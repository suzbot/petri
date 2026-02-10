package ui

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

func TestIsValidTillTarget_WaterTile(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.AddWater(pos, game.WaterPond)

	if isValidTillTarget(pos, m) {
		t.Error("isValidTillTarget() should return false for water tiles")
	}
}

func TestIsValidTillTarget_FeatureTile(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	f := entity.NewLeafPile(5, 5)
	m.AddFeature(f)

	if isValidTillTarget(pos, m) {
		t.Error("isValidTillTarget() should return false for feature tiles")
	}
}

func TestIsValidTillTarget_AlreadyTilled(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.SetTilled(pos)

	if isValidTillTarget(pos, m) {
		t.Error("isValidTillTarget() should return false for already-tilled tiles")
	}
}

func TestIsValidTillTarget_AlreadyMarked(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.MarkForTilling(pos)

	if isValidTillTarget(pos, m) {
		t.Error("isValidTillTarget() should return false for already-marked tiles")
	}
}

func TestIsValidTillTarget_EmptyWalkable(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}

	if !isValidTillTarget(pos, m) {
		t.Error("isValidTillTarget() should return true for empty walkable tiles")
	}
}

func TestIsValidUnmarkTarget_MarkedNotTilled(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.MarkForTilling(pos)

	if !isValidUnmarkTarget(pos, m) {
		t.Error("isValidUnmarkTarget() should return true for marked-but-not-tilled tiles")
	}
}

func TestIsValidUnmarkTarget_Unmarked(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}

	if isValidUnmarkTarget(pos, m) {
		t.Error("isValidUnmarkTarget() should return false for unmarked tiles")
	}
}

func TestIsValidUnmarkTarget_AlreadyTilled(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.SetTilled(pos)

	if isValidUnmarkTarget(pos, m) {
		t.Error("isValidUnmarkTarget() should return false for already-tilled tiles")
	}
}

func TestGetValidPositions_FiltersRectangle(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	// Place water at (6,5) to exclude it
	m.AddWater(types.Position{X: 6, Y: 5}, game.WaterPond)

	anchor := types.Position{X: 5, Y: 5}
	cursor := types.Position{X: 7, Y: 5}

	positions := getValidPositions(anchor, cursor, m, isValidTillTarget)
	// Should include (5,5) and (7,5) but NOT (6,5)
	if len(positions) != 2 {
		t.Errorf("getValidPositions() should return 2 positions, got %d", len(positions))
	}

	posSet := make(map[types.Position]bool)
	for _, p := range positions {
		posSet[p] = true
	}
	if posSet[types.Position{X: 6, Y: 5}] {
		t.Error("getValidPositions() should not include water tile at (6,5)")
	}
}

func TestGetValidPositions_ExcludesOutOfBounds(t *testing.T) {
	t.Parallel()
	m := game.NewMap(10, 10)

	// Anchor inside, cursor outside map bounds
	anchor := types.Position{X: 8, Y: 8}
	cursor := types.Position{X: 11, Y: 11}

	positions := getValidPositions(anchor, cursor, m, isValidTillTarget)
	// Only (8,8), (9,8), (8,9), (9,9) are in bounds
	if len(positions) != 4 {
		t.Errorf("getValidPositions() should return 4 in-bounds positions, got %d", len(positions))
	}
}

func TestGetValidPositions_AnchorGreaterThanCursor(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)

	// Anchor at bottom-right, cursor at top-left (drag up-left)
	anchor := types.Position{X: 7, Y: 7}
	cursor := types.Position{X: 5, Y: 5}

	positions := getValidPositions(anchor, cursor, m, isValidTillTarget)
	// 3x3 rectangle = 9 positions
	if len(positions) != 9 {
		t.Errorf("getValidPositions() should return 9 positions for 3x3 rect, got %d", len(positions))
	}
}
