package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// Intent Priority and Fallback
// =============================================================================

func TestCalculateIntent_HighestTierPrioritized(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 90 // Severe
	char.Thirst = 50 // Mild
	char.Energy = 50 // Mild

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{entity.NewBerry(5, 5, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.DrivingStat != types.StatHunger {
		t.Errorf("DrivingStat: got %q, want %q", intent.DrivingStat, types.StatHunger)
	}
}

func TestCalculateIntent_TieBreakerFavorsThirstOverHunger(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 75 // Moderate
	char.Thirst = 75 // Moderate (same tier)
	char.Energy = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddFeature(entity.NewSpring(5, 5))
	items := []*entity.Item{entity.NewBerry(10, 10, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.DrivingStat != types.StatThirst {
		t.Errorf("DrivingStat: got %q, want %q (tie-breaker)", intent.DrivingStat, types.StatThirst)
	}
}

func TestCalculateIntent_TieBreakerFavorsHungerOverEnergy(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 75 // Moderate
	char.Thirst = 0
	char.Energy = 25 // Moderate (same tier)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddFeature(entity.NewLeafPile(5, 5))
	items := []*entity.Item{entity.NewBerry(10, 10, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.DrivingStat != types.StatHunger {
		t.Errorf("DrivingStat: got %q, want %q (tie-breaker)", intent.DrivingStat, types.StatHunger)
	}
}

func TestCalculateIntent_FallsBackWhenHighestCantBeFulfilled(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 75 // Moderate
	char.Thirst = 90 // Severe (highest, but no water)
	char.Energy = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// No springs!
	items := []*entity.Item{entity.NewBerry(5, 5, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent after fallback, got nil")
	}
	if intent.DrivingStat != types.StatHunger {
		t.Errorf("DrivingStat after fallback: got %q, want %q", intent.DrivingStat, types.StatHunger)
	}
}

func TestCalculateIntent_ReturnsNilWhenNoNeeds(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 0  // TierNone
	char.Thirst = 0  // TierNone
	char.Energy = 100 // TierNone

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	intent := CalculateIntent(char, nil, gameMap, nil, nil)

	if intent != nil {
		t.Errorf("Expected nil intent when no needs, got %+v", intent)
	}
}

func TestCalculateIntent_DeadCharacterReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsDead = true
	char.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{entity.NewBerry(5, 5, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent != nil {
		t.Error("Dead character should return nil intent")
	}
}

func TestCalculateIntent_SleepingCharacterReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsSleeping = true
	char.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{entity.NewBerry(5, 5, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent != nil {
		t.Error("Sleeping character should return nil intent")
	}
}

func TestCalculateIntent_FrustratedCharacterReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsFrustrated = true
	char.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{entity.NewBerry(5, 5, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent != nil {
		t.Error("Frustrated character should return nil intent")
	}
}

// =============================================================================
// Frustration Trigger
// =============================================================================

func TestCalculateIntent_FailedIntentIncrementsCounterAtSevere(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 90  // Severe
	char.Thirst = 90  // Severe
	char.Energy = 100 // TierNone (so ground sleep not available as fallback)
	char.FailedIntentCount = 0

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// No food, no water - hunger and thirst unfulfillable

	CalculateIntent(char, nil, gameMap, nil, nil)

	if char.FailedIntentCount != 1 {
		t.Errorf("FailedIntentCount: got %d, want 1", char.FailedIntentCount)
	}
}

func TestCalculateIntent_FailedIntentDoesNotIncrementAtLowerTiers(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 75 // Moderate (not Severe)
	char.Thirst = 75 // Moderate
	char.Energy = 25 // Moderate
	char.FailedIntentCount = 0

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// No resources

	CalculateIntent(char, nil, gameMap, nil, nil)

	if char.FailedIntentCount != 0 {
		t.Errorf("FailedIntentCount should stay 0 at Moderate tier, got %d", char.FailedIntentCount)
	}
}

func TestCalculateIntent_FrustrationTriggersAtThreshold(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 90 // Severe
	char.FailedIntentCount = config.FrustrationThreshold - 1

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// No food

	CalculateIntent(char, nil, gameMap, nil, nil)

	if !char.IsFrustrated {
		t.Error("Character should become frustrated at threshold")
	}
	if char.FrustrationTimer != config.FrustrationDuration {
		t.Errorf("FrustrationTimer: got %.2f, want %.2f", char.FrustrationTimer, config.FrustrationDuration)
	}
	if char.FailedIntentCount != 0 {
		t.Errorf("FailedIntentCount should reset to 0, got %d", char.FailedIntentCount)
	}
}

func TestCalculateIntent_SuccessfulIntentResetsFailureCounter(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 90
	char.FailedIntentCount = 2

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{entity.NewBerry(5, 5, types.ColorRed, false, false)}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected successful intent")
	}
	if char.FailedIntentCount != 0 {
		t.Errorf("FailedIntentCount should reset on success, got %d", char.FailedIntentCount)
	}
}

// =============================================================================
// Food Target Selection - Gradient Scoring
// =============================================================================

// Crisis tier (90+): Pick nearest edible item regardless of preference
func TestFindFoodTarget_CrisisPicksNearest(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 95           // Crisis
	char.SetPos(types.Position{X: 0, Y: 0})

	// Disliked but close vs liked but far
	// Add a dislike preference for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	nearMushroom := entity.NewMushroom(2, 2, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	farRedBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{farRedBerry, nearMushroom}

	result := FindFoodTarget(char, items)

	if result.Item != nearMushroom {
		t.Error("Crisis should pick nearest regardless of preference")
	}
}

// Severe tier (75-89): Uses gradient scoring, considers all items including disliked
func TestFindFoodTarget_SevereUsesGradientScoring(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80           // Severe
	char.SetPos(types.Position{X: 0, Y: 0})

	// Liked item far away vs neutral item close
	// Score(redBerry) = 5*2 - 20 = -10
	// Score(brownMushroom) = 5*0 - 6 = -6
	// brownMushroom wins (less negative score)
	redBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	brownMushroom := entity.NewMushroom(3, 3, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{redBerry, brownMushroom}

	result := FindFoodTarget(char, items)

	if result.Item != brownMushroom {
		t.Errorf("Severe should use gradient - expected closer neutral item, got %v", result.Item)
	}
}

func TestFindFoodTarget_SevereConsidersDislikedItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80           // Severe
	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Only disliked food available
	items := []*entity.Item{entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)}

	result := FindFoodTarget(char, items)

	if result.Item == nil {
		t.Error("Severe should consider disliked items when nothing else available")
	}
}

func TestFindFoodTarget_SeverePrefersLikedOverDisliked(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80           // Severe
	char.SetPos(types.Position{X: 0, Y: 0})
	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Liked item same distance as disliked
	// Score(blueBerry at 5,5) = 5*1 - 10 = -5 (likes berries)
	// Score(brownMushroom at 5,5) = 5*(-1) - 10 = -15 (dislikes mushrooms)
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	brownMushroom := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{brownMushroom, blueBerry}

	result := FindFoodTarget(char, items)

	if result.Item != blueBerry {
		t.Error("Severe should prefer liked over disliked at same distance")
	}
}

// Moderate tier (50-74): Uses gradient scoring, filters out disliked items (NetPref < 0)
func TestFindFoodTarget_ModerateUsesGradientScoring(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 0, Y: 0})

	// Perfect match far vs partial match close
	// Score(redBerry at 10,10) = 20*2 - 20 = 20
	// Score(blueBerry at 3,3) = 20*1 - 6 = 14
	// redBerry wins despite distance
	redBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	blueBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)

	items := []*entity.Item{blueBerry, redBerry}

	result := FindFoodTarget(char, items)

	if result.Item != redBerry {
		t.Errorf("Moderate should use gradient - expected higher preference item, got %v", result.Item)
	}
}

func TestFindFoodTarget_ModerateFiltersDislikedItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Only disliked food available
	items := []*entity.Item{entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)}

	result := FindFoodTarget(char, items)

	if result.Item != nil {
		t.Error("Moderate should filter out disliked items (return nil)")
	}
}

func TestFindFoodTarget_ModerateTakesNeutralItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate

	// Neutral food (no preference match, but not disliked either)
	brownMushroom := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{brownMushroom}

	result := FindFoodTarget(char, items)

	if result.Item != brownMushroom {
		t.Error("Moderate should accept neutral items (NetPref >= 0)")
	}
}

func TestFindFoodTarget_ModeratePrefersLikedOverNeutral(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 0, Y: 0})

	// Liked item same distance as neutral
	// Score(blueBerry at 5,5) = 20*1 - 10 = 10
	// Score(brownMushroom at 5,5) = 20*0 - 10 = -10
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	brownMushroom := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{brownMushroom, blueBerry}

	result := FindFoodTarget(char, items)

	if result.Item != blueBerry {
		t.Error("Moderate should prefer liked over neutral at same distance")
	}
}

// Distance tiebreaker: when gradient scores are equal, prefer closer item
func TestFindFoodTarget_DistanceTiebreaker(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 0, Y: 0})

	// Two items with same NetPreference at different distances
	nearRedBerry := entity.NewBerry(2, 2, types.ColorRed, false, false)
	farRedBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{farRedBerry, nearRedBerry}

	result := FindFoodTarget(char, items)

	if result.Item != nearRedBerry {
		t.Error("Should prefer closer item when preference is equal")
	}
}

// Edge case: no edible items
func TestFindFoodTarget_ReturnsNilForNoEdibleItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 90

	// Only non-edible items (flowers)
	items := []*entity.Item{entity.NewFlower(5, 5, types.ColorRed)}

	result := FindFoodTarget(char, items)

	if result.Item != nil {
		t.Error("Should return nil when no edible items exist")
	}
}

func TestFindFoodTarget_VesselNotEdible(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 95 // Crisis - would eat anything edible

	// Vessel is not edible (crafted from gourd)
	gourd := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)

	items := []*entity.Item{vessel}

	result := FindFoodTarget(char, items)

	if result.Item != nil {
		t.Error("Vessel should not be edible - got a food target when none expected")
	}
}

func TestFindFoodIntent_CarriedVesselNotEaten(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 95 // Crisis - would eat anything edible from inventory

	// Character is carrying a vessel (not edible)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	char.AddToInventory(vessel)

	cpos := char.Pos()
	intent := findFoodIntent(char, cpos, nil, entity.TierCrisis, nil)

	if intent != nil {
		t.Error("Should not create eat intent for non-edible carried vessel")
	}
}

// =============================================================================
// Healing Bonus in Food Selection (Sub-phase F)
// =============================================================================

func TestFindFoodTarget_HealingBonus_OnlyWhenKnown(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60        // Moderate hunger
	char.Health = 50        // Moderate health (hurt)
	char.SetPos(types.Position{X: 0, Y: 0})
	// No healing knowledge

	// Healing item vs regular item at same distance
	healingBerry := entity.NewBerry(5, 5, types.ColorBlue, false, true) // is healing
	regularBerry := entity.NewBerry(5, 0, types.ColorBlue, false, false)

	items := []*entity.Item{healingBerry, regularBerry}

	result := FindFoodTarget(char, items)

	// Without knowledge, healing item shouldn't get bonus
	// regularBerry is at distance 5, healingBerry is at distance 10
	// Both have same preference, so closer one wins
	if result.Item != regularBerry {
		t.Error("Without healing knowledge, should not apply healing bonus")
	}
}

func TestFindFoodTarget_HealingBonus_OnlyWhenHurt(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60    // Moderate hunger
	char.Health = 100   // Full health (not hurt)
	char.SetPos(types.Position{X: 0, Y: 0})

	// Give character healing knowledge for blue berries
	char.Knowledge = append(char.Knowledge, entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	})

	// Healing item far vs regular item close
	healingBerry := entity.NewBerry(10, 10, types.ColorBlue, false, true)
	regularBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)

	items := []*entity.Item{healingBerry, regularBerry}

	result := FindFoodTarget(char, items)

	// At full health, no bonus - closer item wins
	if result.Item != regularBerry {
		t.Error("At full health, should not apply healing bonus")
	}
}

func TestFindFoodTarget_HealingBonus_PrefersKnownHealingWhenHurt(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate hunger
	char.Health = 50           // Moderate health tier (hurt)
	char.SetPos(types.Position{X: 0, Y: 0})

	// Give character healing knowledge for blue berries
	char.Knowledge = append(char.Knowledge, entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	})

	// Known healing (blue berry) vs unknown (red berry) - only blue matches knowledge
	// At Moderate hunger: prefWeight=20, distWeight=1
	// At Moderate health: healingBonus=10
	// Score(blueBerry at 5,5) = 20*1 + 10 - 10 = 20 (likes berries + healing bonus)
	// Score(redBerry at 3,3) = 20*2 - 6 = 34 (likes berries AND red)
	// Red berry would win without healing bonus, so we need to adjust distances
	// Let's make red berry farther so healing bonus can tip the balance:
	// Score(blueBerry at 3,3) = 20*1 + 10 - 6 = 24 (likes berries + healing bonus)
	// Score(redBerry at 6,6) = 20*2 - 12 = 28 (likes berries AND red)
	// Still loses... make red even farther:
	// Score(blueBerry at 3,3) = 20*1 + 10 - 6 = 24
	// Score(redBerry at 8,8) = 20*2 - 16 = 24 (tie, blue wins on distance tiebreaker? No, same score means distance tiebreaker, red at 16 vs blue at 6)
	// Actually wait - we need healing bonus to OVERCOME the preference difference
	// Red has +2 preference, blue has +1, so red is 20 points ahead
	// Healing bonus at Moderate is 10, not enough to overcome
	// At Severe health, bonus is 20, which would tie
	// At Crisis health, bonus is 40, which would overcome

	// For this test, let's use Severe health to make it clearly work:
	char.Health = 25 // Severe health tier - healing bonus = 20

	// Score(blueBerry at 5,5) = 20*1 + 20 - 10 = 30 (likes berries + healing bonus)
	// Score(redBerry at 5,5) = 20*2 - 10 = 30 (tie)
	// Let's make red slightly farther:
	// Score(blueBerry at 5,5) = 20*1 + 20 - 10 = 30
	// Score(redBerry at 6,6) = 20*2 - 12 = 28
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, true)
	redBerry := entity.NewBerry(6, 6, types.ColorRed, false, false)

	items := []*entity.Item{redBerry, blueBerry}

	result := FindFoodTarget(char, items)

	if result.Item != blueBerry {
		t.Errorf("When hurt and has knowledge, should prefer known healing item, got %v", result.Item.Description())
	}
}

func TestFindFoodTarget_HealingBonus_ScalesWithHealthTier(t *testing.T) {
	t.Parallel()

	// Test that at Crisis health, healing bonus is large enough to overcome preference difference
	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate hunger
	char.Health = 10           // Crisis health tier
	char.SetPos(types.Position{X: 0, Y: 0})

	// Give character healing knowledge for blue berries
	char.Knowledge = append(char.Knowledge, entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	})

	// At Crisis health, bonus=40 should overcome preference difference
	// Red berry: +2 preference (likes berries AND red), Blue berry: +1 (likes berries only)
	// At Moderate hunger: prefWeight=20, distWeight=1
	// At Crisis health: healingBonus=40
	// Score(blueBerry at 8,8) = 20*1 + 40 - 16 = 44 (likes berries + crisis healing bonus)
	// Score(redBerry at 3,3) = 20*2 - 6 = 34 (likes berries AND red)
	// Blue wins due to massive healing bonus!
	blueBerry := entity.NewBerry(8, 8, types.ColorBlue, false, true)
	redBerry := entity.NewBerry(3, 3, types.ColorRed, false, false)

	items := []*entity.Item{redBerry, blueBerry}

	result := FindFoodTarget(char, items)

	if result.Item != blueBerry {
		t.Errorf("At Crisis health, larger healing bonus should win, got %v", result.Item.Description())
	}
}

// =============================================================================
// Continue Intent
// =============================================================================

func TestContinueIntent_ContinuesIfTargetItemExists(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0})

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	char.Intent = &entity.Intent{
		Target:      types.Position{X: 1, Y: 0},
		Action:      entity.ActionMove,
		TargetItem:  item,
		DrivingStat: types.StatHunger,
		DrivingTier: 2,
	}

	intent := continueIntent(char, 0, 0, gameMap, nil)

	if intent == nil {
		t.Error("Should continue intent when target item exists")
	}
}

func TestContinueIntent_AbandonsIfTargetItemConsumed(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0})

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	// Item NOT added to map (simulates it was consumed)

	char.Intent = &entity.Intent{
		Target:      types.Position{X: 1, Y: 0},
		Action:      entity.ActionMove,
		TargetItem:  item,
		DrivingStat: types.StatHunger,
		DrivingTier: 2,
	}

	intent := continueIntent(char, 0, 0, gameMap, nil)

	if intent != nil {
		t.Error("Should abandon intent when target item no longer exists")
	}
}

func TestContinueIntent_AbandonsIfAllSpringAdjacentTilesBlocked(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0})

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)

	// Block all 4 cardinal-adjacent tiles of the spring
	gameMap.AddCharacter(entity.NewCharacter(2, 5, 4, "N", "berry", types.ColorBlue))
	gameMap.AddCharacter(entity.NewCharacter(3, 6, 5, "E", "berry", types.ColorBlue))
	gameMap.AddCharacter(entity.NewCharacter(4, 5, 6, "S", "berry", types.ColorBlue))
	gameMap.AddCharacter(entity.NewCharacter(5, 4, 5, "W", "berry", types.ColorBlue))

	char.Intent = &entity.Intent{
		Target:        types.Position{X: 1, Y: 0},
		Action:        entity.ActionMove,
		TargetFeature: spring,
		DrivingStat:   types.StatThirst,
		DrivingTier:   2,
	}

	intent := continueIntent(char, 0, 0, gameMap, nil)

	if intent != nil {
		t.Error("Should abandon intent when all spring adjacent tiles are blocked")
	}
}

func TestContinueIntent_AbandonsIfTargetBedOccupied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0})

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	bed := entity.NewLeafPile(5, 5)
	gameMap.AddFeature(bed)

	// Another character at the bed
	otherChar := entity.NewCharacter(2, 5, 5, "Other", "berry", types.ColorBlue)
	gameMap.AddCharacter(otherChar)

	char.Intent = &entity.Intent{
		Target:        types.Position{X: 1, Y: 0},
		Action:        entity.ActionMove,
		TargetFeature: bed,
		DrivingStat:   types.StatEnergy,
		DrivingTier:   2,
	}

	intent := continueIntent(char, 0, 0, gameMap, nil)

	if intent != nil {
		t.Error("Should abandon intent when target bed is occupied")
	}
}

func TestContinueIntent_OwnPositionDoesNotAbandon(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 5, Y: 5})

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)
	gameMap.AddCharacter(char) // Character is at the spring

	char.Intent = &entity.Intent{
		Target:        types.Position{X: 5, Y: 5},
		Action:        entity.ActionDrink,
		TargetFeature: spring,
		DrivingStat:   types.StatThirst,
		DrivingTier:   2,
	}

	intent := continueIntent(char, 5, 5, gameMap, nil)

	if intent == nil {
		t.Error("Should not abandon intent when character is at their own target")
	}
	if intent.Action != entity.ActionDrink {
		t.Errorf("Action: got %d, want ActionDrink", intent.Action)
	}
}

func TestContinueIntent_PickupNotConvertedToLook(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 4, Y: 5}) // Adjacent to item at (5,5)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	// ActionPickup intent (foraging) should NOT be converted to ActionLook
	char.Intent = &entity.Intent{
		Target:     types.Position{X: 5, Y: 5},
		Action:     entity.ActionPickup,
		TargetItem: item,
		// No DrivingStat - idle activity
	}

	intent := continueIntent(char, 4, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Should continue pickup intent")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Action: got %d, want ActionPickup (should NOT convert to Look)", intent.Action)
	}
}

func TestContinueIntent_PickupContinuesToItem(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0}) // Not adjacent to item at (5,5)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	char.Intent = &entity.Intent{
		Target:     types.Position{X: 1, Y: 0},
		Action:     entity.ActionPickup,
		TargetItem: item,
	}

	intent := continueIntent(char, 0, 0, gameMap, nil)

	if intent == nil {
		t.Fatal("Should continue pickup intent when item exists")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Action: got %d, want ActionPickup", intent.Action)
	}
	if intent.TargetItem != item {
		t.Error("Should maintain same target item")
	}
}

// =============================================================================
// findHealingIntent (E2: Health-driven seeking)
// =============================================================================

func TestFindHealingIntent_NoKnowledge_ReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{} // No healing knowledge
	char.SetPos(types.Position{X: 0, Y: 0})

	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorBlue, false, true), // Healing but unknown
	}

	intent := findHealingIntent(char, types.Position{X: 0, Y: 0}, items, entity.TierModerate, nil)

	if intent != nil {
		t.Error("findHealingIntent() should return nil when character has no healing knowledge")
	}
}

func TestFindHealingIntent_HasKnowledge_FindsNearestKnown(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	// Character knows blue berries are healing
	healingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{healingKnowledge}
	char.SetPos(types.Position{X: 0, Y: 0})

	blueBerryNear := entity.NewBerry(2, 2, types.ColorBlue, false, true)
	blueBerryFar := entity.NewBerry(10, 10, types.ColorBlue, false, true)
	redBerry := entity.NewBerry(1, 1, types.ColorRed, false, false) // Unknown, closer
	items := []*entity.Item{blueBerryFar, redBerry, blueBerryNear}

	intent := findHealingIntent(char, types.Position{X: 0, Y: 0}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("findHealingIntent() should return intent when known healing item exists")
	}
	if intent.TargetItem != blueBerryNear {
		t.Error("findHealingIntent() should target nearest KNOWN healing item")
	}
	if intent.DrivingStat != types.StatHealth {
		t.Errorf("DrivingStat: got %q, want %q", intent.DrivingStat, types.StatHealth)
	}
}

func TestFindHealingIntent_NoMatchingItems_ReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	// Character knows blue berries are healing
	healingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{healingKnowledge}
	char.SetPos(types.Position{X: 0, Y: 0})

	// Only red berries available (not known healing)
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorRed, false, false),
	}

	intent := findHealingIntent(char, types.Position{X: 0, Y: 0}, items, entity.TierModerate, nil)

	if intent != nil {
		t.Error("findHealingIntent() should return nil when no known healing items available")
	}
}

func TestFindHealingIntent_EmptyItemList_ReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	healingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{healingKnowledge}
	char.SetPos(types.Position{X: 0, Y: 0})

	intent := findHealingIntent(char, types.Position{X: 0, Y: 0}, []*entity.Item{}, entity.TierModerate, nil)

	if intent != nil {
		t.Error("findHealingIntent() should return nil with empty item list")
	}
}

// =============================================================================
// E3: Health in CalculateIntent Priority System
// =============================================================================

func TestCalculateIntent_HealthDrivesIntent_WhenHasHealingKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 25  // Severe tier
	char.Hunger = 0   // No need
	char.Thirst = 0   // No need
	char.Energy = 100 // No need

	// Character knows blue berries are healing
	healingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{healingKnowledge}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorBlue, false, true), // Known healing
	}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent when health is urgent and healing knowledge exists")
	}
	if intent.DrivingStat != types.StatHealth {
		t.Errorf("DrivingStat: got %q, want %q", intent.DrivingStat, types.StatHealth)
	}
}

func TestCalculateIntent_HealthIgnored_WhenNoHealingKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 25  // Severe tier
	char.Hunger = 75  // Moderate tier
	char.Thirst = 0
	char.Energy = 100
	char.Knowledge = []entity.Knowledge{} // No healing knowledge

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorBlue, false, true), // Healing but unknown
		entity.NewBerry(6, 6, types.ColorRed, false, false), // Regular food
	}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent for hunger")
	}
	// Should fall back to hunger since health can't be fulfilled (no knowledge)
	if intent.DrivingStat != types.StatHunger {
		t.Errorf("DrivingStat: got %q, want %q (health unfulfillable)", intent.DrivingStat, types.StatHunger)
	}
}

func TestCalculateIntent_HealthPriority_SameTierAsOtherStats(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 25 // Severe
	char.Hunger = 90 // Severe (same tier)
	char.Thirst = 0
	char.Energy = 100

	// Character knows blue berries are healing
	healingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{healingKnowledge}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorBlue, false, true), // Known healing
		entity.NewBerry(6, 6, types.ColorRed, false, false), // Regular food
	}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent")
	}
	// At same tier, tie-breaker order should be: Thirst > Hunger > Energy > Health
	// So hunger should win over health at same tier
	if intent.DrivingStat != types.StatHunger {
		t.Errorf("DrivingStat: got %q, want %q (tie-breaker)", intent.DrivingStat, types.StatHunger)
	}
}

func TestCalculateIntent_HealthWins_WhenHigherTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 10 // Crisis tier
	char.Hunger = 75 // Moderate tier
	char.Thirst = 0
	char.Energy = 100

	// Character knows blue berries are healing
	healingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{healingKnowledge}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorBlue, false, true), // Known healing
		entity.NewBerry(6, 6, types.ColorRed, false, false), // Regular food
	}

	intent := CalculateIntent(char, items, gameMap, nil, nil)

	if intent == nil {
		t.Fatal("Expected intent")
	}
	// Health at Crisis should beat Hunger at Moderate
	if intent.DrivingStat != types.StatHealth {
		t.Errorf("DrivingStat: got %q, want %q (higher tier)", intent.DrivingStat, types.StatHealth)
	}
}

// =============================================================================
// Eating from Inventory (5.3)
// =============================================================================

func TestFindFoodIntent_ReturnsConsumeIntent_WhenCarryingEdibleItem(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60 // Moderate tier - should trigger food seeking
	char.SetPos(types.Position{X: 5, Y: 5})

	// Character is carrying an edible item
	carriedBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.AddToInventory(carriedBerry)

	// Map item exists but is farther away
	mapItem := entity.NewBerry(10, 10, types.ColorRed, false, false)
	items := []*entity.Item{mapItem}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierMild, nil)

	if intent == nil {
		t.Fatal("Expected intent when carrying edible item")
	}
	if intent.Action != entity.ActionConsume {
		t.Errorf("Action: got %d, want ActionConsume (%d)", intent.Action, entity.ActionConsume)
	}
	if intent.TargetItem != carriedBerry {
		t.Error("TargetItem should be the carried item")
	}
	if intent.DrivingStat != types.StatHunger {
		t.Errorf("DrivingStat: got %q, want %q", intent.DrivingStat, types.StatHunger)
	}
}

func TestFindFoodIntent_IgnoresCarriedItem_WhenNotEdible(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60
	char.SetPos(types.Position{X: 5, Y: 5})

	// Character is carrying a non-edible item (flower)
	carriedFlower := entity.NewFlower(0, 0, types.ColorRed)
	char.AddToInventory(carriedFlower)

	// Map has edible item
	mapBerry := entity.NewBerry(6, 6, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierMild, nil)

	if intent == nil {
		t.Fatal("Expected intent when map has edible items")
	}
	// Should seek map item, not try to eat non-edible carried item
	if intent.Action == entity.ActionConsume {
		t.Error("Should not return ActionConsume for non-edible carried item")
	}
	if intent.TargetItem != mapBerry {
		t.Error("Should target map item when carried item is not edible")
	}
}

func TestFindFoodIntent_FallsBackToMapItems_WhenNotCarrying(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60
	char.SetPos(types.Position{X: 5, Y: 5})
	char.Inventory = nil // Not carrying anything

	mapBerry := entity.NewBerry(6, 6, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierMild, nil)

	if intent == nil {
		t.Fatal("Expected intent for map item")
	}
	if intent.Action == entity.ActionConsume {
		t.Error("Should not return ActionConsume when not carrying anything")
	}
	if intent.TargetItem != mapBerry {
		t.Error("Should target map item")
	}
}

func TestContinueIntent_ActionConsume_PreservesIntent(t *testing.T) {
	t.Parallel()

	// Regression test: ActionConsume intent should not be abandoned
	// because the carried item isn't on the map (it's in inventory)
	char := newTestCharacter()
	char.Hunger = 60
	char.SetPos(types.Position{X: 5, Y: 5})

	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.AddToInventory(carriedItem)

	// Set up an ActionConsume intent (eating from inventory)
	char.Intent = &entity.Intent{
		Target:      types.Position{X: 5, Y: 5},
		Action:      entity.ActionConsume,
		TargetItem:  carriedItem,
		DrivingStat: types.StatHunger,
		DrivingTier: entity.TierMild,
	}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	// continueIntent should preserve the ActionConsume intent
	// (not return nil because the item isn't on the map)
	result := continueIntent(char, 5, 5, gameMap, nil)

	if result == nil {
		t.Fatal("continueIntent should not abandon ActionConsume intent")
	}
	if result.Action != entity.ActionConsume {
		t.Errorf("Action: got %d, want ActionConsume", result.Action)
	}
	if result.TargetItem != carriedItem {
		t.Error("TargetItem should be preserved")
	}
}

// =============================================================================
// Foraging Filter - IsGrowing (Feature 3b)
// =============================================================================

func TestScoreForageItems_SkipsNonGrowingItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0})

	// Growing berry (should be targeted)
	growingBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	// growingBerry.Plant.IsGrowing is true by default from constructor

	// Non-growing berry (dropped item - should be skipped)
	droppedBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)
	droppedBerry.Plant.IsGrowing = false // Simulates picked up and dropped

	items := []*entity.Item{droppedBerry, growingBerry}

	result, _ := scoreForageItems(char, types.Position{X: 0, Y: 0}, items, nil) // nil vessel = no variety filter

	if result != growingBerry {
		t.Errorf("Expected growing berry, got %v", result)
	}
}

func TestScoreForageItems_ReturnsNilWhenOnlyNonGrowingItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 0, Y: 0})

	// Only non-growing items
	droppedBerry := entity.NewBerry(3, 3, types.ColorRed, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry}

	result, _ := scoreForageItems(char, types.Position{X: 0, Y: 0}, items, nil) // nil vessel = no variety filter

	if result != nil {
		t.Error("Should return nil when only non-growing items exist")
	}
}

// =============================================================================
// Unified Food Selection (Stage 5b) - Carried items use same scoring as map items
// =============================================================================

func TestFindFoodIntent_CarriedDislikedItem_FilteredAtModerate(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})

	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Carrying a disliked mushroom
	carriedMushroom := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	char.AddToInventory(carriedMushroom)

	// Liked berry on map
	mapBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent when map has liked food")
	}
	// Should seek map food, not eat disliked carried item
	if intent.Action == entity.ActionConsume {
		t.Error("Should not eat disliked carried item at Moderate hunger")
	}
	if intent.TargetItem != mapBerry {
		t.Error("Should target liked map berry instead of disliked carried mushroom")
	}
}

func TestFindFoodIntent_CarriedDislikedItem_EatenAtCrisis(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 95 // Crisis
	char.SetPos(types.Position{X: 5, Y: 5})

	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Carrying a disliked mushroom
	carriedMushroom := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	char.AddToInventory(carriedMushroom)

	// Liked berry far away on map
	mapBerry := entity.NewBerry(15, 15, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierCrisis, nil)

	if intent == nil {
		t.Fatal("Expected intent at Crisis hunger")
	}
	// At Crisis, distance wins - should eat carried item despite dislike
	if intent.Action != entity.ActionConsume {
		t.Error("Should eat carried item at Crisis hunger (distance=0 wins)")
	}
	if intent.TargetItem != carriedMushroom {
		t.Error("Should eat carried mushroom at Crisis (distance advantage)")
	}
}

func TestFindFoodIntent_CarriedLikedItem_WinsOverFarLikedItem(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})

	// Carrying a liked red berry
	carriedBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.AddToInventory(carriedBerry)

	// Another liked red berry far away
	mapBerry := entity.NewBerry(15, 15, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent when carrying food")
	}
	// Carried berry should win due to distance=0
	// Score(carried) = 20*2 - 0 = 40
	// Score(map at 15,15) = 20*2 - 20 = 20
	if intent.Action != entity.ActionConsume {
		t.Error("Should eat carried liked item (distance advantage)")
	}
	if intent.TargetItem != carriedBerry {
		t.Error("Should target carried berry")
	}
}

func TestFindFoodIntent_CarriedNeutralItem_FilteredWhenLikedAvailable(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})

	// Carrying a neutral item (brown mushroom - no preference match)
	carriedMushroom := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	char.AddToInventory(carriedMushroom)

	// Liked red berry nearby
	mapBerry := entity.NewBerry(7, 7, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent")
	}
	// Score(carried neutral) = 20*0 - 0 = 0
	// Score(liked berry at 7,7) = 20*2 - 4 = 36
	// Liked berry should win
	if intent.Action == entity.ActionConsume {
		t.Error("Should prefer liked map berry over neutral carried item")
	}
	if intent.TargetItem != mapBerry {
		t.Error("Should target liked map berry")
	}
}

func TestFindFoodIntent_NoFood_ReturnsNil(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60 // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})

	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Only carrying disliked food, no map food
	carriedMushroom := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	char.AddToInventory(carriedMushroom)

	items := []*entity.Item{} // No map food

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	// At Moderate, disliked carried item should be filtered, no alternatives
	if intent != nil {
		t.Error("Should return nil when only disliked carried food and no map food")
	}
}

func TestFindFoodIntent_VesselContents_RecognizedAsFood(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})

	// Create vessel with red berries (liked)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	variety := &entity.ItemVariety{
		ID:       "berry-red",
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	}
	vessel.Container.Contents = []entity.Stack{{Variety: variety, Count: 5}}
	char.AddToInventory(vessel)

	items := []*entity.Item{} // No map food

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent when carrying vessel with edible contents")
	}
	if intent.Action != entity.ActionConsume {
		t.Errorf("Expected ActionConsume, got %d", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the vessel")
	}
}

func TestFindFoodIntent_VesselWithDislikedContents_FilteredAtModerate(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60 // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})

	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Create vessel with disliked mushrooms
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	variety := &entity.ItemVariety{
		ID:       "mushroom-brown",
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Edible: &entity.EdibleProperties{},
	}
	vessel.Container.Contents = []entity.Stack{{Variety: variety, Count: 5}}
	char.AddToInventory(vessel)

	// Liked berry on map
	mapBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent")
	}
	// Should seek map food, not eat disliked vessel contents
	if intent.Action == entity.ActionConsume {
		t.Error("Should not eat disliked vessel contents at Moderate hunger")
	}
	if intent.TargetItem != mapBerry {
		t.Error("Should target liked map berry")
	}
}

func TestFindFoodIntent_DroppedVessel_RecognizedAsFood(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})
	char.Inventory = nil // Not carrying anything

	// Create dropped vessel with red berries (liked)
	gourd := entity.NewGourd(7, 7, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	vessel.SetPos(types.Position{X: 7, Y: 7})
	variety := &entity.ItemVariety{
		ID:       "berry-red",
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	}
	vessel.Container.Contents = []entity.Stack{{Variety: variety, Count: 5}}

	items := []*entity.Item{vessel}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent when dropped vessel has edible contents")
	}
	if intent.Action != entity.ActionMove {
		t.Errorf("Expected ActionMove to vessel, got %d", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the dropped vessel")
	}
}

func TestFindFoodIntent_DroppedVesselWithDislikedContents_FilteredAtModerate(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 60 // Moderate
	char.SetPos(types.Position{X: 5, Y: 5})
	char.Inventory = nil

	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Create dropped vessel with disliked mushrooms
	gourd := entity.NewGourd(7, 7, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	vessel.SetPos(types.Position{X: 7, Y: 7})
	variety := &entity.ItemVariety{
		ID:       "mushroom-brown",
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Edible: &entity.EdibleProperties{},
	}
	vessel.Container.Contents = []entity.Stack{{Variety: variety, Count: 5}}

	// Liked berry farther away
	mapBerry := entity.NewBerry(15, 15, types.ColorRed, false, false)
	items := []*entity.Item{vessel, mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected intent")
	}
	// Should prefer liked berry over disliked vessel contents
	if intent.TargetItem != mapBerry {
		t.Error("Should target liked map berry, not vessel with disliked contents")
	}
}

func TestFindFoodIntent_DroppedVesselCloser_WinsOverFarFood(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 95           // Crisis - distance wins
	char.SetPos(types.Position{X: 5, Y: 5})
	char.Inventory = nil

	// Create dropped vessel nearby with berries
	gourd := entity.NewGourd(6, 6, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	vessel.SetPos(types.Position{X: 6, Y: 6})
	variety := &entity.ItemVariety{
		ID:       "berry-blue",
		ItemType: "berry",
		Color:    types.ColorBlue,
		Edible: &entity.EdibleProperties{},
	}
	vessel.Container.Contents = []entity.Stack{{Variety: variety, Count: 5}}

	// Red berry far away (more liked but farther)
	mapBerry := entity.NewBerry(20, 20, types.ColorRed, false, false)
	items := []*entity.Item{vessel, mapBerry}

	intent := findFoodIntent(char, types.Position{X: 5, Y: 5}, items, entity.TierCrisis, nil)

	if intent == nil {
		t.Fatal("Expected intent")
	}
	// At Crisis, distance wins - should target closer vessel
	if intent.TargetItem != vessel {
		t.Error("At Crisis hunger, should target closer dropped vessel")
	}
}

// =============================================================================
// Cardinal Adjacency Helpers
// =============================================================================

func TestIsCardinallyAdjacent_Cardinal_True(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		x1, y1   int
		x2, y2   int
		expected bool
	}{
		{"north", 5, 5, 5, 4, true},
		{"south", 5, 5, 5, 6, true},
		{"east", 5, 5, 6, 5, true},
		{"west", 5, 5, 4, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCardinallyAdjacent(tt.x1, tt.y1, tt.x2, tt.y2)
			if got != tt.expected {
				t.Errorf("isCardinallyAdjacent(%d,%d,%d,%d): got %v, want %v",
					tt.x1, tt.y1, tt.x2, tt.y2, got, tt.expected)
			}
		})
	}
}

func TestIsCardinallyAdjacent_Diagonal_False(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		x1, y1 int
		x2, y2 int
	}{
		{"northeast", 5, 5, 6, 4},
		{"northwest", 5, 5, 4, 4},
		{"southeast", 5, 5, 6, 6},
		{"southwest", 5, 5, 4, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCardinallyAdjacent(tt.x1, tt.y1, tt.x2, tt.y2)
			if got {
				t.Errorf("isCardinallyAdjacent(%d,%d,%d,%d): got true for diagonal, want false",
					tt.x1, tt.y1, tt.x2, tt.y2)
			}
		})
	}
}

func TestIsCardinallyAdjacent_SamePosition_False(t *testing.T) {
	t.Parallel()

	got := isCardinallyAdjacent(5, 5, 5, 5)
	if got {
		t.Error("isCardinallyAdjacent(5,5,5,5): got true for same position, want false")
	}
}

func TestIsCardinallyAdjacent_TooFar_False(t *testing.T) {
	t.Parallel()

	got := isCardinallyAdjacent(5, 5, 7, 5)
	if got {
		t.Error("isCardinallyAdjacent(5,5,7,5): got true for distance 2, want false")
	}
}

func TestFindClosestCardinalTile_FindsClosest(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Character at (3, 5), target spring at (5, 5)
	// Cardinal tiles around spring: (5,4), (6,5), (5,6), (4,5)
	// Closest to (3,5) is (4,5) with distance 1
	x, y := findClosestCardinalTile(3, 5, 5, 5, gameMap)

	if x != 4 || y != 5 {
		t.Errorf("findClosestCardinalTile: got (%d,%d), want (4,5)", x, y)
	}
}

func TestFindClosestCardinalTile_SkipsBlocked(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Block the closest tile with a character
	gameMap.AddCharacter(entity.NewCharacter(1, 4, 5, "Blocker", "berry", types.ColorRed))

	// Character at (3, 5), target spring at (5, 5)
	// (4,5) is blocked, next closest should be (5,4) or (5,6) with distance 3
	x, y := findClosestCardinalTile(3, 5, 5, 5, gameMap)

	// Should not return the blocked tile
	if x == 4 && y == 5 {
		t.Error("findClosestCardinalTile should skip blocked tiles")
	}
	// Should return a valid tile
	if x == -1 {
		t.Error("findClosestCardinalTile should find an unblocked tile")
	}
}

func TestFindClosestCardinalTile_AllBlocked(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Block all cardinal tiles around (5, 5)
	gameMap.AddCharacter(entity.NewCharacter(1, 5, 4, "N", "berry", types.ColorRed))
	gameMap.AddCharacter(entity.NewCharacter(2, 6, 5, "E", "berry", types.ColorRed))
	gameMap.AddCharacter(entity.NewCharacter(3, 5, 6, "S", "berry", types.ColorRed))
	gameMap.AddCharacter(entity.NewCharacter(4, 4, 5, "W", "berry", types.ColorRed))

	x, y := findClosestCardinalTile(3, 5, 5, 5, gameMap)

	if x != -1 || y != -1 {
		t.Errorf("findClosestCardinalTile: got (%d,%d), want (-1,-1) when all blocked", x, y)
	}
}

// =============================================================================
// Drink Intent with Cardinal Adjacency
// =============================================================================

func TestFindDrinkIntent_DrinksWhenCardinallyAdjacent(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 75 // Moderate
	char.SetPos(types.Position{X: 5, Y: 4}) // North of spring at (5,5)

	gameMap := game.NewMap(20, 20)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)

	intent := findDrinkIntent(char, types.Position{X: 5, Y: 4}, gameMap, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected drink intent when cardinally adjacent to spring")
	}
	if intent.Action != entity.ActionDrink {
		t.Errorf("Action: got %d, want ActionDrink", intent.Action)
	}
	// Should stay in place (not move onto spring)
	if intent.Target.X != 5 || intent.Target.Y != 4 {
		t.Errorf("Target: got (%d,%d), want (5,4) - should stay in place", intent.Target.X, intent.Target.Y)
	}
}

func TestFindDrinkIntent_DoesNotDrinkWhenDiagonallyAdjacent(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 75
	char.SetPos(types.Position{X: 4, Y: 4}) // Diagonally adjacent to spring at (5,5)

	gameMap := game.NewMap(20, 20)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)

	intent := findDrinkIntent(char, types.Position{X: 4, Y: 4}, gameMap, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected move intent when diagonally adjacent")
	}
	// Should be moving, not drinking
	if intent.Action == entity.ActionDrink {
		t.Error("Should not drink when only diagonally adjacent - must be cardinally adjacent")
	}
	if intent.Action != entity.ActionMove {
		t.Errorf("Action: got %d, want ActionMove", intent.Action)
	}
}

func TestFindDrinkIntent_MovesToAdjacentTile(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 75
	char.SetPos(types.Position{X: 0, Y: 5}) // Far from spring at (5,5)

	gameMap := game.NewMap(20, 20)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)

	intent := findDrinkIntent(char, types.Position{X: 0, Y: 5}, gameMap, entity.TierModerate, nil)

	if intent == nil {
		t.Fatal("Expected move intent")
	}
	if intent.Action != entity.ActionMove {
		t.Errorf("Action: got %d, want ActionMove", intent.Action)
	}
	if intent.TargetFeature != spring {
		t.Error("Should target the spring")
	}
}

func TestContinueIntent_DrinkFromAdjacentTile(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPos(types.Position{X: 6, Y: 5}) // East of spring at (5,5)

	gameMap := game.NewMap(20, 20)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)

	// Set up intent targeting the spring
	char.Intent = &entity.Intent{
		Target:        types.Position{X: 5, Y: 5}, // Was moving toward spring
		Action:        entity.ActionMove,
		TargetFeature: spring,
		DrivingStat:   types.StatThirst,
		DrivingTier:   entity.TierModerate,
	}

	intent := continueIntent(char, 6, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent when cardinally adjacent to spring")
	}
	if intent.Action != entity.ActionDrink {
		t.Errorf("Action: got %d, want ActionDrink when cardinally adjacent", intent.Action)
	}
	// Should stay at current position
	if intent.Target.X != 6 || intent.Target.Y != 5 {
		t.Errorf("Target: got (%d,%d), want (6,5) - should stay in place to drink", intent.Target.X, intent.Target.Y)
	}
}
