This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## Issues To resolve


## UI Improvements

1. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
3. **Natural Language Descriptions**
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."


## Tech Stuff



## Unallocated Features


3. **Activity Preferences** — Characters form opinions about order categories (Garden, Harvest, Craft) through experience, similar to how food preferences form from eating.

   **Formation:** Chance to form a preference for the order's activity category on order completion, based on mood (same trigger conditions as item preference formation — Joyful/Happy → like, Unhappy/Miserable → dislike). Same formation chance as item preferences initially; can be tuned separately later if needed.

   **Player-visible impact (v1):** Mood change rate during the activity while performing it. A character who likes gardening feels happier while gardening; one who dislikes it feels worse. Visible in the preferences panel as a separate category (like how knowledge panel has Facts and Know-how sections).

   **Preference structure:** Separate category in the preferences panel, not the same Preference struct as item preferences (different target type — category string vs item attributes). Displayed alongside item preferences but visually distinct.

   **Future extensions (not v1):**
   - Preference affects which orders characters voluntarily accept
   - Personality emergence — characters develop work identities over time

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
