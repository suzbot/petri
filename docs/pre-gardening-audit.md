# Pre-Gardening Extensibility Audit

## Purpose

Before implementing Simple Gardening, audit the codebase to reduce touchpoints when adding new actions, orders, recipes, and items. This makes Gardening implementation smoother and sets up better patterns for Construction, Threats, and Pigment features that follow.

## Context from Future Features

Reviewed requirements for upcoming features to understand common patterns:

**Gardening-Reqs.txt** introduces:
- Inventory expansion to 2 slots (explicit prerequisite)
- New item types: Seeds, Shells, Sticks, Hoe
- New orderable activities: Craft Hoe, Till Soil, Plant, Water Garden
- Area selection UI for tilling
- Activity structuring explicitly called out as a prerequisite

**Construction-Reqs.txt** introduces:
- Stackable items (bundles of grass/sticks)
- Clay tiles and gathering
- Multi-recipe activities (Fence/Hut with thatch/stick/brick variants)
- Material durability

**Pigment-Reqs.txt** introduces:
- Component procurement pattern (get vessel + flowers before crafting)
- Item decoration system
- Face painting (character appearance modification)

**Common pattern across all**: Check inventory for components → drop non-components → seek components (weighted by preference/distance) → perform activity OR abandon. This mirrors existing harvest/craft patterns but will be used by many more activities.

## Extension Exercise: Adding "Water Plant" Action

Walking through a hypothetical "water plant" action to count touchpoints:

### Files That Would Need Changes

1. **internal/entity/character.go** - Add `ActionWater` to ActionType enum
2. **internal/entity/activity.go** - Add "waterGarden" activity to ActivityRegistry
3. **internal/system/order_execution.go** - Add case to `findOrderIntent` switch for "waterGarden"
4. **internal/system/order_execution.go** - New `findWaterGardenIntent` function (~80 lines if following harvest pattern)
5. **internal/ui/update.go** - Add `case entity.ActionWater:` to applyIntent switch (~20-50 lines)
6. **internal/simulation/simulation.go** - Add `case entity.ActionWater:` to applyIntent switch (if tests need it)
7. **internal/entity/item.go** - Possibly add "Plantable" attribute checks
8. **internal/game/spawning.go** or **variety_generation.go** - If water/tilled soil are new features/states

**Total: 6-8 files, ~150-200 lines of new code**

### Patterns Observed

**Duplicated vessel-seeking logic** (~70 lines each):
- `foraging.go:findForageIntent` (lines 267-320)
- `order_execution.go:findHarvestIntent` (lines 115-236)

Both do: check if carrying vessel → look for compatible vessel → create intent to move/pickup vessel. Differences: target finding method, conflict resolution (skip vs drop), log category.

**Duplicated applyIntent logic**:
- `ui/update.go:applyIntent` (lines 521-852, ~330 lines)
- `simulation/simulation.go:applyIntent` (lines 110-210, ~160 lines)

simulation.go is intentionally simpler (missing ActionLook, ActionTalk, ActionCraft). It serves as a lightweight test harness. **Recommendation**: Keep separate; simulation.go can stay minimal.

**Order completion logic scattered in update.go**:
- Harvest completion: lines 733-765 (checks vessel full, removes order)
- Craft completion: lines 840-846 (removes order after crafting)
- Each new order type adds more conditionals

## Findings

### Priority 1: Inventory Expansion (BLOCKER)

**Status**: Must complete before gardening

Gardening-Reqs explicitly lists as prerequisite: "Give character 2 inventory slots."

Current state: `Carrying *Item` (single slot)

**Impact**: Every inventory check, pickup logic, and drop logic needs updating. Better to do this as focused work before adding new activities.

### Priority 2: Pickup/Drop Unification

**Status**: Triggered by gardening - recommend doing BEFORE gardening

From triggered-enhancements.md: "Unify pickup/drop code (picking.go)" with trigger "Adding order/activity that involves picking up items."

Gardening adds 4+ activities that involve picking up items:
- Craft Hoe (pickup stick + shell)
- Till Soil (pickup/carry hoe)
- Plant (pickup plantable items)
- Fetch Water (fill vessel at pond)

**Recommendation**: Create `internal/system/picking.go` before gardening with:
- Unified vessel-seeking pattern (encapsulates the ~70 lines duplicated between foraging and harvest)
- `PickupIntentBuilder` with configurable target-finding and conflict-resolution strategies
- Move vessel helpers from foraging.go (`AddToVessel`, `IsVesselFull`, `CanVesselAccept`, `FindAvailableVessel`)

**Estimated savings**: Each new pickup-based activity saves ~40-50 lines by reusing the builder pattern.

### Priority 3: Activity Structuring

**Status**: Current pattern is good; minor improvements possible

Current ActivityRegistry pattern is clean. Adding a new activity requires:
1. Registry entry in activity.go (~10 lines)
2. Case in findOrderIntent switch (~3 lines)
3. Intent-finding function (~50-100 lines depending on complexity)

**Recommendation**: No major refactor needed. The Gardening-Reqs note about "ensure Activities are structured for easy addition" is largely already satisfied by the existing registry pattern.

### Priority 4: Order Completion Refactor

**Status**: Keep as triggered enhancement - do DURING gardening if it becomes unwieldy

Currently order completion is in update.go's applyIntent with activity-specific conditionals. With 2 current order types (harvest, craft), this is manageable.

Gardening adds: Till Soil, Plant, Water Garden (3 more order types)

**Recommendation**: Monitor during gardening. If completion logic exceeds ~50 lines of conditionals, refactor to a completion handler pattern:
```go
type OrderCompletionHandler func(char *Character, order *Order, result PickupResult) bool
var OrderCompletionHandlers = map[string]OrderCompletionHandler{...}
```

### Priority 5: Item Type Extensibility

**Status**: Keep as triggered enhancement (after gardening)

From triggered-enhancements.md: "ItemType constants" trigger is "Adding new item types."

Gardening adds: seed, shell, stick, hoe. This may reach the threshold where compile-time safety becomes valuable.

**Recommendation**: Evaluate after gardening. If ItemType string comparisons become error-prone, formalize to constants.

### Priority 6: Category Formalization

**Status**: Keep as triggered enhancement (during/after Construction)

From triggered-enhancements.md: "Category type formalization" trigger is "Adding non-plant categories (tools, crafted items)."

Gardening introduces hoe (tool). Construction introduces more non-plant items.

**Recommendation**: Evaluate during Construction phase when multiple non-plant categories exist.

## Scope

### Do BEFORE Gardening

| Refactor | Estimated Effort | Benefit |
|----------|------------------|---------|
| Inventory expansion (2 slots) | Medium | Required prerequisite |
| Pickup/drop unification (picking.go) | Medium | Saves ~40-50 lines per new pickup activity |

### Do DURING Gardening

| Item | Notes |
|------|-------|
| New item types (seed, shell, stick, hoe) | Natural part of feature |
| New activities (Craft Hoe, Till, Plant, Water) | Using existing patterns |
| Order completion refactor | Only if complexity warrants |
| Area selection UI | New UI pattern for tilling |

### Keep as Triggered (AFTER Gardening)

| Enhancement | Trigger Point |
|-------------|--------------|
| ItemType constants | If string comparisons become error-prone |
| Category formalization | When multiple non-plant categories exist (Construction) |
| applyIntent unification | simulation.go intentionally simpler; keep separate |
| Order completion criteria refactor | If scattered logic exceeds ~50 lines |
| Performance optimizations | When noticeable lag or profiling shows bottlenecks |

## Out of Scope (Unchanged)

- **Knowledge/Learning patterns** - Wait until Enhanced Learning feature is closer
- **Comprehensive test restructuring** - Observation tests were just tuned; only fix obviously flaky tests
- **UI refactoring** - Unless directly blocking extensibility

## Success Criteria

- [x] Extension exercise completed with file/line counts
- [x] Identified duplication patterns and quantified savings
- [x] Findings from triggered-enhancements.md integrated
- [x] Clear before/during/after categorization
- [ ] Inventory expansion implemented
- [ ] Pickup/drop unification implemented (picking.go)
- [ ] Patterns established benefit Construction, Threats, and Pigment features

## Next Steps

1. Discuss and confirm before/during/after categorization
2. Implement inventory expansion (2 slots)
3. Implement picking.go with vessel-seeking pattern unification
4. Begin gardening implementation with reduced touchpoints
