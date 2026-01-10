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

	// Reduce hunger
	char.Hunger -= config.FoodHungerReduction
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
	if item.Poisonous {
		char.Poisoned = true
		char.PoisonTimer = config.PoisonDuration

		if log != nil {
			log.Add(char.ID, char.Name, "poison",
				fmt.Sprintf("Became poisoned! (duration: %ds)", int(config.PoisonDuration)))
		}

		// Learn that this item type is poisonous
		knowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgePoisonous)
		if char.LearnKnowledge(knowledge) && log != nil {
			log.Add(char.ID, char.Name, "learning", "Learned something!")
		}
	}

	// Apply healing effect
	if item.Healing {
		oldHealth := char.Health
		prevTier := char.HealthTier()
		char.Health += config.HealAmount
		if char.Health > 100 {
			char.Health = 100
		}

		// Log health change (debug-only, filtered in view layer)
		if log != nil {
			log.Add(char.ID, char.Name, "health",
				fmt.Sprintf("Eating %s impacted health (%d→%d)", itemName, int(oldHealth), int(char.Health)))
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

		// Learn that this item type is healing
		knowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgeHealing)
		if char.LearnKnowledge(knowledge) && log != nil {
			log.Add(char.ID, char.Name, "learning", "Learned something!")
		}
	}

	// Try to form preference based on mood (C2)
	TryFormPreference(char, item, log)

	// Remove item from map
	gameMap.RemoveItem(item)
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
