package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
)

// CompleteLook handles completion of a looking action
// This provides an opportunity for preference formation and mood impact
func CompleteLook(char *entity.Character, item *entity.Item, log *ActionLog) {
	if item == nil {
		return
	}

	itemName := item.Description()

	// Log completion
	if log != nil {
		log.Add(char.ID, char.Name, "activity", "Looked at "+itemName)
	}

	// Remember last looked item (to avoid looking at same item twice in a row)
	ipos := item.Pos()
	char.LastLookedX = ipos.X
	char.LastLookedY = ipos.Y
	char.HasLastLooked = true

	// Mood adjustment from preferences (same as eating)
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
				fmt.Sprintf("Looking at %s %s (mood %d→%d)", itemName, moodChange, int(oldMood), int(char.Mood)))
		}
		// Log mood tier transition from preference
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Try to form a preference based on mood
	TryFormPreference(char, item, log)

	// Try to discover know-how from looking at edible items
	TryDiscoverKnowHow(char, entity.ActionLook, item, log, GetDiscoveryChance(char))
}

// CompleteLookAtConstruct handles completion of looking at a construct.
// Adjusts mood from existing preferences and attempts preference formation.
// No discovery triggers (DD-36 — deferred to Step 8).
func CompleteLookAtConstruct(char *entity.Character, construct *entity.Construct, log *ActionLog) {
	if construct == nil {
		return
	}

	name := construct.DisplayName()

	if log != nil {
		log.Add(char.ID, char.Name, "activity", "Looked at "+name)
	}

	// Remember last looked position (to avoid looking at same construct twice in a row)
	cpos := construct.Pos()
	char.LastLookedX = cpos.X
	char.LastLookedY = cpos.Y
	char.HasLastLooked = true

	// Mood adjustment from existing preferences
	netPref := char.NetPreferenceForConstruct(construct)
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
				fmt.Sprintf("Looking at %s %s (mood %d→%d)", name, moodChange, int(oldMood), int(char.Mood)))
		}
		if log != nil && char.MoodTier() != prevTier {
			log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
		}
	}

	// Try to form a preference based on mood
	TryFormConstructPreference(char, construct, log)
}
