// Package persistence provides MongoDB implementations of domain repositories.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/ecommerce-platform/services/order/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/order/internal/domain/valueobject"
	"github.com/your-org/ecommerce-platform/shared/pkg/mongodb"
	"github.com/your-org/ecommerce-platform/shared/pkg/outbox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ordersCollection = "orders"
	outboxCollection = "outbox"
)

// MongoOrderRepository implements repository.OrderRepository using MongoDB.
type MongoOrderRepository struct {
	client     *mongodb.Client
	collection *mongo.Collection
	outbox     *mongo.Collection
}

// NewMongoOrderRepository creates a new MongoDB order repository.
func NewMongoOrderRepository(client *mongodb.Client) *MongoOrderRepository {
	repo := &MongoOrderRepository{
		client:     client,
		collection: client.Collection(ordersCollection),
		outbox:     client.Collection(outboxCollection),
	}
	return repo
}

// orderDocument is the MongoDB document representation of an Order.
type orderDocument struct {
	ID          string              `bson:"_id"`
	CustomerID  string              `bson:"customer_id"`
	Items       []orderItemDocument `bson:"items"`
	TotalAmount moneyDocument       `bson:"total_amount"`
	Status      string              `bson:"status"`
	CreatedAt   time.Time           `bson:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at"`
	Version     int                 `bson:"version"`
}

type orderItemDocument struct {
	ID        string        `bson:"id"`
	ProductID string        `bson:"product_id"`
	Name      string        `bson:"name"`
	Quantity  int           `bson:"quantity"`
	UnitPrice moneyDocument `bson:"unit_price"`
}

type moneyDocument struct {
	Amount   int64  `bson:"amount"`
	Currency string `bson:"currency"`
}

// Save saves an order and its domain events atomically.
func (r *MongoOrderRepository) Save(ctx context.Context, order *aggregate.Order) error {
	return r.client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		doc := toOrderDocument(order)

		opts := options.Replace().SetUpsert(true)
		_, err := r.collection.ReplaceOne(sc, bson.M{"_id": doc.ID}, doc, opts)
		if err != nil {
			return fmt.Errorf("failed to save order: %w", err)
		}

		// Save domain events to outbox
		for _, event := range order.Events() {
			entry, err := outbox.NewEntryFromRaw(
				event.AggregateID(),
				"Order",
				event.EventType(),
				event,
			)
			if err != nil {
				return fmt.Errorf("failed to create outbox entry: %w", err)
			}
			outboxDoc := bson.M{
				"_id":            entry.ID,
				"aggregate_id":   entry.AggregateID,
				"aggregate_type": entry.AggregateType,
				"event_type":     entry.EventType,
				"payload":        entry.Payload,
				"created_at":     entry.CreatedAt,
				"published":      entry.Published,
			}
			_, err = r.outbox.InsertOne(sc, outboxDoc)
			if err != nil {
				return fmt.Errorf("failed to save outbox entry: %w", err)
			}
		}

		order.ClearEvents()
		return nil
	})
}

// FindByID finds an order by its ID.
func (r *MongoOrderRepository) FindByID(ctx context.Context, id valueobject.OrderID) (*aggregate.Order, error) {
	var doc orderDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}
	return toOrder(&doc), nil
}

// FindByCustomerID finds all orders for a customer.
func (r *MongoOrderRepository) FindByCustomerID(ctx context.Context, customerID valueobject.CustomerID) ([]*aggregate.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"customer_id": customerID.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*aggregate.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode order: %w", err)
		}
		orders = append(orders, toOrder(&doc))
	}

	return orders, nil
}

// toOrderDocument converts an Order aggregate to a MongoDB document.
func toOrderDocument(order *aggregate.Order) *orderDocument {
	items := make([]orderItemDocument, len(order.Items))
	for i, item := range order.Items {
		items[i] = orderItemDocument{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: moneyDocument{
				Amount:   item.UnitPrice.Amount,
				Currency: string(item.UnitPrice.Currency),
			},
		}
	}

	return &orderDocument{
		ID:         order.ID.String(),
		CustomerID: order.CustomerID.String(),
		Items:      items,
		TotalAmount: moneyDocument{
			Amount:   order.TotalAmount.Amount,
			Currency: string(order.TotalAmount.Currency),
		},
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
		Version:   order.Version,
	}
}

// toOrder converts a MongoDB document to an Order aggregate.
func toOrder(doc *orderDocument) *aggregate.Order {
	items := make([]*entity.OrderItem, len(doc.Items))
	for i, itemDoc := range doc.Items {
		items[i] = &entity.OrderItem{
			ID:        itemDoc.ID,
			ProductID: itemDoc.ProductID,
			Name:      itemDoc.Name,
			Quantity:  itemDoc.Quantity,
			UnitPrice: valueobject.Money{
				Amount:   itemDoc.UnitPrice.Amount,
				Currency: valueobject.Currency(itemDoc.UnitPrice.Currency),
			},
		}
	}

	orderID, _ := valueobject.ParseOrderID(doc.ID)
	customerID, _ := valueobject.ParseCustomerID(doc.CustomerID)

	return &aggregate.Order{
		ID:         orderID,
		CustomerID: customerID,
		Items:      items,
		TotalAmount: valueobject.Money{
			Amount:   doc.TotalAmount.Amount,
			Currency: valueobject.Currency(doc.TotalAmount.Currency),
		},
		Status:    valueobject.OrderStatus(doc.Status),
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
		Version:   doc.Version,
	}
}

// EnsureIndexes creates the necessary indexes for the orders collection.
func (r *MongoOrderRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "customer_id", Value: 1}},
			Options: options.Index().SetName("idx_customer_id"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_status"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}

// FindByStatus finds all orders with the given status.
func (r *MongoOrderRepository) FindByStatus(ctx context.Context, status valueobject.OrderStatus) ([]*aggregate.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": string(status)})
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*aggregate.Order
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode order: %w", err)
		}
		orders = append(orders, toOrder(&doc))
	}

	return orders, nil
}

// Delete removes an order by its ID.
func (r *MongoOrderRepository) Delete(ctx context.Context, id valueobject.OrderID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	if result.DeletedCount == 0 {
		return repository.ErrOrderNotFound
	}
	return nil
}

// Exists checks if an order exists.
func (r *MongoOrderRepository) Exists(ctx context.Context, id valueobject.OrderID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return false, fmt.Errorf("failed to check order existence: %w", err)
	}
	return count > 0, nil
}
