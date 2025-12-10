package ingest

import (
	"context"
	"fmt"

	"cyrene/internal/platform/vectorstore"

	"github.com/google/uuid"
)

type service struct {
	embedService   embedService
	store          vectorStore
	pokemonService pokemonService
	repository     Repository
}

func NewService(embedService embedService, store vectorStore, pokemonService pokemonService, repository Repository) *service {
	return &service{
		embedService:   embedService,
		store:          store,
		pokemonService: pokemonService,
		repository:     repository,
	}
}

func (s *service) Ingest(ctx context.Context, event IngestionEvent) error {
	switch event.Type {
	case DocumentTypePokemon:
		return s.ingestPokemon(ctx, event.ID)
	default:
		return fmt.Errorf("unsupported document type: %s", event.Type)
	}
}

func (s *service) ingestPokemon(ctx context.Context, pokemonID string) error {
	pokemon, err := s.pokemonService.GetPokemonByID(ctx, pokemonID)
	if err != nil {
		return fmt.Errorf("fetch pokemon: %w", err)
	}

	embeddingText := pokemon.EmbeddingText()
	if embeddingText == "" {
		return fmt.Errorf("pokemon %s has empty embedding text", pokemonID)
	}

	vectors, err := s.embedService.Embed(ctx, s.store.Dimensions(), embeddingText)
	if err != nil {
		return fmt.Errorf("embed pokemon: %w", err)
	}

	if len(vectors) == 0 {
		return fmt.Errorf("no embeddings generated for pokemon %s", pokemonID)
	}

	return s.repository.InTx(ctx, func(repo Repository) error {
		return s.ingestDocument(ctx, repo, DocumentTypePokemon, pokemonID, embeddingText, vectors)
	})
}

func (s *service) ingestDocument(
	ctx context.Context,
	repo Repository,
	docType DocumentType,
	externalID string,
	content string,
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
		typeKey:      string(docType),
		contentKey:   content,
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
