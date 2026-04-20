package config

import (
	"time"
)

const (
	MapWidth  = 58
	MapHeight = 58

	ItemSpawnCount   = 20
	FlowerSpawnCount = 20
	SpringCount      = 2
	LeafPileCount    = 4
	PondMinCount     = 1
	PondMaxCount     = 5
	PondMinSize      = 4
	PondMaxSize      = 16
	ClayMinCount     = 6
	ClayMaxCount     = 10
	ClayLooseItems   = 2 // loose clay items spawned on clay tiles at world gen (min; max is +1)
	UpdateInterval   = 150 * time.Millisecond

	// Symbols
	CharRobot       = '@'
	CharBerry       = '●'
	CharMushroom    = '♠'
	CharFlower      = '✿'
	CharGourd       = 'G'
	CharVessel      = 'U'
	CharSpring      = '☉'
	CharLeafPile    = '#'
	CharWater       = '▓'
	CharStick       = '/'
	CharNut         = 'o'
	CharShell       = '<'
	CharHoe         = 'L'
	CharSeed        = '.'
	CharTilledSoil  = '═'
	CharGrass       = 'W'
	CharBundle      = 'X'
	CharSprout      = '𖧧'
	CharSleeping    = 'z'
	CharClay        = '☗'
	CharClayTile    = '░'
	CharBrick       = '▬'
	CharFence       = '╬'
	CharHutCornerTL = '┏'
	CharHutCornerTR = '┓'
	CharHutCornerBL = '┗'
	CharHutCornerBR = '┛'
	CharHutEdgeH    = '━'
	CharHutEdgeV    = '┃'
	CharHutDoor     = '≡'
	CharHutTDown    = '┳'
	CharHutTUp      = '┻'
	CharHutTRight   = '┣'
	CharHutTLeft    = '┫'
	CharHutCross    = '╋'

	// Speed system
	BaseSpeed              = 50 // baseline speed (0-100 scale)
	MinSpeed               = 5  // minimum speed floor
	PoisonSpeedPenalty     = 25
	ParchedSpeedPenalty    = 10 // thirst >= 90
	DehydratedSpeedPenalty = 10 // thirst >= 100 (additional)
	VeryTiredSpeedPenalty  = 10 // energy <= 25
	ExhaustedSpeedPenalty  = 10 // energy <= 10 (additional)

	// Survival mechanics
	// Time scale: 1 game second = 12 world minutes, 1 world day = 120 game seconds
	PoisonDuration        = 20.0 // seconds (~4 world hours)
	HungerIncreaseRate    = 0.14 // per second (starving in ~6 world days)
	ThirstIncreaseRate    = 0.28 // per second (dehydrated in ~3 world days)
	EnergyDecreaseRate    = 0.5  // per second (base rate)
	EnergyMovementDrain   = 0.2  // additional per movement tick
	StarvationDamageRate  = 0.5  // health per second
	DehydrationDamageRate = 0.5  // health per second
	PoisonDamageRate      = 0.33 // health per second
	// Satiation tiers — see MealSize below for per-food satiation + duration
	DrinkThirstReduction    = 20.0 // thirst reduced per drink
	BedEnergyRestoreRate    = 2.86 // energy per second in bed (~7 world hours to full)
	GroundEnergyRestoreRate = 1.67 // energy per second on ground (~12 world hours to full)
	SatisfactionCooldown    = 5.0  // seconds (~1 world hour) before stat starts changing after reaching optimal
	// Action duration tiers
	// TODO: Consider Extra Short and Extra Long tiers as more actions are added
	ActionDurationShort  = 0.83 // seconds (~10 world minutes) for eat, drink, pickup
	ActionDurationMedium = 4.0  // seconds (~48 world minutes) for till soil, look
	ActionDurationLong   = 10.0 // seconds (~2 world hours) for craft hoe, craft vessel

	// Idle activities (looking, talking)
	IdleCooldown = 5.0 // seconds (~1 world hour) between idle activity attempts
	LookDuration = 4.0 // seconds (~48 world minutes) to complete looking at an item
	TalkDuration = 5.0 // seconds (~1 world hour) to complete a conversation

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

	// Sprout maturation tiers
	SproutDurationFast   = 120.0 // ~1 world day (mushroom)
	SproutDurationMedium = 360.0 // ~3 world days (berry, flower)
	SproutDurationSlow   = 600.0 // ~5 world days (gourd)

	// Reproduction interval tiers (base seconds between spawn attempts per plant)
	ReproductionFast       = 12.0  // berry: ~2 world days per plant
	ReproductionMedium     = 18.0  // mushroom, flower: ~3 world days per plant (current rate)
	ReproductionSlow       = 30.0  // gourd: ~5 world days per plant
	TilledGrowthMultiplier = 1.25  // 25% faster growth on tilled soil
	WetGrowthMultiplier    = 1.25  // 25% faster growth on wet tiles
	WateredTileDuration    = 360.0 // 3 world days (360 game seconds) until manual watering wears off

	// Healing
	HealAmount = 20.0 // health restored by healing items (instant)

	// Item lifecycle
	ItemSpawnChance           = 0.50 // 50% chance per spawn opportunity
	ItemSpawnMaxDensity       = 0.50 // max 50% of map coordinates occupied by items
	LifecycleIntervalVariance = 0.20 // ±20% randomization for spawn/death timers

	// Preference formation
	PrefFormationChanceMiserable = 0.10 // 10% chance when Miserable
	PrefFormationChanceUnhappy   = 0.05 // 5% chance when Unhappy
	PrefFormationChanceHappy     = 0.05 // 5% chance when Happy
	PrefFormationChanceJoyful    = 0.10 // 10% chance when Joyful
	PrefFormationWeightSingle    = 0.30 // 30% chance to form single-attribute preference
	PrefFormationWeightCombo     = 0.70 // 70% chance to form combo preference (2+ attributes)

	// Know-how discovery (Joyful mood rate)
	// Happy mood uses 20% of this rate. Neutral and below: 0%.
	// Set high (50%) for testing. For gameplay balance, use 5% (Happy: 1%).
	KnowHowDiscoveryChance = 0.50

	// Variety generation
	VarietyDivisor        = 4    // varietyCount = max(2, spawnCount / divisor)
	VarietyMinCount       = 2    // minimum varieties per item type
	VarietyPoisonPercent  = 0.20 // 20% of edible varieties are poisonous
	VarietyHealingPercent = 0.20 // 20% of edible varieties are healing

	// Ground spawning (sticks, nuts, shells)
	GroundSpawnInterval = 600.0 // ~5 world days between spawns per item type (±LifecycleIntervalVariance)

	// Bundle defaults
	DefaultMaxBundleSize = 6

	// Auto-save
	AutoSaveInterval = 60.0 // seconds of game time between auto-saves

	// Order abandonment cooldown (one world day = 2 real minutes = 120 seconds)
	OrderAbandonCooldown = 120.0

	// Food seeking - gradient scoring
	// Score = (NetPreference × PrefWeight) - (Distance × DistWeight)
	// At Moderate: only consider items with NetPreference >= 0
	// At Severe+: consider all items
	FoodSeekPrefWeightModerate = 20.0 // Strong preference influence at moderate hunger
	FoodSeekPrefWeightSevere   = 5.0  // Moderate preference influence at severe hunger
	FoodSeekPrefWeightCrisis   = 0.0  // No preference influence, just distance
	FoodSeekDistWeight         = 1.0  // Distance penalty per tile (base, used by vessel scoring)
	FoodSeekDistWeightModerate = 1.0  // Distance weight at moderate hunger
	FoodSeekDistWeightSevere   = 1.5  // Distance weight at severe hunger (fit can overcome distance)
	FoodSeekDistWeightCrisis   = 3.0  // Distance weight at crisis hunger (nearest food wins)

	// Item seeking - preference-weighted scoring (DD-52)
	// Score = (weightedPref × PrefWeight) - (dist × DistWeight)
	// No urgency tiers — item seeking doesn't vary by need level
	ItemSeekPrefWeight = 20.0 // How many tiles one preference point is worth
	ItemSeekDistWeight = 1.0  // Distance penalty per tile

	// Healing bonus in food selection - when hungry AND hurt, known healing items score higher
	// Only applies when health tier >= Mild and character knows item is healing
	HealingBonusMild     = 5.0  // Bonus when health at Mild tier
	HealingBonusModerate = 10.0 // Bonus when health at Moderate tier
	HealingBonusSevere   = 20.0 // Bonus when health at Severe tier
	HealingBonusCrisis   = 40.0 // Bonus when health at Crisis tier
)

// LifecycleConfig defines spawn and death intervals for an item type
type LifecycleConfig struct {
	SpawnInterval float64 // base seconds between spawn attempts (multiplied by initial item count)
	DeathInterval float64 // base seconds until death (0 = immortal, multiplied by initial item count)
}

// ItemLifecycle maps item types to their lifecycle configuration
// Time scale: 1 game second = 12 world minutes, 1 world day = 120 game seconds
// NOTE: These values are multiplied by initialItemCount in lifecycle.go to spread
// spawn attempts across all plants. To get target world-time intervals, divide by
// expected item count (typically 20). E.g., 18 * 20 = 360s = 3 world days.
var ItemLifecycle = map[string]LifecycleConfig{
	"berry":    {SpawnInterval: ReproductionFast, DeathInterval: 0},      // ~2 world days between spawns, immortal until eaten
	"mushroom": {SpawnInterval: ReproductionMedium, DeathInterval: 0},    // ~3 world days between spawns, immortal until eaten
	"flower":   {SpawnInterval: ReproductionMedium, DeathInterval: 48.0}, // ~3 world days between spawns, dies after ~8 world days
	"gourd":    {SpawnInterval: ReproductionSlow, DeathInterval: 0},      // ~5 world days between spawns, immortal until eaten
	"grass":    {SpawnInterval: ReproductionFast, DeathInterval: 48.0},   // ~2 world days between spawns, dies after ~8 world days
}

// ExtractableTypes defines which plant types yield seeds via extraction
var ExtractableTypes = map[string]bool{
	"flower": true,
	"grass":  true,
}

// StackSize maps item types to how many fit in one stack (for vessel storage)
var StackSize = map[string]int{
	"berry":    20,
	"mushroom": 10,
	"flower":   10,
	"gourd":    1,
	"nut":      10,
	"shell":    4,
	"seed":     20,
	"liquid":   4,
}

// GetStackSize returns the stack size for an item type, defaulting to 1 if not defined
func GetStackSize(itemType string) int {
	if size, ok := StackSize[itemType]; ok {
		return size
	}
	return 1
}

// MaxBundleSize maps bundleable item types to their max bundle count.
// Used only for bundle logic (merging, capacity checks, rendering).
// For vessel exclusion, see VesselExcludedTypes.
var MaxBundleSize = map[string]int{
	"stick": 6,
	"grass": 6,
}

// VesselExcludedTypes is the set of item types that cannot be stored in vessels.
// Distinct from MaxBundleSize — items can be vessel-excluded without being bundleable (clay, brick).
var VesselExcludedTypes = map[string]bool{
	"stick": true,
	"grass": true,
	"clay":  true,
	"brick": true,
}

// GroundSpawnCount maps ground-spawned item types to their initial world-gen count.
// Types not listed default to 1 (via GetGroundSpawnCount).
var GroundSpawnCount = map[string]int{
	"stick": 6,
	"nut":   6,
}

// GetGroundSpawnCount returns the initial spawn count for a ground item type, defaulting to 1
func GetGroundSpawnCount(itemType string) int {
	if count, ok := GroundSpawnCount[itemType]; ok {
		return count
	}
	return 1
}

// MealSize defines the satiation and eating duration for a food tier
type MealSize struct {
	Satiation float64 // hunger reduction amount
	Duration  float64 // game seconds to complete eating
}

// Meal size tier constants
// Time scale: 1 game second = 12 world minutes
var (
	MealSizeFeast = MealSize{Satiation: 50.0, Duration: 3.75}  // ~45 world mins (gourd)
	MealSizeMeal  = MealSize{Satiation: 25.0, Duration: 1.25}  // ~15 world mins (mushroom)
	MealSizeSnack = MealSize{Satiation: 10.0, Duration: 0.417} // ~5 world mins (berry, nut)
)

// ItemMealSize maps item types to their meal size configuration
var ItemMealSize = map[string]MealSize{
	"gourd":    MealSizeFeast,
	"mushroom": MealSizeMeal,
	"berry":    MealSizeSnack,
	"nut":      MealSizeSnack,
}

// GetMealSize returns the meal size for an item type, defaulting to Meal tier
func GetMealSize(itemType string) MealSize {
	if ms, ok := ItemMealSize[itemType]; ok {
		return ms
	}
	return MealSizeMeal
}

// SproutDurationTier maps item types to their sprout maturation duration
var SproutDurationTier = map[string]float64{
	"mushroom": SproutDurationFast,
	"berry":    SproutDurationMedium,
	"flower":   SproutDurationMedium,
	"gourd":    SproutDurationSlow,
	"grass":    SproutDurationFast,
}

// GetSproutDuration returns the sprout maturation duration for an item type, defaulting to Medium tier
func GetSproutDuration(itemType string) float64 {
	if duration, ok := SproutDurationTier[itemType]; ok {
		return duration
	}
	return SproutDurationMedium
}
