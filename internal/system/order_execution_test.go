package system

import (
	"testing"

	"petri/internal/config"
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
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
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

func TestFindHarvestIntent_KeepsIncompatibleVessel_WhenSpaceAvailable(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Set up variety registry
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
	})
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("mushroom", "", types.ColorBrown, types.PatternSpotted, types.TextureSlimy),
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
		Edible:   &entity.EdibleProperties{},
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Character is carrying vessel with mushrooms (incompatible with berries)
	// but has a free inventory slot — should keep the vessel
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{
				Variety: registry.Get(entity.GenerateVarietyID("mushroom", "", types.ColorBrown, types.PatternSpotted, types.TextureSlimy)),
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

	intent := findHarvestIntent(char, types.Position{X: 5, Y: 5}, items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}

	// Vessel should still be in inventory — free slot available
	if char.GetCarriedVessel() != vessel {
		t.Error("Vessel should be kept when inventory has space")
	}

	// Intent should be to move toward the berry (no vessel on map to procure)
	if intent.TargetItem != berry {
		t.Error("Intent should target the berry for harvest")
	}
}

func TestFindHarvestIntent_UsesCompatibleVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Set up variety registry
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
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
				Variety: registry.Get(entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone)),
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
	if order.Status != entity.OrderCompleted {
		t.Errorf("Step 6: Order status should be Completed, got %s", order.Status)
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
	// Then selectOrderActivity should find berry2
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

	seed := entity.NewSeed(7, 5, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	// Need an available tilled tile
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

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

	// Need an available tilled tile
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

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
				{Variety: &entity.ItemVariety{ItemType: "berry", Color: types.ColorRed, Plantable: true}, Count: 3},
			},
		},
	}
	char.Inventory = []*entity.Item{vessel}
	gameMap.AddCharacter(char)

	// Need an available tilled tile
	gameMap.SetTilled(types.Position{X: 3, Y: 3})

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
	seed := entity.NewSeed(7, 5, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
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
	seed := entity.NewSeed(7, 5, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
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
	order.LockedVariety = entity.GenerateVarietyID("berry", "", types.ColorRed, types.PatternNone, types.TextureNone)
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	// Should return nil — blue berry doesn't match locked red variety
	if intent != nil {
		t.Error("Expected nil intent — only blue berry available but locked to red variety")
	}
}

func TestFindPlantIntent_SkipsTilledTilesWithGrowingPlants(t *testing.T) {
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
	berryVariety := &entity.ItemVariety{
		ID:       "berry-red",
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   &entity.EdibleProperties{},
		Sym:      config.CharBerry,
	}
	sprout := entity.CreateSprout(occupiedPos.X, occupiedPos.Y, berryVariety)
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

func TestFindPlantIntent_AllowsTilledTilesWithLooseItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true
	char.AddToInventory(berry)
	gameMap.AddCharacter(char)

	// One tilled tile with a loose seed on it (not growing)
	tilePos := types.Position{X: 6, Y: 5}
	gameMap.SetTilled(tilePos)
	looseSeed := entity.NewSeed(tilePos.X, tilePos.Y, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(looseSeed)

	order := entity.NewOrder(1, "plant", "berry")
	items := gameMap.Items()

	intent := findPlantIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent — tilled tile with loose item should be available for planting")
	}
	if intent.Dest.X != tilePos.X || intent.Dest.Y != tilePos.Y {
		t.Errorf("Expected Dest %v, got (%d,%d)", tilePos, intent.Dest.X, intent.Dest.Y)
	}
}

func TestIsOrderFeasible_PlantUnfeasible_WhenNoEmptyTilledTiles(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Seed exists (so PlantableItemExists is true)
	seed := entity.NewSeed(7, 5, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	// One tilled tile, but occupied by a growing plant
	tilePos := types.Position{X: 3, Y: 3}
	gameMap.SetTilled(tilePos)
	growingBerry := entity.NewBerry(tilePos.X, tilePos.Y, types.ColorRed, false, false)
	gameMap.AddItem(growingBerry)

	order := entity.NewOrder(1, "plant", "gourd seed")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Plant should be unfeasible when all tilled tiles have growing plants")
	}
}

func TestIsOrderFeasible_PlantFeasible_WhenTilledTileHasLooseItem(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"plant"}
	gameMap.AddCharacter(char)

	// Seed exists
	seed := entity.NewSeed(7, 5, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
	gameMap.AddItem(seed)

	// One tilled tile with a loose item (not a growing plant)
	tilePos := types.Position{X: 3, Y: 3}
	gameMap.SetTilled(tilePos)
	looseItem := entity.NewStick(tilePos.X, tilePos.Y)
	gameMap.AddItem(looseItem)

	order := entity.NewOrder(1, "plant", "gourd seed")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Plant should be feasible — tilled tile has only a loose item (not a growing plant)")
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
	seed := entity.NewSeed(7, 5, "gourd", "gourd-green", "", types.ColorGreen, types.PatternNone, types.TextureNone)
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

// =============================================================================
// CompleteOrder + OrderCompleted status
// =============================================================================

func TestCompleteOrder_SetsOrderCompletedStatus(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)
	CompleteOrder(char, order, log)

	if order.Status != entity.OrderCompleted {
		t.Errorf("Order status: got %s, want %s", order.Status, entity.OrderCompleted)
	}
	if char.AssignedOrderID != 0 {
		t.Error("Expected char.AssignedOrderID to be cleared")
	}
}

func TestSelectOrderActivity_SkipsCompletedOrders(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	// Completed order should be skipped, open order should be taken
	completedOrder := entity.NewOrder(1, "harvest", "berry")
	completedOrder.Status = entity.OrderCompleted
	openOrder := entity.NewOrder(2, "harvest", "berry")
	orders := []*entity.Order{completedOrder, openOrder}

	items := gameMap.Items()
	intent := selectOrderActivity(char, char.Pos(), items, gameMap, orders, nil)

	if intent == nil {
		t.Fatal("Expected intent from open order")
	}
	if char.AssignedOrderID != openOrder.ID {
		t.Errorf("Should have taken open order (id=2), got assigned to %d", char.AssignedOrderID)
	}
}

func TestFindAvailableOrder_SkipsCompletedOrders(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"harvest"}
	gameMap.AddCharacter(char)

	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	completedOrder := entity.NewOrder(1, "harvest", "berry")
	completedOrder.Status = entity.OrderCompleted
	orders := []*entity.Order{completedOrder}

	items := gameMap.Items()
	result := findAvailableOrder(char, orders, items, gameMap)

	if result != nil {
		t.Error("Should not return completed order")
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

// Water Garden feasibility tests

func TestIsOrderFeasible_WaterGardenFeasible(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Vessel on ground
	vessel := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'U', EType: entity.TypeItem},
		ItemType:   "vessel",
	}
	gameMap.AddItem(vessel)

	// Water source
	gameMap.AddWater(types.Position{X: 15, Y: 15}, game.WaterPond)

	// Tilled + planted tile that is NOT wet
	pos := types.Position{X: 3, Y: 3}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 3, Y: 3, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if !feasible {
		t.Error("Water garden should be feasible when vessel, water, and dry planted tilled tile exist")
	}
}

func TestIsOrderFeasible_WaterGardenUnfeasible_NoDryPlantedTiles(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Vessel on ground
	vessel := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'U', EType: entity.TypeItem},
		ItemType:   "vessel",
	}
	gameMap.AddItem(vessel)

	// Water source
	gameMap.AddWater(types.Position{X: 15, Y: 15}, game.WaterPond)

	// Tilled tile with plant, but it's already wet (adjacent to water)
	pos := types.Position{X: 14, Y: 15} // adjacent to pond
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 14, Y: 15, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Water garden should be unfeasible when all planted tiles are already wet")
	}
}

func TestIsOrderFeasible_WaterGardenUnfeasible_NoVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Water source but no vessel
	gameMap.AddWater(types.Position{X: 15, Y: 15}, game.WaterPond)

	// Dry tilled planted tile
	pos := types.Position{X: 3, Y: 3}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 3, Y: 3, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	feasible, _ := IsOrderFeasible(order, items, gameMap)

	if feasible {
		t.Error("Water garden should be unfeasible when no vessel exists")
	}
}

// =============================================================================
// findWaterGardenIntent
// =============================================================================

func TestFindWaterGardenIntent_ReturnsActionWaterGarden(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Character carries vessel with water
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

	// Dry tilled planted tile
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.Action != entity.ActionWaterGarden {
		t.Errorf("Expected ActionWaterGarden, got %v", intent.Action)
	}
	if intent.Dest != pos {
		t.Errorf("Expected dest %v, got %v", pos, intent.Dest)
	}
	if intent.TargetItem != vessel {
		t.Error("Expected TargetItem to be the water vessel")
	}
}

func TestFindWaterGardenIntent_EmptyVessel_TargetsWaterSource(t *testing.T) {
	// Anchor: character has empty vessel → Phase 2: targets water-adjacent position for filling
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Character carries empty vessel
	vessel := &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	char.AddToInventory(vessel)

	// Water source
	waterPos := types.Position{X: 5, Y: 10}
	gameMap.AddWater(waterPos, game.WaterPond)

	// Dry tilled planted tile
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent for water fill phase, got nil")
	}
	if intent.Action != entity.ActionWaterGarden {
		t.Errorf("Expected ActionWaterGarden, got %v", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("Expected TargetItem to be the empty vessel")
	}
}

func TestFindWaterGardenIntent_NoVessel_TargetsGroundVessel(t *testing.T) {
	// Anchor: character has no vessel, ground vessel exists → Phase 1: targets ground vessel
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Ground vessel nearby
	groundVessel := &entity.Item{
		ItemType:   "vessel",
		Name:       "Ground Vessel",
		BaseEntity: entity.BaseEntity{X: 8, Y: 5, Sym: 'U', EType: entity.TypeItem},
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	gameMap.AddItem(groundVessel)

	// Dry tilled planted tile
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent for vessel procurement phase, got nil")
	}
	if intent.Action != entity.ActionWaterGarden {
		t.Errorf("Expected ActionWaterGarden, got %v", intent.Action)
	}
	if intent.TargetItem != groundVessel {
		t.Error("Expected TargetItem to be the ground vessel")
	}
}

func TestFindWaterGardenIntent_NoVessel_TargetsGroundWaterVessel(t *testing.T) {
	// Anchor: character has no vessel, ground water vessel exists → Phase 1: targets ground water vessel
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Ground vessel with water nearby
	waterVessel := &entity.Item{
		ItemType:   "vessel",
		Name:       "Water Vessel",
		BaseEntity: entity.BaseEntity{X: 8, Y: 5, Sym: 'U', EType: entity.TypeItem},
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{
				Variety: &entity.ItemVariety{
					ItemType: "liquid",
					Kind:     "water",
				},
				Count: 4,
			}},
		},
	}
	gameMap.AddItem(waterVessel)

	// Dry tilled planted tile
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent to pick up ground water vessel, got nil")
	}
	if intent.Action != entity.ActionWaterGarden {
		t.Errorf("Expected ActionWaterGarden, got %v", intent.Action)
	}
	if intent.TargetItem != waterVessel {
		t.Error("Expected TargetItem to be the ground water vessel")
	}
}

func TestFindWaterGardenIntent_PrefersGroundWaterOverGroundEmpty(t *testing.T) {
	// Anchor: ground water vessel AND ground empty vessel both exist.
	// Water vessel is preferred because it skips the fill phase.
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Empty vessel closer
	emptyVessel := &entity.Item{
		ItemType:   "vessel",
		Name:       "Empty Vessel",
		BaseEntity: entity.BaseEntity{X: 6, Y: 5, Sym: 'U', EType: entity.TypeItem},
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	gameMap.AddItem(emptyVessel)

	// Water vessel farther
	waterVessel := &entity.Item{
		ItemType:   "vessel",
		Name:       "Water Vessel",
		BaseEntity: entity.BaseEntity{X: 9, Y: 5, Sym: 'U', EType: entity.TypeItem},
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{{
				Variety: &entity.ItemVariety{
					ItemType: "liquid",
					Kind:     "water",
				},
				Count: 4,
			}},
		},
	}
	gameMap.AddItem(waterVessel)

	// Water source for fill path
	gameMap.AddWater(types.Position{X: 15, Y: 5}, game.WaterPond)

	// Dry tilled planted tile
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent, got nil")
	}
	if intent.TargetItem != waterVessel {
		t.Error("Should prefer ground water vessel over ground empty vessel (skips fill phase)")
	}
}

func TestFindWaterGardenIntent_NoVesselAnywhere_ReturnsNil(t *testing.T) {
	// Anchor: no vessel in inventory or on ground → abandon (return nil)
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Dry tilled planted tile
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no vessel available anywhere")
	}
}

func TestFindWaterGardenIntent_ReturnsNilWhenNoDryTiles(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"waterGarden"}
	gameMap.AddCharacter(char)

	// Character carries vessel with water
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

	// Tilled planted tile that is already wet (manually watered)
	pos := types.Position{X: 7, Y: 5}
	gameMap.SetTilled(pos)
	gameMap.SetManuallyWatered(pos)
	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 7, Y: 5, Sym: 'v', EType: entity.TypeItem},
		ItemType:   "berry",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "waterGarden", "")
	items := gameMap.Items()

	intent := findWaterGardenIntent(char, char.Pos(), items, order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when all tilled planted tiles are already wet")
	}
}

// =============================================================================
// Gather Order Tests
// =============================================================================

// TestFindGatherIntent_ReturnsPickupForNearestItem is the anchor test:
// a character with a gather order gets a pickup intent targeting the nearest item of that type.
func TestFindGatherIntent_ReturnsPickupForNearestItem(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	nut := entity.NewNut(7, 5)
	gameMap.AddItem(nut)

	order := entity.NewOrder(1, "gather", "nut")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected pickup intent for nearest gatherable item, got nil")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Intent.Action: got %v, want ActionPickup", intent.Action)
	}
	if intent.TargetItem != nut {
		t.Errorf("Intent.TargetItem: got %v, want nut", intent.TargetItem)
	}
}

// TestFindGatherIntent_VesselProcurementForNut verifies that when gathering nuts
// (which have registered varieties), the character seeks a vessel.
func TestFindGatherIntent_VesselProcurementForNut(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has no vessel; a vessel is on the ground
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 5
	gameMap.AddItem(vessel)

	nut := entity.NewNut(7, 5)
	gameMap.AddItem(nut)

	order := entity.NewOrder(1, "gather", "nut")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	// Should return an intent to pick up the vessel first
	if intent == nil {
		t.Fatal("Expected intent to procure vessel, got nil")
	}
	if intent.TargetItem != vessel {
		t.Errorf("Intent.TargetItem: expected vessel (for procurement), got %v", intent.TargetItem)
	}
}

// TestFindGatherIntent_StickSkipsVessel verifies that when gathering sticks
// (no registered variety), the character picks up directly without vessel procurement.
func TestFindGatherIntent_StickSkipsVessel(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Vessel on the ground — should NOT be targeted for sticks
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 5
	gameMap.AddItem(vessel)

	stick := entity.NewStick(7, 5)
	gameMap.AddItem(stick)

	order := entity.NewOrder(1, "gather", "stick")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected pickup intent for stick, got nil")
	}
	if intent.TargetItem != stick {
		t.Errorf("Intent.TargetItem: expected stick, got %v", intent.TargetItem)
	}
}

// TestFindGatherIntent_StickNilWhenBundlesFull verifies that when gathering sticks
// with full bundles and no inventory space, findGatherIntent returns nil.
func TestFindGatherIntent_StickNilWhenBundlesFull(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	fullBundle1 := entity.NewStick(0, 0)
	fullBundle1.BundleCount = 6
	fullBundle2 := entity.NewStick(0, 0)
	fullBundle2.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle1, fullBundle2}
	gameMap.AddCharacter(char)

	stick := entity.NewStick(7, 5)
	gameMap.AddItem(stick)

	order := entity.NewOrder(1, "gather", "stick")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent != nil {
		t.Errorf("Expected nil intent when bundles full and no inventory space, got %v", intent)
	}
}

// TestFindGatherIntent_StickAllowedWhenFullWithNonSticks verifies that a character
// with full inventory of non-stick items can still take a gather sticks order.
// The order handler (applyPickup) will drop an item on arrival.
func TestFindGatherIntent_StickAllowedWhenFullWithNonSticks(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	// Full inventory with non-stick items (seeds)
	seed1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	seed1.Plantable = true
	seed2 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	seed2.Plantable = true
	char.Inventory = []*entity.Item{seed1, seed2}
	gameMap.AddCharacter(char)

	stick := entity.NewStick(7, 5)
	gameMap.AddItem(stick)

	order := entity.NewOrder(1, "gather", "stick")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent (character can drop non-stick item), got nil")
	}
	if intent.TargetItem != stick {
		t.Errorf("Expected target to be the stick, got %v", intent.TargetItem)
	}
}

// TestFindGatherIntent_NilWhenFullBundle verifies one-bundle-per-order: if the character
// already has a full bundle of the target type, findGatherIntent returns nil even with
// more sticks on the map.
func TestFindGatherIntent_NilWhenFullBundle(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	fullBundle := entity.NewStick(0, 0)
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil} // full bundle + empty slot
	gameMap.AddCharacter(char)

	// More sticks exist on the map
	gameMap.AddItem(entity.NewStick(7, 5))
	gameMap.AddItem(entity.NewStick(8, 5))

	order := entity.NewOrder(1, "gather", "stick")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	intent := FindGatherIntentForTest(char, char.Pos(), gameMap.Items(), order, nil, gameMap)
	if intent != nil {
		t.Errorf("Expected nil (one bundle per order), got intent targeting %v", intent.TargetItem)
	}
}

// TestIsMultiStepOrderComplete_GatherWithFullBundle verifies that a gather order
// is considered complete when the character has a full bundle of the target type.
func TestIsMultiStepOrderComplete_GatherWithFullBundle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	fullBundle := entity.NewStick(0, 0)
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "gather", "stick")

	if !IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected gather order to be complete with full bundle")
	}

	// Partial bundle should NOT be complete
	char.Inventory[0].BundleCount = 3
	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected gather order to NOT be complete with partial bundle")
	}
}

// TestFindNearestItemByType_SkipsFullBundles verifies that full bundles on the ground
// are not valid gather targets — they're finished products, not raw material.
func TestFindNearestItemByType_SkipsFullBundles(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Full bundle on the ground (closer)
	fullBundle := entity.NewStick(6, 5)
	fullBundle.BundleCount = 6
	gameMap.AddItem(fullBundle)

	// Loose stick on the ground (further)
	looseStick := entity.NewStick(9, 5)
	gameMap.AddItem(looseStick)

	items := gameMap.Items()
	target := FindNearestItemByTypeForTest(5, 5, items, "stick", false)

	if target == nil {
		t.Fatal("Expected to find the loose stick, got nil")
	}
	if target == fullBundle {
		t.Error("Should skip full bundle, got the full bundle instead of the loose stick")
	}
	if target != looseStick {
		t.Errorf("Expected loose stick, got %v", target)
	}
}

// TestFindNextGatherTarget_NilWhenInventoryFull verifies completion signal when full.
func TestFindNextGatherTarget_NilWhenInventoryFull(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.Inventory = []*entity.Item{entity.NewNut(0, 0), entity.NewNut(0, 0)}
	gameMap.AddCharacter(char)

	nut := entity.NewNut(7, 5)
	gameMap.AddItem(nut)

	intent := FindNextGatherTarget(char, 5, 5, gameMap.Items(), "nut", gameMap)
	if intent != nil {
		t.Errorf("Expected nil when inventory full, got %v", intent)
	}
}

// TestFindNextGatherTarget_NilWhenNoItems verifies completion signal when no more items.
func TestFindNextGatherTarget_NilWhenNoItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	intent := FindNextGatherTarget(char, 5, 5, []*entity.Item{}, "nut", gameMap)
	if intent != nil {
		t.Errorf("Expected nil when no items exist, got %v", intent)
	}
}

// TestIsOrderFeasible_GatherOrder verifies feasibility when target type exists on ground.
func TestIsOrderFeasible_GatherOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	nut := entity.NewNut(5, 5)
	gameMap.AddItem(nut)

	order := entity.NewOrder(1, "gather", "nut")

	feasible, noKnowHow := IsOrderFeasible(order, gameMap.Items(), gameMap)
	if !feasible {
		t.Error("Expected gather order to be feasible when nut exists on ground")
	}
	if noKnowHow {
		t.Error("Expected noKnowHow=false for gather (AvailabilityDefault)")
	}
}

// TestIsOrderFeasible_GatherOrder_NoItems verifies infeasibility when nothing on ground.
func TestIsOrderFeasible_GatherOrder_NoItems(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	order := entity.NewOrder(1, "gather", "nut")

	feasible, _ := IsOrderFeasible(order, []*entity.Item{}, gameMap)
	if feasible {
		t.Error("Expected gather order to be infeasible when no items on ground")
	}
}

// Anchor test: Exercises the full gather-order flow by chaining system functions
// in the sequence the UI handler calls them. Catches regressions in any function
// in the chain: EnsureHasVesselFor (drop + procurement), Pickup (vessel + inventory
// paths), FindNextVesselTarget/FindNextGatherTarget (continuation), and HasItemOnMap
// (pointer identity when items share positions).
//
// Scenario: Character has full inventory (2 sticks), gets a gather-nuts order.
// Expected flow: drop stick → pick up vessel → pick up nut (to vessel) → continue
// to next nut → no more nuts → order complete.
func TestGatherOrder_VesselPath_EndToEnd(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(20, 20)
	gameMap.SetVarieties(registry)

	// Character starts with full inventory (2 sticks)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.Inventory = []*entity.Item{entity.NewStick(0, 0), entity.NewStick(0, 0)}
	gameMap.AddCharacter(char)

	// Vessel on the ground
	vessel := createTestVessel()
	vessel.X = 5
	vessel.Y = 5
	gameMap.AddItem(vessel)

	// Two nuts on the ground (one at character's position, one elsewhere)
	nut1 := entity.NewNut(5, 5)
	nut2 := entity.NewNut(7, 5)
	gameMap.AddItem(nut1)
	gameMap.AddItem(nut2)

	order := entity.NewOrder(1, "gather", "nut")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// --- Phase 1: findGatherIntent with full inventory ---
	// EnsureHasVesselFor should drop a stick to make room, then return vessel pickup intent
	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)
	if intent == nil {
		t.Fatal("Phase 1: Expected vessel procurement intent, got nil")
	}
	if intent.TargetItem != vessel {
		t.Fatal("Phase 1: Expected intent to target vessel for procurement")
	}
	if len(char.Inventory) != 1 {
		t.Fatalf("Phase 1: Expected 1 item after drop, got %d", len(char.Inventory))
	}

	// --- Phase 2: Pick up the vessel ---
	result := Pickup(char, vessel, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Phase 2: Expected PickupToInventory for vessel, got %d", result)
	}
	if char.GetCarriedVessel() == nil {
		t.Fatal("Phase 2: Character should now have vessel in inventory")
	}

	// Vessel pickup is a prerequisite — not gather work. UI handler skips
	// continuation for items with Container != nil. Verify vessel is a container
	// (this is the guard condition from bug #5).
	if vessel.Container == nil {
		t.Fatal("Phase 2: Vessel should have Container (prerequisite guard check)")
	}

	// --- Phase 3: findGatherIntent again — should now target nut ---
	items = gameMap.Items()
	intent = FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)
	if intent == nil {
		t.Fatal("Phase 3: Expected nut pickup intent, got nil")
	}
	if intent.TargetItem != nut1 && intent.TargetItem != nut2 {
		t.Fatal("Phase 3: Expected intent to target one of the nuts")
	}
	targetNut := intent.TargetItem

	// --- Phase 4: Pick up the first nut ---
	result = Pickup(char, targetNut, gameMap, nil, registry)
	if result != PickupToVessel {
		t.Fatalf("Phase 4: Expected PickupToVessel for nut, got %d", result)
	}

	// Verify nut went into vessel
	if len(vessel.Container.Contents) == 0 {
		t.Fatal("Phase 4: Vessel should contain the gathered nut")
	}

	// --- Phase 5: Continuation — FindNextVesselTarget finds next nut ---
	cpos := char.Pos()
	items = gameMap.Items()
	nextIntent := FindNextVesselTarget(char, cpos.X, cpos.Y, items, registry, gameMap, false)
	if nextIntent == nil {
		t.Fatal("Phase 5: Expected continuation intent for second nut, got nil")
	}

	// Verify it targets the remaining nut (not the one we just picked up)
	remainingNut := nut2
	if targetNut == nut2 {
		remainingNut = nut1
	}
	// The remaining nut should still be on the map
	if !gameMap.HasItemOnMap(remainingNut) {
		t.Fatal("Phase 5: Remaining nut should still be on the map")
	}

	// --- Phase 6: Pick up the second nut ---
	result = Pickup(char, remainingNut, gameMap, nil, registry)
	if result != PickupToVessel {
		t.Fatalf("Phase 6: Expected PickupToVessel for second nut, got %d", result)
	}

	// --- Phase 7: Continuation — no more nuts, order should complete ---
	items = gameMap.Items()
	nextIntent = FindNextVesselTarget(char, cpos.X, cpos.Y, items, registry, gameMap, false)
	if nextIntent != nil {
		t.Error("Phase 7: Expected nil (no more nuts) — signals order completion")
	}

	// Verify final state: vessel has 2 nuts
	if len(vessel.Container.Contents) != 1 {
		t.Fatalf("Final: Expected 1 stack in vessel, got %d", len(vessel.Container.Contents))
	}
	if vessel.Container.Contents[0].Count != 2 {
		t.Errorf("Final: Expected 2 nuts in vessel stack, got %d", vessel.Container.Contents[0].Count)
	}
}

// Anchor test: Exercises gather-order flow for inventory-path items (sticks).
// Sticks have no registered variety, so they go directly to inventory instead of vessels.
//
// Scenario: Character has empty inventory, gets a gather-sticks order. Two sticks
// on the ground. Expected flow: pick up stick → continue → pick up second stick →
// inventory full → order complete.
func TestGatherOrder_InventoryPath_EndToEnd(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(20, 20)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	stick1 := entity.NewStick(5, 5)
	stick2 := entity.NewStick(7, 5)
	gameMap.AddItem(stick1)
	gameMap.AddItem(stick2)

	order := entity.NewOrder(1, "gather", "stick")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// --- Phase 1: findGatherIntent targets nearest stick ---
	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)
	if intent == nil {
		t.Fatal("Phase 1: Expected stick pickup intent")
	}
	if intent.TargetItem != stick1 {
		t.Fatal("Phase 1: Expected nearest stick (stick1)")
	}

	// --- Phase 2: Pick up first stick ---
	result := Pickup(char, stick1, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Phase 2: Expected PickupToInventory, got %d", result)
	}

	// Stick goes to inventory (no vessel path — no variety)
	if len(char.Inventory) != 1 {
		t.Fatalf("Phase 2: Expected 1 item in inventory, got %d", len(char.Inventory))
	}

	// --- Phase 3: Continuation — FindNextGatherTarget finds second stick ---
	cpos := char.Pos()
	items = gameMap.Items()
	nextIntent := FindNextGatherTarget(char, cpos.X, cpos.Y, items, "stick", gameMap)
	if nextIntent == nil {
		t.Fatal("Phase 3: Expected continuation intent for second stick")
	}
	if nextIntent.TargetItem != stick2 {
		t.Fatal("Phase 3: Expected second stick as target")
	}

	// --- Phase 4: Pick up second stick — merges into existing bundle ---
	result = Pickup(char, stick2, gameMap, nil, registry)
	if result != PickupToBundle {
		t.Fatalf("Phase 4: Expected PickupToBundle, got %d", result)
	}

	// --- Phase 5: Continuation — no more sticks on map, signals completion ---
	items = gameMap.Items()
	nextIntent = FindNextGatherTarget(char, cpos.X, cpos.Y, items, "stick", gameMap)
	if nextIntent != nil {
		t.Error("Phase 5: Expected nil (no more sticks on map) — signals order completion")
	}

	// Verify final state: 1 bundle of 2 sticks
	if len(char.Inventory) != 1 {
		t.Errorf("Final: Expected 1 item in inventory (bundle), got %d", len(char.Inventory))
	}
	if char.Inventory[0].BundleCount != 2 {
		t.Errorf("Final: Expected bundle count 2, got %d", char.Inventory[0].BundleCount)
	}
}

// Regression: Character with a vessel in inventory gathers sticks. After picking up
// the first stick, FindNextGatherTarget should find the next stick even though the
// character is carrying a vessel. The original bug was a blanket vessel guard in the
// UI-layer PickupToInventory handler that blocked gather continuation for characters
// with vessels — this test validates the system-layer function works for this scenario.
func TestFindNextGatherTarget_FindsNextStickWithVesselInInventory(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	// Character has a vessel and a full stick bundle
	vessel := createTestVessel()
	fullBundle := entity.NewStick(0, 0)
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{vessel, fullBundle}
	gameMap.AddCharacter(char)

	// Another stick on the ground
	nextStick := entity.NewStick(7, 5)
	gameMap.AddItem(nextStick)

	// Should NOT find next — inventory full and bundle at max
	intent := FindNextGatherTarget(char, 5, 5, gameMap.Items(), "stick", gameMap)
	if intent != nil {
		t.Error("Expected nil when inventory full and bundle at max size")
	}

	// Now with only the vessel (one slot free) — should find next stick
	char.Inventory = []*entity.Item{vessel}
	intent = FindNextGatherTarget(char, 5, 5, gameMap.Items(), "stick", gameMap)
	if intent == nil {
		t.Fatal("Expected intent to gather next stick when inventory has space")
	}
	if intent.TargetItem != nextStick {
		t.Error("Intent should target the next stick on the ground")
	}
}

// Regression: FindNextGatherTarget must return nil when character has a full bundle,
// even if more sticks exist on the map and inventory has space. Without this check,
// the PickupToBundle handler would send the character to pick up more sticks after
// the bundle hit max size, delaying order completion.
func TestFindNextGatherTarget_NilWhenFullBundle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	// Character has a full bundle AND an empty inventory slot
	fullBundle := entity.NewStick(0, 0)
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}
	gameMap.AddCharacter(char)

	// More sticks on the ground
	nextStick := entity.NewStick(7, 5)
	gameMap.AddItem(nextStick)

	intent := FindNextGatherTarget(char, 5, 5, gameMap.Items(), "stick", gameMap)
	if intent != nil {
		t.Error("Expected nil when character has full bundle (one bundle per order)")
	}
}

// Regression: CanMergeIntoBundle returns true when target can merge into an existing
// non-full bundle. Used to skip drop-before-pickup logic — bundle merges don't need
// a free inventory slot.
func TestCanMergeIntoBundle(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	// Character has a non-full stick bundle and a berry (full inventory)
	stickBundle := entity.NewStick(0, 0)
	stickBundle.BundleCount = 3
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Inventory = []*entity.Item{stickBundle, berry}

	// Stick target should merge
	targetStick := entity.NewStick(6, 5)
	if !CanMergeIntoBundle(char, targetStick) {
		t.Error("Expected true — stick can merge into existing non-full bundle")
	}

	// Berry target should NOT merge (not bundleable)
	targetBerry := entity.NewBerry(6, 5, types.ColorRed, false, false)
	if CanMergeIntoBundle(char, targetBerry) {
		t.Error("Expected false — berry is not bundleable")
	}

	// Full bundle should NOT merge
	stickBundle.BundleCount = 6
	if CanMergeIntoBundle(char, targetStick) {
		t.Error("Expected false — bundle is already full")
	}
}

// Regression: DropCompletedBundle drops a full bundle when a gather order completes.
// Without this, the character would keep the full bundle in inventory after completion.
func TestDropCompletedBundle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	fullBundle := entity.NewStick(0, 0)
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}

	order := entity.NewOrder(1, "gather", "stick")

	log := NewActionLog(100)
	DropCompletedBundle(char, order, gameMap, log)

	// Bundle should be removed from inventory
	if char.Inventory[0] != nil {
		t.Error("Expected full bundle to be removed from inventory")
	}

	// Bundle should appear on the map
	found := false
	for _, item := range gameMap.Items() {
		if item.ItemType == "stick" && item.BundleCount == 6 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected full bundle to appear on the map")
	}
}

// Anchor test: Full gather-order flow through one complete bundle (6 sticks).
// Exercises the chain from first pickup through bundle merges to completion with
// bundle drop. Catches regressions in: CanMergeIntoBundle (skip unnecessary drops),
// FindNextGatherTarget (full-bundle check), DropCompletedBundle (drop on complete).
func TestGatherOrder_InventoryPath_FullBundle_EndToEnd(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(20, 20)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	// Start with full inventory of non-stick items
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry2 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Inventory = []*entity.Item{berry1, berry2}
	gameMap.AddCharacter(char)

	// Place 6 sticks at character position (simplify by co-locating)
	sticks := make([]*entity.Item, 6)
	for i := range sticks {
		sticks[i] = entity.NewStick(5, 5)
		gameMap.AddItem(sticks[i])
	}

	order := entity.NewOrder(1, "gather", "stick")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)

	// --- Phase 1: Intent should target first stick (needs to drop a berry) ---
	intent := FindGatherIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil {
		t.Fatal("Phase 1: Expected gather intent despite full inventory")
	}

	// Simulate the drop-before-pickup that applyPickup does for order characters
	Drop(char, gameMap, log)

	// --- Phase 2: Pick up first stick into empty slot ---
	result := Pickup(char, sticks[0], gameMap, log, registry)
	if result != PickupToInventory {
		t.Fatalf("Phase 2: Expected PickupToInventory, got %d", result)
	}

	// --- Phase 3-6: Pick up sticks 2-6, each should merge into bundle ---
	for i := 1; i < 6; i++ {
		// CanMergeIntoBundle should be true — no need to drop the remaining berry
		if !CanMergeIntoBundle(char, sticks[i]) {
			t.Fatalf("Phase %d: Expected CanMergeIntoBundle=true", i+2)
		}

		result = Pickup(char, sticks[i], gameMap, log, registry)
		if result != PickupToBundle {
			t.Fatalf("Phase %d: Expected PickupToBundle, got %d", i+2, result)
		}

		if i < 5 {
			// Not full yet — continuation should find next stick
			nextIntent := FindNextGatherTarget(char, 5, 5, gameMap.Items(), "stick", gameMap)
			if nextIntent == nil {
				t.Fatalf("Phase %d: Expected continuation intent (bundle at %d/6)", i+2, i+1)
			}
		}
	}

	// --- Phase 7: Bundle at 6/6 — FindNextGatherTarget should return nil ---
	nextIntent := FindNextGatherTarget(char, 5, 5, gameMap.Items(), "stick", gameMap)
	if nextIntent != nil {
		t.Error("Phase 7: Expected nil — full bundle, one bundle per order")
	}

	// --- Phase 8: Drop completed bundle and complete order ---
	DropCompletedBundle(char, order, gameMap, log)
	CompleteOrder(char, order, log)

	// Verify: bundle on ground, berry still in inventory, order completed
	bundleOnGround := false
	for _, item := range gameMap.Items() {
		if item.ItemType == "stick" && item.BundleCount == 6 {
			bundleOnGround = true
			break
		}
	}
	if !bundleOnGround {
		t.Error("Final: Expected full bundle on the ground")
	}

	hasBerry := false
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == "berry" {
			hasBerry = true
		}
	}
	if !hasBerry {
		t.Error("Final: Expected berry still in inventory (only one was dropped)")
	}

	if order.Status != entity.OrderCompleted {
		t.Errorf("Final: Expected order completed, got %s", order.Status)
	}
	if char.AssignedOrderID != 0 {
		t.Error("Final: Expected character unassigned from order")
	}
}

// =============================================================================
// Step 1c-ii: Harvest + Gather Bundle Integration
// =============================================================================

// TestFindHarvestIntent_SkipsVesselForVesselExcluded verifies that harvesting a
// vessel-excluded type (grass) skips vessel procurement and targets the grass directly.
func TestFindHarvestIntent_SkipsVesselForVesselExcluded(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Grass plant on the map (growing, non-sprout)
	grass := entity.NewGrass(7, 5)
	gameMap.AddItem(grass)

	// Vessel on the ground — should NOT be targeted for grass harvest
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 5
	gameMap.AddItem(vessel)

	order := entity.NewOrder(1, "harvest", "grass")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindHarvestIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected pickup intent for grass, got nil")
	}
	if intent.TargetItem != grass {
		t.Errorf("Intent should target grass directly (not vessel), got %v", intent.TargetItem)
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Intent.Action: got %v, want ActionPickup", intent.Action)
	}
}

// TestFindHarvestIntent_NilWhenFullBundle verifies that findHarvestIntent returns nil
// when the character already has a full bundle of the target type (safety net).
func TestFindHarvestIntent_NilWhenFullBundle(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	fullBundle := entity.NewGrass(0, 0)
	fullBundle.Plant = nil // In inventory, not growing
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}
	gameMap.AddCharacter(char)

	// More grass on the map
	gameMap.AddItem(entity.NewGrass(7, 5))

	order := entity.NewOrder(1, "harvest", "grass")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	intent := FindHarvestIntentForTest(char, char.Pos(), gameMap.Items(), order, nil, gameMap)
	if intent != nil {
		t.Errorf("Expected nil (full bundle safety net), got intent targeting %v", intent.TargetItem)
	}
}

// TestFindHarvestIntent_UsesVesselForNonExcluded verifies that non-vessel-excluded
// types (berry) still use vessel procurement during harvest.
func TestFindHarvestIntent_UsesVesselForNonExcluded(t *testing.T) {
	t.Parallel()

	// Use explicit registry with known red berry variety (GenerateVarieties is random)
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Berry plant on the map
	berry := entity.NewBerry(7, 5, types.ColorRed, false, false)
	berry.Plant = &entity.PlantProperties{IsGrowing: true}
	gameMap.AddItem(berry)

	// Vessel on the ground — should be targeted for berry harvest (vessel procurement)
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 5
	gameMap.AddItem(vessel)

	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindHarvestIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent for berry harvest, got nil")
	}
	// Should target the vessel (procurement), not the berry directly
	if intent.TargetItem != vessel {
		t.Errorf("Expected vessel procurement intent, got target %v", intent.TargetItem)
	}
}

// TestFindGatherIntent_VesselExcludedWithVariety_SkipsVessel verifies that
// vessel-excluded types skip vessel procurement even when they have a registered variety.
func TestFindGatherIntent_VesselExcludedWithVariety_SkipsVessel(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Grass on the ground (non-growing, gatherable) — grass has a variety in the registry
	grass := entity.NewGrass(7, 5)
	grass.Plant = nil // Not growing — it's a loose item for gathering
	gameMap.AddItem(grass)

	// Vessel on the ground — should NOT be targeted for grass gathering
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 5
	gameMap.AddItem(vessel)

	order := entity.NewOrder(1, "gather", "grass")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	items := gameMap.Items()
	intent := FindGatherIntentForTest(char, char.Pos(), items, order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected pickup intent for grass, got nil")
	}
	if intent.TargetItem != grass {
		t.Errorf("Intent should target grass directly (not vessel), got %v", intent.TargetItem)
	}
}

// TestFindNextHarvestTarget_ReturnsIntentWhenBundleHasRoom verifies that
// FindNextHarvestTarget returns a continuation intent when the character's
// bundle has room for more items.
func TestFindNextHarvestTarget_ReturnsIntentWhenBundleHasRoom(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	// Non-full grass bundle in inventory (fills both slots with bundle + other item)
	grassBundle := entity.NewGrass(0, 0)
	grassBundle.Plant = nil
	grassBundle.BundleCount = 3
	otherItem := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Inventory = []*entity.Item{grassBundle, otherItem}
	gameMap.AddCharacter(char)

	// More grass on the map (growing)
	nextGrass := entity.NewGrass(7, 5)
	gameMap.AddItem(nextGrass)

	intent := FindNextHarvestTarget(char, 5, 5, gameMap.Items(), "grass", gameMap)
	if intent == nil {
		t.Fatal("Expected continuation intent — bundle has room")
	}
	if intent.TargetItem != nextGrass {
		t.Error("Intent should target the next grass on the map")
	}
}

// TestFindNextHarvestTarget_NilWhenBundleFull verifies that FindNextHarvestTarget
// returns nil when the character's bundle is full, even with more targets on the map.
func TestFindNextHarvestTarget_NilWhenBundleFull(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	// Full grass bundle
	fullBundle := entity.NewGrass(0, 0)
	fullBundle.Plant = nil
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}
	gameMap.AddCharacter(char)

	// More grass on the map
	gameMap.AddItem(entity.NewGrass(7, 5))

	intent := FindNextHarvestTarget(char, 5, 5, gameMap.Items(), "grass", gameMap)
	if intent != nil {
		t.Error("Expected nil — bundle is full")
	}
}

// TestIsMultiStepOrderComplete_HarvestWithFullBundle verifies that a harvest order
// is considered complete when the character has a full bundle of the target type.
func TestIsMultiStepOrderComplete_HarvestWithFullBundle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	fullBundle := entity.NewGrass(0, 0)
	fullBundle.Plant = nil
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "harvest", "grass")

	if !IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected harvest order complete when character has full bundle")
	}

	// Non-full bundle should NOT be complete
	fullBundle.BundleCount = 3
	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected harvest order NOT complete with non-full bundle")
	}
}

// TestDropCompletedBundle_HandlesHarvestOrder verifies that DropCompletedBundle
// drops a full bundle for harvest orders (not just gather orders).
func TestDropCompletedBundle_HandlesHarvestOrder(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	fullBundle := entity.NewGrass(0, 0)
	fullBundle.Plant = nil
	fullBundle.BundleCount = 6
	char.Inventory = []*entity.Item{fullBundle, nil}

	order := entity.NewOrder(1, "harvest", "grass")

	log := NewActionLog(100)
	DropCompletedBundle(char, order, gameMap, log)

	// Bundle should be removed from inventory
	if char.Inventory[0] != nil {
		t.Error("Expected full bundle to be removed from inventory")
	}

	// Bundle should appear on the map
	found := false
	for _, item := range gameMap.Items() {
		if item.ItemType == "grass" && item.BundleCount == 6 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected full bundle to appear on the map")
	}
}

// Anchor test: Full harvest-grass-to-bundle flow. Character harvests grass plants
// into a bundle, continues until full (6), drops bundle, order completes.
func TestHarvestGrass_Bundle_EndToEnd(t *testing.T) {
	t.Parallel()

	registry := game.GenerateVarieties()
	gameMap := game.NewMap(20, 20)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Place 6 grass plants at character position
	grasses := make([]*entity.Item, 6)
	for i := range grasses {
		grasses[i] = entity.NewGrass(5, 5)
		gameMap.AddItem(grasses[i])
	}

	order := entity.NewOrder(1, "harvest", "grass")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)

	// --- Phase 1: findHarvestIntent targets nearest grass (no vessel procurement) ---
	intent := FindHarvestIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil {
		t.Fatal("Phase 1: Expected grass harvest intent")
	}
	if intent.Action != entity.ActionPickup {
		t.Fatalf("Phase 1: Expected ActionPickup, got %v", intent.Action)
	}

	// --- Phase 2: Pick up first grass → PickupToInventory (new bundle of 1) ---
	result := Pickup(char, grasses[0], gameMap, log, registry)
	if result != PickupToInventory {
		t.Fatalf("Phase 2: Expected PickupToInventory, got %d", result)
	}
	if char.Inventory[0].BundleCount != 1 {
		t.Fatalf("Phase 2: Expected bundle count 1, got %d", char.Inventory[0].BundleCount)
	}

	// --- Phase 3: Continuation — FindNextHarvestTarget finds next grass ---
	cpos := char.Pos()
	nextIntent := FindNextHarvestTarget(char, cpos.X, cpos.Y, gameMap.Items(), "grass", gameMap)
	if nextIntent == nil {
		t.Fatal("Phase 3: Expected continuation intent for next grass")
	}

	// --- Phase 4-7: Pick up grasses 2-6, each merges into bundle ---
	for i := 1; i < 6; i++ {
		result = Pickup(char, grasses[i], gameMap, log, registry)
		if result != PickupToBundle {
			t.Fatalf("Phase %d: Expected PickupToBundle, got %d", i+3, result)
		}

		if i < 5 {
			// Not full yet — continuation should find next grass
			nextIntent = FindNextHarvestTarget(char, cpos.X, cpos.Y, gameMap.Items(), "grass", gameMap)
			if nextIntent == nil {
				t.Fatalf("Phase %d: Expected continuation (bundle at %d/6)", i+3, i+1)
			}
		}
	}

	// --- Phase 8: Bundle at 6/6 — FindNextHarvestTarget should return nil ---
	nextIntent = FindNextHarvestTarget(char, cpos.X, cpos.Y, gameMap.Items(), "grass", gameMap)
	if nextIntent != nil {
		t.Error("Phase 8: Expected nil — full bundle")
	}

	// --- Phase 9: Drop completed bundle and complete order ---
	DropCompletedBundle(char, order, gameMap, log)
	CompleteOrder(char, order, log)

	// Verify: bundle on ground, order completed
	bundleOnGround := false
	for _, item := range gameMap.Items() {
		if item.ItemType == "grass" && item.BundleCount == 6 {
			bundleOnGround = true
			break
		}
	}
	if !bundleOnGround {
		t.Error("Final: Expected full grass bundle on the ground")
	}
	if order.Status != entity.OrderCompleted {
		t.Errorf("Final: Expected order completed, got %s", order.Status)
	}
	if char.AssignedOrderID != 0 {
		t.Error("Final: Expected character unassigned from order")
	}
}

// Regression: harvest berry still uses vessel path (bundle changes don't affect it).
func TestHarvestBerry_VesselPath_Unchanged(t *testing.T) {
	t.Parallel()

	// Use explicit registry with known red berry variety (GenerateVarieties is random)
	registry := createTestRegistry()
	gameMap := game.NewMap(20, 20)
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Berry plant and vessel on map
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	vessel := createTestVessel()
	vessel.X = 7
	vessel.Y = 5
	gameMap.AddItem(vessel)

	order := entity.NewOrder(1, "harvest", "berry")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	// Phase 1: Intent should target vessel (procurement)
	items := gameMap.Items()
	intent := FindHarvestIntentForTest(char, char.Pos(), items, order, nil, gameMap)
	if intent == nil {
		t.Fatal("Expected vessel procurement intent")
	}
	if intent.TargetItem != vessel {
		t.Fatalf("Expected vessel target, got %v", intent.TargetItem)
	}

	// Phase 2: Pick up vessel
	result := Pickup(char, vessel, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Expected PickupToInventory for vessel, got %d", result)
	}

	// Phase 3: With vessel in inventory, intent should target berry
	items = gameMap.Items()
	intent = FindHarvestIntentForTest(char, char.Pos(), items, order, nil, gameMap)
	if intent == nil {
		t.Fatal("Expected berry harvest intent")
	}
	if intent.TargetItem != berry {
		t.Fatalf("Expected berry target, got %v", intent.TargetItem)
	}

	// Phase 4: Pick up berry → goes into vessel
	result = Pickup(char, berry, gameMap, nil, registry)
	if result != PickupToVessel {
		t.Fatalf("Expected PickupToVessel for berry, got %d", result)
	}
}

// =============================================================================
// findExtractIntent
// =============================================================================

func TestFindExtractIntent_FindsNearestExtractablePlant(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	// Create a vessel and give to character
	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	// Plant far away
	farFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 18, Y: 18, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	// Plant close
	nearFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(farFlower)
	gameMap.AddItem(nearFlower)

	order := entity.NewOrder(1, "extract", "flower")
	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected non-nil intent")
	}
	if intent.Action != entity.ActionExtract {
		t.Errorf("Expected ActionExtract, got %d", intent.Action)
	}
	if intent.TargetItem != nearFlower {
		t.Error("Expected intent to target the nearer flower")
	}
}

func TestFindExtractIntent_BFSStepNotTeleport(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 2, 2, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	// Plant far away (distance > 1)
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 15, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected non-nil intent")
	}

	// Target should be a BFS step (adjacent to character), NOT the plant position
	plantPos := types.Position{X: 15, Y: 15}
	if intent.Target == plantPos {
		t.Error("Target should be a BFS step toward the plant, not the plant position itself (teleport bug)")
	}
	// Dest should be the plant position
	if intent.Dest != plantPos {
		t.Errorf("Dest should be plant position %v, got %v", plantPos, intent.Dest)
	}
	// Target should be close to the character (one step away)
	charPos := types.Position{X: 2, Y: 2}
	dist := intent.Target.DistanceTo(charPos)
	if dist > 2 {
		t.Errorf("Target should be one BFS step from character, but distance is %d", dist)
	}
}

func TestFindExtractIntent_SkipsSeedTimerPositive(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	// Plant with active seed timer (recently extracted)
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 50.0},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when all plants have active SeedTimer")
	}
}

func TestFindExtractIntent_SkipsSprouts(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	sprout := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '🌱', EType: entity.TypeItem},
		ItemType:   "flower",
		Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true, SeedTimer: 0},
	}
	gameMap.AddItem(sprout)

	order := entity.NewOrder(1, "extract", "flower")
	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent for sprouts")
	}
}

func TestFindExtractIntent_VesselProcurement(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Set up variety registry with seed varieties
	registry := game.NewVarietyRegistry()
	registry.Register(&entity.ItemVariety{
		ID:       "flower seed-yellow",
		ItemType: "seed",
		Kind:     "flower seed",
		Color:    types.ColorYellow,
		Sym:      '·',
	})
	gameMap.SetVarieties(registry)

	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	char.Inventory = nil // No vessel
	gameMap.AddCharacter(char)

	// Vessel on ground
	vessel := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 12, Y: 10, Sym: 'U', EType: entity.TypeItem},
		ItemType:   "vessel",
		Container:  &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	gameMap.AddItem(vessel)

	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected vessel procurement intent")
	}
	// Should be ActionPickup targeting the vessel, not the flower
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup for vessel, got %d", intent.Action)
	}
	if intent.TargetItem != vessel {
		t.Error("Expected intent to target the vessel")
	}
}

func TestFindExtractIntent_ReturnsNilWhenNoTargets(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	// No flowers on the map
	order := entity.NewOrder(1, "extract", "flower")
	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when no extractable plants exist")
	}
}

func TestFindExtractIntent_RespectsLockedVariety(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	// Yellow flower (near) and pink flower (far) — both extractable
	yellowFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	pinkFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorPink,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(yellowFlower)
	gameMap.AddItem(pinkFlower)

	// Lock order to pink variety — should skip nearby yellow and target pink
	order := entity.NewOrder(1, "extract", "flower")
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorPink, "", "")

	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected intent targeting pink flower")
	}
	if intent.TargetItem != pinkFlower {
		t.Error("Expected intent to target pink flower (locked variety), not yellow")
	}
}

func TestFindExtractIntent_LockedVarietyNilWhenDepleted(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	vessel := &entity.Item{
		ItemType:  "vessel",
		Container: &entity.ContainerData{Capacity: 1, Contents: []entity.Stack{}},
	}
	char.AddToInventory(vessel)

	// Yellow flower on cooldown, pink flower available
	yellowFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 100},
	}
	pinkFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorPink,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(yellowFlower)
	gameMap.AddItem(pinkFlower)

	// Lock to yellow — should return nil even though pink is available
	order := entity.NewOrder(1, "extract", "flower")
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")

	pos := types.Position{X: char.X, Y: char.Y}
	intent := findExtractIntent(char, pos, gameMap.Items(), order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil intent when locked variety is depleted, even though other varieties are available")
	}
}

func TestExtractOrderCompletion_CompletesWhenLockedVarietyDepleted(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Yellow flower on cooldown, pink flower available
	yellowFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 100},
	}
	pinkFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorPink,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(yellowFlower)
	gameMap.AddItem(pinkFlower)

	order := entity.NewOrder(1, "extract", "flower")
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")

	// Locked to yellow, yellow is depleted → should complete
	if !isMultiStepOrderComplete(char, order, gameMap) {
		t.Error("Expected extract order to complete when locked variety has no extractable plants")
	}
}

func TestExtractOrderCompletion_DoesNotCompleteBeforeLock(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// All flowers on cooldown but no variety lock yet
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 100},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	// LockedVariety is empty — not yet locked

	if isMultiStepOrderComplete(char, order, gameMap) {
		t.Error("Expected extract order NOT to complete before variety is locked (no extraction done yet)")
	}
}

func TestExtractFeasibility_IgnoresLockedVariety(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	// Only pink flower available (yellow on cooldown)
	yellowFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 100},
	}
	pinkFlower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorPink,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(yellowFlower)
	gameMap.AddItem(pinkFlower)

	order := entity.NewOrder(1, "extract", "flower")
	// Feasibility check should pass — pink flower is extractable
	feasible, _ := IsOrderFeasible(order, gameMap.Items(), gameMap)
	if !feasible {
		t.Error("Expected extract order to be feasible when any flower has seeds available")
	}
}

func TestExtractFeasibility_InfeasibleWhenAllOnCooldown(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"extract"}
	gameMap.AddCharacter(char)

	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 11, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 100},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	feasible, _ := IsOrderFeasible(order, gameMap.Items(), gameMap)
	if feasible {
		t.Error("Expected extract order to be infeasible when all flowers are on seed cooldown")
	}
}

// =============================================================================
// Extract order: inventory-full completion
// =============================================================================

func TestExtractOrderCompletion_CompletesWhenInventoryFullNoVessel(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Character has 2 seeds in inventory (full, no vessel)
	seed1 := entity.NewSeed(0, 0, "flower", "flower-yellow", "", types.ColorYellow, "", "")
	seed2 := entity.NewSeed(0, 0, "flower", "flower-yellow", "", types.ColorYellow, "", "")
	char.AddToInventory(seed1)
	char.AddToInventory(seed2)

	// Extractable flower still exists on map
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")

	// Should complete: locked variety set + inventory full + no vessel
	if !isMultiStepOrderComplete(char, order, gameMap) {
		t.Error("Expected extract order to complete when inventory full and no vessel")
	}
}

func TestExtractOrderCompletion_DoesNotCompleteWhenVesselHasSpace(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 10, 10, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Character has a vessel with space for seeds
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

	vessel := &entity.Item{
		ItemType: "vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
	char.AddToInventory(vessel)

	// Extractable flower still exists
	flower := &entity.Item{
		BaseEntity: entity.BaseEntity{X: 15, Y: 10, Sym: '✿', EType: entity.TypeItem},
		ItemType:   "flower",
		Color:      types.ColorYellow,
		Plant:      &entity.PlantProperties{IsGrowing: true, SeedTimer: 0},
	}
	gameMap.AddItem(flower)

	order := entity.NewOrder(1, "extract", "flower")
	order.LockedVariety = entity.GenerateVarietyID("flower", "", types.ColorYellow, "", "")

	// Should NOT complete: vessel can still accept seeds
	if isMultiStepOrderComplete(char, order, gameMap) {
		t.Error("Expected extract order NOT to complete when vessel has space for seeds")
	}
}

// =============================================================================
// Step 3b: findDigIntent
// =============================================================================

func TestFindDigIntent_FindsNearestClayTile(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// Clay tile nearby
	clayPos := types.Position{X: 7, Y: 5}
	gameMap.SetClay(clayPos)

	order := entity.NewOrder(1, "dig", "clay")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	intent := FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, nil, gameMap)

	if intent == nil {
		t.Fatal("Expected dig intent, got nil")
	}
	if intent.Action != entity.ActionDig {
		t.Errorf("Intent.Action: got %v, want ActionDig", intent.Action)
	}
	if intent.Dest != clayPos {
		t.Errorf("Intent.Dest: got %v, want %v (clay tile)", intent.Dest, clayPos)
	}
}

func TestFindDigIntent_NilWhenNoClay(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// No clay tiles
	order := entity.NewOrder(1, "dig", "clay")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	intent := FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil when no clay tiles exist")
	}
}

func TestFindDigIntent_NilWhenInventoryFullOfClay(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// Fill both inventory slots with clay
	clay1 := entity.NewClay(0, 0)
	clay2 := entity.NewClay(0, 0)
	char.Inventory = []*entity.Item{clay1, clay2}

	gameMap.SetClay(types.Position{X: 7, Y: 5})

	order := entity.NewOrder(1, "dig", "clay")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	intent := FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, nil, gameMap)

	if intent != nil {
		t.Error("Expected nil when both inventory slots have clay (triggers completion)")
	}
}

func TestFindDigIntent_DropsInventoryFirst(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// Full inventory with non-clay items
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry2 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Inventory = []*entity.Item{berry1, berry2}

	gameMap.SetClay(types.Position{X: 7, Y: 5})

	order := entity.NewOrder(1, "dig", "clay")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)
	intent := FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)

	// Should return a dig intent (drop happens inside findDigIntent, then proceeds)
	if intent == nil {
		t.Fatal("Expected dig intent after dropping non-clay items")
	}
	if intent.Action != entity.ActionDig {
		t.Errorf("Intent.Action: got %v, want ActionDig", intent.Action)
	}
	// Inventory should have been cleared of non-clay items
	for _, item := range char.Inventory {
		if item != nil && item.ItemType != "clay" {
			t.Errorf("Expected non-clay items to be dropped, but found %s in inventory", item.ItemType)
		}
	}
}

// =============================================================================
// Step 3b: isMultiStepOrderComplete for dig
// =============================================================================

func TestIsMultiStepOrderComplete_Dig(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	clay1 := entity.NewClay(0, 0)
	clay2 := entity.NewClay(0, 0)
	char.Inventory = []*entity.Item{clay1, clay2}
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "dig", "clay")

	if !IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected dig order complete when both inventory slots have clay")
	}
}

func TestIsMultiStepOrderComplete_Dig_OneSlot(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)

	clay1 := entity.NewClay(0, 0)
	char.Inventory = []*entity.Item{clay1, nil}
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "dig", "clay")

	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected dig order NOT complete with only one clay")
	}
}

func TestIsMultiStepOrderComplete_Dig_Empty(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	order := entity.NewOrder(1, "dig", "clay")

	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected dig order NOT complete with empty inventory")
	}
}

// =============================================================================
// Step 3b: DropCompletedDigItems
// =============================================================================

func TestDropCompletedDigItems_DropsAllClay(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	clay1 := entity.NewClay(0, 0)
	clay2 := entity.NewClay(0, 0)
	char.Inventory = []*entity.Item{clay1, clay2}

	order := entity.NewOrder(1, "dig", "clay")

	log := NewActionLog(100)
	DropCompletedDigItems(char, order, gameMap, log)

	// Both clay items should be dropped
	for i, item := range char.Inventory {
		if item != nil && item.ItemType == "clay" {
			t.Errorf("Expected clay slot %d to be nil after drop, but got clay item", i)
		}
	}

	// Clay items should appear on the map
	clayOnMap := 0
	for _, item := range gameMap.Items() {
		if item.ItemType == "clay" {
			clayOnMap++
		}
	}
	if clayOnMap != 2 {
		t.Errorf("Expected 2 clay items on the map, got %d", clayOnMap)
	}
}

// =============================================================================
// Step 3b: IsOrderFeasible for dig
// =============================================================================

func TestIsOrderFeasible_Dig_ClayExists(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	gameMap.SetClay(types.Position{X: 7, Y: 5})

	order := entity.NewOrder(1, "dig", "clay")
	feasible, noKnowHow := IsOrderFeasible(order, gameMap.Items(), gameMap)

	if !feasible {
		t.Error("Expected dig order to be feasible when clay tiles exist")
	}
	if noKnowHow {
		t.Error("Expected noKnowHow=false when character knows dig")
	}
}

func TestIsOrderFeasible_Dig_NoClay(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// No clay tiles
	order := entity.NewOrder(1, "dig", "clay")
	feasible, noKnowHow := IsOrderFeasible(order, gameMap.Items(), gameMap)

	if feasible {
		t.Error("Expected dig order to be infeasible when no clay tiles exist")
	}
	if noKnowHow {
		t.Error("Expected noKnowHow=false (clay absence, not know-how absence)")
	}
}

// =============================================================================
// Step 3b: TestDigOrder_EndToEnd
// =============================================================================

// TestDigOrder_EndToEnd traces the full dig flow: drop inventory → walk to clay →
// dig → walk to next clay → dig → completion check → drop both clay.
// Follows TestGatherOrder_InventoryPath_FullBundle_EndToEnd pattern.
func TestDigOrder_EndToEnd(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"dig"}
	gameMap.AddCharacter(char)

	// Start with a non-clay item in inventory
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Inventory = []*entity.Item{berry}

	// Place two clay tiles at character position (simplify by co-locating)
	clay1Pos := types.Position{X: 5, Y: 5}
	clay2Pos := types.Position{X: 6, Y: 5}
	gameMap.SetClay(clay1Pos)
	gameMap.SetClay(clay2Pos)

	order := entity.NewOrder(1, "dig", "clay")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)

	// --- Phase 1: findDigIntent drops non-clay inventory, returns dig intent ---
	intent := FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil {
		t.Fatal("Phase 1: Expected dig intent, got nil")
	}
	if intent.Action != entity.ActionDig {
		t.Fatalf("Phase 1: Expected ActionDig, got %v", intent.Action)
	}
	// Berry should have been dropped
	for _, item := range char.Inventory {
		if item != nil && item.ItemType == "berry" {
			t.Error("Phase 1: Berry should have been dropped before digging")
		}
	}

	// --- Phase 2: Simulate walking to clay tile and digging (add clay to inventory) ---
	clay1 := entity.NewClay(clay1Pos.X, clay1Pos.Y)
	char.AddToInventory(clay1)

	// --- Phase 3: findDigIntent should not return nil yet (only 1 clay) ---
	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Fatal("Phase 3: Order should not be complete with only 1 clay")
	}

	// findDigIntent should find next clay tile
	intent = FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil {
		t.Fatal("Phase 3: Expected second dig intent")
	}

	// --- Phase 4: Dig second clay ---
	clay2 := entity.NewClay(clay2Pos.X, clay2Pos.Y)
	char.AddToInventory(clay2)

	// --- Phase 5: findDigIntent returns nil (both slots have clay) ---
	intent = FindDigIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent != nil {
		t.Fatal("Phase 5: Expected nil intent when both slots have clay")
	}

	// --- Phase 6: isMultiStepOrderComplete returns true ---
	if !IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Fatal("Phase 6: Expected order complete with 2 clay items")
	}

	// --- Phase 7: Drop both clay on completion ---
	DropCompletedDigItems(char, order, gameMap, log)
	CompleteOrder(char, order, log)

	// Verify: 2 clay on map, inventory empty, order completed
	clayOnMap := 0
	for _, item := range gameMap.Items() {
		if item.ItemType == "clay" {
			clayOnMap++
		}
	}
	if clayOnMap != 2 {
		t.Errorf("Final: Expected 2 clay items on map, got %d", clayOnMap)
	}
	for i, item := range char.Inventory {
		if item != nil && item.ItemType == "clay" {
			t.Errorf("Final: Expected clay slot %d cleared, but still has clay", i)
		}
	}
	if order.Status != entity.OrderCompleted {
		t.Errorf("Final: Expected order completed, got %s", order.Status)
	}
	if char.AssignedOrderID != 0 {
		t.Error("Final: Expected character unassigned from order")
	}
}

// =============================================================================
// Step 4a: IsOrderFeasible for craftBrick
// =============================================================================

func TestIsOrderFeasible_CraftBrick_ClayExists(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftBrick"}
	char.KnownRecipes = []string{"clay-brick"}
	gameMap.AddCharacter(char)

	clay := entity.NewClay(3, 3)
	clay.ID = gameMap.NextItemID()
	gameMap.AddItem(clay)

	order := entity.NewOrder(1, "craftBrick", "")
	feasible, noKnowHow := IsOrderFeasible(order, gameMap.Items(), gameMap)

	if !feasible {
		t.Error("craftBrick order should be feasible when clay exists on the ground")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false when character knows craftBrick")
	}
}

func TestIsOrderFeasible_CraftBrick_NoClay(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftBrick"}
	char.KnownRecipes = []string{"clay-brick"}
	gameMap.AddCharacter(char)

	// No clay on map
	order := entity.NewOrder(1, "craftBrick", "")
	feasible, noKnowHow := IsOrderFeasible(order, gameMap.Items(), gameMap)

	if feasible {
		t.Error("craftBrick order should be unfeasible when no clay exists")
	}
	if noKnowHow {
		t.Error("noKnowHow should be false (clay missing, not know-how missing)")
	}
}

func TestIsOrderFeasible_CraftBrick_NoKnowHow(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	// Character has no craftBrick know-how
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	clay := entity.NewClay(3, 3)
	clay.ID = gameMap.NextItemID()
	gameMap.AddItem(clay)

	order := entity.NewOrder(1, "craftBrick", "")
	feasible, noKnowHow := IsOrderFeasible(order, gameMap.Items(), gameMap)

	if feasible {
		t.Error("craftBrick order should be unfeasible when no character knows craftBrick")
	}
	if !noKnowHow {
		t.Error("noKnowHow should be true when clay exists but no character knows craftBrick")
	}
}

// =============================================================================
// Step 4b: findCraftIntent for craftBrick
// =============================================================================

func TestFindCraftIntent_CraftBrick_FindsClayOnGround(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftBrick"}
	char.KnownRecipes = []string{"clay-brick"}
	gameMap.AddCharacter(char)

	clay := entity.NewClay(7, 5)
	clay.ID = gameMap.NextItemID()
	gameMap.AddItem(clay)

	order := entity.NewOrder(1, "craftBrick", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)
	intent := FindCraftIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)

	if intent == nil {
		t.Fatal("Expected pickup intent targeting clay, got nil")
	}
	if intent.Action != entity.ActionPickup {
		t.Errorf("Expected ActionPickup, got %v", intent.Action)
	}
	if intent.TargetItem != clay {
		t.Errorf("Expected TargetItem to be clay, got %v", intent.TargetItem)
	}
}

func TestFindCraftIntent_CraftBrick_ReturnsActionCraft(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftBrick"}
	char.KnownRecipes = []string{"clay-brick"}
	gameMap.AddCharacter(char)

	// Character already has clay in inventory
	clay := entity.NewClay(5, 5)
	clay.ID = gameMap.NextItemID()
	char.AddToInventory(clay)

	order := entity.NewOrder(1, "craftBrick", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)
	intent := FindCraftIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)

	if intent == nil {
		t.Fatal("Expected ActionCraft intent, got nil")
	}
	if intent.Action != entity.ActionCraft {
		t.Errorf("Expected ActionCraft, got %v", intent.Action)
	}
	if intent.RecipeID != "clay-brick" {
		t.Errorf("Expected RecipeID 'clay-brick', got %q", intent.RecipeID)
	}
}

// =============================================================================
// Step 4b: isMultiStepOrderComplete for craftBrick
// =============================================================================

func TestIsMultiStepOrderComplete_CraftBrick_ClayExists(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Clay on the ground — order not complete
	clay := entity.NewClay(3, 3)
	clay.ID = gameMap.NextItemID()
	gameMap.AddItem(clay)

	order := entity.NewOrder(1, "craftBrick", "")

	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected craftBrick order NOT complete when clay exists on ground")
	}
}

func TestIsMultiStepOrderComplete_CraftBrick_NoClay(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// No clay on ground — order complete
	order := entity.NewOrder(1, "craftBrick", "")

	if !IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Error("Expected craftBrick order complete when no clay on ground")
	}
}

// =============================================================================
// Step 4b: TestCraftBrickOrder_EndToEnd
// =============================================================================

// TestCraftBrickOrder_EndToEnd traces the full brick craft flow:
// character picks up clay → crafts brick (brick appears on map) →
// picks up next clay → crafts another brick → no more clay →
// isMultiStepOrderComplete true → CompleteOrder.
// Follows TestGatherOrder_InventoryPath_FullBundle_EndToEnd pattern.
func TestCraftBrickOrder_EndToEnd(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	char.KnownActivities = []string{"craftBrick"}
	char.KnownRecipes = []string{"clay-brick"}
	gameMap.AddCharacter(char)

	// Place two clay items on the ground
	clay1 := entity.NewClay(5, 5)
	clay1.ID = gameMap.NextItemID()
	gameMap.AddItem(clay1)
	clay2 := entity.NewClay(6, 5)
	clay2.ID = gameMap.NextItemID()
	gameMap.AddItem(clay2)

	order := entity.NewOrder(1, "craftBrick", "")
	order.Status = entity.OrderAssigned
	order.AssignedTo = char.ID
	char.AssignedOrderID = order.ID

	log := NewActionLog(100)

	// --- Phase 1: findCraftIntent returns pickup targeting first clay ---
	intent := FindCraftIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil || intent.Action != entity.ActionPickup {
		t.Fatalf("Phase 1: Expected pickup intent, got %v", intent)
	}

	// --- Phase 2: Pick up first clay ---
	registry := game.GenerateVarieties()
	gameMap.SetVarieties(registry)
	result := Pickup(char, clay1, gameMap, log, registry)
	if result == PickupFailed {
		t.Fatal("Phase 2: Expected successful pickup of clay1")
	}
	gameMap.RemoveItem(clay1)

	// --- Phase 3: findCraftIntent returns ActionCraft (clay in inventory) ---
	intent = FindCraftIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil || intent.Action != entity.ActionCraft {
		t.Fatalf("Phase 3: Expected ActionCraft intent, got %v", intent)
	}
	if intent.RecipeID != "clay-brick" {
		t.Errorf("Phase 3: Expected RecipeID 'clay-brick', got %q", intent.RecipeID)
	}

	// --- Phase 4: Simulate craft — consume clay, produce brick ---
	consumed := char.ConsumeAccessibleItem("clay")
	if consumed == nil {
		t.Fatal("Phase 4: Expected clay to be consumed from inventory")
	}
	brick1 := CreateBrick(consumed, entity.RecipeRegistry["clay-brick"])
	brick1.X, brick1.Y = char.X, char.Y
	brick1.ID = gameMap.NextItemID()
	gameMap.AddItem(brick1)

	// Order should NOT be complete — clay2 still on ground
	if IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Fatal("Phase 4: Order should not be complete while clay2 is on ground")
	}

	// --- Phase 5: findCraftIntent returns pickup targeting clay2 ---
	intent = FindCraftIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent == nil || intent.Action != entity.ActionPickup {
		t.Fatalf("Phase 5: Expected pickup intent for clay2, got %v", intent)
	}
	if intent.TargetItem != clay2 {
		t.Errorf("Phase 5: Expected TargetItem clay2, got %v", intent.TargetItem)
	}

	// --- Phase 6: Pick up clay2 ---
	result = Pickup(char, clay2, gameMap, log, registry)
	if result == PickupFailed {
		t.Fatal("Phase 6: Expected successful pickup of clay2")
	}
	gameMap.RemoveItem(clay2)

	// --- Phase 7: Craft second brick ---
	consumed = char.ConsumeAccessibleItem("clay")
	if consumed == nil {
		t.Fatal("Phase 7: Expected clay2 to be consumed")
	}
	brick2 := CreateBrick(consumed, entity.RecipeRegistry["clay-brick"])
	brick2.X, brick2.Y = char.X, char.Y
	brick2.ID = gameMap.NextItemID()
	gameMap.AddItem(brick2)

	// --- Phase 8: No more clay — findCraftIntent returns nil ---
	intent = FindCraftIntentForTest(char, char.Pos(), gameMap.Items(), order, log, gameMap)
	if intent != nil {
		t.Error("Phase 8: Expected nil intent when no clay remains")
	}

	// --- Phase 9: isMultiStepOrderComplete returns true ---
	if !IsMultiStepOrderCompleteForTest(char, order, gameMap) {
		t.Fatal("Phase 9: Expected order complete when no clay on ground")
	}

	// --- Phase 10: CompleteOrder ---
	CompleteOrder(char, order, log)

	// Verify: 2 bricks on map, order completed
	brickCount := 0
	for _, item := range gameMap.Items() {
		if item.ItemType == "brick" {
			brickCount++
		}
	}
	if brickCount != 2 {
		t.Errorf("Final: Expected 2 bricks on map, got %d", brickCount)
	}
	if order.Status != entity.OrderCompleted {
		t.Errorf("Final: Expected order completed, got %s", order.Status)
	}
	if char.AssignedOrderID != 0 {
		t.Error("Final: Expected character unassigned from order")
	}
}
