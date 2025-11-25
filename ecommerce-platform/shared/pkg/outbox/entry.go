// Package outbox implements the transactional outbox pattern.
package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/shared/pkg/events"
)

// Entry represents an outbox entry for reliable event publishing.
type Entry struct {
	ID            string    `json:"id" bson:"_id"`
	AggregateID   string    `json:"aggregate_id" bson:"aggregate_id"`
	AggregateType string    `json:"aggregate_type" bson:"aggregate_type"`
	EventType     string    `json:"event_type" bson:"event_type"`
	Payload       []byte    `json:"payload" bson:"payload"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	Published     bool      `json:"published" bson:"published"`
	PublishedAt   time.Time `json:"published_at,omitempty" bson:"published_at,omitempty"`
}

// NewEntry creates a new outbox entry from a domain event.
func NewEntry(event events.DomainEvent) (*Entry, error) {
	payload := event.Payload()
	if payload == nil {
		// If payload is nil, serialize the entire event
		var err error
		payload, err = json.Marshal(event)
		if err != nil {
			return nil, err
		}
	}

	return &Entry{
		ID:            uuid.New().String(),
		AggregateID:   event.AggregateID(),
		AggregateType: event.AggregateType(),
		EventType:     event.EventType(),
		Payload:       payload,
		CreatedAt:     time.Now().UTC(),
		Published:     false,
	}, nil
}

// NewEntryFromRaw creates an outbox entry from raw data.
func NewEntryFromRaw(aggregateID, aggregateType, eventType string, payload interface{}) (*Entry, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Entry{
		ID:            uuid.New().String(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       payloadBytes,
		CreatedAt:     time.Now().UTC(),
		Published:     false,
	}, nil
}

// MarkAsPublished marks the entry as published.
func (e *Entry) MarkAsPublished() {
	e.Published = true
	e.PublishedAt = time.Now().UTC()
}

// IsExpired checks if the entry is older than the given duration.
func (e *Entry) IsExpired(retention time.Duration) bool {
	return time.Since(e.CreatedAt) > retention
}
