// Package command contains command definitions for the Order service.
package command

import "errors"

// Command validation errors.
var (
	ErrCustomerIDRequired  = errors.New("customer ID is required")
	ErrOrderIDRequired     = errors.New("order ID is required")
	ErrItemsRequired       = errors.New("at least one item is required")
	ErrProductIDRequired   = errors.New("product ID is required")
	ErrProductNameRequired = errors.New("product name is required")
	ErrInvalidQuantity     = errors.New("quantity must be positive")
	ErrInvalidPrice        = errors.New("price must be positive")
	ErrInvalidCurrency     = errors.New("invalid currency")
	ErrInvalidStatus       = errors.New("invalid order status")
	ErrReasonRequired      = errors.New("reason is required")
)
