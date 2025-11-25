// Package aggregate contains aggregate roots for the Order domain.
package aggregate

import (
	"fmt"
	"time"

	"github.com/your-org/ecommerce-platform/services/order/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/event"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// Order is the aggregate root for orders.
type Order struct {
	ID          valueobject.OrderID     `json:"id" bson:"_id"`
	CustomerID  valueobject.CustomerID  `json:"customer_id" bson:"customer_id"`
	Items       []*entity.OrderItem     `json:"items" bson:"items"`
	TotalAmount valueobject.Money       `json:"total_amount" bson:"total_amount"`
	Status      valueobject.OrderStatus `json:"status" bson:"status"`
	CreatedAt   time.Time               `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at" bson:"updated_at"`
	Version     int                     `json:"version" bson:"version"`

	// Uncommitted domain events
	events []event.DomainEvent
}

// NewOrder creates a new order with the given customer ID and items.
func NewOrder(customerID valueobject.CustomerID, items []*entity.OrderItem, currency valueobject.Currency) (*Order, error) {
	if customerID.IsEmpty() {
		return nil, fmt.Errorf("customer ID is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	order := &Order{
		ID:         valueobject.NewOrderID(),
		CustomerID: customerID,
		Items:      items,
		Status:     valueobject.OrderStatusCreated,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Version:    1,
		events:     make([]event.DomainEvent, 0),
	}

	// Calculate total
	if err := order.recalculateTotal(currency); err != nil {
		return nil, err
	}

	// Emit OrderCreated event
	order.addEvent(event.NewOrderCreated(
		order.ID,
		order.CustomerID,
		order.Items,
		order.TotalAmount,
	))

	return order, nil
}

// recalculateTotal recalculates the order total from items.
func (o *Order) recalculateTotal(currency valueobject.Currency) error {
	total := valueobject.Zero(currency)
	for _, item := range o.Items {
		var err error
		total, err = total.Add(item.TotalPrice())
		if err != nil {
			return fmt.Errorf("failed to calculate total: %w", err)
		}
	}
	o.TotalAmount = total
	return nil
}

// AddItem adds an item to the order.
func (o *Order) AddItem(item *entity.OrderItem) error {
	if o.Status != valueobject.OrderStatusCreated {
		return fmt.Errorf("cannot add items to order in status: %s", o.Status)
	}
	if item == nil {
		return fmt.Errorf("item cannot be nil")
	}

	o.Items = append(o.Items, item)
	if err := o.recalculateTotal(o.TotalAmount.Currency); err != nil {
		return err
	}
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// RemoveItem removes an item from the order by ID.
func (o *Order) RemoveItem(itemID string) error {
	if o.Status != valueobject.OrderStatusCreated {
		return fmt.Errorf("cannot remove items from order in status: %s", o.Status)
	}

	for i, item := range o.Items {
		if item.ID == itemID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			if err := o.recalculateTotal(o.TotalAmount.Currency); err != nil {
				return err
			}
			o.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("item not found: %s", itemID)
}

// MarkAsPaid transitions the order to PAID status.
func (o *Order) MarkAsPaid() error {
	if !o.Status.CanTransitionTo(valueobject.OrderStatusPaid) {
		return fmt.Errorf("cannot transition from %s to PAID", o.Status)
	}

	oldStatus := o.Status
	o.Status = valueobject.OrderStatusPaid
	o.UpdatedAt = time.Now().UTC()

	o.addEvent(event.NewOrderStatusChanged(o.ID, oldStatus, o.Status))
	return nil
}

// MarkAsShipped transitions the order to SHIPPED status.
func (o *Order) MarkAsShipped() error {
	if !o.Status.CanTransitionTo(valueobject.OrderStatusShipped) {
		return fmt.Errorf("cannot transition from %s to SHIPPED", o.Status)
	}

	oldStatus := o.Status
	o.Status = valueobject.OrderStatusShipped
	o.UpdatedAt = time.Now().UTC()

	o.addEvent(event.NewOrderStatusChanged(o.ID, oldStatus, o.Status))
	return nil
}

// MarkAsDelivered transitions the order to DELIVERED status.
func (o *Order) MarkAsDelivered() error {
	if !o.Status.CanTransitionTo(valueobject.OrderStatusDelivered) {
		return fmt.Errorf("cannot transition from %s to DELIVERED", o.Status)
	}

	oldStatus := o.Status
	o.Status = valueobject.OrderStatusDelivered
	o.UpdatedAt = time.Now().UTC()

	o.addEvent(event.NewOrderStatusChanged(o.ID, oldStatus, o.Status))
	return nil
}

// Complete transitions the order to COMPLETED status.
func (o *Order) Complete() error {
	if !o.Status.CanTransitionTo(valueobject.OrderStatusCompleted) {
		return fmt.Errorf("cannot transition from %s to COMPLETED", o.Status)
	}

	oldStatus := o.Status
	o.Status = valueobject.OrderStatusCompleted
	o.UpdatedAt = time.Now().UTC()

	o.addEvent(event.NewOrderCompleted(o.ID))
	o.addEvent(event.NewOrderStatusChanged(o.ID, oldStatus, o.Status))
	return nil
}

// Cancel cancels the order with a reason.
func (o *Order) Cancel(reason string) error {
	if !o.Status.CanTransitionTo(valueobject.OrderStatusCancelled) {
		return fmt.Errorf("cannot cancel order in status: %s", o.Status)
	}

	oldStatus := o.Status
	o.Status = valueobject.OrderStatusCancelled
	o.UpdatedAt = time.Now().UTC()

	o.addEvent(event.NewOrderCancelled(o.ID, reason))
	o.addEvent(event.NewOrderStatusChanged(o.ID, oldStatus, o.Status))
	return nil
}

// Events returns the uncommitted domain events.
func (o *Order) Events() []event.DomainEvent {
	return o.events
}

// ClearEvents clears the uncommitted domain events.
func (o *Order) ClearEvents() {
	o.events = make([]event.DomainEvent, 0)
}

// addEvent adds a domain event to the uncommitted events.
func (o *Order) addEvent(e event.DomainEvent) {
	o.events = append(o.events, e)
}

// IsActive returns true if the order is in an active state.
func (o *Order) IsActive() bool {
	return o.Status.IsActive()
}

// IsCancelled returns true if the order is cancelled.
func (o *Order) IsCancelled() bool {
	return o.Status == valueobject.OrderStatusCancelled
}

// IsCompleted returns true if the order is completed.
func (o *Order) IsCompleted() bool {
	return o.Status == valueobject.OrderStatusCompleted
}
