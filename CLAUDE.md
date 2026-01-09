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
  simulation/               # Integration test utilities
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

## Development

### Current Work

**Phase 3** See: [Plan](docs/Phase%203/phase03-plan.md), [Reqs](docs/Phase%203/phase03reqs.txt)
Sub-phases:

- ✅ A-C. Mood and Preference interactions
- 🔄 **D. World Balancing** ← In progress

**Completed**: D1-D6, D8-D9

**Next**: D7 (Seeking/avoidance refinements)

Key D6 decisions (Preference formation):

- **Solo (30%)**: Any single attribute; Pattern/Texture use noun forms ("Likes Spots", "Likes Slime")
- **Combo (70%)**: Must include ItemType + 1-2 other attributes (max 3 total)
- **Generative approach**: Not hardcoded permutations, scales with future attributes

### Next Priorities

1. Complete Phase III sub-phases (A → D)
2. Balance tuning pass (end of Phase III, see docs/futureEnhancements.md)
3. Post-Phase III: Review audit doc and architecture doc for archival/consolidation. Create appropriate documents for tracking session progress and bigger picture architecture decisions/planning.
4. Feature flags cleanup as per docs/audit (not blocking, can happen anytime), use cobra cli to manage flags going forward. Polish conventions (eg. use --help or -h instead of -help)

### Deferred Enhancements & Trigger Points

Technical items analyzed and consciously deferred until trigger conditions are met. see docs/futureEnhancements.md for more.

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
