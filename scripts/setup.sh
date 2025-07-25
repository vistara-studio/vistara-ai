#!/bin/bash

# Vistara AI Service Setup Script
echo "🚀 Setting up Vistara AI Service..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'o' -f2)
echo "✅ Go version: $GO_VERSION"

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p bin
mkdir -p tmp  
mkdir -p logs
echo "✅ Directories created"

# Install development tools
echo "📦 Installing development tools..."

# Install Air for hot reload
if ! command -v air &> /dev/null; then
    echo "  � Installing Air for hot reload..."
    go install github.com/air-verse/air@latest
    if [ $? -eq 0 ]; then
        echo "  ✅ Air installed successfully"
    else
        echo "  ⚠️  Air installation failed, trying alternative..."
        go install github.com/cosmtrek/air@latest
        echo "  ✅ Air installed successfully (alternative)"
    fi
else
    echo "  ✅ Air is already installed"
fi

# Install golangci-lint for code linting  
if ! command -v golangci-lint &> /dev/null; then
    echo "  🔍 Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo "  ✅ golangci-lint installed successfully"
else
    echo "  ✅ golangci-lint is already installed"
fi

# Copy environment file if not exists
if [ ! -f .env ]; then
    echo "📋 Creating .env file from template..."
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
    echo "⚠️  Please update .env file with your actual configuration"
else
    echo "✅ .env file already exists"
fi

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy
go mod download

# Build the application
echo "🔨 Building application..."
mkdir -p bin
go build -ldflags="-s -w" -o bin/vistara-ai cmd/api/main.go

# Verify build
if [ -f bin/vistara-ai ]; then
    echo "✅ Application built successfully"
else
    echo "❌ Build failed"
    exit 1
fi

# Setup Docker environment
echo "🐳 Setting up Docker environment..."
echo "📋 This will create PostgreSQL + PostGIS database and Redis in Docker"
read -p "Do you want to setup Docker containers? (y/N): " setup_docker

if [[ $setup_docker =~ ^[Yy]$ ]]; then
    echo "🔄 Building and starting Docker containers..."
    docker-compose up -d --build
    
    echo "⏳ Waiting for database to be ready..."
    sleep 10
    
    # Check if database is ready
    echo "🔍 Checking database connection..."
    docker-compose exec -T db pg_isready -U vistara -d vistara_ai
    
    if [ $? -eq 0 ]; then
        echo "✅ Database is ready"
        echo "🗄️  Database setup completed with Docker!"
        echo "📊 Database URL: postgres://vistara:vistara123@localhost:5432/vistara_ai"
    else
        echo "⚠️  Database may still be starting up. Check with: docker-compose logs db"
    fi
else
    echo "⏭️  Skipping Docker setup"
fi

# Run tests to ensure everything works
echo "🧪 Running tests..."
go test ./... -v

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📝 Next steps:"
echo "1. 📝 Update .env file with your configuration:"
echo "   - Set GEMINI_API_KEY to your Google Gemini API key"
echo "   - Set API_SECRET_KEY for API authentication"  
echo "   - Set JWT_SECRET for JWT token signing"
echo ""
echo "2. 🚀 Start the application:"
echo "   🐳 With Docker (Recommended - includes database):"
echo "     - 'docker-compose up -d'     - Start all services"
echo "     - 'docker-compose logs app'  - View app logs"
echo "     - 'docker-compose down'      - Stop all services"
echo ""
echo "   🖥️  Local development:"
echo "     - 'make dev'   - Development with hot reload"
echo "     - 'make run'   - Production mode"
echo "     - 'make build' - Build binary"
echo ""
echo "3. 🧪 Test the application:"
echo "   - 'make test-api'              - Test health endpoint"
echo "   - 'curl http://localhost:5000/api/v1/health' - Direct health check"
echo ""
echo "4. 🔧 Development tools:"
echo "   - 'make lint' - Run code linter"
echo "   - 'make fmt'  - Format code"
echo ""
echo "📊 Service URLs:"
echo "   - API: http://localhost:5000"
echo "   - Database: postgres://vistara:vistara123@localhost:5432/vistara_ai"
echo "   - Redis: redis://localhost:6379"
echo ""
