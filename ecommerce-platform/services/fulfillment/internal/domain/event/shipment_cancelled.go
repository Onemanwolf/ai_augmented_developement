package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentCancelled is emitted when a shipment is cancelled.
type ShipmentCancelled struct {
	BaseEvent
	OrderID string `json:"order_id"`
}

// NewShipmentCancelled creates a new ShipmentCancelled event.
func NewShipmentCancelled(
	shipmentID valueobject.ShipmentID,
	orderID string,
) *ShipmentCancelled {
	return &ShipmentCancelled{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "ShipmentCancelled",
			AggregateIDV: shipmentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID: orderID,
	}
}
