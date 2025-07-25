#!/bin/bash

# Vistara AI Service Reset Script
echo "🧹 Resetting Vistara AI Service..."

# Stop any running processes - more comprehensive
echo "🛑 Stopping all running processes..."
pkill -f "air" 2>/dev/null || true
pkill -f "go run cmd/api/main.go" 2>/dev/null || true
pkill -f "go run.*main.go" 2>/dev/null || true
pkill -f "bin/vistara-ai" 2>/dev/null || true
pkill -f "vistara-ai" 2>/dev/null || true

# Kill processes on specific ports
echo "🔌 Killing processes on ports 5000 and 8080..."
lsof -ti:5000 | xargs kill -9 2>/dev/null || true
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

# Wait for processes to terminate
sleep 2

# Clean build artifacts
echo "🗑️  Cleaning build artifacts..."
rm -rf bin/
rm -rf tmp/
rm -rf logs/
rm -rf coverage.out coverage.html
rm -f .air.toml.bak

# Stop and remove Docker containers/images
echo "🐳 Stopping and removing Docker containers..."
docker-compose down -v 2>/dev/null || true
docker stop vistara-ai-app vistara-ai-db vistara-ai-redis 2>/dev/null || true
docker rm vistara-ai-app vistara-ai-db vistara-ai-redis 2>/dev/null || true
docker rmi vistara-ai-app vistara-ai:latest 2>/dev/null || true
docker volume prune -f 2>/dev/null || true
docker network prune -f 2>/dev/null || true

# Clean Go module cache for this project - more thorough
echo "🧼 Cleaning Go module cache..."
go clean -modcache -cache -testcache -fuzzcache 2>/dev/null || true

# Clean development tools
echo "🔧 Cleaning development tools..."
go clean -i github.com/air-verse/air 2>/dev/null || true
go clean -i github.com/cosmtrek/air 2>/dev/null || true
go clean -i github.com/golangci/golangci-lint/cmd/golangci-lint 2>/dev/null || true

# Reset .env file
if [ -f .env ]; then
    echo "📋 Backing up current .env to .env.backup..."
    cp .env .env.backup
    rm .env
fi

echo "📋 Creating fresh .env from template..."
if [ -f .env.example ]; then
    cp .env.example .env
else
    # Create basic .env template
    cat > .env << EOF
# Vistara AI Service Environment Configuration

# Application Configuration
GO_ENV=development
PORT=5000
LOG_LEVEL=info

# Security Keys (for API authentication)
API_SECRET_KEY=your_api_secret_key_here

# JWT Configuration  
JWT_SECRET=your_jwt_secret_here

# AI Configuration
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_MODEL_NAME=gemini-2.5-flash

# CORS Configuration for integration
ALLOWED_ORIGINS=http://localhost:8080,http://localhost:3000

# Integration with Vistara BE
VISTARA_BE_URL=http://localhost:8080
EOF
fi

# Reinstall dependencies
echo "📦 Reinstalling dependencies..."
rm -f go.sum 2>/dev/null || true
go mod tidy
go mod download

echo ""
echo "✅ Reset completed successfully!"
echo ""
echo "Your previous .env has been backed up to .env.backup"
echo "Please update the new .env file with your configuration"
echo "Run './scripts/setup.sh' to complete the setup"
echo ""
