package genkit

import (
	"context"

	"cyrene/internal/platform/config"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Clients struct {
	Genkit    *genkit.Genkit
	Embedder  ai.Embedder
	Model     ai.Model
	FastModel ai.Model
}

func New(ctx context.Context, cfg *config.GenkitConfig) (*Clients, error) {
	agentProvider := &compat_oai.OpenAICompatible{
		APIKey:   cfg.AgentAPIKey,
		BaseURL:  cfg.AgentURL,
		Provider: "agent",
	}

	g := genkit.Init(ctx, genkit.WithPlugins(agentProvider))

	embedClient := openai.NewClient(
		option.WithAPIKey(cfg.EmbedAPIKey),
		option.WithBaseURL(cfg.EmbedURL),
	)
	embedder := newEmbedder(embedClient, cfg.EmbedModel)

	agentModel := agentProvider.DefineModel("agent", cfg.AgentModel, ai.ModelOptions{})
	fastModel := agentProvider.DefineModel("agent", cfg.FastModel, ai.ModelOptions{})

	return &Clients{
		Genkit:    g,
		Embedder:  embedder,
		Model:     agentModel,
		FastModel: fastModel,
	}, nil
}

func newEmbedder(client openai.Client, model string) ai.Embedder {
	return ai.NewEmbedder(
		"embed/"+model,
		nil,
		func(ctx context.Context, req *ai.EmbedRequest) (*ai.EmbedResponse, error) {
			var texts []string
			for _, doc := range req.Input {
				for _, p := range doc.Content {
					texts = append(texts, p.Text)
				}
			}

			params := openai.EmbeddingNewParams{
				Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
				Model: model,
			}

			if dim, ok := req.Options.(int); ok && dim > 0 {
				params.Dimensions = openai.Int(int64(dim))
			}

			resp, err := client.Embeddings.New(ctx, params)
			if err != nil {
				return nil, err
			}

			embeddings := make([]*ai.Embedding, len(resp.Data))
			for i, emb := range resp.Data {
				vec := make([]float32, len(emb.Embedding))
				for j, v := range emb.Embedding {
					vec[j] = float32(v)
				}
				embeddings[i] = &ai.Embedding{Embedding: vec}
			}

			return &ai.EmbedResponse{Embeddings: embeddings}, nil
		},
	)
}
