package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implements Logger using uber-go/zap.
type ZapLogger struct {
	logger        *zap.Logger
	correlationID string
}

// Config holds logger configuration.
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	OutputPath string // stdout, stderr, or file path
}

// DefaultConfig returns default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}
}

// NewZapLogger creates a new ZapLogger with the given configuration.
func NewZapLogger(cfg Config) (*ZapLogger, error) {
	level := parseLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var output zapcore.WriteSyncer
	switch cfg.OutputPath {
	case "stdout":
		output = zapcore.AddSync(os.Stdout)
	case "stderr":
		output = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	core := zapcore.NewCore(encoder, output, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{logger: logger}, nil
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *ZapLogger) toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)+1)
	if l.correlationID != "" {
		zapFields = append(zapFields, zap.String("correlation_id", l.correlationID))
	}
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}

// Debug logs a debug message.
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.toZapFields(fields)...)
}

// Info logs an info message.
func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.toZapFields(fields)...)
}

// Warn logs a warning message.
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.toZapFields(fields)...)
}

// Error logs an error message.
func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, l.toZapFields(fields)...)
}

// Fatal logs a fatal message and exits.
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, l.toZapFields(fields)...)
}

// WithField returns a new logger with the given field.
func (l *ZapLogger) WithField(key string, value interface{}) Logger {
	return &ZapLogger{
		logger:        l.logger.With(zap.Any(key, value)),
		correlationID: l.correlationID,
	}
}

// WithFields returns a new logger with the given fields.
func (l *ZapLogger) WithFields(fields ...Field) Logger {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return &ZapLogger{
		logger:        l.logger.With(zapFields...),
		correlationID: l.correlationID,
	}
}

// WithCorrelationID returns a new logger with the given correlation ID.
func (l *ZapLogger) WithCorrelationID(correlationID string) Logger {
	return &ZapLogger{
		logger:        l.logger,
		correlationID: correlationID,
	}
}

// WithContext returns a new logger with correlation ID from context.
func (l *ZapLogger) WithContext(ctx context.Context) Logger {
	correlationID := CorrelationIDFromContext(ctx)
	return l.WithCorrelationID(correlationID)
}

// Sync flushes any buffered log entries.
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
