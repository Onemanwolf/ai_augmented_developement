// Package valueobject contains value objects for the Order domain.
package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

// OrderID represents a unique order identifier.
type OrderID string

// NewOrderID generates a new OrderID.
func NewOrderID() OrderID {
	return OrderID(uuid.New().String())
}

// ParseOrderID creates an OrderID from a string, validating the format.
func ParseOrderID(s string) (OrderID, error) {
	if s == "" {
		return "", fmt.Errorf("order ID cannot be empty")
	}
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid order ID format: %w", err)
	}
	return OrderID(s), nil
}

// String returns the string representation of the OrderID.
func (id OrderID) String() string {
	return string(id)
}

// IsEmpty returns true if the OrderID is empty.
func (id OrderID) IsEmpty() bool {
	return id == ""
}

// Equals checks if two OrderIDs are equal.
func (id OrderID) Equals(other OrderID) bool {
	return id == other
}
