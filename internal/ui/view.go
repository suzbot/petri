package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/system"
	"petri/internal/types"
)

// Regex patterns for stripping debug info from log messages
var (
	parenPattern  = regexp.MustCompile(`\s*\([^)]*\)`)         // (anything in parens)
	healthPattern = regexp.MustCompile(`\s*Health:\s*\d+/\d+`) // Health: X/100
)

// colorByTier applies color to a level name based on severity tier
func colorByTier(level string, tier int) string {
	switch tier {
	case entity.TierNone:
		return optimalStyle.Render(level) // dark green for optimal
	case entity.TierCrisis:
		return crisisStyle.Render(level)
	case entity.TierSevere:
		return severeStyle.Render(level)
	default:
		return level
	}
}

// colorLogMessage colors action log messages based on content
func colorLogMessage(line, message string) string {
	// Learning/discovery messages (darker blue) - check first as these are important
	if strings.Contains(message, "Learned") || strings.Contains(message, "Discovered") {
		return learnedStyle.Render(line)
	}

	// Effect wore off / recovery messages (cyan)
	woreOffKeywords := []string{"Calmed down", "Woke up", "Poison wore off"}
	for _, kw := range woreOffKeywords {
		if strings.Contains(message, kw) {
			return woreOffStyle.Render(line)
		}
	}

	// Crisis tier messages (red)
	crisisKeywords := []string{
		"Starving", "Dehydrated", "Collapsed", "Dying", "Died", "Miserable",
	}
	for _, kw := range crisisKeywords {
		if strings.Contains(message, kw) {
			return crisisStyle.Render(line)
		}
	}

	// Severe tier messages (yellow)
	severeKeywords := []string{
		"Ravenous", "Parched", "Exhausted", "Critical", "Unhappy",
	}
	for _, kw := range severeKeywords {
		if strings.Contains(message, kw) {
			return severeStyle.Render(line)
		}
	}

	// Frustrated messages (orange)
	if strings.Contains(message, "Frustrated") {
		return frustratedStyle.Render(line)
	}

	// Sleeping messages (purple)
	if strings.Contains(message, "asleep") {
		return sleepingStyle.Render(line)
	}

	// Order-related messages (coral/salmon)
	if strings.Contains(message, "order:") {
		return orderStyle.Render(line)
	}

	// Poison messages (green)
	if strings.Contains(message, "poisoned") {
		return poisonedStyle.Render(line)
	}

	// Optimal tier messages (dark green)
	optimalKeywords := []string{"Joyful", "impacted health"}
	for _, kw := range optimalKeywords {
		if strings.Contains(message, kw) {
			return optimalStyle.Render(line)
		}
	}

	// Preference messages
	// "No longer" checked first since it contains "Likes"/"Dislikes"
	if strings.Contains(message, "No longer") {
		return woreOffStyle.Render(line)
	}
	if strings.Contains(message, "New Opinion: Likes") {
		return optimalStyle.Render(line)
	}
	if strings.Contains(message, "New Opinion: Dislikes") {
		return severeStyle.Render(line)
	}

	return line
}

// View implements tea.Model
func (m Model) View() string {
	switch m.phase {
	case phaseWorldSelect:
		return m.viewWorldSelect()
	case phaseSelectMode:
		return m.viewModeSelect()
	case phaseCharacterCreate:
		return m.viewCharacterCreate()
	case phaseSelectFood:
		return m.viewFoodSelect()
	case phaseSelectColor:
		return m.viewColorSelect()
	default:
		return m.viewGame()
	}
}

// viewWorldSelect renders the world selection screen
func (m Model) viewWorldSelect() string {
	// Ensure we have valid dimensions for centering
	width, height := m.width, m.height
	if width < 40 {
		width = 80
	}
	if height < 20 {
		height = 40
	}

	var lines []string
	lines = append(lines, titleStyle.Render("=== PETRI PROJECT ==="))
	lines = append(lines, "")

	// Check if we're confirming a delete
	if m.confirmingDelete >= 0 && m.confirmingDelete < len(m.worlds) {
		worldName := m.worlds[m.confirmingDelete].Name
		lines = append(lines, fmt.Sprintf("Delete \"%s\"? This cannot be undone.", worldName))
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(
			"Y: Confirm   N: Cancel"))

		content := lipgloss.JoinVertical(lipgloss.Center, lines...)
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}

	if len(m.worlds) == 0 {
		// No existing worlds
		lines = append(lines, "No saved worlds found.")
		lines = append(lines, "")
		if m.selectedWorld == 0 {
			lines = append(lines, highlightStyle.Render("> New World"))
		} else {
			lines = append(lines, "  New World")
		}
	} else {
		// List existing worlds
		for i, world := range m.worlds {
			lastPlayed := formatTimeAgo(world.LastPlayedAt)
			entry := fmt.Sprintf("Continue \"%s\" (%d alive, %s)", world.Name, world.AliveCount, lastPlayed)
			if i == m.selectedWorld {
				lines = append(lines, highlightStyle.Render("> "+entry))
			} else {
				lines = append(lines, "  "+entry)
			}
		}
		// "New World" option at the end
		newWorldIdx := len(m.worlds)
		if m.selectedWorld == newWorldIdx {
			lines = append(lines, highlightStyle.Render("> New World"))
		} else {
			lines = append(lines, "  New World")
		}
	}

	lines = append(lines, "")
	// Show D: Delete hint only when a saved world is selected (not "New World")
	if m.selectedWorld < len(m.worlds) {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(
			"↑/↓ Select   Enter: Continue   D: Delete   Q: Quit"))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(
			"↑/↓ Select   Enter: Continue   Q: Quit"))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// formatTimeAgo formats a time as a human-readable "X ago" string
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// viewModeSelect renders the game mode selection screen
func (m Model) viewModeSelect() string {
	var options string
	if m.testCfg.Debug {
		// Show both options in debug mode
		options = "(1) Single character - customize preferences\n(M) Multiple characters - customize your team"
	} else {
		// Only show multi-character option in normal mode
		options = "(M) Start game - customize your team"
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		titleStyle.Render("=== PETRI PROJECT ==="),
		"",
		options,
		"",
		"Press Q to quit",
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewFoodSelect renders the food selection screen
func (m Model) viewFoodSelect() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		titleStyle.Render("=== PETRI PROJECT ==="),
		"",
		"Choose your character's FAVORITE FOOD:",
		"",
		"(B) Berries  or  (M) Mushrooms",
		"",
		"Press Q to quit",
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewColorSelect renders the color selection screen
func (m Model) viewColorSelect() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		titleStyle.Render("=== PETRI PROJECT ==="),
		"",
		"Choose your character's FAVORITE COLOR:",
		"",
		"(R) Red   (L) Blue   (W) White   (N) Brown",
		"",
		"Press Q to quit",
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewCharacterCreate renders the character creation screen
func (m Model) viewCharacterCreate() string {
	if m.creationState == nil {
		return "Loading..."
	}

	// Styles for character cards
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(22)

	selectedCardStyle := cardStyle.Copy().
		BorderForeground(lipgloss.Color("45")) // cyan highlight

	fieldLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	selectedFieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("45")).
		Bold(true)

	// Build character cards
	var cards []string
	for i, charData := range m.creationState.Characters {
		isSelectedChar := i == m.creationState.SelectedChar

		// Build field displays
		nameLabel := "Name:"
		foodLabel := "Fav Food:"
		colorLabel := "Fav Color:"

		nameValue := charData.Name
		foodValue := charData.Food
		colorValue := charData.Color

		// Apply styles based on selection
		if isSelectedChar {
			switch m.creationState.SelectedField {
			case FieldName:
				nameLabel = selectedFieldStyle.Render("Name:")
				nameValue = selectedFieldStyle.Render(charData.Name + "_")
			case FieldFood:
				foodLabel = selectedFieldStyle.Render("Fav Food:")
				foodValue = selectedFieldStyle.Render("[" + charData.Food + "]")
			case FieldColor:
				colorLabel = selectedFieldStyle.Render("Fav Color:")
				colorValue = selectedFieldStyle.Render("[" + charData.Color + "]")
			}
		} else {
			nameLabel = fieldLabelStyle.Render(nameLabel)
			foodLabel = fieldLabelStyle.Render(foodLabel)
			colorLabel = fieldLabelStyle.Render(colorLabel)
		}

		cardContent := lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("%s %s", nameLabel, nameValue),
			fmt.Sprintf("%s %s", foodLabel, foodValue),
			fmt.Sprintf("%s %s", colorLabel, colorValue),
		)

		style := cardStyle
		if isSelectedChar {
			style = selectedCardStyle
		}

		cards = append(cards, style.Render(cardContent))
	}

	// Join cards horizontally
	cardRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	// Key hints - context aware for space
	var spaceHint string
	if m.creationState.IsNameFieldSelected() {
		spaceHint = ""
	} else {
		spaceHint = "Space: Change option  "
	}

	hints := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render(fmt.Sprintf("← → Navigate characters   ↑ ↓ Navigate fields   Tab: Next field\n%sCtrl+R: Randomize   Enter: Start game   Esc: Back", spaceHint))

	// Build full content
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		titleStyle.Render("=== CHARACTER CREATION ==="),
		"",
		"Customize your characters:",
		"",
		cardRow,
		"",
		hints,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewGame renders the main game view
func (m Model) viewGame() string {
	// Full-screen activity log
	if m.viewMode == viewModeAllActivity && m.activityFullScreen {
		return m.viewFullScreenActivity()
	}

	// Full-screen orders panel
	if m.showOrdersPanel && m.ordersFullScreen {
		return m.viewFullScreenOrders()
	}

	// Build map view
	var mapBuilder strings.Builder
	for y := 0; y < m.gameMap.Height; y++ {
		for x := 0; x < m.gameMap.Width; x++ {
			cell := m.renderCell(x, y)
			mapBuilder.WriteString(cell)
		}
		if y < m.gameMap.Height-1 {
			mapBuilder.WriteRune('\n')
		}
	}

	mapView := borderStyle.Render(mapBuilder.String())

	// Right panel layout depends on view mode
	panelWidth := 52
	totalContentHeight := m.gameMap.Height - 2 // Account for extra borders on right panel

	var rightPanel string
	if m.showOrdersPanel {
		// Orders panel: full-height like All Activity mode
		allActivityHeight := totalContentHeight + 2
		ordersView := borderStyle.Width(panelWidth).Height(allActivityHeight).Render(m.renderOrdersPanel())
		rightPanel = ordersView
	} else if m.viewMode == viewModeAllActivity {
		// All Activity mode: full-height combined log
		// Use full map height so panel bottom aligns with map bottom
		// (Select mode has 2 panels with 2 borders each = 4 border rows total,
		// but All Activity has 1 panel with 2 border rows, so we need +2 content)
		allActivityHeight := totalContentHeight + 2
		logView := borderStyle.Width(panelWidth).Height(allActivityHeight).Render(m.renderCombinedLog())
		rightPanel = logView
	} else {
		// Select mode: Details on top, Action Log/Knowledge/Inventory Panel below
		detailsHeight := totalContentHeight / 2
		logHeight := totalContentHeight - detailsHeight
		detailsView := borderStyle.Width(panelWidth).Height(detailsHeight).Render(m.renderDetails())
		var bottomPanel string
		if m.showKnowledgePanel {
			bottomPanel = borderStyle.Width(panelWidth).Height(logHeight).Render(m.renderKnowledgePanel())
		} else if m.showInventoryPanel {
			bottomPanel = borderStyle.Width(panelWidth).Height(logHeight).Render(m.renderInventoryPanel())
		} else if m.showPreferencesPanel {
			bottomPanel = borderStyle.Width(panelWidth).Height(logHeight).Render(m.renderPreferencesPanel(logHeight))
		} else {
			bottomPanel = borderStyle.Width(panelWidth).Height(logHeight).Render(m.renderActionLog())
		}
		rightPanel = lipgloss.JoinVertical(lipgloss.Left, detailsView, bottomPanel)
	}

	// Horizontal layout: Map | Right Panel
	gameArea := lipgloss.JoinHorizontal(lipgloss.Top, mapView, " ", rightPanel)

	// World time display (120 game seconds = 1 world day)
	worldDay := int(m.elapsedGameTime/120) + 1

	// Status bar with mode-specific hints
	status := "RUNNING"
	stepHint := ""
	if m.paused {
		status = "PAUSED"
		stepHint = " | .=step"
	}

	// Speed indicator and control hints
	speedHint := ""
	speedControls := ""
	if m.speedMultiplier == 2 {
		speedHint = " [½x]"
		speedControls = " | > =fast"
	} else if m.speedMultiplier == 4 {
		speedHint = " [¼x]"
		speedControls = " | > =fast"
	}
	if m.speedMultiplier < 4 {
		speedControls = " | < =slow" + speedControls
	}

	// Saving indicator
	saveHint := ""
	if time.Now().Before(m.saveIndicatorEnd) {
		saveHint = " " + orangeStyle.Render("[Saved]")
	}

	// Build hints for all currently applicable actions
	var hints []string

	// View mode switching (always available)
	if m.viewMode == viewModeAllActivity {
		hints = append(hints, "s=select", "x=expand")
	} else {
		hints = append(hints, "a=all activity")
	}
	hints = append(hints, "b/n=back/next", "o=orders")

	// Cursor movement (not available in orders add/cancel mode)
	inOrdersInput := m.showOrdersPanel && (m.ordersAddMode || m.ordersCancelMode)
	if !inOrdersInput {
		hints = append(hints, "ARROWS=cursor")
	}

	// ESC goes to menu unless it would do something else first
	inSubpanel := m.showKnowledgePanel || m.showInventoryPanel || m.showPreferencesPanel
	if !inOrdersInput && !inSubpanel {
		hints = append(hints, "ESC=menu")
	}

	statusBar := fmt.Sprintf("\nDay %d | [%s]%s%s SPACE=pause%s%s | %s", worldDay, status, speedHint, saveHint, speedControls, stepHint, strings.Join(hints, " | "))

	// Debug line (only shown with -debug flag)
	debugLine := ""
	if m.testCfg.Debug {
		var charInfo []string
		for _, c := range m.gameMap.Characters() {
			pos := c.Pos()
			found := m.gameMap.CharacterAt(pos)
			marker := "✓"
			if found != c {
				marker = "✗"
			}
			charInfo = append(charInfo, fmt.Sprintf("%s(%d,%d)%s", c.Name, pos.X, pos.Y, marker))
		}
		debugLine = fmt.Sprintf("\nChars: %v", charInfo)
	}

	return gameArea + statusBar + debugLine
}

// renderCell renders a single map cell
func (m Model) renderCell(x, y int) string {
	isCursor := x == m.cursorX && y == m.cursorY
	pos := types.Position{X: x, Y: y}

	var sym string
	var fill string // terrain fill for padding (empty = use spaces)

	// Check for character first (takes visual precedence)
	if char := m.gameMap.CharacterAt(pos); char != nil {
		sym = m.styledSymbol(char)
	} else if item := m.gameMap.ItemAt(pos); item != nil {
		sym = m.styledSymbol(item)
	} else if wtype := m.gameMap.WaterAt(pos); wtype != game.WaterNone {
		// Water terrain
		switch wtype {
		case game.WaterSpring:
			sym = waterStyle.Render(string(config.CharSpring))
		case game.WaterPond:
			// Full terrain fill — ▓▓▓ avoids vertical stripe appearance
			waterFill := waterStyle.Render(string(config.CharWater))
			sym = waterFill
			fill = waterFill
		}
	} else if m.gameMap.IsTilled(pos) {
		// Empty tilled tile — full terrain fill (green if wet, olive if dry)
		tilledStyle := growingStyle
		if m.gameMap.IsWet(pos) {
			tilledStyle = greenStyle
		}
		tilledFill := tilledStyle.Render(string(config.CharTilledSoil))
		sym = tilledFill
		fill = tilledFill
	} else if feature := m.gameMap.FeatureAt(pos); feature != nil {
		sym = m.styledSymbol(feature)
	} else {
		sym = " "
	}

	// Entities on tilled soil get terrain fill padding (green if wet, olive if dry)
	if fill == "" && m.gameMap.IsTilled(pos) {
		tilledStyle := growingStyle
		if m.gameMap.IsWet(pos) {
			tilledStyle = greenStyle
		}
		fill = tilledStyle.Render(string(config.CharTilledSoil))
	}

	// Area selection highlighting (only visible during tillSoil step 2)
	if m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID == "tillSoil" {
		// Rectangle highlight when anchor is set
		if m.areaSelectAnchor != nil && !isCursor {
			cursor := types.Position{X: m.cursorX, Y: m.cursorY}
			if isInRect(pos, *m.areaSelectAnchor, cursor) {
				var validator func(types.Position, *game.Map) bool
				var bgStyle lipgloss.Style
				if m.areaSelectUnmarkMode {
					validator = isValidUnmarkTarget
					bgStyle = areaUnselectStyle
				} else {
					validator = isValidTillTarget
					bgStyle = areaSelectStyle
				}
				if validator(pos, m.gameMap) {
					hasEntity := m.gameMap.CharacterAt(pos) != nil || m.gameMap.ItemAt(pos) != nil || m.gameMap.FeatureAt(pos) != nil
					if hasEntity {
						bg := bgStyle.Render(" ")
						return bg + sym + bg
					}
					padded := " " + sym + " "
					if fill != "" {
						padded = fill + sym + fill
					}
					return bgStyle.Render(padded)
				}
			}
		}

		// Pre-highlight existing marked-for-tilling tiles
		if m.gameMap.IsMarkedForTilling(pos) && !isCursor {
			hasEntity := m.gameMap.CharacterAt(pos) != nil || m.gameMap.ItemAt(pos) != nil || m.gameMap.FeatureAt(pos) != nil
			if hasEntity {
				bg := areaSelectStyle.Render(" ")
				return bg + sym + bg
			}
			padded := " " + sym + " "
			if fill != "" {
				padded = fill + sym + fill
			}
			return areaSelectStyle.Render(padded)
		}
	}

	if isCursor {
		return "[" + sym + "]"
	}
	if fill != "" {
		return fill + sym + fill
	}
	return " " + sym + " "
}

// styledSymbol returns a colored symbol for an entity
func (m Model) styledSymbol(e entity.Entity) string {
	sym := string(e.Symbol())

	switch v := e.(type) {
	case *entity.Character:
		if v.IsDead {
			return deadStyle.Render(sym)
		}

		// Collect active status symbols (sleeping, frustrated, in crisis)
		type statusDisplay struct {
			symbol string
			style  lipgloss.Style
		}
		var statuses []statusDisplay

		if v.IsSleeping {
			statuses = append(statuses, statusDisplay{"z", sleepingStyle})
		}
		if v.IsFrustrated {
			statuses = append(statuses, statusDisplay{"?", frustratedStyle})
		}
		if v.IsInCrisis() {
			statuses = append(statuses, statusDisplay{"!", crisisStyle})
		}

		// No status symbols to flash - show @ (green if poisoned)
		if len(statuses) == 0 {
			if v.Poisoned {
				return poisonedStyle.Render(sym)
			}
			return sym
		}

		// One or more statuses - flash between @ and status symbols
		// Build cycle: @ first (green if poisoned), then each status symbol
		atStyle := lipgloss.NewStyle()
		if v.Poisoned {
			atStyle = poisonedStyle
		}

		allSymbols := make([]statusDisplay, 0, len(statuses)+1)
		allSymbols = append(allSymbols, statusDisplay{sym, atStyle})
		allSymbols = append(allSymbols, statuses...)

		idx := m.flashIndex % len(allSymbols)
		return allSymbols[idx].style.Render(allSymbols[idx].symbol)

	case *entity.Item:
		// Sprout rendering: sage for most, green on wet ground, variety color for mushrooms
		if v.Plant != nil && v.Plant.IsSprout && v.ItemType != "mushroom" {
			if m.gameMap.IsWet(v.Pos()) {
				return greenStyle.Render(sym)
			}
			return sproutStyle.Render(sym)
		}
		switch v.Color {
		case types.ColorRed:
			return redStyle.Render(sym)
		case types.ColorBlue:
			return blueStyle.Render(sym)
		case types.ColorBrown:
			return brownStyle.Render(sym)
		case types.ColorWhite:
			return whiteStyle.Render(sym)
		case types.ColorOrange:
			return orangeStyle.Render(sym)
		case types.ColorYellow:
			return yellowStyle.Render(sym)
		case types.ColorPurple:
			return purpleStyle.Render(sym)
		case types.ColorTan:
			return tanStyle.Render(sym)
		case types.ColorPink:
			return pinkStyle.Render(sym)
		case types.ColorBlack:
			return blackStyle.Render(sym)
		case types.ColorGreen:
			return greenStyle.Render(sym)
		case types.ColorPalePink:
			return palePinkStyle.Render(sym)
		case types.ColorPaleYellow:
			return paleYellowStyle.Render(sym)
		case types.ColorSilver:
			return silverStyle.Render(sym)
		case types.ColorGray:
			return grayStyle.Render(sym)
		case types.ColorLavender:
			return lavenderStyle.Render(sym)
		}

	case *entity.Feature:
		if v.IsBed() {
			return leafStyle.Render(sym)
		}
	}

	return sym
}

// renderDetails renders the details sidebar
func (m Model) renderDetails() string {
	var lines []string
	lines = append(lines, titleStyle.Render("       DETAILS"), "")

	// Check for character first, then item, then feature
	cursorPos := types.Position{X: m.cursorX, Y: m.cursorY}
	e := m.gameMap.EntityAt(cursorPos)
	item := m.gameMap.ItemAt(cursorPos)
	feature := m.gameMap.FeatureAt(cursorPos)

	waterType := m.gameMap.WaterAt(cursorPos)

	if e == nil && item == nil && feature == nil && waterType == game.WaterNone {
		if m.gameMap.IsTilled(cursorPos) {
			lines = append(lines, " Type: "+growingStyle.Render("Tilled soil"))
		} else if m.gameMap.IsMarkedForTilling(cursorPos) {
			lines = append(lines, " Type: "+growingStyle.Render("Marked for tilling"))
		} else {
			lines = append(lines, " Type: Empty")
		}
		if m.gameMap.IsManuallyWatered(cursorPos) {
			label := "Watered"
			if m.testCfg.Debug {
				label = fmt.Sprintf("Watered (%.0fs)", m.gameMap.WateredTimer(cursorPos))
			}
			lines = append(lines, " "+waterStyle.Render(label))
		} else if m.gameMap.IsWet(cursorPos) {
			lines = append(lines, " "+waterStyle.Render("Wet"))
		}
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
	} else if char, ok := e.(*entity.Character); ok {
		// Show name with edit UI if editing this character
		if m.editingCharacterName && m.editingCharacterID == char.ID {
			lines = append(lines, fmt.Sprintf(" Name: %s_", m.editingNameBuffer))
			lines = append(lines, " [Enter=save, Esc=cancel]")
		} else {
			lines = append(lines, fmt.Sprintf(" Name: %s", char.Name))
		}
		lines = append(lines, " Type: Character")
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}

		// Stats with tier-based coloring
		healthLevel := colorByTier(char.HealthLevel(), char.HealthTier())
		hungerLevel := colorByTier(char.HungerLevel(), char.HungerTier())
		thirstLevel := colorByTier(char.ThirstLevel(), char.ThirstTier())
		energyLevel := colorByTier(char.EnergyLevel(), char.EnergyTier())
		moodLevel := colorByTier(char.MoodLevel(), char.MoodTier())

		if m.testCfg.Debug {
			lines = append(lines, "",
				fmt.Sprintf(" Health: %d/100 (%s)", int(char.Health), healthLevel),
				fmt.Sprintf(" Hunger: %d/100 (%s)", int(char.Hunger), hungerLevel),
				fmt.Sprintf(" Thirst: %d/100 (%s)", int(char.Thirst), thirstLevel),
				fmt.Sprintf(" Energy: %d/100 (%s)", int(char.Energy), energyLevel),
				fmt.Sprintf(" Mood: %d/100 (%s)", int(char.Mood), moodLevel),
			)
		} else {
			lines = append(lines, "",
				fmt.Sprintf(" Health: %s", healthLevel),
				fmt.Sprintf(" Hunger: %s", hungerLevel),
				fmt.Sprintf(" Thirst: %s", thirstLevel),
				fmt.Sprintf(" Energy: %s", energyLevel),
				fmt.Sprintf(" Mood: %s", moodLevel),
			)
		}

		// Speed info (debug only)
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Speed: %d/100", char.EffectiveSpeed()))
		}

		// Status with color - supports multiple simultaneous statuses
		var statusLine string
		if char.IsDead {
			statusLine = " Status: " + deadStyle.Render("DEAD")
		} else {
			var statusParts []string

			if char.IsSleeping {
				location := "ground"
				if char.AtBed {
					location = "bed"
				}
				statusParts = append(statusParts, sleepingStyle.Render(fmt.Sprintf("SLEEPING (%s)", location)))
			}
			if char.IsFrustrated {
				statusParts = append(statusParts, frustratedStyle.Render(fmt.Sprintf("FRUSTRATED (%.0fs)", char.FrustrationTimer)))
			}
			if char.Poisoned {
				statusParts = append(statusParts, poisonedStyle.Render(fmt.Sprintf("POISONED (%.0fs)", char.PoisonTimer)))
			}
			if char.IsInCrisis() {
				statusParts = append(statusParts, crisisStyle.Render("IN CRISIS"))
			}

			if len(statusParts) == 0 {
				statusLine = " Status: -"
			} else {
				statusLine = " Status: " + strings.Join(statusParts, ", ")
			}
		}
		lines = append(lines, statusLine)

		// Activity with optional debug progress
		activityLine := fmt.Sprintf(" Activity: %s", char.CurrentActivity)
		if m.testCfg.Debug && char.ActionProgress > 0 {
			activityLine = fmt.Sprintf(" Activity: %s (%.1fs)", char.CurrentActivity, char.ActionProgress)
		}
		lines = append(lines, activityLine)

		// Show assigned order if any
		if char.AssignedOrderID != 0 {
			if order := m.findOrderByID(char.AssignedOrderID); order != nil {
				orderLine := fmt.Sprintf(" Order: %s [%s]", order.DisplayName(), order.StatusDisplay())
				lines = append(lines, orderLine)
			}
		}

		if m.following == char {
			lines = append(lines, "", highlightStyle.Render(" [FOLLOWING]"), " Press F to unfollow")
		} else {
			lines = append(lines, "", " Press F to follow")
		}
		lines = append(lines, " P: Preferences  K: Knowledge  I: Inventory")
		if !m.editingCharacterName {
			lines = append(lines, " Press E to edit name")
		}

	} else if item != nil {
		if item.Plant != nil && item.Plant.IsSprout {
			lines = append(lines, " Type: "+growingStyle.Render("Sprout"))
		} else {
			lines = append(lines, " Type: Item")
		}
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
		// Show Name for crafted items
		if item.Name != "" {
			lines = append(lines, fmt.Sprintf(" Name: %s", item.Name))
		}
		kindLabel := item.ItemType
		if item.Plant != nil && item.Plant.IsSprout {
			kindLabel += " sprout"
		}
		lines = append(lines,
			fmt.Sprintf(" Kind: %s", kindLabel),
			fmt.Sprintf(" Color: %s", item.Color),
		)
		// Show Pattern/Texture if item has them (works for both natural and crafted items)
		if item.Pattern != "" {
			lines = append(lines, fmt.Sprintf(" Pattern: %s", item.Pattern))
		}
		if item.Texture != "" {
			lines = append(lines, fmt.Sprintf(" Texture: %s", item.Texture))
		}
		// Show Poisonous/Healing only for edible items
		if item.IsEdible() {
			poison := "No"
			if item.IsPoisonous() {
				poison = redStyle.Render("Yes")
			}
			healing := "No"
			if item.IsHealing() {
				healing = optimalStyle.Render("Yes")
			}
			lines = append(lines,
				fmt.Sprintf(" Poisonous: %s", poison),
				fmt.Sprintf(" Healing: %s", healing),
			)
		}
		// Show Growing status for plants that can spread
		if item.Plant != nil && item.Plant.IsGrowing {
			lines = append(lines, " "+growingStyle.Render("Growing"))
		}
		// Show Plantable status
		if item.Plantable {
			lines = append(lines, " "+growingStyle.Render("Plantable"))
		}
		// Show vessel contents if this is a container
		if item.Container != nil {
			lines = append(lines, "")
			if len(item.Container.Contents) == 0 {
				lines = append(lines, " Contents: (empty)")
			} else {
				lines = append(lines, " Contents:")
				for _, stack := range item.Container.Contents {
					if stack.Variety != nil {
						stackSize := config.GetStackSize(stack.Variety.ItemType)
						contentLine := fmt.Sprintf("   %s: %d/%d",
							stack.Variety.Description(), stack.Count, stackSize)
						if stack.Variety.ItemType == "liquid" {
							contentLine = "   " + waterStyle.Render(fmt.Sprintf("%s: %d/%d",
								stack.Variety.Description(), stack.Count, stackSize))
						}
						lines = append(lines, contentLine)
					}
				}
			}
		}
		// Show tilled/marked soil annotation
		if m.gameMap.IsTilled(cursorPos) {
			lines = append(lines, " "+growingStyle.Render("On tilled soil"))
		} else if m.gameMap.IsMarkedForTilling(cursorPos) {
			lines = append(lines, " "+growingStyle.Render("Marked for tilling"))
		}
		if m.gameMap.IsManuallyWatered(cursorPos) {
			label := "Watered"
			if m.testCfg.Debug {
				label = fmt.Sprintf("Watered (%.0fs)", m.gameMap.WateredTimer(cursorPos))
			}
			lines = append(lines, " "+waterStyle.Render(label))
		} else if m.gameMap.IsWet(cursorPos) {
			lines = append(lines, " "+waterStyle.Render("Wet"))
		}
	} else if waterType != game.WaterNone {
		lines = append(lines, " Type: Water")
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
		switch waterType {
		case game.WaterSpring:
			lines = append(lines, " Kind: spring")
		case game.WaterPond:
			lines = append(lines, " Kind: pond")
		}
		lines = append(lines, " Use: Drinking")
	} else if feature != nil {
		lines = append(lines, " Type: Feature")
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
		lines = append(lines, fmt.Sprintf(" Kind: %s", feature.Description()))
		if feature.IsBed() {
			lines = append(lines, " Use: Sleeping")
		}
	}

	return strings.Join(lines, "\n")
}

// renderActionLog renders the action log sidebar
func (m Model) renderActionLog() string {
	var lines []string
	lines = append(lines, titleStyle.Render("       ACTION LOG"), "")

	e := m.gameMap.EntityAt(types.Position{X: m.cursorX, Y: m.cursorY})
	char, isChar := e.(*entity.Character)

	if !isChar {
		lines = append(lines, " Select character to", " view action log")
		return strings.Join(lines, "\n")
	}

	events := m.actionLog.Events(char.ID, 200)

	// Pre-filter debug-only messages in non-debug mode
	if !m.testCfg.Debug {
		var filtered []system.Event
		for _, event := range events {
			if strings.Contains(event.Message, "Improved Mood") ||
				strings.Contains(event.Message, "Worsened Mood") ||
				strings.Contains(event.Message, "impacted health") ||
				event.Message == "Collapsed from exhaustion!" {
				continue
			}
			filtered = append(filtered, event)
		}
		events = filtered
	}

	if len(events) == 0 {
		lines = append(lines, " No events yet")
		return strings.Join(lines, "\n")
	}

	// Calculate display range (panel height = (mapHeight-2)/2, minus header lines)
	logHeight := (m.gameMap.Height - 2) - (m.gameMap.Height-2)/2
	maxDisplay := logHeight - 3 // Account for header and borders
	total := len(events)

	// Apply scroll offset
	endIdx := total - m.logScrollOffset
	if endIdx < 0 {
		endIdx = 0
	}
	startIdx := endIdx - maxDisplay
	if startIdx < 0 {
		startIdx = 0
	}

	if m.logScrollOffset > 0 {
		lines = append(lines, fmt.Sprintf(" [Scrolled: -%d]", m.logScrollOffset))
	}

	// Format and display events (most recent first)
	displayEvents := events[startIdx:endIdx]

	for i := len(displayEvents) - 1; i >= 0; i-- {
		event := displayEvents[i]
		elapsedSecs := m.elapsedGameTime - event.GameTime

		// Strip numeric details from messages in non-debug mode
		message := event.Message
		if !m.testCfg.Debug {
			message = parenPattern.ReplaceAllString(message, "")
			message = healthPattern.ReplaceAllString(message, "")
		}

		var line string
		if elapsedSecs >= 60 {
			line = fmt.Sprintf(" %s %s", system.FormatGameTime(elapsedSecs), message)
		} else {
			line = " " + message
		}

		// Truncate if too long (use runes for proper Unicode handling)
		runes := []rune(line)
		if len(runes) > 50 {
			line = string(runes[:49]) + "…"
		}

		// Color based on message content
		coloredLine := colorLogMessage(line, message)

		// Highlight most recent (only if not already colored)
		isNewest := (i == len(displayEvents)-1) && m.logScrollOffset == 0
		if coloredLine != line {
			lines = append(lines, coloredLine)
		} else if isNewest {
			lines = append(lines, highlightStyle.Render(line))
		} else {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// renderCombinedLog renders the combined activity log for all characters
func (m Model) renderCombinedLog() string {
	var lines []string
	lines = append(lines, titleStyle.Render("     ALL ACTIVITY"), "")
	lines = append(lines, m.renderActivityContent(false)...)
	return strings.Join(lines, "\n")
}

// viewFullScreenActivity renders full-screen activity log
func (m Model) viewFullScreenActivity() string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("                  ALL ACTIVITY"))
	lines = append(lines, "")
	lines = append(lines, m.renderActivityContent(true)...)

	// Footer with controls
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(
		"PgUp/PgDn: Scroll | X: Collapse | S: Select Mode"))

	content := strings.Join(lines, "\n")
	return borderStyle.Width(m.width - 4).Height(m.height - 4).Render(content)
}

// renderActivityContent generates the activity log content lines
// expanded: true for full-screen (no truncation), false for side panel (truncated)
func (m Model) renderActivityContent(expanded bool) []string {
	var lines []string

	events := m.actionLog.AllEvents(200)

	// Pre-filter debug-only messages in non-debug mode
	if !m.testCfg.Debug {
		var filtered []system.Event
		for _, event := range events {
			if strings.Contains(event.Message, "Improved Mood") ||
				strings.Contains(event.Message, "Worsened Mood") ||
				strings.Contains(event.Message, "impacted health") ||
				event.Message == "Collapsed from exhaustion!" {
				continue
			}
			filtered = append(filtered, event)
		}
		events = filtered
	}

	if len(events) == 0 {
		lines = append(lines, " No events yet")
		return lines
	}

	// Calculate display range based on mode
	var maxDisplay int
	if expanded {
		// Full screen: use most of the screen height
		maxDisplay = m.height - 8 // Account for header, footer, borders
	} else {
		// Side panel: use map height
		allActivityHeight := m.gameMap.Height
		maxDisplay = allActivityHeight - 3 // Account for header and borders
	}

	total := len(events)

	// Apply scroll offset
	endIdx := total - m.logScrollOffset
	if endIdx < 0 {
		endIdx = 0
	}
	startIdx := endIdx - maxDisplay
	if startIdx < 0 {
		startIdx = 0
	}

	if m.logScrollOffset > 0 {
		lines = append(lines, fmt.Sprintf(" [Scrolled: -%d]", m.logScrollOffset))
	}

	// Format and display events (most recent first)
	displayEvents := events[startIdx:endIdx]

	for i := len(displayEvents) - 1; i >= 0; i-- {
		event := displayEvents[i]
		elapsedSecs := m.elapsedGameTime - event.GameTime

		// Strip numeric details from messages in non-debug mode
		message := event.Message
		if !m.testCfg.Debug {
			message = parenPattern.ReplaceAllString(message, "")
			message = healthPattern.ReplaceAllString(message, "")
		}

		// Format with character name prefix
		var line string
		namePrefix := fmt.Sprintf("[%s] ", event.CharName)
		if elapsedSecs >= 60 {
			line = fmt.Sprintf(" %s%s%s", system.FormatGameTime(elapsedSecs), namePrefix, message)
		} else {
			line = " " + namePrefix + message
		}

		// Truncate if too long in collapsed mode (use runes for proper Unicode handling)
		if !expanded {
			runes := []rune(line)
			if len(runes) > 50 {
				line = string(runes[:49]) + "…"
			}
		}

		// Color based on message content
		coloredLine := colorLogMessage(line, message)

		// Highlight most recent (only if not already colored)
		isNewest := (i == len(displayEvents)-1) && m.logScrollOffset == 0
		if coloredLine != line {
			lines = append(lines, coloredLine)
		} else if isNewest {
			lines = append(lines, highlightStyle.Render(line))
		} else {
			lines = append(lines, line)
		}
	}

	return lines
}

// renderKnowledgePanel renders the knowledge panel for the selected character
func (m Model) renderKnowledgePanel() string {
	var lines []string
	lines = append(lines, titleStyle.Render("       KNOWLEDGE"), "")

	// Get character at cursor
	if e := m.gameMap.EntityAt(types.Position{X: m.cursorX, Y: m.cursorY}); e != nil {
		if char, ok := e.(*entity.Character); ok {
			hasKnowHow := len(char.KnownActivities) > 0
			hasFacts := len(char.Knowledge) > 0
			hasRecipes := len(char.KnownRecipes) > 0

			if !hasKnowHow && !hasFacts && !hasRecipes {
				// Empty state - just show title, panel appears empty
				return strings.Join(lines, "\n")
			}

			// Show facts section
			if hasFacts {
				lines = append(lines, " Facts:")
				for _, k := range char.Knowledge {
					lines = append(lines, "   "+k.Description())
				}
				if hasKnowHow || hasRecipes {
					lines = append(lines, "") // spacing between sections
				}
			}

			// Show know-how section
			if hasKnowHow {
				lines = append(lines, " Knows how to:")
				for _, activityID := range char.KnownActivities {
					if activity, ok := entity.ActivityRegistry[activityID]; ok {
						if activity.Category != "" {
							if display, ok := categoryDisplayName[activity.Category]; ok {
								lines = append(lines, "   "+display+": "+activity.Name)
							} else {
								lines = append(lines, "   "+activity.Category+": "+activity.Name)
							}
						} else {
							lines = append(lines, "   "+activity.Name)
						}
					}
				}
				if hasRecipes {
					lines = append(lines, "") // spacing between sections
				}
			}

			// Show recipes section
			if hasRecipes {
				lines = append(lines, " Recipes:")
				for _, recipeID := range char.KnownRecipes {
					if recipe, ok := entity.RecipeRegistry[recipeID]; ok {
						lines = append(lines, "   "+recipe.Name)
					}
				}
			}
		} else {
			lines = append(lines, " Select a character")
		}
	} else {
		lines = append(lines, " Select a character")
	}

	lines = append(lines, "", " K or Esc to return")

	return strings.Join(lines, "\n")
}

// renderInventoryPanel renders the inventory panel for the selected character
func (m Model) renderInventoryPanel() string {
	var lines []string
	lines = append(lines, titleStyle.Render("       INVENTORY"), "")

	// Get character at cursor
	if e := m.gameMap.EntityAt(types.Position{X: m.cursorX, Y: m.cursorY}); e != nil {
		if char, ok := e.(*entity.Character); ok {
			if len(char.Inventory) == 0 {
				lines = append(lines, " Inventory: empty")
			} else {
				lines = append(lines, fmt.Sprintf(" Inventory: %d/%d slots", len(char.Inventory), entity.InventoryCapacity))
				for i, item := range char.Inventory {
					label := item.Description()
					if attrs := inventoryItemAttrs(item); attrs != "" {
						label += " (" + attrs + ")"
					}
					lines = append(lines, fmt.Sprintf("  [%d] %s", i+1, label))

					// Show vessel contents if this item is a vessel
					if item.Container != nil {
						if len(item.Container.Contents) == 0 {
							lines = append(lines, "      (empty)")
						} else {
							for _, stack := range item.Container.Contents {
								if stack.Variety != nil {
									stackSize := config.GetStackSize(stack.Variety.ItemType)
									contentLine := fmt.Sprintf("      %s: %d/%d",
										stack.Variety.Description(), stack.Count, stackSize)
									if stack.Variety.ItemType == "liquid" {
										contentLine = "      " + waterStyle.Render(fmt.Sprintf("%s: %d/%d",
											stack.Variety.Description(), stack.Count, stackSize))
									}
									lines = append(lines, contentLine)
								}
							}
						}
					}
				}
			}
		} else {
			lines = append(lines, " Select a character")
		}
	} else {
		lines = append(lines, " Select a character")
	}

	lines = append(lines, "", " I or Esc to return")

	return strings.Join(lines, "\n")
}

// inventoryItemAttrs returns a parenthetical summary of boolean item attributes.
// Returns empty string if no attributes are true.
func inventoryItemAttrs(item *entity.Item) string {
	var parts []string
	if item.IsEdible() {
		parts = append(parts, "Edible")
	}
	if item.IsPoisonous() {
		parts = append(parts, "Poison")
	}
	if item.IsHealing() {
		parts = append(parts, "Healing")
	}
	if item.Plantable {
		parts = append(parts, "Plantable")
	}
	return strings.Join(parts, ", ")
}

// renderPreferencesPanel renders the preferences panel for the selected character
func (m Model) renderPreferencesPanel(panelHeight int) string {
	var lines []string
	lines = append(lines, titleStyle.Render("      PREFERENCES"), "")

	// Get character at cursor
	if e := m.gameMap.EntityAt(types.Position{X: m.cursorX, Y: m.cursorY}); e != nil {
		if char, ok := e.(*entity.Character); ok {
			if len(char.Preferences) == 0 {
				lines = append(lines, " No preferences yet")
			} else {
				// Calculate display range for scrolling
				maxDisplay := panelHeight - 5 // Account for header, footer, borders
				total := len(char.Preferences)

				// Apply scroll offset
				endIdx := total - m.logScrollOffset
				if endIdx < 0 {
					endIdx = 0
				}
				startIdx := endIdx - maxDisplay
				if startIdx < 0 {
					startIdx = 0
				}
				if endIdx > total {
					endIdx = total
				}

				// Show scroll indicator if scrolled
				if m.logScrollOffset > 0 {
					lines = append(lines, fmt.Sprintf(" [Scrolled: -%d]", m.logScrollOffset))
				}

				// Display preferences in range
				for i := startIdx; i < endIdx; i++ {
					pref := char.Preferences[i]
					verb := "Likes"
					if !pref.IsPositive() {
						verb = "Dislikes"
					}
					line := fmt.Sprintf("   %s %s", verb, pref.Description())
					// Color based on valence
					if pref.IsPositive() {
						lines = append(lines, optimalStyle.Render(line))
					} else {
						lines = append(lines, severeStyle.Render(line))
					}
				}

				// Show hint if more items exist
				if total > maxDisplay {
					lines = append(lines, fmt.Sprintf(" (%d total, PgUp/PgDn to scroll)", total))
				}
			}
		} else {
			lines = append(lines, " Select a character")
		}
	} else {
		lines = append(lines, " Select a character")
	}

	lines = append(lines, "", " P or Esc to return")

	return strings.Join(lines, "\n")
}

// renderOrdersContent generates the orders panel content lines
// expanded: true for full-screen, false for side panel
func (m Model) renderOrdersContent(expanded bool) []string {
	var lines []string

	// Indentation varies by mode
	indent := " "
	selectIndent := "   "
	selectPrefix := " > "
	if expanded {
		indent = "  "
		selectIndent = "    "
		selectPrefix = "  > "
	}

	// Get orderable activities (known by at least one living character)
	orderableActivities := m.getOrderableActivities()

	// Add hints at top so they're always visible
	if m.ordersAddMode && m.ordersAddStep == 2 && m.step2ActivityID == "tillSoil" {
		// Area selection hints
		modeName := "Mark"
		if m.areaSelectUnmarkMode {
			modeName = "Unmark"
		}
		lines = append(lines, indent+growingStyle.Render("Till Soil: "+modeName), "")
		if m.areaSelectAnchor == nil {
			lines = append(lines, indent+"arrows: move cursor")
			lines = append(lines, indent+"p: set anchor")
		} else {
			lines = append(lines, indent+"arrows: resize")
			lines = append(lines, indent+"p: confirm plot")
		}
		lines = append(lines, indent+"tab: toggle mark/unmark")
		lines = append(lines, indent+"enter: done  esc: cancel")
		lines = append(lines, "")
	} else if m.ordersAddMode || m.ordersCancelMode {
		lines = append(lines, indent+"enter: confirm  esc: back", "")
	} else {
		var hints []string
		if len(orderableActivities) > 0 {
			hints = append(hints, "+: add")
		}
		if len(m.orders) > 0 {
			hints = append(hints, "c: cancel")
		}
		if expanded {
			hints = append(hints, "x: collapse")
		} else {
			hints = append(hints, "x: expand")
		}
		hints = append(hints, "o: close")
		lines = append(lines, indent+strings.Join(hints, "  "), "")
	}

	if m.ordersAddMode {
		// Add order flow
		if m.ordersAddStep == 0 {
			// Step 0: Select activity
			lines = append(lines, indent+"Select activity:", "")
			if len(orderableActivities) == 0 {
				lines = append(lines, selectIndent+"(no activities available)")
				lines = append(lines, selectIndent+"Characters must discover")
				lines = append(lines, selectIndent+"know-how first.")
			} else {
				for i, activity := range orderableActivities {
					prefix := selectIndent
					if i == m.selectedActivityIndex {
						prefix = selectPrefix
					}
					lines = append(lines, prefix+activity.Name)
				}
			}
		} else if m.ordersAddStep == 2 {
			if m.step2ActivityID == "plant" {
				// Step 2 (plant): show plantable type selection
				plantTypes := game.GetPlantableTypes()
				lines = append(lines, indent+"Select type to plant:", "")
				for i, pt := range plantTypes {
					prefix := selectIndent
					if i == m.selectedPlantTypeIndex {
						prefix = selectPrefix
					}
					lines = append(lines, prefix+pt.DisplayName)
				}
			}
			// tillSoil: hints already rendered above, nothing else needed
		} else {
			// Step 1: Select sub-item based on selected category/activity
			if m.selectedActivityIndex < len(orderableActivities) &&
				isSyntheticCategory(orderableActivities[m.selectedActivityIndex].ID) {
				category := syntheticCategoryID(orderableActivities[m.selectedActivityIndex].ID)
				categoryActivities := m.getCategoryActivities(category)
				prompt := "Select activity:"
				if category == "craft" {
					prompt = "Select item to craft:"
				}
				lines = append(lines, indent+prompt, "")
				for i, activity := range categoryActivities {
					prefix := selectIndent
					if i == m.selectedTargetIndex {
						prefix = selectPrefix
					}
					lines = append(lines, prefix+activity.Name)
				}
			} else {
				// Harvest selected - show item types
				edibleTypes := m.getEdibleItemTypes()
				lines = append(lines, indent+"Select item type:", "")
				for i, itemType := range edibleTypes {
					prefix := selectIndent
					if i == m.selectedTargetIndex {
						prefix = selectPrefix
					}
					lines = append(lines, prefix+itemType)
				}
			}
		}
	} else if m.ordersCancelMode {
		// Cancel mode - show orders with selection
		lines = append(lines, indent+"Select order to cancel:", "")
		if len(m.orders) == 0 {
			lines = append(lines, selectIndent+"(no orders)")
		} else {
			// Build character lookup map for assigned names
			charByID := make(map[int]*entity.Character)
			for _, c := range m.gameMap.Characters() {
				charByID[c.ID] = c
			}

			items := m.gameMap.Items()
			for i, order := range m.orders {
				prefix := selectIndent
				if i == m.selectedOrderIndex {
					prefix = selectPrefix
				}
				statusStr := order.StatusDisplay()
				if order.AssignedTo != 0 {
					if char, ok := charByID[order.AssignedTo]; ok {
						statusStr = fmt.Sprintf("%s: %s", order.StatusDisplay(), char.Name)
					}
				}
				// Check feasibility for unfulfillable display
				feasible, noKnowHow := system.IsOrderFeasible(order, items, m.gameMap)
				if !feasible {
					if noKnowHow {
						statusStr = "No one knows how"
					} else {
						statusStr = "Unfulfillable"
					}
					lines = append(lines, unfulfillableStyle.Render(fmt.Sprintf("%s%s [%s]", prefix, order.DisplayName(), statusStr)))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s [%s]", prefix, order.DisplayName(), statusStr))
				}
			}
		}
	} else {
		// Normal view - show order list
		if len(m.orders) == 0 {
			lines = append(lines, indent+"No orders.")
			lines = append(lines, "")
			if len(orderableActivities) > 0 {
				lines = append(lines, indent+"Press + to add an order.")
			} else {
				lines = append(lines, indent+"Characters must discover")
				lines = append(lines, indent+"know-how before orders")
				lines = append(lines, indent+"can be created.")
			}
		} else {
			// Build character lookup map for assigned names
			charByID := make(map[int]*entity.Character)
			for _, c := range m.gameMap.Characters() {
				charByID[c.ID] = c
			}

			items := m.gameMap.Items()
			for _, order := range m.orders {
				statusStr := order.StatusDisplay()
				if order.AssignedTo != 0 {
					if char, ok := charByID[order.AssignedTo]; ok {
						statusStr = fmt.Sprintf("%s: %s", order.StatusDisplay(), char.Name)
					}
				}
				// Check feasibility for unfulfillable display
				feasible, noKnowHow := system.IsOrderFeasible(order, items, m.gameMap)
				if !feasible {
					if noKnowHow {
						statusStr = "No one knows how"
					} else {
						statusStr = "Unfulfillable"
					}
					lines = append(lines, unfulfillableStyle.Render(fmt.Sprintf("%s%s [%s]", indent, order.DisplayName(), statusStr)))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s [%s]", indent, order.DisplayName(), statusStr))
				}
			}
		}
	}

	// Order creation flash confirmation
	if time.Now().Before(m.orderFlashEnd) {
		flash := indent + "+ " + m.orderFlashMessage + " added"
		if m.orderFlashCount > 1 {
			flash += fmt.Sprintf(" (x%d)", m.orderFlashCount)
		}
		lines = append(lines, "", orderStyle.Render(flash))
	}

	return lines
}

// renderOrdersPanel renders the collapsed orders panel
func (m Model) renderOrdersPanel() string {
	var lines []string
	lines = append(lines, titleStyle.Render("         ORDERS"), "")
	lines = append(lines, m.renderOrdersContent(false)...)
	return strings.Join(lines, "\n")
}

// categoryDisplayName maps category IDs to their display names in the order UI
var categoryDisplayName = map[string]string{
	"craft":  "Craft",
	"garden": "Garden",
}

// getOrderableActivities returns activities that can be ordered
// (known by at least one living character)
// Returns uncategorized activities first, then synthetic category entries
func (m Model) getOrderableActivities() []entity.Activity {
	var result []entity.Activity
	chars := m.gameMap.Characters()
	knownCategories := map[string]bool{}

	for _, activity := range entity.ActivityRegistry {
		if activity.Availability != entity.AvailabilityKnowHow {
			continue // Only knowhow activities are orderable
		}
		if activity.IntentFormation != entity.IntentOrderable {
			continue // Only orderable activities
		}
		// Check if any living character knows this activity
		for _, char := range chars {
			if !char.IsDead && char.KnowsActivity(activity.ID) {
				if activity.Category != "" {
					knownCategories[activity.Category] = true
				} else {
					result = append(result, activity)
				}
				break
			}
		}
	}

	// Add synthetic category entries for categories with known activities
	for category := range knownCategories {
		name := category
		if display, ok := categoryDisplayName[category]; ok {
			name = display
		}
		result = append(result, entity.Activity{
			ID:       "category:" + category,
			Name:     name,
			Category: category,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// isSyntheticCategory returns true if this is a synthetic category entry (not a real activity)
func isSyntheticCategory(activityID string) bool {
	return len(activityID) > 9 && activityID[:9] == "category:"
}

// syntheticCategoryID returns the category from a synthetic category activity ID
func syntheticCategoryID(activityID string) string {
	if isSyntheticCategory(activityID) {
		return activityID[9:]
	}
	return ""
}

// getCategoryActivities returns available activities in a category that at least one character knows
func (m Model) getCategoryActivities(category string) []entity.Activity {
	var result []entity.Activity
	chars := m.gameMap.Characters()

	for _, activity := range entity.ActivityRegistry {
		if activity.Category != category {
			continue
		}
		if activity.Availability != entity.AvailabilityKnowHow {
			continue
		}
		if activity.IntentFormation != entity.IntentOrderable {
			continue
		}
		// Check if any living character knows this activity
		for _, char := range chars {
			if !char.IsDead && char.KnowsActivity(activity.ID) {
				result = append(result, activity)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// getEdibleItemTypes returns a list of edible item type names
func (m Model) getEdibleItemTypes() []string {
	configs := game.GetItemTypeConfigs()
	var result []string
	for itemType, cfg := range configs {
		if cfg.Edible {
			result = append(result, itemType)
		}
	}
	// Sort for consistent ordering
	sort.Strings(result)
	return result
}

// findOrderByID returns the order with the given ID, or nil if not found
func (m Model) findOrderByID(id int) *entity.Order {
	for _, order := range m.orders {
		if order.ID == id {
			return order
		}
	}
	return nil
}

// removeOrder removes an order by ID from the orders list
func (m *Model) removeOrder(id int) {
	for i, order := range m.orders {
		if order.ID == id {
			m.orders = append(m.orders[:i], m.orders[i+1:]...)
			return
		}
	}
}

// sweepCompletedOrders removes all orders with OrderCompleted status.
// Called once per game tick after intents are applied.
func (m *Model) sweepCompletedOrders() {
	n := 0
	for _, order := range m.orders {
		if order.Status != entity.OrderCompleted {
			m.orders[n] = order
			n++
		}
	}
	m.orders = m.orders[:n]
}

// viewFullScreenOrders renders full-screen orders panel
func (m Model) viewFullScreenOrders() string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("                    ORDERS"))
	lines = append(lines, "")
	lines = append(lines, m.renderOrdersContent(true)...)

	content := strings.Join(lines, "\n")
	return borderStyle.Width(m.width - 4).Height(m.height - 4).Render(content)
}
