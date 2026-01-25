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
	case phaseWorldSelect:
		return m.handleWorldSelectKey(msg)

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
		case "q", "ctrl+c":
			m.saveGame() // Save before quitting
			return m, tea.Quit
		case "esc":
			// If in orders add/cancel mode, exit that mode
			if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
				m.ordersAddMode = false
				m.ordersCancelMode = false
				return m, nil
			}
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
			m.showOrdersPanel = false
			m.logScrollOffset = 0
			return m, tea.Batch(tea.ClearScreen, tea.WindowSize())
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
		case "enter":
			// Confirm selection in orders panel
			if m.showOrdersPanel {
				if m.ordersAddMode {
					// Handle add order confirmation inline
					if m.ordersAddStep == 0 {
						activities := m.getOrderableActivities()
						if len(activities) > 0 && m.selectedActivityIndex < len(activities) {
							// Both Harvest and Craft need step 1 selection
							// Harvest selects target type, Craft selects craft activity
							m.ordersAddStep = 1
							m.selectedTargetIndex = 0
						}
					} else {
						// Step 1: Create the order based on what was selected
						activities := m.getOrderableActivities()
						if len(activities) > 0 && m.selectedActivityIndex < len(activities) {
							selectedActivity := activities[m.selectedActivityIndex]

							if isCraftCategory(selectedActivity.ID) {
								// Craft category - get selected craft activity
								craftActivities := m.getCraftActivities()
								if m.selectedTargetIndex < len(craftActivities) {
									craftActivity := craftActivities[m.selectedTargetIndex]
									order := entity.NewOrder(m.nextOrderID, craftActivity.ID, "")
									m.nextOrderID++
									m.orders = append(m.orders, order)

									m.ordersAddMode = false
									m.ordersAddStep = 0
								}
							} else {
								// Harvest - get selected target type
								types := m.getEdibleItemTypes()
								if m.selectedTargetIndex < len(types) {
									targetType := types[m.selectedTargetIndex]
									order := entity.NewOrder(m.nextOrderID, selectedActivity.ID, targetType)
									m.nextOrderID++
									m.orders = append(m.orders, order)

									m.ordersAddMode = false
									m.ordersAddStep = 0
								}
							}
						}
					}
				} else if m.ordersCancelMode {
					// Handle cancel order confirmation inline
					if m.selectedOrderIndex < len(m.orders) {
						order := m.orders[m.selectedOrderIndex]

						// Clear assignment from character if order was assigned
						if order.AssignedTo != 0 {
							for _, char := range m.gameMap.Characters() {
								if char.ID == order.AssignedTo {
									char.AssignedOrderID = 0
									char.Intent = nil // Clear intent so they re-evaluate
									break
								}
							}
						}

						// Remove the selected order
						m.orders = append(m.orders[:m.selectedOrderIndex], m.orders[m.selectedOrderIndex+1:]...)

						// Adjust selection if needed
						if m.selectedOrderIndex >= len(m.orders) && m.selectedOrderIndex > 0 {
							m.selectedOrderIndex--
						}

						// Exit cancel mode if no orders left
						if len(m.orders) == 0 {
							m.ordersCancelMode = false
						}
					}
				}
				return m, nil
			}
		case "k", "K":
			// Toggle knowledge panel (only in select mode, mutually exclusive with inventory)
			if m.viewMode == viewModeSelect {
				m.showKnowledgePanel = !m.showKnowledgePanel
				if m.showKnowledgePanel {
					m.showInventoryPanel = false
				}
			}
		case "i", "I":
			// Toggle inventory panel (only in select mode, mutually exclusive with knowledge)
			if m.viewMode == viewModeSelect {
				m.showInventoryPanel = !m.showInventoryPanel
				if m.showInventoryPanel {
					m.showKnowledgePanel = false
				}
			}
		case "n":
			// Cycle to next alive character
			m.cycleToNextCharacter()
		case "up":
			if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
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
			if m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode) {
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
							isCraftCategory(activities[m.selectedActivityIndex].ID) {
							maxIndex = len(m.getCraftActivities()) - 1
						} else {
							maxIndex = len(m.getEdibleItemTypes()) - 1
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
			if !(m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode)) {
				m.moveCursor(-1, 0)
			}
		case "right":
			if !(m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode)) {
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

	// Center position for cursor fallback
	cx, cy := config.MapWidth/2, config.MapHeight/2
	m.cursorX, m.cursorY = cx, cy

	// Spawn characters unless disabled
	if !m.testCfg.NoCharacters {
		names := []string{"Len", "Macca", "Hari", "Starr"}
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
		m.cursorX, m.cursorY = chars[followIdx].Position()
	}

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
	}

	// Update item death timers (flowers die regardless of no-food mode)
	system.UpdateDeathTimers(m.gameMap, delta)

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

	// Collapse is immediate and involuntary - check before any action
	if char.Energy <= 0 {
		system.StartSleep(char, false, m.actionLog)
		return
	}

	switch char.Intent.Action {
	case entity.ActionMove:
		cx, cy := char.Position()
		tx, ty := char.Intent.TargetX, char.Intent.TargetY

		// Check if at target item for eating - only if driven by hunger and item is edible
		if char.Intent.TargetItem != nil &&
			char.Intent.DrivingStat == types.StatHunger &&
			char.Intent.TargetItem.Edible {
			ix, iy := char.Intent.TargetItem.Position()
			if cx == ix && cy == iy {
				// At target item - eating in progress
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
			// Clear intent so CalculateIntent will re-evaluate next tick
			char.Intent = nil
			// Set idle cooldown so we don't immediately try another idle activity
			char.IdleCooldown = config.IdleCooldown
		}

	case entity.ActionTalk:
		target := char.Intent.TargetCharacter
		if target == nil {
			return
		}

		// If not already talking, start the conversation
		if char.TalkingWith == nil {
			system.StartTalking(char, target, m.actionLog)
		}

		// Decrement talk timer
		char.TalkTimer -= delta
		if char.TalkTimer <= 0 {
			// Talk complete - transmit knowledge, then stop talking
			system.TransmitKnowledge(char, target, m.actionLog)
			system.StopTalking(char, target, m.actionLog)
		}

	case entity.ActionPickup:
		// Picking up an item (used by both foraging and harvest orders)
		cx, cy := char.Position()

		if char.Intent.TargetItem == nil {
			return
		}

		ix, iy := char.Intent.TargetItem.Position()

		// Check if at target item
		if cx == ix && cy == iy {
			// At item - pickup in progress
			char.ActionProgress += delta
			if char.ActionProgress >= config.ActionDuration {
				char.ActionProgress = 0
				if item := m.gameMap.ItemAt(cx, cy); item == char.Intent.TargetItem {
					// If on an order and inventory full, drop current item first
					// BUT don't drop if carrying a vessel with space (can add to it)
					// (If carrying a recipe input, we'd have ActionCraft intent instead)
					if char.AssignedOrderID != 0 && char.IsInventoryFull() {
						// Check if carrying vessel with space - don't drop, will add to it
						canAddToVessel := char.Carrying != nil &&
							char.Carrying.Container != nil &&
							!system.IsVesselFull(char.Carrying, m.gameMap.Varieties())
						if !canAddToVessel {
							system.Drop(char, m.gameMap, m.actionLog)
						}
					}
					result := system.Pickup(char, item, m.gameMap, m.actionLog, m.gameMap.Varieties())

					// Handle vessel filling continuation
					if result == system.PickupToVessel {
						// Item was added to vessel - check if we should continue
						if nextIntent := system.FindNextVesselTarget(char, cx, cy, m.gameMap.Items(), m.gameMap.Varieties()); nextIntent != nil {
							char.Intent = nextIntent
							return
						}
						// Vessel full or no more matching targets - stop
						char.Intent = nil
						char.IdleCooldown = config.IdleCooldown
						char.CurrentActivity = "Idle"

						// Check for harvest order completion (vessel full or no matching items)
						if char.AssignedOrderID != 0 {
							if order := m.findOrderByID(char.AssignedOrderID); order != nil {
								if order.ActivityID == "harvest" {
									system.CompleteOrder(char, order, m.actionLog)
									m.removeOrder(order.ID)
								}
							}
						}
						return
					}

					// Handle pickup failure (variety mismatch with carried vessel)
					// This shouldn't happen with proper intent filtering, but handle gracefully
					if result == system.PickupFailed {
						char.Intent = nil
						char.IdleCooldown = config.IdleCooldown
						char.CurrentActivity = "Idle"
						return
					}

					// PickupToInventory - inventory now full (one item)
					// Check for harvest order completion
					// Craft orders don't complete on pickup - they complete after crafting
					// If picked up a vessel, continue harvesting into it (don't complete order)
					if char.AssignedOrderID != 0 && (char.Carrying == nil || char.Carrying.Container == nil) {
						if order := m.findOrderByID(char.AssignedOrderID); order != nil {
							if order.ActivityID == "harvest" {
								system.CompleteOrder(char, order, m.actionLog)
								m.removeOrder(order.ID)
							}
						}
					}
				}
			}
			return
		}

		// Not at item yet - move toward it
		speed := char.EffectiveSpeed()
		char.SpeedAccumulator += float64(speed) * delta

		const movementThreshold = 7.5

		if char.SpeedAccumulator < movementThreshold {
			return
		}

		char.SpeedAccumulator -= movementThreshold

		// Move toward target item
		tx, ty := char.Intent.TargetX, char.Intent.TargetY
		if m.gameMap.MoveCharacter(char, tx, ty) {
			// Successfully moved - update intent for next step
			newX, newY := char.Position()
			if newX != ix || newY != iy {
				// Not at item yet, calculate next step
				nx, ny := system.NextStep(newX, newY, ix, iy)
				char.Intent.TargetX = nx
				char.Intent.TargetY = ny
			}
		}

	case entity.ActionConsume:
		// Eating from inventory - no movement needed, just duration
		char.ActionProgress += delta
		if char.ActionProgress >= config.ActionDuration {
			char.ActionProgress = 0
			// Verify target item matches what we're carrying
			if char.Carrying == char.Intent.TargetItem {
				system.ConsumeFromInventory(char, char.Carrying, m.actionLog)
			}
		}

	case entity.ActionCraft:
		// Crafting - uses recipe duration
		if char.Carrying == nil || char.Carrying.ItemType != "gourd" {
			// No gourd to craft with - clear intent
			char.Intent = nil
			return
		}

		// Get recipe duration (hollow-gourd recipe)
		recipe := entity.RecipeRegistry["hollow-gourd"]
		if recipe == nil {
			char.Intent = nil
			return
		}

		char.ActionProgress += delta
		if char.ActionProgress >= recipe.Duration {
			char.ActionProgress = 0

			// Complete the craft
			gourd := char.Carrying
			vessel := system.CreateVessel(gourd, recipe)
			char.Carrying = vessel

			// Log the craft
			if m.actionLog != nil {
				m.actionLog.Add(char.ID, char.Name, "activity", "Crafted "+recipe.Name)
			}

			// Complete the order
			if char.AssignedOrderID != 0 {
				if order := m.findOrderByID(char.AssignedOrderID); order != nil {
					system.CompleteOrder(char, order, m.actionLog)
					m.removeOrder(order.ID)
				}
			}

			char.CurrentActivity = "Idle"
			char.Intent = nil
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

	// Update item death timers (flowers die regardless of no-food mode)
	system.UpdateDeathTimers(m.gameMap, delta)

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
				save.DeleteWorld(worldID)
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

	return m, tickCmd()
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
	m.cursorX, m.cursorY = chars[followIdx].Position()

	// Clear creation state
	m.creationState = nil

	// Spawn items and features (respecting test config)
	if !m.testCfg.NoFood {
		game.SpawnItems(m.gameMap, m.testCfg.MushroomsOnly)
	}
	game.SpawnFeatures(m.gameMap, m.testCfg.NoWater, m.testCfg.NoBeds)

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

