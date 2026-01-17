package rag

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"cyrene/internal/ingest"
	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const maxResponseBytes = 50000

func checkResponseSize(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return fmt.Errorf("response too large (%d bytes, max %d), narrow your search or reduce included fields", len(data), maxResponseBytes)
	}
	return nil
}

type SearchToolResponse struct {
	Results []vectorstore.SearchResult `json:"results,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

type PokemonToolResponse struct {
	Results []pokemon.AgentPokemon `json:"results,omitempty"`
	Total   int                    `json:"total,omitempty"`
	Limit   int                    `json:"limit,omitempty"`
	Offset  int                    `json:"offset,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type MoveToolResponse struct {
	Results []pokemon.AgentMove `json:"results,omitempty"`
	Total   int                 `json:"total,omitempty"`
	Limit   int                 `json:"limit,omitempty"`
	Offset  int                 `json:"offset,omitempty"`
	Error   string              `json:"error,omitempty"`
}

type AbilityToolResponse struct {
	Results []pokemon.AgentAbility `json:"results,omitempty"`
	Total   int                    `json:"total,omitempty"`
	Limit   int                    `json:"limit,omitempty"`
	Offset  int                    `json:"offset,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type ItemToolResponse struct {
	Results []pokemon.AgentItem `json:"results,omitempty"`
	Total   int                 `json:"total,omitempty"`
	Limit   int                 `json:"limit,omitempty"`
	Offset  int                 `json:"offset,omitempty"`
	Error   string              `json:"error,omitempty"`
}

type ArticleSearchToolResponse struct {
	Results []pokemon.AgentArticleSearch `json:"results,omitempty"`
	Total   int                          `json:"total,omitempty"`
	Limit   int                          `json:"limit,omitempty"`
	Offset  int                          `json:"offset,omitempty"`
	Error   string                       `json:"error,omitempty"`
}

type ArticleToolResponse struct {
	Article *pokemon.AgentArticle `json:"article,omitempty"`
	Error   string                `json:"error,omitempty"`
}

func (s *service) registerTools(g *genkit.Genkit) {
	s.searchPokemonTool = s.defineSearchPokemonTool(g)
	s.searchMovesTool = s.defineSearchMovesTool(g)
	s.searchAbilitiesTool = s.defineSearchAbilitiesTool(g)
	s.searchItemsTool = s.defineSearchItemsTool(g)
	s.searchArticlesTool = s.defineSearchArticlesTool(g)
	s.getArticleTool = s.defineGetArticleTool(g)
	s.searchTool = s.defineVectorSearchTool(g)
}

func (s *service) defineSearchPokemonTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchPokemon",
		`Search Pokemon by name, type, ability, move, egg group, label, or generation. Returns form data. All filters use fuzzy matching. Use include flags to control response size - only request what you need.`,
		func(ctx *ai.ToolContext, input struct {
			Names              []string `json:"names,omitempty" jsonschema_description:"Fuzzy match Pokemon names (e.g., ['pikachu', 'char'])"`
			Types              []string `json:"types,omitempty" jsonschema_description:"Filter by type names (e.g., ['fire', 'water']). Returns Pokemon with any matching type."`
			Abilities          []string `json:"abilities,omitempty" jsonschema_description:"Filter by ability names (e.g., ['levitate']). Returns Pokemon with any matching ability."`
			Moves              []string `json:"moves,omitempty" jsonschema_description:"Filter by move names (e.g., ['thunderbolt']). Returns Pokemon that can learn any matching move."`
			EggGroups          []string `json:"eggGroups,omitempty" jsonschema_description:"Filter by egg group names (e.g., ['dragon', 'monster'])."`
			Labels             []string `json:"labels,omitempty" jsonschema_description:"Filter by labels (e.g., ['legendary', 'starter'])."`
			DropsItems         []string `json:"dropsItems,omitempty" jsonschema_description:"Filter by dropped item names (e.g., ['oran berry', 'potion']). Returns Pokemon that drop any matching item."`
			Generation         []int    `json:"generation,omitempty" jsonschema_description:"Filter by generation numbers (e.g., [1, 2, 3])."`
			IncludeDescription bool     `json:"includeDescription,omitempty" jsonschema_description:"Include Pokemon description/flavor text."`
			IncludeGeneration  bool     `json:"includeGeneration,omitempty" jsonschema_description:"Include generation number."`
			IncludeStats       bool     `json:"includeStats,omitempty" jsonschema_description:"Include base stats (HP, Attack, Defense, etc.)."`
			IncludeEvYield     bool     `json:"includeEvYield,omitempty" jsonschema_description:"Include EV yield data."`
			IncludePhysical    bool     `json:"includePhysical,omitempty" jsonschema_description:"Include physical attributes (height, weight)."`
			IncludeTypes       bool     `json:"includeTypes,omitempty" jsonschema_description:"Include type information."`
			IncludeAbilities   bool     `json:"includeAbilities,omitempty" jsonschema_description:"Include ability information."`
			IncludeMoves       bool     `json:"includeMoves,omitempty" jsonschema_description:"Include learnable moves."`
			IncludeDrops       bool     `json:"includeDrops,omitempty" jsonschema_description:"Include item drop information."`
			IncludeBreeding    bool     `json:"includeBreeding,omitempty" jsonschema_description:"Include breeding data (egg cycles, gender ratio)."`
			IncludeEggGroups   bool     `json:"includeEggGroups,omitempty" jsonschema_description:"Include egg group information."`
			IncludeExpGroup    bool     `json:"includeExpGroup,omitempty" jsonschema_description:"Include experience group."`
			IncludeLabels      bool     `json:"includeLabels,omitempty" jsonschema_description:"Include labels (legendary, starter, etc.)."`
			IncludeAspects     bool     `json:"includeAspects,omitempty" jsonschema_description:"Include aspect/variant information."`
			IncludeHitboxes    bool     `json:"includeHitboxes,omitempty" jsonschema_description:"Include hitbox dimensions."`
			IncludeLighting    bool     `json:"includeLighting,omitempty" jsonschema_description:"Include lighting/glow effects."`
			IncludeRiding      bool     `json:"includeRiding,omitempty" jsonschema_description:"Include riding/mount data."`
			IncludeBehaviour   bool     `json:"includeBehaviour,omitempty" jsonschema_description:"Include AI behaviour patterns."`
			IncludeSpawns      bool     `json:"includeSpawns,omitempty" jsonschema_description:"Include spawn location and condition data."`
			Limit              int      `json:"limit,omitempty" jsonschema_description:"Max results to return (default 5, max 100)"`
			Offset             int      `json:"offset,omitempty" jsonschema_description:"Number of results to skip for pagination."`
		}) (*PokemonToolResponse, error) {
			slog.Info("tool searchPokemon",
				"names", input.Names,
				"types", input.Types,
				"abilities", input.Abilities,
				"moves", input.Moves,
				"eggGroups", input.EggGroups,
				"labels", input.Labels,
				"dropsItems", input.DropsItems,
				"generation", input.Generation,
				"limit", input.Limit,
				"offset", input.Offset,
			)

			limit := input.Limit
			if limit <= 0 {
				limit = 5
			}
			if limit > 100 {
				limit = 100
			}

			params := pokemon.AgentPokemonParams{
				Names:              input.Names,
				Types:              input.Types,
				Abilities:          input.Abilities,
				Moves:              input.Moves,
				EggGroups:          input.EggGroups,
				Labels:             input.Labels,
				DropsItems:         input.DropsItems,
				Generation:         input.Generation,
				IncludeDescription: input.IncludeDescription,
				IncludeGeneration:  input.IncludeGeneration,
				IncludeStats:       input.IncludeStats,
				IncludeEvYield:     input.IncludeEvYield,
				IncludePhysical:    input.IncludePhysical,
				IncludeTypes:       input.IncludeTypes,
				IncludeAbilities:   input.IncludeAbilities,
				IncludeMoves:       input.IncludeMoves,
				IncludeDrops:       input.IncludeDrops,
				IncludeBreeding:    input.IncludeBreeding,
				IncludeEggGroups:   input.IncludeEggGroups,
				IncludeExpGroup:    input.IncludeExpGroup,
				IncludeLabels:      input.IncludeLabels,
				IncludeAspects:     input.IncludeAspects,
				IncludeHitboxes:    input.IncludeHitboxes,
				IncludeLighting:    input.IncludeLighting,
				IncludeRiding:      input.IncludeRiding,
				IncludeBehaviour:   input.IncludeBehaviour,
				IncludeSpawns:      input.IncludeSpawns,
				Limit:              limit,
				Offset:             input.Offset,
			}

			resp, err := s.pokemon.SearchPokemon(ctx, params)
			if err != nil {
				slog.Error("tool searchPokemon failed", "error", err)
				return &PokemonToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool searchPokemon completed", "results", len(resp.Results), "total", resp.Total)
			result := &PokemonToolResponse{
				Results: resp.Results,
				Total:   resp.Total,
				Limit:   resp.Limit,
				Offset:  resp.Offset,
			}
			if err := checkResponseSize(result); err != nil {
				return &PokemonToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}

func (s *service) defineSearchMovesTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchMoves",
		`Search moves by name, type, or category. Returns move data. All filters use fuzzy matching. Use include flags to control response size.`,
		func(ctx *ai.ToolContext, input struct {
			Names              []string `json:"names,omitempty" jsonschema_description:"Fuzzy match move names (e.g., ['thunderbolt', 'earthquake'])"`
			Types              []string `json:"types,omitempty" jsonschema_description:"Filter by type names (e.g., ['fire', 'water'])"`
			Categories         []string `json:"categories,omitempty" jsonschema_description:"Filter by category (e.g., ['Physical', 'Special', 'Status'])"`
			IncludeDescription bool     `json:"includeDescription,omitempty" jsonschema_description:"Include move description/effect text."`
			IncludeFlags       bool     `json:"includeFlags,omitempty" jsonschema_description:"Include move flags (contact, sound-based, etc.)."`
			IncludeBoosts      bool     `json:"includeBoosts,omitempty" jsonschema_description:"Include stat boost information."`
			IncludeEffects     bool     `json:"includeEffects,omitempty" jsonschema_description:"Include secondary effect data."`
			IncludeZData       bool     `json:"includeZData,omitempty" jsonschema_description:"Include Z-Move data."`
			Limit              int      `json:"limit,omitempty" jsonschema_description:"Max results to return (default 5, max 100)"`
			Offset             int      `json:"offset,omitempty" jsonschema_description:"Number of results to skip for pagination."`
		}) (*MoveToolResponse, error) {
			slog.Info("tool searchMoves",
				"names", input.Names,
				"types", input.Types,
				"categories", input.Categories,
				"limit", input.Limit,
				"offset", input.Offset,
			)

			limit := input.Limit
			if limit <= 0 {
				limit = 5
			}
			if limit > 100 {
				limit = 100
			}

			params := pokemon.AgentMoveParams{
				Names:              input.Names,
				Types:              input.Types,
				Categories:         input.Categories,
				IncludeDescription: input.IncludeDescription,
				IncludeFlags:       input.IncludeFlags,
				IncludeBoosts:      input.IncludeBoosts,
				IncludeEffects:     input.IncludeEffects,
				IncludeZData:       input.IncludeZData,
				Limit:              limit,
				Offset:             input.Offset,
			}

			resp, err := s.pokemon.SearchMoves(ctx, params)
			if err != nil {
				slog.Error("tool searchMoves failed", "error", err)
				return &MoveToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool searchMoves completed", "results", len(resp.Results), "total", resp.Total)
			result := &MoveToolResponse{
				Results: resp.Results,
				Total:   resp.Total,
				Limit:   resp.Limit,
				Offset:  resp.Offset,
			}
			if err := checkResponseSize(result); err != nil {
				return &MoveToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}

func (s *service) defineSearchAbilitiesTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchAbilities",
		`Search abilities by name. Returns ability data. Uses fuzzy matching. Use include flags to control response size.`,
		func(ctx *ai.ToolContext, input struct {
			Names              []string `json:"names,omitempty" jsonschema_description:"Fuzzy match ability names (e.g., ['levitate', 'intimidate'])"`
			IncludeDescription bool     `json:"includeDescription,omitempty" jsonschema_description:"Include ability description/effect text."`
			IncludeFlags       bool     `json:"includeFlags,omitempty" jsonschema_description:"Include ability flags."`
			Limit              int      `json:"limit,omitempty" jsonschema_description:"Max results to return (default 5, max 100)"`
			Offset             int      `json:"offset,omitempty" jsonschema_description:"Number of results to skip for pagination."`
		}) (*AbilityToolResponse, error) {
			slog.Info("tool searchAbilities", "names", input.Names, "limit", input.Limit, "offset", input.Offset)

			limit := input.Limit
			if limit <= 0 {
				limit = 5
			}
			if limit > 100 {
				limit = 100
			}

			params := pokemon.AgentAbilityParams{
				Names:              input.Names,
				IncludeDescription: input.IncludeDescription,
				IncludeFlags:       input.IncludeFlags,
				Limit:              limit,
				Offset:             input.Offset,
			}

			resp, err := s.pokemon.SearchAbilities(ctx, params)
			if err != nil {
				slog.Error("tool searchAbilities failed", "error", err)
				return &AbilityToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool searchAbilities completed", "results", len(resp.Results), "total", resp.Total)
			result := &AbilityToolResponse{
				Results: resp.Results,
				Total:   resp.Total,
				Limit:   resp.Limit,
				Offset:  resp.Offset,
			}
			if err := checkResponseSize(result); err != nil {
				return &AbilityToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}

func (s *service) defineSearchItemsTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchItems",
		`Search items by name or tag. Returns item data. Uses fuzzy matching. Use include flags to control response size.`,
		func(ctx *ai.ToolContext, input struct {
			Names              []string `json:"names,omitempty" jsonschema_description:"Fuzzy match item names (e.g., ['potion', 'pokeball'])"`
			Tags               []string `json:"tags,omitempty" jsonschema_description:"Filter by tag names (e.g., ['medicine', 'ball'])"`
			IncludeDescription bool     `json:"includeDescription,omitempty" jsonschema_description:"Include item description text."`
			IncludeBoosts      bool     `json:"includeBoosts,omitempty" jsonschema_description:"Include stat boost information."`
			IncludeTags        bool     `json:"includeTags,omitempty" jsonschema_description:"Include item tags."`
			IncludeRecipes     bool     `json:"includeRecipes,omitempty" jsonschema_description:"Include crafting recipes that produce this item. Recipes contain inputs with slot positions (0-8 for 3x3 crafting grid, left-to-right, top-to-bottom) and slotType indicating input role: input (smelting/stonecutting), template (smithing), base (smithing armor), addition (smithing material)."`
			Limit              int      `json:"limit,omitempty" jsonschema_description:"Max results to return (default 5, max 100)"`
			Offset             int      `json:"offset,omitempty" jsonschema_description:"Number of results to skip for pagination."`
		}) (*ItemToolResponse, error) {
			slog.Info("tool searchItems", "names", input.Names, "tags", input.Tags, "limit", input.Limit, "offset", input.Offset)

			limit := input.Limit
			if limit <= 0 {
				limit = 5
			}
			if limit > 100 {
				limit = 100
			}

			params := pokemon.AgentItemParams{
				Names:              input.Names,
				Tags:               input.Tags,
				IncludeDescription: input.IncludeDescription,
				IncludeBoosts:      input.IncludeBoosts,
				IncludeTags:        input.IncludeTags,
				IncludeRecipes:     input.IncludeRecipes,
				Limit:              limit,
				Offset:             input.Offset,
			}

			resp, err := s.pokemon.SearchItems(ctx, params)
			if err != nil {
				slog.Error("tool searchItems failed", "error", err)
				return &ItemToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool searchItems completed", "results", len(resp.Results), "total", resp.Total)
			result := &ItemToolResponse{
				Results: resp.Results,
				Total:   resp.Total,
				Limit:   resp.Limit,
				Offset:  resp.Offset,
			}
			if err := checkResponseSize(result); err != nil {
				return &ItemToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}

func (s *service) defineSearchArticlesTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchArticles",
		`Search articles by title or category. Returns article metadata. Use include flags to control response size. Use getArticle to fetch full article content by slug.`,
		func(ctx *ai.ToolContext, input struct {
			Titles            []string `json:"titles,omitempty" jsonschema_description:"Fuzzy match article titles (e.g., ['breeding', 'ev training'])"`
			Categories        []string `json:"categories,omitempty" jsonschema_description:"Filter by category names (e.g., ['guide', 'tutorial'])"`
			IncludeContent    bool     `json:"includeContent,omitempty" jsonschema_description:"Include full article content."`
			IncludeCategories bool     `json:"includeCategories,omitempty" jsonschema_description:"Include article categories."`
			Limit             int      `json:"limit,omitempty" jsonschema_description:"Max results to return (default 5, max 100)"`
			Offset            int      `json:"offset,omitempty" jsonschema_description:"Number of results to skip for pagination."`
		}) (*ArticleSearchToolResponse, error) {
			slog.Info("tool searchArticles", "titles", input.Titles, "categories", input.Categories, "limit", input.Limit, "offset", input.Offset)

			limit := input.Limit
			if limit <= 0 {
				limit = 5
			}
			if limit > 100 {
				limit = 100
			}

			params := pokemon.AgentArticleParams{
				Titles:            input.Titles,
				Categories:        input.Categories,
				IncludeContent:    input.IncludeContent,
				IncludeCategories: input.IncludeCategories,
				Limit:             limit,
				Offset:            input.Offset,
			}

			resp, err := s.pokemon.SearchArticles(ctx, params)
			if err != nil {
				slog.Error("tool searchArticles failed", "error", err)
				return &ArticleSearchToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool searchArticles completed", "results", len(resp.Results), "total", resp.Total)
			result := &ArticleSearchToolResponse{
				Results: resp.Results,
				Total:   resp.Total,
				Limit:   resp.Limit,
				Offset:  resp.Offset,
			}
			if err := checkResponseSize(result); err != nil {
				return &ArticleSearchToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}

func (s *service) defineGetArticleTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"getArticle",
		`Fetch a specific article by its slug. Use this after searchArticles to get the full article content.`,
		func(ctx *ai.ToolContext, input struct {
			Slug string `json:"slug" jsonschema_description:"Article slug from search results (e.g., 'getting-started', 'breeding-guide')"`
		}) (*ArticleToolResponse, error) {
			slog.Info("tool getArticle", "slug", input.Slug)

			if input.Slug == "" {
				return &ArticleToolResponse{Error: "slug is required"}, nil
			}

			article, err := s.pokemon.GetArticleBySlug(ctx, input.Slug)
			if err != nil {
				slog.Error("tool getArticle failed", "error", err)
				return &ArticleToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool getArticle completed", "title", article.Title)
			result := &ArticleToolResponse{Article: article}
			if err := checkResponseSize(result); err != nil {
				return &ArticleToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}

func (s *service) defineVectorSearchTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"search",
		"Semantic search for text-based content. Best for: articles (guides, strategies, game mechanics, tier lists), move effects by concept (e.g. 'moves that cause burn'), and ability effects by description. Use for general knowledge questions or discovering relevant articles. NOT for Pokemon stat/type/ability filtering - use searchPokemon tool instead which has structured filters for accurate queries. Filter by type (move/ability/article) to narrow results.",
		func(ctx *ai.ToolContext, input struct {
			Query string                `json:"query" jsonschema_description:"Natural language search query describing what you're looking for"`
			Types []ingest.DocumentType `json:"types,omitempty" jsonschema_description:"Filter by document type: form (Pokemon data), move, ability, item, article. Must be an array, e.g. [\"form\"] or [\"move\", \"ability\"]."`
			Limit int                   `json:"limit" jsonschema_description:"Max results to return (default 5)"`
		}) (*SearchToolResponse, error) {
			slog.Info("tool search", "query", input.Query, "types", input.Types, "limit", input.Limit)

			embeddings, err := s.Embed(ctx, s.vectorStore.Dimensions(), input.Query)
			if err != nil {
				slog.Error("tool search embed failed", "error", err)
				return &SearchToolResponse{Error: err.Error()}, nil
			}

			var filter *vectorstore.Filter
			if len(input.Types) > 0 {
				filters := make([]vectorstore.StringFilter, len(input.Types))
				for i, t := range input.Types {
					filters[i] = vectorstore.StringFilter{
						Field: "type",
						Value: string(t),
						Op:    vectorstore.FilterOR,
					}
				}
				filter = &vectorstore.Filter{StringFilters: filters}
			}

			results, err := s.vectorStore.Search(ctx, embeddings[0], input.Limit, filter)
			if err != nil {
				slog.Error("tool search failed", "error", err)
				return &SearchToolResponse{Error: err.Error()}, nil
			}

			slog.Info("tool search completed", "results", len(results))
			result := &SearchToolResponse{Results: results}
			if err := checkResponseSize(result); err != nil {
				return &SearchToolResponse{Error: err.Error()}, nil
			}
			return result, nil
		},
	)
}
