package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"petri/internal/ui"
)

const Version = "0.0.1"

func main() {
	// Test mode flags
	noFood := flag.Bool("no-food", false, "Skip spawning food items (test mode)")
	noWater := flag.Bool("no-water", false, "Skip spawning water sources (test mode)")
	noBeds := flag.Bool("no-beds", false, "Skip spawning beds (test mode)")
	noCharacters := flag.Bool("no-characters", false, "Skip spawning characters (test mode)")
	debug := flag.Bool("debug", false, "Show debug info (action progress, etc.)")
	mushroomsOnly := flag.Bool("mushrooms-only", false, "Replace all items with mushroom varieties (test mode)")
	version := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *version {
		fmt.Println("Version:", Version)
		os.Exit(0)
	}

	testCfg := ui.TestConfig{
		NoFood:        *noFood,
		NoWater:       *noWater,
		NoBeds:        *noBeds,
		NoCharacters:  *noCharacters,
		Debug:         *debug,
		MushroomsOnly: *mushroomsOnly,
	}

	p := tea.NewProgram(
		ui.NewModel(testCfg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
