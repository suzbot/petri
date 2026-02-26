This is a document where I'm dumping my random ideas as I have them,
that haven't been added to the roadmap or other docs yet.

Goal of this document is that it can be automatically reviewed when roadmap planning
to see if any of these smaller items make sense to incorporate into the plan as we go along.

When items are completed OR documented as part of a roadmap, requirement, or plan elsewhere,
then they can be removed from this list.

# Ready Ideas

## Issues To resolve

1. **Action log events appear in character order instead of tick order** — When stepping through ticks manually (pressing '.'), events from different ticks display in character ID order rather than chronological order. This happens because `GameTime` advances by wall-clock delta between ticks, and rapid keypresses produce deltas so small that events from separate ticks share effectively the same timestamp. The action log then sorts them by character ID.

   **Existing diagnostic:** `update.go:521-525` logs a warning to `~/.petri/debug.log` when `delta <= 0.001s`, but the threshold is too tight to catch most occurrences — the delta from manual stepping is typically above 1ms but still too small for meaningful ordering.

   **Root cause:** The action log relies on `GameTime` (a float64) for ordering, but `GameTime` advances by wall-clock delta which doesn't guarantee meaningful separation between ticks. This is a systemic issue, not specific to any feature. Primarily observed when stepping with '.' — the manual step path may not be incrementing game time correctly, or increments are too small to differentiate.

   **Likely fix direction:** Monotonic sequence numbers on log entries (e.g., a tick counter) instead of or in addition to `GameTime` for ordering purposes. Also investigate whether the '.' step code path handles time advancement differently from the normal tick path.

2. **Characters target unreachable/invisible items** — Observed during playtesting: a character would target an item, move toward it, but the item wasn't visible on the map. The character would spam movement attempts until a higher-priority need interrupted.

   **Root cause (suspected):** No reachability or validity check at targeting time. Characters select items by iterating all world items and scoring by distance + preference. Two likely scenarios:
   - **Stale reference after pickup:** Character A targets item X. Character B picks it up first. A's intent still holds a reference to item X, now in B's inventory and no longer on the map. A chases an invisible item.
   - **Stacked items:** `ItemAt()` returns the first item at a position for rendering, but scoring sees all items. A character could target a second item at the same tile — invisible on screen but present in data.

   **Steps to resolve:**
   - Add diagnostic logging: when pathfinding fails for a targeted item, log the item description, its position, whether it's still on the map (`gameMap.ItemAt()`), and whether any character is carrying it. This makes the invisible-item scenario observable during playtesting.
   - Audit intent invalidation: trace what happens to Character A's intent when Character B picks up the targeted item. Confirm whether there's a stale-reference gap.
   - Audit item stacking: check whether multiple items can end up at the same position, and whether only the top one renders.
   - Fix: either invalidate intents when items are picked up by others, or add a validity check (item still on map, reachable) at targeting time.


## UI Improvements

1. **Capitalization** consistency clean-up (not quick - lots of strings to sweep through)
3. **Natural Language Descriptions**
   - Key words shown in relevant color (color text is color, poison red, healing green, growing olive) and bolded
   - Everything non-null/non-false reflected in a sentence
   - Design must easily expand as new attribute types added
   - Example: "This is a hollow gourd. It is a vessel that can be used to carry things. It is warty and green."


## Tech Stuff



## Unallocated Features


2. **Helping** — Characters help community members in crisis by finding and delivering items to address their needs. The community is helpful at baseline — this isn't a rare event, it's the default social response to crisis.

   **Handoff:** Helper drops item adjacent to the needy character. The needy character's existing food/water scoring picks it up naturally — no direct "give" action needed.

   **Trigger pattern (needs refinement — compare these during design):**
   - **A. Order-like:** When a character hits crisis, a "help" order is implicitly created. First eligible character takes it.
   - **B. Need-like:** Closest character treats the needy character's crisis as their own overriding need, interrupting whatever they're doing (idle or working).
   - **C. Idle override:** Characters doing an idle roll are forced to choose helping as long as someone is in crisis. Multiple idle helpers may respond simultaneously.

   Not limited to idle characters — working characters should be able to help too. Size based on largest option (B, which interrupts any activity).

   **Item selection:** Helper chooses an item the way they'd choose food for themselves if they had the needy character's hunger level. Formula: `(Helper's Pref × Needer's severity weight) - (Helper's Distance × Needer's severity weight) - (Needer's Fit Delta)`. This naturally biases toward carried items (distance=0), incorporates the helper's poison/healing knowledge if they have it, and scales urgency against distance. Helper does not need to share knowledge — they just use their own knowledge when scoring.

   **Multi-step flow:** Helper may need to go find an item (not just give what they're carrying). Existing patterns for multi-step idle activities and ordered actions handle this. Helper drops unhelpful items if inventory is full, picks up the selected item, navigates to needy character, drops it adjacent.

   **Resumption:** Helper resumes their previous order if they had one assigned.

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
