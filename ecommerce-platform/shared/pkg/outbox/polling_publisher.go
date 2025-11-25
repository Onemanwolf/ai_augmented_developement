package outbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/your-org/ecommerce-platform/shared/pkg/kafka"
)

// PollingPublisherConfig holds configuration for the polling publisher.
type PollingPublisherConfig struct {
	PollInterval    time.Duration `json:"poll_interval" yaml:"poll_interval"`
	BatchSize       int           `json:"batch_size" yaml:"batch_size"`
	RetentionPeriod time.Duration `json:"retention_period" yaml:"retention_period"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// DefaultPollingPublisherConfig returns default configuration.
func DefaultPollingPublisherConfig() PollingPublisherConfig {
	return PollingPublisherConfig{
		PollInterval:    5 * time.Second,
		BatchSize:       100,
		RetentionPeriod: 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}
}

// PollingPublisher implements a fallback polling-based outbox publisher.
type PollingPublisher struct {
	repository Repository
	producer   *kafka.Producer
	config     PollingPublisherConfig
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
}

// NewPollingPublisher creates a new polling publisher.
func NewPollingPublisher(repo Repository, producer *kafka.Producer, cfg PollingPublisherConfig) *PollingPublisher {
	return &PollingPublisher{
		repository: repo,
		producer:   producer,
		config:     cfg,
		stopCh:     make(chan struct{}),
	}
}

// Start starts the polling publisher.
func (p *PollingPublisher) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("polling publisher already running")
	}
	p.running = true
	p.mu.Unlock()

	// Start polling goroutine
	p.wg.Add(1)
	go p.pollLoop(ctx)

	// Start cleanup goroutine
	p.wg.Add(1)
	go p.cleanupLoop(ctx)

	return nil
}

// Stop stops the polling publisher.
func (p *PollingPublisher) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	close(p.stopCh)
	p.wg.Wait()
}

func (p *PollingPublisher) pollLoop(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			if err := p.publishPendingEntries(ctx); err != nil {
				fmt.Printf("Error publishing pending entries: %v\n", err)
			}
		}
	}
}

func (p *PollingPublisher) publishPendingEntries(ctx context.Context) error {
	entries, err := p.repository.FindUnpublished(ctx, p.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to find unpublished entries: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	publishedIDs := make([]string, 0, len(entries))

	for _, entry := range entries {
		msg := kafka.Message{
			Key:   entry.AggregateID,
			Value: entry.Payload,
			Headers: map[string]string{
				"event_type":     entry.EventType,
				"aggregate_type": entry.AggregateType,
				"event_id":       entry.ID,
			},
		}

		if err := p.producer.Publish(ctx, msg); err != nil {
			fmt.Printf("Failed to publish entry %s: %v\n", entry.ID, err)
			continue
		}

		publishedIDs = append(publishedIDs, entry.ID)
	}

	if len(publishedIDs) > 0 {
		if err := p.repository.MarkBatchAsPublished(ctx, publishedIDs); err != nil {
			return fmt.Errorf("failed to mark entries as published: %w", err)
		}
	}

	return nil
}

func (p *PollingPublisher) cleanupLoop(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			deleted, err := p.repository.DeletePublished(ctx, p.config.RetentionPeriod)
			if err != nil {
				fmt.Printf("Error cleaning up old entries: %v\n", err)
			} else if deleted > 0 {
				fmt.Printf("Cleaned up %d old outbox entries\n", deleted)
			}
		}
	}
}

// IsRunning returns whether the publisher is running.
func (p *PollingPublisher) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
