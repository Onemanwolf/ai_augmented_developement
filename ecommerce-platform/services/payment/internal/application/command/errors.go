// Package command contains command definitions for the Payment service.
package command

import "errors"

// Command validation errors.
var (
	ErrPaymentIDRequired    = errors.New("payment ID is required")
	ErrOrderIDRequired      = errors.New("order ID is required")
	ErrCustomerIDRequired   = errors.New("customer ID is required")
	ErrInvalidAmount        = errors.New("amount must be positive")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrReasonRequired       = errors.New("reason is required")
)
