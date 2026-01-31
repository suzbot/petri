package save

import (
	"encoding/json"
	"strings"
	"testing"

	"petri/internal/types"
)

func TestJSONFormat_BackwardsCompatibility(t *testing.T) {
	// Test CharacterSave JSON format - should be flat, not nested
	cs := CharacterSave{
		ID:       1,
		Name:     "Test",
		Position: types.Position{X: 10, Y: 15},
		Health:   100,
	}

	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("Failed to marshal CharacterSave: %v", err)
	}
	jsonStr := string(data)

	// Verify flat format (x and y at top level, not nested under "pos")
	if !strings.Contains(jsonStr, `"x":10`) && !strings.Contains(jsonStr, `"x": 10`) {
		t.Errorf("Expected flat x field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"y":15`) && !strings.Contains(jsonStr, `"y": 15`) {
		t.Errorf("Expected flat y field, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"pos"`) || strings.Contains(jsonStr, `"position"`) {
		t.Errorf("Position should not be nested, got: %s", jsonStr)
	}

	// Test ItemSave
	is := ItemSave{
		ID:       2,
		Position: types.Position{X: 5, Y: 7},
		ItemType: "mushroom",
	}

	data, err = json.Marshal(is)
	if err != nil {
		t.Fatalf("Failed to marshal ItemSave: %v", err)
	}
	jsonStr = string(data)

	if !strings.Contains(jsonStr, `"x":5`) && !strings.Contains(jsonStr, `"x": 5`) {
		t.Errorf("Expected flat x field in ItemSave, got: %s", jsonStr)
	}

	// Test FeatureSave
	fs := FeatureSave{
		ID:          3,
		Position:    types.Position{X: 8, Y: 9},
		FeatureType: 1,
		DrinkSource: true,
	}

	data, err = json.Marshal(fs)
	if err != nil {
		t.Fatalf("Failed to marshal FeatureSave: %v", err)
	}
	jsonStr = string(data)

	if !strings.Contains(jsonStr, `"x":8`) && !strings.Contains(jsonStr, `"x": 8`) {
		t.Errorf("Expected flat x field in FeatureSave, got: %s", jsonStr)
	}
}

func TestJSONFormat_LoadOldSave(t *testing.T) {
	// Simulate old save format with flat x/y fields
	oldCharFormat := `{"id":1,"name":"OldChar","x":20,"y":25,"health":50,"hunger":0,"thirst":0,"energy":0,"mood":0,"poisoned":false,"poison_timer":0,"is_dead":false,"is_sleeping":false,"at_bed":false,"is_frustrated":false,"frustration_timer":0,"failed_intent_count":0,"idle_cooldown":0,"last_looked_x":0,"last_looked_y":0,"has_last_looked":false,"talking_with_id":0,"talk_timer":0,"hunger_cooldown":0,"thirst_cooldown":0,"energy_cooldown":0,"action_progress":0,"speed_accumulator":0,"current_activity":"","preferences":null,"knowledge":null,"assigned_order_id":0}`

	var loaded CharacterSave
	if err := json.Unmarshal([]byte(oldCharFormat), &loaded); err != nil {
		t.Fatalf("Failed to load old format: %v", err)
	}

	if loaded.X != 20 {
		t.Errorf("Expected X=20, got %d", loaded.X)
	}
	if loaded.Y != 25 {
		t.Errorf("Expected Y=25, got %d", loaded.Y)
	}
	if loaded.Name != "OldChar" {
		t.Errorf("Expected name 'OldChar', got '%s'", loaded.Name)
	}

	// Test old item format
	oldItemFormat := `{"id":2,"x":5,"y":7,"item_type":"mushroom","color":"red","pattern":"","texture":"","edible":false,"poisonous":false,"healing":false,"death_timer":0}`

	var loadedItem ItemSave
	if err := json.Unmarshal([]byte(oldItemFormat), &loadedItem); err != nil {
		t.Fatalf("Failed to load old item format: %v", err)
	}

	if loadedItem.X != 5 {
		t.Errorf("Expected item X=5, got %d", loadedItem.X)
	}
	if loadedItem.Y != 7 {
		t.Errorf("Expected item Y=7, got %d", loadedItem.Y)
	}

	// Test old feature format
	oldFeatureFormat := `{"id":3,"x":8,"y":9,"feature_type":1,"drink_source":true,"bed":false,"passable":false}`

	var loadedFeature FeatureSave
	if err := json.Unmarshal([]byte(oldFeatureFormat), &loadedFeature); err != nil {
		t.Fatalf("Failed to load old feature format: %v", err)
	}

	if loadedFeature.X != 8 {
		t.Errorf("Expected feature X=8, got %d", loadedFeature.X)
	}
	if loadedFeature.Y != 9 {
		t.Errorf("Expected feature Y=9, got %d", loadedFeature.Y)
	}
}
