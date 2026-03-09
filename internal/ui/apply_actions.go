package ui

import (
	"fmt"
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/system"
	"petri/internal/types"
)

// applyIntent executes a character's intent by dispatching to the appropriate handler.
func (m *Model) applyIntent(char *entity.Character, delta float64) {
	if char.Intent == nil || char.IsDead || char.IsSleeping {
		return
	}

	// Collapse is immediate and involuntary - check before any action
	if char.Energy <= 0 {
		system.StartSleep(char, false, m.actionLog)
		return
	}

	switch char.Intent.Action {
	case entity.ActionMove:
		m.applyMove(char, delta)
	case entity.ActionDrink:
		m.applyDrink(char, delta)
	case entity.ActionSleep:
		m.applySleep(char, delta)
	case entity.ActionLook:
		m.applyLook(char, delta)
	case entity.ActionTalk:
		m.applyTalk(char, delta)
	case entity.ActionPickup:
		m.applyPickup(char, delta)
	case entity.ActionConsume:
		m.applyConsume(char, delta)
	case entity.ActionCraft:
		m.applyCraft(char, delta)
	case entity.ActionTillSoil:
		m.applyTillSoil(char, delta)
	case entity.ActionPlant:
		m.applyPlant(char, delta)
	case entity.ActionForage:
		m.applyForage(char, delta)
	case entity.ActionFillVessel:
		m.applyFillVessel(char, delta)
	case entity.ActionWaterGarden:
		m.applyWaterGarden(char, delta)
	case entity.ActionHelpFeed:
		m.applyHelpFeed(char, delta)
	case entity.ActionHelpWater:
		m.applyHelpWater(char, delta)
	case entity.ActionExtract:
		m.applyExtract(char, delta)
	case entity.ActionDig:
		m.applyDig(char, delta)
	case entity.ActionBuildFence:
		m.applyBuildFence(char, delta)
	case entity.ActionBuildHut:
		m.applyBuildHut(char, delta)
	}
}

// applyMove handles ActionMove: speed-gated movement with displacement-aware collision
// handling and energy drain. Also detects arrival at ground food for in-place eating.
func (m *Model) applyMove(char *entity.Character, delta float64) {
	cpos := char.Pos()
	cx, cy := cpos.X, cpos.Y
	tx, ty := char.Intent.Target.X, char.Intent.Target.Y

	// Check if at target item for eating - only if driven by hunger
	// Item can be edible directly OR be a vessel with edible contents
	if char.Intent.TargetItem != nil && char.Intent.DrivingStat == types.StatHunger {
		targetItem := char.Intent.TargetItem
		isEdible := targetItem.IsEdible()
		isVesselWithFood := targetItem.Container != nil && len(targetItem.Container.Contents) > 0

		if isEdible || isVesselWithFood {
			ipos := targetItem.Pos()
			if cx == ipos.X && cy == ipos.Y {
				// At target item - eating in progress, duration varies by food tier
				// Update activity from "Moving to X" to "Eating X"
				if isVesselWithFood {
					char.CurrentActivity = "Eating " + targetItem.Container.Contents[0].Variety.Description() + " from vessel"
				} else {
					char.CurrentActivity = "Eating " + targetItem.Description()
				}
				duration := config.GetMealSize(getEatenItemType(targetItem)).Duration
				char.ActionProgress += delta
				if char.ActionProgress >= duration {
					char.ActionProgress = 0
					if m.gameMap.HasItemOnMap(targetItem) {
						if isVesselWithFood {
							// Eat from vessel contents (vessel stays on map)
							system.ConsumeFromVessel(char, targetItem, m.gameMap, m.actionLog)
						} else {
							// Eat the item directly (removes from map)
							system.Consume(char, targetItem, m.gameMap, m.actionLog)
						}
					}
				}
				return
			}
		}
	}

	// Movement is gated by speed accumulator
	speed := char.EffectiveSpeed()
	char.SpeedAccumulator += float64(speed) * delta

	// Check if we've accumulated enough "movement points" to act
	// At speed 50 (baseline), character moves once per 0.02 seconds (50 ticks/sec)
	// This is scaled so baseline speed = ~1 action per game tick (0.15s)
	const movementThreshold = 7.5 // 50 speed * 0.15s delta = 7.5

	if char.SpeedAccumulator < movementThreshold {
		return
	}

	// Consume accumulated points
	char.SpeedAccumulator -= movementThreshold

	// Try to move with displacement-aware collision handling
	moved := false
	if char.DisplacementStepsLeft > 0 {
		// Continue displacement: move perpendicular instead of following path
		moved = m.takeDisplacementStep(char, cx, cy)
	} else {
		// Normal movement with character-collision displacement
		triedPositions := map[[2]int]bool{{tx, ty}: true}
		if m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
			moved = true
		} else {
			// Character collision → initiate perpendicular displacement
			if m.gameMap.CharacterAt(types.Position{X: tx, Y: ty}) != nil {
				moved = m.initiateDisplacement(char, cx, cy, tx, ty)
			}
			if !moved {
				// Non-character obstacle or displacement unavailable → findAlternateStep
				for attempts := 0; attempts < 5 && !moved; attempts++ {
					altStep := m.findAlternateStep(char, cx, cy, triedPositions)
					if altStep == nil {
						break
					}
					tx, ty = altStep[0], altStep[1]
					triedPositions[[2]int{tx, ty}] = true
					if m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
						moved = true
					}
				}
			}
		}
	}

	if !moved {
		// Couldn't move anywhere this tick, refund some speed points
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
		if !char.IsSleeping && m.actionLog != nil {
			if prevEnergy > 50 && char.Energy <= 50 {
				m.actionLog.Add(char.ID, char.Name, "energy", "Getting tired")
			}
			if prevEnergy > 25 && char.Energy <= 25 {
				m.actionLog.Add(char.ID, char.Name, "energy", "Very tired!")
			}
			if prevEnergy > 10 && char.Energy <= 10 {
				m.actionLog.Add(char.ID, char.Name, "energy", "Exhausted!")
			}
			if prevEnergy > 0 && char.Energy <= 0 {
				m.actionLog.Add(char.ID, char.Name, "energy", "Collapsed from exhaustion!")
			}
		}
	}
}

// applyDrink handles ActionDrink: timed drinking from terrain water or a vessel.
func (m *Model) applyDrink(char *entity.Character, delta float64) {
	// Drinking requires duration to complete
	char.ActionProgress += delta
	if char.ActionProgress >= config.ActionDurationShort {
		char.ActionProgress = 0

		if char.Intent.TargetItem != nil {
			// Vessel drinking: find vessel in inventory or on ground
			vessel := char.FindInInventory(func(item *entity.Item) bool {
				return item == char.Intent.TargetItem
			})
			if vessel == nil {
				// Check ground
				if m.gameMap.HasItemOnMap(char.Intent.TargetItem) {
					vessel = char.Intent.TargetItem
				}
			}
			if vessel != nil {
				system.DrinkFromVessel(vessel)
			}
			system.Drink(char, m.actionLog)
			// Clear intent to force re-evaluation for next drink source
			char.Intent = nil
		} else {
			// Terrain drinking: existing behavior, intent persists
			system.Drink(char, m.actionLog)
		}
	}
}

// applySleep handles ActionSleep: voluntary sleep initiation (bed or ground collapse).
func (m *Model) applySleep(char *entity.Character, delta float64) {
	atBed := char.Intent.TargetFeature != nil && char.Intent.TargetFeature.IsBed()

	// Collapse is immediate (involuntary) - only at Energy 0
	if !atBed && char.Energy <= 0 {
		system.StartSleep(char, false, m.actionLog)
		return
	}

	// Voluntary sleep (bed or ground) requires duration to complete
	char.ActionProgress += delta
	if char.ActionProgress >= config.ActionDurationShort {
		char.ActionProgress = 0
		system.StartSleep(char, atBed, m.actionLog)
	}
}

// applyLook handles ActionLook: walk to target then observe it.
func (m *Model) applyLook(char *entity.Character, delta float64) {
	cpos := char.Pos()

	// Walking phase: not yet adjacent to target
	if char.Intent.TargetConstruct != nil {
		tpos := char.Intent.TargetConstruct.Pos()
		if !cpos.IsAdjacentTo(tpos) {
			m.moveWithCollision(char, cpos, delta)
			return
		}
	} else if char.Intent.TargetItem != nil {
		ipos := char.Intent.TargetItem.Pos()
		if !cpos.IsAdjacentTo(ipos) {
			m.moveWithCollision(char, cpos, delta)
			return
		}
	}

	// Looking phase: already adjacent
	char.ActionProgress += delta
	if char.ActionProgress >= config.LookDuration {
		char.ActionProgress = 0
		if char.Intent.TargetConstruct != nil {
			system.CompleteLookAtConstruct(char, char.Intent.TargetConstruct, m.actionLog)
		} else {
			system.CompleteLook(char, char.Intent.TargetItem, m.actionLog)
		}
		char.Intent = nil
		char.IdleCooldown = config.IdleCooldown
	}
}

// applyTalk handles ActionTalk: initiate conversation and transmit knowledge on completion.
func (m *Model) applyTalk(char *entity.Character, delta float64) {
	target := char.Intent.TargetCharacter
	if target == nil {
		return
	}

	// If not already talking, start the conversation
	if char.TalkingWith == nil {
		system.StartTalking(char, target, m.actionLog)
	}

	// Decrement talk timer
	char.TalkTimer -= delta
	if char.TalkTimer <= 0 {
		// Talk complete - transmit knowledge, then stop talking
		system.TransmitKnowledge(char, target, m.actionLog)
		system.StopTalking(char, target, m.actionLog)
	}
}

// applyPickup handles ActionPickup: timed item pickup used by harvest orders and
// order prerequisites. Handles vessel filling continuation and order completion.
func (m *Model) applyPickup(char *entity.Character, delta float64) {
	// Picking up an item (used by harvest orders and order prerequisites)
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
			if item := char.Intent.TargetItem; item != nil && item.Pos() == (types.Position{X: cx, Y: cy}) && m.gameMap.HasItemOnMap(item) {
				// If on an order and inventory full, drop current item first
				// BUT don't drop if ANY carried vessel can accept the item
				// (If carrying a recipe input, we'd have ActionCraft intent instead)
				if char.AssignedOrderID != 0 && char.IsInventoryFull() {
					// Check ALL vessels for compatibility — not just the first
					canAddToVessel := system.FindCarriedVesselFor(char, item, m.gameMap.Varieties()) != nil
					// Check if target can merge into existing bundle - don't drop
					canMergeBundle := system.CanMergeIntoBundle(char, item)
					if !canAddToVessel && !canMergeBundle {
						system.Drop(char, m.gameMap, m.actionLog)
					}
				}
				result := system.Pickup(char, item, m.gameMap, m.actionLog, m.gameMap.Varieties())

				// Handle vessel filling continuation
				if result == system.PickupToVessel {
					// Continue filling for orders (autonomous foraging uses ActionForage)
					if char.AssignedOrderID != 0 {
						// Determine growing-only filter: harvest picks growing plants, gather picks any ground item
						growingOnly := true
						if order := m.findOrderByID(char.AssignedOrderID); order != nil && order.ActivityID == "gather" {
							growingOnly = false
						}
						// Continue until vessel full
						if nextIntent := system.FindNextVesselTarget(char, cx, cy, m.gameMap.Items(), m.gameMap.Varieties(), m.gameMap, growingOnly); nextIntent != nil {
							char.Intent = nextIntent
							return
						}
						// Vessel full or no more matching targets - complete order
						if order := m.findOrderByID(char.AssignedOrderID); order != nil {
							if order.ActivityID == "harvest" || order.ActivityID == "gather" {
								system.CompleteOrder(char, order, m.actionLog)
							}
						}
					}
					// Order complete or no continuation target - go idle
					char.Intent = nil
					char.IdleCooldown = config.IdleCooldown
					char.CurrentActivity = "Idle"
					return
				}

				// Handle bundle merge continuation (vessel-excluded items like sticks, grass)
				if result == system.PickupToBundle {
					if char.AssignedOrderID != 0 {
						if order := m.findOrderByID(char.AssignedOrderID); order != nil {
							if order.ActivityID == "buildFence" {
								// Procurement pickup — clear intent, findBuildFenceIntent re-evaluates next tick
								char.Intent = nil
								return
							}
							var nextIntent *entity.Intent
							switch order.ActivityID {
							case "harvest":
								nextIntent = system.FindNextHarvestTarget(char, cx, cy, m.gameMap.Items(), order.TargetType, m.gameMap)
							case "gather":
								nextIntent = system.FindNextGatherTarget(char, cx, cy, m.gameMap.Items(), order.TargetType, m.gameMap)
							}
							if nextIntent != nil {
								char.Intent = nextIntent
								return
							}
							// Bundle full or no more targets — drop completed bundle and finish
							system.DropCompletedBundle(char, order, m.gameMap, m.actionLog)
							system.CompleteOrder(char, order, m.actionLog)
						}
					}
					char.Intent = nil
					char.IdleCooldown = config.IdleCooldown
					char.CurrentActivity = "Idle"
					return
				}

				// Handle pickup failure (variety mismatch with carried vessel)
				// This shouldn't happen with proper intent filtering, but handle gracefully
				if result == system.PickupFailed {
					char.Intent = nil
					char.IdleCooldown = config.IdleCooldown
					char.CurrentActivity = "Idle"
					return
				}

				// PickupToInventory - item added to inventory
				// Check for order continuation or completion
				// Craft orders don't complete on pickup - they complete after crafting
				if char.AssignedOrderID != 0 {
					if order := m.findOrderByID(char.AssignedOrderID); order != nil {
						if order.ActivityID == "harvest" && char.GetCarriedVessel() == nil {
							// Harvest: only continue if no vessel (vessel pickup is a prerequisite, not work)
							if nextIntent := system.FindNextHarvestTarget(char, cx, cy, m.gameMap.Items(), order.TargetType, m.gameMap); nextIntent != nil {
								char.Intent = nextIntent
								return
							}
							system.CompleteOrder(char, order, m.actionLog)
						} else if order.ActivityID == "extract" {
							// Extract: vessel pickup is a prerequisite — clear intent
							// so findExtractIntent re-evaluates with vessel in hand
							char.Intent = nil
							return
						} else if order.ActivityID == "gather" {
							// If we just picked up a vessel, that's a prerequisite — not gather work
							if item.Container != nil {
								return
							}
							// Gather: inventory pickup IS the work — continue regardless of vessel
							if nextIntent := system.FindNextGatherTarget(char, cx, cy, m.gameMap.Items(), order.TargetType, m.gameMap); nextIntent != nil {
								char.Intent = nextIntent
								return
							}
							system.DropCompletedBundle(char, order, m.gameMap, m.actionLog)
							system.CompleteOrder(char, order, m.actionLog)
						}
					}
				}
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
	if m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
		// Successfully moved - update intent for next step
		newPos := char.Pos()
		if newPos.X != ipos.X || newPos.Y != ipos.Y {
			// Not at item yet, calculate next step
			nx, ny := system.NextStepBFS(newPos.X, newPos.Y, ipos.X, ipos.Y, m.gameMap)
			char.Intent.Target.X = nx
			char.Intent.Target.Y = ny
		}
	}
}

// applyConsume handles ActionConsume: timed eating from carried food or a ground vessel.
func (m *Model) applyConsume(char *entity.Character, delta float64) {
	// Eating - duration varies by food tier
	targetItem := char.Intent.TargetItem
	duration := config.GetMealSize(getEatenItemType(targetItem)).Duration
	char.ActionProgress += delta
	if char.ActionProgress >= duration {
		char.ActionProgress = 0
		// Check if target item is in inventory
		inInventory := char.FindInInventory(func(i *entity.Item) bool { return i == targetItem }) != nil
		if inInventory {
			// Check if it's a vessel with edible contents
			if targetItem.Container != nil && len(targetItem.Container.Contents) > 0 {
				// Eat from vessel contents
				system.ConsumeFromVessel(char, targetItem, m.gameMap, m.actionLog)
			} else {
				// Eat the carried item directly
				system.ConsumeFromInventory(char, targetItem, m.gameMap, m.actionLog)
			}
		} else if m.gameMap.HasItemOnMap(targetItem) &&
			targetItem.Container != nil && len(targetItem.Container.Contents) > 0 &&
			targetItem.Container.Contents[0].Variety.IsEdible() {
			// Ground food vessel: eat in place without picking up
			system.ConsumeFromVessel(char, targetItem, m.gameMap, m.actionLog)
			// Clear intent after each unit (like vessel drinking) so character re-evaluates
			char.Intent = nil
		}
	}
}

// applyCraft handles ActionCraft: timed crafting using recipe inputs, dispatched by RecipeID.
func (m *Model) applyCraft(char *entity.Character, delta float64) {
	// Crafting - uses recipe duration, dispatches by intent.RecipeID

	recipe := entity.RecipeRegistry[char.Intent.RecipeID]
	if recipe == nil {
		char.Intent = nil
		return
	}

	// Verify all inputs are still accessible (might have been consumed during pause)
	for _, input := range recipe.Inputs {
		if !char.HasAccessibleItem(input.ItemType) {
			char.Intent = nil
			return
		}
	}

	char.ActionProgress += delta
	if char.ActionProgress >= recipe.Duration {
		char.ActionProgress = 0

		// Consume all recipe inputs
		consumed := make(map[string]*entity.Item)
		for _, input := range recipe.Inputs {
			item := char.ConsumeAccessibleItem(input.ItemType)
			if item == nil {
				char.Intent = nil
				return
			}
			consumed[input.ItemType] = item
		}

		// Dispatch to per-recipe creation function
		var crafted *entity.Item
		switch recipe.ID {
		case "hollow-gourd":
			crafted = system.CreateVessel(consumed["gourd"], recipe)
		case "shell-hoe":
			crafted = system.CreateHoe(consumed["shell"], recipe)
		case "clay-brick":
			crafted = system.CreateBrick(consumed["clay"], recipe)
		}

		if crafted != nil {
			crafted.X = char.X
			crafted.Y = char.Y
			m.gameMap.AddItem(crafted)
		}

		// Log the craft
		if m.actionLog != nil {
			m.actionLog.Add(char.ID, char.Name, "activity", "Crafted "+recipe.Name)
		}

		// Complete the order (skip for repeatable recipes — order loops until world-state condition is met)
		if char.AssignedOrderID != 0 && !recipe.Repeatable {
			if order := m.findOrderByID(char.AssignedOrderID); order != nil {
				system.CompleteOrder(char, order, m.actionLog)
			}
		}

		char.CurrentActivity = "Idle"
		char.Intent = nil
	}
}

// applyTillSoil handles ActionTillSoil: ordered action — walk to tile, till it, check order completion.
func (m *Model) applyTillSoil(char *entity.Character, delta float64) {
	cpos := char.Pos()
	dest := char.Intent.Dest

	if cpos.X != dest.X || cpos.Y != dest.Y {
		// Not at destination — move toward it
		tx, ty := char.Intent.Target.X, char.Intent.Target.Y
		speed := char.EffectiveSpeed()
		char.SpeedAccumulator += float64(speed) * delta
		const movementThreshold = 7.5
		if char.SpeedAccumulator < movementThreshold {
			return
		}
		char.SpeedAccumulator -= movementThreshold
		m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty})
		return
	}

	// At destination — accumulate tilling progress
	char.ActionProgress += delta
	if char.ActionProgress >= config.ActionDurationMedium {
		char.ActionProgress = 0

		// Till the soil
		m.gameMap.SetTilled(dest)
		m.gameMap.UnmarkForTilling(dest)

		// Handle items at the tilled position
		if item := m.gameMap.ItemAt(dest); item != nil {
			isGrowing := item.Plant != nil && item.Plant.IsGrowing
			if isGrowing {
				m.gameMap.RemoveItem(item)
			} else {
				adjX, adjY, found := system.FindEmptyAdjacent(dest.X, dest.Y, m.gameMap)
				if found {
					item.X = adjX
					item.Y = adjY
				}
			}
		}

		if m.actionLog != nil {
			m.actionLog.Add(char.ID, char.Name, "activity", "Tilled soil")
		}

		char.CurrentActivity = "Idle"
		char.Intent = nil

		// Check if till order is complete (pool exhausted)
		if char.AssignedOrderID != 0 {
			if order := m.findOrderByID(char.AssignedOrderID); order != nil && order.ActivityID == "tillSoil" {
				if !system.HasUnfilledTillingPositions(m.gameMap) {
					system.CompleteOrder(char, order, m.actionLog)
				}
			}
		}
	}
}

// applyPlant handles ActionPlant: ordered action — walk to tilled tile, plant item, check order completion.
func (m *Model) applyPlant(char *entity.Character, delta float64) {
	cpos := char.Pos()
	dest := char.Intent.Dest

	if cpos.X != dest.X || cpos.Y != dest.Y {
		// Not at destination — move toward it
		tx, ty := char.Intent.Target.X, char.Intent.Target.Y
		speed := char.EffectiveSpeed()
		char.SpeedAccumulator += float64(speed) * delta
		const movementThreshold = 7.5
		if char.SpeedAccumulator < movementThreshold {
			return
		}
		char.SpeedAccumulator -= movementThreshold
		m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty})
		return
	}

	// At destination — accumulate planting progress
	char.ActionProgress += delta
	if char.ActionProgress >= config.ActionDurationMedium {
		char.ActionProgress = 0

		// Find the order to get target type and locked variety
		var order *entity.Order
		if char.AssignedOrderID != 0 {
			order = m.findOrderByID(char.AssignedOrderID)
		}
		if order == nil {
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}

		// Consume a plantable item matching the order
		plantedItem := system.ConsumePlantable(char, order.TargetType, order.LockedVariety)
		if plantedItem == nil {
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}

		// Look up parent variety for sprout creation
		registry := m.gameMap.Varieties()
		var parentVariety *entity.ItemVariety
		if plantedItem.SourceVarietyID != "" {
			parentVariety = registry.Get(plantedItem.SourceVarietyID)
		}
		if parentVariety == nil {
			// Berries/mushrooms planted directly — look up their own variety
			parentVariety = registry.GetByAttributes(plantedItem.ItemType, plantedItem.Kind, plantedItem.Color, plantedItem.Pattern, plantedItem.Texture)
		}
		if parentVariety == nil {
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}

		// Push aside any loose items on the tile before planting
		pushLooseItemsAside(dest, cpos, m.gameMap)

		// Create sprout from the parent variety
		sprout := entity.CreateSprout(dest.X, dest.Y, parentVariety)
		m.gameMap.AddItem(sprout)

		// Lock the variety on the order (subsequent plants use same variety)
		if order.LockedVariety == "" {
			order.LockedVariety = entity.GenerateVarietyID(
				plantedItem.ItemType, plantedItem.Kind, plantedItem.Color, plantedItem.Pattern, plantedItem.Texture,
			)
		}

		if m.actionLog != nil {
			m.actionLog.Add(char.ID, char.Name, "activity", fmt.Sprintf("Planted %s", plantedItem.Description()))
		}

		// Check if plant order is complete (no more tiles or no more items)
		if !system.HasEmptyTilledTile(m.gameMap) ||
			!system.PlantableItemExists(m.gameMap.Items(), m.gameMap.Characters(), order.TargetType) {
			system.CompleteOrder(char, order, m.actionLog)
		}

		char.CurrentActivity = "Idle"
		char.Intent = nil
	}
}

// applyForage handles ActionForage: self-managing idle foraging — optional vessel procurement
// then food pickup.
func (m *Model) applyForage(char *entity.Character, delta float64) {
	// Self-managing foraging action — two phases:
	// Phase 1 (optional): If TargetItem is a vessel on the ground, pick it up via RunVesselProcurement
	// Phase 2: Move to food target, pick it up, go idle
	cpos := char.Pos()
	target := char.Intent.TargetItem

	// Phase 1: vessel procurement (if target is a vessel on the ground)
	if target != nil && target.Container != nil {
		status := system.RunVesselProcurement(char, target, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
		switch status {
		case system.ProcureApproaching:
			m.moveWithCollision(char, cpos, delta)
			return
		case system.ProcureInProgress:
			return
		case system.ProcureFailed:
			return
		case system.ProcureReady:
			// Vessel in hand — find food target and continue
			foodIntent := system.FindForageFoodIntent(char, cpos, m.gameMap.Items(), m.actionLog, m.gameMap)
			if foodIntent == nil {
				// No food available — go idle
				char.CurrentActivity = "Idle"
				char.Intent = nil
				char.IdleCooldown = config.IdleCooldown
				return
			}
			char.Intent = foodIntent
			return
		}
	}

	// Phase 2: food pickup
	if target == nil {
		char.CurrentActivity = "Idle"
		char.Intent = nil
		return
	}

	ipos := target.Pos()
	if cpos.X == ipos.X && cpos.Y == ipos.Y {
		// At food — accumulate pickup progress
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDurationShort {
			char.ActionProgress = 0
			if m.gameMap.HasItemOnMap(target) {
				system.Pickup(char, target, m.gameMap, m.actionLog, m.gameMap.Varieties())
			}
			// Foraging completes after one food item — go idle
			// (Pickup already clears intent and sets idle cooldown for PickupToInventory)
			// For PickupToVessel, Pickup does NOT clear intent, so we do it explicitly
			char.Intent = nil
			char.IdleCooldown = config.IdleCooldown
			char.CurrentActivity = "Idle"
		}
		return
	}

	// Not at food yet — move toward it
	m.moveWithCollision(char, cpos, delta)
}

// applyFillVessel handles ActionFillVessel: self-managing fetch-water — vessel procurement
// then fill at water source.
func (m *Model) applyFillVessel(char *entity.Character, delta float64) {
	// Fetch water action — two phases:
	// Phase 1 (via RunVesselProcurement): pick up ground vessel if needed
	// Phase 2 (via RunWaterFill): move to water and fill the vessel
	cpos := char.Pos()
	vessel := char.Intent.TargetItem

	// Phase 1: vessel procurement (shared helper)
	status := system.RunVesselProcurement(char, vessel, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
	switch status {
	case system.ProcureApproaching:
		m.moveWithCollision(char, cpos, delta)
		return
	case system.ProcureInProgress:
		return
	case system.ProcureFailed:
		return
	case system.ProcureReady:
		// Vessel in hand — check if already has water (ground water vessel pickup)
		if vessel != nil && vessel.Container != nil && len(vessel.Container.Contents) > 0 {
			// Already has water — mission accomplished, go idle
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}
		// Empty vessel — fall through to Phase 2
	}

	// Phase 2: fill vessel at water (shared helper)
	fillStatus := system.RunWaterFill(char, vessel, entity.ActionFillVessel, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
	switch fillStatus {
	case system.FillApproaching:
		m.moveWithCollision(char, cpos, delta)
		return
	case system.FillInProgress:
		return
	case system.FillFailed:
		return
	case system.FillReady:
		char.CurrentActivity = "Idle"
		char.Intent = nil
	}
}

// applyWaterGarden handles ActionWaterGarden: ordered action — vessel procurement, fill,
// then walk to dry tile and water it.
func (m *Model) applyWaterGarden(char *entity.Character, delta float64) {
	// Water Garden — ordered action pattern (like TillSoil, Plant).
	// Three phases, detected statelessly each tick:
	// Phase 1: vessel not in inventory → RunVesselProcurement (pick up ground vessel)
	// Phase 2: vessel in inventory, no water → RunWaterFill (fill at water source)
	// Phase 3: vessel has water → move to dry tile, water it, clear intent
	cpos := char.Pos()
	vessel := char.Intent.TargetItem

	// Phase 1: vessel procurement (if vessel is on the ground)
	vesselOnGround := vessel != nil && m.gameMap.HasItemOnMap(vessel)
	if vesselOnGround {
		status := system.RunVesselProcurement(char, vessel, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
		switch status {
		case system.ProcureApproaching:
			m.moveWithCollision(char, cpos, delta)
			return
		case system.ProcureInProgress:
			return
		case system.ProcureFailed:
			return
		case system.ProcureReady:
			// Vessel in hand — clear intent so findWaterGardenIntent
			// re-evaluates for Phase 2 (fill) or Phase 3 (water tiles)
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}
	}

	// Phase 2: fill vessel at water source (vessel in inventory, empty)
	vesselEmpty := vessel != nil && vessel.Container != nil && len(vessel.Container.Contents) == 0
	if vesselEmpty {
		fillStatus := system.RunWaterFill(char, vessel, entity.ActionWaterGarden, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
		switch fillStatus {
		case system.FillApproaching:
			m.moveWithCollision(char, cpos, delta)
			return
		case system.FillInProgress:
			return
		case system.FillFailed:
			return
		case system.FillReady:
			// Vessel filled — clear intent so findWaterGardenIntent
			// re-evaluates for Phase 3 (water tiles)
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}
	}

	// Phase 3: water tiles (vessel has water)
	dest := char.Intent.Dest
	if cpos.X != dest.X || cpos.Y != dest.Y {
		m.moveWithCollision(char, cpos, delta)
		return
	}

	// At destination — accumulate watering progress
	if char.CurrentActivity != "Watering garden" {
		char.CurrentActivity = "Watering garden"
	}
	char.ActionProgress += delta
	if char.ActionProgress >= config.ActionDurationShort {
		char.ActionProgress = 0

		// Water the tile
		m.gameMap.SetManuallyWatered(dest)

		// Consume 1 unit of water from vessel
		if vessel != nil {
			system.DrinkFromVessel(vessel)
		}

		if m.actionLog != nil {
			m.actionLog.Add(char.ID, char.Name, "activity", "Watered the garden")
		}

		char.CurrentActivity = "Idle"
		char.Intent = nil

		// Check order completion — no dry tilled planted tiles remain
		if char.AssignedOrderID != 0 {
			if order := m.findOrderByID(char.AssignedOrderID); order != nil && order.ActivityID == "waterGarden" {
				if !system.DryTilledPlantedTileExists(m.gameMap.Items(), m.gameMap) {
					system.CompleteOrder(char, order, m.actionLog)
				}
			}
		}
	}
}

// applyHelpFeed handles ActionHelpFeed: self-managing — procure food, walk to needy
// character, drop food cardinal-adjacent.
func (m *Model) applyHelpFeed(char *entity.Character, delta float64) {
	// Help Feed — self-managing action: procure food, deliver to needy character
	// Phase 1: food on ground → walk to it, pick up
	// Phase 2: food in inventory → walk to needer, drop cardinal-adjacent
	cpos := char.Pos()
	target := char.Intent.TargetItem
	needer := char.Intent.TargetCharacter

	if needer == nil || needer.IsDead {
		// Needer gone — drop food at current position if carrying, go idle
		if target != nil {
			for _, item := range char.Inventory {
				if item == target {
					system.DropItem(char, target, m.gameMap, m.actionLog)
					break
				}
			}
		}
		char.CurrentActivity = "Idle"
		char.Intent = nil
		return
	}

	// Phase 1: procurement — TargetItem is on the ground
	if target != nil {
		ipos := target.Pos()
		if m.gameMap.HasItemOnMap(target) {
			if cpos.X == ipos.X && cpos.Y == ipos.Y {
				// At food/vessel — pick up
				char.ActionProgress += delta
				if char.ActionProgress >= config.ActionDurationShort {
					char.ActionProgress = 0
					result := system.Pickup(char, target, m.gameMap, m.actionLog, m.gameMap.Varieties())

					// Determine what to deliver
					deliveryItem := target
					if result == system.PickupToVessel {
						// Food absorbed into carried vessel — deliver the vessel instead
						deliveryItem = char.GetCarriedVessel()
					} else if result == system.PickupFailed {
						char.CurrentActivity = "Idle"
						char.Intent = nil
						return
					}

					// Rebuild intent for delivery phase
					npos := needer.Pos()
					nx, ny := system.NextStepBFS(cpos.X, cpos.Y, npos.X, npos.Y, m.gameMap)
					char.Intent = &entity.Intent{
						Target:          types.Position{X: nx, Y: ny},
						Dest:            npos,
						Action:          entity.ActionHelpFeed,
						TargetItem:      deliveryItem,
						TargetCharacter: needer,
					}
					char.CurrentActivity = "Bringing food to " + needer.Name
					if m.actionLog != nil {
						m.actionLog.Add(char.ID, char.Name, "activity", "Bringing food to "+needer.Name)
					}
				}
				return
			}
			// Not at food yet — move toward it
			m.moveWithCollision(char, cpos, delta)
			return
		}
		// Check if target is in inventory (carried food — skip to delivery)
		inInventory := false
		for _, item := range char.Inventory {
			if item == target {
				inInventory = true
				break
			}
		}
		if !inInventory {
			// Food gone, not in inventory — give up
			char.CurrentActivity = "Idle"
			char.Intent = nil
			return
		}
	}

	// Phase 2: delivery — food in inventory, walk toward needer
	npos := needer.Pos()
	if cpos.IsCardinallyAdjacentTo(npos) {
		// Adjacent to needer — drop food at an empty cardinal tile next to needer
		// (not where the helper is standing, so the needer can reach it)
		if target != nil {
			for _, item := range char.Inventory {
				if item == target {
					// Find empty cardinal tile adjacent to needer for the drop
					dropPos := findEmptyCardinalTile(npos, cpos, m.gameMap)
					char.RemoveFromInventory(target)
					target.X = dropPos.X
					target.Y = dropPos.Y
					m.gameMap.AddItem(target)
					if m.actionLog != nil {
						m.actionLog.Add(char.ID, char.Name, "activity",
							"Brought "+target.Description()+" to "+needer.Name)
					}
					// Signal the needer to re-evaluate — clear their current intent
					// so they notice the closer food on their next tick
					needer.Intent = nil
					if m.actionLog != nil {
						m.actionLog.Add(char.ID, char.Name, "social",
							char.Name+" called out to "+needer.Name)
					}
					break
				}
			}
		}
		char.CurrentActivity = "Idle"
		char.Intent = nil
		char.IdleCooldown = config.IdleCooldown
		return
	}

	// Not adjacent yet — move toward needer
	char.CurrentActivity = "Bringing food to " + needer.Name
	m.moveWithCollision(char, cpos, delta)
}

// applyHelpWater handles ActionHelpWater: self-managing — procure vessel, fill at water,
// walk to needy character, drop vessel cardinal-adjacent.
func (m *Model) applyHelpWater(char *entity.Character, delta float64) {
	// Help Water — self-managing action: procure vessel, fill at water, deliver to needy character
	// Phase 1: vessel on ground → RunVesselProcurement
	// Phase 2: vessel in inventory, empty → RunWaterFill
	// Phase 3: vessel has water → walk to needer, drop cardinal-adjacent
	cpos := char.Pos()
	vessel := char.Intent.TargetItem
	needer := char.Intent.TargetCharacter // Capture before procurement may nil intent

	if needer == nil || needer.IsDead {
		// Needer gone — drop vessel at current position if carrying, go idle
		if vessel != nil {
			for _, item := range char.Inventory {
				if item == vessel {
					system.DropItem(char, vessel, m.gameMap, m.actionLog)
					break
				}
			}
		}
		char.CurrentActivity = "Idle"
		char.Intent = nil
		return
	}

	// Phase 1: vessel procurement (if vessel is on the ground)
	vesselOnGround := vessel != nil && m.gameMap.HasItemOnMap(vessel)
	if vesselOnGround {
		status := system.RunVesselProcurement(char, vessel, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
		switch status {
		case system.ProcureApproaching:
			m.moveWithCollision(char, cpos, delta)
			return
		case system.ProcureInProgress:
			return
		case system.ProcureFailed:
			return
		case system.ProcureReady:
			// Vessel in hand — check if already has water (ground water vessel pickup)
			if vessel != nil && vessel.Container != nil && len(vessel.Container.Contents) > 0 {
				// Already has water — transition to delivery (skip fill)
				npos := needer.Pos()
				nx, ny := system.NextStepBFS(cpos.X, cpos.Y, npos.X, npos.Y, m.gameMap)
				char.Intent = &entity.Intent{
					Target:          types.Position{X: nx, Y: ny},
					Dest:            npos,
					Action:          entity.ActionHelpWater,
					TargetItem:      vessel,
					TargetCharacter: needer,
				}
				char.CurrentActivity = "Bringing water to " + needer.Name
				if m.actionLog != nil {
					m.actionLog.Add(char.ID, char.Name, "activity", "Bringing water to "+needer.Name)
				}
				return
			}
			// Empty vessel — fall through to fill phase (same tick)
		}
	}

	// Phase 2: fill vessel at water source (vessel in inventory, empty)
	vesselEmpty := vessel != nil && vessel.Container != nil && len(vessel.Container.Contents) == 0
	if vesselEmpty {
		fillStatus := system.RunWaterFill(char, vessel, entity.ActionHelpWater, m.gameMap, m.actionLog, m.gameMap.Varieties(), delta)
		// Restore TargetCharacter after RunWaterFill (may have rebuilt intent without it)
		if char.Intent != nil && char.Intent.TargetCharacter == nil && needer != nil {
			char.Intent.TargetCharacter = needer
		}
		switch fillStatus {
		case system.FillApproaching:
			m.moveWithCollision(char, cpos, delta)
			return
		case system.FillInProgress:
			return
		case system.FillFailed:
			return
		case system.FillReady:
			// Vessel filled — transition to delivery
			npos := needer.Pos()
			nx, ny := system.NextStepBFS(cpos.X, cpos.Y, npos.X, npos.Y, m.gameMap)
			char.Intent = &entity.Intent{
				Target:          types.Position{X: nx, Y: ny},
				Dest:            npos,
				Action:          entity.ActionHelpWater,
				TargetItem:      vessel,
				TargetCharacter: needer,
			}
			char.CurrentActivity = "Bringing water to " + needer.Name
			if m.actionLog != nil {
				m.actionLog.Add(char.ID, char.Name, "activity", "Bringing water to "+needer.Name)
			}
			return
		}
	}

	// Phase 3: delivery — vessel has water, walk toward needer
	npos := needer.Pos()
	if cpos.IsCardinallyAdjacentTo(npos) {
		// Adjacent to needer — drop water vessel
		if vessel != nil {
			for _, item := range char.Inventory {
				if item == vessel {
					dropPos := findEmptyCardinalTile(npos, cpos, m.gameMap)
					char.RemoveFromInventory(vessel)
					vessel.X = dropPos.X
					vessel.Y = dropPos.Y
					m.gameMap.AddItem(vessel)
					if m.actionLog != nil {
						m.actionLog.Add(char.ID, char.Name, "activity",
							"Brought water to "+needer.Name)
					}
					// Signal the needer to re-evaluate
					needer.Intent = nil
					if m.actionLog != nil {
						m.actionLog.Add(char.ID, char.Name, "social",
							char.Name+" called out to "+needer.Name)
					}
					break
				}
			}
		}
		char.CurrentActivity = "Idle"
		char.Intent = nil
		char.IdleCooldown = config.IdleCooldown
		return
	}

	// Not adjacent yet — move toward needer
	char.CurrentActivity = "Bringing water to " + needer.Name
	m.moveWithCollision(char, cpos, delta)
}

// findEmptyCardinalTile finds an unoccupied cardinal-adjacent tile next to center.
// Prefers tiles not occupied by the helper (avoidPos). Falls back to avoidPos if all others blocked.
func findEmptyCardinalTile(center types.Position, avoidPos types.Position, gameMap *game.Map) types.Position {
	cardinalDirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for _, dir := range cardinalDirs {
		pos := types.Position{X: center.X + dir[0], Y: center.Y + dir[1]}
		if pos == avoidPos {
			continue // Skip helper's position
		}
		if !gameMap.IsValid(pos) || gameMap.IsWater(pos) {
			continue
		}
		if occupant := gameMap.CharacterAt(pos); occupant != nil {
			continue
		}
		if f := gameMap.FeatureAt(pos); f != nil && !f.IsPassable() {
			continue
		}
		return pos
	}
	// All cardinal tiles blocked — fall back to helper's position (best effort)
	return avoidPos
}

// pushLooseItemsAside moves any non-growing items at the given position to an adjacent empty tile.
// Used during planting to clear loose items (seeds, vessels) from a tilled tile before placing a sprout.
func pushLooseItemsAside(pos types.Position, charPos types.Position, gameMap *game.Map) {
	items := gameMap.ItemsAt(pos)
	for _, item := range items {
		if item.Plant != nil && item.Plant.IsGrowing {
			continue // Don't push growing plants
		}
		adjPos := findEmptyAdjacentTile(pos, charPos, gameMap)
		if adjPos != pos {
			item.X = adjPos.X
			item.Y = adjPos.Y
		}
	}
}

// findEmptyAdjacentTile finds a cardinal-adjacent tile with no items, characters, or blocking features.
// Used for pushing loose items aside during planting. Falls back to the original position if no empty tile found.
func findEmptyAdjacentTile(center types.Position, avoidPos types.Position, gameMap *game.Map) types.Position {
	cardinalDirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for _, dir := range cardinalDirs {
		pos := types.Position{X: center.X + dir[0], Y: center.Y + dir[1]}
		if pos == avoidPos {
			continue
		}
		if !gameMap.IsValid(pos) || gameMap.IsWater(pos) {
			continue
		}
		if f := gameMap.FeatureAt(pos); f != nil && !f.IsPassable() {
			continue
		}
		if gameMap.ItemAt(pos) != nil {
			continue
		}
		return pos
	}
	// No empty tile found — leave item in place (best effort)
	return center
}

// moveWithCollision handles speed accumulation and collision-aware movement for self-managing actions.
// Used by ActionFillVessel and ActionTillSoil-style actions that handle their own movement.
func (m *Model) moveWithCollision(char *entity.Character, cpos types.Position, delta float64) {
	speed := char.EffectiveSpeed()
	char.SpeedAccumulator += float64(speed) * delta
	const movementThreshold = 7.5
	if char.SpeedAccumulator < movementThreshold {
		return
	}
	char.SpeedAccumulator -= movementThreshold

	cx, cy := cpos.X, cpos.Y
	moved := false

	if char.DisplacementStepsLeft > 0 {
		moved = m.takeDisplacementStep(char, cx, cy)
	} else {
		tx, ty := char.Intent.Target.X, char.Intent.Target.Y
		triedPositions := map[[2]int]bool{{tx, ty}: true}
		if m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
			moved = true
		} else {
			if m.gameMap.CharacterAt(types.Position{X: tx, Y: ty}) != nil {
				moved = m.initiateDisplacement(char, cx, cy, tx, ty)
			}
			if !moved {
				for attempts := 0; attempts < 5 && !moved; attempts++ {
					altStep := m.findAlternateStep(char, cx, cy, triedPositions)
					if altStep == nil {
						break
					}
					tx, ty = altStep[0], altStep[1]
					triedPositions[[2]int{tx, ty}] = true
					if m.gameMap.MoveCharacter(char, types.Position{X: tx, Y: ty}) {
						moved = true
					}
				}
			}
		}
	}

	if !moved {
		char.SpeedAccumulator += movementThreshold * 0.5
	}
}

// takeDisplacementStep moves the character one step in the current displacement direction.
// If the primary direction is blocked, tries the opposite perpendicular.
// If both are blocked, clears displacement state and returns false.
func (m *Model) takeDisplacementStep(char *entity.Character, cx, cy int) bool {
	ddx, ddy := char.DisplacementDX, char.DisplacementDY

	dispPos := types.Position{X: cx + ddx, Y: cy + ddy}
	if m.gameMap.IsValid(dispPos) && m.gameMap.MoveCharacter(char, dispPos) {
		char.DisplacementStepsLeft--
		if char.DisplacementStepsLeft == 0 {
			char.DisplacementDX, char.DisplacementDY = 0, 0
		}
		return true
	}

	// Primary direction blocked — try opposite perpendicular
	oddx, oddy := -ddx, -ddy
	otherPos := types.Position{X: cx + oddx, Y: cy + oddy}
	if m.gameMap.IsValid(otherPos) && m.gameMap.MoveCharacter(char, otherPos) {
		char.DisplacementDX, char.DisplacementDY = oddx, oddy
		char.DisplacementStepsLeft--
		if char.DisplacementStepsLeft == 0 {
			char.DisplacementDX, char.DisplacementDY = 0, 0
		}
		return true
	}

	// Both directions blocked — clear displacement
	char.DisplacementStepsLeft = 0
	char.DisplacementDX, char.DisplacementDY = 0, 0
	return false
}

// initiateDisplacement sets displacement state after a character-character collision.
// Randomly selects one of the two perpendicular directions. If both are blocked, returns false.
// On success, takes the first displacement step this tick and returns true.
func (m *Model) initiateDisplacement(char *entity.Character, cx, cy, tx, ty int) bool {
	moveDX := sign(tx - cx)
	moveDY := sign(ty - cy)
	if moveDX == 0 && moveDY == 0 {
		return false
	}

	// Perpendicular directions to the movement direction
	perp1DX, perp1DY := -moveDY, moveDX
	perp2DX, perp2DY := moveDY, -moveDX

	// Randomly select primary and secondary perpendicular directions
	var pDX, pDY, sDX, sDY int
	if rand.Intn(2) == 0 {
		pDX, pDY, sDX, sDY = perp1DX, perp1DY, perp2DX, perp2DY
	} else {
		pDX, pDY, sDX, sDY = perp2DX, perp2DY, perp1DX, perp1DY
	}

	// Find an available perpendicular direction
	var chDX, chDY int
	var found bool
	p1Pos := types.Position{X: cx + pDX, Y: cy + pDY}
	if m.gameMap.IsValid(p1Pos) && !m.gameMap.IsBlocked(p1Pos) {
		chDX, chDY, found = pDX, pDY, true
	} else {
		p2Pos := types.Position{X: cx + sDX, Y: cy + sDY}
		if m.gameMap.IsValid(p2Pos) && !m.gameMap.IsBlocked(p2Pos) {
			chDX, chDY, found = sDX, sDY, true
		}
	}

	if !found {
		return false
	}

	char.DisplacementStepsLeft = 3
	char.DisplacementDX = chDX
	char.DisplacementDY = chDY
	char.UsingBFS = false // Clear sticky BFS — new direction after sidestepping

	// Take the first displacement step this tick
	dispPos := types.Position{X: cx + chDX, Y: cy + chDY}
	if m.gameMap.MoveCharacter(char, dispPos) {
		char.DisplacementStepsLeft--
		if char.DisplacementStepsLeft == 0 {
			char.DisplacementDX, char.DisplacementDY = 0, 0
		}
		return true
	}

	// IsBlocked said clear but MoveCharacter still failed — state set for next tick
	return false
}

// findAlternateStep finds an alternate step when the preferred step is blocked.
// Returns [x, y] of alternate position, or nil if no valid alternative.
// triedPositions contains positions already attempted this tick.
func (m *Model) findAlternateStep(char *entity.Character, cx, cy int, triedPositions map[[2]int]bool) []int {
	// Use destination position (where we need to stand to interact)
	// This is set correctly for adjacency-based interactions (springs, talking, looking)
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

	// Try adjacent positions that still move toward goal
	// Priority: diagonal toward goal > orthogonal toward goal > wait
	dx := sign(goalX - cx)
	dy := sign(goalY - cy)

	// Build list of candidate positions
	candidates := [][]int{}

	// Primary alternates: move in one axis toward goal
	if dx != 0 {
		candidates = append(candidates, []int{cx + dx, cy})
	}
	if dy != 0 {
		candidates = append(candidates, []int{cx, cy + dy})
	}

	// Secondary: orthogonal moves that don't move away from goal
	if dx == 0 {
		candidates = append(candidates, []int{cx + 1, cy}, []int{cx - 1, cy})
	}
	if dy == 0 {
		candidates = append(candidates, []int{cx, cy + 1}, []int{cx, cy - 1})
	}

	// Tertiary: all other adjacent positions as last resort
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

	// Find first valid candidate not already tried
	for _, pos := range candidates {
		x, y := pos[0], pos[1]
		key := [2]int{x, y}
		if triedPositions[key] {
			continue
		}
		candidatePos := types.Position{X: x, Y: y}
		if !m.gameMap.IsValid(candidatePos) {
			continue
		}
		if m.gameMap.IsBlocked(candidatePos) {
			continue
		}
		return pos
	}

	return nil
}

// sign returns -1, 0, or 1 based on the value
func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// getEatenItemType returns the item type being eaten, resolving vessel contents.
// For vessels, the food type comes from the first stack's variety, not the vessel itself.
func getEatenItemType(item *entity.Item) string {
	if item.Container != nil && len(item.Container.Contents) > 0 {
		return item.Container.Contents[0].Variety.ItemType
	}
	return item.ItemType
}

// applyExtract handles ActionExtract: walk to a living plant, then extract seeds.
// Walk-then-act pattern with ActionDurationShort. Seeds are routed directly to vessel
// or inventory — never placed on the ground.
func (m *Model) applyExtract(char *entity.Character, delta float64) {
	plant := char.Intent.TargetItem
	if plant == nil || plant.Plant == nil {
		char.Intent = nil
		return
	}

	cpos := char.Pos()
	ppos := plant.Pos()

	// Walking phase: not at target plant
	if cpos != ppos {
		m.moveWithCollision(char, cpos, delta)
		return
	}

	// Working phase: at plant, accumulate progress
	char.ActionProgress += delta
	if char.ActionProgress < config.ActionDurationShort {
		return
	}

	// Extraction complete
	char.ActionProgress = 0

	// Create seed from the plant's attributes
	sourceVarietyID := entity.GenerateVarietyID(plant.ItemType, plant.Kind, plant.Color, plant.Pattern, plant.Texture)
	seed := entity.NewSeed(char.X, char.Y, plant.ItemType, sourceVarietyID, plant.Kind, plant.Color, plant.Pattern, plant.Texture)

	// Route seed: vessel first, then inventory
	registry := m.gameMap.Varieties()
	routed := false

	// Try ALL carried vessels for seed routing — not just the first
	if vessel := system.FindCarriedVesselFor(char, seed, registry); vessel != nil {
		if system.AddToVessel(vessel, seed, registry) {
			routed = true
		}
	}

	if !routed {
		if char.HasInventorySpace() {
			char.AddToInventory(seed)
			routed = true
		}
	}

	if !routed {
		// No room for seeds — log and pause
		if m.actionLog != nil {
			m.actionLog.Add(char.ID, char.Name, "extract", "No room for seeds")
		}
		char.Intent = nil
		return
	}

	// Set seed timer on the plant
	if cfg, ok := config.ItemLifecycle[plant.ItemType]; ok {
		plant.Plant.SeedTimer = cfg.SpawnInterval * float64(len(m.gameMap.Items()))
	}

	// Lock the variety on the order (subsequent extractions target same variety)
	if char.AssignedOrderID != 0 {
		if order := m.findOrderByID(char.AssignedOrderID); order != nil && order.LockedVariety == "" {
			order.LockedVariety = entity.GenerateVarietyID(
				plant.ItemType, plant.Kind, plant.Color, plant.Pattern, plant.Texture,
			)
		}
	}

	if m.actionLog != nil {
		m.actionLog.Add(char.ID, char.Name, "activity",
			fmt.Sprintf("Extracted %s from %s", seed.Kind, plant.Description()))
	}

	// Check if extract order is complete: inventory full and no vessel can accept more seeds
	if char.AssignedOrderID != 0 {
		if order := m.findOrderByID(char.AssignedOrderID); order != nil && order.ActivityID == "extract" {
			if !char.HasInventorySpace() && system.FindCarriedVesselFor(char, seed, m.gameMap.Varieties()) == nil {
				system.CompleteOrder(char, order, m.actionLog)
			}
		}
	}

	// Clear intent — ordered action pattern: next tick re-evaluates via findExtractIntent
	char.Intent = nil
}

// applyDig handles ActionDig: walk to a clay terrain tile, then dig up a lump of clay.
// Walk-then-act pattern with ActionDurationShort. Clay added directly to inventory.
// Ordered action pattern: clear intent after digging so next tick re-evaluates.
func (m *Model) applyDig(char *entity.Character, delta float64) {
	cpos := char.Pos()

	// Guard: no inventory space (safety net — findDigIntent should prevent this)
	if !char.HasInventorySpace() {
		char.Intent = nil
		return
	}

	// Walking phase: not yet at clay tile
	if cpos != char.Intent.Dest {
		m.moveWithCollision(char, cpos, delta)
		return
	}

	// Working phase: at clay tile, accumulate progress
	char.ActionProgress += delta
	if char.ActionProgress < config.ActionDurationMedium {
		return
	}

	// Dig complete
	char.ActionProgress = 0
	clay := entity.NewClay(cpos.X, cpos.Y)
	char.AddToInventory(clay)

	if m.actionLog != nil {
		m.actionLog.Add(char.ID, char.Name, "activity", "Dug clay")
	}

	// Clear intent — ordered action pattern: next tick re-evaluates via findDigIntent
	char.Intent = nil
}

// applyBuildFence handles ActionBuildFence: walk to adjacent tile, build fence on marked tile.
// Walk-then-act pattern with ActionDurationMedium. Ordered action pattern: clear intent after
// building so next tick re-evaluates via findBuildFenceIntent.
func (m *Model) applyBuildFence(char *entity.Character, delta float64) {
	if char.Intent.TargetBuildPos == nil {
		char.Intent = nil
		return
	}
	buildPos := *char.Intent.TargetBuildPos
	dest := char.Intent.Dest
	cpos := char.Pos()

	// Delivery mode: character is at the build tile with bricks in inventory (supply-drop)
	if cpos == buildPos && m.hasMaterialInInventory(char, "brick") {
		m.deliverBricks(char, buildPos)
		return
	}

	// Walking phase: not yet at destination
	if cpos != dest {
		m.moveWithCollision(char, cpos, delta)
		return
	}

	// Arrival check (DD-28 layer 2): if a character now occupies the build tile, re-evaluate
	if m.gameMap.CharacterAt(buildPos) != nil {
		char.Intent = nil
		return
	}

	// Working phase: accumulate progress
	if char.CurrentActivity != "Building fence" {
		char.CurrentActivity = "Building fence"
	}
	char.ActionProgress += delta
	if char.ActionProgress < config.ActionDurationMedium {
		return
	}
	char.ActionProgress = 0

	// Determine material source: bundle in inventory or bricks on ground
	var material string
	var bundle *entity.Item
	for _, inv := range char.Inventory {
		if inv != nil && inv.BundleCount > 0 {
			bundle = inv
			break
		}
	}

	if bundle != nil {
		// Bundle build: consume from inventory
		material = bundle.ItemType
		char.RemoveFromInventory(bundle)
	} else {
		// Brick build: consume 6 bricks from ground at build position
		mark, ok := m.gameMap.GetConstructionMark(buildPos)
		if !ok {
			char.Intent = nil
			return
		}
		material = mark.Material
		consumed := 0
		for _, item := range m.gameMap.ItemsAt(buildPos) {
			if item.ItemType == material && consumed < 6 {
				m.gameMap.RemoveItem(item)
				consumed++
			}
		}
		if consumed < 6 {
			char.Intent = nil
			return
		}
	}

	var materialColor types.Color
	switch material {
	case "grass":
		materialColor = types.ColorPaleYellow
	case "stick":
		materialColor = types.ColorBrown
	case "brick":
		materialColor = types.ColorTerracotta
	default:
		materialColor = types.ColorBrown
	}

	// Place fence
	fence := entity.NewFence(buildPos.X, buildPos.Y, material, materialColor)
	m.gameMap.AddConstruct(fence)
	m.gameMap.UnmarkForConstruction(buildPos)

	// DD-28 layer 3: displace any character standing on the build tile (safety net)
	cardinalDirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	if occupant := m.gameMap.CharacterAt(buildPos); occupant != nil {
		for _, dir := range cardinalDirs {
			adj := types.Position{X: buildPos.X + dir[0], Y: buildPos.Y + dir[1]}
			if m.gameMap.MoveCharacter(occupant, adj) {
				break
			}
		}
	}

	// DD-33: displace items at build tile to adjacent tiles
	for _, item := range m.gameMap.ItemsAt(buildPos) {
		for _, dir := range cardinalDirs {
			adj := types.Position{X: buildPos.X + dir[0], Y: buildPos.Y + dir[1]}
			if !m.gameMap.IsBlocked(adj) {
				item.X = adj.X
				item.Y = adj.Y
				break
			}
		}
	}

	if m.actionLog != nil {
		m.actionLog.Add(char.ID, char.Name, "activity", fmt.Sprintf("Built %s fence", material))
	}

	// Clear intent — ordered action pattern: next tick re-evaluates via findBuildFenceIntent
	char.Intent = nil
}

// hasMaterialInInventory checks if a character has any items of the given type in inventory.
func (m *Model) hasMaterialInInventory(char *entity.Character, material string) bool {
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material {
			return true
		}
	}
	return false
}

// deliverBricks drops all bricks from inventory at the current position and clears intent.
func (m *Model) deliverBricks(char *entity.Character, buildPos types.Position) {
	var toDrop []*entity.Item
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == "brick" {
			toDrop = append(toDrop, inv)
		}
	}
	for _, item := range toDrop {
		char.RemoveFromInventory(item)
		item.X = buildPos.X
		item.Y = buildPos.Y
		m.gameMap.AddItem(item)
	}
	char.Intent = nil
}

// deliverMaterial drops all matching material items from inventory at the build position and clears intent.
func (m *Model) deliverMaterial(char *entity.Character, material string, buildPos types.Position) {
	var toDrop []*entity.Item
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material {
			toDrop = append(toDrop, inv)
		}
	}
	for _, item := range toDrop {
		char.RemoveFromInventory(item)
		item.X = buildPos.X
		item.Y = buildPos.Y
		m.gameMap.AddItem(item)
	}
	char.Intent = nil
}

// applyBuildHut handles ActionBuildHut: delivery of supplies and construction of hut wall/door segments.
func (m *Model) applyBuildHut(char *entity.Character, delta float64) {
	if char.Intent.TargetBuildPos == nil {
		char.Intent = nil
		return
	}
	buildPos := *char.Intent.TargetBuildPos
	dest := char.Intent.Dest
	cpos := char.Pos()

	// Delivery mode: character is at the build tile with material in inventory → drop supplies
	mark, markOk := m.gameMap.GetConstructionMark(buildPos)
	if markOk && cpos == buildPos && m.hasMaterialInInventory(char, mark.Material) {
		m.deliverMaterial(char, mark.Material, buildPos)
		return
	}

	// Walking phase: not yet at destination
	if cpos != dest {
		m.moveWithCollision(char, cpos, delta)
		return
	}

	// Arrival check (DD-28 layer 2): if a character now occupies the build tile, re-evaluate
	if m.gameMap.CharacterAt(buildPos) != nil {
		char.Intent = nil
		return
	}

	// Working phase: accumulate progress
	if char.CurrentActivity != "Building hut" {
		char.CurrentActivity = "Building hut"
	}
	char.ActionProgress += delta
	if char.ActionProgress < config.ActionDurationMedium {
		return
	}
	char.ActionProgress = 0

	// Read material and wallRole from mark
	if !markOk {
		char.Intent = nil
		return
	}
	material := mark.Material
	wallRole := mark.WallRole

	// Consume materials from ground at build position
	_, isBundleMaterial := config.MaxBundleSize[material]
	if isBundleMaterial {
		// Bundle consumption: find and remove 2 full bundles at buildPos
		consumed := 0
		for _, item := range m.gameMap.ItemsAt(buildPos) {
			if item.ItemType == material && item.BundleCount >= config.MaxBundleSize[material] && consumed < 2 {
				m.gameMap.RemoveItem(item)
				consumed++
			}
		}
		if consumed < 2 {
			char.Intent = nil
			return
		}
	} else {
		// Brick consumption: find and remove 12 items at buildPos
		consumed := 0
		for _, item := range m.gameMap.ItemsAt(buildPos) {
			if item.ItemType == material && consumed < 12 {
				m.gameMap.RemoveItem(item)
				consumed++
			}
		}
		if consumed < 12 {
			char.Intent = nil
			return
		}
	}

	// Material color mapping
	var materialColor types.Color
	switch material {
	case "grass":
		materialColor = types.ColorPaleYellow
	case "stick":
		materialColor = types.ColorBrown
	case "brick":
		materialColor = types.ColorTerracotta
	default:
		materialColor = types.ColorBrown
	}

	// Place hut construct
	construct := entity.NewHutConstruct(buildPos.X, buildPos.Y, material, materialColor, wallRole)
	m.gameMap.AddConstruct(construct)
	m.gameMap.UnmarkForConstruction(buildPos)

	// DD-28 layer 3: displace any character standing on the build tile (safety net)
	cardinalDirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	if occupant := m.gameMap.CharacterAt(buildPos); occupant != nil {
		for _, dir := range cardinalDirs {
			adj := types.Position{X: buildPos.X + dir[0], Y: buildPos.Y + dir[1]}
			if m.gameMap.MoveCharacter(occupant, adj) {
				break
			}
		}
	}

	// DD-33: displace items at build tile to adjacent tiles
	for _, item := range m.gameMap.ItemsAt(buildPos) {
		for _, dir := range cardinalDirs {
			adj := types.Position{X: buildPos.X + dir[0], Y: buildPos.Y + dir[1]}
			if !m.gameMap.IsBlocked(adj) {
				item.X = adj.X
				item.Y = adj.Y
				break
			}
		}
	}

	roleLabel := "wall"
	if wallRole == "door" {
		roleLabel = "door"
	}
	if m.actionLog != nil {
		m.actionLog.Add(char.ID, char.Name, "activity", fmt.Sprintf("Built %s hut %s", material, roleLabel))
	}

	// Clear intent — ordered action pattern: next tick re-evaluates via findBuildHutIntent
	char.Intent = nil
}
