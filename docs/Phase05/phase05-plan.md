# Phase 5 Plan: Picking up Items and Inventory

## Overview

Phase 5 introduces inventory, know-how (activity knowledge), discovery mechanics, and user-ordered activities. This enables characters to gather items intentionally and lays groundwork for future crafting and job systems.

**Source:** [phase05reqs.txt](phase05reqs.txt)

---

## Architecture Decisions

1. **Activity definition**: Formalize now with `Activity` struct and registry (tech tree planned). Structure:
   - ID, Name (display)
   - IntentFormation (automatic vs orderable)
   - Availability (default vs know-how)
   - DiscoveryTriggers (type TBD - discuss when approaching 5.4)
2. **Know-how vs Fact Knowledge**: Separate `KnowHow` struct (distinct from `Knowledge`)
3. **Order storage**: On `World` struct for persistence across save/load
4. **Foraging preference weights**: Same as eating (reuse existing config values)
5. **Harvest know-how starting state**: All characters must discover it (none start with it)
6. **Gourd symbol**: 'G'

---

## Sub-Phase Breakdown

### 5.1: Item Category & Gourd
**Requirements:** A.1-A.8

**User-facing outcome:** Gourds spawn in the world with varied colors, patterns, and textures. Gourds appear as a food option during character creation. Characters can eat, look at, and form preferences about gourds.

**Scope:**
- Add `Category` field to Item (string for now, "plant" for spawning items)
- Add new patterns to AllPatterns: striped, speckled; define GourdPatterns subset
- Add new texture to AllTextures: warty; define GourdTextures subset
- Add ColorGreen to AllColors; define GourdColors subset
- Create gourd item type (edible, never poison/healing)
- Add gourd to variety generation
- Refactor character creation to derive food options from edible item types (not hardcoded)

**Decisions:**
- Green color: terminal color "34"
- Patterns/Textures: Generalized list with per-item-type subsets (like colors)
- Category: String for now; trigger for formalization noted in futureEnhancements.md
- Food options: Derived dynamically from GetItemTypeConfigs() where Edible=true

**[TEST]:** Start new world, verify gourds spawn with various attributes. Create character with gourd as favorite food. Verify characters eat gourds and can form preferences.

---

### 5.2: Inventory & Foraging
**Requirements:** B.1-B.4, C.1-C.2

**User-facing outcome:** Characters have inventory (capacity: 1 item). Player can view inventory with 'i' key. Idle characters sometimes forage - walking to a nearby edible item and picking it up based on preference/distance scoring. Foraging unavailable when inventory full.

**Scope:**
- Add `Carrying *Item` to Character
- Add `ActionPickup` action type
- Implement pickup execution (remove from map, add to inventory)
- Add 'i' key for inventory panel view
- Add inventory serialization for save/load
- Add foraging to `selectIdleActivity()` options
- Implement `findForageIntent()` using preference/distance scoring (same weights as eating)
- Character moves to item, then picks up (ActionPickup)
- Foraging excluded from idle options when inventory full

**[TEST]:** Run simulation, observe characters foraging during idle time. Verify they pick up items (not eat). Press 'i' to view inventory. Verify foraging stops when carrying an item. Save/load and verify inventory persists.

**Implementation Notes (5.2):**
- Character.Carrying field added (`*Item`, nil = empty)
- ActionPickup constant added to ActionType enum
- IsInventoryFull() helper method added to Character
- Serialization: CharacterSave.Carrying field (`*ItemSave`, omitempty), timers set to 0 for carried items
- itemFromSave() updated to handle "gourd" item type
- UI: showInventoryPanel bool in Model, 'i'/'k' mutually exclusive toggles
- renderInventoryPanel() shows "Carrying: [description]" or "Carrying: nothing"
- Foraging: 4th idle activity at 25%, uses findForageIntent() with preference/distance scoring (same weights as eating)
- Foraging logs: "Foraging for [item type]" on start, "[name] picked up [item]" on pickup

---

### 5.3: Eating from Inventory
**Requirements:** D.1-D.3

**User-facing outcome:** Characters who are mildly hungry (or more) and carrying an edible item will eat from inventory. Eating from inventory has standard consumption effects. Empty inventory re-enables foraging.

**Scope:**
- Modify hunger intent to check inventory first
- Carried edible treated as "closer than any map item" in scoring
- Consume from inventory applies standard effects (preference formation, poison/healing, knowledge)
- Clear inventory slot after consumption

**[TEST]:** Let character forage, then get hungry. Verify they eat carried item before seeking map items. Verify effects apply normally.

---

### 5.4: Know-how System
**Requirements:** E.1-E.5

**User-facing outcome:** Characters can discover "Harvest" know-how through experience. Know-how is displayed in character knowledge panel. Know-how cannot be transmitted via talking.

**Scope:**
- Create `KnowHow` struct (separate from `Knowledge`)
- Add `KnowHow []KnowHow` to Character
- Define Harvest know-how
- Implement discovery chance when: foraging, eating edible, looking at edible
- Display know-how in knowledge panel
- Exclude know-how from `TransmitKnowledge()`

**[TEST]:** Run simulation until a character discovers Harvest. Verify it appears in knowledge panel. Verify talking does NOT transmit know-how.

---

### 5.5: Orders System - Data & UI
**Requirements:** F.1-F.4, F.6

**User-facing outcome:** Player can press 'O' to view Orders panel. Panel shows list of orders with status. Player can add new Harvest orders by selecting an edible item type. Player can cancel orders.

**Scope:**
- Create `Order` struct with: ID, Activity, TargetType, Status, AssignedTo
- Add `Orders []*Order` to World
- Add 'O' key for Orders panel
- Implement order list display (activity, target, status)
- Implement "add order" flow:
  - Show orderable activities (known by at least one living character)
  - For Harvest: select from edible item types
- Implement order cancellation (player removes order from list)
- Add orders serialization

**[TEST]:** Press 'O' to see empty orders panel. After a character has Harvest know-how, add a Harvest order. Verify order appears in list as "open". Cancel order, verify it's removed.

---

### 5.6: Orders System - Execution
**Requirements:** F.5, F.7-F.11

**User-facing outcome:** Characters with Harvest know-how take open orders when idle. They seek specific item types, pick them up, and complete order when inventory full. Orders can be interrupted by needs and resumed. Unfulfillable orders are abandoned.

**Scope:**
- Add `AssignedOrder *Order` to Character
- Implement order assignment: idle + has know-how + open order → assign
- Implement `findHarvestIntent()`: seek specific item type, pick up
- Order interruption: moderate+ needs pause order, resume when satisfied
- Order abandonment: no matching items on map → remove order
- Order completion: inventory full → remove order
- Activity log entries for: taking order, pausing, resuming, completing, abandoning

**[TEST]:**
1. Create Harvest order for gourds
2. Character with know-how takes order (shown in activity log)
3. Character seeks and picks up gourd
4. Order completes when inventory full
5. Test interruption: make character thirsty, verify order pauses
6. Test abandonment: order item type with none on map

---

## Deferred/Future Considerations

- **Mood impacted by carrying**: Noted in vision, not in Phase 5 requirements. Defer to future phase.
- **Multiple inventory slots**: Current capacity is 1. Expand when needed.

---

## Dependencies

```
5.1 Category & Gourd (independent)
         │
         ▼
5.2 Inventory & Foraging
         │
         ├─────────────────┐
         ▼                 ▼
5.3 Eat from Inventory   5.4 Know-how
                           │
                           ▼
                      5.5 Orders UI
                           │
                           ▼
                      5.6 Orders Execution
```

---

## Feature-Specific Questions (to address during implementation)

### 5.2
- Foraging selection probability vs looking/talking? **Decision: 4th option at 25% each (look/talk/idle/forage)**
- Empty inventory display? **Decision: "Carrying: nothing"**
- Action log entries? **Decision: Log both "starting to forage" and "picked up [item]"**
- Movement mechanics? **Decision: Multi-tick move-then-act like eating**
- Panel toggling? **Decision: 'i' and 'k' are mutually exclusive (pressing one closes the other)**
- Carried item timers? **Decision: Carried items become static (no spawn/death timers). Timers cleared on pickup. Future planting feature may re-enable timers, but only when planted.**

### 5.4
- Discovery probability per trigger event?
- DiscoveryTriggers type: strings, activity references, or something else?

### 5.5
- Order panel replaces which existing panels? (Req says "details/log panels")

### 5.6
- When order is paused and character resumes, do they continue same order or re-evaluate?
