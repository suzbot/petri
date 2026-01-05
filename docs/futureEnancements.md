### Deferred Enhancements & Trigger Points

Technical and Feature items analyzed and consciously deferred until trigger conditions are met.

| Enhancement                            | Triggers (implement when ANY is met)                                              |
| -------------------------------------- | --------------------------------------------------------------------------------- |
| **Parallel Intent Calculation**        | Character count ≥ 16; Intent calc exceeds ~1ms/char; Tick time exceeds 50ms       |
| **EventType for ActionLog**            | Event filtering in UI; Event-driven behavior; Event persistence                   |
| **Category field for Items**           | When adding non-spawining itemtypes                                               |
| **Performance optimizations**          | Noticeable lag; Profiling shows bottlenecks                                       |
| **Preference formation for beverages** | Other beverages (non-spring) introduced                                           |
| **Depression and Rage mechanics**      | Job acceptance logic implemented                                                  |
| **Testify test package**               | Test complexity warrants it; Current assertion patterns become unwieldy           |
| **Wandering activity**                 | More idle activities needed; Looking feels repetitive; Want more organic movement |

### Future Enhancement Details

**Balance tuning candidates** (see config.go for current values):

- Activity durations
- Satisfaction cooldown
- Sleep wake thresholds
- Movement energy drain
- Health tier thresholds
- Preference formation chances (inflated for testing)
- Mood adjustment rates and modifiers
- Spawn rate vs consumption rate: 4 characters eating shouldn't outpace plant respawning (adjust spawn interval/chance or hunger rate)
- Flower overpopulation: Flowers reproduce but aren't consumed, may outpace edible items over time. Consider: slower flower reproduction rate, OR allowing flower consumption in severe hunger situations

**Performance optimizations:**

- Skip wake checks for sleeping characters unless tier changed
- Dead character filtering at map level
- Cache nearest features if map static

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
