package entity

import (
	"testing"

	"petri/internal/config"
	"petri/internal/types"
)

// TestNewBerry_Properties verifies NewBerry creates berry with correct properties
func TestNewBerry_Properties(t *testing.T) {
	t.Parallel()

	item := NewBerry(5, 10, types.ColorRed, false, false)

	pos := item.Pos()
	if pos.X != 5 || pos.Y != 10 {
		t.Errorf("NewBerry Pos(): got (%d, %d), want (5, 10)", pos.X, pos.Y)
	}
	if item.ItemType != "berry" {
		t.Errorf("NewBerry ItemType: got %q, want %q", item.ItemType, "berry")
	}
	if item.Color != types.ColorRed {
		t.Errorf("NewBerry Color: got %q, want %q", item.Color, types.ColorRed)
	}
	if item.IsPoisonous() {
		t.Error("NewBerry IsPoisonous: got true, want false")
	}
	if item.Symbol() != config.CharBerry {
		t.Errorf("NewBerry Symbol(): got %q, want %q", item.Symbol(), config.CharBerry)
	}
	if item.Type() != TypeItem {
		t.Errorf("NewBerry Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewBerry_Poisonous verifies NewBerry can create poisonous berry
func TestNewBerry_Poisonous(t *testing.T) {
	t.Parallel()

	item := NewBerry(0, 0, types.ColorWhite, true, false)
	if !item.IsPoisonous() {
		t.Error("NewBerry with poisonous=true: got IsPoisonous()=false")
	}
}

// TestNewMushroom_Properties verifies NewMushroom creates mushroom with correct properties
func TestNewMushroom_Properties(t *testing.T) {
	t.Parallel()

	item := NewMushroom(8, 12, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)

	pos := item.Pos()
	if pos.X != 8 || pos.Y != 12 {
		t.Errorf("NewMushroom Pos(): got (%d, %d), want (8, 12)", pos.X, pos.Y)
	}
	if item.ItemType != "mushroom" {
		t.Errorf("NewMushroom ItemType: got %q, want %q", item.ItemType, "mushroom")
	}
	if item.Color != types.ColorBrown {
		t.Errorf("NewMushroom Color: got %q, want %q", item.Color, types.ColorBrown)
	}
	if item.IsPoisonous() {
		t.Error("NewMushroom IsPoisonous: got true, want false")
	}
	if item.Symbol() != config.CharMushroom {
		t.Errorf("NewMushroom Symbol(): got %q, want %q", item.Symbol(), config.CharMushroom)
	}
	if item.Type() != TypeItem {
		t.Errorf("NewMushroom Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewMushroom_Poisonous verifies NewMushroom can create poisonous mushroom
func TestNewMushroom_Poisonous(t *testing.T) {
	t.Parallel()

	item := NewMushroom(0, 0, types.ColorBlue, types.PatternNone, types.TextureNone, true, false)
	if !item.IsPoisonous() {
		t.Error("NewMushroom with poisonous=true: got IsPoisonous()=false")
	}
}

// TestItem_Description_Berry verifies Description combines color and type for berry
func TestItem_Description_Berry(t *testing.T) {
	t.Parallel()

	item := NewBerry(0, 0, types.ColorRed, false, false)
	got := item.Description()
	if got != "red berry" {
		t.Errorf("Berry Description(): got %q, want %q", got, "red berry")
	}
}

// TestItem_Description_Mushroom verifies Description combines color and type for mushroom
func TestItem_Description_Mushroom(t *testing.T) {
	t.Parallel()

	item := NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	got := item.Description()
	if got != "brown mushroom" {
		t.Errorf("Mushroom Description(): got %q, want %q", got, "brown mushroom")
	}
}

// TestNewStick_Properties verifies NewStick creates stick with correct properties
func TestNewStick_Properties(t *testing.T) {
	t.Parallel()

	item := NewStick(3, 7)

	pos := item.Pos()
	if pos.X != 3 || pos.Y != 7 {
		t.Errorf("NewStick Pos(): got (%d, %d), want (3, 7)", pos.X, pos.Y)
	}
	if item.ItemType != "stick" {
		t.Errorf("NewStick ItemType: got %q, want %q", item.ItemType, "stick")
	}
	if item.Symbol() != config.CharStick {
		t.Errorf("NewStick Symbol(): got %c, want %c", item.Symbol(), config.CharStick)
	}
	if item.Color != types.ColorBrown {
		t.Errorf("NewStick Color: got %q, want %q", item.Color, types.ColorBrown)
	}
	if item.IsEdible() {
		t.Error("NewStick IsEdible: got true, want false")
	}
	if item.Plant != nil {
		t.Error("NewStick Plant: got non-nil, want nil")
	}
	if item.Type() != TypeItem {
		t.Errorf("NewStick Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewNut_Properties verifies NewNut creates nut with correct properties
func TestNewNut_Properties(t *testing.T) {
	t.Parallel()

	item := NewNut(4, 9)

	pos := item.Pos()
	if pos.X != 4 || pos.Y != 9 {
		t.Errorf("NewNut Pos(): got (%d, %d), want (4, 9)", pos.X, pos.Y)
	}
	if item.ItemType != "nut" {
		t.Errorf("NewNut ItemType: got %q, want %q", item.ItemType, "nut")
	}
	if item.Symbol() != config.CharNut {
		t.Errorf("NewNut Symbol(): got %c, want %c", item.Symbol(), config.CharNut)
	}
	if item.Color != types.ColorBrown {
		t.Errorf("NewNut Color: got %q, want %q", item.Color, types.ColorBrown)
	}
	if !item.IsEdible() {
		t.Error("NewNut IsEdible: got false, want true")
	}
	if item.IsPoisonous() {
		t.Error("NewNut IsPoisonous: got true, want false")
	}
	if item.IsHealing() {
		t.Error("NewNut IsHealing: got true, want false")
	}
	if item.Plant != nil {
		t.Error("NewNut Plant: got non-nil, want nil")
	}
}

// TestNewShell_Properties verifies NewShell creates shell with correct properties
func TestNewShell_Properties(t *testing.T) {
	t.Parallel()

	item := NewShell(2, 5, types.ColorLavender)

	pos := item.Pos()
	if pos.X != 2 || pos.Y != 5 {
		t.Errorf("NewShell Pos(): got (%d, %d), want (2, 5)", pos.X, pos.Y)
	}
	if item.ItemType != "shell" {
		t.Errorf("NewShell ItemType: got %q, want %q", item.ItemType, "shell")
	}
	if item.Symbol() != config.CharShell {
		t.Errorf("NewShell Symbol(): got %c, want %c", item.Symbol(), config.CharShell)
	}
	if item.Color != types.ColorLavender {
		t.Errorf("NewShell Color: got %q, want %q", item.Color, types.ColorLavender)
	}
	if item.IsEdible() {
		t.Error("NewShell IsEdible: got true, want false")
	}
	if item.Plant != nil {
		t.Error("NewShell Plant: got non-nil, want nil")
	}
}

// TestNewShell_ColorPreserved verifies shell color is set from argument
func TestNewShell_ColorPreserved(t *testing.T) {
	t.Parallel()

	colors := []types.Color{types.ColorWhite, types.ColorPalePink, types.ColorTan, types.ColorPaleYellow, types.ColorSilver, types.ColorGray, types.ColorLavender}
	for _, c := range colors {
		item := NewShell(0, 0, c)
		if item.Color != c {
			t.Errorf("NewShell Color: got %q, want %q", item.Color, c)
		}
	}
}

// TestItem_Description_Stick verifies stick description
func TestItem_Description_Stick(t *testing.T) {
	t.Parallel()

	item := NewStick(0, 0)
	got := item.Description()
	if got != "brown stick" {
		t.Errorf("Stick Description(): got %q, want %q", got, "brown stick")
	}
}

// TestItem_Description_Nut verifies nut description
func TestItem_Description_Nut(t *testing.T) {
	t.Parallel()

	item := NewNut(0, 0)
	got := item.Description()
	if got != "brown nut" {
		t.Errorf("Nut Description(): got %q, want %q", got, "brown nut")
	}
}

// TestItem_Description_Shell verifies shell description includes color
func TestItem_Description_Shell(t *testing.T) {
	t.Parallel()

	item := NewShell(0, 0, types.ColorPalePink)
	got := item.Description()
	if got != "pale pink shell" {
		t.Errorf("Shell Description(): got %q, want %q", got, "pale pink shell")
	}
}

// =============================================================================
// Kind field tests
// =============================================================================

// TestItem_Description_UsesKindWhenPresent verifies Description uses Kind over ItemType
func TestItem_Description_UsesKindWhenPresent(t *testing.T) {
	t.Parallel()

	item := &Item{
		ItemType: "hoe",
		Kind:     "shell hoe",
		Color:    types.ColorSilver,
	}
	got := item.Description()
	if got != "silver shell hoe" {
		t.Errorf("Description() with Kind: got %q, want %q", got, "silver shell hoe")
	}
}

// TestItem_Description_FallsBackToItemTypeWhenNoKind verifies Description uses ItemType when Kind is empty
func TestItem_Description_FallsBackToItemTypeWhenNoKind(t *testing.T) {
	t.Parallel()

	item := NewBerry(0, 0, types.ColorRed, false, false)
	got := item.Description()
	if got != "red berry" {
		t.Errorf("Description() without Kind: got %q, want %q", got, "red berry")
	}
}

// =============================================================================
// Hoe tests
// =============================================================================

// TestNewHoe_Properties verifies NewHoe creates hoe with correct properties
func TestNewHoe_Properties(t *testing.T) {
	t.Parallel()

	item := NewHoe(6, 11, types.ColorSilver)

	pos := item.Pos()
	if pos.X != 6 || pos.Y != 11 {
		t.Errorf("NewHoe Pos(): got (%d, %d), want (6, 11)", pos.X, pos.Y)
	}
	if item.ItemType != "hoe" {
		t.Errorf("NewHoe ItemType: got %q, want %q", item.ItemType, "hoe")
	}
	if item.Kind != "shell hoe" {
		t.Errorf("NewHoe Kind: got %q, want %q", item.Kind, "shell hoe")
	}
	if item.Symbol() != config.CharHoe {
		t.Errorf("NewHoe Symbol(): got %c, want %c", item.Symbol(), config.CharHoe)
	}
	if item.Color != types.ColorSilver {
		t.Errorf("NewHoe Color: got %q, want %q", item.Color, types.ColorSilver)
	}
	if item.IsEdible() {
		t.Error("NewHoe IsEdible: got true, want false")
	}
	if item.Plant != nil {
		t.Error("NewHoe Plant: got non-nil, want nil")
	}
	if item.Container != nil {
		t.Error("NewHoe Container: got non-nil, want nil")
	}
	if item.Type() != TypeItem {
		t.Errorf("NewHoe Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewHoe_ColorInherited verifies hoe color comes from parameter (shell color)
func TestNewHoe_ColorInherited(t *testing.T) {
	t.Parallel()

	colors := []types.Color{types.ColorWhite, types.ColorPalePink, types.ColorLavender, types.ColorSilver}
	for _, c := range colors {
		item := NewHoe(0, 0, c)
		if item.Color != c {
			t.Errorf("NewHoe Color: got %q, want %q", item.Color, c)
		}
	}
}

// TestNewHoe_Description verifies hoe description includes color and kind
func TestNewHoe_Description(t *testing.T) {
	t.Parallel()

	item := NewHoe(0, 0, types.ColorSilver)
	got := item.Description()
	if got != "silver shell hoe" {
		t.Errorf("Hoe Description(): got %q, want %q", got, "silver shell hoe")
	}
}

// =============================================================================
// Plantable field tests
// =============================================================================

// TestNewBerry_NotPlantableByDefault verifies berries start non-plantable (only becomes plantable when picked)
func TestNewBerry_NotPlantableByDefault(t *testing.T) {
	t.Parallel()
	item := NewBerry(0, 0, types.ColorRed, false, false)
	if item.Plantable {
		t.Error("NewBerry should have Plantable=false (only set on pickup)")
	}
}

// TestNewMushroom_NotPlantableByDefault verifies mushrooms start non-plantable
func TestNewMushroom_NotPlantableByDefault(t *testing.T) {
	t.Parallel()
	item := NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	if item.Plantable {
		t.Error("NewMushroom should have Plantable=false (only set on pickup)")
	}
}

// TestNewGourd_NotPlantable verifies gourds are never directly plantable
func TestNewGourd_NotPlantable(t *testing.T) {
	t.Parallel()
	item := NewGourd(0, 0, types.ColorGreen, types.PatternStriped, types.TextureWarty, false, false)
	if item.Plantable {
		t.Error("NewGourd should have Plantable=false (gourds produce seeds when eaten, not directly plantable)")
	}
}

// TestNewFlower_NotPlantable verifies flowers are not plantable
func TestNewFlower_NotPlantable(t *testing.T) {
	t.Parallel()
	item := NewFlower(0, 0, types.ColorBlue)
	if item.Plantable {
		t.Error("NewFlower should have Plantable=false")
	}
}

// =============================================================================
// Seed tests
// =============================================================================

// TestNewSeed_Properties verifies NewSeed creates seed with correct properties
func TestNewSeed_Properties(t *testing.T) {
	t.Parallel()

	item := NewSeed(3, 7, "gourd", types.ColorGreen, types.PatternSpotted, types.TextureWarty)

	pos := item.Pos()
	if pos.X != 3 || pos.Y != 7 {
		t.Errorf("NewSeed Pos(): got (%d, %d), want (3, 7)", pos.X, pos.Y)
	}
	if item.ItemType != "seed" {
		t.Errorf("NewSeed ItemType: got %q, want %q", item.ItemType, "seed")
	}
	if item.Kind != "gourd seed" {
		t.Errorf("NewSeed Kind: got %q, want %q", item.Kind, "gourd seed")
	}
	if item.Symbol() != config.CharSeed {
		t.Errorf("NewSeed Symbol(): got %c, want %c", item.Symbol(), config.CharSeed)
	}
	if item.Color != types.ColorGreen {
		t.Errorf("NewSeed Color: got %q, want %q", item.Color, types.ColorGreen)
	}
	if item.Pattern != types.PatternSpotted {
		t.Errorf("NewSeed Pattern: got %q, want %q", item.Pattern, types.PatternSpotted)
	}
	if item.Texture != types.TextureWarty {
		t.Errorf("NewSeed Texture: got %q, want %q", item.Texture, types.TextureWarty)
	}
	if item.IsEdible() {
		t.Error("NewSeed IsEdible: got true, want false")
	}
	if item.Plant != nil {
		t.Error("NewSeed Plant: got non-nil, want nil")
	}
	if !item.Plantable {
		t.Error("NewSeed Plantable: got false, want true")
	}
	if item.Type() != TypeItem {
		t.Errorf("NewSeed Type(): got %d, want %d", item.Type(), TypeItem)
	}
}

// TestNewSeed_Description verifies seed description combines parent attributes with kind
func TestNewSeed_Description(t *testing.T) {
	t.Parallel()

	item := NewSeed(0, 0, "gourd", types.ColorGreen, types.PatternSpotted, types.TextureWarty)
	got := item.Description()
	want := "warty spotted green gourd seed"
	if got != want {
		t.Errorf("Seed Description(): got %q, want %q", got, want)
	}
}

// TestNewSeed_DescriptionColorOnly verifies seed description for color-only parent
func TestNewSeed_DescriptionColorOnly(t *testing.T) {
	t.Parallel()

	item := NewSeed(0, 0, "gourd", types.ColorOrange, types.PatternNone, types.TextureNone)
	got := item.Description()
	want := "orange gourd seed"
	if got != want {
		t.Errorf("Seed Description(): got %q, want %q", got, want)
	}
}

// =============================================================================
// CreateSprout tests
// =============================================================================

// TestCreateSprout_FromGourdSeed verifies sprout from seed has parent type and attributes
func TestCreateSprout_FromGourdSeed(t *testing.T) {
	t.Parallel()

	seed := NewSeed(0, 0, "gourd", types.ColorGreen, types.PatternSpotted, types.TextureWarty)
	edible := &EdibleProperties{} // gourds are edible, not poison/healing

	sprout := CreateSprout(5, 10, seed, edible)

	if sprout.ItemType != "gourd" {
		t.Errorf("CreateSprout ItemType: got %q, want %q", sprout.ItemType, "gourd")
	}
	if sprout.Color != types.ColorGreen {
		t.Errorf("CreateSprout Color: got %q, want %q", sprout.Color, types.ColorGreen)
	}
	if sprout.Pattern != types.PatternSpotted {
		t.Errorf("CreateSprout Pattern: got %q, want %q", sprout.Pattern, types.PatternSpotted)
	}
	if sprout.Texture != types.TextureWarty {
		t.Errorf("CreateSprout Texture: got %q, want %q", sprout.Texture, types.TextureWarty)
	}
	if sprout.Symbol() != config.CharSprout {
		t.Errorf("CreateSprout Symbol: got %c, want %c", sprout.Symbol(), config.CharSprout)
	}
	if sprout.Plant == nil {
		t.Fatal("CreateSprout Plant: got nil")
	}
	if !sprout.Plant.IsSprout {
		t.Error("CreateSprout IsSprout: got false, want true")
	}
	if !sprout.Plant.IsGrowing {
		t.Error("CreateSprout IsGrowing: got false, want true")
	}
	if sprout.Plant.SproutTimer != config.GetSproutDuration("gourd") {
		t.Errorf("CreateSprout SproutTimer: got %f, want %f", sprout.Plant.SproutTimer, config.GetSproutDuration("gourd"))
	}
	pos := sprout.Pos()
	if pos.X != 5 || pos.Y != 10 {
		t.Errorf("CreateSprout Pos: got (%d, %d), want (5, 10)", pos.X, pos.Y)
	}
	if !sprout.IsEdible() {
		t.Error("CreateSprout from gourd seed should be edible")
	}
}

// TestCreateSprout_FromBerry verifies sprout from berry preserves type and edible properties
func TestCreateSprout_FromBerry(t *testing.T) {
	t.Parallel()

	berry := NewBerry(0, 0, types.ColorRed, false, false)
	berry.Plantable = true

	sprout := CreateSprout(3, 7, berry, berry.Edible)

	if sprout.ItemType != "berry" {
		t.Errorf("CreateSprout ItemType: got %q, want %q", sprout.ItemType, "berry")
	}
	if sprout.Color != types.ColorRed {
		t.Errorf("CreateSprout Color: got %q, want %q", sprout.Color, types.ColorRed)
	}
	if sprout.Symbol() != config.CharSprout {
		t.Errorf("CreateSprout Symbol: got %c, want %c", sprout.Symbol(), config.CharSprout)
	}
	if sprout.Plant == nil {
		t.Fatal("CreateSprout Plant: got nil")
	}
	if !sprout.Plant.IsSprout {
		t.Error("CreateSprout IsSprout: got false, want true")
	}
	if !sprout.Plant.IsGrowing {
		t.Error("CreateSprout IsGrowing: got false, want true")
	}
	if !sprout.IsEdible() {
		t.Error("CreateSprout from berry should be edible")
	}
}

// TestCreateSprout_FromMushroom_PreservesEdible verifies poisonous/healing survives
func TestCreateSprout_FromMushroom_PreservesEdible(t *testing.T) {
	t.Parallel()

	mush := NewMushroom(0, 0, types.ColorBlue, types.PatternSpotted, types.TextureSlimy, true, false)
	mush.Plantable = true

	sprout := CreateSprout(2, 4, mush, mush.Edible)

	if sprout.ItemType != "mushroom" {
		t.Errorf("CreateSprout ItemType: got %q, want %q", sprout.ItemType, "mushroom")
	}
	if !sprout.IsPoisonous() {
		t.Error("CreateSprout from poisonous mushroom should be poisonous")
	}
	if sprout.IsHealing() {
		t.Error("CreateSprout from non-healing mushroom should not be healing")
	}
	if sprout.Pattern != types.PatternSpotted {
		t.Errorf("CreateSprout Pattern: got %q, want %q", sprout.Pattern, types.PatternSpotted)
	}
	if sprout.Texture != types.TextureSlimy {
		t.Errorf("CreateSprout Texture: got %q, want %q", sprout.Texture, types.TextureSlimy)
	}
}

// TestItem_Description_KindWithMultipleAttributes verifies Description with Kind + pattern + texture + color
func TestItem_Description_KindWithMultipleAttributes(t *testing.T) {
	t.Parallel()

	item := &Item{
		ItemType: "vessel",
		Kind:     "hollow gourd",
		Color:    types.ColorGreen,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureWarty,
	}
	got := item.Description()
	if got != "warty spotted green hollow gourd" {
		t.Errorf("Description() with Kind + attrs: got %q, want %q", got, "warty spotted green hollow gourd")
	}
}
