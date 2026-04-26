package entity

import "testing"

func TestOrder_DisplayName_ConstructionBuildPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		activityID string
		expected   string
	}{
		{"buildFence", "Build fence"},
		{"buildHut", "Build hut"},
	}

	for _, tc := range tests {
		order := NewOrder(1, tc.activityID, "")
		got := order.DisplayName()
		if got != tc.expected {
			t.Errorf("DisplayName() for %s: got %q, want %q", tc.activityID, got, tc.expected)
		}
	}
}
