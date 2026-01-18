package system

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// Eating
// =============================================================================

func TestConsume_ReducesHunger(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 80

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	expected := 80 - config.FoodHungerReduction
	if char.Hunger != expected {
		t.Errorf("Hunger: got %.2f, want %.2f", char.Hunger, expected)
	}
}

func TestConsume_HungerFloorsAtZero(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 10 // Less than FoodHungerReduction

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.Hunger != 0 {
		t.Errorf("Hunger should floor at 0, got %.2f", char.Hunger)
	}
}

func TestConsume_SetsCooldownWhenFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 10 // Will reach 0 after eating

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.HungerCooldown != config.SatisfactionCooldown {
		t.Errorf("HungerCooldown: got %.2f, want %.2f", char.HungerCooldown, config.SatisfactionCooldown)
	}
}

func TestConsume_DoesNotSetCooldownWhenNotFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 80 // Will not reach 0 after eating

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.HungerCooldown != 0 {
		t.Errorf("HungerCooldown should be 0 when not fully satisfied, got %.2f", char.HungerCooldown)
	}
}

func TestConsume_PoisonousFoodAppliesPoison(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Poisoned = false

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, true, false) // Poisonous
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if !char.Poisoned {
		t.Error("Eating poisonous food should apply poison")
	}
	if char.PoisonTimer != config.PoisonDuration {
		t.Errorf("PoisonTimer: got %.2f, want %.2f", char.PoisonTimer, config.PoisonDuration)
	}
}

func TestConsume_NonPoisonousFoodDoesNotPoison(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Poisoned = false

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false) // Not poisonous
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.Poisoned {
		t.Error("Eating non-poisonous food should not apply poison")
	}
}

func TestConsume_RemovesItemFromMap(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	// Verify item exists
	if gameMap.ItemAt(5, 5) != item {
		t.Fatal("Item should exist on map before consumption")
	}

	Consume(char, item, gameMap, nil)

	// Verify item removed
	if gameMap.ItemAt(5, 5) != nil {
		t.Error("Item should be removed from map after consumption")
	}
}

func TestConsume_BoostsMoodWhenFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berry, likes red
	char.Mood = 50
	char.Hunger = 10 // Will reach 0 after eating

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Use brown mushroom - doesn't match berry or red preferences (NetPreference = 0)
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	expected := 50 + config.MoodBoostOnConsumption
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestConsume_NoMoodBoostWhenNotFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berry, likes red
	char.Mood = 50
	char.Hunger = 80 // Will not reach 0 after eating

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Use brown mushroom - doesn't match berry or red preferences (NetPreference = 0)
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.Mood != 50 {
		t.Errorf("Mood should remain 50 when not fully satisfied (and no preference match), got %.2f", char.Mood)
	}
}

func TestConsume_MoodCapsAt100(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 98   // Close to max
	char.Hunger = 10 // Will reach 0 after eating

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.Mood != 100 {
		t.Errorf("Mood should cap at 100, got %.2f", char.Mood)
	}
}

// =============================================================================
// Preference-based Mood Adjustment
// =============================================================================

func TestConsume_MoodBoostFromPositivePreference(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berry, likes red
	char.Mood = 50
	char.Hunger = 80 // Won't reach 0 (no satisfaction boost)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Red berry matches both preferences: NetPreference = +2
	item := entity.NewBerry(5, 5, types.ColorRed, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Mood should increase by NetPreference(+2) * MoodPreferenceModifier
	expected := 50 + 2*config.MoodPreferenceModifier
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestConsume_MoodBoostFromPartialPreference(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berry, likes red
	char.Mood = 50
	char.Hunger = 80 // Won't reach 0 (no satisfaction boost)

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Blue berry matches only berry preference: NetPreference = +1
	item := entity.NewBerry(5, 5, types.ColorBlue, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Mood should increase by NetPreference(+1) * MoodPreferenceModifier
	expected := 50 + 1*config.MoodPreferenceModifier
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestConsume_MoodPenaltyFromNegativePreference(t *testing.T) {
	t.Parallel()

	// Create character with a dislike
	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Mood: 50,
		Hunger: 80,
		Preferences: []entity.Preference{
			entity.NewNegativePreference("mushroom", ""), // Dislikes mushrooms
		},
	}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Mood should decrease by NetPreference(-1) * MoodPreferenceModifier
	expected := 50 - 1*config.MoodPreferenceModifier
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestConsume_MoodFloorsAtZero(t *testing.T) {
	t.Parallel()

	// Create character with strong dislike and low mood
	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Mood: 3, // Very low
		Hunger: 80,
		Preferences: []entity.Preference{
			entity.NewNegativePreference("mushroom", ""),
			entity.NewNegativePreference("", types.ColorBrown),
		},
	}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Brown mushroom matches both negative preferences: NetPreference = -2
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Mood should floor at 0, not go negative
	if char.Mood != 0 {
		t.Errorf("Mood should floor at 0, got %.2f", char.Mood)
	}
}

// =============================================================================
// Healing
// =============================================================================

func TestConsume_HealingItemRestoresHealth(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 60 // Damaged

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, true) // Healing item
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	expected := 60 + config.HealAmount
	if char.Health != expected {
		t.Errorf("Health: got %.2f, want %.2f", char.Health, expected)
	}
}

func TestConsume_HealingCapsAt100(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 95 // Near max

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, true) // Healing item
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.Health != 100 {
		t.Errorf("Health should cap at 100, got %.2f", char.Health)
	}
}

func TestConsume_HealingBoostsMoodWhenFullyHealed(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 90 // Will reach 100 after healing
	char.Mood = 50

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Use brown mushroom to avoid preference mood effects
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, true) // Healing item
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	expectedMood := 50 + config.MoodBoostOnConsumption
	if char.Mood != expectedMood {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expectedMood)
	}
}

func TestConsume_HealingNoMoodBoostIfNotFullyHealed(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 50 // Won't reach 100 after healing
	char.Mood = 50

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	// Use brown mushroom to avoid preference mood effects
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternNone, types.TextureNone, false, true) // Healing item
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Mood should be unchanged (no boost, no preference effects)
	if char.Mood != 50 {
		t.Errorf("Mood should remain 50 when not fully healed, got %.2f", char.Mood)
	}
}

func TestConsume_NonHealingItemDoesNotHeal(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 60

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false) // Not healing
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if char.Health != 60 {
		t.Errorf("Health should be unchanged for non-healing item, got %.2f", char.Health)
	}
}

// =============================================================================
// Drinking
// =============================================================================

func TestDrink_ReducesThirst(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 80

	Drink(char, nil)

	expected := 80 - config.DrinkThirstReduction
	if char.Thirst != expected {
		t.Errorf("Thirst: got %.2f, want %.2f", char.Thirst, expected)
	}
}

func TestDrink_ThirstFloorsAtZero(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 10 // Less than DrinkThirstReduction

	Drink(char, nil)

	if char.Thirst != 0 {
		t.Errorf("Thirst should floor at 0, got %.2f", char.Thirst)
	}
}

func TestDrink_SetsCooldownWhenFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 10 // Will reach 0 after drinking

	Drink(char, nil)

	if char.ThirstCooldown != config.SatisfactionCooldown {
		t.Errorf("ThirstCooldown: got %.2f, want %.2f", char.ThirstCooldown, config.SatisfactionCooldown)
	}
}

func TestDrink_DoesNotSetCooldownWhenNotFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Thirst = 80 // Will not reach 0 after drinking

	Drink(char, nil)

	if char.ThirstCooldown != 0 {
		t.Errorf("ThirstCooldown should be 0 when not fully satisfied, got %.2f", char.ThirstCooldown)
	}
}

func TestDrink_BoostsMoodWhenFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Thirst = 10 // Will reach 0 after drinking

	Drink(char, nil)

	expected := 50 + config.MoodBoostOnConsumption
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestDrink_NoMoodBoostWhenNotFullySatisfied(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 50
	char.Thirst = 80 // Will not reach 0 after drinking

	Drink(char, nil)

	if char.Mood != 50 {
		t.Errorf("Mood should remain 50 when not fully satisfied, got %.2f", char.Mood)
	}
}

func TestDrink_MoodCapsAt100(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 98   // Close to max
	char.Thirst = 10 // Will reach 0 after drinking

	Drink(char, nil)

	if char.Mood != 100 {
		t.Errorf("Mood should cap at 100, got %.2f", char.Mood)
	}
}

// =============================================================================
// Sleeping
// =============================================================================

func TestStartSleep_SetsStateinBed(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsSleeping = false
	char.AtBed = false

	StartSleep(char, true, nil)

	if !char.IsSleeping {
		t.Error("IsSleeping should be true")
	}
	if !char.AtBed {
		t.Error("AtBed should be true when sleeping in bed")
	}
}

func TestStartSleep_SetsStateOnGround(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.IsSleeping = false
	char.AtBed = false

	StartSleep(char, false, nil)

	if !char.IsSleeping {
		t.Error("IsSleeping should be true")
	}
	if char.AtBed {
		t.Error("AtBed should be false when sleeping on ground")
	}
}

// =============================================================================
// Knowledge Learning
// =============================================================================

func TestConsume_PoisonousItemTeachesKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{} // Start with no knowledge

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewMushroom(5, 5, types.ColorRed, types.PatternSpotted, types.TextureSlimy, true, false) // Poisonous
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Character should have learned that this mushroom type is poisonous
	if len(char.Knowledge) != 1 {
		t.Fatalf("Expected 1 knowledge entry, got %d", len(char.Knowledge))
	}

	k := char.Knowledge[0]
	if k.Category != entity.KnowledgePoisonous {
		t.Errorf("Expected category %s, got %s", entity.KnowledgePoisonous, k.Category)
	}
	if k.ItemType != "mushroom" {
		t.Errorf("Expected itemType mushroom, got %s", k.ItemType)
	}
	if k.Color != types.ColorRed {
		t.Errorf("Expected color red, got %s", k.Color)
	}
	if k.Pattern != types.PatternSpotted {
		t.Errorf("Expected pattern spotted, got %s", k.Pattern)
	}
	if k.Texture != types.TextureSlimy {
		t.Errorf("Expected texture slimy, got %s", k.Texture)
	}
}

func TestConsume_PoisonousItemDoesNotDuplicateKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	// Character already knows this mushroom is poisonous
	existingKnowledge := entity.Knowledge{
		Category: entity.KnowledgePoisonous,
		ItemType: "mushroom",
		Color:    types.ColorRed,
		Pattern:  types.PatternSpotted,
		Texture:  types.TextureSlimy,
	}
	char.Knowledge = []entity.Knowledge{existingKnowledge}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewMushroom(5, 5, types.ColorRed, types.PatternSpotted, types.TextureSlimy, true, false) // Same type
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Should still have only 1 knowledge entry
	if len(char.Knowledge) != 1 {
		t.Errorf("Expected 1 knowledge entry (no duplicate), got %d", len(char.Knowledge))
	}
}

func TestConsume_NonPoisonousItemDoesNotTeachPoisonKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false) // Not poisonous
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Should have no poison knowledge
	for _, k := range char.Knowledge {
		if k.Category == entity.KnowledgePoisonous {
			t.Error("Should not learn poison knowledge from non-poisonous item")
		}
	}
}

func TestConsume_HealingItemTeachesKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{}
	char.Health = 80 // Damaged so healing applies

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorBlue, false, true) // Healing
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Character should have learned that this berry type is healing
	if len(char.Knowledge) != 1 {
		t.Fatalf("Expected 1 knowledge entry, got %d", len(char.Knowledge))
	}

	k := char.Knowledge[0]
	if k.Category != entity.KnowledgeHealing {
		t.Errorf("Expected category %s, got %s", entity.KnowledgeHealing, k.Category)
	}
	if k.ItemType != "berry" {
		t.Errorf("Expected itemType berry, got %s", k.ItemType)
	}
	if k.Color != types.ColorBlue {
		t.Errorf("Expected color blue, got %s", k.Color)
	}
}

func TestConsume_HealingItemDoesNotDuplicateKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	existingKnowledge := entity.Knowledge{
		Category: entity.KnowledgeHealing,
		ItemType: "berry",
		Color:    types.ColorBlue,
	}
	char.Knowledge = []entity.Knowledge{existingKnowledge}
	char.Health = 80

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorBlue, false, true) // Same type
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	if len(char.Knowledge) != 1 {
		t.Errorf("Expected 1 knowledge entry (no duplicate), got %d", len(char.Knowledge))
	}
}

func TestConsume_NonHealingItemDoesNotTeachHealingKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewBerry(5, 5, types.ColorRed, false, false) // Not healing
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	for _, k := range char.Knowledge {
		if k.Category == entity.KnowledgeHealing {
			t.Error("Should not learn healing knowledge from non-healing item")
		}
	}
}

// =============================================================================
// Poison Knowledge Creates Dislike (Sub-Phase D)
// =============================================================================

func TestConsume_PoisonousItemCreatesDislike(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{}
	char.Preferences = []entity.Preference{}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternSpotted, types.TextureNone, true, false)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Should have 1 knowledge entry
	if len(char.Knowledge) != 1 {
		t.Fatalf("Expected 1 knowledge entry, got %d", len(char.Knowledge))
	}

	// Should have 1 dislike preference for the full variety
	if len(char.Preferences) != 1 {
		t.Fatalf("Expected 1 preference (dislike), got %d", len(char.Preferences))
	}

	pref := char.Preferences[0]
	if pref.Valence != -1 {
		t.Errorf("Expected dislike (valence -1), got %d", pref.Valence)
	}
	if pref.ItemType != "mushroom" {
		t.Errorf("Expected ItemType 'mushroom', got '%s'", pref.ItemType)
	}
	if pref.Color != types.ColorBrown {
		t.Errorf("Expected Color brown, got '%s'", pref.Color)
	}
	if pref.Pattern != types.PatternSpotted {
		t.Errorf("Expected Pattern spotted, got '%s'", pref.Pattern)
	}
}

func TestConsume_PoisonousItemRemovesExactMatchLike(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{}
	char.Mood = 40 // Low enough that preference boost (+20) keeps mood in Neutral tier

	// Character has an exact matching "like" for the same variety
	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternSpotted, types.TextureNone, true, false)
	existingLike := entity.NewFullPreferenceFromItem(item, 1)
	char.Preferences = []entity.Preference{existingLike}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Existing like should be removed (not replaced with dislike)
	if len(char.Preferences) != 0 {
		t.Errorf("Expected 0 preferences (like removed), got %d", len(char.Preferences))
	}
}

func TestConsume_PoisonousItemAlreadyKnown_NoDislikeChange(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()

	item := entity.NewMushroom(5, 5, types.ColorBrown, types.PatternSpotted, types.TextureNone, true, false)

	// Character already knows AND dislikes this item
	existingKnowledge := entity.NewKnowledgeFromItem(item, entity.KnowledgePoisonous)
	existingDislike := entity.NewFullPreferenceFromItem(item, -1)
	char.Knowledge = []entity.Knowledge{existingKnowledge}
	char.Preferences = []entity.Preference{existingDislike}

	gameMap := game.NewMap(config.MapWidth, config.MapHeight)
	gameMap.AddItem(item)

	Consume(char, item, gameMap, nil)

	// Knowledge should still be 1 (no duplicate)
	if len(char.Knowledge) != 1 {
		t.Errorf("Expected 1 knowledge entry (no duplicate), got %d", len(char.Knowledge))
	}

	// Preferences should still be 1 (no change since dislike already exists)
	if len(char.Preferences) != 1 {
		t.Errorf("Expected 1 preference (unchanged), got %d", len(char.Preferences))
	}
}

// =============================================================================
// Eating from Inventory (5.3)
// =============================================================================

func TestConsumeFromInventory_ClearsCarrying(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	item := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = item

	ConsumeFromInventory(char, item, nil)

	if char.Carrying != nil {
		t.Error("Carrying should be nil after consuming from inventory")
	}
}

func TestConsumeFromInventory_ReducesHunger(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Hunger = 80
	item := entity.NewBerry(0, 0, types.ColorRed, false, false)
	char.Carrying = item

	ConsumeFromInventory(char, item, nil)

	expected := 80 - config.FoodHungerReduction
	if char.Hunger != expected {
		t.Errorf("Hunger: got %.2f, want %.2f", char.Hunger, expected)
	}
}

func TestConsumeFromInventory_AppliesPoison(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Poisoned = false
	item := entity.NewBerry(0, 0, types.ColorRed, true, false) // Poisonous
	char.Carrying = item

	ConsumeFromInventory(char, item, nil)

	if !char.Poisoned {
		t.Error("Eating poisonous item from inventory should apply poison")
	}
}

func TestConsumeFromInventory_AppliesHealing(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Health = 60
	item := entity.NewBerry(0, 0, types.ColorRed, false, true) // Healing
	char.Carrying = item

	ConsumeFromInventory(char, item, nil)

	expected := 60 + config.HealAmount
	if char.Health != expected {
		t.Errorf("Health: got %.2f, want %.2f", char.Health, expected)
	}
}

func TestConsumeFromInventory_AppliesMoodFromPreference(t *testing.T) {
	t.Parallel()

	char := newTestCharacter() // Likes berry, likes red
	char.Mood = 50
	char.Hunger = 80 // Won't reach 0
	item := entity.NewBerry(0, 0, types.ColorRed, false, false) // +2 preference
	char.Carrying = item

	ConsumeFromInventory(char, item, nil)

	expected := 50 + 2*config.MoodPreferenceModifier
	if char.Mood != expected {
		t.Errorf("Mood: got %.2f, want %.2f", char.Mood, expected)
	}
}

func TestConsumeFromInventory_TriggersPreferenceFormation(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Mood = 95 // High mood - very likely to form preference
	char.Preferences = []entity.Preference{}
	item := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
	char.Carrying = item

	// Run multiple times to increase chance of preference formation
	// (preference formation is probabilistic based on mood)
	// At mood 95 (Joyful tier), formation chance is 10% per attempt
	// With 75 attempts: P(all fail) = 0.9^75 = 0.04%
	formed := false
	for i := 0; i < 75; i++ {
		testChar := newTestCharacter()
		testChar.Mood = 95
		testChar.Preferences = []entity.Preference{}
		testItem := entity.NewMushroom(0, 0, types.ColorBrown, types.PatternNone, types.TextureNone, false, false)
		testChar.Carrying = testItem

		ConsumeFromInventory(testChar, testItem, nil)

		if len(testChar.Preferences) > 0 {
			formed = true
			break
		}
	}

	if !formed {
		t.Error("ConsumeFromInventory should trigger preference formation (failed after 75 attempts)")
	}
}

func TestConsumeFromInventory_LearnsKnowledge(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	char.Knowledge = []entity.Knowledge{}
	item := entity.NewBerry(0, 0, types.ColorRed, true, false) // Poisonous
	char.Carrying = item

	ConsumeFromInventory(char, item, nil)

	if len(char.Knowledge) != 1 {
		t.Fatalf("Expected 1 knowledge entry, got %d", len(char.Knowledge))
	}
	if char.Knowledge[0].Category != entity.KnowledgePoisonous {
		t.Errorf("Expected poison knowledge, got %s", char.Knowledge[0].Category)
	}
}
