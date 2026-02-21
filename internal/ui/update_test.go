package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/config"
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

// =============================================================================
// ActionTillSoil Tests
// =============================================================================

func TestApplyIntent_TillSoil_SetsTilledAfterDuration(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// Mark tile at character's position
	target := types.Position{X: 5, Y: 5}
	gameMap.MarkForTilling(target)

	// Set up till intent
	char.Intent = &entity.Intent{
		Action: entity.ActionTillSoil,
		Target: target,
		Dest:   target,
	}

	order := entity.NewOrder(1, "tillSoil", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Apply with enough time to complete (ActionDurationMedium = 4.0s)
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Tile should be tilled
	if !gameMap.IsTilled(target) {
		t.Error("Expected tile to be tilled after action completes")
	}

	// Tile should be removed from marked-for-tilling pool
	if gameMap.IsMarkedForTilling(target) {
		t.Error("Expected tile to be removed from marked-for-tilling pool")
	}
}

func TestApplyIntent_TillSoil_RequiresDuration(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	target := types.Position{X: 5, Y: 5}
	gameMap.MarkForTilling(target)

	char.ActionProgress = 0
	char.Intent = &entity.Intent{
		Action: entity.ActionTillSoil,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply small delta — not enough to complete
	m.applyIntent(char, 0.1)

	if char.ActionProgress == 0 {
		t.Error("Expected ActionProgress to increase")
	}
	if gameMap.IsTilled(target) {
		t.Error("Tile should NOT be tilled yet (duration not complete)")
	}
}

func TestApplyIntent_TillSoil_DestroysGrowingItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	target := types.Position{X: 5, Y: 5}
	gameMap.MarkForTilling(target)

	// Growing berry at the target position
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	// berry.Plant.IsGrowing is true by default
	gameMap.AddItem(berry)

	char.Intent = &entity.Intent{
		Action: entity.ActionTillSoil,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply with enough time to complete
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Growing item should be destroyed
	if gameMap.ItemAt(target) != nil {
		t.Error("Expected growing item at target to be destroyed by tilling")
	}
}

func TestApplyIntent_TillSoil_DisplacesNonGrowingItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	target := types.Position{X: 5, Y: 5}
	gameMap.MarkForTilling(target)

	// Non-growing stick at the target position (sticks have Plant == nil)
	stick := entity.NewStick(5, 5)
	gameMap.AddItem(stick)

	char.Intent = &entity.Intent{
		Action: entity.ActionTillSoil,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply with enough time to complete
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Stick should still exist on the map but NOT at the tilled position
	items := gameMap.Items()
	stickFound := false
	for _, item := range items {
		if item == stick {
			stickFound = true
			if item.Pos() == target {
				t.Error("Non-growing item should be displaced away from tilled position")
			}
		}
	}
	if !stickFound {
		t.Error("Non-growing item should still exist on map (displaced, not destroyed)")
	}
}

func TestApplyIntent_TillSoil_DisplacesToCharPosWhenNoAdjacentSpace(t *testing.T) {
	t.Parallel()

	// Use a tiny 3x3 map to easily block all adjacent tiles
	gameMap := game.NewMap(3, 3)
	char := entity.NewCharacter(1, 1, 1, "TestChar", "berry", types.ColorRed)
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	target := types.Position{X: 1, Y: 1}
	gameMap.MarkForTilling(target)

	// Place a non-growing item at target
	stick := entity.NewStick(1, 1)
	gameMap.AddItem(stick)

	// Block all 8 adjacent tiles with water (impassable)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			gameMap.AddWater(types.Position{X: 1 + dx, Y: 1 + dy}, game.WaterPond)
		}
	}

	char.Intent = &entity.Intent{
		Action: entity.ActionTillSoil,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply with enough time to complete
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Stick should still exist — dropped at character position (same as target here)
	items := gameMap.Items()
	stickFound := false
	for _, item := range items {
		if item == stick {
			stickFound = true
		}
	}
	if !stickFound {
		t.Error("Non-growing item should still exist when no adjacent space (dropped at char pos)")
	}
}

// =============================================================================
// Order Flash Confirmation Tests
// =============================================================================

func TestSetOrderFlash_SetsMessageAndTimer(t *testing.T) {
	t.Parallel()

	m := Model{}
	m.setOrderFlash("Harvest berries")

	if m.orderFlashMessage != "Harvest berries" {
		t.Errorf("Expected orderFlashMessage 'Harvest berries', got %q", m.orderFlashMessage)
	}
	if m.orderFlashCount != 1 {
		t.Errorf("Expected orderFlashCount 1, got %d", m.orderFlashCount)
	}
	if m.orderFlashEnd.IsZero() {
		t.Error("Expected orderFlashEnd to be set")
	}
}

func TestSetOrderFlash_IncrementsCountForSameOrder(t *testing.T) {
	t.Parallel()

	m := Model{}
	m.setOrderFlash("Harvest berries")
	m.setOrderFlash("Harvest berries")
	m.setOrderFlash("Harvest berries")

	if m.orderFlashMessage != "Harvest berries" {
		t.Errorf("Expected orderFlashMessage 'Harvest berries', got %q", m.orderFlashMessage)
	}
	if m.orderFlashCount != 3 {
		t.Errorf("Expected orderFlashCount 3, got %d", m.orderFlashCount)
	}
}

func TestSetOrderFlash_ResetsCountForDifferentOrder(t *testing.T) {
	t.Parallel()

	m := Model{}
	m.setOrderFlash("Harvest berries")
	m.setOrderFlash("Harvest berries")
	m.setOrderFlash("Craft hollow gourd")

	if m.orderFlashMessage != "Craft hollow gourd" {
		t.Errorf("Expected orderFlashMessage 'Craft hollow gourd', got %q", m.orderFlashMessage)
	}
	if m.orderFlashCount != 1 {
		t.Errorf("Expected orderFlashCount 1, got %d", m.orderFlashCount)
	}
}

func TestSetOrderFlash_ResetsCountWhenTimerExpired(t *testing.T) {
	t.Parallel()

	m := Model{}
	m.setOrderFlash("Harvest berries")
	m.setOrderFlash("Harvest berries")

	// Expire the timer
	m.orderFlashEnd = m.orderFlashEnd.Add(-3 * time.Second)

	// Same order name but timer expired — should reset to 1
	m.setOrderFlash("Harvest berries")

	if m.orderFlashCount != 1 {
		t.Errorf("Expected orderFlashCount to reset to 1 after timer expired, got %d", m.orderFlashCount)
	}
}

func TestApplyIntent_TillSoil_CompletesAndRemovesOrderWhenPoolEmpty(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// Mark only one tile — after tilling it, pool will be empty
	target := types.Position{X: 5, Y: 5}
	gameMap.MarkForTilling(target)

	order := entity.NewOrder(1, "tillSoil", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionTillSoil,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Apply with enough time to complete
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Tile should be tilled
	if !gameMap.IsTilled(target) {
		t.Fatal("Expected tile to be tilled")
	}

	// Order should be marked completed
	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}

	// Character should be unassigned
	if char.AssignedOrderID != 0 {
		t.Error("Expected char.AssignedOrderID to be cleared after completion")
	}

	// Sweep should remove completed orders
	m.sweepCompletedOrders()
	if len(m.orders) != 0 {
		t.Errorf("Expected order to be removed after sweep, got %d orders", len(m.orders))
	}
}

// =============================================================================
// ActionPlant Tests
// =============================================================================

func TestApplyIntent_Plant_ConsumesItemAndCreatesSprout(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Create a plantable berry and add to inventory
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)

	// Tile must be tilled
	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)

	// Create and assign order
	order := entity.NewOrder(1, "plant", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Apply with enough time to complete (ActionDurationMedium = 4.0)
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Berry should be consumed from inventory
	if len(char.Inventory) != 0 {
		nonNil := 0
		for _, item := range char.Inventory {
			if item != nil {
				nonNil++
			}
		}
		if nonNil != 0 {
			t.Errorf("Expected berry consumed from inventory, got %d items", nonNil)
		}
	}

	// Sprout should exist at the position
	sprout := gameMap.ItemAt(target)
	if sprout == nil {
		t.Fatal("Expected sprout at tilled position")
	}
	if sprout.Plant == nil || !sprout.Plant.IsSprout {
		t.Error("Expected sprout to have IsSprout=true")
	}
	if !sprout.Plant.IsGrowing {
		t.Error("Expected sprout to have IsGrowing=true")
	}
	if sprout.Sym != config.CharSprout {
		t.Errorf("Expected sprout symbol %c, got %c", config.CharSprout, sprout.Sym)
	}
	if sprout.ItemType != "berry" {
		t.Errorf("Expected sprout ItemType 'berry', got %q", sprout.ItemType)
	}
	if sprout.Color != types.ColorRed {
		t.Errorf("Expected sprout color red, got %q", sprout.Color)
	}

	// Intent should be cleared
	if char.Intent != nil {
		t.Error("Expected intent to be cleared after planting")
	}
}

func TestApplyIntent_Plant_GourdSeed_SproutHasGourdType(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Create a gourd seed and add to inventory
	seed := entity.NewSeed(0, 0, "gourd", types.ColorGreen, types.PatternSpotted, types.TextureWarty)
	char.AddToInventory(seed)

	// Tile must be tilled
	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)

	// Create and assign order for gourd seeds
	order := entity.NewOrder(1, "plant", "gourd seed")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	sprout := gameMap.ItemAt(target)
	if sprout == nil {
		t.Fatal("Expected sprout at tilled position")
	}
	// Gourd seed → sprout with ItemType "gourd"
	if sprout.ItemType != "gourd" {
		t.Errorf("Expected sprout ItemType 'gourd', got %q", sprout.ItemType)
	}
	// Should preserve the seed's visual attributes
	if sprout.Color != types.ColorGreen {
		t.Errorf("Expected sprout color green, got %q", sprout.Color)
	}
	if sprout.Pattern != types.PatternSpotted {
		t.Errorf("Expected sprout pattern spotted, got %q", sprout.Pattern)
	}
}

func TestApplyIntent_Plant_SetsLockedVarietyOnOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(0, 0, types.ColorBlue, true, false)
	berry.Plantable = true
	char.AddToInventory(berry)

	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)

	order := entity.NewOrder(1, "plant", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	// LockedVariety should be empty initially
	if order.LockedVariety != "" {
		t.Fatal("Expected LockedVariety to be empty initially")
	}

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// LockedVariety should now be set to the berry's variety ID
	expectedVariety := entity.GenerateVarietyID("berry", types.ColorBlue, types.PatternNone, types.TextureNone)
	if order.LockedVariety != expectedVariety {
		t.Errorf("Expected LockedVariety %q, got %q", expectedVariety, order.LockedVariety)
	}
}

func TestApplyIntent_Plant_PreservesEdibleOnSprout(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Create a poisonous mushroom
	mush := entity.NewMushroom(0, 0, types.ColorPurple, types.PatternSpotted, types.TextureSlimy, true, false)
	mush.Plantable = true
	char.AddToInventory(mush)

	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)

	order := entity.NewOrder(1, "plant", "mushroom")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	sprout := gameMap.ItemAt(target)
	if sprout == nil {
		t.Fatal("Expected sprout at tilled position")
	}
	if sprout.Edible == nil {
		t.Fatal("Expected sprout to have Edible properties")
	}
	if !sprout.Edible.Poisonous {
		t.Error("Expected sprout to be poisonous (inherited from parent)")
	}
}

func TestApplyIntent_Plant_ExtractsFromVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Create variety registry with a plantable berry variety
	registry := game.NewVarietyRegistry()
	berryVariety := &entity.ItemVariety{
		ID:        entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType:  "berry",
		Color:     types.ColorRed,
		Plantable: true,
		Edible:    &entity.EdibleProperties{},
		Sym:       config.CharBerry,
	}
	registry.Register(berryVariety)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Vessel containing 2 plantable berries — no loose berries in inventory
	vessel := &entity.Item{
		ItemType: "vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: berryVariety, Count: 2},
			},
		},
	}
	vessel.EType = entity.TypeItem
	vessel.Sym = config.CharVessel
	char.AddToInventory(vessel)

	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)

	order := entity.NewOrder(1, "plant", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Sprout should exist
	sprout := gameMap.ItemAt(target)
	if sprout == nil {
		t.Fatal("Expected sprout at tilled position (berry extracted from vessel)")
	}
	if sprout.ItemType != "berry" {
		t.Errorf("Expected sprout ItemType 'berry', got %q", sprout.ItemType)
	}
	if sprout.Plant == nil || !sprout.Plant.IsSprout {
		t.Error("Expected sprout to have IsSprout=true")
	}

	// Vessel should have 1 berry remaining
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Expected vessel to still have berries")
	}
	if vessel.Container.Contents[0].Count != 1 {
		t.Errorf("Expected 1 berry remaining in vessel, got %d", vessel.Container.Contents[0].Count)
	}
}

// =============================================================================
// Order Completion Sweep Tests
// =============================================================================

func TestSweepCompletedOrders_RemovesCompletedOrders(t *testing.T) {
	t.Parallel()

	order1 := entity.NewOrder(1, "harvest", "berry")
	order1.Status = entity.OrderCompleted
	order2 := entity.NewOrder(2, "harvest", "mushroom")
	order2.Status = entity.OrderOpen
	order3 := entity.NewOrder(3, "tillSoil", "")
	order3.Status = entity.OrderCompleted

	m := Model{
		orders: []*entity.Order{order1, order2, order3},
	}

	m.sweepCompletedOrders()

	if len(m.orders) != 1 {
		t.Fatalf("Expected 1 order remaining, got %d", len(m.orders))
	}
	if m.orders[0].ID != 2 {
		t.Errorf("Expected remaining order ID=2, got ID=%d", m.orders[0].ID)
	}
}

func TestApplyIntent_Plant_CompletesOrderWhenLastItemPlanted(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Only one plantable berry — after planting it, no more items
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)

	// Two tilled tiles — but only one plantable item, so order completes after first plant
	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)
	gameMap.SetTilled(types.Position{X: 6, Y: 5})

	order := entity.NewOrder(1, "plant", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Sprout should exist
	if gameMap.ItemAt(target) == nil {
		t.Fatal("Expected sprout at tilled position")
	}

	// Order should be marked completed (last plantable item used)
	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}

	if char.AssignedOrderID != 0 {
		t.Error("Expected char.AssignedOrderID to be cleared after completion")
	}
}

func TestApplyIntent_Plant_CompletesOrderWhenNoMoreTilledTiles(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Two plantable berries but only one tilled tile
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry1.Plantable = true
	char.AddToInventory(berry1)

	berry2 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry2.Plantable = true
	char.AddToInventory(berry2)

	// Only one tilled tile — after planting, no more empty tiles
	target := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(target)

	order := entity.NewOrder(1, "plant", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action: entity.ActionPlant,
		Target: target,
		Dest:   target,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Order should be marked completed (no more empty tilled tiles)
	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}
}

func TestApplyIntent_FillVessel_FillsAfterDuration(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Give character an empty vessel
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	char.AddToInventory(vessel)

	// Water adjacent to character
	waterPos := types.Position{X: 5, Y: 6}
	gameMap.AddWater(waterPos, game.WaterPond)

	// Set up fill vessel intent
	char.Intent = &entity.Intent{
		Action:     entity.ActionFillVessel,
		Target:     char.Pos(),
		Dest:       char.Pos(),
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply with enough time to complete (ActionDurationShort = 0.83s)
	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Vessel should now contain water
	if len(vessel.Container.Contents) != 1 {
		t.Fatalf("Expected 1 stack in vessel, got %d", len(vessel.Container.Contents))
	}
	stack := vessel.Container.Contents[0]
	if stack.Variety == nil || stack.Variety.ItemType != "liquid" {
		t.Error("Expected liquid variety in vessel")
	}
	if stack.Count != 4 {
		t.Errorf("Expected 4 water units, got %d", stack.Count)
	}
	// Intent should be cleared
	if char.Intent != nil {
		t.Error("Expected intent to be cleared after filling")
	}
}

func TestApplyIntent_FillVessel_DoesNotOverfillPartialVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Give character a vessel with 2 water units already
	waterVariety := registry.VarietiesOfType("liquid")[0]
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: waterVariety, Count: 2},
			},
		},
	}
	char.AddToInventory(vessel)

	// Water adjacent to character
	waterPos := types.Position{X: 5, Y: 6}
	gameMap.AddWater(waterPos, game.WaterPond)

	// Set up fill vessel intent
	char.Intent = &entity.Intent{
		Action:     entity.ActionFillVessel,
		Target:     char.Pos(),
		Dest:       char.Pos(),
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply with enough time to complete
	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Vessel should be capped at 4 (stack size), not 6
	if vessel.Container.Contents[0].Count != 4 {
		t.Errorf("Expected 4 water units (capped), got %d", vessel.Container.Contents[0].Count)
	}
}

// =============================================================================
// ActionForage Tests
// =============================================================================

func TestApplyIntent_Forage_DirectFoodPickup_GoesIdle(t *testing.T) {
	// Anchor: character forages food directly (no vessel). After pickup, goes idle.
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Growing berry at character's position
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	berry.ID = 10
	gameMap.AddItem(berry)

	char.Intent = &entity.Intent{
		Action:     entity.ActionForage,
		Target:     char.Pos(),
		Dest:       char.Pos(),
		TargetItem: berry,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply enough ticks to complete pickup
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Berry should be in inventory
	hasBerry := false
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == "berry" {
			hasBerry = true
			break
		}
	}
	if !hasBerry {
		t.Error("Expected berry in inventory after forage pickup")
	}

	// Character should be idle
	if char.Intent != nil {
		t.Error("Expected intent to be nil after forage completes")
	}
	if char.CurrentActivity != "Idle" {
		t.Errorf("Expected activity 'Idle', got %q", char.CurrentActivity)
	}
}

func TestApplyIntent_Forage_VesselThenFood_ContinuousAction(t *testing.T) {
	// Anchor: character picks up vessel (phase 1), then continues to food (phase 2)
	// as one continuous action — no idle gap in between.
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	// Registry must include a red berry variety so AddToVessel can look it up
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Empty vessel at character's position
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.ID = 20
	vessel.X = 5
	vessel.Y = 5
	vessel.EType = entity.TypeItem
	gameMap.AddItem(vessel)

	// Growing berry nearby
	berry := entity.NewBerry(5, 6, types.ColorRed, false, false)
	berry.ID = 21
	gameMap.AddItem(berry)

	// Intent targets vessel (phase 1)
	char.Intent = &entity.Intent{
		Action:     entity.ActionForage,
		Target:     char.Pos(),
		Dest:       char.Pos(),
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Run ticks one at a time until vessel is picked up
	vesselPickedUp := false
	for i := 0; i < 30; i++ {
		m.applyIntent(char, 0.1)
		if char.GetCarriedVessel() != nil && !vesselPickedUp {
			vesselPickedUp = true

			// KEY ASSERTION: after vessel pickup, intent should target food — NOT idle.
			// This is the core of the fix: no idle gap between vessel and food.
			if char.Intent == nil {
				t.Fatal("Expected intent to exist after vessel pickup (should continue to food)")
			}
			if char.Intent.Action != entity.ActionForage {
				t.Errorf("Expected ActionForage intent for food phase, got %d", char.Intent.Action)
			}
			if char.Intent.TargetItem == vessel {
				t.Error("Intent should target food item, not vessel")
			}
		}
	}

	if !vesselPickedUp {
		t.Fatal("Vessel was never picked up")
	}

	// After all ticks, foraging should have completed: berry in vessel, character idle
	hasFood := false
	for _, item := range char.Inventory {
		if item != nil && item.Container != nil && len(item.Container.Contents) > 0 {
			if item.Container.Contents[0].Variety != nil && item.Container.Contents[0].Variety.ItemType == "berry" {
				hasFood = true
			}
		}
	}
	if !hasFood {
		t.Error("Expected berry in vessel after food pickup")
	}

	if char.CurrentActivity != "Idle" {
		t.Errorf("Expected activity 'Idle' after completing forage, got %q", char.CurrentActivity)
	}
}

func TestApplyIntent_Forage_VesselGone_GoesIdle(t *testing.T) {
	// When vessel target disappears during procurement, character goes idle.
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Vessel exists as struct but NOT on the map (simulates being taken)
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.ID = 30
	vessel.X = 5
	vessel.Y = 5

	char.Intent = &entity.Intent{
		Action:     entity.ActionForage,
		Target:     char.Pos(),
		Dest:       char.Pos(),
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	m.applyIntent(char, 0.1)

	// Should fail and go idle
	if char.Intent != nil {
		t.Error("Expected intent to be nil after procurement failure")
	}
	if char.CurrentActivity != "Idle" {
		t.Errorf("Expected 'Idle', got %q", char.CurrentActivity)
	}
}

// =============================================================================
// ActionFillVessel Tests
// =============================================================================

func TestApplyIntent_FillVessel_GroundVesselPickupTransitionsToPhase2(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Place empty vessel on the ground at same position as character
	vessel := &entity.Item{
		ItemType: "vessel",
		Kind:     "hollow gourd",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.X = 5
	vessel.Y = 5
	vessel.EType = entity.TypeItem
	gameMap.AddItem(vessel)

	// Water nearby
	waterPos := types.Position{X: 5, Y: 8}
	gameMap.AddWater(waterPos, game.WaterPond)

	// Intent targets ground vessel (phase 1)
	char.Intent = &entity.Intent{
		Action:     entity.ActionFillVessel,
		Target:     char.Pos(),
		Dest:       char.Pos(),
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply enough ticks to pick up vessel (phase 1) and transition to phase 2
	// This previously panicked with nil pointer dereference on char.Intent
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Vessel should be in inventory
	if char.GetCarriedVessel() == nil {
		t.Fatal("Expected vessel in inventory after phase 1 pickup")
	}

	// Intent should still exist (transitioning to phase 2, moving to water)
	if char.Intent == nil {
		t.Fatal("Expected intent to exist for phase 2 water-seeking")
	}
	if char.Intent.Action != entity.ActionFillVessel {
		t.Errorf("Expected ActionFillVessel intent, got %d", char.Intent.Action)
	}
}

// =============================================================================
// ActionDrink — Vessel Drinking (Phase 2)
// =============================================================================

func createTestWaterVessel(x, y int, units int) *entity.Item {
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.X = x
	vessel.Y = y
	if units > 0 {
		waterVariety := &entity.ItemVariety{
			ID:       "liquid-water",
			ItemType: "liquid",
			Kind:     "water",
		}
		system.AddLiquidToVessel(vessel, waterVariety, units)
	}
	return vessel
}

func TestApplyIntent_DrinkFromCarriedVessel_ReducesThirstAndClearsIntent(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Thirst = 40 // Moderately thirsty
	gameMap.AddCharacter(char)

	vessel := createTestWaterVessel(0, 0, 3)
	char.AddToInventory(vessel)

	char.Intent = &entity.Intent{
		Target:      char.Pos(),
		Dest:        char.Pos(),
		Action:      entity.ActionDrink,
		TargetItem:  vessel,
		DrivingStat: types.StatThirst,
		DrivingTier: entity.TierModerate,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	// Apply enough ticks to complete ActionDurationShort (0.83s)
	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Thirst should be reduced
	if char.Thirst >= 40 {
		t.Errorf("Thirst should be reduced after drinking, got %f", char.Thirst)
	}

	// Vessel should have lost one unit
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Vessel should still have contents after one drink")
	}
	if vessel.Container.Contents[0].Count != 2 {
		t.Errorf("Vessel should have 2 units remaining, got %d", vessel.Container.Contents[0].Count)
	}

	// Intent should be cleared (forces re-eval for next drink source)
	if char.Intent != nil {
		t.Error("Expected intent to be cleared after vessel drink")
	}
}

func TestApplyIntent_DrinkFromGroundVessel_ReducesThirstAndClearsIntent(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Thirst = 40
	gameMap.AddCharacter(char)

	// Ground vessel at same position as character
	vessel := createTestWaterVessel(5, 5, 3)
	gameMap.AddItem(vessel)

	char.Intent = &entity.Intent{
		Target:      char.Pos(),
		Dest:        char.Pos(),
		Action:      entity.ActionDrink,
		TargetItem:  vessel,
		DrivingStat: types.StatThirst,
		DrivingTier: entity.TierModerate,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Thirst should be reduced
	if char.Thirst >= 40 {
		t.Errorf("Thirst should be reduced after drinking from ground vessel, got %f", char.Thirst)
	}

	// Vessel should have lost one unit
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Vessel should still have contents")
	}
	if vessel.Container.Contents[0].Count != 2 {
		t.Errorf("Vessel should have 2 units remaining, got %d", vessel.Container.Contents[0].Count)
	}

	// Intent cleared
	if char.Intent != nil {
		t.Error("Expected intent to be cleared after ground vessel drink")
	}
}

func TestApplyIntent_DrinkFromVessel_LastUnit_EmptiesVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Thirst = 40
	gameMap.AddCharacter(char)

	// Vessel with only 1 unit
	vessel := createTestWaterVessel(0, 0, 1)
	char.AddToInventory(vessel)

	char.Intent = &entity.Intent{
		Target:      char.Pos(),
		Dest:        char.Pos(),
		Action:      entity.ActionDrink,
		TargetItem:  vessel,
		DrivingStat: types.StatThirst,
		DrivingTier: entity.TierModerate,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Vessel should be empty (stack removed)
	if len(vessel.Container.Contents) != 0 {
		t.Errorf("Vessel should be empty after last unit consumed, got %d stacks", len(vessel.Container.Contents))
	}

	// Intent cleared
	if char.Intent != nil {
		t.Error("Expected intent to be cleared after emptying vessel")
	}
}

func TestApplyIntent_DrinkFromTerrain_DoesNotClearIntent(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Thirst = 40
	gameMap.AddCharacter(char)

	// Water adjacent to character
	waterPos := types.Position{X: 5, Y: 6}
	gameMap.AddWater(waterPos, game.WaterSpring)

	char.Intent = &entity.Intent{
		Target:         char.Pos(),
		Dest:           char.Pos(),
		Action:         entity.ActionDrink,
		TargetWaterPos: &waterPos,
		DrivingStat:    types.StatThirst,
		DrivingTier:    entity.TierModerate,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
	}

	for i := 0; i < 10; i++ {
		m.applyIntent(char, 0.1)
	}

	// Thirst should be reduced
	if char.Thirst >= 40 {
		t.Errorf("Thirst should be reduced after terrain drinking, got %f", char.Thirst)
	}

	// Intent should NOT be cleared for terrain drinking (continues until sated)
	if char.Intent == nil {
		t.Error("Expected intent to persist for terrain drinking")
	}
}
