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
	carriedVessel := char.GetCarriedVessel()
	if carriedVessel == nil {
		// No vessel in inventory - look for available vessel first (4d) if we have space
		if char.HasInventorySpace() {
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
		}
		// No vessel available or no space - proceed to harvest directly (single item)
	} else {
		// Have a vessel - check if it can accept the target
		if !CanVesselAccept(carriedVessel, target, registry) {
			// Vessel cannot accept target (full or variety mismatch)
			// Drop vessel since order takes priority (4e)
			DropItem(char, carriedVessel, gameMap, log)
			// Now look for compatible vessel
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
// If carrying a gourd (in inventory or vessel contents), returns ActionCraft.
// Otherwise returns ActionPickup for a gourd.
func findCraftVesselIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog) *entity.Intent {
	// Check if already carrying a gourd in direct inventory - ready to craft
	gourd := char.FindInInventory(func(i *entity.Item) bool { return i.ItemType == "gourd" })
	if gourd != nil {
		newActivity := "Crafting vessel"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos, // Already at destination (crafting in place)
			Action:     entity.ActionCraft,
			TargetItem: gourd, // The gourd being used
		}
	}

	// Check if recipe input is inside a carried vessel - extract it first
	recipe := entity.RecipeRegistry["hollow-gourd"]
	if recipe != nil {
		extractedItem := extractRecipeInputFromVessel(char, recipe)
		if extractedItem != nil {
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Took "+extractedItem.ItemType+" from vessel for crafting")
			}
			// Now have input in inventory - ready to craft
			newActivity := "Crafting vessel"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				Target:     pos,
				Dest:       pos,
				Action:     entity.ActionCraft,
				TargetItem: extractedItem,
			}
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

// extractRecipeInputFromVessel checks if a carried vessel contains any required recipe input,
// and if so, extracts one item to character inventory. Returns the extracted item or nil.
func extractRecipeInputFromVessel(char *entity.Character, recipe *entity.Recipe) *entity.Item {
	vessel := char.GetCarriedVessel()
	if vessel == nil || vessel.Container == nil || recipe == nil {
		return nil
	}

	// Check each recipe input against vessel contents
	for _, input := range recipe.Inputs {
		for i, stack := range vessel.Container.Contents {
			if stack.Variety != nil && stack.Variety.ItemType == input.ItemType && stack.Count > 0 {
				// Found matching input - extract one
				vessel.Container.Contents[i].Count--

				// Remove empty stack
				if vessel.Container.Contents[i].Count <= 0 {
					vessel.Container.Contents = append(
						vessel.Container.Contents[:i],
						vessel.Container.Contents[i+1:]...,
					)
				}

				// Create item from variety
				item := &entity.Item{
					ItemType: stack.Variety.ItemType,
					Color:    stack.Variety.Color,
					Pattern:  stack.Variety.Pattern,
					Texture:  stack.Variety.Texture,
				}
				item.EType = entity.TypeItem

				// Add to inventory
				char.AddToInventory(item)

				return item
			}
		}
	}

	return nil
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

// FindNextHarvestTarget finds the next item to harvest for order continuation.
// Returns nil if inventory is full or no matching targets exist.
func FindNextHarvestTarget(char *entity.Character, cx, cy int, items []*entity.Item, targetType string) *entity.Intent {
	if !char.HasInventorySpace() {
		return nil
	}

	// Find nearest item matching the target type
	target := findNearestItemByType(cx, cy, items, targetType)
	if target == nil {
		return nil
	}

	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	if cx == tx && cy == ty {
		// Already at target
		return &entity.Intent{
			Target:     types.Position{X: cx, Y: cy},
			Dest:       types.Position{X: cx, Y: cy},
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	// Move toward target
	nx, ny := NextStep(cx, cy, tx, ty)
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty},
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
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
