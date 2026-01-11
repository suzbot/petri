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
