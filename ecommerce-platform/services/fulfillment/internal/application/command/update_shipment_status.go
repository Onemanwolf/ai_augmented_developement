// Package command contains command definitions for the Fulfillment service.
package command

import "github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"

// UpdateShipmentStatus represents a command to update a shipment's status.
type UpdateShipmentStatus struct {
	ShipmentID string                     `json:"shipment_id"`
	NewStatus  valueobject.ShipmentStatus `json:"new_status"`
	Reason     string                     `json:"reason,omitempty"` // Required for FAILED status
}

// Validate validates the command.
func (c *UpdateShipmentStatus) Validate() error {
	if c.ShipmentID == "" {
		return ErrShipmentIDRequired
	}
	if !c.NewStatus.IsValid() {
		return ErrInvalidStatus
	}
	if c.NewStatus == valueobject.ShipmentStatusFailed && c.Reason == "" {
		return ErrReasonRequired
	}
	return nil
}
