// Package command contains command definitions for the Payment service.
package command

// RefundPayment represents a command to refund a payment.
type RefundPayment struct {
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount,omitempty"` // 0 means full refund
	Reason    string `json:"reason"`
}

// Validate validates the command.
func (c *RefundPayment) Validate() error {
	if c.PaymentID == "" {
		return ErrPaymentIDRequired
	}
	if c.Reason == "" {
		return ErrReasonRequired
	}
	if c.Amount < 0 {
		return ErrInvalidAmount
	}
	return nil
}

// IsFullRefund returns true if this is a full refund request.
func (c *RefundPayment) IsFullRefund() bool {
	return c.Amount == 0
}
