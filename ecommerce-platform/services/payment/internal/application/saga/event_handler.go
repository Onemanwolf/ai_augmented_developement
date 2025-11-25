// Package saga handles saga events for the Payment service.
package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
	"github.com/your-org/ecommerce-platform/shared/pkg/kafka"
	"github.com/your-org/ecommerce-platform/shared/pkg/saga"
)

// PaymentGateway defines the interface for payment processing.
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, amount valueobject.Money, method valueobject.PaymentMethod) (transactionID string, err error)
	RefundPayment(ctx context.Context, transactionID string, amount valueobject.Money) error
}

// EventHandler handles saga events for the Payment service.
type EventHandler struct {
	repo           repository.PaymentRepository
	gateway        PaymentGateway
	eventPublisher *kafka.EventPublisher
}

// NewEventHandler creates a new saga event handler.
func NewEventHandler(
	repo repository.PaymentRepository,
	gateway PaymentGateway,
	publisher *kafka.EventPublisher,
) *EventHandler {
	return &EventHandler{
		repo:           repo,
		gateway:        gateway,
		eventPublisher: publisher,
	}
}

// HandleOrderCreated processes an OrderCreated event by creating a payment.
func (h *EventHandler) HandleOrderCreated(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Payment SAGA] Received OrderCreated event")

	// Parse the event envelope
	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	// Extract order data from payload
	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var orderEvent struct {
		CustomerID  string `json:"customer_id"`
		TotalAmount struct {
			Amount   int64  `json:"amount"`
			Currency string `json:"currency"`
		} `json:"total_amount"`
	}
	if err := json.Unmarshal(payloadBytes, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal order event: %w", err)
	}

	orderID := envelope.AggregateID
	log.Printf("[Payment SAGA] Processing payment for order %s, amount: %d %s",
		orderID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency)

	// Create payment amount
	amount, err := valueobject.NewMoney(orderEvent.TotalAmount.Amount, valueobject.Currency(orderEvent.TotalAmount.Currency))
	if err != nil {
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, err.Error())
	}

	// Create the payment aggregate (customerID is a string, not a value object)
	customerID := orderEvent.CustomerID
	if customerID == "" {
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, "customer ID is required")
	}

	payment, err := aggregate.NewPayment(orderID, customerID, amount, valueobject.PaymentMethodCreditCard)
	if err != nil {
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, err.Error())
	}

	// Save payment in pending state
	if err := h.repo.Save(ctx, payment); err != nil {
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, err.Error())
	}

	// Transition to PROCESSING state before calling gateway
	if err := payment.Process(); err != nil {
		payment.Fail(err.Error())
		h.repo.Save(ctx, payment)
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, err.Error())
	}

	// Process payment through gateway
	transactionID, err := h.gateway.ProcessPayment(ctx, amount, valueobject.PaymentMethodCreditCard)
	if err != nil {
		// Mark payment as failed
		payment.Fail(err.Error())
		h.repo.Save(ctx, payment)
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, err.Error())
	}

	// Mark payment as completed (now allowed because we're in PROCESSING state)
	if err := payment.Complete(transactionID); err != nil {
		payment.Fail(err.Error())
		h.repo.Save(ctx, payment)
		return h.publishPaymentFailed(ctx, orderID, orderEvent.CustomerID, orderEvent.TotalAmount.Amount, orderEvent.TotalAmount.Currency, err.Error())
	}

	// Save completed payment
	if err := h.repo.Save(ctx, payment); err != nil {
		return fmt.Errorf("failed to save completed payment: %w", err)
	}

	// Publish PaymentProcessed event
	return h.publishPaymentProcessed(ctx, payment, transactionID, orderEvent.CustomerID)
}

// HandleShipmentFailed processes a ShipmentFailed event by refunding the payment.
func (h *EventHandler) HandleShipmentFailed(ctx context.Context, msg *kafka.ConsumedMessage) error {
	log.Printf("[Payment SAGA] Received ShipmentFailed event - initiating refund")

	var envelope kafka.DomainEvent
	if err := msg.Unmarshal(&envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var shipmentFailed saga.ShipmentFailedEvent
	if err := json.Unmarshal(payloadBytes, &shipmentFailed); err != nil {
		return fmt.Errorf("failed to unmarshal shipment failed event: %w", err)
	}

	// Find payment by order ID
	payment, err := h.repo.FindByOrderID(ctx, shipmentFailed.OrderID)
	if err != nil {
		return fmt.Errorf("failed to find payment for order %s: %w", shipmentFailed.OrderID, err)
	}

	// Refund through gateway
	if err := h.gateway.RefundPayment(ctx, payment.TransactionID, payment.Amount); err != nil {
		log.Printf("[Payment SAGA] Failed to refund payment: %v", err)
		return err
	}

	// Mark payment as refunded
	if err := payment.Refund("Shipment failed: " + shipmentFailed.Reason); err != nil {
		return err
	}

	if err := h.repo.Save(ctx, payment); err != nil {
		return err
	}

	// Publish refund event
	return h.publishPaymentRefunded(ctx, payment, shipmentFailed.Reason)
}

func (h *EventHandler) publishPaymentProcessed(ctx context.Context, payment *aggregate.Payment, transactionID, customerID string) error {
	event := saga.PaymentProcessedEvent{
		PaymentID:     payment.ID.String(),
		OrderID:       payment.OrderID,
		CustomerID:    customerID,
		Amount:        saga.Money{Amount: payment.Amount.Amount, Currency: string(payment.Amount.Currency)},
		TransactionID: transactionID,
		OccurredAt:    time.Now().UTC(),
	}

	log.Printf("[Payment SAGA] Publishing PaymentProcessed for order %s", payment.OrderID)
	return h.eventPublisher.Publish(ctx, []interface{}{&paymentProcessedWrapper{event}})
}

func (h *EventHandler) publishPaymentFailed(ctx context.Context, orderID, customerID string, amount int64, currency, reason string) error {
	event := saga.PaymentFailedEvent{
		PaymentID:  uuid.New().String(),
		OrderID:    orderID,
		CustomerID: customerID,
		Amount:     saga.Money{Amount: amount, Currency: currency},
		Reason:     reason,
		OccurredAt: time.Now().UTC(),
	}

	log.Printf("[Payment SAGA] Publishing PaymentFailed for order %s: %s", orderID, reason)
	return h.eventPublisher.Publish(ctx, []interface{}{&paymentFailedWrapper{event}})
}

func (h *EventHandler) publishPaymentRefunded(ctx context.Context, payment *aggregate.Payment, reason string) error {
	event := saga.PaymentRefundedEvent{
		PaymentID:     payment.ID.String(),
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        saga.Money{Amount: payment.Amount.Amount, Currency: string(payment.Amount.Currency)},
		Reason:        reason,
		OccurredAt:    time.Now().UTC(),
	}

	log.Printf("[Payment SAGA] Publishing PaymentRefunded for order %s", payment.OrderID)
	return h.eventPublisher.Publish(ctx, []interface{}{&paymentRefundedWrapper{event}})
}

// Event wrappers to implement the event interface
type paymentProcessedWrapper struct {
	saga.PaymentProcessedEvent
}

func (e *paymentProcessedWrapper) EventType() string     { return saga.EventPaymentProcessed }
func (e *paymentProcessedWrapper) AggregateID() string   { return e.OrderID }
func (e *paymentProcessedWrapper) AggregateType() string { return "Payment" }

type paymentFailedWrapper struct {
	saga.PaymentFailedEvent
}

func (e *paymentFailedWrapper) EventType() string     { return saga.EventPaymentFailed }
func (e *paymentFailedWrapper) AggregateID() string   { return e.OrderID }
func (e *paymentFailedWrapper) AggregateType() string { return "Payment" }

type paymentRefundedWrapper struct {
	saga.PaymentRefundedEvent
}

func (e *paymentRefundedWrapper) EventType() string     { return saga.EventPaymentRefunded }
func (e *paymentRefundedWrapper) AggregateID() string   { return e.OrderID }
func (e *paymentRefundedWrapper) AggregateType() string { return "Payment" }
