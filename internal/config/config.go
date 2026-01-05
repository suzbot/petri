package config

import (
	"time"
)

const (
	MapWidth  = 60
	MapHeight = 60

	ItemSpawnCount   = 20
	FlowerSpawnCount = 20
	SpringCount      = 2
	LeafPileCount    = 4
	UpdateInterval = 150 * time.Millisecond

	// Symbols
	CharRobot    = '@'
	CharBerry    = '*'
	CharMushroom = '^'
	CharFlower   = '❀'
	CharSpring   = '~'
	CharLeafPile = '&'
	CharSleeping = 'z'

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
	SatisfactionCooldown = 5.0 // seconds before stat starts changing after reaching optimal
	ActionDuration       = 1.5 // seconds for consume/drink/sleep actions to complete

	// Looking activity
	LookChance   = 0.50  // 50% chance to look when idle
	LookDuration = 3.0   // seconds to complete looking at an item
	LookCooldown = 10.0  // seconds before can look again after completing a look

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

	// Item spawning
	ItemSpawnChance           = 0.50 // 50% chance per spawn opportunity
	ItemSpawnIntervalBase     = 8.0  // seconds, multiplied by initial item count
	ItemSpawnIntervalVariance = 0.20 // ±20% randomization
	ItemSpawnMaxDensity       = 0.50 // max 50% of map coordinates occupied by items

	// Preference formation (values inflated for testing - see CLAUDE.md balance tuning)
	PrefFormationChanceMiserable = 0.20 // 20% chance when Miserable
	PrefFormationChanceUnhappy   = 0.10 // 10% chance when Unhappy
	PrefFormationChanceHappy     = 0.10 // 10% chance when Happy
	PrefFormationChanceJoyful    = 0.20 // 20% chance when Joyful
	PrefFormationWeightSingle    = 0.30 // 30% chance to form single-attribute preference
	PrefFormationWeightCombo     = 0.70 // 70% chance to form Combo preference (2+ attributes)

	// Variety generation
	VarietyDivisor        = 4    // varietyCount = max(2, spawnCount / divisor)
	VarietyMinCount       = 2    // minimum varieties per item type
	VarietyPoisonPercent  = 0.20 // 20% of edible varieties are poisonous
	VarietyHealingPercent = 0.20 // 20% of edible varieties are healing
)
