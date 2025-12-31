package pokemon

import (
	"fmt"
	"strings"
)

type PaginatedResponse[T any] struct {
	Results []T `json:"results"`
	Total   int `json:"total"`
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
}

type AgentPokemon struct {
	Name            string                `json:"name"`
	Slug            string                `json:"slug"`
	SpeciesName     string                `json:"speciesName"`
	Description     *string               `json:"description,omitempty"`
	Generation      int                   `json:"generation,omitempty"`
	Stats           *AgentStats           `json:"stats,omitempty"`
	EvYield         *AgentEvYield         `json:"evYield,omitempty"`
	Physical        *AgentPhysical        `json:"physical,omitempty"`
	Types           []string              `json:"types,omitempty"`
	Abilities       []AgentPokemonAbility `json:"abilities,omitempty"`
	Moves           []AgentPokemonMove    `json:"moves,omitempty"`
	Drops           []AgentDrop           `json:"drops,omitempty"`
	Breeding        *AgentBreeding        `json:"breeding,omitempty"`
	EggGroups       []string              `json:"eggGroups,omitempty"`
	ExperienceGroup *string               `json:"experienceGroup,omitempty"`
	Labels          []string              `json:"labels,omitempty"`
	Cosmetics       *AgentCosmetics       `json:"cosmetics,omitempty"`
	Hitbox          *AgentHitbox          `json:"hitbox,omitempty"`
	Lighting        *AgentLighting        `json:"lighting,omitempty"`
	Riding          *AgentRiding          `json:"riding,omitempty"`
	Behaviour       *AgentBehaviour       `json:"behaviour,omitempty"`
	Spawns          []AgentSpawn          `json:"spawns,omitempty"`
}

type AgentStats struct {
	HP      int `json:"hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	SpAtk   int `json:"spAtk"`
	SpDef   int `json:"spDef"`
	Speed   int `json:"speed"`
	Total   int `json:"total"`
}

type AgentEvYield struct {
	HP      int `json:"hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	SpAtk   int `json:"spAtk"`
	SpDef   int `json:"spDef"`
	Speed   int `json:"speed"`
}

type AgentPhysical struct {
	Height int `json:"height"`
	Weight int `json:"weight"`
}

type AgentPokemonAbility struct {
	Name string `json:"name"`
	Slot string `json:"slot"`
}

type AgentPokemonMove struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Level  *int   `json:"level"`
}

type AgentDrop struct {
	Item        string `json:"item"`
	Chance      int    `json:"chance,omitempty"`
	QuantityMin int    `json:"quantityMin,omitempty"`
	QuantityMax int    `json:"quantityMax,omitempty"`
}

type AgentBreeding struct {
	EggCycles      int      `json:"eggCycles"`
	BaseFriendship int      `json:"baseFriendship"`
	MaleRatio      *float64 `json:"maleRatio"`
}

type AgentCosmetics struct {
	AspectChoices []string   `json:"aspectChoices"`
	AspectCombos  [][]string `json:"aspectCombos"`
}

type AgentHitbox struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Fixed  bool    `json:"fixed"`
}

type AgentLighting struct {
	LightLevel    int     `json:"lightLevel"`
	LiquidGlowMode *string `json:"liquidGlowMode"`
}

type AgentRiding struct {
	Data any `json:"data"`
}

type AgentBehaviour struct {
	Data any `json:"data"`
}

type AgentSpawn struct {
	Bucket       string                `json:"bucket"`
	PositionType string                `json:"positionType"`
	Weight       float64               `json:"weight"`
	LevelMin     int                   `json:"levelMin"`
	LevelMax     int                   `json:"levelMax"`
	Presets      []string              `json:"presets"`
	Conditions   []AgentSpawnCondition `json:"conditions"`
}

type AgentSpawnCondition struct {
	Type       string             `json:"type"`
	Multiplier *float64           `json:"multiplier,omitempty"`
	Biomes     []string           `json:"biomes,omitempty"`
	BiomeTags  []string           `json:"biomeTags,omitempty"`
	TimeRanges []string           `json:"timeRanges,omitempty"`
	MoonPhases []string           `json:"moonPhases,omitempty"`
	Weather    *AgentSpawnWeather `json:"weather,omitempty"`
	Sky        *AgentSpawnSky     `json:"sky,omitempty"`
	Position   *AgentSpawnPos     `json:"position,omitempty"`
	Lure       *AgentSpawnLure    `json:"lure,omitempty"`
}

type AgentSpawnWeather struct {
	IsRaining    *bool `json:"isRaining,omitempty"`
	IsThundering *bool `json:"isThundering,omitempty"`
}

type AgentSpawnSky struct {
	CanSeeSky   *bool `json:"canSeeSky,omitempty"`
	MinSkyLight *int  `json:"minSkyLight,omitempty"`
	MaxSkyLight *int  `json:"maxSkyLight,omitempty"`
}

type AgentSpawnPos struct {
	MinY *int `json:"minY,omitempty"`
	MaxY *int `json:"maxY,omitempty"`
}

type AgentSpawnLure struct {
	MinLureLevel *int `json:"minLureLevel,omitempty"`
	MaxLureLevel *int `json:"maxLureLevel,omitempty"`
}

type AgentMove struct {
	Name      string            `json:"name"`
	Slug      string            `json:"slug"`
	Type      string            `json:"type"`
	Category  string            `json:"category"`
	Power     *int              `json:"power"`
	Accuracy  *int              `json:"accuracy"`
	PP        int               `json:"pp"`
	Priority  int               `json:"priority"`
	Target    *string           `json:"target,omitempty"`
	ShortDesc *string           `json:"shortDesc,omitempty"`
	Desc      *string           `json:"desc,omitempty"`
	Flags     []string          `json:"flags,omitempty"`
	Boosts    []AgentMoveBoost  `json:"boosts,omitempty"`
	Effects   []AgentMoveEffect `json:"effects,omitempty"`
	ZData     *AgentZData       `json:"zData,omitempty"`
}

type AgentMoveBoost struct {
	Stat   string `json:"stat"`
	Stages int    `json:"stages"`
	IsSelf bool   `json:"isSelf"`
}

type AgentMoveEffect struct {
	Effect    string  `json:"effect"`
	Chance    int     `json:"chance"`
	IsSelf    bool    `json:"isSelf"`
	Condition *string `json:"condition,omitempty"`
}

type AgentZData struct {
	ZPower       *int    `json:"zPower"`
	ZEffect      *string `json:"zEffect"`
	ZCrystal     *string `json:"zCrystal"`
	IsZExclusive bool    `json:"isZExclusive"`
}

type AgentAbility struct {
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	ShortDesc *string  `json:"shortDesc,omitempty"`
	Desc      *string  `json:"desc,omitempty"`
	Flags     []string `json:"flags,omitempty"`
}

type AgentItem struct {
	Name      string           `json:"name"`
	Slug      string           `json:"slug"`
	ShortDesc *string          `json:"shortDesc,omitempty"`
	Desc      *string          `json:"desc,omitempty"`
	Boosts    []AgentItemBoost `json:"boosts,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
}

type AgentItemBoost struct {
	Stat   string `json:"stat"`
	Stages int    `json:"stages"`
}

type AgentArticleSearch struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Subtitle    *string  `json:"subtitle,omitempty"`
	Description *string  `json:"description,omitempty"`
	Body        string   `json:"body,omitempty"`
	Author      *string  `json:"author,omitempty"`
	Categories  []string `json:"categories,omitempty"`
}

type AgentArticle struct {
	ID          int     `json:"id"`
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Subtitle    *string `json:"subtitle"`
	Description *string `json:"description"`
	Body        string  `json:"body"`
	Author      *string `json:"author"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type AgentPokemonParams struct {
	Names              []string
	Types              []string
	Abilities          []string
	Moves              []string
	EggGroups          []string
	Labels             []string
	Generation         []int
	IncludeDescription bool
	IncludeGeneration  bool
	IncludeStats       bool
	IncludeEvYield     bool
	IncludePhysical    bool
	IncludeTypes       bool
	IncludeAbilities   bool
	IncludeMoves       bool
	IncludeDrops       bool
	IncludeBreeding    bool
	IncludeEggGroups   bool
	IncludeExpGroup    bool
	IncludeLabels      bool
	IncludeAspects     bool
	IncludeHitboxes    bool
	IncludeLighting    bool
	IncludeRiding      bool
	IncludeBehaviour   bool
	IncludeSpawns      bool
	Limit              int
	Offset             int
}

type AgentMoveParams struct {
	Names              []string
	Types              []string
	Categories         []string
	IncludeDescription bool
	IncludeFlags       bool
	IncludeBoosts      bool
	IncludeEffects     bool
	IncludeZData       bool
	Limit              int
	Offset             int
}

type AgentAbilityParams struct {
	Names              []string
	IncludeDescription bool
	IncludeFlags       bool
	Limit              int
	Offset             int
}

type AgentItemParams struct {
	Names              []string
	Tags               []string
	IncludeDescription bool
	IncludeBoosts      bool
	IncludeTags        bool
	Limit              int
	Offset             int
}

type AgentArticleParams struct {
	Titles            []string
	Categories        []string
	IncludeBody       bool
	IncludeCategories bool
	Limit             int
	Offset            int
}

// Get-one response types for ingestion

type NamedResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Species struct {
	ID    int           `json:"id"`
	Name  string        `json:"name"`
	Slug  string        `json:"slug"`
	Forms []SpeciesForm `json:"forms"`
}

type SpeciesForm struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

type Pokemon struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	Slug            string           `json:"slug"`
	Description     *string          `json:"description"`
	Generation      int              `json:"generation"`
	BaseFriendship  int              `json:"baseFriendship"`
	CatchRate       int              `json:"catchRate"`
	EggCycles       int              `json:"eggCycles"`
	ExperienceGroup *ExperienceGroup `json:"experienceGroup"`
	MaleRatio       *float64         `json:"maleRatio"`
	EggGroups       []NamedResource  `json:"eggGroups"`
	Form            *PokemonForm     `json:"form"`
}

type ExperienceGroup struct {
	ID      int    `json:"id"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Formula string `json:"formula"`
}

type PokemonForm struct {
	ID                 int                  `json:"id"`
	Name               string               `json:"name"`
	FullName           string               `json:"fullName"`
	Slug               string               `json:"slug"`
	Description        *string              `json:"description"`
	Generation         *int                 `json:"generation"`
	Height             int                  `json:"height"`
	Weight             int                  `json:"weight"`
	BaseHP             int                  `json:"baseHp"`
	BaseAttack         int                  `json:"baseAttack"`
	BaseDefence        int                  `json:"baseDefence"`
	BaseSpecialAttack  int                  `json:"baseSpecialAttack"`
	BaseSpecialDefence int                  `json:"baseSpecialDefence"`
	BaseSpeed          int                  `json:"baseSpeed"`
	Types              []PokemonFormType    `json:"types"`
	Abilities          []PokemonFormAbility `json:"abilities"`
	Moves              []PokemonFormMove    `json:"moves"`
	Labels             []NamedResource      `json:"labels"`
}

type PokemonFormType struct {
	Type NamedResource `json:"type"`
	Slot int           `json:"slot"`
}

type PokemonFormAbility struct {
	Ability NamedResource `json:"ability"`
	Slot    NamedResource `json:"slot"`
}

type PokemonFormMove struct {
	Move   NamedResource `json:"move"`
	Method NamedResource `json:"method"`
	Level  *int          `json:"level"`
}

type Move struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Desc      *string         `json:"desc"`
	ShortDesc *string         `json:"shortDesc"`
	Type      NamedResource   `json:"type"`
	Category  NamedResource   `json:"category"`
	Target    *NamedResource  `json:"target"`
	Power     *int            `json:"power"`
	Accuracy  *int            `json:"accuracy"`
	PP        int             `json:"pp"`
	Priority  int             `json:"priority"`
	Flags     []NamedResource `json:"flags"`
}

type Ability struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Desc      *string         `json:"desc"`
	ShortDesc *string         `json:"shortDesc"`
	Flags     []NamedResource `json:"flags"`
}

type Item struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Desc      *string         `json:"desc"`
	ShortDesc *string         `json:"shortDesc"`
	Boosts    []ItemBoost     `json:"boosts"`
	Flags     []NamedResource `json:"flags"`
	Tags      []NamedResource `json:"tags"`
}

type ItemBoost struct {
	Stat   NamedResource `json:"stat"`
	Stages int           `json:"stages"`
}

type Article struct {
	ID          int               `json:"id"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Subtitle    *string           `json:"subtitle"`
	Description *string           `json:"description"`
	Body        string            `json:"body"`
	Author      *string           `json:"author"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
	Categories  []ArticleCategory `json:"categories"`
}

type ArticleCategory struct {
	ID          int     `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// EmbeddingText methods for ingestion

func (p *Pokemon) EmbeddingText() string {
	var sb strings.Builder

	if p.Form != nil {
		sb.WriteString(fmt.Sprintf("Pokemon: %s\n", p.Form.FullName))
	} else {
		sb.WriteString(fmt.Sprintf("Pokemon: %s\n", p.Name))
	}
	sb.WriteString(fmt.Sprintf("Generation: %d\n", p.Generation))

	if p.Form != nil {
		if len(p.Form.Types) > 0 {
			var types []string
			for _, t := range p.Form.Types {
				types = append(types, t.Type.Name)
			}
			sb.WriteString(fmt.Sprintf("Types: %s\n", strings.Join(types, ", ")))
		}

		if len(p.Form.Abilities) > 0 {
			var abilities []string
			for _, a := range p.Form.Abilities {
				name := a.Ability.Name
				if a.Slot.Slug == "hidden" {
					name += " (Hidden)"
				}
				abilities = append(abilities, name)
			}
			sb.WriteString(fmt.Sprintf("Abilities: %s\n", strings.Join(abilities, ", ")))
		}

		sb.WriteString(fmt.Sprintf("Stats: HP %d, Atk %d, Def %d, SpA %d, SpD %d, Spe %d\n",
			p.Form.BaseHP, p.Form.BaseAttack, p.Form.BaseDefence,
			p.Form.BaseSpecialAttack, p.Form.BaseSpecialDefence, p.Form.BaseSpeed))

		if len(p.Form.Moves) > 0 {
			var moves []string
			for i, m := range p.Form.Moves {
				if i >= 20 {
					break
				}
				moves = append(moves, m.Move.Name)
			}
			sb.WriteString(fmt.Sprintf("Moves: %s\n", strings.Join(moves, ", ")))
		}

		if len(p.Form.Labels) > 0 {
			var labels []string
			for _, l := range p.Form.Labels {
				labels = append(labels, l.Name)
			}
			sb.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(labels, ", ")))
		}
	}

	if len(p.EggGroups) > 0 {
		var groups []string
		for _, g := range p.EggGroups {
			groups = append(groups, g.Name)
		}
		sb.WriteString(fmt.Sprintf("Egg Groups: %s\n", strings.Join(groups, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Catch Rate: %d\n", p.CatchRate))
	sb.WriteString(fmt.Sprintf("Base Friendship: %d\n", p.BaseFriendship))

	return sb.String()
}

func (m *Move) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Move: %s\n", m.Name))
	sb.WriteString(fmt.Sprintf("Type: %s\n", m.Type.Name))
	sb.WriteString(fmt.Sprintf("Category: %s\n", m.Category.Name))

	if m.Power != nil {
		sb.WriteString(fmt.Sprintf("Power: %d\n", *m.Power))
	}
	if m.Accuracy != nil {
		sb.WriteString(fmt.Sprintf("Accuracy: %d\n", *m.Accuracy))
	}
	sb.WriteString(fmt.Sprintf("PP: %d\n", m.PP))
	sb.WriteString(fmt.Sprintf("Priority: %d\n", m.Priority))

	if m.Target != nil {
		sb.WriteString(fmt.Sprintf("Target: %s\n", m.Target.Name))
	}

	if m.ShortDesc != nil && *m.ShortDesc != "" {
		sb.WriteString(fmt.Sprintf("Effect: %s\n", *m.ShortDesc))
	} else if m.Desc != nil && *m.Desc != "" {
		sb.WriteString(fmt.Sprintf("Effect: %s\n", *m.Desc))
	}

	if len(m.Flags) > 0 {
		var flags []string
		for _, f := range m.Flags {
			flags = append(flags, f.Name)
		}
		sb.WriteString(fmt.Sprintf("Flags: %s\n", strings.Join(flags, ", ")))
	}

	return sb.String()
}

func (a *Ability) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Ability: %s\n", a.Name))

	if a.ShortDesc != nil && *a.ShortDesc != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", *a.ShortDesc))
	} else if a.Desc != nil && *a.Desc != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", *a.Desc))
	}

	if len(a.Flags) > 0 {
		var flags []string
		for _, f := range a.Flags {
			flags = append(flags, f.Name)
		}
		sb.WriteString(fmt.Sprintf("Flags: %s\n", strings.Join(flags, ", ")))
	}

	return sb.String()
}

func (i *Item) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Item: %s\n", i.Name))

	if i.ShortDesc != nil && *i.ShortDesc != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", *i.ShortDesc))
	} else if i.Desc != nil && *i.Desc != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", *i.Desc))
	}

	if len(i.Boosts) > 0 {
		var boosts []string
		for _, b := range i.Boosts {
			boosts = append(boosts, fmt.Sprintf("%s %+d", b.Stat.Name, b.Stages))
		}
		sb.WriteString(fmt.Sprintf("Boosts: %s\n", strings.Join(boosts, ", ")))
	}

	if len(i.Tags) > 0 {
		var tags []string
		for _, t := range i.Tags {
			tags = append(tags, t.Name)
		}
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(tags, ", ")))
	}

	return sb.String()
}

func (a *Article) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Article: %s\n", a.Title))

	if a.Subtitle != nil && *a.Subtitle != "" {
		sb.WriteString(fmt.Sprintf("Subtitle: %s\n", *a.Subtitle))
	}

	if a.Author != nil && *a.Author != "" {
		sb.WriteString(fmt.Sprintf("Author: %s\n", *a.Author))
	}

	if len(a.Categories) > 0 {
		var cats []string
		for _, c := range a.Categories {
			cats = append(cats, c.Name)
		}
		sb.WriteString(fmt.Sprintf("Categories: %s\n", strings.Join(cats, ", ")))
	}

	if a.Description != nil && *a.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", *a.Description))
	}

	sb.WriteString(fmt.Sprintf("\n%s\n", a.Body))

	return sb.String()
}
