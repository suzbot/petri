package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// Cardinal Adjacency Helpers
// =============================================================================

func TestIsCardinallyAdjacent_Cardinal_True(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		x1, y1   int
		x2, y2   int
		expected bool
	}{
		{"north", 5, 5, 5, 4, true},
		{"south", 5, 5, 5, 6, true},
		{"east", 5, 5, 6, 5, true},
		{"west", 5, 5, 4, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCardinallyAdjacent(tt.x1, tt.y1, tt.x2, tt.y2)
			if got != tt.expected {
				t.Errorf("isCardinallyAdjacent(%d,%d,%d,%d): got %v, want %v",
					tt.x1, tt.y1, tt.x2, tt.y2, got, tt.expected)
			}
		})
	}
}

func TestIsCardinallyAdjacent_Diagonal_False(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		x1, y1 int
		x2, y2 int
	}{
		{"northeast", 5, 5, 6, 4},
		{"northwest", 5, 5, 4, 4},
		{"southeast", 5, 5, 6, 6},
		{"southwest", 5, 5, 4, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCardinallyAdjacent(tt.x1, tt.y1, tt.x2, tt.y2)
			if got {
				t.Errorf("isCardinallyAdjacent(%d,%d,%d,%d): got true for diagonal, want false",
					tt.x1, tt.y1, tt.x2, tt.y2)
			}
		})
	}
}

func TestIsCardinallyAdjacent_SamePosition_False(t *testing.T) {
	t.Parallel()

	got := isCardinallyAdjacent(5, 5, 5, 5)
	if got {
		t.Error("isCardinallyAdjacent(5,5,5,5): got true for same position, want false")
	}
}

func TestIsCardinallyAdjacent_TooFar_False(t *testing.T) {
	t.Parallel()

	got := isCardinallyAdjacent(5, 5, 7, 5)
	if got {
		t.Error("isCardinallyAdjacent(5,5,7,5): got true for distance 2, want false")
	}
}

func TestFindClosestCardinalTile_FindsClosest(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Character at (3, 5), target spring at (5, 5)
	// Cardinal tiles around spring: (5,4), (6,5), (5,6), (4,5)
	// Closest to (3,5) is (4,5) with distance 1
	x, y := FindClosestCardinalTile(3, 5, 5, 5, gameMap)

	if x != 4 || y != 5 {
		t.Errorf("FindClosestCardinalTile: got (%d,%d), want (4,5)", x, y)
	}
}

func TestFindClosestCardinalTile_SkipsBlocked(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Block the closest tile with a character
	gameMap.AddCharacter(entity.NewCharacter(1, 4, 5, "Blocker", "berry", types.ColorRed))

	// Character at (3, 5), target spring at (5, 5)
	// (4,5) is blocked, next closest should be (5,4) or (5,6) with distance 3
	x, y := FindClosestCardinalTile(3, 5, 5, 5, gameMap)

	// Should not return the blocked tile
	if x == 4 && y == 5 {
		t.Error("FindClosestCardinalTile should skip blocked tiles")
	}
	// Should return a valid tile
	if x == -1 {
		t.Error("FindClosestCardinalTile should find an unblocked tile")
	}
}

func TestFindClosestCardinalTile_AllBlocked(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// Block all cardinal tiles around (5, 5)
	gameMap.AddCharacter(entity.NewCharacter(1, 5, 4, "N", "berry", types.ColorRed))
	gameMap.AddCharacter(entity.NewCharacter(2, 6, 5, "E", "berry", types.ColorRed))
	gameMap.AddCharacter(entity.NewCharacter(3, 5, 6, "S", "berry", types.ColorRed))
	gameMap.AddCharacter(entity.NewCharacter(4, 4, 5, "W", "berry", types.ColorRed))

	x, y := FindClosestCardinalTile(3, 5, 5, 5, gameMap)

	if x != -1 || y != -1 {
		t.Errorf("FindClosestCardinalTile: got (%d,%d), want (-1,-1) when all blocked", x, y)
	}
}

// =============================================================================
// BFS Pathfinding
// =============================================================================

func TestNextStepBFS_ClearPath_MatchesGreedy(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)

	// No obstacles - BFS should match greedy NextStep
	nx, ny := NextStepBFS(5, 5, 10, 5, gameMap)
	gx, gy := NextStep(5, 5, 10, 5)
	if nx != gx || ny != gy {
		t.Errorf("Clear path: BFS got (%d,%d), greedy got (%d,%d)", nx, ny, gx, gy)
	}
}

func TestNextStepBFS_RoutesAroundWater(t *testing.T) {
	t.Parallel()

	// Character at (5,5), target at (5,2), water wall at y=3 from x=3 to x=7
	gameMap := game.NewMap(20, 20)
	for x := 3; x <= 7; x++ {
		gameMap.AddWater(types.Position{X: x, Y: 3}, game.WaterPond)
	}

	nx, ny := NextStepBFS(5, 5, 5, 2, gameMap)

	// Greedy would try (5,4) then get stuck at (5,3). BFS should route around.
	// The first step should NOT be directly toward the target (which leads to water wall).
	// It should go sideways toward an end of the wall.
	// Valid first steps: (4,5) or (6,5) — moving laterally to go around
	// OR (5,4) is also valid if BFS finds a path through (5,4) then sideways
	// The key test: the step should be part of a valid path (not into water)
	stepPos := types.Position{X: nx, Y: ny}
	if gameMap.IsWater(stepPos) {
		t.Errorf("BFS stepped into water at (%d,%d)", nx, ny)
	}
	// Step should be adjacent to start
	start := types.Position{X: 5, Y: 5}
	if start.DistanceTo(stepPos) != 1 {
		t.Errorf("BFS step (%d,%d) is not adjacent to start (5,5)", nx, ny)
	}
}

func TestNextStepBFS_RoutesAroundPond(t *testing.T) {
	t.Parallel()

	// Character at (5,5), target at (5,1)
	// 2x2 pond blocking direct path at (5,3),(6,3),(5,4),(6,4)
	gameMap := game.NewMap(20, 20)
	gameMap.AddWater(types.Position{X: 5, Y: 3}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 6, Y: 3}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 5, Y: 4}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 6, Y: 4}, game.WaterPond)

	nx, ny := NextStepBFS(5, 5, 5, 1, gameMap)

	// Should step sideways (left) to route around the pond
	// (4,5) is the expected first step to go around the left side
	if nx != 4 || ny != 5 {
		t.Errorf("Route around pond: got (%d,%d), want (4,5)", nx, ny)
	}
}

func TestNextStepBFS_AlreadyAtTarget(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	nx, ny := NextStepBFS(5, 5, 5, 5, gameMap)
	if nx != 5 || ny != 5 {
		t.Errorf("At target: got (%d,%d), want (5,5)", nx, ny)
	}
}

func TestNextStepBFS_NoPath_FallsBackToGreedy(t *testing.T) {
	t.Parallel()

	// Surround character with water (no path possible)
	gameMap := game.NewMap(20, 20)
	gameMap.AddWater(types.Position{X: 4, Y: 5}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 6, Y: 5}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 5, Y: 4}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 5, Y: 6}, game.WaterPond)

	nx, ny := NextStepBFS(5, 5, 10, 10, gameMap)

	// Falls back to greedy (which doesn't check obstacles)
	gx, gy := NextStep(5, 5, 10, 10)
	if nx != gx || ny != gy {
		t.Errorf("No path fallback: got (%d,%d), want greedy (%d,%d)", nx, ny, gx, gy)
	}
}

func TestNextStepBFS_LShapedPond(t *testing.T) {
	t.Parallel()

	// L-shaped pond forcing character to go around
	// Character at (3,5), target at (3,1)
	// Pond: (3,3),(3,4),(4,3),(4,4),(5,3),(5,4) - wide horizontal wall
	gameMap := game.NewMap(20, 20)
	for x := 3; x <= 5; x++ {
		gameMap.AddWater(types.Position{X: x, Y: 3}, game.WaterPond)
		gameMap.AddWater(types.Position{X: x, Y: 4}, game.WaterPond)
	}

	nx, ny := NextStepBFS(3, 5, 3, 1, gameMap)

	// Should go sideways to route around (either left or right)
	stepPos := types.Position{X: nx, Y: ny}
	if gameMap.IsWater(stepPos) {
		t.Errorf("BFS stepped into water at (%d,%d)", nx, ny)
	}
	start := types.Position{X: 3, Y: 5}
	if start.DistanceTo(stepPos) != 1 {
		t.Errorf("BFS step (%d,%d) is not adjacent to start (3,5)", nx, ny)
	}
}

// =============================================================================
// Greedy-first pathfinding (Slice 9 Step 2: path diversity)
// =============================================================================

// Anchor: two characters from different positions heading to the same destination
// take different first steps, reducing path convergence.
func TestNextStepBFS_GreedyFirst_PathDiversity(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(30, 30)

	// Target: (15, 15)
	// Character A at (5, 10) — dx=10, dy=5 → larger dx → steps right
	// Character B at (10, 5) — dx=5, dy=10 → larger dy → steps down
	ax, ay := NextStepBFS(5, 10, 15, 15, gameMap)
	bx, by := NextStepBFS(10, 5, 15, 15, gameMap)

	// A should step right (X+1)
	if ax != 6 || ay != 10 {
		t.Errorf("Character A: got (%d,%d), want (6,10) — should step along larger dx", ax, ay)
	}
	// B should step down (Y+1)
	if bx != 10 || by != 6 {
		t.Errorf("Character B: got (%d,%d), want (10,6) — should step along larger dy", bx, by)
	}
	// Different first steps
	if ax == bx && ay == by {
		t.Error("Both characters took the same first step — paths should diverge")
	}
}

// Greedy produces zigzag: alternates X and Y steps as relative deltas shift.
func TestNextStepBFS_GreedyFirst_ZigzagPath(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(30, 30)

	// Walk from (0,0) to (4,3) — should zigzag, not L-shape
	x, y := 0, 0
	steps := make([][2]int, 0, 7)
	for x != 4 || y != 3 {
		nx, ny := NextStepBFS(x, y, 4, 3, gameMap)
		steps = append(steps, [2]int{nx, ny})
		x, y = nx, ny
	}

	// Count direction changes (X-move vs Y-move transitions)
	dirChanges := 0
	for i := 1; i < len(steps); i++ {
		prevDX := steps[i-1][0] - (steps[i-1][0] - (steps[i][0] - steps[i-1][0]))
		_ = prevDX
		// Simpler: track whether each step was X or Y
		xMove := steps[i][0] != steps[i-1][0]
		prevXMove := false
		if i == 1 {
			prevXMove = steps[0][0] != 0 || (steps[0][0] == 0 && steps[0][1] == 0)
		} else {
			prevXMove = steps[i-1][0] != steps[i-2][0]
		}
		if xMove != prevXMove {
			dirChanges++
		}
	}

	// A zigzag path of 7 steps should have multiple direction changes.
	// An L-shape has exactly 1 direction change. Zigzag should have more.
	if dirChanges < 2 {
		t.Errorf("Expected zigzag path (>= 2 direction changes), got %d changes. Steps: %v", dirChanges, steps)
	}
}

// Greedy step blocked by water → falls back to BFS routing.
func TestNextStepBFS_GreedyBlocked_FallsBackToBFS(t *testing.T) {
	t.Parallel()

	// Character at (5,5), target (5,1). Greedy step is (5,4).
	// Place water at (5,4) so greedy fails but BFS can route around.
	gameMap := game.NewMap(20, 20)
	gameMap.AddWater(types.Position{X: 5, Y: 4}, game.WaterPond)

	nx, ny := NextStepBFS(5, 5, 5, 1, gameMap)

	// Should not step into water
	if nx == 5 && ny == 4 {
		t.Error("Should not step into water at (5,4) — BFS fallback should route around")
	}
	// Should be adjacent to start
	start := types.Position{X: 5, Y: 5}
	stepPos := types.Position{X: nx, Y: ny}
	if start.DistanceTo(stepPos) != 1 {
		t.Errorf("Step (%d,%d) is not adjacent to start (5,5)", nx, ny)
	}
}

// =============================================================================
// Sticky BFS (characters stay in BFS mode once obstacle detected)
// =============================================================================

// Anchor test: character navigates around a pocket-shaped obstacle and reaches
// its target. Without sticky BFS, greedy-first oscillates between the pocket
// entrance (1,3) and the tile south of it (1,4) forever:
//   - At (1,3): greedy tries (1,2)-water, BFS routes south to (1,4)
//   - At (1,4): greedy tries (1,3)-clear, moves back into the pocket
//
// With sticky BFS, once BFS fires at (1,3), the character stays in BFS mode
// at (1,4) and routes east around the obstacle instead of re-entering the pocket.
func TestStickyBFS_AntiThrashing_PocketObstacle(t *testing.T) {
	t.Parallel()

	// Map layout (10x10):
	//   col: 0 1 2 3 4 5
	//   row 0: T . . . . .    target at (0,0)
	//   row 1: . . . . . .
	//   row 2: . W . . . .    water at (1,2)
	//   row 3: W . W . . .    water at (0,3) and (2,3) — pocket at (1,3)
	//   row 4: . . . . . .
	//   row 5: . C . . . .    character starts at (1,5)
	//
	// The pocket at (1,3) has water on three sides (N, W, E) with only
	// south (1,4) as exit. Character approaches via greedy, enters pocket,
	// then must use BFS to escape and navigate east.

	gameMap := game.NewMap(10, 10)
	gameMap.AddWater(types.Position{X: 1, Y: 2}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 0, Y: 3}, game.WaterPond)
	gameMap.AddWater(types.Position{X: 2, Y: 3}, game.WaterPond)

	x, y := 1, 5
	tx, ty := 0, 0
	usingBFS := false

	for i := 0; i < 50; i++ {
		if x == tx && y == ty {
			break
		}
		nx, ny, usedBFS := nextStepBFSCore(x, y, tx, ty, gameMap, usingBFS)
		if usedBFS {
			usingBFS = true
		}
		x, y = nx, ny
	}

	if x != tx || y != ty {
		t.Errorf("Character did not reach target: stuck at (%d,%d), target (%d,%d)", x, y, tx, ty)
	}
}

// nextStepBFSCore with preferBFS=true skips greedy step and uses BFS directly.
func TestNextStepBFSCore_PreferBFS_SkipsGreedy(t *testing.T) {
	t.Parallel()

	// Character at (5,5), target at (5,1). No obstacles.
	// Greedy would step (5,4) — along Y (larger delta).
	// BFS only has cardinal directions — would step (5,4) or one of the cardinals.
	// With preferBFS=true, the result should be a valid BFS step (cardinal only).
	gameMap := game.NewMap(20, 20)

	// Place water at (5,4) so greedy and BFS diverge
	gameMap.AddWater(types.Position{X: 5, Y: 4}, game.WaterPond)

	// Without preferBFS: greedy tries (5,4-water), falls through to BFS
	nx1, ny1, used1 := nextStepBFSCore(5, 5, 5, 1, gameMap, false)
	if !used1 {
		t.Error("Expected usedBFS=true when greedy step is blocked by water")
	}

	// With preferBFS: skips greedy entirely, goes straight to BFS
	nx2, ny2, used2 := nextStepBFSCore(5, 5, 5, 1, gameMap, true)
	if !used2 {
		t.Error("Expected usedBFS=true when preferBFS=true")
	}

	// Both should produce the same result (BFS from same position)
	if nx1 != nx2 || ny1 != ny2 {
		t.Errorf("preferBFS should produce same BFS result: got (%d,%d) vs (%d,%d)", nx1, ny1, nx2, ny2)
	}
}

// nextStepBFSCore returns usedBFS=false when greedy step succeeds.
func TestNextStepBFSCore_GreedyClear_ReturnsFalse(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	// No obstacles — greedy should work fine
	nx, ny, usedBFS := nextStepBFSCore(5, 5, 10, 5, gameMap, false)

	gx, gy := NextStep(5, 5, 10, 5)
	if nx != gx || ny != gy {
		t.Errorf("Clear path: got (%d,%d), want greedy (%d,%d)", nx, ny, gx, gy)
	}
	if usedBFS {
		t.Error("Expected usedBFS=false when greedy step is clear")
	}
}

// Sticky BFS flag clears when CalculateIntent creates a new intent
// (not continuing an existing one).
func TestStickyBFS_ClearsOnNewIntent(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	char := entity.NewCharacter(1, 5, 5, "Test", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Set UsingBFS flag as if character was navigating around an obstacle
	char.UsingBFS = true
	// No intent — CalculateIntent will create a new one from scratch
	char.Intent = nil

	// CalculateIntent should clear UsingBFS before creating new intent
	CalculateIntent(char, nil, gameMap, nil, nil)

	if char.UsingBFS {
		t.Error("Expected UsingBFS=false after CalculateIntent creates a new intent")
	}
}

func TestNextStepBFS_RoutesAroundConstruct(t *testing.T) {
	t.Parallel()

	// Character at (5,5), target at (5,2), fence wall at y=3 from x=3 to x=7
	gameMap := game.NewMap(20, 20)
	for x := 3; x <= 7; x++ {
		fence := entity.NewFence(x, 3, "stick", types.ColorBrown)
		gameMap.AddConstruct(fence)
	}

	nx, ny := NextStepBFS(5, 5, 5, 2, gameMap)

	// Step should not be into a construct
	stepPos := types.Position{X: nx, Y: ny}
	if gameMap.ConstructAt(stepPos) != nil {
		t.Errorf("BFS stepped into construct at (%d,%d)", nx, ny)
	}
	// Step should be adjacent to start
	start := types.Position{X: 5, Y: 5}
	if start.DistanceTo(stepPos) != 1 {
		t.Errorf("BFS step (%d,%d) is not adjacent to start (5,5)", nx, ny)
	}
}
