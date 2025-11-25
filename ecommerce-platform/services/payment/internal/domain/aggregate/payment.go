// Package aggregate contains aggregate roots for the Payment domain.
package aggregate

import (
	"fmt"
	"time"

	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/event"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// DomainEvent represents a domain event interface.
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// Payment is the aggregate root for payments.
type Payment struct {
	ID              valueobject.PaymentID     `bson:"_id"`
	OrderID         string                    `bson:"order_id"`
	CustomerID      string                    `bson:"customer_id"`
	Amount          valueobject.Money         `bson:"amount"`
	Method          valueobject.PaymentMethod `bson:"method"`
	Status          valueobject.PaymentStatus `bson:"status"`
	TransactionID   string                    `bson:"transaction_id,omitempty"`
	FailureReason   string                    `bson:"failure_reason,omitempty"`
	RefundedAmount  valueobject.Money         `bson:"refunded_amount"`
	CreatedAt       time.Time                 `bson:"created_at"`
	UpdatedAt       time.Time                 `bson:"updated_at"`
	ProcessedAt     *time.Time                `bson:"processed_at,omitempty"`
	Version         int                       `bson:"version"`

	events []DomainEvent
}

// NewPayment creates a new Payment aggregate.
func NewPayment(
	orderID string,
	customerID string,
	amount valueobject.Money,
	method valueobject.PaymentMethod,
) (*Payment, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}
	if !amount.IsPositive() {
		return nil, fmt.Errorf("payment amount must be positive")
	}
	if !method.IsValid() {
		return nil, fmt.Errorf("invalid payment method")
	}

	now := time.Now().UTC()
	payment := &Payment{
		ID:             valueobject.NewPaymentID(),
		OrderID:        orderID,
		CustomerID:     customerID,
		Amount:         amount,
		Method:         method,
		Status:         valueobject.PaymentStatusPending,
		RefundedAmount: valueobject.Zero(amount.Currency),
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        1,
		events:         make([]DomainEvent, 0),
	}

	payment.addEvent(event.NewPaymentCreated(
		payment.ID,
		orderID,
		customerID,
		amount,
		method,
	))

	return payment, nil
}

// Process starts processing the payment.
func (p *Payment) Process() error {
	if !p.Status.CanTransitionTo(valueobject.PaymentStatusProcessing) {
		return fmt.Errorf("cannot process payment in status %s", p.Status)
	}

	oldStatus := p.Status
	p.Status = valueobject.PaymentStatusProcessing
	p.UpdatedAt = time.Now().UTC()
	p.Version++

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	return nil
}

// Complete marks the payment as completed.
func (p *Payment) Complete(transactionID string) error {
	if !p.Status.CanTransitionTo(valueobject.PaymentStatusCompleted) {
		return fmt.Errorf("cannot complete payment in status %s", p.Status)
	}
	if transactionID == "" {
		return fmt.Errorf("transaction ID is required")
	}

	now := time.Now().UTC()
	oldStatus := p.Status
	p.Status = valueobject.PaymentStatusCompleted
	p.TransactionID = transactionID
	p.ProcessedAt = &now
	p.UpdatedAt = now
	p.Version++

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	p.addEvent(event.NewPaymentReceived(p.ID, p.OrderID, p.Amount, transactionID))
	return nil
}

// Fail marks the payment as failed.
func (p *Payment) Fail(reason string) error {
	if !p.Status.CanTransitionTo(valueobject.PaymentStatusFailed) {
		return fmt.Errorf("cannot fail payment in status %s", p.Status)
	}

	now := time.Now().UTC()
	oldStatus := p.Status
	p.Status = valueobject.PaymentStatusFailed
	p.FailureReason = reason
	p.ProcessedAt = &now
	p.UpdatedAt = now
	p.Version++

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	p.addEvent(event.NewPaymentFailed(p.ID, p.OrderID, reason))
	return nil
}

// Refund processes a full refund.
func (p *Payment) Refund(reason string) error {
	if !p.Status.CanTransitionTo(valueobject.PaymentStatusRefunded) {
		return fmt.Errorf("cannot refund payment in status %s", p.Status)
	}

	oldStatus := p.Status
	p.Status = valueobject.PaymentStatusRefunded
	p.RefundedAmount = p.Amount
	p.UpdatedAt = time.Now().UTC()
	p.Version++

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	p.addEvent(event.NewRefundProcessed(p.ID, p.OrderID, p.Amount, reason))
	return nil
}

// PartialRefund processes a partial refund.
func (p *Payment) PartialRefund(amount valueobject.Money, reason string) error {
	if p.Status != valueobject.PaymentStatusCompleted && p.Status != valueobject.PaymentStatusPartiallyRefunded {
		return fmt.Errorf("cannot partially refund payment in status %s", p.Status)
	}
	if amount.Currency != p.Amount.Currency {
		return fmt.Errorf("refund currency must match payment currency")
	}

	remainingAmount, err := p.Amount.Subtract(p.RefundedAmount)
	if err != nil {
		return fmt.Errorf("failed to calculate remaining amount: %w", err)
	}

	if amount.Amount > remainingAmount.Amount {
		return fmt.Errorf("refund amount exceeds remaining balance")
	}

	oldStatus := p.Status
	newRefunded, _ := p.RefundedAmount.Add(amount)
	p.RefundedAmount = newRefunded
	p.UpdatedAt = time.Now().UTC()
	p.Version++

	if p.RefundedAmount.Equals(p.Amount) {
		p.Status = valueobject.PaymentStatusRefunded
	} else {
		p.Status = valueobject.PaymentStatusPartiallyRefunded
	}

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	p.addEvent(event.NewRefundProcessed(p.ID, p.OrderID, amount, reason))
	return nil
}

// Cancel cancels a pending payment.
func (p *Payment) Cancel() error {
	if !p.Status.CanTransitionTo(valueobject.PaymentStatusCancelled) {
		return fmt.Errorf("cannot cancel payment in status %s", p.Status)
	}

	oldStatus := p.Status
	p.Status = valueobject.PaymentStatusCancelled
	p.UpdatedAt = time.Now().UTC()
	p.Version++

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	p.addEvent(event.NewPaymentCancelled(p.ID, p.OrderID))
	return nil
}

// Retry retries a failed payment.
func (p *Payment) Retry() error {
	if !p.Status.CanTransitionTo(valueobject.PaymentStatusPending) {
		return fmt.Errorf("cannot retry payment in status %s", p.Status)
	}

	oldStatus := p.Status
	p.Status = valueobject.PaymentStatusPending
	p.FailureReason = ""
	p.UpdatedAt = time.Now().UTC()
	p.Version++

	p.addEvent(event.NewPaymentStatusChanged(p.ID, oldStatus, p.Status))
	return nil
}

// GetEvents returns and clears domain events.
func (p *Payment) GetEvents() []DomainEvent {
	events := p.events
	p.events = make([]DomainEvent, 0)
	return events
}

// ClearEvents clears all domain events.
func (p *Payment) ClearEvents() {
	p.events = make([]DomainEvent, 0)
}

func (p *Payment) addEvent(e DomainEvent) {
	p.events = append(p.events, e)
}

// IsPending checks if payment is pending.
func (p *Payment) IsPending() bool {
	return p.Status == valueobject.PaymentStatusPending
}

// IsCompleted checks if payment is completed.
func (p *Payment) IsCompleted() bool {
	return p.Status == valueobject.PaymentStatusCompleted
}

// IsFailed checks if payment failed.
func (p *Payment) IsFailed() bool {
	return p.Status == valueobject.PaymentStatusFailed
}

// GetRemainingAmount returns the amount that can still be refunded.
func (p *Payment) GetRemainingAmount() (valueobject.Money, error) {
	return p.Amount.Subtract(p.RefundedAmount)
}
