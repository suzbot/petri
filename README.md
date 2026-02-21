# Petri Project

A simulation game inspired by Dwarf Fortress, with dreams to eventually model emergent development of culture within a community.

Currently, players can observe and interact with characters in a cozy forest world. Watch as they form opinions and try to stay happy while managing their basic survival needs.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Petri Screenshot](docs/Screenshot.png)

## Project Vision

See [VISION.txt](docs/VISION.txt) for the full vision statement and development roadmap.

## Latest Updates

- **Water Garden Orders**: Players can order characters to water planted tiles using carried water vessels. Characters move through the garden, watering each dry tilled planted tile and consuming one unit of water per tile.
- **Drinking from Vessels**: Characters can drink from water vessels they carry or find on the ground.
- **Fetch Water**: Idle charachters can now fill a vessel with water to keep with them
- **Plant Orders**: Players can order characters to plant on tilled soil.
- **Water Appearance + Wet Tiles**: Ponds now render as filled blocks (`▓`) instead of waves. Tiles adjacent to water (including diagonally) show "Wet" in the details panel.
- **Gourd Seeds**: Eating a gourd now drops a seed on the ground. Seeds are plantable and stack in vessels.
- **Till Soil**: Players can mark rectangular areas for tilling using an area selection tool. Characters with till know-how procure a hoe and till the marked tiles.

## How It Works

1. Create characters with names, favorite food, and favorite color, then start the simulation
2. The world contains edible plants, flowers, water sources for drinking, and leaf piles for sleep
3. Characters manage needs (hunger, thirst, energy, health) prioritized by urgency
4. Mood reflects emotional state, affected by need urgency and preferences
5. Characters form preferences based on their mood when interacting with items
6. When idle, characters may look at items, talk with each other, forage, or fill vessels with water.
7. Characters learn from experience, gaining knowledge that affects future behavior
8. Characters gain 'know-how' by making discoveries during item interactions
9. Player can issue orders (harvest, craft, garden) that characters with relevant know-how will complete

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
- `O` - Toggle orders panel (+: add, c: cancel, x: expand)
- `A` / `S` - All Activity / Select mode (x: expand to full screen)
- `PgUp` / `PgDn` - Scroll panels
- `ESC` - Close panel, or save and return to world selection
- `Q` - Save and quit

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
