package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// Prerequisite Helpers
// =============================================================================

// EnsureHasVesselFor returns an intent to get a vessel that can accept the target item,
// or nil if the character already has a compatible vessel.
// If dropConflict is true and the character has an incompatible/full vessel, drops it.
// category is used for logging (e.g., "order", "activity").
func EnsureHasVesselFor(char *entity.Character, target *entity.Item, items []*entity.Item, gameMap *game.Map, log *ActionLog, dropConflict bool, category string) *entity.Intent {
	registry := gameMap.Varieties()

	// Check if already carrying a compatible vessel
	carriedVessel := char.GetCarriedVessel()
	if carriedVessel != nil {
		if CanVesselAccept(carriedVessel, target, registry) {
			// Already have compatible vessel
			return nil
		}

		// Have incompatible/full vessel
		if !dropConflict {
			// Can't drop - can't get vessel
			return nil
		}

		// Drop the incompatible vessel
		DropItem(char, carriedVessel, gameMap, log)
	}

	// Need to find a vessel - check if we have inventory space
	if !char.HasInventorySpace() {
		return nil
	}

	// Find available vessel on map
	pos := char.Pos()
	availableVessel := FindAvailableVessel(pos.X, pos.Y, items, target, registry)
	if availableVessel == nil {
		return nil
	}

	// Create intent to pick up the vessel
	return createVesselPickupIntent(char, pos, availableVessel, log, category, gameMap)
}

// createVesselPickupIntent creates an intent to move to and pick up a vessel.
func createVesselPickupIntent(char *entity.Character, pos types.Position, vessel *entity.Item, log *ActionLog, category string, gameMap *game.Map) *entity.Intent {
	vpos := vessel.Pos()
	vx, vy := vpos.X, vpos.Y

	if pos.X == vx && pos.Y == vy {
		// Already at vessel
		newActivity := "Picking up vessel"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, category, "Picking up vessel")
			}
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos,
			Action:     entity.ActionPickup,
			TargetItem: vessel,
		}
	}

	// Move toward vessel
	nx, ny := NextStepBFS(pos.X, pos.Y, vx, vy, gameMap)
	newActivity := "Moving to pick up vessel"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
	}
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: vx, Y: vy},
		Action:     entity.ActionPickup,
		TargetItem: vessel,
	}
}

// =============================================================================
// EnsureHasItem
// =============================================================================

// EnsureHasItem returns an intent to acquire a specific item type,
// or nil if the character already has one (or none exists on the map).
// Caller distinguishes "ready" vs "impossible" by checking inventory after nil return.
func EnsureHasItem(char *entity.Character, itemType string, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Check if already carrying the item
	if char.FindInInventory(func(i *entity.Item) bool { return i.ItemType == itemType }) != nil {
		return nil
	}

	// Need to find one — make inventory space if needed
	if !char.HasInventorySpace() {
		dropped := false
		for _, inv := range char.Inventory {
			if inv == nil {
				continue
			}
			// Don't drop the target item type (shouldn't happen, but safe)
			if inv.ItemType == itemType {
				continue
			}
			DropItem(char, inv, gameMap, log)
			dropped = true
			break
		}
		if !dropped {
			return nil
		}
	}

	// Find nearest item of type on map
	target := findNearestItemByType(char.X, char.Y, items, itemType, false)
	if target == nil {
		return nil // Not available on map
	}

	return createItemPickupIntent(char, char.Pos(), target, gameMap, log)
}

// =============================================================================
// EnsureHasRecipeInputs
// =============================================================================

// EnsureHasRecipeInputs returns an intent to acquire a missing recipe input,
// or nil if the character already has all inputs (or inputs are unavailable on map).
// Caller distinguishes "ready" vs "impossible" by checking HasAccessibleItem after nil return.
//
// Component seeking uses nearest-distance for now. See triggered-enhancements.md
// for preference-weighted component procurement — when triggered, this should use
// the same scoring patterns as foraging.go's scoreForageItems.
func EnsureHasRecipeInputs(char *entity.Character, recipe *entity.Recipe, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	if recipe == nil {
		return nil
	}

	// Check which inputs are missing
	for _, input := range recipe.Inputs {
		if char.HasAccessibleItem(input.ItemType) {
			continue // Have this input
		}

		// Missing this input — need to make room and go get it

		// Drop non-recipe loose items to make inventory space
		if !char.HasInventorySpace() {
			dropped := false
			for _, inv := range char.Inventory {
				if inv == nil {
					continue
				}
				// Don't drop containers that hold a recipe input
				if inv.Container != nil && containerHasRecipeInput(inv, recipe) {
					continue
				}
				// Don't drop items that ARE recipe inputs
				if isRecipeInput(inv.ItemType, recipe) {
					continue
				}
				// Drop this non-recipe item
				DropItem(char, inv, gameMap, log)
				dropped = true
				break
			}
			if !dropped {
				// Couldn't make space
				return nil
			}
		}

		// Find nearest matching item on map
		target := findNearestItemByType(char.X, char.Y, items, input.ItemType, false)
		if target == nil {
			return nil // Input not available on map
		}

		// Create intent to pick up the item
		return createItemPickupIntent(char, char.Pos(), target, gameMap, log)
	}

	// All inputs accessible
	return nil
}

// containerHasRecipeInput checks if a container holds any item type needed by the recipe.
func containerHasRecipeInput(container *entity.Item, recipe *entity.Recipe) bool {
	if container.Container == nil {
		return false
	}
	for _, stack := range container.Container.Contents {
		if stack.Variety == nil || stack.Count <= 0 {
			continue
		}
		if isRecipeInput(stack.Variety.ItemType, recipe) {
			return true
		}
	}
	return false
}

// isRecipeInput returns true if the given item type is an input for the recipe.
func isRecipeInput(itemType string, recipe *entity.Recipe) bool {
	for _, input := range recipe.Inputs {
		if input.ItemType == itemType {
			return true
		}
	}
	return false
}

// createItemPickupIntent creates an intent to move to and pick up an item.
func createItemPickupIntent(char *entity.Character, pos types.Position, target *entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	if pos.X == tx && pos.Y == ty {
		newActivity := "Picking up " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "order", "Picking up "+target.Description())
			}
		}
		return &entity.Intent{
			Target:     pos,
			Dest:       pos,
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	// Move toward target
	nx, ny := NextStepBFS(pos.X, pos.Y, tx, ty, gameMap)
	newActivity := "Moving to pick up " + target.Description()
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

// =============================================================================
// Map Search Utilities
// =============================================================================

// findNearestItemByType finds the closest item of a specific type.
// If growingOnly is true, only considers items with Plant.IsGrowing == true (for harvest).
// If growingOnly is false, considers any item of that type on the map (for recipe components, etc).
func findNearestItemByType(cx, cy int, items []*entity.Item, itemType string, growingOnly bool) *entity.Item {
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
		if growingOnly {
			// Only consider growing items
			if item.Plant == nil || !item.Plant.IsGrowing {
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

// =============================================================================
// Vessel Helpers
// =============================================================================

// AddToVessel attempts to add an item to a vessel's contents.
// Returns true if the item was added, false if:
// - vessel has no container, or registry is nil, or variety not found
// - vessel stack is at capacity (based on item's StackSize)
// - vessel already has a different variety (variety lock)
func AddToVessel(vessel, item *entity.Item, registry *game.VarietyRegistry) bool {
	if vessel == nil || vessel.Container == nil || registry == nil {
		return false
	}

	// Look up the item's variety in the registry
	variety := registry.GetByAttributes(item.ItemType, item.Color, item.Pattern, item.Texture)
	if variety == nil {
		return false
	}

	stackSize := config.GetStackSize(item.ItemType)

	// If vessel is empty, create a new stack
	if len(vessel.Container.Contents) == 0 {
		vessel.Container.Contents = []entity.Stack{{
			Variety: variety,
			Count:   1,
		}}
		return true
	}

	// Vessel has contents - check if variety matches
	stack := &vessel.Container.Contents[0]
	if stack.Variety.ID != variety.ID {
		return false // Variety mismatch
	}

	// Check if stack is at capacity
	if stack.Count >= stackSize {
		return false
	}

	// Add to stack
	stack.Count++
	return true
}

// CanVesselAccept returns true if a vessel can accept a specific item.
// True if: vessel is empty, OR vessel has same variety and space remaining.
func CanVesselAccept(vessel, item *entity.Item, registry *game.VarietyRegistry) bool {
	if vessel == nil || vessel.Container == nil || registry == nil {
		return false
	}

	// Empty vessel can accept anything
	if len(vessel.Container.Contents) == 0 {
		return true
	}

	// Check if variety matches
	stack := vessel.Container.Contents[0]
	if stack.Variety == nil {
		return false
	}

	// Must match variety
	if item.ItemType != stack.Variety.ItemType ||
		item.Color != stack.Variety.Color ||
		item.Pattern != stack.Variety.Pattern ||
		item.Texture != stack.Variety.Texture {
		return false
	}

	// Check if stack has space
	stackSize := config.GetStackSize(stack.Variety.ItemType)
	return stack.Count < stackSize
}

// IsVesselFull returns true if the vessel cannot hold more items.
// A vessel is full when its stack is at capacity for that item type.
// Returns false if vessel has no container or is empty (can accept new items).
func IsVesselFull(vessel *entity.Item, registry *game.VarietyRegistry) bool {
	if vessel == nil || vessel.Container == nil {
		return false
	}

	// Empty vessel is not full
	if len(vessel.Container.Contents) == 0 {
		return false
	}

	stack := vessel.Container.Contents[0]
	if stack.Variety == nil {
		return false
	}

	stackSize := config.GetStackSize(stack.Variety.ItemType)
	return stack.Count >= stackSize
}

// FindAvailableVessel finds the nearest vessel on the map that can accept a target item.
// Returns nil if no suitable vessel is found.
// A vessel is suitable if it's empty OR has matching variety with space.
func FindAvailableVessel(cx, cy int, items []*entity.Item, targetItem *entity.Item, registry *game.VarietyRegistry) *entity.Item {
	if targetItem == nil || registry == nil {
		return nil
	}

	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range items {
		// Must be a vessel (has container)
		if item.Container == nil {
			continue
		}

		// Check if vessel can accept the target item
		if !CanVesselAccept(item, targetItem, registry) {
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

// =============================================================================
// Pickup Capacity
// =============================================================================

// CanPickUpMore returns true if the character can pick up more items.
// True if: has empty inventory slot, OR carrying a vessel with space.
func CanPickUpMore(char *entity.Character, registry *game.VarietyRegistry) bool {
	if char.HasInventorySpace() {
		return true
	}
	// Inventory is full, but check if any vessel has space
	vessel := char.GetCarriedVessel()
	if vessel != nil {
		return !IsVesselFull(vessel, registry)
	}
	return false
}

// =============================================================================
// Drop Actions
// =============================================================================

// Drop handles a character dropping the first item from inventory
func Drop(char *entity.Character, gameMap *game.Map, log *ActionLog) {
	if len(char.Inventory) == 0 {
		return // Nothing to drop
	}

	// Drop first item in inventory
	item := char.Inventory[0]
	itemName := item.Description()

	// Place item on map at character's position
	item.X = char.X
	item.Y = char.Y
	gameMap.AddItem(item)

	// Remove from inventory
	char.RemoveFromInventory(item)

	// Log drop
	if log != nil {
		log.Add(char.ID, char.Name, "activity",
			fmt.Sprintf("Dropped %s", itemName))
	}
}

// DropItem handles a character dropping a specific item from inventory
func DropItem(char *entity.Character, item *entity.Item, gameMap *game.Map, log *ActionLog) {
	if !char.RemoveFromInventory(item) {
		return // Item not in inventory
	}

	itemName := item.Description()

	// Place item on map at character's position
	item.X = char.X
	item.Y = char.Y
	gameMap.AddItem(item)

	// Log drop
	if log != nil {
		log.Add(char.ID, char.Name, "activity",
			fmt.Sprintf("Dropped %s", itemName))
	}
}

// =============================================================================
// Pickup Actions
// =============================================================================

// PickupResult indicates what happened during a pickup action
type PickupResult int

const (
	// PickupToInventory - item was placed in empty inventory slot
	PickupToInventory PickupResult = iota
	// PickupToVessel - item was added to a carried vessel
	PickupToVessel
	// PickupFailed - could not pick up (vessel variety mismatch or full)
	PickupFailed
)

// Pickup handles a character picking up an item (foraging/harvesting).
// If carrying a vessel with space, adds item to vessel and returns PickupToVessel.
// Otherwise places item in inventory and returns PickupToInventory.
// When adding to vessel, intent is NOT cleared (caller should continue foraging).
// When adding to inventory, intent IS cleared and idle cooldown set.
func Pickup(char *entity.Character, item *entity.Item, gameMap *game.Map, log *ActionLog, registry *game.VarietyRegistry) PickupResult {
	itemName := item.Description()

	// Try to add to any vessel that can accept the item
	for _, vessel := range char.Inventory {
		if vessel == nil || vessel.Container == nil {
			continue
		}
		if AddToVessel(vessel, item, registry) {
			// Successfully added to vessel
			gameMap.RemoveItem(item)

			// Mark as no longer growing
			if item.Plant != nil {
				item.Plant.IsGrowing = false
				item.Plant.SpawnTimer = 0
			}
			item.DeathTimer = 0

			// Berries and mushrooms become plantable when picked
			if item.ItemType == "berry" || item.ItemType == "mushroom" {
				item.Plantable = true
			}

			// Log the addition
			if log != nil {
				count := vessel.Container.Contents[0].Count
				log.Add(char.ID, char.Name, "activity",
					fmt.Sprintf("Added %s to vessel (%d)", itemName, count))
			}

			// Try to discover know-how
			TryDiscoverKnowHow(char, entity.ActionPickup, item, log, GetDiscoveryChance(char))

			// DON'T clear intent - caller will decide if foraging continues
			return PickupToVessel
		}
	}

	// No vessel could accept - check if there's inventory space for loose pickup
	if !char.HasInventorySpace() {
		// No space - return failure so caller can decide
		return PickupFailed
	}

	// Check if inventory has space
	if !char.HasInventorySpace() {
		return PickupFailed
	}

	// Standard pickup to inventory
	gameMap.RemoveItem(item)

	// Mark as no longer growing (won't respawn if dropped)
	if item.Plant != nil {
		item.Plant.IsGrowing = false
		item.Plant.SpawnTimer = 0
	}
	// Clear death timer - carried items don't decay
	item.DeathTimer = 0

	// Berries and mushrooms become plantable when picked
	if item.ItemType == "berry" || item.ItemType == "mushroom" {
		item.Plantable = true
	}

	// Add to inventory
	char.AddToInventory(item)

	// Update activity
	char.CurrentActivity = "Idle"

	// Log pickup
	if log != nil {
		log.Add(char.ID, char.Name, "activity",
			fmt.Sprintf("Picked up %s", itemName))
	}

	// Try to discover know-how from foraging
	TryDiscoverKnowHow(char, entity.ActionPickup, item, log, GetDiscoveryChance(char))

	// Clear intent and set idle cooldown
	char.Intent = nil
	char.IdleCooldown = config.IdleCooldown

	return PickupToInventory
}
