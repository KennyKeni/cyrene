package llm

import (
	"context"
	"cyrene/internal/config"
	"cyrene/internal/elysia"

	"github.com/KennyKeni/elysia/types"
)

type Service interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

type service struct {
	embedClient types.Client
	embedModel  string
}

func New(cfg *config.ElysiaConfig, clients *elysia.Clients) Service {
	return &service{
		embedClient: clients.EmbedClient,
		embedModel:  cfg.EmbedModel,
	}
}

func (s *service) Embed(ctx context.Context, text string) ([]float64, error) {
	params := types.NewEmbeddingParams(
		types.WithEmbeddingModel(s.embedModel),
		types.WithInput([]string{text}),
	)

	resp, err := s.embedClient.Embed(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Embeddings[0].Vector, nil
}
