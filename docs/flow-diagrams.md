# Petri Flow Diagrams

Visual reference for how character decision-making and game systems connect across the codebase. Ordered from most foundational (structural map) to most specific (individual action state machines).

---

## Architecture: Complete Call Graph

Every cross-file dependency. If the code is spaghetti, this is the spaghetti.

```mermaid
flowchart TD
    subgraph "Game Loop (ui/)"
        update["update.go
        updateGame()
        applyIntent()"]
    end

    subgraph "Decision Layer (system/)"
        intent["intent.go
        CalculateIntent()
        continueIntent()"]
        movement_spatial["movement.go
        NextStepBFS()
        Pathfinding"]
        idle["idle.go
        selectIdleActivity()"]
        orders["order_execution.go
        selectOrderActivity()
        findOrderIntent()"]
        helping["helping.go
        findHelpFeedIntent()
        findHelpWaterIntent()"]
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

    %% Game Loop → Action (execution via applyIntent)
    update --> survival
    update --> consume
    update --> pick
    update --> craft
    update --> talk
    update --> forage
    update --> orders

    %% Game Loop → Data
    update --> char
    update --> item
    update --> recipe

    %% Decision: intent internals
    intent --> movement_spatial
    intent --> idle
    intent --> talk

    %% Decision: intent → Data
    intent --> char
    intent --> item

    %% Decision: idle dispatches
    idle --> orders
    idle --> helping
    idle --> forage
    idle --> fetch
    idle --> talk
    idle --> pick

    %% Decision: idle → Data
    idle --> activity
    idle --> char

    %% Decision: orders dispatches
    orders --> pick
    orders --> craft

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

    idle["idle.go
    selectIdleActivity()"]
    orders["order_execution.go
    selectOrderActivity()
    findOrderIntent()"]
    helping["helping.go
    findHelpFeedIntent()
    findHelpWaterIntent()"]
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

    %% Intent tree
    intent -->|"Tier 0 (no needs)"| idle
    intent -->|"Needs unfulfillable"| idle
    intent --> char
    intent --> item

    %% Idle dispatches
    idle -->|"Has/claims order"| orders
    idle -->|"Crisis nearby"| helping
    idle -->|"Idle roll"| forage
    idle -->|"Idle roll"| fetch
    idle -->|"Idle roll"| talk
    idle --> pick
    idle --> activity
    idle --> char

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
    update["update.go
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
    calc --> apply["applyIntent() — per character
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
    tier -->|"Tier 0 (None)<br/>All stats OK"| idle["selectIdleActivity()"]
    tier -->|"Tier 1 (Mild)"| mild{Has assigned<br/>order?}
    mild -->|Yes| invcheck{Food/water<br/>in inventory?}
    invcheck -->|Yes| pause1["Pause order,<br/>consume from inventory"]
    invcheck -->|No| idle
    mild -->|No| idle

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
    found -->|No, Moderate| idle
    frustcheck --> idle
```

---

## Idle Activity Selection

`selectIdleActivity()` (`internal/system/idle.go`) — what happens when a character has no urgent needs.

```mermaid
flowchart TD
    start([selectIdleActivity])

    start --> hasorder{Has assigned<br/>order?}
    hasorder -->|Yes| tryorder1["selectOrderActivity()
    Resume/find order work"]
    tryorder1 --> ordersuccess1{Found<br/>work?}
    ordersuccess1 -->|Yes| ret1([Return order intent])
    ordersuccess1 -->|No| cooldown

    hasorder -->|No| cooldown{Idle cooldown<br/>expired?}
    cooldown -->|No| none([No intent — wait])
    cooldown -->|Yes| setcooldown["Set next cooldown"]

    setcooldown --> neworder{Available<br/>order to claim?}
    neworder -->|Yes| tryorder2["selectOrderActivity()
    Assign + begin order"]
    tryorder2 --> ret2([Return order intent])

    neworder -->|No| crisis{Nearby character<br/>in crisis?}
    crisis -->|Thirst crisis| helpw["findHelpWaterIntent()"]
    crisis -->|Hunger crisis| helpf["findHelpFeedIntent()"]
    crisis -->|Both| helpboth["Try helpWater first,
    fall back to helpFeed"]
    helpw --> helpok{Found<br/>solution?}
    helpf --> helpok
    helpboth --> helpok
    helpok -->|Yes| ret3([Return helping intent])
    helpok -->|No| roll

    crisis -->|No crisis| roll

    roll{{"Random roll (0-4)"}}
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
    [*] --> GetVessel: No vessel in inventory
    [*] --> FillAtWater: Empty vessel in inventory
    GetVessel --> FillAtWater: Picked up vessel
    FillAtWater --> [*]: Vessel filled, intent complete
```

### Water Garden

```mermaid
stateDiagram-v2
    [*] --> FindVessel: No vessel
    [*] --> FillWater: Empty vessel
    [*] --> WaterTile: Full vessel
    FindVessel --> FillWater: Vessel acquired
    FillWater --> WaterTile: Vessel filled
    WaterTile --> FillWater: Vessel empty,\nmore tiles to water
    WaterTile --> [*]: All tiles watered
```
