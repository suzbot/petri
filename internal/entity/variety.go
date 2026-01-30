package entity

import (
	"strings"

	"petri/internal/types"
)

// ItemVariety defines a type of item that can exist in the world.
// Items reference varieties by ID rather than storing attributes directly.
// This allows world generation to define "what exists" before spawning items.
type ItemVariety struct {
	ID string // unique identifier, e.g., "mushroom-spotted-red-slimy"

	// Descriptive attributes (opinion-formable)
	ItemType string
	Color    types.Color
	Pattern  types.Pattern // zero value if not applicable to this item type
	Texture  types.Texture // zero value if not applicable to this item type

	// Edible properties (nil for non-edible varieties like flowers)
	Edible *EdibleProperties

	// Display
	Sym rune // symbol for rendering
}

// IsEdible returns true if this variety can be consumed
func (v *ItemVariety) IsEdible() bool {
	return v.Edible != nil
}

// IsPoisonous returns true if this variety is edible and poisonous
func (v *ItemVariety) IsPoisonous() bool {
	return v.Edible != nil && v.Edible.Poisonous
}

// IsHealing returns true if this variety is edible and healing
func (v *ItemVariety) IsHealing() bool {
	return v.Edible != nil && v.Edible.Healing
}

// Description returns a human-readable description like "slimy spotted red mushroom"
// Format: [Texture] [Pattern] [Color] [ItemType]
func (v *ItemVariety) Description() string {
	var parts []string

	if v.Texture != types.TextureNone {
		parts = append(parts, string(v.Texture))
	}
	if v.Pattern != types.PatternNone {
		parts = append(parts, string(v.Pattern))
	}
	if v.Color != "" {
		parts = append(parts, string(v.Color))
	}
	parts = append(parts, v.ItemType)

	return strings.Join(parts, " ")
}

// GenerateVarietyID creates a unique ID from the variety's attributes
func GenerateVarietyID(itemType string, color types.Color, pattern types.Pattern, texture types.Texture) string {
	var parts []string
	parts = append(parts, itemType)

	if color != "" {
		parts = append(parts, string(color))
	}
	if pattern != types.PatternNone {
		parts = append(parts, string(pattern))
	}
	if texture != types.TextureNone {
		parts = append(parts, string(texture))
	}

	return strings.Join(parts, "-")
}

// HasPattern returns true if this variety has a pattern attribute
func (v *ItemVariety) HasPattern() bool {
	return v.Pattern != types.PatternNone
}

// HasTexture returns true if this variety has a texture attribute
func (v *ItemVariety) HasTexture() bool {
	return v.Texture != types.TextureNone
}
