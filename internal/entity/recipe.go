package entity

import "petri/internal/config"

// RecipeInput defines an input requirement for a recipe
type RecipeInput struct {
	ItemType string // e.g., "gourd"
	Count    int    // How many required
}

// RecipeOutput defines what a recipe produces
type RecipeOutput struct {
	ItemType          string // broad category: "vessel", "hoe"
	Kind              string // recipe subtype: "hollow gourd", "shell hoe"
	ContainerCapacity int    // If > 0, item has ContainerData with this capacity
}

// Recipe defines how to craft an item
type Recipe struct {
	ID                string // e.g., "hollow-gourd"
	ActivityID        string // e.g., "craftVessel" - links recipe to activity
	Name              string // e.g., "Hollow Gourd"
	Inputs            []RecipeInput
	Output            RecipeOutput
	Duration          float64            // Craft time in game seconds
	Repeatable        bool               // When true, craft order loops until world-state completion condition is met (DD-19)
	DiscoveryTriggers []DiscoveryTrigger // Triggers for discovering this recipe
	BundledActivities []string           // Additional activities granted on recipe discovery
}

// RecipeRegistry contains all defined recipes
var RecipeRegistry = map[string]*Recipe{
	"thatch-fence": {
		ID:         "thatch-fence",
		ActivityID: "buildFence",
		Name:       "Thatch Fence",
		Inputs:     []RecipeInput{{ItemType: "grass", Count: 6}},
		Output:     RecipeOutput{ItemType: "fence"}, // display only; actual output is a construct
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "grass"},
			{Action: ActionPickup, ItemType: "grass"},
		},
	},
	"stick-fence": {
		ID:         "stick-fence",
		ActivityID: "buildFence",
		Name:       "Stick Fence",
		Inputs:     []RecipeInput{{ItemType: "stick", Count: 6}},
		Output:     RecipeOutput{ItemType: "fence"},
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "stick"},
			{Action: ActionPickup, ItemType: "stick"},
		},
	},
	"brick-fence": {
		ID:         "brick-fence",
		ActivityID: "buildFence",
		Name:       "Brick Fence",
		Inputs:     []RecipeInput{{ItemType: "brick", Count: 6}},
		Output:     RecipeOutput{ItemType: "fence"},
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "brick"},
			{Action: ActionPickup, ItemType: "brick"},
		},
	},
	"clay-brick": {
		ID:         "clay-brick",
		ActivityID: "craftBrick",
		Name:       "Clay Brick",
		Inputs:     []RecipeInput{{ItemType: "clay", Count: 1}},
		Output:     RecipeOutput{ItemType: "brick"},
		Duration:   config.ActionDurationLong,
		Repeatable: true,
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "clay"},
			{Action: ActionPickup, ItemType: "clay"},
			{Action: ActionDig, ItemType: "clay"},
		},
	},
	"hollow-gourd": {
		ID:         "hollow-gourd",
		ActivityID: "craftVessel",
		Name:       "Hollow Gourd",
		Inputs:     []RecipeInput{{ItemType: "gourd", Count: 1}},
		Output:     RecipeOutput{ItemType: "vessel", Kind: "hollow gourd", ContainerCapacity: 1},
		Duration:   config.ActionDurationLong,
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "gourd"},    // looking at gourd
			{Action: ActionPickup, ItemType: "gourd"},  // picking up gourd
			{Action: ActionConsume, ItemType: "gourd"}, // eating gourd
			{Action: ActionDrink},                      // drinking from spring (no item)
		},
	},
	"shell-hoe": {
		ID:         "shell-hoe",
		ActivityID: "craftHoe",
		Name:       "Shell Hoe",
		Inputs: []RecipeInput{
			{ItemType: "stick", Count: 1},
			{ItemType: "shell", Count: 1},
		},
		Output:   RecipeOutput{ItemType: "hoe", Kind: "shell hoe"},
		Duration: config.ActionDurationLong,
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "stick"},   // looking at stick
			{Action: ActionPickup, ItemType: "stick"}, // picking up stick
			{Action: ActionLook, ItemType: "shell"},   // looking at shell
			{Action: ActionPickup, ItemType: "shell"}, // picking up shell
		},
		BundledActivities: []string{"tillSoil"}, // inventing a hoe implies knowing how to till
	},
	"thatch-hut": {
		ID:         "thatch-hut",
		ActivityID: "buildHut",
		Name:       "Thatch Hut",
		Inputs:     []RecipeInput{{ItemType: "grass", Count: 12}},
		Output:     RecipeOutput{ItemType: "hut"}, // display only; actual output is a construct
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ConstructKind: "fence", ConstructMaterial: "grass"},
		},
	},
	"stick-hut": {
		ID:         "stick-hut",
		ActivityID: "buildHut",
		Name:       "Stick Hut",
		Inputs:     []RecipeInput{{ItemType: "stick", Count: 12}},
		Output:     RecipeOutput{ItemType: "hut"},
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ConstructKind: "fence", ConstructMaterial: "stick"},
		},
	},
	"brick-hut": {
		ID:         "brick-hut",
		ActivityID: "buildHut",
		Name:       "Brick Hut",
		Inputs:     []RecipeInput{{ItemType: "brick", Count: 12}},
		Output:     RecipeOutput{ItemType: "hut"},
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ConstructKind: "fence", ConstructMaterial: "brick"},
		},
	},
}

// GetRecipesForActivity returns all recipes that belong to a given activity
func GetRecipesForActivity(activityID string) []*Recipe {
	var result []*Recipe
	for _, recipe := range RecipeRegistry {
		if recipe.ActivityID == activityID {
			result = append(result, recipe)
		}
	}
	return result
}

// GetDiscoverableRecipes returns all recipes that have discovery triggers
func GetDiscoverableRecipes() []*Recipe {
	var result []*Recipe
	for _, recipe := range RecipeRegistry {
		if len(recipe.DiscoveryTriggers) > 0 {
			result = append(result, recipe)
		}
	}
	return result
}
