package elysia

import (
	"cyrene/internal/config"

	"github.com/KennyKeni/elysia/adapter/openai"
	"github.com/KennyKeni/elysia/agent"
	"github.com/KennyKeni/elysia/client"
	"github.com/KennyKeni/elysia/types"
)

type empty struct{}

type Clients struct {
	EmbedClient types.Client
	Agent       *agent.Agent[empty, empty]
}

func New(cfg *config.ElysiaConfig) *Clients {
	ec := openai.NewClient(client.WithAPIKey(cfg.EmbedAPIKey), client.WithBaseURL(cfg.EmbedURL))

	ac := openai.NewClient(client.WithAPIKey(cfg.AgentAPIKey), client.WithBaseURL(cfg.AgentURL))

	a, _ := agent.New[empty, empty](ac,
		agent.WithModel[empty, empty](cfg.AgentModel),
		agent.WithSystemPrompt[empty, empty]("You are a Pokemon expert..."),
		agent.WithTools[empty, empty](),
	)

	return &Clients{EmbedClient: ec, Agent: a}
}
