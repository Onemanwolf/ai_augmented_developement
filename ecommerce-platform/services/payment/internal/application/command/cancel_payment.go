// Package command contains command definitions for the Payment service.
package command

// CancelPayment represents a command to cancel a pending payment.
type CancelPayment struct {
	PaymentID string `json:"payment_id"`
}

// Validate validates the command.
func (c *CancelPayment) Validate() error {
	if c.PaymentID == "" {
		return ErrPaymentIDRequired
	}
	return nil
}
