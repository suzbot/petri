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

// TryDiscoverKnowHow attempts to discover know-how based on the action performed.
// Checks both activity triggers (direct discovery) and recipe triggers (grants activity + recipe).
// Returns true if something new was discovered.
// The chance parameter allows testing with deterministic values; use config.KnowHowDiscoveryChance in production.
func TryDiscoverKnowHow(char *entity.Character, action entity.ActionType, item *entity.Item, log *ActionLog, chance float64) bool {
	// Try activity-based discovery (e.g., harvest)
	if tryDiscoverActivity(char, action, item, log, chance) {
		return true
	}

	// Try recipe-based discovery (e.g., craftVessel via hollow-gourd recipe)
	if tryDiscoverRecipe(char, action, item, log, chance) {
		return true
	}

	return false
}

// tryDiscoverActivity attempts to discover activities with direct triggers (like harvest)
func tryDiscoverActivity(char *entity.Character, action entity.ActionType, item *entity.Item, log *ActionLog, chance float64) bool {
	for _, activity := range entity.GetDiscoverableActivities() {
		// Skip if already known
		if char.KnowsActivity(activity.ID) {
			continue
		}

		// Skip activities without direct triggers (discovered via recipes instead)
		if len(activity.DiscoveryTriggers) == 0 {
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

// tryDiscoverRecipe attempts to discover recipes, granting both the recipe and its activity
func tryDiscoverRecipe(char *entity.Character, action entity.ActionType, item *entity.Item, log *ActionLog, chance float64) bool {
	for _, recipe := range entity.GetDiscoverableRecipes() {
		// Skip if recipe already known
		if char.KnowsRecipe(recipe.ID) {
			continue
		}

		// Check if any trigger matches
		for _, trigger := range recipe.DiscoveryTriggers {
			if !triggerMatches(trigger, action, item) {
				continue
			}

			// Roll for discovery
			if rand.Float64() < chance {
				// Grant the activity (if not already known)
				activityLearned := char.LearnActivity(recipe.ActivityID)

				// Grant the recipe
				char.LearnRecipe(recipe.ID)

				// Log discovery
				if log != nil {
					activity := entity.ActivityRegistry[recipe.ActivityID]
					if activityLearned {
						log.Add(char.ID, char.Name, "discovery",
							fmt.Sprintf("Discovered how to %s %s!", activityCategoryVerb(activity.Category), activity.Name))
					}
					log.Add(char.ID, char.Name, "discovery",
						fmt.Sprintf("Learned %s recipe!", recipe.Name))
				}

				// Grant bundled activities
				for _, bundledID := range recipe.BundledActivities {
					if char.LearnActivity(bundledID) {
						if log != nil {
							bundledActivity := entity.ActivityRegistry[bundledID]
							log.Add(char.ID, char.Name, "discovery",
								fmt.Sprintf("Discovered how to %s!", bundledActivity.Name))
						}
					}
				}

				return true
			}
		}
	}

	return false
}

// activityCategoryVerb returns a verb suitable for "Discovered how to [verb] [name]!" messages.
func activityCategoryVerb(category string) string {
	switch category {
	case "construction":
		return "build"
	default:
		return "craft"
	}
}

// TryDiscoverFromConstruct is the entry point for construct-based discovery.
// Mirrors TryDiscoverKnowHow: tries activities then recipes.
func TryDiscoverFromConstruct(char *entity.Character, action entity.ActionType, constructKind string, constructMaterial string, log *ActionLog, chance float64) bool {
	if tryDiscoverActivityFromConstruct(char, action, constructKind, constructMaterial, log, chance) {
		return true
	}

	if tryDiscoverRecipeFromConstruct(char, action, constructKind, constructMaterial, log, chance) {
		return true
	}

	return false
}

// tryDiscoverActivityFromConstruct attempts to discover activities with construct-based triggers
func tryDiscoverActivityFromConstruct(char *entity.Character, action entity.ActionType, constructKind string, constructMaterial string, log *ActionLog, chance float64) bool {
	for _, activity := range entity.GetDiscoverableActivities() {
		if char.KnowsActivity(activity.ID) {
			continue
		}

		if len(activity.DiscoveryTriggers) == 0 {
			continue
		}

		for _, trigger := range activity.DiscoveryTriggers {
			if !constructTriggerMatches(trigger, action, constructKind, constructMaterial) {
				continue
			}

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

// tryDiscoverRecipeFromConstruct attempts to discover recipes with construct-based triggers
func tryDiscoverRecipeFromConstruct(char *entity.Character, action entity.ActionType, constructKind string, constructMaterial string, log *ActionLog, chance float64) bool {
	for _, recipe := range entity.GetDiscoverableRecipes() {
		if char.KnowsRecipe(recipe.ID) {
			continue
		}

		for _, trigger := range recipe.DiscoveryTriggers {
			if !constructTriggerMatches(trigger, action, constructKind, constructMaterial) {
				continue
			}

			if rand.Float64() < chance {
				activityLearned := char.LearnActivity(recipe.ActivityID)
				char.LearnRecipe(recipe.ID)

				if log != nil {
					activity := entity.ActivityRegistry[recipe.ActivityID]
					if activityLearned {
						log.Add(char.ID, char.Name, "discovery",
							fmt.Sprintf("Discovered how to %s %s!", activityCategoryVerb(activity.Category), activity.Name))
					}
					log.Add(char.ID, char.Name, "discovery",
						fmt.Sprintf("Learned %s recipe!", recipe.Name))
				}

				for _, bundledID := range recipe.BundledActivities {
					if char.LearnActivity(bundledID) {
						if log != nil {
							bundledActivity := entity.ActivityRegistry[bundledID]
							log.Add(char.ID, char.Name, "discovery",
								fmt.Sprintf("Discovered how to %s!", bundledActivity.Name))
						}
					}
				}

				return true
			}
		}
	}

	return false
}

// constructTriggerMatches checks if a discovery trigger matches a construct interaction.
// A trigger with empty ConstructKind never matches construct interactions (it's an item trigger).
// When ConstructMaterial is set on the trigger, the construct's material must also match.
func constructTriggerMatches(trigger entity.DiscoveryTrigger, action entity.ActionType, constructKind string, constructMaterial string) bool {
	if trigger.ConstructKind == "" {
		return false
	}
	if trigger.Action != action {
		return false
	}
	if trigger.ConstructKind != constructKind {
		return false
	}
	if trigger.ConstructMaterial != "" && trigger.ConstructMaterial != constructMaterial {
		return false
	}
	return true
}

// triggerMatches checks if a discovery trigger matches the current action and item
// item can be nil for actions like ActionDrink that don't involve items
func triggerMatches(trigger entity.DiscoveryTrigger, action entity.ActionType, item *entity.Item) bool {
	// Construct-only triggers never match item interactions
	if trigger.ConstructKind != "" {
		return false
	}

	// Action must match
	if trigger.Action != action {
		return false
	}

	// If trigger requires a specific item type, item must not be nil and must match
	if trigger.ItemType != "" {
		if item == nil || trigger.ItemType != item.ItemType {
			return false
		}
	}

	// Check edible requirement (only if item exists)
	if trigger.RequiresEdible {
		if item == nil || !item.IsEdible() {
			return false
		}
	}

	// Check plantable requirement
	if trigger.RequiresPlantable {
		if item == nil || !item.Plantable {
			return false
		}
	}

	// Check harvestable requirement (growing non-sprout plant)
	if trigger.RequiresHarvestable {
		if item == nil || item.Plant == nil || !item.Plant.IsGrowing || item.Plant.IsSprout {
			return false
		}
	}

	return true
}
