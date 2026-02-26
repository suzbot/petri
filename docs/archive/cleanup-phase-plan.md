# Post-Gardening Cleanup Phase Plan

Requirements: [post-gardening-cleanup.md](post-gardening-cleanup.md)

---

## Pre-Work

### 0A. Whitespace Investigation ✅

**Outcome:** Root cause is Edit tool ambiguity with Go tab indentation in line-number output format. Not a codebase issue — a Claude workflow issue. Mitigations added:
- `gofmt ./...` post-edit step added to `/implement-feature` skill
- Tab-preservation reminder added to MEMORY.md

### 0B. Test File Organization ✅

**Outcome:** Evaluated — not actionable now. 26K total lines across 37 files. Top 3: update_test (3.2K), movement_test (2.9K), order_execution_test (2.5K). These are big because the systems they test are big. Splitting by sub-concern is possible but creates file proliferation without reducing total lines. Paring down risks losing coverage before a phase that touches multiple systems. Targeted Grep/Read handles large files fine for most searches. Added to triggered-enhancements.md with trigger: "a specific file becomes painful to navigate during implementation."

---

## Implementation Order

1. **2A** — Sticky BFS pathing ✅
2. **2C** — Numbered order selection ✅
3. **2B** — Esc key cleanup ✅
4. **2D** — Till soil selection colors ✅
5. **3A** — Gather orders ✅
6. **3B** — Satiation & consumption duration ✅
7. **3C** — Satiation-aware food selection ✅
8. **4A** — Character creation streamline: dead code + R/C screen ✅
9. **4B** — Variable-count creation grid ✅

---

## Part 1: Investigations

Lightweight assessments that inform whether follow-up work exists. Fine outcome for any of these: "no action needed."

### 1A. Misplaced Functions Audit ✅

**Outcome:** No action needed. `createItemFromVariety` is in `game/world.go` (correctly placed — world generation). `ConsumeAccessibleItem` is in `entity/character.go` (correctly placed — character inventory). `CreateVessel`/`CreateHoe` in `system/crafting.go` are correctly placed — crafting logic belongs in the system layer. Everything follows the entity/system/game separation cleanly.

### 1B. ItemType Constants Evaluation ✅

**Outcome:** No action needed. 11 distinct ItemType values across 45 occurrences in 14 files. Zero typos or inconsistencies found. Constants would add indirection without solving an observed problem. Revisit if a typo actually causes a bug or the count grows significantly.

### 1C. Action Pattern Unification Investigation ✅

**Outcome:** No action needed. Patterns are appropriately different, not diverging:
- Eating (`Consume`/`ConsumeFromInventory`/`ConsumeFromVessel`): applies effects (hunger, poison, preferences)
- Crafting/planting (`ConsumeAccessibleItem`/`ConsumePlantable`): extracts item only, no effects
- `HasAccessibleItem` is always read-only, never extracts
- No "extract at intent time" anti-pattern exists anywhere — all consumption happens at action completion

---

## Part 2: Quick Fixes

Each is independently shippable and human-testable, with its own [TEST], [DOCS], [RETRO] cycle.

### 2A. Pathing Improvement ✅

**Player impact:** Characters no longer thrash back and forth when a target is across an irregularly shaped pond. Once a character switches to BFS pathfinding to navigate around an obstacle, they stay in BFS mode until they reach their target or bump into another character.

**Reqs reconciliation:** Post-gardening-cleanup.md section A — "once a character switches to BFS, they stay in BFS until they get to target or until they run into a character. If they run into a character they sidestep (as in current state) and then switch back to greedy step."

**Architecture alignment:** Movement & Pathfinding in architecture.md — `NextStepBFS` is greedy-first-then-BFS. Currently stateless (decides greedy vs BFS from scratch each tick). Sticky BFS adds per-character ephemeral state following the displacement precedent (`DisplacementStepsLeft`, `DisplacementDX`, `DisplacementDY`) — not serialized, clears on save/load.

**Values alignment:** Consistency Over Local Cleverness — follows the existing ephemeral state pattern used by displacement. Start With the Simpler Rule — simple flag with two clear-conditions, no edge-case logic.

**Resolved questions:**
- Flag clears when intent is nilled (covers reaching target + intent changes) and when displacement initiates (covers character collision). No special clearing logic needed — piggybacks on existing state transitions.
- Flag does NOT persist across intent changes. A new intent may path in a different direction where greedy is appropriate.

**Implementation:**

**Step 1: Sticky BFS**

Character paths smoothly around obstacles instead of oscillating between greedy and BFS steps.

1. **Tests first:** Test that a character with a target across an irregular pond converges toward the target without revisiting positions (anti-thrashing). Test that the BFS flag clears after displacement initiation.
2. Add ephemeral `UsingBFS bool` to Character struct (not serialized, like displacement fields).
3. Modify `NextStepBFS` to accept a `preferBFS bool` parameter and return `usedBFS bool`. When `preferBFS` is true, skip the greedy attempt and go straight to BFS. Return `true` for `usedBFS` whenever BFS was actually used.
4. Caller sets `char.UsingBFS = true` when `usedBFS` is returned. Caller passes `char.UsingBFS` as `preferBFS`.
5. Clear `UsingBFS` in two places: (a) when intent is nilled (wherever `Intent` is set to nil), (b) when `initiateDisplacement` fires.
6. Update all `NextStepBFS` callers for the new signature.

[TEST] Create a test world with a character and a target item on the other side of an irregularly shaped pond. Verify the character paths smoothly around the pond without oscillating.

**Step 2:** [DOCS] + [RETRO]

---

### 2B. Esc Key Cleanup

**Player impact:** Esc consistently means "go back one level" everywhere in the UI. `q` becomes the "go to world select" key (no longer quits). `l` returns to the action log from any details subpanel. After creating a plant order, you stay at the Gardening sub-category instead of jumping to the top. Status bar hints update contextually to reflect what Esc/q will do.

**Reqs reconciliation:** Post-gardening-cleanup.md section B — desired behavior list. All items addressed:
- Esc from any expanded view → collapse
- Esc from all-activity view → no-op
- Esc from orders (normal view) → all-activity view
- Esc from within orders add/cancel → back one level
- Esc from select details > action log or no panel → all-activity view
- Esc from select details > any other panel → details action log
- `l` from select details > any other panel → details action log (NEW keystroke)
- `q` from anywhere → world select screen
- `q` from world select → quit

Additional scope from discussion:
- Post-order-creation auto-back: step 2 plant confirm currently jumps to step 0 (top-level activity list). Should go to step 1 (Gardening sub-category) so you can immediately create another order. (tillSoil already goes to step 1 — this makes plant match.)
- `L: Log` hint in details panel when a subpanel (knowledge/inventory/preferences) is open.
- Status bar hints: always-visible, context-dependent. `ESC=collapse` when expanded, `ESC=back` when in orders/select, no esc hint in all-activity (no-op). `q=menu` always shown.

**Architecture alignment:** MVU pattern (architecture.md "Key Design Patterns") — all key handling flows through `handleKey()` in update.go, view hints in view.go. Changes are to key dispatch logic and hint rendering. No new architecture patterns.

**Values alignment:** Consistency Over Local Cleverness — "back one level" is a single principle applied uniformly, replacing ad-hoc per-state behavior. Start With the Simpler Rule — one esc rule instead of per-state special cases.

**Implementation:**

**Step 1: Investigate current behavior delta ✅**

Investigated. x-key expansion only applies to ordersFullScreen and activityFullScreen (knowledge/inventory/preferences panels don't have expanded views). 8 behavior changes needed plus hint updates.

**Step 2: Implement Esc/q/l changes ✅**

Player experiences consistent "back one level" for Esc, `q` as navigate-to-world-select, `l` as return-to-action-log, and contextual status bar hints.

1. **Tests first:** Key handling is UI logic (no unit tests per CLAUDE.md testing policy). Verified entirely via human testing.

2. **Esc handler restructure** (update.go `handleKey`, phasePlaying esc case). New priority order:
   1. `ordersFullScreen` true → set false, return (collapse expanded orders)
   2. `activityFullScreen` true → set false, return (collapse expanded activity)
   3. ordersAddMode → back one level (UNCHANGED — existing logic)
   4. ordersCancelMode → exit (UNCHANGED — existing logic)
   5. subpanels open → close (UNCHANGED — existing logic)
   6. `showOrdersPanel` true (normal, not add/cancel) → close panel, switch to `viewModeAllActivity`
   7. `viewModeSelect` → switch to `viewModeAllActivity`, reset logScrollOffset
   8. `viewModeAllActivity` → no-op (return)

   The current "save + world select" block is REMOVED from esc entirely.

3. **`q` key change** (update.go `handleKey`, phasePlaying). Split current `"q", "ctrl+c"` case:
   - `q` → save game, navigate to world select (reuses the cleanup block from the old esc fallthrough: reset viewMode, clear panels, reload world list, etc.)
   - `ctrl+c` → save + tea.Quit (standard terminal kill, unchanged)

4. **`l` key** (update.go `handleKey`, phasePlaying). New case:
   - If `viewModeSelect` and any subpanel is open → close all subpanels, reset logScrollOffset (shows action log)

5. **Post-order auto-back** (update.go `applyOrdersConfirm`). In the `ordersAddStep == 2` / `step2ActivityID == "plant"` branch: change `m.ordersAddStep = 0` to `m.ordersAddStep = 1` and remove `m.selectedActivityIndex = 0` reset (stay on the Gardening category). tillSoil already does step 1 — this makes plant match.

6. **`L: Log` hint** (view.go `renderDetails`). Line 858 currently shows `P: Preferences  K: Knowledge  I: Inventory`. When any subpanel is open (`showKnowledgePanel || showInventoryPanel || showPreferencesPanel`), prepend `L: Log` to the hint line.

7. **Status bar hints** (view.go status bar section). Replace the current conditional `ESC=menu` hint with always-visible context-dependent hints:
   - `q=menu` — always shown
   - Esc hint based on state:
     - `ordersFullScreen` or `activityFullScreen` → `ESC=collapse`
     - orders add/cancel mode → `ESC=back` (existing: already hidden, now show it)
     - subpanel open → `ESC=back`
     - `showOrdersPanel` normal → `ESC=back`
     - `viewModeSelect` → `ESC=back`
     - `viewModeAllActivity` (nothing expanded) → no esc hint

8. Verify all existing Esc behavior within orders add/cancel mode still works (those cases fire before the new ones in the priority order).

[TEST] Walk through each scenario from the desired behavior list:
- Esc from expanded all-activity → collapses
- Esc from expanded orders → collapses
- Esc from all-activity (not expanded) → nothing happens
- Esc from orders normal → goes to all-activity
- Esc from orders add/cancel → back one level (unchanged)
- Esc from select with action log → goes to all-activity
- Esc from select with knowledge/inventory/preferences panel → closes panel, shows action log
- `l` from select with any subpanel → closes panel, shows action log
- `q` from game → saves, goes to world select
- `q` from world select → quits
- `ctrl+c` → quits (unchanged)
- Create plant order at step 2 → returns to step 1 (Gardening sub-category)
- `L: Log` hint visible when subpanel open, hidden when not
- Status bar hints update contextually

**Step 3:** [DOCS] ✅ + [RETRO]

---

### 2C. Order Selection UX ✅

**Player impact:** Order lists show numbered items. Player can press a number key to instantly select an item, in addition to the existing arrow-key scrolling.

**Reqs reconciliation:** Post-gardening-cleanup.md section C — "Replace with single keypress selection (numbered list)." Refined during discussion: numbers are added alongside scrolling, not replacing it. Applies to every step within orders.

**Architecture alignment:** MVU pattern — view layer adds number display, update layer adds key handling. No new architecture patterns.

**Values alignment:** Reuse Before Invention — extends existing list rendering and selection logic rather than replacing it.

**Resolved questions:**
- More than 9 items: deferred. Current max is ~4 items per step, well within single-digit range.

**Implementation:**

**Step 1: Numbered order selection**

Player sees numbered items in all order list contexts and can press a number key to select instantly.

1. **Tests first:** UI rendering/key handling — no unit tests per CLAUDE.md testing policy. Verified via human testing.
2. Add number display (e.g., `1. Harvest`, `2. Craft`) to order list rendering in `renderOrdersContent()` for all contexts:
   - Step 0: orderable activity list
   - Step 1: target item type / category list
   - Step 2: plant variety list
   - Cancel mode: existing order list
3. Add number key handling (1-9) in the order key handler that sets the corresponding selection index. Pressing a number both selects and confirms (equivalent to arrow-to-item + Enter), since the point is single-keypress selection.

**Design note:** Number keys select-and-confirm in one press (otherwise it's just a different way to scroll, not a UX improvement). If an invalid number is pressed (e.g., 5 when there are only 3 items), it's a no-op.

[TEST] Open orders. Verify numbered display in each step. Verify pressing a number key selects and advances. Verify arrow+Enter still works.

**Step 2:** [DOCS] + [RETRO]



---
### 2D. Till Soil Color Improvement ✅

**Player impact:** Three visually distinct states during tilling: teal highlight for active rectangle selection, sage for confirmed-but-not-yet-tilled plots, and dusky earth for tilled soil. Currently the active selection and confirmed plots both use olive green, making it hard to distinguish "still selecting" from "already marked."

**Reqs reconciliation:** Post-gardening-cleanup.md doesn't have a separate section for this — it was added directly to the plan. Requirement: "highlight (teal) color in select mode, tilled color (dusky earth) after plot confirmed." Refined during discussion: three distinct states instead of two — sage (growing style, color 108) for marked-for-tilling instead of dusky earth, so "pending work" looks different from "done."

**Architecture alignment:** MVU pattern — rendering changes only, in `view.go` and `styles.go`. No update logic changes.

**Values alignment:** Consistency Over Local Cleverness — visual progression (teal → sage → earth) maps to workflow progression (selecting → pending → done).

**Implementation:**

**Step 1: Update area selection and marked-for-tilling colors**

Player sees teal highlight while dragging a selection rectangle, sage for confirmed marked-for-tilling tiles, and dusky earth for already-tilled soil.

1. **Tests first:** UI rendering — no unit tests per CLAUDE.md testing policy. Verified via human testing.
2. **`styles.go`**: Change `areaSelectStyle` background from color 58 (olive) to a teal (e.g., color 30 or 38). Add `markedForTillingStyle` with sage background (color 108, matching `growingStyle` foreground). Leave `areaUnselectStyle` (dark red, color 52), `tilledStyle` (color 138), and `wetTilledStyle` (color 94) unchanged.
3. **`view.go` `renderCell()`**: Where marked-for-tilling tiles are rendered with pre-highlight (lines ~611-623, the section that shows already-marked tiles when no anchor is set), use `markedForTillingStyle` instead of `areaSelectStyle`. The active rectangle selection (inside the anchor-set block) continues using the updated `areaSelectStyle` (now teal).

[TEST] Enter area selection for till soil. Verify teal highlight during rectangle drag. Confirm a plot. Verify sage background on marked tiles. Have a worker till some tiles. Verify dusky earth on tilled tiles. Toggle to unmark mode — verify dark red still works.

**Step 2:** [DOCS] + [RETRO]

---

## Part 3: Small Features

### 3A. Gather Orders ✅

**Player impact:** Players can direct characters to gather loose items (sticks, nuts, shells, seeds) via the orders menu. "Gather" appears alongside Harvest, Craft, and Garden at the top level. Player selects an item type from a dynamic list of what's currently on the ground. Characters seek and pick up items of that type, filling vessels for seeds, nuts, and shells (supporting planting and snacking workflows), while sticks go directly to hand (bundling deferred to Construction). Anchor story: "Player gathers gourd seeds into a vessel for easy planting later."

**Reqs reconciliation:** Post-gardening-cleanup.md doesn't have a separate Gather section — originated from plan discussion. Refined design decisions below.

**Resolved design questions:**
- **One activity with sub-selection** (like Harvest, not like Craft's category grouping). Single "gather" activity in the registry. `Order.TargetType` stores the item type to gather. The item type selection is part of the order creation flow (step 1), same as Harvest.
- **Completion condition:** Follow Harvest model exactly — continue until no more items of target type on ground OR inventory (including vessel space) is full. Only ground items count (not other characters' inventories or future storage).
- **Available items list:** Dynamic scan of world items, filtered to gatherable types (`Plant == nil && Container == nil && ItemType != "hoe"`). This captures sticks, nuts, shells, and seeds while excluding growing plants, dropped plant items, crafted items, and tools. Consistent with the world not "knowing" things outside of it.
- **Vessel support:** Seeds, nuts, and shells use vessel procurement (harvest-identical via `EnsureHasVesselFor`). Sticks go to hand only — no stick varieties registered, so `AddToVessel` naturally fails and items go to inventory. `findGatherIntent` adds a variety check: if target item has no registered variety AND character has no inventory space, return nil (prevents stuck-loop where `CanVesselAccept` returns true for empty vessels but `AddToVessel` fails).
- **Discovery:** `AvailabilityDefault` — no discovery needed. Characters know how to pick things up.
- **No category:** Top-level orderable activity (empty Category field), same shape as Harvest. If Construction adds more gather-like activities, promote to a category at that point.

**Architecture patterns:**
- **Activity Registry** (architecture.md: Activity Registry & Know-How Discovery): New entry with `IntentOrderable`, `AvailabilityDefault`, no Category.
- **Adding an Ordered Action** (architecture.md): `findGatherIntent()` in order_execution.go, wire into `findOrderIntent` switch + `IsOrderFeasible`. NOT multi-step — completion via continuation chaining in ActionPickup handler (same as Harvest).
- **Item Acquisition** (architecture.md): `findNearestItemByType` with `growingOnly=false`. `EnsureHasVesselFor` for items with registered varieties. Conditional inventory check for items without varieties (sticks).
- **`continueIntent`**: Gather targets items on the map throughout — generic path handles it, no early-return block needed.
- **No new entity fields** — no serialization concern. Reuses `Order.TargetType` and existing `ActionPickup`.
- **No new ActionType** — gather uses `ActionPickup` (same physical action as Harvest).

**Prerequisite: nut variety registration.** Nuts are currently created by `NewNut()` with fixed Color=Brown but have no variety in the registry — `AddToVessel` fails without a variety. Add "nut" to `GetItemTypeConfigs()` with `NonPlantSpawned: true` (prevents double-spawning) so a nut variety is generated. Seeds and shells already have varieties. Also add `"shell": 4` to `config.StackSize` (currently defaults to 1, which is not useful for gathering).

**Implementation:**

**Step 1: Nut variety registration + shell stack size ✅**

Prerequisite for vessel support during gathering. Nuts get a variety in the registry so they can be stacked in vessels. Shells get an explicit stack size of 4.

1. **Tests first:** Verify nut variety is generated by `GenerateVarieties()`. Verify `AddToVessel` succeeds for nuts (currently would fail). Verify shell stack size is 4.
2. Add "nut" to `GetItemTypeConfigs()` in `variety_generation.go`: `Colors: []types.Color{types.ColorBrown}`, `Edible: true`, `Sym: config.CharNut`, `SpawnCount: config.GetGroundSpawnCount("nut")`, `NonPlantSpawned: true`.
3. Add `"shell": 4` to `config.StackSize` map.

[TEST] No human testing — pure logic. Verified via unit tests.

**Step 2: Gather order execution ✅**

Character assigned a gather order seeks and picks up items of the target type. Seeds, nuts, and shells fill vessels; sticks go to hand. Order completes when no more items remain or inventory is full.

**Progress:** All tests written in `order_execution_test.go` and `variety_generation_test.go`. Tests compile-fail as expected (functions don't exist yet). Next: implement activity registry entry, `findGatherIntent`, `FindNextGatherTarget`, `IsOrderFeasible` gather case, `GetGatherableTypes`, then ActionPickup handler extension in `update.go`.

1. **Tests first:** Test `findGatherIntent` returns pickup intent for nearest gatherable item. Test vessel procurement for seeds/nuts (items with varieties). Test sticks skip vessel procurement and check inventory space. Test `FindNextGatherTarget` returns nil when inventory full. Test `IsOrderFeasible` for gather orders. Test `GetGatherableTypes` returns correct dynamic list.
2. **Activity registry** (`entity/activity.go`): Add "gather" entry — `IntentOrderable`, `AvailabilityDefault`, no Category, no DiscoveryTriggers.
3. **`findGatherIntent`** (`system/order_execution.go`): Follows `findHarvestIntent` with two differences:
   - `findNearestItemByType(pos.X, pos.Y, items, order.TargetType, false)` — `growingOnly=false` instead of `true`.
   - Before `EnsureHasVesselFor`: check if target item has a registered variety via `registry.GetByAttributes(target.ItemType, target.Color, target.Pattern, target.Texture)`. If variety exists → call `EnsureHasVesselFor` (harvest-identical). If no variety → check `char.HasInventorySpace()`, return nil if full.
   - Everything else identical to `findHarvestIntent`: vessel pickup, movement, ActionPickup intent.
4. **Wire into `findOrderIntent`** switch: add `case "gather": return findGatherIntent(...)`.
5. **`IsOrderFeasible`**: add gather case — check if any item of `order.TargetType` exists on the ground (simple type match scan, same granularity as harvest's `growingItemExists` but without the growing filter).
6. **`FindNextGatherTarget`** (`system/order_execution.go`): Same shape as `FindNextHarvestTarget` — check `HasInventorySpace`, find nearest by type with `growingOnly=false`. Returns nil when full or no more items.
7. **ActionPickup handler extension** (`ui/update.go`): In the `PickupToVessel` continuation block, add `order.ActivityID == "gather"` alongside `"harvest"` for `FindNextVesselTarget` chaining and `CompleteOrder`. In the `PickupToInventory` continuation block, add `order.ActivityID == "gather"` alongside `"harvest"` for `FindNextGatherTarget` chaining and `CompleteOrder`.
8. **`GetGatherableTypes`** (`game/variety_generation.go`): Scans items on the map, filters to `Plant == nil && Container == nil && ItemType != "hoe"`, returns deduplicated list of `{DisplayName, TargetType}` sorted alphabetically. Same shape as `GetPlantableTypes`.

[TEST] No human testing yet — unit tests only. UI wiring in Step 3.

**Step 3: Order UI integration + human testing ✅**

Player sees "Gather" in the orders menu and can select from available gatherable item types.

1. **Tests first:** UI — no unit tests per CLAUDE.md testing policy.
2. **Step 0 (activity list)**: "gather" activity already appears via the standard `getOrderableActivities()` flow since it's `IntentOrderable` + `AvailabilityDefault`.
3. **Step 1 (target selection)**: When selected activity is "gather", populate the target list with `GetGatherableTypes(m.gameMap.Items())` instead of the harvest/plant-specific lists. Wire into the same step 1 rendering and selection logic.
4. **Order creation**: Set `order.TargetType` from the selected gatherable type. Standard `CreateOrder` flow.

[TEST] Create a test world with sticks, nuts, shells, and seeds scattered on the ground, plus a few empty vessels. Issue gather orders for each type:
- Gather seeds → character finds vessel, gathers seeds into it. Verify seeds stack in vessel.
- Gather nuts → character finds vessel, gathers nuts into it. Verify nuts stack in vessel.
- Gather shells → character finds vessel, gathers shells into it (up to 4 per vessel).
- Gather sticks → character picks up sticks to inventory only (no vessel seeking). Verify order completes when inventory full.
- Verify "Gather" appears in orders menu. Verify only types present on the ground show up.
- Verify order completes when no more items of target type remain.

**Step 3.5: Regression tests for bugs found during human testing ✅**

Two flow-level anchor tests chain the system functions in the sequence the UI handler calls them, exercising the full gather order lifecycle:
- **`TestGatherOrder_VesselPath_EndToEnd`** — Full inventory → drop item → pick up vessel → gather nuts into vessel → continuation → completion. Covers bugs #1 (HasItemOnMap pointer identity), #2 (EnsureHasVesselFor drop), #5 (vessel prerequisite guard).
- **`TestGatherOrder_InventoryPath_EndToEnd`** — Empty inventory → pick up stick → continuation → second stick → inventory full → completion. Covers bug #4 (vessel guard doesn't block stick continuation).

Additional targeted tests:
- **`TestHasItemOnMap_TwoItemsSamePosition`** (`map_test.go`) — Two items at same position, pointer identity works where `ItemAt` would return wrong item (bug #1).
- **`TestFindNextGatherTarget_FindsNextStickWithVesselInInventory`** (`order_execution_test.go`) — Validates the system function works for character carrying vessel + gathering sticks (bug #4).

Previously covered: Bug #2 by `TestEnsureHasVesselFor_DropsItemWhenFullOnOrder`, Bug #3 by `TestFindNextVesselTarget_FindsNonGrowingWhenNotGrowingOnly`.

**Note:** The UI-layer branching in update.go's ActionPickup handler (which function to call after pickup, and when to skip continuation for vessel prerequisites) is verified by human testing. When Construction adds more ordered actions, consider making the simulation package order-aware for automated e2e coverage of the full handler.

**Step 4:** [DOCS] + [RETRO]

### 3B. Satiation & Consumption Duration

**Player impact:** Eating a feast takes longer than eating a snack. Berries and nuts are a quick ~5-minute nibble. Mushrooms are a proper ~15-minute meal. Gourds are a lengthy ~45-minute feast. Currently all eating takes a fixed ~10 world minutes regardless of food type.

**Reqs reconciliation:** Post-gardening-cleanup.md section H — "Satiation tier of food should modify the amount of time it takes to consume. Meal = 15 world mins; Snack 1/3 that time; Feast 3x that time." All three tiers addressed with exact specified ratios.

**Architecture alignment:** Config-driven behavior — extends the existing satiation tier concept by making it a first-class struct with both satiation and duration properties. The `ActionConsume` handler in update.go (line ~991) accumulates `ActionProgress` against a threshold; this changes the threshold from fixed (`ActionDurationShort`) to tier-resolved. No new action types, no new entity fields, no serialization changes.

**Values alignment:**
- **Follow the Existing Shape** — satiation tiers already exist as constants (`SatiationFeast`/`SatiationMeal`/`SatiationSnack`) with an item→value map. This restructures them into an explicit tier struct, making the implicit concept explicit.
- **Source of Truth Clarity** — tier properties (satiation amount, duration) live on the tier definition, not scattered across parallel maps or derived from multipliers.
- **Start With the Simpler Rule** — flat duration per tier, no character-specific modifiers or hunger-level scaling.

**Resolved questions:**
- **Drinking unchanged.** Requirement says "satiation tier of food." Water has no satiation tiers. `ActionDrink` continues using `ActionDurationShort`.
- **Vessel eating.** When eating from a vessel, `char.Intent.TargetItem` is the vessel, not the food. Duration lookup resolves the actual food type from vessel contents (`Container.Contents[0].Variety.ItemType`).

**Opportunistic assessment:** After playtesting 3B, evaluate whether satiation-aware targeting (characters preferring higher-satiation food when hungrier, snacking on nearby food at lower hunger) should be tackled now or deferred. See triggered-enhancements.md for full description.

**Implementation:**

**Step 1: Restructure satiation config to tier-based struct ✅**

No player-visible change — satiation values and eating behavior remain identical. This restructures the config to support duration in Step 2.

1. **Tests first:**
   - Test that looking up satiation for each item type returns the same values as today (gourd=50, mushroom=25, berry=10, nut=10).
   - Test that looking up duration for each item type returns the correct values (gourd=3.75, mushroom=1.25, berry=0.417, nut=0.417 game seconds).
   - Test that an unknown item type falls back gracefully (returns a default — use `ActionDurationShort` and 0 satiation, or whatever the current implicit behavior is for unlisted types).

2. **Config restructure** (`config/config.go`):
   - Define `MealSize` struct: `type MealSize struct { Satiation float64; Duration float64 }`.
   - Define tier constants as `MealSize` values:
     ```
     MealSizeFeast = MealSize{Satiation: 50.0, Duration: 3.75}  // ~45 world mins
     MealSizeMeal  = MealSize{Satiation: 25.0, Duration: 1.25}  // ~15 world mins
     MealSizeSnack = MealSize{Satiation: 10.0, Duration: 0.417} // ~5 world mins
     ```
   - Replace `SatiationTier map[string]float64` with `ItemMealSize map[string]MealSize`:
     ```
     "gourd" → MealSizeFeast, "mushroom" → MealSizeMeal,
     "berry" → MealSizeSnack, "nut" → MealSizeSnack
     ```
   - Add helper: `GetMealSize(itemType string) MealSize` — returns the tier for an item type, with a sensible default for unknown types (Meal tier as middle ground).
   - Remove the old `SatiationFeast`/`SatiationMeal`/`SatiationSnack` float constants and `SatiationTier` map.

3. **Update callsites** (`system/consumption.go`): Everywhere that currently reads `config.SatiationTier[itemType]` for the satiation value, change to `config.GetMealSize(itemType).Satiation`. The resolved values are identical — this is a mechanical refactor.

   **Anti-pattern:** Do NOT wire duration into the handler yet. Step 1 only restructures config; Step 2 adds the behavior change. This keeps the refactor verifiable in isolation.

No human testing — pure config refactor. Verified via unit tests.

**Step 2: Tier-based consumption duration ✅**

A character eats a berry in ~5 world minutes (quick snack), a mushroom in ~15 minutes (proper meal), and a gourd in ~45 minutes (lengthy feast). The player can see the difference watching characters eat different foods.

1. **Tests first:**
   - Test that eating a snack-tier item completes after accumulating 0.417 game seconds of progress.
   - Test that eating a feast-tier item does NOT complete after 1.25 game seconds but DOES complete after 3.75 game seconds.
   - Test eating from a vessel uses the vessel contents' tier, not a default.

2. **Resolve eaten item type** (`ui/update.go`): Add a small helper `getEatenItemType(item *entity.Item) string` used by the `ActionConsume` handler:
   - If `item.Container != nil && len(item.Container.Contents) > 0` → return `item.Container.Contents[0].Variety.ItemType` (vessel eating: food type from contents).
   - Otherwise → return `item.ItemType` (loose item eating).

3. **ActionConsume handler** (`ui/update.go`, ~line 991): Replace the fixed `config.ActionDurationShort` threshold:
   - Before accumulating progress, resolve the duration: `duration := config.GetMealSize(getEatenItemType(char.Intent.TargetItem)).Duration`.
   - Change `if char.ActionProgress >= config.ActionDurationShort` to `if char.ActionProgress >= duration`.
   - Everything else in the handler stays the same — consumption functions, vessel/inventory branching, side effects.

   **Note:** `ActionDurationShort` is still used by `ActionDrink`, `ActionPickup`, and other actions. Only `ActionConsume` changes.

[TEST] Start a game with varied food. Observe:
- Berry/nut eating completes noticeably fast (~5 world minutes).
- Mushroom eating takes a moderate amount of time (~15 world minutes).
- Gourd eating takes significantly longer (~45 world minutes).
- Eating from a vessel with berries is fast; vessel with mushrooms is moderate.
- Drinking speed is unchanged.

**Step 3:** [DOCS] ✅ + [RETRO]

### 3C. Satiation-Aware Food Selection ✅

**Anchor story:** An unprovisioned worker is tilling fields all day. He only pulls himself away when hunger reaches Severe — time to walk home for a hearty filling meal. He passes the nearby berry bush without stopping; the gourd waiting 20 tiles away is worth the walk. But if a mushroom is right there on the path (5-6 tiles), that's a sensible enough meal to stop for. And if things get truly desperate — Crisis hunger — he grabs whatever is closest just to survive. When foraging proactively at low hunger, characters stock up on snack-sized food (berries, mushrooms) rather than gourds they can't practically eat during a workday.

**Player impact:** Characters choose food that fits their current hunger level. At moderate hunger, they eat what they like nearby. At severe hunger, they bypass snacks and seek filling meals worth the walk — unless a decent meal is right on the path. At crisis, nearest food wins. When idle-foraging at low hunger, they prefer snacks over feasts they don't need.

**Reqs reconciliation:** Post-gardening-cleanup.md section 3C / triggered-enhancements.md "satiation-aware targeting." The snacking threshold (characters grabbing nearby food at lower hunger before reaching Moderate) is a future enhancement — evaluate after playtesting 3C. This feature lays the groundwork by incorporating satiation fit into scoring, which naturally supports lower hunger thresholds if added later.

**Architecture alignment:** Modifies two existing scoring paths — `FindFoodTarget` in movement.go (hunger-driven food seeking) and `scoreForageItems` in foraging.go (idle foraging). Both already use `(pref × PrefWeight) - distance` scoring. Adds two terms: raw fit delta as a negative modifier, and per-tier distance weight. New config constants follow the existing per-tier weight pattern (`FoodSeekPrefWeight*`). No new functions, no new entity fields, no new action types.

**Values alignment:**
- **Anchor to Intent** — formula designed and validated against the worker story, not abstract scoring math.
- **Follow the Existing Shape** — `FoodSeekDistWeight*` per tier sits alongside existing `FoodSeekPrefWeight*` per tier, same config pattern.
- **Start With the Simpler Rule** — raw delta with no weight coefficient. Tuning surface is DistWeight per tier only.

**Resolved questions:**
- **How does satiation factor into scoring?** Not raw satiation (which always favors the biggest meal), but *fit*: `|hunger - satiation|`. A food that matches the character's current hunger scores best. Too small = leftover hunger; too big = wasted satiation. This naturally scales with hunger level without needing per-tier fit weights.
- **Distance weight scaling creates the behavioral arc.** At Moderate hunger, distance is cheap (DistW=1) — preference and fit compete freely. At Severe, distance is moderate (DistW=1.5) — fit can overcome distance for filling food, but only for reasonable walks. At Crisis, distance is expensive (DistW=3) — nearest food wins regardless of fit.
- **Order-context consumption?** No. Workers at Mild eat from carried inventory with no scoring. Workers at Moderate+ pause and go through `FindFoodTarget`, which picks up fit scoring naturally.
- **Idle foraging?** Yes. `scoreForageItems` in foraging.go gets the same fit delta. At low hunger (sub-50), characters prefer snack-sized food: berry (delta 10-20) over gourd (delta 30-50). This means workers forage berries for the field, not gourds they'd waste 45 minutes eating. Vessel scoring (`scoreForageVessels`) is unchanged — that's about container decisions, not food-fit decisions.

**Formula:**
```
score = (pref × PrefWeight) - (dist × DistWeight) - |hunger - satiation| + healingBonus
```

**Starting weight values (tune after playtesting):**

| Tier | PrefWeight (existing) | DistWeight (new) |
|------|----------------------|-----------------|
| None (idle) | 20 | 1.0 |
| Moderate | 20 | 1.0 |
| Severe | 5 | 1.5 |
| Crisis | 0 | 3.0 |

**Reference scenarios (berry pref +2, mushroom pref +1, gourd neutral):**

| Hunger | Berry (d=2) | Mushroom (d=10) | Gourd (d=20) | Winner |
|--------|-------------|-----------------|--------------|--------|
| 30 (idle) | (40)-2-20 = 18 | (20)-10-5 = 5 | (0)-20-20 = -40 | Berry |
| 55 (Moderate) | (40)-2-45 = -7 | (20)-10-30 = -20 | (0)-20-5 = -25 | Berry |
| 75 (Severe) | (10)-3-65 = -58 | (5)-15-50 = -60 | (0)-30-25 = -55 | Gourd (by 3) |
| 75 + mush d=5 | (10)-3-65 = -58 | (5)-7.5-50 = -52.5 | (0)-30-25 = -55 | Mushroom |
| 95 (Crisis) | (0)-6-85 = -91 | (0)-30-70 = -100 | (0)-60-45 = -105 | Berry |

**Dependency:** Uses `GetMealSize(itemType).Satiation` from 3B's restructured config.

**Implementation:**

**Step 1: Satiation fit scoring ✅**

The worker at Severe hunger walks past a nearby berry to reach a filling gourd — unless a decent mushroom is right on the path. At Crisis, the nearest food wins. When foraging at low hunger, characters prefer snacks over feasts.

1. **Tests first:**
   - Test `FindFoodTarget`: character at Severe hunger (75), neutral preferences, berry at dist 3, gourd at dist 20 → chooses gourd (fit overcomes distance).
   - Test `FindFoodTarget`: same layout at Crisis hunger (95) → chooses berry (distance overcomes fit).
   - Test `FindFoodTarget`: Severe hunger, mushroom at dist 5, berry at 2, gourd at 20 → chooses mushroom (right-sized meal on the path).
   - Test `scoreForageItems`: character at hunger 20, berry and gourd equidistant → prefers berry (better fit at low hunger).
   - Test that inventory items (distance 0) use fit as the deciding factor: character carrying berry and gourd at hunger 75 → prefers gourd.

2. **Config constants** (`config/config.go`): Add per-tier distance weights alongside existing preference weights:
   ```
   FoodSeekDistWeightModerate = 1.0
   FoodSeekDistWeightSevere   = 1.5
   FoodSeekDistWeightCrisis   = 3.0
   ```

3. **Shared scoring helper** (`system/food_scoring.go`): Extract a `ScoreFoodFit(netPref int, dist int, hunger float64, itemType string, prefWeight, distWeight float64) float64` function that computes `(pref × prefWeight) - (dist × distWeight) - |hunger - satiation|`. Both `FindFoodTarget` and `scoreForageItems` call this helper for the core formula, keeping scoring math as a single source of truth. Each callsite still owns its own candidate filtering, iteration, vessel logic, and healing bonus.

   **Design decision:** The formula has three terms that must stay in sync across hunger-driven and idle-foraging paths. Per Values.md "Don't duplicate logic across call sites" — a shared helper prevents future tuning changes (e.g., adding a fit weight coefficient) from needing updates in two places. Foraging naturally gets Moderate/1.0 weights since it only fires at TierNone.

4. **`FindFoodTarget`** (`system/movement.go`): Resolve per-tier `distWeight` alongside `prefWeight`. Replace inline scoring with `ScoreFoodFit()`. Vessel-contents paths resolve `itemType` from variety. Healing bonus still applied after the shared score.

5. **`scoreForageItems`** (`system/foraging.go`): Replace inline scoring with `ScoreFoodFit()` using `FoodSeekPrefWeightModerate` and `FoodSeekDistWeightModerate` (idle = TierNone weights).

   **Scope note:** `scoreForageVessels` is unchanged — vessel scoring is about container decisions (vessel bonus), not food-fit decisions.

[TEST] Start a world with varied food at different distances from a tilling worker:
- Worker at Severe hunger bypasses nearby berries, heads for a distant gourd.
- If a mushroom is on the path (5-6 tiles), worker stops for it instead.
- At Crisis hunger, worker grabs nearest food regardless.
- Idle characters forage berries/mushrooms over gourds when not very hungry.
- Characters with strong food preferences still eat preferred food at Moderate hunger (preference wins when not desperate).

**Step 2:** [DOCS] ✅ + [RETRO]

---

## Part 4: Character Creation Streamline

**Player impact:** Cleaner game start — no single/multi mode choice. R for instant random community (4 characters), C for a creation grid where you can customize any number of characters (1-16) with +/- to add/remove. All characters start randomized and are equally editable — there's no protagonist, just a community you're setting up.

**Reqs reconciliation:** Post-gardening-cleanup.md section G — "Remove single char mode, adjust start screen to R/C, add character count." Refined during discussion: all characters are equal (no protagonist/companion model), count merged into the creation grid via +/- (no separate count step), no summary screen (creation grid → game directly).

**Architecture alignment:** MVU pattern — phase handlers in update.go, views in view.go, creation state in creation.go. No new architecture patterns. Extends the existing CharacterCreationState by making it variable-length.

**Values alignment:**
- Follow the Existing Shape — extends creation.go's card-per-character pattern rather than replacing it.
- Consider Extensibility — slice-based character list supports future count changes.
- Start With the Simpler Rule — +/- on the grid is simpler than a separate count step.

**Dead code to remove:** `phaseSelectFood`, `phaseSelectColor`, `startGame()`, `startGameMulti()`, `viewFoodSelect()`, `viewColorSelect()`, model fields `multiCharMode`, `selectedFood`, `selectedColor`. All self-contained in ui/ — no external references.

**Deferred:** Character count does not affect world generation (map size, item/feature spawning) yet. Revisit if playtesting reveals scaling needs.

**Implementation:**

### 4A. Clean Up Dead Code + R/C Start Screen ✅

**Anchor story:** Player selects "New World" and sees "R: Random Characters / C: Create Characters" instead of the old Single/Multi mode screen. Pressing R drops them straight into a world with 4 random characters. Pressing C takes them to the creation screen.

1. **Tests first:** UI-only changes — no unit tests per CLAUDE.md testing policy.

2. **Remove dead phases** (model.go): Delete `phaseSelectFood` and `phaseSelectColor` from the phase enum. `phasePlaying` renumbers via iota (internal only, not serialized).

3. **Remove dead model fields** (model.go): Delete `multiCharMode`, `selectedFood`, `selectedColor`.

4. **Remove dead functions and handlers** (update.go):
   - Delete `phaseSelectFood` case and `phaseSelectColor` case from `handleKey`.
   - Delete `startGame()` function (single-char start).
   - Delete `startGameMulti()` function (unused random start).

5. **Remove dead views** (view.go): Delete `viewFoodSelect()` and `viewColorSelect()`.

6. **Repurpose phaseSelectMode handler** (update.go): Replace current handler with:
   - `"r"/"R"` → `startGameRandom()` (new function) + `tickCmd`
   - `"c"/"C"` → `m.creationState = NewCharacterCreationState()` + `m.phase = phaseCharacterCreate`
   - `"esc"` → `m.phase = phaseWorldSelect`
   - Remove "1" single-char debug path, remove "m"/"M" multi-char path.

7. **New `startGameRandom()`** (update.go): Generates 4 random characters and starts the game. Extract logic from the deleted `startGameMulti()`:
   - Generate 4 unique random names via `randomUniqueNames(4)`
   - Random food/color for each
   - Clustered placement at offsets `{0,0}, {2,0}, {0,2}, {2,2}` around map center
   - Spawn world (ponds, features, items)
   - Same initialization as `startGameFromCreation()` (world ID, cursor, following)

8. **Repurpose viewModeSelect** (view.go): Render the R/C screen:
   ```
   === Petri ===

   R  Random Characters
   C  Create Characters

   Esc: Back
   ```

[TEST] Start a new world via R — 4 random characters appear, game plays normally. Start via C — creation screen works as before. Load an existing world — still works. Esc from R/C screen goes back to world select.

**Step 2:** [DOCS] ✅ + [RETRO]

---

### 4B. Variable-Count Creation Grid ✅

**Anchor story:** Player presses C, sees 4 randomized character cards in a row. They press + three times — three more cards appear in a second row. They edit a couple of names, randomize the rest with Ctrl+R, and press Enter to start with 7 characters.

1. **Tests first:** Test that `NewCharacterCreationState()` creates 4 characters with unique names (default count). Test that `AddCharacter()` adds a card (up to 16) with a unique name and `RemoveLastCharacter()` removes the last card (min 1, adjusts SelectedChar if needed). Test that `NavigateCharacter` wraps correctly at variable lengths.

2. **CharacterCreationState restructure** (creation.go):
   - `Characters` from `[4]CharacterCreationData` to `[]CharacterCreationData`.
   - `NewCharacterCreationState()` takes no arguments, defaults to 4 characters.
   - New `AddCharacter()`: appends a new random card with a unique name. Returns false if at max (16).
   - New `RemoveLastCharacter()`: removes last card. Adjusts `SelectedChar` if it pointed at the removed card. Returns false if at min (1).
   - `NavigateCharacter`: modulo `len(s.Characters)` instead of `4`.
   - `RandomizeAll`: regenerates `len(s.Characters)` characters.

3. **Key handling** (update.go `handleCharacterCreationKey`): Add `"+"/"="` → `AddCharacter()`, `"-"` → `RemoveLastCharacter()`.

4. **View** (view.go `viewCharacterCreate`):
   - Render cards in rows of 4 using `lipgloss.JoinHorizontal` per row.
   - Show count + "+/- to add/remove" header line.
   - Each card shows `#N`, name, food, color with selected card highlighted (cyan border) and selected field highlighted within the active card.
   - Navigation hints include +/- controls.

5. **Dynamic character placement** (update.go `startGameFromCreation`):
   - Generate offsets dynamically: `col = i % 4`, `row = i / 4`, `x = cx + col*2`, `y = cy + row*2`.
   - Iterate over `m.creationState.Characters` slice (no fixed-length assumption).

[TEST] Create characters with counts 1, 4, 8, and 16:
- Grid renders correctly at each size (1 card, 1 row of 4, 2 rows, 4 rows)
- + adds a card with unique name, - removes last card
- Cannot go below 1 or above 16
- Navigate across cards with left/right/tab, fields with up/down
- Edit names, cycle food/color on any card
- Ctrl+R randomizes all cards (preserving count)
- Enter starts game with correct number of characters at correct positions

**Step 2:** [DOCS] ✅ + [RETRO]

---

## Part 5: Game Mechanics Doc Reorg ✅

Reorganized game-mechanics.md:
- Reordered by gameflow (25 sections → 14)
- ~870 lines → ~330 lines — trimmed implementation details, merged related sections
- Grouped: Characters (stats/speed/sleep/frustration/mood), Food & Consumption (eating/drinking/scoring), Gardening (tilled soil/sprouts/watered tiles)
- Added Overview paragraph, Character Creation section, moved Character Names out of Preferences
- Removed: Intent Re-evaluation Guards, ConsumePlantable Helper, Preference Log Messages, Variety Generation formula, Mood Display colors
- Added game-mechanics.md as a reference in skills: required for `/new-phase`, optional for `/refine-feature` and `/test-world`

[DOCS] — this step *was* the docs update.

---

## Final [RETRO] on full cleanup phase ✅

Skipped — process reflections were captured incrementally after each feature.
