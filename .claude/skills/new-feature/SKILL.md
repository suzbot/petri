---
name: new-feature
description: "Start implementing a feature, step, or task — whether part of a phase plan or stand-alone. Enforces discussion-first, TDD, and human testing checkpoints. Trigger phrases: 'start the next step', 'implement', 'work on', 'build'."
---

## Starting a New Feature

**Do NOT enter plan mode.** Work within the existing planning document in `docs/` and discuss as conversation.

### Step 1: Review Context (REQUIRED)
Read these before discussing approach:
- Reread the **original requirements document** for full context (linked at top of phase plan doc). Proposals that contradict or miss requirements waste discussion time.
- Phase plan document in `docs/` (if exists) — this is the planning artifact, not a new file
- Any questions deferred from phase planning
- Run /architecture if exploring unfamiliar areas of the codebase

### Step 2: Discussion First (REQUIRED)
**Do NOT write code yet.** First:
- Ask any deferred clarifying questions
- Present implementation approach with trade-offs as **conversation** (not structured multiple-choice — reserve that for simple bounded decisions)
- Critically evaluate assumptions in existing plans, not just implement what's written
- Get user confirmation on approach
- Update the existing planning document in `docs/` with design decisions and implementation steps
  - Tests listed first (TDD)
  - Human testing checkpoints after each meaningful milestone, not just at the end
  - Reconcile back to original requirements at each checkpoint

### Step 3: TDD Implementation
Once approach is confirmed:
- Write tests first
- Implement minimum code to pass tests
- Run tests to verify
- **Pause at each [TEST] checkpoint** for user to rebuild and manually test before continuing
- Pause when broader design questions emerge for discussion before proceeding
- After each [TEST] checkpoint passes, follow the [DOCS] and [RETRO] checkpoints in the plan doc

**Note:** Steps 4-6 below describe what [TEST], [DOCS], and [RETRO] checkpoints involve. The plan doc controls *when* they happen — these steps won't stay in context through long implementation, so the plan doc is the source of truth for sequencing.

### Step 4: Human Testing Checkpoint ([TEST])
**Do NOT mark feature complete until user has tested.**
- If scenario for verification is complex, consider running /test-world
- Wait for explicit confirmation from user before continuing

### Step 5: Update Documentation ([DOCS])
Only after human testing confirms success:
- Update README (including Latest Updates), CLAUDE.md, docs/game-mechanics, and docs/architecture as needed
- Mark feature complete in phase plan

### Step 6: Retro ([RETRO])
After documentation is updated, run /retro.

---

**Collaboration Norms:**
- TDD for bug fixes too — write regression tests
- Small iterations — keep changes focused
- Planning lives in `docs/`, not in ephemeral plan files
