package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/game"
)

// Pickup handles a character picking up an item (foraging)
func Pickup(char *entity.Character, item *entity.Item, gameMap *game.Map, log *ActionLog) {
	itemName := item.Description()

	// Remove item from map
	gameMap.RemoveItem(item)

	// Clear spawn/death timers - carried items are static
	item.SpawnTimer = 0
	item.DeathTimer = 0

	// Add to inventory
	char.Carrying = item

	// Update activity
	char.CurrentActivity = "Idle"

	// Log pickup
	if log != nil {
		log.Add(char.ID, char.Name, "activity",
			fmt.Sprintf("Picked up %s", itemName))
	}

	// Try to discover know-how from foraging
	TryDiscoverKnowHow(char, entity.ActionPickup, item, log, GetDiscoveryChance(char))

	// Clear intent and set idle cooldown
	char.Intent = nil
	char.IdleCooldown = config.IdleCooldown
}
