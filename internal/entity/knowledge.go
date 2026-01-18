package entity

import (
	"petri/internal/types"
	"strings"
)

// KnowledgeCategory represents the type of knowledge about an item
type KnowledgeCategory string

const (
	KnowledgePoisonous KnowledgeCategory = "poisonous"
	KnowledgeHealing   KnowledgeCategory = "healing"
)

// Knowledge represents something a character has learned about items
type Knowledge struct {
	Category KnowledgeCategory
	ItemType string
	Color    types.Color
	Pattern  types.Pattern
	Texture  types.Texture
}

// NewKnowledgeFromItem creates a Knowledge entry from an item
func NewKnowledgeFromItem(item *Item, category KnowledgeCategory) Knowledge {
	return Knowledge{
		Category: category,
		ItemType: item.ItemType,
		Color:    item.Color,
		Pattern:  item.Pattern,
		Texture:  item.Texture,
	}
}

// Description returns a human-readable description of this knowledge
// Format: "[Texture] [pattern] [color] [itemType]s are [category]"
// First letter is capitalized
func (k Knowledge) Description() string {
	var parts []string

	if k.Texture != types.TextureNone && k.Texture != "" {
		parts = append(parts, string(k.Texture))
	}
	if k.Pattern != types.PatternNone && k.Pattern != "" {
		parts = append(parts, string(k.Pattern))
	}
	if k.Color != "" {
		parts = append(parts, string(k.Color))
	}
	parts = append(parts, Pluralize(k.ItemType))

	description := strings.Join(parts, " ")
	description += " are " + string(k.Category)

	// Capitalize first letter
	if len(description) > 0 {
		description = strings.ToUpper(string(description[0])) + description[1:]
	}

	return description
}

// Matches returns true if this knowledge applies to the given item
func (k Knowledge) Matches(item *Item) bool {
	if k.ItemType != item.ItemType {
		return false
	}
	if k.Color != item.Color {
		return false
	}
	if k.Pattern != item.Pattern {
		return false
	}
	if k.Texture != item.Texture {
		return false
	}
	return true
}

// Equals returns true if two knowledge entries are identical
func (k Knowledge) Equals(other Knowledge) bool {
	return k.Category == other.Category &&
		k.ItemType == other.ItemType &&
		k.Color == other.Color &&
		k.Pattern == other.Pattern &&
		k.Texture == other.Texture
}
