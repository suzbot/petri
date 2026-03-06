package entity

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/types"
)

// bundlePluralName maps bundleable item types to their plural form for display.
// Only needed for types without Kind set — types with Kind use Pluralize(Kind) instead.
var bundlePluralName = map[string]string{
	"stick": "sticks",
}

// PlantProperties contains properties specific to growing plants
type PlantProperties struct {
	IsGrowing   bool    // Can reproduce at location (gates spawning)
	SpawnTimer  float64 // Countdown until next spawn opportunity
	IsSprout    bool    // True if this is a newly planted sprout (not yet mature)
	SproutTimer float64 // Countdown until sprout matures into full plant
	SeedTimer   float64 // Cooldown after seed extraction (0 = seeds available)
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

	// SourceVarietyID — for seeds, the variety registry ID of the parent plant.
	// Used to reconstruct the parent plant when planted.
	SourceVarietyID string

	// Bundle count for stackable materials (sticks, grass). 0 = not a bundle.
	BundleCount int

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

// NewGrass creates a new tall grass item (plant, non-edible, bundleable)
func NewGrass(x, y int) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharGrass,
			EType: TypeItem,
		},
		ItemType:    "grass",
		Kind:        "tall grass",
		Color:       types.ColorPaleGreen,
		Plant:       &PlantProperties{IsGrowing: true},
		BundleCount: 1,
	}
}

// NewStick creates a new stick item (non-edible, non-plant, bundleable)
func NewStick(x, y int) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharStick,
			EType: TypeItem,
		},
		ItemType:    "stick",
		Color:       types.ColorBrown,
		BundleCount: 1,
	}
}

// NewClay creates a new clay item (non-edible, non-plant, non-bundleable, no variety)
func NewClay(x, y int) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharClay,
			EType: TypeItem,
		},
		Name:     "lump of clay",
		ItemType: "clay",
		Color:    types.ColorEarthy,
	}
}

// NewBrick creates a new brick item (non-edible, non-plant, non-bundleable, no variety, DD-21)
func NewBrick(x, y int) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharBrick,
			EType: TypeItem,
		},
		Name:     "brick",
		ItemType: "brick",
		Color:    types.ColorTerracotta,
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

// NewSeed creates a new seed item from a parent plant type.
// sourceVarietyID is the parent plant's variety registry ID.
// parentKind is the parent's Kind (e.g., "tall grass"); when non-empty, seed Kind = parentKind + " seed".
// When parentKind is empty, seed Kind = parentItemType + " seed".
func NewSeed(x, y int, parentItemType string, sourceVarietyID string, parentKind string, color types.Color, pattern types.Pattern, texture types.Texture) *Item {
	kind := parentItemType + " seed"
	if parentKind != "" {
		kind = parentKind + " seed"
	}
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharSeed,
			EType: TypeItem,
		},
		ItemType:        "seed",
		Kind:            kind,
		Color:           color,
		Pattern:         pattern,
		Texture:         texture,
		Plantable:       true,
		SourceVarietyID: sourceVarietyID,
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

	if i.BundleCount >= 2 {
		var plural string
		if i.Kind != "" {
			plural = Pluralize(i.Kind)
		} else {
			plural = bundlePluralName[i.ItemType]
			if plural == "" {
				plural = i.ItemType + "s"
			}
		}
		return fmt.Sprintf("bundle of %s (%d)", plural, i.BundleCount)
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

// CreateSprout creates a sprout item from a parent variety.
// The caller resolves the parent variety: for seeds via SourceVarietyID registry lookup,
// for berries/mushrooms via their own variety. All attributes come from the variety.
func CreateSprout(x, y int, parentVariety *ItemVariety) *Item {
	return &Item{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharSprout,
			EType: TypeItem,
		},
		ItemType: parentVariety.ItemType,
		Kind:     parentVariety.Kind,
		Color:    parentVariety.Color,
		Pattern:  parentVariety.Pattern,
		Texture:  parentVariety.Texture,
		Plant: &PlantProperties{
			IsGrowing:   true,
			IsSprout:    true,
			SproutTimer: config.GetSproutDuration(parentVariety.ItemType),
		},
		Edible: parentVariety.Edible,
	}
}
