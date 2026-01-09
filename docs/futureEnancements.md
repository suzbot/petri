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

**Applied fix** (2026-01-09): Reduced spawn interval and hunger rate.
See `config.ItemLifecycle` and `config.HungerIncreaseRate` for current values.

**Results after tuning**:
| Metric | Before | After |
|--------|--------|-------|
| First death (avg) | ~8 min | ~16-19 min |
| Total wipe | ~10 min | ~22+ min |
| Edible items | Depletes to 0 | Stable when population balanced |
| Death causes | 100% starvation | Mixed (starvation + poison) |

Target was ~30 min to wipe. Provides gameplay breathing room for future agriculture features.

#### Flower Overpopulation - RESOLVED

**Original issue**: Flowers grew unchecked (20 → 874) since they spawn but aren't consumed.

**Applied fix** (2026-01-09): Flower death timer system

Implementation:
- Added `DeathTimer` field to Item struct
- Added `ItemLifecycle` config map with spawn/death intervals per item type
- Simple timer + variance (no chance roll): timer expires → flower dies
- Renamed `spawning.go` → `lifecycle.go` to contain both spawn and death logic
- Initial flowers: staggered death timers to avoid synchronized die-off

**Results after tuning**:
| Metric | Before | After |
|--------|--------|-------|
| Flower population | 20 → 874 (exploding) | 20 → 35-50 (stable) |
| Edible crowding | Flowers dominated map | Balanced ecosystem |

Scalability path: Config map by ItemType now; migrate to per-variety rates when needed.
See `config.ItemLifecycle` for current values.

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
