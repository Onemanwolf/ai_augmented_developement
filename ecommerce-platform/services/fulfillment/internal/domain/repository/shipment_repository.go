// Package repository defines repository interfaces for the Fulfillment domain.
package repository

import (
	"context"
	"errors"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// Repository errors.
var (
	ErrShipmentNotFound = errors.New("shipment not found")
)

// ShipmentRepository defines the interface for shipment persistence.
type ShipmentRepository interface {
	// Save persists a shipment (insert or update).
	Save(ctx context.Context, shipment *aggregate.Shipment) error

	// FindByID retrieves a shipment by its ID.
	FindByID(ctx context.Context, id valueobject.ShipmentID) (*aggregate.Shipment, error)

	// FindByOrderID retrieves the shipment for an order.
	FindByOrderID(ctx context.Context, orderID string) (*aggregate.Shipment, error)

	// FindByCustomerID retrieves all shipments for a customer.
	FindByCustomerID(ctx context.Context, customerID string) ([]*aggregate.Shipment, error)

	// FindByStatus retrieves shipments with the given status.
	FindByStatus(ctx context.Context, status valueobject.ShipmentStatus) ([]*aggregate.Shipment, error)

	// FindByTrackingNumber retrieves a shipment by tracking number.
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*aggregate.Shipment, error)

	// Delete removes a shipment by its ID.
	Delete(ctx context.Context, id valueobject.ShipmentID) error

	// Exists checks if a shipment exists.
	Exists(ctx context.Context, id valueobject.ShipmentID) (bool, error)
}
