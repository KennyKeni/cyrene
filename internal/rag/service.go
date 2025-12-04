package rag

import (
	"context"
	"cyrene/internal/platform/config"
	"cyrene/internal/platform/genkit"

	"github.com/firebase/genkit/go/ai"
)

type Service struct {
	model    ai.Model
	embedder ai.Embedder
}

func New(cfg *config.GenkitConfig, clients *genkit.Clients) *Service {
	return &Service{
		model:    clients.Model,
		embedder: clients.Embedder,
	}
}

func (s *Service) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	docs := make([]*ai.Document, len(texts))
	for i, text := range texts {
		docs[i] = ai.DocumentFromText(text, nil)
	}

	req := &ai.EmbedRequest{
		Input: docs,
	}

	resp, err := s.embedder.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		embeddings[i] = emb.Embedding
	}

	return embeddings, nil
}
