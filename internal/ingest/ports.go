package ingest

import (
	"context"

	"cyrene/internal/pokemon"
)

type embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type pokemonGetter interface {
	GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error)
}
