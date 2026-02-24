package system

import (
	"math"

	"petri/internal/config"
)

// ScoreFoodFit computes the unified food selection score.
// Score = (netPref × prefWeight) - (dist × distWeight) - |hunger - satiation|
// Used by both hunger-driven food seeking (FindFoodTarget) and idle foraging (scoreForageItems).
func ScoreFoodFit(netPref int, dist int, hunger float64, itemType string, prefWeight, distWeight float64) float64 {
	satiation := config.GetMealSize(itemType).Satiation
	return float64(netPref)*prefWeight - float64(dist)*distWeight - math.Abs(hunger-satiation)
}
