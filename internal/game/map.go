package game

import (
	"petri/internal/entity"
)

// Pos represents a 2D position on the map
type Pos struct {
	X, Y int
}

// Map represents the game world as a sparse grid
type Map struct {
	Width, Height int
	entities      map[Pos]entity.Entity

	// Indexed lookups for performance with many entities
	characters     []*entity.Character
	characterByPos map[Pos]*entity.Character // O(1) position lookup, max 1 character per position
	items          []*entity.Item
	features       []*entity.Feature

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
		entities:       make(map[Pos]entity.Entity),
		characters:     make([]*entity.Character, 0),
		characterByPos: make(map[Pos]*entity.Character),
		items:          make([]*entity.Item, 0),
		features:       make([]*entity.Feature, 0),
	}
}

// AddCharacter adds a character to the map
// Returns false if position is already occupied by another character
func (m *Map) AddCharacter(c *entity.Character) bool {
	pos := Pos{c.X, c.Y}
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
func (m *Map) EntityAt(x, y int) entity.Entity {
	pos := Pos{x, y}
	if char := m.characterByPos[pos]; char != nil {
		return char
	}
	return m.entities[pos]
}

// CharacterAt returns the character at the given position, or nil (O(1) lookup)
func (m *Map) CharacterAt(x, y int) *entity.Character {
	return m.characterByPos[Pos{x, y}]
}

// MoveEntity moves a non-character entity from one position to another
func (m *Map) MoveEntity(fromX, fromY, toX, toY int) {
	pos := Pos{fromX, fromY}
	if e, ok := m.entities[pos]; ok {
		delete(m.entities, pos)
		e.SetPosition(toX, toY)
		m.entities[Pos{toX, toY}] = e
	}
}

// MoveCharacter moves a character to a new position, updating the position index
// Returns true if the move succeeded, false if blocked (position already occupied)
func (m *Map) MoveCharacter(char *entity.Character, toX, toY int) bool {
	oldX, oldY := char.Position()
	oldPos := Pos{oldX, oldY}
	newPos := Pos{toX, toY}

	// Refuse move if target is occupied by another character
	if existing := m.characterByPos[newPos]; existing != nil && existing != char {
		return false
	}

	// Remove from old position - but verify it's actually this character
	if m.characterByPos[oldPos] == char {
		delete(m.characterByPos, oldPos)
	}

	// Update character position
	char.SetPosition(toX, toY)

	// Add to new position
	m.characterByPos[newPos] = char
	return true
}

// IsValid returns true if the position is within map bounds
func (m *Map) IsValid(x, y int) bool {
	return x >= 0 && x < m.Width && y >= 0 && y < m.Height
}

// IsOccupied returns true if there's a character at the position
func (m *Map) IsOccupied(x, y int) bool {
	return m.characterByPos[Pos{x, y}] != nil
}

// IsEmpty returns true if no entity (character, item, or feature) is at the position
func (m *Map) IsEmpty(x, y int) bool {
	if m.characterByPos[Pos{x, y}] != nil {
		return false
	}
	if m.ItemAt(x, y) != nil {
		return false
	}
	if m.FeatureAt(x, y) != nil {
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
func (m *Map) ItemAt(x, y int) *entity.Item {
	for _, item := range m.items {
		ix, iy := item.Position()
		if ix == x && iy == y {
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
func (m *Map) FeatureAt(x, y int) *entity.Feature {
	for _, f := range m.features {
		fx, fy := f.Position()
		if fx == x && fy == y {
			return f
		}
	}
	return nil
}

// DrinkSourceAt returns a drink source feature at the given position, or nil
func (m *Map) DrinkSourceAt(x, y int) *entity.Feature {
	f := m.FeatureAt(x, y)
	if f != nil && f.IsDrinkSource() {
		return f
	}
	return nil
}

// BedAt returns a bed feature at the given position, or nil
func (m *Map) BedAt(x, y int) *entity.Feature {
	f := m.FeatureAt(x, y)
	if f != nil && f.IsBed() {
		return f
	}
	return nil
}

// FindNearestDrinkSource finds the nearest unoccupied drink source to the given position
// Excludes springs occupied by other characters (the requesting character at x,y is allowed)
func (m *Map) FindNearestDrinkSource(x, y int) *entity.Feature {
	var nearest *entity.Feature
	nearestDist := int(^uint(0) >> 1)
	requestingChar := m.characterByPos[Pos{x, y}]

	for _, f := range m.features {
		if !f.IsDrinkSource() {
			continue
		}
		fx, fy := f.Position()

		// Skip springs occupied by another character
		occupant := m.characterByPos[Pos{fx, fy}]
		if occupant != nil && occupant != requestingChar {
			continue
		}

		dist := abs(x-fx) + abs(y-fy)
		if dist < nearestDist {
			nearestDist = dist
			nearest = f
		}
	}
	return nearest
}

// FindNearestBed finds the nearest unoccupied bed to the given position
// Excludes beds occupied by other characters (the requesting character at x,y is allowed)
func (m *Map) FindNearestBed(x, y int) *entity.Feature {
	var nearest *entity.Feature
	nearestDist := int(^uint(0) >> 1)
	requestingChar := m.characterByPos[Pos{x, y}]

	for _, f := range m.features {
		if !f.IsBed() {
			continue
		}
		fx, fy := f.Position()

		// Skip beds occupied by another character
		occupant := m.characterByPos[Pos{fx, fy}]
		if occupant != nil && occupant != requestingChar {
			continue
		}

		dist := abs(x-fx) + abs(y-fy)
		if dist < nearestDist {
			nearestDist = dist
			nearest = f
		}
	}
	return nearest
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
