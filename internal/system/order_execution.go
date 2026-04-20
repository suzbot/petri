package system

import (
	"fmt"
	"math"
	"sort"

	"petri/internal/config"
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
			// Check if order goal is achieved vs failure
			if isMultiStepOrderComplete(char, order, gameMap) {
				DropCompletedBundle(char, order, gameMap, log)
				DropCompletedDigItems(char, order, gameMap, log)
				DropCompletedFenceMaterials(char, order, gameMap, log)
				CompleteOrder(char, order, log)
				return nil
			}
			// Transient nil: order is still feasible but temporarily blocked (e.g., another
			// worker occupies the only remaining build tile). Idle for a tick and retry.
			if feasible, _ := IsOrderFeasible(order, items, gameMap); feasible {
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

	// Check if order goal is achieved vs failure
	if isMultiStepOrderComplete(char, order, gameMap) {
		DropCompletedBundle(char, order, gameMap, log)
		DropCompletedDigItems(char, order, gameMap, log)
		CompleteOrder(char, order, log)
		return nil
	}

	// Transient nil: order is still feasible but temporarily blocked — idle and retry
	if feasible, _ := IsOrderFeasible(order, items, gameMap); feasible {
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
// Generic recipe-based activities (craftVessel, craftHoe, craftBrick) fall through to findCraftIntent.
func findOrderIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	switch order.ActivityID {
	case "harvest":
		return findHarvestIntent(char, pos, items, order, log, gameMap)
	case "tillSoil":
		return findTillSoilIntent(char, pos, items, order, log, gameMap)
	case "plant":
		return findPlantIntent(char, pos, items, order, log, gameMap)
	case "waterGarden":
		return findWaterGardenIntent(char, pos, items, order, log, gameMap)
	case "gather":
		return findGatherIntent(char, pos, items, order, log, gameMap)
	case "extract":
		return findExtractIntent(char, pos, items, order, log, gameMap)
	case "dig":
		return findDigIntent(char, pos, items, order, log, gameMap)
	case "buildFence":
		return findBuildFenceIntent(char, pos, items, order, log, gameMap)
	case "buildHut":
		return findBuildHutIntent(char, pos, items, order, log, gameMap)
	default:
		// Recipe-based activities (craftVessel, craftHoe, craftBrick, etc.) use generic craft handler
		if len(entity.GetRecipesForActivity(order.ActivityID)) > 0 {
			return findCraftIntent(char, pos, items, order, log, gameMap)
		}
		return nil
	}
}

// findHarvestIntent creates an intent to harvest (pick up) a specific item type per order.
// Returns nil if no matching items exist on the map.
// Vessel-excluded types (grass) skip vessel procurement and go straight to pickup.
// Non-excluded types (berry) use EnsureHasVesselFor for vessel acquisition.
func findHarvestIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Full-bundle safety net: if character already has a full bundle, signal completion
	if hasFullBundle(char, order.TargetType) {
		return nil
	}

	// Find nearest item matching the order's target type
	target := findNearestItemByType(pos.X, pos.Y, items, order.TargetType, true)
	if target == nil {
		return nil // No matching items - will trigger abandonment
	}

	// Only do vessel procurement for non-vessel-excluded types
	if !config.VesselExcludedTypes[target.ItemType] {
		// dropConflict=true: orders take priority, drop incompatible vessels
		if intent := EnsureHasVesselFor(char, target, items, gameMap, log, true, "order"); intent != nil {
			return intent
		}
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

	// Move toward target tile (use sticky BFS — position-based orders recalculate
	// each tick, so the character needs BFS to persist across recalculations)
	nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, nearest.X, nearest.Y, gameMap, char.UsingBFS)
	if usedBFS {
		char.UsingBFS = true
	}
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
	// Step 1: Find nearest plantable tilled tile (no growing plant at position)
	// Tiles with loose non-growing items are still available — items get pushed aside during planting.
	var nearestTile *types.Position
	nearestTileDist := int(^uint(0) >> 1)

	for _, tpos := range gameMap.TilledPositions() {
		if tileHasGrowingPlant(tpos, gameMap) {
			continue // Tile already has a growing plant or sprout
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

		// Move toward tilled tile (use sticky BFS — position-based orders recalculate each tick)
		nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, nearestTile.X, nearestTile.Y, gameMap, char.UsingBFS)
		if usedBFS {
			char.UsingBFS = true
		}
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
// (vs abandoned due to failure). Returns true when the order's goal is achieved.
func isMultiStepOrderComplete(char *entity.Character, order *entity.Order, gameMap *game.Map) bool {
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
	case "waterGarden":
		// Complete when no dry tilled planted tiles remain
		return !DryTilledPlantedTileExists(gameMap.Items(), gameMap)
	case "harvest":
		// Bundleable harvest (grass) — complete when character has a full bundle
		return hasFullBundle(char, order.TargetType)
	case "gather":
		// One bundle per order — complete when character has a full bundle of the target type
		return hasFullBundle(char, order.TargetType)
	case "extract":
		if order.LockedVariety == "" {
			return false
		}
		// Complete when no more extractable plants of that variety exist
		if !extractableItemExists(gameMap.Items(), order.TargetType, order.LockedVariety) {
			return true
		}
		// Safety net: complete when inventory full and no vessel can accept seeds
		return !char.HasInventorySpace() && char.GetCarriedVessel() == nil
	case "dig":
		// Complete when both inventory slots have clay
		clayCount := 0
		for _, item := range char.Inventory {
			if item != nil && item.ItemType == "clay" {
				clayCount++
			}
		}
		return clayCount >= 2
	case "craftBrick":
		// Complete when no clay items remain on the ground
		return !groundItemOfTypeExists(gameMap.Items(), "clay")
	case "buildFence":
		return !gameMap.HasUnbuiltConstructionPositions("fence")
	case "buildHut":
		return !gameMap.HasUnbuiltConstructionPositions("hut")
	default:
		return false
	}
}

// hasFullBundle returns true if the character is carrying a full bundle of the given type.
func hasFullBundle(char *entity.Character, targetType string) bool {
	maxSize := config.MaxBundleSize[targetType]
	if maxSize == 0 {
		return false
	}
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == targetType && item.BundleCount >= maxSize {
			return true
		}
	}
	return false
}

// DropCompletedBundle drops a full bundle of the order's target type when a
// harvest or gather order completes. Called from selectOrderActivity before
// CompleteOrder. No-op for other order types or when no full bundle exists.
func DropCompletedBundle(char *entity.Character, order *entity.Order, gameMap *game.Map, log *ActionLog) {
	if order.ActivityID != "gather" && order.ActivityID != "harvest" {
		return
	}
	maxSize := config.MaxBundleSize[order.TargetType]
	if maxSize == 0 {
		return
	}
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == order.TargetType && item.BundleCount >= maxSize {
			DropItem(char, item, gameMap, log)
			return
		}
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

// abandonOrder marks an order as abandoned with a cooldown and clears the character's assignment.
func abandonOrder(char *entity.Character, order *entity.Order, orders []*entity.Order, log *ActionLog) {
	if log != nil {
		log.Add(char.ID, char.Name, "order", fmt.Sprintf("Abandoning order: %s (no items available)", order.DisplayName()))
	}

	// Clear character's assignment
	char.AssignedOrderID = 0

	// Set abandoned status with cooldown — prevents take/abandon spam
	order.Status = entity.OrderAbandoned
	order.AssignedTo = 0
	order.AbandonCooldown = config.OrderAbandonCooldown
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
// Bundle-aware: for bundleable types, uses canGatherMore (non-full bundle check)
// instead of HasInventorySpace. Returns nil if no capacity or no matching targets.
func FindNextHarvestTarget(char *entity.Character, cx, cy int, items []*entity.Item, targetType string, gameMap *game.Map) *entity.Intent {
	if config.MaxBundleSize[targetType] > 0 {
		if !canGatherMore(char, targetType) {
			return nil
		}
	} else {
		if !char.HasInventorySpace() {
			return nil
		}
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

	// Construction activities use recipes for discovery only — check construction-specific feasibility
	// before the generic recipe check intercepts it.
	if order.ActivityID == "buildFence" {
		return gameMap.HasUnbuiltConstructionPositions("fence") && constructionMaterialFeasible("fence", gameMap), false
	}
	if order.ActivityID == "buildHut" {
		return gameMap.HasUnbuiltConstructionPositions("hut") && constructionMaterialFeasible("hut", gameMap), false
	}

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
		return PlantableItemExists(items, chars, order.TargetType) && plantableTilledTileExists(gameMap), false
	case "waterGarden":
		return waterGardenFeasible(chars, items, gameMap), false
	case "gather":
		return groundItemOfTypeExists(items, order.TargetType), false
	case "extract":
		return extractableItemExists(items, order.TargetType, ""), false
	case "dig":
		return gameMap.HasClay(), false
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
func PlantableItemExists(items []*entity.Item, chars []*entity.Character, targetType string) bool {
	for _, item := range items {
		if isPlantableMatch(item, targetType) {
			return true
		}
		// Check ground vessel contents
		if item.Container != nil {
			for _, stack := range item.Container.Contents {
				if isPlantableVarietyMatch(stack.Variety, stack.Count, targetType) {
					return true
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
			// Check vessel contents
			if item.Container != nil {
				for _, stack := range item.Container.Contents {
					if isPlantableVarietyMatch(stack.Variety, stack.Count, targetType) {
						return true
					}
				}
			}
		}
	}
	return false
}

// isPlantableVarietyMatch checks if a vessel stack variety matches the target type.
// Mirrors isPlantableMatch but for varieties: checks Plantable flag and matches on
// either ItemType or Kind.
func isPlantableVarietyMatch(variety *entity.ItemVariety, count int, targetType string) bool {
	return variety != nil && count > 0 && variety.Plantable &&
		(variety.ItemType == targetType || variety.Kind == targetType)
}

func isPlantableMatch(item *entity.Item, targetType string) bool {
	return item.Plantable && (item.ItemType == targetType || item.Kind == targetType)
}

// tileHasGrowingPlant checks if any item at the position is a growing plant.
// Used to determine if a tilled tile is available for planting — tiles with only
// non-growing loose items (seeds, vessels) are still plantable.
func tileHasGrowingPlant(pos types.Position, gameMap *game.Map) bool {
	for _, item := range gameMap.ItemsAt(pos) {
		if item.Plant != nil && item.Plant.IsGrowing {
			return true
		}
	}
	return false
}

// plantableTilledTileExists checks if at least one tilled tile has no growing plant.
// Used by IsOrderFeasible to avoid the take-abandon spam loop when seeds exist
// but all tilled tiles are occupied by growing plants.
func plantableTilledTileExists(gameMap *game.Map) bool {
	for _, tpos := range gameMap.TilledPositions() {
		if !tileHasGrowingPlant(tpos, gameMap) {
			return true
		}
	}
	return false
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

// extractableItemExists checks if any growing non-sprout plant of the given type
// has seeds available (SeedTimer <= 0).
// If lockedVariety is non-empty, only plants matching that variety ID are considered.
func extractableItemExists(items []*entity.Item, itemType string, lockedVariety string) bool {
	for _, item := range items {
		if item.ItemType != itemType || item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout || item.Plant.SeedTimer > 0 {
			continue
		}
		if lockedVariety != "" {
			vid := entity.GenerateVarietyID(item.ItemType, item.Kind, item.Color, item.Pattern, item.Texture)
			if vid != lockedVariety {
				continue
			}
		}
		return true
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
	return plantableTilledTileExists(gameMap)
}

// waterGardenFeasible checks if Water Garden can be fulfilled:
// (1) vessel exists in world, (2) water exists on map, (3) at least one dry tilled planted tile.
func waterGardenFeasible(chars []*entity.Character, items []*entity.Item, gameMap *game.Map) bool {
	if !itemExistsInWorld("vessel", chars, items) {
		return false
	}
	if len(gameMap.WaterPositions()) == 0 {
		return false
	}
	return DryTilledPlantedTileExists(items, gameMap)
}

// findWaterGardenIntent creates an intent for watering garden tiles.
// Full phase detection: Phase 1 (procure vessel), Phase 2 (fill at water), Phase 3 (water tiles).
// Returns nil if no vessel available anywhere or no dry tilled planted tiles.
func findWaterGardenIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Check for dry tilled planted tiles first — if none, order is complete
	target := FindNearestDryTilledPlanted(pos, items, gameMap)
	if target == nil {
		return nil // No dry tiles — order complete
	}

	// Phase 3: vessel with water in inventory → water tiles
	vessel := findCarriedVesselWithWater(char)
	if vessel != nil {
		if pos.X == target.X && pos.Y == target.Y {
			char.CurrentActivity = "Watering garden"
			return &entity.Intent{
				Target:     *target,
				Dest:       *target,
				Action:     entity.ActionWaterGarden,
				TargetItem: vessel,
			}
		}
		nx, ny := NextStepBFS(pos.X, pos.Y, target.X, target.Y, gameMap)
		newActivity := "Moving to water garden"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     types.Position{X: nx, Y: ny},
			Dest:       *target,
			Action:     entity.ActionWaterGarden,
			TargetItem: vessel,
		}
	}

	// Phase 2: empty vessel in inventory → fill at water source
	carriedVessel := char.GetCarriedVessel()
	if carriedVessel != nil {
		waterPos, found := gameMap.FindNearestWater(pos)
		if !found {
			return nil // No water reachable — abandon
		}
		adjX, adjY := FindClosestCardinalTile(pos.X, pos.Y, waterPos.X, waterPos.Y, gameMap)
		if adjX == -1 {
			return nil // Water blocked — abandon
		}
		dest := types.Position{X: adjX, Y: adjY}
		nx, ny := NextStepBFS(pos.X, pos.Y, adjX, adjY, gameMap)
		newActivity := "Fetching water for garden"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     types.Position{X: nx, Y: ny},
			Dest:       dest,
			Action:     entity.ActionWaterGarden,
			TargetItem: carriedVessel,
		}
	}

	// Phase 1: no vessel → procure ground vessel
	if !char.HasInventorySpace() {
		return nil // No space to pick up vessel — abandon
	}

	// Prefer ground water vessel — already filled, fewer steps than empty vessel + fill phase
	if waterVessel := findGroundWaterVessel(pos, items); waterVessel != nil {
		vpos := waterVessel.Pos()
		nx, ny := NextStepBFS(pos.X, pos.Y, vpos.X, vpos.Y, gameMap)
		newActivity := "Getting water for garden"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     types.Position{X: nx, Y: ny},
			Dest:       vpos,
			Action:     entity.ActionWaterGarden,
			TargetItem: waterVessel,
		}
	}

	// Fall back to ground empty vessel (needs fill phase)
	groundVessel := findEmptyGroundVessel(pos, items)
	if groundVessel == nil {
		return nil // No vessel available anywhere — abandon
	}
	vpos := groundVessel.Pos()
	nx, ny := NextStepBFS(pos.X, pos.Y, vpos.X, vpos.Y, gameMap)
	newActivity := "Getting vessel for garden"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       vpos,
		Action:     entity.ActionWaterGarden,
		TargetItem: groundVessel,
	}
}

// findCarriedVesselWithWater returns the first vessel in the character's inventory
// that contains liquid (water). Returns nil if no such vessel exists.
func findCarriedVesselWithWater(char *entity.Character) *entity.Item {
	for _, item := range char.Inventory {
		if item == nil || item.Container == nil {
			continue
		}
		for _, stack := range item.Container.Contents {
			if stack.Variety != nil && stack.Variety.ItemType == "liquid" && stack.Count > 0 {
				return item
			}
		}
	}
	return nil
}

// FindWaterGardenIntentForTest is an exported wrapper for integration tests in other packages.
func FindWaterGardenIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findWaterGardenIntent(char, pos, items, order, log, gameMap)
}

// FindNearestDryTilledPlanted finds the nearest tilled tile with a growing plant that is not wet.
func FindNearestDryTilledPlanted(pos types.Position, items []*entity.Item, gameMap *game.Map) *types.Position {
	var nearest *types.Position
	nearestDist := int(^uint(0) >> 1)

	for _, item := range items {
		if item.Plant == nil || !item.Plant.IsGrowing {
			continue
		}
		ipos := item.Pos()
		if !gameMap.IsTilled(ipos) || gameMap.IsWet(ipos) {
			continue
		}
		dist := pos.DistanceTo(ipos)
		if dist < nearestDist {
			nearestDist = dist
			p := ipos
			nearest = &p
		}
	}
	return nearest
}

// DryTilledPlantedTileExists returns true if any tilled tile has a growing plant and is not wet.
func DryTilledPlantedTileExists(items []*entity.Item, gameMap *game.Map) bool {
	for _, item := range items {
		if item.Plant == nil || !item.Plant.IsGrowing {
			continue
		}
		pos := item.Pos()
		if gameMap.IsTilled(pos) && !gameMap.IsWet(pos) {
			return true
		}
	}
	return false
}

// findGatherIntent creates an intent to gather (pick up) a specific item type per order.
// Follows findHarvestIntent with two differences:
//   - Uses growingOnly=false to find non-plant ground items
//   - Checks variety registry: items with a variety use vessel procurement; items without (sticks)
//     check inventory space directly.
func findGatherIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// One bundle per order — if character already has a full bundle, order goal is achieved
	if hasFullBundle(char, order.TargetType) {
		return nil
	}

	// Find nearest item matching the order's target type (growingOnly=false)
	target := findNearestItemByType(pos.X, pos.Y, items, order.TargetType, false)
	if target == nil {
		return nil // No matching items - will trigger abandonment
	}

	// Vessel-excluded types skip vessel procurement regardless of variety
	if config.VesselExcludedTypes[target.ItemType] {
		// Bundleable or vessel-excluded — direct pickup, check capacity
		if !canGatherMore(char, order.TargetType) && !hasNonTargetToDrop(char, order.TargetType) {
			return nil
		}
	} else {
		// Check if item has a registered variety (determines vessel vs. direct inventory path)
		registry := gameMap.Varieties()
		variety := registry.GetByAttributes(target.ItemType, target.Kind, target.Color, target.Pattern, target.Texture)

		if variety != nil {
			// Item has variety — use vessel procurement (same as harvest)
			if intent := EnsureHasVesselFor(char, target, items, gameMap, log, true, "order"); intent != nil {
				return intent
			}
		} else {
			// No variety — must have room to pick up (inventory slot or non-full bundle).
			if !canGatherMore(char, order.TargetType) && !hasNonTargetToDrop(char, order.TargetType) {
				return nil
			}
		}
	}

	// Ready to gather
	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	if pos.X == tx && pos.Y == ty {
		newActivity := "Gathering " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos,
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)
	newActivity := "Moving to gather " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty},
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// FindGatherIntentForTest is an exported wrapper for integration tests in other packages.
func FindGatherIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findGatherIntent(char, pos, items, order, log, gameMap)
}

// FindHarvestIntentForTest is an exported wrapper for integration tests in other packages.
func FindHarvestIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findHarvestIntent(char, pos, items, order, log, gameMap)
}

// IsMultiStepOrderCompleteForTest is an exported wrapper for tests.
func IsMultiStepOrderCompleteForTest(char *entity.Character, order *entity.Order, gameMap *game.Map) bool {
	return isMultiStepOrderComplete(char, order, gameMap)
}

// FindCraftIntentForTest is an exported wrapper for integration tests in other packages.
func FindCraftIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findCraftIntent(char, pos, items, order, log, gameMap)
}

// FindNextGatherTarget finds the next item to gather for order continuation.
// Returns nil if no capacity (inventory slot or non-full bundle of target type) or no matching targets exist.
func FindNextGatherTarget(char *entity.Character, cx, cy int, items []*entity.Item, targetType string, gameMap *game.Map) *entity.Intent {
	// One bundle per order — stop continuation when a full bundle exists
	if hasFullBundle(char, targetType) {
		return nil
	}
	if !canGatherMore(char, targetType) {
		return nil
	}

	target := findNearestItemByType(cx, cy, items, targetType, false)
	if target == nil {
		return nil
	}

	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	if cx == tx && cy == ty {
		return &entity.Intent{
			Target:     types.Position{X: cx, Y: cy},
			Dest:       types.Position{X: cx, Y: cy},
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	nx, ny := NextStepBFS(cx, cy, tx, ty, gameMap)
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty},
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// canGatherMore returns true if the character can pick up more items of the given type.
// Checks inventory space first, then checks for non-full bundles of the target type.
// This is target-type-aware unlike CanPickUpMore (which counts vessel space that
// vessel-excluded items like sticks can't use).
func canGatherMore(char *entity.Character, targetType string) bool {
	if char.HasInventorySpace() {
		return true
	}
	maxSize := config.MaxBundleSize[targetType]
	if maxSize > 0 {
		for _, item := range char.Inventory {
			if item == nil {
				continue
			}
			if item.ItemType == targetType && item.BundleCount > 0 && item.BundleCount < maxSize {
				return true
			}
		}
	}
	return false
}

// hasNonTargetToDrop returns true if the character has an inventory item that isn't
// the target type (and could be dropped to make room for the target).
// Used by findGatherIntent to allow order-assigned characters to proceed even with
// full inventory — applyPickup handles the drop when they arrive.
func hasNonTargetToDrop(char *entity.Character, targetType string) bool {
	for _, item := range char.Inventory {
		if item == nil {
			continue
		}
		if item.ItemType != targetType {
			return true
		}
	}
	return false
}

// groundItemOfTypeExists checks if any gatherable item of the given type exists on the ground.
// Full bundles are excluded — they're finished products, not raw material.
func groundItemOfTypeExists(items []*entity.Item, itemType string) bool {
	for _, item := range items {
		if item.ItemType != itemType {
			continue
		}
		// Skip full bundles
		if maxBundle := config.MaxBundleSize[item.ItemType]; maxBundle > 0 && item.BundleCount >= maxBundle {
			continue
		}
		return true
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

// findExtractIntent creates an intent to extract seeds from a living plant.
// Follows the walk-then-act pattern with vessel procurement.
// Returns nil if no extractable targets are available.
func findExtractIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Find nearest extractable plant: growing, non-sprout, matching target type, SeedTimer <= 0
	target := findNearestExtractable(pos.X, pos.Y, items, order.TargetType, order.LockedVariety)
	if target == nil {
		return nil
	}

	// Vessel procurement: create synthetic seed for compatibility checking
	sourceVarietyID := entity.GenerateVarietyID(target.ItemType, target.Kind, target.Color, target.Pattern, target.Texture)
	syntheticSeed := entity.NewSeed(0, 0, target.ItemType, sourceVarietyID, target.Kind, target.Color, target.Pattern, target.Texture)
	if intent := EnsureHasVesselFor(char, syntheticSeed, items, gameMap, log, true, "extract"); intent != nil {
		return intent
	}

	// Ready to extract — walk-then-act pattern
	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	// Check if already at target
	if pos == tpos {
		newActivity := fmt.Sprintf("Extracting %s seeds", target.ItemType)
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       tpos,
			Action:     entity.ActionExtract,
			TargetItem: target,
		}
	}

	// Move toward target — calculate BFS step
	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)

	newActivity := fmt.Sprintf("Moving to extract from %s", target.Description())
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}

	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       tpos,
		Action:     entity.ActionExtract,
		TargetItem: target,
	}
}

// findNearestExtractable finds the nearest growing non-sprout plant of the given type
// with SeedTimer <= 0 (seeds available for extraction).
// If lockedVariety is non-empty, only plants matching that variety ID are considered.
func findNearestExtractable(cx, cy int, items []*entity.Item, itemType string, lockedVariety string) *entity.Item {
	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range items {
		if item.ItemType != itemType {
			continue
		}
		if item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
			continue
		}
		if item.Plant.SeedTimer > 0 {
			continue
		}
		if lockedVariety != "" {
			vid := entity.GenerateVarietyID(item.ItemType, item.Kind, item.Color, item.Pattern, item.Texture)
			if vid != lockedVariety {
				continue
			}
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

// findDigIntent creates an intent to dig clay from a clay terrain tile.
// Walk-then-act pattern (follows applyExtract). No vessel procurement.
// Step 1: If both inventory slots have clay, return nil (triggers completion).
// Step 2: Drop all non-clay inventory items (procurement drop pattern).
// Step 3: Find nearest clay tile. Return nil if none (triggers abandonment).
// Step 4: Return ActionDig intent targeting the clay tile.
func findDigIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Step 1: If both slots have clay, return nil to trigger completion
	clayCount := 0
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == "clay" {
			clayCount++
		}
	}
	if clayCount >= 2 {
		return nil
	}

	// Step 2: Drop all non-clay inventory items (procurement drop pattern)
	var toDrop []*entity.Item
	for _, item := range char.Inventory {
		if item != nil && item.ItemType != "clay" {
			toDrop = append(toDrop, item)
		}
	}
	for _, item := range toDrop {
		DropItem(char, item, gameMap, log)
	}

	// Step 3: Find nearest clay tile
	clayPos, found := gameMap.FindNearestClay(pos)
	if !found {
		return nil // No clay tiles — triggers abandonment
	}

	// Step 4: Walk-then-act
	if pos == clayPos {
		char.CurrentActivity = "Digging clay"
		return &entity.Intent{
			Target: pos,
			Dest:   clayPos,
			Action: entity.ActionDig,
		}
	}

	// Use sticky BFS — position-based orders recalculate each tick
	nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, clayPos.X, clayPos.Y, gameMap, char.UsingBFS)
	if usedBFS {
		char.UsingBFS = true
	}
	newActivity := "Moving to dig clay"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target: types.Position{X: nx, Y: ny},
		Dest:   clayPos,
		Action: entity.ActionDig,
	}
}

// DropCompletedFenceMaterials drops any remaining fence materials from inventory when a
// buildFence order completes. Called from selectOrderActivity before CompleteOrder.
func DropCompletedFenceMaterials(char *entity.Character, order *entity.Order, gameMap *game.Map, log *ActionLog) {
	if order.ActivityID != "buildFence" {
		return
	}
	fenceMaterials := map[string]bool{"grass": true, "stick": true, "brick": true}
	var toDrop []*entity.Item
	for _, item := range char.Inventory {
		if item != nil && fenceMaterials[item.ItemType] {
			toDrop = append(toDrop, item)
		}
	}
	for _, item := range toDrop {
		DropItem(char, item, gameMap, log)
	}
}

// DropCompletedDigItems drops all clay items from inventory when a dig order completes.
// Called from selectOrderActivity before CompleteOrder. No-op for other order types.
func DropCompletedDigItems(char *entity.Character, order *entity.Order, gameMap *game.Map, log *ActionLog) {
	if order.ActivityID != "dig" {
		return
	}
	// Collect first, then drop (avoid modifying inventory during iteration)
	var toDrop []*entity.Item
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == "clay" {
			toDrop = append(toDrop, item)
		}
	}
	for _, item := range toDrop {
		DropItem(char, item, gameMap, log)
	}
}

// FindDigIntentForTest is an exported wrapper for integration tests in other packages.
func FindDigIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findDigIntent(char, pos, items, order, log, gameMap)
}

// =============================================================================
// Build Fence
// =============================================================================

// findBuildFenceIntent creates an intent to build a fence on a marked tile.
// Phase detection flow (position-based ordered action, re-evaluated each tick):
// 1. Find nearest unbuilt marked tile not occupied by a character (DD-28).
// 2. Determine material for that tile's line; stamp if not yet set.
// 3. Drop non-material inventory items.
// 4. If character has full bundle → build phase (find adjacent standing tile).
// 5. Otherwise → procurement phase (find nearest material item).
func findBuildFenceIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Step 1: Collect unbuilt fence-marked tiles not occupied by a character (DD-28, DD-49)
	var candidates []types.Position
	for _, mpos := range gameMap.MarkedForConstructionPositions() {
		mark, ok := gameMap.GetConstructionMark(mpos)
		if !ok || mark.ConstructKind != "fence" {
			continue // Only fence marks (DD-49)
		}
		if gameMap.ConstructAt(mpos) != nil {
			continue // Already built
		}
		if occ := gameMap.CharacterAt(mpos); occ != nil && occ != char {
			continue // Occupied by another character — skip per DD-28
		}
		candidates = append(candidates, mpos)
	}
	if len(candidates) == 0 {
		return nil // No work → triggers completion check
	}
	sort.Slice(candidates, func(i, j int) bool {
		return pos.DistanceTo(candidates[i]) < pos.DistanceTo(candidates[j])
	})
	nearest := candidates[0]

	// Step 2: Determine material for the nearest tile's line
	mark, _ := gameMap.GetConstructionMark(nearest)
	material := mark.Material
	if material == "" {
		material = selectConstructionMaterial(char, pos, items, gameMap, "buildFence")
		if material == "" {
			return nil // No material available → triggers abandonment
		}
		gameMap.SetLineMaterial(mark.LineID, material)
	}

	// Step 3: Drop non-material inventory items (procurement drop pattern)
	var toDrop []*entity.Item
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType != material {
			toDrop = append(toDrop, inv)
		}
	}
	for _, item := range toDrop {
		DropItem(char, item, gameMap, log)
	}

	// Branch: brick (non-bundle) materials use supply-drop pattern (DD-23)
	_, isBundleMaterial := config.MaxBundleSize[material]
	if !isBundleMaterial {
		return findBrickFenceIntent(char, pos, items, nearest, candidates, material, gameMap, log)
	}

	// Step 4: Build phase — character has a full bundle
	if hasFullBundle(char, material) {
		return findBundleBuildIntent(char, pos, candidates, material, gameMap)
	}

	// Step 5: Procurement phase
	// Only search for full/partial bundles when character has no bundle yet (DD-30 overflow guard)
	hasPartialBundle := false
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material && inv.BundleCount > 0 {
			hasPartialBundle = true
			break
		}
	}

	var target *entity.Item
	if !hasPartialBundle {
		target = findNearestBundleByType(pos.X, pos.Y, items, material)
	}
	if target == nil {
		// Fallback: individual items (findNearestItemByType skips full bundles)
		target = findNearestItemByType(pos.X, pos.Y, items, material, false)
	}
	if target == nil {
		return nil // No materials → triggers abandonment
	}

	return createItemPickupIntent(char, pos, target, gameMap, log)
}

// findBundleBuildIntent returns a build intent for bundle materials (grass/sticks).
// Extracted from findBuildFenceIntent for clarity.
func findBundleBuildIntent(char *entity.Character, pos types.Position, candidates []types.Position, material string, gameMap *game.Map) *entity.Intent {
	for _, candidate := range candidates {
		cm, _ := gameMap.GetConstructionMark(candidate)
		if cm.Material != "" && cm.Material != material {
			continue // Different line with different material
		}
		adjPos := findAdjacentStandingTile(candidate, gameMap)
		if adjPos == nil {
			continue // All adjacent tiles blocked — try next candidate
		}
		buildPos := candidate
		nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, adjPos.X, adjPos.Y, gameMap, char.UsingBFS)
		if usedBFS {
			char.UsingBFS = true
		}
		newActivity := "Building fence"
		if pos != *adjPos {
			newActivity = "Moving to build fence"
		}
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:         types.Position{X: nx, Y: ny},
			Dest:           *adjPos,
			Action:         entity.ActionBuildFence,
			TargetBuildPos: &buildPos,
		}
	}
	return nil // No viable candidate (all adjacent tiles blocked)
}

// findBrickFenceIntent handles the brick supply-drop pattern (DD-23).
// Brick materials are not bundled — characters carry 2 at a time, drop at build site, repeat.
func findBrickFenceIntent(char *entity.Character, pos types.Position, items []*entity.Item, nearest types.Position, candidates []types.Position, material string, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Phase 1: Check build-ready — 6+ bricks at target build position
	bricksAtSite := countItemsAtPosition(items, material, nearest)
	if bricksAtSite >= 6 {
		return findBundleBuildIntent(char, pos, candidates, material, gameMap)
	}

	// Phase 2: Delivery — character has bricks in inventory → deliver to build site
	hasBricks := false
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material {
			hasBricks = true
			break
		}
	}
	if hasBricks {
		buildPos := nearest
		nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, buildPos.X, buildPos.Y, gameMap, char.UsingBFS)
		if usedBFS {
			char.UsingBFS = true
		}
		newActivity := "Delivering materials"
		if pos == buildPos {
			newActivity = "Dropping materials"
		}
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:         types.Position{X: nx, Y: ny},
			Dest:           buildPos,
			Action:         entity.ActionBuildFence,
			TargetBuildPos: &buildPos,
		}
	}

	// Phase 3: Pickup — find nearest brick NOT at a construction site
	target := findNearestMaterialNotAtSite(pos, items, material, gameMap, false)
	if target == nil {
		return nil // No bricks → triggers abandonment
	}
	return createItemPickupIntent(char, pos, target, gameMap, log)
}

// findNearestMaterialNotAtSite finds the nearest item of the given type that is NOT
// on a construction-marked tile. If bundlesOnly is true, only matches items with BundleCount > 0.
// Used by all supply-drop procurement phases (brick fence, brick hut, bundle hut).
func findNearestMaterialNotAtSite(pos types.Position, items []*entity.Item, material string, gameMap *game.Map, bundlesOnly bool) *entity.Item {
	var target *entity.Item
	bestDist := int(^uint(0) >> 1)
	for _, item := range items {
		if item.ItemType != material {
			continue
		}
		if bundlesOnly && item.BundleCount == 0 {
			continue
		}
		if gameMap.IsMarkedForConstruction(item.Pos()) {
			continue
		}
		d := pos.DistanceTo(item.Pos())
		if d < bestDist {
			bestDist = d
			target = item
		}
	}
	return target
}

// countItemsAtPosition counts items of a specific type at a given position.
func countItemsAtPosition(items []*entity.Item, itemType string, pos types.Position) int {
	count := 0
	for _, item := range items {
		if item.ItemType == itemType && item.Pos() == pos {
			count++
		}
	}
	return count
}

// selectConstructionMaterial picks the best available construction material for a character.
// Scores each known recipe by preference (via synthetic Construct) and distance (DD-52).
// Returns the material type of the highest-scoring feasible recipe, or "" if none available.
func selectConstructionMaterial(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, activityID string) string {
	_ = gameMap // reserved for future proximity-to-map checks
	recipes := char.GetKnownRecipesForActivity(activityID)
	bestMaterial := ""
	bestScore := -math.MaxFloat64

	for _, recipe := range recipes {
		if len(recipe.Inputs) == 0 {
			continue
		}
		itemType := recipe.Inputs[0].ItemType

		if countItemsOnMap(items, itemType) < 1 {
			continue
		}
		nearest := findNearestItemOrBundle(pos.X, pos.Y, items, itemType)
		if nearest == nil {
			continue
		}

		synthetic := &entity.Construct{
			Kind:          recipe.Output.ItemType,
			Material:      itemType,
			MaterialColor: nearest.Color,
		}
		weightedPref := ScoreConstructPreference(char, synthetic)
		dist := pos.DistanceTo(nearest.Pos())
		score := ScoreItemFit(weightedPref, dist)

		if score > bestScore {
			bestScore = score
			bestMaterial = itemType
		}
	}
	return bestMaterial
}

// countItemsOnMap counts the total number of items of a given type on the map,
// summing BundleCount for bundleable items.
func countItemsOnMap(items []*entity.Item, itemType string) int {
	total := 0
	for _, item := range items {
		if item.ItemType != itemType {
			continue
		}
		if item.BundleCount > 0 {
			total += item.BundleCount
		} else {
			total++
		}
	}
	return total
}

// findNearestBundleByType finds the nearest item of a given type with BundleCount > 0,
// including full bundles (unlike findNearestItemByType which skips full bundles).
func findNearestBundleByType(cx, cy int, items []*entity.Item, itemType string) *entity.Item {
	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1)

	for _, item := range items {
		if item.ItemType != itemType || item.BundleCount <= 0 {
			continue
		}
		dist := pos.DistanceTo(item.Pos())
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}
	return nearest
}

// findNearestItemOrBundle finds the nearest item of a given type regardless of bundle state.
// Used for material availability proximity checks.
func findNearestItemOrBundle(cx, cy int, items []*entity.Item, itemType string) *entity.Item {
	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1)

	for _, item := range items {
		if item.ItemType != itemType {
			continue
		}
		dist := pos.DistanceTo(item.Pos())
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}
	return nearest
}

// findAdjacentStandingTile finds an empty cardinal tile adjacent to buildPos where
// a character can stand (not blocked, not occupied by another character).
// Returns nil if all adjacent tiles are blocked.
func findAdjacentStandingTile(buildPos types.Position, gameMap *game.Map) *types.Position {
	cardinalDirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for _, dir := range cardinalDirs {
		adj := types.Position{X: buildPos.X + dir[0], Y: buildPos.Y + dir[1]}
		if gameMap.IsBlocked(adj) {
			continue
		}
		if gameMap.CharacterAt(adj) != nil {
			continue
		}
		return &adj
	}
	return nil
}

// constructionMaterialExistsOnMap returns true if any construction material (grass, stick, or brick)
// exists on the map. Used by both fence and hut feasibility checks (DD-44).
// constructionMaterialFeasible checks whether free (non-staged) construction materials
// exist that match the locked material on marked tiles. If lines are unlocked (no material
// stamped yet), any free construction material counts.
func constructionMaterialFeasible(constructKind string, gameMap *game.Map) bool {
	// Collect distinct locked materials from marks of this kind
	lockedMaterials := make(map[string]bool)
	hasUnlockedLine := false
	for _, pos := range gameMap.MarkedForConstructionPositions() {
		mark, _ := gameMap.GetConstructionMark(pos)
		if mark.ConstructKind != constructKind {
			continue
		}
		if mark.Material == "" {
			hasUnlockedLine = true
		} else {
			lockedMaterials[mark.Material] = true
		}
	}

	items := gameMap.Items()
	for _, item := range items {
		switch item.ItemType {
		case "grass", "stick", "brick":
			if gameMap.IsMarkedForConstruction(item.Pos()) {
				continue // staged — don't count
			}
			// Free item: check if it matches a locked material or any unlocked line exists
			if lockedMaterials[item.ItemType] || hasUnlockedLine {
				return true
			}
		}
	}
	return false
}

// FindBuildFenceIntentForTest is an exported wrapper for integration tests.
func FindBuildFenceIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findBuildFenceIntent(char, pos, items, order, log, gameMap)
}

// FindBuildHutIntentForTest is an exported wrapper for integration tests.
func FindBuildHutIntentForTest(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	return findBuildHutIntent(char, pos, items, order, log, gameMap)
}

// SelectConstructionMaterialForTest is an exported wrapper for tests.
func SelectConstructionMaterialForTest(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, activityID string) string {
	return selectConstructionMaterial(char, pos, items, gameMap, activityID)
}

// CountFullBundlesAtPositionForTest is an exported wrapper for tests.
func CountFullBundlesAtPositionForTest(items []*entity.Item, itemType string, pos types.Position) int {
	return countFullBundlesAtPosition(items, itemType, pos)
}

// SuppliesMetForHutTileForTest is an exported wrapper for tests.
func SuppliesMetForHutTileForTest(items []*entity.Item, material string, pos types.Position) bool {
	return suppliesMetForHutTile(items, material, pos)
}

// countFullBundlesAtPosition counts items at pos where BundleCount >= MaxBundleSize for that type.
func countFullBundlesAtPosition(items []*entity.Item, itemType string, pos types.Position) int {
	maxSize := config.MaxBundleSize[itemType]
	if maxSize == 0 {
		return 0
	}
	count := 0
	for _, item := range items {
		if item.ItemType == itemType && item.Pos() == pos && item.BundleCount >= maxSize {
			count++
		}
	}
	return count
}

// suppliesMetForHutTile checks if a tile has enough supplies to build a hut wall/door.
// For bundleable materials: 2 full bundles (recipe count 12 / MaxBundleSize 6).
// For non-bundleable (brick): 12 items (recipe count directly).
func suppliesMetForHutTile(items []*entity.Item, material string, pos types.Position) bool {
	_, isBundleMaterial := config.MaxBundleSize[material]
	if isBundleMaterial {
		return countFullBundlesAtPosition(items, material, pos) >= 2
	}
	return countItemsAtPosition(items, material, pos) >= 12
}

// findBuildHutIntent finds the next intent for a character building a hut.
// Follows the same pattern as findBuildFenceIntent but with supply-drop for all materials.
func findBuildHutIntent(char *entity.Character, pos types.Position, items []*entity.Item, order *entity.Order, log *ActionLog, gameMap *game.Map) *entity.Intent {
	// Step 1: Collect unbuilt hut-marked tiles not occupied by another character (DD-28, DD-49)
	var candidates []types.Position
	for _, mpos := range gameMap.MarkedForConstructionPositions() {
		mark, ok := gameMap.GetConstructionMark(mpos)
		if !ok || mark.ConstructKind != "hut" {
			continue // Only hut marks (DD-49)
		}
		if gameMap.ConstructAt(mpos) != nil {
			continue // Already built
		}
		if occ := gameMap.CharacterAt(mpos); occ != nil && occ != char {
			continue // Occupied by another character — skip per DD-28
		}
		candidates = append(candidates, mpos)
	}
	if len(candidates) == 0 {
		return nil // No work → triggers completion check
	}
	sort.Slice(candidates, func(i, j int) bool {
		return pos.DistanceTo(candidates[i]) < pos.DistanceTo(candidates[j])
	})
	nearest := candidates[0]

	// Step 2: Determine material for the nearest tile's line
	mark, _ := gameMap.GetConstructionMark(nearest)
	material := mark.Material
	if material == "" {
		material = selectConstructionMaterial(char, pos, items, gameMap, "buildHut")
		if material == "" {
			return nil // No material available → triggers abandonment
		}
		gameMap.SetLineMaterial(mark.LineID, material)
	}

	// Step 3: Drop non-material inventory items (procurement drop pattern)
	var toDrop []*entity.Item
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType != material {
			toDrop = append(toDrop, inv)
		}
	}
	for _, item := range toDrop {
		DropItem(char, item, gameMap, log)
	}

	// Step 4: Check if nearest tile has enough supplies — branch to build or supply phase
	_, isBundleMaterial := config.MaxBundleSize[material]

	if isBundleMaterial {
		return findBundleHutIntent(char, pos, items, nearest, candidates, material, gameMap, log)
	}
	return findBrickHutIntent(char, pos, items, nearest, candidates, material, gameMap, log)
}

// findBundleHutIntent handles hut building with bundle materials (grass/sticks).
// Supply-drop pattern: deliver 2 full bundles to the tile, then build from adjacent.
func findBundleHutIntent(char *entity.Character, pos types.Position, items []*entity.Item, nearest types.Position, candidates []types.Position, material string, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Phase 1: Check if nearest tile has enough supplies to build
	if suppliesMetForHutTile(items, material, nearest) {
		return findHutBuildIntent(char, pos, candidates, material, gameMap)
	}

	// Count full bundles in inventory
	maxSize := config.MaxBundleSize[material]
	fullBundleCount := 0
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material && inv.BundleCount >= maxSize {
			fullBundleCount++
		}
	}

	// Phase 2: Delivery — character has 2 full bundles (both slots filled) → deliver (DD-39)
	if fullBundleCount >= 2 {
		return createHutDeliveryIntent(char, pos, nearest, gameMap)
	}

	// Phase 3: Procurement — try to fill remaining inventory slots
	// Only search for full/partial bundles when character has no bundle yet (DD-30 overflow guard)
	hasPartialBundle := false
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material && inv.BundleCount > 0 {
			hasPartialBundle = true
			break
		}
	}

	var target *entity.Item
	if !hasPartialBundle {
		target = findNearestMaterialNotAtSite(pos, items, material, gameMap, true)
	}
	if target == nil {
		// Fallback: individual items not at construction sites
		target = findNearestMaterialNotAtSite(pos, items, material, gameMap, false)
	}
	if target != nil {
		return createItemPickupIntent(char, pos, target, gameMap, log)
	}

	// Phase 4: No more materials available — deliver what we have (partial load)
	if fullBundleCount >= 1 {
		return createHutDeliveryIntent(char, pos, nearest, gameMap)
	}

	return nil // No materials at all → triggers abandonment
}

// findBrickHutIntent handles hut building with brick (non-bundle) materials.
// Supply-drop pattern: deliver bricks to the tile (12 per tile), then build from adjacent.
// createHutDeliveryIntent creates a delivery intent to move to the build tile and drop materials.
func createHutDeliveryIntent(char *entity.Character, pos types.Position, buildPos types.Position, gameMap *game.Map) *entity.Intent {
	nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, buildPos.X, buildPos.Y, gameMap, char.UsingBFS)
	if usedBFS {
		char.UsingBFS = true
	}
	newActivity := "Delivering materials"
	if pos == buildPos {
		newActivity = "Dropping materials"
	}
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target:         types.Position{X: nx, Y: ny},
		Dest:           buildPos,
		Action:         entity.ActionBuildHut,
		TargetBuildPos: &buildPos,
	}
}

func findBrickHutIntent(char *entity.Character, pos types.Position, items []*entity.Item, nearest types.Position, candidates []types.Position, material string, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Phase 1: Check if nearest tile has enough supplies to build
	if suppliesMetForHutTile(items, material, nearest) {
		return findHutBuildIntent(char, pos, candidates, material, gameMap)
	}

	// Count bricks in inventory
	brickCount := 0
	for _, inv := range char.Inventory {
		if inv != nil && inv.ItemType == material {
			brickCount++
		}
	}

	// Phase 2: Delivery — both inventory slots full → deliver
	if brickCount >= 2 {
		return createHutDeliveryIntent(char, pos, nearest, gameMap)
	}

	// Phase 3: Procurement — try to fill remaining inventory slots
	target := findNearestMaterialNotAtSite(pos, items, material, gameMap, false)
	if target != nil {
		return createItemPickupIntent(char, pos, target, gameMap, log)
	}

	// Phase 4: No more materials available — deliver what we have (partial load)
	if brickCount >= 1 {
		return createHutDeliveryIntent(char, pos, nearest, gameMap)
	}

	return nil // No bricks at all → triggers abandonment
}

// findHutBuildIntent returns a build intent for hut tiles when supplies are met.
// Finds the nearest candidate with enough supplies and an adjacent standing tile.
func findHutBuildIntent(char *entity.Character, pos types.Position, candidates []types.Position, material string, gameMap *game.Map) *entity.Intent {
	items := gameMap.Items()
	for _, candidate := range candidates {
		cm, _ := gameMap.GetConstructionMark(candidate)
		if cm.Material != "" && cm.Material != material {
			continue // Different line with different material
		}
		if !suppliesMetForHutTile(items, material, candidate) {
			continue // Not enough supplies at this tile yet
		}
		adjPos := findAdjacentStandingTile(candidate, gameMap)
		if adjPos == nil {
			continue // All adjacent tiles blocked — try next candidate
		}
		buildPos := candidate
		nx, ny, usedBFS := nextStepBFSCore(pos.X, pos.Y, adjPos.X, adjPos.Y, gameMap, char.UsingBFS)
		if usedBFS {
			char.UsingBFS = true
		}
		newActivity := "Building hut"
		if pos != *adjPos {
			newActivity = "Moving to build hut"
		}
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			Target:         types.Position{X: nx, Y: ny},
			Dest:           *adjPos,
			Action:         entity.ActionBuildHut,
			TargetBuildPos: &buildPos,
		}
	}
	return nil // No viable candidate
}
