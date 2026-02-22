package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
)

// Consume handles a character eating an item
func Consume(char *entity.Character, item *entity.Item, gameMap *game.Map, log *ActionLog) {
	itemName := item.Description()
	oldHunger := char.Hunger

	// Update activity
	char.CurrentActivity = "Consuming " + itemName

	// Reduce hunger (per-item satiation tier)
	char.Hunger -= config.GetSatiationAmount(item.ItemType)
	if char.Hunger < 0 {
		char.Hunger = 0
	}

	// Set cooldown and boost mood when fully satisfied
	if char.Hunger == 0 {
		char.HungerCooldown = config.SatisfactionCooldown
		prevTier := char.MoodTier()
		char.Mood += config.MoodBoostOnConsumption
		if char.Mood > 100 {
			char.Mood = 100
		}
		// Log mood tier transition from boost
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Mood adjustment from preferences (C3)
	netPref := char.NetPreference(item)
	if netPref != 0 {
		oldMood := char.Mood
		prevTier := char.MoodTier()
		char.Mood += float64(netPref) * config.MoodPreferenceModifier
		if char.Mood > 100 {
			char.Mood = 100
		}
		if char.Mood < 0 {
			char.Mood = 0
		}
		// Log mood change from preference
		if log != nil {
			moodChange := "Improved Mood"
			if netPref < 0 {
				moodChange = "Worsened Mood"
			}
			log.Add(char.ID, char.Name, "mood",
				fmt.Sprintf("Eating %s %s (mood %d→%d)", itemName, moodChange, int(oldMood), int(char.Mood)))
		}
		// Log mood tier transition from preference
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Log consumption
	if log != nil {
		log.Add(char.ID, char.Name, "consumption",
			fmt.Sprintf("Consumed %s (hunger %d→%d)", itemName, int(oldHunger), int(char.Hunger)))
	}

	// Apply poison effect
	if item.IsPoisonous() {
		char.Poisoned = true
		char.PoisonTimer = config.PoisonDuration

		if log != nil {
			log.Add(char.ID, char.Name, "poison",
				fmt.Sprintf("Became poisoned! (duration: %ds)", int(config.PoisonDuration)))
		}

		// Learn that this item type is poisonous (includes dislike formation)
		knowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgePoisonous)
		LearnKnowledgeWithEffects(char, knowledge, log)
	}

	// Apply healing effect
	if item.IsHealing() {
		oldHealth := char.Health
		prevTier := char.HealthTier()
		char.Health += config.HealAmount
		if char.Health > 100 {
			char.Health = 100
		}

		// Only log and learn if health actually increased
		if char.Health > oldHealth {
			// Log health change (debug-only, filtered in view layer)
			if log != nil {
				log.Add(char.ID, char.Name, "health",
					fmt.Sprintf("Eating %s impacted health (%d→%d)", itemName, int(oldHealth), int(char.Health)))
			}

			// Learn that this item type is healing (only if we experienced healing)
			knowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgeHealing)
			LearnKnowledgeWithEffects(char, knowledge, log)
		}

		// Boost mood when fully healed
		if char.Health == 100 {
			char.Mood += config.MoodBoostOnConsumption
			if char.Mood > 100 {
				char.Mood = 100
			}
		}

		// Log health tier transition
		if log != nil && char.HealthTier() != prevTier {
			log.Add(char.ID, char.Name, "health", fmt.Sprintf("%s", char.HealthLevel()))
		}
	}

	// Try to form preference based on mood (C2)
	TryFormPreference(char, item, log)

	// Try to discover know-how from eating
	TryDiscoverKnowHow(char, entity.ActionConsume, item, log, GetDiscoveryChance(char))

	// Remove item from map
	gameMap.RemoveItem(item)

	// Gourd consumption creates a seed at character's position
	if item.ItemType == "gourd" {
		seed := entity.NewSeed(char.X, char.Y, "gourd", item.Color, item.Pattern, item.Texture)
		gameMap.AddItem(seed)
	}
}

// ConsumeFromInventory handles a character eating an item from their inventory
// Same effects as Consume, but clears inventory instead of removing from map
func ConsumeFromInventory(char *entity.Character, item *entity.Item, gameMap *game.Map, log *ActionLog) {
	itemName := item.Description()
	oldHunger := char.Hunger

	// Update activity
	char.CurrentActivity = "Consuming " + itemName

	// Reduce hunger (per-item satiation tier)
	char.Hunger -= config.GetSatiationAmount(item.ItemType)
	if char.Hunger < 0 {
		char.Hunger = 0
	}

	// Set cooldown and boost mood when fully satisfied
	if char.Hunger == 0 {
		char.HungerCooldown = config.SatisfactionCooldown
		prevTier := char.MoodTier()
		char.Mood += config.MoodBoostOnConsumption
		if char.Mood > 100 {
			char.Mood = 100
		}
		// Log mood tier transition from boost
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Mood adjustment from preferences (C3)
	netPref := char.NetPreference(item)
	if netPref != 0 {
		oldMood := char.Mood
		prevTier := char.MoodTier()
		char.Mood += float64(netPref) * config.MoodPreferenceModifier
		if char.Mood > 100 {
			char.Mood = 100
		}
		if char.Mood < 0 {
			char.Mood = 0
		}
		// Log mood change from preference
		if log != nil {
			moodChange := "Improved Mood"
			if netPref < 0 {
				moodChange = "Worsened Mood"
			}
			log.Add(char.ID, char.Name, "mood",
				fmt.Sprintf("Eating %s %s (mood %d→%d)", itemName, moodChange, int(oldMood), int(char.Mood)))
		}
		// Log mood tier transition from preference
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Log consumption
	if log != nil {
		log.Add(char.ID, char.Name, "consumption",
			fmt.Sprintf("Ate carried %s (hunger %d→%d)", itemName, int(oldHunger), int(char.Hunger)))
	}

	// Apply poison effect
	if item.IsPoisonous() {
		char.Poisoned = true
		char.PoisonTimer = config.PoisonDuration

		if log != nil {
			log.Add(char.ID, char.Name, "poison",
				fmt.Sprintf("Became poisoned! (duration: %ds)", int(config.PoisonDuration)))
		}

		// Learn that this item type is poisonous (includes dislike formation)
		knowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgePoisonous)
		LearnKnowledgeWithEffects(char, knowledge, log)
	}

	// Apply healing effect
	if item.IsHealing() {
		oldHealth := char.Health
		prevTier := char.HealthTier()
		char.Health += config.HealAmount
		if char.Health > 100 {
			char.Health = 100
		}

		// Only log and learn if health actually increased
		if char.Health > oldHealth {
			// Log health change (debug-only, filtered in view layer)
			if log != nil {
				log.Add(char.ID, char.Name, "health",
					fmt.Sprintf("Eating %s impacted health (%d→%d)", itemName, int(oldHealth), int(char.Health)))
			}

			// Learn that this item type is healing (only if we experienced healing)
			knowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgeHealing)
			LearnKnowledgeWithEffects(char, knowledge, log)
		}

		// Boost mood when fully healed
		if char.Health == 100 {
			char.Mood += config.MoodBoostOnConsumption
			if char.Mood > 100 {
				char.Mood = 100
			}
		}

		// Log health tier transition
		if log != nil && char.HealthTier() != prevTier {
			log.Add(char.ID, char.Name, "health", fmt.Sprintf("%s", char.HealthLevel()))
		}
	}

	// Try to form preference based on mood (C2)
	TryFormPreference(char, item, log)

	// Try to discover know-how from eating
	TryDiscoverKnowHow(char, entity.ActionConsume, item, log, GetDiscoveryChance(char))

	// Remove item from inventory (item consumed from inventory, not map)
	char.RemoveFromInventory(item)

	// Gourd consumption creates a seed at character's position
	if item.ItemType == "gourd" && gameMap != nil {
		seed := entity.NewSeed(char.X, char.Y, "gourd", item.Color, item.Pattern, item.Texture)
		gameMap.AddItem(seed)
	}
}

// Drink handles a character drinking from a water source
func Drink(char *entity.Character, log *ActionLog) {
	oldThirst := char.Thirst

	// Update activity
	char.CurrentActivity = "Drinking"

	// Reduce thirst
	char.Thirst -= config.DrinkThirstReduction
	if char.Thirst < 0 {
		char.Thirst = 0
	}

	// Set cooldown and boost mood when fully satisfied
	if char.Thirst == 0 {
		char.ThirstCooldown = config.SatisfactionCooldown
		prevTier := char.MoodTier()
		char.Mood += config.MoodBoostOnConsumption
		if char.Mood > 100 {
			char.Mood = 100
		}
		// Log mood tier transition from boost
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Log drinking
	if log != nil {
		log.Add(char.ID, char.Name, "thirst",
			fmt.Sprintf("Drank water (thirst %d→%d)", int(oldThirst), int(char.Thirst)))
	}
}

// StartSleep begins the sleeping state for a character
func StartSleep(char *entity.Character, atBed bool, log *ActionLog) {
	char.IsSleeping = true
	char.AtBed = atBed

	location := "on ground"
	if atBed {
		location = "in leaf pile"
	}
	char.CurrentActivity = "Sleeping (" + location + ")"

	if log != nil {
		if atBed {
			log.Add(char.ID, char.Name, "sleep",
				fmt.Sprintf("Fell asleep in leaf pile (energy: %d)", int(char.Energy)))
		} else if char.Energy <= 0 {
			log.Add(char.ID, char.Name, "sleep",
				fmt.Sprintf("Collapsed from exhaustion (energy: %d)", int(char.Energy)))
		} else {
			log.Add(char.ID, char.Name, "sleep",
				fmt.Sprintf("Fell asleep on ground (energy: %d)", int(char.Energy)))
		}
	}
}

// ConsumeFromVessel handles eating from a vessel's contents.
// Decrements the stack count and removes the stack when empty.
// The vessel remains in the character's inventory.
func ConsumeFromVessel(char *entity.Character, vessel *entity.Item, gameMap *game.Map, log *ActionLog) {
	if vessel.Container == nil || len(vessel.Container.Contents) == 0 {
		return // Nothing to eat
	}

	// Get the first stack
	stack := &vessel.Container.Contents[0]
	variety := stack.Variety
	varietyName := variety.Description()
	oldHunger := char.Hunger

	// Update activity
	char.CurrentActivity = "Consuming " + varietyName + " from vessel"

	// Reduce hunger (per-item satiation tier, uses variety's ItemType)
	char.Hunger -= config.GetSatiationAmount(variety.ItemType)
	if char.Hunger < 0 {
		char.Hunger = 0
	}

	// Set cooldown and boost mood when fully satisfied
	if char.Hunger == 0 {
		char.HungerCooldown = config.SatisfactionCooldown
		prevTier := char.MoodTier()
		char.Mood += config.MoodBoostOnConsumption
		if char.Mood > 100 {
			char.Mood = 100
		}
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Mood adjustment from preferences
	netPref := char.NetPreferenceForVariety(variety)
	if netPref != 0 {
		oldMood := char.Mood
		prevTier := char.MoodTier()
		char.Mood += float64(netPref) * config.MoodPreferenceModifier
		if char.Mood > 100 {
			char.Mood = 100
		}
		if char.Mood < 0 {
			char.Mood = 0
		}
		if log != nil {
			moodChange := "Improved Mood"
			if netPref < 0 {
				moodChange = "Worsened Mood"
			}
			log.Add(char.ID, char.Name, "mood",
				fmt.Sprintf("Eating %s %s (mood %d→%d)", varietyName, moodChange, int(oldMood), int(char.Mood)))
		}
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Log consumption
	if log != nil {
		log.Add(char.ID, char.Name, "consumption",
			fmt.Sprintf("Ate %s from vessel (hunger %d→%d, %d remaining)",
				varietyName, int(oldHunger), int(char.Hunger), stack.Count-1))
	}

	// Apply poison effect
	if variety.IsPoisonous() {
		char.Poisoned = true
		char.PoisonTimer = config.PoisonDuration

		if log != nil {
			log.Add(char.ID, char.Name, "poison",
				fmt.Sprintf("Became poisoned! (duration: %ds)", int(config.PoisonDuration)))
		}

		// Learn poison knowledge
		knowledge := entity.NewKnowledgeFromVariety(variety, entity.KnowledgePoisonous)
		LearnKnowledgeWithEffects(char, knowledge, log)
	}

	// Apply healing effect
	if variety.IsHealing() {
		oldHealth := char.Health
		prevTier := char.HealthTier()
		char.Health += config.HealAmount
		if char.Health > 100 {
			char.Health = 100
		}

		if char.Health > oldHealth {
			if log != nil {
				log.Add(char.ID, char.Name, "health",
					fmt.Sprintf("Eating %s impacted health (%d→%d)", varietyName, int(oldHealth), int(char.Health)))
			}

			// Learn healing knowledge
			knowledge := entity.NewKnowledgeFromVariety(variety, entity.KnowledgeHealing)
			LearnKnowledgeWithEffects(char, knowledge, log)
		}

		// Boost mood when fully healed
		if char.Health == 100 {
			char.Mood += config.MoodBoostOnConsumption
			if char.Mood > 100 {
				char.Mood = 100
			}
		}

		if log != nil && char.HealthTier() != prevTier {
			log.Add(char.ID, char.Name, "health", fmt.Sprintf("%s", char.HealthLevel()))
		}
	}

	// Decrement stack count
	stack.Count--

	// Remove stack if empty
	if stack.Count <= 0 {
		vessel.Container.Contents = vessel.Container.Contents[1:]
	}

	// Gourd consumption creates a seed at character's position
	if variety.ItemType == "gourd" && gameMap != nil {
		seed := entity.NewSeed(char.X, char.Y, "gourd", variety.Color, variety.Pattern, variety.Texture)
		gameMap.AddItem(seed)
	}

	// Note: We don't remove from inventory - vessel stays with character
}

