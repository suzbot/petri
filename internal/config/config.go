package config

import (
	"time"

	"petri/internal/types"
)

const (
	MapWidth  = 60
	MapHeight = 60

	ItemSpawnCount = 20
	SpringCount    = 2
	LeafPileCount  = 4
	UpdateInterval = 150 * time.Millisecond

	// Symbols
	CharRobot     = '@'
	CharBerry     = '*'
	CharMushroom  = '^'
	CharSpring    = '~'
	CharLeafPile  = '&'
	CharSleeping  = 'z'

	// Speed system
	BaseSpeed          = 50  // baseline speed (0-100 scale)
	MinSpeed           = 5   // minimum speed floor
	PoisonSpeedPenalty = 25
	ParchedSpeedPenalty    = 10  // thirst >= 90
	DehydratedSpeedPenalty = 10  // thirst >= 100 (additional)
	VeryTiredSpeedPenalty  = 10  // energy <= 25
	ExhaustedSpeedPenalty  = 10  // energy <= 10 (additional)

	// Survival mechanics
	PoisonDuration          = 20.0 // seconds
	HungerIncreaseRate      = 0.7  // per second (slower than thirst)
	ThirstIncreaseRate      = 0.85 // per second (slightly faster than hunger)
	EnergyDecreaseRate      = 0.5  // per second (base rate)
	EnergyMovementDrain     = 0.2  // additional per movement tick
	StarvationDamageRate    = 0.5  // health per second
	DehydrationDamageRate   = 0.5  // health per second
	PoisonDamageRate        = 0.33 // health per second
	FoodHungerReduction     = 20.0 // hunger reduced per food
	DrinkThirstReduction    = 20.0 // thirst reduced per drink
	BedEnergyRestoreRate    = 5.0  // energy per second in bed
	GroundEnergyRestoreRate = 2.0  // energy per second on ground
	SatisfactionCooldown    = 5.0  // seconds before stat starts changing after reaching optimal
	ActionDuration          = 1.0  // seconds for consume/drink/sleep actions to complete

	// Frustration mechanics
	FrustrationThreshold = 3   // consecutive failed intents before frustrated
	FrustrationDuration  = 5.0 // seconds to stay frustrated

	// Mood mechanics
	MoodIncreaseRate       = 0.5 // per second when all needs at TierNone
	MoodDecreaseRateSlow   = 0.5 // per second at Moderate highest need
	MoodDecreaseRateMedium = 1.5 // per second at Severe highest need
	MoodDecreaseRateFast   = 3.0 // per second at Crisis highest need
	MoodBoostOnConsumption = 5.0 // mood boost when eating or drinking
	MoodPreferenceModifier = 5.0 // mood change per NetPreference point on consumption
	MoodPenaltyPoisoned    = 2.0 // per second while poisoned (additive with need decay)
	MoodPenaltyFrustrated  = 2.0 // per second while frustrated (additive with need decay)

	// Healing
	HealAmount = 20.0 // health restored by healing items (instant)

	// Preference formation (values inflated for testing - see CLAUDE.md balance tuning)
	PrefFormationChanceMiserable = 0.20 // 20% chance when Miserable
	PrefFormationChanceUnhappy   = 0.10 // 10% chance when Unhappy
	PrefFormationChanceHappy     = 0.10 // 10% chance when Happy
	PrefFormationChanceJoyful    = 0.20 // 20% chance when Joyful
	PrefFormationWeightItemType  = 0.20 // 20% chance to form ItemType-only preference
	PrefFormationWeightColor     = 0.20 // 20% chance to form Color-only preference
	PrefFormationWeightCombo     = 0.60 // 60% chance to form Combo preference
)

// PoisonCombo represents a food type + color combination
type PoisonCombo struct {
	ItemType string
	Color    types.Color
}

// AllPoisonCombos returns all possible poisonable combinations
func AllPoisonCombos() []PoisonCombo {
	return []PoisonCombo{
		{"berry", types.ColorRed},
		{"berry", types.ColorBlue},
		{"mushroom", types.ColorBrown},
		{"mushroom", types.ColorWhite},
		{"mushroom", types.ColorRed},
	}
}

// HealingCombo represents a food type + color combination that can heal
type HealingCombo struct {
	ItemType string
	Color    types.Color
}

// AllHealingCombos returns all possible healing combinations
// Uses same pool as poison - actual healing combos chosen at world gen (no overlap)
func AllHealingCombos() []HealingCombo {
	return []HealingCombo{
		{"berry", types.ColorRed},
		{"berry", types.ColorBlue},
		{"mushroom", types.ColorBrown},
		{"mushroom", types.ColorWhite},
		{"mushroom", types.ColorRed},
	}
}
