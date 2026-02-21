package ui

import (
	"testing"

	"petri/internal/config"
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
	if !found.IsPoisonous() {
		t.Error("Expected item to be poisonous")
	}
}

func TestFromSaveState_RestoresFeatures(t *testing.T) {
	m := createTestModel()

	// Add a leaf pile feature
	leafPile := entity.NewLeafPile(7, 7)
	m.gameMap.AddFeature(leafPile)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Find leaf pile
	foundBed := restored.gameMap.BedAt(types.Position{X: 7, Y: 7})
	if foundBed == nil {
		t.Error("Expected to find leaf pile at (7,7)")
	}
}

func TestFromSaveState_MigratesSpringFeaturesToWater(t *testing.T) {
	m := createTestModel()

	// Add a spring as a feature (simulating old save format)
	spring := entity.NewSpring(3, 3)
	m.gameMap.AddFeature(spring)

	// Also add a leaf pile to verify it's not affected
	leafPile := entity.NewLeafPile(7, 7)
	m.gameMap.AddFeature(leafPile)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Spring should be migrated to water tile, not restored as feature
	if !restored.gameMap.IsWater(types.Position{X: 3, Y: 3}) {
		t.Error("Expected spring at (3,3) to be migrated to water tile")
	}
	if restored.gameMap.WaterAt(types.Position{X: 3, Y: 3}) != game.WaterSpring {
		t.Error("Expected water type WaterSpring at (3,3)")
	}

	// Leaf pile should still be a feature
	foundBed := restored.gameMap.BedAt(types.Position{X: 7, Y: 7})
	if foundBed == nil {
		t.Error("Expected leaf pile at (7,7) to remain as feature")
	}
}

func TestFromSaveState_RestoresWaterTiles(t *testing.T) {
	m := createTestModel()

	// Add water tiles (new save format)
	m.gameMap.AddWater(types.Position{X: 5, Y: 5}, game.WaterSpring)
	m.gameMap.AddWater(types.Position{X: 8, Y: 8}, game.WaterPond)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Verify water tiles restored
	if restored.gameMap.WaterAt(types.Position{X: 5, Y: 5}) != game.WaterSpring {
		t.Error("Expected WaterSpring at (5,5)")
	}
	if restored.gameMap.WaterAt(types.Position{X: 8, Y: 8}) != game.WaterPond {
		t.Error("Expected WaterPond at (8,8)")
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

func TestFromSaveState_RestoresVesselWithContents(t *testing.T) {
	m := createTestModel()

	// Create a gourd to make vessel from
	gourd := entity.NewGourd(20, 20, types.ColorOrange, types.PatternStriped, types.TextureWarty, false, false)

	// Create vessel from gourd (matches CreateVessel output)
	vessel := &entity.Item{
		ID:       999,
		ItemType: "vessel",
		Kind:     "hollow gourd",
		Color:   gourd.Color,
		Pattern: gourd.Pattern,
		Texture: gourd.Texture,
		// Edible is nil - vessels are not edible
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.X = 20
	vessel.Y = 20
	vessel.EType = entity.TypeItem

	// Create a variety for the contents and register it
	berryVariety := &entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Pattern:  types.PatternNone,
		Texture:  types.TextureNone,
		Edible:   &entity.EdibleProperties{Healing: true},
	}
	m.gameMap.Varieties().Register(berryVariety)

	// Add contents to vessel
	vessel.Container.Contents = []entity.Stack{
		{Variety: berryVariety, Count: 5},
	}

	m.gameMap.AddItemDirect(vessel)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Find the vessel
	var restoredVessel *entity.Item
	for _, item := range restored.gameMap.Items() {
		if item.X == 20 && item.Y == 20 && item.ItemType == "vessel" {
			restoredVessel = item
			break
		}
	}

	if restoredVessel == nil {
		t.Fatal("Expected to find vessel at (20,20)")
	}

	// Verify vessel attributes
	if restoredVessel.Kind != "hollow gourd" {
		t.Errorf("Expected Kind 'hollow gourd', got '%s'", restoredVessel.Kind)
	}
	if restoredVessel.Color != types.ColorOrange {
		t.Errorf("Expected orange color, got %s", restoredVessel.Color)
	}
	if restoredVessel.Pattern != types.PatternStriped {
		t.Errorf("Expected striped pattern, got %s", restoredVessel.Pattern)
	}
	if restoredVessel.Texture != types.TextureWarty {
		t.Errorf("Expected warty texture, got %s", restoredVessel.Texture)
	}

	// Verify container exists
	if restoredVessel.Container == nil {
		t.Fatal("Expected vessel to have Container")
	}
	if restoredVessel.Container.Capacity != 1 {
		t.Errorf("Expected capacity 1, got %d", restoredVessel.Container.Capacity)
	}

	// Verify contents
	if len(restoredVessel.Container.Contents) != 1 {
		t.Fatalf("Expected 1 stack in contents, got %d", len(restoredVessel.Container.Contents))
	}

	stack := restoredVessel.Container.Contents[0]
	if stack.Count != 5 {
		t.Errorf("Expected count 5, got %d", stack.Count)
	}
	if stack.Variety == nil {
		t.Fatal("Expected stack to have Variety")
	}
	if stack.Variety.ItemType != "berry" {
		t.Errorf("Expected variety item type 'berry', got '%s'", stack.Variety.ItemType)
	}
	if stack.Variety.Color != types.ColorRed {
		t.Errorf("Expected variety color red, got %s", stack.Variety.Color)
	}
	if !stack.Variety.IsEdible() {
		t.Error("Expected variety to be edible")
	}
	if !stack.Variety.IsHealing() {
		t.Error("Expected variety to be healing")
	}
}

func TestFromSaveState_RestoresTwoSlotInventory(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		t.Fatal("Need at least 1 character")
	}

	// Create two items for inventory
	berry := entity.NewBerry(0, 0, types.ColorRed, false, true) // healing berry
	berry.ID = 901

	vessel := &entity.Item{
		ID:       902,
		ItemType: "vessel",
		Kind:     "hollow gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternNone,
		Texture:  types.TextureWaxy,
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.EType = entity.TypeItem

	// Fill both inventory slots
	chars[0].AddToInventory(berry)
	chars[0].AddToInventory(vessel)

	if len(chars[0].Inventory) != 2 {
		t.Fatalf("Expected 2 items in inventory, got %d", len(chars[0].Inventory))
	}

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Verify both items restored
	restoredChar := restored.gameMap.Characters()[0]
	if len(restoredChar.Inventory) != 2 {
		t.Fatalf("Expected 2 items in restored inventory, got %d", len(restoredChar.Inventory))
	}

	// Verify first item (berry)
	if restoredChar.Inventory[0].ItemType != "berry" {
		t.Errorf("Expected first item to be berry, got %s", restoredChar.Inventory[0].ItemType)
	}
	if restoredChar.Inventory[0].ID != 901 {
		t.Errorf("Expected first item ID 901, got %d", restoredChar.Inventory[0].ID)
	}
	if !restoredChar.Inventory[0].IsHealing() {
		t.Error("Expected first item to be healing")
	}

	// Verify second item (vessel)
	if restoredChar.Inventory[1].ItemType != "vessel" {
		t.Errorf("Expected second item to be vessel, got %s", restoredChar.Inventory[1].ItemType)
	}
	if restoredChar.Inventory[1].ID != 902 {
		t.Errorf("Expected second item ID 902, got %d", restoredChar.Inventory[1].ID)
	}
	if restoredChar.Inventory[1].Container == nil {
		t.Error("Expected second item to have container")
	}
}

func TestFromSaveState_RestoresCarriedVesselWithContents(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		t.Fatal("Need at least 1 character")
	}

	// Create a variety for the contents
	mushroomVariety := &entity.ItemVariety{
		ID:       entity.GenerateVarietyID("mushroom", types.ColorBlue, types.PatternSpotted, types.TextureSlimy),
		ItemType: "mushroom",
		Color:    types.ColorBlue,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
		Edible:   &entity.EdibleProperties{Poisonous: true},
	}
	m.gameMap.Varieties().Register(mushroomVariety)

	// Create vessel with contents
	vessel := &entity.Item{
		ID:       888,
		ItemType: "vessel",
		Kind:     "hollow gourd",
		Color:   types.ColorGreen,
		Pattern: types.PatternNone,
		Texture: types.TextureWaxy,
		// Edible is nil - vessels are not edible
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: mushroomVariety, Count: 7},
			},
		},
	}
	vessel.EType = entity.TypeItem

	// Give vessel to character
	chars[0].AddToInventory(vessel)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Get restored character
	restoredChar := restored.gameMap.Characters()[0]

	carriedVessel := restoredChar.GetCarriedVessel()
	if carriedVessel == nil {
		t.Fatal("Expected character to be carrying vessel")
	}
	if carriedVessel.ItemType != "vessel" {
		t.Errorf("Expected item type 'vessel', got '%s'", carriedVessel.ItemType)
	}
	if carriedVessel.Container == nil {
		t.Fatal("Expected carried vessel to have Container")
	}

	// Verify contents
	if len(carriedVessel.Container.Contents) != 1 {
		t.Fatalf("Expected 1 stack, got %d", len(carriedVessel.Container.Contents))
	}

	stack := carriedVessel.Container.Contents[0]
	if stack.Count != 7 {
		t.Errorf("Expected count 7, got %d", stack.Count)
	}
	if stack.Variety == nil {
		t.Fatal("Expected stack to have Variety")
	}
	if stack.Variety.ItemType != "mushroom" {
		t.Errorf("Expected 'mushroom', got '%s'", stack.Variety.ItemType)
	}
	if !stack.Variety.IsPoisonous() {
		t.Error("Expected variety to be poisonous")
	}
}

func TestFromSaveState_RestoresItemKind(t *testing.T) {
	m := createTestModel()

	// Add a crafted item with Kind
	hoe := &entity.Item{
		ID:       950,
		ItemType: "hoe",
		Kind:     "shell hoe",
		Color:    types.ColorSilver,
	}
	hoe.X = 25
	hoe.Y = 25
	hoe.EType = entity.TypeItem
	hoe.Sym = 'L'
	m.gameMap.AddItemDirect(hoe)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// Find the hoe
	var found *entity.Item
	for _, item := range restored.gameMap.Items() {
		if item.X == 25 && item.Y == 25 && item.ItemType == "hoe" {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find hoe at (25,25)")
	}
	if found.Kind != "shell hoe" {
		t.Errorf("Expected Kind 'shell hoe', got %q", found.Kind)
	}
}

func TestFromSaveState_RestoresPreferenceKind(t *testing.T) {
	m := createTestModel()
	chars := m.gameMap.Characters()

	// Add preference with Kind
	pref := entity.Preference{
		Valence: 1,
		Kind:    "shell hoe",
		Color:   types.ColorSilver,
	}
	chars[0].Preferences = append(chars[0].Preferences, pref)

	// Round trip
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	restoredChars := restored.gameMap.Characters()
	restoredPref := restoredChars[0].Preferences[len(restoredChars[0].Preferences)-1]

	if restoredPref.Kind != "shell hoe" {
		t.Errorf("Expected Kind 'shell hoe', got %q", restoredPref.Kind)
	}
	if restoredPref.Color != types.ColorSilver {
		t.Errorf("Expected Color silver, got %q", restoredPref.Color)
	}
}

func TestFromSaveState_OldSaveWithoutKind_LoadsEmpty(t *testing.T) {
	m := createTestModel()

	// Regular item without Kind (simulating old save)
	state := m.ToSaveState()
	restored := FromSaveState(state, "test-world", m.testCfg)

	// All items should have empty Kind
	for _, item := range restored.gameMap.Items() {
		if item.Kind != "" {
			t.Errorf("Expected empty Kind for old save item %s, got %q", item.ItemType, item.Kind)
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

func TestSproutSerialization_RoundTrip(t *testing.T) {
	t.Parallel()

	m := createTestModel()

	// Create a sprout on the map
	berry := entity.NewBerry(0, 0, types.ColorRed, true, false)
	berry.Plantable = true
	sprout := entity.CreateSprout(15, 15, berry, berry.Edible)
	m.gameMap.AddItem(sprout)

	// Save
	state := m.ToSaveState()

	// Load into a new model
	m2 := FromSaveState(state, "test", TestConfig{})

	// Find the sprout
	restoredSprout := m2.gameMap.ItemAt(types.Position{X: 15, Y: 15})
	if restoredSprout == nil {
		t.Fatal("Expected sprout at (15,15) after round-trip")
	}
	if restoredSprout.Plant == nil {
		t.Fatal("Expected plant properties on restored sprout")
	}
	if !restoredSprout.Plant.IsSprout {
		t.Error("Expected IsSprout=true after round-trip")
	}
	if !restoredSprout.Plant.IsGrowing {
		t.Error("Expected IsGrowing=true after round-trip")
	}
	if restoredSprout.Plant.SproutTimer <= 0 {
		t.Errorf("Expected positive SproutTimer, got %f", restoredSprout.Plant.SproutTimer)
	}
	if restoredSprout.ItemType != "berry" {
		t.Errorf("Expected ItemType 'berry', got %q", restoredSprout.ItemType)
	}
	if restoredSprout.Sym != config.CharSprout {
		t.Errorf("Expected sprout symbol %c, got %c", config.CharSprout, restoredSprout.Sym)
	}
}

func TestVarietySerialization_PreservesPlantableAndKind(t *testing.T) {
	t.Parallel()

	m := createTestModel()

	// Register a seed variety with Kind and Plantable
	registry := m.gameMap.Varieties()
	seedVariety := &entity.ItemVariety{
		ID:        entity.GenerateVarietyID("seed", types.ColorGreen, types.PatternSpotted, types.TextureWarty),
		ItemType:  "seed",
		Kind:      "gourd seed",
		Color:     types.ColorGreen,
		Pattern:   types.PatternSpotted,
		Texture:   types.TextureWarty,
		Plantable: true,
		Sym:       '.',
	}
	registry.Register(seedVariety)

	state := m.ToSaveState()

	m2 := FromSaveState(state, "test", TestConfig{})

	restoredRegistry := m2.gameMap.Varieties()
	restored := restoredRegistry.Get(seedVariety.ID)
	if restored == nil {
		t.Fatal("Expected seed variety in registry after round-trip")
	}
	if restored.Kind != "gourd seed" {
		t.Errorf("Expected Kind 'gourd seed', got %q", restored.Kind)
	}
	if !restored.Plantable {
		t.Error("Expected Plantable=true after round-trip")
	}
	if restored.Sym != '.' {
		t.Errorf("Expected Sym '.', got %c", restored.Sym)
	}
}

func TestWaterVesselSerialization_RoundTrip(t *testing.T) {
	t.Parallel()

	m := createTestModel()

	// Find the water variety from the registry
	registry := m.gameMap.Varieties()
	waterVariety := registry.VarietiesOfType("liquid")
	if len(waterVariety) == 0 {
		t.Fatal("Expected water variety in registry")
	}

	// Create a vessel with water contents
	vessel := &entity.Item{
		ItemType: "vessel",
		Kind:     "hollow gourd",
		Name:     "Hollow Gourd",
		Color:    types.ColorOrange,
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: waterVariety[0], Count: 3},
			},
		},
	}
	vessel.X = 20
	vessel.Y = 20
	vessel.EType = entity.TypeItem
	vessel.Sym = 'U'
	m.gameMap.AddItemDirect(vessel)

	// Save
	state := m.ToSaveState()

	// Load into a new model
	m2 := FromSaveState(state, "test", TestConfig{})

	// Find the vessel
	var restoredVessel *entity.Item
	for _, item := range m2.gameMap.Items() {
		if item.X == 20 && item.Y == 20 && item.ItemType == "vessel" {
			restoredVessel = item
			break
		}
	}
	if restoredVessel == nil {
		t.Fatal("Expected to find vessel at (20,20)")
	}

	// Verify container exists with contents
	if restoredVessel.Container == nil {
		t.Fatal("Expected vessel to have Container")
	}
	if len(restoredVessel.Container.Contents) != 1 {
		t.Fatalf("Expected 1 stack in contents, got %d", len(restoredVessel.Container.Contents))
	}

	// Verify liquid stack
	stack := restoredVessel.Container.Contents[0]
	if stack.Count != 3 {
		t.Errorf("Expected count 3, got %d", stack.Count)
	}
	if stack.Variety == nil {
		t.Fatal("Expected variety on stack")
	}
	if stack.Variety.ItemType != "liquid" {
		t.Errorf("Expected ItemType 'liquid', got %q", stack.Variety.ItemType)
	}
	if stack.Variety.Kind != "water" {
		t.Errorf("Expected Kind 'water', got %q", stack.Variety.Kind)
	}
}

func TestWateredTileSerialization_RoundTrip(t *testing.T) {
	t.Parallel()

	m := createTestModel()

	// Manually water two tiles
	pos1 := types.Position{X: 3, Y: 3}
	pos2 := types.Position{X: 7, Y: 7}
	m.gameMap.SetManuallyWatered(pos1)
	m.gameMap.SetManuallyWatered(pos2)

	// Partially decay one tile so timers differ
	m.gameMap.UpdateWateredTimers(100.0)

	// Save
	state := m.ToSaveState()

	// Load into a new model
	m2 := FromSaveState(state, "test", TestConfig{})

	// Both tiles should still be watered
	if !m2.gameMap.IsManuallyWatered(pos1) {
		t.Error("pos1 should be watered after round-trip")
	}
	if !m2.gameMap.IsManuallyWatered(pos2) {
		t.Error("pos2 should be watered after round-trip")
	}

	// Both should show as wet
	if !m2.gameMap.IsWet(pos1) {
		t.Error("pos1 should be wet after round-trip")
	}
	if !m2.gameMap.IsWet(pos2) {
		t.Error("pos2 should be wet after round-trip")
	}

	// Unwatered tile should not be wet (no water terrain nearby)
	if m2.gameMap.IsWet(types.Position{X: 15, Y: 15}) {
		t.Error("unwatered tile should not be wet")
	}
}
