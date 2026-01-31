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

## UI Improvements

1. **Preferences Panel Restructure**
   - Move preferences list to lower panel (viewable with keypress 'p')
   - Make lower panel scrollable
2. **Order Selection UX**
   - Pain point: Scrolling through list every time, reopening menu for multiple orders.
   - Single keypress selection (underlined unique letter or numbered list)
   - Stay in add mode after adding an order (don't close menu)
3. **Capitalization** consistency clean-up
4. **Natural Language Descriptions**
   - Replace details view in regular mode, supplement in debug mode.
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."
5. Show passage of world time below map

## Tech Updates

1. **Character Names**: put in their own special file so its super easy for users/me to add more names.
   - Names to remember to add next: Bog, Log, Fen, Pod, Bough, Toady, Tarn, Weir, Bine, Brome, Cress, Daub, Nook, Fir, Frond, Furl, Rye, Muld, Grue, Gnarl, Peat, Pad, Sod, Sprout
   - Alphabetize all names, to make it easy for humans to not add duplicates
2. Create Code Review Agent
3. Assess process docs for claude skills
4. Consider sequential problem solving MCS

## Small Features

1. Ability to edit name of existing character on map
2. Remove single char mode from UI, streamline flow:
   - Adjust start screen title and start keys
   - === Petri ===
   - R to start with Random Characters
   - C to create Characters
3. Edible Nuts drop from canopy
4. Ability to slow down game speed

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
