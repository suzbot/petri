package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
)

// UpdateSpawnTimers decrements spawn timers for all items and spawns new items when timers expire
func UpdateSpawnTimers(gameMap *game.Map, initialItemCount int, delta float64) {
	items := gameMap.Items()

	// Check if map is at capacity (50% of coordinates)
	maxItems := int(float64(gameMap.Width*gameMap.Height) * config.ItemSpawnMaxDensity)
	if len(items) >= maxItems {
		// Still decrement timers but don't spawn
		for _, item := range items {
			item.SpawnTimer -= delta
			if item.SpawnTimer <= 0 {
				item.SpawnTimer = calculateSpawnInterval(initialItemCount)
			}
		}
		return
	}

	// Process each item's spawn timer
	for _, item := range items {
		item.SpawnTimer -= delta

		if item.SpawnTimer <= 0 {
			// Reset timer regardless of spawn success
			item.SpawnTimer = calculateSpawnInterval(initialItemCount)

			// Roll for spawn chance
			if rand.Float64() >= config.ItemSpawnChance {
				continue
			}

			// Try to find empty adjacent tile
			ix, iy := item.Position()
			adjX, adjY, found := findEmptyAdjacent(ix, iy, gameMap)
			if !found {
				continue
			}

			// Spawn new item matching parent's properties (type, color, pattern, texture, etc.)
			spawnItem(gameMap, item, adjX, adjY, initialItemCount)
		}
	}
}

// calculateSpawnInterval returns a randomized spawn interval
func calculateSpawnInterval(initialItemCount int) float64 {
	base := config.ItemSpawnIntervalBase * float64(initialItemCount)
	variance := base * config.ItemSpawnIntervalVariance
	// Random value in range [base - variance, base + variance]
	return base + (rand.Float64()*2-1)*variance
}

// findEmptyAdjacent finds a random empty tile adjacent to (x, y) using 8-directional adjacency
// Returns the coordinates and true if found, or (0, 0, false) if no empty adjacent tile exists
func findEmptyAdjacent(x, y int, gameMap *game.Map) (int, int, bool) {
	// 8 directions: N, NE, E, SE, S, SW, W, NW
	directions := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}

	// Shuffle directions for randomness
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]
		if gameMap.IsValid(nx, ny) && gameMap.IsEmpty(nx, ny) {
			return nx, ny, true
		}
	}

	return 0, 0, false
}

// spawnItem creates a new item matching the parent's properties
func spawnItem(gameMap *game.Map, parent *entity.Item, x, y int, initialItemCount int) {
	var newItem *entity.Item

	switch parent.ItemType {
	case "berry":
		newItem = entity.NewBerry(x, y, parent.Color, parent.Poisonous, parent.Healing)
	case "mushroom":
		newItem = entity.NewMushroom(x, y, parent.Color, parent.Pattern, parent.Texture, parent.Poisonous, parent.Healing)
	case "flower":
		newItem = entity.NewFlower(x, y, parent.Color)
	default:
		return
	}

	// Set spawn timer for the new item (full interval, not staggered)
	newItem.SpawnTimer = calculateSpawnInterval(initialItemCount)
	gameMap.AddItem(newItem)
}
