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
// Food Target Selection
// =============================================================================

func TestFindFoodTarget_RavenousTakesAnyFood(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red (via NewCharacter)
	char.Hunger = 90           // Ravenous

	// Only non-matching food available
	items := []*entity.Item{entity.NewMushroom(5, 5, types.ColorBrown, false, false)}

	target := findFoodTarget(char, items)

	if target == nil {
		t.Error("Ravenous character should take any food")
	}
}

func TestFindFoodTarget_VeryHungryTakesPartialMatch(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80           // Very hungry (75-89)

	// Partial match (food matches, color doesn't)
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	// No match
	brownMushroom := entity.NewMushroom(3, 3, types.ColorBrown, false, false)

	items := []*entity.Item{brownMushroom, blueBerry}

	target := findFoodTarget(char, items)

	if target != blueBerry {
		t.Errorf("Very hungry should prefer partial match, got %v", target)
	}
}

func TestFindFoodTarget_VeryHungryTakesAnyIfNoPartial(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 80

	// Only non-matching food
	items := []*entity.Item{entity.NewMushroom(5, 5, types.ColorBrown, false, false)}

	target := findFoodTarget(char, items)

	if target == nil {
		t.Error("Very hungry should take any food if no partial match")
	}
}

func TestFindFoodTarget_ModeratelyHungryPrefersPerfectMatch(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60           // Moderately hungry (50-74)

	// Perfect match
	redBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	// Partial match (closer)
	blueBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)

	items := []*entity.Item{blueBerry, redBerry}

	target := findFoodTarget(char, items)

	if target != redBerry {
		t.Errorf("Moderately hungry should prefer perfect match, got %v", target)
	}
}

func TestFindFoodTarget_ModeratelyHungryTakesPartialIfNoPerfect(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60

	// Only partial match
	blueBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false)

	items := []*entity.Item{blueBerry}

	target := findFoodTarget(char, items)

	if target != blueBerry {
		t.Error("Moderately hungry should take partial match if no perfect")
	}
}

func TestFindFoodTarget_ModeratelyHungryReturnsNilForNonMatching(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60

	// Only non-matching food
	items := []*entity.Item{entity.NewMushroom(5, 5, types.ColorBrown, false, false)}

	target := findFoodTarget(char, items)

	if target != nil {
		t.Error("Moderately hungry should return nil for non-matching food")
	}
}

func TestFindFoodTarget_SelectsNearestWithinPreferenceTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berries and red
	char.Hunger = 60
	char.SetPosition(0, 0)

	// Two perfect matches at different distances
	nearRedBerry := entity.NewBerry(2, 2, types.ColorRed, false, false)
	farRedBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{farRedBerry, nearRedBerry}

	target := findFoodTarget(char, items)

	if target != nearRedBerry {
		t.Error("Should select nearest food within preference tier")
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
