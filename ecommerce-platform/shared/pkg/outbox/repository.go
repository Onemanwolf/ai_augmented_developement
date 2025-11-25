package outbox

import (
	"context"
	"time"
)

// Repository defines the interface for outbox storage operations.
type Repository interface {
	// Save saves an outbox entry within the current transaction.
	Save(ctx context.Context, entry *Entry) error

	// SaveBatch saves multiple outbox entries within the current transaction.
	SaveBatch(ctx context.Context, entries []*Entry) error

	// FindUnpublished returns unpublished entries up to the given limit.
	FindUnpublished(ctx context.Context, limit int) ([]*Entry, error)

	// MarkAsPublished marks an entry as published.
	MarkAsPublished(ctx context.Context, id string) error

	// MarkBatchAsPublished marks multiple entries as published.
	MarkBatchAsPublished(ctx context.Context, ids []string) error

	// DeletePublished deletes published entries older than the given duration.
	DeletePublished(ctx context.Context, olderThan time.Duration) (int64, error)
}

// Publisher defines the interface for publishing outbox entries.
type Publisher interface {
	// Publish publishes an outbox entry to the message broker.
	Publish(ctx context.Context, entry *Entry) error

	// PublishBatch publishes multiple outbox entries to the message broker.
	PublishBatch(ctx context.Context, entries []*Entry) error
}
