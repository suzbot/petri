# Step Spec: Step 2b — Seed Identity (DD-13) + Planting + Gone to Seed

Design doc: [construction-design.md](construction-design.md)

---

## Sub-step 2b-1: SourceVarietyID + seed Kind + serialization

**Anchor story:** The player extracts seeds from tall grass. Looking at the seed in the details panel, it shows "tall grass seed" — specific to the kind of plant it came from, not just the generic "grass seed." The seed carries the genetic identity of its parent plant. The player saves and reloads — the seed's identity is preserved.

**What changes and why (DD-13):** Seeds currently encode parent identity via string derivation (`Kind = parentItemType + " seed"`, then `CreateSprout` does `TrimSuffix` to recover the parent type). This loses the parent's Kind ("tall grass") and is brittle. The fix: seeds carry `SourceVarietyID` — the parent plant's variety registry ID — and use the parent's Kind for their own Kind (`parentKind + " seed"` when Kind exists, `parentItemType + " seed"` otherwise).

**Implementation:**

1. **Data model — add `SourceVarietyID string` to:**
   - `Item` struct (`entity/item.go`) — for loose seeds in inventory or on ground
   - `ItemVariety` struct (`entity/variety.go`) — for seed varieties used in vessel stacks
   - `ItemSave` struct (`save/state.go`) — `json:"source_variety_id,omitempty"`
   - `VarietySave` struct (`save/state.go`) — `json:"source_variety_id,omitempty"`

2. **Update `NewSeed`** (`entity/item.go:222-238`) — new signature:
   - Add `sourceVarietyID string` and `parentKind string` parameters
   - Kind = `parentKind + " seed"` when parentKind is non-empty, else `parentItemType + " seed"`
   - Store `SourceVarietyID: sourceVarietyID`

3. **Update seed variety registration** (`game/variety_generation.go:261-304`) — three blocks (gourd, flower, grass):
   - Each seed variety gets `SourceVarietyID` set to the parent variety's ID
   - Grass seed variety Kind changes from `"grass seed"` to `"tall grass seed"` (uses parent's Kind)
   - Flower/gourd seed variety Kind unchanged (parents have no Kind)

4. **Update all seed creation call sites** to pass sourceVarietyID and parentKind:
   - `applyExtract` (`apply_actions.go:1420`) — generate variety ID from plant attributes, pass `plant.Kind`
   - `Consume` (`consumption.go:134`) — generate variety ID from gourd item attributes, pass `""` (gourds have no Kind)
   - `ConsumeFromInventory` (`consumption.go:263`) — same as Consume
   - `ConsumeFromVessel` (`consumption.go:453`) — generate variety ID from variety attributes, pass `""` (gourds have no Kind)
   - `findExtractIntent` synthetic seed (`order_execution.go:1089`) — pass target variety ID and Kind

5. **Update `ConsumePlantable`** (`picking.go:407-419`) — copy `SourceVarietyID` from `stack.Variety.SourceVarietyID` when reconstructing item from vessel stack

6. **Serialization** — add SourceVarietyID to:
   - `itemsToSave` / `itemFromSave` (`serialize.go`)
   - `varietiesToSave` / `varietiesFromSave` (`serialize.go`)
   - `StackSave` does NOT need SourceVarietyID — vessel stacks look up their variety from the registry, which carries it after loading

**Tests (TDD):**
- `NewSeed` with grass parent (Kind="tall grass") → seed has Kind="tall grass seed" and correct SourceVarietyID
- `NewSeed` with flower parent (no Kind) → seed has Kind="flower seed" and correct SourceVarietyID
- Seed variety registration: grass seed variety has Kind="tall grass seed" and SourceVarietyID matching parent grass variety
- `ConsumePlantable` from vessel: reconstructed seed item has SourceVarietyID
- Round-trip serialization: loose seed with SourceVarietyID survives save/load
- Round-trip serialization: seed variety with SourceVarietyID survives save/load (variety registry round-trip)

**Architecture:** Follows serialization checklist in architecture.md (new entity field → save struct + serialize/deserialize). Seed variety registration follows existing pattern in variety_generation.go. `SourceVarietyID` on both Item and ItemVariety follows the pattern of `Plantable` which also lives on both.

**Values:** Isomorphism (seed carries genetic identity of parent), Source of Truth Clarity (variety registry is single source for plant identity), Follow the Existing Shape (variety registration, serialization patterns).

**Files:**
- `internal/entity/item.go` — SourceVarietyID on Item, NewSeed signature
- `internal/entity/variety.go` — SourceVarietyID on ItemVariety
- `internal/save/state.go` — ItemSave, VarietySave
- `internal/ui/serialize.go` — itemsToSave/itemFromSave, varietiesToSave/varietiesFromSave
- `internal/game/variety_generation.go` — seed variety registration
- `internal/ui/apply_actions.go` — applyExtract
- `internal/system/consumption.go` — three consumption paths
- `internal/system/order_execution.go` — findExtractIntent synthetic seed
- `internal/system/picking.go` — ConsumePlantable
- `internal/entity/item_test.go` — NewSeed tests
- `internal/game/variety_generation_test.go` — seed variety registration tests
- `internal/system/picking_test.go` — ConsumePlantable tests
- `internal/ui/serialize_test.go` — round-trip tests

[TEST] Extract seeds from tall grass — verify details panel shows "tall grass seed." Extract flower seeds — verify "flower seed." Save and reload — verify seeds preserved with correct Kind and identity.

[DOCS]

[RETRO]

---

## Sub-step 2b-2: CreateSprout + PlantableItemExists + planting verification

**Anchor story:** The player creates a Plant order for tall grass seeds. The order shows as fulfillable (seeds exist in a vessel). A character plants a tall grass seed in tilled soil — a sprout appears and matures into tall grass, identical to the wild tall grass it came from, with Kind="tall grass" and matching color. Flower seeds planted likewise grow into flowers with correct color.

**Implementation:**

1. **Update `CreateSprout`** (`entity/item.go:313-337`) — replace string derivation with variety-based creation. Change signature to accept `*ItemVariety` (the parent variety, resolved by caller). The function sets ItemType, Kind, Color, Pattern, Texture, and Edible all from the variety. No TrimSuffix.
   - Callers: `applyPlant` (production), two test files
   - `applyPlant` (`apply_actions.go:639`): after `ConsumePlantable` returns the seed, look up parent variety via `registry.Get(plantedItem.SourceVarietyID)`. Pass to `CreateSprout`.

2. **Fix `PlantableItemExists`** (`order_execution.go:605-646`) — the vessel-contents check (lines 613-619 and 633-641) currently compares `stack.Variety.ItemType == targetType`, which fails for seeds (ItemType="seed" vs targetType="tall grass seed"). Replace with `stack.Variety.Plantable && (stack.Variety.ItemType == targetType || stack.Variety.Kind == targetType)`. This mirrors how `isPlantableMatch` already works for loose items — **Follow the Existing Shape**.

**Tests (TDD):**
- `PlantableItemExists` returns true when flower seeds are in a ground vessel
- `PlantableItemExists` returns true when tall grass seeds are in a carried vessel
- Plant flower seed → sprout appears with correct ItemType ("flower") and color
- Plant grass seed → sprout appears with correct ItemType ("grass") and Kind ("tall grass")

**Architecture:** `PlantableItemExists` fix mirrors `isPlantableMatch` (same check on variety instead of item). `CreateSprout` signature change follows Isomorphism — the variety IS the parent identity.

**Values:** Anchor to Intent (tests validate "planted seed grows into correct plant"), Follow the Existing Shape (PlantableItemExists mirrors isPlantableMatch), Source of Truth Clarity (variety registry resolves parent identity, not string derivation).

**Files:**
- `internal/entity/item.go` — CreateSprout signature change
- `internal/ui/apply_actions.go` — applyPlant variety lookup
- `internal/system/order_execution.go` — PlantableItemExists vessel fix
- `internal/system/picking_test.go` — PlantableItemExists tests
- `internal/entity/item_test.go` — CreateSprout tests (update existing + new grass test)

[TEST] Create a Plant order when seeds exist only in a vessel — verify order shows as fulfillable. Plant tall grass seeds in tilled soil — verify sprout matures into tall grass with correct Kind. Plant flower seeds — verify correct flower with color.

[DOCS]

[RETRO]

---

## Sub-step 2b-3: "Gone to seed" details panel indicator

**Anchor story:** The player selects a flower that has seeds available. The details panel shows "Gone to seed" in dusky earth color below the "Growing" line. They extract from it — "Gone to seed" disappears. In debug mode, while on cooldown, the panel shows "Seed cooldown: 45s". When the timer expires, "Gone to seed" reappears. Non-extractable plants (berries, mushrooms, gourds) never show either indicator.

**Implementation:** Add conditional display in `renderDetails()` in `view.go`, after the existing "Growing" line (around line 911). For items where `config.ExtractableTypes[item.ItemType]` is true and `item.Plant != nil && item.Plant.IsGrowing && !item.Plant.IsSprout`:
- If `item.Plant.SeedTimer <= 0`: show `" " + tilledDryStyle.Render("Gone to seed")`
- If `item.Plant.SeedTimer > 0` and debug mode: show `fmt.Sprintf(" Seed cooldown: %.0fs", item.Plant.SeedTimer)`

Sprouts never show either indicator (gated by `!item.Plant.IsSprout`). Non-extractable types never show either indicator (gated by `ExtractableTypes` check).

- Follows **Follow the Existing Shape** — mirrors the conditional "Growing" and "Plantable" lines in the same section of `renderDetails()`. Uses `tilledDryStyle` which already exists for dry tilled soil rendering.
- Follows **Anchor to Intent** — the indicator tells the player "this plant has seeds ready for extraction" in functional terms.

**Tests (TDD):** No unit tests — UI rendering in `renderDetails()`, falls under "no tests for UI rendering" policy.

**Files:**
- `internal/ui/view.go` — add conditional lines in `renderDetails()` item section

[TEST] Select an extractable flower in details panel — verify "Gone to seed" appears in dusky earth style below "Growing". Extract from it — verify "Gone to seed" disappears. Enable debug mode — verify "Seed cooldown: Xs" appears while timer is active. Wait for timer to expire — verify "Gone to seed" reappears. Select a berry, mushroom, or gourd — verify neither indicator ever appears. Select a sprout of an extractable type — verify no indicator appears.

[DOCS]

[RETRO]
