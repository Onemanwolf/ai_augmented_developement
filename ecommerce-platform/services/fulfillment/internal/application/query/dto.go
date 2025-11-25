// Package query contains query definitions for the Fulfillment service.
package query

import (
	"time"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/aggregate"
)

// ShipmentDTO represents a shipment in query responses.
type ShipmentDTO struct {
	ID                string            `json:"id"`
	OrderID           string            `json:"order_id"`
	CustomerID        string            `json:"customer_id"`
	Items             []ShipmentItemDTO `json:"items"`
	ShippingAddress   AddressDTO        `json:"shipping_address"`
	Carrier           string            `json:"carrier"`
	TrackingNumber    string            `json:"tracking_number,omitempty"`
	Status            string            `json:"status"`
	FailureReason     string            `json:"failure_reason,omitempty"`
	EstimatedDelivery *time.Time        `json:"estimated_delivery,omitempty"`
	ActualDelivery    *time.Time        `json:"actual_delivery,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	ShippedAt         *time.Time        `json:"shipped_at,omitempty"`
}

// ShipmentItemDTO represents a shipment item in query responses.
type ShipmentItemDTO struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	Weight    int    `json:"weight_grams"`
}

// AddressDTO represents an address in query responses.
type AddressDTO struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// FromAggregate converts a Shipment aggregate to ShipmentDTO.
func FromAggregate(shipment *aggregate.Shipment) *ShipmentDTO {
	if shipment == nil {
		return nil
	}

	items := make([]ShipmentItemDTO, len(shipment.Items))
	for i, item := range shipment.Items {
		items[i] = ShipmentItemDTO{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			Weight:    item.Weight.Value,
		}
	}

	return &ShipmentDTO{
		ID:         shipment.ID.String(),
		OrderID:    shipment.OrderID,
		CustomerID: shipment.CustomerID,
		Items:      items,
		ShippingAddress: AddressDTO{
			Street:     shipment.ShippingAddress.Street,
			City:       shipment.ShippingAddress.City,
			State:      shipment.ShippingAddress.State,
			PostalCode: shipment.ShippingAddress.PostalCode,
			Country:    shipment.ShippingAddress.Country,
		},
		Carrier:           string(shipment.Carrier),
		TrackingNumber:    shipment.TrackingNumber,
		Status:            string(shipment.Status),
		FailureReason:     shipment.FailureReason,
		EstimatedDelivery: shipment.EstimatedDelivery,
		ActualDelivery:    shipment.ActualDelivery,
		CreatedAt:         shipment.CreatedAt,
		UpdatedAt:         shipment.UpdatedAt,
		ShippedAt:         shipment.ShippedAt,
	}
}

// FromAggregateList converts a list of Shipment aggregates to ShipmentDTOs.
func FromAggregateList(shipments []*aggregate.Shipment) []*ShipmentDTO {
	dtos := make([]*ShipmentDTO, len(shipments))
	for i, shipment := range shipments {
		dtos[i] = FromAggregate(shipment)
	}
	return dtos
}
