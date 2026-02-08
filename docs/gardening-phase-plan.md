# Gardening Phase Plan

**Requirements source:** [docs/Gardening-Reqs.txt](Gardening-Reqs.txt)
**Vision context:** Precedes Construction, which reuses area selection UI and component procurement patterns established here. May reference ahead when making design decisions to set a good foundation.

---

## Decisions Log

| Decision              | Choice                                 | Rationale                                                                                                        |
| --------------------- | -------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Water tile model      | Map terrain, not features              | Springs and ponds should use the same system. Water is terrain, not a placed object. O(1) lookups, scales well. Leaf piles will eventually leave the feature system too (→ construction/beds), so features shrink over time. |
| Component procurement | Generic `EnsureHasRecipeInputs` helper | Reqs explicitly call for reuse by every future recipe. Model from existing `EnsureHasVesselFor`.                 |
| Unfulfillable orders  | Address in Part I alongside Till Soil  | Gardening orders are first to commonly become unfulfillable (no hoe, no seeds). Solve before it becomes painful. |
| Kind field on Items   | `Kind` subtype field on Item + Preference | Enables hierarchical preferences: "likes hoes" vs "likes shell hoes" vs "likes silver shell hoes". Recipe output sets ItemType (broad category) and Kind (recipe-specific subtype). Natural items leave Kind empty. |
| Hoe symbol            | `L`                                    | Simple, readable, distinct from other symbols.                                                                   |
| Action duration tiers | Short (0.83) / Medium (4.0) / Long (10.0) | Formalized for Craft Hoe (Long) and Till Soil (Medium). Code notes potential Extra Short / Extra Long tiers.    |
| Craft auto-drop       | Crafted items always drop to ground    | Available to any character. Consistent with existing vessel behavior.                                            |
| Recipe-level naming   | ItemType "hoe", Kind "shell hoe"       | User orders "Craft Hoe", character produces "shell hoe" based on recipe. Future recipes: "metal hoe", "wooden hoe". Description builds naturally: "silver shell hoe". |

---

## Part I: Gardening Preparation

### Slice 1: Ponds, Sticks, Shells, and Nuts

**Reqs covered:**

- New Feature: Ponds (Gardening-Reqs lines 12-16)
- Shells (Gardening-Reqs lines 18-20)
- Sticks (Gardening-Reqs lines 22-23)
- Edible Nuts (from randomideas.md — opportunistic, same "drop from canopy" pattern as sticks)

#### Design: Water as Map Terrain (unifying springs + ponds)

Springs and ponds are both water/drink sources and use the same system. Water is modeled as map terrain, not features. (See Decisions Log.)

- Add `water map[Position]WaterType` to `Map` struct — O(1) lookups everywhere
- `WaterType` distinguishes spring vs pond for rendering: springs render as `☉`, ponds as `≈`, both in `waterStyle`
- Springs become 1-tile water entries. Ponds become multi-tile water clusters. Same system.
- `FindNearestDrinkSource` → `FindNearestWater`: iterates water positions, same cardinal adjacency logic
- `IsBlocked(pos)` adds water check. New `IsWater(pos)` method on Map.
- `FeatureSpring` removed from feature system. `FeatureLeafPile` stays for now.
- Save migration: on load, detect old `FeatureSpring` entries, convert to water tiles. New saves use `WaterTiles` field.

#### Design: Ponds (Gardening-Reqs lines 12-16)

> _"1-5 ponds... contiguous blob shaped clusters of water tiles in size from 4-16 tiles... not passable... can be drank from, when standing in an adjacent tile"_

- 1-5 ponds of 4-16 contiguous water tiles, blob-shaped, impassable
- Symbol: `≈` (waves) rendered in `waterStyle` (blue)
- Drinking from pond: characters drink from cardinally adjacent tiles (same pattern as springs)
- Generation order: ponds generate **before** items and other features. `findEmptySpot` updated to respect water tiles so nothing spawns on them.
- Connectivity check: after generating all ponds, verify walkable map is still fully connected (BFS). Regenerate if partitioned.

#### Design: Sticks (Gardening-Reqs lines 22-23)

> _"A stick will occasionally spawn on the map, as though being 'dropped' from the canopy above. they can fall onto any unoccupied tile."_

- Symbol: `/` rendered in `woodStyle` (brown). Single variety (no color/pattern/texture variation). Future: different stick types per tree species.
- Spawning: random empty tiles ("dropped from canopy"). Initial handful at world start + periodic respawn every ~5 world days via non-plant spawning mechanism.
- Not edible. Forageable (pickup) and preference-formable.
- No `PlantProperties` — does not reproduce via plant system.

#### Design: Nuts (from randomideas.md)

- Symbol: `o` rendered in `woodStyle` (brown). Single variety for now. Future: color varieties per tree species.
- Edible (not poisonous, not healing — single variety), forageable, preference-formable. NOT plantable yet (deferred to Tree requirements).
- Spawning: identical to sticks — random empty tiles, initial spawn at world start, periodic respawn every ~5 world days.
- No `PlantProperties` — does not reproduce via plant system.

#### Design: Shells (Gardening-Reqs lines 18-20)

> _"a shell item will occasionally spawn adjacent to a water tile '<'... shells can have colors of: white, pale pink, tan, pale yellow, silver, gray, lavender"_

- Symbol: `<` per reqs. Colors: white, pale pink, tan, pale yellow, silver, gray, lavender (new Color constants needed for pale pink, pale yellow, silver, gray, lavender).
- Spawning: adjacent to pond tiles only (not springs). Initial handful at world start + periodic respawn every ~5 world days.
- Shell varieties: color-only (no pattern/texture). Preference-formable. Add to `GetItemTypeConfigs()` for variety generation; spawned by ground spawning system, not `SpawnItems()`.

#### Design: Non-Plant Spawning Mechanism

Sticks, nuts, and shells all use a new periodic spawning system separate from plant reproduction. No parent item needed — if count is below target, spawn on appropriate tile (random empty for sticks/nuts, pond-adjacent for shells). Rate: ~5 world days between spawn opportunities.

- New function `UpdateGroundSpawning(gameMap, delta, spawnTimer)` in `internal/system/ground_spawning.go`
- Timer-based: fires every `NonPlantSpawnInterval` (~600 game seconds = ~5 world days), with variance
- Count-based: counts current sticks/nuts/shells on map, spawns deficit up to target
- Shell spawning finds pond-adjacent empty tiles; skips if no ponds exist
- Timer state tracked in Model (`groundSpawnTimer float64`), included in save/load
- Called from `updateGame()` alongside `UpdateSpawnTimers`/`UpdateDeathTimers`

#### Design: Foraging Update for Nuts

Nuts have `Plant == nil` (not a growing plant) but are edible and should be foraged by characters seeking food. Change the foraging predicate in `system/foraging.go` from:

```
if item.Plant == nil || !item.Plant.IsGrowing { continue }
```

to:

```
if item.Plant != nil && !item.Plant.IsGrowing { continue }
```

Items with `Plant == nil` (nuts on ground) become forageable if `IsEdible()`. Items with `Plant.IsGrowing == false` (dropped plant items) remain excluded.

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → add minimal stubs to compile → verify red → implement → verify green → human testing checkpoint.

##### Step 1: Water Terrain System + Spring Migration ✅

**Tests** (in `internal/game/map_test.go`):
- `IsWater()` returns true for water tiles, false for empty tiles
- `WaterAt()` returns correct `WaterType` (spring vs pond)
- `IsBlocked()` returns true for water tiles
- `IsEmpty()` returns false for water tiles
- `FindNearestWater()` finds nearest water with available cardinal-adjacent tile
- `FindNearestWater()` skips water with all adjacent tiles blocked
- `FindNearestWater()` allows requester's adjacent position
- `FindNearestWater()` returns false when no water exists
- `WaterPositions()` returns all water positions
- `findEmptySpot()` doesn't place on water tiles

**Types & constants** needed: `WaterType` (spring/pond) in `map.go`, `CharWater = '≈'` in `config.go`

**Implementation** (`internal/game/map.go`, `internal/game/world.go`, `internal/entity/feature.go`):
- Add `water map[types.Position]WaterType` to Map struct
- New methods: `AddWater(pos, waterType)`, `IsWater(pos)`, `WaterAt(pos)`, `WaterPositions()`
- Update `IsBlocked(pos)` and `IsEmpty(pos)` to check water
- New `FindNearestWater()` replacing `FindNearestDrinkSource` (iterates water positions with cardinal adjacency)
- Remove `FeatureSpring` from `SpawnFeatures` — springs placed via `m.AddWater(pos, WaterSpring)`
- Update `findEmptySpot()` to check `!m.IsWater(pos)` and `m.FeatureAt(pos) == nil`

**Styles** (`internal/ui/styles.go`, `internal/ui/view.go`): water rendering in `renderCell()` before feature check. Springs render as `☉`, ponds as `≈`, both in `waterStyle`.

**Serialization** (pulled forward from Step 6 — needed for saves to work):
- Add `WaterTiles []WaterTileSave` to `SaveState` (position + water type)
- `ToSaveState`: serialize water positions
- `FromSaveState`: restore water tiles from save
- Migration: detect old `FeatureSpring` entries, convert to water tiles instead of features

**Legacy cleanup** (dead code from spring→water migration):
- Remove `FindNearestDrinkSource`, `DrinkSourceAt` from map.go (replaced by `FindNearestWater`)
- Remove `IsDrinkSource` rendering branch in view.go (water terrain renders first now)
- Remove `FindNearestDrinkSource` tests and `newTestSpring` helper from map_test.go
- Update `serialize_test.go` to test water tile round-trip instead of spring feature

**[TEST] Checkpoint — Water terrain:**
- `go test ./...` passes
- Build and run: springs still appear as `☉` and function identically to before (characters drink from adjacent tiles)
- Save, load — springs persist correctly. Load an old save — springs migrate from features to water tiles.
- Verify no regression in spring behavior before adding ponds

**Reqs reconciliation:** Gardening-Reqs line 16 _"can be drank from, when standing in an adjacent tile"_ — same drinking mechanic as existing springs, now unified in water terrain system.

##### Step 2: Pond Generation ✅

**Tests** (in `internal/game/world_test.go`):
- `SpawnPonds()` produces water tile count within valid range [4, 80]
- No undersized pond blobs: flood-fill all pond tiles into connected components, verify each component has >= `PondMinSize` tiles
- `isMapConnected()` returns true for empty/open maps
- `isMapConnected()` returns true when water doesn't partition map
- `isMapConnected()` returns false when water wall partitions map

**Config** needed: `PondMinCount = 1`, `PondMaxCount = 5`, `PondMinSize = 4`, `PondMaxSize = 16`

**Implementation** (`internal/game/world.go`, `internal/game/map.go`):
- New `SpawnPonds(m *Map)`: pick random count in [min, max], for each pond pick random size in [min, max], generate blob by random cardinal growth from a center tile, place water tiles as `WaterPond`
- Blob growth: pick random tile already in blob, add a random cardinal neighbor that's in-bounds and not already water. At pond generation time, map is clean (no items/features yet), so only need to check water and bounds.
- New `isMapConnected(m *Map) bool`: BFS from first non-blocked tile, verify all non-blocked tiles reached
- Retry strategy: if `isMapConnected` fails after placing all ponds, clear all pond water tiles and regenerate all ponds (not just the offending one). Max retry count to avoid infinite loops — partitioning should be extremely rare on 60x60 map with max 80 water tiles.
- Generation order: `SpawnPonds` (retry if not connected) → `SpawnFeatures` (leaf piles only now) → `SpawnItems` → spawn initial sticks/nuts/shells

**[TEST] Checkpoint — Ponds:**
- `go test ./...` passes
- Build and run: blue `≈` pond blobs visible on map, 1-5 ponds of varying sizes
- Walk characters toward ponds — they cannot enter water tiles
- Characters drink from ponds when thirsty (same as springs)
- No items or features spawned on water tiles
- Restart several times to confirm map is always fully connected

**Reqs reconciliation:** Gardening-Reqs lines 12-16 fully covered. _"random number of ponds (1-5)"_ ✓, _"contiguous blob shaped clusters... 4-16 tiles"_ ✓, _"not passable"_ ✓, _"drank from"_ ✓.

##### Step 3: New Item Types + Initial Spawn ✅

**Tests** (in `internal/entity/item_test.go`):
- `NewStick` returns correct properties (symbol `/`, `ItemType "stick"`, not edible, `Plant == nil`)
- `NewNut` returns correct properties (symbol `o`, `ItemType "nut"`, edible, not poisonous, not healing, `Plant == nil`)
- `NewShell` returns correct properties (symbol `<`, `ItemType "shell"`, not edible, `Plant == nil`, correct color)

**Types & config** needed:
- New Color constants (for shells): `ColorPalePink`, `ColorPaleYellow`, `ColorSilver`, `ColorGray`, `ColorLavender`. Add to `AllColors`. New `ShellColors` list.
- New character constants: `CharStick = '/'`, `CharNut = 'o'`, `CharShell = '<'`
- New config: `StickSpawnCount = 6`, `NutSpawnCount = 6`, `ShellSpawnCount = 6`
- Add `"nut"` to `StackSize` (10) and `ItemLifecycle` (immortal)

**Implementation** (`internal/entity/item.go`, `internal/game/world.go`, `internal/game/variety_generation.go`):
- New constructors: `NewStick(x, y)`, `NewNut(x, y)`, `NewShell(x, y, color)`
- Add shells to `GetItemTypeConfigs()` with `ShellColors`, non-edible, `SpawnCount: ShellSpawnCount`. Add `NonPlantSpawned bool` to skip in `SpawnItems()`
- Spawn initial sticks, nuts, and shells during world generation (after ponds and items). Shells adjacent to pond tiles; sticks/nuts on random empty tiles.

**Styles** (`internal/ui/styles.go`, `internal/ui/view.go`):
- New styles: `woodStyle` (brown "136"), `palePinkStyle` ("218"), `paleYellowStyle` ("229"), `silverStyle` ("188"), `grayStyle` ("250"), `lavenderStyle` ("183")
- In `styledSymbol()`: add cases for new colors. Sticks/nuts use `woodStyle` regardless of color field.

**[TEST] Checkpoint — New items visible:**
- `go test ./...` passes
- Build and run: brown `/` sticks and `o` nuts scattered across map
- Colored `<` shells appear near pond edges in various colors (white, pale pink, tan, etc.)
- Cursor over each item shows correct description in details panel
- No sticks/nuts/shells on water tiles

**Reqs reconciliation:** Gardening-Reqs lines 18-20 _"shell item... spawn adjacent to a water tile '<'"_ ✓, _"colors of: white, pale pink, tan, pale yellow, silver, gray, lavender"_ ✓. Lines 22-23 _"stick... spawn on the map... any unoccupied tile"_ ✓.

**Bug fix during testing:** Characters thrashed at pond edges when greedy pathfinding couldn't route around multi-tile obstacles. Added `NextStepBFS` in `movement.go` — BFS pathfinding that routes around permanent obstacles (water, impassable features) while ignoring characters (temporary). Falls back to greedy `NextStep` if no path exists. Updated all callers with gameMap access: `continueIntent`, `findDrinkIntent`, `findSleepIntent`, `findLookIntent`, `findTalkIntent`, `EnsureHasVesselFor`, `findHarvestIntent`, and pickup movement in `update.go`.

##### Step 4: Foraging Update for Nuts ✅

**Tests** (in `internal/system/foraging_test.go`):
- `scoreForageItems()` includes edible items with `Plant == nil` (nuts on ground)
- `scoreForageItems()` excludes non-edible items with `Plant == nil` (sticks)

**Implementation** (`internal/system/foraging.go`):
- Change predicate so items with `Plant == nil` and `IsEdible()` (nuts) are forageable. See design section above.

**[TEST] Checkpoint — Nuts are foraged:**
- `go test ./...` passes
- Build and run: observe characters seeking and eating nuts when hungry
- Nuts reduce hunger like other food
- Characters do NOT try to eat sticks or shells
- Existing foraging of berries/mushrooms/gourds unaffected

**[DOCS]** Update README (Latest Updates), CLAUDE.md, game-mechanics, architecture as needed for new item types (sticks, nuts, shells).

**[RETRO]** Run /retro.

##### Step 5: Non-Plant Spawning ✅

#### Design: Periodic Ground Spawning (no count cap)

Sticks and nuts fall from the canopy periodically; shells wash up near ponds. Each item type has its own independent timer on a random ~5 world day cycle. Items accumulate naturally — there is no target count or cap. This matches the simulation fiction: sticks fall from trees whether or not there are already sticks on the ground.

- Three independent timers (`GroundSpawnTimers` struct): `Stick`, `Nut`, `Shell`
- Each timer uses `GroundSpawnInterval` (600s / ~5 world days) ± `LifecycleIntervalVariance` (±20%), giving a 4-6 world day range per spawn
- When a timer fires: spawn one item, reset timer to new random interval
- Sticks/nuts: random empty tile (try up to 10 times, skip if no spot found)
- Shells: random pond-adjacent empty tile (skip if no ponds or no valid spots)
- Initial timer values randomized at world gen so all three don't fire simultaneously
- Export `FindPondAdjacentEmptyTiles` from world.go for reuse by ground spawning

**Tests** (in `internal/system/ground_spawning_test.go`):
- Timer does not fire before interval elapses
- Timer fires after interval elapses, spawns item, resets timer
- Spawns stick on empty tile (not on water)
- Spawns nut on empty tile (not on water)
- Spawns shell adjacent to pond tile only
- No shells spawn when no ponds exist
- Each type spawns independently (one type firing doesn't affect others)

**Config** needed: `GroundSpawnInterval = 600.0` (reuses existing `LifecycleIntervalVariance`)

**Implementation**:
- New file `internal/system/ground_spawning.go`: `GroundSpawnTimers` struct, `UpdateGroundSpawning()`, `RandomGroundSpawnInterval()` helper
- `internal/game/world.go`: Export `FindPondAdjacentEmptyTiles` (rename from `findPondAdjacentEmptyTiles`)
- `internal/ui/model.go`: Add `groundSpawnTimers GroundSpawnTimers` to Model, initialize with random values at world gen
- `internal/ui/update.go`: Call `UpdateGroundSpawning()` from `updateGame()` and `stepForward()` after `UpdateDeathTimers()`
- `internal/save/state.go`: Add `GroundSpawnTimers` fields to `SaveState`
- `internal/ui/serialize.go`: Save/load ground spawn timers

**[TEST] Checkpoint — Respawning:**
- `go test ./...` passes
- Build and run, let simulation run for extended time
- New sticks, nuts, and shells appear periodically
- Shells only appear near ponds, sticks/nuts anywhere
- Items accumulate over time (no cap)

**Reqs reconciliation:** Gardening-Reqs line 19 _"occasionally spawn"_ ✓, line 23 _"occasionally spawn"_ ✓.

---

**[TEST] Final Checkpoint — Full Slice 1:**

Start a new world. Verify ponds appear as blue `≈` blobs, are impassable, characters drink from them. Springs still appear as `☉` and function identically. Shells appear near pond edges in various colors. Sticks and nuts appear scattered in brown. All three respawn over time. No items spawn on water tiles. Map is fully connected. Save and load preserves everything.

**[DOCS]** Final doc pass for Slice 1 if anything was missed.

**[RETRO]** Run /retro.

---

### Slice 2: Craft Hoe

**Reqs covered:**

- New Activity: Craft Hoe (Gardening-Reqs lines 25-34)

#### Design: Kind Field for Hierarchical Item Identity

Crafted items have two levels of type identity:

- **ItemType**: broad product category ("hoe", "vessel") — what the user orders
- **Kind**: recipe-specific subtype ("shell hoe", "hollow gourd") — what the character produces

`Description()` uses Kind when present, falls back to ItemType. A shell hoe made with a silver shell displays as `"silver shell hoe"`. Preferences can form at any level: "likes hoes", "likes shell hoes", "likes silver shell hoes", "likes silver".

For natural items (berries, mushrooms, etc.), Kind is empty — no change to existing behavior.

This also updates the existing vessel: drop the hardcoded `Name: "Hollow Gourd"`, set `Kind: "hollow gourd"` from recipe output. Vessels now display their full variety description (e.g., "warty spotted green hollow gourd").

**Vision for recipe selection:** User orders a product ("Craft Hoe"). Character picks which recipe to use based on knowledge, resource availability, and preference. Currently one recipe per activity (shell-hoe, hollow-gourd), but the pattern supports future additions (metal-hoe, clay-pot, wooden-bucket). Character learns the activity + first recipe together; additional recipes come from talking to other characters.

#### Design: Action Duration Tiers

| Tier | Constant | Game seconds | World time | Used for |
|------|----------|-------------|------------|----------|
| Short | `ActionDurationShort` | 0.83 | ~10 min | Eat, drink, pickup |
| Medium | `ActionDurationMedium` | 4.0 | ~48 min | Till Soil, Look |
| Long | `ActionDurationLong` | 10.0 | ~2 hours | Craft Hoe, Craft Vessel |

Code comment noting potential Extra Short and Extra Long tiers as more actions are added.

#### Design: Hoe Item

- Symbol: `L` rendered in `woodStyle` (brown)
- ItemType: `"hoe"`, Kind: `"shell hoe"`
- Color inherited from shell input (enables preference variety: "silver shell hoe")
- Not edible, no Plant properties, no Container. Cannot go in vessel.

#### Design: Craft Hoe Recipe & Activity

Recipe `"shell-hoe"`:

- Inputs: 1 stick + 1 shell
- Output: ItemType `"hoe"`, Kind `"shell hoe"`
- Duration: `ActionDurationLong` (10.0)
- DiscoveryTriggers: look at or pick up stick or shell

Activity `"craftHoe"`:

- Name: `"Hoe"`
- IntentFormation: orderable, Availability: knowhow
- Appears under Craft category in order UI (auto-grouped by `"craft"` prefix on activity ID)

#### Design: EnsureHasRecipeInputs (Generic Multi-Component Procurement)

New helper in picking.go. Same return pattern as `EnsureHasVesselFor`: returns intent to go get something, or nil when ready (or impossible).

1. Check which recipe inputs are accessible (inventory items + container contents via `HasAccessibleItem`)
2. All present → return nil (ready to craft)
3. Need inventory space? Drop non-recipe loose items synchronously. Skip containers holding recipe inputs.
4. Seek nearest missing input on map → return pickup intent
5. Nothing available → return nil (triggers abandonment by caller)

New helper `findNearestGroundItemByType`: like `findNearestItemByType` but without the `Plant.IsGrowing` filter, for finding sticks/shells/non-growing items on the ground.

Component seeking uses nearest-distance. Preference-weighted seeking deferred (see triggered-enhancements.md).

#### Design: Generalized Craft Execution

**Intent**: Add `RecipeID string` to Intent struct. Craft intents carry the recipe ID so the ActionCraft handler knows which recipe to execute.

**findCraftIntent**: Replaces per-activity functions (`findCraftVesselIntent`). Works for any craft activity:

1. Get recipes for the order's activity
2. Filter to recipes whose inputs exist in the world
3. Pick first feasible recipe (future: preference scoring for recipe selection)
4. Use `EnsureHasRecipeInputs` to gather components
5. When ready → return ActionCraft intent with RecipeID

**ActionCraft handler**: Dispatch by `intent.RecipeID`. Per-recipe creation functions:

- `CreateVessel` (existing, updated): vessel from gourd, sets Kind from recipe
- `CreateHoe` (new): hoe from stick + shell, inherits shell color

**Auto-drop**: All crafted items placed on the ground at character's position. Available to any character who needs it.

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → add minimal stubs to compile → verify red → implement → verify green → human testing checkpoint.

##### Step 1: Kind Field on Item + Preference ✅

**Tests** (in `internal/entity/item_test.go`, `internal/entity/preference_test.go`):
- ✅ `Item.Description()` uses Kind when present, falls back to ItemType
- ✅ `Preference` with Kind matches items with matching Kind
- ✅ `Preference` with ItemType matches items regardless of Kind value
- ✅ `Preference` with Kind does NOT match items with different Kind
- ✅ `Preference.MatchScore` counts Kind as an attribute

**Implementation** ✅:
- Kind field added to Item struct + Description() updated
- Kind field added to Preference struct + Matches/MatchesVariety/AttributeCount/Description/ExactMatch updated
- Kind field added to RecipeOutput, hollow-gourd recipe updated with Kind "hollow gourd"
- Recipe duration updated to use ActionDurationLong (config constant added in Step 2)

##### Step 2: Duration Tiers + ActionDuration Rename ✅

##### Step 3: RecipeID on Intent + Preference Formation for Kind ✅

**Tests** (in `internal/system/preference_test.go`):
- `collectItemAttributes` includes "kind" for items with Kind set
- `collectItemAttributes` excludes "kind" for items without Kind
- `buildPreference` with "kind" attr sets Kind from item
- Combo preference with Kind uses Kind instead of ItemType

**Implementation:**
- Add `RecipeID string` to Intent struct in character.go
- Update `collectItemAttributes` / `collectExtraAttributes` / `buildPreference` in preference.go to handle Kind

##### Step 4: CreateVessel Update + Save/Load Kind ✅

**[TEST] Checkpoint — Kind field + durations:** ✅
- `go test ./...` passes
- Build and run: vessels now display full variety name (e.g., "warty spotted green hollow gourd") instead of "Hollow Gourd"
- Craft Vessel order still works end-to-end
- Save, load — no regression
- Preference formation works for Kind items (Bob formed "Likes orange hollow gourds")

##### Step 5: Hoe Item + Shell-Hoe Recipe ✅

**Tests** (in `internal/entity/item_test.go`):
- ✅ `NewHoe` returns correct properties (symbol `L`, ItemType `"hoe"`, Kind `"shell hoe"`, color, not edible)
- ✅ `NewHoe` color inherited from parameter (shell color)
- ✅ `NewHoe` description uses Kind ("silver shell hoe")

**Tests** (in `internal/entity/recipe_test.go`):
- ✅ shell-hoe recipe registered with correct inputs/output/duration/discovery triggers
- ✅ `GetRecipesForActivity("craftHoe")` returns shell-hoe
- ✅ craftHoe activity registered with correct properties

**Config:** `CharHoe = 'L'`

**Implementation:**
- ✅ `NewHoe(x, y int, color types.Color)` in item.go
- ✅ shell-hoe recipe in `RecipeRegistry`
- ✅ craftHoe activity in `ActivityRegistry`
- ✅ Hoe rendering uses shell's color via existing color switch (not hardcoded woodStyle)
- ✅ `itemFromSave` case for `"hoe"`
- ✅ Discovery: characters learn craftHoe + shell-hoe recipe from looking at/picking up sticks and shells

##### Step 6: EnsureHasRecipeInputs ✅

**Design decisions:**
- **Unified `findNearestItemByType`**: Instead of adding a separate `findNearestGroundItemByType`, add `growingOnly bool` parameter to the existing `findNearestItemByType` and move it from order_execution.go to picking.go. This keeps all item-seeking utilities in picking.go (alongside `FindAvailableVessel`) and maintains the "call downward" pattern where order_execution.go calls into picking.go for prerequisites and search.
- **Shell in vessel**: Shells have color varieties and can end up in vessels. `HasAccessibleItem("shell")` already checks vessel contents. Drop logic must not drop a container holding a recipe input.
- **Nearest-distance for now**: Component seeking uses `findNearestItemByType`. Preference-weighted seeking deferred (see triggered-enhancements.md — candidates to generalize from foraging.go's scoring).

**Tests** (in `internal/system/picking_test.go`):
- `findNearestItemByType` with `growingOnly=false` finds non-growing items (sticks, shells)
- `findNearestItemByType` with `growingOnly=true` still only finds growing items (regression)
- `findNearestItemByType` ignores wrong types
- `EnsureHasRecipeInputs` returns nil when all inputs accessible (inventory)
- `EnsureHasRecipeInputs` returns nil when input accessible in container
- `EnsureHasRecipeInputs` returns pickup intent for missing input
- `EnsureHasRecipeInputs` drops non-recipe loose items to make space
- `EnsureHasRecipeInputs` does NOT drop container holding recipe input
- `EnsureHasRecipeInputs` returns nil when inputs not available on map

**Implementation:**
- Move `findNearestItemByType` from order_execution.go to picking.go, add `growingOnly bool` param, update all callers
- `EnsureHasRecipeInputs` in picking.go

##### Step 7: Generalized Craft Execution ✅

**Design: `findCraftIntent` replaces `findCraftVesselIntent`**

Generic craft intent finder for any recipe-based activity:

1. `GetRecipesForActivity(order.ActivityID)` → get all recipes for this activity
2. Filter to feasible recipes (at least one of each input type exists in the world)
3. Pick first feasible recipe (future: score by preference for inputs — see triggered-enhancements.md)
4. `EnsureHasRecipeInputs(char, recipe, ...)` → gather components
5. When ready (all inputs accessible) → return `ActionCraft` intent with `RecipeID` set

**Design: `findOrderIntent` dispatch**

Instead of string-matching craft activity IDs, check whether the activity has recipes via `GetRecipesForActivity`. Any activity with registered recipes routes to `findCraftIntent`. Harvest and other non-recipe activities keep explicit cases.

**Design: Generalized `ActionCraft` handler**

Look up recipe from `intent.RecipeID`. Verify all inputs still accessible. On completion, dispatch to per-recipe creation function by recipe ID. `CreateVessel` (existing, updated to work with generalized path). `CreateHoe` (new): consumes stick + shell, creates hoe inheriting shell's color.

**Behavior changes from generalization (improvements):**
- Craft Vessel now uses `EnsureHasRecipeInputs` → picks up non-growing gourds too (dropped gourds now craftable, correct behavior)
- Craft Vessel now uses BFS pathfinding via `createItemPickupIntent` (no more thrashing around ponds)

**Tests** (in `internal/system/order_execution_test.go`):
- `findCraftIntent` returns ActionCraft with correct RecipeID when inputs gathered
- `findCraftIntent` returns pickup intent via EnsureHasRecipeInputs when inputs missing
- `findCraftIntent` returns nil when no recipe feasible (inputs don't exist in world)

**Tests** (in `internal/ui/update_test.go`):
- ActionCraft handler creates shell hoe from stick + shell (correct ItemType, Kind, color from shell)
- ActionCraft handler still creates vessel from gourd (regression test)
- Crafted items placed on ground at character position (not in inventory)

**Implementation:**
- `findCraftIntent` replaces `findCraftVesselIntent` in order_execution.go
- Update `findOrderIntent` dispatch: recipe-based activities → `findCraftIntent`
- `CreateHoe(stick, shell *entity.Item, recipe *entity.Recipe)` in crafting.go
- Generalize ActionCraft handler in update.go: look up recipe by `intent.RecipeID`, dispatch to per-recipe creation function
- Per-recipe creation: `CreateVessel` (existing), `CreateHoe` (new)
- Both auto-drop crafted item to ground
- Remove `findCraftVesselIntent` (dead code after generalization)

**[TEST] Checkpoint — Full Craft Hoe:**
- `go test ./...` passes
- Build and run: order Craft Hoe. Character drops non-components, seeks stick and shell, picks them up, crafts. Hoe appears on ground with shell's color in description (e.g., "silver shell hoe").
- Verify discovery: character looks at stick/shell, may learn Craft Hoe
- Verify Craft Vessel still works through generalized path
- Cursor over hoe shows correct description in details panel
- Save and load preserves hoe correctly

**Reqs reconciliation:** Gardening-Reqs lines 25-34. _"Components: 1 stick + 1 shell"_ ✓, _"Orderable"_ ✓, _"Discoverable"_ ✓, _"Duration: Medium"_ → Long per discussion ✓, _"creates 'Hoe' item"_ ✓, _"Cannot go in a vessel"_ ✓, _"goes into open inventory slot or is dropped on ground"_ → always drops per discussion ✓.

**[DOCS]** Update README (Latest Updates), CLAUDE.md, game-mechanics, architecture as needed.

**[RETRO]** Run /retro.

---

### Slice 3: Till Soil with Area Selection UI

**Reqs covered:**

- New Activity: Garden > Till Soil (lines 36-59)

**Outcomes:**

1. New order type: Till Soil, orderable, discoverable (from picking up or looking at hoe)
2. Area selection UI: new input mode for highlighting map tiles
   - Keyboard-driven area highlighting
   - Tiles with features (springs, leaf piles, ponds) excluded from selection
   - Already-tilled tiles not selectable
   - Pending-till tiles shown as pre-highlighted
   - Marked tiles visible in details screen (select view)
3. Tilling action: medium duration per tile, requires hoe in inventory
   - On completion: tile gets "Growing" style (olive background)
   - Growing items on tile destroyed; non-growing items displaced to adjacent tile
   - Character moves to next nearest marked tile automatically
4. Multi-assignment: multiple characters can work on the same set of marked tiles
5. Required item procurement: character must have hoe before starting (uses `EnsureHasItem` helper)

**[TEST] Checkpoint:** Character discovers Till Soil from hoe. Issue Till Soil order, select area on map. Character fetches hoe, tills tiles one by one (olive background appears). Non-growing items displaced. Test with 2 characters tilling same area. Verify feature tiles excluded from selection.

**Feature questions:**

- Area selection keyboard controls: arrow keys to move cursor, hold shift to extend? Or click-and-drag style with start/end corners?
  - Make a reccomendation
- Visual indicator for "marked for tilling" vs "tilled" on map?
  - Marked for tilling should only be visible during area selection UI
  - For 'Tilled', Show some options with different symbols and options
- Should the area selection UI pattern be documented for reuse by Construction's fence/hut placement?
  - Let's assume yes

---

### Slice 4: Unfulfillable Orders

**Reqs covered:**

- Orders Unfulfillable Logic (lines 137-148)

**Outcomes:**

1. Characters skip orders they can't complete (move to next open order)
2. Unfulfillable detection: no one has required know-how, OR required components don't exist in world
3. Unfulfillable status shown on orders screen
4. Characters don't assign themselves to unfulfillable orders (no assign/abandon loop)

**[TEST] Checkpoint:** Create a Till Soil order when no hoe exists and no stick/shell available. Verify it shows as unfulfillable. Characters skip it and take other orders. Create a Craft Hoe order; after hoe is crafted, Till Soil becomes fulfillable again.

**Feature questions:**

- Should unfulfillable be re-evaluated every tick, or only when world state changes (item added/removed)?
  - Make reccomendation
- UI treatment: grayed out? Different status text? Icon?
  - Show examples of different options

[Pause for evaluation before continuing for Part II, opportunity to pull in quick wins or opportunistic random items. Ensure user has finished Part II requirements and plan is updated accordingly.]

---

## Part II: Seeds, Planting, and Watering

### Slice 5: Seeds and Plantable Attribute

**Reqs covered:**

- New ItemType: Seed (lines 63-75)
- New Item Attribute: Plantable (lines 77-80)

**Outcomes:**

1. Seed item type: carries variety of source item, max stack size 20, same-variety stacking only
2. Gourd seeds: created when eating a gourd (1 seed per gourd consumed)
3. Flower seeds: created when foraging a flower (1 seed, flower not removed)
4. Plantable attribute on items: all seeds are plantable; picked berries and mushrooms become plantable
5. Standard pickup logic applies to seeds (matching container > empty container > inventory slot > drop)

**[TEST] Checkpoint:** Eat a gourd, seed appears. Forage a flower, seed appears and flower remains. Seeds stack in vessels by variety. Pick a growing berry - it becomes plantable. Verify seed descriptions carry parent variety info.

**Feature questions:**

- How do flower seeds display? "Blue flower seed"? Symbol?
  - likely a '.' in its color
- Does foraging flowers require any know-how, or is it default like berry foraging?
  - Flowers should become a targetable item in the existing foraging activity using the same preference targeting, except instead of being removed from map it yields a seed.
  - - Will need to figure out how often a seed can be collected from a flower, probably on a similar cadance as flower reproduction, maybe even using same config

---

### Slice 6: Plant Activity

**Reqs covered:**

- New Activity: Plant (lines 82-94)

**Outcomes:**

1. Activity: orderable, discoverable (trigger from looking at or picking up plantable items)
2. Order UI: Garden > Plant > Select plantable item type (gourd seeds, flower seeds, mushrooms, berries)
3. Character plants selected type on tilled soil tiles
   - Uses carried plantable item if available, otherwise seeks by preference/distance
   - Abandons if no plantable item available
   - Planted item becomes a growing "sprout" of that variety
4. Continuation: character keeps planting same specific variety until supply exhausted or no unplanted tilled tiles remain
5. Multi-assignment: multiple characters can plant simultaneously with different varieties

**[TEST] Checkpoint:** Till some soil, then issue Plant > gourd seeds. Character fetches gourd seeds, plants them on tilled tiles. Sprouts appear. Test planting berries directly (no seed needed). Test abandonment when no plantable items available.

---

### Slice 7: Fetch Water

**Reqs covered:**

- New Idle Activity: Fetch Water (lines 96-103)

**Outcomes:**

1. Idle activity: characters choose to fill empty vessel with water (same vessel logic as foraging)
2. Water-filled vessel: 4 drinks/units per vessel
3. Characters can drink from carried water vessel (potentially triggered earlier than walking to water source)
4. Dropped water vessels can be targeted for drinking by other characters
5. Water vessel display in UI

**[TEST] Checkpoint:** Character with empty vessel goes to pond/spring, fills it. Character drinks from carried vessel when thirsty instead of walking to water. Drop water vessel, another character drinks from it. Verify 4 drinks per vessel.

**Feature questions:**

- Should this be a variation of forage or its own activity? (Reqs ask this explicitly)
- Can drinking from carried vessel be triggered at an earlier thirst threshold than walking to a source?
- Preference formation for water vessels? (See triggered enhancement - likely defer)

---

### Slice 8: Garden Plant Growth

**Reqs covered:**

- Enhanced Logic: Garden Plant Growth and Reproduction (lines 108-113)
- Food Turning (lines 115-117)

**Outcomes:**

1. Sprout phase: planted items start as sprouts, grow to full plants (6 min duration)
2. Extend normal plant reproduction to include sprout phase with baseline duration
3. Tilled ground growth bonus: faster growth and reproduction
4. Watered growth bonus: faster growth and reproduction (requires Water Garden - partial reqs)
5. Food turning: different edible items have different satiation amounts
6. Different growing items have different lifecycle times

**[TEST] Checkpoint:** Plant seeds, watch sprouts grow into full plants. Verify tilled ground plants grow faster than wild ones. Compare lifecycle times across plant types. Verify different food items restore different hunger amounts.

**Feature questions:**

- Water Garden activity reqs say "REQS TO DO" - skip or define during implementation?
- Specific growth rate bonuses for tilled vs watered?
- Specific satiation values per food type?

---

## Triggered Enhancements to Monitor

These are from [docs/triggered-enhancements.md](triggered-enhancements.md). They may be triggered during Gardening but don't need to be planned upfront.

| Enhancement                            | Trigger During Gardening                                  | Action                                                                                         |
| -------------------------------------- | --------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| **Order completion criteria refactor** | Adding Till Soil, Plant, Water Garden (3 new order types) | Monitor if completion logic in update.go exceeds ~50 lines. Refactor to handler pattern if so. |
| **ItemType constants**                 | Adding stick, shell, hoe, seed (4 new types, total ~9)    | Evaluate after Part I whether string comparisons are error-prone.                              |
| **Category formalization**             | Hoe is first "tool" category                              | Note pattern but defer to Construction per triggered-enhancements.md.                          |
| **Preference formation for beverages** | Fetch Water introduces water vessels as drinkable         | Evaluate during Slice 7; likely defer until actual beverage variety exists.                    |
| **Action duration tiers**              | Craft Hoe and Till Soil both need "medium" duration       | Define Short/Medium/Long tiers if not already formalized by Slice 2.                           |
| **UI extensibility refactoring**       | Area selection UI is new pattern                          | Document approach for reuse by Construction.                                                   |

## Opportunistic Additions to Consider

From [docs/randomideas.md](randomideas.md):

| Idea                   | Opportunity                                                                     | Status                                                                           |
| ---------------------- | ------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| **Edible Nuts**        | Same "drop from canopy" pattern as sticks.                                      | **Included in Slice 1.** Edible, forageable, brown `o`. Plantable deferred to Tree reqs. |
| **Order Selection UX** | Gardening adds more order types, making scrolling more painful.                 | Consider after Part I when the pain is fresh. Not in scope for gardening itself.  |
