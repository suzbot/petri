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

**Anchor story:** The player sees tall, pale green grass growing and spreading across the world. A character harvests some grass — the living green plant becomes pale yellow dried material in their inventory, showing "bundle of tall grass (1)." They harvest another nearby — the bundle grows to 2. Sticks now also stack as bundles when picked up. The player notices neither grass nor sticks can be placed in vessels — they're too large. When a character fills a bundle to 6, they drop it on the ground and complete the order.

**Resolved design decisions (from discussion):**
- Grass seed vs grass item: **Different items.** Grass plant on map = ItemType "grass" with PlantProperties. Harvested grass = "grass" material with BundleCount (construction material). Grass seed = ItemType "seed", Kind "grass seed" (produced by extraction in Step 2, not harvest). Harvesting is destructive (removes plant). Extraction (Step 2) is non-destructive.
- Grass symbol: `W` — visually reads as blades/fronds, distinct from all current symbols.
- Grass spawn count: Same as other plant types (`ItemSpawnCount` = 20), not 2x as originally required. Aggressive reproduction via Fast tier handles proliferation naturally.
- Grass lifecycle: Fast maturation (same as mushroom, ~1 world day), Fast reproduction (same as berry, ~2 world days per plant). Fastest-growing plant overall.
- Grass varieties: Single variety (pale green, no pattern, no texture). Still registered in VarietyRegistry — needed for seed extraction in Step 2.
- Grass identity: ItemType "grass", Kind "tall grass". Kind enables natural display in details panel and preferences ("Likes tall grass" vs "Likes grasses"). Follows the seed pattern (ItemType "seed", Kind "gourd seed").
- Bundle-full behavior: When harvest/gather fills a bundle to max (6), character **drops the bundle** at current position and order **completes**. One bundle per order. Applies to both grass harvest and stick gather.
- Dried grass color: Harvested grass changes from ColorPaleGreen to ColorPaleYellow on pickup, representing real-world drying. Visually distinguishes living plants from harvested material. Participates in preference system — characters can form opinions about pale yellow items. See triggered-enhancements.md for reassessment if warmer gold/wheat colors enter the palette.
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

**Anchor story:** The player opens the orders panel and sees grass in the harvest list alongside berries and mushrooms. They create a Harvest > Grass order. A character walks to tall grass, harvests it into a bundle, moves to the next grass, harvests again (bundle grows to 2, 3...). When the bundle reaches 6, the character drops it on the ground and completes the order. The player creates a Gather > Sticks order — same flow. The character gathers sticks into a bundle of 6, drops it, done.

**Resolved design decisions (from refinement):**
- Harvest target list filter: **Growing non-sprout items**, not edible items. The current `getEdibleItemTypes()` filter excludes grass (non-edible). The harvest list should show all item types that have growing, non-sprout instances — this naturally includes berries, mushrooms, gourds, grass, and flowers. No config flag needed; scan map items for distinct types with `Plant.IsGrowing && !Plant.IsSprout`. Follows `GetGatherableTypes()` being map-based. If a future plant type shouldn't be harvestable, add an exclusion then.
- `applyPickup` PickupToInventory harvest handler: The existing `char.GetCarriedVessel() == nil` check accidentally works correctly for vessel-excluded harvest targets (grass never has a vessel, so the check always passes). This is correct behavior from coincidental logic. Track in triggered-enhancements.md for rework when the handler becomes confusing or a second vessel-excluded harvestable type arrives.

---

#### Step 1c-i: Harvest UI — Show All Harvestable Types ✅

**Anchor story:** The player opens the orders panel, selects Harvest, and sees grass listed alongside berries, mushrooms, gourds, and flowers. They can select grass and create a Harvest > Grass order.

**What's new:**
- New `getHarvestableItemTypes()` function (on Model, in `view.go`): scans `m.gameMap.Items()` for distinct `ItemType` values where `Plant != nil && Plant.IsGrowing && !Plant.IsSprout`. Returns sorted list of type name strings. Follows the same map-based pattern as `GetGatherableTypes()` in `game/world.go`.
- Replace `m.getEdibleItemTypes()` with `m.getHarvestableItemTypes()` in three harvest UI locations:
  - `view.go` ~line 1526: rendering the harvest target selection list
  - `update.go` ~line 1104: creating the order with selected target type (`applyOrdersConfirm`)
  - `update.go` ~line 396: cursor navigation bounds for the target list
- `getEdibleItemTypes()` on Model (view.go) may become unused after this change — remove if so. The package-level `getEdibleItemTypes()` in update.go (used for character food preferences at creation) is a separate function and stays.

**Architecture patterns:**
- **Follow the Existing Shape** — `GetGatherableTypes()` is the sibling pattern: scan map items for distinct types matching a criterion. Harvest version filters on growing non-sprout plants.
- **Source of Truth Clarity** — the map is the source of truth for what's harvestable, not a config flag.

**Reqs reconciliation:**
- "As a plant, should show up on the harvest list" → growing non-sprout filter includes grass (and all other plant types) naturally.

**Tests (TDD):**
- Unit: `getHarvestableItemTypes` returns grass when growing grass exists on map
- Unit: `getHarvestableItemTypes` returns edible plants (berry, mushroom) alongside non-edible (grass, flower)
- Unit: `getHarvestableItemTypes` excludes sprouts
- Unit: `getHarvestableItemTypes` returns empty when no growing plants exist

[TEST] Open orders panel, select Harvest. Verify grass appears in the list. Verify berries/mushrooms/gourds still appear. Verify flowers also appear (growing non-edible plant). Create a Harvest > Flower order and verify it executes without broken loops or spam — flowers go through the vessel path (not vessel-excluded, have varieties), so a character should procure a vessel, pick flowers into it, and complete normally. This is the first time flower harvest has been exercisable. **Do not test Harvest > Grass execution yet** — the bundle integration isn't wired until Step 1c-ii. Creating a grass harvest order at this point would produce a one-at-a-time loop that stalls at bundle capacity.

---

#### Step 1c-ii: Harvest + Gather Bundle Integration ✅

**Anchor story:** A character with a Harvest > Grass order walks to tall grass, harvests it (the plant is destroyed, grass appears in inventory as "bundle of grass (1)"). They walk to the next grass, harvest again — the bundle grows to (2). At (6), they drop the full bundle on the ground and the order completes. A Gather > Sticks order works the same way. Meanwhile, a Harvest > Berry order still works as before — character gets a vessel, fills it with berries.

**What's new:**

**`findHarvestIntent` changes** (`order_execution.go`):
- Add full-bundle safety check at start: if target is bundleable and `hasFullBundle(char, order.TargetType)`, return nil. Parallel to the same check in `findGatherIntent`. Safety net — normally the previous tick's handler drops the bundle and completes, but this catches edge cases where intent was cleared mid-flow.
- After finding target via `findNearestItemByType`: check `config.IsVesselExcluded(target.ItemType)`. If true, skip `EnsureHasVesselFor` entirely and go straight to move-and-pickup. If false, proceed with existing vessel procurement path.
- No changes to the vessel path (berries, mushrooms, gourds) — those continue working exactly as before.

**`findGatherIntent` changes** (`order_execution.go`):
- Add explicit vessel-excluded check **before** the variety registry check. Currently, variety presence is a proxy for "needs vessel" — sticks have no variety so they skip vessels. But grass has a variety (needed for seed extraction in Step 2) yet is vessel-excluded. The fix: after finding target, check `config.IsVesselExcluded(target.ItemType)` first. If true, take the direct (bundle) path regardless of variety presence. If false, proceed to existing variety check for vessel procurement.
- This makes the bundle path explicit rather than relying on variety-absence as a proxy.

**`applyPickup` PickupToBundle handler changes** (`apply_actions.go`):
- Currently (from Step 1a), the PickupToBundle handler only processes gather orders. Extend to also handle harvest orders.
- For harvest orders: continuation uses `FindNextHarvestTarget` (growingOnly=true). For gather orders: continuation uses `FindNextGatherTarget` (unchanged).
- On no-more-targets or bundle full: call `DropCompletedBundle` + `CompleteOrder` for both harvest and gather.
- Structure: the PickupToBundle block checks `order.ActivityID` and dispatches to the appropriate continuation function. Follows the same shape as the PickupToVessel block which already handles both harvest and gather contexts.

**`FindNextHarvestTarget` changes** (`order_execution.go`):
- Make bundle-aware: for bundleable target types, use `CanPickUpMore(char, targetType)` as the capacity check instead of only `HasInventorySpace()`. A character with a non-full bundle can pick up more even if inventory is "full."
- Return nil when bundle is full (same as returning nil when inventory is full for vessel-based harvest).

**`applyPickup` PickupToInventory harvest path** (`apply_actions.go`):
- No changes needed. The existing check `order.ActivityID == "harvest" && char.GetCarriedVessel() == nil` works correctly for grass: vessel is always nil for vessel-excluded types, so the first grass pickup (PickupToInventory as a new bundle of 1) correctly triggers continuation via `FindNextHarvestTarget`. This is accidental correctness — the check was designed for "did I just pick up a vessel prerequisite?" but it correctly handles "vessel-excluded type never has a vessel." Tracked in triggered-enhancements.md for future rework.

**Triggered enhancement entry** (`triggered-enhancements.md`):
- Add entry: "Harvest PickupToInventory handler clarity" — the `GetCarriedVessel() == nil` check in the harvest PickupToInventory path works for vessel-excluded types by coincidence. Trigger: when a second vessel-excluded harvestable type is added, or when the handler's intent becomes confusing during debugging.

**Architecture patterns:**
- **Follow the Existing Shape** — PickupToBundle harvest handling mirrors PickupToBundle gather handling. findHarvestIntent's vessel-excluded branch mirrors findGatherIntent's.
- **Source of Truth Clarity** — `config.IsVesselExcluded()` is the single check for vessel exclusion, replacing variety-absence as a proxy in findGatherIntent.
- **Anti-pattern to avoid:** Don't create a separate "bundle harvest" action type. Use the same `ActionPickup` action — the bundle mechanics are in the `Pickup()` function, not the action type. This follows **Separation of Concerns by Cognitive Role** — what changes is the physical storage, not the intent.
- No `continueIntent` early-return block needed — harvest/gather targets are always growing plants or ground items on the map, never inventory items. The generic path handles path recalculation.

**Reqs reconciliation:**
- "As a plant, should show up on the harvest list" → Step 1c-i handles UI. This step handles the execution path.
- "Harvested grass CANNOT be placed in a vessel, but can be carried or dropped as a stack" → `config.IsVesselExcluded` skips vessel procurement in findHarvestIntent; bundle merge in `Pickup()` handles stacking.
- "A bundle of sticks can have up to 6 pieces" + "a picked up stack can be added to until max pieces" → harvest/gather continues until `CanPickUpMore` returns false (bundle at max).
- Drop-on-complete → `DropCompletedBundle` + `CompleteOrder` for both harvest bundles and gather bundles.

**Tests (TDD):**
- Unit: `findHarvestIntent` skips vessel procurement for vessel-excluded types (grass)
- Unit: `findHarvestIntent` returns nil when character has full bundle of target type
- Unit: `findHarvestIntent` still uses vessel procurement for non-excluded types (berry)
- Unit: `findGatherIntent` takes bundle path for vessel-excluded types with varieties (grass)
- Unit: `FindNextHarvestTarget` returns intent when bundle has room (CanPickUpMore true)
- Unit: `FindNextHarvestTarget` returns nil when bundle full
- Integration: harvest grass end-to-end — character harvests grass into bundle, continues until full (6), drops bundle, order completes
- Integration: gather sticks end-to-end — same flow with sticks (already working from Step 1a, but verify regression-free)
- Integration: harvest berry end-to-end — character still uses vessel path, unchanged behavior

[TEST] Create harvest grass and gather sticks orders. Watch characters collect into bundles, continuing to next target after each pickup. When bundle hits 6, character drops it and order completes. Verify bundles on ground show correct count. Verify existing harvest (berries, mushrooms) still works with vessel path unchanged.

[DOCS] ✅

[RETRO]

---

#### ✅ Step 1d: Polish, UI, + Save/Load Verification

**Anchor story:** A character harvests tall grass — the pale green living plant becomes pale yellow dried material in their inventory, showing "bundle of tall grass (1)." They harvest more — the count climbs to (2), (3)... When the bundle is dropped on the ground, it renders as a pale yellow `X` — visually distinct from both living grass (`W` in pale green) and a single stick (`/`). The player selects the bundle and the details panel shows "Kind: tall grass" and "Bundle: 6/6" — the same capacity format as vessel contents. They save, reload — everything preserved. Living grass is still pale green, harvested bundles still pale yellow with correct counts.

**Resolved design decisions (from refinement):**
- **Grass Kind: "tall grass".** Set Kind="tall grass" on grass items (ItemType remains "grass"). This gives the details panel a natural display name, produces better preference language ("Likes tall grass" vs "Likes grasses"), and follows the existing Kind pattern (seeds use Kind "gourd seed" on ItemType "seed"). The variety registration also gets Kind="tall grass" to support vessel restoration in future extensions.
- **Dried grass color: ColorPaleYellow.** When grass is harvested (transitions from growing plant to material), its color changes from ColorPaleGreen to ColorPaleYellow. Represents real-world drying. Participates in existing preference system for free — characters looking at dried grass bundles can form opinions about pale yellow and/or tall grass. See triggered-enhancements.md for reassessment if warmer gold/wheat colors enter the palette.
- **Bundle description at count 1: show for harvested/non-growing items.** The current `BundleCount >= 2` threshold in `Description()` means a single harvested grass shows no bundle indicator. Fix: show bundle format for `BundleCount >= 1` when the item is not a growing plant (`Plant == nil || !Plant.IsGrowing`). Growing plants keep their normal description. Affects both grass and sticks — a freshly picked up stick shows "bundle of sticks (1)."
- **Details panel "Kind" line fix.** The details panel (view.go:864) currently shows `item.ItemType` for all items, even those with Kind set — the label says "Kind" but displays ItemType. Fix: prefer `item.Kind` when present, fall back to `item.ItemType`. This benefits all items with Kind: seeds show "Kind: gourd seed" instead of "Kind: seed", water shows "Kind: water" instead of "Kind: liquid". Straightforward correctness fix.

---

#### ✅ Step 1d-i: Grass Kind + Dried Color + Harvest Helper

**Anchor story:** A character harvests tall grass. In their inventory, the material appears as pale yellow "bundle of tall grass (1)" — clearly different from the pale green living plants on the map. They can tell what they're carrying and that it's dried material.

**What's new:**

- **Kind="tall grass" on grass items** (`entity/item.go`): Add `Kind: "tall grass"` to `NewGrass()` constructor. `Description()` already prioritizes Kind over ItemType, so this produces "pale yellow tall grass" (single) and "bundle of tall grass (N)" (bundle) with no further changes needed to Description's non-bundle path.
- **Kind="tall grass" on grass variety** (`game/variety_generation.go`): Set Kind on the grass ItemVariety registration. Needed for variety restoration on extraction (architecture.md: Serialization Checklist) and for correct preference matching via `MatchesVariety`.
- **`Pluralize("tall grass")` case** (`entity/preference.go`): Add `"tall grass": "tall grass"` — grass is an uncountable noun. Without this, preference descriptions would say "tall grasss."
- **Remove `itemDisplayName["grass"]` entry** (`entity/item.go`): The "tall grass" display name was in this map as a workaround. With Kind set, `Description()` uses Kind directly — the map entry becomes dead code.
- **Remove `bundlePluralName["grass"]` entry** (`entity/item.go`): The bundle description path (`BundleCount >= 1`) should use Kind when present, falling back to bundlePluralName. With Kind="tall grass", the plural name comes from `Pluralize("tall grass")` → "tall grass". Update the bundle description logic: if `Kind != ""`, use `Pluralize(Kind)` for the bundle name; else use `bundlePluralName[ItemType]` as before.
- **`harvestItem(item)` helper in `picking.go`:** Extracts the duplicated harvest transition logic from the three `Pickup()` paths (vessel ~line 1053, bundle merge ~line 1091, inventory ~line 1118) into a single function. Each path currently duplicates: `IsGrowing = false`, `SpawnTimer = 0`, `DeathTimer = 0`. The new helper does all three plus applies the grass color change (`if item.ItemType == "grass" { item.Color = types.ColorPaleYellow }`). Called from all three paths where `IsGrowing` was previously set to false. Does NOT replace the `Plantable = true` mutation for berries/mushrooms — that's a separate concern (pickup activation, not harvest transformation).
- **Copy Kind in `spawnItem()`** (`system/lifecycle.go`): Add `Kind: parent.Kind` to the new item in `spawnItem()`, alongside the existing `Color`, `Pattern`, `Texture` copies. Without this, grass plants that reproduce via sprouts would mature with `Kind=""` — after removing `itemDisplayName["grass"]`, they'd display as "pale green grass" instead of "pale green tall grass". Grass is the first reproducing plant with a Kind value (seeds, hoes, liquids all have Kind but don't reproduce via `spawnItem`), so this gap wasn't visible before.

**Architecture patterns:**
- **Follow the Existing Shape** — Kind on grass mirrors Kind on seeds. The three `Pickup()` paths already share identical harvest logic; `harvestItem` makes the shared shape explicit.
- **Isomorphism** — harvested grass IS materially different from living grass (dried vs growing), so it looks different. Kind="tall grass" is the item's specific identity, not a display-only label.
- **Source of Truth Clarity** — the color lives on the item, not derived at render time. Kind is the identity field, not a separate displayName map.

**Reqs reconciliation:**
- "Pale Green" (Construction-Reqs line 35) → applies to the living plant. Harvested material changes to pale yellow. No conflict — the req describes the plant, not the material.
- "A stack of grass is called a bundle" (Construction-Reqs line 40) → `Description()` produces "bundle of tall grass (N)" using Kind.
- Preference participation: characters looking at dropped dried grass bundles can form preferences about ColorPaleYellow and/or Kind "tall grass" — free through existing `CompleteLook` → `TryFormPreference` path.

**Tests (TDD):**
- Unit: `NewGrass()` sets Kind="tall grass"
- Unit: `Description()` for grass with Kind uses "tall grass" (not "grass")
- Unit: `Pluralize("tall grass")` returns "tall grass"
- Unit: `harvestItem` sets `IsGrowing = false`, clears `SpawnTimer`, clears `DeathTimer`
- Unit: `harvestItem` changes grass color to `ColorPaleYellow`
- Unit: `harvestItem` does not change color for non-grass items (sticks, berries)
- Unit: `Pickup` all three paths call `harvestItem` (verify via color change on grass)
- Unit: grass variety has Kind="tall grass"
- Unit: `spawnItem` copies Kind from parent (grass sprout inherits Kind="tall grass")

---

#### ✅ Step 1d-ii: Bundle Description Fix + Details Panel + Map Rendering

**Anchor story:** A character picks up one stick — their inventory shows "bundle of sticks (1)." They pick up one grass — "bundle of tall grass (1)." Meanwhile, growing grass on the map still shows as "pale green tall grass" (no bundle label). The player selects a dropped bundle of 6 and the details panel shows "Kind: tall grass", "Color: pale yellow", and "Bundle: 6/6" — the same capacity format as vessel contents. When a bundle of 2+ items is on the ground, it renders as `X`.

**What's new:**

- **`Description()` bundle threshold fix** (`entity/item.go`): Change the bundle description condition from `BundleCount >= 2` to `BundleCount >= 1` with a growing-plant guard: `BundleCount >= 1 && (i.Plant == nil || !i.Plant.IsGrowing)`. Growing plants with BundleCount=1 (their constructor-set default) keep their normal descriptive name. Harvested materials and non-plant bundleables (sticks) show bundle format at any count.
- **Details panel "Kind" line fix** (`ui/view.go` ~line 864): Change `kindLabel := item.ItemType` to prefer `item.Kind` when present: `kindLabel := item.ItemType; if item.Kind != "" { kindLabel = item.Kind }`. The sprout suffix logic stays. This fixes all items with Kind — seeds, crafted items, liquids, and now grass.
- **Details panel bundle count line** (`ui/view.go` after color line ~870): When `item.BundleCount > 0 && (item.Plant == nil || !item.Plant.IsGrowing)`, add a line: `Bundle: N/M` where N is BundleCount and M is `config.MaxBundleSize[item.ItemType]`. Follows the vessel contents display pattern (`Contents: red berries: 5/20`). Not shown for growing plants.
- **Bundle rendering on map** (`ui/view.go` in `renderCell()`): Items with `BundleCount >= 2` render as `X` instead of their normal symbol. Single items (BundleCount=1) render with their normal symbol (`/` for sticks, `W` for grass). Add `CharBundle = 'X'` to config symbols. Follows the sprout symbol override precedent.
- **Inventory panel verification:** Verify bundle items show their bundle description in the inventory list (`I` key). Same source — `Description()`. Confirm the inventory rendering path handles it.

**Architecture patterns:**
- **Source of Truth Clarity** — `Description()` is the single source of item identity text for inventory and action log. Details panel shows structured fields (Kind, Color, Bundle) for inspection. Both draw from the same item fields — no separate display logic.
- **Follow the Existing Shape** — bundle count display mirrors vessel contents display. Symbol override mirrors sprout symbol override. Kind line fix is a correctness improvement to an existing pattern.
- **Anti-pattern to avoid:** Don't add a separate `BundleDescription()` method. `Description()` handles all contexts — map tooltips, inventory, action log. The details panel shows structured fields, not a single description string.

**Reqs reconciliation:**
- "A stack of sticks is called a bundle" (Construction-Reqs line 29) → `Description()` shows "bundle of sticks (N)" at all counts for picked-up items.
- "A stack of grass is called a bundle" (Construction-Reqs line 40) → `Description()` shows "bundle of tall grass (N)" at all counts for harvested grass.
- "A bundle can have up to 6 pieces" (Construction-Reqs lines 31, 41) → Details panel shows "Bundle: N/6" capacity.

**Tests (TDD):**
- Unit: `Description()` returns "bundle of sticks (1)" for stick with BundleCount=1
- Unit: `Description()` returns "bundle of tall grass (1)" for harvested grass (IsGrowing=false) with BundleCount=1
- Unit: `Description()` returns normal description for growing grass with BundleCount=1 (NOT "bundle of...")
- Unit: `Description()` returns "bundle of tall grass (3)" for BundleCount=3
- Unit: details panel shows `item.Kind` when Kind is set (grass, seeds)
- Unit: details panel shows `item.ItemType` when Kind is empty (berries, mushrooms)
- Unit: details panel shows "Bundle: 6/6" for full bundle
- Unit: details panel does not show Bundle line for growing plants
- Unit: item with BundleCount >= 2 renders as 'X' on map
- Unit: item with BundleCount == 1 renders as normal symbol

---

#### ✅ Step 1d-iii: Harvest Discovery Triggers

**Anchor story:** A character picks up some tall grass and has a realization — they could harvest plants! Previously, only encountering edible items triggered this insight. Now any harvestable plant (grass, flowers) can teach a character about harvesting.

**What's new:**
- **`RequiresHarvestable bool` on `DiscoveryTrigger`** (`entity/activity.go`): New trigger gate, parallel to `RequiresEdible` and `RequiresPlantable`. Matching logic in `triggerMatches()`: checks `item.Plant != nil && item.Plant.IsGrowing && !item.Plant.IsSprout`. Any growing non-sprout plant counts as harvestable.
- **Update harvest activity triggers:** Replace `RequiresEdible` with `RequiresHarvestable` on the harvest activity's `ActionPickup` and `ActionLook` discovery triggers. `ActionConsume` triggers keep `RequiresEdible` — eating is inherently about edible items.
- This enables grass and flowers (non-edible plants) to trigger harvest know-how discovery on pickup or looking.

**Architecture patterns:**
- **Follow the Existing Shape** — `RequiresHarvestable` parallels `RequiresEdible` and `RequiresPlantable` in DiscoveryTrigger. Same bool gate pattern, same evaluation point in `triggerMatches()`.
- **Source of Truth Clarity** — harvestability is derived from plant state (`IsGrowing && !IsSprout`), not a config flag. Matches the same logic used by `getHarvestableItemTypes()` in Step 1c-i.

**Reqs reconciliation:**
- "As a plant, should show up on the harvest list" (Construction-Reqs line 38) → Step 1c-i handled UI listing. This step ensures characters can also *discover* harvest know-how from grass, completing the integration.

**Tests (TDD):**
- Unit: `triggerMatches` returns true for `RequiresHarvestable` with growing non-sprout plant
- Unit: `triggerMatches` returns false for `RequiresHarvestable` with sprout
- Unit: `triggerMatches` returns false for `RequiresHarvestable` with non-plant item
- Unit: picking up grass triggers harvest know-how discovery (when chance succeeds)
- Unit: looking at grass triggers harvest know-how discovery
- Unit: eating still uses `RequiresEdible` (eating grass doesn't trigger harvest — grass isn't edible)

---

#### ✅ Step 1d-iv: Save/Load Verification + Backward Compatibility

**Anchor story:** The player saves a world with living grass, harvested grass bundles (pale yellow), and stick bundles. They reload — living grass is still pale green and growing, harvested bundles are still pale yellow with correct counts, Kind="tall grass" preserved. They load a pre-bundle save — sticks work correctly.

**What's new:**
- **Save/load round-trip verification:** BundleCount on items (inventory and ground), grass plant type (growing, sprouts), grass Kind ("tall grass"), grass variety in registry (with Kind), harvested grass color (ColorPaleYellow). Kind and Color are already-serialized fields — verify they round-trip correctly with the new values.
- **Backward compatibility:** Loading a pre-bundle save where sticks have no BundleCount field. Sticks with BundleCount=0 are migrated to BundleCount=1 on load. Loading a pre-Kind save where grass has no Kind — grass items without Kind should get Kind="tall grass" on load (migration in deserialize, same pattern as BundleCount migration). Verify `NewStick()` sets BundleCount=1 for newly ground-spawned sticks.

**Architecture patterns:**
- **Serialization Checklist** (architecture.md) — verify all new/changed fields round-trip. Color change and Kind addition don't add new struct fields — they change existing field values, which serialize normally. Migration handles old saves.

**Tests (TDD):**
- Round-trip test: save world with grass (growing + harvested) + bundles → load → verify all fields preserved, including harvested grass color=ColorPaleYellow and Kind="tall grass"
- Migration test: load old save without BundleCount → sticks get BundleCount=1
- Migration test: load old save without Kind on grass → grass gets Kind="tall grass"

[TEST] Verify full visual flow: harvest grass → pale yellow in inventory with "bundle of tall grass (1)." Harvest more → count increments. Drop bundle → pale yellow `X` on map, distinct from pale green living `W`. Select dropped bundle → details panel shows "Kind: tall grass", "Color: pale yellow", "Bundle: 6/6". Check inventory panel shows bundle description with count. Verify sticks also show "bundle of sticks (1)" at count 1. Select a gourd seed → details panel shows "Kind: gourd seed" (not "Kind: seed") — verify the Kind line fix works for other items too. Save and reload — living grass still pale green, harvested bundles still pale yellow with correct counts. Load an old save — verify sticks and grass migrate correctly.

[DOCS] ✅

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
- **Post-phase review: code touchpoints for new item types.** Adding grass as a harvestable/bundleable type required changes across many files (config, entity, system, UI rendering, serialization, lifecycle, order execution, apply_actions). Sticks had the same pattern. Investigate whether a checklist, registry, or structural change could reduce the number of touchpoints and lower the risk of missing one when adding future item types. Include in the review: preference formation and mood participation should be part of any new-item-type checklist — any entity that characters can look at or interact with should participate in the preference/mood system (appropriate descriptive attributes, lookable, correct Kind for natural preference language).
