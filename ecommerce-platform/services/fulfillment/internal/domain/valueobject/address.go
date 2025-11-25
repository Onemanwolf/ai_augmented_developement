// Package valueobject contains value objects for the Fulfillment domain.
package valueobject

import "fmt"

// Address represents a shipping address.
type Address struct {
	Street     string `json:"street" bson:"street"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	PostalCode string `json:"postal_code" bson:"postal_code"`
	Country    string `json:"country" bson:"country"`
}

// NewAddress creates a new Address value object.
func NewAddress(street, city, state, postalCode, country string) (Address, error) {
	if street == "" {
		return Address{}, fmt.Errorf("street is required")
	}
	if city == "" {
		return Address{}, fmt.Errorf("city is required")
	}
	if state == "" {
		return Address{}, fmt.Errorf("state is required")
	}
	if postalCode == "" {
		return Address{}, fmt.Errorf("postal code is required")
	}
	if country == "" {
		return Address{}, fmt.Errorf("country is required")
	}

	return Address{
		Street:     street,
		City:       city,
		State:      state,
		PostalCode: postalCode,
		Country:    country,
	}, nil
}

// Equals checks equality with another Address.
func (a Address) Equals(other Address) bool {
	return a.Street == other.Street &&
		a.City == other.City &&
		a.State == other.State &&
		a.PostalCode == other.PostalCode &&
		a.Country == other.Country
}

// String returns a formatted address string.
func (a Address) String() string {
	return fmt.Sprintf("%s, %s, %s %s, %s",
		a.Street, a.City, a.State, a.PostalCode, a.Country)
}

// IsEmpty checks if the address is empty.
func (a Address) IsEmpty() bool {
	return a.Street == "" && a.City == "" && a.State == "" &&
		a.PostalCode == "" && a.Country == ""
}
