package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentReceived is emitted when a payment is successfully processed.
type PaymentReceived struct {
	BaseEvent
	OrderID       string            `json:"order_id"`
	Amount        valueobject.Money `json:"amount"`
	TransactionID string            `json:"transaction_id"`
}

// NewPaymentReceived creates a new PaymentReceived event.
func NewPaymentReceived(
	paymentID valueobject.PaymentID,
	orderID string,
	amount valueobject.Money,
	transactionID string,
) *PaymentReceived {
	return &PaymentReceived{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "PaymentReceived",
			AggregateIDV: paymentID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		OrderID:       orderID,
		Amount:        amount,
		TransactionID: transactionID,
	}
}
