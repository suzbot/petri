package system

import (
	"testing"

	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/types"
)

// =============================================================================
// findTalkIntent
// =============================================================================

func TestFindTalkIntent_ReturnsNilWithNoOtherCharacters(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent != nil {
		t.Error("Should return nil when no other characters exist")
	}
}

func TestFindTalkIntent_ReturnsNilWhenOnlyNonIdleCharacters(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	other := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	other.IsSleeping = true // Not idle - sleeping

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(other)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent != nil {
		t.Error("Should return nil when other characters are not doing idle activities")
	}
}

func TestFindTalkIntent_ReturnsTalkIntentWhenAdjacent(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	other := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	other.CurrentActivity = "Idle"

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(other)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Should return an intent when idle character is adjacent")
	}
	if intent.Action != entity.ActionTalk {
		t.Errorf("Expected ActionTalk, got %v", intent.Action)
	}
	if intent.TargetCharacter != other {
		t.Error("TargetCharacter should be the adjacent idle character")
	}
}

func TestFindTalkIntent_ReturnsMoveIntentWhenNotAdjacent(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	other := entity.NewCharacter(2, 8, 8, "Bob", "mushroom", types.ColorBlue)
	other.CurrentActivity = "Idle"

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(other)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Should return an intent when idle character exists")
	}
	if intent.Action != entity.ActionMove {
		t.Errorf("Expected ActionMove when not adjacent, got %v", intent.Action)
	}
	if intent.TargetCharacter != other {
		t.Error("TargetCharacter should be set for move toward target")
	}
}

func TestFindTalkIntent_FindsClosestIdleCharacter(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	far := entity.NewCharacter(2, 9, 9, "Far", "mushroom", types.ColorBlue)
	far.CurrentActivity = "Idle"
	near := entity.NewCharacter(3, 6, 6, "Near", "berry", types.ColorWhite)
	near.CurrentActivity = "Idle"

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(far)
	gameMap.AddCharacter(near)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Should return an intent")
	}
	if intent.TargetCharacter != near {
		t.Error("Should target the closest idle character")
	}
}

func TestFindTalkIntent_TargetsCharacterLooking(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	other := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	other.CurrentActivity = "Looking at Purple Flower"

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(other)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Should return an intent when character is looking")
	}
	if intent.TargetCharacter != other {
		t.Error("Should target character who is looking")
	}
}

func TestFindTalkIntent_TargetsCharacterTalking(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	other := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	other.CurrentActivity = "Talking with Carol"

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(other)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent == nil {
		t.Fatal("Should return an intent when character is talking")
	}
	if intent.TargetCharacter != other {
		t.Error("Should target character who is talking")
	}
}

func TestFindTalkIntent_SkipsDeadCharacters(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	dead := entity.NewCharacter(2, 6, 5, "Dead", "mushroom", types.ColorBlue)
	dead.CurrentActivity = "Idle"
	dead.IsDead = true

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(dead)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent != nil {
		t.Error("Should not target dead characters")
	}
}

func TestFindTalkIntent_SkipsSleepingCharacters(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	sleeping := entity.NewCharacter(2, 6, 5, "Sleepy", "mushroom", types.ColorBlue)
	sleeping.IsSleeping = true

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(sleeping)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent != nil {
		t.Error("Should not target sleeping characters")
	}
}

func TestFindTalkIntent_SkipsCharactersWithActiveNeeds(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	busy := entity.NewCharacter(2, 6, 5, "Busy", "mushroom", types.ColorBlue)
	busy.CurrentActivity = "Moving to berry" // Not an idle activity

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)
	gameMap.AddCharacter(busy)

	intent := findTalkIntent(char, 5, 5, gameMap, nil)

	if intent != nil {
		t.Error("Should not target characters with active needs")
	}
}

// =============================================================================
// isIdleActivity helper
// =============================================================================

func TestIsIdleActivity_ReturnsTrueForIdleActivities(t *testing.T) {
	t.Parallel()

	idleActivities := []string{
		"Idle",
		"Idle (no needs)",
		"Looking at Red Berry",
		"Talking with Bob",
	}

	for _, activity := range idleActivities {
		if !isIdleActivity(activity) {
			t.Errorf("Activity %q should be considered idle", activity)
		}
	}
}

func TestIsIdleActivity_ReturnsFalseForNonIdleActivities(t *testing.T) {
	t.Parallel()

	nonIdleActivities := []string{
		"Moving to berry",
		"Moving to spring",
		"Drinking",
		"Sleeping (in bed)",
		"Sleeping (on ground)",
		"Frustrated",
	}

	for _, activity := range nonIdleActivities {
		if isIdleActivity(activity) {
			t.Errorf("Activity %q should not be considered idle", activity)
		}
	}
}

// =============================================================================
// selectIdleActivity
// =============================================================================

func TestSelectIdleActivity_ReturnsNilWhenCooldownActive(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	char.IdleCooldown = 5.0 // Cooldown active

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)

	// Add an item so looking would be possible
	item := entity.NewFlower(6, 5, types.ColorPurple)
	items := []*entity.Item{item}

	intent := selectIdleActivity(char, 5, 5, items, gameMap, nil)

	if intent != nil {
		t.Error("Should return nil when cooldown is active")
	}
}

func TestSelectIdleActivity_SetsCooldownWhenCalled(t *testing.T) {
	t.Parallel()

	char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	char.IdleCooldown = 0 // No cooldown

	gameMap := game.NewMap(10, 10)
	gameMap.AddCharacter(char)

	items := []*entity.Item{}

	// Call multiple times - cooldown should be set
	selectIdleActivity(char, 5, 5, items, gameMap, nil)

	if char.IdleCooldown <= 0 {
		t.Error("Should set IdleCooldown after being called")
	}
}

func TestSelectIdleActivity_ReturnsVariedIntents(t *testing.T) {
	t.Parallel()

	// Run multiple times and track what we get
	lookCount := 0
	talkCount := 0
	idleCount := 0

	for i := 0; i < 90; i++ {
		gameMap := game.NewMap(20, 20)

		// Add an item for looking
		item := entity.NewFlower(6, 5, types.ColorPurple)
		items := []*entity.Item{item}

		// Add another character for talking
		other := entity.NewCharacter(2, 7, 5, "Bob", "mushroom", types.ColorBlue)
		other.CurrentActivity = "Idle"
		gameMap.AddCharacter(other)

		char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
		char.IdleCooldown = 0
		gameMap.AddCharacter(char)

		intent := selectIdleActivity(char, 5, 5, items, gameMap, nil)

		if intent == nil {
			idleCount++
		} else if intent.TargetItem != nil {
			lookCount++
		} else if intent.TargetCharacter != nil {
			talkCount++
		}
	}

	// With 1/3 probability each over 90 trials, we expect ~30 of each
	// Allow for variance - each should be at least 5 (very conservative)
	if lookCount < 5 {
		t.Errorf("Expected some looking intents, got %d", lookCount)
	}
	if talkCount < 5 {
		t.Errorf("Expected some talking intents, got %d", talkCount)
	}
	if idleCount < 5 {
		t.Errorf("Expected some idle outcomes, got %d", idleCount)
	}
}

func TestSelectIdleActivity_FallsBackWhenLookingNotPossible(t *testing.T) {
	t.Parallel()

	// No items - looking not possible
	items := []*entity.Item{}

	// Run multiple times - should never get a looking intent
	for i := 0; i < 30; i++ {
		gameMap := game.NewMap(10, 10)

		// Add another character for talking
		other := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
		other.CurrentActivity = "Idle"
		gameMap.AddCharacter(other)

		char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
		char.IdleCooldown = 0
		gameMap.AddCharacter(char)

		intent := selectIdleActivity(char, 5, 5, items, gameMap, nil)

		if intent != nil && intent.TargetItem != nil {
			t.Error("Should not return looking intent when no items exist")
		}
	}
}

func TestSelectIdleActivity_FallsBackWhenTalkingNotPossible(t *testing.T) {
	t.Parallel()

	// Add an item for looking
	item := entity.NewFlower(6, 5, types.ColorPurple)
	items := []*entity.Item{item}

	// No other characters - talking not possible

	// Run multiple times - should never get a talking intent
	for i := 0; i < 30; i++ {
		gameMap := game.NewMap(10, 10)

		char := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
		char.IdleCooldown = 0
		gameMap.AddCharacter(char)

		intent := selectIdleActivity(char, 5, 5, items, gameMap, nil)

		if intent != nil && intent.TargetCharacter != nil {
			t.Error("Should not return talking intent when no other characters exist")
		}
	}
}

// =============================================================================
// StartTalking
// =============================================================================

func TestStartTalking_SetsBothCharactersToTalkingState(t *testing.T) {
	t.Parallel()

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)

	StartTalking(alice, bob, nil)

	if alice.TalkingWith != bob {
		t.Error("Alice should have TalkingWith pointing to Bob")
	}
	if bob.TalkingWith != alice {
		t.Error("Bob should have TalkingWith pointing to Alice")
	}
}

func TestStartTalking_SetsTalkTimerForBoth(t *testing.T) {
	t.Parallel()

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)

	StartTalking(alice, bob, nil)

	if alice.TalkTimer <= 0 {
		t.Error("Alice should have TalkTimer set")
	}
	if bob.TalkTimer <= 0 {
		t.Error("Bob should have TalkTimer set")
	}
}

func TestStartTalking_UpdatesCurrentActivity(t *testing.T) {
	t.Parallel()

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)

	StartTalking(alice, bob, nil)

	if alice.CurrentActivity != "Talking with Bob" {
		t.Errorf("Alice CurrentActivity should be 'Talking with Bob', got %q", alice.CurrentActivity)
	}
	if bob.CurrentActivity != "Talking with Alice" {
		t.Errorf("Bob CurrentActivity should be 'Talking with Alice', got %q", bob.CurrentActivity)
	}
}

// =============================================================================
// StopTalking
// =============================================================================

func TestStopTalking_ClearsTalkingState(t *testing.T) {
	t.Parallel()

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)

	// Start talking
	StartTalking(alice, bob, nil)

	// Stop talking
	StopTalking(alice, bob, nil)

	if alice.TalkingWith != nil {
		t.Error("Alice TalkingWith should be nil after stopping")
	}
	if bob.TalkingWith != nil {
		t.Error("Bob TalkingWith should be nil after stopping")
	}
	if alice.TalkTimer != 0 {
		t.Error("Alice TalkTimer should be 0 after stopping")
	}
	if bob.TalkTimer != 0 {
		t.Error("Bob TalkTimer should be 0 after stopping")
	}
}

func TestStopTalking_SetsIdleCooldown(t *testing.T) {
	t.Parallel()

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)

	StartTalking(alice, bob, nil)
	StopTalking(alice, bob, nil)

	if alice.IdleCooldown <= 0 {
		t.Error("Alice should have IdleCooldown set after stopping")
	}
	if bob.IdleCooldown <= 0 {
		t.Error("Bob should have IdleCooldown set after stopping")
	}
}

// =============================================================================
// Talking continuation in CalculateIntent
// =============================================================================

func TestCalculateIntent_ContinuesTalkingWhenNoUrgentNeeds(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	gameMap.AddCharacter(alice)
	gameMap.AddCharacter(bob)

	// Set up talking state
	alice.TalkingWith = bob
	alice.TalkTimer = 3.0
	alice.CurrentActivity = "Talking with Bob"
	alice.Intent = &entity.Intent{
		Action:          entity.ActionTalk,
		TargetCharacter: bob,
	}

	items := []*entity.Item{}

	// Calculate intent - should continue talking
	intent := CalculateIntent(alice, items, gameMap, nil)

	if intent == nil || intent.Action != entity.ActionTalk {
		t.Error("Should continue talking when no urgent needs")
	}
}

func TestCalculateIntent_InterruptsTalkingForModerateNeed(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	alice := entity.NewCharacter(1, 5, 5, "Alice", "berry", types.ColorRed)
	alice.Hunger = 80 // Moderate hunger
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	gameMap.AddCharacter(alice)
	gameMap.AddCharacter(bob)

	// Set up talking state
	alice.TalkingWith = bob
	alice.TalkTimer = 3.0
	alice.CurrentActivity = "Talking with Bob"
	alice.Intent = &entity.Intent{
		Action:          entity.ActionTalk,
		TargetCharacter: bob,
	}

	// Add food so hunger can be addressed
	food := entity.NewBerry(7, 5, types.ColorRed, false, false)
	items := []*entity.Item{food}

	// Calculate intent - should interrupt talking for food
	intent := CalculateIntent(alice, items, gameMap, nil)

	if intent != nil && intent.Action == entity.ActionTalk {
		t.Error("Should interrupt talking for Moderate hunger")
	}
}

// =============================================================================
// Talking approach continuation (regression test)
// =============================================================================

func TestCalculateIntent_ContinuesApproachingTalkTarget(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Alice at (2, 5), Bob at (6, 5) - not adjacent, 4 tiles apart
	alice := entity.NewCharacter(1, 2, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	bob.CurrentActivity = "Idle"
	gameMap.AddCharacter(alice)
	gameMap.AddCharacter(bob)

	items := []*entity.Item{}

	// Set up initial talking intent (as if findTalkIntent was just called)
	alice.Intent = &entity.Intent{
		TargetX:         3, // Next step toward Bob
		TargetY:         5,
		Action:          entity.ActionMove,
		TargetCharacter: bob,
	}
	alice.CurrentActivity = "Moving to talk with Bob"

	// Simulate multiple ticks - alice should keep pursuing bob
	for i := 0; i < 10; i++ {
		cx, cy := alice.Position()
		intent := CalculateIntent(alice, items, gameMap, nil)

		if intent == nil {
			t.Fatalf("Tick %d: Intent should not be nil while approaching talk target", i)
		}

		// Should either be moving toward bob or talking (if adjacent)
		if intent.Action != entity.ActionMove && intent.Action != entity.ActionTalk {
			t.Fatalf("Tick %d: Expected ActionMove or ActionTalk, got %v", i, intent.Action)
		}

		// TargetCharacter should be preserved
		if intent.TargetCharacter != bob {
			t.Fatalf("Tick %d: TargetCharacter should be Bob", i)
		}

		// Update alice's intent for next tick
		alice.Intent = intent

		// If we got ActionTalk, we've arrived - test passed
		if intent.Action == entity.ActionTalk {
			return
		}

		// Simulate alice actually moving (like applyIntent would do)
		if intent.Action == entity.ActionMove && (intent.TargetX != cx || intent.TargetY != cy) {
			gameMap.MoveCharacter(alice, intent.TargetX, intent.TargetY)
		}
	}

	t.Error("Alice should have reached Bob and started ActionTalk within 10 ticks")
}

func TestCalculateIntent_ApproachingTalkTargetReachesAndTalks(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	// Alice at (4, 5), Bob at (6, 5) - 2 tiles apart, should reach in ~2 moves
	alice := entity.NewCharacter(1, 4, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	bob.CurrentActivity = "Idle"
	gameMap.AddCharacter(alice)
	gameMap.AddCharacter(bob)

	items := []*entity.Item{}

	// Initial intent to move toward Bob
	alice.Intent = &entity.Intent{
		TargetX:         5,
		TargetY:         5,
		Action:          entity.ActionMove,
		TargetCharacter: bob,
	}

	// First tick - should continue moving
	intent1 := CalculateIntent(alice, items, gameMap, nil)
	if intent1 == nil || intent1.Action != entity.ActionMove {
		t.Fatal("First tick should continue moving toward Bob")
	}
	if intent1.TargetCharacter != bob {
		t.Fatal("First tick should preserve TargetCharacter")
	}

	// Simulate alice moving to (5, 5) - now adjacent to Bob at (6, 5)
	alice.Intent = intent1
	gameMap.MoveCharacter(alice, 5, 5)

	// Second tick - should switch to ActionTalk since now adjacent
	intent2 := CalculateIntent(alice, items, gameMap, nil)
	if intent2 == nil {
		t.Fatal("Second tick intent should not be nil")
	}
	if intent2.Action != entity.ActionTalk {
		t.Errorf("Second tick should be ActionTalk (now adjacent), got %v", intent2.Action)
	}
	if intent2.TargetCharacter != bob {
		t.Error("Second tick should have Bob as TargetCharacter")
	}
}

func TestCalculateIntent_ApproachStopsIfTargetBecomesNonIdle(t *testing.T) {
	t.Parallel()

	gameMap := game.NewMap(10, 10)

	alice := entity.NewCharacter(1, 2, 5, "Alice", "berry", types.ColorRed)
	bob := entity.NewCharacter(2, 6, 5, "Bob", "mushroom", types.ColorBlue)
	bob.CurrentActivity = "Idle"
	gameMap.AddCharacter(alice)
	gameMap.AddCharacter(bob)

	items := []*entity.Item{}

	// Alice is approaching Bob
	alice.Intent = &entity.Intent{
		TargetX:         3,
		TargetY:         5,
		Action:          entity.ActionMove,
		TargetCharacter: bob,
	}

	// Bob stops being idle (e.g., gets hungry and starts moving)
	bob.CurrentActivity = "Moving to berry"

	// Alice should abandon approach and re-evaluate
	intent := CalculateIntent(alice, items, gameMap, nil)

	// Should not still be targeting Bob
	if intent != nil && intent.TargetCharacter == bob {
		t.Error("Should stop approaching when target is no longer idle")
	}
}
