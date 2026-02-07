---
name: update-docs
description: Update project documentation after feature implementation. Use at [DOCS] checkpoints or when asked to update docs.
user-invocable: true
argument-hint: Brief summary of what changed (e.g., "Added sticks, nuts, shells; BFS pathfinding fix")
model: sonnet
agent: true
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
---

## Update Documentation

You are updating project documentation after a feature implementation or bug fix. You will be given a summary of what changed as your argument.

### Files to Update

Read each of these files, then apply only the changes warranted by the summary:

| File | Audience | Update Rules |
|------|----------|-------------|
| `README.md` | **Players** | Latest Updates section only. Player-visible changes only. No implementation details, no specific counts or enumerations (not "seven variants" — just "multiple variants" or omit). One line per feature, plain language. |
| `CLAUDE.md` | **AI context (always loaded)** | Current features line and roadmap only. Generalize into categories, never enumerate specifics. Keep lightweight — detailed docs exist elsewhere. |
| `docs/game-mechanics.md` | **Detailed reference** | Add new mechanics, update existing sections. Never duplicate config values — reference config source instead (e.g., "See `config.GroundSpawnInterval`"). This includes approximate values like "~5 days" — if the number comes from config, reference the config. |
| `docs/architecture.md` | **Developer reference** | Update design patterns, data flow, code organization. Implementation details welcome here. |
| `docs/gardening-phase-plan.md` | **Planning artifact** | Mark completed steps with ✅. Add notes about bugs found/fixed during testing. Do not modify future steps. |

### Principles

1. **Audience awareness**: Each doc has a different reader. Apply the rules above strictly.
2. **No config duplication**: Never enumerate specific config values (spawn counts, stack sizes, thresholds) in docs. Reference the config source.
3. **Minimal changes**: Only add/update what the summary warrants. Do not reorganize, rewrite, or "improve" existing content.
4. **Consistency**: Match the style and formatting of the existing content in each file.
5. **No new files**: Only edit existing files listed above. If a file doesn't need changes, skip it.

### Process

1. Read all five files
2. For each file, determine what (if anything) needs updating based on the summary
3. Make edits
4. Report what was changed in each file (or "no changes needed")
