Guidance for Claude Code when working with code in this repo.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development.

**Current features:** Procedural world generation, urgency-based survival AI, social behavior with preferences and knowledge, crafting and inventory, player-directed orders.

**Vision:** Complex roguelike simulation world with complex interactions between characters, items, and attributes. History exists only in character memories and created artifacts. As characters die, their knowledge dies with them except what they've communicated or created. See [docs/VISION.txt](docs/VISION.txt).

---

## Quick Commands

Build and run
go build -o petri ./cmd/petri
./petri

Test flags
./petri -no-food -no-water -no-beds -debug -mushrooms-only -help

Testing
go test ./...
go test -race ./...

## Architecture

Go project using Bubble Tea's Model-View-Update (MVU) pattern. See docs/architecture.md for details.

### File Structure

```
cmd/petri/main.go           # Entry point, flag parsing
internal/
  config/config.go          # Game constants
  types/types.go            # Shared type constants (Color, StatType)
  entity/                   # Character, Item, Feature, Construct
  game/                     # Map, world generation
  save/                     # Save/load, world management, serialization types
  system/                   # Movement, survival, consumption, action log
  ui/                       # Bubble Tea model, update, view, styles
  simulation/               # Integration test utilities, balance observation tests
```

### Key Files for Context

**Configuration & Types**

- `internal/config/config.go` - Tunable game constants (rates, thresholds, durations). Also contains legacy item property maps (`StackSize`, `ItemMealSize`, etc.) that should migrate to entity registries — see triggered-enhancements.md
- `internal/types/types.go` - Core enums: Color, Pattern, Texture, StatType
- `internal/types/position.go` - Position struct with distance/adjacency methods

**Entity Definitions**

- `internal/entity/character.go` - Character struct, stat tiers, status effects
- `internal/entity/item.go` - Item struct, spawning, descriptions
- `internal/entity/variety.go` - ItemVariety for world generation
- `internal/entity/preference.go` - Preference struct, matching logic
- `internal/entity/knowledge.go` - Knowledge struct, learning from experience
- `internal/entity/activity.go` - Activity struct, ActivityRegistry, know-how discovery triggers
- `internal/entity/recipe.go` - Recipe struct, RecipeRegistry, crafting definitions
- `internal/entity/order.go` - Order struct for player-directed tasks
- `internal/entity/construct.go` - Construct struct (ConstructType/Kind/Material/MaterialColor/Passable/Movable); NewFence constructor

**Core Systems**

- `internal/system/intent.go` - Intent calculation, four-bucket routing (Needs → Orders → Helping → Discretionary), urgency tiers, stat fallback, need evaluators (findFoodIntent, findDrinkIntent, etc.), feasibility checks, carried-inventory checks
- `internal/system/movement.go` - Spatial mechanics only: NextStepBFS, pathfinding, tile queries, adjacency checks
- `internal/system/survival.go` - Stat decay, damage, sleep/wake mechanics
- `internal/system/consumption.go` - Eating, drinking, poison/healing effects
- `internal/system/preference.go` - Preference formation on eat/look
- `internal/system/talking.go` - Talking state, knowledge transmission (LearnKnowledgeWithEffects)
- `internal/system/discretionary.go` - Discretionary activity selection, selectDiscretionaryActivity, isDiscretionaryAction (ActionType-based discretionary check)
- `internal/system/helping.go` - Crisis detection, findHelpFeedIntent/findHelpWaterIntent (food/water delivery to crisis characters)
- `internal/system/order_execution.go` - Order assignment, intent finding, completion/abandonment
- `internal/system/crafting.go` - CreateVessel, crafted item creation
- `internal/system/picking.go` - Pickup, Drop, vessel helpers, EnsureHasVesselFor, FindCarriedVesselFor, EnsureHasItem
- `internal/system/foraging.go` - Foraging intent, unified scoring

**World & Generation**

- `internal/game/world.go` - World struct, entity management; constructs slice, AddConstruct/AddConstructDirect/ConstructAt/Constructs/RemoveConstruct + ID methods
- `internal/game/variety_generation.go` - Variety creation, poison/healing assignment

**Game Loop**

- `internal/ui/model.go` - Game state, Bubble Tea model
- `internal/ui/update.go` - Tick processing, game loop orchestration, input handling, lifecycle (startGame, loadWorld, saveGame)
- `internal/ui/apply_actions.go` - Intent execution: applyIntent dispatch table + 17 handler methods (applyMove, applyDrink, applySleep, applyLook, applyTalk, applyPickup, applyConsume, applyCraft, applyTillSoil, applyPlant, applyForage, applyFillVessel, applyWaterGarden, applyHelpFeed, applyHelpWater, applyExtract, applyDig) + execution helpers (moveWithCollision, findEmptyCardinalTile, etc.)

**Save System**

- `internal/save/state.go` - SaveState struct, all serialization types
- `internal/save/io.go` - World management, file I/O, backup rotation
- `internal/ui/serialize.go` - ToSaveState/FromSaveState conversion

## Collaboration

**Pause -> Skill -> Discussion → Tests -> Implementation → Human Testing → Documentation**

- **Before writing code**: Load the relevant skill, check architecture.md, discuss approach. Use project skills (`/refine-feature`, `/implement-feature`, `/new-phase`, `/retro`) — not generic plan mode. Phase design docs and step specs live in `docs/`.
- **Frame the problem first**: Before solving, confirm the problem — is this actually a problem? Is it worth solving now? Specify current state and desired state in functional terms. Consider impact on the larger system. Confirm alignment before implementing. When human testing surfaces a new gap mid-implementation, that's a new problem to scope and discuss, not a defect to fix in place.
- **Communication**: Functional terms, not code mechanics. Prose for tradeoffs, not multiple-choice. Recommend with options. Qualify claims precisely. When claiming alignment with patterns or docs, cite specific evidence — assertions aren't demonstrations.
- **When things go wrong**: Gather evidence first — examine the save file, add logging, ask what the user observes — before reasoning about causes. Don't be dismissive of user reports. The cost of one diagnostic is always lower than three rounds of speculation. Second bug in the same feature = step back and restate the intended flow before patching further. Surface when stuck.
- **Quality gates**: TDD. User must test before marking complete. Keep docs current.

## Testing

- Add regression tests when making bug fixes
- No tests needed for UI rendering, Bubble Tea integration, brittle string matching (log wording, display names, UI text), configuration constants
- Don't test for absence unless requirements call for prevention
- Headless simulation tests for measuring game balance. Located in `internal/simulation/observation_test.go`.

## Development Roadmap

**Up Next:**

- Construction Phase: In progress. See [docs/construction-design.md](docs/construction-design.md) for design and [docs/step-spec.md](docs/step-spec.md) for current step.

## Reference and Planning Documents

| Document                                                         | Purpose                                       |
| ---------------------------------------------------------------- | --------------------------------------------- |
| [docs/VISION.txt](docs/VISION.txt)                               | Project vision and phases                     |
| [docs/architecture.md](docs/architecture.md)                     | Design patterns, decision rationale, "adding new X" checklists |
| [docs/game-mechanics.md](docs/game-mechanics.md)                 | Detailed stat thresholds, rates, systems      |
| [docs/triggered-enhancements.md](docs/triggered-enhancements.md) | Deferred items with triggers, balance tuning  |
| [docs/construction-design.md](docs/construction-design.md)       | Construction phase design: steps, decisions, scope |
| [docs/step-spec.md](docs/step-spec.md)                           | Current step implementation spec (replaced each step) |
| [docs/post-gardening-cleanup.md](docs/post-gardening-cleanup.md) | Small improvements to bundle after Gardening  |
