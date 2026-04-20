package system

import (
	"petri/internal/entity"
)

// CreateVessel creates a vessel item from an input material and recipe.
// The vessel inherits the input's appearance (color, pattern, texture).
func CreateVessel(input *entity.Item, recipe *entity.Recipe) *entity.Item {
	vessel := entity.NewVessel(input.X, input.Y, recipe.Output.Kind, input.ItemType)
	vessel.Color = input.Color
	vessel.Pattern = input.Pattern
	vessel.Texture = input.Texture
	return vessel
}

// CreateHoe creates a hoe item from a stick and shell.
// The hoe inherits the shell's color (e.g., "silver shell hoe").
func CreateHoe(shell *entity.Item, recipe *entity.Recipe) *entity.Item {
	return entity.NewHoe(shell.X, shell.Y, shell.Color)
}

// CreateBrick creates a brick item from a consumed clay input.
// Bricks are uniform — position is taken from the clay item.
func CreateBrick(clay *entity.Item, recipe *entity.Recipe) *entity.Item {
	return entity.NewBrick(clay.X, clay.Y)
}
