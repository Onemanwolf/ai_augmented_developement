package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

// CustomerID represents a unique customer identifier.
type CustomerID string

// NewCustomerID generates a new CustomerID.
func NewCustomerID() CustomerID {
	return CustomerID(uuid.New().String())
}

// ParseCustomerID creates a CustomerID from a string, validating the format.
func ParseCustomerID(s string) (CustomerID, error) {
	if s == "" {
		return "", fmt.Errorf("customer ID cannot be empty")
	}
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid customer ID format: %w", err)
	}
	return CustomerID(s), nil
}

// String returns the string representation of the CustomerID.
func (id CustomerID) String() string {
	return string(id)
}

// IsEmpty returns true if the CustomerID is empty.
func (id CustomerID) IsEmpty() bool {
	return id == ""
}

// Equals checks if two CustomerIDs are equal.
func (id CustomerID) Equals(other CustomerID) bool {
	return id == other
}
