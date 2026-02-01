package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

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

	// Check if carrying a vessel with space
	vessel := char.GetCarriedVessel()
	if vessel != nil {
		if AddToVessel(vessel, item, registry) {
			// Successfully added to vessel
			gameMap.RemoveItem(item)

			// Mark as no longer growing
			if item.Plant != nil {
				item.Plant.IsGrowing = false
				item.Plant.SpawnTimer = 0
			}
			item.DeathTimer = 0

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
		// Vessel full or variety mismatch - check if there's inventory space for loose pickup
		if !char.HasInventorySpace() {
			// No space - return failure so caller can decide
			return PickupFailed
		}
		// Fall through to standard pickup
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

// findForageIntent creates an intent to forage (pick up) an edible item.
// Called by selectIdleActivity when foraging is selected.
// Uses unified scoring of growing items vs vessels, with vessel value scaled by hunger.
// Lower hunger = more willing to invest in vessel; higher hunger = grab immediate food.
// If carrying a vessel with contents, only targets matching variety.
func findForageIntent(char *entity.Character, pos types.Position, items []*entity.Item, log *ActionLog, registry *game.VarietyRegistry) *entity.Intent {
	// If already carrying a vessel, just find a target item
	vessel := char.GetCarriedVessel()
	if vessel != nil {
		return findForageItemIntent(char, pos, items, vessel, log)
	}

	// No vessel - use unified scoring to decide between growing items and vessels
	// Vessel bonus scales with hunger: low hunger = willing to plan ahead
	vesselBonus := config.FoodSeekPrefWeightModerate * (1 - char.Hunger/100)

	// Find best growing item
	bestItem, bestItemScore := scoreForageItems(char, pos, items, nil)

	// Find best vessel (if inventory has space)
	var bestVessel *entity.Item
	bestVesselScore := float64(int(^uint(0)>>1)) * -1 // Negative max float

	if char.HasInventorySpace() {
		bestVessel, bestVesselScore = scoreForageVessels(char, pos, items, vesselBonus, registry)
	}

	// Nothing to forage
	if bestItem == nil && bestVessel == nil {
		return nil
	}

	// Pick the better option (vessel wins ties - investment pays off)
	if bestVessel != nil && bestVesselScore >= bestItemScore {
		return createPickupIntent(char, pos, bestVessel, "vessel", log)
	}

	if bestItem != nil {
		return createPickupIntent(char, pos, bestItem, bestItem.ItemType, log)
	}

	return nil
}

// scoreForageItems scores all growing edible items and returns the best one.
// If vessel is provided and has contents, only considers matching variety.
func scoreForageItems(char *entity.Character, pos types.Position, items []*entity.Item, vessel *entity.Item) (*entity.Item, float64) {
	// Get variety constraint from vessel if it has contents
	var requiredVariety *entity.ItemVariety
	if vessel != nil && vessel.Container != nil && len(vessel.Container.Contents) > 0 {
		requiredVariety = vessel.Container.Contents[0].Variety
	}

	var bestItem *entity.Item
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float

	for _, item := range items {
		// Only consider edible, growing items
		if !item.IsEdible() || item.Plant == nil || !item.Plant.IsGrowing {
			continue
		}

		// If vessel has variety constraint, item must match
		if requiredVariety != nil {
			if item.ItemType != requiredVariety.ItemType ||
				item.Color != requiredVariety.Color ||
				item.Pattern != requiredVariety.Pattern ||
				item.Texture != requiredVariety.Texture {
				continue
			}
		}

		netPref := char.NetPreference(item)
		dist := pos.DistanceTo(item.Pos())
		score := float64(netPref)*config.FoodSeekPrefWeightModerate - float64(dist)*config.FoodSeekDistWeight

		if score > bestScore {
			bestItem = item
			bestScore = score
		}
	}

	return bestItem, bestScore
}

// scoreForageVessels scores all available vessels and returns the best one.
// Empty vessels get vesselBonus. Partial vessels get vesselBonus + content preference.
// Partial vessels are only scored if matching growing items exist.
func scoreForageVessels(char *entity.Character, pos types.Position, items []*entity.Item, vesselBonus float64, registry *game.VarietyRegistry) (*entity.Item, float64) {
	var bestVessel *entity.Item
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float

	for _, item := range items {
		// Must be a vessel (has container)
		if item.Container == nil {
			continue
		}

		// Must not be full
		if IsVesselFull(item, registry) {
			continue
		}

		dist := pos.DistanceTo(item.Pos())
		var score float64

		if len(item.Container.Contents) == 0 {
			// Empty vessel - just vesselBonus minus distance
			score = vesselBonus - float64(dist)*config.FoodSeekDistWeight
		} else {
			// Partial vessel - check if matching growing items exist
			contentVariety := item.Container.Contents[0].Variety
			if !hasMatchingGrowingItems(items, contentVariety) {
				continue // No point picking up vessel we can't fill
			}

			// Score with content preference
			contentPref := char.NetPreferenceForVariety(contentVariety)
			score = vesselBonus + float64(contentPref)*config.FoodSeekPrefWeightModerate - float64(dist)*config.FoodSeekDistWeight
		}

		if score > bestScore {
			bestVessel = item
			bestScore = score
		}
	}

	return bestVessel, bestScore
}

// hasMatchingGrowingItems checks if any growing items match the given variety.
func hasMatchingGrowingItems(items []*entity.Item, variety *entity.ItemVariety) bool {
	if variety == nil {
		return false
	}
	for _, item := range items {
		if item.Plant == nil || !item.Plant.IsGrowing {
			continue
		}
		if item.ItemType == variety.ItemType &&
			item.Color == variety.Color &&
			item.Pattern == variety.Pattern &&
			item.Texture == variety.Texture {
			return true
		}
	}
	return false
}

// findForageItemIntent creates intent to pick up a growing item when already carrying a vessel.
func findForageItemIntent(char *entity.Character, pos types.Position, items []*entity.Item, vessel *entity.Item, log *ActionLog) *entity.Intent {
	target, _ := scoreForageItems(char, pos, items, vessel)
	if target == nil {
		return nil
	}
	return createPickupIntent(char, pos, target, target.ItemType, log)
}

// createPickupIntent creates an intent to move to and pick up an item.
func createPickupIntent(char *entity.Character, pos types.Position, target *entity.Item, itemType string, log *ActionLog) *entity.Intent {
	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	if pos.X == tx && pos.Y == ty {
		// Already at target
		newActivity := "Foraging " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Foraging for "+itemType)
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
	nx, ny := NextStep(pos.X, pos.Y, tx, ty)
	newActivity := "Moving to forage " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "activity", "Foraging for "+itemType)
		}
	}

	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty},
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// FindNextVesselTarget finds the next item to pick up when filling a vessel.
// Only considers growing items matching the variety already in the vessel.
// Returns nil if vessel is empty, full, or no matching items exist.
func FindNextVesselTarget(char *entity.Character, cx, cy int, items []*entity.Item, registry *game.VarietyRegistry) *entity.Intent {
	vessel := char.GetCarriedVessel()
	if vessel == nil {
		return nil
	}
	if len(vessel.Container.Contents) == 0 {
		return nil // Empty vessel - shouldn't happen mid-forage
	}
	if IsVesselFull(vessel, registry) {
		return nil
	}

	targetVariety := vessel.Container.Contents[0].Variety
	if targetVariety == nil {
		return nil
	}

	// Find nearest item matching the variety
	pos := types.Position{X: cx, Y: cy}
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range items {
		// Must be growing (not dropped)
		if item.Plant == nil || !item.Plant.IsGrowing {
			continue
		}
		// Must match variety
		if item.ItemType != targetVariety.ItemType ||
			item.Color != targetVariety.Color ||
			item.Pattern != targetVariety.Pattern ||
			item.Texture != targetVariety.Texture {
			continue
		}

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
	if cx == tx && cy == ty {
		// Already at target
		char.CurrentActivity = "Foraging " + nearest.Description()
		return &entity.Intent{
			Target:          types.Position{X: cx, Y: cy},
			Dest:            types.Position{X: cx, Y: cy}, // Already at destination
			Action:     entity.ActionPickup,
			TargetItem: nearest,
		}
	}

	// Move toward target
	nx, ny := NextStep(cx, cy, tx, ty)
	char.CurrentActivity = "Moving to forage " + nearest.Description()
	return &entity.Intent{
		Target:          types.Position{X: nx, Y: ny},
		Dest:            types.Position{X: tx, Y: ty}, // Destination is the item's position
		Action:     entity.ActionPickup,
		TargetItem: nearest,
	}
}

