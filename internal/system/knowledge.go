package system

import (
	"math/rand"

	"petri/internal/entity"
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
