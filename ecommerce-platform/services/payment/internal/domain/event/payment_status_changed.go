package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentStatusChanged is emitted when a payment's status changes.
type PaymentStatusChanged struct {
	BaseEvent
	OldStatus valueobject.PaymentStatus `json:"old_status"`
	NewStatus valueobject.PaymentStatus `json:"new_status"`
}

// NewPaymentStatusChanged creates a new PaymentStatusChanged event.
func NewPaymentStatusChanged(
	paymentID valueobject.PaymentID,
	oldStatus valueobject.PaymentStatus,
	newStatus valueobject.PaymentStatus,
) *PaymentStatusChanged {
	return &PaymentStatusChanged{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "PaymentStatusChanged",
			AggregateIDV: paymentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}
}
