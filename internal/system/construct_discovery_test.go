package system

import (
	"strings"
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

// Anchor test: character looks at a stick fence, discovers buildHut activity + stick-hut recipe.
// Second character who already knows buildHut + stick-hut looks at a thatch fence
// and discovers thatch-hut (no re-discovery of activity).
func TestConstructDiscovery_AnchorStory(t *testing.T) {
	// Character 1: knows nothing, looks at a stick fence
	char1 := &entity.Character{
		ID:              1,
		Name:            "Alice",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}
	stickFence := entity.NewFence(5, 5, "stick", types.ColorBrown)
	log := NewActionLog(100)

	// Looking at the stick fence should discover stick-hut (material-matched)
	TryDiscoverFromConstruct(char1, entity.ActionLook, stickFence.Kind, stickFence.Material, log, 1.0)

	// Should have discovered buildHut activity
	if !char1.KnowsActivity("buildHut") {
		t.Error("Expected Alice to discover buildHut activity from looking at fence")
	}

	// Should have discovered stick-hut specifically
	if !char1.KnowsRecipe("stick-hut") {
		t.Error("Expected Alice to discover stick-hut from looking at stick fence")
	}
	// Should NOT have discovered other material hut recipes
	if char1.KnowsRecipe("thatch-hut") {
		t.Error("Alice should not discover thatch-hut from a stick fence")
	}
	if char1.KnowsRecipe("brick-hut") {
		t.Error("Alice should not discover brick-hut from a stick fence")
	}

	// Should have log entries for both activity and recipe discovery
	entries := log.Events(1, 0)
	discoveryCount := 0
	for _, e := range entries {
		if e.Type == "discovery" {
			discoveryCount++
		}
	}
	if discoveryCount < 2 {
		t.Errorf("Expected at least 2 discovery log entries (activity + recipe), got %d", discoveryCount)
	}

	// Character 2: already knows buildHut + stick-hut, looks at a thatch fence
	char2 := &entity.Character{
		ID:              2,
		Name:            "Bob",
		KnownActivities: []string{"buildHut"},
		KnownRecipes:    []string{"stick-hut"},
	}
	log2 := NewActionLog(100)

	thatchFence := entity.NewFence(5, 6, "grass", types.ColorPaleGreen)
	TryDiscoverFromConstruct(char2, entity.ActionLook, thatchFence.Kind, thatchFence.Material, log2, 1.0)

	// Should have discovered thatch-hut from the thatch fence
	if !char2.KnowsRecipe("thatch-hut") {
		t.Error("Expected Bob to discover thatch-hut from looking at thatch fence")
	}

	// Should NOT re-discover buildHut activity (check log has no activity discovery entry)
	entries2 := log2.Events(2, 0)
	for _, e := range entries2 {
		if e.Type == "discovery" {
			if strings.Contains(e.Message, "Discovered how to") {
				t.Error("Bob should not re-discover buildHut activity")
			}
		}
	}
}

// Regression test: looking at a stick fence does NOT discover thatch-hut or brick-hut
func TestConstructDiscovery_MaterialMismatchBlocksRecipe(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}

	// Look at stick fence with 100% chance — should only discover stick-hut
	TryDiscoverFromConstruct(char, entity.ActionLook, "fence", "stick", nil, 1.0)

	if !char.KnowsRecipe("stick-hut") {
		t.Error("Expected stick-hut from stick fence")
	}
	if char.KnowsRecipe("thatch-hut") {
		t.Error("Should not discover thatch-hut from stick fence")
	}
	if char.KnowsRecipe("brick-hut") {
		t.Error("Should not discover brick-hut from stick fence")
	}
}

// =============================================================================
// constructTriggerMatches unit tests
// =============================================================================

func TestConstructTriggerMatches_MatchesWhenAllMatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:            entity.ActionLook,
		ConstructKind:     "fence",
		ConstructMaterial: "stick",
	}
	if !constructTriggerMatches(trigger, entity.ActionLook, "fence", "stick") {
		t.Error("Expected match when Action, ConstructKind, and ConstructMaterial all match")
	}
}

func TestConstructTriggerMatches_MatchesWithoutMaterialFilter(t *testing.T) {
	t.Parallel()
	// Trigger with no ConstructMaterial matches any material
	trigger := entity.DiscoveryTrigger{
		Action:        entity.ActionLook,
		ConstructKind: "fence",
	}
	if !constructTriggerMatches(trigger, entity.ActionLook, "fence", "stick") {
		t.Error("Expected match when trigger has no material filter")
	}
}

func TestConstructTriggerMatches_RejectsMaterialMismatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:            entity.ActionLook,
		ConstructKind:     "fence",
		ConstructMaterial: "stick",
	}
	if constructTriggerMatches(trigger, entity.ActionLook, "fence", "grass") {
		t.Error("Expected rejection when ConstructMaterial mismatches")
	}
}

func TestConstructTriggerMatches_RejectsActionMismatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:        entity.ActionLook,
		ConstructKind: "fence",
	}
	if constructTriggerMatches(trigger, entity.ActionPickup, "fence", "stick") {
		t.Error("Expected rejection when Action mismatches")
	}
}

func TestConstructTriggerMatches_RejectsConstructKindMismatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:        entity.ActionLook,
		ConstructKind: "fence",
	}
	if constructTriggerMatches(trigger, entity.ActionLook, "hut", "stick") {
		t.Error("Expected rejection when ConstructKind mismatches")
	}
}

func TestConstructTriggerMatches_RejectsEmptyConstructKind(t *testing.T) {
	t.Parallel()
	// Item-only trigger (no ConstructKind) should never match construct interactions
	trigger := entity.DiscoveryTrigger{
		Action:   entity.ActionLook,
		ItemType: "grass",
	}
	if constructTriggerMatches(trigger, entity.ActionLook, "fence", "grass") {
		t.Error("Expected rejection when trigger has empty ConstructKind (item-only trigger)")
	}
}

// =============================================================================
// TryDiscoverFromConstruct unit tests
// =============================================================================

func TestTryDiscoverFromConstruct_DiscoversHutRecipe(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}

	discovered := TryDiscoverFromConstruct(char, entity.ActionLook, "fence", "stick", nil, 1.0)

	if !discovered {
		t.Error("Expected discovery with 100% chance")
	}
	if !char.KnowsActivity("buildHut") {
		t.Error("Expected character to know buildHut after discovery")
	}
}

func TestTryDiscoverFromConstruct_GrantsBuildHutOnFirstRecipe(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}
	log := NewActionLog(100)

	TryDiscoverFromConstruct(char, entity.ActionLook, "fence", "stick", log, 1.0)

	if !char.KnowsActivity("buildHut") {
		t.Error("Expected buildHut activity to be granted on first recipe discovery")
	}

	// Check log has activity discovery entry
	entries := log.Events(0, 0)
	hasActivityLog := false
	for _, e := range entries {
		if e.Type == "discovery" && strings.Contains(e.Message, "Discovered how to") {
			hasActivityLog = true
		}
	}
	if !hasActivityLog {
		t.Error("Expected activity discovery log entry on first recipe")
	}
}

func TestTryDiscoverFromConstruct_NoReGrantActivity(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{"buildHut"},
		KnownRecipes:    []string{"stick-hut"},
	}
	log := NewActionLog(100)

	// Look at a thatch fence — should discover thatch-hut without re-granting buildHut
	discovered := TryDiscoverFromConstruct(char, entity.ActionLook, "fence", "grass", log, 1.0)

	if !discovered {
		t.Error("Expected discovery of new recipe")
	}

	// Should NOT log activity discovery
	entries := log.Events(0, 0)
	for _, e := range entries {
		if e.Type == "discovery" && strings.Contains(e.Message, "Discovered how to") {
			t.Error("Should not re-log activity discovery when already known")
		}
	}
}

func TestTryDiscoverFromConstruct_RespectsZeroChance(t *testing.T) {
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}

	discovered := TryDiscoverFromConstruct(char, entity.ActionLook, "fence", "stick", nil, 0.0)

	if discovered {
		t.Error("Should not discover with 0% chance")
	}
	if char.KnowsActivity("buildHut") {
		t.Error("Should not know buildHut with 0% chance")
	}
}

// =============================================================================
// tryDiscoverActivityFromConstruct unit tests
// =============================================================================

func TestTryDiscoverActivityFromConstruct_NoConstructTriggeredActivities(t *testing.T) {
	t.Parallel()
	// Currently no activities have construct triggers (buildHut is discovered via recipes)
	char := &entity.Character{
		Name:            "Test",
		KnownActivities: []string{},
	}

	discovered := tryDiscoverActivityFromConstruct(char, entity.ActionLook, "fence", "stick", nil, 1.0)

	if discovered {
		t.Error("Expected no discovery — no activities have construct triggers")
	}
}
