---
name: refine-feature
description: "Discuss and refine a feature, step, or task — whether part of a phase plan or stand-alone. Enforces discussion-first and produces a step spec with TDD and human testing checkpoints. Trigger phrases: 'refine the next step', 'discuss the next feature', or similar."
user-invocable: true
argument-hint: Feature or step to refine (e.g., "Step 3" or "clay terrain")
---

## Refining a Feature

**Do NOT enter plan mode.** Work within the existing planning documents in `docs/` and discuss as conversation.

### Step 1: Build Context (REQUIRED)
Read these before discussing approach:
- **Phase design doc** in `docs/` (e.g., `docs/construction-design.md`) — read the step's anchor story, scope, open questions, and triggered enhancements. Also read the Design Decisions section for prior decisions that affect this step.
- **`docs/step-spec.md`** — if one exists from a prior partial refinement, read it for continuity.
- **`docs/architecture.md`** — Read sections relevant to this feature FIRST. Identify which established patterns apply (Component Procurement, Order execution, Pickup helpers, Recipe system, etc.). This is the routing table that prevents expensive broad code exploration — use it before reaching for Explore or reading implementation files.
- **`docs/Values.md`** — Design principles that shape implementation decisions (consistency, source of truth, reuse). Keep these in mind when evaluating approaches.
- **Original requirements document** for full context (linked at top of design doc). Proposals that contradict or miss requirements waste discussion time.
- **`docs/game-mechanics.md`** (optional) — If the feature interacts with existing game systems and you need to understand current player-visible behavior without code diving, read relevant sections here. Covers stats, food selection, orders, gardening, etc.
- **Implementation code** (if needed) — If the feature extends existing systems and the docs above don't give enough detail on how those systems currently work, use targeted reads or an Explore agent to understand the relevant code. Architecture.md often points to the right files; start there before doing broad exploration.

### Step 2: Discussion First (REQUIRED)
**Do NOT write code yet.** First:
- **Address all open questions** listed in the step's section of the design doc
- **Evaluate all triggered enhancements** listed in the step's section
- Ask any additional clarifying questions
- **Reconcile before proposing** (see below)
- Present implementation approach with trade-offs as **conversation** (not structured multiple-choice — reserve that for simple bounded decisions)
- Critically evaluate assumptions in existing plans
- Get user confirmation on approach

**Reconciliation check — do this before presenting your approach:**

If the step already has implementation details from a prior planning pass, cross-check each detail against its sources before discussing. Explicitly state alignment or drift:
1. **Requirements** — the specific lines this step traces to. Does the implementation match the requirement's language and intent, or has it drifted (e.g., flattened a concept, lost an abstraction)?
2. **Design Decisions** — any DD entries in the design doc that affect this step. Does the implementation honor the recorded decision, or did a later refinement erode it?
3. **Architecture patterns** — does the implementation follow established patterns? Name them.
4. **Values** — which Values.md principles apply? Does the implementation honor them?

Surface any drift in conversation so it gets discussed, not silently carried forward. Prior planning passes can erode design decisions — refinement that doesn't check against earlier decisions can make things worse, not better.

**When presenting options:**
- State what you're deciding between explicitly (don't assume it's obvious)
- For design questions, name the trade-off axis (e.g., "cardinal vs 8-directional adjacency")
- Describe each option's implications — for player experience, structural alignment with vision (code structure mirrors character knowledge per VISION.txt), scalability, performance. Don't describe options in implementation jargon (function signatures, parameter patterns). If the user can't evaluate the difference without reading code, the description needs translating.
- If an option has no evaluable implications (purely code-organizational), decide unilaterally and note it in passing rather than presenting it as a choice.
- If you realize mid-explanation that you haven't actually surfaced a question, pause and reframe
- When a step involves multiple related items (similar names, overlapping systems), ensure each item is explicitly delineated in both the step spec and updated to be clearly separate concerns in the design document, to prevent conflation within and across implementation sessions.

**Anti-pattern:** Stating an interpretation as if it's an open question without naming the alternatives.
**Anti-pattern:** Treating an existing plan's implementation details as ground truth. The requirements and Design Decisions are ground truth; the plan is a draft that may have drifted.

### Step 2.5: Document Decisions Before Deep Exploration (REQUIRED)

Before invoking expensive exploration (Explore agents, broad code reads for implementation details), update the **phase design doc** with:
- New design decisions: add as DD-N entries in the Design Decisions section (next available number), with Context/Decision/Rationale/Affects
- Strike through resolved open questions in the step's section with a DD cross-reference (e.g., `~~question~~ → DD-14`)
- Update step scope in the design doc if it changed during discussion

**Why:** Decisions are fresh in context. If exploration gets interrupted or expensive, the resolved thinking is already captured. The design doc is the durable artifact — it should reflect discussion outcomes immediately.

**When to skip:** If no decisions were made yet (still in Q&A phase), continue discussion before documenting.

### Step 2.7: Trace Execution Path (REQUIRED)

Before writing the step spec, read the actual code for each function the plan will modify or extend. Walk through what happens at runtime when the new behavior executes:
- For ordered actions: trace tick-by-tick from intent creation through handler execution. What does Target get set to? What does the handler read? What happens on the second tick?
- For pickup/procurement chains: trace from `findXxxIntent` through `Pickup()` result handling in `applyPickup` through continuation/completion.
- For new entity fields: trace all code paths that read or write the parent struct.

This is targeted reads (3-5 files the plan already names), not broad exploration. The goal is to identify assumptions in the plan that don't match the code — the class of bugs where the plan describes the right behavior but misses a code-level detail (like Target needing to be a BFS step, not a destination).

Surface any mismatches as discussion items before proceeding to the step spec.

### Step 3: Update Plan (REQUIRED)

#### Step 3a: Present Outline Conversationally

**Before presenting the outline:** If the design shifted materially during Step 2 discussion (different approach, broader scope, changed mechanics), re-run the reconciliation check (reqs, decisions log, architecture, values) against the new design. The Step 2 reconciliation validated the *starting point* — Step 3a must validate what the discussion actually produced.

Present the step breakdown to the user as conversation at a high-to-medium level of detail:
- "Here are the N sub-steps I see this breaking into: ..."
- Include enough detail to evaluate sequencing, scope per sub-step, and dependencies
- For each sub-step, name the architecture pattern and Values.md principle it follows — this surfaces the Step 2 reconciliation work for the user to validate before the written plan
- Do NOT write into the step spec yet — this is a digestibility and alignment check

Get user feedback on the outline before proceeding. Adjust if needed.

#### Step 3b: Write Step Spec

Once the outline is aligned, write the detailed implementation plan to **`docs/step-spec.md`** (replacing prior contents). The step spec is the handoff artifact — context will likely be cleared between refinement and implementation. `/implement-feature` will rely on the step spec cold, with only the design doc for broader context. Design decisions, rationale, pattern choices, and anti-patterns must all be captured — anything left in conversation context is effectively lost.

**At the top of the step spec, include:**
```markdown
# Step Spec: Step N — [Name]

Design doc: [phase-design.md](phase-design.md)
```

**Refinement checklist (verify for each sub-step before writing):**
- [ ] **Human testing checkpoint:** Is there a user-verifiable behavior at this sub-step? If yes, add [TEST] checkpoint. If no, state why (e.g., "pure logic, no UI").
- [ ] **Reqs reconciliation:** Verify this was addressed in Step 2 discussion. Show the work: cite the requirement lines and confirm the implementation matches.
- [ ] **Architecture alignment:** Verify this was addressed in Step 2 discussion. Show the work: name the pattern and confirm the implementation follows it.
- [ ] **Values alignment:** Verify this was addressed in Step 2 discussion. Show the work: cite which Values.md principles apply and how the design honors them.

**If the feature resolved open design questions during Step 2, explicitly document:**
- [ ] Deferred scope (what was descoped and where is it tracked?)

**Behavioral completeness — resolve before writing:**
- Ensure all behavioral details of the feature are specified, even if they weren't in the original requirements. For ordered actions, check the "Behavioral details the plan must specify" checklist in architecture.md (targeting, duration, completion criteria, feasibility criteria, variety lock). Make recommendations and get approval from the user on all such details before recording them in the step spec.

A step spec is implementation-ready when it has:
- **Anchor story per sub-step (REQUIRED)** — each sub-step MUST open with a 1-2 sentence narrative of what the user/character experiences. This is a critical handoff artifact: `/implement-feature` derives its anchor tests directly from these stories. A sub-step without an anchor story will produce tests that validate code structure instead of user intent. Example: "Character gets a Water Garden order but has no vessel. They procure one, fill it at the pond, and start watering." The anchor test derived from this would verify: "ground vessel ends up filled with water and tiles get watered" — not "returns ActionWaterGarden."
- **Granular implementation sub-steps** with iterative testable checkpoints: [TEST] → [DOCS] → [RETRO] as a unit at every testable milestone
  - Each functional accomplishment reconciled with references to original requirements doc
  - **Architecture pattern references** — name which patterns each sub-step extends (e.g., "follows Component Procurement pattern", "uses EnsureHasVesselFor") so `/implement-feature` can validate without re-deriving. Include anti-patterns when ambiguity is likely (e.g., "follows ordered action pattern, NOT self-managing like ActionFillVessel"). Also name prior-step artifacts this step depends on (e.g., "calls RunWaterFill extracted in Step 5a").
  - **New entity fields require same-step serialization** — if a sub-step adds fields to an entity struct, the corresponding Save struct and serialize/deserialize code must be updated in that same sub-step
- **Tests first:** planned before/at the beginning of each sub-step (TDD)
- **Human testing checkpoints** after each testable milestone, not just at the end
  - It is OK if human testing is only for a partial workflow — just note what won't function until a subsequent sub-step
- **[TEST] → [DOCS] → [RETRO] as a unit** — every [TEST] checkpoint must be followed by [DOCS] and [RETRO]. These three always appear together. This ensures documentation and process reflection happen at each testable milestone, not just at the end of a feature.

## Checkpoint Definitions

Sections below describe what [TEST], [DOCS], and [RETRO] checkpoints involve. The step spec controls *when* they happen — these definitions won't stay in context through long implementation, so the step spec is the source of truth for sequencing.

### Human Testing Checkpoint ([TEST])
- After implementation, run the test suite then prompt the user to test
- If scenario for verification is complex, plan to create a save file via /test-world
- Force pause for explicit confirmation from user before marking complete

### Update Documentation ([DOCS])
Run only after human testing confirms success, using /update-docs

### Retro ([RETRO])
Run after documentation is updated, using /retro
