package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentCreated is emitted when a new shipment is created.
type ShipmentCreated struct {
	BaseEvent
	OrderID         string              `json:"order_id"`
	CustomerID      string              `json:"customer_id"`
	Carrier         valueobject.Carrier `json:"carrier"`
	ShippingAddress valueobject.Address `json:"shipping_address"`
}

// NewShipmentCreated creates a new ShipmentCreated event.
func NewShipmentCreated(
	shipmentID valueobject.ShipmentID,
	orderID string,
	customerID string,
	carrier valueobject.Carrier,
	address valueobject.Address,
) *ShipmentCreated {
	return &ShipmentCreated{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "ShipmentCreated",
			AggregateIDV: shipmentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID:         orderID,
		CustomerID:      customerID,
		Carrier:         carrier,
		ShippingAddress: address,
	}
}
