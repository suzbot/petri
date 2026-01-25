# Petri Project

A simulation game inspired by Dwarf Fortress, to explore the emergent development of culture within a community.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Petri Screenshot](docs/Screenshot.png)

## Latest Updates

- **Eating from Vessels**: Hungry characters eat from vessel contents (carried or dropped). Food selection uses unified scoring: preferences, distance, and healing knowledge all factor in.
- **Vessel Contents**: Vessels hold stacks of items. Characters automatically seek out vessels when foraging or harvesting, filling them until full.
- **Crafting System**: Characters can craft vessels from gourds. Discover crafting by interacting with gourds or drinking. Order crafting via Orders panel (Craft > Vessel).
- **Orders System**: Press 'O' to add Harvest or Craft orders. Characters with know-how automatically take and complete orders.
- **Know-how Discovery**: Characters discover skills through actions. Discovery chance depends on mood (Joyful > Happy > none).

## Features

- **Save/Load**: Auto-saves on pause/quit, multiple worlds, create or delete from title screen
- **Multi-character simulation** with character creation (names, food/color preferences)
- **Multi-stat survival**: hunger, thirst, energy, health, mood with urgency-based AI
- **Inventory system**: Characters carry items or vessels, forage as idle activity, drop items when needed. Vessels hold stacks of same-variety items; characters eat from vessel contents when hungry. View with I key
- **Crafting system**: Characters craft vessels from gourds, inheriting appearance. Crafted items have display names
- **Orders system**: Direct characters to harvest or craft via Orders panel (O key)
- **Social behavior**: Characters talk with each other when idle, transmitting knowledge
- **Knowledge system**: Learn facts (poison/healing) through experience, discover know-how (skills) through actions. Facts transmit via talking; know-how does not. View with K key
- **Dynamic preferences**: Characters form opinions about items based on mood and attributes
- **Item variety**: Berries, mushrooms, gourds (with patterns/textures), flowers, and crafted vessels
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

## How It Works

1. Create characters with names and preferences, then start the simulation
2. Characters manage needs (hunger, thirst, energy, health) prioritized by urgency
3. The world contains edible items (berries, mushrooms, gourds), decorative flowers, springs for water, and leaf piles for sleep
4. When idle, characters may look at items, talk with each other, or forage (pick up items to carry)
5. Characters learn from experience: eating poison/healing items creates knowledge that affects future behavior
6. Characters discover crafting know-how through interacting with gourds or drinking at springs
7. Player can issue orders (harvest, craft) that characters with relevant know-how will complete
8. Characters form preferences based on their mood when interacting with items
9. Mood reflects emotional state, affected by need urgency and preferences

For detailed mechanics, see [docs/game-mechanics.md](docs/game-mechanics.md). For configuration values, see `internal/config/config.go`.

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
