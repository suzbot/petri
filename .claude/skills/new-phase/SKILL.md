---
name: new-phase
description: "Start planning a new development phase. Reads requirements, facilitates discussion, and creates the phase plan."
---

## Starting a New Phase

Follow this process to plan the next development phase:

### Step 1: Identify the Next Phase
Read the roadmap section of CLAUDE.md to find what's next. This skill only needs to be leveraged if the next item is a phase or large feature with a requirements doc.

### Step 2: Read Requirements
- Find and read the requirements document for this phase (typically in `docs/`).
- For context on where this phase fits within the longer term plan and how it sets up future phases, reference docs/VISION.txt.

### Step 3: Discuss Approach
**Do NOT enter plan mode.** Discuss as conversation.
- Use /architecture to get a jumpstart on finding and understanding current design patterns
- **Identify which established patterns apply** to each planned step (Component Procurement, Order execution, Pickup helpers, Recipe system, etc.). Steps that involve item acquisition should name which `EnsureHas*` variant they'll use. Steps that add entity fields should include serialization. Catching pattern mismatches here prevents rework during implementation.
- Ask clarifying questions about requirements
- Present high-level approach options with trade-offs as **prose discussion**
- Identify an iterative approach with frequent human testing checkpoints
- Only ask questions that affect high-level decisions - feature-specific questions wait

### Step 4: Create Phase Plan
Save the high-level phase plan to `docs/[name]-phase-plan.md` — this is the **single planning artifact** for the phase. Do not create separate plan files. Contents:
- Original requirements reference — include an explicit link to the requirements document (e.g., `[Gardening-Reqs.txt](Gardening-Reqs.txt)`) at the top of the plan so it's easy to find during implementation
- User-facing outcomes for each sub-phase
- [TEST] checkpoints where user can verify behavior
- [DOCS] checkpoint after each [TEST] to update README, CLAUDE.md, game-mechanics, and architecture as needed
- [RETRO] checkpoint after each [DOCS] to run /retro on any human-testable, README-worthy feature
- Suggest any opportunistic or triggered enhancements from docs/randomideas.md or triggered-enhancements.md that could be worked into the plan.

Aim for shippable steps — plan to avoid regressions where reasonably feasible, so each step leaves the codebase in a working state.

**Assumption discipline**: Plan steps should reflect what the requirements say, not invent mechanics or constraints beyond them. If a design decision is needed that the requirements don't address, flag it explicitly as a deferred question for the implementation discussion (Step 5), not as a baked-in assumption.

### Step 5: Note Feature Questions
For clarifications that don't impact the high-level approach, save those questions in feature plan docs to ask in context during implementation.

---

**Reminder:** Features within a phase are approached one at a time with fresh discussion, todos, and testing for each. Use /new-feature when starting each one.
