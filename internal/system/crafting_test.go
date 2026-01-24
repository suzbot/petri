package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

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
		Edible:   true, // gourd is edible
	}
	recipe := entity.RecipeRegistry["hollow-gourd"]

	vessel := CreateVessel(gourd, recipe)

	if vessel.Edible {
		t.Error("Vessel should not be edible")
	}
}

func TestCreateVessel_UsesRecipeName(t *testing.T) {
	t.Parallel()

	gourd := &entity.Item{
		ItemType: "gourd",
		Color:    types.ColorGreen,
	}
	recipe := entity.RecipeRegistry["hollow-gourd"]

	vessel := CreateVessel(gourd, recipe)

	if vessel.Name != "Hollow Gourd" {
		t.Errorf("Expected Name 'Hollow Gourd', got %s", vessel.Name)
	}
	if vessel.Description() != "Hollow Gourd" {
		t.Errorf("Expected Description 'Hollow Gourd', got %s", vessel.Description())
	}
}
