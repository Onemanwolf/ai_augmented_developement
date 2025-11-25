// Package main is the entry point for the Fulfillment service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httpapi "github.com/your-org/ecommerce-platform/services/fulfillment/internal/api/http"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/handler"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/application/saga"
	"github.com/your-org/ecommerce-platform/services/fulfillment/internal/infrastructure/persistence"
	"github.com/your-org/ecommerce-platform/shared/pkg/kafka"
	"github.com/your-org/ecommerce-platform/shared/pkg/mongodb"
	sagapkg "github.com/your-org/ecommerce-platform/shared/pkg/saga"
)

// Config holds the service configuration.
type Config struct {
	Port         string
	MongoURI     string
	MongoDBName  string
	KafkaBrokers []string
	KafkaTopic   string
}

func main() {
	cfg := loadConfig()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MongoDB client
	mongoConfig := mongodb.Config{
		URI:            cfg.MongoURI,
		Database:       cfg.MongoDBName,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    10,
	}

	mongoClient, err := mongodb.NewClient(ctx, mongoConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Close(ctx)

	// Initialize Kafka event publisher
	eventPublisher := kafka.NewEventPublisher(kafka.EventPublisherConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
	})
	defer eventPublisher.Close()
	log.Printf("Kafka event publisher initialized for topic: %s", cfg.KafkaTopic)

	// Initialize repositories
	shipmentRepo := persistence.NewMongoShipmentRepository(mongoClient)

	// Initialize SAGA event handler
	sagaHandler := saga.NewEventHandler(shipmentRepo, eventPublisher)

	// Initialize Kafka consumer for payment-events
	consumerConfig := kafka.DefaultConsumerConfig()
	consumerConfig.Brokers = cfg.KafkaBrokers
	consumerConfig.GroupID = "fulfillment-service"
	consumerConfig.Topics = []string{sagapkg.TopicPaymentEvents}
	consumer := kafka.NewConsumer(consumerConfig)
	defer consumer.Close()

	// Register event handlers
	consumer.RegisterHandler(sagapkg.EventPaymentProcessed, sagaHandler.HandlePaymentProcessed)
	consumer.RegisterHandler(sagapkg.EventPaymentRefunded, sagaHandler.HandlePaymentRefunded)
	log.Printf("Kafka consumer initialized for topics: %v", consumerConfig.Topics)

	// Start consumer in background
	go func() {
		log.Println("Starting Fulfillment SAGA event consumer...")
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Ensure indexes
	if err := shipmentRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to ensure indexes: %v", err)
	}

	// Initialize handlers with event publisher
	cmdHandler := handler.NewShipmentCommandHandler(shipmentRepo, eventPublisher)
	qryHandler := handler.NewShipmentQueryHandler(shipmentRepo)

	// Initialize HTTP handlers
	httpHandler := httpapi.NewShipmentHandler(cmdHandler, qryHandler)

	// Setup routes
	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, httpHandler)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Fulfillment service starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func loadConfig() Config {
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	return Config{
		Port:         getEnv("PORT", "8082"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:  getEnv("MONGO_DB", "fulfillment"),
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "fulfillment-events"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	fmt.Println("Fulfillment Service v1.0.0")
}
