// Package handler contains command and query handlers for the Payment service.
package handler

import (
	"context"
	"fmt"

	"github.com/your-org/ecommerce-platform/services/payment/internal/application/query"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentQueryHandler handles payment queries.
type PaymentQueryHandler struct {
	repo repository.PaymentRepository
}

// NewPaymentQueryHandler creates a new PaymentQueryHandler.
func NewPaymentQueryHandler(repo repository.PaymentRepository) *PaymentQueryHandler {
	return &PaymentQueryHandler{repo: repo}
}

// HandleGetPayment handles the GetPayment query.
func (h *PaymentQueryHandler) HandleGetPayment(ctx context.Context, q *query.GetPayment) (*query.PaymentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	paymentID, err := valueobject.ParsePaymentID(q.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	payment, err := h.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return query.FromAggregate(payment), nil
}

// HandleGetPaymentByOrder handles the GetPaymentByOrder query.
func (h *PaymentQueryHandler) HandleGetPaymentByOrder(ctx context.Context, q *query.GetPaymentByOrder) (*query.PaymentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	payment, err := h.repo.FindByOrderID(ctx, q.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return query.FromAggregate(payment), nil
}

// HandleGetPaymentsByCustomer handles the GetPaymentsByCustomer query.
func (h *PaymentQueryHandler) HandleGetPaymentsByCustomer(ctx context.Context, q *query.GetPaymentsByCustomer) ([]*query.PaymentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	payments, err := h.repo.FindByCustomerID(ctx, q.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}

	return query.FromAggregateList(payments), nil
}
