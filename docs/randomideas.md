This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## Issues To resolve

1. **Gardening Phase testing feedback**
   a. **"Foraging for vessel" text is unintuitive.** When foraging rolls and scores an empty vessel (as a container to carry future food), the activity reads "Foraging for vessel." Characters aren't foraging *for* vessels — they're picking one up to aid foraging. Root cause: `scoreForageVessels` gives empty vessels a positive score (`vesselBonus - distance`) even when no growing items exist to put in them. Consider: (1) only score empty vessels when matching growing items exist (like partial vessels already do), and/or (2) change activity text to "Picking up vessel" when the target is a container.

   c. **Edibility Bug:** Gourds grown from seed to not inherit edible property so characters do not eat them. It is expected that an item grown from seed would inherit all the properties of its parent item, in this case gourds grown from seed should be edible, and targetable by hungry characters.
      - looking at the code, there's a spot where it pulls down the G symbol upon maturing from sprout. Could it also pull down the edible (and in the future any other attributes) from the registry at that point?

   d. **Discoverability Friction:** 
      - **Original Thinking:** The current pattern I've been trying to establish is that actions/recipes are discoverable by picking up or looking at the components that would be used in that action/recipe. My thought here was that this would be programatic enough to make it easy to add more discoverable actions/recipes over time. 
      - **Problem:** this actually feels off during play in the balance between tilling and planting
         - Tilling is only discoverable by looking at a hoe (no trigger to pick up a hoe before tilling). Very few hoes can exist at first, as shells are relatively rare. Therefore, Tilling can take a while to discover.
         - Planting is discoverable by looking at any plantable item, of which there is an abundance. It is disocovered quickly.
         - It feels strange that everyone knows how to plant, but no one can because know one knows how do to the prerequisite activity.
      - **Possible Solutions:**
         - Make Tilling pre-requisite knowledge to Planting
            Pro: helps enforce an intuitive 'tecnology tree'
            Con: Loses some emergence of different skill/knowledge sets for specialization
            Con: Achieves balance by making planting harder to discover instead of tilling easier to discover (maybe this is ok?)
         - Give Tilling more discoverability triggers (looking at seed)
            Pro: easier to discover
            Con: Gets in the way of my hope for consistent discoverability pattern / implementation
         - Bundle tilling discovery together with Craft Hoe Discovery -- when you look at a hoe component you have a chance to discover both tilling and hoe crafting together
            Pro: A little more intuitive -- why would you think to invent a hoe if you didn't also have tilling in mind?
            Pro: Could still have emergence of specialization through future non-bundled know-how transfer (teaching, observation)
            Con: Still kind of unintuitive -- why would you think of planting if you don't know about digging? why would you discover how to build a hoe if you weren't thinking about eventually planting?
            Con: Bulky messaging (could figure out how to streamline this -- maybe just one discovery message of "Had an idea!")


   e. **Pathing Thrashing:** we're seeing characters thrash more often when pathing now that we've 'upgraded' our pathing logic. The most direct route to an item is blocked by a character. Characrers aren't in the patching calculation because they can move, but often they are engaged in an activity that keeps them stationary for up to 20 seconds. I see characters take one step to try and path around them, but it appears that the new path is the same as the old path, they take one step back and they are blocked again. They spend a lot of energy moving back and forth until the other character moves, and its unpleasant to observe, and doesn't feel like natural behavior. What are some different approaches to improving this?
      i. idea one - force the character to take a few random steps away before re-calculating path, in hopes that new path doesn't have the exact same blockage as the old path
      ii. idea 2 - disallow new paths that have any of the same first three steps as the last path
      iii. anything else?

3. **esc key clean up** (after Gardening): still a little bit inconsistant how esc behaves. Proposed new pattern: esc always takes you 'back' a level. Ideal if there's a way to generalize this behavior. desired behavior:
   - esc from any expanded view collapses the view
   - esc from all activity view goes nowhere
   - esc from orders start goes back to all activity
   - esc from within orders only goes back one level
   - esc from select: details view > action log or no panel goes back to all activity view
   - esc from select: details view > any other panel goes back to details action log
   - 'l' from select: details view > any other panel goes back to details action log
   - q from anywhere goes to start file menu
   - q from start file menu quits

4. **Game Mechanics Doc Reorg** (after Gardening): they are little disorganized, inconsistent about what level of detail it shares, not in the most intuitive order.
   a. Reorganize in order of gameflow
   b. Remove anything unnecessary for user, or summarize where makes sense


## UI Improvements 

1. **Order Selection UX** (after Gardening)
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

## Tech Stuff

3. why is create item from variety logic in character.go? are there other functions that ended up in unintuitive places?
4. Why have we started running into all these tab/whitespace problems while editing just in the last couple weeks or so? especially since claude is the only one reading/writing code?

## Unallocated Features

1. **Helping** - The closest character who isn't already addressing a need (ie they are doing an idle activity or an order)
   can be interrupted to help a character who is in crisis.
   - the helping character will give them the item they are carrying, if it will address a severe or crisis need of the needy char
   - otherwise they will target an item that will address the worst-off need of the needy char, and creates the shortest path from helper to item to needy char
   - helping character will drop one unhelpful item if their inventory is too full to pick up the helpful item
   - helper will resume their order if they had one assigned
2. **Activity Preferences**
   - A Character who does something in an order category (eg: 'Garden', 'Harvest', 'Craft') has a chance to form a preference for orders in that category, based on their mood at completion.
   - They will get a mood impact from completing orders in that category.
   - Future: preference could affect what orders they accept

## Flower Seeds / Flower Cultivation (After Construction)

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
