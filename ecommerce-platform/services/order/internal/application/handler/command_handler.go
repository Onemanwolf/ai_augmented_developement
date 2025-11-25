// Package handler contains command and query handlers for the Order service.
package handler

import (
	"context"
	"fmt"

	"github.com/your-org/ecommerce-platform/services/order/internal/application/command"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderCommandHandler handles order commands.
type OrderCommandHandler struct {
	repo           repository.OrderRepository
	eventPublisher EventPublisher
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, events []interface{}) error
}

// NewOrderCommandHandler creates a new OrderCommandHandler.
func NewOrderCommandHandler(repo repository.OrderRepository, publisher EventPublisher) *OrderCommandHandler {
	return &OrderCommandHandler{
		repo:           repo,
		eventPublisher: publisher,
	}
}

// HandleCreateOrder handles the CreateOrder command.
func (h *OrderCommandHandler) HandleCreateOrder(ctx context.Context, cmd *command.CreateOrder) (*aggregate.Order, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Convert command items to domain entities
	items := make([]*entity.OrderItem, len(cmd.Items))
	for i, item := range cmd.Items {
		unitPrice, err := valueobject.NewMoney(item.UnitPrice, cmd.Currency)
		if err != nil {
			return nil, fmt.Errorf("invalid unit price for item %s: %w", item.ProductID, err)
		}

		orderItem, err := entity.NewOrderItem(item.ProductID, item.Name, item.Quantity, unitPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
		items[i] = orderItem
	}

	// Create the order aggregate
	customerID, err := valueobject.ParseCustomerID(cmd.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	order, err := aggregate.NewOrder(customerID, items, cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Capture events BEFORE saving (Save clears events for outbox pattern)
	events := order.Events()

	// Save the order (this will also save events to outbox and clear them)
	if err := h.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Publish domain events to Kafka for real-time consumers
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		if err := h.eventPublisher.Publish(ctx, eventInterfaces); err != nil {
			// Log but don't fail - events are also saved to outbox as backup
			fmt.Printf("failed to publish events: %v\n", err)
		}
	}

	return order, nil
}

// HandleCancelOrder handles the CancelOrder command.
func (h *OrderCommandHandler) HandleCancelOrder(ctx context.Context, cmd *command.CancelOrder) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	orderID, err := valueobject.ParseOrderID(cmd.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	if err := order.Cancel(cmd.Reason); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Capture events BEFORE saving
	events := order.Events()

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish domain events to Kafka
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		if err := h.eventPublisher.Publish(ctx, eventInterfaces); err != nil {
			fmt.Printf("failed to publish events: %v\n", err)
		}
	}

	return nil
}

// HandleUpdateOrderStatus handles the UpdateOrderStatus command.
func (h *OrderCommandHandler) HandleUpdateOrderStatus(ctx context.Context, cmd *command.UpdateOrderStatus) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	orderID, err := valueobject.ParseOrderID(cmd.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Handle status transitions
	switch cmd.NewStatus {
	case valueobject.OrderStatusPaid:
		if err := order.MarkAsPaid(); err != nil {
			return fmt.Errorf("failed to mark order as paid: %w", err)
		}
	case valueobject.OrderStatusShipped:
		if err := order.MarkAsShipped(); err != nil {
			return fmt.Errorf("failed to mark order as shipped: %w", err)
		}
	case valueobject.OrderStatusDelivered:
		if err := order.MarkAsDelivered(); err != nil {
			return fmt.Errorf("failed to mark order as delivered: %w", err)
		}
	case valueobject.OrderStatusCompleted:
		if err := order.Complete(); err != nil {
			return fmt.Errorf("failed to complete order: %w", err)
		}
	default:
		return fmt.Errorf("unsupported status transition to %s", cmd.NewStatus)
	}

	// Capture events BEFORE saving
	events := order.Events()

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish domain events to Kafka
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		if err := h.eventPublisher.Publish(ctx, eventInterfaces); err != nil {
			fmt.Printf("failed to publish events: %v\n", err)
		}
	}

	return nil
}
