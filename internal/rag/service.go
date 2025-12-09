package rag

import (
	"context"
	"time"

	platformgenkit "cyrene/internal/platform/genkit"
	"cyrene/internal/platform/vectorstore"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/google/uuid"
)

const (
	cacheScoreThreshold = float32(0.95)
	payloadTypeCache    = "qa_cache"
)

type service struct {
	clients *platformgenkit.Clients
	pokemon pokemonService
	store   vectorStore

	getPokemonTool ai.Tool
	searchTool     ai.Tool
}

func NewService(clients *platformgenkit.Clients, pokemon pokemonService, store vectorStore) Service {
	s := &service{
		clients: clients,
		pokemon: pokemon,
		store:   store,
	}
	s.registerTools(clients.Genkit)
	return s
}

func (s *service) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	docs := make([]*ai.Document, len(texts))
	for i, text := range texts {
		docs[i] = ai.DocumentFromText(text, nil)
	}

	resp, err := s.clients.Embedder.Embed(ctx, &ai.EmbedRequest{Input: docs})
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		embeddings[i] = emb.Embedding
	}

	return embeddings, nil
}

func (s *service) Chat(ctx context.Context, prompt string) (string, error) {
	embeddings, err := s.Embed(ctx, prompt)
	if err != nil {
		return "", err
	}
	embedding := embeddings[0]

	if cached, err := s.findCachedAnswer(ctx, embedding); err == nil && cached != nil {
		return cached.Answer, nil
	}

	resp, err := genkit.Generate(ctx, s.clients.Genkit,
		ai.WithModel(s.clients.Model),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(prompt),
		ai.WithTools(s.getPokemonTool, s.searchTool),
	)
	if err != nil {
		return "", err
	}

	answer := resp.Text()
	_ = s.storeCachedAnswer(ctx, prompt, embedding, answer)

	return answer, nil
}

func (s *service) findCachedAnswer(ctx context.Context, embedding []float32) (*CachedAnswer, error) {
	filter := &vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: "type", Value: payloadTypeCache},
		},
	}

	results, err := s.store.Search(ctx, embedding, 1, filter)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 || results[0].Score < cacheScoreThreshold {
		return nil, nil
	}

	r := results[0]
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
	return s.store.Upsert(ctx, point)
}
