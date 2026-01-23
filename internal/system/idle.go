package system

import (
	"math/rand"
	"strings"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
)

// selectIdleActivity checks for order work first, then randomly selects an idle activity.
// Returns nil if cooldown is active or if the selected activity cannot be performed.
// Sets IdleCooldown after being called (regardless of outcome).
func selectIdleActivity(char *entity.Character, cx, cy int, items []*entity.Item, gameMap *game.Map, log *ActionLog, orders []*entity.Order) *entity.Intent {
	// Priority: if character has an assigned order, always try to resume it (bypass cooldown)
	// This ensures order work isn't blocked by idle cooldown when target changes
	if char.AssignedOrderID != 0 {
		if intent := selectOrderActivity(char, cx, cy, items, orders, log); intent != nil {
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
	if intent := selectOrderActivity(char, cx, cy, items, orders, log); intent != nil {
		return intent
	}

	// Roll 0-3 for activity selection (equal 1/4 probability each)
	roll := rand.Intn(4)

	switch roll {
	case 0:
		// Try looking
		if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try other activities
		if intent := findTalkIntent(char, cx, cy, gameMap, log); intent != nil {
			return intent
		}
		if !char.IsInventoryFull() {
			if intent := findForageIntent(char, cx, cy, items, log); intent != nil {
				return intent
			}
		}
	case 1:
		// Try talking
		if intent := findTalkIntent(char, cx, cy, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try looking
		if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
		if !char.IsInventoryFull() {
			if intent := findForageIntent(char, cx, cy, items, log); intent != nil {
				return intent
			}
		}
	case 2:
		// Try foraging (only if inventory not full)
		if !char.IsInventoryFull() {
			if intent := findForageIntent(char, cx, cy, items, log); intent != nil {
				return intent
			}
		}
		// Fall through to try other activities
		if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
		if intent := findTalkIntent(char, cx, cy, gameMap, log); intent != nil {
			return intent
		}
	case 3:
		// Stay idle - return nil
		return nil
	}

	// Nothing available
	return nil
}

// isIdleActivity returns true if the activity string represents an idle activity
// that can be interrupted for talking. Idle activities are: Idle, Looking, Talking, Foraging.
func isIdleActivity(activity string) bool {
	if strings.HasPrefix(activity, "Idle") {
		return true
	}
	if strings.HasPrefix(activity, "Looking") {
		return true
	}
	if strings.HasPrefix(activity, "Talking") {
		return true
	}
	if strings.HasPrefix(activity, "Foraging") {
		return true
	}
	if strings.HasPrefix(activity, "Moving to forage") {
		return true
	}
	return false
}
