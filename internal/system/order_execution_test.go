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
	// Fill inventory to capacity
	char.AddToInventory(entity.NewBerry(0, 0, types.ColorRed, false, false))
	char.AddToInventory(entity.NewBerry(0, 0, types.ColorRed, false, false))
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
	char.AddToInventory(vessel)

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
	if char.GetCarriedVessel() != nil {
		t.Error("Vessel should have been dropped due to variety mismatch")
	}

	// Vessel should be on the map at character's position
	droppedVessel := gameMap.ItemAt(types.Position{X: 5, Y: 5})
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
	char.AddToInventory(vessel)

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
	if char.GetCarriedVessel() != vessel {
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
	char.AddToInventory(berry) // Simulate item picked up
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

// =============================================================================
// findCraftIntent (generic, replaces findCraftVesselIntent)
// =============================================================================

func TestFindCraftIntent_ReturnsActionCraftWithRecipeID_WhenInputsGathered(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	char.KnownRecipes = []string{"shell-hoe"}
	// Character has both recipe inputs
	char.AddToInventory(entity.NewStick(0, 0))
	char.AddToInventory(entity.NewShell(0, 0, types.ColorSilver))
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "craftHoe", "")

	items := gameMap.Items()
	intent := findCraftIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected craft intent")
	}
	if intent.Action != entity.ActionCraft {
		t.Errorf("Expected ActionCraft, got %d", intent.Action)
	}
	if intent.RecipeID != "shell-hoe" {
		t.Errorf("Expected RecipeID 'shell-hoe', got %q", intent.RecipeID)
	}
}

func TestFindCraftIntent_ReturnsPickupIntent_WhenInputsMissing(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	char.KnownRecipes = []string{"shell-hoe"}
	// Character has no inputs, but they exist on map
	gameMap.AddCharacter(char)

	stick := entity.NewStick(7, 5)
	shell := entity.NewShell(8, 5, types.ColorSilver)
	gameMap.AddItem(stick)
	gameMap.AddItem(shell)

	order := entity.NewOrder(1, "craftHoe", "")

	items := gameMap.Items()
	intent := findCraftIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected pickup intent for missing input")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
	// Should target one of the recipe inputs
	if intent.TargetItem != stick && intent.TargetItem != shell {
		t.Error("Expected TargetItem to be stick or shell")
	}
}

func TestFindCraftIntent_ReturnsNil_WhenNoRecipeFeasible(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	char.KnownRecipes = []string{"shell-hoe"}
	gameMap.AddCharacter(char)

	// No sticks or shells on map — only berries
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "craftHoe", "")

	items := gameMap.Items()
	intent := findCraftIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no recipe inputs exist in world")
	}
}

func TestFindCraftIntent_VesselRegression_WithGourdInInventory(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.KnownRecipes = []string{"hollow-gourd"}
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	char.AddToInventory(gourd)
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected craft intent")
	}
	if intent.Action != entity.ActionCraft {
		t.Errorf("Expected ActionCraft, got %d", intent.Action)
	}
	if intent.RecipeID != "hollow-gourd" {
		t.Errorf("Expected RecipeID 'hollow-gourd', got %q", intent.RecipeID)
	}
	// Gourd should still be in inventory (not extracted yet)
	if char.FindInInventory(func(i *entity.Item) bool { return i.ItemType == "gourd" }) == nil {
		t.Error("Gourd should still be in inventory (consumed only when craft completes)")
	}
}

func TestFindCraftIntent_VesselRegression_FindsGourdOnMap(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.KnownRecipes = []string{"hollow-gourd"}
	gameMap.AddCharacter(char)

	gourd := entity.NewGourd(7, 5, types.ColorGreen, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(gourd)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

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

func TestFindCraftIntent_VesselRegression_NoGourdsAvailable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftVessel"}
	char.KnownRecipes = []string{"hollow-gourd"}
	gameMap.AddCharacter(char)

	// No gourds on map
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "craftVessel", "")

	items := gameMap.Items()
	intent := findCraftIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no gourds available")
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
	char.Inventory = []*entity.Item{berry}
	gameMap.AddCharacter(char)

	Drop(char, gameMap, nil)

	if len(char.Inventory) != 0 {
		t.Error("Expected inventory to be empty after drop")
	}

	// Item should be on map at character's position
	droppedItem := gameMap.ItemAt(types.Position{X: 5, Y: 5})
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
	existingBerry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	existingBerry2 := entity.NewBerry(0, 0, types.ColorGreen, false, false)
	char.Inventory = []*entity.Item{existingBerry1, existingBerry2} // inventory full (2 slots)
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

	result := findNearestItemByType(0, 0, items, "berry", true)

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

	result := findNearestItemByType(0, 0, items, "berry", true)

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

	result := findNearestItemByType(0, 0, items, "berry", true)

	if result != growingBerry {
		t.Errorf("Expected growing berry (not vessel), got %v", result)
	}
}

// =============================================================================
// findTillSoilIntent
// =============================================================================

func TestFindTillSoilIntent_ReturnsHoeProcurementIntent_WhenNoHoeCarried(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	gameMap.AddCharacter(char)

	// Hoe on the map, not in inventory
	hoe := entity.NewHoe(7, 5, types.ColorSilver)
	gameMap.AddItem(hoe)

	// Mark a tile for tilling
	gameMap.MarkForTilling(types.Position{X: 3, Y: 3})

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	intent := findTillSoilIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected hoe procurement intent")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
	if intent.TargetItem != hoe {
		t.Error("Expected intent to target the hoe")
	}
}

func TestFindTillSoilIntent_ReturnsMovementIntent_TowardNearestMarkedTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	// Character carrying a hoe
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// Mark a tile for tilling (not at character's position)
	gameMap.MarkForTilling(types.Position{X: 8, Y: 5})

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	intent := findTillSoilIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected movement intent toward marked tile")
	}
	if intent.Action != entity.ActionTillSoil {
		t.Errorf("Expected ActionTillSoil, got %d", intent.Action)
	}
	// Dest should be the marked tile
	if intent.Dest.X != 8 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest (8,5), got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
	// Target (next step) should be moving toward the marked tile
	if intent.Target.X <= 5 {
		t.Error("Expected Target to move toward marked tile (X > 5)")
	}
}

func TestFindTillSoilIntent_ReturnsActionTillSoil_WhenAtMarkedPosition(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// Mark the character's current position for tilling
	gameMap.MarkForTilling(types.Position{X: 5, Y: 5})

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	intent := findTillSoilIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected ActionTillSoil intent")
	}
	if intent.Action != entity.ActionTillSoil {
		t.Errorf("Expected ActionTillSoil, got %d", intent.Action)
	}
	// Both Target and Dest should be the current position
	if intent.Target.X != 5 || intent.Target.Y != 5 {
		t.Errorf("Expected Target (5,5), got (%d,%d)", intent.Target.X, intent.Target.Y)
	}
	if intent.Dest.X != 5 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest (5,5), got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindTillSoilIntent_ReturnsNil_WhenPoolEmpty(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// No tiles marked for tilling
	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	intent := findTillSoilIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no tiles marked for tilling")
	}
}

func TestFindTillSoilIntent_SkipsAlreadyTilledPositions(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// Mark two tiles: one already tilled, one not
	alreadyTilled := types.Position{X: 6, Y: 5}
	notYetTilled := types.Position{X: 8, Y: 5}
	gameMap.MarkForTilling(alreadyTilled)
	gameMap.MarkForTilling(notYetTilled)
	gameMap.SetTilled(alreadyTilled) // This one is already done

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	intent := findTillSoilIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent for remaining untilled position")
	}
	if intent.Dest.X != 8 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest (8,5) for untilled tile, got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
}

// =============================================================================
// selectOrderActivity — tillSoil order completion vs abandonment
// =============================================================================

func TestSelectOrderActivity_CompletesTillSoil_WhenPoolEmpty(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	// No tiles marked for tilling — pool is empty
	order := entity.NewOrder(1, "tillSoil", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	orders := []*entity.Order{order}

	log := NewActionLog(100)
	items := gameMap.Items()
	intent := selectOrderActivity(char, char.Pos(), items, gameMap, orders, log)

	// Should return nil (no work to do)
	if intent != nil {
		t.Error("Expected nil intent when till pool is empty")
	}

	// Assignment should be cleared
	if char.AssignedOrderID != 0 {
		t.Error("Expected char.AssignedOrderID to be cleared after completion")
	}

	// Should log "Completed", not "Abandoning"
	events := log.Events(char.ID, 10)
	if len(events) == 0 {
		t.Fatal("Expected a log entry for order completion")
	}
	// CompleteOrder logs "Completed order: ..."
	// abandonOrder logs "Abandoning order: ..."
	found := false
	for _, e := range events {
		if len(e.Message) >= 9 && e.Message[:9] == "Completed" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected 'Completed' log message, got events: %v", events)
	}
}

func TestSelectOrderActivity_AbandonsTillSoil_WhenNoHoeAvailable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	// No hoe in inventory or on map
	gameMap.AddCharacter(char)

	// Tiles ARE marked for tilling — work exists but can't be done
	gameMap.MarkForTilling(types.Position{X: 3, Y: 3})

	order := entity.NewOrder(1, "tillSoil", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID
	orders := []*entity.Order{order}

	items := gameMap.Items()
	intent := selectOrderActivity(char, char.Pos(), items, gameMap, orders, nil)

	if intent != nil {
		t.Error("Expected nil intent when no hoe available")
	}

	// Order should be abandoned (status back to Open)
	if order.Status != entity.OrderOpen {
		t.Errorf("Order should be abandoned (Open), got %s", order.Status)
	}
	if char.AssignedOrderID != 0 {
		t.Error("Expected char.AssignedOrderID to be cleared after abandonment")
	}
}

// =============================================================================
// IsOrderFeasible
// =============================================================================

func TestIsOrderFeasible_HarvestFeasible_WhenGrowingItemsExist(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Harvest should be feasible when growing berries exist")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false")
	}
}

func TestIsOrderFeasible_HarvestUnfeasible_WhenNoGrowingItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	// No berries on map

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Harvest should be unfeasible when no growing berries exist")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false (components missing, not know-how)")
	}
}

func TestIsOrderFeasible_CraftFeasible_WhenRecipeInputsExist(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	gameMap.AddCharacter(char)

	stick := entity.NewStick(7, 5)
	shell := entity.NewShell(8, 5, types.ColorSilver)
	gameMap.AddItem(stick)
	gameMap.AddItem(shell)

	order := entity.NewOrder(1, "craftHoe", "")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Craft hoe should be feasible when stick and shell exist on map")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false")
	}
}

func TestIsOrderFeasible_CraftFeasible_WhenInputInCharacterInventory(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	char.AddToInventory(entity.NewStick(0, 0))
	gameMap.AddCharacter(char)

	// Shell on map, stick in inventory
	shell := entity.NewShell(8, 5, types.ColorSilver)
	gameMap.AddItem(shell)

	order := entity.NewOrder(1, "craftHoe", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Craft should be feasible when inputs exist across inventory and map")
	}
}

func TestIsOrderFeasible_CraftUnfeasible_WhenInputMissing(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftHoe"}
	gameMap.AddCharacter(char)

	// Only a stick, no shell
	stick := entity.NewStick(7, 5)
	gameMap.AddItem(stick)

	order := entity.NewOrder(1, "craftHoe", "")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Craft hoe should be unfeasible when shell is missing")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false (components missing, not know-how)")
	}
}

func TestIsOrderFeasible_TillSoilFeasible_WhenHoeOnGround(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	gameMap.AddCharacter(char)

	hoe := entity.NewHoe(7, 5, types.ColorSilver)
	gameMap.AddItem(hoe)
	gameMap.MarkForTilling(types.Position{X: 3, Y: 3})

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Till soil should be feasible when hoe exists on ground and tiles are marked")
	}
}

func TestIsOrderFeasible_TillSoilFeasible_WhenHoeInInventory(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char.AddToInventory(hoe)
	gameMap.AddCharacter(char)

	gameMap.MarkForTilling(types.Position{X: 3, Y: 3})

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Till soil should be feasible when hoe is in character inventory")
	}
}

func TestIsOrderFeasible_TillSoilUnfeasible_WhenNoHoe(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	gameMap.AddCharacter(char)

	gameMap.MarkForTilling(types.Position{X: 3, Y: 3})

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Till soil should be unfeasible when no hoe exists")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false")
	}
}

func TestIsOrderFeasible_TillSoilUnfeasible_WhenNoMarkedPositions(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"tillSoil"}
	gameMap.AddCharacter(char)

	hoe := entity.NewHoe(7, 5, types.ColorSilver)
	gameMap.AddItem(hoe)
	// No tiles marked for tilling

	order := entity.NewOrder(1, "tillSoil", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Till soil should be unfeasible when no tiles are marked for tilling")
	}
}

func TestIsOrderFeasible_NoKnowHow_WhenNoCharacterKnowsActivity(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	// Character does NOT know harvest
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Should be unfeasible when no character knows the activity")
	}
	if !noKnowHow {
		t.Error("noKnowHow should be true")
	}
}

func TestIsOrderFeasible_KnowHow_WhenAtLeastOneCharacterKnows(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	// First character doesn't know harvest
	char1 := entity.NewCharacter(1, 5, 5, "Test1", "berry", types.ColorRed)
	gameMap.AddCharacter(char1)
	// Second character knows harvest
	char2 := entity.NewCharacter(2, 3, 3, "Test2", "berry", types.ColorBlue)
	char2.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char2)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "harvest", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Should be feasible when at least one character knows the activity")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false when a character has the know-how")
	}
}

func TestIsOrderFeasible_PlantFeasible_WhenPlantableSeedExistsByKind(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	seed := entity.NewSeed(7, 5, "gourd", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	// Target is the Kind "gourd seed", not the ItemType "seed"
	order := entity.NewOrder(1, "plant", "gourd seed")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Plant should be feasible when plantable gourd seed exists (matched by Kind)")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false")
	}
}

func TestIsOrderFeasible_PlantFeasible_WhenPlantableBerryExists(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	berry := &entity.Item{
		ItemType:  "berry",
		Plantable: true,
	}
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Plant should be feasible when plantable berry exists (matched by ItemType)")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false")
	}
}

func TestIsOrderFeasible_PlantFeasible_WhenPlantableBerryInVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	// Berry stored in vessel (Plantable flag lost, but config says berry is plantable)
	vessel := &entity.Item{
		ItemType: "vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{
				{Variety: &entity.ItemVariety{ItemType: "berry", Color: types.ColorRed}, Count: 3},
			},
		},
	}
	char.Inventory = []*entity.Item{vessel}
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Plant should be feasible when berries exist in a vessel")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false")
	}
}

func TestIsOrderFeasible_PlantUnfeasible_WhenNoPlantableItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Berry on map is growing (not plantable)
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Plant should be unfeasible when no plantable berries exist")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false (items missing, not know-how)")
	}
}

func TestIsOrderFeasible_PlantUnfeasible_WhenWrongSeedKind(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Gourd seed exists but order targets flower seeds
	seed := entity.NewSeed(7, 5, "gourd", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	order := entity.NewOrder(1, "plant", "flower seed")
	items := gameMap.Items()

	feasible, noKnowHow := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Plant should be unfeasible when no flower seeds exist (only gourd seeds)")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false (items missing, not know-how)")
	}
}

// =============================================================================
// findAvailableOrder — skips unfeasible orders
// =============================================================================

func TestFindAvailableOrder_SkipsUnfeasibleOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	// Berry exists for the second order but not for first
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// First order: harvest mushrooms (no mushrooms on map — unfeasible)
	order1 := entity.NewOrder(1, "harvest", "mushroom")
	// Second order: harvest berries (berries exist — feasible)
	order2 := entity.NewOrder(2, "harvest", "berry")
	orders := []*entity.Order{order1, order2}

	items := gameMap.Items()
	result := findAvailableOrder(char, orders, items, gameMap)

	if result != order2 {
		t.Errorf("Should skip unfeasible order1 and return order2, got %v", result)
	}
}

// =============================================================================
// findPlantIntent
// =============================================================================

func TestFindPlantIntent_ReturnsProcurementIntent_WhenNoPlantableCarried(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Tilled empty tile
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

	// Plantable seed on the ground (not carried)
	seed := entity.NewSeed(7, 5, "gourd", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	order := entity.NewOrder(1, "plant", "gourd seed")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected procurement intent to pick up seed")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
	if intent.TargetItem != seed {
		t.Error("Expected intent to target the seed on the ground")
	}
}

func TestFindPlantIntent_ReturnsMovementIntent_WhenCarryingPlantable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	// Carrying a plantable berry
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)
	gameMap.AddCharacter(char)

	// Tilled empty tile away from character
	gameMap.SetTilled(types.Position{X: 8, Y: 5})

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected movement intent toward tilled tile")
	}
	if intent.Action != entity.ActionPlant {
		t.Errorf("Expected ActionPlant, got %d", intent.Action)
	}
	// Dest should be the tilled tile
	if intent.Dest.X != 8 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest (8,5), got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
	// Target (next step) should be moving toward the tilled tile
	if intent.Target.X <= 5 {
		t.Error("Expected Target to move toward tilled tile (X > 5)")
	}
}

func TestFindPlantIntent_ReturnsActionPlant_WhenAtEmptyTilledTileWithPlantable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)
	gameMap.AddCharacter(char)

	// Character is standing on tilled empty tile
	gameMap.SetTilled(types.Position{X: 5, Y: 5})

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected ActionPlant intent")
	}
	if intent.Action != entity.ActionPlant {
		t.Errorf("Expected ActionPlant, got %d", intent.Action)
	}
	if intent.Target.X != 5 || intent.Target.Y != 5 {
		t.Errorf("Expected Target (5,5), got (%d,%d)", intent.Target.X, intent.Target.Y)
	}
	if intent.Dest.X != 5 || intent.Dest.Y != 5 {
		t.Errorf("Expected Dest (5,5), got (%d,%d)", intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindPlantIntent_ReturnsNil_WhenNoEmptyTilledTiles(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)
	gameMap.AddCharacter(char)

	// No tilled tiles at all
	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no tilled tiles exist")
	}
}

func TestFindPlantIntent_ReturnsNil_WhenNoPlantableItemsAvailable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Tilled tile exists
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

	// No plantable items on map or in inventory
	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no plantable items available")
	}
}

func TestFindPlantIntent_WithLockedVariety_OnlySeeksMatchingVariety(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Tilled tile
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

	// Blue berry on map (plantable but wrong variety)
	blueBerry := entity.NewBerry(7, 5, types.ColorBlue, false, false)
	blueBerry.Plantable = true
	blueBerry.Plant.IsGrowing = false
	gameMap.AddItem(blueBerry)

	// Order has locked variety to "red berry"
	order := entity.NewOrder(1, "plant", "berry")
	order.LockedVariety = entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone)
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	// Should return nil — blue berry doesn't match locked red variety
	if intent != nil {
		t.Error("Expected nil intent — only blue berry available but locked to red variety")
	}
}

func TestFindPlantIntent_SkipsTilledTilesWithItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)
	gameMap.AddCharacter(char)

	// Two tilled tiles: one has a sprout, one is empty
	occupiedPos := types.Position{X: 6, Y: 5}
	emptyPos := types.Position{X: 8, Y: 5}
	gameMap.SetTilled(occupiedPos)
	gameMap.SetTilled(emptyPos)

	// Place a sprout on the occupied tile
	sprout := entity.CreateSprout(occupiedPos.X, occupiedPos.Y, berry, berry.Edible)
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent targeting empty tilled tile")
	}
	// Should target the empty tile, not the occupied one
	if intent.Dest.X != emptyPos.X || intent.Dest.Y != emptyPos.Y {
		t.Errorf("Expected Dest %v, got (%d,%d)", emptyPos, intent.Dest.X, intent.Dest.Y)
	}
}

func TestFindPlantIntent_DropsUnneededItem_WhenInventoryFull(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Fill inventory with non-plantable items (hoe + stick)
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	stick := entity.NewStick(0, 0)
	char.AddToInventory(hoe)
	char.AddToInventory(stick)

	// Tilled empty tile
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

	// Plantable seed on the ground
	seed := entity.NewSeed(7, 5, "gourd", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	order := entity.NewOrder(1, "plant", "gourd seed")
	actionLog := NewActionLog(100)
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, actionLog, gameMap)

	// Should have dropped an item to make room
	if !char.HasInventorySpace() {
		t.Error("Expected character to drop an item to make inventory space")
	}
	// Should return pickup intent for the seed
	if intent == nil {
		t.Fatal("Expected procurement intent after dropping item")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %d", intent.Action)
	}
}

func TestFindAvailableOrder_ReturnsNilWhenAllUnfeasible(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	// No items on map at all

	order1 := entity.NewOrder(1, "harvest", "berry")
	order2 := entity.NewOrder(2, "harvest", "mushroom")
	orders := []*entity.Order{order1, order2}

	items := gameMap.Items()
	result := findAvailableOrder(char, orders, items, gameMap)

	if result != nil {
		t.Error("Should return nil when all orders are unfeasible")
	}
}
