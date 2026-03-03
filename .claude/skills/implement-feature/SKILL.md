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
- `docs/Values.md` and `docs/game-mechanics.md` are available if needed — consult Values.md when a gap requires framing a design discussion, game-mechanics.md when you need to understand expected player-visible behavior

### Step 2: Create Task List (REQUIRED)

**For each sub-step in the step spec**, create the following tasks in order. These are the workflow — do not skip tasks unless the user explicitly says to.

1. **Validate readiness** — confirm spec meets readiness criteria (see below)
2. **Invoke `/refine-feature`** — only if gaps found in readiness review; otherwise mark completed immediately
3. **Write tests** — TDD: tests before implementation
4. **Implement** — minimum code to pass tests
5. **Run tests and format** — `go test ./...` and `gofmt ./...`
6. **[TEST] Human testing** — offer `/test-world`, relay checklist, wait for confirmation
7. **Invoke `/fix-bug`** — only if issues found during testing; otherwise mark completed immediately
8. **[DOCS] Invoke `/update-docs`** — relay summary of changes to user
9. **[RETRO] Invoke `/retro`**

Set up dependencies so each task blocks the next. **Wait for user confirmation before moving past task 6.**

### Readiness Criteria (for task 1)

Confirm the sub-step spec has:
- [ ] Anchor story (1-2 sentence narrative of what the user/character experiences)
- [ ] Detailed implementation breakdown (not "TBD" or single-line bullets)
- [ ] Architecture patterns named explicitly (e.g., "follows ordered action pattern" not just "follows existing pattern")
- [ ] Tests listed before implementation tasks (TDD order)
- [ ] At least one test traces the anchor story end-to-end
- [ ] [TEST] checkpoint
- [ ] If step spec references an architecture.md "Adding New X" checklist, verify spec covers every item

**Pattern alignment:**
- Cite the architecture.md section and state how it applies
- Verify patterns match current code (patterns may have evolved since planning)
- When the step spec changes a constant or behavior, grep all usage sites and verify the spec addresses each one

**Wait for explicit user confirmation before writing any code.**

### Implementation Guidelines (for tasks 3-5)

**Announce** what you're about to write — one sentence: "I'm about to write [these tests] and [this implementation]"

**Only implement behavior specified in the step spec.** If a detail isn't specified, surface the gap, make a recommendation, and seek confirmation from the user.

**Write tests immediately before code** — anchor to the step's anchor story, not implementation paths. "Ground vessel ends up filled with water" validates intent; "returns ActionPickup" validates structure.

When modifying a shared function (Pickup, CanPickUpMore, FindNextTarget, etc.), trace all callers before writing code — new return values and behavior changes must be handled at every call site.

**When to stop coding and invoke `/refine-feature`:**
- You find a gap in the implementation plan
- You're proposing design alternatives, not just implementation details
- You're re-deriving an approach you already considered (first: re-read architecture.md; second: `/refine-feature`; if circling on a test failure: run a diagnostic instead)
- You've been stuck for several minutes without progress — surface what's blocking you

#### Test Patterns Reference

- **No brittle string assertions** — don't assert on exact display text: log messages, activity wording, item descriptions, names, or UI strings. Remove existing brittle assertions rather than updating them.
- **Ordered-action integration tests:** Test loop must mirror `continueIntent`: (1) recalculate `char.Intent.Target` each tick via `NextStepBFS`, (2) rebuild intent when nil. `IsWet()` uses 8-directional adjacency — dry tiles must be >1 tile from water.
- **Flow-level anchor tests for procurement chains:** Chain system functions in handler order: `findXxxIntent` → `Pickup` → `FindNextTarget` → repeat → nil. See `TestGatherOrder_VesselPath_EndToEnd`.
- **`continueIntent` and TargetItem rules:** Read the "`continueIntent` Rules" and "Self-Managing Actions" sections in architecture.md when adding/modifying item-targeting actions. Trace the full fall-through path for new cases.

### Human Testing Guidelines (for task 6)

- Before relaying [TEST] items to user: verify each item matches the step's implemented behavior. Surface contradictions rather than forwarding verbatim.
- Offer `/test-world` if the [TEST] checkpoint calls for it or the scenario is complex
- Wait for explicit confirmation from user before continuing

### Documentation Guidelines (for task 8)

- Invoke `/update-docs` via the **Skill tool** with a summary of what changed
- **After `/update-docs` completes: relay the summary of doc changes to the user.**
- Update the step's **Status** in the phase design doc to "Complete"
- Replace `docs/step-spec.md` content with: `Step N complete. Next: Step M — run /refine-feature.`
- Suggest a commit message for the completed work
