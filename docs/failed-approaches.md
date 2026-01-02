# Failed Approaches

Documentation of approaches that were attempted and abandoned, for future reference.

## Multi-character per position (attempted, abandoned)

We attempted to allow up to 2 characters per position to enable behaviors like:
- Characters passing each other in narrow spaces
- Carrying/grappling another character
- Sharing a bed

### Implementation tried

1. Changed `characterByPos map[Pos]*entity.Character` to `charactersByPos map[Pos][]*entity.Character` (slice per position)
2. Added `CharactersAt(x, y)` returning all characters at position
3. Added `CharacterCountAt(x, y)` for collision checks
4. Modified `MoveCharacter` to manage slice operations (append to new position, remove from old)
5. Added rendering logic to flash between multiple characters at same position
6. Added collision prevention to cap at 2 characters per position

### Problems encountered

- Characters would "disappear" after converging - only one would be visible/moving going forward
- The slice manipulation in `MoveCharacter` had subtle bugs with slice references
- Despite hard limits in `MoveCharacter`, 3+ characters would end up at same position
- Multiple debugging attempts failed to identify the root cause

### Root cause suspected but not confirmed

- Possible race condition in sequential intent application within same tick
- Slice reference issues when modifying `charactersByPos` entries
- Map lookup returning stale data after modification

### Recommendations for future implementation

For future implementation (e.g., carrying, grappling, bed sharing):
- Consider explicit "passenger" state on characters rather than position sharing
- Use separate tracking for characters in special states (being carried, grappling, etc.)
- Test thoroughly with debug logging at each step of position changes
- Consider using unique IDs for position verification rather than pointer comparison

## Single-character per position (current, working)

After the multi-character approach failed, we simplified to max 1 character per position:
- `characterByPos map[Pos]*entity.Character` (single pointer, not slice)
- `MoveCharacter()` returns false if position occupied
- `findAlternateStep()` finds alternate routes when blocked

This approach works correctly. The earlier "failures" were due to testing with a stale binary.
