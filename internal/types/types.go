// Package types defines shared type constants used across the codebase.
// This centralizes type definitions to avoid scattered string literals
// and provides compile-time safety.
package types

// Color represents item/entity colors (descriptive attribute, opinion-formable)
type Color string

const (
	ColorRed    Color = "red"
	ColorBlue   Color = "blue"
	ColorBrown  Color = "brown"
	ColorWhite  Color = "white"
	ColorOrange Color = "orange"
	ColorYellow Color = "yellow"
	ColorPurple Color = "purple"
)

// AllColors returns all valid colors
var AllColors = []Color{ColorRed, ColorBlue, ColorBrown, ColorWhite, ColorOrange, ColorYellow, ColorPurple}

// BerryColors returns valid colors for berries
var BerryColors = []Color{ColorRed, ColorBlue}

// MushroomColors returns valid colors for mushrooms
var MushroomColors = []Color{ColorBrown, ColorWhite, ColorRed}

// FlowerColors returns valid colors for flowers
var FlowerColors = []Color{ColorRed, ColorOrange, ColorYellow, ColorBlue, ColorPurple, ColorWhite}

// StatType represents character survival stats
type StatType string

const (
	StatHunger StatType = "hunger"
	StatThirst StatType = "thirst"
	StatEnergy StatType = "energy"
	StatHealth StatType = "health"
	StatMood   StatType = "mood"
)
