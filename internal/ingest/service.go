package ingest

import (
	"context"
	"cyrene/internal/vectorstore"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	embedder       embedder
	store          vectorstore.Store
	pokemonService pokemonGetter
	repository     Repository
}

func New(embedder embedder, store vectorstore.Store, pokemonService pokemonGetter, repository Repository) *Service {
	return &Service{
		embedder:       embedder,
		store:          store,
		pokemonService: pokemonService,
		repository:     repository,
	}
}

func (s *Service) IngestPokemon(ctx context.Context, pokemonID string) error {
	pokemon, err := s.pokemonService.GetPokemonByID(ctx, pokemonID)
	if err != nil {
		return fmt.Errorf("fetch pokemon: %w", err)
	}

	vectors, err := s.embedder.Embed(ctx, pokemon.RawJSON)
	if err != nil {
		return fmt.Errorf("embed pokemon: %w", err)
	}

	return s.repository.InTx(ctx, func(repo Repository) error {
		return s.ingestDocument(ctx, repo, DocumentTypePokemon, pokemonID, vectors)
	})
}

// ingestDocument upserts a document record to Postgres and stores vectors in Qdrant.
// Embedding/splitting logic is handled by the caller.
func (s *Service) ingestDocument(
	ctx context.Context,
	repo Repository,
	docType DocumentType,
	externalID string,
	vectors [][]float32,
) error {
	reference := NewDocumentID(docType, externalID)

	err := repo.Upsert(ctx, &IngestedDocument{
		ID:           uuid.Must(uuid.NewV7()),
		DocumentType: docType,
		ExternalID:   externalID,
	})
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}

	err = s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: referenceKey, Value: reference, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete vectors: %w", err)
	}

	points := make([]vectorstore.Point, len(vectors))
	metadata := map[string]any{
		referenceKey: reference,
		typeKey:      docType,
	}

	for i, vector := range vectors {
		points[i] = vectorstore.Point{
			ID:      uuid.Must(uuid.NewV7()).String(),
			Vector:  vector,
			Payload: metadata,
		}
	}

	return s.store.Upsert(ctx, points...)
}
