# Step Spec: Step 4 — Craft Bricks

Design doc: [construction-design.md](construction-design.md)

## Sub-step 4a: Brick Item Type, Recipe, Activity, Discovery

**Anchor story:** A character wanders near the clay deposits, picks up a lump of clay, and has a flash of insight — they could shape this into bricks! The player opens the orders panel and sees "Craft > Brick" as a new option. The order shows as unfulfillable if no loose clay exists in the world.

### Tests

**Infrastructure — VesselExcludedTypes split:**
- `TestVesselExcluded_Clay` — `AddToVessel` returns false for clay items
- `TestVesselExcluded_Brick` — `AddToVessel` returns false for brick items
- `TestCanVesselAccept_Clay` — `CanVesselAccept` returns false for clay items
- `TestCanVesselAccept_Brick` — `CanVesselAccept` returns false for brick items
- Existing vessel exclusion tests for sticks/grass must still pass (they now use VesselExcludedTypes instead of MaxBundleSize)

**NewBrick constructor:**
- `TestNewBrick_Properties` — ItemType "brick", symbol `▬` (config.CharBrick), Color ColorTerracotta, no Kind, no Plant, no Container, BundleCount 0

**Save/load:**
- `TestBrickSymbol_RoundTrip` — brick item survives save/load with correct symbol restored (follows existing pattern in itemFromSave switch)

**Activity registry (checkpoint 2):**
- `TestCraftBrickActivity_InRegistry` — "craftBrick" activity exists with Category "craft", IntentOrderable, AvailabilityKnowHow

**Recipe:**
- `TestCraftBrickRecipe_InRegistry` — "clay-brick" recipe exists with Input clay (count 1), Output ItemType "brick", Repeatable true

**Discovery (checkpoint 1):**
- `TestCraftBrickDiscovery_LookAtClay` — ActionLook on a clay item triggers "craftBrick" know-how
- `TestCraftBrickDiscovery_PickupClay` — ActionPickup on a clay item triggers "craftBrick" know-how
- `TestCraftBrickDiscovery_DigClay` — ActionDig on a clay item triggers "craftBrick" know-how

**Feasibility (checkpoint 3):**
- `TestIsOrderFeasible_CraftBrick_ClayExists` — returns (true, false) when clay items exist on the ground
- `TestIsOrderFeasible_CraftBrick_NoClay` — returns (false, false) when no clay items exist
- `TestIsOrderFeasible_CraftBrick_NoKnowHow` — returns (false, true) when clay exists but no character knows craftBrick

### Implementation

**VesselExcludedTypes split (config/config.go):**
- Add `VesselExcludedTypes = map[string]bool{"stick": true, "grass": true, "clay": true, "brick": true}`
- `MaxBundleSize` stays as-is (stick: 6, grass: 6) — used only for bundle logic
- Update comment on `MaxBundleSize` to remove the "also vessel-excluded" note
- Pattern: resolves triggered enhancement (DD-20)

**Update vessel exclusion check sites:**
- `picking.go` — `AddToVessel` (~line 582): change `config.MaxBundleSize[item.ItemType] > 0` → `config.VesselExcludedTypes[item.ItemType]`
- `picking.go` — `CanVesselAccept` (~line 690): same change
- `picking.go` — `Pickup` vessel path skip (~line 1082): change `config.MaxBundleSize[item.ItemType] == 0` → `!config.VesselExcludedTypes[item.ItemType]`
- `order_execution.go` — `findHarvestIntent` (~line 166): change `config.MaxBundleSize[target.ItemType] == 0` → `!config.VesselExcludedTypes[target.ItemType]`
- `order_execution.go` — `findGatherIntent` (~line 923): change `config.MaxBundleSize[target.ItemType] > 0` → `config.VesselExcludedTypes[target.ItemType]`
- **Anti-pattern:** Do NOT change bundle-logic sites (hasFullBundle, DropCompletedBundle, CanMergeIntoBundle, canGatherMore, FindNextHarvestTarget, lifecycle.go sprout maturation, serialize.go backward compat). Those correctly use MaxBundleSize for bundle behavior.

**New color (types/types.go, ui/styles.go):**
- Add `ColorTerracotta Color = "terracotta"` to types.go constants
- Add `ColorTerracotta` to `AllColors` slice
- Add `terracottaStyle` to styles.go — warm reddish-brown ANSI color (suggest 166, verify visually)
- Add `ColorTerracotta` case to the color-to-style switch in view.go's `getItemStyle` (or equivalent rendering function)

**Brick symbol (config/config.go):**
- Add `CharBrick = '▬'`

**NewBrick constructor (entity/item.go):**
- `NewBrick(x, y int) *Item` — ItemType "brick", Sym config.CharBrick, Color types.ColorTerracotta, Name "brick", no Kind, no Plant, no Container, no Edible, BundleCount 0
- Pattern: follows NewClay (uniform item, no variety, DD-21)

**Repeatable field (entity/recipe.go):**
- Add `Repeatable bool` to Recipe struct — when true, craft order loops until world-state completion condition is met rather than completing after one craft (DD-19)

**Recipe (entity/recipe.go):**
- Add "clay-brick" to RecipeRegistry:
  - ID: "clay-brick", ActivityID: "craftBrick", Name: "Clay Brick"
  - Inputs: `[]RecipeInput{{ItemType: "clay", Count: 1}}`
  - Output: `RecipeOutput{ItemType: "brick"}`
  - Duration: `config.ActionDurationLong`
  - Repeatable: true
  - DiscoveryTriggers: ActionLook on clay, ActionPickup on clay, ActionDig on clay

**Activity (entity/activity.go):**
- Add "craftBrick" to ActivityRegistry:
  - ID: "craftBrick", Name: "Brick", Category: "craft"
  - IntentFormation: IntentOrderable, Availability: AvailabilityKnowHow
  - No DiscoveryTriggers on the activity (discovery is via recipe triggers)

**Save/load (ui/serialize.go):**
- Add `case "brick": item.Sym = config.CharBrick` to the symbol restoration switch in `itemFromSave()`

**Color rendering (ui/view.go or styles.go):**
- Add `case types.ColorTerracotta: return terracottaStyle` to the item color rendering switch (trace the actual function name — likely `getItemStyle` or a switch in `renderCell`)

**Architecture pattern:** Follows **Adding a New Recipe** checklist (architecture.md). VesselExcludedTypes split follows **Source of Truth Clarity** (DD-20). Brick item follows **Start With the Simpler Rule** (DD-21). Recipe discovery follows **Follow the Existing Shape** — recipe triggers are the established pattern for craft activities.

**Values:** **Isomorphism** — bricks are a distinct material from clay, visually and conceptually. **Source of Truth Clarity** — VesselExcludedTypes is now its own config set, not piggybacked on MaxBundleSize.

**Reqs:** Construction-Reqs lines 54-57: "Craft > Bricks, orderable, discoverable chance by picking up or looking at clay." Discovery also triggers on digging clay (ActionDig added in Step 3 wasn't anticipated by original reqs). "User selects number of bricks to craft" deferred to triggered enhancements.

### [TEST]

Build and run the game. Verify:
- A character looks at or picks up loose clay and discovers brick crafting (check knowledge panel)
- "Craft > Brick" appears in the orders panel under the Craft category after discovery
- If no clay items exist on the ground, the order shows as unfulfillable
- Creating a Craft > Brick order succeeds (order appears in list) — **crafting won't execute yet, that's 4b**
- Save and reload preserves brick items if manually spawned via test world
- Clay items cannot be placed in vessels (verify via character behavior — they skip vessel path for clay)

Use `/test-world` with a character near clay tiles and pre-discovered craftBrick know-how for faster testing.

### [DOCS]
### [RETRO]

---

## Sub-step 4b: Craft Execution and Multi-Step Completion

**Anchor story:** The player has created a Craft > Brick order. A character picks up a lump of clay, pauses to shape it, and sets down a terracotta brick. Without hesitation they head for the next lump of clay. Another character joins the effort. They work through all the loose clay, and when the last lump is shaped, the order completes.

### Tests

**CreateBrick function:**
- `TestCreateBrick_Properties` — output has ItemType "brick", Color ColorTerracotta, Sym CharBrick, position matches input

**Execution start (checkpoint 4):**
- `TestFindCraftIntent_CraftBrick_FindsClayOnGround` — returns pickup intent targeting nearest clay when character has no clay
- `TestFindCraftIntent_CraftBrick_ReturnsActionCraft` — returns ActionCraft with RecipeID "clay-brick" when character has clay in inventory

**Produces brick (checkpoint 5):**
- `TestApplyCraft_CraftBrick_CreatesBrick` — after ActionDurationLong, clay consumed from inventory, brick item placed on map at character position

**Repeatable behavior (checkpoint 6):**
- `TestApplyCraft_CraftBrick_DoesNotCompleteOrder` — after crafting a brick, order is NOT marked OrderCompleted (recipe.Repeatable skips inline completion)
- `TestApplyCraft_NonRepeatable_CompletesOrder` — after crafting a vessel, order IS marked OrderCompleted (existing behavior unchanged)
- `TestIsMultiStepOrderComplete_CraftBrick_ClayExists` — returns false when clay items exist on the ground
- `TestIsMultiStepOrderComplete_CraftBrick_NoClay` — returns true when no clay items exist on the ground

**End-to-end (checkpoints 4-6 chained):**
- `TestCraftBrickOrder_EndToEnd` — full flow: character assigned to craftBrick order → picks up clay → crafts brick (brick appears on map) → picks up next clay → crafts another brick → no more clay → isMultiStepOrderComplete true → CompleteOrder. Follows `TestGatherOrder_InventoryPath_FullBundle_EndToEnd` pattern.

### Implementation

**CreateBrick function (system/crafting.go):**
- `CreateBrick(clay *entity.Item, recipe *entity.Recipe) *entity.Item` — returns `NewBrick(clay.X, clay.Y)`
- Pattern: follows CreateVessel/CreateHoe — receives consumed input, creates output item

**applyCraft dispatch (ui/apply_actions.go):**
- Add case to the recipe dispatch switch in applyCraft:
  ```
  case "clay-brick":
      crafted = system.CreateBrick(consumed["clay"], recipe)
  ```

**Repeatable skip in applyCraft (ui/apply_actions.go):**
- Wrap the existing inline CompleteOrder block (~lines 523-528) with a Repeatable check:
  ```
  if char.AssignedOrderID != 0 && !recipe.Repeatable {
      // existing CompleteOrder logic
  }
  ```
- When Repeatable is true, intent clears (existing line 531) but order stays active. Next tick, selectOrderActivity re-evaluates via findCraftIntent.
- **Anti-pattern:** Do NOT remove the existing inline completion for vessel/hoe. Only skip it when Repeatable is true.

**isMultiStepOrderComplete (system/order_execution.go):**
- Add case to the switch:
  ```
  case "craftBrick":
      return !groundItemOfTypeExists(gameMap.Items(), "clay")
  ```
- This is reached when findCraftIntent returns nil (no clay accessible). If no clay on the ground either → order complete. If clay is in another character's inventory → this character's work is done; the other will finish their craft independently.

**No changes needed to:**
- `findCraftIntent` — already generic for all recipe-based activities
- `findOrderIntent` — already routes recipe-based activities to findCraftIntent
- `IsOrderFeasible` — already uses `isAnyRecipeWorldFeasible` for recipe-based activities
- `continueIntent` — ActionCraft has no early-return block (single-phase, no vessel procurement)
- Order UI — craft sub-category already groups by Category "craft"; no sub-menu or target type selection needed

**Behavioral details (architecture.md checklist):**
- **Targeting:** same-tile (character crafts at their current position)
- **Duration:** `ActionDurationLong` (~2 world hours)
- **Completion criteria:** no clay items exist on the ground (multi-step via Repeatable + isMultiStepOrderComplete)
- **Feasibility criteria:** any recipe input (clay) exists in the world (generic recipe feasibility)
- **Variety lock:** none — clay is uniform (DD-14)
- **DisplayName:** "Craft brick" (via Category "craft" → "Craft " + lowercase activity.Name)
- **Sub-menu:** none — craftBrick is a leaf activity under the Craft category, no target type selection

**Architecture pattern:** Follows **Adding a New Recipe** checklist. Multi-step completion follows **Follow the Existing Shape** — gather/dig use the same selectOrderActivity → isMultiStepOrderComplete loop (DD-19). Repeatable field on Recipe follows **Consider Extensibility**.

**Values:** **Follow the Existing Shape** — applyCraft handler, recipe dispatch, and multi-step loop are all existing patterns. **Anchor to Intent, Not Structure** — end-to-end test validates "clay becomes bricks until clay is gone", not individual function returns.

**Reqs:** Construction-Reqs lines 54-57. "Craft > Bricks, orderable" — implemented via recipe-based craft order. "User selects number" — deferred; order instead processes all available clay.

### [TEST]

Build and run the game. Verify:
- Creating a Craft > Brick order: character drops non-clay items, walks to clay, picks it up, shapes it (visible pause), brick appears on the ground
- Character immediately seeks the next lump of clay and repeats
- Assign a second character to the same order — both work in parallel
- When the last clay is consumed, the order completes
- Bricks are terracotta-colored `▬` symbols, visually distinct from clay's earthy `☗`
- Character pauses to eat/drink if needs become pressing, then resumes crafting
- Save and reload mid-craft preserves order state and brick items

Use `/test-world` with a character near clay tiles, pre-discovered craftBrick know-how, and several loose clay items for faster testing.

### [DOCS]
### [RETRO]
