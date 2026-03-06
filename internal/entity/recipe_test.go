package entity

import (
	"petri/internal/config"
	"testing"
)

// TestShellHoeRecipe_Registered verifies shell-hoe recipe exists with correct properties
func TestShellHoeRecipe_Registered(t *testing.T) {
	t.Parallel()

	recipe, ok := RecipeRegistry["shell-hoe"]
	if !ok {
		t.Fatal("shell-hoe recipe not found in RecipeRegistry")
	}

	if recipe.ActivityID != "craftHoe" {
		t.Errorf("shell-hoe ActivityID: got %q, want %q", recipe.ActivityID, "craftHoe")
	}

	// Verify inputs: 1 stick + 1 shell
	if len(recipe.Inputs) != 2 {
		t.Fatalf("shell-hoe Inputs count: got %d, want 2", len(recipe.Inputs))
	}
	inputTypes := map[string]int{}
	for _, input := range recipe.Inputs {
		inputTypes[input.ItemType] = input.Count
	}
	if inputTypes["stick"] != 1 {
		t.Errorf("shell-hoe stick input count: got %d, want 1", inputTypes["stick"])
	}
	if inputTypes["shell"] != 1 {
		t.Errorf("shell-hoe shell input count: got %d, want 1", inputTypes["shell"])
	}

	// Verify output
	if recipe.Output.ItemType != "hoe" {
		t.Errorf("shell-hoe Output.ItemType: got %q, want %q", recipe.Output.ItemType, "hoe")
	}
	if recipe.Output.Kind != "shell hoe" {
		t.Errorf("shell-hoe Output.Kind: got %q, want %q", recipe.Output.Kind, "shell hoe")
	}

	// Verify duration
	if recipe.Duration != config.ActionDurationLong {
		t.Errorf("shell-hoe Duration: got %v, want %v", recipe.Duration, config.ActionDurationLong)
	}

	// Verify discovery triggers exist for stick and shell
	if len(recipe.DiscoveryTriggers) == 0 {
		t.Fatal("shell-hoe DiscoveryTriggers: got empty, want triggers for stick/shell")
	}
	hasStickTrigger := false
	hasShellTrigger := false
	for _, trigger := range recipe.DiscoveryTriggers {
		if trigger.ItemType == "stick" {
			hasStickTrigger = true
		}
		if trigger.ItemType == "shell" {
			hasShellTrigger = true
		}
	}
	if !hasStickTrigger {
		t.Error("shell-hoe DiscoveryTriggers: missing stick trigger")
	}
	if !hasShellTrigger {
		t.Error("shell-hoe DiscoveryTriggers: missing shell trigger")
	}
}

// TestShellHoeRecipe_GetRecipesForActivity verifies recipe links to craftHoe activity
func TestShellHoeRecipe_GetRecipesForActivity(t *testing.T) {
	t.Parallel()

	recipes := GetRecipesForActivity("craftHoe")
	if len(recipes) == 0 {
		t.Fatal("GetRecipesForActivity(craftHoe): got 0 recipes, want at least 1")
	}

	found := false
	for _, r := range recipes {
		if r.ID == "shell-hoe" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetRecipesForActivity(craftHoe): shell-hoe not found")
	}
}

// TestCraftHoeActivity_Registered verifies craftHoe activity exists with correct properties
func TestCraftHoeActivity_Registered(t *testing.T) {
	t.Parallel()

	activity, ok := ActivityRegistry["craftHoe"]
	if !ok {
		t.Fatal("craftHoe activity not found in ActivityRegistry")
	}

	if activity.Name != "Hoe" {
		t.Errorf("craftHoe Name: got %q, want %q", activity.Name, "Hoe")
	}
	if activity.IntentFormation != IntentOrderable {
		t.Errorf("craftHoe IntentFormation: got %q, want %q", activity.IntentFormation, IntentOrderable)
	}
	if activity.Availability != AvailabilityKnowHow {
		t.Errorf("craftHoe Availability: got %q, want %q", activity.Availability, AvailabilityKnowHow)
	}
}

// TestCraftBrickRecipe_InRegistry verifies clay-brick recipe exists with correct properties (DD-19)
func TestCraftBrickRecipe_InRegistry(t *testing.T) {
	t.Parallel()

	recipe, ok := RecipeRegistry["clay-brick"]
	if !ok {
		t.Fatal("clay-brick recipe not found in RecipeRegistry")
	}

	if recipe.ActivityID != "craftBrick" {
		t.Errorf("clay-brick ActivityID: got %q, want %q", recipe.ActivityID, "craftBrick")
	}

	// Verify input: 1 clay
	if len(recipe.Inputs) != 1 {
		t.Fatalf("clay-brick Inputs count: got %d, want 1", len(recipe.Inputs))
	}
	if recipe.Inputs[0].ItemType != "clay" {
		t.Errorf("clay-brick Input ItemType: got %q, want %q", recipe.Inputs[0].ItemType, "clay")
	}
	if recipe.Inputs[0].Count != 1 {
		t.Errorf("clay-brick Input Count: got %d, want 1", recipe.Inputs[0].Count)
	}

	// Verify output
	if recipe.Output.ItemType != "brick" {
		t.Errorf("clay-brick Output.ItemType: got %q, want %q", recipe.Output.ItemType, "brick")
	}

	// Verify duration
	if recipe.Duration != config.ActionDurationLong {
		t.Errorf("clay-brick Duration: got %v, want %v", recipe.Duration, config.ActionDurationLong)
	}

	// Verify Repeatable — crafting loops until no clay remains (DD-19)
	if !recipe.Repeatable {
		t.Error("clay-brick Repeatable: got false, want true (processes all available clay)")
	}

	// Verify discovery triggers exist for clay
	if len(recipe.DiscoveryTriggers) == 0 {
		t.Fatal("clay-brick DiscoveryTriggers: got empty, want triggers for clay")
	}
	hasLookClay := false
	hasPickupClay := false
	hasDigClay := false
	for _, trigger := range recipe.DiscoveryTriggers {
		if trigger.Action == ActionLook && trigger.ItemType == "clay" {
			hasLookClay = true
		}
		if trigger.Action == ActionPickup && trigger.ItemType == "clay" {
			hasPickupClay = true
		}
		if trigger.Action == ActionDig && trigger.ItemType == "clay" {
			hasDigClay = true
		}
	}
	if !hasLookClay {
		t.Error("clay-brick DiscoveryTriggers: missing ActionLook clay trigger")
	}
	if !hasPickupClay {
		t.Error("clay-brick DiscoveryTriggers: missing ActionPickup clay trigger")
	}
	if !hasDigClay {
		t.Error("clay-brick DiscoveryTriggers: missing ActionDig clay trigger")
	}
}
