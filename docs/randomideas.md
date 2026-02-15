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
2. **esc key clean up**: still a little bit inconsistant how esc behaves. New pattern: esc always takes you 'back' a level. Ideal if there's a way to generalize this behavior. desired behavior:
   - esc from any expanded view collapses the view
   - esc from all activity view goes nowhere
   - esc from orders goes back to all activity
   - esc from within orders only goes back one level
   - esc from select: details view/action log or no panel goes back to all activity view
   - esc from select: details view/any other panel goes back to action log
   - 'l' from select: details view/any other panel goes back to action log
   - q from anywhere goes to start file menu
   - q from start file menu quits
3. Why have we started running into all these tab/whitespace problems while editing just in the last week or so? especially since claude is the only one reading/writing code?

## UI Improvements (after Gardening)

1. **Order Selection UX**
   - Pain point: Scrolling through list every time
   - Single keypress selection (numbered list)
2. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
3. **Natural Language Descriptions**
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

3. why is create item from variety logic in character.go? are there other functions that ended up in unintuitive places?

## Unallocated Features

1. **Helping** - The closest character who isn't already addressing a need (ie they are doing an idle activity or an order)
   can be interrupted to help a character who is in crisis.
   - the helping character will give them the item they are carrying, if it will address a severe or crisis need of the needy char
   - otherwise they will target an item that will address the worst-off need of the needy char, and creates the shortest path from helper to item to needy char
   - helping character will drop one unhelpful item if their inventory is too full to pick up the helpful item
   - helper will resume their order if they had one assigned
2. **Activity Preferences**
   - A Character who does something in an order category (eg: 'Garden', 'Harvest', 'Craft') has a chance to form a preference for orders in thaat category.
   - They will get a mood imapct from completing orders in that category.
   - Future: preference would affect what orders they accept

## Flower Seeds / Flower Cultivation

Deferred from Gardening Phase (was in Feature Set 2). Flower cultivation is cultural/aesthetic, not survival-critical — the core gardening loop works with gourd seeds + plantable berries/mushrooms.

**The mechanic tension:** Flower seed gathering doesn't fit cleanly into any existing activity archetype:

- **Foraging** is becoming food-centric ("go find snacks"). Flower seeds aren't food.
- **Collecting** (future, for Construction) is about gathering raw materials from the ground. But flower seeds come _from_ an interaction with a living plant, not the ground.
- **Flower seed gathering** is "interact with a living plant to extract something, without destroying it" — a different verb from foraging, collecting, or harvesting (which kills the plant).

This is genuinely a fourth pattern without enough examples yet to know the right abstraction. Revisit after Construction introduces more "interact with world objects" patterns, which may reveal a natural home for this mechanic.

**Original requirement (Gardening-Reqs line 67):** "Foraging a flower produces 1 flower variety seed without removing the flower."

**Open questions:**

- Does this become its own know-how? Part of a broader "tend" or "gather" activity?
- Can the same flower be gathered from repeatedly? Timer per flower? Match flower reproduction cadence?
- Standard pickup logic for the seed, or special handling?

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
