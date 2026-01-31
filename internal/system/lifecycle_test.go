package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// calculateSpawnInterval
// =============================================================================

func TestCalculateSpawnInterval_ReturnsValueInExpectedRange(t *testing.T) {
	t.Parallel()

	initialItemCount := 40 // typical: 20 berries + 20 mushrooms

	// Run multiple times to check range
	for i := 0; i < 100; i++ {
		interval := CalculateSpawnInterval("berry", initialItemCount)

		base := config.ItemLifecycle["berry"].SpawnInterval * float64(initialItemCount)
		variance := base * config.LifecycleIntervalVariance
		minExpected := base - variance
		maxExpected := base + variance

		if interval < minExpected || interval > maxExpected {
			t.Errorf("Interval %.2f outside expected range [%.2f, %.2f]", interval, minExpected, maxExpected)
		}
	}
}

// =============================================================================
// findEmptyAdjacent
// =============================================================================

func TestFindEmptyAdjacent_FindsEmptyTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Position with all adjacent tiles empty
	x, y, found := findEmptyAdjacent(5, 5, gameMap)

	if !found {
		t.Fatal("Should find empty adjacent tile")
	}

	// Check returned position is adjacent (8-directional)
	dx := types.Abs(x - 5)
	dy := types.Abs(y - 5)
	if dx > 1 || dy > 1 || (dx == 0 && dy == 0) {
		t.Errorf("Position (%d, %d) is not adjacent to (5, 5)", x, y)
	}
}

func TestFindEmptyAdjacent_ReturnsFalseWhenNoEmpty(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Fill all 8 adjacent tiles with items
	offsets := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}
	for _, off := range offsets {
		gameMap.AddItem(entity.NewBerry(5+off[0], 5+off[1], types.ColorRed, false, false))
	}

	_, _, found := findEmptyAdjacent(5, 5, gameMap)

	if found {
		t.Error("Should not find empty adjacent tile when all are occupied")
	}
}

func TestFindEmptyAdjacent_RespectsMapBounds(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Position at corner - only 3 valid adjacent tiles
	x, y, found := findEmptyAdjacent(0, 0, gameMap)

	if !found {
		t.Fatal("Should find empty adjacent tile at corner")
	}

	// Check position is within bounds
	if !gameMap.IsValid(types.Position{X: x, Y: y}) {
		t.Errorf("Position (%d, %d) is outside map bounds", x, y)
	}
}

// =============================================================================
// UpdateSpawnTimers
// =============================================================================

func TestUpdateSpawnTimers_DecrementsTimers(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	item.Plant.SpawnTimer = 100.0
	gameMap.AddItem(item)

	delta := 10.0

	UpdateSpawnTimers(gameMap, 40, delta)

	if item.Plant.SpawnTimer != 90.0 {
		t.Errorf("SpawnTimer: got %.2f, want 90.0", item.Plant.SpawnTimer)
	}
}

func TestUpdateSpawnTimers_ResetsTimerWhenExpired(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	item.Plant.SpawnTimer = 0.5 // Will expire with delta of 1.0
	gameMap.AddItem(item)

	initialItemCount := 40

	UpdateSpawnTimers(gameMap, initialItemCount, 1.0)

	// Timer should have been reset to a new interval
	base := config.ItemLifecycle["berry"].SpawnInterval * float64(initialItemCount)
	variance := base * config.LifecycleIntervalVariance

	if item.Plant.SpawnTimer < base-variance || item.Plant.SpawnTimer > base+variance {
		t.Errorf("SpawnTimer %.2f should be in range [%.2f, %.2f]", item.Plant.SpawnTimer, base-variance, base+variance)
	}
}

func TestUpdateSpawnTimers_RespectsMapCap(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(4, 4) // 16 tiles, cap at 8 items (50%)

	// Fill to capacity
	for i := 0; i < 8; i++ {
		item := entity.NewBerry(i%4, i/4, types.ColorRed, false, false)
		item.Plant.SpawnTimer = 0.1 // Will expire immediately
		gameMap.AddItem(item)
	}

	// Run spawn with delta that would trigger all timers
	UpdateSpawnTimers(gameMap, 40, 1.0)

	// Should still have 8 items (cap enforced)
	if len(gameMap.Items()) != 8 {
		t.Errorf("Item count should remain at cap (8), got %d", len(gameMap.Items()))
	}
}

func TestUpdateSpawnTimers_SpawnsAdjacentToParent(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Create a parent item with expired timer
	parent := entity.NewBerry(5, 5, types.ColorBlue, true, true)
	parent.Plant.SpawnTimer = 0.0 // Expired
	gameMap.AddItem(parent)

	// Run multiple times to increase chance of spawn (50% chance)
	for i := 0; i < 20; i++ {
		parent.Plant.SpawnTimer = 0.0 // Reset to expired each iteration
		UpdateSpawnTimers(gameMap, 40, 1.0)
	}

	// Check if any spawn occurred
	items := gameMap.Items()
	if len(items) <= 1 {
		t.Skip("No spawn occurred in 20 attempts (probabilistic)")
	}

	// Verify spawned item is adjacent to parent
	for _, item := range items {
		if item == parent {
			continue
		}
		ipos := item.Pos()
		dx := types.Abs(ipos.X - 5)
		dy := types.Abs(ipos.Y - 5)
		if dx > 1 || dy > 1 {
			t.Errorf("Spawned item at (%d, %d) is not adjacent to parent at (5, 5)", ipos.X, ipos.Y)
		}
	}
}

// =============================================================================
// spawnItem
// =============================================================================

func TestSpawnItem_InheritsParentProperties(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Create poisonous, healing blue berry parent
	parent := entity.NewBerry(5, 5, types.ColorBlue, true, true)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 6, 5, 40)

	items := gameMap.Items()
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// Find the spawned item (not the parent)
	var spawned *entity.Item
	for _, item := range items {
		if item != parent {
			spawned = item
			break
		}
	}

	if spawned == nil {
		t.Fatal("Could not find spawned item")
	}

	// Verify properties match parent
	if spawned.ItemType != parent.ItemType {
		t.Errorf("ItemType: got %s, want %s", spawned.ItemType, parent.ItemType)
	}
	if spawned.Color != parent.Color {
		t.Errorf("Color: got %s, want %s", spawned.Color, parent.Color)
	}
	if spawned.IsPoisonous() != parent.IsPoisonous() {
		t.Errorf("IsPoisonous: got %v, want %v", spawned.IsPoisonous(), parent.IsPoisonous())
	}
	if spawned.IsHealing() != parent.IsHealing() {
		t.Errorf("IsHealing: got %v, want %v", spawned.IsHealing(), parent.IsHealing())
	}

	// Verify position
	spos := spawned.Pos()
	if spos.X != 6 || spos.Y != 5 {
		t.Errorf("Position: got (%d, %d), want (6, 5)", spos.X, spos.Y)
	}
}

func TestSpawnItem_MushroomInheritsProperties(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Create brown mushroom parent
	parent := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, true)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 4, 5, 40)

	items := gameMap.Items()
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	var spawned *entity.Item
	for _, item := range items {
		if item != parent {
			spawned = item
			break
		}
	}

	if spawned.ItemType != "mushroom" {
		t.Errorf("ItemType: got %s, want mushroom", spawned.ItemType)
	}
	if spawned.Color != types.ColorBrown {
		t.Errorf("Color: got %s, want brown", spawned.Color)
	}
}

func TestSpawnItem_SetsSpawnTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	initialItemCount := 40

	parent := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 6, 5, initialItemCount)

	var spawned *entity.Item
	for _, item := range gameMap.Items() {
		if item != parent {
			spawned = item
			break
		}
	}

	base := config.ItemLifecycle["berry"].SpawnInterval * float64(initialItemCount)
	variance := base * config.LifecycleIntervalVariance

	if spawned.Plant.SpawnTimer < base-variance || spawned.Plant.SpawnTimer > base+variance {
		t.Errorf("SpawnTimer %.2f should be in range [%.2f, %.2f]", spawned.Plant.SpawnTimer, base-variance, base+variance)
	}
}

