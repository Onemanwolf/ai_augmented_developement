// Package valueobject contains value objects for the Fulfillment domain.
package valueobject

import "fmt"

// ShipmentStatus represents the status of a shipment.
type ShipmentStatus string

const (
	// ShipmentStatusPending indicates shipment is awaiting processing.
	ShipmentStatusPending ShipmentStatus = "PENDING"
	// ShipmentStatusProcessing indicates shipment is being prepared.
	ShipmentStatusProcessing ShipmentStatus = "PROCESSING"
	// ShipmentStatusReadyForPickup indicates shipment is ready for carrier pickup.
	ShipmentStatusReadyForPickup ShipmentStatus = "READY_FOR_PICKUP"
	// ShipmentStatusPickedUp indicates shipment was picked up by carrier.
	ShipmentStatusPickedUp ShipmentStatus = "PICKED_UP"
	// ShipmentStatusInTransit indicates shipment is in transit.
	ShipmentStatusInTransit ShipmentStatus = "IN_TRANSIT"
	// ShipmentStatusOutForDelivery indicates shipment is out for delivery.
	ShipmentStatusOutForDelivery ShipmentStatus = "OUT_FOR_DELIVERY"
	// ShipmentStatusDelivered indicates shipment was delivered.
	ShipmentStatusDelivered ShipmentStatus = "DELIVERED"
	// ShipmentStatusFailed indicates shipment delivery failed.
	ShipmentStatusFailed ShipmentStatus = "FAILED"
	// ShipmentStatusReturned indicates shipment was returned.
	ShipmentStatusReturned ShipmentStatus = "RETURNED"
	// ShipmentStatusCancelled indicates shipment was cancelled.
	ShipmentStatusCancelled ShipmentStatus = "CANCELLED"
)

// validShipmentTransitions defines allowed state transitions.
var validShipmentTransitions = map[ShipmentStatus][]ShipmentStatus{
	ShipmentStatusPending:        {ShipmentStatusProcessing, ShipmentStatusCancelled},
	ShipmentStatusProcessing:     {ShipmentStatusReadyForPickup, ShipmentStatusCancelled},
	ShipmentStatusReadyForPickup: {ShipmentStatusPickedUp, ShipmentStatusCancelled},
	ShipmentStatusPickedUp:       {ShipmentStatusInTransit},
	ShipmentStatusInTransit:      {ShipmentStatusOutForDelivery, ShipmentStatusFailed},
	ShipmentStatusOutForDelivery: {ShipmentStatusDelivered, ShipmentStatusFailed},
	ShipmentStatusDelivered:      {ShipmentStatusReturned},
	ShipmentStatusFailed:         {ShipmentStatusReturned, ShipmentStatusPending}, // Retry or return
	ShipmentStatusReturned:       {},                                              // Terminal state
	ShipmentStatusCancelled:      {},                                              // Terminal state
}

// IsValid checks if the status is valid.
func (s ShipmentStatus) IsValid() bool {
	switch s {
	case ShipmentStatusPending, ShipmentStatusProcessing, ShipmentStatusReadyForPickup,
		ShipmentStatusPickedUp, ShipmentStatusInTransit, ShipmentStatusOutForDelivery,
		ShipmentStatusDelivered, ShipmentStatusFailed, ShipmentStatusReturned,
		ShipmentStatusCancelled:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if transition to target status is allowed.
func (s ShipmentStatus) CanTransitionTo(target ShipmentStatus) bool {
	allowed, ok := validShipmentTransitions[s]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

// String returns the string representation.
func (s ShipmentStatus) String() string {
	return string(s)
}

// ParseShipmentStatus parses a string into a ShipmentStatus.
func ParseShipmentStatus(s string) (ShipmentStatus, error) {
	status := ShipmentStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid shipment status: %s", s)
	}
	return status, nil
}

// IsTerminal checks if the status is a terminal state.
func (s ShipmentStatus) IsTerminal() bool {
	return s == ShipmentStatusReturned || s == ShipmentStatusCancelled
}

// IsDelivered checks if the shipment was successfully delivered.
func (s ShipmentStatus) IsDelivered() bool {
	return s == ShipmentStatusDelivered
}
