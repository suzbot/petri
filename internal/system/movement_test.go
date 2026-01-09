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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	intent := CalculateIntent(char, nil, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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

	CalculateIntent(char, nil, gameMap, nil)

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

	CalculateIntent(char, nil, gameMap, nil)

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

	CalculateIntent(char, nil, gameMap, nil)

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

	intent := CalculateIntent(char, items, gameMap, nil)

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
