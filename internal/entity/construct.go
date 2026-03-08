package entity

import (
	"petri/internal/config"
	"petri/internal/types"
)

// Construct represents a character-built structure (fence, hut, furniture)
type Construct struct {
	BaseEntity
	ID            int
	ConstructType string      // "structure", future: "furniture"
	Kind          string      // "fence", "hut"
	Material      string      // ItemType of material: "grass", "stick", "brick"
	MaterialColor types.Color // rendering color
	Passable      bool
	Movable       bool   // false for structures, true for future furniture
	WallRole      string // position-aware role for hut constructs: "corner-tl", "corner-tr", "corner-bl", "corner-br", "edge-h", "edge-v", "door"
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

// NewHutConstruct creates a hut construct (wall or door) at the given position.
// WallRole determines the symbol and passability: "door" is passable, all others are walls.
func NewHutConstruct(x, y int, material string, materialColor types.Color, wallRole string) *Construct {
	sym := wallRoleToSymbol(wallRole)
	return &Construct{
		BaseEntity: BaseEntity{
			X:     x,
			Y:     y,
			Sym:   sym,
			EType: TypeConstruct,
		},
		ConstructType: "structure",
		Kind:          "hut",
		Material:      material,
		MaterialColor: materialColor,
		Passable:      wallRole == "door",
		Movable:       false,
		WallRole:      wallRole,
	}
}

// wallRoleToSymbol maps a WallRole string to its display symbol.
func wallRoleToSymbol(wallRole string) rune {
	switch wallRole {
	case "corner-tl":
		return config.CharHutCornerTL
	case "corner-tr":
		return config.CharHutCornerTR
	case "corner-bl":
		return config.CharHutCornerBL
	case "corner-br":
		return config.CharHutCornerBR
	case "edge-h":
		return config.CharHutEdgeH
	case "edge-v":
		return config.CharHutEdgeV
	case "door":
		return config.CharHutDoor
	case "t-down":
		return config.CharHutTDown
	case "t-up":
		return config.CharHutTUp
	case "t-right":
		return config.CharHutTRight
	case "t-left":
		return config.CharHutTLeft
	case "cross":
		return config.CharHutCross
	default:
		return config.CharHutEdgeH
	}
}

// IsPassable returns whether characters can walk onto this construct
func (c *Construct) IsPassable() bool {
	return c.Passable
}

// DisplayName returns a formatted name like "Stick Fence", "Stick Hut Wall", or "Stick Hut Door"
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
	if c.Kind == "hut" {
		if c.WallRole == "door" {
			return materialDisplay + " Hut Door"
		}
		return materialDisplay + " Hut Wall"
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
