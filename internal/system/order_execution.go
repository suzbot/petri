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
// Recipe-based activities (any activity with registered recipes) route to findCraftIntent.
func findOrderIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Check if this activity has recipes — route to generic craft handler
	if len(entity.GetRecipesForActivity(order.ActivityID)) > 0 {
		return findCraftIntent(char, pos, items, order, log, gameMap)
	}

	switch order.ActivityID {
	case "harvest":
		return findHarvestIntent(char, pos, items, order, log, gameMap)
	default:
		// Unknown activity type - cannot create intent
		return nil
	}
}

// findHarvestIntent creates an intent to harvest (pick up) a specific item type per order.
// Returns nil if no matching items exist on the map.
// Uses EnsureHasVesselFor to handle vessel acquisition with drop-when-blocked logic.
func findHarvestIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Find nearest item matching the order's target type
	target := findNearestItemByType(pos.X, pos.Y, items, order.TargetType, true)
	if target == nil {
		return nil // No matching items - will trigger abandonment
	}

	// Ensure we have a compatible vessel (or get one if available)
	// dropConflict=true: orders take priority, drop incompatible vessels
	if intent := EnsureHasVesselFor(char, target, items, gameMap, log, true, "order"); intent != nil {
		return intent
	}

	// Ready to harvest - either have compatible vessel or will harvest directly
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
	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)

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



// findCraftIntent creates an intent to craft an item using the recipe system.
// Generic replacement for findCraftVesselIntent — works for any recipe-based activity.
// 1. Gets recipes for the order's activity
// 2. Filters to feasible recipes (inputs exist in world)
// 3. Picks first feasible recipe (future: score by preference — see triggered-enhancements.md)
// 4. Uses EnsureHasRecipeInputs to gather components
// 5. When ready → returns ActionCraft intent with RecipeID
func findCraftIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	recipes := entity.GetRecipesForActivity(order.ActivityID)
	if len(recipes) == 0 {
		return nil
	}

	// Find a feasible recipe: at least one of each input type exists
	// (in character's accessible items or on the map)
	var feasible *entity.Recipe
	for _, recipe := range recipes {
		if isRecipeFeasible(char, recipe, items) {
			feasible = recipe
			break
		}
	}
	if feasible == nil {
		return nil
	}

	// Use EnsureHasRecipeInputs to gather components
	if intent := EnsureHasRecipeInputs(char, feasible, items, gameMap, log); intent != nil {
		return intent
	}

	// All inputs accessible — check if we're actually ready to craft
	// (EnsureHasRecipeInputs returns nil both when ready AND when impossible)
	for _, input := range feasible.Inputs {
		if !char.HasAccessibleItem(input.ItemType) {
			return nil // Impossible — input not available
		}
	}

	// Ready to craft
	newActivity := "Crafting " + feasible.Name
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target:   pos,
		Dest:     pos,
		Action:   entity.ActionCraft,
		RecipeID: feasible.ID,
	}
}

// isRecipeFeasible returns true if all recipe inputs exist somewhere accessible
// (in character inventory/vessels or on the map).
func isRecipeFeasible(char *entity.Character, recipe *entity.Recipe, items []*entity.Item) bool {
	for _, input := range recipe.Inputs {
		// Check if character already has this input
		if char.HasAccessibleItem(input.ItemType) {
			continue
		}
		// Check if input exists on the map
		if findNearestItemByType(char.X, char.Y, items, input.ItemType, false) == nil {
			return false
		}
	}
	return true
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
	target := findNearestItemByType(cx, cy, items, targetType, true)
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
