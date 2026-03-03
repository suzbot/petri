package game

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/types"
)

func TestGenerateVarieties_CreatesExpectedCounts(t *testing.T) {
	registry := GenerateVarieties()

	// Check we have varieties for each item type
	berries := registry.VarietiesOfType("berry")
	mushrooms := registry.VarietiesOfType("mushroom")
	flowers := registry.VarietiesOfType("flower")

	if len(berries) < config.VarietyMinCount {
		t.Errorf("Expected at least %d berry varieties, got %d", config.VarietyMinCount, len(berries))
	}

	if len(mushrooms) < config.VarietyMinCount {
		t.Errorf("Expected at least %d mushroom varieties, got %d", config.VarietyMinCount, len(mushrooms))
	}

	if len(flowers) < config.VarietyMinCount {
		t.Errorf("Expected at least %d flower varieties, got %d", config.VarietyMinCount, len(flowers))
	}
}

func TestGenerateVarieties_MushroomsCanHavePatternOrTexture(t *testing.T) {
	// Test that mushroom config allows patterns and textures
	// (unlike berries which don't support them)
	configs := GetItemTypeConfigs()

	mushroomCfg, ok := configs["mushroom"]
	if !ok {
		t.Fatal("Expected mushroom config to exist")
	}

	if mushroomCfg.Patterns == nil || len(mushroomCfg.Patterns) == 0 {
		t.Error("Expected mushroom config to support patterns")
	}
	if mushroomCfg.Textures == nil || len(mushroomCfg.Textures) == 0 {
		t.Error("Expected mushroom config to support textures")
	}

	// Verify patterns include a non-None option
	hasNonNonePattern := false
	for _, p := range mushroomCfg.Patterns {
		if p != types.PatternNone {
			hasNonNonePattern = true
			break
		}
	}
	if !hasNonNonePattern {
		t.Error("Expected mushroom patterns to include at least one non-None pattern")
	}

	// Verify textures include a non-None option
	hasNonNoneTexture := false
	for _, tex := range mushroomCfg.Textures {
		if tex != types.TextureNone {
			hasNonNoneTexture = true
			break
		}
	}
	if !hasNonNoneTexture {
		t.Error("Expected mushroom textures to include at least one non-None texture")
	}
}

func TestGenerateVarieties_BerriesHaveNoPatternOrTexture(t *testing.T) {
	registry := GenerateVarieties()
	berries := registry.VarietiesOfType("berry")

	for _, b := range berries {
		if b.Pattern != "" {
			t.Errorf("Berry variety %q should not have pattern, got %q", b.ID, b.Pattern)
		}
		if b.Texture != "" {
			t.Errorf("Berry variety %q should not have texture, got %q", b.ID, b.Texture)
		}
	}
}

func TestGenerateVarieties_FlowersAreNotEdible(t *testing.T) {
	registry := GenerateVarieties()
	flowers := registry.VarietiesOfType("flower")

	for _, f := range flowers {
		if f.IsEdible() {
			t.Errorf("Flower variety %q should not be edible", f.ID)
		}
		if f.IsPoisonous() {
			t.Errorf("Flower variety %q should not be poisonous", f.ID)
		}
		if f.IsHealing() {
			t.Errorf("Flower variety %q should not be healing", f.ID)
		}
	}
}

func TestGenerateVarieties_PoisonAndHealingAssigned(t *testing.T) {
	registry := GenerateVarieties()
	edible := registry.EdibleVarieties()

	if len(edible) < 2 {
		t.Skip("Need at least 2 edible varieties to test poison/healing")
	}

	var poisonCount, healingCount int
	for _, v := range edible {
		if v.IsPoisonous() {
			poisonCount++
		}
		if v.IsHealing() {
			healingCount++
		}
		// Check no variety is both poisonous and healing
		if v.IsPoisonous() && v.IsHealing() {
			t.Errorf("Variety %q is both poisonous and healing", v.ID)
		}
	}

	if poisonCount == 0 {
		t.Error("Expected at least one poisonous variety")
	}
	if healingCount == 0 {
		t.Error("Expected at least one healing variety")
	}
}

func TestGenerateVarieties_UniqueIDs(t *testing.T) {
	registry := GenerateVarieties()
	all := registry.AllVarieties()

	seen := make(map[string]bool)
	for _, v := range all {
		if seen[v.ID] {
			t.Errorf("Duplicate variety ID: %q", v.ID)
		}
		seen[v.ID] = true
	}
}

func TestGenerateVarieties_SeedVarietiesForGourds(t *testing.T) {
	registry := GenerateVarieties()

	gourds := registry.VarietiesOfType("gourd")
	seeds := registry.VarietiesOfType("seed")

	if len(seeds) == 0 {
		t.Fatal("Expected seed varieties to be registered")
	}
	// Filter to gourd seeds only (flower and grass seeds also exist now)
	var gourdSeeds []*entity.ItemVariety
	for _, s := range seeds {
		if s.Kind == "gourd seed" {
			gourdSeeds = append(gourdSeeds, s)
		}
	}

	if len(gourdSeeds) != len(gourds) {
		t.Errorf("Expected %d gourd seed varieties (one per gourd), got %d", len(gourds), len(gourdSeeds))
	}

	// Each gourd variety should have a corresponding seed variety
	for _, g := range gourds {
		seedVariety := registry.GetByAttributes("seed", "gourd seed", g.Color, g.Pattern, g.Texture)
		if seedVariety == nil {
			t.Errorf("Expected seed variety for gourd %q, got nil", g.ID)
			continue
		}
		if seedVariety.ItemType != "seed" {
			t.Errorf("Seed variety ItemType: got %q, want %q", seedVariety.ItemType, "seed")
		}
		if seedVariety.IsEdible() {
			t.Errorf("Seed variety %q should not be edible", seedVariety.ID)
		}
	}
}

func TestGenerateVarieties_WaterVariety(t *testing.T) {
	registry := GenerateVarieties()

	liquids := registry.VarietiesOfType("liquid")
	if len(liquids) != 1 {
		t.Fatalf("Expected exactly 1 liquid variety, got %d", len(liquids))
	}

	water := liquids[0]
	if water.Kind != "water" {
		t.Errorf("Water variety Kind: got %q, want %q", water.Kind, "water")
	}
	if water.ItemType != "liquid" {
		t.Errorf("Water variety ItemType: got %q, want %q", water.ItemType, "liquid")
	}
	if water.Sym != 0 {
		t.Errorf("Water variety Sym: got %c, want 0 (never rendered as ground item)", water.Sym)
	}
	if water.IsEdible() {
		t.Error("Water variety should not be edible")
	}
	if water.Plantable {
		t.Error("Water variety should not be plantable")
	}
}

func TestGenerateVarieties_LiquidStackSize(t *testing.T) {
	size := config.GetStackSize("liquid")
	if size != 4 {
		t.Errorf("GetStackSize(\"liquid\"): got %d, want 4", size)
	}
}

func TestGenerateVarieties_NutVarietiesGenerated(t *testing.T) {
	registry := GenerateVarieties()

	nuts := registry.VarietiesOfType("nut")
	if len(nuts) == 0 {
		t.Fatal("Expected nut varieties to be registered, got 0")
	}

	for _, n := range nuts {
		if !n.IsEdible() {
			t.Errorf("Nut variety %q should be edible", n.ID)
		}
		if n.Sym != config.CharNut {
			t.Errorf("Nut variety %q has wrong symbol %c, want %c", n.ID, n.Sym, config.CharNut)
		}
	}
}

func TestShellStackSize(t *testing.T) {
	size := config.GetStackSize("shell")
	if size != 4 {
		t.Errorf("GetStackSize(\"shell\"): got %d, want 4", size)
	}
}

func TestGetGatherableTypes_ReturnsTypesOnGround(t *testing.T) {
	items := []*entity.Item{
		entity.NewNut(1, 1),
		entity.NewStick(2, 2),
		entity.NewShell(3, 3, types.ColorWhite),
	}

	entries := GetGatherableTypes(items)

	if len(entries) == 0 {
		t.Fatal("Expected gatherable entries, got 0")
	}

	// All returned types should be present in item list
	typeSet := map[string]bool{}
	for _, item := range items {
		typeSet[item.ItemType] = true
	}
	for _, e := range entries {
		if !typeSet[e.TargetType] {
			t.Errorf("GetGatherableTypes returned type %q not present in items", e.TargetType)
		}
	}
}

func TestGetGatherableTypes_ExcludesPlants(t *testing.T) {
	items := []*entity.Item{
		entity.NewNut(1, 1),
		entity.NewBerry(2, 2, types.ColorRed, false, false),
	}

	entries := GetGatherableTypes(items)

	for _, e := range entries {
		if e.TargetType == "berry" {
			t.Errorf("GetGatherableTypes should not include growing plants (berry)")
		}
	}
}

func TestGetGatherableTypes_Deduplicated(t *testing.T) {
	items := []*entity.Item{
		entity.NewNut(1, 1),
		entity.NewNut(2, 2),
		entity.NewNut(3, 3),
	}

	entries := GetGatherableTypes(items)

	if len(entries) != 1 {
		t.Errorf("GetGatherableTypes: got %d entries for 3 nuts, want 1 (deduplicated)", len(entries))
	}
}

func TestGetItemTypeConfigs_IncludesGrass(t *testing.T) {
	configs := GetItemTypeConfigs()

	grassCfg, ok := configs["grass"]
	if !ok {
		t.Fatal("Expected grass config to exist in GetItemTypeConfigs")
	}
	if grassCfg.Edible {
		t.Error("Grass should not be edible")
	}
	if grassCfg.Plantable {
		t.Error("Grass should not be plantable (grass seeds are, not grass material)")
	}
	if grassCfg.Sym != config.CharGrass {
		t.Errorf("Grass Sym: got %c, want %c", grassCfg.Sym, config.CharGrass)
	}
	if len(grassCfg.Colors) != 1 || grassCfg.Colors[0] != types.ColorPaleGreen {
		t.Errorf("Grass Colors: got %v, want [%s]", grassCfg.Colors, types.ColorPaleGreen)
	}
}

func TestGenerateVarieties_GrassVarietyRegistered(t *testing.T) {
	registry := GenerateVarieties()

	grasses := registry.VarietiesOfType("grass")
	if len(grasses) == 0 {
		t.Fatal("Expected grass varieties to be registered, got 0")
	}

	// Single variety: pale green, no pattern, no texture, Kind="tall grass"
	g := grasses[0]
	if g.Color != types.ColorPaleGreen {
		t.Errorf("Grass variety Color: got %q, want %q", g.Color, types.ColorPaleGreen)
	}
	if g.Pattern != types.PatternNone {
		t.Errorf("Grass variety Pattern: got %q, want %q", g.Pattern, types.PatternNone)
	}
	if g.Texture != types.TextureNone {
		t.Errorf("Grass variety Texture: got %q, want %q", g.Texture, types.TextureNone)
	}
	if g.Kind != "tall grass" {
		t.Errorf("Grass variety Kind: got %q, want %q", g.Kind, "tall grass")
	}
	if g.IsEdible() {
		t.Error("Grass variety should not be edible")
	}
	if g.Sym != config.CharGrass {
		t.Errorf("Grass variety Sym: got %c, want %c", g.Sym, config.CharGrass)
	}
}

func TestGenerateVarieties_CorrectSymbols(t *testing.T) {
	registry := GenerateVarieties()

	for _, v := range registry.VarietiesOfType("berry") {
		if v.Sym != config.CharBerry {
			t.Errorf("Berry variety %q has wrong symbol %c, want %c", v.ID, v.Sym, config.CharBerry)
		}
	}

	for _, v := range registry.VarietiesOfType("mushroom") {
		if v.Sym != config.CharMushroom {
			t.Errorf("Mushroom variety %q has wrong symbol %c, want %c", v.ID, v.Sym, config.CharMushroom)
		}
	}

	for _, v := range registry.VarietiesOfType("flower") {
		if v.Sym != config.CharFlower {
			t.Errorf("Flower variety %q has wrong symbol %c, want %c", v.ID, v.Sym, config.CharFlower)
		}
	}
}

func TestGenerateVarieties_FlowerSeedVarietiesRegistered(t *testing.T) {
	registry := GenerateVarieties()

	flowers := registry.VarietiesOfType("flower")
	seeds := registry.VarietiesOfType("seed")

	// Count flower seeds (Kind = "flower seed")
	var flowerSeeds []*entity.ItemVariety
	for _, s := range seeds {
		if s.Kind == "flower seed" {
			flowerSeeds = append(flowerSeeds, s)
		}
	}

	if len(flowerSeeds) != len(flowers) {
		t.Errorf("Expected %d flower seed varieties (one per flower), got %d", len(flowers), len(flowerSeeds))
	}

	// Each flower variety should have a corresponding seed variety with matching color
	for _, f := range flowers {
		seedVariety := registry.GetByAttributes("seed", "flower seed", f.Color, f.Pattern, f.Texture)
		if seedVariety == nil || seedVariety.Kind != "flower seed" {
			t.Errorf("Expected flower seed variety for flower color %q", f.Color)
			continue
		}
		if !seedVariety.Plantable {
			t.Errorf("Flower seed variety %q should be plantable", seedVariety.ID)
		}
		if seedVariety.Sym != config.CharSeed {
			t.Errorf("Flower seed variety Sym: got %c, want %c", seedVariety.Sym, config.CharSeed)
		}
		if seedVariety.IsEdible() {
			t.Errorf("Flower seed variety %q should not be edible", seedVariety.ID)
		}
	}
}

func TestGenerateVarieties_GrassSeedVarietyRegistered(t *testing.T) {
	registry := GenerateVarieties()

	grasses := registry.VarietiesOfType("grass")

	// Count grass seeds (Kind = "tall grass seed" — uses parent Kind)
	seeds := registry.VarietiesOfType("seed")
	var grassSeeds []*entity.ItemVariety
	for _, s := range seeds {
		if s.Kind == "tall grass seed" {
			grassSeeds = append(grassSeeds, s)
		}
	}

	if len(grassSeeds) != len(grasses) {
		t.Errorf("Expected %d grass seed varieties (one per grass), got %d", len(grasses), len(grassSeeds))
	}

	// Verify the seed has correct attributes and SourceVarietyID
	for _, g := range grasses {
		seedVariety := registry.GetByAttributes("seed", "tall grass seed", g.Color, g.Pattern, g.Texture)
		if seedVariety == nil || seedVariety.Kind != "tall grass seed" {
			t.Errorf("Expected tall grass seed variety for grass color %q", g.Color)
			continue
		}
		if !seedVariety.Plantable {
			t.Errorf("Grass seed variety %q should be plantable", seedVariety.ID)
		}
		if seedVariety.Sym != config.CharSeed {
			t.Errorf("Grass seed variety Sym: got %c, want %c", seedVariety.Sym, config.CharSeed)
		}
		if seedVariety.SourceVarietyID != g.ID {
			t.Errorf("Grass seed SourceVarietyID: got %q, want %q", seedVariety.SourceVarietyID, g.ID)
		}
	}
}

// TestGetHarvestableItemTypes_ReturnsGrowingNonSprout verifies that growing, non-sprout
// plants appear in the harvestable list.
func TestGetHarvestableItemTypes_ReturnsGrowingNonSprout(t *testing.T) {
	items := []*entity.Item{
		entity.NewBerry(1, 1, types.ColorRed, false, false),
		entity.NewGrass(2, 2),
	}

	types_ := GetHarvestableItemTypes(items)

	found := map[string]bool{}
	for _, tp := range types_ {
		found[tp] = true
	}
	if !found["berry"] {
		t.Error("Expected berry in harvestable types")
	}
	if !found["grass"] {
		t.Error("Expected grass in harvestable types")
	}
}

// TestGetHarvestableItemTypes_ExcludesSprouts verifies that sprouts are not harvestable.
func TestGetHarvestableItemTypes_ExcludesSprouts(t *testing.T) {
	sprout := entity.NewGrass(1, 1)
	sprout.Plant.IsSprout = true

	types_ := GetHarvestableItemTypes([]*entity.Item{sprout})

	for _, tp := range types_ {
		if tp == "grass" {
			t.Error("Sprouts should not appear in harvestable types")
		}
	}
}

// TestGetHarvestableItemTypes_ReturnsEmptyWhenNoPlants verifies empty result when no growing plants exist.
func TestGetHarvestableItemTypes_ReturnsEmptyWhenNoPlants(t *testing.T) {
	items := []*entity.Item{
		entity.NewStick(1, 1),
		entity.NewNut(2, 2),
	}

	types_ := GetHarvestableItemTypes(items)

	if len(types_) != 0 {
		t.Errorf("Expected empty harvestable types with no plants, got %v", types_)
	}
}

// TestGetHarvestableItemTypes_IncludesNonEdible verifies that non-edible growing plants
// (grass, flower) appear alongside edible ones (berry, mushroom).
func TestGetHarvestableItemTypes_IncludesNonEdible(t *testing.T) {
	items := []*entity.Item{
		entity.NewBerry(1, 1, types.ColorRed, false, false),
		entity.NewFlower(2, 2, types.ColorYellow),
		entity.NewGrass(3, 3),
	}

	types_ := GetHarvestableItemTypes(items)

	found := map[string]bool{}
	for _, tp := range types_ {
		found[tp] = true
	}
	if !found["berry"] {
		t.Error("Expected berry (edible plant) in harvestable types")
	}
	if !found["flower"] {
		t.Error("Expected flower (non-edible plant) in harvestable types")
	}
	if !found["grass"] {
		t.Error("Expected grass (non-edible plant) in harvestable types")
	}
}

// TestGetHarvestableItemTypes_Deduplicated verifies that multiple instances of the same
// plant type produce a single entry.
func TestGetHarvestableItemTypes_Deduplicated(t *testing.T) {
	items := []*entity.Item{
		entity.NewGrass(1, 1),
		entity.NewGrass(2, 2),
		entity.NewGrass(3, 3),
	}

	types_ := GetHarvestableItemTypes(items)

	count := 0
	for _, tp := range types_ {
		if tp == "grass" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 1 entry for 3 grass plants, got %d", count)
	}
}

// =============================================================================
// GetExtractableItemTypes
// =============================================================================

func TestGetExtractableItemTypes_ReturnsExtractablePlantTypes(t *testing.T) {
	t.Parallel()

	items := []*entity.Item{
		{
			BaseEntity: entity.BaseEntity{X: 1, Y: 1, Sym: config.CharFlower, EType: entity.TypeItem},
			ItemType:   "flower",
			Plant:      &entity.PlantProperties{IsGrowing: true},
		},
		{
			BaseEntity: entity.BaseEntity{X: 2, Y: 2, Sym: config.CharGrass, EType: entity.TypeItem},
			ItemType:   "grass",
			Plant:      &entity.PlantProperties{IsGrowing: true},
		},
	}

	result := GetExtractableItemTypes(items)
	if len(result) != 2 {
		t.Fatalf("Expected 2 extractable types, got %d", len(result))
	}
	// Sorted alphabetically by TargetType
	if result[0].TargetType != "flower" {
		t.Errorf("Expected first type 'flower', got %q", result[0].TargetType)
	}
	if result[1].TargetType != "grass" {
		t.Errorf("Expected second type 'grass', got %q", result[1].TargetType)
	}
}

func TestGetExtractableItemTypes_ExcludesNonExtractable(t *testing.T) {
	t.Parallel()

	items := []*entity.Item{
		// Berry is a growing plant but not extractable
		{
			BaseEntity: entity.BaseEntity{X: 1, Y: 1, Sym: config.CharBerry, EType: entity.TypeItem},
			ItemType:   "berry",
			Plant:      &entity.PlantProperties{IsGrowing: true},
		},
		// Flower is extractable
		{
			BaseEntity: entity.BaseEntity{X: 2, Y: 2, Sym: config.CharFlower, EType: entity.TypeItem},
			ItemType:   "flower",
			Plant:      &entity.PlantProperties{IsGrowing: true},
		},
	}

	result := GetExtractableItemTypes(items)
	if len(result) != 1 {
		t.Fatalf("Expected 1 extractable type, got %d", len(result))
	}
	if result[0].TargetType != "flower" {
		t.Errorf("Expected 'flower', got %q", result[0].TargetType)
	}
}

func TestGetExtractableItemTypes_ExcludesSprouts(t *testing.T) {
	t.Parallel()

	items := []*entity.Item{
		{
			BaseEntity: entity.BaseEntity{X: 1, Y: 1, Sym: config.CharSprout, EType: entity.TypeItem},
			ItemType:   "flower",
			Plant:      &entity.PlantProperties{IsGrowing: true, IsSprout: true},
		},
	}

	result := GetExtractableItemTypes(items)
	if len(result) != 0 {
		t.Errorf("Expected 0 extractable types for sprouts, got %d", len(result))
	}
}
