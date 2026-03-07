# Step Spec: Step 7 — Construct Interaction

Design doc: [construction-design.md](construction-design.md)

---

## Sub-step 7a: Look at Constructs + Construct Preference Formation

**Anchor story:** A character wanders past a stick fence and pauses to look at it. They form an opinion — "Likes stick fences" appears in their preferences panel and action log. Another character in a bad mood looks at a brick fence and develops "Dislikes brick fences." A character who already likes stick fences looks at one and their mood improves slightly.

### Scope

**Entity layer:**

- `TargetConstruct *entity.Construct` field on Intent struct (ephemeral, not serialized — follows `TargetBuildPos` pattern)
- `PreferenceKind() string` method on Construct — returns lowercase composed identity for preference matching:
  - "grass" → "thatch fence", "stick" → "stick fence", "brick" → "brick fence"
  - Computed from material display name (lowercase) + " " + Kind
  - This is the construct's recipe-level identity, parallel to Item.Kind for crafted items ("shell hoe", "hollow gourd")

**Preference layer (entity/preference.go + system/preference.go):**

- `MatchesConstruct(c *Construct) bool` on Preference:
  - If Kind is set: compare against `c.PreferenceKind()`
  - If Color is set: compare against `c.MaterialColor`
  - Empty preference matches nothing (same guard as Matches)
  - ItemType, Pattern, Texture: if set, return false (constructs don't have these item attributes)
- `MatchScoreConstruct(c *Construct) int` on Preference — returns `Valence * AttributeCount()` if matches, 0 otherwise (same pattern as MatchScore)
- `NetPreferenceForConstruct(c *Construct) int` on Character — sums MatchScoreConstruct across all preferences (same pattern as NetPreference)
- `TryFormConstructPreference(char, construct, log) (PreferenceFormationResult, *Preference)` in system/preference.go:
  - Same mood-gated formation as `TryFormPreference` (calls `getFormationParams`)
  - `rollConstructPreferenceType(construct, valence)` picks from available attributes:
    - Kind (construct identity): always available — "stick fence", etc.
    - Color (material color): always available
    - Solo vs combo roll uses same `PrefFormationWeightSingle` config
    - Solo: pick one of Kind or Color randomly
    - Combo: Kind + Color (e.g., "brown stick fences")
  - Duplicate/opposite check against existing preferences (same logic as TryFormPreference)

**Looking system (system/looking.go):**

- `CompleteLookAtConstruct(char *Character, construct *Construct, log *ActionLog)`:
  - Log "Looked at [DisplayName]"
  - Update last-looked position (`LastLookedX`, `LastLookedY`, `HasLastLooked`) — same position-based mechanism as items
  - Mood adjustment from existing preferences: `char.NetPreferenceForConstruct(construct)` × `MoodPreferenceModifier`, with mood tier transition logging
  - Call `TryFormConstructPreference` for new preference formation
  - No discovery triggers (DD-36 — deferred to Step 8)

**Intent finding (system/intent.go — findLookIntent):**

- Current: searches items only via `findNearestItemExcluding`
- Extended: also search constructs via `gameMap.Constructs()`, excluding last-looked position
  - Find nearest item (existing logic)
  - Find nearest construct (new: iterate constructs, skip last-looked position, track nearest by distance)
  - Pick whichever is closer; if tie, prefer item (arbitrary but deterministic)
- When targeting a construct:
  - Activity string: "Looking at " + construct.DisplayName() / "Moving to look at " + construct.DisplayName()
  - Intent: `Action: ActionLook`, `TargetConstruct: construct`, `TargetItem: nil`
  - Dest/Target: adjacent tile (using `findClosestAdjacentTile` — same as item path), since constructs are impassable

**Handler (ui/apply_actions.go — applyLook):**

- Current: checks `TargetItem` for walking phase adjacency, calls `CompleteLook` on completion
- Extended: add parallel branch for `TargetConstruct`:
  - Walking phase: if `TargetConstruct != nil`, check adjacency to construct position. If not adjacent, `moveWithCollision`. Return.
  - Looking phase: accumulate `ActionProgress` until `LookDuration` (same duration as items)
  - On completion: call `system.CompleteLookAtConstruct(char, construct, log)`, clear intent, set idle cooldown

### Architecture patterns

- **TargetConstruct on Intent**: follows the typed-target-field pattern — `TargetItem`, `TargetFeature`, `TargetCharacter`, `TargetWaterPos`, `TargetBuildPos` are all precedents. Ephemeral, not serialized.
- **CompleteLookAtConstruct**: parallel to `CompleteLook` — same structure (log, last-looked update, mood adjustment, preference formation), different entity type. Follows **Isomorphism** — constructs are not items.
- **Preference.Kind for recipe identity**: follows "shell hoe" / "hollow gourd" pattern. The construct's `PreferenceKind()` returns the same kind of string that crafted items carry in their Kind field. This is the foundation for preference-weighted recipe selection in Step 10.
- **Nearest-lookable search**: extends `findLookIntent` to search two entity pools (items + constructs) and pick the nearest. Position-based exclusion works across entity types.

### Anti-patterns to avoid

- Do NOT put construct data into TargetItem — constructs are not items. Use the dedicated TargetConstruct field.
- Do NOT call TryFormPreference (item-based) with construct data — use TryFormConstructPreference. The formation logic rolls from different attributes (Kind + Color from construct, not ItemType + Color + Pattern + Texture from item).
- Do NOT add discovery triggers in CompleteLookAtConstruct — deferred to Step 8 (DD-36).
- Do NOT serialize TargetConstruct — it's ephemeral like TargetBuildPos.

### Tests

- Unit: `Construct.PreferenceKind()` returns correct composed identity ("stick fence", "thatch fence", "brick fence")
- Unit: `Preference.MatchesConstruct` — Kind-only preference matches correct construct, Color-only matches correct color, combo matches both, ItemType-only preference does NOT match constructs
- Unit: `Preference.MatchScoreConstruct` — returns Valence × AttributeCount when matching, 0 when not
- Unit: `Character.NetPreferenceForConstruct` — sums across multiple preferences correctly
- Unit: `TryFormConstructPreference` — forms Kind preference, Color preference, or combo based on roll; respects mood gating (neutral mood → no formation); handles duplicate/opposite removal
- Unit: `CompleteLookAtConstruct` — updates last-looked position, adjusts mood from existing preferences, calls formation
- Unit: `findLookIntent` with both items and constructs on map — returns intent targeting nearest entity (item or construct)
- Unit: `findLookIntent` with only constructs (no items) — returns intent targeting nearest construct
- Unit: `findLookIntent` excludes last-looked position for constructs
- Regression: existing Look-at-item tests still pass after findLookIntent changes

### [TEST] Checkpoint

Human testing:
- Create test world with: 2-3 characters (varied moods — one happy, one unhappy, one neutral), several fences of different materials (stick, thatch, brick), some items on the ground for comparison
- Verify: characters look at fences (not just items) — "Looking at Stick Fence" appears in activity
- Verify: preference formation — "Likes stick fences" or "Dislikes brick fences" appears in action log and preferences panel
- Verify: mood adjustment — character with existing fence preference gets mood change when looking at matching fence
- Verify: characters alternate between looking at items and constructs (don't get stuck on one type)
- Verify: last-looked exclusion works (don't look at same fence twice in a row)
- Verify: details panel still shows correct construct info when cursor is on a fence (existing Step 5 behavior, confirm not broken)

### [DOCS]
Run /update-docs after human testing confirms success.

### [RETRO]
Run /retro after docs are updated.

---

## Implementation Sequencing Notes

### Why a single sub-step
Step 7 is a cohesive feature with no separable infrastructure phase. The entity layer (TargetConstruct, PreferenceKind), preference layer (matching, formation), looking system (intent, handler), and completion handler all need each other to produce any testable behavior. Splitting would create sub-steps with no human-testable checkpoint.

### Deferred from Step 7
- **Discovery from looking at constructs**: DD-36 — revisit for Step 8 (huts). Looking at a fence could trigger "build hut" know-how discovery.
- **Preference-weighted recipe selection**: Step 10 scope. These preferences lay the foundation — a character who "likes stick fences" will prefer the stick-fence recipe when that selection logic is built.
