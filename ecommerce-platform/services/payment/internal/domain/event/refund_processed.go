package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// RefundProcessed is emitted when a refund is processed.
type RefundProcessed struct {
	BaseEvent
	OrderID string            `json:"order_id"`
	Amount  valueobject.Money `json:"amount"`
	Reason  string            `json:"reason"`
}

// NewRefundProcessed creates a new RefundProcessed event.
func NewRefundProcessed(
	paymentID valueobject.PaymentID,
	orderID string,
	amount valueobject.Money,
	reason string,
) *RefundProcessed {
	return &RefundProcessed{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "RefundProcessed",
			AggregateIDV: paymentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID: orderID,
		Amount:  amount,
		Reason:  reason,
	}
}
