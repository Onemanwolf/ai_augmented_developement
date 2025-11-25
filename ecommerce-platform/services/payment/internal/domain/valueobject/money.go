// Package valueobject contains value objects for the Payment domain.
package valueobject

import (
	"fmt"
)

// Currency represents a currency code (ISO 4217).
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
)

// Money represents a monetary value with currency.
type Money struct {
	Amount   int64    `json:"amount" bson:"amount"`       // Amount in smallest unit (cents)
	Currency Currency `json:"currency" bson:"currency"`
}

// NewMoney creates a new Money value object.
func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("amount cannot be negative")
	}
	if !currency.IsValid() {
		return Money{}, fmt.Errorf("invalid currency: %s", currency)
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// Zero returns a zero money value.
func Zero(currency Currency) Money {
	return Money{Amount: 0, Currency: currency}
}

// Add adds two money values.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract subtracts money values.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	if m.Amount < other.Amount {
		return Money{}, fmt.Errorf("cannot subtract: result would be negative")
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// IsZero checks if the money value is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive checks if the money value is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// Equals checks equality with another Money value.
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// IsValid checks if currency is valid.
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP:
		return true
	default:
		return false
	}
}

// String returns string representation.
func (c Currency) String() string {
	return string(c)
}
