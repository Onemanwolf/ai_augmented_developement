// Package saga contains the SAGA orchestrator for order processing.
package saga

import "context"

// SagaRepository defines the interface for saga persistence.
type SagaRepository interface {
	// Save persists a saga (insert or update).
	Save(ctx context.Context, saga *OrderSagaData) error

	// FindByID retrieves a saga by its ID.
	FindByID(ctx context.Context, id string) (*OrderSagaData, error)

	// FindByOrderID retrieves a saga by order ID.
	FindByOrderID(ctx context.Context, orderID string) (*OrderSagaData, error)

	// FindPending retrieves sagas that need processing.
	FindPending(ctx context.Context, limit int) ([]*OrderSagaData, error)

	// FindStale retrieves sagas that may be stuck.
	FindStale(ctx context.Context, olderThan int64, limit int) ([]*OrderSagaData, error)
}
