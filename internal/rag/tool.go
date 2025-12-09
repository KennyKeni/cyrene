package rag

import (
	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func (s *service) registerTools(g *genkit.Genkit) {
	s.getPokemonTool = s.defineGetPokemonTool(g)
	s.searchTool = s.defineVectorSearchTool(g)
}

func (s *service) defineGetPokemonTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"getPokemon",
		"Fetches complete Pokemon data by ID or name from the PokeAPI. Returns exact stats (HP, attack, defense, speed, etc.), types, abilities, moves, height, and weight. Use this when you need precise details about a specific Pokemon.",
		func(ctx *ai.ToolContext, input struct {
			ID string `json:"id" jsonschema_description:"Pokemon ID (e.g. '25') or name (e.g. 'pikachu')"`
		}) (*pokemon.Pokemon, error) {
			return s.pokemon.GetPokemonByID(ctx, input.ID)
		},
	)
}

func (s *service) defineVectorSearchTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchPokemon",
		"Searches the Pokemon database using semantic similarity. Use for exploratory queries like finding Pokemon by type, abilities, characteristics, or conceptual similarities (e.g. 'fast electric Pokemon', 'tanky water types', 'Pokemon that can learn fire moves'). Returns ranked results with relevance scores.",
		func(ctx *ai.ToolContext, input struct {
			Query string `json:"query" jsonschema_description:"Natural language search query describing the Pokemon you're looking for"`
			Limit int    `json:"limit" jsonschema_description:"Max results to return (default 5)"`
		}) ([]vectorstore.SearchResult, error) {
			embeddings, err := s.Embed(ctx, input.Query)
			if err != nil {
				return nil, err
			}
			return s.store.Search(ctx, embeddings[0], input.Limit, nil)
		},
	)
}
