// Package handler contains command and query handlers for the Fulfillment service.
package handler

import (
	"context"
	"fmt"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/query"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentQueryHandler handles shipment queries.
type ShipmentQueryHandler struct {
	repo repository.ShipmentRepository
}

// NewShipmentQueryHandler creates a new ShipmentQueryHandler.
func NewShipmentQueryHandler(repo repository.ShipmentRepository) *ShipmentQueryHandler {
	return &ShipmentQueryHandler{repo: repo}
}

// HandleGetShipment handles the GetShipment query.
func (h *ShipmentQueryHandler) HandleGetShipment(ctx context.Context, q *query.GetShipment) (*query.ShipmentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	shipmentID, err := valueobject.ParseShipmentID(q.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid shipment ID: %w", err)
	}

	shipment, err := h.repo.FindByID(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}

	return query.FromAggregate(shipment), nil
}

// HandleGetShipmentByOrder handles the GetShipmentByOrder query.
func (h *ShipmentQueryHandler) HandleGetShipmentByOrder(ctx context.Context, q *query.GetShipmentByOrder) (*query.ShipmentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	shipment, err := h.repo.FindByOrderID(ctx, q.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}

	return query.FromAggregate(shipment), nil
}

// HandleGetShipmentByTracking handles the GetShipmentByTracking query.
func (h *ShipmentQueryHandler) HandleGetShipmentByTracking(ctx context.Context, q *query.GetShipmentByTracking) (*query.ShipmentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	shipment, err := h.repo.FindByTrackingNumber(ctx, q.TrackingNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}

	return query.FromAggregate(shipment), nil
}

// HandleGetShipmentsByCustomer handles the GetShipmentsByCustomer query.
func (h *ShipmentQueryHandler) HandleGetShipmentsByCustomer(ctx context.Context, q *query.GetShipmentsByCustomer) ([]*query.ShipmentDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	shipments, err := h.repo.FindByCustomerID(ctx, q.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shipments: %w", err)
	}

	return query.FromAggregateList(shipments), nil
}
