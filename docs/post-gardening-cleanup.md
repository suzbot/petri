# Post-Gardening Cleanup

Small improvements and investigations to tackle after the Gardening phase is complete. These are individually scoped items that don't warrant a full phase plan — bundle them into a short cleanup sprint.

**Sources:** Items consolidated and removed from [randomideas.md](randomideas.md) and [triggered-enhancements.md](triggered-enhancements.md).

---

## Tech Stuff

1. Test files are really long, and seem to use a lot of time/tokens/context to search through. How can we handle?
    - paring down valueless tests?
    - combining similar tests?
    - breaking into smaller files?
    - anything in triggered enhancements that would help?
    - or, is this not really a problem and we don't need to worry about it right now?
3. why is create item from variety logic in character.go? are there other functions that ended up in unintuitive places?
4. Why have we started running into all these tab/whitespace problems while editing just in the last couple weeks or so? especially since claude is the only one reading/writing code?


---

## A. Pathing Improvement

1. Still occasional issues with pathing. Example: a target item is on the other side of an irregularly shaped pond. The character greedy steps and runs into the pond. Then they take a BFS step to go around the pond. Then they take a greedy step that puts them back where they were. They thrash between those two modes.
   Suggestion: once a character switches to BFS, they stay in BFS until they get to target or until they run into a character. If they run into a character they sidestep (as in current state) and then switch back to greedy step. 
---

## B. Esc Key Cleanup

Still a little inconsistent how esc behaves. Proposed new pattern: esc always takes you "back" a level. Ideal if there's a way to generalize this behavior.

**Desired behavior:**
- esc from any expanded view collapses the view
- esc from all activity view goes nowhere
- esc from orders start goes back to all activity
- esc from within orders only goes back one level
- esc from select: details view > action log or no panel goes back to all activity view
- esc from select: details view > any other panel goes back to details action log
- 'l' from select: details view > any other panel goes back to details action log
- q from anywhere goes to start file menu
- q from start file menu quits

---

## C. Order Selection UX

Pain point: scrolling through the order list every time. Replace with single keypress selection (numbered list).

---

## D. Game Mechanics Doc Reorg

The game-mechanics doc is a little disorganized, inconsistent about what level of detail it shares, not in the most intuitive order.

- Reorganize in order of gameflow
- Remove anything unnecessary for user, or summarize where makes sense

---

## E. ItemType Constants Evaluation

Gardening added seed, shell, stick, hoe (4 new item types). Total is now ~9 item types. Evaluate whether string comparisons like `item.ItemType == "gourd"` scattered through the codebase have become error-prone. If typos or inconsistencies emerge, formalize to constants:

```go
const (
    ItemTypeBerry    = "berry"
    ItemTypeMushroom = "mushroom"
    // ...
)
```

---

## F. Action Pattern Unification Investigation

Current action patterns for item consumption:
- **Eating**: Pre-selects specific item via `FindFoodTarget` scoring, passes as `TargetItem`, consumes via `Consume`/`ConsumeFromVessel`/`ConsumeFromInventory`
- **Crafting**: Uses `HasAccessibleItem` to check availability, `ConsumeAccessibleItem` to consume when complete

Both defer consumption to execution time (good), but use different patterns for checking/consuming.

Now that Gardening has added more actions (planting seeds, watering, tool usage), investigate:
1. Are there common patterns that could share helpers?
2. Should `HasAccessibleItem`/`ConsumeAccessibleItem` be generalized for all item consumption?
3. Do any actions have the "extract at intent time" anti-pattern that caused the craft loop bug?
4. Could a unified `ItemConsumer` interface simplify action handlers?

Only act if patterns are diverging or duplicating. If each action's needs are sufficiently different, the current approach may be fine.

---

## G. Streamlined Character Creation 

Remove single/multi mode distinction from UI and add character count control:

1. **UI Streamlining**
   - Remove single char mode from game start
   - Adjust start screen title and keys:
     - === Petri ===
     - R to start with Random Characters
     - C to create Characters
   - "C for Create" creates first character, randomizes the rest

2. **Character Count Modification**
   - Add an intermediate step to choose the number of characters to start with (max 16)
   - Note: Not a quick standalone win - requires refactoring the current character creation flow, which is why it's scoped together with UI Streamlining above

---

## H. Satiation and consumption duration:

1. Satiation tier of food should modify the amount of time it takes to consume. Meal = 15 world mins; Snack 1/3 that time; Feast 3x that time.
