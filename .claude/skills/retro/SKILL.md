---
name: retro
description: "Post-feature reflection to identify process changes that reduce future friction and context overhead. Run after a feature is marked complete, or mid-session when the user wants a process check-in."
user-invocable: true
argument-hint: Brief context about what was just completed (optional)
model: sonnet
agent: true
allowed-tools:
  - Read
  - Glob
  - Grep
---

**Note for caller (not the retro agent):** Before invoking this skill, ask the user if they want conversation history searched for uncaptured friction. If yes, search history using the process in "Conversation History Search" at the bottom of this file, then pass findings as part of the retro arguments. After the retro completes and proposals are resolved, update `memory/last-retro.txt` with the current ISO timestamp.

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

#### 3. Options & Decision Signals
- Were multiple options presented?
- If yes:
  - What criteria did the user appear to value when selecting an option?
    (e.g., simplicity, flexibility, explicitness, speed, extensibility)

#### 4. Reminders of Prior Context
- Did the user remind you of something already read or provided this session?
- If yes:
  - Where did context slip?
  - How could documents, skills, or tools provide better breadcrumbs so the
    information is used when needed, without adding unnecessary overhead?

#### 5. Human-Caught Bugs or Issues
- Were any bugs, logical gaps, or misalignments caught during user testing checkpoints?
- If yes:
  - What recurring pattern or pitfall caused them?
  - Could a lightweight guardrail have prevented them?

If none of the above occurred, explicitly state that **no meaningful
collaboration friction was observed**. If any of the above did occur, carry the insights to Step 2.

---

### Step 2: Process Refinement

A. Assess whether Step 1 insights revealed **clear, recurring, or token/context-heavy friction**
- If yes:
  - are improvements best addressed by one or more of the following:

    - Updates to documents
    - Updates to Claude skills
    - Updates to custom agents
    - Creation of a new Claude skill
    - Creation of a new custom agent

    For each proposed change:
    - Explain **why it is worth the cost**
    - Specify **what future context, effort, or friction it saves**
    - Keep recommendations minimal, concrete, and scoped
    - **Prefer docs, skills, or agents over CLAUDE.md or Memories** — CLAUDE.md and Memories are always-loaded context and should stay lightweight and should only contain broadly applicable notes. Place refinements where they'll be loaded contextually.

- If No: changes do not clearly pass the cost-benefit threshold
  - Did the user request the retro? If so, provide the analysis and the reasoning for no change.
  - Was the retro automatic? No output is needed, do not solicit user response.

B. Take what you learned about user values, and propose additions to docs/Values.md.

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

### 2. Find session files

```bash
# List recent session JSONL files by filesystem mtime
ls -lt ~/.claude/projects/-Users-suzanneerin-projects-petri/*.jsonl | head -N
```

Filter to files modified after the last-retro timestamp (or take last 3 if no timestamp).

### 3. Extract user messages and scan for friction

```bash
# For each session file, extract user messages and look for friction signals
python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    for line in f:
        try:
            obj = json.loads(line)
            if obj.get('type') == 'user':
                content = obj.get('message', {}).get('content', '')
                if isinstance(content, str) and len(content) > 10:
                    print(content[:300])
                    print('---')
        except: pass
" <session-file>
```

Scan user messages for friction signals:
- Corrections ("no, I meant...", "that's not what I...")
- Reminders ("don't forget", "we already discussed", "did we", "does the skill say")
- Process observations ("friction", "missed", "should have", "too strong", "too weak")

### 4. Pass findings to retro

Include a summary of friction signals found (with direct quotes) in the retro skill arguments.

### 5. After retro completes

Write current timestamp to `memory/last-retro.txt`:
```bash
date -u +"%Y-%m-%dT%H:%M:%SZ" > memory/last-retro.txt
```
