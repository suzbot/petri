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
| **Descriptive** | ItemType, Color, Pattern, Texture, Kind | Yes | Identity, appearance |
| **Functional** | Edible (nil/non-nil), Poisonous, Healing, Plantable | No | Capabilities, effects |

This separation supports the preference/opinion system — characters form opinions about what things *are*, not what they *do*. It reflects **Source of Truth Clarity** (see Values.md): descriptive attributes are the ground truth for identity, functional attributes are the ground truth for behavior, and these concerns don't cross.

Functional attributes live in optional property structs (`EdibleProperties`, `PlantProperties`, `ContainerData`) — nil when not applicable. See `entity/item.go` for the full Item struct definition.

**PlantProperties** controls spawning behavior and sprout state:
- `IsGrowing bool` — item can spawn new items of its variety. Set to false when picked up.
- `IsSprout bool` — item is in the sprout phase. Lifecycle skips reproduction for sprouts. Maturation logic converts sprout to full-grown item when `SproutTimer` expires.
- `SproutTimer float64` — countdown to maturation (see `config.SproutDuration`).

**ContainerData** enables storage — vessels have `Capacity: 1` (one stack). Contents tracked as `[]Stack` where each Stack has a Variety pointer and count. Liquids are stored as stacks with ItemType "liquid" and Kind (e.g., "water"), reusing all existing vessel infrastructure.

**EdibleProperties** marks items as edible with optional effects. Items with `Edible != nil` can be eaten; Poisonous/Healing determine effects.

### Kind Pattern for Subtypes

Some item types use a `Kind` subtype field for hierarchical identity:
- **ItemType**: broad product category (e.g., `"hoe"`, `"seed"`) — what the player orders or conceptually groups
- **Kind**: recipe- or origin-specific subtype (e.g., `"shell hoe"`, `"gourd seed"`) — what the character produces or what was consumed to create it
- `Description()` uses Kind when set, falls back to ItemType
- Natural items leave Kind empty

This follows **Design Types for Future Siblings** (Values.md) — `ItemType: "liquid", Kind: "water"` accommodates future beverages (mead, beer, wine) as Kind variants with no structural changes needed.

`ItemVariety` also carries a `Kind` field (mirrors `Item.Kind`) so vessel contents can restore the correct Kind when items are extracted. See "Variety Restoration on Extraction" below.

### Crafted Items

Crafted items (like vessels) differ from natural items:
- Have `Name` set (e.g., "Hollow Gourd") which overrides `Description()`
- Have `Plant = nil` (don't spawn/reproduce)
- Inherit appearance (Color, Pattern, Texture) from input materials
- May have `Container` for storage capability

### Adding New Plant Types

1. `entity/item.go` — Add `NewX()` constructor
2. `config/config.go` — Add to `ItemLifecycle` map (spawn/death intervals); if edible, add to `ItemMealSize` map (use `GetMealSize` to look up satiation and eating duration); add to `SproutDurationTier` map (maturation speed tier)
3. `game/variety_generation.go` — Add variety generation logic
4. `game/world.go` — Add to `GetItemTypeConfigs()` for UI/character creation. Set `Plantable: true` if items of this type can be planted directly; set `CanProduceSeeds: true` if consuming produces seeds.

Note: `spawnItem()` in `lifecycle.go` is generic — copies parent properties, no changes needed.

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

Each item type has its own timer (`GroundSpawnTimers` struct) with intervals defined in `config.GroundSpawnInterval`. When a timer fires, one item spawns and the timer resets.

Ground spawning is count-independent — items continue spawning regardless of how many exist. This matches the simulation fiction of natural processes (canopy drops, tidal washing).

See `internal/system/ground_spawning.go`.

### Death Timers

Items with `DeathTimer > 0` decay over time via `UpdateDeathTimers()`. When the timer reaches zero, the item is removed from the world.

## Memory & Knowledge Model

Per BUILD CONCEPT in VISION.txt — history exists only in character memories and artifacts.

### ActionLog (Working Memory)

- Per-character recent events, bounded by count
- Displayed in UI, provides player visibility into character experience
- No omniscient world log — player sees aggregate of character experiences

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
- `Knowledge []Knowledge` — Facts about items (poison, healing properties)
- `KnownActivities []string` — Activity IDs discovered (know-how like "harvest", "craftVessel")
- `KnownRecipes []string` — Recipe IDs learned (e.g., "hollow-gourd")

Knowledge transmission happens through talking — idle characters can share knowledge with each other.

### Future: Long-Term Memory

- Selective storage of notable events
- Persists until character death
- Basis for storytelling, artifact creation

## Action System

The action system is how characters do things in the world. It separates *what* is being done from *why*, enabling the same physical actions to serve different purposes with different continuation logic.

### Separation of Concerns

| Concept | Description | Example |
|---------|-------------|---------|
| **ActionType** | Physical action being performed | `ActionPickup`, `ActionCraft` |
| **Context** | Why the action is being performed | Foraging (idle) vs Harvest (order) |
| **Completion Criteria** | When the task is done | Foraging: after one pickup; Harvest: when inventory full |

ActionTypes describe *what* is physically being done. Context is tracked via `char.AssignedOrderID` (0 = no order, non-zero = working on order). After an action completes, the handler checks context to determine next steps.

### Action Categories

Actions fall into three categories with different continuation and interruption semantics. Choosing the right category for a new action is a key architectural decision.

**Need-fulfilling actions** (`ActionConsume`, `ActionDrink`, `ActionSleep`): Driven by stat urgency tiers. Direct actions — eat food, drink water, go to sleep. The intent system prioritizes these when stats are elevated.

**Idle activities** (`ActionForage`, `ActionFillVessel`, `ActionLook`, `ActionTalk`): Chosen autonomously when no needs are pressing (`maxTier == TierNone`). **Idle activities are interruptible** — `CalculateIntent` checks `maxTier < TierModerate` before continuing an idle intent. If needs become Moderate+, the idle activity is dropped. **Idle activities should be non-destructive** — they should never destroy player-relevant state. Need-driven behavior may consume or displace things for survival, but autonomous idle choices should be safe by default (e.g., Fetch Water only triggers when a character has or can find an empty vessel — it never dumps existing contents).

Some idle activities are **self-managing** — they own their complete multi-phase flow (procurement, movement, execution) as a single continuous intent. The character never goes idle between phases and never needs to win another idle roll to continue. See "Self-Managing Actions" below.

**Ordered actions** (`ActionTillSoil`, `ActionPlant`, `ActionWaterGarden`, `ActionPickup` with order context, `ActionCraft`): Player-directed via orders. Handler completes one unit of work, then **clears intent**. Next tick, `CalculateIntent` re-evaluates via a tiered intercept:

- **maxTier == TierNone**: `selectIdleActivity` → `selectOrderActivity` resumes the order (bypasses idle cooldown).
- **maxTier == TierMild && AssignedOrderID != 0**: Inventory-only intercept — check carried inventory for water (if thirsty) or food (if hungry). If found, briefly pause order (`PauseOrder`) and consume. If not found, `selectIdleActivity` → `selectOrderActivity` continues working through Mild.
- **maxTier >= TierModerate**: Priority loop fires as normal; order is paused (`PauseOrder`) and character seeks a fulfillable need.

**The one-tick idle window** between ordered work units is the key mechanism. It gives `CalculateIntent` a natural re-evaluation point, which is how needs interruption and the pause/resume system work. Without it, ordered actions would need their own interruption logic.

**Key distinction:** Self-managing idle actions and ordered actions are both interruptible by needs, but through different mechanisms. Self-managing actions yield when `CalculateIntent` detects elevated tiers mid-flow. Ordered actions yield because intent is cleared between work units, giving `CalculateIntent` a natural re-evaluation point.

| Scenario | Pattern | Why |
|----------|---------|-----|
| Multi-phase idle activity (forage with vessel, fetch water) | Self-managing | Continuous flow; doesn't need pause/resume |
| Player-directed multi-step work (till, plant, water garden) | Ordered (clear intent between units) | Needs interruption + pause/resume via `AssignedOrderID` |
| Simple one-off pickup (order prerequisite, single item) | `ActionPickup` | No continuation needed |

### Self-Managing Actions

Some actions manage their full lifecycle across multiple phases without returning to idle activity selection. This follows **Reuse Before Invention** (Values.md) — rather than inventing a central dispatcher, each action distributes lifecycle knowledge to itself.

**Key principle:** Procurement within a multi-phase action should be a *phase within the action's own handler*, not a separate intent routed through `ActionPickup`. The action type stays the same throughout; the handler detects which phase it's in and acts accordingly.

**Current self-managing actions:**
- `ActionFillVessel` — Phase 1: procure empty vessel (if on ground). Phase 2: move to water and fill.
- `ActionForage` — Phase 1: procure vessel (optional, based on scoring). Phase 2: move to food and pick up.
- `ActionHelpFeed` — Phase 1: procure food (if not already carried). Phase 2: walk to needer. Phase 3: drop food cardinal-adjacent to needer.
- `ActionHelpWater` — Phase 1: procure vessel (if not already carried). Phase 2: fill at water (if vessel empty). Phase 3: walk to needer. Phase 4: drop water vessel cardinal-adjacent to needer. Falls back to helpFeed if no vessel is available but needer also has Crisis hunger.

**Shared tick helpers** (`picking.go`): Self-managing actions reuse shared helpers rather than duplicating logic:

- `RunVesselProcurement` — returns a `ProcurementStatus` enum (`ProcureReady`, `ProcureApproaching`, `ProcureInProgress`, `ProcureFailed`). The action handler calls this each tick during its procurement phase and switches on the result.
- `RunWaterFill` — same shape with `WaterFillStatus` enum (`FillReady`, `FillApproaching`, `FillInProgress`, `FillFailed`). Handles finding water, navigating, and filling. **Key contract:** always fills to full; it's the caller's responsibility to decide *when* filling is needed.

```
Action handler pseudocode:
    if needsVessel && !hasVessel {
        status := RunVesselProcurement(char, items, gameMap, log)
        switch status {
        case ProcureReady:
            // proceed to main phase
        case ProcureApproaching, ProcureInProgress:
            return  // still procuring
        case ProcureFailed:
            // handle failure (nil intent or fall back)
        }
    }
    // Main phase: do the real work
```

**Critical: `char.Intent` is nil after `ProcureReady`.** `RunVesselProcurement` calls `Pickup()`, which nils `char.Intent`. If the handler falls through to a subsequent phase (rather than clearing intent and returning), it must rebuild `char.Intent` before calling `moveWithCollision` or accessing `char.Intent` fields. This applies even when the handler uses a captured `vessel` variable from the top of the case — the variable is still valid, but `char.Intent` is not. Handlers that clear intent and return on `ProcureReady` (e.g., `ActionWaterGarden`) avoid this because re-evaluation rebuilds intent next tick. Handlers that fall through (e.g., `ActionHelpWater`, `ActionFillVessel`) must check world state and rebuild intent for the correct next phase.

**Phase detection:** Handlers detect their current phase by checking world state (e.g., "is my target vessel on the ground or in inventory?") rather than storing explicit phase numbers. This is stateless and survives save/load without additional serialization.

**Why not centralize in ActionPickup?** An earlier design considered routing all post-pickup decisions through the `ActionPickup` handler. This was rejected because it creates a central dispatcher that must understand every context — the opposite of self-managing. It scales poorly as new multi-phase activities are added. Self-managing actions distribute that knowledge to the actions that need it.

### `continueIntent` Rules

`continueIntent` (intent.go) runs every tick for characters with existing intents. It recalculates paths and handles arrival transitions. Understanding its structure is critical when adding new actions.

**Two layers:**

1. **Action-specific early returns** — actions where `TargetItem` can be in different locations across phases (ground vs inventory) handle their own path recalculation and return before reaching the generic logic.

2. **Generic fallthrough** — all other actions use the shared path: verify target item still exists on map, recalculate path toward target, apply arrival transitions.

**The decision rule for new actions:**

- **Needs an early-return block if:** Your action's `TargetItem` can be in the character's inventory (not on the map) during any phase. The generic path's `ItemAt` check would nil the intent when the item moves to inventory.
  - Examples: `ActionConsume` (food in inventory), `ActionDrink` (carried vessel), `ActionFillVessel` (vessel moves ground → inventory), `ActionWaterGarden` (vessel moves ground → inventory), `ActionHelpFeed` (food moves ground → inventory during procurement), `ActionHelpWater` (vessel moves ground → inventory; three navigation targets across phases)

- **Uses the generic path if:** Your action's `TargetItem` stays on the map throughout, targets a character, targets a position, or has no `TargetItem`.
  - Examples: `ActionLook` (item stays on map), `ActionTalk` (targets character via Dest), `ActionTillSoil` (targets position), `ActionPickup` (item on map until picked up), `ActionConsume` targeting a ground food vessel (vessel stays on map; arrival detected in the handler)

**Walk-then-act pattern:** Actions like `ActionLook` and `ActionTalk` set their final action type from the start (not `ActionMove`). The intent creator owns the action type; `continueIntent` just recalculates paths via the generic fallthrough. The handler in `update.go` has a walking phase (moves via `moveWithCollision` while not yet at target) and an acting phase (performs the action when arrived). This follows **Consistency Over Local Cleverness** (Values.md) — all walk-then-act actions share the same shape.

### Adding New Actions

Three checklists organized by category. Each includes every touchpoint; see the sections above for the rationale behind each pattern.

**Adding a Need-Driven Action** (e.g., ActionConsume, ActionDrink, ActionSleep):

1. Action constant in `character.go`
2. Intent finder in `intent.go` (driven by stat urgency tiers)
3. `applyIntent` handler in `update.go` — performs the action, clears intent when stat satisfied or source exhausted
4. `continueIntent`: if `TargetItem` can be in inventory, add an early-return block. If `TargetItem` is always on the map, generic path handles it. (See `continueIntent` Rules above.)
5. `simulation.go`: add handler if simulation tests exercise this action
6. No activity registry entry needed (need-driven actions aren't idle activities)

**Adding an Idle Activity** (e.g., ActionLook, ActionTalk, ActionForage, ActionFillVessel):

1. Action constant in `character.go`
2. Activity entry in `ActivityRegistry` (`entity/activity.go`) — omit if the activity is a pre-roll override (like `ActionHelpFeed`, `ActionHelpWater`) rather than a rolled idle activity
3. Intent finder (location depends on context: `intent.go` for social/observation, `foraging.go` for food-seeking, `picking.go` for resource-seeking, `helping.go` for crisis response)
4. Wire into `selectIdleActivity` in `idle.go` — either as a pre-roll override (checked before the roll) or as a rollable option
5. Add action to `isIdleAction` in `idle.go` if the action should be treated as idle for talking availability (i.e., it is a true idle activity, not a helping/delivery action). `isIdleAction` checks `ActionType` enum — simpler and more robust than string matching.
6. `applyIntent` handler in `update.go`
7. `continueIntent`: self-managing multi-phase actions (where `TargetItem` moves between ground and inventory) need an early-return block. Walk-then-act actions (Look, Talk) use the generic path. (See `continueIntent` Rules above.)
8. `simulation.go`: add handler if simulation exercises this activity

**Adding an Ordered Action** (e.g., ActionTillSoil, ActionPlant, ActionWaterGarden, ActionCraft):

1. Action constant in `character.go`
2. Activity entry in `ActivityRegistry` (with `IntentOrderable` and appropriate `Category`)
3. `findXxxIntent()` in `order_execution.go` — handles target selection on each resumption tick
4. Wire into `findOrderIntent` switch, `isMultiStepOrderComplete`, `IsOrderFeasible`
5. `applyIntent` handler in `update.go` — complete one work unit, clear intent, check order completion inline
6. `continueIntent`: multi-phase ordered actions with vessel procurement (e.g., WaterGarden) need an early-return block. Single-phase ordered actions (TillSoil, Plant) use the generic path. (See `continueIntent` Rules above.)
7. `simulation.go`: add handler if simulation exercises this action
8. If the action uses vessel procurement: use `RunVesselProcurement` tick helper
9. If the action uses water fill: use `RunWaterFill` tick helper

## Activity Registry & Know-How Discovery

Activities are the named behaviors that characters learn and perform. The registry controls availability, discovery, and order UI grouping.

### Activity Properties

| Field | Purpose |
|-------|---------|
| `ID` | Unique identifier (e.g., "harvest", "craftVessel") |
| `Category` | Groups orderable activities in the order UI ("craft", "garden", or empty for top-level) |
| `IntentFormation` | `IntentAutomatic` (needs/idle) or `IntentOrderable` (player orders) |
| `Availability` | `AvailabilityDefault` (all characters) or `AvailabilityKnowHow` (must discover) |
| `DiscoveryTriggers` | What experience unlocks this activity |

**Category field**: Groups orderable activities for the order UI menu hierarchy. `getOrderableActivities()` generates synthetic category entries (e.g., "Craft", "Garden") when any known activity has that category. Uncategorized activities (e.g., Harvest) appear at the top level.

### Discovery Triggers

Know-how activities are discovered through experience. Each trigger specifies an action, optional item type, and optional requirements (`RequiresEdible`, `RequiresPlantable`).

Example: Harvest is discovered by picking up, eating, or looking at edible items. Plant is discovered by picking up or looking at plantable items.

### Adding a New Activity

This checklist covers the activity registry specifically. For the full action system touchpoints (handler, intent finder, `continueIntent`), see "Adding New Actions" above.

1. Add entry to `ActivityRegistry` in `entity/activity.go`
2. Set `Category` for grouped activities (generates synthetic category entries in order UI)
3. Add discovery triggers if `AvailabilityKnowHow`
4. If orderable: add case to `findOrderIntent()` switch in `order_execution.go`

## Orders

Orders are player-directed tasks. They share physical actions with idle activities but have different triggering, completion, and interruption semantics.

### Order Execution

- `idle.go` calls `selectOrderActivity()` first, giving orders priority over idle activities
- Order eligibility checks activity's `Availability` against character's known activities
- `IsOrderFeasible(order, items, gameMap)` is computed on demand at assignment and render time — returns `(feasible bool, noKnowHow bool)`. Unfeasible orders are skipped during `findAvailableOrder` and rendered dimmed with `[Unfulfillable]` or `[No one knows how]`.
- `LockedVariety string` on Order: set when the first item is planted. After locking, the character only seeks items of that variety, keeping a single order focused.

### Unified Order Completion

All order types call `CompleteOrder()`, which sets `OrderCompleted` status. A sweep in the game loop after intent application removes all `OrderCompleted` orders that tick. Action handlers contain inline completion checks; `selectOrderActivity` and `findAvailableOrder` skip `OrderCompleted` orders as a safety net.

### Marked-for-Tilling Pool

Tilling separates the player's plan from worker assignments:
- **Marked tiles** (`gameMap.markedForTilling`): User's tilling plan, persistent, independent of orders
- **Till Soil orders**: Worker assignments. Multiple orders = multiple workers on the shared pool.
- Cancelling an order removes the worker, not the plan. Unmarking tiles removes from plan (via area selection in unmark mode).

Pool is serialized in `SaveState.MarkedForTillingPositions`.

## Item Acquisition

`picking.go` is the shared home for all item acquisition logic, organized in three layers:

1. **Map Search** — `findNearestItemByType`, `FindAvailableVessel` (vessels that can receive items), `FindVesselContaining` (vessels whose contents match a target)
2. **Prerequisite Orchestration** — `EnsureHasVesselFor`, `EnsureHasRecipeInputs`, `EnsureHasItem` (check-or-go-get helpers)
3. **Physical Actions** — `Pickup`, `Drop`, vessel operations

### Pickup Result Pattern

`Pickup()` returns a `PickupResult` to distinguish outcomes:
- `PickupToInventory` — item picked up directly (inventory was empty)
- `PickupToVessel` — item added to carried vessel's stack
- `PickupFailed` — could not pick up (variety mismatch with vessel)

Callers handle continuation differently based on result and context (foraging vs harvesting).

### Component Procurement Flow

Many activities require gathering specific items before performing the action (harvest, craft, gardening).

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

### Variety Restoration on Extraction

When items are reconstructed from vessel stacks, `ConsumeAccessibleItem` and `ConsumePlantable` restore constructor-set fields (`Sym`, `Plantable`, `Kind`, `Edible`) from the variety. This is necessary because vessel stacks store variety references, not full item structs — direct struct reconstruction skips constructor logic.

**Rule:** Any new field set by a constructor must be added to the variety and restored on extraction. This is a serialization concern — see the Serialization Checklist below.

**Pattern:** Check availability at intent creation (no extraction), consume at action completion. This supports pause/resume — the item stays accessible until actually consumed.

### Consumption Side Effects

`ConsumeFromInventory` and `ConsumeFromVessel` accept a `gameMap` parameter for side effects at the consumption site. Currently: consuming a gourd drops a seed of the same variety. This pattern generalizes to any future consumption side effect (leaving a core, shell, or husk).

### Look-for-Container Pattern

When foraging or harvesting without a vessel:
1. `FindAvailableVessel()` searches for empty or compatible vessel on ground
2. If found, intent targets vessel first
3. After vessel pickup, continues to harvest/forage into vessel
4. If no vessel, picks up item directly to inventory

### Future: Unified Item-Seeking

Currently, `foraging.go` has preference-weighted scoring while `picking.go`'s prerequisite helpers use nearest-distance. These should converge — character item-seeking should be consistent regardless of context. Preference shapes material culture (e.g., a character who prefers silver shells will craft silver shell hoes). See `triggered-enhancements.md`.

## Recipe System

Recipes define how to craft items from components.

### Crafting Flow

1. Order assigned → `findCraftIntent` gets recipes for activity, filters to feasible recipes
2. Picks first feasible recipe (future: preference-weighted selection)
3. `EnsureHasRecipeInputs` gathers missing components
4. Once all components accessible: perform crafting action
5. On completion: consume all inputs, create output item via recipe-specific creation function

### Adding a New Recipe

1. Add recipe to `RecipeRegistry` in `entity/recipe.go`
2. Add activity to `ActivityRegistry` for the crafting activity (if new)
3. Add creation function in `system/crafting.go` (e.g., `CreateHoe`)
4. Add case to `applyIntent()` ActionCraft handler dispatch (by recipe ID)

## World & Terrain

### Water Terrain

Water tiles (springs, ponds) are stored as map terrain (`water map[Position]WaterType`), not as features. This enables O(1) lookups and clean separation from the feature system.

| Water Type | Symbol | Rendering |
|------------|--------|-----------|
| WaterSpring | `☉` | Single character |
| WaterPond | `▓` | Three-character fill `▓▓▓` |

Water tiles are impassable. Characters interact from cardinal-adjacent tiles. Tiles 8-directionally adjacent to any water are "wet" — computed on the fly via `IsWet(pos)`, no persistent state.

Manually-watered tiles track wet status with a decay timer (see `config.WateredTileDecayTime`). `IsWet(pos)` checks both water adjacency and the watered-tile timer.

### Drinking Sources

Characters can drink from three source types, unified by distance scoring in `findDrinkIntent`:
- **Terrain water** — infinite, drinks continuously until sated
- **Carried vessel** — finite, clears intent after each drink (re-evaluate nearest source)
- **Ground vessel** — finite, character moves to vessel and drinks in place

Intent clearing for vessel drinking ensures characters naturally handle vessel depletion and re-prioritize the nearest source.

### Food Sources

Characters score food across four candidate pools in `FindFoodTarget`:
- **Carried loose food** — distance 0
- **Carried food vessel contents** — distance 0, scored by contents' variety
- **Ground loose food** — Manhattan distance from character
- **Ground food vessel contents** — Manhattan distance to vessel, scored by contents' variety; vessel is TargetItem

Ground vessel eating follows the same eat-in-place pattern as ground vessel drinking: character walks to the vessel, consumes one unit, clears intent, and re-evaluates. The vessel stays on the map throughout — no early-return block needed in `continueIntent` (vessel never enters inventory). `ConsumeFromVessel` operates on the vessel pointer directly, whether it is in inventory or on the ground.

### Tilled Soil

Tilled soil is map terrain state (`tilled map[Position]bool`), following the same pattern as water. Key difference: tilled tiles are walkable and items can exist on them.

Rendering: `═══` fill for empty tilled tiles, `═X═` fill around entities on tilled soil. Wet tilled soil uses distinct styles from dry.

### Pond Generation

`SpawnPonds()` generates 1-5 ponds of 4-16 contiguous water tiles each via blob growth. After placing all ponds, `isMapConnected()` verifies walkability via BFS. If partitioned, regenerates (max 10 retries). Ponds generate before features and items.

### Features

Features are map elements that aren't items or characters. Currently only leaf piles (passable, used as beds). Springs migrated to water terrain.

Features have a `Passable` boolean — impassable features are handled by `IsBlocked()` and pathfinding automatically.

### Movement & Pathfinding

**`NextStepBFS`**: Greedy-first pathfinding. Tries a greedy diagonal step (moving along the larger of X or Y delta) before running BFS. If the greedy step is clear, takes it — this produces natural zigzag paths and spreads characters heading to the same destination across different routes. BFS only runs when the greedy step hits water or an impassable feature. Falls back to greedy `NextStep` if no BFS path exists. Used by all callers with gameMap access. The public function is a thin wrapper over `nextStepBFSCore(preferBFS bool)`.

**Sticky BFS (`UsingBFS` flag)**: Once a character uses BFS to navigate around an obstacle, they stay in BFS mode for the remainder of that intent. `UsingBFS bool` is an ephemeral field on Character (not serialized, like displacement fields). `continueIntent` passes `char.UsingBFS` as `preferBFS` to `nextStepBFSCore` and sets `char.UsingBFS = true` when BFS was actually used. The flag clears in two places: (a) when `Intent` is nilled in `CalculateIntent` (covers reaching target and intent changes), and (b) when `initiateDisplacement` fires (covers character collision). This follows the displacement precedent — ephemeral, not serialized, clears on save/load.

**`NextStep`**: Greedy single-step toward target along larger axis delta. No obstacle awareness. Fallback only.

**`findAlternateStep`**: Per-tick reactive routing around blocked tiles. Used by `MoveCharacter` when next step is occupied.

**Perpendicular displacement on character collision**: When `MoveCharacter` fails due to another character (not terrain), the blocked character enters displacement mode: 3 perpendicular sidesteps before resuming BFS. Direction is chosen randomly from the two perpendiculars; if blocked, tries the opposite. If both blocked, displacement is skipped. Displacement state (`DisplacementStepsLeft`, `DisplacementDX`, `DisplacementDY`) is ephemeral — not serialized. On save/load, displacement clears and the character re-pathfinds normally. This extends `findAlternateStep`'s reactive-routing pattern to multi-step intentional routing without modifying BFS semantics or treating characters as obstacles in pathfinding.

**Movement blocking**: `IsBlocked(pos)` returns true if character, water tile, or impassable feature occupies the position.

## Area Selection UI Pattern

Area selection enables players to define rectangular regions for terrain modification (tilling, future construction).

**Flow:**
1. Player selects activity → enters area selection mode
2. Move cursor, press `p` to anchor first corner
3. Move cursor to resize rectangle (valid tiles highlighted)
4. Press `p` to confirm (marks tiles / creates order)
5. Press `Tab` to toggle mark/unmark mode
6. Press `Enter` when done, `Esc` to cancel

**Rendering**: Rectangle highlight uses full background for empty tiles, padding-only for entities (avoids ANSI nesting). Three distinct color states map directly to workflow progression: active selection (teal) → confirmed pending (sage) → completed work (dusky earth). Each state should be visually distinct from the others — reusing a color across states collapses workflow stages the player needs to distinguish.

### Reuse for Future Activities

When adding new area-based orders (e.g., fence placement, building zones):
1. Add activity check in step 1 Enter handler
2. Write activity-specific validator function (like `isValidTillTarget`)
3. Reuse `getValidPositions` with custom validator
4. Handle plot confirm logic in `p` key handler

## Save/Load Serialization

Save files stored in `~/.petri/worlds/world-XXXX/` with `state.json`, `state.backup`, and `meta.json`.

### Serialization Checklist

When adding fields to saved structs:

1. **Display fields**: Symbols (`Sym`), colors, styles set by constructors — must be explicitly restored on deserialization
2. **All attribute fields**: Easy to miss nested fields (e.g., Pattern/Texture on preferences)
3. **Round-trip tests**: Save → load → verify all fields match
4. **Variety fields**: `VarietySave` must include all fields that `ConsumeAccessibleItem` / `ConsumePlantable` need to restore (currently: `Kind`, `Plantable`, `Sym`)

Constructor-set fields won't be populated when deserializing directly into structs — must be explicitly restored based on type.

**PlantProperties serialization**: `IsSprout` and `SproutTimer` must round-trip correctly to preserve sprout state across save/load.

**Save compatibility when changing entity storage**: When changing how entities are stored (e.g., moving data between fields, maps, or types), verify save/load round-trip in the same step. Check: (1) new state serializes, (2) old saves migrate, (3) serialize tests updated.

## Position Handling

All coordinates use `types.Position` struct with `X, Y int`.

**Do:**
- Use `pos.DistanceTo(other)` for distance calculations
- Use `pos.IsAdjacentTo(other)` for 8-direction adjacency checks
- Use `pos.IsCardinallyAdjacentTo(other)` for N/E/S/W only
- Create Position with `types.Position{X: x, Y: y}`

**Don't:**
- Inline distance calculations like `abs(x1-x2) + abs(y1-y2)`
- Create new position-like structs
- Define local `abs()` or `sign()` functions — use `types.Abs` and `types.Sign`

## Common Implementation Pitfalls

**Game time vs wall clock**: UI indicators that should work when paused (like "Saved" message) need wall clock time (`time.Now()`), not game time which only advances when unpaused.

**Sorting stability**: When displaying merged data from maps (e.g., AllEvents from ActionLog), use `sort.SliceStable` with deterministic tiebreakers (like CharID) to prevent visual jitter from Go's random map iteration order.

**View transitions**: When switching between views with different rendering approaches (game view uses direct rendering, menus use lipgloss.Place for centering), add dimension safeguards for edge cases.

**Terrain fill in `renderCell()`**: Terrain that renders as solid blocks (tilled soil `═══`, water `▓▓▓`) requires both `sym` AND `fill` set to the styled terrain character. Setting only `sym` produces a single character flanked by spaces (` ▓ `), creating a vertical stripe appearance.
