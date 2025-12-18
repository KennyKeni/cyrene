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
	GetMoveByID(ctx context.Context, id string) (*pokemon.Move, error)
	SearchMoves(ctx context.Context, query string, limit int) ([]pokemon.MoveSearchResult, error)
	SearchTypes(ctx context.Context, query string, limit int) ([]pokemon.TypeSearchResult, error)
	GetAbilityByID(ctx context.Context, id string) (*pokemon.Ability, error)
	SearchAbilities(ctx context.Context, query string, limit int) ([]pokemon.AbilitySearchResult, error)
	GetArticleByID(ctx context.Context, id string) (*pokemon.Article, error)
	SearchArticles(ctx context.Context, query string, limit int) ([]pokemon.ArticleSearchResult, error)
	SearchForms(ctx context.Context, params pokemon.FormSearchParams) (*pokemon.FormSearchResponse, error)
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
