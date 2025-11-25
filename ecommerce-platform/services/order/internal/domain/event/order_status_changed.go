package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderStatusChanged is emitted when an order's status changes.
type OrderStatusChanged struct {
	BaseEvent
	OldStatus valueobject.OrderStatus `json:"old_status"`
	NewStatus valueobject.OrderStatus `json:"new_status"`
}

// NewOrderStatusChanged creates a new OrderStatusChanged event.
func NewOrderStatusChanged(
	orderID valueobject.OrderID,
	oldStatus valueobject.OrderStatus,
	newStatus valueobject.OrderStatus,
) *OrderStatusChanged {
	return &OrderStatusChanged{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "OrderStatusChanged",
			AggregateIDV: orderID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}
}
