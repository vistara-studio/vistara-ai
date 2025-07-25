# Vistara AI Service Makefile
.PHONY: help setup reset dev run build test clean docker-build docker-run

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := vistara-ai
BINARY_DIR := bin
MAIN_PATH := cmd/api/main.go
DOCKER_IMAGE := vistara-ai:latest
PORT := 5000

## help: Show this help message
help:
	@echo "Vistara AI Service - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "🔥 Quick Start:"
	@echo "  1. make setup         (first time setup with confirmation)"
	@echo "  2. make dev           (run with hot reload)"  
	@echo "  3. make test-api      (test if it's working)"
	@echo ""
	@echo "🧹 Reset Commands:"
	@echo "  make reset-setup      (complete reset with confirmation)"
	@echo "  make reset            (reset without confirmation)"
	@echo "  make stop             (stop running server)"

## setup: Initial project setup
setup:
	@echo "🚀 Setting up Vistara AI..."
	@echo "⚠️  This will install dependencies and configure the project."
	@read -p "Do you want to continue? (y/N): " confirm && [ "$$confirm" = "y" ] || { echo "❌ Setup cancelled."; exit 1; }
	@./scripts/setup.sh

## reset-setup: Reset project to clean state
reset-setup:
	@echo "🔄 Resetting Vistara AI..."
	@echo "⚠️  This will STOP all running processes and CLEAN all build artifacts!"
	@read -p "Are you sure you want to reset everything? (y/N): " confirm && [ "$$confirm" = "y" ] || { echo "❌ Reset cancelled."; exit 1; }
	@./scripts/reset-setup.sh

## reset: Reset project to clean state
reset:
	@echo "🔄 Resetting Vistara AI..."
	@./scripts/reset-setup.sh

## dev: Run with hot reload (development)
dev:
	@echo "🔥 Starting development server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	elif [ -f $$(go env GOPATH)/bin/air ]; then \
		$$(go env GOPATH)/bin/air; \
	else \
		echo "⚠️  Air not found. Installing Air..."; \
		go install github.com/air-verse/air@latest || go install github.com/cosmtrek/air@latest; \
		if command -v air >/dev/null 2>&1; then \
			air; \
		elif [ -f $$(go env GOPATH)/bin/air ]; then \
			$$(go env GOPATH)/bin/air; \
		else \
			echo "❌ Air installation failed. Falling back to normal run..."; \
			$(MAKE) run; \
		fi \
	fi

## stop: Stop running development server
stop:
	@echo "🛑 Stopping development server..."
	@pkill -f "air" 2>/dev/null || true
	@pkill -f "go run.*main.go" 2>/dev/null || true
	@pkill -f "vistara-ai" 2>/dev/null || true
	@lsof -ti:$(PORT) | xargs kill -9 2>/dev/null || true
	@echo "✅ Development server stopped"

## run: Run application (production)
run:
	@echo "🚀 Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

## build: Build the application
build:
	@echo "🔨 Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	@go build -ldflags="-s -w" -o $(BINARY_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✅ Build completed: $(BINARY_DIR)/$(APP_NAME)"

## test: Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)/
	@rm -rf tmp/
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache

## install-deps: Install development dependencies
install-deps:
	@echo "📦 Installing development dependencies..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Dependencies installed"

## lint: Run linter
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Run: make install-deps"; \
	fi

## fmt: Format code
fmt:
	@echo "📝 Formatting code..."
	@go fmt ./...

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

## docker-run: Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	@docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE)

## test-api: Test API health endpoint
test-api:
	@echo "🧪 Testing API health endpoint..."
	@sleep 2
	@curl -s http://localhost:$(PORT)/api/v1/health | jq . || echo "API not responding or jq not installed"

## test-smart-plan: Test smart plan endpoint
test-smart-plan:
	@echo "🧪 Testing smart plan endpoint..."
	@curl -X POST http://localhost:$(PORT)/api/v1/smart-planner \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer test-jwt-token" \
		-d '{"destination":"Bali","start_date":"2025-01-15T00:00:00Z","end_date":"2025-01-20T00:00:00Z","budget":5000000,"travel_style":"romantic_couple","activity_preferences":["beach","culture"],"activity_intensity":"balanced"}' \
		| jq . || echo "API not responding or jq not installed"

## test-historian: Test AI historian endpoint
test-historian:
	@echo "🧪 Testing AI historian endpoint..."
	@curl -X POST http://localhost:$(PORT)/api/v1/historical-story \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer test-jwt-token" \
		-d '{"location":"Borobudur Temple","additional_context":"ancient Buddhist temple"}' \
		| jq . || echo "API not responding or jq not installed"

## test-nusalingo: Test Nusalingo endpoint
test-nusalingo:
	@echo "🧪 Testing Nusalingo endpoint..."
	@curl -X POST http://localhost:$(PORT)/api/v1/nusalingo \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer test-jwt-token" \
		-d '{"input_text":"Halo, apa kabar?","target_language":"en","source_language":"id"}' \
		| jq . || echo "API not responding or jq not installed"

## test-all-endpoints: Test all API endpoints
test-all-endpoints: test-api test-smart-plan test-historian test-nusalingo
	@echo "✅ All API endpoint tests completed"

## mod-tidy: Tidy Go modules
mod-tidy:
	@echo "📦 Tidying Go modules..."
	@go mod tidy
	@go mod download
