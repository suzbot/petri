package system

import (
	"fmt"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// selectOrderActivity checks if a character should work on an order instead of a random idle activity.
// Returns an intent if the character takes or resumes an order, nil otherwise.
// Handles order assignment, resumption, and abandonment.
func selectOrderActivity(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, orders []*entity.Order, log *ActionLog) *entity.Intent {
	if orders == nil {
		return nil
	}

	// First priority: resume assigned order if character has one
	if char.AssignedOrderID != 0 {
		order := findOrderByID(orders, char.AssignedOrderID)
		if order != nil && (order.Status == entity.OrderAssigned || order.Status == entity.OrderPaused) {
			// Resume the order
			if order.Status == entity.OrderPaused {
				order.Status = entity.OrderAssigned
				if log != nil {
					log.Add(char.ID, char.Name, "order", fmt.Sprintf("Resuming order: %s", order.DisplayName()))
				}
			}
			if intent := findOrderIntent(char, pos, items, order, log, gameMap); intent != nil {
				return intent
			}
			// Order cannot be fulfilled - abandon it
			abandonOrder(char, order, orders, log)
			return nil
		}
		// Order was cancelled or doesn't exist - clear assignment
		char.AssignedOrderID = 0
	}

	// Second priority: take a new order if one is available and character can execute it
	// Orders can be taken with full inventory - will drop current item if needed during execution
	order := findAvailableOrder(char, orders)
	if order == nil {
		return nil
	}

	// Assign the order to this character
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	if log != nil {
		log.Add(char.ID, char.Name, "order", fmt.Sprintf("Taking order: %s", order.DisplayName()))
	}

	if intent := findOrderIntent(char, pos, items, order, log, gameMap); intent != nil {
		return intent
	}

	// Order cannot be fulfilled immediately - abandon it
	abandonOrder(char, order, orders, log)
	return nil
}

// findAvailableOrder finds the first open order that the character can execute.
// Checks activity requirements (know-how) against character's known activities.
// Future: can be extended to consider character preference, order urgency, etc.
func findAvailableOrder(char *entity.Character, orders []*entity.Order) *entity.Order {
	for _, order := range orders {
		if order.Status != entity.OrderOpen {
			continue
		}
		if canExecuteOrder(char, order) {
			return order
		}
	}
	return nil
}

// canExecuteOrder checks if a character meets the requirements to execute an order.
// Currently checks know-how requirements based on the activity definition.
func canExecuteOrder(char *entity.Character, order *entity.Order) bool {
	activity, ok := entity.ActivityRegistry[order.ActivityID]
	if !ok {
		return false // Unknown activity
	}

	// Check know-how requirement
	if activity.Availability == entity.AvailabilityKnowHow {
		if !char.KnowsActivity(order.ActivityID) {
			return false
		}
	}

	return true
}

// findOrderIntent creates an intent for executing an order based on its activity type.
// Dispatches to activity-specific intent finding logic.
func findOrderIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	switch order.ActivityID {
	case "harvest":
		return findHarvestIntent(char, pos, items, order, log, gameMap)
	case "craftVessel":
		return findCraftVesselIntent(char, pos, items, order, log)
	default:
		// Unknown activity type - cannot create intent
		return nil
	}
}

// findHarvestIntent creates an intent to harvest (pick up) a specific item type per order.
// Returns nil if no matching items exist on the map.
// Handles look-for-container (4d) and drop-when-blocked (4e) logic.
func findHarvestIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Find nearest item matching the order's target type
	target := findNearestItemByType(pos.X, pos.Y, items, order.TargetType)
	if target == nil {
		return nil // No matching items - will trigger abandonment
	}

	registry := gameMap.Varieties()

	// Check inventory state and handle vessel logic
	if char.Carrying == nil {
		// Not carrying anything - look for available vessel first (4d)
		availableVessel := FindAvailableVessel(pos.X, pos.Y, items, target, registry)
		if availableVessel != nil {
			// Go pick up vessel first
			vpos := availableVessel.Pos()
			vx, vy := vpos.X, vpos.Y
			if pos.X == vx && pos.Y == vy {
				newActivity := "Picking up vessel"
				if char.CurrentActivity != newActivity {
					char.CurrentActivity = newActivity
					if log != nil {
						log.Add(char.ID, char.Name, "order", "Picking up vessel for harvesting")
					}
				}
				return &entity.Intent{
					Target:     pos,
					Dest:       pos, // Already at destination
					Action:     entity.ActionPickup,
					TargetItem: availableVessel,
				}
			}
			// Move toward vessel
			nx, ny := NextStep(pos.X, pos.Y, vx, vy)
			newActivity := "Moving to pick up vessel"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				Target:     types.Position{X: nx, Y: ny},
				Dest:       types.Position{X: vx, Y: vy}, // Destination is the vessel's position
				Action:     entity.ActionPickup,
				TargetItem: availableVessel,
			}
		}
		// No vessel available - proceed to harvest directly (single item)
	} else if char.Carrying.Container != nil {
		// Carrying a vessel - check if it can accept the target
		if !CanVesselAccept(char.Carrying, target, registry) {
			// Vessel cannot accept target (full or variety mismatch)
			// Drop vessel since order takes priority (4e)
			Drop(char, gameMap, log)
			// Now not carrying anything - look for compatible vessel
			availableVessel := FindAvailableVessel(pos.X, pos.Y, items, target, registry)
			if availableVessel != nil {
				vpos := availableVessel.Pos()
				vx, vy := vpos.X, vpos.Y
				if pos.X == vx && pos.Y == vy {
					newActivity := "Picking up vessel"
					if char.CurrentActivity != newActivity {
						char.CurrentActivity = newActivity
					}
					return &entity.Intent{
						Target:     pos,
						Dest:       pos, // Already at destination
						Action:     entity.ActionPickup,
						TargetItem: availableVessel,
					}
				}
				nx, ny := NextStep(pos.X, pos.Y, vx, vy)
				newActivity := "Moving to pick up vessel"
				if char.CurrentActivity != newActivity {
					char.CurrentActivity = newActivity
				}
				return &entity.Intent{
					Target:     types.Position{X: nx, Y: ny},
					Dest:       types.Position{X: vx, Y: vy}, // Destination is the vessel's position
					Action:     entity.ActionPickup,
					TargetItem: availableVessel,
				}
			}
			// No compatible vessel - proceed to harvest directly
		}
		// Vessel can accept target - proceed to harvest (will add to vessel)
	} else {
		// Carrying non-vessel item - this case is handled by update.go
		// which drops items when on order before pickup
	}

	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	// Check if already at target
	if pos.X == tx && pos.Y == ty {
		// Start harvesting immediately (uses ActionPickup - same physical action as foraging)
		newActivity := "Harvesting " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos, // Already at destination
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	// Move toward target
	nx, ny := NextStep(pos.X, pos.Y, tx, ty)

	newActivity := "Moving to harvest " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}

	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty}, // Destination is the item's position
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// findCraftVesselIntent creates an intent to craft a vessel.
// If carrying a gourd, returns ActionCraft. Otherwise returns ActionPickup for a gourd.
func findCraftVesselIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog) *entity.Intent {
	// Check if already carrying a gourd - ready to craft
	if char.Carrying != nil && char.Carrying.ItemType == "gourd" {
		newActivity := "Crafting vessel"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos, // Already at destination (crafting in place)
			Action:     entity.ActionCraft,
			TargetItem: char.Carrying, // The gourd being used
		}
	}

	// Need to find a gourd to pick up
	target := findNearestItemByType(pos.X, pos.Y, items, "gourd")
	if target == nil {
		return nil // No gourds available - will trigger abandonment
	}

	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	// Check if already at target
	if pos.X == tx && pos.Y == ty {
		newActivity := "Picking up " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos, // Already at destination
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	// Move toward target
	nx, ny := NextStep(pos.X, pos.Y, tx, ty)

	newActivity := "Moving to pick up " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}

	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty}, // Destination is the item's position
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// findNearestItemByType finds the closest item of a specific type.
func findNearestItemByType(cx, cy int, items []*entity.Item, itemType string) *entity.Item {
	if len(items) == 0 {
		return nil
	}

	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range items {
		if item.ItemType != itemType {
			continue
		}
		// Only consider growing items for harvest
		if item.Plant == nil || !item.Plant.IsGrowing {
			continue
		}

		ipos := item.Pos()
		dist := pos.DistanceTo(ipos)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	return nearest
}

// abandonOrder removes an order that cannot be fulfilled and clears the character's assignment.
func abandonOrder(char *entity.Character, order *entity.Order, orders []*entity.Order, log *ActionLog) {
	if log != nil {
		log.Add(char.ID, char.Name, "order", fmt.Sprintf("Abandoning order: %s (no items available)", order.DisplayName()))
	}

	// Clear character's assignment
	char.AssignedOrderID = 0

	// Mark order for removal by setting a special status
	// The UI layer will need to clean up removed orders
	order.Status = entity.OrderOpen
	order.AssignedTo = 0

	// Note: Actually removing from the slice should be done by the caller
	// to avoid issues with slice modification during iteration
}

// findOrderByID returns the order with the given ID, or nil if not found.
func findOrderByID(orders []*entity.Order, id int) *entity.Order {
	for _, order := range orders {
		if order.ID == id {
			return order
		}
	}
	return nil
}

// CompleteOrder handles order completion (e.g., when inventory is full for Harvest).
// Removes the order and clears the character's assignment.
func CompleteOrder(char *entity.Character, order *entity.Order, log *ActionLog) {
	if log != nil {
		log.Add(char.ID, char.Name, "order", fmt.Sprintf("Completed order: %s", order.DisplayName()))
	}

	// Clear character's assignment
	char.AssignedOrderID = 0

	// Mark order as completed (will be removed by UI layer)
	// For now, just reset it - actual removal handled elsewhere
	order.Status = entity.OrderOpen
	order.AssignedTo = 0
}

// PauseOrder marks an order as paused due to character needs interruption.
func PauseOrder(order *entity.Order, log *ActionLog, charID int, charName string) {
	if order.Status == entity.OrderAssigned {
		order.Status = entity.OrderPaused
		if log != nil {
			log.Add(charID, charName, "order", fmt.Sprintf("Pausing order: %s (needs attention)", order.DisplayName()))
		}
	}
}
