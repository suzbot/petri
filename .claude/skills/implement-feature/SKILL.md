---
name: implement-feature
description: "Implement an already planned feature, step, or task — whether part of a phase or stand-alone. Confirms and develops against detailed plan including TDD, and human testing checkpoints. Trigger phrases: 'implement the next step', 'build the next feature', 'start development' or similar.
---

## Confirming a New Feature

**Do NOT enter plan mode.** Work within the existing planning document in `docs/` and discuss as conversation.

### Step 1: Review Context (REQUIRED)
Read these before confirming approach:
- Reread the **original requirements document** for full context (linked at top of phase plan doc if part of larger phase). Plans that contradict or miss requirements waste discussion time.
- Read Phase plan document in `docs/` (if exists) — this is the discussion artifact
- Check for any unanswered questions from phase planning
- Run /architecture if exploring unfamiliar areas of the codebase or looking for applicable patterns

### Step 2: Confirm Approach First (REQUIRED)
**Do NOT write code yet.** First:
- Critically evaluate assumptions in existing plans, especially if there are discrepancies in approach from patterns in docs/architecture.md, or conflics with docs/Values.md
	- If there is any conflict with patterns or values, present reccomended approach with trade-offs as conversation
	- You don't need generate something to discuss if there is no conflict, it is acceptable to say the plan is detailed, follows intended patterns, and aligns with delvelopment values.
- If needed, ask any clarifying questions, deffered or new
- Ensure that the plan has small iterative steps with appropriate testing checkpoints
- Get user confirmation on approach

### Step 3: Update Plan (REQUIRED)
- Update the existing planning document in `docs/` with any changes or clarifying details

### Step 4: TDD Implementation
Once approach is confirmed:
- Write tests first
- Implement minimum code to pass tests
- Run tests to verify
- **Pause at each [TEST] checkpoint** for user to rebuild and manually test before continuing
- Pause when broader design questions emerge for discussion before proceeding
	- Re-invoke /refine-feature if user has feedback questioning the approach to ensure plan is thought through and updated appropriately
- After each [TEST] checkpoint passes, follow the [DOCS] and [RETRO] checkpoints in the plan doc

**Note:** Sections describe what [TEST], [DOCS], and [RETRO] checkpoints involve. The plan doc controls *when* they happen — these steps won't stay in context through long implementation, so the plan doc is the source of truth for sequencing.

### Step 5: Human Testing Checkpoint ([TEST])
**Do NOT mark feature complete until user has tested.**
- If scenario for verification is complex, consider running /test-world
- Wait for explicit confirmation from user before continuing

### Step 6: Update Documentation ([DOCS])
Only after human testing confirms success:
- run /update-docs
	- Updates README (including Latest Updates), CLAUDE.md, docs/game-mechanics, and docs/architecture as needed
	- Mark feature complete in phase plan

### Step 7: Retro ([RETRO])
After documentation is updated, run /retro.

---

**Collaboration Norms:**
- TDD for bug fixes too — write regression tests
- Small iterations — keep changes focused
- Planning lives in `docs/`, not in ephemeral plan files