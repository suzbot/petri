package ui

import (
	"testing"
)

// =============================================================================
// Character Creation State Tests
// =============================================================================

func TestNewCharacterCreationState_DefaultValues(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()

	// Should have 4 characters
	if len(state.Characters) != 4 {
		t.Errorf("Expected 4 characters, got %d", len(state.Characters))
	}

	// Check default names
	expectedNames := []string{"Len", "Macca", "Hari", "Starr"}
	for i, expected := range expectedNames {
		if state.Characters[i].Name != expected {
			t.Errorf("Character %d name: expected %q, got %q", i, expected, state.Characters[i].Name)
		}
	}

	// First character should be selected
	if state.SelectedChar != 0 {
		t.Errorf("Expected SelectedChar 0, got %d", state.SelectedChar)
	}

	// Name field should be focused initially
	if state.SelectedField != FieldName {
		t.Errorf("Expected SelectedField FieldName, got %d", state.SelectedField)
	}
}

func TestNewCharacterCreationState_RandomFoodAndColor(t *testing.T) {
	t.Parallel()

	// Build valid foods map from foodOptions
	validFoods := make(map[string]bool)
	for _, f := range foodOptions {
		validFoods[f] = true
	}

	// Build valid colors map from colorOptions
	validColors := make(map[string]bool)
	for _, c := range colorOptions {
		validColors[c] = true
	}

	// Run multiple times to check randomization produces valid values
	for i := 0; i < 10; i++ {
		state := NewCharacterCreationState()

		for j, char := range state.Characters {
			// Food should be valid (from dynamically generated foodOptions)
			if !validFoods[char.Food] {
				t.Errorf("Character %d has invalid food: %q", j, char.Food)
			}

			// Color should be valid (from dynamically generated colorOptions)
			if !validColors[char.Color] {
				t.Errorf("Character %d has invalid color: %q", j, char.Color)
			}
		}
	}
}

// =============================================================================
// Navigation Tests
// =============================================================================

func TestCharacterCreationState_NavigateCharacterRight(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0

	state.NavigateCharacter(1)
	if state.SelectedChar != 1 {
		t.Errorf("Expected SelectedChar 1, got %d", state.SelectedChar)
	}
}

func TestCharacterCreationState_NavigateCharacterWrapsRight(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 3

	state.NavigateCharacter(1)
	if state.SelectedChar != 0 {
		t.Errorf("Expected SelectedChar to wrap to 0, got %d", state.SelectedChar)
	}
}

func TestCharacterCreationState_NavigateCharacterLeft(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 2

	state.NavigateCharacter(-1)
	if state.SelectedChar != 1 {
		t.Errorf("Expected SelectedChar 1, got %d", state.SelectedChar)
	}
}

func TestCharacterCreationState_NavigateCharacterWrapsLeft(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0

	state.NavigateCharacter(-1)
	if state.SelectedChar != 3 {
		t.Errorf("Expected SelectedChar to wrap to 3, got %d", state.SelectedChar)
	}
}

func TestCharacterCreationState_NavigateFieldDown(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedField = FieldName

	state.NavigateField(1)
	if state.SelectedField != FieldFood {
		t.Errorf("Expected SelectedField FieldFood, got %d", state.SelectedField)
	}
}

func TestCharacterCreationState_NavigateFieldWrapsDown(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedField = FieldColor

	state.NavigateField(1)
	if state.SelectedField != FieldName {
		t.Errorf("Expected SelectedField to wrap to FieldName, got %d", state.SelectedField)
	}
}

func TestCharacterCreationState_NavigateFieldUp(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedField = FieldColor

	state.NavigateField(-1)
	if state.SelectedField != FieldFood {
		t.Errorf("Expected SelectedField FieldFood, got %d", state.SelectedField)
	}
}

func TestCharacterCreationState_NavigateFieldWrapsUp(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedField = FieldName

	state.NavigateField(-1)
	if state.SelectedField != FieldColor {
		t.Errorf("Expected SelectedField to wrap to FieldColor, got %d", state.SelectedField)
	}
}

func TestCharacterCreationState_NextField(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedField = FieldName

	state.NextField()
	if state.SelectedField != FieldFood {
		t.Errorf("Expected SelectedField FieldFood, got %d", state.SelectedField)
	}

	state.NextField()
	if state.SelectedField != FieldColor {
		t.Errorf("Expected SelectedField FieldColor, got %d", state.SelectedField)
	}

	state.NextField()
	if state.SelectedField != FieldName {
		t.Errorf("Expected SelectedField to wrap to FieldName, got %d", state.SelectedField)
	}
}

// =============================================================================
// Name Editing Tests
// =============================================================================

func TestCharacterCreationState_TypeCharacter(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.Characters[0].Name = "Len"

	state.TypeCharacter('a')
	if state.Characters[0].Name != "Lena" {
		t.Errorf("Expected name 'Lena', got %q", state.Characters[0].Name)
	}
}

func TestCharacterCreationState_TypeSpace(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.Characters[0].Name = "Len"

	state.TypeCharacter(' ')
	if state.Characters[0].Name != "Len " {
		t.Errorf("Expected name 'Len ', got %q", state.Characters[0].Name)
	}
}

func TestCharacterCreationState_Backspace(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.Characters[0].Name = "Len"

	state.Backspace()
	if state.Characters[0].Name != "Le" {
		t.Errorf("Expected name 'Le', got %q", state.Characters[0].Name)
	}
}

func TestCharacterCreationState_BackspaceOnEmpty(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.Characters[0].Name = ""

	state.Backspace() // Should not panic
	if state.Characters[0].Name != "" {
		t.Errorf("Expected name to remain empty, got %q", state.Characters[0].Name)
	}
}

func TestCharacterCreationState_NameMaxLength(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.Characters[0].Name = "1234567890123456" // 16 chars (max)

	state.TypeCharacter('X')
	if len(state.Characters[0].Name) != 16 {
		t.Errorf("Expected name to stay at 16 chars, got %d", len(state.Characters[0].Name))
	}
	if state.Characters[0].Name != "1234567890123456" {
		t.Errorf("Expected name unchanged, got %q", state.Characters[0].Name)
	}
}

// =============================================================================
// Option Cycling Tests
// =============================================================================

func TestCharacterCreationState_CycleFoodOption(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.SelectedField = FieldFood
	state.Characters[0].Food = foodOptions[0] // Start with first food option

	// Cycle through all food options and verify it wraps back
	for i := 1; i <= len(foodOptions); i++ {
		state.CycleOption()
		expected := foodOptions[i%len(foodOptions)]
		if state.Characters[0].Food != expected {
			t.Errorf("After %d cycles: expected food %q, got %q", i, expected, state.Characters[0].Food)
		}
	}
}

func TestCharacterCreationState_CycleColorOption(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.SelectedField = FieldColor
	state.Characters[0].Color = colorOptions[0] // Start with first color

	// Cycle through all colors and verify it wraps back
	for i := 1; i <= len(colorOptions); i++ {
		state.CycleOption()
		expected := colorOptions[i%len(colorOptions)]
		if state.Characters[0].Color != expected {
			t.Errorf("After %d cycles: expected color %q, got %q", i, expected, state.Characters[0].Color)
		}
	}
}

func TestCharacterCreationState_CycleOptionOnNameFieldDoesNothing(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	state.SelectedChar = 0
	state.SelectedField = FieldName
	state.Characters[0].Name = "Len"

	state.CycleOption()
	// Should not change anything
	if state.Characters[0].Name != "Len" {
		t.Errorf("CycleOption on Name field should not change name, got %q", state.Characters[0].Name)
	}
}

// =============================================================================
// Randomize Tests
// =============================================================================

func TestCharacterCreationState_RandomizeAll(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()
	// Modify all names
	state.Characters[0].Name = "Modified1"
	state.Characters[1].Name = "Modified2"
	state.Characters[2].Name = "Modified3"
	state.Characters[3].Name = "Modified4"

	state.RandomizeAll()

	// Names should be reset to defaults
	expectedNames := []string{"Len", "Macca", "Hari", "Starr"}
	for i, expected := range expectedNames {
		if state.Characters[i].Name != expected {
			t.Errorf("Character %d name: expected %q, got %q", i, expected, state.Characters[i].Name)
		}
	}

	// Build valid foods map from foodOptions
	validFoods := make(map[string]bool)
	for _, f := range foodOptions {
		validFoods[f] = true
	}

	// Build valid colors map from colorOptions
	validColors := make(map[string]bool)
	for _, c := range colorOptions {
		validColors[c] = true
	}

	// Food and color should still be valid (randomized)
	for i, char := range state.Characters {
		if !validFoods[char.Food] {
			t.Errorf("Character %d has invalid food after randomize: %q", i, char.Food)
		}
		if !validColors[char.Color] {
			t.Errorf("Character %d has invalid color after randomize: %q", i, char.Color)
		}
	}
}

// =============================================================================
// Field Type Detection Tests
// =============================================================================

func TestCharacterCreationState_IsNameFieldSelected(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()

	state.SelectedField = FieldName
	if !state.IsNameFieldSelected() {
		t.Error("Expected IsNameFieldSelected to be true for FieldName")
	}

	state.SelectedField = FieldFood
	if state.IsNameFieldSelected() {
		t.Error("Expected IsNameFieldSelected to be false for FieldFood")
	}

	state.SelectedField = FieldColor
	if state.IsNameFieldSelected() {
		t.Error("Expected IsNameFieldSelected to be false for FieldColor")
	}
}

func TestCharacterCreationState_IsOptionFieldSelected(t *testing.T) {
	t.Parallel()

	state := NewCharacterCreationState()

	state.SelectedField = FieldName
	if state.IsOptionFieldSelected() {
		t.Error("Expected IsOptionFieldSelected to be false for FieldName")
	}

	state.SelectedField = FieldFood
	if !state.IsOptionFieldSelected() {
		t.Error("Expected IsOptionFieldSelected to be true for FieldFood")
	}

	state.SelectedField = FieldColor
	if !state.IsOptionFieldSelected() {
		t.Error("Expected IsOptionFieldSelected to be true for FieldColor")
	}
}
