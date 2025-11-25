package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentCreated is emitted when a new payment is created.
type PaymentCreated struct {
	BaseEvent
	OrderID    string                    `json:"order_id"`
	CustomerID string                    `json:"customer_id"`
	Amount     valueobject.Money         `json:"amount"`
	Method     valueobject.PaymentMethod `json:"method"`
}

// NewPaymentCreated creates a new PaymentCreated event.
func NewPaymentCreated(
	paymentID valueobject.PaymentID,
	orderID string,
	customerID string,
	amount valueobject.Money,
	method valueobject.PaymentMethod,
) *PaymentCreated {
	return &PaymentCreated{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "PaymentCreated",
			AggregateIDV: paymentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID:    orderID,
		CustomerID: customerID,
		Amount:     amount,
		Method:     method,
	}
}
