package entity

import (
	"testing"

	"petri/internal/config"
)

// TestNewSpring_Properties verifies NewSpring creates spring with correct properties
func TestNewSpring_Properties(t *testing.T) {
	t.Parallel()

	f := NewSpring(10, 15)

	pos := f.Pos()
	if pos.X != 10 || pos.Y != 15 {
		t.Errorf("NewSpring Pos(): got (%d, %d), want (10, 15)", pos.X, pos.Y)
	}
	if f.FType != FeatureSpring {
		t.Errorf("NewSpring FType: got %d, want %d", f.FType, FeatureSpring)
	}
	if !f.DrinkSource {
		t.Error("NewSpring DrinkSource: got false, want true")
	}
	if f.Bed {
		t.Error("NewSpring Bed: got true, want false")
	}
	if f.Symbol() != config.CharSpring {
		t.Errorf("NewSpring Symbol(): got %q, want %q", f.Symbol(), config.CharSpring)
	}
	if f.Type() != TypeFeature {
		t.Errorf("NewSpring Type(): got %d, want %d", f.Type(), TypeFeature)
	}
}

// TestNewLeafPile_Properties verifies NewLeafPile creates bed with correct properties
func TestNewLeafPile_Properties(t *testing.T) {
	t.Parallel()

	f := NewLeafPile(20, 25)

	pos := f.Pos()
	if pos.X != 20 || pos.Y != 25 {
		t.Errorf("NewLeafPile Pos(): got (%d, %d), want (20, 25)", pos.X, pos.Y)
	}
	if f.FType != FeatureLeafPile {
		t.Errorf("NewLeafPile FType: got %d, want %d", f.FType, FeatureLeafPile)
	}
	if f.DrinkSource {
		t.Error("NewLeafPile DrinkSource: got true, want false")
	}
	if !f.Bed {
		t.Error("NewLeafPile Bed: got false, want true")
	}
	if f.Symbol() != config.CharLeafPile {
		t.Errorf("NewLeafPile Symbol(): got %q, want %q", f.Symbol(), config.CharLeafPile)
	}
	if f.Type() != TypeFeature {
		t.Errorf("NewLeafPile Type(): got %d, want %d", f.Type(), TypeFeature)
	}
}

// TestFeature_IsDrinkSource_Spring verifies IsDrinkSource returns true for spring
func TestFeature_IsDrinkSource_Spring(t *testing.T) {
	t.Parallel()

	f := NewSpring(0, 0)
	if !f.IsDrinkSource() {
		t.Error("Spring IsDrinkSource(): got false, want true")
	}
}

// TestFeature_IsDrinkSource_LeafPile verifies IsDrinkSource returns false for leaf pile
func TestFeature_IsDrinkSource_LeafPile(t *testing.T) {
	t.Parallel()

	f := NewLeafPile(0, 0)
	if f.IsDrinkSource() {
		t.Error("LeafPile IsDrinkSource(): got true, want false")
	}
}

// TestFeature_IsBed_LeafPile verifies IsBed returns true for leaf pile
func TestFeature_IsBed_LeafPile(t *testing.T) {
	t.Parallel()

	f := NewLeafPile(0, 0)
	if !f.IsBed() {
		t.Error("LeafPile IsBed(): got false, want true")
	}
}

// TestFeature_IsBed_Spring verifies IsBed returns false for spring
func TestFeature_IsBed_Spring(t *testing.T) {
	t.Parallel()

	f := NewSpring(0, 0)
	if f.IsBed() {
		t.Error("Spring IsBed(): got true, want false")
	}
}

// TestFeature_Description_Spring verifies Description returns "spring"
func TestFeature_Description_Spring(t *testing.T) {
	t.Parallel()

	f := NewSpring(0, 0)
	got := f.Description()
	if got != "spring" {
		t.Errorf("Spring Description(): got %q, want %q", got, "spring")
	}
}

// TestFeature_Description_LeafPile verifies Description returns "leaf pile"
func TestFeature_Description_LeafPile(t *testing.T) {
	t.Parallel()

	f := NewLeafPile(0, 0)
	got := f.Description()
	if got != "leaf pile" {
		t.Errorf("LeafPile Description(): got %q, want %q", got, "leaf pile")
	}
}

// TestNewSpring_IsImpassable verifies springs are impassable
func TestNewSpring_IsImpassable(t *testing.T) {
	t.Parallel()

	f := NewSpring(0, 0)
	if f.IsPassable() {
		t.Error("Spring IsPassable(): got true, want false")
	}
	if f.Passable {
		t.Error("Spring Passable field: got true, want false")
	}
}

// TestNewLeafPile_IsPassable verifies leaf piles are passable
func TestNewLeafPile_IsPassable(t *testing.T) {
	t.Parallel()

	f := NewLeafPile(0, 0)
	if !f.IsPassable() {
		t.Error("LeafPile IsPassable(): got false, want true")
	}
	if !f.Passable {
		t.Error("LeafPile Passable field: got false, want true")
	}
}
