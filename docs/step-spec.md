# Step Spec: Step 8 — Construct Discovery + Hut Placement UI

Design doc: [construction-design.md](construction-design.md)

---

## Sub-step 8a: Construct Discovery + buildHut Registry

**Anchor story:** A character wanders past a stick fence and pauses to look at it. They form a preference, and also have a moment of insight — they realize they could build a stick hut! "Discovered how to build Hut!" and "Learned Stick Hut recipe!" appear in the action log. Another character who already knows how to build huts looks at a brick fence and discovers the brick hut recipe specifically — "Learned Brick Hut recipe!" — but doesn't re-discover the activity.

### Scope

**Discovery mechanism (entity/activity.go, system/discovery.go):**

- `ConstructKind string` field on `DiscoveryTrigger` — parallel to `ItemType`. When set, the trigger matches construct interactions rather than item interactions.
- `TryDiscoverFromConstruct(char, action, constructKind, log, chance)` in discovery.go — entry point for construct-based discovery. Mirrors `TryDiscoverKnowHow` structure: tries `tryDiscoverActivityFromConstruct` then `tryDiscoverRecipeFromConstruct`, returns true if something new was discovered.
- `tryDiscoverActivityFromConstruct` — iterates `GetDiscoverableActivities()`, checks triggers where `ConstructKind` matches. Same structure as `tryDiscoverActivity`.
- `tryDiscoverRecipeFromConstruct` — iterates `GetDiscoverableRecipes()`, checks triggers where `ConstructKind` matches. Same structure as `tryDiscoverRecipe` (grants activity via `LearnActivity`, grants recipe via `LearnRecipe`, logs discovery, processes `BundledActivities`).
- `constructTriggerMatches(trigger, action, constructKind)` — parallel to `triggerMatches`. Checks: `trigger.Action == action` AND `trigger.ConstructKind != ""` AND `trigger.ConstructKind == constructKind`. A trigger with empty `ConstructKind` never matches construct interactions (it's an item trigger).

**CompleteLookAtConstruct integration (system/looking.go):**

- Add call to `TryDiscoverFromConstruct(char, entity.ActionLook, construct.Kind, log, GetDiscoveryChance(char))` at the end of `CompleteLookAtConstruct`, after preference formation (replacing the DD-36 deferral comment).
- Follows the same position in `CompleteLook`: preference formation, then discovery.

**buildHut activity (entity/activity.go):**

- New entry in `ActivityRegistry`:
  - ID: `"buildHut"`
  - Name: `"Hut"`
  - Category: `"construction"`
  - IntentFormation: `IntentOrderable`
  - Availability: `AvailabilityKnowHow`
  - No `DiscoveryTriggers` — discovered via recipe triggers (same pattern as `buildFence`, DD-27)

**Hut recipes (entity/recipe.go):**

Three recipes following the fence recipe pattern. Per-tile inputs (DD-39). `ActivityID: "buildHut"`. Output is display-only (actual output is a construct, same as fence recipes).

- `"thatch-hut"`: Name "Thatch Hut", Inputs: `[{ItemType: "grass", Count: 12}]` (2 bundles of 6 per tile), Output: `{ItemType: "hut"}`, DiscoveryTriggers: `[{Action: ActionLook, ConstructKind: "fence"}]`
- `"stick-hut"`: Name "Stick Hut", Inputs: `[{ItemType: "stick", Count: 12}]`, Output: `{ItemType: "hut"}`, DiscoveryTriggers: `[{Action: ActionLook, ConstructKind: "fence"}]`
- `"brick-hut"`: Name "Brick Hut", Inputs: `[{ItemType: "brick", Count: 12}]`, Output: `{ItemType: "hut"}`, DiscoveryTriggers: `[{Action: ActionLook, ConstructKind: "fence"}]`

**Order pipeline guards (system/order_execution.go):**

Prevents `findOrderIntent`'s default branch from routing `buildHut` to `findCraftIntent` before execution exists.

- `case "buildHut": return nil` in `findOrderIntent` switch — replaced with `findBuildHutIntent` in Step 10.
- `case "buildHut": return false, false` in `IsOrderFeasible` — replaced with real feasibility check in sub-step 8b.

### Architecture patterns

- **Adding a New Activity checklist** (architecture.md): items 1-3 (registry entry, category, discovery triggers via recipes). Item 4 (findOrderIntent case) is a nil stub. Item 5 (new category) — not needed, reuses existing "construction" category.
- **Recipe discovery pattern** (DD-27): recipes carry discovery triggers, first recipe grants activity know-how via `LearnActivity(recipe.ActivityID)`. Same pattern as fence recipes.
- **Isomorphism**: `constructTriggerMatches` is parallel to `triggerMatches` — constructs aren't items, so they get their own trigger matching function rather than being shoehorned into item triggers.
- **Follow the Existing Shape**: `TryDiscoverFromConstruct` mirrors `TryDiscoverKnowHow` (activities then recipes). `activityCategoryVerb("construction")` already returns `"build"`.

### Anti-patterns to avoid

- Do NOT add `ConstructKind` checks to the existing `triggerMatches` function — it takes an `*Item`, not a construct. Keep item and construct trigger matching separate.
- Do NOT let `buildHut` fall through to `findCraftIntent` — hut recipes use a custom execution path (same as fence recipes, DD-27).

### Tests

- Unit: `constructTriggerMatches` — matches when Action and ConstructKind both match; rejects when Action mismatches; rejects when ConstructKind mismatches; rejects when trigger has empty ConstructKind (item-only trigger)
- Unit: `TryDiscoverFromConstruct` — discovers hut recipe when looking at fence (chance=1.0); grants buildHut activity on first recipe discovery; second recipe discovery doesn't re-grant activity; respects chance=0 (no discovery)
- Unit: `tryDiscoverActivityFromConstruct` — returns false when no activities have construct triggers (current state); would discover if a construct-triggered activity existed (future-proofing)
- Regression: existing item-based discovery still works (triggerMatches unchanged, TryDiscoverKnowHow unchanged)

### [TEST] Checkpoint

Create test world with: 2-3 characters (happy mood for discovery), several fences of different materials (stick, thatch, brick), some items on the ground.

Verify:
- Characters look at fences and discover hut recipes — "Discovered how to build Hut!" and "Learned [Material] Hut recipe!" appear in action log
- A character who already knows buildHut but looks at a different material fence discovers the new recipe without re-discovering the activity
- "Hut" appears in the Construction menu (under orders) after discovery
- Selecting "Hut" from the menu does nothing useful yet (order is unfulfillable) — this is expected

### [DOCS]
Run /update-docs after human testing confirms success.

### [RETRO]
Run /retro after docs are updated.

---

## Sub-step 8b: Hut Placement UI

**Anchor story:** The player selects Construction > Build Hut and enters placement mode. A 5×5 footprint preview follows the cursor, showing which tiles would become walls (perimeter) with the door marked on the south side. Invalid placements (overlapping water, constructs, or map edges) show in a different color. The player presses `p` to confirm — 16 perimeter tiles are marked for construction. The details panel shows "Marked for construction (Hut)" on marked tiles.

### Scope

**ConstructKind on ConstructionMark (game/map.go, save/state.go, ui/serialize.go):**

- Add `ConstructKind string` field to `ConstructionMark` struct (DD-41)
- Add `ConstructKind string` field to `ConstructionMarkSave` struct
- Update `MarkForConstruction(pos, lineID)` signature to `MarkForConstruction(pos, lineID, constructKind)`
- Update all existing `MarkForConstruction` callers (fence line-drawing) to pass `"fence"`
- Update `constructionMarksToSave` and `FromSaveState` to serialize/deserialize `ConstructKind`
- `HasUnbuiltConstructionPositions(constructKind string)` — add filter parameter. Update all callers (fence feasibility, fence completion) to pass `"fence"`.
- `isMultiStepOrderComplete` fence case: pass `"fence"` to `HasUnbuiltConstructionPositions`
- `IsOrderFeasible` fence case: pass `"fence"` to `HasUnbuiltConstructionPositions`

**IsOrderFeasible + findOrderIntent (system/order_execution.go):**

- Replace 8a's `IsOrderFeasible` stub with real check: `gameMap.HasUnbuiltConstructionPositions("hut") && hutMaterialExistsOnMap(items)` — follows fence feasibility pattern
- `hutMaterialExistsOnMap` checks the same material types as fence (grass, stick, brick) but with hut per-tile quantity threshold (12)
- 8a's `findOrderIntent` nil stub stays — replaced with `findBuildHutIntent` in Step 10

**Footprint placement UI (ui/update.go):**

- `case "buildHut"` branch in the category activity selection (alongside existing `tillSoil` and `buildFence` branches)
- 5×5 footprint placement mode (DD-38): cursor positions top-left corner of footprint
- Footprint preview rendering: show 5×5 area with perimeter highlighted, door position marked (south wall center)
- Validation: full 5×5 area must be clear of water, existing constructs, impassable features, and map edges
- Invalid footprints show in a distinct color (follow existing area selection error highlighting)
- `p` confirms: marks 16 perimeter tiles with `ConstructKind: "hut"`, shared LineID, material empty
- Door position: bottom edge center tile (x+2, y+4 relative to top-left)
- `Enter` creates order: `NewOrder(id, "buildHut", "")`
- `Esc` cancels placement mode

**Details panel (ui/view.go):**

- When cursor is on a hut construction mark: show "Marked for construction (Hut)" — follows fence mark display pattern, using `ConstructKind` to determine label
- When material is stamped: show "Marked for construction (Stick Hut)" etc.

### Architecture patterns

- **Marked-for-Construction Pool** (architecture.md): extends the existing pool with `ConstructKind` filter. Fence and hut marks share the same `markedForConstruction` map — kind is a filter, not a separate data structure (DD-41).
- **Area Selection UI Pattern** (architecture.md): reuses area selection infrastructure. Footprint mode is a third shape (after rectangle for tilling and line for fences). Fixed 5×5, no anchor/drag — cursor positions the footprint, `p` confirms.
- **Follow the Existing Shape**: fence mark flow is the template. `MarkForConstruction` gains a parameter; callers specify kind.
- **Same-step serialization**: `ConstructKind` is added to both runtime and save structs in this sub-step.

### Anti-patterns to avoid

- Do NOT create a separate `markedForHutConstruction` pool — use the shared pool with `ConstructKind` filter (DD-41).
- Do NOT change the fence placement UX — fence still uses line-drawing mode. Only hut uses footprint mode.

### Tests

- Unit: `MarkForConstruction` with `ConstructKind` — fence marks carry `"fence"`, hut marks carry `"hut"`
- Unit: `HasUnbuiltConstructionPositions("fence")` only counts fence marks; `HasUnbuiltConstructionPositions("hut")` only counts hut marks
- Unit: `IsOrderFeasible` for buildHut — feasible when hut marks + materials exist; unfeasible when no marks; unfeasible when no materials
- Unit: serialization round-trip — `ConstructKind` survives save/load for both fence and hut marks
- Regression: fence placement, feasibility, and completion still work after `ConstructKind` addition

### [TEST] Checkpoint

Create test world with: characters who know buildHut (via 8a discovery or pre-granted), materials on the ground (sticks, grass, bricks), open terrain for placement.

Verify:
- Construction > Hut enters footprint placement mode with 5×5 preview following cursor
- Invalid placements (over water, map edges, existing constructs) show distinct highlighting
- `p` confirms valid placement — 16 perimeter tiles marked
- Details panel on marked tile shows "Marked for construction (Hut)"
- Build Hut order can be created (Enter after marking)
- Order shows feasible when marks + materials exist, `[Unfulfillable]` when materials exhausted
- Fence placement still works correctly (regression)
- Save/load preserves both fence and hut marks with correct `ConstructKind`
- Characters assigned to Build Hut order don't do anything (no execution yet) — this is expected

### [DOCS]
Run /update-docs after human testing confirms success.

### [RETRO]
Run /retro after docs are updated.

---

## Implementation Sequencing Notes

### Why two sub-steps

8a is a cohesive unit: the discovery mechanism, activity, recipes, and pipeline guards all need each other for testable behavior (character discovers hut recipe by looking at fence). It can be tested independently — discovery works, menu shows the option, guards prevent misrouting.

8b is a separate concern: UI placement mode, ConstructKind filtering, and mark management. It depends on 8a (the activity must exist for the order UI), but is independently testable (marks appear on map, details panel shows them, feasibility works).

### Deferred from Step 8

- **Build Hut execution** (`findBuildHutIntent`, `applyBuildHut`): Step 10. The `findOrderIntent` nil stub from 8a stays until then.
- **Hut construct types** (wall/door entities, line-drawing symbols): Step 9. Tested with pre-placed constructs via test-world.
- **Preference-weighted recipe selection**: Step 12 (Phase Wrap-Up). Characters currently select nearest available material.
