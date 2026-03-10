# Step Spec: Step 11 — Preference-Weighted Item Seeking

Design doc: [construction-design.md](construction-design.md)

---

## Overview

Characters currently ignore their preferences when seeking non-food items. All item acquisition uses pure nearest-distance. After this step, preferences influence which materials characters choose for building, which recipe they select when crafting, which specific item they grab when procuring components, which vessel they pick up, and which plantable variety they choose before variety lock.

One formula, one function, all call sites:

```
score = recipe_netPref × prefWeight × 2 + component_netPref × prefWeight - dist × distWeight
```

- `recipe_netPref`: character's preference for the recipe output (e.g., "stick fence"). Zero when no recipe context.
- `component_netPref`: character's preference for the specific item being considered.
- `dist`: distance to that item.
- `prefWeight`: `ItemSeekPrefWeight` (20.0). One preference point = 20 tiles of walking.
- `distWeight`: `ItemSeekDistWeight` (1.0). Each tile costs 1 point.
- The 2× on recipe preference is hardcoded — "what you're making matters twice as much as what you're making it with."

When there's no recipe context, the recipe term is zero and the formula reduces to:
```
score = component_netPref × prefWeight - dist × distWeight
```

The formula lives in `ScoreItemFit` in `item_scoring.go` (sibling of `food_scoring.go`). Every call site uses this one function.

---

## Sub-step 11a: Scoring Infrastructure + Construction Material Selection

**Anchor story:** A character who likes stick fences and dislikes grass takes a Build Fence order. Thatch grass bundles are 5 tiles away; sticks are 25 tiles away. The character walks past the grass to the sticks — "likes stick fences" (+1 × 20 × 2 = 40) outweighs the 20-tile distance penalty (20 × 1 = 20).

### Changes

**1. `ScoreItemFit` in `internal/system/item_scoring.go`**

New file, sibling of `food_scoring.go`. Single function used by all call sites in this step:

```go
func ScoreItemFit(recipePref, componentPref, dist int) float64 {
    return float64(recipePref)*config.ItemSeekPrefWeight*2 +
        float64(componentPref)*config.ItemSeekPrefWeight -
        float64(dist)*config.ItemSeekDistWeight
}
```

Recipe selection sites pass `recipePref` from `NetPreference` on a synthetic recipe output item. Item procurement sites pass `recipePref = 0`.

**Pattern:** Follows food_scoring.go (ScoreFoodFit). Follow the Existing Shape.

**2. Config constants in `config.go`**

```go
ItemSeekPrefWeight = 20.0  // How many tiles one preference point is worth
ItemSeekDistWeight = 1.0   // Distance penalty per tile
```

Separate from food seeking constants to allow independent tuning — food seeking has urgency tiers (Moderate/Severe/Crisis), item seeking does not.

**3. Output.Kind on construction recipes (DD-59)**

Add Kind to all 6 construction recipe outputs so `NetPreference` can match against them:

| Recipe ID | Current Output | New Output.Kind |
|---|---|---|
| thatch-fence | `{ItemType: "fence"}` | `Kind: "thatch fence"` |
| stick-fence | `{ItemType: "fence"}` | `Kind: "stick fence"` |
| brick-fence | `{ItemType: "fence"}` | `Kind: "brick fence"` |
| thatch-hut | `{ItemType: "hut"}` | `Kind: "thatch hut"` |
| stick-hut | `{ItemType: "hut"}` | `Kind: "stick hut"` |
| brick-hut | `{ItemType: "hut"}` | `Kind: "brick hut"` |

These values match what `PreferenceKind()` returns on the corresponding constructs — so "likes stick fences" formed by looking at a stick fence will match the stick-fence recipe output.

**Pattern:** Follows existing craft recipe Output.Kind ("shell hoe", "hollow gourd"). Follow the Existing Shape; Source of Truth Clarity (recipe carries its own preference identity).

**4. `selectConstructionMaterial` uses `ScoreItemFit`**

Currently (`order_execution.go:1558`): iterates `char.GetKnownRecipesForActivity(activityID)`, finds nearest material of each type via `findNearestItemOrBundle`, picks the one with shortest distance.

Change: for each recipe with available materials, compute:
- `recipePref = char.NetPreference(&Item{ItemType: recipe.Output.ItemType, Kind: recipe.Output.Kind})`
- `nearestItem` from existing `findNearestItemOrBundle` (construction materials don't vary within a type — all sticks are brown — so nearest is still correct per type)
- `componentPref = char.NetPreference(nearestItem)`
- `dist = pos.DistanceTo(nearestItem.Pos())`
- `score = ScoreItemFit(recipePref, componentPref, dist)`

Pick the recipe with the highest score. Return its input material type (same return interface as today).

Note: the existing `countItemsOnMap >= 1` feasibility check stays — infeasible recipes (zero materials) are eliminated before scoring.

**Pattern:** Replaces pure-nearest with scored selection. Same function shape, different selection logic.

### Tests

- **Preference overrides distance:** Character with "likes stick fences" (+1), sticks at distance 25, grass at distance 5. `selectConstructionMaterial` returns "stick" (stick score: 1×20×2 + 0×20 - 25×1 = 15; grass score: 0×20×2 + 0×20 - 5×1 = -5... wait, grass component pref might be negative if character dislikes grass). Verify stick wins.
- **No preferences = nearest wins:** Character with no preferences, sticks at distance 25, grass at distance 5. Returns "grass" (nearest). Existing behavior preserved.
- **Infeasible recipe excluded:** Character prefers sticks but no sticks on map. Returns other available material.
- Existing `TestSelectConstructionMaterial_FenceAndHut` and `TestSelectConstructionMaterial_BelowRecipeCount` must still pass.

### [TEST]

Create test world via `/test-world`:
- Character with know-how for all fence recipes
- Character has "likes stick fences" preference
- Grass bundles (6) near the character, stick bundles (6) far away
- Fence marks placed
- Build Fence order assigned

Observe: character walks past grass to sticks. Action log shows stick material selection.

### [DOCS]
### [RETRO]

---

## Sub-step 11b: Craft Recipe Selection + Item Procurement

**Anchor story:** A character crafting a hoe needs a shell. Two silver shells and a brown shell are on the ground. The silver shell is a few tiles further, but the character prefers silver — they walk to it.

### Changes

**1. `findPreferredItemByType` in `picking.go`**

New helper, sibling of `findNearestItemByType`. Same filtering logic (itemType match, growingOnly, skip full bundles) but scores by preference + distance instead of distance alone:

```go
func findPreferredItemByType(char *entity.Character, cx, cy int, items []*entity.Item, itemType string, growingOnly bool) *entity.Item
```

For each matching item: `score = ScoreItemFit(0, char.NetPreference(item), dist)` (recipe term zero — recipe is already chosen or irrelevant). Returns highest-scoring item. Falls back to nearest behavior when character has no relevant preferences (all scores equal, distance dominates).

Exported test wrapper: `FindPreferredItemByTypeForTest`.

The existing `findNearestItemByType` stays — it's still used for existence checks (`isRecipeFeasible`, `isMultiStepOrderComplete`, etc.) where preference doesn't matter.

**Pattern:** Follows `findNearestItemByType` structure. Follow the Existing Shape. Item Acquisition layer (architecture.md picking.go section).

**2. `findCraftIntent` scores feasible recipes**

Currently (`order_execution.go:438`): iterates recipes, picks first feasible (`break` on first match), calls `EnsureHasRecipeInputs`.

Change: collect all feasible recipes, score each using `ScoreItemFit`:
- `recipePref = char.NetPreference(&Item{ItemType: recipe.Output.ItemType, Kind: recipe.Output.Kind})`
- For each recipe input: find best item via `findPreferredItemByType`. Sum `componentPref` across inputs. Use max distance across inputs (bottleneck).
- `score = ScoreItemFit(recipePref, totalComponentPref, maxDist)`

Pick highest-scoring recipe. Proceed with `EnsureHasRecipeInputs` for that recipe.

Today this is effectively a no-op — one recipe per craft activity. But the infrastructure is ready for future competing recipes (clay pots, Post-Con-Reqs item 2). Existing tests pass because single-recipe scoring produces the same result as first-feasible.

**Pattern:** Extends existing recipe iteration loop. Consider Extensibility.

**3. `EnsureHasRecipeInputs` uses `findPreferredItemByType`**

Currently (`picking.go:257`): calls `findNearestItemByType(char.X, char.Y, items, input.ItemType, false)` for each missing input.

Change: replace with `findPreferredItemByType(char, char.X, char.Y, items, input.ItemType, false)`. Character picks preferred item of the required type (silver shell over brown shell) instead of just nearest.

**4. `EnsureHasItem` uses `findPreferredItemByType`**

Currently (`picking.go:138`): calls `findNearestItemByType(char.X, char.Y, items, itemType, false)`.

Change: replace with `findPreferredItemByType(char, char.X, char.Y, items, itemType, false)`.

### Tests

- **`findPreferredItemByType` picks preferred:** Two shells on map — silver at distance 10, brown at distance 5. Character likes silver. Returns silver shell.
- **`findPreferredItemByType` falls back to nearest:** No preferences. Returns brown (nearer) shell.
- **`EnsureHasItem` picks preferred:** Character needs a hoe, two hoes on map (silver, brown) at different distances. Character prefers silver. Intent targets silver hoe.
- **`EnsureHasRecipeInputs` picks preferred input:** Shell-hoe recipe, two shells of different colors. Character with color preference. Intent targets preferred shell.
- Existing `TestEnsureHasItem_*` and `TestEnsureHasRecipeInputs_*` and `TestFindCraftIntent_*` tests must still pass.

### [TEST]

Create test world via `/test-world`:
- Character with know-how for shell-hoe recipe
- Character has "likes silver" color preference
- Silver shell far away, brown shell nearby, stick nearby
- Craft Hoe order assigned

Observe: character picks up stick (nearest, no color variation), then walks past brown shell to silver shell.

### [DOCS]
### [RETRO]

---

## Sub-step 11c: Vessel Selection + Plantable Selection

**Anchor story:** A character with an Extract order needs a vessel. A brown gourd vessel is nearby; a green gourd vessel is further away. The character prefers green — they walk to it. Another character with a Plant order and no variety lock sees red and blue berries. They prefer red and grab the red berry despite the blue being closer.

### Changes

**1. `FindAvailableVessel` gains `char` parameter and scores by preference**

Currently (`picking.go:743`): iterates all vessels, returns nearest compatible (can accept target item).

Change signature:
```go
func FindAvailableVessel(char *entity.Character, cx, cy int, items []*entity.Item, targetItem *entity.Item, registry *game.VarietyRegistry) *entity.Item
```

For each compatible vessel: `score = ScoreItemFit(0, char.NetPreference(vessel), dist)`. Returns highest-scoring vessel. Vessel attributes (Color, Pattern, Texture, future Kind) are matchable against preferences.

**Caller updates:** `EnsureHasVesselFor` passes `char` (already has it as parameter). All other callers of `FindAvailableVessel` need `char` added — trace and update each. If a caller doesn't have access to a character (unlikely but check), pass `nil` and handle nil char as "use nearest" fallback.

**Anti-pattern:** Do NOT score by vessel contents — this is vessel selection for procurement ("which empty/compatible vessel do I grab"), not plantable selection. The character cares about the vessel's own appearance attributes here.

**Pattern:** Same `ScoreItemFit` with recipe term zero. Item Acquisition layer.

**2. `FindVesselContaining` gains `char` parameter and scores when unlocked**

Currently (`picking.go:524`): iterates vessels with matching plantable contents, returns nearest.

Change signature to add `char *entity.Character`. When called from `EnsureHasPlantable` with `lockedVariety == ""`: score each matching vessel by contents preference: `score = ScoreItemFit(0, char.NetPreferenceForVariety(contentVariety), dist)`. When `lockedVariety != ""`, only one variety matches — scoring is moot, keep nearest.

**Pattern:** Approach B — vessel bucket keeps priority over loose items, scored within.

**3. `findPreferredPlantableOnGround` in picking.go**

New helper, sibling of `findNearestPlantableOnGround`. Used when `lockedVariety == ""`:

```go
func findPreferredPlantableOnGround(char *entity.Character, cx, cy int, items []*entity.Item, targetType string) *entity.Item
```

Scores each matching plantable item by `ScoreItemFit(0, char.NetPreference(item), dist)`. Returns highest-scoring. When `lockedVariety != ""`, caller continues to use `findNearestPlantableOnGround` (only one variety matches).

**4. `EnsureHasPlantable` uses scored helpers when unlocked**

Currently (`picking.go:150`): calls `FindVesselContaining` (vessels first), then `findNearestPlantableOnGround` (loose items fallback).

Change: when `lockedVariety == ""`:
- Vessel bucket: call `FindVesselContaining(char, ...)` — returns preferred-scored vessel
- Loose bucket: call `findPreferredPlantableOnGround(char, ...)` — returns preferred-scored loose item
- Vessel-first priority preserved (approach B): if vessel bucket has a result, use it; otherwise fall back to loose bucket

When `lockedVariety != ""`: no change — existing nearest behavior, single variety match.

**Pattern:** Approach B — buckets preserved, scored within. Follow the Existing Shape (vessel-first priority is established behavior).

### Tests

- **`FindAvailableVessel` picks preferred:** Two empty vessels on map — green gourd at distance 10, brown gourd at distance 5. Character likes green. Returns green vessel.
- **`FindAvailableVessel` falls back to nearest:** No preferences. Returns brown (nearer).
- **`FindVesselContaining` picks preferred contents (unlocked):** Two vessels with different berry varieties. Character prefers red berries. Returns vessel with red berries.
- **`findPreferredPlantableOnGround` picks preferred:** Red and blue loose berries. Character prefers red. Returns red.
- **`EnsureHasPlantable` locked = unchanged:** Locked to red variety, blue berry closer. Still picks red (only match). Verify existing behavior.
- **`EnsureHasPlantable` unlocked = preferred:** No lock, red and blue berries, character prefers red. Picks red.
- Existing `TestEnsureHasPlantable_*` and `TestEnsureHasVesselFor_*` tests must still pass.

### [TEST]

Create test world via `/test-world`:
- Character with know-how for planting
- Character has color preference (e.g., "likes red")
- Red and blue berries at different distances, both plantable
- Plant order with no variety lock
- Tilled soil available

Observe: character walks to preferred berry variety despite another being closer.

Second scenario (vessel):
- Character with Extract order
- Two empty vessels of different colors at different distances
- Character has color preference

Observe: character walks to preferred vessel.

### [DOCS]
### [RETRO]

---

## Post-completion

- Resolve architecture.md "Future: Unified Item-Seeking" note — the gap is closed. Update to describe the `ScoreItemFit` pattern and which call sites use it.
- Resolve `triggered-enhancements.md` entry referencing unified item seeking.
