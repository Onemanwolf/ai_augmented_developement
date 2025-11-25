package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentFailed is emitted when a shipment delivery fails.
type ShipmentFailed struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// NewShipmentFailed creates a new ShipmentFailed event.
func NewShipmentFailed(
	shipmentID valueobject.ShipmentID,
	orderID string,
	reason string,
) *ShipmentFailed {
	return &ShipmentFailed{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "ShipmentFailed",
			AggregateIDV: shipmentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID: orderID,
		Reason:  reason,
	}
}
