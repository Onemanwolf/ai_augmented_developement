// Package http provides HTTP handlers for the Fulfillment service API.
package http

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/command"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/handler"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/query"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
)

// ShipmentHandler handles HTTP requests for shipments.
type ShipmentHandler struct {
	commandHandler *handler.ShipmentCommandHandler
	queryHandler   *handler.ShipmentQueryHandler
}

// NewShipmentHandler creates a new ShipmentHandler.
func NewShipmentHandler(cmdHandler *handler.ShipmentCommandHandler, qryHandler *handler.ShipmentQueryHandler) *ShipmentHandler {
	return &ShipmentHandler{
		commandHandler: cmdHandler,
		queryHandler:   qryHandler,
	}
}

// CreateShipmentRequest represents the request body for creating a shipment.
type CreateShipmentRequest struct {
	OrderID string                      `json:"order_id"`
	Items   []CreateShipmentItemRequest `json:"items"`
	Address AddressRequest              `json:"address"`
	Carrier string                      `json:"carrier"`
}

// CreateShipmentItemRequest represents an item in the create shipment request.
type CreateShipmentItemRequest struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
}

// AddressRequest represents a shipping address.
type AddressRequest struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// CreateShipmentResponse represents the response for creating a shipment.
type CreateShipmentResponse struct {
	ShipmentID string `json:"shipment_id"`
	Status     string `json:"status"`
}

// ShipOrderRequest represents the request body for shipping an order.
type ShipOrderRequest struct {
	TrackingNumber string `json:"tracking_number"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// CreateShipment handles POST /shipments
func (h *ShipmentHandler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	var req CreateShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	items := make([]command.CreateShipmentItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = command.CreateShipmentItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
		}
	}

	cmd := &command.CreateShipment{
		OrderID: req.OrderID,
		Items:   items,
		ShippingAddress: command.ShippingAddressData{
			Street:     req.Address.Street,
			City:       req.Address.City,
			State:      req.Address.State,
			PostalCode: req.Address.PostalCode,
			Country:    req.Address.Country,
		},
		Carrier: valueobject.Carrier(req.Carrier),
	}

	shipment, err := h.commandHandler.HandleCreateShipment(r.Context(), cmd)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, CreateShipmentResponse{
		ShipmentID: shipment.ID.String(),
		Status:     string(shipment.Status),
	})
}

// GetShipment handles GET /shipments/{id}
func (h *ShipmentHandler) GetShipment(w http.ResponseWriter, r *http.Request) {
	shipmentID := r.PathValue("id")
	if shipmentID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Shipment ID is required")
		return
	}

	qry := &query.GetShipment{ShipmentID: shipmentID}
	result, err := h.queryHandler.HandleGetShipment(r.Context(), qry)
	if err != nil {
		if err == query.ErrShipmentNotFound {
			respondError(w, http.StatusNotFound, "not_found", "Shipment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "query_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ShipOrder handles POST /shipments/{id}/ship
func (h *ShipmentHandler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	shipmentID := r.PathValue("id")
	if shipmentID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Shipment ID is required")
		return
	}

	var req ShipOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	cmd := &command.ShipOrder{
		ShipmentID:     shipmentID,
		TrackingNumber: req.TrackingNumber,
	}

	if err := h.commandHandler.HandleShipOrder(r.Context(), cmd); err != nil {
		respondError(w, http.StatusInternalServerError, "ship_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "shipped"})
}

// CancelShipment handles POST /shipments/{id}/cancel
func (h *ShipmentHandler) CancelShipment(w http.ResponseWriter, r *http.Request) {
	shipmentID := r.PathValue("id")
	if shipmentID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Shipment ID is required")
		return
	}

	cmd := &command.CancelShipment{
		ShipmentID: shipmentID,
	}

	if err := h.commandHandler.HandleCancelShipment(r.Context(), cmd); err != nil {
		respondError(w, http.StatusInternalServerError, "cancel_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// HealthCheck handles GET /health
func (h *ShipmentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, ErrorResponse{
		Error:   code,
		Message: message,
	})
}
