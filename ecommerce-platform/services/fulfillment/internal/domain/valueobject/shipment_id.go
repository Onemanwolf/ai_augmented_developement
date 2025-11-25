// Package valueobject contains value objects for the Fulfillment domain.
package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

// ShipmentID represents a unique shipment identifier.
type ShipmentID string

// NewShipmentID creates a new ShipmentID.
func NewShipmentID() ShipmentID {
	return ShipmentID(uuid.New().String())
}

// ParseShipmentID parses a string into a ShipmentID.
func ParseShipmentID(s string) (ShipmentID, error) {
	if s == "" {
		return "", fmt.Errorf("shipment ID cannot be empty")
	}
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid shipment ID format: %w", err)
	}
	return ShipmentID(s), nil
}

// String returns the string representation.
func (id ShipmentID) String() string {
	return string(id)
}

// IsEmpty checks if the ID is empty.
func (id ShipmentID) IsEmpty() bool {
	return id == ""
}
