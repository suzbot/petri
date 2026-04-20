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
- Door passability mechanism: currently passability is boolean. Doors need "character-passable but creature-impassable" (for future Threats phase). May need a passability enum or just Passable=true until creatures exist. (Deferred — relevant to Step 10 hut execution, not Step 5.)

---

### Step 6: Build Fence
**Status:** Complete

**Anchor story:** The player creates a Construction > Fence order. They enter a line-placement mode and mark several contiguous tiles for fence construction. A character takes the order and decides to use sticks. They find a bundle of 6 sticks (open question: gather their own bundle or rely on pre-gathered full bundles), walk to one of the marked tiles, and build. Another character takes a fence construction order and starts building from the other end. The line already has a material choice — sticks — so they build their section out of sticks too. For a brick fence segment, a character makes multiple trips — carrying 2 bricks at a time, dropping them at the build site, and returning for more until 6 are accumulated, then building.

**Scope:**
- buildFence activity in ActivityRegistry, "Construction" order UI category
- Three fence recipes in RecipeRegistry: thatch (1 grass bundle of 6), stick (1 stick bundle of 6), brick (6 bricks) — per-recipe discovery, custom execution path (DD-23, DD-27)
- Fence placement UI: line/series tile marking (DD-24)
- Character-driven material selection: nearest available for v1, preference-based deferred to Step 12 — no player sub-menu (DD-22, DD-31)
- Material lock via line ID: per-line marking with material propagation on first build (DD-25)
- Material procurement: bundle gathering for grass/sticks, multi-trip supply-drop for bricks (DD-23, DD-30)
- Build fence handler with adjacent-tile building, layered collision handling (DD-26, DD-28)
- Order feasibility, completion, multiple workers with no tile claiming (DD-29)
- Details panel shows material on marked-for-construction tiles
- Discovery triggers on fence recipes (activity know-how auto-granted on first recipe discovery)

**Open questions:**
- ~~Material sub-menu~~ → DD-22 (character chooses, no player sub-menu)
- ~~Brick bundleability~~ → DD-23 (bricks stay individual, supply-drop pattern)
- ~~Bundle procurement~~ → DD-30 (gather one-by-one using existing pickup/merge; full bundles picked up in one step)
- ~~Line selection UI~~ → DD-24 (line drawing with anchor/confirm, cardinal snap)
- ~~Material lock mechanism~~ → DD-25 (line ID on marked tiles; material set when first tile built, propagated to all tiles with same line ID)
- ~~Build position collision~~ → DD-28 (skip occupied tiles, abandon if occupied during build, displace as safety net)
- ~~Multiple workers~~ → DD-29 (no tile claiming; overlap benign for bundles; excess bricks reusable for next tile)

**Triggered enhancements:**
- `continueIntent` early-return block consolidation — already at 6 blocks, trigger threshold was met before Step 5 (not worsened by fence building, which uses position-based intent with no new block). Evaluate independently of construction.
- Order-aware simulation for e2e testing — construction adds multiple new ordered actions with supply procurement.
- Category type formalization — evaluate whether `VesselExcludedTypes` set is sufficient or whether formal item categories are needed.

---

### Step 7: Construct Interaction
**Status:** Complete

**Anchor story:** A character wanders past a stick fence and pauses to look at it. They decide they like the look of stick fences — "Likes stick fences" appears in their preferences panel. Another character, in a bad mood, looks at a brick fence and develops "Dislikes brick fences."

**Scope:**
- Extend Look idle activity to target constructs: `findLookIntent` searches both items and constructs, picks nearest (DD-34)
- `TargetConstruct *Construct` on Intent (ephemeral, not serialized) (DD-34)
- `applyLook` branches on TargetConstruct vs TargetItem (DD-34)
- `CompleteLookAtConstruct` handles construct-specific mood adjustment and preference formation (DD-34)
- Material preference formation using existing Preference.Kind field — "stick fence", "thatch fence", "brick fence" as recipe-level identities (DD-35)
- `MatchesConstruct`, `NetPreferenceForConstruct` for preference matching against constructs (DD-35)
- `TryFormConstructPreference` for construct-specific formation rolls (DD-35)
- Discovery from looking at constructs: skipped for now (DD-36)

**Open questions:**
- ~~How should Look be extended to target constructs?~~ → DD-34
- ~~Preference scope: material only, construct+material combo, or both?~~ → DD-35

---

### Step 8: Construct Discovery + Hut Placement UI
**Status:** Complete

**Anchor story:** A character wanders past a stick fence and pauses to look at it. They form a preference, and also have a moment of insight — they realize they could build a stick hut! "Discovered: Build Hut" appears in the action log. Another character who already knows how to build looks at a brick fence and discovers the brick hut recipe specifically. The player selects Construction > Build Hut and enters placement mode. A 5×5 footprint preview follows the cursor. They place a hut overlapping some planned fence marks — the fence marks are overwritten. They place a second hut sharing a wall — the shared tiles keep the first hut's LineID. They confirm — 16 perimeter tiles are marked for construction. No character can build yet, but the order and marks exist.

**Scope:**

*Sub-step 8a — Construct discovery (complete):*
- `ConstructKind string` field on `DiscoveryTrigger` (DD-37)
- `TryDiscoverFromConstruct` function called from `CompleteLookAtConstruct` — mirrors `TryDiscoverKnowHow` structure (tries activities then recipes)
- buildHut activity in ActivityRegistry with `AvailabilityKnowHow`, Category "construction"
- Three hut recipes in RecipeRegistry (thatch-hut, stick-hut, brick-hut) with per-tile inputs and discovery triggers: `ActionLook` + `ConstructKind: "fence"`
- First recipe discovery auto-grants buildHut activity know-how (same pattern as fence recipes)
- Order pipeline guards: `findOrderIntent` nil stub, `IsOrderFeasible` false stub

*Sub-step 8b — Hut placement UI:*
- `ConstructKind string` field on `ConstructionMark` to distinguish fence vs hut marks (DD-41)
- Retroactive: fence marks carry `ConstructKind: "fence"`
- Fixed 5×5 footprint placement mode (DD-38) — cursor positions top-left corner, `p` confirms
- Nuanced placement overlap rules (DD-46): perimeter tiles can overlap existing marks (first-wins for hut shared walls, overwrite for fence marks), interior tiles block on existing marks
- Items/plants/tilled soil don't block placement (DD-47) — displaced/destroyed/reverted during construction
- Whole-footprint unmark via LineID (DD-43) — Tab toggles, `p` removes entire footprint
- Fence marks render in grey during hut placement; amber when covered by preview (DD-48)
- Details panel labels include ConstructKind (DD-45): "Marked for construction (Hut)" / "Marked for construction (Stick Hut)"
- Feasibility follows fence pattern — any material on map = feasible (DD-44)
- `HasUnbuiltConstructionPositions`, `isMultiStepOrderComplete`, `IsOrderFeasible` filter by ConstructKind
- `UnmarkByLineID` for whole-footprint removal
- `findOrderIntent` nil stub stays — execution deferred to Step 10
- Save/load: `ConstructKind` in serialized `ConstructionMark`, backward compat defaults to "fence"

**Rationale for merge:** The buildHut activity and recipes must land alongside the order pipeline (`findOrderIntent` case, `IsOrderFeasible`, area selection UI) to avoid a gap where recipes exist but `findOrderIntent`'s default branch routes them to `findCraftIntent`. This mirrors how fence recipes were added in Step 6 alongside the full fence pipeline.

---

### Step 9: Hut Construct Types + Rendering
**Status:** Complete

**Anchor story:** The player creates a test world with a completed hut. The hut renders as a clean 5×5 enclosure using heavy box-drawing characters — corners (`┏┓┗┛`), horizontal edges (`━`), vertical edges (`┃`), and a door (`▯`) on the south wall. Corners and doors have asymmetric horizontal fill for visual continuity. The walls are colored by material (brown for stick, pale yellow for thatch, terracotta for brick). Characters path around the walls but walk freely through the door. The details panel shows "Stick Hut Wall" or "Stick Hut Door" with "Structure" and passability info.

**Scope:**
- ~~Demo program to evaluate line-drawing character options for walls and door symbols~~ → DD-42 (resolved: heavy box-drawing)
- `WallRole string` field on Construct for position-aware symbol and display name (DD-50)
- `NewHutConstruct(x, y, material, materialColor, wallRole)` constructor — single constructor for walls and doors; Kind `"hut"` for both (DD-50); wallRole determines symbol and passability (door role → `Passable: true`)
- Heavy box-drawing symbols: `┏ ┓ ┗ ┛ ━ ┃` for walls, `▯` for door (DD-42)
- Construct rendering with asymmetric horizontal fill, details panel (`DisplayName` uses WallRole), movement blocking (follows Adding a New Construct Type checklist)
- Save/load serialization for WallRole field on ConstructSave
- Test with pre-placed constructs via test-world

**Open questions:**
- ~~Hut wall/door visual symbols: line-drawing Unicode characters for walls, door symbol TBD — resolve via demo program during step refinement~~ → DD-42

---

### Step 10: Build Hut Execution
**Status:** Complete

**Anchor story:** A character who knows how to build huts takes a Build Hut order. They choose stick material (nearest available), and the material is stamped across all 16 hut marks. They carry bundles of sticks to the nearest unbuilt wall tile, dropping 2 full bundles, then work the tile into a wall segment. They move to the next tile. Another character joins and works from the other end. Eventually all 16 tiles are built — walls and a door. Characters walk through the door into the 3×3 interior.

**Scope:**
- `findBuildHutIntent` in order_execution.go — follows fence pattern: nearest unbuilt hut mark, material selection (DD-40), supply-drop procurement for all materials (DD-39)
- Material costs per tile: thatch = 2 bundles of 6 grass, stick = 2 bundles of 6 sticks, brick = 12 bricks (DD-39)
- Supply-drop for all material types: bundles dropped at site then consumed, same as bricks (DD-39)
- `applyBuildHut` handler — walk-then-act, supply delivery, build phase, construct placement with WallRole "wall"/"door" from mark (DD-51)
- Render-time symbol refactor: simplify WallRole on Construct to "wall"/"door", replace `wallRoleToSymbol` with adjacency-based symbol computation in `renderCell()` (DD-42, DD-50 updated)
- Material selection generalized: `selectConstructionMaterial(activityID)` replaces `selectFenceMaterial`, shared by fence and hut building
- Multiple workers, no tile claiming (same as fences DD-29)
- Collision handling: skip/abandon/displace layers (same as fences DD-28)
- Item displacement after construct placement (same as fences DD-33)
- Order feasibility: unbuilt hut marks exist AND hut materials exist on map (already stubbed)
- Order completion: no unbuilt hut marks remain (already stubbed)

**Open questions:**
- ~~Hut supply management~~ → DD-39 (per-tile sequential, supply-drop for all materials)
- ~~Multiple hut materials~~ → DD-40 (no mixing; character selects, same as fences)
- ~~WallRole on marks~~ → DD-51 ("wall"/"door" set at placement time, read at build time)

---

### Step 11: Preference-Weighted Item Seeking
**Status:** In progress (11a Complete, 11b Complete, 11c Complete)

**Anchor story:** A character who likes stick fences and dislikes grass takes a Build Fence order. Thatch grass bundles are nearby and a stick bundle is 25 tiles away. The character walks past the grass to the sticks — scoring against the anticipated output (a synthetic stick fence), the Kind preference ("likes stick fences", weighted 2×) outweighs the extra distance. Another character gets a Build Hut order. They dislike loose bricks but like brick huts — Kind weight (+2) outweighs material dislike (-1), and they choose bricks. A third character crafting a hoe needs a shell — two silver shells and a brown shell are on the ground. The silver is a few tiles further but the character prefers silver, so they walk to it. A fourth character with a Plant order and no variety lock yet sees red and blue berries — they prefer red, and grab the red berry despite the blue being closer.

**Scope:**
- Recipe Output.Kind on construction recipes: add Kind to fence and hut recipe outputs so all recipes carry a preference-matchable identity (DD-59)
- Combined synthetic output scoring with weighted attribute matching: build a synthetic output (Construct for construction, Item for crafts) merging recipe identity with material attributes, score preferences against it with Kind weighted 2× (DD-52)
- Material field on Item: crafted items carry input material identity so material preferences cross-apply to crafted products (DD-61)
- Preference-weighted recipe selection in `selectConstructionMaterial`: score feasible recipes via synthetic Construct, replacing pure nearest (DD-52)
- Preference-weighted recipe selection in `findCraftIntent`: score feasible recipes via synthetic Item with Material field, replacing first-feasible — no competing craft recipes exist yet, but infrastructure ready (DD-52)
- Preference-weighted item procurement in `EnsureHasRecipeInputs` and `EnsureHasItem`: score items of target type by preference + distance, replacing pure nearest (DD-52)
- Preference-weighted vessel selection in `EnsureHasVesselFor` / `FindAvailableVessel`: score vessels by preference + distance, replacing pure nearest (DD-52)
- Preference-weighted plantable selection in `EnsureHasPlantable`: score by preference + distance before variety lock; after lock, only one variety matches so scoring is moot (DD-52)
- Resolve architecture.md "Future: Unified Item-Seeking" gap

**Sub-steps:**
- 11a: Scoring infrastructure (`ScoreConstructPreference`, `ScoreItemFit`, config) + Output.Kind on construction recipes + `selectConstructionMaterial`
- 11b: Item procurement — `findPreferredItemByType` + `EnsureHasItem` + `EnsureHasRecipeInputs` use preference scoring. No new fields; scores real items with existing attributes.
- 11c: Material field on Item (structural + save/load + `Preference.Matches` update) + `ScoreItemPreference` (weighted Item scoring) + `findCraftIntent` recipe scoring. Effectively no-op today (one recipe per craft activity), but infrastructure for future competing recipes.
- 11d: Vessel selection (`FindAvailableVessel`) + plantable selection (`FindVesselContaining`, `EnsureHasPlantable`) use preference scoring.

**Open questions:**
- ~~Scope may be too narrow~~ → Evaluated: all item-seeking paths reviewed. Variety-locked paths (`FindNextVesselTarget`, `EnsureHasPlantable` after lock) have no meaningful choice to weight. All other paths included.
- ~~Behavioral scope of DD-61 Material field~~ → Confirmed: Material cross-applies to all contexts (mood, preference formation, not just item seeking). Parallels how construct preferences already work ("likes bricks" boosts mood for brick fences via `MatchesConstruct`).

---

### Step 12: Display, Label, and Color Polish
**Status:** Planned

**Scope:**
- Redundant display names: sticks display as "stick" not "brown stick"; tall grass displays as "tall grass" not "pale green tall grass" (DD-53)
- Validate bricks display as "brick" (Name field should handle this; verify in action log)
- Pluralization: "Gather clays" → "Gather lumps of clay" (DD-54)
- Abandoned order cooldown timer: show remaining time as "Abandoned (1:53)" in orders panel (DD-55)
- Order menu labels: construction orders currently show as "Fence" / "Hut" — add "Build " prefix to match craft pattern ("Craft vessel"), producing "Build fence" / "Build hut" (DD-56)
- Discovery randomization: randomize activity-vs-recipe order in `TryDiscoverKnowHow` so overlapping triggers don't always resolve in the same order (DD-57)
- Thatch color evaluation: demo app comparing ColorPaleYellow (ANSI 229) against gold/wheat candidates (ANSI 178, 179, 186) for dried grass and thatch constructs (DD-58)

---

### Step 13: Tuning Pass
**Status:** Planned

**Scope:**
- Extraction duration: ActionDurationShort (0.83s) may be too fast for seed extraction — evaluate and adjust
- Extraction seed yield: currently 1 seed per extraction — evaluate whether yield should increase
- Duration and yield should be tuned together via in-game observation


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

### DD-21: Brick is uniform with terracotta color
**Context:** Brick appearance properties. Only one brick type exists (from clay). Options: inherit clay's earthy color, or give bricks a distinct color.
**Decision:** Brick uses new `ColorTerracotta` (warm reddish-brown, distinct from clay's earthy), symbol `▬` (`CharBrick`), ItemType "brick", no Kind, no variety. Uniform item like clay (DD-14).
**Rationale:** Terracotta is visually distinct from raw clay — shaped bricks look different from loose lumps. Follows **Isomorphism** — the transformation from clay to brick should be visible. Follows **Start With the Simpler Rule** — no Kind until multiple brick types exist.
**Affects:** Step 4

### DD-22: Character chooses fence material — no player sub-menu
**Context:** How does the player specify which material a fence should use? Options: player selects material in sub-menu (like harvest target type selection), or character chooses based on preference and availability (like the game's vision of character agency).
**Decision:** No material sub-menu. Order UI is Construction > Fence (no material step). The character assigned to the order evaluates available fence recipes and selects one based on material preference and availability. This aligns with the game vision: the player says *what* to build, the character decides *how*.
**Rationale:** Follows **Isomorphism** — characters have preferences and agency; material choice is a character decision, not a player directive. Follows the game vision: history exists in character decisions, not player micromanagement.
**Affects:** Step 6

### DD-23: Bricks stay individual — supply-drop for brick fences
**Context:** Brick fences require 6 bricks. Characters have 2 inventory slots. Options: (A) make bricks bundleable at 6 for one-trip carry, (B) multi-trip supply-drop where character carries 2 at a time, drops at build site, repeats.
**Decision:** Bricks stay individual (not bundled). Characters carry 2 bricks per trip, drop at build site, return for more until 6 are accumulated, then build. This creates a visible, narrative supply chain — characters shuttle materials to a construction site.
**Rationale:** The original requirements describe bricks as "6 bricks" distinct from "1 bundle of 6." The multi-trip pattern is isomorphic to real construction — materials are stockpiled before building. This pattern is also needed for huts (Step 10: "supplies dropped on tiles, then worked"). Introducing it at fence scale (one tile, 6 items) before hut scale (16 tiles, 12+ items each) follows **Start With the Simpler Rule** for infrastructure introduction.
**Affects:** Step 6, Step 8

### DD-24: Fence placement uses line drawing with anchor/confirm
**Context:** Fences are linear structures — walls, borders, enclosures. Placement UI options: single-tile orders with TargetPosition, rectangle area selection (like tilling), or line/series marking.
**Decision:** Fence placement uses a line-drawing mode that reuses the area selection infrastructure with a line constraint. UX flow: press `p` to anchor start point → move cursor → preview shows a cardinal line (horizontal or vertical; diagonal cursor movement snaps to the axis with the larger delta) → press `p` to confirm, marking all valid tiles along the line. Single-tile marking is the degenerate case (anchor and confirm on the same tile). Tab toggles mark/unmark mode. Player can draw multiple lines before pressing Enter to create the order. Validator rejects tiles with water, existing constructs, or impassable features.
**Rationale:** Fences are lines, not rectangles. Line drawing gives good UX for walls and borders while reusing area selection infrastructure. Cardinal-only lines prevent awkward diagonal fences. The snap behavior (larger axis wins) is simple and predictable. Follows **Isomorphism** — the placement tool mirrors what the player is building. Follows **Follow the Existing Shape** — reuses area selection's anchor/confirm/mode-toggle pattern.
**Affects:** Step 6

### DD-25: Material lock via line ID on marked tiles
**Context:** When a character builds a fence line, contiguous tiles should use the same material — a wall shouldn't alternate between thatch and brick. But the character chooses the material (DD-22), not the player. How is consistency enforced?
**Decision:** Each line-drawing operation assigns a shared Line ID (incrementing counter) to all marked tiles. Storage: `markedForConstruction map[Position]ConstructionMark` where `ConstructionMark` has `LineID int` and `Material string`. Material starts as `""` (unset). When the first character builds any tile in the line, they pick a material (nearest available for v1) and that material is stamped onto all `ConstructionMark` entries with the same Line ID. Subsequent workers on the same line see the material already set. Cross-line adjacency does NOT enforce consistency — two separately-drawn lines that meet at a corner can have different materials (interesting emergent behavior). When the locked material runs out, the order becomes unfulfillable via `IsOrderFeasible`. Details panel shows the material on marked tiles (e.g., "Marked for construction (Stick)").
**Rationale:** The line-drawing operation is a natural unit of player intent — one draw = one wall = one material. Line ID captures this cleanly without complex adjacency scanning or multi-axis analysis. Visual and structural consistency within a drawn line. Follows **Isomorphism** — a real fence section is one material. Follows **Source of Truth Clarity** — the line ID is the single source of truth for which tiles share a material constraint.
**Affects:** Step 6

### DD-26: Building from adjacent tile
**Context:** Should a character stand on the build tile or work from adjacent? Standing on the tile and placing an impassable construct creates a "trapped inside my own fence" problem.
**Decision:** Characters build from a cardinally adjacent tile, not the build tile itself. Handler checks `char.Pos().IsCardinallyAdjacentTo(buildPos)`. Additional collision scenarios (another character on the build tile, blocked access) are open questions.
**Rationale:** Follows **Follow the Existing Shape** — water drinking uses adjacent-tile interaction for impassable targets.
**Affects:** Step 6

### DD-27: Fence material requirements use RecipeRegistry — custom execution path
**Context:** Fence building requires specific materials per tile (1 grass bundle of 6, 1 stick bundle of 6, or 6 bricks). Options: (A) define material requirements as ad-hoc config data with a single "buildFence" activity, (B) define as RecipeRegistry entries with per-recipe discovery.
**Decision:** Fence material requirements are RecipeRegistry entries (thatch-fence, stick-fence, brick-fence), each with its own DiscoveryTriggers. Characters discover each fence recipe independently — a character who's worked with sticks might know stick fences but not brick fences. The existing `tryDiscoverRecipe` mechanism automatically grants the "buildFence" activity when the first fence recipe is discovered. However, `findOrderIntent` routes "buildFence" to a custom `findBuildFenceIntent` (not `findCraftIntent`), because the procurement pattern (bundle gathering, multi-trip supply-drop) and output type (construct, not item) don't fit the craft execution path. The recipes serve as the data/discovery/knowledge-display layer; execution is handled by the buildFence action's own intent finder and handler.
**Rationale:** Recipes are the game's mental model for "what do I need to create this thing?" — this applies equally to items and constructs. Per-recipe discovery gives characters individual knowledge about specific building methods, which is richer than all-or-nothing activity discovery. The existing recipe discovery system (`tryDiscoverRecipe` grants activity + recipe) handles this with no new plumbing. Custom execution is necessary because fence building is fundamentally different from crafting: position-based with adjacent-tile building (DD-26), multi-phase procurement (DD-23), and construct output. Follows **Follow the Existing Shape** for discovery, **Isomorphism** for separating craft from construction execution.
**Affects:** Step 6

### DD-28: Build position collision — layered skip/abandon/displace
**Context:** What happens when another character stands on or walks into the fence build tile during construction?
**Decision:** Three layers of handling, increasing in rarity: (1) **Skip** — the builder's intent finder skips marked tiles that have a character standing on them, targeting the next available tile instead. (2) **Abandon** — if a character moves onto the build tile during the build action (between ticks), the builder abandons that tile and re-evaluates on the next tick. (3) **Displace** — safety net for race conditions: if a construct is placed and a character happens to be at that position (simultaneous movement), displace the occupant to the nearest empty cardinally-adjacent tile. Layer 3 should be rare given layers 1 and 2.
**Rationale:** Belt-and-suspenders approach prevents edge cases without complex mutex logic. Primary avoidance (skip) handles the common case; abandonment handles mid-build changes; displacement handles the timing edge case. Follows **Start With the Simpler Rule** — each layer is simple, combined they cover all cases.
**Affects:** Step 6

### DD-29: Multiple workers — no tile claiming
**Context:** Can two workers supply-drop bricks to the same tile, or is each tile claimed by one worker?
**Decision:** No tile claiming mechanism. Workers independently target the nearest unbuilt marked tile. For bundle materials (grass/sticks), each worker carries one full bundle per tile — no overlap issues since the bundle is consumed atomically at build time. For bricks, two workers might both deliver to the same tile, but excess bricks stay on the ground and are picked up for the next tile. Materials are never wasted, just relocated.
**Rationale:** Follows **Start With the Simpler Rule** — claiming requires tracking state that isn't necessary. The worst case (two workers delivering bricks to the same tile) is self-correcting — excess materials are reused.
**Affects:** Step 6

### DD-30: Bundle procurement with enhanced Pickup merge
**Context:** Fence building with bundle materials (grass/sticks) requires a full bundle of 6. Characters have 2 inventory slots. How does the character assemble the bundle?
**Decision:** The intent finder searches for the nearest bundle-compatible item on the ground (including full and partial bundles), then falls back to individual items. `Pickup()` is enhanced to merge by the picked-up item's `BundleCount` (not always +1): `carried.BundleCount += item.BundleCount`, with guard `carried.BundleCount + item.BundleCount <= maxSize` (skip if merge would overflow — no partial splitting in v1). This makes partial and full bundle pickup work correctly everywhere (gather, fence, any future bundle user). Full bundles on the ground go into empty inventory slots without merging. The intent finder loops (intent cleared after pickup → re-evaluated next tick → next pickup) until the character has a full bundle of 6, then transitions to the build phase.
**Rationale:** Fixing `Pickup()` to merge by actual BundleCount is a general improvement — partial bundles dropped by one character can be efficiently picked up by another. The fence-specific search (including full bundles) is needed because `findNearestItemByType` skips full bundles by design (correct for gather, wrong for fence procurement). Follows **Follow the Existing Shape** — pickup/merge pattern, component procurement flow.
**Affects:** Step 6 (also improves gather behavior globally)

### DD-31: Material selection — nearest available for v1
**Context:** When a line's material is unset and the first tile needs to be built, how does the character choose which material to use? DD-22 says character chooses based on preference and availability.
**Decision:** For v1, the character selects the nearest available material. "Available" means enough material exists in the world to build one fence tile (6+ of any fence material type). The character picks the material type whose nearest instance is closest. Preference-based selection (characters favoring materials they like) is deferred to Step 12 alongside the broader preference-weighted target selection enhancement.
**Rationale:** Follows **Start With the Simpler Rule** — nearest-distance selection works and produces reasonable results. Preference-based selection is already scoped for Step 12 where it can be done properly across all item-seeking activities.
**Affects:** Step 6

### DD-32: findOrderIntent restructure — switch first, recipe fallback in default
**Context:** `findOrderIntent` currently checks for recipes first and routes all recipe-having activities to `findCraftIntent`. This means every non-craft activity with recipes (fence building, future cooking) needs a hardcoded exception before the recipe check. This doesn't scale.
**Decision:** Restructure `findOrderIntent` to put the switch statement first, with the recipe check in the default case. Specific activities (harvest, tillSoil, buildFence, future cook) get their custom intent finders via switch cases. Only truly generic craft activities (craftVessel, craftHoe, craftBrick) fall through to `findCraftIntent` via the default case. Adding a new recipe-using activity category just adds a switch case — no exceptions, no new fields.
**Rationale:** Follows **Consider Extensibility** — the dispatch structure accommodates future recipe categories (cooking, alchemy) with no structural changes. Follows **Isomorphism** — activities with distinct execution patterns (crafting vs building vs cooking) get distinct handlers, with recipes as shared data infrastructure.
**Affects:** Step 6

### DD-33: Item displacement when fence is placed
**Context:** When a fence construct is placed on a tile, items may exist on that tile (excess bricks from supply-drop, food someone dropped, etc.). An impassable construct would make these items inaccessible.
**Decision:** After placing a fence, displace all remaining items at the build position to nearest cardinally-adjacent empty tiles. Follows the `applyTillSoil` pattern which displaces non-growing items when tilling. For brick fences, the handler consumes exactly 6 bricks from the tile first, then displaces any remainder.
**Rationale:** Follows **Follow the Existing Shape** — `applyTillSoil` already handles item displacement on terrain change. Items trapped under impassable constructs would be permanently inaccessible — displacement prevents item loss.
**Affects:** Step 6

### DD-34: Look targets both items and constructs via nearest-lookable search
**Context:** Look currently only targets items via `findNearestItemExcluding`. Constructs need to be lookable for preference formation. Options: separate "look at construct" activity, or extend existing Look to search both entity types.
**Decision:** Extend `findLookIntent` to search both items and constructs, picking the nearest lookable entity. Add `TargetConstruct *Construct` to Intent (ephemeral, not serialized — follows `TargetBuildPos` pattern). `applyLook` branches: if `TargetConstruct != nil`, walk to it and call `CompleteLookAtConstruct`; otherwise use existing item path. The "exclude last-looked" mechanism already works by position, so it naturally prevents looking at the same construct twice in a row.
**Rationale:** Follows **Follow the Existing Shape** — Intent already has typed target fields per entity type (`TargetItem`, `TargetFeature`, `TargetCharacter`). Adding `TargetConstruct` is the same pattern. Follows **Isomorphism** — constructs are not items, so they get their own target field and completion handler.
**Affects:** Step 7

### DD-35: Construct preferences use Kind field — recipe-level identity
**Context:** What preference attributes form when looking at constructs? The requirements say "form a like/dislike materials in the construction." Options: (A) material only as ItemType ("Likes sticks"), (B) material+construct combo as Kind ("Likes stick fences"), (C) new Preference field for construct attributes.
**Decision:** Use the existing `Kind` field on Preference for the construct's recipe-level identity: "stick fence", "thatch fence", "brick fence". This parallels how crafted items use Kind for recipe subtypes ("shell hoe", "hollow gourd"). Color maps to `MaterialColor`. Formation roll picks from: solo Kind ("Likes stick fences"), solo Color ("Likes brown"), or Kind+Color combo ("Likes brown stick fences"). No generic "likes fences" level — that doesn't help differentiate recipes, which is the primary use case for these preferences. Add `MatchesConstruct(*Construct)` on Preference and `NetPreferenceForConstruct(*Construct)` on Character. `TryFormConstructPreference` handles construct-specific formation rolls.
**Rationale:** "Stick fence" is a recipe identity, just like "shell hoe." Kind already serves this purpose — no new fields needed. This lays the foundation for preference-weighted recipe selection (Step 10): a character who "likes stick fences" will prefer the stick-fence recipe. Follows **Follow the Existing Shape** (Kind for recipe identity), **Isomorphism** (preferences about constructs reflect the recipe that produced them), and **Consider Extensibility** (future construct types like huts produce "stick hut", "brick hut" preferences using the same mechanism).
**Affects:** Step 7, Step 12

### DD-36: Discovery from looking at constructs — deferred from Step 7, resolved in DD-37
**Context:** `CompleteLook` currently calls `TryDiscoverKnowHow` for item-based discovery triggers. Should looking at constructs also trigger discovery?
**Decision:** Skipped for Step 7. Resolved for Step 8 in DD-37 — `ConstructKind` field on `DiscoveryTrigger`, with `CompleteLookAtConstruct` calling construct-specific discovery.
**Rationale:** Follows **Start With the Simpler Rule** for Step 7 (no requirements). Step 8 hut discovery provides the concrete use case.
**Affects:** Step 7 (skipped); Step 8 (resolved via DD-37)

### DD-37: Discovery from looking at constructs — construct-based triggers
**Context:** DD-36 deferred construct-based discovery to Step 8. Characters need to discover hut recipes by looking at fences. Current discovery triggers use `ActionType` + `ItemType` + optional requirements. Constructs aren't items, so item-based triggers can't express "looked at a fence." Options: (A) add `ConstructKind` field to `DiscoveryTrigger`, (B) overload `ItemType` for construct material.
**Decision:** Add `ConstructKind string` field to `DiscoveryTrigger`, parallel to `ItemType`. `CompleteLookAtConstruct` calls a new `tryDiscoverFromConstruct(char, ActionLook, constructKind)` function that checks triggers with matching `ConstructKind`. Each hut recipe has a discovery trigger: `ActionLook` + `ConstructKind: "fence"` — looking at any fence can trigger discovery of hut recipes. Discovery of the first hut recipe auto-grants "buildHut" activity know-how (same pattern as fence recipes granting "buildFence" activity).
**Rationale:** Clean parallel to `ItemType` — constructs get their own trigger field rather than being shoehorned into item triggers. Follows **Isomorphism** — constructs are not items. Follows **Follow the Existing Shape** — same trigger mechanism, new field. Follows **Consider Extensibility** — future construct-triggered discoveries (e.g., looking at a hut triggers furniture recipes) use the same field.
**Affects:** Step 8

### DD-38: Hut uses 5×5 footprint with 16 perimeter tiles and 3×3 interior
**Context:** Original design doc said "4×4 outline ... 3×3 opening inside" but a 4×4 grid has only a 2×2 interior (12 perimeter tiles). The requirements say "16 tiles" for material costs. A 5×5 grid has 16 perimeter tiles and a 3×3 interior — matching both "16 tiles" and "3×3 opening."
**Decision:** Hut footprint is 5×5. 16 perimeter tiles are marked for construction (walls + door). 9 interior tiles (3×3) are open space. Area selection uses a fixed 5×5 footprint positioned by cursor (top-left corner). Player positions and presses `p` to confirm — marks all 16 perimeter tiles. Validator checks the full 5×5 area: no water, no existing constructs, no impassable features.
**Rationale:** Reconciles "16 tiles" with "3×3 opening." Follows **Follow the Existing Shape** — reuses area selection infrastructure (DD-6). Fixed footprint (not free-form) because huts have structural requirements (walls must form a complete enclosure).
**Affects:** Step 8

### DD-39: Per-tile sequential supply-drop and build
**Context:** How does the worker distribute supplies? Options: (A) drop all supplies on all tiles first, then build all tiles; (B) per-tile: drop supplies for one tile, build it, move to next.
**Decision:** Per-tile sequential with supply-drop for all material types. Worker picks the nearest unbuilt marked tile, delivers supplies by dropping them at the tile until it has the required amount, then builds that tile, then moves to the next. Material costs per tile: thatch = 2 full bundles of 6 grass, stick = 2 full bundles of 6 sticks, brick = 12 bricks. All materials are supply-dropped at the build site before working the tile (per requirements: "supplies dropped on them and then each tile with all its supplies is worked"). Characters carry materials using existing inventory (2 slots), making multiple trips per tile as needed. For bundles: 1 bundle per slot, 1 trip to deliver both bundles (carry 2 → drop 2 → build). For bricks: 2 per trip, 6 trips per tile.
**Rationale:** Matches requirements language ("supplies dropped on tiles, then worked"). Per-tile produces visible incremental progress (each completed wall segment is immediately visible). Global "drop everything first" would require dozens of trips before any construction happens. Uniform supply-drop for both bundles and bricks simplifies the execution flow.
**Affects:** Step 10

### DD-40: Character selects hut material — same as fences, no player sub-menu
**Context:** Should the player select hut material via sub-menu, or should the character choose? Huts are large (16 tiles, major investment). Fences use character choice (DD-22).
**Decision:** Same as fences — no material sub-menu. Character selects material based on known recipes and nearest availability. Material is stamped across all hut perimeter marks (same line-ID mechanism as fences). One material per hut — no mixing.
**Rationale:** Consistency with fence pattern (DD-22) — the player says *what* to build, the character decides *how*. Follows **Follow the Existing Shape**. Character agency in material choice is a core game vision element. If the player wants a specific material, they ensure it's the most available — indirect control through world state, not direct micromanagement.
**Affects:** Step 8

### DD-41: ConstructKind field on ConstructionMark — distinguishes fence vs hut marks
**Context:** `isMultiStepOrderComplete` and `IsOrderFeasible` currently check ALL marked-for-construction positions. With both fences and huts sharing the pool, a fence order shouldn't complete based on hut marks, and vice versa. Options: (A) add `ConstructKind` to `ConstructionMark`, (B) separate pools.
**Decision:** Add `ConstructKind string` field to `ConstructionMark` ("fence" or "hut"). `HasUnbuiltConstructionPositions`, `isMultiStepOrderComplete`, and `IsOrderFeasible` filter by kind. Fence orders only see fence marks; hut orders only see hut marks. The shared `markedForConstruction` pool stays unified — kind is a filter, not a separate data structure.
**Rationale:** Follows **Follow the Existing Shape** — extends the existing pool rather than creating a parallel one. Follows **Consider Extensibility** — future construct types (watchtower, workshop) add kind values, not new pools. Follows **Source of Truth Clarity** — the mark carries its own identity.
**Affects:** Step 8; retroactively updates fence marks to carry `ConstructKind: "fence"`

### DD-43: Hut footprint unmark — whole-footprint removal via LineID
**Context:** Hut placement uses a fixed 5×5 footprint (16 perimeter marks sharing a LineID). If the player mis-places a hut, they need a way to remove it. Options: (A) Tab toggles to unmark mode, `p` on any tile removes all marks with that LineID (whole footprint); (B) no unmark, defer; (C) Esc undoes last placement.
**Decision:** Option A. Tab toggles mark/unmark mode (same as fence pattern). In unmark mode, pressing `p` on any tile that's part of a hut footprint removes all marks sharing that tile's LineID. Player can see which tiles are marked via olive highlighting during placement mode.
**Rationale:** Follows **Follow the Existing Shape** — fence uses Tab toggle for mark/unmark. Whole-footprint removal via LineID is consistent with how LineID already groups related marks. Per-tile unmark would leave partial footprints, which are structurally invalid for huts.
**Affects:** Step 8b

### DD-44: Hut feasibility follows fence pattern — any material on map
**Context:** Should hut feasibility check for sufficient quantity (12 per tile × 16 tiles = 192 total) or just any material exists? Fence feasibility uses a simple boolean: any grass/stick/brick on map = feasible.
**Decision:** Follow fence pattern. `hutMaterialExistsOnMap` = any construction material on map. No quantity threshold. Order becomes unfulfillable naturally when the character can't find enough materials during execution.
**Rationale:** Follows **Follow the Existing Shape**. Quantity checking at feasibility time is fragile (materials consumed between check and execution). The fence pattern handles this gracefully — feasibility is a rough gate, not an exact resource accounting.
**Affects:** Step 8b

### DD-45: Details panel labels include ConstructKind for all mark types
**Context:** Current fence marks show "Marked for construction" (no kind) or "Marked for construction (Stick)" (material only). With huts sharing the pool, the label needs to distinguish construct types.
**Decision:** Use ConstructKind in labels for all marks. Format: "Marked for construction (Kind)" when no material, "Marked for construction (Material Kind)" when material is stamped. Examples: "Marked for construction (Fence)", "Marked for construction (Stick Fence)", "Marked for construction (Hut)", "Marked for construction (Stick Hut)". Capitalize Kind. Title-case Material.
**Rationale:** Follows **Isomorphism** — the mark carries its kind, the label reflects it. Consistent treatment for all construct types.
**Affects:** Step 8b

### DD-46: Hut placement overlap rules — shared walls, fence overwrite, interior clear
**Context:** How should hut footprint placement interact with existing construction marks? Options: (A) all marks block, (B) marks don't block at all, (C) nuanced per-tile rules.
**Decision:** Option C. Perimeter tiles can overlap existing marks: hut marks use first-wins (shared walls, original LineID preserved), fence marks are overwritten with new hut mark. Interior tiles cannot overlap any existing marks (walls inside a hut's open space are invalid). On placement: perimeter tiles that land on fence marks overwrite them; interior tiles that land on any marks cause the placement to be rejected (invalid).
**Rationale:** Shared hut walls are architecturally natural and material-efficient. First-wins on hut marks means unmarking hut B doesn't remove shared walls that belong to hut A. Fence overwrite is acceptable because the hut wall serves the same blocking function. Interior blocking prevents structurally invalid configurations. Fence placement continues to block on all marks (including hut marks) — the asymmetry is intentional: walls replace fences, but fences don't replace walls.
**Affects:** Step 8b

### DD-47: Items, plants, and tilled soil don't block hut placement
**Context:** Should items, growing plants, and tilled tiles block hut footprint placement? Fence placement already allows items/plants (they're displaced/destroyed during construction). Tilled soil is terrain state.
**Decision:** None of these block placement. Items are displaced during construction, plants are destroyed, tilled soil reverts to normal ground. The hut validator only blocks on: water, built constructs, impassable features, map edges, and marks in interior positions.
**Rationale:** Follows **Follow the Existing Shape** — fence and tilling placement already allow items/plants. Construction is a destructive terrain operation that supersedes lighter uses of the land.
**Affects:** Step 8b

### DD-48: Fence marks render in grey during hut placement; amber when covered by preview
**Context:** During hut placement mode, unbuilt fence marks are visible on the map. They should be visually distinct from hut marks and respond to the footprint preview.
**Decision:** Fence marks render in grey (distinct from olive hut marks) during hut placement mode. When the footprint preview overlaps fence marks, they show in amber (constructionSelectStyle) to indicate they'll be overwritten. This reuses the existing preview highlighting approach — the footprint preview is always rendered on top of confirmed marks.
**Rationale:** Visual distinction helps the player understand what they're overwriting. Follows **Isomorphism** — different mark kinds look different.
**Affects:** Step 8b

### DD-49: findBuildFenceIntent must filter by ConstructKind
**Context:** `findBuildFenceIntent` iterates `MarkedForConstructionPositions()` to find unbuilt tiles. With hut marks now sharing the pool, it would pick up hut-marked positions and try to build fences on them.
**Decision:** Add a `ConstructKind` filter inside `findBuildFenceIntent`'s candidate loop. After getting each position, call `GetConstructionMark(pos)` and skip marks where `ConstructKind != "fence"`. Same pattern needed for `findBuildHutIntent` in Step 10 (filter for `"hut"`).
**Rationale:** Follows **Isomorphism** — marks carry their kind, workers respect it. Follows **Follow the Existing Shape** — extends the existing candidate-filtering loop rather than adding a new API method.
**Affects:** Step 8b (fence side), Step 10 (hut side)

### DD-42: Hut wall segments use heavy box-drawing characters; symbols computed from adjacency at render time
**Context:** Hut walls need visual distinction from fences (`╬`). Requirements call for a 5×5 enclosure with one door. Options for wall rendering: (A) single character for all walls (like fences), (B) line-drawing Unicode characters with corners and edges. Character set options: thin (`┌─┐│└┘`), heavy (`┏━┓┃┗┛`), double (`╔═╗║╚╝`), rounded (`╭─╮│╰╯`).
**Decision:** Heavy box-drawing characters. The symbol set: corners (`┏ ┓ ┗ ┛`), horizontal edges (`━`), vertical edges (`┃`), door (`▯`), T-junctions (`┳ ┻ ┣ ┫`), cross (`╋`). The specific symbol for each wall tile is **computed at render time** from adjacent hut constructs — the renderer checks which cardinal neighbors are also hut constructs and selects the appropriate box-drawing character (corner if two perpendicular neighbors, edge if one axis, T-junction if three neighbors, cross if four). This means junction symbols emerge naturally when two huts share a wall — no special build-time logic needed. Door is at the center of the south wall (bottom edge, middle tile of 5-wide wall). Door construct has `Passable: true`. Door rotation/position choice deferred to Post-Con-Reqs item 5. Rendering uses asymmetric horizontal fill: left corners get fill on right only (` ┏━`), right corners get fill on left only (`━┓ `), edges and doors get fill on both sides (`━━━`, `━▯━`). T-junctions follow the same logic: T-left (`━┫ `) gets fill on left, T-right (` ┣━`) gets fill on right, T-down (`━┳━`) and T-up (`━┻━`) get fill on both sides. Cross (`━╋━`) gets fill on both sides.
**Rationale:** Heavy lines create a clean, substantial enclosure — visually distinct from fence `╬` while reading as a solid building. Evaluated via demo program (`cmd/hutdemo`). Render-time adjacency computation keeps constructs simple (no fine-grained role storage) and handles junctions automatically when huts share walls. Follows **Isomorphism** — the visual reflects what's actually built around a tile, not what was planned.
**Affects:** Step 9, Step 10 (refactor from stored WallRole to adjacency-based rendering)

### DD-50: Hut walls and doors share Kind "hut" — WallRole is "wall" or "door"
**Context:** Hut constructs need a Kind for `PreferenceKind()` matching and future recipe selection. Options: (A) Kind `"hut wall"` and `"hut door"` — precise but creates two preference identities for one building type, (B) Kind `"hut"` for both — WallRole carries the wall/door distinction.
**Decision:** Kind `"hut"` for both walls and doors. `WallRole string` field on Construct carries only the semantic distinction: `"wall"` or `"door"`. WallRole drives `DisplayName()` ("Stick Hut Wall" vs "Stick Hut Door") and `Passable` (door = passable, wall = impassable). The visual symbol (which box-drawing character to render) is **not** determined by WallRole — it is computed at render time from adjacent constructs (DD-42). `PreferenceKind()` returns `"stick hut"` regardless of which tile a character looks at.
**Rationale:** A character looking at any part of a hut should form an opinion about huts of that material, not hut walls vs hut doors separately. Kind `"hut"` maps cleanly to recipe `"stick-hut"` — same 1:1 pattern as `"fence"` → `"stick-fence"`. WallRole is semantic (door vs wall for passability and display name), not visual (symbol selection). Follows **Semantic Precision** — Kind represents what the player cares about (hut identity), WallRole represents the functional distinction (can you walk through it?). Visual rendering is the renderer's job, derived from what's actually built (DD-42).
**Affects:** Step 9, Step 10 (simplified from fine-grained roles to wall/door)

### DD-51: ConstructionMark carries WallRole "wall" or "door" for hut marks
**Context:** When a character builds a hut tile, they need to know whether to create a wall (impassable) or a door (passable). The door position is deterministic (center of south wall), but marks are removed as tiles are built, making it impossible to reliably reconstruct the footprint geometry at build time. Options: (A) store WallRole on the mark at placement time, (B) derive from footprint geometry at build time.
**Decision:** Add `WallRole string` field to `ConstructionMark`. For hut marks, set to `"wall"` or `"door"` at placement time — the placement code already knows the footprint geometry and which tile is the door. For fence marks, WallRole is empty for now (future: fences may have a "gate" role serving the same passable-opening purpose as a hut door). At build time, the character reads WallRole from the mark to create the construct with correct `Passable` value. The visual symbol is not derived from WallRole — it's computed at render time from adjacency (DD-42).
**Rationale:** Placement time is when the footprint geometry is fully known. Storing the semantic role on the mark avoids fragile runtime derivation from partially-built footprints. Follows **Follow the Existing Shape** — marks already carry per-tile metadata (LineID, Material, ConstructKind). Follows **Source of Truth Clarity** — the mark carries its own identity.
**Affects:** Step 10; retroactive: Step 8b placement code sets WallRole on hut marks

### DD-52: Preference-weighted item seeking replaces pure-nearest selection
**Context:** Construction material selection (`selectConstructionMaterial`), craft recipe selection (`findCraftIntent`), item procurement (`EnsureHasItem`, `EnsureHasRecipeInputs`), vessel selection (`EnsureHasVesselFor`, `FindAvailableVessel`), and plantable selection (`EnsureHasPlantable` before variety lock) all use pure nearest-distance. Characters ignore their preferences when seeking materials. Architecture.md's "Future: Unified Item-Seeking" identifies this gap: foraging.go has preference-weighted scoring, but picking.go doesn't.
**Decision:** Score preferences against a combined synthetic output that merges recipe identity with actual material attributes — the anticipated result the character will experience. For construction recipes, build a synthetic Construct (`Kind` from recipe `Output.ItemType`, `Material` from recipe input, `MaterialColor` from nearest material item). For craft recipes, build a synthetic Item with a Material field (DD-61). Within the score, weight attribute matches: Kind contributes 2 points per valence, all other attributes (Material/ItemType, Color) contribute 1 point. Final score: `weightedPref × prefWeight - dist × distWeight`. Config constants `ItemSeekPrefWeight` (20.0) and `ItemSeekDistWeight` (1.0). The weighted scoring functions (`ScoreConstructPreference`, future `ScoreItemPreference`) are separate from `NetPreferenceForConstruct` / `NetPreference` — the 2× Kind weight is a decision-making rule for item seeking, not a general preference strength. Mood changes from looking at constructs/items stay unweighted. For item procurement (no recipe context), vessel selection, and plantable selection: score = `netPref × prefWeight - dist × distWeight` using unweighted preference score.
**Rationale:** Preference shapes material culture — a character who likes stick fences should build stick fences even when grass is closer. The anticipated output carries the full identity the character will experience, so combo preferences (Kind + Color) fire correctly. Weighted Kind ensures recipe identity outweighs material identity (conflict: "likes brick fences" at +2 outweighs "dislikes bricks" at -1 → net +1, character still chooses bricks). Follows **Follow the Existing Shape** (food scoring pattern), **Isomorphism** (score what the character will see, personality expressed through choices), **Source of Truth Clarity** (one synthetic represents the full decision).
**Changed:** Originally a two-term formula scoring recipe and component preferences separately against incomplete synthetic Items. Changed because combo preferences including Color (35% of construct preference formations) scored zero against synthetics missing Color. Also, `Preference.Matches` maps `p.ItemType` to `item.ItemType`, but construct preferences use `p.ItemType` for Material — synthetic Items caused material preferences to go silent. The combined synthetic approach uses `MatchesConstruct` (for construction) which handles this mapping correctly, and embeds the 2:1 weighting in attribute weights rather than formula terms.
**Affects:** Step 11

### DD-53: Redundant color suppression in item display names
**Context:** Sticks always display as "brown stick" and tall grass as "pale green tall grass" because `Description()` builds Color + ItemType/Kind. The color is redundant when all items of that type share the same color — it adds visual noise without information.
**Decision:** Set `Name` field on stick items ("stick") and suppress color for tall grass in `Description()`. Same pattern as brick (`Name: "brick"`) and clay (`Name: "lump of clay"`). Validate that bricks still display correctly after save/load.
**Rationale:** Follows **Semantic Precision** — display names should convey distinguishing information. Color is only meaningful when multiple colors exist for the same type (berries, mushrooms).
**Affects:** Step 12

### DD-54: "Gather lumps of clay" pluralization
**Context:** `Pluralize("clay")` falls through to default "s" suffix, producing "Gather clays" in order display. "Clays" is awkward; clay is used as a mass noun elsewhere ("Dig Clay").
**Decision:** Add "clay" → "lumps of clay" case to `Pluralize()`. Matches the item's Name field ("lump of clay") and differentiates from "Dig Clay" which is a different action.
**Rationale:** Follows **Follow the Existing Shape** — Pluralize already has special cases for berry/mushroom/flower/tall grass. Follows **Isomorphism** — order label matches what the character actually picks up.
**Affects:** Step 12

### DD-55: Abandoned order cooldown timer in orders panel
**Context:** Abandoned orders show status text but no indication of when they'll become available again. The `AbandonCooldown` field exists on Order but isn't surfaced to the player.
**Decision:** Show remaining cooldown as "Abandoned (M:SS)" in the orders panel when an order is in cooldown. When cooldown expires, order returns to "Open" status as before.
**Rationale:** Gives the player visibility into when an order will retry. Follows **Anchor to Intent** — the player wants to know "when will this resume?"
**Affects:** Step 12

### DD-56: Construction order menu labels add "Build " prefix
**Context:** Construction orders display as just "Fence" / "Hut" in the order menu. Other categories prefix a verb: "Craft vessel", "Harvest berries". Construction orders are missing their verb.
**Decision:** Add "Build " prefix for construction category in `DisplayName()`, producing "Build fence" / "Build hut". Lowercase after verb, matching existing craft pattern ("Craft vessel", "Craft hoe").
**Rationale:** Follows **Follow the Existing Shape** — craft category already uses verb + lowercase activity name. Consistent verb prefix across order categories.
**Affects:** Step 12

### DD-57: Discovery randomization — coin flip on activity-vs-recipe order
**Context:** `TryDiscoverKnowHow` always tries activity discovery before recipe discovery. When both share the same trigger (e.g., ActionLook on clay triggers both "dig" and "craftBrick"), the activity always wins deterministically.
**Decision:** Randomize the order: 50% chance to try activity first, 50% chance to try recipe first. Preserves the "only one discovery per interaction" limit while removing fixed ordering.
**Rationale:** Characters should discover capabilities in varied order — deterministic discovery feels artificial. Follows **Start With the Simpler Rule** — coin flip is the simplest randomization.
**Affects:** Step 12

### DD-58: Thatch color evaluation via demo app
**Context:** ColorPaleYellow (ANSI 229) may be too washed out for thatch fences/huts. Thatch constructs make the color more prominent than it was for loose grass bundles.
**Decision:** Build a small demo app comparing ANSI 229 against candidates (178, 179, 186) for dried grass items and thatch constructs side by side. User selects preferred color.
**Rationale:** Color is a visual judgment best made by seeing the options in context, not decided theoretically.
**Affects:** Step 12

### DD-59: Construction recipes carry Output.Kind for unified preference matching
**Context:** Craft recipes have `Output.Kind` (e.g., "shell hoe", "hollow gourd") which is directly matchable against the character's `Preference.Kind` field. Construction recipes don't — their outputs are `RecipeOutput{ItemType: "fence"}` with no Kind. The construct's preference identity ("stick fence") is computed at runtime by `PreferenceKind()` from Material + Kind, but this isn't available during recipe selection because no construct exists yet.
**Decision:** Add `Kind` to construction recipe outputs: `"stick-fence"` gets `Output.Kind: "stick fence"`, `"thatch-fence"` gets `Output.Kind: "thatch fence"`, etc. Same for hut recipes. For construction recipe scoring, the synthetic Construct derives `PreferenceKind()` from its Material + Kind fields (DD-52), so `Output.Kind` is not directly used in that path. `Output.Kind` aligns construction recipes with craft recipes for consistency and enables future scoring paths that may reference it.
**Rationale:** Every recipe should carry its own preference identity rather than requiring runtime derivation from a not-yet-built construct. Follows **Source of Truth Clarity** — the recipe declares what it produces. Follows **Consider Extensibility** — future recipe types (clay pot vessels, etc.) get preference matching with no new plumbing. Follows **Follow the Existing Shape** — craft recipes already have Output.Kind; construction recipes are brought into alignment.
**Changed:** Usage mechanism updated. Originally stated scoring used `NetPreference` against a synthetic Item built from Output. Construction recipes now use synthetic Constructs (DD-52) which derive preference identity from Material + Kind, not from Output.Kind directly.
**Affects:** Step 11

### DD-61: Material field on Item for craft recipe preference matching
**Context:** Construction recipes produce Constructs which have a Material field — `MatchesConstruct` maps `p.ItemType` to `c.Material`, so "likes gourds" (ItemType: "gourd") matches a gourd-material construct. Craft recipes produce Items which have no Material field — `Matches` maps `p.ItemType` to `item.ItemType`, so "likes gourds" cannot match a crafted item with ItemType: "vessel" (the product category). Material preferences are invisible during craft recipe scoring when using combined synthetic Items (DD-52).
**Decision:** Add a `Material string` field to Item. Crafted items carry their primary input material's ItemType (e.g., a hollow gourd vessel has Material: "gourd"). Update `Preference.Matches` to check `p.ItemType` against both `item.ItemType` and `item.Material`, so material preferences cross-apply to crafted products.
**Rationale:** Follows **Isomorphism** (a hollow gourd IS made from a gourd; that identity should be matchable). Follows **Follow the Existing Shape** (mirrors Construct.Material). Enables combined synthetic scoring for craft recipes using the same weighted attribute approach as construction recipes (DD-52).
**Affects:** Step 11b (deferred from 11a; construction recipes use the Construct scoring path which already has Material)

### DD-62: Entity constructor pattern for item types
**Context:** Items like hoe and brick have entity constructors (`NewHoe`, `NewBrick` in `entity/item.go`) that define inherent identity — symbol, type, kind, and now material. Crafting creation functions (`CreateHoe`, `CreateBrick` in `system/crafting.go`) delegate to these constructors and add crafting-specific assembly (color from input, recipe output fields). Vessel is the only item type created inline in its creation function with no entity constructor.
**Decision:** All item types must have entity constructors in `entity/item.go` that define inherent identity (symbol, type, kind, material). Crafting creation functions in `system/crafting.go` delegate to them. Extract `NewVessel` as boy-scout cleanup in Step 11c when adding Material to all constructors.
**Rationale:** Follows **Source of Truth Clarity** — inherent item identity lives in one place (entity constructors), not scattered across crafting code. Follows **Follow the Existing Shape** — hoe and brick already use this pattern; vessel should too. Supports mass-adding craft recipes in later phases without duplicating identity definitions.
**Affects:** Step 11c, future craft recipe additions

