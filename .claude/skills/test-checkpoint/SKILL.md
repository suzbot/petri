---
name: test-checkpoint
description: "Run the [TEST] checkpoint sequence: present test items, wait for confirmation, route issues to /fix-bug. Invoked by /implement-feature at each [TEST] milestone."
user-invocable: false
---

## Test Checkpoint

Runs when `/implement-feature` reaches a [TEST] task. The implement-feature skill creates the task — this skill executes it.

### Step 1: Prepare Test Items

1. Re-read the [TEST] items from `docs/step-spec.md` for this sub-step
2. Cross-check each item against what was actually implemented — surface contradictions before presenting to user
3. If the checkpoint calls for a test world, offer `/test-world`. **Before invoking: remind user to close the game first (auto-save on quit overwrites the test world). Wait for acknowledgment.**

### Step 2: Present and Wait

Present the test items to the user as a numbered checklist. Then:

**Wait for the user to explicitly confirm testing is complete. Do not move to the next task until confirmed.**

Do not prompt "shall I continue?" or interpret partial responses as confirmation. The user will say when they're done.

### Step 3: Route Issues

- **User reports ANY issue → invoke `/fix-bug` via the Skill tool immediately.**
- Do not diagnose, hypothesize, or propose fixes inline. The `/fix-bug` skill's evidence-first protocol exists to prevent premature conclusions.
- After `/fix-bug` completes and the fix is confirmed, return to Step 2 — re-present remaining test items.

### Step 4: Confirm Complete

Only after the user confirms all test items pass:
- Mark the [TEST] task complete
- Proceed to [DOCS] and [RETRO] tasks
