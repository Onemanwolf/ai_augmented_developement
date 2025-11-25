// Package logging provides structured logging utilities using zap.
package logging

import "go.uber.org/zap"

// Common field constructors for structured logging.

// OrderID creates a field for order ID.
func OrderID(id string) zap.Field {
	return zap.String("order_id", id)
}

// PaymentID creates a field for payment ID.
func PaymentID(id string) zap.Field {
	return zap.String("payment_id", id)
}

// ShipmentID creates a field for shipment ID.
func ShipmentID(id string) zap.Field {
	return zap.String("shipment_id", id)
}

// CustomerID creates a field for customer ID.
func CustomerID(id string) zap.Field {
	return zap.String("customer_id", id)
}

// Amount creates a field for monetary amount.
func Amount(cents int64) zap.Field {
	return zap.Int64("amount_cents", cents)
}

// Currency creates a field for currency code.
func Currency(code string) zap.Field {
	return zap.String("currency", code)
}

// Status creates a field for status.
func Status(status string) zap.Field {
	return zap.String("status", status)
}

// EventType creates a field for event type.
func EventType(eventType string) zap.Field {
	return zap.String("event_type", eventType)
}

// Topic creates a field for Kafka topic.
func Topic(topic string) zap.Field {
	return zap.String("topic", topic)
}

// Partition creates a field for Kafka partition.
func Partition(partition int) zap.Field {
	return zap.Int("partition", partition)
}

// Offset creates a field for Kafka offset.
func Offset(offset int64) zap.Field {
	return zap.Int64("offset", offset)
}

// Collection creates a field for MongoDB collection.
func Collection(name string) zap.Field {
	return zap.String("collection", name)
}

// Operation creates a field for operation type.
func Operation(op string) zap.Field {
	return zap.String("operation", op)
}

// Count creates a field for count/quantity.
func Count(n int) zap.Field {
	return zap.Int("count", n)
}

// Carrier creates a field for shipping carrier.
func Carrier(name string) zap.Field {
	return zap.String("carrier", name)
}

// TrackingNumber creates a field for tracking number.
func TrackingNumber(number string) zap.Field {
	return zap.String("tracking_number", number)
}

// PaymentMethod creates a field for payment method.
func PaymentMethod(method string) zap.Field {
	return zap.String("payment_method", method)
}

// TransactionID creates a field for transaction ID.
func TransactionID(id string) zap.Field {
	return zap.String("transaction_id", id)
}

// UserAgent creates a field for user agent.
func UserAgent(ua string) zap.Field {
	return zap.String("user_agent", ua)
}

// RemoteAddr creates a field for remote address.
func RemoteAddr(addr string) zap.Field {
	return zap.String("remote_addr", addr)
}
