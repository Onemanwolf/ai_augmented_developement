// Package event contains domain events for the Order service.
package event

import (
	"time"
)

// DomainEvent is the interface for all domain events.
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	OccurredAt() time.Time
}

// BaseEvent provides common fields for domain events.
type BaseEvent struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	AggregateIDV string    `json:"aggregate_id"`
	OccurredAtV  time.Time `json:"occurred_at"`
}

// EventID returns the event ID.
func (e *BaseEvent) EventID() string {
	return e.ID
}

// EventType returns the event type.
func (e *BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID.
func (e *BaseEvent) AggregateID() string {
	return e.AggregateIDV
}

// AggregateType returns "Order" for all order events.
func (e *BaseEvent) AggregateType() string {
	return "Order"
}

// OccurredAt returns when the event occurred.
func (e *BaseEvent) OccurredAt() time.Time {
	return e.OccurredAtV
}
