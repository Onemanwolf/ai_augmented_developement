# Developer Guidelines & Agent Prompts

## Document Cross-References

| Document | Purpose | Key Sections |
|----------|---------|--------------|
| **Requirements.md** | What to build | Business flows, tech stack, acceptance criteria |
| **Plan.md** | Implementation phases | Phase 1-7, architecture diagrams |
| **Task.md** | Detailed tasks | 209 tasks, use task IDs in commits |
| **CHANGELOG.md** | Track progress | Task completion log, verification metadata |
| **Setup.md** | Environment config | Prerequisites, Docker Compose, agent bootstrap |
| **tasks.json** | Machine-readable | Query tasks, DoD checklists, guidelinesRef links |

---

## Overview

This document provides coding standards, acceptance criteria, and copy-paste prompts for developer agents working on the E-Commerce Microservices Platform.

---

## Table of Contents

1. [Coding Standards](#coding-standards)
2. [Layer-Specific Guidelines](#layer-specific-guidelines)
3. [Acceptance Criteria](#acceptance-criteria)
4. [Agent Prompts by Phase](#agent-prompts-by-phase)

---

## Coding Standards

### Go Language Standards

```go
// Package naming: lowercase, single word
package domain

// Interface naming: describe behavior, use -er suffix
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
}

// Struct naming: PascalCase, noun
type Order struct {
    ID        OrderID
    Status    OrderStatus
    CreatedAt time.Time
}

// Function naming: PascalCase for exported, camelCase for private
func (o *Order) MarkAsPaid() {
    o.Status = OrderStatusPaid
    o.addEvent(NewOrderStatusChangedEvent(o.ID, o.Status))
}

func (o *Order) addEvent(event DomainEvent) {
    o.events = append(o.events, event)
}

// Error handling: wrap with context
if err != nil {
    return fmt.Errorf("failed to save order %s: %w", order.ID, err)
}

// Context: always first parameter
func (r *MongoOrderRepository) Save(ctx context.Context, order *Order) error
```

### File Organization

```
service/
├── cmd/
│   └── main.go              # Entry point only
├── internal/
│   ├── domain/              # Business logic (no external deps)
│   │   ├── aggregate/       # Aggregate roots
│   │   ├── entity/          # Entities
│   │   ├── valueobject/     # Value objects
│   │   ├── event/           # Domain events
│   │   └── repository/      # Repository interfaces
│   ├── application/         # Use cases
│   │   ├── command/         # Command definitions
│   │   ├── query/           # Query definitions
│   │   ├── handler/         # Command/Query handlers
│   │   └── saga/            # SAGA orchestration
│   ├── infrastructure/      # External integrations
│   │   ├── persistence/     # Database implementations
│   │   ├── messaging/       # Kafka implementations
│   │   └── outbox/          # Outbox pattern
│   └── api/                 # API layer
│       ├── http/            # HTTP handlers
│       └── dto/             # Data transfer objects
├── Dockerfile
├── go.mod
└── go.sum
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Package | lowercase | `domain`, `persistence` |
| Interface | PascalCase + behavior | `OrderRepository`, `EventPublisher` |
| Struct | PascalCase + noun | `Order`, `PaymentReceived` |
| Method | PascalCase (exported) | `CreateOrder`, `HandleEvent` |
| Method | camelCase (private) | `validateOrder`, `addEvent` |
| Constant | PascalCase | `OrderStatusPaid` |
| Variable | camelCase | `orderID`, `customerName` |
| File | snake_case | `order_repository.go` |
| Test file | `*_test.go` | `order_test.go` |

### Error Handling

```go
// Define domain errors
var (
    ErrOrderNotFound    = errors.New("order not found")
    ErrInvalidOrderID   = errors.New("invalid order ID")
    ErrPaymentRequired  = errors.New("payment required before shipping")
)

// Wrap errors with context
func (r *repo) FindByID(ctx context.Context, id OrderID) (*Order, error) {
    result, err := r.collection.FindOne(ctx, bson.M{"_id": id})
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, ErrOrderNotFound
        }
        return nil, fmt.Errorf("database error finding order %s: %w", id, err)
    }
    // ...
}
```

### Testing Standards

```go
// Table-driven tests
func TestOrder_MarkAsPaid(t *testing.T) {
    tests := []struct {
        name           string
        initialStatus  OrderStatus
        expectedStatus OrderStatus
        expectEvent    bool
    }{
        {
            name:           "pending order can be marked as paid",
            initialStatus:  OrderStatusPending,
            expectedStatus: OrderStatusPaid,
            expectEvent:    true,
        },
        // more cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            order := NewOrder(...)
            order.Status = tt.initialStatus

            order.MarkAsPaid()

            assert.Equal(t, tt.expectedStatus, order.Status)
            if tt.expectEvent {
                assert.Len(t, order.Events(), 1)
            }
        })
    }
}
```

---

## Layer-Specific Guidelines

### Domain Layer

**Rules:**
- NO external dependencies (no database, no Kafka, no HTTP)
- Pure business logic only
- Aggregates protect invariants
- Domain events capture state changes
- Repository interfaces defined here, implemented elsewhere

**Structure:**
```go
// Aggregate must:
// 1. Have an identity (ID)
// 2. Protect invariants
// 3. Emit domain events
// 4. Be the unit of persistence

type Order struct {
    id        OrderID           // Identity
    items     []OrderItem       // Child entities
    status    OrderStatus       // State
    events    []DomainEvent     // Uncommitted events
}

// Methods enforce business rules
func (o *Order) AddItem(item OrderItem) error {
    if o.status != OrderStatusDraft {
        return ErrCannotModifyOrder
    }
    if item.Quantity <= 0 {
        return ErrInvalidQuantity
    }
    o.items = append(o.items, item)
    return nil
}
```

### Application Layer

**Rules:**
- Orchestrates domain objects
- No business logic (delegate to domain)
- Handles transactions
- Publishes integration events

**Structure:**
```go
type CreateOrderHandler struct {
    repo      OrderRepository
    publisher EventPublisher
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) error {
    // 1. Create domain object
    order, err := NewOrder(cmd.CustomerID, cmd.Items)
    if err != nil {
        return err
    }

    // 2. Persist (including outbox)
    if err := h.repo.Save(ctx, order); err != nil {
        return err
    }

    // Domain events are saved to outbox in same transaction
    // Debezium will publish them to Kafka

    return nil
}
```

### Infrastructure Layer

**Rules:**
- Implements interfaces from domain/application
- Handles external system integration
- Manages connections and retries
- Translates between domain and external formats

**Structure:**
```go
type MongoOrderRepository struct {
    client     *mongo.Client
    collection *mongo.Collection
    outbox     *OutboxRepository
}

func (r *MongoOrderRepository) Save(ctx context.Context, order *Order) error {
    session, err := r.client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)

    _, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
        // Save aggregate
        if err := r.saveOrder(sc, order); err != nil {
            return nil, err
        }

        // Save events to outbox (same transaction)
        for _, event := range order.Events() {
            if err := r.outbox.Save(sc, event); err != nil {
                return nil, err
            }
        }

        return nil, nil
    })

    return err
}
```

### API Layer

**Rules:**
- Thin layer, no business logic
- Input validation only
- Maps DTOs to commands/queries
- Consistent error responses

**Structure:**
```go
func (h *OrderHTTPHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := req.Validate(); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    cmd := CreateOrderCommand{
        CustomerID: req.CustomerID,
        Items:      mapItems(req.Items),
    }

    if err := h.handler.Handle(r.Context(), cmd); err != nil {
        respondError(w, http.StatusInternalServerError, "failed to create order")
        return
    }

    respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}
```

---

## Acceptance Criteria

### Domain Layer Acceptance Criteria

| Criterion | Description |
|-----------|-------------|
| No external imports | Domain package imports only stdlib and shared/pkg |
| Aggregate invariants | All business rules enforced within aggregate |
| Event emission | State changes produce domain events |
| Immutable value objects | Value objects cannot be modified after creation |
| Unit test coverage | Minimum 80% coverage |
| Test isolation | Tests don't require external dependencies |

### Application Layer Acceptance Criteria

| Criterion | Description |
|-----------|-------------|
| Single responsibility | One handler per command/query |
| No business logic | Logic delegated to domain |
| Transaction handling | Operations wrapped in transactions |
| Event publishing | Domain events routed to outbox |
| Error handling | Errors wrapped with context |
| Integration tests | All handlers tested with mocks |

### Infrastructure Layer Acceptance Criteria

| Criterion | Description |
|-----------|-------------|
| Interface compliance | Implements all domain interfaces |
| Connection management | Proper pooling and lifecycle |
| Retry logic | Transient failures handled |
| Outbox integration | Events saved atomically |
| Integration tests | Tested with testcontainers |

### API Layer Acceptance Criteria

| Criterion | Description |
|-----------|-------------|
| Input validation | All inputs validated before processing |
| Consistent responses | Standardized error format |
| HTTP status codes | Appropriate codes for all scenarios |
| Request logging | All requests logged with correlation ID |
| Health endpoint | `/health` returns service status |

### DevOps Acceptance Criteria

| Criterion | Description |
|-----------|-------------|
| Docker builds | Images build successfully |
| Size optimization | Multi-stage builds, minimal image size |
| Security scans | No critical vulnerabilities |
| CI pipeline | All gates pass before merge |
| CD pipeline | Manifests updated automatically |
| GitOps sync | ArgoCD deploys changes |

---

## Quality Agent Role

### Overview

The Quality Agent is a dedicated AI agent responsible for ensuring all work meets requirements, coding standards, and quality benchmarks. This agent acts as the final gatekeeper before code is merged or deployed, providing automated and manual quality assurance.

### Responsibilities

1. **Pre-Merge Quality Gates**: Review all pull requests for compliance
2. **Automated Testing**: Run comprehensive test suites and quality checks
3. **Code Standards Enforcement**: Verify adherence to coding standards and best practices
4. **Requirements Validation**: Ensure implementation meets acceptance criteria
5. **Security Scanning**: Identify security vulnerabilities and compliance issues
6. **Documentation Review**: Verify documentation is complete and accurate
7. **Performance Validation**: Check for performance regressions and bottlenecks
8. **Cross-Cutting Concerns**: Validate logging, error handling, observability

### Quality Gates

#### Gate 1: Code Quality (Automated)
- **Linting**: golangci-lint passes with zero warnings
- **Formatting**: gofmt and goimports applied correctly
- **Security**: gosec identifies no critical vulnerabilities
- **Dependencies**: go mod tidy executed, no vulnerable packages
- **Secrets**: gitleaks finds no exposed secrets

#### Gate 2: Testing (Automated)
- **Unit Tests**: Minimum 80% coverage per package
- **Integration Tests**: All critical paths tested
- **Performance Tests**: Benchmarks meet requirements
- **Security Tests**: Penetration tests pass
- **Load Tests**: Throughput meets SLAs

#### Gate 3: Requirements Compliance (Semi-Automated)
- **Acceptance Criteria**: All DoD checklist items verified
- **Business Logic**: Requirements.md specifications met
- **API Contracts**: OpenAPI specs match implementation
- **Data Contracts**: Event schemas validated
- **Performance**: Latency/throughput requirements satisfied

#### Gate 4: Documentation (Manual Review)
- **Code Comments**: Complex logic documented
- **API Documentation**: Endpoints documented with examples
- **Architecture Decisions**: ADRs updated for changes
- **Runbooks**: Operational procedures documented
- **CHANGELOG**: Changes logged with proper versioning

#### Gate 5: Security & Compliance (Automated + Manual)
- **OWASP Top 10**: No violations in implementation
- **Data Protection**: GDPR/CCPA compliance verified
- **Infrastructure Security**: CIS benchmarks met
- **Access Control**: RBAC properly implemented
- **Audit Logging**: Security events logged appropriately

### Quality Agent Workflow

#### Pre-Commit Quality Check
```
When a developer creates a PR:
1. Quality Agent automatically triggered
2. Run automated quality gates (1-2)
3. Generate quality report with findings
4. Block merge if critical issues found
5. Request fixes for non-critical issues
```

#### Manual Quality Review
```
For complex changes or high-risk features:
1. Quality Agent performs deep code review
2. Validate requirements compliance manually
3. Review architecture and design decisions
4. Assess test coverage adequacy
5. Verify documentation completeness
6. Approve or request changes with detailed feedback
```

#### Post-Merge Validation
```
After merge to main:
1. Run full integration test suite
2. Performance regression testing
3. Security scanning on full codebase
4. Generate quality metrics report
5. Update quality dashboard
```

### Quality Metrics

#### Code Quality Metrics
- **Test Coverage**: Target >80% overall, >90% for critical paths
- **Cyclomatic Complexity**: Average <10 per function
- **Technical Debt**: Maintain debt ratio <5%
- **Code Duplication**: <3% duplicate lines
- **Security Score**: Maintain A grade from security scanners

#### Process Quality Metrics
- **Defect Density**: <0.5 bugs per 1000 lines of code
- **Mean Time to Detection**: <4 hours for critical bugs
- **Review Coverage**: 100% of PRs reviewed
- **Quality Gate Pass Rate**: >95% first-time passes

#### Performance Metrics
- **Build Time**: <10 minutes for full CI pipeline
- **Test Execution Time**: <5 minutes for unit tests
- **Security Scan Time**: <3 minutes
- **Quality Gate Time**: <15 minutes total

### Quality Agent Prompts

#### Prompt: Pre-Merge Quality Review

```
You are the Quality Agent performing a comprehensive pre-merge review.

REF: Guidelines.md Quality Agent Role
REF: Requirements.md acceptance criteria
REF: Task.md DoD checklists

Review the following PR: [PR_LINK]

Perform these checks:

1. CODE QUALITY:
   - Run golangci-lint, check for zero warnings
   - Verify gofmt/goimports formatting
   - Check gosec security issues
   - Review dependency vulnerabilities

2. TESTING:
   - Verify test coverage >80%
   - Run integration tests
   - Check performance benchmarks
   - Validate test isolation

3. REQUIREMENTS COMPLIANCE:
   - Match implementation to acceptance criteria
   - Verify business logic correctness
   - Check API contract compliance
   - Validate event schemas

4. DOCUMENTATION:
   - Review code comments for complex logic
   - Check API documentation updates
   - Verify CHANGELOG entries
   - Confirm runbook updates

5. SECURITY & COMPLIANCE:
   - OWASP Top 10 compliance
   - Data protection requirements
   - Access control implementation
   - Audit logging completeness

Generate a quality report with:
- Overall quality score (A/B/C/D/F)
- Critical issues (blockers)
- Major issues (must fix)
- Minor issues (should fix)
- Recommendations for improvement

If critical issues found, block the merge with detailed remediation steps.
```

#### Prompt: Security Vulnerability Assessment

```
You are the Quality Agent performing a security vulnerability assessment.

REF: Requirements.md Section 5 (Security Requirements)
REF: Guidelines.md Security Standards

Assess the following code changes for security vulnerabilities:

1. INPUT VALIDATION:
   - Check for SQL injection vulnerabilities
   - Verify XSS prevention in web handlers
   - Validate input sanitization
   - Check for command injection risks

2. AUTHENTICATION & AUTHORIZATION:
   - Verify JWT token validation
   - Check RBAC implementation
   - Validate session management
   - Review access control logic

3. DATA PROTECTION:
   - Check encryption at rest/transit
   - Verify sensitive data handling
   - Validate GDPR compliance
   - Check for data leakage risks

4. INFRASTRUCTURE SECURITY:
   - Review Docker image security
   - Check Kubernetes security configs
   - Validate network policies
   - Assess secret management

5. DEPENDENCY SECURITY:
   - Scan for vulnerable packages
   - Check for outdated dependencies
   - Verify dependency integrity
   - Review third-party component security

Generate security assessment report with:
- Risk severity levels (Critical/High/Medium/Low)
- CVSS scores where applicable
- Remediation recommendations
- Compliance status vs requirements
- False positive identifications
```

#### Prompt: Performance Quality Gate

```
You are the Quality Agent validating performance requirements.

REF: Requirements.md Section 4 (Performance Requirements)
REF: Plan.md scalability targets

Validate the following implementation against performance benchmarks:

1. LATENCY REQUIREMENTS:
   - Order creation: <500ms p95
   - Payment processing: <2s p95
   - API response times: <200ms p95
   - Database queries: <50ms average

2. THROUGHPUT REQUIREMENTS:
   - 100 orders/second sustained load
   - 1000 concurrent users supported
   - Event processing: 500 events/second
   - Database connections: efficient pooling

3. RESOURCE EFFICIENCY:
   - Memory usage: <512MB per service instance
   - CPU utilization: <70% under load
   - Database connection pooling
   - Cache hit rates >90%

4. SCALABILITY VALIDATION:
   - Horizontal scaling capability
   - Load balancer configuration
   - Auto-scaling triggers
   - Resource limits and requests

Run performance tests and generate report with:
- Benchmark results vs requirements
- Performance regression analysis
- Bottleneck identification
- Optimization recommendations
- Capacity planning data
```

### Integration with CI/CD

#### Quality Gates in GitHub Actions

```yaml
# .github/workflows/quality-gate.yml
name: Quality Gate
on:
  pull_request:
    branches: [ main, develop ]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Quality Agent
        uses: ./.github/actions/quality-agent
        with:
          pr_number: ${{ github.event.pull_request.number }}
          quality_level: strict
          
      - name: Quality Report
        uses: actions/upload-artifact@v3
        with:
          name: quality-report
          path: quality-report.json
          
      - name: Block on Critical Issues
        if: steps.quality.outputs.critical_issues > 0
        run: exit 1
```

#### Quality Dashboard

The Quality Agent maintains a dashboard showing:
- Quality metrics over time
- Common issue patterns
- Team performance trends
- Quality gate pass/fail rates
- Security vulnerability trends

---

## Agent Prompts by Phase

### How to Use These Prompts

1. Copy the relevant prompt for your task
2. Paste into the AI agent
3. Answer any clarifying questions
4. Review the generated code
5. Ensure acceptance criteria are met
6. Run tests before committing

---

### Phase 1: Foundation Prompts

#### Prompt 1.2.2: Create Events Package

```
You are building a shared events package for a Go microservices project.

Create the following in shared/pkg/events:

1. event.go - Base event interface:
   - DomainEvent interface with methods: EventID(), EventType(), AggregateID(), OccurredAt(), Payload()
   - IntegrationEvent interface embedding DomainEvent
   - BaseEvent struct implementing common fields

2. registry.go - Event type registry:
   - Map event type strings to event structs
   - Register function to add event types
   - Deserialize function to unmarshal events by type

Requirements:
- Use UUID for event IDs (google/uuid package)
- Events must be JSON serializable
- Include comprehensive unit tests
- No external dependencies except stdlib and uuid

Acceptance Criteria:
- All interfaces clearly defined
- BaseEvent implements DomainEvent
- Registry can register and deserialize events
- 80%+ test coverage
```

#### Prompt 1.2.4: Create MongoDB Package

```
You are building a shared MongoDB package for a Go microservices project.

Create the following in shared/pkg/mongodb:

1. client.go - MongoDB client wrapper:
   - Config struct with connection settings (URI, database, timeouts)
   - NewClient function that returns connected client
   - Close function for cleanup
   - Health check function

2. repository.go - Base repository helpers:
   - Generic helper functions for CRUD operations
   - Transaction wrapper function
   - Pagination helpers

3. errors.go - MongoDB-specific error types:
   - IsNotFoundError helper
   - IsDuplicateKeyError helper

Requirements:
- Use official mongo-go-driver
- Support connection pooling
- Include retry logic for transient errors
- Proper context handling
- Unit tests with mocks

Acceptance Criteria:
- Client connects successfully
- Transactions work correctly
- Errors properly categorized
- Connection cleanup on shutdown
```

#### Prompt 1.2.5: Create Kafka Package

```
You are building a shared Kafka package for a Go microservices project.

Create the following in shared/pkg/kafka:

1. producer.go - Kafka producer:
   - Config struct (brokers, topic, retries)
   - NewProducer function
   - Publish method accepting event interface
   - Close method

2. consumer.go - Kafka consumer:
   - ConsumerConfig struct (brokers, groupID, topics)
   - NewConsumer function
   - Subscribe method with handler function
   - Start/Stop methods
   - Commit handling

3. serialization.go:
   - Serialize events to JSON
   - Deserialize events from JSON
   - Include headers for event type

Requirements:
- Use segmentio/kafka-go (standardized across all services)
- Support consumer groups
- Handle rebalancing
- Dead letter queue support
- Graceful shutdown

Acceptance Criteria:
- Producer publishes events successfully
- Consumer receives and processes events
- Consumer groups work correctly
- Graceful shutdown implemented
- Integration tests with testcontainers
```

#### Prompt 1.2.3: Create Logging Package

```
You are building a shared logging package for a Go microservices project.

REF: Requirements.md Section 3.2 (Technology Stack - Go)
REF: Task.md Section 1.2.3

Create the following in shared/pkg/logging:

1. logger.go - Structured logger wrapper:
   - Logger interface with methods: Debug, Info, Warn, Error, Fatal
   - WithField, WithFields for contextual logging
   - WithCorrelationID for distributed tracing

2. zap.go - Zap implementation:
   - NewZapLogger constructor
   - Configure JSON output for production
   - Configure console output for development
   - Log levels configurable via environment

3. middleware.go - HTTP logging middleware:
   - Log request method, path, status, duration
   - Include correlation ID from headers
   - Redact sensitive data (auth headers, passwords)

Requirements:
- Use uber-go/zap for implementation
- Support structured JSON logging
- Correlation ID propagation
- Configurable log levels
- No sensitive data in logs

Acceptance Criteria:
- Logger interface allows swapping implementations
- Zap implementation passes all tests
- Middleware logs all requests
- Correlation IDs propagate correctly
- 80%+ test coverage
```

#### Prompt 1.4.1: Create Docker Compose for Local Development

```
You are creating a Docker Compose setup for local development.

REF: Requirements.md Section 3.2 (Technology Stack)
REF: Task.md Section 1.4.1

Create docker-compose.yml in project root:

Services to include:
1. MongoDB (replica set for transaction support)
2. Kafka + Zookeeper
3. Kafka UI (for debugging)
4. Debezium Connect
5. Order service (with hot reload)
6. Payment service (with hot reload)
7. Fulfillment service (with hot reload)

Configuration:
- Use named volumes for persistence
- Network isolation between services
- Health checks for all services
- Environment variable configuration
- Port mappings for local access

Requirements:
- Single `docker-compose up` starts everything
- Hot reload for Go services (use air or similar)
- Kafka topics auto-created
- MongoDB replica set initialized
- Debezium connectors auto-registered

Acceptance Criteria:
- All services start successfully
- Services can communicate
- Data persists between restarts
- Hot reload works for development
- Kafka UI accessible at localhost:8080
```

#### Prompt 1.4.2: Configure golangci-lint

```
You are configuring golangci-lint for the Go project.

REF: Guidelines.md Coding Standards section
REF: Task.md Section 1.4.2

Create .golangci.yml in project root:

Linters to enable:
- errcheck (unchecked errors)
- govet (suspicious constructs)
- staticcheck (static analysis)
- unused (unused code)
- gosimple (simplifications)
- ineffassign (ineffectual assignments)
- gocritic (code style)
- gofmt (formatting)
- goimports (import organization)
- misspell (spelling)
- gosec (security issues)

Configuration:
- Set appropriate severity levels
- Exclude generated code
- Exclude test files from some checks
- Configure line length (120 chars)
- Set timeout for CI

Requirements:
- Zero warnings in CI (warnings are errors)
- Fast execution (<30s for full codebase)
- Clear error messages
- Exclude vendor directory

Acceptance Criteria:
- Lint passes on initial codebase
- All configured linters run
- CI integration documented
- Pre-commit hook compatible
```

#### Prompt 1.4.3: Setup Pre-commit Hooks

```
You are setting up pre-commit hooks for code quality.

REF: Task.md Section 1.4.3

Create .pre-commit-config.yaml:

Hooks to include:
1. go fmt (formatting)
2. go vet (static analysis)
3. golangci-lint (linting)
4. go mod tidy (dependency check)
5. gitleaks (secret detection)
6. conventional commits (commit message format)

Create scripts/pre-commit.sh as backup:
- Run all checks manually
- Exit with error if any fail

Requirements:
- Use pre-commit framework
- Fast execution (<10s typical)
- Skip on --no-verify flag
- Clear error messages

Acceptance Criteria:
- Hooks install with `pre-commit install`
- Commits blocked if checks fail
- Manual script works as alternative
- Secret detection catches common patterns
```

#### Prompt 1.4.4: Create Makefile

```
You are creating a Makefile for common development tasks.

REF: Task.md Section 1.4.4

Create Makefile in project root:

Targets:
- `make build` - Build all services
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linter
- `make fmt` - Format all code
- `make docker-build` - Build Docker images
- `make docker-up` - Start Docker Compose
- `make docker-down` - Stop Docker Compose
- `make clean` - Clean build artifacts
- `make help` - Show available targets

Per-service targets:
- `make build-order` - Build order service
- `make test-order` - Test order service
- etc.

Requirements:
- Use .PHONY for non-file targets
- Support parallel builds
- Colorized output
- Default target shows help

Acceptance Criteria:
- All targets work correctly
- Help text is clear
- CI can use Makefile targets
- Works on Linux and macOS
```

#### Prompt 1.2.6: Create Outbox Package

```
You are building an outbox pattern implementation for reliable event publishing.

Create the following in shared/pkg/outbox:

1. entry.go - Outbox entry model:
   - OutboxEntry struct with fields: ID, AggregateID, EventType, Payload, CreatedAt, Published
   - NewOutboxEntry constructor from DomainEvent

2. repository.go - Outbox repository interface:
   - Save(ctx, entry) error
   - MarkAsPublished(ctx, id) error
   - FindUnpublished(ctx, limit) ([]OutboxEntry, error)

3. publisher.go - Polling publisher (fallback if Debezium fails):
   - Poll outbox table periodically
   - Publish to Kafka
   - Mark as published

Requirements:
- Entries must be saved in same transaction as aggregate
- Support MongoDB and could support other DBs
- Idempotent publishing
- Include cleanup of old entries

Acceptance Criteria:
- Entries saved atomically with aggregate
- Debezium can capture changes
- Fallback polling works
- Old entries cleaned up
```

---

### Phase 2: Domain Layer Prompts

#### Prompt 2.1.6: Create Order Aggregate

```
You are implementing the Order aggregate root for an e-commerce domain.

Create the following in services/order/internal/domain/aggregate:

1. order.go - Order aggregate:
   - Order struct with: ID, CustomerID, Items, TotalAmount, Status, CreatedAt, UpdatedAt
   - Private events slice for domain events
   - NewOrder constructor (validates inputs, sets initial state)
   - AddItem, RemoveItem methods
   - MarkAsPaid, Ship, Deliver, Cancel methods
   - Events() method returning uncommitted events
   - ClearEvents() method

2. State transitions:
   - PENDING -> CREATED (on creation)
   - CREATED -> PAID (on PaymentReceived)
   - CREATED -> CANCELLED (on Cancel or PaymentFailed)
   - PAID -> SHIPPED (on OrderShipped)
   - PAID -> CANCELLED (on ShipmentFailed)
   - SHIPPED -> COMPLETED (on OrderDelivered)

Requirements:
- Enforce state machine rules
- Emit domain events on state changes
- Calculate total from items
- Immutable value objects for ID, Money
- Comprehensive validation

Acceptance Criteria:
- Cannot add items after order confirmed
- Cannot ship unpaid order
- State transitions emit events
- Invalid transitions return errors
- 80%+ unit test coverage
```

#### Prompt 2.1.7-2.1.9: Create Order Domain Events

```
You are implementing domain events for the Order aggregate.

Create the following in services/order/internal/domain/event:

1. order_created.go:
   - OrderCreated struct implementing DomainEvent
   - Fields: OrderID, CustomerID, Items, TotalAmount, OccurredAt
   - NewOrderCreated constructor

2. order_cancelled.go:
   - OrderCancelled struct implementing DomainEvent
   - Fields: OrderID, Reason, OccurredAt
   - NewOrderCancelled constructor

3. order_status_changed.go:
   - OrderStatusChanged struct implementing DomainEvent
   - Fields: OrderID, OldStatus, NewStatus, OccurredAt
   - NewOrderStatusChanged constructor

Requirements:
- Implement DomainEvent interface from shared/pkg/events
- Include all necessary data for consumers
- Events must be JSON serializable
- Use value objects for IDs

Acceptance Criteria:
- All events implement DomainEvent interface
- Events contain all required fields
- Events serialize/deserialize correctly
- Unit tests for all events
```

#### Prompt 2.2.3: Create Payment Aggregate

```
You are implementing the Payment aggregate root for the payment service.

Create the following in services/payment/internal/domain/aggregate:

1. payment.go - Payment aggregate:
   - Payment struct with: ID, OrderID, Amount, Status, PaymentMethod, CreatedAt
   - Private events slice
   - NewPayment constructor
   - Process method (simulates payment processing)
   - Refund method
   - Events() and ClearEvents() methods

2. State transitions:
   - PENDING -> PROCESSING (on Process called)
   - PROCESSING -> COMPLETED (on successful processing)
   - PROCESSING -> FAILED (on processing failure)
   - COMPLETED -> REFUNDED (on Refund)

Requirements:
- Use Money value object for amount
- Emit PaymentReceived on success
- Emit PaymentFailed on failure
- Emit RefundProcessed on refund
- Simulate processing with configurable success rate

Acceptance Criteria:
- Payment can only be refunded if completed
- Events emitted for all state changes
- Money calculations correct
- 80%+ unit test coverage
```

#### Prompt 2.3.4: Create Shipment Aggregate

```
You are implementing the Shipment aggregate root for the fulfillment service.

Create the following in services/fulfillment/internal/domain/aggregate:

1. shipment.go - Shipment aggregate:
   - Shipment struct with: ID, OrderID, TrackingNumber, Carrier, Status, CreatedAt, ShippedAt, DeliveredAt
   - Private events slice
   - NewShipment constructor
   - Ship method (assigns tracking number)
   - MarkDelivered method
   - Cancel method
   - Events() and ClearEvents() methods

2. State transitions:
   - PENDING -> PREPARING (on creation)
   - PREPARING -> SHIPPED (on Ship)
   - PREPARING -> CANCELLED (on Cancel)
   - SHIPPED -> DELIVERED (on MarkDelivered)
   - SHIPPED -> FAILED (on delivery failure)

Requirements:
- Generate tracking number on ship
- Emit OrderShipped, ShipmentFailed, OrderDelivered events
- Support multiple carriers
- Track shipping timestamps

Acceptance Criteria:
- Cannot ship cancelled shipment
- Tracking number generated on ship
- Events contain tracking info
- Timestamps correctly set
- 80%+ unit test coverage
```

---

### Phase 3: Application Layer Prompts

#### Prompt 3.1.6: Implement CreateOrderHandler

```
You are implementing the CreateOrderHandler for the order service.

Create the following in services/order/internal/application/handler:

1. create_order.go:
   - CreateOrderCommand struct with: CustomerID, Items
   - CreateOrderHandler struct with: repository, eventPublisher
   - Handle method implementing use case

Implementation:
1. Validate command inputs
2. Create Order aggregate using domain constructor
3. Save order to repository (includes outbox)
4. Return order ID

Requirements:
- Use constructor injection for dependencies
- Wrap errors with context
- No business logic (delegate to domain)
- Transaction handled by repository

Acceptance Criteria:
- Command validated before processing
- Order created via domain constructor
- Repository save includes events in outbox
- Errors properly wrapped
- Unit tests with mocked repository
```

#### Prompt 3.2.1-3.2.7: Implement Order SAGA

```
You are implementing the Order SAGA orchestrator for distributed transaction management.

Create the following in services/order/internal/application/saga:

1. order_saga.go:
   - OrderSaga struct with: repository, eventPublisher
   - HandlePaymentReceived method
   - HandlePaymentFailed method
   - HandleOrderShipped method
   - HandleShipmentFailed method
   - HandleOrderDelivered method

2. compensation.go:
   - CancelOrderCompensation
   - TriggerRefundCompensation

Event Handlers:
- PaymentReceived: Update order to PAID status
- PaymentFailed: Cancel order, emit OrderCancelled
- OrderShipped: Update order to SHIPPED status
- ShipmentFailed: Cancel order, trigger refund
- OrderDelivered: Update order to COMPLETED status

Requirements:
- Idempotent event handling
- Compensation on failures
- Log all state transitions
- Handle out-of-order events gracefully

Acceptance Criteria:
- All event handlers implemented
- Compensation triggers correctly
- Idempotency maintained
- State transitions logged
- Integration tests for all flows
```

#### Prompt 3.3.5: Create OrderCreatedEventHandler (Payment)

```
You are implementing the event handler that processes OrderCreated events in the payment service.

Create the following in services/payment/internal/application/handler:

1. order_created_handler.go:
   - OrderCreatedEventHandler struct
   - Handle method that processes OrderCreated events

Implementation:
1. Receive OrderCreated event from Kafka
2. Create Payment aggregate for the order
3. Process payment (simulate)
4. Save payment (includes outbox with PaymentReceived or PaymentFailed)

Requirements:
- Idempotent (check if payment already exists for order)
- Handle processing failures
- Emit appropriate success/failure events
- Log processing steps

Acceptance Criteria:
- Payment created on OrderCreated
- Duplicate events handled gracefully
- Success emits PaymentReceived
- Failure emits PaymentFailed
- Proper error handling
```

#### Prompt 3.4.7: Create PaymentReceivedEventHandler (Fulfillment)

```
You are implementing the event handler that processes PaymentReceived events in the fulfillment service.

Create the following in services/fulfillment/internal/application/handler:

1. payment_received_handler.go:
   - PaymentReceivedEventHandler struct
   - Handle method that processes PaymentReceived events

Implementation:
1. Receive PaymentReceived event from Kafka
2. Find or create Shipment for the order
3. Ship the order (generate tracking, update status)
4. Save shipment (includes outbox with OrderShipped)

Requirements:
- Idempotent (check shipment status)
- Handle shipping failures
- Emit OrderShipped or ShipmentFailed
- Include tracking information

Acceptance Criteria:
- Shipment created/updated on PaymentReceived
- Duplicate events handled gracefully
- Success emits OrderShipped with tracking
- Failure emits ShipmentFailed
- Proper error handling
```

---

### Phase 4: Infrastructure Layer Prompts

#### Prompt 4.1.1: Implement MongoOrderRepository

```
You are implementing the MongoDB repository for the Order aggregate.

Create the following in services/order/internal/infrastructure/persistence:

1. order_repository.go:
   - MongoOrderRepository struct with: collection, outbox
   - NewMongoOrderRepository constructor
   - Save method (with outbox transaction)
   - FindByID method
   - FindByCustomerID method
   - Update method

2. order_mapper.go:
   - Map Order aggregate to MongoDB document
   - Map MongoDB document to Order aggregate

Requirements:
- Use transactions for Save (order + outbox)
- Proper index creation
- Map domain objects to/from BSON
- Handle not found errors

Acceptance Criteria:
- Save includes outbox in same transaction
- Documents correctly mapped to/from domain
- Indexes created for common queries
- Not found returns domain error
- Integration tests with testcontainers
```

#### Prompt 4.2.3: Integrate Outbox with Repositories

```
You are integrating the outbox pattern with the MongoDB repositories.

Modify repositories to save domain events atomically:

1. Update MongoOrderRepository.Save:
   - Start transaction session
   - Save order document
   - For each domain event, save to outbox collection
   - Commit transaction
   - Clear events from aggregate

2. Create outbox collection schema:
   - id: string (UUID)
   - aggregate_id: string
   - event_type: string
   - payload: document (JSON)
   - created_at: timestamp
   - published: boolean (for polling fallback)

Requirements:
- Events saved in same transaction as aggregate
- Debezium will read from outbox collection
- Fallback polling if Debezium fails
- Cleanup old published events

Acceptance Criteria:
- Transaction ensures atomicity
- Events appear in outbox collection
- Debezium can capture changes
- Polling publisher works as fallback
- Old events cleaned up
```

#### Prompt 4.4.1-4.4.2: Configure Debezium Connector

```
You are configuring Debezium to capture outbox events and route them to Kafka.

Create the following in infrastructure/debezium:

1. order-outbox-connector.json:
   - MongoDB connector configuration
   - Capture changes from orders.outbox collection
   - Route to order.events Kafka topic
   - Transform using outbox event router

2. payment-outbox-connector.json:
   - Same for payment service
   - Route to payment.events topic

3. fulfillment-outbox-connector.json:
   - Same for fulfillment service
   - Route to fulfillment.events topic

Configuration:
- Use MongoDB Change Streams
- Event router transformation
- Route by aggregate type
- Include original event metadata

Requirements:
- Connector configs are valid JSON
- Correct MongoDB connection settings
- Proper SMT (Single Message Transform) configuration
- Error handling configuration

Acceptance Criteria:
- Connectors deploy without errors
- Changes captured from outbox collections
- Events routed to correct topics
- Event format matches domain events
- Delete operations handled (cleanup)
```

#### Prompt 4.5.1: Create Order HTTP Handlers

```
You are implementing HTTP handlers for the Order service API.

Create the following in services/order/internal/api/http:

1. handlers.go:
   - OrderHandler struct with command/query handlers
   - CreateOrder: POST /orders
   - GetOrder: GET /orders/{id}
   - ListOrders: GET /orders?customer_id=xxx
   - CancelOrder: POST /orders/{id}/cancel

2. dto.go:
   - CreateOrderRequest/Response
   - GetOrderResponse
   - ListOrdersResponse
   - ErrorResponse

3. routes.go:
   - Configure routes using chi or gorilla/mux
   - Apply middleware

Requirements:
- Validate all inputs
- Return appropriate HTTP status codes
- Consistent error response format
- Include request logging
- CORS if needed

Acceptance Criteria:
- All endpoints return correct status codes
- Invalid inputs return 400
- Not found returns 404
- Server errors return 500
- Responses match DTO schemas
```

#### Prompt 1.2.7: Create Shared Value Objects

```
You are implementing shared value objects used across all services.

REF: Requirements.md Section 3.4 (Service Architecture)
REF: Task.md Section 1.2.7

Create the following in shared/pkg/domain:

1. orderid.go - OrderID value object:
   - OrderID type (string wrapper)
   - NewOrderID() generates UUID
   - ParseOrderID(string) validates and creates
   - String() method
   - IsEmpty() method

2. money.go - Money value object:
   - Money struct with Amount (int64 cents) and Currency
   - NewMoney(amount, currency) constructor
   - Add, Subtract methods (return new Money)
   - Equals method
   - Validate currency codes (ISO 4217)

3. customerid.go - CustomerID value object:
   - Similar to OrderID

4. correlationid.go - CorrelationID for distributed tracing:
   - CorrelationID type
   - NewCorrelationID() generates UUID
   - FromContext/ToContext for propagation

Requirements:
- Value objects are immutable
- Validation on construction
- JSON serialization support
- BSON serialization for MongoDB
- No business logic, only validation

Acceptance Criteria:
- All value objects immutable
- Invalid values rejected on construction
- Serialization works correctly
- Used consistently across services
- 80%+ unit test coverage
```

#### Prompt 4.6.1: Create Order Service Main

```
You are implementing the main entry point for the Order service.

REF: Requirements.md Section 3.4 (Service Architecture)
REF: Task.md Section 4.6.1

Create the following in services/order/cmd/main.go:

1. Configuration loading (env vars or config file)
2. Dependency injection (wire up all components)
3. Start HTTP server
4. Start Kafka consumer
5. Graceful shutdown handling

Components to wire:
- MongoDB client and repositories
- Outbox repository
- Kafka consumer for incoming events
- Event handlers (SAGA)
- HTTP handlers
- Health check endpoint

Requirements:
- Clean dependency injection
- Graceful shutdown on SIGTERM/SIGINT
- Health endpoint at /health
- Structured logging
- Configuration via environment variables

Acceptance Criteria:
- Service starts and connects to dependencies
- HTTP endpoints accessible
- Kafka consumer receives events
- Graceful shutdown works
- Health check reports status
```

#### Prompt 5.7.1-5.7.12: Setup Istio Gateway API

```
You are setting up Kubernetes Gateway API with Istio for the microservices platform.

REF: Requirements.md Section 3.3 (Gateway API + Istio)
REF: Task.md Section 5.7

Create infrastructure/k8s/base/gateway-api:

1. istio-install.yaml - Istio installation config:
   - IstioOperator resource
   - Enable Gateway API support
   - Configure telemetry

2. gateway-class.yaml:
   - GatewayClass resource
   - controllerName: istio.io/gateway-controller

3. gateway.yaml:
   - Gateway resource
   - Listeners for HTTP (80) and HTTPS (443)
   - TLS configuration with cert-manager
   - Reference to GatewayClass

4. routes/order-route.yaml:
   - HTTPRoute for /api/orders/*
   - Backend reference to order-service
   - Path prefix matching

5. routes/payment-route.yaml:
   - HTTPRoute for /api/payments/*
   - Backend reference to payment-service

6. routes/fulfillment-route.yaml:
   - HTTPRoute for /api/shipments/*
   - Backend reference to fulfillment-service

7. auth/request-authentication.yaml:
   - RequestAuthentication for JWT validation
   - JWKS URI configuration

8. auth/authorization-policy.yaml:
   - AuthorizationPolicy for access control

9. filters/rate-limit.yaml:
   - EnvoyFilter for rate limiting
   - 1000 req/min per client

10. filters/correlation-id.yaml:
    - EnvoyFilter to inject X-Correlation-ID
    - Generate UUID if not present

Requirements:
- Use Kubernetes Gateway API v1 spec
- Istio 1.20+ for full Gateway API support
- cert-manager for TLS certificate management
- All resources in infrastructure/k8s/base

Acceptance Criteria:
- Istio installed and healthy
- Gateway API resources created
- Routes work for all services
- JWT authentication enforced
- Rate limiting working
- Correlation IDs injected
- TLS certificates auto-renewed
- Istio telemetry visible in Prometheus
```

#### Prompt 4.8.1: Implement Outbox Polling Fallback

```
You are implementing a polling-based fallback for the outbox pattern.

REF: Requirements.md Section 4.6 (Outbox Polling Fallback)
REF: Task.md Section 4.8

Create the following in shared/pkg/outbox:

1. polling_publisher.go:
   - PollingPublisher struct
   - Start/Stop methods
   - Configurable poll interval (default 5s)
   - Batch size limit (max 100)
   - Publish to Kafka

2. debezium_health.go:
   - DebeziumHealthChecker interface
   - HTTP health check implementation
   - Fallback activation logic

3. cleanup.go:
   - Cleanup old published events
   - Configurable retention (default 24h)
   - Run on schedule

Requirements:
- Idempotent publishing (check event ID)
- Maintain ordering per aggregate
- Graceful shutdown
- Metrics for monitoring
- Automatic fallback activation

Acceptance Criteria:
- Polling activates when Debezium unhealthy
- Events published in order
- No duplicate events
- Cleanup removes old events
- Metrics track fallback status
- 80%+ test coverage
```

---

### Phase 5: DevOps Prompts

#### Prompt 5.1.1: Create Order Service Dockerfile

```
You are creating an optimized Dockerfile for the Order service.

Create services/order/Dockerfile:

Multi-stage build:
1. Builder stage:
   - Use golang:1.21-alpine
   - Set WORKDIR
   - Copy go.mod/go.sum, download dependencies
   - Copy source code
   - Build binary with CGO_ENABLED=0

2. Runtime stage:
   - Use alpine:3.18 or scratch
   - Add ca-certificates
   - Copy binary from builder
   - Set non-root user
   - Expose port
   - Set entrypoint

Requirements:
- Minimal image size (<50MB)
- Non-root user for security
- Health check instruction
- Build args for version info
- .dockerignore file

Acceptance Criteria:
- Image builds successfully
- Image size under 50MB
- Runs as non-root
- Container starts correctly
- Health check works
```

#### Prompt 5.2.1-5.2.10: Create GitHub Actions CI Pipeline

```
You are creating the GitHub Actions CI pipeline for the microservices.

Create .github/workflows/ci-order.yml:

Jobs:
1. test:
   - Checkout code
   - Setup Go
   - Run go mod download
   - Run go vet
   - Run golangci-lint
   - Run unit tests with coverage
   - Upload coverage report

2. security:
   - OWASP dependency check
   - SonarQube scan
   - Upload reports

3. build (needs: test, security):
   - Build Docker image
   - Run Trivy scan
   - Push to registry (only on main)

4. notify:
   - Send email on success/failure

Requirements:
- Trigger on push to main/develop
- Trigger on PR to main
- Cache Go modules
- Fail fast on security issues
- Store artifacts

Acceptance Criteria:
- Pipeline runs on PR
- Tests execute and report coverage
- Security scans run
- Docker image built
- Image pushed only on main
- Notifications sent
```

#### Prompt 5.3.1: Create Terraform AKS Module

```
You are creating a Terraform module to provision Azure AKS.

Create infrastructure/terraform/modules/aks:

1. main.tf:
   - Azure Kubernetes Service cluster
   - Default node pool
   - User node pool for workloads
   - Managed identity
   - Azure CNI networking

2. variables.tf:
   - cluster_name
   - location
   - resource_group_name
   - node_count
   - node_size
   - kubernetes_version

3. outputs.tf:
   - cluster_id
   - kube_config
   - cluster_fqdn
   - node_resource_group

Requirements:
- Enable RBAC
- Configure autoscaling
- Azure Monitor integration
- Network policy enabled
- Private cluster option

Acceptance Criteria:
- Module creates valid AKS cluster
- Node pools configured correctly
- Autoscaling works
- Outputs available for other modules
- Passes terraform validate
```

#### Prompt 5.4.1: Create Order Service Helm Chart

```
You are creating a Helm chart for the Order service.

Create infrastructure/helm/charts/order-service:

1. Chart.yaml - Chart metadata
2. values.yaml - Default values
3. templates/:
   - deployment.yaml
   - service.yaml
   - configmap.yaml
   - secret.yaml (external secret reference)
   - hpa.yaml
   - pdb.yaml
   - networkpolicy.yaml
   - serviceaccount.yaml

Values to parameterize:
- image.repository, image.tag
- replicas, resources
- env vars (config)
- service port
- ingress settings
- HPA min/max replicas

Requirements:
- Follow Helm best practices
- Include health checks
- Resource limits required
- Pod anti-affinity for HA
- ConfigMap for non-sensitive config

Acceptance Criteria:
- helm lint passes
- helm template renders correctly
- Deploys successfully to cluster
- Health checks work
- HPA scales correctly
```

#### Prompt 5.5.1-5.5.5: Create GitHub Actions CD Workflow

```
You are creating a GitHub Actions CD workflow for continuous deployment.

Create .github/workflows/cd-deploy.yml:

Triggers:
1. workflow_run: Triggered when CI workflow completes successfully
2. workflow_dispatch: Manual trigger with image tag input

Jobs:
1. update-manifests:
   - Checkout repository with PAT for push access
   - Update image tag in Helm values or kustomize
   - Use sed or yq for updates
   - Commit and push changes

2. notify:
   - Post status to Slack/email
   - Include deployment details

Requirements:
- PAT_TOKEN secret for Git push access
- Atomic updates (single commit)
- Rollback procedure documented
- Use github-actions[bot] for commits

Acceptance Criteria:
- Workflow triggers on CI success
- Manifests updated correctly
- Changes pushed to repo
- ArgoCD detects changes
- Notifications sent
```

#### Prompt 5.6.2: Create ArgoCD Application Manifests

```
You are creating ArgoCD Application manifests for GitOps deployment.

Create argocd/applications/:

1. order-service.yaml:
   - Application resource for Order service
   - Source: GitHub repo, path to Helm chart
   - Destination: AKS cluster, ecommerce namespace
   - Sync policy: automated for staging

2. payment-service.yaml:
   - Same structure for Payment service

3. fulfillment-service.yaml:
   - Same structure for Fulfillment service

4. infrastructure.yaml:
   - App-of-apps for shared infrastructure
   - Kafka, MongoDB, Debezium, monitoring

Requirements:
- Use Helm as source type
- Configure value files per environment
- Enable auto-prune and self-heal for staging
- Manual sync for production
- Health checks configured

Acceptance Criteria:
- Applications sync successfully
- Helm values applied correctly
- Auto-sync works for staging
- Manual approval for prod
- Health status accurate
```

---

### Phase 6: Observability Prompts

#### Prompt 6.1.3: Add Custom Prometheus Metrics

```
You are adding custom Prometheus metrics to the microservices.

Add metrics to each service:

1. metrics/metrics.go:
   - Define metrics using prometheus/client_golang
   - Counter: requests_total, orders_created_total, payments_processed_total
   - Histogram: request_duration_seconds, order_processing_duration_seconds
   - Gauge: active_connections, pending_orders

2. Instrument handlers:
   - Wrap HTTP handlers with metrics middleware
   - Record request count, duration, status
   - Record business metrics in domain operations

3. Expose /metrics endpoint:
   - Prometheus scrape endpoint
   - Include Go runtime metrics

Requirements:
- Consistent naming (snake_case)
- Include labels (service, method, status)
- Histogram buckets for latency
- Document each metric

Acceptance Criteria:
- /metrics returns valid Prometheus format
- All handlers instrumented
- Business metrics recorded
- Labels consistent across services
- Prometheus can scrape metrics
```

#### Prompt 6.2.2-6.2.5: Create Grafana Dashboards

```
You are creating Grafana dashboards for the e-commerce platform.

Create infrastructure/grafana/dashboards/:

1. service-overview.json:
   - Request rate per service
   - Error rate per service
   - Latency percentiles (p50, p90, p99)
   - Active connections

2. order-pipeline.json:
   - Orders created over time
   - Orders by status (gauge)
   - Order processing duration
   - SAGA success/failure rate

3. kafka-metrics.json:
   - Consumer lag per topic
   - Message throughput
   - Partition distribution
   - Consumer group status

4. infrastructure.json:
   - CPU usage per pod
   - Memory usage per pod
   - Network I/O
   - Disk usage

Requirements:
- Use variables for service/namespace selection
- Include alert thresholds as annotations
- Consistent time ranges
- Mobile-friendly layout

Acceptance Criteria:
- Dashboards import without errors
- Variables work correctly
- Panels show data from Prometheus
- Thresholds visible
- Refresh rate appropriate
```

#### Prompt 6.3.1-6.3.6: Configure AlertManager

```
You are configuring Prometheus AlertManager for the platform.

Create infrastructure/prometheus/:

1. alertmanager.yml:
   - Configure receivers (email, Slack)
   - Route rules by severity
   - Inhibition rules
   - Repeat interval settings

2. alert-rules.yml:
   - HighErrorRate: >1% 5xx errors for 5min
   - KafkaConsumerLag: lag >1000 for 10min
   - PodRestartLoop: >3 restarts in 10min
   - HighMemoryUsage: >85% for 15min
   - HighCPUUsage: >85% for 15min
   - ServiceDown: probe fails for 2min

Requirements:
- Group alerts by service
- Include runbook links
- Configure silence during maintenance
- Test alert firing

Acceptance Criteria:
- Alerts fire correctly
- Notifications sent to configured channels
- Grouping works
- Silences can be applied
- Runbook links included
```

---

## Quick Reference

### Common Commands

```bash
# Run tests
go test ./... -v -cover

# Run linter
golangci-lint run

# Build service
go build -o bin/service ./cmd/main.go

# Build Docker image
docker build -t service:latest .

# Run locally with Docker Compose
docker-compose up -d

# Deploy with Helm
helm upgrade --install order-service ./charts/order-service

# Check ArgoCD sync status
argocd app get order-service

# View logs
kubectl logs -f deployment/order-service

# Port forward
kubectl port-forward svc/order-service 8080:80
```

### Checklist Before PR

- [ ] All tests pass locally
- [ ] Lint issues resolved
- [ ] Code reviewed by peer
- [ ] Acceptance criteria met
- [ ] Documentation updated
- [ ] No secrets in code
- [ ] Docker image builds
- [ ] Helm chart lints

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-11-24 | AI Agent | Initial version |
