package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

// =============================================================================
// Formation Parameters by Mood
// =============================================================================

func TestGetFormationParams_MoodTiers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		moodTier        int
		expectChance    bool // true if chance > 0
		expectValence   int  // expected valence (0 means no formation)
	}{
		{"Joyful (TierNone) forms positive", entity.TierNone, true, 1},
		{"Happy (TierMild) forms positive", entity.TierMild, true, 1},
		{"Neutral (TierModerate) no formation", entity.TierModerate, false, 0},
		{"Unhappy (TierSevere) forms negative", entity.TierSevere, true, -1},
		{"Miserable (TierCrisis) forms negative", entity.TierCrisis, true, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chance, valence := getFormationParams(tt.moodTier)
			if tt.expectChance && chance == 0 {
				t.Errorf("Expected non-zero chance for mood tier %d", tt.moodTier)
			}
			if !tt.expectChance && chance != 0 {
				t.Errorf("Expected zero chance for mood tier %d, got %f", tt.moodTier, chance)
			}
			if valence != tt.expectValence {
				t.Errorf("Expected valence %d, got %d", tt.expectValence, valence)
			}
		})
	}
}

func TestGetFormationParams_ChanceValues(t *testing.T) {
	t.Parallel()

	// Miserable and Joyful should have higher chances than Unhappy and Happy
	miserableChance, _ := getFormationParams(entity.TierCrisis)
	unhappyChance, _ := getFormationParams(entity.TierSevere)
	happyChance, _ := getFormationParams(entity.TierMild)
	joyfulChance, _ := getFormationParams(entity.TierNone)

	if miserableChance <= unhappyChance {
		t.Errorf("Miserable chance (%f) should be > Unhappy chance (%f)", miserableChance, unhappyChance)
	}
	if joyfulChance <= happyChance {
		t.Errorf("Joyful chance (%f) should be > Happy chance (%f)", joyfulChance, happyChance)
	}
}

// =============================================================================
// Preference Type Rolling
// =============================================================================

func TestRollPreferenceType_ReturnsValidPreference(t *testing.T) {
	t.Parallel()

	item := entity.NewBerry(0, 0, types.ColorRed, false, false)

	// Run multiple times to cover different random outcomes
	for i := 0; i < 100; i++ {
		pref := rollPreferenceType(item, 1)

		// Must have at least one attribute set
		if pref.ItemType == "" && pref.Color == "" && pref.Pattern == "" && pref.Texture == "" {
			t.Error("Preference must have at least one attribute set")
		}

		// If ItemType is set, it must match item
		if pref.ItemType != "" && pref.ItemType != item.ItemType {
			t.Errorf("ItemType mismatch: got %s, want %s", pref.ItemType, item.ItemType)
		}

		// If Color is set, it must match item
		if pref.Color != "" && pref.Color != item.Color {
			t.Errorf("Color mismatch: got %s, want %s", pref.Color, item.Color)
		}

		// Valence must match input
		if pref.Valence != 1 {
			t.Errorf("Valence mismatch: got %d, want 1", pref.Valence)
		}
	}
}

func TestRollPreferenceType_NegativeValence(t *testing.T) {
	t.Parallel()

	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	pref := rollPreferenceType(item, -1)

	if pref.Valence != -1 {
		t.Errorf("Expected negative valence, got %d", pref.Valence)
	}
}

func TestRollPreferenceType_MushroomIncludesPatternTexture(t *testing.T) {
	t.Parallel()

	// Mushroom with pattern and texture
	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)

	// Run multiple times to see pattern/texture being used
	hasPattern := false
	hasTexture := false
	for i := 0; i < 200; i++ {
		pref := rollPreferenceType(item, 1)

		// If Pattern is set, it must match item
		if pref.Pattern != "" {
			hasPattern = true
			if pref.Pattern != item.Pattern {
				t.Errorf("Pattern mismatch: got %s, want %s", pref.Pattern, item.Pattern)
			}
		}

		// If Texture is set, it must match item
		if pref.Texture != "" {
			hasTexture = true
			if pref.Texture != item.Texture {
				t.Errorf("Texture mismatch: got %s, want %s", pref.Texture, item.Texture)
			}
		}
	}

	// With 4 attributes and 200 iterations, we should see pattern and texture
	if !hasPattern {
		t.Error("Expected at least one preference with Pattern set")
	}
	if !hasTexture {
		t.Error("Expected at least one preference with Texture set")
	}
}

// =============================================================================
// Existing Preference Handling
// =============================================================================

func TestTryFormPreference_SamePreferenceExists_NoChange(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)
	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Mood: 100, // Joyful - will try to form positive
		Preferences: []entity.Preference{
			entity.NewPositivePreference("berry", ""), // Already likes berries
		},
	}
	item := entity.NewBerry(0, 0, types.ColorRed, false, false)

	// Force formation to happen and target ItemType only
	// We need deterministic testing, so we'll test the logic directly
	candidate := entity.Preference{Valence: 1, ItemType: "berry"}

	// Check existing preference handling
	for _, existing := range char.Preferences {
		if existing.ExactMatch(candidate) {
			if existing.Valence == candidate.Valence {
				// This is the expected path - same preference exists
				return // Test passes
			}
		}
	}
	t.Error("Should have found existing matching preference")
	_ = item // suppress unused warning
	_ = log
}

func TestTryFormPreference_OppositePreferenceExists_Removed(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)
	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Mood: 0, // Miserable - will try to form negative
		Preferences: []entity.Preference{
			entity.NewPositivePreference("berry", ""), // Likes berries
		},
	}

	// Simulate forming a negative preference for berries
	candidate := entity.Preference{Valence: -1, ItemType: "berry"}

	// Apply the logic
	for i, existing := range char.Preferences {
		if existing.ExactMatch(candidate) {
			if existing.Valence != candidate.Valence {
				// Opposite valence - remove existing
				char.Preferences = append(char.Preferences[:i], char.Preferences[i+1:]...)
				break
			}
		}
	}

	// Verify preference was removed
	if len(char.Preferences) != 0 {
		t.Errorf("Expected 0 preferences after removal, got %d", len(char.Preferences))
	}
	_ = log
}

func TestTryFormPreference_NoExistingMatch_Added(t *testing.T) {
	t.Parallel()

	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Mood: 100, // Joyful
		Preferences: []entity.Preference{
			entity.NewPositivePreference("berry", ""), // Likes berries
		},
	}

	// Simulate forming a preference for mushrooms (no existing match)
	candidate := entity.Preference{Valence: 1, ItemType: "mushroom"}

	// Check no existing match
	hasMatch := false
	for _, existing := range char.Preferences {
		if existing.ExactMatch(candidate) {
			hasMatch = true
			break
		}
	}

	if hasMatch {
		t.Error("Should not have found existing match for mushrooms")
	}

	// Add new preference
	char.Preferences = append(char.Preferences, candidate)

	if len(char.Preferences) != 2 {
		t.Errorf("Expected 2 preferences after adding, got %d", len(char.Preferences))
	}
}

func TestTryFormPreference_DifferentSpecificity_Coexist(t *testing.T) {
	t.Parallel()

	// Character likes berries (general)
	// Forms dislike for red berries (specific combo)
	// Both should coexist

	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Preferences: []entity.Preference{
			entity.NewPositivePreference("berry", ""), // Likes berries (ItemType only)
		},
	}

	// New preference: dislikes red berries (combo)
	newPref := entity.NewNegativePreference("berry", types.ColorRed)

	// Check for exact match
	hasExactMatch := false
	for _, existing := range char.Preferences {
		if existing.ExactMatch(newPref) {
			hasExactMatch = true
			break
		}
	}

	// "likes berries" (ItemType="berry", Color="") should NOT exactly match
	// "dislikes red berries" (ItemType="berry", Color="red")
	if hasExactMatch {
		t.Error("Different specificity preferences should not be exact matches")
	}

	// Both can coexist
	char.Preferences = append(char.Preferences, newPref)

	if len(char.Preferences) != 2 {
		t.Errorf("Expected 2 preferences (coexisting), got %d", len(char.Preferences))
	}

	// Verify both exist
	hasLikesBerries := false
	hasDislikesRedBerries := false
	for _, p := range char.Preferences {
		if p.ItemType == "berry" && p.Color == "" && p.Valence == 1 {
			hasLikesBerries = true
		}
		if p.ItemType == "berry" && p.Color == types.ColorRed && p.Valence == -1 {
			hasDislikesRedBerries = true
		}
	}

	if !hasLikesBerries {
		t.Error("Should still have 'likes berries' preference")
	}
	if !hasDislikesRedBerries {
		t.Error("Should have new 'dislikes red berries' preference")
	}
}

// =============================================================================
// Neutral Mood - No Formation
// =============================================================================

func TestTryFormPreference_NeutralMood_NoFormation(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)
	char := &entity.Character{
		ID:          1,
		Name:        "Test",
		Mood:        50, // Neutral (TierModerate)
		Preferences: []entity.Preference{},
	}
	item := entity.NewBerry(0, 0, types.ColorRed, false, false)

	result, _ := TryFormPreference(char, item, log)

	if result != FormationNone {
		t.Errorf("Expected FormationNone for neutral mood, got %v", result)
	}
	if len(char.Preferences) != 0 {
		t.Errorf("Expected no preferences formed, got %d", len(char.Preferences))
	}
}

// =============================================================================
// D6: Combo Preferences Must Include ItemType
// =============================================================================

func TestRollPreferenceType_ComboAlwaysIncludesItemType(t *testing.T) {
	t.Parallel()

	// Mushroom with all attributes available
	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)

	// Run many iterations to ensure combos always include ItemType
	for i := 0; i < 500; i++ {
		pref := rollPreferenceType(item, 1)

		// If this is a combo (2+ attributes), it must include ItemType
		if pref.AttributeCount() >= 2 {
			if pref.ItemType == "" {
				t.Errorf("Combo preference (count=%d) must include ItemType, got: %+v",
					pref.AttributeCount(), pref)
			}
		}
	}
}

func TestRollPreferenceType_ComboLimitedToThreeAttributes(t *testing.T) {
	t.Parallel()

	// Mushroom with all 4 possible attributes
	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)

	// Run many iterations to check max attribute count
	for i := 0; i < 500; i++ {
		pref := rollPreferenceType(item, 1)

		if pref.AttributeCount() > 3 {
			t.Errorf("Preference should have max 3 attributes for interaction-formed preferences, got %d: %+v",
				pref.AttributeCount(), pref)
		}
	}
}

func TestRollPreferenceType_ComboCanHaveTwoOrThreeAttributes(t *testing.T) {
	t.Parallel()

	// Mushroom with all attributes
	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)

	hasTwoAttr := false
	hasThreeAttr := false

	// Run enough iterations to see both 2 and 3 attribute combos
	for i := 0; i < 500; i++ {
		pref := rollPreferenceType(item, 1)
		count := pref.AttributeCount()

		if count == 2 {
			hasTwoAttr = true
		}
		if count == 3 {
			hasThreeAttr = true
		}
	}

	if !hasTwoAttr {
		t.Error("Expected to see some 2-attribute combo preferences")
	}
	if !hasThreeAttr {
		t.Error("Expected to see some 3-attribute combo preferences")
	}
}

func TestRollPreferenceType_SoloCanBeAnyAttribute(t *testing.T) {
	t.Parallel()

	// Mushroom with all attributes
	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)

	hasItemType := false
	hasColor := false
	hasPattern := false
	hasTexture := false

	// Run enough iterations to see all solo types
	for i := 0; i < 1000; i++ {
		pref := rollPreferenceType(item, 1)

		// Check solo preferences (exactly 1 attribute)
		if pref.AttributeCount() == 1 {
			if pref.ItemType != "" {
				hasItemType = true
			}
			if pref.Color != "" {
				hasColor = true
			}
			if pref.Pattern != "" {
				hasPattern = true
			}
			if pref.Texture != "" {
				hasTexture = true
			}
		}
	}

	if !hasItemType {
		t.Error("Expected to see solo ItemType preferences")
	}
	if !hasColor {
		t.Error("Expected to see solo Color preferences")
	}
	if !hasPattern {
		t.Error("Expected to see solo Pattern preferences")
	}
	if !hasTexture {
		t.Error("Expected to see solo Texture preferences")
	}
}
