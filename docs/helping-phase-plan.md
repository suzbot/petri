# Helping Feature Plan

Requirements: [randomideas.md](randomideas.md) lines 33-48

---

## Overview

Characters help community members in crisis by finding and delivering items. Helping is baseline social behavior — when an idle character notices someone in Crisis hunger or thirst, they find food or water and bring it to them. The community is helpful by default; this isn't rare altruism, it's the normal social response to crisis.

---

## Scope

**v1 (this plan):**
- Idle activity override: before the idle roll, check for Crisis characters
- Hunger crisis: helper finds food, delivers it to needy character
- Thirst crisis: helper procures/fills water vessel, delivers it
- Helper uses their own knowledge/preferences when selecting items (poison avoidance, healing awareness apply naturally)
- Helper drops item cardinal-adjacent to needy character
- Prerequisite: ground vessel eating (independently valuable improvement)

**Deferred (tracked in randomideas.md):**
- Option B trigger: working characters interrupted by crisis (order pause/resume to help)
- Health/energy crisis response
- Helper drops non-food to make room when inventory full

---

## Resolved Design Decisions

1. **Trigger pattern: Idle override (C).** Before the idle roll, scan for crisis characters. Most architecturally aligned with existing idle activity patterns. Working-character interruption (B) deferred to randomideas.md as a future enhancement.

2. **Crisis stats: Hunger and Thirst only.** These are the stats where delivering an item helps directly. Health crisis (healing items) is a future extension.

3. **No helper limit.** Each idle roll independently checks for crisis. Multiple helpers may respond — more food near a starving character is a good thing. Attempting to help isn't wasted effort.

4. **No explicit abandonment.** Helper commits to delivery. Drops item even if needy character satisfies their need before arrival. The only bail-out is "can't find a helpful item." Next idle roll, crisis is gone, no more helping triggers.

5. **Item selection: ScoreFoodFit with Severe-tier weights and needer's hunger for satiation fit.** Helper uses existing food scoring formula with Severe-tier pref/dist weights (not the needer's actual Crisis tier). The needer is always at Crisis hunger (that's the trigger), and Crisis weights zero out preference — a helper who knows a mushroom is poisonous would still grab it if nearest. Severe weights preserve the helper's judgment: poison knowledge (auto-formed dislike) penalizes known-poisonous items, distance still matters significantly, and satiation fit uses the needer's actual hunger (100) so large meals are preferred.

6. **Carried items are candidates.** Distance 0 in the formula means carried food scores very well. A helper already carrying food walks straight to the needy character — no procurement phase needed.

7. **Carried food vessels are delivery candidates.** If the helper is carrying a vessel of berries, they deliver the whole vessel. With ground vessel eating (Step 1), the needy character eats from the dropped vessel naturally.

8. **Drop position: cardinal adjacency.** Consistent with water tile interaction and other "deliver to" patterns. Avoids diagonal distance weirdness in the needy character's food scoring.

9. **Two separate activities: helpFeed and helpWater.** Different delivery mechanics (loose food/food vessel vs water vessel). Keeps self-managing flows clean and lets each action borrow from its closest existing analog (foraging vs fill-vessel).

10. **Full-inventory helpers with no carried food can't help.** They bail and do the normal idle roll. "Drop non-food to make room" is a future enhancement.

11. **Target character tracking: TargetCharacter on Intent.** Reuses the existing `TargetCharacter *Character` field on Intent (already used by ActionTalk). Intent is not serialized — recalculated on load. No new field needed. Helper uses TargetCharacter for: food scoring (needer's hunger level), navigation (needer's current position via `.Pos()`), and drop location (cardinal-adjacent to needer).

12. **Ground vessel eating: eat in place.** Walk to vessel, eat from it without picking up. Follows the ground water vessel drinking precedent (drink in place, clear intent after each unit, re-evaluate). No inventory management needed — works with full inventory. Communal: multiple characters can eat from the same ground vessel.

13. **Ground food vessels are procurement candidates.** A ground vessel of berries near the helper is structurally identical to loose ground food: walk to vessel, pick up, walk to needer, drop. The "two-leg routing" concern that originally deferred this isn't real — all ground food procurement is two-leg. Scoring follows the existing ground-vessel scoring pattern in `FindFoodTarget` (movement.go). Four candidate pools: carried loose food (dist 0), carried food vessel (dist 0), ground loose food, ground food vessel contents (scored by contents' variety, vessel is the TargetItem).

14. **Crisis target selection: nearest wins, stat tie-break for ties.** `findNearestCrisisCharacter` returns the nearest character with any Crisis stat (hunger or thirst). Distance is the primary criterion — a starving character 5 tiles away is helped before a dehydrated character 20 tiles away. The existing stat tie-breaker (thirst > hunger) applies only for equidistant characters or when the same character has multiple Crisis stats. This matches the intent system's pattern: stat priority resolves ties, distance resolves most decisions.

15. **helpWater fallback to helpFeed.** If the nearest crisis character has Crisis thirst but no vessel is available, and they also have Crisis hunger, try `findHelpFeedIntent` as fallback. Some help is better than none.

---

## Implementation Steps

### ✅ Step 1: Ground Vessel Eating

**Anchor story:** A character is hungry. There's a vessel of mushrooms on the ground 5 tiles away and a loose berry 12 tiles away. They walk to the vessel and eat from it — closer, better-fitting food.

**Reqs reconciliation:** Not in the original helping requirements — this is an independently valuable improvement that also supports helping (Decision 12: helpers may drop food vessels, needy characters eat from them on the ground). Currently `FindFoodTarget` scores ground items, carried items, and carried vessel contents, but not ground vessel contents. A character ignoring a vessel of food sitting next to them isn't intentional design. Fixing it makes food sources consistent across all seeking contexts.

**Architecture alignment:**
- Extends food scoring in `movement.go` (`FindFoodTarget`). Adds ground vessel contents as a fourth candidate pool alongside ground items, carried items, carried vessel contents. Follows the existing carried-vessel scoring pattern (lines ~1074-1129): score by contents' variety using `ScoreFoodFit` and `NetPreferenceForVariety`, return the vessel as TargetItem.
- Eat-in-place consumption follows the **ground water vessel drinking pattern**: character walks to vessel, consumes in place, clears intent after each unit, re-evaluates. Direct parallel to `ActionDrink` handler's ground vessel path and the arrival detection in `continueIntent` (lines ~497-514).
- `continueIntent` generic path handles movement (vessel stays on map throughout). Arrival detection block switches action to `ActionConsume` when character reaches the ground vessel — same shape as the thirst/ground-vessel arrival detection.
- `ConsumeFromVessel` (`consumption.go`) already operates on a vessel `*Item` pointer without assuming inventory ownership — it modifies the stack directly. Works as-is for ground vessels.
- **Not** a self-managing action — single phase (walk to vessel, eat). No procurement, no early-return block needed.

**Values alignment:**
- **Follow the Existing Shape** — drinking from ground water vessels is the direct precedent for both scoring (distance-based source selection) and consumption (eat in place, clear intent, re-evaluate).
- **Reuse Before Invention** — `ConsumeFromVessel` works unchanged. `ScoreFoodFit` and `NetPreferenceForVariety` reused for scoring.

**Anti-pattern:** Do NOT route ground vessel eating through pickup. Eat-in-place is a distinct path, not "pick up then eat." The vessel stays on the ground throughout (Decision 12).

**Serialization:** No new entity fields. Ground vessel eating is a new code path using existing data structures.

**Implementation:**

1. **Tests first:**
   - Character at Moderate hunger, ground vessel with mushrooms at distance 5, loose berry at distance 12 → `FindFoodTarget` chooses vessel (better fit and closer). [Regression: scoring already exists]
   - Character at Crisis hunger, loose berry at distance 2, ground vessel with mushrooms at distance 8 → chooses berry (distance wins at Crisis). [Regression: scoring already exists]
   - Character with full inventory eats from ground vessel without picking it up.
   - Ground vessel with non-edible contents (water, sticks) is not scored as food. [Regression: scoring already exists]
   - Empty ground vessel is not scored as food. [Regression: scoring already exists]
   - Intent clears after eating one unit from ground vessel (re-evaluation, same as vessel drinking).

2. ~~**FindFoodTarget scoring**~~ **Already implemented.** `FindFoodTarget` (movement.go lines ~1103-1135) already scores ground vessel contents as a fourth candidate pool alongside ground items, carried items, and carried vessel contents. Added during post-gardening cleanup. No changes needed.

3. **Arrival detection in `continueIntent`** (`system/movement.go`): When character arrives at a ground vessel targeted for eating (`DrivingStat == StatHunger`, TargetItem is on the map with edible container contents), switch action to `ActionConsume`. Directly parallels the ground water vessel arrival detection block (lines ~497-514 which checks `DrivingStat == StatThirst`).

4. **Extend ActionConsume handler** (`ui/update.go`): Currently only checks inventory for TargetItem. Add a ground vessel path: if TargetItem is not in inventory, check if it's on the map at the character's position with edible container contents. If found, call `ConsumeFromVessel` (works as-is on ground vessels). **Clear intent after each unit** (like vessel drinking) so the character re-evaluates — may eat again from the same vessel or seek another source.

✅ [TEST] Passed. Hungry character walks to ground food vessel and eats from it without picking it up. Multiple characters eat from the same vessel. Crisis hunger prefers closer loose food. Drinking behavior unchanged.

[RETRO]

---

### ✅ Step 2: helpFeed (Hunger Crisis Response)

**Anchor story:** A character hits Crisis hunger — they're starving and slow. Another idle character nearby notices, grabs a mushroom from the ground, walks over, and drops it next to them. The hungry character's food scoring finds the mushroom and they eat it. If the helper was already carrying a berry, they skip procurement and walk straight to the starving character.

**Reqs reconciliation:** randomideas.md line 33 — "Characters help community members in crisis by finding and delivering items to address their needs." Line 36 — "Helper drops item adjacent to the needy character. The needy character's existing food/water scoring picks it up naturally." Line 44 — item selection formula using helper's preferences weighted by needer's severity. Implementation uses ScoreFoodFit with **Severe-tier weights** (Decision 5) — needer's actual Crisis tier would zero out preference and disable poison avoidance, contradicting the requirement's intent that helper's knowledge applies.

**Architecture alignment:**
- **Adding an Idle Activity** (architecture.md): self-managing, like ActionForage. New `ActionHelpFeed` constant. Multi-phase flow with stateless phase detection.
- **Self-managing action pattern** (architecture.md: Self-Managing Actions): phases detected by world state — "do I have food in inventory?" determines whether to procure or deliver. Action type stays `ActionHelpFeed` throughout.
- **`continueIntent` early-return block**: needed because TargetItem (food) moves from ground to inventory during procurement. Without it, the generic path's `ItemAt` check would nil the intent after pickup. Same pattern as ActionFillVessel.
- **`ScoreFoodFit`** (from 3C): reused with Severe-tier pref/dist weights, needer's hunger for satiation fit, helper's preferences/knowledge. Four candidate pools (Decision 13): helper's carried loose food (distance 0), carried food vessel (distance 0, scored by contents), ground loose food, ground food vessel contents (scored by contents' variety, vessel is TargetItem).
- **`TargetCharacter` on Intent** (Decision 11): reuses the existing `TargetCharacter *Character` field already used by ActionTalk. Intent is not serialized. No new field needed.
- **No ActivityRegistry entry needed.** Helping isn't discovered, ordered, or part of the idle roll. It's a pre-roll override. Action log messages use the action type directly.
- **`isIdleActivity` → `isIdleAction` refactor**: replaces fragile string-prefix checking with ActionType checking. Helpers (and any future non-idle actions) are automatically excluded from talking availability. Follows **Start With the Simpler Rule** (Values.md) — checking a type enum is simpler and more robust than maintaining a string prefix list.

**Values alignment:**
- **Follow the Existing Shape** — food scoring reuses ScoreFoodFit, ground vessel scoring follows FindFoodTarget's pattern (movement.go lines 1103-1134), pickup reuses existing helpers, self-managing phases follow ActionForage's pattern. TargetCharacter reuses the ActionTalk targeting pattern.
- **Reuse Before Invention** — no new scoring formula, no new pickup mechanics, no new navigation pattern, no new Intent field.
- **Start With the Simpler Rule** — `isIdleAction` replaces string prefix matching with an ActionType check. Simpler, robust, no maintenance burden for new actions.
- **Anchor to Intent** — the anchor test validates "food ends up next to the starving character," not "returns ActionHelpFeed."

**Anti-pattern:** Do NOT use Crisis-tier scoring weights. The needer is always at Crisis hunger, and Crisis weights zero out preference — helpers would ignore poison knowledge. Use Severe-tier weights so the helper exercises judgment while still prioritizing proximity.

**Serialization:** No new entity fields. TargetCharacter already exists on Intent and is not serialized (Intent is recalculated on load).

**Implementation:**

1. **Tests first:**
   - Idle character, another character at Crisis hunger, food on ground → helper gets `ActionHelpFeed` intent targeting the food.
   - No character at Crisis hunger → normal idle roll (no helping override).
   - Helper already carrying food (distance 0) → intent targets needy character directly, no procurement.
   - Helper already carrying a food vessel → intent delivers the vessel.
   - Helper scores food using Severe-tier weights with needer's hunger for fit: mushroom (meal-tier, closer) scores better than berry (snack-tier, farther) because distance matters and satiation fit favors larger meals at hunger 100.
   - Helper knows an item is poisonous → that item scores lower (helper's dislike preference penalizes it at Severe weights).
   - Ground food vessel with edible contents → scored as candidate, vessel is TargetItem.
   - Helper's inventory full with no food → can't help, falls through to normal idle roll.
   - Multiple crisis characters → helper targets nearest.
   - Helper drops food cardinal-adjacent to needy character.
   - After dropping, helper's intent clears (returns to idle).
   - Helper actively delivering food is NOT targetable for talking (isIdleAction excludes ActionHelpFeed).

2. **`ActionHelpFeed` constant** (`entity/character.go`): New action type constant.

3. **Replace `isIdleActivity` with `isIdleAction`** (`system/idle.go`): The existing `isIdleActivity(activity string)` checks string prefixes to determine if a character is available for talking. This is fragile — any new activity with a colliding prefix (e.g., "Fetching water for [Name]" matches "Fetching") would silently make helpers interruptible. Replace with `isIdleAction(action ActionType)` that checks against the known idle action types: `ActionNone`, `ActionLook`, `ActionTalk`, `ActionForage`, `ActionFillVessel`. New actions like `ActionHelpFeed`/`ActionHelpWater` are automatically excluded. Update the caller in `findTalkIntent` (talking.go) and the continuation check in `CalculateIntent` (movement.go) to use ActionType instead of string matching. Test: existing talking behavior unchanged (idle characters still find each other for conversation).

4. **Crisis detection** (`system/helping.go`): `findNearestCrisisCharacter(helper *Character, characters []*Character) *Character` — scans all characters, returns the nearest character with Crisis hunger (nil if none). Skips the helper themselves and dead/sleeping characters. Uses Manhattan distance for "nearest." Step 3 extends to include Crisis thirst.

4. **Idle override** (`system/idle.go`, `selectIdleActivity`): After order checks, after idle cooldown (same cadence as other idle activities), before the random roll. Call `findNearestCrisisCharacter`. If a character with Crisis hunger is found, call `findHelpFeedIntent` instead of rolling. If the intent finder returns nil (can't find food), fall through to normal idle roll.

5. **`findHelpFeedIntent`** (new function, `system/helping.go`): Scores food candidates using `ScoreFoodFit` with **Severe-tier weights** (`config.FoodSeekPrefWeightSevere`, `config.FoodSeekDistWeightSevere`) and **needer's hunger** for satiation fit. Also applies helper's healing knowledge bonus (following `FindFoodTarget`'s pattern) if needer's health is low. Four candidate pools:
   - Helper's carried loose food items (distance 0)
   - Helper's carried food vessel (distance 0, scored by contents' variety)
   - Ground loose food items (distance from helper)
   - Ground food vessels with edible contents (distance from helper, scored by contents' variety, vessel is TargetItem)

   If best candidate is carried (loose food or food vessel) → create intent: `ActionHelpFeed`, `Dest` = needer's position, `TargetCharacter` = needer. TargetItem = the carried item (identifies what to drop).

   If best candidate is on ground (loose food or food vessel) → create intent: `ActionHelpFeed`, `TargetItem` = the item/vessel, `TargetCharacter` = needer. Procurement phase: walk to food, pick up.

   If no candidates → return nil (can't help).

6. **`ActionHelpFeed` handler** (`ui/update.go`, `applyIntent`): Self-managing phases detected by world state:

   **Phase: Procurement** (TargetItem on the ground, not in inventory)
   - Walking toward TargetItem (ground food or ground food vessel): handled by `continueIntent` movement.
   - On arrival at food/vessel: pick up (standard `Pickup`). Transition to delivery phase.

   **Phase: Delivery** (food in inventory, not cardinal-adjacent to needer)
   - Update Dest to needy character's current position (via `TargetCharacter.Pos()`).
   - Walk toward needy character: handled by `continueIntent` movement.

   **Phase: Drop** (food in inventory, cardinal-adjacent to needer)
   - Find empty cardinal-adjacent tile next to needy character.
   - Drop the food item (or food vessel) at that tile via `Drop`.
   - Log action: "[Helper] brought [food] to [Needer]"
   - Clear intent. Helper returns to idle.

   **Edge cases:**
   - Needy character moved and no cardinal-adjacent tile is empty → keep walking toward them (recalculate Dest).
   - Needy character died → drop food at current position and clear intent.
   - TargetItem was taken by someone else during procurement → intent nils via continueIntent, character re-evaluates next idle roll.

7. **`continueIntent` early-return block** (`system/movement.go`): Add `ActionHelpFeed` to the early-return section. Same structure as ActionFillVessel:
   - TargetItem on the ground (`gameMap.ItemAt` matches) → recalculate path toward TargetItem. (Procurement phase.)
   - TargetItem in inventory → recalculate path toward `TargetCharacter.Pos()`. (Delivery phase — needer may have moved.)
   - TargetItem gone (not on ground, not in inventory) → return nil. (Taken by someone else.)

8. **Action log messages**:
   - On intent creation: "[Helper] is bringing food to [Needer]"
   - On drop: "[Helper] brought [food description] to [Needer]"

✅ [TEST] Create a test world with a character near Crisis hunger and another character nearby with food available:
- Idle character notices crisis and walks to pick up food
- Helper delivers food by dropping it cardinal-adjacent to the starving character
- Starving character eats the delivered food
- Helper carrying food skips procurement, walks directly to needer
- Helper carrying a food vessel delivers the whole vessel
- Ground food vessel nearby → helper picks it up and delivers it
- With no food available, helper does normal idle activity instead
- Multiple idle characters may help simultaneously (multiple deliveries)
- Helper with full inventory and no food does normal idle activity

[RETRO]

---

### Step 3: helpWater (Thirst Crisis Response)

**Anchor story:** A character hits Crisis thirst. An idle character nearby finds an empty vessel on the ground, picks it up, fills it at the pond, walks to the thirsty character, and drops the water vessel next to them. The thirsty character's drink scoring finds the ground water vessel and drinks from it. If the helper was already carrying a water vessel, they skip procurement and fill, walking straight to the dehydrated character.

**Reqs reconciliation:** randomideas.md line 33 — "Characters help community members in crisis by finding and delivering items to address their needs." Thirst crisis addressed by delivering water. Needy character's existing water scoring (which already handles ground water vessels — see game-mechanics.md Drinking section) picks up the delivery naturally.

**Architecture alignment:**
- **Adding an Idle Activity** (architecture.md): self-managing, like ActionFillVessel. New `ActionHelpWater` constant. Multi-phase flow with stateless phase detection.
- **Self-managing action pattern** (architecture.md: Self-Managing Actions): four phases detected by world state — "do I have a vessel?", "is it filled?", "am I adjacent to needer?" Reuses `RunVesselProcurement` and `RunWaterFill` tick helpers unchanged.
- **`continueIntent` early-return block**: needed because vessel moves between ground and inventory, and has three distinct navigation targets across phases (TargetItem on ground, water Dest, TargetCharacter). Extends ActionFillVessel's two-path block with a third path using `vesselHasLiquid` (already exists in movement.go) to distinguish fill from delivery.
- **`TargetCharacter` on Intent** (Decision 11): reuses existing field. Same pattern as helpFeed.
- **Capture-and-restore pattern for TargetCharacter**: `RunWaterFill` rebuilds `char.Intent` after procurement (because `Pickup` nils it), but doesn't set TargetCharacter. Handler captures `needer` from the intent before procurement, then patches it back after `RunWaterFill` rebuilds. One line — RunWaterFill is unchanged. See "Anti-pattern" below.

**Values alignment:**
- **Reuse Before Invention** — `RunVesselProcurement`, `RunWaterFill`, `Drop`, `vesselHasLiquid` all reused unchanged. Capture-and-restore avoids modifying shared helpers.
- **Follow the Existing Shape** — ActionFillVessel (fetch_water.go, update.go) is the direct structural analog for procurement + fill. Delivery phase follows helpFeed's pattern.

**Anti-pattern:** Do NOT modify `RunWaterFill` to accept/preserve TargetCharacter. It's a shared helper used by ActionFillVessel and ActionWaterGarden. The capture-and-restore pattern keeps the coupling one-directional — the handler knows about the helper's behavior, the helper doesn't know about helping.

**Serialization:** No new entity fields. TargetCharacter already exists on Intent and is not serialized.

**Implementation:**

1. **Tests first:**
   - Idle character, another character at Crisis thirst, empty vessel available → helper gets `ActionHelpWater` intent.
   - Helper already carrying water vessel → skips procurement and fill, walks directly to needer.
   - Helper carrying empty vessel → skips procurement, goes to fill, then delivers.
   - No vessel available and can't procure one → can't help, falls through to normal idle roll.
   - Same character at both Crisis hunger and Crisis thirst → thirst takes priority (existing stat tie-breaker on same character).
   - Two crisis characters at different distances → helper targets nearest regardless of crisis type (distance primary, stat tie-break only for equidistant).
   - No vessel available but needer also has Crisis hunger → falls back to helpFeed.
   - Helper drops water vessel cardinal-adjacent to needy character.
   - After dropping, helper's intent clears.
   - TargetCharacter preserved across procurement → fill transition (capture-and-restore).

2. **`ActionHelpWater` constant** (`entity/character.go`): New action type constant.

3. **Extend crisis detection** (`system/helping.go`): `findNearestCrisisCharacter` (from Step 2) extended to check Crisis thirst alongside Crisis hunger. Returns nearest character with any Crisis stat. **Distance is primary** — stat tie-break (thirst > hunger) only for equidistant characters or same character with multiple crises. Skips dead/sleeping characters and the helper themselves.

4. **Extend idle override** (`system/idle.go`, `selectIdleActivity`): After `findNearestCrisisCharacter` returns a character, check their crisis stats:
   - Crisis thirst → `findHelpWaterIntent`
   - Crisis hunger → `findHelpFeedIntent` (from Step 2)
   - Same character at both → thirst first (existing stat tie-breaker)
   - **Fallback:** If `findHelpWaterIntent` returns nil (no vessel) and needer also has Crisis hunger, try `findHelpFeedIntent`. Some help is better than none.

5. **`findHelpWaterIntent`** (new function, `system/helping.go`): Determines water delivery approach based on helper's current state:
   - Helper carrying full water vessel → create intent: `ActionHelpWater`, `Dest` = needer's position, `TargetCharacter` = needer, `TargetItem` = vessel. Skip to delivery.
   - Helper carrying empty vessel → create intent: `ActionHelpWater`, `TargetItem` = vessel, `TargetCharacter` = needer. Fill phase: `RunWaterFill` finds water and navigates.
   - No vessel in inventory, has inventory space → find nearest empty ground vessel via `findEmptyGroundVessel` (fetch_water.go). Create intent: `ActionHelpWater`, `TargetItem` = ground vessel, `TargetCharacter` = needer. Procurement phase first.
   - No vessel available anywhere (no carried vessel, no ground vessel, no inventory space) → return nil (can't help).

6. **`ActionHelpWater` handler** (`ui/update.go`, `applyIntent`): Self-managing phases detected by world state. Follows ActionFillVessel's structure (update.go lines 1207-1241) for procurement and fill, then adds delivery.

   **Critical: capture needer at handler top.** `needer := char.Intent.TargetCharacter` — must be captured before `RunVesselProcurement` might call `Pickup` which nils `char.Intent`.

   **Phase: Vessel Procurement** (TargetItem on the ground, not in inventory)
   - `RunVesselProcurement` each tick. On `ProcureApproaching` → moveWithCollision. On `ProcureInProgress` → return. On `ProcureFailed` → return (intent already nilled).
   - On `ProcureReady` → fall through to fill phase (same tick, same as ActionFillVessel).

   **Phase: Fill** (vessel in inventory, empty — `len(vessel.Container.Contents) == 0`)
   - `RunWaterFill` each tick with `entity.ActionHelpWater` as actionType.
   - **After RunWaterFill returns** (any status): restore TargetCharacter if lost:
     `if char.Intent != nil && char.Intent.TargetCharacter == nil && needer != nil { char.Intent.TargetCharacter = needer }`
   - On `FillApproaching` → moveWithCollision. On `FillInProgress` → return. On `FillFailed` → return.
   - On `FillReady` → transition to delivery: build new intent with `Dest` = `needer.Pos()`, `TargetItem` = vessel, `TargetCharacter` = needer.

   **Phase: Delivery** (vessel has water — `vesselHasLiquid(vessel)`, not cardinal-adjacent to needer)
   - Update Dest to needy character's current position (via `TargetCharacter.Pos()`).
   - Walk toward needy character: moveWithCollision.

   **Phase: Drop** (vessel has water, cardinal-adjacent to needer)
   - Find empty cardinal-adjacent tile next to needy character.
   - Drop the water vessel at that tile via `DropItem`.
   - Log action: "[Helper] brought water to [Needer]"
   - Clear intent. Helper returns to idle.

   **Edge cases:**
   - Needy character moved and no cardinal-adjacent tile is empty → keep walking toward them (recalculate Dest).
   - Needy character died → drop vessel at current position and clear intent.
   - Vessel taken by someone else during procurement → intent nils via `ProcureFailed`.

7. **`continueIntent` early-return block** (`system/movement.go`): Add `ActionHelpWater`. Three navigation paths, distinguished by world state:
   - TargetItem on the ground (`gameMap.ItemAt` matches) → recalculate path toward TargetItem. (Procurement phase.)
   - TargetItem in inventory, vessel has no water (`!vesselHasLiquid`) → recalculate path toward Dest (water-adjacent tile). (Fill phase.)
   - TargetItem in inventory, vessel has water (`vesselHasLiquid`) → recalculate path toward `TargetCharacter.Pos()`. (Delivery phase — needer may have moved.)
   - TargetItem gone (not on ground, not in inventory) → return nil.

8. **Action log messages**:
   - On intent creation: "[Helper] is fetching water for [Needer]"
   - On drop: "[Helper] brought water to [Needer]"

[TEST] Create a test world with a character near Crisis thirst, an empty vessel, and a water source:
- Idle character notices crisis, picks up vessel, fills at pond, delivers to thirsty character
- Thirsty character drinks from the delivered ground water vessel
- Helper already carrying water skips procurement and fill, walks directly to needer
- Helper carrying empty vessel fills it then delivers
- No vessel available → helper does normal idle activity
- No vessel available but needer also has Crisis hunger → helper brings food instead (fallback)
- Two crisis characters at different distances → helper targets nearer one regardless of crisis type
- Same character with both Crisis hunger and thirst → thirst addressed first
- Multiple idle characters may respond to the same thirst crisis

[RETRO]

---

### Step 4: [DOCS] + [RETRO]

After all three steps are human-tested and confirmed working:
- [DOCS] Update game-mechanics.md, architecture.md, CLAUDE.md via `/update-docs`
- [RETRO] Run `/retro` on the full helping feature
- Update randomideas.md: move Helping from "Ready Ideas" to completed, keep deferred items (Option B trigger, health crisis, inventory management) as future enhancements
