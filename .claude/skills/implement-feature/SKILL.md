---
name: implement-feature
description: "Implement an already-planned feature, step, or task. Validates the plan against architecture patterns, then executes with TDD and human testing checkpoints. Trigger phrases: 'implement the next step', 'build the next feature', 'start development' or similar."
user-invocable: true
argument-hint: Feature or step to implement (e.g., "Step 3" or "clay terrain")
model: sonnet
---

## Implementing a Planned Feature

**Do NOT enter plan mode.** The step spec and design doc are the sources of truth — produced by `/refine-feature` or `/new-phase`.

### Step 1: Load the Plan
- Read **`docs/step-spec.md`** — the implementation plan for the current step
- Read the **phase design doc** in `docs/` (linked at top of step spec) — for design decisions and broader context. Focus on the DD entries that affect this step.
- Read the **relevant sections of `docs/architecture.md`** that the step spec references
- Read **`docs/Values.md`**
- Do NOT re-read the full requirements doc unless the step spec explicitly flags an open question

### Step 2: Validate Readiness (REQUIRED)
**Do NOT write code yet.** Confirm the step is implementation-ready — check for presence and alignment, not correctness of decisions already made.

**Step spec completeness — does the step have:**
- [ ] Anchor story (1-2 sentence narrative of what the user/character experiences — this is what you derive tests from)
- [ ] Detailed implementation breakdown (not "TBD" or single-line bullets)
- [ ] Architecture patterns named explicitly (e.g., "follows ordered action pattern" not just "follows existing pattern")
- [ ] Tests listed before implementation tasks (TDD order)
- [ ] At least one test traces the anchor story end-to-end (not just unit-level checks of individual functions)
- [ ] [TEST] checkpoint (or explicit note why human testing isn't possible)
- [ ] [RETRO] checkpoint
- [ ] If step spec references an architecture.md "Adding New X" checklist, verify spec covers every item on that checklist

**Pattern alignment:**
- Cite the architecture.md section and state how it applies — don't just assert "follows existing pattern"
- Verify patterns match current code (patterns may have evolved since planning)
- When the step spec changes a constant or behavior, grep all usage sites and verify the spec addresses each one

**If any box is unchecked or alignment issues found:** invoke `/refine-feature`. After returning, re-invoke `/implement-feature` from the top — do not resume mid-skill.

**Wait for explicit user confirmation before writing any code.**

### Step 3: Implement (TDD)

**Core loop:**
1. **Announce** what you're about to write — one sentence: "I'm about to write [these tests] and [this implementation]"
2. **Only implement behavior specified in the step spec, which has been reconciled to requirements.** If a detail isn't specified, surface the gap, make a recommendation, and seek confirmation from the user.
3. **Write each sub-step's tests immediately before that sub-step's code** — do not batch tests for multiple sub-steps. Anchor to the step's anchor story, not implementation paths. "Ground vessel ends up filled with water" validates intent; "returns ActionPickup" validates structure. When modifying a shared function (Pickup, CanPickUpMore, FindNextTarget, etc.), trace all callers before writing code — new return values and behavior changes must be handled at every call site.
4. **Implement** minimum code to pass tests
5. **Verify** — run tests, run `gofmt ./...`
6. **Pause at each [TEST] checkpoint** for user to rebuild and test
7. After [TEST] passes, follow [DOCS] and [RETRO] checkpoints in the step spec

**When to stop coding and invoke `/refine-feature`:**
- You're proposing design alternatives, not just implementation details
- You're re-deriving an approach you already considered (first: re-read architecture.md; second: `/refine-feature`; if circling on a test failure: run a diagnostic instead)
- You've been stuck for several minutes without progress — surface what's blocking you instead of churning

#### Test Patterns Reference

- **No brittle string assertions** — don't assert on exact display text: log messages, activity wording, item descriptions, names, or UI strings. Remove existing brittle assertions rather than updating them.
- **Ordered-action integration tests:** Test loop must mirror `continueIntent`: (1) recalculate `char.Intent.Target` each tick via `NextStepBFS`, (2) rebuild intent when nil. `IsWet()` uses 8-directional adjacency — dry tiles must be >1 tile from water.
- **Flow-level anchor tests for procurement chains:** Chain system functions in handler order: `findXxxIntent` → `Pickup` → `FindNextTarget` → repeat → nil. See `TestGatherOrder_VesselPath_EndToEnd`.
- **`continueIntent` and TargetItem rules:** Read the "`continueIntent` Rules" and "Self-Managing Actions" sections in architecture.md when adding/modifying item-targeting actions. Trace the full fall-through path for new cases.

### Step 4: Human Testing ([TEST])
**Do NOT mark feature complete until user has tested.**
- Before relaying [TEST] items to user: verify each item matches the step's implemented behavior (thresholds, rules). Surface contradictions rather than forwarding verbatim.
- Offer `/test-world` if the [TEST] checkpoint calls for it or the scenario is complex
- Wait for explicit confirmation from user before continuing

**When testing surfaces issues:**

*Gap (missing behavior, not a bug):*
- **Enumerate all sibling flows with the same structure** and ask "does this gap also exist in X, Y, Z?" Finding one is not enough.
- Discuss scope before writing code (see Values.md: "Pause Before Solving")

*Bug:*
- **Always gather direct evidence first** — examine the most recently modified save file (`ls -t ~/.petri/worlds/*/state.json | head -1`), check logs, add `t.Logf` or `-v`. Do this before forming any hypothesis about the cause. Never guess what the game state is; read it.
- **Restate the user's observation in their words** before offering a causal theory. If evidence doesn't match the report, ask clarifying questions — don't assume the user is misreporting. Ambiguity in what they observed is more likely than a wrong report.
- **After fixing any human-caught bug:** write a regression test that reproduces the scenario before moving to the next testing round. Don't wait for the user to ask.
- **Second bug in same feature** → stop patching. Restate the intended end-to-end flow and evaluate whether the design is sound (Values.md: "Step Back on Cascading Bugs"). Also check: does an automated end-to-end test exist for this flow? If not, write one before fixing — it catches remaining bugs in the same pass.

### Step 5: Update Documentation ([DOCS])
Only after human testing confirms success:
- Launch an **Agent tool** subagent (general-purpose, sonnet model) to run update-docs autonomously. Read `.claude/skills/update-docs/SKILL.md` and pass its full instructions + the change summary as the agent prompt. Do NOT use the Skill tool — that requires per-edit approval.
- Update the step's **Status** in the phase design doc to "Complete"
- Replace `docs/step-spec.md` content with: `Step N complete. Next: Step M — run /refine-feature.`

### Step 6: Retro ([RETRO])
After documentation is updated, run /retro

---

**Communication rules:**
- When explaining changes: name both the mechanism and the visible artifact. "I added the shout" is ambiguous — say "The mechanism is X. It's visible as Y."
