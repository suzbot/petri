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

	// Calculate initial spawn cycle for staggered timers
	initialItemCount := config.ItemSpawnCount*2 + config.FlowerSpawnCount // berries + mushrooms + flowers
	maxInitialTimer := config.ItemSpawnIntervalBase * float64(initialItemCount)

	if mushroomsOnly {
		// Replace all items with mushroom varieties for testing preference formation
		totalCount := config.ItemSpawnCount*2 + config.FlowerSpawnCount
		spawnItemsOfType(m, registry, "mushroom", totalCount, maxInitialTimer)
	} else {
		// Spawn items for each item type
		spawnItemsOfType(m, registry, "berry", config.ItemSpawnCount, maxInitialTimer)
		spawnItemsOfType(m, registry, "mushroom", config.ItemSpawnCount, maxInitialTimer)
		spawnItemsOfType(m, registry, "flower", config.FlowerSpawnCount, maxInitialTimer)
	}
}

// spawnItemsOfType spawns count items of the given type, distributed across varieties
func spawnItemsOfType(m *Map, registry *VarietyRegistry, itemType string, count int, maxInitialTimer float64) {
	varieties := registry.VarietiesOfType(itemType)
	if len(varieties) == 0 {
		return
	}

	for i := 0; i < count; i++ {
		// Pick a random variety of this type
		v := varieties[rand.Intn(len(varieties))]

		x, y := findEmptySpot(m)
		item := createItemFromVariety(v, x, y)
		item.SpawnTimer = rand.Float64() * maxInitialTimer // stagger across first cycle
		m.AddItem(item)
	}
}

// createItemFromVariety creates an Item by copying attributes from a variety
func createItemFromVariety(v *entity.ItemVariety, x, y int) *entity.Item {
	switch v.ItemType {
	case "berry":
		return entity.NewBerry(x, y, v.Color, v.Poisonous, v.Healing)
	case "mushroom":
		return entity.NewMushroom(x, y, v.Color, v.Pattern, v.Texture, v.Poisonous, v.Healing)
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
