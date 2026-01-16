package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// CalculateIntent determines what a character wants to do next tick
// This is safe to call concurrently - it only reads world state
func CalculateIntent(char *entity.Character, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	if char.IsDead || char.IsSleeping {
		return nil
	}

	// If frustrated, stay idle until timer expires
	if char.IsFrustrated {
		if char.CurrentActivity != "Frustrated" {
			char.CurrentActivity = "Frustrated"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Frustrated (can't meet needs)")
			}
		}
		return nil
	}

	cx, cy := char.Position()

	// Cache tier values (calculated once, reused throughout)
	hungerTier := char.HungerTier()
	thirstTier := char.ThirstTier()
	energyTier := char.EnergyTier()
	healthTier := char.HealthTier()

	// Check if we should continue an idle activity (looking or talking)
	// Idle activities can be interrupted by urgent needs (tier >= Moderate)
	if char.Intent != nil && char.Intent.DrivingStat == "" {
		maxTier := hungerTier
		if thirstTier > maxTier {
			maxTier = thirstTier
		}
		if energyTier > maxTier {
			maxTier = energyTier
		}
		// Health only counts if we can fulfill it (have healing knowledge)
		if healthTier > maxTier && canFulfillHealth(char, items) {
			maxTier = healthTier
		}

		// Continue looking if no urgent needs
		if char.Intent.TargetItem != nil && maxTier < entity.TierModerate {
			return continueIntent(char, cx, cy, gameMap, log)
		}

		// Continue talking/approaching if no urgent needs
		if char.Intent.TargetCharacter != nil && maxTier < entity.TierModerate {
			target := char.Intent.TargetCharacter
			// If already talking, continue
			if char.TalkingWith != nil {
				return &entity.Intent{
					TargetX:         cx,
					TargetY:         cy,
					Action:          entity.ActionTalk,
					TargetCharacter: char.TalkingWith,
				}
			}
			// If approaching to talk, check target is still valid (idle activity, not dead/sleeping)
			if !target.IsDead && !target.IsSleeping && isIdleActivity(target.CurrentActivity) {
				return continueIntent(char, cx, cy, gameMap, log)
			}
			// Target no longer valid, fall through to re-evaluate
		}

		// If talking but interrupted by Moderate+ need, stop partner too
		if char.TalkingWith != nil && maxTier >= entity.TierModerate {
			StopTalking(char, char.TalkingWith, log)
		}
		// Fall through to re-evaluate
	}

	// Check if current intent should be kept (tier-based evaluation)
	if char.Intent != nil {
		currentDrivingTier := char.Intent.DrivingTier

		// Get the current tier of the driving stat
		var currentStatTier int
		switch char.Intent.DrivingStat {
		case types.StatHunger:
			currentStatTier = hungerTier
		case types.StatThirst:
			currentStatTier = thirstTier
		case types.StatEnergy:
			currentStatTier = energyTier
		case types.StatHealth:
			currentStatTier = healthTier
		}

		shouldReEval := false

		// Re-evaluate if the driving stat has been satisfied
		// Special case: when drinking at a spring, keep drinking until thirst == 0
		if char.Intent.DrivingStat == types.StatThirst && char.Intent.Action == entity.ActionDrink {
			// At a spring - only stop when fully satisfied
			if char.Thirst == 0 {
				shouldReEval = true
			}
		} else if currentStatTier == entity.TierNone {
			// Standard behavior: re-evaluate when tier drops to None
			shouldReEval = true
		}

		// Re-evaluate if a different stat has reached a higher tier AND can be fulfilled
		// This prevents thrashing between an unfulfillable higher-tier stat and a fulfillable lower-tier stat
		if !shouldReEval {
			switch char.Intent.DrivingStat {
			case types.StatHunger:
				if thirstTier > currentDrivingTier && canFulfillThirst(gameMap, cx, cy) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cx, cy) {
					shouldReEval = true
				} else if healthTier > currentDrivingTier && canFulfillHealth(char, items) {
					shouldReEval = true
				}
			case types.StatThirst:
				if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cx, cy) {
					shouldReEval = true
				} else if healthTier > currentDrivingTier && canFulfillHealth(char, items) {
					shouldReEval = true
				}
			case types.StatEnergy:
				if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if thirstTier > currentDrivingTier && canFulfillThirst(gameMap, cx, cy) {
					shouldReEval = true
				} else if healthTier > currentDrivingTier && canFulfillHealth(char, items) {
					shouldReEval = true
				}
			case types.StatHealth:
				if thirstTier > currentDrivingTier && canFulfillThirst(gameMap, cx, cy) {
					shouldReEval = true
				} else if hungerTier > currentDrivingTier && canFulfillHunger(char, items) {
					shouldReEval = true
				} else if energyTier > currentDrivingTier && canFulfillEnergy(char, gameMap, cx, cy) {
					shouldReEval = true
				}
			default:
				shouldReEval = true
			}
		}

		if !shouldReEval {
			// Continue with current intent - recalculate next step
			return continueIntent(char, cx, cy, gameMap, log)
		}
	}

	// Evaluate which stat to address based on tier and tie-breakers

	// Find the highest tier (health only counts if we can fulfill it)
	maxTier := hungerTier
	if thirstTier > maxTier {
		maxTier = thirstTier
	}
	if energyTier > maxTier {
		maxTier = energyTier
	}
	if healthTier > maxTier && canFulfillHealth(char, items) {
		maxTier = healthTier
	}

	// No urgent needs - try an idle activity (looking, talking, or staying idle)
	if maxTier == entity.TierNone {
		if intent := selectIdleActivity(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no needs)")
			}
		}
		return nil
	}

	// Build priority list: stats with needs, sorted by tier (desc), then tie-breaker (Thirst > Hunger > Health > Energy)
	type statPriority struct {
		stat types.StatType
		tier int
	}
	var priorities []statPriority

	// Add stats that have needs (tier > 0), in tie-breaker order within same tier
	if thirstTier > 0 {
		priorities = append(priorities, statPriority{types.StatThirst, thirstTier})
	}
	if hungerTier > 0 {
		priorities = append(priorities, statPriority{types.StatHunger, hungerTier})
	}
	// Health only added if can be fulfilled (requires healing knowledge)
	if healthTier > 0 && canFulfillHealth(char, items) {
		priorities = append(priorities, statPriority{types.StatHealth, healthTier})
	}
	if energyTier > 0 {
		priorities = append(priorities, statPriority{types.StatEnergy, energyTier})
	}

	// Sort by tier descending (higher tier = more urgent)
	// Tie-breaker order is already correct since we added in Thirst > Hunger > Health > Energy order
	for i := 0; i < len(priorities)-1; i++ {
		for j := i + 1; j < len(priorities); j++ {
			if priorities[j].tier > priorities[i].tier {
				priorities[i], priorities[j] = priorities[j], priorities[i]
			}
		}
	}

	// Try each stat in priority order, falling back if intent can't be fulfilled
	for _, p := range priorities {
		var intent *entity.Intent
		switch p.stat {
		case types.StatThirst:
			intent = findDrinkIntent(char, cx, cy, gameMap, p.tier, log)
		case types.StatHunger:
			intent = findFoodIntent(char, cx, cy, items, p.tier, log)
		case types.StatHealth:
			intent = findHealingIntent(char, cx, cy, items, p.tier, log)
		case types.StatEnergy:
			intent = findSleepIntent(char, cx, cy, gameMap, p.tier, log)
		}
		if intent != nil {
			// Successfully found an intent - reset failure counter
			char.FailedIntentCount = 0
			return intent
		}
	}

	// No intent could be fulfilled - track failure only at Severe+ tier (prevents thrashing where it matters)
	if maxTier >= entity.TierSevere {
		char.FailedIntentCount++
		if char.FailedIntentCount >= config.FrustrationThreshold {
			char.IsFrustrated = true
			char.FrustrationTimer = config.FrustrationDuration
			char.FailedIntentCount = 0
			char.CurrentActivity = "Frustrated"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Frustrated (can't meet needs)")
			}
			return nil
		}
	}

	// No needs could be fulfilled - try an idle activity
	if intent := selectIdleActivity(char, cx, cy, items, gameMap, log); intent != nil {
		return intent
	}

	if char.CurrentActivity != "Idle" {
		char.CurrentActivity = "Idle"
		if log != nil {
			log.Add(char.ID, char.Name, "activity", "Idle (no options)")
		}
	}
	return nil
}

// continueIntent recalculates the next step for an existing intent
func continueIntent(char *entity.Character, cx, cy int, gameMap *game.Map, log *ActionLog) *entity.Intent {
	intent := char.Intent

	// Check if target item still exists at expected position (O(1) instead of O(n))
	if intent.TargetItem != nil {
		ix, iy := intent.TargetItem.Position()
		if gameMap.ItemAt(ix, iy) != intent.TargetItem {
			return nil // Target consumed by someone else
		}
	}

	// Check if target feature is now occupied by another character
	if intent.TargetFeature != nil {
		fx, fy := intent.TargetFeature.Position()
		occupant := gameMap.CharacterAt(fx, fy)
		if occupant != nil && occupant != char {
			return nil // Target occupied by someone else, find new target
		}
	}

	// Recalculate next step toward target
	var tx, ty int
	if intent.TargetItem != nil {
		tx, ty = intent.TargetItem.Position()
	} else if intent.TargetFeature != nil {
		tx, ty = intent.TargetFeature.Position()
	} else if intent.TargetCharacter != nil {
		tx, ty = intent.TargetCharacter.Position()
	} else {
		tx, ty = intent.TargetX, intent.TargetY
	}

	// Check if we've arrived at a feature target - switch to appropriate action
	if intent.TargetFeature != nil && cx == tx && cy == ty {
		feature := intent.TargetFeature
		if feature.IsDrinkSource() {
			newActivity := "Drinking"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
				if log != nil {
					log.Add(char.ID, char.Name, "thirst", "Drinking from spring")
				}
			}
			return &entity.Intent{
				TargetX:       tx,
				TargetY:       ty,
				Action:        entity.ActionDrink,
				TargetFeature: feature,
				DrivingStat:   intent.DrivingStat,
				DrivingTier:   intent.DrivingTier,
			}
		}
		if feature.IsBed() {
			newActivity := "Sleeping (in bed)"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				TargetX:       tx,
				TargetY:       ty,
				Action:        entity.ActionSleep,
				TargetFeature: feature,
				DrivingStat:   intent.DrivingStat,
				DrivingTier:   intent.DrivingTier,
			}
		}
	}

	// Check if we've arrived adjacent to an item for looking (no DrivingStat means looking intent)
	// Skip if ActionPickup (foraging) - those need to move onto the item, not stay adjacent
	if intent.TargetItem != nil && intent.DrivingStat == "" && intent.Action != entity.ActionPickup && isAdjacent(cx, cy, tx, ty) {
		newActivity := "Looking at " + intent.TargetItem.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			TargetX:    cx, // Stay in place
			TargetY:    cy,
			Action:     entity.ActionLook,
			TargetItem: intent.TargetItem,
		}
	}

	// Check if we've arrived adjacent to a character for talking
	if intent.TargetCharacter != nil && isAdjacent(cx, cy, tx, ty) {
		return &entity.Intent{
			TargetX:         cx, // Stay in place
			TargetY:         cy,
			Action:          entity.ActionTalk,
			TargetCharacter: intent.TargetCharacter,
		}
	}

	nx, ny := NextStep(cx, cy, tx, ty)

	return &entity.Intent{
		TargetX:         nx,
		TargetY:         ny,
		Action:          intent.Action,
		TargetItem:      intent.TargetItem,
		TargetFeature:   intent.TargetFeature,
		TargetCharacter: intent.TargetCharacter,
		DrivingStat:     intent.DrivingStat,
		DrivingTier:     intent.DrivingTier,
	}
}

// findDrinkIntent finds a spring to drink from
func findDrinkIntent(char *entity.Character, cx, cy int, gameMap *game.Map, tier int, log *ActionLog) *entity.Intent {
	spring := gameMap.FindNearestDrinkSource(cx, cy)
	if spring == nil {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no water source)")
			}
		}
		return nil
	}

	tx, ty := spring.Position()

	// Already at spring - drink
	if cx == tx && cy == ty {
		newActivity := "Drinking"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "thirst", "Drinking from spring")
			}
		}
		return &entity.Intent{
			TargetX:       tx,
			TargetY:       ty,
			Action:        entity.ActionDrink,
			TargetFeature: spring,
			DrivingStat:   types.StatThirst,
			DrivingTier:   tier,
		}
	}

	// Move toward spring
	nx, ny := NextStep(cx, cy, tx, ty)
	newActivity := "Moving to spring"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Heading to spring")
		}
	}

	return &entity.Intent{
		TargetX:       nx,
		TargetY:       ny,
		Action:        entity.ActionMove,
		TargetFeature: spring,
		DrivingStat:   types.StatThirst,
		DrivingTier:   tier,
	}
}

// findFoodIntent finds food based on hunger priority
func findFoodIntent(char *entity.Character, cx, cy int, items []*entity.Item, tier int, log *ActionLog) *entity.Intent {
	result := findFoodTarget(char, items)
	if result.Item == nil {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no suitable food)")
			}
		}
		return nil
	}

	tx, ty := result.Item.Position()
	nx, ny := NextStep(cx, cy, tx, ty)

	newActivity := "Moving to " + result.Item.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			// Include scores in parentheses (stripped in non-debug mode)
			log.Add(char.ID, char.Name, "movement",
				fmt.Sprintf("Started moving to %s (pref:%d score:%.0f)",
					result.Item.Description(), result.NetPreference, result.GradientScore))
		}
	}

	return &entity.Intent{
		TargetX:     nx,
		TargetY:     ny,
		Action:      entity.ActionMove,
		TargetItem:  result.Item,
		DrivingStat: types.StatHunger,
		DrivingTier: tier,
	}
}

// findHealingIntent finds a known healing item to consume.
// Only considers items the character knows are healing.
// Returns nil if no known healing items are available.
func findHealingIntent(char *entity.Character, cx, cy int, items []*entity.Item, tier int, log *ActionLog) *entity.Intent {
	// Get only items the character knows are healing
	knownHealing := char.KnownHealingItems(items)
	if len(knownHealing) == 0 {
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no known healing items)")
			}
		}
		return nil
	}

	// Find nearest known healing item
	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1) // Max int

	for _, item := range knownHealing {
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
	nx, ny := NextStep(cx, cy, tx, ty)

	newActivity := "Moving to " + nearest.Description() + " (healing)"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement",
				fmt.Sprintf("Seeking healing: %s", nearest.Description()))
		}
	}

	return &entity.Intent{
		TargetX:     nx,
		TargetY:     ny,
		Action:      entity.ActionMove,
		TargetItem:  nearest,
		DrivingStat: types.StatHealth,
		DrivingTier: tier,
	}
}

// findSleepIntent finds a bed to sleep in
func findSleepIntent(char *entity.Character, cx, cy int, gameMap *game.Map, tier int, log *ActionLog) *entity.Intent {
	bed := gameMap.FindNearestBed(cx, cy)

	// If no bed, can sleep on ground when exhausted (voluntary) or collapsed (involuntary)
	if bed == nil {
		if char.Energy <= 10 { // Exhausted - ground sleep available
			newActivity := "Sleeping (on ground)"
			if char.CurrentActivity != newActivity {
				char.CurrentActivity = newActivity
			}
			return &entity.Intent{
				TargetX:     cx,
				TargetY:     cy,
				Action:      entity.ActionSleep,
				DrivingStat: types.StatEnergy,
				DrivingTier: tier,
			}
		}
		// Not tired enough for ground sleep, need a bed
		if char.CurrentActivity != "Idle" {
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Idle (no bed nearby)")
			}
		}
		return nil
	}

	tx, ty := bed.Position()

	// Already at bed - sleep
	if cx == tx && cy == ty {
		newActivity := "Sleeping (in bed)"
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
		}
		return &entity.Intent{
			TargetX:       tx,
			TargetY:       ty,
			Action:        entity.ActionSleep,
			TargetFeature: bed,
			DrivingStat:   types.StatEnergy,
			DrivingTier:   tier,
		}
	}

	// Move toward bed
	nx, ny := NextStep(cx, cy, tx, ty)
	newActivity := "Moving to leaf pile"
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Heading to leaf pile")
		}
	}

	return &entity.Intent{
		TargetX:       nx,
		TargetY:       ny,
		Action:        entity.ActionMove,
		TargetFeature: bed,
		DrivingStat:   types.StatEnergy,
		DrivingTier:   tier,
	}
}

// FoodTargetResult contains the selected food target and scoring info
type FoodTargetResult struct {
	Item          *entity.Item
	NetPreference int
	GradientScore float64
}

// findFoodTarget finds the best item for a character based on hunger level
// Uses gradient scoring: Score = (NetPreference × PrefWeight) - (Distance × DistWeight) + HealingBonus
// Hunger tier affects both the preference weight and which items are considered:
// - Moderate (50-74): High pref weight, only NetPreference >= 0 items considered
// - Severe (75-89): Medium pref weight, all items considered
// - Crisis (90+): No pref weight (just distance), all items considered
// Healing bonus: When health tier >= Moderate and character knows item is healing,
// adds bonus to score (larger bonus at worse health tiers)
func findFoodTarget(char *entity.Character, items []*entity.Item) FoodTargetResult {
	if len(items) == 0 {
		return FoodTargetResult{}
	}

	cx, cy := char.Position()

	// Determine hunger tier and corresponding weights/filters
	var prefWeight float64
	var filterDisliked bool

	if char.Hunger >= 90 {
		// Crisis: just pick nearest
		prefWeight = config.FoodSeekPrefWeightCrisis
		filterDisliked = false
	} else if char.Hunger >= 75 {
		// Severe: gradient with medium pref weight, consider all items
		prefWeight = config.FoodSeekPrefWeightSevere
		filterDisliked = false
	} else {
		// Moderate: gradient with high pref weight, filter disliked
		prefWeight = config.FoodSeekPrefWeightModerate
		filterDisliked = true
	}

	// Determine healing bonus based on health tier (only if hurt)
	var healingBonus float64
	healthTier := char.HealthTier()
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
		healingBonus = 0 // No bonus at TierNone (full health)
	}

	var bestItem *entity.Item
	var bestNetPref int
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float
	bestDist := int(^uint(0) >> 1)              // Max int for distance tiebreaker

	for _, item := range items {
		// Skip non-edible items (e.g., flowers)
		if !item.Edible {
			continue
		}

		netPref := char.NetPreference(item)

		// At Moderate hunger, filter out disliked items (NetPreference < 0)
		if filterDisliked && netPref < 0 {
			continue
		}

		ix, iy := item.Position()
		dist := abs(cx-ix) + abs(cy-iy)

		// Calculate gradient score
		score := float64(netPref)*prefWeight - float64(dist)*config.FoodSeekDistWeight

		// Apply healing bonus if character knows this item is healing
		if healingBonus > 0 && char.KnowsItemIsHealing(item) {
			score += healingBonus
		}

		// Update best if better score, or same score but closer (distance tiebreaker)
		if score > bestScore || (score == bestScore && dist < bestDist) {
			bestItem = item
			bestNetPref = netPref
			bestScore = score
			bestDist = dist
		}
	}

	return FoodTargetResult{
		Item:          bestItem,
		NetPreference: bestNetPref,
		GradientScore: bestScore,
	}
}

// NextStep calculates the next position moving toward target
func NextStep(fromX, fromY, toX, toY int) (int, int) {
	dx := toX - fromX
	dy := toY - fromY

	if dx == 0 && dy == 0 {
		return fromX, fromY
	}

	// Move toward larger distance
	if abs(dx) > abs(dy) {
		return fromX + sign(dx), fromY
	}
	return fromX, fromY + sign(dy)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// canFulfillThirst checks if thirst can be addressed (water source exists)
func canFulfillThirst(gameMap *game.Map, cx, cy int) bool {
	return gameMap.FindNearestDrinkSource(cx, cy) != nil
}

// canFulfillHunger checks if hunger can be addressed (suitable food exists)
func canFulfillHunger(char *entity.Character, items []*entity.Item) bool {
	return findFoodTarget(char, items).Item != nil
}

// canFulfillEnergy checks if energy can be addressed (bed exists or exhausted enough for ground sleep)
func canFulfillEnergy(char *entity.Character, gameMap *game.Map, cx, cy int) bool {
	// Can sleep on ground if exhausted (voluntary at ≤10, involuntary collapse at 0)
	if char.Energy <= 10 {
		return true
	}
	// Otherwise need a bed
	return gameMap.FindNearestBed(cx, cy) != nil
}

// canFulfillHealth checks if health can be addressed (character knows healing items and they exist)
func canFulfillHealth(char *entity.Character, items []*entity.Item) bool {
	return len(char.KnownHealingItems(items)) > 0
}

// findLookIntent creates an intent to look at the nearest item.
// Called by selectIdleActivity when looking is selected.
func findLookIntent(char *entity.Character, cx, cy int, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Find nearest item, excluding last looked item
	target := findNearestItemExcluding(cx, cy, items, char.LastLookedX, char.LastLookedY, char.HasLastLooked)
	if target == nil {
		return nil
	}

	tx, ty := target.Position()

	// Check if already adjacent to target
	if isAdjacent(cx, cy, tx, ty) {
		// Start looking immediately
		newActivity := "Looking at " + target.Description()
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Looking at "+target.Description())
			}
		}
		return &entity.Intent{
			TargetX:    cx, // Stay in place
			TargetY:    cy,
			Action:     entity.ActionLook,
			TargetItem: target,
		}
	}

	// Find closest adjacent tile to target
	adjX, adjY := findClosestAdjacentTile(cx, cy, tx, ty, gameMap)
	if adjX == -1 {
		return nil // No accessible adjacent tile
	}

	// Move toward adjacent tile
	nx, ny := NextStep(cx, cy, adjX, adjY)

	newActivity := "Moving to look at " + target.Description()
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Moving to look at "+target.Description())
		}
	}

	return &entity.Intent{
		TargetX:    nx,
		TargetY:    ny,
		Action:     entity.ActionMove,
		TargetItem: target,
	}
}

// findForageIntent creates an intent to forage (pick up) an edible item.
// Called by selectIdleActivity when foraging is selected.
// Uses preference/distance scoring similar to eating but without hunger-based filtering.
func findForageIntent(char *entity.Character, cx, cy int, items []*entity.Item, log *ActionLog) *entity.Intent {
	// Find best edible item using preference/distance gradient
	target := findForageTarget(char, cx, cy, items)
	if target == nil {
		return nil
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
func findForageTarget(char *entity.Character, cx, cy int, items []*entity.Item) *entity.Item {
	if len(items) == 0 {
		return nil
	}

	var bestItem *entity.Item
	bestScore := float64(int(^uint(0)>>1)) * -1 // Negative max float
	bestDist := int(^uint(0) >> 1)              // Max int for distance tiebreaker

	for _, item := range items {
		// Only consider edible items for foraging
		if !item.Edible {
			continue
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

// findNearestItem finds the closest item of any type to the given position
func findNearestItem(cx, cy int, items []*entity.Item) *entity.Item {
	return findNearestItemExcluding(cx, cy, items, 0, 0, false)
}

// findNearestItemExcluding finds the closest item, optionally excluding a specific position
func findNearestItemExcluding(cx, cy int, items []*entity.Item, excludeX, excludeY int, hasExclude bool) *entity.Item {
	if len(items) == 0 {
		return nil
	}

	var nearest *entity.Item
	nearestDist := int(^uint(0) >> 1)

	for _, item := range items {
		ix, iy := item.Position()

		// Skip excluded position
		if hasExclude && ix == excludeX && iy == excludeY {
			continue
		}

		dist := abs(cx-ix) + abs(cy-iy)
		if dist < nearestDist {
			nearestDist = dist
			nearest = item
		}
	}

	return nearest
}

// isAdjacent checks if two positions are adjacent (including diagonals)
func isAdjacent(x1, y1, x2, y2 int) bool {
	dx := abs(x1 - x2)
	dy := abs(y1 - y2)
	return dx <= 1 && dy <= 1 && !(dx == 0 && dy == 0)
}

// findClosestAdjacentTile finds the closest unoccupied tile adjacent to (tx, ty) from position (cx, cy)
func findClosestAdjacentTile(cx, cy, tx, ty int, gameMap *game.Map) (int, int) {
	// 8 directions
	directions := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}

	bestX, bestY := -1, -1
	bestDist := int(^uint(0) >> 1)

	for _, dir := range directions {
		ax, ay := tx+dir[0], ty+dir[1]
		if !gameMap.IsValid(ax, ay) {
			continue
		}
		if gameMap.IsOccupied(ax, ay) {
			continue
		}

		dist := abs(cx-ax) + abs(cy-ay)
		if dist < bestDist {
			bestDist = dist
			bestX, bestY = ax, ay
		}
	}

	return bestX, bestY
}
