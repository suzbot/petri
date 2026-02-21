package system

import (
	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// findForageIntent creates an intent to forage (pick up) an edible item.
// Called by selectIdleActivity when foraging is selected.
// Uses unified scoring of growing items vs vessels, with vessel value scaled by hunger.
// Lower hunger = more willing to invest in vessel; higher hunger = grab immediate food.
// If carrying a vessel with contents, only targets matching variety.
func findForageIntent(char *entity.Character, pos types.Position, items []*entity.Item, log *ActionLog, registry *game.VarietyRegistry, gameMap *game.Map) *entity.Intent {
	// If already carrying a vessel, just find a target item
	vessel := char.GetCarriedVessel()
	if vessel != nil {
		return findForageItemIntent(char, pos, items, vessel, log, gameMap)
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
		return createPickupIntent(char, pos, bestVessel, "vessel", log, gameMap)
	}

	if bestItem != nil {
		return createPickupIntent(char, pos, bestItem, bestItem.ItemType, log, gameMap)
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
		// Only consider edible, growing items (not sprouts — they're still maturing)
		if !item.IsEdible() || (item.Plant != nil && (!item.Plant.IsGrowing || item.Plant.IsSprout)) {
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
		if item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
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
func findForageItemIntent(char *entity.Character, pos types.Position, items []*entity.Item, vessel *entity.Item, log *ActionLog, gameMap *game.Map) *entity.Intent {
	target, _ := scoreForageItems(char, pos, items, vessel)
	if target == nil {
		return nil
	}
	return createPickupIntent(char, pos, target, target.ItemType, log, gameMap)
}

// createPickupIntent creates an intent to move to and pick up an item.
func createPickupIntent(char *entity.Character, pos types.Position, target *entity.Item, itemType string, log *ActionLog, gameMap *game.Map) *entity.Intent {
	tpos := target.Pos()
	tx, ty := tpos.X, tpos.Y

	// Vessels use "Picking up" text; food items use "Foraging"
	isVessel := target.Container != nil
	atVerb := "Foraging "
	movingVerb := "Moving to forage "
	logVerb := "Foraging for "
	if isVessel {
		atVerb = "Picking up "
		movingVerb = "Moving to pick up "
		logVerb = "Picking up "
	}

	if pos.X == tx && pos.Y == ty {
		// Already at target
		newActivity := atVerb + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", logVerb+itemType)
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
	newActivity := movingVerb + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "activity", logVerb+itemType)
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
func FindNextVesselTarget(char *entity.Character, cx, cy int, items []*entity.Item, registry *game.VarietyRegistry, gameMap *game.Map) *entity.Intent {
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
		// Must be growing (not dropped) and not a sprout
		if item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
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
			Target:     types.Position{X: cx, Y: cy},
			Dest:       types.Position{X: cx, Y: cy}, // Already at destination
			Action:     entity.ActionPickup,
			TargetItem: nearest,
		}
	}

	// Move toward target
	nx, ny := NextStepBFS(cx, cy, tx, ty, gameMap)
	char.CurrentActivity = "Moving to forage " + nearest.Description()
	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       types.Position{X: tx, Y: ty}, // Destination is the item's position
		Action:     entity.ActionPickup,
		TargetItem: nearest,
	}
}
