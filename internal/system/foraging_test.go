package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// createTestVessel creates an empty vessel for testing
func createTestVessel() *entity.Item {
	return &entity.Item{
		ItemType: "vessel",
		Name:     "Test Vessel",
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
}

// createTestRegistry creates a registry with common test varieties
func createTestRegistry() *game.VarietyRegistry {
	reg := game.NewVarietyRegistry()
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorRed, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   true,
	})
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorBlue, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorBlue,
		Edible:   true,
	})
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("mushroom", types.ColorBrown, types.PatternSpotted, types.TextureSlimy),
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
		Edible:   true,
	})
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("gourd", types.ColorGreen, types.PatternStriped, types.TextureWarty),
		ItemType: "gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternStriped,
		Texture:  types.TextureWarty,
		Edible:   true,
	})
	return reg
}

func TestAddToVessel_EmptyVessel(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)

	added := AddToVessel(vessel, berry, registry)

	if !added {
		t.Error("AddToVessel should return true for empty vessel")
	}
	if len(vessel.Container.Contents) != 1 {
		t.Errorf("vessel should have 1 stack, got %d", len(vessel.Container.Contents))
	}
	if vessel.Container.Contents[0].Count != 1 {
		t.Errorf("stack count should be 1, got %d", vessel.Container.Contents[0].Count)
	}
	if vessel.Container.Contents[0].Variety.ItemType != "berry" {
		t.Errorf("stack variety should be berry, got %s", vessel.Container.Contents[0].Variety.ItemType)
	}
	if vessel.Container.Contents[0].Variety.Color != types.ColorRed {
		t.Errorf("stack variety color should be red, got %s", vessel.Container.Contents[0].Variety.Color)
	}
}

func TestAddToVessel_SameVariety(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add first berry
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry1, registry)

	// Add second berry of same variety
	berry2 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	added := AddToVessel(vessel, berry2, registry)

	if !added {
		t.Error("AddToVessel should return true for same variety")
	}
	if len(vessel.Container.Contents) != 1 {
		t.Errorf("vessel should still have 1 stack, got %d", len(vessel.Container.Contents))
	}
	if vessel.Container.Contents[0].Count != 2 {
		t.Errorf("stack count should be 2, got %d", vessel.Container.Contents[0].Count)
	}
}

func TestAddToVessel_DifferentVariety(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add red berry
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry1, registry)

	// Try to add blue berry (different variety)
	berry2 := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	added := AddToVessel(vessel, berry2, registry)

	if added {
		t.Error("AddToVessel should return false for different variety")
	}
	if len(vessel.Container.Contents) != 1 {
		t.Errorf("vessel should still have 1 stack, got %d", len(vessel.Container.Contents))
	}
	if vessel.Container.Contents[0].Count != 1 {
		t.Errorf("stack count should still be 1, got %d", vessel.Container.Contents[0].Count)
	}
}

func TestAddToVessel_StackFull(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	stackSize := config.GetStackSize("berry") // 20

	// Fill the stack
	for i := 0; i < stackSize; i++ {
		berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
		added := AddToVessel(vessel, berry, registry)
		if !added {
			t.Errorf("AddToVessel should succeed for item %d", i+1)
		}
	}

	// Try to add one more
	extraBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	added := AddToVessel(vessel, extraBerry, registry)

	if added {
		t.Error("AddToVessel should return false when stack is full")
	}
	if vessel.Container.Contents[0].Count != stackSize {
		t.Errorf("stack count should be %d, got %d", stackSize, vessel.Container.Contents[0].Count)
	}
}

func TestAddToVessel_GourdStackSize(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Gourds have stack size of 1
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	added := AddToVessel(vessel, gourd, registry)

	if !added {
		t.Error("AddToVessel should succeed for first gourd")
	}

	// Try to add second gourd - should fail (stack size is 1)
	gourd2 := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	added = AddToVessel(vessel, gourd2, registry)

	if added {
		t.Error("AddToVessel should fail for second gourd (stack size is 1)")
	}
}

func TestAddToVessel_NilContainer(t *testing.T) {
	// Item without container
	notVessel := entity.NewBerry(0, 0, types.ColorRed, false, false)
	registry := createTestRegistry()

	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	added := AddToVessel(notVessel, berry, registry)

	if added {
		t.Error("AddToVessel should return false for non-container item")
	}
}

func TestAddToVessel_NilRegistry(t *testing.T) {
	vessel := createTestVessel()
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)

	added := AddToVessel(vessel, berry, nil)

	if added {
		t.Error("AddToVessel should return false with nil registry")
	}
}

func TestAddToVessel_VarietyNotInRegistry(t *testing.T) {
	vessel := createTestVessel()
	registry := game.NewVarietyRegistry() // Empty registry

	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	added := AddToVessel(vessel, berry, registry)

	if added {
		t.Error("AddToVessel should return false when variety not in registry")
	}
}

func TestIsVesselFull_Empty(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	if IsVesselFull(vessel, registry) {
		t.Error("empty vessel should not be full")
	}
}

func TestIsVesselFull_PartiallyFilled(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry, registry)

	if IsVesselFull(vessel, registry) {
		t.Error("vessel with 1 berry should not be full (stack size is 20)")
	}
}

func TestIsVesselFull_StackFull(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Fill with berries (stack size 20)
	for i := 0; i < 20; i++ {
		berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
		AddToVessel(vessel, berry, registry)
	}

	if !IsVesselFull(vessel, registry) {
		t.Error("vessel should be full when stack is at capacity")
	}
}

func TestIsVesselFull_GourdStack(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add one gourd (stack size 1)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	AddToVessel(vessel, gourd, registry)

	if !IsVesselFull(vessel, registry) {
		t.Error("vessel should be full with one gourd (stack size is 1)")
	}
}

// =============================================================================
// CanPickUpMore Tests
// =============================================================================

func TestCanPickUpMore_EmptyInventory(t *testing.T) {
	registry := createTestRegistry()
	char := &entity.Character{Carrying: nil}

	if !CanPickUpMore(char, registry) {
		t.Error("Should be able to pick up with empty inventory")
	}
}

func TestCanPickUpMore_CarryingNonVessel(t *testing.T) {
	registry := createTestRegistry()
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char := &entity.Character{Carrying: berry}

	if CanPickUpMore(char, registry) {
		t.Error("Should not be able to pick up when carrying non-vessel")
	}
}

func TestCanPickUpMore_CarryingEmptyVessel(t *testing.T) {
	registry := createTestRegistry()
	vessel := createTestVessel()
	char := &entity.Character{Carrying: vessel}

	if !CanPickUpMore(char, registry) {
		t.Error("Should be able to pick up with empty vessel")
	}
}

func TestCanPickUpMore_CarryingFullVessel(t *testing.T) {
	registry := createTestRegistry()
	vessel := createTestVessel()

	// Fill with gourd (stack size 1)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	AddToVessel(vessel, gourd, registry)

	char := &entity.Character{Carrying: vessel}

	if CanPickUpMore(char, registry) {
		t.Error("Should not be able to pick up with full vessel")
	}
}

// =============================================================================
// FindNextVesselTarget Tests
// =============================================================================

func TestFindNextVesselTarget_EmptyVessel(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	char := &entity.Character{Carrying: vessel}
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorRed, false, false),
	}

	intent := FindNextVesselTarget(char, 0, 0, items, registry)

	if intent != nil {
		t.Error("FindNextVesselTarget should return nil for empty vessel")
	}
}

func TestFindNextVesselTarget_FindsMatchingVariety(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add a red berry to vessel
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry, registry)

	char := &entity.Character{Carrying: vessel}

	// Create items on map - one matching, one different variety
	redBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	blueBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)
	items := []*entity.Item{blueBerry, redBerry}

	intent := FindNextVesselTarget(char, 0, 0, items, registry)

	if intent == nil {
		t.Fatal("FindNextVesselTarget should find matching berry")
	}
	if intent.TargetItem != redBerry {
		t.Error("FindNextVesselTarget should target the red berry, not blue")
	}
}

func TestFindNextVesselTarget_IgnoresNonGrowing(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add a red berry to vessel
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry, registry)

	char := &entity.Character{Carrying: vessel}

	// Create a dropped (non-growing) berry
	droppedBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry}

	intent := FindNextVesselTarget(char, 0, 0, items, registry)

	if intent != nil {
		t.Error("FindNextVesselTarget should ignore non-growing items")
	}
}

func TestFindNextVesselTarget_VesselFull(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Fill vessel with gourds (stack size 1)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	AddToVessel(vessel, gourd, registry)

	char := &entity.Character{Carrying: vessel}

	// Another gourd on map
	gourd2 := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty)
	items := []*entity.Item{gourd2}

	intent := FindNextVesselTarget(char, 0, 0, items, registry)

	if intent != nil {
		t.Error("FindNextVesselTarget should return nil when vessel is full")
	}
}

// =============================================================================
// Pickup with Vessel Tests
// =============================================================================

func TestPickup_AddsToVessel(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)

	vessel := createTestVessel()
	char := &entity.Character{
		ID:       1,
		Name:     "Test",
		Carrying: vessel,
	}

	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	result := Pickup(char, berry, gameMap, nil, registry)

	if result != PickupToVessel {
		t.Error("Pickup should return PickupToVessel when adding to vessel")
	}
	if char.Carrying != vessel {
		t.Error("Character should still be carrying vessel")
	}
	if len(vessel.Container.Contents) != 1 {
		t.Error("Vessel should have 1 stack")
	}
	if vessel.Container.Contents[0].Count != 1 {
		t.Error("Stack should have count 1")
	}
	if char.Intent != nil {
		t.Error("Intent should NOT be cleared when adding to vessel")
	}
}

func TestPickup_ToInventory(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)

	char := &entity.Character{
		ID:       1,
		Name:     "Test",
		Carrying: nil,
	}

	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	result := Pickup(char, berry, gameMap, nil, registry)

	if result != PickupToInventory {
		t.Error("Pickup should return PickupToInventory when not carrying vessel")
	}
	if char.Carrying != berry {
		t.Error("Character should be carrying the berry")
	}
	if char.Intent != nil {
		t.Error("Intent should be cleared")
	}
}
