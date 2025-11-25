// Package valueobject contains value objects for the Payment domain.
package valueobject

import "fmt"

// PaymentMethod represents the method of payment.
type PaymentMethod string

const (
	// PaymentMethodCreditCard represents credit card payment.
	PaymentMethodCreditCard PaymentMethod = "CREDIT_CARD"
	// PaymentMethodDebitCard represents debit card payment.
	PaymentMethodDebitCard PaymentMethod = "DEBIT_CARD"
	// PaymentMethodBankTransfer represents bank transfer payment.
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	// PaymentMethodDigitalWallet represents digital wallet payment (PayPal, etc.).
	PaymentMethodDigitalWallet PaymentMethod = "DIGITAL_WALLET"
)

// IsValid checks if the payment method is valid.
func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCreditCard, PaymentMethodDebitCard,
		PaymentMethodBankTransfer, PaymentMethodDigitalWallet:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (m PaymentMethod) String() string {
	return string(m)
}

// ParsePaymentMethod parses a string into a PaymentMethod.
func ParsePaymentMethod(s string) (PaymentMethod, error) {
	method := PaymentMethod(s)
	if !method.IsValid() {
		return "", fmt.Errorf("invalid payment method: %s", s)
	}
	return method, nil
}
