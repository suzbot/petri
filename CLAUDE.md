# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Petri is a Dwarf Fortress-inspired simulation exploring emergent culture development. Built with Go and Bubble Tea for flicker-free terminal rendering.

**Current features:** Multi-character mode, character creation screen, multi-stat survival (hunger, thirst, energy, health, mood), dynamic preference system, urgency-based AI with stat fallback and frustration mechanics, landscape features (springs, leaf piles), sleep mechanics, view modes, and action logging.

**Vision:** History exists only in character memories and created artifacts. As characters die, their knowledge dies with them except what they've communicated or created. See [docs/VISION.txt](docs/VISION.txt).

---

## Quick Reference

### Commands

```bash
# Build and run
go build -o petri ./cmd/petri
./petri

# Test flags
./petri -no-food -no-water -no-beds -debug -help

# Testing
go test ./...
go test -race ./...
```

### Human Testing Protocol

**CRITICAL**: Always rebuild before testing:
```bash
go build -o petri ./cmd/petri && echo "✓ Binary rebuilt at $(date)"
```

---

## Architecture

Go project using Bubble Tea's Model-View-Update (MVU) pattern.

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

### Key Design Patterns

- **MVU Architecture**: Bubble Tea handles rendering diffs automatically
- **Intent System**: Characters calculate Intent, then intents applied atomically (enables future parallelization)
- **Multi-Stat Urgency**: Tiers 0-4, highest wins, tie-breaker: Thirst > Hunger > Energy
- **Stat Fallback**: If intent can't be fulfilled, falls through to next urgent stat
- **Sparse Grid + Indexed Slices**: O(1) character lookups, separate slices for characters/items/features

### Data Flow

1. `Update()` → `updateGame()`
2. `UpdateSurvival()`: timers, stat changes, damage, sleep/wake
3. `CalculateIntent()`: evaluate tiers, try each in priority, track failures
4. `applyIntent()`: accumulate speed/progress, execute actions
5. `View()`: render UI (Bubble Tea diffs automatically)

### Memory Model

Per BUILD CONCEPT in VISION.txt:

- **ActionLog (Working Memory)**: Per-character recent events, bounded by count
- **Memory (Long-Term)**: Future - selective storage of notable events, persists until death
- No omniscient log - player sees aggregate of character experiences

---

## Development

### Status

- ✅ Phase II complete (Features A, B, C, D)
- ✅ All test priorities complete (P1-P4)
- ✅ Code audit complete
- ✅ High priority audit items: typed constants (Color, StatType, Edible)
- ✅ Test strategy assessment: streamlined process, archived Gherkin docs
- ✅ Tier/level calculation logic consolidated (StatThresholds, StatLevels structs)
- 🔄 **Phase III in progress**: Mood and preference formation

### Current Work

**Phase III Implementation Plan**: See [docs/phase03-plan.md](docs/phase03-plan.md)

Sub-phases:
- ✅ A. Mood Stat Foundation
- ✅ B. Preference System Refactor
- ✅ C. Preference Formation
- 🔄 **D. World Balancing** ← in progress

**D1 (Healing items) decisions:**
- Instant heal (not over time) for simplicity
- HealAmount = 20 (matches FoodHungerReduction, tunable)
- HealingConfig mirrors PoisonConfig pattern
- 1-2 healing combos, must not overlap with poison
- Debug visibility: Log combos at world creation (avoids UI coupling)

### Next Priorities

1. Complete Phase III sub-phases (A → B → C → D)
2. Balance tuning pass (end of Phase III)
3. Post-Phase III: Review audit doc and architecture doc for archival/consolidation
4. Feature flags cleanup (not blocking, can happen anytime)

### Deferred Enhancements & Trigger Points

Items analyzed and consciously deferred until trigger conditions are met.

| Enhancement | Triggers (implement when ANY is met) |
|-------------|--------------------------------------|
| **Parallel Intent Calculation** | Character count ≥ 16; Intent calc exceeds ~1ms/char; Tick time exceeds 50ms |
| **EventType for ActionLog** | Event filtering in UI; Event-driven behavior (Phase III); Event persistence |
| **Category field for Items** | Non-food items introduced (D3 Flowers will trigger this) |
| **Performance optimizations** | Noticeable lag; Profiling shows bottlenecks |
| **Preference formation for beverages** | Other beverages (non-spring) introduced |
| **Depression and Rage mechanics** | Job acceptance logic implemented |
| **Testify test package** | Test complexity warrants it; Current assertion patterns become unwieldy |

**Implementation details:** See docs/architecture-review-typed-constants.md

### Future Enhancements

**Balance tuning candidates** (see config.go for current values):
- Activity durations
- Satisfaction cooldown
- Sleep wake thresholds
- Movement energy drain
- Health tier thresholds
- Preference formation chances (inflated for testing)
- Mood adjustment rates and modifiers
- Spawn rate vs consumption rate: 4 characters eating shouldn't outpace plant respawning (adjust spawn interval/chance or hunger rate)
- Flower overpopulation: Flowers reproduce but aren't consumed, may outpace edible items over time. Consider: slower flower reproduction rate, OR allowing flower consumption in severe hunger situations

**Performance optimizations:**
- Skip wake checks for sleeping characters unless tier changed
- Dead character filtering at map level
- Cache nearest features if map static

**Parallel Intent pattern:**
```go
var wg sync.WaitGroup
for _, char := range chars {
    wg.Add(1)
    go func(c *entity.Character) {
        defer wg.Done()
        c.Intent = system.CalculateIntent(c, items, gameMap, log)
    }(char)
}
wg.Wait()
```

---

## Testing

### Process

1. **Clarify functional intent** - For complex features, discuss expected behavior before coding. Use Gherkin-style scenarios when helpful, but not required as formal documents.
2. **Write tests first** - Define expected behavior in test form.
3. **Implement feature** - Write minimum code to pass tests.
4. **Refactor** - Clean up while keeping tests green.
5. **Regression tests for bugs** - When fixing bugs, add a regression test to prevent recurrence.

### What to Test

- **Core logic**: Tier calculations, intent priority, survival mechanics, stat changes
- **Behavior correctness**: System interactions, state transitions
- **Invariants**: Multi-tick simulation tests (no position conflicts, consistent state)
- **Regression**: Specific bugs that were fixed

### What NOT to Test

- **UI rendering** (view.go) - Visual verification done manually
- **Bubble Tea integration** - Framework responsibility
- **Exact log message wording** - Brittle; test that events ARE logged, not exact text
- **Configuration constants** - Just values, no logic

### Test Files

```
internal/
  entity/*_test.go       # Character, Item, Feature, Entity (98.9% coverage)
  game/map_test.go       # Position tracking, collision (43% - spawning untested, ok)
  system/*_test.go       # Survival, movement, consumption, action log (72.8%)
  simulation/*_test.go   # Multi-tick invariants (93.6%)
  ui/creation_test.go    # Character creation state logic
```

### Running Tests

```bash
go test ./...              # All tests
go test -race ./...        # With race detector (important for actionlog)
go test -v ./internal/system/...  # Verbose, specific package
go test -run TestHungerTier ./... # Specific test
go test -cover ./...       # Coverage report
```

---

## Collaboration

### Working Patterns

- **Verify against requirements**: Before implementing a feature, re-read the relevant requirements to ensure correct interpretation. Don't rely on memory or plan summaries alone.
- **Confirm rather than guess**: Ask clarifying questions instead of making assumptions about intent.
- **Compare options before committing**: Present alternatives with trade-offs; confirm approach before implementing.
- **Number discussion points**: When presenting multiple thoughts/options for discussion, number them so the user can easily reference specific points in their reply.
- **Pause for architecture review**: When implementing reveals broader design questions, pause to review architecture before proceeding. Document decisions in docs/ to maintain context.
- **Test-driven development**: See Testing section for process. Clarify intent verbally when needed; formal Gherkin docs not required.
- **Iterative development**: Small features with frequent human testing checkpoints.
- **Discuss questions in context**: When starting a new feature area, ask clarifying questions as each sub-feature is approached rather than presenting a large batch of questions upfront. Easier to answer in context.
- **Rebuild before testing**: Always remind to rebuild binary (see Human Testing Protocol).
- **Keep docs current**: Include README and reference doc updates as todo items when implementing features. Don't defer documentation to end of session.
- **Maintain context documents**: Reference CLAUDE.md and docs/ regularly at the start of work sessions. Update documents and keep them clean and organized as we go. Move detailed content to separate docs/ files when CLAUDE.md becomes cluttered. This maintains project knowledge across sessions.

### Reference Documents

| Document | Purpose |
|----------|---------|
| [docs/VISION.txt](docs/VISION.txt) | Project vision and phases |
| [docs/phase03reqs.txt](docs/phase03reqs.txt) | Phase III requirements |
| [docs/phase03-plan.md](docs/phase03-plan.md) | Phase III implementation plan (current) |
| [docs/game-mechanics.md](docs/game-mechanics.md) | Detailed stat thresholds, rates, systems |
| [docs/audit-findings.md](docs/audit-findings.md) | Code audit results |
| [docs/architecture-review-typed-constants.md](docs/architecture-review-typed-constants.md) | Typed constants decisions |
| [docs/bug-fixes.md](docs/bug-fixes.md) | Bug documentation and regression tests |
| [docs/failed-approaches.md](docs/failed-approaches.md) | Approaches tried and abandoned |

**Archived:** `tests/archive/` contains historical testing approach and Gherkin acceptance criteria files from initial test implementation.
