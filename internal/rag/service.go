package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	platformgenkit "cyrene/internal/platform/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type service struct {
	clients     *platformgenkit.Clients
	pokemon     pokemonService
	chatStore   chatStore
	vectorStore vectorStore

	searchPokemonTool   ai.Tool
	searchMovesTool     ai.Tool
	searchAbilitiesTool ai.Tool
	searchItemsTool     ai.Tool
	listArticlesTool    ai.Tool
	getArticleTool      ai.Tool
	searchTool          ai.Tool
}

func NewService(clients *platformgenkit.Clients, pokemon pokemonService, store vectorStore, chatStore chatStore) Service {
	s := &service{
		clients:     clients,
		pokemon:     pokemon,
		vectorStore: store,
		chatStore:   chatStore,
	}
	s.registerTools(clients.Genkit)
	return s
}

func (s *service) Embed(ctx context.Context, dimensions int, texts ...string) ([][]float32, error) {
	docs := make([]*ai.Document, len(texts))
	for i, text := range texts {
		docs[i] = ai.DocumentFromText(text, nil)
	}

	req := &ai.EmbedRequest{Input: docs}
	if dimensions > 0 {
		req.Options = dimensions
	}

	resp, err := s.clients.Embedder.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		embeddings[i] = emb.Embedding
	}

	return embeddings, nil
}

func (s *service) Chat(ctx context.Context, prompt string, user string) (answer string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("chat panic recovered", "panic", r)
			err = fmt.Errorf("internal error: %v", r)
		}
	}()

	slog.Info("chat request", "prompt", prompt)

	chatHistory, err := s.chatStore.Get(ctx, user)
	if err != nil {
		slog.Warn("Unable to retrieve chat history", "error", err)
	}

	slog.Info("chat length", "length", len(chatHistory))

	newPrompt, err := s.rewritePrompt(ctx, prompt, chatHistory)
	if err != nil {
		return "", err
	}
	if newPrompt.Rejected {
		slog.Info("prompt rejected", "reason", newPrompt.Reason)
		return fmt.Sprintf("I am unable to answer you: %s", newPrompt.Reason), nil
	}

	if len(chatHistory) == 0 {
		prompt = newPrompt.Prompt
	}

	slog.Info("prompt rewritten", "prompt", newPrompt.Prompt)

	generateOpts := []ai.GenerateOption{
		ai.WithModel(s.clients.Model),
		ai.WithSystem(systemPrompt),
		ai.WithTools(s.searchPokemonTool, s.searchMovesTool, s.searchAbilitiesTool, s.searchItemsTool, s.listArticlesTool, s.getArticleTool, s.searchTool),
	}
	if len(chatHistory) > 0 {
		generateOpts = append(generateOpts, ai.WithMessages(chatHistory...))
	}
	generateOpts = append(generateOpts, ai.WithPrompt(prompt))

	resp, err := genkit.Generate(ctx, s.clients.Genkit, generateOpts...)
	if err != nil {
		slog.Error("LLM generation failed", "error", err)
		if errMsg := err.Error(); strings.Contains(errMsg, "max") || strings.Contains(errMsg, "loop") || strings.Contains(errMsg, "iteration") {
			return "I wasn't able to find the information after several attempts. Could you try rephrasing your question or being more specific?", nil
		}
		return "", err
	}

	answer = resp.Text()

	if err := s.chatStore.Append(ctx, user,
		ai.NewUserTextMessage(prompt),
		ai.NewModelTextMessage(answer),
	); err != nil {
		slog.Warn("failed to append chat history", "error", err)
	}

	return answer, nil
}

func (s *service) ChatStream(ctx context.Context, prompt string, user string, onChunk func(string) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("chat stream panic recovered", "panic", r)
			err = fmt.Errorf("internal error: %v", r)
		}
	}()

	slog.Info("chat stream request", "prompt", prompt)

	chatHistory, err := s.chatStore.Get(ctx, user)
	if err != nil {
		slog.Warn("Unable to retrieve chat history", "error", err)
	}

	newPrompt, err := s.rewritePrompt(ctx, prompt, chatHistory)
	if err != nil {
		return err
	}
	if newPrompt.Rejected {
		slog.Info("prompt rejected", "reason", newPrompt.Reason)
		return onChunk(fmt.Sprintf("I am unable to answer you: %s", newPrompt.Reason))
	}

	if len(chatHistory) == 0 {
		prompt = newPrompt.Prompt
	}

	var answer strings.Builder
	generateOpts := []ai.GenerateOption{
		ai.WithModel(s.clients.Model),
		ai.WithSystem(systemPrompt),
		ai.WithTools(s.searchPokemonTool, s.searchMovesTool, s.searchAbilitiesTool, s.searchItemsTool, s.listArticlesTool, s.getArticleTool, s.searchTool),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			text := chunk.Text()
			if text != "" {
				answer.WriteString(text)
				return onChunk(text)
			}
			return nil
		}),
	}
	if len(chatHistory) > 0 {
		generateOpts = append(generateOpts, ai.WithMessages(chatHistory...))
	}
	generateOpts = append(generateOpts, ai.WithPrompt(prompt))

	resp, err := genkit.Generate(ctx, s.clients.Genkit, generateOpts...)
	if err != nil {
		slog.Error("LLM streaming failed", "error", err)
		if errMsg := err.Error(); strings.Contains(errMsg, "max") || strings.Contains(errMsg, "loop") || strings.Contains(errMsg, "iteration") {
			return onChunk("I wasn't able to find the information after several attempts. Could you try rephrasing your question or being more specific?")
		}
		return err
	}

	fullAnswer := resp.Text()
	if fullAnswer == "" {
		fullAnswer = answer.String()
	}

	if err := s.chatStore.Append(ctx, user,
		ai.NewUserTextMessage(prompt),
		ai.NewModelTextMessage(fullAnswer),
	); err != nil {
		slog.Warn("failed to append chat history", "error", err)
	}

	return nil
}

func (s *service) rewritePrompt(ctx context.Context, query string, chatHistory []*ai.Message) (*rewriteResult, error) {
	resp, _, err := genkit.GenerateData[rewriteResult](ctx, s.clients.Genkit,
		ai.WithModel(s.clients.FastModel),
		ai.WithSystem(rewritePrompt),
		ai.WithMessages(chatHistory...),
		ai.WithPrompt(query),
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

