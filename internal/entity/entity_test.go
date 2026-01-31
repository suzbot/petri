package entity

import (
	"testing"

	"petri/internal/types"
)

// TestBaseEntity_Pos verifies Pos returns correct position
func TestBaseEntity_Pos(t *testing.T) {
	t.Parallel()

	e := &BaseEntity{X: 15, Y: 20}
	pos := e.Pos()
	if pos.X != 15 || pos.Y != 20 {
		t.Errorf("Pos(): got (%d, %d), want (15, 20)", pos.X, pos.Y)
	}
}

// TestBaseEntity_SetPos verifies SetPos updates position
func TestBaseEntity_SetPos(t *testing.T) {
	t.Parallel()

	e := &BaseEntity{X: 0, Y: 0}
	e.SetPos(types.Position{X: 25, Y: 30})
	pos := e.Pos()
	if pos.X != 25 || pos.Y != 30 {
		t.Errorf("SetPos({25, 30}) then Pos(): got (%d, %d), want (25, 30)", pos.X, pos.Y)
	}
}

// TestBaseEntity_Symbol verifies Symbol returns configured symbol
func TestBaseEntity_Symbol(t *testing.T) {
	t.Parallel()

	e := &BaseEntity{Sym: '@'}
	got := e.Symbol()
	if got != '@' {
		t.Errorf("Symbol(): got %q, want '@'", got)
	}
}

// TestBaseEntity_Type verifies Type returns configured entity type
func TestBaseEntity_Type(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		eType    EntityType
		expected EntityType
	}{
		{"TypeCharacter", TypeCharacter, TypeCharacter},
		{"TypeItem", TypeItem, TypeItem},
		{"TypeFeature", TypeFeature, TypeFeature},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &BaseEntity{EType: tt.eType}
			got := e.Type()
			if got != tt.expected {
				t.Errorf("Type(): got %d, want %d", got, tt.expected)
			}
		})
	}
}
