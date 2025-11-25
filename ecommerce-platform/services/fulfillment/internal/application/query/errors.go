// Package query contains query definitions for the Fulfillment service.
package query

import "errors"

// Query validation errors.
var (
	ErrShipmentIDRequired     = errors.New("shipment ID is required")
	ErrOrderIDRequired        = errors.New("order ID is required")
	ErrCustomerIDRequired     = errors.New("customer ID is required")
	ErrTrackingNumberRequired = errors.New("tracking number is required")
)
