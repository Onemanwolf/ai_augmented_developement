// Package entity contains domain entities for the Order service.
package entity

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderItem represents a line item in an order.
type OrderItem struct {
	ID        string           `json:"id" bson:"id"`
	ProductID string           `json:"product_id" bson:"product_id"`
	Name      string           `json:"name" bson:"name"`
	Quantity  int              `json:"quantity" bson:"quantity"`
	UnitPrice valueobject.Money `json:"unit_price" bson:"unit_price"`
}

// NewOrderItem creates a new order item.
func NewOrderItem(productID, name string, quantity int, unitPrice valueobject.Money) (*OrderItem, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("product name cannot be empty")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive: %d", quantity)
	}
	if !unitPrice.IsPositive() {
		return nil, fmt.Errorf("unit price must be positive")
	}

	return &OrderItem{
		ID:        uuid.New().String(),
		ProductID: productID,
		Name:      name,
		Quantity:  quantity,
		UnitPrice: unitPrice,
	}, nil
}

// TotalPrice calculates the total price for this item.
func (i *OrderItem) TotalPrice() valueobject.Money {
	return i.UnitPrice.Multiply(int64(i.Quantity))
}

// UpdateQuantity updates the quantity of the item.
func (i *OrderItem) UpdateQuantity(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive: %d", quantity)
	}
	i.Quantity = quantity
	return nil
}

// Equals checks if two order items are equal by ID.
func (i *OrderItem) Equals(other *OrderItem) bool {
	if other == nil {
		return false
	}
	return i.ID == other.ID
}
