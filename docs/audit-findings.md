# Code Audit Findings

**Date**: 2025-12-31
**Focus**: Scalable patterns, hard-coded strings, extensibility concerns
**Scope**: Quick scan of entity, system, game, ui packages

## Summary

This audit focuses on identifying patterns that may not scale well as the game adds more entity types, attributes, and behaviors. Primary concern: hard-coded strings that should be constants or enums.

---

## Findings

### Finding 1: Hard-coded Stat Names as Strings

**Location**: `internal/system/movement.go`, `internal/entity/character.go`

**Issue**: Stat names ("hunger", "thirst", "energy") are hard-coded strings throughout the codebase.

**Examples**:
```go
// movement.go - Intent.DrivingStat uses strings
intent.DrivingStat = "hunger"
intent.DrivingStat = "thirst"
intent.DrivingStat = "energy"

// switch statements match on strings
switch char.Intent.DrivingStat {
case "hunger":
case "thirst":
case "energy":
}
```

**Risk**:
- Typos won't be caught at compile time
- Adding new stats requires finding all string occurrences
- No IDE autocomplete support

**Recommendation**: Define stat type constants:
```go
type StatType string
const (
    StatHunger StatType = "hunger"
    StatThirst StatType = "thirst"
    StatEnergy StatType = "energy"
)
```

---

### Finding 2: Hard-coded Food Types as Strings

**Location**: `internal/entity/item.go`, `internal/system/movement.go`

**Issue**: Food types ("berry", "mushroom") are hard-coded strings.

**Examples**:
```go
// item.go
ItemType: "berry",
ItemType: "mushroom",

// movement.go - findFoodTarget
foodMatch := item.ItemType == char.FavoriteFood
```

**Risk**:
- Adding new food types requires string matching everywhere
- No validation that FavoriteFood is a valid food type

**Recommendation**: Define food type constants:
```go
type FoodType string
const (
    FoodBerry    FoodType = "berry"
    FoodMushroom FoodType = "mushroom"
)
```

---

### Finding 3: Hard-coded Color Names as Strings

**Location**: `internal/entity/item.go`, `internal/config/config.go`, `internal/ui/creation.go`

**Issue**: Colors ("red", "blue", "brown", "white") are strings scattered across files.

**Examples**:
```go
// config.go
BerryColors = []string{"red", "blue"}
MushroomColors = []string{"brown", "white", "red"}
AllColors = []string{"red", "blue", "brown", "white"}

// creation.go
colorOptions = []string{"red", "blue", "white", "brown"}
```

**Risk**:
- Color lists are duplicated and could drift out of sync
- No single source of truth for valid colors

**Recommendation**: Centralize color definitions:
```go
type Color string
const (
    ColorRed   Color = "red"
    ColorBlue  Color = "blue"
    ColorBrown Color = "brown"
    ColorWhite Color = "white"
)
var AllColors = []Color{ColorRed, ColorBlue, ColorBrown, ColorWhite}
```

---

### Finding 4: Action Log Event Types as Strings

**Location**: `internal/system/actionlog.go`, throughout system package

**Issue**: Event types ("hunger", "thirst", "activity", "movement", "consumption", "poison", "sleep", "death", "health") are untyped strings.

**Examples**:
```go
log.Add(char.ID, char.Name, "hunger", "Getting hungry")
log.Add(char.ID, char.Name, "activity", "Idle (no needs)")
log.Add(char.ID, char.Name, "consumption", fmt.Sprintf(...))
```

**Risk**:
- No validation of event types
- Filtering/categorizing events requires string matching
- Adding new event types has no compile-time safety

**Recommendation**: Define event type constants:
```go
type EventType string
const (
    EventHunger      EventType = "hunger"
    EventThirst      EventType = "thirst"
    EventActivity    EventType = "activity"
    // etc.
)
```

---

### Finding 5: Activity Strings for CurrentActivity

**Location**: `internal/entity/character.go`, `internal/system/*.go`

**Issue**: CurrentActivity uses free-form strings with no constants.

**Examples**:
```go
char.CurrentActivity = "Idle"
char.CurrentActivity = "Drinking"
char.CurrentActivity = "Moving to spring"
char.CurrentActivity = "Sleeping (in bed)"
char.CurrentActivity = "Frustrated"
char.CurrentActivity = "Consuming " + itemName
```

**Risk**:
- UI display depends on string matching
- Inconsistent formatting (some have parentheses, some don't)
- Hard to enumerate all possible activities

**Recommendation**: Consider activity enum with display strings:
```go
type Activity int
const (
    ActivityIdle Activity = iota
    ActivityDrinking
    ActivityMovingToSpring
    // etc.
)
func (a Activity) String() string { ... }
```

---

### Finding 6: Tier Threshold Logic Repeated

**Location**: `internal/entity/character.go`

**Issue**: Four nearly identical tier calculation methods (HungerTier, ThirstTier, EnergyTier, HealthTier) with similar switch logic.

**Pattern**:
```go
func (c *Character) HungerTier() int {
    switch {
    case c.Hunger >= hungerCrisisThreshold: return TierCrisis
    case c.Hunger >= hungerSevereThreshold: return TierSevere
    // ...
    }
}
// Repeated for Thirst, Energy, Health with inverted logic for some
```

**Risk**:
- Adding a new stat requires copying the entire pattern
- Changes to tier logic must be made in multiple places

**Recommendation**: Consider a generic tier calculator:
```go
func calculateTier(value float64, thresholds [4]float64, inverted bool) int
```

---

### Finding 7: Level Description Methods Repeated

**Location**: `internal/entity/character.go`

**Issue**: Four nearly identical level description methods (HungerLevel, ThirstLevel, EnergyLevel, HealthLevel).

**Risk**: Same as Finding 6 - adding stats requires copying boilerplate.

**Recommendation**: Data-driven approach with level descriptions in a map or struct.

---

### Finding 8: Feature Type Identification

**Location**: `internal/entity/feature.go`

**Issue**: Feature types use both enum (FeatureType) AND boolean flags (DrinkSource, Bed).

**Examples**:
```go
type Feature struct {
    FType       FeatureType  // FeatureSpring or FeatureLeafPile
    DrinkSource bool         // Redundant with FType?
    Bed         bool         // Redundant with FType?
}
```

**Risk**:
- Redundant data that could get out of sync
- Adding new feature types requires updating both enum AND flags

**Recommendation**: Derive capabilities from FType:
```go
func (f *Feature) IsDrinkSource() bool {
    return f.FType == FeatureSpring
}
```
Or use a capabilities map if features can have multiple capabilities.

---

## Recommendations Summary

### High Priority (Address Before Adding New Content)

1. **Define type constants for strings** used in logic:
   - StatType (hunger, thirst, energy)
   - FoodType (berry, mushroom)
   - Color (red, blue, brown, white)
   - EventType (for action log)

2. **Centralize color definitions** - single source of truth in config

### Medium Priority (Address When Refactoring)

3. **Consolidate tier calculation logic** into a reusable function
4. **Consolidate level description logic** into data-driven approach
5. **Remove redundant Feature flags** - derive from FType

### Low Priority (Nice to Have)

6. **Activity enum** - less critical since it's mostly display

---

### Finding 9: Poison Config Uses String Concatenation Keys

**Location**: `internal/game/world.go`

**Issue**: Poison configuration uses string concatenation as map keys.

**Examples**:
```go
type PoisonConfig map[string]bool

key := combo.ItemType + ":" + combo.Color
cfg[key] = true

func (pc PoisonConfig) IsPoisonous(itemType, color string) bool {
    return pc[itemType+":"+color]
}
```

**Risk**:
- Delimiter could appear in values (unlikely but possible)
- No type safety on the composite key
- Pattern doesn't scale to more complex attribute combinations

**Recommendation**: Use a struct key or nested map:
```go
type ItemKey struct {
    Type  FoodType
    Color Color
}
type PoisonConfig map[ItemKey]bool
```

---

### Finding 10: UI Color Styling Uses String Switch

**Location**: `internal/ui/view.go:407-414`

**Issue**: Color-to-style mapping uses string switch statement.

**Examples**:
```go
switch v.Color {
case "red":
    sym = redStyle.Render(sym)
case "blue":
    sym = blueStyle.Render(sym)
case "brown":
    sym = brownStyle.Render(sym)
case "white":
    sym = whiteStyle.Render(sym)
}
```

**Risk**:
- Adding new colors requires updating this switch
- No compile-time check if a color is missing from the switch
- Duplicates color knowledge from config

**Recommendation**: Use a map from Color type to style:
```go
var colorStyles = map[Color]lipgloss.Style{
    ColorRed:   redStyle,
    ColorBlue:  blueStyle,
    // etc.
}
sym = colorStyles[v.Color].Render(sym)
```

---

## String Occurrence Counts

Quick grep analysis of hard-coded strings in `/internal`:

| Pattern | Occurrences | Files |
|---------|-------------|-------|
| Stat names ("hunger", "thirst", etc.) | 52 | 6 |
| Food types ("berry", "mushroom") | 39 | 10 |
| Color names ("red", "blue", etc.) | 80 | 10 |

---

## Scalability Concerns

When adding new entities/attributes, current patterns require:

| To Add | Files to Modify | Risk |
|--------|-----------------|------|
| New stat (e.g., "morale") | character.go (4 methods), movement.go (priority logic), survival.go (update logic), view.go (display) | High - many touch points |
| New food type | item.go, config.go, movement.go (findFoodTarget), creation.go | Medium |
| New color | config.go (multiple arrays), creation.go, potentially view.go | Medium - arrays could drift |
| New feature type | feature.go (enum + flags), game/map.go (find methods), movement.go | Medium |

---

## Next Steps

1. Discuss findings and prioritize
2. Create constants file (e.g., `internal/types/types.go`) for shared type definitions
3. Refactor incrementally, ensuring tests pass at each step
