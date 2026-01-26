# Petri Project

A simulation game inspired by Dwarf Fortress, as a project to explore the emergent development of culture within a community.

Currently, players can observe and interact with a cozy forest world. Watch as characters form opinions about the things around them, and as they try to stay happy while managing their basic survival needs.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Petri Screenshot](docs/Screenshot.png)

## Latest Updates

- **Eating from Vessels**: Hungry characters eat from vessel contents (carried or dropped). Food selection uses unified scoring: preferences, distance, and healing knowledge all factor in.
- **Vessel Contents**: Vessels hold stacks of items. Characters automatically seek out vessels when foraging or harvesting, filling them until full.
- **Crafting System**: Characters can craft vessels from gourds. Discover crafting by interacting with gourds or drinking. Order crafting via Orders panel (Craft > Vessel).

## How It Works

1. Create characters with names and preferences, then start the simulation
2. The world contains edible items (berries, mushrooms, gourds), decorative flowers, springs for water, and leaf piles for sleep
3. Characters manage needs (hunger, thirst, energy, health) prioritized by urgency
4. Mood reflects emotional state, affected by need urgency and preferences
5. Characters form preferences based on their mood when interacting with items
6. When idle, characters may look at items, talk with each other, or forage (pick up items to carry)
7. Characters learn from experience: eating poison/healing items creates knowledge that affects future behavior
8. Characters discover crafting know-how through interacting with gourds or drinking at springs
9. Player can issue orders (harvest, craft) that characters with relevant know-how will complete

For detailed mechanics, see [docs/game-mechanics.md](docs/game-mechanics.md). For configuration values, see `internal/config/config.go`.

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
- Arrow keys - Move cursor
- `F` - Follow/unfollow character
- `N` / `B` - Cycle next/previous character
- `I` - Toggle inventory panel (select mode)
- `K` - Toggle knowledge panel (select mode)
- `O` - Toggle orders panel (+: add, c: cancel, x: expand)
- `A` / `S` - All Activity / Select mode (x: expand to full screen)
- `PgUp` / `PgDn` - Scroll action log
- `ESC` - Save and return to world selection
- `Q` - Save and quit


## Debug Mode

```bash
./petri -debug           # Show detailed numeric info
./petri -no-food         # No food items spawned
./petri -no-water        # No springs spawned
./petri -no-beds         # No leaf piles spawned
./petri -mushrooms-only  # Replace all items with mushroom varieties
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

## Project Vision

See [VISION.txt](docs/VISION.txt) for the full vision statement and development roadmap.
