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
