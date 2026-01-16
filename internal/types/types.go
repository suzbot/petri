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
	ColorTan    Color = "tan"
	ColorPink   Color = "pink"
	ColorBlack  Color = "black"
	ColorGreen  Color = "green"
)

// AllColors returns all valid colors
var AllColors = []Color{ColorRed, ColorBlue, ColorBrown, ColorWhite, ColorOrange, ColorYellow, ColorPurple, ColorTan, ColorPink, ColorBlack, ColorGreen}

// BerryColors returns valid colors for berries
var BerryColors = []Color{ColorRed, ColorBlue, ColorPink, ColorPurple, ColorWhite, ColorYellow, ColorOrange, ColorBlack}

// MushroomColors returns valid colors for mushrooms
var MushroomColors = []Color{ColorBrown, ColorWhite, ColorRed, ColorTan, ColorOrange, ColorYellow, ColorBlue, ColorBlack}

// FlowerColors returns valid colors for flowers
var FlowerColors = []Color{ColorRed, ColorOrange, ColorYellow, ColorBlue, ColorPurple, ColorWhite, ColorPink}

// GourdColors returns valid colors for gourds
var GourdColors = []Color{ColorWhite, ColorGreen, ColorYellow, ColorOrange, ColorTan}

// StatType represents character survival stats
type StatType string

const (
	StatHunger StatType = "hunger"
	StatThirst StatType = "thirst"
	StatEnergy StatType = "energy"
	StatHealth StatType = "health"
	StatMood   StatType = "mood"
)

// Pattern represents item surface patterns (descriptive attribute, opinion-formable)
type Pattern string

const (
	PatternNone     Pattern = ""
	PatternSpotted  Pattern = "spotted"
	PatternStriped  Pattern = "striped"
	PatternSpeckled Pattern = "speckled"
)

// AllPatterns returns all valid patterns (excluding None)
var AllPatterns = []Pattern{PatternSpotted, PatternStriped, PatternSpeckled}

// MushroomPatterns returns valid patterns for mushrooms (includes None for no pattern)
var MushroomPatterns = []Pattern{PatternNone, PatternSpotted}

// GourdPatterns returns valid patterns for gourds (includes None for no pattern)
var GourdPatterns = []Pattern{PatternNone, PatternStriped, PatternSpeckled}

// Texture represents item surface textures (descriptive attribute, opinion-formable)
type Texture string

const (
	TextureNone  Texture = ""
	TextureSlimy Texture = "slimy"
	TextureWaxy  Texture = "waxy"
	TextureWarty Texture = "warty"
)

// AllTextures returns all valid textures (excluding None)
var AllTextures = []Texture{TextureSlimy, TextureWaxy, TextureWarty}

// MushroomTextures returns valid textures for mushrooms (includes None for no texture)
var MushroomTextures = []Texture{TextureNone, TextureSlimy, TextureWaxy}

// GourdTextures returns valid textures for gourds (includes None for no texture)
var GourdTextures = []Texture{TextureNone, TextureWaxy, TextureWarty}
