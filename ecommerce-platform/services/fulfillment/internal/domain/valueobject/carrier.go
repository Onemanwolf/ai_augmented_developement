// Package valueobject contains value objects for the Fulfillment domain.
package valueobject

import "fmt"

// Carrier represents a shipping carrier.
type Carrier string

const (
	// CarrierFedEx represents FedEx.
	CarrierFedEx Carrier = "FEDEX"
	// CarrierUPS represents UPS.
	CarrierUPS Carrier = "UPS"
	// CarrierUSPS represents USPS.
	CarrierUSPS Carrier = "USPS"
	// CarrierDHL represents DHL.
	CarrierDHL Carrier = "DHL"
)

// IsValid checks if the carrier is valid.
func (c Carrier) IsValid() bool {
	switch c {
	case CarrierFedEx, CarrierUPS, CarrierUSPS, CarrierDHL:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (c Carrier) String() string {
	return string(c)
}

// ParseCarrier parses a string into a Carrier.
func ParseCarrier(s string) (Carrier, error) {
	carrier := Carrier(s)
	if !carrier.IsValid() {
		return "", fmt.Errorf("invalid carrier: %s", s)
	}
	return carrier, nil
}
