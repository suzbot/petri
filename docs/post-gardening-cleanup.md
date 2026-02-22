# Post-Gardening Cleanup

Small improvements and investigations to tackle after the Gardening phase is complete. These are individually scoped items that don't warrant a full phase plan — bundle them into a short cleanup sprint.

**Sources:** Items consolidated from [randomideas.md](randomideas.md) and [triggered-enhancements.md](triggered-enhancements.md).

---

## 1. Esc Key Cleanup

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

## 2. Order Selection UX

Pain point: scrolling through the order list every time. Replace with single keypress selection (numbered list).

---

## 3. Game Mechanics Doc Reorg

The game-mechanics doc is a little disorganized, inconsistent about what level of detail it shares, not in the most intuitive order.

- Reorganize in order of gameflow
- Remove anything unnecessary for user, or summarize where makes sense

---

## 4. ItemType Constants Evaluation

Gardening added seed, shell, stick, hoe (4 new item types). Total is now ~9 item types. Evaluate whether string comparisons like `item.ItemType == "gourd"` scattered through the codebase have become error-prone. If typos or inconsistencies emerge, formalize to constants:

```go
const (
    ItemTypeBerry    = "berry"
    ItemTypeMushroom = "mushroom"
    // ...
)
```

---

## 5. Action Pattern Unification Investigation

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
