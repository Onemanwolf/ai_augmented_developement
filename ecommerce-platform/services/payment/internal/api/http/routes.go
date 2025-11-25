// Package http provides HTTP handlers for the Payment service API.
package http

import (
	"net/http"
)

// RegisterRoutes registers all HTTP routes for the Payment service.
func RegisterRoutes(mux *http.ServeMux, h *PaymentHandler) {
	// Health check
	mux.HandleFunc("GET /health", h.HealthCheck)

	// Payment endpoints
	mux.HandleFunc("POST /api/payments", h.ProcessPayment)
	mux.HandleFunc("GET /api/payments/{id}", h.GetPayment)
	mux.HandleFunc("POST /api/payments/{id}/refund", h.RefundPayment)
}
