---
name: refine-feature
description: "Discuss and refine a feature, step, or task — whether part of a phase plan or stand-alone. Enforces discussion-first and plans for TDD, and human testing checkpoints. Trigger phrases: 'refine the next step', 'discuss the next feature', or similar.
---

## Refining a New Feature

**Do NOT enter plan mode.** Work within the existing planning document in `docs/` and discuss as conversation.

### Step 1: Review Context (REQUIRED)
Read these before discussing approach:
- Reread the **original requirements document** for full context (linked at top of phase plan doc if part of larger phase). Proposals that contradict or miss requirements waste discussion time.
- Phase plan document in `docs/` (if exists) — this is the planning artifact, not a new file
- Any questions deferred from phase planning
- Run /architecture if exploring unfamiliar areas of the codebase
- **Check `docs/architecture.md` for applicable patterns** — For example: read the Component Procurement Pattern (item acquisition), Pickup Activity, and Order execution sections, if the feature involves acquiring, dropping, or seeking items. Identify which `EnsureHas*` variant applies before proposing an approach.


### Step 2: Discussion First (REQUIRED)
**Do NOT write code yet.** First:
- Ask any clarifying questions, deffered or new
- Present implementation approach with trade-offs as **conversation** (not structured multiple-choice — reserve that for simple bounded decisions)
- Critically evaluate assumptions in existing plans
- Get user confirmation on approach

### Step 3: Update Plan (REQUIRED)
- Update the existing planning document in `docs/` with design decisions and implementation steps
  - Break down plan into **granular implementation steps**, with iterative testable checkpoints: [TEST], [DOCS], and [RETRO]
    - Ensure each functional accomplishment is reconciled with references to original requirements doc
    - **New entity fields require same-step serialization** — if a step adds fields to an entity struct (Character, Item, ItemVariety, PlantProperties, etc.), the corresponding Save struct and serialize/deserialize code must be updated in that same step, not deferred to a later phase. Otherwise the user won't be able to test from created save files. Fields that aren't serialized silently revert to zero values on load.
  - **Tests first:** plan before/at the beginning of each step (TDD)
  - **Human testing checkpoints** after each testable milestone, not just at the end
    - It is OK if human testing is only for a partial workflow and there are still parts that won't function as intended until a subsequent step, just make sure this is noted
  - **/retro checkpoints** after each human testing checkpoint
 
 ## Definitions

Sections below describe what [TEST], [DOCS], and [RETRO] checkpoints will involve. The plan doc controls *when* they happen — these steps won't stay in context through long implementation, so the plan doc is the source of truth for sequencing.

### Step 4: Human Testing Checkpoint ([TEST])
- After implementation, the system will run its test suite and then prompt the user to test
- If scenario for verification is complex, it may be worth planning to create a save file for the user via /test-world
- This checkpoint will force pause for explicit confirmation from user before marking the feature as complete

### Step 5: Update Documentation ([DOCS])
Will be run only after human testing confirms success using /update-docs

### Step 6: Retro ([RETRO])
Will be run after documentation is updated, using /retro.


