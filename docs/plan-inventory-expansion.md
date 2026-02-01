# Plan: Pre-Gardening Refactors

## Overview

Two refactors to complete before implementing Gardening features:

1. **Inventory Expansion** - Expand from 1 slot to 2 slots
2. **Pickup/Drop Unification** - Consolidate duplicated vessel-seeking pattern into `picking.go`

This document tracks context across sessions.

---

## Part 1: Inventory Expansion

### Goal

Replace `Carrying *Item` (single slot) with `Inventory []*Item` (2 slots, expandable later).

### Design Decisions

- **Data structure**: `Inventory []*Item` with capacity 2
- **Semantics**: Each slot holds one item OR one container (vessel with contents)
- **Backward compatibility**: Existing logic for "pick up until full" should work with 2 slots
- **No over-engineering**: Don't add gardening-specific logic yet, just expand capacity

### Files to Modify

Based on grep for `.Carrying`:

| File | Changes Needed |
|------|----------------|
| `internal/entity/character.go` | Replace `Carrying *Item` with `Inventory []*Item`, update `IsInventoryFull()`, add helpers |
| `internal/save/state.go` | Update `CharacterSave` struct |
| `internal/ui/serialize.go` | Update `ToSaveState`/`FromSaveState` |
| `internal/ui/serialize_test.go` | Update test for inventory serialization |
| `internal/system/foraging.go` | Update `Pickup()`, `Drop()`, `CanPickUpMore()`, `findForageIntent()` |
| `internal/system/foraging_test.go` | Update tests |
| `internal/system/consumption.go` | Update inventory clearing after eating |
| `internal/system/consumption_test.go` | Update tests |
| `internal/system/order_execution.go` | Update `findHarvestIntent()`, `findCraftVesselIntent()` |
| `internal/system/order_execution_test.go` | Update tests |
| `internal/system/movement.go` | Update `findFoodIntent()` carried item checks |
| `internal/system/movement_test.go` | Update tests |
| `internal/ui/update.go` | Update `applyIntent` for ActionPickup, ActionConsume, ActionCraft |
| `internal/ui/update_test.go` | Update tests |
| `internal/ui/view.go` | Update character details display |
| `internal/simulation/simulation.go` | Update `applyIntent` consume logic |

### New Helper Methods (character.go)

```go
// Inventory capacity
const InventoryCapacity = 2

// HasInventorySpace returns true if character has at least one empty slot
func (c *Character) HasInventorySpace() bool

// AddToInventory adds item to first empty slot, returns false if full
func (c *Character) AddToInventory(item *Item) bool

// RemoveFromInventory removes item from inventory, returns false if not found
func (c *Character) RemoveFromInventory(item *Item) bool

// GetCarriedVessel returns first vessel in inventory, or nil
func (c *Character) GetCarriedVessel() *Item

// GetCarriedItem returns first non-vessel item, or nil (for eating carried food)
func (c *Character) GetCarriedItem() *Item

// FindInInventory returns item matching predicate, or nil
func (c *Character) FindInInventory(predicate func(*Item) bool) *Item
```

### Migration Strategy

1. Add new `Inventory` field alongside `Carrying` temporarily
2. Add helper methods that work with `Inventory`
3. Migrate each file one at a time, switching from `Carrying` to helpers
4. Remove `Carrying` field once all references updated
5. Update serialization to handle both old (single Carrying) and new (Inventory slice) formats for save compatibility

### Key Behavioral Changes

| Current | New |
|---------|-----|
| `char.Carrying != nil` means full | `!char.HasInventorySpace()` means full |
| `char.Carrying = item` to add | `char.AddToInventory(item)` |
| `char.Carrying = nil` to remove | `char.RemoveFromInventory(item)` |
| Single item display | Loop through inventory slots |

### Test Strategy

- Update existing tests to use new helpers
- Add tests for 2-slot scenarios:
  - Pick up item when one slot occupied
  - Pick up until both slots full
  - Drop from specific slot
  - Vessel in one slot, loose item in other

---

## Part 2: Pickup/Drop Unification (picking.go)

### Goal

Create `picking.go` with composable prerequisite helpers that encapsulate the common "check inventory → find on map → drop if needed → create pickup intent" pattern. This reduces code in each activity's intent-finding function by ~40-50 lines and establishes patterns for Gardening, Construction, and Pigment phases.

### Design: Composable Helpers with Shared Core

After analyzing patterns across Gardening-Reqs, Construction-Reqs, and Pigment-Reqs, we chose composable helpers over a full builder pattern because:
- Patterns aren't fully orthogonal (some combinations don't make sense)
- Clear, specific function names are easier to understand and debug
- Easy to add new helpers as patterns emerge without modifying existing code
- Incremental cost is low (~30 lines per new helper)

### Architecture

**Layer 1: Low-level operations** (moved from foraging.go)
- `Pickup`, `Drop`, `DropItem` - actual pickup/drop actions
- `AddToVessel`, `IsVesselFull`, `CanVesselAccept`, `FindAvailableVessel`, `CanPickUpMore` - vessel utilities

**Layer 2: Internal shared core** (new)
```go
// Internal: common check → find → drop → pickup logic
type pickupParams struct {
    filter       func(*Item) bool
    scorer       func(*Item) float64
    dropConflict bool
    category     string
    activityName string
}
func findAndCreatePickupIntent(char, pos, items, gameMap, log, params) *Intent
```

**Layer 3: Public prerequisite helpers** (new)
```go
// EnsureHasItem returns intent to get item by type, or nil if already have it
func EnsureHasItem(char, itemType, items, gameMap, log, dropConflict, category) *Intent

// EnsureHasItemByPreference returns intent to get item weighted by preference/distance
func EnsureHasItemByPreference(char, itemType, items, gameMap, log, dropConflict, category) *Intent

// EnsureHasVesselFor returns intent to get vessel compatible with target, or nil if ready
func EnsureHasVesselFor(char, target, items, gameMap, log, dropConflict, category) *Intent

// EnsureHasRecipeInputs returns intent to get next missing recipe input, or nil if ready
func EnsureHasRecipeInputs(char, recipe, items, gameMap, log, category) *Intent
```

### How This Simplifies Activity Code

**Current `findHarvestIntent`** (~120 lines) becomes:
```go
func findHarvestIntent(...) *Intent {
    target := findNearestItemByType(...)
    if target == nil { return nil }

    // One line handles all vessel-seeking logic
    if intent := EnsureHasVesselFor(char, target, items, gameMap, log, "order"); intent != nil {
        return intent
    }

    return createHarvestPickupIntent(char, pos, target, ...)
}
```

**Future `findCraftHoeIntent`** will be simple:
```go
func findCraftHoeIntent(...) *Intent {
    recipe := entity.RecipeRegistry["craft-hoe"]

    // One line handles multi-component gathering
    if intent := EnsureHasRecipeInputs(char, recipe, items, gameMap, log, "order"); intent != nil {
        return intent
    }

    return &Intent{Action: ActionCraft, ...}
}
```

### Files to Modify

| File | Changes |
|------|---------|
| `internal/system/picking.go` | NEW - vessel helpers + internal core + public helpers |
| `internal/system/foraging.go` | Remove vessel helpers (moved to picking.go) |
| `internal/system/order_execution.go` | Refactor `findHarvestIntent` to use `EnsureHasVesselFor` |

### Move to picking.go

From `foraging.go`:
- `AddToVessel()`
- `IsVesselFull()`
- `CanVesselAccept()`
- `FindAvailableVessel()`
- `CanPickUpMore()`
- `Pickup()`, `PickupResult`
- `Drop()`, `DropItem()`

Keep in `foraging.go`:
- `findForageIntent()` - uses unified scoring, not the prerequisite pattern
- `scoreForageItems()`, `scoreForageVessels()`, `hasMatchingGrowingItems()`
- `findForageItemIntent()`, `createPickupIntent()`
- `FindNextVesselTarget()`

### Future Extensions (not in scope now)

Additional helpers will be added as needed:
- `EnsureHasBundle` - for Construction bundles (discuss approach first, see Construction-Reqs.txt)
- `EnsureHasVesselContaining` - for Pigment (vessel with flowers inside)
- `EnsureHasFromCategory` - for Pigment decoration

---

## Progress Tracking

### Part 1: Inventory Expansion

- [x] Add `Inventory []*Item` field and helpers to character.go
- [x] Update foraging.go (Pickup, Drop, CanPickUpMore, findForageIntent, FindNextVesselTarget)
- [x] Update foraging_test.go
- [x] Update consumption.go (eating from inventory)
- [x] Update movement.go (findFoodIntent carried item checks)
- [x] Update order_execution.go (harvest and craft intents)
- [x] Update ui/update.go (applyIntent cases)
- [x] Update ui/view.go (inventory display)
- [x] Update simulation/simulation.go
- [x] Update serialization - backward compat bridge done, saves first item for compat
- [x] Update remaining tests (consumption_test, movement_test, order_execution_test, update_test, serialize_test)
- [x] Remove old `Carrying` field
- [x] Manual testing (all confirmed)

### Part 2: Pickup/Drop Unification

- [x] Create picking.go, move vessel helpers from foraging.go
- [x] Add `EnsureHasVesselFor` helper
- [x] Refactor `findHarvestIntent` to use `EnsureHasVesselFor`
- [x] Update/add tests (9 tests for EnsureHasVesselFor)
- [ ] Add `EnsureHasItem` helper (deferred - add when needed)
- [ ] Add `EnsureHasItemByPreference` helper (deferred - add when needed)
- [ ] Add `EnsureHasRecipeInputs` helper (deferred - add when needed)
- [ ] Manual testing

---

## Session Notes

*(Add notes here as work progresses across sessions)*

### Session 1 (2026-01-31)
- Created plan document
- Analyzed current codebase: 18 files reference `.Carrying`
- Design decisions finalized: `Inventory []*Item` with capacity 2
- Ready to begin implementation

### Session 1 continued - Implementation started
**Completed:**
- Added `Inventory []*Item` field to Character struct (kept `Carrying` temporarily for migration)
- Added `InventoryCapacity = 2` constant
- Added 6 helper methods: `HasInventorySpace()`, `AddToInventory()`, `RemoveFromInventory()`, `GetCarriedVessel()`, `GetCarriedItem()`, `FindInInventory()`
- Updated `IsInventoryFull()` to use new Inventory slice
- Wrote 20 new tests for inventory helpers (all passing)
- Updated foraging.go core functions to use new helpers
- Added `DropItem()` function for dropping specific items
- Updated foraging_test.go and partial order_execution_test.go
- Updated ui/view.go inventory panel to show all slots with format "Inventory: N/2 slots"
- Updated ui/serialize.go for backward-compatible save/load:
  - Load: migrates old `Carrying` field → new `Inventory` slice
  - Save: writes first item from `Inventory` → `Carrying` field (temporary until save format updated)
  - Note: Currently only saves/restores 1 item - full 2-slot serialization still needed

**User testing confirmed:**
- Character successfully forages into second slot when first slot occupied
- Inventory view correctly shows "Inventory: 2/2 slots" with both items listed
- Created test world `world-test-inv2` for testing 2-slot behavior

**Remaining files by usage count:**
| File | Usages | Status |
|------|--------|--------|
| ui/serialize.go | 18 | Partially done (backward compat bridge, not full 2-slot support) |
| ui/update.go | 11 | Not started |
| system/movement.go | 8 | Not started |
| ui/view.go | 5 | Done |
| system/order_execution.go | 5 | Partially updated (tests) |
| simulation/simulation.go | 4 | Not started |
| system/consumption.go | 2 | Not started |

**Current test status:** Some tests failing in order_execution_test.go and ui/update_test.go due to still using old `Carrying` field

**Next steps when resuming:**
1. ~~Continue migrating remaining files (movement.go, consumption.go, order_execution.go, update.go, simulation.go)~~ Done
2. Update save format to support full 2-slot inventory (currently saves only first item)
3. ~~Update remaining tests~~ Done
4. Remove deprecated `Carrying` field
5. Manual testing of all features

### Session 2 (2026-02-01)
**Completed:**
- Migrated all remaining files to use new `Inventory` slice and helpers
- Files updated: consumption.go, movement.go, order_execution.go, update.go, simulation.go
- Updated all tests to use `AddToInventory()`, `GetCarriedVessel()`, `FindInInventory()` instead of `char.Carrying`
- Tests passing: consumption_test.go, movement_test.go, order_execution_test.go, update_test.go, serialize_test.go
- All tests pass including race detection

**Current status:**
- Inventory expansion functionally complete - all code uses new `Inventory` slice
- `Carrying` field still exists but is deprecated and unused (kept for migration safety)
- Serialization works but only saves/restores first inventory item (needs update for full 2-slot support)

**Remaining before marking inventory expansion complete:**
1. ~~Update serialization to save/restore full 2-slot inventory~~ Done
2. ~~Remove deprecated `Carrying` field from character.go~~ Done
3. Manual testing:
   - ~~Foraging 2nd slot~~ ✓ (confirmed Session 1)
   - ~~Save/load with 2 items in inventory~~ ✓ (confirmed Session 3)
   - ~~Crafting vessel~~ ✓ (confirmed Session 3)
   - ~~Harvesting fills both slots~~ ✓ (confirmed Session 3 after bug fix)

### Session 3 (2026-02-01)
**Completed:**
- Updated `CharacterSave` in state.go: added `Inventory []ItemSave`, kept `Carrying *ItemSave` for backward compat
- Updated `ToSaveState`: now saves all inventory items to new `Inventory` field
- Updated `FromSaveState`: loads from `Inventory` (new format), falls back to `Carrying` (old format)
- Added `TestFromSaveState_RestoresTwoSlotInventory` test for 2-slot round-trip
- Updated remaining test files: consumption_test.go, movement_test.go, update_test.go
- Removed deprecated `Carrying` field from Character struct
- All tests pass including race detection

**Ready for manual testing:**
- Foraging 2nd slot (previously confirmed working)
- Save/load with 2 items in inventory
- Harvesting with vessel
- Crafting

### Session 3 continued - Bug fix and foraging redesign

**Bug found during manual testing:**
- Harvest order completed after picking up 1 item, even with empty inventory slot
- Root cause: `update.go` didn't check for inventory space before completing order
- Fix: Added `FindNextHarvestTarget()` in order_execution.go to continue harvesting until inventory full
- Added test `TestApplyIntent_HarvestOrderWithoutVessel_ContinuesUntilInventoryFull`
- **Manual testing: PASSED** - Harvest orders now fill both inventory slots

**Foraging redesign (to prevent world resource stripping):**

Problem: With 2 slots and vessel-filling, eager foragers could strip resources too quickly.

New design - foraging completes after collecting ONE growing item:
- Fetching a vessel doesn't complete foraging (it's setup)
- Picking up one growing item (loose or into vessel) completes foraging
- More casual, interleaves with other idle activities

Unified scoring for target selection (growing items + vessels):
- Growing items: `netPreference * prefWeight - distance * distWeight`
- Empty vessels: `vesselBonus - distance * distWeight`
- Partial vessels: `vesselBonus + netPreference(contents) - distance * distWeight`
  - Only scored if matching growing items exist on map

Vessel bonus scales with hunger: `prefWeight * (1 - hunger/100)`
- Uses same base value as single preferred attribute (+1 valence * prefWeight)
- Low hunger → higher vessel bonus (willing to plan ahead)
- High hunger → lower vessel bonus (grab immediate food)

Emergent behaviors:
- Satiated characters fetch vessels, hungry characters grab loose food
- Partial vessels ignored if contents disliked or no matching growers
- Empty vessels opportunistically grabbed when convenient
- Unused vessels dropped when other tasks need the slot

**Implementation completed:**
- Modified `update.go`: Only continue vessel filling for orders, not autonomous foraging
- Rewrote `findForageIntent` in foraging.go with unified scoring
- Added helper functions: `scoreForageItems`, `scoreForageVessels`, `hasMatchingGrowingItems`, `createPickupIntent`
- Removed old `findForageTarget` function (replaced by `scoreForageItems`)
- Updated tests in foraging_test.go and movement_test.go
- All tests pass including race detection
- Created test world `world-harvest-test` for manual testing

**Ready for manual testing:**
- Autonomous foraging stops after one growing item (not filling vessel)
- Vessel vs loose item selection based on hunger-scaled scoring
- Partial vessel only picked up if matching growers exist

**Foraging manual testing: PASSED**

**Additional fixes from manual testing feedback:**

1. **Crafted items auto-drop** (update.go)
   - Crafted vessels now drop on ground instead of staying in inventory
   - Any character can use them for their next task

2. **Crafting can use vessel contents** (order_execution.go)
   - Added generic `extractRecipeInputFromVessel(char, recipe)` function
   - Checks vessel contents against recipe inputs
   - Extracts matching item to inventory for crafting
   - Works for any recipe, not just gourd→vessel

3. **Foraging frequency** - confirmed working as designed
   - Empty vessels allow foraging (CanPickUpMore returns true)
   - 25% primary chance is acceptable

**Test worlds created:**
- `world-harvest-test` - harvest order testing
- `world-craft-test` - crafting from vessel contents (character holding vessel with gourds)

**Manual testing: PASSED**
- Crafted vessel drops on ground ✓
- Craft order uses gourd from inside carried vessel ✓

### Session 3 continued - Documentation updates
**Completed:**
- Updated README.md latest updates section
- Updated game-mechanics.md: 2-slot inventory, craft auto-drop, foraging redesign
- Updated architecture.md: code example for foraging vs orders distinction
- Removed backward compatibility code for old `Carrying` field (state.go, serialize.go)

**Part 1: Inventory Expansion - COMPLETE**

### Session 4 (2026-02-01) - Part 2: Pickup/Drop Unification

**TDD Implementation Plan:**

**Phase A: Move existing functions to picking.go**
Functions to move from `foraging.go`:
- `AddToVessel`, `CanVesselAccept`, `IsVesselFull`, `FindAvailableVessel`
- `CanPickUpMore`, `Pickup`, `PickupResult` + constants
- `Drop`, `DropItem`

Existing tests in `foraging_test.go` verify correctness (same package).

**Phase B: Add new prerequisite helpers (TDD)**
Write tests first, then implement:
1. `EnsureHasVesselFor(char, target, items, gameMap, log, dropConflict, category) *Intent`
   - Returns intent to get vessel for target, or nil if ready
2. `EnsureHasItem(char, itemType, items, gameMap, log, dropConflict, category) *Intent`
   - Returns intent to get item by type, or nil if already have it
3. `EnsureHasItemByPreference(char, itemType, items, gameMap, log, dropConflict, category) *Intent`
   - Weighted by preference/distance
4. `EnsureHasRecipeInputs(char, recipe, items, gameMap, log, category) *Intent`
   - Returns intent to get next missing recipe input

**Phase C: Refactor findHarvestIntent**
Replace ~90 lines of vessel-seeking logic with `EnsureHasVesselFor` call.

**Progress:**
- [x] Phase A: Created picking.go with moved functions (AddToVessel, CanVesselAccept, IsVesselFull, FindAvailableVessel, CanPickUpMore, Pickup, Drop, DropItem)
- [x] Phase B: Added EnsureHasVesselFor with 9 tests
- [x] Phase C: Refactored findHarvestIntent (~80 lines reduced to ~5)
- All tests pass including race detection

**Deferred helpers:**
The following helpers were planned but deferred since they're not yet needed:
- `EnsureHasItem` - returns intent to get item by type
- `EnsureHasItemByPreference` - weighted by preference/distance
- `EnsureHasRecipeInputs` - returns intent to get next missing recipe input

These can be added when needed for Gardening, Construction, or Pigment phases.

---

## References

- [Gardening-Reqs.txt](Gardening-Reqs.txt) - Feature requirements driving these refactors
- [pre-gardening-audit.md](pre-gardening-audit.md) - Audit that identified these refactors
- [architecture.md](architecture.md) - Current patterns to follow
- [triggered-enhancements.md](triggered-enhancements.md) - Related deferred items
