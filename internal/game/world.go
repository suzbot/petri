package game

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
)

// SpawnItems populates the map with random items using the variety system
func SpawnItems(m *Map, mushroomsOnly bool) {
	// Generate varieties for this world (defines what combos exist, assigns poison/healing)
	registry := GenerateVarieties()
	m.SetVarieties(registry)

	configs := GetItemTypeConfigs()

	// Calculate total spawn count for timer staggering
	totalSpawnCount := 0
	for _, cfg := range configs {
		totalSpawnCount += cfg.SpawnCount
	}
	// Use berry spawn interval as reference (all types currently have same interval)
	maxInitialTimer := config.ItemLifecycle["berry"].SpawnInterval * float64(totalSpawnCount)

	if mushroomsOnly {
		// Replace all items with mushroom varieties for testing preference formation
		spawnItemsOfType(m, registry, "mushroom", totalSpawnCount, maxInitialTimer, totalSpawnCount)
	} else {
		// Spawn items for each type using their configured spawn counts
		for itemType, cfg := range configs {
			spawnItemsOfType(m, registry, itemType, cfg.SpawnCount, maxInitialTimer, totalSpawnCount)
		}
	}
}

// spawnItemsOfType spawns count items of the given type, distributed across varieties
func spawnItemsOfType(m *Map, registry *VarietyRegistry, itemType string, count int, maxInitialTimer float64, totalSpawnCount int) {
	varieties := registry.VarietiesOfType(itemType)
	if len(varieties) == 0 {
		return
	}

	// Calculate max death timer for staggering (if this type has death)
	lifecycleCfg := config.ItemLifecycle[itemType]
	maxDeathTimer := lifecycleCfg.DeathInterval * float64(totalSpawnCount)

	for i := 0; i < count; i++ {
		// Pick a random variety of this type
		v := varieties[rand.Intn(len(varieties))]

		x, y := findEmptySpot(m)
		item := createItemFromVariety(v, x, y)
		// Stagger spawn timers across first cycle (all spawned items are plants)
		if item.Plant != nil {
			item.Plant.SpawnTimer = rand.Float64() * maxInitialTimer
		}

		// Set death timer if this item type is mortal (stagger to avoid synchronized die-off)
		if maxDeathTimer > 0 {
			item.DeathTimer = rand.Float64() * maxDeathTimer
		}

		m.AddItem(item)
	}
}

// createItemFromVariety creates an Item by copying attributes from a variety
func createItemFromVariety(v *entity.ItemVariety, x, y int) *entity.Item {
	switch v.ItemType {
	case "berry":
		return entity.NewBerry(x, y, v.Color, v.IsPoisonous(), v.IsHealing())
	case "mushroom":
		return entity.NewMushroom(x, y, v.Color, v.Pattern, v.Texture, v.IsPoisonous(), v.IsHealing())
	case "gourd":
		return entity.NewGourd(x, y, v.Color, v.Pattern, v.Texture, v.IsPoisonous(), v.IsHealing())
	case "flower":
		return entity.NewFlower(x, y, v.Color)
	default:
		// Fallback for unknown types
		return entity.NewFlower(x, y, v.Color)
	}
}

// SpawnFeatures populates the map with landscape features (springs, leaf piles)
func SpawnFeatures(m *Map, noWater, noBeds bool) {
	// Spawn springs (drink sources)
	if !noWater {
		for i := 0; i < config.SpringCount; i++ {
			x, y := findEmptySpot(m)
			m.AddFeature(entity.NewSpring(x, y))
		}
	}

	// Spawn leaf piles (beds)
	if !noBeds {
		for i := 0; i < config.LeafPileCount; i++ {
			x, y := findEmptySpot(m)
			m.AddFeature(entity.NewLeafPile(x, y))
		}
	}
}

// findEmptySpot finds a random unoccupied position on the map
func findEmptySpot(m *Map) (int, int) {
	for {
		x := rand.Intn(m.Width)
		y := rand.Intn(m.Height)
		if !m.IsOccupied(x, y) {
			return x, y
		}
	}
}
