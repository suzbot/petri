---
name: architecture
description: "Quick reference to Petri's architecture patterns. Use before reading lots of code to understand current design patterns."
---

## Architecture Quick Reference

Read `docs/architecture.md` for the complete reference and details on desired patterns. Key patterns:

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
| Pattern                 | Where Documented                        | When Relevant                     |
|-------------------------|-----------------------------------------|-----------------------------------|
| Action System           | architecture.md § Action System         | Adding any new character behavior |
| Adding New Actions      | architecture.md § Adding New Actions    | Checklists for need/idle/ordered  |
| `continueIntent` Rules  | architecture.md § continueIntent Rules  | Multi-phase actions, early returns|
| Orders                  | architecture.md § Orders                | Player-directed tasks             |
| Item Acquisition        | architecture.md § Item Acquisition      | Pickup, vessels, procurement      |
| Activity Registry       | architecture.md § Activity Registry     | Adding new activities             |
| Recipe System           | architecture.md § Recipe System         | Adding craftable items            |
| World & Terrain         | architecture.md § World & Terrain       | Map elements, water, tilling      |
| Position Handling       | architecture.md § Position Handling     | Coordinates, distance             |

### Common Pitfalls
- **Game time vs wall clock**: UI indicators that work when paused need `time.Now()`
- **Sorting stability**: Use `sort.SliceStable` with deterministic tiebreakers for merged map data
- **View transitions**: Add dimension safeguards when switching rendering approaches

---

**Next step:** Read the specific section of `docs/architecture.md` relevant to your task, then explore the code files listed there.
