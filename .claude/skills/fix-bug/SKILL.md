---
name: fix-bug
description: "Investigate and fix bugs or gaps surfaced during human testing. Evidence first, agree on the problem, scan siblings, then fix. Trigger phrases: 'fix bug', 'investigate issue', 'something went wrong'."
user-invocable: true
argument-hint: Description of the observed bug or unexpected behavior
---

## Fix Bug

Investigate and fix a bug or gap reported during human testing. The goal is to agree on what the problem is before any fixing happens.

### Step 1: Gather Evidence

**Do this BEFORE forming any hypothesis about the cause.**

- Examine the most recently modified save file: `ls -t ~/.petri/worlds/*/state.json | head -1`
- Check logs, add `t.Logf` or `-v` to relevant tests if needed
- Never guess what the game state is — read it

### Step 2: Agree on the Problem

- Restate the user's observation **in their words**
- Present the evidence gathered — what the save file, logs, or diagnostics show
- If the expected behavior is unclear, consult `docs/game-mechanics.md` for how the system should work from the player's perspective
- If evidence doesn't match the report, ask clarifying questions — don't assume the user is misreporting. Ambiguity in what they observed is more likely than a wrong report
- **Stop here and wait for user confirmation that the problem statement is correct.** Do not propose fixes yet.

### Step 3: Scan Sibling Flows and Check Coverage

Before fixing anything:

- **Sibling flows:** Identify flows that share the same structure as the affected code. Check each for the same issue. Report all affected flows — fixing one at a time leads to whack-a-mole cascades. `docs/architecture.md` documents which flows share patterns — use it as a starting point, then trace the actual code to confirm.
- **End-to-end test coverage:** Does an automated end-to-end test exist for this flow? If not, write one before fixing — it catches remaining bugs in the same pass and prevents regressions.

### Step 4: Classify and Route

**Gap** (missing behavior, not a bug):
- Small gaps: confirm the gap is worth solving now and discuss scope before writing code, then fix in place
- Design-level gaps: invoke `/refine-feature`

**Bug** (code doesn't match intended behavior):
- Proceed to Step 5

**Any fix beyond a simple bug fix** — design changes, new behaviors, changed constants, scope decisions — must be reflected in the phase design doc or step spec before implementation.

### Step 5: Fix

1. **Write regression tests first** — cover the reported issue and any sibling flows found in Step 3
2. Fix the minimum code to pass tests
3. Run full test suite: `go test ./...`
4. Run `gofmt ./...`

### Step 6: Escalation Check

**Second bug in the same feature?** Stop patching. Multiple bugs in the same workflow signal a design problem, not an implementation problem.
- Restate the intended end-to-end flow
- Evaluate whether the design is sound before fixing further
- Surface the pattern to the user before continuing

### Step 7: Re-test

After the fix is confirmed:
- Re-run the full [TEST] checklist from the step spec — not just the specific fix
- Offer `/test-world` if the scenario warrants a fresh test world
- Wait for user confirmation before marking the fix-bug task complete
