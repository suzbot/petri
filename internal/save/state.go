package save

import (
	"time"

	"petri/internal/types"
)

// CurrentVersion is the save file format version
const CurrentVersion = 1

// SaveState represents the complete saveable state of a world
type SaveState struct {
	Version         int                  `json:"version"`
	SavedAt         time.Time            `json:"saved_at"`
	ElapsedGameTime float64              `json:"elapsed_game_time"` // Total simulation time in seconds

	MapWidth  int `json:"map_width"`
	MapHeight int `json:"map_height"`

	Varieties   []VarietySave            `json:"varieties"`   // Full variety registry
	Characters  []CharacterSave          `json:"characters"`
	Items       []ItemSave               `json:"items"`
	Features    []FeatureSave            `json:"features"`
	WaterTiles  []WaterTileSave          `json:"water_tiles,omitempty"`
	ActionLogs  map[int][]EventSave      `json:"action_logs"` // Per-character event logs, keyed by char ID
	Orders      []OrderSave              `json:"orders,omitempty"`
	NextOrderID int                      `json:"next_order_id,omitempty"`

	// Ground spawning timers
	GroundSpawnStick float64 `json:"ground_spawn_stick,omitempty"`
	GroundSpawnNut   float64 `json:"ground_spawn_nut,omitempty"`
	GroundSpawnShell float64 `json:"ground_spawn_shell,omitempty"`
}

// OrderSave represents an order for serialization
type OrderSave struct {
	ID         int    `json:"id"`
	ActivityID string `json:"activity_id"`
	TargetType string `json:"target_type"`
	Status     string `json:"status"`
	AssignedTo int    `json:"assigned_to"`
}

// WorldMeta contains display info for world selection (separate from full state)
type WorldMeta struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	LastPlayedAt   time.Time `json:"last_played_at"`
	CharacterCount int       `json:"character_count"`
	AliveCount     int       `json:"alive_count"`
}

// EventSave represents a logged event for serialization
type EventSave struct {
	GameTime float64 `json:"game_time"` // Elapsed game time when event occurred
	CharID   int     `json:"char_id"`
	CharName string  `json:"char_name"`
	Type     string  `json:"type"`
	Message  string  `json:"message"`
}

// CharacterSave represents a character for serialization
type CharacterSave struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	types.Position

	// Stats
	Health float64 `json:"health"`
	Hunger float64 `json:"hunger"`
	Thirst float64 `json:"thirst"`
	Energy float64 `json:"energy"`
	Mood   float64 `json:"mood"`

	// Status
	Poisoned    bool    `json:"poisoned"`
	PoisonTimer float64 `json:"poison_timer"`
	IsDead      bool    `json:"is_dead"`
	IsSleeping  bool    `json:"is_sleeping"`
	AtBed       bool    `json:"at_bed"`

	// Frustration
	IsFrustrated      bool    `json:"is_frustrated"`
	FrustrationTimer  float64 `json:"frustration_timer"`
	FailedIntentCount int     `json:"failed_intent_count"`

	// Idle state
	IdleCooldown  float64 `json:"idle_cooldown"`
	LastLookedX   int     `json:"last_looked_x"`
	LastLookedY   int     `json:"last_looked_y"`
	HasLastLooked bool    `json:"has_last_looked"`

	// Talking (partner stored as ID, -1 if none)
	TalkingWithID int     `json:"talking_with_id"`
	TalkTimer     float64 `json:"talk_timer"`

	// Cooldowns
	HungerCooldown float64 `json:"hunger_cooldown"`
	ThirstCooldown float64 `json:"thirst_cooldown"`
	EnergyCooldown float64 `json:"energy_cooldown"`

	// Action state
	ActionProgress   float64 `json:"action_progress"`
	SpeedAccumulator float64 `json:"speed_accumulator"`
	CurrentActivity  string  `json:"current_activity"`

	// Mind
	Preferences     []PreferenceSave `json:"preferences"`
	Knowledge       []KnowledgeSave  `json:"knowledge"`
	KnownActivities []string         `json:"known_activities,omitempty"`
	KnownRecipes    []string         `json:"known_recipes,omitempty"`

	// Inventory (2 slots)
	Inventory []ItemSave `json:"inventory,omitempty"`

	// Orders
	AssignedOrderID int `json:"assigned_order_id,omitempty"` // ID of assigned order (0 = none)
}

// PlantPropertiesSave represents plant properties for serialization
type PlantPropertiesSave struct {
	IsGrowing  bool    `json:"is_growing"`
	SpawnTimer float64 `json:"spawn_timer"`
}

// StackSave represents a stack in a container for serialization
type StackSave struct {
	// Variety attributes (serialized by value, not pointer)
	ItemType string `json:"item_type"`
	Color    string `json:"color"`
	Pattern  string `json:"pattern"`
	Texture  string `json:"texture"`
	Count    int    `json:"count"`
}

// ContainerDataSave represents container properties for serialization
type ContainerDataSave struct {
	Capacity int         `json:"capacity"`
	Contents []StackSave `json:"contents,omitempty"`
}

// ItemSave represents an item for serialization
type ItemSave struct {
	ID int `json:"id"`
	types.Position
	Name     string `json:"name,omitempty"` // Display name for crafted items
	ItemType string `json:"item_type"`
	Color    string `json:"color"`
	Pattern  string `json:"pattern"`
	Texture  string `json:"texture"`

	// Plant properties (nil for non-plants)
	Plant *PlantPropertiesSave `json:"plant,omitempty"`

	// Container properties (nil for non-containers)
	Container *ContainerDataSave `json:"container,omitempty"`

	Edible    bool `json:"edible"`
	Poisonous bool `json:"poisonous"`
	Healing   bool `json:"healing"`

	DeathTimer float64 `json:"death_timer"`
}

// WaterTileSave represents a water tile for serialization
type WaterTileSave struct {
	types.Position
	WaterType int `json:"water_type"` // WaterType enum value (1=spring, 2=pond)
}

// FeatureSave represents a feature for serialization
type FeatureSave struct {
	ID int `json:"id"`
	types.Position
	FeatureType int  `json:"feature_type"` // FeatureType enum value
	DrinkSource bool `json:"drink_source"`
	Bed         bool `json:"bed"`
	Passable    bool `json:"passable"`
}

// PreferenceSave represents a preference for serialization
type PreferenceSave struct {
	ItemType string `json:"item_type"`
	Color    string `json:"color"`
	Pattern  string `json:"pattern"`
	Texture  string `json:"texture"`
	Valence  int    `json:"valence"`
}

// KnowledgeSave represents knowledge for serialization
type KnowledgeSave struct {
	Category string `json:"category"`
	ItemType string `json:"item_type"`
	Color    string `json:"color"`
	Pattern  string `json:"pattern"`
	Texture  string `json:"texture"`
}

// VarietySave represents an item variety for serialization
type VarietySave struct {
	ItemType  string `json:"item_type"`
	Color     string `json:"color"`
	Pattern   string `json:"pattern"`
	Texture   string `json:"texture"`
	Edible    bool   `json:"edible"`
	Poisonous bool   `json:"poisonous"`
	Healing   bool   `json:"healing"`
}
