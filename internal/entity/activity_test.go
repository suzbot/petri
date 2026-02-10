package entity

import "testing"

// TestTillSoilActivity_Registered verifies tillSoil activity exists with correct properties
func TestTillSoilActivity_Registered(t *testing.T) {
	t.Parallel()

	activity, ok := ActivityRegistry["tillSoil"]
	if !ok {
		t.Fatal("tillSoil activity not found in ActivityRegistry")
	}

	if activity.Name != "Till Soil" {
		t.Errorf("tillSoil Name: got %q, want %q", activity.Name, "Till Soil")
	}
	if activity.IntentFormation != IntentOrderable {
		t.Errorf("tillSoil IntentFormation: got %q, want %q", activity.IntentFormation, IntentOrderable)
	}
	if activity.Availability != AvailabilityKnowHow {
		t.Errorf("tillSoil Availability: got %q, want %q", activity.Availability, AvailabilityKnowHow)
	}
	if activity.Category != "garden" {
		t.Errorf("tillSoil Category: got %q, want %q", activity.Category, "garden")
	}

	// Verify discovery triggers from hoe
	if len(activity.DiscoveryTriggers) == 0 {
		t.Fatal("tillSoil DiscoveryTriggers: got empty, want triggers for hoe")
	}
	hasLookHoe := false
	hasPickupHoe := false
	for _, trigger := range activity.DiscoveryTriggers {
		if trigger.Action == ActionLook && trigger.ItemType == "hoe" {
			hasLookHoe = true
		}
		if trigger.Action == ActionPickup && trigger.ItemType == "hoe" {
			hasPickupHoe = true
		}
	}
	if !hasLookHoe {
		t.Error("tillSoil DiscoveryTriggers: missing ActionLook hoe trigger")
	}
	if !hasPickupHoe {
		t.Error("tillSoil DiscoveryTriggers: missing ActionPickup hoe trigger")
	}
}

// TestCraftActivities_HaveCategory verifies existing craft activities have Category set
func TestCraftActivities_HaveCategory(t *testing.T) {
	t.Parallel()

	craftActivities := []string{"craftVessel", "craftHoe"}
	for _, id := range craftActivities {
		activity, ok := ActivityRegistry[id]
		if !ok {
			t.Errorf("activity %q not found in ActivityRegistry", id)
			continue
		}
		if activity.Category != "craft" {
			t.Errorf("activity %q Category: got %q, want %q", id, activity.Category, "craft")
		}
	}
}

