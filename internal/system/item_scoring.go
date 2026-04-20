package system

import (
	"petri/internal/config"
	"petri/internal/entity"
)

// ScoreConstructPreference computes a weighted preference score for a Construct.
// Kind contributes 2 points per valence; Material (ItemType) and Color contribute 1 each.
// Used for construction recipe selection where a synthetic Construct represents the anticipated output.
// Separate from NetPreferenceForConstruct (which uses unweighted AttributeCount) —
// the 2× Kind weight is a decision-making rule for item seeking, not a general preference strength.
func ScoreConstructPreference(char *entity.Character, construct *entity.Construct) int {
	score := 0
	for _, pref := range char.Preferences {
		if !pref.MatchesConstruct(construct) {
			continue
		}
		weight := 0
		if pref.Kind != "" {
			weight += 2
		}
		if pref.ItemType != "" {
			weight += 1
		}
		if pref.Color != "" {
			weight += 1
		}
		score += pref.Valence * weight
	}
	return score
}

// ScoreItemFit computes the unified item-seeking score.
// Score = (weightedPref × prefWeight) - (dist × distWeight)
// Used by all item-seeking call sites (construction material, craft recipe, item procurement,
// vessel selection, plantable selection).
func ScoreItemFit(weightedPref int, dist int) float64 {
	return float64(weightedPref)*config.ItemSeekPrefWeight -
		float64(dist)*config.ItemSeekDistWeight
}
