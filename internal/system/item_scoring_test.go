package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/types"
)

func TestScoreConstructPreference_SoloKind(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Preferences: []entity.Preference{
			{Valence: 1, Kind: "stick fence"},
		},
	}
	construct := &entity.Construct{Kind: "fence", Material: "stick", MaterialColor: types.ColorBrown}
	got := ScoreConstructPreference(char, construct)
	if got != 2 {
		t.Errorf("Solo Kind: got %d, want 2 (Kind weight)", got)
	}
}

func TestScoreConstructPreference_SoloMaterial(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Preferences: []entity.Preference{
			{Valence: 1, ItemType: "stick"},
		},
	}
	construct := &entity.Construct{Kind: "fence", Material: "stick", MaterialColor: types.ColorBrown}
	got := ScoreConstructPreference(char, construct)
	if got != 1 {
		t.Errorf("Solo Material: got %d, want 1", got)
	}
}

func TestScoreConstructPreference_SoloColor(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Preferences: []entity.Preference{
			{Valence: 1, Color: types.ColorBrown},
		},
	}
	construct := &entity.Construct{Kind: "fence", Material: "stick", MaterialColor: types.ColorBrown}
	got := ScoreConstructPreference(char, construct)
	if got != 1 {
		t.Errorf("Solo Color: got %d, want 1", got)
	}
}

func TestScoreConstructPreference_ComboKindAndColor(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Preferences: []entity.Preference{
			{Valence: 1, Kind: "brick fence", Color: types.ColorTerracotta},
		},
	}
	construct := &entity.Construct{Kind: "fence", Material: "brick", MaterialColor: types.ColorTerracotta}
	got := ScoreConstructPreference(char, construct)
	if got != 3 {
		t.Errorf("Combo Kind+Color: got %d, want 3 (2+1)", got)
	}
}

func TestScoreConstructPreference_ConflictKindVsMaterial(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Preferences: []entity.Preference{
			{Valence: 1, Kind: "brick fence"},
			{Valence: -1, ItemType: "brick"},
		},
	}
	construct := &entity.Construct{Kind: "fence", Material: "brick", MaterialColor: types.ColorTerracotta}
	got := ScoreConstructPreference(char, construct)
	if got != 1 {
		t.Errorf("Conflict (likes recipe, dislikes material): got %d, want 1 (2-1)", got)
	}
}

func TestScoreConstructPreference_NoMatch(t *testing.T) {
	t.Parallel()
	char := &entity.Character{
		Preferences: []entity.Preference{
			{Valence: 1, Kind: "stick fence"},
		},
	}
	construct := &entity.Construct{Kind: "fence", Material: "brick", MaterialColor: types.ColorTerracotta}
	got := ScoreConstructPreference(char, construct)
	if got != 0 {
		t.Errorf("No match: got %d, want 0", got)
	}
}

func TestScoreItemFit_Basic(t *testing.T) {
	t.Parallel()
	got := ScoreItemFit(2, 25)
	want := 15.0 // 2*20 - 25*1
	if got != want {
		t.Errorf("ScoreItemFit(2, 25): got %f, want %f", got, want)
	}
}

func TestScoreItemFit_ZeroPreference(t *testing.T) {
	t.Parallel()
	got := ScoreItemFit(0, 10)
	want := -10.0 // 0*20 - 10*1
	if got != want {
		t.Errorf("ScoreItemFit(0, 10): got %f, want %f", got, want)
	}
}
