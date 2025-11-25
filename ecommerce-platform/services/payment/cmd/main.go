// Package main is the entry point for the Payment service.
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

	httpapi "github.com/your-org/ecommerce-platform/services/payment/internal/api/http"
	"github.com/your-org/ecommerce-platform/services/payment/internal/application/handler"
	"github.com/your-org/ecommerce-platform/services/payment/internal/application/saga"
	"github.com/your-org/ecommerce-platform/services/payment/internal/infrastructure/gateway"
	"github.com/your-org/ecommerce-platform/services/payment/internal/infrastructure/persistence"
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

	// Initialize payment gateway (mock for now)
	paymentGateway := gateway.NewMockPaymentGateway()

	// Initialize repositories
	paymentRepo := persistence.NewMongoPaymentRepository(mongoClient)

	// Initialize SAGA event handler
	sagaHandler := saga.NewEventHandler(paymentRepo, paymentGateway, eventPublisher)

	// Initialize Kafka consumer for order-events
	consumerConfig := kafka.DefaultConsumerConfig()
	consumerConfig.Brokers = cfg.KafkaBrokers
	consumerConfig.GroupID = "payment-service"
	consumerConfig.Topics = []string{sagapkg.TopicOrderEvents, sagapkg.TopicFulfillmentEvents}
	consumer := kafka.NewConsumer(consumerConfig)
	defer consumer.Close()

	// Register event handlers
	consumer.RegisterHandler(sagapkg.EventOrderCreated, sagaHandler.HandleOrderCreated)
	consumer.RegisterHandler(sagapkg.EventShipmentFailed, sagaHandler.HandleShipmentFailed)
	log.Printf("Kafka consumer initialized for topics: %v", consumerConfig.Topics)

	// Start consumer in background
	go func() {
		log.Println("Starting Payment SAGA event consumer...")
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Ensure indexes
	if err := paymentRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to ensure indexes: %v", err)
	}

	// Initialize handlers with event publisher and gateway
	cmdHandler := handler.NewPaymentCommandHandler(paymentRepo, eventPublisher, paymentGateway)
	qryHandler := handler.NewPaymentQueryHandler(paymentRepo)

	// Initialize HTTP handlers
	httpHandler := httpapi.NewPaymentHandler(cmdHandler, qryHandler)

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
		log.Printf("Payment service starting on port %s", cfg.Port)
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
		Port:         getEnv("PORT", "8081"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:  getEnv("MONGO_DB", "payments"),
		KafkaBrokers: strings.Split(brokers, ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "payment-events"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	fmt.Println("Payment Service v1.0.0")
}
