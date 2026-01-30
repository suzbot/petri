package game

import (
	"petri/internal/entity"
	"petri/internal/types"
)

// VarietyRegistry stores all item varieties that exist in a world.
// This replaces the old PoisonConfig/HealingConfig approach by defining
// "what exists" upfront, with properties attached to varieties.
type VarietyRegistry struct {
	varieties map[string]*entity.ItemVariety
}

// NewVarietyRegistry creates an empty variety registry
func NewVarietyRegistry() *VarietyRegistry {
	return &VarietyRegistry{
		varieties: make(map[string]*entity.ItemVariety),
	}
}

// Register adds a variety to the registry
func (r *VarietyRegistry) Register(v *entity.ItemVariety) {
	r.varieties[v.ID] = v
}

// Get retrieves a variety by ID, returns nil if not found
func (r *VarietyRegistry) Get(id string) *entity.ItemVariety {
	return r.varieties[id]
}

// VarietiesOfType returns all varieties of a given item type
func (r *VarietyRegistry) VarietiesOfType(itemType string) []*entity.ItemVariety {
	var result []*entity.ItemVariety
	for _, v := range r.varieties {
		if v.ItemType == itemType {
			result = append(result, v)
		}
	}
	return result
}

// AllVarieties returns all registered varieties
func (r *VarietyRegistry) AllVarieties() []*entity.ItemVariety {
	result := make([]*entity.ItemVariety, 0, len(r.varieties))
	for _, v := range r.varieties {
		result = append(result, v)
	}
	return result
}

// EdibleVarieties returns all varieties that are edible
func (r *VarietyRegistry) EdibleVarieties() []*entity.ItemVariety {
	var result []*entity.ItemVariety
	for _, v := range r.varieties {
		if v.IsEdible() {
			result = append(result, v)
		}
	}
	return result
}

// Count returns the total number of registered varieties
func (r *VarietyRegistry) Count() int {
	return len(r.varieties)
}

// GetByAttributes looks up a variety by item attributes.
// Returns nil if no matching variety is registered.
func (r *VarietyRegistry) GetByAttributes(itemType string, color types.Color, pattern types.Pattern, texture types.Texture) *entity.ItemVariety {
	id := entity.GenerateVarietyID(itemType, color, pattern, texture)
	return r.varieties[id]
}
