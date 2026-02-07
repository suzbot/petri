package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// Timer behavior
// =============================================================================

func TestUpdateGroundSpawning_DoesNotFireBeforeInterval(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	timers := &GroundSpawnTimers{
		Stick: config.GroundSpawnInterval,
		Nut:   config.GroundSpawnInterval,
		Shell: config.GroundSpawnInterval,
	}

	// Advance by half the interval — nothing should spawn
	UpdateGroundSpawning(gameMap, config.GroundSpawnInterval/2, timers)

	if len(gameMap.Items()) != 0 {
		t.Errorf("Expected no items before interval elapses, got %d", len(gameMap.Items()))
	}
}

func TestUpdateGroundSpawning_FiresAfterInterval(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	// Add a pond so shells can spawn too
	gameMap.AddWater(types.Position{X: 10, Y: 10}, game.WaterPond)

	timers := &GroundSpawnTimers{
		Stick: 1.0, // About to fire
		Nut:   1.0,
		Shell: 1.0,
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	if len(gameMap.Items()) == 0 {
		t.Error("Expected items to spawn after interval elapses")
	}
}

func TestUpdateGroundSpawning_ResetsTimerAfterFiring(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	timers := &GroundSpawnTimers{
		Stick: 1.0, // About to fire
		Nut:   config.GroundSpawnInterval * 2, // Won't fire
		Shell: config.GroundSpawnInterval * 2, // Won't fire
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	// Stick timer should have been reset to a new interval in the expected range
	base := config.GroundSpawnInterval
	variance := base * config.LifecycleIntervalVariance
	if timers.Stick < base-variance || timers.Stick > base+variance {
		t.Errorf("Stick timer %.2f should be in range [%.2f, %.2f]", timers.Stick, base-variance, base+variance)
	}
}

// =============================================================================
// Stick spawning
// =============================================================================

func TestUpdateGroundSpawning_SpawnsStickOnEmptyTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	timers := &GroundSpawnTimers{
		Stick: 1.0,
		Nut:   config.GroundSpawnInterval * 2,
		Shell: config.GroundSpawnInterval * 2,
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	sticks := 0
	for _, item := range gameMap.Items() {
		if item.ItemType == "stick" {
			sticks++
			pos := item.Pos()
			if gameMap.IsWater(pos) {
				t.Errorf("Stick spawned on water tile at (%d, %d)", pos.X, pos.Y)
			}
		}
	}

	if sticks != 1 {
		t.Errorf("Expected 1 stick, got %d", sticks)
	}
}

// =============================================================================
// Nut spawning
// =============================================================================

func TestUpdateGroundSpawning_SpawnsNutOnEmptyTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	timers := &GroundSpawnTimers{
		Stick: config.GroundSpawnInterval * 2,
		Nut:   1.0,
		Shell: config.GroundSpawnInterval * 2,
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	nuts := 0
	for _, item := range gameMap.Items() {
		if item.ItemType == "nut" {
			nuts++
			pos := item.Pos()
			if gameMap.IsWater(pos) {
				t.Errorf("Nut spawned on water tile at (%d, %d)", pos.X, pos.Y)
			}
		}
	}

	if nuts != 1 {
		t.Errorf("Expected 1 nut, got %d", nuts)
	}
}

// =============================================================================
// Shell spawning (pond-adjacent only)
// =============================================================================

func TestUpdateGroundSpawning_SpawnsShellAdjacentToPond(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	gameMap.AddWater(types.Position{X: 10, Y: 10}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 11, Y: 10}, game.WaterPond)

	timers := &GroundSpawnTimers{
		Stick: config.GroundSpawnInterval * 2,
		Nut:   config.GroundSpawnInterval * 2,
		Shell: 1.0,
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	shells := 0
	for _, item := range gameMap.Items() {
		if item.ItemType == "shell" {
			shells++
			pos := item.Pos()
			adjacent := false
			cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
			for _, dir := range cardinalDirs {
				adjPos := types.Position{X: pos.X + dir[0], Y: pos.Y + dir[1]}
				if gameMap.WaterAt(adjPos) == game.WaterPond {
					adjacent = true
					break
				}
			}
			if !adjacent {
				t.Errorf("Shell at (%d, %d) is not adjacent to any pond tile", pos.X, pos.Y)
			}
		}
	}

	if shells != 1 {
		t.Errorf("Expected 1 shell, got %d", shells)
	}
}

func TestUpdateGroundSpawning_NoShellsWithoutPonds(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	// Only a spring, no ponds
	gameMap.AddWater(types.Position{X: 5, Y: 5}, game.WaterSpring)

	timers := &GroundSpawnTimers{
		Stick: config.GroundSpawnInterval * 2,
		Nut:   config.GroundSpawnInterval * 2,
		Shell: 1.0,
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	for _, item := range gameMap.Items() {
		if item.ItemType == "shell" {
			t.Error("Shells should not spawn without ponds")
		}
	}
}

// =============================================================================
// Independence: each type spawns on its own timer
// =============================================================================

func TestUpdateGroundSpawning_TimersFireIndependently(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	gameMap.AddWater(types.Position{X: 10, Y: 10}, game.WaterPond)

	timers := &GroundSpawnTimers{
		Stick: 1.0,                            // Will fire
		Nut:   config.GroundSpawnInterval * 2,  // Won't fire
		Shell: config.GroundSpawnInterval * 2,  // Won't fire
	}

	UpdateGroundSpawning(gameMap, 2.0, timers)

	sticks, nuts, shells := 0, 0, 0
	for _, item := range gameMap.Items() {
		switch item.ItemType {
		case "stick":
			sticks++
		case "nut":
			nuts++
		case "shell":
			shells++
		}
	}

	if sticks != 1 {
		t.Errorf("Expected 1 stick, got %d", sticks)
	}
	if nuts != 0 {
		t.Errorf("Expected 0 nuts, got %d", nuts)
	}
	if shells != 0 {
		t.Errorf("Expected 0 shells, got %d", shells)
	}
}

// =============================================================================
// RandomGroundSpawnInterval helper
// =============================================================================

func TestRandomGroundSpawnInterval_InExpectedRange(t *testing.T) {
	t.Parallel()

	base := config.GroundSpawnInterval
	variance := base * config.LifecycleIntervalVariance

	for i := 0; i < 100; i++ {
		interval := RandomGroundSpawnInterval()
		if interval < base-variance || interval > base+variance {
			t.Errorf("Interval %.2f outside expected range [%.2f, %.2f]", interval, base-variance, base+variance)
		}
	}
}
