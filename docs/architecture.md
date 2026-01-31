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
    // Display name (crafted items override Description())
    Name      string         // "Hollow Gourd" for crafted items, empty for natural items

    // Descriptive attributes (opinion-formable)
    ItemType  string         // "berry", "mushroom", "gourd", "flower", "vessel"
    Color     types.Color
    Pattern   types.Pattern  // mushrooms, gourds, vessels
    Texture   types.Texture  // mushrooms, gourds, vessels

    // Functional attributes (not opinion-formable)
    Edible    bool
    Poisonous bool
    Healing   bool

    // Optional properties (nil if not applicable)
    Plant     *PlantProperties  // Growing/spawning behavior (nil for crafted items)
    Container *ContainerData    // Storage capacity (nil for non-containers)
}
```

**PlantProperties** controls spawning behavior - items with `Plant.IsGrowing = true` can spawn new items of their variety. When picked up, `IsGrowing` is set to false.

**ContainerData** enables storage - vessels have `Capacity: 1` (one stack). Contents tracked as `[]Stack` where each Stack has a Variety pointer and count.

### Adding New Plant Types

When adding a new plant type (like gourd was added in Phase 5):

1. `entity/item.go` - Add `NewX()` constructor
2. `config/config.go` - Add to `ItemLifecycle` map (spawn/death intervals)
3. `game/variety_generation.go` - Add variety generation logic
4. `game/world.go` - Add to `GetItemTypeConfigs()` for UI/character creation

Note: `spawnItem()` in `lifecycle.go` is generic - copies parent properties, no changes needed.

### Crafted Items

Crafted items (like vessels) differ from natural items:
- Have `Name` set (e.g., "Hollow Gourd") which overrides `Description()`
- Have `Plant = nil` (don't spawn/reproduce)
- Inherit appearance (Color, Pattern, Texture) from input materials
- May have `Container` for storage capability

### Future Extensions

When adding new item types (tools, materials):
- Add new functional flags as needed (Craftable, Wearable, Drinkable)
- Descriptive attributes remain the basis for preference formation
- Use `Name` field for crafted item display names

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

## Orders and Actions Pattern

Orders (player-directed tasks) and idle activities share physical actions but have different triggering contexts and completion criteria.

### Separation of Concerns

| Concept | Description | Example |
|---------|-------------|---------|
| **ActionType** | Physical action being performed | `ActionPickup`, `ActionCraft` |
| **Order/Activity Context** | Why the action is being performed | Foraging (idle) vs Harvest (order) |
| **Completion Criteria** | When the task is done | Foraging: after one pickup; Harvest: when inventory full |

### Implementation

- ActionTypes describe *what* is physically being done, not *why*
- Context is tracked via `char.AssignedOrderID` (0 = no order, non-zero = working on order)
- After an action completes, check context to determine next steps:

```go
// After ActionPickup completes:
result := system.Pickup(char, item, gameMap, actionLog)

// Handle based on result and context
switch result {
case PickupToVessel:
    // Continue filling vessel until full or no targets
    if nextTarget := FindNextVesselTarget(...); nextTarget != nil {
        char.Intent = newIntentFor(nextTarget)
    } else if char.AssignedOrderID != 0 {
        CompleteOrder(char, order, actionLog)
    }
case PickupToInventory:
    // Inventory full - complete order if on one
    if char.AssignedOrderID != 0 && !isVessel(char.Carrying) {
        CompleteOrder(char, order, actionLog)
    }
}
```

### Benefits

- No duplicate switch cases in `applyIntent` for same physical action
- No need to update exclusion lists when adding new triggering contexts
- Clean separation enables future multi-step orders (e.g., Craft: ActionPickup → ActionMove → ActionCraft)

### Code Structure

- `idle.go` - Orchestrates idle eligibility, calls `selectOrderActivity()` first
- `order_execution.go` - Order selection, assignment, intent finding, completion logic
- Order eligibility is generic: checks activity's `Availability` requirement against character's known activities

## Pickup Activity Code Organization

Picking up items is shared across multiple activities (foraging, harvesting) with different target selection and completion criteria. All use `ActionPickup` but have different triggering contexts.

| File | Contents | Responsibility |
|------|----------|----------------|
| `foraging.go` | `Pickup()`, `Drop()`, vessel helpers, foraging intent | Physical pickup, vessel logic, foraging targeting |
| `order_execution.go` | `findHarvestIntent()`, `findNearestItemByType()` | Harvesting target selection (by item type) |
| `idle.go` | `selectIdleActivity()` | Calls foraging as one idle option |
| `update.go` | `applyIntent()` ActionPickup case | Executes pickup, handles vessel continuation |

### Pickup Result Pattern

`Pickup()` returns a `PickupResult` to distinguish outcomes:
- `PickupToInventory` - Item picked up directly (inventory was empty)
- `PickupToVessel` - Item added to carried vessel's stack
- `PickupFailed` - Could not pick up (variety mismatch with vessel)

Callers handle continuation differently based on result and context (foraging vs harvesting).

### Vessel Helper Functions

| Function | Purpose |
|----------|---------|
| `AddToVessel()` | Add item to vessel stack, returns false if can't fit |
| `IsVesselFull()` | Check if vessel stack at max capacity |
| `CanVesselAccept()` | Check if vessel can accept specific item (empty or matching variety) |
| `FindAvailableVessel()` | Find nearest vessel on ground that can hold target item |
| `FindNextVesselTarget()` | Find next growing item matching vessel's variety |
| `CanPickUpMore()` | Check if character can pick up more (has room or has vessel with space) |
| `ConsumeFromVessel()` | Eat from vessel contents, decrement stack, apply effects from variety |

### Look-for-Container Pattern

When foraging or harvesting without a vessel:
1. `FindAvailableVessel()` searches for empty or compatible vessel on ground
2. If found, intent targets vessel first
3. After vessel pickup, continues to harvest/forage into vessel
4. If no vessel, picks up item directly to inventory

## Feature Passability Model

Features have a `Passable` boolean that controls movement:

| Feature | Passable | Interaction |
|---------|----------|-------------|
| Spring | No | Drink from cardinally adjacent tile (N/E/S/W) |
| Leaf Pile | Yes | Walk onto to sleep |

### Movement Blocking

`Map.IsBlocked(x, y)` returns true if:
- A character occupies the position, OR
- An impassable feature (like a spring) is at the position

`MoveCharacter()` checks `IsBlocked()` before allowing movement.

### Cardinal Adjacency for Impassable Features

For impassable features like springs:
- `isCardinallyAdjacent()` checks 4-direction adjacency (no diagonals)
- `findClosestCardinalTile()` finds nearest unblocked adjacent tile
- `FindNearestDrinkSource()` checks if any cardinal tile is available (not just if spring is "occupied")

This allows multiple characters to drink from the same spring simultaneously (from different sides).

### Pathfinding

`findAlternateStep()` uses `IsBlocked()` to route around both characters and impassable features.

## Common Implementation Pitfalls

**Game time vs wall clock**: UI indicators that should work when paused (like "Saved" message) need wall clock time (`time.Now()`), not game time which only advances when unpaused.

**Sorting stability**: When displaying merged data from maps (e.g., AllEvents from ActionLog), use `sort.SliceStable` with deterministic tiebreakers (like CharID) to prevent visual jitter from Go's random map iteration order.

**View transitions**: When switching between views with different rendering approaches (game view uses direct rendering, menus use lipgloss.Place for centering), add dimension safeguards for edge cases.
