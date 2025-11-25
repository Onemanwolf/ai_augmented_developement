// Package persistence provides MongoDB implementations of domain repositories.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/payment/internal/domain/valueobject"
	"github.com/your-org/ecommerce-platform/shared/pkg/mongodb"
	"github.com/your-org/ecommerce-platform/shared/pkg/outbox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	paymentsCollection = "payments"
	outboxCollection   = "outbox"
)

// MongoPaymentRepository implements repository.PaymentRepository using MongoDB.
type MongoPaymentRepository struct {
	client     *mongodb.Client
	collection *mongo.Collection
	outbox     *mongo.Collection
}

// NewMongoPaymentRepository creates a new MongoDB payment repository.
func NewMongoPaymentRepository(client *mongodb.Client) *MongoPaymentRepository {
	return &MongoPaymentRepository{
		client:     client,
		collection: client.Collection(paymentsCollection),
		outbox:     client.Collection(outboxCollection),
	}
}

// paymentDocument is the MongoDB document representation of a Payment.
type paymentDocument struct {
	ID            string        `bson:"_id"`
	OrderID       string        `bson:"order_id"`
	Amount        moneyDocument `bson:"amount"`
	Status        string        `bson:"status"`
	PaymentMethod string        `bson:"payment_method"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
	Version       int           `bson:"version"`
}

type moneyDocument struct {
	Amount   int64  `bson:"amount"`
	Currency string `bson:"currency"`
}

// Save saves a payment and its domain events atomically.
func (r *MongoPaymentRepository) Save(ctx context.Context, payment *aggregate.Payment) error {
	return r.client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		doc := toPaymentDocument(payment)

		opts := options.Replace().SetUpsert(true)
		_, err := r.collection.ReplaceOne(sc, bson.M{"_id": doc.ID}, doc, opts)
		if err != nil {
			return fmt.Errorf("failed to save payment: %w", err)
		}

		// Save domain events to outbox
		for _, event := range payment.GetEvents() {
			entry, err := outbox.NewEntryFromRaw(
				event.AggregateID(),
				"Payment",
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

		payment.ClearEvents()
		return nil
	})
}

// FindByID finds a payment by its ID.
func (r *MongoPaymentRepository) FindByID(ctx context.Context, id valueobject.PaymentID) (*aggregate.Payment, error) {
	var doc paymentDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	return toPayment(&doc), nil
}

// FindByOrderID finds a payment by order ID.
func (r *MongoPaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*aggregate.Payment, error) {
	var doc paymentDocument
	err := r.collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}
	return toPayment(&doc), nil
}

// toPaymentDocument converts a Payment aggregate to a MongoDB document.
func toPaymentDocument(payment *aggregate.Payment) *paymentDocument {
	return &paymentDocument{
		ID:      payment.ID.String(),
		OrderID: payment.OrderID,
		Amount: moneyDocument{
			Amount:   payment.Amount.Amount,
			Currency: string(payment.Amount.Currency),
		},
		Status:        string(payment.Status),
		PaymentMethod: string(payment.Method),
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
		Version:       payment.Version,
	}
}

// toPayment converts a MongoDB document to a Payment aggregate.
func toPayment(doc *paymentDocument) *aggregate.Payment {
	paymentID, _ := valueobject.ParsePaymentID(doc.ID)

	return &aggregate.Payment{
		ID:      paymentID,
		OrderID: doc.OrderID,
		Amount: valueobject.Money{
			Amount:   doc.Amount.Amount,
			Currency: valueobject.Currency(doc.Amount.Currency),
		},
		Status: valueobject.PaymentStatus(doc.Status),
		Method: valueobject.PaymentMethod(doc.PaymentMethod),
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
		Version:       doc.Version,
	}
}

// EnsureIndexes creates the necessary indexes for the payments collection.
func (r *MongoPaymentRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "order_id", Value: 1}},
			Options: options.Index().SetName("idx_order_id").SetUnique(true),
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

// FindByCustomerID finds all payments for a customer.
func (r *MongoPaymentRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*aggregate.Payment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*aggregate.Payment
	for cursor.Next(ctx) {
		var doc paymentDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode payment: %w", err)
		}
		payments = append(payments, toPayment(&doc))
	}
	return payments, nil
}

// FindByStatus finds all payments with the given status.
func (r *MongoPaymentRepository) FindByStatus(ctx context.Context, status valueobject.PaymentStatus) ([]*aggregate.Payment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": string(status)})
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*aggregate.Payment
	for cursor.Next(ctx) {
		var doc paymentDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode payment: %w", err)
		}
		payments = append(payments, toPayment(&doc))
	}
	return payments, nil
}

// Delete removes a payment by its ID.
func (r *MongoPaymentRepository) Delete(ctx context.Context, id valueobject.PaymentID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}
	if result.DeletedCount == 0 {
		return repository.ErrPaymentNotFound
	}
	return nil
}

// Exists checks if a payment exists.
func (r *MongoPaymentRepository) Exists(ctx context.Context, id valueobject.PaymentID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return false, fmt.Errorf("failed to check payment existence: %w", err)
	}
	return count > 0, nil
}
