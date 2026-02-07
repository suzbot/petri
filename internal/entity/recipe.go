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
	ID                string             // e.g., "hollow-gourd"
	ActivityID        string             // e.g., "craftVessel" - links recipe to activity
	Name              string             // e.g., "Hollow Gourd"
	Inputs            []RecipeInput
	Output            RecipeOutput
	Duration          float64            // Craft time in game seconds
	DiscoveryTriggers []DiscoveryTrigger // Triggers for discovering this recipe
}

// RecipeRegistry contains all defined recipes
var RecipeRegistry = map[string]*Recipe{
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
