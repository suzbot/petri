package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

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
	// Learning messages (darker blue) - check first as these are important
	if strings.Contains(message, "Learned something") {
		return learnedStyle.Render(line)
	}

	// Effect wore off messages (light blue)
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
	sleepKeywords := []string{"sleeping", "Falling asleep"}
	for _, kw := range sleepKeywords {
		if strings.Contains(message, kw) {
			return sleepingStyle.Render(line)
		}
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
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(
		"↑/↓ Select   Enter: Continue   Q: Quit"))

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
		Width(18)

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
		foodLabel := "Food:"
		colorLabel := "Color:"

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
				foodLabel = selectedFieldStyle.Render("Food:")
				foodValue = selectedFieldStyle.Render("[" + charData.Food + "]")
			case FieldColor:
				colorLabel = selectedFieldStyle.Render("Color:")
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
	// Full-screen log view
	if m.viewMode == viewModeFullLog {
		return m.viewFullLog()
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
	if m.viewMode == viewModeAllActivity {
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
		} else {
			bottomPanel = borderStyle.Width(panelWidth).Height(logHeight).Render(m.renderActionLog())
		}
		rightPanel = lipgloss.JoinVertical(lipgloss.Left, detailsView, bottomPanel)
	}

	// Horizontal layout: Map | Right Panel
	gameArea := lipgloss.JoinHorizontal(lipgloss.Top, mapView, " ", rightPanel)

	// Status bar with mode-specific hints
	status := "RUNNING"
	stepHint := ""
	if m.paused {
		status = "PAUSED"
		stepHint = " | .=step"
	}

	// Saving indicator
	saveHint := ""
	if time.Now().Before(m.saveIndicatorEnd) {
		saveHint = " " + orangeStyle.Render("[Saved]")
	}

	modeHint := ""
	if m.viewMode == viewModeAllActivity {
		modeHint = " | s=select | l=full log"
	} else {
		modeHint = " | a=all activity | l=full log | n=next char"
	}

	statusBar := fmt.Sprintf("\n[%s]%s SPACE=pause%s%s | ARROWS=cursor | ESC=menu", status, saveHint, stepHint, modeHint)

	// Debug line (only shown with -debug flag)
	debugLine := ""
	if m.testCfg.Debug {
		var charInfo []string
		for _, c := range m.gameMap.Characters() {
			x, y := c.Position()
			found := m.gameMap.CharacterAt(x, y)
			marker := "✓"
			if found != c {
				marker = "✗"
			}
			charInfo = append(charInfo, fmt.Sprintf("%s(%d,%d)%s", c.Name, x, y, marker))
		}
		debugLine = fmt.Sprintf("\nChars: %v", charInfo)
	}

	return gameArea + statusBar + debugLine
}

// viewFullLog renders a full-screen log view with complete (non-truncated) messages
func (m Model) viewFullLog() string {
	var lines []string

	// Header
	lines = append(lines, titleStyle.Render("=== FULL LOG VIEW ==="))
	lines = append(lines, "")

	// Get all events across all characters, sorted by time (0 = no limit)
	allEvents := m.actionLog.AllEvents(0)

	// Calculate how many lines we can show - use map height for consistency with game view
	// Leave room for header (2 lines), footer (2 lines), and status bar (1 line)
	availableLines := m.gameMap.Height - 3

	// Apply scroll offset (from end, showing most recent first)
	total := len(allEvents)
	endIdx := total - m.logScrollOffset
	if endIdx < 0 {
		endIdx = 0
	}
	startIdx := endIdx - availableLines
	if startIdx < 0 {
		startIdx = 0
	}

	if m.logScrollOffset > 0 {
		lines = append(lines, fmt.Sprintf(" [Scrolled: -%d]", m.logScrollOffset))
	}

	// Render events (most recent first, no truncation)
	displayEvents := allEvents[startIdx:endIdx]
	for i := len(displayEvents) - 1; i >= 0; i-- {
		event := displayEvents[i]
		elapsedSecs := m.elapsedGameTime - event.GameTime

		// Full message with character name prefix
		// Only show timestamp if >= 1 minute (like other logs)
		var line string
		if elapsedSecs >= 60 {
			line = fmt.Sprintf(" %s [%s] %s", system.FormatGameTime(elapsedSecs), event.CharName, event.Message)
		} else {
			line = fmt.Sprintf(" [%s] %s", event.CharName, event.Message)
		}

		coloredLine := colorLogMessage(line, event.Message)

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

	// Footer with controls
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(
		"PgUp/PgDn: Scroll | L: Back"))

	content := strings.Join(lines, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
}

// renderCell renders a single map cell
func (m Model) renderCell(x, y int) string {
	isCursor := x == m.cursorX && y == m.cursorY

	var sym string

	// Check for character first (takes visual precedence)
	if char := m.gameMap.CharacterAt(x, y); char != nil {
		sym = m.styledSymbol(char)
	} else if item := m.gameMap.ItemAt(x, y); item != nil {
		// Check for item
		sym = m.styledSymbol(item)
	} else if feature := m.gameMap.FeatureAt(x, y); feature != nil {
		// Check for feature
		sym = m.styledSymbol(feature)
	} else {
		sym = " "
	}

	if isCursor {
		return "[" + sym + "]"
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
		}

	case *entity.Feature:
		if v.IsDrinkSource() {
			return waterStyle.Render(sym)
		}
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
	e := m.gameMap.EntityAt(m.cursorX, m.cursorY)
	item := m.gameMap.ItemAt(m.cursorX, m.cursorY)
	feature := m.gameMap.FeatureAt(m.cursorX, m.cursorY)

	if e == nil && item == nil && feature == nil {
		lines = append(lines, " Type: Empty")
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
	} else if char, ok := e.(*entity.Character); ok {
		lines = append(lines, fmt.Sprintf(" Name: %s", char.Name))
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
		lines = append(lines,
			activityLine,
			"",
			" Preferences:",
		)
		for _, pref := range char.Preferences {
			verb := "Likes"
			if !pref.IsPositive() {
				verb = "Dislikes"
			}
			lines = append(lines, fmt.Sprintf("   %s %s", verb, pref.Description()))
		}
		if len(char.Preferences) == 0 {
			lines = append(lines, "   (none)")
		}

		if m.following == char {
			lines = append(lines, "", highlightStyle.Render(" [FOLLOWING]"), " Press F to unfollow")
		} else {
			lines = append(lines, "", " Press F to follow")
		}
		lines = append(lines, " Press K for Knowledge")

	} else if item != nil {
		poison := "No"
		if item.Poisonous {
			poison = redStyle.Render("Yes")
		}
		healing := "No"
		if item.Healing {
			healing = optimalStyle.Render("Yes")
		}
		lines = append(lines, " Type: Item")
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
		lines = append(lines,
			fmt.Sprintf(" Kind: %s", item.ItemType),
			fmt.Sprintf(" Color: %s", item.Color),
		)
		// Show Pattern/Texture for item types that can have them
		configs := game.GetItemTypeConfigs()
		if cfg, ok := configs[item.ItemType]; ok {
			if cfg.Patterns != nil {
				pattern := "none"
				if item.Pattern != "" {
					pattern = string(item.Pattern)
				}
				lines = append(lines, fmt.Sprintf(" Pattern: %s", pattern))
			}
			if cfg.Textures != nil {
				texture := "none"
				if item.Texture != "" {
					texture = string(item.Texture)
				}
				lines = append(lines, fmt.Sprintf(" Texture: %s", texture))
			}
		}
		lines = append(lines,
			fmt.Sprintf(" Poisonous: %s", poison),
			fmt.Sprintf(" Healing: %s", healing),
		)
	} else if feature != nil {
		lines = append(lines, " Type: Feature")
		if m.testCfg.Debug {
			lines = append(lines, fmt.Sprintf(" Pos: (%d, %d)", m.cursorX, m.cursorY))
		}
		lines = append(lines, fmt.Sprintf(" Kind: %s", feature.Description()))
		if feature.IsDrinkSource() {
			lines = append(lines, " Use: Drinking")
		}
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

	e := m.gameMap.EntityAt(m.cursorX, m.cursorY)
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
				strings.Contains(event.Message, "impacted health") {
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

	events := m.actionLog.AllEvents(200)

	// Pre-filter debug-only messages in non-debug mode
	if !m.testCfg.Debug {
		var filtered []system.Event
		for _, event := range events {
			if strings.Contains(event.Message, "Improved Mood") ||
				strings.Contains(event.Message, "Worsened Mood") ||
				strings.Contains(event.Message, "impacted health") {
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

	// Calculate display range (full panel height minus header lines)
	// All Activity panel uses full map height (gameMap.Height) for content
	allActivityHeight := m.gameMap.Height
	maxDisplay := allActivityHeight - 3 // Account for header and borders

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

// renderKnowledgePanel renders the knowledge panel for the selected character
func (m Model) renderKnowledgePanel() string {
	var lines []string
	lines = append(lines, titleStyle.Render("       KNOWLEDGE"), "")

	// Get character at cursor
	if e := m.gameMap.EntityAt(m.cursorX, m.cursorY); e != nil {
		if char, ok := e.(*entity.Character); ok {
			if len(char.Knowledge) == 0 {
				// Empty state - just show title, panel appears empty
				return strings.Join(lines, "\n")
			}

			// Show knowledge entries
			for _, k := range char.Knowledge {
				line := " " + k.Description()
				lines = append(lines, line)
			}
		} else {
			lines = append(lines, " Select a character")
		}
	} else {
		lines = append(lines, " Select a character")
	}

	lines = append(lines, "", " Press K to return")

	return strings.Join(lines, "\n")
}

// renderInventoryPanel renders the inventory panel for the selected character
func (m Model) renderInventoryPanel() string {
	var lines []string
	lines = append(lines, titleStyle.Render("       INVENTORY"), "")

	// Get character at cursor
	if e := m.gameMap.EntityAt(m.cursorX, m.cursorY); e != nil {
		if char, ok := e.(*entity.Character); ok {
			if char.Carrying != nil {
				lines = append(lines, " Carrying: "+char.Carrying.Description())
			} else {
				lines = append(lines, " Carrying: nothing")
			}
		} else {
			lines = append(lines, " Select a character")
		}
	} else {
		lines = append(lines, " Select a character")
	}

	lines = append(lines, "", " Press I to return")

	return strings.Join(lines, "\n")
}
