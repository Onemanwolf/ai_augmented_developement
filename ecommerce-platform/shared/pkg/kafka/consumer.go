package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// ConsumerConfig holds Kafka consumer configuration.
type ConsumerConfig struct {
	Brokers        []string      `json:"brokers" yaml:"brokers"`
	GroupID        string        `json:"group_id" yaml:"group_id"`
	Topics         []string      `json:"topics" yaml:"topics"`
	MinBytes       int           `json:"min_bytes" yaml:"min_bytes"`
	MaxBytes       int           `json:"max_bytes" yaml:"max_bytes"`
	MaxWait        time.Duration `json:"max_wait" yaml:"max_wait"`
	CommitInterval time.Duration `json:"commit_interval" yaml:"commit_interval"`
	StartOffset    int64         `json:"start_offset" yaml:"start_offset"`
}

// DefaultConsumerConfig returns default consumer configuration.
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		Brokers:        []string{"localhost:9092"},
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.FirstOffset,
	}
}

// MessageHandler is a function that handles consumed messages.
type MessageHandler func(ctx context.Context, msg *ConsumedMessage) error

// ConsumedMessage represents a message received from Kafka.
type ConsumedMessage struct {
	Topic     string
	Partition int
	Offset    int64
	Key       string
	Value     []byte
	Headers   map[string]string
	Timestamp time.Time
}

// Unmarshal deserializes the message value into the given target.
func (m *ConsumedMessage) Unmarshal(target interface{}) error {
	return json.Unmarshal(m.Value, target)
}

// Consumer wraps kafka-go reader with additional functionality.
type Consumer struct {
	reader   *kafka.Reader
	config   ConsumerConfig
	handlers map[string]MessageHandler
	stopCh   chan struct{}
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		GroupTopics:    cfg.Topics,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		CommitInterval: cfg.CommitInterval,
		StartOffset:    cfg.StartOffset,
	})

	return &Consumer{
		reader:   reader,
		config:   cfg,
		handlers: make(map[string]MessageHandler),
		stopCh:   make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a specific event type.
func (c *Consumer) RegisterHandler(eventType string, handler MessageHandler) {
	c.handlers[eventType] = handler
}

// Start begins consuming messages.
func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopCh:
			return nil
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			}

			consumed := &ConsumedMessage{
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       string(msg.Key),
				Value:     msg.Value,
				Headers:   make(map[string]string),
				Timestamp: msg.Time,
			}

			for _, h := range msg.Headers {
				consumed.Headers[h.Key] = string(h.Value)
			}

			// Get event type from headers (check both formats)
			eventType := consumed.Headers["event-type"]
			if eventType == "" {
				eventType = consumed.Headers["event_type"]
			}
			if handler, ok := c.handlers[eventType]; ok {
				if err := handler(ctx, consumed); err != nil {
					// Log error but continue processing
					fmt.Printf("Error handling message: %v\n", err)
					continue
				}
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				fmt.Printf("Error committing message: %v\n", err)
			}
		}
	}
}

// Stop stops the consumer.
func (c *Consumer) Stop() {
	close(c.stopCh)
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	c.Stop()
	return c.reader.Close()
}
