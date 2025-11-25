// Package event contains domain events for the Fulfillment service.
package event

import (
	"time"
)

// DomainEvent is the interface all domain events must implement.
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// BaseEvent provides common event fields.
type BaseEvent struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	AggregateIDV string    `json:"aggregate_id"`
	OccurredAtV  time.Time `json:"occurred_at"`
}

// EventType returns the event type.
func (e BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID.
func (e BaseEvent) AggregateID() string {
	return e.AggregateIDV
}

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.OccurredAtV
}
