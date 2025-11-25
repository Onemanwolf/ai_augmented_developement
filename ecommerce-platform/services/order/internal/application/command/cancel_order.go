// Package command contains command definitions for the Order service.
package command

// CancelOrder represents a command to cancel an order.
type CancelOrder struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// Validate validates the command.
func (c *CancelOrder) Validate() error {
	if c.OrderID == "" {
		return ErrOrderIDRequired
	}
	if c.Reason == "" {
		return ErrReasonRequired
	}
	return nil
}
