// Package handler contains command and query handlers for the Payment service.
package handler

import (
	"context"
	"fmt"

	"github.com/your-org/ecommerce-platform/services/payment/internal/application/command"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentCommandHandler handles payment commands.
type PaymentCommandHandler struct {
	repo           repository.PaymentRepository
	eventPublisher EventPublisher
	gateway        PaymentGateway
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, events []interface{}) error
}

// PaymentGateway defines the interface for payment processing.
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, amount valueobject.Money, method valueobject.PaymentMethod) (transactionID string, err error)
	RefundPayment(ctx context.Context, transactionID string, amount valueobject.Money) error
}

// NewPaymentCommandHandler creates a new PaymentCommandHandler.
func NewPaymentCommandHandler(
	repo repository.PaymentRepository,
	publisher EventPublisher,
	gateway PaymentGateway,
) *PaymentCommandHandler {
	return &PaymentCommandHandler{
		repo:           repo,
		eventPublisher: publisher,
		gateway:        gateway,
	}
}

// HandleProcessPayment handles the ProcessPayment command.
func (h *PaymentCommandHandler) HandleProcessPayment(ctx context.Context, cmd *command.ProcessPayment) (*aggregate.Payment, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create money value object
	amount, err := valueobject.NewMoney(cmd.Amount, cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Create payment aggregate
	payment, err := aggregate.NewPayment(cmd.OrderID, cmd.CustomerID, amount, cmd.Method)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Start processing
	if err := payment.Process(); err != nil {
		return nil, fmt.Errorf("failed to start processing: %w", err)
	}

	// Save initial state
	if err := h.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Process with payment gateway
	transactionID, err := h.gateway.ProcessPayment(ctx, amount, cmd.Method)
	if err != nil {
		// Mark as failed
		if failErr := payment.Fail(err.Error()); failErr != nil {
			fmt.Printf("failed to mark payment as failed: %v\n", failErr)
		}
		if saveErr := h.repo.Save(ctx, payment); saveErr != nil {
			fmt.Printf("failed to save failed payment: %v\n", saveErr)
		}
		h.publishEvents(ctx, payment)
		return payment, fmt.Errorf("payment processing failed: %w", err)
	}

	// Mark as completed
	if err := payment.Complete(transactionID); err != nil {
		return nil, fmt.Errorf("failed to complete payment: %w", err)
	}

	// Save final state
	if err := h.repo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save completed payment: %w", err)
	}

	h.publishEvents(ctx, payment)
	return payment, nil
}

// HandleRefundPayment handles the RefundPayment command.
func (h *PaymentCommandHandler) HandleRefundPayment(ctx context.Context, cmd *command.RefundPayment) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	paymentID, err := valueobject.ParsePaymentID(cmd.PaymentID)
	if err != nil {
		return fmt.Errorf("invalid payment ID: %w", err)
	}

	payment, err := h.repo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}

	// Determine refund amount
	var refundAmount valueobject.Money
	if cmd.IsFullRefund() {
		remainingAmount, err := payment.GetRemainingAmount()
		if err != nil {
			return fmt.Errorf("failed to get remaining amount: %w", err)
		}
		refundAmount = remainingAmount
	} else {
		refundAmount, err = valueobject.NewMoney(cmd.Amount, payment.Amount.Currency)
		if err != nil {
			return fmt.Errorf("invalid refund amount: %w", err)
		}
	}

	// Process refund with gateway
	if err := h.gateway.RefundPayment(ctx, payment.TransactionID, refundAmount); err != nil {
		return fmt.Errorf("refund processing failed: %w", err)
	}

	// Update payment state
	if cmd.IsFullRefund() || refundAmount.Equals(payment.Amount) {
		if err := payment.Refund(cmd.Reason); err != nil {
			return fmt.Errorf("failed to refund payment: %w", err)
		}
	} else {
		if err := payment.PartialRefund(refundAmount, cmd.Reason); err != nil {
			return fmt.Errorf("failed to partially refund payment: %w", err)
		}
	}

	if err := h.repo.Save(ctx, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	h.publishEvents(ctx, payment)
	return nil
}

// HandleCancelPayment handles the CancelPayment command.
func (h *PaymentCommandHandler) HandleCancelPayment(ctx context.Context, cmd *command.CancelPayment) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	paymentID, err := valueobject.ParsePaymentID(cmd.PaymentID)
	if err != nil {
		return fmt.Errorf("invalid payment ID: %w", err)
	}

	payment, err := h.repo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}

	if err := payment.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}

	if err := h.repo.Save(ctx, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	h.publishEvents(ctx, payment)
	return nil
}

func (h *PaymentCommandHandler) publishEvents(ctx context.Context, payment *aggregate.Payment) {
	events := payment.GetEvents()
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
