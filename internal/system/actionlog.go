package system

import (
	"fmt"
	"sync"
	"time"
)

// Event represents a single logged event
type Event struct {
	Timestamp time.Time
	CharID    int
	CharName  string
	Type      string
	Message   string
}

// ActionLog maintains a log of significant character events
type ActionLog struct {
	mu        sync.RWMutex
	logs      map[int][]Event
	maxEvents int
	startTime time.Time
}

// NewActionLog creates a new action log
func NewActionLog(maxEvents int) *ActionLog {
	return &ActionLog{
		logs:      make(map[int][]Event),
		maxEvents: maxEvents,
		startTime: time.Now(),
	}
}

// Add records an event for a character
func (al *ActionLog) Add(charID int, charName, eventType, message string) {
	al.mu.Lock()
	defer al.mu.Unlock()

	event := Event{
		Timestamp: time.Now(),
		CharID:    charID,
		CharName:  charName,
		Type:      eventType,
		Message:   message,
	}

	al.logs[charID] = append(al.logs[charID], event)

	// Trim if over limit
	if len(al.logs[charID]) > al.maxEvents {
		al.logs[charID] = al.logs[charID][len(al.logs[charID])-al.maxEvents:]
	}
}

// Events returns events for a character
func (al *ActionLog) Events(charID int, limit int) []Event {
	al.mu.RLock()
	defer al.mu.RUnlock()

	events := al.logs[charID]
	if limit > 0 && len(events) > limit {
		return events[len(events)-limit:]
	}

	// Return a copy to avoid race conditions
	result := make([]Event, len(events))
	copy(result, events)
	return result
}

// EventCount returns the number of events for a character
func (al *ActionLog) EventCount(charID int) int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return len(al.logs[charID])
}

// AllEvents returns all events from all characters, sorted by timestamp (oldest first)
func (al *ActionLog) AllEvents(limit int) []Event {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Collect all events
	var all []Event
	for _, events := range al.logs {
		all = append(all, events...)
	}

	// Sort by timestamp (oldest first)
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Timestamp.Before(all[i].Timestamp) {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	// Apply limit (return most recent)
	if limit > 0 && len(all) > limit {
		return all[len(all)-limit:]
	}

	return all
}

// AllEventCount returns total number of events across all characters
func (al *ActionLog) AllEventCount() int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	total := 0
	for _, events := range al.logs {
		total += len(events)
	}
	return total
}

// FormatElapsed formats elapsed time for display
func FormatElapsed(elapsed time.Duration) string {
	secs := int(elapsed.Seconds())
	if secs < 60 {
		return fmt.Sprintf("[%ds]", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("[%dm]", mins)
	}
	hours := mins / 60
	return fmt.Sprintf("[%dh]", hours)
}
