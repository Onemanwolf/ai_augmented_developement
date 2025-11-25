// Package valueobject contains value objects for the Payment domain.
package valueobject

import "fmt"

// PaymentStatus represents the status of a payment.
type PaymentStatus string

const (
	// PaymentStatusPending indicates payment is awaiting processing.
	PaymentStatusPending PaymentStatus = "PENDING"
	// PaymentStatusProcessing indicates payment is being processed.
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	// PaymentStatusCompleted indicates payment was successful.
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	// PaymentStatusFailed indicates payment failed.
	PaymentStatusFailed PaymentStatus = "FAILED"
	// PaymentStatusRefunded indicates payment was refunded.
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
	// PaymentStatusPartiallyRefunded indicates partial refund was issued.
	PaymentStatusPartiallyRefunded PaymentStatus = "PARTIALLY_REFUNDED"
	// PaymentStatusCancelled indicates payment was cancelled.
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
)

// validTransitions defines allowed state transitions.
var validPaymentTransitions = map[PaymentStatus][]PaymentStatus{
	PaymentStatusPending:           {PaymentStatusProcessing, PaymentStatusCancelled},
	PaymentStatusProcessing:        {PaymentStatusCompleted, PaymentStatusFailed},
	PaymentStatusCompleted:         {PaymentStatusRefunded, PaymentStatusPartiallyRefunded},
	PaymentStatusFailed:            {PaymentStatusPending}, // Allow retry
	PaymentStatusRefunded:          {},                     // Terminal state
	PaymentStatusPartiallyRefunded: {PaymentStatusRefunded},
	PaymentStatusCancelled:         {}, // Terminal state
}

// IsValid checks if the status is valid.
func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted,
		PaymentStatusFailed, PaymentStatusRefunded, PaymentStatusPartiallyRefunded,
		PaymentStatusCancelled:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if transition to target status is allowed.
func (s PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	allowed, ok := validPaymentTransitions[s]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

// String returns the string representation.
func (s PaymentStatus) String() string {
	return string(s)
}

// ParsePaymentStatus parses a string into a PaymentStatus.
func ParsePaymentStatus(s string) (PaymentStatus, error) {
	status := PaymentStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid payment status: %s", s)
	}
	return status, nil
}

// IsTerminal checks if the status is a terminal state.
func (s PaymentStatus) IsTerminal() bool {
	return s == PaymentStatusRefunded || s == PaymentStatusCancelled
}

// IsSuccessful checks if the payment was successful.
func (s PaymentStatus) IsSuccessful() bool {
	return s == PaymentStatusCompleted || s == PaymentStatusPartiallyRefunded
}
