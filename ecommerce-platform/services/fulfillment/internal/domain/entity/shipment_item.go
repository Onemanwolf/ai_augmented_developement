// Package entity contains entities for the Fulfillment domain.
package entity

import (
	"fmt"

	"github.com/google/uuid"
)

// ShipmentItem represents an item in a shipment.
type ShipmentItem struct {
	ID        string `bson:"id"`
	ProductID string `bson:"product_id"`
	Name      string `bson:"name"`
	Quantity  int    `bson:"quantity"`
	Weight    Weight `bson:"weight"` // Weight in grams
}

// Weight represents weight with unit.
type Weight struct {
	Value int    `json:"value" bson:"value"` // Weight in grams
	Unit  string `json:"unit" bson:"unit"`   // Always "g" for grams
}

// NewWeight creates a new Weight value.
func NewWeight(grams int) (Weight, error) {
	if grams < 0 {
		return Weight{}, fmt.Errorf("weight cannot be negative")
	}
	return Weight{Value: grams, Unit: "g"}, nil
}

// NewShipmentItem creates a new ShipmentItem.
func NewShipmentItem(productID, name string, quantity int, weight Weight) (*ShipmentItem, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	return &ShipmentItem{
		ID:        uuid.New().String(),
		ProductID: productID,
		Name:      name,
		Quantity:  quantity,
		Weight:    weight,
	}, nil
}

// TotalWeight returns the total weight of all items.
func (i *ShipmentItem) TotalWeight() int {
	return i.Weight.Value * i.Quantity
}

// Equals checks equality with another ShipmentItem.
func (i *ShipmentItem) Equals(other *ShipmentItem) bool {
	if other == nil {
		return false
	}
	return i.ID == other.ID
}
