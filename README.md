# Petri Project

A simulation game inspired by Dwarf Fortress, exploring the emergent development of culture within a community.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea) for flicker-free terminal rendering.

## Current Features

- **Multi-character mode**: Run simulation with multiple customizable characters
- **Character creation screen**: Set names and preferences (food type, color) for each character
- **Dynamic preference system**: Characters like/dislike items based on attributes
- Single-character mode with customizable preferences (debug mode only)
- Multi-stat survival system: hunger, thirst, energy, health, mood
- Urgency-based priority system with stat fallback
- Tier-based stat coloring: dark green (optimal), yellow (severe), red (crisis)
- Landscape features: springs (water), leaf piles (beds)
- Sleep mechanics with early wake on urgent needs
- Action duration system (eating, drinking, falling asleep take time)
- Satisfaction cooldown (stats pause briefly after reaching optimal)
- Poison effects with speed penalties
- **View modes**: Select mode (examine entities) and All Activity mode (combined log)
- Action logging with scrollable history
- Cursor navigation and character following
- World starts paused for observation

## Requirements

- Go 1.21+

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
- `1` - Single character mode (debug mode only)

**Character Creation:**

- `← →` - Navigate between characters
- `↑ ↓` - Navigate between fields
- `Tab` - Next field
- `Space` - Cycle option (Food/Color fields)
- `Ctrl+R` - Randomize all characters
- `Enter` - Start game
- `Esc` - Back to start screen

**Single Character Setup (debug mode):**

- `B` / `M` - Select berries or mushrooms as favorite food
- `R` / `L` / `W` / `N` - Select red, blue, white, or brown as favorite color

**During game:**

- `Space` - Pause/unpause (world starts paused)
- `.` - Step forward one tick (while paused)
- Arrow keys - Move cursor
- `F` - Follow/unfollow character at cursor
- `N` - Cycle to next alive character
- `A` - Switch to All Activity mode (combined log from all characters)
- `S` - Switch to Select mode (details panel + per-character log)
- `PgUp` / `PgDn` - Scroll action log
- `ESC` or `Q` - Quit

## Test & Debug Mode

Command-line flags for testing and debugging:

```bash
./petri -no-food      # No food items spawned
./petri -no-water     # No springs spawned
./petri -no-beds      # No leaf piles spawned
./petri -debug        # Show detailed numeric info
./petri -no-food -no-water  # Combine multiple flags
./petri -help         # Show all available flags
```

**Test flags:** Useful for testing stat fallback behavior (e.g., when hunger can't be satisfied, character pursues next urgent need)

**Debug flag reveals:**

- Position coordinates for all entities
- Full stat values (e.g., "Hunger: 45/100 (Hungry)", "Mood: 50/100 (Neutral)")
- Speed value
- Action progress timers (e.g., "Drinking (0.8s)")
- Stat change numbers in action log (e.g., "hunger 50→30")
- Duration values in action log
- Poison combo information

## How It Works

1. Press M to enter character creation - customize names, food preferences, and colors
2. Press Enter to start the game - world begins paused, press Space to begin
3. Each character manages three needs: hunger, thirst, and energy
   - Thirst increases faster than hunger
   - After reaching optimal levels, stats briefly pause before changing again
4. Needs are prioritized by urgency tier, with tie-breaker order: Thirst > Hunger > Energy
5. If a need can't be fulfilled, character falls back to next most urgent need
6. Actions take time to complete:
   - Eating, drinking, and falling asleep all have duration
   - At springs, character drinks continuously until fully satisfied
7. Preferences affect food selection:
   - Characters prefer items matching their likes
   - When moderately hungry: prefers best matches, settles for partial
   - When very hungry: accepts any liked item
   - When ravenous: eats anything
8. Sleep mechanics:
   - Character sleeps in leaf pile (wakes fully rested) or on ground (wakes partially rested)
   - Wakes early if hunger/thirst becomes more urgent than current energy tier
9. In multi-character mode, characters compete for resources:
   - Springs and beds become occupied when in use
   - Characters find alternative targets when blocked
10. Use view modes to observe:
    - Select mode (S): Examine individual entities, view per-character logs
    - All Activity mode (A): See combined activity from all characters
11. Character dies if health reaches 0 (from starvation, dehydration, or poison)
12. Mood reflects emotional state based on needs:
    - Increases slowly when all needs are met
    - Decreases when needs become urgent (faster at higher urgency)
    - Receives a boost when a need is fully satisfied (hunger/thirst→0, energy→100)

## Stat Levels

| Urgency  | Hunger      | Thirst        | Energy       | Health    | Mood      |
| -------- | ----------- | ------------- | ------------ | --------- | --------- |
| None     | Not Hungry  | Hydrated      | Rested       | Healthy   | Joyful    |
| Mild     | Hungry      | Thirsty       | Tired        | Poor      | Happy     |
| Moderate | Very Hungry | Very Thirsty  | Very Tired   | Very Poor | Neutral   |
| Severe   | Ravenous    | Parched       | Exhausted 🛏️ | Critical  | Unhappy   |
| Crisis   | Starving 💔 | Dehydrated 💔 | Collapsed ⚡ | Dying     | Miserable |

💔 Takes health damage
🛏️ Voluntary ground sleep available
⚡ Involuntary collapse (immediate)

**Stat colors:** Stats at optimal level (None) display in dark green, Severe in yellow, Crisis in red.

## Project Vision

See [VISION.txt](docs/VISION.txt) for the full vision statement and development roadmap.
