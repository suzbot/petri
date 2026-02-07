package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Border style for panels
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	// Character status colors
	poisonedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))  // green
	deadStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // gray
	sleepingStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // purple/lavender
	frustratedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // orange

	// Severity colors for stat tiers
	optimalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // green (best)
	severeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // yellow
	crisisStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red

	// Effect wore off color
	woreOffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("45")) // cyan

	// Learning color
	learnedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33")) // darker blue

	// Order-related color
	orderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("174")) // dusty rose

	// Item colors
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	blueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	brownStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("136"))
	whiteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	orangeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	purpleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("135"))
	tanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("180"))
	pinkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	blackStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // dark gray for visibility
	greenStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	palePinkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("218"))
	paleYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	silverStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("188"))
	grayStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	lavenderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("183"))

	// Feature colors
	waterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // bright blue
	leafStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("106")) // olive/leaf green

	// Agricultural/plant status
	growingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("106")) // olive green

	// UI highlight (background)
	highlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("23")).Foreground(lipgloss.Color("255")) // dark cyan bg, white text

	// Title style
	titleStyle = lipgloss.NewStyle().Bold(true)
)
