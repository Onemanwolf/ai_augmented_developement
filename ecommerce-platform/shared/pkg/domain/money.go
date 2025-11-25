package domain

import (
	"fmt"
)

// Currency represents an ISO 4217 currency code.
type Currency string

// Common currency codes.
const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
)

// ValidCurrencies contains all supported currency codes.
var ValidCurrencies = map[Currency]bool{
	USD: true,
	EUR: true,
	GBP: true,
}

// Money represents a monetary amount with currency.
// Amount is stored in the smallest currency unit (e.g., cents for USD).
type Money struct {
	Amount   int64    `json:"amount" bson:"amount"`
	Currency Currency `json:"currency" bson:"currency"`
}

// NewMoney creates a new Money value object.
func NewMoney(amount int64, currency Currency) (Money, error) {
	if !ValidCurrencies[currency] {
		return Money{}, fmt.Errorf("invalid currency: %s", currency)
	}
	if amount < 0 {
		return Money{}, fmt.Errorf("amount cannot be negative: %d", amount)
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// MustNewMoney creates a new Money and panics if invalid.
func MustNewMoney(amount int64, currency Currency) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero returns a zero amount for the given currency.
func Zero(currency Currency) Money {
	return Money{Amount: 0, Currency: currency}
}

// Add returns a new Money with the sum of the amounts.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract returns a new Money with the difference of the amounts.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	result := m.Amount - other.Amount
	if result < 0 {
		return Money{}, fmt.Errorf("subtraction would result in negative amount")
	}
	return Money{Amount: result, Currency: m.Currency}, nil
}

// Multiply returns a new Money with the amount multiplied by the factor.
func (m Money) Multiply(factor int64) Money {
	return Money{Amount: m.Amount * factor, Currency: m.Currency}
}

// Equals returns true if the amounts and currencies are equal.
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive returns true if the amount is greater than zero.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// String returns a human-readable representation.
func (m Money) String() string {
	// Convert cents to dollars for display
	dollars := float64(m.Amount) / 100
	return fmt.Sprintf("%.2f %s", dollars, m.Currency)
}
