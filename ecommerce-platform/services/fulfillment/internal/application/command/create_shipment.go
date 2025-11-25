// Package command contains command definitions for the Fulfillment service.
package command

import "github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"

// CreateShipment represents a command to create a new shipment.
type CreateShipment struct {
	OrderID         string                    `json:"order_id"`
	CustomerID      string                    `json:"customer_id"`
	Items           []CreateShipmentItem      `json:"items"`
	ShippingAddress ShippingAddressData       `json:"shipping_address"`
	Carrier         valueobject.Carrier       `json:"carrier"`
}

// CreateShipmentItem represents an item in the create shipment command.
type CreateShipmentItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	WeightGrams int  `json:"weight_grams"`
}

// ShippingAddressData represents shipping address data.
type ShippingAddressData struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Validate validates the command.
func (c *CreateShipment) Validate() error {
	if c.OrderID == "" {
		return ErrOrderIDRequired
	}
	if c.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	if len(c.Items) == 0 {
		return ErrItemsRequired
	}
	if !c.Carrier.IsValid() {
		return ErrInvalidCarrier
	}
	if c.ShippingAddress.Street == "" || c.ShippingAddress.City == "" ||
		c.ShippingAddress.State == "" || c.ShippingAddress.PostalCode == "" ||
		c.ShippingAddress.Country == "" {
		return ErrAddressRequired
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
	}
	return nil
}
