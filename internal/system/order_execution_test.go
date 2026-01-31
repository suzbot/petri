package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// selectOrderActivity
// =============================================================================

func TestSelectOrderActivity_AssignsOpenOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"} // knows harvest
	gameMap.AddCharacter(char)

	// Add a berry to harvest
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Create open order
	order := entity.NewOrder(1, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	if intent == nil {
		t.Fatal("Expected intent to be returned, got nil")
	}

	if order.Status != entity.OrderAssigned {
		t.Errorf("Order status: got %s, want Assigned", order.Status)
	}

	if order.AssignedTo != char.ID {
		t.Errorf("Order.AssignedTo: got %d, want %d", order.AssignedTo, char.ID)
	}

	if char.AssignedOrderID != order.ID {
		t.Errorf("char.AssignedOrderID: got %d, want %d", char.AssignedOrderID, order.ID)
	}

	if intent.TargetItem != berry {
		t.Errorf("Intent.TargetItem: got %v, want %v", intent.TargetItem, berry)
	}

	if intent.Action != entity.ActionPickup {
		t.Errorf("Intent.Action: got %v, want ActionPickup", intent.Action)
	}
}

func TestSelectOrderActivity_ResumesAssignedOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	// Add a berry to harvest
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Create already-assigned order
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	if intent == nil {
		t.Fatal("Expected intent to be returned for resume, got nil")
	}

	if intent.TargetItem != berry {
		t.Errorf("Intent.TargetItem: got %v, want %v", intent.TargetItem, berry)
	}
}

func TestSelectOrderActivity_ResumesPausedOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	// Add a berry to harvest
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Create paused order
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderPaused
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	if intent == nil {
		t.Fatal("Expected intent to be returned for paused resume, got nil")
	}

	// Order should be set back to Assigned
	if order.Status != entity.OrderAssigned {
		t.Errorf("Order status: got %s, want Assigned", order.Status)
	}
}

func TestSelectOrderActivity_RequiresKnowHow(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	// char does NOT know harvest
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	if intent != nil {
		t.Error("Expected nil intent for character without know-how")
	}

	if order.Status != entity.OrderOpen {
		t.Errorf("Order should remain Open, got %s", order.Status)
	}
}

func TestSelectOrderActivity_FullInventoryCanTakeNew(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Carrying = entity.NewBerry(0, 0, types.ColorRed, false, false) // inventory full
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	// Characters can now take orders with full inventory - will drop during execution
	if intent == nil {
		t.Error("Expected intent - characters can take orders with full inventory")
	}

	if order.Status != entity.OrderAssigned {
		t.Errorf("Order should be Assigned, got %s", order.Status)
	}
}

func TestSelectOrderActivity_AbandonsWhenNoItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.AssignedOrderID = 1
	gameMap.AddCharacter(char)

	// NO berries on map

	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	if intent != nil {
		t.Error("Expected nil intent when no items to harvest")
	}

	// Order should be abandoned (status back to Open)
	if order.Status != entity.OrderOpen {
		t.Errorf("Order should be abandoned (Open), got %s", order.Status)
	}

	if char.AssignedOrderID != 0 {
		t.Errorf("char.AssignedOrderID should be cleared, got %d", char.AssignedOrderID)
	}
}

// =============================================================================
// Order ID validation
// =============================================================================

// TestSelectOrderActivity_OrderIDMustBeNonZero verifies that order IDs must be
// non-zero for proper assignment tracking. This is a regression test for a bug
// where order ID 0 caused char.AssignedOrderID = 0, which looked like "no order".
func TestSelectOrderActivity_OrderIDMustBeNonZero(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Create order with ID 0 - this is the problematic case
	order := entity.NewOrder(0, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	_ = selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	// After assignment, char.AssignedOrderID = order.ID = 0
	// This looks like "no order assigned" because we check AssignedOrderID != 0
	// This test documents why the UI MUST start order IDs at 1, not 0

	// Verify the bidirectional relationship is broken with ID 0:
	// order thinks it's assigned to the character...
	if order.AssignedTo != char.ID {
		t.Errorf("order.AssignedTo should be %d, got %d", char.ID, order.AssignedTo)
	}
	// ...but the character appears to have no order (because ID is 0)
	// This would cause the character to not resume the order on subsequent ticks
	hasOrder := char.AssignedOrderID != 0
	if hasOrder {
		t.Error("With order ID 0, char.AssignedOrderID != 0 check incorrectly shows an order")
	}
}

// TestSelectOrderActivity_ValidatesAssignmentBidirectional ensures that after
// assignment, both order.AssignedTo and char.AssignedOrderID are properly set
// and can be used to look up the relationship in either direction.
func TestSelectOrderActivity_ValidatesAssignmentBidirectional(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Use a valid order ID (non-zero)
	order := entity.NewOrder(42, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	// Verify bidirectional relationship:
	// 1. Order knows which character has it
	if order.AssignedTo != char.ID {
		t.Errorf("order.AssignedTo: got %d, want %d", order.AssignedTo, char.ID)
	}

	// 2. Character knows which order they have
	if char.AssignedOrderID != order.ID {
		t.Errorf("char.AssignedOrderID: got %d, want %d", char.AssignedOrderID, order.ID)
	}

	// 3. The "has order" check works correctly
	hasOrder := char.AssignedOrderID != 0
	if !hasOrder {
		t.Error("char.AssignedOrderID != 0 should be true for assigned order")
	}

	// 4. Can look up order from character's AssignedOrderID
	foundOrder := findOrderByID(orders, char.AssignedOrderID)
	if foundOrder != order {
		t.Error("Should be able to find order using char.AssignedOrderID")
	}
}

// =============================================================================
// findHarvestIntent
// =============================================================================

func TestFindHarvestIntent_FindsNearestItem(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Add two berries - one near, one far
	nearBerry := entity.NewBerry(6, 5, types.ColorRed, false, false)
	farBerry := entity.NewBerry(9, 9, types.ColorBlue, false, false)
	gameMap.AddItem(nearBerry)
	gameMap.AddItem(farBerry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	if intent.TargetItem != nearBerry {
		t.Errorf("Should target nearest berry, got %v", intent.TargetItem)
	}
}

func TestFindHarvestIntent_MatchesTargetType(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Add mushroom (close) and berry (far)
	mushroom := entity.NewMushroom(6, 5, types.ColorBrown, "", "", false, false)
	berry := entity.NewBerry(9, 9, types.ColorRed, false, false)
	gameMap.AddItem(mushroom)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	if intent.TargetItem != berry {
		t.Errorf("Should target berry (matching type), not mushroom. Got: %v", intent.TargetItem)
	}
}

func TestFindHarvestIntent_ReturnsNilWhenNoMatchingItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Only mushrooms, no berries
	mushroom := entity.NewMushroom(6, 5, types.ColorBrown, "", "", false, false)
	gameMap.AddItem(mushroom)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no matching items")
	}
}

// =============================================================================
// findHarvestIntent with Vessel Logic (4d/4e)
// =============================================================================

func TestFindHarvestIntent_LooksForVesselFirst(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Set up variety registry so FindAvailableVessel works
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Berry to harvest
	berry := entity.NewBerry(8, 8, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Empty vessel closer than berry
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	vessel.SetPos(types.Position{X: 6, Y: 5})
	gameMap.AddItem(vessel)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	// Should target vessel first, not berry
	if intent.TargetItem != vessel {
		t.Errorf("Should target vessel first. Got: %v", intent.TargetItem)
	}
}

func TestFindHarvestIntent_DropsIncompatibleVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Set up variety registry
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	})
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("mushroom", types.ColorBrown, types.PatternSpotted, types.TextureSlimy),
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
		Edible: &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Character is carrying vessel with mushrooms
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{
				Variety: registry.Get(entity.GenerateVarietyID("mushroom", types.ColorBrown, types.PatternSpotted, types.TextureSlimy)),
				Count:   5,
			}},
		},
	}
	char.Carrying = vessel

	// Berry to harvest (incompatible with vessel contents)
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	// This should drop the vessel
	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	// Vessel should have been dropped
	if char.Carrying != nil {
		t.Error("Vessel should have been dropped due to variety mismatch")
	}

	// Vessel should be on the map at character's position
	droppedVessel := gameMap.ItemAt(5, 5)
	if droppedVessel == nil || droppedVessel.Container == nil {
		t.Error("Dropped vessel should be on map at character position")
	}
}

func TestFindHarvestIntent_UsesCompatibleVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Set up variety registry
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible: &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Character is carrying vessel with red berries (compatible)
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{
				Variety: registry.Get(entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone)),
				Count:   5,
			}},
		},
	}
	char.Carrying = vessel

	// Red berry to harvest (compatible with vessel contents)
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	// Vessel should NOT be dropped (it's compatible)
	if char.Carrying != vessel {
		t.Error("Compatible vessel should not be dropped")
	}

	// Should target the berry
	if intent.TargetItem != berry {
		t.Errorf("Should target berry. Got: %v", intent.TargetItem)
	}
}

// =============================================================================
// Integration tests with CalculateIntent
// =============================================================================

func TestCalculateIntent_AssignsOrderWhenIdle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	// Ensure character has no needs (all stats healthy)
	char.Hunger = 0
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100
	char.IdleCooldown = 0 // No cooldown
	gameMap.AddCharacter(char)

	// Add a berry to harvest
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Create open order
	order := entity.NewOrder(1, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := CalculateIntent(char, items, gameMap, nil, orders)

	if intent == nil {
		t.Fatal("Expected intent to be returned, got nil")
	}

	if order.Status != entity.OrderAssigned {
		t.Errorf("Order status: got %s, want Assigned", order.Status)
	}

	if char.AssignedOrderID != order.ID {
		t.Errorf("char.AssignedOrderID: got %d, want %d", char.AssignedOrderID, order.ID)
	}

	if intent.Action != entity.ActionPickup {
		t.Errorf("Intent.Action: got %v, want ActionPickup", intent.Action)
	}

	if intent.TargetItem != berry {
		t.Errorf("Intent.TargetItem: got %v, want %v", intent.TargetItem, berry)
	}
}

func TestCalculateIntent_ContinuesOrderWork(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Hunger = 0
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100
	gameMap.AddCharacter(char)

	// Add a berry to harvest (not adjacent)
	berry := entity.NewBerry(8, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Set up order as already assigned
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	orders := []*entity.Order{order}

	// Set up existing intent (character is working on order)
	char.Intent = &entity.Intent{
		Target:     types.Position{X: 6, Y: 5}, // Moving toward berry
		Action:     entity.ActionPickup,
		TargetItem: berry,
	}

	items := gameMap.Items()
	intent := CalculateIntent(char, items, gameMap, nil, orders)

	if intent == nil {
		t.Fatal("Expected intent to be continued, got nil")
	}

	if intent.Action != entity.ActionPickup {
		t.Errorf("Intent.Action: got %v, want ActionPickup (continued)", intent.Action)
	}

	if intent.TargetItem != berry {
		t.Errorf("Intent.TargetItem: got %v, want %v", intent.TargetItem, berry)
	}
}

func TestCalculateIntent_PausesOrderWhenModerateNeed(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Hunger = 80 // Moderate hunger - should interrupt order
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100
	gameMap.AddCharacter(char)

	// Add a berry to harvest
	berry := entity.NewBerry(8, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Add food for character to eat
	food := entity.NewBerry(6, 5, types.ColorRed, false, false)
	gameMap.AddItem(food)

	// Set up order as already assigned
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	orders := []*entity.Order{order}

	// Set up existing intent (character was working on order)
	char.Intent = &entity.Intent{
		Target:     types.Position{X: 7, Y: 5},
		Action:     entity.ActionPickup,
		TargetItem: berry,
	}

	items := gameMap.Items()
	intent := CalculateIntent(char, items, gameMap, nil, orders)

	if intent == nil {
		t.Fatal("Expected food-seeking intent, got nil")
	}

	// Character should seek food, not continue order
	if intent.DrivingStat == "" {
		t.Error("Expected need-driven intent (with DrivingStat), got idle intent")
	}

	// Order should be paused
	if order.Status != entity.OrderPaused {
		t.Errorf("Order status: got %s, want Paused", order.Status)
	}
}

// =============================================================================
// Full Order Lifecycle Test
// =============================================================================

// TestOrderLifecycle_FullFlow tests the complete order lifecycle:
// 1. Create order (Open status)
// 2. Character takes order (Assigned status, bidirectional relationship)
// 3. Character works on order (intent continues across ticks)
// 4. Order is paused when character has needs
// 5. Order is resumed when needs are satisfied
// 6. Order is completed when inventory is full
//
// This test would have caught the ID=0 bug because it verifies the character
// can be identified as having an assigned order across multiple ticks.
func TestOrderLifecycle_FullFlow(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Hunger = 0
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100
	char.IdleCooldown = 0
	gameMap.AddCharacter(char)

	// Add a berry at the character's position (for immediate pickup)
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Step 1: Create order with valid ID (simulating UI behavior)
	// IMPORTANT: Order ID must be >= 1, not 0
	nextOrderID := 1 // This is what the UI should initialize to
	order := entity.NewOrder(nextOrderID, "harvest", "berry")
	orders := []*entity.Order{order}

	// Verify initial state
	if order.Status != entity.OrderOpen {
		t.Fatalf("Step 1: Order should start as Open, got %s", order.Status)
	}

	// Step 2: Character takes order
	items := gameMap.Items()
	intent := CalculateIntent(char, items, gameMap, nil, orders)

	if intent == nil {
		t.Fatal("Step 2: Expected intent after taking order, got nil")
	}
	if order.Status != entity.OrderAssigned {
		t.Errorf("Step 2: Order status should be Assigned, got %s", order.Status)
	}
	if char.AssignedOrderID != order.ID {
		t.Errorf("Step 2: char.AssignedOrderID should be %d, got %d", order.ID, char.AssignedOrderID)
	}
	if order.AssignedTo != char.ID {
		t.Errorf("Step 2: order.AssignedTo should be %d, got %d", char.ID, order.AssignedTo)
	}

	// Verify the "has order" check works (this is what failed with ID=0)
	if char.AssignedOrderID == 0 {
		t.Error("Step 2: char.AssignedOrderID == 0 means character appears to have no order!")
	}

	// Step 3: Simulate working on order - character should continue intent
	// Set up the intent as if character is at the item
	char.Intent = intent
	items = gameMap.Items() // Refresh items list

	// On subsequent tick, intent should continue
	intent2 := CalculateIntent(char, items, gameMap, nil, orders)
	if intent2 == nil {
		t.Fatal("Step 3: Expected intent to continue, got nil")
	}
	if intent2.Action != entity.ActionPickup {
		t.Errorf("Step 3: Intent should still be ActionPickup, got %v", intent2.Action)
	}

	// Step 4: Pause order due to moderate need
	char.Hunger = 80 // Moderate hunger
	char.Intent = intent2

	intent3 := CalculateIntent(char, items, gameMap, nil, orders)
	if intent3 == nil {
		t.Fatal("Step 4: Expected food-seeking intent, got nil")
	}
	if order.Status != entity.OrderPaused {
		t.Errorf("Step 4: Order should be Paused, got %s", order.Status)
	}
	// Character should still have the order assigned
	if char.AssignedOrderID != order.ID {
		t.Errorf("Step 4: char.AssignedOrderID should still be %d, got %d", order.ID, char.AssignedOrderID)
	}

	// Step 5: Resume order after needs satisfied
	char.Hunger = 0 // Needs satisfied
	char.Intent = nil
	char.IdleCooldown = 0

	intent4 := CalculateIntent(char, items, gameMap, nil, orders)
	if intent4 == nil {
		t.Fatal("Step 5: Expected order to resume, got nil")
	}
	if order.Status != entity.OrderAssigned {
		t.Errorf("Step 5: Order should be Assigned (resumed), got %s", order.Status)
	}
	if intent4.Action != entity.ActionPickup {
		t.Errorf("Step 5: Intent should be ActionPickup after resume, got %v", intent4.Action)
	}

	// Step 6: Complete order (simulate pickup completing)
	// The actual pickup and completion happens in applyIntent (UI layer)
	// Here we just verify the CompleteOrder function works correctly
	char.Carrying = berry // Simulate inventory is now full
	gameMap.RemoveItem(berry)

	CompleteOrder(char, order, nil)

	if char.AssignedOrderID != 0 {
		t.Errorf("Step 6: char.AssignedOrderID should be 0 after completion, got %d", char.AssignedOrderID)
	}
	if order.Status != entity.OrderOpen {
		t.Errorf("Step 6: Order status should be Open (ready for removal), got %s", order.Status)
	}
	if order.AssignedTo != 0 {
		t.Errorf("Step 6: order.AssignedTo should be 0 after completion, got %d", order.AssignedTo)
	}
}

// TestOrderLifecycle_ResumesAfterTargetTaken tests that when a character's
// target item is taken by another character, the order finds a new target.
func TestOrderLifecycle_ResumesAfterTargetTaken(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	char.Hunger = 0
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100
	char.IdleCooldown = 0
	gameMap.AddCharacter(char)

	// Add two berries
	berry1 := entity.NewBerry(6, 5, types.ColorRed, false, false)
	berry2 := entity.NewBerry(7, 5, types.ColorBlue, false, false)
	gameMap.AddItem(berry1)
	gameMap.AddItem(berry2)

	order := entity.NewOrder(1, "harvest", "berry")
	orders := []*entity.Order{order}

	// Character takes order
	items := gameMap.Items()
	intent := CalculateIntent(char, items, gameMap, nil, orders)

	if intent == nil || intent.TargetItem != berry1 {
		t.Fatal("Should target nearest berry (berry1)")
	}

	// Simulate berry1 being taken by another character
	gameMap.RemoveItem(berry1)
	char.Intent = intent // Character still has old intent

	// On next tick, continueIntent should fail (target gone)
	// Then selectIdleActivity -> selectOrderActivity should find berry2
	items = gameMap.Items()
	intent2 := CalculateIntent(char, items, gameMap, nil, orders)

	if intent2 == nil {
		t.Fatal("Should get new intent targeting berry2")
	}
	if intent2.TargetItem != berry2 {
		t.Errorf("Should target berry2 after berry1 taken, got %v", intent2.TargetItem)
	}

	// Order should still be assigned
	if order.Status != entity.OrderAssigned {
		t.Errorf("Order should still be Assigned, got %s", order.Status)
	}
}

// =============================================================================
// findCraftVesselIntent
// =============================================================================

func TestFindCraftVesselIntent_WithGourdInInventory(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.Carrying = entity.NewGourd(0, 0, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftVesselIntent(char, types.Position{X: 5, Y: 5}, items, order, nil)

	if intent == nil {
		t.Fatal("Expected craft intent")
	}
	if intent.Action != entity.ActionCraft {
		t.Errorf("Expected ActionCraft, got %d", intent.Action)
	}
	if intent.TargetItem != char.Carrying {
		t.Error("Expected TargetItem to be the carried gourd")
	}
}

func TestFindCraftVesselIntent_WithoutGourd_FindsGourdOnMap(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	// Not carrying anything
	gameMap.AddCharacter(char)

	gourd := entity.NewGourd(7, 5, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(gourd)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftVesselIntent(char, types.Position{X: 5, Y: 5}, items, order, nil)

	if intent == nil {
		t.Fatal("Expected pickup intent for gourd")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
	if intent.TargetItem != gourd {
		t.Error("Expected TargetItem to be the gourd on map")
	}
}

func TestFindCraftVesselIntent_NoGourdsAvailable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	gameMap.AddCharacter(char)

	// No gourds on map
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftVesselIntent(char, types.Position{X: 5, Y: 5}, items, order, nil)

	if intent != nil {
		t.Error("Expected nil intent when no gourds available")
	}
}

func TestFindCraftVesselIntent_CarryingNonGourd_FindsGourd(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.Carrying = entity.NewBerry(0, 0, types.ColorRed, false, false) // carrying berry, not gourd
	gameMap.AddCharacter(char)

	gourd := entity.NewGourd(7, 5, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(gourd)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftVesselIntent(char, types.Position{X: 5, Y: 5}, items, order, nil)

	if intent == nil {
		t.Fatal("Expected pickup intent for gourd")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
	if intent.TargetItem != gourd {
		t.Error("Expected TargetItem to be the gourd on map")
	}
}

// =============================================================================
// Drop behavior during order execution
// =============================================================================

func TestDrop_RemovesFromInventoryAndPlacesOnMap(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = berry
	gameMap.AddCharacter(char)

	Drop(char, gameMap, nil)

	if char.Carrying != nil {
		t.Error("Expected inventory to be empty after drop")
	}

	// Item should be on map at character's position
	droppedItem := gameMap.ItemAt(5, 5)
	if droppedItem != berry {
		t.Error("Expected dropped item at character position")
	}
}

func TestSelectOrderActivity_FullInventory_DropsOnPickup(t *testing.T) {
	// This tests the full flow: character with full inventory takes order,
	// then during execution drops current item to pick up target.
	// The drop happens in update.go during ActionPickup execution,
	// so here we just verify the intent is created correctly.
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	existingBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = existingBerry // inventory full
	gameMap.AddCharacter(char)

	targetBerry := entity.NewBerry(5, 5, types.ColorBlue, false, false) // at same position
	gameMap.AddItem(targetBerry)

	order := entity.NewOrder(1, "harvest", "berry")
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, types.Position{X: 5, Y: 5}, items, gameMap, orders, nil)

	// Should get pickup intent even with full inventory
	if intent == nil {
		t.Fatal("Expected pickup intent")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
	if intent.TargetItem != targetBerry {
		t.Error("Expected target to be the berry on map")
	}
	// Order should be assigned
	if order.Status != entity.OrderAssigned {
		t.Errorf("Expected OrderAssigned, got %s", order.Status)
	}
}

// =============================================================================
// Harvest Filter - IsGrowing (Feature 3b)
// =============================================================================

func TestFindNearestItemByType_SkipsNonGrowingItems(t *testing.T) {
	t.Parallel()

	// Growing berry (should be targeted)
	growingBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	// growingBerry.Plant.IsGrowing is true by default

	// Non-growing berry (dropped - should be skipped)
	droppedBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry, growingBerry}

	result := findNearestItemByType(0, 0, items, "berry")

	if result != growingBerry {
		t.Errorf("Expected growing berry, got %v", result)
	}
}

func TestFindNearestItemByType_ReturnsNilWhenOnlyNonGrowingItems(t *testing.T) {
	t.Parallel()

	// Only non-growing items
	droppedBerry := entity.NewBerry(3, 3, types.ColorRed, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry}

	result := findNearestItemByType(0, 0, items, "berry")

	if result != nil {
		t.Error("Should return nil when only non-growing items exist")
	}
}

func TestFindNearestItemByType_SkipsItemsWithNilPlant(t *testing.T) {
	t.Parallel()

	// Vessel (no Plant property)
	gourd := entity.NewGourd(3, 3, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)
	vessel.ItemType = "berry" // Artificially set type to test Plant filter

	// Growing berry
	growingBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{vessel, growingBerry}

	result := findNearestItemByType(0, 0, items, "berry")

	if result != growingBerry {
		t.Errorf("Expected growing berry (not vessel), got %v", result)
	}
}
