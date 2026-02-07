package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// GroundSpawnTimers holds independent timers for periodic ground item spawning.
// Each item type (stick, nut, shell) spawns on its own random cycle.
type GroundSpawnTimers struct {
	Stick float64
	Nut   float64
	Shell float64
}

// UpdateGroundSpawning decrements each ground spawn timer and spawns one item
// when a timer fires. Each timer resets to a new random interval after firing.
func UpdateGroundSpawning(gameMap *game.Map, delta float64, timers *GroundSpawnTimers) {
	timers.Stick -= delta
	if timers.Stick <= 0 {
		timers.Stick = RandomGroundSpawnInterval()
		spawnGroundItem(gameMap, "stick")
	}

	timers.Nut -= delta
	if timers.Nut <= 0 {
		timers.Nut = RandomGroundSpawnInterval()
		spawnGroundItem(gameMap, "nut")
	}

	timers.Shell -= delta
	if timers.Shell <= 0 {
		timers.Shell = RandomGroundSpawnInterval()
		spawnShell(gameMap)
	}
}

// RandomGroundSpawnInterval returns a randomized spawn interval for ground items.
// Uses GroundSpawnInterval ± LifecycleIntervalVariance (same pattern as plant lifecycle).
func RandomGroundSpawnInterval() float64 {
	base := config.GroundSpawnInterval
	variance := base * config.LifecycleIntervalVariance
	return base + (rand.Float64()*2-1)*variance
}

// spawnGroundItem spawns one item of the given type on a random empty tile.
// Tries up to 10 times to find a valid spot; gives up silently if the map is too full.
func spawnGroundItem(gameMap *game.Map, itemType string) {
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		x := rand.Intn(gameMap.Width)
		y := rand.Intn(gameMap.Height)
		pos := types.Position{X: x, Y: y}
		if !gameMap.IsEmpty(pos) {
			continue
		}

		switch itemType {
		case "stick":
			gameMap.AddItem(entity.NewStick(x, y))
		case "nut":
			gameMap.AddItem(entity.NewNut(x, y))
		}
		return
	}
}

// spawnShell spawns one shell adjacent to a random pond tile.
// Does nothing if no ponds exist or no pond-adjacent tiles are available.
func spawnShell(gameMap *game.Map) {
	tiles := game.FindPondAdjacentEmptyTiles(gameMap)
	if len(tiles) == 0 {
		return
	}

	pos := tiles[rand.Intn(len(tiles))]
	color := types.ShellColors[rand.Intn(len(types.ShellColors))]
	gameMap.AddItem(entity.NewShell(pos.X, pos.Y, color))
}
