// Package query contains query definitions for the Order service.
package query

import (
	"time"

	"github.com/your-org/ecommerce-platform/services/order/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderDTO represents an order in query responses.
type OrderDTO struct {
	ID          string         `json:"id"`
	CustomerID  string         `json:"customer_id"`
	Items       []OrderItemDTO `json:"items"`
	TotalAmount MoneyDTO       `json:"total_amount"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// OrderItemDTO represents an order item in query responses.
type OrderItemDTO struct {
	ID        string   `json:"id"`
	ProductID string   `json:"product_id"`
	Name      string   `json:"name"`
	Quantity  int      `json:"quantity"`
	UnitPrice MoneyDTO `json:"unit_price"`
}

// MoneyDTO represents money in query responses.
type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// FromAggregate converts an Order aggregate to OrderDTO.
func FromAggregate(order *aggregate.Order) *OrderDTO {
	if order == nil {
		return nil
	}

	items := make([]OrderItemDTO, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemDTO{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: MoneyDTO{
				Amount:   item.UnitPrice.Amount,
				Currency: string(item.UnitPrice.Currency),
			},
		}
	}

	return &OrderDTO{
		ID:         order.ID.String(),
		CustomerID: order.CustomerID.String(),
		Items:      items,
		TotalAmount: MoneyDTO{
			Amount:   order.TotalAmount.Amount,
			Currency: string(order.TotalAmount.Currency),
		},
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}

// FromAggregateList converts a list of Order aggregates to OrderDTOs.
func FromAggregateList(orders []*aggregate.Order) []*OrderDTO {
	dtos := make([]*OrderDTO, len(orders))
	for i, order := range orders {
		dtos[i] = FromAggregate(order)
	}
	return dtos
}

// MoneyFromValueObject converts a Money value object to MoneyDTO.
func MoneyFromValueObject(m valueobject.Money) MoneyDTO {
	return MoneyDTO{
		Amount:   m.Amount,
		Currency: string(m.Currency),
	}
}
