# CLAUDE.md

Guidance for Claude Code when working with code in this repo.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development.

**Current features:** Character creation, world generation with items and features, multi-stat survival, dynamic preference system, urgency-based AI with stat fallback and frustration mechanics, view modes, and action logging.

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

**Core Systems**
- `internal/system/intent.go` - Intent calculation, urgency tiers, stat fallback
- `internal/system/survival.go` - Stat decay, damage, sleep/wake mechanics
- `internal/system/consumption.go` - Eating, drinking, poison/healing effects
- `internal/system/preference.go` - Preference formation on eat/look

**World & Generation**
- `internal/game/world.go` - World struct, entity management
- `internal/game/variety_generation.go` - Variety creation, poison/healing assignment

**Game Loop**
- `internal/ui/model.go` - Game state, Bubble Tea model
- `internal/ui/update.go` - Tick processing, intent application

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

| Test | Purpose |
|------|---------|
| `TestObserveBalanceMetrics` | 5 runs × 300s: survival rate, mood distribution, preferences |
| `TestObserveFoodScarcity` | Tracks food availability vs consumption over time |
| `TestObserveFlowerGrowth` | Monitors flower population growth |
| `TestObserveTimeToFirstDeath` | 10 runs: measures time until first death |
| `TestObserveDeathProgression` | Single extended run tracking all deaths |

Results are documented in `docs/futureEnancements.md` under "Balance Observation Results".

## Development

### Current Work

**Phase 3 - WRAPPING UP**
See: [Plan](docs/Phase%203/phase03-plan.md), [Reqs](docs/Phase%203/phase03reqs.txt)

- ✅ A-C. Mood and Preference interactions
- ✅ D. World Balancing (D1-D9 complete)
- ✅ Balance tuning pass (food scarcity, flower overpopulation resolved)
- ✅ Tune preference formation chances (reduced to 10%/5%/5%/10%)

### Near-Term Roadmap

**Phase 3 Closeout**
1. ~~Tune preference formation chances~~ ✅
2. Review audit doc and architecture doc for archival/consolidation
3. Create session progress tracking document for Phase 4+

**Phase 4 Prep**
- Feature flags cleanup / migrate to Cobra CLI (cleaner `--help`, new phase may need flags like `--no-talking`, `--debug-knowledge`)

**Phase 4: Basic Knowledge & Transmission**
See: [Reqs](docs/Phase%204/phase04reqs.txt)

| Sub-phase | Description | Opportunistic Items |
|-----------|-------------|---------------------|
| A | Knowledge by experience (learn poison/healing from eating) | |
| B | Knowledge panel UI (toggle in select mode) | ESC key: return to start screen vs quit |
| C | Action log: "[Char] learned something!" | Action log audit: vertical space, limits; L log alignment with action log |
| D | Poison knowledge → dislike preference | |
| E-F | Healing knowledge → seek healing intent + conditional food matching | |
| G | Talking as idle activity (5s duration, targets idle chars) | |
| H | Knowledge transmission via talking | |

**Post-Phase 4: UI Polish Pass**
Low-priority items to address after core Phase 4 features:
- Reorder fields in Details panel
- Assess unicode symbols for entities
- Start screen improvements (title, key hints)
- Loading screen
- ASCII art mushrooms on title screen

**Pre-Phase 5 Decision Point**
- Save game feature: assess placement before or during Phase 5 (Resources/Inventory)
- Rationale: Save becomes more valuable as game state complexity increases

### Deferred Enhancements & Trigger Points

Technical items analyzed and consciously deferred until trigger conditions are met. See docs/futureEnancements.md for more.

| Enhancement                     | Triggers (implement when ANY is met)                                        |
| ------------------------------- | --------------------------------------------------------------------------- |
| **Parallel Intent Calculation** | Character count ≥ 16; Intent calc exceeds ~1ms/char; Tick time exceeds 50ms |
| **EventType for ActionLog**     | Event filtering in UI; Event-driven behavior (Phase III); Event persistence |
| **Performance optimizations**   | Noticeable lag; Profiling shows bottlenecks                                 |
| **Testify test package**        | Test complexity warrants it; Current assertion patterns become unwieldy     |

## Reference Documents

| Document                                               | Purpose                                  |
| ------------------------------------------------------ | ---------------------------------------- |
| [docs/VISION.txt](docs/VISION.txt)                     | Project vision and phases                |
| [docs/game-mechanics.md](docs/game-mechanics.md)       | Detailed stat thresholds, rates, systems |
| [docs/failed-approaches.md](docs/failed-approaches.md) | Approaches tried and abandoned           |
