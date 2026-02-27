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
| **Motivation** — what do I need, how urgently? | Urgency tiers, stat priority, frustration | intent.go (CalculateIntent) |
| **Evaluation** — what options exist for each need? | Food scoring, drink finding, sleep finding | intent.go (findFoodIntent, findDrinkIntent, findSleepIntent) |
| **Selection** — which option wins? | Priority ordering, idle roll, order priority | intent.go + idle.go + order_execution.go |
| **Spatial planning** — how do I get there? | Pathfinding, obstacle avoidance, displacement | movement.go (NextStepBFS) |
| **Execution** — do the thing | Consume, pickup, craft, talk, till, plant | update.go (applyIntent) |
| **Body** — passive changes | Stat decay, sleep/wake, damage | survival.go |

### Encapsulated Workflows

Some actions involve multi-phase workflows (fetch water: get vessel → fill → done). These workflows encapsulate sub-decisions within a single course of action. This is intentional — once you've committed to fetching water, the sub-decisions (which vessel, which water source) are implementation details of that commitment, not top-level cognitive choices. The architecture should preserve this encapsulation.

### Intent as the Universal Contract

The Intent struct is the handoff between deciding and doing. This works well and should remain the single interface between the decision phase and execution phase.

---

## Current Gaps

### ✅ 1. movement.go Conflates Motivation, Evaluation, and Spatial Planning (Resolved)

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

### ✅ Priority 1: Extract intent.go from movement.go (Complete)

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

**✅ Step 1: Extract intent.go from movement.go**

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

**✅ Step 2: Update docs**

- `docs/flow-diagrams.md`: Update "Decision Layer" boxes to show intent.go and movement.go as separate concerns
- `docs/architecture.md`: Update file references (CalculateIntent location, continueIntent rules header, Adding New Actions checklists)
- `docs/architecture-requirements.md`: Mark Priority 1 as complete
- `CLAUDE.md`: Update Key Files list (add intent.go, update movement.go description)
- [DOCS] via `/update-docs`
- [RETRO] via `/retro`

### Priority 2: Align Code to the Four-Bucket Decision Model

The character's decision-making has four distinct motivational categories. The code should reflect these as separate concerns rather than nesting three of them inside a function named for the fourth.

#### The Four Buckets

| Bucket | Question the character asks | Current location |
|---|---|---|
| **Needs** | What is my body demanding? | intent.go — priority loop, findFoodIntent, findDrinkIntent, etc. |
| **Orders** | Has the player told me to do something? | order_execution.go — but called from idle.go |
| **Helping** | Is someone nearby in crisis? | helping.go — but called from idle.go |
| **Discretionary** | What do I feel like doing? | idle.go random roll — look, talk, forage, fetch water |

#### What's Wrong Today

**`selectIdleActivity` conflates buckets 2–4.** It evaluates orders first, then helping, then the leisure roll — all under a function named "idle." This hides the fact that order work and crisis response are distinct motivational categories from leisure. The orchestrator (`CalculateIntent`) calls `selectIdleActivity` and gets back an intent without knowing whether the character decided to help a dying neighbor or go look at a rock.

**"Idle" means two different things.** The code uses "Idle" for both:
- **Discretionary** — all stats are fine, character is choosing a leisure activity. Content.
- **Stuck** — character has needs but can't fulfill them (no food, no water, no bed). Helpless.

These are fundamentally different states. A content character proactively chooses activities. A stuck character is waiting for the world to change. The code routes both through the same function and labels both "Idle."

**The orchestrator has need-evaluation interleaved with routing logic.** `CalculateIntent` contains both the top-level "which bucket do I evaluate?" logic and the need-specific evaluators (findFoodIntent, findDrinkIntent, etc.). This makes intent.go ~1500 lines and harder to read as an orchestrator. Lower priority than the idle.go problem since needs and orchestration are at least in the same domain.

#### Priority Ordering Between Buckets

The buckets aren't a flat list — there's intentional priority logic:

1. **Continue current activity** if still valid (any bucket)
2. **Needs** always win at Moderate+ urgency (interrupt everything)
3. **Mild needs + assigned order** — eat from inventory, keep working (needs/orders hybrid)
4. **No pressing needs** — evaluate Orders → Helping → Discretionary in that priority
5. **Needs unfulfillable** — also evaluate Orders → Helping → Discretionary (the "stuck" path)

This priority ordering should be explicit in the orchestrator, not buried inside `selectIdleActivity`.

#### What Changes

- `selectIdleActivity` breaks apart: order evaluation, helping evaluation, and discretionary activity selection become separate calls from the orchestrator
- `idle.go` becomes `discretionary.go` and contains only the leisure activity roll
- "Idle (no options)" becomes "Stuck"; "Idle (no needs)" stays "Idle" (no longer overloaded once stuck is separated)
- The orchestrator in `CalculateIntent` gains explicit bucket-routing that reads like the priority list above

#### What Stays the Same

- Need-driven and discretionary evaluators for the same resource remain separate on purpose (findFoodIntent vs findForageIntent, findDrinkIntent vs findFetchWaterIntent) — urgency affects willingness, and these are different motivations producing different scoring
- `continueIntent` stays with `CalculateIntent` — both are orchestration
- The Intent struct remains the universal contract between deciding and doing
- Need evaluators stay in intent.go for now (separating them from the orchestrator is a future conversation)

#### Implementation Plan

**Step 1: Extract bucket functions from idle.go**

*Anchor story:* A developer asks "where does crisis-helping logic get triggered?" After this step, they find `selectHelpingActivity` in helping.go and see it called explicitly from CalculateIntent — not buried inside an "idle" function alongside random leisure rolls. A developer asks "where is the discretionary activity roll?" and finds it in discretionary.go, clearly separated from order work and helping.

This step is a pure extraction — no logic changes except cooldown scoping (see below). Game behavior is identical.

- Rename `idle.go` → `discretionary.go`
- Rename `selectIdleActivity` → `selectDiscretionaryActivity` — keeps only the cooldown check + random roll (look/talk/forage/fetchWater/stay put). Removes order-checking code (lines 18-34) and helping code (lines 38-55).
- Rename `isIdleAction` → `isDiscretionaryAction` — same logic, name aligns with bucket terminology.
- Add `selectHelpingActivity(char, pos, items, gameMap, log) *entity.Intent` to helping.go — wraps the crisis detection + help intent logic extracted from selectIdleActivity. No cooldown.
- `selectOrderActivity` stays in order_execution.go unchanged — it's already a standalone function, just called from a different place now.
- **Cooldown scoping change:** Move `IdleCooldown` check and set inside `selectDiscretionaryActivity` only. Orders and helping become no-cooldown (checked every tick from the orchestrator). This is better behavior: crisis response is faster, order pickup is immediate. Cost: two extra O(n) scans per tick (characters + items), negligible at current scale.
- Update all call sites: `selectIdleActivity` → the three separate calls happen in Step 2, but test files referencing `selectIdleActivity` by name need updating to `selectDiscretionaryActivity`.
- Update test files: `talking_test.go` calls `selectIdleActivity` directly in 5 places — rename to `selectDiscretionaryActivity`. `order_execution_test.go` has comments referencing `selectIdleActivity` — update comments.
- Architecture pattern: Separation of Concerns by Cognitive Role (architecture-requirements.md). Each bucket gets its own file: needs in intent.go, orders in order_execution.go, helping in helping.go, discretionary in discretionary.go.
- Values: Follow the Existing Shape — making file structure match the four-bucket conceptual structure already present in the code.
- Verify: `go build ./...`, `go test ./...`, `gofmt ./...`
- [TEST] Human spot-check: characters eat, drink, sleep, take orders, help crisis characters, do discretionary activities (look, talk, forage, fetch water). No behavior change expected. Verify that helping and order-taking still happen (cooldown scoping change means they should be at least as responsive as before).

**Step 2: Restructure CalculateIntent's bucket routing**

*Anchor story:* A developer reads CalculateIntent and can follow the character's decision process as a clear priority list: "continue what I'm doing, handle urgent needs, handle mild needs while working, then evaluate orders → helping → discretionary." The three `selectIdleActivity` call sites collapse into one unified bucket-routing section that serves both the "no pressing needs" and "stuck" paths.

This step rewrites the orchestrator to call each bucket explicitly. Same intents produced — the routing is equivalent, just structured to match the documented priority ordering.

- Replace the three `selectIdleActivity(...)` call sites in CalculateIntent (lines 186, 222, 304) with explicit bucket routing:
  ```
  1. Early exits (dead/sleeping/frustrated)
  2. Continue current activity if valid (any bucket, unchanged)
  3. New intent from scratch:
     a. Moderate+ needs → priority loop → return if found (pause order if needed)
     b. Mild needs + assigned order → carried food/drink intercept → fall through to 3e
     c. Mild needs + no order → priority loop → return if found
     d. Frustration tracking at Severe+ (if priority loop failed in 3a)
     e. Bucket routing: selectOrderActivity → selectHelpingActivity → selectDiscretionaryActivity
     f. If nothing: maxTier > TierNone → "Stuck" / maxTier == TierNone → "Idle"
  ```
- The "continue current activity" block (lines 41-86) stays unchanged but update comment: "idle activity" → "non-need activity" since it continues orders and helping too, not just discretionary.
- The need-driven continuation block (lines 89-163) stays unchanged.
- Architecture pattern: This makes the orchestrator read like the Priority Ordering table above — explicit bucket evaluation, not implicit delegation.
- Values: Anchor to Intent — the code structure mirrors how a character thinks: "Do I have urgent needs? Am I working on something? Is someone in crisis? What do I feel like doing?"
- Anti-pattern to avoid: Do NOT change the need evaluation logic (priority loop, tier sorting, stat fallback). Only the routing around it changes.
- Verify: `go build ./...`, `go test ./...`, `gofmt ./...`
- [TEST] Human testing — thorough check since this rewrites the orchestrator:
  - Characters with Moderate+ needs eat/drink/sleep (needs bucket)
  - Characters with assigned orders resume them at TierNone/TierMild (orders bucket)
  - Characters notice crisis neighbors and deliver food/water (helping bucket)
  - Characters look, talk, forage, fetch water when content (discretionary bucket)
  - Characters with Mild hunger + no order seek food (Mild needs through priority loop)
  - Characters with Mild hunger + assigned order eat from inventory then resume (Mild intercept)
  - Characters that can't meet needs show "Stuck" label, not "Idle"
- [RETRO]

**Step 3: Update activity labels**

*Anchor story:* A player watching a character who has Severe thirst but no water source sees "Stuck (can't meet needs)" instead of "Idle (no options)" — the label tells them the character is helpless, not content. In the action log, "No water source available" replaces "Idle (no water source)" — the message describes what happened, not a misleading state.

- **Final state labels in CalculateIntent:**
  - "Idle (no needs)" → "Idle" (unchanged meaning, but shorter — the parenthetical was disambiguating from the other "Idle" which is now "Stuck")
  - "Idle (no options)" → "Stuck (can't meet needs)"
- **Intermediate log messages in need evaluators** (these flash during evaluation and get overwritten, but persist in the action log):
  - findDrinkIntent: "Idle (no water source)" → log message "No water source available", CurrentActivity stays as whatever it was (don't set to "Idle")
  - findFoodIntent: "Idle (no suitable food)" → log message "No suitable food available", don't set CurrentActivity
  - findSleepIntent: "Idle (no bed nearby)" → log message "No bed available", don't set CurrentActivity
  - findHealingIntent: "Idle (no known healing items)" → log message "No known healing items available", don't set CurrentActivity
- These evaluators currently set `char.CurrentActivity = "Idle"` as a side effect before returning nil. Since the orchestrator now sets the final label ("Idle" or "Stuck") after all evaluators fail, the evaluators should stop setting CurrentActivity — it gets overwritten anyway and the intermediate value was misleading.
- Values: Anchor to Intent — labels describe what the character is experiencing (stuck vs content), not code state.
- Verify: `go build ./...`, `go test ./...`, `gofmt ./...`
- [TEST] Human spot-check: verify "Stuck" label appears for characters with unmet needs, "Idle" for content characters, and action log messages read naturally.
- [DOCS] via `/update-docs`. Additionally, update diagrams:
  - `docs/flow-diagrams.md`: "Complete Call Graph" — update Decision Layer subgraph: `discretionary.go` replaces `idle.go`, `intent.go` calls orders/helping/discretionary directly instead of idle dispatching to them. "Decision Flow Only" — same restructure: remove idle.go as intermediary, show CalculateIntent calling each bucket. "Intent Priority Hierarchy" — update Tier 0 and fallback paths to show explicit bucket routing (Orders → Helping → Discretionary) instead of `selectIdleActivity()`. "Idle Activity Selection" — rename to "Discretionary Activity Selection", remove order and helping sections, show only cooldown + random roll.
  - `docs/proposed-decision-flow.md`: Remove — transitional artifact from Priority 1, superseded by the updated flow-diagrams.md.
- [RETRO] via `/retro`

---

## Decisions

- **Intent and movement are separate concerns.** Intent is about deciding what to do (motivation, evaluation, selection). Movement is about getting there (pathfinding, spatial queries). They belong in separate files.
- **This is a minimal extraction, not a reorganization.** Need-driven evaluators (findFoodIntent, etc.) move to intent.go as a unit. Whether they eventually migrate further to domain-specific files is a future conversation.
- **continueIntent stays with CalculateIntent.** Both are intent orchestration. If the phase-detection vs. path-recalculation coupling ever causes friction, revisit then.
- **Four decision buckets: Needs, Orders, Helping, Discretionary.** These are the character's motivational categories. The code structure should reflect them as separate concerns, not nest three inside the fourth.
- **"Stuck" replaces overloaded "idle."** When a character has needs but can't fulfill them, that's a "stuck" state, not an idle one. "Idle" was conflating contentment with helplessness.
- **Discretionary, not idle.** The leisure activity bucket (look, talk, forage, fetch water) is "discretionary" — chosen freely when nothing is pressing. "Idle" implied the character has nothing to do; "discretionary" implies they're choosing what to do.
- **Need-driven and discretionary paths for the same resource are separate motivations.** findFoodIntent and findForageIntent represent different reasons to seek food. Same for findDrinkIntent and findFetchWaterIntent. The paths should remain conceptually separate, but shared helpers and duplicate code between them can be evaluated further later.

## Future Exploration

- **Should need evaluators separate from the orchestrator?** intent.go currently houses both CalculateIntent (the cognitive loop) and the need-specific evaluators. Separating them would make the orchestrator easier to read but isn't blocking. Revisit once the four-bucket separation is done.
- **applyIntent structure.** Worth revisiting once the decision side is properly organized.
