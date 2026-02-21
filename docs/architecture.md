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
| **Functional** | Edible (nil/non-nil), Poisonous, Healing | No | Capabilities, effects |

This separation is important for the preference/opinion system - characters form opinions about what things *are*, not what they *do*. Functional attributes are now in optional property structs (`EdibleProperties`, `PlantProperties`, `ContainerData`).

### Current Item Structure

```go
type Item struct {
    BaseEntity
    ID       int     // Unique identifier for save/load
    Name     string  // "Hollow Gourd" for crafted items, empty for natural items

    // Descriptive attributes (opinion-formable)
    ItemType string         // "berry", "mushroom", "gourd", "flower", "vessel"
    Color    types.Color
    Pattern  types.Pattern  // mushrooms, gourds, vessels
    Texture  types.Texture  // mushrooms, gourds, vessels

    // Optional property structs (nil if not applicable)
    Plant     *PlantProperties   // Growing/spawning behavior (nil for crafted items)
    Container *ContainerData     // Storage capacity (nil for non-containers)
    Edible    *EdibleProperties  // Edible items (nil for vessels, flowers)

    DeathTimer float64  // countdown until death (0 = immortal)
}

type EdibleProperties struct {
    Poisonous bool
    Healing   bool
}
```

**PlantProperties** controls spawning behavior and sprout state:
- `IsGrowing bool` — item can spawn new items of its variety. Set to false when picked up.
- `IsSprout bool` — item is in the sprout phase. Lifecycle skips reproduction for sprouts. Maturation logic converts sprout to full-grown item when `SproutTimer` expires.
- `SproutTimer float64` — countdown to maturation (see `config.SproutDuration`).

**ContainerData** enables storage - vessels have `Capacity: 1` (one stack). Contents tracked as `[]Stack` where each Stack has a Variety pointer and count. Liquids are stored as stacks with ItemType "liquid" and Kind (e.g., "water"), reusing all existing vessel infrastructure.

**EdibleProperties** marks items as edible with optional effects. Items with `Edible != nil` can be eaten; Poisonous/Healing determine effects.

### Adding New Plant Types

When adding a new plant type (like gourd was added in Phase 5):

1. `entity/item.go` - Add `NewX()` constructor
2. `config/config.go` - Add to `ItemLifecycle` map (spawn/death intervals)
3. `game/variety_generation.go` - Add variety generation logic
4. `game/world.go` - Add to `GetItemTypeConfigs()` for UI/character creation. Set `Plantable: true` if items of this type can be planted directly; set `CanProduceSeeds: true` if consuming produces seeds. These flags drive the Plant order UI sub-menu and the `IsOrderFeasible` check for plant orders.

Note: `spawnItem()` in `lifecycle.go` is generic - copies parent properties, no changes needed.

### Crafted Items

Crafted items (like vessels) differ from natural items:
- Have `Name` set (e.g., "Hollow Gourd") which overrides `Description()`
- Have `Plant = nil` (don't spawn/reproduce)
- Inherit appearance (Color, Pattern, Texture) from input materials
- May have `Container` for storage capability

### Kind Pattern for Subtypes

Some item types use a `Kind` subtype field for hierarchical identity:
- **ItemType**: broad product category (e.g., `"hoe"`, `"seed"`) — what the player orders or conceptually groups
- **Kind**: recipe- or origin-specific subtype (e.g., `"shell hoe"`, `"gourd seed"`) — what the character produces or what was consumed to create it
- `Description()` uses Kind when set, falls back to ItemType
- Seeds use this pattern: `ItemType: "seed"`, `Kind: "gourd seed"`, inheriting the parent gourd's full variety (Color, Pattern, Texture)
- `ItemVariety` also carries a `Kind` field (mirrors `Item.Kind`) so vessel contents can restore the correct Kind when items are extracted via `ConsumeAccessibleItem`
- Natural items leave Kind empty

### Future Extensions

When adding new item types (tools, materials):
- Add new functional flags as needed (Craftable, Wearable, Drinkable)
- Descriptive attributes remain the basis for preference formation
- Use `Name` field for crafted item display names

## Item Lifecycle

Items have lifecycle stages managed by timers and spawning systems.

### Plant-Based Spawning

Items with `Plant.IsGrowing = true` can spawn new items via `UpdateSpawnTimers()`. When total item count drops below target, spawn timers countdown. When they fire, `spawnItem()` creates a copy of a growing parent at a random empty adjacent tile. See `internal/system/lifecycle.go`.

Spawning is controlled by `ItemLifecycle` map in `config.go` which defines spawn/death intervals for each item type.

### Ground Spawning

Non-plant items (sticks, nuts, shells) spawn periodically via independent timers:

- **Sticks**: Fall from the canopy onto random empty tiles
- **Nuts**: Fall from the canopy onto random empty tiles
- **Shells**: Wash up adjacent to pond tiles

Each item type has its own timer (`GroundSpawnTimers` struct) with ~5 world day intervals (see `config.GroundSpawnInterval`). When a timer fires, one item spawns and the timer resets to a new random interval.

Ground spawning is count-independent - items continue spawning periodically regardless of how many exist in the world. This matches the simulation fiction of natural processes (canopy drops, tidal washing).

See `internal/system/ground_spawning.go`.

### Death Timers

Items with `DeathTimer > 0` decay over time via `UpdateDeathTimers()`. When the timer reaches zero, the item is removed from the world.

## Memory & Knowledge Model

Per BUILD CONCEPT in VISION.txt - history exists only in character memories and artifacts.

### ActionLog (Working Memory)

- Per-character recent events, bounded by count
- Displayed in UI, provides player visibility into character experience
- No omniscient world log - player sees aggregate of character experiences

### Knowledge System

Knowledge is discovered through experience and stored per-character:

```
Character eats poisonous mushroom
    → Takes damage, event logged
    → Gains knowledge: "Spotted Red Mushrooms are poisonous"
    → Knowledge may trigger opinion formation (dislike)
    → Knowledge can be shared via talking
```

Key distinction:
- **Poisonous** is an objective property on the Item (`Edible.Poisonous`)
- **Knowledge of poisonousness** is stored per-character, discovered through experience
- **Opinion about the item** may form based on the experience (separate from knowledge)

Characters track three types of learned information:
- `Knowledge []Knowledge` - Facts about items (poison, healing properties)
- `KnownActivities []string` - Activity IDs discovered (know-how like "harvest", "craftVessel")
- `KnownRecipes []string` - Recipe IDs learned (e.g., "hollow-gourd")

Knowledge transmission happens through talking - idle characters can share knowledge with each other.

### Future: Long-Term Memory

- Selective storage of notable events
- Persists until character death
- Basis for storytelling, artifact creation

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
4. **Variety fields**: `VarietySave` must include all fields that `ConsumeAccessibleItem` / `ConsumePlantable` need to restore (currently: `Kind`, `Plantable`, `Sym`)

Constructor-set fields like `Sym` won't be populated when deserializing directly into structs - must be explicitly restored based on type.

**PlantProperties serialization**: `IsSprout` and `SproutTimer` are saved/loaded as part of the plant save fields. Both must round-trip correctly to preserve sprout state across save/load.

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

### Self-Managing Actions

Some actions manage their full lifecycle across multiple phases without returning to idle activity selection. Each self-managing action owns its complete flow — procurement, movement, and execution — as a single continuous intent. The character never goes idle between phases and never needs to win another idle roll to continue.

**Key principle:** Vessel procurement within a multi-phase action should not be a separate intent routed through `ActionPickup`. It should be a *phase within the action's own handler*. The action type stays the same throughout (e.g., `ActionForage` or `ActionFillVessel`); the handler detects which phase it's in and acts accordingly.

**Current self-managing actions:**
- `ActionFillVessel` — Phase 1: procure empty vessel (if on ground). Phase 2: move to water and fill.
- `ActionForage` — Phase 1: procure vessel (optional, based on scoring). Phase 2: move to food and pick up.

**Shared procurement helper** (`picking.go`): Self-managing actions share a `RunVesselProcurement` tick helper rather than duplicating vessel procurement logic. Each tick, the action handler calls the helper; it returns `true` when the vessel is in hand (proceed to main phase), `false` if still working on procurement (moved or picked up this tick), or signals failure by nilling the intent. This keeps procurement logic in one place while each action owns its lifecycle.

```
Action handler pseudocode:
    if needsVessel && !hasVessel {
        ready := RunVesselProcurement(char, items, gameMap, log)
        if !ready { return }  // still procuring, or failed (intent nil)
    }
    // Main phase: do the real work
```

**Why not centralize in ActionPickup?** An earlier design (Option B) considered making `Pickup()` a pure helper and routing all post-pickup decisions through the `ActionPickup` handler. This was rejected because it creates a central dispatcher that must understand every context that triggers a pickup — the opposite of self-managing. It scales poorly as new multi-phase activities are added (Water Garden, future Construction activities). Self-managing actions distribute that knowledge to the actions that need it.

**Phase detection:** Handlers detect their current phase by checking world state (e.g., "is my target vessel on the ground or in inventory?") rather than storing explicit phase numbers. This is stateless and survives save/load without additional serialization.

### When to Use Self-Managing Actions vs ActionPickup

| Scenario | Action Type | Why |
|----------|-------------|-----|
| Multi-phase idle activity (forage with vessel, fetch water) | Self-managing (`ActionForage`, `ActionFillVessel`) | Needs continuous flow across phases without returning to idle selection |
| Simple one-off pickup (order prerequisite, single item) | `ActionPickup` | No continuation needed; re-selection on next tick handles follow-up |
| Order-driven harvest (fill vessel with food) | `ActionPickup` with order context | Continuation logic is in the ActionPickup handler, driven by `AssignedOrderID` |

**Adding a new self-managing action:**
1. Define the ActionType constant in `character.go`
2. Add the `applyIntent` handler in `update.go`
3. Use `RunVesselProcurement` (or future shared helpers) for procurement phases
4. Intent-finding function returns the new ActionType; handler owns all phases from there
5. Messaging is naturally contextual — the handler knows its own purpose

### Benefits

- No duplicate switch cases in `applyIntent` for same physical action
- No need to update exclusion lists when adding new triggering contexts
- Clean separation enables future multi-step orders (e.g., Craft: ActionPickup → ActionMove → ActionCraft)
- Self-managing actions enable complex autonomous behaviors without polluting intent system
- Shared procurement helpers prevent duplication across self-managing actions

### Code Structure

- `idle.go` - Orchestrates idle eligibility, calls `selectOrderActivity()` first
- `order_execution.go` - Order selection, assignment, intent finding, completion logic
- Order eligibility is generic: checks activity's `Availability` requirement against character's known activities
- `IsOrderFeasible(order, items, gameMap)` is computed on demand (not cached) at assignment time and render time — returns `(feasible bool, noKnowHow bool)`. Unfeasible orders are skipped during `findAvailableOrder`; the orders panel renders them dimmed with `[Unfulfillable]` or `[No one knows how]`. For plant orders, feasibility checks `item.Plantable && (item.ItemType == targetType || item.Kind == targetType)`, including contents of ground vessels.
- `LockedVariety string` on Order: set when the first item is planted. After locking, the character only seeks items of that exact variety, keeping a single order focused on one plant variety.
- **Unified order completion**: all order types (harvest, craft, till, plant) call `CompleteOrder()`, which sets `OrderCompleted` status. A sweep in the game loop after intent application removes all `OrderCompleted` orders that tick. Action handlers contain inline completion checks; `selectOrderActivity` and `findAvailableOrder` skip `OrderCompleted` orders as a safety net.

## Pickup Activity Code Organization

Picking up items is shared across multiple activities with different target selection, continuation, and completion criteria.

| File | Contents | Responsibility |
|------|----------|----------------|
| `picking.go` | `Pickup()`, `Drop()`, `DropItem()`, vessel helpers, `EnsureHasVesselFor()`, `EnsureHasRecipeInputs()`, `RunVesselProcurement()`, `findNearestItemByType()`, `ConsumePlantable()` | Physical pickup/drop, vessel operations, prerequisite helpers, procurement tick helpers, map search utilities |
| `foraging.go` | `findForageIntent()`, scoring functions | Foraging targeting and unified scoring. Returns `ActionForage` intent. |
| `fetch_water.go` | `findFetchWaterIntent()` | Fetch water targeting. Returns `ActionFillVessel` intent. |
| `order_execution.go` | `findHarvestIntent()`, `findCraftIntent()`, `findPlantIntent()` | Order-specific intent finding |
| `idle.go` | `selectIdleActivity()` | Calls foraging/fetch water as idle options |
| `update.go` | `applyIntent()` ActionPickup/ActionCraft cases | Executes actions, handles continuation |

### Pickup Result Pattern

`Pickup()` returns a `PickupResult` to distinguish outcomes:
- `PickupToInventory` - Item picked up directly (inventory was empty)
- `PickupToVessel` - Item added to carried vessel's stack
- `PickupFailed` - Could not pick up (variety mismatch with vessel)

Callers handle continuation differently based on result and context (foraging vs harvesting).

### Vessel Helper Functions (picking.go)

| Function | Purpose |
|----------|---------|
| `AddToVessel()` | Add item to vessel stack, returns false if can't fit |
| `IsVesselFull()` | Check if vessel stack at max capacity |
| `CanVesselAccept()` | Check if vessel can accept specific item (empty or matching variety) |
| `FindAvailableVessel()` | Find nearest vessel on ground that can hold target item |
| `CanPickUpMore()` | Check if character can pick up more (has room or has vessel with space) |
| `EnsureHasVesselFor()` | Returns intent to get compatible vessel, or nil if already have one |
| `EnsureHasItem()` | Returns intent to acquire a single item type, or nil if already carried |

### Accessible Item Helpers (character.go)

For actions that consume items from inventory or vessel contents:

| Function | Purpose |
|----------|---------|
| `HasAccessibleItem(itemType)` | Check if item exists in inventory OR inside carried vessel |
| `ConsumeAccessibleItem(itemType)` | Remove and return item from inventory or vessel contents |
| `ConsumePlantable(targetType, lockedVariety)` | Remove and return a plantable item from inventory or vessel contents, restoring `Kind`, `Plantable`, `Sym`, and `Edible` from the variety |

**Pattern**: Check availability at intent creation (no extraction), consume at action completion. This supports pause/resume - item stays accessible until actually consumed.

**Variety restoration on extraction**: When items are reconstructed from vessel stacks, `ConsumeAccessibleItem` and `ConsumePlantable` restore constructor-set fields (`Sym`, `Plantable`, `Kind`, `Edible`) from the variety. This is necessary because vessel stacks store variety references, not full item structs — direct struct reconstruction skips constructor logic. Serialization checklist: any new field set by a constructor must be listed in the variety and restored on extraction.

### Side Effects on Consumption (consumption.go)

`ConsumeFromInventory` and `ConsumeFromVessel` accept a `gameMap *game.Map` parameter to support side effects at the consumption site. Currently used to drop seeds when a gourd is consumed: after the item is removed from inventory or vessel, a seed of the same variety is created at the character's position and added to the map. This pattern generalizes to any future consumption side effect (e.g., leaving a core, shell, or husk).

### Look-for-Container Pattern

When foraging or harvesting without a vessel:
1. `FindAvailableVessel()` searches for empty or compatible vessel on ground
2. If found, intent targets vessel first
3. After vessel pickup, continues to harvest/forage into vessel
4. If no vessel, picks up item directly to inventory

## Water Terrain & Feature Passability

### Water Terrain

Water tiles (springs, ponds) are stored as map terrain (`water map[Position]WaterType`), not as features. This enables O(1) lookups and clean separation from the feature system.

| Water Type | Symbol | Interaction |
|------------|--------|-------------|
| WaterSpring | `☉` | Drink from cardinally adjacent tile (N/E/S/W) |
| WaterPond | `▓` | Drink from cardinally adjacent tile (N/E/S/W) |

Pond tiles render as a three-character fill `▓▓▓` (matching the terrain fill system used by tilled soil `═══`). Tiles 8-directionally adjacent to any water tile are "wet" — computed on the fly via `IsWet(pos)`, no persistent state.

Water tiles are impassable. Characters drink from cardinal-adjacent tiles, allowing multiple characters to drink from the same water source simultaneously (from different sides).

`FindNearestWater()` returns the closest water tile that has at least one available cardinal-adjacent tile. Returns `(Position, bool)` rather than a feature pointer.

### Pond Generation

`SpawnPonds()` generates 1-5 ponds of 4-16 contiguous water tiles each via blob growth (random cardinal expansion from a center tile). After placing all ponds, `isMapConnected()` verifies all non-water tiles are reachable via BFS. If the map is partitioned, all pond tiles are cleared and regenerated (max 10 retries). Ponds generate before features and items in the world generation order.

### Feature Passability

Features have a `Passable` boolean that controls movement:

| Feature | Passable | Interaction |
|---------|----------|-------------|
| Leaf Pile | Yes | Walk onto to sleep |

### Movement Blocking

`Map.IsBlocked(pos)` returns true if:
- A character occupies the position, OR
- A water tile is at the position, OR
- An impassable feature is at the position

`MoveCharacter()` checks these before allowing movement.

### Pathfinding

Two pathfinding strategies:

- **`NextStepBFS(fromX, fromY, toX, toY, gameMap)`**: BFS pathfinding that routes around permanent obstacles (water tiles, impassable features) while ignoring characters (temporary obstacles). Returns the first step toward the target. Falls back to greedy `NextStep` if no path exists. Used by all callers with gameMap access (intent continuation, drink/sleep/look/talk seeking, vessel/harvest seeking, pickup movement).

- **`NextStep(fromX, fromY, toX, toY)`**: Greedy single-step movement toward target along the larger axis delta. No obstacle awareness. Used as fallback and by callers without gameMap access.

- **`findAlternateStep()`**: Per-tick reactive routing around blocked tiles (characters and features). Used by `MoveCharacter` when the next step is occupied.

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
- `IsWater(pos)`, `WaterAt(pos)`, `IsWet(pos)`, `FindNearestWater(pos)`, `FindNearestBed(pos)`

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
    Category          string           // "craft", "garden", "" for uncategorized
    IntentFormation   IntentFormation  // automatic vs orderable
    Availability      Availability     // default vs knowhow
    DiscoveryTriggers []DiscoveryTrigger
}
```

| IntentFormation | Description | Examples |
|-----------------|-------------|----------|
| `IntentAutomatic` | Triggered by needs or idle selection | eat, drink, forage, look |
| `IntentOrderable` | Triggered by player orders | harvest, craftVessel, tillSoil |

| Availability | Description | Examples |
|--------------|-------------|----------|
| `AvailabilityDefault` | All characters can do it | eat, drink, forage |
| `AvailabilityKnowHow` | Must discover first | harvest, craftVessel, tillSoil |

**Category field**: Groups orderable activities for the order UI menu hierarchy. `getOrderableActivities()` generates synthetic category entries (e.g., "Craft", "Garden") when any known activity has that category. Uncategorized activities (e.g., Harvest) appear at the top level.

### Discovery Triggers

Know-how activities are discovered through experience:

```go
type DiscoveryTrigger struct {
    Action            ActionType  // ActionPickup, ActionLook, ActionConsume
    ItemType          string      // Specific type or empty for any
    RequiresEdible    bool        // Only trigger if item is edible
    RequiresPlantable bool        // Only trigger if item is plantable
}
```

Example: Harvest is discovered by picking up, eating, or looking at edible items. Plant is discovered by picking up or looking at plantable items (`RequiresPlantable: true`).

### Adding a New Activity

1. Add entry to `ActivityRegistry` in `entity/activity.go` (set `Category` for grouped activities)
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

Recipes are stored in `RecipeRegistry` map, keyed by recipe ID. Pattern supports multiple recipes per activity and per output type.

### Crafting Flow

1. Order assigned → `findCraftIntent` gets recipes for activity, filters to feasible recipes (inputs exist in world)
2. Picks first feasible recipe (future: preference-weighted selection)
3. `EnsureHasRecipeInputs` gathers missing components (drop non-recipe items, seek nearest input)
4. Once all components accessible: perform crafting action
5. On completion: consume all inputs, create output item via recipe-specific creation function

### Adding a New Recipe

1. Add recipe to `RecipeRegistry` in `entity/recipe.go`
2. Add activity to `ActivityRegistry` for the crafting activity (if new)
3. Add creation function in `system/crafting.go` (e.g., `CreateHoe`)
4. Add case to `applyIntent()` ActionCraft handler dispatch (by recipe ID)

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

### picking.go: Shared Home for Item Acquisition

picking.go is organized into three responsibility layers:

**1. Map Search Utilities:**
- `findNearestItemByType(cx, cy, items, itemType, growingOnly)` - Find nearest item by type, with optional growing-only filter
- `FindAvailableVessel(cx, cy, items, targetItem, registry)` - Find nearest vessel that can hold target item (i.e., can *receive* items)
- `FindVesselContaining(cx, cy, items, targetType, lockedVariety)` - Find nearest ground vessel whose *contents* match a plant target type/variety. Sibling of `FindAvailableVessel` used during plant order procurement.

**2. Prerequisite Orchestration:**
- `EnsureHasVesselFor(char, target, ...)` - Ensure character has compatible vessel or go get one (handles drop conflicts, availability checks)
- `EnsureHasRecipeInputs(char, recipe, ...)` - Ensure character has all recipe inputs or go get missing ones (handles drop logic, container protection)
- `EnsureHasItem(char, itemType, ...)` - Ensure character carries a single item type or go get one (simpler single-item variant for tools like the hoe)

**3. Physical Actions:**
- `Pickup()`, `Drop()`, `DropItem()` - Execute item transfers
- Vessel helpers: `AddToVessel()`, `CanVesselAccept()`, `IsVesselFull()`, `CanPickUpMore()`

### Future Vision: Unified Item-Seeking

Currently, foraging.go has preference-weighted item scoring (`scoreForageItems`) while picking.go's prerequisite helpers use nearest-distance via `findNearestItemByType`. These should converge — character item-seeking behavior should be consistent whether foraging, gathering craft components, or fulfilling orders. Preference shapes material culture (e.g., a character who prefers silver shells will craft silver shell hoes).

When preference-weighted component procurement is triggered (see triggered-enhancements.md), candidates to generalize into picking.go:
- `scoreForageItems` (foraging.go) → generic `scoreItemsByPreference` in picking.go
- `createPickupIntent` (foraging.go) → generic intent builder in picking.go (currently duplicated with different logging across foraging.go, order_execution.go, picking.go)

This consolidation makes picking.go the canonical answer to "how does a character go about acquiring an item" regardless of context.

## Feature Types

Features are map elements that aren't items or characters. They have fixed positions and provide interaction capabilities.

### Current Features

| Feature | Passable | Capability | Interaction |
|---------|----------|------------|-------------|
| Leaf Pile | Yes | Bed | Walk onto to sleep |

Note: Springs are now water terrain, not features. See "Water Terrain & Feature Passability" above.

### Feature Definition

```go
type Feature struct {
    BaseEntity
    FType       string  // "leafpile"
    Passable    bool
    DrinkSource bool  // legacy, kept for save migration
    Bed         bool
}
```

### Adding New Feature Types

1. Add constructor in `entity/feature.go`
2. Set appropriate capability flags
3. Add spawning logic in `game/spawning.go`
4. If impassable: pathfinding already handles via `IsBlocked()`
5. If new interaction type: add intent handling

## Area Selection UI Pattern

Area selection enables players to define rectangular regions for terrain modification (tilling, future construction). Designed for reusability.

### Pattern Overview

**State** (`internal/ui/model.go`):
- `ordersAddStep` uses step 2 for area selection mode
- `areaSelectAnchor *Position` tracks first corner (nil when unset)
- `areaSelectUnmarkMode bool` toggles mark/unmark behavior

**Validation** (`internal/ui/area_select.go`):
- `isValidTillTarget(pos, gameMap)` - position eligibility check (pluggable for other activities)
- `getValidPositions(anchor, cursor, gameMap, validator)` - rectangle filter with custom validator
- `isInRect(pos, corner1, corner2)` - rectangle bounds check

**Flow**:
1. Player selects activity (e.g., Garden > Till Soil) → enters step 2
2. Move cursor with arrow keys, press `p` to anchor
3. Move cursor to resize rectangle (valid tiles highlighted)
4. Press `p` to confirm (marks tiles / creates order)
5. Press `Tab` to toggle mark/unmark mode
6. Press `Enter` when done → returns to step 1 (activity selection)
7. Press `Esc` to cancel (clear anchor or exit)

**Rendering** (`internal/ui/view.go` `renderCell`):
- Rectangle highlight: full background for empty tiles, padding-only for entities (avoids ANSI nesting)
- Pre-highlight: shows existing marked tiles during area selection
- Background styles: `areaSelectStyle` (olive) for mark, `areaUnselectStyle` (dark red) for unmark

### Reuse for Future Activities

When adding new area-based orders (e.g., fence placement, building zones):
1. Add activity check in step 1 Enter handler (same as tillSoil)
2. Write activity-specific validator function (like `isValidTillTarget`)
3. Reuse `getValidPositions` with custom validator
4. Handle plot confirm logic in `p` key handler (activity-specific)

### Marked-for-Tilling Pool

Tilling separates plan (marked tiles) from work (orders):
- **Marked tiles** (`gameMap.markedForTilling map[Position]bool`): User's tilling plan, persistent, independent of orders
- **Till Soil orders**: Worker assignments. Multiple orders = multiple workers on the shared pool.
- Cancelling order removes worker, plan stays intact
- Unmarking tiles removes from plan (via area selection in unmark mode)

Pool is serialized in `SaveState.MarkedForTillingPositions`.

## Common Implementation Pitfalls

**Game time vs wall clock**: UI indicators that should work when paused (like "Saved" message) need wall clock time (`time.Now()`), not game time which only advances when unpaused.

**Sorting stability**: When displaying merged data from maps (e.g., AllEvents from ActionLog), use `sort.SliceStable` with deterministic tiebreakers (like CharID) to prevent visual jitter from Go's random map iteration order.

**View transitions**: When switching between views with different rendering approaches (game view uses direct rendering, menus use lipgloss.Place for centering), add dimension safeguards for edge cases.

**Save compatibility when changing entity storage**: When changing how entities are stored (e.g., moving data between fields, maps, or types), verify save/load round-trip as part of the same step. Check: (1) new state serializes, (2) old saves migrate, (3) serialize tests updated.

**Terrain fill in `renderCell()`**: Terrain that renders as solid blocks (tilled soil `═══`, water `▓▓▓`) requires both `sym` AND `fill` set to the styled terrain character. Setting only `sym` produces a single character flanked by spaces (` ▓ `), creating a vertical stripe appearance. Future terrain types (e.g., construction walls) should follow this pattern.
