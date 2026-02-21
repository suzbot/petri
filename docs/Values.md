Name: Values

Description: A record of observations about the principles and values that are important to the user in this project.

## Consistency Over Local Cleverness

When a new case resembles an existing pattern, follow the existing pattern's shape — even if an ad-hoc inline solution is faster to write. Diverging from the pattern creates conceptual debt and inconsistency that will need to be reconciled later. If the existing pattern doesn't quite fit, that's a discussion point, not a reason to quietly do something different.

Examples: Kind on ItemVariety (mirrors Edible on ItemVariety). EnsureHasPlantable (mirrors EnsureHasItem). ConsumePlantable (mirrors ConsumeAccessibleItem).

## Source of Truth Clarity

Every piece of data should live in exactly one place. Don't reconstruct a field from surrounding context when the entity itself should carry that field. "Where does this value rightfully live?" is a design question worth asking explicitly before writing code.

Example: item.Kind belongs on ItemVariety, not reconstructed from order.targetType.

## Reuse Before Invention

Before creating a new utility, helper, or pattern, ask: "Does something existing already handle this, or can it be extended?" Even if the answer is no, the question itself often reveals the right shape for the new code by identifying what's adjacent. The cost of a brief analysis is always lower than the cost of discovering redundant abstractions later.

Example: FindVesselContaining — checked whether FindAvailableVessel could serve the need first. It couldn't (inverted semantics), but the analysis confirmed the new utility should be a structural sibling.

## Design Types for Future Siblings

When introducing a new type or category, ask: "What else will eventually live alongside this?" and design the type hierarchy to accommodate those future siblings — even if only one member exists today. A narrow type that fits only the current case creates refactoring debt when the next sibling arrives.

Example: Water as ItemType "liquid", Kind "water" (not ItemType "water"). Future beverages (mead, beer, wine) become Kind variants under the same ItemType, with no structural changes needed.

## Requirements as Ground Truth

Implementation and tests should trace back to the stated requirement. When a test validates code structure ("returns ActionPickup") rather than user intent ("vessel gets filled with water"), it's testing the wrong thing. The requirement is the ground truth — code and tests are accountable to it, not the other way around.

Example: Fetch Water tests initially validated that ground-vessel pickup returned ActionPickup. But the requirement was "get the vessel and fill it as one continuous activity." The structural test passed while the intent was broken.

## Step Back on Cascading Bugs

Multiple bugs in the same workflow signal a design problem, not an implementation problem. Fix the design, don't keep patching symptoms. The cost of one step-back conversation is always lower than the cost of another round of testing, bug-finding, and context consumption.

Example: Fetch Water went through four rounds of testing with new bugs each time. Each fix was correct in isolation, but the underlying two-roll design was wrong. A step back after the second bug would have saved two rounds.

## Idle Activities Should Be Non-Destructive

Idle activities — the things characters choose to do on their own when no needs or orders are pressing — should never destroy player-relevant state. Need-driven behavior (eating, drinking, sleeping) may consume or displace things to ensure survival, and that's appropriate for a simulation. But idle choices should be safe by default.

Example: Fetch Water only triggers when a character has or can find an empty vessel. It never dumps existing vessel contents to make room for water.

## Demonstrate Alignment, Don't Assert It

When claiming a new feature follows an existing pattern, cite the specific evidence — from architecture.md or from the code itself. "This follows the ordered action pattern" is an assertion; "This follows the ordered action pattern (architecture.md 'Action Categories' section) — handler clears intent after each work unit, resumption via AssignedOrderID bypass in selectIdleActivity" is a demonstration. If the architecture doc doesn't cover the relevant distinction, that's a signal to update it (via `/refine-feature`, `/new-phase`, or a general agent — not during implementation).

Example: ActionWaterGarden was initially proposed as self-managing, then as ordered, with assertions of consistency both times. The actual code trace through CalculateIntent → selectIdleActivity → AssignedOrderID revealed which pattern was correct and why.

## Evidence Before Reasoning

When something fails unexpectedly, gather evidence first — run a diagnostic, add logging, check with `-v` — then reason about causes. Reasoning without evidence leads to circular re-derivation of the same candidate explanations. The cost of one diagnostic run is always lower than the cost of three rounds of speculative reasoning.

Example: WaterGarden integration tests failed for unclear reasons. Three rounds of reasoning about positioning and pathfinding produced wrong fixes. One diagnostic pass with `t.Logf` immediately revealed that `findWaterGardenIntent` was returning nil because `IsWet` treated tiles adjacent to water as already wet.
