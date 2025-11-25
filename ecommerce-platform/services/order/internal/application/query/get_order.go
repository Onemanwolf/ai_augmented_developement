// Package query contains query definitions for the Order service.
package query

// GetOrder represents a query to get an order by ID.
type GetOrder struct {
	OrderID string `json:"order_id"`
}

// Validate validates the query.
func (q *GetOrder) Validate() error {
	if q.OrderID == "" {
		return ErrOrderIDRequired
	}
	return nil
}

// GetOrdersByCustomer represents a query to get orders by customer ID.
type GetOrdersByCustomer struct {
	CustomerID string `json:"customer_id"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// Validate validates the query.
func (q *GetOrdersByCustomer) Validate() error {
	if q.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	return nil
}

// GetOrdersByStatus represents a query to get orders by status.
type GetOrdersByStatus struct {
	Status string `json:"status"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// Validate validates the query.
func (q *GetOrdersByStatus) Validate() error {
	if q.Status == "" {
		return ErrStatusRequired
	}
	return nil
}
