package types

import "testing"

// =============================================================================
// Position.DistanceTo (Manhattan distance)
// =============================================================================

func TestPosition_DistanceTo_SamePosition(t *testing.T) {
	t.Parallel()

	p := Position{5, 5}
	got := p.DistanceTo(Position{5, 5})
	if got != 0 {
		t.Errorf("DistanceTo same position: got %d, want 0", got)
	}
}

func TestPosition_DistanceTo_Horizontal(t *testing.T) {
	t.Parallel()

	p := Position{0, 0}
	got := p.DistanceTo(Position{5, 0})
	if got != 5 {
		t.Errorf("DistanceTo horizontal: got %d, want 5", got)
	}
}

func TestPosition_DistanceTo_Vertical(t *testing.T) {
	t.Parallel()

	p := Position{0, 0}
	got := p.DistanceTo(Position{0, 7})
	if got != 7 {
		t.Errorf("DistanceTo vertical: got %d, want 7", got)
	}
}

func TestPosition_DistanceTo_Diagonal(t *testing.T) {
	t.Parallel()

	p := Position{0, 0}
	got := p.DistanceTo(Position{3, 4})
	if got != 7 {
		t.Errorf("DistanceTo diagonal: got %d, want 7", got)
	}
}

func TestPosition_DistanceTo_IsSymmetric(t *testing.T) {
	t.Parallel()

	p1 := Position{2, 3}
	p2 := Position{7, 9}

	d1 := p1.DistanceTo(p2)
	d2 := p2.DistanceTo(p1)
	if d1 != d2 {
		t.Errorf("Distance should be symmetric: got %d and %d", d1, d2)
	}
}

func TestPosition_DistanceTo_NegativeCoords(t *testing.T) {
	t.Parallel()

	p1 := Position{-2, -3}
	p2 := Position{2, 3}
	got := p1.DistanceTo(p2)
	if got != 10 {
		t.Errorf("DistanceTo with negative coords: got %d, want 10", got)
	}
}

// =============================================================================
// Position.IsAdjacentTo (8-directional)
// =============================================================================

func TestPosition_IsAdjacentTo_AllEightDirections(t *testing.T) {
	t.Parallel()

	center := Position{5, 5}
	adjacent := []Position{
		{4, 4}, {5, 4}, {6, 4}, // Top row
		{4, 5}, {6, 5}, // Middle (excluding center)
		{4, 6}, {5, 6}, {6, 6}, // Bottom row
	}

	for _, pos := range adjacent {
		if !center.IsAdjacentTo(pos) {
			t.Errorf("Position %v should be adjacent to %v", pos, center)
		}
	}
}

func TestPosition_IsAdjacentTo_SamePosition(t *testing.T) {
	t.Parallel()

	p := Position{5, 5}
	if p.IsAdjacentTo(p) {
		t.Error("Same position should not be considered adjacent")
	}
}

func TestPosition_IsAdjacentTo_TwoAway(t *testing.T) {
	t.Parallel()

	center := Position{5, 5}
	distant := []Position{
		{3, 3}, {5, 3}, {7, 3},
		{3, 5}, {7, 5},
		{3, 7}, {5, 7}, {7, 7},
	}

	for _, pos := range distant {
		if center.IsAdjacentTo(pos) {
			t.Errorf("Position %v should not be adjacent to %v", pos, center)
		}
	}
}

// =============================================================================
// Position.IsCardinallyAdjacentTo (4-directional: N/S/E/W)
// =============================================================================

func TestPosition_IsCardinallyAdjacentTo_CardinalDirections(t *testing.T) {
	t.Parallel()

	center := Position{5, 5}
	cardinal := []Position{
		{5, 4}, // North
		{5, 6}, // South
		{6, 5}, // East
		{4, 5}, // West
	}

	for _, pos := range cardinal {
		if !center.IsCardinallyAdjacentTo(pos) {
			t.Errorf("Position %v should be cardinally adjacent to %v", pos, center)
		}
	}
}

func TestPosition_IsCardinallyAdjacentTo_DiagonalsAreFalse(t *testing.T) {
	t.Parallel()

	center := Position{5, 5}
	diagonals := []Position{
		{4, 4}, // NW
		{6, 4}, // NE
		{4, 6}, // SW
		{6, 6}, // SE
	}

	for _, pos := range diagonals {
		if center.IsCardinallyAdjacentTo(pos) {
			t.Errorf("Diagonal position %v should not be cardinally adjacent to %v", pos, center)
		}
	}
}

func TestPosition_IsCardinallyAdjacentTo_SamePosition(t *testing.T) {
	t.Parallel()

	p := Position{5, 5}
	if p.IsCardinallyAdjacentTo(p) {
		t.Error("Same position should not be cardinally adjacent")
	}
}

func TestPosition_IsCardinallyAdjacentTo_TwoAway(t *testing.T) {
	t.Parallel()

	center := Position{5, 5}
	if center.IsCardinallyAdjacentTo(Position{5, 3}) {
		t.Error("Position two steps away should not be cardinally adjacent")
	}
}

// =============================================================================
// Position.NextStepToward
// =============================================================================

func TestPosition_NextStepToward_SamePosition(t *testing.T) {
	t.Parallel()

	p := Position{5, 5}
	got := p.NextStepToward(Position{5, 5})
	if got != p {
		t.Errorf("NextStepToward same position: got %v, want %v", got, p)
	}
}

func TestPosition_NextStepToward_CardinalDirections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		from   Position
		target Position
		want   Position
	}{
		{"north", Position{5, 5}, Position{5, 0}, Position{5, 4}},
		{"south", Position{5, 5}, Position{5, 10}, Position{5, 6}},
		{"east", Position{5, 5}, Position{10, 5}, Position{6, 5}},
		{"west", Position{5, 5}, Position{0, 5}, Position{4, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.NextStepToward(tt.target)
			if got != tt.want {
				t.Errorf("NextStepToward %s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPosition_NextStepToward_DiagonalDirections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		from   Position
		target Position
		want   Position
	}{
		{"northeast", Position{5, 5}, Position{10, 0}, Position{6, 4}},
		{"northwest", Position{5, 5}, Position{0, 0}, Position{4, 4}},
		{"southeast", Position{5, 5}, Position{10, 10}, Position{6, 6}},
		{"southwest", Position{5, 5}, Position{0, 10}, Position{4, 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.NextStepToward(tt.target)
			if got != tt.want {
				t.Errorf("NextStepToward %s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPosition_NextStepToward_AlreadyAdjacent(t *testing.T) {
	t.Parallel()

	from := Position{5, 5}
	target := Position{6, 5}
	got := from.NextStepToward(target)
	if got != target {
		t.Errorf("NextStepToward adjacent: got %v, want %v", got, target)
	}
}

// =============================================================================
// Abs helper
// =============================================================================

func TestAbs_Zero(t *testing.T) {
	t.Parallel()

	if Abs(0) != 0 {
		t.Errorf("Abs(0): got %d, want 0", Abs(0))
	}
}

func TestAbs_Positive(t *testing.T) {
	t.Parallel()

	if Abs(5) != 5 {
		t.Errorf("Abs(5): got %d, want 5", Abs(5))
	}
}

func TestAbs_Negative(t *testing.T) {
	t.Parallel()

	if Abs(-5) != 5 {
		t.Errorf("Abs(-5): got %d, want 5", Abs(-5))
	}
}

// =============================================================================
// Sign helper
// =============================================================================

func TestSign_Zero(t *testing.T) {
	t.Parallel()

	if Sign(0) != 0 {
		t.Errorf("Sign(0): got %d, want 0", Sign(0))
	}
}

func TestSign_Positive(t *testing.T) {
	t.Parallel()

	if Sign(5) != 1 {
		t.Errorf("Sign(5): got %d, want 1", Sign(5))
	}
	if Sign(100) != 1 {
		t.Errorf("Sign(100): got %d, want 1", Sign(100))
	}
}

func TestSign_Negative(t *testing.T) {
	t.Parallel()

	if Sign(-5) != -1 {
		t.Errorf("Sign(-5): got %d, want -1", Sign(-5))
	}
	if Sign(-100) != -1 {
		t.Errorf("Sign(-100): got %d, want -1", Sign(-100))
	}
}
