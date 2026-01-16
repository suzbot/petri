package entity

import (
	"testing"

	"petri/internal/types"
)

// =============================================================================
// Preference Matching
// =============================================================================

func TestPreference_Matches_ItemTypeOnly(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("berry", "")

	redBerry := &Item{ItemType: "berry", Color: types.ColorRed}
	blueBerry := &Item{ItemType: "berry", Color: types.ColorBlue}
	redMushroom := &Item{ItemType: "mushroom", Color: types.ColorRed}

	if !pref.Matches(redBerry) {
		t.Error("ItemType preference should match red berry")
	}
	if !pref.Matches(blueBerry) {
		t.Error("ItemType preference should match blue berry")
	}
	if pref.Matches(redMushroom) {
		t.Error("ItemType preference should not match mushroom")
	}
}

func TestPreference_Matches_ColorOnly(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("", types.ColorRed)

	redBerry := &Item{ItemType: "berry", Color: types.ColorRed}
	redMushroom := &Item{ItemType: "mushroom", Color: types.ColorRed}
	blueBerry := &Item{ItemType: "berry", Color: types.ColorBlue}

	if !pref.Matches(redBerry) {
		t.Error("Color preference should match red berry")
	}
	if !pref.Matches(redMushroom) {
		t.Error("Color preference should match red mushroom")
	}
	if pref.Matches(blueBerry) {
		t.Error("Color preference should not match blue berry")
	}
}

func TestPreference_Matches_Combo(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("berry", types.ColorRed)

	redBerry := &Item{ItemType: "berry", Color: types.ColorRed}
	blueBerry := &Item{ItemType: "berry", Color: types.ColorBlue}
	redMushroom := &Item{ItemType: "mushroom", Color: types.ColorRed}

	if !pref.Matches(redBerry) {
		t.Error("Combo preference should match red berry")
	}
	if pref.Matches(blueBerry) {
		t.Error("Combo preference should not match blue berry (wrong color)")
	}
	if pref.Matches(redMushroom) {
		t.Error("Combo preference should not match red mushroom (wrong type)")
	}
}

func TestPreference_Matches_EmptyPreference(t *testing.T) {
	t.Parallel()

	pref := Preference{Valence: 1} // No attributes set

	redBerry := &Item{ItemType: "berry", Color: types.ColorRed}

	if pref.Matches(redBerry) {
		t.Error("Empty preference should not match anything")
	}
}

func TestPreference_Matches_ValenceDoesNotAffectMatching(t *testing.T) {
	t.Parallel()

	positive := NewPositivePreference("berry", "")
	negative := NewNegativePreference("berry", "")

	redBerry := &Item{ItemType: "berry", Color: types.ColorRed}

	if !positive.Matches(redBerry) {
		t.Error("Positive preference should match")
	}
	if !negative.Matches(redBerry) {
		t.Error("Negative preference should also match (valence doesn't affect matching)")
	}
}

// =============================================================================
// Preference Description
// =============================================================================

func TestPreference_Description_ItemTypeOnly(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("berry", "")
	if pref.Description() != "berries" {
		t.Errorf("Expected 'berries', got '%s'", pref.Description())
	}

	pref2 := NewPositivePreference("mushroom", "")
	if pref2.Description() != "mushrooms" {
		t.Errorf("Expected 'mushrooms', got '%s'", pref2.Description())
	}
}

func TestPreference_Description_ColorOnly(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("", types.ColorRed)
	if pref.Description() != "red" {
		t.Errorf("Expected 'red', got '%s'", pref.Description())
	}
}

func TestPreference_Description_Combo(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("berry", types.ColorRed)
	if pref.Description() != "red berries" {
		t.Errorf("Expected 'red berries', got '%s'", pref.Description())
	}
}

func TestPreference_Description_Empty(t *testing.T) {
	t.Parallel()

	pref := Preference{Valence: 1}
	if pref.Description() != "" {
		t.Errorf("Expected empty string, got '%s'", pref.Description())
	}
}

// =============================================================================
// Preference Helpers
// =============================================================================

func TestPreference_IsPositive(t *testing.T) {
	t.Parallel()

	positive := NewPositivePreference("berry", "")
	negative := NewNegativePreference("berry", "")

	if !positive.IsPositive() {
		t.Error("Positive preference should return true for IsPositive()")
	}
	if negative.IsPositive() {
		t.Error("Negative preference should return false for IsPositive()")
	}
}

func TestNewPositivePreference(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("berry", types.ColorRed)

	if pref.Valence != 1 {
		t.Errorf("Expected Valence 1, got %d", pref.Valence)
	}
	if pref.ItemType != "berry" {
		t.Errorf("Expected ItemType 'berry', got '%s'", pref.ItemType)
	}
	if pref.Color != types.ColorRed {
		t.Errorf("Expected Color red, got '%s'", pref.Color)
	}
}

func TestNewNegativePreference(t *testing.T) {
	t.Parallel()

	pref := NewNegativePreference("mushroom", types.ColorWhite)

	if pref.Valence != -1 {
		t.Errorf("Expected Valence -1, got %d", pref.Valence)
	}
	if pref.ItemType != "mushroom" {
		t.Errorf("Expected ItemType 'mushroom', got '%s'", pref.ItemType)
	}
	if pref.Color != types.ColorWhite {
		t.Errorf("Expected Color white, got '%s'", pref.Color)
	}
}

// =============================================================================
// Exact Match
// =============================================================================

func TestPreference_ExactMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref1    Preference
		pref2    Preference
		expected bool
	}{
		{
			name:     "same ItemType only",
			pref1:    NewPositivePreference("berry", ""),
			pref2:    NewNegativePreference("berry", ""),
			expected: true, // same attributes, different valence
		},
		{
			name:     "same Color only",
			pref1:    NewPositivePreference("", types.ColorRed),
			pref2:    NewPositivePreference("", types.ColorRed),
			expected: true,
		},
		{
			name:     "same Combo",
			pref1:    NewPositivePreference("berry", types.ColorRed),
			pref2:    NewNegativePreference("berry", types.ColorRed),
			expected: true,
		},
		{
			name:     "different ItemType",
			pref1:    NewPositivePreference("berry", ""),
			pref2:    NewPositivePreference("mushroom", ""),
			expected: false,
		},
		{
			name:     "different Color",
			pref1:    NewPositivePreference("", types.ColorRed),
			pref2:    NewPositivePreference("", types.ColorBlue),
			expected: false,
		},
		{
			name:     "different specificity - ItemType vs Combo",
			pref1:    NewPositivePreference("berry", ""),
			pref2:    NewPositivePreference("berry", types.ColorRed),
			expected: false, // different attributes set
		},
		{
			name:     "different specificity - Color vs Combo",
			pref1:    NewPositivePreference("", types.ColorRed),
			pref2:    NewPositivePreference("berry", types.ColorRed),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pref1.ExactMatch(tt.pref2)
			if got != tt.expected {
				t.Errorf("ExactMatch(): got %v, want %v", got, tt.expected)
			}
			// Symmetry check
			got2 := tt.pref2.ExactMatch(tt.pref1)
			if got2 != tt.expected {
				t.Errorf("ExactMatch() (symmetric): got %v, want %v", got2, tt.expected)
			}
		})
	}
}

// =============================================================================
// Attribute Count
// =============================================================================

func TestPreference_AttributeCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref     Preference
		expected int
	}{
		{
			name:     "empty preference",
			pref:     Preference{Valence: 1},
			expected: 0,
		},
		{
			name:     "itemType only",
			pref:     NewPositivePreference("berry", ""),
			expected: 1,
		},
		{
			name:     "color only",
			pref:     NewPositivePreference("", types.ColorRed),
			expected: 1,
		},
		{
			name:     "combo (both attributes)",
			pref:     NewPositivePreference("berry", types.ColorRed),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pref.AttributeCount()
			if got != tt.expected {
				t.Errorf("AttributeCount(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Match Score
// =============================================================================

func TestPreference_MatchScore_NoMatch(t *testing.T) {
	t.Parallel()

	pref := NewPositivePreference("berry", "")
	mushroom := &Item{ItemType: "mushroom", Color: types.ColorRed}

	if pref.MatchScore(mushroom) != 0 {
		t.Error("MatchScore should return 0 for non-matching item")
	}
}

func TestPreference_MatchScore_SingleAttribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref     Preference
		item     *Item
		expected int
	}{
		{
			name:     "positive itemType match",
			pref:     NewPositivePreference("berry", ""),
			item:     &Item{ItemType: "berry", Color: types.ColorRed},
			expected: 1, // valence(+1) × attrCount(1)
		},
		{
			name:     "positive color match",
			pref:     NewPositivePreference("", types.ColorRed),
			item:     &Item{ItemType: "mushroom", Color: types.ColorRed},
			expected: 1, // valence(+1) × attrCount(1)
		},
		{
			name:     "negative itemType match",
			pref:     NewNegativePreference("berry", ""),
			item:     &Item{ItemType: "berry", Color: types.ColorRed},
			expected: -1, // valence(-1) × attrCount(1)
		},
		{
			name:     "negative color match",
			pref:     NewNegativePreference("", types.ColorRed),
			item:     &Item{ItemType: "mushroom", Color: types.ColorRed},
			expected: -1, // valence(-1) × attrCount(1)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pref.MatchScore(tt.item)
			if got != tt.expected {
				t.Errorf("MatchScore(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPreference_MatchScore_ComboAttribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref     Preference
		item     *Item
		expected int
	}{
		{
			name:     "positive combo match",
			pref:     NewPositivePreference("berry", types.ColorRed),
			item:     &Item{ItemType: "berry", Color: types.ColorRed},
			expected: 2, // valence(+1) × attrCount(2)
		},
		{
			name:     "negative combo match",
			pref:     NewNegativePreference("berry", types.ColorRed),
			item:     &Item{ItemType: "berry", Color: types.ColorRed},
			expected: -2, // valence(-1) × attrCount(2)
		},
		{
			name:     "combo partial match (wrong color)",
			pref:     NewPositivePreference("berry", types.ColorRed),
			item:     &Item{ItemType: "berry", Color: types.ColorBlue},
			expected: 0, // no match - must match ALL attributes
		},
		{
			name:     "combo partial match (wrong type)",
			pref:     NewPositivePreference("berry", types.ColorRed),
			item:     &Item{ItemType: "mushroom", Color: types.ColorRed},
			expected: 0, // no match - must match ALL attributes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pref.MatchScore(tt.item)
			if got != tt.expected {
				t.Errorf("MatchScore(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Pattern and Texture Matching
// =============================================================================

func TestPreference_Matches_PatternOnly(t *testing.T) {
	t.Parallel()

	pref := Preference{Valence: 1, Pattern: types.PatternSpotted}

	spottedMushroom := &Item{ItemType: "mushroom", Color: types.ColorBrown, Pattern: types.PatternSpotted}
	plainMushroom := &Item{ItemType: "mushroom", Color: types.ColorBrown, Pattern: types.PatternNone}
	redBerry := &Item{ItemType: "berry", Color: types.ColorRed} // No pattern

	if !pref.Matches(spottedMushroom) {
		t.Error("Pattern preference should match spotted mushroom")
	}
	if pref.Matches(plainMushroom) {
		t.Error("Pattern preference should not match plain mushroom")
	}
	if pref.Matches(redBerry) {
		t.Error("Pattern preference should not match berry (no pattern)")
	}
}

func TestPreference_Matches_TextureOnly(t *testing.T) {
	t.Parallel()

	pref := Preference{Valence: 1, Texture: types.TextureSlimy}

	slimyMushroom := &Item{ItemType: "mushroom", Color: types.ColorBrown, Texture: types.TextureSlimy}
	normalMushroom := &Item{ItemType: "mushroom", Color: types.ColorBrown, Texture: types.TextureNone}
	redBerry := &Item{ItemType: "berry", Color: types.ColorRed} // No texture

	if !pref.Matches(slimyMushroom) {
		t.Error("Texture preference should match slimy mushroom")
	}
	if pref.Matches(normalMushroom) {
		t.Error("Texture preference should not match non-slimy mushroom")
	}
	if pref.Matches(redBerry) {
		t.Error("Texture preference should not match berry (no texture)")
	}
}

func TestPreference_Matches_MushroomCombo(t *testing.T) {
	t.Parallel()

	// Preference for spotted brown mushrooms
	pref := Preference{
		Valence:  1,
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
	}

	spottedBrown := &Item{ItemType: "mushroom", Color: types.ColorBrown, Pattern: types.PatternSpotted}
	plainBrown := &Item{ItemType: "mushroom", Color: types.ColorBrown, Pattern: types.PatternNone}
	spottedWhite := &Item{ItemType: "mushroom", Color: types.ColorWhite, Pattern: types.PatternSpotted}

	if !pref.Matches(spottedBrown) {
		t.Error("Should match spotted brown mushroom")
	}
	if pref.Matches(plainBrown) {
		t.Error("Should not match plain brown mushroom (wrong pattern)")
	}
	if pref.Matches(spottedWhite) {
		t.Error("Should not match spotted white mushroom (wrong color)")
	}
}

func TestPreference_AttributeCount_WithPatternTexture(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref     Preference
		expected int
	}{
		{
			name:     "pattern only",
			pref:     Preference{Valence: 1, Pattern: types.PatternSpotted},
			expected: 1,
		},
		{
			name:     "texture only",
			pref:     Preference{Valence: 1, Texture: types.TextureSlimy},
			expected: 1,
		},
		{
			name:     "pattern + texture",
			pref:     Preference{Valence: 1, Pattern: types.PatternSpotted, Texture: types.TextureSlimy},
			expected: 2,
		},
		{
			name:     "all four attributes",
			pref:     Preference{Valence: 1, ItemType: "mushroom", Color: types.ColorBrown, Pattern: types.PatternSpotted, Texture: types.TextureSlimy},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pref.AttributeCount()
			if got != tt.expected {
				t.Errorf("AttributeCount(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPreference_ExactMatch_WithPatternTexture(t *testing.T) {
	t.Parallel()

	pref1 := Preference{Valence: 1, ItemType: "mushroom", Pattern: types.PatternSpotted}
	pref2 := Preference{Valence: -1, ItemType: "mushroom", Pattern: types.PatternSpotted}
	pref3 := Preference{Valence: 1, ItemType: "mushroom", Pattern: types.PatternNone}

	if !pref1.ExactMatch(pref2) {
		t.Error("Same attributes with different valence should be exact match")
	}
	if pref1.ExactMatch(pref3) {
		t.Error("Different pattern should not be exact match")
	}
}

// =============================================================================
// NewFullPreferenceFromItem
// =============================================================================

func TestNewFullPreferenceFromItem_Berry(t *testing.T) {
	t.Parallel()

	item := &Item{ItemType: "berry", Color: types.ColorRed}
	pref := NewFullPreferenceFromItem(item, -1)

	if pref.Valence != -1 {
		t.Errorf("Expected Valence -1, got %d", pref.Valence)
	}
	if pref.ItemType != "berry" {
		t.Errorf("Expected ItemType 'berry', got '%s'", pref.ItemType)
	}
	if pref.Color != types.ColorRed {
		t.Errorf("Expected Color red, got '%s'", pref.Color)
	}
	if pref.Pattern != "" {
		t.Errorf("Expected empty Pattern, got '%s'", pref.Pattern)
	}
	if pref.Texture != "" {
		t.Errorf("Expected empty Texture, got '%s'", pref.Texture)
	}
}

func TestNewFullPreferenceFromItem_MushroomWithAllAttributes(t *testing.T) {
	t.Parallel()

	item := &Item{
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
	}
	pref := NewFullPreferenceFromItem(item, -1)

	if pref.Valence != -1 {
		t.Errorf("Expected Valence -1, got %d", pref.Valence)
	}
	if pref.ItemType != "mushroom" {
		t.Errorf("Expected ItemType 'mushroom', got '%s'", pref.ItemType)
	}
	if pref.Color != types.ColorBrown {
		t.Errorf("Expected Color brown, got '%s'", pref.Color)
	}
	if pref.Pattern != types.PatternSpotted {
		t.Errorf("Expected Pattern spotted, got '%s'", pref.Pattern)
	}
	if pref.Texture != types.TextureSlimy {
		t.Errorf("Expected Texture slimy, got '%s'", pref.Texture)
	}
}

func TestNewFullPreferenceFromItem_MushroomPartialAttributes(t *testing.T) {
	t.Parallel()

	// Mushroom with only pattern, no texture
	item := &Item{
		ItemType: "mushroom",
		Color:    types.ColorWhite,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureNone,
	}
	pref := NewFullPreferenceFromItem(item, 1)

	if pref.Valence != 1 {
		t.Errorf("Expected Valence 1, got %d", pref.Valence)
	}
	if pref.ItemType != "mushroom" {
		t.Errorf("Expected ItemType 'mushroom', got '%s'", pref.ItemType)
	}
	if pref.Color != types.ColorWhite {
		t.Errorf("Expected Color white, got '%s'", pref.Color)
	}
	if pref.Pattern != types.PatternSpotted {
		t.Errorf("Expected Pattern spotted, got '%s'", pref.Pattern)
	}
	// TextureNone should be copied as-is (empty string equivalent)
	if pref.Texture != types.TextureNone {
		t.Errorf("Expected Texture none, got '%s'", pref.Texture)
	}
}

// =============================================================================
// Pluralization
// =============================================================================

func TestPluralize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"berry", "berries"},
		{"mushroom", "mushrooms"},
		{"flower", "flowers"},
		{"unknown", "unknowns"},
	}

	for _, tc := range tests {
		result := pluralize(tc.input)
		if result != tc.expected {
			t.Errorf("pluralize(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}
