package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// OrderShipped is emitted when an order is shipped.
type OrderShipped struct {
	BaseEvent
	OrderID        string              `json:"order_id"`
	TrackingNumber string              `json:"tracking_number"`
	Carrier        valueobject.Carrier `json:"carrier"`
}

// NewOrderShipped creates a new OrderShipped event.
func NewOrderShipped(
	shipmentID valueobject.ShipmentID,
	orderID string,
	trackingNumber string,
	carrier valueobject.Carrier,
) *OrderShipped {
	return &OrderShipped{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "OrderShipped",
			AggregateIDV: shipmentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID:        orderID,
		TrackingNumber: trackingNumber,
		Carrier:        carrier,
	}
}
