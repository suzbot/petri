---
name: refine-feature
description: "Discuss and refine a feature, step, or task — whether part of a phase plan or stand-alone. Enforces discussion-first and produces a detailed plan with TDD and human testing checkpoints. Trigger phrases: 'refine the next step', 'discuss the next feature', or similar."
user-invocable: true
argument-hint: Feature or step to refine (e.g., "Slice 6 Step 3" or "sprout maturation")
---

## Refining a Feature

**Do NOT enter plan mode.** Work within the existing planning document in `docs/` and discuss as conversation.

### Step 1: Build Context (REQUIRED)
Read these before discussing approach:
- **`docs/architecture.md`** — Read sections relevant to this feature FIRST. Identify which established patterns apply (Component Procurement, Order execution, Pickup helpers, Recipe system, etc.). This is the routing table that prevents expensive broad code exploration — use it before reaching for Explore or reading implementation files.
- **`docs/Values.md`** — Design principles that shape implementation decisions (consistency, source of truth, reuse). Keep these in mind when evaluating approaches.
- **Original requirements document** for full context (linked at top of phase plan doc if part of larger phase). Proposals that contradict or miss requirements waste discussion time.
- **Phase plan document** in `docs/` (if exists) — this is the planning artifact, not a new file
- Any questions deferred from phase planning
- **Implementation code** (if needed) — If the feature extends existing systems and the docs above don't give enough detail on how those systems currently work, use targeted reads or an Explore agent to understand the relevant code. Architecture.md often points to the right files; start there before doing broad exploration.

### Step 2: Discussion First (REQUIRED)
**Do NOT write code yet.** First:
- Ask any clarifying questions, deferred or new
- Present implementation approach with trade-offs as **conversation** (not structured multiple-choice — reserve that for simple bounded decisions)
- Critically evaluate assumptions in existing plans
- Get user confirmation on approach

### Step 3: Update Plan (REQUIRED)
Update the existing planning document in `docs/` with design decisions and implementation steps. **The plan doc is the handoff artifact** — after `/clear`, `/implement-feature` will rely on it cold as the source of truth.

A plan is implementation-ready when it has:
- **Granular implementation steps** with iterative testable checkpoints: [TEST], [DOCS], and [RETRO]
  - Each functional accomplishment reconciled with references to original requirements doc
  - **Architecture pattern references** — name which patterns each step extends (e.g., "follows Component Procurement pattern", "uses EnsureHasVesselFor") so `/implement-feature` can validate without re-deriving
  - **New entity fields require same-step serialization** — if a step adds fields to an entity struct, the corresponding Save struct and serialize/deserialize code must be updated in that same step
- **Tests first:** planned before/at the beginning of each step (TDD)
- **Human testing checkpoints** after each testable milestone, not just at the end
  - It is OK if human testing is only for a partial workflow — just note what won't function until a subsequent step
- **/retro checkpoints** after each human testing checkpoint

## Checkpoint Definitions

Sections below describe what [TEST], [DOCS], and [RETRO] checkpoints involve. The plan doc controls *when* they happen — these definitions won't stay in context through long implementation, so the plan doc is the source of truth for sequencing.

### Human Testing Checkpoint ([TEST])
- After implementation, run the test suite then prompt the user to test
- If scenario for verification is complex, plan to create a save file via /test-world
- Force pause for explicit confirmation from user before marking complete

### Update Documentation ([DOCS])
Run only after human testing confirms success, using /update-docs

### Retro ([RETRO])
Run after documentation is updated, using /retro
