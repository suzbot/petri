# Game Mechanics Reference

Detailed game mechanics for Petri. For exact values, see `internal/config/config.go`.

## Table of Contents

- [Overview](#overview)
- [Controls & Views](#controls--views)
- [World](#world)
- [Items](#items)
- [Characters](#characters)
- [Character Creation](#character-creation)
- [Preferences](#preferences)
- [Food & Consumption](#food--consumption)
- [Idle Activities](#idle-activities)
- [Knowledge & Know-how](#knowledge--know-how)
- [Orders](#orders)
- [Crafting](#crafting)
- [Inventory & Vessels](#inventory--vessels)
- [Gardening](#gardening)
- [Constructs](#constructs)

## Overview

Petri is a simulation where characters survive, learn, and develop culture autonomously. Characters have physical needs (hunger, thirst, energy, health) that drive their behavior through an urgency-based AI system. When needs are satisfied, characters idle — looking at items, talking to each other, foraging food, and fetching water. Through these interactions they form preferences, gain knowledge, and discover skills. Players can direct characters via orders to harvest, gather, craft, and garden.

## Controls & Views

### View Modes

**Select Mode** (default, press `S`):
- Details panel shows selected entity info
- Action log shows events for selected character
- Subpanels: `K` knowledge, `I` inventory, `P` preferences (replace action log)
- `L` or `Esc` closes subpanel and returns to action log

**All Activity Mode** (press `A`):
- Combined log showing all character events
- `X` to expand to full-screen, `X` again to collapse

### Navigation

- **Esc** always means "go back one level": collapse expanded view → close subpanel → close orders panel → return to All Activity. No-op from All Activity with nothing expanded.
- **Q** saves and returns to world select from anywhere during gameplay.

### Speed Control

- `<` — Slow down (1x → ½x → ¼x)
- `>` — Speed up (¼x → ½x → 1x)

Current speed shown in status bar when not at normal speed.

## World

### Time

The simulation uses a compressed time scale: 1 game second ≈ 12 world minutes, so 1 world day = 2 real-time minutes.

All durations are tuned to feel narratively appropriate:
- **Actions** (eating, drinking, looking): Minutes to under an hour of world time
- **Need cycles** (hunger, thirst): Multiple world days to reach critical levels
- **Item lifecycles** (spawning, flower death): Days of world time between events
- **Sleep**: Several world hours to fully rest

The current world day is displayed in the status bar.

### Terrain

- **Springs**: Single water tiles (`☉`), impassable, drink from adjacent tile
- **Ponds**: Contiguous blob-shaped clusters of 4-16 water tiles (`▓`), impassable, drink from adjacent tile. 1-5 ponds generated per world.
- **Wet tiles**: Tiles 8-directionally adjacent to any water tile are considered wet. Shown as "Wet" (blue) in the details panel. See also [Watered Tiles](#watered-tiles) for manually-watered tiles.
- **Clay deposits**: Clusters of passable dusky earth tiles (`░`) adjacent to water. Generated during world creation near ponds. Loose clay items spawn on some clay tiles. Clay tiles display "Clay deposit" in the details panel. See `config.ClayMinCount`, `config.ClayMaxCount`.
- **Tilled soil**: Ground modified by characters working Till Soil orders. See [Gardening](#gardening).
- **Leaf Piles**: Passable terrain features, used as beds for sleeping.

### Item Spawning

**Plant-based spawning:** Items with `IsGrowing = true` can spawn new items when total count drops below target. Children inherit all parent attributes. Spawn intervals configured per item type in `config.ItemLifecycle`.

**Ground spawning:** Non-plant items spawn periodically via independent timers, regardless of how many exist:
- **Sticks & Nuts**: Fall from the canopy onto random empty tiles
- **Shells**: Wash up adjacent to pond tiles

See `config.GroundSpawnInterval` for intervals.

**Death timers:** Items with a death interval are removed when their timer expires. Flowers and grass have death timers; edibles are immortal until eaten. See `config.ItemLifecycle`.

## Items

### Item Types

- **Berries**: Color, optional poisonous/healing; edible. Plantable when picked up.
- **Mushrooms**: Color + optional Pattern + optional Texture, optional poisonous/healing; edible. Plantable when picked up.
- **Gourds**: Color + optional Pattern + optional Texture; edible, never poisonous/healing. Consuming a gourd drops a seed.
- **Flowers**: Color; non-edible, decorative. Have death timers.
- **Nuts**: Brown `o`; edible, not poisonous/healing. Falls from canopy periodically.
- **Sticks**: Brown `/`; non-edible, crafting material. Falls from canopy periodically.
- **Shells**: Colored `<`; non-edible, crafting material. Multiple color variants. Washes up adjacent to ponds.
- **Grass**: Pale green `W`; non-edible, construction material. Kind "tall grass". Grows and spreads across the world; fastest lifecycle (fast maturation and reproduction). Has a death timer so it doesn't take over the map. Harvested into bundles; color changes to pale yellow on harvest, representing drying.
- **Seeds**: Dot `.` in parent's color; not edible, plantable. Carries the parent plant's variety ID so planted seeds grow into the correct variety with full fidelity. Kind reflects parent Kind when present (e.g., "tall grass seed", "flower seed", "gourd seed"). Auto-dropped when a gourd is consumed.
- **Clay**: Earthy `░` lump; non-edible, construction material. No varieties. Obtained via Dig Clay orders. Cannot be placed in vessels.
- **Bricks**: Terracotta `▬`; non-edible, construction material. No varieties. Crafted from clay via Craft Bricks orders. Cannot be placed in vessels.

### Varieties

Items are generated from varieties at world creation. Each variety defines a unique combination of attributes (type, color, pattern, texture). Multiple varieties per item type are generated, with 20% of edible varieties marked poisonous and 20% marked healing (mutually exclusive).

## Characters

### Stats & Thresholds

Characters have five stats with four severity tiers: Mild, Moderate, Severe, Crisis.

| Stat   | Direction | Notes |
| ------ | --------- | ----- |
| Hunger | Higher = worse | 0 is optimal, 100 is starving |
| Thirst | Higher = worse | 0 is optimal, 100 is dehydrated |
| Energy | Lower = worse | 100 is optimal, 0 triggers collapse |
| Health | Lower = worse | 100 is optimal, 0 is death |
| Mood   | Lower = worse | 100 is Joyful, 0 is Miserable |

<!-- Exception to "don't duplicate config values" rule: tier thresholds are referenced
     frequently across intent, orders, idle activities, and mood systems. Keeping them
     here avoids constant code lookups. Source: internal/entity/character.go -->

| Stat   | Mild | Moderate | Severe | Crisis |
| ------ | ---- | -------- | ------ | ------ |
| Hunger | ≥50  | ≥75      | ≥90    | 100    |
| Thirst | ≥50  | ≥75      | ≥90    | 100    |
| Energy | ≤50  | ≤25      | ≤10    | 0      |
| Health | ≤75  | ≤50      | ≤25    | ≤10    |
| Mood   | ≤89  | ≤64      | ≤34    | ≤10    |

**Stat rates:** Hunger and thirst increase continuously. Energy decreases when awake, with additional drain per movement step. Health decreases from starvation, dehydration, or poison damage. After a stat is fully satisfied, it holds at optimal for a brief cooldown before resuming natural change.

**Intent priority:** When multiple needs are elevated, the highest tier wins. Tie-breaker order: Thirst > Hunger > Health > Energy. A stat can only drive intent if it can actually be fulfilled (e.g., energy won't drive intent if no beds exist and the character isn't Exhausted).

### Speed & Movement

Base movement speed is modified by stacking penalties: poison, high thirst (tiered), and exhaustion (tiered). A minimum speed floor prevents complete immobility.

### Pathfinding

Characters use greedy-first pathfinding: they move diagonally toward their target and only run obstacle-aware (BFS) pathfinding if the direct step is blocked. Once a character switches to BFS to navigate around an obstacle (such as a pond), they stay in BFS mode for the rest of that intent — no oscillation back to greedy. BFS mode clears when the character reaches their target, changes intent, or bumps into another character (which triggers sideways displacement instead).

### Sleep

- Wake at full energy (in a bed) or partial energy (on the ground)
- Early wake only if another stat is Moderate+ and more urgent than energy
- Ground sleep available at Exhausted tier
- Collapse (immediate, involuntary) at 0 energy

### Frustration

When a character's need is at **Mild tier** and no resource is available, they fall through silently to a lower-priority need or discretionary activity — Mild needs do not interrupt idle activities.

When a need is at **Moderate+ tier** and no resource is available, the character's activity shows the specific blocker (e.g., "No bed available", "No suitable food available", "No water source available") and they wait for the world to change. The lower-priority needs and the discretionary bucket are still tried — the character is not stuck entirely, just blocked on that specific need. When a resource appears, the character acts on it immediately.

When a character's most urgent unfulfilled need remains unavailable and no other activity is possible, their activity shows **"Stuck (can't meet needs)"** — this is a safety-net fallback that should rarely appear. When urgent needs (Severe+) go unmet repeatedly, they become Frustrated. While frustrated they skip intent calculation, display "?" (orange), and have status "FRUSTRATED." Frustration clears after a timer expires.

### Mood

Mood reflects emotional state (0-100, higher is better). Levels: Joyful, Happy, Neutral, Unhappy, Miserable.

**Need-based changes:** Mood changes based on the highest need tier:
- All optimal: mood increases slowly
- Mild: no change
- Moderate: mood decreases slowly
- Severe: mood decreases at medium rate
- Crisis: mood decreases quickly

**Status effect penalties:** Poisoned and Frustrated each apply additional mood penalties that stack with each other and with need-based decay.

**Fulfillment boost:** When a need is fully satisfied (hunger→0, thirst→0, energy→100, health→100), mood receives a one-time boost.

**Food preference impact:** Eating liked or disliked food adjusts mood based on NetPreference (see [Preferences](#preferences)).

### Character Names

Random character names are drawn from `internal/entity/names.go`. Names can be edited during gameplay by pressing `E` in select mode when the cursor is on a character (max 16 characters).

## Character Creation

### Start Screen

After selecting "New World":
- **R** — Start with random characters (default count, randomized names/preferences)
- **C** — Open creation grid to customize characters

### Creation Grid

The creation grid shows character cards in rows of 4. Each card displays a name, food preference, and color preference — all randomized by default and individually editable.

- **+/-** — Add/remove characters (min 1, max 16)
- **Arrow keys/Tab** — Navigate between cards and fields
- **Ctrl+R** — Randomize all cards (preserving count)
- **Enter** — Start game with current characters

## Preferences

Characters have dynamic preferences that affect food selection and mood.

### Structure

Each preference targets item attributes:
- **ItemType only**: e.g., "Likes berries" (matches any berry)
- **Color only**: e.g., "Likes red" (matches any red item)
- **Pattern only**: e.g., "Likes Spots" (uses noun form)
- **Texture only**: e.g., "Likes Slime" (uses noun form)
- **Combo (2-3 attributes)**: e.g., "Likes spotted brown mushrooms" — always includes ItemType

Each preference has a **valence**: +1 (likes) or -1 (dislikes).

### NetPreference

When evaluating an item, sum all matching preference scores:
- Single-attribute preference: contributes `Valence × 1`
- Combo preference (2 attributes): contributes `Valence × 2`

| Character Preferences | Item | NetPreference |
|-----------------------|------|---------------|
| Likes berries, Likes red | Red berry | +2 (perfect) |
| Likes berries, Likes red | Blue berry | +1 (partial) |
| Likes red berries (combo) | Red berry | +2 (perfect) |
| Likes red berries (combo) | Blue berry | 0 (no match) |
| Likes red, Likes berries, Likes red berries | Red berry | +4 (all stack) |
| Dislikes slimy, Dislikes mushrooms, Dislikes slimy mushrooms | Slimy mushroom | -4 (all stack) |

### Initial Preferences

Characters start with two positive preferences based on character creation: likes [selected food type] and likes [selected color].

### Formation

Preferences form dynamically when consuming or looking at items and constructs, based on current mood:
- **Joyful/Happy**: chance to form positive preference (likes)
- **Neutral**: no formation
- **Unhappy/Miserable**: chance to form negative preference (dislikes)

**From items:**
- **Solo**: Single attribute (ItemType, Color, Pattern, or Texture)
- **Combo**: ItemType + 1-2 other attributes (max 3 total)

**From constructs:**
- **Solo**: Recipe identity (e.g., "stick fence"), material (e.g., "stick"), or color (e.g., "brown")
- **Combo**: Recipe identity + color, or material + color

If a character already has the exact same preference with opposite valence, the existing preference is removed instead of creating a new one.

### Cross-Application

Preferences formed from one source can affect mood when encountering another. Material preferences (ItemType) match constructs by their building material — "Likes sticks" boosts mood when looking at a stick fence. Color preferences also cross-apply between items and constructs.

### Viewing Preferences

Press `P` in select mode to toggle the Preferences panel. Likes shown in green, dislikes in yellow. Scrollable with `PgUp`/`PgDn`. Press `P` or `Esc` to return.

## Food & Consumption

### Eating

Eating duration varies by food size tier (see `config.ItemMealSize`). Feast-tier foods (gourds) take significantly longer than snack-tier foods (berries, nuts). Mushrooms are meal-tier (middle).

### Food Selection

Characters choose food using gradient scoring across all sources (map items, carried items, vessel contents, and ground vessel contents):

`Score = (NetPreference × PrefWeight) - (Distance × DistWeight) - |hunger - satiation| + HealingBonus`

Carried items and carried vessel contents have distance = 0; map items and ground vessels use Manhattan distance.

When a ground food vessel scores best, the character walks to it and eats in place — one unit at a time, re-evaluating between units, same as ground vessel drinking. The vessel stays on the ground; multiple characters can eat from the same communal vessel. No pickup is needed, so this works even with a full inventory.

The **satiation fit** term (`|hunger - satiation|`) penalizes food that's too small or too large for current hunger. A meal that matches current hunger scores best; snacks are penalized at high hunger, feasts are penalized at low hunger.

**Behavior by hunger tier:**
- **Moderate (50-74)**: High preference weight, low distance weight. Only considers non-disliked items. Preference and fit compete freely.
- **Severe (75-89)**: Medium preference weight, moderate distance weight. Considers all items including disliked. Filling food is worth a longer walk, but a decent meal on the path wins over a distant feast.
- **Crisis (90+)**: No preference weight, high distance weight. Nearest food wins regardless of fit or preference.

Idle foraging uses Moderate-tier weights. When scores are equal, closer item wins.

### Drinking

Characters prioritize water sources by distance:

| Source | Distance | Behavior |
|--------|----------|----------|
| Carried water vessel | 0 (always closest) | Drinks immediately without moving |
| Ground water vessel | Manhattan distance | Walks to vessel, drinks in place |
| Terrain (spring/pond) | Manhattan distance to adjacent tile | Walks adjacent, drinks from terrain |

**From terrain:** Characters drink continuously until thirst reaches 0. Intent persists across drinks.

**From vessels:** Each drink consumes 1 water unit and clears intent. Character re-evaluates and may drink again from the same vessel or seek another source.

### Poison & Healing Effects

Poisonous items deal damage and reduce speed. Healing items restore health. Both effects are properties of the item variety — characters don't know about them until they eat the item and gain knowledge (see [Knowledge](#knowledge--know-how)).

## Idle Activities

When characters have no urgent needs (all stats below Moderate tier), they select from idle activities. Idle activities are interruptible by any Moderate+ need that can be fulfilled. Mild needs do not interrupt idle activities.

### Activity Selection

Every 10 seconds, characters roll for an idle activity (equal probability each):
- **Look** at nearest item
- **Talk** with nearby idle character
- **Forage** for edible item (if inventory not full)
- **Fetch water** (fill empty vessel from water source)
- **Stay idle**

If the selected activity isn't possible, the system falls back to the next option.

### Looking

Characters look at nearby items and constructs, which provides opportunity for preference formation (based on mood) and adjusts mood based on existing preferences. Characters pick the nearest lookable target (item or construct) and walk to an adjacent tile. Avoids looking at the same target twice in a row.

### Talking

Characters seek other idle characters. When adjacent, both enter a conversation. On completion, knowledge transmission occurs (see [Knowledge Transmission](#knowledge-transmission)). Either character can be interrupted by Moderate+ needs; if one is interrupted, both stop and no knowledge is transmitted.

### Foraging

Characters pick up edible items to carry:
- Uses food selection scoring with satiation fit (see [Food Selection](#food-selection))
- If carrying a vessel, items are added to the stack instead of filling inventory
- **Vessel bonus scales with hunger**: at low hunger characters prefer stocking vessels (planning ahead); at high hunger they grab immediate food
- **Completes after one growing item** — casual activity, doesn't strip world resources
- Skips items incompatible with carried vessel

### Fetch Water

Characters fill vessels with water to carry:
- Skipped if already carrying water in a vessel
- Prefers a ground vessel already filled with water (pick up and carry — no fill phase)
- Otherwise seeks an empty vessel (carried or on ground), then moves to nearest water source and fills it
- Characters with non-water vessel contents seek a different vessel

### Helping (Crisis Response)

Before the idle roll, characters check if any community member is in Crisis hunger or thirst. If so, the idle character attempts to deliver food or water rather than doing a random idle activity.

**Trigger:** Another character reaches Crisis hunger or thirst (see stat thresholds above). The nearest character in crisis is targeted first — distance is the primary criterion. For equidistant characters, thirst takes priority over hunger. Multiple idle characters may respond to the same crisis simultaneously.

**Food delivery (helpFeed):** Triggered when the nearest crisis character has Crisis hunger. Helper scores available food using their own knowledge and preferences (poison avoidance, healing awareness, preference weights) with Severe-tier scoring weights — not the needer's Crisis-tier weights. This preserves the helper's judgment: they avoid known poisons and prefer more fitting food, while still prioritizing proximity. Satiation fit uses the needer's hunger level (100), so larger meals score better. Four candidate pools are scored: helper's carried loose food (distance 0), helper's carried food vessel (distance 0), ground loose food, and ground food vessels.

If the helper already carries food, they walk directly to the needy character. Otherwise, they procure food first (walk to it, pick it up), then deliver. The helper drops food on a cardinal-adjacent empty tile next to the needy character and calls out — visible in the action log as a social event (e.g., "Ada called out to Kira"). The call causes the needy character to re-evaluate, so they notice the closer delivered food instead of continuing toward a distant source.

**Water delivery (helpWater):** Triggered when the nearest crisis character has Crisis thirst. Helper procures a vessel with water and delivers it. The helper checks in order: carried water vessel (walk directly to needer), carried empty vessel (fill at water then deliver), ground water vessel (pick up and deliver, no fill needed), ground empty vessel (pick up, fill at water, deliver). If no vessel is available anywhere and the same crisis character also has Crisis hunger, the helper falls back to food delivery.

**Delivery:** In both cases, the helper drops the item on a cardinal-adjacent empty tile next to the needy character. The needy character's existing food/water scoring finds the delivered item naturally — ground food vessels and ground water vessels are already candidate sources.

**Limits:** Helpers with a full inventory and no usable food or vessel fall through to the normal idle roll. Helpers commit to delivery once started — no abandonment except when no candidate can be found.

## Knowledge & Know-how

Characters learn about the world through experience. Knowledge persists per-character and affects future behavior.

### Knowledge (Facts)

When a character eats a poisonous or healing item, they gain knowledge about that variety:
- **Poisonous**: Learns "[Variety] are poisonous" — automatically forms a dislike preference
- **Healing**: Learns "[Variety] are healing" — only if health actually increased

Knowledge is gained once per variety. Duplicates are not created.

**Poison knowledge → avoidance:** The auto-formed dislike preference affects food selection — at Moderate hunger the character avoids the item entirely; at Severe it scores lower but may be eaten; at Crisis nearest food wins regardless.

**Healing knowledge → health seeking:** When a character knows a healing item AND health is below full, health becomes a need that can drive intent (priority: Thirst > Hunger > Health > Energy). Known healing items also receive a scoring bonus when seeking food while hurt, scaling with health severity.

### Knowledge Transmission

Characters share knowledge through conversation:
- When a conversation completes naturally, each character shares one random piece of knowledge
- If the partner doesn't already have that knowledge, they learn it
- Learning poison knowledge via transmission also creates a dislike preference
- Interrupted conversations do not transmit knowledge

### Know-how

Know-how represents activity skills discovered through experience. Unlike facts, know-how cannot be transmitted through talking.

**Discovery triggers:**
- **Harvest**: Discovered when foraging, eating edible items, or picking up or looking at any harvestable plant (growing non-sprout plants, including grass and flowers)
- **Plant**: Discovered when picking up or looking at plantable items (berries, mushrooms, gourd seeds)
- **Extract**: Discovered when looking at a flower or tall grass, or when picking up or looking at a seed
- **Dig Clay**: Discovered when looking at or picking up a clay item
- **Build Fence**: Discovered per recipe — see [Fence Recipes](#fence-recipes)
- **Build Hut**: Discovered per recipe — discovered by looking at a fence of matching material (stick fence → stick-hut, thatch fence → thatch-hut, brick fence → brick-hut)
- **Crafting know-how**: See [Crafting Discovery](#discovery)

**Discovery chance depends on mood:**
- **Joyful**: Highest chance (see `config.KnowHowDiscoveryChance`)
- **Happy**: 20% of Joyful rate
- **Neutral and below**: No discovery possible

### Knowledge Panel

Press `K` in select mode to view. Shows two sections:
- **Facts**: Learned poison/healing knowledge
- **Knows how to**: Discovered activity skills

| Aspect | Facts | Know-how |
|--------|-------|----------|
| Examples | "Red berries are poisonous" | "Harvest", "Garden: Plant" |
| Learned by | Experience (eating) | Discovery (foraging/eating/looking) |
| Can be transmitted | Yes (via talking) | No |

## Orders

Players issue orders to direct characters to perform specific tasks. Orders take priority over random idle activities.

### Orders Panel

Press `O` to toggle. Controls:
- `+` — Add new order
- `c` — Cancel mode to remove orders
- `x` — Toggle side panel / full-screen view
- `O` — Close panel

### Adding Orders

1. Press `+` to start
2. Select an activity (only activities known by at least one living character appear; Gather requires no know-how)
3. Select a target type (varies by activity)
4. Press Enter or a number key (1-9) to select and confirm in one keypress

Press `Esc` at any step to go back one level.

### Order Status

- **Open**: Available to be taken by an eligible character
- **Assigned**: Currently being worked on
- **Paused**: Interrupted by character needs; will resume when needs are satisfied
- **Unfulfillable** *(display)*: Required items don't exist. Shown dimmed; clears automatically when world state changes.
- **No one knows how** *(display)*: No living character has the required skill. Shown dimmed.

### Order Execution

When a character becomes idle-eligible:
1. Check for assigned order to resume → continue if found
2. Check for open orders character can take → take first available
3. Fall through to random idle activity

**Harvest orders:** Character seeks growing items of the target type, picks them up into a vessel (or inventory). Continues until vessel is full, inventory is full with no vessel available, or no matching items remain.

**Extract orders:** Character procures a vessel and extracts seeds from matching plants without removing the plant. Plants can only be extracted from when their `SeedTimer` has expired (seeds regenerated). After extraction, the plant's seed timer resets — tied to the plant type's reproduction interval (fast-reproducing grass regenerates seeds faster than flowers). If all plants of the target type are temporarily depleted, the order shows as **[Unfulfillable]** and is skipped until seeds regenerate. Order completes when all plants of the target variety have been extracted (none left with seeds available), or inventory is full with no vessel space remaining for seeds. Requires Extract know-how.

When a fully grown extractable plant (flower or tall grass) has seeds available, the details panel shows **"Gone to seed"** below "Growing" — a signal that the plant is ready to extract from. Sprouts and non-extractable plants never show this indicator.

**Gather orders:** Character picks up loose (non-growing, non-container, non-tool) items from the ground. For items with registered varieties (seeds, nuts, shells), uses vessel procurement. For bundleable items (sticks), successive pickups merge into a bundle in inventory. When the bundle reaches max size (see `config.MaxBundleSize`), the character drops the completed bundle on the ground and the order completes — one bundle per order. No know-how required — all characters can gather.

**Dig Clay orders:** Character drops any non-clay items from inventory, walks to the nearest clay tile, and digs for a lump of clay (see `config.ActionDurationMedium`). Continues digging until both inventory slots hold clay, then drops both lumps on the ground and the order completes. Clay deposits are inexhaustible. Only available when clay deposits exist on the map. Requires Dig Clay know-how.

### Order Interruption and Resumption

**At Mild tier:** Characters with assigned orders check carried inventory first:
- If thirsty and carrying water → briefly pause to drink, then resume
- If hungry and carrying food → briefly pause to eat, then resume
- If no provisions carried → keep working through Mild

**At Moderate+ tier:** Order is paused. Character addresses needs, then resumes the same order when idle-eligible again.

### Order Abandonment

Orders are abandoned if no items matching the target type exist on the map, or if cancelled by the player while assigned.

When an order is abandoned (not cancelled), it enters a **cooldown period** (see `config.OrderAbandonCooldown`) before it can be retaken. During cooldown the order appears greyed out with an "Abandoned" status in the orders panel. Once the cooldown expires the order returns to "Open" and becomes available again. This prevents a rapid take/abandon loop when materials are temporarily unavailable.

## Crafting

### Recipes

- **hollow-gourd**: 1 gourd → 1 vessel (container). Vessel inherits gourd's appearance.
- **shell-hoe**: 1 stick + 1 shell → 1 hoe (tool for tilling soil). Hoe inherits shell's color.
- **clay-brick**: 1 clay → 1 brick. Repeatable — order continues until no loose clay remains on the map.

### Discovery

Crafting know-how and recipes are discovered together:
- **craftVessel + hollow-gourd**: Discovered via gourd interaction (look, pickup, eat) or drinking at a spring
- **craftHoe + shell-hoe**: Discovered via stick or shell interaction (look, pickup)
- **craftBrick + clay-brick**: Discovered via clay interaction (look, pickup, dig)

Discovery chance depends on mood (same as other know-how discovery — see [Know-how](#know-how)).

### Craft Orders

1. Player creates a Craft order (Orders panel → Craft → select activity)
2. Character with relevant know-how takes the order
3. If missing inputs: drops non-recipe items, seeks missing components
4. Crafting takes recipe duration (see `config.ActionDurationLong`)
5. On completion: crafted item drops on ground
6. For repeatable recipes (clay-brick): order continues until the completion condition is met (no more loose clay on the map); for non-repeatable recipes: order completes immediately

## Inventory & Vessels

### Inventory

Characters carry items in a 2-slot inventory. Each slot holds one item, one vessel (with contents), or one bundle. Press `I` in select mode to view.

When an item is picked up:
- Removed from the map, no longer grows or has spawn/death timers
- Berries and mushrooms become Plantable

Characters drop items when working an order that requires picking up a different item, or when inventory is full and they need something else.

### Bundles

Some item types (sticks, grass) stack as bundles when picked up. Each successive pickup of the same bundleable type merges into the carried bundle, incrementing its count. A bundle occupies one inventory slot. When a bundle reaches max size (see `config.MaxBundleSize`), no more items can merge into it — the slot is full.

Bundleable items cannot be placed in vessels. A bundle with count ≥ 2 on the ground shows as `X` on the map and as "bundle of sticks (N)" or "bundle of tall grass (N)" in descriptions. A single picked-up bundleable item (count 1) also shows bundle format — "bundle of sticks (1)" — since it has been removed from its growing state. Growing plants on the map keep their normal description regardless of BundleCount. The details panel shows "Bundle: N/M" capacity when selecting a bundle item.

### Vessels

Vessels are containers crafted from gourds that hold stacks of items. Different item types stack to different limits (see `config.GetStackSize()`).

**Variety lock:** When an item is added to an empty vessel, only items of the exact same variety can be added afterward. When emptied, the vessel accepts any variety again.

**Liquid contents:** Vessels can hold liquids (currently water) stacking to 4 units per vessel.

**Look-for-container:** When starting to forage or harvest without a vessel, characters first look for an available vessel on the ground (empty or compatible with space). If none found, they pick up items directly.

**Drop-when-blocked:** If no carried vessel can accept the target item — for orders, the character drops a vessel only when inventory is full (need space for a compatible vessel); for idle foraging, the character skips the incompatible item (don't lose vessel contents for casual activity). All vessel compatibility checks iterate every carried vessel, not just the first.

## Gardening

### Tilled Soil

Tilled soil is a terrain state that doesn't block movement or item placement. Created by characters working Till Soil orders.

**Visual:** Empty tilled tiles render as `═══` in dusky earth. Entities on tilled tiles render as `═X═`. Wet tilled tiles render in dark brown.

**Marked-for-tilling pool:** The player's tilling plan is stored as marked tiles, separate from Till Soil orders that assign workers:
- Marked tiles are added via area selection and persist independently of orders
- Till Soil orders are worker slots — multiple orders = multiple workers on the same pool
- Cancelling an order removes a worker but keeps marked tiles
- Unmarking tiles (via Garden > Unmark Tilling) removes them from the plan

**Till Soil action:** When a character tills a marked tile, growing items at the position are destroyed and non-growing items are displaced to an adjacent empty tile.

### Area Selection (Till Soil)

1. Select Garden > Till Soil from orders panel → enters area selection mode
2. Move cursor, press `p` to set anchor (first corner)
3. Move cursor to define rectangle (valid tiles highlight in teal)
4. Press `p` to confirm (tiles marked in sage)
5. Press `Tab` to toggle mark/unmark mode
6. Press `Enter` when done

### Planting

The plant order menu lists all plantable item types currently found in the world — on the ground, in vessels, or in character inventories. This includes flower seeds, tall grass seeds, gourd seeds, berries, and mushrooms. Only types that actually exist in the world appear in the menu.

When a character works a Plant order:
1. Procures a plantable item matching the order's target type (checks inventory, ground vessels, loose items)
2. Moves to the nearest available tilled tile — tiles with only loose non-growing items (seeds, vessels) are considered available; only tiles with a growing plant are skipped
3. Plants: pushes any loose non-growing items on the tile to an adjacent empty tile, then consumes the plantable item and places a sprout
4. On first plant, the order locks to that exact variety — subsequent procurement seeks only the same variety
5. Completes when no available tilled tiles remain (all occupied by growing plants) or no matching items are available

### Sprouts

Planted items are consumed and replaced with a sprout. Sprouts render as `𖧧` — dark teal on wet tiles, variety's color for mushrooms, muted green otherwise. Sprouts cannot be picked up, foraged, or harvested.

**Growth tiers** — maturation and reproduction are tuned per plant type (see `config.GetSproutDuration()` and `config.ItemLifecycle`):

| | Fast | Medium | Slow |
|--|------|--------|------|
| **Maturation** (sprout → plant) | Mushroom, Grass | Berry, Flower | Gourd |
| **Reproduction** (parent → new sprout) | Berry, Grass | Mushroom, Flower | Gourd |

Tilled and wet tile multipliers apply to both timers (see `config.TilledGrowthMultiplier`, `config.WetGrowthMultiplier`).

### Watered Tiles

Tiles can be wet from two sources:
- **Water-adjacent**: Tiles next to water tiles are always wet (computed on the fly)
- **Manually watered**: Tiles watered by a Water Garden order. Wetness decays over time (see `config.WateredTileDuration`).

**Water Garden orders:** Character discovers Water Garden know-how by filling a vessel with water. Character procures a vessel with water, walks to each dry tilled planted tile, and waters it (consuming 1 water unit per tile). Order completes when no dry tilled planted tiles remain.

**Growth effect:** Wet tiles accelerate sprout maturation and plant reproduction (see `config.WetGrowthMultiplier`).

## Constructs

Constructs are player-built structures placed on the map — distinct from natural terrain features like leaf piles. Current types: fences and hut walls/doors.

### Passability

Constructs have a passability property. Impassable constructs (such as fences) block movement for characters and are treated as obstacles by the pathfinding system. Characters navigate around impassable constructs the same way they navigate around water tiles.

### Details Panel

Selecting a tile that contains a construct shows:
- **Display name** — e.g., "Stick Fence", "Brick Hut Wall", "Stick Hut Door"
- **Type label** — e.g., "Structure"
- **"Not passable"** — shown for impassable constructs

### Rendering

Constructs render using their material's color.

**Fences:** A horizontal fence segment fills the full tile width (`╬╬╬`); a vertical segment is centered (`  ╬  `).

**Hut walls and doors:** Use heavy box-drawing characters with asymmetric horizontal fill. Symbols are computed at render time from adjacency — corners (`┏ ┓ ┗ ┛`), horizontal edges (`━━━`), vertical edges (`  ┃  `), door (`  ≡  `), and T/cross junctions for shared walls. Corners and door tiles use asymmetric fill to produce visual continuity with adjacent wall segments. Doors are passable; walls are not.

### Fence Recipes

Three fence recipes exist. Each is discovered independently:

| Recipe | Material | Discovery trigger |
|--------|----------|-------------------|
| Thatch Fence | 6 grass (one bundle) | Looking at or picking up harvested grass |
| Stick Fence | 6 sticks (one bundle) | Looking at or picking up a stick |
| Brick Fence | 6 bricks | Looking at or picking up a brick |

Discovering any fence recipe also grants **Build Fence** know-how. Characters learn individual recipes based on which materials they encounter — a character who has only handled sticks may know stick fences but not brick fences.

### Construction Orders

Construction orders appear under a **Construction** category in the orders panel. **Build Fence** requires at least one character to know a fence recipe. **Build Hut** requires at least one character to know a hut recipe (discovered by looking at any fence).

**Marking fence tiles** (step 2 of order creation):
1. Press `p` to anchor a start point
2. Move cursor — a cardinal line preview (horizontal or vertical, snapping to the larger axis) highlights valid tiles
3. Press `p` to confirm the line
4. Press `Tab` to toggle mark/unmark mode
5. Draw additional lines as needed; press `Enter` to create the order

**Marking hut footprints** (step 2 of order creation):
1. Move cursor — a 5×5 footprint preview follows the cursor (perimeter tiles in warm brown, interior tiles subtly shaded)
2. Press `p` to place the hut footprint; 16 perimeter tiles are marked
3. Press `Tab` to toggle unmark mode — hover over any marked tile and press `p` to remove the entire footprint
4. Place additional huts as needed; press `Enter` to create the order

Invalid footprint positions (water, built constructs, map edges, or existing marks in the interior) show red. Existing fence marks visible during hut placement render grey; the preview turns amber over them to indicate they'll be overwritten. Two huts can share a wall — shared perimeter tiles keep the first hut's mark (first-wins).

Marked-for-construction tiles are only highlighted during the active marking phase. In regular select mode, move the cursor over a tile to see "Marked for construction (Fence)" or "Marked for construction (Hut Wall)" / "Marked for construction (Hut Door)" in the details panel. Once a material is assigned, it shows e.g. "Marked for construction (Stick Fence)", "Marked for construction (Stick Hut Wall)", or "Marked for construction (Stick Hut Door)".

The character assigns a material to the line/footprint when they begin building the first tile — selecting whichever construction material is nearest and available. Once one tile in a line is built, all remaining tiles in that line use the same material.

### Building

When taking a fence order, a character selects the nearest fence material with enough supply (6+ items) and walks to build.

**Bundle materials (grass, sticks):** The character picks up a full bundle of 6 in one trip and builds immediately.

**Non-bundle materials (bricks):** The character uses a multi-trip supply-drop pattern — carrying 2 bricks per trip to the build site, dropping them, and returning for more until 6 are accumulated. They then build from a cardinally adjacent tile.

Characters always build from an adjacent tile — never from the build tile itself — to avoid being blocked inside finished sections. After the fence is placed, any items remaining on that tile are displaced to nearby clear tiles.

Multiple characters can work the same fence order simultaneously. Each worker picks the nearest unbuilt tile independently — no tile claiming occurs.

### Hut Building

When taking a hut order, a character selects the nearest available hut material and uses the **supply-drop pattern** for all material types: they carry 2 bundles or 2 bricks per trip to the target tile, drop them, and repeat until enough material has been staged. Material costs per tile:

| Material | Cost per tile |
|----------|--------------|
| Thatch (grass) | 2 full bundles (12 grass) |
| Stick | 2 full bundles (12 sticks) |
| Brick | 12 bricks |

After enough material is staged at a tile, the character builds from a cardinally adjacent standing tile. The constructed piece is a wall or door based on the tile's mark (the door is the center-south tile of the 5×5 footprint).

**Multi-worker support:** Multiple characters can work the same hut order simultaneously. Each worker picks the nearest unbuilt tile independently — no tile claiming occurs. If another worker occupies a candidate tile mid-build, the worker skips that tile rather than abandoning entirely.

**Material assignment:** The character assigns a material to the whole footprint when they begin building the first tile. All 16 tiles in the footprint use the same material.

**Feasibility:** An order is feasible if unbuilt hut marks exist and at least one free (un-staged) construction material exists on the map. Items that have already been delivered to a construction-marked tile are excluded from feasibility counting — they are committed to that site.

**Unfulfillable display:** If an order is both abandoned and infeasible, the orders panel shows "Unfulfillable" instead of "Abandoned."
