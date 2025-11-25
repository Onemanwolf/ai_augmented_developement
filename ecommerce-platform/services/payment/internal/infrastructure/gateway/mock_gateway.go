// Package gateway provides payment gateway implementations.
package gateway

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// MockPaymentGateway is a mock payment gateway for testing.
type MockPaymentGateway struct{}

// NewMockPaymentGateway creates a new mock payment gateway.
func NewMockPaymentGateway() *MockPaymentGateway {
	return &MockPaymentGateway{}
}

// ProcessPayment simulates processing a payment.
func (g *MockPaymentGateway) ProcessPayment(ctx context.Context, amount valueobject.Money, method valueobject.PaymentMethod) (string, error) {
	// Simulate successful payment processing
	transactionID := fmt.Sprintf("txn_%s", uuid.New().String())
	return transactionID, nil
}

// RefundPayment simulates refunding a payment.
func (g *MockPaymentGateway) RefundPayment(ctx context.Context, transactionID string, amount valueobject.Money) error {
	// Simulate successful refund
	return nil
}
