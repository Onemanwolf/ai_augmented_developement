// Package query contains query definitions for the Payment service.
package query

import (
	"time"

	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/aggregate"
)

// PaymentDTO represents a payment in query responses.
type PaymentDTO struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	CustomerID     string    `json:"customer_id"`
	Amount         MoneyDTO  `json:"amount"`
	Method         string    `json:"method"`
	Status         string    `json:"status"`
	TransactionID  string    `json:"transaction_id,omitempty"`
	FailureReason  string    `json:"failure_reason,omitempty"`
	RefundedAmount MoneyDTO  `json:"refunded_amount"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
}

// MoneyDTO represents money in query responses.
type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// FromAggregate converts a Payment aggregate to PaymentDTO.
func FromAggregate(payment *aggregate.Payment) *PaymentDTO {
	if payment == nil {
		return nil
	}

	return &PaymentDTO{
		ID:         payment.ID.String(),
		OrderID:    payment.OrderID,
		CustomerID: payment.CustomerID,
		Amount: MoneyDTO{
			Amount:   payment.Amount.Amount,
			Currency: string(payment.Amount.Currency),
		},
		Method:        string(payment.Method),
		Status:        string(payment.Status),
		TransactionID: payment.TransactionID,
		FailureReason: payment.FailureReason,
		RefundedAmount: MoneyDTO{
			Amount:   payment.RefundedAmount.Amount,
			Currency: string(payment.RefundedAmount.Currency),
		},
		CreatedAt:   payment.CreatedAt,
		UpdatedAt:   payment.UpdatedAt,
		ProcessedAt: payment.ProcessedAt,
	}
}

// FromAggregateList converts a list of Payment aggregates to PaymentDTOs.
func FromAggregateList(payments []*aggregate.Payment) []*PaymentDTO {
	dtos := make([]*PaymentDTO, len(payments))
	for i, payment := range payments {
		dtos[i] = FromAggregate(payment)
	}
	return dtos
}
