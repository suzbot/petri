---
name: new-phase
description: "Start planning a new development phase. Reads requirements, facilitates discussion, and creates the phase design document."
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
- **Read `docs/architecture.md`, `docs/game-mechanics.md`, `docs/Values.md`, and `docs/VISION.txt` first** — game-mechanics.md is the behavioral reference for how the whole game works from the player's perspective — use it as a starting point for understanding what systems exist and how they interact before investigating code. architecture.md is the routing table that prevents expensive broad code exploration. Values.md shapes design decisions (consistency, source of truth, reuse). VISION.txt grounds design in the project's core conceits (e.g., code structure mirrors character knowledge of the world). Understand current behavior, patterns, principles, and vision before reaching for Explore or reading implementation files.
- **Identify which established patterns apply** to each planned step (Component Procurement, Order execution, Pickup helpers, Recipe system, etc.). Steps that involve item acquisition should name which `EnsureHas*` variant they'll use. Steps that add entity fields should include serialization. Catching pattern mismatches here prevents rework during implementation.
- **Surface architecture and values alignment in discussion** — for each planned step, name the architecture pattern it follows and which Values.md principles apply. This is part of the conversation, not internal bookkeeping. The user evaluates whether patterns and principles are correctly applied; silently identifying them doesn't allow that.
- Ask clarifying questions about requirements
- Present high-level approach options with trade-offs as **prose discussion**
- Identify an iterative approach with frequent human testing checkpoints
- Only ask questions that affect high-level decisions - feature-specific questions wait

### Step 4: Create Phase Design Document

Save the phase design document to `docs/[name]-design.md`. This is the **design artifact** for the phase — it captures what we're building and why. Implementation details for each step are written separately by `/refine-feature` into `docs/step-spec.md`.

**Use this format:**

```markdown
# [Phase Name] Design

Requirements: [Requirements-Doc.txt](Requirements-Doc.txt)

## Overview
[2-3 sentences: what this phase adds to the game, what it sets up for the future]

---

## Steps

### Step N: [Name]
**Status:** Planned

**Anchor story:** [1-3 sentences of what the player/character experiences.
This is the primary handoff artifact — /refine-feature derives scope from it,
/implement-feature derives tests from it. REQUIRED for every step.]

**Scope:**
[Brief bullet list of what this step introduces — new entity fields, new
activities, new UI, config changes. Declarative, not implementation-level.
"BundleCount field on Item" not "modify Pickup() to check BundleCount."]

**Open questions:**
[Questions deferred for refinement. Each resolved question gets struck through
with a cross-reference: ~~question~~ → DD-N]

**Triggered enhancements:**
[From triggered-enhancements.md or randomideas.md. Evaluate during this step.]

---

## Design Decisions

### DD-1: [Title]
**Context:** [What question arose and why]
**Decision:** [What we chose]
**Rationale:** [Why — name the architecture pattern or Values.md principle]
**Affects:** Step N, Step M
```

**Format rules:**
- Every step MUST have an anchor story. A step without an anchor story cannot be refined or implemented — it has no test derivation source.
- Open questions and triggered enhancements live with their steps, not in separate sections. This prevents them being missed during refinement.
- Design decisions are numbered globally (DD-1, DD-2, ...) and live in a single section. New decisions from refinement sessions get the next number. Each decision cross-references which steps it affects.
- Scope bullets are brief and declarative. The step spec (produced by `/refine-feature`) holds the implementation details.
- [TEST], [DOCS], [RETRO] checkpoints are NOT in the design doc — they go in the step spec.

**Assumption discipline**: Plan steps should reflect what the requirements say, not invent mechanics or constraints beyond them. If a design decision is needed that the requirements don't address, flag it as an open question on the relevant step, not as a baked-in assumption.

Aim for shippable steps — plan to avoid regressions where reasonably feasible, so each step leaves the codebase in a working state.

Suggest any opportunistic or triggered enhancements from docs/randomideas.md or triggered-enhancements.md that could be worked into the plan — place them with the relevant step.

### Step 5: Note Feature Questions
For clarifications that don't impact the high-level approach, add them as open questions on the relevant step in the design doc. They'll be addressed during `/refine-feature`.

---

**Reminder:** Features within a phase are approached one at a time with fresh discussion, todos, and testing for each. Use `/refine-feature` to design each step, then `/implement-feature` to build it.
