# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **SAGA Pattern Implementation** - Complete choreography-based saga for order processing
  - Order service publishes `OrderCreated` events to Kafka `order-events` topic
  - Payment service consumes `OrderCreated` events, processes payments, publishes `PaymentProcessed`/`PaymentFailed` events
  - Fulfillment service consumes `PaymentProcessed` events, creates shipments, publishes `ShipmentCreated`/`ShipmentFailed` events
  - Order service consumes payment and fulfillment events to update order status accordingly

- **Compensation/Rollback Logic**
  - Order cancellation on `PaymentFailed` events
  - Payment refunds on `ShipmentFailed` events (triggers `PaymentRefunded`)
  - Shipment cancellation on `PaymentRefunded` events
  - Order cancellation on `PaymentRefunded` events

- **Shared SAGA Event Types** (`shared/pkg/saga/events.go`)
  - `OrderCreatedEvent`, `PaymentProcessedEvent`, `PaymentFailedEvent`, `PaymentRefundedEvent`
  - `ShipmentCreatedEvent`, `ShipmentFailedEvent`, `ShipmentShippedEvent`, `ShipmentDeliveredEvent`
  - Topic constants: `order-events`, `payment-events`, `fulfillment-events`

- **Service-specific SAGA Event Handlers**
  - `services/order/internal/application/saga/event_handler.go`
  - `services/payment/internal/application/saga/event_handler.go`
  - `services/fulfillment/internal/application/saga/event_handler.go`

- **Kafka Consumer Integration**
  - All services now start Kafka consumers in background goroutines
  - Consumers registered with appropriate event handlers per topic
  - Consumer groups: `order-service`, `payment-service`, `fulfillment-service`

### Fixed
- Fixed Kafka event publishing - events are now captured before repository `Save()` clears them
- Fixed consumer header lookup to check both `event-type` and `event_type` header formats
- Fixed Payment service state transition - properly transitions through `PROCESSING` state before `COMPLETED`

## [1.0.0] - 2025-11-25

### Added
- Initial release with three microservices: Order, Payment, Fulfillment
- Domain-Driven Design architecture with aggregates, entities, value objects
- MongoDB persistence with transactional outbox pattern
- Kafka event publishing infrastructure
- Docker Compose setup for local development
- Health check endpoints for all services
