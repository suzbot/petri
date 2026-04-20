# Petri Architecture

## Table of Contents

- [Design Model](#design-model)
  - [The Character's Cognitive Loop](#the-characters-cognitive-loop)
  - [External Direction](#external-direction)
  - [Social Awareness](#social-awareness)
  - [Separation of Concerns by Cognitive Role](#separation-of-concerns-by-cognitive-role)
  - [Encapsulated Workflows](#encapsulated-workflows)
  - [Intent as the Universal Contract](#intent-as-the-universal-contract)
- [Key Design Patterns](#key-design-patterns)
- [Data Flow](#data-flow)
- [World & Terrain](#world--terrain)
  - [Water Terrain](#water-terrain)
  - [Drinking Sources](#drinking-sources)
  - [Food Sources](#food-sources)
  - [Tilled Soil](#tilled-soil)
  - [Pond Generation](#pond-generation)
  - [Features](#features)
  - [Constructs](#constructs)
  - [Movement & Pathfinding](#movement--pathfinding)
- [Position Handling](#position-handling)
- [Item Model](#item-model)
  - [Attribute Classification](#attribute-classification)
  - [Kind Pattern for Subtypes](#kind-pattern-for-subtypes)
  - [Crafted Items](#crafted-items)
  - [Adding New Plant Types](#adding-new-plant-types)
  - [Adding New Item Types (Non-Plant)](#adding-new-item-types-non-plant)
  - [Adding New Terrain Types](#adding-new-terrain-types)
  - [Adding a New Construct Type](#adding-a-new-construct-type)
- [Item Lifecycle](#item-lifecycle)
  - [Plant-Based Spawning](#plant-based-spawning)
  - [Ground Spawning](#ground-spawning)
  - [Death Timers](#death-timers)
- [Memory & Knowledge Model](#memory--knowledge-model)
  - [ActionLog (Working Memory)](#actionlog-working-memory)
  - [Knowledge System](#knowledge-system)
  - [Future: Long-Term Memory](#future-long-term-memory)
- [Action System](#action-system)
  - [Separation of Concerns](#separation-of-concerns)
  - [Action Categories](#action-categories)
  - [Self-Managing Actions](#self-managing-actions)
  - [`continueIntent` Rules](#continueintent-rules)
  - [Adding New Actions](#adding-new-actions)
- [Activity Registry & Know-How Discovery](#activity-registry--know-how-discovery)
  - [Activity Properties](#activity-properties)
  - [Discovery Triggers](#discovery-triggers)
  - [Adding a New Activity](#adding-a-new-activity)
- [Orders](#orders)
  - [Order Execution](#order-execution)
  - [Unified Order Completion](#unified-order-completion)
  - [Marked-for-Tilling Pool](#marked-for-tilling-pool)
  - [Marked-for-Construction Pool](#marked-for-construction-pool)
- [Item Acquisition](#item-acquisition)
  - [Pickup Result Pattern](#pickup-result-pattern)
  - [Component Procurement Flow](#component-procurement-flow)
  - [Variety Restoration on Extraction](#variety-restoration-on-extraction)
  - [Consumption Side Effects](#consumption-side-effects)
  - [Look-for-Container Pattern](#look-for-container-pattern)
  - [Future: Unified Item-Seeking](#future-unified-item-seeking)
- [Recipe System](#recipe-system)
  - [Crafting Flow](#crafting-flow)
  - [Repeatable Recipes](#repeatable-recipes)
  - [Adding a New Recipe](#adding-a-new-recipe)
- [Area Selection UI Pattern](#area-selection-ui-pattern)
  - [Reuse for Future Activities](#reuse-for-future-activities)
- [Save/Load Serialization](#saveload-serialization)
  - [Serialization Checklist](#serialization-checklist)
- [Common Implementation Pitfalls](#common-implementation-pitfalls)

## Design Model

The architecture mirrors what the game simulates (see Values.md: Isomorphism). These mental models explain *why* the code is structured the way it is.

### The Character's Cognitive Loop

Each tick, a character does something like:

1. **What do I need?** — My body tells me I'm thirsty, hungry, tired. How urgent is it?
2. **What are my options?** — I can see water over there, there's food nearby, I could ask for help.
3. **Which option do I pick?** — I'm very thirsty and the water is close. I'll drink.
4. **How do I get there?** — There's a rock in the way, I'll path around it.
5. **Do the thing.** — Walk, pick up, eat, drink, craft, talk.

Between ticks, the world acts on the character: hunger grows, energy drains, poison damages, sleep restores.

### External Direction

Orders from the player don't replace the cognitive loop — they insert into it. A character with an order still gets hungry, still evaluates urgency, and may pause the order to eat. The order adds an option (step 2) and gets prioritized when needs aren't pressing (step 3).

### Social Awareness

Characters notice when others are in crisis and may choose to help. This competes with discretionary activities — it's part of "what are my options?" when a character has no pressing needs.

### Separation of Concerns by Cognitive Role

The cognitive loop maps to code boundaries:

| Cognitive Step | Responsibility | Location |
|---|---|---|
| **Motivation** — what do I need? | Urgency tiers, stat priority | intent.go (CalculateIntent) |
| **Evaluation** — what options exist? | Food scoring, drink finding, sleep finding | intent.go (findFoodIntent, etc.) |
| **Selection** — which option wins? | Priority ordering, bucket routing | intent.go + discretionary.go + order_execution.go |
| **Spatial planning** — how do I get there? | Pathfinding, obstacle avoidance | movement.go (NextStepBFS) |
| **Execution** — do the thing | Consume, pickup, craft, talk, till, plant | apply_actions.go (applyIntent) |
| **Body** — passive changes | Stat decay, sleep/wake, damage | survival.go |

### Encapsulated Workflows

Some actions involve multi-phase workflows (fetch water: get vessel → fill → done). Once committed to a course of action, the sub-decisions (which vessel, which water source) are implementation details of that commitment, not top-level cognitive choices. The architecture preserves this encapsulation — see "Self-Managing Actions" below.

### Intent as the Universal Contract

The Intent struct is the handoff between deciding and doing. It remains the single interface between the decision phase and execution phase.

## Key Design Patterns

- **MVU Architecture**: Bubble Tea handles rendering diffs automatically
- **Intent System**: Characters calculate Intent, then intents applied atomically (enables future parallelization)
- **Four-Bucket Decision Model**: `CalculateIntent` routes through explicit priority buckets — Needs → Orders → Helping → Discretionary. Each bucket has its own file: `intent.go` (needs), `order_execution.go` (orders), `helping.go` (helping), `discretionary.go` (discretionary). The orchestrator in `CalculateIntent` calls each bucket explicitly.
- **"Stuck" vs "Idle" terminal states**: When `CalculateIntent` exhausts all buckets, the outcome depends on max tier: `maxTier >= TierModerate` → "Stuck (can't meet needs)" (has Moderate+ needs but nothing available); otherwise → "Idle" (content or only Mild needs, nothing to do). These are fundamentally different states. Note: a character with Moderate+ needs that can't be fulfilled sets `CurrentActivity` to a need-specific guard string (e.g., "No bed available") in the need evaluator itself; the terminal "Stuck" label is a safety-net fallback that appears only when no need evaluator set an activity.
- **Tier-worsening re-evaluation**: If a character's driving stat crosses a tier boundary (e.g., Energy from Moderate to Severe), the existing intent is cleared and `CalculateIntent` runs fresh. This enables behaviors that are only available at higher tiers (e.g., ground sleep becomes available at Exhausted/Severe, not at Moderate) to unlock without waiting for intent completion.
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

See [docs/flow-diagrams.md](flow-diagrams.md) for visual call graphs, the intent priority hierarchy, and multi-phase action state machines.

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

Features are natural map elements that aren't items or characters. Currently only leaf piles (passable, used as beds). Springs migrated to water terrain.

Features have a `Passable` boolean — impassable features are handled by `IsBlocked()` and pathfinding automatically.

### Constructs

Constructs are player-built structures. They are a distinct entity type from Features (natural world elements) — see DD-4 in `construction-design.md`. Constructs have a `ConstructType` (e.g., "structure") and `Kind` (e.g., "fence") hierarchy mirroring ItemType/Kind. The `Passable` boolean on a Construct determines whether characters can move through it; impassable constructs are respected by `IsBlocked()`, `MoveCharacter`, and all three call sites of `nextStepBFSCore`. The `Movable` boolean reserves future furniture-relocation behavior.

Construct storage on `Map`: the `constructs []Construct` slice with `AddConstruct`, `AddConstructDirect`, `ConstructAt`, `Constructs`, `RemoveConstruct`, and ID-generation methods.

### Movement & Pathfinding

**`NextStepBFS`**: Greedy-first pathfinding. Tries a greedy diagonal step (moving along the larger of X or Y delta) before running BFS. If the greedy step is clear, takes it — this produces natural zigzag paths and spreads characters heading to the same destination across different routes. BFS only runs when the greedy step hits water or an impassable feature. Falls back to greedy `NextStep` if no BFS path exists. Used by all callers with gameMap access. The public function is a thin wrapper over `nextStepBFSCore(preferBFS bool)`.

**Sticky BFS (`UsingBFS` flag)**: Once a character uses BFS to navigate around an obstacle, they stay in BFS mode for the remainder of that intent. `UsingBFS bool` is an ephemeral field on Character (not serialized, like displacement fields). `continueIntent` passes `char.UsingBFS` as `preferBFS` to `nextStepBFSCore` and sets `char.UsingBFS = true` when BFS was actually used. The flag clears in two places: (a) when `Intent` is nilled in `CalculateIntent` (covers reaching target and intent changes), and (b) when `initiateDisplacement` fires (covers character collision). This follows the displacement precedent — ephemeral, not serialized, clears on save/load.

**`NextStep`**: Greedy single-step toward target along larger axis delta. No obstacle awareness. Fallback only.

**`findAlternateStep`**: Per-tick reactive routing around blocked tiles. Used by `MoveCharacter` when next step is occupied.

**Perpendicular displacement on character collision**: When `MoveCharacter` fails due to another character (not terrain), the blocked character enters displacement mode: 3 perpendicular sidesteps before resuming BFS. Direction is chosen randomly from the two perpendiculars; if blocked, tries the opposite. If both blocked, displacement is skipped. Displacement state (`DisplacementStepsLeft`, `DisplacementDX`, `DisplacementDY`) is ephemeral — not serialized. On save/load, displacement clears and the character re-pathfinds normally. This extends `findAlternateStep`'s reactive-routing pattern to multi-step intentional routing without modifying BFS semantics or treating characters as obstacles in pathfinding.

**Movement blocking**: `IsBlocked(pos)` returns true if character, water tile, impassable feature, or impassable construct occupies the position.

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
- `SeedTimer float64` — cooldown before this plant can be extracted from again. Starts at 0 (seeds available). Reset to the plant type's spawn interval after extraction. Decremented every tick independently of `SpawnTimer`. Only extractable plant types (see `config.ExtractableTypes`) use this field.

`SourceVarietyID string` — registry ID of the parent plant's variety. Set at seed creation (extraction, gourd consumption). Used at planting time to look up the parent variety and create the sprout with full fidelity — no string derivation. Carried by `Item` (loose seeds) and `ItemVariety` (vessel-stacked seeds).

**ContainerData** enables storage — vessels have `Capacity: 1` (one stack). Contents tracked as `[]Stack` where each Stack has a Variety pointer and count. Liquids are stored as stacks with ItemType "liquid" and Kind (e.g., "water"), reusing all existing vessel infrastructure.

**EdibleProperties** marks items as edible with optional effects. Items with `Edible != nil` can be eaten; Poisonous/Healing determine effects.

**BundleCount** (`int` field on Item, default 0 for non-bundleable items, 1 for bundleable items): tracks how many units are in this bundle slot. Constructors for bundleable types (e.g., `NewStick()`, `NewGrass()`) set `BundleCount: 1`. `Pickup()` increments BundleCount when merging. `Description()` uses BundleCount to produce "bundle of sticks (N)" display text; when Kind is set, uses `Pluralize(Kind)` for the bundle name (e.g., "bundle of tall grass (N)"). The bundle format shows for `BundleCount >= 2` (single items use their normal description); growing plants also keep their normal description. Items with `BundleCount >= 2` render as `X` on the map (`config.CharBundle`). `config.IsBundleable(itemType)` and `config.MaxBundleSize` drive bundle eligibility and capacity.

**VesselExcludedTypes** (`config.VesselExcludedTypes map[string]bool`): explicit set of item types that cannot be placed in vessels (sticks, grass, clay, brick). This is distinct from bundle logic — non-bundleable items (clay, brick) can still be vessel-excluded. Checked in `AddToVessel`, `CanVesselAccept`, and gather/harvest order vessel-procurement branching. **Decision rule**: if a new item type should never go in a vessel, add it here. Do not rely on variety absence as implicit exclusion.

### Kind Pattern for Subtypes

Some item types use a `Kind` subtype field for hierarchical identity:
- **ItemType**: broad product category (e.g., `"hoe"`, `"seed"`) — what the player orders or conceptually groups
- **Kind**: recipe- or origin-specific subtype (e.g., `"shell hoe"`, `"gourd seed"`) — what the character produces or what was consumed to create it
- `Description()` uses Kind when set, falls back to ItemType
- Natural items leave Kind empty

This follows **Design Types for Future Siblings** (Values.md) — `ItemType: "liquid", Kind: "water"` accommodates future beverages (mead, beer, wine) as Kind variants with no structural changes needed.

`ItemVariety` also carries a `Kind` field (mirrors `Item.Kind`) so vessel contents can restore the correct Kind when items are extracted. See "Variety Restoration on Extraction" below.

**Kind and plant reproduction:** `spawnItem()` copies `Kind` from the parent alongside `Color`, `Pattern`, and `Texture`. Grass is the first plant type with a Kind value that reproduces via `spawnItem` — without this copy, reproduced plants would lose their Kind. When adding a new plant type with Kind, verify that `spawnItem` copies it (it does, generically) and that the kind survives the sprout→mature transition in `UpdateSproutTimers`.

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
5. `system/lifecycle.go` — Add case to `MatureSymbol()` for the mature plant symbol. Verify that any constructor-set fields are restored on maturation in `UpdateSproutTimers`. `spawnItem()` copies descriptive attributes (Color, Pattern, Texture, Kind) — functional fields set by the constructor (e.g., `BundleCount`, `Edible`) must be restored when the sprout matures, since they won't survive the sprout→mature transition otherwise.
6. `game/world.go` — Add case to `createItemFromVariety()` to call the constructor. Without this, the type falls through to the default (currently `NewFlower`).
7. `ui/serialize.go` — Add case to the symbol restoration switch in `itemFromSave()`. Symbols are not serialized — they're restored from ItemType on load.

### Adding New Item Types (Non-Plant)

When adding new item types (tools, materials, resources):

1. `entity/item.go` — Add `NewX()` constructor. Set `Color` (required for rendering via `styledSymbol`). Confirm with user how the item should appear in the details panel — set `Name` if the auto-generated `Description()` (color + itemType) isn't right.
2. `config/config.go` — Add character constant (`CharX`). If edible, add to `ItemMealSize`.
3. `types/types.go` — Add new `Color` constant if needed (e.g., `ColorEarthy`).
4. `ui/styles.go` — Add rendering style for the new color if it doesn't have one.
5. `ui/view.go` — Add case in `styledSymbol()` for the new color. If the item has associated terrain, add terrain rendering in `renderCell()` and terrain annotation in details panel (both empty-tile and entity-on-tile paths).
6. `ui/serialize.go` — Add case to the symbol restoration switch in `itemFromSave()`.
7. Add new functional flags as needed (Craftable, Wearable, Drinkable). Descriptive attributes remain the basis for preference formation.

### Adding New Terrain Types

1. `game/map.go` — Add field (e.g., `clay map[types.Position]bool`), initialize in `NewMap`, add Set/Is/Has/Positions query methods.
2. `game/world.go` — Add spawn function, wire into world gen in `ui/update.go` (both `startGameRandom` and `startGameFromCreation`).
3. `ui/styles.go` — Add terrain style.
4. `ui/view.go` — Add terrain rendering in `renderCell()` (check rendering order: water → clay → tilled). Add terrain fill behind entities. Add terrain annotation in details panel (both empty-tile "Type:" section and entity-on-terrain annotation).
5. `save/state.go` — Add positions field to `SaveState`. Add serialization in `ui/serialize.go` (both `ToSaveState` and `FromSaveState`).

### Adding a New Construct Type

Constructs are player-built structures stored as a slice on the `Map`. Each has a `ConstructType` (broad category, e.g., "structure") and `Kind` (specific type, e.g., "fence").

1. `entity/construct.go` — Add `NewX()` constructor. Set `ConstructType`, `Kind`, `Material`, `MaterialColor`, `Passable`, and `Movable`. The constructor is the source of truth for these values. For position-aware constructs (e.g., hut walls), add a `WallRole string` field to encode structural position — use coarse semantic values only (`"wall"`, `"door"`); do not encode fine-grained visual roles (corner, edge, junction) here. Passability is set from `WallRole` in the constructor.
2. `config/config.go` — Add character symbol constants for each distinct glyph (e.g., `CharFence = '╬'`, or the `CharHut*` constants for heavy box-drawing).
3. `types/types.go` — Add new `Color` constant if needed.
4. `ui/styles.go` — Add rendering style for any new color.
5. `ui/view.go` — Add construct rendering in `renderCell()`. Constructs use `colorToStyle(color)` (the shared color-to-style helper) for consistent color resolution. Add details panel display (DisplayName, type label, "Not passable" when `!Passable`). Constructs appear in both the empty-tile and entity-on-tile rendering paths. For constructs with position-dependent symbols (e.g., hut walls), compute the box-drawing character and horizontal fill at render time via adjacency lookup — call a helper like `hutSymbolFromAdjacency(pos, world)` that queries cardinal neighbor constructs of the same kind; do not store the symbol on the construct. For asymmetric horizontal fill (e.g., hut corners/doors), use `leftFill`/`rightFill` variables returned by the adjacency helper.
6. `system/movement.go` — If the new type can be impassable: verify `IsBlocked`, `MoveCharacter`, and all three `nextStepBFSCore` call sites already handle constructs via `ConstructAt`. No per-type changes needed if `Passable` is false — the existing checks suffice.
7. `save/state.go` — Add fields to `ConstructSave` struct if the new type has additional properties not already covered (e.g., `WallRole string` with `json:"wall_role,omitempty"`).
8. `ui/serialize.go` — Add constructor call in `FromSaveState` construct restoration. For old saves with fine-grained WallRole values (corner-tl, edge-h, etc.), map them to the coarse semantic equivalents ("wall"/"door") in `constructFromSave` for backward compatibility. The symbol is computed at render time, so no symbol restoration is needed.

**`colorToStyle` helper**: `renderCell()` uses a shared `colorToStyle(color types.Color) lipgloss.Style` helper to resolve Color constants to lipgloss styles. Both item rendering and construct rendering call this helper — when adding a new color for any entity type, add a case to `colorToStyle` rather than duplicating style logic per entity.

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

- **maxTier == TierNone**: bucket routing → `selectOrderActivity` resumes the order (bypasses idle cooldown).
- **maxTier == TierMild && AssignedOrderID != 0**: Inventory-only intercept — check carried inventory for water (if thirsty) or food (if hungry). If found, briefly pause order (`PauseOrder`) and consume. If not found, bucket routing → `selectOrderActivity` continues working through Mild.
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

2. **Generic fallthrough** — all other actions use the shared path: verify target still exists, recalculate path toward target, apply arrival transitions. Adding a new target field to Intent requires updating **four touchpoints** across `CalculateIntent` and `continueIntent`:
   1. **`CalculateIntent` continuation gate** — the `if/else if` chain that decides whether to call `continueIntent` for non-need intents (checks `TargetItem != nil`, `TargetConstruct != nil`, `TargetBuildPos != nil`, `TargetCharacter != nil`). A missing entry means the intent falls through to full re-evaluation every tick — the character picks a new intent each tick and never completes the action.
   2. **Target recalculation chain** (in `continueIntent`) — the `if/else if` block that maps target fields to navigation coordinates (tx, ty). A missing entry causes the fallthrough to use `intent.Target` (a single BFS step) as the destination, which produces a no-op path.
   3. **Arrival detection** (in `continueIntent`) — action-specific blocks that detect adjacency/co-location and transition the intent to the "stay and act" phase (e.g., item-look adjacency, construct-look adjacency, talk adjacency, bed co-location). A missing block means the character walks to the target but never transitions.
   4. **Final return intent** (in `continueIntent`) — the struct literal at the bottom that preserves all target fields for the next tick. A missing field means the target is silently dropped after one tick.

**The decision rule for new actions:**

- **Needs an early-return block if:** Your action's `TargetItem` can be in the character's inventory (not on the map) during any phase. The generic path's `ItemAt` check would nil the intent when the item moves to inventory.
  - Examples: `ActionConsume` (food in inventory), `ActionDrink` (carried vessel), `ActionFillVessel` (vessel moves ground → inventory), `ActionWaterGarden` (vessel moves ground → inventory), `ActionHelpFeed` (food moves ground → inventory during procurement), `ActionHelpWater` (vessel moves ground → inventory; three navigation targets across phases)

- **Uses the generic path if:** Your action's `TargetItem` stays on the map throughout, targets a character, targets a position, or has no `TargetItem`.
  - Examples: `ActionLook` (item stays on map, or construct via `TargetConstruct`), `ActionTalk` (targets character via Dest), `ActionTillSoil` (targets position), `ActionPickup` (item on map until picked up), `ActionConsume` targeting a ground food vessel (vessel stays on map; arrival detected in the handler)

- **Adjacent-tile position-based ordered actions** (e.g., `ActionBuildFence`): A variant where the character stands on a cardinally adjacent tile rather than the work tile. These use `TargetBuildPos` (the work tile) and `Dest` (the standing tile). The generic `continueIntent` fallthrough navigates toward `Dest` when `TargetBuildPos` is set, and preserves `TargetBuildPos` in the returned intent. **Critical:** `CalculateIntent`'s non-need continuation gate must include `TargetBuildPos != nil` as a condition to call `continueIntent` — without it, the intent falls through to full re-evaluation every tick, causing the character to thrash by selecting a new target tile each tick instead of walking toward the current one.

**Same-tile rerouting:** `ActionLook` and `ActionTalk` require adjacency, not co-location. If a character is on the same tile as the target (item or character), `continueIntent` reroutes to an adjacent tile via `findClosestAdjacentTile` rather than falling through to the generic path. Without this, the generic path's `isAdjacent` check (which excludes distance zero) would never trigger, and `NextStepBFS` from the current tile to the current tile would return a no-op — creating an infinite loop. This is not an early-return block (the target doesn't move between phases); it's an adjacency-contract edge case specific to actions that observe rather than interact.

**Walk-then-act pattern:** Actions like `ActionLook` and `ActionTalk` set their final action type from the start (not `ActionMove`). The intent creator owns the action type; `continueIntent` just recalculates paths via the generic fallthrough. The handler in `apply_actions.go` has a walking phase (moves via `moveWithCollision` while not yet at target) and an acting phase (performs the action when arrived). This follows **Consistency Over Local Cleverness** (Values.md) — all walk-then-act actions share the same shape.

**Intent target contract:** `Target` is the *next step* (one BFS move from current position); `Dest` is the *final destination*. Intent creators must call `NextStepBFS` to calculate `Target` — never set `Target` directly to the destination. `moveWithCollision` and `continueIntent` both consume `Target` as a single-step move, not a teleport destination.

### Adding New Actions

Three checklists organized by category. Each includes every touchpoint; see the sections above for the rationale behind each pattern.

**Adding a Need-Driven Action** (e.g., ActionConsume, ActionDrink, ActionSleep):

1. Action constant in `character.go`
2. Intent finder in `intent.go` (driven by stat urgency tiers)
3. Handler method in `apply_actions.go` — add a named `applyXxx` method on `Model` and a case in the `applyIntent` dispatch table; performs the action, clears intent when stat satisfied or source exhausted
4. `continueIntent`: if `TargetItem` can be in inventory, add an early-return block. If `TargetItem` is always on the map, generic path handles it. (See `continueIntent` Rules above.)
5. `simulation.go`: add handler if simulation tests exercise this action
6. No activity registry entry needed (need-driven actions aren't idle activities)

**Adding an Idle Activity** (e.g., ActionLook, ActionTalk, ActionForage, ActionFillVessel):

1. Action constant in `character.go`
2. Activity entry in `ActivityRegistry` (`entity/activity.go`) — omit if the activity is a pre-roll override (like `ActionHelpFeed`, `ActionHelpWater`) rather than a rolled idle activity
3. Intent finder (location depends on context: `intent.go` for social/observation, `foraging.go` for food-seeking, `picking.go` for resource-seeking, `helping.go` for crisis response)
4. Wire into `selectDiscretionaryActivity` in `discretionary.go` — either as a pre-roll override (checked before the roll) or as a rollable option. Crisis-response actions (helpFeed, helpWater) wire into `selectHelpingActivity` in `helping.go` instead.
5. Add action to `isDiscretionaryAction` in `discretionary.go` if the action should be treated as discretionary for talking availability (i.e., it is a true leisure activity, not a helping/delivery action). `isDiscretionaryAction` checks `ActionType` enum — simpler and more robust than string matching.
6. Handler method in `apply_actions.go` — add a named `applyXxx` method on `Model` and a case in the `applyIntent` dispatch table
7. `continueIntent`: self-managing multi-phase actions (where `TargetItem` moves between ground and inventory) need an early-return block. Walk-then-act actions (Look, Talk) use the generic path. (See `continueIntent` Rules above.)
8. `simulation.go`: add handler if simulation exercises this activity

**Adding an Ordered Action** (e.g., ActionTillSoil, ActionPlant, ActionWaterGarden, ActionCraft):

1. Action constant in `character.go`
2. Activity entry in `ActivityRegistry` (with `IntentOrderable` and appropriate `Category`)
3. `findXxxIntent()` in `order_execution.go` — handles target selection on each resumption tick. **Position-based intents** (no `TargetItem`, e.g., TillSoil, Plant, Dig) bypass `continueIntent` and recalculate each tick — use `nextStepBFSCore(... char.UsingBFS)` and set `char.UsingBFS = true` when BFS is used, so sticky BFS survives across ticks. Item-based intents flow through `continueIntent` which handles this automatically. **Adjacent-tile variant** (e.g., BuildFence): set `TargetBuildPos` to the work tile and `Dest` to the adjacent standing tile — `continueIntent`'s generic fallthrough handles navigation toward `Dest`. See `continueIntent` Rules above for the non-need continuation gate requirement.
4. Wire into `findOrderIntent` switch, `isMultiStepOrderComplete`, `IsOrderFeasible`
5. Handler method in `apply_actions.go` — add a named `applyXxx` method on `Model` and a case in the `applyIntent` dispatch table; complete one work unit, clear intent, check order completion inline
5a. **Completion criteria**: every ordered action must define and test both (a) inline completion in the handler (checked after each work unit) and (b) a safety-net case in `isMultiStepOrderComplete` (checked each tick before resuming). Missing either allows the order to loop forever if world state changes between ticks. Add a regression test that exercises the completion boundary (e.g., full inventory, no remaining targets).
6. `continueIntent`: multi-phase ordered actions with vessel procurement (e.g., WaterGarden) need an early-return block. Single-phase ordered actions (TillSoil, Plant) use the generic path. (See `continueIntent` Rules above.)
7. `simulation.go`: add handler if simulation exercises this action
8. If the action uses vessel procurement: use `RunVesselProcurement` tick helper
9. If the action uses water fill: use `RunWaterFill` tick helper

**Behavioral details the plan must specify** (these are design decisions, not implementation details — the plan should resolve them before implementation begins):
- **Targeting**: same-tile or adjacent? (Most actions use same-tile; Look uses adjacent.)
- **Duration**: which `ActionDuration` constant? (Short, Medium, Long)
- **Completion criteria**: what makes a multi-step order complete? (e.g., "no more items of locked variety on map")
- **Feasibility criteria**: what makes the order unfulfillable? (greyed out, skipped during assignment). **Alignment rule**: `IsOrderFeasible` must use the same tile/item eligibility definition as the execution intent (`findXxxIntent`). If execution skips tiles with growing plants but `IsOrderFeasible` counts those tiles as available, characters will take the order, find no valid target, and abandon — looping until feasibility is fixed.
- **Variety lock**: does the order lock to a specific variety on first action? (Harvest, Plant, Extract do; TillSoil, Craft, Dig don't.)
- **DisplayName suffix**: does the order display name include a target type suffix? (Most do: `activity.Name + " " + Pluralize(targetType)`. Activities whose name already implies the target — like "Dig Clay" — return `activity.Name` alone. Add a case to `DisplayName()` in `order.go`.)
- **Sub-menu (step 1)**: does the order need a target type selection sub-menu? If there is only one possible target type (e.g., clay for dig), skip step 1: create the order immediately at step 0 in `applyOrdersConfirm` and omit the step-1 rendering branch in `view.go`. If multiple target types exist, add `GetXxxTypes()` in `variety_generation.go` and wire into the step-1 nav, confirm, and render paths.

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

**Category field**: Groups orderable activities for the order UI menu hierarchy. `getOrderableActivities()` generates synthetic category entries (e.g., "Craft", "Garden", "Construction") when any known activity has that category. Uncategorized activities (e.g., Harvest) appear at the top level. Known categories: `"craft"`, `"garden"`, `"construction"`.

### Discovery Triggers

Know-how activities are discovered through experience. Each trigger specifies an action, optional item type, and optional requirements (`RequiresEdible`, `RequiresPlantable`, `RequiresHarvestable`).

`RequiresHarvestable` gates discovery on any growing non-sprout plant (`Plant != nil && Plant.IsGrowing && !Plant.IsSprout`). This enables non-edible plants like grass and flowers to trigger harvest know-how. `RequiresEdible` and `RequiresPlantable` follow the same bool-gate pattern.

**Construct-triggered discovery:** A trigger with `ConstructKind` set is a construct-only trigger — it fires when a character looks at a construct of that kind, and is explicitly rejected during item interactions. `CompleteLookAtConstruct` calls `TryDiscoverFromConstruct(char, construct, material)`, which runs the same activity-then-recipe search as item-based discovery but matches only triggers with a `ConstructKind` field. When a trigger also has `ConstructMaterial` set, it only matches constructs of that specific material — enabling per-material recipe discovery. This is a parallel path to item-based discovery, not an extension of it.

Example: Harvest is discovered by foraging/eating edible items, or by picking up or looking at any harvestable plant (including non-edible plants like grass and flowers). Plant is discovered by picking up or looking at plantable items. Hut recipes are discovered by looking at a fence of matching material — `ConstructKind: "fence"` and `ConstructMaterial: "stick"` on the stick-hut trigger, so only a stick fence reveals the stick-hut recipe.

### Adding a New Activity

This checklist covers the activity registry specifically. For the full action system touchpoints (handler, intent finder, `continueIntent`), see "Adding New Actions" above.

1. Add entry to `ActivityRegistry` in `entity/activity.go`
2. Set `Category` for grouped activities (generates synthetic category entries in order UI)
3. Add discovery triggers if `AvailabilityKnowHow`
4. If orderable: add case to `findOrderIntent()` switch in `order_execution.go`
5. **If introducing a new category:** grep for all code that branches on activity category (e.g., `activityCategoryVerb`, `categoryDisplayName`). Each branch site must handle the new category, and any user-facing strings (log messages, display labels) must be explicitly stated in the step spec. If a branch site has accumulated 3+ cases, consider replacing it with a map lookup driven by the `ActivityRegistry` or a category metadata struct — prefer dynamic/data-driven over repeated switch statements.

## Orders

Orders are player-directed tasks. They share physical actions with idle activities but have different triggering, completion, and interruption semantics.

### Order Execution

- `CalculateIntent` calls `selectOrderActivity()` before `selectDiscretionaryActivity`, giving orders priority over discretionary activities
- Order eligibility checks activity's `Availability` against character's known activities
- `IsOrderFeasible(order, items, gameMap)` is computed on demand at assignment and render time — returns `(feasible bool, noKnowHow bool)`. Unfeasible orders are skipped during `findAvailableOrder` and rendered dimmed with `[Unfulfillable]` or `[No one knows how]`.
- `LockedVariety string` on Order: set when the first item is planted. After locking, the character only seeks items of that variety, keeping a single order focused.
- **Full bundle exclusion from gather targeting**: `FindNextGatherTarget` skips items that would merge into a bundle that's already at max size — the same way full vessels are skipped during harvest targeting. This prevents continuation past bundle capacity.

### Unified Order Completion

All order types call `CompleteOrder()`, which sets `OrderCompleted` status. A sweep in the game loop after intent application removes all `OrderCompleted` orders that tick. Action handlers contain inline completion checks; `selectOrderActivity` and `findAvailableOrder` skip `OrderCompleted` orders as a safety net.

### Marked-for-Tilling Pool

Tilling separates the player's plan from worker assignments:
- **Marked tiles** (`gameMap.markedForTilling`): User's tilling plan, persistent, independent of orders
- **Till Soil orders**: Worker assignments. Multiple orders = multiple workers on the shared pool.
- Cancelling an order removes the worker, not the plan. Unmarking tiles removes from plan (via area selection in unmark mode).

Pool is serialized in `SaveState.MarkedForTillingPositions`.

### Marked-for-Construction Pool

Construction uses a parallel pool following the same separation of plan from worker:
- **Marked tiles** (`gameMap.markedForConstruction map[Position]ConstructionMark`): User's construction plan. Each `ConstructionMark` carries a `LineID int` (which line-drawing operation created it), `Material string` (empty until first tile is built), and `ConstructKind string` (`"fence"` or `"hut"`).
- **Build Fence / Build Hut orders**: Worker assignments. Multiple orders = multiple workers on the same pool. `HasUnbuiltConstructionPositions(constructKind string)` filters the shared pool by kind — fence orders only count fence marks, hut orders only count hut marks.
- **Line ID**: All tiles from one placement operation share a `LineID`. For fences, this is a cardinal line. For huts, this is the full 16-tile perimeter of one footprint. When the first tile in a line is built, the chosen material is stamped onto all tiles with the same `LineID`. `UnmarkByLineID(lineID)` removes all marks with that ID — used for whole-footprint hut unmarking.
- **Material lock**: Once set, a line's material is read by subsequent workers. If the locked material runs out, the order becomes unfulfillable.
- **Kind filter rule**: `findBuildFenceIntent` passes `"fence"` to `HasUnbuiltConstructionPositions`; `IsOrderFeasible` and `isMultiStepOrderComplete` filter by the order's kind. Adding a new construct kind means adding a `ConstructKind` value and filtering callers by it — no new pool.
- **Feasibility (coarse gate, DD-44)**: `IsOrderFeasible` for construction checks for the existence of at least one free (un-staged) construction material — not a full supply count. `constructionMaterialFeasible` counts only items *not* at construction-marked positions. Items already staged at a marked tile are committed to that site and excluded. Locked lines check only for matching material type; unlocked lines accept any construction material.
- **Pickup loop prevention**: `findNearestMaterialNotAtSite` (shared by fence and hut procurement) skips items whose position is a construction-marked tile. This prevents a character from picking up material they just delivered, re-delivering it, and looping. Used by `findBundleHutIntent`, `findBrickHutIntent`, and brick fence procurement.
- **Transient nil guard**: When a construction intent finder returns nil (e.g., all candidate tiles temporarily occupied by other workers), the caller checks `IsOrderFeasible` before abandoning — if still feasible, the nil is treated as a transient block rather than exhaustion, and the order is not abandoned. See triggered-enhancements.md "Typed nil-intent signals" for the known limitation of this approach.

Pool is serialized in `SaveState.MarkedForConstructionTiles`. Line ID counter in `SaveState.ConstructionLineID`. `ConstructKind` is serialized on each mark with backward compatibility: old saves without `ConstructKind` default to `"fence"`.

### Shared Construction Helpers

Construction intent finders share helpers in `order_execution.go` to avoid duplication:
- **`selectConstructionMaterial(char, world, items, activityID)`**: Selects the preferred available construction material for the given activity (fence or hut). Replaces `selectFenceMaterial` — shared by both fence and hut building. Scores each feasible recipe via `ScoreConstructPreference` (builds a synthetic Construct, scores character preferences against it with Kind weighted 2×) combined with distance via `ScoreItemFit`. No preferences → nearest wins (preserves prior behavior).
- **`findNearestMaterialNotAtSite(items, gameMap, materialType)`**: Finds the nearest item of a given type whose position is not a construction-marked tile. Prevents pickup loops where a character re-picks staged materials.
- **`createHutDeliveryIntent(char, item, targetPos)`**: Shared delivery intent constructor for hut bundle/brick handlers — sets `TargetBuildPos`, `Dest` (adjacent standing tile), and `TargetItem`.

In `apply_actions.go`, `deliverMaterial` is a generic delivery helper shared across ordered construction actions — it handles the drop-at-site phase for the supply-drop pattern.

## Item Acquisition

`picking.go` is the shared home for all item acquisition logic, organized in three layers:

1. **Map Search** — `findNearestItemByType`, `findPreferredItemByType` (preference + distance scoring), `FindAvailableVessel` (vessels that can receive items), `FindVesselContaining` (vessels whose contents match a target)
2. **Prerequisite Orchestration** — `EnsureHasVesselFor`, `EnsureHasRecipeInputs`, `EnsureHasItem` (check-or-go-get helpers; use `findPreferredItemByType` for preference-weighted targeting)
3. **Inventory Query** — `FindCarriedVesselFor(char, item, registry)` finds the first carried vessel that can accept a specific item. Use this instead of `GetCarriedVessel()` when vessel compatibility matters — `GetCarriedVessel()` only returns the first vessel, which may be incompatible (e.g., holds water when you need to store berries).
4. **Physical Actions** — `Pickup`, `Drop`, vessel operations

### Pickup Result Pattern

`Pickup()` returns a `PickupResult` to distinguish outcomes:
- `PickupToInventory` — item picked up directly (inventory was empty)
- `PickupToVessel` — item added to carried vessel's stack
- `PickupToBundle` — item merged into carried bundle (count incremented)
- `PickupFailed` — could not pick up (variety mismatch with vessel)

`PickupToBundle` and `PickupToVessel` share the same continuation semantics: the caller (gather/harvest handler) continues working rather than clearing intent. Both also skip intent clearing — the action keeps going. Callers that need to check for bundle completion (gather orders) additionally call `CanPickUpMore()` after the merge; if the bundle is full, they drop the bundle and complete the order.

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

### Preference-Weighted Item Seeking

`selectConstructionMaterial` uses preference-weighted scoring via `ScoreConstructPreference` + `ScoreItemFit` (Step 11a). `EnsureHasItem` and `EnsureHasRecipeInputs` use `findPreferredItemByType` — preference + distance scoring — replacing pure nearest-distance (Step 11b). Remaining gap: `FindAvailableVessel`, `EnsureHasVesselFor`, and `EnsureHasPlantable` (before variety lock) still use nearest-distance (Step 11d). See `triggered-enhancements.md` for the "Generic scored-item search helper" trigger.

## Recipe System

Recipes define how to craft items from components.

### Crafting Flow

1. Order assigned → `findCraftIntent` gets recipes for activity, filters to feasible recipes
2. Picks first feasible recipe (future: preference-weighted selection)
3. `EnsureHasRecipeInputs` gathers missing components
4. Once all components accessible: perform crafting action
5. On completion: consume all inputs, create output item via recipe-specific creation function

### Repeatable Recipes

Recipes with `Repeatable: true` loop after each craft cycle instead of completing the order. After crafting one output, `applyCraft` clears intent (no inline `CompleteOrder`). The next tick, `selectOrderActivity` re-evaluates: if `findCraftIntent` finds more inputs, it crafts again; if `isMultiStepOrderComplete` returns true (e.g., no clay on map), the order completes. This enables "process all available material" orders (e.g., clay-brick) without quantity selection UI. Non-repeatable recipes complete the order immediately after one craft.

### Adding a New Recipe

1. Add recipe to `RecipeRegistry` in `entity/recipe.go`. Set `Repeatable: true` if the order should loop until a world-state condition is met.
2. Add activity to `ActivityRegistry` for the crafting activity (if new)
3. Add creation function in `system/crafting.go` (e.g., `CreateHoe`)
4. Add case to the `ActionCraft` handler in `apply_actions.go` dispatch (by recipe ID)
5. If `Repeatable`: add completion case to `isMultiStepOrderComplete` in `order_execution.go`
6. If discovery is triggered by looking at a construct: set `ConstructKind` on the recipe's `DiscoveryTrigger`. If the recipe should only be discovered from a specific material variant (e.g., stick-hut from stick fences only), also set `ConstructMaterial`. See [Construct-triggered discovery](#discovery-triggers). These triggers fire from `TryDiscoverFromConstruct` (called by `CompleteLookAtConstruct`), not from item interactions.

## Area Selection UI Pattern

Area selection enables players to define regions for terrain modification or construction. Three selection shapes are supported:

- **Rectangle** (tillSoil): anchor + move cursor → fills rectangle between anchor and cursor
- **Line** (buildFence): anchor + move cursor → snaps to a cardinal line (horizontal or vertical, whichever axis has the larger delta); diagonal cursor movement resolves to the dominant axis
- **Fixed footprint** (buildHut): no anchor — cursor positions the top-left corner of a fixed 5×5 footprint; `p` confirms in one press

**Flow (rectangle and line):**
1. Player selects activity → enters area selection mode
2. Move cursor, press `p` to anchor start point
3. Move cursor to resize selection (valid tiles highlighted)
4. Press `p` to confirm (marks tiles)
5. Press `Tab` to toggle mark/unmark mode
6. Press `Enter` when done, `Esc` to cancel

**Flow (fixed footprint):**
1. Player selects activity → enters placement mode (no anchor)
2. Move cursor to position the footprint preview
3. Press `p` to place (marks tiles immediately, no anchor step)
4. Press `Tab` to toggle mark/unmark mode; in unmark mode, hover over any marked tile and press `p` to remove the entire footprint by LineID
5. Press `Enter` when done, `Esc` to cancel

**Rendering**: Selection highlight uses full background for empty tiles, padding-only for entities (avoids ANSI nesting). Three distinct color states map directly to workflow progression: active selection (teal/warm brown) → confirmed pending (sage/olive) → completed work (dusky earth). Each state should be visually distinct from the others — reusing a color across states collapses workflow stages the player needs to distinguish.

**Confirmed marks visibility**: Confirmed marks are only highlighted during the active marking phase for their activity (step 2). In regular select mode, the cursor over a marked tile reveals the status in the details panel. This matches the principle that persistent UI state should not clutter the default view.

### Reuse for Future Activities

When adding new area-based orders:
1. Add activity check in step 1 Enter handler
2. Choose a selection shape (rectangle, line, or fixed footprint) and write an activity-specific validator function (like `isValidTillTarget`, `isValidFenceTarget`, or `isValidHutFootprint`)
3. For rectangle/line: reuse `getValidPositions` or `getValidLinePositions` with custom validator. For fixed footprint: compute positions directly from cursor (top-left anchor convention) without a separate `getValid*` helper
4. Handle confirm logic in `p` key handler (fixed footprint: no anchor state needed — `p` always confirms immediately)
5. Add confirmed-marks rendering inside the `m.ordersAddMode && step2 && activityID == X` block (not outside it)

## Save/Load Serialization

Save files stored in `~/.petri/worlds/world-XXXX/` with `state.json`, `state.backup`, and `meta.json`.

### Serialization Checklist

When adding fields to saved structs:

1. **Display fields**: Symbols (`Sym`), colors, styles set by constructors — must be explicitly restored on deserialization
2. **All attribute fields**: Easy to miss nested fields (e.g., Pattern/Texture on preferences)
3. **Round-trip tests**: Save → load → verify all fields match
4. **Variety fields**: `VarietySave` must include all fields that `ConsumeAccessibleItem` / `ConsumePlantable` need to restore (currently: `Kind`, `Plantable`, `Sym`, `SourceVarietyID`)

Constructor-set fields won't be populated when deserializing directly into structs — must be explicitly restored based on type.

**PlantProperties serialization**: `IsSprout` and `SproutTimer` must round-trip correctly to preserve sprout state across save/load.

**BundleCount serialization**: `BundleCount` is stored as `bundle_count` in `ItemSave`. Backward compatibility: old saves without the field default to `BundleCount=0`; on load, bundleable types with `BundleCount=0` are migrated to `BundleCount=1` (a single item is a bundle of 1).

**Kind migration**: When a new Kind value is added to an existing item type, old saves that pre-date the Kind field have `Kind=""` for those items. Migration in `FromSaveState` (`serialize.go`) restores the expected Kind (e.g., grass items without Kind get `Kind="tall grass"` on load). Follow this pattern for any future item type that gains a Kind retroactively.

**Save compatibility when changing entity storage**: When changing how entities are stored (e.g., moving data between fields, maps, or types), verify save/load round-trip in the same step. Check: (1) new state serializes, (2) old saves migrate, (3) serialize tests updated.

## Common Implementation Pitfalls

**Game time vs wall clock**: UI indicators that should work when paused (like "Saved" message) need wall clock time (`time.Now()`), not game time which only advances when unpaused.

**Sorting stability**: When displaying merged data from maps (e.g., AllEvents from ActionLog), use `sort.SliceStable` with deterministic tiebreakers (like CharID) to prevent visual jitter from Go's random map iteration order.

**View transitions**: When switching between views with different rendering approaches (game view uses direct rendering, menus use lipgloss.Place for centering), add dimension safeguards for edge cases.

**Terrain fill in `renderCell()`**: Terrain that renders as solid blocks (tilled soil `═══`, water `▓▓▓`) requires both `sym` AND `fill` set to the styled terrain character. Setting only `sym` produces a single character flanked by spaces (` ▓ `), creating a vertical stripe appearance.
