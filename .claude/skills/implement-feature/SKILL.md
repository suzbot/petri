---
name: implement-feature
description: "Implement an already-planned feature, step, or task. Validates the plan against architecture patterns, then executes with TDD and human testing checkpoints. Trigger phrases: 'implement the next step', 'build the next feature', 'start development' or similar."
user-invocable: true
argument-hint: Feature or step to implement (e.g., "Slice 6 Step 3" or "sprout maturation")
model: sonnet
---

## Implementing a Planned Feature

**Do NOT enter plan mode.** The plan doc in `docs/` is the source of truth — it was produced by `/refine-feature` or `/new-phase` and contains the design decisions, pattern references, and step breakdown.

### Step 1: Load the Plan
- Read the **phase plan document** in `docs/` — find the step(s) to implement
- Read the **relevant sections of `docs/architecture.md`** that the plan references
- Read **`docs/Values.md`**
- Do NOT re-read the full requirements doc unless the plan explicitly flags an open question

### Step 2: Validate Readiness (REQUIRED)
**Do NOT write code yet.** Confirm the step is implementation-ready — check for presence and alignment, not correctness of decisions already made.

**Plan completeness — does the step have:**
- [ ] Anchor story (1-2 sentence narrative of what the user/character experiences — this is what you derive tests from)
- [ ] Detailed implementation breakdown (not "TBD" or single-line bullets)
- [ ] Architecture patterns named explicitly (e.g., "follows ordered action pattern" not just "follows existing pattern")
- [ ] Tests listed before implementation tasks (TDD order)
- [ ] [TEST] checkpoint (or explicit note why human testing isn't possible)
- [ ] [RETRO] checkpoint

**Pattern alignment:**
- Cite the architecture.md section and state how it applies — don't just assert "follows existing pattern"
- Verify patterns match current code (patterns may have evolved since planning)
- When the plan changes a constant or behavior, grep all usage sites and verify the plan addresses each one

**If any box is unchecked or alignment issues found:** invoke `/refine-feature`. After returning, re-invoke `/implement-feature` from the top — do not resume mid-skill.

**Wait for explicit user confirmation before writing any code.**

### Step 3: Implement (TDD)

**Core loop:**
1. **Announce** what you're about to write — one sentence: "I'm about to write [these tests] and [this implementation]"
2. **Write tests first** — anchor to the step's anchor story, not implementation paths. "Ground vessel ends up filled with water" validates intent; "returns ActionPickup" validates structure.
3. **Implement** minimum code to pass tests
4. **Verify** — run tests, run `gofmt ./...`
5. **Pause at each [TEST] checkpoint** for user to rebuild and test
6. After [TEST] passes, follow [DOCS] and [RETRO] checkpoints in the plan

**When to stop coding and invoke `/refine-feature`:**
- You're proposing design alternatives, not just implementation details
- You're re-deriving an approach you already considered (first: re-read architecture.md; second: `/refine-feature`; if circling on a test failure: run a diagnostic instead)
- You've been stuck for several minutes without progress — surface what's blocking you instead of churning

#### Test Patterns Reference

- **No brittle string assertions** — don't assert on exact log/activity wording. Remove existing brittle assertions rather than updating them.
- **Ordered-action integration tests:** Test loop must mirror `continueIntent`: (1) recalculate `char.Intent.Target` each tick via `NextStepBFS`, (2) rebuild intent when nil. `IsWet()` uses 8-directional adjacency — dry tiles must be >1 tile from water.
- **Flow-level anchor tests for procurement chains:** Chain system functions in handler order: `findXxxIntent` → `Pickup` → `FindNextTarget` → repeat → nil. See `TestGatherOrder_VesselPath_EndToEnd`.
- **`continueIntent` and TargetItem rules:** Read the "`continueIntent` Rules" and "Self-Managing Actions" sections in architecture.md when adding/modifying item-targeting actions. Trace the full fall-through path for new cases.

### Step 4: Human Testing ([TEST])
**Do NOT mark feature complete until user has tested.**
- Offer `/test-world` if the [TEST] checkpoint calls for it or the scenario is complex
- Wait for explicit confirmation from user before continuing

**When testing surfaces issues:**

*Gap (missing behavior, not a bug):*
- **Enumerate all sibling flows with the same structure** and ask "does this gap also exist in X, Y, Z?" Finding one is not enough.
- Discuss scope before writing code (see Values.md: "Pause Before Solving")

*Bug:*
- **Evidence first** — ask what the user observes, check logs, add `t.Logf` or `-v`. Don't propose fixes from speculation.
- **Restate the user's observation in their words** before offering a causal theory. If observation and theory don't match, ask rather than reframe.
- **Second bug in same feature** → stop patching. Restate the intended end-to-end flow and evaluate whether the design is sound (Values.md: "Step Back on Cascading Bugs").

### Step 5: Update Documentation ([DOCS])
Only after human testing confirms success:
- Run /update-docs via the **Task tool** (general-purpose subagent, sonnet model). Read `.claude/skills/update-docs/SKILL.md` and pass its full instructions + arguments as the task prompt.
- Mark feature complete in phase plan

### Step 6: Retro ([RETRO])
After documentation is updated, run /retro

---

**Communication rules:**
- When explaining changes: name both the mechanism and the visible artifact. "I added the shout" is ambiguous — say "The mechanism is X. It's visible as Y."
