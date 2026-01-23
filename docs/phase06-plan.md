# Phase 6: Containers and Storage - Implementation Plan

**Status:** Planning
**Requirements:** [docs/phase06reqs.txt](phase06reqs.txt)

---

## Overview

Characters can craft containers and store things in them. Items can be placed on the ground and picked up. This phase introduces crafting, the vessel item type, and item stacking.

---

## Phase 6 Prep: Category Type Formalization

**Trigger:** Adding non-plant categories (tools, crafted items) - from futureEnhancements.md

This foundational work should be completed before feature implementation to avoid retrofitting.

### Prep Tasks

- [ ] Create `PlantProperties` struct and refactor plant fields
  - `IsGrowing bool`: can reproduce at location (gates spawning)
  - `SpawnTimer float64`: countdown until next spawn
  - Move from flat Item fields into `Plant *PlantProperties`
  - On pickup: set `IsGrowing = false`
- [ ] Create `Stack` struct for container contents
  - `Variety *ItemVariety`: what variety this stack holds
  - `Count int`: how many in the stack
- [ ] Create `ContainerData` struct for containers
  - `Capacity int`: how many stacks it can hold
  - `Contents []Stack`: slice of stacks (supports future multi-stack containers)
  - Add as `Container *ContainerData` on Item (nil for non-containers)
  - Serialization: save slice of {variety ID, count} pairs, restore on load
- [ ] Remove `Category string` from Item
  - Derive plant status from `item.Plant != nil` where needed
- [ ] Update spawning logic
  - Check `item.Plant != nil && item.Plant.IsGrowing`
- [ ] Audit existing item creation
  - World-gen plants: `Plant = &PlantProperties{IsGrowing: true, SpawnTimer: ...}`
  - Crafted items: `Plant = nil`
- [ ] Keep on Item (broader applicability beyond plants)
  - `Edible`, `Poisonous`, `Healing` (future crafted consumables)
  - `DeathTimer` (future spoiling/degrading items)

### Deferred Fields (not needed for Phase 6)

- `IsPortable` on ContainerData - all items currently carriable
- `IsWatertight` on ContainerData - all vessels are watertight, defer until non-watertight containers
- `LockedVariety` on ContainerData - future vessel restriction

---

## Feature 1: Item Placement System

Characters can drop items; dropped items can be picked up.

### Requirements
- Characters can drop carried items
- Dropped items appear on map at character's location
- Dropped items have `IsGrowing = false` (set on pickup, stays false when dropped)
- Existing pickup logic works for dropped items
- Dropped items don't spawn new items

### Tasks
- [ ] Implement drop action/intent
- [ ] Update map rendering for dropped items
- [ ] Verify existing pickup logic handles dropped items
- [ ] Verify spawning correctly skips `IsGrowing = false` items (from Prep)
- [ ] Tests for drop/pickup cycle

### Testing Checkpoint
- Manual: Drop item, verify map display, pick it up again

---

## Feature 2: Crafting Foundation

General crafting system infrastructure.

### Requirements
- "Crafting" know-how that enables craft menu access
- Recipe system for defining craftable items
- Craft activity with duration
- Discovery triggers for crafting know-how

### Tasks
- [ ] Define Recipe struct (inputs, output, duration)
- [ ] Add "crafting" know-how to knowledge system
- [ ] Add discovery triggers (gourd interaction, spring interaction)
- [ ] Implement craft activity type
- [ ] Add craft order type to orders system
- [ ] UI: Craft option in order menu (gated by know-how)
- [ ] Tests for recipe system, know-how discovery

### Open Design Questions (Feature 2)

1. **Recipe storage:** Where do recipes live? Config? Separate registry?
2. **Recipe know-how:** Is knowing "crafting" enough to see all recipes, or must each recipe be discovered separately?

### Testing Checkpoint
- Manual: Character discovers crafting, craft menu appears

---

## Feature 3: Hollow Gourd Vessel

First craftable item: vessel from gourd.

### Requirements
- Recipe: 1 gourd -> 1 hollow gourd vessel, 2 minutes game time
- Character must have gourd in inventory to craft
- If no gourd in inventory: target gourd, drop current item if needed, pick up gourd, then craft
- Crafted vessel goes to inventory if room, else dropped

### Tasks
- [ ] Define hollow gourd vessel recipe
- [ ] Add "hollow gourd vessel recipe" know-how
- [ ] Implement vessel item type with `Container = &ContainerData{Capacity: 1}`
- [ ] Craft execution: check inventory, acquire gourd if needed
- [ ] Post-craft: add to inventory or drop
- [ ] Tests for craft flow, inventory handling

### Testing Checkpoint
- Manual: Order craft vessel, watch character acquire gourd and craft

---

## Feature 4: Vessel Contents

Vessels can hold stacks of items.

### Requirements
- Stack sizes by item type (Mushrooms: 10, Berries: 20, Gourds: 1, Flowers: 10)
- Vessel capacity: 1 stack (configurable per vessel type)
- Once an item variety is added, only that variety can be added
- Empty vessel accepts any variety
- Foraging fills vessel before stopping
- Harvesting completes when vessel full or no matching items remain

### Design Decision: Stack Slice
Contents tracked via `Contents []Stack` where each Stack has `Variety` + `Count`:
- Adding item: if empty, append new Stack; if same variety exists, increment count
- Eating: decrement count; remove Stack from slice when count hits 0
- Vessel variety-lock: enforced because `Capacity == 1` means only one Stack allowed
- Future containers with `Capacity > 1` can hold multiple different-variety Stacks

### Tasks
- [ ] Add StackSize to item type definitions (config or variety level)
- [ ] Implement add-to-vessel logic (variety lock, count increment)
- [ ] Update foraging to fill vessel before stopping
- [ ] Update harvesting order completion logic (full OR no matching items)
- [ ] Tests for stacking, variety lock, forage/harvest with vessel

### Testing Checkpoint
- Manual: Forage with vessel, verify stacking and variety lock

---

## Feature 5: Eating from Vessels

Hungry characters can eat from carried vessels.

### Requirements
- Vessel contents count as carried items for hunger intent
- Eating decrements Stack count by 1
- When Stack count hits 0, remove Stack from Contents (vessel accepts any variety again)
- Knowledge-based eating decisions apply to vessel contents

### Tasks
- [ ] Update hunger intent to check vessel `Contents` for edible Stacks
- [ ] Implement eating from vessel (decrement count, apply effects from variety)
- [ ] Handle Stack removal when count = 0
- [ ] Apply poison/healing knowledge to vessel contents
- [ ] Tests for eating from vessel, knowledge application

### Open Design Questions (Feature 5)

1. **Priority:** Loose carried item vs item in vessel - which eaten first?
2. **Poison avoidance:** Should characters refuse to eat known-poison items from vessels?

### Testing Checkpoint
- Manual: Character eats from vessel when hungry

---

## Feature 6: UI Updates

Visual feedback for vessels and contents.

### Requirements
- Inventory panel shows vessel contents (item, count)
- Dropped vessels visible on map (symbol TBD)
- Vessel contents visible in item details

### Tasks
- [ ] Update inventory panel rendering for vessels
- [ ] Choose and implement map symbol for dropped vessels
- [ ] Update item detail view for vessel contents
- [ ] Tests: N/A (UI rendering)

### Testing Checkpoint
- Manual: Verify all vessel states display correctly

---

## Quick Wins (Parallel Work)

Low-effort improvements from randomideas.txt, can be done between features:

- [x] Randomize starting names from curated list
- [ ] Remove single char mode from UI
- [ ] Add flag for character count control

---

## Post-Phase 6 Considerations

Items to revisit after Phase 6 stabilizes:

- **Time config reset:** Adjust so "day" = 2 game minutes
- **Carried item eating logic:** Re-evaluate with vessel context
- **Save/Load:** Verify nested item serialization works correctly

---

## Architecture Audit Points

1. **After Prep completion:** Review category system design before building on it
2. **After Phase 6 completion:** Full audit of new systems, test coverage review

---

## Reference

- [phase06reqs.txt](phase06reqs.txt) - Original requirements
- [futureEnhancements.md](futureEnhancements.md) - Deferred items and triggers
- [architecture.md](architecture.md) - System design patterns
