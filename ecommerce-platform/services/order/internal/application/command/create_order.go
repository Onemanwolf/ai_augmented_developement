// Package command contains command definitions for the Order service.
package command

import "github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"

// CreateOrder represents a command to create a new order.
type CreateOrder struct {
	CustomerID string             `json:"customer_id"`
	Items      []CreateOrderItem  `json:"items"`
	Currency   valueobject.Currency `json:"currency"`
}

// CreateOrderItem represents an item in the create order command.
type CreateOrderItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"` // In smallest currency unit (cents)
}

// Validate validates the command.
func (c *CreateOrder) Validate() error {
	if c.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	if len(c.Items) == 0 {
		return ErrItemsRequired
	}
	if !c.Currency.IsValid() {
		return ErrInvalidCurrency
	}
	for _, item := range c.Items {
		if item.ProductID == "" {
			return ErrProductIDRequired
		}
		if item.Name == "" {
			return ErrProductNameRequired
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.UnitPrice <= 0 {
			return ErrInvalidPrice
		}
	}
	return nil
}
