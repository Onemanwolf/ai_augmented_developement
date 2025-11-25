// Package saga contains the SAGA orchestrator for order processing.
package saga

import (
	"time"
)

// SagaState represents the state of a saga execution.
type SagaState string

const (
	// SagaStateStarted indicates the saga has started.
	SagaStateStarted SagaState = "STARTED"
	// SagaStatePaymentPending indicates payment is being processed.
	SagaStatePaymentPending SagaState = "PAYMENT_PENDING"
	// SagaStatePaymentCompleted indicates payment was successful.
	SagaStatePaymentCompleted SagaState = "PAYMENT_COMPLETED"
	// SagaStatePaymentFailed indicates payment failed.
	SagaStatePaymentFailed SagaState = "PAYMENT_FAILED"
	// SagaStateFulfillmentPending indicates fulfillment is being processed.
	SagaStateFulfillmentPending SagaState = "FULFILLMENT_PENDING"
	// SagaStateFulfillmentCompleted indicates fulfillment was successful.
	SagaStateFulfillmentCompleted SagaState = "FULFILLMENT_COMPLETED"
	// SagaStateFulfillmentFailed indicates fulfillment failed.
	SagaStateFulfillmentFailed SagaState = "FULFILLMENT_FAILED"
	// SagaStateCompleted indicates the saga completed successfully.
	SagaStateCompleted SagaState = "COMPLETED"
	// SagaStateCompensating indicates compensation is in progress.
	SagaStateCompensating SagaState = "COMPENSATING"
	// SagaStateCompensated indicates compensation completed.
	SagaStateCompensated SagaState = "COMPENSATED"
	// SagaStateFailed indicates the saga failed.
	SagaStateFailed SagaState = "FAILED"
)

// OrderSagaData holds all data needed for the order saga.
type OrderSagaData struct {
	ID              string            `bson:"_id" json:"id"`
	OrderID         string            `bson:"order_id" json:"order_id"`
	CustomerID      string            `bson:"customer_id" json:"customer_id"`
	State           SagaState         `bson:"state" json:"state"`
	PaymentID       string            `bson:"payment_id,omitempty" json:"payment_id,omitempty"`
	ShipmentID      string            `bson:"shipment_id,omitempty" json:"shipment_id,omitempty"`
	FailureReason   string            `bson:"failure_reason,omitempty" json:"failure_reason,omitempty"`
	CompensationLog []CompensationLog `bson:"compensation_log" json:"compensation_log"`
	CreatedAt       time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at"`
	Version         int               `bson:"version" json:"version"`
}

// CompensationLog records compensation actions taken.
type CompensationLog struct {
	Action    string    `bson:"action" json:"action"`
	Status    string    `bson:"status" json:"status"`
	Error     string    `bson:"error,omitempty" json:"error,omitempty"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// NewOrderSagaData creates a new OrderSagaData.
func NewOrderSagaData(sagaID, orderID, customerID string) *OrderSagaData {
	now := time.Now().UTC()
	return &OrderSagaData{
		ID:              sagaID,
		OrderID:         orderID,
		CustomerID:      customerID,
		State:           SagaStateStarted,
		CompensationLog: make([]CompensationLog, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
}

// TransitionTo transitions the saga to a new state.
func (s *OrderSagaData) TransitionTo(newState SagaState) {
	s.State = newState
	s.UpdatedAt = time.Now().UTC()
	s.Version++
}

// SetPaymentID sets the payment ID.
func (s *OrderSagaData) SetPaymentID(paymentID string) {
	s.PaymentID = paymentID
	s.UpdatedAt = time.Now().UTC()
}

// SetShipmentID sets the shipment ID.
func (s *OrderSagaData) SetShipmentID(shipmentID string) {
	s.ShipmentID = shipmentID
	s.UpdatedAt = time.Now().UTC()
}

// SetFailure records a failure.
func (s *OrderSagaData) SetFailure(reason string) {
	s.FailureReason = reason
	s.UpdatedAt = time.Now().UTC()
}

// AddCompensationLog adds a compensation log entry.
func (s *OrderSagaData) AddCompensationLog(action, status, err string) {
	s.CompensationLog = append(s.CompensationLog, CompensationLog{
		Action:    action,
		Status:    status,
		Error:     err,
		Timestamp: time.Now().UTC(),
	})
	s.UpdatedAt = time.Now().UTC()
}

// IsCompleted checks if the saga is in a terminal completed state.
func (s *OrderSagaData) IsCompleted() bool {
	return s.State == SagaStateCompleted
}

// IsFailed checks if the saga is in a terminal failed state.
func (s *OrderSagaData) IsFailed() bool {
	return s.State == SagaStateFailed || s.State == SagaStateCompensated
}

// IsTerminal checks if the saga is in any terminal state.
func (s *OrderSagaData) IsTerminal() bool {
	return s.IsCompleted() || s.IsFailed()
}

// NeedsCompensation checks if the saga needs compensation.
func (s *OrderSagaData) NeedsCompensation() bool {
	return s.State == SagaStatePaymentFailed || s.State == SagaStateFulfillmentFailed
}
