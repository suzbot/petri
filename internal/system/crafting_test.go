package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

// =============================================================================
// Step 4b: CreateBrick
// =============================================================================

func TestCreateBrick_Properties(t *testing.T) {
	t.Parallel()

	clay := entity.NewClay(3, 7)
	recipe := entity.RecipeRegistry["clay-brick"]
	if recipe == nil {
		t.Fatal("clay-brick recipe not found")
	}

	brick := CreateBrick(clay, recipe)

	if brick.ItemType != "brick" {
		t.Errorf("Expected ItemType 'brick', got %q", brick.ItemType)
	}
	if brick.Color != types.ColorTerracotta {
		t.Errorf("Expected Color ColorTerracotta, got %v", brick.Color)
	}
	if brick.Sym != config.CharBrick {
		t.Errorf("Expected Sym %c (CharBrick), got %c", config.CharBrick, brick.Sym)
	}
	if brick.X != clay.X || brick.Y != clay.Y {
		t.Errorf("Expected position (%d,%d), got (%d,%d)", clay.X, clay.Y, brick.X, brick.Y)
	}
}

func TestCreateVessel_InheritsGourdAttributes(t *testing.T) {
	t.Parallel()

	gourd := &entity.Item{
		ItemType: "gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternStriped,
		Texture:  types.TextureWarty,
	}
	recipe := entity.RecipeRegistry["hollow-gourd"]

	vessel := CreateVessel(gourd, recipe)

	if vessel.ItemType != "vessel" {
		t.Errorf("Expected ItemType 'vessel', got %s", vessel.ItemType)
	}
	if vessel.Color != types.ColorGreen {
		t.Errorf("Expected Color green, got %s", vessel.Color)
	}
	if vessel.Pattern != types.PatternStriped {
		t.Errorf("Expected Pattern striped, got %s", vessel.Pattern)
	}
	if vessel.Texture != types.TextureWarty {
		t.Errorf("Expected Texture warty, got %s", vessel.Texture)
	}
	if vessel.Sym != config.CharVessel {
		t.Errorf("Expected Sym %c, got %c", config.CharVessel, vessel.Sym)
	}
}

func TestCreateVessel_HasContainer(t *testing.T) {
	t.Parallel()

	gourd := &entity.Item{
		ItemType: "gourd",
		Color:    types.ColorGreen,
	}
	recipe := entity.RecipeRegistry["hollow-gourd"]

	vessel := CreateVessel(gourd, recipe)

	if vessel.Container == nil {
		t.Fatal("Expected vessel to have Container")
	}
	if vessel.Container.Capacity != 1 {
		t.Errorf("Expected Capacity 1, got %d", vessel.Container.Capacity)
	}
	if len(vessel.Container.Contents) != 0 {
		t.Errorf("Expected empty Contents, got %d items", len(vessel.Container.Contents))
	}
}

func TestCreateVessel_NotEdible(t *testing.T) {
	t.Parallel()

	gourd := &entity.Item{
		ItemType: "gourd",
		Edible:   &entity.EdibleProperties{}, // gourd is edible
	}
	recipe := entity.RecipeRegistry["hollow-gourd"]

	vessel := CreateVessel(gourd, recipe)

	if vessel.IsEdible() {
		t.Error("Vessel should not be edible")
	}
}

func TestCreateVessel_UsesRecipeKind(t *testing.T) {
	t.Parallel()

	gourd := &entity.Item{
		ItemType: "gourd",
		Color:    types.ColorGreen,
	}
	recipe := entity.RecipeRegistry["hollow-gourd"]

	vessel := CreateVessel(gourd, recipe)

	if vessel.Kind != "hollow gourd" {
		t.Errorf("Expected Kind 'hollow gourd', got %s", vessel.Kind)
	}
	if vessel.Description() != "green hollow gourd" {
		t.Errorf("Expected Description 'green hollow gourd', got %s", vessel.Description())
	}
}

func TestCreateVessel_Material(t *testing.T) {
	t.Parallel()

	gourd := &entity.Item{ItemType: "gourd", Color: types.ColorGreen}
	recipe := entity.RecipeRegistry["hollow-gourd"]
	vessel := CreateVessel(gourd, recipe)

	if vessel.Material != "gourd" {
		t.Errorf("Expected Material 'gourd', got %q", vessel.Material)
	}
}

func TestCreateHoe_Material(t *testing.T) {
	t.Parallel()

	hoe := CreateHoe(&entity.Item{ItemType: "shell", Color: types.ColorSilver}, entity.RecipeRegistry["shell-hoe"])

	if hoe.Material != "shell" {
		t.Errorf("Expected Material 'shell', got %q", hoe.Material)
	}
}

func TestCreateBrick_Material(t *testing.T) {
	t.Parallel()

	brick := CreateBrick(entity.NewClay(0, 0), entity.RecipeRegistry["clay-brick"])

	if brick.Material != "clay" {
		t.Errorf("Expected Material 'clay', got %q", brick.Material)
	}
}
