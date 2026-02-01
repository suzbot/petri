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
