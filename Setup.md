# Setup & Environment Guide

## Document Cross-References

| Document | Purpose |
|----------|---------|
| **Requirements.md** | Technology stack specifications |
| **Plan.md** | Repository structure |
| **Task.md** | Setup tasks (1.0.x, 1.4.x) |
| **Guidelines.md** | Coding standards to configure |
| **tasks.json** | Machine-readable task queries |

---

## Quick Start for AI Agents

```bash
# 1. Clone and navigate
cd /path/to/GitOps

# 2. Verify prerequisites
./scripts/check-prerequisites.sh

# 3. Initialize repository (Task 1.0.1-1.0.6)
git init
git add .
git commit -m "[TASK-1.0.4] Initial commit with planning documents"

# 4. Start local development environment
cd ecommerce-platform
docker-compose up -d

# 5. Verify services
make health-check
```

---

## Prerequisites

### Required Software

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Go** | 1.21+ | Application runtime | [go.dev/dl](https://go.dev/dl/) |
| **Docker** | 24.0+ | Containerization | [docker.com](https://docker.com) |
| **Docker Compose** | 2.20+ | Local orchestration | Included with Docker Desktop |
| **kubectl** | 1.28+ | Kubernetes CLI | `brew install kubectl` |
| **Helm** | 3.12+ | K8s package manager | `brew install helm` |
| **Terraform** | 1.5+ | Infrastructure as Code | `brew install terraform` |
| **Azure CLI** | 2.50+ | Azure management | `brew install azure-cli` |
| **golangci-lint** | 1.54+ | Go linter | `brew install golangci-lint` |
| **pre-commit** | 3.3+ | Git hooks | `pip install pre-commit` |

### Verification Script

```bash
#!/bin/bash
# scripts/check-prerequisites.sh

echo "Checking prerequisites..."

check_version() {
    local cmd=$1
    local min_version=$2
    local actual=$($cmd --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+' | head -1)

    if [ -z "$actual" ]; then
        echo "❌ $cmd not installed"
        return 1
    fi
    echo "✅ $cmd: $actual (min: $min_version)"
}

check_version "go version | grep -oE 'go[0-9]+\.[0-9]+'" "1.21"
check_version "docker" "24.0"
check_version "docker-compose" "2.20"
check_version "kubectl" "1.28"
check_version "helm" "3.12"
check_version "terraform" "1.5"
check_version "az" "2.50"
check_version "golangci-lint" "1.54"
check_version "pre-commit" "3.3"

echo ""
echo "Prerequisites check complete."
```

---

## Environment Configuration

### Environment Variables

```bash
# .env.example (DO NOT commit actual .env)

# Application
APP_ENV=development
LOG_LEVEL=debug
LOG_FORMAT=console  # console | json

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=ecommerce
MONGO_REPLICA_SET=rs0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=ecommerce-local

# Debezium
DEBEZIUM_CONNECT_URL=http://localhost:8083

# Azure (for production)
AZURE_SUBSCRIPTION_ID=
AZURE_TENANT_ID=
AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=

# Container Registry
REGISTRY_URL=
REGISTRY_USERNAME=
REGISTRY_PASSWORD=
```

### Secret Management

| Environment | Strategy |
|-------------|----------|
| **Local** | `.env` files (gitignored) |
| **CI/CD** | GitHub Secrets |
| **Kubernetes** | Azure Key Vault + External Secrets Operator |

```yaml
# Example: External Secrets for K8s
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
spec:
  secretStoreRef:
    name: azure-keyvault
    kind: ClusterSecretStore
  target:
    name: app-secrets
  data:
    - secretKey: mongo-uri
      remoteRef:
        key: ecommerce-mongo-uri
```

---

## Repository Layout

```
GitOps/
├── .env.example              # Environment template
├── .gitignore               # Ignore patterns
├── .editorconfig            # Editor settings
├── .golangci.yml            # Linter config
├── .pre-commit-config.yaml  # Git hooks
├── Makefile                 # Build commands
├── docker-compose.yml       # Local dev stack
├── scripts/
│   ├── check-prerequisites.sh
│   ├── init-mongo-replica.sh
│   └── register-debezium-connectors.sh
├── CHANGELOG.md
├── Requirements.md
├── Plan.md
├── Task.md
├── Guidelines.md
├── Setup.md                 # This file
├── tasks.json               # Machine-readable tasks
└── ecommerce-platform/      # Application code
    ├── services/
    │   ├── order/
    │   ├── payment/
    │   └── fulfillment/
    ├── shared/
    │   └── pkg/
    ├── infrastructure/
    │   ├── terraform/
    │   ├── helm/
    │   └── k8s/
    ├── .github/
    │   └── workflows/
    └── argocd/
```

---

## Local Development Stack

### Docker Compose Services

```yaml
# docker-compose.yml
version: '3.8'

services:
  # MongoDB Replica Set
  mongo:
    image: mongo:7.0
    command: ["--replSet", "rs0", "--bind_ip_all"]
    ports:
      - "27017:27017"
    volumes:
      - mongo-data:/data/db
    healthcheck:
      test: mongosh --eval "rs.status()"
      interval: 10s
      timeout: 5s
      retries: 5

  mongo-init:
    image: mongo:7.0
    depends_on:
      mongo:
        condition: service_healthy
    command: >
      mongosh --host mongo --eval "rs.initiate({_id: 'rs0', members: [{_id: 0, host: 'mongo:27017'}]})"

  # Kafka + Zookeeper
  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    healthcheck:
      test: kafka-topics --bootstrap-server localhost:9092 --list
      interval: 10s
      timeout: 5s
      retries: 5

  # Kafka UI (debugging)
  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    ports:
      - "8080:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:9092
    depends_on:
      - kafka

  # Debezium Connect
  debezium:
    image: debezium/connect:2.4
    ports:
      - "8083:8083"
    environment:
      BOOTSTRAP_SERVERS: kafka:9092
      GROUP_ID: debezium-connect
      CONFIG_STORAGE_TOPIC: debezium-configs
      OFFSET_STORAGE_TOPIC: debezium-offsets
      STATUS_STORAGE_TOPIC: debezium-status
    depends_on:
      kafka:
        condition: service_healthy
    healthcheck:
      test: curl -f http://localhost:8083/connectors
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mongo-data:

networks:
  default:
    name: ecommerce-network
```

### Starting the Stack

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f kafka

# Stop all
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

---

## Git Configuration

### Branch Strategy

| Branch | Purpose | Protection |
|--------|---------|------------|
| `main` | Production-ready code | Required reviews, CI pass |
| `develop` | Integration branch | CI pass |
| `feature/*` | Feature development | None |
| `hotfix/*` | Production fixes | Expedited review |

### Branch Protection Rules (GitHub)

```json
{
  "main": {
    "required_pull_request_reviews": {
      "required_approving_review_count": 1,
      "dismiss_stale_reviews": true
    },
    "required_status_checks": {
      "strict": true,
      "contexts": ["test", "lint", "security"]
    },
    "enforce_admins": true,
    "restrictions": null
  }
}
```

### Commit Convention

```
[TASK-X.X.X] Brief description (50 chars max)

- Detailed change 1
- Detailed change 2

Updates CHANGELOG.md

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: Go Format
        entry: gofmt -w
        language: system
        types: [go]

      - id: go-vet
        name: Go Vet
        entry: go vet ./...
        language: system
        pass_filenames: false

      - id: golangci-lint
        name: GolangCI Lint
        entry: golangci-lint run
        language: system
        pass_filenames: false

      - id: go-mod-tidy
        name: Go Mod Tidy
        entry: go mod tidy
        language: system
        pass_filenames: false

  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks

  - repo: https://github.com/compilerla/conventional-pre-commit
    rev: v3.0.0
    hooks:
      - id: conventional-pre-commit
        stages: [commit-msg]
```

---

## Azure AKS Setup

### Prerequisites

```bash
# Login to Azure
az login

# Set subscription
az account set --subscription "<subscription-id>"

# Register providers
az provider register --namespace Microsoft.ContainerService
az provider register --namespace Microsoft.OperationsManagement
```

### Terraform Initialization

```bash
cd infrastructure/terraform/environments/dev

# Initialize
terraform init

# Plan
terraform plan -out=tfplan

# Apply
terraform apply tfplan
```

### Kubernetes Context

```bash
# Get credentials
az aks get-credentials \
  --resource-group ecommerce-rg \
  --name ecommerce-aks-dev

# Verify
kubectl get nodes
kubectl cluster-info
```

---

## Istio Installation

```bash
# Install Istio CLI
brew install istioctl

# Install Istio with Gateway API support
istioctl install --set profile=default \
  --set values.pilot.env.PILOT_ENABLE_GATEWAY_API=true

# Verify
kubectl get pods -n istio-system
istioctl analyze

# Enable sidecar injection for namespace
kubectl label namespace ecommerce istio-injection=enabled
```

---

## Verification Commands

### Health Checks

```bash
# Check all local services
make health-check

# Or manually:
curl http://localhost:27017  # MongoDB
curl http://localhost:9092   # Kafka (will fail but confirms port)
curl http://localhost:8083/connectors  # Debezium
curl http://localhost:8080   # Kafka UI
```

### Smoke Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint

# Build all services
make build
```

---

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| MongoDB connection refused | Replica set not initialized | Run `mongo-init` container |
| Kafka consumer lag | Consumer not started | Check consumer group ID |
| Debezium connector fails | MongoDB not in replica set mode | Verify `--replSet` flag |
| Permission denied (Docker) | User not in docker group | `sudo usermod -aG docker $USER` |
| Go module errors | Wrong Go version | Check `go version` >= 1.21 |

### Debug Commands

```bash
# MongoDB
mongosh --eval "rs.status()"
mongosh --eval "db.adminCommand({listDatabases: 1})"

# Kafka
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092
docker exec -it kafka kafka-consumer-groups --list --bootstrap-server localhost:9092

# Debezium
curl http://localhost:8083/connectors
curl http://localhost:8083/connectors/<name>/status

# Kubernetes
kubectl get pods -A
kubectl describe pod <pod-name>
kubectl logs <pod-name> -c <container>
```

---

## AI Agent Bootstrap

### Agent Environment Setup

```bash
# Clone repository
git clone <repo-url>
cd GitOps

# Read tasks.json for current state
cat tasks.json | jq '.phases[0].sections[0].tasks[] | select(.status=="pending")'

# Get next unblocked task
NEXT_TASK=$(cat tasks.json | jq -r '
  .phases[].sections[].tasks[]
  | select(.status=="pending")
  | select(all(.dependencies[]; . as $dep |
      (.phases[].sections[].tasks[] | select(.id==$dep) | .status) == "completed"
    ))
  | .id' | head -1)

echo "Next task: $NEXT_TASK"
```

### Agent Verification Protocol

Before marking a task complete, agents must verify:

```bash
# 1. Tests pass
go test ./... -v

# 2. Lint passes
golangci-lint run

# 3. Coverage meets threshold
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{print $3}'

# 4. Build succeeds
go build ./...

# 5. Definition of Done checklist complete
# (manual verification against tasks.json definitionOfDone array)
```

### Agent Output Format

When completing a task, agents should output:

```json
{
  "taskId": "1.2.3",
  "status": "completed",
  "verification": {
    "testsPass": true,
    "lintPass": true,
    "coverage": 85.2,
    "buildPass": true,
    "definitionOfDone": ["✅ Item 1", "✅ Item 2", "✅ Item 3"]
  },
  "filesModified": [
    "shared/pkg/logging/logger.go",
    "shared/pkg/logging/zap.go",
    "shared/pkg/logging/logger_test.go"
  ],
  "commit": {
    "hash": "abc123",
    "message": "[TASK-1.2.3] Create logging package"
  },
  "changelogEntry": "..."
}
```

---

## Version Matrix

| Component | Dev | Staging | Prod |
|-----------|-----|---------|------|
| Go | 1.21 | 1.21 | 1.21 |
| MongoDB | 7.0 | 7.0 | 7.0 |
| Kafka | 3.5 | 3.5 | 3.5 |
| Kubernetes | 1.28 | 1.28 | 1.28 |
| Istio | 1.20 | 1.20 | 1.20 |
| Terraform | 1.5 | 1.5 | 1.5 |

---

## Rollback & Recovery Procedures

### ArgoCD Rollback

ArgoCD provides built-in rollback capabilities for GitOps deployments:

```bash
# List application history
argocd app history ecommerce-order

# Rollback to specific revision
argocd app rollback ecommerce-order <revision-id>

# Rollback via UI
# Navigate to Application → History & Rollback → Select revision → Rollback

# Sync to previous Git commit (alternative approach)
argocd app sync ecommerce-order --revision <git-commit-sha>
```

### Manual Kubernetes Rollback

```bash
# View deployment history
kubectl rollout history deployment/order-service -n ecommerce

# Rollback to previous revision
kubectl rollout undo deployment/order-service -n ecommerce

# Rollback to specific revision
kubectl rollout undo deployment/order-service -n ecommerce --to-revision=2

# Check rollout status
kubectl rollout status deployment/order-service -n ecommerce
```

### Database Rollback Strategy

| Scenario | Strategy |
|----------|----------|
| Schema migration failure | Use versioned migrations with rollback scripts |
| Data corruption | Restore from Azure Backup (point-in-time recovery) |
| CDC lag issues | Reset Debezium connector offset |

```bash
# Reset Debezium connector
curl -X DELETE http://localhost:8083/connectors/order-outbox-connector
# Re-register with updated config

# MongoDB point-in-time restore (Azure)
az cosmosdb mongodb restorable-database-accounts show \
  --location eastus \
  --name ecommerce-mongo
```

### Kafka Consumer Offset Reset

```bash
# Reset consumer group to earliest
kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group order-service \
  --reset-offsets --to-earliest \
  --topic order-events \
  --execute

# Reset to specific timestamp
kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group order-service \
  --reset-offsets --to-datetime 2025-01-15T00:00:00.000 \
  --topic order-events \
  --execute
```

### Recovery Runbook

1. **Service Failure**: Check logs → ArgoCD rollback → Verify health
2. **Database Corruption**: Stop writes → Point-in-time restore → Replay events
3. **Kafka Issues**: Check consumer lag → Reset offsets if needed → Verify processing
4. **Debezium Failure**: Check connector status → Delete and recreate connector
5. **Full Cluster Recovery**: Terraform apply → ArgoCD sync → Health checks
