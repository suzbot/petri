---
name: refine-feature
description: "Discuss and refine a feature, step, or task — whether part of a phase plan or stand-alone. Enforces discussion-first and produces a step spec with TDD and human testing checkpoints. Trigger phrases: 'refine the next step', 'discuss the next feature', or similar."
user-invocable: true
argument-hint: Feature or step to refine (e.g., "Step 3" or "clay terrain")
---

## Purpose

Refinement takes a planned step and produces an implementation-ready spec through discussion. The goal is a shared, documented understanding of what we're building and why — so that implementation can proceed in a separate context with confidence.

The design doc's scope items are decisions about *what* to build. Refinement discusses *how* — resolving open questions, catching drift between the plan and its sources, documenting decisions, and verifying assumptions against code. The output is a step spec detailed enough to hand off cold.

**Do NOT enter plan mode.** Work within the existing planning documents in `docs/` and discuss as conversation.

---

### Step 1: Build Context (REQUIRED)

Read these before discussing approach:
- **Original requirements document** for full context (linked at top of design doc). Proposals that contradict or miss requirements waste discussion time.
- **Phase design doc** in `docs/` (e.g., `docs/construction-design.md`) — read the step's anchor story, scope, open questions, and triggered enhancements. Also read the Design Decisions section for prior decisions that affect this step.
- **`docs/step-spec.md`** — if one exists from a prior partial refinement, read it for continuity.
- **`docs/architecture.md`** — Read sections relevant to this feature FIRST. Identify which established patterns apply (Component Procurement, Order execution, Pickup helpers, Recipe system, etc.). This is the routing table that prevents expensive broad code exploration — use it before reaching for Explore or reading implementation files.
- **`docs/Values.md`** — Design principles that shape implementation decisions (consistency, source of truth, reuse). Keep these in mind when evaluating approaches.
- **`docs/game-mechanics.md`** (optional) — If the feature interacts with existing game systems and you need to understand current player-visible behavior without code diving, read relevant sections here. Covers stats, food selection, orders, gardening, etc.
- **Implementation code** — Use targeted reads or an Explore agent for initial exploration to understand the relevant code. Architecture.md often points to the right files; start there before doing broad exploration. Deeper exploration comes later in Step 2.7.

### Step 2: Discussion First (REQUIRED)

**Do NOT write code yet.**

#### Step 2a: Align on Intent (REQUIRED — do this BEFORE any analysis)

State your understanding of what this step accomplishes in 2-3 sentences — what changes for the player or the codebase, and why it matters now. **Stop and confirm alignment.** Do NOT present drift checks, scope evaluations, approach analysis, or implementation details until the user confirms or corrects your understanding. The user may have vision or context that fundamentally shapes the approach — learn it before analyzing.

**Anti-pattern:** Presenting a wall of analysis (drift check + scope evaluation + approach + questions) before understanding the user's intent. This wastes discussion time when the user's vision differs from what the plan implies.

#### Step 2b: Discuss Approach (after intent alignment)

With intent aligned:
- **Address all open questions** listed in the step's section of the design doc
- **Evaluate all triggered enhancements** listed in the step's section
- **Propose new triggered enhancements** when a change builds infrastructure ahead of its real test conditions (e.g., scoring logic with only one candidate today). Name the trigger condition that would make real test coverage meaningful.
- **Check for drift** (see below)
- Present implementation approach with trade-offs as **conversation** (not structured multiple-choice — reserve that for simple bounded decisions)

**Present scope evaluations as proposals**, not closed decisions. Show the reasoning and recommendation, but let the user confirm or correct before striking through open questions.

**Discuss implementation specifics before the outline.** Before presenting the sub-step outline (Step 3a), explicitly propose and get alignment on: where new code will live (files, functions), what config values are needed and why, what the core algorithm/formula is, and any signature changes to existing functions. These are design choices the user should evaluate — don't embed them in the outline for the first time. The outline should confirm decisions already discussed, not introduce them.

**Drift check:**

The design doc scope says what to build. Prior planning passes may have added implementation details that have drifted from their sources. Cross-check implementation details against:
1. **Requirements** — does the implementation match the requirement's language and intent?
2. **Design Decisions** — does it honor recorded DD entries, or did a later pass erode one?
3. **Architecture patterns** — does it follow established patterns? Name them.
4. **Values** — which Values.md principles apply?
5. **System interactions** — does it introduce states that existing systems (feasibility, completion, abandonment) don't handle?

Surface drift so it gets discussed. The scope items themselves are not subject to re-evaluation — they represent decisions already made. Drift checking verifies the *how* honors the *what*.

**Anchor story quality check — do this after drift check:**

The anchor story drives the anchor test. If the story is vague, the test will validate structure instead of intent. A good anchor story:
- Names a concrete scenario with specific quantities or conditions (not "character builds a hut" but "character delivers 2 stick bundles to a marked tile, then builds from an adjacent tile")
- Encodes design decisions inline — if a DD constrains a behavior (threshold, gate, material choice), the story should include a scenario that exercises that constraint
- Covers at least one intermediate decision point, not just the end state
- If the step references DD entries, re-read each DD and verify the anchor story doesn't contradict it

**When presenting options:**
- State what you're deciding between explicitly (don't assume it's obvious)
- For design questions, name the trade-off axis (e.g., "cardinal vs 8-directional adjacency")
- Describe each option's implications — for player experience, structural alignment with vision (code structure mirrors character knowledge per VISION.txt), scalability, performance. Don't describe options in implementation jargon (function signatures, parameter patterns). If the user can't evaluate the difference without reading code, the description needs translating.
- If an option has no evaluable implications (purely code-organizational), decide unilaterally and note it in passing rather than presenting it as a choice.
- If you realize mid-explanation that you haven't actually surfaced a question, pause and reframe
- If you catch yourself revisiting the same scope/ordering concern twice within your own reasoning (before surfacing to the user), that's churning — stop immediately and surface the specific blocker. State what's causing the concern and what tradeoff you see. The user may have a simple solution. Check existing patterns (architecture.md, analogous features) before raising scope concerns.
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

Before writing the step spec, grep existing symbol constants and helper functions for any names or values the plan will introduce — conflicts in a small namespace are findable before proposing specifics. Then read the actual code for each function the plan will modify or extend. Walk through what happens at runtime when the new behavior executes:
- For ordered actions: trace tick-by-tick from intent creation through handler execution. What does Target get set to? What does the handler read? What happens on the second tick?
- For pickup/procurement chains: trace from `findXxxIntent` through `Pickup()` result handling in `applyPickup` through continuation/completion.
- For new entity fields: trace all code paths that read or write the parent struct, including ToSaveState/FromSaveState for any fields that touch serialized state.

- For borrowed algorithms: when the plan reuses an existing algorithm for a new context (e.g., SpawnPonds logic for SpawnClay), trace the original's constraints and verify they hold in the new context. Different entity types often have different spatial or adjacency constraints that make the algorithm fail silently or produce degenerate results.
- When the plan adds a discriminating field to a shared predicate, trace both directions: what the new field matches *and* what it causes the predicate to reject in existing call sites. If the predicate is called from multiple contexts, verify rejection behavior is correct in each.

- When extending an existing entity type, grep for existing tests of that type — follow their patterns and conventions in the step spec's test plan.

This is targeted reads (3-5 files the plan already names), not broad exploration. The goal is to identify assumptions in the plan that don't match the code — the class of bugs where the plan describes the right behavior but misses a code-level detail (like Target needing to be a BFS step, not a destination).

Surface any mismatches as discussion items before proceeding to the step spec.

### Step 3: Write Step Spec (REQUIRED)

#### Step 3a: Present Outline Conversationally

Present the step breakdown to the user as conversation at a high-to-medium level of detail:
- "Here are the N sub-steps I see this breaking into: ..."
- Include enough detail to evaluate sequencing, scope per sub-step, and dependencies
- For each sub-step, name the architecture pattern and Values.md principle it follows
- **For each change, connect it to its functional outcome** — describe what it does and why in functional terms (per CLAUDE.md communication norms), not code mechanics. A change listed without its "why" is hard to evaluate. Don't list implementation artifacts (new files, helpers) without explaining what player-visible or system behavior they enable.
- For each sub-step, confirm where [TEST] falls and whether new item types need save/load in the same sub-step
- **Evaluate intermediate [TEST] opportunities within each sub-step.** When a sub-step has a natural split between player-visible changes and no-op infrastructure, propose an intermediate human testing checkpoint after the visible changes. Catching issues in a smaller blast radius is cheaper than debugging after everything is in.
- Do NOT write into the step spec yet — this is a digestibility and alignment check

**Before presenting:** If the design shifted materially during Step 2 discussion (different approach, broader scope, changed mechanics), re-run the drift check against the new design. Step 2's check validated the starting point — the outline must validate what discussion actually produced.

Get user feedback on the outline before proceeding. Adjust if needed.

#### Step 3b: Write Step Spec

Once the outline is aligned, write the detailed implementation plan to **`docs/step-spec.md`** (replacing prior contents). The step spec is the handoff artifact — context will likely be cleared between refinement and implementation. `/implement-feature` will rely on the step spec cold, with only the design doc for broader context. Design decisions, rationale, pattern choices, and anti-patterns must all be captured — anything left in conversation context is effectively lost.

**At the top of the step spec, include:**
```markdown
# Step Spec: Step N — [Name]

Design doc: [phase-design.md](phase-design.md)
```

**Implementation-ready checklist — verify before writing:**

Discussion should have resolved these. If any are unresolved, surface them now rather than writing an incomplete spec.
- All open questions from the design doc addressed
- All triggered enhancements evaluated
- Deferred scope documented (what was descoped and where is it tracked?)
- Drift check passed — implementation honors requirements, DDs, architecture patterns, and values
- Label and display text decisions confirmed with user (what the player reads are design choices)
- All applicable "Adding New X" checklists from architecture.md reviewed — each item explicitly addressed in the spec

**Behavioral completeness:**
- Ensure all behavioral details of the feature are specified, even if they weren't in the original requirements. Make recommendations and get approval from the user on details before recording them. Label and display text decisions require user input — present as open questions, don't prescribe.

**A step spec is implementation-ready when it has:**
- **Anchor story per sub-step (REQUIRED)** — each sub-step MUST open with a 1-2 sentence narrative of what the user/character experiences. This is a critical handoff artifact: `/implement-feature` derives its anchor tests directly from these stories. A sub-step without an anchor story will produce tests that validate code structure instead of user intent.
- **Granular implementation sub-steps** with iterative testable checkpoints: [TEST] → [DOCS] → [RETRO] as a unit at every testable milestone
  - Each functional accomplishment reconciled with references to original requirements doc
  - **Architecture pattern references** — name which patterns each sub-step extends (e.g., "follows Component Procurement pattern", "uses EnsureHasVesselFor") so `/implement-feature` can validate without re-deriving. Include anti-patterns when ambiguity is likely (e.g., "follows ordered action pattern, NOT self-managing like ActionFillVessel"). Also name prior-step artifacts this step depends on (e.g., "calls RunWaterFill extracted in Step 5a").
  - **New entity fields require same-step serialization** — if a sub-step adds fields to an entity struct, the corresponding Save struct and serialize/deserialize code must be updated in that same sub-step
- **Pseudocode conditions are implementation instructions.** Name the identity field as the check, not a related condition. Be sure it is clear what a concept IS vs what it DOES. `If WallRole == "door"`, not `If construct is a door (Passable == true)`.
- **Tests first:** planned before/at the beginning of each sub-step (TDD)
- **Human testing checkpoints** after each testable milestone, not just at the end
  - It is OK if human testing is only for a partial workflow — just note what won't function until a subsequent sub-step
- **[TEST] → [DOCS] → [RETRO] as a unit** — every [TEST] checkpoint must be followed by [DOCS] and [RETRO]. These three always appear together. This ensures documentation and process reflection happen at each testable milestone, not just at the end of a feature.

## Checkpoint Definitions

Sections below describe what [TEST], [DOCS], and [RETRO] checkpoints involve. The step spec controls *when* they happen — these definitions won't stay in context through long implementation, so the step spec is the source of truth for sequencing.

### Human Testing Checkpoint ([TEST])
- After implementation, run the test suite then prompt the user to test
- If scenario for verification is complex, plan to create a save file via /test-world. For placement/creation flows, the test world should give characters know-how and let the user exercise the flow — not pre-populate the data being tested.
- Force pause for explicit confirmation from user before marking complete

### Update Documentation ([DOCS])
Run only after human testing confirms success, using /update-docs

### Retro ([RETRO])
Run after documentation is updated, using /retro
