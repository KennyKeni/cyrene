package ingest

import (
	"context"
	"errors"
	"testing"

	"cyrene/internal/pokemon"
	"cyrene/internal/vectorstore"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockEmbedder struct {
	embedFn func(ctx context.Context, texts ...string) ([][]float32, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, texts ...string) ([][]float32, error) {
	return m.embedFn(ctx, texts...)
}

type mockPokemonGetter struct {
	getFn func(ctx context.Context, id string) (*pokemon.Pokemon, error)
}

func (m *mockPokemonGetter) GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error) {
	return m.getFn(ctx, id)
}

type mockStore struct {
	upsertFn   func(ctx context.Context, points ...vectorstore.Point) error
	deleteFn   func(ctx context.Context, filter vectorstore.Filter) error
	upserted   []vectorstore.Point
	deletedRef string
}

func (m *mockStore) Upsert(ctx context.Context, points ...vectorstore.Point) error {
	m.upserted = append(m.upserted, points...)
	if m.upsertFn != nil {
		return m.upsertFn(ctx, points...)
	}
	return nil
}

func (m *mockStore) Delete(ctx context.Context, filter vectorstore.Filter) error {
	if len(filter.StringFilters) > 0 {
		m.deletedRef = filter.StringFilters[0].Value
	}
	if m.deleteFn != nil {
		return m.deleteFn(ctx, filter)
	}
	return nil
}

func (m *mockStore) Search(ctx context.Context, vector []float32, limit int, filter *vectorstore.Filter) ([]vectorstore.SearchResult, error) {
	return nil, nil
}

func (m *mockStore) DeleteByID(ctx context.Context, ids ...string) error {
	return nil
}

type mockRepository struct {
	upsertFn func(ctx context.Context, doc *IngestedDocument) error
	upserted *IngestedDocument
}

func (m *mockRepository) Upsert(ctx context.Context, doc *IngestedDocument) error {
	m.upserted = doc
	if m.upsertFn != nil {
		return m.upsertFn(ctx, doc)
	}
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockRepository) DeleteByRef(ctx context.Context, dt DocumentType, externalID string) error {
	return nil
}

func (m *mockRepository) FindByRef(ctx context.Context, dt DocumentType, externalID string) (*IngestedDocument, error) {
	return nil, nil
}

func (m *mockRepository) InTx(ctx context.Context, fn func(Repository) error) error {
	return fn(m)
}

// Tests

func TestIngestPokemon_Success(t *testing.T) {
	ctx := context.Background()
	pokemonID := "25"
	rawJSON := `{"id":25,"name":"pikachu"}`
	vectors := [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}}

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, texts ...string) ([][]float32, error) {
			assert.Equal(t, rawJSON, texts[0])
			return vectors, nil
		},
	}

	pokemonGetter := &mockPokemonGetter{
		getFn: func(ctx context.Context, id string) (*pokemon.Pokemon, error) {
			assert.Equal(t, pokemonID, id)
			return &pokemon.Pokemon{
				ID:         pokemonID,
				Identifier: "pikachu",
				RawJSON:    rawJSON,
			}, nil
		},
	}

	store := &mockStore{}
	repo := &mockRepository{}

	svc := New(embedder, store, pokemonGetter, repo)

	err := svc.IngestPokemon(ctx, pokemonID)
	require.NoError(t, err)

	// Verify repository upsert
	require.NotNil(t, repo.upserted)
	assert.Equal(t, DocumentTypePokemon, repo.upserted.DocumentType)
	assert.Equal(t, pokemonID, repo.upserted.ExternalID)

	// Verify old vectors deleted
	assert.Equal(t, "pokemon_25", store.deletedRef)

	// Verify new vectors upserted
	require.Len(t, store.upserted, 2)
	for i, point := range store.upserted {
		assert.Equal(t, vectors[i], point.Vector)
		assert.Equal(t, "pokemon_25", point.Payload[referenceKey])
		assert.Equal(t, DocumentTypePokemon, point.Payload[typeKey])
	}
}

func TestIngestPokemon_PokemonFetchError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("pokemon not found")

	pokemonGetter := &mockPokemonGetter{
		getFn: func(ctx context.Context, id string) (*pokemon.Pokemon, error) {
			return nil, expectedErr
		},
	}

	svc := New(&mockEmbedder{}, &mockStore{}, pokemonGetter, &mockRepository{})

	err := svc.IngestPokemon(ctx, "999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetch pokemon")
	assert.ErrorIs(t, err, expectedErr)
}

func TestIngestPokemon_EmbedError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("embedding failed")

	pokemonGetter := &mockPokemonGetter{
		getFn: func(ctx context.Context, id string) (*pokemon.Pokemon, error) {
			return &pokemon.Pokemon{ID: id, RawJSON: "{}"}, nil
		},
	}

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, texts ...string) ([][]float32, error) {
			return nil, expectedErr
		},
	}

	svc := New(embedder, &mockStore{}, pokemonGetter, &mockRepository{})

	err := svc.IngestPokemon(ctx, "25")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed pokemon")
	assert.ErrorIs(t, err, expectedErr)
}

func TestIngestPokemon_RepositoryUpsertError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("db error")

	pokemonGetter := &mockPokemonGetter{
		getFn: func(ctx context.Context, id string) (*pokemon.Pokemon, error) {
			return &pokemon.Pokemon{ID: id, RawJSON: "{}"}, nil
		},
	}

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, texts ...string) ([][]float32, error) {
			return [][]float32{{0.1}}, nil
		},
	}

	repo := &mockRepository{
		upsertFn: func(ctx context.Context, doc *IngestedDocument) error {
			return expectedErr
		},
	}

	svc := New(embedder, &mockStore{}, pokemonGetter, repo)

	err := svc.IngestPokemon(ctx, "25")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upsert document")
	assert.ErrorIs(t, err, expectedErr)
}

func TestIngestPokemon_StoreDeleteError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("qdrant delete failed")

	pokemonGetter := &mockPokemonGetter{
		getFn: func(ctx context.Context, id string) (*pokemon.Pokemon, error) {
			return &pokemon.Pokemon{ID: id, RawJSON: "{}"}, nil
		},
	}

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, texts ...string) ([][]float32, error) {
			return [][]float32{{0.1}}, nil
		},
	}

	store := &mockStore{
		deleteFn: func(ctx context.Context, filter vectorstore.Filter) error {
			return expectedErr
		},
	}

	svc := New(embedder, store, pokemonGetter, &mockRepository{})

	err := svc.IngestPokemon(ctx, "25")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete vectors")
	assert.ErrorIs(t, err, expectedErr)
}

func TestIngestPokemon_StoreUpsertError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("qdrant upsert failed")

	pokemonGetter := &mockPokemonGetter{
		getFn: func(ctx context.Context, id string) (*pokemon.Pokemon, error) {
			return &pokemon.Pokemon{ID: id, RawJSON: "{}"}, nil
		},
	}

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, texts ...string) ([][]float32, error) {
			return [][]float32{{0.1}}, nil
		},
	}

	store := &mockStore{
		upsertFn: func(ctx context.Context, points ...vectorstore.Point) error {
			return expectedErr
		},
	}

	svc := New(embedder, store, pokemonGetter, &mockRepository{})

	err := svc.IngestPokemon(ctx, "25")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upsert vectors")
	assert.ErrorIs(t, err, expectedErr)
}

func TestNewDocumentID(t *testing.T) {
	tests := []struct {
		docType    DocumentType
		externalID string
		expected   string
	}{
		{DocumentTypePokemon, "25", "pokemon_25"},
		{DocumentTypePokemon, "pikachu", "pokemon_pikachu"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := NewDocumentID(tc.docType, tc.externalID)
			assert.Equal(t, tc.expected, result)
		})
	}
}
