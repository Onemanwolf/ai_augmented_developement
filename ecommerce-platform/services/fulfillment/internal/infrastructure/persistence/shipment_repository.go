// Package persistence provides MongoDB implementations of domain repositories.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/aggregate"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/entity"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/repository"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/domain/valueobject"
	"github.com/your-org/ecommerce-platform/shared/pkg/mongodb"
	"github.com/your-org/ecommerce-platform/shared/pkg/outbox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	shipmentsCollection = "shipments"
	outboxCollection    = "outbox"
)

// MongoShipmentRepository implements repository.ShipmentRepository using MongoDB.
type MongoShipmentRepository struct {
	client     *mongodb.Client
	collection *mongo.Collection
	outbox     *mongo.Collection
}

// NewMongoShipmentRepository creates a new MongoDB shipment repository.
func NewMongoShipmentRepository(client *mongodb.Client) *MongoShipmentRepository {
	return &MongoShipmentRepository{
		client:     client,
		collection: client.Collection(shipmentsCollection),
		outbox:     client.Collection(outboxCollection),
	}
}

// shipmentDocument is the MongoDB document representation of a Shipment.
type shipmentDocument struct {
	ID             string                 `bson:"_id"`
	OrderID        string                 `bson:"order_id"`
	Items          []shipmentItemDocument `bson:"items"`
	TrackingNumber string                 `bson:"tracking_number"`
	Carrier        string                 `bson:"carrier"`
	Status         string                 `bson:"status"`
	Address        addressDocument        `bson:"address"`
	CreatedAt      time.Time              `bson:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at"`
	ShippedAt      *time.Time             `bson:"shipped_at,omitempty"`
	DeliveredAt    *time.Time             `bson:"delivered_at,omitempty"`
	Version        int                    `bson:"version"`
}

type shipmentItemDocument struct {
	ProductID string `bson:"product_id"`
	Name      string `bson:"name"`
	Quantity  int    `bson:"quantity"`
}

type addressDocument struct {
	Street     string `bson:"street"`
	City       string `bson:"city"`
	State      string `bson:"state"`
	PostalCode string `bson:"postal_code"`
	Country    string `bson:"country"`
}

// Save saves a shipment and its domain events atomically.
func (r *MongoShipmentRepository) Save(ctx context.Context, shipment *aggregate.Shipment) error {
	return r.client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		doc := toShipmentDocument(shipment)

		opts := options.Replace().SetUpsert(true)
		_, err := r.collection.ReplaceOne(sc, bson.M{"_id": doc.ID}, doc, opts)
		if err != nil {
			return fmt.Errorf("failed to save shipment: %w", err)
		}

		// Save domain events to outbox
		for _, event := range shipment.GetEvents() {
			entry, err := outbox.NewEntryFromRaw(
				event.AggregateID(),
				"Shipment",
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

		shipment.ClearEvents()
		return nil
	})
}

// FindByID finds a shipment by its ID.
func (r *MongoShipmentRepository) FindByID(ctx context.Context, id valueobject.ShipmentID) (*aggregate.Shipment, error) {
	var doc shipmentDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrShipmentNotFound
		}
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}
	return toShipment(&doc), nil
}

// FindByOrderID finds a shipment by order ID.
func (r *MongoShipmentRepository) FindByOrderID(ctx context.Context, orderID string) (*aggregate.Shipment, error) {
	var doc shipmentDocument
	err := r.collection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrShipmentNotFound
		}
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}
	return toShipment(&doc), nil
}

// FindByTrackingNumber finds a shipment by tracking number.
func (r *MongoShipmentRepository) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*aggregate.Shipment, error) {
	var doc shipmentDocument
	err := r.collection.FindOne(ctx, bson.M{"tracking_number": trackingNumber}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repository.ErrShipmentNotFound
		}
		return nil, fmt.Errorf("failed to find shipment: %w", err)
	}
	return toShipment(&doc), nil
}

// toShipmentDocument converts a Shipment aggregate to a MongoDB document.
func toShipmentDocument(shipment *aggregate.Shipment) *shipmentDocument {
	items := make([]shipmentItemDocument, len(shipment.Items))
	for i, item := range shipment.Items {
		items[i] = shipmentItemDocument{
			ProductID: item.ProductID,
			Name:      item.Name,
			Quantity:  item.Quantity,
		}
	}

	doc := &shipmentDocument{
		ID:             shipment.ID.String(),
		OrderID:        shipment.OrderID,
		Items:          items,
		TrackingNumber: shipment.TrackingNumber,
		Carrier:        string(shipment.Carrier),
		Status:         string(shipment.Status),
		Address: addressDocument{
			Street:     shipment.ShippingAddress.Street,
			City:       shipment.ShippingAddress.City,
			State:      shipment.ShippingAddress.State,
			PostalCode: shipment.ShippingAddress.PostalCode,
			Country:    shipment.ShippingAddress.Country,
		},
		CreatedAt: shipment.CreatedAt,
		UpdatedAt: shipment.UpdatedAt,
		ShippedAt: shipment.ShippedAt,
		DeliveredAt: shipment.ActualDelivery,
		Version:   shipment.Version,
	}

	return doc
}

// toShipment converts a MongoDB document to a Shipment aggregate.
func toShipment(doc *shipmentDocument) *aggregate.Shipment {
	items := make([]*entity.ShipmentItem, len(doc.Items))
	for i, itemDoc := range doc.Items {
		items[i] = &entity.ShipmentItem{
			ProductID: itemDoc.ProductID,
			Name:      itemDoc.Name,
			Quantity:  itemDoc.Quantity,
		}
	}

	shipmentID, _ := valueobject.ParseShipmentID(doc.ID)

	shipment := &aggregate.Shipment{
		ID:             shipmentID,
		OrderID:        doc.OrderID,
		Items:          items,
		TrackingNumber: doc.TrackingNumber,
		Carrier:        valueobject.Carrier(doc.Carrier),
		Status:         valueobject.ShipmentStatus(doc.Status),
		ShippingAddress: valueobject.Address{
			Street:     doc.Address.Street,
			City:       doc.Address.City,
			State:      doc.Address.State,
			PostalCode: doc.Address.PostalCode,
			Country:    doc.Address.Country,
		},
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
		ShippedAt:      doc.ShippedAt,
		ActualDelivery: doc.DeliveredAt,
		Version:        doc.Version,
	}

	return shipment
}

// EnsureIndexes creates the necessary indexes for the shipments collection.
func (r *MongoShipmentRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "order_id", Value: 1}},
			Options: options.Index().SetName("idx_order_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "tracking_number", Value: 1}},
			Options: options.Index().SetName("idx_tracking_number").SetSparse(true),
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

// FindByCustomerID finds all shipments for a customer.
func (r *MongoShipmentRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*aggregate.Shipment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return nil, fmt.Errorf("failed to find shipments: %w", err)
	}
	defer cursor.Close(ctx)

	var shipments []*aggregate.Shipment
	for cursor.Next(ctx) {
		var doc shipmentDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode shipment: %w", err)
		}
		shipments = append(shipments, toShipment(&doc))
	}
	return shipments, nil
}

// FindByStatus finds all shipments with the given status.
func (r *MongoShipmentRepository) FindByStatus(ctx context.Context, status valueobject.ShipmentStatus) ([]*aggregate.Shipment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": string(status)})
	if err != nil {
		return nil, fmt.Errorf("failed to find shipments: %w", err)
	}
	defer cursor.Close(ctx)

	var shipments []*aggregate.Shipment
	for cursor.Next(ctx) {
		var doc shipmentDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode shipment: %w", err)
		}
		shipments = append(shipments, toShipment(&doc))
	}
	return shipments, nil
}

// Delete removes a shipment by its ID.
func (r *MongoShipmentRepository) Delete(ctx context.Context, id valueobject.ShipmentID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return fmt.Errorf("failed to delete shipment: %w", err)
	}
	if result.DeletedCount == 0 {
		return repository.ErrShipmentNotFound
	}
	return nil
}

// Exists checks if a shipment exists.
func (r *MongoShipmentRepository) Exists(ctx context.Context, id valueobject.ShipmentID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}
