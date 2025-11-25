// Package command contains command definitions for the Order service.
package command

import "github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"

// UpdateOrderStatus represents a command to update an order's status.
type UpdateOrderStatus struct {
	OrderID   string                  `json:"order_id"`
	NewStatus valueobject.OrderStatus `json:"new_status"`
}

// Validate validates the command.
func (c *UpdateOrderStatus) Validate() error {
	if c.OrderID == "" {
		return ErrOrderIDRequired
	}
	if !c.NewStatus.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}
