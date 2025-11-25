# Project Plan: E-Commerce Microservices Platform

## Document Cross-References

| Document | Purpose | Key Sections |
|----------|---------|--------------|
| **Requirements.md** | What to build | Business flows, tech stack, acceptance criteria |
| **Task.md** | Detailed tasks | 209 tasks with dependencies, sequential/concurrent |
| **Guidelines.md** | How to code | Layer standards, copy-paste agent prompts |
| **CHANGELOG.md** | Track progress | Task completion log, verification metadata |
| **Setup.md** | Environment config | Prerequisites, Docker Compose, agent bootstrap |
| **tasks.json** | Machine-readable | Programmatic queries, DoD checklists, acceptance criteria |

---

## Executive Summary

This plan outlines the phased approach to build a scalable e-commerce microservices platform using Go, Kafka, MongoDB, and GitOps practices. The project is divided into 7 phases, with each phase building upon the previous.

---

## Phase Overview

| Phase | Name | Description | Dependencies |
|-------|------|-------------|--------------|
| 1 | Foundation | Project structure, shared libraries, base configuration | None |
| 2 | Domain Layer | Domain models, events, aggregates for all services | Phase 1 |
| 3 | Application Layer | Use cases, handlers, SAGA orchestration | Phase 2 |
| 4 | Infrastructure Layer | MongoDB, Kafka, Debezium integration | Phase 2 |
| 5 | DevOps & CI/CD | Containers, pipelines, GitOps setup | Phase 3, 4 |
| 6 | Observability | Monitoring, dashboards, alerting | Phase 5 |
| 7 | Quality Assurance | Quality gates, security scanning, performance validation | Phase 6 |

---

## Phase 1: Foundation

**Objective**: Establish project structure, shared libraries, development environment, and GitOps repository.

### 1.0 GitOps Repository Setup

**Critical First Step**: Initialize the GitOps repository structure before any code development.

```
GitOps/
├── CHANGELOG.md              # Project state tracking (updated after each task)
├── Requirements.md           # Project requirements
├── Plan.md                   # This file
├── Task.md                   # Task breakdown
├── Guidelines.md             # Developer guidelines
└── ecommerce-platform/       # Application monorepo
    ├── services/
    ├── shared/
    ├── infrastructure/
    └── ...
```

#### Git Workflow Requirements

1. **Commit per Task**: Each completed task MUST have a dedicated commit
2. **Commit Message Format**:
   ```
   [TASK-X.X.X] Brief description

   - Detailed change 1
   - Detailed change 2

   Closes: #issue (if applicable)
   ```
3. **Changelog Update**: After each task completion:
   - Update CHANGELOG.md with task details
   - Include files changed, features added
   - Commit changelog update with the task

#### Deliverables
- [ ] Initialize Git repository
- [ ] Create initial branch structure (main, develop)
- [ ] Setup branch protection rules (document for GitHub)
- [ ] Create CHANGELOG.md with initial structure
- [ ] First commit with planning documents

### 1.1 Repository Structure

```
ecommerce-platform/
├── services/
│   ├── order/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   │   ├── aggregate/
│   │   │   │   ├── entity/
│   │   │   │   ├── valueobject/
│   │   │   │   ├── event/
│   │   │   │   └── repository/
│   │   │   ├── application/
│   │   │   │   ├── command/
│   │   │   │   ├── query/
│   │   │   │   ├── handler/
│   │   │   │   └── saga/
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/
│   │   │   │   ├── messaging/
│   │   │   │   └── outbox/
│   │   │   └── api/
│   │   │       └── http/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── payment/
│   │   └── [same structure]
│   └── fulfillment/
│       └── [same structure]
├── shared/
│   ├── pkg/
│   │   ├── events/
│   │   ├── outbox/
│   │   ├── kafka/
│   │   ├── mongodb/
│   │   └── logging/
│   └── go.mod
├── infrastructure/
│   ├── terraform/
│   │   ├── modules/
│   │   │   ├── aks/
│   │   │   ├── networking/
│   │   │   └── storage/
│   │   ├── environments/
│   │   │   ├── dev/
│   │   │   ├── staging/
│   │   │   └── prod/
│   │   └── main.tf
│   ├── helm/
│   │   ├── charts/
│   │   │   ├── order-service/
│   │   │   ├── payment-service/
│   │   │   ├── fulfillment-service/
│   │   │   ├── kafka/
│   │   │   ├── mongodb/
│   │   │   └── monitoring/
│   │   └── values/
│   └── k8s/
│       ├── base/
│       └── overlays/
├── .github/
│   └── workflows/
│       ├── ci-order.yml
│       ├── ci-payment.yml
│       ├── ci-fulfillment.yml
│       └── cd-deploy.yml
├── argocd/
│   └── applications/
└── docs/
```

### 1.2 Shared Libraries

| Library | Purpose |
|---------|---------|
| `shared/pkg/events` | Base event interfaces and types |
| `shared/pkg/outbox` | Outbox pattern implementation |
| `shared/pkg/kafka` | Kafka producer/consumer wrappers |
| `shared/pkg/mongodb` | MongoDB connection and utilities |
| `shared/pkg/logging` | Structured logging (zerolog/zap) |

### 1.3 Deliverables

- [ ] Initialize Go modules for all services
- [ ] Create shared library package
- [ ] Setup development environment (Docker Compose)
- [ ] Configure linting (golangci-lint)
- [ ] Setup pre-commit hooks

---

## Phase 2: Domain Layer

**Objective**: Implement domain models following DDD principles for all three services.

### 2.1 Order Service Domain

```go
// Aggregate Root
type Order struct {
    ID          OrderID
    CustomerID  CustomerID
    Items       []OrderItem
    TotalAmount Money
    Status      OrderStatus
    Events      []DomainEvent
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Value Objects
type OrderID string
type CustomerID string
type Money struct { Amount int64; Currency string }
type OrderStatus string

// Domain Events
type OrderCreated struct { ... }
type OrderStatusChanged struct { ... }
```

### 2.2 Payment Service Domain

```go
// Aggregate Root
type Payment struct {
    ID        PaymentID
    OrderID   OrderID
    Amount    Money
    Status    PaymentStatus
    Events    []DomainEvent
    CreatedAt time.Time
}

// Domain Events
type PaymentReceived struct { ... }
type PaymentFailed struct { ... }
type RefundProcessed struct { ... }
```

### 2.3 Fulfillment Service Domain

```go
// Aggregate Root
type Shipment struct {
    ID             ShipmentID
    OrderID        OrderID
    TrackingNumber string
    Carrier        string
    Status         ShipmentStatus
    Events         []DomainEvent
    CreatedAt      time.Time
}

// Domain Events
type OrderShipped struct { ... }
type ShipmentFailed struct { ... }
type OrderDelivered struct { ... }
```

### 2.4 Deliverables

- [ ] Order domain: Aggregate, entities, value objects, events
- [ ] Payment domain: Aggregate, entities, value objects, events
- [ ] Fulfillment domain: Aggregate, entities, value objects, events
- [ ] Domain repository interfaces
- [ ] Unit tests for domain logic (>80% coverage)

---

## Phase 3: Application Layer

**Objective**: Implement use cases, command/query handlers, and SAGA orchestration.

### 3.1 CQRS Pattern

```
┌─────────────────────────────────────────────────┐
│                  Application Layer               │
├─────────────────────────────────────────────────┤
│                                                  │
│  Commands (Write)          Queries (Read)        │
│  ┌───────────────┐         ┌───────────────┐    │
│  │CreateOrderCmd │         │ GetOrderQuery │    │
│  │CancelOrderCmd │         │ListOrdersQuery│    │
│  │UpdateStatusCmd│         │               │    │
│  └───────┬───────┘         └───────┬───────┘    │
│          │                         │            │
│          ▼                         ▼            │
│  ┌───────────────┐         ┌───────────────┐    │
│  │CommandHandler │         │ QueryHandler  │    │
│  └───────────────┘         └───────────────┘    │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 3.2 SAGA Orchestration (Order Service)

```go
type OrderSaga struct {
    orderRepo    OrderRepository
    eventPublisher EventPublisher
}

func (s *OrderSaga) HandlePaymentReceived(event PaymentReceived) error {
    order, err := s.orderRepo.FindByID(event.OrderID)
    if err != nil {
        return err
    }

    order.MarkAsPaid()

    if err := s.orderRepo.Save(order); err != nil {
        return err
    }

    return s.eventPublisher.Publish(order.Events...)
}

func (s *OrderSaga) HandlePaymentFailed(event PaymentFailed) error {
    // Compensation: Cancel order
    order, _ := s.orderRepo.FindByID(event.OrderID)
    order.Cancel(event.Reason)
    s.orderRepo.Save(order)
    return s.eventPublisher.Publish(OrderCancelled{...})
}
```

### 3.3 Event Handlers per Service

| Service | Incoming Events | Actions |
|---------|-----------------|---------|
| Order | PaymentReceived | Update status to PAID |
| Order | PaymentFailed | Cancel order (compensation) |
| Order | OrderShipped | Update status to SHIPPED |
| Order | ShipmentFailed | Cancel order, trigger refund |
| Order | OrderDelivered | Update status to COMPLETED |
| Payment | OrderCreated | Process payment |
| Payment | OrderCancelled | Process refund if paid |
| Fulfillment | OrderCreated | Prepare shipment |
| Fulfillment | PaymentReceived | Ship order |
| Fulfillment | OrderCancelled | Cancel shipment |

### 3.4 Deliverables

- [ ] Command handlers for all services
- [ ] Query handlers for all services
- [ ] SAGA orchestrator in Order service
- [ ] Event handlers for cross-service communication
- [ ] Compensation logic for failure scenarios
- [ ] Integration tests for SAGA flows

---

## Phase 4: Infrastructure Layer

**Objective**: Implement persistence, messaging, and outbox pattern integration.

### 4.1 MongoDB Persistence

```go
type MongoOrderRepository struct {
    collection *mongo.Collection
    outbox     *OutboxRepository
}

func (r *MongoOrderRepository) Save(order *Order) error {
    session, _ := r.client.StartSession()
    defer session.EndSession(context.Background())

    _, err := session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
        // Save order
        _, err := r.collection.ReplaceOne(sc, bson.M{"_id": order.ID}, order, options.Replace().SetUpsert(true))
        if err != nil {
            return nil, err
        }

        // Save events to outbox (same transaction)
        for _, event := range order.Events {
            if err := r.outbox.Save(sc, event); err != nil {
                return nil, err
            }
        }

        return nil, nil
    })

    return err
}
```

### 4.2 Outbox Pattern

```go
// Outbox table schema
type OutboxEntry struct {
    ID          string
    AggregateID string
    EventType   string
    Payload     []byte
    CreatedAt   time.Time
    Published   bool
}

// Debezium will capture changes to this collection
// and publish to Kafka topic: outbox.events
```

### 4.3 Kafka Integration

```go
type KafkaEventPublisher struct {
    producer *kafka.Producer
    topics   map[string]string
}

type KafkaEventConsumer struct {
    consumer *kafka.Consumer
    handlers map[string]EventHandler
}

func (c *KafkaEventConsumer) Start() {
    for {
        msg, _ := c.consumer.ReadMessage(-1)
        eventType := getEventType(msg)
        handler := c.handlers[eventType]
        handler.Handle(msg.Value)
    }
}
```

### 4.4 Debezium Configuration

```json
{
  "name": "outbox-connector",
  "config": {
    "connector.class": "io.debezium.connector.mongodb.MongoDbConnector",
    "mongodb.connection.string": "mongodb://mongo:27017",
    "collection.include.list": "*.outbox",
    "transforms": "outbox",
    "transforms.outbox.type": "io.debezium.transforms.outbox.EventRouter",
    "transforms.outbox.table.field.event.key": "aggregate_id",
    "transforms.outbox.table.field.event.type": "event_type",
    "transforms.outbox.table.field.event.payload": "payload"
  }
}
```

### 4.5 Deliverables

- [ ] MongoDB repositories for all services
- [ ] Outbox table and repository
- [ ] Kafka producer/consumer implementations
- [ ] Debezium connector configuration
- [ ] Event serialization/deserialization
- [ ] Connection pooling and retry logic
- [ ] Integration tests with testcontainers

---

## Phase 5: DevOps & CI/CD

**Objective**: Containerize services, setup CI/CD pipelines, and configure GitOps.

### 5.1 Dockerization

```dockerfile
# Multi-stage build for Go services
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/main.go

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
COPY --from=builder /service /service
ENTRYPOINT ["/service"]
```

### 5.2 GitHub Actions CI Pipeline

```yaml
name: CI Pipeline
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run unit tests
        run: go test ./... -cover

  security:
    runs-on: ubuntu-latest
    steps:
      - name: OWASP Dependency Check
        uses: dependency-check/Dependency-Check_Action@main
      - name: SonarQube Scan
        uses: sonarsource/sonarqube-scan-action@master

  build:
    needs: [test, security]
    runs-on: ubuntu-latest
    steps:
      - name: Build Docker image
        run: docker build -t $IMAGE_NAME:${{ github.sha }} .
      - name: Trivy scan
        uses: aquasecurity/trivy-action@master
      - name: Push to registry
        run: docker push $IMAGE_NAME:${{ github.sha }}
```

### 5.3 GitHub Actions CD Workflow

```yaml
name: CD Deploy
on:
  workflow_run:
    workflows: ["CI Pipeline"]
    types: [completed]
    branches: [main]

jobs:
  update-manifests:
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    steps:
      - name: Checkout manifests repo
        uses: actions/checkout@v4
        with:
          token: ${{ secrets.PAT_TOKEN }}

      - name: Update image tag
        run: |
          sed -i "s|image:.*|image: ${{ env.REGISTRY }}/${{ env.IMAGE }}:${{ github.sha }}|" k8s/deployment.yaml

      - name: Commit and push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add k8s/
          git commit -m "Update image to ${{ github.sha }}"
          git push
```

### 5.4 Terraform AKS Provisioning

```hcl
module "aks" {
  source              = "./modules/aks"
  resource_group_name = var.resource_group_name
  location            = var.location
  cluster_name        = var.cluster_name
  node_count          = var.node_count
  node_size           = var.node_size

  tags = {
    Environment = var.environment
    Project     = "ecommerce-platform"
  }
}
```

### 5.5 ArgoCD Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: order-service
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/org/ecommerce-platform.git
    targetRevision: HEAD
    path: infrastructure/k8s/overlays/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: ecommerce
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 5.6 Deliverables

- [ ] Dockerfiles for all services
- [ ] Docker Compose for local development
- [ ] Terraform modules for AKS
- [ ] Helm charts for all services
- [ ] GitHub Actions CI workflows
- [ ] GitHub Actions CD workflow
- [ ] ArgoCD application manifests
- [ ] Kubernetes base manifests

---

## Phase 6: Observability

**Objective**: Implement comprehensive monitoring, logging, and alerting.

### 6.1 Prometheus Metrics

```go
var (
    ordersCreated = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "orders_created_total",
        Help: "Total number of orders created",
    })

    orderProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "order_processing_duration_seconds",
        Help:    "Time spent processing orders",
        Buckets: prometheus.DefBuckets,
    })
)
```

### 6.2 Grafana Dashboards

| Dashboard | Panels |
|-----------|--------|
| Service Overview | Request rate, error rate, latency percentiles |
| Order Pipeline | Orders by status, processing time, failure rate |
| Kafka Metrics | Consumer lag, throughput, partition health |
| Infrastructure | CPU, memory, network, disk by pod |

### 6.3 Alerting Rules

```yaml
groups:
  - name: ecommerce-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected

      - alert: KafkaConsumerLag
        expr: kafka_consumer_lag > 1000
        for: 10m
        labels:
          severity: warning
```

### 6.4 Deliverables

- [ ] Prometheus configuration
- [ ] Custom metrics in all services
- [ ] Grafana dashboard JSON files
- [ ] AlertManager configuration
- [ ] Structured logging implementation
- [ ] Log aggregation setup (optional: EFK stack)

---

## Dependency Graph

```
Phase 1 (Foundation)
    │
    ▼
Phase 2 (Domain Layer)
    │
    ├──────────────┬──────────────┐
    ▼              ▼              │
Phase 3        Phase 4            │
(Application)  (Infrastructure)   │
    │              │              │
    └──────┬───────┘              │
           ▼                      │
       Phase 5 ◄──────────────────┘
       (DevOps)
           │
           ▼
       Phase 6
    (Observability)
           │
           ▼
       Phase 7
  (Quality Assurance)
```

---

## Timeline & Milestones

| Phase | Milestone | Checkpoint |
|-------|-----------|------------|
| 1 | Project structure complete | All services scaffolded, shared libs ready |
| 2 | Domain models complete | All aggregates, events, repos defined |
| 3 | Application layer complete | SAGA working, all handlers implemented |
| 4 | Infrastructure layer complete | Services can persist and communicate |
| 5 | CI/CD pipeline complete | Automated build, test, deploy working |
| 6 | Observability complete | Monitoring, alerting, dashboards live |
| 7 | Production ready | Quality gates passing, security validated |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Kafka complexity | Use managed Kafka (Confluent/Azure Event Hubs) if needed |
| Debezium setup | Provide detailed configuration, fallback to polling publisher |
| AKS provisioning | Document manual steps as backup to Terraform |
| SAGA complexity | Thorough testing of all failure scenarios |

---

## Cost Estimation & Resource Planning

### Infrastructure Cost Breakdown

#### Azure AKS Cluster (Production)
| Component | Specification | Monthly Cost | Notes |
|-----------|---------------|--------------|-------|
| **Control Plane** | Standard tier | $75 | Managed by Azure |
| **Worker Nodes** | 3x D4s v3 (4 vCPU, 16GB RAM) | $450 | Base cluster |
| **Auto-scaling** | 1-10 nodes | Variable | Based on load |
| **Load Balancer** | Standard tier | $20 | For ingress |
| **Network** | VNet, NSG, VPN | $50 | Basic networking |

**Total AKS**: ~$595/month (base) + variable scaling

#### Azure Container Registry (ACR)
| Component | Specification | Monthly Cost |
|-----------|---------------|--------------|
| **Storage** | 100GB | $5 |
| **Data Transfer** | 100GB out | $10 |
| **Build Tasks** | 10 hours | $2 |

**Total ACR**: ~$17/month

#### Azure Database for MongoDB
| Component | Specification | Monthly Cost |
|-----------|---------------|--------------|
| **Compute** | 2 vCPU, 8GB RAM | $150 |
| **Storage** | 128GB | $25 |
| **Backup** | 7-day retention | $10 |
| **High Availability** | 3-node replica set | Included |

**Total MongoDB**: ~$185/month

#### Azure Event Hubs (Kafka-compatible)
| Component | Specification | Monthly Cost |
|-----------|---------------|--------------|
| **Throughput** | 20 TU (20MB/s) | $200 |
| **Storage** | 100GB | $10 |
| **Data Transfer** | 100GB out | $10 |

**Total Event Hubs**: ~$220/month

#### Monitoring & Observability
| Component | Specification | Monthly Cost |
|-----------|---------------|--------------|
| **Azure Monitor** | Logs + Metrics | $50 |
| **Application Insights** | Full observability | $100 |
| **Log Analytics** | 1GB/day | $150 |

**Total Monitoring**: ~$300/month

### Development Environment Costs

#### Azure AKS Cluster (Development)
| Component | Specification | Monthly Cost |
|-----------|---------------|--------------|
| **Control Plane** | Standard tier | $75 |
| **Worker Nodes** | 2x D2s v3 (2 vCPU, 8GB RAM) | $150 |
| **Load Balancer** | Basic tier | $10 |

**Total Dev AKS**: ~$235/month

### Total Cost Summary

| Environment | Monthly Cost | Annual Cost |
|-------------|--------------|-------------|
| **Production** | $1,317 | $15,804 |
| **Development** | $235 | $2,820 |
| **Total** | $1,552 | $18,624 |

### Cost Optimization Strategies

1. **Reserved Instances**: 1-year reservation saves ~20%
2. **Auto-scaling**: Scale down during off-peak hours
3. **Spot Instances**: Use for non-critical workloads
4. **Resource Rightsizing**: Monitor and adjust pod limits
5. **Data Lifecycle**: Implement data retention policies

### Resource Planning

#### Team Resources Required

| Role | Count | Time Allocation | Duration |
|------|-------|-----------------|----------|
| **Backend Developer** | 3 | 100% | 6 months |
| **DevOps Engineer** | 1 | 100% | 6 months |
| **QA Engineer** | 1 | 50% | 6 months |
| **Technical Lead** | 1 | 50% | 6 months |

#### Infrastructure Setup Timeline

| Phase | Duration | Key Activities |
|-------|----------|----------------|
| **Planning** | 2 weeks | Requirements, architecture design |
| **Development** | 16 weeks | Core implementation, testing |
| **DevOps Setup** | 4 weeks | CI/CD, infrastructure, monitoring |
| **Testing** | 4 weeks | Integration, performance, security |
| **Deployment** | 2 weeks | Production rollout, validation |

**Total Timeline**: 28 weeks (7 months)

---

## Success Criteria

1. All microservices deployed and communicating
2. SAGA pattern working with compensation
3. Zero-downtime deployments via ArgoCD
4. All security gates passing in CI
5. Grafana dashboards showing real-time metrics
6. End-to-end order flow functional
