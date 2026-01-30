package game

import (
	"testing"

	"petri/internal/config"
	"petri/internal/types"
)

func TestGenerateVarieties_CreatesExpectedCounts(t *testing.T) {
	registry := GenerateVarieties()

	// Check we have varieties for each item type
	berries := registry.VarietiesOfType("berry")
	mushrooms := registry.VarietiesOfType("mushroom")
	flowers := registry.VarietiesOfType("flower")

	if len(berries) < config.VarietyMinCount {
		t.Errorf("Expected at least %d berry varieties, got %d", config.VarietyMinCount, len(berries))
	}

	if len(mushrooms) < config.VarietyMinCount {
		t.Errorf("Expected at least %d mushroom varieties, got %d", config.VarietyMinCount, len(mushrooms))
	}

	if len(flowers) < config.VarietyMinCount {
		t.Errorf("Expected at least %d flower varieties, got %d", config.VarietyMinCount, len(flowers))
	}
}

func TestGenerateVarieties_MushroomsCanHavePatternOrTexture(t *testing.T) {
	// Test that mushroom config allows patterns and textures
	// (unlike berries which don't support them)
	configs := GetItemTypeConfigs()

	mushroomCfg, ok := configs["mushroom"]
	if !ok {
		t.Fatal("Expected mushroom config to exist")
	}

	if mushroomCfg.Patterns == nil || len(mushroomCfg.Patterns) == 0 {
		t.Error("Expected mushroom config to support patterns")
	}
	if mushroomCfg.Textures == nil || len(mushroomCfg.Textures) == 0 {
		t.Error("Expected mushroom config to support textures")
	}

	// Verify patterns include a non-None option
	hasNonNonePattern := false
	for _, p := range mushroomCfg.Patterns {
		if p != types.PatternNone {
			hasNonNonePattern = true
			break
		}
	}
	if !hasNonNonePattern {
		t.Error("Expected mushroom patterns to include at least one non-None pattern")
	}

	// Verify textures include a non-None option
	hasNonNoneTexture := false
	for _, tex := range mushroomCfg.Textures {
		if tex != types.TextureNone {
			hasNonNoneTexture = true
			break
		}
	}
	if !hasNonNoneTexture {
		t.Error("Expected mushroom textures to include at least one non-None texture")
	}
}

func TestGenerateVarieties_BerriesHaveNoPatternOrTexture(t *testing.T) {
	registry := GenerateVarieties()
	berries := registry.VarietiesOfType("berry")

	for _, b := range berries {
		if b.Pattern != "" {
			t.Errorf("Berry variety %q should not have pattern, got %q", b.ID, b.Pattern)
		}
		if b.Texture != "" {
			t.Errorf("Berry variety %q should not have texture, got %q", b.ID, b.Texture)
		}
	}
}

func TestGenerateVarieties_FlowersAreNotEdible(t *testing.T) {
	registry := GenerateVarieties()
	flowers := registry.VarietiesOfType("flower")

	for _, f := range flowers {
		if f.IsEdible() {
			t.Errorf("Flower variety %q should not be edible", f.ID)
		}
		if f.IsPoisonous() {
			t.Errorf("Flower variety %q should not be poisonous", f.ID)
		}
		if f.IsHealing() {
			t.Errorf("Flower variety %q should not be healing", f.ID)
		}
	}
}

func TestGenerateVarieties_PoisonAndHealingAssigned(t *testing.T) {
	registry := GenerateVarieties()
	edible := registry.EdibleVarieties()

	if len(edible) < 2 {
		t.Skip("Need at least 2 edible varieties to test poison/healing")
	}

	var poisonCount, healingCount int
	for _, v := range edible {
		if v.IsPoisonous() {
			poisonCount++
		}
		if v.IsHealing() {
			healingCount++
		}
		// Check no variety is both poisonous and healing
		if v.IsPoisonous() && v.IsHealing() {
			t.Errorf("Variety %q is both poisonous and healing", v.ID)
		}
	}

	if poisonCount == 0 {
		t.Error("Expected at least one poisonous variety")
	}
	if healingCount == 0 {
		t.Error("Expected at least one healing variety")
	}
}

func TestGenerateVarieties_UniqueIDs(t *testing.T) {
	registry := GenerateVarieties()
	all := registry.AllVarieties()

	seen := make(map[string]bool)
	for _, v := range all {
		if seen[v.ID] {
			t.Errorf("Duplicate variety ID: %q", v.ID)
		}
		seen[v.ID] = true
	}
}

func TestGenerateVarieties_CorrectSymbols(t *testing.T) {
	registry := GenerateVarieties()

	for _, v := range registry.VarietiesOfType("berry") {
		if v.Sym != config.CharBerry {
			t.Errorf("Berry variety %q has wrong symbol %c, want %c", v.ID, v.Sym, config.CharBerry)
		}
	}

	for _, v := range registry.VarietiesOfType("mushroom") {
		if v.Sym != config.CharMushroom {
			t.Errorf("Mushroom variety %q has wrong symbol %c, want %c", v.ID, v.Sym, config.CharMushroom)
		}
	}

	for _, v := range registry.VarietiesOfType("flower") {
		if v.Sym != config.CharFlower {
			t.Errorf("Flower variety %q has wrong symbol %c, want %c", v.ID, v.Sym, config.CharFlower)
		}
	}
}
