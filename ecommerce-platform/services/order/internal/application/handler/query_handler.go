// Package handler contains command and query handlers for the Order service.
package handler

import (
	"context"
	"fmt"

	"github.com/your-org/ecommerce-platform/services/order/internal/application/query"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderQueryHandler handles order queries.
type OrderQueryHandler struct {
	repo repository.OrderRepository
}

// NewOrderQueryHandler creates a new OrderQueryHandler.
func NewOrderQueryHandler(repo repository.OrderRepository) *OrderQueryHandler {
	return &OrderQueryHandler{repo: repo}
}

// HandleGetOrder handles the GetOrder query.
func (h *OrderQueryHandler) HandleGetOrder(ctx context.Context, q *query.GetOrder) (*query.OrderDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	orderID, err := valueobject.ParseOrderID(q.OrderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	return query.FromAggregate(order), nil
}

// HandleGetOrdersByCustomer handles the GetOrdersByCustomer query.
func (h *OrderQueryHandler) HandleGetOrdersByCustomer(ctx context.Context, q *query.GetOrdersByCustomer) ([]*query.OrderDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	customerID, err := valueobject.ParseCustomerID(q.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	orders, err := h.repo.FindByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}

	return query.FromAggregateList(orders), nil
}

// HandleGetOrdersByStatus handles the GetOrdersByStatus query.
func (h *OrderQueryHandler) HandleGetOrdersByStatus(ctx context.Context, q *query.GetOrdersByStatus) ([]*query.OrderDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	status, err := valueobject.ParseOrderStatus(q.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	orders, err := h.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}

	return query.FromAggregateList(orders), nil
}
