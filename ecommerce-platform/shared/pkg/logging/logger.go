// Package logging provides structured logging utilities for microservices.
package logging

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// CorrelationIDKey is the context key for correlation IDs.
	CorrelationIDKey contextKey = "correlation_id"
)

// Logger defines the interface for structured logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	WithField(key string, value interface{}) Logger
	WithFields(fields ...Field) Logger
	WithCorrelationID(correlationID string) Logger
	WithContext(ctx context.Context) Logger
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field.
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field.
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a field with any value.
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// CorrelationIDFromContext extracts the correlation ID from context.
func CorrelationIDFromContext(ctx context.Context) string {
	if v := ctx.Value(CorrelationIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// ContextWithCorrelationID adds a correlation ID to the context.
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}
