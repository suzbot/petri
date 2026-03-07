package entity

import (
	"petri/internal/config"
	"petri/internal/types"
)

// Construct represents a character-built structure (fence, future: hut, door, furniture)
type Construct struct {
	BaseEntity
	ID            int
	ConstructType string      // "structure", future: "furniture"
	Kind          string      // "fence", future: "hut", "door"
	Material      string      // ItemType of material: "grass", "stick", "brick"
	MaterialColor types.Color // rendering color
	Passable      bool
	Movable       bool // false for structures, true for future furniture
}

// NewFence creates a new fence construct at the given position with the specified material
func NewFence(x, y int, material string, materialColor types.Color) *Construct {
	return &Construct{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   config.CharFence,
			EType: TypeConstruct,
		},
		ConstructType: "structure",
		Kind:          "fence",
		Material:      material,
		MaterialColor: materialColor,
		Passable:      false,
		Movable:       false,
	}
}

// IsPassable returns whether characters can walk onto this construct
func (c *Construct) IsPassable() bool {
	return c.Passable
}

// DisplayName returns a formatted name like "Stick Fence"
func (c *Construct) DisplayName() string {
	materialDisplay := c.Material
	switch c.Material {
	case "grass":
		materialDisplay = "Thatch"
	case "stick":
		materialDisplay = "Stick"
	case "brick":
		materialDisplay = "Brick"
	}
	kind := c.Kind
	if len(kind) > 0 {
		kind = string(kind[0]-32) + kind[1:] // capitalize first letter
	}
	return materialDisplay + " " + kind
}

// Description returns the construct type capitalized
func (c *Construct) Description() string {
	if c.ConstructType == "structure" {
		return "Structure"
	}
	return c.ConstructType
}

// PreferenceKind returns the lowercase composed identity for preference matching.
// Maps material to its display name and combines with Kind: "stick fence", "thatch fence", "brick fence".
func (c *Construct) PreferenceKind() string {
	materialDisplay := c.Material
	switch c.Material {
	case "grass":
		materialDisplay = "thatch"
	case "stick":
		materialDisplay = "stick"
	case "brick":
		materialDisplay = "brick"
	}
	return materialDisplay + " " + c.Kind
}
