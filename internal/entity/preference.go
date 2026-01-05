package entity

import (
	"petri/internal/types"
)

// Preference represents a character's like or dislike for item attributes.
// Uses implicit typing: which attributes are set determines the preference type.
// - Only ItemType set: preference about item type (e.g., "berries")
// - Only Color set: preference about color (e.g., "red")
// - Both set: preference about combination (e.g., "red berries")
// - Pattern/Texture: only relevant for mushrooms
type Preference struct {
	Valence  int           // +1 (likes) or -1 (dislikes)
	ItemType string        // empty if not part of preference
	Color    types.Color   // zero value if not part of preference
	Pattern  types.Pattern // mushrooms only: spotted, plain
	Texture  types.Texture // mushrooms only: slimy, none
	// Future: Material, Origin
}

// Matches returns true if the item matches all set attributes of this preference.
// An empty preference (no attributes set) matches nothing.
func (p Preference) Matches(item *Item) bool {
	if p.ItemType == "" && p.Color == "" && p.Pattern == "" && p.Texture == "" {
		return false // Empty preference matches nothing
	}

	if p.ItemType != "" && p.ItemType != item.ItemType {
		return false
	}

	if p.Color != "" && p.Color != item.Color {
		return false
	}

	if p.Pattern != "" && p.Pattern != item.Pattern {
		return false
	}

	if p.Texture != "" && p.Texture != item.Texture {
		return false
	}

	return true
}

// AttributeCount returns the number of attributes specified in this preference.
func (p Preference) AttributeCount() int {
	count := 0
	if p.ItemType != "" {
		count++
	}
	if p.Color != "" {
		count++
	}
	if p.Pattern != "" {
		count++
	}
	if p.Texture != "" {
		count++
	}
	return count
}

// MatchScore returns the preference score for an item.
// Returns 0 if the preference doesn't match the item.
// Otherwise returns Valence × AttributeCount, reflecting that more specific
// preferences (combos) contribute proportionally more to the net preference.
func (p Preference) MatchScore(item *Item) int {
	if !p.Matches(item) {
		return 0
	}
	return p.Valence * p.AttributeCount()
}

// Description returns a human-readable description of what this preference targets.
// Examples: "berries", "red", "red berries", "spotted mushrooms", "slimy red mushrooms"
func (p Preference) Description() string {
	// Build parts in order: texture, pattern, color, type
	parts := []string{}

	if p.Texture != "" {
		parts = append(parts, string(p.Texture))
	}
	if p.Pattern != "" {
		parts = append(parts, string(p.Pattern))
	}
	if p.Color != "" {
		parts = append(parts, string(p.Color))
	}
	if p.ItemType != "" {
		parts = append(parts, pluralize(p.ItemType))
	}

	if len(parts) == 0 {
		return ""
	}

	// Join with spaces
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " " + parts[i]
	}
	return result
}

// IsPositive returns true if this is a "likes" preference.
func (p Preference) IsPositive() bool {
	return p.Valence > 0
}

// ExactMatch returns true if both preferences target the same attributes
// (regardless of valence). Used for preference formation logic.
func (p Preference) ExactMatch(other Preference) bool {
	return p.ItemType == other.ItemType &&
		p.Color == other.Color &&
		p.Pattern == other.Pattern &&
		p.Texture == other.Texture
}

// pluralize returns the plural form of an item type.
func pluralize(itemType string) string {
	switch itemType {
	case "berry":
		return "berries"
	case "mushroom":
		return "mushrooms"
	case "flower":
		return "flowers"
	default:
		return itemType + "s"
	}
}

// NewPositivePreference creates a "likes" preference for the given attributes.
// Pass empty string or zero value to omit an attribute.
func NewPositivePreference(itemType string, color types.Color) Preference {
	return Preference{
		Valence:  1,
		ItemType: itemType,
		Color:    color,
	}
}

// NewNegativePreference creates a "dislikes" preference for the given attributes.
// Pass empty string or zero value to omit an attribute.
func NewNegativePreference(itemType string, color types.Color) Preference {
	return Preference{
		Valence:  -1,
		ItemType: itemType,
		Color:    color,
	}
}
