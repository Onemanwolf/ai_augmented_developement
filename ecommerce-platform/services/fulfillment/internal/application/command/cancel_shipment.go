// Package command contains command definitions for the Fulfillment service.
package command

// CancelShipment represents a command to cancel a shipment.
type CancelShipment struct {
	ShipmentID string `json:"shipment_id"`
}

// Validate validates the command.
func (c *CancelShipment) Validate() error {
	if c.ShipmentID == "" {
		return ErrShipmentIDRequired
	}
	return nil
}
