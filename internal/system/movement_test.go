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
	char.SetPosition(0, 0)

	// Disliked but close vs liked but far
	// Add a dislike preference for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	nearMushroom := entity.NewMushroom(2, 2, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	farRedBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{farRedBerry, nearMushroom}

	result := findFoodTarget(char, items)

	if result.Item != nearMushroom {
		t.Error("Crisis should pick nearest regardless of preference")
	}
}

// Severe tier (75-89): Uses gradient scoring, considers all items including disliked
func TestFindFoodTarget_SevereUsesGradientScoring(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80           // Severe
	char.SetPosition(0, 0)

	// Liked item far away vs neutral item close
	// Score(redBerry) = 5*2 - 20 = -10
	// Score(brownMushroom) = 5*0 - 6 = -6
	// brownMushroom wins (less negative score)
	redBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	brownMushroom := entity.NewMushroom(3, 3, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{redBerry, brownMushroom}

	result := findFoodTarget(char, items)

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

	result := findFoodTarget(char, items)

	if result.Item == nil {
		t.Error("Severe should consider disliked items when nothing else available")
	}
}

func TestFindFoodTarget_SeverePrefersLikedOverDisliked(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80           // Severe
	char.SetPosition(0, 0)
	// Add dislike for mushrooms
	char.Preferences = append(char.Preferences, entity.NewNegativePreference("mushroom", ""))

	// Liked item same distance as disliked
	// Score(blueBerry at 5,5) = 5*1 - 10 = -5 (likes berries)
	// Score(brownMushroom at 5,5) = 5*(-1) - 10 = -15 (dislikes mushrooms)
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	brownMushroom := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{brownMushroom, blueBerry}

	result := findFoodTarget(char, items)

	if result.Item != blueBerry {
		t.Error("Severe should prefer liked over disliked at same distance")
	}
}

// Moderate tier (50-74): Uses gradient scoring, filters out disliked items (NetPref < 0)
func TestFindFoodTarget_ModerateUsesGradientScoring(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPosition(0, 0)

	// Perfect match far vs partial match close
	// Score(redBerry at 10,10) = 20*2 - 20 = 20
	// Score(blueBerry at 3,3) = 20*1 - 6 = 14
	// redBerry wins despite distance
	redBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	blueBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)

	items := []*entity.Item{blueBerry, redBerry}

	result := findFoodTarget(char, items)

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

	result := findFoodTarget(char, items)

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

	result := findFoodTarget(char, items)

	if result.Item != brownMushroom {
		t.Error("Moderate should accept neutral items (NetPref >= 0)")
	}
}

func TestFindFoodTarget_ModeratePrefersLikedOverNeutral(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPosition(0, 0)

	// Liked item same distance as neutral
	// Score(blueBerry at 5,5) = 20*1 - 10 = 10
	// Score(brownMushroom at 5,5) = 20*0 - 10 = -10
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	brownMushroom := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	items := []*entity.Item{brownMushroom, blueBerry}

	result := findFoodTarget(char, items)

	if result.Item != blueBerry {
		t.Error("Moderate should prefer liked over neutral at same distance")
	}
}

// Distance tiebreaker: when gradient scores are equal, prefer closer item
func TestFindFoodTarget_DistanceTiebreaker(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderate
	char.SetPosition(0, 0)

	// Two items with same NetPreference at different distances
	nearRedBerry := entity.NewBerry(2, 2, types.ColorRed, false, false)
	farRedBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{farRedBerry, nearRedBerry}

	result := findFoodTarget(char, items)

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

	result := findFoodTarget(char, items)

	if result.Item != nil {
		t.Error("Should return nil when no edible items exist")
	}
}

func TestFindFoodTarget_VesselNotEdible(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 95 // Crisis - would eat anything edible

	// Vessel is not edible (crafted from gourd)
	gourd := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)

	items := []*entity.Item{vessel}

	result := findFoodTarget(char, items)

	if result.Item != nil {
		t.Error("Vessel should not be edible - got a food target when none expected")
	}
}

func TestFindFoodIntent_CarriedVesselNotEaten(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 95 // Crisis - would eat anything edible from inventory

	// Character is carrying a vessel (not edible)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	char.Carrying = vessel

	cx, cy := char.Position()
	intent := findFoodIntent(char, cx, cy, nil, entity.TierCrisis, nil)

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
	char.SetPosition(0, 0)
	// No healing knowledge

	// Healing item vs regular item at same distance
	healingBerry := entity.NewBerry(5, 5, types.ColorBlue, false, true) // is healing
	regularBerry := entity.NewBerry(5, 0, types.ColorBlue, false, false)

	items := []*entity.Item{healingBerry, regularBerry}

	result := findFoodTarget(char, items)

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
	char.SetPosition(0, 0)

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

	result := findFoodTarget(char, items)

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
	char.SetPosition(0, 0)

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

	result := findFoodTarget(char, items)

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
	char.SetPosition(0, 0)

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

	result := findFoodTarget(char, items)

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
	char.SetPosition(0, 0)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	char.Intent = &entity.Intent{
		TargetX:     1,
		TargetY:     0,
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
	char.SetPosition(0, 0)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	// Item NOT added to map (simulates it was consumed)

	char.Intent = &entity.Intent{
		TargetX:     1,
		TargetY:     0,
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

func TestContinueIntent_AbandonsIfTargetSpringOccupied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPosition(0, 0)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)

	// Another character at the spring
	otherChar := entity.NewCharacter(2, 5, 5, "Other", "berry", types.ColorBlue)
	gameMap.AddCharacter(otherChar)

	char.Intent = &entity.Intent{
		TargetX:       1,
		TargetY:       0,
		Action:        entity.ActionMove,
		TargetFeature: spring,
		DrivingStat:   types.StatThirst,
		DrivingTier:   2,
	}

	intent := continueIntent(char, 0, 0, gameMap, nil)

	if intent != nil {
		t.Error("Should abandon intent when target spring is occupied")
	}
}

func TestContinueIntent_AbandonsIfTargetBedOccupied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPosition(0, 0)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	bed := entity.NewLeafPile(5, 5)
	gameMap.AddFeature(bed)

	// Another character at the bed
	otherChar := entity.NewCharacter(2, 5, 5, "Other", "berry", types.ColorBlue)
	gameMap.AddCharacter(otherChar)

	char.Intent = &entity.Intent{
		TargetX:       1,
		TargetY:       0,
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
	char.SetPosition(5, 5)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	spring := entity.NewSpring(5, 5)
	gameMap.AddFeature(spring)
	gameMap.AddCharacter(char) // Character is at the spring

	char.Intent = &entity.Intent{
		TargetX:       5,
		TargetY:       5,
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
	char.SetPosition(4, 5) // Adjacent to item at (5,5)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	// ActionPickup intent (foraging) should NOT be converted to ActionLook
	char.Intent = &entity.Intent{
		TargetX:    5,
		TargetY:    5,
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
	char.SetPosition(0, 0) // Not adjacent to item at (5,5)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	char.Intent = &entity.Intent{
		TargetX:    1,
		TargetY:    0,
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
	char.SetPosition(0, 0)

	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorBlue, false, true), // Healing but unknown
	}

	intent := findHealingIntent(char, 0, 0, items, entity.TierModerate, nil)

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
	char.SetPosition(0, 0)

	blueBerryNear := entity.NewBerry(2, 2, types.ColorBlue, false, true)
	blueBerryFar := entity.NewBerry(10, 10, types.ColorBlue, false, true)
	redBerry := entity.NewBerry(1, 1, types.ColorRed, false, false) // Unknown, closer
	items := []*entity.Item{blueBerryFar, redBerry, blueBerryNear}

	intent := findHealingIntent(char, 0, 0, items, entity.TierModerate, nil)

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
	char.SetPosition(0, 0)

	// Only red berries available (not known healing)
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorRed, false, false),
	}

	intent := findHealingIntent(char, 0, 0, items, entity.TierModerate, nil)

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
	char.SetPosition(0, 0)

	intent := findHealingIntent(char, 0, 0, []*entity.Item{}, entity.TierModerate, nil)

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
	char.SetPosition(5, 5)

	// Character is carrying an edible item
	carriedBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = carriedBerry

	// Map item exists but is farther away
	mapItem := entity.NewBerry(10, 10, types.ColorRed, false, false)
	items := []*entity.Item{mapItem}

	intent := findFoodIntent(char, 5, 5, items, entity.TierMild, nil)

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
	char.SetPosition(5, 5)

	// Character is carrying a non-edible item (flower)
	carriedFlower := entity.NewFlower(0, 0, types.ColorRed)
	char.Carrying = carriedFlower

	// Map has edible item
	mapBerry := entity.NewBerry(6, 6, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, 5, 5, items, entity.TierMild, nil)

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
	char.SetPosition(5, 5)
	char.Carrying = nil // Not carrying anything

	mapBerry := entity.NewBerry(6, 6, types.ColorRed, false, false)
	items := []*entity.Item{mapBerry}

	intent := findFoodIntent(char, 5, 5, items, entity.TierMild, nil)

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
	char.SetPosition(5, 5)

	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = carriedItem

	// Set up an ActionConsume intent (eating from inventory)
	char.Intent = &entity.Intent{
		TargetX:     5,
		TargetY:     5,
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

func TestFindForageTarget_SkipsNonGrowingItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPosition(0, 0)

	// Growing berry (should be targeted)
	growingBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	// growingBerry.Plant.IsGrowing is true by default from constructor

	// Non-growing berry (dropped item - should be skipped)
	droppedBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)
	droppedBerry.Plant.IsGrowing = false // Simulates picked up and dropped

	items := []*entity.Item{droppedBerry, growingBerry}

	result := findForageTarget(char, 0, 0, items)

	if result != growingBerry {
		t.Errorf("Expected growing berry, got %v", result)
	}
}

func TestFindForageTarget_ReturnsNilWhenOnlyNonGrowingItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPosition(0, 0)

	// Only non-growing items
	droppedBerry := entity.NewBerry(3, 3, types.ColorRed, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry}

	result := findForageTarget(char, 0, 0, items)

	if result != nil {
		t.Error("Should return nil when only non-growing items exist")
	}
}

func TestFindForageTarget_SkipsItemsWithNilPlant(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.SetPosition(0, 0)

	// Vessel (no Plant property, not forageable)
	gourd := entity.NewGourd(3, 3, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	vessel.Edible = true // Artificially make edible to test Plant filter

	// Growing berry
	growingBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)

	items := []*entity.Item{vessel, growingBerry}

	result := findForageTarget(char, 0, 0, items)

	if result != growingBerry {
		t.Errorf("Expected growing berry (not vessel), got %v", result)
	}
}
