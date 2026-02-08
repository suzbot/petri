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
		WaterTiles:      waterTilesToSave(m.gameMap),
		ActionLogs:      actionLogsToSave(m.actionLog),
		Orders:          ordersToSave(m.orders),
		NextOrderID:     m.nextOrderID,

		GroundSpawnStick: m.groundSpawnTimers.Stick,
		GroundSpawnNut:   m.groundSpawnTimers.Nut,
		GroundSpawnShell: m.groundSpawnTimers.Shell,
	}
	return state
}

// ordersToSave converts orders to save format
func ordersToSave(orders []*entity.Order) []save.OrderSave {
	result := make([]save.OrderSave, len(orders))
	for i, o := range orders {
		result[i] = save.OrderSave{
			ID:         o.ID,
			ActivityID: o.ActivityID,
			TargetType: o.TargetType,
			Status:     string(o.Status),
			AssignedTo: o.AssignedTo,
		}
	}
	return result
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
			Edible:    v.IsEdible(),
			Poisonous: v.IsPoisonous(),
			Healing:   v.IsHealing(),
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

		// Convert inventory items
		var inventory []save.ItemSave
		if len(c.Inventory) > 0 {
			inventory = make([]save.ItemSave, len(c.Inventory))
			for idx, item := range c.Inventory {
				var plantSave *save.PlantPropertiesSave
				if item.Plant != nil {
					plantSave = &save.PlantPropertiesSave{
						IsGrowing:  item.Plant.IsGrowing,
						SpawnTimer: item.Plant.SpawnTimer,
					}
				}
				inventory[idx] = save.ItemSave{
					ID:         item.ID,
					Position:   item.Pos(),
					Name:       item.Name,
					ItemType:   item.ItemType,
					Color:      string(item.Color),
					Pattern:    string(item.Pattern),
					Texture:    string(item.Texture),
					Plant:      plantSave,
					Container:  containerToSave(item.Container),
					Edible:     item.IsEdible(),
					Poisonous:  item.IsPoisonous(),
					Healing:    item.IsHealing(),
					DeathTimer: item.DeathTimer,
				}
			}
		}

		result[i] = save.CharacterSave{
			ID:       c.ID,
			Name:     c.Name,
			Position: c.Pos(),

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

			Preferences:     preferencesToSave(c.Preferences),
			Knowledge:       knowledgeToSave(c.Knowledge),
			KnownActivities: c.KnownActivities,
			KnownRecipes:    c.KnownRecipes,

			Inventory:       inventory,
			AssignedOrderID: c.AssignedOrderID,
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
			Kind:     p.Kind,
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

// containerToSave converts ContainerData to save format
func containerToSave(container *entity.ContainerData) *save.ContainerDataSave {
	if container == nil {
		return nil
	}
	contents := make([]save.StackSave, len(container.Contents))
	for i, stack := range container.Contents {
		contents[i] = save.StackSave{
			ItemType: stack.Variety.ItemType,
			Color:    string(stack.Variety.Color),
			Pattern:  string(stack.Variety.Pattern),
			Texture:  string(stack.Variety.Texture),
			Count:    stack.Count,
		}
	}
	return &save.ContainerDataSave{
		Capacity: container.Capacity,
		Contents: contents,
	}
}

// itemsToSave converts items to save format
func itemsToSave(items []*entity.Item) []save.ItemSave {
	result := make([]save.ItemSave, len(items))
	for i, item := range items {
		var plantSave *save.PlantPropertiesSave
		if item.Plant != nil {
			plantSave = &save.PlantPropertiesSave{
				IsGrowing:  item.Plant.IsGrowing,
				SpawnTimer: item.Plant.SpawnTimer,
			}
		}
		result[i] = save.ItemSave{
			ID:       item.ID,
			Position: item.Pos(),
			Name:     item.Name,
			ItemType: item.ItemType,
			Kind:     item.Kind,
			Color:      string(item.Color),
			Pattern:    string(item.Pattern),
			Texture:    string(item.Texture),
			Plant:      plantSave,
			Container:  containerToSave(item.Container),
			Edible:     item.IsEdible(),
			Poisonous:  item.IsPoisonous(),
			Healing:    item.IsHealing(),
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
			Position:    f.Pos(),
			FeatureType: int(f.FType),
			DrinkSource: f.DrinkSource,
			Bed:         f.Bed,
			Passable:    f.Passable,
		}
	}
	return result
}

// waterTilesToSave converts water tiles to save format
func waterTilesToSave(gameMap *game.Map) []save.WaterTileSave {
	positions := gameMap.WaterPositions()
	result := make([]save.WaterTileSave, len(positions))
	for i, pos := range positions {
		result[i] = save.WaterTileSave{
			Position:  pos,
			WaterType: int(gameMap.WaterAt(pos)),
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
		speedMultiplier:  1,                     // Normal speed
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
		char := characterFromSave(cs, registry)
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
		item := itemFromSave(is, registry)
		m.gameMap.AddItemDirect(item)
		if is.ID > maxItemID {
			maxItemID = is.ID
		}
	}

	// Restore water tiles
	for _, ws := range state.WaterTiles {
		m.gameMap.AddWater(ws.Position, game.WaterType(ws.WaterType))
	}

	// Restore features (without auto-assigning IDs)
	// Migrate old spring features to water tiles
	for _, fs := range state.Features {
		if fs.DrinkSource {
			// Old save: spring was stored as feature — migrate to water tile
			m.gameMap.AddWater(fs.Position, game.WaterSpring)
			continue
		}
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

	// Restore orders
	m.orders = ordersFromSave(state.Orders)
	m.nextOrderID = state.NextOrderID
	// Ensure nextOrderID is at least 1 (ID 0 means "no order assigned")
	if m.nextOrderID < 1 {
		m.nextOrderID = 1
	}

	// Restore ground spawn timers (default to random if loading old save without them)
	m.groundSpawnTimers = system.GroundSpawnTimers{
		Stick: state.GroundSpawnStick,
		Nut:   state.GroundSpawnNut,
		Shell: state.GroundSpawnShell,
	}
	if m.groundSpawnTimers.Stick <= 0 {
		m.groundSpawnTimers.Stick = system.RandomGroundSpawnInterval()
	}
	if m.groundSpawnTimers.Nut <= 0 {
		m.groundSpawnTimers.Nut = system.RandomGroundSpawnInterval()
	}
	if m.groundSpawnTimers.Shell <= 0 {
		m.groundSpawnTimers.Shell = system.RandomGroundSpawnInterval()
	}

	// Set cursor to first character position if any
	chars := m.gameMap.Characters()
	if len(chars) > 0 {
		pos := chars[0].Pos()
		m.cursorX, m.cursorY = pos.X, pos.Y
	}

	return m
}

// ordersFromSave converts saved orders back to entities
func ordersFromSave(orders []save.OrderSave) []*entity.Order {
	result := make([]*entity.Order, len(orders))
	for i, os := range orders {
		result[i] = &entity.Order{
			ID:         os.ID,
			ActivityID: os.ActivityID,
			TargetType: os.TargetType,
			Status:     entity.OrderStatus(os.Status),
			AssignedTo: os.AssignedTo,
		}
	}
	return result
}

// varietiesFromSave converts saved varieties back to a registry
func varietiesFromSave(varieties []save.VarietySave) *game.VarietyRegistry {
	registry := game.NewVarietyRegistry()
	for _, vs := range varieties {
		var edible *entity.EdibleProperties
		if vs.Edible {
			edible = &entity.EdibleProperties{
				Poisonous: vs.Poisonous,
				Healing:   vs.Healing,
			}
		}
		v := &entity.ItemVariety{
			ID:       entity.GenerateVarietyID(vs.ItemType, types.Color(vs.Color), types.Pattern(vs.Pattern), types.Texture(vs.Texture)),
			ItemType: vs.ItemType,
			Color:    types.Color(vs.Color),
			Pattern:  types.Pattern(vs.Pattern),
			Texture:  types.Texture(vs.Texture),
			Edible:   edible,
		}
		registry.Register(v)
	}
	return registry
}

// characterFromSave converts a saved character back to an entity
// registry is optional - only needed if carried item has container with contents
func characterFromSave(cs save.CharacterSave, registry *game.VarietyRegistry) *entity.Character {
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

		Preferences:     preferencesFromSave(cs.Preferences),
		Knowledge:       knowledgeFromSave(cs.Knowledge),
		KnownActivities: cs.KnownActivities,
		KnownRecipes:    cs.KnownRecipes,
	}

	// Set position and symbol via BaseEntity
	char.X = cs.Position.X
	char.Y = cs.Position.Y
	char.Sym = config.CharRobot
	char.EType = entity.TypeCharacter

	// Restore inventory
	if len(cs.Inventory) > 0 {
		char.Inventory = make([]*entity.Item, len(cs.Inventory))
		for i, is := range cs.Inventory {
			char.Inventory[i] = itemFromSave(is, registry)
		}
	}

	// Restore assigned order
	char.AssignedOrderID = cs.AssignedOrderID

	return char
}

// preferencesFromSave converts saved preferences back to entities
func preferencesFromSave(prefs []save.PreferenceSave) []entity.Preference {
	result := make([]entity.Preference, len(prefs))
	for i, ps := range prefs {
		result[i] = entity.Preference{
			ItemType: ps.ItemType,
			Kind:     ps.Kind,
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

// containerFromSave converts saved ContainerData back to entity
// registry is used to look up variety pointers for stacks
func containerFromSave(cs *save.ContainerDataSave, registry *game.VarietyRegistry) *entity.ContainerData {
	if cs == nil {
		return nil
	}
	contents := make([]entity.Stack, len(cs.Contents))
	for i, ss := range cs.Contents {
		// Look up variety from registry by attributes
		var variety *entity.ItemVariety
		if registry != nil {
			varietyID := entity.GenerateVarietyID(
				ss.ItemType,
				types.Color(ss.Color),
				types.Pattern(ss.Pattern),
				types.Texture(ss.Texture),
			)
			variety = registry.Get(varietyID)
		}
		contents[i] = entity.Stack{
			Variety: variety,
			Count:   ss.Count,
		}
	}
	return &entity.ContainerData{
		Capacity: cs.Capacity,
		Contents: contents,
	}
}

// itemFromSave converts a saved item back to an entity
// registry is optional - only needed if item has container with contents
func itemFromSave(is save.ItemSave, registry *game.VarietyRegistry) *entity.Item {
	var plant *entity.PlantProperties
	if is.Plant != nil {
		plant = &entity.PlantProperties{
			IsGrowing:  is.Plant.IsGrowing,
			SpawnTimer: is.Plant.SpawnTimer,
		}
	}

	var edible *entity.EdibleProperties
	if is.Edible {
		edible = &entity.EdibleProperties{
			Poisonous: is.Poisonous,
			Healing:   is.Healing,
		}
	}

	item := &entity.Item{
		ID:         is.ID,
		Name:       is.Name,
		ItemType:   is.ItemType,
		Kind:       is.Kind,
		Color:      types.Color(is.Color),
		Pattern:    types.Pattern(is.Pattern),
		Texture:    types.Texture(is.Texture),
		Plant:      plant,
		Container:  containerFromSave(is.Container, registry),
		Edible:     edible,
		DeathTimer: is.DeathTimer,
	}
	item.X = is.Position.X
	item.Y = is.Position.Y
	item.EType = entity.TypeItem

	// Set display symbol based on item type
	switch is.ItemType {
	case "berry":
		item.Sym = config.CharBerry
	case "mushroom":
		item.Sym = config.CharMushroom
	case "flower":
		item.Sym = config.CharFlower
	case "gourd":
		item.Sym = config.CharGourd
	case "vessel":
		item.Sym = config.CharVessel
	case "stick":
		item.Sym = config.CharStick
	case "nut":
		item.Sym = config.CharNut
	case "shell":
		item.Sym = config.CharShell
	case "hoe":
		item.Sym = config.CharHoe
	}

	return item
}

// featureFromSave converts a saved feature back to an entity
func featureFromSave(fs save.FeatureSave) *entity.Feature {
	// Handle backward compatibility for Passable field
	// Old saves won't have this field, so we infer from feature type:
	// - Springs (DrinkSource) are impassable
	// - Leaf piles (Bed) are passable
	passable := fs.Passable
	if !fs.Passable && fs.Bed {
		// Old save format - beds should be passable
		passable = true
	}
	// Springs stay impassable (fs.Passable would be false, which is correct)

	feature := &entity.Feature{
		ID:          fs.ID,
		FType:       entity.FeatureType(fs.FeatureType),
		DrinkSource: fs.DrinkSource,
		Bed:         fs.Bed,
		Passable:    passable,
	}
	feature.X = fs.Position.X
	feature.Y = fs.Position.Y
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
