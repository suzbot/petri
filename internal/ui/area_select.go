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
