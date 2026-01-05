# CLAUDE.md

Guidance for Claude Code when working with code in this repo.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development.

**Current features:** Multi-character mode, character creation screen, multi-stat survival (hunger, thirst, energy, health, mood), dynamic preference system, urgency-based AI with stat fallback and frustration mechanics, landscape features (springs, leaf piles), sleep mechanics, view modes, and action logging.

**Vision:** Complex roguelike simulation world with complex interactions between characters, items, and attributes. History exists only in character memories and created artifacts. As characters die, their knowledge dies with them except what they've communicated or created. See [docs/VISION.txt](docs/VISION.txt).

---

## Quick Commands

Build and run
go build -o petri ./cmd/petri
./petri

Test flags
./petri -no-food -no-water -no-beds -debug -help

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

## Collaboration

TTD, Iterative Approach. Frequent discussion. Present options with trade-offs. Frequent human testing checkpoints. Update claude.md, REAMDE, and other documentation along the way.

- For specific process when starting a new Phase, see docs/new-phase-process.md
- For specific process when starting a new feature within a phase, see docs/feature-dev-process.md

## Testing

- TDD process. See docs/testingProcess for details
- Add regression tests when making bug fixes
- No tests needed for UI rendering, Bubble Tea integration, brittle log wording, configuration constants

## Development

### Current Work

**Phase 3** See: [Plan](docs/phase03-plan.md), [Reqs](docs/phase03reqs.txt)
Sub-phases:

- ✅ A-C. Mood and Preference interactions <- Complete
- 🔄 **D. World Balancing** ← In progress

**Completed**: D1-D4, D8

**D5 (Pattern/Texture)** ← in progress

See `docs/phase03-plan.md` for implementation plan. Key decisions:

- **Variety for generation only**: Varieties define what combos exist and assign poison/healing at world gen
- **Items keep embedded fields**: No registry lookups at runtime, simpler architecture
- Variety count = max(2, spawnCount / 4), 20% poisonous, 20% healing
- Pattern: Spotted, None (mushrooms only)
- Texture: Slimy, None (mushrooms only)

### Next Priorities

1. Complete Phase III sub-phases (A → B → C → D)
2. Balance tuning pass (end of Phase III, see docs/futureEnhancements.md)
3. Improve README for Phase II roll-out
   - Add a section for Latest Updates
   - Make feature list more high level and concise
   - Make game mechanics section more high level and concise, point readers to game-mechanics doc and config file, respectively
4. Post-Phase III: Review audit doc and architecture doc for archival/consolidation. Create appropriate documents for tracking session progress and bigger picture architecture decisions/planning.
5. Feature flags cleanup from docs/audit (not blocking, can happen anytime), use cobra cli to manage flags going forward. Polish conventions (eg. use --help or -h instead of -help)

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
