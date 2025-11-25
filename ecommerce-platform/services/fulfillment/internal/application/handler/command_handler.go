// Package handler contains command and query handlers for the Fulfillment service.
package handler

import (
	"context"
	"fmt"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/command"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentCommandHandler handles shipment commands.
type ShipmentCommandHandler struct {
	repo           repository.ShipmentRepository
	eventPublisher EventPublisher
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, events []interface{}) error
}

// NewShipmentCommandHandler creates a new ShipmentCommandHandler.
func NewShipmentCommandHandler(repo repository.ShipmentRepository, publisher EventPublisher) *ShipmentCommandHandler {
	return &ShipmentCommandHandler{
		repo:           repo,
		eventPublisher: publisher,
	}
}

// HandleCreateShipment handles the CreateShipment command.
func (h *ShipmentCommandHandler) HandleCreateShipment(ctx context.Context, cmd *command.CreateShipment) (*aggregate.Shipment, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Convert command items to domain entities
	items := make([]*entity.ShipmentItem, len(cmd.Items))
	for i, item := range cmd.Items {
		weight, err := entity.NewWeight(item.WeightGrams)
		if err != nil {
			return nil, fmt.Errorf("invalid weight for item %s: %w", item.ProductID, err)
		}

		shipmentItem, err := entity.NewShipmentItem(item.ProductID, item.Name, item.Quantity, weight)
		if err != nil {
			return nil, fmt.Errorf("failed to create shipment item: %w", err)
		}
		items[i] = shipmentItem
	}

	// Create address value object
	address, err := valueobject.NewAddress(
		cmd.ShippingAddress.Street,
		cmd.ShippingAddress.City,
		cmd.ShippingAddress.State,
		cmd.ShippingAddress.PostalCode,
		cmd.ShippingAddress.Country,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping address: %w", err)
	}

	// Create shipment aggregate
	shipment, err := aggregate.NewShipment(cmd.OrderID, cmd.CustomerID, items, address, cmd.Carrier)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Save the shipment
	if err := h.repo.Save(ctx, shipment); err != nil {
		return nil, fmt.Errorf("failed to save shipment: %w", err)
	}

	h.publishEvents(ctx, shipment)
	return shipment, nil
}

// HandleShipOrder handles the ShipOrder command.
func (h *ShipmentCommandHandler) HandleShipOrder(ctx context.Context, cmd *command.ShipOrder) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	shipmentID, err := valueobject.ParseShipmentID(cmd.ShipmentID)
	if err != nil {
		return fmt.Errorf("invalid shipment ID: %w", err)
	}

	shipment, err := h.repo.FindByID(ctx, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to find shipment: %w", err)
	}

	// Start processing if still pending
	if shipment.IsPending() {
		if err := shipment.StartProcessing(); err != nil {
			return fmt.Errorf("failed to start processing: %w", err)
		}
	}

	// Mark ready for pickup with tracking info
	if err := shipment.MarkReadyForPickup(cmd.TrackingNumber, cmd.EstimatedDelivery); err != nil {
		return fmt.Errorf("failed to mark ready for pickup: %w", err)
	}

	// Mark as shipped (picked up by carrier)
	if err := shipment.Ship(); err != nil {
		return fmt.Errorf("failed to ship: %w", err)
	}

	if err := h.repo.Save(ctx, shipment); err != nil {
		return fmt.Errorf("failed to save shipment: %w", err)
	}

	h.publishEvents(ctx, shipment)
	return nil
}

// HandleUpdateShipmentStatus handles the UpdateShipmentStatus command.
func (h *ShipmentCommandHandler) HandleUpdateShipmentStatus(ctx context.Context, cmd *command.UpdateShipmentStatus) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	shipmentID, err := valueobject.ParseShipmentID(cmd.ShipmentID)
	if err != nil {
		return fmt.Errorf("invalid shipment ID: %w", err)
	}

	shipment, err := h.repo.FindByID(ctx, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to find shipment: %w", err)
	}

	// Handle status transitions
	switch cmd.NewStatus {
	case valueobject.ShipmentStatusProcessing:
		if err := shipment.StartProcessing(); err != nil {
			return fmt.Errorf("failed to start processing: %w", err)
		}
	case valueobject.ShipmentStatusInTransit:
		if err := shipment.UpdateToInTransit(); err != nil {
			return fmt.Errorf("failed to update to in transit: %w", err)
		}
	case valueobject.ShipmentStatusOutForDelivery:
		if err := shipment.UpdateToOutForDelivery(); err != nil {
			return fmt.Errorf("failed to update to out for delivery: %w", err)
		}
	case valueobject.ShipmentStatusDelivered:
		if err := shipment.MarkDelivered(); err != nil {
			return fmt.Errorf("failed to mark delivered: %w", err)
		}
	case valueobject.ShipmentStatusFailed:
		if err := shipment.MarkFailed(cmd.Reason); err != nil {
			return fmt.Errorf("failed to mark as failed: %w", err)
		}
	case valueobject.ShipmentStatusReturned:
		if err := shipment.MarkReturned(cmd.Reason); err != nil {
			return fmt.Errorf("failed to mark as returned: %w", err)
		}
	default:
		return fmt.Errorf("unsupported status transition to %s", cmd.NewStatus)
	}

	if err := h.repo.Save(ctx, shipment); err != nil {
		return fmt.Errorf("failed to save shipment: %w", err)
	}

	h.publishEvents(ctx, shipment)
	return nil
}

// HandleCancelShipment handles the CancelShipment command.
func (h *ShipmentCommandHandler) HandleCancelShipment(ctx context.Context, cmd *command.CancelShipment) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	shipmentID, err := valueobject.ParseShipmentID(cmd.ShipmentID)
	if err != nil {
		return fmt.Errorf("invalid shipment ID: %w", err)
	}

	shipment, err := h.repo.FindByID(ctx, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to find shipment: %w", err)
	}

	if err := shipment.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel shipment: %w", err)
	}

	if err := h.repo.Save(ctx, shipment); err != nil {
		return fmt.Errorf("failed to save shipment: %w", err)
	}

	h.publishEvents(ctx, shipment)
	return nil
}

func (h *ShipmentCommandHandler) publishEvents(ctx context.Context, shipment *aggregate.Shipment) {
	events := shipment.GetEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		if err := h.eventPublisher.Publish(ctx, eventInterfaces); err != nil {
			fmt.Printf("failed to publish events: %v\n", err)
		}
	}
}
