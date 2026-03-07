---
name: implement-feature
description: "Implement an already-planned feature, step, or task. Validates the plan against architecture patterns, then executes with TDD and human testing checkpoints. Trigger phrases: 'implement the next step', 'build the next feature', 'start development' or similar."
user-invocable: true
argument-hint: Feature or step to implement (e.g., "Step 3" or "clay terrain")
model: sonnet
---

## Implementing a Planned Feature

**Do NOT enter plan mode.** The step spec and design doc are the sources of truth — produced by `/refine-feature` or `/new-phase`.

### Step 1: Read Spec → Create Task List → Show User

This is the only step before work begins. Do all three in sequence, with no codebase reading in between. **No Grep or file reads until the task list is confirmed.**

1. **Read the plan**: Read `docs/step-spec.md` and the phase design doc (DD entries that affect this step only). Do not read `architecture.md` or any source files yet.
2. **Create the task list** based on the spec. For each sub-step, create these tasks in order:
   1. **Validate readiness** — see Readiness Criteria below
   2. **Invoke `/refine-feature`** — only if gaps found; otherwise mark completed
   3. **Write anchor test** — end-to-end functional test based on the anchor story, before any implementation
   4. **Write unit tests + Implement** — additional tests and minimum code to pass all tests
   5. **Run tests and format** — `go test ./...` and `gofmt ./...`
   6. **[TEST] Human testing** — offer `/test-world`, wait for confirmation
   7. **Invoke `/fix-bug`** — only if issues found; otherwise mark completed
   8. **[DOCS] Invoke `/update-docs`** — relay summary to user
   9. **[RETRO] Invoke `/retro`**
3. **Show the task list to the user** and wait for confirmation before beginning task 1.

Set up dependencies so each task blocks the next. Wait for user confirmation before moving past task 6.

---

### Reference: Readiness Criteria (for task 1)

**Context budget discipline — every file read consumes context. Protect it.**

- Use Grep, not Read, to locate functions. Find exact lines, then read only those offsets. Never read a full file >200 lines to find one function.
- Trust prior knowledge. If you've seen a pattern before, grep to confirm the signature, don't re-read the handler.
- Do not read `architecture.md` end-to-end — grep for specific sections by heading or keyword.

**Confirm the spec has:**
- [ ] Anchor story (1-2 sentence narrative of what the user/character experiences)
- [ ] Detailed implementation breakdown (not "TBD" or single-line bullets)
- [ ] Architecture patterns named explicitly (e.g., "follows ordered action pattern" not just "follows existing pattern")
- [ ] Tests listed before implementation tasks (TDD order)
- [ ] At least one test traces the anchor story end-to-end
- [ ] [TEST] checkpoint
- [ ] Identify which "Adding New X" checklists apply — grep `architecture.md` for the heading, read only that section, verify spec covers each checklist item

**Pattern alignment (targeted, not comprehensive):**
- Grep for key functions the spec references to confirm signatures and call sites — do not read the whole file
- When the spec changes a shared function, grep all callers to verify the spec addresses each one
- New visual indicators: grep for the analogous existing pattern; read only those lines

**Report the readiness result to the user before proceeding to task 3. Wait for explicit user confirmation before writing any code.**

---

### Reference: Implementation Guidelines (for tasks 3-4)

**The anchor test (task 3) must exist and fail before writing implementation code.** Additional unit tests for specific behaviors can be written alongside implementation in task 4.

**Announce** what you're about to write — one sentence: "I'm about to write [these tests] and [this implementation]"

**Only implement behavior specified in the step spec.** If a detail isn't specified, surface the gap, make a recommendation, and seek confirmation.

**Write tests immediately before code** — anchor to the step's anchor story, not implementation paths. "Ground vessel ends up filled with water" validates intent; "returns ActionPickup" validates structure.

When modifying a shared function, grep for callers before writing code — new return values must be handled at every call site. Read only the relevant lines, not the whole file.

**Before reading any source file:** grep for the specific function or type you need, then read just the relevant lines. Reading a few examples is enough — you don't need every instance.

**When to stop coding and invoke `/refine-feature`:**
- You find a gap in the implementation plan
- You're proposing design alternatives, not just implementation details
- You're re-deriving an approach you already considered (first: re-read architecture.md; second: `/refine-feature`; if circling on a test failure: run a diagnostic instead)
- You've been stuck for several minutes without progress — surface what's blocking you

#### Test Patterns Reference

- **No brittle string assertions** — don't assert on exact display text. Remove existing brittle assertions rather than updating them.
- **Ordered-action integration tests:** Test loop must mirror `continueIntent`: (1) recalculate `char.Intent.Target` each tick via `NextStepBFS`, (2) rebuild intent when nil. `IsWet()` uses 8-directional adjacency — dry tiles must be >1 tile from water.
- **Flow-level anchor tests for procurement chains:** Chain system functions in handler order: `findXxxIntent` → `Pickup` → `FindNextTarget` → repeat → nil.
- **Game-loop integration tests:** Call `CalculateIntent` every tick (not only when intent is nil) — the real loop runs it each tick for `continueIntent`.
- **`continueIntent` and TargetItem rules:** Read the "`continueIntent` Rules" and "Self-Managing Actions" sections in architecture.md when adding/modifying item-targeting actions.

---

### Reference: Human Testing (for task 6)

- Verify [TEST] items match implemented behavior before relaying. Surface contradictions.
- Offer `/test-world` if the checkpoint calls for it. **Before invoking, read the test-world skill's Caller Requirements and build a structured spec.**
- **When the user reports any issues: invoke `/fix-bug`.** Do not diagnose or propose fixes inline — the fix-bug skill's evidence-first protocol exists for a reason.
- Wait for explicit confirmation before continuing

---

### Reference: Documentation (for task 8)

- Invoke `/update-docs` via the **Skill tool** with a summary of what changed
- **After `/update-docs` completes: relay the summary of doc changes to the user.**
- Update the step's **Status** in the phase design doc to "Complete"
- Replace `docs/step-spec.md` content with: `Step N complete. Next: Step M — run /refine-feature.`
- Suggest a commit message

---

### Reference: Retro (for task 9)

1. Ask the user if they want conversation history searched for uncaptured friction
2. If yes: run the search process from the retro skill's "Conversation History Search" section, pass findings as `/retro` arguments
3. Invoke `/retro` via the Skill tool
4. Relay proposals to user for approval; implement only what approved
5. After proposals resolved: update `memory/last-retro.txt` with current ISO timestamp
