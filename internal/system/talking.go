package system

import (
	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
)

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
	nx, ny := NextStep(cx, cy, tx, ty)

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
