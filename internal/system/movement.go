package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// CalculateIntent determines what a character wants to do next tick
// This is safe to call concurrently - it only reads world state
func CalculateIntent(char *entity.Character, items []*entity.Item, gameMap *game.Map, log *ActionLog, orders []*entity.Order) *entity.Intent {
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

	cpos := char.Pos()
	cx, cy := cpos.X, cpos.Y

	// Cache tier values (calculated once, reused throughout)
	hungerTier := char.HungerTier()
	thirstTier := char.ThirstTier()
	energyTier := char.EnergyTier()
	healthTier := char.HealthTier()

	// Check if we should continue an idle activity (looking or talking)
	// Idle activities can be interrupted by urgent needs (tier >= Moderate)
	if char.Intent != nil && char.Intent.DrivingStat == "" {
		maxTier := hungerTier
		if thirstTier > maxTier {
			maxTier = thirstTier
		}
		if energyTier > maxTier {
			maxTier = energyTier
		}
		// Health only counts if we can fulfill it (have healing knowledge)
		if healthTier > maxTier && canFulfillHealth(char, items) {
			maxTier = healthTier
		}

		// Continue looking/working if no urgent needs
		if char.Intent.TargetItem != nil && maxTier < entity.TierModerate {
			if result := continueIntent(char, cx, cy, gameMap, log); result != nil {
				return result
			}
			// Target gone (e.g., taken by another character), fall through to re-evaluate
		}

		// Continue talking/approaching if no urgent needs
		if char.Intent.TargetCharacter != nil && maxTier < entity.TierModerate {
			target := char.Intent.TargetCharacter
			// If already talking, continue
			if char.TalkingWith != nil {
				return &entity.Intent{
					Target:          types.Position{X: cx, Y: cy},
					Dest:            types.Position{X: cx, Y: cy}, // Already at destination (talking)
					Action:          entity.ActionTalk,
					TargetCharacter: char.TalkingWith,
				}
			}
			// If approaching to talk, check target is still valid (idle activity, not dead/sleeping)
			if !target.IsDead && !target.IsSleeping && isIdleActivity(target.CurrentActivity) {
				return continueIntent(char, cx, cy, gameMap, log)
			}
			// Target no longer valid, fall through to re-evaluate
		}

		// If talking but interrupted by Moderate+ need, stop partner too
		if char.TalkingWith != nil && maxTier >= entity.TierModerate {
			StopTalking(char, char.TalkingWith, log)
		}
		// Fall through to re-evaluate
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
		case types.StatHealth:
			currentStatTier = healthTier
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
				if thirstTier > currentDrivingTier && canFulfillThirst(char, gameMap, cpos, items) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cpos) {
					shouldReEval = true
				} else if healthTier > currentDrivingTier && canFulfillHealth(char, items) {
					shouldReEval = true
				}
			case types.StatThirst:
				if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cpos) {
					shouldReEval = true
				} else if healthTier > currentDrivingTier && canFulfillHealth(char, items) {
					shouldReEval = true
				}
			case types.StatEnergy:
				if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if thirstTier > currentDrivingTier && canFulfillThirst(char, gameMap, cpos, items) {
					shouldReEval = true
				} else if healthTier > currentDrivingTier && canFulfillHealth(char, items) {
					shouldReEval = true
				}
			case types.StatHealth:
				if thirstTier > currentDrivingTier && canFulfillThirst(char, gameMap, cpos, items) {
					shouldReEval = true
				} else if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cpos) {
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

	// Find the highest tier (health only counts if we can fulfill it)
	maxTier := hungerTier
	if thirstTier > maxTier {
		maxTier = thirstTier
	}
	if energyTier > maxTier {
		maxTier = energyTier
	}
	if healthTier > maxTier && canFulfillHealth(char, items) {
		maxTier = healthTier
	}

	// No urgent needs - try an idle activity (looking, talking, or staying idle)
	if maxTier == entity.TierNone {
		if intent := selectIdleActivity(char, cpos, items, gameMap, log, orders); intent != nil {
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

	// Build priority list: stats with needs, sorted by tier (desc), then tie-breaker (Thirst > Hunger > Health > Energy)
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
	// Health only added if can be fulfilled (requires healing knowledge)
	if healthTier > 0 && canFulfillHealth(char, items) {
		priorities = append(priorities, statPriority{types.StatHealth, healthTier})
	}
	if energyTier > 0 {
		priorities = append(priorities, statPriority{types.StatEnergy, energyTier})
	}

	// Sort by tier descending (higher tier = more urgent)
	// Tie-breaker order is already correct since we added in Thirst > Hunger > Health > Energy order
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
			intent = findDrinkIntent(char, cpos, gameMap, p.tier, log, items)
		case types.StatHunger:
			intent = findFoodIntent(char, cpos, items, p.tier, log, gameMap)
		case types.StatHealth:
			intent = findHealingIntent(char, cpos, items, p.tier, log, gameMap)
		case types.StatEnergy:
			intent = findSleepIntent(char, cpos, gameMap, p.tier, log)
		}
		if intent != nil {
			// Check for order interruption - character has assigned order but pursuing a need
			// (DrivingStat being set means this is a need-driven intent, not order work)
			if char.AssignedOrderID != 0 && intent.DrivingStat != "" {
				order := findOrderByID(orders, char.AssignedOrderID)
				if order != nil {
					PauseOrder(order, log, char.ID, char.Name)
				}
			}
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

	// No needs could be fulfilled - try an idle activity
	if intent := selectIdleActivity(char, cpos, items, gameMap, log, orders); intent != nil {
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

	// For ActionConsume (eating from inventory), just continue - item is in inventory, not on map
	if intent.Action == entity.ActionConsume {
		return intent
	}

	// For ActionDrink with TargetItem in character inventory (carried vessel), just continue
	if intent.Action == entity.ActionDrink && intent.TargetItem != nil {
		if char.FindInInventory(func(item *entity.Item) bool { return item == intent.TargetItem }) != nil {
			return intent
		}
	}

	// ActionForage has two phases:
	// Phase 1 (optional): TargetItem is a vessel on the ground — move toward it for pickup
	// Phase 2: TargetItem is food on the ground — move toward it for pickup
	// Both phases use the generic TargetItem-on-map check below, so no special handling needed.

	// ActionFillVessel has two phases:
	// Phase 1: TargetItem is on the map (ground vessel) — move toward it for pickup
	// Phase 2: TargetItem is in inventory — move toward water Dest for filling
	if intent.Action == entity.ActionFillVessel {
		if intent.TargetItem != nil {
			ipos := intent.TargetItem.Pos()
			if gameMap.ItemAt(ipos) == intent.TargetItem {
				// Phase 1: vessel is on the ground — recalculate toward it
				if cx == ipos.X && cy == ipos.Y {
					return intent // At vessel, ready for pickup
				}
				nx, ny := NextStepBFS(cx, cy, ipos.X, ipos.Y, gameMap)
				intent.Target = types.Position{X: nx, Y: ny}
				return intent
			}
			// Check if vessel is in character's inventory (phase 2)
			inInventory := false
			for _, item := range char.Inventory {
				if item == intent.TargetItem {
					inInventory = true
					break
				}
			}
			if !inInventory {
				return nil // Vessel was taken by someone else
			}
		}
		// Phase 2: vessel in inventory — recalculate toward water Dest
		dest := intent.Dest
		if cx == dest.X && cy == dest.Y {
			return intent // At water destination, ready to fill
		}
		nx, ny := NextStepBFS(cx, cy, dest.X, dest.Y, gameMap)
		intent.Target = types.Position{X: nx, Y: ny}
		return intent
	}

	// ActionWaterGarden has three phases:
	// Phase 1: TargetItem is on the map (ground vessel) — move toward it for pickup
	// Phase 2: TargetItem is in inventory (empty vessel) — move toward water Dest for filling
	// Phase 3: TargetItem is in inventory (vessel with water) — move toward dry tile Dest for watering
	if intent.Action == entity.ActionWaterGarden {
		if intent.TargetItem != nil {
			ipos := intent.TargetItem.Pos()
			if gameMap.ItemAt(ipos) == intent.TargetItem {
				// Phase 1: vessel is on the ground — recalculate toward it
				if cx == ipos.X && cy == ipos.Y {
					return intent // At vessel, ready for pickup
				}
				nx, ny := NextStepBFS(cx, cy, ipos.X, ipos.Y, gameMap)
				intent.Target = types.Position{X: nx, Y: ny}
				return intent
			}
			// Check if vessel is in character's inventory (phase 2 or 3)
			inInventory := false
			for _, item := range char.Inventory {
				if item == intent.TargetItem {
					inInventory = true
					break
				}
			}
			if !inInventory {
				return nil // Vessel was taken by someone else
			}
		}
		// Phase 2/3: vessel in inventory — recalculate toward Dest
		dest := intent.Dest
		if cx == dest.X && cy == dest.Y {
			return intent // At destination, ready for fill or water
		}
		nx, ny := NextStepBFS(cx, cy, dest.X, dest.Y, gameMap)
		intent.Target = types.Position{X: nx, Y: ny}
		return intent
	}

	// Check if target item still exists at expected position (O(1) instead of O(n))
	if intent.TargetItem != nil {
		ipos := intent.TargetItem.Pos()
		if gameMap.ItemAt(ipos) != intent.TargetItem {
			return nil // Target consumed by someone else
		}

		// If adjacent to target item and tile is occupied, abandon to find alternative
		// (matches bed/water occupied pattern — character can't reach item)
		if isAdjacent(cx, cy, ipos.X, ipos.Y) {
			occupant := gameMap.CharacterAt(ipos)
			if occupant != nil && occupant != char {
				return nil
			}
		}
	}

	// Check if target water tile is still available (for drinking)
	if intent.TargetWaterPos != nil {
		wpos := *intent.TargetWaterPos
		hasAvailableTile := false
		cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
		for _, dir := range cardinalDirs {
			adjPos := types.Position{X: wpos.X + dir[0], Y: wpos.Y + dir[1]}
			if !gameMap.IsValid(adjPos) {
				continue
			}
			occupant := gameMap.CharacterAt(adjPos)
			if occupant == nil || occupant == char {
				if adjFeature := gameMap.FeatureAt(adjPos); adjFeature != nil && !adjFeature.IsPassable() {
					continue
				}
				if gameMap.IsWater(adjPos) {
					continue
				}
				hasAvailableTile = true
				break
			}
		}
		if !hasAvailableTile {
			return nil // All cardinal tiles blocked, find new water
		}
	}

	// Check if target feature is still available (for beds)
	if intent.TargetFeature != nil {
		feature := intent.TargetFeature
		fpos := feature.Pos()
		// For passable features (beds), check if occupied by another character
		occupant := gameMap.CharacterAt(fpos)
		if occupant != nil && occupant != char {
			return nil // Target occupied by someone else, find new target
		}
	}

	// Recalculate next step toward target
	var tx, ty int
	if intent.TargetItem != nil {
		tpos := intent.TargetItem.Pos()
		tx, ty = tpos.X, tpos.Y
	} else if intent.TargetWaterPos != nil {
		tx, ty = intent.TargetWaterPos.X, intent.TargetWaterPos.Y
	} else if intent.TargetFeature != nil {
		tpos := intent.TargetFeature.Pos()
		tx, ty = tpos.X, tpos.Y
	} else if intent.TargetCharacter != nil {
		tpos := intent.TargetCharacter.Pos()
		tx, ty = tpos.X, tpos.Y
	} else {
		tx, ty = intent.Target.X, intent.Target.Y
	}

	// Check if we've arrived at a ground vessel for drinking - switch to drink action
	if intent.TargetItem != nil && intent.DrivingStat == types.StatThirst && cx == tx && cy == ty {
		newActivity := "Drinking"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "thirst", "Drinking from vessel")
			}
		}
		return &entity.Intent{
			Target:      types.Position{X: cx, Y: cy},
			Dest:        types.Position{X: cx, Y: cy},
			Action:      entity.ActionDrink,
			TargetItem:  intent.TargetItem,
			DrivingStat: intent.DrivingStat,
			DrivingTier: intent.DrivingTier,
		}
	}

	// Check if we've arrived at a water target - switch to drink action
	if intent.TargetWaterPos != nil {
		if isCardinallyAdjacent(cx, cy, tx, ty) {
			newActivity := "Drinking"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
				if log != nil {
					log.Add(char.ID, char.Name, "thirst",
						fmt.Sprintf("Drinking from %s", waterSourceName(gameMap, *intent.TargetWaterPos)))
				}
			}
			return &entity.Intent{
				Target:         types.Position{X: cx, Y: cy}, // Stay in place
				Dest:           types.Position{X: cx, Y: cy}, // Already at destination
				Action:         entity.ActionDrink,
				TargetWaterPos: intent.TargetWaterPos,
				DrivingStat:    intent.DrivingStat,
				DrivingTier:    intent.DrivingTier,
			}
		}
		// Not adjacent yet - check if any cardinal tile is still available
		adjX, adjY := FindClosestCardinalTile(cx, cy, tx, ty, gameMap)
		if adjX == -1 {
			return nil // All cardinal tiles blocked, find new water
		}
		tx, ty = adjX, adjY
	}

	// Check if we've arrived at a feature target - switch to appropriate action
	if intent.TargetFeature != nil {
		feature := intent.TargetFeature

		// Beds are passable - arrive when at the feature position
		if feature.IsBed() && cx == tx && cy == ty {
			newActivity := "Sleeping (in bed)"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				Target:        types.Position{X: tx, Y: ty},
				Dest:          types.Position{X: tx, Y: ty}, // Destination is the bed
				Action:        entity.ActionSleep,
				TargetFeature: feature,
				DrivingStat:   intent.DrivingStat,
				DrivingTier:   intent.DrivingTier,
			}
		}
	}

	// Check if we've arrived adjacent to an item for looking (no DrivingStat means looking intent)
	// Skip if ActionPickup/ActionForage/ActionWaterGarden - those need to move onto the item, not stay adjacent
	if intent.TargetItem != nil && intent.DrivingStat == "" && intent.Action != entity.ActionPickup && intent.Action != entity.ActionForage && intent.Action != entity.ActionWaterGarden && isAdjacent(cx, cy, tx, ty) {
		newActivity := "Looking at " + intent.TargetItem.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:          types.Position{X: cx, Y: cy}, // Stay in place
			Dest:            types.Position{X: cx, Y: cy}, // Already at destination (adjacent to item)
			Action:     entity.ActionLook,
			TargetItem: intent.TargetItem,
		}
	}

	// Check if we've arrived adjacent to a character for talking
	if intent.TargetCharacter != nil && isAdjacent(cx, cy, tx, ty) {
		return &entity.Intent{
			Target:          types.Position{X: cx, Y: cy}, // Stay in place
			Dest:            types.Position{X: cx, Y: cy}, // Already at destination (adjacent to character)
			Action:          entity.ActionTalk,
			TargetCharacter: intent.TargetCharacter,
		}
	}

	nx, ny := NextStepBFS(cx, cy, tx, ty, gameMap)

	return &entity.Intent{
		Target:          types.Position{X: nx, Y: ny},
		Dest:            types.Position{X: tx, Y: ty}, // Destination we're moving toward
		Action:          intent.Action,
		TargetItem:      intent.TargetItem,
		TargetFeature:   intent.TargetFeature,
		TargetWaterPos:  intent.TargetWaterPos,
		TargetCharacter: intent.TargetCharacter,
		DrivingStat:     intent.DrivingStat,
		DrivingTier:     intent.DrivingTier,
	}
}

// vesselHasLiquid returns true if the item is a vessel containing liquid
func vesselHasLiquid(item *entity.Item) bool {
	return item != nil && item.Container != nil &&
		len(item.Container.Contents) > 0 &&
		item.Container.Contents[0].Variety != nil &&
		item.Container.Contents[0].Variety.ItemType == "liquid"
}

// waterSourceName returns a human-readable name for a water terrain type
func waterSourceName(gameMap *game.Map, waterPos types.Position) string {
	switch gameMap.WaterAt(waterPos) {
	case game.WaterSpring:
		return "spring"
	case game.WaterPond:
		return "pond"
	default:
		return "water"
	}
}

// findDrinkIntent finds the closest water source to drink from.
// Searches three source types by distance: carried water vessel (distance 0),
// ground water vessel (distance to vessel), water terrain (distance to cardinal-adjacent tile).
func findDrinkIntent(char *entity.Character, pos types.Position, gameMap *game.Map, tier int, log *ActionLog, items []*entity.Item) *entity.Intent {
	// Track best source across all types
	type drinkSource struct {
		distance int
		kind     string // "carried", "ground", "terrain"
		vessel   *entity.Item
		waterPos types.Position
		adjX     int // cardinal-adjacent tile for terrain
		adjY     int
	}
	var best *drinkSource

	// 1. Check carried water vessel (distance 0)
	for _, item := range char.Inventory {
		if vesselHasLiquid(item) {
			best = &drinkSource{distance: 0, kind: "carried", vessel: item}
			break // Can't beat distance 0
		}
	}

	// 2. Check ground water vessels (only if carried vessel not found)
	if best == nil {
		for _, item := range items {
			if !vesselHasLiquid(item) {
				continue
			}
			// Skip vessels in character inventories (only ground vessels)
			if gameMap.ItemAt(item.Pos()) != item {
				continue
			}
			dist := pos.DistanceTo(item.Pos())
			if best == nil || dist < best.distance {
				best = &drinkSource{distance: dist, kind: "ground", vessel: item}
			}
		}
	}

	// 3. Check water terrain
	waterPos, found := gameMap.FindNearestWater(pos)
	if found {
		// Distance to terrain = distance to the cardinal-adjacent tile the character would stand on
		adjX, adjY := FindClosestCardinalTile(pos.X, pos.Y, waterPos.X, waterPos.Y, gameMap)
		if adjX != -1 {
			terrainDist := pos.DistanceTo(types.Position{X: adjX, Y: adjY})
			if best == nil || terrainDist < best.distance {
				best = &drinkSource{distance: terrainDist, kind: "terrain", waterPos: waterPos, adjX: adjX, adjY: adjY}
			}
		}
	}

	// No water source found
	if best == nil {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no water source)")
			}
		}
		return nil
	}

	// Build intent based on source type
	switch best.kind {
	case "carried":
		// Distance 0 — drink immediately from carried vessel
		newActivity := "Drinking"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "thirst", "Drinking from vessel")
			}
		}
		return &entity.Intent{
			Target:      pos,
			Dest:        pos,
			Action:      entity.ActionDrink,
			TargetItem:  best.vessel,
			DrivingStat: types.StatThirst,
			DrivingTier: tier,
		}

	case "ground":
		vpos := best.vessel.Pos()
		if pos.X == vpos.X && pos.Y == vpos.Y {
			// Already at ground vessel — drink in place
			newActivity := "Drinking"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
				if log != nil {
					log.Add(char.ID, char.Name, "thirst", "Drinking from vessel")
				}
			}
			return &entity.Intent{
				Target:      pos,
				Dest:        pos,
				Action:      entity.ActionDrink,
				TargetItem:  best.vessel,
				DrivingStat: types.StatThirst,
				DrivingTier: tier,
			}
		}
		// Move toward ground vessel
		nx, ny := NextStepBFS(pos.X, pos.Y, vpos.X, vpos.Y, gameMap)
		newActivity := "Moving to water"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "movement", "Heading to water")
			}
		}
		return &entity.Intent{
			Target:      types.Position{X: nx, Y: ny},
			Dest:        vpos,
			Action:      entity.ActionMove,
			TargetItem:  best.vessel,
			DrivingStat: types.StatThirst,
			DrivingTier: tier,
		}

	case "terrain":
		wp := best.waterPos
		if isCardinallyAdjacent(pos.X, pos.Y, wp.X, wp.Y) {
			// Already cardinally adjacent — drink from terrain
			newActivity := "Drinking"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
				if log != nil {
					log.Add(char.ID, char.Name, "thirst",
						fmt.Sprintf("Drinking from %s", waterSourceName(gameMap, wp)))
				}
			}
			return &entity.Intent{
				Target:         pos,
				Dest:           pos,
				Action:         entity.ActionDrink,
				TargetWaterPos: &wp,
				DrivingStat:    types.StatThirst,
				DrivingTier:    tier,
			}
		}
		// Move toward cardinal-adjacent tile
		nx, ny := NextStepBFS(pos.X, pos.Y, best.adjX, best.adjY, gameMap)
		newActivity := "Moving to water"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "movement", "Heading to water")
			}
		}
		return &entity.Intent{
			Target:         types.Position{X: nx, Y: ny},
			Dest:           types.Position{X: best.adjX, Y: best.adjY},
			Action:         entity.ActionMove,
			TargetWaterPos: &wp,
			DrivingStat:    types.StatThirst,
			DrivingTier:    tier,
		}
	}

	return nil
}

// findFoodIntent finds food based on hunger priority
// Uses unified scoring for both carried and map items (carried items have distance=0)
func findFoodIntent(char *entity.Character, pos types.Position, items []*entity.Item, tier int, log *ActionLog, gameMap *game.Map) *entity.Intent {
	result := FindFoodTarget(char, items)
	if result.Item == nil {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no suitable food)")
			}
		}
		return nil
	}

	// Check if best food is in inventory (any slot)
	isInInventory := char.FindInInventory(func(i *entity.Item) bool { return i == result.Item }) != nil
	if isInInventory {
		newActivity := "Eating carried " + result.Item.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity",
					fmt.Sprintf("Eating from inventory (pref:%d score:%.0f)",
						result.NetPreference, result.GradientScore))
			}
		}
		return &entity.Intent{
			Target:      pos,
			Dest:        pos, // Already at destination (eating from inventory)
			Action:      entity.ActionConsume,
			TargetItem:  result.Item,
			DrivingStat: types.StatHunger,
			DrivingTier: tier,
		}
	}

	// Best food is on map - move to it
	ipos := result.Item.Pos()
	tx, ty := ipos.X, ipos.Y
	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)

	newActivity := "Moving to " + result.Item.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement",
				fmt.Sprintf("Started moving to %s (pref:%d score:%.0f)",
					result.Item.Description(), result.NetPreference, result.GradientScore))
		}
	}

	return &entity.Intent{
		Target:      types.Position{X: nx, Y: ny},
		Dest:        types.Position{X: tx, Y: ty}, // Destination is the item's position
		Action:      entity.ActionMove,
		TargetItem:  result.Item,
		DrivingStat: types.StatHunger,
		DrivingTier: tier,
	}
}

// findHealingIntent finds a known healing item to consume.
// Only considers items the character knows are healing.
// Returns nil if no known healing items are available.
func findHealingIntent(char *entity.Character, pos types.Position, items []*entity.Item, tier int, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Get only items the character knows are healing
	knownHealing := char.KnownHealingItems(items)
	if len(knownHealing) == 0 {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no known healing items)")
			}
		}
		return nil
	}

	// Find nearest known healing item
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range knownHealing {
		ipos := item.Pos()
		dist := pos.DistanceTo(ipos)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	if nearest == nil {
		return nil
	}

	npos := nearest.Pos()
	tx, ty := npos.X, npos.Y
	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)

	newActivity := "Moving to " + nearest.Description() + " (healing)"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement",
				fmt.Sprintf("Seeking healing: %s", nearest.Description()))
		}
	}

	return &entity.Intent{
		Target:      types.Position{X: nx, Y: ny},
		Dest:        types.Position{X: tx, Y: ty}, // Destination is the item's position
		Action:      entity.ActionMove,
		TargetItem:  nearest,
		DrivingStat: types.StatHealth,
		DrivingTier: tier,
	}
}

// findSleepIntent finds a bed to sleep in
func findSleepIntent(char *entity.Character, pos types.Position, gameMap *game.Map, tier int, log *ActionLog) *entity.Intent {
	bed := gameMap.FindNearestBed(pos)

	// If no bed, can sleep on ground when exhausted (voluntary) or collapsed (involuntary)
	if bed == nil {
		if char.Energy <= 10 { // Exhausted - ground sleep available
			newActivity := "Sleeping (on ground)"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				Target:      pos,
				Dest:        pos, // Already at destination (ground sleep)
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

	bpos := bed.Pos()
	tx, ty := bpos.X, bpos.Y

	// Already at bed - sleep
	if pos.X == tx && pos.Y == ty {
		newActivity := "Sleeping (in bed)"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:        types.Position{X: tx, Y: ty},
			Dest:          types.Position{X: tx, Y: ty}, // Already at destination (the bed)
			Action:        entity.ActionSleep,
			TargetFeature: bed,
			DrivingStat:   types.StatEnergy,
			DrivingTier:   tier,
		}
	}

	// Move toward bed
	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)
	newActivity := "Moving to leaf pile"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Heading to leaf pile")
		}
	}

	return &entity.Intent{
		Target:        types.Position{X: nx, Y: ny},
		Dest:          types.Position{X: tx, Y: ty}, // Destination is the bed
		Action:        entity.ActionMove,
		TargetFeature: bed,
		DrivingStat:   types.StatEnergy,
		DrivingTier:   tier,
	}
}

// FoodTargetResult contains the selected food target and scoring info
type FoodTargetResult struct {
	Item          *entity.Item
	NetPreference int
	GradientScore float64
}

// FindFoodTarget finds the best food for a character based on hunger level
// Uses gradient scoring: Score = (NetPreference × PrefWeight) - (Distance × DistWeight) + HealingBonus
// Considers both carried items (distance=0) and map items using the same scoring.
// Hunger tier affects both the preference weight and which items are considered:
// - Moderate (50-74): High pref weight, only NetPreference >= 0 items considered
// - Severe (75-89): Medium pref weight, all items considered
// - Crisis (90+): No pref weight (just distance), all items considered
// Healing bonus: When health tier >= Moderate and character knows item is healing,
// adds bonus to score (larger bonus at worse health tiers)
func FindFoodTarget(char *entity.Character, items []*entity.Item) FoodTargetResult {
	cpos := char.Pos()

	// Determine hunger tier and corresponding weights/filters
	var prefWeight float64
	var filterDisliked bool

	if char.Hunger >= 90 {
		// Crisis: just pick nearest
		prefWeight = config.FoodSeekPrefWeightCrisis
		filterDisliked = false
	} else if char.Hunger >= 75 {
		// Severe: gradient with medium pref weight, consider all items
		prefWeight = config.FoodSeekPrefWeightSevere
		filterDisliked = false
	} else {
		// Moderate: gradient with high pref weight, filter disliked
		prefWeight = config.FoodSeekPrefWeightModerate
		filterDisliked = true
	}

	// Determine healing bonus based on health tier (only if hurt)
	var healingBonus float64
	healthTier := char.HealthTier()
	switch healthTier {
	case entity.TierCrisis:
		healingBonus = config.HealingBonusCrisis
	case entity.TierSevere:
		healingBonus = config.HealingBonusSevere
	case entity.TierModerate:
		healingBonus = config.HealingBonusModerate
	case entity.TierMild:
		healingBonus = config.HealingBonusMild
	default:
		healingBonus = 0 // No bonus at TierNone (full health)
	}

	var bestItem *entity.Item
	var bestNetPref int
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float
	bestDist := int(^uint(0) >> 1)              // Max int for distance tiebreaker

	// Helper to score and potentially update best candidate
	scoreCandidate := func(item *entity.Item, dist int) {
		if !item.IsEdible() {
			return
		}

		netPref := char.NetPreference(item)

		// At Moderate hunger, filter out disliked items (NetPreference < 0)
		if filterDisliked && netPref < 0 {
			return
		}

		// Calculate gradient score
		score := float64(netPref)*prefWeight - float64(dist)*config.FoodSeekDistWeight

		// Apply healing bonus if character knows this item is healing
		if healingBonus > 0 && char.KnowsItemIsHealing(item) {
			score += healingBonus
		}

		// Update best if better score, or same score but closer (distance tiebreaker)
		if score > bestScore || (score == bestScore && dist < bestDist) {
			bestItem = item
			bestNetPref = netPref
			bestScore = score
			bestDist = dist
		}
	}

	// Score inventory items first (distance = 0)
	for _, invItem := range char.Inventory {
		if invItem == nil {
			continue
		}
		// Check if carrying a vessel with edible contents
		if invItem.Container != nil && len(invItem.Container.Contents) > 0 {
			variety := invItem.Container.Contents[0].Variety
			if variety.IsEdible() {
				netPref := char.NetPreferenceForVariety(variety)

				// At Moderate hunger, filter out disliked items
				if !filterDisliked || netPref >= 0 {
					score := float64(netPref)*prefWeight - 0*config.FoodSeekDistWeight // distance = 0

					// Apply healing bonus if character knows this variety is healing
					if healingBonus > 0 && char.KnowsVarietyIsHealing(variety) {
						score += healingBonus
					}

					if score > bestScore || (score == bestScore && 0 < bestDist) {
						bestItem = invItem // Return vessel, not the variety
						bestNetPref = netPref
						bestScore = score
						bestDist = 0
					}
				}
			}
		} else {
			// Carrying a loose item - score it directly
			scoreCandidate(invItem, 0)
		}
	}

	// Score map items (including dropped vessels with edible contents)
	for _, item := range items {
		ipos := item.Pos()
		dist := cpos.DistanceTo(ipos)

		// Check if item is a vessel with edible contents
		if item.Container != nil && len(item.Container.Contents) > 0 {
			variety := item.Container.Contents[0].Variety
			if variety.IsEdible() {
				netPref := char.NetPreferenceForVariety(variety)

				// At Moderate hunger, filter out disliked items
				if !filterDisliked || netPref >= 0 {
					score := float64(netPref)*prefWeight - float64(dist)*config.FoodSeekDistWeight

					// Apply healing bonus if character knows this variety is healing
					if healingBonus > 0 && char.KnowsVarietyIsHealing(variety) {
						score += healingBonus
					}

					if score > bestScore || (score == bestScore && dist < bestDist) {
						bestItem = item // Return vessel
						bestNetPref = netPref
						bestScore = score
						bestDist = dist
					}
				}
			}
		} else {
			// Regular edible item
			scoreCandidate(item, dist)
		}
	}

	return FoodTargetResult{
		Item:          bestItem,
		NetPreference: bestNetPref,
		GradientScore: bestScore,
	}
}

// NextStepBFS calculates the next position moving toward target using BFS pathfinding.
// Routes around permanent obstacles (water, impassable features). Ignores characters
// since they move and per-tick collision is handled separately.
// Falls back to greedy NextStep if no path exists.
func NextStepBFS(fromX, fromY, toX, toY int, gameMap *game.Map) (int, int) {
	if fromX == toX && fromY == toY {
		return fromX, fromY
	}

	// Nil map fallback - used in tests that don't need pathfinding
	if gameMap == nil {
		return NextStep(fromX, fromY, toX, toY)
	}

	from := types.Position{X: fromX, Y: fromY}
	to := types.Position{X: toX, Y: toY}

	// BFS tracking the first step from origin that leads to each visited tile
	type node struct {
		pos       types.Position
		firstStep types.Position
	}

	visited := make(map[types.Position]bool)
	visited[from] = true

	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	// Seed queue with walkable neighbors of start
	var queue []node
	for _, dir := range cardinalDirs {
		neighbor := types.Position{X: fromX + dir[0], Y: fromY + dir[1]}
		if !gameMap.IsValid(neighbor) || visited[neighbor] {
			continue
		}
		if gameMap.IsWater(neighbor) {
			continue
		}
		if f := gameMap.FeatureAt(neighbor); f != nil && !f.IsPassable() {
			continue
		}
		visited[neighbor] = true
		if neighbor == to {
			return neighbor.X, neighbor.Y
		}
		queue = append(queue, node{pos: neighbor, firstStep: neighbor})
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, dir := range cardinalDirs {
			neighbor := types.Position{X: cur.pos.X + dir[0], Y: cur.pos.Y + dir[1]}
			if !gameMap.IsValid(neighbor) || visited[neighbor] {
				continue
			}
			if gameMap.IsWater(neighbor) {
				continue
			}
			if f := gameMap.FeatureAt(neighbor); f != nil && !f.IsPassable() {
				continue
			}
			visited[neighbor] = true
			if neighbor == to {
				return cur.firstStep.X, cur.firstStep.Y
			}
			queue = append(queue, node{pos: neighbor, firstStep: cur.firstStep})
		}
	}

	// No path found - fall back to greedy
	return NextStep(fromX, fromY, toX, toY)
}

// NextStep calculates the next position moving toward target
func NextStep(fromX, fromY, toX, toY int) (int, int) {
	dx := toX - fromX
	dy := toY - fromY

	if dx == 0 && dy == 0 {
		return fromX, fromY
	}

	// Move toward larger distance
	if types.Abs(dx) > types.Abs(dy) {
		return fromX + types.Sign(dx), fromY
	}
	return fromX, fromY + types.Sign(dy)
}

// canFulfillThirst checks if thirst can be addressed (water source exists).
// Checks carried water vessel, ground water vessels, and water terrain.
func canFulfillThirst(char *entity.Character, gameMap *game.Map, pos types.Position, items []*entity.Item) bool {
	// Carried water vessel
	for _, item := range char.Inventory {
		if vesselHasLiquid(item) {
			return true
		}
	}
	// Ground water vessels
	for _, item := range items {
		if vesselHasLiquid(item) && gameMap.ItemAt(item.Pos()) == item {
			return true
		}
	}
	// Water terrain
	_, found := gameMap.FindNearestWater(pos)
	return found
}

// canFulfillHunger checks if hunger can be addressed (suitable food exists)
func canFulfillHunger(char *entity.Character, items []*entity.Item) bool {
	return FindFoodTarget(char, items).Item != nil
}

// canFulfillEnergy checks if energy can be addressed (bed exists or exhausted enough for ground sleep)
func canFulfillEnergy(char *entity.Character, gameMap *game.Map, pos types.Position) bool {
	// Can sleep on ground if exhausted (voluntary at ≤10, involuntary collapse at 0)
	if char.Energy <= 10 {
		return true
	}
	// Otherwise need a bed
	return gameMap.FindNearestBed(pos) != nil
}

// canFulfillHealth checks if health can be addressed (character knows healing items and they exist)
func canFulfillHealth(char *entity.Character, items []*entity.Item) bool {
	return len(char.KnownHealingItems(items)) > 0
}

// findLookIntent creates an intent to look at the nearest item.
// Called by selectIdleActivity when looking is selected.
func findLookIntent(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Find nearest item, excluding last looked item
	target := findNearestItemExcluding(pos.X, pos.Y, items, char.LastLookedX, char.LastLookedY, char.HasLastLooked)
	if target == nil {
		return nil
	}

	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	// Check if already adjacent to target
	if isAdjacent(pos.X, pos.Y, tx, ty) {
		// Start looking immediately
		newActivity := "Looking at " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Looking at "+target.Description())
			}
		}
		return &entity.Intent{
			Target:     pos, // Stay in place
			Dest:       pos, // Already at destination (adjacent to item)
			Action:     entity.ActionLook,
			TargetItem: target,
		}
	}

	// Find closest adjacent tile to target
	adjX, adjY := findClosestAdjacentTile(pos.X, pos.Y, tx, ty, gameMap)
	if adjX == -1 {
		return nil // No accessible adjacent tile
	}

	// Move toward adjacent tile
	nx, ny := NextStepBFS(pos.X, pos.Y, adjX, adjY, gameMap)

	newActivity := "Moving to look at " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Moving to look at "+target.Description())
		}
	}

	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: adjX, Y: adjY}, // Destination is adjacent to the item
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

	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1)

	for _, item := range items {
		ipos := item.Pos()

		// Skip excluded position
		if hasExclude && ipos.X == excludeX && ipos.Y == excludeY {
			continue
		}

		dist := pos.DistanceTo(ipos)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	return nearest
}

// isAdjacent checks if two positions are adjacent (including diagonals)
func isAdjacent(x1, y1, x2, y2 int) bool {
	return types.Position{X: x1, Y: y1}.IsAdjacentTo(types.Position{X: x2, Y: y2})
}

// isCardinallyAdjacent checks 4-direction adjacency (N/E/S/W, no diagonals)
func isCardinallyAdjacent(x1, y1, x2, y2 int) bool {
	return types.Position{X: x1, Y: y1}.IsCardinallyAdjacentTo(types.Position{X: x2, Y: y2})
}

// FindClosestCardinalTile finds closest unblocked cardinally adjacent tile to target
func FindClosestCardinalTile(cx, cy, tx, ty int, gameMap *game.Map) (int, int) {
	pos := types.Position{X: cx, Y: cy}
	directions := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	bestX, bestY := -1, -1
	bestDist := int(^uint(0) >> 1)

	for _, dir := range directions {
		adjPos := types.Position{X: tx + dir[0], Y: ty + dir[1]}
		if !gameMap.IsValid(adjPos) || gameMap.IsBlocked(adjPos) {
			continue
		}
		if dist := pos.DistanceTo(adjPos); dist < bestDist {
			bestDist, bestX, bestY = dist, adjPos.X, adjPos.Y
		}
	}
	return bestX, bestY
}

// findClosestAdjacentTile finds the closest unoccupied tile adjacent to (tx, ty) from position (cx, cy)
func findClosestAdjacentTile(cx, cy, tx, ty int, gameMap *game.Map) (int, int) {
	pos := types.Position{X: cx, Y: cy}
	// 8 directions
	directions := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}

	bestX, bestY := -1, -1
	bestDist := int(^uint(0) >> 1)

	for _, dir := range directions {
		adjPos := types.Position{X: tx + dir[0], Y: ty + dir[1]}
		if !gameMap.IsValid(adjPos) {
			continue
		}
		if gameMap.IsOccupied(adjPos) {
			continue
		}

		dist := pos.DistanceTo(adjPos)
		if dist < bestDist {
			bestDist = dist
			bestX, bestY = adjPos.X, adjPos.Y
		}
	}

	return bestX, bestY
}
