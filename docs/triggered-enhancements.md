### Deferred Enhancements & Trigger Points

Technical and Feature items analyzed and consciously deferred until trigger conditions are met.

| Enhancement                            | Triggers (implement when ANY is met)                                              |
| -------------------------------------- | --------------------------------------------------------------------------------- |
| **Parallel Intent Calculation**        | Character count ≥ 16; Intent calc exceeds ~1ms/char; Tick time exceeds 50ms       |
| **EventType for ActionLog**            | Event filtering in UI; Event-driven behavior; Event persistence                   |
| **Category type formalization**        | Adding non-plant categories (tools, crafted items); Need category-based filtering |
| **Category-based world spawning**      | When Category is formalized; Filter spawn loop by category (plants spawn, tools don't) |
| **Performance optimizations**          | Noticeable lag; Profiling shows bottlenecks                                       |
| **Preference formation for beverages** | Other beverages (non-spring) introduced                                           |
| **Depression and Rage mechanics**      | Job acceptance logic implemented                                                  |
| **Testify test package**               | Test complexity warrants it; Current assertion patterns become unwieldy           |
| **Wandering activity**                 | More idle activities needed; Looking feels repetitive; Want more organic movement |
| **ItemType constants**                 | Adding new item types; Want compile-time safety for item type checks              |
| **Activity enum**                      | Need to enumerate/filter activities; Activity-based game logic beyond display     |
| **Order completion criteria refactor** | Adding new order types; Completion logic scattered across update.go becomes unwieldy |
| **Action pattern unification investigation** | Post-Gardening; Have 3+ action types with item consumption; Patterns diverge or duplicate |
| **Feature capability derivation**      | Adding new feature types; DrinkSource/Bed bools become redundant                  |
| **Action log retention policy**        | Implementing character memory; May need world time vs real time consideration     |
| **UI color style map**                 | Adding new colors frequently; Switch statement maintenance becomes tedious        |
| **Cobra CLI migration**                | Next time we want to add a flag; Current flag parsing becomes unwieldy            |
| **Knowledge/Learning pattern review** | Enhanced Learning phase begins; Adding new knowledge types or transmission methods |
| **UI extensibility refactoring** | UI structure blocks adding new activities or features; Area selection pattern needs generalization |
| **Preference-weighted component procurement** | Multiple craft recipes with varied inputs; Characters feel flat when seeking components; Scoring math exists in foraging.go to reuse |
| **applyIntent duplication (simulation.go)** | INTENTIONALLY SEPARATE - simulation.go is lighter test harness; only unify if maintaining both becomes burdensome |

### Future Enhancement Details

**Balance tuning candidates** (see config.go for current values):

- Prefernce formation still feels too easy
- Activity durations
- Satisfaction cooldown
- Sleep wake thresholds
- Movement energy drain
- Health tier thresholds
- Mood adjustment rates and modifiers

---

**Performance optimizations:**

- Skip wake checks for sleeping characters unless tier changed
- Dead character filtering at map level
- Cache nearest features if map static-

---

**Parallel Intent pattern:**

```go
var wg sync.WaitGroup
for _, char := range chars {
    wg.Add(1)
    go func(c *entity.Character) {
        defer wg.Done()
        c.Intent = system.CalculateIntent(c, items, gameMap, log)
    }(char)
}
wg.Wait()
```

---

**Order completion criteria refactor (post-Gardening evaluation):**

Pre-Gardening audit identified this as "monitor during Gardening." Current state:
- Harvest completion: update.go lines 733-765
- Craft completion: update.go lines 840-846

Gardening adds 3+ order types (Till Soil, Plant, Water Garden). If completion logic exceeds ~50 lines of conditionals, refactor to handler pattern:
```go
type OrderCompletionHandler func(char *Character, order *Order, result PickupResult) bool
var OrderCompletionHandlers = map[string]OrderCompletionHandler{
    "harvest": handleHarvestCompletion,
    "craftVessel": handleCraftCompletion,
    // ... new handlers
}
```

---

**ItemType constants (post-Gardening evaluation):**

Gardening adds: seed, shell, stick, hoe (4 new item types).
Current item types: berry, mushroom, gourd, flower, vessel (~5 types).

After Gardening, there will be ~9 item types. Evaluate whether string comparisons like `item.ItemType == "gourd"` scattered through the codebase have become error-prone. If typos or inconsistencies emerge, formalize to constants:
```go
const (
    ItemTypeBerry    = "berry"
    ItemTypeMushroom = "mushroom"
    // ...
)
```

---

**Category formalization (evaluate during Construction):**

Gardening introduces first tool (hoe). Construction introduces more categories:
- Tools: hoe, (future: axe, hammer)
- Building materials: grass bundles, stick bundles, bricks
- Structures: fence, hut

When Construction begins, formalize categories to enable:
- Category-based spawning (plants spawn naturally, tools don't)
- Category-based inventory rules (tools can't go in vessels)
- Category-based preference formation

---

**Action pattern unification investigation (post-Gardening):**

Current action patterns for item consumption:
- **Eating**: Pre-selects specific item via `FindFoodTarget` scoring, passes as `TargetItem`, consumes via `Consume`/`ConsumeFromVessel`/`ConsumeFromInventory`
- **Crafting**: Uses `HasAccessibleItem` to check availability, `ConsumeAccessibleItem` to consume when complete

Both defer consumption to execution time (good), but use different patterns for checking/consuming.

After Gardening adds more actions (planting seeds, watering, tool usage), investigate:
1. Are there common patterns that could share helpers?
2. Should `HasAccessibleItem`/`ConsumeAccessibleItem` be generalized for all item consumption?
3. Do any actions have the "extract at intent time" anti-pattern that caused the craft loop bug?
4. Could a unified `ItemConsumer` interface simplify action handlers?

Only investigate if patterns are diverging or duplicating. If each action's needs are sufficiently different, the current approach may be fine.

---

**Preference-weighted component procurement → unified item-seeking in picking.go:**

When gathering recipe inputs (e.g., shells for Craft Hoe), characters currently pick the nearest available component by distance. With shell color variety (7 colors), characters could instead score components by preference and distance, similar to food seeking's gradient scoring in foraging.go.

picking.go is the shared home for "how characters acquire items." Currently, foraging.go has the most mature item-seeking logic (`scoreForageItems`, `createPickupIntent` — preference + distance scoring), while recipe procurement in picking.go uses simpler nearest-distance via `findNearestItemByType`. These should converge: characters' item-seeking behavior should be consistent whether they're foraging, gathering craft components, or fulfilling orders. Preference shapes material culture — a character who prefers silver shells will craft silver shell hoes, developing a personal aesthetic.

Candidates to generalize into picking.go when this triggers:
- `scoreForageItems` (foraging.go) → generic `scoreItemsByPreference` in picking.go
- `createPickupIntent` (foraging.go) → generic intent builder in picking.go (currently duplicated across foraging.go, order_execution.go, picking.go with different logging)
- `EnsureHasRecipeInputs` component seeking → use preference-weighted scoring instead of nearest-distance

**Recipe selection by preference** is a natural extension of this same system. `findCraftIntent` currently picks the first feasible recipe for an activity. When multiple recipes produce the same product (e.g., shell hoe, metal hoe, wooden hoe), the character could score each feasible recipe by net preference for its inputs — a character who likes silver shells would prefer the shell-hoe recipe and then prefer silver shells as input. The `findCraftIntent` structure (get feasible recipes → pick one → gather inputs) has the selection point built in at step 3.

Deferred because: Craft Hoe is the first multi-component recipe, and there's only one recipe per activity. Nearest-distance works fine. Revisit when multiple recipes per activity create enough variety that "character ignores preferred recipe/material" feels wrong.

---

**Preference formation for beverages (evaluate during Gardening):**

Gardening's Fetch Water activity introduces water-filled vessels as drinkable items. This technically triggers "Other beverages (non-spring) introduced."

However, water vessels may not warrant full preference formation since:
- Water is water (no variety like berry colors)
- The vessel itself already has preference-forming attributes

Evaluate whether characters should form preferences about "drinking from vessel" vs "drinking from spring" or defer until actual beverage variety exists (juice, tea, etc.).
