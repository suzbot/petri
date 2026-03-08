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

// getValidLinePositions tests

func TestGetValidLinePositions_Horizontal(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)

	anchor := types.Position{X: 3, Y: 5}
	cursor := types.Position{X: 7, Y: 5}

	positions := getValidLinePositions(anchor, cursor, m, isValidFenceTarget)
	if len(positions) != 5 {
		t.Errorf("getValidLinePositions() horizontal: got %d positions, want 5", len(positions))
	}
	posSet := make(map[types.Position]bool)
	for _, p := range positions {
		posSet[p] = true
		if p.Y != 5 {
			t.Errorf("getValidLinePositions() horizontal: position (%d,%d) has wrong Y, want 5", p.X, p.Y)
		}
	}
}

func TestGetValidLinePositions_Vertical(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)

	anchor := types.Position{X: 5, Y: 3}
	cursor := types.Position{X: 5, Y: 7}

	positions := getValidLinePositions(anchor, cursor, m, isValidFenceTarget)
	if len(positions) != 5 {
		t.Errorf("getValidLinePositions() vertical: got %d positions, want 5", len(positions))
	}
	for _, p := range positions {
		if p.X != 5 {
			t.Errorf("getValidLinePositions() vertical: position (%d,%d) has wrong X, want 5", p.X, p.Y)
		}
	}
}

func TestGetValidLinePositions_DiagonalSnapsToLargerAxis(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)

	// dx=4, dy=2 — horizontal axis wins
	anchor := types.Position{X: 5, Y: 5}
	cursor := types.Position{X: 9, Y: 7}

	positions := getValidLinePositions(anchor, cursor, m, isValidFenceTarget)
	// Should snap to horizontal: y=5, x from 5 to 9 = 5 positions
	if len(positions) != 5 {
		t.Errorf("getValidLinePositions() diagonal (dx>dy): got %d positions, want 5 (horizontal snap)", len(positions))
	}
	for _, p := range positions {
		if p.Y != 5 {
			t.Errorf("getValidLinePositions() diagonal: snapped to horizontal but got Y=%d, want 5", p.Y)
		}
	}

	// dx=1, dy=3 — vertical axis wins
	anchor2 := types.Position{X: 5, Y: 5}
	cursor2 := types.Position{X: 6, Y: 8}
	positions2 := getValidLinePositions(anchor2, cursor2, m, isValidFenceTarget)
	// Should snap to vertical: x=5, y from 5 to 8 = 4 positions
	if len(positions2) != 4 {
		t.Errorf("getValidLinePositions() diagonal (dy>dx): got %d positions, want 4 (vertical snap)", len(positions2))
	}
	for _, p := range positions2 {
		if p.X != 5 {
			t.Errorf("getValidLinePositions() diagonal: snapped to vertical but got X=%d, want 5", p.X)
		}
	}
}

func TestGetValidLinePositions_FiltersInvalidTiles(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	// Water tile in the middle of the line
	m.AddWater(types.Position{X: 5, Y: 5}, game.WaterPond)

	anchor := types.Position{X: 3, Y: 5}
	cursor := types.Position{X: 7, Y: 5}

	positions := getValidLinePositions(anchor, cursor, m, isValidFenceTarget)
	// Should return 4 positions (3,5),(4,5),(6,5),(7,5) — (5,5) excluded by water
	if len(positions) != 4 {
		t.Errorf("getValidLinePositions() with water: got %d positions, want 4", len(positions))
	}
	for _, p := range positions {
		if p == (types.Position{X: 5, Y: 5}) {
			t.Error("getValidLinePositions() should not include water tile at (5,5)")
		}
	}
}

func TestGetValidLinePositions_SingleTile(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)

	anchor := types.Position{X: 5, Y: 5}
	cursor := types.Position{X: 5, Y: 5}

	positions := getValidLinePositions(anchor, cursor, m, isValidFenceTarget)
	if len(positions) != 1 {
		t.Errorf("getValidLinePositions() single tile: got %d positions, want 1", len(positions))
	}
}

func TestIsValidFenceTarget_RejectsWater(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.AddWater(pos, game.WaterPond)

	if isValidFenceTarget(pos, m) {
		t.Error("isValidFenceTarget() should return false for water tiles")
	}
}

func TestIsValidFenceTarget_RejectsExistingConstruct(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	fence := entity.NewFence(5, 5, "stick", types.ColorBrown)
	m.AddConstruct(fence)

	if isValidFenceTarget(pos, m) {
		t.Error("isValidFenceTarget() should return false for tiles with existing constructs")
	}
}

func TestIsValidFenceTarget_RejectsAlreadyMarked(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.MarkForConstruction(pos, 1, "fence")

	if isValidFenceTarget(pos, m) {
		t.Error("isValidFenceTarget() should return false for already-marked-for-construction tiles")
	}
}

func TestIsValidFenceTarget_AcceptsEmptyTile(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}

	if !isValidFenceTarget(pos, m) {
		t.Error("isValidFenceTarget() should return true for empty walkable tiles")
	}
}

func TestIsValidUnmarkFenceTarget_ReturnsTrueForMarked(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.MarkForConstruction(pos, 1, "fence")

	if !isValidUnmarkFenceTarget(pos, m) {
		t.Error("isValidUnmarkFenceTarget() should return true for marked-for-construction tiles without construct")
	}
}

func TestIsValidUnmarkFenceTarget_ReturnsFalseForUnmarked(t *testing.T) {
	t.Parallel()
	m := game.NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}

	if isValidUnmarkFenceTarget(pos, m) {
		t.Error("isValidUnmarkFenceTarget() should return false for unmarked tiles")
	}
}
