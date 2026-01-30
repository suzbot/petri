### Deferred Enhancements & Trigger Points

Technical and Feature items analyzed and consciously deferred until trigger conditions are met.

| Enhancement                            | Triggers (implement when ANY is met)                                              |
| -------------------------------------- | --------------------------------------------------------------------------------- |
| **Parallel Intent Calculation**        | Character count ≥ 16; Intent calc exceeds ~1ms/char; Tick time exceeds 50ms       |
| **EventType for ActionLog**            | Event filtering in UI; Event-driven behavior; Event persistence                   |
| **Category type formalization**        | Adding non-plant categories (tools, crafted items); Need category-based filtering |
| **Category-based world spawning**      | When Category is formalized; Filter spawn loop by category (plants spawn, tools don't) |
| **Performance optimizations**          | Noticeable lag; Profiling shows bottlenecks                                       |
| **Preference formation for beverages** | Other beverages (non-spring) introduced                                           |
| **Depression and Rage mechanics**      | Job acceptance logic implemented                                                  |
| **Testify test package**               | Test complexity warrants it; Current assertion patterns become unwieldy           |
| **Wandering activity**                 | More idle activities needed; Looking feels repetitive; Want more organic movement |
| **ItemType constants**                 | Adding new item types; Want compile-time safety for item type checks              |
| **Activity enum**                      | Need to enumerate/filter activities; Activity-based game logic beyond display     |
| **Order completion criteria refactor** | Adding new order types; Completion logic scattered across update.go becomes unwieldy |
| **Unify pickup/drop code (picking.go)** | Adding order/activity that involves picking up items; Duplicated vessel-seeking pattern becomes third instance; update.go ActionPickup handler grows with order-specific logic |
| **Feature capability derivation**      | Adding new feature types; DrinkSource/Bed bools become redundant                  |
| **Action log retention policy**        | Implementing character memory; May need world time vs real time consideration     |
| **UI color style map**                 | Adding new colors frequently; Switch statement maintenance becomes tedious        |
| **Cobra CLI migration**                | Next time we want to add a flag; Current flag parsing becomes unwieldy            |
| **Kind field for fine-grained preferences** | Adding second recipe with same ItemType (e.g., clay pot vessel); Preferences need to distinguish recipe outputs |

### Future Enhancement Details

**Balance tuning candidates** (see config.go for current values):

- Prefernce formation still feels too easy
- Activity durations
- Satisfaction cooldown
- Sleep wake thresholds
- Movement energy drain
- Health tier thresholds
- Mood adjustment rates and modifiers

---

**Performance optimizations:**

- Skip wake checks for sleeping characters unless tier changed
- Dead character filtering at map level
- Cache nearest features if map static-

---

**Unify pickup/drop code (picking.go):**

Current concerns identified in Phase 6 Clean-up:

1. **Mixed abstraction levels in foraging.go**: Contains both low-level helpers (`AddToVessel`, `IsVesselFull`) and high-level intent finding (`findForageIntent`). File name suggests foraging-specific but functions are used by harvesting.

2. **Duplicated "vessel-seeking" pattern**: Both `findForageIntent` and `findHarvestIntent` have ~30 lines of similar code: check if carrying vessel → look for available vessel → create intent to move/pickup vessel. Different only in: target finding method, conflict resolution (skip vs drop), log category.

3. **update.go ActionPickup knows too much**: Handler has order-specific completion logic (harvest), vessel continuation logic, drop-before-pickup logic. Each new order type would add more conditionals here.

4. **Scattered vessel logic**: `AddToVessel`, `IsVesselFull`, `CanVesselAccept`, `FindAvailableVessel`, `FindNextVesselTarget` all in foraging.go but used across activities.

Suggested refactor approach:
- Create `picking.go` with: `Pickup()`, `Drop()`, vessel helpers, and a `PickupIntentBuilder` that encapsulates the vessel-seeking pattern with configurable target-finding and conflict-resolution strategies
- Move completion logic out of update.go into order-specific handlers or a completion criteria system
- Consider `vessel.go` for container-specific logic if it grows further

---

**Kind field for fine-grained preferences:**

Currently preferences form on `ItemType` (e.g., "vessel"), not specific recipe outputs (e.g., "hollow gourd" vs "clay pot"). When multiple recipes produce the same ItemType, characters can't distinguish them for preferences or recipe selection.

Options:
- **Option A: Add `Kind` field to Item** - New field set from recipe ID. `ItemType` = broad category ("vessel"), `Kind` = specific type ("hollow-gourd"). Clean separation from `Name` which could be customizable later.
- **Option C: Store `RecipeID` on Item** - Look up recipe for display/preference purposes. More indirect but clear provenance.

Also consider: should `Kind` default to `ItemType` for non-crafted items, enabling finer preference distinctions for natural items too?

---

**Parallel Intent pattern:**

```go
var wg sync.WaitGroup
for _, char := range chars {
    wg.Add(1)
    go func(c *entity.Character) {
        defer wg.Done()
        c.Intent = system.CalculateIntent(c, items, gameMap, log)
    }(char)
}
wg.Wait()
```
