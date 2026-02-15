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
// FindEmptyAdjacent
// =============================================================================

func TestFindEmptyAdjacent_FindsEmptyTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Position with all adjacent tiles empty
	x, y, found := FindEmptyAdjacent(5, 5, gameMap)

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

	_, _, found := FindEmptyAdjacent(5, 5, gameMap)

	if found {
		t.Error("Should not find empty adjacent tile when all are occupied")
	}
}

func TestFindEmptyAdjacent_RespectsMapBounds(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Position at corner - only 3 valid adjacent tiles
	x, y, found := FindEmptyAdjacent(0, 0, gameMap)

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

func TestSpawnItem_SetsSproutTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	parent := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 6, 5, 40)

	var spawned *entity.Item
	for _, item := range gameMap.Items() {
		if item != parent {
			spawned = item
			break
		}
	}

	if spawned.Plant.SproutTimer != config.SproutDuration {
		t.Errorf("SproutTimer: got %.2f, want %.2f", spawned.Plant.SproutTimer, config.SproutDuration)
	}
}

// =============================================================================
// MatureSymbol
// =============================================================================

func TestMatureSymbol_ReturnsCorrectSymbols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		itemType string
		want     rune
	}{
		{"berry", config.CharBerry},
		{"mushroom", config.CharMushroom},
		{"flower", config.CharFlower},
		{"gourd", config.CharGourd},
	}

	for _, tt := range tests {
		if got := MatureSymbol(tt.itemType); got != tt.want {
			t.Errorf("MatureSymbol(%q): got %c, want %c", tt.itemType, got, tt.want)
		}
	}
}

// =============================================================================
// UpdateSproutTimers
// =============================================================================

func TestUpdateSproutTimers_DecrementsSproutTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 20.0
	gameMap.AddItem(sprout)

	UpdateSproutTimers(gameMap, 40, 5.0)

	if sprout.Plant.SproutTimer != 15.0 {
		t.Errorf("SproutTimer: got %.2f, want 15.0", sprout.Plant.SproutTimer)
	}
}

func TestUpdateSproutTimers_MaturesWhenTimerReachesZero(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 1.0
	gameMap.AddItem(sprout)

	UpdateSproutTimers(gameMap, 40, 2.0)

	if sprout.Plant.IsSprout {
		t.Error("Expected IsSprout=false after maturation")
	}
}

func TestUpdateSproutTimers_MatureSproutHasCorrectSymbol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		itemType string
		newFn    func() *entity.Item
		wantSym  rune
	}{
		{"berry", func() *entity.Item { return entity.NewBerry(5, 5, types.ColorRed, false, false) }, config.CharBerry},
		{"mushroom", func() *entity.Item {
			return entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
		}, config.CharMushroom},
		{"flower", func() *entity.Item { return entity.NewFlower(5, 5, types.ColorBlue) }, config.CharFlower},
		{"gourd", func() *entity.Item {
			return entity.NewGourd(5, 5, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
		}, config.CharGourd},
	}

	for _, tt := range tests {
		gameMap := game.NewMap(10, 10)
		sprout := tt.newFn()
		sprout.Sym = config.CharSprout
		sprout.Plant.IsSprout = true
		sprout.Plant.SproutTimer = 1.0
		gameMap.AddItem(sprout)

		UpdateSproutTimers(gameMap, 40, 2.0)

		if sprout.Sym != tt.wantSym {
			t.Errorf("MatureSymbol(%s): got %c, want %c", tt.itemType, sprout.Sym, tt.wantSym)
		}
	}
}

func TestUpdateSproutTimers_MaturePlantHasSpawnTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 1.0
	sprout.Plant.SpawnTimer = 0 // no spawn timer yet
	gameMap.AddItem(sprout)

	initialItemCount := 40
	UpdateSproutTimers(gameMap, initialItemCount, 2.0)

	base := config.ItemLifecycle["berry"].SpawnInterval * float64(initialItemCount)
	variance := base * config.LifecycleIntervalVariance

	if sprout.Plant.SpawnTimer < base-variance || sprout.Plant.SpawnTimer > base+variance {
		t.Errorf("SpawnTimer %.2f should be in range [%.2f, %.2f]", sprout.Plant.SpawnTimer, base-variance, base+variance)
	}
}

func TestUpdateSproutTimers_MaturePlantHasDeathTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	// Flowers have a death interval
	sprout := entity.NewFlower(5, 5, types.ColorBlue)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 1.0
	gameMap.AddItem(sprout)

	initialItemCount := 40
	UpdateSproutTimers(gameMap, initialItemCount, 2.0)

	base := config.ItemLifecycle["flower"].DeathInterval * float64(initialItemCount)
	variance := base * config.LifecycleIntervalVariance

	if sprout.DeathTimer < base-variance || sprout.DeathTimer > base+variance {
		t.Errorf("DeathTimer %.2f should be in range [%.2f, %.2f]", sprout.DeathTimer, base-variance, base+variance)
	}
}

func TestUpdateSproutTimers_ImmortalPlantHasZeroDeathTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	// Berries are immortal (DeathInterval = 0)
	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 1.0
	gameMap.AddItem(sprout)

	UpdateSproutTimers(gameMap, 40, 2.0)

	if sprout.DeathTimer != 0 {
		t.Errorf("DeathTimer: got %.2f, want 0 (immortal)", sprout.DeathTimer)
	}
}

func TestUpdateSproutTimers_TilledSoilMultiplier(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	pos := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(pos)

	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 20.0
	gameMap.AddItem(sprout)

	delta := 10.0
	UpdateSproutTimers(gameMap, 40, delta)

	expectedTimer := 20.0 - (delta * config.TilledGrowthMultiplier)
	if sprout.Plant.SproutTimer != expectedTimer {
		t.Errorf("SproutTimer on tilled: got %.2f, want %.2f", sprout.Plant.SproutTimer, expectedTimer)
	}
}

func TestUpdateSproutTimers_WetTileMultiplier(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	// Place water adjacent to sprout position so (5,5) is wet
	gameMap.AddWater(types.Position{X: 4, Y: 5}, game.WaterPond)

	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 20.0
	gameMap.AddItem(sprout)

	delta := 10.0
	UpdateSproutTimers(gameMap, 40, delta)

	expectedTimer := 20.0 - (delta * config.WetGrowthMultiplier)
	if sprout.Plant.SproutTimer != expectedTimer {
		t.Errorf("SproutTimer on wet: got %.2f, want %.2f", sprout.Plant.SproutTimer, expectedTimer)
	}
}

func TestUpdateSproutTimers_TilledAndWetStack(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	pos := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(pos)
	gameMap.AddWater(types.Position{X: 4, Y: 5}, game.WaterPond)

	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 20.0
	gameMap.AddItem(sprout)

	delta := 10.0
	UpdateSproutTimers(gameMap, 40, delta)

	expectedTimer := 20.0 - (delta * config.TilledGrowthMultiplier * config.WetGrowthMultiplier)
	if sprout.Plant.SproutTimer != expectedTimer {
		t.Errorf("SproutTimer on tilled+wet: got %.2f, want %.2f", sprout.Plant.SproutTimer, expectedTimer)
	}
}

func TestUpdateSproutTimers_SkipsNonSprouts(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	// Mature plant, not a sprout
	mature := entity.NewBerry(5, 5, types.ColorRed, false, false)
	mature.Plant.SpawnTimer = 100.0
	gameMap.AddItem(mature)

	UpdateSproutTimers(gameMap, 40, 5.0)

	// SpawnTimer should be unchanged (UpdateSproutTimers doesn't touch non-sprouts)
	if mature.Plant.SpawnTimer != 100.0 {
		t.Errorf("SpawnTimer: got %.2f, want 100.0 (should be unchanged)", mature.Plant.SpawnTimer)
	}
}

// =============================================================================
// spawnItem creates sprouts
// =============================================================================

func TestSpawnItem_CreatesSproutInsteadOfMaturePlant(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	parent := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 6, 5, 40)

	var spawned *entity.Item
	for _, item := range gameMap.Items() {
		if item != parent {
			spawned = item
			break
		}
	}

	if spawned == nil {
		t.Fatal("Could not find spawned item")
	}

	if !spawned.Plant.IsSprout {
		t.Error("Spawned item should be a sprout (IsSprout=true)")
	}
	if spawned.Sym != config.CharSprout {
		t.Errorf("Spawned sprout symbol: got %c, want %c", spawned.Sym, config.CharSprout)
	}
	if spawned.Plant.SproutTimer != config.SproutDuration {
		t.Errorf("SproutTimer: got %.2f, want %.2f", spawned.Plant.SproutTimer, config.SproutDuration)
	}
}

func TestSpawnItem_SproutHasParentVarietyAndEdible(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	parent := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternSpotted, types.TextureWarty, true, true)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 6, 5, 40)

	var spawned *entity.Item
	for _, item := range gameMap.Items() {
		if item != parent {
			spawned = item
			break
		}
	}

	if spawned.Color != types.ColorBrown {
		t.Errorf("Color: got %s, want brown", spawned.Color)
	}
	if spawned.Pattern != types.PatternSpotted {
		t.Errorf("Pattern: got %s, want spotted", spawned.Pattern)
	}
	if spawned.Texture != types.TextureWarty {
		t.Errorf("Texture: got %s, want warty", spawned.Texture)
	}
	if !spawned.IsPoisonous() {
		t.Error("Spawned sprout should be poisonous like parent")
	}
	if !spawned.IsHealing() {
		t.Error("Spawned sprout should be healing like parent")
	}
}

func TestSpawnItem_SproutHasNoSpawnOrDeathTimer(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	parent := entity.NewFlower(5, 5, types.ColorBlue)
	gameMap.AddItem(parent)

	spawnItem(gameMap, parent, 6, 5, 40)

	var spawned *entity.Item
	for _, item := range gameMap.Items() {
		if item != parent {
			spawned = item
			break
		}
	}

	if spawned.Plant.SpawnTimer != 0 {
		t.Errorf("Sprout SpawnTimer: got %.2f, want 0 (set on maturation)", spawned.Plant.SpawnTimer)
	}
	if spawned.DeathTimer != 0 {
		t.Errorf("Sprout DeathTimer: got %.2f, want 0 (set on maturation)", spawned.DeathTimer)
	}
}

// =============================================================================
// UpdateSpawnTimers skips sprouts + growth multipliers
// =============================================================================

func TestUpdateSpawnTimers_SkipsSproutsForReproduction(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	sprout := entity.NewBerry(5, 5, types.ColorRed, false, false)
	sprout.Sym = config.CharSprout
	sprout.Plant.IsSprout = true
	sprout.Plant.SproutTimer = 20.0
	sprout.Plant.SpawnTimer = 10.0 // shouldn't be decremented
	gameMap.AddItem(sprout)

	UpdateSpawnTimers(gameMap, 40, 5.0)

	if sprout.Plant.SpawnTimer != 10.0 {
		t.Errorf("SpawnTimer: got %.2f, want 10.0 (sprouts should be skipped)", sprout.Plant.SpawnTimer)
	}
}

func TestUpdateSpawnTimers_TilledMultiplierOnReproduction(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	pos := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(pos)

	plant := entity.NewBerry(5, 5, types.ColorRed, false, false)
	plant.Plant.SpawnTimer = 100.0
	gameMap.AddItem(plant)

	delta := 10.0
	UpdateSpawnTimers(gameMap, 40, delta)

	expectedTimer := 100.0 - (delta * config.TilledGrowthMultiplier)
	if plant.Plant.SpawnTimer != expectedTimer {
		t.Errorf("SpawnTimer on tilled: got %.2f, want %.2f", plant.Plant.SpawnTimer, expectedTimer)
	}
}

func TestUpdateSpawnTimers_WetMultiplierOnReproduction(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	gameMap.AddWater(types.Position{X: 4, Y: 5}, game.WaterPond)

	plant := entity.NewBerry(5, 5, types.ColorRed, false, false)
	plant.Plant.SpawnTimer = 100.0
	gameMap.AddItem(plant)

	delta := 10.0
	UpdateSpawnTimers(gameMap, 40, delta)

	expectedTimer := 100.0 - (delta * config.WetGrowthMultiplier)
	if plant.Plant.SpawnTimer != expectedTimer {
		t.Errorf("SpawnTimer on wet: got %.2f, want %.2f", plant.Plant.SpawnTimer, expectedTimer)
	}
}

