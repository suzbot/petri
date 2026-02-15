Name: Values

Description: A record of observations about the principles and values that are important to the user in this project.

## Consistency Over Local Cleverness

When a new case resembles an existing pattern, follow the existing pattern's shape — even if an ad-hoc inline solution is faster to write. Diverging from the pattern creates conceptual debt and inconsistency that will need to be reconciled later. If the existing pattern doesn't quite fit, that's a discussion point, not a reason to quietly do something different.

Examples: Kind on ItemVariety (mirrors Edible on ItemVariety). EnsureHasPlantable (mirrors EnsureHasItem). ConsumePlantable (mirrors ConsumeAccessibleItem).

## Source of Truth Clarity

Every piece of data should live in exactly one place. Don't reconstruct a field from surrounding context when the entity itself should carry that field. "Where does this value rightfully live?" is a design question worth asking explicitly before writing code.

Example: item.Kind belongs on ItemVariety, not reconstructed from order.targetType.
