---
name: implement-feature
description: "Implement an already-planned feature, step, or task. Validates the plan against architecture patterns, then executes with TDD and human testing checkpoints. Trigger phrases: 'implement the next step', 'build the next feature', 'start development' or similar."
user-invocable: true
argument-hint: Feature or step to implement (e.g., "Slice 6 Step 3" or "sprout maturation")
---

## Implementing a Planned Feature

**Do NOT enter plan mode.** The plan doc in `docs/` is the source of truth — it was produced by `/refine-feature` or `/new-phase` and contains the design decisions, pattern references, and step breakdown.

### Step 1: Load the Plan
- Read the **phase plan document** in `docs/` — find the step(s) to implement
- Read the **relevant sections of `docs/architecture.md`** that the plan references — confirm you understand the patterns this step extends
- Read **`docs/Values.md`** — these are the design principles that implementation must not drift from
- Do NOT re-read the full requirements doc — the plan should capture what matters. Only consult requirements if the plan explicitly flags an open question.

### Step 2: Quick Validation (REQUIRED)
**Do NOT write code yet.** This is a sanity check, not a re-discussion of design.
- Verify the plan's pattern references match what's actually in the code (patterns may have evolved since planning)
- Verify no code changes since planning invalidate the approach
- **If the plan is detailed, follows intended patterns, and has no conflicts: say so and move on.** No need to generate discussion where none is needed.
- If genuine issues are found: present them and re-invoke `/refine-feature` to update the plan before proceeding
- Get user confirmation to proceed

### Step 3: TDD Implementation
Once approach is confirmed:
- Write tests first
  - **Anchor tests to requirements, not implementation.** Before writing tests, restate the user story in one sentence. Write at least one test that validates the end-to-end intent of that story. Implementation-path tests are fine as supplements, but the anchor test should be: "does the user's described outcome happen?" A test like "returns ActionPickup" validates code structure; a test like "ground vessel ends up filled with water after action completes" validates intent.
  - **Don't assert on exact log/activity wording** (per CLAUDE.md brittle string matching guideline). If you encounter existing tests that assert on exact message text, remove the brittle assertions — don't update them to match new wording.
- Implement minimum code to pass tests
- Run tests to verify
- **Pause at each [TEST] checkpoint** for user to rebuild and manually test
- When broader design questions emerge mid-implementation, pause for discussion
  - Re-invoke `/refine-feature` if the plan needs updating
- After each [TEST] checkpoint passes, follow the [DOCS] and [RETRO] checkpoints in the plan doc

### Step 4: Human Testing ([TEST])
**Do NOT mark feature complete until user has tested.**
- If scenario for verification is complex, consider running /test-world
- Wait for explicit confirmation from user before continuing

### Step 5: Update Documentation ([DOCS])
Only after human testing confirms success:
- Run /update-docs
- Mark feature complete in phase plan

### Step 6: Retro ([RETRO])
After documentation is updated, run /retro

---

**Collaboration Norms:**
- TDD for bug fixes too — write regression tests
- Small iterations — keep changes focused
- Planning lives in `docs/`, not in ephemeral plan files
