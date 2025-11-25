// Package repository defines repository interfaces for the Payment domain.
package repository

import (
	"context"

	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentRepository defines the interface for payment persistence.
type PaymentRepository interface {
	// Save persists a payment (insert or update).
	Save(ctx context.Context, payment *aggregate.Payment) error

	// FindByID retrieves a payment by its ID.
	FindByID(ctx context.Context, id valueobject.PaymentID) (*aggregate.Payment, error)

	// FindByOrderID retrieves the payment for an order.
	FindByOrderID(ctx context.Context, orderID string) (*aggregate.Payment, error)

	// FindByCustomerID retrieves all payments for a customer.
	FindByCustomerID(ctx context.Context, customerID string) ([]*aggregate.Payment, error)

	// FindByStatus retrieves payments with the given status.
	FindByStatus(ctx context.Context, status valueobject.PaymentStatus) ([]*aggregate.Payment, error)

	// Delete removes a payment by its ID.
	Delete(ctx context.Context, id valueobject.PaymentID) error

	// Exists checks if a payment exists.
	Exists(ctx context.Context, id valueobject.PaymentID) (bool, error)
}
