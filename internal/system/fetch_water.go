package system

import (
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// findFetchWaterIntent creates an intent to find an empty vessel and fill it with water.
// Called by selectIdleActivity when fetch water is selected.
// Non-destructive: never dumps existing vessel contents.
//
// Returns ActionFillVessel which manages the full lifecycle:
//   - Phase 1 (if needed): Move to ground vessel, pick it up
//   - Phase 2: Move to water, fill vessel
//
// Phase detection in continueIntent/applyIntent: if TargetItem is on the map,
// we're in phase 1 (acquiring vessel). If it's in inventory, we're in phase 2 (filling).
func findFetchWaterIntent(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Scan inventory: find empty vessel, check if already carrying water
	var emptyVessel *entity.Item
	for _, item := range char.Inventory {
		if item == nil || item.Container == nil {
			continue
		}
		if len(item.Container.Contents) == 0 {
			if emptyVessel == nil {
				emptyVessel = item
			}
		} else if item.Container.Contents[0].Variety != nil &&
			item.Container.Contents[0].Variety.ItemType == "liquid" {
			return nil // Already carrying water — no need to fetch more
		}
	}

	if emptyVessel == nil {
		// No empty vessel in inventory — look for one on the ground
		groundVessel := findEmptyGroundVessel(pos, items)
		if groundVessel == nil {
			return nil // No empty vessel available anywhere
		}

		// Need inventory space to pick it up
		if !char.HasInventorySpace() {
			return nil
		}

		// Phase 1: ActionFillVessel with Dest at vessel position
		// applyIntent will pick up the vessel, then transition to phase 2
		vpos := groundVessel.Pos()
		nx, ny := NextStepBFS(pos.X, pos.Y, vpos.X, vpos.Y, gameMap)
		newActivity := "Fetching water"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Picking up vessel for water")
			}
		}

		return &entity.Intent{
			Target:     types.Position{X: nx, Y: ny},
			Dest:       vpos,
			Action:     entity.ActionFillVessel,
			TargetItem: groundVessel,
		}
	}

	vessel := emptyVessel

	// Has empty vessel — find nearest water
	waterPos, found := gameMap.FindNearestWater(pos)
	if !found {
		return nil
	}

	// Find closest cardinal tile adjacent to water
	adjX, adjY := FindClosestCardinalTile(pos.X, pos.Y, waterPos.X, waterPos.Y, gameMap)
	if adjX == -1 {
		return nil // Water is blocked
	}

	dest := types.Position{X: adjX, Y: adjY}

	// Phase 2: ActionFillVessel with Dest at water-adjacent tile
	nx, ny := NextStepBFS(pos.X, pos.Y, adjX, adjY, gameMap)
	newActivity := "Fetching water"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "activity", "Heading to water to fill vessel")
		}
	}

	return &entity.Intent{
		Target:     types.Position{X: nx, Y: ny},
		Dest:       dest,
		Action:     entity.ActionFillVessel,
		TargetItem: vessel,
	}
}

// findEmptyGroundVessel finds the nearest empty vessel on the ground.
func findEmptyGroundVessel(pos types.Position, items []*entity.Item) *entity.Item {
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range items {
		if item.Container == nil {
			continue
		}
		// Must be empty (no contents)
		if len(item.Container.Contents) > 0 {
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
