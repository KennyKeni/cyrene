package elysia

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

type Service interface {
	Chat(ctx context.Context, message string, user string) (string, error)
}

type chatStore interface {
	Get(ctx context.Context, username string) ([]*ai.Message, error)
	Append(ctx context.Context, username string, msgs ...*ai.Message) error
}
