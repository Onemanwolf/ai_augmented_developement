// Package http provides HTTP handlers for the Payment service API.
package http

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/ecommerce-platform/services/payment/internal/application/command"
	"github.com/your-org/ecommerce-platform/services/payment/internal/application/handler"
	"github.com/your-org/ecommerce-platform/services/payment/internal/application/query"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
)

// PaymentHandler handles HTTP requests for payments.
type PaymentHandler struct {
	commandHandler *handler.PaymentCommandHandler
	queryHandler   *handler.PaymentQueryHandler
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(cmdHandler *handler.PaymentCommandHandler, qryHandler *handler.PaymentQueryHandler) *PaymentHandler {
	return &PaymentHandler{
		commandHandler: cmdHandler,
		queryHandler:   qryHandler,
	}
}

// ProcessPaymentRequest represents the request body for processing a payment.
type ProcessPaymentRequest struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
}

// ProcessPaymentResponse represents the response for processing a payment.
type ProcessPaymentResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

// RefundPaymentRequest represents the request body for refunding a payment.
type RefundPaymentRequest struct {
	Reason string `json:"reason"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ProcessPayment handles POST /payments
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	cmd := &command.ProcessPayment{
		OrderID:  req.OrderID,
		Amount:   req.Amount,
		Currency: valueobject.Currency(req.Currency),
		Method:   valueobject.PaymentMethod(req.PaymentMethod),
	}

	payment, err := h.commandHandler.HandleProcessPayment(r.Context(), cmd)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "process_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, ProcessPaymentResponse{
		PaymentID: payment.ID.String(),
		Status:    string(payment.Status),
	})
}

// GetPayment handles GET /payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Payment ID is required")
		return
	}

	qry := &query.GetPayment{PaymentID: paymentID}
	result, err := h.queryHandler.HandleGetPayment(r.Context(), qry)
	if err != nil {
		if err == query.ErrPaymentNotFound {
			respondError(w, http.StatusNotFound, "not_found", "Payment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "query_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// RefundPayment handles POST /payments/{id}/refund
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		respondError(w, http.StatusBadRequest, "invalid_id", "Payment ID is required")
		return
	}

	var req RefundPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	cmd := &command.RefundPayment{
		PaymentID: paymentID,
		Reason:    req.Reason,
	}

	if err := h.commandHandler.HandleRefundPayment(r.Context(), cmd); err != nil {
		respondError(w, http.StatusInternalServerError, "refund_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "refunded"})
}

// HealthCheck handles GET /health
func (h *PaymentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
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
