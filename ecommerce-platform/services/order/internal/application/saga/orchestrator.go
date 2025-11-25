// Package saga contains the SAGA orchestrator for order processing.
package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OrderSagaOrchestrator orchestrates the order processing saga.
type OrderSagaOrchestrator struct {
	sagaRepo       SagaRepository
	paymentService PaymentService
	fulfillmentService FulfillmentService
	orderService   OrderService
	eventPublisher EventPublisher
}

// PaymentService defines the interface for payment operations.
type PaymentService interface {
	ProcessPayment(ctx context.Context, req ProcessPaymentRequest) (paymentID string, err error)
	RefundPayment(ctx context.Context, paymentID string, reason string) error
}

// FulfillmentService defines the interface for fulfillment operations.
type FulfillmentService interface {
	CreateShipment(ctx context.Context, req CreateShipmentRequest) (shipmentID string, err error)
	CancelShipment(ctx context.Context, shipmentID string) error
}

// OrderService defines the interface for order operations.
type OrderService interface {
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
	CancelOrder(ctx context.Context, orderID string, reason string) error
}

// EventPublisher defines the interface for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, events []interface{}) error
}

// ProcessPaymentRequest represents a payment request.
type ProcessPaymentRequest struct {
	OrderID    string
	CustomerID string
	Amount     int64
	Currency   string
	Method     string
}

// CreateShipmentRequest represents a shipment creation request.
type CreateShipmentRequest struct {
	OrderID    string
	CustomerID string
	Items      []ShipmentItem
	Address    ShippingAddress
	Carrier    string
}

// ShipmentItem represents an item for shipment.
type ShipmentItem struct {
	ProductID   string
	Name        string
	Quantity    int
	WeightGrams int
}

// ShippingAddress represents a shipping address.
type ShippingAddress struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

// NewOrderSagaOrchestrator creates a new OrderSagaOrchestrator.
func NewOrderSagaOrchestrator(
	sagaRepo SagaRepository,
	paymentService PaymentService,
	fulfillmentService FulfillmentService,
	orderService OrderService,
	eventPublisher EventPublisher,
) *OrderSagaOrchestrator {
	return &OrderSagaOrchestrator{
		sagaRepo:           sagaRepo,
		paymentService:     paymentService,
		fulfillmentService: fulfillmentService,
		orderService:       orderService,
		eventPublisher:     eventPublisher,
	}
}

// StartSaga starts a new order saga.
func (o *OrderSagaOrchestrator) StartSaga(ctx context.Context, orderID, customerID string) (*OrderSagaData, error) {
	sagaID := uuid.New().String()
	saga := NewOrderSagaData(sagaID, orderID, customerID)

	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return nil, fmt.Errorf("failed to save saga: %w", err)
	}

	return saga, nil
}

// ProcessPaymentStep executes the payment step.
func (o *OrderSagaOrchestrator) ProcessPaymentStep(ctx context.Context, sagaID string, req ProcessPaymentRequest) error {
	saga, err := o.sagaRepo.FindByID(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("failed to find saga: %w", err)
	}

	if saga.State != SagaStateStarted {
		return fmt.Errorf("invalid saga state for payment: %s", saga.State)
	}

	saga.TransitionTo(SagaStatePaymentPending)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	paymentID, err := o.paymentService.ProcessPayment(ctx, req)
	if err != nil {
		saga.TransitionTo(SagaStatePaymentFailed)
		saga.SetFailure(err.Error())
		if saveErr := o.sagaRepo.Save(ctx, saga); saveErr != nil {
			fmt.Printf("failed to save saga after payment failure: %v\n", saveErr)
		}
		return fmt.Errorf("payment failed: %w", err)
	}

	saga.SetPaymentID(paymentID)
	saga.TransitionTo(SagaStatePaymentCompleted)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	// Update order status
	if err := o.orderService.UpdateOrderStatus(ctx, saga.OrderID, "PAID"); err != nil {
		fmt.Printf("failed to update order status: %v\n", err)
	}

	return nil
}

// ProcessFulfillmentStep executes the fulfillment step.
func (o *OrderSagaOrchestrator) ProcessFulfillmentStep(ctx context.Context, sagaID string, req CreateShipmentRequest) error {
	saga, err := o.sagaRepo.FindByID(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("failed to find saga: %w", err)
	}

	if saga.State != SagaStatePaymentCompleted {
		return fmt.Errorf("invalid saga state for fulfillment: %s", saga.State)
	}

	saga.TransitionTo(SagaStateFulfillmentPending)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	shipmentID, err := o.fulfillmentService.CreateShipment(ctx, req)
	if err != nil {
		saga.TransitionTo(SagaStateFulfillmentFailed)
		saga.SetFailure(err.Error())
		if saveErr := o.sagaRepo.Save(ctx, saga); saveErr != nil {
			fmt.Printf("failed to save saga after fulfillment failure: %v\n", saveErr)
		}
		return fmt.Errorf("fulfillment failed: %w", err)
	}

	saga.SetShipmentID(shipmentID)
	saga.TransitionTo(SagaStateFulfillmentCompleted)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	return nil
}

// CompleteSaga marks the saga as completed.
func (o *OrderSagaOrchestrator) CompleteSaga(ctx context.Context, sagaID string) error {
	saga, err := o.sagaRepo.FindByID(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("failed to find saga: %w", err)
	}

	if saga.State != SagaStateFulfillmentCompleted {
		return fmt.Errorf("invalid saga state for completion: %s", saga.State)
	}

	saga.TransitionTo(SagaStateCompleted)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	return nil
}

// Compensate runs compensation for a failed saga.
func (o *OrderSagaOrchestrator) Compensate(ctx context.Context, sagaID string) error {
	saga, err := o.sagaRepo.FindByID(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("failed to find saga: %w", err)
	}

	if !saga.NeedsCompensation() {
		return fmt.Errorf("saga does not need compensation: %s", saga.State)
	}

	saga.TransitionTo(SagaStateCompensating)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	// Compensate fulfillment if shipment was created
	if saga.ShipmentID != "" {
		if err := o.fulfillmentService.CancelShipment(ctx, saga.ShipmentID); err != nil {
			saga.AddCompensationLog("cancel_shipment", "failed", err.Error())
			fmt.Printf("failed to cancel shipment: %v\n", err)
		} else {
			saga.AddCompensationLog("cancel_shipment", "success", "")
		}
	}

	// Compensate payment if payment was made
	if saga.PaymentID != "" {
		if err := o.paymentService.RefundPayment(ctx, saga.PaymentID, "Order saga compensation"); err != nil {
			saga.AddCompensationLog("refund_payment", "failed", err.Error())
			fmt.Printf("failed to refund payment: %v\n", err)
		} else {
			saga.AddCompensationLog("refund_payment", "success", "")
		}
	}

	// Cancel the order
	if err := o.orderService.CancelOrder(ctx, saga.OrderID, "Order processing failed: "+saga.FailureReason); err != nil {
		saga.AddCompensationLog("cancel_order", "failed", err.Error())
		fmt.Printf("failed to cancel order: %v\n", err)
	} else {
		saga.AddCompensationLog("cancel_order", "success", "")
	}

	saga.TransitionTo(SagaStateCompensated)
	if err := o.sagaRepo.Save(ctx, saga); err != nil {
		return fmt.Errorf("failed to save saga: %w", err)
	}

	return nil
}

// GetSagaStatus retrieves the current status of a saga.
func (o *OrderSagaOrchestrator) GetSagaStatus(ctx context.Context, sagaID string) (*OrderSagaData, error) {
	return o.sagaRepo.FindByID(ctx, sagaID)
}

// GetSagaByOrderID retrieves a saga by order ID.
func (o *OrderSagaOrchestrator) GetSagaByOrderID(ctx context.Context, orderID string) (*OrderSagaData, error) {
	return o.sagaRepo.FindByOrderID(ctx, orderID)
}

// ProcessPendingSagas processes sagas that need attention.
func (o *OrderSagaOrchestrator) ProcessPendingSagas(ctx context.Context) error {
	// Find sagas that need compensation
	sagas, err := o.sagaRepo.FindPending(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to find pending sagas: %w", err)
	}

	for _, saga := range sagas {
		if saga.NeedsCompensation() {
			if err := o.Compensate(ctx, saga.ID); err != nil {
				fmt.Printf("failed to compensate saga %s: %v\n", saga.ID, err)
			}
		}
	}

	return nil
}

// RecoverStaleSagas recovers sagas that may be stuck.
func (o *OrderSagaOrchestrator) RecoverStaleSagas(ctx context.Context) error {
	// Find sagas older than 5 minutes that aren't in terminal state
	staleThreshold := time.Now().Add(-5 * time.Minute).Unix()
	sagas, err := o.sagaRepo.FindStale(ctx, staleThreshold, 100)
	if err != nil {
		return fmt.Errorf("failed to find stale sagas: %w", err)
	}

	for _, saga := range sagas {
		if !saga.IsTerminal() {
			// Mark as failed and compensate
			saga.SetFailure("Saga timed out")
			saga.TransitionTo(SagaStateFailed)
			if err := o.sagaRepo.Save(ctx, saga); err != nil {
				fmt.Printf("failed to save stale saga %s: %v\n", saga.ID, err)
				continue
			}
			if err := o.Compensate(ctx, saga.ID); err != nil {
				fmt.Printf("failed to compensate stale saga %s: %v\n", saga.ID, err)
			}
		}
	}

	return nil
}
