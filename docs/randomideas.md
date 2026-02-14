This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## Issues To resolve

1. **Game Mechanics Doc Reorg**: they are little disorganized, inconsistent about what level of detail it shares, not in the most intuitive order.
   a. Reorganize in order of gameflow
   b. Remove anything unnecessary for user, or summarize where makes sense

## UI Improvements (after Gardening)

1. **Order Selection UX**
   - Pain point: Scrolling through list every time
   - Single keypress selection (underlined unique letter or numbered list)
   - Add confirmation feedback when order is created (flash message or highlight new order)
2. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
3. **Natural Language Descriptions**
   - Replace details view in regular mode, supplement in debug mode.
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."

## Streamlined Character Creation (after Gardening)

Remove single/multi mode distinction from UI and add character count control:

1. **UI Streamlining**
   - Remove single char mode from game start
   - Adjust start screen title and keys:
     - === Petri ===
     - R to start with Random Characters
     - C to create Characters
   - "C for Create" creates first character, randomizes the rest

2. **Character Count Flag**
   - Add `-characters=N` flag to control spawn count (default 4)
   - `-characters=0` equivalent to `-no-characters`
   - Reasonable max cap (e.g., 20)
   - Note: Not a quick standalone win - requires refactoring the current character creation flow, which is why it's scoped together with UI Streamlining above

## Tech Updates

3. Consider sequential problem solving MCS
4. Evaluate how to make architecture doc more useful
5. move name file to config dir for clearer access to users

## Unallocated Features

1. **Helping** - The closest character who isn't already addressing a need (ie they are doing an idle activity or an order) 
can be interrupted to help a character who is in crisis.
   - the helping character will give them the item they are carrying, if it will address a severe or crisis need of the needy char
   - otherwise they will target an item that will address the worst-off need of the needy char, and creates the shortest path from helper to item to needy char
   - helping character will drop one unhelpful item if their inventory is too full to pick up the helpful item
   - helper will resume their order if they had one assigned
2. **Activity Preferences**
   - A Character who does something in an order category (eg: 'Garden', 'Harvest', 'Craft') has a chance to form a preference for orders in thaat category. 
   - They will 
   - They will get a mood increase from completing orders in that category.
   - Future: preference would affect what orders they accept


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
