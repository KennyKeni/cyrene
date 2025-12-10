package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	platformgenkit "cyrene/internal/platform/genkit"
	"cyrene/internal/platform/vectorstore"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/google/uuid"
)

type service struct {
	clients     *platformgenkit.Clients
	pokemon     pokemonService
	chatStore   chatStore
	vectorStore vectorStore
	cacheStore  vectorStore

	getPokemonTool ai.Tool
	searchTool     ai.Tool
}

func NewService(clients *platformgenkit.Clients, pokemon pokemonService, store vectorStore, cacheStore vectorStore, chatStore chatStore) Service {
	s := &service{
		clients:     clients,
		pokemon:     pokemon,
		vectorStore: store,
		cacheStore:  cacheStore,
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

	embeddings, err := s.Embed(ctx, s.cacheStore.Dimensions(), newPrompt.Prompt)
	if err != nil {
		slog.Error("failed to embed prompt", "error", err)
		return "", err
	}
	embedding := embeddings[0]

	if cached, err := s.findCachedAnswer(ctx, prompt, embedding); err == nil && cached != nil {
		slog.Info("cache hit", "cached_question", cached.Question)
		if err := s.chatStore.Append(ctx, user,
			ai.NewUserTextMessage(prompt),
			ai.NewModelTextMessage(cached.Answer),
		); err != nil {
			slog.Warn("failed to append chat history", "error", err)
		}
		return cached.Answer, nil
	}
	slog.Info("cache miss, calling LLM")

	resp, err := genkit.Generate(ctx, s.clients.Genkit,
		ai.WithModel(s.clients.Model),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(prompt),
		ai.WithTools(s.getPokemonTool, s.searchTool),
	)
	if err != nil {
		slog.Error("LLM generation failed", "error", err)
		return "", err
	}

	answer = resp.Text()
	if err := s.storeCachedAnswer(ctx, newPrompt.Prompt, embedding, answer); err != nil {
		slog.Warn("failed to cache answer", "error", err)
	} else {
		slog.Info("cached answer stored")
	}

	if err := s.chatStore.Append(ctx, user,
		ai.NewUserTextMessage(prompt),
		ai.NewModelTextMessage(answer),
	); err != nil {
		slog.Warn("failed to append chat history", "error", err)
	}

	return answer, nil
}

func (s *service) findCachedAnswer(ctx context.Context, query string, embedding []float32) (*CachedAnswer, error) {
	filter := &vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: "type", Value: payloadTypeCache},
		},
	}

	results, err := s.cacheStore.Search(ctx, embedding, cacheTopN, filter)
	if err != nil {
		slog.Error("cache search failed", "error", err)
		return nil, err
	}
	slog.Info("cache search results", "count", len(results))

	var candidates []vectorstore.SearchResult
	for _, r := range results {
		// Essentially the same question
		if r.Score >= cacheHeuristicScoreThreshold {
			slog.Info("Returning answer due to heuristic threshold")
			return &CachedAnswer{
				Question: r.Payload["question"].(string),
				Answer:   r.Payload["answer"].(string),
			}, nil
		}
		if r.Score >= cacheScoreThreshold {
			candidates = append(candidates, r)
		}
	}
	if len(candidates) == 0 {
		slog.Info("No candidates")
		return nil, nil
	}

	slog.Info("cache candidates found", "count", len(candidates), "top_score", candidates[0].Score)

	var sb strings.Builder
	for i, r := range candidates {
		answer := r.Payload["answer"].(string)
		if len(answer) > cacheAnswerMaxLen {
			answer = answer[:cacheAnswerMaxLen] + "..."
		}
		_, err := fmt.Fprintf(&sb, "%d. Q: %s\n   A: %s\n", i, r.Payload["question"], answer)
		if err != nil {
			return nil, err
		}
	}

	validation, _, err := genkit.GenerateData[cacheValidation](ctx, s.clients.Genkit,
		ai.WithModel(s.clients.FastModel),
		ai.WithSystem("Pick the cached Q&A that answers the user's query. Return match_index as -1 if none are applicable."),
		ai.WithPrompt("User query: %s\n\nCached Q&A:\n%s", query, sb.String()),
	)
	if err != nil {
		slog.Warn("cache validation failed", "error", err)
		return nil, nil
	}

	slog.Info("cache validation result", "match_index", validation.MatchIndex, "reason", validation.Reason)

	if validation.MatchIndex < 0 || validation.MatchIndex >= len(candidates) {
		return nil, nil
	}

	r := candidates[validation.MatchIndex]
	return &CachedAnswer{
		Question: r.Payload["question"].(string),
		Answer:   r.Payload["answer"].(string),
	}, nil
}

func (s *service) storeCachedAnswer(ctx context.Context, question string, embedding []float32, answer string) error {
	point := vectorstore.Point{
		ID:     uuid.New().String(),
		Vector: embedding,
		Payload: map[string]any{
			"type":       payloadTypeCache,
			"question":   question,
			"answer":     answer,
			"created_at": time.Now().Unix(),
		},
	}
	return s.cacheStore.Upsert(ctx, point)
}

func (s *service) fastModelAsk(ctx context.Context, system string, prompt string) (string, error) {
	resp, err := genkit.Generate(ctx, s.clients.Genkit,
		ai.WithModel(s.clients.FastModel),
		ai.WithSystem(system),
		ai.WithPrompt(prompt),
	)
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
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
