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
	optimalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("28"))  // dark green (best)
	severeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // yellow
	crisisStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red

	// Effect wore off color
	woreOffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117")) // light blue

	// Item colors
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	blueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	brownStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("136"))
	whiteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	orangeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	purpleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("135"))

	// Feature colors
	waterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // bright blue
	leafStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("106")) // olive/leaf green

	// UI highlight
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("45")) // cyan/teal

	// Title style
	titleStyle = lipgloss.NewStyle().Bold(true)
)
