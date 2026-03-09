# Petri Flow Diagrams

Visual reference for how character decision-making and game systems connect across the codebase. Ordered from most foundational (structural map) to most specific (individual action state machines).

---

## Architecture: Complete Call Graph

Every cross-file dependency. If the code is spaghetti, this is the spaghetti.

```mermaid
flowchart TD
    subgraph "Game Loop (ui/)"
        update["update.go
        updateGame()"]
        apply["apply_actions.go
        applyIntent()"]
    end

    subgraph "Decision Layer (system/)"
        intent["intent.go
        CalculateIntent()
        continueIntent()"]
        movement_spatial["movement.go
        NextStepBFS()
        Pathfinding"]
        discretionary["discretionary.go
        selectDiscretionaryActivity()"]
        orders["order_execution.go
        selectOrderActivity()
        findOrderIntent()
        findBuildFenceIntent()
        findBuildHutIntent()
        selectConstructionMaterial()
        findNearestMaterialNotAtSite()"]
        helping["helping.go
        selectHelpingActivity()
        findHelpFeedIntent()
        findHelpWaterIntent()"]
    end

    subgraph "World Layer (game/)"
        worldmap["map.go
        ConstructAt()
        HasUnbuiltConstructionPositions()
        constructionMaterialFeasible()"]
    end

    subgraph "Action Layer (system/)"
        survival["survival.go
        UpdateSurvival()"]
        consume["consumption.go
        Consume/Drink"]
        forage["foraging.go
        findForageIntent()"]
        fetch["fetch_water.go
        findFetchWaterIntent()"]
        pick["picking.go
        Pickup/Drop
        EnsureHas*/RunVessel*"]
        craft["crafting.go
        CreateVessel/CreateHoe"]
        talk["talking.go
        Knowledge transmission"]
        pref["preference.go
        TryFormPreference()"]
    end

    subgraph "Data Layer (entity/)"
        char["character.go
        Stats, tiers, inventory"]
        item["item.go
        Items, properties"]
        activity["activity.go
        ActivityRegistry"]
        order["order.go
        Order struct"]
        recipe["recipe.go
        RecipeRegistry"]
        knowledge["knowledge.go
        Knowledge struct"]
    end

    %% Game Loop → Decision
    update --> intent

    %% Game Loop → Execution
    update --> apply
    update --> survival

    %% Execution (applyIntent) → Action
    apply --> consume
    apply --> pick
    apply --> craft
    apply --> talk
    apply --> forage
    apply --> orders

    %% Execution → Data
    apply --> char
    apply --> item
    apply --> recipe

    %% Decision: intent internals
    intent --> movement_spatial
    intent --> talk

    %% Decision: intent → bucket routing
    intent --> orders
    intent --> helping
    intent --> discretionary

    %% Decision: intent → Data
    intent --> char
    intent --> item

    %% Decision: discretionary dispatches
    discretionary --> forage
    discretionary --> fetch
    discretionary --> talk
    discretionary --> pick

    %% Decision: discretionary → Data
    discretionary --> activity
    discretionary --> char

    %% Decision: orders dispatches
    orders --> pick
    orders --> craft
    orders --> worldmap

    %% Decision: orders → Data
    orders --> activity
    orders --> order
    orders --> item
    orders --> char
    orders --> recipe

    %% Decision: helping
    helping --> pick
    helping --> char
    helping --> item

    %% Action: forage
    forage --> pick
    forage --> item
    forage --> char

    %% Action: fetch
    fetch --> pick
    fetch --> item

    %% Action: consumption
    consume --> char
    consume --> item
    consume --> pref
    consume --> activity
    consume --> knowledge

    %% Action: survival
    survival --> char

    %% Action: picking
    pick --> char
    pick --> item
    pick --> recipe

    %% Action: crafting
    craft --> char
    craft --> item
    craft --> recipe

    %% Action: talking
    talk --> char
    talk --> knowledge

    %% Action: preference
    pref --> char
    pref --> item
```

---

## Architecture: Decision Flow Only

What happens during `CalculateIntent()` — choosing what to do. No action execution.

```mermaid
flowchart TD
    intent["intent.go
    CalculateIntent()
    continueIntent()
    findFoodIntent()
    findDrinkIntent()
    findSleepIntent()"]

    orders["order_execution.go
    selectOrderActivity()
    findOrderIntent()"]
    helping["helping.go
    selectHelpingActivity()
    findHelpFeedIntent()
    findHelpWaterIntent()"]
    discretionary["discretionary.go
    selectDiscretionaryActivity()"]
    forage["foraging.go
    findForageIntent()"]
    fetch["fetch_water.go
    findFetchWaterIntent()"]
    talk["talking.go
    findTalkIntent()"]
    pick["picking.go
    EnsureHasVesselFor()
    CanPickUpMore()"]

    char["character.go
    Stats, tiers, preferences"]
    item["item.go
    Food/drink/vessel scoring"]
    activity["activity.go
    ActivityRegistry"]
    order["order.go
    Order struct"]
    recipe["recipe.go
    RecipeRegistry"]

    %% Priority routing (Orders → Helping → Discretionary)
    intent -->|"Priority routing"| orders
    intent -->|"Priority routing"| helping
    intent -->|"Priority routing"| discretionary
    intent --> char
    intent --> item

    %% Order evaluation
    orders --> activity
    orders --> order
    orders --> item
    orders --> char
    orders --> recipe

    %% Helping evaluation
    helping --> pick
    helping --> char
    helping --> item

    %% Discretionary roll
    discretionary -->|"Random roll"| forage
    discretionary -->|"Random roll"| fetch
    discretionary -->|"Random roll"| talk
    discretionary --> pick
    discretionary --> activity
    discretionary --> char

    %% Forage/fetch evaluation
    forage --> pick
    forage --> item
    forage --> char
    fetch --> pick
    fetch --> item
```

---

## Architecture: Execution Flow Only

What happens during `applyIntent()` and `UpdateSurvival()` — carrying out decisions.

```mermaid
flowchart TD
    update["apply_actions.go
    applyIntent()"]
    survival["survival.go
    UpdateSurvival()"]

    consume["consumption.go
    Consume()
    ConsumeFromInventory()
    ConsumeFromVessel()
    Drink()"]
    pick["picking.go
    Pickup() / Drop()
    RunVesselProcurement()
    RunWaterFill()"]
    craft["crafting.go
    CreateVessel()
    CreateHoe()"]
    talk["talking.go
    StartTalking()
    StopTalking()
    TransmitKnowledge()"]
    forage["foraging.go
    FindForageFoodIntent()"]
    orders["order_execution.go
    CompleteOrder()"]
    pref["preference.go
    TryFormPreference()"]
    movement["movement.go
    NextStepBFS()"]

    char["character.go
    Stats, inventory"]
    item["item.go
    Properties, spawning"]
    recipe["recipe.go
    RecipeRegistry"]
    knowledge["knowledge.go
    Knowledge struct"]
    activity["activity.go
    KnowHow discovery"]

    %% update.go dispatches by ActionType
    update -->|"ActionConsume/Drink"| consume
    update -->|"ActionPickup/Drop"| pick
    update -->|"ActionCraft"| craft
    update -->|"ActionTalk"| talk
    update -->|"ActionForage"| forage
    update -->|"ActionHelpFeed/Water"| movement
    update -->|"ActionBuildFence/BuildHut"| movement
    update -->|"Order completion"| orders
    update --> char
    update --> item
    update --> recipe

    %% Survival (pre-intent, every tick)
    survival --> char

    %% Consumption side effects
    consume --> char
    consume --> item
    consume --> pref
    consume --> activity
    consume --> knowledge

    %% Picking
    pick --> char
    pick --> item
    pick --> recipe

    %% Crafting
    craft --> char
    craft --> item
    craft --> recipe

    %% Talking
    talk --> char
    talk --> knowledge

    %% Preference
    pref --> char
    pref --> item
```

---

## Main Game Tick Loop

The per-tick processing order in `updateGame()` (`internal/ui/update.go`).

```mermaid
flowchart TD
    tick[/"Tick (every frame)"/]
    tick --> survival["UpdateSurvival()
    Decay hunger/thirst, drain/restore energy,
    sleep/wake checks"]
    survival --> items["Item Updates
    Spawn timers, death timers,
    sprout timers, ground spawning"]
    items --> calc["CalculateIntent() — per character
    Evaluate needs, choose action"]
    calc --> apply["applyIntent() — per character (apply_actions.go)
    Execute movement, consumption,
    crafting, helping, etc."]
    apply --> cleanup["Sweep completed orders"]
    cleanup --> save{"Auto-save
    due?"}
    save -->|Yes| dosave[Save game]
    save -->|No| done([End tick])
    dosave --> done
```

---

## Intent Priority Hierarchy

The core decision tree in `CalculateIntent()` (`internal/system/intent.go`) — what a character decides to do each tick.

```mermaid
flowchart TD
    start([CalculateIntent])
    start --> dead{Dead or<br/>sleeping?}
    dead -->|Yes| none1([No intent])
    dead -->|No| frust{Frustrated?}
    frust -->|Yes| none2([Idle — wait<br/>out timer])
    frust -->|No| existing{Has existing<br/>intent?}

    existing -->|Idle activity,<br/>no urgent need| cont1["continueIntent()
    Keep doing current activity"]
    existing -->|Need-driven,<br/>still valid| cont2["continueIntent()
    Keep pursuing current need"]
    existing -->|Need satisfied<br/>or preempted| clear[Clear intent,<br/>re-evaluate]

    clear --> tier
    existing -->|No existing| tier

    tier{{"Determine max<br/>urgency tier"}}

    tier -->|"Tier 2-4<br/>(Moderate–Crisis)"| priority["Build priority list
    Thirst > Hunger > Health > Energy
    (within same tier)"]
    priority --> trystat["For each stat in priority:
    findDrinkIntent / findFoodIntent /
    findHealIntent / findSleepIntent"]
    trystat --> found{Intent<br/>found?}
    found -->|Yes| pauseorder["Pause order if any,
    return intent"]
    found -->|No, Severe+| frustcheck["Track failure count
    → may trigger frustration"]
    found -->|No| routing

    tier -->|"Tier 1 (Mild)"| mild{Has assigned<br/>order?}
    mild -->|Yes| invcheck{Food/water<br/>in inventory?}
    invcheck -->|Yes| pause1["Pause order,<br/>consume from inventory"]
    invcheck -->|No| routing
    mild -->|No| mildloop["Priority loop
    (same as Moderate)"]
    mildloop --> mildresult{Intent<br/>found?}
    mildresult -->|Yes| ret1([Return intent])
    mildresult -->|No| routing

    tier -->|"Tier 0 (None)<br/>All stats OK"| routing

    frustcheck --> routing
    routing["Priority routing:
    selectOrderActivity()
    → selectHelpingActivity()
    → selectDiscretionaryActivity()"]
    routing --> routeresult{Intent<br/>found?}
    routeresult -->|Yes| ret2([Return intent])
    routeresult -->|"No, Moderate+ needs"| stuck(["Stuck
    (can't meet needs)"])
    routeresult -->|"No, Mild or no needs"| idle([Idle])
```

---

## Discretionary Activity Selection

`selectDiscretionaryActivity()` (`internal/system/discretionary.go`) — leisure activities chosen when no needs, orders, or helping are active.

```mermaid
flowchart TD
    start([selectDiscretionaryActivity])

    start --> cooldown{Cooldown<br/>expired?}
    cooldown -->|No| none([No intent — wait])
    cooldown -->|Yes| setcooldown["Set next cooldown"]

    setcooldown --> roll{{"Random roll (0-4)"}}
    roll -->|0| r0["Look → Talk → Forage"]
    roll -->|1| r1["Talk → Look → Forage"]
    roll -->|2| r2["Forage → Look → Talk"]
    roll -->|3| r3["FetchWater → Look"]
    roll -->|4| r4([Stay idle])
    r0 --> fallback([First that succeeds,<br/>or nil])
    r1 --> fallback
    r2 --> fallback
    r3 --> fallback
```

---

## Multi-Phase Actions

Complex actions that manage internal phase transitions across ticks. Phase is detected from world state each tick (e.g., "is the vessel in my inventory or on the ground?"), not stored explicitly.

### Help Water

```mermaid
stateDiagram-v2
    [*] --> ProcureVessel: No vessel in inventory
    [*] --> FillVessel: Empty vessel in inventory
    [*] --> DeliverWater: Full vessel in inventory
    ProcureVessel --> FillVessel: Picked up vessel
    FillVessel --> DeliverWater: Vessel filled at water
    DeliverWater --> [*]: Dropped adjacent to needer
```

### Help Feed

```mermaid
stateDiagram-v2
    [*] --> ProcureFood: Food on ground
    [*] --> DeliverFood: Food in inventory
    ProcureFood --> DeliverFood: Picked up food
    DeliverFood --> [*]: Dropped adjacent to needer
```

### Fetch Water

```mermaid
stateDiagram-v2
    [*] --> ProcureVessel: No vessel in inventory
    [*] --> FillAtWater: Empty vessel in inventory
    ProcureVessel --> FillAtWater: Picked up vessel
    FillAtWater --> [*]: Vessel filled, intent complete
```

### Water Garden

```mermaid
stateDiagram-v2
    [*] --> ProcureVessel: No vessel
    [*] --> FillWater: Empty vessel
    [*] --> WaterTile: Full vessel
    ProcureVessel --> FillWater: Vessel acquired
    FillWater --> WaterTile: Vessel filled
    WaterTile --> FillWater: Vessel empty,\nmore tiles to water
    WaterTile --> [*]: All tiles watered
```

### Construction Order (Fence / Hut)

Same machinery for both. Material is stamped on first tile; subsequent tiles use the same material. Bundles (grass/sticks) can be consumed directly for fences (no supply-drop needed). Bricks and all hut materials use the supply-drop path: carry 2 per trip, deliver, repeat until threshold met, then build.

```mermaid
stateDiagram-v2
    [*] --> SelectMaterial: Order taken,\nno material assigned yet
    SelectMaterial --> Pickup: Material stamped\nto line/footprint
    [*] --> Pickup: Material already assigned

    state "Can build directly?" as direct_check
    Pickup --> direct_check: Items picked up
    direct_check --> Build: Yes (bundle fence:\nfull bundle in hand)
    direct_check --> Deliver: No (supply-drop:\ncarry 2 to site)

    Deliver --> Pickup: Dropped at site,\nthreshold not yet met
    Deliver --> Build: Threshold met at tile

    Build --> Pickup: Construct placed,\nnext unbuilt tile
    Build --> [*]: All tiles built
```

---

## Order Lifecycle

How an order transitions from creation to completion or abandonment.

```mermaid
stateDiagram-v2
    [*] --> Open: Player creates order
    Open --> Assigned: Character takes order\n(passes feasibility + know-how check)
    Assigned --> Paused: Character's needs interrupt\n(Moderate+ tier)
    Paused --> Assigned: Needs satisfied,\ncharacter resumes
    Assigned --> Completed: Order completion condition met
    Assigned --> Abandoned: No matching items on map\nor player cancels
    Abandoned --> Open: Cooldown expires\n(order retried)
    Completed --> [*]: Swept from order list
    Abandoned --> [*]: If infeasible at render time,\nshows "Unfulfillable" instead of "Abandoned"
```

**Feasibility at assignment**: `IsOrderFeasible` is checked before a character takes an open order. Infeasible orders are skipped and shown dimmed. For construction orders, feasibility counts only free (un-staged) materials — items already at construction-marked tiles are excluded.

**Transient nil guard**: For construction orders, when `findBuildHutIntent` returns nil (all candidates temporarily occupied), the guard checks `IsOrderFeasible` — if still feasible, the nil is treated as a transient block and the order is not abandoned.
