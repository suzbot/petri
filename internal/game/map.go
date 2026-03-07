package game

import (
	"petri/internal/config"
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
	constructs     []*entity.Construct

	// Water terrain (springs and ponds)
	water map[types.Position]WaterType

	// Clay terrain positions (passable, items can exist on them)
	clay map[types.Position]bool

	// Tilled soil positions (walkable, items can exist on them)
	tilled map[types.Position]bool

	// Marked-for-tilling pool (user's tilling plan, independent of orders)
	markedForTilling map[types.Position]bool

	// Marked-for-construction pool (fence/hut placement plan, independent of orders)
	markedForConstruction map[types.Position]ConstructionMark

	// Manually watered tiles with decay timers (seconds remaining)
	wateredTimers map[types.Position]float64

	// ID counters for save/load
	nextItemID             int
	nextFeatureID          int
	nextConstructID        int
	nextConstructionLineID int

	// Variety registry for this world (determines poison/healing for item types)
	varieties *VarietyRegistry
}

// ConstructionMark records that a tile has been designated for fence construction.
type ConstructionMark struct {
	LineID   int    // All marks from one line-drawing operation share a LineID
	Material string // Empty until first character starts building this line
}

// NewMap creates a new map with the given dimensions
func NewMap(width, height int) *Map {
	return &Map{
		Width:                 width,
		Height:                height,
		entities:              make(map[types.Position]entity.Entity),
		characters:            make([]*entity.Character, 0),
		characterByPos:        make(map[types.Position]*entity.Character),
		items:                 make([]*entity.Item, 0),
		features:              make([]*entity.Feature, 0),
		constructs:            make([]*entity.Construct, 0),
		water:                 make(map[types.Position]WaterType),
		clay:                  make(map[types.Position]bool),
		tilled:                make(map[types.Position]bool),
		markedForTilling:      make(map[types.Position]bool),
		markedForConstruction: make(map[types.Position]ConstructionMark),
		wateredTimers:         make(map[types.Position]float64),
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

	// Refuse move if target has an impassable construct
	if c := m.ConstructAt(to); c != nil && !c.IsPassable() {
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
	if c := m.ConstructAt(pos); c != nil && !c.IsPassable() {
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
	if m.ConstructAt(pos) != nil {
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

// ItemsAt returns all items at the given position.
func (m *Map) ItemsAt(pos types.Position) []*entity.Item {
	var result []*entity.Item
	for _, item := range m.items {
		if item.Pos() == pos {
			result = append(result, item)
		}
	}
	return result
}

// HasItemOnMap returns true if the given item pointer is in the map's item list.
func (m *Map) HasItemOnMap(item *entity.Item) bool {
	for _, i := range m.items {
		if i == item {
			return true
		}
	}
	return false
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

// AddConstruct adds a construct to the map, assigning a unique ID
func (m *Map) AddConstruct(c *entity.Construct) {
	m.nextConstructID++
	c.ID = m.nextConstructID
	m.constructs = append(m.constructs, c)
}

// AddConstructDirect adds a construct to the map without assigning an ID (for save/load)
func (m *Map) AddConstructDirect(c *entity.Construct) {
	m.constructs = append(m.constructs, c)
}

// Constructs returns all constructs on the map
func (m *Map) Constructs() []*entity.Construct {
	return m.constructs
}

// ConstructAt returns the construct at the given position, or nil
func (m *Map) ConstructAt(pos types.Position) *entity.Construct {
	for _, c := range m.constructs {
		if c.Pos() == pos {
			return c
		}
	}
	return nil
}

// RemoveConstruct removes a construct from the map
func (m *Map) RemoveConstruct(c *entity.Construct) {
	for i, con := range m.constructs {
		if con == c {
			m.constructs = append(m.constructs[:i], m.constructs[i+1:]...)
			break
		}
	}
}

// NextConstructID returns the current next construct ID (for save/load)
func (m *Map) NextConstructID() int {
	return m.nextConstructID
}

// SetNextConstructID sets the next construct ID (for save/load)
func (m *Map) SetNextConstructID(id int) {
	m.nextConstructID = id
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

// IsWet returns true if the position is wet from any source:
// water-adjacent (8-directional) or manually watered by a character.
// Water tiles themselves are not "wet" — they ARE water (impassable).
func (m *Map) IsWet(pos types.Position) bool {
	if m.IsWater(pos) {
		return false
	}
	if m.wateredTimers[pos] > 0 {
		return true
	}
	dirs := [][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}
	for _, d := range dirs {
		neighbor := types.Position{X: pos.X + d[0], Y: pos.Y + d[1]}
		if m.IsWater(neighbor) {
			return true
		}
	}
	return false
}

// SetClay marks a position as clay terrain
func (m *Map) SetClay(pos types.Position) {
	m.clay[pos] = true
}

// IsClay returns true if the position has clay terrain
func (m *Map) IsClay(pos types.Position) bool {
	return m.clay[pos]
}

// HasClay returns true if any clay tiles exist on the map
func (m *Map) HasClay() bool {
	return len(m.clay) > 0
}

// ClayPositions returns all positions that have clay terrain
func (m *Map) ClayPositions() []types.Position {
	positions := make([]types.Position, 0, len(m.clay))
	for pos := range m.clay {
		positions = append(positions, pos)
	}
	return positions
}

// FindNearestClay finds the nearest clay tile to the given position.
// Clay is passable so no adjacency check is needed (unlike FindNearestWater).
// Returns the clay position and true if found, or zero position and false if not.
func (m *Map) FindNearestClay(pos types.Position) (types.Position, bool) {
	var nearestPos types.Position
	nearestDist := int(^uint(0) >> 1)
	found := false

	for clayPos := range m.clay {
		dist := pos.DistanceTo(clayPos)
		if dist < nearestDist {
			nearestDist = dist
			nearestPos = clayPos
			found = true
		}
	}
	return nearestPos, found
}

// SetTilled marks a position as tilled soil
func (m *Map) SetTilled(pos types.Position) {
	m.tilled[pos] = true
}

// IsTilled returns true if the position has been tilled
func (m *Map) IsTilled(pos types.Position) bool {
	return m.tilled[pos]
}

// TilledPositions returns all positions that have been tilled
func (m *Map) TilledPositions() []types.Position {
	positions := make([]types.Position, 0, len(m.tilled))
	for pos := range m.tilled {
		positions = append(positions, pos)
	}
	return positions
}

// MarkForTilling adds a position to the marked-for-tilling pool.
// Returns false if the position is already tilled (no-op).
func (m *Map) MarkForTilling(pos types.Position) bool {
	if m.tilled[pos] {
		return false
	}
	m.markedForTilling[pos] = true
	return true
}

// UnmarkForTilling removes a position from the marked-for-tilling pool.
func (m *Map) UnmarkForTilling(pos types.Position) {
	delete(m.markedForTilling, pos)
}

// IsMarkedForTilling returns true if the position is in the marked-for-tilling pool.
func (m *Map) IsMarkedForTilling(pos types.Position) bool {
	return m.markedForTilling[pos]
}

// MarkedForTillingPositions returns all positions in the marked-for-tilling pool.
func (m *Map) MarkedForTillingPositions() []types.Position {
	positions := make([]types.Position, 0, len(m.markedForTilling))
	for pos := range m.markedForTilling {
		positions = append(positions, pos)
	}
	return positions
}

// MarkForConstruction adds a position to the marked-for-construction pool with the given lineID.
// Returns false if the position is already marked or has an existing construct.
func (m *Map) MarkForConstruction(pos types.Position, lineID int) bool {
	if _, exists := m.markedForConstruction[pos]; exists {
		return false
	}
	if m.ConstructAt(pos) != nil {
		return false
	}
	m.markedForConstruction[pos] = ConstructionMark{LineID: lineID}
	return true
}

// UnmarkForConstruction removes a position from the marked-for-construction pool.
func (m *Map) UnmarkForConstruction(pos types.Position) {
	delete(m.markedForConstruction, pos)
}

// IsMarkedForConstruction returns true if the position is in the marked-for-construction pool.
func (m *Map) IsMarkedForConstruction(pos types.Position) bool {
	_, exists := m.markedForConstruction[pos]
	return exists
}

// GetConstructionMark returns the construction mark at pos, if any.
func (m *Map) GetConstructionMark(pos types.Position) (ConstructionMark, bool) {
	mark, exists := m.markedForConstruction[pos]
	return mark, exists
}

// MarkedForConstructionPositions returns all positions in the marked-for-construction pool.
func (m *Map) MarkedForConstructionPositions() []types.Position {
	positions := make([]types.Position, 0, len(m.markedForConstruction))
	for pos := range m.markedForConstruction {
		positions = append(positions, pos)
	}
	return positions
}

// SetLineMaterialAt stamps the given material onto the construction mark at pos (used by save/load).
func (m *Map) SetLineMaterialAt(pos types.Position, material string) {
	if mark, exists := m.markedForConstruction[pos]; exists {
		mark.Material = material
		m.markedForConstruction[pos] = mark
	}
}

// SetLineMaterial stamps the given material onto all construction marks with the matching lineID.
func (m *Map) SetLineMaterial(lineID int, material string) {
	for pos, mark := range m.markedForConstruction {
		if mark.LineID == lineID {
			mark.Material = material
			m.markedForConstruction[pos] = mark
		}
	}
}

// HasUnbuiltConstructionPositions returns true if any marked-for-construction position has no construct yet.
func (m *Map) HasUnbuiltConstructionPositions() bool {
	for pos := range m.markedForConstruction {
		if m.ConstructAt(pos) == nil {
			return true
		}
	}
	return false
}

// NextConstructionLineID increments and returns the next construction line ID.
func (m *Map) NextConstructionLineID() int {
	m.nextConstructionLineID++
	return m.nextConstructionLineID
}

// SetConstructionLineID sets the line ID counter (used by save/load).
func (m *Map) SetConstructionLineID(id int) {
	m.nextConstructionLineID = id
}

// ConstructionLineID returns the current line ID counter (used by save/load).
func (m *Map) ConstructionLineID() int {
	return m.nextConstructionLineID
}

// SetManuallyWatered marks a position as manually watered with the full duration timer.
func (m *Map) SetManuallyWatered(pos types.Position) {
	m.wateredTimers[pos] = config.WateredTileDuration
}

// IsManuallyWatered returns true if the position has been manually watered and the timer hasn't expired.
func (m *Map) IsManuallyWatered(pos types.Position) bool {
	return m.wateredTimers[pos] > 0
}

// WateredPositions returns all positions that are currently manually watered.
func (m *Map) WateredPositions() []types.Position {
	positions := make([]types.Position, 0, len(m.wateredTimers))
	for pos := range m.wateredTimers {
		positions = append(positions, pos)
	}
	return positions
}

// SetWateredTimer sets a specific timer value for a position (used by save/load).
func (m *Map) SetWateredTimer(pos types.Position, remaining float64) {
	if remaining > 0 {
		m.wateredTimers[pos] = remaining
	}
}

// WateredTimer returns the remaining watered timer for a position (0 if not watered).
func (m *Map) WateredTimer(pos types.Position) float64 {
	return m.wateredTimers[pos]
}

// UpdateWateredTimers decrements all watered tile timers and removes expired ones.
func (m *Map) UpdateWateredTimers(delta float64) {
	for pos, remaining := range m.wateredTimers {
		remaining -= delta
		if remaining <= 0 {
			delete(m.wateredTimers, pos)
		} else {
			m.wateredTimers[pos] = remaining
		}
	}
}
