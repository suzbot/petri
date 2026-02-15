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
			// Check if this is a tillSoil completion (pool empty) vs failure
			if isMultiStepOrderComplete(order, gameMap) {
				CompleteOrder(char, order, log)
				return nil
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
	order := findAvailableOrder(char, orders, items, gameMap)
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

	// Check if this is a tillSoil completion (pool empty) vs failure
	if isMultiStepOrderComplete(order, gameMap) {
		CompleteOrder(char, order, log)
		return nil
	}

	// Order cannot be fulfilled immediately - abandon it
	abandonOrder(char, order, orders, log)
	return nil
}

// findAvailableOrder finds the first open order that the character can execute.
// Checks activity requirements (know-how) and world feasibility (components exist).
// Future: can be extended to consider character preference, order urgency, etc.
func findAvailableOrder(char *entity.Character, orders []*entity.Order, items []*entity.Item, gameMap *game.Map) *entity.Order {
	for _, order := range orders {
		if order.Status != entity.OrderOpen {
			continue
		}
		if !canExecuteOrder(char, order) {
			continue
		}
		feasible, _ := IsOrderFeasible(order, items, gameMap)
		if !feasible {
			continue
		}
		return order
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
	case "tillSoil":
		return findTillSoilIntent(char, pos, items, order, log, gameMap)
	case "plant":
		return findPlantIntent(char, pos, items, order, log, gameMap)
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



// findTillSoilIntent creates an intent to till a marked tile from the shared pool.
// Flow: procure hoe → find nearest marked-but-not-tilled tile → move to it → till it.
// Returns nil when pool is empty (order complete) or when no hoe exists (triggers abandonment).
func findTillSoilIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Step 1: Ensure character has a hoe
	if intent := EnsureHasItem(char, "hoe", items, gameMap, log); intent != nil {
		return intent
	}

	// Check if hoe procurement failed (no hoe in inventory and none on map)
	if char.FindInInventory(func(i *entity.Item) bool { return i.ItemType == "hoe" }) == nil {
		return nil // No hoe available — triggers abandonment
	}

	// Step 2: Find nearest marked-but-not-tilled position
	marked := gameMap.MarkedForTillingPositions()
	var nearest *types.Position
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, mpos := range marked {
		if gameMap.IsTilled(mpos) {
			continue // Already tilled
		}
		dist := pos.DistanceTo(mpos)
		if dist < nearestDist {
			nearestDist = dist
			p := mpos // Copy for pointer
			nearest = &p
		}
	}

	if nearest == nil {
		return nil // Pool empty — order complete
	}

	// Step 3: Move to or till the target tile
	if pos.X == nearest.X && pos.Y == nearest.Y {
		// At target — start tilling
		char.CurrentActivity = "Tilling soil"
		return &entity.Intent{
			Target: *nearest,
			Dest:   *nearest,
			Action: entity.ActionTillSoil,
		}
	}

	// Move toward target tile
	nx, ny := NextStepBFS(pos.X, pos.Y, nearest.X, nearest.Y, gameMap)
	newActivity := "Moving to till soil"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target: types.Position{X: nx, Y: ny},
		Dest:   *nearest,
		Action: entity.ActionTillSoil,
	}
}

// findPlantIntent creates an intent to plant items on tilled soil.
// Flow: find empty tilled tile → check for accessible plantable item → procure if needed → plant.
// Returns nil when no empty tilled tiles (order complete) or no plantable items (abandon/complete).
func findPlantIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Step 1: Find nearest empty tilled tile (no item at position)
	var nearestTile *types.Position
	nearestTileDist := int(^uint(0) >> 1)

	for _, tpos := range gameMap.TilledPositions() {
		if gameMap.ItemAt(tpos) != nil {
			continue // Tile already has a plant/sprout
		}
		dist := pos.DistanceTo(tpos)
		if dist < nearestTileDist {
			nearestTileDist = dist
			p := tpos
			nearestTile = &p
		}
	}

	if nearestTile == nil {
		return nil // No empty tilled tiles — order complete
	}

	// Step 2: Check if character has accessible plantable item
	if hasAccessiblePlantable(char, order.TargetType, order.LockedVariety) {
		// Has item — move to tilled tile or plant
		if pos.X == nearestTile.X && pos.Y == nearestTile.Y {
			// At tilled tile — plant
			char.CurrentActivity = "Planting"
			return &entity.Intent{
				Target: *nearestTile,
				Dest:   *nearestTile,
				Action: entity.ActionPlant,
			}
		}

		// Move toward tilled tile
		nx, ny := NextStepBFS(pos.X, pos.Y, nearestTile.X, nearestTile.Y, gameMap)
		newActivity := "Moving to plant"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target: types.Position{X: nx, Y: ny},
			Dest:   *nearestTile,
			Action: entity.ActionPlant,
		}
	}

	// Step 3: No accessible item — use EnsureHasPlantable to acquire one
	// (handles dropping unneeded items, finding on ground, creating pickup intent)
	return EnsureHasPlantable(char, order.TargetType, order.LockedVariety, items, gameMap, log)
}

// isMultiStepOrderComplete checks if a multi-step order should be considered complete
// (vs abandoned due to failure). Returns true when the work pool is exhausted.
func isMultiStepOrderComplete(order *entity.Order, gameMap *game.Map) bool {
	switch order.ActivityID {
	case "tillSoil":
		for _, pos := range gameMap.MarkedForTillingPositions() {
			if !gameMap.IsTilled(pos) {
				return false // Still has work to do
			}
		}
		return true
	case "plant":
		// Plant order is complete when LockedVariety is set (character planted at least once)
		// and findPlantIntent returned nil (no more empty tiles or no more items).
		return order.LockedVariety != ""
	default:
		return false
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

// CompleteOrder handles order completion. Logs, unassigns the character, and marks
// the order as OrderCompleted. The game loop sweep removes completed orders.
func CompleteOrder(char *entity.Character, order *entity.Order, log *ActionLog) {
	if log != nil {
		log.Add(char.ID, char.Name, "order", fmt.Sprintf("Completed order: %s", order.DisplayName()))
	}

	char.AssignedOrderID = 0
	order.Status = entity.OrderCompleted
	order.AssignedTo = 0
}

// FindNextHarvestTarget finds the next item to harvest for order continuation.
// Returns nil if inventory is full or no matching targets exist.
func FindNextHarvestTarget(char *entity.Character, cx, cy int, items []*entity.Item, targetType string, gameMap *game.Map) *entity.Intent {
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
	nx, ny := NextStepBFS(cx, cy, tx, ty, gameMap)
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty},
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// IsOrderFeasible checks whether an order can potentially be fulfilled given current world state.
// Returns (feasible, noKnowHow) — noKnowHow is true when the failure is due to no character
// having the required know-how (vs missing components).
// This is computed on demand, not cached — cheap O(n) check over items and characters.
func IsOrderFeasible(order *entity.Order, items []*entity.Item, gameMap *game.Map) (bool, bool) {
	activity, ok := entity.ActivityRegistry[order.ActivityID]
	if !ok {
		return false, false
	}

	// Check know-how: does any living character know this activity?
	if activity.Availability == entity.AvailabilityKnowHow {
		chars := gameMap.Characters()
		knowHowExists := false
		for _, c := range chars {
			if c.KnowsActivity(order.ActivityID) {
				knowHowExists = true
				break
			}
		}
		if !knowHowExists {
			return false, true
		}
	}

	// Check components per activity type
	chars := gameMap.Characters()

	// Recipe-based activities (craft): check if any recipe's inputs all exist in world
	if len(entity.GetRecipesForActivity(order.ActivityID)) > 0 {
		return isAnyRecipeWorldFeasible(order.ActivityID, chars, items), false
	}

	switch order.ActivityID {
	case "harvest":
		return growingItemExists(items, order.TargetType), false
	case "tillSoil":
		return itemExistsInWorld("hoe", chars, items) && HasUnfilledTillingPositions(gameMap), false
	case "plant":
		return PlantableItemExists(items, chars, order.TargetType), false
	default:
		return true, false // Unknown activity type, assume feasible
	}
}

// itemExistsInWorld checks if an item of the given type exists anywhere —
// in any character's inventory/vessels or on the ground.
func itemExistsInWorld(itemType string, chars []*entity.Character, items []*entity.Item) bool {
	for _, c := range chars {
		if c.HasAccessibleItem(itemType) {
			return true
		}
	}
	for _, item := range items {
		if item.ItemType == itemType {
			return true
		}
	}
	return false
}

// PlantableItemExists checks if any plantable item matching the target exists in the world
// (on the ground, in character inventories, or inside carried vessels). Matches on ItemType
// or Kind to support both direct plantables ("berry") and seed kinds ("gourd seed").
// Vessel contents lose the Plantable flag, so we infer plantability from ItemTypeConfig.
func PlantableItemExists(items []*entity.Item, chars []*entity.Character, targetType string) bool {
	configs := game.GetItemTypeConfigs()
	for _, item := range items {
		if isPlantableMatch(item, targetType) {
			return true
		}
		// Check ground vessel contents
		if item.Container != nil {
			for _, stack := range item.Container.Contents {
				if stack.Variety != nil && stack.Count > 0 &&
					stack.Variety.ItemType == targetType {
					if cfg, ok := configs[stack.Variety.ItemType]; ok && cfg.Plantable {
						return true
					}
				}
			}
		}
	}
	for _, c := range chars {
		for _, item := range c.Inventory {
			if item == nil {
				continue
			}
			if isPlantableMatch(item, targetType) {
				return true
			}
			// Check vessel contents — infer plantability from config since
			// the Plantable flag is lost when items enter vessels
			if item.Container != nil {
				for _, stack := range item.Container.Contents {
					if stack.Variety != nil && stack.Count > 0 &&
						stack.Variety.ItemType == targetType {
						if cfg, ok := configs[stack.Variety.ItemType]; ok && cfg.Plantable {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func isPlantableMatch(item *entity.Item, targetType string) bool {
	return item.Plantable && (item.ItemType == targetType || item.Kind == targetType)
}

// growingItemExists checks if any growing item of the given type exists on the map.
func growingItemExists(items []*entity.Item, itemType string) bool {
	for _, item := range items {
		if item.ItemType == itemType && item.Plant != nil && item.Plant.IsGrowing && !item.Plant.IsSprout {
			return true
		}
	}
	return false
}

// isAnyRecipeWorldFeasible checks if any recipe for the given activity has all its
// input types present somewhere in the world (character inventories or ground items).
func isAnyRecipeWorldFeasible(activityID string, chars []*entity.Character, items []*entity.Item) bool {
	recipes := entity.GetRecipesForActivity(activityID)
	for _, recipe := range recipes {
		allInputsExist := true
		for _, input := range recipe.Inputs {
			if !itemExistsInWorld(input.ItemType, chars, items) {
				allInputsExist = false
				break
			}
		}
		if allInputsExist {
			return true
		}
	}
	return false
}

// HasUnfilledTillingPositions checks if the marked-for-tilling pool has any positions
// that haven't been tilled yet.
func HasUnfilledTillingPositions(gameMap *game.Map) bool {
	for _, pos := range gameMap.MarkedForTillingPositions() {
		if !gameMap.IsTilled(pos) {
			return true
		}
	}
	return false
}

// HasEmptyTilledTile checks if any tilled tile on the map has no item on it.
func HasEmptyTilledTile(gameMap *game.Map) bool {
	for _, tpos := range gameMap.TilledPositions() {
		if gameMap.ItemAt(tpos) == nil {
			return true
		}
	}
	return false
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
