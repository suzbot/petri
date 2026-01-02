package entity

import (
	"petri/internal/config"
	"petri/internal/types"
)

// Item represents an item in the game world
type Item struct {
	BaseEntity

	// Descriptive attributes (opinion-formable)
	ItemType string      // "berry", "mushroom", etc.
	Color    types.Color

	// Functional attributes (not opinion-formable)
	Edible    bool
	Poisonous bool
	Healing   bool
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
func NewMushroom(x, y int, color types.Color, poisonous, healing bool) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharMushroom,
			EType: TypeItem,
		},
		ItemType:  "mushroom",
		Color:     color,
		Edible:    true,
		Poisonous: poisonous,
		Healing:   healing,
	}
}

// Description returns a human-readable item description
func (i *Item) Description() string {
	return string(i.Color) + " " + i.ItemType
}
