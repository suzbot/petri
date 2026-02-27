# Architecture Requirements

What we want our architecture to represent, what we require of it, and how we prioritize getting there.

---

## What We're Modeling

Petri simulates autonomous characters surviving in a world. Each character has a body, a mind, and capabilities. The architecture should reflect this.

### The Character's Cognitive Loop

A character, each tick, does something like this:

1. **What do I need?** — My body tells me I'm thirsty, hungry, tired. How urgent is it?
2. **What are my options?** — I can see water over there, there's food nearby, I could ask for help.
3. **Which option do I pick?** — I'm very thirsty and the water is close. I'll drink.
4. **How do I get there?** — There's a rock in the way, I'll path around it.
5. **Do the thing.** — Walk, pick up, eat, drink, craft, talk.

Additionally, between ticks, the world acts on the character:

- **My body changes.** — Hunger grows, energy drains, poison damages, sleep restores.

### External Direction

Characters can also receive orders from the player. An order doesn't replace the cognitive loop — it inserts into it. A character with an order still gets hungry, still evaluates urgency, and may pause the order to eat. The order influences step 2 (adds an option) and step 3 (prioritizes it when needs aren't pressing).

### Social Awareness

Characters notice when others are in crisis and may choose to help. This is a motivation that competes with idle activities — it's part of "what are my options?" when a character has no pressing needs of their own.

---

## Architectural Principles

### Isomorphism

The code structure should mirror what it represents. If the game has a concept, there should be a clear place in the code where that concept lives. Someone asking "how does a character decide what to eat?" should find the answer in a place that makes intuitive sense.

### Separation of Concerns by Cognitive Role

The character's cognitive loop suggests natural boundaries:

| Cognitive Step | Responsibility | Current Location |
|---|---|---|
| **Motivation** — what do I need, how urgently? | Urgency tiers, stat priority, frustration | movement.go (CalculateIntent) |
| **Evaluation** — what options exist for each need? | Food scoring, drink finding, sleep finding | movement.go (findFoodIntent, findDrinkIntent, findSleepIntent) |
| **Selection** — which option wins? | Priority ordering, idle roll, order priority | movement.go + idle.go + order_execution.go |
| **Spatial planning** — how do I get there? | Pathfinding, obstacle avoidance, displacement | movement.go (NextStepBFS, continueIntent) |
| **Execution** — do the thing | Consume, pickup, craft, talk, till, plant | update.go (applyIntent) |
| **Body** — passive changes | Stat decay, sleep/wake, damage | survival.go |

### Encapsulated Workflows

Some actions involve multi-phase workflows (fetch water: get vessel → fill → done). These workflows encapsulate sub-decisions within a single course of action. This is intentional — once you've committed to fetching water, the sub-decisions (which vessel, which water source) are implementation details of that commitment, not top-level cognitive choices. The architecture should preserve this encapsulation.

### Intent as the Universal Contract

The Intent struct is the handoff between deciding and doing. This works well and should remain the single interface between the decision phase and execution phase.

---

## Current Gaps

### 1. movement.go Conflates Motivation, Evaluation, and Spatial Planning

**The problem:** movement.go houses three distinct responsibilities:
- The decision brain (CalculateIntent — motivation + selection)
- Need-specific evaluators (findFoodIntent, findDrinkIntent, findSleepIntent — option evaluation)
- The spatial engine (NextStepBFS, continueIntent, pathfinding — getting there physically)

**Why it matters:** "I'm hungry so I should find food" and "here's the BFS path around that rock" are different concerns. A new need type or a new food-scoring factor both require editing the same large file. Someone looking for "how does food selection work" finds it in the pathfinding file, not alongside other food logic. The idle path already delegates cleanly (idle.go → foraging.go for food scoring) — the need-driven path should follow the same pattern.

**What the game world suggests:** Motivation, evaluation, and locomotion are separate faculties. A character knows they're thirsty (motivation) separately from knowing where water is (evaluation) separately from knowing how to walk there (locomotion). Separating these would also naturally resolve the question of where need-driven evaluation lives — it would move to the domain-specific evaluator files where it conceptually belongs.

### 2. applyIntent May Outgrow Its Current Shape

**The problem:** update.go's applyIntent function handles execution for every ActionType. Each new action adds another branch.

**Current assessment:** At current scale this is manageable and grows linearly. It's worth revisiting after the movement.go separation — the right structure for execution dispatch may become clearer once the decision side is properly organized. Not a priority now.

---

## Prioritization

### Priority 1: Extract intent.go from movement.go

A pure extraction — no logic changes, no merging with other files. See [proposed-decision-flow.md](proposed-decision-flow.md) for the diagram.

**intent.go (new)** gets everything about deciding what to do:
- CalculateIntent, continueIntent (orchestration)
- findFoodIntent, FindFoodTarget, FoodTargetResult, findDrinkIntent, findHealingIntent, findSleepIntent (need-driven evaluators)
- findLookIntent (idle evaluator)
- canFulfillThirst, canFulfillHunger, canFulfillEnergy, canFulfillHealth (feasibility checks)
- findCarriedDrinkIntent, findCarriedFoodIntent (Mild-tier inventory checks)
- Small helpers: findNearestItem, findNearestItemExcluding, vesselHasLiquid, waterSourceName, eatingActivityName

**movement.go (slimmed)** keeps only spatial mechanics:
- NextStepBFS, nextStepBFSCore, NextStep (pathfinding)
- FindClosestCardinalTile, findClosestAdjacentTile (tile queries)
- isAdjacent, isCardinallyAdjacent (proximity checks)

**Everything else untouched.** foraging.go, fetch_water.go, idle.go, helping.go, order_execution.go stay exactly as they are.

#### Implementation Plan

**Step 1: Extract intent.go from movement.go**

*Anchor story:* A developer asks "where does food selection logic live?" After this step, they find it in intent.go alongside all other decision-making, not buried in the pathfinding file. The game itself behaves identically — no player-visible change.

- Create `internal/system/intent.go` with all decision functions listed above
  - Imports: `fmt`, `petri/internal/config`, `petri/internal/entity`, `petri/internal/game`, `petri/internal/types`
  - Pure cut-and-paste from movement.go, zero logic changes
- Slim `internal/system/movement.go` to only spatial functions listed above
  - Imports slim to: `petri/internal/game`, `petri/internal/types`
- Same `package system` — cross-file calls (intent.go calling `NextStepBFS`, etc.) work with no changes
- No test files to reorganize (no movement_test.go exists)
- Verify: `go build ./...`, `go test ./...`, `gofmt ./...`
- Architecture pattern: Separation of Concerns by Cognitive Role (architecture-requirements.md). Intent = motivation + evaluation + selection. Movement = spatial planning + locomotion.
- Values: Follow the Existing Shape — making file structure match conceptual structure that already exists in the code.
- [TEST] Human spot-check: characters eat, drink, sleep, path around obstacles, idle activities work, orders work, helping works. No behavior change expected.

**Step 2: Update docs**

- `docs/flow-diagrams.md`: Update "Decision Layer" boxes to show intent.go and movement.go as separate concerns
- `docs/architecture.md`: Update file references (CalculateIntent location, continueIntent rules header, Adding New Actions checklists)
- `docs/architecture-requirements.md`: Mark Priority 1 as complete
- `CLAUDE.md`: Update Key Files list (add intent.go, update movement.go description)
- [DOCS] via `/update-docs`
- [RETRO] via `/retro`

### Priority 2: Re-evaluate After Separation

Once intent.go and movement.go are separated, reassess what else needs attention. The right next step may become obvious — whether that's migrating need-driven evaluators to domain files, restructuring applyIntent, or something we haven't identified yet.

---

## Decisions

- **Intent and movement are separate concerns.** Intent is about deciding what to do (motivation, evaluation, selection). Movement is about getting there (pathfinding, spatial queries). They belong in separate files.
- **This is a minimal extraction, not a reorganization.** Need-driven evaluators (findFoodIntent, etc.) move to intent.go as a unit. Whether they eventually migrate further to domain-specific files is a future conversation.
- **continueIntent stays with CalculateIntent.** Both are intent orchestration. If the phase-detection vs. path-recalculation coupling ever causes friction, revisit then.

## Future Exploration (Not Blocking Priority 1)

- **Should need-driven evaluators migrate to domain files?** e.g., findFoodIntent → foraging.go, findDrinkIntent → fetch_water.go. Need-driven, idle, and ordered behaviors are distinct flavors of thinking — any consolidation should respect that distinction rather than blur it.
- **Should need-driven and idle evaluation share code or share a pattern?** Foraging and findFoodIntent score food differently on purpose (urgency affects willingness). Whether that becomes a parameterized shared evaluator is worth exploring after we see the post-refactor shape.
- **applyIntent structure.** Worth revisiting once the decision side is properly organized.
