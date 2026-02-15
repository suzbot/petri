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

	// Item colors (Bold helps Unicode symbols render more prominently)
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	blueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("27")).Bold(true)
	brownStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("136")).Bold(true)
	whiteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	orangeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	purpleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("135")).Bold(true)
	tanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("180")).Bold(true)
	pinkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	blackStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true) // dark gray for visibility
	greenStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
	palePinkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("218")).Bold(true)
	paleYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	silverStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("188")).Bold(true)
	grayStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Bold(true)
	lavenderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("183")).Bold(true)

	// Feature colors
	waterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)  // bright blue
	leafStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("106")).Bold(true) // olive/leaf green

	// Agricultural/plant status
	growingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("142")).Bold(true) // olive (gardening: growing, tilled soil)
	sproutStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("108")).Bold(true) // sage (sprouts on dry ground)

	// UI highlight (background)
	highlightStyle    = lipgloss.NewStyle().Background(lipgloss.Color("23")).Foreground(lipgloss.Color("255"))  // dark cyan bg, white text
	areaSelectStyle   = lipgloss.NewStyle().Background(lipgloss.Color("58"))                                   // olive bg for area selection
	areaUnselectStyle = lipgloss.NewStyle().Background(lipgloss.Color("52"))                                   // dark red bg for unmark selection

	// Unfulfillable order style (dimmed)
	unfulfillableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // gray

	// Title style
	titleStyle = lipgloss.NewStyle().Bold(true)
)
