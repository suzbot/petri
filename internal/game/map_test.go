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

func newTestSpring(x, y int) *entity.Feature {
	return entity.NewSpring(x, y)
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

	got := m.CharacterAt(5, 5)
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

	got := m.CharacterAt(5, 5)
	if got != c1 {
		t.Error("CharacterAt(5,5) should still return original character")
	}
}

func TestCharacterAt_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	got := m.CharacterAt(10, 10)
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

	ok := m.MoveCharacter(c, 6, 5)
	if !ok {
		t.Error("MoveCharacter() should succeed for empty target position")
	}

	// Old position should be empty
	if m.CharacterAt(5, 5) != nil {
		t.Error("CharacterAt(5,5) should return nil after move")
	}

	// New position should have character
	if m.CharacterAt(6, 5) != c {
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

	ok := m.MoveCharacter(c1, 6, 5)
	if ok {
		t.Error("MoveCharacter() should fail for occupied target position")
	}

	// Both characters should remain at original positions
	if m.CharacterAt(5, 5) != c1 {
		t.Error("CharacterAt(5,5) should still return c1")
	}
	if m.CharacterAt(6, 5) != c2 {
		t.Error("CharacterAt(6,5) should still return c2")
	}
}

func TestIsOccupied_OccupiedPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)
	m.AddCharacter(c)

	if !m.IsOccupied(5, 5) {
		t.Error("IsOccupied() should return true for occupied position")
	}
}

func TestIsOccupied_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	if m.IsOccupied(10, 10) {
		t.Error("IsOccupied() should return false for empty position")
	}
}

// =============================================================================
// Feature Finding Tests
// =============================================================================

func TestFindNearestDrinkSource_ReturnsClosest(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring1 := newTestSpring(10, 10)
	spring2 := newTestSpring(20, 20)
	m.AddFeature(spring1)
	m.AddFeature(spring2)

	// Character at (12, 10) - closer to spring1 at (10, 10)
	got := m.FindNearestDrinkSource(12, 10)
	if got != spring1 {
		t.Error("FindNearestDrinkSource() should return closest spring")
	}
}

func TestFindNearestDrinkSource_SkipsWhenAllAdjacentOccupied(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring1 := newTestSpring(10, 10)
	spring2 := newTestSpring(20, 20)
	m.AddFeature(spring1)
	m.AddFeature(spring2)

	// Block all 4 cardinal-adjacent tiles of spring1
	m.AddCharacter(newTestCharacter(1, 10, 9))  // North
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	m.AddCharacter(newTestCharacter(4, 9, 10))  // West

	// Character at (12, 10) should skip blocked spring1
	got := m.FindNearestDrinkSource(12, 10)
	if got != spring2 {
		t.Error("FindNearestDrinkSource() should skip springs with all adjacent tiles blocked")
	}
}

func TestFindNearestDrinkSource_AllowsRequesterAdjacentPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring := newTestSpring(10, 10)
	m.AddFeature(spring)

	// Character standing adjacent to spring (springs are now impassable)
	c := newTestCharacter(1, 10, 9) // North of spring
	m.AddCharacter(c)

	// Block other 3 cardinal tiles
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	m.AddCharacter(newTestCharacter(4, 9, 10))  // West

	// Should still return the spring - requester's adjacent position counts as available
	got := m.FindNearestDrinkSource(10, 9)
	if got != spring {
		t.Error("FindNearestDrinkSource() should allow requesting character's adjacent position")
	}
}

func TestFindNearestDrinkSource_AllAdjacentBlocked(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring1 := newTestSpring(10, 10)
	spring2 := newTestSpring(20, 20)
	m.AddFeature(spring1)
	m.AddFeature(spring2)

	// Block all cardinal tiles around both springs
	// Spring1 at (10, 10)
	m.AddCharacter(newTestCharacter(1, 10, 9))   // North
	m.AddCharacter(newTestCharacter(2, 11, 10))  // East
	m.AddCharacter(newTestCharacter(3, 10, 11))  // South
	m.AddCharacter(newTestCharacter(4, 9, 10))   // West
	// Spring2 at (20, 20)
	m.AddCharacter(newTestCharacter(5, 20, 19))  // North
	m.AddCharacter(newTestCharacter(6, 21, 20))  // East
	m.AddCharacter(newTestCharacter(7, 20, 21))  // South
	m.AddCharacter(newTestCharacter(8, 19, 20))  // West

	// Character at (15, 15) looking for spring
	got := m.FindNearestDrinkSource(15, 15)
	if got != nil {
		t.Error("FindNearestDrinkSource() should return nil when all springs have all adjacent tiles blocked")
	}
}

func TestFindNearestDrinkSource_NoSprings(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	got := m.FindNearestDrinkSource(10, 10)
	if got != nil {
		t.Error("FindNearestDrinkSource() should return nil when no springs exist")
	}
}

func TestFindNearestBed_ReturnsClosest(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	bed1 := newTestBed(10, 10)
	bed2 := newTestBed(20, 20)
	m.AddFeature(bed1)
	m.AddFeature(bed2)

	// Character at (12, 10) - closer to bed1 at (10, 10)
	got := m.FindNearestBed(12, 10)
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
	got := m.FindNearestBed(12, 10)
	if got != bed2 {
		t.Error("FindNearestBed() should skip occupied beds")
	}
}

func TestFindNearestBed_NoBeds(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	got := m.FindNearestBed(10, 10)
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
			if !m.IsValid(tt.x, tt.y) {
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
			if m.IsValid(tt.x, tt.y) {
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
	spring := newTestSpring(10, 10)
	m.AddFeature(spring)

	if !m.IsBlocked(10, 10) {
		t.Error("IsBlocked() should return true for impassable spring")
	}
}

func TestIsBlocked_PassableFeature(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	bed := newTestBed(10, 10)
	m.AddFeature(bed)

	if m.IsBlocked(10, 10) {
		t.Error("IsBlocked() should return false for passable leaf pile")
	}
}

func TestIsBlocked_Character(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 10, 10)
	m.AddCharacter(c)

	if !m.IsBlocked(10, 10) {
		t.Error("IsBlocked() should return true for character position")
	}
}

func TestIsBlocked_EmptyPosition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)

	if m.IsBlocked(10, 10) {
		t.Error("IsBlocked() should return false for empty position")
	}
}

func TestMoveCharacter_BlockedByImpassableFeature(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	c := newTestCharacter(1, 5, 5)
	m.AddCharacter(c)

	// Add impassable spring at target
	spring := newTestSpring(6, 5)
	m.AddFeature(spring)

	ok := m.MoveCharacter(c, 6, 5)
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

	ok := m.MoveCharacter(c, 6, 5)
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
// Cardinal Adjacency Drink Source Tests
// =============================================================================

func TestFindNearestDrinkSource_SkipsWhenAllAdjacentBlocked(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring := newTestSpring(10, 10)
	m.AddFeature(spring)

	// Block all 4 cardinal-adjacent tiles with characters
	m.AddCharacter(newTestCharacter(1, 10, 9))  // North
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	m.AddCharacter(newTestCharacter(4, 9, 10))  // West

	// Character at (15, 15) looking for spring - should find none
	got := m.FindNearestDrinkSource(15, 15)
	if got != nil {
		t.Error("FindNearestDrinkSource() should return nil when all adjacent tiles blocked")
	}
}

func TestFindNearestDrinkSource_AllowsWhenOneAdjacentFree(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring := newTestSpring(10, 10)
	m.AddFeature(spring)

	// Block 3 of 4 cardinal-adjacent tiles
	m.AddCharacter(newTestCharacter(1, 10, 9))  // North
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	// West (9, 10) is free

	// Character at (15, 15) looking for spring - should find it
	got := m.FindNearestDrinkSource(15, 15)
	if got != spring {
		t.Error("FindNearestDrinkSource() should return spring when at least one adjacent tile is free")
	}
}

func TestFindNearestDrinkSource_AllowsRequesterAtAdjacentTile(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring := newTestSpring(10, 10)
	m.AddFeature(spring)

	// Requester is at adjacent tile
	requester := newTestCharacter(1, 10, 9) // North of spring
	m.AddCharacter(requester)

	// Block other 3 adjacent tiles
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East
	m.AddCharacter(newTestCharacter(3, 10, 11)) // South
	m.AddCharacter(newTestCharacter(4, 9, 10))  // West

	// Requester should still find the spring (they're at an adjacent tile)
	got := m.FindNearestDrinkSource(10, 9)
	if got != spring {
		t.Error("FindNearestDrinkSource() should allow requester's own position as available")
	}
}

func TestFindNearestDrinkSource_MultipleDrinkersAtDifferentTiles(t *testing.T) {
	t.Parallel()

	m := NewMap(30, 30)
	spring := newTestSpring(10, 10)
	m.AddFeature(spring)

	// Two characters at adjacent tiles (simulating drinking)
	m.AddCharacter(newTestCharacter(1, 10, 9))  // North
	m.AddCharacter(newTestCharacter(2, 11, 10)) // East

	// Third character looking for spring - should find it (South and West still free)
	got := m.FindNearestDrinkSource(15, 15)
	if got != spring {
		t.Error("FindNearestDrinkSource() should allow spring when some adjacent tiles remain free")
	}
}
