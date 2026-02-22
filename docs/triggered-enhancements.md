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
| **Activity enum**                      | Need to enumerate/filter activities; Activity-based game logic beyond display     |
| **`continueIntent` early-return block consolidation** | 5+ action-specific early-return blocks in `continueIntent`; New multi-phase action needs early-return and the pattern feels repetitive |
| **Feature capability derivation**      | Adding new feature types; DrinkSource/Bed bools become redundant                  |
| **Action log retention policy**        | Implementing character memory; May need world time vs real time consideration     |
| **UI color style map**                 | Adding new colors frequently; Switch statement maintenance becomes tedious        |
| **Cobra CLI migration**                | Next time we want to add a flag; Current flag parsing becomes unwieldy            |
| **Knowledge/Learning pattern review** | Enhanced Learning phase begins; Adding new knowledge types or transmission methods |
| **UI extensibility refactoring** | UI structure blocks adding new activities or features; Area selection pattern needs generalization |
| **Preference-weighted component procurement** | Multiple craft recipes with varied inputs; Characters feel flat when seeking components; Scoring math exists in foraging.go to reuse |
| **applyIntent duplication (simulation.go)** | INTENTIONALLY SEPARATE - simulation.go is lighter test harness; only unify if maintaining both becomes burdensome |
| **Temporarily-blocked order cooldown** | Assign/abandon churn noticeable for temporarily-blocked (but feasible) orders; Character visibly thrashes between taking and abandoning the same order |
| **Interactive inventory details panel** | Items gain enough attributes that parenthetical summary in inventory list isn't sufficient; Want to inspect individual inventory items with full details view |
| **Drinkable bool on ItemVariety**      | Non-drinkable liquid introduced (lamp oil, dye); Need to distinguish consumable vs non-consumable liquids |
| **Watertight bool on ContainerData**   | Non-watertight container introduced (basket, sack); Need to prevent liquid storage in permeable containers |
| **Idle activity registry refactor**    | Idle activity needs discovery/know-how gating; Idle activities need data-driven selection (weighted probabilities, personality); Adding 7th+ hardcoded idle option feels unwieldy |
| **Satiation-aware targeting & snacking threshold** | Characters walk past nearby berries because hunger isn't high enough to trigger foraging, then starve later; Snack-tier food feels useless because characters only eat when already quite hungry |
| **Terrain-aware Look + discovery**     | Adding activity discovered from terrain interaction; Characters feel artificially limited to item-only observation; Want richer idle contemplation behaviors |

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

_(ItemType constants evaluation moved to [post-gardening-cleanup.md](post-gardening-cleanup.md).)_

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

_(Action pattern unification investigation moved to [post-gardening-cleanup.md](post-gardening-cleanup.md).)_

---

**`continueIntent` early-return block consolidation:**

`continueIntent` (movement.go) has action-specific early-return blocks for actions whose `TargetItem` can be in different locations across phases (ground vs inventory). Current blocks: `ActionConsume`, `ActionDrink` (carried vessel), `ActionFillVessel`, `ActionWaterGarden`. These exist because the generic path's `ItemAt` check would nil the intent when the item moves to inventory.

The blocks share a common shape: check if item is in inventory → return unchanged (no movement needed) or recalculate toward `Dest`. If this grows to 5+ blocks, evaluate whether a general "multi-phase item tracking" pattern could replace per-action blocks — e.g., a helper that checks inventory-then-map for `TargetItem` and recalculates accordingly, called by each action's block or replacing them entirely.

Don't consolidate prematurely — the current 4 blocks are manageable and each has slight phase-detection differences. The trigger is when adding a new block feels like copy-paste.

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

**Plant order variety locking → character-driven behavior:**

Currently, `LockedVariety` is a field on the Order struct — the order itself remembers which variety the character started planting. This is a game-mechanical constraint rather than emergent character behavior. Two directions to explore:

1. **Knowledge/memory approach**: The character develops working memory ("I'm planting green gourds") rather than the order dictating it. Aligns with VISION.txt's principle that state lives in characters, not global structures. Evaluate when the knowledge system gets richer.

2. **Emergent from pickup limitations**: The order just says "plant gourds." The variety lock emerges naturally because the character picks up whatever gourd seeds are nearest/preferred, and once they have a vessel full of one variety, they keep planting that variety until it runs out. No explicit lock needed — the lock is a side effect of inventory state. This is more realistic and removes a special-case field from Order.

Deferred because: Current implementation works. Revisit when orders gain more character-agency features or when the knowledge system supports working memory.

---

**Terrain-aware Look + discovery:**

Characters can look at items but not terrain — narratively artificial. Water, tilled soil, and future terrain types (cliff faces, mineral deposits) should be observable. Two extensions:

1. **Lookable terrain**: Extend the Look idle activity to target terrain positions (water-adjacent, etc.). Characters "contemplate" a pond or spring. No preference formation (terrain has no item attributes), but fires discovery checks. Requires: look-targeting logic to include terrain positions alongside items, ActionLook handler to accept terrain targets.

2. **Terrain discovery triggers**: Add `TerrainType string` to `DiscoveryTrigger`. New `CheckTerrainDiscovery(char, action, terrainType)` called when characters interact with terrain. Enables "character looks at water → discovers Water Garden" or "character looks at cliff → discovers Mining."

These are complementary — lookable terrain provides the idle behavior, terrain triggers provide the discovery mechanism. Together they create a richer observation loop where characters learn from their environment, not just from items.

Deferred because: Water Garden discovery is adequately served by item-based triggers (sprouts) + action-completion triggers (ActionFillVessel). Lookable terrain is flavor, not blocking. Revisit when a future activity's discovery genuinely requires terrain observation, or when idle behaviors need more variety.

---

**Satiation-aware targeting & snacking threshold:**

Slice 9 introduces per-item satiation tiers (Feast 50, Meal 25, Snack 10). Two related enhancements to consider after living with these values:

1. **Satiation-aware foraging targeting**: When hungry, characters could factor satiation value into foraging scores — preferring a gourd (50 pts) over a berry (10 pts) when both are nearby, especially at high hunger. Currently foraging uses distance + preference scoring without considering how much the food will actually help.

2. **Lower foraging threshold for snacks**: Characters could be willing to snack at a lower hunger level (e.g., Mild tier) if food is immediately adjacent or in inventory — "grab a berry while walking past" behavior. Currently all food triggers foraging at the same hunger threshold regardless of convenience. This parallels the vessel drinking threshold concept (Slice 9 Step 5): lower the trigger when the cost is near-zero.

Deferred because: Need to observe how the new satiation tiers play out before adding complexity. The current system may produce acceptable behavior through existing proximity scoring — characters naturally eat nearby food first. Revisit after playtesting Slice 9.

---

**Preference formation for beverages (evaluate during Gardening):**

Gardening's Fetch Water activity introduces water-filled vessels as drinkable items. This technically triggers "Other beverages (non-spring) introduced."

However, water vessels may not warrant full preference formation since:
- Water is water (no variety like berry colors)
- The vessel itself already has preference-forming attributes

Evaluate whether characters should form preferences about "drinking from vessel" vs "drinking from spring" or defer until actual beverage variety exists (juice, tea, etc.).

