# Step Spec: Step 12 — Display, Label, and Color Polish

Design doc: [construction-design.md](construction-design.md)

## Overview

Step 12 is a polish pass on player-facing display and feedback that has accumulated through the construction phase. The seven scoped items (DD-53, brick validation, DD-54, DD-55, DD-56, DD-57, DD-58) are independent — they share theme but not implementation surface. The step is broken into four sub-steps grouped by code surface and test surface, not dependency.

**Checkpoint structure (deviation from default):** Per refinement discussion, this step uses **intermediate [TEST] checkpoints** between sub-steps but **one consolidated [DOCS] and [RETRO] at the end**. Rationale: each sub-step's documentation footprint is tiny (a display name, a label tweak), and a single retro at the end captures cross-cutting observations from a polish pass without ceremony overhead per item.

## Sub-step structure (breakout decision)

Four sub-steps grouped by **code surface** (where the changes land) and **test surface** (what the user verifies in one pass):

- **12a — Display name and label cleanups** (DD-53, brick validation, DD-54, DD-56) — clusters all text/label fixes; user verifies in same UI session.
- **12b — Abandoned order cooldown timer** (DD-55) — standalone UI work with state.
- **12c — Discovery randomization** (DD-57) — touches behavior code; verifiable via observation tests.
- **12d — Thatch color evaluation + propagate** (DD-58) — last because user-judgment pause for color evaluation doesn't block downstream sub-steps.

Sequencing: easiest mechanical fixes first → UI state work → behavior change → user-judgment work last.

---

## Sub-step 12a — Display name and label cleanups — COMPLETE

**Anchor story:** A character drops a stick. The action log reads "drops stick" (not "drops brown stick"). They pick up another and bundle it — the inventory shows "bundle of sticks (2)". The player opens the orders menu and creates a Gather > Tall Grass order — the menu reads "Gather tall grass" (not "Gather pale green tall grass"). They create a Gather > Clay order — it reads "Gather lumps of clay" (not "Gather clays"). They open the construction submenu — entries read "Build fence" and "Build hut" (not bare "Fence" / "Hut").

### Drift point — DD-53 vs bundleable items

DD-53 prescribes `Name: "stick"` on NewStick, but `Item.Description()` (`internal/entity/item.go:322`) returns Name as an early return — bundleable items with Name set lose their "bundle of X (N)" formatting. **Resolution:** Reorder `Description()` so bundle check precedes Name check. Clay (the only existing Name-having item) is not bundleable, so this change has no effect on existing display. Enables Name to coexist with bundling.

**Bringing tall grass in line:** With the reorder, the same Name-field pattern applies cleanly to tall grass too. Set `Name: "tall grass"` on NewGrass; the bundle path still uses `Pluralize(Kind)` which already returns "tall grass" as collective noun (preference.go:274), so bundle display stays "bundle of tall grass (N)". This avoids a "suppress color when Kind == 'tall grass'" special case in Description(). Verified via test world: single tall grass exists transiently in inventory after first harvest before subsequent merges, so the single-item path matters.

### Implementation

**Files:**
- `internal/entity/item.go`:
  - Add `Name: "stick"` to `NewStick` (line 183).
  - Add `Name: "tall grass"` to `NewGrass` (line 166).
  - Reorder `Description()` so the `BundleCount >= 2` check runs before the `Name != ""` check.
- `internal/entity/preference.go`:
  - Add `case "clay": return "lumps of clay"` to `Pluralize()`.
- Order `DisplayName()` (location: `internal/entity/order.go` per grep — confirm during implementation):
  - For construction-category orders, prefix `"Build "` and lowercase the activity name (e.g., "Build fence", "Build hut"). Matches existing craft pattern ("Craft vessel", "Craft hoe").

**Architecture pattern:** Follows existing Name-field convention (brick, clay precedent), now extended uniformly to bundleable items via the reorder. Follows Pluralize special-case pattern (existing entries: berry, mushroom, flower, tall grass). Follows craft-category verb-prefix pattern.

**Brick validation:** Bricks already have `Name: "brick"` (entity/item.go:221). After reordering Description(), confirm bricks still display as "brick" in:
- Action log entries (drops, picks up, crafts)
- Details panel
- Inventory display
- After save/load round-trip

### Tests

Add unit tests for:
- `Item.Description()`: stick single → "stick"; stick bundle of 5 → "bundle of sticks (5)"
- `Item.Description()`: tall grass single → "tall grass"; tall grass bundle of 5 → "bundle of tall grasses (5)"
- `Pluralize("clay")` → "lumps of clay"
- Construction order `DisplayName()`: fence order → "Build fence"; hut order → "Build hut"

Regression checks:
- Brick `Description()` still returns "brick" (single)
- Brick save/load preserves Name (likely already covered by existing serialize tests; verify)

### [TEST] checkpoint — 12a

Run unit and existing integration tests. Then human test:
1. Load a normal game (no test world needed). Have characters perform actions producing each affected item type.
2. Open orders menu → confirm "Build fence" / "Build hut" labels and "Gather lumps of clay".
3. Drop and pick up sticks → confirm action log reads "stick" / "bundle of sticks (N)".
4. Create a Gather > Tall Grass order → confirm "Gather tall grass" label.
5. Confirm bricks still display correctly (action log, details panel).

---

## Sub-step 12b — Abandoned order cooldown timer

**Anchor story:** A character abandons a Gather > Berries order (no berries reachable). The orders panel shows "Abandoned (1:53)" — the timer counts down each tick. At "0:00" the status flips back to "Open" and another character can take it again.

### Implementation

**Files:**
- `internal/entity/order.go`: AbandonCooldown field already exists per grep. Add or update a status-text helper that returns `"Abandoned (M:SS)"` when in cooldown, falling through to current "Abandoned" / "Open" logic otherwise.
- `internal/ui/update.go`: Confirm AbandonCooldown is being decremented per tick (likely already true since the field exists and orders flip back to Open today).
- Orders panel renderer (location TBD — likely `internal/ui/view.go` or a panel-specific file): Use the new status helper.

**Architecture pattern:** Extends existing order-status display. Follows existing time-formatting convention if one exists (otherwise add a small `formatCooldown(seconds float64) string` helper, format `M:SS`).

### Tests

Add unit tests for:
- Time format helper: 113s → "1:53"; 60s → "1:00"; 5s → "0:05"
- Order status helper: order with AbandonCooldown > 0 → "Abandoned (M:SS)"; order with AbandonCooldown == 0 and not abandoned → "Open" (or current value)

### [TEST] checkpoint — 12b

Use `/test-world` to set up a scenario with a fresh order that's about to be abandoned (e.g., Gather order with no targets reachable, or a known abandon scenario from existing tests).

1. Watch the orders panel as the order abandons → confirm "Abandoned (M:SS)" appears.
2. Confirm timer counts down each tick.
3. Confirm transition back to "Open" when timer expires.

---

## Sub-step 12c — Discovery randomization

**Anchor story:** A character looks at clay. Both a "dig" activity discovery and a "craftBrick" recipe discovery share the same trigger. Across many such interactions over time, the character discovers the activity first about half the time and the recipe first the other half — varied rather than always activity-first.

### Implementation

**Files:**
- `internal/system/discovery.go`: `TryDiscoverKnowHow` (line 29) currently calls `tryDiscoverActivity` then `tryDiscoverRecipe` deterministically. Change to flip a coin (50/50): half the time call activity-then-recipe, half the time call recipe-then-activity. Preserves the "only one discovery per interaction" limit (early return on success).

**Architecture pattern:** Minimal change to discovery system structure. Follows existing `rand.Float64()` usage already present in discovery code.

### Tests

Add unit tests for:
- Seeded RNG verifying both call orderings can occur
- With seed favoring 0–0.5: activity tried first (existing path)
- With seed favoring 0.5–1.0: recipe tried first (new path)

Add observation test in `internal/simulation/observation_test.go`:
- Set up a character with no discovered activities/recipes and a trigger that matches both an activity and a recipe (e.g., looking at clay matches both "dig" activity and "craftBrick" recipe per design)
- Run N iterations (N = 1000 or whatever existing observation tests use), recording which discovery fires first when both could
- Verify distribution sits within statistical bounds (e.g., 40–60% activity-first over 1000 trials)

**Architecture pattern:** Follows existing observation-test conventions in `internal/simulation/observation_test.go` for measuring emergent/probabilistic behavior.

### [TEST] checkpoint — 12c

Run unit tests + observation test (both must pass).

Human spot-check: load a normal game with a character who can discover both activity and recipe from the same trigger. Watch action log over a few play sessions to confirm both orderings show up. (Lighter-weight than 12a/12b human testing since observation test is doing the heavy verification.)

---

## Sub-step 12d — Thatch color evaluation + propagate

**Anchor story (12d-i):** The user runs `go run ./cmd/colordemo` and sees ColorPaleYellow (ANSI 229) alongside candidate colors (ANSI 178, 179, 186) rendered side-by-side as both loose dried grass items and as thatch fence wall and hut wall segments. They evaluate which reads best in context and report the choice.

**Anchor story (12d-ii):** With the user's chosen ANSI value applied, thatch constructs in the actual game render with the new color. Loose dried grass items use the same color. Brown sticks and terracotta bricks remain unchanged.

### Implementation

**12d-i — Build demo:**
- New file: `cmd/colordemo/main.go` (follows precedent: `cmd/hutdemo` from DD-42 evaluation).
- Render a small grid showing each candidate color (229, 178, 179, 186) used for:
  - Loose dried-grass item symbol (`config.CharGrass` post-harvest)
  - Thatch fence wall segment (`╬` with material color)
  - Thatch hut wall segments (`━`, `┃`, corner pieces) per hut rendering
- Print legend showing which ANSI value is which.

**12d-ii — Propagate chosen color:**
- After user picks: update the ColorPaleYellow constant (or replace with a new color name if user prefers a different identifier) in `internal/types/types.go`.
- Verify all uses (12 files per grep) still render appropriately — most should be transparent since they reference the constant by name.
- If user chooses a new identifier (e.g., ColorThatch), introduce the new constant and migrate references in dried-grass items and thatch construction recipes only; leave any unrelated ColorPaleYellow uses alone.

**Architecture pattern:** Demo program pattern (precedent: cmd/hutdemo for DD-42 box-drawing evaluation). Color is a visual judgment best made in context, not decided theoretically.

### Tests

No code tests for color choice (visual judgment).

After 12d-ii propagation:
- Existing rendering tests should still pass without modification (color is referenced by constant name).
- If a new color constant is introduced, add a smoke test that thatch constructs reference the new constant.

### [TEST] checkpoint — 12d

After 12d-i: user runs `go run ./cmd/colordemo`, evaluates, reports chosen ANSI value (or new identifier choice).

After 12d-ii: user loads a game with thatch fence and thatch hut visible (test world or natural play), confirms new color reads well for both items and constructs. Confirms unrelated colors unchanged.

---

## [DOCS] — End of step

After all four sub-steps test successfully, run `/update-docs`. Updates likely include:
- `docs/construction-design.md`: Mark Step 12 complete; update DD-53 (note bundle/Name reorder), DD-54, DD-55, DD-56, DD-57, DD-58 as resolved if any details changed during implementation.
- `docs/game-mechanics.md`: Update if any display behavior is documented there (likely minor or no changes).
- `docs/step-spec.md`: Replace contents with pointer to Step 13.

## [RETRO] — End of step

After [DOCS] completes, run `/retro` for the whole step. Single retro covers cross-cutting observations from the polish pass — likely lighter than feature-step retros, but worth running before moving to Step 13.

---

## Implementation-ready checklist

- [x] All open questions from design doc addressed (Step 12 had none listed)
- [x] All triggered enhancements evaluated (Step 12 had none listed)
- [x] Drift check passed — DD-53 / bundle interaction surfaced and resolved
- [x] Label and display text decisions confirmed with user (DD-56 lowercase-after-verb matches craft pattern; DD-53 unified Name treatment for sticks and tall grass confirmed via test world observation 2026-04-21)
- [x] Architecture patterns named per sub-step
- [x] Anchor story per sub-step
- [x] Sub-step boundaries align with test surface

## Confirmed decisions (refinement, 2026-04-21)

1. **DD-53 / bundle drift resolution:** Option 1 — reorder Description() so bundle check precedes Name check. Apply Name field to **both** sticks and tall grass uniformly. Single stick → "stick"; stick bundle → "bundle of sticks (N)". Single tall grass → "tall grass"; tall grass bundle → "bundle of tall grass (N)" (Pluralize already handles "tall grass" as collective noun). Clay (only existing Name-having item) is not bundleable, so reorder is invisible to it. Eliminates the proposed special-case "suppress color when Kind == 'tall grass'" logic; uses the same pattern uniformly.
2. **[DOCS] placement:** Single [DOCS] at end alongside [RETRO]. Each sub-step's documentation footprint is small enough that consolidating is cleaner than per-sub-step doc updates.

## Decision deferred to 12d-i evaluation

3. **Color identifier (general vs thatch-specific):** The chosen color may warrant a general-purpose constant name (e.g., ColorWheat, ColorStraw — usable by future items characters could form preferences about) or a thatch-specific name (ColorThatch — if so distinctly thatch-toned that no other item would use it). Decide during 12d-i once the color is selected. If general-purpose, also evaluate whether to rename or replace ColorPaleYellow rather than introduce a parallel constant.
