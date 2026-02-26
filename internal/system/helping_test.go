package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// isIdleAction
// =============================================================================

func TestIsIdleAction_NilIntent_ReturnsTrue(t *testing.T) {
	t.Parallel()
	char := newTestCharacter()
	char.Intent = nil
	if !isIdleAction(char) {
		t.Error("Character with nil intent should be idle")
	}
}

func TestIsIdleAction_IdleActions_ReturnTrue(t *testing.T) {
	t.Parallel()
	idleActions := []entity.ActionType{
		entity.ActionNone,
		entity.ActionLook,
		entity.ActionTalk,
		entity.ActionForage,
		entity.ActionFillVessel,
	}
	for _, action := range idleActions {
		char := newTestCharacter()
		char.Intent = &entity.Intent{Action: action}
		if !isIdleAction(char) {
			t.Errorf("ActionType %d should be idle", action)
		}
	}
}

func TestIsIdleAction_NonIdleActions_ReturnFalse(t *testing.T) {
	t.Parallel()
	nonIdleActions := []entity.ActionType{
		entity.ActionHelpFeed,
		entity.ActionWaterGarden,
		entity.ActionTillSoil,
		entity.ActionPlant,
		entity.ActionPickup,
		entity.ActionCraft,
		entity.ActionConsume,
		entity.ActionDrink,
		entity.ActionSleep,
	}
	for _, action := range nonIdleActions {
		char := newTestCharacter()
		char.Intent = &entity.Intent{Action: action}
		if isIdleAction(char) {
			t.Errorf("ActionType %d should not be idle", action)
		}
	}
}

// =============================================================================
// findNearestCrisisCharacter
// =============================================================================

func TestFindNearestCrisisCharacter_ReturnsCrisisHungryCharacter(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 8, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100 // Crisis

	chars := []*entity.Character{helper, needer}
	result := findNearestCrisisCharacter(helper, chars)

	if result != needer {
		t.Error("Should return the character at Crisis hunger")
	}
}

func TestFindNearestCrisisCharacter_ReturnsNilWhenNoCrisis(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	other := entity.NewCharacter(2, 8, 5, "Other", "berry", types.ColorBlue)
	other.Hunger = 75 // Severe, not Crisis

	chars := []*entity.Character{helper, other}
	result := findNearestCrisisCharacter(helper, chars)

	if result != nil {
		t.Error("Should return nil when no character is in crisis")
	}
}

func TestFindNearestCrisisCharacter_SkipsSelf(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 100 // Helper is in crisis themselves

	chars := []*entity.Character{helper}
	result := findNearestCrisisCharacter(helper, chars)

	if result != nil {
		t.Error("Should skip the helper themselves")
	}
}

func TestFindNearestCrisisCharacter_SkipsDeadAndSleeping(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)

	dead := entity.NewCharacter(2, 6, 5, "Dead", "berry", types.ColorBlue)
	dead.Hunger = 100
	dead.IsDead = true

	sleeping := entity.NewCharacter(3, 7, 5, "Sleeping", "berry", types.ColorGreen)
	sleeping.Hunger = 100
	sleeping.IsSleeping = true

	chars := []*entity.Character{helper, dead, sleeping}
	result := findNearestCrisisCharacter(helper, chars)

	if result != nil {
		t.Error("Should skip dead and sleeping characters")
	}
}

func TestFindNearestCrisisCharacter_ReturnsNearest(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)

	far := entity.NewCharacter(2, 20, 5, "Far", "berry", types.ColorBlue)
	far.Hunger = 100

	near := entity.NewCharacter(3, 7, 5, "Near", "berry", types.ColorGreen)
	near.Hunger = 100

	chars := []*entity.Character{helper, far, near}
	result := findNearestCrisisCharacter(helper, chars)

	if result != near {
		t.Errorf("Should return nearest crisis character (Near at dist 2), got %v", result)
	}
}

// =============================================================================
// findHelpFeedIntent
// =============================================================================

func TestFindHelpFeedIntent_GroundFood_ReturnsHelpFeedIntent(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0 // Well fed
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100 // Crisis

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	mushroom := entity.NewMushroom(7, 5, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)
	gameMap.AddItem(mushroom)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected ActionHelpFeed intent, got nil")
	}
	if intent.Action != entity.ActionHelpFeed {
		t.Errorf("Action: got %d, want ActionHelpFeed", intent.Action)
	}
	if intent.TargetItem != mushroom {
		t.Error("TargetItem should be the ground mushroom")
	}
	if intent.TargetCharacter != needer {
		t.Error("TargetCharacter should be the needer")
	}
}

func TestFindHelpFeedIntent_CarriedFood_SkipsProcurement(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	helper.AddToInventory(berry)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != berry {
		t.Error("TargetItem should be the carried berry")
	}
	// Dest should be needer's position (delivery phase, no procurement needed)
	npos := needer.Pos()
	if intent.Dest.X != npos.X || intent.Dest.Y != npos.Y {
		t.Errorf("Dest should be needer's position (%d,%d), got (%d,%d)", npos.X, npos.Y, intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindHelpFeedIntent_CarriedFoodVessel_DeliversVessel(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	vessel := createTestVessel()
	vessel.Container.Contents = []entity.Stack{{
		Variety: &entity.ItemVariety{
			ItemType: "berry",
			Color:    types.ColorRed,
			Edible:   &entity.EdibleProperties{},
		},
		Count: 3,
	}}
	helper.AddToInventory(vessel)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the carried food vessel")
	}
}

func TestFindHelpFeedIntent_FullInventoryNoFood_ReturnsNil(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	// Fill inventory with non-food items
	stick1 := entity.NewStick(0, 0)
	stick2 := entity.NewStick(0, 0)
	helper.AddToInventory(stick1)
	helper.AddToInventory(stick2)

	// Food exists on ground but helper can't carry it
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent != nil {
		t.Error("Helper with full inventory and no food should not be able to help")
	}
}

func TestFindHelpFeedIntent_NoFoodAvailable_ReturnsNil(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// No food on map, nothing in inventory

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent != nil {
		t.Error("Should return nil when no food is available")
	}
}

func TestFindHelpFeedIntent_UsesSevereTierWeights(t *testing.T) {
	t.Parallel()
	// Mushroom is meal-tier (higher satiation), closer. Berry is snack-tier, farther.
	// With Severe weights and needer's hunger at 100, mushroom should win:
	// larger satiation fit for hunger 100, AND closer distance.
	helper := entity.NewCharacter(1, 10, 10, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	needer := entity.NewCharacter(2, 20, 10, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	mushroom := entity.NewMushroom(12, 10, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)
	gameMap.AddItem(mushroom)

	berry := entity.NewBerry(18, 10, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != mushroom {
		t.Error("Mushroom (meal, closer) should score better than berry (snack, farther) with Severe weights and hunger 100")
	}
}

func TestFindHelpFeedIntent_PoisonKnowledge_PenalizesKnownPoison(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 10, 10, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	// Helper knows the nearby mushroom is poisonous (auto-formed dislike)
	helper.Preferences = append(helper.Preferences, entity.NewNegativePreference("mushroom", types.ColorBrown))
	needer := entity.NewCharacter(2, 20, 10, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	// Poisonous mushroom very close to helper
	mushroom := entity.NewMushroom(11, 10, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, true, false)
	gameMap.AddItem(mushroom)

	// Berry farther away but not disliked
	berry := entity.NewBerry(14, 10, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != berry {
		t.Error("Helper's dislike of poisonous mushroom should penalize it; berry should be chosen instead")
	}
}

func TestFindHelpFeedIntent_GroundFoodVessel_ScoredAsCandidate(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	// Ground vessel with berries — closer than the loose mushroom
	vessel := createTestVessel()
	vessel.X = 6
	vessel.Y = 5
	vessel.Container.Contents = []entity.Stack{{
		Variety: &entity.ItemVariety{
			ItemType: "berry",
			Color:    types.ColorRed,
			Edible:   &entity.EdibleProperties{},
		},
		Count: 3,
	}}
	gameMap.AddItem(vessel)

	// Loose mushroom much farther
	mushroom := entity.NewMushroom(20, 5, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)
	gameMap.AddItem(mushroom)

	intent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != vessel {
		t.Error("Ground food vessel (closer) should be scored as a candidate")
	}
}

// =============================================================================
// Talking availability: helpers are NOT targetable for talking
// =============================================================================

func TestFindTalkIntent_SkipsHelpingCharacters(t *testing.T) {
	t.Parallel()
	talker := entity.NewCharacter(1, 5, 5, "Talker", "berry", types.ColorRed)
	talker.Hunger = 0
	talker.Thirst = 0
	talker.Energy = 100

	helper := entity.NewCharacter(2, 6, 5, "Helper", "berry", types.ColorBlue)
	helper.Hunger = 0
	helper.Thirst = 0
	helper.Energy = 100
	helper.Intent = &entity.Intent{Action: entity.ActionHelpFeed}
	helper.CurrentActivity = "Bringing food to someone"

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddCharacter(talker)
	gameMap.AddCharacter(helper)

	intent := findTalkIntent(talker, talker.Pos(), gameMap, nil)

	if intent != nil {
		t.Error("Helper with ActionHelpFeed should not be targetable for talking")
	}
}
