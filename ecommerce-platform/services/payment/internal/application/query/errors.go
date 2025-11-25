// Package query contains query definitions for the Payment service.
package query

import "errors"

// Query validation errors.
var (
	ErrPaymentIDRequired  = errors.New("payment ID is required")
	ErrOrderIDRequired    = errors.New("order ID is required")
	ErrCustomerIDRequired = errors.New("customer ID is required")
	ErrPaymentNotFound    = errors.New("payment not found")
)
