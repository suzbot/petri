# CLAUDE.md

Guidance for Claude Code when working with code in this repo.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development.

**Current features:** Character creation, world generation with items (berries, mushrooms, gourds, flowers) and features, multi-stat survival, dynamic preference system, knowledge system (learning poison/healing, knowledge-driven behavior, knowledge transmission via talking), social behavior (talking between idle characters), inventory system (foraging to pick up items), urgency-based AI with stat fallback and frustration mechanics, view modes, and action logging.

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

**Entity Definitions**

- `internal/entity/character.go` - Character struct, stat tiers, status effects
- `internal/entity/item.go` - Item struct, spawning, descriptions
- `internal/entity/variety.go` - ItemVariety for world generation
- `internal/entity/preference.go` - Preference struct, matching logic
- `internal/entity/knowledge.go` - Knowledge struct, learning from experience

**Core Systems**

- `internal/system/intent.go` - Intent calculation, urgency tiers, stat fallback
- `internal/system/survival.go` - Stat decay, damage, sleep/wake mechanics
- `internal/system/consumption.go` - Eating, drinking, poison/healing effects
- `internal/system/preference.go` - Preference formation on eat/look
- `internal/system/talking.go` - Idle activity selection, talking state, knowledge transmission (LearnKnowledgeWithEffects)

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

TTD, Iterative Approach. Frequent discussion. Present options with trade-offs. Frequent human testing checkpoints. Update claude.md, REAMDE, and other documentation along the way.

- When starting a new Phase, see docs/new-phase-process.md
- When starting a new feature within a phase, see docs/feature-dev-process.md

## Testing

- TDD process. See docs/testingProcess for details
- Add regression tests when making bug fixes
- No tests needed for UI rendering, Bubble Tea integration, brittle log wording, configuration constants

### Balance Observation Tests

Headless simulation tests for measuring game balance. Located in `internal/simulation/observation_test.go`.

Run all observation tests:

```bash
go test -v -run TestObserve ./internal/simulation/
```

| Test                          | Purpose                                                      |
| ----------------------------- | ------------------------------------------------------------ |
| `TestObserveBalanceMetrics`   | 5 runs × 300s: survival rate, mood distribution, preferences |
| `TestObserveFoodScarcity`     | Tracks food availability vs consumption over time            |
| `TestObserveFlowerGrowth`     | Monitors flower population growth                            |
| `TestObserveTimeToFirstDeath` | 10 runs: measures time until first death                     |
| `TestObserveDeathProgression` | Single extended run tracking all deaths                      |

Results are documented in `docs/futureEnancements.md` under "Balance Observation Results".

## Development

### Current Work

Phase 5: Picking up Items and Inventory. Sub-phases 5.1 (Item Category & Gourd), 5.2 (Inventory & Foraging), 5.3 (Eating from Inventory), and 5.4 (Know-how System) complete.

### Near-Term Roadmap

- Phase 5.5+ (see docs/phase05-plan.md for full plan)

### Deferred Enhancements & Trigger Points

Technical items analyzed and consciously deferred until trigger conditions are met. See docs/futureEnancements.md for more.

| Enhancement                     | Triggers (implement when ANY is met)                                        |
| ------------------------------- | --------------------------------------------------------------------------- |
| **Parallel Intent Calculation** | Character count ≥ 16; Intent calc exceeds ~1ms/char; Tick time exceeds 50ms |
| **EventType for ActionLog**     | Event filtering in UI; Event-driven behavior (Phase III); Event persistence |
| **Performance optimizations**   | Noticeable lag; Profiling shows bottlenecks                                 |
| **Testify test package**        | Test complexity warrants it; Current assertion patterns become unwieldy     |

## Reference Documents

| Document                                               | Purpose                                       |
| ------------------------------------------------------ | --------------------------------------------- |
| [docs/VISION.txt](docs/VISION.txt)                     | Project vision and phases                     |
| [docs/architecture.md](docs/architecture.md)           | Design patterns, data flow, item/memory model |
| [docs/game-mechanics.md](docs/game-mechanics.md)       | Detailed stat thresholds, rates, systems      |
| [docs/futureEnancements.md](docs/futureEnancements.md) | Deferred items with triggers, balance tuning  |
| [docs/failed-approaches.md](docs/failed-approaches.md) | Approaches tried and abandoned                |
| [docs/phase05-plan.md](docs/phase05-plan.md)           | Phase 5 implementation plan                   |
