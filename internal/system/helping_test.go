package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// isDiscretionaryAction
// =============================================================================

func TestIsDiscretionaryAction_NilIntent_ReturnsTrue(t *testing.T) {
	t.Parallel()
	char := newTestCharacter()
	char.Intent = nil
	if !isDiscretionaryAction(char) {
		t.Error("Character with nil intent should be idle")
	}
}

func TestIsDiscretionaryAction_IdleActions_ReturnTrue(t *testing.T) {
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
		if !isDiscretionaryAction(char) {
			t.Errorf("ActionType %d should be idle", action)
		}
	}
}

func TestIsDiscretionaryAction_NonIdleActions_ReturnFalse(t *testing.T) {
	t.Parallel()
	nonIdleActions := []entity.ActionType{
		entity.ActionHelpFeed,
		entity.ActionHelpWater,
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
		if isDiscretionaryAction(char) {
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
// findNearestCrisisCharacter — thirst extension
// =============================================================================

func TestFindNearestCrisisCharacter_ReturnsCrisisThirstyCharacter(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 8, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100 // Crisis thirst

	chars := []*entity.Character{helper, needer}
	result := findNearestCrisisCharacter(helper, chars)

	if result != needer {
		t.Error("Should return the character at Crisis thirst")
	}
}

func TestFindNearestCrisisCharacter_DistancePrimary_NearerHungryOverFartherThirsty(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)

	nearHungry := entity.NewCharacter(2, 7, 5, "NearHungry", "berry", types.ColorBlue)
	nearHungry.Hunger = 100 // Crisis hunger, distance 2

	farThirsty := entity.NewCharacter(3, 20, 5, "FarThirsty", "berry", types.ColorGreen)
	farThirsty.Thirst = 100 // Crisis thirst, distance 15

	chars := []*entity.Character{helper, nearHungry, farThirsty}
	result := findNearestCrisisCharacter(helper, chars)

	if result != nearHungry {
		t.Error("Distance should be primary — nearer hungry character should be chosen over farther thirsty one")
	}
}

func TestFindNearestCrisisCharacter_Equidistant_ThirstWinsOverHunger(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 10, 10, "Helper", "berry", types.ColorRed)

	hungry := entity.NewCharacter(2, 13, 10, "Hungry", "berry", types.ColorBlue)
	hungry.Hunger = 100 // Crisis hunger, distance 3

	thirsty := entity.NewCharacter(3, 10, 13, "Thirsty", "berry", types.ColorGreen)
	thirsty.Thirst = 100 // Crisis thirst, distance 3

	chars := []*entity.Character{helper, hungry, thirsty}
	result := findNearestCrisisCharacter(helper, chars)

	if result != thirsty {
		t.Error("At equal distance, thirst crisis should win over hunger crisis (stat tie-breaker)")
	}
}

func TestFindNearestCrisisCharacter_BothCrisesOnSameCharacter(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 8, 5, "Needer", "berry", types.ColorBlue)
	needer.Hunger = 100 // Crisis hunger
	needer.Thirst = 100 // Crisis thirst

	chars := []*entity.Character{helper, needer}
	result := findNearestCrisisCharacter(helper, chars)

	if result != needer {
		t.Error("Should return character with both crisis stats")
	}
}

// =============================================================================
// findHelpWaterIntent
// =============================================================================

func TestFindHelpWaterIntent_CarryingWaterVessel_DeliveryIntent(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	// Helper carrying a water vessel
	vessel := createTestVessel()
	vessel.Container.Contents = []entity.Stack{{
		Variety: &entity.ItemVariety{
			ItemType: "liquid",
			Kind:     "water",
		},
		Count: 5,
	}}
	helper.AddToInventory(vessel)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected ActionHelpWater intent, got nil")
	}
	if intent.Action != entity.ActionHelpWater {
		t.Errorf("Action: got %d, want ActionHelpWater", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the carried water vessel")
	}
	if intent.TargetCharacter != needer {
		t.Error("TargetCharacter should be the needer")
	}
	// Dest should be needer's position (delivery phase — skip procurement and fill)
	npos := needer.Pos()
	if intent.Dest.X != npos.X || intent.Dest.Y != npos.Y {
		t.Errorf("Dest should be needer's position (%d,%d), got (%d,%d)", npos.X, npos.Y, intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindHelpWaterIntent_CarryingEmptyVessel_FillPhaseIntent(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	vessel := createTestVessel()
	helper.AddToInventory(vessel)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Add water source for fill destination
	gameMap.AddWater(types.Position{X: 10, Y: 5}, game.WaterSpring)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected ActionHelpWater intent, got nil")
	}
	if intent.Action != entity.ActionHelpWater {
		t.Errorf("Action: got %d, want ActionHelpWater", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the carried empty vessel")
	}
	if intent.TargetCharacter != needer {
		t.Error("TargetCharacter should be the needer")
	}
	// Dest should be water-adjacent tile (fill phase), NOT needer's position
	npos := needer.Pos()
	if intent.Dest.X == npos.X && intent.Dest.Y == npos.Y {
		t.Error("Dest should be water-adjacent tile for fill phase, not needer's position")
	}
}

func TestFindHelpWaterIntent_GroundVessel_ProcurementIntent(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	// Empty vessel on the ground
	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddItem(vessel)
	gameMap.AddWater(types.Position{X: 10, Y: 5}, game.WaterSpring)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected ActionHelpWater intent, got nil")
	}
	if intent.Action != entity.ActionHelpWater {
		t.Errorf("Action: got %d, want ActionHelpWater", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the ground vessel")
	}
	if intent.TargetCharacter != needer {
		t.Error("TargetCharacter should be the needer")
	}
	// Dest should be vessel's position (procurement phase)
	vpos := vessel.Pos()
	if intent.Dest.X != vpos.X || intent.Dest.Y != vpos.Y {
		t.Errorf("Dest should be vessel's position (%d,%d), got (%d,%d)", vpos.X, vpos.Y, intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindHelpWaterIntent_GroundWaterVessel_ProcurementDeliveryIntent(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	// Water vessel on the ground (already filled)
	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	vessel.Container.Contents = []entity.Stack{{
		Variety: &entity.ItemVariety{
			ItemType: "liquid",
			Kind:     "water",
		},
		Count: 5,
	}}
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddItem(vessel)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected ActionHelpWater intent for ground water vessel, got nil")
	}
	if intent.Action != entity.ActionHelpWater {
		t.Errorf("Action: got %d, want ActionHelpWater", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("TargetItem should be the ground water vessel")
	}
	if intent.TargetCharacter != needer {
		t.Error("TargetCharacter should be the needer")
	}
	// Dest should be vessel's position (procurement phase — pick up, then deliver)
	vpos := vessel.Pos()
	if intent.Dest.X != vpos.X || intent.Dest.Y != vpos.Y {
		t.Errorf("Dest should be vessel's position (%d,%d), got (%d,%d)", vpos.X, vpos.Y, intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindHelpWaterIntent_GroundWaterVessel_PreferredOverGroundEmpty(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	// Water vessel on the ground (farther but already filled)
	waterVessel := createTestVessel()
	waterVessel.X = 9
	waterVessel.Y = 5
	waterVessel.Container.Contents = []entity.Stack{{
		Variety: &entity.ItemVariety{
			ItemType: "liquid",
			Kind:     "water",
		},
		Count: 5,
	}}

	// Empty vessel on the ground (closer but needs filling)
	emptyVessel := createTestVessel()
	emptyVessel.X = 7
	emptyVessel.Y = 5

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddItem(waterVessel)
	gameMap.AddItem(emptyVessel)
	gameMap.AddWater(types.Position{X: 20, Y: 5}, game.WaterSpring)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != waterVessel {
		t.Error("Ground water vessel should be preferred over ground empty vessel (skips fill phase)")
	}
}

func TestFindHelpWaterIntent_NoVessel_ReturnsNil(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddWater(types.Position{X: 10, Y: 5}, game.WaterSpring)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent != nil {
		t.Error("Should return nil when no vessel is available")
	}
}

func TestFindHelpWaterIntent_NoInventorySpace_NoCarriedVessel_ReturnsNil(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	// Fill inventory with non-vessel items
	stick1 := entity.NewStick(0, 0)
	stick2 := entity.NewStick(0, 0)
	helper.AddToInventory(stick1)
	helper.AddToInventory(stick2)

	// Ground vessel exists but helper can't pick it up
	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddItem(vessel)
	gameMap.AddWater(types.Position{X: 10, Y: 5}, game.WaterSpring)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent != nil {
		t.Error("Helper with full inventory and no carried vessel should not be able to help")
	}
}

func TestFindHelpWaterIntent_CarryingEmptyVessel_NoWater_ReturnsNil(t *testing.T) {
	t.Parallel()
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 15, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100

	vessel := createTestVessel()
	helper.AddToInventory(vessel)

	// No water on map
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)

	intent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)

	if intent != nil {
		t.Error("Should return nil when carrying empty vessel but no water source exists")
	}
}

// =============================================================================
// Idle override: thirst crisis routing and fallback
// =============================================================================

func TestIdleOverride_ThirstCrisis_NoVessel_HungerFallback(t *testing.T) {
	t.Parallel()
	// Needer has both Crisis thirst and Crisis hunger. No vessel available.
	// Helper should fall back to helpFeed.
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	helper.Hunger = 0
	helper.Thirst = 0
	helper.Energy = 100

	needer := entity.NewCharacter(2, 8, 5, "Needer", "berry", types.ColorBlue)
	needer.Thirst = 100 // Crisis thirst
	needer.Hunger = 100 // Crisis hunger

	// Food available but no vessel
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)
	gameMap.AddCharacter(helper)
	gameMap.AddCharacter(needer)

	// findNearestCrisisCharacter should find needer (thirst crisis)
	found := findNearestCrisisCharacter(helper, gameMap.Characters())
	if found != needer {
		t.Fatal("Should find the needer with crisis thirst")
	}

	// helpWater should fail (no vessel)
	waterIntent := findHelpWaterIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)
	if waterIntent != nil {
		t.Fatal("helpWater should return nil when no vessel available")
	}

	// Fallback: needer also has Crisis hunger, so helpFeed should work
	feedIntent := findHelpFeedIntent(helper, needer, helper.Pos(), gameMap.Items(), gameMap, nil)
	if feedIntent == nil {
		t.Fatal("helpFeed should succeed as fallback when needer also has Crisis hunger")
	}
	if feedIntent.Action != entity.ActionHelpFeed {
		t.Error("Fallback intent should be ActionHelpFeed")
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

func TestFindTalkIntent_SkipsHelpWaterCharacters(t *testing.T) {
	t.Parallel()
	talker := entity.NewCharacter(1, 5, 5, "Talker", "berry", types.ColorRed)
	talker.Hunger = 0
	talker.Thirst = 0
	talker.Energy = 100

	helper := entity.NewCharacter(2, 6, 5, "Helper", "berry", types.ColorBlue)
	helper.Hunger = 0
	helper.Thirst = 0
	helper.Energy = 100
	helper.Intent = &entity.Intent{Action: entity.ActionHelpWater}
	helper.CurrentActivity = "Fetching water for someone"

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddCharacter(talker)
	gameMap.AddCharacter(helper)

	intent := findTalkIntent(talker, talker.Pos(), gameMap, nil)

	if intent != nil {
		t.Error("Helper with ActionHelpWater should not be targetable for talking")
	}
}
