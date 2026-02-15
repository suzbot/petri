package entity

import (
	"strings"

	"petri/internal/config"
	"petri/internal/types"
)

// PlantProperties contains properties specific to growing plants
type PlantProperties struct {
	IsGrowing   bool    // Can reproduce at location (gates spawning)
	SpawnTimer  float64 // Countdown until next spawn opportunity
	IsSprout    bool    // True if this is a newly planted sprout (not yet mature)
	SproutTimer float64 // Countdown until sprout matures into full plant
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
	ItemType string        // broad category: "berry", "hoe", "vessel"
	Kind     string        // recipe subtype: "shell hoe", "hollow gourd" (empty for natural items)
	Color    types.Color   // all items have color
	Pattern  types.Pattern // mushrooms, gourds (spotted, striped, speckled)
	Texture  types.Texture // mushrooms, gourds (slimy, waxy, warty)

	// Plant properties (nil for non-plants like crafted items)
	Plant *PlantProperties

	// Container properties (nil for non-containers)
	Container *ContainerData

	// Edible properties (nil for non-edible items like vessels, flowers)
	Edible *EdibleProperties

	// Plantable - set when berries/mushrooms are picked, or for seeds
	Plantable bool

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

// NewStick creates a new stick item (non-edible, non-plant)
func NewStick(x, y int) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharStick,
			EType: TypeItem,
		},
		ItemType: "stick",
		Color:    types.ColorBrown,
	}
}

// NewNut creates a new nut item (edible, non-plant)
func NewNut(x, y int) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharNut,
			EType: TypeItem,
		},
		ItemType: "nut",
		Color:    types.ColorBrown,
		Edible:   &EdibleProperties{},
	}
}

// NewShell creates a new shell item (non-edible, non-plant)
func NewShell(x, y int, color types.Color) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharShell,
			EType: TypeItem,
		},
		ItemType: "shell",
		Color:    color,
	}
}

// NewSeed creates a new seed item from a parent plant type
func NewSeed(x, y int, parentItemType string, color types.Color, pattern types.Pattern, texture types.Texture) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharSeed,
			EType: TypeItem,
		},
		ItemType:  "seed",
		Kind:      parentItemType + " seed",
		Color:     color,
		Pattern:   pattern,
		Texture:   texture,
		Plantable: true,
	}
}

// NewHoe creates a new hoe item (non-edible, non-plant, crafted from stick + shell)
func NewHoe(x, y int, color types.Color) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharHoe,
			EType: TypeItem,
		},
		ItemType: "hoe",
		Kind:     "shell hoe",
		Color:    color,
	}
}

// Description returns a human-readable item description
// If Name is set, returns Name.
// Otherwise returns format: [texture] [pattern] [color] [kind or itemType]
// Kind is used when present (crafted items), ItemType as fallback (natural items).
// e.g., "silver shell hoe", "warty spotted green hollow gourd", "red berry"
func (i *Item) Description() string {
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
	if i.Kind != "" {
		parts = append(parts, i.Kind)
	} else {
		parts = append(parts, i.ItemType)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	if i.Plant != nil && i.Plant.IsSprout {
		result += " sprout"
	}
	return result
}

// CreateSprout creates a sprout item from a plantable item.
// For seeds (ItemType="seed"), derives parent type from Kind ("gourd seed" → "gourd").
// For berries/mushrooms, uses the item's own type.
// Caller provides edible properties (from item.Edible for berries/mushrooms,
// from registry lookup for seeds since seeds themselves aren't edible).
func CreateSprout(x, y int, plantedItem *Item, edible *EdibleProperties) *Item {
	parentType := plantedItem.ItemType
	if plantedItem.ItemType == "seed" {
		parentType = strings.TrimSuffix(plantedItem.Kind, " seed")
	}

	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharSprout,
			EType: TypeItem,
		},
		ItemType: parentType,
		Color:    plantedItem.Color,
		Pattern:  plantedItem.Pattern,
		Texture:  plantedItem.Texture,
		Plant: &PlantProperties{
			IsGrowing:   true,
			IsSprout:    true,
			SproutTimer: config.SproutDuration,
		},
		Edible: edible,
	}
}
