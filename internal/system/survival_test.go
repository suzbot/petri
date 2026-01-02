package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

// Helper to create a basic test character
func newTestCharacter() *entity.Character {
	return entity.NewCharacter(1, 0, 0, "Test", "berry", types.ColorRed)
}

// =============================================================================
// Stat Changes Over Time
// =============================================================================

func TestUpdateSurvival_HungerIncreasesOverTime(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 50
	char.HungerCooldown = 0

	UpdateSurvival(char, 1.0, nil)

	expected := 50 + config.HungerIncreaseRate
	if char.Hunger != expected {
		t.Errorf("Hunger: got %.2f, want %.2f", char.Hunger, expected)
	}
}

func TestUpdateSurvival_HungerDoesNotIncreaseOnCooldown(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 0
	char.HungerCooldown = 5.0

	UpdateSurvival(char, 1.0, nil)

	if char.Hunger != 0 {
		t.Errorf("Hunger should remain 0 on cooldown, got %.2f", char.Hunger)
	}
	if char.HungerCooldown != 4.0 {
		t.Errorf("HungerCooldown: got %.2f, want 4.0", char.HungerCooldown)
	}
}

func TestUpdateSurvival_HungerCapsAt100(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 99.5
	char.HungerCooldown = 0

	UpdateSurvival(char, 10.0, nil) // Enough to exceed 100

	if char.Hunger != 100 {
		t.Errorf("Hunger should cap at 100, got %.2f", char.Hunger)
	}
}

func TestUpdateSurvival_ThirstIncreasesOverTime(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 50
	char.ThirstCooldown = 0

	UpdateSurvival(char, 1.0, nil)

	expected := 50 + config.ThirstIncreaseRate
	if char.Thirst != expected {
		t.Errorf("Thirst: got %.2f, want %.2f", char.Thirst, expected)
	}
}

func TestUpdateSurvival_ThirstDoesNotIncreaseOnCooldown(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 0
	char.ThirstCooldown = 5.0

	UpdateSurvival(char, 1.0, nil)

	if char.Thirst != 0 {
		t.Errorf("Thirst should remain 0 on cooldown, got %.2f", char.Thirst)
	}
}

func TestUpdateSurvival_EnergyDecreasesWhenAwake(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 100
	char.EnergyCooldown = 0
	char.IsSleeping = false

	UpdateSurvival(char, 1.0, nil)

	expected := 100 - config.EnergyDecreaseRate
	if char.Energy != expected {
		t.Errorf("Energy: got %.2f, want %.2f", char.Energy, expected)
	}
}

func TestUpdateSurvival_EnergyDoesNotDecreaseOnCooldown(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 100
	char.EnergyCooldown = 5.0
	char.IsSleeping = false

	UpdateSurvival(char, 1.0, nil)

	if char.Energy != 100 {
		t.Errorf("Energy should remain 100 on cooldown, got %.2f", char.Energy)
	}
}

func TestUpdateSurvival_EnergyFloorsAt0(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 1
	char.EnergyCooldown = 0
	char.IsSleeping = false

	UpdateSurvival(char, 10.0, nil) // Enough to go below 0

	if char.Energy != 0 {
		t.Errorf("Energy should floor at 0, got %.2f", char.Energy)
	}
}

// =============================================================================
// Sleep and Wake Mechanics
// =============================================================================

func TestUpdateSurvival_EnergyRestoresFasterInBed(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 50
	char.IsSleeping = true
	char.AtBed = true

	UpdateSurvival(char, 1.0, nil)

	expected := 50 + config.BedEnergyRestoreRate
	if char.Energy != expected {
		t.Errorf("Energy in bed: got %.2f, want %.2f", char.Energy, expected)
	}
}

func TestUpdateSurvival_EnergyRestoresSlowerOnGround(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 50
	char.IsSleeping = true
	char.AtBed = false

	UpdateSurvival(char, 1.0, nil)

	expected := 50 + config.GroundEnergyRestoreRate
	if char.Energy != expected {
		t.Errorf("Energy on ground: got %.2f, want %.2f", char.Energy, expected)
	}
}

func TestUpdateSurvival_WakesAt100InBed(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 99
	char.IsSleeping = true
	char.AtBed = true

	UpdateSurvival(char, 1.0, nil) // BedEnergyRestoreRate should push past 100

	if char.IsSleeping {
		t.Error("Character should wake up at 100 energy in bed")
	}
	if char.EnergyCooldown != config.SatisfactionCooldown {
		t.Errorf("EnergyCooldown: got %.2f, want %.2f", char.EnergyCooldown, config.SatisfactionCooldown)
	}
}

func TestUpdateSurvival_WakesAt75OnGround(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 74
	char.IsSleeping = true
	char.AtBed = false

	UpdateSurvival(char, 1.0, nil) // GroundEnergyRestoreRate should push past 75

	if char.IsSleeping {
		t.Error("Character should wake up at 75 energy on ground")
	}
	if char.AtBed {
		t.Error("AtBed should be false after waking")
	}
}

func TestUpdateSurvival_MoodBoostWhenFullyRested(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 99
	char.Mood = 50
	char.IsSleeping = true
	char.AtBed = true

	UpdateSurvival(char, 1.0, nil) // BedEnergyRestoreRate should push past 100

	expected := 50 + config.MoodBoostOnConsumption
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateSurvival_NoMoodBoostWhenPartiallyRested(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 74
	char.Mood = 50
	char.IsSleeping = true
	char.AtBed = false

	UpdateSurvival(char, 1.0, nil) // GroundEnergyRestoreRate should push past 75

	if char.Mood != 50 {
		t.Errorf("Mood should remain 50 when partially rested, got %.2f", char.Mood)
	}
}

func TestUpdateSurvival_WakesEarlyDueToHunger(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 50 // EnergyUrgency = 50
	char.Hunger = 80 // Moderate tier, HungerUrgency = 80 > 50
	char.IsSleeping = true
	char.AtBed = true

	UpdateSurvival(char, 0.01, nil) // Small delta to not change energy much

	if char.IsSleeping {
		t.Error("Character should wake early due to hunger at Moderate+ tier")
	}
}

func TestUpdateSurvival_WakesEarlyDueToThirst(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 50 // EnergyUrgency = 50
	char.Thirst = 80 // Moderate tier, ThirstUrgency = 80 > 50
	char.IsSleeping = true
	char.AtBed = true

	UpdateSurvival(char, 0.01, nil)

	if char.IsSleeping {
		t.Error("Character should wake early due to thirst at Moderate+ tier")
	}
}

func TestUpdateSurvival_DoesNotWakeEarlyForMildHunger(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 50 // EnergyUrgency = 50
	char.Hunger = 60 // Mild tier (50-74)
	char.IsSleeping = true
	char.AtBed = true

	UpdateSurvival(char, 0.01, nil)

	if !char.IsSleeping {
		t.Error("Character should not wake early for Mild hunger")
	}
}

// =============================================================================
// Damage and Death
// =============================================================================

func TestUpdateSurvival_PoisonDealsDamage(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 100
	char.Poisoned = true
	char.PoisonTimer = 10.0

	UpdateSurvival(char, 1.0, nil)

	expected := 100 - config.PoisonDamageRate
	if char.Health != expected {
		t.Errorf("Health after poison: got %.2f, want %.2f", char.Health, expected)
	}
}

func TestUpdateSurvival_PoisonWearsOff(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Poisoned = true
	char.PoisonTimer = 1.0

	UpdateSurvival(char, 1.5, nil)

	if char.Poisoned {
		t.Error("Poison should wear off when timer expires")
	}
	if char.PoisonTimer != 0 {
		t.Errorf("PoisonTimer: got %.2f, want 0", char.PoisonTimer)
	}
}

func TestUpdateSurvival_StarvationDealsDamage(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 100
	char.Health = 100

	UpdateSurvival(char, 1.0, nil)

	expected := 100 - config.StarvationDamageRate
	if char.Health != expected {
		t.Errorf("Health after starvation: got %.2f, want %.2f", char.Health, expected)
	}
}

func TestUpdateSurvival_NoStarvationDamageBelow100Hunger(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 99
	char.Health = 100

	UpdateSurvival(char, 1.0, nil)

	if char.Health != 100 {
		t.Errorf("Health should remain 100 when hunger < 100, got %.2f", char.Health)
	}
}

func TestUpdateSurvival_DehydrationDealsDamage(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 100
	char.Health = 100

	UpdateSurvival(char, 1.0, nil)

	expected := 100 - config.DehydrationDamageRate
	if char.Health != expected {
		t.Errorf("Health after dehydration: got %.2f, want %.2f", char.Health, expected)
	}
}

func TestUpdateSurvival_DeathAtZeroHealth(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 0.1
	char.Hunger = 100 // Taking starvation damage
	char.IsSleeping = true

	UpdateSurvival(char, 1.0, nil)

	if !char.IsDead {
		t.Error("Character should be dead when health reaches 0")
	}
	if char.Health != 0 {
		t.Errorf("Health should be exactly 0, got %.2f", char.Health)
	}
	if char.IsSleeping {
		t.Error("Dead character should not be sleeping")
	}
}

func TestUpdateSurvival_DeadCharacterNotUpdated(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsDead = true
	char.Hunger = 50
	char.Thirst = 50

	UpdateSurvival(char, 1.0, nil)

	if char.Hunger != 50 {
		t.Errorf("Dead character hunger should not change, got %.2f", char.Hunger)
	}
	if char.Thirst != 50 {
		t.Errorf("Dead character thirst should not change, got %.2f", char.Thirst)
	}
}

// =============================================================================
// Frustration Timer
// =============================================================================

func TestUpdateSurvival_FrustrationTimerDecrements(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsFrustrated = true
	char.FrustrationTimer = 5.0

	UpdateSurvival(char, 1.0, nil)

	if char.FrustrationTimer != 4.0 {
		t.Errorf("FrustrationTimer: got %.2f, want 4.0", char.FrustrationTimer)
	}
}

func TestUpdateSurvival_FrustrationClearsWhenTimerExpires(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsFrustrated = true
	char.FrustrationTimer = 0.5

	UpdateSurvival(char, 1.0, nil)

	if char.IsFrustrated {
		t.Error("Frustration should clear when timer expires")
	}
	if char.FrustrationTimer != 0 {
		t.Errorf("FrustrationTimer: got %.2f, want 0", char.FrustrationTimer)
	}
}

// =============================================================================
// Regression Tests
// =============================================================================

// TestRegression_EnergyMilestonesAllLogged verifies that when energy drops across
// multiple thresholds in one tick, all milestones are logged (not just the first).
// Bug: else-if chain caused only first threshold crossing to be logged.
func TestRegression_EnergyMilestonesAllLogged(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Energy = 60 // Above all thresholds
	char.EnergyCooldown = 0
	char.IsSleeping = false

	log := NewActionLog(100)

	// Drop energy across multiple thresholds in one tick
	// We'll manually set prevEnergy high and current energy low to simulate
	// a large drop (e.g., from 60 to 5 in one tick)
	// Since UpdateSurvival captures prevEnergy internally, we need to use
	// a very large delta to cross multiple thresholds

	// Energy decreases at 0.5/sec, so delta of 110 would drop 55 points
	// From 60 to 5, crossing 50, 25, and 10 thresholds
	UpdateSurvival(char, 110.0, log)

	// Check that energy dropped significantly
	if char.Energy > 10 {
		t.Fatalf("Energy should have dropped below 10, got %.2f", char.Energy)
	}

	// Check that multiple milestones were logged
	events := log.AllEvents(100)
	energyMessages := []string{}
	for _, e := range events {
		if e.Type == "energy" {
			energyMessages = append(energyMessages, e.Message)
		}
	}

	// Should have logged: "Getting tired" (50), "Very tired!" (25), "Exhausted!" (10)
	expectedCount := 3
	if len(energyMessages) < expectedCount {
		t.Errorf("Expected at least %d energy milestones, got %d: %v",
			expectedCount, len(energyMessages), energyMessages)
	}
}

// =============================================================================
// Mood Updates
// =============================================================================

func TestUpdateMood_IncreasesWhenAllNeedsMet(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 0  // TierNone
	char.Thirst = 0  // TierNone
	char.Energy = 100 // TierNone
	char.Health = 100 // TierNone

	UpdateMood(char, 1.0, nil)

	expected := 50 + config.MoodIncreaseRate
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateMood_NoChangeAtMildTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 60 // TierMild (50-74)
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100

	UpdateMood(char, 1.0, nil)

	if char.Mood != 50 {
		t.Errorf("Mood should remain 50 at TierMild, got %.2f", char.Mood)
	}
}

func TestUpdateMood_DecreasesSlowlyAtModerateTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 80 // TierModerate (75-89)
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100

	UpdateMood(char, 1.0, nil)

	expected := 50 - config.MoodDecreaseRateSlow
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateMood_DecreasesMediumAtSevereTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 95 // TierSevere (90-99)
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100

	UpdateMood(char, 1.0, nil)

	expected := 50 - config.MoodDecreaseRateMedium
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateMood_DecreasesFastAtCrisisTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 100 // TierCrisis
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100

	UpdateMood(char, 1.0, nil)

	expected := 50 - config.MoodDecreaseRateFast
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateMood_CapsAt100(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 99.9
	char.Hunger = 0
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100

	UpdateMood(char, 10.0, nil) // Enough to exceed 100

	if char.Mood != 100 {
		t.Errorf("Mood should cap at 100, got %.2f", char.Mood)
	}
}

func TestUpdateMood_FloorsAt0(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 1
	char.Hunger = 100 // TierCrisis
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100

	UpdateMood(char, 10.0, nil) // Enough to go below 0

	if char.Mood != 0 {
		t.Errorf("Mood should floor at 0, got %.2f", char.Mood)
	}
}

func TestUpdateMood_DeadCharacterNotUpdated(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsDead = true
	char.Mood = 50
	char.Hunger = 100

	UpdateMood(char, 1.0, nil)

	if char.Mood != 50 {
		t.Errorf("Dead character mood should not change, got %.2f", char.Mood)
	}
}

func TestUpdateMood_UsesHighestNeedTier(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 60  // TierMild
	char.Thirst = 80  // TierModerate
	char.Energy = 100 // TierNone
	char.Health = 100 // TierNone

	UpdateMood(char, 1.0, nil)

	// Should use TierModerate (from thirst) → slow decrease
	expected := 50 - config.MoodDecreaseRateSlow
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f (should use highest tier)", char.Mood, expected)
	}
}

func TestUpdateMood_IncludesHealthInCalculation(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 0   // TierNone
	char.Thirst = 0   // TierNone
	char.Energy = 100 // TierNone
	char.Health = 20  // TierSevere (≤25)

	UpdateMood(char, 1.0, nil)

	// Should use TierSevere (from health) → medium decrease
	expected := 50 - config.MoodDecreaseRateMedium
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f (should include Health)", char.Mood, expected)
	}
}

func TestUpdateMood_LogsTierTransitions(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	// Mood threshold: TierMild > 64, TierModerate ≤ 64
	// Start just above threshold so small decrease crosses it
	char.Mood = 64.2 // Just above Neutral threshold
	char.Hunger = 80 // TierModerate → slow decrease (0.3/sec)

	log := NewActionLog(100)
	UpdateMood(char, 1.0, log)

	// After decrease of 0.3, mood should cross from Happy (>64) to Neutral (≤64)
	if char.Mood > 64 {
		t.Fatalf("Mood should have dropped to ≤64, got %.2f", char.Mood)
	}

	events := log.AllEvents(100)
	found := false
	for _, e := range events {
		if e.Type == "mood" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected mood tier transition to be logged")
	}
}

func TestUpdateMood_PoisonedPenaltyApplied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 0   // TierNone
	char.Thirst = 0   // TierNone
	char.Energy = 100 // TierNone
	char.Health = 100 // TierNone
	char.Poisoned = true

	UpdateMood(char, 1.0, nil)

	// Should get increase from TierNone, minus poison penalty
	expected := 50 + config.MoodIncreaseRate - config.MoodPenaltyPoisoned
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateMood_FrustratedPenaltyApplied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 0   // TierNone
	char.Thirst = 0   // TierNone
	char.Energy = 100 // TierNone
	char.Health = 100 // TierNone
	char.IsFrustrated = true

	UpdateMood(char, 1.0, nil)

	// Should get increase from TierNone, minus frustration penalty
	expected := 50 + config.MoodIncreaseRate - config.MoodPenaltyFrustrated
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestUpdateMood_StatusPenaltiesStack(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Hunger = 80  // TierModerate
	char.Thirst = 0
	char.Energy = 100
	char.Health = 100
	char.Poisoned = true
	char.IsFrustrated = true

	UpdateMood(char, 1.0, nil)

	// Should get: base decay (moderate) + poison penalty + frustration penalty
	expected := 50 - config.MoodDecreaseRateSlow - config.MoodPenaltyPoisoned - config.MoodPenaltyFrustrated
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}
