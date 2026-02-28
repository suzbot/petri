# Construction Phase Plan

Requirements: [Construction-Reqs.txt](Construction-Reqs.txt)

---

## Overview

Characters gather materials and construct small buildings from grass, sticks, or bricks. Construction introduces bundles (stackable materials), a new plant type (tall grass), clay terrain, brick crafting, a new entity type (Construct) for structures and future furniture, fence and hut building, and two new preference systems (material preferences and activity preferences). This sets the stage for future phases where buildings protect stored items and provide better sleep.

---

## Resolved Design Decisions

1. **Bundles: BundleCount field on Item.** Bundleable items (sticks, grass) carry a `BundleCount int` on the Item struct. `MaxBundleSize` config per item type (sticks: 6, grass: 6). Single items start at BundleCount=1. Picking up a matching item increments the count. Dropped bundles are one entity on the ground with a count. This is the simplest representation — one field, no new structs. Follows **Start With the Simpler Rule**.

2. **Vessel exclusion: Explicit config set.** Add `VesselExcludedTypes` set in config (sticks, grass). Check in `AddToVessel`, `CanVesselAccept`, and gather order vessel-procurement branching. Replaces the current implicit exclusion (sticks have no variety, so `AddToVessel` fails on registry lookup). Making this explicit is necessary because grass needs registered varieties (for seed inheritance) but still can't go in vessels.

3. **Seed extraction: "Extract" as new orderable activity category.** Subcategories: "Flower Seeds", "Grass Seeds". Non-destructive interaction with a growing plant — character approaches, performs extraction, seed appears. The plant is unharmed. This is a distinct pattern from foraging (food-centric), harvesting (destructive), and gathering (picks up ground items). Follows **Consider Extensibility** — future extraction targets include pigment, sap, essence.

4. **Construct: New entity type.** Separate from Feature (natural world elements like leaf piles). Construct represents character-built things with a ConstructType/Kind hierarchy mirroring ItemType/Kind on Items. ConstructType: "structure" (immovable, impassable except doors) and future "furniture" (movable, passability varies). Kind: "wall", "fence", "door", and future "bed", "table", etc. Follows **Isomorphism** — constructs and natural features are different world concepts, so they get different places in code.

5. **Clay: New terrain type.** `clay map[Position]bool` on Map with `IsClay(pos)` query. Passable tiles adjacent to water, clustered (6-10 tiles, must be adjacent to water and to at least one other clay tile). Infinite source like water for drinking. Rendering: reddish brown pattern.

6. **Hut area selection: Fixed 4x4 footprint.** Reuses the area selection UI pattern from tilling. Player positions a 4x4 outline. Interior is 3x3. One outline tile is a door (character-passable only). Worker marks tiles, drops supplies, then works each tile.

---

## Deferred Questions (for refinement sessions)

- **EnsureHasBundle helper design**: How does a construction worker procure a full bundle of 6? Does EnsureHasBundle gather one at a time, or is there a "fill bundle" loop? Discuss when refining Step 5 (Build Fence).
- **Door passability mechanism**: Currently passability is boolean. Doors need "character-passable but creature-impassable" (for future Threats phase). May need a passability enum or just Passable=true until creatures exist. Discuss when refining Step 5.
- **Hut supply management**: The reqs say "marked tiles for construction get all their supplies dropped on them and then each tile with all its supplies is worked for a duration." How does the worker distribute supplies across tiles? Discuss when refining Step 6.
- **Multiple hut materials**: Can a single hut mix materials (some thatch tiles, some brick)? Reqs imply one material per hut. Confirm when refining Step 6.
- **Extract discovery triggers**: What experience discovers Extract know-how? Looking at flowers/grass? Picking up seeds? Discuss when refining Step 2.
- ~~**Grass seed vs grass item**~~: **Resolved in Step 1 refinement.** Different items. Harvested grass = "grass" material with BundleCount. Grass seed = ItemType "seed", Kind "grass seed" (produced by extraction in Step 2).

---

## Triggered Enhancements to Evaluate

From [triggered-enhancements.md](triggered-enhancements.md):

- **Category type formalization** — Explicitly triggered by Construction. Evaluate whether `VesselExcludedTypes` set is sufficient or whether formal item categories are needed. Evaluate during Step 1 (Bundles) when the vessel exclusion pattern is implemented.
- **`continueIntent` early-return block consolidation** — Already at 6 blocks, trigger threshold met. If construction adds more multi-phase actions, consolidate first. Evaluate during Step 5 (Build Fence) when the first construction action is added.
- **Order-aware simulation for e2e testing** — Explicitly triggered: "Construction adds multiple new ordered actions with supply procurement." Add simulation tests for construction orders. Evaluate during Step 5.

From [randomideas.md](randomideas.md):

- **Activity Preferences** — Already in the construction requirements. Implemented as Step 8.
- **Flower seed extraction** — Resolved by the Extract activity (Step 2).

---

## Implementation Steps

### Step 1: Tall Grass + Bundles

**Anchor story:** The player sees tall, pale green grass growing and spreading across the world. A character harvests some grass — it appears in their inventory as "a bundle of grass (1)." They harvest another nearby — the bundle grows to 2. Sticks now also stack as bundles when picked up. The player notices neither grass nor sticks can be placed in vessels — they're too large. When a character fills a bundle to 6, they drop it on the ground and complete the order.

**Resolved design decisions (from discussion):**
- Grass seed vs grass item: **Different items.** Grass plant on map = ItemType "grass" with PlantProperties. Harvested grass = "grass" material with BundleCount (construction material). Grass seed = ItemType "seed", Kind "grass seed" (produced by extraction in Step 2, not harvest). Harvesting is destructive (removes plant). Extraction (Step 2) is non-destructive.
- Grass symbol: `W` — visually reads as blades/fronds, distinct from all current symbols.
- Grass spawn count: Same as other plant types (`ItemSpawnCount` = 20), not 2x as originally required. Aggressive reproduction via Fast tier handles proliferation naturally.
- Grass lifecycle: Fast maturation (same as mushroom, ~1 world day), Fast reproduction (same as berry, ~2 world days per plant). Fastest-growing plant overall.
- Grass varieties: Single variety (pale green, no pattern, no texture). Still registered in VarietyRegistry — needed for seed extraction in Step 2.
- Bundle-full behavior: When harvest/gather fills a bundle to max (6), character **drops the bundle** at current position and order **completes**. One bundle per order. Applies to both grass harvest and stick gather.
- EnsureHasBundle helper: **Deferred to Step 5** (Build Fence). Step 1 only needs pickup-merges-into-bundle, not "gather until bundle reaches target count."

---

#### ✅ Step 1a: Bundle Mechanics + Modified Pickup

**Anchor story:** A character picks up a stick and sees "bundle of sticks (1)" in their inventory. They pick up another — the bundle grows to (2). They try to put sticks in a vessel — it won't fit. When the bundle reaches 6, they can't pick up any more sticks.

**What's new:**
- `BundleCount int` field on Item struct (`entity/item.go`)
- `MaxBundleSize` config map in `config/config.go`: `{"stick": 6, "grass": 6}`
- `VesselExcludedTypes` config set in `config/config.go`: `{"stick", "grass"}`
- `IsBundleable(itemType)` config helper — returns true if type has a MaxBundleSize entry
- `NewStick()` constructor sets `BundleCount: 1`
- Modified `Pickup()` in `picking.go`:
  - Before vessel check: if item is vessel-excluded, skip vessel path entirely
  - New path: if character carries a bundle of same ItemType with room, increment BundleCount on carried bundle, remove picked-up item from map, return new `PickupToBundle` result
  - `PickupToBundle` does NOT clear intent (parallel to `PickupToVessel` — caller decides continuation)
  - If no existing bundle and inventory has space: add item to inventory as new bundle (BundleCount already set by constructor)
- Modified `CanPickUpMore()` in `picking.go`: returns true if character carries a non-full bundle of a bundleable type (even if inventory is otherwise full)
- Modified `Description()` in `entity/item.go`: when `BundleCount > 0`, returns `"bundle of [plural type] ([count])"` (e.g., "bundle of sticks (3)")
- `AddToVessel` / `CanVesselAccept` in `picking.go`: early return false for VesselExcludedTypes (belt-and-suspenders — explicit check alongside existing variety-absence check for sticks)
- Serialization: `BundleCount` added to `SaveItem` in `save/state.go`, round-trip in `serialize.go`

**Architecture patterns:**
- Item Model extension — new field with config-driven behavior. Follows **Start With the Simpler Rule** — one int field, not a new struct.
- **Source of Truth Clarity** — VesselExcludedTypes is explicit config replacing implicit variety-absence check for sticks. Grass has a registered variety but still can't go in vessels — without explicit exclusion, AddToVessel would succeed for grass.
- Pickup Result Pattern — `PickupToBundle` parallels `PickupToVessel` (don't clear intent, let caller decide). Follows **Follow the Existing Shape**.
- **Anti-pattern to avoid:** Don't add bundle-merging logic inside `AddToVessel` or the vessel path. Bundles are a separate concept from vessel stacking — they share the "multiple items in one slot" idea but have completely different mechanics (inventory item vs container contents).

**Reqs reconciliation:**
- "A stack of sticks is called a bundle. A bundle of sticks can have up to 6 pieces in it." → BundleCount field, MaxBundleSize config, Description() format
- "a picked up stack can be added to until it has max pieces" → Pickup() bundle merge path, CanPickUpMore() bundle check
- "Harvested grass CANNOT be placed in a vessel (too large)" → VesselExcludedTypes config, AddToVessel/CanVesselAccept checks
- Sticks implicit vessel exclusion → now explicit via VesselExcludedTypes

**Tests (TDD):**
- Unit: Pickup merges into existing bundle (count increments, map item removed)
- Unit: Pickup starts new bundle when none carried (BundleCount=1 in inventory)
- Unit: Pickup fails when bundle is at max (6) and no inventory space
- Unit: CanPickUpMore returns true when bundle has room, false when full
- Unit: AddToVessel returns false for VesselExcludedTypes items
- Unit: CanVesselAccept returns false for VesselExcludedTypes items
- Unit: Description() returns "bundle of sticks (3)" format for bundled items
- Unit: Save/load round-trips BundleCount

[TEST] ✅ Create a world with sticks. Verify sticks show as "bundle of sticks (1)" in inventory. Pick up more — count increments. Can't exceed 6. Can't put in vessel. Descriptions correct. Save/load preserves bundle state.

Bugs found during testing:
- Characters with full inventory of non-target items couldn't take gather orders — `canGatherMore` rejected them before drop-non-target logic could run. Fixed by checking bundle mergeability before the inventory-full gate.
- Continuation handler didn't check for a full bundle, so the character kept gathering past max. Fixed by adding a full-bundle check in `FindNextGatherTarget`.
- Drop-before-pickup didn't account for bundle mergeability — items that merge into an existing bundle don't need a free slot, so the drop was unnecessary.
- Completion handler in `apply_actions.go` wasn't calling `DropCompletedBundle`, so the full bundle stayed in inventory instead of landing on the ground.

[DOCS] ✅

[RETRO]

---

#### ✅ Step 1b: Tall Grass Plant Type

**Anchor story:** The player starts a new world and sees pale green `W` characters scattered across the map — tall grass. Over time, sprouts appear adjacent to grass and mature into new grass. Grass is everywhere.

**What's new:**
- `CharGrass = 'W'` symbol in `config/config.go`
- `NewGrass(x, y int)` constructor in `entity/item.go` — ItemType "grass", Color pale green, Plant with IsGrowing=true, BundleCount=1, not edible
- `config.ItemLifecycle["grass"]` — SpawnInterval: ReproductionFast (12.0), DeathInterval: 0 (immortal)
- `config.SproutDurationTier["grass"]` — SproutDurationFast (120.0)
- Grass in `GetItemTypeConfigs()` — Colors: `{types.ColorPaleGreen}`, no patterns/textures, not edible, not poisonable, Plantable: false (grass material is not plantable — grass *seeds* are, in Step 2), Sym: CharGrass, SpawnCount: ItemSpawnCount
- Variety generation: grass gets variety registration (single variety — pale green). Needed for seed extraction in Step 2.
- `createItemFromVariety` in `world.go`: add "grass" case → `entity.NewGrass(x, y)`
- `MatureSymbol` in `lifecycle.go`: add "grass" case → `config.CharGrass`
- `types.ColorPaleGreen` — new Color constant if not already defined (check types.go)

**Architecture patterns:**
- Adding New Plant Types checklist (architecture.md): constructor, config entries, variety generation, GetItemTypeConfigs. Every touchpoint covered.
- **Follow the Existing Shape** — grass follows berry/mushroom/gourd pattern exactly. Same lifecycle system, same spawn system, same variety registration.
- **Consider Extensibility** — variety registration with single variety supports future grass variants if needed.

**Reqs reconciliation:**
- "Spawns at world creation" → SpawnItems via GetItemTypeConfigs with SpawnCount
- "Pale Green" → ColorPaleGreen, single variety
- "Growing, reproduces like other plants, fast lifecycle" → PlantProperties, ItemLifecycle Fast tier, SproutDurationTier Fast
- "As a plant, should show up on the harvest list" → grass has Plant.IsGrowing=true, so growingItemExists() and findNearestItemByType(growingOnly=true) find it naturally. No special wiring needed.

**Tests (TDD):**
- Unit: NewGrass creates item with correct attributes (type, color, symbol, plant properties, BundleCount=1)
- Unit: Grass variety is registered during GenerateVarieties
- Unit: Grass appears in GetItemTypeConfigs
- Verify grass spawns on world generation (visual check)

[TEST] Start new world. See pale green W characters (grass). Over time, grass sprouts appear and mature. Grass spreads naturally via the lifecycle system. Details panel shows grass description correctly.

Bugs found during testing:
- Grass rendered without color — `ColorPaleGreen` missing from render switch in view.go. Added `paleGreenStyle` (sage, ANSI 108) and wired into the item color switch.
- Grass vanished on save/load — `serialize.go` symbol restoration switch missing `case "grass"`. Added.
- Grass sprouts matured without `BundleCount` — `spawnItem` doesn't copy BundleCount, and `UpdateSproutTimers` didn't restore it. Added BundleCount restoration for bundleable types on maturation.
- Grass population unchecked — no death timer. Added `DeathInterval: 48.0` (same as flowers). Observation tests show stable ~51 avg (down from ~65 immortal).

[DOCS] ✅

[RETRO] ✅

---

#### Step 1c: Harvest + Gather Integration with Bundles

**Anchor story:** The player creates a Harvest > Grass order. A character walks to tall grass, harvests it into a bundle, moves to the next grass, harvests again (bundle grows to 2, 3...). When the bundle reaches 6, the character drops it on the ground and completes the order. The player creates a Gather > Sticks order — same flow. The character gathers sticks into a bundle of 6, drops it, done.

**What's new:**
- `findHarvestIntent`: skip `EnsureHasVesselFor` when target is a vessel-excluded type. Character goes directly to pickup without vessel procurement. Check: if `config.IsVesselExcluded(target.ItemType)`, skip vessel procurement and go straight to "Ready to harvest."
- `findGatherIntent`: currently branches on variety presence. Add explicit VesselExcludedTypes check before the variety check — vessel-excluded types always take the direct path (same as current "no variety" path for sticks, but now explicit). For bundleable types, `CanPickUpMore()` replaces `HasInventorySpace()` as the capacity check (since bundle room counts).
- `applyPickup` handler in `apply_actions.go`: handle `PickupToBundle` result in both harvest-order and gather-order contexts. Treat like `PickupToVessel` — continue working (find next target). When `CanPickUpMore()` returns false (bundle full), drop the bundle and complete the order.
- Drop-on-complete behavior: when harvest/gather order completes because bundle is full, call `DropItem()` on the full bundle before `CompleteOrder()`. Bundle lands at character's current position.
- `FindNextHarvestTarget` / `FindNextGatherTarget`: use `CanPickUpMore()` instead of `HasInventorySpace()` so bundle room is considered.

**Architecture patterns:**
- Order execution pattern — harvest/gather bundle path parallels vessel path. Follows **Follow the Existing Shape**.
- `PickupToBundle` continuation mirrors `PickupToVessel` continuation. Same control flow, different storage mechanism.
- Drop-on-complete follows craft order pattern (crafted item drops on ground). Follows **Follow the Existing Shape**.
- **Anti-pattern to avoid:** Don't create a separate "bundle harvest" action type. Use the same `ActionPickup` action — the bundle mechanics are in the Pickup function, not the action type. This follows **Separation of Concerns by Cognitive Role** — what changes is the physical storage, not the intent.

**Reqs reconciliation:**
- "As a plant, should show up on the harvest list" → grass has IsGrowing, appears in harvest target list naturally
- "Harvested grass CANNOT be placed in a vessel, but can be carried or dropped as a stack" → VesselExcludedTypes skips vessel procurement, bundle merge in Pickup handles stacking
- "A bundle of sticks can have up to 6 pieces" + "a picked up stack can be added to until max pieces" → harvest/gather continues until CanPickUpMore returns false (bundle at max)
- Drop-on-complete → discussed and agreed during refinement, extends to both grass harvest and stick gather

**Tests (TDD):**
- Unit: findHarvestIntent skips vessel procurement for vessel-excluded types
- Unit: findGatherIntent uses CanPickUpMore for bundleable types
- Integration: harvest grass order end-to-end — character harvests grass into bundle, continues until full, drops bundle, order completes
- Integration: gather sticks order end-to-end — same flow with sticks
- Unit: FindNextHarvestTarget/FindNextGatherTarget return nil when bundle full
- Unit: drop-on-complete places bundle on map at character position

[TEST] Create harvest grass and gather sticks orders. Watch characters collect into bundles, continuing to next target after each pickup. When bundle hits 6, character drops it and order completes. Verify bundles on ground show correct count. Verify existing harvest (berries, mushrooms) still works with vessel path unchanged.

[DOCS]

[RETRO]

---

#### Step 1d: Polish, UI, + Save/Load Verification

**Anchor story:** The player drops a bundle of 3 sticks on the ground and sees an `X` symbol — visually distinct from a single stick's `/`. They select the bundle and the details panel shows "bundle of sticks (3)." They save, reload — everything preserved. Grass is still growing, bundles still have their counts.

**What's new:**
- **Bundle rendering on map:** Items with `BundleCount >= 2` render as `X` instead of their normal symbol. This gives the player a visual signal that a bundle is sitting on the ground, distinct from a single item. Single items (BundleCount=1) still render with their normal symbol (`/` for sticks, `W` for grass). Implementation: `renderCell()` in view code checks `BundleCount >= 2` and overrides the symbol. Add `CharBundle = 'X'` to config symbols.
- **Details panel:** When a bundle is selected, the details panel should show the bundle description (e.g., "bundle of sticks (3)") — this comes from `Description()` which was updated in Step 1a. Verify it renders correctly in the details panel — check that the existing item detail rendering path uses `Description()` and doesn't need modification.
- **Inventory panel:** Verify bundle items show their bundle description in the inventory list (`I` key). Same source — `Description()` — but confirm the inventory rendering path handles it.
- Save/load round-trip: verify BundleCount on items (inventory and ground), grass plant type (growing, sprouts), grass variety in registry
- Backward compatibility: loading a pre-bundle save where sticks have no BundleCount field. Sticks should default to BundleCount=0 (not bundled) or BundleCount=1 on load — need to handle gracefully. Decision: sticks with BundleCount=0 in old saves are treated as BundleCount=1 (they're always individual items, which is a bundle of 1). Add migration logic in deserialize.
- Verify NewStick() sets BundleCount=1 for newly ground-spawned sticks

**Architecture patterns:**
- Rendering: follows existing `renderCell()` pattern — entity-aware symbol override. Same approach used for sprouts (CharSprout overrides normal symbol). **Follow the Existing Shape.**
- Details/inventory display: relies on `Description()` as the single source of item identity text. **Source of Truth Clarity** — no separate bundle display logic, Description() handles it.

**Tests (TDD):**
- Unit: item with BundleCount >= 2 renders as 'X' on map
- Unit: item with BundleCount == 1 renders as normal symbol
- Round-trip test: save world with grass + bundles → load → verify all fields preserved
- Migration test: load old save without BundleCount → sticks get BundleCount=1

[TEST] Drop a bundle of 2+ items — verify `X` symbol on map. Select it — verify details panel shows "bundle of sticks (3)" (or similar). Check inventory panel shows bundle description. Save and reload — verify everything preserved. Start an old save (if available) — verify sticks work correctly.

[DOCS]

[RETRO]

---

### Step 2: Seed Extraction (Extract Activity)

**Anchor story:** The player creates an Extract > Flower Seeds order. A character who has discovered how to extract walks to a growing flower, spends time carefully collecting seeds, and obtains a flower seed — without harming the flower. The same mechanic works for Extract > Grass Seeds with tall grass. Seeds can be planted in tilled soil.

**What's new:**
- New `ActionExtract` action constant
- "Extract" activity category with subcategories in ActivityRegistry
- Extract intent finder and handler
- Non-destructive plant interaction pattern (plant stays, seed produced)
- Discovery triggers for extract know-how

**Architecture patterns:**
- Adding an Ordered Action checklist (architecture.md) — action constant, activity registry, intent finder, handler, order execution wiring
- Order execution pattern — single work unit, clear intent, re-evaluate
- **Consider Extensibility** — extraction pattern accommodates future targets (pigment, sap, essence)
- **Anchor to Intent** — "character gets seeds from a living plant without killing it"

**Serialization:** No new entity fields beyond the activity/order entries.

[TEST] Extract order appears in orders panel when known. Character walks to target plant, extracts seed, plant remains alive. Seeds inherit parent variety. Order continues until no more target plants or inventory full. Flower extraction produces flower seeds; grass extraction produces grass seeds.

[DOCS]

[RETRO]

---

### Step 3: Clay Terrain + Gather Clay

**Anchor story:** The player notices reddish-brown patches near the pond — clay deposits. They create a Gather > Clay order. A character walks to the clay, scoops some up, and carries a lump of clay. Clay deposits never run out.

**What's new:**
- Clay terrain type on Map (`clay map[Position]bool`)
- Clay tile generation during world creation (6-10 tiles, adjacent to water, clustered)
- Clay tile rendering (reddish brown pattern)
- Clay as a gatherable item type (new item constructor)
- Gather order extended to support clay (infinite source, like drinking from water)
- `IsClay(pos)` query on Map

**Architecture patterns:**
- Water terrain pattern (architecture.md: World & Terrain) — infinite source, map terrain, generated at world creation
- Gather order pattern (order_execution.go) — extends existing gather with a new target type
- **Follow the Existing Shape** — clay follows water's infinite-source pattern
- **Isomorphism** — clay is terrain state, stored as terrain (like water, tilled soil)

**Key difference from other gathered items:** Clay comes from terrain, not from item entities on the ground. Gathering clay creates a new item from nothing (like filling a vessel creates water). The gather handler needs a terrain-source path alongside the existing ground-item path.

**Serialization:** Clay tile positions must be saved and restored. Follow the tilled soil serialization pattern.

[TEST] Clay tiles visible near water on world generation. Clay tiles are passable. Characters gather clay via order (walk to clay tile, pick up clay item). Clay tile is not depleted. Clay items can be carried. Save/load preserves clay terrain.

[DOCS]

[RETRO]

---

### Step 4: Craft Bricks

**Anchor story:** A character picks up some clay and has a moment of insight — they could shape this into bricks! The player creates a Craft > Bricks order. The character takes clay from their inventory and forms it into a brick.

**What's new:**
- Brick item type (new item constructor)
- "craftBrick" recipe in RecipeRegistry (input: clay, output: brick)
- "craftBrick" activity in ActivityRegistry with discovery triggers (picking up or looking at clay)
- Brick creation function in crafting.go

**Architecture patterns:**
- Adding a New Recipe checklist (architecture.md: Recipe System)
- Adding a New Activity checklist (architecture.md: Activity Registry)
- Crafting flow: order → check recipe feasibility → gather inputs → craft → output
- **Follow the Existing Shape** — mirrors craftVessel/craftHoe patterns exactly

**Serialization:** New recipe and activity IDs must be recognized on load. Brick items need save/load support (follows existing item patterns).

[TEST] Looking at or picking up clay can trigger craftBrick know-how discovery. Craft > Bricks order appears when known. Character gathers clay, crafts brick. Brick item appears on ground after crafting. Save/load preserves bricks and know-how.

[DOCS]

[RETRO]

---

### Step 5: Construct Entity + Build Fence

**Anchor story:** A character who has been gathering sticks discovers they can build a fence. The player creates a Construction > Build Fence order (selecting stick fence). The character gathers a full bundle of 6 sticks, walks to an open tile, and builds a fence segment. The fence blocks movement — characters path around it. Another character looks at the fence and decides they like the look of stick construction.

**What's new:**
- `Construct` entity type with ConstructType/Kind/Material/Color/Passable/Movable fields
- Map integration: construct storage, AddConstruct, RemoveConstruct, ConstructAt, IsBlocked integration
- Construct rendering in game view
- Construct save/load serialization
- `ActionBuildFence` action and "buildFence" activity
- Three fence recipes: thatch (1 grass bundle of 6), stick (1 stick bundle of 6), brick (6 bricks)
- Build fence intent finder with material procurement (EnsureHasBundle or equivalent)
- Build fence handler: walk to target tile, build construct
- "Construction" activity category in order UI
- Discovery triggers for buildFence know-how
- Looking at constructs + material preference formation

**Architecture patterns:**
- Adding an Ordered Action checklist — with new entity type integration
- New entity type addition: Map storage, rendering, save/load, pathfinding integration
- Recipe/procurement pattern — gather materials, then build
- Area Selection reuse potential (single tile for fences, but area for future huts)
- **Isomorphism** — Construct is a distinct world concept, gets a distinct entity type
- **Consider Extensibility** — ConstructType/Kind hierarchy accommodates future furniture
- **Follow the Existing Shape** — ItemType/Kind hierarchy pattern from Items

**Serialization:** Full Construct entity serialization. New save state types for constructs.

**Evaluate triggered enhancements:**
- `continueIntent` consolidation — evaluate whether fence building adds a multi-phase action that needs an early-return block
- Order-aware simulation tests — add for construction orders

[TEST] Construct entity renders on map. Fences block movement (characters path around). Three fence material types buildable. Character procures full bundle/bricks before building. Build fence order works end to end. Characters can look at constructs. Material preference formation works (like/dislike thatch, sticks, bricks). Save/load preserves constructs.

[DOCS]

[RETRO]

---

### Step 6: Build Hut

**Anchor story:** The player selects Construction > Build Hut > Thatch Hut and enters area selection mode. They position a 4x4 footprint and confirm. Marked tiles appear. A character begins the long process of building — dropping bundles of grass on each of the 16 tiles, then working each tile one by one. Eventually a hut stands: a 4x4 outline of thatch walls with a 3x3 interior space accessible through a single door tile. Characters can walk through the door but not through walls.

**What's new:**
- Build Hut activity with three recipes (thatch: 2 grass bundles per tile x 16, stick: 2 stick bundles per tile x 16, brick: 12 bricks per tile x 16)
- Area selection for hut placement (fixed 4x4 footprint, reusing tilling's area selection UI)
- Hut validator (clear 4x4 space, no water/existing constructs)
- Marked-for-construction pool (parallel to marked-for-tilling)
- Multi-phase build per tile: supply dropping phase, then working phase
- Door tile: one outline tile that's character-passable
- Order becomes unfulfillable when materials run out, re-fulfillable when available again

**Architecture patterns:**
- Area Selection UI Pattern (architecture.md) — reuse with hut-specific validator and fixed size
- Marked-for-Tilling Pool pattern — adapted for construction (marked-for-building)
- Order execution with supply management — new complexity: per-tile material requirements
- **Follow the Existing Shape** — tilling area selection is the direct precedent
- **Anchor to Intent** — "player places a hut blueprint, characters build it over time"

**Serialization:** Marked-for-building positions. Partially-built hut state (which tiles are complete). Door position.

[TEST] Area selection works for hut placement (fixed 4x4). Hut validator rejects invalid positions (water, existing constructs). Workers drop supplies on tiles then build each one. Completed hut has walls (impassable) and door (character-passable). Order pauses when materials run out, resumes when available. Multiple workers can work on the same hut. Save/load preserves partial and complete huts.

[DOCS]

[RETRO]

---

### Step 7: Activity Preferences

**Anchor story:** A character finishes a gardening order while in a good mood and develops a fondness for gardening. The next time they work a garden order, their mood improves slightly faster. Another character who disliked their construction work finds their mood dipping while building. The player opens the preferences panel and sees a new section: "Activity Preferences" showing which work each character enjoys or dislikes.

**What's new:**
- `ActivityPreference` struct (separate from item Preference — target is category string, not item attributes)
- Storage on Character: `ActivityPreferences []ActivityPreference`
- Formation: chance on order completion based on mood (Joyful/Happy → like, Unhappy/Miserable → dislike)
- Effect: mood change rate modifier during ordered work
- UI: new section in preferences panel displaying activity preferences
- Categories: Garden, Harvest, Craft, Construction (and future categories as added)

**Architecture patterns:**
- Preference formation pattern (system/preference.go) — same trigger conditions as item preferences
- Mood system extension (system/survival.go) — activity-based mood rate modifier during ordered work
- UI panel extension — new section in preferences display
- **Consider Extensibility** — separate struct accommodates future extensions (affects order acceptance, personality emergence)
- **Source of Truth Clarity** — activity preferences are a distinct concept from item preferences, get a distinct type

**Serialization:** ActivityPreference on Character must round-trip. New save state types.

[TEST] Completing an order while happy can form a positive activity preference. Completing while unhappy can form negative. Activity preference affects mood rate during that category of work. Preferences panel shows activity preferences as a separate section. Duplicate/opposite preference handling matches item preference rules. Save/load preserves activity preferences.

[DOCS]

[RETRO]

---

## Step 8: Phase Wrap-Up

- [DOCS] Final documentation pass — update CLAUDE.md roadmap, game-mechanics.md with all new systems, architecture.md with Construct entity and new patterns
- [RETRO] Run `/retro` on full construction phase
- Update triggered-enhancements.md — mark resolved triggers, add new deferred items discovered during phase
- Update randomideas.md — remove completed items, add new observations
