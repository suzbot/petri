package entity

import "testing"

// TestBaseEntity_Position verifies Position returns correct coordinates
func TestBaseEntity_Position(t *testing.T) {
	t.Parallel()

	e := &BaseEntity{X: 15, Y: 20}
	x, y := e.Position()
	if x != 15 || y != 20 {
		t.Errorf("Position(): got (%d, %d), want (15, 20)", x, y)
	}
}

// TestBaseEntity_SetPosition verifies SetPosition updates coordinates
func TestBaseEntity_SetPosition(t *testing.T) {
	t.Parallel()

	e := &BaseEntity{X: 0, Y: 0}
	e.SetPosition(25, 30)
	x, y := e.Position()
	if x != 25 || y != 30 {
		t.Errorf("SetPosition(25, 30) then Position(): got (%d, %d), want (25, 30)", x, y)
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
