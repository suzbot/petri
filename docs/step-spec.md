# Step Spec: Step 6 — Build Fence

Design doc: [construction-design.md](construction-design.md)

---

## Sub-step 6a: Infrastructure + Marking UI + Discovery

**Anchor story:** The player opens the orders menu and sees "Construction" as a new category. They select Construction > Fence and enter a line-marking mode. They press `p` to anchor the start of a fence line, move the cursor horizontally, and press `p` again — a row of tiles lights up as "Marked for construction." They move to another spot and draw a vertical line. They press Enter and a "Build Fence" order appears in the order list. Selecting a marked tile in the details panel shows "Marked for construction." Meanwhile, a character in a good mood looks at some bricks and has an insight — "Learned Brick Fence recipe!" and "Discovered how to build fences!" appear in their action log. The knowledge panel shows "Build Fence" under Know-how and "Brick Fence" under Recipes.

### Scope

**Data layer:**
- `ConstructionMark` struct: `LineID int`, `Material string` (empty until first tile built)
- `markedForConstruction map[types.Position]ConstructionMark` on Map with methods:
  - `MarkForConstruction(pos, lineID) bool` — returns false if already marked or has construct
  - `UnmarkForConstruction(pos)`
  - `IsMarkedForConstruction(pos) bool`
  - `GetConstructionMark(pos) (ConstructionMark, bool)`
  - `MarkedForConstructionPositions() []types.Position`
  - `SetLineMaterial(lineID int, material string)` — stamps material on all marks with matching LineID
  - `HasUnbuiltConstructionPositions() bool` — any marked position without a construct
  - `nextConstructionLineID int` with `NextConstructionLineID() int` (increments and returns)
- Serialization: `ConstructionMarkSave` struct in `save/state.go`, serialize/deserialize in `serialize.go` (LineID + Material + Position)

**Entity layer:**
- `ActionBuildFence` constant in `character.go`
- `TargetBuildPos *types.Position` field on Intent struct (ephemeral, not serialized — follows `TargetWaterPos` pattern)
- "buildFence" activity in ActivityRegistry:
  - Category: "construction"
  - IntentFormation: IntentOrderable
  - Availability: AvailabilityKnowHow
  - No direct DiscoveryTriggers (discovered via recipes, like craftVessel/craftHoe/craftBrick)
- Three fence recipes in RecipeRegistry:
  - `"thatch-fence"`: ActivityID "buildFence", Name "Thatch Fence", Inputs: [{ItemType: "grass", Count: 6}], DiscoveryTriggers: [{Action: ActionLook, ItemType: "grass"}, {Action: ActionPickup, ItemType: "grass"}]
  - `"stick-fence"`: ActivityID "buildFence", Name "Stick Fence", Inputs: [{ItemType: "stick", Count: 6}], DiscoveryTriggers: [{Action: ActionLook, ItemType: "stick"}, {Action: ActionPickup, ItemType: "stick"}]
  - `"brick-fence"`: ActivityID "buildFence", Name "Brick Fence", Inputs: [{ItemType: "brick", Count: 6}], DiscoveryTriggers: [{Action: ActionLook, ItemType: "brick"}, {Action: ActionPickup, ItemType: "brick"}]
  - Output field: `{ItemType: "fence"}` (for display; actual output is a construct, handled by custom handler)
  - Duration: not used by fence handler (fence has its own duration in applyBuildFence)
  - Repeatable: false (fence completion is tile-pool-based, not recipe-loop-based)
- Discovery log message fix: `tryDiscoverRecipe` currently logs "Discovered how to craft %s!" — change to use category-aware wording. If the activity's category is "construction", log "Discovered how to build fences!" instead of "Discovered how to craft Fence!"

**UI layer — Line marking:**
- Line-drawing mode activated when player selects "Fence" under "Construction" category (step 2 with `step2ActivityID = "buildFence"`)
- `isValidFenceTarget(pos, gameMap)` validator in `area_select.go`: rejects water, impassable features, existing constructs, already-marked-for-construction tiles
- `isValidUnmarkFenceTarget(pos, gameMap)` validator: returns true for marked-for-construction tiles without a built construct
- `getValidLinePositions(anchor, cursor, gameMap, validator)` in `area_select.go`: constrains to cardinal line (axis with larger delta wins for diagonal cursor movement). Returns positions along the line that pass the validator.
- `p` key handling in update.go step 2 for buildFence:
  - First `p`: set anchor at cursor position
  - Second `p`: get valid line positions from anchor to cursor, mark all with shared LineID from `gameMap.NextConstructionLineID()`. If in unmark mode, unmark instead. Clear anchor (stay in step 2 for next line).
  - Single-tile: anchor and confirm on same tile is the degenerate case (line of length 1)
- Tab toggles mark/unmark mode (reuses `areaSelectUnmarkMode`)
- Enter creates order: `entity.NewOrder(nextOrderID, "buildFence", "")` — TargetType is empty (material determined by character at build time)

**UI layer — Rendering:**
- Marked-for-construction tiles get a highlight style (distinct from tilling's sage/teal — use a construction-appropriate color)
- Active line preview during anchor-to-cursor drawing
- Details panel at cursor:
  - Marked tile with no material: "Marked for construction"
  - Marked tile with material stamped: "Marked for construction (Stick)"
  - Construct already built: existing construct display (from Step 5)

**Order menu flow:**
- `getOrderableActivities` already generates synthetic category entries from Activity.Category — adding `Category: "construction"` makes "Construction" appear automatically when any character knows buildFence
- `getCategoryActivities("construction")` returns buildFence activity
- Step 0: player sees "Construction" category → step 1
- Step 1: player sees "Fence" under Construction → step 2 (line marking mode)
- Step 2: line marking → Enter creates order

### Architecture patterns
- **Marked-tile pool**: follows `markedForTilling` pattern (Map field, methods, serialization). Extended with `ConstructionMark` struct for LineID and Material.
- **Area selection reuse**: follows area_select.go pattern with line-constrained validator instead of rectangle.
- **Recipe-based discovery**: follows existing `tryDiscoverRecipe` which auto-grants activity know-how (DD-27).
- **Order menu category**: follows existing synthetic category pattern in `getOrderableActivities` (DD-22).

### Anti-patterns to avoid
- Do NOT route buildFence through `findCraftIntent` — the recipes exist for discovery/display only, execution uses custom handler (DD-27).
- Do NOT add buildFence to `isDiscretionaryAction` — it's an ordered action, not a discretionary activity.
- Do NOT serialize `TargetBuildPos` on Intent — it's ephemeral like displacement fields.

### Tests
- Unit: `MarkForConstruction` / `UnmarkForConstruction` / `SetLineMaterial` round-trip
- Unit: `getValidLinePositions` returns correct cardinal line for horizontal, vertical, and diagonal cursor positions
- Unit: Fence recipe discovery triggers grant buildFence activity know-how
- Unit: Serialization round-trip for ConstructionMark (LineID + Material preserved)
- Integration: `getOrderableActivities` includes "Construction" category when character knows buildFence

### [TEST] Checkpoint
Human testing:
- Create a test world with a character who knows buildFence + all fence recipes
- Verify: Construction > Fence menu flow works (step 0 → 1 → 2)
- Verify: Line marking with `p` (anchor, draw line, confirm), unmark with Tab toggle
- Verify: Marked tiles render with highlight style
- Verify: Details panel shows "Marked for construction"
- Verify: Enter creates "Build Fence" order in order list
- Verify: Marked tiles persist through save/load
- Separately: test world with character in good mood + bricks on ground → verify recipe discovery + know-how auto-grant in action log and knowledge panel

### [DOCS]
Run /update-docs after human testing confirms success.

### [RETRO]
Run /retro after docs are updated.

---

## Sub-step 6b: Stick/Grass Fence Building

**Anchor story:** A character takes the Build Fence order. They spot a full bundle of 6 sticks that someone gathered earlier, pick it up in one step, walk to a tile adjacent to the first marked position, and build a stick fence. The marked tile now shows "Stick" as the material, and all other tiles in the same line also update to show "Stick." For the next tile, no full bundles exist, so they find a partial bundle of 3 sticks on the ground and pick it up, then gather individual sticks until their bundle reaches 6, then build again. When all marked tiles have fences, the order completes.

### Scope

**findOrderIntent restructure (DD-32):**
- Move the recipe check (`GetRecipesForActivity`) from before the switch to inside the `default` case
- All existing activity-specific cases (harvest, tillSoil, plant, waterGarden, gather, extract, dig) remain as switch cases
- Add `case "buildFence": return findBuildFenceIntent(...)`
- Generic craft activities (craftVessel, craftHoe, craftBrick) fall through to default → findCraftIntent
- This restructure is a prerequisite — do it first, verify existing tests still pass

**Pickup bundle merge enhancement (DD-30):**
- In `Pickup()` bundle merge path: change `carried.BundleCount++` to `carried.BundleCount += item.BundleCount`
- Change guard: `carried.BundleCount < maxSize` to `carried.BundleCount + item.BundleCount <= maxSize`
- If merge would overflow, skip (no partial splitting in v1) — character finds a different item
- This is a general improvement; verify existing gather tests still pass after the change

**findBuildFenceIntent (new function in order_execution.go):**

Phase detection flow (checked each tick, position-based ordered action):
1. **Find target tile**: nearest unbuilt marked-for-construction position (no character standing on it per DD-28). If none → return nil (triggers completion check).
2. **Determine material**: read `ConstructionMark.Material` for target tile's LineID. If empty → select material:
   - Get character's known fence recipes (`char.GetKnownRecipesForActivity("buildFence")`)
   - For each known recipe's input ItemType, count available items on map (6+ needed)
   - Pick the material whose nearest item is closest (DD-31)
   - Call `gameMap.SetLineMaterial(lineID, material)` to stamp all marks in the line
   - If no material has 6+ items → return nil (triggers abandonment)
3. **Drop non-material inventory**: drop any inventory items that aren't the target material (procurement drop pattern, follows findDigIntent)
4. **Check procurement complete**: `hasFullBundle(char, material)` for bundle types. If yes → go to build phase.
5. **Procurement phase**: find nearest item of target material on ground:
   - First: search all items for any bundle of the material type (including full bundles — unlike `findNearestItemByType` which skips full bundles). Only search for full/partial bundles when character has no bundle of this type in inventory (avoids the overflow merge issue).
   - Fallback: `findNearestItemByType` for individual items (skips full bundles, which is correct when character already has a partial bundle)
   - Return `ActionPickup` intent for the found item
   - If no items found → return nil (triggers abandonment)
6. **Build phase** (has full bundle):
   - Find cardinally adjacent empty tile to the target build position (for character to stand on)
   - If no adjacent tile available → try next marked tile
   - Return `ActionBuildFence` intent: `Dest` = adjacent standing tile, `TargetBuildPos` = build position
   - Use sticky BFS (`nextStepBFSCore` with `char.UsingBFS`)
   - Set `CurrentActivity` to "Building fence" or "Moving to build fence"

**applyBuildFence handler (new method in apply_actions.go):**

Walking phase:
- If not at `Dest` → `moveWithCollision`

Arrival check (DD-28 layer 2):
- If a character is now standing on `TargetBuildPos` → clear intent, return (re-evaluate next tick will skip this tile)

Working phase:
- Accumulate `ActionProgress` until `ActionDurationMedium`
- On completion:
  - Determine material color: "grass" → ColorPaleYellow, "stick" → ColorBrown, "brick" → ColorTerracotta
  - Consume full bundle from inventory (remove from inventory slot)
  - Create fence: `entity.NewFence(buildPos.X, buildPos.Y, material, materialColor)`
  - Add to map: `gameMap.AddConstruct(fence)`
  - Unmark tile: `gameMap.UnmarkForConstruction(buildPos)`
  - Displace character at buildPos if present (DD-28 layer 3): find empty adjacent tile, move them
  - Displace any items at buildPos to adjacent empty tiles (DD-33)
  - Log: "Built [Material] fence"
  - Clear intent

**applyIntent dispatch:**
- Add `case entity.ActionBuildFence:` → `m.applyBuildFence(char, delta)`

**isMultiStepOrderComplete:**
- Add `case "buildFence":` → `!gameMap.HasUnbuiltConstructionPositions()`

**IsOrderFeasible:**
- Add `case "buildFence":` → unbuilt marked-for-construction positions exist AND at least one fence material type (grass, stick, brick) has items on the map

**Order completion cleanup:**
- In `selectOrderActivity`, after `isMultiStepOrderComplete` returns true for buildFence: drop any remaining fence materials from inventory (follows `DropCompletedBundle` / `DropCompletedDigItems` pattern)

**Discovery log wording:**
- In `tryDiscoverRecipe` (discovery.go): when the activity's category is "construction", use "Discovered how to build fences!" instead of "Discovered how to craft Fence!"

**Order DisplayName:**
- Add case in `DisplayName()` for buildFence: return "Build Fence" (activity name is sufficient — no target type suffix since material is character-chosen)

### Architecture patterns
- **Position-based ordered action**: follows TillSoil/Dig pattern — intent cleared after each work unit, re-evaluated via findBuildFenceIntent each tick. Uses sticky BFS.
- **Component Procurement**: follows findTillSoilIntent pattern (EnsureHasItem for hoe) but with bundle gathering instead of single-item procurement.
- **Marked-tile pool consumption**: follows `UnmarkForTilling` pattern — tile unmarked when work completes.
- **Adjacent-tile interaction**: follows water drinking pattern — `Dest` is standing tile, `TargetBuildPos` is interaction target.
- **findOrderIntent restructure**: switch-first, recipe-fallback-in-default (DD-32). Extensible for future recipe-using activities.
- **No early-return block in continueIntent**: position-based ordered actions use the generic fallthrough path. No new block needed.

### Anti-patterns to avoid
- Do NOT use `findCraftIntent` or `EnsureHasRecipeInputs` for fence procurement — the bundle gathering and supply-drop patterns don't fit the craft procurement model.
- Do NOT set `Target` directly to the destination — always use `nextStepBFSCore` to calculate `Target` as a single BFS step.
- Do NOT search for full bundles when the character already carries a partial bundle of the same type — the `Pickup` merge would only add `BundleCount` and could silently overflow if the guard is wrong.
- Do NOT check `TargetBuildPos` before confirming the character is at `Dest` — the handler must arrive at the standing tile first.

### Tests
- Unit: `findOrderIntent` restructure — verify craftVessel/craftHoe/craftBrick still route to `findCraftIntent` via default case
- Unit: `Pickup` bundle merge — picking up a bundle of 3 when carrying a bundle of 2 results in bundle of 5
- Unit: `Pickup` bundle merge overflow — picking up a bundle of 4 when carrying a bundle of 4 returns PickupFailed (would exceed max 6)
- Unit: `findBuildFenceIntent` returns ActionPickup when character needs materials, ActionBuildFence when character has full bundle
- Unit: `findBuildFenceIntent` skips marked tiles occupied by characters (DD-28)
- Unit: `findBuildFenceIntent` sets line material on first build (DD-25)
- Unit: `applyBuildFence` places fence construct at TargetBuildPos, not at character position
- Unit: `applyBuildFence` unmarks tile and displaces items (DD-33)
- Unit: `isMultiStepOrderComplete` returns true when no unbuilt marked positions remain
- Unit: `IsOrderFeasible` returns false when no materials exist, true when materials + marks exist
- Regression: existing gather/harvest tests still pass after Pickup merge enhancement

### [TEST] Checkpoint
Human testing:
- Create test world with: character who knows buildFence + stick-fence recipe, several sticks and a full bundle of sticks on the ground, tiles marked for construction (from 6a)
- Verify: character picks up full bundle first, walks to adjacent tile, builds stick fence
- Verify: material stamps on all marks in the same line
- Verify: character gathers remaining sticks one-by-one for next tile, builds again
- Verify: picking up partial bundles adds correct count (not just +1)
- Verify: order completes when all marked tiles have fences
- Verify: fence renders correctly (stick color, fence symbol, impassable)
- Verify: details panel shows fence info when selecting built fence
- Verify: items on build tile get displaced to adjacent tiles
- Verify: fence persists through save/load

### [DOCS]
Run /update-docs after human testing confirms success.

### [RETRO]
Run /retro after docs are updated.

---

## Sub-step 6c: Brick Supply-Drop

**Anchor story:** The player has tiles marked for construction where the line material is set to "brick" (either from an adjacent brick fence or determined by nearest-available in a world with bricks). A character takes the order, picks up 2 bricks (filling both inventory slots), walks to the marked tile, and drops them there. They return for more bricks, pick up 2 more, carry them to the same tile, drop them. After three trips, 6 bricks sit on the tile. The character walks to an adjacent tile and builds — the bricks are consumed, a brick fence appears, and any excess bricks on the tile get displaced to neighboring tiles.

### Scope

**Extend findBuildFenceIntent for brick path:**

The brick material path is distinct from bundle materials because bricks aren't bundled (DD-23). The intent finder adds these phases:

1. **After material determination**: if material is "brick" (not in `config.MaxBundleSize`), use brick-specific procurement instead of bundle procurement.
2. **Check build-ready**: count brick items at the target build position. If 6+ → build phase (same as bundle: adjacent tile intent with TargetBuildPos).
3. **Check delivery needed**: if character has bricks in inventory → deliver to build site:
   - Return `ActionBuildFence` intent with `Dest` = build tile position (NOT adjacent — character walks TO the tile to drop)
   - Handler detects delivery mode: character is at the build tile with bricks in inventory → drop all bricks, clear intent
4. **Pickup phase**: if character has no bricks → find nearest brick on map → return `ActionPickup` intent
5. **No bricks available** → return nil (triggers abandonment)

**Extend applyBuildFence for delivery and brick build modes:**

The handler needs three modes, detected by world state:

1. **Delivery mode**: character is at `TargetBuildPos` AND has bricks in inventory → drop all bricks at current position, clear intent. (Dest was set to build tile for delivery.)
2. **Build mode**: character is at `Dest` (adjacent to `TargetBuildPos`) AND 6+ bricks at TargetBuildPos → accumulate progress, consume 6 bricks from the tile, place fence, displace remainder + characters (DD-28, DD-33).
3. **Walking mode**: not yet at destination → `moveWithCollision`.

Mode detection: the handler checks `TargetBuildPos`:
- If `char.Pos() == *TargetBuildPos` and character has bricks → delivery mode
- If `char.Pos() == Dest` and `char.Pos() != *TargetBuildPos` → build mode (character is adjacent)
- Otherwise → walking mode

**Brick consumption from ground:**
- When building with bricks, consume 6 brick items from the build tile (iterate items at position, remove 6 with ItemType "brick")
- Any remaining items at position displaced to adjacent tiles (DD-33)

### Architecture patterns
- **Supply-drop pattern**: novel to this step. Character shuttles materials (2 at a time) to a build site. Intent re-evaluation each tick drives the state machine: pickup → deliver → pickup → deliver → build. This pattern is reused in Step 8 (huts) at larger scale.
- **Delivery mode in handler**: the handler distinguishes delivery from building by checking character position relative to `TargetBuildPos`. Delivery = at build tile with materials; building = adjacent to build tile with materials on ground.

### Anti-patterns to avoid
- Do NOT try to carry all 6 bricks at once — characters have 2 inventory slots. The multi-trip pattern is by design (DD-23).
- Do NOT merge bricks into bundles — bricks are individual items, not bundleable (DD-23). The `config.MaxBundleSize` map has no entry for "brick".
- Do NOT set `Dest` to the adjacent tile during delivery — `Dest` must be the build tile so the character walks TO it to drop materials. Only set `Dest` to adjacent for the build phase.

### Tests
- Unit: `findBuildFenceIntent` with brick material: returns ActionPickup when no bricks in inventory, ActionBuildFence with Dest=buildTile when carrying bricks (delivery), ActionBuildFence with Dest=adjacentTile when 6+ bricks at site (build)
- Unit: `applyBuildFence` delivery mode: drops bricks at position, clears intent
- Unit: `applyBuildFence` brick build mode: consumes 6 bricks from ground, places fence, displaces excess
- Unit: Full brick supply-drop cycle: 3 pickup-deliver trips followed by build
- Unit: Displace safety net — character standing on build tile when fence placed gets moved to adjacent tile (DD-28 layer 3)
- Regression: bundle fence building (6b) still works after handler changes

### [TEST] Checkpoint
Human testing:
- Create test world with: character who knows buildFence + brick-fence recipe, 8+ bricks on the ground, tiles marked for construction
- Verify: character picks up 2 bricks, carries to marked tile, drops them
- Verify: character returns for more bricks, repeats until 6 at build site
- Verify: character moves to adjacent tile and builds brick fence
- Verify: excess bricks (if any) displaced to adjacent tiles, not trapped under fence
- Verify: order completion when all marked tiles built
- Verify: brick fence renders correctly (terracotta color, fence symbol)
- Verify: save/load preserves brick fences

### [DOCS]
Run /update-docs after human testing confirms success.

### [RETRO]
Run /retro after docs are updated.

---

## Implementation Sequencing Notes

### Why this ordering
- **6a first**: establishes all data structures, UI, and discovery without character behavior. Clean test of infrastructure.
- **6b second**: introduces the core building loop with the simpler procurement pattern (bundles). Includes the findOrderIntent restructure (DD-32) and Pickup enhancement (DD-30) as prerequisites since they affect the broader system.
- **6c last**: adds the novel supply-drop pattern (bricks) on top of the working fence system. The handler gains delivery mode but the build mode is already tested.

### Deferred from Step 6
- **Preference-based material selection**: deferred to Step 10 (DD-31). Current: nearest available.
- **Partial bundle splitting on overflow**: v1 skips merges that would overflow. Future: split the ground bundle, take what fits.
- **continueIntent consolidation**: triggered enhancement, evaluate independently of construction (Step 6 adds no new early-return block).
- **Order-aware simulation for e2e testing**: triggered enhancement, not blocking.
- **Category type formalization**: triggered enhancement, not blocking.
