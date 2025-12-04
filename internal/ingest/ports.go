package ingest

import (
	"context"

	"cyrene/internal/pokemon"

	"github.com/google/uuid"
)

type embedder interface {
	Embed(ctx context.Context, texts ...string) ([][]float32, error)
}

type pokemonGetter interface {
	GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error)
}

type Repository interface {
	Upsert(ctx context.Context, doc *IngestedDocument) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByRef(ctx context.Context, dt DocumentType, externalID string) error
	FindByRef(ctx context.Context, dt DocumentType, externalID string) (*IngestedDocument, error)
	InTx(ctx context.Context, fn func(Repository) error) error
}
