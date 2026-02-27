package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// selectDiscretionaryActivity randomly selects a leisure activity (look, talk, forage, fetch water).
// Returns nil if cooldown is active or if the selected activity cannot be performed.
// Sets IdleCooldown after being called (regardless of outcome).
// Orders and helping are evaluated separately by the orchestrator — this function
// handles only the discretionary bucket.
func selectDiscretionaryActivity(char *entity.Character, pos types.Position, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Check cooldown for discretionary activities (looking, talking, foraging, fetch water)
	if char.IdleCooldown > 0 {
		return nil
	}

	// Set cooldown for next attempt
	char.IdleCooldown = config.IdleCooldown

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

// isDiscretionaryAction returns true if the character is performing a discretionary activity
// that can be interrupted for talking. Checks ActionType instead of activity strings.
// Characters with no intent (nil) are considered available.
func isDiscretionaryAction(char *entity.Character) bool {
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
