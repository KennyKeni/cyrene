package elysia

import (
	"context"
	"log/slog"

	platformgenkit "cyrene/internal/platform/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type service struct {
	clients   *platformgenkit.Clients
	chatStore chatStore
}

func NewService(clients *platformgenkit.Clients, chatStore chatStore) Service {
	return &service{
		clients:   clients,
		chatStore: chatStore,
	}
}

func (s *service) Chat(ctx context.Context, message string, user string) (string, error) {
	chatHistory, err := s.chatStore.Get(ctx, user)
	if err != nil {
		slog.Warn("unable to retrieve chat history", "error", err)
	}

	resp, err := genkit.Generate(ctx, s.clients.Genkit,
		ai.WithModel(s.clients.Model),
		ai.WithSystem(ElysiaPrompt),
		ai.WithMessages(chatHistory...),
		ai.WithPrompt(message),
	)
	if err != nil {
		slog.Error("LLM generation failed", "error", err)
		return "", err
	}

	answer := resp.Text()

	if err := s.chatStore.Append(ctx, user,
		ai.NewUserTextMessage(message),
		ai.NewModelTextMessage(answer),
	); err != nil {
		slog.Warn("failed to append chat history", "error", err)
	}

	return answer, nil
}
