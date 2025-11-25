package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// OrderDelivered is emitted when an order is delivered.
type OrderDelivered struct {
	BaseEvent
	OrderID     string    `json:"order_id"`
	DeliveredAt time.Time `json:"delivered_at"`
}

// NewOrderDelivered creates a new OrderDelivered event.
func NewOrderDelivered(
	shipmentID valueobject.ShipmentID,
	orderID string,
) *OrderDelivered {
	now := time.Now().UTC()
	return &OrderDelivered{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "OrderDelivered",
			AggregateIDV: shipmentID.String(),
			OccurredAtV:  now,
		},
		OrderID:     orderID,
		DeliveredAt: now,
	}
}
