package rag

import (
	"context"

	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"
)

type Service interface {
	Chat(ctx context.Context, prompt string) (string, error)
	Embed(ctx context.Context, texts ...string) ([][]float32, error)
}

type pokemonService interface {
	GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error)
}

type vectorStore interface {
	Search(ctx context.Context, vector []float32, limit int, filter *vectorstore.Filter) ([]vectorstore.SearchResult, error)
	Upsert(ctx context.Context, points ...vectorstore.Point) error
}
