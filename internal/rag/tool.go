package rag

import (
	"cyrene/internal/platform/vectorstore"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type PokemonToolResponse struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	Height         int            `json:"height"`
	Weight         int            `json:"weight"`
	BaseExperience int            `json:"base_experience"`
	Types          []string       `json:"types"`
	Abilities      []string       `json:"abilities"`
	Stats          map[string]int `json:"stats"`
	Moves          []string       `json:"moves"`
}

func (s *service) registerTools(g *genkit.Genkit) {
	s.getPokemonTool = s.defineGetPokemonTool(g)
	s.searchTool = s.defineVectorSearchTool(g)
}

func (s *service) defineGetPokemonTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"getPokemon",
		"Fetches Pokemon data by ID or name. Returns stats, types, abilities, moves, height, and weight.",
		func(ctx *ai.ToolContext, input struct {
			ID string `json:"id" jsonschema_description:"Pokemon ID (e.g. '25') or name (e.g. 'pikachu')"`
		}) (*PokemonToolResponse, error) {
			p, err := s.pokemon.GetPokemonByID(ctx, input.ID)
			if err != nil {
				return nil, err
			}
			return toToolResponse(p.Metadata, p.Identifier), nil
		},
	)
}

func toToolResponse(meta map[string]any, name string) *PokemonToolResponse {
	resp := &PokemonToolResponse{
		Name:  name,
		Stats: make(map[string]int),
	}

	if id, ok := meta["id"].(float64); ok {
		resp.ID = int(id)
	}
	if height, ok := meta["height"].(float64); ok {
		resp.Height = int(height)
	}
	if weight, ok := meta["weight"].(float64); ok {
		resp.Weight = int(weight)
	}
	if exp, ok := meta["base_experience"].(float64); ok {
		resp.BaseExperience = int(exp)
	}

	if types, ok := meta["types"].([]any); ok {
		for _, t := range types {
			if tm, ok := t.(map[string]any); ok {
				if typeInfo, ok := tm["type"].(map[string]any); ok {
					if n, ok := typeInfo["name"].(string); ok {
						resp.Types = append(resp.Types, n)
					}
				}
			}
		}
	}

	if abilities, ok := meta["abilities"].([]any); ok {
		for _, a := range abilities {
			if am, ok := a.(map[string]any); ok {
				if abilityInfo, ok := am["ability"].(map[string]any); ok {
					if n, ok := abilityInfo["name"].(string); ok {
						resp.Abilities = append(resp.Abilities, n)
					}
				}
			}
		}
	}

	if stats, ok := meta["stats"].([]any); ok {
		for _, s := range stats {
			if sm, ok := s.(map[string]any); ok {
				if statInfo, ok := sm["stat"].(map[string]any); ok {
					if n, ok := statInfo["name"].(string); ok {
						if baseStat, ok := sm["base_stat"].(float64); ok {
							resp.Stats[n] = int(baseStat)
						}
					}
				}
			}
		}
	}

	if moves, ok := meta["moves"].([]any); ok {
		for _, m := range moves {
			if mm, ok := m.(map[string]any); ok {
				if moveInfo, ok := mm["move"].(map[string]any); ok {
					if n, ok := moveInfo["name"].(string); ok {
						resp.Moves = append(resp.Moves, n)
					}
				}
			}
		}
	}

	return resp
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
			embeddings, err := s.Embed(ctx, s.vectorStore.Dimensions(), input.Query)
			if err != nil {
				return nil, err
			}
			return s.vectorStore.Search(ctx, embeddings[0], input.Limit, nil)
		},
	)
}
