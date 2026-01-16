# Petri Architecture

## Key Design Patterns

- **MVU Architecture**: Bubble Tea handles rendering diffs automatically
- **Intent System**: Characters calculate Intent, then intents applied atomically (enables future parallelization)
- **Multi-Stat Urgency**: Tiers 0-4, highest wins, tie-breaker: Thirst > Hunger > Energy
- **Stat Fallback**: If intent can't be fulfilled, falls through to next urgent stat
- **Sparse Grid + Indexed Slices**: O(1) character lookups, separate slices for characters/items/features
- **Simple Flags over ECS**: Interaction capabilities use boolean flags (Edible, Poisonous) rather than full Entity Component System. Can evolve toward ECS later if needed.

## Data Flow

1. `Update()` → `updateGame()`
2. `UpdateSurvival()`: timers, stat changes, damage, sleep/wake
3. `CalculateIntent()`: evaluate tiers, try each in priority, track failures
4. `applyIntent()`: accumulate speed/progress, execute actions
5. `View()`: render UI (Bubble Tea diffs automatically)

## Item Model

### Attribute Classification

Items have two types of attributes with different roles:

| Attribute Type | Examples | Can form opinions? | Purpose |
|----------------|----------|-------------------|---------|
| **Descriptive** | ItemType, Color, Pattern, Texture | Yes | Identity, appearance |
| **Functional** | Edible, Poisonous, Healing | No | Capabilities, effects |

This separation is important for the preference/opinion system - characters form opinions about what things *are*, not what they *do*.

### Current Item Structure

```go
type Item struct {
    // Descriptive attributes (opinion-formable)
    ItemType  string         // "berry", "mushroom", "flower"
    Color     types.Color
    Pattern   types.Pattern  // mushrooms only
    Texture   types.Texture  // mushrooms only

    // Functional attributes (not opinion-formable)
    Edible    bool
    Poisonous bool
    Healing   bool
}
```

### Future Extensions

When adding new item types (tools, materials, crafted items):
- Add `Category` field if needed for organization
- Add new functional flags as needed (Craftable, Wearable, Drinkable)
- Descriptive attributes remain the basis for preference formation

## Memory & Knowledge Model

Per BUILD CONCEPT in VISION.txt - history exists only in character memories and artifacts.

### Current: ActionLog (Working Memory)

- Per-character recent events, bounded by count
- Displayed in UI, provides player visibility into character experience
- No omniscient world log - player sees aggregate of character experiences

### Future: Knowledge System (Phase 4)

Knowledge is discovered through experience and stored per-character:

```
Character eats poisonous mushroom
    → Takes damage, event logged
    → Gains knowledge: "Spotted Red Mushrooms are poisonous"
    → Knowledge may trigger opinion formation (dislike)
    → Future: Knowledge can be shared via talking
```

Key distinction:
- **Poisonous** is an objective property on the Item
- **Knowledge of poisonousness** is stored per-character, discovered through experience
- **Opinion about the item** may form based on the experience (separate from knowledge)

### Future: Long-Term Memory

- Selective storage of notable events
- Persists until character death
- Basis for knowledge transmission, storytelling, artifact creation

## Save/Load Serialization

Save files stored in `~/.petri/worlds/world-XXXX/` with:
- `state.json` - Full game state
- `state.backup` - Previous save (backup rotation)
- `meta.json` - World metadata for selection screen

### Serialization Checklist

When adding fields to saved structs, ensure ALL fields are included:

1. **Display fields**: Symbols (`Sym`), colors, styles set by constructors
2. **All attribute fields**: Easy to miss nested fields (e.g., Pattern/Texture on preferences)
3. **Round-trip tests**: Save → load → verify all fields match

Constructor-set fields like `Sym` won't be populated when deserializing directly into structs - must be explicitly restored based on type.

## Common Implementation Pitfalls

**Game time vs wall clock**: UI indicators that should work when paused (like "Saved" message) need wall clock time (`time.Now()`), not game time which only advances when unpaused.

**Sorting stability**: When displaying merged data from maps (e.g., AllEvents from ActionLog), use `sort.SliceStable` with deterministic tiebreakers (like CharID) to prevent visual jitter from Go's random map iteration order.

**View transitions**: When switching between views with different rendering approaches (game view uses direct rendering, menus use lipgloss.Place for centering), add dimension safeguards for edge cases.
