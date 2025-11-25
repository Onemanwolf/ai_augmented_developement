// Package http provides HTTP handlers for the Fulfillment service API.
package http

import (
	"net/http"
)

// RegisterRoutes registers all HTTP routes for the Fulfillment service.
func RegisterRoutes(mux *http.ServeMux, h *ShipmentHandler) {
	// Health check
	mux.HandleFunc("GET /health", h.HealthCheck)

	// Shipment endpoints
	mux.HandleFunc("POST /api/shipments", h.CreateShipment)
	mux.HandleFunc("GET /api/shipments/{id}", h.GetShipment)
	mux.HandleFunc("POST /api/shipments/{id}/ship", h.ShipOrder)
	mux.HandleFunc("POST /api/shipments/{id}/cancel", h.CancelShipment)
}
