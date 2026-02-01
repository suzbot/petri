package ui

import (
	"math/rand"
	"sort"
	"strings"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// Field indices for character creation
const (
	FieldName  = 0
	FieldFood  = 1
	FieldColor = 2
	numFields  = 3
)

// Maximum name length
const MaxNameLength = 16

// foodOptions built dynamically from edible item types
var foodOptions = buildFoodOptions()

// colorOptions built dynamically from types.AllColors
var colorOptions = buildColorOptions()

// buildFoodOptions generates display strings from edible item types
func buildFoodOptions() []string {
	configs := game.GetItemTypeConfigs()
	var options []string
	for itemType, cfg := range configs {
		if cfg.Edible {
			options = append(options, capitalizeItemType(itemType))
		}
	}
	// Sort for consistent ordering (maps iterate randomly)
	sort.Strings(options)
	return options
}

// capitalizeItemType converts an item type to display string (e.g., "berry" -> "Berry")
func capitalizeItemType(itemType string) string {
	if len(itemType) == 0 {
		return itemType
	}
	return strings.ToUpper(itemType[:1]) + itemType[1:]
}

// DisplayToItemType converts a display string back to item type (e.g., "Berry" -> "berry")
func DisplayToItemType(display string) string {
	return strings.ToLower(display)
}

// buildColorOptions generates display strings from types.AllColors
func buildColorOptions() []string {
	options := make([]string, len(types.AllColors))
	for i, c := range types.AllColors {
		options[i] = capitalizeColor(c)
	}
	return options
}

// capitalizeColor converts a types.Color to a display string (e.g., "red" -> "Red")
func capitalizeColor(c types.Color) string {
	s := string(c)
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// DisplayToColor converts a display string back to types.Color (e.g., "Red" -> "red")
func DisplayToColor(display string) types.Color {
	return types.Color(strings.ToLower(display))
}

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

	// Initialize characters with random unique names and random food/color
	names := randomUniqueNames(4)
	for i := 0; i < 4; i++ {
		state.Characters[i] = CharacterCreationData{
			Name:  names[i],
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

// RandomizeAll resets all characters to random names with random food/color
func (s *CharacterCreationState) RandomizeAll() {
	names := randomUniqueNames(4)
	for i := 0; i < 4; i++ {
		s.Characters[i] = CharacterCreationData{
			Name:  names[i],
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

// randomUniqueNames returns n unique random names from entity.CharacterNames
func randomUniqueNames(n int) []string {
	// Shuffle a copy of CharacterNames
	shuffled := make([]string, len(entity.CharacterNames))
	copy(shuffled, entity.CharacterNames)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	// Return first n names
	return shuffled[:n]
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
