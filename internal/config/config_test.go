package config

import "testing"

func TestGetMealSize_SatiationValuesUnchanged(t *testing.T) {
	t.Parallel()

	// Verify satiation amounts match the previous constant values exactly
	tests := []struct {
		itemType string
		want     float64
	}{
		{"gourd", 50.0},
		{"mushroom", 25.0},
		{"berry", 10.0},
		{"nut", 10.0},
	}

	for _, tt := range tests {
		ms := GetMealSize(tt.itemType)
		if ms.Satiation != tt.want {
			t.Errorf("GetMealSize(%q).Satiation = %.3f, want %.3f", tt.itemType, ms.Satiation, tt.want)
		}
	}
}

func TestGetMealSize_DurationValues(t *testing.T) {
	t.Parallel()

	// Feast ~45 world mins, Meal ~15 world mins, Snack ~5 world mins
	tests := []struct {
		itemType string
		want     float64
	}{
		{"gourd", 3.75},
		{"mushroom", 1.25},
		{"berry", 0.417},
		{"nut", 0.417},
	}

	for _, tt := range tests {
		ms := GetMealSize(tt.itemType)
		if ms.Duration != tt.want {
			t.Errorf("GetMealSize(%q).Duration = %.3f, want %.3f", tt.itemType, ms.Duration, tt.want)
		}
	}
}

func TestGetMealSize_UnknownFallsBackToMeal(t *testing.T) {
	t.Parallel()

	ms := GetMealSize("unknown_type")
	if ms.Satiation != MealSizeMeal.Satiation {
		t.Errorf("Unknown item satiation: got %.2f, want %.2f", ms.Satiation, MealSizeMeal.Satiation)
	}
	if ms.Duration != MealSizeMeal.Duration {
		t.Errorf("Unknown item duration: got %.3f, want %.3f", ms.Duration, MealSizeMeal.Duration)
	}
}
