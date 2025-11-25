.PHONY: help build test lint fmt clean docker-up docker-down docker-logs \
        build-order build-payment build-fulfillment \
        test-order test-payment test-fulfillment test-shared \
        run-order run-payment run-fulfillment \
        tidy coverage health-check

# Default target
.DEFAULT_GOAL := help

# Colors for terminal output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Docker parameters
DOCKER_COMPOSE=docker compose

# Directories
SERVICES_DIR=services
SHARED_DIR=shared
ORDER_DIR=$(SERVICES_DIR)/order
PAYMENT_DIR=$(SERVICES_DIR)/payment
FULFILLMENT_DIR=$(SERVICES_DIR)/fulfillment

## help: Show this help message
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## build: Build all services
build: build-order build-payment build-fulfillment ## Build all services
	@echo "${GREEN}All services built successfully${RESET}"

## build-order: Build order service
build-order: ## Build order service
	@echo "${CYAN}Building order service...${RESET}"
	cd $(ORDER_DIR) && $(GOBUILD) -o bin/order ./cmd/...

## build-payment: Build payment service
build-payment: ## Build payment service
	@echo "${CYAN}Building payment service...${RESET}"
	cd $(PAYMENT_DIR) && $(GOBUILD) -o bin/payment ./cmd/...

## build-fulfillment: Build fulfillment service
build-fulfillment: ## Build fulfillment service
	@echo "${CYAN}Building fulfillment service...${RESET}"
	cd $(FULFILLMENT_DIR) && $(GOBUILD) -o bin/fulfillment ./cmd/...

## test: Run all tests
test: test-shared test-order test-payment test-fulfillment ## Run all tests
	@echo "${GREEN}All tests passed${RESET}"

## test-shared: Run shared package tests
test-shared: ## Run shared package tests
	@echo "${CYAN}Testing shared packages...${RESET}"
	cd $(SHARED_DIR) && $(GOTEST) -v -race ./...

## test-order: Run order service tests
test-order: ## Run order service tests
	@echo "${CYAN}Testing order service...${RESET}"
	cd $(ORDER_DIR) && $(GOTEST) -v -race ./...

## test-payment: Run payment service tests
test-payment: ## Run payment service tests
	@echo "${CYAN}Testing payment service...${RESET}"
	cd $(PAYMENT_DIR) && $(GOTEST) -v -race ./...

## test-fulfillment: Run fulfillment service tests
test-fulfillment: ## Run fulfillment service tests
	@echo "${CYAN}Testing fulfillment service...${RESET}"
	cd $(FULFILLMENT_DIR) && $(GOTEST) -v -race ./...

## coverage: Run tests with coverage report
coverage: ## Run tests with coverage report
	@echo "${CYAN}Running tests with coverage...${RESET}"
	cd $(SHARED_DIR) && $(GOTEST) -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: shared/coverage.html${RESET}"

## lint: Run linter
lint: ## Run golangci-lint
	@echo "${CYAN}Running linter...${RESET}"
	cd $(SHARED_DIR) && golangci-lint run ./...
	cd $(ORDER_DIR) && golangci-lint run ./...
	cd $(PAYMENT_DIR) && golangci-lint run ./...
	cd $(FULFILLMENT_DIR) && golangci-lint run ./...
	@echo "${GREEN}Lint passed${RESET}"

## fmt: Format all Go files
fmt: ## Format Go code
	@echo "${CYAN}Formatting code...${RESET}"
	$(GOFMT) -w $(SHARED_DIR)
	$(GOFMT) -w $(ORDER_DIR)
	$(GOFMT) -w $(PAYMENT_DIR)
	$(GOFMT) -w $(FULFILLMENT_DIR)
	@echo "${GREEN}Code formatted${RESET}"

## tidy: Run go mod tidy on all modules
tidy: ## Tidy Go modules
	@echo "${CYAN}Tidying Go modules...${RESET}"
	cd $(SHARED_DIR) && $(GOMOD) tidy
	cd $(ORDER_DIR) && $(GOMOD) tidy
	cd $(PAYMENT_DIR) && $(GOMOD) tidy
	cd $(FULFILLMENT_DIR) && $(GOMOD) tidy
	@echo "${GREEN}Modules tidied${RESET}"

## clean: Clean build artifacts
clean: ## Clean build artifacts
	@echo "${CYAN}Cleaning build artifacts...${RESET}"
	rm -rf $(ORDER_DIR)/bin
	rm -rf $(PAYMENT_DIR)/bin
	rm -rf $(FULFILLMENT_DIR)/bin
	rm -f */coverage.out */coverage.html
	@echo "${GREEN}Clean complete${RESET}"

## docker-up: Start Docker Compose stack
docker-up: ## Start Docker Compose services
	@echo "${CYAN}Starting Docker Compose stack...${RESET}"
	$(DOCKER_COMPOSE) up -d
	@echo "${GREEN}Stack started. Kafka UI available at http://localhost:8080${RESET}"

## docker-down: Stop Docker Compose stack
docker-down: ## Stop Docker Compose services
	@echo "${CYAN}Stopping Docker Compose stack...${RESET}"
	$(DOCKER_COMPOSE) down
	@echo "${GREEN}Stack stopped${RESET}"

## docker-logs: Show Docker Compose logs
docker-logs: ## Show Docker Compose logs
	$(DOCKER_COMPOSE) logs -f

## docker-clean: Stop stack and remove volumes
docker-clean: ## Stop stack and remove volumes
	@echo "${YELLOW}Stopping stack and removing volumes...${RESET}"
	$(DOCKER_COMPOSE) down -v
	@echo "${GREEN}Stack cleaned${RESET}"

## health-check: Check health of all services
health-check: ## Check health of Docker services
	@echo "${CYAN}Checking service health...${RESET}"
	@echo "MongoDB:"
	@docker exec ecommerce-mongo mongosh --eval "rs.status().ok" 2>/dev/null && echo "${GREEN}  OK${RESET}" || echo "${YELLOW}  Not ready${RESET}"
	@echo "Kafka:"
	@docker exec ecommerce-kafka kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1 && echo "${GREEN}  OK${RESET}" || echo "${YELLOW}  Not ready${RESET}"
	@echo "Debezium:"
	@curl -s http://localhost:8083/connectors >/dev/null && echo "${GREEN}  OK${RESET}" || echo "${YELLOW}  Not ready${RESET}"
	@echo "Kafka UI: http://localhost:8080"
