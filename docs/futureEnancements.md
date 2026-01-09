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

---

### Balance Observation Results (2026-01-09)

Ran simulation observations: 5 runs × 2000 ticks (300 game-seconds each).
See `internal/simulation/observation_test.go` for test code.

#### Food Scarcity - CONFIRMED

| Metric | Value |
|--------|-------|
| Eat events | 46 |
| Spawn events | 20 |
| Consumption:Spawn ratio | **2.3:1** |
| Edible item trend | 40 → 0-15 |
| Starvation deaths | 1/20 (5%) |

**Root cause**: Spawn interval too slow.
- `ItemSpawnIntervalBase` (8.0s) × `initialItemCount` (60) = 480s base interval per item
- With 50% spawn chance → ~16 minutes average per successful spawn

**Tuning options** (pick one or combine):
1. Reduce `ItemSpawnIntervalBase` from 8.0 → 2.0-4.0
2. Increase `ItemSpawnChance` from 0.50 → 0.70
3. Reduce `HungerIncreaseRate` from 0.7 → 0.5

#### Flower Overpopulation - CONFIRMED

| Tick | Edible | Flowers | Flower % |
|------|--------|---------|----------|
| 0 | 40 | 20 | 33% |
| 1000 | 22-26 | 24-26 | ~50% |
| 2500+ | 0 | 24-34 | 100% |

Flowers grow 20-70% while edibles deplete. No consumers = unchecked growth.

**Tuning options** (pick one):
1. Make flowers non-reproducing (simplest)
2. Separate slower flower spawn interval
3. Allow desperate flower eating at Crisis hunger (adds gameplay depth)

#### Survival & Mood

- 95% survival rate over 300 game-seconds
- Mood distribution: 53% Joyful, 21% Neutral, 21% Miserable, 5% Unhappy
- Would degrade with longer simulation runs as food depletes

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
