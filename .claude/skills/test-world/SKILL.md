---
name: test-world
description: "Create a test world with specific preconditions, when asking user to verify a behavior that's hard to observe naturally."
user-invocable: true
argument-hint: "Structured world spec — read this skill's Caller Requirements before invoking"
model: sonnet
agent: true
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
---

## Caller Requirements

**The caller must read this section and build a structured argument before invoking this skill.**

Specify all of the following in the skill argument:

- **Characters**: count, positions, stat overrides (if any), and exact inventory contents — item type, color, kind for each item; for vessels, include container contents
- **Map items**: what's on the ground, with full variety attributes (item_type, color, pattern, texture, kind, edible, poisonous, healing); for plants, specify growth state (mature, sprout, gone-to-seed)
- **Orders**: activity ID, target type, status, assignment (if pre-assigned)
- **Varieties to register**: every unique item_type/color/kind combination that appears anywhere — inventory, container contents, map items, seeds. If an item exists anywhere in the world, its variety must be listed here.
- **Support infrastructure**: water tile positions, leaf piles, other needs-satisfying features
- **What to observe**: the specific behavior the user should watch for

Do not leave entity attributes for the skill to infer. If a vessel contains water, specify `"item_type": "liquid", "color": "", "kind": "water"`. If a vessel contains berries, specify the exact color. The skill will translate this spec into a valid save file.

---

## Creating a Test World

You are creating a test world save file so the user can verify specific game behaviors. You will be given a structured world spec as your argument.

### Step 1: Read the Schema

Read `internal/save/state.go` for the complete `SaveState` struct and all serialization types. This is your source of truth for JSON field names and structure.

Also read `internal/config/config.go` for character/item symbols, spawn counts, and any relevant constants.

Check `docs/game-mechanics.md` for system interactions (e.g., hunger tiers, order interruption, idle activity triggers) before designing preconditions — this is faster than searching code and captures player-visible behavior that code comments may not.

### Step 2: Plan the World

Based on the test description, determine:
- What behavior needs verification?
- What preconditions make it observable? (character positions, stat levels, items, knowledge)
- If testing orderable activities, ensure characters have the required `known_activities` AND `known_recipes`. Check ActivityRegistry and RecipeRegistry for the activity's prerequisites.
- What items/features are needed for supporting needs (water for thirst, leaf piles for sleep)?
- What should be excluded to isolate the behavior?
- **Observability:** For behavioral sequences (A does X before B does Y), ensure the world forces the intended sequence — use position and stat asymmetry so the behavior can't play out ambiguously. Prefer one item, one actor for the key narrative. Ask: "Could a race condition obscure what I'm testing?"

### Step 3: Create World Directory

Use Bash to create: `~/.petri/worlds/world-test-<feature>/`

### Step 4: Write meta.json

```json
{
  "id": "world-test-<feature>",
  "name": "world-test-<feature>",
  "created_at": "<timestamp>",
  "last_played_at": "<timestamp>",
  "character_count": <n>,
  "alive_count": <n>
}
```

### Step 4.5: Remind User to Close Game (REQUIRED)

Before writing any files, remind the user: **"Close the game if it's running — auto-save on quit will overwrite the test world."** Wait for acknowledgment before proceeding to Step 5.

### Step 5: Write state.json

Start from the **recipe template** below, then customize for the test scenario.

**Stat semantics (IMPORTANT):**
- `hunger`: 0 = fully satisfied, 100 = starving to death
- `thirst`: 0 = fully satisfied, 100 = dying of thirst
- `energy`: 0 = exhausted, 100 = fully rested

**Default stat levels for test worlds:** hunger=0, thirst=0, energy=98. This saturates survival needs so characters focus on the target behavior. Override only when the test specifically requires a needy character (e.g., testing hunger-driven eating).

**Key considerations:**
- Use `version: 1` and set `map_width: 60`, `map_height: 60`
- Position characters near relevant items/features
- Pre-populate knowledge/known_activities if testing knowledge-dependent behavior
- Include water_tiles for drinking, features (leaf piles, feature_type: 2) for sleeping
- Include varieties for all item types present
- Set `talking_with_id: -1` for characters not in conversation
- All items need fields matching the `ItemSave` struct in `state.go`. See templates below for common examples.
- **Position format**: Characters, items, features, and water tiles all use flat `"x": N, "y": N` fields (NOT nested `"position": {"x": N, "y": N}`)
- **Vessel availability:** If testing a behavior that uses vessels, provide enough vessels to fill both inventory slots per character. Characters may autonomously pick up vessels for other idle activities (forage, fetch water) before the target behavior triggers — extra vessels prevent this from blocking the test.
- **Order-item match:** If testing orders, verify test items match the order's target type (e.g., growing plants for harvest, not loose berries).
- **Vessel varieties:** Water vessels require `"color": ""` (not `"blue"`) and `"kind": "water"` in both the stack contents and the varieties array — wrong color causes a nil-pointer crash on save. Extraction requires seed varieties pre-registered in varieties (e.g. `{"item_type": "seed", "kind": "tall grass seed", ...}`) — `AddToVessel` fails silently without them.

#### Type Reference

**Color, Pattern, Texture are strings** in the save format (not integers):
- Colors: `"red"`, `"blue"`, `"brown"`, `"white"`, `"orange"`, `"yellow"`, `"purple"`, `"tan"`, `"pink"`, `"black"`, `"green"`, `"pale_pink"`, `"pale_yellow"`, `"silver"`, `"gray"`, `"lavender"`
- Patterns: `""` (none), `"spotted"`, `"striped"`, `"speckled"`
- Textures: `""` (none), `"smooth"`, `"slimy"`, `"waxy"`, `"warty"`

#### Recipe Template

Use this as the starting structure. Add/remove characters, items, features, and varieties as needed for the specific test. Read `internal/save/state.go` for any fields not shown here.

```json
{
  "version": 1,
  "map_width": 60,
  "map_height": 60,
  "elapsed_game_time": 0,
  "characters": [
    {
      "id": 1,
      "name": "Kai",
      "x": 30,
      "y": 30,
      "health": 100,
      "hunger": 0,
      "thirst": 0,
      "energy": 98,
      "mood": 50,
      "talking_with_id": -1,
      "inventory": [],
      "preferences": [],
      "knowledge": [],
      "known_activities": [],
      "known_recipes": [],
      "assigned_order_id": 0
    }
  ],
  "items": [],
  "features": [],
  "water_tiles": [
    {"x": 25, "y": 25, "water_type": 2}
  ],
  "tilled_positions": [],
  "watered_tiles": [],
  "varieties": [],
  "orders": []
}
```

**Item template (loose edible):**
```json
{
  "id": 1,
  "x": 31,
  "y": 30,
  "item_type": "berry",
  "color": "red",
  "pattern": "",
  "texture": "",
  "edible": true,
  "poisonous": false,
  "healing": false,
  "death_timer": 0
}
```

**Growing plant template** (mature, can reproduce):
```json
{
  "id": 2,
  "x": 32,
  "y": 30,
  "item_type": "berry",
  "color": "red",
  "pattern": "",
  "texture": "",
  "edible": true,
  "poisonous": false,
  "healing": false,
  "death_timer": 0,
  "plant": {"is_growing": true, "spawn_timer": 0}
}
```

**Sprout template** (still maturing, NOT yet edible — `edible` must be `false`):
```json
{
  "id": 3,
  "x": 33,
  "y": 30,
  "item_type": "berry",
  "color": "red",
  "pattern": "",
  "texture": "",
  "edible": false,
  "poisonous": false,
  "healing": false,
  "death_timer": 0,
  "plant": {"is_growing": true, "spawn_timer": 0, "is_sprout": true, "sprout_timer": 120.0}
}
```

**Vessel with contents** (in inventory or on map). `name` is the display name, `container` holds stacks. Stack variety attributes must match an entry in the top-level `varieties` array:
```json
{
  "id": 10,
  "x": 0,
  "y": 0,
  "name": "Hollow Gourd",
  "item_type": "vessel",
  "kind": "hollow gourd",
  "color": "green",
  "pattern": "",
  "texture": "",
  "edible": false,
  "poisonous": false,
  "healing": false,
  "death_timer": 0,
  "container": {
    "capacity": 1,
    "contents": [
      {
        "item_type": "berry",
        "color": "red",
        "pattern": "",
        "texture": "",
        "count": 5
      }
    ]
  }
}
```

To put a vessel in a character's inventory, add it to the character's `"inventory"` array (same ItemSave format). Characters have 2 inventory slots.

**Feature template (leaf pile for sleeping):**
```json
{
  "id": 1,
  "x": 35,
  "y": 30,
  "feature_type": 2
}
```

**Order template** (pre-assign to a character by setting `assigned_to` and matching `assigned_order_id` on the character):
```json
{
  "id": 1,
  "activity_id": "gather",
  "target_type": "stick",
  "status": "assigned",
  "assigned_to": 1
}
```

**Order status values (IMPORTANT — must use exact strings):**
- `"open"` — available to be taken by any character
- `"assigned"` — currently being worked on (use this when pre-assigning to a character)
- `"paused"` — interrupted by character needs
- `"completed"` — finished (swept up by game loop)

**Variety template:**
```json
{
  "item_type": "berry",
  "color": "red",
  "pattern": "",
  "texture": "",
  "edible": true,
  "poisonous": false,
  "healing": false,
  "kind": ""
}
```

### Step 5.5: Verify File Exists (REQUIRED)

After writing state.json, verify the file was actually created (e.g., `ls` the directory). Do NOT report to user until the file exists.

### Step 6: Report to User

Tell the user:
1. What the world contains and why
2. To run `go build -o petri ./cmd/petri && ./petri`
3. Select the test world from the world list
4. What specific behavior to observe
5. Remind them to report results so the test world can be cleaned up
