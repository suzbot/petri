# Step Spec: Step 10 — Build Hut Execution

Design doc: [construction-design.md](construction-design.md)

---

## Sub-step 10a: WallRole on ConstructionMark + Placement Update

**Status:** Complete

---

## Sub-step 10b: Render-time Adjacency Symbol Computation

**Status:** Complete

**Anchor story:** The player loads a test world with a completed stick hut. It renders with the same box-drawing characters as before — corners, edges, door — but now the symbols are computed from what's actually built around each tile, not from a stored role. The hut looks identical to before. When two huts share a wall, junction symbols (T-pieces) appear automatically where walls meet.

### Implementation

**1. Simplify WallRole on Construct to "wall"/"door" (DD-50 updated)**

In `internal/entity/construct.go`:
- `NewHutConstruct(x, y, material, materialColor, wallRole)` — wallRole is now `"wall"` or `"door"` only
- Set `Passable: wallRole == "door"` (unchanged)
- Set `Sym` to a default placeholder (e.g., `config.CharHutEdgeH` for walls, `config.CharHutDoor` for doors) — the renderer will override this
- Remove `wallRoleToSymbol` function entirely
- `DisplayName()` logic unchanged — already checks `WallRole == "door"` vs anything else

**2. Add adjacency-based symbol computation in `renderCell()` (DD-42 updated)**

In `internal/ui/view.go`, replace the WallRole-based symbol/fill switch for hut constructs with:

```
func hutSymbolFromAdjacency(pos, gameMap) (symbol rune, leftFill string, rightFill string):
  If construct is a door (Passable == true):
    symbol = CharHutDoor
  Else:
    Check 4 cardinal neighbors for hut constructs (gameMap.ConstructAt)
    hasN, hasS, hasE, hasW = neighbor checks
    Map neighbor pattern to box-drawing symbol:
      {E,S} → ┏    {W,S} → ┓    {E,N} → ┗    {W,N} → ┛
      {W,E} → ━    {N,S} → ┃
      {W,E,S} → ┳  {W,E,N} → ┻  {N,S,E} → ┣  {N,S,W} → ┫
      {N,S,E,W} → ╋
      1 neighbor or 0: fallback to edge (━ or ┃ based on which neighbor exists)

  leftFill = ━ if hasW, else " "
  rightFill = ━ if hasE, else " "
  (Door always gets both fills since it's mid-wall: leftFill = ━, rightFill = ━)
```

The renderer calls this helper instead of switching on WallRole for symbol/fill.

**3. Update save/load deserialization (backward compat)**

In `internal/ui/serialize.go`, construct restoration:
- Old saves have fine-grained WallRole values ("corner-tl", "edge-h", etc.)
- Map all non-"door" values to "wall": `if wallRole != "door" { wallRole = "wall" }`
- New saves will only ever have "wall" or "door"

**4. Update test-world hut creation**

If test-world code creates huts with fine-grained WallRole, simplify to "wall"/"door".

### Tests

- Unit test: `hutSymbolFromAdjacency` returns correct symbols for each neighbor pattern (all 11 combinations: 4 corners, 2 edges, 4 T-junctions, 1 cross)
- Unit test: door always returns ▯ regardless of neighbors
- Unit test: fill logic — leftFill/rightFill match expected values for each pattern
- Unit test: old save with fine-grained WallRole loads correctly (mapped to "wall"/"door")

### [TEST] Checkpoint

Load a test world with a pre-placed completed hut. Verify rendering is identical to before (correct box-drawing, fills, colors). Create a second test world with two huts sharing a wall — verify T-junction symbols appear at the shared tiles automatically.

### [DOCS]
Run `/update-docs` after human testing confirms success.

### [RETRO]
Run `/retro` after documentation is updated.

---

## Sub-step 10c: Build Hut Intent + Handler (Full Execution)

**Anchor story:** A character who knows how to build huts takes a Build Hut order. They choose stick material (nearest available), and the material is stamped across all 16 hut marks. They walk to the nearest bundle of sticks, pick it up, carry it to the nearest unbuilt wall tile, and drop it. They return for another bundle, drop it at the same tile — now 2 full bundles sit on the tile. They work the tile, consuming both bundles, and a stick wall segment appears. They move to the next tile and repeat. Another character joins and works from the other end. Eventually all 16 tiles are built — walls and a door. Characters walk through the door into the 3×3 interior.

### Implementation

**1. Generalize material selection**

In `internal/system/order_execution.go`:
- Rename `selectFenceMaterial` → `selectConstructionMaterial`
- Add `activityID string` parameter, replacing hardcoded `"buildFence"`
- Uses `char.GetKnownRecipesForActivity(activityID)` instead of hardcoded activity
- Update the existing `findBuildFenceIntent` call site to pass `"buildFence"`

**2. New helpers for supply-drop counting**

In `internal/system/order_execution.go`:
- `countFullBundlesAtPosition(items, itemType, pos) int` — counts items at position where `BundleCount >= MaxBundleSize` for that type
- `countSuppliesAtPosition(items, material, pos, recipe) int` — unified check:
  - For bundleable materials: returns `countFullBundlesAtPosition(...)` (threshold: 2 bundles per tile)
  - For non-bundleable (brick): returns `countItemsAtPosition(...)` (threshold: 12 per tile)
- `suppliesMetForHutTile(items, material, pos) bool` — checks if the tile has enough supplies to build (bundles: 2 full bundles; bricks: 12 items). Uses recipe `Inputs[0].Count` to derive thresholds: for bundleable materials, count = recipe count / MaxBundleSize (12/6 = 2 bundles); for bricks, count = recipe count (12).

**3. `findBuildHutIntent` — follows fence pattern with supply-drop for all materials**

In `internal/system/order_execution.go`:

Flow mirrors `findBuildFenceIntent` (DD-49 filter for `ConstructKind == "hut"`):

1. Collect unbuilt hut-marked candidates (no construct at pos, no character at pos)
2. Find nearest candidate
3. Determine material: if mark.Material empty → `selectConstructionMaterial(char, pos, items, gameMap, "buildHut")` → `SetLineMaterial(mark.LineID, material)`. Return nil if no material available.
4. Drop non-material inventory items (procurement drop pattern)
5. Check if tile has enough supplies (`suppliesMetForHutTile`):
   - Yes → go to build phase (`findHutBuildIntent` — find adjacent standing tile, create ActionBuildHut intent with TargetBuildPos)
   - No → go to supply phase:
     - **Bundle materials**: if character has a full bundle → deliver to tile (move to tile, intent for delivery). If no full bundle → gather (same as fence: find nearest bundle or individual item, create pickup intent)
     - **Brick materials**: same pattern as `findBrickFenceIntent` but threshold 12 instead of 6. If character has bricks → deliver. Otherwise → find nearest brick not on a construction-marked tile → pickup intent.

Key differences from fence intent finder:
- Filters by `ConstructKind == "hut"` (not "fence")
- Uses `selectConstructionMaterial("buildHut")` for material selection
- Supply-drop for bundle materials (fences consume bundles from inventory; huts drop bundles at site then consume from ground)
- Threshold: 2 full bundles or 12 bricks per tile (vs 1 bundle or 6 bricks for fences)

**4. `applyBuildHut` handler**

In `internal/ui/apply_actions.go`:

Structure mirrors `applyBuildFence`:

1. **Validate**: check `TargetBuildPos != nil`
2. **Delivery phase**: if at build position and has material in inventory → drop all matching items at position, clear intent, return
3. **Walking phase**: if not at destination → `moveWithCollision()`
4. **Arrival collision check** (DD-28 layer 2): if another character now occupies build tile → clear intent, return
5. **Working phase**: accumulate `ActionProgress` over `config.ActionDurationMedium`
6. **Build action**:
   - Read material from mark (`GetConstructionMark(buildPos)`)
   - Read WallRole from mark (DD-51)
   - **Bundle consumption from ground**: find 2 items at buildPos with `BundleCount >= MaxBundleSize` for that material, remove them from world
   - **Brick consumption from ground**: find and remove 12 items of material type at buildPos
   - If insufficient materials at position: clear intent, return (re-evaluate next tick)
   - Material color mapping: grass → ColorPaleYellow, stick → ColorBrown, brick → ColorTerracotta
   - Create construct: `entity.NewHutConstruct(buildPos.X, buildPos.Y, material, materialColor, wallRole)`
   - `gameMap.AddConstruct(construct)`
   - `gameMap.UnmarkForConstruction(buildPos)`
7. **Character displacement** (DD-28 layer 3): if character occupies build tile after construct placement → displace to nearest cardinal empty tile
8. **Item displacement** (DD-33): displace remaining items at build position to adjacent tiles
9. **Log and clear**: log "Built [material] hut [wall/door]", clear intent

**5. Wire up dispatch**

In `internal/system/order_execution.go`:
- Replace `return nil` stub in `findOrderIntent` switch case `"buildHut"` with `return findBuildHutIntent(...)`

In `internal/ui/apply_actions.go`:
- Add `case entity.ActionBuildHut:` to `applyIntent` dispatch, calling `applyBuildHut`

In `internal/entity/intent.go` (or wherever action constants live):
- Add `ActionBuildHut` constant

### Architecture Patterns

- **Component Procurement** (architecture.md): supply-drop is a variant — instead of "ensure I have the item," it's "ensure the tile has the items"
- **Marked-for-Construction Pool** (architecture.md): filters by ConstructKind "hut", uses LineID for material lock
- **Collision handling** (DD-28): skip occupied tiles in candidate list → abandon if occupied mid-build → displace as safety net
- **Item displacement** (DD-33): same pattern as fence — displace items from build position after construct placement
- **No tile claiming** (DD-29): workers independently target nearest unbuilt tile; overlap is benign
- **Sticky BFS**: reuse `char.UsingBFS` for position-based movement

### Requirements Reconciliation

- Construction-Reqs.txt line 79: "2 full bundles per tile x 16 tiles" → per-tile cost derives from recipe Input Count (12) divided by MaxBundleSize (6) = 2 bundles
- Construction-Reqs.txt line 81: "12 bricks per tile x 16 tiles" → per-tile cost is recipe Input Count (12)
- Construction-Reqs.txt line 82: "marked tiles for construction get all their supplies dropped on them and then each tile with all its supplies is worked" → supply-drop for all materials, then build
- Construction-Reqs.txt line 87: "order becomes unavailable or unfulfillable when materials run out" → `IsOrderFeasible` already checks `HasUnbuiltConstructionPositions("hut") && constructionMaterialExistsOnMap(items)` (stubbed in Step 8)

### Values Alignment

- **Follow the Existing Shape**: mirrors fence execution flow — same phases (candidate selection → material lock → procurement → build → displace), same helpers, same collision layers
- **Isomorphism**: supply-drop is visible — characters deliver materials to a build site, creating a visible stockpile before construction
- **Start With the Simpler Rule**: no tile claiming, no complex scheduling — workers independently pick nearest tile

### Tests

- Unit test: `selectConstructionMaterial` returns correct material for both fence and hut activities (existing fence behavior preserved)
- Unit test: `countFullBundlesAtPosition` counts only full bundles at the right position
- Unit test: `suppliesMetForHutTile` correctly checks thresholds for bundle vs brick materials
- Unit test: `findBuildHutIntent` returns nil when no unbuilt hut marks exist
- Unit test: `findBuildHutIntent` filters by ConstructKind "hut" (ignores fence marks)
- Unit test: material stamping via LineID propagates to all hut marks in the line

### [TEST] Checkpoint

**Test 1 — Bundle material (sticks):** Create a test world with hut marks placed, a character who knows buildHut + stick-hut recipe, and plenty of stick bundles on the ground. Assign Build Hut order. Watch character select material, deliver bundles to tiles, build walls. Verify:
- Material stamped across all 16 marks
- Character drops bundles at tile, then builds
- Correct wall/door symbols render (adjacency-based from 10b)
- Door tile is passable, wall tiles are not
- Details panel shows "Stick Hut Wall" / "Stick Hut Door"
- Order completes when all 16 tiles are built

**Test 2 — Brick material:** Same setup but with bricks and brick-hut recipe. Verify multi-trip delivery (2 bricks per trip, 6 trips per tile), then build.

**Test 3 — Multiple workers:** Two characters with Build Hut orders. Verify both work on the same hut, building from different tiles, no conflicts.

**Test 4 — Feasibility:** Remove all construction materials from the world. Verify the order becomes unfulfillable. Add materials back — verify it becomes feasible again.

### [DOCS]
Run `/update-docs` after human testing confirms success.

### [RETRO]
Run `/retro` after documentation is updated.

---

## Sub-step 10d: Multi-Hut Junction Test

**Anchor story:** The player places two huts sharing a wall. Both are built. Where the walls meet, T-junction symbols appear automatically — the shared tiles show the correct box-drawing characters reflecting connections in three directions. A cross symbol appears if walls connect in all four directions.

### Implementation

No new code expected — junction rendering should work from 10b's adjacency computation. This sub-step is a targeted test of the multi-hut scenario.

If junctions don't render correctly, investigate and fix the adjacency computation in `renderCell()`.

### [TEST] Checkpoint

Create a test world with two completed huts sharing a wall (e.g., hut A at (5,5) and hut B at (9,5) — sharing the column at x=9). Verify:
- Shared wall tiles show T-junction symbols (┳ ┻ ┣ ┫) where appropriate
- Corner tiles at the junction show correct symbols
- Fill characters are correct for junctions (T-left gets left fill, T-right gets right fill, etc.)
- Details panel shows correct display names for junction tiles
- Both huts' doors are passable and non-shared

If time permits, test a third hut creating a cross junction (╋).

### [DOCS]
Run `/update-docs` after human testing confirms success.

### [RETRO]
Run `/retro` after documentation is updated.
