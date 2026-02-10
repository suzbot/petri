package entity

import "strings"

// OrderStatus represents the current state of an order
type OrderStatus string

const (
	OrderOpen     OrderStatus = "open"     // Available to be taken
	OrderAssigned OrderStatus = "assigned" // Currently being worked on
	OrderPaused   OrderStatus = "paused"   // Interrupted by character needs
)

// Order represents a player-issued work order
type Order struct {
	ID         int         // Unique identifier
	ActivityID string      // Activity to perform (e.g., "harvest")
	TargetType string      // Item type to target (e.g., "berry", "gourd")
	Status     OrderStatus // Current status
	AssignedTo int         // Character ID working on this order, 0 if unassigned
}

// NewOrder creates a new order with the given activity and target
func NewOrder(id int, activityID, targetType string) *Order {
	return &Order{
		ID:         id,
		ActivityID: activityID,
		TargetType: targetType,
		Status:     OrderOpen,
		AssignedTo: 0,
	}
}

// DisplayName returns a human-readable description of the order
func (o *Order) DisplayName() string {
	activity, ok := ActivityRegistry[o.ActivityID]
	if !ok {
		return o.ActivityID + " " + o.TargetType
	}
	switch activity.Category {
	case "craft":
		return "Craft " + strings.ToLower(activity.Name)
	case "garden":
		return activity.Name
	default:
		return activity.Name + " " + Pluralize(o.TargetType)
	}
}

// StatusDisplay returns a human-readable status string
func (o *Order) StatusDisplay() string {
	switch o.Status {
	case OrderOpen:
		return "Open"
	case OrderAssigned:
		return "Assigned"
	case OrderPaused:
		return "Paused"
	default:
		return string(o.Status)
	}
}
