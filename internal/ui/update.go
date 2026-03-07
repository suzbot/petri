package ui

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/save"
	"petri/internal/system"
	"petri/internal/types"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickMsg:
		if m.phase != phasePlaying || m.paused {
			return m, tickCmd(m.speedMultiplier)
		}
		newModel, _ := m.updateGame(time.Time(msg))
		return newModel, tickCmd(newModel.speedMultiplier)
	}

	return m, nil
}

// handleKey processes keyboard input
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseWorldSelect:
		return m.handleWorldSelectKey(msg)

	case phaseSelectMode:
		switch msg.String() {
		case "r", "R":
			return m.startGameRandom(), tickCmd(m.speedMultiplier)
		case "c", "C":
			m.creationState = NewCharacterCreationState()
			m.phase = phaseCharacterCreate
		case "esc":
			m.phase = phaseWorldSelect
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case phaseCharacterCreate:
		return m.handleCharacterCreationKey(msg)

	case phasePlaying:
		// Handle character name editing mode
		if m.editingCharacterName {
			return m.handleNameEditKey(msg)
		}

		switch msg.String() {
		case "ctrl+c":
			m.saveGame()
			return m, tea.Quit
		case "q":
			// Save and return to world select screen
			m.saveGame()
			m.phase = phaseWorldSelect
			// Reload world list to show updated metadata
			m.worlds, _ = save.ListWorlds()
			m.selectedWorld = 0
			m.confirmingDelete = -1
			m.gameMap = nil
			m.following = nil
			m.paused = true
			m.viewMode = viewModeSelect
			m.activityFullScreen = false
			m.showKnowledgePanel = false
			m.showInventoryPanel = false
			m.showPreferencesPanel = false
			m.showOrdersPanel = false
			m.logScrollOffset = 0
			// Clear world state so new worlds get fresh IDs and logs
			m.worldID = ""
			m.actionLog = system.NewActionLog(200)
			m.elapsedGameTime = 0
			m.orders = nil
			m.nextOrderID = 1
			return m, tea.Batch(tea.ClearScreen, tea.WindowSize())
		case "esc":
			// Collapse expanded views first
			if m.ordersFullScreen {
				m.ordersFullScreen = false
				return m, nil
			}
			if m.activityFullScreen {
				m.activityFullScreen = false
				return m, nil
			}
			// Orders add mode: back one level
			if m.showOrdersPanel && m.ordersAddMode {
				if m.ordersAddStep == 2 {
					if (m.step2ActivityID == "tillSoil" || m.step2ActivityID == "buildFence") && m.areaSelectAnchor != nil {
						// Clear anchor first, then back to step 1 on next esc
						m.areaSelectAnchor = nil
					} else {
						// Back to step 1
						m.ordersAddStep = 1
						m.selectedTargetIndex = 0
						m.areaSelectUnmarkMode = false
					}
				} else if m.ordersAddStep == 1 {
					// Step 1 (sub-menu): back to step 0
					m.ordersAddStep = 0
					m.selectedActivityIndex = 0
				} else {
					// Step 0: exit add mode
					m.ordersAddMode = false
				}
				return m, nil
			}
			// Orders cancel mode: exit
			if m.showOrdersPanel && m.ordersCancelMode {
				m.ordersCancelMode = false
				return m, nil
			}
			// Subpanel open: close it, show action log
			if m.showKnowledgePanel || m.showInventoryPanel || m.showPreferencesPanel {
				m.showKnowledgePanel = false
				m.showInventoryPanel = false
				m.showPreferencesPanel = false
				m.logScrollOffset = 0
				return m, nil
			}
			// Orders normal view: close panel, go to all-activity
			if m.showOrdersPanel {
				m.showOrdersPanel = false
				m.viewMode = viewModeAllActivity
				m.logScrollOffset = 0
				return m, nil
			}
			// Select view: go to all-activity
			if m.viewMode == viewModeSelect {
				m.viewMode = viewModeAllActivity
				m.logScrollOffset = 0
				return m, nil
			}
			// All-activity view: no-op
			return m, nil
		case " ":
			m.paused = !m.paused
			if m.paused {
				// Save when pausing
				m.saveGame()
			} else {
				// Reset lastUpdate when unpausing to prevent accumulated delta
				m.lastUpdate = time.Now()
			}
		case "f", "F":
			m.toggleFollow()
		case "a":
			// Switch to All Activity mode
			m.viewMode = viewModeAllActivity
			m.showOrdersPanel = false
			m.logScrollOffset = 0
		case "s":
			// Switch to Select mode
			m.viewMode = viewModeSelect
			m.showOrdersPanel = false
			m.activityFullScreen = false
			m.logScrollOffset = 0
		case "o", "O":
			// Toggle orders panel
			m.showOrdersPanel = !m.showOrdersPanel
			if m.showOrdersPanel {
				// Reset orders panel state
				m.ordersCancelMode = false
				m.ordersAddMode = false
				m.selectedOrderIndex = 0
			}
		case "x", "X":
			// Toggle full-screen for orders panel or activity log
			if m.showOrdersPanel {
				m.ordersFullScreen = !m.ordersFullScreen
			} else if m.viewMode == viewModeAllActivity {
				m.activityFullScreen = !m.activityFullScreen
				m.logScrollOffset = 0
			}
		case "<":
			// Slow down: 1 -> 2 -> 4
			if m.speedMultiplier < 4 {
				m.speedMultiplier *= 2
			}
		case ">":
			// Speed up: 4 -> 2 -> 1
			if m.speedMultiplier > 1 {
				m.speedMultiplier /= 2
			}
		case "+", "=":
			// Start add order mode (= is unshifted + on most keyboards)
			if m.showOrdersPanel && !m.ordersCancelMode && !m.ordersAddMode {
				orderableActivities := m.getOrderableActivities()
				if len(orderableActivities) > 0 {
					m.ordersAddMode = true
					m.ordersAddStep = 0
					m.selectedActivityIndex = 0
					m.selectedTargetIndex = 0
				}
			}
		case "c", "C":
			// Start cancel order mode
			if m.showOrdersPanel && !m.ordersAddMode && !m.ordersCancelMode {
				if len(m.orders) > 0 {
					m.ordersCancelMode = true
					m.selectedOrderIndex = 0
				}
			}
		case "tab":
			// Toggle mark/unmark mode during area selection (tillSoil and buildFence)
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 &&
				(m.step2ActivityID == "tillSoil" || m.step2ActivityID == "buildFence") {
				m.areaSelectUnmarkMode = !m.areaSelectUnmarkMode
				m.areaSelectAnchor = nil // Reset anchor when toggling mode
				return m, nil
			}
		case "enter":
			// Confirm selection in orders panel
			if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
				m.applyOrdersConfirm()
				return m, nil
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Number key: select-and-confirm in orders lists
			if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
				// tillSoil step 2 uses map cursor, not numbered list
				if m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID != "plant" {
					return m, nil
				}
				idx := int(msg.String()[0]-'0') - 1
				if m.ordersAddMode {
					switch m.ordersAddStep {
					case 0:
						m.selectedActivityIndex = idx
					case 1:
						m.selectedTargetIndex = idx
					case 2:
						m.selectedPlantTypeIndex = idx
					}
				} else if m.ordersCancelMode {
					m.selectedOrderIndex = idx
				}
				m.applyOrdersConfirm()
				return m, nil
			}
		case "k", "K":
			// Toggle knowledge panel (only in select mode, mutually exclusive with others)
			if m.viewMode == viewModeSelect {
				m.showKnowledgePanel = !m.showKnowledgePanel
				m.logScrollOffset = 0
				if m.showKnowledgePanel {
					m.showInventoryPanel = false
					m.showPreferencesPanel = false
				}
			}
		case "i", "I":
			// Toggle inventory panel (only in select mode, mutually exclusive with others)
			if m.viewMode == viewModeSelect {
				m.showInventoryPanel = !m.showInventoryPanel
				m.logScrollOffset = 0
				if m.showInventoryPanel {
					m.showKnowledgePanel = false
					m.showPreferencesPanel = false
				}
			}
		case "l", "L":
			// Return to action log from any details subpanel (select mode only)
			if m.viewMode == viewModeSelect && (m.showKnowledgePanel || m.showInventoryPanel || m.showPreferencesPanel) {
				m.showKnowledgePanel = false
				m.showInventoryPanel = false
				m.showPreferencesPanel = false
				m.logScrollOffset = 0
			}
		case "p":
			// Plot: anchor/confirm during area selection (tillSoil = rectangle, buildFence = line)
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID == "tillSoil" {
				if m.areaSelectAnchor == nil {
					anchor := types.Position{X: m.cursorX, Y: m.cursorY}
					m.areaSelectAnchor = &anchor
				} else {
					cursor := types.Position{X: m.cursorX, Y: m.cursorY}
					if m.areaSelectUnmarkMode {
						positions := getValidPositions(*m.areaSelectAnchor, cursor, m.gameMap, isValidUnmarkTarget)
						for _, pos := range positions {
							m.gameMap.UnmarkForTilling(pos)
						}
					} else {
						positions := getValidPositions(*m.areaSelectAnchor, cursor, m.gameMap, isValidTillTarget)
						for _, pos := range positions {
							m.gameMap.MarkForTilling(pos)
						}
					}
					m.areaSelectAnchor = nil // Clear anchor, stay in step 2
				}
				return m, nil
			}
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID == "buildFence" {
				if m.areaSelectAnchor == nil {
					anchor := types.Position{X: m.cursorX, Y: m.cursorY}
					m.areaSelectAnchor = &anchor
				} else {
					cursor := types.Position{X: m.cursorX, Y: m.cursorY}
					if m.areaSelectUnmarkMode {
						positions := getValidLinePositions(*m.areaSelectAnchor, cursor, m.gameMap, isValidUnmarkFenceTarget)
						for _, pos := range positions {
							m.gameMap.UnmarkForConstruction(pos)
						}
					} else {
						lineID := m.gameMap.NextConstructionLineID()
						positions := getValidLinePositions(*m.areaSelectAnchor, cursor, m.gameMap, isValidFenceTarget)
						for _, pos := range positions {
							m.gameMap.MarkForConstruction(pos, lineID)
						}
					}
					m.areaSelectAnchor = nil // Clear anchor, stay in step 2 for next line
				}
				return m, nil
			}
			// Toggle preferences panel (only in select mode, mutually exclusive with others)
			if m.viewMode == viewModeSelect {
				m.showPreferencesPanel = !m.showPreferencesPanel
				m.logScrollOffset = 0
				if m.showPreferencesPanel {
					m.showKnowledgePanel = false
					m.showInventoryPanel = false
				}
			}
		case "n":
			// Cycle to next alive character
			m.cycleToNextCharacter()
		case "b", "B":
			// Cycle to previous alive character
			m.cycleToPreviousCharacter()
		case "e", "E":
			// Enter character name edit mode (only in select mode with details panel)
			if m.viewMode == viewModeSelect {
				if char := m.characterAtCursor(); char != nil {
					m.editingCharacterName = true
					m.editingCharacterID = char.ID
					m.editingNameBuffer = char.Name
				}
			}
		case "up":
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 {
				if m.step2ActivityID == "plant" {
					if m.selectedPlantTypeIndex > 0 {
						m.selectedPlantTypeIndex--
					}
				} else {
					// Area selection: move cursor on map
					m.moveCursor(0, -1)
				}
			} else if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
				// Handle orders panel navigation inline
				if m.ordersAddMode {
					if m.ordersAddStep == 0 {
						if m.selectedActivityIndex > 0 {
							m.selectedActivityIndex--
						}
					} else {
						if m.selectedTargetIndex > 0 {
							m.selectedTargetIndex--
						}
					}
				} else if m.ordersCancelMode {
					if m.selectedOrderIndex > 0 {
						m.selectedOrderIndex--
					}
				}
			} else {
				m.moveCursor(0, -1)
			}
		case "down":
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 {
				if m.step2ActivityID == "plant" {
					plantTypes := game.GetPlantableTypes(m.gameMap.Items(), m.gameMap.Characters())
					if m.selectedPlantTypeIndex < len(plantTypes)-1 {
						m.selectedPlantTypeIndex++
					}
				} else {
					// Area selection: move cursor on map
					m.moveCursor(0, 1)
				}
			} else if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
				// Handle orders panel navigation inline
				if m.ordersAddMode {
					if m.ordersAddStep == 0 {
						activities := m.getOrderableActivities()
						if m.selectedActivityIndex < len(activities)-1 {
							m.selectedActivityIndex++
						}
					} else {
						// Step 1: determine list length based on selected activity
						activities := m.getOrderableActivities()
						var maxIndex int
						if m.selectedActivityIndex < len(activities) &&
							isSyntheticCategory(activities[m.selectedActivityIndex].ID) {
							category := syntheticCategoryID(activities[m.selectedActivityIndex].ID)
							maxIndex = len(m.getCategoryActivities(category)) - 1
						} else if m.selectedActivityIndex < len(activities) &&
							activities[m.selectedActivityIndex].ID == "gather" {
							maxIndex = len(game.GetGatherableTypes(m.gameMap.Items())) - 1
						} else if m.selectedActivityIndex < len(activities) &&
							activities[m.selectedActivityIndex].ID == "extract" {
							maxIndex = len(game.GetExtractableItemTypes(m.gameMap.Items())) - 1
						} else {
							maxIndex = len(m.getHarvestableItemTypes()) - 1
						}
						if m.selectedTargetIndex < maxIndex {
							m.selectedTargetIndex++
						}
					}
				} else if m.ordersCancelMode {
					if m.selectedOrderIndex < len(m.orders)-1 {
						m.selectedOrderIndex++
					}
				}
			} else {
				m.moveCursor(0, 1)
			}
		case "left":
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID != "plant" {
				m.moveCursor(-1, 0) // Area selection: move cursor on map
			} else if !(m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode)) {
				m.moveCursor(-1, 0)
			}
		case "right":
			if m.showOrdersPanel && m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID != "plant" {
				m.moveCursor(1, 0) // Area selection: move cursor on map
			} else if !(m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode)) {
				m.moveCursor(1, 0)
			}
		case "pgup":
			m.logScrollOffset += 5
		case "pgdown":
			if m.logScrollOffset >= 5 {
				m.logScrollOffset -= 5
			} else {
				m.logScrollOffset = 0
			}
		case "home":
			// Jump to oldest
			if e := m.gameMap.EntityAt(types.Position{X: m.cursorX, Y: m.cursorY}); e != nil {
				if char, ok := e.(*entity.Character); ok {
					m.logScrollOffset = m.actionLog.EventCount(char.ID) - 1
					if m.logScrollOffset < 0 {
						m.logScrollOffset = 0
					}
				}
			}
		case "end":
			m.logScrollOffset = 0
		case ".":
			// Step forward by one tick while paused
			if m.paused {
				m.stepForward()
			}
		}
	}

	return m, nil
}

// startGameRandom initializes the game world with 4 random characters
func (m Model) startGameRandom() Model {
	m.gameMap = game.NewMap(config.MapWidth, config.MapHeight)
	m.phase = phasePlaying
	m.lastUpdate = time.Now()

	// Center position
	cx, cy := config.MapWidth/2, config.MapHeight/2
	m.cursorX, m.cursorY = cx, cy

	// Spawn characters unless disabled
	if !m.testCfg.NoCharacters {
		names := randomUniqueNames(4)
		foods := getEdibleItemTypes()
		colors := types.AllColors
		offsets := [][2]int{{0, 0}, {2, 0}, {0, 2}, {2, 2}}

		var chars []*entity.Character
		for i, name := range names {
			x := cx + offsets[i][0]
			y := cy + offsets[i][1]
			food := foods[rand.Intn(len(foods))]
			color := colors[rand.Intn(len(colors))]
			char := entity.NewCharacter(i+1, x, y, name, food, color)
			m.gameMap.AddCharacter(char)
			chars = append(chars, char)
		}

		// Randomly select one character to follow
		followIdx := rand.Intn(len(chars))
		m.following = chars[followIdx]
		pos := chars[followIdx].Pos()
		m.cursorX, m.cursorY = pos.X, pos.Y
	}

	// Spawn world: ponds first (before items/features), then clay, then features, then items
	if !m.testCfg.NoWater {
		game.SpawnPonds(m.gameMap)
		game.SpawnClay(m.gameMap)
	}
	game.SpawnFeatures(m.gameMap, m.testCfg.NoWater, m.testCfg.NoBeds)
	if !m.testCfg.NoFood {
		game.SpawnItems(m.gameMap, m.testCfg.MushroomsOnly)
	}
	game.SpawnGroundItems(m.gameMap)
	m.groundSpawnTimers = system.GroundSpawnTimers{
		Stick: system.RandomGroundSpawnInterval(),
		Nut:   system.RandomGroundSpawnInterval(),
		Shell: system.RandomGroundSpawnInterval(),
	}

	// Create world for saving
	if m.worldID == "" {
		worldID, err := save.CreateWorld()
		if err == nil {
			m.worldID = worldID
		}
	}

	return m
}

// updateGame processes one game tick
func (m Model) updateGame(now time.Time) (Model, tea.Cmd) {
	delta := now.Sub(m.lastUpdate).Seconds()
	m.lastUpdate = now
	m.elapsedGameTime += delta

	// Warn if delta is suspiciously small (time not advancing)
	if delta <= 0.001 {
		save.LogWarning("Tick delta very small (%.6fs) - time may not be advancing properly. lastUpdate=%v, now=%v, elapsedGameTime=%.2f",
			delta, m.lastUpdate, now, m.elapsedGameTime)
	}

	// Update action log with current game time
	m.actionLog.SetGameTime(m.elapsedGameTime)

	// Periodic auto-save check
	if m.elapsedGameTime-m.lastSaveGameTime >= config.AutoSaveInterval {
		m.saveGame()
	}

	// Update flash timer for status symbol cycling (0.5s intervals)
	m.flashTimer += delta
	if m.flashTimer >= 0.5 {
		m.flashTimer = 0
		m.flashIndex++
	}

	// Update survival mechanics for all characters
	for _, char := range m.gameMap.Characters() {
		system.UpdateSurvival(char, delta, m.actionLog)
	}

	// Update item spawning (unless no food mode)
	if !m.testCfg.NoFood {
		initialItemCount := config.ItemSpawnCount*2 + config.FlowerSpawnCount // berries + mushrooms + flowers
		system.UpdateSpawnTimers(m.gameMap, initialItemCount, delta)
		system.UpdateSproutTimers(m.gameMap, initialItemCount, delta)
	}

	// Update item death timers (flowers die regardless of no-food mode)
	system.UpdateDeathTimers(m.gameMap, delta)

	// Update seed extraction cooldown timers
	system.UpdateSeedTimers(m.gameMap, delta)

	// Update manually watered tile timers
	m.gameMap.UpdateWateredTimers(delta)

	// Update ground spawning (sticks, nuts, shells)
	system.UpdateGroundSpawning(m.gameMap, delta, &m.groundSpawnTimers)

	// Tick down abandoned order cooldowns
	for _, order := range m.orders {
		if order.Status == entity.OrderAbandoned && order.AbandonCooldown > 0 {
			order.AbandonCooldown -= delta
			if order.AbandonCooldown <= 0 {
				order.AbandonCooldown = 0
				order.Status = entity.OrderOpen
			}
		}
	}

	// Calculate intents (Phase II ready: can parallelize this)
	items := m.gameMap.Items()
	for _, char := range m.gameMap.Characters() {
		oldIntent := char.Intent
		char.Intent = system.CalculateIntent(char, items, m.gameMap, m.actionLog, m.orders)

		// Reset action progress if intent action changed
		if oldIntent == nil || char.Intent == nil || oldIntent.Action != char.Intent.Action {
			char.ActionProgress = 0
		}
	}

	// Apply intents atomically
	for _, char := range m.gameMap.Characters() {
		m.applyIntent(char, delta)
	}

	// Remove completed orders
	m.sweepCompletedOrders()

	// Update cursor if following
	if m.following != nil {
		fpos := m.following.Pos()
		m.cursorX, m.cursorY = fpos.X, fpos.Y
	}

	return m, nil
}

// moveCursor moves the cursor and stops following
func (m *Model) moveCursor(dx, dy int) {
	m.following = nil
	m.logScrollOffset = 0
	nx := m.cursorX + dx
	ny := m.cursorY + dy
	if m.gameMap.IsValid(types.Position{X: nx, Y: ny}) {
		m.cursorX, m.cursorY = nx, ny
	}
}

// toggleFollow toggles following the character at cursor
func (m *Model) toggleFollow() {
	if char := m.characterAtCursor(); char != nil {
		if m.following == char {
			m.following = nil
		} else {
			m.following = char
		}
	}
}

// characterAtCursor returns the character at cursor position, or nil if none
func (m *Model) characterAtCursor() *entity.Character {
	if e := m.gameMap.EntityAt(types.Position{X: m.cursorX, Y: m.cursorY}); e != nil {
		if char, ok := e.(*entity.Character); ok {
			return char
		}
	}
	return nil
}

// handleNameEditKey processes keyboard input during character name editing
func (m Model) handleNameEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel editing, discard changes
		m.editingCharacterName = false
		m.editingCharacterID = 0
		m.editingNameBuffer = ""
		return m, nil

	case tea.KeyEnter:
		// Confirm editing if name is not empty
		if m.editingNameBuffer == "" {
			// Don't allow empty names, stay in edit mode
			return m, nil
		}
		// Find and update the character
		for _, char := range m.gameMap.Characters() {
			if char.ID == m.editingCharacterID {
				char.Name = m.editingNameBuffer
				break
			}
		}
		m.editingCharacterName = false
		m.editingCharacterID = 0
		m.editingNameBuffer = ""
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from buffer
		if len(m.editingNameBuffer) > 0 {
			m.editingNameBuffer = m.editingNameBuffer[:len(m.editingNameBuffer)-1]
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to buffer (respecting max length)
		if len(m.editingNameBuffer) < MaxNameLength {
			for _, r := range msg.Runes {
				if len(m.editingNameBuffer) < MaxNameLength {
					m.editingNameBuffer += string(r)
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// cycleToNextCharacter moves cursor and follow to the next alive character
func (m *Model) cycleToNextCharacter() {
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		return
	}

	// Build list of alive characters
	var alive []*entity.Character
	for _, c := range chars {
		if !c.IsDead {
			alive = append(alive, c)
		}
	}
	if len(alive) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, c := range alive {
		if c == m.following {
			currentIdx = i
			break
		}
	}

	// Move to next (wrap around)
	nextIdx := (currentIdx + 1) % len(alive)
	next := alive[nextIdx]

	m.following = next
	npos := next.Pos()
	m.cursorX, m.cursorY = npos.X, npos.Y
	m.logScrollOffset = 0
}

// cycleToPreviousCharacter moves cursor and follow to the previous alive character
func (m *Model) cycleToPreviousCharacter() {
	chars := m.gameMap.Characters()
	if len(chars) == 0 {
		return
	}

	// Build list of alive characters
	var alive []*entity.Character
	for _, c := range chars {
		if !c.IsDead {
			alive = append(alive, c)
		}
	}
	if len(alive) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, c := range alive {
		if c == m.following {
			currentIdx = i
			break
		}
	}

	// Move to previous (wrap around)
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(alive) - 1
	}
	prev := alive[prevIdx]

	m.following = prev
	ppos := prev.Pos()
	m.cursorX, m.cursorY = ppos.X, ppos.Y
	m.logScrollOffset = 0
}

// stepForward advances the game by one tick while paused
// One tick = 0.15s, which equals one move at speed 50
func (m *Model) stepForward() {
	delta := config.UpdateInterval.Seconds()
	m.elapsedGameTime += delta
	m.actionLog.SetGameTime(m.elapsedGameTime)

	// Update flash timer for status symbol cycling
	m.flashTimer += delta
	if m.flashTimer >= 0.5 {
		m.flashTimer = 0
		m.flashIndex++
	}

	// Update survival mechanics
	for _, char := range m.gameMap.Characters() {
		system.UpdateSurvival(char, delta, m.actionLog)
	}

	// Update item spawning (unless no food mode)
	if !m.testCfg.NoFood {
		initialItemCount := config.ItemSpawnCount*2 + config.FlowerSpawnCount // berries + mushrooms + flowers
		system.UpdateSpawnTimers(m.gameMap, initialItemCount, delta)
		system.UpdateSproutTimers(m.gameMap, initialItemCount, delta)
	}

	// Update item death timers (flowers die regardless of no-food mode)
	system.UpdateDeathTimers(m.gameMap, delta)

	// Update seed extraction cooldown timers
	system.UpdateSeedTimers(m.gameMap, delta)

	// Update manually watered tile timers
	m.gameMap.UpdateWateredTimers(delta)

	// Update ground spawning (sticks, nuts, shells)
	system.UpdateGroundSpawning(m.gameMap, delta, &m.groundSpawnTimers)

	// Calculate and apply intents
	items := m.gameMap.Items()
	for _, char := range m.gameMap.Characters() {
		oldIntent := char.Intent
		char.Intent = system.CalculateIntent(char, items, m.gameMap, m.actionLog, m.orders)
		if oldIntent == nil || char.Intent == nil || oldIntent.Action != char.Intent.Action {
			char.ActionProgress = 0
		}
	}

	for _, char := range m.gameMap.Characters() {
		m.applyIntent(char, delta)
	}

	// Remove completed orders
	m.sweepCompletedOrders()

	// Update cursor if following
	if m.following != nil {
		fpos := m.following.Pos()
		m.cursorX, m.cursorY = fpos.X, fpos.Y
	}
}

// getEdibleItemTypes returns item types that are edible (for character preferences)
func getEdibleItemTypes() []string {
	configs := game.GetItemTypeConfigs()
	var types []string
	for itemType, cfg := range configs {
		if cfg.Edible {
			types = append(types, itemType)
		}
	}
	return types
}

// handleWorldSelectKey handles input during world selection phase
func (m Model) handleWorldSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle delete confirmation mode
	if m.confirmingDelete >= 0 {
		switch msg.String() {
		case "y", "Y":
			// Confirm delete
			if m.confirmingDelete < len(m.worlds) {
				worldID := m.worlds[m.confirmingDelete].ID
				if err := save.DeleteWorld(worldID); err != nil {
					save.LogWarning("Failed to delete world %s: %v", worldID, err)
				}
				// Refresh world list
				m.worlds, _ = save.ListWorlds()
				// Adjust selection if needed
				if m.selectedWorld >= len(m.worlds) {
					m.selectedWorld = len(m.worlds) // Point to "New World"
				}
			}
			m.confirmingDelete = -1
		case "n", "N", "esc":
			// Cancel delete
			m.confirmingDelete = -1
		}
		return m, nil
	}

	maxIdx := len(m.worlds) // "New World" is at index len(worlds)

	switch msg.String() {
	case "up", "k":
		if m.selectedWorld > 0 {
			m.selectedWorld--
		}
	case "down", "j":
		if m.selectedWorld < maxIdx {
			m.selectedWorld++
		}
	case "enter":
		if m.selectedWorld < len(m.worlds) {
			// Load existing world
			return m.loadWorld(m.worlds[m.selectedWorld].ID)
		}
		// New World selected - go to character creation
		m.phase = phaseSelectMode
	case "d", "x":
		// Start delete confirmation (only for saved worlds, not "New World")
		if m.selectedWorld < len(m.worlds) {
			m.confirmingDelete = m.selectedWorld
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// loadWorld loads an existing world and returns to playing phase
func (m Model) loadWorld(worldID string) (Model, tea.Cmd) {
	state, err := save.LoadWorld(worldID)
	if err != nil {
		// Failed to load - stay on world select
		// TODO: Show error message
		return m, nil
	}

	// Restore model from save state
	m = FromSaveState(state, worldID, m.testCfg)
	m.paused = true // Start paused

	return m, tickCmd(m.speedMultiplier)
}

// handleCharacterCreationKey handles input during character creation phase
func (m Model) handleCharacterCreationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Return to start screen
		m.phase = phaseSelectMode
		m.creationState = nil
		return m, nil

	case tea.KeyEnter:
		// Start the game with current character settings
		return m.startGameFromCreation(), tickCmd(m.speedMultiplier)

	case tea.KeyLeft:
		m.creationState.NavigateCharacter(-1)

	case tea.KeyRight:
		m.creationState.NavigateCharacter(1)

	case tea.KeyUp:
		m.creationState.NavigateField(-1)

	case tea.KeyDown:
		m.creationState.NavigateField(1)

	case tea.KeyTab:
		m.creationState.NextField()

	case tea.KeySpace:
		if m.creationState.IsNameFieldSelected() {
			// Add space to name
			m.creationState.TypeCharacter(' ')
		} else {
			// Cycle option
			m.creationState.CycleOption()
		}

	case tea.KeyBackspace:
		if m.creationState.IsNameFieldSelected() {
			m.creationState.Backspace()
		}

	case tea.KeyCtrlR:
		// Randomize all characters
		m.creationState.RandomizeAll()

	case tea.KeyRunes:
		key := msg.String()
		switch key {
		case "+", "=":
			m.creationState.AddCharacter()
		case "-":
			m.creationState.RemoveLastCharacter()
		default:
			if m.creationState.IsNameFieldSelected() && len(key) == 1 {
				m.creationState.TypeCharacter(rune(key[0]))
			}
		}
	}

	return m, nil
}

// startGameFromCreation initializes the game from character creation settings
func (m Model) startGameFromCreation() Model {
	m.gameMap = game.NewMap(config.MapWidth, config.MapHeight)
	m.phase = phasePlaying
	m.lastUpdate = time.Now()

	// Clustered starting positions near center, 4 per row
	cx, cy := config.MapWidth/2, config.MapHeight/2

	var chars []*entity.Character
	for i, charData := range m.creationState.Characters {
		col := i % 4
		row := i / 4
		x := cx + col*2
		y := cy + row*2
		// Convert food/color display strings to lowercase for consistency
		food := DisplayToItemType(charData.Food)
		color := DisplayToColor(charData.Color)
		char := entity.NewCharacter(i+1, x, y, charData.Name, food, color)
		m.gameMap.AddCharacter(char)
		chars = append(chars, char)
	}

	// Randomly select one character to follow
	followIdx := rand.Intn(len(chars))
	m.following = chars[followIdx]
	fpos := chars[followIdx].Pos()
	m.cursorX, m.cursorY = fpos.X, fpos.Y

	// Clear creation state
	m.creationState = nil

	// Spawn world: ponds first (before items/features), then clay, then features, then items
	if !m.testCfg.NoWater {
		game.SpawnPonds(m.gameMap)
		game.SpawnClay(m.gameMap)
	}
	game.SpawnFeatures(m.gameMap, m.testCfg.NoWater, m.testCfg.NoBeds)
	if !m.testCfg.NoFood {
		game.SpawnItems(m.gameMap, m.testCfg.MushroomsOnly)
	}
	game.SpawnGroundItems(m.gameMap)
	m.groundSpawnTimers = system.GroundSpawnTimers{
		Stick: system.RandomGroundSpawnInterval(),
		Nut:   system.RandomGroundSpawnInterval(),
		Shell: system.RandomGroundSpawnInterval(),
	}

	// Create world for saving if not already set
	if m.worldID == "" {
		worldID, err := save.CreateWorld()
		if err == nil {
			m.worldID = worldID
		}
	}

	return m
}

// saveGame saves the current game state to disk
// Returns error if save fails, nil on success
func (m *Model) saveGame() error {
	if m.worldID == "" || m.gameMap == nil {
		return nil // Nothing to save
	}

	state := m.ToSaveState()
	if err := save.SaveWorld(m.worldID, state); err != nil {
		return err
	}

	// Update metadata
	chars := m.gameMap.Characters()
	aliveCount := 0
	for _, c := range chars {
		if !c.IsDead {
			aliveCount++
		}
	}

	meta, err := save.LoadMeta(m.worldID)
	if err != nil {
		return err
	}
	meta.LastPlayedAt = time.Now()
	meta.CharacterCount = len(chars)
	meta.AliveCount = aliveCount

	if err := save.SaveMeta(m.worldID, meta); err != nil {
		return err
	}

	m.lastSaveGameTime = m.elapsedGameTime
	m.saveIndicatorEnd = time.Now().Add(1 * time.Second) // Show "Saving" for 1 second
	return nil
}

// applyOrdersConfirm executes the confirm action for the current orders selection.
// Used by both Enter key and number key (select-and-confirm) handlers.
func (m *Model) applyOrdersConfirm() {
	if m.ordersAddMode {
		if m.ordersAddStep == 0 {
			activities := m.getOrderableActivities()
			if len(activities) > 0 && m.selectedActivityIndex < len(activities) {
				selectedActivity := activities[m.selectedActivityIndex]
				// Dig has no sub-menu — create order immediately
				if selectedActivity.ID == "dig" {
					order := entity.NewOrder(m.nextOrderID, "dig", "clay")
					m.nextOrderID++
					m.orders = append(m.orders, order)
					m.setOrderFlash(order.DisplayName())
					m.ordersAddStep = 0
					m.selectedActivityIndex = 0
				} else {
					m.ordersAddStep = 1
					m.selectedTargetIndex = 0
				}
			}
		} else if m.ordersAddStep == 1 {
			activities := m.getOrderableActivities()
			if len(activities) > 0 && m.selectedActivityIndex < len(activities) {
				selectedActivity := activities[m.selectedActivityIndex]

				if isSyntheticCategory(selectedActivity.ID) {
					category := syntheticCategoryID(selectedActivity.ID)
					categoryActivities := m.getCategoryActivities(category)
					if m.selectedTargetIndex < len(categoryActivities) {
						catActivity := categoryActivities[m.selectedTargetIndex]

						if catActivity.ID == "tillSoil" {
							m.ordersAddStep = 2
							m.step2ActivityID = "tillSoil"
							m.areaSelectAnchor = nil
							m.areaSelectUnmarkMode = false
						} else if catActivity.ID == "plant" {
							m.ordersAddStep = 2
							m.step2ActivityID = "plant"
							m.selectedPlantTypeIndex = 0
						} else if catActivity.ID == "buildFence" {
							m.ordersAddStep = 2
							m.step2ActivityID = "buildFence"
							m.areaSelectAnchor = nil
							m.areaSelectUnmarkMode = false
						} else {
							order := entity.NewOrder(m.nextOrderID, catActivity.ID, "")
							m.nextOrderID++
							m.orders = append(m.orders, order)
							m.setOrderFlash(order.DisplayName())
							m.ordersAddStep = 0
							m.selectedActivityIndex = 0
						}
					}
				} else if selectedActivity.ID == "gather" {
					gatherTypes := game.GetGatherableTypes(m.gameMap.Items())
					if m.selectedTargetIndex < len(gatherTypes) {
						targetType := gatherTypes[m.selectedTargetIndex].TargetType
						order := entity.NewOrder(m.nextOrderID, "gather", targetType)
						m.nextOrderID++
						m.orders = append(m.orders, order)
						m.setOrderFlash(order.DisplayName())
						m.ordersAddStep = 0
						m.selectedActivityIndex = 0
					}
				} else if selectedActivity.ID == "extract" {
					extractTypes := game.GetExtractableItemTypes(m.gameMap.Items())
					if m.selectedTargetIndex < len(extractTypes) {
						targetType := extractTypes[m.selectedTargetIndex].TargetType
						order := entity.NewOrder(m.nextOrderID, "extract", targetType)
						m.nextOrderID++
						m.orders = append(m.orders, order)
						m.setOrderFlash(order.DisplayName())
						m.ordersAddStep = 0
						m.selectedActivityIndex = 0
					}
				} else {
					types := m.getHarvestableItemTypes()
					if m.selectedTargetIndex < len(types) {
						targetType := types[m.selectedTargetIndex]
						order := entity.NewOrder(m.nextOrderID, selectedActivity.ID, targetType)
						m.nextOrderID++
						m.orders = append(m.orders, order)
						m.setOrderFlash(order.DisplayName())
						m.ordersAddStep = 0
						m.selectedActivityIndex = 0
					}
				}
			}
		} else if m.ordersAddStep == 2 {
			if m.step2ActivityID == "plant" {
				plantTypes := game.GetPlantableTypes(m.gameMap.Items(), m.gameMap.Characters())
				if m.selectedPlantTypeIndex < len(plantTypes) {
					order := entity.NewOrder(m.nextOrderID, "plant", plantTypes[m.selectedPlantTypeIndex].TargetType)
					m.nextOrderID++
					m.orders = append(m.orders, order)
					m.setOrderFlash(order.DisplayName())
					// Go back to step 1 (Gardening sub-category) so player can immediately create another order
					m.ordersAddStep = 1
					m.selectedTargetIndex = 0
				}
			} else if m.step2ActivityID == "buildFence" {
				// buildFence: Enter = done, create order if tiles marked
				if len(m.gameMap.MarkedForConstructionPositions()) > 0 {
					order := entity.NewOrder(m.nextOrderID, "buildFence", "")
					m.nextOrderID++
					m.orders = append(m.orders, order)
					m.setOrderFlash(order.DisplayName())
				}
				m.ordersAddStep = 1
				m.selectedTargetIndex = 0
				m.areaSelectAnchor = nil
				m.areaSelectUnmarkMode = false
			} else {
				// tillSoil: Enter = done, create order if tiles marked
				if len(m.gameMap.MarkedForTillingPositions()) > 0 {
					order := entity.NewOrder(m.nextOrderID, "tillSoil", "")
					m.nextOrderID++
					m.orders = append(m.orders, order)
					m.setOrderFlash(order.DisplayName())
				}
				m.ordersAddStep = 1
				m.selectedTargetIndex = 0
				m.areaSelectAnchor = nil
				m.areaSelectUnmarkMode = false
			}
		}
	} else if m.ordersCancelMode {
		if m.selectedOrderIndex < len(m.orders) {
			order := m.orders[m.selectedOrderIndex]

			// Clear assignment from character if order was assigned
			if order.AssignedTo != 0 {
				for _, char := range m.gameMap.Characters() {
					if char.ID == order.AssignedTo {
						char.AssignedOrderID = 0
						char.Intent = nil
						break
					}
				}
			}

			m.orders = append(m.orders[:m.selectedOrderIndex], m.orders[m.selectedOrderIndex+1:]...)

			if m.selectedOrderIndex >= len(m.orders) && m.selectedOrderIndex > 0 {
				m.selectedOrderIndex--
			}

			if len(m.orders) == 0 {
				m.ordersCancelMode = false
			}
		}
	}
}

// setOrderFlash sets or updates the order creation flash confirmation.
// If the same order type is created consecutively within the flash duration,
// the count increments. Otherwise, it resets to 1.
func (m *Model) setOrderFlash(displayName string) {
	now := time.Now()
	if displayName == m.orderFlashMessage && now.Before(m.orderFlashEnd) {
		m.orderFlashCount++
	} else {
		m.orderFlashMessage = displayName
		m.orderFlashCount = 1
	}
	m.orderFlashEnd = now.Add(2 * time.Second)
}
