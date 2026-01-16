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
| **Feature capability derivation**      | Adding new feature types; DrinkSource/Bed bools become redundant                  |
| **UI color style map**                 | Adding new colors frequently; Switch statement maintenance becomes tedious        |
| **Cobra CLI migration**                | Next time we want to add a flag; Current flag parsing becomes unwieldy            |

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
