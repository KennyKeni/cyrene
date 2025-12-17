package pokemon

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Classification struct {
	Name string `json:"name"`
}

type EggGroup struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

type Stats struct {
	HP             int `json:"hp"`
	Attack         int `json:"attack"`
	Defense        int `json:"defense"`
	SpecialAttack  int `json:"specialAttack"`
	SpecialDefense int `json:"specialDefense"`
	Speed          int `json:"speed"`
}

type AttributeType struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Order      int    `json:"order,omitempty"`
}

type Attribute struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Identifier    string        `json:"identifier"`
	AttributeType AttributeType `json:"attributeType"`
}

type Component struct {
	Attribute Attribute `json:"attribute"`
}

type AbilityInfo struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Identifier       string `json:"identifier"`
	Generation       int    `json:"generation"`
	Description      string `json:"description,omitempty"`
	ShortDescription string `json:"shortDescription,omitempty"`
}

type FormAbility struct {
	Slot     int         `json:"slot"`
	IsHidden bool        `json:"isHidden"`
	Ability  AbilityInfo `json:"ability"`
}

type TypeInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Generation int    `json:"generation"`
}

type FormType struct {
	Slot int      `json:"slot"`
	Type TypeInfo `json:"type"`
}

type MoveTarget struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

type MoveDamageClass struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

type MoveInfo struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	Identifier      string           `json:"identifier"`
	Generation      int              `json:"generation"`
	Power           int              `json:"power"`
	PowerPoints     int              `json:"powerPoints"`
	Accuracy        int              `json:"accuracy"`
	Priority        int              `json:"priority"`
	ShortEffect     string           `json:"shortEffect,omitempty"`
	Effect          string           `json:"effect,omitempty"`
	EffectChance    *int             `json:"effectChance"`
	Type            *int             `json:"type"`
	MoveTarget      *int             `json:"moveTarget"`
	MoveDamageClass *int             `json:"moveDamageClass"`
}

type MoveMethod struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

type FormMove struct {
	Level             int        `json:"level"`
	Move              MoveInfo   `json:"move"`
	PokemonMoveMethod MoveMethod `json:"pokemonMoveMethod"`
}

type PokemonAppearance struct {
	ID     int    `json:"id"`
	Gender string `json:"gender"`
	Shiny  bool   `json:"shiny"`
}

type PokemonVariation struct {
	ID                 int                 `json:"id"`
	Name               string              `json:"name"`
	Identifier         string              `json:"identifier"`
	Description        string              `json:"description,omitempty"`
	Components         []Component         `json:"components,omitempty"`
	PokemonAppearances []PokemonAppearance `json:"pokemonAppearances,omitempty"`
}

type SpeciesRef struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

type SpeciesInfo struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	Identifier     string         `json:"identifier"`
	Generation     int            `json:"generation"`
	GenderRate     int            `json:"genderRate"`
	CatchRate      int            `json:"catchRate"`
	GrowthRate     int            `json:"growthRate"`
	BaseFriendship int            `json:"baseFriendship"`
	EggCycle       int            `json:"eggCycle"`
	IsBaby         bool           `json:"isBaby"`
	Classification Classification `json:"classification"`
}

type FormRef struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	FormIdentifier string     `json:"formIdentifier"`
	Species        SpeciesRef `json:"species,omitempty"`
}

type FormTransition struct {
	Method          string  `json:"method"`
	ToPokemonForm   FormRef `json:"toPokemonForm,omitempty"`
	FromPokemonForm FormRef `json:"fromPokemonForm,omitempty"`
}

type PokemonForm struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	FormIdentifier    string             `json:"formIdentifier"`
	Generation        int                `json:"generation"`
	Description       string             `json:"description,omitempty"`
	Height            int                `json:"height"`
	Weight            int                `json:"weight"`
	BaseExperience    int                `json:"baseExperience"`
	Stats             Stats              `json:"stats"`
	Efforts           Stats              `json:"efforts"`
	Components        []Component        `json:"components,omitempty"`
	Abilities         []FormAbility      `json:"abilities,omitempty"`
	Types             []FormType         `json:"types,omitempty"`
	Moves             []FormMove         `json:"moves,omitempty"`
	PokemonVariations []PokemonVariation `json:"pokemonVariations,omitempty"`
	ToPokemonForms    []FormTransition   `json:"toPokemonForms,omitempty"`
	FromPokemonForms  []FormTransition   `json:"fromPokemonForms,omitempty"`
}

type Species struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	Identifier     string         `json:"identifier"`
	Generation     int            `json:"generation"`
	GenderRate     int            `json:"genderRate"`
	CatchRate      int            `json:"catchRate"`
	GrowthRate     int            `json:"growthRate"`
	BaseFriendship int            `json:"baseFriendship"`
	EggCycle       int            `json:"eggCycle"`
	IsBaby         bool           `json:"isBaby"`
	Classification Classification `json:"classification"`
	EggGroups      []EggGroup     `json:"eggGroups"`
	PokemonForms   []PokemonForm  `json:"pokemonForms,omitempty"`
	RawJSON        string         `json:"-"`
}

func (s *Species) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Pokemon: %s\n", s.Name))
	sb.WriteString(fmt.Sprintf("Generation: %d\n", s.Generation))

	if s.Classification.Name != "" {
		sb.WriteString(fmt.Sprintf("Classification: %s\n", s.Classification.Name))
	}

	if len(s.EggGroups) > 0 {
		var groups []string
		for _, g := range s.EggGroups {
			groups = append(groups, g.Name)
		}
		sb.WriteString(fmt.Sprintf("Egg Groups: %s\n", strings.Join(groups, ", ")))
	}

	if len(s.PokemonForms) > 0 {
		form := s.PokemonForms[0]

		if len(form.Types) > 0 {
			var types []string
			for _, t := range form.Types {
				types = append(types, t.Type.Name)
			}
			sb.WriteString(fmt.Sprintf("Types: %s\n", strings.Join(types, ", ")))
		}

		if len(form.Abilities) > 0 {
			var abilities []string
			for _, a := range form.Abilities {
				name := a.Ability.Name
				if a.IsHidden {
					name += " (Hidden)"
				}
				abilities = append(abilities, name)
			}
			sb.WriteString(fmt.Sprintf("Abilities: %s\n", strings.Join(abilities, ", ")))
		}

		sb.WriteString(fmt.Sprintf("Stats: HP %d, Atk %d, Def %d, SpA %d, SpD %d, Spe %d\n",
			form.Stats.HP, form.Stats.Attack, form.Stats.Defense,
			form.Stats.SpecialAttack, form.Stats.SpecialDefense, form.Stats.Speed))

		sb.WriteString(fmt.Sprintf("Height: %.1fm, Weight: %.1fkg\n",
			float64(form.Height)/10, float64(form.Weight)/10))

		if len(form.Moves) > 0 {
			var moves []string
			for _, m := range form.Moves {
				if len(moves) < 20 {
					moves = append(moves, m.Move.Name)
				}
			}
			sb.WriteString(fmt.Sprintf("Moves: %s\n", strings.Join(moves, ", ")))
		}
	}

	sb.WriteString(fmt.Sprintf("Catch Rate: %d\n", s.CatchRate))
	sb.WriteString(fmt.Sprintf("Base Friendship: %d\n", s.BaseFriendship))

	if s.IsBaby {
		sb.WriteString("Baby Pokemon: Yes\n")
	}

	return sb.String()
}

func (s *Species) MetadataJSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type MoveFlag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

type MoveSearchResult struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Similarity float64 `json:"similarity"`
}

type TypeSearchResult struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Similarity float64 `json:"similarity"`
}

type Move struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Identifier      string          `json:"identifier"`
	Generation      int             `json:"generation"`
	Type            TypeInfo        `json:"type"`
	Power           int             `json:"power"`
	PowerPoints     int             `json:"powerPoints"`
	Accuracy        int             `json:"accuracy"`
	Priority        int             `json:"priority"`
	MoveTarget      MoveTarget      `json:"moveTarget"`
	ShortEffect     string          `json:"shortEffect,omitempty"`
	Effect          string          `json:"effect,omitempty"`
	EffectChance    *int            `json:"effectChance"`
	MoveDamageClass MoveDamageClass `json:"moveDamageClass"`
	Flags           []MoveFlag      `json:"flags"`
	RawJSON         string          `json:"-"`
}

func (m *Move) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Move: %s\n", m.Name))
	sb.WriteString(fmt.Sprintf("Type: %s\n", m.Type.Name))
	sb.WriteString(fmt.Sprintf("Category: %s\n", m.MoveDamageClass.Name))
	sb.WriteString(fmt.Sprintf("Power: %d\n", m.Power))
	sb.WriteString(fmt.Sprintf("Accuracy: %d\n", m.Accuracy))
	sb.WriteString(fmt.Sprintf("PP: %d\n", m.PowerPoints))
	sb.WriteString(fmt.Sprintf("Priority: %d\n", m.Priority))
	sb.WriteString(fmt.Sprintf("Target: %s\n", m.MoveTarget.Name))

	if m.ShortEffect != "" {
		sb.WriteString(fmt.Sprintf("Effect: %s\n", m.ShortEffect))
	} else if m.Effect != "" {
		sb.WriteString(fmt.Sprintf("Effect: %s\n", m.Effect))
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

func (m *Move) MetadataJSON() string {
	b, _ := json.Marshal(m)
	return string(b)
}

type AbilitySearchResult struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Similarity float64 `json:"similarity"`
}

type Ability struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Identifier       string `json:"identifier"`
	Generation       int    `json:"generation"`
	Description      string `json:"description,omitempty"`
	ShortDescription string `json:"shortDescription,omitempty"`
	RawJSON          string `json:"-"`
}

func (a *Ability) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Ability: %s\n", a.Name))
	sb.WriteString(fmt.Sprintf("Generation: %d\n", a.Generation))

	if a.ShortDescription != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", a.ShortDescription))
	} else if a.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", a.Description))
	}

	return sb.String()
}

func (a *Ability) MetadataJSON() string {
	b, _ := json.Marshal(a)
	return string(b)
}

type Category struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

type ArticleCategory struct {
	Category Category `json:"category"`
}

type ArticleSearchResult struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Identifier string  `json:"identifier"`
	Similarity float64 `json:"similarity"`
}

type Article struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Subtitle    string            `json:"subtitle,omitempty"`
	Description string            `json:"description,omitempty"`
	Body        string            `json:"body"`
	Identifier  string            `json:"identifier"`
	Categories  []ArticleCategory `json:"categories,omitempty"`
	RawJSON     string            `json:"-"`
}

func (a *Article) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Article: %s\n", a.Title))

	if a.Subtitle != "" {
		sb.WriteString(fmt.Sprintf("Subtitle: %s\n", a.Subtitle))
	}

	if len(a.Categories) > 0 {
		var cats []string
		for _, c := range a.Categories {
			cats = append(cats, c.Category.Name)
		}
		sb.WriteString(fmt.Sprintf("Categories: %s\n", strings.Join(cats, ", ")))
	}

	if a.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", a.Description))
	}

	sb.WriteString(fmt.Sprintf("\n%s\n", a.Body))

	return sb.String()
}

func (a *Article) MetadataJSON() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func GrowthRateName(id int) string {
	switch id {
	case 1:
		return "Slow"
	case 2:
		return "Medium"
	case 3:
		return "Fast"
	case 4:
		return "Medium Slow"
	case 5:
		return "Erratic"
	case 6:
		return "Fluctuating"
	default:
		return "Unknown"
	}
}

func GenderRateDescription(rate int) string {
	switch rate {
	case -1:
		return "Genderless"
	case 0:
		return "100% male"
	case 1:
		return "87.5% male, 12.5% female"
	case 2:
		return "75% male, 25% female"
	case 4:
		return "50% male, 50% female"
	case 6:
		return "25% male, 75% female"
	case 7:
		return "12.5% male, 87.5% female"
	case 8:
		return "100% female"
	default:
		return "Unknown"
	}
}

type FormSearchParams struct {
	Query             string
	FormID            string
	SpeciesID         string
	VariationID       string
	Types             []string
	Abilities         []string
	Moves             []string
	Generation        *int
	MinHP             *int
	MaxHP             *int
	MinAttack         *int
	MaxAttack         *int
	MinDefense        *int
	MaxDefense        *int
	MinSpecialAttack  *int
	MaxSpecialAttack  *int
	MinSpecialDefense *int
	MaxSpecialDefense *int
	MinSpeed          *int
	MaxSpeed          *int
	MinBST            *int
	MaxBST            *int
	Include           []string
	Limit             int
	Offset            int
}

type FormSearchResult struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	FormName          string             `json:"formName"`
	FormIdentifier    string             `json:"formIdentifier"`
	Generation        int                `json:"generation"`
	Height            int                `json:"height"`
	Weight            int                `json:"weight"`
	BaseExperience    int                `json:"baseExperience"`
	Species           SpeciesInfo        `json:"species"`
	Stats             *Stats             `json:"stats,omitempty"`
	Abilities         []FormAbility      `json:"abilities,omitempty"`
	Types             []FormType         `json:"types,omitempty"`
	Moves             []FormMove         `json:"moves,omitempty"`
	PokemonVariations []PokemonVariation `json:"pokemonVariations,omitempty"`
}

type FormSearchResponse struct {
	Data   []FormSearchResult `json:"data"`
	Total  int                `json:"total"`
	Offset int                `json:"offset"`
	Limit  int                `json:"limit"`
}

func (f *FormSearchResult) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Pokemon: %s\n", f.Name))
	if f.FormName != "" && f.FormName != "Base" {
		sb.WriteString(fmt.Sprintf("Form: %s\n", f.FormName))
	}
	sb.WriteString(fmt.Sprintf("Species: %s\n", f.Species.Name))
	sb.WriteString(fmt.Sprintf("Generation: %d\n", f.Generation))

	if len(f.Types) > 0 {
		var types []string
		for _, t := range f.Types {
			types = append(types, t.Type.Name)
		}
		sb.WriteString(fmt.Sprintf("Types: %s\n", strings.Join(types, ", ")))
	}

	if len(f.Abilities) > 0 {
		var abilities []string
		for _, a := range f.Abilities {
			name := a.Ability.Name
			if a.IsHidden {
				name += " (Hidden)"
			}
			abilities = append(abilities, name)
		}
		sb.WriteString(fmt.Sprintf("Abilities: %s\n", strings.Join(abilities, ", ")))
	}

	if f.Stats != nil {
		sb.WriteString(fmt.Sprintf("Stats: HP %d, Atk %d, Def %d, SpA %d, SpD %d, Spe %d\n",
			f.Stats.HP, f.Stats.Attack, f.Stats.Defense,
			f.Stats.SpecialAttack, f.Stats.SpecialDefense, f.Stats.Speed))
	}

	sb.WriteString(fmt.Sprintf("Height: %.1fm, Weight: %.1fkg\n",
		float64(f.Height)/10, float64(f.Weight)/10))

	if len(f.Moves) > 0 {
		var moves []string
		for _, m := range f.Moves {
			moves = append(moves, m.Move.Name)
		}
		sb.WriteString(fmt.Sprintf("Moves: %s\n", strings.Join(moves, ", ")))
	}

	return sb.String()
}
