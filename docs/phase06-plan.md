# Phase 6: Containers and Storage - Implementation Plan

**Status:** Feature 4 Complete - Ready for Feature 5
**Requirements:** [docs/phase06reqs.txt](phase06reqs.txt)

---

## Overview

Characters can craft containers and store things in them. Items can be placed on the ground and picked up. This phase introduces crafting, the vessel item type, and item stacking.

---

## Phase 6 Prep

**Trigger:** Adding non-plant categories (tools, crafted items) - from futureEnhancements.md
**Addresses:** OQ-A (growing vs picked plants), OQ-C (item categories)

### Prep Stage 1: PlantProperties Refactor ✓

- [x] Create `PlantProperties` struct
  - `IsGrowing bool`: can reproduce at location (gates spawning)
  - `SpawnTimer float64`: countdown until next spawn
- [x] Add `Plant *PlantProperties` to Item struct
- [x] Update item creation functions (`NewBerry`, `NewMushroom`, `NewFlower`, `NewGourd`)
- [x] Update spawning logic: `Plant != nil && Plant.IsGrowing`
- [x] Remove `Category string` from Item
- [x] Update serialization for PlantProperties
- [x] Update tests

**Test checkpoint:** Game works, plants spawn normally, save/load works

### Prep Stage 2: Container Structs ✓

- [x] Create `Stack` struct
  - `Variety *ItemVariety`: what variety this stack holds
  - `Count int`: how many in the stack
- [x] Create `ContainerData` struct
  - `Capacity int`: how many stacks it can hold
  - `Contents []Stack`: slice of stacks
- [x] Add `Container *ContainerData` to Item (nil for non-containers)
- [x] Update serialization (containers nil for existing items)

**Test checkpoint:** Save/load still works ✓

### Design Decisions (Prep)

**Fields staying on Item** (broader applicability):

- `Edible`, `Poisonous`, `Healing` - future crafted consumables
- `DeathTimer` - future spoiling/degrading items

**Deferred fields** (not needed for Phase 6):

- `IsPortable` on ContainerData - all items currently carriable
- `IsWatertight` on ContainerData - all vessels are watertight
- `LockedVariety` on ContainerData - future vessel restriction

---

## Feature 1: Item Placement System ✓

**Addresses:** OQ-A (picked plants don't respawn)

### Already Complete (from Prep Stage 1)

- `IsGrowing = false` set on pickup (foraging.go)
- Spawning correctly skips `IsGrowing = false` items (lifecycle.go)
- Existing pickup logic works for any item on map

### Scope Decision

Drop action deferred until Feature 2/3 when crafting logic needs it. No player-initiated drops - characters decide when to drop based on simulation needs.

### Tasks

- [x] Show "Growing" status in item details panel for growing plants

### Deferred to Feature 2/3

- Implement `Drop` function:
  - Remove item from `char.Carrying`
  - Place item on map at character's position
  - `IsGrowing` already false (set on pickup), stays false
  - Dropped item can be picked up again via existing pickup logic
- AI logic for dropping (when inventory full and need to pick up something else)

**Test checkpoint:** Select growing plant on map, verify "Growing" shown in details panel ✓

---

## Feature 2: Crafting Foundation ✓

**Addresses:** Req A.1 (discovery), A.2 (orderable), OQ-B (crafting know-how structure)

### Requirements

- Req A.1.i: Discovery when engaging in gourd interaction (looking, picking up, consuming)
- Req A.1.ii: Discovery when engaging in spring interaction (drinking)
- Req A.2: Orderable via Craft option in task menu
- OQ-B: General "crafting" know-how + specific recipe know-how

### Design Decisions

**Know-how structure:**

- Activity-level: `craftVessel` in `KnownActivities` (like `harvest`)
- Recipe-level: `KnownRecipes []string` on Character (e.g., `["hollow-gourd"]`)
- UI groups all `craft*` activities under a "Craft" menu header (presentation convention)

**Recipe struct** (`entity/recipe.go`):

```go
type Recipe struct {
    ID                string             // "hollow-gourd"
    ActivityID        string             // "craftVessel" - links recipe to activity
    Name              string             // "Hollow Gourd"
    Inputs            []RecipeInput      // [{ItemType: "gourd", Count: 1}]
    Output            RecipeOutput       // ItemType: "vessel", creates ContainerData
    Duration          float64            // Game time in seconds
    DiscoveryTriggers []DiscoveryTrigger // Triggers for discovering this recipe
}
```

**Discovery triggers - two patterns:**

- **Direct activity discovery** (e.g., harvest): Triggers on Activity, grants activity only
- **Recipe-based discovery** (e.g., craftVessel): Triggers on Recipe, grants activity + recipe

For craftVessel:

- Activity has no DiscoveryTriggers (discovered via recipes)
- hollow-gourd recipe has triggers: gourd interaction (look, pickup, consume), drinking (ActionDrink)
- Discovery grants: `craftVessel` activity + `hollow-gourd` recipe together

**Orders:**

- Order: `{ActivityID: "craftVessel", TargetType: ""}` (no target - recipe determines output)
- Character selects recipe based on known recipes for that activity + available materials

**Knowledge transfer (talking):**

- Can only receive a recipe if you already know the corresponding activity
- e.g., must know `craftVessel` to receive a vessel recipe

### Tasks

- [x] Create `entity/recipe.go` with Recipe struct (including DiscoveryTriggers) and RecipeRegistry
- [x] Add `KnownRecipes []string` to Character
- [x] Add `craftVessel` activity to ActivityRegistry (no discovery triggers - discovered via recipes)
- [x] Update discovery system to also check recipe triggers
- [x] Update discovery system to handle ActionDrink (item can be nil)
- [x] Recipe discovery grants activity + recipe together
- [x] UI: Two-step order selection (Craft → Vessel)
- [x] UI: Knowledge panel shows "Craft: Vessel" and Recipes section
- [x] UI: Order list shows "Craft vessel [status]"
- [x] Update serialization for KnownRecipes
- [x] Tests for recipe system, discovery

**Test checkpoint:** Character discovers craftVessel via gourd/spring interaction, Craft > Vessel appears in order menu ✓

---

## Feature 3: Hollow Gourd Vessel ✓

**Addresses:** Req A.3 (recipe), B.1 (post-craft inventory)

### Requirements

- Req A.3: Recipe is 1 gourd for 2 minutes game time
- Req A.3.i: If gourd in inventory, begin crafting
- Req A.3.ii: If no gourd, target gourd to pick up; drop current item if inventory full
- Req B.1: Crafted item goes to inventory if room, else dropped

### Design Decisions

**Recipe definition:** (already in RecipeRegistry from Feature 2)

- ID: `"hollow-gourd"`
- ActivityID: `"craftVessel"`
- Input: 1 gourd (any variety)
- Output: vessel item with `Container = &ContainerData{Capacity: 1}`
- Duration: 10 seconds (temporary - see Recipe Timing Note)

**Crafting state:**

- Use existing `ActionProgress` pattern (same as eat, drink, look, pickup)
- Add `ActionCraft` to entity actions
- When `ActionProgress >= Recipe.Duration`, craft completes
- Pause/resume uses existing `OrderPaused` mechanism

**Vessel appearance:**

- Item already has `Color`, `Pattern`, `Texture` fields
- Vessel inherits appearance from input gourd
- Crafted items use Item.Name for display (supports future painting feature)
- Display name: "Hollow Gourd" (from recipe.Name via Item.Name field)

**Naming:**

- Recipe: "Hollow Gourd"
- Created item: ItemType "vessel", Name "Hollow Gourd", display "Hollow Gourd"

### Tasks

- [x] Add `Name` field to Item for crafted item display names
- [x] Update `Description()` to return Name if set
- [x] Update Item serialization for Name field
- [x] Add ActionCraft to entity actions
- [x] Implement vessel item creation (`CreateVessel` in system/crafting.go)
- [x] Implement Drop function (system/foraging.go)
- [x] Craft intent/execution: acquire gourd, drop if blocked, craft duration
- [x] Post-craft: vessel goes to inventory
- [x] Tests for craft flow, vessel creation, vessel not edible
- [x] Bug fix: ActionMove was eating non-edible items (look intents triggered eating)

**Test checkpoint:** Order craft vessel, watch character acquire gourd, craft, and hold result ✓

---

## Feature 3b: IsGrowing Filter for Foraging/Harvesting ✓

**Addresses:** Bug - dropped items shouldn't be targeted by foraging/harvesting

Dropped (non-growing) items should only be eligible for looking and eating, not foraging and harvesting. Foraging and harvesting should only target growing plants.

### Tasks

- [x] `findForageTarget`: filter for `Plant != nil && Plant.IsGrowing`
- [x] `findNearestItemByType` (harvest): filter for `Plant != nil && Plant.IsGrowing`
- [x] Tests for filter behavior
- [x] Item details panel: show Name and Pattern/Texture for all items (including crafted)

**Test checkpoint:** Drop an item, verify characters don't forage/harvest it but can still eat/look at it ✓

---

## Feature 4: Vessel Contents ✓

**Addresses:** Req B.2 (filling vessel), B.3 (look for container), B.4 (drop when blocked)

### Requirements

- Req B.2.i: Stack sizes - Mushrooms: 10, Berries: 20, Gourds: 1, Flowers: 10
- Req B.2.ii: Once variety added to a Stack, only that variety can be added to that Stack
- Req B.2.iii: Foraging fills vessel before stopping
- Req B.2.iv: Harvesting complete when vessel full OR no matching varieties
- Req B.3: If not carrying container, look for empty/matching vessel first
- Req B.4: Drop container when it blocks action

### Design Decision: Stack Slice

Contents tracked via `Contents []Stack` where each Stack has `Variety` + `Count`:

- Adding item: if empty, append new Stack; if same variety, increment count
- Eating: decrement count; remove Stack when count hits 0
- Vessel variety-lock: enforced because `Capacity == 1` means only one Stack allowed
- Future containers with larger capacity could hold multiple Stacks of different varieties

### Clarifications

- **Variety lock is on the Stack**, not the vessel. Any item can go in an empty vessel. Subsequent items must match the first item's variety.
- **Completion condition** for foraging/harvesting: vessel full, OR inventory full (no vessel case), OR no eligible items remaining.
- **Both foraging and harvesting** continue until one of the completion conditions is met.

### Stage 4a: StackSize and Add-to-Vessel ✓

- [x] Add `StackSize` map to config (berry: 20, mushroom: 10, gourd: 1, flower: 10)
- [x] Add `GetStackSize(itemType string) int` helper
- [x] Add `GetByAttributes` helper to VarietyRegistry for looking up variety by item attributes
- [x] Implement `AddToVessel(vessel, item *Item, registry) bool` function
- [x] Implement `IsVesselFull(vessel, registry) bool` helper
- [x] Tests for add-to-vessel logic, variety lock, stack size limits (12 tests)

**Test checkpoint 4a:** Unit tests pass ✓

### Code Organization (Option A) ✓

Pickup logic is split across multiple files (see `architecture.md`). For Phase 6, relocated `findForageIntent`/`findForageTarget` from `movement.go` to `foraging.go` to consolidate foraging logic. Deferring full unification to `picking.go` until after Feature 4 testing (see `futureEnhancements.md` for triggers).

- [x] Move `findForageIntent` and `findForageTarget` from `movement.go` to `foraging.go`

### Stage 4b: Foraging with Vessel ✓

- [x] Modify `Pickup()` to detect when carrying a vessel
  - If carrying vessel with space: call `AddToVessel`, returns `PickupToVessel`
  - Intent NOT cleared when adding to vessel (caller handles continuation)
- [x] Add `FindNextVesselTarget()` for vessel filling continuation
  - Finds next growing item matching vessel's variety
- [x] Update `applyIntent` in update.go to continue foraging after `PickupToVessel`
  - Calls `FindNextVesselTarget`, sets new intent if found
  - Stops when: vessel full, or no eligible items remaining
- [x] If not carrying vessel, current behavior applies (pick up one item)
- [x] Tests for foraging with vessel (16 tests for vessel logic)
- [x] Fix: Add `CanPickUpMore()` helper - foraging was blocked when carrying vessel because `IsInventoryFull()` returned true
- [x] Update `selectIdleActivity` to use `CanPickUpMore` instead of `!IsInventoryFull()`

**Test checkpoint 4b:** Manually test foraging with vessel - verify items stack, variety locks, foraging continues until full ✓

### Stage 4c: Harvesting with Vessel ✓

- [x] Update harvesting order completion logic
  - `PickupToVessel`: order completes when `FindNextVesselTarget` returns nil (vessel full or no matching items)
  - `PickupToInventory`: order completes immediately (inventory full)
- [x] Fix: Don't drop vessel when on order if vessel has space (was dropping before adding)
- [x] Tests for harvesting with vessel (3 tests)

**Test checkpoint 4c:** Manually test harvesting with vessel - verify order continues until vessel full or map empty ✓

### Stage 4d: Look-for-Container (Req B.3) ✓

**Design:**

- When starting forage/harvest, check if current inventory can accept target item
- If not carrying anything OR carrying incompatible item, look for available vessel first
- Available vessel = empty OR partially-filled with matching variety and space
- Pick up vessel first, then continue to forage/harvest target

**Tasks:**

- [x] Add `PickupFailed` result to handle variety mismatch bug
- [x] Fix `Pickup()` to return `PickupFailed` instead of overwriting vessel
- [x] Add `CanVesselAccept(vessel, item, registry)` helper
- [x] Add `FindAvailableVessel(cx, cy, items, targetItem, registry)` function
- [x] Update `findForageIntent` - look for vessel first if not carrying one, filter targets by vessel variety
- [x] Update `findHarvestIntent` - look for vessel first if not carrying one
- [x] Tests for look-for-container behavior (8 new tests)

**Test checkpoint 4d:** Manually test - character without vessel finds and picks up vessel before foraging/harvesting ✓

### Stage 4e: Drop-when-Blocked (Req B.4) ✓

**Design:**

- If character's vessel has variety mismatch with target item:
  - For orders (harvesting): drop vessel and pick up item directly
  - For idle foraging: skip target (don't lose vessel contents for casual pickup)
- Uses existing `Drop()` function

**Tasks:**

- [x] Update `findForageIntent` - skip target if carrying incompatible vessel (filters via `findForageTarget`)
- [x] Update `findHarvestIntent` - drop vessel if it blocks harvest (order takes priority)
- [x] Update callers to handle `PickupFailed` result (update.go)
- [x] Tests for drop-when-blocked scenarios (3 harvest tests)

**Test checkpoint 4e:** Manually test - character drops vessel when it blocks their harvest order ✓

---

## Feature 5: Eating from Vessels ✓

**Addresses:** Req B.5 (eating from vessel)

### Requirements

- Req B.5: Vessel contents count as carried item for eating when hungry
- Eating decrements Stack count by 1
- When Stack empty, vessel accepts any variety again
- Characters can also eat from dropped vessels on the ground

### Design Decisions

**Unified Food Selection (Option D):** All food sources use the same proximity-based scoring system. This also addresses the deferred enhancement of applying preference/poison logic to carried items.

**Food sources with distances:**
- Carried loose item: distance = 0
- Carried vessel contents: distance = 0
- Dropped vessel (same tile): distance = 0
- Dropped vessel (nearby): distance = Manhattan distance
- Map item: distance = Manhattan distance

**Scoring formula (unchanged from current map food):**
```
Score = (NetPreference × PrefWeight) - (Distance × DistWeight) + HealingBonus
```

**Hunger tier behavior (unchanged):**
- Moderate (50-74): Filter disliked (NetPreference < 0), high preference weight
- Severe (75-89): Allow all, medium preference weight
- Crisis (90+): Allow all, distance only (no preference weight)

**Key behavior changes:**
- Carried disliked item filtered at Moderate hunger → character seeks better food
- At Severe+, carried food still has distance advantage but doesn't auto-win
- At Crisis, carried food wins on distance (as before, but now explicit)
- Poison knowledge creates dislike preference → filtered at Moderate

**Dropped vessel interaction:** Same tile required (consistent with eating map items).

**Intent handling:** Infer food source from context in applyIntent:
- ActionConsume: if Carrying is vessel with contents → eat from contents
- ActionMove arriving at target: if TargetItem is vessel with contents → eat from contents

### Stage 5a: Variety-based Methods ✓

Add methods to check preferences/knowledge against ItemVariety (for vessel Stack contents):

- [x] `Preference.MatchesVariety(v *ItemVariety) bool`
- [x] `Preference.MatchScoreVariety(v *ItemVariety) int`
- [x] `Knowledge.MatchesVariety(v *ItemVariety) bool`
- [x] `Character.NetPreferenceForVariety(v *ItemVariety) int`
- [x] `Character.KnowsVarietyIsHealing(v *ItemVariety) bool`
- [x] `NewKnowledgeFromVariety(v *ItemVariety, category) Knowledge`
- [x] Tests for variety-based methods (19 tests)

**Test checkpoint 5a:** Unit tests pass for variety methods ✓

### Stage 5b: Unified Food Selection ✓

Refactor findFoodIntent/findFoodTarget to use unified candidate scoring:

- [x] Modify `findFoodTarget` to include carried items with distance=0
- [x] Modify `findFoodTarget` to score carried vessel contents by variety
- [x] Update `findFoodIntent` to check if result is carried item → ActionConsume
- [x] Tests for unified food selection (5 new tests)

**Note:** Simplified implementation - no FoodCandidate struct needed. Context inference (Option C) allows existing return type to work.

**Test checkpoint 5b:** Unit tests verify carried disliked item filtered at Moderate hunger ✓

### Stage 5c: Eating from Carried Vessel ✓

- [x] Update ActionConsume handling: detect vessel with contents, eat from contents
- [x] Implement `ConsumeFromVessel()`: decrement stack, remove when empty, apply effects
- [x] Effects come from Variety (poisonous, healing attributes)
- [x] Knowledge formation from vessel contents via `NewKnowledgeFromVariety`
- [x] Mood adjustment via `NetPreferenceForVariety`
- [x] Tests for eating from carried vessel (7 tests)

**Human test checkpoint 5c:** Character with carried vessel eats from contents when hungry ✓

### Stage 5d: Eating from Dropped Vessel ✓

- [x] Include dropped vessels with edible contents in `findFoodTarget` scoring
- [x] Update arrival handling in update.go: if TargetItem is vessel with contents → eat from contents
- [x] Same `ConsumeFromVessel()` logic as carried vessel
- [x] Tests for eating from dropped vessel (4 tests)

**Human test checkpoint 5d:** Character finds and eats from dropped vessel on ground ✓

### Bug Fix: VarietySave Missing Edible Field

During human testing, discovered that `VarietySave` was missing the `Edible` field. When loading saved games, variety `Edible` defaulted to `false`, so dropped vessel contents weren't recognized as food.

- [x] Add `Edible` field to `VarietySave` in `internal/save/state.go`
- [x] Update `varietiesToSave()` to include Edible
- [x] Update `varietiesFromSave()` to restore Edible

### Stage 5e: Documentation ✓

- [x] Update README, Game Mechanics, claude.md, and architecture docs
- [x] Document test world creation process in feature-dev-process.md

---

## Feature 6: UI Updates

**Prioritization:** Quick wins first, then items needing more design discussion.

### Stage 6a: Action Log Colors ✓

- [x] Order-related events: dusty rose (174), keyword `"order:"`
- [x] Recovery events: changed from light blue (117) to cyan (45)
- [x] Sleep events: fixed keyword from "Falling asleep" to `"asleep"`
- [x] Highlight style: changed from cyan text to dark cyan background (23) with white text

### Stage 6b: Quick Wins

- [x] Keypress 'b' for back when cycling through selected characters

---

## Phase Clean-up and Close-out

After each of the below items, Update README, Game Mechanics, claude.md, and architecture docs as applicable, for changes since last doc updates.

### Clean-up 1: Verification Tasks ✓

- [x] **Investigate mood from looking at hollow gourds** - Verified: preferences CAN form for vessels (itemType + color). Pattern/texture not available (mushroom-only). Limitation: preferences are for "vessel" category, not specific recipe like "hollow-gourd". Kind field enhancement deferred (see triggered-enhancements.md).
- [x] **Verify nested item serialization** - Verified with tests: `TestFromSaveState_RestoresVesselWithContents` and `TestFromSaveState_RestoresCarriedVesselWithContents` both pass.

### Clean-up 2: Structure Refactor ✓

- [x] **Revisit container and plant structures** - Created `EdibleProperties` struct with `Poisonous` and `Healing` fields. `Item` and `ItemVariety` now use `*EdibleProperties` (nil for non-edible items). Added `IsEdible()`, `IsPoisonous()`, `IsHealing()` helper methods. UI only shows Poisonous/Healing for edible items. Also updated `NewGourd` to accept poisonous/healing parameters for future flexibility.

### Clean-up 3: Code Unification ✓

- [x] **Evaluate pickup code unification** - Evaluated: duplication exists (~30 lines vessel-seeking pattern) plus mixed concerns (foraging.go has generic pickup functions, update.go has order-specific completion logic). Refactor deferred until adding another pickup-based order/activity. Detailed guidance added to triggered-enhancements.md.

### Clean-up 4: Balance Tuning

- [ ] **Time config reset:** Adjust so "world day" = 2 game minutes (0.5 game seconds = ~6 "world minutes")