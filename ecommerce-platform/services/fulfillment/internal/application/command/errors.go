// Package command contains command definitions for the Fulfillment service.
package command

import "errors"

// Command validation errors.
var (
	ErrShipmentIDRequired        = errors.New("shipment ID is required")
	ErrOrderIDRequired           = errors.New("order ID is required")
	ErrCustomerIDRequired        = errors.New("customer ID is required")
	ErrItemsRequired             = errors.New("at least one item is required")
	ErrProductIDRequired         = errors.New("product ID is required")
	ErrProductNameRequired       = errors.New("product name is required")
	ErrInvalidQuantity           = errors.New("quantity must be positive")
	ErrInvalidCarrier            = errors.New("invalid carrier")
	ErrAddressRequired           = errors.New("complete shipping address is required")
	ErrTrackingNumberRequired    = errors.New("tracking number is required")
	ErrEstimatedDeliveryRequired = errors.New("estimated delivery date is required")
	ErrInvalidStatus             = errors.New("invalid shipment status")
	ErrReasonRequired            = errors.New("reason is required for failed status")
)
