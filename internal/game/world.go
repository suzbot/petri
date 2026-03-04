package game

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
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
			if cfg.NonPlantSpawned {
				continue // spawned by ground spawning system
			}
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
	case "grass":
		return entity.NewGrass(x, y)
	case "shell":
		return entity.NewShell(x, y, v.Color)
	default:
		// Fallback for unknown types
		return entity.NewFlower(x, y, v.Color)
	}
}

// SpawnFeatures populates the map with landscape features (leaf piles) and water (springs)
func SpawnFeatures(m *Map, noWater, noBeds bool) {
	// Spawn springs as water terrain (drink sources)
	if !noWater {
		for i := 0; i < config.SpringCount; i++ {
			x, y := findEmptySpot(m)
			m.AddWater(types.Position{X: x, Y: y}, WaterSpring)
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

// SpawnPonds generates 1-5 ponds of 4-16 contiguous water tiles each.
// Retries if the resulting map is not fully connected (max 10 attempts).
func SpawnPonds(m *Map) {
	maxRetries := 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		pondCount := config.PondMinCount + rand.Intn(config.PondMaxCount-config.PondMinCount+1)

		for i := 0; i < pondCount; i++ {
			pondSize := config.PondMinSize + rand.Intn(config.PondMaxSize-config.PondMinSize+1)
			spawnPondBlob(m, pondSize)
		}

		if isMapConnected(m) {
			return
		}

		// Clear all pond tiles and retry
		clearPondTiles(m)
	}
}

// SpawnClay generates clay tiles adjacent to water, with loose clay items on some tiles.
// Must be called after SpawnPonds so water tiles exist.
// Uses pair-placement: picks random water-adjacent candidates and pairs each with a cardinal
// neighbor also in the candidate pool. Pairs are scattered independently — clay is a deposit,
// not a contiguous body like a pond.
func SpawnClay(m *Map) {
	waterPositions := m.WaterPositions()
	if len(waterPositions) == 0 {
		return // No water — no clay
	}

	targetSize := config.ClayMinCount + rand.Intn(config.ClayMaxCount-config.ClayMinCount+1)
	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	// Build candidate pool: all non-water tiles cardinal-adjacent to water
	candidateSet := make(map[types.Position]bool)
	for _, wp := range waterPositions {
		for _, dir := range cardinalDirs {
			adj := types.Position{X: wp.X + dir[0], Y: wp.Y + dir[1]}
			if m.IsValid(adj) && !m.IsWater(adj) {
				candidateSet[adj] = true
			}
		}
	}

	// Convert to shuffled slice for random iteration
	candidates := make([]types.Position, 0, len(candidateSet))
	for pos := range candidateSet {
		candidates = append(candidates, pos)
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// Phase 1: Place pairs until we reach target size
	placed := make(map[types.Position]bool)
	var placedList []types.Position

	for _, pos := range candidates {
		if len(placedList) >= targetSize {
			break
		}
		if placed[pos] {
			continue
		}

		// Find a cardinal neighbor that's also in the candidate pool and not yet placed
		var partner types.Position
		foundPartner := false
		// Shuffle directions for variety
		dirs := make([][2]int, len(cardinalDirs))
		copy(dirs, cardinalDirs)
		rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
		for _, dir := range dirs {
			neighbor := types.Position{X: pos.X + dir[0], Y: pos.Y + dir[1]}
			if candidateSet[neighbor] && !placed[neighbor] {
				partner = neighbor
				foundPartner = true
				break
			}
		}
		if !foundPartner {
			continue // Skip candidates with no available partner
		}

		// Place both
		placed[pos] = true
		placed[partner] = true
		placedList = append(placedList, pos, partner)
	}

	// Phase 2: Fill in — add remaining candidates that touch both water and placed clay.
	// Repeats until target reached or no more candidates can be added (each added tile
	// may enable further candidates in the next pass).
	for len(placedList) < targetSize {
		added := false
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		for _, pos := range candidates {
			if len(placedList) >= targetSize {
				break
			}
			if placed[pos] {
				continue
			}
			for _, dir := range cardinalDirs {
				neighbor := types.Position{X: pos.X + dir[0], Y: pos.Y + dir[1]}
				if placed[neighbor] {
					placed[pos] = true
					placedList = append(placedList, pos)
					added = true
					break
				}
			}
		}
		if !added {
			break
		}
	}

	// Set clay terrain
	for _, pos := range placedList {
		m.SetClay(pos)
	}

	// Spawn loose clay items on randomly selected clay tiles
	if len(placedList) > 0 {
		looseCount := config.ClayLooseItems + rand.Intn(2) // ClayLooseItems to ClayLooseItems+1
		shuffled := make([]types.Position, len(placedList))
		copy(shuffled, placedList)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		for i := 0; i < looseCount && i < len(shuffled); i++ {
			m.AddItem(entity.NewClay(shuffled[i].X, shuffled[i].Y))
		}
	}
}

// spawnPondBlob grows a single contiguous pond of the given size from a random starting tile.
func spawnPondBlob(m *Map, size int) {
	// Pick a random starting position that's not already water
	var startX, startY int
	for {
		startX = rand.Intn(m.Width)
		startY = rand.Intn(m.Height)
		pos := types.Position{X: startX, Y: startY}
		if !m.IsWater(pos) {
			break
		}
	}

	start := types.Position{X: startX, Y: startY}
	m.AddWater(start, WaterPond)
	blob := []types.Position{start}

	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	for len(blob) < size {
		// Pick a random tile already in the blob
		source := blob[rand.Intn(len(blob))]

		// Collect valid cardinal neighbors
		var candidates []types.Position
		for _, dir := range cardinalDirs {
			neighbor := types.Position{X: source.X + dir[0], Y: source.Y + dir[1]}
			if m.IsValid(neighbor) && !m.IsWater(neighbor) {
				candidates = append(candidates, neighbor)
			}
		}

		if len(candidates) == 0 {
			// This tile is fully surrounded — try a different blob tile
			// If no tile in the blob can grow, stop early
			canGrow := false
			for _, tile := range blob {
				for _, dir := range cardinalDirs {
					neighbor := types.Position{X: tile.X + dir[0], Y: tile.Y + dir[1]}
					if m.IsValid(neighbor) && !m.IsWater(neighbor) {
						canGrow = true
						break
					}
				}
				if canGrow {
					break
				}
			}
			if !canGrow {
				break
			}
			continue
		}

		chosen := candidates[rand.Intn(len(candidates))]
		m.AddWater(chosen, WaterPond)
		blob = append(blob, chosen)
	}
}

// clearPondTiles removes all pond water tiles from the map (preserves springs).
func clearPondTiles(m *Map) {
	for _, pos := range m.WaterPositions() {
		if m.WaterAt(pos) == WaterPond {
			m.RemoveWater(pos)
		}
	}
}

// isMapConnected returns true if all non-blocked tiles on the map are reachable from each other.
// Uses BFS from the first walkable tile and verifies all walkable tiles are reached.
func isMapConnected(m *Map) bool {
	// Find first walkable tile
	var start types.Position
	found := false
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			pos := types.Position{X: x, Y: y}
			if !m.IsWater(pos) {
				start = pos
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return true // entirely water — vacuously connected
	}

	// BFS from start
	visited := make(map[types.Position]bool)
	queue := []types.Position{start}
	visited[start] = true

	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, dir := range cardinalDirs {
			neighbor := types.Position{X: cur.X + dir[0], Y: cur.Y + dir[1]}
			if m.IsValid(neighbor) && !visited[neighbor] && !m.IsWater(neighbor) {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// Verify all non-water tiles were reached
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			pos := types.Position{X: x, Y: y}
			if !m.IsWater(pos) && !visited[pos] {
				return false
			}
		}
	}
	return true
}

// SpawnGroundItems places initial sticks, nuts, and shells on the map.
// Sticks and nuts go on random empty tiles; shells go adjacent to pond tiles.
func SpawnGroundItems(m *Map) {
	// Spawn sticks on random empty tiles
	for i := 0; i < config.GetGroundSpawnCount("stick"); i++ {
		x, y := findEmptySpot(m)
		m.AddItem(entity.NewStick(x, y))
	}

	// Spawn nuts on random empty tiles
	for i := 0; i < config.GetGroundSpawnCount("nut"); i++ {
		x, y := findEmptySpot(m)
		m.AddItem(entity.NewNut(x, y))
	}

	// Spawn shells adjacent to pond tiles
	pondAdjacentTiles := FindPondAdjacentEmptyTiles(m)
	shellColors := types.ShellColors
	for i := 0; i < config.GetGroundSpawnCount("shell") && len(pondAdjacentTiles) > 0; i++ {
		// Pick a random pond-adjacent tile
		idx := rand.Intn(len(pondAdjacentTiles))
		pos := pondAdjacentTiles[idx]
		// Remove chosen tile to avoid duplicates
		pondAdjacentTiles = append(pondAdjacentTiles[:idx], pondAdjacentTiles[idx+1:]...)

		color := shellColors[rand.Intn(len(shellColors))]
		m.AddItem(entity.NewShell(pos.X, pos.Y, color))
	}
}

// FindPondAdjacentEmptyTiles returns all empty tiles cardinally adjacent to pond water tiles
func FindPondAdjacentEmptyTiles(m *Map) []types.Position {
	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	seen := make(map[types.Position]bool)
	var result []types.Position

	for _, waterPos := range m.WaterPositions() {
		if m.WaterAt(waterPos) != WaterPond {
			continue
		}
		for _, dir := range cardinalDirs {
			adj := types.Position{X: waterPos.X + dir[0], Y: waterPos.Y + dir[1]}
			if !m.IsValid(adj) || seen[adj] {
				continue
			}
			if m.IsEmpty(adj) {
				seen[adj] = true
				result = append(result, adj)
			}
		}
	}
	return result
}

// findEmptySpot finds a random position on the map with no character, water, or feature
func findEmptySpot(m *Map) (int, int) {
	for {
		x := rand.Intn(m.Width)
		y := rand.Intn(m.Height)
		pos := types.Position{X: x, Y: y}
		if !m.IsOccupied(pos) && !m.IsWater(pos) && m.FeatureAt(pos) == nil {
			return x, y
		}
	}
}
