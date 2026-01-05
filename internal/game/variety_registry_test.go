package game

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

func TestVarietyRegistry_RegisterAndGet(t *testing.T) {
	reg := NewVarietyRegistry()

	v := &entity.ItemVariety{
		ID:       "berry-red",
		ItemType: "berry",
		Color:    types.ColorRed,
		Edible:   true,
	}

	reg.Register(v)

	got := reg.Get("berry-red")
	if got == nil {
		t.Fatal("Get() returned nil for registered variety")
	}
	if got.ID != "berry-red" {
		t.Errorf("Get() returned variety with ID %q, want %q", got.ID, "berry-red")
	}

	notFound := reg.Get("nonexistent")
	if notFound != nil {
		t.Error("Get() should return nil for unregistered variety")
	}
}

func TestVarietyRegistry_VarietiesOfType(t *testing.T) {
	reg := NewVarietyRegistry()

	reg.Register(&entity.ItemVariety{ID: "berry-red", ItemType: "berry", Color: types.ColorRed})
	reg.Register(&entity.ItemVariety{ID: "berry-blue", ItemType: "berry", Color: types.ColorBlue})
	reg.Register(&entity.ItemVariety{ID: "mushroom-brown", ItemType: "mushroom", Color: types.ColorBrown})

	berries := reg.VarietiesOfType("berry")
	if len(berries) != 2 {
		t.Errorf("VarietiesOfType(berry) returned %d varieties, want 2", len(berries))
	}

	mushrooms := reg.VarietiesOfType("mushroom")
	if len(mushrooms) != 1 {
		t.Errorf("VarietiesOfType(mushroom) returned %d varieties, want 1", len(mushrooms))
	}

	flowers := reg.VarietiesOfType("flower")
	if len(flowers) != 0 {
		t.Errorf("VarietiesOfType(flower) returned %d varieties, want 0", len(flowers))
	}
}

func TestVarietyRegistry_EdibleVarieties(t *testing.T) {
	reg := NewVarietyRegistry()

	reg.Register(&entity.ItemVariety{ID: "berry-red", ItemType: "berry", Edible: true})
	reg.Register(&entity.ItemVariety{ID: "mushroom-brown", ItemType: "mushroom", Edible: true})
	reg.Register(&entity.ItemVariety{ID: "flower-purple", ItemType: "flower", Edible: false})

	edible := reg.EdibleVarieties()
	if len(edible) != 2 {
		t.Errorf("EdibleVarieties() returned %d varieties, want 2", len(edible))
	}

	for _, v := range edible {
		if !v.Edible {
			t.Errorf("EdibleVarieties() returned non-edible variety %q", v.ID)
		}
	}
}

func TestVarietyRegistry_AllVarieties(t *testing.T) {
	reg := NewVarietyRegistry()

	reg.Register(&entity.ItemVariety{ID: "a"})
	reg.Register(&entity.ItemVariety{ID: "b"})
	reg.Register(&entity.ItemVariety{ID: "c"})

	all := reg.AllVarieties()
	if len(all) != 3 {
		t.Errorf("AllVarieties() returned %d varieties, want 3", len(all))
	}
}

func TestVarietyRegistry_Count(t *testing.T) {
	reg := NewVarietyRegistry()

	if reg.Count() != 0 {
		t.Errorf("Count() on empty registry = %d, want 0", reg.Count())
	}

	reg.Register(&entity.ItemVariety{ID: "a"})
	reg.Register(&entity.ItemVariety{ID: "b"})

	if reg.Count() != 2 {
		t.Errorf("Count() after registering 2 = %d, want 2", reg.Count())
	}
}
