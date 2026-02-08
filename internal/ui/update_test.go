package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/save"
	"petri/internal/system"
	"petri/internal/types"
)

// =============================================================================
// applyIntent Tests
// =============================================================================

func TestApplyIntent_CollapseIsImmediate(t *testing.T) {
	t.Parallel()

	// Setup: character at (5,5) with energy=0, intent to move toward bed at (10,10)
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berries", types.ColorRed)
	char.Energy = 0 // Collapsed state
	bed := entity.NewLeafPile(10, 10)
	char.Intent = &entity.Intent{
		Action:        entity.ActionMove,
		Target:        types.Position{X: 6, Y: 5},
		TargetFeature: bed,
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent - should trigger collapse instead of movement
	m.applyIntent(char, 0.1)

	// Assert: character should be sleeping (collapsed on ground)
	if !char.IsSleeping {
		t.Error("Expected character to be sleeping after collapse, but IsSleeping is false")
	}

	// Assert: character should NOT have moved (still at 5,5)
	pos := char.Pos()
	if pos.X != 5 || pos.Y != 5 {
		t.Errorf("Expected character to stay at (5,5) after collapse, but moved to (%d,%d)", pos.X, pos.Y)
	}
}

func TestApplyIntent_NoCollapseWhenEnergyAboveZero(t *testing.T) {
	t.Parallel()

	// Setup: character at (5,5) with energy=1 (just above collapse), intent to move
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berries", types.ColorRed)
	char.Energy = 1 // Just above collapse threshold
	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
	}
	// Need enough speed to actually move (threshold is 7.5)
	char.SpeedAccumulator = 10.0
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent - should move normally
	m.applyIntent(char, 0.1)

	// Assert: character should NOT be sleeping
	if char.IsSleeping {
		t.Error("Expected character with energy=1 to not collapse, but IsSleeping is true")
	}

	// Assert: character should have moved to (6,5)
	pos := char.Pos()
	if pos.X != 6 || pos.Y != 5 {
		t.Errorf("Expected character to move to (6,5), but is at (%d,%d)", pos.X, pos.Y)
	}
}

// =============================================================================
// ActionConsume (Eating from Inventory) Tests - 5.3
// =============================================================================

func TestApplyIntent_ActionConsume_ConsumesFromInventory(t *testing.T) {
	t.Parallel()

	// Setup: character carrying an edible item with ActionConsume intent
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.AddToInventory(carriedItem)
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: carriedItem,
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent with enough time to complete action
	// ActionDuration is typically around 1.0s, so we'll apply multiple small deltas
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: inventory should be cleared
	if len(char.Inventory) != 0 {
		t.Error("Expected inventory to be empty after ActionConsume")
	}

	// Assert: hunger should be reduced
	if char.Hunger >= 80 {
		t.Errorf("Expected hunger to be reduced from 80, got %.2f", char.Hunger)
	}
}

func TestApplyIntent_ActionConsume_RequiresDuration(t *testing.T) {
	t.Parallel()

	// Setup: character carrying an edible item with ActionConsume intent
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.AddToInventory(carriedItem)
	char.ActionProgress = 0
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: carriedItem,
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent with small delta (not enough to complete)
	m.applyIntent(char, 0.1)

	// Assert: action progress should increase but item not consumed yet
	if char.ActionProgress == 0 {
		t.Error("Expected ActionProgress to increase")
	}
	if len(char.Inventory) == 0 {
		t.Error("Expected item to NOT be consumed yet (duration not complete)")
	}
}

func TestApplyIntent_ActionConsume_VerifiesTargetMatchesCarrying(t *testing.T) {
	t.Parallel()

	// Setup: character with intent targeting different item than what they're carrying
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	carriedItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	differentItem := entity.NewBerry(0, 0, types.ColorBlue, false, false) // Different item
	char.AddToInventory(carriedItem)
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: differentItem, // Mismatched!
	}
	gameMap.AddCharacter(char)

	actionLog := system.NewActionLog(100)
	m := &Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Act: apply intent with enough time
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: item should NOT be consumed (target mismatch)
	if len(char.Inventory) == 0 {
		t.Error("Expected item to NOT be consumed when target doesn't match inventory")
	}
}

// =============================================================================
// Craft Order Tests
// =============================================================================

func TestApplyIntent_CraftOrderNotCompletedOnPickup(t *testing.T) {
	t.Parallel()

	// Setup: character with craftVessel order, at position with gourd
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.KnownRecipes = []string{"hollow-gourd"}
	gameMap.AddCharacter(char)

	// Add gourd at character's position
	gourd := entity.NewGourd(5, 5, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(gourd)

	// Create craft order and assign to character
	order := entity.NewOrder(1, "craftVessel", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Set pickup intent for gourd
	char.Intent = &entity.Intent{
		Action:     entity.ActionPickup,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: gourd,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Act: apply intent with enough time to complete pickup
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: character should now be carrying the gourd
	carriedGourd := char.FindInInventory(func(i *entity.Item) bool { return i.ItemType == "gourd" })
	if carriedGourd == nil {
		t.Fatal("Expected character to be carrying gourd after pickup")
	}

	// Assert: order should NOT be completed - should still be assigned
	if order.Status != entity.OrderAssigned {
		t.Errorf("Craft order should still be Assigned after pickup, got %s", order.Status)
	}
	if char.AssignedOrderID == 0 {
		t.Error("Character should still have order assigned after picking up craft input")
	}
}

func TestApplyIntent_CraftConsumesGourdFromVessel(t *testing.T) {
	t.Parallel()

	// Setup: character with full inventory (2 vessels), one vessel contains a gourd
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.KnownRecipes = []string{"hollow-gourd"}
	gameMap.AddCharacter(char)

	// Create variety registry with gourd
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("gourd", types.ColorGreen, types.PatternNone, types.TextureNone),
		ItemType: "gourd",
		Color:    types.ColorGreen,
	})
	gameMap.SetVarieties(registry)

	// First vessel is empty
	vessel1 := &entity.Item{
		ItemType: "vessel",
		Name:     "Vessel 1",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	// Second vessel contains a gourd
	gourdVariety := registry.GetByAttributes("gourd", types.ColorGreen, types.PatternNone, types.TextureNone)
	vessel2 := &entity.Item{
		ItemType: "vessel",
		Name:     "Vessel 2",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: gourdVariety, Count: 1},
			},
		},
	}
	char.AddToInventory(vessel1)
	char.AddToInventory(vessel2)

	// Verify gourd is accessible (in vessel, not loose)
	if !char.HasAccessibleItem("gourd") {
		t.Fatal("Test setup error: gourd should be accessible in vessel")
	}
	if char.FindInInventory(func(i *entity.Item) bool { return i.ItemType == "gourd" }) != nil {
		t.Fatal("Test setup error: gourd should NOT be loose in inventory")
	}

	// Create craft order and assign to character
	order := entity.NewOrder(1, "craftVessel", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Set craft intent with RecipeID
	char.Intent = &entity.Intent{
		Action:   entity.ActionCraft,
		Target:   types.Position{X: 5, Y: 5},
		Dest:     types.Position{X: 5, Y: 5},
		RecipeID: "hollow-gourd",
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Act: apply intent with enough time to complete crafting
	recipe := entity.RecipeRegistry["hollow-gourd"]
	if recipe == nil {
		t.Fatal("hollow-gourd recipe not found")
	}
	iterations := int(recipe.Duration/0.1) + 5
	for i := 0; i < iterations; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: crafting should have completed
	if char.Intent != nil {
		t.Error("Intent should be cleared after crafting completes")
	}

	// Assert: gourd should have been consumed from vessel
	if char.HasAccessibleItem("gourd") {
		t.Error("Gourd should have been consumed from vessel")
	}

	// Assert: a new vessel should have been dropped on the map
	items := gameMap.Items()
	newVesselFound := false
	for _, item := range items {
		if item.Container != nil && item.X == 5 && item.Y == 5 {
			newVesselFound = true
			break
		}
	}
	if !newVesselFound {
		t.Error("Expected crafted vessel to be dropped at character's position")
	}

	// Assert: order should be completed
	if order.Status == entity.OrderAssigned {
		t.Error("Order should be completed after crafting")
	}
}

// =============================================================================
// Harvest Order with Vessel Tests
// =============================================================================

// createTestVesselWithRegistry creates a vessel and registry for testing
func createTestVesselWithRegistry() (*entity.Item, *game.VarietyRegistry) {
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	})
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorBlue, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorBlue,
		Edible: &entity.EdibleProperties{},
	})

	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	return vessel, registry
}

func TestApplyIntent_HarvestOrderWithVessel_ContinuesUntilFull(t *testing.T) {
	t.Parallel()

	// Setup
	gameMap := game.NewMap(20, 20)
	vessel, registry := createTestVesselWithRegistry()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.AddToInventory(vessel)
	gameMap.AddCharacter(char)

	// Add multiple red berries at character's position
	berry1 := entity.NewBerry(5, 5, types.ColorRed, false, false)
	berry2 := entity.NewBerry(5, 6, types.ColorRed, false, false) // Nearby
	gameMap.AddItem(berry1)
	gameMap.AddItem(berry2)

	// Create harvest order
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Set pickup intent for first berry
	char.Intent = &entity.Intent{
		Action:     entity.ActionPickup,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: berry1,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Act: apply intent with enough time to complete first pickup
	// ActionDuration is 0.83s, so 1.0s should complete exactly one pickup
	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: berry should be in vessel
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Expected berry to be added to vessel")
	}
	if vessel.Container.Contents[0].Count != 1 {
		t.Errorf("Expected vessel to have 1 berry, got %d", vessel.Container.Contents[0].Count)
	}

	// Assert: order should NOT be completed yet (vessel not full, more berries exist)
	if order.Status != entity.OrderAssigned {
		t.Errorf("Harvest order should still be Assigned, got %v", order.Status)
	}

	// Assert: character should have new intent to get next berry
	if char.Intent == nil {
		t.Error("Character should have intent to continue harvesting")
	}
}

func TestApplyIntent_HarvestOrderWithVessel_CompletesWhenFull(t *testing.T) {
	t.Parallel()

	// Setup: vessel with gourd (stack size 1 = full after one item)
	gameMap := game.NewMap(20, 20)
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("gourd", types.ColorGreen, types.PatternStriped, types.TextureWarty),
		ItemType: "gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternStriped,
		Texture:  types.TextureWarty,
		Edible: &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}

	char := entity.NewCharacter(1, 5, 5, "TestChar", "gourd", types.ColorGreen)
	char.KnownActivities = []string{"harvest"}
	char.AddToInventory(vessel)
	gameMap.AddCharacter(char)

	// Add gourd at character's position
	gourd := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	gameMap.AddItem(gourd)

	// Create harvest order
	order := entity.NewOrder(1, "harvest", "gourd")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action:     entity.ActionPickup,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: gourd,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Act: apply intent
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: vessel should be full (1 gourd = full)
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Expected gourd to be added to vessel")
	}

	// Assert: order should be completed (vessel full)
	if char.AssignedOrderID != 0 {
		t.Error("Character should have no assigned order after vessel is full")
	}
}

func TestApplyIntent_HarvestOrderWithoutVessel_CompletesAfterOneItem(t *testing.T) {
	t.Parallel()

	// Setup: no vessel, just picking up to inventory
	gameMap := game.NewMap(20, 20)
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Inventory = nil // No vessel
	gameMap.AddCharacter(char)

	// Add berry at character's position
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Create harvest order
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action:     entity.ActionPickup,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: berry,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Act: apply intent
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: character should be carrying berry
	hasBerry := char.FindInInventory(func(i *entity.Item) bool { return i == berry }) != nil
	if !hasBerry {
		t.Error("Character should be carrying berry")
	}

	// Assert: order should be completed (no vessel means single pickup completes)
	if char.AssignedOrderID != 0 {
		t.Error("Character should have no assigned order after pickup")
	}
}

func TestApplyIntent_HarvestOrderWithoutVessel_ContinuesUntilInventoryFull(t *testing.T) {
	t.Parallel()

	// Setup: no vessel, multiple berries to harvest
	gameMap := game.NewMap(20, 20)
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Inventory = nil // Empty inventory
	gameMap.AddCharacter(char)

	// Add multiple berries - one at position, one nearby
	berry1 := entity.NewBerry(5, 5, types.ColorRed, false, false)
	berry2 := entity.NewBerry(6, 5, types.ColorRed, false, false)
	berry3 := entity.NewBerry(7, 5, types.ColorRed, false, false) // Third berry won't fit
	gameMap.AddItem(berry1)
	gameMap.AddItem(berry2)
	gameMap.AddItem(berry3)

	// Create harvest order
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action:     entity.ActionPickup,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: berry1,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Act: apply intent many times to allow movement and pickup
	for i := 0; i < 100; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: character should have 2 items in inventory (inventory capacity)
	if len(char.Inventory) != 2 {
		t.Errorf("Expected 2 items in inventory, got %d", len(char.Inventory))
	}

	// Assert: order should be completed (inventory full)
	if char.AssignedOrderID != 0 {
		t.Error("Character should have no assigned order after inventory is full")
	}

	// Assert: third berry should still be on map (couldn't fit in inventory)
	mapItems := gameMap.Items()
	berryCount := 0
	for _, item := range mapItems {
		if item.ItemType == "berry" {
			berryCount++
		}
	}
	if berryCount != 1 {
		t.Errorf("Expected 1 berry left on map, got %d", berryCount)
	}
}

// =============================================================================
// Generalized Craft Tests (Step 7 — shell hoe + vessel regression)
// =============================================================================

func TestApplyIntent_CraftHoe_CreatesHoeFromStickAndShell(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	char.KnownRecipes = []string{"shell-hoe"}
	gameMap.AddCharacter(char)

	// Character has both inputs
	stick := entity.NewStick(0, 0)
	shell := entity.NewShell(0, 0, types.ColorSilver)
	char.AddToInventory(stick)
	char.AddToInventory(shell)

	// Create craft order and assign to character
	order := entity.NewOrder(1, "craftHoe", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Set craft intent with RecipeID
	char.Intent = &entity.Intent{
		Action:   entity.ActionCraft,
		Target:   types.Position{X: 5, Y: 5},
		Dest:     types.Position{X: 5, Y: 5},
		RecipeID: "shell-hoe",
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Apply intent with enough time to complete crafting
	recipe := entity.RecipeRegistry["shell-hoe"]
	if recipe == nil {
		t.Fatal("shell-hoe recipe not found")
	}
	iterations := int(recipe.Duration/0.1) + 5
	for i := 0; i < iterations; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: crafting should have completed
	if char.Intent != nil {
		t.Error("Intent should be cleared after crafting completes")
	}

	// Assert: stick and shell should be consumed
	if char.HasAccessibleItem("stick") {
		t.Error("Stick should have been consumed")
	}
	if char.HasAccessibleItem("shell") {
		t.Error("Shell should have been consumed")
	}

	// Assert: a hoe should be on the map at character's position
	items := gameMap.Items()
	var hoe *entity.Item
	for _, item := range items {
		if item.ItemType == "hoe" && item.X == 5 && item.Y == 5 {
			hoe = item
			break
		}
	}
	if hoe == nil {
		t.Fatal("Expected crafted hoe to be dropped at character's position")
	}
	if hoe.Kind != "shell hoe" {
		t.Errorf("Expected Kind 'shell hoe', got %q", hoe.Kind)
	}
	if hoe.Color != types.ColorSilver {
		t.Errorf("Expected hoe color Silver (from shell), got %v", hoe.Color)
	}

	// Assert: order should be completed
	if char.AssignedOrderID != 0 {
		t.Error("Character should have no assigned order after crafting completes")
	}
}

func TestApplyIntent_CraftVessel_Regression_StillWorksViaGeneralizedPath(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.KnownRecipes = []string{"hollow-gourd"}
	gameMap.AddCharacter(char)

	// Character has gourd in inventory
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	char.AddToInventory(gourd)

	// Create craft order and assign
	order := entity.NewOrder(1, "craftVessel", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Set craft intent with RecipeID
	char.Intent = &entity.Intent{
		Action:   entity.ActionCraft,
		Target:   types.Position{X: 5, Y: 5},
		Dest:     types.Position{X: 5, Y: 5},
		RecipeID: "hollow-gourd",
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Apply intent
	recipe := entity.RecipeRegistry["hollow-gourd"]
	iterations := int(recipe.Duration/0.1) + 5
	for i := 0; i < iterations; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: crafting completed
	if char.Intent != nil {
		t.Error("Intent should be cleared after crafting completes")
	}

	// Assert: gourd consumed
	if char.HasAccessibleItem("gourd") {
		t.Error("Gourd should have been consumed")
	}

	// Assert: vessel on map
	items := gameMap.Items()
	vesselFound := false
	for _, item := range items {
		if item.Container != nil && item.X == 5 && item.Y == 5 {
			vesselFound = true
			if item.Kind != "hollow gourd" {
				t.Errorf("Expected Kind 'hollow gourd', got %q", item.Kind)
			}
			break
		}
	}
	if !vesselFound {
		t.Fatal("Expected crafted vessel to be dropped at character's position")
	}

	// Assert: order completed
	if char.AssignedOrderID != 0 {
		t.Error("Character should have no assigned order after crafting")
	}
}

// =============================================================================
// World State Reset Tests
// =============================================================================

func TestReturnToWorldSelect_ClearsWorldState(t *testing.T) {
	// Setup: use temp directory for saves
	tempDir := t.TempDir()
	save.SetBaseDir(tempDir)
	defer save.ResetBaseDir()

	// Create a model in playing state with world data
	m := Model{
		phase:           phasePlaying,
		worldID:         "test-world-123",
		actionLog:       system.NewActionLog(100),
		elapsedGameTime: 500.0,
		orders:          []*entity.Order{{ID: 1}},
		nextOrderID:     5,
		gameMap:         game.NewMap(20, 20),
	}

	// Add some log entries
	m.actionLog.Add(1, "TestChar", "test", "Some log entry")

	// Act: simulate pressing ESC to return to world select
	// We can't fully simulate the key press without more infrastructure,
	// but we can call handleKey directly
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := newModel.(Model)

	// Assert: world state should be cleared
	if m2.worldID != "" {
		t.Errorf("Expected worldID to be cleared, got '%s'", m2.worldID)
	}
	if m2.elapsedGameTime != 0 {
		t.Errorf("Expected elapsedGameTime to be 0, got %f", m2.elapsedGameTime)
	}
	if len(m2.orders) != 0 {
		t.Errorf("Expected orders to be cleared, got %d orders", len(m2.orders))
	}
	if m2.nextOrderID != 1 {
		t.Errorf("Expected nextOrderID to be 1, got %d", m2.nextOrderID)
	}
	if m2.actionLog.AllEventCount() != 0 {
		t.Errorf("Expected actionLog to be cleared, got %d events", m2.actionLog.AllEventCount())
	}
	if m2.phase != phaseWorldSelect {
		t.Errorf("Expected phase to be phaseWorldSelect, got %d", m2.phase)
	}
}

// =============================================================================
// Edit Character Name Tests
// =============================================================================

func TestEditName_PressE_EntersEditMode(t *testing.T) {
	t.Parallel()

	// Setup: character at cursor position in select mode
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:     phasePlaying,
		viewMode:  viewModeSelect,
		gameMap:   gameMap,
		cursorX:   5,
		cursorY:   5,
		actionLog: system.NewActionLog(100),
	}

	// Act: press E
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m2 := newModel.(Model)

	// Assert: should be in edit mode with name buffer populated
	if !m2.editingCharacterName {
		t.Error("Expected editingCharacterName to be true")
	}
	if m2.editingNameBuffer != "Alice" {
		t.Errorf("Expected editingNameBuffer to be 'Alice', got %q", m2.editingNameBuffer)
	}
	if m2.editingCharacterID != 1 {
		t.Errorf("Expected editingCharacterID to be 1, got %d", m2.editingCharacterID)
	}
}

func TestEditName_PressE_NoCharacterAtCursor_DoesNothing(t *testing.T) {
	t.Parallel()

	// Setup: no character at cursor position
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:     phasePlaying,
		viewMode:  viewModeSelect,
		gameMap:   gameMap,
		cursorX:   5,
		cursorY:   5, // No character here
		actionLog: system.NewActionLog(100),
	}

	// Act: press E
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m2 := newModel.(Model)

	// Assert: should NOT be in edit mode
	if m2.editingCharacterName {
		t.Error("Expected editingCharacterName to be false when no character at cursor")
	}
}

func TestEditName_PressE_NotInSelectMode_DoesNothing(t *testing.T) {
	t.Parallel()

	// Setup: character at cursor but in AllActivity mode (not select mode)
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:     phasePlaying,
		viewMode:  viewModeAllActivity, // Not in select mode
		gameMap:   gameMap,
		cursorX:   5,
		cursorY:   5,
		actionLog: system.NewActionLog(100),
	}

	// Act: press E
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m2 := newModel.(Model)

	// Assert: should NOT be in edit mode
	if m2.editingCharacterName {
		t.Error("Expected editingCharacterName to be false when not in select mode")
	}
}

func TestEditName_TypeCharacter_AddsToBuffer(t *testing.T) {
	t.Parallel()

	// Setup: in edit mode
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Ali", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:                phasePlaying,
		gameMap:              gameMap,
		cursorX:              5,
		cursorY:              5,
		actionLog:            system.NewActionLog(100),
		editingCharacterName: true,
		editingCharacterID:   1,
		editingNameBuffer:    "Ali",
	}

	// Act: type 'c' and 'e'
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m2 := newModel.(Model)
	newModel, _ = m2.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m3 := newModel.(Model)

	// Assert: buffer should be updated
	if m3.editingNameBuffer != "Alice" {
		t.Errorf("Expected editingNameBuffer to be 'Alice', got %q", m3.editingNameBuffer)
	}
	// Character name should NOT be updated yet
	if char.Name != "Ali" {
		t.Errorf("Expected character name to still be 'Ali', got %q", char.Name)
	}
}

func TestEditName_Backspace_RemovesFromBuffer(t *testing.T) {
	t.Parallel()

	// Setup: in edit mode
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:                phasePlaying,
		gameMap:              gameMap,
		cursorX:              5,
		cursorY:              5,
		actionLog:            system.NewActionLog(100),
		editingCharacterName: true,
		editingCharacterID:   1,
		editingNameBuffer:    "Alice",
	}

	// Act: press backspace
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	m2 := newModel.(Model)

	// Assert: buffer should have last char removed
	if m2.editingNameBuffer != "Alic" {
		t.Errorf("Expected editingNameBuffer to be 'Alic', got %q", m2.editingNameBuffer)
	}
}

func TestEditName_Enter_ConfirmsAndUpdatesCharacter(t *testing.T) {
	t.Parallel()

	// Setup: in edit mode with modified buffer
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:                phasePlaying,
		gameMap:              gameMap,
		cursorX:              5,
		cursorY:              5,
		actionLog:            system.NewActionLog(100),
		editingCharacterName: true,
		editingCharacterID:   1,
		editingNameBuffer:    "Bob",
	}

	// Act: press Enter
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := newModel.(Model)

	// Assert: edit mode should be exited
	if m2.editingCharacterName {
		t.Error("Expected editingCharacterName to be false after Enter")
	}
	// Character name should be updated
	if char.Name != "Bob" {
		t.Errorf("Expected character name to be 'Bob', got %q", char.Name)
	}
}

func TestEditName_Escape_CancelsAndReverts(t *testing.T) {
	t.Parallel()

	// Setup: in edit mode with modified buffer
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:                phasePlaying,
		gameMap:              gameMap,
		cursorX:              5,
		cursorY:              5,
		actionLog:            system.NewActionLog(100),
		editingCharacterName: true,
		editingCharacterID:   1,
		editingNameBuffer:    "Bob",
	}

	// Act: press Escape
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := newModel.(Model)

	// Assert: edit mode should be exited
	if m2.editingCharacterName {
		t.Error("Expected editingCharacterName to be false after Escape")
	}
	// Character name should NOT be changed
	if char.Name != "Alice" {
		t.Errorf("Expected character name to remain 'Alice', got %q", char.Name)
	}
}

func TestEditName_MaxLength_EnforcedAt16(t *testing.T) {
	t.Parallel()

	// Setup: in edit mode with 16-char name
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:                phasePlaying,
		gameMap:              gameMap,
		cursorX:              5,
		cursorY:              5,
		actionLog:            system.NewActionLog(100),
		editingCharacterName: true,
		editingCharacterID:   1,
		editingNameBuffer:    "1234567890123456", // 16 chars - max
	}

	// Act: try to type another character
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	m2 := newModel.(Model)

	// Assert: buffer should not exceed 16 chars
	if len(m2.editingNameBuffer) != 16 {
		t.Errorf("Expected buffer length to remain 16, got %d", len(m2.editingNameBuffer))
	}
	if m2.editingNameBuffer != "1234567890123456" {
		t.Errorf("Expected buffer unchanged, got %q", m2.editingNameBuffer)
	}
}

func TestEditName_EmptyName_NotAllowed(t *testing.T) {
	t.Parallel()

	// Setup: in edit mode with empty buffer
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	m := Model{
		phase:                phasePlaying,
		gameMap:              gameMap,
		cursorX:              5,
		cursorY:              5,
		actionLog:            system.NewActionLog(100),
		editingCharacterName: true,
		editingCharacterID:   1,
		editingNameBuffer:    "",
	}

	// Act: press Enter with empty name
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := newModel.(Model)

	// Assert: should still be in edit mode (empty name not allowed)
	if !m2.editingCharacterName {
		t.Error("Expected to remain in edit mode when name is empty")
	}
	// Character name should be unchanged
	if char.Name != "Alice" {
		t.Errorf("Expected character name to remain 'Alice', got %q", char.Name)
	}
}
