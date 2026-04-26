This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## UI Improvements

1. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
2. **Natural Language Descriptions**
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."

## Tech Stuff

## Flaky Tests

1. **TestSpawnClay_CountInRange** (internal/game/world_test.go) — intermittently fails when clay spawn count falls outside 6-10 range. Probabilistic bounds may be too tight.

## Bugs to Investigate

1. **Fill vessel abandoned mid-trip** — Observed character picking up a vessel for water, heading toward the pond, then abandoning the fill-vessel intent partway there (wandered off to talk instead). Needs investigation into why ActionFillVessel intent was dropped before reaching water. Observed during extract testing, world-test-extract - only investigate if user reports another instance.

2. **Build fence ignores unlocked lines when locked lines exist** — Two characters with build fence orders sit idle (looking at sticks, not gathering materials) when locked lines are exhausted but unlocked lines with available materials exist. Save evidence (world-test-step12a): line 3 has no material set, bricks and sticks on ground, characters know brick-fence and stick-fence recipes. Possibly the intent finder only evaluates locked lines first and never falls through to unlocked ones, or the transient nil guard prevents abandonment without triggering re-evaluation of unlocked lines.

3. **Material lock stamps too early** — `selectConstructionMaterial` stamps material onto an entire line when a character first evaluates it, before any tile is built. If that material runs out before any construction happens, the line is permanently locked to an unavailable material. Save evidence (world-test-step12a): line 2 locked to grass with no tiles built. Consider: should stamping happen at build time instead of selection time? Or should locked-but-exhausted lines be unlockable?

## Unallocated Features

# Ideas that aren't ready yet

- hello world loading screen
- Extended ascii art mushrooms on title screen
- Reorder the fields in the Details panel (low - still needs product decision)
- Auto pause at certain events?
- Fire for cooking (consider after Gardening - depends on water carrying)
  - Pumpkin Pie recipe to use those hollow gourds that contain 1 gourd
  - Porridge! made out of tall grass seeds or nuts or berries + water
  - Soup! made out of mushrooms + water
  - feed fire with fuel?
  - food quality for mood
  - tea
- Streams and Bridges (requires Construction)
