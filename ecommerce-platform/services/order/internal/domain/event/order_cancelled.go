package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderCancelled is emitted when an order is cancelled.
type OrderCancelled struct {
	BaseEvent
	Reason string `json:"reason"`
}

// NewOrderCancelled creates a new OrderCancelled event.
func NewOrderCancelled(orderID valueobject.OrderID, reason string) *OrderCancelled {
	return &OrderCancelled{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "OrderCancelled",
			AggregateIDV: orderID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		Reason: reason,
	}
}
