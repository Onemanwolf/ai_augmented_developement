// Package kafka provides Kafka producer and consumer utilities.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// EventPublisher publishes domain events to Kafka topics.
type EventPublisher struct {
	writer *kafka.Writer
	topic  string
}

// EventPublisherConfig holds configuration for the event publisher.
type EventPublisherConfig struct {
	Brokers []string
	Topic   string
}

// NewEventPublisher creates a new Kafka event publisher.
func NewEventPublisher(cfg EventPublisherConfig) *EventPublisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1, // Send immediately for real-time events
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
		RequiredAcks: kafka.RequireAll,
	}

	return &EventPublisher{
		writer: writer,
		topic:  cfg.Topic,
	}
}

// DomainEvent represents a domain event envelope.
type DomainEvent struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`
	AggregateID   string      `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	Timestamp     time.Time   `json:"timestamp"`
	Version       int         `json:"version"`
	Payload       interface{} `json:"payload"`
}

// Publish publishes domain events to Kafka.
func (p *EventPublisher) Publish(ctx context.Context, events []interface{}) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		// Wrap event in envelope
		envelope := DomainEvent{
			ID:        uuid.New().String(),
			Timestamp: time.Now().UTC(),
			Version:   1,
			Payload:   event,
		}

		// Extract type information if available
		if typed, ok := event.(interface{ EventType() string }); ok {
			envelope.Type = typed.EventType()
		} else {
			envelope.Type = fmt.Sprintf("%T", event)
		}

		if agg, ok := event.(interface{ AggregateID() string }); ok {
			envelope.AggregateID = agg.AggregateID()
		}

		if aggType, ok := event.(interface{ AggregateType() string }); ok {
			envelope.AggregateType = aggType.AggregateType()
		}

		value, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(envelope.AggregateID),
			Value: value,
			Headers: []kafka.Header{
				{Key: "event-type", Value: []byte(envelope.Type)},
				{Key: "event-id", Value: []byte(envelope.ID)},
				{Key: "timestamp", Value: []byte(envelope.Timestamp.Format(time.RFC3339))},
			},
		})
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("failed to publish events to Kafka: %w", err)
	}

	return nil
}

// Close closes the event publisher.
func (p *EventPublisher) Close() error {
	return p.writer.Close()
}
