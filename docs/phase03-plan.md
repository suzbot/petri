# Phase III Implementation Plan: Mood and Preference Formation

## Overview

Implement mood stat, dynamic preference formation, and world balancing features from `docs/phase03reqs.txt`. Implementation will be incremental across 4 sub-phases.

## Collaboration Approach

Following our working patterns from CLAUDE.md:

- **Discuss before implementing**: Each step begins with a brief discussion/confirmation
- **Human testing touchpoints**: Marked with 🎮 - pause for manual testing with rebuilt binary
- **Small iterations**: Steps are sized for quick feedback loops
- **Numbered options**: When presenting choices, number them for easy reference

---

## User Decisions

- **Best tier color**: Dark green applies to ALL stats at optimal level (not just Mood)
- **Approach**: Incremental (4 sub-phases, test each before proceeding)
- **Preference weights**: 20% itemType, 20% color, 60% combo (for formation chance, not preference strength)
- **Initial preferences**: FavoriteFood and FavoriteColor convert to two positive preferences (no combo at creation)
- **Field migration**: Remove FavoriteFood/FavoriteColor entirely, use only Preferences slice
- **Preference struct design**: Option B — implicit type from set attributes (extensible without enum explosion)
- **MatchScore calculation**: Option 4 — each matching preference contributes `Valence × AttributeCount`, so combos (2 attrs) count as ±2, single-attr preferences as ±1. Principled, extensible, preserves current behavior.

---

## Sub-Phase A: Mood Stat Foundation

### A1. Add Mood field and tier methods

**Discuss**: Confirm threshold values match requirements (0-10 Miserable, 11-34 Unhappy, etc.)

**File**: `internal/entity/character.go`

- Add `Mood float64` field (starts at 50)
- Add `moodThresholds` and `moodLevels` using existing pattern
- Add `MoodTier()` and `MoodLevel()` methods

**File**: `internal/types/types.go`

- Add `StatMood StatType = "mood"`

**Tests**: Unit tests for mood tier boundaries

### A2. Display Mood in UI + optimal styling

**Discuss**: Confirm dark green color choice, placement in details panel

**File**: `internal/ui/styles.go`

- Add `optimalStyle` (dark green)

**File**: `internal/ui/view.go`

- Add Mood to Details Panel
- Update `colorByTier()` to handle optimal tier for all stats

🎮 **Human Testing**: Rebuild and verify Mood displays in details panel with correct coloring

### A3. Mood updates from needs + tier logging

**Discuss**: Confirm rate values feel right (may need tuning later)

**Decisions made:**

- TierMild = no mood change (neutral)
- Include Health in highest-need calculation (will drive behavior in future)
- Tier transition logging happens here (in UpdateMood)

**File**: `internal/config/config.go`

- Add mood rate constants:
  - `MoodIncreaseRate = 0.5` (when all needs at TierNone)
  - `MoodDecreaseRateSlow = 0.3` (at Moderate)
  - `MoodDecreaseRateMedium = 0.8` (at Severe)
  - `MoodDecreaseRateFast = 1.5` (at Crisis)

**File**: `internal/system/survival.go`

- Add `UpdateMood()` function
- Call from end of `UpdateSurvival()`
- Log tier transitions: "Feeling [MoodLevel]"

**File**: `internal/ui/view.go`

- Add mood keywords to `colorLogMessage()`

**Tests**: Unit tests for mood changes under various need states

🎮 **Human Testing**: Watch mood change over time based on character needs

### A4. Mood boost on consumption

**Discuss**: Confirm boost amount

**File**: `internal/system/consumption.go`

- Add mood boost in `Consume()` and `Drink()`

**Tests**: Integration test for mood evolution

🎮 **Human Testing**: Verify mood boosts when eating/drinking

---

**Sub-Phase A Complete Checkpoint**: All mood mechanics working, tested, documented

---

## Sub-Phase B: Preference System Refactor

### Architecture Context

Per `docs/architecture-review-typed-constants.md`, item attributes are classified as:

| Attribute Type | Examples                                             | Can form opinions? |
| -------------- | ---------------------------------------------------- | ------------------ |
| Descriptive    | ItemType, Color, (future: Material, Pattern, Origin) | Yes                |
| Functional     | Edible, Poisonous, (future: Craftable, Wearable)     | No                 |

Preferences target **descriptive attributes only**. The Preference struct uses implicit typing (which attributes are set) rather than an explicit enum, allowing easy extension for future attributes without combinatorial explosion.

### B1. Create Preference type

**Decided**: Use implicit type from set attributes (Option B)

**File**: `internal/entity/preference.go` (new)

```go
type Preference struct {
    Valence  int           // +1 (likes) or -1 (dislikes)
    ItemType string        // empty if not part of preference
    Color    types.Color   // zero value if not part of preference
    // Future: Material, Pattern, Origin
}
```

- `Matches(item *Item) bool` - true if item matches all set attributes
- `Description() string` - e.g., "red", "berries", "red berries"

**Tests**: Unit tests for preference matching

### B2. Add Preferences to Character

**Decided**: Remove FavoriteFood/FavoriteColor entirely

**File**: `internal/entity/character.go`

- Remove `FavoriteFood string` and `FavoriteColor types.Color` fields
- Add `Preferences []Preference` field
- Update `NewCharacter()` to create two positive preferences from params
- Add `NetPreference(item *Item) int` helper (sum of matching preference valences)

**File**: Update all references to FavoriteFood/FavoriteColor

- `internal/system/movement.go` - food selection logic (updated in B3)
- `internal/ui/view.go` - details panel (updated in B4)
- Test files as needed

**Tests**: Unit tests for NetPreference calculation

🎮 **Human Testing**: Verify character creation still works

### B3. Update food selection logic

**Discuss**: Confirm new matching behavior maintains existing feel

**File**: `internal/system/movement.go`

- Replace FavoriteFood/FavoriteColor checks with `NetPreference()`
- Maintain tiered matching (positive > neutral > negative based on hunger)

**Tests**: Integration tests for food selection

🎮 **Human Testing**: Observe character food choices match preferences

### B4. Update Details Panel for dynamic preferences

**Discuss**: Confirm display format for multiple preferences

**File**: `internal/ui/view.go`

- Replace static preference display with dynamic list
- Handle 0, 1, or many preferences

🎮 **Human Testing**: Verify preferences display correctly

---

**Sub-Phase B Complete Checkpoint**: Preference system refactored, old behavior preserved

---

## Sub-Phase C: Preference Formation

### Decisions Made

- **Formation probabilities**: Miserable 20%, Unhappy 10%, Neutral 0%, Happy 10%, Joyful 20%
  - ⚠️ _Balance note_: Values inflated for testing; will need tuning later
- **Mood adjustment from preferences**: ±5 per point of NetPreference
  - ⚠️ _Balance note_: Inflated for testing; will need tuning later
- **NetPreference scales mood**: Yes, mood boost/decrement scales with NetPreference score
- **Preference coexistence**: Different-specificity preferences can coexist (e.g., "likes berries" + "dislikes red berries")
- **Applies to**: Food consumption only (not springs)
  - ⚠️ _Future_: Extend to other beverages when added

### Order of Operations (Consumption)

1. Reduce hunger, apply poison
2. Mood boost if fully satisfied (existing)
3. Mood adjustment from preferences (new - C3)
4. Try form preference (new - C2)

### C1. Create preference formation logic ✅

**File**: `internal/system/preference.go` (new)

- `TryFormPreference(char, item, log)` function
- Formation chances from config constants
- Type weights: 20% ItemType, 20% Color, 60% Combo
- Existing preference handling: only exact matches count (same specificity)

**File**: `internal/config/config.go`

- Add preference chance constants (easily tunable)

**File**: `internal/entity/preference.go`

- Added `ExactMatch(other)` helper method

**Tests**: Unit tests for formation probability and existing preference handling

### C2. Integrate formation with consumption ✅

**File**: `internal/system/consumption.go`

- Call `TryFormPreference()` at end of `Consume()`

### C3. Mood adjustments from preferences ✅

**File**: `internal/system/consumption.go`

- Add preference-based mood adjustment: `MoodPreferenceModifier × NetPreference`

**File**: `internal/config/config.go`

- Add `MoodPreferenceModifier = 5.0` constant

**Tests**: Unit tests for mood adjustments (4 new tests in consumption_test.go)

---

🎮 **Human Testing Checkpoint**: Binary rebuilt. Test preference formation and mood adjustments before proceeding to C4.

---

### C4. Log preference changes with colors ✅

**File**: `internal/system/preference.go`

- Log "Likes [x]", "Dislikes [x]", "No longer likes/dislikes [x]"

**File**: `internal/ui/view.go`

- Added preference keywords to `colorLogMessage()`:
  - "New Opinion: Likes" → dark green (optimalStyle)
  - "New Opinion: Dislikes" → yellow (severeStyle)
  - "No longer" → light blue (woreOffStyle)
  - "Improved Mood" → dark green (optimalStyle)
  - "Worsened Mood" → yellow (severeStyle)

**File**: `internal/system/consumption.go`

- Added mood change log: "Eating [item] Improved/Worsened Mood (mood X→Y)"

🎮 **Human Testing**: Verify preference changes appear in action log with correct colors

---

**Sub-Phase C Complete Checkpoint**: All C1-C4 implementation complete. Ready for final human testing.

---

## Sub-Phase D: World Balancing

### D8. Status effects on mood

**Discuss**: Confirm mood penalty rates for Poisoned and Frustrated states

**File**: `internal/system/survival.go`

- In `UpdateMood()`, apply mood penalty when `char.Poisoned` is true
- Apply mood penalty when `char.IsFrustrated` is true

**File**: `internal/config/config.go`

- Add `MoodPenaltyPoisoned` constant
- Add `MoodPenaltyFrustrated` constant

**Tests**: Unit tests for mood penalties from status effects

🎮 **Human Testing**: Verify poisoned/frustrated characters have mood decrease

---

### D1. Healing items

**Discuss**: Confirm healing config approach (similar to poison), heal amount

**File**: `internal/entity/item.go`

- Add `Healing bool` field

**File**: `internal/game/world.go`

- Create `HealingConfig` similar to `PoisonConfig`
- Generate 1-2 healing combos (must not overlap with poison)

**File**: `internal/system/consumption.go`

- Apply healing effect when consuming healing item

**File**: `internal/config/config.go`

- Add `HealAmount` constant

**Tests**: Unit tests for healing mechanics

🎮 **Human Testing**: Verify healing items restore health, debug bar shows healing combos

### D2. Item spawning ✅ DECISIONS MADE

**Goal**: Average ~1 spawn per 15 seconds globally, with organic staggered timing.

**Decisions made:**
- Per-item spawn timer approach (each item "reproduces" on its own schedule)
- 8-directional adjacency (including diagonals)
- 50% map coordinate cap to prevent overflow
- Staggered initial timers to avoid synchronized spawning

**Config constants** (`internal/config/config.go`):
```go
ItemSpawnChance          = 0.50  // 50% chance per spawn opportunity
ItemSpawnIntervalBase    = 8.0   // seconds, multiplied by initial item count
ItemSpawnIntervalVariance = 0.20 // ±20% randomization
ItemSpawnMaxDensity      = 0.50  // max 50% of map coordinates
```

**Implementation** (`internal/system/spawning.go`):

1. Add `SpawnTimer float64` field to Item struct
2. On world gen: `item.SpawnTimer = rand(0, intervalBase × initialItemCount)` (stagger across one cycle)
3. Each tick: decrement all item timers by delta
4. When timer ≤ 0:
   - Check map not at 50% capacity
   - Find random empty 8-adjacent tile
   - Roll 50% chance
   - If success: spawn matching item (same type, color, poison/healing status)
   - Reset timer to `intervalBase × initialItemCount × (1 ± variance)`

**Math verification**: 40 items × 50% ÷ 320s = 1 spawn per 16 seconds ✓

**File**: `internal/entity/item.go`
- Add `SpawnTimer float64` field

**File**: `internal/system/spawning.go` (new)
- `UpdateSpawnTimers(items, delta, gameMap, poisonCfg, healingCfg, initialItemCount)`
- `findEmptyAdjacent(x, y, gameMap)` - 8-directional

**File**: `internal/game/map.go`
- Add `TotalCoordinates()` and `EntityCount()` helpers for cap check

**Tests**: Unit tests for spawn timer logic, cap enforcement

🎮 **Human Testing**: Watch items spawn over time, verify ~15s average spacing

### D3. Flower item type ✅ DECISIONS MADE

**Decisions**:
- Symbol: `❀` (unicode flower)
- Colors: Red, Orange, Yellow, Blue, Purple, White (per requirements)
- Spawn count: 20 (same as berries/mushrooms)
- Not edible (purely decorative)
- Reproduces like other items (same spawn behavior)
- Note: Future enhancement - different item types may have different reproduction rates

**File**: `internal/types/types.go`

- Add colors: Orange, Yellow, Purple
- Add `FlowerColors` slice

**File**: `internal/entity/item.go`

- Add `NewFlower()` constructor (Edible: false)

**File**: `internal/config/config.go`

- Add `CharFlower = '❀'`, `FlowerSpawnCount = 20`

**File**: `internal/game/world.go`

- Add flower spawning to `SpawnItems()`

**File**: `internal/ui/view.go`

- Add orange, yellow, purple color rendering

🎮 **Human Testing**: Verify flowers appear on map with correct symbols/colors

### D4. Looking activity

**Discuss**: Confirm looking mechanics (adjacent positioning, duration, interruption)

**File**: `internal/entity/character.go`

- Add `ActionLook` to ActionType enum

**File**: `internal/system/movement.go`

- Add `findLookIntent()` for idle characters

**File**: `internal/ui/update.go`

- Handle `ActionLook` with duration

**File**: `internal/system/consumption.go` (or new file)

- Add `CompleteLook()` function

**File**: `internal/config/config.go`

- Add `LookChance`, `LookDuration`

**Tests**: Integration tests for looking activity and interruption

🎮 **Human Testing**: Observe characters looking at items when idle, verify interruption by needs

---

### D5. Pattern and Texture attributes

**Discuss**: Confirm attribute values, how they affect poison/healing combos

**File**: `internal/types/types.go`

- Add `Pattern` type (Spotted, Plain)
- Add `Texture` type (Slimy, None)

**File**: `internal/entity/item.go`

- Add `Pattern` and `Texture` fields to Item
- Update `NewMushroom()` to accept pattern/texture params
- Update `Description()` to include pattern/texture

**File**: `internal/game/world.go`

- Update mushroom spawning to generate random pattern/texture combinations
- Update poison/healing config to include pattern/texture in combo matching

**File**: `internal/entity/preference.go`

- Add `Pattern` and `Texture` fields to Preference struct
- Update `Matches()` and `Description()` methods

**Tests**: Unit tests for pattern/texture matching

🎮 **Human Testing**: Verify mushrooms display with patterns/textures, poison/healing combos work

---

### D6. Preference formation with expanded attributes

**Discuss**: Confirm combo constraints (ItemType + 1-2 other attributes), formation weights

**File**: `internal/system/preference.go`

- Update `rollPreferenceType()` to handle Pattern/Texture
- Ensure combo preferences include ItemType + other attributes

**File**: `internal/config/config.go`

- Adjust formation weights if needed

**Tests**: Unit tests for expanded preference formation

🎮 **Human Testing**: Observe preference formation includes new attributes

---

### D7. Seeking and avoidance refinements

**Discuss**: Confirm threshold adjustments for expanded NetPreference range, avoidance behavior

**File**: `internal/system/movement.go`

- Update `findFoodTarget()` thresholds for higher NetPreference values
- Consider avoidance logic for strongly disliked items (if desired)

**Tests**: Integration tests for food selection with expanded preferences

🎮 **Human Testing**: Verify food selection behavior with new attribute preferences

---

### Balance Tuning Pass

**Discuss**: Review overall game feel after D1-D7 complete

Areas to evaluate:

- Activity durations (eating feels too quick?)
- Stat change rates
- Preference formation chances
- Mood adjustment rates and modifiers
- Any other config values that feel off

This is an iterative tuning session based on gameplay feel, not a fixed implementation.

🎮 **Human Testing**: Extended play session to identify balance issues

---

**Sub-Phase D Complete Checkpoint**: World balancing features complete, game has more variety

---

## Critical Files Summary

| File                             | Changes                                                               |
| -------------------------------- | --------------------------------------------------------------------- |
| `internal/entity/character.go`   | Mood field, Preferences slice, ActionLook                             |
| `internal/entity/preference.go`  | NEW - Preference type and methods, Pattern/Texture fields             |
| `internal/entity/item.go`        | Healing field, Pattern/Texture fields                                 |
| `internal/system/survival.go`    | UpdateMood() integration, status effect mood penalties                |
| `internal/system/consumption.go` | Mood boosts, preference formation, healing, looking                   |
| `internal/system/preference.go`  | NEW - TryFormPreference(), expanded attribute formation               |
| `internal/system/movement.go`    | findLookIntent(), update findFoodTarget() with expanded thresholds    |
| `internal/ui/view.go`            | Mood display, optimal styling, preference list                        |
| `internal/ui/styles.go`          | optimalStyle (dark green)                                             |
| `internal/ui/update.go`          | ActionLook handling                                                   |
| `internal/game/world.go`         | HealingConfig, flower spawning, item spawning, pattern/texture combos |
| `internal/config/config.go`      | All new constants                                                     |
| `internal/types/types.go`        | StatMood, new colors, Pattern, Texture                                |

---

## Workflow Summary

Each step follows this pattern:

1. **Discuss** - Brief confirmation before implementing
2. **Implement** - Code changes with tests
3. **🎮 Test** - Human testing touchpoint (rebuild binary first!)

Sub-phases are designed to be independently testable. Complete one before starting the next.
