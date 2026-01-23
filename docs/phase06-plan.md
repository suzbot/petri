# Phase 6: Containers and Storage - Implementation Plan

**Status:** Feature 2 Complete - Ready for Feature 3
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
- [ ] Create `entity/recipe.go` with Recipe struct (including DiscoveryTriggers) and RecipeRegistry
- [ ] Add `KnownRecipes []string` to Character
- [ ] Add `craftVessel` activity to ActivityRegistry (no discovery triggers - discovered via recipes)
- [ ] Update discovery system to also check recipe triggers
- [ ] Update discovery system to handle ActionDrink (item can be nil)
- [ ] Recipe discovery grants activity + recipe together
- [ ] UI: Group craft* activities under "Craft" menu header
- [ ] UI: Craft activities skip target selection step
- [ ] Update serialization for KnownRecipes
- [ ] Tests for recipe system, discovery

**Test checkpoint:** Character discovers craftVessel via gourd/spring interaction, Craft > Vessel appears in order menu

---

## Feature 3: Hollow Gourd Vessel

**Addresses:** Req A.3 (recipe), B.1 (post-craft inventory)

### Requirements
- Req A.3: Recipe is 1 gourd for 2 minutes game time
- Req A.3.i: If gourd in inventory, begin crafting
- Req A.3.ii: If no gourd, target gourd to pick up; drop current item if inventory full
- Req B.1: Crafted item goes to inventory if room, else dropped

### Design Decisions

**Recipe definition:**
- ID: `"hollow-gourd"`
- ActivityID: `"craftVessel"`
- Input: 1 gourd (any variety)
- Output: vessel item with `Container = &ContainerData{Capacity: 1}`
- Duration: 2 minutes game time
- Vessel variety inherits color/pattern/texture from input gourd

**Naming:**
- Recipe: "Hollow Gourd"
- Created item: ItemType "vessel", description includes variety (e.g., "spotted red vessel")

### Tasks
- [ ] Add `hollow-gourd` recipe to RecipeRegistry
- [ ] Implement vessel item creation (ItemType "vessel" + ContainerData)
- [ ] Implement Drop function (deferred from Feature 1):
  - Remove item from `char.Carrying`
  - Place item on map at character's position
  - Item retains `IsGrowing = false`
- [ ] Craft execution: check inventory, acquire gourd if needed, drop if blocked
- [ ] Post-craft: add to inventory or drop
- [ ] Tests for craft flow, inventory handling, drop mechanics

**Test checkpoint:** Order craft vessel, watch character acquire gourd, craft, and hold/drop result

---

## Feature 4: Vessel Contents

**Addresses:** Req B.2 (filling vessel), B.3 (look for container), B.4 (drop when blocked)

### Requirements
- Req B.2.i: Stack sizes - Mushrooms: 10, Berries: 20, Gourds: 1, Flowers: 10
- Req B.2.ii: Once variety added, only that variety can be added (vessel-specific)
- Req B.2.iii: Foraging fills vessel before stopping
- Req B.2.iv: Harvesting complete when vessel full OR no matching varieties
- Req B.3: If not carrying container, look for empty/matching vessel first
- Req B.4: Drop container when it blocks action

### Design Decision: Stack Slice
Contents tracked via `Contents []Stack` where each Stack has `Variety` + `Count`:
- Adding item: if empty, append new Stack; if same variety, increment count
- Eating: decrement count; remove Stack when count hits 0
- Vessel variety-lock: enforced because `Capacity == 1` means only one Stack allowed

### Tasks
- [ ] Add StackSize to item type definitions (config)
- [ ] Implement add-to-vessel logic (variety lock, count increment)
- [ ] Update foraging to fill vessel before stopping
- [ ] Update harvesting order completion logic
- [ ] Implement look-for-container behavior (Req B.3)
- [ ] Implement drop-when-blocked behavior (Req B.4)
- [ ] Tests for stacking, variety lock, forage/harvest with vessel

**Test checkpoint:** Forage with vessel, verify stacking and variety lock

---

## Feature 5: Eating from Vessels

**Addresses:** Req B.5 (eating from vessel)

### Requirements
- Req B.5: Vessel contents count as carried item for eating when hungry
- Eating decrements Stack count by 1
- When Stack empty, vessel accepts any variety again

### Tasks
- [ ] Update hunger intent to check vessel `Contents` for edible Stacks
- [ ] Implement eating from vessel (decrement count, apply effects from variety)
- [ ] Handle Stack removal when count = 0
- [ ] Apply poison/healing knowledge to vessel contents
- [ ] Tests for eating from vessel, knowledge application

### Open Design Questions
1. **Priority:** Loose carried item vs item in vessel - which eaten first?
2. **Poison avoidance:** Should characters refuse to eat known-poison items from vessels?

**Test checkpoint:** Character eats from vessel when hungry

---

## Feature 6: UI Updates

**Addresses:** Req C.1 (inventory panel), C.2 (map symbol)

### Requirements
- Req C.1: Carried vessels listed in inventory panel, showing contents
- Req C.2: Dropped vessels appear on map (symbol TBD)

### Tasks
- [ ] Update inventory panel rendering for vessels
- [ ] Choose and implement map symbol for dropped vessels
- [ ] Update item detail view for vessel contents

**Test checkpoint:** Verify all vessel states display correctly

---

## Recipe Timing Note

**Addresses:** Req D

Recipe time (2 minutes for hollow gourd vessel) is in game time. May revisit timing values after initial implementation.

---

## Quick Wins (Parallel Work)

- [x] Randomize starting names from curated list
- [ ] Remove single char mode from UI
- [ ] Add flag for character count control

---

## Post-Phase 6 Considerations

- **Time config reset:** Adjust so "day" = 2 game minutes
- **Carried item eating logic:** Re-evaluate with vessel context
- **Save/Load:** Verify nested item serialization works correctly

---

## Architecture Audit Points

1. **After Prep completion:** Review category system design before building on it
2. **After Phase 6 completion:** Full audit of new systems, test coverage review
