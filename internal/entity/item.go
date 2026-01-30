package entity

import (
	"petri/internal/config"
	"petri/internal/types"
)

// PlantProperties contains properties specific to growing plants
type PlantProperties struct {
	IsGrowing  bool    // Can reproduce at location (gates spawning)
	SpawnTimer float64 // Countdown until next spawn opportunity
}

// Stack represents a quantity of items of a single variety in a container
type Stack struct {
	Variety *ItemVariety // What variety this stack holds
	Count   int          // How many items in the stack
}

// ContainerData contains properties specific to containers (vessels, etc.)
type ContainerData struct {
	Capacity int     // How many stacks this container can hold
	Contents []Stack // Stacks currently in the container
}

// EdibleProperties contains properties specific to consumable items
type EdibleProperties struct {
	Poisonous bool
	Healing   bool
}

// Item represents an item in the game world
type Item struct {
	BaseEntity
	ID int // Unique identifier for save/load

	// Display name (if set, used instead of Description() for crafted items)
	Name string

	// Descriptive attributes (opinion-formable)
	ItemType string        // "berry", "mushroom", "gourd", etc.
	Color    types.Color   // all items have color
	Pattern  types.Pattern // mushrooms, gourds (spotted, striped, speckled)
	Texture  types.Texture // mushrooms, gourds (slimy, waxy, warty)

	// Plant properties (nil for non-plants like crafted items)
	Plant *PlantProperties

	// Container properties (nil for non-containers)
	Container *ContainerData

	// Edible properties (nil for non-edible items like vessels, flowers)
	Edible *EdibleProperties

	// Lifecycle
	DeathTimer float64 // countdown until death (0 = immortal)
}

// IsEdible returns true if this item can be consumed
func (i *Item) IsEdible() bool {
	return i.Edible != nil
}

// IsPoisonous returns true if this item is edible and poisonous
func (i *Item) IsPoisonous() bool {
	return i.Edible != nil && i.Edible.Poisonous
}

// IsHealing returns true if this item is edible and healing
func (i *Item) IsHealing() bool {
	return i.Edible != nil && i.Edible.Healing
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
		ItemType: "berry",
		Color:    color,
		Plant:    &PlantProperties{IsGrowing: true},
		Edible:   &EdibleProperties{Poisonous: poisonous, Healing: healing},
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
		ItemType: "mushroom",
		Color:    color,
		Pattern:  pattern,
		Texture:  texture,
		Plant:    &PlantProperties{IsGrowing: true},
		Edible:   &EdibleProperties{Poisonous: poisonous, Healing: healing},
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
		ItemType: "flower",
		Color:    color,
		Plant:    &PlantProperties{IsGrowing: true},
		// Edible is nil - flowers are not edible
	}
}

// NewGourd creates a new gourd item
func NewGourd(x, y int, color types.Color, pattern types.Pattern, texture types.Texture, poisonous, healing bool) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharGourd,
			EType: TypeItem,
		},
		ItemType: "gourd",
		Color:    color,
		Pattern:  pattern,
		Texture:  texture,
		Plant:    &PlantProperties{IsGrowing: true},
		Edible:   &EdibleProperties{Poisonous: poisonous, Healing: healing},
	}
}

// Description returns a human-readable item description
// If Name is set (crafted items), returns Name.
// Otherwise returns format: [texture] [pattern] [color] [itemType]
// e.g., "slimy spotted red mushroom", "red berry", "purple flower"
func (i *Item) Description() string {
	// Crafted items use their Name
	if i.Name != "" {
		return i.Name
	}

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
