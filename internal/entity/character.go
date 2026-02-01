package entity

import (
	"petri/internal/config"
	"petri/internal/types"
)

// Character represents a game character with preferences and survival stats
type Character struct {
	BaseEntity
	ID          int
	Name        string
	Preferences      []Preference // Dynamic likes/dislikes for item attributes
	Knowledge        []Knowledge  // Things learned through experience (facts)
	KnownActivities  []string     // Activity IDs discovered (know-how)
	KnownRecipes     []string     // Recipe IDs learned (e.g., "hollow-gourd")

	// Survival attributes
	Health      float64
	Hunger      float64
	Thirst      float64
	Energy      float64
	Mood        float64 // 0-100, higher is better (Joyful at 90+, Miserable at 0-10)
	Poisoned    bool
	PoisonTimer float64
	IsDead      bool
	IsSleeping  bool
	AtBed       bool // true if sleeping at a leaf pile

	// Frustration tracking (when needs can't be met)
	IsFrustrated      bool
	FrustrationTimer  float64
	FailedIntentCount int

	// Idle activity tracking
	IdleCooldown  float64 // Time until next idle activity attempt
	LastLookedX   int     // Position of last item looked at (to avoid repetition)
	LastLookedY   int
	HasLastLooked bool // Whether LastLookedX/Y are valid

	// Talking activity tracking
	TalkingWith *Character // Current conversation partner (nil if not talking)
	TalkTimer   float64    // Time remaining in conversation

	// Satisfaction cooldowns (delay before stat starts changing after reaching optimal)
	HungerCooldown float64
	ThirstCooldown float64
	EnergyCooldown float64

	// Action progress (time spent on current action)
	ActionProgress float64

	// Speed tracking (accumulator for fractional movement)
	SpeedAccumulator float64

	// Activity tracking
	CurrentActivity string

	// Inventory (2 slots, each can hold a loose item or container)
	Inventory []*Item // Items being carried (capacity: InventoryCapacity)

	// Orders
	AssignedOrderID int // ID of currently assigned order (0 = none)

	// Intent-based movement for Phase II concurrency
	Intent *Intent
}

// Intent represents what a character wants to do next tick
type Intent struct {
	Target           types.Position // Next step position (immediate move)
	Dest             types.Position // Destination position (where we need to stand to interact)
	Action           ActionType
	TargetItem       *Item          // The specific item being pursued (nil if none)
	TargetFeature    *Feature       // The specific feature being pursued (nil if none)
	TargetCharacter  *Character     // The character being pursued for talking (nil if none)
	DrivingStat      types.StatType // Which stat is driving this intent
	DrivingTier      int            // The urgency tier when intent was set
}

// ActionType represents the type of action a character intends to take
type ActionType int

const (
	ActionNone ActionType = iota
	ActionMove
	ActionConsume
	ActionDrink
	ActionSleep
	ActionLook
	ActionTalk
	ActionPickup // Picking up an item (used by both foraging and harvest orders)
	ActionCraft  // Crafting an item (uses ActionProgress with Recipe.Duration)
)

// NewCharacter creates a new character with the given preferences
func NewCharacter(id, x, y int, name, food string, color types.Color) *Character {
	// Create initial preferences from food and color params
	preferences := []Preference{
		NewPositivePreference(food, ""),  // Likes this food type
		NewPositivePreference("", color), // Likes this color
	}

	return &Character{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharRobot,
			EType: TypeCharacter,
		},
		ID:              id,
		Name:            name,
		Preferences:     preferences,
		Health:          100,
		Hunger:          50,
		Thirst:          50,
		Energy:          100,
		Mood:            50, // Neutral mood
		CurrentActivity: "Idle",
	}
}

// NetPreference returns the total preference score for an item.
// Each matching preference contributes Valence × AttributeCount, so combo
// preferences (2 attributes) contribute twice as much as single-attribute ones.
// Positive = character likes the item, negative = dislikes, zero = neutral.
func (c *Character) NetPreference(item *Item) int {
	sum := 0
	for _, pref := range c.Preferences {
		sum += pref.MatchScore(item)
	}
	return sum
}

// NetPreferenceForVariety returns the total preference score for a variety.
// Used for checking preferences against vessel contents (which are Stacks of Varieties).
func (c *Character) NetPreferenceForVariety(v *ItemVariety) int {
	sum := 0
	for _, pref := range c.Preferences {
		sum += pref.MatchScoreVariety(v)
	}
	return sum
}

// HasKnowledge returns true if the character already has this knowledge
func (c *Character) HasKnowledge(k Knowledge) bool {
	for _, existing := range c.Knowledge {
		if existing.Equals(k) {
			return true
		}
	}
	return false
}

// LearnKnowledge adds knowledge if not already known.
// Returns true if new knowledge was learned, false if already known.
func (c *Character) LearnKnowledge(k Knowledge) bool {
	if c.HasKnowledge(k) {
		return false
	}
	c.Knowledge = append(c.Knowledge, k)
	return true
}

// KnownHealingItems returns items that the character knows are healing.
// Only items matching the character's healing knowledge are returned.
func (c *Character) KnownHealingItems(items []*Item) []*Item {
	var result []*Item
	for _, item := range items {
		for _, k := range c.Knowledge {
			if k.Category == KnowledgeHealing && k.Matches(item) {
				result = append(result, item)
				break // Don't add same item twice if multiple knowledge entries match
			}
		}
	}
	return result
}

// KnowsItemIsHealing returns true if the character has knowledge that this item is healing.
func (c *Character) KnowsItemIsHealing(item *Item) bool {
	for _, k := range c.Knowledge {
		if k.Category == KnowledgeHealing && k.Matches(item) {
			return true
		}
	}
	return false
}

// KnowsVarietyIsHealing returns true if the character has knowledge that this variety is healing.
// Used for checking knowledge against vessel contents (which are Stacks of Varieties).
func (c *Character) KnowsVarietyIsHealing(v *ItemVariety) bool {
	for _, k := range c.Knowledge {
		if k.Category == KnowledgeHealing && k.MatchesVariety(v) {
			return true
		}
	}
	return false
}

// KnowsActivity returns true if the character has discovered the specified activity.
func (c *Character) KnowsActivity(activityID string) bool {
	for _, id := range c.KnownActivities {
		if id == activityID {
			return true
		}
	}
	return false
}

// LearnActivity adds an activity to the character's known activities if not already known.
// Returns true if the activity was newly learned, false if already known.
func (c *Character) LearnActivity(activityID string) bool {
	if c.KnowsActivity(activityID) {
		return false
	}
	c.KnownActivities = append(c.KnownActivities, activityID)
	return true
}

// KnowsRecipe returns true if the character has learned the specified recipe.
func (c *Character) KnowsRecipe(recipeID string) bool {
	for _, id := range c.KnownRecipes {
		if id == recipeID {
			return true
		}
	}
	return false
}

// LearnRecipe adds a recipe to the character's known recipes if not already known.
// Returns true if the recipe was newly learned, false if already known.
func (c *Character) LearnRecipe(recipeID string) bool {
	if c.KnowsRecipe(recipeID) {
		return false
	}
	c.KnownRecipes = append(c.KnownRecipes, recipeID)
	return true
}

// GetKnownRecipesForActivity returns the character's known recipes for a given activity.
func (c *Character) GetKnownRecipesForActivity(activityID string) []*Recipe {
	var result []*Recipe
	for _, recipeID := range c.KnownRecipes {
		if recipe, ok := RecipeRegistry[recipeID]; ok {
			if recipe.ActivityID == activityID {
				result = append(result, recipe)
			}
		}
	}
	return result
}

// HungerLevel returns a human-readable hunger description
func (c *Character) HungerLevel() string {
	return hungerLevels.forTier(c.HungerTier())
}

// StatusText returns a human-readable status description
func (c *Character) StatusText() string {
	if c.IsDead {
		return "DEAD"
	}
	if c.IsSleeping {
		return "SLEEPING"
	}
	if c.Poisoned {
		return "POISONED"
	}
	return "Healthy"
}

// ThirstLevel returns a human-readable thirst description
func (c *Character) ThirstLevel() string {
	return thirstLevels.forTier(c.ThirstTier())
}

// EnergyLevel returns a human-readable energy description
func (c *Character) EnergyLevel() string {
	return energyLevels.forTier(c.EnergyTier())
}

// HealthLevel returns a human-readable health description
func (c *Character) HealthLevel() string {
	return healthLevels.forTier(c.HealthTier())
}

// EffectiveSpeed calculates current speed after all penalties
func (c *Character) EffectiveSpeed() int {
	speed := config.BaseSpeed

	// Poison penalty
	if c.Poisoned {
		speed -= config.PoisonSpeedPenalty
	}

	// Thirst penalties
	if c.Thirst >= 100 {
		speed -= config.DehydratedSpeedPenalty
	}
	if c.Thirst >= 90 {
		speed -= config.ParchedSpeedPenalty
	}

	// Energy penalties
	if c.Energy <= 10 {
		speed -= config.ExhaustedSpeedPenalty
	}
	if c.Energy <= 25 {
		speed -= config.VeryTiredSpeedPenalty
	}

	// Apply minimum floor
	if speed < config.MinSpeed {
		speed = config.MinSpeed
	}

	return speed
}

// Urgency tier constants
const (
	TierNone     = 0
	TierMild     = 1
	TierModerate = 2
	TierSevere   = 3
	TierCrisis   = 4
)

// StatThresholds defines tier boundaries for a stat
type StatThresholds struct {
	Mild     float64
	Moderate float64
	Severe   float64
	Crisis   float64
	Inverted bool // true if lower values are worse (energy, health)
}

// StatLevels defines human-readable descriptions for each tier
type StatLevels struct {
	None     string
	Mild     string
	Moderate string
	Severe   string
	Crisis   string
}

// Stat configurations
var (
	hungerThresholds = StatThresholds{50, 75, 90, 100, false}
	thirstThresholds = StatThresholds{50, 75, 90, 100, false}
	energyThresholds = StatThresholds{50, 25, 10, 0, true}
	healthThresholds = StatThresholds{75, 50, 25, 10, true}
	moodThresholds   = StatThresholds{89, 64, 34, 10, true} // Inverted: lower is worse

	hungerLevels = StatLevels{"Not Hungry", "Hungry", "Very Hungry", "Ravenous", "Starving"}
	thirstLevels = StatLevels{"Hydrated", "Thirsty", "Very Thirsty", "Parched", "Dehydrated"}
	energyLevels = StatLevels{"Rested", "Tired", "Very Tired", "Exhausted", "Collapsed"}
	healthLevels = StatLevels{"Healthy", "Poor", "Very Poor", "Critical", "Dying"}
	moodLevels   = StatLevels{"Joyful", "Happy", "Neutral", "Unhappy", "Miserable"}
)

// calculateTier returns the urgency tier for a value given thresholds
func calculateTier(value float64, t StatThresholds) int {
	if t.Inverted {
		switch {
		case value <= t.Crisis:
			return TierCrisis
		case value <= t.Severe:
			return TierSevere
		case value <= t.Moderate:
			return TierModerate
		case value <= t.Mild:
			return TierMild
		default:
			return TierNone
		}
	}
	switch {
	case value >= t.Crisis:
		return TierCrisis
	case value >= t.Severe:
		return TierSevere
	case value >= t.Moderate:
		return TierModerate
	case value >= t.Mild:
		return TierMild
	default:
		return TierNone
	}
}

// levelForTier returns the description for a given tier
func (l StatLevels) forTier(tier int) string {
	switch tier {
	case TierCrisis:
		return l.Crisis
	case TierSevere:
		return l.Severe
	case TierModerate:
		return l.Moderate
	case TierMild:
		return l.Mild
	default:
		return l.None
	}
}

// HungerTier returns the urgency tier for hunger
func (c *Character) HungerTier() int {
	return calculateTier(c.Hunger, hungerThresholds)
}

// ThirstTier returns the urgency tier for thirst
func (c *Character) ThirstTier() int {
	return calculateTier(c.Thirst, thirstThresholds)
}

// EnergyTier returns the urgency tier for energy
func (c *Character) EnergyTier() int {
	return calculateTier(c.Energy, energyThresholds)
}

// HealthTier returns the urgency tier for health
func (c *Character) HealthTier() int {
	return calculateTier(c.Health, healthThresholds)
}

// MoodTier returns the tier for mood (0=Joyful, 4=Miserable)
func (c *Character) MoodTier() int {
	return calculateTier(c.Mood, moodThresholds)
}

// MoodLevel returns human-readable mood description
func (c *Character) MoodLevel() string {
	return moodLevels.forTier(c.MoodTier())
}

// HungerUrgency returns normalized urgency score (0-100, higher = more urgent)
func (c *Character) HungerUrgency() float64 {
	return c.Hunger
}

// ThirstUrgency returns normalized urgency score (0-100, higher = more urgent)
func (c *Character) ThirstUrgency() float64 {
	return c.Thirst
}

// EnergyUrgency returns normalized urgency score (0-100, higher = more urgent)
func (c *Character) EnergyUrgency() float64 {
	return 100 - c.Energy
}

// IsInCrisis returns true if any stat is at crisis tier
func (c *Character) IsInCrisis() bool {
	return c.HungerTier() == TierCrisis ||
		c.ThirstTier() == TierCrisis ||
		c.EnergyTier() == TierCrisis ||
		c.HealthTier() == TierCrisis
}

// InventoryCapacity is the maximum number of items a character can carry
const InventoryCapacity = 2

// IsInventoryFull returns true if the character cannot carry more items
func (c *Character) IsInventoryFull() bool {
	return len(c.Inventory) >= InventoryCapacity
}

// HasInventorySpace returns true if character has at least one empty slot
func (c *Character) HasInventorySpace() bool {
	return len(c.Inventory) < InventoryCapacity
}

// AddToInventory adds item to first empty slot, returns false if full
func (c *Character) AddToInventory(item *Item) bool {
	if c.IsInventoryFull() {
		return false
	}
	c.Inventory = append(c.Inventory, item)
	return true
}

// RemoveFromInventory removes item from inventory, returns false if not found
func (c *Character) RemoveFromInventory(item *Item) bool {
	for i, inv := range c.Inventory {
		if inv == item {
			// Remove by replacing with last element and truncating
			c.Inventory[i] = c.Inventory[len(c.Inventory)-1]
			c.Inventory = c.Inventory[:len(c.Inventory)-1]
			return true
		}
	}
	return false
}

// GetCarriedVessel returns first vessel (item with Container) in inventory, or nil
func (c *Character) GetCarriedVessel() *Item {
	for _, item := range c.Inventory {
		if item != nil && item.Container != nil {
			return item
		}
	}
	return nil
}

// GetCarriedItem returns first non-vessel item in inventory, or nil
// Used for finding loose food items to eat
func (c *Character) GetCarriedItem() *Item {
	for _, item := range c.Inventory {
		if item != nil && item.Container == nil {
			return item
		}
	}
	return nil
}

// FindInInventory returns first item matching predicate, or nil
func (c *Character) FindInInventory(predicate func(*Item) bool) *Item {
	for _, item := range c.Inventory {
		if item != nil && predicate(item) {
			return item
		}
	}
	return nil
}

// HasAccessibleItem returns true if the character has an item of the given type
// either directly in inventory OR inside a carried vessel.
// Use this for checking availability without extraction.
func (c *Character) HasAccessibleItem(itemType string) bool {
	// Check direct inventory
	for _, item := range c.Inventory {
		if item != nil && item.ItemType == itemType {
			return true
		}
	}
	// Check vessel contents
	for _, item := range c.Inventory {
		if item != nil && item.Container != nil {
			for _, stack := range item.Container.Contents {
				if stack.Variety != nil && stack.Variety.ItemType == itemType && stack.Count > 0 {
					return true
				}
			}
		}
	}
	return false
}

// ConsumeAccessibleItem finds and removes one item of the given type from
// inventory or vessel contents. Returns the consumed item, or nil if not found.
// For items in vessels, creates a new Item from the variety.
func (c *Character) ConsumeAccessibleItem(itemType string) *Item {
	// First check direct inventory (prefer loose items)
	for _, item := range c.Inventory {
		if item != nil && item.ItemType == itemType {
			c.RemoveFromInventory(item)
			return item
		}
	}
	// Check vessel contents
	for _, vessel := range c.Inventory {
		if vessel == nil || vessel.Container == nil {
			continue
		}
		for i, stack := range vessel.Container.Contents {
			if stack.Variety != nil && stack.Variety.ItemType == itemType && stack.Count > 0 {
				// Decrement count
				vessel.Container.Contents[i].Count--
				// Remove empty stack
				if vessel.Container.Contents[i].Count <= 0 {
					vessel.Container.Contents = append(
						vessel.Container.Contents[:i],
						vessel.Container.Contents[i+1:]...,
					)
				}
				// Create item from variety
				item := &Item{
					ItemType: stack.Variety.ItemType,
					Color:    stack.Variety.Color,
					Pattern:  stack.Variety.Pattern,
					Texture:  stack.Variety.Texture,
				}
				item.EType = TypeItem
				return item
			}
		}
	}
	return nil
}
