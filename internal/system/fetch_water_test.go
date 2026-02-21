package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

func TestFindFetchWaterIntent_PicksUpEmptyGroundVessel(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character at same position as empty vessel
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 7
	char.Y = 5

	// Empty vessel on the ground at character's position
	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	vessel.ID = 100
	gameMap.AddItemDirect(vessel)

	// Water on the map
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent == nil {
		t.Fatal("Expected intent to pick up empty ground vessel")
	}
	// Unified flow: always ActionFillVessel, vessel pickup handled within applyIntent
	if intent.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("Expected intent to target the empty vessel")
	}
	// Dest should be the vessel position (phase 1 target)
	if intent.Dest.X != 7 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest at vessel position (7,5), got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindFetchWaterIntent_MovesTowardGroundVessel(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character far from vessel
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 5
	char.Y = 5

	// Empty vessel on the ground, 2 tiles away
	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	vessel.ID = 100
	gameMap.AddItemDirect(vessel)

	// Water on the map
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent == nil {
		t.Fatal("Expected intent to move toward ground vessel")
	}
	// Unified flow: always ActionFillVessel, handles own movement
	if intent.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("Expected intent to target the empty vessel")
	}
	// Dest should be the vessel position (phase 1)
	if intent.Dest.X != 7 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest at vessel position (7,5), got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
	// Target should be first BFS step toward vessel
	if intent.Target.X <= 5 {
		t.Errorf("Expected Target.X > 5 (step toward vessel), got %d", intent.Target.X)
	}
}

func TestFindFetchWaterIntent_MovesTowardWater(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries an empty vessel
	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 5
	char.Y = 5

	// Water far away
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent == nil {
		t.Fatal("Expected intent toward water")
	}
	// ActionFillVessel handles its own movement (like ActionTillSoil)
	if intent.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", intent.Action)
	}
	// Dest should be near water, not at character position
	if intent.Dest.X == char.X && intent.Dest.Y == char.Y {
		t.Error("Expected Dest to be near water, not at character position")
	}
	if intent.TargetItem != vessel {
		t.Error("Expected intent to target the carried vessel")
	}
}

func TestFindFetchWaterIntent_FillsVesselWhenAdjacentToWater(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries an empty vessel, adjacent to water
	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 10
	char.Y = 5

	// Water cardinally adjacent
	gameMap.AddWater(types.Position{X: 10, Y: 6}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent == nil {
		t.Fatal("Expected ActionFillVessel intent")
	}
	if intent.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("Expected intent to target the carried vessel")
	}
}

func TestFindFetchWaterIntent_NilWhenCarryingWater(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries a vessel with water contents
	vessel := createTestVessel()
	waterVariety := createWaterVariety()
	AddLiquidToVessel(vessel, waterVariety, 4)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 5
	char.Y = 5

	// Water on the map
	gameMap.AddWater(types.Position{X: 10, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent != nil {
		t.Error("Expected nil when character already carries water")
	}
}

func TestFindFetchWaterIntent_NilWhenCarryingWater_EvenWithGroundVessel(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries a vessel with water — has inventory space
	vessel := createTestVessel()
	waterVariety := createWaterVariety()
	AddLiquidToVessel(vessel, waterVariety, 4)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 5
	char.Y = 5

	// Empty ground vessel nearby
	groundVessel := createTestVessel()
	groundVessel.X = 7
	groundVessel.Y = 5
	groundVessel.ID = 100
	gameMap.AddItemDirect(groundVessel)

	// Water on the map
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent != nil {
		t.Error("Expected nil — character already has water, shouldn't pick up another vessel")
	}
}

func TestFindFetchWaterIntent_BerryVesselSeeksGroundVessel(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries a vessel with berry contents (not water)
	berryVessel := createTestVessel()
	berryVariety := &entity.ItemVariety{
		ID:       "berry-red",
		ItemType: "berry",
		Color:    types.ColorRed,
	}
	berryVessel.Container.Contents = []entity.Stack{
		{Variety: berryVariety, Count: 2},
	}

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{berryVessel},
	}
	char.X = 5
	char.Y = 5

	// Empty ground vessel nearby
	groundVessel := createTestVessel()
	groundVessel.X = 5
	groundVessel.Y = 5
	groundVessel.ID = 100
	gameMap.AddItemDirect(groundVessel)

	// Water on the map
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent == nil {
		t.Fatal("Expected intent — vessel has berries not water, should pick up ground vessel")
	}
	// Unified flow: ActionFillVessel manages vessel pickup internally
	if intent.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", intent.Action)
	}
	if intent.TargetItem != groundVessel {
		t.Error("Expected intent to target the ground vessel")
	}
}

func TestFindFetchWaterIntent_WaterVesselPlusEmptyVessel_Nil(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries water vessel AND an empty vessel
	waterVessel := createTestVessel()
	waterVariety := createWaterVariety()
	AddLiquidToVessel(waterVessel, waterVariety, 4)

	emptyVessel := createTestVessel()

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{waterVessel, emptyVessel},
	}
	char.X = 5
	char.Y = 5

	// Water on the map
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent != nil {
		t.Error("Expected nil — character already has water, even though they have an empty vessel")
	}
}

func TestFindFetchWaterIntent_NilWhenNoWaterOnMap(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character carries an empty vessel
	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 5
	char.Y = 5

	// No water on map

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent != nil {
		t.Error("Expected nil when no water on map")
	}
}

func TestContinueIntent_FillVessel_RecalculatesTarget(t *testing.T) {
	gameMap := game.NewMap(20, 20)

	// Water at (15, 5)
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	// Character at (5, 5) with ActionFillVessel intent toward (14, 5) (adjacent to water)
	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 5
	char.Y = 5

	char.Intent = &entity.Intent{
		Target:     types.Position{X: 6, Y: 5}, // Stale first-step
		Dest:       types.Position{X: 14, Y: 5},
		Action:     entity.ActionFillVessel,
		TargetItem: vessel,
	}

	// Simulate: character moved to (10, 5) but Target wasn't updated
	char.X = 10
	char.Y = 5

	result := continueIntent(char, 10, 5, gameMap, nil)
	if result == nil {
		t.Fatal("Expected continueIntent to return intent for ActionFillVessel")
	}
	// Target should be recalculated from (10,5) toward (14,5), not stale (6,5)
	if result.Target.X <= 10 {
		t.Errorf("Expected Target.X > 10 (next step toward dest), got %d", result.Target.X)
	}
	if result.Target.X != 11 {
		t.Errorf("Expected Target.X = 11 (one step east), got %d", result.Target.X)
	}
}

func TestContinueIntent_FillVessel_AtDestination(t *testing.T) {
	gameMap := game.NewMap(20, 20)

	// Water at (15, 5)
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	// Character already at destination
	char.X = 14
	char.Y = 5

	char.Intent = &entity.Intent{
		Target:     types.Position{X: 14, Y: 5},
		Dest:       types.Position{X: 14, Y: 5},
		Action:     entity.ActionFillVessel,
		TargetItem: vessel,
	}

	result := continueIntent(char, 14, 5, gameMap, nil)
	if result == nil {
		t.Fatal("Expected continueIntent to return intent at destination")
	}
	if result.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", result.Action)
	}
}

func TestContinueIntent_FillVessel_Phase1_RecalculatesTarget(t *testing.T) {
	gameMap := game.NewMap(20, 20)

	// Vessel on the ground at (12, 5)
	vessel := createTestVessel()
	vessel.X = 12
	vessel.Y = 5
	vessel.ID = 100
	gameMap.AddItemDirect(vessel)

	// Character at (5, 5) — vessel is on map (phase 1)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 5
	char.Y = 5

	char.Intent = &entity.Intent{
		Target:     types.Position{X: 6, Y: 5}, // Stale first-step
		Dest:       types.Position{X: 12, Y: 5},
		Action:     entity.ActionFillVessel,
		TargetItem: vessel,
	}

	// Simulate: character moved to (10, 5)
	char.X = 10
	char.Y = 5

	result := continueIntent(char, 10, 5, gameMap, nil)
	if result == nil {
		t.Fatal("Expected continueIntent to return intent for phase 1")
	}
	// Target should be recalculated toward the vessel at (12, 5)
	if result.Target.X != 11 {
		t.Errorf("Expected Target.X = 11 (next step toward vessel), got %d", result.Target.X)
	}
}

func TestContinueIntent_FillVessel_Phase1_AtVessel(t *testing.T) {
	gameMap := game.NewMap(20, 20)

	// Vessel on the ground at (12, 5)
	vessel := createTestVessel()
	vessel.X = 12
	vessel.Y = 5
	vessel.ID = 100
	gameMap.AddItemDirect(vessel)

	// Character at vessel position (phase 1, ready for pickup)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 12
	char.Y = 5

	char.Intent = &entity.Intent{
		Target:     types.Position{X: 12, Y: 5},
		Dest:       types.Position{X: 12, Y: 5},
		Action:     entity.ActionFillVessel,
		TargetItem: vessel,
	}

	result := continueIntent(char, 12, 5, gameMap, nil)
	if result == nil {
		t.Fatal("Expected continueIntent to return intent at vessel position")
	}
	if result.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel, got %d", result.Action)
	}
}

func TestContinueIntent_FillVessel_Phase1_VesselTaken(t *testing.T) {
	gameMap := game.NewMap(20, 20)

	// Vessel was on ground but has been picked up by someone else (not on map)
	vessel := createTestVessel()
	vessel.X = -1
	vessel.Y = -1
	vessel.ID = 100
	// Don't add to map — simulates it being picked up

	// Character at (10, 5) heading toward where vessel was
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 10
	char.Y = 5

	char.Intent = &entity.Intent{
		Target:     types.Position{X: 11, Y: 5},
		Dest:       types.Position{X: 12, Y: 5},
		Action:     entity.ActionFillVessel,
		TargetItem: vessel,
	}

	result := continueIntent(char, 10, 5, gameMap, nil)
	if result != nil {
		t.Error("Expected nil — vessel was taken by another character")
	}
}

func TestFindFetchWaterIntent_SkipsVesselsWithContents(t *testing.T) {
	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	// Character has no vessel
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 5
	char.Y = 5

	// Ground vessel with water contents (not empty)
	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	vessel.ID = 100
	waterVariety := createWaterVariety()
	AddLiquidToVessel(vessel, waterVariety, 4)
	gameMap.AddItemDirect(vessel)

	// Water on the map
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	intent := findFetchWaterIntent(char, char.Pos(), gameMap.Items(), gameMap, nil)
	if intent != nil {
		t.Error("Expected nil — only ground vessel has contents, should be skipped")
	}
}
