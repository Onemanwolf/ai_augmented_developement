// Package query contains query definitions for the Order service.
package query

import "errors"

// Query validation errors.
var (
	ErrOrderIDRequired    = errors.New("order ID is required")
	ErrCustomerIDRequired = errors.New("customer ID is required")
	ErrStatusRequired     = errors.New("status is required")
)
