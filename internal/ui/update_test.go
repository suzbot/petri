package ui

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/system"
	"petri/internal/types"
)

// =============================================================================
// applyIntent Tests
// =============================================================================

func TestApplyIntent_CollapseIsImmediate(t *testing.T) {
	t.Parallel()

	// Setup: character at (5,5) with energy=0, intent to move toward bed at (10,10)
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berries", types.ColorRed)
	char.Energy = 0 // Collapsed state
	bed := entity.NewLeafPile(10, 10)
	char.Intent = &entity.Intent{
		Action:        entity.ActionMove,
		TargetX:       6,
		TargetY:       5,
		TargetFeature: bed,
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent - should trigger collapse instead of movement
	m.applyIntent(char, 0.1)

	// Assert: character should be sleeping (collapsed on ground)
	if !char.IsSleeping {
		t.Error("Expected character to be sleeping after collapse, but IsSleeping is false")
	}

	// Assert: character should NOT have moved (still at 5,5)
	cx, cy := char.Position()
	if cx != 5 || cy != 5 {
		t.Errorf("Expected character to stay at (5,5) after collapse, but moved to (%d,%d)", cx, cy)
	}
}

func TestApplyIntent_NoCollapseWhenEnergyAboveZero(t *testing.T) {
	t.Parallel()

	// Setup: character at (5,5) with energy=1 (just above collapse), intent to move
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berries", types.ColorRed)
	char.Energy = 1 // Just above collapse threshold
	char.Intent = &entity.Intent{
		Action:  entity.ActionMove,
		TargetX: 6,
		TargetY: 5,
	}
	// Need enough speed to actually move (threshold is 7.5)
	char.SpeedAccumulator = 10.0
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent - should move normally
	m.applyIntent(char, 0.1)

	// Assert: character should NOT be sleeping
	if char.IsSleeping {
		t.Error("Expected character with energy=1 to not collapse, but IsSleeping is true")
	}

	// Assert: character should have moved to (6,5)
	cx, cy := char.Position()
	if cx != 6 || cy != 5 {
		t.Errorf("Expected character to move to (6,5), but is at (%d,%d)", cx, cy)
	}
}

// =============================================================================
// ActionConsume (Eating from Inventory) Tests - 5.3
// =============================================================================

func TestApplyIntent_ActionConsume_ConsumesFromInventory(t *testing.T) {
	t.Parallel()

	// Setup: character carrying an edible item with ActionConsume intent
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = carriedItem
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		TargetX:    5,
		TargetY:    5,
		TargetItem: carriedItem,
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent with enough time to complete action
	// ActionDuration is typically around 1.0s, so we'll apply multiple small deltas
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: inventory should be cleared
	if char.Carrying != nil {
		t.Error("Expected Carrying to be nil after ActionConsume")
	}

	// Assert: hunger should be reduced
	if char.Hunger >= 80 {
		t.Errorf("Expected hunger to be reduced from 80, got %.2f", char.Hunger)
	}
}

func TestApplyIntent_ActionConsume_RequiresDuration(t *testing.T) {
	t.Parallel()

	// Setup: character carrying an edible item with ActionConsume intent
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = carriedItem
	char.ActionProgress = 0
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		TargetX:    5,
		TargetY:    5,
		TargetItem: carriedItem,
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent with small delta (not enough to complete)
	m.applyIntent(char, 0.1)

	// Assert: action progress should increase but item not consumed yet
	if char.ActionProgress == 0 {
		t.Error("Expected ActionProgress to increase")
	}
	if char.Carrying == nil {
		t.Error("Expected item to NOT be consumed yet (duration not complete)")
	}
}

func TestApplyIntent_ActionConsume_VerifiesTargetMatchesCarrying(t *testing.T) {
	t.Parallel()

	// Setup: character with intent targeting different item than what they're carrying
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	differentItem := entity.NewBerry(0, 0, types.ColorBlue, false, false) // Different item
	char.Carrying = carriedItem
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		TargetX:    5,
		TargetY:    5,
		TargetItem: differentItem, // Mismatched!
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent with enough time
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: item should NOT be consumed (target mismatch)
	if char.Carrying == nil {
		t.Error("Expected item to NOT be consumed when target doesn't match carrying")
	}
}
