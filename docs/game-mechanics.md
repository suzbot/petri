# Game Mechanics Reference

Detailed game mechanics. For exact values, see `internal/config/config.go`.

## Stat Thresholds

Stats have four severity tiers: Mild, Moderate, Severe, Crisis. Thresholds defined in `internal/entity/character.go`.

| Stat   | Direction | Notes |
| ------ | --------- | ----- |
| Hunger | Higher = worse | 0 is optimal, 100 is starving |
| Thirst | Higher = worse | 0 is optimal, 100 is dehydrated |
| Energy | Lower = worse | 100 is optimal, 0 triggers collapse |
| Health | Lower = worse | 100 is optimal, 0 is death |
| Mood   | Lower = worse | 100 is Joyful, 0 is Miserable |

## Stat Rates

Stats change over time at rates defined in config. Key behaviors:
- Hunger and Thirst increase continuously
- Energy decreases when awake, plus additional drain per movement step
- Health decreases from starvation, dehydration, or poison damage

## Frustration System

When a character cannot fulfill urgent needs (Severe+), they accumulate failed intent counts. After reaching the threshold, they become Frustrated for a duration, during which they skip intent calculation and display "?" symbol.

Flow:
1. `CalculateIntent` returns nil when no stat can be fulfilled
2. Only if maxTier >= Severe: increment failed count
3. If count >= threshold: set Frustrated with timer
4. While frustrated: skip intent, show "?" (orange), status "FRUSTRATED"
5. Timer decrements; when expired: clear frustration, log "Calmed down"

## Intent Re-evaluation Guards

Re-evaluation only triggers when a higher-tier stat can actually be fulfilled. Prevents thrashing when e.g., energy tier > thirst tier but no beds exist.

## Sleep Mechanics

- Wake at full energy (bed) or partial energy (ground)
- Early wake only if another stat is Moderate+ tier AND has worse raw urgency than energy
- Ground sleep available at Exhausted tier
- Collapse (immediate, involuntary) at 0 energy

## Satisfaction Cooldown

Stats hang at optimal for a cooldown period before starting natural change. This gives a "freshly satisfied" feel.

## Speed System

Base speed modified by penalties that stack:
- Poison penalty
- High thirst penalties (tiered)
- Exhaustion penalties (tiered)
- Minimum speed floor prevents complete immobility

Speed accumulator gates movement; higher speed = more moves per tick.

## Action Duration

Drinking, eating, and falling asleep have a duration before completing. Collapse at Energy=0 is immediate (involuntary).

## Continuous Drinking

At springs, characters drink until thirst == 0 (not just until tier boundary).

## Mood System

Mood reflects character emotional state (0-100, higher is better).

Levels: Joyful, Happy, Neutral, Unhappy, Miserable (thresholds in character.go)

### Mood Changes from Need States

Mood changes based on the highest need tier (Hunger, Thirst, Energy, Health):
- All optimal: mood increases slowly
- Mild: no change
- Moderate: mood decreases slowly
- Severe: mood decreases at medium rate
- Crisis: mood decreases quickly

### Mood Penalties from Status Effects

Status effects apply additional mood penalties (additive with need-based decay):
- **Poisoned**: mood penalty per second
- **Frustrated**: mood penalty per second

These stack with each other and with need-based decay.

### Mood Boost on Need Fulfillment

When a need is **fully satisfied**, mood receives a boost:
- Hunger reaches 0 (from eating)
- Thirst reaches 0 (from drinking)
- Energy reaches 100 (from sleeping in bed)
- Health reaches 100 (from healing items)

### Mood Display

- Mood tier transitions logged in Action Log (e.g., "Feeling Joyful")
- Log colors: Joyful (dark green), Unhappy (yellow), Miserable (red)
- Details Panel shows mood with tier-based coloring

## Item Varieties

Items are generated from varieties at world creation. Each variety defines a unique combination of attributes.

### Item Types

- **Berries**: Color only (red, blue)
- **Mushrooms**: Color + optional Pattern + optional Texture
  - Colors: brown, white, red
  - Pattern: spotted or none
  - Texture: slimy or none
- **Flowers**: Color only (red, orange, yellow, blue, purple, white) - non-edible

### Variety Generation

At world creation:
1. Generate varieties for each item type
2. Variety count = max(2, spawnCount / VarietyDivisor)
3. 20% of edible varieties marked poisonous
4. 20% of edible varieties marked healing (mutually exclusive with poison)

### Item Lifecycle

Items have spawn and death timers managed by the lifecycle system (`internal/system/lifecycle.go`).

**Spawning:**
- Each item has a spawn timer
- When timer expires, chance to spawn adjacent copy with same attributes
- Children inherit all parent attributes (color, pattern, texture, poison, healing)
- Spawn interval configured per item type in `config.ItemLifecycle`

**Death:**
- Items with a death interval have a death timer; when it expires, the item is removed
- Items with death interval of 0 are immortal (removed only when consumed)
- Currently only flowers have death timers; edibles are immortal until eaten
- Death creates natural population equilibrium for decorative items

See `config.ItemLifecycle` for per-item-type spawn and death intervals.

## Preference System

Characters have dynamic preferences that affect food selection and mood.

### Preference Structure

Each preference targets item attributes:
- **ItemType only**: e.g., "Likes berries" (matches any berry)
- **Color only**: e.g., "Likes red" (matches any red item)
- **Pattern only**: e.g., "Likes Spots" (matches any spotted mushroom) - uses noun form
- **Texture only**: e.g., "Likes Slime" (matches any slimy mushroom) - uses noun form
- **Combo (2-3 attributes)**: e.g., "Likes spotted brown mushrooms" - always includes ItemType

Each preference has a **valence**: +1 (likes) or -1 (dislikes).

### NetPreference Calculation

When evaluating an item, sum all matching preference scores:
- Single-attribute preference: contributes `Valence × 1`
- Combo preference (2 attributes): contributes `Valence × 2`

Examples:
| Character Preferences | Item | NetPreference |
|-----------------------|------|---------------|
| Likes berries, Likes red | Red berry | +2 (perfect) |
| Likes berries, Likes red | Blue berry | +1 (partial) |
| Likes red berries (combo) | Red berry | +2 (perfect) |
| Likes red berries (combo) | Blue berry | 0 (no match) |
| Likes red, Likes berries, Likes red berries | Red berry | +4 (all stack) |
| Dislikes slimy, Dislikes mushrooms, Dislikes slimy mushrooms | Slimy mushroom | -4 (all stack) |

### Food Selection by Hunger Tier

Uses gradient scoring: `Score = (NetPreference × PrefWeight) - (Distance × DistWeight)`

Higher hunger = lower preference weight + willingness to eat disliked items:
- **Moderate (50-74)**: High PrefWeight, only considers NetPreference >= 0 items (filters disliked)
- **Severe (75-89)**: Medium PrefWeight, considers all items including disliked
- **Crisis (90+)**: No PrefWeight (nearest wins), considers all items

Weights configured in config.go. When scores are equal, closer item wins (distance tiebreaker).

### Initial Preferences

Characters start with two positive preferences based on character creation:
- Likes [selected food type]
- Likes [selected color]

### Preference Formation

Preferences form dynamically when consuming or looking at items, based on current mood:
- Joyful/Happy: chance to form positive preference (likes)
- Neutral: no formation
- Unhappy/Miserable: chance to form negative preference (dislikes)

Formation types (weights configured in config.go):
- **Solo**: Single attribute (ItemType, Color, Pattern, or Texture)
  - Pattern/Texture solo use noun forms: "Likes Spots", "Likes Slime"
- **Combo**: ItemType + 1-2 other attributes (max 3 total)
  - Combos always include ItemType: "spotted mushrooms", "slimy red mushrooms"
  - Uses adjective forms in combos: "spotted", "slimy"

If character already has exact same preference:
- Same valence: No change
- Opposite valence: Removes existing preference

### Preference Mood Impact

When consuming food, mood adjusts based on NetPreference (scaled by config modifier).

### Preference Log Messages

**Formation:**
- "New Opinion: Likes [x]" → dark green
- "New Opinion: Dislikes [x]" → yellow
- "No longer likes/dislikes [x]" → light blue

**Mood impact (debug mode only):**
- "Eating [item] Improved Mood (mood X→Y)"
- "Eating [item] Worsened Mood (mood X→Y)"

## Knowledge System

Characters learn about items through experience. Knowledge persists and affects future behavior.

### Learning by Experience

When a character eats a poisonous or healing item, they gain knowledge about that specific variety:
- **Poisonous item**: Learns "[Variety] are poisonous" (e.g., "Spotted red mushrooms are poisonous")
- **Healing item**: Learns "[Variety] are healing" (e.g., "Blue berries are healing") - only if health actually increased

Knowledge is only gained once per variety - eating the same type again does not create duplicate entries.

When knowledge is gained, "Learned something!" appears in the action log (darker blue color).

### Knowledge Panel

Press `K` in select mode to toggle the Knowledge panel (replaces action log). Shows all knowledge the selected character has learned. Press `K` again to return to action log.

### Knowledge Affects Behavior

**Poison Knowledge → Dislike Preference:**
When a character learns an item is poisonous, they automatically form a dislike preference for the full variety (e.g., "Dislikes spotted brown mushrooms"). This affects food selection:
- At Moderate hunger: character avoids the disliked item entirely
- At Severe hunger: disliked items score lower but may still be eaten
- At Crisis hunger: nearest food is eaten regardless of preference

If the character had an exact matching "like" preference, it is removed instead of creating a new dislike.

### Healing Knowledge → Health Seeking

When a character knows an item is healing AND their health is below full:
- Health becomes a need that can drive intent (priority: Thirst > Hunger > Health > Energy)
- Character seeks the nearest known healing item when health is the most urgent stat
- If no known healing items exist, health cannot drive intent (character won't seek healing without knowledge)

### Healing Knowledge → Food Selection Bonus

When seeking food while hurt (health below full), known healing items receive a bonus to their gradient score:
- **Mild (health ≤75)**: +5 bonus
- **Moderate (health ≤50)**: +10 bonus
- **Severe (health ≤25)**: +20 bonus
- **Crisis (health ≤10)**: +40 bonus

This makes injured characters prefer known healing food over equally-distant alternatives, with stronger preference at lower health.

### Future Behaviors (Planned)

- Knowledge can be shared between characters through conversation (Phase 4H)

## Idle Activities

When characters have no urgent needs (all stats below Moderate tier), they select from idle activities:

### Activity Selection

Every 10 seconds (IdleCooldown), characters roll for an idle activity:
- **1/3 chance**: Look at nearest item
- **1/3 chance**: Talk with nearby idle character
- **1/3 chance**: Stay idle

If the selected activity isn't possible (no items to look at, no idle characters nearby), the system falls back to the next option.

### Looking

Characters look at nearby items, which:
- Provides opportunity for preference formation (based on mood)
- Adjusts mood based on existing preferences
- Takes 3 seconds to complete
- Avoids looking at the same item twice in a row

### Talking

Characters seek out other characters doing idle activities (Idle, Looking, or already Talking):
- Initiator moves to adjacent position of target
- When adjacent, both characters enter "Talking with [Name]" state
- Conversation lasts 5 seconds
- Either character can be interrupted by Moderate+ needs
- When one partner is interrupted, both stop talking

Idle activities are interruptible by any Moderate or higher tier need that can be fulfilled.

## View Modes

Three view modes available during gameplay:

### Select Mode (default)
- Details panel shows selected entity info
- Action log shows events for selected character
- Press `S` to enter

### All Activity Mode
- Combined log showing all character events
- No details panel
- Press `A` to enter

### Full Log View
- Full-screen log with complete (non-truncated) messages
- Shows all events with character name prefix
- Useful for reading long debug messages (e.g., gradient scores)
- Press `L` to enter, `L` or `Esc` to exit
