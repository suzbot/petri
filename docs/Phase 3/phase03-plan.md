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

### D8. Status effects on mood ✅ COMPLETE

Mood penalties applied for Poisoned and Frustrated states in `UpdateMood()`.

---

### D1. Healing items ✅ COMPLETE

Healing items implemented with HealingConfig mirroring PoisonConfig pattern.

### D2. Item spawning ✅ COMPLETE

Per-item spawn timers with 8-directional adjacency and 50% map cap implemented.

### D3. Flower item type ✅ COMPLETE

Flowers added with 6 colors (Red, Orange, Yellow, Blue, Purple, White), non-edible.

### D4. Looking activity ✅ COMPLETE

**Decisions**:
- LookChance: 50% when idle
- LookDuration: 3.0 seconds (eating changed to 1.5s)
- Target: Closest item of any type (berries, mushrooms, flowers)
- Position: Closest available adjacent tile to target item
- If already adjacent: Skip travel, start looking immediately
- Preference formation: Call TryFormPreference when looking completes
- Mood impact: Looking at liked/disliked items affects mood (same as eating)
- Cooldown: 10 seconds between look checks (applies whether roll succeeds or fails)
- Different target: After looking at an item, next look must target a different item
- Interruption: Moderate+ needs interrupt looking (stickier than idle, which responds to Mild+)

**Files changed**:

- `internal/entity/character.go` - ActionLook enum, LookCooldown/LastLookedX/Y tracking fields
- `internal/system/movement.go` - findLookIntent(), helper functions for adjacency/distance
- `internal/ui/update.go` - Handle ActionLook with 3.0s duration
- `internal/system/looking.go` (new) - CompleteLook() with preference mood impact
- `internal/system/survival.go` - Decrement LookCooldown timer
- `internal/config/config.go` - LookChance, LookDuration, LookCooldown constants

**Tests**: looking_test.go with tests for CompleteLook, adjacency, targeting, intent creation

🎮 **Human Testing**: Complete - looking works with cooldown, mood impact, and preference formation

---

### D5. Pattern and Texture attributes ✅ COMPLETE

**Status**: ✅ COMPLETE

#### Approach: Variety for Generation Only

After evaluating the full refactor scope (15+ files, 100+ call sites), we pivoted to a simpler approach:

- **Variety system used for world generation only** - decides what combos exist, assigns poison/healing
- **Items keep embedded fields** - no registry lookups at runtime, items remain self-contained
- **Much smaller refactor** - ~70% less code churn than variety-by-reference approach

#### Decisions Made

- **Pattern values**: Spotted, Plain (mushrooms only)
- **Texture values**: Slimy, None (mushrooms only)
- **Variety count**: `max(2, spawnCount / 4)` (configurable via VarietyDivisor)
- **Poison/Healing**: 20% of edible varieties each (configurable), mutually exclusive

#### Completed Steps

- ✅ D5.1: Pattern/Texture types added to `types/types.go`
- ✅ D5.2: ItemVariety and VarietyRegistry created (for generation)
- ✅ D5.3: Variety generation logic with poison/healing assignment
- ✅ D5.4: Added Pattern/Texture fields to Item, updated NewMushroom signature
- ✅ D5.5: Updated world generation to use varieties, removed poisonCfg/healingCfg from Model
- ✅ D5.6: Updated spawning system to inherit Pattern/Texture from parent
- ✅ D5.7: Updated preference system (30% single / 70% combo weights, Pattern/Texture support)
- ✅ D5.8: Updated UI to display Pattern/Texture for mushrooms
- ✅ D5.9: Cleaned up old PoisonConfig/HealingConfig code

🎮 **Human Testing**: Ready for testing

---

### D6. Preference formation with expanded attributes ✅ COMPLETE

**Status**: ✅ COMPLETE

#### Decisions Made

- **Solo preferences (30%)**: Can be any single attribute (ItemType, Color, Pattern, or Texture)
  - Pattern/Texture solos use noun forms: "Likes Spots", "Likes Slime"
- **Combo preferences (70%)**: Must include ItemType + 1-2 other attributes
  - Max 3 attributes total (4-attribute combos deferred to Phase 4)
  - Distribution within combo: generative, not hardcoded permutations
- **Weights**: Keep simple - 30% solo, 70% combo (no sub-weights needed)

#### Implementation

**D6.1: Update rollPreferenceType() for combo logic**

**File**: `internal/system/preference.go`

- Combos always include ItemType
- Pick 1-2 additional attributes from available [Color, Pattern, Texture]
- Use generative approach: collect extras, shuffle, take 1 or 2

```go
// Combo: Always include ItemType, plus 1-2 other attributes
attrs := []string{"itemType"}
extras := collectExtraAttributes(item) // excludes itemType
numExtras := 1
if len(extras) >= 2 && rand.Float64() < 0.5 {
    numExtras = 2
}
// shuffle extras, append first numExtras to attrs
```

**D6.2: Update Description() for solo noun forms**

**File**: `internal/entity/preference.go`

- When Pattern is solo (no other attrs): use "Spots" instead of "spotted"
- When Texture is solo (no other attrs): use "Slime" instead of "slimy"
- Combo descriptions unchanged: "spotted mushrooms", "slimy red mushrooms"

**D6.3: Tests**

**File**: `internal/system/preference_test.go`

- Test solo preferences include all attribute types
- Test combos always include ItemType
- Test combos limited to 2-3 attributes
- Test Description() noun forms for solo Pattern/Texture

🎮 **Human Testing**: Observe preference formation includes new attributes with correct descriptions

---

### D7. Seeking and avoidance refinements

**Discuss**: Confirm threshold adjustments for expanded NetPreference range, avoidance behavior

**File**: `internal/system/movement.go`

- Update `findFoodTarget()` thresholds for higher NetPreference values
- Consider avoidance logic for strongly disliked items (if desired)

**Tests**: Integration tests for food selection with expanded preferences

🎮 **Human Testing**: Verify food selection behavior with new attribute preferences

---

### D9. Dynamic color options in character creation ✅ COMPLETE

**Status**: ✅ COMPLETE

Character creation now dynamically builds color options from `types.AllColors`. When new colors are added to the types package, they automatically appear in character creation.

**Files changed**:
- `internal/ui/creation.go` - Removed hardcoded color constants, added `buildColorOptions()`, `capitalizeColor()`, `DisplayToColor()`
- `internal/ui/update.go` - Use `DisplayToColor()` instead of switch statement
- `internal/ui/creation_test.go` - Updated tests to use dynamic `colorOptions`

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
