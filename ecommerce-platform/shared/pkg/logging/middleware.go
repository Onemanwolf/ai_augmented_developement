// Package logging provides structured logging utilities using zap.
package logging

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// HTTPMiddleware returns HTTP middleware that adds logging to requests.
func (l *Logger) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get or generate correlation ID
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Get or generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Create request-scoped logger
		requestLogger := l.WithCorrelationID(correlationID).WithRequestID(requestID)

		// Add logger to context
		ctx := requestLogger.WithContext(r.Context())
		r = r.WithContext(ctx)

		// Set response headers
		w.Header().Set("X-Correlation-ID", correlationID)
		w.Header().Set("X-Request-ID", requestID)

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request completion
		duration := time.Since(start)
		requestLogger.HTTPRequest(
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
