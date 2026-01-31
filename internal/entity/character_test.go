package entity

import (
	"testing"

	"petri/internal/config"
	"petri/internal/types"
)

// TestHungerTier verifies hunger tier calculation at all boundaries
func TestHungerTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hunger   float64
		expected int
	}{
		// TierNone: below 50
		{"hunger 0 is TierNone", 0, TierNone},
		{"hunger 49 is TierNone", 49, TierNone},
		// TierMild: 50-74
		{"hunger 50 is TierMild", 50, TierMild},
		{"hunger 74 is TierMild", 74, TierMild},
		// TierModerate: 75-89
		{"hunger 75 is TierModerate", 75, TierModerate},
		{"hunger 89 is TierModerate", 89, TierModerate},
		// TierSevere: 90-99
		{"hunger 90 is TierSevere", 90, TierSevere},
		{"hunger 99 is TierSevere", 99, TierSevere},
		// TierCrisis: 100
		{"hunger 100 is TierCrisis", 100, TierCrisis},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Hunger: tt.hunger}
			got := c.HungerTier()
			if got != tt.expected {
				t.Errorf("HungerTier() with hunger=%.0f: got %d, want %d", tt.hunger, got, tt.expected)
			}
		})
	}
}

// TestThirstTier verifies thirst tier calculation at all boundaries
func TestThirstTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		thirst   float64
		expected int
	}{
		// TierNone: below 50
		{"thirst 0 is TierNone", 0, TierNone},
		{"thirst 49 is TierNone", 49, TierNone},
		// TierMild: 50-74
		{"thirst 50 is TierMild", 50, TierMild},
		{"thirst 74 is TierMild", 74, TierMild},
		// TierModerate: 75-89
		{"thirst 75 is TierModerate", 75, TierModerate},
		{"thirst 89 is TierModerate", 89, TierModerate},
		// TierSevere: 90-99
		{"thirst 90 is TierSevere", 90, TierSevere},
		{"thirst 99 is TierSevere", 99, TierSevere},
		// TierCrisis: 100
		{"thirst 100 is TierCrisis", 100, TierCrisis},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Thirst: tt.thirst}
			got := c.ThirstTier()
			if got != tt.expected {
				t.Errorf("ThirstTier() with thirst=%.0f: got %d, want %d", tt.thirst, got, tt.expected)
			}
		})
	}
}

// TestEnergyTier verifies energy tier calculation (inverted scale)
func TestEnergyTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		energy   float64
		expected int
	}{
		// TierNone: above 50
		{"energy 100 is TierNone", 100, TierNone},
		{"energy 51 is TierNone", 51, TierNone},
		// TierMild: 26-50
		{"energy 50 is TierMild", 50, TierMild},
		{"energy 26 is TierMild", 26, TierMild},
		// TierModerate: 11-25
		{"energy 25 is TierModerate", 25, TierModerate},
		{"energy 11 is TierModerate", 11, TierModerate},
		// TierSevere: 1-10
		{"energy 10 is TierSevere", 10, TierSevere},
		{"energy 1 is TierSevere", 1, TierSevere},
		// TierCrisis: 0
		{"energy 0 is TierCrisis", 0, TierCrisis},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Energy: tt.energy}
			got := c.EnergyTier()
			if got != tt.expected {
				t.Errorf("EnergyTier() with energy=%.0f: got %d, want %d", tt.energy, got, tt.expected)
			}
		})
	}
}

// TestHealthTier verifies health tier calculation (inverted scale)
func TestHealthTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		health   float64
		expected int
	}{
		// TierNone: above 75
		{"health 100 is TierNone", 100, TierNone},
		{"health 76 is TierNone", 76, TierNone},
		// TierMild: 51-75
		{"health 75 is TierMild", 75, TierMild},
		{"health 51 is TierMild", 51, TierMild},
		// TierModerate: 26-50
		{"health 50 is TierModerate", 50, TierModerate},
		{"health 26 is TierModerate", 26, TierModerate},
		// TierSevere: 11-25
		{"health 25 is TierSevere", 25, TierSevere},
		{"health 11 is TierSevere", 11, TierSevere},
		// TierCrisis: 10 or below
		{"health 10 is TierCrisis", 10, TierCrisis},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Health: tt.health}
			got := c.HealthTier()
			if got != tt.expected {
				t.Errorf("HealthTier() with health=%.0f: got %d, want %d", tt.health, got, tt.expected)
			}
		})
	}
}

// TestHungerUrgency verifies hunger urgency equals hunger value
func TestHungerUrgency(t *testing.T) {
	t.Parallel()

	tests := []float64{0, 25, 50, 75, 100}
	for _, hunger := range tests {
		c := &Character{Hunger: hunger}
		got := c.HungerUrgency()
		if got != hunger {
			t.Errorf("HungerUrgency() with hunger=%.0f: got %.0f, want %.0f", hunger, got, hunger)
		}
	}
}

// TestThirstUrgency verifies thirst urgency equals thirst value
func TestThirstUrgency(t *testing.T) {
	t.Parallel()

	tests := []float64{0, 25, 50, 75, 100}
	for _, thirst := range tests {
		c := &Character{Thirst: thirst}
		got := c.ThirstUrgency()
		if got != thirst {
			t.Errorf("ThirstUrgency() with thirst=%.0f: got %.0f, want %.0f", thirst, got, thirst)
		}
	}
}

// TestEnergyUrgency verifies energy urgency is inverted (100 - energy)
func TestEnergyUrgency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		energy   float64
		expected float64
	}{
		{100, 0},
		{75, 25},
		{50, 50},
		{25, 75},
		{0, 100},
	}

	for _, tt := range tests {
		c := &Character{Energy: tt.energy}
		got := c.EnergyUrgency()
		if got != tt.expected {
			t.Errorf("EnergyUrgency() with energy=%.0f: got %.0f, want %.0f", tt.energy, got, tt.expected)
		}
	}
}

// TestEffectiveSpeed_Healthy verifies base speed with no penalties
func TestEffectiveSpeed_Healthy(t *testing.T) {
	t.Parallel()

	c := &Character{
		Thirst: 0,  // below parched threshold
		Energy: 100, // above tired threshold
	}
	got := c.EffectiveSpeed()
	if got != config.BaseSpeed {
		t.Errorf("EffectiveSpeed() healthy character: got %d, want %d", got, config.BaseSpeed)
	}
}

// TestEffectiveSpeed_Poisoned verifies poison penalty
func TestEffectiveSpeed_Poisoned(t *testing.T) {
	t.Parallel()

	c := &Character{
		Poisoned: true,
		Thirst:   0,
		Energy:   100,
	}
	expected := config.BaseSpeed - config.PoisonSpeedPenalty
	got := c.EffectiveSpeed()
	if got != expected {
		t.Errorf("EffectiveSpeed() poisoned: got %d, want %d", got, expected)
	}
}

// TestEffectiveSpeed_Parched verifies thirst >= 90 penalty
func TestEffectiveSpeed_Parched(t *testing.T) {
	t.Parallel()

	c := &Character{
		Thirst: 90, // parched but not dehydrated
		Energy: 100,
	}
	expected := config.BaseSpeed - config.ParchedSpeedPenalty
	got := c.EffectiveSpeed()
	if got != expected {
		t.Errorf("EffectiveSpeed() parched: got %d, want %d", got, expected)
	}
}

// TestEffectiveSpeed_Dehydrated verifies thirst >= 100 stacks penalties
func TestEffectiveSpeed_Dehydrated(t *testing.T) {
	t.Parallel()

	c := &Character{
		Thirst: 100, // dehydrated (parched + dehydrated penalties)
		Energy: 100,
	}
	expected := config.BaseSpeed - config.ParchedSpeedPenalty - config.DehydratedSpeedPenalty
	got := c.EffectiveSpeed()
	if got != expected {
		t.Errorf("EffectiveSpeed() dehydrated: got %d, want %d", got, expected)
	}
}

// TestEffectiveSpeed_VeryTired verifies energy <= 25 penalty
func TestEffectiveSpeed_VeryTired(t *testing.T) {
	t.Parallel()

	c := &Character{
		Thirst: 0,
		Energy: 25, // very tired but not exhausted
	}
	expected := config.BaseSpeed - config.VeryTiredSpeedPenalty
	got := c.EffectiveSpeed()
	if got != expected {
		t.Errorf("EffectiveSpeed() very tired: got %d, want %d", got, expected)
	}
}

// TestEffectiveSpeed_Exhausted verifies energy <= 10 stacks penalties
func TestEffectiveSpeed_Exhausted(t *testing.T) {
	t.Parallel()

	c := &Character{
		Thirst: 0,
		Energy: 10, // exhausted (very tired + exhausted penalties)
	}
	expected := config.BaseSpeed - config.VeryTiredSpeedPenalty - config.ExhaustedSpeedPenalty
	got := c.EffectiveSpeed()
	if got != expected {
		t.Errorf("EffectiveSpeed() exhausted: got %d, want %d", got, expected)
	}
}

// TestEffectiveSpeed_AllPenalties verifies all penalties stack
func TestEffectiveSpeed_AllPenalties(t *testing.T) {
	t.Parallel()

	c := &Character{
		Poisoned: true,
		Thirst:   100, // dehydrated
		Energy:   10,  // exhausted
	}
	expected := config.BaseSpeed -
		config.PoisonSpeedPenalty -
		config.ParchedSpeedPenalty -
		config.DehydratedSpeedPenalty -
		config.VeryTiredSpeedPenalty -
		config.ExhaustedSpeedPenalty

	// Should not fall below minimum
	if expected < config.MinSpeed {
		expected = config.MinSpeed
	}

	got := c.EffectiveSpeed()
	if got != expected {
		t.Errorf("EffectiveSpeed() all penalties: got %d, want %d", got, expected)
	}
}

// TestEffectiveSpeed_MinimumFloor verifies speed doesn't go below minimum
func TestEffectiveSpeed_MinimumFloor(t *testing.T) {
	t.Parallel()

	c := &Character{
		Poisoned: true,
		Thirst:   100,
		Energy:   0,
	}
	got := c.EffectiveSpeed()
	if got < config.MinSpeed {
		t.Errorf("EffectiveSpeed() should not go below MinSpeed: got %d, want >= %d", got, config.MinSpeed)
	}
	if got != config.MinSpeed {
		t.Errorf("EffectiveSpeed() with max penalties should equal MinSpeed: got %d, want %d", got, config.MinSpeed)
	}
}

// TestIsInCrisis_HungerCrisis verifies crisis detection for hunger
func TestIsInCrisis_HungerCrisis(t *testing.T) {
	t.Parallel()

	c := &Character{
		Hunger: 100, // crisis threshold
		Thirst: 0,
		Energy: 100,
		Health: 100,
	}
	if !c.IsInCrisis() {
		t.Error("IsInCrisis() should return true when hunger is at crisis threshold")
	}
}

// TestIsInCrisis_ThirstCrisis verifies crisis detection for thirst
func TestIsInCrisis_ThirstCrisis(t *testing.T) {
	t.Parallel()

	c := &Character{
		Hunger: 0,
		Thirst: 100, // crisis threshold
		Energy: 100,
		Health: 100,
	}
	if !c.IsInCrisis() {
		t.Error("IsInCrisis() should return true when thirst is at crisis threshold")
	}
}

// TestIsInCrisis_EnergyCrisis verifies crisis detection for energy
func TestIsInCrisis_EnergyCrisis(t *testing.T) {
	t.Parallel()

	c := &Character{
		Hunger: 0,
		Thirst: 0,
		Energy: 0, // crisis threshold (inverted scale)
		Health: 100,
	}
	if !c.IsInCrisis() {
		t.Error("IsInCrisis() should return true when energy is at crisis threshold")
	}
}

// TestIsInCrisis_HealthCrisis verifies crisis detection for health
func TestIsInCrisis_HealthCrisis(t *testing.T) {
	t.Parallel()

	c := &Character{
		Hunger: 0,
		Thirst: 0,
		Energy: 100,
		Health: 10, // crisis threshold (inverted scale)
	}
	if !c.IsInCrisis() {
		t.Error("IsInCrisis() should return true when health is at crisis threshold")
	}
}

// TestIsInCrisis_NoCrisis verifies no crisis when all stats are safe
func TestIsInCrisis_NoCrisis(t *testing.T) {
	t.Parallel()

	c := &Character{
		Hunger: 0,
		Thirst: 0,
		Energy: 100,
		Health: 100,
	}
	if c.IsInCrisis() {
		t.Error("IsInCrisis() should return false when no stats are at crisis threshold")
	}
}

// TestIsInCrisis_SevereButNotCrisis verifies severe tier doesn't trigger crisis
func TestIsInCrisis_SevereButNotCrisis(t *testing.T) {
	t.Parallel()

	c := &Character{
		Hunger: 90, // severe but not crisis (hunger severe = 90)
		Thirst: 90, // severe but not crisis (thirst severe = 90)
		Energy: 10, // severe but not crisis (energy severe = 10, inverted)
		Health: 25, // severe but not crisis (health severe = 25, inverted)
	}
	if c.IsInCrisis() {
		t.Error("IsInCrisis() should return false when stats are at severe but not crisis")
	}
}

// TestHungerLevel verifies hunger level descriptions at all thresholds
func TestHungerLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hunger   float64
		expected string
	}{
		{0, "Not Hungry"},
		{49, "Not Hungry"},
		{50, "Hungry"},
		{74, "Hungry"},
		{75, "Very Hungry"},
		{89, "Very Hungry"},
		{90, "Ravenous"},
		{99, "Ravenous"},
		{100, "Starving"},
	}

	for _, tt := range tests {
		c := &Character{Hunger: tt.hunger}
		got := c.HungerLevel()
		if got != tt.expected {
			t.Errorf("HungerLevel() with hunger=%.0f: got %q, want %q", tt.hunger, got, tt.expected)
		}
	}
}

// TestThirstLevel verifies thirst level descriptions at all thresholds
func TestThirstLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		thirst   float64
		expected string
	}{
		{0, "Hydrated"},
		{49, "Hydrated"},
		{50, "Thirsty"},
		{74, "Thirsty"},
		{75, "Very Thirsty"},
		{89, "Very Thirsty"},
		{90, "Parched"},
		{99, "Parched"},
		{100, "Dehydrated"},
	}

	for _, tt := range tests {
		c := &Character{Thirst: tt.thirst}
		got := c.ThirstLevel()
		if got != tt.expected {
			t.Errorf("ThirstLevel() with thirst=%.0f: got %q, want %q", tt.thirst, got, tt.expected)
		}
	}
}

// TestEnergyLevel verifies energy level descriptions at all thresholds
func TestEnergyLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		energy   float64
		expected string
	}{
		{100, "Rested"},
		{51, "Rested"},
		{50, "Tired"},
		{26, "Tired"},
		{25, "Very Tired"},
		{11, "Very Tired"},
		{10, "Exhausted"},
		{1, "Exhausted"},
		{0, "Collapsed"},
	}

	for _, tt := range tests {
		c := &Character{Energy: tt.energy}
		got := c.EnergyLevel()
		if got != tt.expected {
			t.Errorf("EnergyLevel() with energy=%.0f: got %q, want %q", tt.energy, got, tt.expected)
		}
	}
}

// TestHealthLevel verifies health level descriptions at all thresholds
func TestHealthLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		health   float64
		expected string
	}{
		{100, "Healthy"},
		{76, "Healthy"},
		{75, "Poor"},
		{51, "Poor"},
		{50, "Very Poor"},
		{26, "Very Poor"},
		{25, "Critical"},
		{11, "Critical"},
		{10, "Dying"},
	}

	for _, tt := range tests {
		c := &Character{Health: tt.health}
		got := c.HealthLevel()
		if got != tt.expected {
			t.Errorf("HealthLevel() with health=%.0f: got %q, want %q", tt.health, got, tt.expected)
		}
	}
}

// TestMoodTier verifies mood tier calculation (inverted scale, higher is better)
func TestMoodTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mood     float64
		expected int
	}{
		// TierNone (Joyful): 90-100
		{"mood 100 is TierNone (Joyful)", 100, TierNone},
		{"mood 90 is TierNone (Joyful)", 90, TierNone},
		// TierMild (Happy): 65-89
		{"mood 89 is TierMild (Happy)", 89, TierMild},
		{"mood 65 is TierMild (Happy)", 65, TierMild},
		// TierModerate (Neutral): 35-64
		{"mood 64 is TierModerate (Neutral)", 64, TierModerate},
		{"mood 35 is TierModerate (Neutral)", 35, TierModerate},
		// TierSevere (Unhappy): 11-34
		{"mood 34 is TierSevere (Unhappy)", 34, TierSevere},
		{"mood 11 is TierSevere (Unhappy)", 11, TierSevere},
		// TierCrisis (Miserable): 0-10
		{"mood 10 is TierCrisis (Miserable)", 10, TierCrisis},
		{"mood 0 is TierCrisis (Miserable)", 0, TierCrisis},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Mood: tt.mood}
			got := c.MoodTier()
			if got != tt.expected {
				t.Errorf("MoodTier() with mood=%.0f: got %d, want %d", tt.mood, got, tt.expected)
			}
		})
	}
}

// TestMoodLevel verifies mood level descriptions at all thresholds
func TestMoodLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mood     float64
		expected string
	}{
		{100, "Joyful"},
		{90, "Joyful"},
		{89, "Happy"},
		{65, "Happy"},
		{64, "Neutral"},
		{35, "Neutral"},
		{34, "Unhappy"},
		{11, "Unhappy"},
		{10, "Miserable"},
		{0, "Miserable"},
	}

	for _, tt := range tests {
		c := &Character{Mood: tt.mood}
		got := c.MoodLevel()
		if got != tt.expected {
			t.Errorf("MoodLevel() with mood=%.0f: got %q, want %q", tt.mood, got, tt.expected)
		}
	}
}

// TestStatusText_Dead verifies dead status takes priority
func TestStatusText_Dead(t *testing.T) {
	t.Parallel()

	c := &Character{IsDead: true}
	got := c.StatusText()
	if got != "DEAD" {
		t.Errorf("StatusText() with IsDead=true: got %q, want %q", got, "DEAD")
	}
}

// TestStatusText_Sleeping verifies sleeping status
func TestStatusText_Sleeping(t *testing.T) {
	t.Parallel()

	c := &Character{IsDead: false, IsSleeping: true}
	got := c.StatusText()
	if got != "SLEEPING" {
		t.Errorf("StatusText() with IsSleeping=true: got %q, want %q", got, "SLEEPING")
	}
}

// TestStatusText_Poisoned verifies poisoned status
func TestStatusText_Poisoned(t *testing.T) {
	t.Parallel()

	c := &Character{IsDead: false, IsSleeping: false, Poisoned: true}
	got := c.StatusText()
	if got != "POISONED" {
		t.Errorf("StatusText() with Poisoned=true: got %q, want %q", got, "POISONED")
	}
}

// TestStatusText_Healthy verifies healthy status
func TestStatusText_Healthy(t *testing.T) {
	t.Parallel()

	c := &Character{IsDead: false, IsSleeping: false, Poisoned: false}
	got := c.StatusText()
	if got != "Healthy" {
		t.Errorf("StatusText() healthy character: got %q, want %q", got, "Healthy")
	}
}

// TestStatusText_Priority verifies Dead > Sleeping > Poisoned > Healthy
func TestStatusText_Priority(t *testing.T) {
	t.Parallel()

	c := &Character{IsDead: true, IsSleeping: true, Poisoned: true}
	got := c.StatusText()
	if got != "DEAD" {
		t.Errorf("StatusText() priority test: got %q, want %q", got, "DEAD")
	}
}

// TestNewCharacter_Pos verifies position is set correctly
func TestNewCharacter_Pos(t *testing.T) {
	t.Parallel()

	c := NewCharacter(1, 5, 10, "Test", "berry", types.ColorRed)
	pos := c.Pos()
	if pos.X != 5 || pos.Y != 10 {
		t.Errorf("NewCharacter Pos(): got (%d, %d), want (5, 10)", pos.X, pos.Y)
	}
}

// TestNewCharacter_Identity verifies identity fields are set correctly
func TestNewCharacter_Identity(t *testing.T) {
	t.Parallel()

	c := NewCharacter(1, 0, 0, "Luna", "mushroom", types.ColorBlue)
	if c.ID != 1 {
		t.Errorf("NewCharacter ID: got %d, want 1", c.ID)
	}
	if c.Name != "Luna" {
		t.Errorf("NewCharacter Name: got %q, want %q", c.Name, "Luna")
	}
	// Should have two preferences: one for food type, one for color
	if len(c.Preferences) != 2 {
		t.Fatalf("NewCharacter Preferences: got %d preferences, want 2", len(c.Preferences))
	}
	// First preference: likes mushrooms
	if c.Preferences[0].ItemType != "mushroom" || c.Preferences[0].Valence != 1 {
		t.Errorf("NewCharacter Preferences[0]: got %+v, want likes mushroom", c.Preferences[0])
	}
	// Second preference: likes blue
	if c.Preferences[1].Color != types.ColorBlue || c.Preferences[1].Valence != 1 {
		t.Errorf("NewCharacter Preferences[1]: got %+v, want likes blue", c.Preferences[1])
	}
}

// TestNewCharacter_SurvivalStats verifies survival stats default values
func TestNewCharacter_SurvivalStats(t *testing.T) {
	t.Parallel()

	c := NewCharacter(1, 0, 0, "Test", "berry", types.ColorRed)
	if c.Health != 100 {
		t.Errorf("NewCharacter Health: got %.0f, want 100", c.Health)
	}
	if c.Hunger != 50 {
		t.Errorf("NewCharacter Hunger: got %.0f, want 50", c.Hunger)
	}
	if c.Thirst != 50 {
		t.Errorf("NewCharacter Thirst: got %.0f, want 50", c.Thirst)
	}
	if c.Energy != 100 {
		t.Errorf("NewCharacter Energy: got %.0f, want 100", c.Energy)
	}
}

// TestNewCharacter_StatusFlags verifies status flags default to healthy
func TestNewCharacter_StatusFlags(t *testing.T) {
	t.Parallel()

	c := NewCharacter(1, 0, 0, "Test", "berry", types.ColorRed)
	if c.IsDead {
		t.Error("NewCharacter IsDead: got true, want false")
	}
	if c.IsSleeping {
		t.Error("NewCharacter IsSleeping: got true, want false")
	}
	if c.Poisoned {
		t.Error("NewCharacter Poisoned: got true, want false")
	}
	if c.IsFrustrated {
		t.Error("NewCharacter IsFrustrated: got true, want false")
	}
}

// TestNewCharacter_SymbolAndType verifies symbol and type are set correctly
func TestNewCharacter_SymbolAndType(t *testing.T) {
	t.Parallel()

	c := NewCharacter(1, 0, 0, "Test", "berry", types.ColorRed)
	if c.Symbol() != config.CharRobot {
		t.Errorf("NewCharacter Symbol(): got %q, want %q", c.Symbol(), config.CharRobot)
	}
	if c.Type() != TypeCharacter {
		t.Errorf("NewCharacter Type(): got %d, want %d", c.Type(), TypeCharacter)
	}
}

// TestNetPreference_NoPreferences verifies empty preferences returns 0
func TestNetPreference_NoPreferences(t *testing.T) {
	t.Parallel()

	c := &Character{Preferences: []Preference{}}
	item := NewBerry(0, 0, types.ColorRed, false, false)
	got := c.NetPreference(item)
	if got != 0 {
		t.Errorf("NetPreference() with no preferences: got %d, want 0", got)
	}
}

// TestNetPreference_SingleMatch verifies single matching preference returns valence
func TestNetPreference_SingleMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref     Preference
		item     *Item
		expected int
	}{
		{
			name:     "positive itemType match",
			pref:     NewPositivePreference("berry", ""),
			item:     NewBerry(0, 0, types.ColorBlue, false, false),
			expected: 1,
		},
		{
			name:     "positive color match",
			pref:     NewPositivePreference("", types.ColorRed),
			item:     NewMushroom(0, 0, types.ColorRed, types.PatternNone, types.TextureNone, false, false),
			expected: 1,
		},
		{
			name:     "negative itemType match",
			pref:     NewNegativePreference("berry", ""),
			item:     NewBerry(0, 0, types.ColorBlue, false, false),
			expected: -1,
		},
		{
			name:     "negative color match",
			pref:     NewNegativePreference("", types.ColorRed),
			item:     NewMushroom(0, 0, types.ColorRed, types.PatternNone, types.TextureNone, false, false),
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Preferences: []Preference{tt.pref}}
			got := c.NetPreference(tt.item)
			if got != tt.expected {
				t.Errorf("NetPreference(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

// TestNetPreference_NoMatch verifies non-matching preferences return 0
func TestNetPreference_NoMatch(t *testing.T) {
	t.Parallel()

	c := &Character{
		Preferences: []Preference{
			NewPositivePreference("berry", ""),       // likes berries
			NewPositivePreference("", types.ColorRed), // likes red
		},
	}
	// Item is white mushroom - matches neither preference
	item := NewMushroom(0, 0, types.ColorWhite, types.PatternNone, types.TextureNone, false, false)
	got := c.NetPreference(item)
	if got != 0 {
		t.Errorf("NetPreference() with no matching preferences: got %d, want 0", got)
	}
}

// TestNetPreference_MultipleMatches verifies multiple matches sum correctly
func TestNetPreference_MultipleMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefs    []Preference
		item     *Item
		expected int
	}{
		{
			name: "two positive matches (perfect)",
			prefs: []Preference{
				NewPositivePreference("berry", ""),
				NewPositivePreference("", types.ColorRed),
			},
			item:     NewBerry(0, 0, types.ColorRed, false, false), // matches both
			expected: 2,
		},
		{
			name: "positive and negative cancel",
			prefs: []Preference{
				NewPositivePreference("berry", ""),        // likes berries
				NewNegativePreference("", types.ColorRed), // dislikes red
			},
			item:     NewBerry(0, 0, types.ColorRed, false, false), // matches both, opposite valence
			expected: 0,
		},
		{
			name: "two negatives stack",
			prefs: []Preference{
				NewNegativePreference("berry", ""),
				NewNegativePreference("", types.ColorRed),
			},
			item:     NewBerry(0, 0, types.ColorRed, false, false),
			expected: -2,
		},
		{
			name: "partial match only",
			prefs: []Preference{
				NewPositivePreference("berry", ""),         // matches
				NewPositivePreference("", types.ColorBlue), // doesn't match
			},
			item:     NewBerry(0, 0, types.ColorRed, false, false),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Preferences: tt.prefs}
			got := c.NetPreference(tt.item)
			if got != tt.expected {
				t.Errorf("NetPreference(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

// TestNetPreference_ComboPreference verifies combo preferences match correctly
// Combo preferences contribute Valence × 2 (since they specify 2 attributes)
func TestNetPreference_ComboPreference(t *testing.T) {
	t.Parallel()

	// Character likes specifically red berries (combo)
	c := &Character{
		Preferences: []Preference{
			NewPositivePreference("berry", types.ColorRed), // likes red berries
		},
	}

	tests := []struct {
		name     string
		item     *Item
		expected int
	}{
		{
			name:     "exact combo match",
			item:     NewBerry(0, 0, types.ColorRed, false, false),
			expected: 2, // valence(+1) × attrCount(2)
		},
		{
			name:     "wrong color for combo",
			item:     NewBerry(0, 0, types.ColorBlue, false, false),
			expected: 0,
		},
		{
			name:     "wrong type for combo",
			item:     NewMushroom(0, 0, types.ColorRed, types.PatternNone, types.TextureNone, false, false),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.NetPreference(tt.item)
			if got != tt.expected {
				t.Errorf("NetPreference(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

// TestHasKnowledge verifies HasKnowledge detects existing knowledge
func TestHasKnowledge(t *testing.T) {
	t.Parallel()

	k1 := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}

	k2 := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}

	c := &Character{
		Knowledge: []Knowledge{k1},
	}

	if !c.HasKnowledge(k1) {
		t.Error("HasKnowledge() should return true for existing knowledge")
	}

	if c.HasKnowledge(k2) {
		t.Error("HasKnowledge() should return false for unknown knowledge")
	}
}

// TestLearnKnowledge_NewKnowledge verifies learning new knowledge
func TestLearnKnowledge_NewKnowledge(t *testing.T) {
	t.Parallel()

	c := &Character{Knowledge: []Knowledge{}}

	k := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
	}

	learned := c.LearnKnowledge(k)

	if !learned {
		t.Error("LearnKnowledge() should return true when learning new knowledge")
	}

	if len(c.Knowledge) != 1 {
		t.Errorf("LearnKnowledge() should add knowledge: got %d items, want 1", len(c.Knowledge))
	}

	if !c.HasKnowledge(k) {
		t.Error("LearnKnowledge() should add knowledge that can be found with HasKnowledge")
	}
}

// TestLearnKnowledge_AlreadyKnown verifies duplicate knowledge is not added
func TestLearnKnowledge_AlreadyKnown(t *testing.T) {
	t.Parallel()

	k := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
	}

	c := &Character{Knowledge: []Knowledge{k}}

	learned := c.LearnKnowledge(k)

	if learned {
		t.Error("LearnKnowledge() should return false when knowledge already exists")
	}

	if len(c.Knowledge) != 1 {
		t.Errorf("LearnKnowledge() should not duplicate: got %d items, want 1", len(c.Knowledge))
	}
}

// TestLearnKnowledge_MultipleKnowledge verifies learning multiple pieces of knowledge
func TestLearnKnowledge_MultipleKnowledge(t *testing.T) {
	t.Parallel()

	c := &Character{Knowledge: []Knowledge{}}

	k1 := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
	}

	k2 := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}

	c.LearnKnowledge(k1)
	c.LearnKnowledge(k2)

	if len(c.Knowledge) != 2 {
		t.Errorf("Character should have 2 knowledge items, got %d", len(c.Knowledge))
	}

	if !c.HasKnowledge(k1) || !c.HasKnowledge(k2) {
		t.Error("Character should have both knowledge items")
	}
}

// =============================================================================
// KnownHealingItems (E1: Filter items by healing knowledge)
// =============================================================================

func TestKnownHealingItems_NoKnowledge(t *testing.T) {
	t.Parallel()

	c := &Character{Knowledge: []Knowledge{}}

	items := []*Item{
		NewBerry(0, 0, types.ColorBlue, false, true),  // healing
		NewBerry(1, 0, types.ColorRed, false, false),  // not healing
	}

	result := c.KnownHealingItems(items)

	if len(result) != 0 {
		t.Errorf("KnownHealingItems() with no knowledge: got %d items, want 0", len(result))
	}
}

func TestKnownHealingItems_HasHealingKnowledge(t *testing.T) {
	t.Parallel()

	// Character knows blue berries are healing
	healingKnowledge := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	c := &Character{Knowledge: []Knowledge{healingKnowledge}}

	blueBerry := NewBerry(0, 0, types.ColorBlue, false, true)
	redBerry := NewBerry(1, 0, types.ColorRed, false, false)
	items := []*Item{blueBerry, redBerry}

	result := c.KnownHealingItems(items)

	if len(result) != 1 {
		t.Fatalf("KnownHealingItems(): got %d items, want 1", len(result))
	}
	if result[0] != blueBerry {
		t.Error("KnownHealingItems() should return the blue berry")
	}
}

func TestKnownHealingItems_OnlyPoisonKnowledge(t *testing.T) {
	t.Parallel()

	// Character only knows about poison, not healing
	poisonKnowledge := Knowledge{
		Category: KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
	}
	c := &Character{Knowledge: []Knowledge{poisonKnowledge}}

	items := []*Item{
		NewBerry(0, 0, types.ColorBlue, false, true),     // healing but unknown
		NewMushroom(1, 0, types.ColorRed, types.PatternNone, types.TextureNone, true, false), // known poison
	}

	result := c.KnownHealingItems(items)

	if len(result) != 0 {
		t.Errorf("KnownHealingItems() with only poison knowledge: got %d items, want 0", len(result))
	}
}

func TestKnownHealingItems_MultipleHealingKnowledge(t *testing.T) {
	t.Parallel()

	// Character knows two types of healing items
	k1 := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	k2 := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "mushroom",
		Color:    types.ColorWhite,
	}
	c := &Character{Knowledge: []Knowledge{k1, k2}}

	blueBerry := NewBerry(0, 0, types.ColorBlue, false, true)
	whiteMushroom := NewMushroom(1, 0, types.ColorWhite, types.PatternNone, types.TextureNone, false, true)
	redBerry := NewBerry(2, 0, types.ColorRed, false, false) // not known healing
	items := []*Item{blueBerry, whiteMushroom, redBerry}

	result := c.KnownHealingItems(items)

	if len(result) != 2 {
		t.Fatalf("KnownHealingItems(): got %d items, want 2", len(result))
	}
}

func TestKnownHealingItems_EmptyItemList(t *testing.T) {
	t.Parallel()

	healingKnowledge := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	c := &Character{Knowledge: []Knowledge{healingKnowledge}}

	result := c.KnownHealingItems([]*Item{})

	if len(result) != 0 {
		t.Errorf("KnownHealingItems() with empty item list: got %d items, want 0", len(result))
	}
}

func TestKnownHealingItems_KnowledgeMustMatchExactly(t *testing.T) {
	t.Parallel()

	// Character knows spotted red mushrooms are healing
	healingKnowledge := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}
	c := &Character{Knowledge: []Knowledge{healingKnowledge}}

	spottedRed := NewMushroom(0, 0, types.ColorRed, types.PatternSpotted, types.TextureNone, false, true)
	plainRed := NewMushroom(1, 0, types.ColorRed, types.PatternNone, types.TextureNone, false, true) // different pattern
	items := []*Item{spottedRed, plainRed}

	result := c.KnownHealingItems(items)

	if len(result) != 1 {
		t.Fatalf("KnownHealingItems(): got %d items, want 1 (exact match only)", len(result))
	}
	if result[0] != spottedRed {
		t.Error("KnownHealingItems() should only return exactly matching item")
	}
}

// =============================================================================
// Variety-based Methods (for vessel contents)
// =============================================================================

// TestNetPreferenceForVariety_NoPreferences verifies empty preferences returns 0
func TestNetPreferenceForVariety_NoPreferences(t *testing.T) {
	t.Parallel()

	c := &Character{Preferences: []Preference{}}
	variety := &ItemVariety{ItemType: "berry", Color: types.ColorRed}

	got := c.NetPreferenceForVariety(variety)
	if got != 0 {
		t.Errorf("NetPreferenceForVariety() with no preferences: got %d, want 0", got)
	}
}

// TestNetPreferenceForVariety_SingleMatch verifies single matching preference
func TestNetPreferenceForVariety_SingleMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pref     Preference
		variety  *ItemVariety
		expected int
	}{
		{
			name:     "positive itemType match",
			pref:     NewPositivePreference("berry", ""),
			variety:  &ItemVariety{ItemType: "berry", Color: types.ColorRed},
			expected: 1,
		},
		{
			name:     "positive color match",
			pref:     NewPositivePreference("", types.ColorRed),
			variety:  &ItemVariety{ItemType: "berry", Color: types.ColorRed},
			expected: 1,
		},
		{
			name:     "negative itemType match",
			pref:     NewNegativePreference("berry", ""),
			variety:  &ItemVariety{ItemType: "berry", Color: types.ColorRed},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{Preferences: []Preference{tt.pref}}
			got := c.NetPreferenceForVariety(tt.variety)
			if got != tt.expected {
				t.Errorf("NetPreferenceForVariety(): got %d, want %d", got, tt.expected)
			}
		})
	}
}

// TestNetPreferenceForVariety_MultipleMatches verifies multiple matches sum correctly
func TestNetPreferenceForVariety_MultipleMatches(t *testing.T) {
	t.Parallel()

	// Likes berries (+1) and likes red (+1), so red berry = +2
	c := &Character{
		Preferences: []Preference{
			NewPositivePreference("berry", ""),
			NewPositivePreference("", types.ColorRed),
		},
	}
	redBerry := &ItemVariety{ItemType: "berry", Color: types.ColorRed}

	got := c.NetPreferenceForVariety(redBerry)
	if got != 2 {
		t.Errorf("NetPreferenceForVariety() with two matches: got %d, want 2", got)
	}

	// Blue berry only matches berry preference = +1
	blueBerry := &ItemVariety{ItemType: "berry", Color: types.ColorBlue}
	got = c.NetPreferenceForVariety(blueBerry)
	if got != 1 {
		t.Errorf("NetPreferenceForVariety() with one match: got %d, want 1", got)
	}
}

// TestNetPreferenceForVariety_MushroomWithPatternTexture verifies complex variety matching
func TestNetPreferenceForVariety_MushroomWithPatternTexture(t *testing.T) {
	t.Parallel()

	// Dislikes spotted things and slimy things
	c := &Character{
		Preferences: []Preference{
			{Valence: -1, Pattern: types.PatternSpotted},
			{Valence: -1, Texture: types.TextureSlimy},
		},
	}

	slimySpotted := &ItemVariety{
		ItemType: "mushroom",
		Color:    types.ColorBrown,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
	}

	got := c.NetPreferenceForVariety(slimySpotted)
	if got != -2 {
		t.Errorf("NetPreferenceForVariety() for slimy spotted: got %d, want -2", got)
	}
}

// TestKnowsVarietyIsHealing verifies healing knowledge check for varieties
func TestKnowsVarietyIsHealing(t *testing.T) {
	t.Parallel()

	healingKnowledge := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	c := &Character{Knowledge: []Knowledge{healingKnowledge}}

	blueVariety := &ItemVariety{ItemType: "berry", Color: types.ColorBlue}
	redVariety := &ItemVariety{ItemType: "berry", Color: types.ColorRed}

	if !c.KnowsVarietyIsHealing(blueVariety) {
		t.Error("KnowsVarietyIsHealing() should return true for known healing variety")
	}
	if c.KnowsVarietyIsHealing(redVariety) {
		t.Error("KnowsVarietyIsHealing() should return false for unknown variety")
	}
}

// TestKnowsVarietyIsHealing_NoKnowledge verifies false when no knowledge
func TestKnowsVarietyIsHealing_NoKnowledge(t *testing.T) {
	t.Parallel()

	c := &Character{Knowledge: []Knowledge{}}
	variety := &ItemVariety{ItemType: "berry", Color: types.ColorBlue, Edible: &EdibleProperties{Healing: true}}

	if c.KnowsVarietyIsHealing(variety) {
		t.Error("KnowsVarietyIsHealing() should return false when no knowledge")
	}
}

// TestKnowsVarietyIsHealing_MustMatchExactly verifies exact variety matching
func TestKnowsVarietyIsHealing_MustMatchExactly(t *testing.T) {
	t.Parallel()

	// Knows spotted red mushrooms are healing
	healingKnowledge := Knowledge{
		Category: KnowledgeHealing,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}
	c := &Character{Knowledge: []Knowledge{healingKnowledge}}

	spottedRed := &ItemVariety{
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
	}
	plainRed := &ItemVariety{
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternNone,
	}

	if !c.KnowsVarietyIsHealing(spottedRed) {
		t.Error("Should know spotted red mushroom is healing")
	}
	if c.KnowsVarietyIsHealing(plainRed) {
		t.Error("Should not know plain red mushroom is healing (different pattern)")
	}
}
