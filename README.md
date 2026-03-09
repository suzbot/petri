# Petri Project

A simulation game inspired by Dwarf Fortress, with dreams to eventually model emergent development of culture within a community.

Currently, players can observe and interact with characters in a cozy forest world. Watch as they form opinions and try to stay happy while managing their basic survival needs.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Petri Screenshot](docs/Screenshot.png)

## Project Vision

Characters don't follow scripts — they will form preferences, learn from experience, share knowledge,
and make decisions based on personality and memory. History exists only in what characters remember and create.
When a character dies, their knowledge dies with them, except what they've taught or built.

See [VISION.txt](docs/VISION.txt) for the full vision and development roadmap.

## Development Process

Petri is built by a product manager and an AI (Claude).
The development process itself is part of the project —
each practice here was designed and iterated on during development.

Core process innovations developed over the course of this project:

- **Skill-driven workflows**: Custom AI skill definitions support agile iterative development, TDD, human testing checkpoints, and requirement traceability at each step
- **Recursive retrospectives**: After each feature, an automated retro reviews collaboration friction, then updates its own skill definitions and notes design values, compounding institutional knowledge across sessions
- **Inferred design values**: Patterns observed during retros are captured as explicit principles that guide future implementation decisions

**See it in action:**

- [`.claude/skills/`](.claude/skills/) — AI interaction protocols developed for this project
- [`docs/Values.md`](docs/Values.md) — design values surfaced through retrospectives
- [`CLAUDE.md`](CLAUDE.md) — codebase context and collaboration norms

## Latest Updates

- **Construction**: Characters can learn to build fences and huts from available materials.
- **Construct preferences**: Characters look at constructs and form opinions about materials and recipes, affecting mood.
- **Craft Bricks orders**: Characters shape loose clay into bricks, repeating until all dug clay is used.
- **Dig Clay orders**: Direct characters to dig clay from deposits.
- **Seed extraction**: Characters extract seeds from flowers and tall grass via Extract orders, without harming the plant.
- **Helping:** Characters bring food or water to other character in crisis
- **Gardening:** Characters can gather seeds, till soil, water the garden, and watch food grow

## How It Works

1. Choose R to start with random characters, or C to customize characters before starting
2. The world contains edible plants, flowers, water sources for drinking, and leaf piles for sleep
3. Characters manage needs (hunger, thirst, energy, health) prioritized by urgency
4. Mood reflects emotional state, affected by need urgency and preferences
5. Characters form preferences based on their mood when interacting with items
6. When idle, characters may look at items, talk with each other, forage, or fill vessels with water.
7. Characters learn from experience, gaining knowledge that affects future behavior
8. Characters gain 'know-how' by making discoveries during item interactions
9. Player can issue orders (harvest, gather, craft, garden) that characters with relevant know-how will complete

For detailed mechanics, see [docs/game-mechanics.md](docs/game-mechanics.md).

## Requirements

- Go version go1.25.5 or higher (https://go.dev/learn/)

## Running the Game

```bash
go build ./cmd/petri
./petri
```

## Controls

**World Selection:**

- `↑ ↓` - Select world
- `Enter` - Continue selected world or start new
- `Q` - Quit

**Character Creation:**

- `← →` - Navigate between characters
- `↑ ↓` / `Tab` - Navigate fields
- `Space` - Cycle option
- `Ctrl+R` - Randomize all
- `Enter` - Start game

**During Game:**

- `Space` - Pause/unpause (saves on pause)
- `.` - Step forward one tick (while paused)
- `<` / `>` - Slow down / speed up (½x, ¼x)
- Arrow keys - Move cursor
- `F` - Follow/unfollow character
- `N` / `B` - Cycle next/previous character
- `E` - Edit character name (select mode)
- `P` - Toggle preferences panel (select mode)
- `I` - Toggle inventory panel (select mode)
- `K` - Toggle knowledge panel (select mode)
- `L` - Return to action log from details subpanel (select mode)
- `O` - Toggle orders panel (+: add, c: cancel, x: expand)
- `A` / `S` - All Activity / Select mode (x: expand to full screen)
- `PgUp` / `PgDn` - Scroll panels
- `ESC` - Go back one level (collapse expanded view → close subpanel → close orders → return to all-activity)
- `Q` - Save and return to world selection

## Debug Mode

```bash
./petri -debug           # Show detailed numeric info
./petri -help            # Show all available flags
```

Debug mode reveals exact stat values, action progress timers, and poison/healing information.

## Save Files

Save data is stored in `~/.petri/worlds/`. Each world has its own directory:

```
~/.petri/
  worlds/
    world-0001/
      state.json      # Current game state
      state.backup    # Previous save (backup)
      meta.json       # World name, character count, last played
    world-0002/
      ...
```

**Managing saves:**

Worlds can be deleted from the title screen by pressing `D` on a saved world (with confirmation).

For advanced management via terminal:

```bash
# List all saved worlds
ls ~/.petri/worlds/

# Delete ALL save data (start fresh)
rm -rf ~/.petri

# Backup your saves
cp -r ~/.petri ~/petri-backup
```

## Customization

**Character Names:** Edit `internal/entity/names.go` to add or remove names from the random name pool.

**Configuration Values:** see `internal/config/config.go`.
