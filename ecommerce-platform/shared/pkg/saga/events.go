// Package saga provides shared event types for the order processing saga.
package saga

import "time"

// Event type constants
const (
	// Order events
	EventOrderCreated   = "OrderCreated"
	EventOrderPaid      = "OrderPaid"
	EventOrderShipped   = "OrderShipped"
	EventOrderCompleted = "OrderCompleted"
	EventOrderCancelled = "OrderCancelled"
	EventOrderFailed    = "OrderFailed"

	// Payment events
	EventPaymentProcessed = "PaymentProcessed"
	EventPaymentFailed    = "PaymentFailed"
	EventPaymentRefunded  = "PaymentRefunded"

	// Fulfillment events
	EventShipmentCreated   = "ShipmentCreated"
	EventShipmentFailed    = "ShipmentFailed"
	EventShipmentShipped   = "ShipmentShipped"
	EventShipmentDelivered = "ShipmentDelivered"
)

// Topic constants
const (
	TopicOrderEvents       = "order-events"
	TopicPaymentEvents     = "payment-events"
	TopicFulfillmentEvents = "fulfillment-events"
)

// OrderCreatedEvent is published when an order is created.
type OrderCreatedEvent struct {
	OrderID     string      `json:"order_id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount Money       `json:"total_amount"`
	OccurredAt  time.Time   `json:"occurred_at"`
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice Money  `json:"unit_price"`
}

// Money represents a monetary amount.
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// PaymentProcessedEvent is published when payment succeeds.
type PaymentProcessedEvent struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	CustomerID    string    `json:"customer_id"`
	Amount        Money     `json:"amount"`
	TransactionID string    `json:"transaction_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// PaymentFailedEvent is published when payment fails.
type PaymentFailedEvent struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Amount     Money     `json:"amount"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentRefundedEvent is published when a payment is refunded.
type PaymentRefundedEvent struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	TransactionID string    `json:"transaction_id"`
	Amount        Money     `json:"amount"`
	Reason        string    `json:"reason"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// ShipmentCreatedEvent is published when a shipment is created.
type ShipmentCreatedEvent struct {
	ShipmentID string      `json:"shipment_id"`
	OrderID    string      `json:"order_id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	OccurredAt time.Time   `json:"occurred_at"`
}

// ShipmentFailedEvent is published when shipment creation fails.
type ShipmentFailedEvent struct {
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ShipmentShippedEvent is published when a shipment is shipped.
type ShipmentShippedEvent struct {
	ShipmentID     string    `json:"shipment_id"`
	OrderID        string    `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	Carrier        string    `json:"carrier"`
	OccurredAt     time.Time `json:"occurred_at"`
}

// ShipmentDeliveredEvent is published when a shipment is delivered.
type ShipmentDeliveredEvent struct {
	ShipmentID  string    `json:"shipment_id"`
	OrderID     string    `json:"order_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// OrderStatusUpdatedEvent is published when order status changes.
type OrderStatusUpdatedEvent struct {
	OrderID    string    `json:"order_id"`
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	Reason     string    `json:"reason,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}
