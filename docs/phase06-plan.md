# Phase 6: Containers and Storage - Implementation Plan

**Status:** Prep Stage 1 Complete - Ready for Prep Stage 2
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

### Prep Stage 2: Container Structs

- [ ] Create `Stack` struct
  - `Variety *ItemVariety`: what variety this stack holds
  - `Count int`: how many in the stack
- [ ] Create `ContainerData` struct
  - `Capacity int`: how many stacks it can hold
  - `Contents []Stack`: slice of stacks
- [ ] Add `Container *ContainerData` to Item (nil for non-containers)
- [ ] Update serialization (containers nil for existing items)

**Test checkpoint:** Save/load still works

### Design Decisions (Prep)

**Fields staying on Item** (broader applicability):
- `Edible`, `Poisonous`, `Healing` - future crafted consumables
- `DeathTimer` - future spoiling/degrading items

**Deferred fields** (not needed for Phase 6):
- `IsPortable` on ContainerData - all items currently carriable
- `IsWatertight` on ContainerData - all vessels are watertight
- `LockedVariety` on ContainerData - future vessel restriction

---

## Feature 1: Item Placement System

**Addresses:** OQ-A (picked plants don't respawn)

### Requirements
- Characters can drop carried items
- Dropped items appear on map at character's location
- Dropped items have `IsGrowing = false` (set on pickup, stays false when dropped)
- Existing pickup logic works for dropped items

### Tasks
- [ ] Implement drop action/intent
- [ ] Update map rendering for dropped items
- [ ] Set `IsGrowing = false` on pickup
- [ ] Verify spawning correctly skips `IsGrowing = false` items
- [ ] Tests for drop/pickup cycle

**Test checkpoint:** Drop item, verify map display, pick it up again

---

## Feature 2: Crafting Foundation

**Addresses:** Req A.1 (discovery), A.2 (orderable), OQ-B (crafting know-how structure)

### Requirements
- Req A.1.i: Discovery when engaging in gourd interaction (looking, picking up, consuming)
- Req A.1.ii: Discovery when engaging in spring interaction (drinking)
- Req A.2: Orderable via Craft option in task menu
- OQ-B: General "crafting" know-how + specific recipe know-how

### Tasks
- [ ] Define Recipe struct (inputs, output, duration)
- [ ] Add "crafting" know-how to knowledge system
- [ ] Add discovery triggers (gourd interaction, spring interaction)
- [ ] Implement craft activity type
- [ ] Add craft order type to orders system
- [ ] UI: Craft option in order menu (gated by know-how)
- [ ] Tests for recipe system, know-how discovery

### Open Design Questions
1. **Recipe storage:** Where do recipes live? Config? Separate registry?

**Test checkpoint:** Character discovers crafting, craft menu appears

---

## Feature 3: Hollow Gourd Vessel

**Addresses:** Req A.3 (recipe), B.1 (post-craft inventory)

### Requirements
- Req A.3: Recipe is 1 gourd for 2 minutes game time
- Req A.3.i: If gourd in inventory, begin crafting
- Req A.3.ii: If no gourd, target gourd to pick up; drop current item if inventory full
- Req B.1: Crafted item goes to inventory if room, else dropped

### Tasks
- [ ] Define hollow gourd vessel recipe
- [ ] Add "hollow gourd vessel recipe" know-how
- [ ] Implement vessel item type with `Container = &ContainerData{Capacity: 1}`
- [ ] Craft execution: check inventory, acquire gourd if needed
- [ ] Post-craft: add to inventory or drop
- [ ] Tests for craft flow, inventory handling

**Test checkpoint:** Order craft vessel, watch character acquire gourd and craft

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
