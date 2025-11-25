// Package http provides HTTP handlers for the Order service API.
package http

import (
	"net/http"
)

// RegisterRoutes registers all HTTP routes for the Order service.
func RegisterRoutes(mux *http.ServeMux, h *OrderHandler) {
	// Health check
	mux.HandleFunc("GET /health", h.HealthCheck)

	// Order endpoints
	mux.HandleFunc("POST /api/orders", h.CreateOrder)
	mux.HandleFunc("GET /api/orders/{id}", h.GetOrder)
	mux.HandleFunc("POST /api/orders/{id}/cancel", h.CancelOrder)
}

// Middleware wraps an http.Handler with common middleware.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Add correlation ID if not present
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}
		w.Header().Set("X-Correlation-ID", correlationID)

		next.ServeHTTP(w, r)
	})
}

// generateCorrelationID generates a simple correlation ID.
func generateCorrelationID() string {
	return "req-" + randomString(16)
}

// randomString generates a random string of the given length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
