package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
)

// OrderCreated is emitted when a new order is created.
type OrderCreated struct {
	BaseEvent
	CustomerID  valueobject.CustomerID `json:"customer_id"`
	Items       []*OrderItemData       `json:"items"`
	TotalAmount valueobject.Money      `json:"total_amount"`
}

// OrderItemData represents item data in the event.
type OrderItemData struct {
	ProductID string            `json:"product_id"`
	Name      string            `json:"name"`
	Quantity  int               `json:"quantity"`
	UnitPrice valueobject.Money `json:"unit_price"`
}

// NewOrderCreated creates a new OrderCreated event.
func NewOrderCreated(
	orderID valueobject.OrderID,
	customerID valueobject.CustomerID,
	items []*entity.OrderItem,
	totalAmount valueobject.Money,
) *OrderCreated {
	itemData := make([]*OrderItemData, len(items))
	for i, item := range items {
		itemData[i] = &OrderItemData{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	return &OrderCreated{
		BaseEvent: BaseEvent{
			ID:           uuid.New().String(),
			Type:         "OrderCreated",
			AggregateIDV: orderID.String(),
			OccurredAtV:  time.Now().UTC(),
		},
		CustomerID:  customerID,
		Items:       itemData,
		TotalAmount: totalAmount,
	}
}
