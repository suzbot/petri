package game

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

// PoisonConfig tracks which food type/color combos are poisonous
type PoisonConfig map[string]bool

// HealingConfig tracks which food type/color combos are healing
type HealingConfig map[string]bool

// GeneratePoisonConfig creates a random poison configuration for a world
func GeneratePoisonConfig() PoisonConfig {
	cfg := make(PoisonConfig)
	allCombos := config.AllPoisonCombos()

	// Pick 1-3 random combinations to be poisonous
	numPoisonous := rand.Intn(3) + 1
	perm := rand.Perm(len(allCombos))

	for i := 0; i < numPoisonous && i < len(perm); i++ {
		combo := allCombos[perm[i]]
		key := combo.ItemType + ":" + string(combo.Color)
		cfg[key] = true
	}

	return cfg
}

// IsPoisonous checks if a given item type/color combo is poisonous
func (pc PoisonConfig) IsPoisonous(itemType string, color types.Color) bool {
	return pc[itemType+":"+string(color)]
}

// IsHealing checks if a given item type/color combo is healing
func (hc HealingConfig) IsHealing(itemType string, color types.Color) bool {
	return hc[itemType+":"+string(color)]
}

// GenerateHealingConfig creates a random healing configuration for a world
// Healing combos must not overlap with poison combos
func GenerateHealingConfig(poisonCfg PoisonConfig) HealingConfig {
	cfg := make(HealingConfig)
	allCombos := config.AllHealingCombos()

	// Filter out combos that are already poisonous
	var available []config.HealingCombo
	for _, combo := range allCombos {
		key := combo.ItemType + ":" + string(combo.Color)
		if !poisonCfg[key] {
			available = append(available, combo)
		}
	}

	// Pick 1-2 random non-poisonous combinations to be healing
	if len(available) == 0 {
		return cfg
	}

	numHealing := rand.Intn(2) + 1 // 1-2 healing combos
	if numHealing > len(available) {
		numHealing = len(available)
	}

	perm := rand.Perm(len(available))
	for i := 0; i < numHealing; i++ {
		combo := available[perm[i]]
		key := combo.ItemType + ":" + string(combo.Color)
		cfg[key] = true
	}

	return cfg
}

// SpawnItems populates the map with random items
func SpawnItems(m *Map, poisonCfg PoisonConfig, healingCfg HealingConfig) {
	// Spawn berries
	for i := 0; i < config.ItemSpawnCount; i++ {
		x, y := findEmptySpot(m)
		color := types.BerryColors[rand.Intn(len(types.BerryColors))]
		poisonous := poisonCfg.IsPoisonous("berry", color)
		healing := healingCfg.IsHealing("berry", color)
		m.AddItem(entity.NewBerry(x, y, color, poisonous, healing))
	}

	// Spawn mushrooms
	for i := 0; i < config.ItemSpawnCount; i++ {
		x, y := findEmptySpot(m)
		color := types.MushroomColors[rand.Intn(len(types.MushroomColors))]
		poisonous := poisonCfg.IsPoisonous("mushroom", color)
		healing := healingCfg.IsHealing("mushroom", color)
		m.AddItem(entity.NewMushroom(x, y, color, poisonous, healing))
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
