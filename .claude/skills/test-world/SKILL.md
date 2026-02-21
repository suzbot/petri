---
name: test-world
description: "Create a test world with specific preconditions, when asking user to verify a behavior that's hard to observe naturally."
user-invocable: true
argument-hint: What behavior to test and what preconditions are needed
model: sonnet
agent: true
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
---

## Creating a Test World

You are creating a test world save file so the user can verify specific game behaviors. You will be given a description of what to test as your argument.

### Step 1: Read the Schema

Read `internal/save/state.go` for the complete `SaveState` struct and all serialization types. This is your source of truth for JSON field names and structure.

Also read `internal/config/config.go` for character/item symbols, spawn counts, and any relevant constants.

### Step 2: Plan the World

Based on the test description, determine:
- What behavior needs verification?
- What preconditions make it observable? (character positions, stat levels, items, knowledge)
- If testing orderable activities, ensure characters have the required `known_activities` AND `known_recipes`. Check ActivityRegistry and RecipeRegistry for the activity's prerequisites.
- What items/features are needed for supporting needs (water for thirst, leaf piles for sleep)?
- What should be excluded to isolate the behavior?

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

### Step 5: Write state.json

Write a complete save state following the schema from Step 1. Key considerations:
- Use `version: 1` and set `map_width: 60`, `map_height: 60`
- Position characters near relevant items/features
- Set stat levels to trigger specific behaviors (e.g., low hunger to trigger eating)
- **Saturate needs aggressively** (hunger=0, thirst=0, energy=98) to prevent characters from pursuing needs instead of the target behavior
- Pre-populate knowledge/known_activities if testing knowledge-dependent behavior
- Include water_tiles for drinking, features (leaf piles, feature_type: 2) for sleeping
- Include varieties for all item types present
- Set `next_item_id` higher than the highest item ID used
- Set `talking_with_id: -1` for characters not in conversation
- All items need complete fields: id, position, item_type, color, pattern, texture, edible, poisonous, healing, death_timer
- **Vessel availability:** If testing a behavior that uses vessels, provide enough vessels to fill both inventory slots per character. Characters may autonomously pick up vessels for other idle activities (forage, fetch water) before the target behavior triggers — extra vessels prevent this from blocking the test.

### Step 5.5: Verify Written Files (REQUIRED)

After writing state.json, read it back and confirm key counts match intent:
- Character count
- Item count
- Tilled positions count (if applicable)
- `elapsed_game_time` is 0 (fresh world)
- Characters have expected inventory contents

Do NOT report to user until verification passes.

### Step 6: Report to User

Tell the user:
1. What the world contains and why
2. To run `go build -o petri ./cmd/petri && ./petri`
3. Select the test world from the world list
4. What specific behavior to observe
5. Remind them to report results so the test world can be cleaned up
