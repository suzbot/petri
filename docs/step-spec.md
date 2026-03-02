# Step Spec: Step 2b — Planting Verification + Gone to Seed + Save/Load

Design doc: [construction-design.md](construction-design.md)

---

## Sub-step 2b-1: Fix `PlantableItemExists` for vessel-stored seeds + planting verification

**Anchor story:** A character extracts flower seeds into a vessel. The player creates a Plant order for flower seeds. The order shows as fulfillable (not greyed out). The character plants a red flower seed — a sprout appears and matures into a red flower. Grass seeds planted in tilled soil grow into tall grass.

**Bug:** `PlantableItemExists` in `order_execution.go:605-646` has a vessel-contents check that compares `stack.Variety.ItemType == targetType`. For seeds, `Variety.ItemType` is `"seed"` but `targetType` is `"flower seed"` (the Kind). The comparison fails, so plant orders show as unfulfillable when seeds are only in vessels. A second issue: the config lookup `configs[stack.Variety.ItemType]` won't find a `"seed"` entry in `GetItemTypeConfigs()` because seeds don't have one.

**Fix:** Replace the vessel-contents check (lines 613-619 and 633-641) with `matchesPlantTargetVariety(stack.Variety, targetType, "")` combined with `stack.Variety.Plantable`. This reuses the existing matching function that already handles the seed Kind pattern (`ItemType == "seed" && targetType ends with " seed"`), and checks `Plantable` on the variety directly instead of looking it up in config.

- Follows **Follow the Existing Shape** — `matchesPlantTargetVariety` already exists in `picking.go` and handles exactly this case. `FindVesselContaining` and `ConsumePlantable` already use it.
- Follows **Anchor to Intent** — the test validates "plant order is fulfillable when seeds are in a vessel", not "returns true from PlantableItemExists".

**If planting doesn't Just Work beyond the feasibility bug:** Do NOT fix inline. Note the gap in Step 8 (Phase Wrap-Up) for investigation. Planting is downstream validation of extraction, not core extraction mechanics.

**Tests (TDD):**
- `PlantableItemExists` returns true when flower seeds are in a ground vessel
- `PlantableItemExists` returns true when grass seeds are in a carried vessel
- Plant flower seed in tilled soil → sprout appears with correct ItemType ("flower") and color
- Plant grass seed in tilled soil → sprout appears with correct ItemType ("grass") and Kind ("tall grass")

**Architecture:** Extends ordered action feasibility check (`PlantableItemExists`). Reuses `matchesPlantTargetVariety` from picking.go. No new patterns.

**Values:** Anchor to Intent (tests validate planting experience), Follow the Existing Shape (reuse existing matching function), Evidence Before Reasoning (trace identified the exact bug before proposing a fix).

**Files:**
- `internal/system/order_execution.go` — fix `PlantableItemExists` vessel checks
- `internal/system/order_execution_test.go` — feasibility tests for vessel-stored seeds
- `internal/ui/update_test.go` or `internal/system/picking_test.go` — planting verification tests (sprout creation from seeds)

[TEST] Plant extracted flower seeds in tilled soil — verify sprout appears and matures into flower with correct color. Plant grass seeds — verify they grow into tall grass. Create a Plant order when seeds exist only inside a vessel — verify the order shows as fulfillable (not greyed out).

[DOCS]

[RETRO]

---

## Sub-step 2b-2: "Gone to seed" details panel indicator

**Anchor story:** The player selects a flower that has seeds available. The details panel shows "Gone to seed" in dusky earth color below the "Growing" line. They extract from it — "Gone to seed" disappears. In debug mode, while on cooldown, the panel shows "Seed cooldown: 45s". When the timer expires, "Gone to seed" reappears. Non-extractable plants (berries, mushrooms, gourds) never show either indicator.

**Implementation:** Add conditional display in `renderDetails()` in `view.go`, after the existing "Growing" line (around line 911). For items where `config.ExtractableTypes[item.ItemType]` is true and `item.Plant != nil && item.Plant.IsGrowing && !item.Plant.IsSprout`:
- If `item.Plant.SeedTimer <= 0`: show `" " + tilledDryStyle.Render("Gone to seed")`
- If `item.Plant.SeedTimer > 0` and debug mode: show `fmt.Sprintf(" Seed cooldown: %.0fs", item.Plant.SeedTimer)`

Sprouts never show either indicator (gated by `!item.Plant.IsSprout`). Non-extractable types never show either indicator (gated by `ExtractableTypes` check).

- Follows **Follow the Existing Shape** — mirrors the conditional "Growing" and "Plantable" lines in the same section of `renderDetails()`. Uses `tilledDryStyle` which already exists for dry tilled soil rendering.
- Follows **Anchor to Intent** — the indicator tells the player "this plant has seeds ready for extraction" in functional terms.

**Tests (TDD):** No unit tests — this is UI rendering in `renderDetails()`, which falls under the "no tests for UI rendering" policy.

**Files:**
- `internal/ui/view.go` — add conditional lines in `renderDetails()` item section

[TEST] Select an extractable flower in details panel — verify "Gone to seed" appears in dusky earth style below "Growing". Extract from it — verify "Gone to seed" disappears. Enable debug mode — verify "Seed cooldown: Xs" appears while timer is active. Wait for timer to expire — verify "Gone to seed" reappears. Select a berry, mushroom, or gourd — verify neither indicator ever appears. Select a sprout of an extractable type — verify no indicator appears.

[DOCS]

[RETRO]

---

## Sub-step 2b-3: Save/load round-trip tests

**Anchor story:** The player saves a world mid-extraction — some flowers have active SeedTimers, seeds sit in vessels. They reload. SeedTimers are preserved (flowers still on cooldown), seeds still in vessels with correct varieties. Loading a pre-extraction save causes no errors — SeedTimer defaults to 0 (seeds available).

**Implementation:** Pure test-writing. The serialization code already exists:
- `PlantPropertiesSave` has `SeedTimer float64` field (`state.go:132`)
- `itemsToSave` and `itemFromSave` both handle SeedTimer (`serialize.go`)
- `json:"seed_timer,omitempty"` means old saves without the field unmarshal to 0

**Tests (TDD):**
- Round-trip: create world with plant having `SeedTimer = 50.0` → save → load → verify `SeedTimer == 50.0`
- Round-trip: create world with flower seeds in vessel → save → load → verify seeds preserved with correct variety (ItemType, Kind, Color, Plantable)
- Migration: load save data without `seed_timer` field → verify `SeedTimer == 0` (seeds available, no error)

**Architecture:** Follows serialization checklist in architecture.md. Uses existing round-trip test patterns in serialize_test.go or update_test.go.

**Values:** Anchor to Intent (tests validate "player's world state survives save/load"), Follow the Existing Shape (matches existing round-trip test patterns).

**Files:**
- `internal/ui/serialize_test.go` or `internal/ui/update_test.go` — round-trip and migration tests

[TEST] Save a world with active extraction (flowers with SeedTimers, seeds in vessels). Reload. Verify SeedTimers preserved, seeds in vessels intact, varieties correct. Load a pre-extraction-era save — verify no errors, SeedTimer defaults to 0.

[DOCS]

[RETRO]
