package genkit

import (
	"context"
	"cyrene/internal/platform/config"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai"
)

type Clients struct {
	Genkit   *genkit.Genkit
	Embedder ai.Embedder
	Model    ai.Model
}

func New(ctx context.Context, cfg *config.GenkitConfig) (*Clients, error) {
	// Configure embedding provider with OpenRouter
	embedProvider := &compat_oai.OpenAICompatible{
		APIKey:   cfg.EmbedAPIKey,
		BaseURL:  cfg.EmbedURL,
		Provider: "embed",
	}

	// Configure agent provider with OpenRouter
	agentProvider := &compat_oai.OpenAICompatible{
		APIKey:   cfg.AgentAPIKey,
		BaseURL:  cfg.AgentURL,
		Provider: "agent",
	}

	// Initialize Genkit with plugins
	g := genkit.Init(ctx, genkit.WithPlugins(embedProvider, agentProvider))

	// Define embedder and model
	embedder := embedProvider.DefineEmbedder("embed", cfg.EmbedModel, nil)
	agentModel := agentProvider.DefineModel("agent", cfg.AgentModel, ai.ModelOptions{})

	return &Clients{
		Genkit:   g,
		Embedder: embedder,
		Model:    agentModel,
	}, nil
}
