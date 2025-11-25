// Package kafka provides Kafka producer and consumer utilities.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// ProducerConfig holds Kafka producer configuration.
type ProducerConfig struct {
	Brokers      []string      `json:"brokers" yaml:"brokers"`
	Topic        string        `json:"topic" yaml:"topic"`
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	RequiredAcks int           `json:"required_acks" yaml:"required_acks"`
}

// DefaultProducerConfig returns default producer configuration.
func DefaultProducerConfig() ProducerConfig {
	return ProducerConfig{
		Brokers:      []string{"localhost:9092"},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxRetries:   3,
		RequiredAcks: -1, // Wait for all replicas
	}
}

// Producer wraps kafka-go writer with additional functionality.
type Producer struct {
	writer *kafka.Writer
	config ProducerConfig
}

// NewProducer creates a new Kafka producer.
func NewProducer(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		MaxAttempts:  cfg.MaxRetries,
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
	}

	return &Producer{
		writer: writer,
		config: cfg,
	}
}

// Message represents a Kafka message to be published.
type Message struct {
	Key     string
	Value   interface{}
	Headers map[string]string
}

// Publish publishes a single message to Kafka.
func (p *Producer) Publish(ctx context.Context, msg Message) error {
	value, err := json.Marshal(msg.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(msg.Key),
		Value: value,
	}

	// Add headers
	for k, v := range msg.Headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple messages to Kafka.
func (p *Producer) PublishBatch(ctx context.Context, messages []Message) error {
	kafkaMessages := make([]kafka.Message, len(messages))

	for i, msg := range messages {
		value, err := json.Marshal(msg.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal message %d: %w", i, err)
		}

		kafkaMessages[i] = kafka.Message{
			Key:   []byte(msg.Key),
			Value: value,
		}

		for k, v := range msg.Headers {
			kafkaMessages[i].Headers = append(kafkaMessages[i].Headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	if err := p.writer.WriteMessages(ctx, kafkaMessages...); err != nil {
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	return nil
}

// Close closes the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
