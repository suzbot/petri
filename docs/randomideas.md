This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## Issues To resolve

1. **Game Mechanics**: they are little disorganized, inconsistent about what level of detail it share, not in the most intuitive order.
   a. Reorganize in order of gameflow
   b. remove anything that is unnessecary for user, or summarize where makes sense
   c. Put table of contents at top (can it link to header sections?)
2. **Orders**:
   a. if an order is not completable, ie abandoned, but its at the top of the list, characters assign and then abandon it over and over
   b. ideally, if they can't do the first order, they move on to the next open order
   c. ideally, if an order is unfufillable (either because no one has the knowledge anymore, or its missing a required component in the world)
   d. then it should be shown as unfufullable on the orders screen, and characters should skip it without assigning themselves
   e. Open to other approaches to solve the experience

## UI Improvements

1. **Preferences Panel Restructure**
   - Move preferences list to lower panel (viewable with keypress 'p')
   - Make lower panel scrollable
2. **Order Selection UX**
   - Pain point: Scrolling through list every time, reopening menu for multiple orders.
   - Single keypress selection (underlined unique letter or numbered list)
   - Stay in add mode after adding an order (don't close menu)
3. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
4. **Natural Language Descriptions**
   - Replace details view in regular mode, supplement in debug mode.
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."

## Streamlined Character Creation (combined scope)

Remove single/multi mode distinction from UI, add character count control, and improve name management:

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

3. **Character Names File**
   - Put names in their own file for easy user/dev additions
   - Alphabetize all names to prevent duplicates, breaking lines by starting letter
   - Names to add (if they aren't already on the list):
     Oak, Ash, Elm, Birch,
     Bine, Bit, Bole, Bog, Bough, Brome, Burr, Buzz, Briar, Branch, Brush
     Cress, Croft, Creft, Chert, Claw,
     Daub, Dew, Dock, Dray, Duff, Eft, Egg,
     Fen, Fir, Frond, Furl, Flax, Flit, Fitch,
     Gnarl, Grub, Grue, Hedge, Herb, Knoll
     Larch, Lil, Log,
     Marsh, Marl, Mite, Mole, Mor, Moth, Muld, Mull, Midge, Mouse
     Newt, Nook, Olm, Nest, Ov,
     Pad, Peat, Pitch, Pith, Pod, Phlom, Phlox,
     Reed, Rime, Root, Rye, Quill, Quitch,
     Sap, Scamps, Scrub, Scurf, Sedge, Shoot, Shrew, Skink, Sod, Sprig, Sprout,
     Stick, Stub, Stump, Snug, Skunk, Spud, Stoat, Swale
     Tarn, Toady, Twig, Tuff, Tuft,
     Verd, Vetch, Vole, Weir, Web, Yew, Wort, Weed

## Tech Updates

1. Create Code Review Agent
2. Assess process docs for claude skills
3. Consider sequential problem solving MCS
4. review pre-gardening-audit for audit areas that were out of scope, and add any triggers for those deferred items to triggered-enhancements.md
5. clean up triggered-enhacements.md to remove complete items or items no longer applicable (after gardening)
6. Evaluate how to make architecture doc more useful

## Small Features

1. Ability to edit name of existing character on map
2. Edible Nuts drop from canopy

# Ideas that aren't ready yet:

- hello world loading screen
- Extended ascii art mushrooms on title screen
- Reorder the fields in the Details panel (low - still needs product decision)
- Auto pause at certain events?
- Fire for cooking
  - Pumpkin Pie recipe to use those hollow gourds that contain 1 gourd
  - Porridge! made out of tall grass seeds or nuts or berries + water
  - Soup! made out of mushrooms + water
  - feed fire with fuel?
  - food quality for mood
  - satiety of food items
- Streams and Bridges (requires construction)
- Use moss, feathers, and leaves to make beds
- Baskets, Bags, and Bins
