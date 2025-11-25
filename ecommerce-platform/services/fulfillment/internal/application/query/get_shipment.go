// Package query contains query definitions for the Fulfillment service.
package query

// GetShipment represents a query to get a shipment by ID.
type GetShipment struct {
	ShipmentID string `json:"shipment_id"`
}

// Validate validates the query.
func (q *GetShipment) Validate() error {
	if q.ShipmentID == "" {
		return ErrShipmentIDRequired
	}
	return nil
}

// GetShipmentByOrder represents a query to get a shipment by order ID.
type GetShipmentByOrder struct {
	OrderID string `json:"order_id"`
}

// Validate validates the query.
func (q *GetShipmentByOrder) Validate() error {
	if q.OrderID == "" {
		return ErrOrderIDRequired
	}
	return nil
}

// GetShipmentByTracking represents a query to get a shipment by tracking number.
type GetShipmentByTracking struct {
	TrackingNumber string `json:"tracking_number"`
}

// Validate validates the query.
func (q *GetShipmentByTracking) Validate() error {
	if q.TrackingNumber == "" {
		return ErrTrackingNumberRequired
	}
	return nil
}

// GetShipmentsByCustomer represents a query to get shipments by customer ID.
type GetShipmentsByCustomer struct {
	CustomerID string `json:"customer_id"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// Validate validates the query.
func (q *GetShipmentsByCustomer) Validate() error {
	if q.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	return nil
}
