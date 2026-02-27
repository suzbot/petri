package system

import (
	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// findNearestCrisisCharacter returns the nearest character with any Crisis stat
// (hunger or thirst). Distance is the primary criterion. Stat tie-break (thirst > hunger)
// applies only for equidistant characters. Skips the helper themselves, dead, and sleeping.
// Returns nil if no character is in crisis.
func findNearestCrisisCharacter(helper *entity.Character, characters []*entity.Character) *entity.Character {
	var nearest *entity.Character
	nearestDist := int(^uint(0) >> 1) // Max int
	nearestStatPriority := 0          // Higher = more urgent for tie-breaking

	hpos := helper.Pos()

	for _, other := range characters {
		if other == helper {
			continue
		}
		if other.IsDead || other.IsSleeping {
			continue
		}

		hasThirstCrisis := other.ThirstTier() == entity.TierCrisis
		hasHungerCrisis := other.HungerTier() == entity.TierCrisis
		if !hasThirstCrisis && !hasHungerCrisis {
			continue
		}

		// Stat priority for tie-breaking: thirst > hunger (matches intent system)
		statPriority := 0
		if hasHungerCrisis {
			statPriority = 1
		}
		if hasThirstCrisis {
			statPriority = 2 // Thirst wins tie-break
		}

		opos := other.Pos()
		dist := hpos.DistanceTo(opos)
		if dist < nearestDist || (dist == nearestDist && statPriority > nearestStatPriority) {
			nearestDist = dist
			nearest = other
			nearestStatPriority = statPriority
		}
	}

	return nearest
}

// findHelpWaterIntent creates an intent to find and deliver water to a needy character.
// Determines water delivery approach based on helper's current state:
//   - Helper carrying full water vessel → skip to delivery
//   - Helper carrying empty vessel → fill phase (find water)
//   - No vessel in inventory, has space → procure ground vessel
//   - No vessel available → return nil
//
// Returns nil if no vessel is available or helper can't carry anything.
func findHelpWaterIntent(helper *entity.Character, needer *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Scan inventory for vessels
	var emptyVessel *entity.Item
	var waterVessel *entity.Item
	for _, item := range helper.Inventory {
		if item == nil || item.Container == nil {
			continue
		}
		if len(item.Container.Contents) > 0 {
			if item.Container.Contents[0].Variety != nil &&
				item.Container.Contents[0].Variety.ItemType == "liquid" {
				waterVessel = item
			}
		} else {
			if emptyVessel == nil {
				emptyVessel = item
			}
		}
	}

	npos := needer.Pos()

	// Carrying water vessel → deliver directly
	if waterVessel != nil {
		nx, ny := NextStepBFS(pos.X, pos.Y, npos.X, npos.Y, gameMap)
		newActivity := "Bringing water to " + needer.Name
		if helper.CurrentActivity != newActivity {
			helper.CurrentActivity = newActivity
			if log != nil {
				log.Add(helper.ID, helper.Name, "activity", "Bringing water to "+needer.Name)
			}
		}
		return &entity.Intent{
			Target:          types.Position{X: nx, Y: ny},
			Dest:            npos,
			Action:          entity.ActionHelpWater,
			TargetItem:      waterVessel,
			TargetCharacter: needer,
		}
	}

	// Carrying empty vessel → fill phase
	if emptyVessel != nil {
		waterPos, found := gameMap.FindNearestWater(pos)
		if !found {
			return nil
		}
		adjX, adjY := FindClosestCardinalTile(pos.X, pos.Y, waterPos.X, waterPos.Y, gameMap)
		if adjX == -1 {
			return nil
		}
		nx, ny := NextStepBFS(pos.X, pos.Y, adjX, adjY, gameMap)
		newActivity := "Fetching water for " + needer.Name
		if helper.CurrentActivity != newActivity {
			helper.CurrentActivity = newActivity
			if log != nil {
				log.Add(helper.ID, helper.Name, "activity", "Fetching water for "+needer.Name)
			}
		}
		return &entity.Intent{
			Target:          types.Position{X: nx, Y: ny},
			Dest:            types.Position{X: adjX, Y: adjY},
			Action:          entity.ActionHelpWater,
			TargetItem:      emptyVessel,
			TargetCharacter: needer,
		}
	}

	// No vessel in inventory — look for ground vessel
	if !helper.HasInventorySpace() {
		return nil
	}

	// Prefer ground water vessel — already filled, fewer steps than empty vessel + fill phase
	if waterVessel := findGroundWaterVessel(pos, items); waterVessel != nil {
		vpos := waterVessel.Pos()
		nx, ny := NextStepBFS(pos.X, pos.Y, vpos.X, vpos.Y, gameMap)
		newActivity := "Bringing water to " + needer.Name
		if helper.CurrentActivity != newActivity {
			helper.CurrentActivity = newActivity
			if log != nil {
				log.Add(helper.ID, helper.Name, "activity", "Picking up water for "+needer.Name)
			}
		}
		return &entity.Intent{
			Target:          types.Position{X: nx, Y: ny},
			Dest:            vpos,
			Action:          entity.ActionHelpWater,
			TargetItem:      waterVessel,
			TargetCharacter: needer,
		}
	}

	// Fall back to ground empty vessel (needs fill phase)
	groundVessel := findEmptyGroundVessel(pos, items)
	if groundVessel == nil {
		return nil
	}

	vpos := groundVessel.Pos()
	nx, ny := NextStepBFS(pos.X, pos.Y, vpos.X, vpos.Y, gameMap)
	newActivity := "Fetching water for " + needer.Name
	if helper.CurrentActivity != newActivity {
		helper.CurrentActivity = newActivity
		if log != nil {
			log.Add(helper.ID, helper.Name, "activity", "Picking up vessel for "+needer.Name)
		}
	}
	return &entity.Intent{
		Target:          types.Position{X: nx, Y: ny},
		Dest:            vpos,
		Action:          entity.ActionHelpWater,
		TargetItem:      groundVessel,
		TargetCharacter: needer,
	}
}

// findHelpFeedIntent creates an intent to find food and deliver it to a needy character.
// Scores food candidates using ScoreFoodFit with Severe-tier weights (helper exercises judgment)
// and needer's hunger for satiation fit.
//
// Four candidate pools:
//   - Helper's carried loose food (distance 0)
//   - Helper's carried food vessel (distance 0, scored by contents)
//   - Ground loose food (distance from helper)
//   - Ground food vessel with edible contents (distance from helper, vessel is TargetItem)
//
// Returns nil if no food is available or helper can't carry anything.
func findHelpFeedIntent(helper *entity.Character, needer *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	prefWeight := config.FoodSeekPrefWeightSevere
	distWeight := config.FoodSeekDistWeightSevere

	// Determine healing bonus based on needer's health tier (helper's knowledge applies)
	var healingBonus float64
	healthTier := needer.HealthTier()
	switch healthTier {
	case entity.TierCrisis:
		healingBonus = config.HealingBonusCrisis
	case entity.TierSevere:
		healingBonus = config.HealingBonusSevere
	case entity.TierModerate:
		healingBonus = config.HealingBonusModerate
	case entity.TierMild:
		healingBonus = config.HealingBonusMild
	default:
		healingBonus = 0
	}

	var bestItem *entity.Item
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float
	bestDist := int(^uint(0) >> 1)              // Max int for tiebreaker
	bestCarried := false                        // Whether best candidate is in inventory

	// Helper to score and update best candidate
	scoreCandidate := func(item *entity.Item, dist int, netPref int, carried bool) {
		score := ScoreFoodFit(netPref, dist, needer.Hunger, item.ItemType, prefWeight, distWeight)

		// Apply healing bonus if helper knows this item is healing
		if healingBonus > 0 && helper.KnowsItemIsHealing(item) {
			score += healingBonus
		}

		if score > bestScore || (score == bestScore && dist < bestDist) {
			bestItem = item
			bestScore = score
			bestDist = dist
			bestCarried = carried
		}
	}

	// Score helper's carried items (distance = 0)
	for _, invItem := range helper.Inventory {
		if invItem == nil {
			continue
		}
		if invItem.Container != nil && len(invItem.Container.Contents) > 0 {
			// Carried food vessel — score by contents
			variety := invItem.Container.Contents[0].Variety
			if variety != nil && variety.IsEdible() {
				netPref := helper.NetPreferenceForVariety(variety)
				score := ScoreFoodFit(netPref, 0, needer.Hunger, variety.ItemType, prefWeight, distWeight)
				if healingBonus > 0 && helper.KnowsVarietyIsHealing(variety) {
					score += healingBonus
				}
				if score > bestScore || (score == bestScore && 0 < bestDist) {
					bestItem = invItem
					bestScore = score
					bestDist = 0
					bestCarried = true
				}
			}
		} else if invItem.IsEdible() {
			// Carried loose food
			netPref := helper.NetPreference(invItem)
			scoreCandidate(invItem, 0, netPref, true)
		}
	}

	// Score ground items (only if helper has inventory space OR already found carried food)
	canPickUp := helper.HasInventorySpace()

	if canPickUp {
		for _, item := range items {
			ipos := item.Pos()
			dist := pos.DistanceTo(ipos)

			if item.Container != nil && len(item.Container.Contents) > 0 {
				// Ground food vessel — score by contents
				variety := item.Container.Contents[0].Variety
				if variety != nil && variety.IsEdible() {
					netPref := helper.NetPreferenceForVariety(variety)
					score := ScoreFoodFit(netPref, dist, needer.Hunger, variety.ItemType, prefWeight, distWeight)
					if healingBonus > 0 && helper.KnowsVarietyIsHealing(variety) {
						score += healingBonus
					}
					if score > bestScore || (score == bestScore && dist < bestDist) {
						bestItem = item
						bestScore = score
						bestDist = dist
						bestCarried = false
					}
				}
			} else if item.IsEdible() {
				// Ground loose food
				netPref := helper.NetPreference(item)
				scoreCandidate(item, dist, netPref, false)
			}
		}
	}

	if bestItem == nil {
		return nil
	}

	npos := needer.Pos()

	if bestCarried {
		// Food already in inventory — go straight to delivery
		nx, ny := NextStepBFS(pos.X, pos.Y, npos.X, npos.Y, gameMap)
		newActivity := "Bringing food to " + needer.Name
		if helper.CurrentActivity != newActivity {
			helper.CurrentActivity = newActivity
			if log != nil {
				log.Add(helper.ID, helper.Name, "activity", "Bringing food to "+needer.Name)
			}
		}
		return &entity.Intent{
			Target:          types.Position{X: nx, Y: ny},
			Dest:            npos,
			Action:          entity.ActionHelpFeed,
			TargetItem:      bestItem,
			TargetCharacter: needer,
		}
	}

	// Food is on the ground — procurement phase: walk to food
	ipos := bestItem.Pos()
	nx, ny := NextStepBFS(pos.X, pos.Y, ipos.X, ipos.Y, gameMap)
	newActivity := "Getting food for " + needer.Name
	if helper.CurrentActivity != newActivity {
		helper.CurrentActivity = newActivity
		if log != nil {
			log.Add(helper.ID, helper.Name, "activity", "Getting food for "+needer.Name)
		}
	}
	return &entity.Intent{
		Target:          types.Position{X: nx, Y: ny},
		Dest:            types.Position{X: ipos.X, Y: ipos.Y},
		Action:          entity.ActionHelpFeed,
		TargetItem:      bestItem,
		TargetCharacter: needer,
	}
}
