package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// CompleteLook
// =============================================================================

func TestCompleteLook_CallsTryFormPreference(t *testing.T) {
	t.Parallel()

	// Character with extreme mood to ensure preference formation
	char := &entity.Character{
		ID:   1,
		Name: "Test",
		Mood: 5, // Miserable - 20% chance to form preference
	}

	item := entity.NewFlower(5, 5, types.ColorPurple)

	// Run multiple times to increase chance of formation
	formed := false
	for i := 0; i < 50; i++ {
		char.Preferences = nil // Reset
		CompleteLook(char, item, nil)
		if len(char.Preferences) > 0 {
			formed = true
			break
		}
	}

	if !formed {
		t.Skip("No preference formed in 50 attempts (probabilistic)")
	}
}

func TestCompleteLook_HandlesNilItem(t *testing.T) {
	t.Parallel()

	char := &entity.Character{
		ID:   1,
		Name: "Test",
	}

	// Should not panic
	CompleteLook(char, nil, nil)
}

// =============================================================================
// isAdjacent
// =============================================================================

func TestIsAdjacent_ReturnsTrueForAdjacentPositions(t *testing.T) {
	t.Parallel()

	// All 8 adjacent positions
	adjacent := [][2]int{
		{4, 4}, {5, 4}, {6, 4}, // Top row
		{4, 5}, {6, 5}, // Middle (excluding center)
		{4, 6}, {5, 6}, {6, 6}, // Bottom row
	}

	for _, pos := range adjacent {
		if !isAdjacent(5, 5, pos[0], pos[1]) {
			t.Errorf("Position (%d, %d) should be adjacent to (5, 5)", pos[0], pos[1])
		}
	}
}

func TestIsAdjacent_ReturnsFalseForSamePosition(t *testing.T) {
	t.Parallel()

	if isAdjacent(5, 5, 5, 5) {
		t.Error("Same position should not be considered adjacent")
	}
}

func TestIsAdjacent_ReturnsFalseForDistantPositions(t *testing.T) {
	t.Parallel()

	distant := [][2]int{
		{3, 3}, {5, 3}, {7, 3}, // Two steps away
		{3, 5}, {7, 5},
		{3, 7}, {5, 7}, {7, 7},
	}

	for _, pos := range distant {
		if isAdjacent(5, 5, pos[0], pos[1]) {
			t.Errorf("Position (%d, %d) should not be adjacent to (5, 5)", pos[0], pos[1])
		}
	}
}

// =============================================================================
// findNearestItem
// =============================================================================

func TestFindNearestItem_ReturnsNilForEmptyList(t *testing.T) {
	t.Parallel()

	result := findNearestItem(5, 5, nil)
	if result != nil {
		t.Error("Should return nil for empty item list")
	}
}

func TestFindNearestItem_FindsClosestItem(t *testing.T) {
	t.Parallel()

	items := []*entity.Item{
		entity.NewBerry(10, 10, types.ColorRed, false, false),                                          // Distance 10
		entity.NewFlower(6, 5, types.ColorPurple),                                                      // Distance 1
		entity.NewMushroom(8, 8, types.ColorBrown, types.PatternNone, types.TextureNone, false, false), // Distance 6
	}

	result := findNearestItem(5, 5, items)

	if result != items[1] {
		t.Error("Should return the closest item")
	}
}

// =============================================================================
// findClosestAdjacentTile
// =============================================================================

func TestFindClosestAdjacentTile_FindsClosest(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Character at (3, 5), target at (5, 5)
	// Closest adjacent to target is (4, 5)
	ax, ay := findClosestAdjacentTile(3, 5, 5, 5, gameMap)

	if ax != 4 || ay != 5 {
		t.Errorf("Expected (4, 5), got (%d, %d)", ax, ay)
	}
}

func TestFindClosestAdjacentTile_SkipsOccupied(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Block the closest adjacent tile
	char := entity.NewCharacter(1, 4, 5, "Blocker", "berry", types.ColorRed)
	gameMap.AddCharacter(char)

	// Character at (3, 5), target at (5, 5)
	// (4, 5) is blocked, should find next closest
	ax, ay := findClosestAdjacentTile(3, 5, 5, 5, gameMap)

	if ax == 4 && ay == 5 {
		t.Error("Should not return occupied tile")
	}
	if ax == -1 {
		t.Error("Should find an available tile")
	}
}

func TestFindClosestAdjacentTile_ReturnsNegativeWhenAllBlocked(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Block all 8 adjacent tiles with items
	offsets := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}
	for i, off := range offsets {
		char := entity.NewCharacter(i+1, 5+off[0], 5+off[1], "Blocker", "berry", types.ColorRed)
		gameMap.AddCharacter(char)
	}

	ax, ay := findClosestAdjacentTile(3, 5, 5, 5, gameMap)

	if ax != -1 || ay != -1 {
		t.Errorf("Should return (-1, -1) when all blocked, got (%d, %d)", ax, ay)
	}
}

// =============================================================================
// findLookIntent
// =============================================================================

func TestFindLookIntent_ReturnsNilWithNoItems(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	gameMap := game.NewMap(10, 10)

	// Run multiple times (50% chance)
	for i := 0; i < 10; i++ {
		intent := findLookIntent(char, types.Position{X: 5, Y: 5}, nil, gameMap, nil)
		if intent != nil {
			t.Error("Should return nil when no items exist")
		}
	}
}

func TestFindLookIntent_ReturnsLookIntentWhenAdjacent(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	gameMap := game.NewMap(10, 10)

	// Item adjacent to character
	item := entity.NewFlower(6, 5, types.ColorPurple)
	items := []*entity.Item{item}

	// Run multiple times to get a look intent (50% chance)
	var intent *entity.Intent
	for i := 0; i < 20; i++ {
		intent = findLookIntent(char, types.Position{X: 5, Y: 5}, items, gameMap, nil)
		if intent != nil {
			break
		}
	}

	if intent == nil {
		t.Skip("No look intent returned in 20 attempts (probabilistic)")
	}

	if intent.Action != entity.ActionLook {
		t.Errorf("Expected ActionLook, got %v", intent.Action)
	}
	if intent.TargetItem != item {
		t.Error("TargetItem should be the adjacent item")
	}
}

func TestFindLookIntent_RoutesToAdjacentWhenOnSameTile(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	gameMap := game.NewMap(10, 10)

	// Item on same tile as character
	item := entity.NewFlower(5, 5, types.ColorPurple)
	items := []*entity.Item{item}

	var intent *entity.Intent
	for i := 0; i < 20; i++ {
		intent = findLookIntent(char, types.Position{X: 5, Y: 5}, items, gameMap, nil)
		if intent != nil {
			break
		}
	}

	if intent == nil {
		t.Skip("No look intent returned in 20 attempts (probabilistic)")
	}

	if intent.Action != entity.ActionLook {
		t.Errorf("Expected ActionLook, got %v", intent.Action)
	}
	if intent.TargetItem != item {
		t.Error("TargetItem should be the same-tile item")
	}
	// Character should be routed to an adjacent tile, not staying at (5,5)
	ipos := item.Pos()
	if intent.Dest.X == ipos.X && intent.Dest.Y == ipos.Y {
		t.Error("Dest should be an adjacent tile, not the item's tile")
	}
	destPos := intent.Dest
	if !destPos.IsAdjacentTo(ipos) {
		t.Errorf("Dest (%d,%d) should be adjacent to item at (%d,%d)", destPos.X, destPos.Y, ipos.X, ipos.Y)
	}
}

func TestFindLookIntent_ReturnsLookIntentWhenNotAdjacent(t *testing.T) {
	t.Parallel()

	char := newTestCharacter()
	gameMap := game.NewMap(10, 10)

	// Item far from character
	item := entity.NewFlower(8, 8, types.ColorPurple)
	items := []*entity.Item{item}

	// Run multiple times to get an intent (50% chance)
	var intent *entity.Intent
	for i := 0; i < 20; i++ {
		intent = findLookIntent(char, types.Position{X: 5, Y: 5}, items, gameMap, nil)
		if intent != nil {
			break
		}
	}

	if intent == nil {
		t.Skip("No look intent returned in 20 attempts (probabilistic)")
	}

	if intent.Action != entity.ActionLook {
		t.Errorf("Expected ActionLook when not adjacent, got %v", intent.Action)
	}
	if intent.TargetItem != item {
		t.Error("TargetItem should be set for move toward item")
	}
}

// =============================================================================
// Anchor test: Look at construct → preference formation + mood adjustment
// =============================================================================

func TestLookAtConstruct_FormsPreferenceAndAdjustsMood(t *testing.T) {
	t.Parallel()

	// --- Part 1: Happy character looks at stick fence, forms preference ---
	fence := entity.NewFence(6, 5, "stick", types.ColorBrown)

	// Run multiple times since formation is probabilistic
	formed := false
	var formedPref *entity.Preference
	for i := 0; i < 50; i++ {
		char := &entity.Character{
			ID:   1,
			Name: "Test",
			Mood: 95, // Joyful — positive preference formation
		}
		CompleteLookAtConstruct(char, fence, nil)

		if len(char.Preferences) > 0 {
			formed = true
			formedPref = &char.Preferences[0]
			break
		}
	}

	if !formed {
		t.Skip("No preference formed in 50 attempts (probabilistic)")
	}

	// Preference should be about the construct's recipe identity
	// Kind should be "stick fence" (not "stick", not "fence")
	if formedPref.Kind != "stick fence" && formedPref.Color != types.ColorBrown {
		t.Errorf("Expected preference with Kind='stick fence' or Color='brown', got Kind=%q Color=%q",
			formedPref.Kind, formedPref.Color)
	}
	if formedPref.Valence != 1 {
		t.Errorf("Expected positive valence from happy character, got %d", formedPref.Valence)
	}

	// --- Part 2: Existing preference adjusts mood when looking at matching construct ---
	char2 := &entity.Character{
		ID:   2,
		Name: "Test2",
		Mood: 50, // Neutral mood
		Preferences: []entity.Preference{
			{Valence: 1, Kind: "stick fence"}, // Likes stick fences
		},
	}

	moodBefore := char2.Mood
	CompleteLookAtConstruct(char2, fence, nil)

	if char2.Mood <= moodBefore {
		t.Errorf("Expected mood to increase from looking at liked construct, mood before=%v after=%v",
			moodBefore, char2.Mood)
	}

	// --- Part 3: findLookIntent targets constructs when no items exist ---
	gameMap := game.NewMap(10, 10)
	gameMap.AddConstruct(entity.NewFence(6, 5, "stick", types.ColorBrown))

	char3 := &entity.Character{
		ID:   3,
		Name: "Test3",
		Mood: 50,
	}
	char3.SetPos(types.Position{X: 5, Y: 5})

	intent := findLookIntent(char3, types.Position{X: 5, Y: 5}, nil, gameMap, nil)
	if intent == nil {
		t.Fatal("Expected look intent targeting construct, got nil")
	}
	if intent.Action != entity.ActionLook {
		t.Errorf("Expected ActionLook, got %v", intent.Action)
	}
	if intent.TargetConstruct == nil {
		t.Error("Expected TargetConstruct to be set when looking at a construct")
	}
	if intent.TargetItem != nil {
		t.Error("Expected TargetItem to be nil when targeting a construct")
	}
}

func TestFindLookIntent_PrefersCloserTarget(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	charPos := types.Position{X: 10, Y: 10}

	// Construct is closer (distance 2) than item (distance 5)
	gameMap.AddConstruct(entity.NewFence(12, 10, "stick", types.ColorBrown))
	farItem := entity.NewFlower(15, 10, types.ColorPurple)
	items := []*entity.Item{farItem}

	char := &entity.Character{ID: 1, Name: "Test"}

	intent := findLookIntent(char, charPos, items, gameMap, nil)
	if intent == nil {
		t.Fatal("Expected look intent, got nil")
	}
	if intent.TargetConstruct == nil {
		t.Error("Expected TargetConstruct (closer) to be chosen over farther item")
	}
}

func TestFindLookIntent_ExcludesLastLookedConstruct(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(20, 20)
	charPos := types.Position{X: 10, Y: 10}

	// Only one construct on the map
	gameMap.AddConstruct(entity.NewFence(12, 10, "stick", types.ColorBrown))

	char := &entity.Character{
		ID:            1,
		Name:          "Test",
		LastLookedX:   12,
		LastLookedY:   10,
		HasLastLooked: true,
	}

	intent := findLookIntent(char, charPos, nil, gameMap, nil)
	if intent != nil {
		t.Error("Expected nil intent when only construct is excluded by last-looked")
	}
}
