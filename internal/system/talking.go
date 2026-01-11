package system

import (
	"math/rand"
	"strings"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
)

// LearnKnowledgeWithEffects teaches knowledge to a character and applies all side effects.
// This is the single entry point for all knowledge learning (eating, talking, etc.).
// Returns true if the knowledge was new and learned.
// Side effects:
// - Logs "Learned something!" if new knowledge
// - If poison knowledge: forms dislike preference
func LearnKnowledgeWithEffects(char *entity.Character, knowledge entity.Knowledge, log *ActionLog) bool {
	if !char.LearnKnowledge(knowledge) {
		return false // Already knew this
	}

	// Log the learning
	if log != nil {
		log.Add(char.ID, char.Name, "learning", "Learned something!")
	}

	// Side effect: poison knowledge creates dislike preference
	if knowledge.Category == entity.KnowledgePoisonous {
		formDislikeFromPoisonKnowledge(char, knowledge, log)
	}

	return true
}

// formDislikeFromPoisonKnowledge creates a dislike preference from poison knowledge.
func formDislikeFromPoisonKnowledge(char *entity.Character, knowledge entity.Knowledge, log *ActionLog) {
	// Create full-variety dislike preference from knowledge attributes
	candidate := entity.Preference{
		Valence:  -1,
		ItemType: knowledge.ItemType,
		Color:    knowledge.Color,
		Pattern:  knowledge.Pattern,
		Texture:  knowledge.Texture,
	}

	// Check for existing preference with exact match
	for i, existing := range char.Preferences {
		if existing.ExactMatch(candidate) {
			if existing.Valence == candidate.Valence {
				// Same dislike already exists - no change
				return
			}
			// Opposite valence (like) - remove existing preference
			char.Preferences = append(char.Preferences[:i], char.Preferences[i+1:]...)
			logPreferenceRemoved(char, existing, log)
			return
		}
	}

	// No existing match - add new dislike
	char.Preferences = append(char.Preferences, candidate)
	logPreferenceFormed(char, candidate, log)
}

// selectIdleActivity randomly selects an idle activity (looking, talking, or staying idle).
// Returns nil if cooldown is active or if the selected activity cannot be performed.
// Sets IdleCooldown after being called (regardless of outcome).
func selectIdleActivity(char *entity.Character, cx, cy int, items []*entity.Item, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Check cooldown
	if char.IdleCooldown > 0 {
		return nil
	}

	// Set cooldown for next attempt
	char.IdleCooldown = config.IdleCooldown

	// Roll 0-2 for activity selection (equal 1/3 probability each)
	roll := rand.Intn(3)

	switch roll {
	case 0:
		// Try looking
		if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try other activities
		if intent := findTalkIntent(char, cx, cy, gameMap, log); intent != nil {
			return intent
		}
	case 1:
		// Try talking
		if intent := findTalkIntent(char, cx, cy, gameMap, log); intent != nil {
			return intent
		}
		// Fall through to try looking
		if intent := findLookIntent(char, cx, cy, items, gameMap, log); intent != nil {
			return intent
		}
	case 2:
		// Stay idle - return nil
		return nil
	}

	// Nothing available
	return nil
}

// StartTalking sets up both characters for a conversation.
// Sets TalkingWith pointers, TalkTimer, and CurrentActivity for both.
func StartTalking(initiator, target *entity.Character, log *ActionLog) {
	initiator.TalkingWith = target
	target.TalkingWith = initiator

	initiator.TalkTimer = config.TalkDuration
	target.TalkTimer = config.TalkDuration

	initiator.CurrentActivity = "Talking with " + target.Name
	target.CurrentActivity = "Talking with " + initiator.Name

	// Set intent for target so they also continue talking
	target.Intent = &entity.Intent{
		Action:          entity.ActionTalk,
		TargetCharacter: initiator,
	}

	if log != nil {
		log.Add(initiator.ID, initiator.Name, "activity", "Started talking with "+target.Name)
		log.Add(target.ID, target.Name, "activity", "Started talking with "+initiator.Name)
	}
}

// StopTalking clears talking state for both characters.
// Called when talk completes or is interrupted.
func StopTalking(char1, char2 *entity.Character, log *ActionLog) {
	char1.TalkingWith = nil
	char1.TalkTimer = 0
	char1.Intent = nil
	char1.IdleCooldown = config.IdleCooldown

	char2.TalkingWith = nil
	char2.TalkTimer = 0
	char2.Intent = nil
	char2.IdleCooldown = config.IdleCooldown
}

// isIdleActivity returns true if the activity string represents an idle activity
// that can be interrupted for talking. Idle activities are: Idle, Looking, Talking.
func isIdleActivity(activity string) bool {
	if strings.HasPrefix(activity, "Idle") {
		return true
	}
	if strings.HasPrefix(activity, "Looking") {
		return true
	}
	if strings.HasPrefix(activity, "Talking") {
		return true
	}
	return false
}

// TransmitKnowledge allows two characters to share knowledge after completing a conversation.
// Each character picks one random piece of knowledge to share with their partner.
// If the partner doesn't already have that knowledge, they learn it.
func TransmitKnowledge(char1, char2 *entity.Character, log *ActionLog) {
	// Select knowledge to share BEFORE any transfers (so we pick from original sets)
	var k1 *entity.Knowledge
	var k2 *entity.Knowledge

	if len(char1.Knowledge) > 0 {
		idx := rand.Intn(len(char1.Knowledge))
		k1 = &char1.Knowledge[idx]
	}
	if len(char2.Knowledge) > 0 {
		idx := rand.Intn(len(char2.Knowledge))
		k2 = &char2.Knowledge[idx]
	}

	// Now perform the transfers
	if k1 != nil {
		teachKnowledge(char1, char2, *k1, log)
	}
	if k2 != nil {
		teachKnowledge(char2, char1, *k2, log)
	}
}

// teachKnowledge attempts to teach a specific piece of knowledge from sharer to learner.
func teachKnowledge(sharer, learner *entity.Character, knowledge entity.Knowledge, log *ActionLog) {
	if LearnKnowledgeWithEffects(learner, knowledge, log) {
		// Successfully learned - log sharing
		if log != nil {
			log.Add(sharer.ID, sharer.Name, "knowledge", "Shared knowledge with "+learner.Name)
			log.Add(learner.ID, learner.Name, "knowledge", "Learned: "+knowledge.Description())
		}
	}
}

// findTalkIntent creates an intent to talk with the closest idle character.
// Returns nil if no suitable conversation partner is available.
func findTalkIntent(char *entity.Character, cx, cy int, gameMap *game.Map, log *ActionLog) *entity.Intent {
	// Find all characters doing idle activities
	var candidates []*entity.Character
	for _, other := range gameMap.Characters() {
		// Skip self
		if other == char {
			continue
		}
		// Skip dead or sleeping characters
		if other.IsDead || other.IsSleeping {
			continue
		}
		// Only target characters doing idle activities
		if !isIdleActivity(other.CurrentActivity) {
			continue
		}
		candidates = append(candidates, other)
	}

	if len(candidates) == 0 {
		return nil
	}

	// Find closest idle character
	var closest *entity.Character
	closestDist := int(^uint(0) >> 1) // Max int

	for _, other := range candidates {
		ox, oy := other.Position()
		dist := abs(cx-ox) + abs(cy-oy)
		if dist < closestDist {
			closestDist = dist
			closest = other
		}
	}

	if closest == nil {
		return nil
	}

	tx, ty := closest.Position()

	// If adjacent, start talking immediately
	if isAdjacent(cx, cy, tx, ty) {
		newActivity := "Talking with " + closest.Name
		if char.CurrentActivity != newActivity {
			char.CurrentActivity = newActivity
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Started talking with "+closest.Name)
			}
		}
		return &entity.Intent{
			TargetX:         cx, // Stay in place
			TargetY:         cy,
			Action:          entity.ActionTalk,
			TargetCharacter: closest,
		}
	}

	// Not adjacent - move toward target
	nx, ny := nextStep(cx, cy, tx, ty)

	newActivity := "Moving to talk with " + closest.Name
	if char.CurrentActivity != newActivity {
		char.CurrentActivity = newActivity
		if log != nil {
			log.Add(char.ID, char.Name, "movement", "Moving to talk with "+closest.Name)
		}
	}

	return &entity.Intent{
		TargetX:         nx,
		TargetY:         ny,
		Action:          entity.ActionMove,
		TargetCharacter: closest,
	}
}
