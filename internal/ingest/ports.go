package ingest

import (
	"context"

	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/google/uuid"
)

type Service interface {
	Ingest(ctx context.Context, event IngestionEvent) error
}

type embedService interface {
	Embed(ctx context.Context, texts ...string) ([][]float32, error)
}

type pokemonService interface {
	GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error)
}

type vectorStore interface {
	Upsert(ctx context.Context, points ...vectorstore.Point) error
	Delete(ctx context.Context, filter vectorstore.Filter) error
}

type Repository interface {
	Upsert(ctx context.Context, doc *IngestedDocument) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByRef(ctx context.Context, dt DocumentType, externalID string) error
	FindByRef(ctx context.Context, dt DocumentType, externalID string) (*IngestedDocument, error)
	InTx(ctx context.Context, fn func(Repository) error) error
}
