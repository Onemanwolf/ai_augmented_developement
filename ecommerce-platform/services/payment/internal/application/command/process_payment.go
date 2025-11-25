// Package command contains command definitions for the Payment service.
package command

import "github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"

// ProcessPayment represents a command to process a payment.
type ProcessPayment struct {
	OrderID    string                    `json:"order_id"`
	CustomerID string                    `json:"customer_id"`
	Amount     int64                     `json:"amount"`
	Currency   valueobject.Currency      `json:"currency"`
	Method     valueobject.PaymentMethod `json:"method"`
}

// Validate validates the command.
func (c *ProcessPayment) Validate() error {
	if c.OrderID == "" {
		return ErrOrderIDRequired
	}
	if c.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	if c.Amount <= 0 {
		return ErrInvalidAmount
	}
	if !c.Currency.IsValid() {
		return ErrInvalidCurrency
	}
	if !c.Method.IsValid() {
		return ErrInvalidPaymentMethod
	}
	return nil
}
