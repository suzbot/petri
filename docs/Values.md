Name: Values

Description: Design principles observed and refined through retrospectives. Collaboration norms live in CLAUDE.md.

## Anchor to Intent, Not Structure

Every phase of work — planning, refinement, implementation, testing, fixing — should trace to what the user or character experiences, not to code structure.

**Planning and refinement:** Prefer to break work into iterative chunks of deliverable value, over technical layers. Each chunk should have an anchor story — a 1-2 sentence narrative of what the player or character experiences — that grounds implementation details in "why" and gives testing a north star.

**Testing:** A test that validates "returns ActionPickup" checks code structure; "vessel gets filled with water after action completes" validates intent. Anchor tests to the story, not the implementation path.

**Fixing:** A bug fix that changes what the player experiences is a feature change, not a fix. A plan whose causal model is wrong produces correct code that addresses the wrong cause. Describe any proposed change in terms of what the player experiences — a fix that sounds neutral technically ("sort map keys") may be a behavior change in gameplay terms ("every character discovers activities in the same fixed order").

Examples: Fetch Water tests validated ActionPickup returns while the actual requirement — continuous vessel-fill activity — was broken. A flaky discovery test was fixed by retrying in the test (preserving production randomness), not by sorting iteration order (removing emergent variety). Pathing plan said displacement fixes thrashing; one human test showed the real cause was BFS convergence, requiring a different approach.

## Isomorphism

The code structure should mirror what it represents. If the game has a concept, there should be a clear place in the code where that concept lives. Someone asking "how does a character decide what to eat?" should find the answer in a place that makes intuitive sense — not buried in the pathfinding file.

This is the *why* behind many structural decisions: intent.go exists because decision-making is a distinct concept from locomotion. discretionary.go exists because leisure is a distinct motivation from orders. apply_actions.go exists because execution is a distinct concern from orchestration.

## Follow the Existing Shape

When a new case resembles an existing pattern, give preference to that pattern — even if an ad-hoc solution is faster. Divergence creates conceptual debt. If the pattern doesn't quite fit, that's a discussion point, not a reason to quietly diverge.

**Before creating something new**, ask: "Does something existing handle this, or can it be extended?" Even when the answer is no, the question reveals the right shape by identifying what's adjacent.

**Every piece of data should live in one place.** Don't reconstruct a field from surrounding context when the entity should carry it. "Where does this value rightfully live?" is a design question worth asking before writing code.

**Don't duplicate logic across call sites.** When multiple actions need the same behavior, extract a shared helper rather than copying the logic. Duplication means bug fixes and behavior changes must be applied in multiple places — and they won't be. (EnsureHasVesselFor, FindNextVesselTarget, and the pickup helpers in picking.go all exist because multiple actions needed the same procurement/continuation logic.)

**When fixing a gap, check sibling flows.** A gap in one flow likely exists in every flow with the same structure. Ground water vessel support was missing from helpWater — but fetchWater and waterGarden had the same gap because all three share the "ensure I have a vessel with water" pattern.

Examples: Kind on ItemVariety mirrors Edible on ItemVariety. FindVesselContaining checked whether FindAvailableVessel could serve the need first — it couldn't, but the analysis confirmed the new utility should be a structural sibling. item.Kind belongs on ItemVariety, not reconstructed from order.targetType.

## Consider Extensibility

When introducing a new type, category, or pattern, ask: "What else will eventually live alongside this?" Design to accommodate future siblings — even if only one member exists today. A narrow type that fits only the current case creates refactoring debt when the next sibling arrives. This applies to type hierarchies, action categories, registry patterns, and any structure that's likely to grow.

Example: Water as ItemType "liquid", Kind "water" (not ItemType "water"). Future beverages become Kind variants under the same ItemType with no structural changes.

## Start With the Simpler Rule

When designing character behavior, default to the simplest version that addresses the observed problem. Don't add conditions without a concrete observed reason. A conditional rule that isn't grounded in something the player actually experiences is complexity with no benefit.

Example: Displacement on any collision vs. only stationary characters — "always sidestep" is simpler, matches real behavior, and avoids a tracking mechanism with no observable upside.
