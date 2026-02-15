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
		Edible: &entity.EdibleProperties{},
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
		Edible: &entity.EdibleProperties{},
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
		Edible: &entity.EdibleProperties{},
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
		// Edible is nil - flowers are not edible
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
		Edible: &entity.EdibleProperties{},
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
		Edible: &entity.EdibleProperties{},
	}

	// With 0% chance, should never discover
	discovered := TryDiscoverKnowHow(char, entity.ActionPickup, item, nil, 0.0)

	if discovered {
		t.Error("Should not discover with 0% chance")
	}
}

func TestTryDiscoverKnowHow_NoHarvestDiscoverOnDrink(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{"hollow-gourd"}, // Already knows recipe so that won't trigger
	}
	item := &entity.Item{
		ItemType: "berry",
		Edible: &entity.EdibleProperties{},
	}

	// Drinking is not a trigger for harvest discovery
	discovered := TryDiscoverKnowHow(char, entity.ActionDrink, item, nil, 1.0)

	if discovered {
		t.Error("Should not discover harvest from drinking")
	}
	if char.KnowsActivity("harvest") {
		t.Error("Should not know harvest after drinking")
	}
}

func TestGetDiscoveryChance_ByMoodTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mood     float64
		wantZero bool
		wantFull bool // true = full config rate (Joyful), false = 20% of config rate (Happy)
	}{
		// TierNone (Joyful): 90-100 - full rate
		{"mood 100 (Joyful) gets full rate", 100, false, true},
		{"mood 90 (Joyful) gets full rate", 90, false, true},
		// TierMild (Happy): 65-89 - 20% of rate
		{"mood 89 (Happy) gets 20% rate", 89, false, false},
		{"mood 65 (Happy) gets 20% rate", 65, false, false},
		// TierModerate (Neutral): 35-64 - no discovery
		{"mood 64 (Neutral) gets zero", 64, true, false},
		{"mood 50 (Neutral) gets zero", 50, true, false},
		// TierSevere (Unhappy): 11-34 - no discovery
		{"mood 34 (Unhappy) gets zero", 34, true, false},
		// TierCrisis (Miserable): 0-10 - no discovery
		{"mood 10 (Miserable) gets zero", 10, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := &entity.Character{Mood: tt.mood}
			chance := GetDiscoveryChance(char)

			if tt.wantZero && chance != 0 {
				t.Errorf("Expected 0 chance for mood %.0f, got %f", tt.mood, chance)
			}
			if tt.wantFull && chance == 0 {
				t.Errorf("Expected non-zero chance for mood %.0f (Joyful)", tt.mood)
			}
			if !tt.wantZero && !tt.wantFull {
				// Happy tier - should be 20% of Joyful
				joyfulChar := &entity.Character{Mood: 100}
				joyfulChance := GetDiscoveryChance(joyfulChar)
				expected := joyfulChance * 0.20
				if chance != expected {
					t.Errorf("Expected Happy chance %f (20%% of %f), got %f", expected, joyfulChance, chance)
				}
			}
		})
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
		Edible: &entity.EdibleProperties{},
		Color:    types.ColorRed,
	}
	log := NewActionLog(100)

	TryDiscoverKnowHow(char, entity.ActionPickup, item, log, 1.0)

	entries := log.Events(1, 0)
	if len(entries) == 0 {
		t.Error("Expected log entry for discovery")
	}
	entry := entries[0]
	if entry.CharID != 1 {
		t.Errorf("Expected CharID 1, got %d", entry.CharID)
	}
	if entry.Type != "discovery" {
		t.Errorf("Expected type 'discovery', got '%s'", entry.Type)
	}
}

// Plant discovery tests (RequiresPlantable trigger)

func TestTryDiscoverKnowHow_DiscoverPlantOnLookAtPlantable(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType:  "seed",
		Plantable: true,
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance on plantable item")
	}
	if !char.KnowsActivity("plant") {
		t.Error("Expected character to know plant after discovery")
	}
}

func TestTryDiscoverKnowHow_DiscoverPlantOnPickupPlantable(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}
	item := &entity.Item{
		ItemType:  "berry",
		Plantable: true,
		Edible:    &entity.EdibleProperties{},
	}

	// Character already knows harvest so plant can trigger
	char.KnownActivities = []string{"harvest"}

	discovered := TryDiscoverKnowHow(char, entity.ActionPickup, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance on plantable item")
	}
	if !char.KnowsActivity("plant") {
		t.Error("Expected character to know plant after discovery")
	}
}

func TestTryDiscoverKnowHow_NoPlantDiscoverOnNonPlantable(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest"}, // already knows harvest
	}
	item := &entity.Item{
		ItemType:  "berry",
		Plantable: false, // not plantable (still growing on map)
		Edible:    &entity.EdibleProperties{},
	}

	TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if char.KnowsActivity("plant") {
		t.Error("Should not discover plant from non-plantable item")
	}
}

// Recipe discovery tests

func TestTryDiscoverKnowHow_DiscoverRecipeOnGourdLook(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest"}, // Already knows harvest so recipe can trigger
		KnownRecipes:    []string{},
	}
	item := &entity.Item{
		ItemType: "gourd",
		Edible: &entity.EdibleProperties{},
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	// Should discover both activity and recipe
	if !char.KnowsActivity("craftVessel") {
		t.Error("Expected character to know craftVessel after recipe discovery")
	}
	if !char.KnowsRecipe("hollow-gourd") {
		t.Error("Expected character to know hollow-gourd recipe after discovery")
	}
}

func TestTryDiscoverKnowHow_DiscoverRecipeOnGourdPickup(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest"}, // Already knows harvest so recipe can trigger
		KnownRecipes:    []string{},
	}
	item := &entity.Item{
		ItemType: "gourd",
		Edible: &entity.EdibleProperties{},
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionPickup, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("craftVessel") {
		t.Error("Expected character to know craftVessel")
	}
	if !char.KnowsRecipe("hollow-gourd") {
		t.Error("Expected character to know hollow-gourd recipe")
	}
}

func TestTryDiscoverKnowHow_DiscoverRecipeOnGourdConsume(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest"}, // Already knows harvest so recipe can trigger
		KnownRecipes:    []string{},
	}
	item := &entity.Item{
		ItemType: "gourd",
		Edible: &entity.EdibleProperties{},
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionConsume, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("craftVessel") {
		t.Error("Expected character to know craftVessel")
	}
	if !char.KnowsRecipe("hollow-gourd") {
		t.Error("Expected character to know hollow-gourd recipe")
	}
}

func TestTryDiscoverKnowHow_DiscoverRecipeOnDrink(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}

	// Drinking from spring - no item needed
	discovered := TryDiscoverKnowHow(char, entity.ActionDrink, nil, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("craftVessel") {
		t.Error("Expected character to know craftVessel")
	}
	if !char.KnowsRecipe("hollow-gourd") {
		t.Error("Expected character to know hollow-gourd recipe")
	}
}

func TestTryDiscoverKnowHow_NoRecipeDiscoverOnNonGourd(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}
	item := &entity.Item{
		ItemType: "berry", // not a gourd
		Edible: &entity.EdibleProperties{},
	}

	// Looking at a berry should discover harvest, not craftVessel
	TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if char.KnowsActivity("craftVessel") {
		t.Error("Should not discover craftVessel from berry")
	}
	if char.KnowsRecipe("hollow-gourd") {
		t.Error("Should not discover hollow-gourd recipe from berry")
	}
	// But should discover harvest
	if !char.KnowsActivity("harvest") {
		t.Error("Should discover harvest from berry")
	}
}

func TestTryDiscoverKnowHow_NoRecipeDiscoverWhenAlreadyKnown(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest", "craftVessel"}, // knows both so nothing new to discover
		KnownRecipes:    []string{"hollow-gourd"},
	}
	item := &entity.Item{
		ItemType: "gourd",
		Edible: &entity.EdibleProperties{},
	}

	// Should return false because everything already known
	discovered := TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if discovered {
		t.Error("Should not discover when everything already known")
	}
}

func TestTryDiscoverKnowHow_ActivityAlreadyKnownButNotRecipe(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"harvest", "craftVessel"}, // already knows activities
		KnownRecipes:    []string{},                         // but not recipe
	}
	item := &entity.Item{
		ItemType: "gourd",
		Edible: &entity.EdibleProperties{},
	}

	discovered := TryDiscoverKnowHow(char, entity.ActionLook, item, nil, 1.0)

	if !discovered {
		t.Error("Expected discovery of new recipe")
	}
	if !char.KnowsRecipe("hollow-gourd") {
		t.Error("Expected character to know hollow-gourd recipe")
	}
}
