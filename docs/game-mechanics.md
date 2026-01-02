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

## Preference System

Characters have dynamic preferences that affect food selection and mood.

### Preference Structure

Each preference targets item attributes:
- **ItemType only**: e.g., "likes berries" (matches any berry)
- **Color only**: e.g., "likes red" (matches any red item)
- **Combo**: e.g., "likes red berries" (matches only red berries)

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

### Food Selection by Hunger Tier

Higher hunger = less picky about preferences:
- Mild: Perfect match only
- Moderate: Perfect, then Partial
- Severe: Partial, then any
- Crisis: Any food

### Initial Preferences

Characters start with two positive preferences based on character creation:
- Likes [selected food type]
- Likes [selected color]

### Preference Formation

Preferences form dynamically when consuming food, based on current mood:
- Joyful/Happy: chance to form positive preference (likes)
- Neutral: no formation
- Unhappy/Miserable: chance to form negative preference (dislikes)

Formation chances and type weights configured in config.go.

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
