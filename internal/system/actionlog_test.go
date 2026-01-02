package system

import (
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Basic Operations Tests
// =============================================================================

func TestActionLog_AddAndRetrieve(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)
	log.Add(1, "Len", "hunger", "Getting hungry")

	events := log.Events(1, 100)
	if len(events) != 1 {
		t.Fatalf("Events() should return 1 event, got %d", len(events))
	}

	e := events[0]
	if e.CharID != 1 {
		t.Errorf("Event CharID should be 1, got %d", e.CharID)
	}
	if e.CharName != "Len" {
		t.Errorf("Event CharName should be 'Len', got '%s'", e.CharName)
	}
	if e.Type != "hunger" {
		t.Errorf("Event Type should be 'hunger', got '%s'", e.Type)
	}
	if e.Message != "Getting hungry" {
		t.Errorf("Event Message should be 'Getting hungry', got '%s'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("Event Timestamp should be set")
	}
}

func TestActionLog_EventsLimit(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)

	// Add 10 events
	for i := 0; i < 10; i++ {
		log.Add(1, "Len", "test", "Event")
	}

	// Request only 5
	events := log.Events(1, 5)
	if len(events) != 5 {
		t.Errorf("Events() with limit 5 should return 5, got %d", len(events))
	}
}

func TestActionLog_EventsReturnsMostRecent(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)

	// Add events with different messages
	for i := 1; i <= 10; i++ {
		log.Add(1, "Len", "test", string(rune('A'+i-1))) // A, B, C, ...
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Request 3 most recent
	events := log.Events(1, 3)
	if len(events) != 3 {
		t.Fatalf("Events() should return 3, got %d", len(events))
	}

	// Should be H, I, J (the last 3)
	expected := []string{"H", "I", "J"}
	for i, e := range events {
		if e.Message != expected[i] {
			t.Errorf("Event %d message should be '%s', got '%s'", i, expected[i], e.Message)
		}
	}
}

func TestActionLog_EventCount(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)

	for i := 0; i < 7; i++ {
		log.Add(1, "Len", "test", "Event")
	}

	count := log.EventCount(1)
	if count != 7 {
		t.Errorf("EventCount() should return 7, got %d", count)
	}
}

func TestActionLog_EventsForNonExistentCharacter(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)
	log.Add(1, "Len", "test", "Event for char 1")

	events := log.Events(99, 100)
	if len(events) != 0 {
		t.Errorf("Events() for non-existent character should return empty, got %d", len(events))
	}
}

func TestActionLog_EventCountForNonExistentCharacter(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)
	count := log.EventCount(99)
	if count != 0 {
		t.Errorf("EventCount() for non-existent character should return 0, got %d", count)
	}
}

// =============================================================================
// Combined Events Tests
// =============================================================================

func TestActionLog_AllEventsMergesCharacters(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)

	// Add events with controlled timing
	log.Add(1, "Len", "test", "Event A")
	time.Sleep(time.Millisecond)
	log.Add(2, "Macca", "test", "Event B")
	time.Sleep(time.Millisecond)
	log.Add(1, "Len", "test", "Event C")

	events := log.AllEvents(100)
	if len(events) != 3 {
		t.Fatalf("AllEvents() should return 3 events, got %d", len(events))
	}

	// Should be sorted by timestamp (oldest first)
	if events[0].Message != "Event A" {
		t.Errorf("First event should be 'Event A', got '%s'", events[0].Message)
	}
	if events[1].Message != "Event B" {
		t.Errorf("Second event should be 'Event B', got '%s'", events[1].Message)
	}
	if events[2].Message != "Event C" {
		t.Errorf("Third event should be 'Event C', got '%s'", events[2].Message)
	}
}

func TestActionLog_AllEventsRespectsLimit(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)

	// Add 20 events across multiple characters
	for i := 0; i < 10; i++ {
		log.Add(1, "Len", "test", "Len Event")
		log.Add(2, "Macca", "test", "Macca Event")
	}

	events := log.AllEvents(10)
	if len(events) != 10 {
		t.Errorf("AllEvents(10) should return 10 events, got %d", len(events))
	}
}

func TestActionLog_AllEventCount(t *testing.T) {
	t.Parallel()

	log := NewActionLog(100)

	log.Add(1, "Len", "test", "Event")
	log.Add(1, "Len", "test", "Event")
	log.Add(1, "Len", "test", "Event")
	log.Add(1, "Len", "test", "Event")
	log.Add(1, "Len", "test", "Event") // 5 for char 1
	log.Add(2, "Macca", "test", "Event")
	log.Add(2, "Macca", "test", "Event")
	log.Add(2, "Macca", "test", "Event") // 3 for char 2
	log.Add(3, "Hari", "test", "Event")
	log.Add(3, "Hari", "test", "Event")
	log.Add(3, "Hari", "test", "Event")
	log.Add(3, "Hari", "test", "Event") // 4 for char 3

	count := log.AllEventCount()
	if count != 12 {
		t.Errorf("AllEventCount() should return 12, got %d", count)
	}
}

// =============================================================================
// Event Trimming Tests
// =============================================================================

func TestActionLog_TrimsAtMaxEvents(t *testing.T) {
	t.Parallel()

	maxEvents := 10
	log := NewActionLog(maxEvents)

	// Add 15 events
	for i := 1; i <= 15; i++ {
		log.Add(1, "Len", "test", string(rune('A'+i-1)))
	}

	events := log.Events(1, 100)
	if len(events) != maxEvents {
		t.Errorf("Events() should be trimmed to %d, got %d", maxEvents, len(events))
	}
}

func TestActionLog_TrimsOldestFirst(t *testing.T) {
	t.Parallel()

	maxEvents := 10
	log := NewActionLog(maxEvents)

	// Add 15 events: A through O
	for i := 1; i <= 15; i++ {
		log.Add(1, "Len", "test", string(rune('A'+i-1)))
		time.Sleep(time.Millisecond)
	}

	events := log.Events(1, 100)

	// Should have F through O (events 6-15, i.e., the last 10)
	expected := []string{"F", "G", "H", "I", "J", "K", "L", "M", "N", "O"}
	for i, e := range events {
		if e.Message != expected[i] {
			t.Errorf("Event %d should be '%s', got '%s'", i, expected[i], e.Message)
		}
	}
}

// =============================================================================
// Thread Safety Tests
// =============================================================================

func TestActionLog_ConcurrentWrites(t *testing.T) {
	t.Parallel()

	log := NewActionLog(10000)
	var wg sync.WaitGroup

	// 10 goroutines, each adding 100 events
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(charID int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				log.Add(charID, "Char", "test", "Event")
			}
		}(g + 1)
	}

	wg.Wait()

	count := log.AllEventCount()
	if count != 1000 {
		t.Errorf("AllEventCount() after concurrent writes should be 1000, got %d", count)
	}
}

func TestActionLog_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	log := NewActionLog(1000)
	var wg sync.WaitGroup

	// Pre-populate with some events
	for i := 0; i < 100; i++ {
		log.Add(1, "Len", "test", "Initial Event")
	}

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			log.Add(1, "Len", "test", "New Event")
		}
	}()

	// Reader goroutines
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = log.Events(1, 50)
				_ = log.AllEvents(50)
				_ = log.EventCount(1)
				_ = log.AllEventCount()
			}
		}()
	}

	wg.Wait()

	// If we got here without panic or race detector complaints, test passes
	// Verify we have events
	if log.EventCount(1) == 0 {
		t.Error("Should have events after concurrent operations")
	}
}

// =============================================================================
// FormatElapsed Tests
// =============================================================================

func TestFormatElapsed_Seconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		elapsed  time.Duration
		expected string
	}{
		{0, "[0s]"},
		{30 * time.Second, "[30s]"},
		{59 * time.Second, "[59s]"},
	}

	for _, tt := range tests {
		got := FormatElapsed(tt.elapsed)
		if got != tt.expected {
			t.Errorf("FormatElapsed(%v) = '%s', want '%s'", tt.elapsed, got, tt.expected)
		}
	}
}

func TestFormatElapsed_Minutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		elapsed  time.Duration
		expected string
	}{
		{60 * time.Second, "[1m]"},
		{5 * time.Minute, "[5m]"},
		{59 * time.Minute, "[59m]"},
	}

	for _, tt := range tests {
		got := FormatElapsed(tt.elapsed)
		if got != tt.expected {
			t.Errorf("FormatElapsed(%v) = '%s', want '%s'", tt.elapsed, got, tt.expected)
		}
	}
}

func TestFormatElapsed_Hours(t *testing.T) {
	t.Parallel()

	tests := []struct {
		elapsed  time.Duration
		expected string
	}{
		{60 * time.Minute, "[1h]"},
		{2 * time.Hour, "[2h]"},
	}

	for _, tt := range tests {
		got := FormatElapsed(tt.elapsed)
		if got != tt.expected {
			t.Errorf("FormatElapsed(%v) = '%s', want '%s'", tt.elapsed, got, tt.expected)
		}
	}
}
