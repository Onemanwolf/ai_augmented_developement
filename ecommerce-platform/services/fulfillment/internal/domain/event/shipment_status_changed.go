package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentStatusChanged is emitted when a shipment's status changes.
type ShipmentStatusChanged struct {
	BaseEvent
	OldStatus valueobject.ShipmentStatus `json:"old_status"`
	NewStatus valueobject.ShipmentStatus `json:"new_status"`
}

// NewShipmentStatusChanged creates a new ShipmentStatusChanged event.
func NewShipmentStatusChanged(
	shipmentID valueobject.ShipmentID,
	oldStatus valueobject.ShipmentStatus,
	newStatus valueobject.ShipmentStatus,
) *ShipmentStatusChanged {
	return &ShipmentStatusChanged{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "ShipmentStatusChanged",
			AggregateIDV: shipmentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}
}
