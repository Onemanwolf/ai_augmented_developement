// Package valueobject contains value objects for the Payment domain.
package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

// PaymentID represents a unique payment identifier.
type PaymentID string

// NewPaymentID creates a new PaymentID.
func NewPaymentID() PaymentID {
	return PaymentID(uuid.New().String())
}

// ParsePaymentID parses a string into a PaymentID.
func ParsePaymentID(s string) (PaymentID, error) {
	if s == "" {
		return "", fmt.Errorf("payment ID cannot be empty")
	}
	if _, err := uuid.Parse(s); err != nil {
		return "", fmt.Errorf("invalid payment ID format: %w", err)
	}
	return PaymentID(s), nil
}

// String returns the string representation.
func (id PaymentID) String() string {
	return string(id)
}

// IsEmpty checks if the ID is empty.
func (id PaymentID) IsEmpty() bool {
	return id == ""
}
