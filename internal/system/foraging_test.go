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
		Edible: &entity.EdibleProperties{},
	})
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("berry", types.ColorBlue, types.PatternNone, types.TextureNone),
		ItemType: "berry",
		Color:    types.ColorBlue,
		Edible: &entity.EdibleProperties{},
	})
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("mushroom", types.ColorBrown, types.PatternSpotted, types.TextureSlimy),
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
		Edible: &entity.EdibleProperties{},
	})
	reg.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("gourd", types.ColorGreen, types.PatternStriped, types.TextureWarty),
		ItemType: "gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternStriped,
		Texture:  types.TextureWarty,
		Edible: &entity.EdibleProperties{},
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
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	added := AddToVessel(vessel, gourd, registry)

	if !added {
		t.Error("AddToVessel should succeed for first gourd")
	}

	// Try to add second gourd - should fail (stack size is 1)
	gourd2 := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
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
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
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
	char := &entity.Character{Inventory: []*entity.Item{}}

	if !CanPickUpMore(char, registry) {
		t.Error("Should be able to pick up with empty inventory")
	}
}

func TestCanPickUpMore_OneSlotUsed(t *testing.T) {
	registry := createTestRegistry()
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char := &entity.Character{Inventory: []*entity.Item{berry}}

	if !CanPickUpMore(char, registry) {
		t.Error("Should be able to pick up with one slot used (has second slot)")
	}
}

func TestCanPickUpMore_FullInventoryNoVessel(t *testing.T) {
	registry := createTestRegistry()
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry2 := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	char := &entity.Character{Inventory: []*entity.Item{berry1, berry2}}

	if CanPickUpMore(char, registry) {
		t.Error("Should not be able to pick up when both slots full with non-vessels")
	}
}

func TestCanPickUpMore_CarryingEmptyVessel(t *testing.T) {
	registry := createTestRegistry()
	vessel := createTestVessel()
	char := &entity.Character{Inventory: []*entity.Item{vessel}}

	if !CanPickUpMore(char, registry) {
		t.Error("Should be able to pick up with empty vessel")
	}
}

func TestCanPickUpMore_FullInventoryWithEmptyVessel(t *testing.T) {
	registry := createTestRegistry()
	vessel := createTestVessel()
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char := &entity.Character{Inventory: []*entity.Item{vessel, berry}}

	if !CanPickUpMore(char, registry) {
		t.Error("Should be able to pick up when full but vessel has space")
	}
}

func TestCanPickUpMore_FullInventoryWithFullVessel(t *testing.T) {
	registry := createTestRegistry()
	vessel := createTestVessel()

	// Fill with gourd (stack size 1)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	AddToVessel(vessel, gourd, registry)

	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char := &entity.Character{Inventory: []*entity.Item{vessel, berry}}

	if CanPickUpMore(char, registry) {
		t.Error("Should not be able to pick up when full and vessel full")
	}
}

// =============================================================================
// FindNextVesselTarget Tests
// =============================================================================

func TestFindNextVesselTarget_EmptyVessel(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	char := &entity.Character{Inventory: []*entity.Item{vessel}}
	items := []*entity.Item{
		entity.NewBerry(5, 5, types.ColorRed, false, false),
	}

	intent := FindNextVesselTarget(char, 0, 0, items, registry, nil)

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

	char := &entity.Character{Inventory: []*entity.Item{vessel}}

	// Create items on map - one matching, one different variety
	redBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	blueBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)
	items := []*entity.Item{blueBerry, redBerry}

	intent := FindNextVesselTarget(char, 0, 0, items, registry, nil)

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

	char := &entity.Character{Inventory: []*entity.Item{vessel}}

	// Create a dropped (non-growing) berry
	droppedBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry}

	intent := FindNextVesselTarget(char, 0, 0, items, registry, nil)

	if intent != nil {
		t.Error("FindNextVesselTarget should ignore non-growing items")
	}
}

func TestFindNextVesselTarget_VesselFull(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Fill vessel with gourds (stack size 1)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	AddToVessel(vessel, gourd, registry)

	char := &entity.Character{Inventory: []*entity.Item{vessel}}

	// Another gourd on map
	gourd2 := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	items := []*entity.Item{gourd2}

	intent := FindNextVesselTarget(char, 0, 0, items, registry, nil)

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
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}

	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	result := Pickup(char, berry, gameMap, nil, registry)

	if result != PickupToVessel {
		t.Error("Pickup should return PickupToVessel when adding to vessel")
	}
	if char.GetCarriedVessel() != vessel {
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
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}

	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	result := Pickup(char, berry, gameMap, nil, registry)

	if result != PickupToInventory {
		t.Error("Pickup should return PickupToInventory when not carrying vessel")
	}
	if len(char.Inventory) != 1 || char.Inventory[0] != berry {
		t.Error("Character should have berry in inventory")
	}
	if char.Intent != nil {
		t.Error("Intent should be cleared")
	}
}

func TestPickup_UsesSecondVesselWhenFirstHasVarietyMismatch(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)

	// First vessel has mushrooms (variety mismatch for gourd)
	// Registry has: mushroom (brown, spotted, slimy)
	mushroomVessel := createTestVessel()
	mushroom := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)
	AddToVessel(mushroomVessel, mushroom, registry)

	// Second vessel is empty (can accept gourd)
	emptyVessel := createTestVessel()

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{mushroomVessel, emptyVessel},
	}

	// Try to pick up a gourd - should go into empty vessel, not fail
	// Registry has: gourd (green, striped, warty)
	gourd := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	gameMap.AddItem(gourd)

	result := Pickup(char, gourd, gameMap, nil, registry)

	if result != PickupToVessel {
		t.Error("Pickup should succeed by using empty vessel when first vessel has variety mismatch")
	}
	if len(emptyVessel.Container.Contents) != 1 {
		t.Error("Empty vessel should now contain the gourd")
	}
	if emptyVessel.Container.Contents[0].Variety.ItemType != "gourd" {
		t.Error("Empty vessel should contain gourd type")
	}
}

// =============================================================================
// CanVesselAccept Tests
// =============================================================================

func TestCanVesselAccept_EmptyVessel(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)

	if !CanVesselAccept(vessel, berry, registry) {
		t.Error("Empty vessel should accept any item")
	}
}

func TestCanVesselAccept_MatchingVariety(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add a red berry to vessel
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry1, registry)

	// Same variety should be accepted
	berry2 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	if !CanVesselAccept(vessel, berry2, registry) {
		t.Error("Vessel should accept matching variety")
	}
}

func TestCanVesselAccept_DifferentVariety(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Add a red berry to vessel
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, berry1, registry)

	// Different variety should be rejected
	berry2 := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	if CanVesselAccept(vessel, berry2, registry) {
		t.Error("Vessel should reject different variety")
	}
}

func TestCanVesselAccept_FullVessel(t *testing.T) {
	vessel := createTestVessel()
	registry := createTestRegistry()

	// Fill vessel with gourds (stack size 1)
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	AddToVessel(vessel, gourd, registry)

	// Same variety should be rejected when full
	gourd2 := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	if CanVesselAccept(vessel, gourd2, registry) {
		t.Error("Full vessel should reject items even if variety matches")
	}
}

// =============================================================================
// FindAvailableVessel Tests
// =============================================================================

func TestFindAvailableVessel_FindsEmptyVessel(t *testing.T) {
	registry := createTestRegistry()
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 3

	items := []*entity.Item{berry, vessel}

	found := FindAvailableVessel(0, 0, items, berry, registry)
	if found != vessel {
		t.Error("Should find empty vessel")
	}
}

func TestFindAvailableVessel_FindsMatchingVessel(t *testing.T) {
	registry := createTestRegistry()

	// Berry to pick up
	targetBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Vessel with same variety
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 3
	existingBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, existingBerry, registry)

	items := []*entity.Item{targetBerry, vessel}

	found := FindAvailableVessel(0, 0, items, targetBerry, registry)
	if found != vessel {
		t.Error("Should find vessel with matching variety")
	}
}

func TestFindAvailableVessel_SkipsIncompatibleVessel(t *testing.T) {
	registry := createTestRegistry()

	// Berry to pick up
	targetBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Vessel with different variety
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 3
	blueBerry := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	AddToVessel(vessel, blueBerry, registry)

	items := []*entity.Item{targetBerry, vessel}

	found := FindAvailableVessel(0, 0, items, targetBerry, registry)
	if found != nil {
		t.Error("Should not find vessel with incompatible variety")
	}
}

func TestFindAvailableVessel_FindsNearest(t *testing.T) {
	registry := createTestRegistry()
	berry := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Far vessel
	farVessel := createTestVessel()
	farVessel.X = 9
	farVessel.Y = 9

	// Near vessel
	nearVessel := createTestVessel()
	nearVessel.X = 1
	nearVessel.Y = 1

	items := []*entity.Item{berry, farVessel, nearVessel}

	found := FindAvailableVessel(0, 0, items, berry, registry)
	if found != nearVessel {
		t.Error("Should find nearest available vessel")
	}
}

// =============================================================================
// scoreForageItems with Vessel Tests
// =============================================================================

func TestScoreForageItems_FiltersToVesselVariety(t *testing.T) {
	registry := createTestRegistry()
	char := &entity.Character{ID: 1, Name: "Test"}

	// Vessel with red berries
	vessel := createTestVessel()
	redBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, redBerry, registry)

	// Closer blue berry (should be skipped)
	blueBerry := entity.NewBerry(1, 1, types.ColorBlue, false, false)
	// Farther red berry (should be found)
	targetBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)

	items := []*entity.Item{blueBerry, targetBerry}

	target, _ := scoreForageItems(char, types.Position{X: 0, Y: 0}, items, vessel)

	if target != targetBerry {
		t.Error("Should find red berry matching vessel variety, not closer blue berry")
	}
}

func TestScoreForageItems_IncludesEdibleNonPlantItems(t *testing.T) {
	char := &entity.Character{ID: 1, Name: "Test"}

	// Nut: edible, Plant == nil
	nut := entity.NewNut(3, 3)

	items := []*entity.Item{nut}

	target, _ := scoreForageItems(char, types.Position{X: 0, Y: 0}, items, nil)

	if target != nut {
		t.Error("scoreForageItems should include edible items with Plant == nil (nuts)")
	}
}

func TestScoreForageItems_ExcludesNonEdibleNonPlantItems(t *testing.T) {
	char := &entity.Character{ID: 1, Name: "Test"}

	// Stick: not edible, Plant == nil
	stick := entity.NewStick(3, 3)

	items := []*entity.Item{stick}

	target, _ := scoreForageItems(char, types.Position{X: 0, Y: 0}, items, nil)

	if target != nil {
		t.Error("scoreForageItems should exclude non-edible items with Plant == nil (sticks)")
	}
}

func TestScoreForageItems_NoFilterWhenVesselEmpty(t *testing.T) {
	char := &entity.Character{ID: 1, Name: "Test"}

	// Empty vessel
	vessel := createTestVessel()

	// Closer blue berry (should be found - no variety filter)
	blueBerry := entity.NewBerry(1, 1, types.ColorBlue, false, false)
	// Farther red berry
	redBerry := entity.NewBerry(5, 5, types.ColorRed, false, false)

	items := []*entity.Item{blueBerry, redBerry}

	target, _ := scoreForageItems(char, types.Position{X: 0, Y: 0}, items, vessel)

	if target != blueBerry {
		t.Error("Empty vessel should not filter - should find closest edible item")
	}
}
