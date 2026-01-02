package ui

import (
	"math/rand"
)

// Field indices for character creation
const (
	FieldName  = 0
	FieldFood  = 1
	FieldColor = 2
	numFields  = 3
)

// Food options
const (
	FoodBerry    = "Berry"
	FoodMushroom = "Mushroom"
)

// Color options
const (
	ColorRed   = "Red"
	ColorBlue  = "Blue"
	ColorWhite = "White"
	ColorBrown = "Brown"
)

// Maximum name length
const MaxNameLength = 16

// Default character names
var defaultNames = []string{"Len", "Macca", "Hari", "Starr"}

// Food options list for cycling
var foodOptions = []string{FoodBerry, FoodMushroom}

// Color options list for cycling
var colorOptions = []string{ColorRed, ColorBlue, ColorWhite, ColorBrown}

// CharacterCreationData holds the editable data for one character
type CharacterCreationData struct {
	Name  string
	Food  string
	Color string
}

// CharacterCreationState holds all state for the character creation screen
type CharacterCreationState struct {
	Characters    [4]CharacterCreationData
	SelectedChar  int // 0-3
	SelectedField int // FieldName, FieldFood, FieldColor
}

// NewCharacterCreationState creates a new character creation state with defaults
func NewCharacterCreationState() *CharacterCreationState {
	state := &CharacterCreationState{
		SelectedChar:  0,
		SelectedField: FieldName,
	}

	// Initialize characters with default names and random food/color
	for i := 0; i < 4; i++ {
		state.Characters[i] = CharacterCreationData{
			Name:  defaultNames[i],
			Food:  randomFood(),
			Color: randomColor(),
		}
	}

	return state
}

// NavigateCharacter moves selection to another character (delta: -1 for left, +1 for right)
func (s *CharacterCreationState) NavigateCharacter(delta int) {
	s.SelectedChar = (s.SelectedChar + delta + 4) % 4
}

// NavigateField moves selection to another field (delta: -1 for up, +1 for down)
func (s *CharacterCreationState) NavigateField(delta int) {
	s.SelectedField = (s.SelectedField + delta + numFields) % numFields
}

// NextField moves to the next field (wraps around)
func (s *CharacterCreationState) NextField() {
	s.NavigateField(1)
}

// TypeCharacter adds a character to the current character's name
func (s *CharacterCreationState) TypeCharacter(ch rune) {
	name := s.Characters[s.SelectedChar].Name
	if len(name) < MaxNameLength {
		s.Characters[s.SelectedChar].Name = name + string(ch)
	}
}

// Backspace removes the last character from the current character's name
func (s *CharacterCreationState) Backspace() {
	name := s.Characters[s.SelectedChar].Name
	if len(name) > 0 {
		s.Characters[s.SelectedChar].Name = name[:len(name)-1]
	}
}

// CycleOption cycles the current option field (Food or Color) to the next value
func (s *CharacterCreationState) CycleOption() {
	char := &s.Characters[s.SelectedChar]

	switch s.SelectedField {
	case FieldFood:
		char.Food = nextOption(char.Food, foodOptions)
	case FieldColor:
		char.Color = nextOption(char.Color, colorOptions)
	// FieldName: do nothing
	}
}

// RandomizeAll resets all characters to default names with random food/color
func (s *CharacterCreationState) RandomizeAll() {
	for i := 0; i < 4; i++ {
		s.Characters[i] = CharacterCreationData{
			Name:  defaultNames[i],
			Food:  randomFood(),
			Color: randomColor(),
		}
	}
}

// IsNameFieldSelected returns true if the Name field is currently selected
func (s *CharacterCreationState) IsNameFieldSelected() bool {
	return s.SelectedField == FieldName
}

// IsOptionFieldSelected returns true if Food or Color field is currently selected
func (s *CharacterCreationState) IsOptionFieldSelected() bool {
	return s.SelectedField == FieldFood || s.SelectedField == FieldColor
}

// Helper functions

func randomFood() string {
	return foodOptions[rand.Intn(len(foodOptions))]
}

func randomColor() string {
	return colorOptions[rand.Intn(len(colorOptions))]
}

func nextOption(current string, options []string) string {
	for i, opt := range options {
		if opt == current {
			return options[(i+1)%len(options)]
		}
	}
	// If not found, return first option
	return options[0]
}
