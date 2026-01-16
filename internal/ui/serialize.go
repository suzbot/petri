package ui

import (
	"time"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
	"petri/internal/save"
	"petri/internal/system"
	"petri/internal/types"
)

// ToSaveState converts the current game state to a SaveState for serialization
func (m Model) ToSaveState() *save.SaveState {
	state := &save.SaveState{
		Version:         save.CurrentVersion,
		SavedAt:         time.Now(),
		ElapsedGameTime: m.elapsedGameTime,
		MapWidth:        m.gameMap.Width,
		MapHeight:       m.gameMap.Height,
		Varieties:       varietiesToSave(m.gameMap.Varieties()),
		Characters:      charactersToSave(m.gameMap.Characters()),
		Items:           itemsToSave(m.gameMap.Items()),
		Features:        featuresToSave(m.gameMap.Features()),
		ActionLogs:      actionLogsToSave(m.actionLog),
	}
	return state
}

// varietiesToSave converts the variety registry to save format
func varietiesToSave(registry *game.VarietyRegistry) []save.VarietySave {
	if registry == nil {
		return nil
	}
	varieties := registry.AllVarieties()
	result := make([]save.VarietySave, len(varieties))
	for i, v := range varieties {
		result[i] = save.VarietySave{
			ItemType:  v.ItemType,
			Color:     string(v.Color),
			Pattern:   string(v.Pattern),
			Texture:   string(v.Texture),
			Poisonous: v.Poisonous,
			Healing:   v.Healing,
		}
	}
	return result
}

// charactersToSave converts characters to save format
func charactersToSave(characters []*entity.Character) []save.CharacterSave {
	result := make([]save.CharacterSave, len(characters))
	for i, c := range characters {
		talkingWithID := -1
		if c.TalkingWith != nil {
			talkingWithID = c.TalkingWith.ID
		}

		result[i] = save.CharacterSave{
			ID:   c.ID,
			Name: c.Name,
			X:    c.X,
			Y:    c.Y,

			Health: c.Health,
			Hunger: c.Hunger,
			Thirst: c.Thirst,
			Energy: c.Energy,
			Mood:   c.Mood,

			Poisoned:    c.Poisoned,
			PoisonTimer: c.PoisonTimer,
			IsDead:      c.IsDead,
			IsSleeping:  c.IsSleeping,
			AtBed:       c.AtBed,

			IsFrustrated:      c.IsFrustrated,
			FrustrationTimer:  c.FrustrationTimer,
			FailedIntentCount: c.FailedIntentCount,

			IdleCooldown:  c.IdleCooldown,
			LastLookedX:   c.LastLookedX,
			LastLookedY:   c.LastLookedY,
			HasLastLooked: c.HasLastLooked,

			TalkingWithID: talkingWithID,
			TalkTimer:     c.TalkTimer,

			HungerCooldown: c.HungerCooldown,
			ThirstCooldown: c.ThirstCooldown,
			EnergyCooldown: c.EnergyCooldown,

			ActionProgress:   c.ActionProgress,
			SpeedAccumulator: c.SpeedAccumulator,
			CurrentActivity:  c.CurrentActivity,

			Preferences: preferencesToSave(c.Preferences),
			Knowledge:   knowledgeToSave(c.Knowledge),
		}
	}
	return result
}

// preferencesToSave converts preferences to save format
func preferencesToSave(prefs []entity.Preference) []save.PreferenceSave {
	result := make([]save.PreferenceSave, len(prefs))
	for i, p := range prefs {
		result[i] = save.PreferenceSave{
			ItemType: p.ItemType,
			Color:    string(p.Color),
			Pattern:  string(p.Pattern),
			Texture:  string(p.Texture),
			Valence:  p.Valence,
		}
	}
	return result
}

// knowledgeToSave converts knowledge to save format
func knowledgeToSave(knowledge []entity.Knowledge) []save.KnowledgeSave {
	result := make([]save.KnowledgeSave, len(knowledge))
	for i, k := range knowledge {
		result[i] = save.KnowledgeSave{
			Category: string(k.Category),
			ItemType: k.ItemType,
			Color:    string(k.Color),
			Pattern:  string(k.Pattern),
			Texture:  string(k.Texture),
		}
	}
	return result
}

// itemsToSave converts items to save format
func itemsToSave(items []*entity.Item) []save.ItemSave {
	result := make([]save.ItemSave, len(items))
	for i, item := range items {
		result[i] = save.ItemSave{
			ID:         item.ID,
			X:          item.X,
			Y:          item.Y,
			ItemType:   item.ItemType,
			Color:      string(item.Color),
			Pattern:    string(item.Pattern),
			Texture:    string(item.Texture),
			Edible:     item.Edible,
			Poisonous:  item.Poisonous,
			Healing:    item.Healing,
			SpawnTimer: item.SpawnTimer,
			DeathTimer: item.DeathTimer,
		}
	}
	return result
}

// featuresToSave converts features to save format
func featuresToSave(features []*entity.Feature) []save.FeatureSave {
	result := make([]save.FeatureSave, len(features))
	for i, f := range features {
		result[i] = save.FeatureSave{
			ID:          f.ID,
			X:           f.X,
			Y:           f.Y,
			FeatureType: int(f.FType),
			DrinkSource: f.DrinkSource,
			Bed:         f.Bed,
		}
	}
	return result
}

// FromSaveState creates a Model from a SaveState
func FromSaveState(state *save.SaveState, worldID string, testCfg TestConfig) Model {
	// Create base model
	m := Model{
		phase:            phasePlaying,
		actionLog:        system.NewActionLog(200),
		width:            80,
		height:           40,
		paused:           true, // Start paused when loading
		testCfg:          testCfg,
		elapsedGameTime:  state.ElapsedGameTime,
		lastUpdate:       time.Now(),
		worldID:          worldID,
		lastSaveGameTime: state.ElapsedGameTime, // Treat load time as last save
	}

	// Create map
	m.gameMap = game.NewMap(state.MapWidth, state.MapHeight)

	// Restore variety registry
	registry := varietiesFromSave(state.Varieties)
	m.gameMap.SetVarieties(registry)

	// Build character ID map for TalkingWith resolution
	charByID := make(map[int]*entity.Character)

	// Restore characters (first pass - create all characters)
	for _, cs := range state.Characters {
		char := characterFromSave(cs)
		m.gameMap.AddCharacter(char)
		charByID[char.ID] = char
	}

	// Resolve TalkingWith references (second pass)
	for _, cs := range state.Characters {
		if cs.TalkingWithID >= 0 {
			if char := charByID[cs.ID]; char != nil {
				char.TalkingWith = charByID[cs.TalkingWithID]
			}
		}
	}

	// Track max IDs for counter restoration
	maxItemID := 0
	maxFeatureID := 0

	// Restore items (without auto-assigning IDs)
	for _, is := range state.Items {
		item := itemFromSave(is)
		m.gameMap.AddItemDirect(item)
		if is.ID > maxItemID {
			maxItemID = is.ID
		}
	}

	// Restore features (without auto-assigning IDs)
	for _, fs := range state.Features {
		feature := featureFromSave(fs)
		m.gameMap.AddFeatureDirect(feature)
		if fs.ID > maxFeatureID {
			maxFeatureID = fs.ID
		}
	}

	// Set ID counters to max + 1 for future spawns
	m.gameMap.SetNextItemID(maxItemID)
	m.gameMap.SetNextFeatureID(maxFeatureID)

	// Restore action logs
	m.actionLog.SetAllLogs(actionLogsFromSave(state.ActionLogs))
	m.actionLog.SetGameTime(state.ElapsedGameTime)

	// Set cursor to first character position if any
	chars := m.gameMap.Characters()
	if len(chars) > 0 {
		m.cursorX, m.cursorY = chars[0].Position()
	}

	return m
}

// varietiesFromSave converts saved varieties back to a registry
func varietiesFromSave(varieties []save.VarietySave) *game.VarietyRegistry {
	registry := game.NewVarietyRegistry()
	for _, vs := range varieties {
		v := &entity.ItemVariety{
			ID:        entity.GenerateVarietyID(vs.ItemType, types.Color(vs.Color), types.Pattern(vs.Pattern), types.Texture(vs.Texture)),
			ItemType:  vs.ItemType,
			Color:     types.Color(vs.Color),
			Pattern:   types.Pattern(vs.Pattern),
			Texture:   types.Texture(vs.Texture),
			Poisonous: vs.Poisonous,
			Healing:   vs.Healing,
		}
		registry.Register(v)
	}
	return registry
}

// characterFromSave converts a saved character back to an entity
func characterFromSave(cs save.CharacterSave) *entity.Character {
	char := &entity.Character{
		ID:   cs.ID,
		Name: cs.Name,

		Health: cs.Health,
		Hunger: cs.Hunger,
		Thirst: cs.Thirst,
		Energy: cs.Energy,
		Mood:   cs.Mood,

		Poisoned:    cs.Poisoned,
		PoisonTimer: cs.PoisonTimer,
		IsDead:      cs.IsDead,
		IsSleeping:  cs.IsSleeping,
		AtBed:       cs.AtBed,

		IsFrustrated:      cs.IsFrustrated,
		FrustrationTimer:  cs.FrustrationTimer,
		FailedIntentCount: cs.FailedIntentCount,

		IdleCooldown:  cs.IdleCooldown,
		LastLookedX:   cs.LastLookedX,
		LastLookedY:   cs.LastLookedY,
		HasLastLooked: cs.HasLastLooked,

		TalkTimer: cs.TalkTimer,

		HungerCooldown: cs.HungerCooldown,
		ThirstCooldown: cs.ThirstCooldown,
		EnergyCooldown: cs.EnergyCooldown,

		ActionProgress:   cs.ActionProgress,
		SpeedAccumulator: cs.SpeedAccumulator,
		CurrentActivity:  cs.CurrentActivity,

		Preferences: preferencesFromSave(cs.Preferences),
		Knowledge:   knowledgeFromSave(cs.Knowledge),
	}

	// Set position and symbol via BaseEntity
	char.X = cs.X
	char.Y = cs.Y
	char.Sym = config.CharRobot
	char.EType = entity.TypeCharacter

	return char
}

// preferencesFromSave converts saved preferences back to entities
func preferencesFromSave(prefs []save.PreferenceSave) []entity.Preference {
	result := make([]entity.Preference, len(prefs))
	for i, ps := range prefs {
		result[i] = entity.Preference{
			ItemType: ps.ItemType,
			Color:    types.Color(ps.Color),
			Pattern:  types.Pattern(ps.Pattern),
			Texture:  types.Texture(ps.Texture),
			Valence:  ps.Valence,
		}
	}
	return result
}

// knowledgeFromSave converts saved knowledge back to entities
func knowledgeFromSave(knowledge []save.KnowledgeSave) []entity.Knowledge {
	result := make([]entity.Knowledge, len(knowledge))
	for i, ks := range knowledge {
		result[i] = entity.Knowledge{
			Category: entity.KnowledgeCategory(ks.Category),
			ItemType: ks.ItemType,
			Color:    types.Color(ks.Color),
			Pattern:  types.Pattern(ks.Pattern),
			Texture:  types.Texture(ks.Texture),
		}
	}
	return result
}

// itemFromSave converts a saved item back to an entity
func itemFromSave(is save.ItemSave) *entity.Item {
	item := &entity.Item{
		ID:         is.ID,
		ItemType:   is.ItemType,
		Color:      types.Color(is.Color),
		Pattern:    types.Pattern(is.Pattern),
		Texture:    types.Texture(is.Texture),
		Edible:     is.Edible,
		Poisonous:  is.Poisonous,
		Healing:    is.Healing,
		SpawnTimer: is.SpawnTimer,
		DeathTimer: is.DeathTimer,
	}
	item.X = is.X
	item.Y = is.Y
	item.EType = entity.TypeItem

	// Set display symbol based on item type
	switch is.ItemType {
	case "berry":
		item.Sym = config.CharBerry
	case "mushroom":
		item.Sym = config.CharMushroom
	case "flower":
		item.Sym = config.CharFlower
	}

	return item
}

// featureFromSave converts a saved feature back to an entity
func featureFromSave(fs save.FeatureSave) *entity.Feature {
	feature := &entity.Feature{
		ID:          fs.ID,
		FType:       entity.FeatureType(fs.FeatureType),
		DrinkSource: fs.DrinkSource,
		Bed:         fs.Bed,
	}
	feature.X = fs.X
	feature.Y = fs.Y
	feature.EType = entity.TypeFeature

	// Set display symbol based on feature type
	if fs.DrinkSource {
		feature.Sym = config.CharSpring
	} else if fs.Bed {
		feature.Sym = config.CharLeafPile
	}

	return feature
}

// actionLogsFromSave converts saved action logs back to Event format
func actionLogsFromSave(logs map[int][]save.EventSave) map[int][]system.Event {
	result := make(map[int][]system.Event)
	for charID, savedEvents := range logs {
		events := make([]system.Event, len(savedEvents))
		for i, se := range savedEvents {
			events[i] = system.Event{
				GameTime: se.GameTime,
				CharID:   se.CharID,
				CharName: se.CharName,
				Type:     se.Type,
				Message:  se.Message,
			}
		}
		result[charID] = events
	}
	return result
}

// actionLogsToSave converts action logs to save format
func actionLogsToSave(log *system.ActionLog) map[int][]save.EventSave {
	result := make(map[int][]save.EventSave)

	allLogs := log.AllLogs()
	for charID, events := range allLogs {
		savedEvents := make([]save.EventSave, len(events))
		for i, e := range events {
			savedEvents[i] = save.EventSave{
				GameTime: e.GameTime,
				CharID:   e.CharID,
				CharName: e.CharName,
				Type:     e.Type,
				Message:  e.Message,
			}
		}
		result[charID] = savedEvents
	}
	return result
}
