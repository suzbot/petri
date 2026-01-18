# Petri Project

A simulation game inspired by Dwarf Fortress, to explore the emergent development of culture within a community.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Petri Screenshot](docs/Screenshot.png)

## Latest Updates

- **Orders System**: Press 'O' to open the Orders panel. Add Harvest orders to direct characters to gather specific item types. Characters with Harvest know-how will automatically take open orders and collect items until their inventory is full.
- **Know-how Discovery**: Characters can discover "Harvest" know-how through foraging, eating, or looking at edible items. Know-how is displayed in the Knowledge panel and cannot be transmitted through talking.
- **Inventory & Foraging**: Characters can now pick up and carry items. Foraging is a new idle activity where characters collect nearby edible items based on preference.
- **Inventory panel**: Press I to view what a character is carrying
- **Gourds**: New edible item type with colors, patterns (striped, speckled), and textures (waxy, warty). Gourds are never poisonous or healing.
- **Expanded item attributes**: More colors, patterns, and textures for existing item types
- **Save system**: Auto-saves on pause, quit, and every 60 seconds of game time. Multiple worlds supported.
- **World selection**: Continue existing worlds or start new ones from the title screen
- **Knowledge transmission**: Characters share knowledge when conversations complete
- **Talking activity**: Characters now talk with each other as an idle activity
- **Knowledge system**: Characters learn about poison/healing items through experience
- **Knowledge panel**: Press K to view what a character has learned

## Features

- **Save/Load**: Auto-saves on pause/quit, multiple worlds, continue or start new
- **Multi-character simulation** with character creation (names, food/color preferences)
- **Multi-stat survival**: hunger, thirst, energy, health, mood with urgency-based AI
- **Inventory system**: Characters can carry items, forage as an idle activity
- **Orders system**: Direct characters to harvest specific item types; characters automatically take and complete orders
- **Social behavior**: Characters talk with each other when idle
- **Knowledge system**: Characters learn facts (poison/healing) through experience and share them via talking. They also discover know-how (skills like Harvest) that cannot be shared.
- **Dynamic preferences**: Characters form opinions about items based on attributes
- **Item variety**: Berries, mushrooms, gourds (with patterns/textures), and flowers
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
- `N` - Cycle to next character
- `I` - Toggle inventory panel (select mode)
- `K` - Toggle knowledge panel (select mode)
- `O` - Toggle orders panel (+: add, c: cancel, x: expand)
- `A` / `S` - All Activity / Select mode
- `L` - Full log view (complete messages, no truncation)
- `PgUp` / `PgDn` - Scroll action log
- `ESC` - Save and return to world selection
- `Q` - Save and quit

## How It Works

1. Create characters with names and preferences, then start the simulation
2. Characters manage needs (hunger, thirst, energy, health) prioritized by urgency
3. The world contains edible items (berries, mushrooms, gourds), decorative flowers, springs for water, and leaf piles for sleep
4. When idle, characters may look at items, talk with each other, or forage (pick up items to carry)
5. Characters learn from experience: eating poison/healing items creates knowledge that affects future behavior
6. Characters form preferences based on their mood when interacting with items
7. Mood reflects emotional state, affected by need urgency and preferences

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

**Managing saves outside the game:**

```bash
# List all saved worlds
ls ~/.petri/worlds/

# Delete a specific world
rm -rf ~/.petri/worlds/world-0001

# Delete ALL save data (start fresh)
rm -rf ~/.petri

# Backup your saves
cp -r ~/.petri ~/petri-backup
```

## Project Vision

See [VISION.txt](docs/VISION.txt) for the full vision statement and development roadmap.
