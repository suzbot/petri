package system

import (
	"math/rand"

	"petri/internal/config"
	"petri/internal/entity"
)

// PreferenceFormationResult represents the outcome of a preference formation attempt
type PreferenceFormationResult int

const (
	FormationNone     PreferenceFormationResult = iota // No formation (wrong mood or failed roll)
	FormationNew                                       // New preference formed
	FormationRemoved                                   // Existing opposite preference removed
	FormationNoChange                                  // Same preference already exists
)

// TryFormPreference attempts to form a preference based on the character's mood
// while interacting with an item. Returns the result and the preference involved (if any).
func TryFormPreference(char *entity.Character, item *entity.Item, log *ActionLog) (PreferenceFormationResult, *entity.Preference) {
	// Determine formation chance and valence based on mood tier
	chance, valence := getFormationParams(char.MoodTier())
	if chance == 0 {
		return FormationNone, nil
	}

	// Roll for formation
	if rand.Float64() >= chance {
		return FormationNone, nil
	}

	// Determine preference type (ItemType, Color, or Combo)
	candidate := rollPreferenceType(item, valence)

	// Check for existing preference with exact match
	for i, existing := range char.Preferences {
		if existing.ExactMatch(candidate) {
			if existing.Valence == candidate.Valence {
				// Same preference already exists - no change
				return FormationNoChange, &candidate
			}
			// Opposite valence - remove existing preference
			char.Preferences = append(char.Preferences[:i], char.Preferences[i+1:]...)
			logPreferenceRemoved(char, existing, log)
			return FormationRemoved, &existing
		}
	}

	// No existing match - add new preference
	char.Preferences = append(char.Preferences, candidate)
	logPreferenceFormed(char, candidate, log)
	return FormationNew, &candidate
}

// getFormationParams returns the formation chance and valence for a mood tier.
// Negative moods form negative preferences, positive moods form positive preferences.
func getFormationParams(moodTier int) (chance float64, valence int) {
	switch moodTier {
	case entity.TierCrisis: // Miserable
		return config.PrefFormationChanceMiserable, -1
	case entity.TierSevere: // Unhappy
		return config.PrefFormationChanceUnhappy, -1
	case entity.TierMild: // Happy
		return config.PrefFormationChanceHappy, 1
	case entity.TierNone: // Joyful
		return config.PrefFormationChanceJoyful, 1
	default: // Neutral (TierModerate)
		return 0, 0
	}
}

// rollPreferenceType randomly selects which type of preference to form
// based on configured weights: single attribute or combo (2+ attributes).
// Solo: any single attribute. Combo: ItemType + 1-2 other attributes (max 3 total).
func rollPreferenceType(item *entity.Item, valence int) entity.Preference {
	roll := rand.Float64()

	// Build list of available attributes for this item
	attrs := collectItemAttributes(item)

	if roll < config.PrefFormationWeightSingle {
		// Single attribute - pick one randomly from all available
		attr := attrs[rand.Intn(len(attrs))]
		return buildPreference(valence, []string{attr}, item)
	}

	// Combo - always include ItemType + 1-2 extra attributes
	extras := collectExtraAttributes(item) // excludes itemType

	if len(extras) == 0 {
		// Fallback: only itemType available, return solo itemType
		return buildPreference(valence, []string{"itemType"}, item)
	}

	// Determine how many extra attributes to include (1 or 2)
	numExtras := 1
	if len(extras) >= 2 && rand.Float64() < 0.5 {
		numExtras = 2
	}

	// Shuffle extras and pick first numExtras
	rand.Shuffle(len(extras), func(i, j int) {
		extras[i], extras[j] = extras[j], extras[i]
	})

	// Build combo: itemType + selected extras
	selected := []string{"itemType"}
	selected = append(selected, extras[:numExtras]...)

	return buildPreference(valence, selected, item)
}

// collectItemAttributes returns the list of available descriptive attributes for an item.
func collectItemAttributes(item *entity.Item) []string {
	attrs := []string{"itemType", "color"}

	// Mushrooms have additional attributes
	if item.ItemType == "mushroom" {
		if item.Pattern != "" {
			attrs = append(attrs, "pattern")
		}
		if item.Texture != "" {
			attrs = append(attrs, "texture")
		}
	}

	return attrs
}

// collectExtraAttributes returns attributes excluding itemType (for combo formation).
func collectExtraAttributes(item *entity.Item) []string {
	attrs := []string{"color"}

	// Mushrooms have additional attributes
	if item.ItemType == "mushroom" {
		if item.Pattern != "" {
			attrs = append(attrs, "pattern")
		}
		if item.Texture != "" {
			attrs = append(attrs, "texture")
		}
	}

	return attrs
}

// buildPreference creates a preference using the specified attributes from the item.
func buildPreference(valence int, attrs []string, item *entity.Item) entity.Preference {
	pref := entity.Preference{Valence: valence}

	for _, attr := range attrs {
		switch attr {
		case "itemType":
			pref.ItemType = item.ItemType
		case "color":
			pref.Color = item.Color
		case "pattern":
			pref.Pattern = item.Pattern
		case "texture":
			pref.Texture = item.Texture
		}
	}

	return pref
}

// logPreferenceFormed logs a new preference formation
func logPreferenceFormed(char *entity.Character, pref entity.Preference, log *ActionLog) {
	if log == nil {
		return
	}
	verb := "Likes"
	if pref.Valence < 0 {
		verb = "Dislikes"
	}
	log.Add(char.ID, char.Name, "preference", "New Opinion: "+verb+" "+pref.Description())
}

// logPreferenceRemoved logs when an existing preference is removed
func logPreferenceRemoved(char *entity.Character, pref entity.Preference, log *ActionLog) {
	if log == nil {
		return
	}
	verb := "No longer likes"
	if pref.Valence < 0 {
		verb = "No longer dislikes"
	}
	log.Add(char.ID, char.Name, "preference", verb+" "+pref.Description())
}
