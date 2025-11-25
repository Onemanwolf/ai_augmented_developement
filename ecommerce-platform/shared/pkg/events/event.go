// Package events provides base event interfaces and types for domain-driven design.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents an event that occurred within a domain aggregate.
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	OccurredAt() time.Time
	Payload() []byte
}

// IntegrationEvent extends DomainEvent for cross-service communication.
type IntegrationEvent interface {
	DomainEvent
	CorrelationID() string
	CausationID() string
}

// BaseEvent provides a base implementation of DomainEvent.
type BaseEvent struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	AggregateIDV  string    `json:"aggregate_id"`
	AggregateT    string    `json:"aggregate_type"`
	OccurredAtV   time.Time `json:"occurred_at"`
	PayloadData   []byte    `json:"payload"`
	CorrelationV  string    `json:"correlation_id,omitempty"`
	CausationV    string    `json:"causation_id,omitempty"`
}

// NewBaseEvent creates a new BaseEvent with generated ID and timestamp.
func NewBaseEvent(eventType, aggregateID, aggregateType string, payload interface{}) (*BaseEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &BaseEvent{
		ID:           uuid.New().String(),
		Type:         eventType,
		AggregateIDV: aggregateID,
		AggregateT:   aggregateType,
		OccurredAtV:  time.Now().UTC(),
		PayloadData:  payloadBytes,
	}, nil
}

// EventID returns the unique identifier of the event.
func (e *BaseEvent) EventID() string {
	return e.ID
}

// EventType returns the type of the event.
func (e *BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the ID of the aggregate that produced this event.
func (e *BaseEvent) AggregateID() string {
	return e.AggregateIDV
}

// AggregateType returns the type of the aggregate.
func (e *BaseEvent) AggregateType() string {
	return e.AggregateT
}

// OccurredAt returns when the event occurred.
func (e *BaseEvent) OccurredAt() time.Time {
	return e.OccurredAtV
}

// Payload returns the event payload as bytes.
func (e *BaseEvent) Payload() []byte {
	return e.PayloadData
}

// CorrelationID returns the correlation ID for distributed tracing.
func (e *BaseEvent) CorrelationID() string {
	return e.CorrelationV
}

// CausationID returns the ID of the event that caused this event.
func (e *BaseEvent) CausationID() string {
	return e.CausationV
}

// WithCorrelation sets the correlation and causation IDs.
func (e *BaseEvent) WithCorrelation(correlationID, causationID string) *BaseEvent {
	e.CorrelationV = correlationID
	e.CausationV = causationID
	return e
}
