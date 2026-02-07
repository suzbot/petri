---
name: architecture
description: "Quick reference to Petri's architecture patterns. Use before reading lots of code to understand current design patterns."
---

## Architecture Quick Reference

Read `docs/architecture.md` for the complete reference. Key patterns:

### Core Design
- **MVU Architecture**: Bubble Tea Model-View-Update, rendering diffs automatic
- **Intent System**: Characters calculate Intent, then intents applied atomically
- **Multi-Stat Urgency**: Tiers 0-4, highest wins; tie-breaker: Thirst > Hunger > Energy
- **Stat Fallback**: If intent can't be fulfilled, falls through to next urgent stat

### Data Structures
- **Sparse Grid + Indexed Slices**: O(1) character lookups via grid, separate slices for entities
- **Simple Flags over ECS**: Boolean flags (Edible, Poisonous) not full ECS. Can evolve later.

### Item Model
- **Descriptive attributes** (ItemType, Color, Pattern, Texture) - opinions form on these
- **Functional attributes** (Edible, Poisonous, Healing) - capabilities, no opinions
- **Optional property structs**: `EdibleProperties`, `PlantProperties`, `ContainerData`

### Key Patterns to Check
| Pattern                 | Where Documented                  | When Relevant                     |
|-------------------------|-----------------------------------|-----------------------------------|
| Orders and Actions      | architecture.md § Orders          | Adding player-directed tasks      |
| Pickup/Vessel helpers   | architecture.md § Pickup Activity | Item manipulation                 |
| Component Procurement   | architecture.md § Component       | Multi-step gathering activities   |
| Activity Registry       | architecture.md § Activity        | Adding new activities             |
| Recipe System           | architecture.md § Recipe          | Adding craftable items            |
| Feature Passability     | architecture.md § Feature         | Map elements, movement            |
| Position Handling       | architecture.md § Position        | Coordinates, distance             |

### Common Pitfalls
- **Game time vs wall clock**: UI indicators that work when paused need `time.Now()`
- **Sorting stability**: Use `sort.SliceStable` with deterministic tiebreakers for merged map data
- **View transitions**: Add dimension safeguards when switching rendering approaches

---

**Next step:** Read the specific section of `docs/architecture.md` relevant to your task, then explore the code files listed there.
