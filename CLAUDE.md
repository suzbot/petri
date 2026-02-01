Guidance for Claude Code when working with code in this repo.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development.

**Current features:** Character creation, world generation with items (berries, mushrooms, gourds, flowers, vessels) and features, multi-stat survival, dynamic preference system, knowledge system (learning poison/healing, knowledge-driven behavior, knowledge transmission via talking), social behavior (talking between idle characters), inventory system (foraging to pick up items, dropping items), orders system (player-directed harvesting and crafting), crafting system (hollow gourd vessels), vessel contents (item stacking, variety locking, automatic vessel seeking), know-how discovery, urgency-based AI with stat fallback and frustration mechanics, view modes, and action logging.

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

- `internal/config/config.go` - All game constants, rates, thresholds, spawn counts
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

- `internal/system/intent.go` - Intent calculation, urgency tiers, stat fallback
- `internal/system/survival.go` - Stat decay, damage, sleep/wake mechanics
- `internal/system/consumption.go` - Eating, drinking, poison/healing effects
- `internal/system/preference.go` - Preference formation on eat/look
- `internal/system/talking.go` - Idle activity selection, talking state, knowledge transmission (LearnKnowledgeWithEffects)
- `internal/system/order_execution.go` - Order assignment, intent finding, completion/abandonment
- `internal/system/crafting.go` - CreateVessel, crafted item creation
- `internal/system/picking.go` - Pickup, Drop, vessel helpers, EnsureHasVesselFor
- `internal/system/foraging.go` - Foraging intent, unified scoring

**World & Generation**

- `internal/game/world.go` - World struct, entity management
- `internal/game/variety_generation.go` - Variety creation, poison/healing assignment

**Game Loop**

- `internal/ui/model.go` - Game state, Bubble Tea model
- `internal/ui/update.go` - Tick processing, intent application

**Save System**

- `internal/save/state.go` - SaveState struct, all serialization types
- `internal/save/io.go` - World management, file I/O, backup rotation
- `internal/ui/serialize.go` - ToSaveState/FromSaveState conversion

## Collaboration

**Discussion → Implementation → Human Testing → Documentation**

- **Discussion first**: Always discuss approach before writing code. Present options with trade-offs.
- **TDD**: Write tests before implementation (see docs/testingProcess)
- **Human testing required**: Do NOT mark features or fixes complete until user has manually tested. Passing unit tests is not sufficient - the user must confirm the fix works in actual gameplay before updating docs or marking complete.
- **Keep docs current**: Update claude.md, README, game-mechanics as part of feature work

Process docs:

- When starting a new Phase, always read docs/new-phase-process.md
- When starting a new feature within a phase, always read docs/feature-dev-process.md

## Testing

- TDD process. See docs/testingProcess for details
- Add regression tests when making bug fixes
- No tests needed for UI rendering, Bubble Tea integration, brittle log wording, configuration constants
- Headless simulation tests for measuring game balance. Located in `internal/simulation/observation_test.go`.

## Development Roadmap

**Up Next:**

- Evaluation of docs/randomideas.md against future phases to prioritize anything opportunistically
- Simple Gardening planning and implementation (see [docs/Gardening-Reqs.txt](docs/Gardening-Reqs.txt))

## Reference and Planning Documents

| Document                                                         | Purpose                                       |
| ---------------------------------------------------------------- | --------------------------------------------- |
| [docs/VISION.txt](docs/VISION.txt)                               | Project vision and phases                     |
| [docs/architecture.md](docs/architecture.md)                     | Design patterns, data flow, item/memory model |
| [docs/game-mechanics.md](docs/game-mechanics.md)                 | Detailed stat thresholds, rates, systems      |
| [docs/triggered-enhancements.md](docs/triggered-enhancements.md) | Deferred items with triggers, balance tuning  |
| [docs/failed-approaches.md](docs/failed-approaches.md)           | Approaches tried and abandoned                |
