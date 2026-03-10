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

### Codebase Navigation

Architecture follows the Character's Cognitive Loop — see [docs/architecture.md](docs/architecture.md) for detailed file roles, patterns, and "adding new X" checklists. Key entry points:

- **Decision-making**: `internal/system/intent.go` (needs) → `order_execution.go` (orders) → `helping.go` → `discretionary.go`
- **Execution**: `internal/ui/apply_actions.go` (applyIntent dispatch + all action handlers)
- **Entities**: `internal/entity/` (character, item, construct, activity/recipe registries)
- **Game loop**: `internal/ui/update.go` (tick orchestration) → `model.go` (state)

## Collaboration

**Pause -> Skill -> Discussion → Tests -> Implementation → Human Testing → Documentation**

- **Before writing code**: Load the relevant skill, check architecture.md, discuss approach. Use project skills (`/refine-feature`, `/implement-feature`, `/new-phase`, `/retro`) — not generic plan mode. Phase design docs and step specs live in `docs/`.
- **Frame the problem first**: Before solving, confirm the problem — is this actually a problem? Is it worth solving now? Specify current state and desired state in functional terms. Consider impact on the larger system. Confirm alignment before implementing. When human testing surfaces a new gap mid-implementation, that's a new problem to scope and discuss, not a defect to fix in place.
- **Communication**: Functional terms, not code mechanics. Prose for tradeoffs, not multiple-choice. Recommend with options. Qualify claims precisely, citing specific evidence — assertions aren't demonstrations.
- **Interaction**: The user has a vision — help realize it. If you don't understand the intent, ask for context. Don't ask "are you sure?" — if there are substantive concerns, present trade-offs. Trust user observations as evidence; verify where they point, don't reason about why they can't be true.
- **When things go wrong**: Gather evidence first — examine the save file the user references, add logging, ask what they observe — before reasoning about causes. Don't speculate. Second bug in the same feature = step back and restate the intended flow before patching further. Surface when stuck.
- **Quality gates**: TDD. User must test before marking complete. Keep docs current.

## Testing

- Add regression tests when making bug fixes
- No tests needed for UI rendering, Bubble Tea integration, brittle string matching (log wording, display names, UI text), configuration constants
- Don't test for absence unless requirements call for prevention
- Headless simulation tests for measuring game balance are located in `internal/simulation/observation_test.go`.

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
