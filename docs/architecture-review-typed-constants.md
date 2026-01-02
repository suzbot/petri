# Architecture Review: Typed Constants

**Date**: 2025-12-31
**Status**: Decisions Made - Ready to Implement
**Triggered by**: High priority audit items (docs/audit-findings.md)

## Background

The code audit identified hard-coded strings that should become typed constants:
- StatType: "hunger", "thirst", "energy" (52 occurrences)
- FoodType: "berry", "mushroom" (39 occurrences)
- Color: "red", "blue", "brown", "white" (80 occurrences)
- EventType: action log event types

## Decision to Pause and Review

Before implementing typed constants, we identified questions that suggest the current data model may need architectural review:

### 1. Health is also a stat

The audit missed that Health follows the same pattern as Hunger/Thirst/Energy:
- Has `HealthTier()` method
- Has `HealthLevel()` method
- Used in survival/damage calculations

StatType should include: hunger, thirst, energy, health, and future: mood

### 2. Phase 3 introduces Mood and dynamic preferences

From `docs/phase03reqs.txt`:
- Mood will be a new stat
- Preferences form about **entities AND attributes** of entities
- Preferences can be positive or negative, can change over time

This suggests we need flexible ways to reference "things characters can have opinions about."

### 3. Current Item model is too narrow

Current structure:
```go
type Item struct {
    ItemType  string  // "berry" or "mushroom"
    Color     string
    Poisonous bool
}
```

Problems:
- `ItemType` is really "food subtype" - assumes all items are food
- No way to represent non-food items (tools, materials, crafted items)
- No explicit `Edible` flag - edibility is implied by existence

Future items might need:
- Different interaction types (eat, drink, use, wear, craft with)
- Different categories (food, tool, material, clothing, container)
- Attributes that preferences can form about

### 4. Risk of typing the wrong abstractions

If we type `FoodType = "berry" | "mushroom"` now, we're encoding the assumption that ItemType == FoodType. This may require significant rework when we add non-food items.

## Decision

**Expand scope to review Item/Entity architecture before implementing typed constants.**

This prevents:
- Typing abstractions that will change
- Creating technical debt during Phase 3 implementation
- Multiple refactoring passes

## Questions for Architecture Review

### Items

1. Should Item have an `Edible bool` field?
2. Should `ItemType` become `Category` + `SubType`?
3. What other item categories will exist? (tools, materials, crafted, etc.)
4. What attributes can preferences form about? (color, type, taste, origin?)

### Stats

1. Is the current stat model (individual fields) sustainable?
2. Should stats be a map or slice for easier extensibility?
3. How does Mood differ from other stats? (affected by interactions vs. time)

### Preferences

1. How are preferences stored? (map of attribute -> sentiment?)
2. What can preferences target? (entity types, attributes, specific entities?)
3. How do preferences affect behavior? (food selection, movement, mood?)

## Research: Industry Patterns

Reviewed Entity Component System (ECS) patterns used in roguelikes and simulations:

**Sources:**
- [RogueBasin: Entity Component System](https://www.roguebasin.com/index.php/Entity_Component_System)
- [Rust Roguelike Tutorial: Items and Inventory](https://bfnightly.bracketproductions.com/chapter_9.html)
- [ECS FAQ](https://github.com/SanderMertens/ecs-faq)
- [How RimWorld fleshes out the Dwarf Fortress formula](https://www.gamedeveloper.com/design/how-i-rimworld-i-fleshes-out-the-i-dwarf-fortress-i-formula)

**Key findings:**

1. ECS is the dominant pattern - entities are IDs, behaviors come from attached components
2. Simple boolean flags for capabilities is a valid approach (used in Rust roguelike tutorial)
3. RimWorld's philosophy: "create a simulation that is not overly complex, and that a player can observe and comprehend"
4. Full ECS can be evolved toward later if simple flags become insufficient

## Decisions Made

### Item Model: Simple Flags Approach

**Decision:** Use interaction flags (booleans) rather than full ECS components.

**Rationale:**
- Simpler to implement and understand
- Can evolve toward full ECS later if needed
- Aligns with RimWorld's "comprehensible simulation" philosophy

### Attribute Classification

**Decision:** Structurally separate descriptive attributes from functional attributes.

| Attribute Type | Examples | Can form opinions? |
|----------------|----------|-------------------|
| Descriptive | ItemType, Color, Material | Yes |
| Functional | Edible, Craftable, Wearable | No |

This distinction is important for the Phase 3 preference system.

### Category Field

**Decision:** Category (food, tool, material) is helpful for organization but not required immediately. Can be added when non-food items are introduced.

### Poisonous as Discovered Knowledge

**Decision:** Poisonous is an objective property on the Item. Character *knowledge* of poisonousness is stored in their Memory (Phase 3).

Example flow:
1. Character eats red mushroom (Poisonous=true)
2. Character gets sick (event logged)
3. Event has chance of forming opinion: "dislikes red mushrooms" OR "dislikes mushrooms" OR "dislikes red things"
4. Opinion stored in character's Memory, influences future decisions

## Approved Item Model

```go
type Item struct {
    BaseEntity

    // Descriptive attributes (opinion-formable)
    ItemType  string  // "berry", "mushroom", "stick", etc.
    Color     Color
    // Future: Material, Pattern, Origin

    // Functional attributes (not opinion-formable)
    Edible    bool
    Poisonous bool
    // Future: Craftable, Wearable, Drinkable
}
```

## Implementation Progress

**Location:** Types defined in `internal/types/types.go` (not config, to avoid bloat)

**Completed:**
- Created `internal/types/types.go` with Color type and constants
- Added `Edible` field to Item struct
- Updated Item struct comments to reflect descriptive vs functional attribute separation
- Updated all production files to use `types.Color`: config.go, entity/item.go, entity/character.go, game/world.go, ui/model.go, ui/update.go, ui/view.go, simulation/simulation.go
- Updated all test files: entity/item_test.go, entity/character_test.go, game/map_test.go, system/survival_test.go, system/consumption_test.go, system/movement_test.go
- Note: ui/creation.go intentionally uses title-case display strings ("Red", "Blue") for UI purposes; mapping to types.Color happens in ui/update.go

**Remaining:**
- Defer Category until non-food items are needed

**Completed (2025-12-31):**
- `StatType` with hunger, thirst, energy, health constants
- Updated `Intent.DrivingStat` from `string` to `types.StatType`
- Updated all switch statements and comparisons in movement.go
- Updated all tests in movement_test.go

**Completed (2026-01-01):**
- Consolidated tier/level calculation logic in `internal/entity/character.go`
- Created `StatThresholds` struct with Mild, Moderate, Severe, Crisis boundaries and Inverted flag
- Created `StatLevels` struct with None, Mild, Moderate, Severe, Crisis descriptions
- Created generic `calculateTier()` function replacing 4 separate tier calculation methods
- Created `forTier()` method on StatLevels replacing 4 separate level description methods
- All tier methods (HungerTier, ThirstTier, EnergyTier, HealthTier) are now one-liners
- All level methods (HungerLevel, ThirstLevel, EnergyLevel, HealthLevel) are now one-liners
- Adding Mood stat in Phase III will require only 2 new data declarations, not duplicating logic

---

## Deferred: EventType for ActionLog

**Date:** 2025-12-31
**Decision:** Defer implementation

### Analysis

The audit (Finding 4) identified EventType as high priority alongside StatType. However, closer analysis reveals a different risk profile.

**Current event types fall into two categories:**

| Stat-related (overlap with StatType) | Non-stat categories |
|--------------------------------------|---------------------|
| "hunger", "thirst", "energy", "health" | "activity", "movement", "consumption", "poison", "sleep", "death" |

**Options considered:**

1. **Reuse StatType for stat events, add EventType for others** - Stat-related events could use `string(types.StatHunger)` etc. Only add EventType for non-stat categories.

2. **Create separate EventType with all categories** - Cleaner API, but introduces redundancy with StatType values.

3. **Defer** - Current risk is low.

### Reasoning

Unlike StatType (used in switch statements that control behavior), EventType is purely cosmetic:
- Event types are only used for logging/display, not logic
- No filtering/categorization code exists yet
- Typos would show up visibly in the UI immediately

The cost/benefit doesn't justify implementation now.

### Trigger Points

See CLAUDE.md "Deferred Enhancements & Trigger Points" for the consolidated list. Summary:
- Event filtering added to UI
- Event-driven behavior (Phase III)
- Event persistence
- Event type proliferation

### Implementation Notes (for future reference)

If implementing, consider:
- Stat-related events could cast from StatType: `EventType(types.StatHunger)`
- Or define EventType separately with all categories for API clarity
- ActionLog.Add signature would change: `Add(charID int, charName string, eventType EventType, message string)`
