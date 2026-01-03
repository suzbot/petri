package system

import (
	"petri/internal/config"
	"petri/internal/entity"
)

// CompleteLook handles completion of a looking action
// This provides an opportunity for preference formation
func CompleteLook(char *entity.Character, item *entity.Item, log *ActionLog) {
	if item == nil {
		return
	}

	// Log completion
	if log != nil {
		log.Add(char.ID, char.Name, "activity", "Looked at "+item.Description())
	}

	// Set cooldown and remember last looked item
	char.LookCooldown = config.LookCooldown
	ix, iy := item.Position()
	char.LastLookedX = ix
	char.LastLookedY = iy
	char.HasLastLooked = true

	// Try to form a preference based on mood
	TryFormPreference(char, item, log)
}
