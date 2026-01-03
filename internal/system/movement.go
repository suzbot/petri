package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// CalculateIntent determines what a character wants to do next tick
// This is safe to call concurrently - it only reads world state
func CalculateIntent(char *entity.Character, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	if char.IsDead || char.IsSleeping {
		return nil
	}

	// If frustrated, stay idle until timer expires
	if char.IsFrustrated {
		if char.CurrentActivity != "Frustrated" {
			char.CurrentActivity = "Frustrated"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Frustrated (can't meet needs)")
			}
		}
		return nil
	}

	cx, cy := char.Position()

	// Cache tier values (calculated once, reused throughout)
	hungerTier := char.HungerTier()
	thirstTier := char.ThirstTier()
	energyTier := char.EnergyTier()

	// Check if we should continue a looking intent (no DrivingStat)
	// Looking can be interrupted by: urgent needs (tier >= Moderate), or cooldown started (look completed)
	if char.Intent != nil && char.Intent.DrivingStat == "" && char.Intent.TargetItem != nil {
		// If cooldown is active, look just completed - fall through to re-evaluate
		if char.LookCooldown <= 0 {
			maxTier := hungerTier
			if thirstTier > maxTier {
				maxTier = thirstTier
			}
			if energyTier > maxTier {
				maxTier = energyTier
			}
			// Keep looking if no urgent needs (Moderate or higher)
			if maxTier < entity.TierModerate {
				return continueIntent(char, cx, cy, gameMap, log)
			}
		}
		// Otherwise fall through to re-evaluate (need interrupts looking, or look completed)
	}

	// Check if current intent should be kept (tier-based evaluation)
	if char.Intent != nil {
		currentDrivingTier := char.Intent.DrivingTier

		// Get the current tier of the driving stat
		var currentStatTier int
		switch char.Intent.DrivingStat {
		case types.StatHunger:
			currentStatTier = hungerTier
		case types.StatThirst:
			currentStatTier = thirstTier
		case types.StatEnergy:
			currentStatTier = energyTier
		}

		shouldReEval := false

		// Re-evaluate if the driving stat has been satisfied
		// Special case: when drinking at a spring, keep drinking until thirst == 0
		if char.Intent.DrivingStat == types.StatThirst && char.Intent.Action == entity.ActionDrink {
			// At a spring - only stop when fully satisfied
			if char.Thirst == 0 {
				shouldReEval = true
			}
		} else if currentStatTier == entity.TierNone {
			// Standard behavior: re-evaluate when tier drops to None
			shouldReEval = true
		}

		// Re-evaluate if a different stat has reached a higher tier AND can be fulfilled
		// This prevents thrashing between an unfulfillable higher-tier stat and a fulfillable lower-tier stat
		if !shouldReEval {
			switch char.Intent.DrivingStat {
			case types.StatHunger:
				if thirstTier > currentDrivingTier && canFulfillThirst(gameMap, cx, cy) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cx, cy) {
					shouldReEval = true
				}
			case types.StatThirst:
				if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cx, cy) {
					shouldReEval = true
				}
			case types.StatEnergy:
				if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if thirstTier > currentDrivingTier && canFulfillThirst(gameMap, cx, cy) {
					shouldReEval = true
				}
			default:
				shouldReEval = true
			}
		}

		if !shouldReEval {
			// Continue with current intent - recalculate next step
			return continueIntent(char, cx, cy, gameMap, log)
		}
	}

	// Evaluate which stat to address based on tier and tie-breakers

	// Find the highest tier
	maxTier := hungerTier
	if thirstTier > maxTier {
		maxTier = thirstTier
	}
	if energyTier > maxTier {
		maxTier = energyTier
	}

	// No urgent needs - try looking at something
	if maxTier == entity.TierNone {
		if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no needs)")
			}
		}
		return nil
	}

	// Build priority list: stats with needs, sorted by tier (desc), then tie-breaker (Thirst > Hunger > Energy)
	type statPriority struct {
		stat types.StatType
		tier int
	}
	var priorities []statPriority

	// Add stats that have needs (tier > 0), in tie-breaker order within same tier
	if thirstTier > 0 {
		priorities = append(priorities, statPriority{types.StatThirst, thirstTier})
	}
	if hungerTier > 0 {
		priorities = append(priorities, statPriority{types.StatHunger, hungerTier})
	}
	if energyTier > 0 {
		priorities = append(priorities, statPriority{types.StatEnergy, energyTier})
	}

	// Sort by tier descending (higher tier = more urgent)
	// Tie-breaker order is already correct since we added in Thirst > Hunger > Energy order
	for i := 0; i < len(priorities)-1; i++ {
		for j := i + 1; j < len(priorities); j++ {
			if priorities[j].tier > priorities[i].tier {
				priorities[i], priorities[j] = priorities[j], priorities[i]
			}
		}
	}

	// Try each stat in priority order, falling back if intent can't be fulfilled
	for _, p := range priorities {
		var intent *entity.Intent
		switch p.stat {
		case types.StatThirst:
			intent = findDrinkIntent(char, cx, cy, gameMap, p.tier, log)
		case types.StatHunger:
			intent = findFoodIntent(char, cx, cy, items, p.tier, log)
		case types.StatEnergy:
			intent = findSleepIntent(char, cx, cy, gameMap, p.tier, log)
		}
		if intent != nil {
			// Successfully found an intent - reset failure counter
			char.FailedIntentCount = 0
			return intent
		}
	}

	// No intent could be fulfilled - track failure only at Severe+ tier (prevents thrashing where it matters)
	if maxTier >= entity.TierSevere {
		char.FailedIntentCount++
		if char.FailedIntentCount >= config.FrustrationThreshold {
			char.IsFrustrated = true
			char.FrustrationTimer = config.FrustrationDuration
			char.FailedIntentCount = 0
			char.CurrentActivity = "Frustrated"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Frustrated (can't meet needs)")
			}
			return nil
		}
	}

	// No needs could be fulfilled - try looking at something
	if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
		return intent
	}

	if char.CurrentActivity != "Idle" {
		char.CurrentActivity = "Idle"
		if log != nil {
			log.Add(char.ID, char.Name, "activity", "Idle (no options)")
		}
	}
	return nil
}

// continueIntent recalculates the next step for an existing intent
func continueIntent(char *entity.Character, cx, cy int, gameMap *game.Map, log *ActionLog) *entity.Intent {
	intent := char.Intent

	// Check if target item still exists at expected position (O(1) instead of O(n))
	if intent.TargetItem != nil {
		ix, iy := intent.TargetItem.Position()
		if gameMap.ItemAt(ix, iy) != intent.TargetItem {
			return nil // Target consumed by someone else
		}
	}

	// Check if target feature is now occupied by another character
	if intent.TargetFeature != nil {
		fx, fy := intent.TargetFeature.Position()
		occupant := gameMap.CharacterAt(fx, fy)
		if occupant != nil && occupant != char {
			return nil // Target occupied by someone else, find new target
		}
	}

	// Recalculate next step toward target
	var tx, ty int
	if intent.TargetItem != nil {
		tx, ty = intent.TargetItem.Position()
	} else if intent.TargetFeature != nil {
		tx, ty = intent.TargetFeature.Position()
	} else {
		tx, ty = intent.TargetX, intent.TargetY
	}

	// Check if we've arrived at a feature target - switch to appropriate action
	if intent.TargetFeature != nil && cx == tx && cy == ty {
		feature := intent.TargetFeature
		if feature.IsDrinkSource() {
			newActivity := "Drinking"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
				if log != nil {
					log.Add(char.ID, char.Name, "thirst", "Drinking from spring")
				}
			}
			return &entity.Intent{
				TargetX:       tx,
				TargetY:       ty,
				Action:        entity.ActionDrink,
				TargetFeature: feature,
				DrivingStat:   intent.DrivingStat,
				DrivingTier:   intent.DrivingTier,
			}
		}
		if feature.IsBed() {
			newActivity := "Sleeping (in bed)"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				TargetX:       tx,
				TargetY:       ty,
				Action:        entity.ActionSleep,
				TargetFeature: feature,
				DrivingStat:   intent.DrivingStat,
				DrivingTier:   intent.DrivingTier,
			}
		}
	}

	// Check if we've arrived adjacent to an item for looking (no DrivingStat means looking intent)
	if intent.TargetItem != nil && intent.DrivingStat == "" && isAdjacent(cx, cy, tx, ty) {
		newActivity := "Looking at " + intent.TargetItem.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			TargetX:    cx, // Stay in place
			TargetY:    cy,
			Action:     entity.ActionLook,
			TargetItem: intent.TargetItem,
		}
	}

	nx, ny := nextStep(cx, cy, tx, ty)

	return &entity.Intent{
		TargetX:       nx,
		TargetY:       ny,
		Action:        intent.Action,
		TargetItem:    intent.TargetItem,
		TargetFeature: intent.TargetFeature,
		DrivingStat:   intent.DrivingStat,
		DrivingTier:   intent.DrivingTier,
	}
}

// findDrinkIntent finds a spring to drink from
func findDrinkIntent(char *entity.Character, cx, cy int, gameMap *game.Map, tier int, log *ActionLog) *entity.Intent {
	spring := gameMap.FindNearestDrinkSource(cx, cy)
	if spring == nil {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no water source)")
			}
		}
		return nil
	}

	tx, ty := spring.Position()

	// Already at spring - drink
	if cx == tx && cy == ty {
		newActivity := "Drinking"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "thirst", "Drinking from spring")
			}
		}
		return &entity.Intent{
			TargetX:       tx,
			TargetY:       ty,
			Action:        entity.ActionDrink,
			TargetFeature: spring,
			DrivingStat:   types.StatThirst,
			DrivingTier:   tier,
		}
	}

	// Move toward spring
	nx, ny := nextStep(cx, cy, tx, ty)
	newActivity := "Moving to spring"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Heading to spring")
		}
	}

	return &entity.Intent{
		TargetX:       nx,
		TargetY:       ny,
		Action:        entity.ActionMove,
		TargetFeature: spring,
		DrivingStat:   types.StatThirst,
		DrivingTier:   tier,
	}
}

// findFoodIntent finds food based on hunger priority
func findFoodIntent(char *entity.Character, cx, cy int, items []*entity.Item, tier int, log *ActionLog) *entity.Intent {
	target := findFoodTarget(char, items)
	if target == nil {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no suitable food)")
			}
		}
		return nil
	}

	tx, ty := target.Position()
	nx, ny := nextStep(cx, cy, tx, ty)

	newActivity := "Moving to " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Started moving to "+target.Description())
		}
	}

	return &entity.Intent{
		TargetX:     nx,
		TargetY:     ny,
		Action:      entity.ActionMove,
		TargetItem:  target,
		DrivingStat: types.StatHunger,
		DrivingTier: tier,
	}
}

// findSleepIntent finds a bed to sleep in
func findSleepIntent(char *entity.Character, cx, cy int, gameMap *game.Map, tier int, log *ActionLog) *entity.Intent {
	bed := gameMap.FindNearestBed(cx, cy)

	// If no bed, can sleep on ground when exhausted (voluntary) or collapsed (involuntary)
	if bed == nil {
		if char.Energy <= 10 { // Exhausted - ground sleep available
			newActivity := "Sleeping (on ground)"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				TargetX:     cx,
				TargetY:     cy,
				Action:      entity.ActionSleep,
				DrivingStat: types.StatEnergy,
				DrivingTier: tier,
			}
		}
		// Not tired enough for ground sleep, need a bed
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no bed nearby)")
			}
		}
		return nil
	}

	tx, ty := bed.Position()

	// Already at bed - sleep
	if cx == tx && cy == ty {
		newActivity := "Sleeping (in bed)"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			TargetX:       tx,
			TargetY:       ty,
			Action:        entity.ActionSleep,
			TargetFeature: bed,
			DrivingStat:   types.StatEnergy,
			DrivingTier:   tier,
		}
	}

	// Move toward bed
	nx, ny := nextStep(cx, cy, tx, ty)
	newActivity := "Moving to leaf pile"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Heading to leaf pile")
		}
	}

	return &entity.Intent{
		TargetX:       nx,
		TargetY:       ny,
		Action:        entity.ActionMove,
		TargetFeature: bed,
		DrivingStat:   types.StatEnergy,
		DrivingTier:   tier,
	}
}

// findFoodTarget finds the best item for a character based on hunger priority
// Uses single-pass algorithm for O(n) instead of O(n*5) performance
func findFoodTarget(char *entity.Character, items []*entity.Item) *entity.Item {
	if len(items) == 0 {
		return nil
	}

	cx, cy := char.Position()
	maxDist := int(^uint(0) >> 1)

	// Track best match in each category (single pass)
	// Perfect = both preferences match (NetPreference >= 2)
	// Partial = one preference matches (NetPreference >= 1)
	// Any = any edible item
	var bestPerfect, bestPartial, bestAny *entity.Item
	distPerfect, distPartial, distAny := maxDist, maxDist, maxDist

	for _, item := range items {
		// Skip non-edible items (e.g., flowers)
		if !item.Edible {
			continue
		}

		ix, iy := item.Position()
		dist := abs(cx-ix) + abs(cy-iy)

		// Check match types using NetPreference
		pref := char.NetPreference(item)
		isPerfect := pref >= 2
		isPartial := pref >= 1

		// Update best in each category if closer
		if isPerfect && dist < distPerfect {
			bestPerfect = item
			distPerfect = dist
		}
		if isPartial && dist < distPartial {
			bestPartial = item
			distPartial = dist
		}
		if dist < distAny {
			bestAny = item
			distAny = dist
		}
	}

	// Return based on hunger tier
	// 90+: Ravenous - eat anything
	if char.Hunger >= 90 {
		return bestAny
	}

	// 75-89: Very hungry - partial match, then any
	if char.Hunger >= 75 {
		if bestPartial != nil {
			return bestPartial
		}
		return bestAny
	}

	// 50-74: Moderately hungry - perfect match, then partial match
	if bestPerfect != nil {
		return bestPerfect
	}
	return bestPartial
}

// nextStep calculates the next position moving toward target
func nextStep(fromX, fromY, toX, toY int) (int, int) {
	dx := toX - fromX
	dy := toY - fromY

	if dx == 0 && dy == 0 {
		return fromX, fromY
	}

	// Move toward larger distance
	if abs(dx) > abs(dy) {
		return fromX + sign(dx), fromY
	}
	return fromX, fromY + sign(dy)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

// canFulfillThirst checks if thirst can be addressed (water source exists)
func canFulfillThirst(gameMap *game.Map, cx, cy int) bool {
	return gameMap.FindNearestDrinkSource(cx, cy) != nil
}

// canFulfillHunger checks if hunger can be addressed (suitable food exists)
func canFulfillHunger(char *entity.Character, items []*entity.Item) bool {
	return findFoodTarget(char, items) != nil
}

// canFulfillEnergy checks if energy can be addressed (bed exists or exhausted enough for ground sleep)
func canFulfillEnergy(char *entity.Character, gameMap *game.Map, cx, cy int) bool {
	// Can sleep on ground if exhausted (voluntary at ≤10, involuntary collapse at 0)
	if char.Energy <= 10 {
		return true
	}
	// Otherwise need a bed
	return gameMap.FindNearestBed(cx, cy) != nil
}

// findLookIntent creates an intent to look at the nearest item (50% chance when idle)
func findLookIntent(char *entity.Character, cx, cy int, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Check cooldown
	if char.LookCooldown > 0 {
		return nil
	}

	// 50% chance to look - set cooldown regardless of outcome so check only happens periodically
	if rand.Float64() >= config.LookChance {
		char.LookCooldown = config.LookCooldown
		return nil
	}

	// Find nearest item, excluding last looked item
	target := findNearestItemExcluding(cx, cy, items, char.LastLookedX, char.LastLookedY, char.HasLastLooked)
	if target == nil {
		return nil
	}

	tx, ty := target.Position()

	// Check if already adjacent to target
	if isAdjacent(cx, cy, tx, ty) {
		// Start looking immediately
		newActivity := "Looking at " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Looking at "+target.Description())
			}
		}
		return &entity.Intent{
			TargetX:    cx, // Stay in place
			TargetY:    cy,
			Action:     entity.ActionLook,
			TargetItem: target,
		}
	}

	// Find closest adjacent tile to target
	adjX, adjY := findClosestAdjacentTile(cx, cy, tx, ty, gameMap)
	if adjX == -1 {
		return nil // No accessible adjacent tile
	}

	// Move toward adjacent tile
	nx, ny := nextStep(cx, cy, adjX, adjY)

	newActivity := "Moving to look at " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Moving to look at "+target.Description())
		}
	}

	return &entity.Intent{
		TargetX:    nx,
		TargetY:    ny,
		Action:     entity.ActionMove,
		TargetItem: target,
	}
}

// findNearestItem finds the closest item of any type to the given position
func findNearestItem(cx, cy int, items []*entity.Item) *entity.Item {
	return findNearestItemExcluding(cx, cy, items, 0, 0, false)
}

// findNearestItemExcluding finds the closest item, optionally excluding a specific position
func findNearestItemExcluding(cx, cy int, items []*entity.Item, excludeX, excludeY int, hasExclude bool) *entity.Item {
	if len(items) == 0 {
		return nil
	}

	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1)

	for _, item := range items {
		ix, iy := item.Position()

		// Skip excluded position
		if hasExclude && ix == excludeX && iy == excludeY {
			continue
		}

		dist := abs(cx-ix) + abs(cy-iy)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	return nearest
}

// isAdjacent checks if two positions are adjacent (including diagonals)
func isAdjacent(x1, y1, x2, y2 int) bool {
	dx := abs(x1 - x2)
	dy := abs(y1 - y2)
	return dx <= 1 && dy <= 1 && !(dx == 0 && dy == 0)
}

// findClosestAdjacentTile finds the closest unoccupied tile adjacent to (tx, ty) from position (cx, cy)
func findClosestAdjacentTile(cx, cy, tx, ty int, gameMap *game.Map) (int, int) {
	// 8 directions
	directions := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}

	bestX, bestY := -1, -1
	bestDist := int(^uint(0) >> 1)

	for _, dir := range directions {
		ax, ay := tx+dir[0], ty+dir[1]
		if !gameMap.IsValid(ax, ay) {
			continue
		}
		if gameMap.IsOccupied(ax, ay) {
			continue
		}

		dist := abs(cx-ax) + abs(cy-ay)
		if dist < bestDist {
			bestDist = dist
			bestX, bestY = ax, ay
		}
	}

	return bestX, bestY
}
