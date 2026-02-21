package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// MatureSymbol returns the correct symbol for a mature plant of the given item type
func MatureSymbol(itemType string) rune {
	switch itemType {
	case "berry":
		return config.CharBerry
	case "mushroom":
		return config.CharMushroom
	case "flower":
		return config.CharFlower
	case "gourd":
		return config.CharGourd
	default:
		return config.CharSprout // fallback
	}
}

// effectiveDelta calculates the growth-adjusted delta for an item position,
// applying tilled and wet multipliers when applicable
func effectiveDelta(delta float64, pos types.Position, gameMap *game.Map) float64 {
	d := delta
	if gameMap.IsTilled(pos) {
		d *= config.TilledGrowthMultiplier
	}
	if gameMap.IsWet(pos) {
		d *= config.WetGrowthMultiplier
	}
	return d
}

// UpdateSproutTimers decrements sprout timers and matures sprouts when timers expire
func UpdateSproutTimers(gameMap *game.Map, initialItemCount int, delta float64) {
	for _, item := range gameMap.Items() {
		if item.Plant == nil || !item.Plant.IsSprout {
			continue
		}

		pos := item.Pos()
		item.Plant.SproutTimer -= effectiveDelta(delta, pos, gameMap)

		if item.Plant.SproutTimer <= 0 {
			// Mature the sprout
			item.Plant.IsSprout = false
			item.Plant.SproutTimer = 0
			item.Sym = MatureSymbol(item.ItemType)
			item.Plant.SpawnTimer = CalculateSpawnInterval(item.ItemType, initialItemCount)
			item.DeathTimer = CalculateDeathInterval(item.ItemType, initialItemCount)

			// Restore variety attributes from registry (sprouts from seeds may have nil Edible)
			if registry := gameMap.Varieties(); registry != nil {
				variety := registry.GetByAttributes(item.ItemType, item.Color, item.Pattern, item.Texture)
				if variety != nil && variety.Edible != nil {
					item.Edible = &entity.EdibleProperties{
						Poisonous: variety.Edible.Poisonous,
						Healing:   variety.Edible.Healing,
					}
				}
			}
		}
	}
}

// UpdateSpawnTimers decrements spawn timers for growing plants and spawns new items when timers expire
func UpdateSpawnTimers(gameMap *game.Map, initialItemCount int, delta float64) {
	items := gameMap.Items()

	// Check if map is at capacity (50% of coordinates)
	maxItems := int(float64(gameMap.Width*gameMap.Height) * config.ItemSpawnMaxDensity)
	if len(items) >= maxItems {
		// Still decrement timers but don't spawn
		for _, item := range items {
			// Only process growing plants (skip sprouts — they can't reproduce yet)
			if item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
				continue
			}
			pos := item.Pos()
			item.Plant.SpawnTimer -= effectiveDelta(delta, pos, gameMap)
			if item.Plant.SpawnTimer <= 0 {
				item.Plant.SpawnTimer = CalculateSpawnInterval(item.ItemType, initialItemCount)
			}
		}
		return
	}

	// Process each growing plant's spawn timer
	for _, item := range items {
		// Only process growing plants (skip sprouts — they can't reproduce yet)
		if item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
			continue
		}

		pos := item.Pos()
		item.Plant.SpawnTimer -= effectiveDelta(delta, pos, gameMap)

		if item.Plant.SpawnTimer <= 0 {
			// Reset timer regardless of spawn success
			item.Plant.SpawnTimer = CalculateSpawnInterval(item.ItemType, initialItemCount)

			// Roll for spawn chance
			if rand.Float64() >= config.ItemSpawnChance {
				continue
			}

			// Try to find empty adjacent tile
			ipos := item.Pos()
			adjX, adjY, found := FindEmptyAdjacent(ipos.X, ipos.Y, gameMap)
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

// FindEmptyAdjacent finds a random empty adjacent tile (8-directional).
// Returns the coordinates and true if found, or (0, 0, false) if no empty adjacent tile exists.
func FindEmptyAdjacent(x, y int, gameMap *game.Map) (int, int, bool) {
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
		adjPos := types.Position{X: nx, Y: ny}
		if gameMap.IsValid(adjPos) && gameMap.IsEmpty(adjPos) {
			return nx, ny, true
		}
	}

	return 0, 0, false
}

// spawnItem creates a new sprout matching the parent's properties.
// Offspring start as sprouts and mature via UpdateSproutTimers.
func spawnItem(gameMap *game.Map, parent *entity.Item, x, y int, initialItemCount int) {
	// Copy edible properties if parent is edible
	var edible *entity.EdibleProperties
	if parent.Edible != nil {
		edible = &entity.EdibleProperties{
			Poisonous: parent.Edible.Poisonous,
			Healing:   parent.Edible.Healing,
		}
	}

	newItem := &entity.Item{
		BaseEntity: entity.BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharSprout,
			EType: entity.TypeItem,
		},
		ItemType: parent.ItemType,
		Color:    parent.Color,
		Pattern:  parent.Pattern,
		Texture:  parent.Texture,
		Plant: &entity.PlantProperties{
			IsGrowing:   true,
			IsSprout:    true,
			SproutTimer: config.SproutDuration,
		},
		Edible: edible,
	}

	// No SpawnTimer or DeathTimer — set on maturation by UpdateSproutTimers

	gameMap.AddItem(newItem)
}
