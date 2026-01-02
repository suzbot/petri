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
	FType       FeatureType
	DrinkSource bool // Can be used to drink
	Bed         bool // Can be used to sleep
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
