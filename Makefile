# Vistara AI Smart Planner Makefile

# Variables
APP_NAME=vistara-ai
BINARY_NAME=bin/$(APP_NAME)
MAIN_PATH=cmd/api/main.go
DOCKER_CONTAINER_NAME=vistara-ai

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Docker related variables
DOCKER_IMAGE=$(APP_NAME)
DOCKER_TAG=latest
DOCKER_PORT=5000

.PHONY: all build clean test deps run dev docker-build docker-run docker-stop help

## Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(GOBIN)
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BINARY_NAME)"

## Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(GOBIN)
	@echo "Clean completed"

## Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## Run the application
run:
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

## Run the application in development mode with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		$(MAKE) run; \
	fi

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

## Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	docker run -d --name $(DOCKER_CONTAINER_NAME) -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started: $(DOCKER_CONTAINER_NAME) on port $(DOCKER_PORT)"

## Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@echo "Container stopped: $(DOCKER_CONTAINER_NAME)"

## View Docker container logs
docker-logs:
	docker logs -f $(DOCKER_CONTAINER_NAME)

## Rebuild and restart Docker container
docker-restart: docker-build docker-run
	@echo "Docker container rebuilt and restarted"

## Test API endpoint
test-api:
	@echo "Testing API health endpoint..."
	@sleep 3
	@curl -s http://localhost:$(DOCKER_PORT)/ | jq . || echo "API not responding or jq not installed"

## Test smart plan endpoint with sample data
test-smart-plan:
	@echo "Testing smart plan endpoint..."
	@sleep 3
	@curl -X POST http://localhost:$(DOCKER_PORT)/api/v1/smart-planner \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer test-key" \
		-d '{"destination":"Bali","start_date":"2025-08-01","end_date":"2025-08-05","budget":5000000,"travel_style":"budget","activity_preferences":["sightseeing","adventure"]}' \
		| jq . || echo "API not responding or jq not installed"

## Run with Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	@if [ -f docker-compose.yml ]; then \
		docker-compose up -d; \
		echo "Services started. Check logs with: make docker-compose-logs"; \
	else \
		echo "docker-compose.yml not found. Using direct Docker commands..."; \
		$(MAKE) docker-run; \
	fi

## Stop Docker Compose services
docker-down:
	@echo "Stopping Docker Compose services..."
	@if [ -f docker-compose.yml ]; then \
		docker-compose down; \
		echo "Services stopped"; \
	else \
		echo "docker-compose.yml not found. Stopping direct Docker container..."; \
		$(MAKE) docker-stop; \
	fi

## View Docker Compose logs
docker-compose-logs:
	@if [ -f docker-compose.yml ]; then \
		docker-compose logs -f; \
	else \
		echo "docker-compose.yml not found. Showing Docker container logs..."; \
		$(MAKE) docker-logs; \
	fi

## Database migrations up
migrate-up:
	@echo "Running database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path db/migrations -database "$(DATABASE_URL)" up; \
	else \
		echo "migrate tool not found. Install it with:"; \
		echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

## Database migrations down
migrate-down:
	@echo "Rolling back database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path db/migrations -database "$(DATABASE_URL)" down; \
	else \
		echo "migrate tool not found. Install it with:"; \
		echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

## Create new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	if command -v migrate > /dev/null; then \
		migrate create -ext sql -dir db/migrations -seq $$name; \
	else \
		echo "migrate tool not found. Install it with:"; \
		echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

## Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/"; \
	fi

## Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOCMD) install github.com/cosmtrek/air@latest
	$(GOCMD) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully"

## Generate API documentation (if using swag)
docs:
	@echo "Generating API documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_PATH); \
	else \
		echo "swag not found. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

## Show this help message
help:
	@echo "Vistara AI - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  build          Build the application"
	@echo "  run            Run the application"
	@echo "  dev            Run with hot reload (requires air)"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  clean          Clean build artifacts"
	@echo "  deps           Download dependencies"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt            Format code"
	@echo "  lint           Lint code"
	@echo "  docs           Generate API documentation"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo "  docker-stop    Stop Docker container"
	@echo "  docker-restart Rebuild and restart container"
	@echo "  docker-logs    View Docker container logs"
	@echo "  docker-up      Start with Docker Compose (if available)"
	@echo "  docker-down    Stop Docker Compose services"
	@echo ""
	@echo "Testing:"
	@echo "  test-api       Test API health endpoint"
	@echo "  test-smart-plan Test smart plan endpoint"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up     Run database migrations"
	@echo "  migrate-down   Rollback database migrations"
	@echo "  migrate-create Create new migration file"
	@echo ""
	@echo "Tools:"
	@echo "  install-tools  Install development tools"
	@echo "  help           Show this help message"

# Default target
all: deps build
