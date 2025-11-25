// Package saga handles saga events for the Fulfillment service.
package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
	"github.com/your-org/ecommerce-platform/shared/pkg/kafka"
	"github.com/your-org/ecommerce-platform/shared/pkg/saga"
)

// EventHandler handles saga events for the Fulfillment service.
type EventHandler struct {
	repo           repository.ShipmentRepository
	eventPublisher *kafka.EventPublisher
}

// NewEventHandler creates a new saga event handler.
func NewEventHandler(
	repo repository.ShipmentRepository,
	publisher *kafka.EventPublisher,
) *EventHandler {
	return &EventHandler{
		repo:           repo,
		eventPublisher: publisher,
	}
}

// HandlePaymentProcessed processes a PaymentProcessed event by creating a shipment.
func (h *EventHandler) HandlePaymentProcessed(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Fulfillment SAGA] Received PaymentProcessed event")

	// Parse the event envelope
	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	// Extract payment data from payload
	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var paymentEvent saga.PaymentProcessedEvent
	if err := json.Unmarshal(payloadBytes, &paymentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payment event: %w", err)
	}

	log.Printf("[Fulfillment SAGA] Creating shipment for order %s", paymentEvent.OrderID)

	// We need to get order details - for now create a basic shipment
	// In production, you'd fetch order details from order service or include in event
	customerID := paymentEvent.CustomerID
	if customerID == "" {
		return h.publishShipmentFailed(ctx, paymentEvent.OrderID, paymentEvent.CustomerID, "customer ID is required")
	}

	// Create default address (in production, this would come from order/customer data)
	address, err := valueobject.NewAddress(
		"123 Default Street",
		"Default City",
		"DS",
		"12345",
		"US",
	)
	if err != nil {
		return h.publishShipmentFailed(ctx, paymentEvent.OrderID, paymentEvent.CustomerID, err.Error())
	}

	// Create a shipment item (in production, these would come from order data)
	weight, _ := entity.NewWeight(500) // 500 grams default
	item, err := entity.NewShipmentItem("default-product", "Order Items", 1, weight)
	if err != nil {
		return h.publishShipmentFailed(ctx, paymentEvent.OrderID, paymentEvent.CustomerID, err.Error())
	}

	// Create shipment
	shipment, err := aggregate.NewShipment(
		paymentEvent.OrderID,
		customerID,
		[]*entity.ShipmentItem{item},
		address,
		valueobject.CarrierFedEx,
	)
	if err != nil {
		return h.publishShipmentFailed(ctx, paymentEvent.OrderID, paymentEvent.CustomerID, err.Error())
	}

	// Save shipment
	if err := h.repo.Save(ctx, shipment); err != nil {
		return h.publishShipmentFailed(ctx, paymentEvent.OrderID, paymentEvent.CustomerID, err.Error())
	}

	// Publish ShipmentCreated event
	return h.publishShipmentCreated(ctx, shipment, paymentEvent.CustomerID)
}

// HandlePaymentRefunded handles a payment refund (cancel any pending shipment).
func (h *EventHandler) HandlePaymentRefunded(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Fulfillment SAGA] Received PaymentRefunded event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var refundEvent saga.PaymentRefundedEvent
	if err := json.Unmarshal(payloadBytes, &refundEvent); err != nil {
		return fmt.Errorf("failed to unmarshal refund event: %w", err)
	}

	// Find and cancel any pending shipment for this order
	shipment, err := h.repo.FindByOrderID(ctx, refundEvent.OrderID)
	if err != nil {
		log.Printf("[Fulfillment SAGA] No shipment found for order %s: %v", refundEvent.OrderID, err)
		return nil // Not an error - shipment may not exist yet
	}

	// Cancel takes no arguments - just cancels the shipment
	if err := shipment.Cancel(); err != nil {
		log.Printf("[Fulfillment SAGA] Failed to cancel shipment: %v (reason: Payment was refunded: %s)", err, refundEvent.Reason)
		return nil // Shipment may already be shipped
	}

	return h.repo.Save(ctx, shipment)
}

func (h *EventHandler) publishShipmentCreated(ctx context.Context, shipment *aggregate.Shipment, customerID string) error {
	event := saga.ShipmentCreatedEvent{
		ShipmentID: shipment.ID.String(),
		OrderID:    shipment.OrderID,
		CustomerID: customerID,
		OccurredAt: time.Now().UTC(),
	}

	log.Printf("[Fulfillment SAGA] Publishing ShipmentCreated for order %s", shipment.OrderID)
	return h.eventPublisher.Publish(ctx, []interface{}{&shipmentCreatedWrapper{event}})
}

func (h *EventHandler) publishShipmentFailed(ctx context.Context, orderID, customerID, reason string) error {
	event := saga.ShipmentFailedEvent{
		OrderID:    orderID,
		CustomerID: customerID,
		Reason:     reason,
		OccurredAt: time.Now().UTC(),
	}

	log.Printf("[Fulfillment SAGA] Publishing ShipmentFailed for order %s: %s", orderID, reason)
	return h.eventPublisher.Publish(ctx, []interface{}{&shipmentFailedWrapper{event}})
}

// Event wrappers to implement the event interface
type shipmentCreatedWrapper struct {
	saga.ShipmentCreatedEvent
}

func (e *shipmentCreatedWrapper) EventType() string     { return saga.EventShipmentCreated }
func (e *shipmentCreatedWrapper) AggregateID() string   { return e.OrderID }
func (e *shipmentCreatedWrapper) AggregateType() string { return "Shipment" }

type shipmentFailedWrapper struct {
	saga.ShipmentFailedEvent
}

func (e *shipmentFailedWrapper) EventType() string     { return saga.EventShipmentFailed }
func (e *shipmentFailedWrapper) AggregateID() string   { return e.OrderID }
func (e *shipmentFailedWrapper) AggregateType() string { return "Shipment" }
