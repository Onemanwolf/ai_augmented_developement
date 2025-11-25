// Package event contains domain events for the Fulfillment service.
package event

import (
	"encoding/json"
	"time"
)

// DomainEvent is the interface all domain events must implement.
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	OccurredAt() time.Time
	Payload() []byte
}

// BaseEvent provides common event fields.
type BaseEvent struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	AggregateIDV string    `json:"aggregate_id"`
	OccurredAtV  time.Time `json:"occurred_at"`
	PayloadData  []byte    `json:"-"`
}

// EventID returns the event ID.
func (e BaseEvent) EventID() string {
	return e.ID
}

// EventType returns the event type.
func (e BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID.
func (e BaseEvent) AggregateID() string {
	return e.AggregateIDV
}

// AggregateType returns "Shipment" for all fulfillment events.
func (e BaseEvent) AggregateType() string {
	return "Shipment"
}

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.OccurredAtV
}

// Payload returns the event payload as bytes.
func (e BaseEvent) Payload() []byte {
	return e.PayloadData
}

// SetPayload sets the payload from any serializable data.
func (e *BaseEvent) SetPayload(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	e.PayloadData = bytes
	return nil
}
