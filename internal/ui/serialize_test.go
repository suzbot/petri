package ui

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/system"
	"petri/internal/types"
)

func TestToSaveState_RoundTrip(t *testing.T) {
	// Create a model with some state
	m := createTestModel()

	// Convert to save state
	state := m.ToSaveState()

	// Verify basic properties
	if state.Version != 1 {
		t.Errorf("Expected version 1, got %d", state.Version)
	}
	if state.MapWidth != m.gameMap.Width {
		t.Errorf("Expected map width %d, got %d", m.gameMap.Width, state.MapWidth)
	}
	if state.MapHeight != m.gameMap.Height {
		t.Errorf("Expected map height %d, got %d", m.gameMap.Height, state.MapHeight)
	}
	if state.ElapsedGameTime != m.elapsedGameTime {
		t.Errorf("Expected elapsed time %f, got %f", m.elapsedGameTime, state.ElapsedGameTime)
	}

	// Verify characters
	if len(state.Characters) != len(m.gameMap.Characters()) {
		t.Errorf("Expected %d characters, got %d", len(m.gameMap.Characters()), len(state.Characters))
	}

	// Verify items
	if len(state.Items) != len(m.gameMap.Items()) {
		t.Errorf("Expected %d items, got %d", len(m.gameMap.Items()), len(state.Items))
	}

	// Verify features
	if len(state.Features) != len(m.gameMap.Features()) {
		t.Errorf("Expected %d features, got %d", len(m.gameMap.Features()), len(state.Features))
	}
}

func TestFromSaveState_RestoresCharacters(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		t.Fatal("Test model should have characters")
	}

	// Modify a character
	chars[0].Health = 75.5
	chars[0].Hunger = 60.0
	chars[0].Poisoned = true
	chars[0].Preferences = append(chars[0].Preferences, entity.Preference{
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Valence:  -1,
	})

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Verify character state
	restoredChars := restored.gameMap.Characters()
	if len(restoredChars) != len(chars) {
		t.Fatalf("Expected %d characters, got %d", len(chars), len(restoredChars))
	}

	restoredChar := restoredChars[0]
	if restoredChar.Health != 75.5 {
		t.Errorf("Expected health 75.5, got %f", restoredChar.Health)
	}
	if restoredChar.Hunger != 60.0 {
		t.Errorf("Expected hunger 60.0, got %f", restoredChar.Hunger)
	}
}

func TestFromSaveState_RestoresItems(t *testing.T) {
	m := createTestModel()

	// Add a specific item at a unique position
	item := entity.NewMushroom(15, 15, types.ColorRed, types.PatternSpotted, types.TextureSlimy, true, false)
	m.gameMap.AddItem(item)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Find the item
	items := restored.gameMap.Items()
	var found *entity.Item
	for _, i := range items {
		if i.X == 15 && i.Y == 15 {
			found = i
			break
		}
	}

	if found == nil {
		t.Fatal("Expected to find item at (15,15)")
	}

	if found.Color != types.ColorRed {
		t.Errorf("Expected red color, got %s", found.Color)
	}
	if found.Pattern != types.PatternSpotted {
		t.Errorf("Expected spotted pattern, got %s", found.Pattern)
	}
	if !found.Poisonous {
		t.Error("Expected item to be poisonous")
	}
}

func TestFromSaveState_RestoresFeatures(t *testing.T) {
	m := createTestModel()

	// Add features
	spring := entity.NewSpring(3, 3)
	leafPile := entity.NewLeafPile(7, 7)
	m.gameMap.AddFeature(spring)
	m.gameMap.AddFeature(leafPile)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	features := restored.gameMap.Features()
	if len(features) < 2 {
		t.Fatalf("Expected at least 2 features, got %d", len(features))
	}

	// Find spring
	foundSpring := restored.gameMap.DrinkSourceAt(3, 3)
	if foundSpring == nil {
		t.Error("Expected to find spring at (3,3)")
	}

	// Find leaf pile
	foundBed := restored.gameMap.BedAt(7, 7)
	if foundBed == nil {
		t.Error("Expected to find leaf pile at (7,7)")
	}
}

func TestFromSaveState_RestoresVarieties(t *testing.T) {
	m := createTestModel()

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Verify variety registry exists
	if restored.gameMap.Varieties() == nil {
		t.Fatal("Expected variety registry to be restored")
	}

	// Verify counts match
	originalCount := m.gameMap.Varieties().Count()
	restoredCount := restored.gameMap.Varieties().Count()
	if restoredCount != originalCount {
		t.Errorf("Expected %d varieties, got %d", originalCount, restoredCount)
	}
}

func TestFromSaveState_RestoresActionLogs(t *testing.T) {
	m := createTestModel()

	// Add some action log entries
	m.actionLog.SetGameTime(10.0)
	m.actionLog.Add(1, "TestChar", "test", "Test message 1")
	m.actionLog.SetGameTime(20.0)
	m.actionLog.Add(1, "TestChar", "test", "Test message 2")

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Verify logs
	events := restored.actionLog.Events(1, 100)
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	if events[0].Message != "Test message 1" {
		t.Errorf("Expected 'Test message 1', got '%s'", events[0].Message)
	}
	if events[0].GameTime != 10.0 {
		t.Errorf("Expected game time 10.0, got %f", events[0].GameTime)
	}
}

func TestFromSaveState_RestoresTalkingWith(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()
	if len(chars) < 2 {
		t.Skip("Need at least 2 characters for this test")
	}

	// Set up talking relationship
	chars[0].TalkingWith = chars[1]
	chars[1].TalkingWith = chars[0]
	chars[0].TalkTimer = 3.5

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	restoredChars := restored.gameMap.Characters()
	if restoredChars[0].TalkingWith == nil {
		t.Error("Expected TalkingWith to be restored for char 0")
	}
	if restoredChars[0].TalkingWith.ID != chars[1].ID {
		t.Errorf("Expected TalkingWith ID %d, got %d", chars[1].ID, restoredChars[0].TalkingWith.ID)
	}
	if restoredChars[0].TalkTimer != 3.5 {
		t.Errorf("Expected TalkTimer 3.5, got %f", restoredChars[0].TalkTimer)
	}
}

func TestFromSaveState_RestoresElapsedGameTime(t *testing.T) {
	m := createTestModel()
	m.elapsedGameTime = 123.456

	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	if restored.elapsedGameTime != 123.456 {
		t.Errorf("Expected elapsed time 123.456, got %f", restored.elapsedGameTime)
	}
}

func TestFromSaveState_RestoresKnowledge(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		t.Fatal("Need at least 1 character")
	}

	// Add knowledge
	knowledge := entity.Knowledge{
		Category: entity.KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
	}
	chars[0].Knowledge = append(chars[0].Knowledge, knowledge)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	restoredChars := restored.gameMap.Characters()
	if len(restoredChars[0].Knowledge) != len(chars[0].Knowledge) {
		t.Fatalf("Expected %d knowledge entries, got %d", len(chars[0].Knowledge), len(restoredChars[0].Knowledge))
	}

	k := restoredChars[0].Knowledge[len(restoredChars[0].Knowledge)-1]
	if k.Category != entity.KnowledgePoisonous {
		t.Errorf("Expected category 'poisonous', got '%s'", k.Category)
	}
	if k.ItemType != "mushroom" {
		t.Errorf("Expected item type 'mushroom', got '%s'", k.ItemType)
	}
}

func TestFromSaveState_RestoresPreferencePatternTexture(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		t.Fatal("Need at least 1 character")
	}

	// Add preference with Pattern and Texture (mushroom-specific attributes)
	pref := entity.Preference{
		Valence:  1,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
	}
	chars[0].Preferences = append(chars[0].Preferences, pref)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	restoredChars := restored.gameMap.Characters()
	restoredPref := restoredChars[0].Preferences[len(restoredChars[0].Preferences)-1]

	if restoredPref.Pattern != types.PatternSpotted {
		t.Errorf("Expected pattern 'spotted', got '%s'", restoredPref.Pattern)
	}
	if restoredPref.Texture != types.TextureSlimy {
		t.Errorf("Expected texture 'slimy', got '%s'", restoredPref.Texture)
	}
}

func TestFromSaveState_RestoresEntitySymbols(t *testing.T) {
	m := createTestModel()

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Check character symbols are set (non-zero)
	for _, char := range restored.gameMap.Characters() {
		if char.Symbol() == 0 {
			t.Errorf("Character %s has unset symbol after restore", char.Name)
		}
	}

	// Check item symbols are set
	for _, item := range restored.gameMap.Items() {
		if item.Symbol() == 0 {
			t.Errorf("Item %s at (%d,%d) has unset symbol after restore", item.ItemType, item.X, item.Y)
		}
	}

	// Check feature symbols are set
	for _, feature := range restored.gameMap.Features() {
		if feature.Symbol() == 0 {
			t.Errorf("Feature at (%d,%d) has unset symbol after restore", feature.X, feature.Y)
		}
	}
}

// createTestModel creates a Model with basic game state for testing
func createTestModel() Model {
	m := Model{
		phase:           phasePlaying,
		actionLog:       system.NewActionLog(200),
		width:           80,
		height:          40,
		paused:          false,
		elapsedGameTime: 50.0,
		testCfg:         TestConfig{},
	}

	// Create map
	m.gameMap = game.NewMap(40, 30)

	// Generate and set varieties
	registry := game.GenerateVarieties()
	m.gameMap.SetVarieties(registry)

	// Add a character
	char := entity.NewCharacter(1, 10, 10, "TestChar", "berry", types.ColorRed)
	m.gameMap.AddCharacter(char)

	// Add second character for talking tests
	char2 := entity.NewCharacter(2, 12, 12, "TestChar2", "mushroom", types.ColorBlue)
	m.gameMap.AddCharacter(char2)

	// Add some items
	m.gameMap.AddItem(entity.NewBerry(5, 5, types.ColorRed, false, false))
	m.gameMap.AddItem(entity.NewFlower(8, 8, types.ColorBlue))

	return m
}
