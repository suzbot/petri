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
// Consumption Duration by Food Tier
// =============================================================================

func TestApplyIntent_ActionConsume_SnackCompletesAfterSnackDuration(t *testing.T) {
	t.Parallel()

	// Anchor: a berry (snack tier) is eaten in ~5 world minutes (0.417 game seconds).
	// After accumulating just under the snack duration, berry should not be consumed.
	// After reaching snack duration, berry should be consumed.
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.AddToInventory(berry)
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: berry,
	}
	gameMap.AddCharacter(char)

	m := &Model{
		gameMap:   gameMap,
		actionLog: system.NewActionLog(100),
	}

	// Accumulate just under snack duration — berry should NOT be consumed
	m.applyIntent(char, config.MealSizeSnack.Duration-0.01)
	if len(char.Inventory) == 0 {
		t.Fatal("Berry should NOT be consumed before snack duration completes")
	}

	// One more tick pushes past threshold — berry should be consumed
	m.applyIntent(char, 0.02)
	if len(char.Inventory) != 0 {
		t.Error("Berry should be consumed after snack duration completes")
	}
	if char.Hunger >= 80 {
		t.Errorf("Hunger should be reduced after eating berry, got %.2f", char.Hunger)
	}
}

func TestApplyIntent_ActionConsume_FeastRequiresFeastDuration(t *testing.T) {
	t.Parallel()

	// Anchor: a gourd (feast tier) takes ~45 world minutes (3.75 game seconds) to eat.
	// After meal-tier duration (1.25s), gourd should NOT be consumed.
	// After feast duration (3.75s), gourd should be consumed.
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	char.AddToInventory(gourd)
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: gourd,
	}
	gameMap.AddCharacter(char)

	m := &Model{
		gameMap:   gameMap,
		actionLog: system.NewActionLog(100),
	}

	// Accumulate meal duration (1.25s) — gourd should NOT be consumed
	m.applyIntent(char, config.MealSizeMeal.Duration)
	if len(char.Inventory) == 0 {
		t.Fatal("Gourd should NOT be consumed after only meal-tier duration")
	}
	if char.Hunger != 80 {
		t.Errorf("Hunger should be unchanged before consumption, got %.2f", char.Hunger)
	}

	// Accumulate remaining feast duration — gourd should be consumed
	remaining := config.MealSizeFeast.Duration - config.MealSizeMeal.Duration + 0.01
	m.applyIntent(char, remaining)
	if len(char.Inventory) != 0 {
		t.Error("Gourd should be consumed after feast duration completes")
	}
	if char.Hunger >= 80 {
		t.Errorf("Hunger should be reduced after eating gourd, got %.2f", char.Hunger)
	}
}

func TestApplyIntent_ActionConsume_VesselUsesContentsFoodTier(t *testing.T) {
	t.Parallel()

	// Anchor: eating berries from a vessel uses snack duration (~5 world minutes),
	// not a default duration. The food type comes from vessel contents, not the vessel itself.
	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.Hunger = 80

	// Create vessel with berries
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := system.CreateVessel(gourd, recipe)
	variety := &entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
	}
	vessel.Container.Contents = []entity.Stack{
		{Variety: variety, Count: 5},
	}
	char.AddToInventory(vessel)
	char.Intent = &entity.Intent{
		Action:     entity.ActionConsume,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: vessel,
	}
	gameMap.AddCharacter(char)

	m := &Model{
		gameMap:   gameMap,
		actionLog: system.NewActionLog(100),
	}

	// Accumulate just under snack duration — should NOT have eaten yet
	m.applyIntent(char, config.MealSizeSnack.Duration-0.01)
	if vessel.Container.Contents[0].Count != 5 {
		t.Fatal("Berry in vessel should NOT be consumed before snack duration completes")
	}

	// Push past snack duration — should have eaten one berry
	m.applyIntent(char, 0.02)
	if vessel.Container.Contents[0].Count != 4 {
		t.Errorf("Expected berry count 4 after eating one from vessel, got %d", vessel.Container.Contents[0].Count)
	}
	if char.Hunger >= 80 {
		t.Errorf("Hunger should be reduced after eating berry from vessel, got %.2f", char.Hunger)
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
		ID:       entity.GenerateVarietyID("gourd", "", types.ColorGreen, types.PatternNone, types.TextureNone),
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
	gourdVariety := registry.GetByAttributes("gourd", "", types.ColorGreen, types.PatternNone, types.TextureNone)
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
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
	})
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", "", types.ColorBlue, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorBlue,
		Edible:   &entity.EdibleProperties{},
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
		ID:       entity.GenerateVarietyID("gourd", "", types.ColorGreen, types.PatternStriped, types.TextureWarty),
		ItemType: "gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternStriped,
		Texture:  types.TextureWarty,
		Edible:   &entity.EdibleProperties{},
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
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
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
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
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

	// Act: simulate pressing q to return to world select
	// We can't fully simulate the key press without more infrastructure,
	// but we can call handleKey directly
	newModel, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
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

	// Create variety registry with berry variety
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:        entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType:  "berry",
		Color:     types.ColorRed,
		Plantable: true,
		Edible:    &entity.EdibleProperties{},
		Sym:       config.CharBerry,
	})
	gameMap.SetVarieties(registry)

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

	// Create variety registry with gourd parent variety
	registry := game.NewVarietyRegistry()
	gourdVariety := &entity.ItemVariety{
		ID:       "gourd-green-spotted-warty",
		ItemType: "gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureWarty,
		Edible:   &entity.EdibleProperties{},
		Sym:      config.CharGourd,
	}
	registry.Register(gourdVariety)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Create a gourd seed and add to inventory
	seed := entity.NewSeed(0, 0, "gourd", "gourd-green-spotted-warty", "", types.ColorGreen, types.PatternSpotted, types.TextureWarty)
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

	// Create variety registry with berry variety
	registry := game.NewVarietyRegistry()
	berryVariety := &entity.ItemVariety{
		ID:        entity.GenerateVarietyID("berry", "", types.ColorBlue, types.PatternNone, types.TextureNone),
		ItemType:  "berry",
		Color:     types.ColorBlue,
		Plantable: true,
		Edible:    &entity.EdibleProperties{Poisonous: true},
		Sym:       config.CharBerry,
	}
	registry.Register(berryVariety)
	gameMap.SetVarieties(registry)

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
	expectedVariety := entity.GenerateVarietyID("berry", "", types.ColorBlue, types.PatternNone, types.TextureNone)
	if order.LockedVariety != expectedVariety {
		t.Errorf("Expected LockedVariety %q, got %q", expectedVariety, order.LockedVariety)
	}
}

func TestApplyIntent_Plant_PreservesEdibleOnSprout(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Create variety registry with mushroom variety
	registry := game.NewVarietyRegistry()
	mushVariety := &entity.ItemVariety{
		ID:       entity.GenerateVarietyID("mushroom", "", types.ColorPurple, types.PatternSpotted, types.TextureSlimy),
		ItemType: "mushroom",
		Color:    types.ColorPurple,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
		Edible:   &entity.EdibleProperties{Poisonous: true},
		Sym:      config.CharMushroom,
	}
	registry.Register(mushVariety)
	gameMap.SetVarieties(registry)

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
		ID:        entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
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

	// Create variety registry with berry variety
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:        entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType:  "berry",
		Color:     types.ColorRed,
		Plantable: true,
		Edible:    &entity.EdibleProperties{},
		Sym:       config.CharBerry,
	})
	gameMap.SetVarieties(registry)

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

	// Create variety registry with berry variety
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:        entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType:  "berry",
		Color:     types.ColorRed,
		Plantable: true,
		Edible:    &entity.EdibleProperties{},
		Sym:       config.CharBerry,
	})
	gameMap.SetVarieties(registry)

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
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
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

// =============================================================================
// ActionWaterGarden Tests
// =============================================================================

// Anchor test: Character at a dry tilled planted tile waters it — tile becomes wet,
// 1 water unit consumed, order completes when it was the last dry tile.
// Validates the user story: "water the closest dry tilled planted tile, watering uses
// 1 unit of water, once there are no remaining dry tilled planted tiles, order is complete."
func TestApplyIntent_WaterGarden_WatersTileConsumesWaterCompletesOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Vessel with 3 units of water in inventory
	waterVariety := &entity.ItemVariety{
		ID:       "liquid-water",
		ItemType: "liquid",
	}
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{Variety: waterVariety, Count: 3}},
		},
	}
	char.AddToInventory(vessel)

	// Single dry tilled planted tile (at character's position)
	tile := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(tile)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 5, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	// Create and assign order
	order := entity.NewOrder(1, "waterGarden", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action:     entity.ActionWaterGarden,
		Target:     tile,
		Dest:       tile,
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Run enough ticks to complete watering (ActionDurationShort = 0.83s)
	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Tile should be wet
	if !gameMap.IsManuallyWatered(tile) {
		t.Error("Expected tile to be manually watered")
	}

	// 1 water unit consumed (3 → 2)
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Expected vessel to still have water remaining")
	}
	if vessel.Container.Contents[0].Count != 2 {
		t.Errorf("Expected 2 water units remaining, got %d", vessel.Container.Contents[0].Count)
	}

	// Order should be completed (only tile is now wet)
	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}

	// Character should be unassigned
	if char.AssignedOrderID != 0 {
		t.Error("Expected char.AssignedOrderID to be cleared after completion")
	}

	// Intent should be cleared
	if char.Intent != nil {
		t.Error("Expected intent to be nil after watering completes")
	}
}

// When dry tiles remain after watering one, the order stays assigned (not completed).
// Next tick, CalculateIntent will re-enter findWaterGardenIntent to find the next tile.
func TestApplyIntent_WaterGarden_DoesNotCompleteWhenDryTilesRemain(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	waterVariety := &entity.ItemVariety{
		ID:       "liquid-water",
		ItemType: "liquid",
	}
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{Variety: waterVariety, Count: 3}},
		},
	}
	char.AddToInventory(vessel)

	// Two dry tilled planted tiles — character will water tile1, tile2 remains dry
	tile1 := types.Position{X: 5, Y: 5}
	tile2 := types.Position{X: 8, Y: 8}
	gameMap.SetTilled(tile1)
	gameMap.SetTilled(tile2)

	sprout1 := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 5, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	sprout2 := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 8, Y: 8, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout1)
	gameMap.AddItem(sprout2)

	order := entity.NewOrder(1, "waterGarden", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action:     entity.ActionWaterGarden,
		Target:     tile1,
		Dest:       tile1,
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 20; i++ {
		m.applyIntent(char, 0.1)
	}

	// Tile1 watered, tile2 still dry
	if !gameMap.IsManuallyWatered(tile1) {
		t.Error("Expected tile1 to be manually watered")
	}
	if gameMap.IsManuallyWatered(tile2) {
		t.Error("Expected tile2 to still be dry")
	}

	// Order should NOT be completed — dry tiles remain
	if order.Status == entity.OrderCompleted {
		t.Error("Order should not be completed when dry tilled planted tiles remain")
	}

	// Character should still be assigned to the order
	if char.AssignedOrderID != order.ID {
		t.Error("Expected char to still be assigned to the order")
	}

	// Intent should be cleared (ordered pattern — re-evaluated next tick)
	if char.Intent != nil {
		t.Error("Expected intent to be nil after watering one tile")
	}
}

// Anchor test: character has no vessel, ground vessel available, water source, dry tile.
// Full lifecycle: procure vessel → fill at water → water tile → order complete.
// Dry tile placed far from water (IsWet checks 8-directional adjacency to water).
// Target recalculated each tick via NextStepBFS, mirroring continueIntent in real game.
func TestApplyIntent_WaterGarden_FullCycleProcureFillWater(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Ground vessel at character's position — Phase 1 pickup is immediate
	groundVessel := &entity.Item{
		ItemType:   "vessel",
		Kind:       "hollow gourd",
		Name:       "Ground Vessel",
		BaseEntity: entity.BaseEntity{X: 5, Y: 5, Sym: 'U', EType: entity.TypeItem},
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	gameMap.AddItem(groundVessel)

	// Water source far from dry tile (IsWet returns true within 1 tile of water)
	waterPos := types.Position{X: 5, Y: 10}
	gameMap.AddWater(waterPos, game.WaterPond)

	// Single dry tilled planted tile at character's position (far from water)
	tile := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(tile)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: tile.X, Y: tile.Y, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	// Create and assign order
	order := entity.NewOrder(1, "waterGarden", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Phase 1 intent: targets ground vessel for procurement
	char.Intent = &entity.Intent{
		Action:     entity.ActionWaterGarden,
		Target:     char.Pos(),
		Dest:       groundVessel.Pos(),
		TargetItem: groundVessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 300; i++ {
		m.applyIntent(char, 0.1)
		// Recalculate Target toward Dest each tick (mirrors continueIntent in real game)
		if char.Intent != nil {
			cpos := char.Pos()
			nx, ny := system.NextStepBFS(cpos.X, cpos.Y, char.Intent.Dest.X, char.Intent.Dest.Y, gameMap)
			char.Intent.Target = types.Position{X: nx, Y: ny}
		}
		// After intent clears (ordered pattern), rebuild from findWaterGardenIntent
		if char.Intent == nil && char.AssignedOrderID != 0 {
			items := gameMap.Items()
			intent := system.FindWaterGardenIntentForTest(char, char.Pos(), items, order, actionLog, gameMap)
			if intent != nil {
				char.Intent = intent
			}
		}
	}

	// Tile should be watered
	if !gameMap.IsManuallyWatered(tile) {
		t.Errorf("Expected tile %v to be watered", tile)
	}

	// Vessel should be in inventory
	if char.GetCarriedVessel() == nil {
		t.Error("Expected vessel in inventory")
	}

	// Order should be completed
	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}
}

// Anchor test: character has vessel with 1 water unit, 2 dry tiles.
// Waters 1 tile, vessel empty, refills at water, waters remaining tile, order complete.
// Dry tiles placed far from water (IsWet checks 8-directional adjacency to water).
// Target recalculated each tick via NextStepBFS, mirroring continueIntent in real game.
func TestApplyIntent_WaterGarden_RefillCycle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Vessel with 1 unit of water — enough for 1 tile, then refill needed
	waterVariety := registry.VarietiesOfType("liquid")[0]
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{Variety: waterVariety, Count: 1}},
		},
	}
	char.AddToInventory(vessel)

	// Water source far from dry tiles (IsWet returns true within 1 tile of water)
	waterPos := types.Position{X: 5, Y: 10}
	gameMap.AddWater(waterPos, game.WaterPond)

	// 2 dry tilled planted tiles: at char position and adjacent (far from water)
	tiles := []types.Position{{X: 5, Y: 5}, {X: 4, Y: 5}}
	for _, tilePos := range tiles {
		gameMap.SetTilled(tilePos)
		sprout := &entity.Item{
			BaseEntity: entity.BaseEntity{X: tilePos.X, Y: tilePos.Y, Sym: 'v', EType: entity.TypeItem},
			ItemType:   "berry",
			Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
		}
		gameMap.AddItem(sprout)
	}

	// Create and assign order
	order := entity.NewOrder(1, "waterGarden", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Start with Phase 3 intent (has water, target tile at char position)
	char.Intent = &entity.Intent{
		Action:     entity.ActionWaterGarden,
		Target:     tiles[0],
		Dest:       tiles[0],
		TargetItem: vessel,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	for i := 0; i < 500; i++ {
		m.applyIntent(char, 0.1)
		// Recalculate Target toward Dest each tick (mirrors continueIntent in real game)
		if char.Intent != nil {
			cpos := char.Pos()
			nx, ny := system.NextStepBFS(cpos.X, cpos.Y, char.Intent.Dest.X, char.Intent.Dest.Y, gameMap)
			char.Intent.Target = types.Position{X: nx, Y: ny}
		}
		// After intent clears (ordered pattern), rebuild from findWaterGardenIntent
		if char.Intent == nil && char.AssignedOrderID != 0 {
			items := gameMap.Items()
			intent := system.FindWaterGardenIntentForTest(char, char.Pos(), items, order, actionLog, gameMap)
			if intent != nil {
				char.Intent = intent
			}
		}
	}

	// Both tiles should be watered
	for _, tilePos := range tiles {
		if !gameMap.IsManuallyWatered(tilePos) {
			t.Errorf("Expected tile %v to be watered", tilePos)
		}
	}

	// Order should be completed
	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}
}

// =============================================================================
// Displacement Tests (Step 2 Slice 9)
// =============================================================================

// Anchor: character blocked by another character sidesteps perpendicular instead of staying stuck
func TestDisplacement_Anchor_CharacterSidestepsAroundBlocker(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	blocker := entity.NewCharacter(2, 6, 5, "Blocker", "berries", types.ColorRed)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(blocker)

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	pos := char.Pos()
	if pos.X == 5 && pos.Y == 5 {
		t.Error("Character should not stay stuck at start after collision with blocker")
	}
	if pos.X == 6 && pos.Y == 5 {
		t.Error("Character should not occupy blocker's position")
	}
	// Should be at (5,4) or (5,6) — perpendicular to rightward movement
	if pos.X != 5 {
		t.Errorf("Should have moved perpendicular (same X=5), got (%d,%d)", pos.X, pos.Y)
	}
}

// Displacement sets 3-step state, first step taken immediately
func TestDisplacement_CharacterCollision_SetsDisplacementState(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	blocker := entity.NewCharacter(2, 6, 5, "Blocker", "berries", types.ColorRed)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(blocker)

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	if char.DisplacementStepsLeft != 2 {
		t.Errorf("Expected DisplacementStepsLeft=2 after first step, got %d", char.DisplacementStepsLeft)
	}
	if char.DisplacementDX == 0 && char.DisplacementDY == 0 {
		t.Error("Expected non-zero displacement direction after collision")
	}
}

// During displacement, character moves in displacement direction not BFS target
func TestDisplacement_MovesInDisplacementDirection(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	gameMap.AddCharacter(char)

	char.DisplacementStepsLeft = 2
	char.DisplacementDX = 0
	char.DisplacementDY = 1

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	pos := char.Pos()
	if pos.X != 5 || pos.Y != 6 {
		t.Errorf("Expected displacement move to (5,6), got (%d,%d)", pos.X, pos.Y)
	}
	if char.DisplacementStepsLeft != 1 {
		t.Errorf("Expected DisplacementStepsLeft=1, got %d", char.DisplacementStepsLeft)
	}
}

// After last displacement step, state is fully cleared
func TestDisplacement_ClearsAfterThreeSteps(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	gameMap.AddCharacter(char)

	char.DisplacementStepsLeft = 1
	char.DisplacementDX = 0
	char.DisplacementDY = 1

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	if char.DisplacementStepsLeft != 0 {
		t.Errorf("Expected DisplacementStepsLeft=0 after last step, got %d", char.DisplacementStepsLeft)
	}
	if char.DisplacementDX != 0 || char.DisplacementDY != 0 {
		t.Error("Expected displacement direction cleared after last step")
	}
}

// If primary displacement direction is blocked, tries opposite perpendicular
func TestDisplacement_PrimaryDirBlocked_TriesOtherPerp(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	blockDisp := entity.NewCharacter(2, 5, 6, "BlockDisp", "berries", types.ColorRed)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(blockDisp)

	char.DisplacementStepsLeft = 2
	char.DisplacementDX = 0
	char.DisplacementDY = 1

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	pos := char.Pos()
	if pos.X != 5 || pos.Y != 4 {
		t.Errorf("Expected (5,4) when +Y blocked, got (%d,%d)", pos.X, pos.Y)
	}
	if char.DisplacementDY != -1 {
		t.Errorf("Expected DisplacementDY=-1 after switching, got %d", char.DisplacementDY)
	}
}

// If both perpendicular directions blocked, displacement clears
func TestDisplacement_BothPerpsBlocked_ClearsDisplacement(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	blockUp := entity.NewCharacter(2, 5, 6, "BlockUp", "berries", types.ColorRed)
	blockDown := entity.NewCharacter(3, 5, 4, "BlockDown", "berries", types.ColorRed)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(blockUp)
	gameMap.AddCharacter(blockDown)

	char.DisplacementStepsLeft = 2
	char.DisplacementDX = 0
	char.DisplacementDY = 1

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	if char.DisplacementStepsLeft != 0 {
		t.Errorf("Expected DisplacementStepsLeft=0, got %d", char.DisplacementStepsLeft)
	}
	pos := char.Pos()
	if pos.X != 5 || pos.Y != 5 {
		t.Errorf("Expected stuck at (5,5), got (%d,%d)", pos.X, pos.Y)
	}
}

// Water collision does NOT trigger displacement — only character collision does
func TestDisplacement_WaterCollision_NoDisplacement(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	gameMap.AddWater(types.Position{X: 6, Y: 5}, game.WaterPond)
	gameMap.AddCharacter(char)

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	if char.DisplacementStepsLeft != 0 {
		t.Errorf("Expected no displacement for water collision, got %d", char.DisplacementStepsLeft)
	}
}

// Intent is preserved through displacement
func TestDisplacement_IntentPreservedThroughDisplacement(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	blocker := entity.NewCharacter(2, 6, 5, "Blocker", "berries", types.ColorRed)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(blocker)

	targetItem := entity.NewBerry(10, 5, types.ColorRed, false, false)
	gameMap.AddItem(targetItem)

	char.Intent = &entity.Intent{
		Action:     entity.ActionMove,
		Target:     types.Position{X: 6, Y: 5},
		Dest:       types.Position{X: 10, Y: 5},
		TargetItem: targetItem,
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	if char.Intent == nil {
		t.Fatal("Intent should not be nil after displacement")
	}
	if char.Intent.TargetItem != targetItem {
		t.Error("TargetItem should be preserved through displacement")
	}
	if char.Intent.Dest.X != 10 || char.Intent.Dest.Y != 5 {
		t.Error("Dest should be preserved through displacement")
	}
}

// Sticky BFS flag clears when displacement is initiated (character collision
// resets to greedy-first mode for the new direction after sidestepping).
func TestDisplacement_ClearsUsingBFS(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Mover", "berries", types.ColorRed)
	blocker := entity.NewCharacter(2, 6, 5, "Blocker", "berries", types.ColorRed)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(blocker)

	// Simulate character was using BFS to navigate around a pond
	char.UsingBFS = true

	char.Intent = &entity.Intent{
		Action: entity.ActionMove,
		Target: types.Position{X: 6, Y: 5},
		Dest:   types.Position{X: 10, Y: 5},
	}
	char.SpeedAccumulator = 10.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.1)

	// Displacement should have fired (character moved perpendicular)
	// and UsingBFS should be cleared
	if char.UsingBFS {
		t.Error("Expected UsingBFS=false after displacement initiation")
	}
}

// =============================================================================
// stepForward Tests
// =============================================================================

func TestStepForward_AdvancesGameTime(t *testing.T) {
	t.Parallel()

	// Setup: a model with some elapsed game time
	m := Model{
		phase:           phasePlaying,
		paused:          true,
		actionLog:       system.NewActionLog(100),
		elapsedGameTime: 100.0,
		gameMap:         game.NewMap(20, 20),
	}
	m.actionLog.SetGameTime(100.0)

	// Act: step forward three times
	m.stepForward()
	m.stepForward()
	m.stepForward()

	// Assert: game time should have advanced by 3 ticks
	delta := config.UpdateInterval.Seconds()
	if m.elapsedGameTime <= 100.0 {
		t.Errorf("elapsedGameTime should have advanced from 100.0, got %.6f", m.elapsedGameTime)
	}
	if m.elapsedGameTime < 100.0+delta*2.9 || m.elapsedGameTime > 100.0+delta*3.1 {
		t.Errorf("elapsedGameTime should be ~%.4f after 3 steps, got %.6f", 100.0+delta*3, m.elapsedGameTime)
	}

	// Assert: action log game time should match model
	if m.actionLog.GameTime() != m.elapsedGameTime {
		t.Errorf("actionLog GameTime (%.6f) should match elapsedGameTime (%.6f)",
			m.actionLog.GameTime(), m.elapsedGameTime)
	}
}

func TestStepForward_EventsFromDifferentTicksHaveDistinctTimes(t *testing.T) {
	t.Parallel()

	// Setup: model with no characters (we'll add events manually to control timing)
	actionLog := system.NewActionLog(100)
	actionLog.SetGameTime(100.0)

	m := Model{
		phase:           phasePlaying,
		paused:          true,
		actionLog:       actionLog,
		elapsedGameTime: 100.0,
		gameMap:         game.NewMap(20, 20),
	}

	// Step tick 1, then add events (higher CharID first to test ordering)
	m.stepForward()
	m.actionLog.Add(2, "Bob", "test", "Tick 1 from Bob")
	m.actionLog.Add(1, "Alice", "test", "Tick 1 from Alice")

	// Step tick 2, then add events
	m.stepForward()
	m.actionLog.Add(2, "Bob", "test", "Tick 2 from Bob")
	m.actionLog.Add(1, "Alice", "test", "Tick 2 from Alice")

	// Assert: AllEvents sorts by time, so tick 1 events appear before tick 2
	events := actionLog.AllEvents(100)

	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}

	// Tick 1 events should come before tick 2 events
	if events[0].GameTime >= events[2].GameTime {
		t.Errorf("Tick 1 events (time=%.4f) should have earlier GameTime than tick 2 (time=%.4f)",
			events[0].GameTime, events[2].GameTime)
	}
}

// =============================================================================
// ActionLook Handler Tests
// =============================================================================

// Anchor: character adjacent to an item looks at it and forms a memory of having looked
func TestApplyIntent_Look_CompletesAndSetsLastLooked(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	item := entity.NewBerry(6, 5, types.ColorBlue, false, false)
	gameMap.AddItem(item)

	char.Intent = &entity.Intent{
		Action:     entity.ActionLook,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: item,
	}

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}

	// Tick enough to exceed LookDuration (4.0s)
	for i := 0; i < 50; i++ {
		m.applyIntent(char, 0.1)
	}

	// Assert: look completed — HasLastLooked set, intent cleared, idle cooldown set
	if !char.HasLastLooked {
		t.Error("Expected HasLastLooked to be true after completing look")
	}
	if char.LastLookedX != 6 || char.LastLookedY != 5 {
		t.Errorf("Expected LastLooked at (6,5), got (%d,%d)", char.LastLookedX, char.LastLookedY)
	}
	if char.Intent != nil {
		t.Error("Expected intent to be nil after completing look")
	}
	if char.IdleCooldown == 0 {
		t.Error("Expected IdleCooldown to be set after completing look")
	}
}

// Looking requires the full duration — partial ticks don't complete
func TestApplyIntent_Look_RequiresDuration(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	item := entity.NewBerry(6, 5, types.ColorBlue, false, false)
	gameMap.AddItem(item)

	char.Intent = &entity.Intent{
		Action:     entity.ActionLook,
		Target:     types.Position{X: 5, Y: 5},
		TargetItem: item,
	}

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}

	// One small tick — not enough to complete
	m.applyIntent(char, 0.1)

	if char.ActionProgress == 0 {
		t.Error("Expected ActionProgress to increase")
	}
	if char.HasLastLooked {
		t.Error("Expected HasLastLooked to remain false before look completes")
	}
	if char.Intent == nil {
		t.Error("Expected intent to still be set — look not yet complete")
	}
}

// =============================================================================
// ActionTalk Handler Tests
// =============================================================================

// Anchor: two characters have a conversation — initiator starts talking, then
// after the full talk duration, both characters stop talking and go idle
func TestApplyIntent_Talk_CompletesConversationCycle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	initiator := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	target := entity.NewCharacter(2, 6, 5, "Bob", "berry", types.ColorBlue)
	gameMap.AddCharacter(initiator)
	gameMap.AddCharacter(target)

	initiator.Intent = &entity.Intent{
		Action:          entity.ActionTalk,
		Target:          types.Position{X: 5, Y: 5},
		TargetCharacter: target,
	}

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}

	// First tick — should start talking
	m.applyIntent(initiator, 0.1)

	if initiator.TalkingWith != target {
		t.Error("Expected initiator.TalkingWith to be target after first tick")
	}
	if target.TalkingWith != initiator {
		t.Error("Expected target.TalkingWith to be initiator after first tick")
	}

	// Tick enough to complete TalkDuration (5.0s)
	for i := 0; i < 60; i++ {
		m.applyIntent(initiator, 0.1)
	}

	// Assert: conversation over — both characters have cleared talk state
	if initiator.TalkingWith != nil {
		t.Error("Expected initiator.TalkingWith to be nil after talk completes")
	}
	if target.TalkingWith != nil {
		t.Error("Expected target.TalkingWith to be nil after talk completes")
	}
	if initiator.Intent != nil {
		t.Error("Expected initiator intent to be nil after talk completes")
	}
}

// Talk with nil TargetCharacter returns immediately without panic
func TestApplyIntent_Talk_NilTarget_NoPanic(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	char.Intent = &entity.Intent{
		Action:          entity.ActionTalk,
		Target:          types.Position{X: 5, Y: 5},
		TargetCharacter: nil,
	}

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}

	// Should not panic
	m.applyIntent(char, 0.1)

	// Intent unchanged (handler returns early)
	if char.TalkingWith != nil {
		t.Error("Expected TalkingWith to remain nil with nil target")
	}
}

// =============================================================================
// ActionHelpFeed Handler Tests
// =============================================================================

// Anchor: helper carrying food walks to hungry character and drops it adjacent
func TestApplyIntent_HelpFeed_DeliversFoodToNeeder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 7, 5, "Hungry", "berry", types.ColorBlue)
	needer.Hunger = 95 // Crisis-level hunger
	gameMap.AddCharacter(helper)
	gameMap.AddCharacter(needer)

	// Food already in inventory (delivery phase)
	food := entity.NewBerry(0, 0, types.ColorRed, false, false)
	food.Plant = nil // Loose food, not growing
	helper.AddToInventory(food)

	npos := needer.Pos()
	helper.Intent = &entity.Intent{
		Action:          entity.ActionHelpFeed,
		Target:          types.Position{X: 6, Y: 5}, // Next step toward needer
		Dest:            npos,
		TargetItem:      food,
		TargetCharacter: needer,
	}
	helper.CurrentActivity = "Bringing food to Hungry"
	helper.SpeedAccumulator = 100.0 // Ensure movement happens

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}

	// Tick enough for helper to walk to needer and deliver
	for i := 0; i < 50; i++ {
		m.applyIntent(helper, 0.1)
	}

	// Assert: food dropped on map near needer, helper goes idle
	if len(helper.Inventory) != 0 {
		t.Error("Expected helper inventory to be empty after delivery")
	}
	if helper.Intent != nil {
		t.Error("Expected helper intent to be nil after delivery")
	}
	if !gameMap.HasItemOnMap(food) {
		t.Error("Expected food to be placed on the map near needer")
	}
	// Food should be cardinally adjacent to needer
	fpos := food.Pos()
	if !fpos.IsCardinallyAdjacentTo(npos) && fpos != npos {
		t.Errorf("Expected food at cardinal-adjacent to needer (%d,%d), got (%d,%d)",
			npos.X, npos.Y, fpos.X, fpos.Y)
	}
	// Needer's intent should be cleared (signaled to re-evaluate)
	if needer.Intent != nil {
		t.Error("Expected needer intent to be nil after food delivered (re-evaluate signal)")
	}
}

// Helper abandons delivery when needer dies
func TestApplyIntent_HelpFeed_AbandonWhenNeederDead(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 10, 5, "Hungry", "berry", types.ColorBlue)
	needer.IsDead = true
	gameMap.AddCharacter(helper)
	gameMap.AddCharacter(needer)

	food := entity.NewBerry(0, 0, types.ColorRed, false, false)
	food.Plant = nil
	helper.AddToInventory(food)

	helper.Intent = &entity.Intent{
		Action:          entity.ActionHelpFeed,
		Target:          types.Position{X: 6, Y: 5},
		Dest:            needer.Pos(),
		TargetItem:      food,
		TargetCharacter: needer,
	}

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(helper, 0.1)

	// Assert: helper drops food and goes idle
	if helper.Intent != nil {
		t.Error("Expected helper intent to be nil when needer is dead")
	}
	if gameMap.HasItemOnMap(food) {
		// Food should be dropped at helper's position
	}
}

// =============================================================================
// ActionHelpWater Handler Tests
// =============================================================================

// Anchor: helper carrying a filled vessel walks to thirsty character and drops it adjacent
func TestApplyIntent_HelpWater_DeliversWaterToNeeder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 7, 5, "Thirsty", "berry", types.ColorBlue)
	needer.Thirst = 95
	gameMap.AddCharacter(helper)
	gameMap.AddCharacter(needer)

	// Vessel with water already in inventory (delivery phase)
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Gourd",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: &entity.ItemVariety{ItemType: "liquid", Kind: "water"}, Count: 1},
			},
		},
	}
	helper.AddToInventory(vessel)

	npos := needer.Pos()
	helper.Intent = &entity.Intent{
		Action:          entity.ActionHelpWater,
		Target:          types.Position{X: 6, Y: 5},
		Dest:            npos,
		TargetItem:      vessel,
		TargetCharacter: needer,
	}
	helper.CurrentActivity = "Bringing water to Thirsty"
	helper.SpeedAccumulator = 100.0

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}

	// Tick enough for helper to reach needer and deliver
	for i := 0; i < 50; i++ {
		m.applyIntent(helper, 0.1)
	}

	// Assert: vessel dropped on map near needer, helper goes idle
	if len(helper.Inventory) != 0 {
		t.Error("Expected helper inventory to be empty after delivery")
	}
	if helper.Intent != nil {
		t.Error("Expected helper intent to be nil after delivery")
	}
	if !gameMap.HasItemOnMap(vessel) {
		t.Error("Expected vessel to be placed on the map near needer")
	}
	// Vessel should be cardinally adjacent to needer
	vpos := vessel.Pos()
	if !vpos.IsCardinallyAdjacentTo(npos) && vpos != npos {
		t.Errorf("Expected vessel at cardinal-adjacent to needer (%d,%d), got (%d,%d)",
			npos.X, npos.Y, vpos.X, vpos.Y)
	}
	// Needer's intent should be cleared
	if needer.Intent != nil {
		t.Error("Expected needer intent to be nil after water delivered")
	}
}

// Helper abandons water delivery when needer dies
func TestApplyIntent_HelpWater_AbandonWhenNeederDead(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	helper := entity.NewCharacter(1, 5, 5, "Helper", "berry", types.ColorRed)
	needer := entity.NewCharacter(2, 10, 5, "Thirsty", "berry", types.ColorBlue)
	needer.IsDead = true
	gameMap.AddCharacter(helper)
	gameMap.AddCharacter(needer)

	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Gourd",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: &entity.ItemVariety{ItemType: "liquid", Kind: "water"}, Count: 1},
			},
		},
	}
	helper.AddToInventory(vessel)

	helper.Intent = &entity.Intent{
		Action:          entity.ActionHelpWater,
		Target:          types.Position{X: 6, Y: 5},
		Dest:            needer.Pos(),
		TargetItem:      vessel,
		TargetCharacter: needer,
	}

	m := &Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(helper, 0.1)

	// Assert: helper drops vessel and goes idle
	if helper.Intent != nil {
		t.Error("Expected helper intent to be nil when needer is dead")
	}
}

// =============================================================================
// Extract order tests
// =============================================================================

func TestApplyExtract_FullFlow_ExtractsSeedToVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Register seed variety
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:        "flower seed-yellow",
		ItemType:  "seed",
		Kind:      "flower seed",
		Color:     types.ColorYellow,
		Plantable: true,
		Sym:       '·',
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
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

	// Place a flower on character's tile
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 5, Y: 5, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	// Assign an extract order to the character
	order := entity.NewOrder(1, "extract", "flower")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Set extract intent targeting the flower
	char.Intent = &entity.Intent{
		Action:     entity.ActionExtract,
		Target:     types.Position{X: 5, Y: 5},
		Dest:       types.Position{X: 5, Y: 5},
		TargetItem: flower,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Apply intent with enough delta to complete extraction
	m.applyIntent(char, config.ActionDurationShort+0.1)

	// Seed should be in vessel
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Expected seed in vessel after extraction")
	}
	stack := vessel.Container.Contents[0]
	if stack.Variety.Kind != "flower seed" {
		t.Errorf("Expected flower seed in vessel, got kind %q", stack.Variety.Kind)
	}

	// Plant's SeedTimer should be set
	if flower.Plant.SeedTimer <= 0 {
		t.Error("Expected positive SeedTimer on plant after extraction")
	}

	// Intent should be cleared
	if char.Intent != nil {
		t.Error("Expected intent cleared after extraction")
	}

	// Flower should still be on the map (not destroyed)
	if !gameMap.HasItemOnMap(flower) {
		t.Error("Expected flower to still be on map after extraction")
	}

	// Order should be locked to yellow flower variety
	updatedOrder := m.findOrderByID(char.AssignedOrderID)
	if updatedOrder == nil {
		t.Fatal("Expected character to have assigned order")
	}
	expectedVariety := entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")
	if updatedOrder.LockedVariety != expectedVariety {
		t.Errorf("Expected order locked to %q, got %q", expectedVariety, updatedOrder.LockedVariety)
	}
}

func TestApplyExtract_CompletesOrderWhenInventoryFullNoVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Register seed variety
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:        "flower seed-yellow",
		ItemType:  "seed",
		Kind:      "flower seed",
		Color:     types.ColorYellow,
		Plantable: true,
		Sym:       '·',
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	// Fill first inventory slot with existing seed — no vessel
	existingSeed := entity.NewSeed(0, 0, "flower", "flower-yellow", "", types.ColorYellow, "", "")
	char.AddToInventory(existingSeed)

	// Place flower on character's tile
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 5, Y: 5, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	// Assign extract order, already locked
	order := entity.NewOrder(1, "extract", "flower")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")
	char.AssignedOrderID = order.ID

	// Set extract intent at the flower
	char.Intent = &entity.Intent{
		Action:     entity.ActionExtract,
		Target:     types.Position{X: 5, Y: 5},
		Dest:       types.Position{X: 5, Y: 5},
		TargetItem: flower,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	// Extract completes — seed goes to inventory slot 2, filling it
	m.applyIntent(char, config.ActionDurationShort+0.1)

	// Inventory should be full (2 seeds)
	if char.HasInventorySpace() {
		t.Error("Expected inventory to be full after extraction")
	}

	// Order should be completed (inline check: inventory full, no vessel)
	if order.Status != entity.OrderCompleted {
		t.Errorf("Expected order completed, got status %v", order.Status)
	}

	// Character should be unassigned
	if char.AssignedOrderID != 0 {
		t.Error("Expected character unassigned from order after completion")
	}
}

func TestApplyExtract_ContinuesWhenVesselHasSpace(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:        "flower seed-yellow",
		ItemType:  "seed",
		Kind:      "flower seed",
		Color:     types.ColorYellow,
		Plantable: true,
		Sym:       '·',
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	// Give character a vessel with space
	vessel := &entity.Item{
		ItemType: "vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	char.AddToInventory(vessel)

	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 5, Y: 5, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")
	char.AssignedOrderID = order.ID

	char.Intent = &entity.Intent{
		Action:     entity.ActionExtract,
		Target:     types.Position{X: 5, Y: 5},
		Dest:       types.Position{X: 5, Y: 5},
		TargetItem: flower,
	}

	actionLog := system.NewActionLog(100)
	m := Model{
		gameMap:   gameMap,
		actionLog: actionLog,
		orders:    []*entity.Order{order},
	}

	m.applyIntent(char, config.ActionDurationShort+0.1)

	// Seed should be in vessel
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Expected seed in vessel after extraction")
	}

	// Order should NOT be completed — vessel still has space
	if order.Status == entity.OrderCompleted {
		t.Error("Expected order to continue when vessel has space")
	}
}

// =============================================================================
// Step 3b: applyDig handler
// =============================================================================

func TestApplyDig_WalksToClayTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	clayPos := types.Position{X: 8, Y: 5}
	gameMap.SetClay(clayPos)

	// Intent: walking toward clay (not yet there)
	char.Intent = &entity.Intent{
		Action: entity.ActionDig,
		Target: types.Position{X: 6, Y: 5},
		Dest:   clayPos,
	}

	m := Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, 0.15)

	// Character should have moved toward clay
	newPos := char.Pos()
	if newPos.X == 5 && newPos.Y == 5 {
		t.Error("Expected character to move toward clay tile, but position unchanged")
	}
	// Should not have dug anything (progress requires being at Dest)
	if char.HasInventorySpace() && countItemType(char, "clay") > 0 {
		t.Error("Expected no clay in inventory while still walking")
	}
}

func TestApplyDig_CreatesClayInInventory(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// Clay tile at character's position (already there)
	clayPos := types.Position{X: 5, Y: 5}
	gameMap.SetClay(clayPos)

	char.Intent = &entity.Intent{
		Action: entity.ActionDig,
		Target: clayPos,
		Dest:   clayPos,
	}

	m := Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, config.ActionDurationMedium+0.1)

	// Clay should be in inventory
	clayCount := countItemType(char, "clay")
	if clayCount != 1 {
		t.Errorf("Expected 1 clay in inventory after digging, got %d", clayCount)
	}
}

func TestApplyDig_ClearsIntent(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "TestChar", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	clayPos := types.Position{X: 5, Y: 5}
	gameMap.SetClay(clayPos)

	char.Intent = &entity.Intent{
		Action: entity.ActionDig,
		Target: clayPos,
		Dest:   clayPos,
	}

	m := Model{gameMap: gameMap, actionLog: system.NewActionLog(100)}
	m.applyIntent(char, config.ActionDurationMedium+0.1)

	// Intent should be cleared (ordered action one-tick-idle pattern)
	if char.Intent != nil {
		t.Error("Expected intent cleared after dig completes")
	}
}

func TestPushLooseItemsAside_MovesNonGrowingItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Place a loose seed on a tilled tile
	tilePos := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(tilePos)
	looseSeed := entity.NewSeed(tilePos.X, tilePos.Y, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(looseSeed)

	charPos := types.Position{X: 5, Y: 4} // Character is north of tile

	pushLooseItemsAside(tilePos, charPos, gameMap)

	// Seed should have been moved off the tile
	seedPos := looseSeed.Pos()
	if seedPos == tilePos {
		t.Error("Expected loose seed to be pushed off the tile")
	}
	// Should be cardinal-adjacent to original position
	dist := tilePos.DistanceTo(seedPos)
	if dist != 1 {
		t.Errorf("Expected seed to be 1 tile away, got distance %d (at %v)", dist, seedPos)
	}
}

func TestPushLooseItemsAside_LeavesGrowingPlantsInPlace(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Place a growing berry on a tile
	tilePos := types.Position{X: 5, Y: 5}
	gameMap.SetTilled(tilePos)
	growingBerry := entity.NewBerry(tilePos.X, tilePos.Y, types.ColorRed, false, false)
	gameMap.AddItem(growingBerry)

	charPos := types.Position{X: 5, Y: 4}

	pushLooseItemsAside(tilePos, charPos, gameMap)

	// Growing berry should NOT have moved
	berryPos := growingBerry.Pos()
	if berryPos != tilePos {
		t.Errorf("Expected growing berry to stay at %v, but moved to %v", tilePos, berryPos)
	}
}

// countItemType counts how many items of the given type are in a character's inventory
func countItemType(char *entity.Character, itemType string) int {
	count := 0
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == itemType {
			count++
		}
	}
	return count
}
