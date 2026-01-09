package ui

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
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
			return m, tickCmd()
		}
		newModel, _ := m.updateGame(time.Time(msg))
		return newModel, tickCmd()
	}

	return m, nil
}

// handleKey processes keyboard input
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseSelectMode:
		switch msg.String() {
		case "1":
			// Single character mode only available in debug mode
			if m.testCfg.Debug {
				m.multiCharMode = false
				m.phase = phaseSelectFood
			}
		case "m", "M":
			m.multiCharMode = true
			m.creationState = NewCharacterCreationState()
			m.phase = phaseCharacterCreate
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case phaseCharacterCreate:
		return m.handleCharacterCreationKey(msg)

	case phaseSelectFood:
		switch msg.String() {
		case "b", "B":
			m.selectedFood = "berry"
			m.phase = phaseSelectColor
		case "m", "M":
			m.selectedFood = "mushroom"
			m.phase = phaseSelectColor
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case phaseSelectColor:
		switch msg.String() {
		case "r", "R":
			m.selectedColor = types.ColorRed
			return m.startGame(), tickCmd()
		case "l", "L":
			m.selectedColor = types.ColorBlue
			return m.startGame(), tickCmd()
		case "w", "W":
			m.selectedColor = types.ColorWhite
			return m.startGame(), tickCmd()
		case "n", "N":
			m.selectedColor = types.ColorBrown
			return m.startGame(), tickCmd()
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case phasePlaying:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			if !m.paused {
				// Reset lastUpdate when unpausing to prevent accumulated delta
				m.lastUpdate = time.Now()
			}
		case "f", "F":
			m.toggleFollow()
		case "a":
			// Switch to All Activity mode
			m.viewMode = viewModeAllActivity
			m.logScrollOffset = 0
		case "s":
			// Switch to Select mode
			m.viewMode = viewModeSelect
			m.logScrollOffset = 0
		case "n":
			// Cycle to next alive character
			m.cycleToNextCharacter()
		case "up":
			m.moveCursor(0, -1)
		case "down":
			m.moveCursor(0, 1)
		case "left":
			m.moveCursor(-1, 0)
		case "right":
			m.moveCursor(1, 0)
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
			if e := m.gameMap.EntityAt(m.cursorX, m.cursorY); e != nil {
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

// startGame initializes the game world
func (m Model) startGame() Model {
	m.gameMap = game.NewMap(config.MapWidth, config.MapHeight)
	m.phase = phasePlaying
	m.lastUpdate = time.Now()

	// Create character at center
	cx, cy := config.MapWidth/2, config.MapHeight/2
	char := entity.NewCharacter(1, cx, cy, "Len", m.selectedFood, m.selectedColor)
	m.gameMap.AddCharacter(char)
	m.cursorX, m.cursorY = cx, cy
	m.following = char

	// Spawn items and features (respecting test config)
	if !m.testCfg.NoFood {
		game.SpawnItems(m.gameMap, m.testCfg.MushroomsOnly)
	}
	game.SpawnFeatures(m.gameMap, m.testCfg.NoWater, m.testCfg.NoBeds)

	return m
}

// startGameMulti initializes the game world with 4 characters
func (m Model) startGameMulti() Model {
	m.gameMap = game.NewMap(config.MapWidth, config.MapHeight)
	m.phase = phasePlaying
	m.lastUpdate = time.Now()

	// Character names and preferences
	names := []string{"Len", "Macca", "Hari", "Starr"}
	foods := []string{"berry", "mushroom"}
	colors := types.AllColors

	// Clustered starting positions near center
	cx, cy := config.MapWidth/2, config.MapHeight/2
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
	m.cursorX, m.cursorY = chars[followIdx].Position()

	// Spawn items and features (respecting test config)
	if !m.testCfg.NoFood {
		game.SpawnItems(m.gameMap, m.testCfg.MushroomsOnly)
	}
	game.SpawnFeatures(m.gameMap, m.testCfg.NoWater, m.testCfg.NoBeds)

	return m
}

// updateGame processes one game tick
func (m Model) updateGame(now time.Time) (Model, tea.Cmd) {
	delta := now.Sub(m.lastUpdate).Seconds()
	m.lastUpdate = now

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
	}

	// Calculate intents (Phase II ready: can parallelize this)
	items := m.gameMap.Items()
	for _, char := range m.gameMap.Characters() {
		oldIntent := char.Intent
		char.Intent = system.CalculateIntent(char, items, m.gameMap, m.actionLog)

		// Reset action progress if intent action changed
		if oldIntent == nil || char.Intent == nil || oldIntent.Action != char.Intent.Action {
			char.ActionProgress = 0
		}
	}

	// Apply intents atomically
	for _, char := range m.gameMap.Characters() {
		m.applyIntent(char, delta)
	}

	// Update cursor if following
	if m.following != nil {
		m.cursorX, m.cursorY = m.following.Position()
	}

	return m, nil
}

// applyIntent executes a character's intent
func (m *Model) applyIntent(char *entity.Character, delta float64) {
	if char.Intent == nil || char.IsDead || char.IsSleeping {
		return
	}

	switch char.Intent.Action {
	case entity.ActionMove:
		cx, cy := char.Position()
		tx, ty := char.Intent.TargetX, char.Intent.TargetY

		// Check if at target item - eating takes duration
		if char.Intent.TargetItem != nil {
			ix, iy := char.Intent.TargetItem.Position()
			if cx == ix && cy == iy {
				// Already at food - eating in progress
				char.ActionProgress += delta
				if char.ActionProgress >= config.ActionDuration {
					char.ActionProgress = 0
					if item := m.gameMap.ItemAt(cx, cy); item == char.Intent.TargetItem {
						system.Consume(char, item, m.gameMap, m.actionLog)
					}
				}
				return
			}
		}

		// Movement is gated by speed accumulator
		speed := char.EffectiveSpeed()
		char.SpeedAccumulator += float64(speed) * delta

		// Check if we've accumulated enough "movement points" to act
		// At speed 50 (baseline), character moves once per 0.02 seconds (50 ticks/sec)
		// This is scaled so baseline speed = ~1 action per game tick (0.15s)
		const movementThreshold = 7.5 // 50 speed * 0.15s delta = 7.5

		if char.SpeedAccumulator < movementThreshold {
			return
		}

		// Consume accumulated points
		char.SpeedAccumulator -= movementThreshold

		// Try to move, with collision handling (max 1 character per position)
		moved := false
		triedPositions := make(map[[2]int]bool)
		triedPositions[[2]int{tx, ty}] = true

		for attempts := 0; attempts < 5 && !moved; attempts++ {
			if m.gameMap.MoveCharacter(char, tx, ty) {
				moved = true
				break
			}
			// Position blocked, try alternate
			altStep := m.findAlternateStep(char, cx, cy, triedPositions)
			if altStep == nil {
				break // No valid alternatives
			}
			tx, ty = altStep[0], altStep[1]
			triedPositions[[2]int{tx, ty}] = true
		}

		if !moved {
			// Couldn't move anywhere this tick, refund some speed points
			char.SpeedAccumulator += movementThreshold * 0.5
			return
		}

		// Drain energy for movement (unless on cooldown - freshly rested burst)
		if char.EnergyCooldown <= 0 {
			prevEnergy := char.Energy
			char.Energy -= config.EnergyMovementDrain
			if char.Energy < 0 {
				char.Energy = 0
			}

			// Log energy milestones crossed by movement drain
			if !char.IsSleeping && m.actionLog != nil {
				if prevEnergy > 50 && char.Energy <= 50 {
					m.actionLog.Add(char.ID, char.Name, "energy", "Getting tired")
				}
				if prevEnergy > 25 && char.Energy <= 25 {
					m.actionLog.Add(char.ID, char.Name, "energy", "Very tired!")
				}
				if prevEnergy > 10 && char.Energy <= 10 {
					m.actionLog.Add(char.ID, char.Name, "energy", "Exhausted!")
				}
				if prevEnergy > 0 && char.Energy <= 0 {
					m.actionLog.Add(char.ID, char.Name, "energy", "Collapsed from exhaustion!")
				}
			}
		}

	case entity.ActionDrink:
		// Drinking requires duration to complete
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDuration {
			char.ActionProgress = 0
			system.Drink(char, m.actionLog)
		}

	case entity.ActionSleep:
		atBed := char.Intent.TargetFeature != nil && char.Intent.TargetFeature.IsBed()

		// Collapse is immediate (involuntary) - only at Energy 0
		if !atBed && char.Energy <= 0 {
			system.StartSleep(char, false, m.actionLog)
			return
		}

		// Voluntary sleep (bed or ground) requires duration to complete
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDuration {
			char.ActionProgress = 0
			system.StartSleep(char, atBed, m.actionLog)
		}

	case entity.ActionLook:
		// Looking requires duration to complete
		char.ActionProgress += delta
		if char.ActionProgress >= config.LookDuration {
			char.ActionProgress = 0
			system.CompleteLook(char, char.Intent.TargetItem, m.actionLog)
		}
	}
}

// moveCursor moves the cursor and stops following
func (m *Model) moveCursor(dx, dy int) {
	m.following = nil
	m.logScrollOffset = 0
	nx := m.cursorX + dx
	ny := m.cursorY + dy
	if m.gameMap.IsValid(nx, ny) {
		m.cursorX, m.cursorY = nx, ny
	}
}

// toggleFollow toggles following the character at cursor
func (m *Model) toggleFollow() {
	if e := m.gameMap.EntityAt(m.cursorX, m.cursorY); e != nil {
		if char, ok := e.(*entity.Character); ok {
			if m.following == char {
				m.following = nil
			} else {
				m.following = char
			}
		}
	}
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
	m.cursorX, m.cursorY = next.Position()
	m.logScrollOffset = 0
}

// stepForward advances the game by one tick while paused
// One tick = 0.15s, which equals one move at speed 50
func (m *Model) stepForward() {
	delta := config.UpdateInterval.Seconds()

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
	}

	// Calculate and apply intents
	items := m.gameMap.Items()
	for _, char := range m.gameMap.Characters() {
		oldIntent := char.Intent
		char.Intent = system.CalculateIntent(char, items, m.gameMap, m.actionLog)
		if oldIntent == nil || char.Intent == nil || oldIntent.Action != char.Intent.Action {
			char.ActionProgress = 0
		}
	}

	for _, char := range m.gameMap.Characters() {
		m.applyIntent(char, delta)
	}

	// Update cursor if following
	if m.following != nil {
		m.cursorX, m.cursorY = m.following.Position()
	}
}

// findAlternateStep finds an alternate step when the preferred step is blocked
// Returns [x, y] of alternate position, or nil if no valid alternative
// triedPositions contains positions already attempted this tick
func (m *Model) findAlternateStep(char *entity.Character, cx, cy int, triedPositions map[[2]int]bool) []int {
	// Get ultimate goal position
	var goalX, goalY int
	if char.Intent.TargetItem != nil {
		goalX, goalY = char.Intent.TargetItem.Position()
	} else if char.Intent.TargetFeature != nil {
		goalX, goalY = char.Intent.TargetFeature.Position()
	} else {
		// No clear goal, can't find alternate
		return nil
	}

	// Try adjacent positions that still move toward goal
	// Priority: diagonal toward goal > orthogonal toward goal > wait
	dx := sign(goalX - cx)
	dy := sign(goalY - cy)

	// Build list of candidate positions
	candidates := [][]int{}

	// Primary alternates: move in one axis toward goal
	if dx != 0 {
		candidates = append(candidates, []int{cx + dx, cy})
	}
	if dy != 0 {
		candidates = append(candidates, []int{cx, cy + dy})
	}

	// Secondary: orthogonal moves that don't move away from goal
	if dx == 0 {
		candidates = append(candidates, []int{cx + 1, cy}, []int{cx - 1, cy})
	}
	if dy == 0 {
		candidates = append(candidates, []int{cx, cy + 1}, []int{cx, cy - 1})
	}

	// Tertiary: all other adjacent positions as last resort
	allAdjacent := [][]int{
		{cx + 1, cy}, {cx - 1, cy}, {cx, cy + 1}, {cx, cy - 1},
		{cx + 1, cy + 1}, {cx + 1, cy - 1}, {cx - 1, cy + 1}, {cx - 1, cy - 1},
	}
	for _, pos := range allAdjacent {
		found := false
		for _, c := range candidates {
			if c[0] == pos[0] && c[1] == pos[1] {
				found = true
				break
			}
		}
		if !found {
			candidates = append(candidates, pos)
		}
	}

	// Find first valid candidate not already tried
	for _, pos := range candidates {
		x, y := pos[0], pos[1]
		key := [2]int{x, y}
		if triedPositions[key] {
			continue
		}
		if !m.gameMap.IsValid(x, y) {
			continue
		}
		if m.gameMap.IsOccupied(x, y) {
			continue
		}
		return pos
	}

	return nil
}

// sign returns -1, 0, or 1 based on the value
func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
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
		return m.startGameFromCreation(), tickCmd()

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
		// Handle regular character input for name field
		if m.creationState.IsNameFieldSelected() {
			key := msg.String()
			if len(key) == 1 {
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

	// Clustered starting positions near center
	cx, cy := config.MapWidth/2, config.MapHeight/2
	offsets := [][2]int{{0, 0}, {2, 0}, {0, 2}, {2, 2}}

	var chars []*entity.Character
	for i, charData := range m.creationState.Characters {
		x := cx + offsets[i][0]
		y := cy + offsets[i][1]
		// Convert food/color to lowercase for consistency with existing code
		food := charData.Food
		if food == FoodBerry {
			food = "berry"
		} else {
			food = "mushroom"
		}
		color := DisplayToColor(charData.Color)
		char := entity.NewCharacter(i+1, x, y, charData.Name, food, color)
		m.gameMap.AddCharacter(char)
		chars = append(chars, char)
	}

	// Randomly select one character to follow
	followIdx := rand.Intn(len(chars))
	m.following = chars[followIdx]
	m.cursorX, m.cursorY = chars[followIdx].Position()

	// Clear creation state
	m.creationState = nil

	// Spawn items and features (respecting test config)
	if !m.testCfg.NoFood {
		game.SpawnItems(m.gameMap, m.testCfg.MushroomsOnly)
	}
	game.SpawnFeatures(m.gameMap, m.testCfg.NoWater, m.testCfg.NoBeds)

	return m
}
