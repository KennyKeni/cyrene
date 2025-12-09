package ingest

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	ingestPokemonFn func(ctx context.Context, pokemonID string) error
}

func (m *mockService) IngestPokemon(ctx context.Context, pokemonID string) error {
	if m.ingestPokemonFn != nil {
		return m.ingestPokemonFn(ctx, pokemonID)
	}
	return nil
}

func TestHandler_HandleIngest_Pokemon_Success(t *testing.T) {
	ctx := context.Background()
	payload := []byte(`{"type":"pokemon","id":"25"}`)

	var calledWith string
	svc := &mockService{
		ingestPokemonFn: func(ctx context.Context, pokemonID string) error {
			calledWith = pokemonID
			return nil
		},
	}

	h := NewHandler(svc)
	err := h.HandleIngest(ctx, payload)

	require.NoError(t, err)
	assert.Equal(t, "25", calledWith)
}

func TestHandler_HandleIngest_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	payload := []byte(`{invalid json}`)

	h := NewHandler(&mockService{})
	err := h.HandleIngest(ctx, payload)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestHandler_HandleIngest_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	payload := []byte(`{"type":"unknown","id":"1"}`)

	h := NewHandler(&mockService{})
	err := h.HandleIngest(ctx, payload)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported event type")
}

func TestHandler_HandleIngest_ServiceError(t *testing.T) {
	ctx := context.Background()
	payload := []byte(`{"type":"pokemon","id":"999"}`)
	expectedErr := errors.New("pokemon not found")

	svc := &mockService{
		ingestPokemonFn: func(ctx context.Context, pokemonID string) error {
			return expectedErr
		},
	}

	h := NewHandler(svc)
	err := h.HandleIngest(ctx, payload)

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}
