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
				item.SpawnTimer = CalculateSpawnInterval(item.ItemType, initialItemCount)
			}
		}
		return
	}

	// Process each item's spawn timer
	for _, item := range items {
		item.SpawnTimer -= delta

		if item.SpawnTimer <= 0 {
			// Reset timer regardless of spawn success
			item.SpawnTimer = CalculateSpawnInterval(item.ItemType, initialItemCount)

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

// UpdateDeathTimers decrements death timers for mortal items and removes them when expired
func UpdateDeathTimers(gameMap *game.Map, delta float64) {
	items := gameMap.Items()

	// Collect items to remove (can't modify slice while iterating)
	var toRemove []*entity.Item

	for _, item := range items {
		// Check if this item type has a death interval
		cfg, ok := config.ItemLifecycle[item.ItemType]
		if !ok || cfg.DeathInterval <= 0 {
			continue // immortal
		}

		// Skip items without death timer set (shouldn't happen, but defensive)
		if item.DeathTimer <= 0 {
			continue
		}

		item.DeathTimer -= delta

		if item.DeathTimer <= 0 {
			toRemove = append(toRemove, item)
		}
	}

	// Remove dead items
	for _, item := range toRemove {
		gameMap.RemoveItem(item)
	}
}

// CalculateSpawnInterval returns a randomized spawn interval for an item type
func CalculateSpawnInterval(itemType string, initialItemCount int) float64 {
	cfg, ok := config.ItemLifecycle[itemType]
	if !ok {
		// Fallback for unknown types
		cfg = config.LifecycleConfig{SpawnInterval: 3.0}
	}

	base := cfg.SpawnInterval * float64(initialItemCount)
	variance := base * config.LifecycleIntervalVariance
	// Random value in range [base - variance, base + variance]
	return base + (rand.Float64()*2-1)*variance
}

// CalculateDeathInterval returns a randomized death interval for an item type
// Returns 0 if the item type is immortal
func CalculateDeathInterval(itemType string, initialItemCount int) float64 {
	cfg, ok := config.ItemLifecycle[itemType]
	if !ok || cfg.DeathInterval <= 0 {
		return 0 // immortal
	}

	base := cfg.DeathInterval * float64(initialItemCount)
	variance := base * config.LifecycleIntervalVariance
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
// Uses generic copying so new plant types don't require code changes here
func spawnItem(gameMap *game.Map, parent *entity.Item, x, y int, initialItemCount int) {
	newItem := &entity.Item{
		BaseEntity: entity.BaseEntity{
			X:     x,
			Y:     y,
			Sym:   parent.Sym,
			EType: entity.TypeItem,
		},
		ItemType:  parent.ItemType,
		Color:     parent.Color,
		Pattern:   parent.Pattern,
		Texture:   parent.Texture,
		Category:  parent.Category,
		Edible:    parent.Edible,
		Poisonous: parent.Poisonous,
		Healing:   parent.Healing,
	}

	// Set lifecycle timers for the new item
	newItem.SpawnTimer = CalculateSpawnInterval(newItem.ItemType, initialItemCount)
	newItem.DeathTimer = CalculateDeathInterval(newItem.ItemType, initialItemCount)

	gameMap.AddItem(newItem)
}
