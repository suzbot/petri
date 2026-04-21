# Process Improvement Pass — Before Step 12

Implements proposals P1, P6, P7, P4, P8, P3 from the meta-claude proposal set (2026-04-18, updated 2026-04-20). This work lands before Step 12 (Display Polish) begins so that Steps 12 and 13 exercise the changes through real implementation cycles.

---

## Execution order

### 1. Memory-framing discipline (P1)

Add a subsection to CLAUDE.md's Collaboration section covering memory-write discipline:

- Describe the signal, don't prohibit universally. "User preferred X for Y" over "NEVER do Z."
- Name the scope. When the preference applies, when it doesn't.
- Prefer descriptive to prescriptive. "User tends to X in these situations" over "must always X."
- Don't generalize from one correction. One signal is evidence for a contextual preference, not a universal law.
- When a memory candidate feels absolute, capture it with more context — not the absolute rule, and not nothing.

**Effort:** One paragraph in CLAUDE.md.

---

### 2. Documentation additions (P6 + P7)

**P6 — Four-touchpoint checklist for intent target fields.** Add an "Adding a new Intent target field" section to `architecture.md`'s "Adding New X" checklists:
1. `CalculateIntent` continuation gate
2. Target-recalculation chain in `continueIntent`
3. Arrival-detection block
4. Final-return-intent preservation

Cross-reference from `/implement-feature` when the task involves intent-layer changes.

**P7 — Split-trigger conditions.** Add two entries to `triggered-enhancements.md`:
- `order_execution.go` (currently ~1,950 LOC): line-count threshold, new order-type that pushes past, or concept cluster ready for extraction.
- `apply_actions.go` (currently ~1,894 LOC): same structure.

PM sets thresholds during execution.

**Effort:** One new section in architecture.md + two table rows in triggered-enhancements.md.

---

### 3. CLAUDE.md restructure + consolidation (P4)

**Three movements:**

**(a) Audit SKILL.md files** using two criteria:
- Cross-cutting rules restated across multiple skills (candidates: context discipline, evidence-first, sibling-flow scan, Values.md reference, DD-N convention, [TEST]/[DOCS]/[RETRO] unit rule, memory-framing from P1)
- Discoverable content — anything Claude can find by reading the codebase. Heuristic: "can Claude find this by reading the code? If yes, delete it."

**(b) Restructure CLAUDE.md into two zones:**
- **Zone 1 — always-loaded system knowledge:** project overview, architecture pointers, collaboration norms, consolidated cross-skill rules. Stable.
- **Zone 2 — current working context:** links to active artifacts (step-spec.md, current phase design doc, roadmap status). Points to sources of truth, doesn't duplicate them.

**(c) Each SKILL.md** replaces restated cross-cutting rules with a reference to CLAUDE.md.

**(d) Effort frontmatter pass** across all skills — declare effort level where appropriate.

**(e) Cleanup pass** — remove stale content, tighten structure.

**Effort:** Medium. Audit 8 SKILL.md files, extract to CLAUDE.md, restructure, cleanup.

---

### 4. 2nd-bug-rule enforcement (P8)

Two pieces:

**(a) CLAUDE.md feedforward rule.** Add to Collaboration section:

> During testing, before writing any code change in response to a reported problem, explicitly name whether this is the first or a subsequent fix in the current testing session on the current step. If subsequent, halt for design review before proceeding.

**(b) Visible recap at fix-close.** When a fix is confirmed or a new problem is reported (implicitly closing the prior fix), state visibly: "Fix #N on [step]." If N >= 2, halt for design review.

**Effort:** Low. One paragraph in CLAUDE.md + one behavioral addition to `/fix-bug`'s close step.

**Escalation trigger** (already in triggered-enhancements.md): if explicit naming gets skipped 3+ times, evaluate computational controls.

---

### 5. Memory capture + Values.md example-extraction (P3)

Two movements:

**(a) Expand capture scope.** Skills add explicit capture-trigger language at signal points:
- Correction received
- Multi-step approach changed mid-work
- User preference explicitly stated
- Discovery, validation, or pattern noticed in the moment

Threshold: if the signal is specific, scoped, and likely to matter in a future session, capture it.

**(b) Values.md example-extraction + retro-time provenance.**
- Core rule: Values entries are rules. Supporting examples and evidence live in referenced memory files, not inline.
- Provenance links form at retro-synthesis time (when `/retro` promotes a raw memory to support a value or forms a new value from accumulated signal).
- `/retro` synthesis step runs four-phase consolidation: orient, gather signal (targeted searches), consolidate (merge overlaps, convert dates, remove contradictions), prune and update MEMORY.md index. Values.md provenance links form during phase 3.

**Effort:** Low-medium. Cross-skill capture guidance + example-extraction rule + `/retro` synthesis enhancement.

---

## Validation

Steps 12 and 13 exercise all five proposals:
- **P1 framing** — any memory written during Steps 12-13 follows the new discipline
- **P4 restructure** — CLAUDE.md loaded for two full cycles; identify if anything is missing or noisy
- **P8 rule** — Step 12 and 13 [TEST] checkpoints exercise the visible-recap pattern
- **P3 capture** — Step 12 and 13 retros exercise expanded capture + example-extraction
- **P6/P7** — passive; validated when next intent-layer or large-file work fires their triggers

---

## What comes after (post-Step 13, before next phase)

Not scoped here. Proposals P9 (go-code-reviewer reshape), P2 (subagent delegation), P5 (schema-validation hook) land after Step 13 completes and before `/new-phase` is invoked for the next game phase.
