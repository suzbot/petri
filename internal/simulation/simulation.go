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
	NoCharacters  bool
	NumCharacters int
}

// TestWorld holds all components needed to run a simulation
type TestWorld struct {
	GameMap           *game.Map
	ActionLog         *system.ActionLog
	GroundSpawnTimers system.GroundSpawnTimers
}

// CreateTestWorld creates a world configured for testing
func CreateTestWorld(opts WorldOptions) *TestWorld {
	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	actionLog := system.NewActionLog(200)

	// Create characters unless explicitly disabled
	if !opts.NoCharacters {
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
	}

	// Spawn world: ponds first (before items/features), then features, then items
	if !opts.NoWater {
		game.SpawnPonds(gameMap)
	}
	game.SpawnFeatures(gameMap, opts.NoWater, opts.NoBeds)
	if !opts.NoFood {
		game.SpawnItems(gameMap, false) // mushroomsOnly=false for tests
	}
	game.SpawnGroundItems(gameMap)

	return &TestWorld{
		GameMap:   gameMap,
		ActionLog: actionLog,
		GroundSpawnTimers: system.GroundSpawnTimers{
			Stick: system.RandomGroundSpawnInterval(),
			Nut:   system.RandomGroundSpawnInterval(),
			Shell: system.RandomGroundSpawnInterval(),
		},
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
		char.Intent = system.CalculateIntent(char, items, world.GameMap, world.ActionLog, nil)

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
	system.UpdateSproutTimers(world.GameMap, initialItemCount, delta)
	system.UpdateDeathTimers(world.GameMap, delta)
	world.GameMap.UpdateWateredTimers(delta)

	// Phase 5: Update ground spawning (sticks, nuts, shells)
	system.UpdateGroundSpawning(world.GameMap, delta, &world.GroundSpawnTimers)
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
		if char.ActionProgress >= config.ActionDurationShort {
			char.ActionProgress = 0
			system.Drink(char, actionLog)
		}

	case entity.ActionConsume:
		// Eating from inventory - no movement needed, just duration
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDurationShort {
			char.ActionProgress = 0
			targetItem := char.Intent.TargetItem
			// Check if target item is in inventory
			inInventory := char.FindInInventory(func(i *entity.Item) bool { return i == targetItem }) != nil
			if inInventory {
				// Check if it's a vessel with edible contents
				if targetItem.Container != nil && len(targetItem.Container.Contents) > 0 {
					system.ConsumeFromVessel(char, targetItem, gameMap, actionLog)
				} else {
					system.ConsumeFromInventory(char, targetItem, gameMap, actionLog)
				}
			}
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
		if char.ActionProgress >= config.ActionDurationShort {
			char.ActionProgress = 0
			system.StartSleep(char, atBed, actionLog)
		}

	case entity.ActionPickup:
		// Picking up an item (used by both foraging and harvest orders)
		// Order completion is handled in UI layer
		applyPickupIntent(char, gameMap, delta, actionLog)
	}
}

// applyPickupIntent handles foraging - movement and pickup at destination
func applyPickupIntent(char *entity.Character, gameMap *game.Map, delta float64, actionLog *system.ActionLog) {
	cpos := char.Pos()
	cx, cy := cpos.X, cpos.Y

	if char.Intent.TargetItem == nil {
		return
	}

	ipos := char.Intent.TargetItem.Pos()

	// Check if at target item
	if cx == ipos.X && cy == ipos.Y {
		// At item - pickup in progress
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDurationShort {
			char.ActionProgress = 0
			if item := gameMap.ItemAt(types.Position{X: cx, Y: cy}); item == char.Intent.TargetItem {
				system.Pickup(char, item, gameMap, actionLog, gameMap.Varieties())
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
	tx, ty := char.Intent.Target.X, char.Intent.Target.Y
	if gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
		// Successfully moved - update intent for next step
		newPos := char.Pos()
		if newPos.X != ipos.X || newPos.Y != ipos.Y {
			// Need to keep moving toward item
			nextX, nextY := system.NextStep(newPos.X, newPos.Y, ipos.X, ipos.Y)
			char.Intent.Target.X = nextX
			char.Intent.Target.Y = nextY
		}
	}
}

// applyMoveIntent handles movement and eating at destination
func applyMoveIntent(char *entity.Character, gameMap *game.Map, delta float64, actionLog *system.ActionLog) {
	cpos := char.Pos()
	cx, cy := cpos.X, cpos.Y
	tx, ty := char.Intent.Target.X, char.Intent.Target.Y

	// Check if at target item - eating takes duration
	if char.Intent.TargetItem != nil {
		ipos := char.Intent.TargetItem.Pos()
		if cx == ipos.X && cy == ipos.Y {
			// At target item - eating in progress
			char.ActionProgress += delta
			if char.ActionProgress >= config.ActionDurationShort {
				char.ActionProgress = 0
				if item := gameMap.ItemAt(types.Position{X: cx, Y: cy}); item == char.Intent.TargetItem {
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
		if gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
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
	// Use destination position (where we need to stand to interact)
	goalX, goalY := char.Intent.Dest.X, char.Intent.Dest.Y
	if goalX == 0 && goalY == 0 {
		// Fallback for intents without destination set
		if char.Intent.TargetItem != nil {
			gpos := char.Intent.TargetItem.Pos()
			goalX, goalY = gpos.X, gpos.Y
		} else if char.Intent.TargetWaterPos != nil {
			goalX, goalY = char.Intent.TargetWaterPos.X, char.Intent.TargetWaterPos.Y
		} else if char.Intent.TargetFeature != nil {
			gpos := char.Intent.TargetFeature.Pos()
			goalX, goalY = gpos.X, gpos.Y
		} else {
			return nil
		}
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
		candidatePos := types.Position{X: x, Y: y}
		if !gameMap.IsValid(candidatePos) {
			continue
		}
		if gameMap.IsOccupied(candidatePos) {
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
