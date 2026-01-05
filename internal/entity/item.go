package entity

import (
	"petri/internal/config"
	"petri/internal/types"
)

// Item represents an item in the game world
type Item struct {
	BaseEntity

	// Descriptive attributes (opinion-formable)
	ItemType string        // "berry", "mushroom", etc.
	Color    types.Color   // all items have color
	Pattern  types.Pattern // mushrooms only (spotted, plain)
	Texture  types.Texture // mushrooms only (slimy, none)

	// Functional attributes (not opinion-formable)
	Edible    bool
	Poisonous bool
	Healing   bool

	// Spawning
	SpawnTimer float64 // countdown until next spawn opportunity
}

// NewBerry creates a new berry item
func NewBerry(x, y int, color types.Color, poisonous, healing bool) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharBerry,
			EType: TypeItem,
		},
		ItemType:  "berry",
		Color:     color,
		Edible:    true,
		Poisonous: poisonous,
		Healing:   healing,
	}
}

// NewMushroom creates a new mushroom item
func NewMushroom(x, y int, color types.Color, pattern types.Pattern, texture types.Texture, poisonous, healing bool) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharMushroom,
			EType: TypeItem,
		},
		ItemType:  "mushroom",
		Color:     color,
		Pattern:   pattern,
		Texture:   texture,
		Edible:    true,
		Poisonous: poisonous,
		Healing:   healing,
	}
}

// NewFlower creates a new flower item (decorative, not edible)
func NewFlower(x, y int, color types.Color) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharFlower,
			EType: TypeItem,
		},
		ItemType:  "flower",
		Color:     color,
		Edible:    false,
		Poisonous: false,
		Healing:   false,
	}
}

// Description returns a human-readable item description
// Format: [texture] [pattern] [color] [itemType]
// e.g., "slimy spotted red mushroom", "red berry", "purple flower"
func (i *Item) Description() string {
	var parts []string

	if i.Texture != types.TextureNone {
		parts = append(parts, string(i.Texture))
	}
	if i.Pattern != types.PatternNone {
		parts = append(parts, string(i.Pattern))
	}
	if i.Color != "" {
		parts = append(parts, string(i.Color))
	}
	parts = append(parts, i.ItemType)

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	return result
}
