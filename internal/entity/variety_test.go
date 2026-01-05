package entity

import (
	"testing"

	"petri/internal/types"
)

func TestItemVariety_Description(t *testing.T) {
	tests := []struct {
		name     string
		variety  ItemVariety
		expected string
	}{
		{
			name: "simple berry with color only",
			variety: ItemVariety{
				ItemType: "berry",
				Color:    types.ColorRed,
			},
			expected: "red berry",
		},
		{
			name: "mushroom with pattern",
			variety: ItemVariety{
				ItemType: "mushroom",
				Color:    types.ColorBrown,
				Pattern:  types.PatternSpotted,
			},
			expected: "spotted brown mushroom",
		},
		{
			name: "mushroom with texture",
			variety: ItemVariety{
				ItemType: "mushroom",
				Color:    types.ColorWhite,
				Texture:  types.TextureSlimy,
			},
			expected: "slimy white mushroom",
		},
		{
			name: "mushroom with texture only",
			variety: ItemVariety{
				ItemType: "mushroom",
				Color:    types.ColorRed,
				Texture:  types.TextureSlimy,
			},
			expected: "slimy red mushroom",
		},
		{
			name: "flower with color only",
			variety: ItemVariety{
				ItemType: "flower",
				Color:    types.ColorPurple,
			},
			expected: "purple flower",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.variety.Description()
			if got != tt.expected {
				t.Errorf("Description() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGenerateVarietyID(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		color    types.Color
		pattern  types.Pattern
		texture  types.Texture
		expected string
	}{
		{
			name:     "berry with color",
			itemType: "berry",
			color:    types.ColorRed,
			expected: "berry-red",
		},
		{
			name:     "mushroom with all attributes",
			itemType: "mushroom",
			color:    types.ColorBrown,
			pattern:  types.PatternSpotted,
			texture:  types.TextureSlimy,
			expected: "mushroom-brown-spotted-slimy",
		},
		{
			name:     "mushroom with color only",
			itemType: "mushroom",
			color:    types.ColorWhite,
			expected: "mushroom-white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateVarietyID(tt.itemType, tt.color, tt.pattern, tt.texture)
			if got != tt.expected {
				t.Errorf("GenerateVarietyID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestItemVariety_HasPatternAndTexture(t *testing.T) {
	v := ItemVariety{
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}

	if !v.HasPattern() {
		t.Error("HasPattern() should return true for spotted mushroom")
	}
	if v.HasTexture() {
		t.Error("HasTexture() should return false when no texture set")
	}

	v.Texture = types.TextureSlimy
	if !v.HasTexture() {
		t.Error("HasTexture() should return true for slimy mushroom")
	}
}
