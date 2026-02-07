package system

import (
	"petri/internal/config"
	"petri/internal/entity"
)

// CreateVessel creates a vessel item from an input material and recipe.
// The vessel inherits the input's appearance (color, pattern, texture).
// Future inputs (bark, wood, clay) may have different/fewer attributes -
// empty values are valid and will result in vessels without those attributes.
func CreateVessel(input *entity.Item, recipe *entity.Recipe) *entity.Item {
	return &entity.Item{
		BaseEntity: entity.BaseEntity{
			X:     input.X,
			Y:     input.Y,
			Sym:   config.CharVessel,
			EType: entity.TypeItem,
		},
		ItemType: "vessel",
		Kind:     recipe.Output.Kind,
		Color:    input.Color,
		Pattern:  input.Pattern,
		Texture:  input.Texture,
		// Edible is nil - vessels are not edible
		Container: &entity.ContainerData{
			Capacity: 1,
			Contents: []entity.Stack{},
		},
	}
}
