// Package logging provides structured logging utilities using zap.
package logging

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// contextKey is a type for context keys.
type contextKey string

const (
	// LoggerKey is the context key for the logger.
	LoggerKey contextKey = "logger"
	// CorrelationIDKey is the context key for correlation ID.
	CorrelationIDKey contextKey = "correlation_id"
	// RequestIDKey is the context key for request ID.
	RequestIDKey contextKey = "request_id"
)

// Config holds logger configuration.
type Config struct {
	Level       string
	Format      string // "json" or "console"
	ServiceName string
	Environment string
	Version     string
}

// Logger wraps zap.Logger with additional functionality.
type Logger struct {
	*zap.Logger
	config Config
}

// New creates a new structured logger.
func New(cfg Config) (*Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	fields := []zap.Field{
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
	}
	if cfg.Version != "" {
		fields = append(fields, zap.String("version", cfg.Version))
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(fields...),
	)

	return &Logger{
		Logger: logger,
		config: cfg,
	}, nil
}

// NewDefault creates a logger with default configuration.
func NewDefault(serviceName string) *Logger {
	cfg := Config{
		Level:       getEnvOrDefault("LOG_LEVEL", "info"),
		Format:      getEnvOrDefault("LOG_FORMAT", "json"),
		ServiceName: serviceName,
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		Version:     getEnvOrDefault("VERSION", "unknown"),
	}

	logger, err := New(cfg)
	if err != nil {
		zapLogger, _ := zap.NewProduction()
		return &Logger{Logger: zapLogger, config: cfg}
	}

	return logger
}

// WithContext adds the logger to a context.
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

// FromContext retrieves the logger from context.
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	return NewDefault("unknown")
}

// WithCorrelationID adds a correlation ID to the logger.
func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	return &Logger{
		Logger: l.With(zap.String("correlation_id", correlationID)),
		config: l.config,
	}
}

// WithRequestID adds a request ID to the logger.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With(zap.String("request_id", requestID)),
		config: l.config,
	}
}

// WithFields adds multiple fields to the logger.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{
		Logger: l.With(zapFields...),
		config: l.config,
	}
}

// WithError adds an error field to the logger.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.With(zap.Error(err)),
		config: l.config,
	}
}

// WithDuration adds a duration field to the logger.
func (l *Logger) WithDuration(d time.Duration) *Logger {
	return &Logger{
		Logger: l.With(zap.Duration("duration", d)),
		config: l.config,
	}
}

// Event logs a domain event.
func (l *Logger) Event(eventType string, aggregateID string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("event_type", eventType),
		zap.String("aggregate_id", aggregateID),
	}, fields...)
	l.Info("domain_event", allFields...)
}

// HTTPRequest logs an HTTP request.
func (l *Logger) HTTPRequest(method, path string, status int, duration time.Duration, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", status),
		zap.Duration("duration", duration),
	}, fields...)
	l.Info("http_request", allFields...)
}

// DBOperation logs a database operation.
func (l *Logger) DBOperation(operation, collection string, duration time.Duration, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("operation", operation),
		zap.String("collection", collection),
		zap.Duration("duration", duration),
	}, fields...)
	l.Debug("db_operation", allFields...)
}

// KafkaMessage logs a Kafka message event.
func (l *Logger) KafkaMessage(topic string, partition int, offset int64, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("topic", topic),
		zap.Int("partition", partition),
		zap.Int64("offset", offset),
	}, fields...)
	l.Debug("kafka_message", allFields...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
