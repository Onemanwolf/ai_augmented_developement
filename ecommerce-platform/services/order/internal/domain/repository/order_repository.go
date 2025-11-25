// Package repository defines repository interfaces for the Order domain.
package repository

import (
	"context"

	"github.com/your-org/ecommerce-platform/services/order/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderRepository defines the interface for order persistence.
type OrderRepository interface {
	// Save persists an order (insert or update).
	Save(ctx context.Context, order *aggregate.Order) error

	// FindByID retrieves an order by its ID.
	FindByID(ctx context.Context, id valueobject.OrderID) (*aggregate.Order, error)

	// FindByCustomerID retrieves all orders for a customer.
	FindByCustomerID(ctx context.Context, customerID valueobject.CustomerID) ([]*aggregate.Order, error)

	// FindByStatus retrieves orders with the given status.
	FindByStatus(ctx context.Context, status valueobject.OrderStatus) ([]*aggregate.Order, error)

	// Delete removes an order by its ID.
	Delete(ctx context.Context, id valueobject.OrderID) error

	// Exists checks if an order exists.
	Exists(ctx context.Context, id valueobject.OrderID) (bool, error)
}
