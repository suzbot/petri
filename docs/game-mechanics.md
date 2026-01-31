# Game Mechanics Reference

Detailed game mechanics. For exact values, see `internal/config/config.go`.

## World Time

The simulation uses a compressed time scale where game time passes faster than "world time" (the narrative time experienced by characters). This allows observing multi-day events in minutes of real time.

**Time scale:** 1 game second ≈ 12 world minutes, meaning 1 world day = 2 game minutes.

All durations in the game are tuned to feel narratively appropriate at this scale:
- **Actions** (eating, drinking, looking): Minutes to under an hour of world time
- **Need cycles** (hunger, thirst): Multiple world days to reach critical levels
- **Item lifecycles** (spawning, flower death): Days of world time between events
- **Sleep**: Several world hours to fully rest

The time scale is documented in config comments and used consistently across all duration-based mechanics.

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

- **Berries**: Color, optional poisonous/healing; edible
- **Mushrooms**: Color + optional Pattern + optional Texture, optional poisonous/healing; edible
- **Gourds**: Color + optional Pattern + optional Texture; edible, never poisonous/healing
- **Flowers**: Color; non-edible (decorative)

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

Uses gradient scoring for all food sources (map items, carried items, vessel contents):

`Score = (NetPreference × PrefWeight) - (Distance × DistWeight) + HealingBonus`

Carried items and vessel contents have distance = 0; map items use Manhattan distance.

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

Press `K` in select mode to toggle the Knowledge panel (replaces action log). Shows two sections:
- **Facts**: Learned poison/healing knowledge (e.g., "Red berries are poisonous")
- **Knows how to**: Discovered activity skills (e.g., "Harvest")

Press `K` again to return to action log.

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

### Knowledge Transmission

Characters can share knowledge through conversation:
- When a conversation completes naturally (5 seconds), each character shares one random piece of knowledge
- If the partner doesn't already have that knowledge, they learn it
- Learning poison knowledge via transmission also creates a dislike preference (same as learning by eating)
- Interrupted conversations do not transmit knowledge

Log messages for transmission:
- Sharer: "Shared knowledge with [Name]"
- Learner: "Learned: [knowledge description]" + "Learned something!"

## Know-how System

Know-how represents activity skills that characters discover through experience. Unlike facts, know-how cannot be transmitted through talking.

### Discovery

Characters can discover know-how by performing related activities. Currently discoverable:
- **Harvest**: Discovered when foraging (picking up items), eating edible items, or looking at edible items

Discovery chance depends on mood:
- **Joyful**: Uses `config.KnowHowDiscoveryChance` (e.g., 5%)
- **Happy**: 20% of Joyful rate (e.g., 1%)
- **Neutral and below**: No discovery possible

When discovery occurs, "Discovered how to [Activity]!" appears in the action log (blue color).

### Know-how vs Facts

| Aspect | Facts (Knowledge) | Know-how |
|--------|-------------------|----------|
| Examples | "Red berries are poisonous" | "Harvest" |
| Learned by | Experience (eating/healing) | Discovery (foraging/eating/looking) |
| Can be transmitted | Yes (via talking) | No |
| Display | Knowledge panel: "Facts:" section | Knowledge panel: "Knows how to:" section |

## Orders System

Players can issue orders to direct characters to perform specific tasks. Orders are managed through the Orders panel.

### Orders Panel

Press `O` to toggle the Orders panel. The panel shows:
- List of current orders with status (Open, Assigned, Paused)
- Hints for available actions

Panel controls:
- `+` - Add a new order
- `c` - Enter cancel mode to remove orders
- `x` - Toggle between side panel and full-screen view
- `o` - Close the panel

### Adding Orders

To add an order:
1. Press `+` to start add order flow
2. Select an activity (only activities known by at least one living character appear)
3. Select a target type (e.g., for Harvest: choose berry, mushroom, or gourd)
4. Press Enter to confirm

### Order Status

- **Open**: Available to be taken by a character with the required know-how
- **Assigned**: Currently being worked on by a character
- **Paused**: Interrupted by character needs; will resume when needs are satisfied

### Requirements

Orders can only be created for activities that at least one living character knows. For example, Harvest orders require at least one character to have discovered Harvest know-how.

### Order Execution

When a character becomes eligible for an idle activity and has relevant know-how:

1. **Assignment**: Character takes the first available open order (first-come-first-served)
2. **Vessel Search**: If not carrying a vessel, looks for one that can hold target items
3. **Seeking**: Character moves toward nearest item matching the order's target type
4. **Pickup**: Character picks up the item into vessel (if carrying one) or inventory
5. **Continuation**: If carrying a vessel with space, continues harvesting more items
6. **Completion**: When vessel is full OR no matching items remain, order is completed

### Order Interruption and Resumption

- If a character's needs reach Moderate+ tier while working on an order, the order is **paused**
- The character addresses their needs first (eating, drinking, sleeping, etc.)
- When needs are satisfied and character becomes idle-eligible again, the **same order resumes**
- No re-evaluation occurs - character continues their assigned order

### Order Abandonment

Orders are abandoned (removed) if:
- No items matching the target type exist on the map
- The order is cancelled by the player while assigned

When cancelled while assigned, the character's assignment is cleared and they return to normal idle behavior.

### Order Priority vs Idle Activities

Order work takes priority over random idle activities:
1. Check for assigned order to resume → if found, continue order
2. Check for open orders character can take → if found, take order
3. Fall through to random idle activity selection (look/talk/forage/idle)

## Inventory

Characters can carry items in their inventory.

### Capacity

Current inventory capacity: 1 item. Characters can carry one item at a time.

### Viewing Inventory

Press `I` in select mode to toggle the Inventory panel (replaces action log). Shows "Carrying: [item description]" or "Carrying: nothing". If carrying a vessel, also shows the vessel's contents with stack count and capacity. Press `I` again to return to action log.

### Carried Item Properties

When an item is picked up:
- Item is removed from the map
- IsGrowing set to false (picked items don't spawn new items)
- Spawn/death timers are cleared (carried items are static)
- Item remains in inventory until consumed or dropped

### Dropping Items

Characters drop items when:
- Working on an order that requires picking up a different item
- Inventory is full and they need to pick up something else for an order

Dropped items:
- Remain where dropped (at character's position)
- Keep IsGrowing = false (don't spawn new items)
- Can be picked up again, eaten, or looked at

## Crafting

Characters can craft items from materials using recipes.

### Recipes

Recipes define what can be crafted:
- **hollow-gourd**: 1 gourd → 1 vessel (container with capacity 1)
  - Duration: 10 seconds (temporary - full duration 120 seconds)
  - Vessel inherits gourd's appearance (color, pattern, texture)
  - Vessel is not edible

### Discovery

Crafting know-how and recipes are discovered together:
- **craftVessel + hollow-gourd recipe**: Discovered via gourd interaction (look, pickup, eat) or drinking at a spring

Discovery chance depends on mood (same as other know-how discovery).

### Craft Orders

To craft items:
1. Player creates a Craft order (Orders panel → + → Craft → Vessel)
2. Character with craftVessel know-how takes the order
3. If carrying a gourd: begin crafting immediately
4. If not carrying a gourd: move to pick one up (dropping current item if needed)
5. Crafting takes recipe duration (uses ActionProgress like eating/drinking)
6. On completion: vessel goes to inventory, order completed

### Crafted Items

Crafted items differ from natural items:
- Have a display name (e.g., "Hollow Gourd") instead of attribute-based description
- Are not edible (vessel Edible = false)
- Don't spawn new items (Plant = nil)
- May have container storage (vessel has Capacity 1)

## Vessels & Containers

Vessels are containers that can hold stacks of items. Currently, vessels are crafted from gourds.

### Stack Sizes

Different item types stack to different limits within a vessel:
- **Berries**: 20 per stack
- **Mushrooms**: 10 per stack
- **Flowers**: 10 per stack
- **Gourds**: 1 per stack

### Variety Lock

When an item is added to an empty vessel, the vessel becomes "variety locked":
- Only items of the exact same variety (type + color + pattern + texture) can be added
- When the vessel is emptied, it accepts any variety again
- Hollow gourd vessels have capacity 1 (one stack)

### Foraging with Vessels

When a character carrying a vessel forages:
- Items are added to the vessel's stack instead of filling inventory
- Foraging continues automatically until:
  - Vessel is full (stack reached max size)
  - No more matching items exist on the map
- If the vessel is variety-locked and no matching items exist, foraging stops
- Foraging skips items incompatible with the carried vessel

### Harvesting with Vessels

When working on a harvest order with a vessel:
- Items are added to the vessel until full or no targets remain
- Order completes when vessel is full OR no matching items exist

### Look-for-Container

When starting to forage or harvest without carrying a vessel:
- Character first looks for an available vessel on the ground
- Available = empty OR partially filled with compatible variety AND has space
- Character picks up the vessel, then continues to forage/harvest into it
- If no suitable vessel found, picks up items directly (single item to inventory)

### Drop-when-Blocked

If a character's vessel cannot accept the target item (full or wrong variety):
- **For orders (harvesting)**: Drop the vessel and pick up the item directly (order takes priority)
- **For idle foraging**: Skip the incompatible item (don't lose vessel contents for casual activity)

### Eating from Vessels

Hungry characters can eat from vessel contents (carried or dropped on the ground):

**Unified food selection**: All food sources use the same scoring system:
- Carried loose item: distance = 0
- Carried vessel contents: distance = 0
- Dropped vessel contents: distance = Manhattan distance to vessel
- Map items: distance = Manhattan distance

Score formula: `Score = (NetPreference × PrefWeight) - (Distance × DistWeight) + HealingBonus`

This means:
- At Moderate hunger: carried disliked items are filtered; character may seek better food on map
- At Severe hunger: carried food has distance advantage but preferences still influence choice
- At Crisis hunger: closest food wins (carried food at distance=0 almost always wins)

**Effects from vessel contents**: When eating from a vessel, effects (poison, healing) come from the item variety stored in the stack. Knowledge and preferences are formed/applied based on the variety.

**Stack decrement**: Each time a character eats from a vessel, the stack count decreases by 1. When empty, the vessel can accept any variety again.

### Viewing Vessel Contents

- When carrying a vessel: Inventory panel shows contents with count and max stack size
- When selecting a dropped vessel: Details panel shows contents

## Idle Activities

When characters have no urgent needs (all stats below Moderate tier), they select from idle activities:

### Activity Selection

Every 10 seconds (IdleCooldown), characters roll for an idle activity:
- **1/4 chance**: Look at nearest item
- **1/4 chance**: Talk with nearby idle character
- **1/4 chance**: Forage for edible item (if inventory not full)
- **1/4 chance**: Stay idle

If the selected activity isn't possible (no items to look at, no idle characters nearby, inventory full), the system falls back to the next option.

### Looking

Characters look at nearby items, which:
- Provides opportunity for preference formation (based on mood)
- Adjusts mood based on existing preferences
- Takes 3 seconds to complete
- Avoids looking at the same item twice in a row

### Talking

Characters seek out other characters doing idle activities (Idle, Looking, Talking, or Foraging):
- Initiator moves to adjacent position of target
- When adjacent, both characters enter "Talking with [Name]" state
- Conversation lasts 5 seconds
- On completion: knowledge transmission occurs (see Knowledge Transmission above)
- Either character can be interrupted by Moderate+ needs
- When one partner is interrupted, both stop talking (no knowledge transmitted)

### Foraging

Characters pick up edible items to carry in inventory:
- Only available when inventory has room (not full, or carrying a vessel with space)
- Uses preference/distance scoring (same weights as moderate hunger eating)
- Character moves to item, then picks it up (takes ActionDuration to complete)
- Picked up item is removed from map and placed in inventory or vessel
- If not carrying a vessel, looks for one on the ground first (see Vessels & Containers)
- When carrying a vessel, continues foraging until vessel full or no matching items
- Logs "Foraging for [item type]" when starting, "Picked up [item]" on completion

Idle activities are interruptible by any Moderate or higher tier need that can be fulfilled.

## View Modes

Two view modes available during gameplay:

### Select Mode (default)
- Details panel shows selected entity info
- Action log shows events for selected character
- Press `S` to enter

### All Activity Mode
- Combined log showing all character events
- No details panel
- Press `A` to enter
- Press `X` to expand to full-screen (no message truncation)
- Press `X` again or `S` to collapse
