package rag

import (
	"context"

	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/firebase/genkit/go/ai"
)

type Service interface {
	Chat(ctx context.Context, prompt string, user string) (string, error)
	Embed(ctx context.Context, dimensions int, texts ...string) ([][]float32, error)
}

type pokemonService interface {
	GetPokemonByID(ctx context.Context, id string) (*pokemon.Pokemon, error)
}

type vectorStore interface {
	Search(ctx context.Context, vector []float32, limit int, filter *vectorstore.Filter) ([]vectorstore.SearchResult, error)
	Upsert(ctx context.Context, points ...vectorstore.Point) error
	Dimensions() int
}

type chatStore interface {
	Get(ctx context.Context, username string) ([]*ai.Message, error)
	Append(ctx context.Context, username string, msgs ...*ai.Message) error
	Clear(ctx context.Context, username string) error
}
