package types

// Position represents a 2D coordinate on the map.
// This is the canonical position type used throughout the codebase.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// DistanceTo returns the Manhattan distance to another position.
func (p Position) DistanceTo(other Position) int {
	return Abs(p.X-other.X) + Abs(p.Y-other.Y)
}

// IsAdjacentTo returns true if the other position is within 1 tile
// in any of the 8 directions (excludes same position).
func (p Position) IsAdjacentTo(other Position) bool {
	dx, dy := Abs(p.X-other.X), Abs(p.Y-other.Y)
	return dx <= 1 && dy <= 1 && !(dx == 0 && dy == 0)
}

// IsCardinallyAdjacentTo returns true if the other position is exactly
// 1 tile away in a cardinal direction (N/S/E/W only).
func (p Position) IsCardinallyAdjacentTo(other Position) bool {
	dx, dy := Abs(p.X-other.X), Abs(p.Y-other.Y)
	return (dx == 1 && dy == 0) || (dx == 0 && dy == 1)
}

// NextStepToward returns the position one step closer to the target.
// Moves diagonally when both X and Y differ.
func (p Position) NextStepToward(target Position) Position {
	return Position{
		X: p.X + Sign(target.X-p.X),
		Y: p.Y + Sign(target.Y-p.Y),
	}
}

// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Sign returns -1, 0, or 1 based on the sign of x.
func Sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
