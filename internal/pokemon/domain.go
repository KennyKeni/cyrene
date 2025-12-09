package pokemon

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Pokemon struct {
	ID         string
	Identifier string
	RawJSON    string
	Metadata   map[string]any
}

func (p *Pokemon) EmbeddingText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Pokemon: %s (ID: %s)\n", p.Identifier, p.ID))

	if types, ok := p.Metadata["types"].([]any); ok {
		var typeNames []string
		for _, t := range types {
			if tm, ok := t.(map[string]any); ok {
				if typeInfo, ok := tm["type"].(map[string]any); ok {
					if name, ok := typeInfo["name"].(string); ok {
						typeNames = append(typeNames, name)
					}
				}
			}
		}
		if len(typeNames) > 0 {
			sb.WriteString(fmt.Sprintf("Types: %s\n", strings.Join(typeNames, ", ")))
		}
	}

	if abilities, ok := p.Metadata["abilities"].([]any); ok {
		var abilityNames []string
		for _, a := range abilities {
			if am, ok := a.(map[string]any); ok {
				if abilityInfo, ok := am["ability"].(map[string]any); ok {
					if name, ok := abilityInfo["name"].(string); ok {
						abilityNames = append(abilityNames, name)
					}
				}
			}
		}
		if len(abilityNames) > 0 {
			sb.WriteString(fmt.Sprintf("Abilities: %s\n", strings.Join(abilityNames, ", ")))
		}
	}

	if moves, ok := p.Metadata["moves"].([]any); ok {
		var moveNames []string
		for _, m := range moves {
			if mm, ok := m.(map[string]any); ok {
				if moveInfo, ok := mm["move"].(map[string]any); ok {
					if name, ok := moveInfo["name"].(string); ok {
						moveNames = append(moveNames, name)
					}
				}
			}
		}
		if len(moveNames) > 0 {
			sb.WriteString(fmt.Sprintf("Moves: %s\n", strings.Join(moveNames, ", ")))
		}
	}

	if stats, ok := p.Metadata["stats"].([]any); ok {
		var statStrs []string
		for _, s := range stats {
			if sm, ok := s.(map[string]any); ok {
				if statInfo, ok := sm["stat"].(map[string]any); ok {
					if name, ok := statInfo["name"].(string); ok {
						if baseStat, ok := sm["base_stat"].(float64); ok {
							statStrs = append(statStrs, fmt.Sprintf("%s: %.0f", name, baseStat))
						}
					}
				}
			}
		}
		if len(statStrs) > 0 {
			sb.WriteString(fmt.Sprintf("Stats: %s\n", strings.Join(statStrs, ", ")))
		}
	}

	if height, ok := p.Metadata["height"].(float64); ok {
		sb.WriteString(fmt.Sprintf("Height: %.1f\n", height))
	}
	if weight, ok := p.Metadata["weight"].(float64); ok {
		sb.WriteString(fmt.Sprintf("Weight: %.1f\n", weight))
	}

	return sb.String()
}

func (p *Pokemon) MetadataJSON() string {
	b, _ := json.Marshal(p.Metadata)
	return string(b)
}
