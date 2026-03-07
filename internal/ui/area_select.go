package ui

import (
	"petri/internal/game"
	"petri/internal/types"
)

// isValidTillTarget returns true if the position can be marked for tilling.
// Rejects water, features, already-tilled, and already-marked positions.
func isValidTillTarget(pos types.Position, gameMap *game.Map) bool {
	if gameMap.IsWater(pos) {
		return false
	}
	if gameMap.FeatureAt(pos) != nil {
		return false
	}
	if gameMap.IsTilled(pos) {
		return false
	}
	if gameMap.IsMarkedForTilling(pos) {
		return false
	}
	return true
}

// isValidUnmarkTarget returns true if the position can be unmarked.
// Only marked-but-not-yet-tilled positions are valid unmark targets.
func isValidUnmarkTarget(pos types.Position, gameMap *game.Map) bool {
	return gameMap.IsMarkedForTilling(pos) && !gameMap.IsTilled(pos)
}

// isValidFenceTarget returns true if the position can be marked for fence construction.
// Rejects water, impassable features, existing constructs, and already-marked-for-construction tiles.
func isValidFenceTarget(pos types.Position, gameMap *game.Map) bool {
	if gameMap.IsWater(pos) {
		return false
	}
	if f := gameMap.FeatureAt(pos); f != nil && !f.Passable {
		return false
	}
	if gameMap.ConstructAt(pos) != nil {
		return false
	}
	if gameMap.IsMarkedForConstruction(pos) {
		return false
	}
	return true
}

// isValidUnmarkFenceTarget returns true if the position can be unmarked from the construction pool.
func isValidUnmarkFenceTarget(pos types.Position, gameMap *game.Map) bool {
	return gameMap.IsMarkedForConstruction(pos) && gameMap.ConstructAt(pos) == nil
}

// getValidLinePositions returns valid positions along a cardinal line from anchor to cursor.
// The line is constrained to horizontal or vertical: the axis with the larger delta wins.
// For equal deltas, horizontal wins. The validator filters out invalid positions.
func getValidLinePositions(anchor, cursor types.Position, gameMap *game.Map, validator func(types.Position, *game.Map) bool) []types.Position {
	dx := cursor.X - anchor.X
	if dx < 0 {
		dx = -dx
	}
	dy := cursor.Y - anchor.Y
	if dy < 0 {
		dy = -dy
	}

	var positions []types.Position
	if dy > dx {
		// Vertical line: fix X at anchor.X, iterate Y
		minY, maxY := anchor.Y, cursor.Y
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY; y <= maxY; y++ {
			pos := types.Position{X: anchor.X, Y: y}
			if !gameMap.IsValid(pos) {
				continue
			}
			if validator(pos, gameMap) {
				positions = append(positions, pos)
			}
		}
	} else {
		// Horizontal line (also handles dx==dy tie): fix Y at anchor.Y, iterate X
		minX, maxX := anchor.X, cursor.X
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX; x <= maxX; x++ {
			pos := types.Position{X: x, Y: anchor.Y}
			if !gameMap.IsValid(pos) {
				continue
			}
			if validator(pos, gameMap) {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}

// isOnLine returns true if pos lies on the cardinal line from anchor to cursor.
// Uses the same axis-snapping logic as getValidLinePositions.
func isOnLine(pos, anchor, cursor types.Position) bool {
	dx := cursor.X - anchor.X
	if dx < 0 {
		dx = -dx
	}
	dy := cursor.Y - anchor.Y
	if dy < 0 {
		dy = -dy
	}

	if dy > dx {
		// Vertical line
		if pos.X != anchor.X {
			return false
		}
		minY, maxY := anchor.Y, cursor.Y
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		return pos.Y >= minY && pos.Y <= maxY
	}
	// Horizontal line
	if pos.Y != anchor.Y {
		return false
	}
	minX, maxX := anchor.X, cursor.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	return pos.X >= minX && pos.X <= maxX
}

// isInRect returns true if pos is within the rectangle defined by two corners (any drag direction).
func isInRect(pos, corner1, corner2 types.Position) bool {
	minX, maxX := corner1.X, corner2.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := corner1.Y, corner2.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY
}

// getValidPositions returns all valid positions within the rectangle defined by anchor and cursor.
// The validator function determines which positions are included.
func getValidPositions(anchor, cursor types.Position, gameMap *game.Map, validator func(types.Position, *game.Map) bool) []types.Position {
	minX, maxX := anchor.X, cursor.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := anchor.Y, cursor.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}

	var positions []types.Position
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			pos := types.Position{X: x, Y: y}
			if !gameMap.IsValid(pos) {
				continue
			}
			if validator(pos, gameMap) {
				positions = append(positions, pos)
			}
		}
	}
	return positions
}
