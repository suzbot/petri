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

Items refined and ready to implement, in order:
1. **2A** — Sticky BFS pathing (meatiest behavior change, get in early)
2. **2C** — Numbered order selection (self-contained UI work)
3. **2B** — Esc key cleanup (investigation step may surface scope surprises, do last)

Refined and ready to implement:
4. **2D** — Till soil selection colors (small UI change)
5. **3A** — Gather orders (follows harvest pattern with vessel support)

Items needing refinement before implementation:
- **3B** — Satiation & consumption duration (minor — duration mechanism check)
- **3C** — Satiation-aware food selection (factor satiation value into hungry food-seeking)
- **4A/4B** — Character creation streamline (deferred questions about flow/scaling)

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

**Player impact:** Eating a feast takes longer than eating a snack. Meals take 15 world minutes; snacks 1/3 that; feasts 3x that.

Satiation tier of food modifies consumption duration. Touches config constants and consumption timing in update.go.

**Pattern reference:** Config-driven behavior — satiation tiers already defined in `config.SatiationTier` map. Consumption duration would follow the same pattern (config map keyed by item type or satiation tier). The `ActionConsume` handler in update.go manages eating duration via progress/speed accumulation.

**Deferred question:** Does duration apply to drinking too, or just eating? The requirement says "satiation tier of food" so likely just eating. Need to check current eating duration mechanism to understand where the multiplier hooks in.

**Opportunistic assessment:** After playtesting 3B, evaluate whether satiation-aware targeting (characters preferring higher-satiation food when hungrier, snacking on nearby food at lower hunger) should be tackled now or deferred. See triggered-enhancements.md for full description.

[TEST] Observe characters eating different satiation tiers. Verify feast items take noticeably longer than snacks.

### 3C. Satiation-Aware Food Selection

**Player impact:** When hungry, characters prefer more filling food — a nearby gourd over a distant berry, or even a slightly farther gourd over a very close berry. Currently foraging scores by distance + preference only, ignoring how much the food will actually help.

This is the first half of the "satiation-aware targeting & snacking threshold" item from triggered-enhancements.md. The snacking threshold (characters grabbing nearby food at lower hunger) is deferred — evaluate after playtesting 3C.

**Deferred questions:**
- How should satiation weight against distance and preference? A gourd (50 pts) 10 tiles away vs a berry (10 pts) 1 tile away — which wins? Need to examine the current scoring formula in `foraging.go` to understand the weight space.
- Does this apply to order-context consumption too (workers eating from inventory at Mild tier), or just idle foraging? Probably just idle foraging since order consumption uses carried items, not seeking.
- Should the weight scale with hunger level? (Starving character cares more about satiation value; mildly hungry character cares more about convenience.) Or keep it simple with a flat weight?

[TEST] Observe characters choosing between high-satiation and low-satiation food at varying distances.
[DOCS] after 3A-3C
[RETRO] after [DOCS]

---

## Part 4: Character Creation Streamline

**Player impact:** Cleaner game start — no single/multi mode choice. New ability to choose how many characters to start with (up to 16).

### 4A. Remove Single/Multi Mode

- Remove single character mode from game start
- Adjust start screen:
  ```
  === Petri ===
  R to start with Random Characters
  C to create Characters
  ```
- "C for Create" creates first character, randomizes the rest

### 4B. Character Count Selection

- Add intermediate step to choose number of characters (max 16)
- Requires refactoring the current character creation flow

**Pattern reference:** UI state machine in model.go/update.go. The character creation flow is a series of view states managed in the Update handler. No new entity fields — character count is a UI/setup concern, not persisted. No save/load impact (character count isn't stored; the resulting characters are what's saved).

**Deferred questions:**
- Does character count affect world generation (map size, item/feature spawning)? Currently these may be fixed — need to check whether spawn counts scale or are hardcoded.
- Where does the count selection step go in the flow? Before or after creating the first character? Likely before — pick count, then create one, then randomize the rest.
- Does "R for Random" also prompt for count, or does it use a default count?
- `phaseSelectFood` and `phaseSelectColor` are only reachable via the debug-only single-character flow (`-debug` flag + press "1"). When removing single/multi mode, clean up these dead view states and their handlers in update.go/view.go.

[TEST] Verify both R and C paths work. Verify character count selection. Verify created + randomized characters all spawn correctly.
[DOCS] after Part 4
[RETRO] after [DOCS]

---

## Part 5: Game Mechanics Doc Reorg

Reorganize game-mechanics.md for player readability:
- Reorder by gameflow
- Remove unnecessary detail or summarize where appropriate
- Consistent level of detail throughout

[DOCS] — this step *is* the docs update.

---

## Final [RETRO] on full cleanup phase
