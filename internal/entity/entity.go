package entity

import "petri/internal/types"

// EntityType distinguishes between different entity categories
type EntityType int

const (
	TypeCharacter EntityType = iota
	TypeItem
	TypeFeature
)

// Entity is the base interface for all game objects
type Entity interface {
	Pos() types.Position
	SetPos(pos types.Position)
	Symbol() rune
	Type() EntityType
}

// BaseEntity provides common fields for all entities
type BaseEntity struct {
	X, Y  int // Kept for direct field access during transition
	Sym   rune
	EType EntityType
}

// Pos returns the entity's current position
func (e *BaseEntity) Pos() types.Position {
	return types.Position{X: e.X, Y: e.Y}
}

// SetPos updates the entity's position
func (e *BaseEntity) SetPos(pos types.Position) {
	e.X, e.Y = pos.X, pos.Y
}

// Symbol returns the display character for this entity
func (e *BaseEntity) Symbol() rune {
	return e.Sym
}

// Type returns the entity type
func (e *BaseEntity) Type() EntityType {
	return e.EType
}
