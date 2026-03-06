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
**Status:** Complete

**Anchor story:** The player creates an Extract > Flower Seeds order. A character who has discovered how to extract procures a vessel, walks to a growing flower, spends time carefully collecting seeds, and obtains a flower seed in their vessel — without harming the flower. They continue to the next flower. The flower can't be extracted from again until its next reproduction cycle. The same mechanic works for Extract > Grass Seeds with tall grass. Seeds can be planted in tilled soil.

**Scope:**
- Seed variety registration (flower seeds, grass seeds) with Plantable=true
- SeedTimer on PlantProperties (cooldown after extraction, tied to reproduction interval)
- ActionExtract constant, extract activity in ActivityRegistry
- Extract order UI with target type selection (Flower Seeds, Grass Seeds)
- findExtractIntent with vessel procurement
- applyExtract handler (walk-then-act, creates seed, routes to vessel/inventory)
- Discovery triggers (looking at flowers/grass, encountering seeds)
- SourceVarietyID on Item and ItemVariety — seeds carry parent plant's variety ID (DD-13)
- Seed Kind uses parent Kind ("tall grass seed" not "grass seed") (DD-13)
- CreateSprout uses variety registry lookup instead of string derivation (DD-13)
- PlantableItemExists fix for vessel-stored seeds
- Planting verification for extracted seeds
- "Gone to seed" indicator on details panel
- Save/load for SeedTimer and SourceVarietyID

**Open questions:**
- None remaining (all resolved during Step 2a and 2b refinement)

---

### Step 3: Clay Terrain + Dig Clay
**Status:** Complete

**Anchor story:** The player notices dusky earth patches with a halftone texture near the pond — clay deposits, with a few loose lumps of clay on the surface. A character looks at one and realizes they could dig for more. The player creates a Dig Clay order. The character drops what they're carrying, walks to the clay, scoops some up, and carries a lump of clay. They dig a second lump, then set both down on the ground. Clay deposits never run out.

**Scope:**
- Clay terrain type on Map (clay map, IsClay query, clustered near water)
- Clay tile generation during world creation (6-10 tiles, adjacent to water)
- Clay tile rendering (dusky earth color 138, light shade `░` fill character)
- Clay as a uniform item type (NewClay constructor, no varieties) (DD-14)
- "Dig Clay" as a new top-level ordered action (ActionDig, activity "dig") (DD-15)
- Dig order: drop inventory → dig until both slots full → drop clay on completion (DD-16)
- Dig discovery via loose clay items spawned at world gen (DD-17)
- Drop-on-completion for non-bundled items (separate from DropCompletedBundle) (DD-18)
- Clay terrain serialization (follows tilled soil pattern)

**Open questions:**
- ~~Clay item properties: does clay have varieties/colors, or is it uniform?~~ → DD-14
- ~~Gather-from-terrain mechanics: how does the gather handler dispatch between ground-item and terrain-source paths?~~ → DD-15

---

### Step 4: Craft Bricks
**Status:** Complete

**Anchor story:** A character picks up some clay and has a moment of insight — they could shape this into bricks! The player creates a Craft > Bricks order. The character picks up clay, shapes it into a terracotta brick, sets it down, and goes for the next lump. When no more loose clay remains, the order completes.

**Scope:**
- VesselExcludedTypes config split from MaxBundleSize (DD-20, triggered enhancement resolved)
- Clay + brick added to VesselExcludedTypes
- ColorTerracotta + terracottaStyle (new color)
- Brick item type: `▬` symbol, terracotta color, no Kind, no variety (DD-21)
- craftBrick recipe in RecipeRegistry (input: clay, output: brick, Repeatable: true) (DD-19)
- craftBrick activity in ActivityRegistry with discovery triggers (looking at, picking up, or digging clay)
- CreateBrick function in crafting.go
- applyCraft dispatch case + Repeatable skip for inline CompleteOrder (DD-19)
- isMultiStepOrderComplete case for craftBrick: no loose clay on map (DD-19)
- Save/load: brick symbol restoration

**Open questions:**
- ~~Quantity selection for brick orders~~ → Deferred to triggered enhancements
- ~~Completion condition~~ → DD-19

---

### Step 5: Construct Entity
**Status:** Complete

**Anchor story:** The player creates a test world with a fence already placed on the map. The fence renders as a colored `#` — brown for stick, pale yellow for thatch, terracotta for brick. A character walking toward the fence paths around it — the fence blocks movement. The details panel shows "Stick Fence" with "Structure" and "Not passable."

**Scope:**
- Construct entity type with ConstructType/Kind/Material/MaterialColor/Passable/Movable fields (DD-4)
- Map integration (construct storage, AddConstruct, RemoveConstruct, ConstructAt)
- IsBlocked, MoveCharacter, BFS updated to respect impassable constructs
- Construct rendering (symbol, material color) and details panel
- Save/load serialization
- Test with pre-placed constructs via test-world

**Open questions:**
- Door passability mechanism: currently passability is boolean. Doors need "character-passable but creature-impassable" (for future Threats phase). May need a passability enum or just Passable=true until creatures exist. (Deferred — relevant to Step 8 huts, not Step 5.)

---

### Step 6: Build Fence
**Status:** Refining

**Anchor story:** The player creates a Construction > Fence order. They enter a line-placement mode and mark several contiguous tiles for fence construction. A character takes the order and decides to use sticks. They find a bundle of 6 sticks (open question: gather their own bundle or rely on pre-gathered full bundles), walk to one of the marked tiles, and build. Another character takes a fence construction order and starts building from the other end. The line already has a material choice — sticks — so they build their section out of sticks too. For a brick fence segment, a character makes multiple trips — carrying 2 bricks at a time, dropping them at the build site, and returning for more until 6 are accumulated, then building.

**Scope:**
- buildFence activity in ActivityRegistry, "Construction" order UI category
- Three fence recipes: thatch (1 grass bundle of 6), stick (1 stick bundle of 6), brick (6 bricks) (DD-23)
- Fence placement UI: line/series tile marking (DD-24)
- Character-driven material/recipe selection based on preference and availability — no player sub-menu for material (DD-22)
- Material lock for contiguous fence segments (DD-25)
- Material procurement: bundle gathering for grass/sticks, multi-trip supply-drop for bricks (DD-23)
- Build fence handler with adjacent-tile building (DD-26)
- Order feasibility, completion, multiple workers on same fence line
- Discovery triggers for buildFence know-how

**Open questions:**
- ~~Material sub-menu~~ → DD-22 (character chooses, no player sub-menu)
- ~~Brick bundleability~~ → DD-23 (bricks stay individual, supply-drop pattern)
- Bundle procurement: does the character gather items one at a time into a bundle (like gather orders), or find a pre-gathered full bundle on the ground, or both?
- Line selection UI: what does the placement mode look like? Single-tile 'p' presses, or click-and-drag line drawing? How do non-contiguous fence segments interact with material lock?
- Material lock mechanism: what defines a "line" or "segment"? When does the lock form — first tile built, or at marking time? Storage: data on marked tiles, on order, or derived from placed constructs? What happens when the locked material runs out?
- Build position collision: what happens when another character stands on or walks into the build tile during construction? Options include waiting, displacement, or building-on-tile with post-build displacement.
- Multiple workers: can two workers supply-drop bricks to the same tile, or is each tile claimed by one worker?

**Triggered enhancements:**
- `continueIntent` early-return block consolidation — already at 6 blocks, trigger threshold was met before Step 5 (not worsened by fence building, which uses position-based intent with no new block). Evaluate independently of construction.
- Order-aware simulation for e2e testing — construction adds multiple new ordered actions with supply procurement.
- Category type formalization — evaluate whether `VesselExcludedTypes` set is sufficient or whether formal item categories are needed.

---

### Step 7: Construct Interaction
**Status:** Refining

**Anchor story:** A character wanders past a stick fence and pauses to look at it. They decide they like the look of stick construction — "Likes sticks" appears in their preferences panel. Another character, in a bad mood, looks at a brick fence and develops "Dislikes brick walls."

**Scope:**
- Extend Look idle activity to target constructs (new TargetConstruct on Intent or similar)
- Material preference formation from looking at constructs
- Construct-specific preference attributes (e.g., "Likes brick walls") — extensibility for future attributes
- Details panel shows construct information when selected

**Open questions:**
- How should Look be extended to target constructs? Structural implications: TargetConstruct field on Intent vs alternative approaches. Isomorphism considerations — constructs are not items, should Look treat them differently?
- Preference scope: should characters form preferences about the material only ("Likes sticks"), the construct+material combo ("Likes brick walls"), or both? What about extensibility for future construct attributes (size, craftsmanship, etc.)?

---

### Step 8: Build Hut
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

### Step 9: Activity Preferences
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

### Step 10: Phase Wrap-Up
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

### DD-13: Seeds carry parent variety ID (SourceVarietyID)
**Context:** Seeds need to reconstruct their parent plant when planted. The current approach encodes parent identity in the seed's Kind string ("grass seed" → TrimSuffix → "grass"), which loses the parent's Kind ("tall grass") and relies on brittle string derivation. The seed's Kind also under-specifies identity — "grass seed" when it should be "tall grass seed" (like "turkey egg" vs "bird egg").
**Decision:** Seeds carry `SourceVarietyID string` — the variety registry ID of the parent plant, generated at world creation and shared by all plants of that variety. Added to both `Item` (loose seeds) and `ItemVariety` (vessel-stored seeds). The seed's Kind uses the parent's Kind for specificity: `parentKind + " seed"` when Kind exists, `parentItemType + " seed"` otherwise. At planting time, the parent variety is looked up from the registry; `CreateSprout` receives the resolved variety and creates the sprout with full fidelity — no string derivation.
**Rationale:** The variety registry is the source of truth for plant identity (**Isomorphism**). A single ID reference is lossless and extensible — new variety attributes flow through automatically (**Consider Extensibility**). The seed carrying its parent's variety ID is isomorphic to biological seeds carrying genetic information. Eliminates string derivation as an identity mechanism (**Source of Truth Clarity**).
**Affects:** Step 2b

**Triggered enhancement:** Generalize plant order targetType to parent ItemType level (show "grass seed" not "tall grass seed") — trigger: when multiple Kinds exist per parent ItemType

### DD-14: Clay is uniform, no varieties
**Context:** Reqs don't specify clay varieties, colors, or differentiation. Clay is a raw material input to bricks.
**Decision:** Uniform item. `NewClay()` constructor with ItemType "clay", no Kind, no variety registration, no color variation.
**Rationale:** No gameplay reason for varieties — clay exists to become bricks. Follows **Start With the Simpler Rule**.
**Affects:** Step 3

### DD-15: Dig is a separate order verb from Gather
**Context:** Gather targets ground items (sticks, nuts, shells). Clay creates items from terrain — fundamentally different source type. Extending gather would require branching in `findGatherIntent`, `isMultiStepOrderComplete`, `IsOrderFeasible`, and `GetGatherableTypes`.
**Decision:** "Dig Clay" is a new top-level ordered action (`ActionDig`, activity ID "dig", `AvailabilityKnowHow`). Separate intent finder `findDigIntent`, separate handler `applyDig`. Top-level order (no category) — can be broken into Dig > Clay, Dig > [Other] when future dig targets appear.
**Rationale:** The physical action (creating material from terrain) is different from gathering (picking up ground items). A separate verb avoids branching in the gather flow and extends naturally to future terrain-extraction actions (roots, trenches, stone, ore). Follows **Consider Extensibility** and **Isomorphism** — digging IS a different action from gathering.
**Affects:** Step 3

### DD-16: Dig order drops inventory first, drops clay on completion
**Context:** Characters have 2 inventory slots. Dig needs empty inventory to collect clay. Order completion needs a clear rule.
**Decision:** On taking a dig order, character drops all non-clay inventory items (procurement drop pattern). Digs until both inventory slots have clay (order complete). On completion, drops both clay items on the ground.
**Rationale:** Drop-on-completion keeps inventory free for other work between orders and makes clay immediately available on the ground for brick crafting (Step 4). Follows **Follow the Existing Shape** — mirrors bundle drop on gather completion.
**Affects:** Step 3

### DD-17: Dig discovery via loose clay items at world gen
**Context:** Dig requires know-how discovery, but clay comes from terrain — there are no items to look at before anyone digs. Options: (A) spawn loose clay items on clay tiles at world gen, (B) new trigger type for walking on terrain.
**Decision:** Spawn 2-3 loose clay items on clay tiles during world generation. Discovery triggers use existing item-based pattern: ActionLook on clay item, ActionPickup on clay item. This also gives characters something to form preferences about before anyone digs.
**Rationale:** Reuses existing discovery system with no new plumbing. Follows **Follow the Existing Shape** — item-based triggers are the established pattern. Terrain-look triggers deferred as a triggered enhancement.
**Affects:** Step 3

### DD-18: Drop-on-completion for dig uses separate logic from DropCompletedBundle
**Context:** `DropCompletedBundle` looks for items with `BundleCount >= MaxBundleSize`. Clay is not bundled — it's individual loose items in inventory slots.
**Decision:** Add a separate drop path for dig orders that iterates inventory and drops all items of the target type. Same call site in `selectOrderActivity` (after `isMultiStepOrderComplete` returns true), different logic from the bundle path.
**Rationale:** Clay isn't bundled and shouldn't have a bundle count. The drop mechanism is structurally different — dropping N individual items vs. dropping one bundle. Follows **Isomorphism** — different things shouldn't be forced into the same representation.
**Affects:** Step 3

### DD-19: Craft brick order is repeatable — completes when no loose clay remains
**Context:** Brick crafting should process all available clay, not just one lump. Quantity selection deferred. Options: (A) single-craft order like vessel/hoe, (B) repeatable order that loops until a world-state condition is met.
**Decision:** Add `Repeatable bool` field to Recipe. When true, `applyCraft` skips inline `CompleteOrder` and clears intent instead. The `selectOrderActivity` loop re-evaluates each tick: `findCraftIntent` finds more clay → craft → repeat. `isMultiStepOrderComplete` returns true when no clay items exist on the ground (`!groundItemOfTypeExists(items, "clay")`). Multiple characters assigned to the same order work in parallel.
**Rationale:** Follows **Follow the Existing Shape** — multi-step orders (gather, dig) already use the `isMultiStepOrderComplete` loop. Adding `Repeatable` to Recipe keeps the existing single-craft path unchanged for vessel/hoe while enabling the new pattern. Follows **Consider Extensibility** — future repeatable recipes (e.g., craft planks from logs) use the same field.
**Affects:** Step 4

### DD-20: VesselExcludedTypes split from MaxBundleSize
**Context:** `MaxBundleSize` map served double duty as both "bundleable" and "vessel-excluded" (DD-2). Triggered enhancement said to split when concepts diverge. Brick and clay are vessel-excluded but not bundleable — trigger condition met.
**Decision:** Add `VesselExcludedTypes map[string]bool` to config (stick, grass, clay, brick). Update vessel exclusion checks in `AddToVessel`, `CanVesselAccept`, `Pickup`, `findHarvestIntent`, and `findGatherIntent` to use `VesselExcludedTypes` instead of `MaxBundleSize`. Bundle logic (bundle merge, canGatherMore, hasFullBundle, DropCompletedBundle) stays on `MaxBundleSize`.
**Rationale:** Follows **Source of Truth Clarity** — each config set means one thing. Follows **Isomorphism** — vessel exclusion and bundleability are different concepts now that items can have one without the other.
**Affects:** Step 4

### DD-22: Character chooses fence material — no player sub-menu
**Context:** How does the player specify which material a fence should use? Options: player selects material in sub-menu (like harvest target type selection), or character chooses based on preference and availability (like the game's vision of character agency).
**Decision:** No material sub-menu. Order UI is Construction > Fence (no material step). The character assigned to the order evaluates available fence recipes and selects one based on material preference and availability. This aligns with the game vision: the player says *what* to build, the character decides *how*.
**Rationale:** Follows **Isomorphism** — characters have preferences and agency; material choice is a character decision, not a player directive. Follows the game vision: history exists in character decisions, not player micromanagement.
**Affects:** Step 6

### DD-23: Bricks stay individual — supply-drop for brick fences
**Context:** Brick fences require 6 bricks. Characters have 2 inventory slots. Options: (A) make bricks bundleable at 6 for one-trip carry, (B) multi-trip supply-drop where character carries 2 at a time, drops at build site, repeats.
**Decision:** Bricks stay individual (not bundled). Characters carry 2 bricks per trip, drop at build site, return for more until 6 are accumulated, then build. This creates a visible, narrative supply chain — characters shuttle materials to a construction site.
**Rationale:** The original requirements describe bricks as "6 bricks" distinct from "1 bundle of 6." The multi-trip pattern is isomorphic to real construction — materials are stockpiled before building. This pattern is also needed for huts (Step 8: "supplies dropped on tiles, then worked"). Introducing it at fence scale (one tile, 6 items) before hut scale (16 tiles, 12+ items each) follows **Start With the Simpler Rule** for infrastructure introduction.
**Affects:** Step 6, Step 8

### DD-24: Fence placement uses line/series tile marking
**Context:** Fences are linear structures — walls, borders, enclosures. Placement UI options: single-tile orders with TargetPosition, rectangle area selection (like tilling), or line/series marking.
**Decision:** Fence placement uses a line/series marking mode. Player marks individual tiles or draws lines of tiles for fence construction. Contiguous marked tiles form fence segments. Details of the line-drawing UX are an open question.
**Rationale:** Fences are lines, not rectangles. Single-tile placement creates order clutter for realistic fence lengths. Line marking matches the structural intent. Follows **Isomorphism** — the placement tool mirrors what the player is building.
**Affects:** Step 6

### DD-25: Material lock for contiguous fence segments
**Context:** When a character builds a fence line, contiguous tiles should use the same material — a wall shouldn't alternate between thatch and brick. But the character chooses the material (DD-22), not the player. How is consistency enforced?
**Decision:** Mechanism TBD — multiple options under discussion. The constraint is: contiguous marked fence tiles should be built with the same material. A second character working on the same line respects the existing material choice. The central design question is what defines a "line" or "segment" — this determines how the lock scopes and propagates. Open questions: segment definition, when the lock forms (first tile built vs marking time), what happens when materials run out, and the storage mechanism (data on marked tiles, on the order, or derived from placed constructs).
**Rationale:** Visual and structural consistency in built structures. Follows **Isomorphism** — a real fence section is one material.
**Affects:** Step 6

### DD-26: Building from adjacent tile
**Context:** Should a character stand on the build tile or work from adjacent? Standing on the tile and placing an impassable construct creates a "trapped inside my own fence" problem.
**Decision:** Characters build from a cardinally adjacent tile, not the build tile itself. Handler checks `char.Pos().IsCardinallyAdjacentTo(buildPos)`. Additional collision scenarios (another character on the build tile, blocked access) are open questions.
**Rationale:** Follows **Follow the Existing Shape** — water drinking uses adjacent-tile interaction for impassable targets.
**Affects:** Step 6

### DD-21: Brick is uniform with terracotta color
**Context:** Brick appearance properties. Only one brick type exists (from clay). Options: inherit clay's earthy color, or give bricks a distinct color.
**Decision:** Brick uses new `ColorTerracotta` (warm reddish-brown, distinct from clay's earthy), symbol `▬` (`CharBrick`), ItemType "brick", no Kind, no variety. Uniform item like clay (DD-14).
**Rationale:** Terracotta is visually distinct from raw clay — shaped bricks look different from loose lumps. Follows **Isomorphism** — the transformation from clay to brick should be visible. Follows **Start With the Simpler Rule** — no Kind until multiple brick types exist.
**Affects:** Step 4
