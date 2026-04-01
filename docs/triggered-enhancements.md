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
| **Testify test package**               | Assertion boilerplate drowns out test intent (currently ~40% of test length — trigger when it tips past 50%); Table-driven tests would reduce duplication but `if/t.Errorf` pattern makes them verbose; Writing a new test and spending more time on assertion scaffolding than on expressing what the test validates |
| **Test file splitting**                | Largest test files (update_test 3K, movement_test 2.9K, order_execution_test 2.5K) cause context overhead during pattern searches; A specific file becomes painful to navigate during implementation |
| **Wandering activity**                 | More idle activities needed; Looking feels repetitive; Want more organic movement |
| **Activity enum**                      | Need to enumerate/filter activities; Activity-based game logic beyond display     |
| **`continueIntent` early-return block consolidation** | 5+ action-specific early-return blocks in `continueIntent`; New multi-phase action needs early-return and the pattern feels repetitive |
| **Need-evaluator split from intent.go** | Adding a 5th stat type to the priority loop; Scrolling past 400+ lines of need evaluators to find routing logic in intent.go |
| **Feature capability derivation**      | Adding new feature types; DrinkSource/Bed bools become redundant                  |
| **Action log retention policy**        | Implementing character memory; May need world time vs real time consideration     |
| **Cobra CLI migration**                | Next time we want to add a flag; Current flag parsing becomes unwieldy            |
| **Knowledge/Learning pattern review** | Enhanced Learning phase begins; Adding new knowledge types or transmission methods |
| **UI extensibility refactoring** | UI structure blocks adding new activities or features; Area selection pattern needs generalization |
| **applyIntent duplication (simulation.go)** | INTENTIONALLY SEPARATE - simulation.go is lighter test harness; only unify if maintaining both becomes burdensome |
| **Order-aware simulation for e2e testing** | Construction adds multiple new ordered actions with supply procurement; Post-pickup handler branching bugs recur across ordered actions |
| **Temporarily-blocked order cooldown** | Assign/abandon churn noticeable for temporarily-blocked (but feasible) orders; Character visibly thrashes between taking and abandoning the same order |
| **Smarter displacement direction** | Displacement oscillation visible in crowded (16-char) worlds; Characters pick a displacement perpendicular that leads back toward the blocker instead of away |
| **Craft order quantity selection** | User creates many individual craft orders to get desired quantity; Brick demand from construction exceeds convenience of "craft all available" |
| **Interactive inventory details panel** | Items gain enough attributes that parenthetical summary in inventory list isn't sufficient; Want to inspect individual inventory items with full details view |
| **Plant order generalizes to ItemType seeds** | Multiple Kinds exist per parent ItemType (e.g., tall grass + bamboo both have ItemType "grass"); Plant order menu should show "grass seed" and let characters choose Kind, rather than listing each Kind separately |
| **SourceVarietyID on all plantable items** | Berry/mushroom planting uses GenerateVarietyID fallback to look up parent variety; a second place needs the same fallback; Simplify by setting SourceVarietyID at pickup time (when Plantable is set) so all plantables carry it uniformly |
| **Drinkable bool on ItemVariety**      | Non-drinkable liquid introduced (lamp oil, dye); Need to distinguish consumable vs non-consumable liquids |
| **Watertight bool on ContainerData**   | Non-watertight container introduced (basket, sack); Need to prevent liquid storage in permeable containers |
| **Item property registry migration** | Adding new item type requires touching config.go for non-tunable structural mappings; Confusion about what belongs in config vs entity registries |
| **Idle activity registry refactor**    | Idle activity needs discovery/know-how gating; Idle activities need data-driven selection (weighted probabilities, personality); Adding 7th+ hardcoded idle option feels unwieldy |
| **Snacking threshold** | Characters walk past nearby berries because hunger isn't high enough to trigger foraging, then starve later; Snack-tier food feels useless because characters only eat when already quite hungry |
| **Terrain-aware Look + discovery**     | Adding activity discovered from terrain interaction; Characters feel artificially limited to item-only observation; Want richer idle contemplation behaviors. *Context: Dig Clay (DD-17) worked around this by spawning loose clay items for item-based discovery triggers. Terrain-look would be the proper solution — characters notice clay terrain directly.* |
| **Character event/signal system**      | Helping needs richer reactions (gratitude, relationship changes); Multiple systems need to signal between characters; Current intent-clearing is too coarse for nuanced responses |
| **Harvest PickupToInventory handler clarity** | A second vessel-excluded harvestable type is added; The `GetCarriedVessel() == nil` check in applyPickup's harvest PickupToInventory path becomes confusing during debugging |
| **Partial bundle splitting on Pickup overflow** | Character frequently encounters bundles too large to merge (e.g., carrying 4, finds bundle of 4); Gather or fence procurement feels inefficient because useful bundles are skipped; Want "take what I can carry, leave the rest" behavior |
| **Typed nil-intent signals for ordered actions** | Intent finder nil return does double duty: "temporarily blocked" (worker collision) vs "nothing left to do" (true exhaustion). Transient nil guard can't distinguish them. Currently safe because feasibility narrows the cases, but risks masking permanent blockers (e.g., walled-off tile unreachable by any path). Trigger: a character idles on a feasible order where the blocker is permanent, not transient; or adding a new ordered action where the distinction matters for correctness. Fix: intent finders return a typed result (intent, nil-transient, nil-exhausted) instead of bare nil. |
| **architecture.md split**                        | architecture.md exceeds ~1000 lines; A section needs rewriting and the monolith makes it hard to scope; Staleness found because section drifted from reality unnoticed. Split: design model + key patterns ("why") stays in architecture.md; terrain/items/memory/action system reference ("what") → systems-reference.md; "Adding New X" checklists ("how") → checklists.md. |

### Future Enhancement Details

**Balance tuning candidates** (see config.go for current values):

- Prefernce formation still feels too easy
- Activity durations
- Satisfaction cooldown
- Sleep wake thresholds
- Movement energy drain
- Health tier thresholds
- Mood adjustment rates and modifiers
- Nut ground spawn rate — currently 600s shared with sticks/shells. Nuts are Snack-tier (10 pts), contributing ~2% of total food economy. Deferred from Slice 9 Step 4 because Step 4 doesn't change nut mechanics and the contribution is negligible. Revisit if nuts gain a gameplay role beyond supplemental forage filler, or if per-type ground spawn intervals are needed for other reasons (convert `GroundSpawnInterval` to a per-type map, same pattern as `ItemLifecycle`).
- Growth multiplier values (TilledGrowthMultiplier, WetGrowthMultiplier) — currently 1.25 each (1.56x combined). Evaluate after Slice 9 Step 4 playtesting whether the garden bonus feels rewarding with the new longer maturation/reproduction durations.

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

`continueIntent` (movement.go) has action-specific early-return blocks for actions whose `TargetItem` can be in different locations across phases (ground vs inventory). Current blocks: `ActionConsume`, `ActionDrink` (carried vessel), `ActionFillVessel`, `ActionWaterGarden`, `ActionHelpFeed`, `ActionHelpWater`. These exist because the generic path's `ItemAt` check would nil the intent when the item moves to inventory.

The blocks share a common shape: check if item is in inventory → return unchanged (no movement needed) or recalculate toward `Dest`. If this grows to 5+ blocks, evaluate whether a general "multi-phase item tracking" pattern could replace per-action blocks — e.g., a helper that checks inventory-then-map for `TargetItem` and recalculates accordingly, called by each action's block or replacing them entirely.

Don't consolidate prematurely — each block has slight phase-detection differences. The trigger is when adding a new block feels like copy-paste. With 6 blocks now present, this trigger is met — evaluate on the next multi-phase action addition.

---

**Smarter displacement direction:**

In crowded worlds (16 characters), the 3-step perpendicular displacement sometimes picks the wrong perpendicular direction — the one that leads back toward the blocker rather than away. This happens because the character may be on an alternating greedy step that makes one perpendicular look clear but ultimately unhelpful. The character then thrashes briefly (a few ticks) before resolving. Sticky BFS doesn't help here because the issue is displacement direction choice, not greedy-vs-BFS oscillation.

Possible approaches:
- Bias displacement toward the BFS-recommended direction (the direction BFS would have gone)
- Try both perpendiculars and pick the one that makes more progress toward the destination
- Increase displacement steps (currently 3) in proportion to how many blockers are nearby

Deferred because: The thrashing is brief and self-resolving (not infinite like the pre-sticky-BFS pond oscillation). Only noticeable in crowded worlds. Sticky BFS handles the primary pathfinding oscillation case.

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

**Snacking threshold:**

Characters could be willing to snack at a lower hunger level (e.g., Mild tier) if food is immediately adjacent or in inventory — "grab a berry while walking past" behavior. Currently all food triggers foraging at the same hunger threshold regardless of convenience. This parallels the vessel drinking threshold concept (Slice 9 Step 5): lower the trigger when the cost is near-zero.

Satiation-aware foraging targeting (first half of original item) extracted to [cleanup-phase-plan.md](cleanup-phase-plan.md) 3C.

Deferred because: Need to observe how the new satiation tiers play out before adding complexity. The current system may produce acceptable behavior through existing proximity scoring — characters naturally eat nearby food first.

---

**Character event/signal system:**

Currently, the helper "shout" that alerts a needy character to dropped food is implemented as a direct `needer.Intent = nil` — simple and effective, but the signaling is invisible to the code structure. A formal event system would allow:
- Characters to react to signals with richer behavior (gratitude expression, relationship score changes)
- Multiple signal types beyond "re-evaluate your needs" (warning, sharing, teaching)
- Observable event flow for debugging and action log richness

The current approach (direct intent clearing + action log entry) works well for the helping use case. Formalize when character relationships are introduced and signals need to carry metadata (who helped, what they provided, how the recipient feels about it).

---

**Vessel-excluded vs bundleable split:** ✓ Resolved in Step 4 (Construction phase, DD-20). Split into `VesselExcludedTypes` and `MaxBundleSize`. Triggered by brick and clay needing vessel exclusion without bundling.

---

**UI color style map:** ✓ Resolved in Step 5a (Construction phase). Extracted `colorToStyle(color types.Color) lipgloss.Style` helper in `renderCell()` (`ui/view.go`), shared by both item and construct rendering. Adding a new color now requires one case in `colorToStyle` — no per-caller switch duplication. Triggered by construct rendering needing the same color-to-style resolution as items.

---

**Item property registry migration:**

`config.go` currently mixes two concerns: tunable values (rates, durations, thresholds) and structural item-type registries (which items have which capabilities). The tunable values belong in config. The structural mappings — `StackSize`, `ItemMealSize`, `ItemLifecycle`, `SproutDurationTier`, `GroundSpawnCount`, `MaxBundleSize` — are game rules, not balance knobs.

Long-term design: tunable values stay in config as named constants (e.g., `DefaultMaxBundleSize = 6`, `BerryStackSize = 20`). Structural mappings move to entity as registries that reference those constants (e.g., `MaxBundleSize = map[string]int{"stick": config.DefaultMaxBundleSize}`). This separates "what can I tune?" from "which items have which properties?"

Deferred because: the current pattern works and all existing code references config. Migrate when adding a new item type and the config-vs-registry confusion causes a real problem.

---

**Preference formation for beverages (evaluate during Gardening):**

Gardening's Fetch Water activity introduces water-filled vessels as drinkable items. This technically triggers "Other beverages (non-spring) introduced."

However, water vessels may not warrant full preference formation since:
- Water is water (no variety like berry colors)
- The vessel itself already has preference-forming attributes

Evaluate whether characters should form preferences about "drinking from vessel" vs "drinking from spring" or defer until actual beverage variety exists (juice, tea, etc.).

---

**Harvest PickupToInventory handler clarity:**

In `applyPickup` (apply_actions.go), the PickupToInventory handler for harvest orders continues work when `char.GetCarriedVessel() == nil`. This check's original intent was "did I just pick up a vessel prerequisite (don't continue harvesting) or actual harvest work (continue)?" For vessel-excluded types like grass, the check accidentally produces correct behavior: grass never has a vessel, so "no vessel" always means "this was actual work — continue." The coincidence is that the original two cases (picked up vessel = don't continue, picked up harvest target without vessel = continue) and the new case (picked up bundle-target that never uses vessels = continue) share the same condition.

Rework when: a second vessel-excluded harvestable type arrives and the shared condition becomes non-obvious, or when debugging this handler and the intent/coincidence distinction wastes time. The fix would be to make the check explicit: `isVesselExcluded(targetType) || carriedVessel == nil`.

