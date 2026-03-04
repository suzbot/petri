# Step Spec: Step 3 — Clay Terrain + Dig Clay

Design doc: [construction-design.md](construction-design.md)

## Sub-step 3a: Clay Terrain, World Generation, and Loose Clay Items

**Anchor story:** The player starts a new world and notices dusky-earth patches with a halftone texture clustered near the pond — clay deposits. A few loose lumps of clay sit on the surface. The player moves the cursor onto a clay tile and sees "Clay deposit" in the details panel. The loose clay items show their normal item description.

### Tests

**Clay terrain on Map:**
- `TestIsClay_ClayTile` — `IsClay` returns true for a position after `SetClay`
- `TestIsClay_NonClayTile` — `IsClay` returns false for unset positions
- `TestFindNearestClay_FindsClosest` — returns the nearest clay tile to a given position
- `TestFindNearestClay_NoClay` — returns false when no clay tiles exist

**World generation:**
- `TestSpawnClay_CountInRange` — generates 6-10 clay tiles
- `TestSpawnClay_AdjacentToWater` — every clay tile is cardinal-adjacent to a water tile
- `TestSpawnClay_Clustered` — every clay tile is cardinal-adjacent to at least one other clay tile
- `TestSpawnClay_SpawnsLooseClayItems` — 2-3 clay items exist on clay tile positions after generation

**Serialization:**
- `TestClayPositions_RoundTrip` — clay positions survive save/load (follows `TestSetTilled_IsTilled` pattern)

**NewClay constructor:**
- `TestNewClay_Properties` — ItemType "clay", symbol `☗`, no Kind, no variety, no Plant, no Container, BundleCount 0

### Implementation

**Map terrain (game/map.go):**
- Add `clay map[types.Position]bool` field to Map struct
- Initialize in `NewMap`
- Add `SetClay(pos)`, `IsClay(pos) bool`, `ClayPositions() []Position`
- Add `FindNearestClay(pos) (Position, bool)` — finds nearest clay tile. Unlike `FindNearestWater` (which checks cardinal-adjacent availability because water is impassable), clay is passable so this just finds the nearest clay tile by distance. No adjacency check needed.
- Pattern: follows water terrain (`IsWater`, `WaterPositions`, `FindNearestWater`) and tilled terrain (`IsTilled`, `TilledPositions`)

**Passability (game/map.go):**
- Clay tiles are passable — no changes to `IsOccupied`, `IsBlocked`, or pathfinding. Characters walk on clay tiles.

**Clay item constructor (entity/item.go):**
- `NewClay(x, y int) *Item` — ItemType "clay", symbol `☗` (config.CharClay), Color types.ColorBrown (or a reddish-brown if one exists — check types.go for available colors; if not, use ColorBrown for now). No Kind, no Plant, no Container, no Edible, BundleCount 0.
- Add `CharClay = '☗'` to config.go

**World generation (game/world.go):**
- `SpawnClay(m *Map)` — called after `SpawnPonds` in the world generation sequence
- Algorithm (pair-placement + fill, NOT blob growth):
  1. Build candidate pool: all tiles cardinal-adjacent to at least one water tile (and not water themselves).
  2. Phase 1 — Pairs: Shuffle candidates. For each, find a cardinal neighbor also in the pool; place both as a pair. Repeat until target reached (6-10, randomized per `ClayMinCount`/`ClayMaxCount`).
  3. Phase 2 — Fill: If still below target, repeatedly scan remaining candidates for any that are both water-adjacent (in pool) and cardinal-adjacent to already-placed clay. Add them one at a time. Loop until target reached or no more candidates qualify. This grows pairs into clusters of 3+ and recovers tiles lost to suboptimal pairing order.
  4. If candidates exhausted before reaching 6, accept what was placed (rare small-pond edge case).
  - Pairs and fills are scattered independently — clay deposits are clusters near water, not one contiguous blob. Each tile satisfies both constraints (water-adjacent, has at least one clay neighbor).
  - Anti-pattern: do NOT use blob growth (`spawnPondBlob` pattern). Clay is a deposit, not a body.
- After placing clay terrain, spawn 2-3 `NewClay` items on randomly selected clay tile positions. These are loose items on the ground, not terrain.

**Rendering (ui/view.go):**
- Add `clayStyle` to styles.go — color 138 (same as tilledStyle / dusky earth)
- In `renderCell`: after the water check and before the tilled check, add clay terrain rendering. When `m.gameMap.IsClay(pos)` and no entity is on the tile, render `░` in `clayStyle` for both sym and fill.
- When an entity IS on a clay tile, use `clayStyle.Render("░")` as the fill (same pattern as tilled soil fill behind entities).

**Details panel (ui/view.go):**
- When cursor is on a clay tile, show "Clay deposit" annotation (similar to "On tilled soil" annotation). Use `clayStyle` for the text.

**Serialization (save/state.go, ui/serialize.go):**
- Add `ClayPositions []types.Position` to `SaveState` (json tag: `"clay_positions,omitempty"`)
- `ToSaveState`: iterate `gameMap.ClayPositions()` into the save state
- `FromSaveState`: iterate saved positions, call `gameMap.SetClay(pos)` for each
- Add "clay" case to the symbol restoration switch in `itemFromSave()` — restores `config.CharClay`

**Architecture pattern:** Follows water/tilled terrain precedent (**Follow the Existing Shape**). Map field + query methods + serialization as position list. No new entity types.

**Values:** **Isomorphism** — clay terrain represents clay deposits in the world. **Start With the Simpler Rule** — uniform clay, no varieties (DD-14). **Source of Truth Clarity** — the map is the source of truth for where clay exists.

**Reqs:** Construction-Reqs lines 43-47: "6-10 tiles, adjacent to water, adjacent to at least one other clay tile, passable, reddish brown pattern." Rendering uses dusky earth + halftone (user choice during refinement). Clustering and water adjacency enforced by generation algorithm.

### [TEST]

Build and run the game. Verify:
- Clay tiles visible near ponds (dusky earth `░` patches)
- 2-3 loose clay items (☗) visible on clay tiles
- Cursor on clay tile shows "Clay deposit" in details panel
- Cursor on loose clay item shows normal item description
- Characters can walk through clay tiles (passable)
- Save and reload preserves clay tile positions

Use `/test-world` if clay tiles are hard to observe naturally.

### [DOCS]
### [RETRO]

---

## Sub-step 3b: Dig Activity, Discovery, Order, and Handler

**Anchor story:** A character wanders near the clay deposits and looks at a loose lump of clay. They realize they could dig for more! The player opens the orders panel and sees "Dig Clay" as a new option. They create a Dig Clay order. The character drops the vessel they were carrying, walks to the nearest clay tile, spends a moment digging, and picks up a lump of clay. They walk to the next clay tile and dig again. With both hands full, the order is complete — they set both lumps down on the ground.

### Tests

**Activity registry and discovery:**
- `TestDigActivity_InRegistry` — "dig" activity exists with IntentOrderable, AvailabilityKnowHow
- `TestDigDiscovery_LookAtClay` — ActionLook on a clay item triggers "dig" know-how discovery
- `TestDigDiscovery_PickupClay` — ActionPickup on a clay item triggers "dig" know-how discovery

**findDigIntent:**
- `TestFindDigIntent_DropsInventoryFirst` — character carrying non-clay items drops them before returning dig intent
- `TestFindDigIntent_FindsNearestClayTile` — returns ActionDig intent targeting the nearest clay tile
- `TestFindDigIntent_NilWhenNoClay` — returns nil when no clay tiles exist on the map
- `TestFindDigIntent_NilWhenInventoryFullOfClay` — returns nil when both inventory slots have clay (triggers completion)

**Order completion:**
- `TestIsMultiStepOrderComplete_Dig` — returns true when both inventory slots have clay
- `TestIsMultiStepOrderComplete_Dig_OneSlot` — returns false when only one slot has clay
- `TestIsMultiStepOrderComplete_Dig_Empty` — returns false when inventory is empty

**Drop on completion:**
- `TestDropCompletedDigItems_DropsAllClay` — drops all clay items from inventory onto the map

**IsOrderFeasible:**
- `TestIsOrderFeasible_Dig_ClayExists` — returns (true, false) when clay tiles exist
- `TestIsOrderFeasible_Dig_NoClay` — returns (false, false) when no clay tiles exist

**Handler (applyDig):**
- `TestApplyDig_WalksToClayTile` — character moves toward clay tile when not yet there
- `TestApplyDig_CreatesClayInInventory` — after ActionDurationShort at clay tile, clay item appears in inventory
- `TestApplyDig_ClearsIntent` — intent is nil after dig completes (ordered action one-tick-idle pattern)

**End-to-end flow:**
- `TestDigOrder_EndToEnd` — chain the full flow: findDigIntent drops inventory → walks to clay → digs → walks to next clay → digs → findDigIntent returns nil → isMultiStepOrderComplete true → drops both clay → CompleteOrder. Follows `TestGatherOrder_InventoryPath_FullBundle_EndToEnd` pattern.

### Implementation

**Action constant (entity/character.go):**
- Add `ActionDig` to the ActionType enum with comment: `// Digging material from terrain (ordered, walk-then-act)`

**Activity registry (entity/activity.go):**
- Add "dig" entry: `ID: "dig"`, `IntentFormation: IntentOrderable`, `Availability: AvailabilityKnowHow`, `Category: ""` (top-level, no category)
- Discovery triggers:
  - `{Action: ActionLook, ItemType: "clay"}` — looking at clay
  - `{Action: ActionPickup, ItemType: "clay"}` — picking up clay
- Display name in order UI: "Dig Clay" (the activity's display handles this — check how other top-level activities like "Harvest" display)

**findDigIntent (system/order_execution.go):**
- New function: `findDigIntent(char, pos, items, order, log, gameMap) *Intent`
- Step 1: If both inventory slots have clay, return nil (triggers completion via `isMultiStepOrderComplete`)
- Step 2: Drop all non-clay inventory items. Iterate inventory, call `DropItem` for each non-clay item. This follows the procurement drop pattern from `EnsureHasRecipeInputs` (DD-16).
- Step 3: Find nearest clay tile via `gameMap.FindNearestClay(pos)`. If none found, return nil (triggers abandonment).
- Step 4: If character is at the clay tile position, return ActionDig intent with Target=pos, Dest=pos (ready to dig).
- Step 5: Otherwise, calculate next step via `NextStepBFS` toward the clay tile. Return ActionDig intent with Target=nextStep, Dest=clayTilePos.
- Set `CurrentActivity` to "Digging clay" (at tile) or "Moving to dig clay" (walking).
- **Anti-pattern:** This is NOT a self-managing action. No procurement phases, no RunVesselProcurement. Simple walk-then-act with ordered action one-tick-idle pattern.

**Wire into order system (system/order_execution.go):**
- `findOrderIntent` switch: add `case "dig": return findDigIntent(char, pos, items, order, log, gameMap)`
- `isMultiStepOrderComplete`: add `case "dig":` — count clay items in inventory, return true when count >= 2 (config.InventorySize or hardcoded 2 for now)
- `IsOrderFeasible`: add `case "dig": return gameMap.HasClay(), false` (add `HasClay() bool` to Map — returns `len(m.clay) > 0`)
- `DropCompletedBundle` call site in `selectOrderActivity`: after the existing `DropCompletedBundle` call, add `DropCompletedDigItems(char, order, gameMap, log)` call. Or restructure: add a general `DropCompletedOrderItems` that dispatches to bundle or dig drop logic. Recommend: keep `DropCompletedBundle` as-is, add `DropCompletedDigItems` as a new function, call both from the same site (they're no-ops for non-matching activity types).

**DropCompletedDigItems (system/order_execution.go):**
- New function: `DropCompletedDigItems(char, order, gameMap, log)`
- Early return if `order.ActivityID != "dig"`
- Collect all inventory items matching `order.TargetType` ("clay"), then drop each via `DropItem`
- Must collect items first, then drop (modifying inventory during iteration is unsafe)

**Handler — applyDig (ui/apply_actions.go):**
- Add `case entity.ActionDig: m.applyDig(char, delta)` to `applyIntent` dispatch table
- Walk-then-act pattern (follows `applyExtract`):
  - Walking phase: if `cpos != char.Intent.Dest`, call `moveWithCollision(char, cpos, delta)` and return
  - Working phase: accumulate `ActionProgress += delta`. When `>= config.ActionDurationShort`:
    - Reset ActionProgress
    - Create `NewClay(cpos.X, cpos.Y)` and add to inventory via `char.AddToInventory(clay)`
    - Log: "Dug clay"
    - Clear intent (`char.Intent = nil`) — ordered action pattern, one-tick-idle for re-evaluation
  - Guard: if `!char.HasInventorySpace()`, clear intent and return (safety net — findDigIntent should prevent this, but defend against edge cases)

**`continueIntent` (system/intent.go):**
- No changes needed. ActionDig has no TargetItem, no TargetWaterPos, no TargetFeature. Ordered actions re-evaluate via `CalculateIntent` each tick — `continueIntent` is not invoked for DrivingStat=="" intents without a TargetItem (line 55 of intent.go).

**Order UI (ui/update.go, ui/view.go):**
- "dig" activity appears in `getOrderableActivities()` as a top-level entry (no category). Since AvailabilityKnowHow, it only appears when at least one character knows "dig."
- Step 1 target selection: add `selectedActivity.ID == "dig"` branch. For now, create order immediately with `targetType: "clay"` — no sub-menu needed since there's only one dig target. When future dig targets appear, this becomes a target list (like gather).
  - Alternative: show a single-entry target list with just "Clay". This is more consistent with gather/extract flow and future-proofs for multiple dig targets. Recommend this approach — minor UI effort, better extensibility.
  - Implement `GetDiggableTypes(gameMap) []string` — returns `["clay"]` when `gameMap.HasClay()`. Called in step 1 to populate the target list.
- Step 1 navigation (up/down keys): add `"dig"` to the maxIndex calculation for step 1.
- Order `DisplayName()`: verify "Dig Clay" renders correctly — "dig" activity + "clay" targetType.

**Simulation handler (simulation/simulation.go):**
- Add `case entity.ActionDig:` to the simulation intent handler. Minimal: if at Dest, create clay item in inventory and clear intent. If not at Dest, move toward it.

**Architecture pattern:** **Adding New Actions — Ordered Action** checklist (architecture.md). All items covered:
1. Action constant — `ActionDig` in character.go
2. Activity entry — "dig" in ActivityRegistry with IntentOrderable, AvailabilityKnowHow
3. findDigIntent — in order_execution.go
4. Wire into findOrderIntent, isMultiStepOrderComplete, IsOrderFeasible
5. Handler — applyDig in apply_actions.go + dispatch table
6. continueIntent — no early-return block needed (no TargetItem, single-phase)
7. simulation.go — basic handler added
8. No vessel procurement (N/A)
9. No water fill (N/A)

**Values:** **Follow the Existing Shape** — ordered action pattern, item-based discovery, procurement drop pattern. **Consider Extensibility** — "dig" verb extends to future terrain extraction. **Isomorphism** — dig is a distinct physical action from gather.

**Reqs:** Construction-Reqs lines 49-53: "loose in hand pick up 1 clay from clay tile, clay tile is infinite source like water source, orderable." Changed from "available by default" to AvailabilityKnowHow per discussion (DD-17). Changed from "Gather > Clay" to "Dig Clay" per discussion (DD-15). Completion rule: both slots full → drop both (DD-16, DD-18).

**Behavioral details (architecture.md checklist):**
- **Targeting:** same-tile (character walks ON the clay tile — clay is passable)
- **Duration:** `ActionDurationShort`
- **Completion criteria:** both inventory slots contain clay items
- **Feasibility criteria:** clay tiles exist on the map (`HasClay()`)
- **Variety lock:** none — clay is uniform, no variety to lock to

### [TEST]

Build and run the game. Verify:
- A character looks at loose clay and discovers "dig" know-how (check knowledge panel)
- "Dig Clay" appears in the orders panel after discovery
- Creating a Dig Clay order: character drops any carried items, walks to clay tile, digs
- Character digs a second clay, then drops both on the ground
- Order shows as complete
- Multiple Dig Clay orders work in sequence
- Character pauses dig order to eat/drink if needs become pressing, then resumes

Use `/test-world` with a character near clay tiles and pre-discovered dig know-how for faster testing. Also test without pre-discovery to verify the discovery trigger works.

### [DOCS]
### [RETRO]
