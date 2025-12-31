package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

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
	if event.Operation == OperationDelete {
		return s.handleDelete(ctx, event)
	}

	switch event.EntityType {
	case EntityTypeSpecies:
		return s.ingestSpecies(ctx, event.EntityID)
	case EntityTypeForm:
		return s.ingestForm(ctx, event.EntityID)
	case EntityTypeMove:
		return s.ingestMove(ctx, event.EntityID)
	case EntityTypeAbility:
		return s.ingestAbility(ctx, event.EntityID)
	case EntityTypeItem:
		return s.ingestItem(ctx, event.EntityID)
	case EntityTypeArticle:
		return s.ingestArticle(ctx, event.EntityID)
	default:
		return fmt.Errorf("unsupported entity type: %s", event.EntityType)
	}
}

func (s *service) handleDelete(ctx context.Context, event IngestionEvent) error {
	var docType DocumentType
	var idKey string

	switch event.EntityType {
	case EntityTypeSpecies, EntityTypeForm:
		docType = DocumentTypeForm
		idKey = formIDKey
	case EntityTypeMove:
		docType = DocumentTypeMove
		idKey = moveIDKey
	case EntityTypeAbility:
		docType = DocumentTypeAbility
		idKey = abilityIDKey
	case EntityTypeItem:
		docType = DocumentTypeItem
		idKey = itemIDKey
	case EntityTypeArticle:
		docType = DocumentTypeArticle
		idKey = articleIDKey
	default:
		return fmt.Errorf("unsupported entity type for delete: %s", event.EntityType)
	}

	if err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: idKey, Value: event.EntityID, Op: vectorstore.FilterAND},
		},
	}); err != nil {
		return fmt.Errorf("delete vectors: %w", err)
	}

	return s.repository.DeleteByRef(ctx, docType, event.EntityID)
}

func (s *service) ingestSpecies(ctx context.Context, id string) error {
	species, err := s.pokemonService.GetSpecies(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch species: %w", err)
	}

	if len(species.Forms) == 0 {
		slog.Warn("species has no forms", "speciesId", id, "speciesName", species.Name)
		return nil
	}

	slog.Info("ingesting species forms", "speciesId", id, "speciesName", species.Name, "formCount", len(species.Forms))

	var lastErr error
	for _, form := range species.Forms {
		formID := strconv.Itoa(form.ID)
		if err := s.ingestForm(ctx, formID); err != nil {
			slog.Error("failed to ingest form", "speciesId", id, "formId", formID, "error", err)
			lastErr = err
			continue
		}
	}

	if lastErr != nil {
		slog.Warn("species ingestion completed with errors", "speciesId", id, "lastError", lastErr)
	}

	return nil
}

func (s *service) ingestForm(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: formIDKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete form vectors: %w", err)
	}

	pokemon, err := s.pokemonService.GetPokemon(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch pokemon form: %w", err)
	}

	content := pokemon.EmbeddingText()
	metadata := map[string]any{
		typeKey:   string(DocumentTypeForm),
		formIDKey: id,
		contentKey: content,
	}

	return s.doIngest(ctx, ingestInput{
		docType:  DocumentTypeForm,
		id:       id,
		content:  content,
		metadata: metadata,
	})
}

func (s *service) ingestMove(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: moveIDKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete move vectors: %w", err)
	}

	move, err := s.pokemonService.GetMove(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch move: %w", err)
	}

	content := move.EmbeddingText()
	metadata := map[string]any{
		typeKey:    string(DocumentTypeMove),
		moveIDKey:  id,
		contentKey: content,
	}

	return s.doIngest(ctx, ingestInput{
		docType:  DocumentTypeMove,
		id:       id,
		content:  content,
		metadata: metadata,
	})
}

func (s *service) ingestAbility(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: abilityIDKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete ability vectors: %w", err)
	}

	ability, err := s.pokemonService.GetAbility(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch ability: %w", err)
	}

	content := ability.EmbeddingText()
	metadata := map[string]any{
		typeKey:      string(DocumentTypeAbility),
		abilityIDKey: id,
		contentKey:   content,
	}

	return s.doIngest(ctx, ingestInput{
		docType:  DocumentTypeAbility,
		id:       id,
		content:  content,
		metadata: metadata,
	})
}

func (s *service) ingestItem(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: itemIDKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete item vectors: %w", err)
	}

	item, err := s.pokemonService.GetItem(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch item: %w", err)
	}

	content := item.EmbeddingText()
	metadata := map[string]any{
		typeKey:    string(DocumentTypeItem),
		itemIDKey:  id,
		contentKey: content,
	}

	return s.doIngest(ctx, ingestInput{
		docType:  DocumentTypeItem,
		id:       id,
		content:  content,
		metadata: metadata,
	})
}

func (s *service) ingestArticle(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: articleIDKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete article vectors: %w", err)
	}

	article, err := s.pokemonService.GetArticle(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch article: %w", err)
	}

	content := article.EmbeddingText()
	metadata := map[string]any{
		typeKey:      string(DocumentTypeArticle),
		articleIDKey: id,
		contentKey:   content,
	}

	return s.doIngest(ctx, ingestInput{
		docType:  DocumentTypeArticle,
		id:       id,
		content:  content,
		metadata: metadata,
	})
}

type ingestInput struct {
	docType  DocumentType
	id       string
	content  string
	metadata map[string]any
}

func (s *service) doIngest(ctx context.Context, input ingestInput) error {
	if input.content == "" {
		return fmt.Errorf("%s %s has empty embedding text", input.docType, input.id)
	}

	vectors, err := s.embedService.Embed(ctx, s.store.Dimensions(), input.content)
	if err != nil {
		return fmt.Errorf("embed %s: %w", input.docType, err)
	}

	if len(vectors) == 0 {
		return fmt.Errorf("no embeddings generated for %s %s", input.docType, input.id)
	}

	return s.repository.InTx(ctx, func(repo Repository) error {
		err := repo.Upsert(ctx, &IngestedDocument{
			ID:           uuid.Must(uuid.NewV7()),
			DocumentType: input.docType,
			ExternalID:   input.id,
		})
		if err != nil {
			return fmt.Errorf("upsert document: %w", err)
		}

		points := make([]vectorstore.Point, len(vectors))
		for i, vector := range vectors {
			points[i] = vectorstore.Point{
				ID:      uuid.Must(uuid.NewV7()).String(),
				Vector:  vector,
				Payload: input.metadata,
			}
		}

		return s.store.Upsert(ctx, points...)
	})
}
