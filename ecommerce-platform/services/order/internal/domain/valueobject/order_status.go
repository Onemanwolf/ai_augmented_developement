package valueobject

import "fmt"

// OrderStatus represents the status of an order.
type OrderStatus string

// Order status constants.
const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// validStatuses contains all valid order statuses.
var validStatuses = map[OrderStatus]bool{
	OrderStatusPending:   true,
	OrderStatusCreated:   true,
	OrderStatusPaid:      true,
	OrderStatusShipped:   true,
	OrderStatusDelivered: true,
	OrderStatusCompleted: true,
	OrderStatusCancelled: true,
}

// validTransitions defines allowed status transitions.
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending:   {OrderStatusCreated, OrderStatusCancelled},
	OrderStatusCreated:   {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPaid:      {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusShipped:   {OrderStatusDelivered, OrderStatusCancelled},
	OrderStatusDelivered: {OrderStatusCompleted},
	OrderStatusCompleted: {},
	OrderStatusCancelled: {},
}

// ParseOrderStatus creates an OrderStatus from a string.
func ParseOrderStatus(s string) (OrderStatus, error) {
	status := OrderStatus(s)
	if !validStatuses[status] {
		return "", fmt.Errorf("invalid order status: %s", s)
	}
	return status, nil
}

// String returns the string representation of the OrderStatus.
func (s OrderStatus) String() string {
	return string(s)
}

// IsValid returns true if the status is valid.
func (s OrderStatus) IsValid() bool {
	return validStatuses[s]
}

// CanTransitionTo checks if transitioning to the target status is allowed.
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// IsFinal returns true if the status is a terminal state.
func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusCompleted || s == OrderStatusCancelled
}

// IsActive returns true if the order is in an active (non-terminal) state.
func (s OrderStatus) IsActive() bool {
	return !s.IsFinal()
}
