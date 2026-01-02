package entity

import (
	"testing"

	"petri/internal/config"
	"petri/internal/types"
)

// TestNewBerry_Properties verifies NewBerry creates berry with correct properties
func TestNewBerry_Properties(t *testing.T) {
	t.Parallel()

	item := NewBerry(5, 10, types.ColorRed, false, false)

	x, y := item.Position()
	if x != 5 || y != 10 {
		t.Errorf("NewBerry Position(): got (%d, %d), want (5, 10)", x, y)
	}
	if item.ItemType != "berry" {
		t.Errorf("NewBerry ItemType: got %q, want %q", item.ItemType, "berry")
	}
	if item.Color != types.ColorRed {
		t.Errorf("NewBerry Color: got %q, want %q", item.Color, types.ColorRed)
	}
	if item.Poisonous != false {
		t.Error("NewBerry Poisonous: got true, want false")
	}
	if item.Symbol() != config.CharBerry {
		t.Errorf("NewBerry Symbol(): got %q, want %q", item.Symbol(), config.CharBerry)
	}
	if item.Type() != TypeItem {
		t.Errorf("NewBerry Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewBerry_Poisonous verifies NewBerry can create poisonous berry
func TestNewBerry_Poisonous(t *testing.T) {
	t.Parallel()

	item := NewBerry(0, 0, types.ColorWhite, true, false)
	if !item.Poisonous {
		t.Error("NewBerry with poisonous=true: got Poisonous=false")
	}
}

// TestNewMushroom_Properties verifies NewMushroom creates mushroom with correct properties
func TestNewMushroom_Properties(t *testing.T) {
	t.Parallel()

	item := NewMushroom(8, 12, types.ColorBrown, false, false)

	x, y := item.Position()
	if x != 8 || y != 12 {
		t.Errorf("NewMushroom Position(): got (%d, %d), want (8, 12)", x, y)
	}
	if item.ItemType != "mushroom" {
		t.Errorf("NewMushroom ItemType: got %q, want %q", item.ItemType, "mushroom")
	}
	if item.Color != types.ColorBrown {
		t.Errorf("NewMushroom Color: got %q, want %q", item.Color, types.ColorBrown)
	}
	if item.Poisonous != false {
		t.Error("NewMushroom Poisonous: got true, want false")
	}
	if item.Symbol() != config.CharMushroom {
		t.Errorf("NewMushroom Symbol(): got %q, want %q", item.Symbol(), config.CharMushroom)
	}
	if item.Type() != TypeItem {
		t.Errorf("NewMushroom Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewMushroom_Poisonous verifies NewMushroom can create poisonous mushroom
func TestNewMushroom_Poisonous(t *testing.T) {
	t.Parallel()

	item := NewMushroom(0, 0, types.ColorBlue, true, false)
	if !item.Poisonous {
		t.Error("NewMushroom with poisonous=true: got Poisonous=false")
	}
}

// TestItem_Description_Berry verifies Description combines color and type for berry
func TestItem_Description_Berry(t *testing.T) {
	t.Parallel()

	item := NewBerry(0, 0, types.ColorRed, false, false)
	got := item.Description()
	if got != "red berry" {
		t.Errorf("Berry Description(): got %q, want %q", got, "red berry")
	}
}

// TestItem_Description_Mushroom verifies Description combines color and type for mushroom
func TestItem_Description_Mushroom(t *testing.T) {
	t.Parallel()

	item := NewMushroom(0, 0, types.ColorBrown, false, false)
	got := item.Description()
	if got != "brown mushroom" {
		t.Errorf("Mushroom Description(): got %q, want %q", got, "brown mushroom")
	}
}
