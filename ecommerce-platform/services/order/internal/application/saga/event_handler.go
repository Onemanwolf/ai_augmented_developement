// Package saga handles saga events for the Order service.
package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/your-org/ecommerce-platform/services/order/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
	"github.com/your-org/ecommerce-platform/shared/pkg/kafka"
	"github.com/your-org/ecommerce-platform/shared/pkg/saga"
)

// EventHandler handles saga events for the Order service.
type EventHandler struct {
	repo           repository.OrderRepository
	eventPublisher *kafka.EventPublisher
}

// NewEventHandler creates a new saga event handler.
func NewEventHandler(
	repo repository.OrderRepository,
	publisher *kafka.EventPublisher,
) *EventHandler {
	return &EventHandler{
		repo:           repo,
		eventPublisher: publisher,
	}
}

// HandlePaymentProcessed updates order status when payment succeeds.
func (h *EventHandler) HandlePaymentProcessed(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received PaymentProcessed event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var paymentEvent saga.PaymentProcessedEvent
	if err := json.Unmarshal(payloadBytes, &paymentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payment event: %w", err)
	}

	log.Printf("[Order SAGA] Updating order %s to PAID status", paymentEvent.OrderID)

	// Find and update order
	orderID, err := valueobject.ParseOrderID(paymentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Capture events before save
	if err := order.MarkAsPaid(); err != nil {
		return fmt.Errorf("failed to mark order as paid: %w", err)
	}
	events := order.Events()

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order status change events
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		h.eventPublisher.Publish(ctx, eventInterfaces)
	}

	log.Printf("[Order SAGA] Order %s marked as PAID", paymentEvent.OrderID)
	return nil
}

// HandlePaymentFailed updates order status when payment fails.
func (h *EventHandler) HandlePaymentFailed(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received PaymentFailed event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var paymentEvent saga.PaymentFailedEvent
	if err := json.Unmarshal(payloadBytes, &paymentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payment failed event: %w", err)
	}

	log.Printf("[Order SAGA] Payment failed for order %s: %s", paymentEvent.OrderID, paymentEvent.Reason)

	// Find and cancel order
	orderID, err := valueobject.ParseOrderID(paymentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Cancel the order due to payment failure
	if err := order.Cancel("Payment failed: " + paymentEvent.Reason); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	events := order.Events()

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order status change events
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		h.eventPublisher.Publish(ctx, eventInterfaces)
	}

	log.Printf("[Order SAGA] Order %s cancelled due to payment failure", paymentEvent.OrderID)
	return nil
}

// HandleShipmentCreated updates order status when shipment is created.
func (h *EventHandler) HandleShipmentCreated(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received ShipmentCreated event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var shipmentEvent saga.ShipmentCreatedEvent
	if err := json.Unmarshal(payloadBytes, &shipmentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal shipment event: %w", err)
	}

	log.Printf("[Order SAGA] Shipment created for order %s", shipmentEvent.OrderID)

	// Find and update order - mark as processing/fulfillment started
	orderID, err := valueobject.ParseOrderID(shipmentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Order is now being fulfilled - in a real system you might have an PROCESSING status
	// For now, we'll just log it as the order is already PAID
	log.Printf("[Order SAGA] Order %s shipment created, current status: %s", shipmentEvent.OrderID, order.Status)
	return nil
}

// HandleShipmentFailed updates order status when shipment fails.
func (h *EventHandler) HandleShipmentFailed(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received ShipmentFailed event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var shipmentEvent saga.ShipmentFailedEvent
	if err := json.Unmarshal(payloadBytes, &shipmentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal shipment failed event: %w", err)
	}

	log.Printf("[Order SAGA] Shipment failed for order %s: %s", shipmentEvent.OrderID, shipmentEvent.Reason)

	// In a real saga, this would trigger payment refund (which we handle in payment service)
	// The order will be cancelled when we receive the PaymentRefunded event
	// For now, just log it
	log.Printf("[Order SAGA] Awaiting payment refund for order %s", shipmentEvent.OrderID)
	return nil
}

// HandlePaymentRefunded updates order status when payment is refunded.
func (h *EventHandler) HandlePaymentRefunded(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received PaymentRefunded event")

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

	log.Printf("[Order SAGA] Payment refunded for order %s: %s", refundEvent.OrderID, refundEvent.Reason)

	// Cancel the order
	orderID, err := valueobject.ParseOrderID(refundEvent.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	if err := order.Cancel("Payment refunded: " + refundEvent.Reason); err != nil {
		// Order might already be cancelled
		log.Printf("[Order SAGA] Could not cancel order %s: %v", refundEvent.OrderID, err)
		return nil
	}
	events := order.Events()

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order status change events
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		h.eventPublisher.Publish(ctx, eventInterfaces)
	}

	log.Printf("[Order SAGA] Order %s cancelled due to refund", refundEvent.OrderID)
	return nil
}

// HandleShipmentShipped updates order status when shipment is shipped.
func (h *EventHandler) HandleShipmentShipped(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received ShipmentShipped event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var shipmentEvent saga.ShipmentShippedEvent
	if err := json.Unmarshal(payloadBytes, &shipmentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal shipment shipped event: %w", err)
	}

	log.Printf("[Order SAGA] Order %s has been shipped", shipmentEvent.OrderID)

	orderID, err := valueobject.ParseOrderID(shipmentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	if err := order.MarkAsShipped(); err != nil {
		return fmt.Errorf("failed to mark order as shipped: %w", err)
	}
	events := order.Events()

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order status change events
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		h.eventPublisher.Publish(ctx, eventInterfaces)
	}

	log.Printf("[Order SAGA] Order %s marked as SHIPPED", shipmentEvent.OrderID)
	return nil
}

// HandleShipmentDelivered updates order status when shipment is delivered.
func (h *EventHandler) HandleShipmentDelivered(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Order SAGA] Received ShipmentDelivered event")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var shipmentEvent saga.ShipmentDeliveredEvent
	if err := json.Unmarshal(payloadBytes, &shipmentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal shipment delivered event: %w", err)
	}

	log.Printf("[Order SAGA] Order %s has been delivered", shipmentEvent.OrderID)

	orderID, err := valueobject.ParseOrderID(shipmentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	order, err := h.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}

	if err := order.MarkAsDelivered(); err != nil {
		return fmt.Errorf("failed to mark order as delivered: %w", err)
	}
	events := order.Events()

	// Complete the order
	if err := order.Complete(); err != nil {
		log.Printf("[Order SAGA] Could not complete order %s: %v", shipmentEvent.OrderID, err)
	}
	events = append(events, order.Events()...)

	if err := h.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order status change events
	if len(events) > 0 && h.eventPublisher != nil {
		eventInterfaces := make([]interface{}, len(events))
		for i, e := range events {
			eventInterfaces[i] = e
		}
		h.eventPublisher.Publish(ctx, eventInterfaces)
	}

	log.Printf("[Order SAGA] Order %s marked as DELIVERED and COMPLETED", shipmentEvent.OrderID)
	return nil
}
