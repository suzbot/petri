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

#### Food Scarcity - RESOLVED

**Original issue**: Consumption outpaced spawning 2.3:1, causing total party wipe at ~10 min.

**Applied fix** (2026-01-09):
| Parameter | Before | After |
|-----------|--------|-------|
| `ItemSpawnIntervalBase` | 8.0 | **3.0** |
| `HungerIncreaseRate` | 0.7 | **0.5** |

**Results after tuning**:
| Metric | Before | After |
|--------|--------|-------|
| First death (avg) | ~8 min | ~19 min |
| Total wipe | ~10 min | >37 min (1 survivor) |
| Edible items | Depletes to 0 | Stable at 33-43 |
| Death causes | 100% starvation | Mixed (starvation + poison) |

Target was ~30 min to wipe. Slightly overshoots but provides good gameplay breathing room.

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
