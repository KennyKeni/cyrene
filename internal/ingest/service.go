package ingest

import (
	"context"
	"fmt"

	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

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
	case EventTypeSpecies:
		return s.ingestSpecies(ctx, event.ID)
	case EventTypeVariation:
		return s.ingestVariation(ctx, event.ID)
	case EventTypeForm:
		return s.ingestForm(ctx, event.ID)
	case EventTypeMove:
		return s.ingestMove(ctx, event.ID)
	case EventTypeAbility:
		return s.ingestAbility(ctx, event.ID)
	case EventTypeArticle:
		return s.ingestArticle(ctx, event.ID)
	default:
		return fmt.Errorf("unsupported event type: %s", event.Type)
	}
}

func (s *service) ingestSpecies(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: speciesIDKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete species vectors: %w", err)
	}

	resp, err := s.pokemonService.SearchForms(ctx, pokemon.FormSearchParams{
		SpeciesID: id,
		Include:   []string{"stats", "types", "abilities", "moves", "variants"},
		Limit:     100,
	})
	if err != nil {
		return fmt.Errorf("fetch forms for species: %w", err)
	}

	for i := range resp.Data {
		if err := s.ingestFormData(ctx, &resp.Data[i]); err != nil {
			return fmt.Errorf("ingest form %s: %w", resp.Data[i].FormIdentifier, err)
		}
	}

	return nil
}

func (s *service) ingestVariation(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, vectorstore.Filter{
		StringFilters: []vectorstore.StringFilter{
			{Field: variantIDsKey, Value: id, Op: vectorstore.FilterAND},
		},
	})
	if err != nil {
		return fmt.Errorf("delete variation vectors: %w", err)
	}

	resp, err := s.pokemonService.SearchForms(ctx, pokemon.FormSearchParams{
		VariationID: id,
		Include:     []string{"stats", "types", "abilities", "moves", "variants"},
		Limit:       1,
	})
	if err != nil {
		return fmt.Errorf("fetch form for variation: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil
	}

	return s.ingestFormData(ctx, &resp.Data[0])
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

	resp, err := s.pokemonService.SearchForms(ctx, pokemon.FormSearchParams{
		FormID:  id,
		Include: []string{"stats", "types", "abilities", "moves", "variants"},
		Limit:   1,
	})
	if err != nil {
		return fmt.Errorf("fetch form: %w", err)
	}

	if len(resp.Data) == 0 {
		return fmt.Errorf("form %s not found", id)
	}

	return s.ingestFormData(ctx, &resp.Data[0])
}

func (s *service) ingestFormData(ctx context.Context, form *pokemon.FormSearchResult) error {
	content := form.EmbeddingText()

	var variantIDs []any
	for _, v := range form.PokemonVariations {
		variantIDs = append(variantIDs, v.Identifier)
	}

	formID := fmt.Sprintf("%s:%s", form.Species.Identifier, form.FormIdentifier)

	metadata := map[string]any{
		typeKey:       string(DocumentTypeForm),
		formIDKey:     formID,
		speciesIDKey:  form.Species.Identifier,
		variantIDsKey: variantIDs,
		formNameKey:   form.FormName,
		contentKey:    content,
	}

	return s.doIngest(ctx, ingestInput{
		docType:  DocumentTypeForm,
		id:       formID,
		content:  content,
		metadata: metadata,
	})
}

func (s *service) ingestMove(ctx context.Context, id string) error {
	move, err := s.pokemonService.GetMoveByID(ctx, id)
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
	ability, err := s.pokemonService.GetAbilityByID(ctx, id)
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

func (s *service) ingestArticle(ctx context.Context, id string) error {
	article, err := s.pokemonService.GetArticleByID(ctx, id)
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
