// Package simulation provides integration testing utilities for running
// multi-tick game simulations and verifying invariants.
package simulation

import (
	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/system"
	"petri/internal/types"
)

// WorldOptions configures test world creation
type WorldOptions struct {
	NoWater       bool
	NoFood        bool
	NoBeds        bool
	NumCharacters int
}

// TestWorld holds all components needed to run a simulation
type TestWorld struct {
	GameMap   *game.Map
	ActionLog *system.ActionLog
}

// CreateTestWorld creates a world configured for testing
func CreateTestWorld(opts WorldOptions) *TestWorld {
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	actionLog := system.NewActionLog(200)

	numChars := opts.NumCharacters
	if numChars == 0 {
		numChars = 4
	}

	// Place characters in a cluster near center
	cx, cy := config.MapWidth/2, config.MapHeight/2
	offsets := [][2]int{{0, 0}, {2, 0}, {0, 2}, {2, 2}, {4, 0}, {0, 4}, {4, 2}, {2, 4}}
	names := []string{"Len", "Macca", "Hari", "Starr", "Test5", "Test6", "Test7", "Test8"}
	foods := []string{"berry", "mushroom"}
	colors := types.AllColors

	for i := 0; i < numChars && i < len(offsets); i++ {
		x := cx + offsets[i][0]
		y := cy + offsets[i][1]
		food := foods[i%len(foods)]
		color := colors[i%len(colors)]
		name := names[i]
		char := entity.NewCharacter(i+1, x, y, name, food, color)
		gameMap.AddCharacter(char)
	}

	// Spawn resources based on options
	if !opts.NoFood {
		game.SpawnItems(gameMap, false) // mushroomsOnly=false for tests
	}
	game.SpawnFeatures(gameMap, opts.NoWater, opts.NoBeds)

	return &TestWorld{
		GameMap:   gameMap,
		ActionLog: actionLog,
	}
}

// RunTick runs one complete simulation tick
func RunTick(world *TestWorld, delta float64) {
	chars := world.GameMap.Characters()
	items := world.GameMap.Items()

	// Phase 1: Update survival for all characters
	for _, char := range chars {
		system.UpdateSurvival(char, delta, world.ActionLog)
	}

	// Phase 2: Calculate intents for all characters
	for _, char := range chars {
		oldIntent := char.Intent
		char.Intent = system.CalculateIntent(char, items, world.GameMap, world.ActionLog)

		// Reset action progress if intent action changed
		if oldIntent == nil || char.Intent == nil || oldIntent.Action != char.Intent.Action {
			char.ActionProgress = 0
		}
	}

	// Phase 3: Apply intents
	for _, char := range chars {
		applyIntent(char, world.GameMap, delta, world.ActionLog)
	}

	// Phase 4: Update item lifecycle
	initialItemCount := config.ItemSpawnCount*2 + config.FlowerSpawnCount // berries + mushrooms + flowers
	system.UpdateSpawnTimers(world.GameMap, initialItemCount, delta)
	system.UpdateDeathTimers(world.GameMap, delta)
}

// RunTicks runs n simulation ticks
func RunTicks(world *TestWorld, n int, delta float64) {
	for i := 0; i < n; i++ {
		RunTick(world, delta)
	}
}

// applyIntent executes a character's intent (mirrors ui/update.go logic)
func applyIntent(char *entity.Character, gameMap *game.Map, delta float64, actionLog *system.ActionLog) {
	if char.Intent == nil || char.IsDead || char.IsSleeping {
		return
	}

	switch char.Intent.Action {
	case entity.ActionMove:
		applyMoveIntent(char, gameMap, delta, actionLog)

	case entity.ActionDrink:
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDuration {
			char.ActionProgress = 0
			system.Drink(char, actionLog)
		}

	case entity.ActionSleep:
		atBed := char.Intent.TargetFeature != nil && char.Intent.TargetFeature.IsBed()

		// Collapse is immediate (involuntary) - only at Energy 0
		if !atBed && char.Energy <= 0 {
			system.StartSleep(char, false, actionLog)
			return
		}

		// Voluntary sleep requires duration
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDuration {
			char.ActionProgress = 0
			system.StartSleep(char, atBed, actionLog)
		}

	case entity.ActionPickup:
		applyPickupIntent(char, gameMap, delta, actionLog)
	}
}

// applyPickupIntent handles foraging - movement and pickup at destination
func applyPickupIntent(char *entity.Character, gameMap *game.Map, delta float64, actionLog *system.ActionLog) {
	cx, cy := char.Position()

	if char.Intent.TargetItem == nil {
		return
	}

	ix, iy := char.Intent.TargetItem.Position()

	// Check if at target item
	if cx == ix && cy == iy {
		// At item - pickup in progress
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDuration {
			char.ActionProgress = 0
			if item := gameMap.ItemAt(cx, cy); item == char.Intent.TargetItem {
				system.Pickup(char, item, gameMap, actionLog)
			}
		}
		return
	}

	// Not at item yet - move toward it
	speed := char.EffectiveSpeed()
	char.SpeedAccumulator += float64(speed) * delta

	const movementThreshold = 7.5

	if char.SpeedAccumulator < movementThreshold {
		return
	}

	char.SpeedAccumulator -= movementThreshold

	// Move toward target item
	tx, ty := char.Intent.TargetX, char.Intent.TargetY
	if gameMap.MoveCharacter(char, tx, ty) {
		// Successfully moved - update intent for next step
		newX, newY := char.Position()
		if newX != ix || newY != iy {
			// Need to keep moving toward item
			nextX, nextY := system.NextStep(newX, newY, ix, iy)
			char.Intent.TargetX = nextX
			char.Intent.TargetY = nextY
		}
	}
}

// applyMoveIntent handles movement and eating at destination
func applyMoveIntent(char *entity.Character, gameMap *game.Map, delta float64, actionLog *system.ActionLog) {
	cx, cy := char.Position()
	tx, ty := char.Intent.TargetX, char.Intent.TargetY

	// Check if at target item - eating takes duration
	if char.Intent.TargetItem != nil {
		ix, iy := char.Intent.TargetItem.Position()
		if cx == ix && cy == iy {
			// At target item - eating in progress
			char.ActionProgress += delta
			if char.ActionProgress >= config.ActionDuration {
				char.ActionProgress = 0
				if item := gameMap.ItemAt(cx, cy); item == char.Intent.TargetItem {
					system.Consume(char, item, gameMap, actionLog)
				}
			}
			return
		}
	}

	// Movement is gated by speed accumulator
	speed := char.EffectiveSpeed()
	char.SpeedAccumulator += float64(speed) * delta

	const movementThreshold = 7.5 // 50 speed * 0.15s delta

	if char.SpeedAccumulator < movementThreshold {
		return
	}

	char.SpeedAccumulator -= movementThreshold

	// Try to move, with collision handling
	moved := false
	triedPositions := make(map[[2]int]bool)
	triedPositions[[2]int{tx, ty}] = true

	for attempts := 0; attempts < 5 && !moved; attempts++ {
		if gameMap.MoveCharacter(char, tx, ty) {
			moved = true
			break
		}
		// Position blocked, try alternate
		altStep := findAlternateStep(char, gameMap, cx, cy, triedPositions)
		if altStep == nil {
			break
		}
		tx, ty = altStep[0], altStep[1]
		triedPositions[[2]int{tx, ty}] = true
	}

	if !moved {
		// Couldn't move, refund some speed points
		char.SpeedAccumulator += movementThreshold * 0.5
		return
	}

	// Drain energy for movement (unless on cooldown - freshly rested burst)
	if char.EnergyCooldown <= 0 {
		prevEnergy := char.Energy
		char.Energy -= config.EnergyMovementDrain
		if char.Energy < 0 {
			char.Energy = 0
		}

		// Log energy milestones crossed by movement drain
		if !char.IsSleeping && actionLog != nil {
			if prevEnergy > 50 && char.Energy <= 50 {
				actionLog.Add(char.ID, char.Name, "energy", "Getting tired")
			}
			if prevEnergy > 25 && char.Energy <= 25 {
				actionLog.Add(char.ID, char.Name, "energy", "Very tired!")
			}
			if prevEnergy > 10 && char.Energy <= 10 {
				actionLog.Add(char.ID, char.Name, "energy", "Exhausted!")
			}
			if prevEnergy > 0 && char.Energy <= 0 {
				actionLog.Add(char.ID, char.Name, "energy", "Collapsed from exhaustion!")
			}
		}
	}
}

// findAlternateStep finds an alternate step when preferred is blocked
func findAlternateStep(char *entity.Character, gameMap *game.Map, cx, cy int, triedPositions map[[2]int]bool) []int {
	var goalX, goalY int
	if char.Intent.TargetItem != nil {
		goalX, goalY = char.Intent.TargetItem.Position()
	} else if char.Intent.TargetFeature != nil {
		goalX, goalY = char.Intent.TargetFeature.Position()
	} else {
		return nil
	}

	dx := sign(goalX - cx)
	dy := sign(goalY - cy)

	// Build candidate positions
	candidates := [][]int{}

	// Primary: move in one axis toward goal
	if dx != 0 {
		candidates = append(candidates, []int{cx + dx, cy})
	}
	if dy != 0 {
		candidates = append(candidates, []int{cx, cy + dy})
	}

	// Secondary: orthogonal moves
	if dx == 0 {
		candidates = append(candidates, []int{cx + 1, cy}, []int{cx - 1, cy})
	}
	if dy == 0 {
		candidates = append(candidates, []int{cx, cy + 1}, []int{cx, cy - 1})
	}

	// Tertiary: all adjacent
	allAdjacent := [][]int{
		{cx + 1, cy}, {cx - 1, cy}, {cx, cy + 1}, {cx, cy - 1},
		{cx + 1, cy + 1}, {cx + 1, cy - 1}, {cx - 1, cy + 1}, {cx - 1, cy - 1},
	}
	for _, pos := range allAdjacent {
		found := false
		for _, c := range candidates {
			if c[0] == pos[0] && c[1] == pos[1] {
				found = true
				break
			}
		}
		if !found {
			candidates = append(candidates, pos)
		}
	}

	// Find first valid candidate
	for _, pos := range candidates {
		x, y := pos[0], pos[1]
		key := [2]int{x, y}
		if triedPositions[key] {
			continue
		}
		if !gameMap.IsValid(x, y) {
			continue
		}
		if gameMap.IsOccupied(x, y) {
			continue
		}
		return pos
	}

	return nil
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
