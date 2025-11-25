// Package command contains command definitions for the Fulfillment service.
package command

import "time"

// ShipOrder represents a command to ship an order.
type ShipOrder struct {
	ShipmentID        string    `json:"shipment_id"`
	TrackingNumber    string    `json:"tracking_number"`
	EstimatedDelivery time.Time `json:"estimated_delivery"`
}

// Validate validates the command.
func (c *ShipOrder) Validate() error {
	if c.ShipmentID == "" {
		return ErrShipmentIDRequired
	}
	if c.TrackingNumber == "" {
		return ErrTrackingNumberRequired
	}
	if c.EstimatedDelivery.IsZero() {
		return ErrEstimatedDeliveryRequired
	}
	return nil
}
