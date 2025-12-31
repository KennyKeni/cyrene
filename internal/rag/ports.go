package rag

import (
	"context"

	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/firebase/genkit/go/ai"
)

type Service interface {
	Chat(ctx context.Context, prompt string, user string) (string, error)
	ChatStream(ctx context.Context, prompt string, user string, onChunk func(string) error) error
	Embed(ctx context.Context, dimensions int, texts ...string) ([][]float32, error)
}

type pokemonService interface {
	SearchPokemon(ctx context.Context, params pokemon.AgentPokemonParams) (*pokemon.PaginatedResponse[pokemon.AgentPokemon], error)
	SearchMoves(ctx context.Context, params pokemon.AgentMoveParams) (*pokemon.PaginatedResponse[pokemon.AgentMove], error)
	SearchAbilities(ctx context.Context, params pokemon.AgentAbilityParams) (*pokemon.PaginatedResponse[pokemon.AgentAbility], error)
	SearchItems(ctx context.Context, params pokemon.AgentItemParams) (*pokemon.PaginatedResponse[pokemon.AgentItem], error)
	SearchArticles(ctx context.Context, params pokemon.AgentArticleParams) (*pokemon.PaginatedResponse[pokemon.AgentArticleSearch], error)
	GetArticleBySlug(ctx context.Context, slug string) (*pokemon.AgentArticle, error)
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
