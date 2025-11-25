package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentCancelled is emitted when a payment is cancelled.
type PaymentCancelled struct {
	BaseEvent
	OrderID string `json:"order_id"`
}

// NewPaymentCancelled creates a new PaymentCancelled event.
func NewPaymentCancelled(
	paymentID valueobject.PaymentID,
	orderID string,
) *PaymentCancelled {
	return &PaymentCancelled{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "PaymentCancelled",
			AggregateIDV: paymentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID: orderID,
	}
}
