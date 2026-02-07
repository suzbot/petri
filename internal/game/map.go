package game

import (
	"petri/internal/entity"
	"petri/internal/types"
)

// WaterType distinguishes between different water sources
type WaterType int

const (
	WaterNone   WaterType = iota
	WaterSpring           // Natural spring (renders as ☉)
	WaterPond             // Pond tile (renders as ≈)
)

// Map represents the game world as a sparse grid
type Map struct {
	Width, Height int
	entities      map[types.Position]entity.Entity

	// Indexed lookups for performance with many entities
	characters     []*entity.Character
	characterByPos map[types.Position]*entity.Character // O(1) position lookup, max 1 character per position
	items          []*entity.Item
	features       []*entity.Feature

	// Water terrain (springs and ponds)
	water map[types.Position]WaterType

	// ID counters for save/load
	nextItemID    int
	nextFeatureID int

	// Variety registry for this world (determines poison/healing for item types)
	varieties *VarietyRegistry
}

// NewMap creates a new map with the given dimensions
func NewMap(width, height int) *Map {
	return &Map{
		Width:          width,
		Height:         height,
		entities:       make(map[types.Position]entity.Entity),
		characters:     make([]*entity.Character, 0),
		characterByPos: make(map[types.Position]*entity.Character),
		items:          make([]*entity.Item, 0),
		features:       make([]*entity.Feature, 0),
		water:          make(map[types.Position]WaterType),
	}
}

// AddCharacter adds a character to the map
// Returns false if position is already occupied by another character
func (m *Map) AddCharacter(c *entity.Character) bool {
	pos := c.Pos()
	if m.characterByPos[pos] != nil {
		return false
	}
	m.characters = append(m.characters, c)
	m.characterByPos[pos] = c
	return true
}

// AddItem adds an item to the map, assigning a unique ID
func (m *Map) AddItem(item *entity.Item) {
	// Assign unique ID
	m.nextItemID++
	item.ID = m.nextItemID

	// Items are stored only in the items slice, not in entities map
	// This allows characters to walk over items without overwriting them
	m.items = append(m.items, item)
}

// AddItemDirect adds an item to the map without assigning an ID (for save/load)
func (m *Map) AddItemDirect(item *entity.Item) {
	m.items = append(m.items, item)
}

// RemoveItem removes an item from the map
func (m *Map) RemoveItem(item *entity.Item) {
	for i, it := range m.items {
		if it == item {
			m.items = append(m.items[:i], m.items[i+1:]...)
			break
		}
	}
}

// EntityAt returns an entity at the given position, or nil
// For characters, returns the character at that position
func (m *Map) EntityAt(pos types.Position) entity.Entity {
	if char := m.characterByPos[pos]; char != nil {
		return char
	}
	return m.entities[pos]
}

// CharacterAt returns the character at the given position, or nil (O(1) lookup)
func (m *Map) CharacterAt(pos types.Position) *entity.Character {
	return m.characterByPos[pos]
}

// MoveEntity moves a non-character entity from one position to another
func (m *Map) MoveEntity(from, to types.Position) {
	if e, ok := m.entities[from]; ok {
		delete(m.entities, from)
		e.SetPos(to)
		m.entities[to] = e
	}
}

// MoveCharacter moves a character to a new position, updating the position index
// Returns true if the move succeeded, false if blocked (position already occupied or impassable feature)
func (m *Map) MoveCharacter(char *entity.Character, to types.Position) bool {
	oldPos := char.Pos()

	// Refuse move if target is occupied by another character
	if existing := m.characterByPos[to]; existing != nil && existing != char {
		return false
	}

	// Refuse move if target is water
	if m.IsWater(to) {
		return false
	}

	// Refuse move if target has an impassable feature
	if f := m.FeatureAt(to); f != nil && !f.IsPassable() {
		return false
	}

	// Remove from old position - but verify it's actually this character
	if m.characterByPos[oldPos] == char {
		delete(m.characterByPos, oldPos)
	}

	// Update character position
	char.SetPos(to)

	// Add to new position
	m.characterByPos[to] = char
	return true
}

// IsValid returns true if the position is within map bounds
func (m *Map) IsValid(pos types.Position) bool {
	return pos.X >= 0 && pos.X < m.Width && pos.Y >= 0 && pos.Y < m.Height
}

// IsOccupied returns true if there's a character at the position
func (m *Map) IsOccupied(pos types.Position) bool {
	return m.characterByPos[pos] != nil
}

// IsBlocked returns true if the position is blocked by a character, impassable feature, or water
func (m *Map) IsBlocked(pos types.Position) bool {
	if m.characterByPos[pos] != nil {
		return true
	}
	if m.IsWater(pos) {
		return true
	}
	if f := m.FeatureAt(pos); f != nil && !f.IsPassable() {
		return true
	}
	return false
}

// IsEmpty returns true if no entity (character, item, feature, or water) is at the position
func (m *Map) IsEmpty(pos types.Position) bool {
	if m.characterByPos[pos] != nil {
		return false
	}
	if m.IsWater(pos) {
		return false
	}
	if m.ItemAt(pos) != nil {
		return false
	}
	if m.FeatureAt(pos) != nil {
		return false
	}
	return true
}

// Characters returns all characters on the map
func (m *Map) Characters() []*entity.Character {
	return m.characters
}

// Items returns all items on the map
func (m *Map) Items() []*entity.Item {
	return m.items
}

// ItemAt returns the item at the given position, or nil
// Searches the items slice directly
func (m *Map) ItemAt(pos types.Position) *entity.Item {
	for _, item := range m.items {
		if item.Pos() == pos {
			return item
		}
	}
	return nil
}

// AddFeature adds a feature to the map, assigning a unique ID
func (m *Map) AddFeature(f *entity.Feature) {
	// Assign unique ID
	m.nextFeatureID++
	f.ID = m.nextFeatureID

	// Features are stored only in the features slice, not in entities map
	// This allows characters to walk over/onto features
	m.features = append(m.features, f)
}

// AddFeatureDirect adds a feature to the map without assigning an ID (for save/load)
func (m *Map) AddFeatureDirect(f *entity.Feature) {
	m.features = append(m.features, f)
}

// Features returns all features on the map
func (m *Map) Features() []*entity.Feature {
	return m.features
}

// FeatureAt returns the feature at the given position, or nil
func (m *Map) FeatureAt(pos types.Position) *entity.Feature {
	for _, f := range m.features {
		if f.Pos() == pos {
			return f
		}
	}
	return nil
}

// BedAt returns a bed feature at the given position, or nil
func (m *Map) BedAt(pos types.Position) *entity.Feature {
	f := m.FeatureAt(pos)
	if f != nil && f.IsBed() {
		return f
	}
	return nil
}

// FindNearestBed finds the nearest unoccupied bed to the given position
// Excludes beds occupied by other characters (the requesting character at pos is allowed)
func (m *Map) FindNearestBed(pos types.Position) *entity.Feature {
	var nearest *entity.Feature
	nearestDist := int(^uint(0) >> 1)
	requestingChar := m.characterByPos[pos]

	for _, f := range m.features {
		if !f.IsBed() {
			continue
		}
		fpos := f.Pos()

		// Skip beds occupied by another character
		occupant := m.characterByPos[fpos]
		if occupant != nil && occupant != requestingChar {
			continue
		}

		dist := pos.DistanceTo(fpos)
		if dist < nearestDist {
			nearestDist = dist
			nearest = f
		}
	}
	return nearest
}

// Varieties returns the variety registry for this map
func (m *Map) Varieties() *VarietyRegistry {
	return m.varieties
}

// SetVarieties sets the variety registry for this map
func (m *Map) SetVarieties(v *VarietyRegistry) {
	m.varieties = v
}

// NextItemID returns the current next item ID (for save/load)
func (m *Map) NextItemID() int {
	return m.nextItemID
}

// SetNextItemID sets the next item ID (for save/load)
func (m *Map) SetNextItemID(id int) {
	m.nextItemID = id
}

// NextFeatureID returns the current next feature ID (for save/load)
func (m *Map) NextFeatureID() int {
	return m.nextFeatureID
}

// SetNextFeatureID sets the next feature ID (for save/load)
func (m *Map) SetNextFeatureID(id int) {
	m.nextFeatureID = id
}

// AddWater adds a water tile at the given position
func (m *Map) AddWater(pos types.Position, wtype WaterType) {
	m.water[pos] = wtype
}

// RemoveWater removes a water tile at the given position
func (m *Map) RemoveWater(pos types.Position) {
	delete(m.water, pos)
}

// IsWater returns true if there is a water tile at the position
func (m *Map) IsWater(pos types.Position) bool {
	return m.water[pos] != WaterNone
}

// WaterAt returns the water type at the given position (WaterNone if no water)
func (m *Map) WaterAt(pos types.Position) WaterType {
	return m.water[pos]
}

// WaterPositions returns all positions that have water tiles
func (m *Map) WaterPositions() []types.Position {
	positions := make([]types.Position, 0, len(m.water))
	for pos := range m.water {
		positions = append(positions, pos)
	}
	return positions
}

// FindNearestWater finds the nearest water tile that has an available cardinal-adjacent tile.
// Water tiles are impassable, so characters drink from cardinally adjacent tiles (N/E/S/W).
// A water tile is available if at least one cardinal-adjacent tile is unblocked or occupied by the requester.
// Returns the water position and true if found, or zero position and false if not.
func (m *Map) FindNearestWater(pos types.Position) (types.Position, bool) {
	var nearestPos types.Position
	nearestDist := int(^uint(0) >> 1)
	found := false
	requestingChar := m.characterByPos[pos]

	cardinalDirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	for waterPos := range m.water {
		// Check if any cardinal-adjacent tile is available
		hasAvailableTile := false
		for _, dir := range cardinalDirs {
			adjPos := types.Position{X: waterPos.X + dir[0], Y: waterPos.Y + dir[1]}
			if !m.IsValid(adjPos) {
				continue
			}
			occupant := m.characterByPos[adjPos]
			if occupant == nil || occupant == requestingChar {
				// Also check for impassable features or water at adjacent tile
				if adjFeature := m.FeatureAt(adjPos); adjFeature != nil && !adjFeature.IsPassable() {
					continue
				}
				if m.IsWater(adjPos) {
					continue
				}
				hasAvailableTile = true
				break
			}
		}

		if !hasAvailableTile {
			continue
		}

		dist := pos.DistanceTo(waterPos)
		if dist < nearestDist {
			nearestDist = dist
			nearestPos = waterPos
			found = true
		}
	}
	return nearestPos, found
}
