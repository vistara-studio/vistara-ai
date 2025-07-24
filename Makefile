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
	@if command -v air > /dev/null 2>&1; then \
		air; \
	elif [ -f $(go env GOPATH)/bin/air ]; then \
		$(go env GOPATH)/bin/air; \
	else \
		echo "Air not found. Install it with: make install-tools"; \
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
	$(GOCMD) install github.com/air-verse/air@latest
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
	@echo "🚀 Setup & Installation:"
	@echo "  setup              Complete setup (recommended for first time)"
	@echo "  quick-setup        Minimal setup for experienced developers"
	@echo "  dev-setup          Setup with all development tools"
	@echo "  reset-setup        Reset everything and start fresh"
	@echo "  verify-setup       Verify that setup was successful"
	@echo "  health-check       Check system health and configuration"
	@echo ""
	@echo "💻 Development:"
	@echo "  build              Build the application"
	@echo "  run                Run the application"
	@echo "  dev                Run with hot reload (requires air)"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Download dependencies"
	@echo ""
	@echo "🔧 Code Quality:"
	@echo "  fmt                Format code"
	@echo "  lint               Lint code"
	@echo "  docs               Generate API documentation"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-run         Run Docker container"
	@echo "  docker-stop        Stop Docker container"
	@echo "  docker-restart     Rebuild and restart container"
	@echo "  docker-logs        View Docker container logs"
	@echo "  docker-up          Start with Docker Compose (if available)"
	@echo "  docker-down        Stop Docker Compose services"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test-api           Test API health endpoint"
	@echo "  test-smart-plan    Test smart plan endpoint"
	@echo "  test-integration   Test integration with vistara-be"
	@echo ""
	@echo "🔗 Integration:"
	@echo "  dev-full           Start both vistara-ai and vistara-be"
	@echo "  stop-dev-full      Stop both services"
	@echo "  test-smart-plan-integration  Test with integration data"
	@echo ""
	@echo "🛠️ Tools:"
	@echo "  install-tools      Install development tools"
	@echo "  install-dev-tools  Install additional dev tools"
	@echo "  setup-hooks        Setup git pre-commit hooks"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  clean-all          Deep clean all artifacts"
	@echo "  docker-clean       Clean Docker resources"
	@echo "  stop-all           Stop all running processes"
	@echo ""
	@echo "📚 Help:"
	@echo "  help               Show this help message"
	@echo ""
	@echo "🔥 Quick Start:"
	@echo "  1. make setup      (first time setup)"
	@echo "  2. make run        (run the application)"
	@echo "  3. make test-api   (test if it's working)"

# Default target
all: deps build

## Test integration with vistara-be (requires both services running)
test-integration:
	@echo "Testing integration with vistara-be..."
	@echo "Note: Make sure vistara-be is running on http://localhost:8080"
	@echo "Checking vistara-be..."
	@curl -f http://localhost:8080/health 2>/dev/null || echo "⚠️  vistara-be is not running on port 8080"
	@echo "Checking vistara-ai..."
	@curl -f http://localhost:5000/api/v1/health 2>/dev/null && echo "✅ vistara-ai is running!" || echo "❌ vistara-ai is not running!"
	@echo ""
	@echo "Testing vistara-ai standalone (since vistara-be may not be running):"
	@curl -s http://localhost:5000/ | jq '.status' 2>/dev/null && echo "✅ Main endpoint works" || echo "❌ Main endpoint failed"

## Test smart plan with integration data
test-smart-plan-integration:
	@echo "Testing smart plan with integration..."
	@curl -X POST http://localhost:5000/api/v1/smart-planner \
		-H "Content-Type: application/json" \
		-H "X-Service: vistara-be" \
		-d '{"destination":"Bali","start_date":"2025-08-01T00:00:00Z","end_date":"2025-08-05T00:00:00Z","budget":5000000,"travel_style":"romantic_couple","activity_preferences":["beach","culture"],"activity_intensity":"balanced","user_id":"test-user-123"}'

## Test full integration (both services must be running)
test-full-integration:
	@echo "🧪 Testing full integration between vistara-ai and vistara-be..."
	@echo ""
	@echo "1️⃣ Testing vistara-be health..."
	@if curl -s http://localhost:8080/health | grep -q "Everything is good"; then \
		echo "✅ vistara-be healthy"; \
	else \
		echo "❌ vistara-be not responding"; \
	fi
	@echo ""
	@echo "2️⃣ Testing vistara-ai health..."
	@if curl -s http://localhost:5000/api/v1/health | jq -r '.status' 2>/dev/null | grep -q "healthy"; then \
		echo "✅ vistara-ai healthy"; \
	else \
		echo "❌ vistara-ai not responding"; \
	fi
	@echo ""
	@echo "3️⃣ Testing AI Smart Planner with service authentication..."
	@RESPONSE=$$(curl -s -X POST http://localhost:5000/api/v1/smart-planner \
		-H "Content-Type: application/json" \
		-H "X-Service: vistara-be" \
		-d '{"destination":"Yogyakarta","start_date":"2025-08-01T00:00:00Z","end_date":"2025-08-03T00:00:00Z","budget":2000000,"travel_style":"solo_traveler","activity_preferences":["culture","culinary","history"],"activity_intensity":"balanced","user_id":"integration-test-123"}'); \
	if echo "$$RESPONSE" | jq -r '.success' 2>/dev/null | grep -q "true"; then \
		echo "✅ AI integration working"; \
		echo "📄 Response preview:"; \
		echo "$$RESPONSE" | jq '.message' 2>/dev/null | head -c 200; \
		echo "..."; \
	else \
		echo "❌ AI integration failed"; \
		echo "🔍 Response:"; \
		echo "$$RESPONSE"; \
	fi
	@echo ""
	@echo "🎉 Full integration test completed!"

## Start both services for development
dev-full:
	@echo "Starting vistara-ai and vistara-be together..."
	@echo ""
	@echo "🔍 Checking if services are already running..."
	@if curl -s http://localhost:5000/api/v1/health >/dev/null 2>&1; then \
		echo "⚠️  vistara-ai already running on port 5000"; \
	else \
		echo "🚀 Starting vistara-ai on port 5000..."; \
		(cd . && make run) & \
	fi
	@if curl -s http://localhost:8080/health >/dev/null 2>&1; then \
		echo "⚠️  vistara-be already running on port 8080"; \
	else \
		echo "🚀 Starting vistara-be on port 8080..."; \
		(cd ../vistara-be && make run) & \
	fi
	@echo ""
	@echo "⏳ Waiting for services to start..."
	@sleep 5
	@echo ""
	@echo "🧪 Testing services..."
	@$(MAKE) test-integration
	@echo ""
	@echo "✅ Both services started successfully!"
	@echo "   Use 'make stop-dev-full' to stop them."
	@echo "   Use 'make test-full-integration' to test integration."

## Stop development services
stop-dev-full:
	@echo "Stopping development services..."
	@pkill -f "vistara-ai" || true
	@pkill -f "vistara-be" || true
	@echo "Services stopped."

## === SETUP COMMANDS ===

## Complete setup - runs everything needed to get started
setup:
	@echo "🚀 Starting Vistara AI complete setup..."
	@echo ""
	@echo "📋 Step 1: Checking prerequisites..."
	$(MAKE) check-prerequisites
	@echo ""
	@echo "📦 Step 2: Installing dependencies..."
	$(MAKE) deps
	@echo ""
	@echo "🔧 Step 3: Installing development tools..."
	$(MAKE) install-tools-silent
	@echo ""
	@echo "📁 Step 4: Creating necessary directories..."
	$(MAKE) create-dirs
	@echo ""
	@echo "⚙️ Step 5: Setting up environment..."
	$(MAKE) setup-env
	@echo ""
	@echo "🏗️ Step 6: Building application..."
	$(MAKE) build
	@echo ""
	@echo "🧪 Step 7: Running tests..."
	$(MAKE) test
	@echo ""
	@echo "✅ Setup completed successfully!"
	@echo ""
	@echo "🔥 Quick start commands:"
	@echo "  make run              - Run the application"
	@echo "  make dev              - Run with hot reload"
	@echo "  make docker-run       - Run in Docker"
	@echo "  make test-api         - Test the API"
	@echo ""
	@echo "📖 For more commands: make help"

## Reset setup - cleans everything and starts fresh
reset-setup:
	@echo "🔄 Resetting Vistara AI setup..."
	@echo ""
	@echo "⚠️  Warning: This will clean all build artifacts and dependencies!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo ""
	@echo "🧹 Step 1: Stopping all running processes..."
	$(MAKE) stop-all
	@echo ""
	@echo "🗑️ Step 2: Cleaning build artifacts..."
	$(MAKE) clean-all
	@echo ""
	@echo "🐳 Step 3: Cleaning Docker resources..."
	$(MAKE) docker-clean
	@echo ""
	@echo "📦 Step 4: Cleaning Go modules..."
	$(MAKE) clean-deps
	@echo ""
	@echo "✅ Reset completed! Run 'make setup' to start fresh."

## Quick setup - minimal setup for experienced developers
quick-setup:
	@echo "⚡ Quick setup for Vistara AI..."
	$(MAKE) deps
	$(MAKE) build
	@echo "✅ Quick setup completed! Run 'make run' to start."

## Development setup - setup with all dev tools
dev-setup:
	@echo "🛠️ Development setup for Vistara AI..."
	$(MAKE) setup
	@echo ""
	@echo "🔧 Installing additional dev tools..."
	$(MAKE) install-dev-tools
	@echo ""
	@echo "📝 Setting up pre-commit hooks..."
	$(MAKE) setup-hooks
	@echo ""
	@echo "✅ Development setup completed!"

## === SETUP HELPER COMMANDS ===

## Check prerequisites
check-prerequisites:
	@echo "Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed. Please install Go 1.19+"; exit 1; }
	@echo "✅ Go found: $$(go version)"
	@command -v docker >/dev/null 2>&1 || echo "⚠️  Docker not found (optional)"
	@command -v curl >/dev/null 2>&1 || echo "⚠️  curl not found (recommended for testing)"
	@command -v jq >/dev/null 2>&1 || echo "⚠️  jq not found (recommended for JSON parsing)"

## Create necessary directories
create-dirs:
	@echo "Creating directories..."
	@mkdir -p bin
	@mkdir -p logs
	@mkdir -p tmp
	@mkdir -p docs
	@echo "✅ Directories created"

## Setup environment file
setup-env:
	@echo "Setting up environment..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo "⚠️  Please edit .env file with your configuration"; \
		echo "   Required: GEMINI_API_KEY"; \
		echo "   Optional: API_SECRET_KEY (for production)"; \
	else \
		echo "✅ .env file already exists"; \
	fi

## Install development tools silently
install-tools-silent:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest >/dev/null 2>&1 || echo "⚠️  Failed to install air"
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest >/dev/null 2>&1 || echo "⚠️  Failed to install migrate"
	@echo "✅ Basic tools installed"

## Install additional dev tools
install-dev-tools:
	@echo "Installing additional development tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest || echo "⚠️  Failed to install swag"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || echo "⚠️  Failed to install golangci-lint"
	@go install github.com/vektra/mockery/v2@latest || echo "⚠️  Failed to install mockery"
	@echo "✅ Additional dev tools installed"

## Setup git hooks
setup-hooks:
	@echo "Setting up git hooks..."
	@if [ -d .git ]; then \
		echo '#!/bin/sh\nmake fmt\nmake lint' > .git/hooks/pre-commit; \
		chmod +x .git/hooks/pre-commit; \
		echo "✅ Pre-commit hooks installed"; \
	else \
		echo "⚠️  Not a git repository, skipping hooks"; \
	fi

## === CLEANUP COMMANDS ===

## Stop all running processes
stop-all:
	@echo "Stopping all processes..."
	@-pkill -f "vistara-ai" 2>/dev/null
	@-pkill -f "vistara-be" 2>/dev/null
	@-docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null
	@-docker-compose down 2>/dev/null
	@echo "✅ All processes stopped"

## Clean all artifacts
clean-all: clean
	@echo "Deep cleaning..."
	@rm -rf vendor/ || true
	@rm -rf coverage.* || true
	@rm -rf docs/swagger* || true
	@rm -rf logs/* || true
	@rm -rf tmp/* || true
	@echo "✅ Deep clean completed"

## Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@docker system prune -f 2>/dev/null || true
	@echo "✅ Docker resources cleaned"

## Clean dependencies
clean-deps:
	@echo "Cleaning Go dependencies..."
	@go clean -modcache || true
	@rm -rf go.sum || true
	@echo "✅ Dependencies cleaned"

## === VERIFICATION COMMANDS ===

## Verify setup
verify-setup:
	@echo "🔍 Verifying setup..."
	@echo ""
	@echo "📁 Checking directories..."
	@ls -la bin/ logs/ tmp/ docs/ 2>/dev/null || echo "⚠️  Some directories missing"
	@echo ""
	@echo "📄 Checking files..."
	@ls -la .env go.mod go.sum 2>/dev/null || echo "⚠️  Some files missing"
	@echo ""
	@echo "🏗️ Testing build..."
	@$(MAKE) build >/dev/null 2>&1 && echo "✅ Build works" || echo "❌ Build failed"
	@echo ""
	@echo "🧪 Testing basic functionality..."
	@go test ./... >/dev/null 2>&1 && echo "✅ Tests pass" || echo "❌ Tests failed"
	@echo ""
	@echo "🔧 Checking tools..."
	@if command -v air >/dev/null 2>&1; then \
		echo "✅ air installed"; \
	elif [ -f $(go env GOPATH)/bin/air ]; then \
		echo "✅ air installed (in GOPATH)"; \
	else \
		echo "⚠️  air not found"; \
	fi
	@if command -v migrate >/dev/null 2>&1; then \
		echo "✅ migrate installed"; \
	elif [ -f $(go env GOPATH)/bin/migrate ]; then \
		echo "✅ migrate installed (in GOPATH)"; \
	else \
		echo "⚠️  migrate not found"; \
	fi

## Health check
health-check:
	@echo "🏥 Health check..."
	@echo "Go version: $$(go version)"
	@echo "Module: $$(grep module go.mod)"
	@echo "Dependencies: $$(go list -m all | wc -l) modules"
	@echo "Binary: $$(ls -lh bin/ 2>/dev/null || echo 'Not built')"
	@echo "Environment: $$([ -f .env ] && echo 'Configured' || echo 'Not configured')"
