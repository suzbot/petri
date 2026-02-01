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
    // Only continue filling for orders, not autonomous foraging
    if char.AssignedOrderID != 0 {
        if nextTarget := FindNextVesselTarget(...); nextTarget != nil {
            char.Intent = newIntentFor(nextTarget)
        } else {
            CompleteOrder(char, order, actionLog)
        }
    }
    // Autonomous foraging stops after one item
case PickupToInventory:
    // Check if order should continue (harvest fills both slots)
    if char.AssignedOrderID != 0 {
        if nextTarget := FindNextHarvestTarget(...); nextTarget != nil {
            char.Intent = nextTarget
        } else {
            CompleteOrder(char, order, actionLog)
        }
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

## Position Handling

All coordinates use `types.Position` struct:

```go
type Position struct {
    X, Y int
}
```

### Entity Positions

- Get position: `entity.Pos()` returns `types.Position`
- Set position: `entity.SetPos(pos)`
- Direct field access: `entity.X`, `entity.Y` (for performance-critical code)

### Position Methods

| Method | Purpose |
|--------|---------|
| `pos.DistanceTo(other)` | Manhattan distance between positions |
| `pos.IsAdjacentTo(other)` | True if within 1 tile (8 directions) |
| `pos.IsCardinallyAdjacentTo(other)` | True if exactly 1 tile away (N/E/S/W only) |
| `pos.NextStepToward(target)` | Position one step closer to target |

### Helper Functions

| Function | Location |
|----------|----------|
| `types.Abs(x)` | Absolute value |
| `types.Sign(x)` | Returns -1, 0, or 1 |

### Map Query Methods

All Map query methods take `types.Position`:
- `EntityAt(pos)`, `CharacterAt(pos)`, `ItemAt(pos)`, `FeatureAt(pos)`
- `IsValid(pos)`, `IsOccupied(pos)`, `IsBlocked(pos)`, `IsEmpty(pos)`
- `FindNearestDrinkSource(pos)`, `FindNearestBed(pos)`

### Guidelines

**Do:**
- Use `pos.DistanceTo(other)` for distance calculations
- Use `pos.IsAdjacentTo(other)` for adjacency checks
- Create Position with `types.Position{X: x, Y: y}` when needed

**Don't:**
- Inline distance calculations like `abs(x1-x2) + abs(y1-y2)`
- Create new position-like structs
- Define local `abs()` or `sign()` functions

## Activity Registry & Know-How Discovery

Activities are defined in `ActivityRegistry` with properties that control availability and discovery.

### Activity Definition

```go
type Activity struct {
    ID                string
    Name              string
    IntentFormation   IntentFormation  // automatic vs orderable
    Availability      Availability     // default vs knowhow
    DiscoveryTriggers []DiscoveryTrigger
}
```

| IntentFormation | Description | Examples |
|-----------------|-------------|----------|
| `IntentAutomatic` | Triggered by needs or idle selection | eat, drink, forage, look |
| `IntentOrderable` | Triggered by player orders | harvest, craftVessel |

| Availability | Description | Examples |
|--------------|-------------|----------|
| `AvailabilityDefault` | All characters can do it | eat, drink, forage |
| `AvailabilityKnowHow` | Must discover first | harvest, craftVessel |

### Discovery Triggers

Know-how activities are discovered through experience:

```go
type DiscoveryTrigger struct {
    Action         ActionType  // ActionPickup, ActionLook, ActionConsume
    ItemType       string      // Specific type or empty for any
    RequiresEdible bool        // Only trigger if item is edible
}
```

Example: Harvest is discovered by picking up, eating, or looking at edible items.

### Adding a New Activity

1. Add entry to `ActivityRegistry` in `entity/activity.go`
2. If orderable: add case to `findOrderIntent()` switch in `order_execution.go`
3. Add intent-finding function (e.g., `findHarvestIntent()`)
4. If new ActionType needed: add to enum in `character.go`, handle in `applyIntent()`

## Recipe System

Recipes define how to craft items from components.

### Recipe Definition

```go
type Recipe struct {
    ID         string
    Name       string           // Display name for crafted item
    Inputs     []RecipeInput    // Required components
    OutputType string           // ItemType of result
    Duration   float64          // Crafting time
}

type RecipeInput struct {
    ItemType string
    Count    int
}
```

### RecipeRegistry

Recipes are stored in `RecipeRegistry` map, keyed by recipe ID. Currently one recipe (`hollow-gourd`), but pattern supports multiple recipes per output type.

### Crafting Flow

1. Order assigned → check for required components in inventory
2. If missing components: drop incompatible items, seek components
3. Once all components gathered: perform crafting action
4. On completion: consume inputs, create output item

### Adding a New Recipe

1. Add recipe to `RecipeRegistry` in `entity/recipe.go`
2. Add activity to `ActivityRegistry` for the crafting activity
3. Add intent-finding function (follows component procurement pattern)
4. Handle in `applyIntent()` ActionCraft case

## Component Procurement Pattern

Many activities require gathering specific items before performing the action. This pattern is used by harvest, craft, and will be used by gardening activities.

### Flow

```
Order Assigned
    ↓
Check inventory for required items
    ↓
┌── Has all components? ──┐
│                         │
Yes                       No
│                         │
Begin activity    Drop non-components
                         │
                  Seek nearest component
                  (preference/distance weighted)
                         │
                  ┌── Found? ──┐
                  │            │
                  Yes          No
                  │            │
            Move & pickup   Abandon order
                  │
            Check inventory again
            (loop until complete)
```

### Implementation Notes

- Preference/distance weighting uses `NetPreference()` and `DistanceTo()`
- Vessel-seeking is a special case: find vessel that can hold target item
- Drop logic: `Drop()` places item on ground at character position
- Abandonment: `abandonOrder()` clears assignment, resets order status

### Future: picking.go Consolidation

The vessel-seeking pattern (~70 lines) is duplicated between `findForageIntent()` and `findHarvestIntent()`. This will be consolidated into `picking.go` with a `PickupIntentBuilder` that encapsulates the pattern with configurable target-finding and conflict-resolution strategies.

## Feature Types

Features are map elements that aren't items or characters. They have fixed positions and provide interaction capabilities.

### Current Features

| Feature | Passable | Capability | Interaction |
|---------|----------|------------|-------------|
| Spring | No | DrinkSource | Drink from adjacent tile |
| Leaf Pile | Yes | Bed | Walk onto to sleep |

### Feature Definition

```go
type Feature struct {
    BaseEntity
    FType       string  // "spring", "leafpile"
    Passable    bool
    DrinkSource bool
    Bed         bool
}
```

### Adding New Feature Types

1. Add constructor in `entity/feature.go` (e.g., `NewPond()`)
2. Set appropriate capability flags
3. Add spawning logic in `game/spawning.go`
4. If impassable: pathfinding already handles via `IsBlocked()`
5. If new interaction type: add intent handling

### Future: Tile States

Gardening introduces tile states (tilled soil) that modify ground behavior. This may evolve into a tile property system separate from features:
- Tilled soil: affects plant growth rate
- Water tiles (ponds): impassable, allow drinking from adjacent

## Common Implementation Pitfalls

**Game time vs wall clock**: UI indicators that should work when paused (like "Saved" message) need wall clock time (`time.Now()`), not game time which only advances when unpaused.

**Sorting stability**: When displaying merged data from maps (e.g., AllEvents from ActionLog), use `sort.SliceStable` with deterministic tiebreakers (like CharID) to prevent visual jitter from Go's random map iteration order.

**View transitions**: When switching between views with different rendering approaches (game view uses direct rendering, menus use lipgloss.Place for centering), add dimension safeguards for edge cases.
