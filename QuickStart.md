# Quick Start Guide

## 🚀 Getting Started in 5 Minutes

### Prerequisites Check
```bash
# Verify your environment
./scripts/check-prerequisites.sh

# Expected output: All tools ✅
```

### Local Development Setup
```bash
# 1. Clone and setup
git clone <repo-url>
cd GitOps

# 2. Start local stack
cd ecommerce-platform
docker-compose up -d

# 3. Verify services
make health-check

# 4. Run tests
make test
```

### First Development Task
```bash
# Get next task
./scripts/get-next-task.sh

# Example output: 1.2.1

# Read task details
cat tasks.json | jq '.phases[].sections[].tasks[] | select(.id=="1.2.1")'

# Start working
git checkout -b feature/task-1.2.1
```

---

## 📋 Development Workflow

### Daily Development Cycle
1. **Morning**: Check task status, pull latest changes
2. **Planning**: Review assigned tasks and dependencies
3. **Coding**: Implement task following guidelines
4. **Testing**: Run tests, verify acceptance criteria
5. **Commit**: Update CHANGELOG.md, commit with convention

### Task Completion Checklist
- [ ] Code implemented according to Guidelines.md
- [ ] Unit tests pass (`go test ./...`)
- [ ] Integration tests pass (if applicable)
- [ ] Linting passes (`golangci-lint run`)
- [ ] CHANGELOG.md updated
- [ ] Commit follows convention: `[TASK-X.X.X] Description`

### Common Commands
```bash
# Development
make build          # Build all services
make test          # Run all tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make fmt           # Format code

# Local stack
docker-compose up -d    # Start services
docker-compose down     # Stop services
docker-compose logs -f  # View logs

# Git workflow
git checkout -b feature/task-X.X.X
git add .
git commit -m "[TASK-X.X.X] Brief description"
git push origin feature/task-X.X.X
```

---

## 🔧 Troubleshooting Quick Fixes

### Service Won't Start
```bash
# Check logs
docker-compose logs <service-name>

# Check health
curl http://localhost:8080/health

# Restart service
docker-compose restart <service-name>
```

### Tests Failing
```bash
# Run specific test
go test ./internal/domain/... -v

# Debug with verbose output
go test -v -run TestSpecificFunction

# Check coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Database Connection Issues
```bash
# Check MongoDB
docker exec -it mongodb mongosh --eval "db.stats()"

# Reset replica set
docker-compose down -v
docker-compose up -d mongo-init
```

### Kafka Issues
```bash
# Check topics
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092

# Check consumer groups
docker exec -it kafka kafka-consumer-groups --list --bootstrap-server localhost:9092
```

---

## 📚 Key Resources

### Documentation
- **Requirements.md**: What we're building
- **Task.md**: Current task breakdown
- **Guidelines.md**: How to implement
- **Setup.md**: Environment setup

### Important Files
```
ecommerce-platform/
├── services/order/        # Order service
├── services/payment/      # Payment service
├── services/fulfillment/  # Fulfillment service
├── shared/pkg/           # Shared libraries
└── infrastructure/       # Terraform, Helm, K8s
```

### Getting Help
1. Check existing documentation first
2. Review similar completed tasks
3. Ask team member for guidance
4. Create issue with detailed description

---

## 🎯 Success Metrics

### Code Quality
- Test coverage > 80%
- Lint score: A or B
- No critical security issues
- Documentation updated

### Development Velocity
- Tasks completed on schedule
- Code reviews within 24 hours
- CI/CD pipeline passing
- Zero production incidents

### Learning Goals
- Understand microservices architecture
- Master Go language patterns
- Learn Kubernetes deployment
- Gain DevOps experience