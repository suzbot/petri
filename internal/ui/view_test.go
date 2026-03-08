package ui

import (
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

func TestHutSymbolFromAdjacency(t *testing.T) {
	// Helper to create a map with hut constructs at specified positions
	setup := func(positions ...types.Position) *game.Map {
		m := game.NewMap(20, 20)
		for _, pos := range positions {
			c := entity.NewHutConstruct(pos.X, pos.Y, "stick", types.ColorBrown, "wall")
			m.AddConstruct(c)
		}
		return m
	}

	center := types.Position{X: 10, Y: 10}
	north := types.Position{X: 10, Y: 9}
	south := types.Position{X: 10, Y: 11}
	east := types.Position{X: 11, Y: 10}
	west := types.Position{X: 9, Y: 10}

	tests := []struct {
		name      string
		neighbors []types.Position // positions with hut constructs (center always present)
		wantSym   rune
		wantLeft  bool // true if leftFill should be ━
		wantRight bool // true if rightFill should be ━
	}{
		// Corners — two perpendicular neighbors
		{"corner-tl (E+S)", []types.Position{east, south}, config.CharHutCornerTL, false, true},
		{"corner-tr (W+S)", []types.Position{west, south}, config.CharHutCornerTR, true, false},
		{"corner-bl (E+N)", []types.Position{east, north}, config.CharHutCornerBL, false, true},
		{"corner-br (W+N)", []types.Position{west, north}, config.CharHutCornerBR, true, false},

		// Edges — one axis
		{"edge-h (W+E)", []types.Position{west, east}, config.CharHutEdgeH, true, true},
		{"edge-v (N+S)", []types.Position{north, south}, config.CharHutEdgeV, false, false},

		// T-junctions — three neighbors
		{"t-down (W+E+S)", []types.Position{west, east, south}, config.CharHutTDown, true, true},
		{"t-up (W+E+N)", []types.Position{west, east, north}, config.CharHutTUp, true, true},
		{"t-right (N+S+E)", []types.Position{north, south, east}, config.CharHutTRight, false, true},
		{"t-left (N+S+W)", []types.Position{north, south, west}, config.CharHutTLeft, true, false},

		// Cross — all four neighbors
		{"cross (N+S+E+W)", []types.Position{north, south, east, west}, config.CharHutCross, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allPos := append([]types.Position{center}, tt.neighbors...)
			m := setup(allPos...)

			sym, leftFill, rightFill := hutSymbolFromAdjacency(center, m)

			if sym != tt.wantSym {
				t.Errorf("symbol: got %c, want %c", sym, tt.wantSym)
			}
			hasLeft := leftFill == string(config.CharHutEdgeH)
			if hasLeft != tt.wantLeft {
				t.Errorf("leftFill: got %q (hasLeft=%v), want hasLeft=%v", leftFill, hasLeft, tt.wantLeft)
			}
			hasRight := rightFill == string(config.CharHutEdgeH)
			if hasRight != tt.wantRight {
				t.Errorf("rightFill: got %q (hasRight=%v), want hasRight=%v", rightFill, hasRight, tt.wantRight)
			}
		})
	}
}

func TestHutSymbolFromAdjacency_DoorAlwaysReturnsDoorsymbol(t *testing.T) {
	// Door should always return ▯ regardless of neighbors, with both fills
	center := types.Position{X: 10, Y: 10}
	north := types.Position{X: 10, Y: 9}
	south := types.Position{X: 10, Y: 11}
	east := types.Position{X: 11, Y: 10}
	west := types.Position{X: 9, Y: 10}

	m := game.NewMap(20, 20)
	// Center is a door
	door := entity.NewHutConstruct(center.X, center.Y, "stick", types.ColorBrown, "door")
	m.AddConstruct(door)
	// Neighbors are walls
	for _, pos := range []types.Position{north, south, east, west} {
		c := entity.NewHutConstruct(pos.X, pos.Y, "stick", types.ColorBrown, "wall")
		m.AddConstruct(c)
	}

	sym, leftFill, rightFill := hutSymbolFromAdjacency(center, m)

	if sym != config.CharHutDoor {
		t.Errorf("door symbol: got %c, want %c", sym, config.CharHutDoor)
	}
	if leftFill != string(config.CharHutEdgeH) {
		t.Errorf("door leftFill: got %q, want %q", leftFill, string(config.CharHutEdgeH))
	}
	if rightFill != string(config.CharHutEdgeH) {
		t.Errorf("door rightFill: got %q, want %q", rightFill, string(config.CharHutEdgeH))
	}
}

func TestHutSymbolFromAdjacency_SingleOrNoNeighbor(t *testing.T) {
	center := types.Position{X: 10, Y: 10}

	tests := []struct {
		name      string
		neighbors []types.Position
		wantSym   rune
		wantLeft  bool
		wantRight bool
	}{
		{"no neighbors", nil, config.CharHutEdgeH, false, false},
		{"only north", []types.Position{{X: 10, Y: 9}}, config.CharHutEdgeV, false, false},
		{"only south", []types.Position{{X: 10, Y: 11}}, config.CharHutEdgeV, false, false},
		{"only east", []types.Position{{X: 11, Y: 10}}, config.CharHutEdgeH, false, true},
		{"only west", []types.Position{{X: 9, Y: 10}}, config.CharHutEdgeH, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := game.NewMap(20, 20)
			wall := entity.NewHutConstruct(center.X, center.Y, "stick", types.ColorBrown, "wall")
			m.AddConstruct(wall)
			for _, pos := range tt.neighbors {
				c := entity.NewHutConstruct(pos.X, pos.Y, "stick", types.ColorBrown, "wall")
				m.AddConstruct(c)
			}

			sym, leftFill, rightFill := hutSymbolFromAdjacency(center, m)

			if sym != tt.wantSym {
				t.Errorf("symbol: got %c, want %c", sym, tt.wantSym)
			}
			hasLeft := leftFill == string(config.CharHutEdgeH)
			if hasLeft != tt.wantLeft {
				t.Errorf("leftFill: got %q (hasLeft=%v), want hasLeft=%v", leftFill, hasLeft, tt.wantLeft)
			}
			hasRight := rightFill == string(config.CharHutEdgeH)
			if hasRight != tt.wantRight {
				t.Errorf("rightFill: got %q (hasRight=%v), want hasRight=%v", rightFill, hasRight, tt.wantRight)
			}
		})
	}
}
