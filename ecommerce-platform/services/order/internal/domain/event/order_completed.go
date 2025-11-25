package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderCompleted is emitted when an order is completed (delivered and confirmed).
type OrderCompleted struct {
	BaseEvent
	CompletedAt time.Time `json:"completed_at"`
}

// NewOrderCompleted creates a new OrderCompleted event.
func NewOrderCompleted(orderID valueobject.OrderID) *OrderCompleted {
	now := time.Now().UTC()
	return &OrderCompleted{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "OrderCompleted",
			AggregateIDV: orderID.String(),
			OccurredAtV:  now,
		},
		CompletedAt: now,
	}
}
