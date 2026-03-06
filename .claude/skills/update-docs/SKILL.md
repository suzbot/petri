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
| `README.md` | **Players** | Primarily Latest Updates section. Other sections only if absolutely needed. Player-visible changes only. Bug fixes and internal improvements are NOT Latest Updates material — only new player-facing capabilities or workflows. Intended behavior that was broken is a bug fix, not a new capability — even if the player couldn't access it before. No implementation details, no specific counts or enumerations (not "seven variants" — just "multiple variants" or omit). **One short line per feature** — Max 20 words. Only describe behavior implemented in the current step — do not describe planned future capabilities. Name the capability, not its steps (e.g., "Water garden orders" not "character procures vessel, fills at water, walks to tiles..."). If details matter, they belong in game-mechanics.md. Only if a major new workflow, make a brief update to 'How it works'. **Sub-steps:** If the summary describes a sub-step (e.g., "4a", "sub-step X of N", or infrastructure-only with no complete player workflow), skip README — only update when the full step produces player-visible, end-to-end behavior. |
| `CLAUDE.md` | **AI context (always loaded)** | Roadmap section only: keep "Up Next" to phase name and status only (e.g., "Construction Phase: In progress"). Link to design doc and step-spec. No step numbers, no feature lists, no completed-work summaries — the step-spec is self-tracking. |
| `docs/game-mechanics.md` | **Behavioral reference** | See game-mechanics rules below. |
| `docs/architecture.md` | **AI developer reference** | Design patterns, decision rationale, and "adding new X" checklists. **Include:** new patterns/categories, decision rules (e.g., when to use self-managing vs ordered), `continueIntent` interaction rules for new actions, new checklists for recurring tasks. **Exclude:** API reference (function tables, parameter lists, file-to-function maps) — these belong in code comments and are discoverable via code navigation. The test is: does this capture a *design decision or rule* that can't be inferred from reading the code? If yes, it belongs. If it's just documenting what functions exist, it doesn't. |
| `docs/flow-diagrams.md` | **AI developer reference** | Visual call graphs, intent priority hierarchy, and multi-phase action state machines. Update when decision-layer routing changes (new buckets, new multi-phase actions, changed priority order). Don't update for pure implementation changes that don't alter call structure. |
| `docs/*-design.md` | **Phase design doc** | Update the relevant step's **Status** to "Complete". Only mark status "Complete" when the full step is done. Sub-step completion → no status change. Do not modify other steps' content. Do not add implementation details — the design doc stays brief and declarative. |
| `docs/step-spec.md` | **Current step spec** | No changes during [DOCS] — `/implement-feature` handles clearing this file on completion. |
| `docs/randomideas.md`| **Small Feature Planning**` | Remove items that have been moved into plans or that have been completed.|
| `docs/triggered-enhancements.md` | **Deferred Feature Planning** | Remove items that have been completed. Update items where intended approaches have changed.|

### game-mechanics.md Rules

**Audience:** The `/remind-me` skill (Grep-based lookup), users wanting to understand the game, and Claude during phase planning. All content should answer: "what does the game do?" — not "how is the code structured?" (that's architecture.md).

**Placement:** The doc is organized into ~14 sections by gameflow. Add new content to the existing section that covers that system. Don't create new top-level sections unless a genuinely new game system is introduced — and if so, place it in gameflow order. When unsure where something goes, read the Table of Contents first.

**Detail level — include:**
- Player-visible behavior and mechanics
- How systems interact from the player's perspective (e.g., "hunger tier affects food selection")
- Stat thresholds, scoring formulas, and tier tables that are frequently referenced
- Config references for exact values (e.g., "see `config.WetGrowthMultiplier`")

**Detail level — exclude:**
- Code flow descriptions (function call order, handler logic, "CalculateIntent returns nil then...")
- Internal helper names (ConsumePlantable, EnsureHasVesselFor, etc.)
- Log message exact strings or colors
- UI rendering implementation details (ANSI, lipgloss, style names)
- Anything that only matters to someone reading the code — it belongs in architecture.md or code comments

**The test:** Would a user or the `/remind-me` skill benefit from this information? If yes, include it. If only a developer modifying the code would care, it belongs elsewhere.

**No config duplication:** Never enumerate specific config values (spawn counts, stack sizes, duration numbers) in the doc. Reference the config source instead (e.g., "See `config.GroundSpawnInterval`"). This includes approximate world-time equivalents — write `(see config.ItemMealSize)` not `~5 world minutes`. Exception: stat tier thresholds are kept inline because they're referenced so frequently (noted with an HTML comment in the doc).

### Principles

1. **Audience awareness**: Each doc has a different reader. Apply the rules above strictly.
2. **No config duplication**: Never enumerate specific config values (spawn counts, stack sizes, thresholds) in docs. Reference the config source. This includes approximate world-time equivalents derived from config — write `(see config.ItemMealSize)` not `~5 world minutes`. If the number traces back to a constant, reference the constant.
3. **Minimal changes**: Only add/update what the summary warrants. Do not reorganize, rewrite, or "improve" existing content.
4. **Consistency**: Match the style and formatting of the existing content in each file.
5. **No new files**: Only edit existing files listed above. If a file doesn't need changes, skip it.

### Process

1. Read all files
2. For each file, determine what (if anything) needs updating based on the summary
3. Make edits
4. **IMPORTANT**: Your final message MUST include a summary of what was changed in each file (or "no changes needed"). This summary is the only output the caller sees — without it, they have no visibility into what you did. 
  - Format as a bulleted list per file.
  - Include a suggestion for a short commit message
