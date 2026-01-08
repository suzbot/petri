# Bug Fixes and Regression Tests

This document tracks bugs that were found and fixed, along with their regression tests.

## 2025-12-31

### Pause/Unpause Delta Accumulation

**Bug:** When game was paused then unpaused, all accumulated time delta was applied at once, causing instant stat changes/damage.

**Root Cause:** `lastUpdate` timestamp continued advancing during pause. When unpaused, delta calculation used stale timestamp.

**Fix:** Reset `lastUpdate` to `time.Now()` when unpausing.

**Location:** `ui/update.go` (space key handler)

**Regression Test:** UI-level timing issue, difficult to unit test. Verified manually.

---

### Energy Milestones Not All Logged

**Bug:** Energy tier logging used `else if` chain, so if energy dropped across multiple thresholds in one tick, only the first was logged.

**Root Cause:** `else if` prevents subsequent conditions from being checked after first match.

**Fix:** Changed `else if` to independent `if` statements.

**Location:** `system/survival.go` (energy milestone logging)

**Regression Test:** `TestRegression_EnergyMilestonesAllLogged` - Drop energy across multiple thresholds in one tick, verify all milestones logged.

---

### Movement Drain Not Logging Milestones

**Bug:** Movement energy drain happened after `UpdateSurvival`, so threshold crossings caused by movement were never logged. The `prevEnergy` captured in the next tick's `UpdateSurvival` was already past the threshold.

**Root Cause:** Energy milestone logging only happened in `UpdateSurvival`, but movement also drains energy in `applyIntent`.

**Fix:** Added milestone logging after movement energy drain in `applyMoveIntent`.

**Location:** `ui/update.go` and `simulation/simulation.go` (movement energy drain section)

**Regression Test:** `TestRegression_MovementDrainLogsMilestones` - Have movement cross an energy threshold, verify milestone is logged.

---

### Cooldown Not Applying to Movement Drain

**Bug:** `EnergyCooldown` only prevented time-based decay, not movement drain. Freshly rested characters (energy = 100) had energy drop immediately on first move, so 100 was never visually displayed.

**Root Cause:** Movement drain didn't check `EnergyCooldown` before draining.

**Fix:** Check `EnergyCooldown > 0` before draining energy for movement. This gives a "freshly rested burst" of free movement.

**Location:** `ui/update.go` and `simulation/simulation.go` (movement energy drain section)

**Regression Test:** `TestRegression_CooldownAppliesToMovementDrain` - Set `EnergyCooldown > 0`, move character, verify energy unchanged.

---

## Pre-existing (documented in testing_approach.txt)

### Feature Occupation Thrashing

**Regression Test:** `TestRegression_FeatureOccupationThrashing`

### Intent Abandoned When Occupied

**Regression Test:** `TestRegression_IntentAbandonedWhenOccupied`
