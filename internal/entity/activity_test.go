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

// TestPlantActivity_Registered verifies plant activity exists with correct properties
func TestPlantActivity_Registered(t *testing.T) {
	t.Parallel()

	activity, ok := ActivityRegistry["plant"]
	if !ok {
		t.Fatal("plant activity not found in ActivityRegistry")
	}

	if activity.Name != "Plant" {
		t.Errorf("plant Name: got %q, want %q", activity.Name, "Plant")
	}
	if activity.IntentFormation != IntentOrderable {
		t.Errorf("plant IntentFormation: got %q, want %q", activity.IntentFormation, IntentOrderable)
	}
	if activity.Availability != AvailabilityKnowHow {
		t.Errorf("plant Availability: got %q, want %q", activity.Availability, AvailabilityKnowHow)
	}
	if activity.Category != "garden" {
		t.Errorf("plant Category: got %q, want %q", activity.Category, "garden")
	}

	// Verify discovery triggers with RequiresPlantable
	if len(activity.DiscoveryTriggers) == 0 {
		t.Fatal("plant DiscoveryTriggers: got empty, want triggers with RequiresPlantable")
	}
	hasLookPlantable := false
	hasPickupPlantable := false
	for _, trigger := range activity.DiscoveryTriggers {
		if trigger.Action == ActionLook && trigger.RequiresPlantable {
			hasLookPlantable = true
		}
		if trigger.Action == ActionPickup && trigger.RequiresPlantable {
			hasPickupPlantable = true
		}
	}
	if !hasLookPlantable {
		t.Error("plant DiscoveryTriggers: missing ActionLook with RequiresPlantable trigger")
	}
	if !hasPickupPlantable {
		t.Error("plant DiscoveryTriggers: missing ActionPickup with RequiresPlantable trigger")
	}
}

// TestDigActivity_InRegistry verifies the dig activity exists with correct properties
func TestDigActivity_InRegistry(t *testing.T) {
	t.Parallel()

	activity, ok := ActivityRegistry["dig"]
	if !ok {
		t.Fatal("dig activity not found in ActivityRegistry")
	}

	if activity.IntentFormation != IntentOrderable {
		t.Errorf("dig IntentFormation: got %q, want %q", activity.IntentFormation, IntentOrderable)
	}
	if activity.Availability != AvailabilityKnowHow {
		t.Errorf("dig Availability: got %q, want %q", activity.Availability, AvailabilityKnowHow)
	}
	if activity.Category != "" {
		t.Errorf("dig Category: got %q, want empty (top-level)", activity.Category)
	}

	hasLookClay := false
	hasPickupClay := false
	for _, trigger := range activity.DiscoveryTriggers {
		if trigger.Action == ActionLook && trigger.ItemType == "clay" {
			hasLookClay = true
		}
		if trigger.Action == ActionPickup && trigger.ItemType == "clay" {
			hasPickupClay = true
		}
	}
	if !hasLookClay {
		t.Error("dig DiscoveryTriggers: missing ActionLook clay trigger")
	}
	if !hasPickupClay {
		t.Error("dig DiscoveryTriggers: missing ActionPickup clay trigger")
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
