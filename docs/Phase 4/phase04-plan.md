# Phase IV Implementation Plan: Basic Knowledge and Transmission

## Overview

Implement character knowledge system, knowledge-based behavior, and knowledge transmission via talking from `docs/Phase 4/phase04reqs.txt`. Characters gain knowledge through experience, act on that knowledge, and can share it through conversation.

## Collaboration Approach

Following patterns from Phase 3:

- **Discuss before implementing**: Each step begins with discussion/confirmation
- **Human testing touchpoints**: Marked with [TEST] - pause for manual testing with rebuilt binary
- **Small iterations**: Steps sized for quick feedback loops
- **TDD approach**: Tests first, then implementation

---

## High-Level Decision Needed

**Knowledge Data Structure**: Should Knowledge store attribute references (like Preference) or a simple description string?

| Approach | Pros | Cons |
|----------|------|------|
| **Attribute refs** | Enables future partial matching, consistent with Preference pattern | More complex |
| **Description string** | Simpler, matches "full variety description" in requirements | Harder to extend later |

**Recommendation**: Attribute refs - maintains consistency with Preference struct and enables future flexibility.

---

## User Decisions

- **Knowledge Data Structure**: Attribute refs (matches Preference pattern, enables matching logic and future flexibility)
- **Key hint placement**: Details panel (may redesign as characters become more complex)
- **ESC key behavior**: Return to start screen (not quit game)
- **Empty knowledge panel**: Show title, otherwise appear empty

---

## Opportunistic Items (from CLAUDE.md roadmap)

Items to address during relevant sub-phases:

| Sub-Phase | Opportunistic Item |
|-----------|-------------------|
| C (Knowledge Panel UI) | ESC key: return to start screen vs quit |
| D (Action Log) | Action log audit: vertical space, limits; L log alignment with action log |

---

## Sub-Phase A: Knowledge by Experience (Req A)

### Requirement
> A. Character gains knowledge by experience
>    1. When a character eats a poisonous item, they gain the knowledge "[items] are poisonous"
>    2. When a character eats a healing item, they gain the knowledge "[items] are healing"

### User-Facing Outcome
- Player watches character eat a poisonous Spotted Red Mushroom
- Character now "knows" that Spotted Red Mushrooms are poisonous
- (Knowledge affects behavior in later requirements - avoiding poison, seeking healing)

### Implementation Steps

**A1. Create Knowledge type**

**File**: `internal/entity/knowledge.go` (new)

```go
type KnowledgeCategory string
const (
    KnowledgePoisonous KnowledgeCategory = "poisonous"
    KnowledgeHealing   KnowledgeCategory = "healing"
)

type Knowledge struct {
    Category    KnowledgeCategory
    ItemType    string
    Color       types.Color
    Pattern     types.Pattern
    Texture     types.Texture
}
```

Methods:
- `Description() string` - e.g., "Spotted Red Mushrooms are poisonous"
- `Matches(item *Item) bool` - check if knowledge applies to an item
- `Equals(other Knowledge) bool` - for duplicate checking

**Tests**: Unit tests for knowledge matching and description

**A2. Add Knowledge slice to Character**

**File**: `internal/entity/character.go`

- Add `Knowledge []Knowledge` field
- Add `HasKnowledge(k Knowledge) bool` helper
- Add `LearnKnowledge(k Knowledge) bool` helper (returns false if already known)

**Tests**: Unit tests for knowledge management

**A3. Learn poison knowledge on consumption**

**File**: `internal/system/consumption.go`

- After poison effect applied, create knowledge entry from item attributes
- Call `char.LearnKnowledge()`

**Tests**: Integration tests for learning on poisoned item consumption

**A4. Learn healing knowledge on consumption**

**File**: `internal/system/consumption.go`

- After healing effect applied, create knowledge entry from item attributes
- Call `char.LearnKnowledge()`

**Tests**: Integration tests for learning on healing item consumption

[TEST] **Human Testing**: Use debugger or temporary log to verify character's Knowledge slice contains entry after eating poison/healing item. (UI visibility comes in Sub-Phase C.)

---

## Sub-Phase C: Knowledge Panel UI

### C1. Add knowledge panel toggle

**Discuss**: Key binding choice (K suggested in reqs), panel placement

**File**: `internal/ui/update.go`

- Add key handler for knowledge panel toggle in select mode

**File**: `internal/ui/model.go`

- Add `showKnowledgePanel bool` field

### C2. Render knowledge panel

**File**: `internal/ui/view.go`

- When toggled, replace action log with knowledge list
- Show scrollable list of character's knowledge entries
- Show hint text for toggling back

**File**: `internal/ui/styles.go`

- Add styles for knowledge panel if needed

[TEST] **Human Testing**: Verify knowledge panel toggle works, displays correctly

---

## Sub-Phase D: Poison Knowledge Creates Dislike

### Requirement
> D. Poison knowledge creates dislike
>    1. When the character learns that something is poisonous, it is affected by a negative opinion for the full variety description
>    2. If a fully matching 'like' preference exists, remove it
>    3. If no fully matching preference exists, create dislike
>    4. If a fully matching dislike exists, nothing happens
>    5. Confirm characters avoid disliked items at less urgent tiers

### User-Facing Outcome
- Character eats poisonous Spotted Red Mushroom
- Character learns "Spotted red mushrooms are poisonous"
- Character forms "Dislikes spotted red mushrooms" preference (full variety)
- At Moderate hunger, character will avoid spotted red mushrooms (seek other food)

### Implementation Steps

**D1. Add `NewFullPreferenceFromItem` to entity/preference.go**

Creates preference with ALL item attributes (full variety description):
- ItemType, Color, Pattern (if present), Texture (if present)

**Tests**: Unit test for full preference creation

**D2. Add `FormDislikeFromKnowledge` to system/preference.go**

- Takes character, item, and log
- Creates full-variety dislike preference
- Handles three cases:
  - Exact-match "like" exists → remove it (log "No longer likes...")
  - No match exists → create dislike (log "New Opinion: Dislikes...")
  - Exact-match dislike exists → no change
- Reuses existing logging functions

**Tests**:
- Unit test: remove existing like
- Unit test: create new dislike
- Unit test: already dislikes (no change)

**D3. Call from consumption.go**

After `char.LearnKnowledge(knowledge)` succeeds for poison, call `FormDislikeFromKnowledge`

**Tests**: Integration test: eat poison → knowledge + dislike formed

**D4. Verify avoidance logic in movement.go**

Review `findFoodTarget()` - disliked items should be filtered at Moderate hunger via existing valence logic. May need no changes.

[TEST] **Human Testing**:
- Verify character eating poison forms dislike
- Verify character with existing "like" loses it when eating poison
- Verify character avoids disliked items at Moderate hunger

---

## Sub-Phase E: Health as Seeking Trigger

### E1. Add health to intent priority calculation

**Discuss**: Where does health fit in the priority order? Same tier comparison as other stats?

**File**: `internal/system/movement.go`

- Modify `CalculateIntent()` to include health tier in priority evaluation
- Health triggers seeking only when character has healing knowledge

### E2. Create findHealingIntent function

**File**: `internal/system/movement.go`

- New function to find known healing items
- Only considers items the character knows are healing
- Uses similar pattern to `findFoodIntent()` but filtered by knowledge

**Tests**: Unit tests for healing intent creation

---

## Sub-Phase F: Healing Knowledge in Food Selection

### F1. Add healing valence bonus when health is low

**Discuss**: How much bonus? Only when health is driving stat, or always when health < 100?

**File**: `internal/system/movement.go`

- Modify `findFoodTarget()` or healing-specific search
- Add bonus to gradient score for items known to be healing
- Only applies when character has knowledge of the item's healing property

**File**: `internal/config/config.go`

- Add healing bonus config constant

**Tests**: Integration tests for healing-aware food selection

[TEST] **Human Testing**: Verify injured characters seek known healing items

---

## Sub-Phase G: Talking Activity

### G1. Add ActionTalk enum and talking state fields

**File**: `internal/entity/character.go`

- Add `ActionTalk ActionType`
- Add `TalkingWith *Character` field (track conversation partner)
- Add `TalkTimer float64` field

### G2. Create findTalkIntent function

**File**: `internal/system/movement.go`

- New function to find nearby idle characters
- "Idle activities" = Idle, Looking, Talking (per requirements)
- Target must be adjacent (including diagonals)
- If not adjacent, move toward target

### G3. Integrate talking as idle activity choice

**Discuss**: Probability distribution between Idle, Looking, Talking when no needs?

**File**: `internal/system/movement.go`

- Modify idle activity selection to include talking option
- Refactor idle/looking selection into general "idle activity" system

### G4. Handle talking action in game loop

**File**: `internal/ui/update.go`

- When ActionTalk is applied:
  - Set both characters to talking state
  - Start 5-second timer for both
  - Update CurrentActivity for both
  - Can only be interrupted by Moderate+ needs

**Tests**: Integration tests for talking initiation and duration

[TEST] **Human Testing**: Verify characters find each other and talk

---

## Sub-Phase H: Knowledge Transmission

### H1. Implement knowledge sharing on talk completion

**File**: `internal/system/talking.go` (new) or extend consumption.go pattern

- When talking completes (timer expires):
  - If talker has knowledge, offer random piece to partner
  - If partner doesn't have it, they learn it
  - Log transmission if it occurs

### H2. Add transmission logging

**File**: Same as H1

- Log "[Name] shared knowledge with [Partner]"
- Log "[Name] learned [knowledge] from [Partner]"

[TEST] **Human Testing**: Verify knowledge spreads between characters through talking

---

## Critical Files Summary

| File | Changes |
|------|---------|
| `internal/entity/knowledge.go` | NEW - Knowledge type and methods |
| `internal/entity/character.go` | Knowledge slice, TalkingWith, TalkTimer |
| `internal/system/consumption.go` | Learning on eat, preference from knowledge |
| `internal/system/movement.go` | Talking intent, health in priority, healing food selection |
| `internal/system/talking.go` | NEW - Knowledge transmission logic |
| `internal/ui/model.go` | showKnowledgePanel |
| `internal/ui/view.go` | Knowledge panel render, learned keyword coloring |
| `internal/ui/update.go` | K key toggle, ActionTalk handling |
| `internal/config/config.go` | Talking duration, healing bonus constants |

---

## Session Progress Tracking

### Current Status: In Progress

| Sub-Phase | Status | Notes |
|-----------|--------|-------|
| A: Knowledge by Experience | Complete | |
| B: Knowledge Panel UI | Complete | + ESC key behavior |
| C: Action Log ("learned something!") | Complete | + Action log vertical space fix, Full log (L) patterns |
| D: Poison Knowledge → Dislike | Complete | Existing avoidance logic handles filtering |
| E-F: Healing Knowledge → Seeking + Food Selection | Not Started | |
| G: Talking Activity | Not Started | |
| H: Knowledge Transmission | Not Started | |

---

## Workflow Summary

Each step follows this pattern:

1. **Discuss** - Brief confirmation before implementing
2. **Implement** - Tests first, then code changes
3. **[TEST]** - Human testing touchpoint (rebuild binary first!)

Sub-phases are designed to be independently testable. Complete one before starting the next.
