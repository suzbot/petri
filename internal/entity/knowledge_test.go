package entity

import (
	"petri/internal/types"
	"testing"
)

func TestKnowledge_Description(t *testing.T) {
	tests := []struct {
		name     string
		k        Knowledge
		expected string
	}{
		{
			name: "poisonous mushroom with all attributes",
			k: Knowledge{
				Category: KnowledgePoisonous,
				ItemType: "mushroom",
				Color:    types.ColorRed,
				Pattern:  types.PatternSpotted,
				Texture:  types.TextureSlimy,
			},
			expected: "Slimy spotted red mushrooms are poisonous",
		},
		{
			name: "healing berry",
			k: Knowledge{
				Category: KnowledgeHealing,
				ItemType: "berry",
				Color:    types.ColorBlue,
			},
			expected: "Blue berries are healing",
		},
		{
			name: "poisonous mushroom without texture",
			k: Knowledge{
				Category: KnowledgePoisonous,
				ItemType: "mushroom",
				Color:    types.ColorBrown,
				Pattern:  types.PatternSpotted,
			},
			expected: "Spotted brown mushrooms are poisonous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.k.Description()
			if got != tt.expected {
				t.Errorf("Description() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestKnowledge_Matches(t *testing.T) {
	knowledge := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
	}

	tests := []struct {
		name     string
		item     *Item
		expected bool
	}{
		{
			name: "exact match",
			item: &Item{
				ItemType: "mushroom",
				Color:    types.ColorRed,
				Pattern:  types.PatternSpotted,
				Texture:  types.TextureSlimy,
			},
			expected: true,
		},
		{
			name: "different color",
			item: &Item{
				ItemType: "mushroom",
				Color:    types.ColorBlue,
				Pattern:  types.PatternSpotted,
				Texture:  types.TextureSlimy,
			},
			expected: false,
		},
		{
			name: "different item type",
			item: &Item{
				ItemType: "berry",
				Color:    types.ColorRed,
			},
			expected: false,
		},
		{
			name: "different pattern",
			item: &Item{
				ItemType: "mushroom",
				Color:    types.ColorRed,
				Pattern:  types.PatternNone,
				Texture:  types.TextureSlimy,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := knowledge.Matches(tt.item)
			if got != tt.expected {
				t.Errorf("Matches() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKnowledge_Matches_Berry(t *testing.T) {
	// Berry knowledge has no pattern/texture
	knowledge := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}

	matchingItem := &Item{
		ItemType: "berry",
		Color:    types.ColorBlue,
	}

	nonMatchingItem := &Item{
		ItemType: "berry",
		Color:    types.ColorRed,
	}

	if !knowledge.Matches(matchingItem) {
		t.Error("expected berry knowledge to match blue berry")
	}

	if knowledge.Matches(nonMatchingItem) {
		t.Error("expected berry knowledge to not match red berry")
	}
}

func TestKnowledge_Equals(t *testing.T) {
	k1 := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}

	k2 := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}

	k3 := Knowledge{
		Category: KnowledgeHealing, // different category
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}

	k4 := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorBlue, // different color
		Pattern:  types.PatternSpotted,
	}

	if !k1.Equals(k2) {
		t.Error("expected k1 to equal k2")
	}

	if k1.Equals(k3) {
		t.Error("expected k1 to not equal k3 (different category)")
	}

	if k1.Equals(k4) {
		t.Error("expected k1 to not equal k4 (different color)")
	}
}

func TestNewKnowledgeFromItem(t *testing.T) {
	item := &Item{
		ItemType:  "mushroom",
		Color:     types.ColorRed,
		Pattern:   types.PatternSpotted,
		Texture:   types.TextureSlimy,
		Poisonous: true,
	}

	k := NewKnowledgeFromItem(item, KnowledgePoisonous)

	if k.Category != KnowledgePoisonous {
		t.Errorf("expected category %s, got %s", KnowledgePoisonous, k.Category)
	}
	if k.ItemType != "mushroom" {
		t.Errorf("expected itemType mushroom, got %s", k.ItemType)
	}
	if k.Color != types.ColorRed {
		t.Errorf("expected color red, got %s", k.Color)
	}
	if k.Pattern != types.PatternSpotted {
		t.Errorf("expected pattern spotted, got %s", k.Pattern)
	}
	if k.Texture != types.TextureSlimy {
		t.Errorf("expected texture slimy, got %s", k.Texture)
	}
}
