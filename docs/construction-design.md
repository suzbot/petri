# Construction Phase Design

Requirements: [Construction-Reqs.txt](Construction-Reqs.txt)

## Overview

Characters gather materials and construct small buildings from grass, sticks, or bricks. Construction introduces bundles (stackable materials), a new plant type (tall grass), clay terrain, brick crafting, a new entity type (Construct) for structures and future furniture, fence and hut building, and two new preference systems (material preferences and activity preferences). This sets the stage for future phases where buildings protect stored items and provide better sleep.

---

## Steps

### Step 1: Tall Grass + Bundles
**Status:** Complete

**Anchor story:** The player sees tall, pale green grass growing and spreading across the world. A character harvests some grass — the living green plant becomes pale yellow dried material in their inventory, showing "bundle of tall grass (1)." They harvest another nearby — the bundle grows to 2. Sticks now also stack as bundles when picked up. The player notices neither grass nor sticks can be placed in vessels — they're too large. When a character fills a bundle to 6, they drop it on the ground and complete the order.

**Scope:**
- BundleCount field on Item, MaxBundleSize config, VesselExcludedTypes config
- Modified Pickup with bundle merge path, CanPickUpMore bundle check
- Tall grass plant type (ItemType "grass", Kind "tall grass", symbol W, pale green, fast lifecycle)
- Harvest UI shows all growing non-sprout plants (not just edible)
- Harvest and gather orders integrate with bundle mechanics
- Harvested grass color change (pale green → pale yellow)
- Bundle rendering (X symbol for count >= 2), details panel bundle display
- RequiresHarvestable discovery trigger for harvest know-how

---

### Step 2: Seed Extraction
**Status:** In Progress (Step 2b remaining)

**Anchor story:** The player creates an Extract > Flower Seeds order. A character who has discovered how to extract procures a vessel, walks to a growing flower, spends time carefully collecting seeds, and obtains a flower seed in their vessel — without harming the flower. They continue to the next flower. The flower can't be extracted from again until its next reproduction cycle. The same mechanic works for Extract > Grass Seeds with tall grass. Seeds can be planted in tilled soil.

**Scope:**
- Seed variety registration (flower seeds, grass seeds) with Plantable=true
- SeedTimer on PlantProperties (cooldown after extraction, tied to reproduction interval)
- ActionExtract constant, extract activity in ActivityRegistry
- Extract order UI with target type selection (Flower Seeds, Grass Seeds)
- findExtractIntent with vessel procurement
- applyExtract handler (walk-then-act, creates seed, routes to vessel/inventory)
- Discovery triggers (looking at flowers/grass, encountering seeds)
- Planting verification for extracted seeds
- "Gone to seed" indicator on details panel
- Save/load for SeedTimer

**Open questions:**
- None remaining (all resolved during Step 2a refinement)

---

### Step 3: Clay Terrain + Gather Clay
**Status:** Planned

**Anchor story:** The player notices reddish-brown patches near the pond — clay deposits. They create a Gather > Clay order. A character walks to the clay, scoops some up, and carries a lump of clay. Clay deposits never run out.

**Scope:**
- Clay terrain type on Map (clay map, IsClay query, clustered near water)
- Clay tile generation during world creation (6-10 tiles, adjacent to water)
- Clay tile rendering (reddish brown)
- Clay as a gatherable item type (new item constructor)
- Gather order extended for terrain-source items (infinite, like drinking from water)
- Clay terrain serialization (follows tilled soil pattern)

**Open questions:**
- Clay item properties: does clay have varieties/colors, or is it uniform? Reqs don't specify.
- Gather-from-terrain mechanics: gather currently targets ground items. Clay creates an item from terrain (like filling a vessel creates water). How does the gather handler dispatch between ground-item and terrain-source paths?

---

### Step 4: Craft Bricks
**Status:** Planned

**Anchor story:** A character picks up some clay and has a moment of insight — they could shape this into bricks! The player creates a Craft > Bricks order. The character takes clay from their inventory and forms it into a brick.

**Scope:**
- Brick item type (new item constructor)
- craftBrick recipe in RecipeRegistry (input: clay, output: brick)
- craftBrick activity in ActivityRegistry with discovery triggers (picking up or looking at clay)
- Brick creation function in crafting.go

---

### Step 5: Construct Entity + Build Fence
**Status:** Planned

**Anchor story:** A character who has been gathering sticks discovers they can build a fence. The player creates a Construction > Build Fence order (selecting stick fence). The character gathers a full bundle of 6 sticks, walks to an open tile, and builds a fence segment. The fence blocks movement — characters path around it. Another character looks at the fence and decides they like the look of stick construction.

**Scope:**
- Construct entity type with ConstructType/Kind/Material/Color/Passable/Movable fields
- Map integration (construct storage, AddConstruct, RemoveConstruct, ConstructAt, IsBlocked)
- Construct rendering and save/load serialization
- ActionBuildFence action, buildFence activity, "Construction" order UI category
- Three fence recipes: thatch (1 grass bundle of 6), stick (1 stick bundle of 6), brick (6 bricks)
- Build fence intent finder with material procurement (EnsureHasBundle or equivalent)
- Discovery triggers for buildFence know-how
- Looking at constructs + material preference formation

**Open questions:**
- ~~EnsureHasBundle helper design~~ → DD-7
- How does a construction worker procure a full bundle of 6? Does EnsureHasBundle gather one at a time, or is there a "fill bundle" loop?
- Door passability mechanism: currently passability is boolean. Doors need "character-passable but creature-impassable" (for future Threats phase). May need a passability enum or just Passable=true until creatures exist.

**Triggered enhancements:**
- `continueIntent` early-return block consolidation — already at 6 blocks, trigger threshold met. Evaluate whether fence building adds a multi-phase action that needs an early-return block.
- Order-aware simulation for e2e testing — construction adds multiple new ordered actions with supply procurement.
- Category type formalization — evaluate whether `VesselExcludedTypes` set is sufficient or whether formal item categories are needed.

---

### Step 6: Build Hut
**Status:** Planned

**Anchor story:** The player selects Construction > Build Hut > Thatch Hut and enters area selection mode. They position a 4x4 footprint and confirm. Marked tiles appear. A character begins the long process of building — dropping bundles of grass on each of the 16 tiles, then working each tile one by one. Eventually a hut stands: a 4x4 outline of thatch walls with a 3x3 interior space accessible through a single door tile. Characters can walk through the door but not through walls.

**Scope:**
- Build Hut activity with three recipes (thatch, stick, brick — per-tile material costs)
- Area selection for hut placement (fixed 4x4 footprint, reusing tilling's area selection UI)
- Hut validator (clear 4x4 space, no water/existing constructs)
- Marked-for-construction pool (parallel to marked-for-tilling)
- Multi-phase build per tile: supply dropping phase, then working phase
- Door tile: one outline tile that's character-passable
- Order becomes unfulfillable when materials run out, re-fulfillable when available again

**Open questions:**
- Hut supply management: how does the worker distribute supplies across tiles? Per-tile drops then per-tile work, or all supplies then all work?
- Multiple hut materials: can a single hut mix materials (some thatch, some brick)? Reqs imply one material per hut.

---

### Step 7: Activity Preferences
**Status:** Planned

**Anchor story:** A character finishes a gardening order while in a good mood and develops a fondness for gardening. The next time they work a garden order, their mood improves slightly faster. Another character who disliked their construction work finds their mood dipping while building. The player opens the preferences panel and sees a new section: "Activity Preferences" showing which work each character enjoys or dislikes.

**Scope:**
- ActivityPreference struct (target is category string, not item attributes)
- Storage on Character: ActivityPreferences
- Formation: chance on order completion based on mood (Joyful/Happy → like, Unhappy/Miserable → dislike)
- Effect: mood change rate modifier during ordered work
- UI: new section in preferences panel
- Categories: Garden, Harvest, Craft, Construction

---

### Step 8: Phase Wrap-Up
**Status:** Planned

**Scope:**
- Final documentation pass (CLAUDE.md roadmap, game-mechanics.md, architecture.md with Construct entity)
- Full-phase retro
- Update triggered-enhancements.md and randomideas.md
- Tuning: extraction duration and seed yield (ActionDurationShort may be too fast; tune duration and yield together)
- Evaluate: preference-weighted target selection → unified item-seeking in picking.go (moved from triggered-enhancements.md — construction's multiple craft recipes with varied inputs meets the trigger)

---

## Design Decisions

### DD-1: Bundles use BundleCount field on Item
**Context:** Bundleable items (sticks, grass) need to stack. Options: new BundleItem struct, separate inventory slot type, or a field on Item.
**Decision:** `BundleCount int` on the Item struct. `MaxBundleSize` config per item type (sticks: 6, grass: 6). Single items start at BundleCount=1.
**Rationale:** Simplest representation — one field, no new structs. Follows **Start With the Simpler Rule**.
**Affects:** Step 1

### DD-2: Vessel exclusion uses explicit config set
**Context:** Sticks implicitly can't go in vessels (no variety, so AddToVessel fails on registry lookup). Grass needs registered varieties (for seed inheritance) but still can't go in vessels — the implicit check would incorrectly allow grass.
**Decision:** `VesselExcludedTypes` set in config (sticks, grass). Check in `AddToVessel`, `CanVesselAccept`, and gather/harvest order vessel-procurement branching.
**Rationale:** Making exclusion explicit is necessary for correctness. Follows **Source of Truth Clarity**.
**Affects:** Step 1, Step 2

### DD-3: Seed extraction is a new "Extract" activity
**Context:** Seeds from flowers/grass need a non-destructive collection mechanic. Options: extend foraging, extend harvesting, or new activity.
**Decision:** "Extract" as a new orderable activity. Subcategories: "Flower Seeds", "Grass Seeds". Non-destructive — plant is unharmed.
**Rationale:** Distinct pattern from foraging (food-centric), harvesting (destructive), and gathering (picks up ground items). Follows **Consider Extensibility** — future targets include pigment, sap, essence.
**Affects:** Step 2

### DD-4: Construct is a new entity type
**Context:** Character-built structures need representation. Options: reuse Feature (natural world elements), extend Item, or new entity type.
**Decision:** Separate Construct entity with ConstructType/Kind hierarchy mirroring ItemType/Kind on Items. ConstructType: "structure" (immovable, impassable except doors) and future "furniture" (movable, passability varies).
**Rationale:** Constructs and natural features are different world concepts. Follows **Isomorphism**.
**Affects:** Step 5, Step 6

### DD-5: Clay is a new terrain type
**Context:** Clay deposits as a material source. Options: item entities on the ground (depletable) or terrain (infinite).
**Decision:** `clay map[Position]bool` on Map with `IsClay(pos)` query. Passable tiles, clustered near water. Infinite source like water.
**Rationale:** Follows **Follow the Existing Shape** — water terrain is the precedent for infinite sources.
**Affects:** Step 3

### DD-6: Hut uses fixed 4x4 footprint with area selection
**Context:** Hut placement UI. Options: free-form, fixed size with area selection, tile-by-tile.
**Decision:** Reuse the area selection UI pattern from tilling. Fixed 4x4 outline. Interior is 3x3. One outline tile is a door (character-passable only).
**Rationale:** Follows **Follow the Existing Shape** — tilling area selection is the direct precedent.
**Affects:** Step 6

### DD-7: Grass identity uses Kind="tall grass"
**Context:** Grass items need a natural display name for details panel and preferences. Options: ItemType "tall grass" (breaks ItemType conventions), separate displayName map, or Kind field.
**Decision:** ItemType "grass", Kind "tall grass". Kind enables natural display ("Likes tall grass" vs "Likes grasses") and follows the seed pattern (ItemType "seed", Kind "gourd seed").
**Rationale:** Follows **Isomorphism** — Kind is the item's specific identity within its type. Follows **Follow the Existing Shape** — seed Kind pattern.
**Affects:** Step 1

### DD-8: Harvested grass changes color to pale yellow
**Context:** Living grass and harvested material should be visually distinct.
**Decision:** Harvested grass changes from ColorPaleGreen to ColorPaleYellow on pickup, representing real-world drying. Participates in preference system for free.
**Rationale:** Follows **Isomorphism** — harvested grass IS materially different from living grass, so it looks different.
**Affects:** Step 1

### DD-9: SeedTimer cooldown tied to reproduction interval
**Context:** After extraction, plants need a cooldown before seeds are available again. Options: one-and-done flag, fixed cooldown, or tied to plant lifecycle.
**Decision:** `SeedTimer float64` on PlantProperties. After extraction, resets to plant type's spawn interval from `config.ItemLifecycle`. Decrements every tick. When <= 0, plant can be extracted again.
**Rationale:** Ties seed regeneration speed to the plant's lifecycle — fast-reproducing grass regenerates seeds faster than flowers. Creates a natural cycle. Follows **Follow the Existing Shape** — parallels death timer decrement pattern.
**Affects:** Step 2

### DD-10: Extract uses vessel procurement
**Context:** With only 2 inventory slots, vesselless extraction is nearly useless. Options: no vessel support, optional vessel, required vessel.
**Decision:** Follows harvest vessel procurement pattern. `findExtractIntent` calls `EnsureHasVesselFor` before finding targets. Seeds created by handler and routed to vessel via `AddToVessel`.
**Rationale:** Follows **Follow the Existing Shape** — harvest vessel flow. Follows **Component Procurement** pattern.
**Affects:** Step 2

### DD-11: Harvest target list uses growing non-sprout filter
**Context:** Harvest UI showed only edible items (via `getEdibleItemTypes()`), excluding grass and flowers. Options: config flag for harvestable, or map-based scan.
**Decision:** `getHarvestableItemTypes()` scans map items for distinct types with `Plant.IsGrowing && !Plant.IsSprout`. Naturally includes all plant types.
**Rationale:** Map is the source of truth for what's harvestable, not a config flag. Follows **Source of Truth Clarity**. Follows `GetGatherableTypes()` being map-based.
**Affects:** Step 1

### DD-12: Extract discovery triggers use ItemType-specific pattern
**Context:** How do characters discover they can extract seeds?
**Decision:** ActionLook on flower/grass (seeing a plant with seeds), ActionPickup/ActionLook on seeds (encountering seeds suggests more sources). Uses existing ItemType-specific trigger pattern (same as tillSoil triggers on hoe).
**Rationale:** Follows **Follow the Existing Shape** — ItemType-specific triggers are an established pattern.
**Affects:** Step 2
