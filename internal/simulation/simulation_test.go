package simulation

import (
	"testing"

	"petri/internal/config"
	"petri/internal/system"
)

// Standard delta for simulation ticks
var tickDelta = config.UpdateInterval.Seconds()

// =============================================================================
// Helper Assertions
// =============================================================================

// assertNoPositionDuplicates verifies no two characters share a position
func assertNoPositionDuplicates(t *testing.T, world *TestWorld) {
	t.Helper()
	chars := world.GameMap.Characters()
	positions := make(map[[2]int]int) // position -> character ID

	for _, char := range chars {
		x, y := char.Position()
		key := [2]int{x, y}
		if existingID, exists := positions[key]; exists {
			t.Errorf("Position conflict: characters %d and %d both at (%d, %d)",
				existingID, char.ID, x, y)
		}
		positions[key] = char.ID
	}
}

// assertCharacterMapConsistency verifies CharacterAt matches actual positions
func assertCharacterMapConsistency(t *testing.T, world *TestWorld) {
	t.Helper()
	chars := world.GameMap.Characters()

	for _, char := range chars {
		x, y := char.Position()
		found := world.GameMap.CharacterAt(x, y)
		if found != char {
			if found == nil {
				t.Errorf("CharacterAt(%d, %d) returned nil, expected character %d",
					x, y, char.ID)
			} else {
				t.Errorf("CharacterAt(%d, %d) returned character %d, expected %d",
					x, y, found.ID, char.ID)
			}
		}
	}
}

// countAlive returns the number of living characters
func countAlive(world *TestWorld) int {
	count := 0
	for _, char := range world.GameMap.Characters() {
		if !char.IsDead {
			count++
		}
	}
	return count
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestSimulation_NoCharacterDuplication(t *testing.T) {
	t.Parallel()

	world := CreateTestWorld(WorldOptions{})
	initialCount := len(world.GameMap.Characters())

	for tick := 0; tick < 1000; tick++ {
		RunTick(world, tickDelta)

		// Check no position duplicates
		assertNoPositionDuplicates(t, world)

		// Check character count (should only decrease due to deaths)
		currentCount := len(world.GameMap.Characters())
		aliveCount := countAlive(world)
		if currentCount != initialCount {
			t.Errorf("Tick %d: Character count changed from %d to %d (unexpected)",
				tick, initialCount, currentCount)
		}

		// Alive count should only decrease
		if aliveCount > initialCount {
			t.Errorf("Tick %d: Alive count %d exceeds initial %d",
				tick, aliveCount, initialCount)
		}
	}
}

func TestSimulation_PositionMapConsistency(t *testing.T) {
	t.Parallel()

	world := CreateTestWorld(WorldOptions{})

	for tick := 0; tick < 1000; tick++ {
		RunTick(world, tickDelta)

		// Verify map consistency after each tick
		assertCharacterMapConsistency(t, world)
	}
}

func TestSimulation_NoPanic(t *testing.T) {
	t.Parallel()

	world := CreateTestWorld(WorldOptions{})

	// This test passes if it completes without panic
	// The defer/recover pattern catches any panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Simulation panicked: %v", r)
		}
	}()

	RunTicks(world, 1000, tickDelta)

	// Verify simulation completed
	chars := world.GameMap.Characters()
	if len(chars) == 0 {
		t.Error("No characters found after simulation")
	}
}

func TestSimulation_DeadCharactersStopUpdating(t *testing.T) {
	t.Parallel()

	// Create world without water - characters will eventually die
	world := CreateTestWorld(WorldOptions{NoWater: true})

	// Run until at least one character dies (or max ticks)
	var deadChar *struct {
		ID       int
		X, Y     int
		Hunger   float64
		Thirst   float64
		Energy   float64
		DeathTick int
	}

	maxTicks := 2000
	for tick := 0; tick < maxTicks; tick++ {
		RunTick(world, tickDelta)

		// Look for newly dead character
		if deadChar == nil {
			for _, char := range world.GameMap.Characters() {
				if char.IsDead {
					x, y := char.Position()
					deadChar = &struct {
						ID        int
						X, Y      int
						Hunger    float64
						Thirst    float64
						Energy    float64
						DeathTick int
					}{
						ID:        char.ID,
						X:         x,
						Y:         y,
						Hunger:    char.Hunger,
						Thirst:    char.Thirst,
						Energy:    char.Energy,
						DeathTick: tick,
					}
					break
				}
			}
		}

		// Once we have a dead character, verify it stops updating
		if deadChar != nil && tick > deadChar.DeathTick+10 {
			// Find the character again
			for _, char := range world.GameMap.Characters() {
				if char.ID == deadChar.ID {
					x, y := char.Position()

					// Position should not change
					if x != deadChar.X || y != deadChar.Y {
						t.Errorf("Dead character %d moved from (%d,%d) to (%d,%d)",
							deadChar.ID, deadChar.X, deadChar.Y, x, y)
					}

					// Stats should not change (they were recorded at death)
					if char.Hunger != deadChar.Hunger {
						t.Errorf("Dead character %d hunger changed from %.2f to %.2f",
							deadChar.ID, deadChar.Hunger, char.Hunger)
					}
					if char.Thirst != deadChar.Thirst {
						t.Errorf("Dead character %d thirst changed from %.2f to %.2f",
							deadChar.ID, deadChar.Thirst, char.Thirst)
					}
					if char.Energy != deadChar.Energy {
						t.Errorf("Dead character %d energy changed from %.2f to %.2f",
							deadChar.ID, deadChar.Energy, char.Energy)
					}
					break
				}
			}
			// Test complete - we verified the dead character doesn't update
			return
		}
	}

	if deadChar == nil {
		t.Skip("No character died within max ticks - cannot verify dead character behavior")
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestSimulation_AllCharactersAtSameStart(t *testing.T) {
	t.Parallel()

	// Create world with characters that will need to spread out
	world := CreateTestWorld(WorldOptions{NumCharacters: 4})

	// Run a few ticks
	RunTicks(world, 100, tickDelta)

	// Verify no duplicates even after movement
	assertNoPositionDuplicates(t, world)
	assertCharacterMapConsistency(t, world)
}

func TestSimulation_NoResources(t *testing.T) {
	t.Parallel()

	// Create world with no resources at all
	world := CreateTestWorld(WorldOptions{
		NoWater: true,
		NoFood:  true,
		NoBeds:  true,
	})

	// Should not panic even with no resources
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Simulation panicked with no resources: %v", r)
		}
	}()

	RunTicks(world, 500, tickDelta)

	// Verify invariants still hold
	assertNoPositionDuplicates(t, world)
	assertCharacterMapConsistency(t, world)
}

// =============================================================================
// Regression Tests
// =============================================================================

// TestRegression_MovementDrainLogsMilestones verifies that energy milestones
// crossed by movement drain are properly logged.
// Bug: Movement drain happened after UpdateSurvival, so thresholds crossed
// by movement were never logged.
func TestRegression_MovementDrainLogsMilestones(t *testing.T) {
	t.Parallel()

	world := CreateTestWorld(WorldOptions{NumCharacters: 1})
	char := world.GameMap.Characters()[0]

	// Set energy just above threshold so one movement crosses it
	char.Energy = 50.1
	char.EnergyCooldown = 0 // Ensure no cooldown protection

	// Clear any existing log entries
	world.ActionLog = system.NewActionLog(100)

	// Run ticks until movement occurs (character will move toward food)
	for i := 0; i < 100; i++ {
		prevEnergy := char.Energy
		RunTick(world, tickDelta)

		// Check if movement crossed the 50 threshold
		if prevEnergy > 50 && char.Energy <= 50 {
			// Look for the milestone log entry
			events := world.ActionLog.AllEvents(100)
			found := false
			for _, e := range events {
				if e.Type == "energy" && e.Message == "Getting tired" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Movement crossed energy threshold 50 but 'Getting tired' was not logged")
			}
			return // Test complete
		}
	}

	t.Skip("Movement did not cross energy threshold in test - inconclusive")
}

// TestRegression_CooldownAppliesToMovementDrain verifies that EnergyCooldown
// prevents movement from draining energy, giving a "freshly rested burst".
// Bug: Cooldown only prevented time-based decay, not movement drain.
func TestRegression_CooldownAppliesToMovementDrain(t *testing.T) {
	t.Parallel()

	world := CreateTestWorld(WorldOptions{NumCharacters: 1})
	char := world.GameMap.Characters()[0]

	// Set up character as freshly rested with cooldown
	char.Energy = 100
	char.EnergyCooldown = 5.0 // Active cooldown

	// Give character intent to move toward food
	char.Hunger = 80 // Hungry enough to seek food

	initialEnergy := char.Energy

	// Run several ticks - movement should occur but not drain energy
	for i := 0; i < 50; i++ {
		RunTick(world, tickDelta)

		// While cooldown is active, energy should not decrease from movement
		if char.EnergyCooldown > 0 && char.Energy < initialEnergy {
			t.Errorf("Energy decreased from %.2f to %.2f while cooldown was %.2f",
				initialEnergy, char.Energy, char.EnergyCooldown+tickDelta)
		}

		// Once cooldown expires, energy can decrease
		if char.EnergyCooldown <= 0 {
			break
		}
	}
}

// TestRegression_CooldownExpiresAndDrainResumes verifies that after cooldown
// expires, movement energy drain resumes normally.
func TestRegression_CooldownExpiresAndDrainResumes(t *testing.T) {
	t.Parallel()

	world := CreateTestWorld(WorldOptions{NumCharacters: 1})
	char := world.GameMap.Characters()[0]

	// Set up character with short cooldown
	char.Energy = 100
	char.EnergyCooldown = 0.5 // Short cooldown
	char.Hunger = 80          // Hungry enough to move

	// Run until cooldown expires
	for char.EnergyCooldown > 0 {
		RunTick(world, tickDelta)
	}

	// Now record energy and run more ticks
	energyAfterCooldown := char.Energy

	// Run more ticks - movement should now drain energy
	RunTicks(world, 50, tickDelta)

	// Energy should have decreased (from movement drain + time decay)
	if char.Energy >= energyAfterCooldown {
		t.Errorf("Energy did not decrease after cooldown expired: was %.2f, now %.2f",
			energyAfterCooldown, char.Energy)
	}
}
