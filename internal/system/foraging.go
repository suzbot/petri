package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
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
// True if: not carrying anything, OR carrying a vessel with space.
func CanPickUpMore(char *entity.Character, registry *game.VarietyRegistry) bool {
	if char.Carrying == nil {
		return true
	}
	if char.Carrying.Container != nil {
		return !IsVesselFull(char.Carrying, registry)
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

		ix, iy := item.Position()
		dist := abs(cx-ix) + abs(cy-iy)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	return nearest
}

// Drop handles a character dropping an item from inventory
func Drop(char *entity.Character, gameMap *game.Map, log *ActionLog) {
	if char.Carrying == nil {
		return // Nothing to drop
	}

	item := char.Carrying
	itemName := item.Description()

	// Place item on map at character's position
	item.X = char.X
	item.Y = char.Y
	gameMap.AddItem(item)

	// Clear from inventory
	char.Carrying = nil

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

	// Check if carrying a vessel
	if char.Carrying != nil && char.Carrying.Container != nil {
		if AddToVessel(char.Carrying, item, registry) {
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
				count := char.Carrying.Container.Contents[0].Count
				log.Add(char.ID, char.Name, "activity",
					fmt.Sprintf("Added %s to vessel (%d)", itemName, count))
			}

			// Try to discover know-how
			TryDiscoverKnowHow(char, entity.ActionPickup, item, log, GetDiscoveryChance(char))

			// DON'T clear intent - caller will decide if foraging continues
			return PickupToVessel
		}
		// Vessel full or variety mismatch - cannot pick up without dropping vessel
		// Return failure so caller can decide whether to drop vessel or skip item
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
	char.Carrying = item

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
// Uses preference/distance scoring similar to eating but without hunger-based filtering.
// If not carrying a vessel, will look for one first before foraging.
// If carrying a vessel with contents, only targets matching variety.
func findForageIntent(char *entity.Character, cx, cy int, items []*entity.Item, log *ActionLog, registry *game.VarietyRegistry) *entity.Intent {
	// Get the vessel if carrying one (used for variety filtering)
	var vessel *entity.Item
	if char.Carrying != nil && char.Carrying.Container != nil {
		vessel = char.Carrying
	}

	// Find best edible item using preference/distance gradient
	// If carrying vessel with contents, only considers matching variety
	target := findForageTarget(char, cx, cy, items, vessel)
	if target == nil {
		return nil
	}

	// If not carrying anything, look for available vessel first
	if char.Carrying == nil {
		availableVessel := FindAvailableVessel(cx, cy, items, target, registry)
		if availableVessel != nil {
			// Go pick up vessel first
			vx, vy := availableVessel.Position()
			if cx == vx && cy == vy {
				newActivity := "Picking up vessel"
				if char.CurrentActivity != newActivity {
					char.CurrentActivity = newActivity
					if log != nil {
						log.Add(char.ID, char.Name, "activity", "Picking up vessel for foraging")
					}
				}
				return &entity.Intent{
					TargetX:    cx,
					TargetY:    cy,
					Action:     entity.ActionPickup,
					TargetItem: availableVessel,
				}
			}
			// Move toward vessel
			nx, ny := NextStep(cx, cy, vx, vy)
			newActivity := "Moving to pick up vessel"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
				if log != nil {
					log.Add(char.ID, char.Name, "activity", "Getting vessel for foraging")
				}
			}
			return &entity.Intent{
				TargetX:    nx,
				TargetY:    ny,
				Action:     entity.ActionPickup,
				TargetItem: availableVessel,
			}
		}
		// No vessel available - proceed to pick up target directly (single item)
	}

	tx, ty := target.Position()

	// Check if already at target
	if cx == tx && cy == ty {
		// Start foraging immediately
		newActivity := "Foraging " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Foraging for "+target.ItemType)
			}
		}
		return &entity.Intent{
			TargetX:    cx,
			TargetY:    cy,
			Action:     entity.ActionPickup,
			TargetItem: target,
		}
	}

	// Move toward target - use ActionPickup to distinguish from looking
	nx, ny := NextStep(cx, cy, tx, ty)

	newActivity := "Moving to forage " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "activity", "Foraging for "+target.ItemType)
		}
	}

	return &entity.Intent{
		TargetX:    nx,
		TargetY:    ny,
		Action:     entity.ActionPickup,
		TargetItem: target,
	}
}

// findForageTarget finds the best edible item for foraging using preference/distance scoring.
// Uses moderate preference weight - foraging is idle activity, not urgent need.
// If vessel is provided and has contents, only returns items matching the vessel's variety.
func findForageTarget(char *entity.Character, cx, cy int, items []*entity.Item, vessel *entity.Item) *entity.Item {
	if len(items) == 0 {
		return nil
	}

	// Get variety constraint from vessel if it has contents
	var requiredVariety *entity.ItemVariety
	if vessel != nil && vessel.Container != nil && len(vessel.Container.Contents) > 0 {
		requiredVariety = vessel.Container.Contents[0].Variety
	}

	var bestItem *entity.Item
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float
	bestDist := int(^uint(0) >> 1)              // Max int for distance tiebreaker

	for _, item := range items {
		// Only consider edible, growing items for foraging
		if !item.IsEdible() {
			continue
		}
		if item.Plant == nil || !item.Plant.IsGrowing {
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
		ix, iy := item.Position()
		dist := abs(cx-ix) + abs(cy-iy)

		// Calculate gradient score (same weights as moderate hunger eating)
		score := float64(netPref)*config.FoodSeekPrefWeightModerate - float64(dist)*config.FoodSeekDistWeight

		// Update best if better score, or same score but closer
		if score > bestScore || (score == bestScore && dist < bestDist) {
			bestItem = item
			bestScore = score
			bestDist = dist
		}
	}

	return bestItem
}

// FindNextVesselTarget finds the next item to pick up when filling a vessel.
// Only considers growing items matching the variety already in the vessel.
// Returns nil if vessel is empty, full, or no matching items exist.
func FindNextVesselTarget(char *entity.Character, cx, cy int, items []*entity.Item, registry *game.VarietyRegistry) *entity.Intent {
	if char.Carrying == nil || char.Carrying.Container == nil {
		return nil
	}
	if len(char.Carrying.Container.Contents) == 0 {
		return nil // Empty vessel - shouldn't happen mid-forage
	}
	if IsVesselFull(char.Carrying, registry) {
		return nil
	}

	targetVariety := char.Carrying.Container.Contents[0].Variety
	if targetVariety == nil {
		return nil
	}

	// Find nearest item matching the variety
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

		ix, iy := item.Position()
		dist := abs(cx-ix) + abs(cy-iy)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	if nearest == nil {
		return nil
	}

	tx, ty := nearest.Position()
	if cx == tx && cy == ty {
		// Already at target
		char.CurrentActivity = "Foraging " + nearest.Description()
		return &entity.Intent{
			TargetX:    cx,
			TargetY:    cy,
			Action:     entity.ActionPickup,
			TargetItem: nearest,
		}
	}

	// Move toward target
	nx, ny := NextStep(cx, cy, tx, ty)
	char.CurrentActivity = "Moving to forage " + nearest.Description()
	return &entity.Intent{
		TargetX:    nx,
		TargetY:    ny,
		Action:     entity.ActionPickup,
		TargetItem: nearest,
	}
}

