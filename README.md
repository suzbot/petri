# Petri Project

A simulation game inspired by Dwarf Fortress, exploring the emergent development of culture within a community.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea) for flicker-free terminal rendering.

![Petri Screenshot](docs/Screenshot.png)

## Latest Updates

- **Item varieties**: Mushrooms now have pattern (spotted) and texture (slimy) attributes
- **Dynamic preference formation**: Characters form likes/dislikes based on mood when eating or looking at items
- **Mood system**: Emotional state affected by need urgency, status effects, and preferences
- **Healing items**: Some food restores health when consumed

## Features

- **Multi-character simulation** with character creation (names, food/color preferences)
- **Multi-stat survival**: hunger, thirst, energy, health, mood with urgency-based AI
- **Dynamic preferences**: Characters form opinions about items based on attributes
- **Item variety**: Berries, mushrooms (with pattern/texture), and flowers
- **World dynamics**: Item spawning, springs, leaf piles, poison and healing effects
- **View modes**: Select mode (examine entities) and All Activity mode (combined log)

## Running the Game

```bash
go build ./cmd/petri
./petri
```

Or directly:

```bash
go run ./cmd/petri
```

## Controls

**Start Screen:**
- `M` - Start game (opens character creation)

**Character Creation:**
- `← →` - Navigate between characters
- `↑ ↓` / `Tab` - Navigate fields
- `Space` - Cycle option
- `Ctrl+R` - Randomize all
- `Enter` - Start game

**During Game:**
- `Space` - Pause/unpause (world starts paused)
- `.` - Step forward one tick (while paused)
- Arrow keys - Move cursor
- `F` - Follow/unfollow character
- `N` - Cycle to next character
- `A` / `S` - All Activity / Select mode
- `PgUp` / `PgDn` - Scroll action log
- `ESC` or `Q` - Quit

## How It Works

1. Create characters with names and preferences, then start the simulation
2. Characters manage needs (hunger, thirst, energy) prioritized by urgency
3. The world contains edible items (berries, mushrooms), decorative flowers, springs for water, and leaf piles for sleep
4. Characters form preferences based on their mood when interacting with items
5. Mood reflects emotional state, affected by need urgency and preferences

For detailed mechanics, see [docs/game-mechanics.md](docs/game-mechanics.md). For configuration values, see `internal/config/config.go`.

## Debug Mode

```bash
./petri -debug        # Show detailed numeric info
./petri -no-food      # No food items spawned
./petri -no-water     # No springs spawned
./petri -no-beds      # No leaf piles spawned
./petri -help         # Show all available flags
```

Debug mode reveals exact stat values, action progress timers, and poison/healing information.

## Project Vision

See [VISION.txt](docs/VISION.txt) for the full vision statement and development roadmap.
