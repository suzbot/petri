package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

func TestTryDiscoverKnowHow_DiscoverHarvestOnPickup(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "berry",
		Edible:   true,
	}

	// With 100% chance, should always discover
	discovered := TryDiscoverKnowHow(char, entity.ActionPickup, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("harvest") {
		t.Error("Expected character to know harvest after discovery")
	}
}

func TestTryDiscoverKnowHow_DiscoverHarvestOnConsume(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "mushroom",
		Edible:   true,
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionConsume, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("harvest") {
		t.Error("Expected character to know harvest after discovery")
	}
}

func TestTryDiscoverKnowHow_DiscoverHarvestOnLook(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "gourd",
		Edible:   true,
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("harvest") {
		t.Error("Expected character to know harvest after discovery")
	}
}

func TestTryDiscoverKnowHow_NoDiscoverOnNonEdible(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "flower",
		Edible:   false, // flowers are not edible
	}

	// Even with 100% chance, should not discover because item is not edible
	discovered := TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if discovered {
		t.Error("Should not discover from non-edible item")
	}
	if char.KnowsActivity("harvest") {
		t.Error("Character should not know harvest")
	}
}

func TestTryDiscoverKnowHow_NoDiscoverWhenAlreadyKnown(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest"}, // already knows
	}
	item := &entity.Item{
		ItemType: "berry",
		Edible:   true,
	}

	// Should return false because already known
	discovered := TryDiscoverKnowHow(char, entity.ActionPickup, item, nil, 1.0)

	if discovered {
		t.Error("Should not discover when already known")
	}
}

func TestTryDiscoverKnowHow_NoDiscoverWithZeroChance(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "berry",
		Edible:   true,
	}

	// With 0% chance, should never discover
	discovered := TryDiscoverKnowHow(char, entity.ActionPickup, item, nil, 0.0)

	if discovered {
		t.Error("Should not discover with 0% chance")
	}
}

func TestTryDiscoverKnowHow_NoDiscoverOnDrink(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "berry",
		Edible:   true,
	}

	// Drinking is not a trigger for harvest discovery
	discovered := TryDiscoverKnowHow(char, entity.ActionDrink, item, nil, 1.0)

	if discovered {
		t.Error("Should not discover harvest from drinking")
	}
}

func TestTryDiscoverKnowHow_LogsDiscovery(t *testing.T) {
	char := &entity.Character{
		ID:              1,
		Name:            "Alice",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType: "berry",
		Edible:   true,
		Color:    types.ColorRed,
	}
	log := NewActionLog()

	TryDiscoverKnowHow(char, entity.ActionPickup, item, log, 1.0)

	if len(log.Entries) == 0 {
		t.Error("Expected log entry for discovery")
	}
	entry := log.Entries[0]
	if entry.CharID != 1 {
		t.Errorf("Expected CharID 1, got %d", entry.CharID)
	}
	if entry.Category != "discovery" {
		t.Errorf("Expected category 'discovery', got '%s'", entry.Category)
	}
}
