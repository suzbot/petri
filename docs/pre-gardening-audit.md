# Pre-Gardening Extensibility Audit

## Purpose

Before implementing Simple Gardening, audit the codebase to reduce touchpoints when adding new actions, orders, recipes, and items. This makes Gardening implementation smoother and sets up better patterns for Construction, Threats, and Pigment features that follow.

## Scope

### Priority 1: Action Type Extensibility

**Goal**: Minimize files touched when adding a new action (e.g., "water plant", "plant seed")

**Areas to examine**:
- `internal/entity/intent.go` - ActionType enum, how new actions are defined
- `internal/system/intent.go` - Intent calculation, action selection logic
- `internal/ui/update.go` - applyIntent switch statement
- `internal/simulation/simulation.go` - applyIntent mirror (test harness)
- `internal/ui/view.go` - Action display/rendering

**Questions to answer**:
- How many files need changes to add one new action?
- Could a registry pattern consolidate action definitions?
- Is the simulation.go duplication of applyIntent necessary, or can it be shared?

### Priority 2: Order Type Extensibility

**Goal**: Minimize files touched when adding a new order type (e.g., "plant", "water")

**Areas to examine**:
- `internal/entity/order.go` - Order struct, OrderType enum
- `internal/system/order_execution.go` - Order assignment, completion logic
- `internal/system/intent.go` - Order-to-intent mapping
- `internal/ui/update.go` - Order UI integration

**Questions to answer**:
- How are orders mapped to intents currently?
- Can order execution be more data-driven?

### Priority 3: Item Type Extensibility

**Goal**: Minimize files touched when adding new item types (seeds, construction materials, pigments)

**Areas to examine**:
- `internal/entity/item.go` - Item struct, item categories
- `internal/entity/variety.go` - ItemVariety, variety generation
- `internal/game/variety_generation.go` - How varieties are created
- `internal/game/spawning.go` - Item spawning logic
- `internal/config/config.go` - Item-related constants

**Questions to answer**:
- How are item behaviors (edible, craftable, plantable) determined?
- Is there a clean way to add item "capabilities" without scattered conditionals?

### Priority 4: Recipe Extensibility

**Goal**: Ensure recipe system can easily accommodate new crafting recipes

**Areas to examine**:
- `internal/entity/recipe.go` - Recipe struct, RecipeRegistry
- `internal/system/crafting.go` - Crafting execution

**Questions to answer**:
- Is the current recipe registry pattern sufficient?
- Any hardcoded assumptions that would break with new recipes?

## Out of Scope (Defer)

- **Knowledge/Learning patterns** - Wait until Enhanced Learning feature is closer
- **Comprehensive test restructuring** - Observation tests were just tuned; only fix obviously flaky tests
- **UI refactoring** - Unless directly blocking extensibility
- **Performance optimization** - Not a current concern

## Suggested Approach

**Extension Exercise**: Walk through adding a hypothetical "water plant" action without writing code:

1. List every file that would need modification
2. Note what kind of change each file needs (new enum value, new case in switch, new function, etc.)
3. Identify patterns where changes could be consolidated
4. Propose specific refactors with clear before/after touchpoint counts

## Success Criteria

- Clear list of refactors with estimated touchpoint reduction
- Each refactor is small and testable independently
- Patterns established will benefit Construction, Threats, and Pigment features
- No regressions in existing functionality

## Output

After completing this audit, update this document with:
- [ ] Findings from extension exercise
- [ ] Proposed refactors (prioritized)
- [ ] Decision on which refactors to implement before vs. during Gardening
