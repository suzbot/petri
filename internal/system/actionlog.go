package system

import (
	"fmt"
	"sort"
	"sync"
)

// Event represents a single logged event
type Event struct {
	GameTime float64 // Elapsed game time in seconds when event occurred
	CharID   int
	CharName string
	Type     string
	Message  string
}

// ActionLog maintains a log of significant character events
type ActionLog struct {
	mu          sync.RWMutex
	logs        map[int][]Event
	maxEvents   int
	currentTime float64 // Current game time, updated each tick
}

// NewActionLog creates a new action log
func NewActionLog(maxEvents int) *ActionLog {
	return &ActionLog{
		logs:      make(map[int][]Event),
		maxEvents: maxEvents,
	}
}

// SetGameTime updates the current game time (call once per tick)
func (al *ActionLog) SetGameTime(gameTime float64) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.currentTime = gameTime
}

// GameTime returns the current game time
func (al *ActionLog) GameTime() float64 {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return al.currentTime
}

// Add records an event for a character using the current game time
func (al *ActionLog) Add(charID int, charName, eventType, message string) {
	al.mu.Lock()
	defer al.mu.Unlock()

	event := Event{
		GameTime: al.currentTime,
		CharID:   charID,
		CharName: charName,
		Type:     eventType,
		Message:  message,
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

// AllEvents returns all events from all characters, sorted by game time (oldest first)
func (al *ActionLog) AllEvents(limit int) []Event {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Collect all events
	var all []Event
	for _, events := range al.logs {
		all = append(all, events...)
	}

	// Stable sort by game time, with CharID as tiebreaker for deterministic order
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].GameTime != all[j].GameTime {
			return all[i].GameTime < all[j].GameTime
		}
		return all[i].CharID < all[j].CharID
	})

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

// AllLogs returns a copy of all per-character event logs (for save/load)
func (al *ActionLog) AllLogs() map[int][]Event {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make(map[int][]Event)
	for charID, events := range al.logs {
		// Make a copy of each event slice
		eventsCopy := make([]Event, len(events))
		copy(eventsCopy, events)
		result[charID] = eventsCopy
	}
	return result
}

// SetAllLogs replaces all per-character event logs (for save/load)
func (al *ActionLog) SetAllLogs(logs map[int][]Event) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.logs = logs
}

// FormatGameTime formats game time in seconds for display
func FormatGameTime(gameTimeSecs float64) string {
	secs := int(gameTimeSecs)
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
