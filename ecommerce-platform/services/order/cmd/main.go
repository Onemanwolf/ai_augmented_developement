// Package main is the entry point for the Order service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/ecommerce-platform/services/order/internal/application/handler"
	httpapi "github.com/your-org/ecommerce-platform/services/order/internal/api/http"
	"github.com/your-org/ecommerce-platform/services/order/internal/infrastructure/persistence"
	"github.com/your-org/ecommerce-platform/shared/pkg/mongodb"
)

// Config holds the service configuration.
type Config struct {
	Port        string
	MongoURI    string
	MongoDBName string
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

	// Initialize repositories
	orderRepo := persistence.NewMongoOrderRepository(mongoClient)

	// Ensure indexes
	if err := orderRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to ensure indexes: %v", err)
	}

	// Initialize handlers
	// Note: EventPublisher is nil for now - would be injected in full implementation
	cmdHandler := handler.NewOrderCommandHandler(orderRepo, nil)
	qryHandler := handler.NewOrderQueryHandler(orderRepo)

	// Initialize HTTP handlers
	httpHandler := httpapi.NewOrderHandler(cmdHandler, qryHandler)

	// Setup routes
	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, httpHandler)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpapi.Middleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Order service starting on port %s", cfg.Port)
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
	return Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB", "orders"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	fmt.Println("Order Service v1.0.0")
}
