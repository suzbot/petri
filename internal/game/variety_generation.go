package game

import (
	"math/rand"
	"sort"
	"unicode"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

// ItemTypeConfig defines which attributes are applicable for each item type
type ItemTypeConfig struct {
	Colors               []types.Color
	Patterns             []types.Pattern // nil if patterns don't apply
	Textures             []types.Texture // nil if textures don't apply
	Kind                 string          // subtype identity for varieties (empty for most types; "tall grass" for grass)
	Edible               bool
	CanBePoisonOrHealing bool // if false, never assigned poison/healing (e.g., gourds)
	Plantable            bool // if true, picked items of this type become plantable
	CanProduceSeeds      bool // if true, this type produces plantable seeds (e.g., gourd → gourd seed)
	Sym                  rune
	SpawnCount           int
	NonPlantSpawned      bool // if true, spawned by ground spawning system, not SpawnItems()
}

// GetItemTypeConfigs returns configuration for all item types
func GetItemTypeConfigs() map[string]ItemTypeConfig {
	return map[string]ItemTypeConfig{
		"berry": {
			Colors:               types.BerryColors,
			Patterns:             nil, // berries don't have patterns
			Textures:             nil, // berries don't have textures
			Edible:               true,
			CanBePoisonOrHealing: true,
			Plantable:            true,
			Sym:                  config.CharBerry,
			SpawnCount:           config.ItemSpawnCount,
		},
		"mushroom": {
			Colors:               types.MushroomColors,
			Patterns:             types.MushroomPatterns,
			Textures:             types.MushroomTextures,
			Edible:               true,
			CanBePoisonOrHealing: true,
			Plantable:            true,
			Sym:                  config.CharMushroom,
			SpawnCount:           config.ItemSpawnCount,
		},
		"flower": {
			Colors:               types.FlowerColors,
			Patterns:             nil,
			Textures:             nil,
			Edible:               false,
			CanBePoisonOrHealing: false,
			Sym:                  config.CharFlower,
			SpawnCount:           config.FlowerSpawnCount,
		},
		"gourd": {
			Colors:               types.GourdColors,
			Patterns:             types.GourdPatterns,
			Textures:             types.GourdTextures,
			Edible:               true,
			CanBePoisonOrHealing: false, // gourds are never poisonous or healing
			CanProduceSeeds:      true,
			Sym:                  config.CharGourd,
			SpawnCount:           config.ItemSpawnCount,
		},
		"grass": {
			Colors:               []types.Color{types.ColorPaleGreen},
			Patterns:             nil,
			Textures:             nil,
			Kind:                 "tall grass",
			Edible:               false,
			CanBePoisonOrHealing: false,
			Sym:                  config.CharGrass,
			SpawnCount:           config.ItemSpawnCount,
		},
		"shell": {
			Colors:               types.ShellColors,
			Patterns:             nil,
			Textures:             nil,
			Edible:               false,
			CanBePoisonOrHealing: false,
			Sym:                  config.CharShell,
			SpawnCount:           config.GetGroundSpawnCount("shell"),
			NonPlantSpawned:      true, // spawned by ground spawning, not SpawnItems
		},
		"nut": {
			Colors:               []types.Color{types.ColorBrown},
			Patterns:             nil,
			Textures:             nil,
			Edible:               true,
			CanBePoisonOrHealing: false, // nuts are never poisonous or healing
			Sym:                  config.CharNut,
			SpawnCount:           config.GetGroundSpawnCount("nut"),
			NonPlantSpawned:      true, // spawned by ground spawning, not SpawnItems
		},
	}
}

// PlantableTypeEntry represents a plantable type for the order UI menu.
type PlantableTypeEntry struct {
	DisplayName string // e.g., "Gourd seeds", "Berries"
	TargetType  string // value stored in order.TargetType: "gourd seed" for seeds, "berry" for direct
}

// GetPlantableTypes returns the list of plantable types derived from item type configs.
// Includes directly plantable types (berry, mushroom) and seed-producing types (gourd → "gourd seed").
func GetPlantableTypes() []PlantableTypeEntry {
	configs := GetItemTypeConfigs()
	var entries []PlantableTypeEntry

	// Collect sorted keys for deterministic ordering
	var keys []string
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, itemType := range keys {
		cfg := configs[itemType]
		if cfg.CanProduceSeeds {
			entries = append(entries, PlantableTypeEntry{
				DisplayName: capitalize(itemType) + " seeds",
				TargetType:  itemType + " seed",
			})
		}
		if cfg.Plantable {
			entries = append(entries, PlantableTypeEntry{
				DisplayName: capitalize(entity.Pluralize(itemType)),
				TargetType:  itemType,
			})
		}
	}

	return entries
}

// GatherableTypeEntry represents a gatherable item type for the order UI menu.
type GatherableTypeEntry struct {
	DisplayName string // e.g., "Nuts", "Sticks"
	TargetType  string // value stored in order.TargetType: "nut", "stick", "shell"
}

// GetGatherableTypes returns the list of gatherable item types currently on the ground.
// Filters to items with Plant == nil && Container == nil && ItemType != "hoe".
// Returns a deduplicated, alphabetically sorted list.
func GetGatherableTypes(items []*entity.Item) []GatherableTypeEntry {
	seen := make(map[string]bool)
	var entries []GatherableTypeEntry

	for _, item := range items {
		if item.Plant != nil {
			continue // growing plant — not gatherable
		}
		if item.Container != nil {
			continue // vessel — not gatherable
		}
		if item.ItemType == "hoe" {
			continue // tool — not gatherable
		}
		if seen[item.ItemType] {
			continue // already added
		}
		seen[item.ItemType] = true
		entries = append(entries, GatherableTypeEntry{
			DisplayName: capitalize(entity.Pluralize(item.ItemType)),
			TargetType:  item.ItemType,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TargetType < entries[j].TargetType
	})

	return entries
}

// GetHarvestableItemTypes returns a sorted, deduplicated list of item type names
// that can be harvested — growing, non-sprout plants currently on the map.
// This is the source of truth for the harvest order target list.
func GetHarvestableItemTypes(items []*entity.Item) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		if item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
			continue
		}
		if seen[item.ItemType] {
			continue
		}
		seen[item.ItemType] = true
		result = append(result, item.ItemType)
	}

	sort.Strings(result)
	return result
}

// capitalize returns s with the first letter uppercased.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
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

	// Generate seed varieties for gourds (one seed variety per gourd variety)
	for _, g := range registry.VarietiesOfType("gourd") {
		seedID := entity.GenerateVarietyID("seed", g.Color, g.Pattern, g.Texture)
		registry.Register(&entity.ItemVariety{
			ID:        seedID,
			ItemType:  "seed",
			Kind:      "gourd seed",
			Color:     g.Color,
			Pattern:   g.Pattern,
			Texture:   g.Texture,
			Plantable: true,
			Sym:       config.CharSeed,
		})
	}

	// Register water variety (liquid type for vessel storage)
	registry.Register(&entity.ItemVariety{
		ID:       "liquid-water",
		ItemType: "liquid",
		Kind:     "water",
		Sym:      0, // never rendered as ground item
	})

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

		var edible *entity.EdibleProperties
		if cfg.Edible {
			edible = &entity.EdibleProperties{}
		}
		variety := &entity.ItemVariety{
			ID:        id,
			ItemType:  itemType,
			Kind:      cfg.Kind,
			Color:     color,
			Pattern:   pattern,
			Texture:   texture,
			Edible:    edible,
			Plantable: cfg.Plantable,
			Sym:       cfg.Sym,
		}
		varieties = append(varieties, variety)
	}

	return varieties
}

// assignPoisonAndHealing randomly assigns poison and healing properties to edible varieties
// Only assigns to varieties whose item type has CanBePoisonOrHealing=true
func assignPoisonAndHealing(registry *VarietyRegistry) {
	configs := GetItemTypeConfigs()

	// Filter to edible varieties that can be poison/healing
	var eligible []*entity.ItemVariety
	for _, v := range registry.EdibleVarieties() {
		if cfg, ok := configs[v.ItemType]; ok && cfg.CanBePoisonOrHealing {
			eligible = append(eligible, v)
		}
	}

	if len(eligible) == 0 {
		return
	}

	// Shuffle to randomize selection
	rand.Shuffle(len(eligible), func(i, j int) {
		eligible[i], eligible[j] = eligible[j], eligible[i]
	})

	// Calculate counts (at least 1 of each if we have enough varieties)
	poisonCount := int(float64(len(eligible)) * config.VarietyPoisonPercent)
	if poisonCount < 1 && len(eligible) >= 2 {
		poisonCount = 1
	}

	healingCount := int(float64(len(eligible)) * config.VarietyHealingPercent)
	if healingCount < 1 && len(eligible) >= 2 {
		healingCount = 1
	}

	// Ensure we don't over-assign (poison + healing can't exceed total)
	if poisonCount+healingCount > len(eligible) {
		// Split evenly
		poisonCount = len(eligible) / 2
		healingCount = len(eligible) - poisonCount
	}

	// Assign poison to first N
	for i := 0; i < poisonCount; i++ {
		eligible[i].Edible.Poisonous = true
	}

	// Assign healing to next N (no overlap with poison)
	for i := poisonCount; i < poisonCount+healingCount; i++ {
		eligible[i].Edible.Healing = true
	}
}
