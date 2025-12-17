package rag

import (
	"fmt"
	"strconv"

	"cyrene/internal/ingest"
	"cyrene/internal/platform/vectorstore"
	"cyrene/internal/pokemon"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type FormSearchStats struct {
	HP             int `json:"hp"`
	Attack         int `json:"attack"`
	Defense        int `json:"defense"`
	SpecialAttack  int `json:"specialAttack"`
	SpecialDefense int `json:"specialDefense"`
	Speed          int `json:"speed"`
}

type MoveToolResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Identifier  string   `json:"identifier"`
	Type        string   `json:"type"`
	Category    string   `json:"category"`
	Power       int      `json:"power"`
	Accuracy    int      `json:"accuracy"`
	PP          int      `json:"pp"`
	Priority    int      `json:"priority"`
	Target      string   `json:"target"`
	Effect      string   `json:"effect"`
	Flags       []string `json:"flags"`
}

type AbilityToolResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Generation  int    `json:"generation"`
	Description string `json:"description"`
}

type ArticleToolResponse struct {
	ID         int      `json:"id"`
	Title      string   `json:"title"`
	Subtitle   string   `json:"subtitle,omitempty"`
	Body       string   `json:"body"`
	Identifier string   `json:"identifier"`
	Categories []string `json:"categories"`
}

type FormSearchToolInput struct {
	Query      string   `json:"query,omitempty" jsonschema_description:"Fuzzy search on Pokemon name (e.g., 'char' matches 'Charizard', 'Charmander')"`
	Types      []string `json:"types,omitempty" jsonschema_description:"Filter by type identifiers (e.g., ['fire', 'flying']). Results must have ALL specified types."`
	Abilities  []string `json:"abilities,omitempty" jsonschema_description:"Filter by ability identifiers (e.g., ['levitate', 'intimidate']). Results must have ALL specified abilities."`
	Moves      []string `json:"moves,omitempty" jsonschema_description:"Filter by move identifiers (e.g., ['earthquake', 'fly']). Results must know ALL specified moves."`
	Generation *int     `json:"generation,omitempty" jsonschema_description:"Filter by generation (1-9)"`
	MinHP      *int     `json:"minHp,omitempty" jsonschema_description:"Minimum HP stat"`
	MaxHP      *int     `json:"maxHp,omitempty" jsonschema_description:"Maximum HP stat"`
	MinAttack  *int     `json:"minAttack,omitempty" jsonschema_description:"Minimum Attack stat"`
	MaxAttack  *int     `json:"maxAttack,omitempty" jsonschema_description:"Maximum Attack stat"`
	MinDefense *int     `json:"minDefense,omitempty" jsonschema_description:"Minimum Defense stat"`
	MaxDefense *int     `json:"maxDefense,omitempty" jsonschema_description:"Maximum Defense stat"`
	MinSpAtk   *int     `json:"minSpecialAttack,omitempty" jsonschema_description:"Minimum Special Attack stat"`
	MaxSpAtk   *int     `json:"maxSpecialAttack,omitempty" jsonschema_description:"Maximum Special Attack stat"`
	MinSpDef   *int     `json:"minSpecialDefense,omitempty" jsonschema_description:"Minimum Special Defense stat"`
	MaxSpDef   *int     `json:"maxSpecialDefense,omitempty" jsonschema_description:"Maximum Special Defense stat"`
	MinSpeed   *int     `json:"minSpeed,omitempty" jsonschema_description:"Minimum Speed stat"`
	MaxSpeed   *int     `json:"maxSpeed,omitempty" jsonschema_description:"Maximum Speed stat"`
	MinBST     *int     `json:"minBst,omitempty" jsonschema_description:"Minimum Base Stat Total"`
	MaxBST     *int     `json:"maxBst,omitempty" jsonschema_description:"Maximum Base Stat Total"`
	Limit      int      `json:"limit,omitempty" jsonschema_description:"Max results to return (default 10, max 100)"`
}

type FormSearchSpecies struct {
	Name           string `json:"name"`
	GenderRate     string `json:"genderRate"`
	CatchRate      int    `json:"catchRate"`
	GrowthRate     string `json:"growthRate"`
	BaseFriendship int    `json:"baseFriendship"`
	EggCycle       int    `json:"eggCycle"`
	IsBaby         bool   `json:"isBaby"`
	Classification string `json:"classification"`
}

type FormSearchToolResult struct {
	ID         int                `json:"id"`
	Name       string             `json:"name"`
	FormName   string             `json:"formName"`
	Generation int                `json:"generation"`
	Height     float64            `json:"height"`
	Weight     float64            `json:"weight"`
	Species    FormSearchSpecies  `json:"species"`
	Stats      FormSearchStats    `json:"stats"`
	Types      []string           `json:"types"`
	Abilities  []string           `json:"abilities"`
}

type FormSearchToolResponse struct {
	Results []FormSearchToolResult `json:"results"`
	Total   int                    `json:"total"`
}

func (s *service) registerTools(g *genkit.Genkit) {
	s.searchPokemonTool = s.defineSearchPokemonTool(g)
	s.getMoveTool = s.defineGetMoveTool(g)
	s.getAbilityTool = s.defineGetAbilityTool(g)
	s.getArticleTool = s.defineGetArticleTool(g)
	s.searchTool = s.defineVectorSearchTool(g)
}

func (s *service) defineSearchPokemonTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"searchPokemon",
		`Search and filter Pokemon forms. Returns form data (stats, types, abilities, height, weight) plus species data (growth rate, gender ratio, catch rate, base friendship, egg cycles, classification, baby status). Filters: name (fuzzy), types, abilities, moves, generation, stat ranges (HP/Atk/Def/SpA/SpD/Spe), BST range.`,
		func(ctx *ai.ToolContext, input FormSearchToolInput) (*FormSearchToolResponse, error) {
			limit := input.Limit
			if limit <= 0 {
				limit = 10
			}
			if limit > 100 {
				limit = 100
			}

			types := s.resolveTypes(ctx, input.Types)
			abilities := s.resolveAbilities(ctx, input.Abilities)
			moves := s.resolveMoves(ctx, input.Moves)

			params := pokemon.FormSearchParams{
				Query:             input.Query,
				Types:             types,
				Abilities:         abilities,
				Moves:             moves,
				Generation:        input.Generation,
				MinHP:             input.MinHP,
				MaxHP:             input.MaxHP,
				MinAttack:         input.MinAttack,
				MaxAttack:         input.MaxAttack,
				MinDefense:        input.MinDefense,
				MaxDefense:        input.MaxDefense,
				MinSpecialAttack:  input.MinSpAtk,
				MaxSpecialAttack:  input.MaxSpAtk,
				MinSpecialDefense: input.MinSpDef,
				MaxSpecialDefense: input.MaxSpDef,
				MinSpeed:          input.MinSpeed,
				MaxSpeed:          input.MaxSpeed,
				MinBST:            input.MinBST,
				MaxBST:            input.MaxBST,
				Include:           []string{"stats", "types", "abilities"},
				Limit:             limit,
			}

			resp, err := s.pokemon.SearchForms(ctx, params)
			if err != nil {
				return nil, err
			}

			results := make([]FormSearchToolResult, len(resp.Data))
			for i, form := range resp.Data {
				var types []string
				for _, t := range form.Types {
					types = append(types, t.Type.Name)
				}

				var abilities []string
				for _, a := range form.Abilities {
					name := a.Ability.Name
					if a.IsHidden {
						name += " (Hidden)"
					}
					abilities = append(abilities, name)
				}

				var stats FormSearchStats
				if form.Stats != nil {
					stats = FormSearchStats{
						HP:             form.Stats.HP,
						Attack:         form.Stats.Attack,
						Defense:        form.Stats.Defense,
						SpecialAttack:  form.Stats.SpecialAttack,
						SpecialDefense: form.Stats.SpecialDefense,
						Speed:          form.Stats.Speed,
					}
				}

				species := FormSearchSpecies{
					Name:           form.Species.Name,
					GenderRate:     pokemon.GenderRateDescription(form.Species.GenderRate),
					CatchRate:      form.Species.CatchRate,
					GrowthRate:     pokemon.GrowthRateName(form.Species.GrowthRate),
					BaseFriendship: form.Species.BaseFriendship,
					EggCycle:       form.Species.EggCycle,
					IsBaby:         form.Species.IsBaby,
					Classification: form.Species.Classification.Name,
				}

				results[i] = FormSearchToolResult{
					ID:         form.ID,
					Name:       form.Name,
					FormName:   form.FormName,
					Generation: form.Generation,
					Height:     float64(form.Height) / 10,
					Weight:     float64(form.Weight) / 10,
					Species:    species,
					Stats:      stats,
					Types:      types,
					Abilities:  abilities,
				}
			}

			return &FormSearchToolResponse{
				Results: results,
				Total:   resp.Total,
			}, nil
		},
	)
}

func (s *service) defineGetMoveTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"getMove",
		"Fetches move data by name or ID. Uses fuzzy search to find the move, then returns full details including type, power, accuracy, PP, priority, target, effect, and flags.",
		func(ctx *ai.ToolContext, input struct {
			Query string `json:"query" jsonschema_description:"Move name (e.g. 'thunderbolt', 'earthquake') or ID (e.g. '85')"`
		}) (*MoveToolResponse, error) {
			results, err := s.pokemon.SearchMoves(ctx, input.Query, 1)
			if err != nil {
				return nil, err
			}
			if len(results) == 0 {
				return nil, fmt.Errorf("move not found: %s", input.Query)
			}

			move, err := s.pokemon.GetMoveByID(ctx, strconv.Itoa(results[0].ID))
			if err != nil {
				return nil, err
			}

			var flags []string
			for _, f := range move.Flags {
				flags = append(flags, f.Name)
			}

			effect := move.ShortEffect
			if effect == "" {
				effect = move.Effect
			}

			return &MoveToolResponse{
				ID:         move.ID,
				Name:       move.Name,
				Identifier: move.Identifier,
				Type:       move.Type.Name,
				Category:   move.MoveDamageClass.Name,
				Power:      move.Power,
				Accuracy:   move.Accuracy,
				PP:         move.PowerPoints,
				Priority:   move.Priority,
				Target:     move.MoveTarget.Name,
				Effect:     effect,
				Flags:      flags,
			}, nil
		},
	)
}

func (s *service) defineGetAbilityTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"getAbility",
		"Fetches ability data by name or ID. Uses fuzzy search to find the ability, then returns full details including name, generation, and description.",
		func(ctx *ai.ToolContext, input struct {
			Query string `json:"query" jsonschema_description:"Ability name (e.g. 'static', 'intimidate') or ID (e.g. '9')"`
		}) (*AbilityToolResponse, error) {
			results, err := s.pokemon.SearchAbilities(ctx, input.Query, 1)
			if err != nil {
				return nil, err
			}
			if len(results) == 0 {
				return nil, fmt.Errorf("ability not found: %s", input.Query)
			}

			ability, err := s.pokemon.GetAbilityByID(ctx, strconv.Itoa(results[0].ID))
			if err != nil {
				return nil, err
			}

			description := ability.ShortDescription
			if description == "" {
				description = ability.Description
			}

			return &AbilityToolResponse{
				ID:          ability.ID,
				Name:        ability.Name,
				Identifier:  ability.Identifier,
				Generation:  ability.Generation,
				Description: description,
			}, nil
		},
	)
}

func (s *service) defineGetArticleTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"getArticle",
		"Fetches article data by title or ID. Uses fuzzy search to find the article, then returns full details including title, subtitle, body content, and categories.",
		func(ctx *ai.ToolContext, input struct {
			Query string `json:"query" jsonschema_description:"Article title (e.g. 'getting started', 'team building') or ID (e.g. '1')"`
		}) (*ArticleToolResponse, error) {
			results, err := s.pokemon.SearchArticles(ctx, input.Query, 1)
			if err != nil {
				return nil, err
			}
			if len(results) == 0 {
				return nil, fmt.Errorf("article not found: %s", input.Query)
			}

			article, err := s.pokemon.GetArticleByID(ctx, strconv.Itoa(results[0].ID))
			if err != nil {
				return nil, err
			}

			var categories []string
			for _, c := range article.Categories {
				categories = append(categories, c.Category.Name)
			}

			return &ArticleToolResponse{
				ID:         article.ID,
				Title:      article.Title,
				Subtitle:   article.Subtitle,
				Body:       article.Body,
				Identifier: article.Identifier,
				Categories: categories,
			}, nil
		},
	)
}

func (s *service) defineVectorSearchTool(g *genkit.Genkit) ai.Tool {
	return genkit.DefineTool(
		g,
		"search",
		"Searches the database using semantic similarity. Includes Pokemon forms, moves, abilities, and articles. Use for exploratory queries like finding Pokemon by characteristics, moves by type or effect, abilities by description, or articles by topic. Filter by type to narrow results.",
		func(ctx *ai.ToolContext, input struct {
			Query string                `json:"query" jsonschema_description:"Natural language search query describing what you're looking for"`
			Types []ingest.DocumentType `json:"types,omitempty" jsonschema_description:"Filter by document type. Must be an array, e.g. [\"form\"] or [\"move\", \"ability\"]"`
			Limit int                   `json:"limit" jsonschema_description:"Max results to return (default 5)"`
		}) ([]vectorstore.SearchResult, error) {
			embeddings, err := s.Embed(ctx, s.vectorStore.Dimensions(), input.Query)
			if err != nil {
				return nil, err
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

			return s.vectorStore.Search(ctx, embeddings[0], input.Limit, filter)
		},
	)
}
