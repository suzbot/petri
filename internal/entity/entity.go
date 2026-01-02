package entity

// EntityType distinguishes between different entity categories
type EntityType int

const (
	TypeCharacter EntityType = iota
	TypeItem
	TypeFeature
)

// Entity is the base interface for all game objects
type Entity interface {
	Position() (int, int)
	SetPosition(x, y int)
	Symbol() rune
	Type() EntityType
}

// BaseEntity provides common fields for all entities
type BaseEntity struct {
	X, Y  int
	Sym   rune
	EType EntityType
}

// Position returns the entity's current coordinates
func (e *BaseEntity) Position() (int, int) {
	return e.X, e.Y
}

// SetPosition updates the entity's coordinates
func (e *BaseEntity) SetPosition(x, y int) {
	e.X, e.Y = x, y
}

// Symbol returns the display character for this entity
func (e *BaseEntity) Symbol() rune {
	return e.Sym
}

// Type returns the entity type
func (e *BaseEntity) Type() EntityType {
	return e.EType
}
