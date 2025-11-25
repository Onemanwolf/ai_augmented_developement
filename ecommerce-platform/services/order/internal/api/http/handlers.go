// Package http provides HTTP handlers for the Order service API.
package http

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/ecommerce-platform/services/order/internal/application/command"
	"github.com/your-org/ecommerce-platform/services/order/internal/application/handler"
	"github.com/your-org/ecommerce-platform/services/order/internal/application/query"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	commandHandler *handler.OrderCommandHandler
	queryHandler   *handler.OrderQueryHandler
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(cmdHandler *handler.OrderCommandHandler, qryHandler *handler.OrderQueryHandler) *OrderHandler {
	return &OrderHandler{
		commandHandler: cmdHandler,
		queryHandler:   qryHandler,
	}
}

// CreateOrderRequest represents the request body for creating an order.
type CreateOrderRequest struct {
	CustomerID string                   `json:"customer_id"`
	Currency   string                   `json:"currency"`
	Items      []CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest represents an item in the create order request.
type CreateOrderItemRequest struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

// CreateOrderResponse represents the response for creating an order.
type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// CancelOrderRequest represents the request body for canceling an order.
type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Convert request to command
	items := make([]command.CreateOrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = command.CreateOrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	cmd := &command.CreateOrder{
		CustomerID: req.CustomerID,
		Currency:   valueobject.Currency(req.Currency),
		Items:      items,
	}

	order, err := h.commandHandler.HandleCreateOrder(r.Context(), cmd)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, CreateOrderResponse{
		OrderID: order.ID.String(),
		Status:  string(order.Status),
	})
}

// GetOrder handles GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Order ID is required")
		return
	}

	qry := &query.GetOrder{OrderID: orderID}
	result, err := h.queryHandler.HandleGetOrder(r.Context(), qry)
	if err != nil {
		if err == query.ErrOrderNotFound {
			respondError(w, http.StatusNotFound, "not_found", "Order not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "query_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// CancelOrder handles POST /orders/{id}/cancel
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Order ID is required")
		return
	}

	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	cmd := &command.CancelOrder{
		OrderID: orderID,
		Reason:  req.Reason,
	}

	if err := h.commandHandler.HandleCancelOrder(r.Context(), cmd); err != nil {
		respondError(w, http.StatusInternalServerError, "cancel_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// HealthCheck handles GET /health
func (h *OrderHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
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
