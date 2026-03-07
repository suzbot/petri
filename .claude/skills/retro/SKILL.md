---
name: retro
description: "Post-feature reflection to identify process changes that reduce future friction and context overhead. Run after a feature is marked complete, or mid-session when the user wants a process check-in."
user-invocable: true
argument-hint: Brief context about what was just completed (optional)
model: sonnet
context: fork
agent: general-purpose
allowed-tools:
  - Read
  - Glob
  - Grep
---

**Note for caller (not the retro agent):** Before invoking this skill, ask the user if they want conversation history searched for uncaptured friction. If the user says yes, **always run the search** — even if the last retro was recent. The user knows what sessions exist; a recent timestamp does not mean there's nothing to find. Search history using the process in "Conversation History Search" at the bottom of this file, then pass findings as part of the retro arguments. After the retro completes and proposals are resolved, update `memory/last-retro.txt` with the current ISO timestamp.

## Retro Skill

Perform a reflection on collaboration friction. Can be triggered:
- **Post-feature**: after a feature is explicitly marked complete
- **Mid-session**: when the user requests a process check-in (e.g., after a rocky planning phase)

This reflection must not interrupt forward progress unless the user requested the retro,
or unless the threshold for proposed changes has been met.

The goal is to identify **high-confidence, high-leverage refinements** to
collaboration that are likely to save more context in the future than they
cost to reflect on, suggest, and implement.

Avoid speculative or "nice-to-have" improvements.

---

### Step 1: Collaboration Assessment

Reflect **only on observable signals from this session**.
Do not infer intent beyond what the user explicitly stated or did.

Address the following **only if applicable**:

#### 1. Interrupted Actions
- Were any actions interrupted by the user?
- If yes:
  - What action was interrupted?
  - What signal suggested *why* the interruption occurred?
    (e.g., wrong direction, missing constraint, pacing issue)

#### 2. Clarifications Requested
- Did the user request clarifications?
- If yes:
  - What type of clarification was requested?
    (e.g., scope, assumptions, terminology, intent, constraints)

#### 3. Reminders of Prior Context
- Did the user remind you of something already read or provided this session?
- If yes:
  - Where did context slip?
  - How could documents, skills, or tools provide better breadcrumbs so the
    information is used when needed, without adding unnecessary overhead?

#### 4. Human-Caught Bugs or Issues
- Were any bugs, logical gaps, or misalignments caught during user testing checkpoints?
- If yes:
  - What recurring pattern or pitfall caused them?
  - Could a lightweight guardrail have prevented them?

#### 5. Intensive processes
- Were there any suspiciusly high-token or high-touch procceses for actions that were expected to be more straightforward? examples
  - test world creation needing to search code and other save files to reason about how to build it
  - internal churning going back and forth over a question without surfacing it
  - many requests for user approvals/interventions on what is supposed to be largely autonomous (test world creation)
- If yes:
  - Was it merited due to actual complexity, or could have a different prompt or different availability of information circumvented the extra effort?

If none of items 1-5 occurred, explicitly state that **no meaningful
collaboration friction was observed**.

#### 6. Value Signals (always assess)
Unlike items 1-6, this applies even when no friction occurred. Observe what choices the user made during the session and name the implicit value behind each:
- When the user selected between options: what did they optimize for? What criteria did the user appear to value when selecting an option?
    (e.g., simplicity, flexibility, explicitness, speed, extensibility)
- When the user redirected an approach: what principle were they enforcing?
- When the user added or removed scope: what trade-off were they making?

For each signal, state: the choice made and the value it expresses. Note if Values.md has something similar, but don't force-fit — if it feels like a distinct principle, present it as one.

Carry friction insights (items 1-6) and value signals (item 7) to Step 2.

---

### Step 1.5: Read Existing Context (REQUIRED before proposing changes)

Before formulating proposals, read what already exists so proposals build on or strengthen existing content rather than duplicating it:

- **`docs/Values.md`** — current design values. Check whether a friction signal maps to an existing value that could be strengthened (new example, broader wording) rather than a new value.
- **Relevant skills in `.claude/skills/`** — skim skills that relate to the friction observed (e.g., if the issue was during implementation, read `implement-feature/SKILL.md`). Check whether existing guidance already covers the issue but wasn't surfaced at the right time, vs. guidance that's genuinely missing.
- **claude.md** - read section on collaboration norms

The goal is: **strengthen or surface existing content first, create new content only when nothing existing covers the gap.** This prevents values and skills from growing redundantly across retros.

---

### Step 2: Process Refinement

A. Assess whether Step 1 insights revealed **clear, recurring, or token/context-heavy friction**
- If yes:
  - are improvements best addressed by one or more of the following:

    - Updates to documents
    - Updates to Claude skills
    - Updates to custom agents
    - Updates to claude.md
    - Creation of a new Claude skill
    - Creation of a new custom agent

    For each proposed change:
    - Explain **why it is worth the cost**
    - Specify **what future context, effort, or friction it saves**
    - Keep recommendations minimal, concrete, and scoped
    - **Prefer structural enforcement over prose instructions.** If a correction was needed because something was forgotten mid-session, the fix should be a task, a skill invocation, or a dependency — not a paragraph to remember. Adding more text to a skill that was already ignored doesn't solve the problem.
    - **Route proposals to the right home:**
      - **Design values** (how to think about the game/code/player) → `docs/Values.md`
      - **Communication norms** (how to present, qualify, escalate) → CLAUDE.md Collaboration section. ≤15 words per bullet. Combine with thematically related existing bullets rather than adding new ones.
      - **Workflow-specific guardrails** (when to stop, what to check) → the relevant skill in `.claude/skills/`
      - **Technical notes** — broad patterns → `docs/architecture.md`; implementation-specific pitfalls → relevant skill or `memory/MEMORY.md`
      - Avoid mixing categories: communication norms don't belong in Values.md; design values don't belong in CLAUDE.md.
    - **Build on what exists** — if Values.md or a skill already covers the topic, propose strengthening (new example, broader wording, better placement) rather than creating parallel content. Cite what you found in Step 1.5. But if the actionable norm already exists where it's in context (CLAUDE.md, a skill), don't propose adding redundant examples to files that aren't in context during the relevant workflow.
    - **Match the target file's format and density.** Skills are terse reference outlines, not prose. Proposals should edit existing bullets or add short ones — not insert paragraphs into a file that uses bullet points. Show the exact edit (old text → new text). Aim for ≤15 new words per change when strengthening existing text.

- If No: changes do not clearly pass the cost-benefit threshold
  - Did the user request the retro? If so, provide the analysis and the reasoning for no change.
  - Was the retro automatic? No output is needed, do not solicit user response.

---

## Output Rules

- Be concise and concrete
- Avoid speculative suggestions
- Present all changes as **proposals**, not decisions
- End with a clear list of proposals for the caller to relay to the user for approval
- All suggestions require **explicit user approval** before implementation
- If conversation history findings were provided, integrate them into the assessment

---

## Conversation History Search (for caller, not retro agent)

When the user opts in to history search before a retro, follow this process:

### 1. Determine search window

Read `memory/last-retro.txt` for the last retro timestamp. Search sessions modified after that timestamp. If no timestamp file exists, fall back to the last 3 sessions.

### 2. Find and scan sessions (single command)

Run one command that finds recent sessions and extracts friction signals. This avoids multiple approval prompts. Note: `last-retro.txt` stores UTC timestamps but file mtimes are local — the script converts UTC cutoff to local time for comparison.

```bash
python3 -c "
import json, glob, os, sys
from datetime import datetime, timezone
project_dir = os.path.expanduser('~/.claude/projects/-Users-suzanneerin-projects-petri')
cutoff_str = sys.argv[1].strip() if len(sys.argv) > 1 and sys.argv[1].strip() else None
files = sorted(glob.glob(os.path.join(project_dir, '*.jsonl')), key=os.path.getmtime, reverse=True)
if cutoff_str:
    cutoff_utc = datetime.fromisoformat(cutoff_str.replace('Z', '+00:00'))
    cutoff_epoch = cutoff_utc.timestamp()
    files = [f for f in files if os.path.getmtime(f) > cutoff_epoch]
else:
    files = files[:3]
friction_words = ['no, ', 'not what', 'already', 'don\\'t forget', 'we discussed', 'did we', 'does the skill', 'friction', 'missed', 'should have', 'too strong', 'too weak', 'that\\'s not', 'I meant', 'confused', 'why are', 'I don', 'wait,', 'wrong', 'actually', 'problem', 'issue', 'stuck', 'churn']
for fpath in files:
    signals = []
    with open(fpath) as f:
        for line in f:
            try:
                obj = json.loads(line)
                if obj.get('type') == 'user':
                    content = obj.get('message', {}).get('content', '')
                    if isinstance(content, str) and any(w in content.lower() for w in friction_words):
                        signals.append(content[:300])
            except: pass
    if signals:
        mtime = datetime.fromtimestamp(os.path.getmtime(fpath)).strftime('%Y-%m-%d %H:%M')
        print(f'\n=== {os.path.basename(fpath)[:12]}... ({mtime}) — {len(signals)} signals ===')
        for s in signals:
            print(s); print('---')
" "$(cat memory/last-retro.txt 2>/dev/null || echo '')"
```

### 3. Pass findings to retro

Include a summary of friction signals found (with direct quotes) in the retro skill arguments.

### 5. After retro completes

Write current timestamp to `memory/last-retro.txt`:
```bash
date -u +"%Y-%m-%dT%H:%M:%SZ" > memory/last-retro.txt
```
