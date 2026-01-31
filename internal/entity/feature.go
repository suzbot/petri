package entity

import "petri/internal/config"

// FeatureType identifies different landscape features
type FeatureType int

const (
	FeatureSpring FeatureType = iota
	FeatureLeafPile
)

// Feature represents a persistent landscape feature (spring, leaf pile, etc.)
type Feature struct {
	BaseEntity
	ID          int // Unique identifier for save/load
	FType       FeatureType
	DrinkSource bool // Can be used to drink
	Bed         bool // Can be used to sleep
	Passable    bool // Can characters walk onto this feature
}

// NewSpring creates a new water spring
func NewSpring(x, y int) *Feature {
	return &Feature{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharSpring,
			EType: TypeFeature,
		},
		FType:       FeatureSpring,
		DrinkSource: true,
		Bed:         false,
		Passable:    false, // Springs are impassable - drink from adjacent tiles
	}
}

// NewLeafPile creates a new leaf pile (bed)
func NewLeafPile(x, y int) *Feature {
	return &Feature{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharLeafPile,
			EType: TypeFeature,
		},
		FType:       FeatureLeafPile,
		DrinkSource: false,
		Bed:         true,
		Passable:    true, // Leaf piles are passable - walk onto them to sleep
	}
}

// IsDrinkSource returns true if this feature can be used for drinking
func (f *Feature) IsDrinkSource() bool {
	return f.DrinkSource
}

// IsBed returns true if this feature can be used for sleeping
func (f *Feature) IsBed() bool {
	return f.Bed
}

// IsPassable returns true if characters can walk onto this feature
func (f *Feature) IsPassable() bool {
	return f.Passable
}

// Description returns a human-readable description
func (f *Feature) Description() string {
	switch f.FType {
	case FeatureSpring:
		return "spring"
	case FeatureLeafPile:
		return "leaf pile"
	default:
		return "feature"
	}
}
