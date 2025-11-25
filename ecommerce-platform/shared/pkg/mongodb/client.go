// Package mongodb provides MongoDB client utilities and helpers.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds MongoDB connection configuration.
type Config struct {
	URI            string        `json:"uri" yaml:"uri"`
	Database       string        `json:"database" yaml:"database"`
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	PingTimeout    time.Duration `json:"ping_timeout" yaml:"ping_timeout"`
	MaxPoolSize    uint64        `json:"max_pool_size" yaml:"max_pool_size"`
	MinPoolSize    uint64        `json:"min_pool_size" yaml:"min_pool_size"`
}

// DefaultConfig returns default MongoDB configuration.
func DefaultConfig() Config {
	return Config{
		URI:            "mongodb://localhost:27017",
		Database:       "ecommerce",
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    10,
	}
}

// Client wraps the MongoDB client with additional functionality.
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

// NewClient creates a new MongoDB client with the given configuration.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetConnectTimeout(cfg.ConnectTimeout)

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	pingCtx, pingCancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
	}, nil
}

// Database returns the configured database.
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection returns a collection from the configured database.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Client returns the underlying MongoDB client.
func (c *Client) Client() *mongo.Client {
	return c.client
}

// Close disconnects the MongoDB client.
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Health checks if the MongoDB connection is healthy.
func (c *Client) Health(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, c.config.PingTimeout)
	defer cancel()
	return c.client.Ping(pingCtx, readpref.Primary())
}

// StartSession starts a new MongoDB session for transactions.
func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return c.client.StartSession(opts...)
}

// WithTransaction executes the given function within a transaction.
func (c *Client) WithTransaction(ctx context.Context, fn func(sc mongo.SessionContext) error) error {
	session, err := c.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return nil, fn(sc)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}
