// Package query contains query definitions for the Payment service.
package query

// GetPayment represents a query to get a payment by ID.
type GetPayment struct {
	PaymentID string `json:"payment_id"`
}

// Validate validates the query.
func (q *GetPayment) Validate() error {
	if q.PaymentID == "" {
		return ErrPaymentIDRequired
	}
	return nil
}

// GetPaymentByOrder represents a query to get a payment by order ID.
type GetPaymentByOrder struct {
	OrderID string `json:"order_id"`
}

// Validate validates the query.
func (q *GetPaymentByOrder) Validate() error {
	if q.OrderID == "" {
		return ErrOrderIDRequired
	}
	return nil
}

// GetPaymentsByCustomer represents a query to get payments by customer ID.
type GetPaymentsByCustomer struct {
	CustomerID string `json:"customer_id"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// Validate validates the query.
func (q *GetPaymentsByCustomer) Validate() error {
	if q.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	return nil
}
