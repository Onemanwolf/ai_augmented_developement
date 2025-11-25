package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentFailed is emitted when a payment fails.
type PaymentFailed struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// NewPaymentFailed creates a new PaymentFailed event.
func NewPaymentFailed(
	paymentID valueobject.PaymentID,
	orderID string,
	reason string,
) *PaymentFailed {
	return &PaymentFailed{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "PaymentFailed",
			AggregateIDV: paymentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID: orderID,
		Reason:  reason,
	}
}
