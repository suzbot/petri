package entity

import (
	"petri/internal/types"
)

// Preference represents a character's like or dislike for item attributes.
// Uses implicit typing: which attributes are set determines the preference type.
// - Only ItemType set: preference about item type (e.g., "berries")
// - Only Color set: preference about color (e.g., "red")
// - Both set: preference about combination (e.g., "red berries")
type Preference struct {
	Valence  int         // +1 (likes) or -1 (dislikes)
	ItemType string      // empty if not part of preference
	Color    types.Color // zero value if not part of preference
	// Future: Material, Pattern, Origin
}

// Matches returns true if the item matches all set attributes of this preference.
// An empty preference (no attributes set) matches nothing.
func (p Preference) Matches(item *Item) bool {
	if p.ItemType == "" && p.Color == "" {
		return false // Empty preference matches nothing
	}

	if p.ItemType != "" && p.ItemType != item.ItemType {
		return false
	}

	if p.Color != "" && p.Color != item.Color {
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
	// Future: if p.Material != "" { count++ }
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
// Examples: "berries", "red", "red berries"
func (p Preference) Description() string {
	hasType := p.ItemType != ""
	hasColor := p.Color != ""

	if hasType && hasColor {
		return string(p.Color) + " " + pluralize(p.ItemType)
	}
	if hasType {
		return pluralize(p.ItemType)
	}
	if hasColor {
		return string(p.Color)
	}
	return ""
}

// IsPositive returns true if this is a "likes" preference.
func (p Preference) IsPositive() bool {
	return p.Valence > 0
}

// ExactMatch returns true if both preferences target the same attributes
// (regardless of valence). Used for preference formation logic.
func (p Preference) ExactMatch(other Preference) bool {
	return p.ItemType == other.ItemType && p.Color == other.Color
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
