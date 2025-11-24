# Development Cheat Sheet

## 🏗️ Architecture Patterns

### Clean Architecture Layers
```
┌─────────────────────────────────────┐
│          Delivery Layer             │  HTTP, gRPC handlers
├─────────────────────────────────────┤
│          Use Case Layer             │  Application services
├─────────────────────────────────────┤
│          Domain Layer               │  Business logic, entities
├─────────────────────────────────────┤
│          Infrastructure Layer       │  DB, messaging, external APIs
└─────────────────────────────────────┘
```

### CQRS Pattern
```go
// Command - Changes state
type CreateOrderCommand struct {
    CustomerID string
    Items      []OrderItem
}

// Query - Reads state
type GetOrderQuery struct {
    OrderID string
}

// Handler separation
type CommandHandler interface {
    Handle(ctx context.Context, cmd interface{}) error
}

type QueryHandler interface {
    Handle(ctx context.Context, query interface{}) (interface{}, error)
}
```

### SAGA Pattern
```go
type Saga interface {
    Start(ctx context.Context, data interface{}) error
    HandleEvent(ctx context.Context, event DomainEvent) error
    Compensate(ctx context.Context) error
}

// State machine
type OrderSaga struct {
    state    SagaState
    orderID  OrderID
    events   []DomainEvent
}
```

## 🗄️ Database Patterns

### Repository Pattern
```go
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    Update(ctx context.Context, order *Order) error
}

type MongoOrderRepository struct {
    collection *mongo.Collection
}
```

### Outbox Pattern
```go
type OutboxEntry struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    AggregateID string            `bson:"aggregate_id"`
    EventType   string            `bson:"event_type"`
    Payload     []byte            `bson:"payload"`
    CreatedAt   time.Time         `bson:"created_at"`
    Processed   bool              `bson:"processed"`
}
```

## 📨 Event-Driven Patterns

### Domain Events
```go
type DomainEvent interface {
    EventType() string
    AggregateID() string
    OccurredAt() time.Time
}

type OrderCreatedEvent struct {
    OrderID    OrderID
    CustomerID CustomerID
    Items      []OrderItem
    CreatedAt  time.Time
}

func (e OrderCreatedEvent) EventType() string {
    return "OrderCreated"
}
```

### Event Handlers
```go
type EventHandler interface {
    HandleEvent(ctx context.Context, event DomainEvent) error
}

type PaymentEventHandler struct {
    paymentService PaymentService
}

func (h *PaymentEventHandler) HandleEvent(ctx context.Context, event DomainEvent) error {
    switch e := event.(type) {
    case OrderCreatedEvent:
        return h.handleOrderCreated(ctx, e)
    default:
        return nil
    }
}
```

## 🧪 Testing Patterns

### Unit Test Structure
```go
func TestOrderCreation(t *testing.T) {
    // Given
    repo := &mockOrderRepository{}
    handler := NewCreateOrderHandler(repo)

    cmd := CreateOrderCommand{
        CustomerID: "customer-123",
        Items: []OrderItem{{
            ProductID: "product-456",
            Quantity:  2,
        }},
    }

    // When
    err := handler.Handle(context.Background(), cmd)

    // Then
    assert.NoError(t, err)
    assert.Len(t, repo.savedOrders, 1)
}
```

### Integration Test
```go
func TestOrderSagaIntegration(t *testing.T) {
    // Setup test containers
    mongoC, kafkaC := setupTestContainers(t)
    defer cleanupTestContainers(t, mongoC, kafkaC)

    // Create services
    orderSvc := NewOrderService(mongoC)
    paymentSvc := NewPaymentService(mongoC, kafkaC)

    // Execute saga
    orderID, err := orderSvc.CreateOrder(ctx, customerID, items)
    assert.NoError(t, err)

    // Verify final state
    order, err := orderSvc.GetOrder(ctx, orderID)
    assert.NoError(t, err)
    assert.Equal(t, OrderStatusPaid, order.Status)
}
```

## 🐳 Docker Patterns

### Multi-stage Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Docker Compose for Development
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MONGO_URI=mongodb://mongo:27017
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - mongo
      - kafka
```

## ☸️ Kubernetes Patterns

### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: order-service
  template:
    metadata:
      labels:
        app: order-service
    spec:
      containers:
      - name: order-service
        image: myregistry.azurecr.io/order-service:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: MONGO_URI
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: mongo-uri
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

### Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: order-service
spec:
  selector:
    app: order-service
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## 🔒 Security Patterns

### JWT Authentication
```go
func authenticateMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Input Validation
```go
type CreateOrderRequest struct {
    CustomerID string        `json:"customer_id" validate:"required,min=1,max=50"`
    Items      []OrderItem   `json:"items" validate:"required,min=1,dive"`
}

type OrderItem struct {
    ProductID string `json:"product_id" validate:"required,min=1,max=50"`
    Quantity  int    `json:"quantity" validate:"required,min=1,max=100"`
}

func validateRequest(r *CreateOrderRequest) error {
    validate := validator.New()
    return validate.Struct(r)
}
```

## 📊 Monitoring Patterns

### Prometheus Metrics
```go
var (
    ordersCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_created_total",
            Help: "Total number of orders created",
        },
        []string{"status"},
    )

    orderCreationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "order_creation_duration_seconds",
            Help:    "Time taken to create orders",
            Buckets: prometheus.DefBuckets,
        },
        []string{"status"},
    )
)

func init() {
    prometheus.MustRegister(ordersCreated)
    prometheus.MustRegister(orderCreationDuration)
}
```

### Structured Logging
```go
type Logger interface {
    Info(msg string, fields ...Field)
    Error(msg string, err error, fields ...Field)
    WithField(key string, value interface{}) Logger
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) error {
    logger := logging.FromContext(ctx).WithField("order_id", cmd.OrderID)

    logger.Info("Creating order", logging.Field("customer_id", cmd.CustomerID))

    // Business logic...

    logger.Info("Order created successfully")
    return nil
}
```

## 🚀 CI/CD Patterns

### GitHub Actions Workflow
```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

    - name: Run tests
      run: go test ./... -v -coverprofile=coverage.out

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Helm Chart Structure
```
mychart/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   └── _helpers.tpl
└── charts/
    └── dependency-chart/
```

## 🐛 Debugging Commands

### Go Debugging
```bash
# Race detection
go test -race ./...

# Memory profiling
go test -memprofile mem.out
go tool pprof mem.out

# CPU profiling
go test -cpuprofile cpu.out
go tool pprof cpu.out

# Debug build
go build -gcflags="all=-N -l"
dlv exec ./main
```

### Kubernetes Debugging
```bash
# Pod logs
kubectl logs -f pod/order-service-12345

# Exec into pod
kubectl exec -it pod/order-service-12345 -- /bin/sh

# Port forward
kubectl port-forward pod/order-service-12345 8080:8080

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp

# Describe pod
kubectl describe pod order-service-12345
```

### Database Debugging
```bash
# MongoDB queries
mongosh --eval "db.orders.find().limit(5)"

# Check indexes
mongosh --eval "db.orders.getIndexes()"

# Explain query
mongosh --eval "db.orders.find({customer_id: '123'}).explain('executionStats')"
```

This cheat sheet covers the most common patterns and commands you'll need. Keep it handy and update it as you learn new patterns!