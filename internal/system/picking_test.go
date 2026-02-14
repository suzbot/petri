package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// EnsureHasVesselFor Tests
// =============================================================================

func TestEnsureHasVesselFor_AlreadyHasCompatibleVessel(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has an empty vessel
	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}

	// Target item to harvest
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Should return nil - already have compatible vessel
	intent := EnsureHasVesselFor(char, target, nil, gameMap, nil, true, "order")
	if intent != nil {
		t.Error("EnsureHasVesselFor should return nil when already carrying compatible vessel")
	}
}

func TestEnsureHasVesselFor_VesselWithMatchingContents(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has a vessel with red berries (space remaining)
	vessel := createTestVessel()
	existingBerry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	AddToVessel(vessel, existingBerry, registry)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}

	// Target is same variety - vessel can accept
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)

	intent := EnsureHasVesselFor(char, target, nil, gameMap, nil, true, "order")
	if intent != nil {
		t.Error("EnsureHasVesselFor should return nil when vessel can accept target variety")
	}
}

func TestEnsureHasVesselFor_NoVessel_FindsOne(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has no vessel
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 0
	char.Y = 0

	// Target item
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Available vessel on map
	availableVessel := createTestVessel()
	availableVessel.X = 3
	availableVessel.Y = 3
	items := []*entity.Item{target, availableVessel}

	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, true, "order")
	if intent == nil {
		t.Fatal("EnsureHasVesselFor should return intent to pick up vessel")
	}
	if intent.TargetItem != availableVessel {
		t.Error("Intent should target the available vessel")
	}
	if intent.Action != entity.ActionPickup {
		t.Error("Intent action should be ActionPickup")
	}
}

func TestEnsureHasVesselFor_NoVessel_NoneAvailable(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has no vessel
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 0
	char.Y = 0

	// Target item, no vessels on map
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)
	items := []*entity.Item{target}

	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, true, "order")
	if intent != nil {
		t.Error("EnsureHasVesselFor should return nil when no vessels available")
	}
}

func TestEnsureHasVesselFor_IncompatibleVessel_DropAndFind(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has a vessel with blue berries (incompatible with red target)
	vessel := createTestVessel()
	blueBerry := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	AddToVessel(vessel, blueBerry, registry)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 0
	char.Y = 0

	// Target is red berry - incompatible with vessel contents
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Compatible vessel on map
	compatibleVessel := createTestVessel()
	compatibleVessel.X = 3
	compatibleVessel.Y = 3
	items := []*entity.Item{target, compatibleVessel}

	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, true, "order")

	// Should have dropped the incompatible vessel
	if len(char.Inventory) != 0 {
		t.Error("Should have dropped incompatible vessel")
	}

	// Vessel should be on map at character's position
	found := false
	for _, item := range gameMap.Items() {
		if item == vessel && item.X == 0 && item.Y == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Dropped vessel should be on map at character position")
	}

	// Intent should be to pick up compatible vessel
	if intent == nil {
		t.Fatal("Should return intent to pick up compatible vessel")
	}
	if intent.TargetItem != compatibleVessel {
		t.Error("Intent should target the compatible vessel")
	}
}

func TestEnsureHasVesselFor_IncompatibleVessel_NoDropWhenDisabled(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has a vessel with blue berries
	vessel := createTestVessel()
	blueBerry := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	AddToVessel(vessel, blueBerry, registry)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 0
	char.Y = 0

	// Target is red berry - incompatible
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Compatible vessel on map
	compatibleVessel := createTestVessel()
	compatibleVessel.X = 3
	compatibleVessel.Y = 3
	items := []*entity.Item{target, compatibleVessel}

	// dropConflict = false - should not drop
	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, false, "order")

	// Should NOT have dropped the vessel
	if len(char.Inventory) != 1 {
		t.Error("Should not drop vessel when dropConflict is false")
	}

	// Should return nil (can't get vessel without dropping)
	if intent != nil {
		t.Error("Should return nil when can't drop incompatible vessel")
	}
}

func TestEnsureHasVesselFor_FullVessel_DropAndFind(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has a full vessel (gourd stack size is 1)
	vessel := createTestVessel()
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	AddToVessel(vessel, gourd, registry)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 0
	char.Y = 0

	// Target is same variety gourd - but vessel is full
	target := entity.NewGourd(5, 5, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)

	// Empty vessel on map
	emptyVessel := createTestVessel()
	emptyVessel.X = 3
	emptyVessel.Y = 3
	items := []*entity.Item{target, emptyVessel}

	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, true, "order")

	// Should have dropped the full vessel
	if len(char.Inventory) != 0 {
		t.Error("Should have dropped full vessel")
	}

	// Intent should be to pick up empty vessel
	if intent == nil {
		t.Fatal("Should return intent to pick up empty vessel")
	}
	if intent.TargetItem != emptyVessel {
		t.Error("Intent should target the empty vessel")
	}
}

func TestEnsureHasVesselFor_NoInventorySpace(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character with full inventory (no vessel)
	berry1 := entity.NewBerry(0, 0, types.ColorRed, false, false)
	berry2 := entity.NewBerry(0, 0, types.ColorBlue, false, false)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{berry1, berry2},
	}
	char.X = 0
	char.Y = 0

	// Target and available vessel
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 3
	items := []*entity.Item{target, vessel}

	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, true, "order")

	// Should return nil - no space for vessel
	if intent != nil {
		t.Error("Should return nil when no inventory space for vessel")
	}
}

func TestEnsureHasVesselFor_AlreadyAtVessel(t *testing.T) {
	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character at position (3,3)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 3
	char.Y = 3

	// Target item
	target := entity.NewBerry(5, 5, types.ColorRed, false, false)

	// Vessel at same position as character
	vessel := createTestVessel()
	vessel.X = 3
	vessel.Y = 3
	items := []*entity.Item{target, vessel}

	intent := EnsureHasVesselFor(char, target, items, gameMap, nil, true, "order")
	if intent == nil {
		t.Fatal("Should return pickup intent")
	}

	// Dest should equal current position (already there)
	if intent.Dest.X != 3 || intent.Dest.Y != 3 {
		t.Error("Dest should be vessel position (3,3)")
	}
	if intent.Target.X != 3 || intent.Target.Y != 3 {
		t.Error("Target should be current position when already at vessel")
	}
}

// =============================================================================
// findNearestItemByType Tests (moved from order_execution_test.go, now with growingOnly param)
// =============================================================================

func TestFindNearestItemByType_GrowingOnlyTrue_FindsGrowingItems(t *testing.T) {
	t.Parallel()

	growingBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)
	droppedBerry := entity.NewBerry(3, 3, types.ColorBlue, false, false)
	droppedBerry.Plant.IsGrowing = false

	items := []*entity.Item{droppedBerry, growingBerry}

	result := findNearestItemByType(0, 0, items, "berry", true)
	if result != growingBerry {
		t.Error("growingOnly=true should find growing berry, skip dropped")
	}
}

func TestFindNearestItemByType_GrowingOnlyTrue_SkipsNilPlant(t *testing.T) {
	t.Parallel()

	// Stick has Plant == nil
	stick := entity.NewStick(3, 3)
	growingBerry := entity.NewBerry(10, 10, types.ColorRed, false, false)

	items := []*entity.Item{stick, growingBerry}

	result := findNearestItemByType(0, 0, items, "berry", true)
	if result != growingBerry {
		t.Error("growingOnly=true should skip items with nil Plant")
	}
}

func TestFindNearestItemByType_GrowingOnlyFalse_FindsNonGrowingItems(t *testing.T) {
	t.Parallel()

	stick := entity.NewStick(5, 5)
	items := []*entity.Item{stick}

	result := findNearestItemByType(0, 0, items, "stick", false)
	if result != stick {
		t.Error("growingOnly=false should find stick (Plant == nil)")
	}
}

func TestFindNearestItemByType_GrowingOnlyFalse_FindsShells(t *testing.T) {
	t.Parallel()

	shell := entity.NewShell(3, 3, types.ColorSilver)
	items := []*entity.Item{shell}

	result := findNearestItemByType(0, 0, items, "shell", false)
	if result != shell {
		t.Error("growingOnly=false should find shell")
	}
}

func TestFindNearestItemByType_IgnoresWrongType(t *testing.T) {
	t.Parallel()

	stick := entity.NewStick(3, 3)
	items := []*entity.Item{stick}

	result := findNearestItemByType(0, 0, items, "shell", false)
	if result != nil {
		t.Error("Should return nil when no items of requested type exist")
	}
}

func TestFindNearestItemByType_FindsNearest(t *testing.T) {
	t.Parallel()

	farStick := entity.NewStick(10, 10)
	nearStick := entity.NewStick(2, 2)
	items := []*entity.Item{farStick, nearStick}

	result := findNearestItemByType(0, 0, items, "stick", false)
	if result != nearStick {
		t.Error("Should return nearest item")
	}
}

// =============================================================================
// EnsureHasRecipeInputs Tests
// =============================================================================

func TestEnsureHasRecipeInputs_AllInputsInInventory(t *testing.T) {
	gameMap := game.NewMap(10, 10)

	stick := entity.NewStick(0, 0)
	shell := entity.NewShell(0, 0, types.ColorSilver)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{stick, shell},
	}

	recipe := entity.RecipeRegistry["shell-hoe"]

	intent := EnsureHasRecipeInputs(char, recipe, nil, gameMap, nil)
	if intent != nil {
		t.Error("Should return nil when all inputs are in inventory")
	}
}

func TestEnsureHasRecipeInputs_InputAccessibleInContainer(t *testing.T) {
	registry := createTestRegistry()
	// Register shell variety so HasAccessibleItem can find it in vessel
	registry.Register(&entity.ItemVariety{
		ID:       entity.GenerateVarietyID("shell", types.ColorSilver, types.PatternNone, types.TextureNone),
		ItemType: "shell",
		Color:    types.ColorSilver,
	})
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	// Character has a stick in inventory and a shell in a vessel
	stick := entity.NewStick(0, 0)
	vessel := createTestVessel()
	vessel.Container.Contents = []entity.Stack{
		{Variety: &entity.ItemVariety{ItemType: "shell", Color: types.ColorSilver}, Count: 1},
	}
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{stick, vessel},
	}

	recipe := entity.RecipeRegistry["shell-hoe"]

	intent := EnsureHasRecipeInputs(char, recipe, nil, gameMap, nil)
	if intent != nil {
		t.Error("Should return nil when input is accessible in container")
	}
}

func TestEnsureHasRecipeInputs_MissingInput_ReturnsPickupIntent(t *testing.T) {
	gameMap := game.NewMap(10, 10)

	// Character has a stick but no shell
	stick := entity.NewStick(0, 0)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{stick},
	}
	char.X = 0
	char.Y = 0

	// Shell available on map
	shell := entity.NewShell(5, 5, types.ColorSilver)
	items := []*entity.Item{shell}

	recipe := entity.RecipeRegistry["shell-hoe"]

	intent := EnsureHasRecipeInputs(char, recipe, items, gameMap, nil)
	if intent == nil {
		t.Fatal("Should return intent to pick up missing shell")
	}
	if intent.TargetItem != shell {
		t.Error("Intent should target the shell")
	}
	if intent.Action != entity.ActionPickup {
		t.Error("Intent action should be ActionPickup")
	}
}

func TestEnsureHasRecipeInputs_DropsNonRecipeLooseItems(t *testing.T) {
	gameMap := game.NewMap(10, 10)

	// Character has full inventory with a nut (not a recipe input) and a stick (recipe input)
	nut := entity.NewNut(0, 0)
	stick := entity.NewStick(0, 0)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{nut, stick},
	}
	char.X = 0
	char.Y = 0

	// Shell available on map
	shell := entity.NewShell(5, 5, types.ColorSilver)
	items := []*entity.Item{shell}

	recipe := entity.RecipeRegistry["shell-hoe"]

	intent := EnsureHasRecipeInputs(char, recipe, items, gameMap, nil)

	// Should have dropped the nut to make room
	if len(char.Inventory) != 1 {
		t.Errorf("Should have dropped non-recipe item, inventory size = %d", len(char.Inventory))
	}
	if char.Inventory[0] != stick {
		t.Error("Should have kept the stick (recipe input)")
	}

	// Should return intent to pick up shell
	if intent == nil {
		t.Fatal("Should return pickup intent for shell")
	}
	if intent.TargetItem != shell {
		t.Error("Intent should target the shell")
	}
}

func TestEnsureHasRecipeInputs_DoesNotDropContainerWithRecipeInput(t *testing.T) {
	gameMap := game.NewMap(10, 10)

	// Character has a vessel containing shells and a loose nut
	vessel := createTestVessel()
	vessel.Container.Contents = []entity.Stack{
		{Variety: &entity.ItemVariety{ItemType: "shell", Color: types.ColorSilver}, Count: 1},
	}
	nut := entity.NewNut(0, 0)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel, nut},
	}
	char.X = 0
	char.Y = 0

	// Stick available on map
	stick := entity.NewStick(5, 5)
	items := []*entity.Item{stick}

	recipe := entity.RecipeRegistry["shell-hoe"]

	intent := EnsureHasRecipeInputs(char, recipe, items, gameMap, nil)

	// Should have dropped the nut, NOT the vessel (vessel has shell = recipe input)
	if len(char.Inventory) != 1 {
		t.Errorf("Should have 1 item in inventory, got %d", len(char.Inventory))
	}
	if char.Inventory[0] != vessel {
		t.Error("Should have kept the vessel (contains recipe input)")
	}

	// Should return intent to pick up stick
	if intent == nil {
		t.Fatal("Should return pickup intent for stick")
	}
	if intent.TargetItem != stick {
		t.Error("Intent should target the stick")
	}
}

func TestEnsureHasRecipeInputs_NilWhenInputsNotOnMap(t *testing.T) {
	gameMap := game.NewMap(10, 10)

	// Character has nothing, no items on map
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}

	recipe := entity.RecipeRegistry["shell-hoe"]

	intent := EnsureHasRecipeInputs(char, recipe, nil, gameMap, nil)
	if intent != nil {
		t.Error("Should return nil when recipe inputs don't exist on map")
	}
}

// =============================================================================
// EnsureHasItem Tests
// =============================================================================

func TestEnsureHasItem_ReturnsNilWhenAlreadyCarrying(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	hoe := entity.NewHoe(0, 0, types.ColorSilver)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{hoe},
	}

	intent := EnsureHasItem(char, "hoe", nil, gameMap, nil)
	if intent != nil {
		t.Error("Should return nil when character already carries a hoe")
	}
}

func TestEnsureHasItem_ReturnsPickupIntentWhenHoeOnMap(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 0
	char.Y = 0

	hoe := entity.NewHoe(5, 5, types.ColorSilver)
	items := []*entity.Item{hoe}

	intent := EnsureHasItem(char, "hoe", items, gameMap, nil)
	if intent == nil {
		t.Fatal("Should return pickup intent when hoe exists on map")
	}
	if intent.TargetItem != hoe {
		t.Error("Intent should target the hoe")
	}
	if intent.Action != entity.ActionPickup {
		t.Error("Intent action should be ActionPickup")
	}
}

func TestEnsureHasItem_DropsNonTargetLooseItemsToMakeSpace(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	nut := entity.NewNut(0, 0)
	berry := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{nut, berry},
	}
	char.X = 0
	char.Y = 0

	hoe := entity.NewHoe(5, 5, types.ColorSilver)
	items := []*entity.Item{hoe}

	intent := EnsureHasItem(char, "hoe", items, gameMap, nil)

	// Should have dropped one item to make room
	if len(char.Inventory) != 1 {
		t.Errorf("Should have dropped one item, inventory size = %d", len(char.Inventory))
	}

	// Should return pickup intent for hoe
	if intent == nil {
		t.Fatal("Should return pickup intent for hoe")
	}
	if intent.TargetItem != hoe {
		t.Error("Intent should target the hoe")
	}
}

func TestEnsureHasItem_ReturnsNilWhenNoItemExists(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}

	// No hoes on map
	intent := EnsureHasItem(char, "hoe", nil, gameMap, nil)
	if intent != nil {
		t.Error("Should return nil when no hoe exists anywhere")
	}
}

// =============================================================================
// Pickup Plantable Tests
// =============================================================================

func TestPickup_SetsPlantableForBerry(t *testing.T) {
	t.Parallel()

	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	berry := entity.NewBerry(3, 3, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 3
	char.Y = 3

	result := Pickup(char, berry, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Expected PickupToInventory, got %d", result)
	}
	if !berry.Plantable {
		t.Error("Picked up berry should have Plantable=true")
	}
}

func TestPickup_SetsPlantableForMushroom(t *testing.T) {
	t.Parallel()

	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	mushroom := entity.NewMushroom(3, 3, types.ColorBrown, types.PatternSpotted, types.TextureSlimy, false, false)
	gameMap.AddItem(mushroom)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 3
	char.Y = 3

	result := Pickup(char, mushroom, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Expected PickupToInventory, got %d", result)
	}
	if !mushroom.Plantable {
		t.Error("Picked up mushroom should have Plantable=true")
	}
}

func TestPickup_DoesNotSetPlantableForGourd(t *testing.T) {
	t.Parallel()

	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	gourd := entity.NewGourd(3, 3, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	gameMap.AddItem(gourd)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 3
	char.Y = 3

	result := Pickup(char, gourd, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Expected PickupToInventory, got %d", result)
	}
	if gourd.Plantable {
		t.Error("Picked up gourd should NOT have Plantable=true")
	}
}

func TestPickup_DoesNotSetPlantableForFlower(t *testing.T) {
	t.Parallel()

	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	flower := entity.NewFlower(3, 3, types.ColorBlue)
	gameMap.AddItem(flower)

	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{},
	}
	char.X = 3
	char.Y = 3

	result := Pickup(char, flower, gameMap, nil, registry)
	if result != PickupToInventory {
		t.Fatalf("Expected PickupToInventory, got %d", result)
	}
	if flower.Plantable {
		t.Error("Picked up flower should NOT have Plantable=true")
	}
}

func TestPickup_VesselPathSetsPlantableForBerry(t *testing.T) {
	t.Parallel()

	registry := createTestRegistry()
	gameMap := game.NewMap(10, 10)
	gameMap.SetVarieties(registry)

	berry := entity.NewBerry(3, 3, types.ColorRed, false, false)
	gameMap.AddItem(berry)

	vessel := createTestVessel()
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{vessel},
	}
	char.X = 3
	char.Y = 3

	result := Pickup(char, berry, gameMap, nil, registry)
	if result != PickupToVessel {
		t.Fatalf("Expected PickupToVessel, got %d", result)
	}
	if !berry.Plantable {
		t.Error("Berry picked up to vessel should have Plantable=true")
	}
}

func TestEnsureHasRecipeInputs_SingleInputRecipe(t *testing.T) {
	gameMap := game.NewMap(10, 10)

	// Character has a gourd already
	gourd := entity.NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	char := &entity.Character{
		ID:        1,
		Name:      "Test",
		Inventory: []*entity.Item{gourd},
	}

	recipe := entity.RecipeRegistry["hollow-gourd"]

	intent := EnsureHasRecipeInputs(char, recipe, nil, gameMap, nil)
	if intent != nil {
		t.Error("Should return nil when single input is in inventory")
	}
}

