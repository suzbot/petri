package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/system"
	"petri/internal/types"
)

// gamePhase represents the current game state
type gamePhase int

const (
	phaseSelectMode gamePhase = iota
	phaseCharacterCreate
	phaseSelectFood
	phaseSelectColor
	phasePlaying
)

// viewMode represents the current view mode during gameplay
type viewMode int

const (
	viewModeSelect      viewMode = iota // Cursor/examine mode with details panel
	viewModeAllActivity                 // Combined activity log, no details panel
)

// TestConfig holds test mode settings
type TestConfig struct {
	NoFood  bool // Skip spawning food items
	NoWater bool // Skip spawning water sources
	NoBeds  bool // Skip spawning beds
	Debug   bool // Show debug info (action progress, etc.)
}

// Model is the main Bubble Tea model
type Model struct {
	phase         gamePhase
	multiCharMode bool // true = 4 characters, false = 1 character
	selectedFood  string
	selectedColor types.Color

	gameMap *game.Map

	cursorX, cursorY int
	following        *entity.Character
	paused           bool
	logScrollOffset  int
	actionLog        *system.ActionLog

	lastUpdate time.Time

	// Terminal size
	width, height int

	// Flash timer for status symbol cycling (0.5s intervals)
	flashTimer float64
	flashIndex int

	// View mode (select vs all-activity)
	viewMode viewMode

	// Character creation state
	creationState *CharacterCreationState

	// Test mode config
	testCfg TestConfig
}

// NewModel creates a new game model
func NewModel(testCfg TestConfig) Model {
	return Model{
		phase:     phaseSelectMode,
		actionLog: system.NewActionLog(200),
		width:     80,
		height:    40,
		paused:    true, // World starts paused
		testCfg:   testCfg,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// tickMsg is sent on each game tick
type tickMsg time.Time

// tickCmd returns a command that sends a tick message
func tickCmd() tea.Cmd {
	return tea.Tick(config.UpdateInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
