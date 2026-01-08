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
	registry := GenerateVarieties()
	mushrooms := registry.VarietiesOfType("mushroom")

	if len(mushrooms) == 0 {
		t.Fatal("Expected at least one mushroom variety")
	}

	// Mushrooms may have pattern (spotted) and/or texture (slimy)
	// PatternNone and TextureNone are valid - not all mushrooms have distinguishing features
	hasSpotted := false
	hasSlimy := false
	for _, m := range mushrooms {
		if m.Pattern == types.PatternSpotted {
			hasSpotted = true
		}
		if m.Texture == types.TextureSlimy {
			hasSlimy = true
		}
	}

	// With enough varieties, we should see both pattern and texture options used
	if len(mushrooms) >= 4 && !hasSpotted {
		t.Error("Expected at least one spotted mushroom variety")
	}
	if len(mushrooms) >= 4 && !hasSlimy {
		t.Error("Expected at least one slimy mushroom variety")
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
		if f.Edible {
			t.Errorf("Flower variety %q should not be edible", f.ID)
		}
		if f.Poisonous {
			t.Errorf("Flower variety %q should not be poisonous", f.ID)
		}
		if f.Healing {
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
		if v.Poisonous {
			poisonCount++
		}
		if v.Healing {
			healingCount++
		}
		// Check no variety is both poisonous and healing
		if v.Poisonous && v.Healing {
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
