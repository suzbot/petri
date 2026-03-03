This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## UI Improvements

1. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
3. **Natural Language Descriptions**
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."


## Tech Stuff

## Display / Naming

1. **Tall grass variety display name** — When a tall grass seed sprout matures, the resulting plant's display name may need review. Discuss whether "pale green tall grass" is the right phrasing vs. just "tall grass" (since all tall grass is pale green — the color is redundant).

## Bugs to Investigate

1. **Fill vessel abandoned mid-trip** — Observed character picking up a vessel for water, heading toward the pond, then abandoning the fill-vessel intent partway there (wandered off to talk instead). Needs investigation into why ActionFillVessel intent was dropped before reaching water. (Observed during extract testing, world-test-extract)
2. **"Moving to look at item" stuck when standing on it** — Character gets stuck in "moving to look at item" state when already on the item's tile. Likely a pathfinding/adjacency check that doesn't handle distance-zero. (Observed during 2b-3 testing)

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
  - satiety of food items
- Streams and Bridges (requires Construction)
- Use moss, feathers, and leaves to make beds (requires Construction)
- Baskets, Bags, and Bins (consider after Gardening)
- When 9 mushrooms of the same variety grow together in a 3x3 patch they become a mega mushroom that can be converted into a mushroom hut
