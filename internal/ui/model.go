package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/save"
	"petri/internal/system"
	"petri/internal/types"
)

// gamePhase represents the current game state
type gamePhase int

const (
	phaseWorldSelect gamePhase = iota // New: world selection screen
	phaseSelectMode
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
	NoFood        bool // Skip spawning food items
	NoWater       bool // Skip spawning water sources
	NoBeds        bool // Skip spawning beds
	NoCharacters  bool // Skip spawning characters (test mode)
	Debug         bool // Show debug info (action progress, etc.)
	MushroomsOnly bool // Replace all items with mushroom varieties
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

	lastUpdate      time.Time
	elapsedGameTime float64 // Total simulation time in seconds

	// Save state tracking
	worldID          string    // Current world ID for saving
	lastSaveGameTime float64   // Game time of last save (for periodic saves)
	saveIndicatorEnd time.Time // When to stop showing "Saving" indicator

	// Terminal size
	width, height int

	// Flash timer for status symbol cycling (0.5s intervals)
	flashTimer float64
	flashIndex int

	// View mode (select vs all-activity)
	viewMode           viewMode
	activityFullScreen bool // When true and in AllActivity mode, show full-screen log

	// Bottom panel toggles (mutually exclusive, replaces action log in select mode)
	showKnowledgePanel bool
	showInventoryPanel bool

	// Orders system
	orders      []*entity.Order
	nextOrderID int

	// Orders panel UI state
	showOrdersPanel       bool
	ordersFullScreen      bool
	ordersCancelMode      bool
	selectedOrderIndex    int
	ordersAddMode         bool
	ordersAddStep         int // 0 = select activity, 1 = select target
	selectedActivityIndex int
	selectedTargetIndex   int

	// Character creation state
	creationState *CharacterCreationState

	// World selection state
	worlds           []save.WorldMeta
	selectedWorld    int // Index into worlds slice, len(worlds) = "New World"
	confirmingDelete int // -1 = not confirming, otherwise index of world to delete

	// Test mode config
	testCfg TestConfig
}

// NewModel creates a new game model
func NewModel(testCfg TestConfig) Model {
	// Load existing worlds
	worlds, _ := save.ListWorlds()

	return Model{
		phase:            phaseWorldSelect,
		actionLog:        system.NewActionLog(200),
		width:            80,
		height:           40,
		paused:           true, // World starts paused
		testCfg:          testCfg,
		worlds:           worlds,
		selectedWorld:    0,  // First world or "New World" if empty
		confirmingDelete: -1, // Not confirming delete
		nextOrderID:      1,  // Start at 1 so ID 0 means "no order"
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
