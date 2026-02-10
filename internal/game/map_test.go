package game

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

// Helper functions for creating test entities

func newTestCharacter(id, x, y int) *entity.Character {
	return entity.NewCharacter(id, x, y, "Test", "berry", types.ColorRed)
}

func newTestBed(x, y int) *entity.Feature {
	return entity.NewLeafPile(x, y)
}

// =============================================================================
// Character Management Tests
// =============================================================================

func TestAddCharacter_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)

	ok := m.AddCharacter(c)
	if !ok {
		t.Error("AddCharacter() should succeed for empty position")
	}

	got := m.CharacterAt(types.Position{X: 5, Y: 5})
	if got != c {
		t.Errorf("CharacterAt(5,5) should return added character, got %v", got)
	}
}

func TestAddCharacter_OccupiedPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c1 := newTestCharacter(1, 5, 5)
	c2 := newTestCharacter(2, 5, 5)

	m.AddCharacter(c1)
	ok := m.AddCharacter(c2)
	if ok {
		t.Error("AddCharacter() should fail for occupied position")
	}

	got := m.CharacterAt(types.Position{X: 5, Y: 5})
	if got != c1 {
		t.Error("CharacterAt(5,5) should still return original character")
	}
}

func TestCharacterAt_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	got := m.CharacterAt(types.Position{X: 10, Y: 10})
	if got != nil {
		t.Errorf("CharacterAt() for empty position should return nil, got %v", got)
	}
}

func TestCharacters_ReturnsAllAdded(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	chars := []*entity.Character{
		newTestCharacter(1, 1, 1),
		newTestCharacter(2, 2, 2),
		newTestCharacter(3, 3, 3),
		newTestCharacter(4, 4, 4),
	}

	for _, c := range chars {
		m.AddCharacter(c)
	}

	got := m.Characters()
	if len(got) != 4 {
		t.Errorf("Characters() should return 4 characters, got %d", len(got))
	}

	// Verify all characters are present
	charMap := make(map[int]bool)
	for _, c := range got {
		charMap[c.ID] = true
	}
	for _, c := range chars {
		if !charMap[c.ID] {
			t.Errorf("Characters() missing character with ID %d", c.ID)
		}
	}
}

// =============================================================================
// Character Movement Tests
// =============================================================================

func TestMoveCharacter_ToEmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)
	m.AddCharacter(c)

	ok := m.MoveCharacter(c, types.Position{X: 6, Y: 5})
	if !ok {
		t.Error("MoveCharacter() should succeed for empty target position")
	}

	// Old position should be empty
	if m.CharacterAt(types.Position{X: 5, Y: 5}) != nil {
		t.Error("CharacterAt(5,5) should return nil after move")
	}

	// New position should have character
	if m.CharacterAt(types.Position{X: 6, Y: 5}) != c {
		t.Error("CharacterAt(6,5) should return moved character")
	}

	// Character's position should be updated
	pos := c.Pos()
	if pos.X != 6 || pos.Y != 5 {
		t.Errorf("Character position should be (6,5), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestMoveCharacter_ToOccupiedPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c1 := newTestCharacter(1, 5, 5)
	c2 := newTestCharacter(2, 6, 5)
	m.AddCharacter(c1)
	m.AddCharacter(c2)

	ok := m.MoveCharacter(c1, types.Position{X: 6, Y: 5})
	if ok {
		t.Error("MoveCharacter() should fail for occupied target position")
	}

	// Both characters should remain at original positions
	if m.CharacterAt(types.Position{X: 5, Y: 5}) != c1 {
		t.Error("CharacterAt(5,5) should still return c1")
	}
	if m.CharacterAt(types.Position{X: 6, Y: 5}) != c2 {
		t.Error("CharacterAt(6,5) should still return c2")
	}
}

func TestIsOccupied_OccupiedPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)
	m.AddCharacter(c)

	if !m.IsOccupied(types.Position{X: 5, Y: 5}) {
		t.Error("IsOccupied() should return true for occupied position")
	}
}

func TestIsOccupied_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	if m.IsOccupied(types.Position{X: 10, Y: 10}) {
		t.Error("IsOccupied() should return false for empty position")
	}
}

// =============================================================================
// Feature Finding Tests
// =============================================================================

func TestFindNearestBed_ReturnsClosest(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	bed1 := newTestBed(10, 10)
	bed2 := newTestBed(20, 20)
	m.AddFeature(bed1)
	m.AddFeature(bed2)

	// Character at (12, 10) - closer to bed1 at (10, 10)
	got := m.FindNearestBed(types.Position{X: 12, Y: 10})
	if got != bed1 {
		t.Error("FindNearestBed() should return closest bed")
	}
}

func TestFindNearestBed_SkipsOccupied(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	bed1 := newTestBed(10, 10)
	bed2 := newTestBed(20, 20)
	m.AddFeature(bed1)
	m.AddFeature(bed2)

	// Put character A at bed1, occupying it
	charA := newTestCharacter(1, 10, 10)
	m.AddCharacter(charA)

	// Character B at (12, 10) should skip occupied bed1
	got := m.FindNearestBed(types.Position{X: 12, Y: 10})
	if got != bed2 {
		t.Error("FindNearestBed() should skip occupied beds")
	}
}

func TestFindNearestBed_NoBeds(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	got := m.FindNearestBed(types.Position{X: 10, Y: 10})
	if got != nil {
		t.Error("FindNearestBed() should return nil when no beds exist")
	}
}

// =============================================================================
// Position Validation Tests
// =============================================================================

func TestIsValid_InsideBounds(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	tests := []struct {
		name string
		x, y int
	}{
		{"origin", 0, 0},
		{"max corner", 19, 19},
		{"center", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !m.IsValid(types.Position{X: tt.x, Y: tt.y}) {
				t.Errorf("IsValid(%d,%d) should return true", tt.x, tt.y)
			}
		})
	}
}

func TestIsValid_OutsideBounds(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	tests := []struct {
		name string
		x, y int
	}{
		{"negative x", -1, 0},
		{"negative y", 0, -1},
		{"x at width", 20, 0},
		{"y at height", 0, 20},
		{"both outside", 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if m.IsValid(types.Position{X: tt.x, Y: tt.y}) {
				t.Errorf("IsValid(%d,%d) should return false", tt.x, tt.y)
			}
		})
	}
}

// =============================================================================
// Impassable Feature Tests
// =============================================================================

func TestIsBlocked_ImpassableFeature(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	spring := entity.NewSpring(10, 10)
	m.AddFeature(spring)

	if !m.IsBlocked(types.Position{X: 10, Y: 10}) {
		t.Error("IsBlocked() should return true for impassable spring")
	}
}

func TestIsBlocked_PassableFeature(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	bed := newTestBed(10, 10)
	m.AddFeature(bed)

	if m.IsBlocked(types.Position{X: 10, Y: 10}) {
		t.Error("IsBlocked() should return false for passable leaf pile")
	}
}

func TestIsBlocked_Character(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 10, 10)
	m.AddCharacter(c)

	if !m.IsBlocked(types.Position{X: 10, Y: 10}) {
		t.Error("IsBlocked() should return true for character position")
	}
}

func TestIsBlocked_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	if m.IsBlocked(types.Position{X: 10, Y: 10}) {
		t.Error("IsBlocked() should return false for empty position")
	}
}

func TestMoveCharacter_BlockedByImpassableFeature(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)
	m.AddCharacter(c)

	// Add impassable spring at target
	spring := entity.NewSpring(6, 5)
	m.AddFeature(spring)

	ok := m.MoveCharacter(c, types.Position{X: 6, Y: 5})
	if ok {
		t.Error("MoveCharacter() should fail when target has impassable feature")
	}

	// Character should remain at original position
	pos := c.Pos()
	if pos.X != 5 || pos.Y != 5 {
		t.Errorf("Character position should be (5,5), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestMoveCharacter_AllowedOntoPassableFeature(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)
	m.AddCharacter(c)

	// Add passable leaf pile at target
	bed := newTestBed(6, 5)
	m.AddFeature(bed)

	ok := m.MoveCharacter(c, types.Position{X: 6, Y: 5})
	if !ok {
		t.Error("MoveCharacter() should succeed when target has passable feature")
	}

	// Character should be at new position
	pos := c.Pos()
	if pos.X != 6 || pos.Y != 5 {
		t.Errorf("Character position should be (6,5), got (%d,%d)", pos.X, pos.Y)
	}
}

// =============================================================================
// Water Terrain Tests
// =============================================================================

func TestIsWater_WaterTile(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.AddWater(types.Position{X: 5, Y: 5}, WaterSpring)

	if !m.IsWater(types.Position{X: 5, Y: 5}) {
		t.Error("IsWater() should return true for water tile")
	}
}

func TestIsWater_EmptyTile(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	if m.IsWater(types.Position{X: 5, Y: 5}) {
		t.Error("IsWater() should return false for empty tile")
	}
}

func TestWaterAt_ReturnsCorrectType(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.AddWater(types.Position{X: 5, Y: 5}, WaterSpring)
	m.AddWater(types.Position{X: 10, Y: 10}, WaterPond)

	if m.WaterAt(types.Position{X: 5, Y: 5}) != WaterSpring {
		t.Error("WaterAt() should return WaterSpring for spring tile")
	}
	if m.WaterAt(types.Position{X: 10, Y: 10}) != WaterPond {
		t.Error("WaterAt() should return WaterPond for pond tile")
	}
}

func TestIsBlocked_WaterTile(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.AddWater(types.Position{X: 10, Y: 10}, WaterPond)

	if !m.IsBlocked(types.Position{X: 10, Y: 10}) {
		t.Error("IsBlocked() should return true for water tile")
	}
}

func TestFindNearestWater_ReturnsClosest(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	m.AddWater(types.Position{X: 10, Y: 10}, WaterSpring)
	m.AddWater(types.Position{X: 20, Y: 20}, WaterPond)

	// Character at (12, 10) - closer to water at (10, 10)
	pos, found := m.FindNearestWater(types.Position{X: 12, Y: 10})
	if !found {
		t.Fatal("FindNearestWater() should find water")
	}
	if pos.X != 10 || pos.Y != 10 {
		t.Errorf("FindNearestWater() should return closest water at (10,10), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestFindNearestWater_SkipsWhenAllAdjacentBlocked(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	m.AddWater(types.Position{X: 10, Y: 10}, WaterSpring)

	// Block all 4 cardinal-adjacent tiles
	m.AddCharacter(newTestCharacter(1, 10, 9))  // North
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	m.AddCharacter(newTestCharacter(4, 9, 10))  // West

	_, found := m.FindNearestWater(types.Position{X: 15, Y: 15})
	if found {
		t.Error("FindNearestWater() should return false when all adjacent tiles blocked")
	}
}

func TestFindNearestWater_NoWater(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	_, found := m.FindNearestWater(types.Position{X: 10, Y: 10})
	if found {
		t.Error("FindNearestWater() should return false when no water exists")
	}
}

func TestFindNearestWater_AllowsRequesterAtAdjacentTile(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	m.AddWater(types.Position{X: 10, Y: 10}, WaterPond)

	// Requester at adjacent tile
	requester := newTestCharacter(1, 10, 9) // North of water
	m.AddCharacter(requester)

	// Block other 3 adjacent tiles
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	m.AddCharacter(newTestCharacter(4, 9, 10))  // West

	// Requester's position counts as available
	_, found := m.FindNearestWater(types.Position{X: 10, Y: 9})
	if !found {
		t.Error("FindNearestWater() should allow requester's adjacent position as available")
	}
}

func TestWaterPositions_ReturnsAll(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.AddWater(types.Position{X: 5, Y: 5}, WaterSpring)
	m.AddWater(types.Position{X: 10, Y: 10}, WaterPond)
	m.AddWater(types.Position{X: 11, Y: 10}, WaterPond)

	positions := m.WaterPositions()
	if len(positions) != 3 {
		t.Errorf("WaterPositions() should return 3 positions, got %d", len(positions))
	}
}

func TestFindEmptySpot_RespectsWater(t *testing.T) {
	// Create a tiny map where most tiles are water
	// The only non-water tile should always be chosen
	m := NewMap(3, 1) // 3 tiles wide, 1 tile tall
	m.AddWater(types.Position{X: 0, Y: 0}, WaterPond)
	m.AddWater(types.Position{X: 2, Y: 0}, WaterPond)
	// Only (1, 0) is free

	x, y := findEmptySpot(m)
	if x != 1 || y != 0 {
		t.Errorf("findEmptySpot() should avoid water tiles, got (%d,%d)", x, y)
	}
}

func TestIsEmpty_WaterTile(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.AddWater(types.Position{X: 5, Y: 5}, WaterPond)

	if m.IsEmpty(types.Position{X: 5, Y: 5}) {
		t.Error("IsEmpty() should return false for water tile")
	}
}

// =============================================================================
// Tilled Soil Tests
// =============================================================================

func TestSetTilled_IsTilled(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}

	m.SetTilled(pos)

	if !m.IsTilled(pos) {
		t.Error("IsTilled() should return true after SetTilled()")
	}
}

func TestIsTilled_NonTilledPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	if m.IsTilled(types.Position{X: 5, Y: 5}) {
		t.Error("IsTilled() should return false for non-tilled position")
	}
}

func TestTilledPositions_ReturnsAll(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.SetTilled(types.Position{X: 3, Y: 3})
	m.SetTilled(types.Position{X: 4, Y: 3})
	m.SetTilled(types.Position{X: 5, Y: 3})

	positions := m.TilledPositions()
	if len(positions) != 3 {
		t.Errorf("TilledPositions() should return 3 positions, got %d", len(positions))
	}

	// Verify all expected positions are present
	posSet := make(map[types.Position]bool)
	for _, p := range positions {
		posSet[p] = true
	}
	for _, expected := range []types.Position{{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 5, Y: 3}} {
		if !posSet[expected] {
			t.Errorf("TilledPositions() missing position (%d,%d)", expected.X, expected.Y)
		}
	}
}

func TestIsBlocked_TilledTile(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 10, Y: 10}
	m.SetTilled(pos)

	if m.IsBlocked(pos) {
		t.Error("IsBlocked() should return false for tilled tile (tilled soil is walkable)")
	}
}

func TestIsEmpty_TilledTileNoOccupants(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 10, Y: 10}
	m.SetTilled(pos)

	if !m.IsEmpty(pos) {
		t.Error("IsEmpty() should return true for tilled tile with no item/character/feature")
	}
}

func TestIsEmpty_TilledTileWithItem(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 10, Y: 10}
	m.SetTilled(pos)

	item := entity.NewBerry(10, 10, types.ColorRed, false, false)
	m.AddItem(item)

	if m.IsEmpty(pos) {
		t.Error("IsEmpty() should return false for tilled tile with an item on it")
	}
}

// Marked-for-Tilling Tests

func TestMarkForTilling_IsMarkedForTilling(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}

	ok := m.MarkForTilling(pos)

	if !ok {
		t.Error("MarkForTilling() should return true for new position")
	}
	if !m.IsMarkedForTilling(pos) {
		t.Error("IsMarkedForTilling() should return true after MarkForTilling()")
	}
}

func TestIsMarkedForTilling_UnmarkedPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	if m.IsMarkedForTilling(types.Position{X: 5, Y: 5}) {
		t.Error("IsMarkedForTilling() should return false for unmarked position")
	}
}

func TestUnmarkForTilling(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.MarkForTilling(pos)

	m.UnmarkForTilling(pos)

	if m.IsMarkedForTilling(pos) {
		t.Error("IsMarkedForTilling() should return false after UnmarkForTilling()")
	}
}

func TestMarkedForTillingPositions_ReturnsAll(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	m.MarkForTilling(types.Position{X: 3, Y: 3})
	m.MarkForTilling(types.Position{X: 4, Y: 3})
	m.MarkForTilling(types.Position{X: 5, Y: 3})

	positions := m.MarkedForTillingPositions()
	if len(positions) != 3 {
		t.Errorf("MarkedForTillingPositions() should return 3 positions, got %d", len(positions))
	}

	expected := map[types.Position]bool{
		{X: 3, Y: 3}: false,
		{X: 4, Y: 3}: false,
		{X: 5, Y: 3}: false,
	}
	for _, p := range positions {
		if _, ok := expected[p]; !ok {
			t.Errorf("MarkedForTillingPositions() returned unexpected position (%d,%d)", p.X, p.Y)
		}
	}
}

func TestMarkForTilling_AlreadyTilledIsNoOp(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 5, Y: 5}
	m.SetTilled(pos)

	ok := m.MarkForTilling(pos)

	if ok {
		t.Error("MarkForTilling() should return false for already-tilled position")
	}
	if m.IsMarkedForTilling(pos) {
		t.Error("IsMarkedForTilling() should return false for already-tilled position")
	}
}

func TestIsBlocked_MarkedTile(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 10, Y: 10}
	m.MarkForTilling(pos)

	if m.IsBlocked(pos) {
		t.Error("IsBlocked() should return false for marked-for-tilling tile")
	}
}

func TestIsEmpty_MarkedTileNoOccupants(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	pos := types.Position{X: 10, Y: 10}
	m.MarkForTilling(pos)

	if !m.IsEmpty(pos) {
		t.Error("IsEmpty() should return true for marked-for-tilling tile with no occupants")
	}
}
