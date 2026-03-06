package system

import (
	"petri/internal/game"
	"petri/internal/types"
)

// NextStepBFS calculates the next position moving toward target using greedy-first pathfinding.
// Prefers the greedy diagonal step (alternating X/Y based on larger delta) for natural
// zigzag movement that spreads characters across different paths. Falls back to BFS only
// when the greedy step is blocked by terrain (water, impassable features).
// Ignores characters since they move and per-tick collision is handled separately.
// Falls back to greedy NextStep if no BFS path exists either.
func NextStepBFS(fromX, fromY, toX, toY int, gameMap *game.Map) (int, int) {
	nx, ny, _ := nextStepBFSCore(fromX, fromY, toX, toY, gameMap, false)
	return nx, ny
}

// nextStepBFSCore is the internal pathfinding implementation.
// When preferBFS is true, skips the greedy step and goes straight to BFS.
// Returns usedBFS=true whenever BFS was actually used (greedy was skipped or blocked).
func nextStepBFSCore(fromX, fromY, toX, toY int, gameMap *game.Map, preferBFS bool) (int, int, bool) {
	if fromX == toX && fromY == toY {
		return fromX, fromY, false
	}

	// Nil map fallback - used in tests that don't need pathfinding
	if gameMap == nil {
		nx, ny := NextStep(fromX, fromY, toX, toY)
		return nx, ny, false
	}

	// Try greedy step first (unless preferBFS forces BFS)
	if !preferBFS {
		gx, gy := NextStep(fromX, fromY, toX, toY)
		greedyPos := types.Position{X: gx, Y: gy}
		if gameMap.IsValid(greedyPos) && !gameMap.IsWater(greedyPos) {
			if f := gameMap.FeatureAt(greedyPos); f == nil || f.IsPassable() {
				if c := gameMap.ConstructAt(greedyPos); c == nil || c.IsPassable() {
					return gx, gy, false
				}
			}
		}
	}

	from := types.Position{X: fromX, Y: fromY}
	to := types.Position{X: toX, Y: toY}

	// BFS tracking the first step from origin that leads to each visited tile
	type node struct {
		pos       types.Position
		firstStep types.Position
	}

	visited := make(map[types.Position]bool)
	visited[from] = true

	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	// Seed queue with walkable neighbors of start
	var queue []node
	for _, dir := range cardinalDirs {
		neighbor := types.Position{X: fromX + dir[0], Y: fromY + dir[1]}
		if !gameMap.IsValid(neighbor) || visited[neighbor] {
			continue
		}
		if gameMap.IsWater(neighbor) {
			continue
		}
		if f := gameMap.FeatureAt(neighbor); f != nil && !f.IsPassable() {
			continue
		}
		if c := gameMap.ConstructAt(neighbor); c != nil && !c.IsPassable() {
			continue
		}
		visited[neighbor] = true
		if neighbor == to {
			return neighbor.X, neighbor.Y, true
		}
		queue = append(queue, node{pos: neighbor, firstStep: neighbor})
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, dir := range cardinalDirs {
			neighbor := types.Position{X: cur.pos.X + dir[0], Y: cur.pos.Y + dir[1]}
			if !gameMap.IsValid(neighbor) || visited[neighbor] {
				continue
			}
			if gameMap.IsWater(neighbor) {
				continue
			}
			if f := gameMap.FeatureAt(neighbor); f != nil && !f.IsPassable() {
				continue
			}
			if c := gameMap.ConstructAt(neighbor); c != nil && !c.IsPassable() {
				continue
			}
			visited[neighbor] = true
			if neighbor == to {
				return cur.firstStep.X, cur.firstStep.Y, true
			}
			queue = append(queue, node{pos: neighbor, firstStep: cur.firstStep})
		}
	}

	// No path found - fall back to greedy
	nx, ny := NextStep(fromX, fromY, toX, toY)
	return nx, ny, false
}

// NextStep calculates the next position moving toward target
func NextStep(fromX, fromY, toX, toY int) (int, int) {
	dx := toX - fromX
	dy := toY - fromY

	if dx == 0 && dy == 0 {
		return fromX, fromY
	}

	// Move toward larger distance
	if types.Abs(dx) > types.Abs(dy) {
		return fromX + types.Sign(dx), fromY
	}
	return fromX, fromY + types.Sign(dy)
}

// isAdjacent checks if two positions are adjacent (including diagonals)
func isAdjacent(x1, y1, x2, y2 int) bool {
	return types.Position{X: x1, Y: y1}.IsAdjacentTo(types.Position{X: x2, Y: y2})
}

// isCardinallyAdjacent checks 4-direction adjacency (N/E/S/W, no diagonals)
func isCardinallyAdjacent(x1, y1, x2, y2 int) bool {
	return types.Position{X: x1, Y: y1}.IsCardinallyAdjacentTo(types.Position{X: x2, Y: y2})
}

// FindClosestCardinalTile finds closest unblocked cardinally adjacent tile to target
func FindClosestCardinalTile(cx, cy, tx, ty int, gameMap *game.Map) (int, int) {
	pos := types.Position{X: cx, Y: cy}
	directions := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	bestX, bestY := -1, -1
	bestDist := int(^uint(0) >> 1)

	for _, dir := range directions {
		adjPos := types.Position{X: tx + dir[0], Y: ty + dir[1]}
		if !gameMap.IsValid(adjPos) || gameMap.IsBlocked(adjPos) {
			continue
		}
		if dist := pos.DistanceTo(adjPos); dist < bestDist {
			bestDist, bestX, bestY = dist, adjPos.X, adjPos.Y
		}
	}
	return bestX, bestY
}

// findClosestAdjacentTile finds the closest unoccupied tile adjacent to (tx, ty) from position (cx, cy)
func findClosestAdjacentTile(cx, cy, tx, ty int, gameMap *game.Map) (int, int) {
	pos := types.Position{X: cx, Y: cy}
	// 8 directions
	directions := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}

	bestX, bestY := -1, -1
	bestDist := int(^uint(0) >> 1)

	for _, dir := range directions {
		adjPos := types.Position{X: tx + dir[0], Y: ty + dir[1]}
		if !gameMap.IsValid(adjPos) {
			continue
		}
		if gameMap.IsOccupied(adjPos) {
			continue
		}

		dist := pos.DistanceTo(adjPos)
		if dist < bestDist {
			bestDist = dist
			bestX, bestY = adjPos.X, adjPos.Y
		}
	}

	return bestX, bestY
}
