package entity

// IntentFormation describes how an activity is triggered
type IntentFormation string

const (
	// IntentAutomatic activities are triggered by needs or chosen as idle activity
	IntentAutomatic IntentFormation = "automatic"
	// IntentOrderable activities are triggered by user orders
	IntentOrderable IntentFormation = "orderable"
)

// Availability describes who can perform an activity
type Availability string

const (
	// AvailabilityDefault means all characters can perform the activity
	AvailabilityDefault Availability = "default"
	// AvailabilityKnowHow means the character must discover the activity first
	AvailabilityKnowHow Availability = "knowhow"
)

// DiscoveryTrigger defines when a know-how activity can be discovered
type DiscoveryTrigger struct {
	Action         ActionType // The action that can trigger discovery
	ItemType       string     // Specific item type required (empty = any)
	RequiresEdible bool       // Only trigger if item is edible
}

// Activity defines a type of activity characters can perform
type Activity struct {
	ID                string
	Name              string
	Category          string // Grouping for order UI (e.g., "craft", "garden"). Empty = uncategorized.
	IntentFormation   IntentFormation
	Availability      Availability
	DiscoveryTriggers []DiscoveryTrigger // nil for default activities
}

// ActivityRegistry contains all defined activities
var ActivityRegistry = map[string]Activity{
	"eat": {
		ID:              "eat",
		Name:            "Eat",
		IntentFormation: IntentAutomatic,
		Availability:    AvailabilityDefault,
	},
	"drink": {
		ID:              "drink",
		Name:            "Drink",
		IntentFormation: IntentAutomatic,
		Availability:    AvailabilityDefault,
	},
	"look": {
		ID:              "look",
		Name:            "Look",
		IntentFormation: IntentAutomatic,
		Availability:    AvailabilityDefault,
	},
	"talk": {
		ID:              "talk",
		Name:            "Talk",
		IntentFormation: IntentAutomatic,
		Availability:    AvailabilityDefault,
	},
	"forage": {
		ID:              "forage",
		Name:            "Forage",
		IntentFormation: IntentAutomatic,
		Availability:    AvailabilityDefault,
	},
	"harvest": {
		ID:              "harvest",
		Name:            "Harvest",
		IntentFormation: IntentOrderable,
		Availability:    AvailabilityKnowHow,
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionPickup, RequiresEdible: true},  // foraging
			{Action: ActionConsume, RequiresEdible: true}, // eating
			{Action: ActionLook, RequiresEdible: true},    // looking
		},
	},
	"craftVessel": {
		ID:              "craftVessel",
		Name:            "Vessel",
		Category:        "craft",
		IntentFormation: IntentOrderable,
		Availability:    AvailabilityKnowHow,
		// No DiscoveryTriggers - discovered via recipes
	},
	"craftHoe": {
		ID:              "craftHoe",
		Name:            "Hoe",
		Category:        "craft",
		IntentFormation: IntentOrderable,
		Availability:    AvailabilityKnowHow,
		// No DiscoveryTriggers - discovered via recipes
	},
	"tillSoil": {
		ID:              "tillSoil",
		Name:            "Till Soil",
		Category:        "garden",
		IntentFormation: IntentOrderable,
		Availability:    AvailabilityKnowHow,
		DiscoveryTriggers: []DiscoveryTrigger{
			{Action: ActionLook, ItemType: "hoe"},
			{Action: ActionPickup, ItemType: "hoe"},
		},
	},
}

// GetDiscoverableActivities returns all activities that require know-how
func GetDiscoverableActivities() []Activity {
	var activities []Activity
	for _, activity := range ActivityRegistry {
		if activity.Availability == AvailabilityKnowHow {
			activities = append(activities, activity)
		}
	}
	return activities
}
