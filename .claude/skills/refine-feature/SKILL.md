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
- **Reconcile before proposing** (see below)
- Present implementation approach with trade-offs as **conversation** (not structured multiple-choice — reserve that for simple bounded decisions)
- Critically evaluate assumptions in existing plans
- Get user confirmation on approach

**Reconciliation check — do this before presenting your approach:**

If the step already has implementation details from a prior planning pass, cross-check each detail against its sources before discussing. Explicitly state alignment or drift:
1. **Requirements** — the specific lines this step traces to. Does the implementation match the requirement's language and intent, or has it drifted (e.g., flattened a concept, lost an abstraction)?
2. **Decisions Log** — any entries that shaped this step's design. Does the implementation honor the recorded decision, or did a later refinement erode it?
3. **Architecture patterns** — does the implementation follow established patterns? Name them.
4. **Values** — which Values.md principles apply? Does the implementation honor them?

Surface any drift in conversation so it gets discussed, not silently carried forward. Prior planning passes can erode design decisions — refinement that doesn't check against earlier decisions can make things worse, not better.

**When presenting options:**
- State what you're deciding between explicitly (don't assume it's obvious)
- For design questions, name the trade-off axis (e.g., "cardinal vs 8-directional adjacency")
- If you realize mid-explanation that you haven't actually surfaced a question, pause and reframe

**Anti-pattern:** Stating an interpretation as if it's an open question without naming the alternatives.
**Anti-pattern:** Treating an existing plan's implementation details as ground truth. The requirements and Decisions Log are ground truth; the plan is a draft that may have drifted.

### Step 2.5: Document Decisions Before Deep Exploration (REQUIRED)

Before invoking expensive exploration (Explore agents, broad code reads for implementation details), update the planning document with:
- Resolved design questions and their rationale
- Scope changes (descoped features, deferred items)
- Key design decisions from discussion

**Why:** Decisions are fresh in context. If exploration gets interrupted or expensive, the resolved thinking is already captured. The plan doc is the handoff artifact — it should reflect discussion outcomes immediately.

**When to skip:** If no decisions were made yet (still in Q&A phase), continue discussion before documenting.

### Step 3: Update Plan (REQUIRED)

#### Step 3a: Present Outline Conversationally

Before writing detailed steps into the plan doc, present the step breakdown to the user as conversation at a high-to-medium level of detail:
- "Here are the N steps I see this breaking into: ..."
- Include enough detail to evaluate sequencing, scope per step, and dependencies
- Do NOT write into the plan doc yet — this is a digestibility and alignment check

Get user feedback on the outline before proceeding. Adjust if needed.

#### Step 3b: Write Detailed Plan

Once the outline is aligned, write the detailed implementation steps into the existing planning document in `docs/`. **The plan doc is the handoff artifact** — context will likely be cleared between refinement and implementation. `/implement-feature` will rely on the plan doc cold as its sole source of truth, with no access to the discussion that produced it. Design decisions, rationale, pattern choices, and anti-patterns must all be captured in the plan — anything left in conversation context is effectively lost.

**Refinement checklist (verify for each step before writing):**
- [ ] **Human testing checkpoint:** Is there a user-verifiable behavior at this step? If yes, add [TEST] checkpoint. If no, state why (e.g., "pure logic, no UI").
- [ ] **Reqs reconciliation:** Verify this was addressed in Step 2 discussion. Show the work: cite the requirement lines and confirm the implementation matches.
- [ ] **Architecture alignment:** Verify this was addressed in Step 2 discussion. Show the work: name the pattern and confirm the implementation follows it.
- [ ] **Values alignment:** Verify this was addressed in Step 2 discussion. Show the work: cite which Values.md principles apply and how the design honors them.

**If the feature resolved open design questions during Step 2, explicitly document:**
- [ ] Deferred scope (what was descoped and where is it tracked?)

A plan is implementation-ready when it has:
- **Anchor story per step** — each step should open with a 1-2 sentence narrative of what the user/character experiences. This grounds the implementation details in the "why" and makes the step readable as a story, not just a task list. Example: "Character gets a Water Garden order but has no vessel. They procure one, fill it at the pond, and start watering."
- **Granular implementation steps** with iterative testable checkpoints: [TEST], [DOCS], and [RETRO]
  - Each functional accomplishment reconciled with references to original requirements doc
  - **Architecture pattern references** — name which patterns each step extends (e.g., "follows Component Procurement pattern", "uses EnsureHasVesselFor") so `/implement-feature` can validate without re-deriving. Include anti-patterns when ambiguity is likely (e.g., "follows ordered action pattern, NOT self-managing like ActionFillVessel"). Also name prior-step artifacts this step depends on (e.g., "calls RunWaterFill extracted in Step 5a").
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
