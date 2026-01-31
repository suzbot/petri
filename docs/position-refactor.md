# Position Struct Refactor

**Status:** In Progress (Phase 4 Complete)
**Goal:** Replace separate `x, y` / `posX, posY` coordinate pairs with a unified `Position` struct

## Key Discovery

A `Pos` struct already exists in `internal/game/map.go`:
```go
type Pos struct {
    X, Y int
}
```
Currently used as map keys for spatial indexing. This should be promoted to the canonical position type.

## Proposed Position Type

```go
// internal/types/position.go (or extend existing Pos)
type Position struct {
    X, Y int
}

func (p Position) DistanceTo(other Position) int {
    return abs(p.X-other.X) + abs(p.Y-other.Y)  // Manhattan
}

func (p Position) IsAdjacentTo(other Position) bool {
    dx, dy := abs(p.X-other.X), abs(p.Y-other.Y)
    return dx <= 1 && dy <= 1 && !(dx == 0 && dy == 0)
}

func (p Position) IsCardinallyAdjacentTo(other Position) bool {
    dx, dy := abs(p.X-other.X), abs(p.Y-other.Y)
    return (dx == 1 && dy == 0) || (dx == 0 && dy == 1)
}

func (p Position) NextStepToward(target Position) Position {
    return Position{p.X + sign(target.X-p.X), p.Y + sign(target.Y-p.Y)}
}
```

## Scope Summary

### Structs to Change

| Struct | File | Current Fields | New Field |
|--------|------|----------------|-----------|
| `BaseEntity` | `entity/entity.go` | `X, Y int` | `Pos Position` |
| `Intent` | `entity/character.go` | `TargetX, TargetY, DestX, DestY` | `Target, Dest Position` |
| `Character` | `entity/character.go` | `LastLookedX, LastLookedY` | `LastLooked Position` |
| `CharacterSave` | `save/state.go` | `X, Y int` | `Pos Position` |
| `ItemSave` | `save/state.go` | `X, Y int` | `Pos Position` |
| `FeatureSave` | `save/state.go` | `X, Y int` | `Pos Position` |

### Function Categories

| Category | Count | Example |
|----------|-------|---------|
| Map query methods | 14 | `EntityAt(x, y)` → `EntityAt(pos)` |
| Intent finding | 22+ | `findFoodIntent(char, cx, cy, ...)` → `findFoodIntent(char, pos, ...)` |
| Entity constructors | 7 | `NewCharacter(id, x, y, ...)` → `NewCharacter(id, pos, ...)` |
| Distance calcs | 12+ sites | Inline `abs(cx-ix) + abs(cy-iy)` → `pos.DistanceTo(other)` |
| Adjacency checks | 3 funcs | `isAdjacent(x1, y1, x2, y2)` → `pos.IsAdjacentTo(other)` |

### Files Requiring Changes

**Core (11 files):**
- `internal/types/position.go` (new)
- `internal/entity/entity.go`
- `internal/entity/character.go`
- `internal/game/map.go`
- `internal/save/state.go`
- `internal/ui/serialize.go`
- `internal/system/movement.go`
- `internal/system/foraging.go`
- `internal/system/talking.go`
- `internal/system/order_execution.go`
- `internal/ui/update.go`

**Secondary (tests, misc):**
- All `_test.go` files that construct entities or check positions
- `internal/system/lifecycle.go`
- `internal/system/crafting.go`

## Code Reduction Opportunities

1. **Duplicate `abs()` functions** - exists in both `map.go` and `movement.go`
2. **Repeated distance patterns** - 12+ inline Manhattan calculations
3. **Adjacency logic** - repeated across multiple files

## Implementation Plan

### Phase 1: Foundation ✓
- [x] Create `internal/types/position.go` with Position type and methods
- [x] Add `Abs()` and `Sign()` helpers to position package (exported)
- [x] Write tests for Position methods (23 tests)

### Phase 2: Entity Layer ✓
- [x] Update `BaseEntity` to use `Pos()` method returning `types.Position`
- [x] Rename `Position()` to `Pos()` returning `types.Position`
- [x] Rename `SetPosition()` to `SetPos()` taking `types.Position`
- [ ] Update entity constructors (deferred - still use x, y ints)

### Phase 3: Intent System ✓
- [x] Update `Intent` struct fields (Target, Dest as types.Position)
- [x] Update all intent-finding functions to use Position parameter
- [x] Update selectIdleActivity, selectOrderActivity to use Position

### Phase 4: Map Operations ✓
- [x] Update Map query methods
- [x] Remove duplicate `Pos` struct (use types.Position)
- [x] Update spatial indexing maps

### Phase 5: Serialization
- [ ] Update save state structs
- [ ] Update ToSaveState/FromSaveState
- [ ] Verify save/load compatibility

### Phase 6: Cleanup
- [ ] Remove duplicate `abs()` functions
- [ ] Replace inline distance calculations with method calls
- [ ] Update tests

### Phase 7: Documentation
- [ ] Update `docs/architecture.md` with Position type usage patterns
- [ ] Update `CLAUDE.md` Key Files section to include `types/position.go`
- [ ] Document when to use Position methods vs inline calculations
- [ ] Add examples of correct Position usage for future reference

## Serialization Compatibility

Current JSON structure:
```json
{"x": 5, "y": 10}
```

Must maintain this format. Position struct needs:
```go
type Position struct {
    X int `json:"x"`
    Y int `json:"y"`
}
```

## Notes

- Existing `Pos` in `map.go` is already the right shape
- All distance calculations use Manhattan distance
- Intent has two positions: immediate target (next step) and destination (final goal)

## Documentation Updates (Phase 7 Details)

After refactor completion, add the following to project docs:

### For CLAUDE.md (Key Files section)
```
- `internal/types/position.go` - Position struct with distance/adjacency methods
```

### For architecture.md (new section or add to existing)
```markdown
## Position Handling

All coordinates use `types.Position`:
- Entity positions: `entity.Pos()` returns Position
- Distance: `pos.DistanceTo(other)` (Manhattan distance)
- Adjacency: `pos.IsAdjacentTo(other)`, `pos.IsCardinallyAdjacentTo(other)`
- Pathfinding: `pos.NextStepToward(target)`

**Do not:**
- Use separate x, y parameters for positions
- Inline distance calculations (`abs(x1-x2) + abs(y1-y2)`)
- Create new position-like structs
```

### Rationale
This ensures future sessions:
1. Know Position exists before proposing new coordinate patterns
2. Use existing methods instead of reimplementing distance/adjacency
3. Maintain consistency across the codebase

## Session Log

### Session 1 (2026-01-31)
- Analyzed codebase scope
- Found existing `Pos` struct in map.go
- Identified ~50+ functions needing updates
- Created this tracking document

### Session 2 (2026-01-31)
- Completed Phase 1: Foundation
- Created `internal/types/position.go` with Position struct and methods
- Added exported `Abs()` and `Sign()` helpers
- Wrote 23 tests covering all Position methods
- Added JSON tags for serialization compatibility
- Completed Phase 2: Entity Layer
- Renamed `Position()` to `Pos()`, `SetPosition()` to `SetPos()`
- Updated 82+ call sites across 22 files
- All tests passing

### Session 3 (2026-01-31)
- Completed Phase 3: Intent System
- Updated Intent struct: `TargetX/Y, DestX/Y` → `Target, Dest types.Position`
- Updated 57 Intent literals across 8 files
- Updated 10 find*Intent function signatures to use Position parameter
- Updated selectIdleActivity and selectOrderActivity signatures
- Added types import to foraging.go, order_execution.go, talking.go, idle.go
- Used pos.DistanceTo() in findHealingIntent, findTalkIntent, findForageTarget
- All tests passing

### Session 4 (2026-01-31)
- Completed Phase 4: Map Operations
- Removed duplicate `Pos` struct from map.go
- Updated internal maps to use `types.Position` as key:
  - `entities map[types.Position]entity.Entity`
  - `characterByPos map[types.Position]*entity.Character`
- Updated 14 Map query method signatures to use `types.Position`:
  - EntityAt, CharacterAt, ItemAt, FeatureAt, DrinkSourceAt, BedAt
  - IsValid, IsOccupied, IsBlocked, IsEmpty
  - FindNearestDrinkSource, FindNearestBed
  - MoveEntity, MoveCharacter
- Removed local `abs()` function, using `pos.DistanceTo()` instead
- Updated canFulfillThirst, canFulfillEnergy signatures
- Updated 70+ call sites across 15 files:
  - map_test.go, world.go, movement.go, update.go, view.go
  - simulation.go, simulation_test.go, lifecycle.go, lifecycle_test.go
  - talking_test.go, consumption_test.go, serialize_test.go, order_execution_test.go
- All tests passing
