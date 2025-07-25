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

## setup: Complete project setup (one command setup)
setup:
	@echo "🚀 Vistara AI - Complete Setup Starting..."
	@echo "========================================"
	@echo ""
	@echo "📋 Setup will do the following:"
	@echo "  1. ✅ Install Go dependencies"
	@echo "  2. 🔧 Setup environment configuration"
	@echo "  3. 🗄️  Initialize database schema"
	@echo "  4. 🏗️  Build application binary"
	@echo "  5. 🧪 Validate setup with health check"
	@echo ""
	@read -p "Continue with complete setup? (y/N): " confirm && [ "$$confirm" = "y" ] || { echo "❌ Setup cancelled."; exit 1; }
	@echo ""
	@echo "🔄 Step 1/5: Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"
	@echo ""
	@echo "🔄 Step 2/5: Setting up environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "📄 .env file created from .env.example"; \
		echo "⚠️  Please update .env with your actual API keys!"; \
	else \
		echo "📄 .env file already exists"; \
	fi
	@echo ""
	@echo "🔄 Step 3/5: Database setup instructions..."
	@echo "⚠️  Database setup requires PostgreSQL with PostGIS."
	@echo "   Run these commands if you have PostgreSQL installed:"
	@echo "   sudo -u postgres createdb vistara_ai"
	@echo "   make db-migrate"
	@echo "   (Skipping automatic database setup)"
	@echo ""
	@echo "🔄 Step 4/5: Building application..."
	@go build -o bin/vistara-ai cmd/api/main.go
	@echo "✅ Application built successfully: bin/vistara-ai"
	@echo ""
	@echo "🔄 Step 5/5: Validation..."
	@if [ -f bin/vistara-ai ]; then \
		echo "✅ Binary exists and is executable"; \
	else \
		echo "❌ Binary not found"; exit 1; \
	fi
	@echo ""
	@echo "🎉 SETUP COMPLETED SUCCESSFULLY!"
	@echo "================================"
	@echo ""
	@echo "📖 NEXT STEPS:"
	@echo "1. 🔑 Update .env file with your API keys:"
	@echo "   - GEMINI_API_KEY=your-actual-gemini-key"
	@echo "   - Update database credentials if needed"
	@echo ""
	@echo "2. 🗄️  Setup database (if not done):"
	@echo "   make db-setup-local    # Setup local PostgreSQL"
	@echo ""
	@echo "3. 🚀 Start the application:"
	@echo "   make run              # Start production server"
	@echo "   make dev              # Start with hot reload"
	@echo ""
	@echo "4. 🧪 Test the API:"
	@echo "   make test-api         # Test if endpoints work"
	@echo ""
	@echo "📡 Default server will run on: http://localhost:5000"
	@echo "📚 API Documentation: http://localhost:5000/docs"

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

## docker-setup: Setup complete environment with Docker (Database + App)
docker-setup:
	@echo "🐳 Setting up complete Docker environment..."
	@echo "📋 This will create:"
	@echo "  - PostgreSQL database with PostGIS"
	@echo "  - Redis cache"
	@echo "  - Vistara AI application"
	@echo ""
	@read -p "Continue with Docker setup? (y/N): " confirm && [ "$$confirm" = "y" ] || { echo "❌ Cancelled."; exit 1; }
	@echo "🔄 Building and starting services..."
	@docker compose up -d --build
	@echo "⏳ Waiting for services to be ready..."
	@sleep 15
	@echo "🔍 Checking service health..."
	@docker compose ps
	@echo ""
	@echo "✅ Docker environment ready!"
	@echo "📊 Services:"
	@echo "  - API: http://localhost:5000"
	@echo "  - Database: postgres://vistara:vistara123@localhost:5432/vistara_ai"
	@echo "  - Redis: redis://localhost:6379"

## docker-logs: View Docker container logs
docker-logs:
	@echo "📋 Docker container logs:"
	@docker compose logs --tail=50 -f

## docker-stop: Stop Docker containers
docker-stop:
	@echo "🛑 Stopping Docker containers..."
	@docker compose down

## docker-clean: Stop and remove all Docker resources
docker-clean:
	@echo "🧹 Cleaning Docker environment..."
	@docker compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Docker environment cleaned"

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

## docker-run: Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	@docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE)

## test-api: Test API endpoints
test-api:
	@echo "🧪 Testing Vistara AI endpoints..."
	@echo "🔍 Health check:"
	@curl -s http://localhost:$(PORT)/api/v1/health | jq . || echo "❌ Service not responding"
	@echo ""
	@echo "🗺️  Testing destination endpoints:"
	@curl -s "http://localhost:$(PORT)/api/v1/destinations?limit=3" | jq .data[0] || echo "❌ Destinations endpoint failed"
	@echo ""
	@echo "📍 Testing map view:"
	@curl -s "http://localhost:$(PORT)/api/v1/destinations/map?lat=-6.2088&long=106.8456&zoom=12" | jq .data || echo "❌ Map view failed"
	@echo ""
	@echo "🗂️ Testing itinerary endpoints (requires auth):"
	@echo "Note: Itinerary endpoints require JWT authentication"

## test-integrations: Test integrations between services
test-integrations:
	@echo "🔗 Testing service integrations..."
	@echo "📱 Testing destination to itinerary flow:"
	@echo "1. Get destinations -> 2. Create itinerary -> 3. Add destination to itinerary"
	@echo "Note: This requires authentication tokens"

## db-setup: Setup PostgreSQL with PostGIS
db-setup:
	@echo "🗄️ Setting up database..."
	@docker run --name vistara-postgres 
		-e POSTGRES_DB=vistara_ai 
		-e POSTGRES_USER=vistara 
		-e POSTGRES_PASSWORD=vistara123 
		-p 5432:5432 
		-d postgis/postgis:14-3.2
	@echo "⏳ Waiting for database to be ready..."
	@sleep 10
	@echo "✅ Database container started"

## db-setup-local: Setup local PostgreSQL database
db-setup-local:
	@echo "🗄️  Setting up local PostgreSQL database..."
	@echo "📋 This will:"
	@echo "  1. Create vistara_ai database"
	@echo "  2. Apply all migrations"
	@echo "  3. Verify database setup"
	@echo ""
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || { echo "❌ Cancelled."; exit 1; }
	@echo "🔄 Creating database..."
	@sudo -u postgres createdb vistara_ai 2>/dev/null || echo "📄 Database may already exist"
	@sudo -u postgres psql -c "CREATE USER vistara WITH PASSWORD 'vistara123';" 2>/dev/null || echo "📄 User may already exist"
	@sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE vistara_ai TO vistara;" 2>/dev/null || echo "📄 Privileges may already be granted"
	@echo "🔄 Applying migrations..."
	@make db-migrate
	@echo "✅ Database setup completed!"

## test-api: Test API endpoints to verify setup
test-api:
	@echo "🧪 Testing Vistara AI API..."
	@echo "🔄 Starting server in background for testing..."
	@if pgrep -f "vistara-ai" > /dev/null; then \
		echo "✅ Server already running"; \
	else \
		echo "🚀 Starting server..."; \
		./bin/vistara-ai > /dev/null 2>&1 & \
		echo $$! > .server.pid; \
		sleep 3; \
	fi
	@echo ""
	@echo "🧪 Testing endpoints..."
	@echo "1. Health Check:"
	@curl -s http://localhost:5000/health 2>/dev/null || echo "❌ Health check failed"
	@echo ""
	@echo "2. AI Service:"
	@curl -s -X POST http://localhost:5000/ai/plan \
		-H "Content-Type: application/json" \
		-d '{"destination":"Bali","start_date":"2024-12-01T00:00:00Z","end_date":"2024-12-05T00:00:00Z","budget":1000}' \
		2>/dev/null | head -200 || echo "❌ AI plan endpoint failed"
	@echo ""
	@echo "3. Destinations:"
	@curl -s "http://localhost:5000/destinations?limit=1" 2>/dev/null || echo "❌ Destinations endpoint failed"
	@echo ""
	@echo "✅ API testing completed!"
	@echo "📖 Full API documentation available at: http://localhost:5000/docs"

## stop-server: Stop running server
stop-server:
	@echo "🛑 Stopping server..."
	@if [ -f .server.pid ]; then \
		kill `cat .server.pid` 2>/dev/null || true; \
		rm .server.pid; \
	fi
	@pkill -f "vistara-ai" 2>/dev/null || true
	@echo "✅ Server stopped"

## check-env: Validate environment configuration
check-env:
	@echo "🔍 Checking environment configuration..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@echo "✅ .env file exists"
	@if grep -q "your-gemini-api-key-here" .env; then \
		echo "⚠️  Please update GEMINI_API_KEY in .env file"; \
	else \
		echo "✅ GEMINI_API_KEY appears to be configured"; \
	fi
	@if grep -q "your_azure_tts_api_key" .env; then \
		echo "⚠️  Consider updating TTS API keys for full functionality"; \
	fi
	@echo "✅ Environment check completed"

## full-setup: Complete setup including database
full-setup: setup db-setup-local check-env
	@echo ""
	@echo "🎉 FULL SETUP COMPLETED!"
	@echo "========================"
	@echo "✅ Dependencies installed"
	@echo "✅ Environment configured"  
	@echo "✅ Database initialized"
	@echo "✅ Application built"
	@echo ""
	@echo "🚀 Ready to start development!"
	@echo "   make dev     # Start with hot reload"
	@echo "   make run     # Start production server"

## db-migrate: Apply database schema
db-migrate:
	@echo "🔄 Applying database migrations..."
	@./database/migrate.sh
	@echo "✅ Database schema applied"

## db-migrate-single: Apply single migration file
db-migrate-single:
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Please specify migration file: make db-migrate-single FILE=001_init_postgis.sql"; \
		exit 1; \
	fi
	@echo "🔄 Applying migration: $(FILE)"
	@PGPASSWORD=vistara123 psql -h localhost -p 5432 -U vistara -d vistara_ai -f database/migrations/$(FILE)
	@echo "✅ Migration $(FILE) applied"

## db-seed: Load sample data
db-seed:
	@echo "🌱 Loading sample destination data..."
	@echo "INSERT INTO destinations (id, name, description, short_description, location, main_image_url, category, price_amount, price_currency, price_unit) VALUES" > /tmp/sample_data.sql
	@echo "('dest-1', 'Monas', 'Monumen Nasional Jakarta', 'Iconic monument in Jakarta', ST_SetSRID(ST_MakePoint(106.8270, -6.1754), 4326), 'https://example.com/monas.jpg', 'Monument', 15000, 'IDR', 'per_person')," >> /tmp/sample_data.sql
	@echo "('dest-2', 'Borobudur', 'Ancient Buddhist temple', 'UNESCO World Heritage Site', ST_SetSRID(ST_MakePoint(110.2038, -7.6079), 4326), 'https://example.com/borobudur.jpg', 'Heritage', 50000, 'IDR', 'per_person');" >> /tmp/sample_data.sql
	@PGPASSWORD=vistara123 psql -h localhost -p 5432 -U vistara -d vistara_ai -f /tmp/sample_data.sql
	@rm /tmp/sample_data.sql
	@echo "✅ Sample data loaded"

## db-test: Test database connection
db-test:
	@echo "🔌 Testing database connection..."
	@PGPASSWORD=vistara123 psql -h localhost -p 5432 -U vistara -d vistara_ai -c "SELECT 'Database connected successfully!' as status, version();"

## db-reset: Reset database
db-reset:
	@echo "🗑️ Resetting database..."
	@docker stop vistara-postgres || true
	@docker rm vistara-postgres || true
	@echo "✅ Database reset complete"

## db-status: Show database status
db-status:
	@echo "📊 Database Status:"
	@docker ps | grep vistara-postgres || echo "❌ Database container not running"
	@echo ""
	@echo "📋 Table Information:"
	@PGPASSWORD=vistara123 psql -h localhost -p 5432 -U vistara -d vistara_ai -c "\dt" || echo "❌ Cannot connect to database"## test-smart-plan: Test smart plan endpoint
test-smart-plan:
	@echo "🧪 Testing smart plan endpoint..."
	@curl -X POST http://localhost:$(PORT)/api/v1/smart-planner \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer test-jwt-token" \
		-d '{"destination":"Bali","start_date":"2025-01-15T00:00:00Z","end_date":"2025-01-20T00:00:00Z","budget":5000000,"travel_style":"romantic_couple","activity_preferences":["beach","culture"],"activity_intensity":"balanced"}' \
		| jq . || echo "API not responding or jq not installed"

## mod-tidy: Tidy Go modules
mod-tidy:
	@echo "📦 Tidying Go modules..."
	@go mod tidy
	@go mod download
