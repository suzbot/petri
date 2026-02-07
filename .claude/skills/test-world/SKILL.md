---
name: test-world
description: "Create a test world with specific preconditions, when asking user to verify a behavior that's hard to observe naturally."
---

## Creating a Test World

Test worlds let users verify specific behaviors by setting up exact preconditions.

### Step 1: Define What to Test
- What behavior needs verification?
- What preconditions make it observable? (character positions, stat levels, items, knowledge)

### Step 2: Create World Directory
~/.petri/worlds/world-test-<feature>/
├── state.json    # Full game state
└── meta.json     # World metadata

### Step 3: Write meta.json
```json
{
  "name": "world-test-<feature>",
  "created_at": "<timestamp>",
  "description": "Test world for <feature> - <brief description>"
}
```

### Step 4: Write state.json

Base structure (reference internal/save/state.go for complete schema):
```json
{
  "game_time": 0,
  "characters": [...],
  "items": [...],
  "features": [...],
  "varieties": [...],
  "next_item_id": <n>,
  "orders": []
}
```

Key setup considerations:
- Position characters near relevant items/features
- Set stat levels to trigger specific behaviors (e.g., low hunger to trigger eating)
- Pre-populate knowledge if testing knowledge-dependent behavior
- Include only items relevant to the test and items/features for needs that may arise (thirst, hunger, energy)

### Step 5: User Testing

Inform user:
1. Run ./petri
2. Select the test world from the world list
3. Observe the specific behavior
4. Report results

### Step 6: Cleanup

After verification, delete the test world directory if no longer needed.
