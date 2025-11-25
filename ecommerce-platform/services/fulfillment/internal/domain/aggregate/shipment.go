// Package aggregate contains aggregate roots for the Fulfillment domain.
package aggregate

import (
	"fmt"
	"time"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/event"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// DomainEvent represents a domain event interface.
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// Shipment is the aggregate root for shipments.
type Shipment struct {
	ID              valueobject.ShipmentID     `bson:"_id"`
	OrderID         string                     `bson:"order_id"`
	CustomerID      string                     `bson:"customer_id"`
	Items           []*entity.ShipmentItem     `bson:"items"`
	ShippingAddress valueobject.Address        `bson:"shipping_address"`
	Carrier         valueobject.Carrier        `bson:"carrier"`
	TrackingNumber  string                     `bson:"tracking_number,omitempty"`
	Status          valueobject.ShipmentStatus `bson:"status"`
	FailureReason   string                     `bson:"failure_reason,omitempty"`
	EstimatedDelivery *time.Time               `bson:"estimated_delivery,omitempty"`
	ActualDelivery    *time.Time               `bson:"actual_delivery,omitempty"`
	CreatedAt       time.Time                  `bson:"created_at"`
	UpdatedAt       time.Time                  `bson:"updated_at"`
	ShippedAt       *time.Time                 `bson:"shipped_at,omitempty"`
	Version         int                        `bson:"version"`

	events []DomainEvent
}

// NewShipment creates a new Shipment aggregate.
func NewShipment(
	orderID string,
	customerID string,
	items []*entity.ShipmentItem,
	address valueobject.Address,
	carrier valueobject.Carrier,
) (*Shipment, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}
	if address.IsEmpty() {
		return nil, fmt.Errorf("shipping address is required")
	}
	if !carrier.IsValid() {
		return nil, fmt.Errorf("invalid carrier")
	}

	now := time.Now().UTC()
	shipment := &Shipment{
		ID:              valueobject.NewShipmentID(),
		OrderID:         orderID,
		CustomerID:      customerID,
		Items:           items,
		ShippingAddress: address,
		Carrier:         carrier,
		Status:          valueobject.ShipmentStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
		events:          make([]DomainEvent, 0),
	}

	shipment.addEvent(event.NewShipmentCreated(
		shipment.ID,
		orderID,
		customerID,
		carrier,
		address,
	))

	return shipment, nil
}

// StartProcessing starts processing the shipment.
func (s *Shipment) StartProcessing() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusProcessing) {
		return fmt.Errorf("cannot process shipment in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusProcessing
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	return nil
}

// MarkReadyForPickup marks the shipment as ready for carrier pickup.
func (s *Shipment) MarkReadyForPickup(trackingNumber string, estimatedDelivery time.Time) error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusReadyForPickup) {
		return fmt.Errorf("cannot mark shipment ready in status %s", s.Status)
	}
	if trackingNumber == "" {
		return fmt.Errorf("tracking number is required")
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusReadyForPickup
	s.TrackingNumber = trackingNumber
	s.EstimatedDelivery = &estimatedDelivery
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	return nil
}

// Ship marks the shipment as shipped (picked up by carrier).
func (s *Shipment) Ship() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusPickedUp) {
		return fmt.Errorf("cannot ship in status %s", s.Status)
	}

	now := time.Now().UTC()
	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusPickedUp
	s.ShippedAt = &now
	s.UpdatedAt = now
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	s.addEvent(event.NewOrderShipped(s.ID, s.OrderID, s.TrackingNumber, s.Carrier))
	return nil
}

// UpdateToInTransit updates status to in transit.
func (s *Shipment) UpdateToInTransit() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusInTransit) {
		return fmt.Errorf("cannot update to in_transit in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusInTransit
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	return nil
}

// UpdateToOutForDelivery updates status to out for delivery.
func (s *Shipment) UpdateToOutForDelivery() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusOutForDelivery) {
		return fmt.Errorf("cannot update to out_for_delivery in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusOutForDelivery
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	return nil
}

// MarkDelivered marks the shipment as delivered.
func (s *Shipment) MarkDelivered() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusDelivered) {
		return fmt.Errorf("cannot mark delivered in status %s", s.Status)
	}

	now := time.Now().UTC()
	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusDelivered
	s.ActualDelivery = &now
	s.UpdatedAt = now
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	s.addEvent(event.NewOrderDelivered(s.ID, s.OrderID))
	return nil
}

// MarkFailed marks the shipment as failed.
func (s *Shipment) MarkFailed(reason string) error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusFailed) {
		return fmt.Errorf("cannot mark failed in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusFailed
	s.FailureReason = reason
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	s.addEvent(event.NewShipmentFailed(s.ID, s.OrderID, reason))
	return nil
}

// Cancel cancels a pending shipment.
func (s *Shipment) Cancel() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusCancelled) {
		return fmt.Errorf("cannot cancel shipment in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusCancelled
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	s.addEvent(event.NewShipmentCancelled(s.ID, s.OrderID))
	return nil
}

// MarkReturned marks the shipment as returned.
func (s *Shipment) MarkReturned(reason string) error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusReturned) {
		return fmt.Errorf("cannot mark returned in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusReturned
	s.FailureReason = reason
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	return nil
}

// Retry retries a failed shipment.
func (s *Shipment) Retry() error {
	if !s.Status.CanTransitionTo(valueobject.ShipmentStatusPending) {
		return fmt.Errorf("cannot retry shipment in status %s", s.Status)
	}

	oldStatus := s.Status
	s.Status = valueobject.ShipmentStatusPending
	s.FailureReason = ""
	s.UpdatedAt = time.Now().UTC()
	s.Version++

	s.addEvent(event.NewShipmentStatusChanged(s.ID, oldStatus, s.Status))
	return nil
}

// GetEvents returns and clears domain events.
func (s *Shipment) GetEvents() []DomainEvent {
	events := s.events
	s.events = make([]DomainEvent, 0)
	return events
}

// ClearEvents clears all domain events.
func (s *Shipment) ClearEvents() {
	s.events = make([]DomainEvent, 0)
}

func (s *Shipment) addEvent(e DomainEvent) {
	s.events = append(s.events, e)
}

// TotalWeight returns the total weight of all items.
func (s *Shipment) TotalWeight() int {
	total := 0
	for _, item := range s.Items {
		total += item.TotalWeight()
	}
	return total
}

// IsPending checks if shipment is pending.
func (s *Shipment) IsPending() bool {
	return s.Status == valueobject.ShipmentStatusPending
}

// IsDelivered checks if shipment is delivered.
func (s *Shipment) IsDelivered() bool {
	return s.Status == valueobject.ShipmentStatusDelivered
}

// IsShipped checks if shipment has been shipped.
func (s *Shipment) IsShipped() bool {
	return s.ShippedAt != nil
}
