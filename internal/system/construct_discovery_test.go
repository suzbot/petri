package system

import (
	"strings"
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

// Anchor test: character looks at a fence, discovers buildHut activity + hut recipe.
// Second character who already knows buildHut looks at a different-material fence
// and discovers only the new recipe (no re-discovery of activity).
func TestConstructDiscovery_AnchorStory(t *testing.T) {
	// Character 1: knows nothing, looks at a stick fence
	char1 := &entity.Character{
		ID:              1,
		Name:            "Alice",
		KnownActivities: []string{},
		KnownRecipes:    []string{},
	}
	fence := entity.NewFence(5, 5, "stick", types.ColorBrown)
	log := NewActionLog(100)

	// Looking at the fence should trigger construct-based discovery
	TryDiscoverFromConstruct(char1, entity.ActionLook, fence.Kind, log, 1.0)

	// Should have discovered buildHut activity
	if !char1.KnowsActivity("buildHut") {
		t.Error("Expected Alice to discover buildHut activity from looking at fence")
	}

	// Should have discovered at least one hut recipe
	knowsAnyHutRecipe := char1.KnowsRecipe("thatch-hut") ||
		char1.KnowsRecipe("stick-hut") ||
		char1.KnowsRecipe("brick-hut")
	if !knowsAnyHutRecipe {
		t.Error("Expected Alice to discover a hut recipe from looking at fence")
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

	// Character 2: already knows buildHut + one recipe, looks at fence
	char2 := &entity.Character{
		ID:              2,
		Name:            "Bob",
		KnownActivities: []string{"buildHut"},
		KnownRecipes:    []string{"stick-hut"},
	}
	log2 := NewActionLog(100)

	// Discover remaining recipes by looking at fences repeatedly
	TryDiscoverFromConstruct(char2, entity.ActionLook, fence.Kind, log2, 1.0)

	// Should have discovered a new recipe (not stick-hut, which is already known)
	newRecipeCount := 0
	if char2.KnowsRecipe("thatch-hut") {
		newRecipeCount++
	}
	if char2.KnowsRecipe("brick-hut") {
		newRecipeCount++
	}
	if newRecipeCount == 0 {
		t.Error("Expected Bob to discover a new hut recipe")
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

// =============================================================================
// constructTriggerMatches unit tests
// =============================================================================

func TestConstructTriggerMatches_MatchesWhenBothMatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:        entity.ActionLook,
		ConstructKind: "fence",
	}
	if !constructTriggerMatches(trigger, entity.ActionLook, "fence") {
		t.Error("Expected match when Action and ConstructKind both match")
	}
}

func TestConstructTriggerMatches_RejectsActionMismatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:        entity.ActionLook,
		ConstructKind: "fence",
	}
	if constructTriggerMatches(trigger, entity.ActionPickup, "fence") {
		t.Error("Expected rejection when Action mismatches")
	}
}

func TestConstructTriggerMatches_RejectsConstructKindMismatch(t *testing.T) {
	t.Parallel()
	trigger := entity.DiscoveryTrigger{
		Action:        entity.ActionLook,
		ConstructKind: "fence",
	}
	if constructTriggerMatches(trigger, entity.ActionLook, "hut") {
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
	if constructTriggerMatches(trigger, entity.ActionLook, "fence") {
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

	discovered := TryDiscoverFromConstruct(char, entity.ActionLook, "fence", nil, 1.0)

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

	TryDiscoverFromConstruct(char, entity.ActionLook, "fence", log, 1.0)

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

	discovered := TryDiscoverFromConstruct(char, entity.ActionLook, "fence", log, 1.0)

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

	discovered := TryDiscoverFromConstruct(char, entity.ActionLook, "fence", nil, 0.0)

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

	discovered := tryDiscoverActivityFromConstruct(char, entity.ActionLook, "fence", nil, 1.0)

	if discovered {
		t.Error("Expected no discovery — no activities have construct triggers")
	}
}
