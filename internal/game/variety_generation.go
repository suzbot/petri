package game

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

// ItemTypeConfig defines which attributes are applicable for each item type
type ItemTypeConfig struct {
	Colors     []types.Color
	Patterns   []types.Pattern // nil if patterns don't apply
	Textures   []types.Texture // nil if textures don't apply
	Edible     bool
	Sym        rune
	SpawnCount int
}

// GetItemTypeConfigs returns configuration for all item types
func GetItemTypeConfigs() map[string]ItemTypeConfig {
	return map[string]ItemTypeConfig{
		"berry": {
			Colors:     types.BerryColors,
			Patterns:   nil, // berries don't have patterns
			Textures:   nil, // berries don't have textures
			Edible:     true,
			Sym:        config.CharBerry,
			SpawnCount: config.ItemSpawnCount,
		},
		"mushroom": {
			Colors:     types.MushroomColors,
			Patterns:   types.MushroomPatterns,
			Textures:   types.MushroomTextures,
			Edible:     true,
			Sym:        config.CharMushroom,
			SpawnCount: config.ItemSpawnCount,
		},
		"flower": {
			Colors:     types.FlowerColors,
			Patterns:   nil,
			Textures:   nil,
			Edible:     false,
			Sym:        config.CharFlower,
			SpawnCount: config.FlowerSpawnCount,
		},
	}
}

// GenerateVarieties creates all item varieties for a new world
func GenerateVarieties() *VarietyRegistry {
	registry := NewVarietyRegistry()
	configs := GetItemTypeConfigs()

	for itemType, cfg := range configs {
		varieties := generateVarietiesForType(itemType, cfg)
		for _, v := range varieties {
			registry.Register(v)
		}
	}

	// Assign poison and healing to edible varieties
	assignPoisonAndHealing(registry)

	return registry
}

// generateVarietiesForType creates varieties for a single item type
func generateVarietiesForType(itemType string, cfg ItemTypeConfig) []*entity.ItemVariety {
	// Calculate target variety count
	targetCount := cfg.SpawnCount / config.VarietyDivisor
	if targetCount < config.VarietyMinCount {
		targetCount = config.VarietyMinCount
	}

	// Generate unique combinations
	seen := make(map[string]bool)
	var varieties []*entity.ItemVariety

	// Try to generate targetCount unique varieties
	// Use a max attempts limit to avoid infinite loops if attribute space is small
	maxAttempts := targetCount * 10
	attempts := 0

	for len(varieties) < targetCount && attempts < maxAttempts {
		attempts++

		// Pick random attributes
		color := cfg.Colors[rand.Intn(len(cfg.Colors))]

		var pattern types.Pattern
		if cfg.Patterns != nil {
			pattern = cfg.Patterns[rand.Intn(len(cfg.Patterns))]
		}

		var texture types.Texture
		if cfg.Textures != nil {
			texture = cfg.Textures[rand.Intn(len(cfg.Textures))]
		}

		// Check for duplicate
		id := entity.GenerateVarietyID(itemType, color, pattern, texture)
		if seen[id] {
			continue
		}
		seen[id] = true

		variety := &entity.ItemVariety{
			ID:       id,
			ItemType: itemType,
			Color:    color,
			Pattern:  pattern,
			Texture:  texture,
			Edible:   cfg.Edible,
			Sym:      cfg.Sym,
		}
		varieties = append(varieties, variety)
	}

	return varieties
}

// assignPoisonAndHealing randomly assigns poison and healing properties to edible varieties
func assignPoisonAndHealing(registry *VarietyRegistry) {
	edible := registry.EdibleVarieties()
	if len(edible) == 0 {
		return
	}

	// Shuffle to randomize selection
	rand.Shuffle(len(edible), func(i, j int) {
		edible[i], edible[j] = edible[j], edible[i]
	})

	// Calculate counts (at least 1 of each if we have enough varieties)
	poisonCount := int(float64(len(edible)) * config.VarietyPoisonPercent)
	if poisonCount < 1 && len(edible) >= 2 {
		poisonCount = 1
	}

	healingCount := int(float64(len(edible)) * config.VarietyHealingPercent)
	if healingCount < 1 && len(edible) >= 2 {
		healingCount = 1
	}

	// Ensure we don't over-assign (poison + healing can't exceed total)
	if poisonCount+healingCount > len(edible) {
		// Split evenly
		poisonCount = len(edible) / 2
		healingCount = len(edible) - poisonCount
	}

	// Assign poison to first N
	for i := 0; i < poisonCount; i++ {
		edible[i].Poisonous = true
	}

	// Assign healing to next N (no overlap with poison)
	for i := poisonCount; i < poisonCount+healingCount; i++ {
		edible[i].Healing = true
	}
}
