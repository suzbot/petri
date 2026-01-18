# Save System Plan

## Overview

Implement persistent world saving with auto-save and multiple world support. Each world is a unique simulation with its own variety registry, characters, and history. No manual save/load - the game auto-saves to preserve emergent consequences.

### Design Philosophy

Aligned with Petri's vision: worlds have unique identities, history exists in character memories, and consequences are permanent. Players don't reload to undo deaths - they live with outcomes or start fresh worlds.

---

## What Gets Saved

### World State

| Data | Notes |
|------|-------|
| **VarietyRegistry** | Preserves unique poison/healing assignments for this world |
| **Map dimensions** | Width, height |
| **Characters** | Full state including action logs |
| **Items** | All items with positions, timers, attributes |
| **Features** | Springs, leaf piles, positions |
| **Game elapsed time** | Tracked from world creation, used for action log timestamps |
| **ActionLog** | Per-character event map (kept as-is, not refactored) |

### Character State

All fields from `Character` struct:
- Identity: ID, Name, Position
- Stats: Health, Hunger, Thirst, Energy, Mood
- Status: Poisoned, PoisonTimer, IsDead, IsSleeping, AtBed
- Preferences: Full preference list
- Knowledge: Full knowledge list
- Frustration: IsFrustrated, FrustrationTimer, FailedIntentCount
- Idle state: IdleCooldown, LastLookedX/Y, HasLastLooked
- Talking state: TalkingWith (as ID reference), TalkTimer
- Cooldowns: HungerCooldown, ThirstCooldown, EnergyCooldown
- Action: ActionProgress, SpeedAccumulator, CurrentActivity
- ActionLog: Recent events (per-character bounded list)

**Note:** `Intent` is transient - recalculated on next tick after load. Don't save.

### Item State

- ID (new field needed)
- Position (X, Y)
- ItemType, Color, Pattern, Texture
- Edible, Poisonous, Healing
- SpawnTimer, DeathTimer

### Feature State

- ID (new field needed)
- Position (X, Y)
- FeatureType
- IsDrinkSource, IsBed flags

### World Metadata (separate file)

- Display name (auto-generated, user-renameable)
- Created timestamp
- Last played timestamp
- Character count, alive count (for list display without full load)

---

## File Structure

```
~/.petri/
  worlds/
    world-001/
      state.json      # Full world state
      state.backup    # Previous save (crash recovery)
      meta.json       # Display info for world list
    world-002/
      state.json
      state.backup
      meta.json
```

### Why Separate Directories

- Clean organization per world
- Backup file lives alongside main save
- Easy for users to manage (copy, delete folders)
- Future: could add per-world screenshots, exports, etc.

---

## Save Format

JSON with explicit DTO structs (not direct entity serialization).

### SaveState Structure

```go
type SaveState struct {
    Version         int                 // For future migrations
    SavedAt         time.Time
    ElapsedGameTime float64             // Total simulation time (seconds)

    MapWidth        int
    MapHeight       int

    Varieties       []VarietySave       // Full registry
    Characters      []CharacterSave
    Items           []ItemSave
    Features        []FeatureSave
    ActionLogs      map[int][]EventSave // Per-character event logs, keyed by char ID
}

type EventSave struct {
    GameTime    float64  // Elapsed game time when event occurred
    CharID      int
    CharName    string
    Type        string
    Message     string
}

type CharacterSave struct {
    ID              int
    Name            string
    X, Y            int

    // Stats
    Health          float64
    Hunger          float64
    Thirst          float64
    Energy          float64
    Mood            float64

    // Status
    Poisoned        bool
    PoisonTimer     float64
    IsDead          bool
    IsSleeping      bool
    AtBed           bool

    // Frustration
    IsFrustrated      bool
    FrustrationTimer  float64
    FailedIntentCount int

    // Idle
    IdleCooldown    float64
    LastLookedX     int
    LastLookedY     int
    HasLastLooked   bool

    // Talking (partner stored as ID, -1 if none)
    TalkingWithID   int
    TalkTimer       float64

    // Cooldowns
    HungerCooldown  float64
    ThirstCooldown  float64
    EnergyCooldown  float64

    // Action
    ActionProgress    float64
    SpeedAccumulator  float64
    CurrentActivity   string

    // Mind
    Preferences     []PreferenceSave
    Knowledge       []KnowledgeSave
}

type ItemSave struct {
    ID          int
    X, Y        int
    ItemType    string
    Color       string
    Pattern     string
    Texture     string
    Edible      bool
    Poisonous   bool
    Healing     bool
    SpawnTimer  float64
    DeathTimer  float64
}

type FeatureSave struct {
    ID          int
    X, Y        int
    FeatureType string
}

type PreferenceSave struct {
    ItemType    string
    Color       string
    Valence     int
}

type KnowledgeSave struct {
    Category    string
    ItemType    string
    Color       string
    Pattern     string
    Texture     string
}

type VarietySave struct {
    ItemType    string
    Color       string
    Pattern     string
    Texture     string
    Poisonous   bool
    Healing     bool
}

type WorldMeta struct {
    Name            string
    CreatedAt       time.Time
    LastPlayedAt    time.Time
    CharacterCount  int
    AliveCount      int
}
```

---

## Pointer Reference Handling

Entities with cross-references need ID-based serialization:

| Field | Save As | Restore By |
|-------|---------|------------|
| `Character.TalkingWith` | `TalkingWithID int` (-1 if nil) | Lookup in character map after all loaded |
| `Intent.TargetItem` | Don't save | Intent recalculated on first tick |
| `Intent.TargetFeature` | Don't save | Intent recalculated on first tick |
| `Intent.TargetCharacter` | Don't save | Intent recalculated on first tick |

### Load Order

1. Load all items → build ID→Item map
2. Load all features → build ID→Feature map
3. Load all characters → build ID→Character map
4. Resolve character cross-references (TalkingWith)
5. Rebuild Map position indices

---

## Auto-Save Behavior

### Triggers

| Event | Action |
|-------|--------|
| Game quit (q key) | Save before exit |
| Pause (space) | Save on pause |
| Periodic | Every 60 seconds of game time (not wall time) |

### Save Process

1. Pause game tick processing
2. Rename `state.json` → `state.backup`
3. Write new `state.json`
4. Update `meta.json`
5. Resume game tick processing

If write fails, `state.backup` remains valid for recovery.

### Crash Recovery

On load, if `state.json` is missing or corrupt but `state.backup` exists:
- Prompt user: "Save file corrupted. Restore from backup? (loses ~60s of progress)"
- Or auto-restore with message

---

## World Naming

### Auto-Generation

Simple approach: "World N" where N is next available number.

Future enhancement: procedural names from word lists ("Misty Hollow", "Amber Grove").

### User Rename

Add to character/details screen or new world info panel. Not in initial implementation - can iterate later.

---

## UI Flow

### Start Screen (Minimal)

Current flow:
```
Select Mode → Character Creation → Playing
```

New flow:
```
World Select → [if new] Select Mode → Character Creation → Playing
             → [if existing] Playing (load state)
```

### World Select Screen

Simple list:
```
PETRI

  > Continue "World 1" (3 alive, last played 2h ago)
    Continue "World 2" (1 alive, last played 1d ago)
    New World

  [Enter] Select  [Q] Quit
```

If only one world exists, could auto-continue. If no worlds, go straight to new world flow.

### During Play

- No save/load keys needed (auto-save handles it)
- "Saving" indicator: brief flag at top of screen during save operation
- Quit key (q) triggers save then exits

---

## Implementation Steps

### Phase A: Foundation

1. Add `ID` field to `Item` struct
2. Add `ID` field to `Feature` struct
3. Assign IDs during spawn (simple incrementing counter on Map)
4. Add `ElapsedGameTime float64` to Model, increment in game tick
5. Update ActionLog to use game time instead of wall time
6. Create `internal/save/` package
7. Define DTO structs in `save/state.go`

### Phase B: Serialization

1. Implement `ToSaveState(model *ui.Model) *SaveState`
2. Implement `FromSaveState(state *SaveState) *ui.Model`
3. Handle TalkingWith ID resolution
4. Serialize ActionLog map (per-character events)
5. Implement JSON read/write with backup rotation
6. Unit tests for round-trip serialization

### Phase C: World Management

1. Create world directory structure utilities
2. Implement `ListWorlds() []WorldMeta`
3. Implement `CreateWorld(name string) (worldID, error)`
4. Implement `LoadWorld(worldID) (*SaveState, error)`
5. Implement `SaveWorld(worldID, *SaveState) error`

### Phase D: Auto-Save Integration

1. Add auto-save timer to Model (tracks game time since last save)
2. Trigger save on pause
3. Trigger save on quit
4. Add periodic save check in game tick

### Phase E: UI Integration

1. Add world select screen (new gamePhase)
2. Modify start flow to check for existing worlds
3. Load world state when continuing
4. Add "Saving" indicator (brief flag at top of screen during save)

---

## Testing Strategy

### Unit Tests

- Round-trip serialization: save → load → compare
- Pointer resolution: TalkingWith survives save/load
- Missing fields: old saves load with defaults (version migration)
- Corrupt file handling: graceful error, backup recovery

### Integration Tests

- New world creation flow
- Continue existing world flow
- Auto-save triggers correctly
- Game state identical after save/load cycle

### Manual Testing

- Kill process mid-game, verify backup recovery works
- Play through character death, quit, continue - death persists
- Multiple worlds don't interfere with each other

---

## Future Considerations

### Not In Scope (Add Later)

- World deletion UI (users can delete folders manually)
- World rename UI
- Export/import worlds
- Save file compression
- Cloud sync

### Version Migration

`SaveState.Version` field allows future migrations:

```go
func migrate(state *SaveState) *SaveState {
    if state.Version < 2 {
        // Add default values for fields added in v2
    }
    if state.Version < 3 {
        // Transform data for v3 changes
    }
    state.Version = CurrentVersion
    return state
}
```

---

## Decisions Made

- [x] **Elapsed game time**: Yes, track from world creation. Use for action log timestamps instead of wall time.
- [x] **"Saving" indicator**: Yes, brief flag at top of screen during save.
- [x] **ActionLog storage**: Keep current design (central object with per-character map). No refactor needed - already aligns with vision.

## Related Changes

### Event Timestamps: Wall Time → Game Time

Current `Event.Timestamp` uses `time.Time` (wall clock). Change to:
- Store `GameTime float64` (seconds since world creation)
- Update `FormatElapsed()` to format game time
- Requires passing elapsed game time to `ActionLog.Add()`

This is a small change to `internal/system/actionlog.go` and call sites.

---

## Dependencies

Requires before starting:
- None (can begin immediately)

Blocks:
- Phase 5 (Resources/Inventory) - save system should be stable first
