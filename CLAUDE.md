Guidance for Claude Code when working with code in this repo.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development.

**Current features:** Procedural world generation, multi-stat survival with urgency-based AI, preferences and knowledge, social behavior, inventory with vessels and bundles, crafting, player-directed orders (gardening, gathering, construction), crisis helping, and action logging.

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
  entity/                   # Character, Item, Feature
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
- `internal/system/picking.go` - Pickup, Drop, vessel helpers, EnsureHasVesselFor, EnsureHasItem
- `internal/system/foraging.go` - Foraging intent, unified scoring

**World & Generation**

- `internal/game/world.go` - World struct, entity management
- `internal/game/variety_generation.go` - Variety creation, poison/healing assignment

**Game Loop**

- `internal/ui/model.go` - Game state, Bubble Tea model
- `internal/ui/update.go` - Tick processing, game loop orchestration, input handling, lifecycle (startGame, loadWorld, saveGame)
- `internal/ui/apply_actions.go` - Intent execution: applyIntent dispatch table + 15 handler methods (applyMove, applyDrink, applySleep, applyLook, applyTalk, applyPickup, applyConsume, applyCraft, applyTillSoil, applyPlant, applyForage, applyFillVessel, applyWaterGarden, applyHelpFeed, applyHelpWater) + execution helpers (moveWithCollision, findEmptyCardinalTile, etc.)

**Save System**

- `internal/save/state.go` - SaveState struct, all serialization types
- `internal/save/io.go` - World management, file I/O, backup rotation
- `internal/ui/serialize.go` - ToSaveState/FromSaveState conversion

## Collaboration

**Pause -> Skill -> Discussion → Tests -> Implementation → Human Testing → Documentation**

- **Before writing code**: Load the relevant skill, check architecture.md, discuss approach. Use project skills (`/refine-feature`, `/implement-feature`, `/new-phase`, `/retro`) — not generic plan mode. Plan docs live in `docs/`; update existing ones.
- **Frame the problem first**: Specify current state and desired state in functional terms before proposing changes. Consider impact on the larger system. Confirm alignment before implementing.
- **Communication**: Functional terms, not code mechanics. Prose for tradeoffs, not multiple-choice. Recommend with options. Qualify claims precisely.
- **When things go wrong**: Evidence before fixes. Second bug = step back and restate the flow. Surface when stuck.
- **Quality gates**: TDD. User must test before marking complete. Keep docs current.

## Testing

- Add regression tests when making bug fixes
- No tests needed for UI rendering, Bubble Tea integration, brittle string matching (log wording, display names, UI text), configuration constants
- Headless simulation tests for measuring game balance. Located in `internal/simulation/observation_test.go`.

## Development Roadmap

**Up Next:**

- Construction Phase in progress: Steps 1a–1b (bundles + tall grass) complete. See [docs/construction-phase-plan.md](docs/construction-phase-plan.md).

## Reference and Planning Documents

| Document                                                         | Purpose                                       |
| ---------------------------------------------------------------- | --------------------------------------------- |
| [docs/VISION.txt](docs/VISION.txt)                               | Project vision and phases                     |
| [docs/architecture.md](docs/architecture.md)                     | Design patterns, decision rationale, "adding new X" checklists |
| [docs/game-mechanics.md](docs/game-mechanics.md)                 | Detailed stat thresholds, rates, systems      |
| [docs/triggered-enhancements.md](docs/triggered-enhancements.md) | Deferred items with triggers, balance tuning  |
| [docs/post-gardening-cleanup.md](docs/post-gardening-cleanup.md) | Small improvements to bundle after Gardening  |
