package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// selectIdleActivity checks for order work first, then randomly selects an idle activity.
// Returns nil if cooldown is active or if the selected activity cannot be performed.
// Sets IdleCooldown after being called (regardless of outcome).
func selectIdleActivity(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, log *ActionLog, orders []*entity.Order) *entity.Intent {
	// Priority: if character has an assigned order, always try to resume it (bypass cooldown)
	// This ensures order work isn't blocked by idle cooldown when target changes
	if char.AssignedOrderID != 0 {
		if intent := selectOrderActivity(char, pos, items, gameMap, orders, log); intent != nil {
			return intent
		}
		// Order couldn't find a target - fall through to idle activities
	}

	// Check cooldown for idle activities (looking, talking, foraging, taking new orders)
	if char.IdleCooldown > 0 {
		return nil
	}

	// Set cooldown for next attempt
	char.IdleCooldown = config.IdleCooldown

	// Check for new order work (taking unassigned orders)
	if intent := selectOrderActivity(char, pos, items, gameMap, orders, log); intent != nil {
		return intent
	}

	// Helping override: before the random roll, check for crisis characters
	if needer := findNearestCrisisCharacter(char, gameMap.Characters()); needer != nil {
		if needer.HungerTier() == entity.TierCrisis {
			if intent := findHelpFeedIntent(char, needer, pos, items, gameMap, log); intent != nil {
				return intent
			}
		}
	}

	// Roll 0-4 for activity selection (equal 1/5 probability each)
	roll := rand.Intn(5)

	switch roll {
	case 0:
		// Try looking
		if intent := findLookIntent(char, pos, items, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try other activities
		if intent := findTalkIntent(char, pos, gameMap, log); intent != nil {
			return intent
		}
		if CanPickUpMore(char, gameMap.Varieties()) {
			if intent := findForageIntent(char, pos, items, log, gameMap.Varieties(), gameMap); intent != nil {
				return intent
			}
		}
	case 1:
		// Try talking
		if intent := findTalkIntent(char, pos, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try looking
		if intent := findLookIntent(char, pos, items, gameMap, log); intent != nil {
			return intent
		}
		if CanPickUpMore(char, gameMap.Varieties()) {
			if intent := findForageIntent(char, pos, items, log, gameMap.Varieties(), gameMap); intent != nil {
				return intent
			}
		}
	case 2:
		// Try foraging (only if can pick up more)
		if CanPickUpMore(char, gameMap.Varieties()) {
			if intent := findForageIntent(char, pos, items, log, gameMap.Varieties(), gameMap); intent != nil {
				return intent
			}
		}
		// Fall through to try other activities
		if intent := findLookIntent(char, pos, items, gameMap, log); intent != nil {
			return intent
		}
		if intent := findTalkIntent(char, pos, gameMap, log); intent != nil {
			return intent
		}
	case 3:
		// Try fetch water
		if intent := findFetchWaterIntent(char, pos, items, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try other activities
		if intent := findLookIntent(char, pos, items, gameMap, log); intent != nil {
			return intent
		}
	case 4:
		// Stay idle - return nil
		return nil
	}

	// Nothing available
	return nil
}

// isIdleAction returns true if the character is performing an idle activity
// that can be interrupted for talking. Checks ActionType instead of activity strings.
// Characters with no intent (nil) are considered idle.
func isIdleAction(char *entity.Character) bool {
	if char.Intent == nil {
		return true
	}
	switch char.Intent.Action {
	case entity.ActionNone, entity.ActionLook, entity.ActionTalk, entity.ActionForage, entity.ActionFillVessel:
		return true
	default:
		return false
	}
}
