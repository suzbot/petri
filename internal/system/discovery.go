package system

import (
	"fmt"
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
)

// GetDiscoveryChance returns the know-how discovery chance based on character mood.
// Joyful (TierNone) uses the configured rate, Happy (TierMild) uses 20% of that rate.
// All other moods return 0 (no discovery possible).
func GetDiscoveryChance(char *entity.Character) float64 {
	switch char.MoodTier() {
	case entity.TierNone: // Joyful
		return config.KnowHowDiscoveryChance
	case entity.TierMild: // Happy
		return config.KnowHowDiscoveryChance * 0.20
	default:
		return 0
	}
}

// TryDiscoverKnowHow attempts to discover know-how activities based on the action performed.
// Returns true if a new activity was discovered.
// The chance parameter allows testing with deterministic values; use config.KnowHowDiscoveryChance in production.
func TryDiscoverKnowHow(char *entity.Character, action entity.ActionType, item *entity.Item, log *ActionLog, chance float64) bool {
	if item == nil {
		return false
	}

	// Check each discoverable activity
	for _, activity := range entity.GetDiscoverableActivities() {
		// Skip if already known
		if char.KnowsActivity(activity.ID) {
			continue
		}

		// Check if any trigger matches
		for _, trigger := range activity.DiscoveryTriggers {
			if !triggerMatches(trigger, action, item) {
				continue
			}

			// Roll for discovery
			if rand.Float64() < chance {
				char.LearnActivity(activity.ID)
				if log != nil {
					log.Add(char.ID, char.Name, "discovery",
						fmt.Sprintf("Discovered how to %s!", activity.Name))
				}
				return true
			}
		}
	}

	return false
}

// triggerMatches checks if a discovery trigger matches the current action and item
func triggerMatches(trigger entity.DiscoveryTrigger, action entity.ActionType, item *entity.Item) bool {
	// Action must match
	if trigger.Action != action {
		return false
	}

	// Check item type filter (empty means any)
	if trigger.ItemType != "" && trigger.ItemType != item.ItemType {
		return false
	}

	// Check edible requirement
	if trigger.RequiresEdible && !item.Edible {
		return false
	}

	return true
}
