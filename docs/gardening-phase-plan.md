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
| Plantable attribute   | Explicit `Plantable bool` on Item      | Clean, extensible. Seeds get it at creation; berries/mushrooms set it when picked. Derived approach (infer from Plant state) is implicit and harder to extend to future items. |
| Liquid storage        | Liquids as vessel Stacks               | Water/future liquids (mead, beer, wine) stored in ContainerData.Contents as Stacks with ItemType and Count (units). Vessel holds items OR liquid, never both. Reuses all existing vessel infrastructure. |
| Wet tile tracking     | On-the-fly + timer for manual watering | Water-adjacent tiles (8-dir) always wet — computed, no state. Character-watered tiles tracked with decay timer (3 world days). Scales if water features change. |
| Tilled soil visual    | Parallel line symbol + underline       | Tilled soil gets a ground symbol. Plants on tilled soil underlined. "Tilled" attribute visible in details panel when ground selected. Highlight only in selection view. |
| Sprout colors         | Olive (dry) / green (wet) / mushroom exception | Sprouts appear olive by default, green when on wet tile. Mushroom sprouts always show variety color. |
| Tilled soil model     | Map terrain state, not feature         | Tilled soil is a terrain modification. Items must exist on tilled tiles (plants grow there). Features block `IsEmpty()`. O(1) lookups via `tilled map[Position]bool`. Same pattern as water terrain. |
| Tilled soil visual    | Option F: `═══` full fill, `═●═` around entities | Box-drawing `═` in olive (142). Empty tilled tiles render `═══`. Entities on tilled soil get `═X═` fill padding. No background color. Tilled state shown in details panel. |
| Activity categories   | `Category` field on Activity struct    | Explicit grouping replaces prefix convention. Cleaner than string prefix hacking. Scales to future categories (Pigment, Construction). Existing craft activities migrated. |
| Area selection UI     | Rectangle anchor-cursor pattern        | Enter to anchor, move cursor for rectangle, Enter to confirm. Invalid tiles silently excluded. Reusable for Construction fence/hut placement. |
| Till order state      | Global marked-for-tilling pool on Map  | Marked tiles = user's tilling plan (persistent, independent of orders). Till orders = worker assignments. Cancel order = remove worker, not the plan. Unmark via separate UI action. Characters with any tillSoil order work the shared pool. |
| Lookable water terrain | Extend Look to water-adjacent positions | Characters contemplate water sources. Triggers know-how discovery (Water Garden) but not preference formation (water has no item attributes). |
| Food satiation tiers  | Feast (50) / Meal (25) / Snack (10)   | Per-item-type satiation. Gourd=Feast, Mushroom=Meal, Berry/Nut=Snack. Replaces flat FoodHungerReduction. |
| Growth speed tiers    | Fast / Medium / Slow                   | Berry/mushroom=fast, flower=medium (6-min sprout), gourd=slow. Affects sprouting duration and reproduction intervals. |
| Watered tile decay    | 3 world days (360 game seconds)        | Manual watering creates temporary wet status. Tiles adjacent to water are permanently wet (computed). |
| Water symbol          | `▓▓▓` dark shade block in waterStyle   | Textured block for ponds (suggests water movement). Springs keep `☉`. Aligns with terrain fill system (`═══` tilled, `▓▓▓` water). Structure walls TBD. |
| Sprout symbol         | `𖧧` (bold style)                       | Visually distinct from mature plants. Suggests early growth stage. |
| Sprout colors         | Olive (dry) / green (wet) / mushroom variety | Consistent with tilled soil olive. Green indicates water benefit. Mushrooms always variety-colored. |
| Tilled plant visual   | `═X═` fill, no underline               | Existing fill pattern communicates tilled soil. Underline has terminal compatibility concerns. |
| Plant discovery       | `RequiresPlantable` on DiscoveryTrigger | Matches existing `RequiresEdible` pattern. Extensible for future plantable types. |
| Variety locking       | `LockedVariety` on Order               | Per-order lock set on first plant. Character seeks only matching variety after lock. Completion when variety exhausted. |
| Sprout maturation     | In-place symbol change from ItemType   | `MatureSymbol()` helper derives mature symbol. No stored state needed. |
| Growth tuning         | Deferred to Slice 9                    | Flat 30s sprout for testing. Tier differentiation and bonus values in Slice 9 tuning. |
| Liquid type model     | ItemType `"liquid"`, Kind `"water"`    | Extensible for future beverages (mead, beer, wine) as Kind variants under same ItemType. Parallels Kind pattern from seeds/hoes. |
| Fetch water idle slot | Own 1/5 slot in idle roll              | Expand from 4 options to 5: look/talk/forage/fetch-water/idle. Equal probability. |
| Vessel filling action | New `ActionFillVessel`                 | Distinct from ActionPickup — creates liquid from terrain interaction, no source item. Vessel-seeking phase reuses foraging patterns. |
| Drink source search   | Unified distance-based (like hunger)   | Carried vessel (dist 0), ground vessel (dist to vessel), water terrain (dist to adjacent). Closest wins. Future: preference scoring for beverages. |
| Vessel drink continuation | Re-evaluate after each drink        | Vessel: drink once, re-evaluate (source is finite). Terrain: existing continuation until sated (infinite source). |
| Ground vessel drinking | Drink in place, don't pick up         | Character moves to ground vessel and drinks from it like a small water source. Multiple characters can share. |
| Earlier vessel thirst trigger | Deferred to Slice 9              | Keep same thirst threshold for now. Distance=0 already prioritizes carried vessel. Tune threshold in Slice 9. |
| Multi-phase action pattern | Option C: self-managing + shared helper | Self-managing actions own lifecycle (proven by ActionFillVessel). Vessel procurement extracted into shared picking.go helper, not duplicated per action. Option B (central ActionPickup routing) rejected — fights self-managing pattern, scales poorly. |

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
- Add shells to `GetItemTypeConfigs()` with `ShellColors`, non-edible, `SpawnCount: ShellSpawnCount`. Add `NonPlantSpawned Bool` to skip in `SpawnItems()`
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

#### Design: Tilled Soil as Map State

Tilled soil is modeled as **map terrain state** (like water), not a feature. Rationale:
- Tilled soil is a terrain modification, not a discrete entity
- Items must be able to exist on tilled tiles (plants grow there)
- Features block `IsEmpty()` / `findEmptySpot()`, which would prevent plant spawning
- O(1) lookups via `tilled map[Position]bool` on Map struct
- Serialization: `TilledPositions []Position` in SaveState (same pattern as `WaterTiles`)

**Visual:** Olive-colored `≡` (triple horizontal lines) on empty tilled tiles. No background color — the symbol itself renders in olive. When items/characters occupy a tilled tile, the entity renders normally (tilled state visible via details panel). Future: Slice 6 adds underline styling for plants on tilled soil.

**Details panel:** Cursor over tilled tile with no item shows "Tilled soil". Cursor over item on tilled tile shows item details + "On tilled soil" note. Cursor over tile marked-for-tilling (from a pending order) shows "Marked for tilling" in olive text.

#### Design: Activity Category Field

Add `Category string` field to `Activity` struct. Replaces the prefix-convention approach used by Craft (which checks first 5 chars of activity ID). Benefits:
- Explicit grouping for UI menu hierarchy (Garden, Craft, future: Pigment, Construction)
- Activity IDs stay descriptive without encoding category (`"tillSoil"` not `"gardenTillSoil"`)
- `getOrderableActivities()` groups by Category instead of string prefix hacking
- Migrate existing craft activities: set `Category: "craft"` on craftVessel, craftHoe

Order UI flow: Step 0 shows categories (Harvest, Craft, Garden). Step 1 shows activities within category. For Garden, step 1 shows "Till Soil" (later: Plant, Water Garden). Selecting "Till Soil" transitions to area selection mode instead of creating order immediately.

#### Design: Area Selection UI (Reusable for Construction)

Rectangle-based selection, keyboard-driven:

1. Player selects Garden > Till Soil from orders panel → orders panel closes, **area select mode** activates
2. Player moves cursor with arrow keys to desired start position, presses **Enter** to set anchor
3. Player moves cursor — rectangle highlights between anchor and cursor in olive background (lipgloss background color)
4. Invalid tiles within rectangle are **silently excluded** (water, features, already-tilled). They don't highlight.
5. Tiles marked-for-tilling by prior orders show as **pre-highlighted** in olive so player sees existing work
6. Press **Enter** again to confirm — creates order with valid `TargetPositions`, exits area select mode
7. Press **Esc** at any point to cancel and return to normal play

**Implementation notes for reuse by Construction:**
- Area select state on Model: `areaSelectMode bool`, `areaSelectAnchor *Position`, `areaSelectActivityID string`
- Validation function is pluggable: `isValidTillTarget(pos)` for tilling, future `isValidFenceTarget(pos)` for construction
- The pattern is: enter area select → anchor → rectangle → filter invalid → confirm → create order with positions

#### Design: Marked-for-Tilling Pool (Decoupled from Orders)

Marked tiles and till orders are **separate concepts**:

- **Marked tiles** = the user's tilling plan. Stored as `markedForTilling map[Position]bool` on Map. Persists independently of orders. Area selection adds to the pool. "Unmark Tilling" removes from the pool. Serialized alongside `tilled` in SaveState.
- **Till Soil orders** = worker assignments. A tillSoil order means "assign a character to work the tilling pool." The order carries no positions — it just represents a worker slot. Cancelling an order removes a worker but leaves the plan intact.
- **Unmark Tilling** = separate Garden menu option using the same area selection rectangle UI but in "unmark" mode. Removes selected positions from the pool. Does not create or cancel orders.

**Work assignment:** Characters with any tillSoil order pick the nearest marked-but-not-yet-tilled tile from the shared pool. Multiple characters naturally share the work.

**Order completion:** A tillSoil order completes when the pool is empty (all marked tiles have been tilled). All active tillSoil orders complete simultaneously.

**Source of truth:** Map stores both the plan (`markedForTilling`) and reality (`tilled`). When a tile is tilled, it leaves `markedForTilling` and enters `tilled`.

#### Design: EnsureHasItem (Hoe Procurement)

New helper in `picking.go`, following `EnsureHasRecipeInputs` pattern:

1. Check if character already carries item of requested type → return nil (ready)
2. No inventory space? Drop non-target loose items
3. Find nearest item of type on map → return pickup intent
4. Nothing available → return nil (triggers abandonment by caller)

Simpler than `EnsureHasRecipeInputs` — single item type, no multi-component logic.

#### Design: Till Soil Activity

- Activity ID: `"tillSoil"`, Name: `"Till Soil"`, Category: `"garden"`
- IntentFormation: IntentOrderable, Availability: AvailabilityKnowHow
- Discovery triggers: look at hoe, pick up hoe
- No recipe needed (similar to harvest)

#### Design: findTillSoilIntent Flow

1. `EnsureHasItem("hoe")` — procure hoe if not carried
2. Find nearest marked-but-not-tilled position from the shared pool (`markedForTilling` where `!IsTilled()`)
3. If no such position exists → return nil (pool empty, order complete)
4. If not at target position → return movement intent (BFS pathfinding)
5. If at target position → return `ActionTillSoil` intent

#### Design: Till Soil Action Execution

- New `ActionTillSoil` action type, duration `ActionDurationMedium` (4.0s / ~48 world minutes)
- On completion:
  - `Map.SetTilled(pos)` marks tile as tilled
  - `Map.UnmarkForTilling(pos)` removes tile from the pool
  - Growing items at position: destroyed (tilling kills wild plants)
  - Non-growing items at position: displaced to adjacent empty tile via `findEmptyAdjacent()`. If no adjacent space, drop at character position.
  - Character automatically finds next marked-but-not-tilled position and continues
- Order completion: when `findTillSoilIntent` returns nil because the marked-for-tilling pool is empty. All active tillSoil orders complete simultaneously.

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → add minimal stubs to compile → verify red → implement → verify green → human testing checkpoint.

##### Step 1: Tilled Soil Map State + Serialization ✅

**Tests** (in `internal/game/map_test.go`):
- `SetTilled()` / `IsTilled()` basic set/get
- `IsTilled()` returns false for non-tilled positions
- `TilledPositions()` returns all tilled positions
- `IsBlocked()` returns false for tilled tiles (doesn't block movement)
- `IsEmpty()` returns true for tilled tiles with no item/character/feature (items can exist on tilled soil)

**Serialization test** (in `internal/game/map_test.go` — no separate serialize_test.go exists):
- Tilled positions round-trip through save/load (test via `TilledPositions()` after set)

**Implementation** (follows water terrain pattern exactly):
- Add `tilled map[types.Position]bool` to Map struct, initialize in `NewMap()`
- `SetTilled(pos)` — sets `m.tilled[pos] = true`
- `IsTilled(pos) bool` — returns `m.tilled[pos]`
- `TilledPositions() []types.Position` — iterates map, returns slice
- Key: `IsBlocked()` does NOT check tilled (tilled tiles are walkable). `IsEmpty()` does NOT check tilled (items can exist on tilled soil).
- `SaveState`: add `TilledPositions []types.Position` field with `json:"tilled_positions,omitempty"`
- `ToSaveState`: serialize via `tilledPositionsToSave(gameMap)` helper
- `FromSaveState`: restore via `gameMap.SetTilled(pos)` loop

##### Step 2: Tilled Soil Rendering + Details ✅

**Tests:** None (UI rendering per CLAUDE.md)

**Config:** `CharTilledSoil = '═'` in config.go

**Implementation:**
- Reuse `growingStyle` for tilled soil rendering (all gardening visuals share one olive). Change `growingStyle` color from `"106"` to `"142"` (yellow-olive, distinct from leaf pile's `"106"`).
- Terrain fill rendering: tilled soil uses Option F style — `═══` full fill for empty tiles, `═●═` fill around entities on tilled soil. Restructured `renderCell()` to support terrain fill padding (reusable for future water `███`, structure walls `▓▓▓`).
- In `renderDetails()`: when cursor is on tilled tile with no entity, show "Tilled soil". When cursor is on item on tilled tile, show item details + "On tilled soil".

**[TEST] Checkpoint — Tilled soil state + rendering:**
- `go test ./...` passes
- Use test world or temporary code to set tiles as tilled, verify olive `≡` renders
- Cursor over tilled tile shows "Tilled soil" in details
- Save/load preserves tilled state

##### Step 3: Category Field on Activity + Garden Category UI ✅

**Design decisions:**
- **Category field replaces prefix hacking**: Add `Category string` to `Activity` struct. `isCraftActivity(id[:5] == "craft")` replaced by `activity.Category != ""`. Any category works generically.
- **Synthetic category entries**: `getOrderableActivities()` generates synthetic `Activity{ID: "craft", Name: "Craft"}` (and `"garden"/"Garden"`) when any activity in that category is known. Step 0 shows these + uncategorized activities.
- **`getCraftActivities()` → `getCategoryActivities(category)`**: Generic helper returns all known activities in a given category.
- **Knowledge panel**: Categorized activities display as `"Category: Name"` (e.g., "Craft: Vessel", "Garden: Till Soil").
- **DisplayName**: Uses Category-aware formatting. Garden orders show activity name only (e.g., "Till Soil"). Craft orders show "Craft vessel". Harvest unchanged.
- **Till Soil selection**: Selecting Till Soil in step 1 creates a basic order (no TargetPositions yet). Step 4 replaces this path with area selection.
- **tillSoil activity**: Category "garden", IntentOrderable, AvailabilityKnowHow, discovery triggers from hoe (ActionLook + ActionPickup with ItemType "hoe").

**Tests** (in `internal/entity/activity_test.go`):
- `tillSoil` activity registered with correct Category, IntentFormation, Availability, DiscoveryTriggers
- `craftVessel` and `craftHoe` activities have `Category: "craft"`
- `DisplayName()` for garden order returns "Till soil" (activity name, lowercased)
- `DisplayName()` for craft order still returns "Craft vessel"
- `DisplayName()` for harvest order still returns "Harvest berries"

**Implementation:**
- Add `Category string` to Activity struct
- Set `Category: "craft"` on existing craftVessel, craftHoe activities
- Register `tillSoil` activity with `Category: "garden"`, discovery triggers from hoe
- Replace `isCraftActivity` prefix check → `activity.Category != ""` / `activity.Category == X`
- Generalize `getOrderableActivities()`: collect categories with known activities, generate synthetic entries
- Rename `getCraftActivities()` → `getCategoryActivities(category string)`
- Update `Order.DisplayName()` to use Category field
- Update order add UI (view.go + update.go): step 0 shows categories + uncategorized, step 1 shows activities within category. Garden step 1 selecting Till Soil creates a basic order for now.
- Update knowledge panel: categorized activities prefixed with `"Category: "`

**[TEST] Checkpoint — Garden category visible:**
- `go test ./...` passes
- Existing Craft ordering still works (no regression from Category refactor)
- Character discovers Till Soil from looking at/picking up hoe
- Orders panel shows "Garden" when Till Soil is known
- Select Garden → "Till Soil" appears
- Selecting Till Soil creates an order (placeholder — Step 4 replaces with area selection)
- Knowledge panel shows "Garden: Till Soil" and "Craft: Vessel" etc.

##### Step 4: Marked-for-Tilling Pool on Map ✅

**Tests** (in `internal/game/map_test.go`):
- `MarkForTilling(pos)` / `IsMarkedForTilling(pos)` basic set/get
- `IsMarkedForTilling()` returns false for unmarked positions
- `UnmarkForTilling(pos)` removes position from pool
- `MarkedForTillingPositions()` returns all marked positions
- `MarkForTilling` on already-tilled position is a no-op (returns false or silently skips)
- `IsBlocked()` returns false for marked tiles (walkable)
- `IsEmpty()` returns true for marked tiles with no occupants

**Tests** (serialization, in `internal/game/map_test.go`):
- Marked-for-tilling positions round-trip through save/load

**Implementation:**
- Add `markedForTilling map[Position]bool` to Map struct, initialize in `NewMap()`
- `MarkForTilling(pos)`, `UnmarkForTilling(pos)`, `IsMarkedForTilling(pos)`, `MarkedForTillingPositions()`
- `MarkForTilling` skips positions that are already tilled
- `SaveState`: add `MarkedForTillingPositions []types.Position` with `json:"marked_for_tilling,omitempty"`
- `ToSaveState` / `FromSaveState`: serialize/restore marked positions

**[TEST] Checkpoint — Pool state:**
- `go test ./...` passes

##### Step 4b: Area Selection Validation Logic ✅

**Tests** (in `internal/ui/area_select_test.go`):
- `isValidTillTarget(pos, gameMap)` returns false for water tiles
- `isValidTillTarget(pos, gameMap)` returns false for feature tiles
- `isValidTillTarget(pos, gameMap)` returns false for already-tilled tiles
- `isValidTillTarget(pos, gameMap)` returns false for already-marked tiles
- `isValidTillTarget(pos, gameMap)` returns true for empty walkable tiles
- `isValidUnmarkTarget(pos, gameMap)` returns true for marked-but-not-tilled tiles
- `isValidUnmarkTarget(pos, gameMap)` returns false for unmarked or already-tilled tiles
- `getValidPositions(anchor, cursor, gameMap, validatorFn)` filters rectangle to valid positions
- `getValidPositions` excludes out-of-bounds positions
- `getValidPositions` handles anchor > cursor (any drag direction)

**Implementation:**
- New file `internal/ui/area_select.go`
- `isValidTillTarget(pos, gameMap)` — not water, not feature, not tilled, not already marked
- `isValidUnmarkTarget(pos, gameMap)` — is marked and not yet tilled
- `getValidPositions(anchor, cursor, gameMap, validatorFn)` — generic rectangle filter, reusable for Construction

**[TEST] Checkpoint — Validation logic:**
- `go test ./...` passes

##### Step 4c: Area Selection UI + Order Flow ✅

**Tests:** None (UI rendering/keyboard handling per CLAUDE.md)

**Implementation — Area selection as order-add step:**
- Add state to Model: `areaSelectAnchor *types.Position`, `areaSelectUnmarkMode bool`
- Detect in order-add flow: when user selects "Till Soil" in step 1, advance to step 2 (area selection). When user selects "Unmark Tilling", advance to step 2 in unmark mode.
- Register `unmarkTilling` as a non-orderable UI-only activity under garden category (or handle as special case in the order UI — decide during implementation)
- Orders panel sidebar remains visible during step 2, showing keypress hints:
  - Before anchor: "Arrow keys: move cursor / Enter: set anchor / Esc: cancel"
  - After anchor: "Arrow keys: resize / Enter: confirm / Esc: cancel"

**Implementation — Keyboard handling in step 2:**
- Arrow keys: move cursor (reuse existing `moveCursor`)
- Enter (no anchor): set anchor at cursor position
- Enter (anchor set): confirm selection
  - Mark mode: add valid positions to `markedForTilling` pool, create tillSoil order, return to step 0
  - Unmark mode: remove valid positions from `markedForTilling` pool, return to step 0 (no order created)
- Esc: cancel, clear anchor, return to step 0

**Implementation — Map rendering during area selection:**
- `renderCell()`: when in area selection step with anchor set, highlight valid tiles in rectangle with olive lipgloss background
- Pre-highlight existing marked-for-tilling tiles (from pool) so user sees prior work
- Unmark mode: highlight tiles that would be unmarked (different visual — maybe dim/strikethrough, or just show which marked tiles are in the rectangle)

**Implementation — Details panel:**
- During area selection: show "Marked for tilling" in olive for marked tiles under cursor
- In normal select view: cursor over marked-but-not-tilled tile shows "Marked for tilling" in details

**[TEST] Checkpoint — Area selection mark flow:**
- `go test ./...` passes
- Build and run. Use `/test-world` or give a character Till Soil know-how
- Orders → Garden → Till Soil → sidebar shows hints, cursor moves on map
- Press Enter to anchor, move cursor — rectangle highlights valid tiles in olive
- Water, feature, already-tilled tiles NOT highlighted within rectangle
- Existing marked tiles show as pre-highlighted
- Enter confirms → marked tiles added to pool, "Till Soil" order appears in orders panel
- Esc cancels → returns to order menu, no tiles marked
- Cursor over marked tile in select view shows "Marked for tilling" in details panel
- Save, load → marked tiles persist

**[TEST] Checkpoint — Area selection unmark flow:**
- Orders → Garden → Unmark Tilling → area selection in unmark mode
- Rectangle highlights currently-marked tiles that would be removed
- Enter confirms → tiles removed from pool
- Previously marked tiles no longer highlighted
- If no marked tiles exist, unmark selection confirms with no effect

**[TEST] Checkpoint — Multiple mark operations:**
- Mark a 3x3 area → 9 tiles marked, one order created
- Mark another 4x4 area nearby → new tiles added to pool, second order created
- Both orders show as "Till Soil" in orders panel
- Cancel one order → tiles remain marked, one worker removed
- Unmark a few tiles → those tiles removed from pool

##### Step 5: EnsureHasItem + findTillSoilIntent ✅

**Tests** (in `internal/system/picking_test.go`):
- `EnsureHasItem("hoe")` returns nil when character already carries hoe
- `EnsureHasItem("hoe")` returns pickup intent when hoe on map
- `EnsureHasItem("hoe")` drops non-hoe loose items to make space
- `EnsureHasItem("hoe")` returns nil when no hoe exists (triggers abandonment)

**Tests** (in `internal/system/order_execution_test.go`):
- `findTillSoilIntent` returns hoe procurement intent when no hoe carried
- `findTillSoilIntent` returns movement intent toward nearest marked-but-not-tilled position
- `findTillSoilIntent` returns `ActionTillSoil` when at marked position with hoe
- `findTillSoilIntent` returns nil when pool is empty (all marked tiles tilled — order complete)
- `findTillSoilIntent` skips already-tilled positions in pool

**Implementation:**
- `EnsureHasItem(char, itemType, items, gameMap, log)` in picking.go
- Add `ActionTillSoil` to ActionType enum in character.go
- `findTillSoilIntent()` in order_execution.go — works from `gameMap.MarkedForTillingPositions()`
- Wire into `findOrderIntent()` switch: `case "tillSoil": return findTillSoilIntent(...)`

**[TEST] Checkpoint — Intent logic:**
- `go test ./...` passes

##### Step 6: Till Soil Action Execution ✅

**Tests** (in `internal/system/tilling_test.go` for pure logic, `internal/ui/update_test.go` for action handler):
- `ActionTillSoil` sets tile as tilled after `ActionDurationMedium`
- `ActionTillSoil` removes tile from marked-for-tilling pool
- Growing items at target position destroyed on tilling
- Non-growing items at target position displaced to adjacent empty tile
- Non-growing items dropped at character position when no adjacent space
- Individual character's tillSoil order completes when no marked-but-not-tilled positions remain (checked after tilling)
- `selectOrderActivity` completes (not abandons) tillSoil order when pool is empty (e.g., another worker finished the last tile)

**Design clarification (from discussion):** Each character's tillSoil order completes independently when they find no more work — no batch "complete all orders" logic. Multiple workers naturally finish on their own cadence: each finishes their current tile, looks for more, finds none, completes. This avoids workers waiting around for the last tile being worked by someone else.

**Implementation:**
- Handle `ActionTillSoil` in `applyIntent()` (same progress pattern as `ActionCraft`)
- On completion:
  - `gameMap.SetTilled(pos)` — marks tile as tilled terrain
  - `gameMap.UnmarkForTilling(pos)` — removes from pool
  - Check for item at pos: growing → remove, non-growing → displace to adjacent via `FindEmptyAdjacent` (exported from lifecycle.go)
  - Log action
  - Character's intent clears → next tick, `findTillSoilIntent` either finds more work or returns nil
- Order completion: in `selectOrderActivity`, when `findOrderIntent` returns nil for tillSoil, check if pool has untilled marked positions. None → `CompleteOrder` (success). Some → `abandonOrder` (no hoe or other blocker).

**[TEST] Checkpoint — Full Till Soil flow:**
- `go test ./...` passes
- Start world, craft a hoe (or `/test-world` with hoe pre-placed)
- Character discovers Till Soil from looking at/picking up hoe
- Order Garden > Till Soil, select area on map
- Character picks up hoe, moves to first marked tile, tills it (olive `═` appears, tile removed from marked pool)
- Character moves to next marked tile automatically
- Growing items destroyed during tilling, non-growing items displaced to adjacent tile
- Order completes when all marked tiles tilled
- Test with 2 till orders (2 workers) — both characters share the pool, work completes faster
- Cancel one order mid-work — remaining worker continues, marked tiles unchanged
- Unmark some tiles mid-work — worker skips newly-unmarked tiles
- Save/load preserves tilled tiles, marked-for-tilling pool, and orders

**[DOCS]** ✅ Update README (Latest Updates), CLAUDE.md, game-mechanics, architecture as needed for area selection UI pattern (document for Construction reuse), tilled soil system, marked-for-tilling pool, Garden order category.

**[RETRO]** Run /retro.

---

### Slice 4: Unfulfillable Orders

**Reqs covered:**

- Orders Unfulfillable Logic (lines 137-148)

**Outcomes:**

1. Characters skip orders they can't complete (move to next open order)
2. Unfulfillable detection: no one has required know-how, OR required components don't exist in world
3. Unfulfillable status shown on orders screen
4. Characters don't assign themselves to unfulfillable orders (no assign/abandon loop)

#### Design: Feasibility as Computed Property

Unfulfillable is **not a persisted order status** — it's computed on the fly from current world state. A Till Soil order becomes feasible the moment someone crafts a hoe. No dirty flags, no event system, no cache invalidation.

`IsOrderFeasible(order, items, gameMap)` is called:
- At assignment time: `findAvailableOrder` skips unfeasible orders
- At render time: orders panel shows dimmed label for unfeasible orders

The check is O(n) over items + characters, which is negligible on a 60x60 map.

#### Design: Two Failure Modes

| Failure | Label | When |
|---------|-------|------|
| Components missing | `[Unfulfillable]` | Required items don't exist in the world |
| Know-how gap | `[No one knows how]` | No living character has learned the activity |

Know-how failure is only reachable via character death (order creation already requires know-how). Included for future-proofing and debugging — essentially free to implement.

Both rendered in dimmed style in the orders panel.

#### Design: Per-Activity Feasibility Checks

Each order type checks only its **direct requirements** — no recursive dependency chains. The player manages the dependency chain (order Craft Hoe before Till Soil if no hoe exists).

- **Harvest**: Growing items of target type exist on map
- **Craft** (recipe-based): At least one recipe has all input types present in the world (ground items + character inventories)
- **Till Soil**: Hoe exists in world (ground or inventory) AND marked-for-tilling pool has untilled positions

#### Implementation Steps

Each step follows the TDD cycle: write tests → implement → verify green.

##### Step 1: IsOrderFeasible + Helper Functions ✅

**Tests** (in `internal/system/order_execution_test.go`):
- Harvest feasible when growing target items exist on map
- Harvest unfeasible when no growing target items exist
- Craft feasible when all recipe inputs exist in world (ground or inventory)
- Craft unfeasible when a recipe input is missing from world
- TillSoil feasible when hoe exists (ground) and marked positions exist
- TillSoil feasible when hoe exists (character inventory) and marked positions exist
- TillSoil unfeasible when no hoe exists in world
- TillSoil unfeasible when no marked-for-tilling positions exist
- Know-how: unfeasible when no living character knows the activity
- Know-how: feasible when at least one character knows it
- Returns `noKnowHow=true` for know-how failures, `false` for component failures

**Implementation** (`internal/system/order_execution.go`):
- `IsOrderFeasible(order, items, gameMap) (feasible bool, noKnowHow bool)` — exported, gets characters via `gameMap.Characters()`
- `itemExistsInWorld(itemType, chars, items)` — checks character inventories + ground items
- `growingItemExists(items, itemType)` — harvest-specific (growing plants only)
- `isAnyRecipeWorldFeasible(activityID, chars, items)` — all inputs exist for at least one recipe

##### Step 2: Wire into findAvailableOrder ✅

**Tests** (in `internal/system/order_execution_test.go`):
- Character skips unfeasible order and takes next feasible one
- Character takes no order when all orders are unfeasible

**Implementation:**
- `findAvailableOrder` gains `items []*entity.Item` and `gameMap *game.Map` parameters
- Adds `IsOrderFeasible` check alongside existing `canExecuteOrder` check
- Update `selectOrderActivity` call site (already has these params)

##### Step 3: UI Rendering (no tests per CLAUDE.md) ✅

**Implementation** (`internal/ui/view.go`, `internal/ui/styles.go`):
- In `renderOrdersContent`: call `system.IsOrderFeasible` for each order
- Unfeasible orders: dimmed style + `[Unfulfillable]` or `[No one knows how]` label
- Dimmed style for the entire order line (name + label)

**[TEST] Checkpoint — Full unfulfillable flow:**
- `go test ./...` passes
- Create a Till Soil order when no hoe exists → shows `[Unfulfillable]`, dimmed, characters skip it
- Create a Craft Hoe order → characters craft hoe → Till Soil becomes feasible, characters take it
- Create Harvest berries when no berries growing → `[Unfulfillable]`
- Berry grows back → order becomes feasible again
- No more assign/abandon log spam for unfeasible orders
- Save, load — orders still display correct feasibility (computed, not persisted)

**[DOCS]** Update README (Latest Updates), CLAUDE.md, game-mechanics, architecture as needed.

**[RETRO]** Run /retro.

[Pause for evaluation before continuing for Part II, opportunity to pull in quick wins or opportunistic random items.]
---

## Part II: Seeds, Planting, and Watering

### Feature Set 2: Seeds, Plantable Attribute, and Food Satiation

**Reqs covered:**
- New ItemType: Seed (lines 65-77) [Descoped: flower seeds deferred to randomideas.md — mechanic doesn't fit existing activity archetypes yet]
- New Item Attribute: Plantable (lines 79-82)

#### Design: Plantable Attribute

Explicit `Plantable bool` on Item struct. (See Decisions Log.) Only berries and mushrooms become plantable when picked — they ARE the plantable item (no seeds needed). Gourds and flowers do NOT become plantable when picked. Gourds produce seeds when eaten instead. Flower seed collection deferred (see randomideas.md).

#### Design: Seed Item Type

Seeds use the Kind pattern: `ItemType: "seed"`, `Kind: "<parent> seed"` (e.g., `Kind: "gourd seed"`). `Description()` naturally produces "warty green gourd seed" from attributes + Kind.

- Symbol: `.` in parent's color
- Inherits parent's full variety (Color, Pattern, Texture)
- Not edible, `Plant: nil`, `Plantable: true`
- Stack size 20, same-variety stacking only (via vessel system)

**Seed varieties in VarietyRegistry**: Registered at world gen alongside parent varieties. A "warty green gourd" variety generates a corresponding "seed" variety with the same color/pattern/texture. This keeps existing vessel stacking infrastructure working (`AddToVessel` uses `VarietyRegistry.GetByAttributes`).

#### Design: Gourd Seed Creation

When consumption completes on a gourd, create a seed with the gourd's variety and **auto-drop** to ground at character's position (same pattern as crafted items). No pickup logic chain — seed is immediately available to any character doing seed-related tasks.

For `ConsumeFromVessel()`, seed inherits variety from the stack's `ItemVariety`.

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → implement → verify green → human testing checkpoint.

##### Step 1: Plantable Field on Item + Serialization ✅

**Tests** (in `internal/entity/item_test.go`):
- `NewBerry` has `Plantable: false` (only becomes plantable when picked)
- `NewMushroom` has `Plantable: false`
- `NewGourd` has `Plantable: false` (gourds are never directly plantable)
- `NewFlower` has `Plantable: false`

**Tests** (in `internal/system/picking_test.go`):
- `Pickup` sets `Plantable = true` for berry items (ItemType "berry")
- `Pickup` sets `Plantable = true` for mushroom items (ItemType "mushroom")
- `Pickup` does NOT set `Plantable` for gourd items
- `Pickup` does NOT set `Plantable` for flower items
- Pickup to vessel also sets `Plantable = true` for berries/mushrooms

**Implementation:**
- Add `Plantable bool` to Item struct
- In `Pickup()` (picking.go): set `Plantable = true` when `item.ItemType == "berry" || item.ItemType == "mushroom"`, alongside existing `IsGrowing = false` logic (both the vessel path and inventory path)
- `ItemSave`: add `Plantable bool` field with `json:"plantable,omitempty"`
- `serialize.go`: wire `Plantable` through both save paths (ground items + inventory) and `itemFromSave`
- Bug fix: inventory save path is missing `Kind` field (ground items path has it). Fix alongside Plantable addition.
- Details panel: show "Plantable" when `item.Plantable == true`

**[TEST] Checkpoint — Plantable attribute:**
- `go test ./...` passes
- Build and run: pick up a berry, check details — shows "Plantable"
- Pick up a gourd, check details — does NOT show "Plantable"
- Pick up a flower, check details — does NOT show "Plantable"
- Save, load — plantable state preserved

##### Step 2: Seed Item Type + Varieties ✅

**Tests** (in `internal/entity/item_test.go`):
- `NewSeed` returns correct properties (symbol `.`, ItemType `"seed"`, Kind `"gourd seed"`, color/pattern/texture from parent, not edible, `Plant == nil`, `Plantable == true`)
- `NewSeed` description: "warty green gourd seed" for a warty green gourd parent

**Tests** (in `internal/game/variety_generation_test.go` or existing test file):
- Seed varieties registered in VarietyRegistry for each edible parent variety
- `GetByAttributes("seed", color, pattern, texture)` returns correct seed variety

**Implementation:**
- `NewSeed(x, y int, parentItemType string, color types.Color, pattern types.Pattern, texture types.Texture)` in item.go
- `CharSeed = '.'` in config.go
- `"seed"` entry in `StackSize` map (20)
- Register seed varieties in `variety_generation.go` — for each gourd variety, generate corresponding seed variety
- `itemFromSave` handles ItemType `"seed"` (no special case needed — generic path already works with Kind)
- Rendering: `.` in parent's color (uses existing color switch in `styledSymbol()`)

**[TEST] Checkpoint — Seed item type:**
- `go test ./...` passes
- No human testing yet (seeds aren't created in gameplay until Step 3)

##### Step 3: Gourd Seed Creation on Consumption ✅

**Tests** (in `internal/system/consumption_test.go`):
- `Consume` on gourd creates a seed at character position with gourd's variety attributes
- `Consume` on gourd: seed has `Kind: "gourd seed"`, `Plantable: true`
- `Consume` on berry does NOT create a seed
- `Consume` on mushroom does NOT create a seed
- `ConsumeFromInventory` on gourd creates a seed at character position
- `ConsumeFromVessel` on gourd creates a seed at character position with stack variety

**Implementation:**
- In `Consume()`: after removing item from map, if `item.ItemType == "gourd"`, create seed via `NewSeed()` at character position, add to map
- In `ConsumeFromInventory()`: same check, create seed at character position
- In `ConsumeFromVessel()`: if variety ItemType is "gourd", create seed from variety attributes at character position
- All three: seed auto-drops to ground (no pickup logic)

**[TEST] Checkpoint — Full gourd seed flow:**
- `go test ./...` passes
- Build and run: character eats a gourd (from ground, inventory, or vessel). Seed appears on ground at character's position with correct variety description (e.g., "warty green gourd seed")
- Cursor over seed shows correct description + "Plantable" in details
- Seeds can be picked up and added to vessels (stacking works)
- Save, load — seeds persist correctly

**Bug fix:** Removed redundant "Tilling soil" log message that fired each time a character started tilling a new tile. Only "Tilled soil" (on completion) remains.

**[DOCS]** ✅ Update README (Latest Updates), CLAUDE.md, game-mechanics, architecture as needed.

**[RETRO]** Run /retro if there were any interruptions

---

### Slice 6: Plant Activity, Sprout Phase, and Growth Mechanics

**Reqs covered:**
- Quick change: Water symbol (line 63)
- New Activity: Plant (lines 84-99)
- Enhanced Logic: Garden Plant Growth and Reproduction (lines 125-131)
- Growth speed tiers (lines 138-141) — flat 30s sprout for testing; tier differentiation deferred to Slice 9

**Resolved feature questions:**
- Sprout symbol: `𖧧` (bold style) ✓
- Occupied tilled tile: character skips it, finds next empty tilled tile ✓
- Underline styling: skip, rely on `═X═` fill pattern ✓
- Sprout → full plant: in-place symbol change, derive mature symbol from ItemType ✓
- Growth bonus values: simple multipliers, exact values tunable in Slice 9 ✓

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → add minimal stubs to compile → verify red → implement → verify green → human testing checkpoint.

##### Step 0: Water Symbol + Wet Tiles ✅

**Tests** (in `internal/game/map_test.go`):
- `IsWet(pos)` returns true for tiles 8-directionally adjacent to water
- `IsWet(pos)` returns false for tiles not adjacent to water
- `IsWet(pos)` returns false for water tiles themselves (impassable, nothing grows on them)

**Config:** Change `CharWater = '▓'`

**Implementation** (`internal/game/map.go`, `internal/config/config.go`, `internal/ui/view.go`):
- Change `CharWater` constant to `'▓'` (dark shade block)
- Add `IsWet(pos Position) bool` to Map — checks 8 neighbors for water
- Rendering: pond symbol changes automatically from constant
- Details panel: show "Wet" when cursor is on a wet tile (not water itself)

**[TEST] Checkpoint — Water symbol + wet tiles:**
- `go test ./...` passes
- Build and run: ponds appear as blue `▓` blocks instead of `≈`
- Springs still appear as `☉` (unchanged)
- Cursor over tile adjacent to water shows "Wet" in details
- Cursor over tile far from water does NOT show "Wet"
- No functional changes to drinking, pathfinding

**Reqs reconciliation:** Line 63 _"convert water symbol to medium-shade tile"_ — `▓` is dark shade block. Line 131 _"tiles adjacent (including diagonal) to water sources are always 'wet'"_ ✓.

**Bug fixed during testing:** The pond fill pattern must be `▓▓▓` (three characters) to match the terrain fill system width, not a single `▓`. A single character left two empty-looking cells flanking the symbol, breaking visual consistency with tilled soil `═══`. Fixed by rendering pond tiles as `▓▓▓`.

**[DOCS]** Update README, game-mechanics as needed.

**[RETRO]** Run /retro.

##### Step 1a: Plant Activity Registration + Order UI ✅

**Tests** (in `internal/entity/activity_test.go`):
- `plant` activity registered with Category `"garden"`, IntentOrderable, AvailabilityKnowHow
- `plant` activity has DiscoveryTriggers with `RequiresPlantable: true` for ActionLook and ActionPickup

**Tests** (in `internal/system/preference_test.go` or knowledge test):
- Know-how discovery fires when looking at plantable item with RequiresPlantable trigger
- Know-how discovery does NOT fire for non-plantable item with RequiresPlantable trigger

**Tests** (in `internal/system/order_execution_test.go`):
- Plant order unfulfillable when no plantable items of target type exist in world
- Plant order feasible when plantable items of target type exist

**Implementation:**
- Add `RequiresPlantable bool` to `DiscoveryTrigger` struct in activity.go
- Update `triggerMatches()` in discovery.go to handle `RequiresPlantable` (check `item.Plantable`)
- Register `plant` activity: ID `"plant"`, Name `"Plant"`, Category `"garden"`, IntentOrderable, AvailabilityKnowHow, triggers on ActionLook + ActionPickup with RequiresPlantable
- Add `LockedVariety string` to Order struct, `json:"locked_variety,omitempty"` in save state
- Add `Plantable bool` and `CanProduceSeeds bool` to `ItemTypeConfig` in variety_generation.go. Berry + mushroom get `Plantable: true`. Gourd gets `CanProduceSeeds: true`.
- `getPlantableTypes()` derives menu entries from config: types with `Plantable` appear as-is ("Berries", "Mushrooms"); types with `CanProduceSeeds` appear as "{type} seeds" ("Gourd seeds"). Per reqs line 89: `Garden > Plant > Select Plantable ItemType (gourd seeds, flower seeds, mushrooms, berries, future: nuts)`.
- Order UI: selecting "Plant" in garden step 1 → reuse step 2 for plantable type sub-menu (branch on activity ID: tillSoil → area selection, plant → type list). Selecting creates order with TargetType set (ItemType for berry/mushroom, Kind for seeds e.g. `"gourd seed"`).
- `DisplayName()`: plant activity uses harvest-style pattern → "Plant gourd seeds", "Plant berries", "Plant mushrooms"
- Unfulfillable check: plant order unfulfillable when no plantable items of target type exist in world. Match on `item.Plantable && (item.ItemType == targetType || item.Kind == targetType)`.
- Serialization: LockedVariety in order save/load
- **Esc back-navigation fix (orders only):** Align Esc behavior within the order add flow with the "esc = back one level" principle from `randomideas.md` #2. Currently Esc at step 1 exits add mode entirely; fix so step 2 → step 1 → step 0 → exit add mode. This scopes the fix to orders only; the broader cross-panel Esc cleanup remains a separate future task.

**[TEST] Checkpoint — Plant activity registration + order UI:**
- `go test ./...` passes
- Build and run: character discovers Plant from looking at/picking up plantable items
- Knowledge panel shows "Garden: Plant"
- Orders → Garden → Plant → sub-menu shows kind-level types: "Gourd seeds", "Berries", "Mushrooms"
- Selecting one creates order showing "Plant gourd seeds" / "Plant berries" etc. in orders panel
- Plant order shows [Unfulfillable] when no plantable items exist
- Esc navigates back through sub-menus correctly (step by step, not jumping out)
- Esc back-navigation also works for existing activities (Till Soil, Craft, Harvest)
- Order can be cancelled normally
- Note: ordering Plant creates the order but character can't execute it yet (Step 1b)

**[DOCS]** Update README, game-mechanics as needed.

**[RETRO]** Run /retro.

##### Step 1b: Sprout Creation + findPlantIntent + ActionPlant ✅

**Investigation resolved: Plantable flag and vessel extraction.** `ConsumeAccessibleItem` reconstructs items from vessel stacks but only copies ItemType/Color/Pattern/Texture — it does NOT restore `Edible`, `Sym`, `Plantable`, or `Kind`. Edible properties appear to survive only because the consumption system handles poison/healing through a separate path. **Fix: Add `Plantable bool` to `ItemVariety`** (mirrors `Edible` pattern), set during variety generation for berry/mushroom/seed varieties. Fix `ConsumeAccessibleItem` to restore `Plantable`, `Edible`, and `Sym` from the variety on extraction. This means items in vessels always know their plantable status via the variety — no config-inference needed.

**Design: Sprout Data Model**

Add `IsSprout bool` and `SproutTimer float64` to PlantProperties. Sprouts have `IsGrowing: true` (they're in the ground) but can't reproduce yet (lifecycle skips them). `SproutTimer` counts down; maturation logic added in Step 2.

**Design: CreateSprout Helper**

Creates a sprout item from a plantable item:
- From seed: ItemType = parent type (extract from Kind: `"gourd seed"` → `"gourd"`), variety from seed. Edible properties looked up from variety registry by parent type + attributes.
- From berry/mushroom: ItemType = same as planted item, variety from item. Edible properties copied directly.
- Sets `Plant: {IsGrowing: true, IsSprout: true, SproutTimer: SproutDuration}`, symbol `CharSprout`

**Design: findPlantIntent Flow**

1. Find nearest empty tilled tile (no item at position). If none → return nil (order complete).
2. Check if character has accessible plantable item of target type (inventory + vessel). If LockedVariety set, filter to matching variety. Accessible check uses `Plantable` on inventory items and `variety.Plantable` on vessel contents. Seed matching: targetType `"gourd seed"` matches items with Kind `"gourd seed"` or vessel stacks where variety is seed type with matching attributes.
3. If no accessible item: seek nearest plantable item of target type on ground. If LockedVariety set, filter to matching variety.
4. If no item findable → return nil (completion if LockedVariety set, abandon if never started).
5. If carrying item but not at tilled tile → move intent (BFS).
6. At tilled tile with plantable item → ActionPlant intent.

**Logging note (from Till Soil retro):** Only log on completion — not on start. Log "Planted [variety]" per tile completion.

**Config:** `CharSprout = '𖧧'`, `SproutDuration = 30.0`

---

**Phase 1: Data model + CreateSprout**

_Tests_ (in `internal/entity/item_test.go`):
- CreateSprout from gourd seed: sprout has ItemType `"gourd"`, IsSprout true, symbol CharSprout, parent variety
- CreateSprout from berry: sprout has ItemType `"berry"`, IsSprout true, symbol CharSprout, parent variety
- CreateSprout from mushroom: sprout preserves edible properties (poisonous/healing)

_Implementation:_
- Add `Plantable bool` to `ItemVariety`; set during generation for berry/mushroom/seed varieties
- Fix `ConsumeAccessibleItem` to restore `Plantable`, `Edible`, `Sym` from variety on extraction
- Add `IsSprout bool`, `SproutTimer float64` to PlantProperties
- Add `CharSprout`, `SproutDuration` constants to config.go
- Add `ActionPlant` to ActionType enum
- `CreateSprout(x, y int, plantedItem *Item, ...)` helper in entity/item.go

**Phase 2: findPlantIntent**

_Tests_ (in `internal/system/order_execution_test.go`):
- `findPlantIntent` returns procurement intent when no plantable item carried
- `findPlantIntent` returns movement intent toward nearest empty tilled tile when carrying plantable
- `findPlantIntent` returns ActionPlant when at empty tilled tile with plantable item
- `findPlantIntent` returns nil when no empty tilled tiles (order complete)
- `findPlantIntent` returns nil when no plantable items available
- `findPlantIntent` with LockedVariety only seeks matching variety
- `findPlantIntent` skips tilled tiles that already have items

_Implementation:_
- `findPlantIntent()` in order_execution.go (see flow above)
- New helpers: `hasAccessiblePlantable(char, targetType, lockedVariety)`, `findNearestPlantableItem(...)` in picking.go
- Wire into `findOrderIntent()`: `case "plant"` → `findPlantIntent()`
- Order completion: nil from `findPlantIntent` → complete if LockedVariety set, abandon if empty

**Phase 3: ActionPlant handler + rendering**

_Tests_ (in `internal/ui/update_test.go`):
- ActionPlant consumes plantable item from inventory, creates sprout at position
- ActionPlant from gourd seed: sprout has ItemType `"gourd"`, correct variety
- ActionPlant from berry: sprout has ItemType `"berry"`, correct variety
- ActionPlant sets LockedVariety on order when first planted
- Sprout has IsSprout=true, IsGrowing=true, symbol CharSprout

_Implementation:_
- ActionPlant handler in update.go: progress-based (ActionDurationMedium), on completion consume item + create sprout + set LockedVariety + log "Planted [variety]"
- Sprout rendering in `renderCell()`: mushroom=variety color, wet=green, else=olive (`growingStyle`)
- Tilled soil rendering: `═𖧧═` for sprout on tilled tile (existing fill pattern)
- Details panel: "Sprout of [variety]" for sprout items

**[TEST] Checkpoint — Plant flow (pre-serialization):**
- `go test ./...` passes
- Build and run (use `/test-world` with Plant know-how + tilled soil + seeds/berries):
  - Order Plant > Seeds: character picks up gourd seed, moves to tilled soil, plants
  - Sprout `𖧧` appears in olive on tilled soil (`═𖧧═`)
  - Sprout near water appears green
  - Mushroom sprout appears in variety color
  - Details panel shows "Sprout of warty green gourd" (or similar)
  - Character continues planting same variety until supply exhausted
  - Order completes when seeds exhausted or no empty tilled tiles
  - Order Plant > Berries: character picks up plantable berry, plants on tilled soil
  - Multiple characters can plant simultaneously with different orders
  - Note: sprouts do NOT mature yet (Step 2)

**Phase 4: Serialization**

_Tests:_
- IsSprout and SproutTimer round-trip through save/load

_Implementation:_
- IsSprout, SproutTimer in plant save fields

**[TEST] Checkpoint — Serialization:**
- Save, load preserves sprouts, locked variety, and sprout timers

**Reqs reconciliation:** Lines 84-99. _"Discoverable Know-How"_ ✓, _"Orderable"_ ✓, _"Garden > Plant > Select Plantable ItemType"_ ✓, _"planted item becomes a growing 'sprout'"_ ✓, _"sprouts appear olive unless tile is watered"_ ✓, _"mushrooms always the color of their variety"_ ✓, _"keep planting until that specific variety... is no longer available OR... no more unplanted tilled soil"_ ✓, _"More than one character can be planting at once"_ ✓.

**[DOCS]** Update README (Plant activity, sprouts), CLAUDE.md, game-mechanics, architecture.

**[RETRO]** Run /retro.

##### Step 2: Sprout Maturation + All Reproduction Through Sprouts ✅

**Design: Growth Multipliers**

Sprout timer and reproduction timer decremented faster on tilled/wet tiles. Applied per-tick as a multiplier on delta:
- `TilledGrowthMultiplier` (> 1.0, e.g., 1.25 = 25% faster on tilled soil)
- `WetGrowthMultiplier` (> 1.0, e.g., 1.25 = 25% faster on wet tile)
- Stacking: multiplicative (tilled × wet). Exact values tunable in Slice 9.

**Design: spawnItem Change**

`spawnItem()` in lifecycle.go creates sprouts instead of mature plants:
- `IsSprout: true`, `SproutTimer: SproutDuration`, symbol `CharSprout`
- `IsGrowing: true` but skipped for reproduction in `UpdateSpawnTimers`
- No SpawnTimer or DeathTimer yet — set on maturation

**Design: MatureSymbol**

Helper deriving mature symbol from ItemType: berry→`●`, mushroom→`♠`, flower→`✿`, gourd→`G`.

**Tests** (in `internal/system/lifecycle_test.go`):
- `UpdateSproutTimers` decrements SproutTimer by delta
- Sprout matures when SproutTimer reaches 0: IsSprout becomes false
- Mature sprout has correct symbol (berry→●, mushroom→♠, flower→✿, gourd→G)
- Mature plant has SpawnTimer set for reproduction
- Mature plant has DeathTimer set (for mortal types like flowers)
- Sprout on tilled soil: timer decremented faster by TilledGrowthMultiplier
- Sprout on wet tile: timer decremented faster by WetGrowthMultiplier
- `spawnItem` creates sprout (IsSprout=true, CharSprout symbol) instead of mature plant
- `spawnItem` sprout has correct parent variety and edible properties
- `UpdateSpawnTimers` skips sprouts for reproduction
- Reproduction timer also affected by tilled/wet multipliers

**Config:**
- `TilledGrowthMultiplier = 1.25` (25% faster — tunable in Slice 9)
- `WetGrowthMultiplier = 1.25` (25% faster — tunable in Slice 9)

**Implementation** (`internal/system/lifecycle.go`, `internal/config/config.go`, `internal/ui/update.go`):
- `MatureSymbol(itemType string) rune` helper
- `UpdateSproutTimers(gameMap *Map, initialItemCount int, delta float64)`:
  - Iterate items, find sprouts (`Plant.IsSprout == true`)
  - Calculate effective delta: `delta × TilledGrowthMultiplier (if tilled) × WetGrowthMultiplier (if wet)`
  - Decrement SproutTimer by effective delta
  - On maturation (timer ≤ 0): `IsSprout = false`, set Sym to `MatureSymbol(itemType)`, set SpawnTimer, set DeathTimer
- Modify `spawnItem()`: create sprout instead of mature plant (CharSprout symbol, IsSprout, SproutTimer)
- Modify `UpdateSpawnTimers`: skip sprouts (`if item.Plant.IsSprout { continue }`)
- Apply growth multipliers to reproduction SpawnTimer delta in `UpdateSpawnTimers`
- Wire `UpdateSproutTimers()` into game loop (after `UpdateSpawnTimers`, before `UpdateDeathTimers`)

**[TEST] Checkpoint — Full sprout lifecycle:**
- `go test ./...` passes
- Build and run:
  - Plant a seed on tilled soil → sprout appears → after ~30s matures into full plant with correct symbol
  - Mature plant reproduces normally → offspring appear as sprouts → they mature too
  - Wild plants (not on tilled soil) also reproduce through sprout phase
  - Sprout on tilled soil near water matures noticeably faster than wild sprout
  - Growth hierarchy: tilled+wet > tilled or wet alone > wild
  - Full food loop: gourd eaten → seed → plant → sprout → mature gourd → eat → more seeds
  - Berry picked → plantable → plant on tilled soil → sprout → mature berry
  - Existing gameplay unbroken: food supply maintained despite sprout phase
  - Save, load preserves sprout timers and maturation state

**Reqs reconciliation:** Lines 125-131. _"sprout eventually becomes a full grown version"_ ✓, _"extend normal plant reproduction to have sprout phase"_ ✓, _"tilled ground makes it grow faster"_ ✓, _"wet tile makes it grow faster"_ ✓, _"tiles adjacent to water sources are always 'wet'"_ ✓ (from Step 0).

##### Step 2b: Order Completion Refactor (bug fix from Step 2 testing) ✅

**Bugs found during Step 2 human testing:**
1. Completed plant orders stay on the order list after "Order Completed" message
2. Characters log "Order Completed" for planting without actually planting anything

**Root cause:** `CompleteOrder()` resets order to `status=open, assignedTo=0` but never removes the order from the list. For harvest/craft, removal happens in their action handlers in update.go via `m.removeOrder()`. But for till/plant, completion is detected in `selectOrderActivity` (system layer), which can't call the UI-layer `removeOrder`. The order sits there as `open` forever, gets re-assigned, and immediately "completes" again because `LockedVariety` is already set from the previous run.

**Design: Unified order completion pattern**

Triggered enhancement "Order completion criteria refactor" — implementing now. The principle: each order type checks its own completion criteria, but once those are met, a single consistent completion path handles logging, unassignment, and removal.

- Add `OrderCompleted` status to Order
- `CompleteOrder()` sets `OrderCompleted` (instead of resetting to `open`)
- Remove all `m.removeOrder()` calls that follow `CompleteOrder` in action handlers — harvest, craft, till, plant all just call `CompleteOrder`
- Add one sweep in the game loop: after applying intents, remove all `OrderCompleted` orders
- For till/plant action handlers: add inline completion checks (same pattern as harvest/craft) so completion is immediate, not deferred to next tick's `selectOrderActivity`
- `isMultiStepOrderComplete` in `selectOrderActivity` remains as safety net for edge cases (another worker finished the last tile while this character was walking)

**Tests** (in `internal/system/order_execution_test.go`):
- `CompleteOrder` sets order status to `OrderCompleted`
- `selectOrderActivity` skips orders with `OrderCompleted` status
- `findAvailableOrder` skips orders with `OrderCompleted` status

**Tests** (in `internal/ui/update_test.go`):
- ActionTillSoil completes and removes order when last marked tile tilled
- ActionPlant completes and removes order when no more plantable items or tilled tiles
- Completed harvest order removed from order list (existing test, verify no regression)
- Completed craft order removed from order list (existing test, verify no regression)

**Implementation:**
- `entity/order.go`: add `OrderCompleted` status, update `StatusDisplay()`
- `system/order_execution.go`: `CompleteOrder` sets `OrderCompleted`; `selectOrderActivity` and `findAvailableOrder` skip `OrderCompleted` orders
- `ui/update.go`: remove `m.removeOrder()` calls paired with `CompleteOrder`; add sweep after intent application; add completion checks in ActionTillSoil and ActionPlant handlers
- `save/state.go`: handle `OrderCompleted` in serialization (or skip — completed orders shouldn't persist)

**[TEST] Checkpoint — Order completion:**
- `go test ./...` passes
- Build and run:
  - Plant order: character plants items, order completes and disappears from list
  - Till order: character tills all tiles, order completes and disappears from list
  - Harvest/craft orders: no regression, still complete and disappear
  - No phantom "Order Completed" messages for characters that didn't do work
  - Save/load during active orders works correctly

##### Step 2c: Plant order procurement — ground vessel awareness (bug fix from Step 2b testing) ✅

**Bug:** Characters abandon plant orders even when a vessel containing matching plantable items sits on the ground. Each character takes the order, can't find items, and abandons.

**Root cause:** Two functions only look at loose ground items, not inside ground vessels:
- `findNearestPlantableOnGround` — searches `item.Plantable` on direct ground items only
- `PlantableItemExists` — checks ground items + character inventory vessels, but not ground vessels

The procurement flow in `EnsureHasPlantable` goes: check inventory → search ground → nothing found → return nil → abandon. Ground vessels are invisible.

**Design: New Layer 1 search utility + adjusted procurement order**

Per architecture.md's picking.go layering (Map Search → Prerequisite Orchestration → Physical Actions):

- New Layer 1 utility: `FindVesselContaining(cx, cy, items, targetType, lockedVariety)` — finds nearest ground vessel whose contents match the plant target. Sibling of `FindAvailableVessel` (which finds vessels that can *receive* items). Uses existing `matchesPlantTargetVariety` for content matching.
- `EnsureHasPlantable` (Layer 2) adjusted search order:
  1. `hasAccessiblePlantable` — already carrying? done
  2. Make inventory space if needed
  3. `FindVesselContaining` — ground vessel with matching contents? pick it up
  4. `findNearestPlantableOnGround` — loose item? pick it up
  5. Nothing → return nil (abandonment)
- `PlantableItemExists` (feasibility) — add ground vessel contents check to the ground-items loop, mirroring the existing character-vessel check.

Deferred: empty vessel + loose item orchestration (pick up empty vessel, fill it, then plant). Adds multi-step complexity for a minor optimization.

**Tests** (in `internal/system/picking_test.go` or `order_execution_test.go`):
- `FindVesselContaining` returns vessel with matching contents
- `FindVesselContaining` returns nil when vessel has wrong type
- `FindVesselContaining` returns nil when no vessels on ground
- `FindVesselContaining` respects locked variety
- `EnsureHasPlantable` returns pickup intent for ground vessel containing target
- `EnsureHasPlantable` prefers ground vessel over loose item (vessel checked first)
- `PlantableItemExists` returns true when matching items in ground vessel

**Implementation:**
- `system/picking.go`: add `FindVesselContaining`; reorder `EnsureHasPlantable` search (vessel before loose)
- `system/order_execution.go`: add ground vessel check to `PlantableItemExists`

**[TEST] Checkpoint — Plant procurement with ground vessels:**
- `go test ./...` passes
- Build and run:
  - Place vessel of berries on ground, issue plant berries order
  - Character picks up vessel, then plants berries from it
  - No order abandonment spam

**[DOCS]** Update README, CLAUDE.md, game-mechanics, architecture.

**[RETRO]** Run /retro.

---

### Slice 7: Fetch Water and Water Vessel Drinking

**Reqs covered:**

- New Idle Activity: Fetch Water (lines 101-108)
- Drinking from carried/dropped water vessels (lines 105-108)

**Resolved feature questions:**

- Fetch Water gets its own 1/5 idle slot (look/talk/forage/fetch-water/idle) ✓
- Earlier thirst trigger for carried vessel: deferred to Slice 9 tuning ✓
- Water vessel display: show contents in existing inventory/details panels (e.g., "Water 3/4"), no vessel name change ✓
- Water vessels can be dropped and re-picked-up normally (existing vessel drop/pickup works) ✓
- Preference formation for beverages: defer until beverage variety exists (tracked in triggered-enhancements.md) ✓
- Ground water vessel drinking: character drinks in place without picking up the vessel ✓

#### Design: Liquids as Vessel Stacks

Water stored in ContainerData.Contents as a Stack. The liquid type system uses the Kind pattern:
- **ItemType**: `"liquid"` — broad category for all liquid/beverage types
- **Kind**: `"water"` — specific liquid type (future: `"mead"`, `"beer"`, `"wine"`)
- **Stack size**: 4 (per reqs line 108: "4 drinks/units of water")

A vessel holds items OR liquid, never both. Existing vessel infrastructure (CanVesselAccept, IsVesselFull, variety-lock) applies naturally — a vessel with water can't accept berries, and vice versa.

Water is not an item that exists on the ground — it is created by interacting with water terrain and exists only inside vessels. This keeps water terrain (map state) separate from water as a usable liquid.

A water `ItemVariety` is registered at world generation: ItemType `"liquid"`, Kind `"water"`, no Color/Pattern/Texture variation (single variety). This allows Stacks in vessels to reference the variety. Config adds `"liquid": 4` to `StackSize`.

**Future extensions (tracked in triggered-enhancements.md):**
- `Drinkable bool` on ItemVariety — when non-drinkable liquids appear (lamp oil, dye)
- `Watertight bool` on ContainerData — when non-watertight containers appear (baskets)

#### Design: Liquid Vessel Helpers

Two new helpers in picking.go (Layer 1 — physical actions):

- `AddLiquidToVessel(vessel *Item, variety *ItemVariety, amount int) bool` — creates/fills a liquid Stack without a source item. Returns false if vessel is non-empty with different variety or already full. Uses existing Stack infrastructure.
- `DrinkFromVessel(vessel *Item) bool` — decrements liquid stack count by 1. Removes stack when count reaches 0 (vessel becomes empty). Returns false if no liquid contents.

These parallel existing vessel helpers (`AddToVessel`, `ConsumeAccessibleItem`) but handle the "liquid from terrain" and "consume one unit" cases that don't map to item-based operations.

#### Design: ActionFillVessel

New ActionType for filling a vessel at water terrain. Distinct from ActionPickup because:
- **ActionPickup**: picks up a ground item → item goes into vessel via `AddToVessel()`
- **ActionFillVessel**: interacts with water terrain → creates liquid Stack in vessel from nothing

Duration: `ActionDurationShort` (quick interaction with water source).
Outcome: vessel filled to capacity with water Stack (4 units).

#### Design: Fetch Water Idle Activity

New 1/5 idle slot in `selectIdleActivity()` (expanding from 4 to 5 options: look/talk/forage/fetch-water/idle).

`findFetchWaterIntent()` flow:
1. Check if character has empty vessel (carried, with no contents) → skip to step 3
2. No empty vessel carried → search ground for empty vessel → ActionPickup intent to pick it up
3. Has empty vessel → find nearest water terrain via `FindNearestWater()`
4. If not adjacent to water → ActionMove toward water
5. If adjacent to water → ActionFillVessel intent
6. If no empty vessel available and none on ground → return nil (skip this idle option)

Non-destructive: never dumps existing vessel contents. Only triggers when character has or can find an empty vessel.

Vessel-seeking reuses existing patterns from foraging (find vessel on ground, pick it up). The "go to water" phase reuses `FindNearestWater()` and cardinal adjacency from drinking.

#### Design: Unified Drink Source Search

Refactor `findDrinkIntent()` to search all water sources by distance, mirroring hunger/foraging's unified distance-based approach:

| Source | Distance | How to drink |
|--------|----------|-------------|
| Carried water vessel | 0 (always closest if available) | Consume 1 unit from carried vessel |
| Ground water vessel | tiles to vessel position | Move to vessel, consume 1 unit in place |
| Water terrain | tiles to nearest cardinal-adjacent spot | Move adjacent, drink from terrain (existing) |

Character picks the closest source. In the future when other beverages exist, preference scoring will enter into beverage selection (paralleling hunger's preference + distance scoring).

**Intent tracking:** Intent carries source info to distinguish vessel vs terrain drinking:
- `TargetItemID` set → drinking from vessel (carried or ground). Character looks up vessel by ID at execution time.
- `TargetWaterPos` set, no `TargetItemID` → drinking from water terrain (existing path).

**Drink continuation:**
- **From terrain**: existing continuation logic — keep drinking until thirst = 0 (infinite source, already there).
- **From vessel**: drink once (consume 1 unit), then clear intent to force re-evaluation. If still thirsty above threshold, next `CalculateIntent` naturally picks the closest source again (might be same vessel if still has water, might be terrain if vessel emptied).

**Ground vessel drinking:** character moves to the vessel and drinks from it in place — doesn't pick it up. Like a small communal water source. Multiple characters can drink from the same dropped vessel.

**`canFulfillThirst()` update:** Check all three source types (carried water vessel, ground water vessel, water terrain). If any source exists, thirst is fulfillable. This ensures the urgency/fallback system doesn't skip thirst when a water vessel is available.

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → add minimal stubs to compile → verify red → implement → verify green → human testing checkpoint.

##### Step 1: Liquid Type + Water Variety + Vessel Helpers

**Tests** (in `internal/entity/variety_test.go` or `internal/game/variety_generation_test.go`):
- Water variety registered with ItemType `"liquid"`, Kind `"water"`
- `GetStackSize("liquid")` returns 4

**Tests** (in `internal/system/picking_test.go`):
- `AddLiquidToVessel` fills empty vessel with liquid Stack (correct variety, count)
- `AddLiquidToVessel` returns false when vessel has non-liquid contents (variety mismatch)
- `AddLiquidToVessel` returns false when vessel already full of liquid
- `AddLiquidToVessel` tops up partially-filled liquid vessel
- `DrinkFromVessel` decrements liquid stack count by 1
- `DrinkFromVessel` removes stack when count reaches 0 (vessel empty)
- `DrinkFromVessel` returns false when vessel has no liquid contents
- `DrinkFromVessel` returns false when vessel is empty

**Config:** `"liquid": 4` in `StackSize`

**Implementation:**
- Register water `ItemVariety` in variety generation: ItemType `"liquid"`, Kind `"water"`, Sym `0` (never rendered as ground item), no Edible/Plantable
- Follows variety registration pattern in `game/variety_generation.go`
- `AddLiquidToVessel(vessel *Item, variety *ItemVariety, amount int) bool` in picking.go
- `DrinkFromVessel(vessel *Item) bool` in picking.go
- Add `ActionFillVessel` to ActionType enum in entity/character.go (needed by Step 2)

**Serialization:** Liquid stacks serialize as vessel contents (existing `StackSave` with variety ID + count). Verify existing save/load handles liquid variety — likely works with no changes since variety is registered. Add round-trip test.

**Tests** (in `internal/ui/serialize_test.go`):
- Water vessel round-trips through save/load: vessel with liquid Stack preserved, variety and count correct

**Architecture patterns:** Extends vessel Stack system. `AddLiquidToVessel` is a sibling of `AddToVessel`; `DrinkFromVessel` is a sibling of `ConsumeAccessibleItem`. Both follow Layer 1 (Physical Actions) in picking.go.

##### Step 2: ActionFillVessel + Fetch Water Idle Activity

**Tests** (in `internal/system/idle_test.go` or new `internal/system/fetch_water_test.go`):
- `findFetchWaterIntent` returns ActionPickup intent for empty ground vessel when character has no vessel
- `findFetchWaterIntent` returns ActionMove toward water when character has empty vessel but not adjacent
- `findFetchWaterIntent` returns ActionFillVessel when character has empty vessel and is adjacent to water
- `findFetchWaterIntent` returns nil when no empty vessel available (character carries full vessel)
- `findFetchWaterIntent` returns nil when character has empty vessel but no water on map
- `findFetchWaterIntent` skips vessels that have contents (non-destructive)

**Tests** (in `internal/ui/update_test.go`):
- ActionFillVessel completes after ActionDurationShort: vessel now contains water Stack (4 units)
- ActionFillVessel does not overfill partially-filled vessel

**Implementation:**
- `findFetchWaterIntent()` in new file `internal/system/fetch_water.go` (follows pattern of foraging.go as separate idle activity file)
- Wire into `selectIdleActivity()`: expand random roll from 0-3 to 0-4, add case 3 for fetch water (shift existing idle/nil to case 4)
- ActionFillVessel handler in update.go: progress-based (ActionDurationShort), on completion call `AddLiquidToVessel` with water variety, log "Filled [vessel name] with water"
- Update `isIdleActivity()` to include fetch water activity names

**Architecture patterns:** Follows foraging idle activity pattern (separate file, `findXIntent()` function, wired into `selectIdleActivity`). ActionFillVessel follows existing progress-based action pattern in update.go.

**[TEST] Checkpoint — Fetch Water idle activity:** ✅
- `go test ./...` passes ✅
- Build and run:
  - Characters occasionally choose fetch water as idle activity ✅
  - Character with empty vessel walks to nearest water, fills it ✅
  - Character without vessel picks up empty vessel from ground, then fills it — not yet observed (foraging rolled first in test world). **Retest after resolving foraging-for-vessel issue (randomideas.md 1a).**
  - Vessel shows water contents in inventory panel ✅
  - Character with full vessel (food or water) does not attempt fetch water ✅
  - Save/load preserves water vessel contents ✅
  - Note: drinking from water vessel does NOT work yet (Step 3)
- Bugs fixed during testing: liquid content check (was checking vessel fullness, now checks for liquid ItemType), collision handling added to ActionFillVessel movement, unified two-phase ActionFillVessel flow (eliminated two-roll problem).

**[DOCS]** Update README, CLAUDE.md, game-mechanics, architecture.

**[RETRO]** Run /retro.

##### Step 3: Unified Drink Source Search + Vessel Drinking

**Design: Three drink sources scored by distance**

Refactor `findDrinkIntent()` to search all water sources by distance (mirrors `findFoodIntent`'s unified scoring approach):

| Source | Distance | Intent shape | How to drink |
|--------|----------|-------------|-------------|
| Carried water vessel | 0 (always closest if available) | `ActionDrink`, `Target=Dest=charPos`, `TargetItem=vessel` | Consume 1 unit from carried vessel |
| Ground water vessel | Manhattan distance to vessel position | `ActionMove` toward vessel, then `ActionDrink` at vessel tile, `TargetItem=vessel` | Move to vessel, consume 1 unit in place (don't pick up) |
| Water terrain | Manhattan distance to nearest cardinal-adjacent tile | `ActionMove` toward adjacent tile, then `ActionDrink`, `TargetWaterPos=waterTile` | Move adjacent, drink from terrain (existing) |

Character picks the closest source. Tie-breaking: carried vessel > ground vessel > terrain (by construction — carried is distance 0).

**Intent field usage:** `TargetItem` set → vessel drinking. `TargetWaterPos` set, `TargetItem` nil → terrain drinking. No new fields needed — follows existing pointer-based pattern.

**Design: continueIntent — follows ActionConsume pattern for carried vessels**

Carried vessel drinking mirrors the eating-from-inventory pattern (`ActionConsume`):
- `findDrinkIntent` returns `ActionDrink` with `Target=Dest=charPos` when vessel is in inventory
- `continueIntent` early-returns for `ActionDrink` when `TargetItem` is in character inventory (same as `ActionConsume` at line 288), skipping the map-existence check
- Ground vessel drinking uses the existing `TargetItem`-on-map continuation path (character walks to vessel, arrives, drinks) — no special handling needed

**Design: Drink continuation**

- **Terrain**: existing behavior — intent persists, character keeps drinking until `CalculateIntent` says thirst is satisfied (infinite source)
- **Vessel**: after each drink, clear intent (`char.Intent = nil`) to force re-evaluation. If still thirsty, next `CalculateIntent` naturally picks closest source (same vessel if still has water, or terrain if vessel emptied). This matches the plan's "re-evaluate after each drink" design from the Decisions Log.

**Design: Source-specific log messages**

Upgrade existing generic "Drinking from water" to source-specific text using `gameMap.WaterAt()` for terrain:

| Source | Log message |
|--------|-------------|
| Spring terrain | "Drinking from spring" |
| Pond terrain | "Drinking from pond" |
| Carried vessel | "Drinking from vessel" |
| Ground vessel | "Drinking from vessel" |

**Design: Adjacent-item-look bypass**

`ActionDrink` bypasses the adjacent-item-look conversion in `movement.go` (same as `ActionForage`). A thirsty character walks directly onto a ground vessel's tile to drink, rather than stopping adjacent to look at it.

---

**Phase 1: Unified findDrinkIntent + canFulfillThirst** ✅

_Implementation notes (for Phase 2 context):_
- `findDrinkIntent` signature is now `(char, pos, gameMap, tier, log, items []*entity.Item)` — last param added
- `canFulfillThirst` signature is now `(char, gameMap, pos, items)` — all 4 callers in re-eval block updated
- New helpers in `movement.go`: `vesselHasLiquid(item)` checks if vessel contains liquid; `waterSourceName(gameMap, waterPos)` returns "spring"/"pond"/"water"
- `continueIntent` arrival path (water target) also updated with source-specific log messages via `waterSourceName`
- Adjacent-item-look bypass for `ActionDrink` is NOT needed — the bypass condition already checks `DrivingStat == ""` which is false for thirst-driven intents. No code change required.

**Phase 2: ActionDrink handler branching + continueIntent** ✅

_Tests_ (in `internal/ui/update_test.go`):
- `ActionDrink` from carried vessel: calls `DrinkFromVessel`, reduces thirst, clears intent
- `ActionDrink` from carried vessel: intent cleared after drink (forces re-eval)
- `ActionDrink` from ground vessel: calls `DrinkFromVessel` on ground vessel item, reduces thirst, clears intent
- `ActionDrink` from vessel that empties (last unit): intent cleared, vessel stack removed
- `ActionDrink` from terrain: existing behavior unchanged (continues until sated, intent NOT cleared)

_Implementation:_
- `ActionDrink` handler in `update.go` (currently at ~line 814): branch on `char.Intent.TargetItem != nil`:
  - **Vessel path**: duration-based (ActionDurationShort). On completion: find vessel (check inventory first via `char.FindInInventory`, then `m.gameMap.ItemAt` for ground vessel). Call `system.DrinkFromVessel(vessel)`. Call `system.Drink(char, m.actionLog)` for thirst reduction. Clear intent (`char.Intent = nil`).
  - **Terrain path**: existing behavior — duration-based, calls `Drink()`, does NOT clear intent.
- `continueIntent` in `movement.go` (~line 288): add early return for `ActionDrink` when `TargetItem` is in character inventory (mirrors `ActionConsume` pattern). Ground vessel drinking uses existing `TargetItem`-on-map continuation path unchanged.
- Ground vessel arrival: existing `continueIntent` handles `TargetItem` on map — character walks to vessel position. Need to add arrival detection: when character is at `TargetItem` position and action is thirst-driven, convert to `ActionDrink` (similar to water terrain arrival at ~line 407).

Architecture patterns: ActionConsume pattern for carried-vessel intent continuation. Self-managing continuation for ground vessel via existing `TargetItem` map-check path.

**[TEST] Checkpoint — Vessel drinking:**
- `go test ./...` passes
- Build and run:
  - Thirsty character with carried water vessel drinks from it immediately (no walking to water)
  - Each drink reduces vessel water by 1 unit
  - After vessel empties, character seeks nearest water source (terrain or another vessel)
  - Drop a water vessel → another character drinks from it in place
  - Character with both water vessel and nearby pond: drinks from vessel (distance 0)
  - Existing spring/pond drinking unchanged — log now says "Drinking from spring" or "Drinking from pond"
  - Full loop: character fetches water → drinks from vessel when thirsty → refills when empty
  - Save/load during drinking works correctly

**[DOCS]** Update README, CLAUDE.md, game-mechanics, architecture.

**[RETRO]** Run /retro.

##### Step 4: UI Display Polish ✅

**Scope:** Existing vessel content display already works for liquids (`water: 3/4` format, same as berries/mushrooms). The only polish needed is rendering liquid stack lines in `waterStyle` (bright blue bold) — consistent with how water terrain and "Wet" annotations already render.

**Implementation** (no tests per CLAUDE.md for UI rendering):
- Details panel (view.go ~line 910): when stack's variety has ItemType `"liquid"`, render the content line with `waterStyle`
- Inventory panel (view.go ~line 1273): same treatment — liquid stack lines in `waterStyle`
- Two rendering paths, same change: detect liquid ItemType on the stack variety, apply waterStyle to the formatted string

**[TEST] Checkpoint — UI polish:**
- `go test ./...` passes (no regressions)
- Build and run:
  - Water vessel contents display in blue bold text in both details panel and inventory panel
  - Non-liquid vessel contents (berries, mushrooms) render unchanged
  - Empty vessels still show "(empty)" in default style

**[DOCS]** Update README, CLAUDE.md, game-mechanics, architecture.

**[RETRO]** Run /retro.

**Reqs reconciliation:** Lines 101-108. _"fill an empty vessel with water as an idle activity option"_ ✓ (Step 2), _"same vessel logic as foraging"_ ✓ (vessel-seeking reused), _"drink from the carried vessel instead of having to move to a water source"_ ✓ (Step 3, distance=0 prioritization), _"a dropped water vessel can be targeted for drinking"_ ✓ (Step 3, ground vessel search), _"a full vessel has 4 'drinks'/units of water in it"_ ✓ (Step 1, StackSize 4). Line 106 _"consider that drinking from carried vessel can be triggered earlier"_ — deferred to Slice 9 tuning.

#### Post-Slice 7 Bug Fixes (from randomideas.md issues 1a, 1c)

Three fixes, tested and committed independently:

**Fix 1A: "Foraging for vessel" text (text-only).** In `createPickupIntent` (foraging.go), when `itemType == "vessel"`, use different wording: "Picking up vessel" instead of "Foraging for vessel." Activity text and action log both affected. Surgical change to the text branch, no scoring logic changes.

**Fix 1B: Foraging vessel pickup should continue to food (not go idle).** Original framing was "tighten empty vessel scoring," but investigation revealed the real problem is deeper — see design discussion below.

**Fix 2: Edibility bug — sprout maturation restores variety attributes.**

Root cause: gourd seeds have `Edible: nil` (seeds aren't food). `CreateSprout` receives `plantedItem.Edible` which is nil for seeds. When sprout matures in `UpdateSproutTimers`, `MatureSymbol()` restores the symbol but nothing restores `Edible`. Matured gourds from seeds are inedible.

Fix site: `UpdateSproutTimers` in `lifecycle.go`, at the maturation block (line ~51). No signature change needed — `gameMap.Varieties()` already provides the registry. At maturation, look up the variety via `GetByAttributes(item.ItemType, item.Color, item.Pattern, item.Texture)` and restore `Edible` from it. Copy the struct (don't share pointer). Forward-compatible for any future variety attributes.

Architecture pattern: Registry as source of truth (variety_registry.go `GetByAttributes`).

Implementation steps:

1. **Test** (`lifecycle_test.go`): Create a sprout with `Edible: nil` and matching variety in registry with `Edible` set. Run `UpdateSproutTimers` past maturation. Assert `item.Edible` is non-nil with correct Poisonous/Healing values.
2. **Fix** (`lifecycle.go`): In the `SproutTimer <= 0` block, after restoring symbol, look up variety and restore `Edible`.
3. **[TEST] Human testing**: Use `/test-world` to create a world with a gourd seed planted as a sprout (Edible: nil). Fast-forward past maturation. Select the matured gourd and verify edible properties display. Also verify a character will eat it.

**Bonus fix (discovered during testing):** `itemFromSave` in `serialize.go` sets symbol by ItemType but never checks `IsSprout`. All sprouts display with mature symbol after save/load. Fix: after the type-based switch, override symbol to `CharSprout` if `plant.IsSprout`. Regression test added to `TestSproutSerialization_RoundTrip`.

✅ Both fixes verified via human testing. Character ate the matured gourd (hunger 55→35), sprout symbol displayed correctly on load.

**Panic fix (discovered during Fix 1A testing):** `ActionFillVessel` phase 1→2 transition crashed with nil pointer dereference. `Pickup` clears `char.Intent`; the transition code then tried to mutate the nil intent. Fix: create a fresh `Intent` struct for phase 2. Regression test added: `TestApplyIntent_FillVessel_GroundVesselPickupTransitionsToPhase2`. ✅

**Circle-back notes from Fix 1A testing:**

1. **Vessel pickup messaging lacks intent context.** ✅ Resolved by Fix 1B Step 1. Log shows "Picking up vessel for water" — contextual messaging from self-managing action.
2. **Fetch water ground vessel pickup confirmed working** after panic fix. ✅
3. **Repeated "Heading to water to fill vessel" during movement.** ✅ Resolved by Fix 1B Step 1. Single "Heading to water" message, no re-acquisition spam.

##### Fix 1B: Self-Managing Actions + Shared Vessel Procurement

**Root cause:** `Pickup()` clears intent and sets idle cooldown unconditionally for `PickupToInventory`. This conflates "the physical act of picking something up" with "the intent that motivated the pickup is complete." For food, those are the same. For vessel-as-setup, they're not. Foraging vessel pickup goes idle, requiring a second idle roll to continue to food. ActionFillVessel worked around this by calling `Pickup()` directly in its handler, but duplicated procurement logic.

**Decision: Option C — self-managing actions + shared vessel procurement helper.** (See Decisions Log.) Three options were evaluated:
- Option A (self-managing ActionForage, duplicate procurement): proven pattern but duplicates vessel procurement per action
- Option B (Pickup stops clearing intent, central ActionPickup routing): fixes root cause but creates central dispatcher that fights self-managing pattern, scales poorly
- Option C (self-managing actions + shared helper): self-managing pattern + shared `RunVesselProcurement` tick helper in picking.go. No duplication. Each action owns its lifecycle. See architecture.md "Self-Managing Actions" for full pattern documentation.

**Design:**

New shared helper in `picking.go`:
```
RunVesselProcurement(char, items, gameMap, log, actionLog, delta) → ready bool
```
Called each tick by self-managing action handlers. Returns `true` when vessel is in hand (proceed to main phase). Returns `false` if still working (moved toward vessel or picked it up this tick). Signals failure by setting `char.Intent = nil`. Handles: find ground vessel, move toward it, pick it up via direct `Pickup()` call (not ActionPickup), log contextually.

**ActionForage (new):** All foraging uses `ActionForage` — both vessel-then-food and simple food pickup. `findForageIntent` returns `ActionForage` instead of `ActionPickup`. Handler phases:
- Phase 1 (optional): If scoring chose a vessel target and vessel is on ground, `RunVesselProcurement` handles it
- Phase 2: Move to food target, pick up. Go idle after one food item (same completion as today's foraging)
- Messaging: "Foraging" throughout, with sub-log for vessel pickup ("Picking up vessel for foraging")

**ActionFillVessel (refactored):** Replace inline phase-1 procurement with `RunVesselProcurement`. Phase 2 (move to water, fill) unchanged. Shorter handler, same behavior.

**What doesn't change:**
- `Pickup()` itself — still clears intent for ActionPickup callers
- `ActionPickup` handler — still handles harvest continuation, order prerequisites
- `findFetchWaterIntent` — still returns `ActionFillVessel`, but ground-vessel detection simplifies since the handler + helper figure out which phase

**Resolves:**
- Fix 1A circle-back #1 (vessel messaging): self-managing actions know their context
- Fix 1A circle-back #3 (repeated log messages): intent persists through procurement, no re-acquisition
- Fix 1B (foraging chain break): vessel pickup is a phase within ActionForage, not a separate intent
- Foundation for Water Garden (Slice 8): same procurement-then-work shape via shared helper

#### Implementation Steps

**Step 1: RunVesselProcurement helper + ActionFillVessel refactor** ✅

Extract vessel procurement from ActionFillVessel's inline phase-1 code into `RunVesselProcurement` in picking.go. Refactor ActionFillVessel handler to call it. Behavior should be identical — this is a pure extraction refactor.

Architecture patterns: Self-Managing Actions, Shared Procurement Helper (picking.go)

Tests:
- Existing ActionFillVessel tests must pass unchanged (ground vessel pickup → fill, inventory vessel → fill)
- New unit test for `RunVesselProcurement`: vessel on ground → moves toward it → picks up → returns ready
- New unit test for `RunVesselProcurement`: no vessel available → signals failure (intent nil)

**[TEST] Checkpoint — ActionFillVessel refactor:** ✅
- Fetch water flow: pickup → heading to water → filled. One continuous action. ✅
- No repeated "Heading to water" spam. ✅
- Contextual messaging: "Picking up vessel for water". ✅

**Step 2: ActionForage — replace ActionPickup for foraging** ✅

Add `ActionForage` constant. `findForageIntent` returns `ActionForage` instead of `ActionPickup`. New `ActionForage` handler in update.go:
- Phase 1 (optional): if intent targets a ground vessel, call `RunVesselProcurement`. On ready, transition to food-seeking phase (find best food target using existing scoring, update intent).
- Phase 2: move to food target, pick up via direct `Pickup()` call. Go idle after one food item.

Remove foraging-specific logic from ActionPickup handler (vessel scoring path). ActionPickup retains: harvest order continuation, order prerequisite pickups.

Architecture patterns: Self-Managing Actions, foraging.go scoring

Tests:
- New test: ActionForage with vessel on ground → picks up vessel → continues to food → picks up food → goes idle
- New test: ActionForage direct food pickup (no vessel needed) → picks up food → goes idle
- New test: ActionForage vessel procurement fails (no vessel) → falls back to direct food pickup
- Existing harvest order tests must pass unchanged (still use ActionPickup)
- Existing craft/till/plant order prerequisite tests must pass unchanged

**[TEST] Checkpoint — Foraging refactor:** ✅
- Start a world. Watch foraging behavior: characters pick up food directly, or pick up vessel then food.
- Verify log messages: "Foraging" activity, "Picking up vessel for foraging" when applicable — not "Foraging for vessel."
- Verify harvest orders still work correctly (fill vessel via ActionPickup, not ActionForage).
- Verify fetch water still works correctly (ActionFillVessel with shared procurement).
- Verify no double-idle-roll: character picks up vessel and immediately continues to food without going idle in between.

**Implementation notes:**
- `ActionForage` bypasses the adjacent-item-look conversion in `movement.go` (characters walk onto food instead of stopping to look).
- `RunVesselProcurement` returns `ProcurementStatus` enum (ProcureReady/ProcureApproaching/ProcureInProgress/ProcureFailed).
- `ActionPickup` handler cleaned up — now only used by harvest orders and order prerequisites.

**[DOCS]** ✅

**[RETRO]**

#### Post-Slice 7 Fix: Discoverability Friction (from randomideas.md issue 1d) ✅

**Problem:** Planting is discovered quickly (plantable items are abundant), but tilling requires seeing a hoe, which requires crafting, which requires rare shells. Characters learn to plant before anyone knows how to till — feels backwards.

**Fix: Recipe-bundled activity discovery.** When a character discovers the shell-hoe recipe, also grant tillSoil know-how. The insight is: inventing a digging tool and knowing you can dig are one idea, not two. Implemented as a `BundledActivities` field on Recipe — data-driven, extensible to future recipes that imply activity knowledge.

**Design:** Add `BundledActivities []string` to `Recipe` struct. In `tryDiscoverRecipe()`, after granting the recipe and its parent activity, also grant each bundled activity. Shell-hoe recipe gets `BundledActivities: []string{"tillSoil"}`.

**Architecture patterns:** Extends recipe discovery system in `discovery.go`. Follows existing pattern where recipe discovery grants parent activity — bundled activities are additional grants in the same roll.

**Implementation:**

**Tests** (in `internal/system/discovery_test.go`):
- Discovering shell-hoe recipe also grants tillSoil activity
- Character who already knows tillSoil still discovers recipe without error
- Bundled activity appears in action log

**Implementation:**
- Add `BundledActivities []string` to `Recipe` struct in `entity/recipe.go`
- Add `BundledActivities: []string{"tillSoil"}` to shell-hoe recipe
- In `tryDiscoverRecipe()` in `discovery.go`: after granting recipe, loop through `BundledActivities` and call `char.LearnActivity()` for each. Log discovery for each newly learned activity.
- No serialization changes needed — `BundledActivities` is static registry data, not saved state

**[TEST] Checkpoint — Discoverability fix:**
- `go test ./...` passes
- Build and run with `/test-world`: character discovers shell-hoe recipe → also learns tillSoil. Verify in action log: both "Learned Shell Hoe recipe!" and "Discovered how to Till Soil!" appear. Verify character can now be assigned till orders.

**[DOCS]** Update README, CLAUDE.md, game-mechanics, architecture.

**[RETRO]** Run /retro.

---

### Slice 8: Water Garden and Wet Tiles

**Reqs covered:**

- New Activity: Water Garden (lines 110-123)
- Wet tile mechanics (line 131, plus manual watering decay)

#### Design: Lookable Water Terrain — DESCOPED

Deferred to triggered-enhancements.md ("Terrain-aware Look + discovery"). Narratively, characters should be able to observe terrain, not just items. But Water Garden discovery is adequately served by item-based and action-completion triggers (see below). Lookable terrain is flavor, not blocking.

#### Design: Water Garden Activity

Orderable, discoverable. No recipe needed. Prerequisite: vessel with water, procured via self-managing action pattern (vessel procurement → fill at water source → water tiles).

**Discovery (Approach B — action-completion triggers):** Fire `CheckDiscovery` from `ActionFillVessel` completion handler with the vessel as the item. Water Garden triggers: `{ActionFillVessel, ItemType: "vessel"}` + `{ActionLook, RequiresPlantable: true}` (looking at sprouts). The insight: "I filled a vessel and see these sprouts — I could water them." This extends the existing trigger system minimally — just add a `CheckDiscovery` call at action completion. Extensible to any future activity discovered from action completion (e.g., brewing, construction).

**Watering action**: Character waters the closest dry tilled planted tile. Uses 1 unit of water per tile. If water remains, continues to next dry tilled planted tile. If vessel empty but more dry tiles exist, character refills at nearest water source and continues. Completion: no remaining dry tilled planted tiles. Multi-assignment supported.

**Refill/resume after interruption:** Stateless re-evaluation. `findWaterGardenIntent` re-derives the correct action each call: "do I have water? → water nearest dry tile. No water? → seek water source. No dry tiles? → complete." If the character gets hungry mid-watering, survives, comes back to the order, the intent function just re-derives. No explicit "was refilling" state needed. Same pattern as tilling.

**Logging note (from Till Soil retro):** For bulk repeated actions, only log on completion — not on start. Log "Watered [variety]" per tile completion.

#### Design: Wet Tile System

Two sources of wetness:

1. **Water-adjacent (8-directional)**: Always wet. Computed on the fly via existing `IsWet(pos)` — no state to track or save.
2. **Character-watered**: Tracked per-tile with a decay timer. Wears off after 3 world days (360 game seconds). Stored as `map[Position]float64` for remaining wet time. Included in save/load.

**Visual:** Wet tilled tiles render `═══` fill in green instead of olive (same `═` character, different style). Applies to both empty wet tilled tiles and `═X═` fill around entities on wet tilled soil. Consistent with existing sprout color logic (olive dry, green wet).

**Outcomes:**

1. Water Garden: orderable, discoverable from filling vessels + looking at sprouts
2. Watering uses water vessel, 1 unit per tile, auto-continues to next dry tile
3. Character refills vessel and continues if more tiles need watering
4. Completion when no dry tilled planted tiles remain
5. Multi-assignment: multiple characters can water simultaneously
6. Wet tiles: water-adjacent always wet + manual watering with 3-world-day decay
7. Visual: wet tilled tiles green, dry tilled tiles olive

**[TEST] Checkpoint — Water Garden:**
- Character discovers Water Garden from filling a vessel or looking at sprouts
- Order Water Garden. Character fills vessel (if needed), waters planted tiles.
- Watered tiles turn green. Dry tilled tiles remain olive.
- Character continues watering until all planted tiles wet or vessel empty
- If vessel empty and more tiles need watering, character refills at nearest water source
- Plants on watered tiles grow faster than dry tilled plants
- Watering effect wears off after ~3 world days (use time skip to verify)
- Tiles adjacent to ponds are always wet (no watering needed)
- Wet tilled-but-unplanted tiles also display green
- Multi-assignment: two characters watering same garden
- Save and load preserves watered tile timers

**[TEST] Final Checkpoint — Full Part II Integration:**

Start a new world and play through the full garden lifecycle:
- Till soil → plant seeds → water garden → sprouts grow → full plants → natural reproduction (also through sprout phase)
- Growth hierarchy visible: watered tilled > dry tilled > water-adjacent wild > plain wild
- Food chain: gourds eaten → seeds → planted → grow → more gourds (sustainable food loop)
- Flower cycle: forage flower → seed → plant → grow → more flowers → more seeds
- Satiation differences noticeable: gourd satisfies much more than berry
- Water management: characters fetch water, drink from vessels, water gardens

**[DOCS]** Final doc pass for Part II.

**[RETRO]** Run /retro.

**Feature questions (resolved):**

- ~~How does "refill and continue" interact with order pause/resume?~~ Stateless re-evaluation, no explicit state needed.
- ~~Does looking at water require standing adjacent or at distance?~~ Moot — lookable water terrain descoped from Slice 8.
- ~~Water Garden discovery triggers?~~ Approach B: ActionFillVessel completion + ActionLook on sprouts. Extends existing system minimally.
- ~~Visual interaction between wet and tilled?~~ Green `═══` instead of olive for wet tilled tiles.

#### Design Alignment

**Values cross-check:**
- *Consistency*: watered tile state follows tilled soil map pattern; self-managing action follows ActionFillVessel/ActionForage; shared fill helper follows RunVesselProcurement shape
- *Source of Truth*: `IsWet()` is the single query for "is this tile wet from any source." Growth code calls only `IsWet()`, never individual sources.
- *Reuse*: RunVesselProcurement for vessel acquisition, extracted fill helper from existing ActionFillVessel Phase 2
- *Future Siblings*: `wateredTimers map[Position]float64` is the first timed tile effect. If a second appears (fertilizer, contamination), generalize into a `TileEffect` system. For now, named map + named methods is clear enough.
- *Requirements as Ground Truth*: "dry tilled planted tile" = tilled AND has growing plant AND `!IsWet(pos)`. Tests should validate this semantic condition for completion, not structural checks.

**Architecture cross-check:**
- Self-managing action pattern (architecture.md "Self-Managing Actions"): handler owns full lifecycle, stateless phase detection, shared procurement helpers
- Component procurement pattern: vessel + water procurement via existing helpers
- Unified order completion pattern: OrderCompleted status, swept by game loop

#### Design: Self-Managing ActionWaterGarden

Water Garden is a self-managing action (like ActionFillVessel, ActionForage). The handler owns its full lifecycle across phases. Phase detection is stateless — check world state each tick to determine current phase. Survives save/load without additional serialization.

**Phases:**
1. **Vessel procurement** (if no vessel in inventory): `RunVesselProcurement` shared helper. Same pattern as ActionFillVessel Phase 1.
2. **Fill vessel** (if vessel empty): Move to nearest water source, fill vessel. Same logic as ActionFillVessel Phase 2 — extract shared helper to avoid duplication.
3. **Water tile** (if vessel has water): Find nearest dry tilled planted tile, move to it, water it (consume 1 unit). Loop: if more water and more dry tiles → next tile. If no water → Phase 2. If no dry tiles → complete.

**Phase detection (stateless):**
- No vessel in inventory? → Phase 1
- Vessel in inventory, no water? → Phase 2
- Vessel in inventory, has water? → Phase 3

**findWaterGardenIntent:** Returns `ActionWaterGarden` always. The handler detects and manages all phases internally. This keeps the intent function simple (just "do water garden work") and the handler self-contained.

**Prerequisite per reqs line 115:** "at least one vessel full with water." On first assignment, `findWaterGardenIntent` checks if a vessel with water exists (carried, ground, or fillable). If no vessel exists at all → return nil (abandon, unfeasible). Otherwise return ActionWaterGarden intent and let the handler figure out procurement/fill/water phases.

**Shared fill helper:** ActionFillVessel Phase 2 (move to water + fill) should be extracted into a shared helper (like `RunVesselProcurement` for pickup). Both ActionFillVessel and ActionWaterGarden call it. Returns status enum (Approaching/InProgress/Ready/Failed). Avoids duplicating the "move to water and fill vessel" logic.

**Order completion (reqs line 122):** "once there are no remaining dry tilled planted tiles, order is complete." A tile is "dry tilled planted" when it is tilled AND has a growing plant AND `!IsWet(pos)` (not wet from any source — water adjacency or manual watering). When no such tiles exist, `CompleteOrder`. Follows unified order completion pattern (OrderCompleted status, swept by game loop). Two characters may target the same tile on the same tick — harmless, one waters it, the other re-evaluates next tick and picks the next dry tile.

**Feasibility:** `IsOrderFeasible` for waterGarden checks: (1) vessel exists in world (ground or inventory), (2) water exists on map, (3) at least one dry tilled planted tile exists (tilled + has growing plant + `!IsWet`). All three must be true. If all planted tiles are already wet from water adjacency, the order is unfulfillable — there's nothing to water.

---

#### Implementation Steps

Each step follows the TDD cycle: write tests → add minimal stubs to compile → verify red → implement → verify green → human testing checkpoint.

##### Step 1: Watered Tile State + Decay + Serialization

**Tests** (in `internal/game/map_test.go`):
- `SetManuallyWatered(pos)` / `IsManuallyWatered(pos)` basic set/get
- `IsManuallyWatered()` returns false for non-watered positions
- `IsWet(pos)` returns true for manually watered tiles (integration with existing wet check)
- `IsWet(pos)` still returns true for water-adjacent tiles (regression)
- `UpdateWateredTimers(delta)` decrements timer
- `UpdateWateredTimers` removes tile when timer reaches 0
- `WateredPositions()` returns all manually watered positions

**Serialization test** (in `internal/ui/serialize_test.go`):
- Watered tile timers round-trip through save/load

**Config:** `WateredTileDuration = 360.0` (3 world days = 360 game seconds)

**Implementation** (follows tilled soil pattern):
- Add `wateredTimers map[types.Position]float64` to Map struct, initialize in `NewMap()`
- `SetManuallyWatered(pos)` — sets `m.wateredTimers[pos] = config.WateredTileDuration`
- `IsManuallyWatered(pos) bool` — returns `m.wateredTimers[pos] > 0`
- `WateredPositions() []types.Position` — iterates map, returns slice
- `UpdateWateredTimers(delta float64)` — decrement all timers, delete entries when ≤ 0
- Update `IsWet(pos)` to also return true when `m.wateredTimers[pos] > 0` (existing water-adjacent check remains)
- `SaveState`: add `WateredTiles []WateredTileSave` with position + remaining time, `json:"watered_tiles,omitempty"`
- `ToSaveState` / `FromSaveState`: serialize/restore watered timers
- Wire `UpdateWateredTimers(delta)` into `updateGame()` alongside `UpdateDeathTimers`

**Architecture patterns:** Follows water terrain pattern (map state, O(1) lookups). `IsWet()` becomes the unified "is this tile wet from any source" check — growth multiplier code in lifecycle.go uses `IsWet()` already, so manually watered tiles automatically get the growth bonus with zero changes to lifecycle.go.

**[TEST] Checkpoint — Watered tile state:**
- `go test ./...` passes
- Build and run: use `/test-world` to create a world with manually watered tiles. Cursor over watered tile shows "Wet" in details panel (existing `IsWet()` drives the annotation). Verify wet status decays — after fast-forwarding, the "Wet" annotation disappears.
- Note: green tilled rendering not visible yet (Step 2). Only the details panel annotation is testable here.

##### Step 2: Wet Tilled Rendering + Details

**Tests:** None (UI rendering per CLAUDE.md)

**Implementation** (`internal/ui/view.go`, `internal/ui/styles.go`):
- In `renderCell()`: when rendering tilled soil fill (`═══` or `═X═`), check `gameMap.IsWet(pos)`. If wet, use green style instead of `growingStyle` (olive). Applies to both empty tilled tiles and entity fill padding.
- Details panel: when cursor is on a manually watered tile, show "Watered" annotation (in addition to existing "Wet" for water-adjacent and "Tilled soil" for tilled)

**[TEST] Checkpoint — Wet tilled rendering:**
- `go test ./...` passes
- Build and run: use `/test-world` to create tilled soil near water. Tiles adjacent to water render green `═══`. Tilled tiles far from water render olive `═══`. Cursor over wet tilled tile shows "Tilled soil" + "Wet". Cursor over watered tile (once watering exists) shows "Tilled soil" + "Watered".

##### Step 3: Discovery from ActionFillVessel + Activity Registration + Order UI

_Detailed breakdown TBD — high-level scope:_
- Add `CheckDiscovery` call in ActionFillVessel completion handler (update.go), passing the vessel as item
- Register `waterGarden` activity in ActivityRegistry: Category "garden", IntentOrderable, AvailabilityKnowHow, triggers: `{ActionFillVessel, ItemType: "vessel"}`
- Order UI: selecting "Water Garden" in garden step 1 creates order immediately (no area selection, no sub-menu)
- `DisplayName()`: "Water garden"
- `IsOrderFeasible` for waterGarden: vessel exists + water exists + dry tilled planted tile exists
- Serialization: no new order fields needed (no TargetType, no LockedVariety)

##### Step 4: findWaterGardenIntent + ActionWaterGarden (basic watering) ✅

**Design decision:** ActionWaterGarden follows the **ordered action pattern** (same as TillSoil, Plant), not the self-managing pattern. The handler completes one watering, clears intent, and lets `CalculateIntent` re-evaluate next tick. This provides needs interruption and order pause/resume. See architecture.md "Action Categories" section. Step 5's procurement/fill phases will use shared helpers (`RunVesselProcurement`) within the same handler, but Phase 3 (tile watering) clears intent between tiles.

**Implementation:**

Tests (TDD):
- **Anchor test** (update_test.go): Character with water-filled vessel, assigned waterGarden order, at dry tilled planted tiles → waters all dry tiles, each consumes 1 water unit, order completes when no dry tiles remain
- **Vessel empty test** (update_test.go): Character with 1 water unit, 2 dry tiles → waters 1 tile, vessel empties, intent clears (Step 5 adds refill)
- `findWaterGardenIntent` returns ActionWaterGarden targeting nearest dry tilled planted tile (order_execution_test.go)
- `findWaterGardenIntent` returns nil when no vessel with water in inventory (order_execution_test.go)
- `findWaterGardenIntent` returns nil when all tilled planted tiles are already wet (order_execution_test.go)

Code changes:
- `ActionWaterGarden` action type constant in character.go
- `findWaterGardenIntent()` in order_execution.go (replaces stub): find vessel with water in inventory → find nearest dry tilled planted tile → return ActionWaterGarden intent with vessel as TargetItem
- Helper: `findCarriedVesselWithWater(char)` — returns first inventory vessel containing liquid
- Helper: `FindNearestDryTilledPlanted(pos, items, gameMap)` — exported, returns nearest tilled+growing+not-wet position
- ActionWaterGarden handler in update.go applyIntent: move to dest → accumulate progress (ActionDurationShort) → SetManuallyWatered + DrinkFromVessel → clear intent → check order completion inline (no dry tiles → CompleteOrder)
- **Architecture pattern:** Ordered action pattern — handler clears intent after each tile, `findWaterGardenIntent` handles re-entry target selection

**[TEST] Checkpoint — Basic watering:** ✅
- `go test ./...` passes ✅
- Build and run with `/test-world` providing character with water-filled vessel + dry tilled planted tiles ✅
- Character moves to nearest dry tile, waters it (tile turns green), continues to next ✅
- Water unit consumed per tile via DrinkFromVessel ✅
- If vessel empties, character stops (Step 5 adds refill) ✅
- Order completes when all planted tiles are wet ✅
- Order abandons appropriately when vessel lacks water ✅
- Polish items deferred to cleanup: log wording "Watered garden tile", wet soil color scheme

##### Step 5a: Extract `RunWaterFill` shared helper from ActionFillVessel

Pure refactor — no behavior change. Extract the "move to water + fill vessel" logic from ActionFillVessel's Phase 2 (update.go) into a shared helper in picking.go, sibling of `RunVesselProcurement`. ActionFillVessel calls the new helper instead of inline code.

**Tests (TDD):**

Existing fetch water tests serve as regression coverage — no new tests needed since behavior is identical. Run `go test ./...` to confirm green after refactor.

**Implementation:**

- Extract `RunWaterFill(char, vessel, actionType, gameMap, log, registry, delta) WaterFillStatus` into picking.go
  - Returns status enum: `FillReady` (vessel full, proceed), `FillApproaching` (moving to water), `FillInProgress` (filling this tick), `FillFailed` (no water reachable)
  - **Fills to full regardless of starting state** — the helper always fills. It's the caller's responsibility to decide when filling is needed. `findFetchWaterIntent` only triggers when the vessel is empty; `findWaterGardenIntent` routes to fill phase only when the vessel has no water.
  - Logic extracted from ActionFillVessel Phase 2: find nearest water → find cardinal-adjacent tile → move to it → accumulate progress → fill vessel → trigger discovery
  - Caller (ActionFillVessel handler) handles movement via `moveWithCollision` when `FillApproaching` returned
  - `actionType` parameter: the helper rebuilds `char.Intent` after procurement clears it; the caller passes its own action type (e.g., `ActionFillVessel`, `ActionWaterGarden`) so the intent stays consistent
  - Note: discovery trigger (`TryDiscoverKnowHow` for ActionFillVessel) stays in the helper — it fires on fill completion regardless of calling context. Water Garden characters can also discover Water Garden know-how from filling, which is narratively correct.
- Remove `TestApplyIntent_FillVessel_DoesNotOverfillPartialVessel` — tests a state (partial vessel triggering fill) that doesn't occur in actual code. Re-add if partial-fill triggers are introduced.
- Refactor ActionFillVessel handler (update.go): replace inline Phase 2 code with `RunWaterFill` call. Phase 1 (`RunVesselProcurement`) unchanged.
- **Architecture pattern:** Follows `RunVesselProcurement` shape exactly — tick helper with status enum, caller owns movement, stateless phase detection.

**No [TEST] checkpoint** — pure refactor, no user-visible change. Existing tests confirm behavior preserved.

##### Step 5b: Wire procurement + fill into Water Garden flow ✅

The full Water Garden lifecycle: character gets order → procures vessel (if needed) → fills at water → waters tiles → vessel empties → refills → continues → no dry tiles → complete. Initial procurement and mid-order refill are the same cycle — the character always needs water, and the phases detect what's missing.

**Design decision (refined during implementation):** ActionWaterGarden is an ordered action — intent clears at every phase boundary and between every tile. `findWaterGardenIntent` detects the phase each resumption tick and sets up the intent for that phase. The handler checks world state to know which phase it's in and calls the appropriate shared helper (`RunVesselProcurement` or `RunWaterFill`) or does the watering work. This follows the ordered action pattern: `findXxxIntent()` handles target selection on each resumption tick (architecture.md "Adding a new ordered action"), and the handler completes one work unit and clears intent. The shared tick helpers handle multi-tick operations within a phase (pickup takes ticks, filling takes ticks) without clearing intent — intent only clears when the phase completes.

This means the ordered action's pause/resume mechanism works at every phase boundary — if needs become pressing during procurement, filling, or between tiles, the order pauses naturally via `CalculateIntent`.

**Tests (TDD):**

- **Full cycle test** (update_test.go): Character with no vessel, ground vessel available, water source, 3 dry tilled planted tiles → character procures vessel → fills at water → waters all 3 tiles → order completes
- **Refill test** (update_test.go): Character with vessel containing 1 water unit, 3 dry tilled planted tiles → waters 1 tile → vessel empty → refills at water → waters remaining 2 → order completes
- **No vessel abandonment test** (order_execution_test.go): `findWaterGardenIntent` with no vessel anywhere → returns nil
- `findWaterGardenIntent` **phase detection tests** (order_execution_test.go):
  - Has vessel with water → returns ActionWaterGarden targeting dry tile (existing test, keep)
  - Has empty vessel → returns ActionWaterGarden targeting water-adjacent position
  - No vessel, ground vessel exists → returns ActionWaterGarden targeting ground vessel
  - No vessel, no ground vessel → returns nil (existing test updated)

**Implementation:**

1. **`findWaterGardenIntent`** — expand from Phase 3 only to full phase detection:
   - Check for dry tilled planted tiles first — if none, return nil (order complete)
   - Has vessel with water in inventory? → Phase 3: target nearest dry tilled planted tile (existing code)
   - Has empty vessel in inventory? → Phase 2: find nearest water source, target cardinal-adjacent tile. Set `TargetItem` to the vessel.
   - No vessel in inventory? → Phase 1: find ground vessel via `findEmptyGroundVessel`. Target ground vessel position. Set `TargetItem` to ground vessel.
   - Nothing available (no vessel anywhere) → return nil (abandon)
   - Always returns `ActionWaterGarden` — the handler detects the phase via stateless world-state checks.

2. **ActionWaterGarden handler** (update.go) — stateless phase detection, shared helpers:
   - Phase 1: vessel is `TargetItem` and is on the ground (`gameMap.ItemAt(vessel.Pos()) == vessel`) → call `RunVesselProcurement`. On `ProcureReady`, clear intent (ordered pattern — next tick `findWaterGardenIntent` routes to Phase 2 or 3). On `ProcureApproaching`, `moveWithCollision`. On `ProcureFailed`, clear intent.
   - Phase 2: vessel is `TargetItem`, in inventory, empty (`len(vessel.Container.Contents) == 0`) → call `RunWaterFill`. On `FillReady`, clear intent (next tick routes to Phase 3). On `FillApproaching`, `moveWithCollision`. On `FillFailed`, clear intent.
   - Phase 3: vessel has water → existing watering code (move to tile → water → clear intent → check completion).
   - **Phase detection uses same raw state checks as the helpers themselves** — `RunVesselProcurement` uses `gameMap.ItemAt()` internally (picking.go:837), `RunWaterFill` checks vessel contents. No new abstractions needed.

3. **Architecture pattern:** Ordered action pattern (architecture.md "Action Categories" + "Adding a new ordered action") — `findWaterGardenIntent` handles target selection on each resumption tick, handler completes one work unit and clears intent. Shared tick helpers (`RunVesselProcurement`, `RunWaterFill`) handle multi-tick operations within a phase without clearing intent. Intent clears at phase completion, giving `CalculateIntent` a re-evaluation point for needs interruption and order pause/resume.

**Bugs fixed during implementation:**
- **continueIntent look-hijack**: `continueIntent` was converting `ActionWaterGarden` intents to `ActionLook` when the character arrived adjacent to the target vessel (due to the `TargetItem != nil && DrivingStat == ""` broad heuristic). Workaround: added `ActionWaterGarden` to the look-transition exclusion list in `movement.go`. Tracked as Post-Slice 8 cleanup.
- **continueIntent multi-phase path recalculation**: Added an `ActionWaterGarden` early-return block in `continueIntent` so multi-phase path recalculation works correctly (vessel on ground in Phase 1, vessel in inventory in Phase 2/3). Without this, the generic path's `ItemAt` check would nil the intent when the vessel moved to inventory.

**[TEST] Checkpoint — Full Water Garden flow:** ✅
- `go test ./...` passes ✅
- Build and run:
  - Order Water Garden when character has no vessel. Character procures vessel, fills at water, waters planted tiles. ✅
  - Order Water Garden when character has empty vessel. Character fills at water, then waters. ✅
  - Order Water Garden when character has water. Character waters immediately. ✅
  - Watered tiles turn green `═══`. Dry tilled tiles remain olive `═══`. ✅
  - Character continues watering until all planted tiles wet or vessel empty ✅
  - If vessel empty and more tiles need watering, character refills at nearest water source and continues ✅
  - Plants on watered tiles grow faster than dry tilled plants ✅
  - Watering effect wears off after ~3 world days (use time skip to verify) ✅
  - Tiles adjacent to ponds are always wet (no watering needed, already green) ✅
  - Wet tilled-but-unplanted tiles also display green ✅
  - Multi-assignment: two Water Garden orders, two characters watering ✅
  - Save and load preserves watered tile timers ✅

**[TEST] Final Checkpoint — Full Part II Integration:**

Start a new world and play through the full garden lifecycle:
- Till soil → plant seeds → water garden → sprouts grow → full plants → natural reproduction (also through sprout phase)
- Growth hierarchy visible: watered tilled > dry tilled > water-adjacent wild > plain wild
- Food chain: gourds eaten → seeds → planted → grow → more gourds (sustainable food loop)
- Flower cycle: forage flower → seed → plant → grow → more flowers → more seeds
- Satiation differences noticeable: gourd satisfies much more than berry
- Water management: characters fetch water, drink from vessels, water gardens

**[DOCS]** Final doc pass for Part II.

**[RETRO]** Run /retro.

**Reqs reconciliation:** Lines 110-123. _"Discoverable by looking at sprout or water source or hollow gourd"_ — sprout via ActionLook trigger, water source via ActionFillVessel trigger (descoped lookable terrain, captured in triggered-enhancements.md), hollow gourd deferred (minor). _"Orderable"_ ✓. _"No Recipe Needed"_ ✓. _"Pre-requisite: at least one vessel full with water"_ ✓ (self-managing procurement + fill). _"water the closest dry tilled planted tile"_ ✓. _"Watering uses 1 unit of water"_ ✓. _"watered tile changes in appearance to green"_ ✓. _"tilled but not planted... green instead of olive"_ ✓. _"If have water left... look for closest dry tilled planted tile"_ ✓. _"vessel is empty, then refill vessel and continue"_ ✓. _"once there are no remaining dry tilled planted tiles, order is complete"_ ✓. _"more than one character can be watering at once"_ ✓.

---
#### Slice 8 Polish (from human testing feedback)

- **Log wording**: "Watered garden tile" → more natural phrasing like "Watered garden plot" or "Watered garden plant"
- **Wet tilled soil colors**: Current green/olive don't look natural and aren't easily differentiable. Watered soil doesn't turn green in real life — needs a color rethink (e.g., darker brown for wet vs lighter for dry, or subtle blue tint)

---
#### Post-Slice 8: `continueIntent` Look-Transition Cleanup

**Status:** Workaround in place (exclusion list). Needs `/refine-feature` to evaluate the right long-term fix.

**Problem:** `continueIntent` (movement.go ~line 483) converts any intent with `TargetItem != nil && DrivingStat == ""` to `ActionLook` when the character arrives adjacent to the target. This is the arrival handler for `findLookIntent`'s walk-to-item flow (which uses `Action: ActionMove`). But the condition is too broad — it catches any action with a TargetItem and no DrivingStat. Current workaround: exclusion list (`ActionPickup`, `ActionForage`, `ActionWaterGarden`).

**Root issue:** `findLookIntent` sets `Action: ActionMove` for the walking phase, then relies on `continueIntent` to convert to `ActionLook` on arrival. This conflates two concerns — `continueIntent` should recalculate paths, not change action types. The decision to look should live entirely in `findLookIntent`.

**Proposed fix (needs evaluation):** `findLookIntent` should set `Action: entity.ActionLook` for the walking phase (not `ActionMove`), and `continueIntent`'s arrival transition should check `intent.Action == entity.ActionLook` instead of the broad heuristic. Then only intents deliberately created as looks get the arrival behavior.

**`RunVesselProcurement` interaction — may also need reevaluation:**

Three actions use `RunVesselProcurement` for vessel pickup phases, but each interacts with `continueIntent` differently:
- `ActionForage` — excluded from look-hijack by name
- `ActionFillVessel` — has its own early-return block in `continueIntent` that bypasses the look-hijack entirely
- `ActionWaterGarden` — needed both an exclusion AND its own early-return block (added during Step 5b)

This asymmetry means the exclusion list and/or the number of action-specific early-return blocks will grow as new actions use `RunVesselProcurement`. The `ActionLook`-based fix would eliminate the exclusion list, but the early-return block pattern would persist for any action where `TargetItem` can be in inventory (not on the map). Evaluate during this cleanup whether:
1. The `ActionLook` proposal is sufficient (eliminates exclusion list, early-return blocks remain)
2. `continueIntent` needs a more general pattern for multi-phase actions (reduces per-action blocks)
3. Ordered actions should use `ActionPickup` for vessel procurement (aligns with TillSoil/Plant, but requires solving the "liquid isn't an existing item" gap for the fill phase)

---
### Slice 9: Tuning and Enhancements

**Reqs covered:**
- Food Turning: Satiation tiers (lines 133-137)
- Food Turning: Growth speed tiers (lines 138-141)

#### Design: Food Satiation Tiers

Replace the flat `FoodHungerReduction` constant with per-item-type values:

| Tier | Points | Items |
|------|--------|-------|
| Feast | 50 | Gourd |
| Meal | 25 | Mushroom |
| Snack | 10 | Berry, Nut |

#### Design: Growth Speed Tiers

Replace flat `SproutDuration` with per-item-type values. Affects both sprouting time and reproduction intervals:

| Tier | Items | Direction |
|------|-------|-----------|
| Fast | Berry, Mushroom | Shorter sprouting and reproduction times |
| Medium | Flower | ~6 min sprout, existing reproduction |
| Slow | Gourd | Longer sprouting and reproduction times |

Also tune growth multiplier values:
- `TilledGrowthMultiplier` (currently 1.25 — placeholder)
- `WetGrowthMultiplier` (currently 1.25 — placeholder)

**Deferred from Slice 6:** Flat 30s sprout duration used for testing. This slice differentiates by tier and tunes multiplier values.

#### Design: Vessel Drinking Thirst Trigger

Evaluate whether carrying a water vessel should allow drinking at a lower thirst threshold than walking to a water source. Currently both trigger at Moderate tier (51+). Carried vessel at distance 0 is already prioritized by proximity.

**Deferred from Slice 7:** Keep same threshold for initial implementation. Tune here if vessel drinking feels like it should trigger earlier.

**Outcomes:**

1. Per-item-type satiation amounts replace flat hunger reduction
2. Per-item-type growth speed tiers replace flat SproutDuration
3. Tuned growth multiplier values for tilled/wet bonuses
4. Evaluate vessel drinking thirst threshold

**[TEST] Checkpoint — Satiation and growth tuning:**
- Eat a gourd → larger hunger reduction than eating a mushroom → larger hunger reduction than eating a berry/nut
- Berry/mushroom sprouts mature noticeably faster than gourd sprouts
- Flower sprouts take ~6 minutes
- Growth hierarchy feels right: tilled+wet > tilled or wet alone > wild
- Vessel drinking trigger feels natural (adjust threshold if needed)

**[DOCS]** Update README, CLAUDE.md, game-mechanics, architecture as needed.

**[RETRO]** Run /retro.

**Feature questions:**

- Seed symbol and color? Likely `.` in parent's color. Confirm during implementation.
- Seed description format: "warty green gourd seed"? Include all parent attributes or just color + type?
- Flower foraging cadence: can the same flower be foraged repeatedly? Timer per flower? Consider matching flower reproduction cadence.
- Does flower seed go through standard pickup logic at character position, or special handling?
- Nut satiation: grouped with berry at Snack (10). Confirm this feels right during testing.
- Exact growth speed tier values: need playtesting to find good feel. Start with 2x/3x ratios between tiers?

## Triggered Enhancements to Monitor

These are from [docs/triggered-enhancements.md](triggered-enhancements.md). They may be triggered during Gardening but don't need to be planned upfront.

| Enhancement                            | Trigger During Gardening                                  | Action                                                                                         |
| -------------------------------------- | --------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| **Order completion criteria refactor** | Adding Till Soil, Plant, Water Garden (3 new order types) | Monitor if completion logic in update.go exceeds ~50 lines. Refactor to handler pattern if so. |
| **ItemType constants**                 | Adding stick, shell, hoe, seed (4 new types, total ~9)    | Evaluate after Part I whether string comparisons are error-prone.                              |
| **Category formalization**             | Hoe is first "tool" category                              | Note pattern but defer to Construction per triggered-enhancements.md.                          |
| **Preference formation for beverages** | Fetch Water introduces water vessels as drinkable         | Evaluate during Slice 7; likely defer until actual beverage variety exists.                    |
| **Action duration tiers**              | Craft Hoe and Till Soil both need "medium" duration       | ✅ Completed in Part I Slice 2. Short/Medium/Long defined.                                     |
| **UI extensibility refactoring**       | Area selection UI is new pattern                          | Document approach for reuse by Construction.                                                   |

## Opportunistic Additions to Consider

From [docs/randomideas.md](randomideas.md):

| Idea                   | Opportunity                                                                     | Status                                                                           |
| ---------------------- | ------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| **Edible Nuts**        | Same "drop from canopy" pattern as sticks.                                      | **Included in Slice 1.** Edible, forageable, brown `o`. Plantable deferred to Tree reqs. |
| **Order Selection UX** | Gardening adds more order types, making scrolling more painful.                 | Consider after Part I when the pain is fresh. Not in scope for gardening itself.  |
