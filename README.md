# Vistara AI - Smart Travel Planner

Vistara AI is an intelligent travel planning service that creates personalized itineraries using Google's Gemini AI. It integrates seamlessly with the Vistara BE platform to provide comprehensive travel planning solutions.

## � Quick Start

```bash
# Setup (first time)
make setup

# Run development server
make dev

# Test the API
make test-api
```

## 📋 Prerequisites

- Go 1.19+
- Google Gemini API Key
- Docker (optional)

## ⚙️ Configuration

Create `.env` file:

```env
# Required
GEMINI_API_KEY=your_gemini_api_key

# Optional (has defaults)
PORT=5000
API_SECRET_KEY=your_api_key
JWT_SECRET=your_jwt_secret
VISTARA_BE_URL=http://localhost:8080
GO_ENV=development
```

## 🔗 API Endpoints

### Health Check
```http
GET /api/v1/health
```

### Smart Travel Planning

**With JWT Authentication:**
```http
POST /api/v1/smart-planner
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "destination": "Bali",
  "start_date": "2025-01-15T00:00:00Z",
  "end_date": "2025-01-20T00:00:00Z",
  "budget": 5000000,
  "travel_style": "romantic_couple",
  "activity_preferences": ["beach", "culture"],
  "activity_intensity": "balanced"
}
```

**Service-to-Service (from Vistara BE):**
```http
POST /api/v1/smart-planner
X-Service: vistara-be
Content-Type: application/json
```

### Authentication

**Login:**
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "password"
}
```

**Get Profile:**
```http
GET /api/v1/auth/profile
Authorization: Bearer <jwt_token>
```

## 🛠️ Development

```bash
# Install dependencies
make install-deps

# Format code
make fmt

# Run linter
make lint

# Run tests with coverage
make test-coverage

# Build application
make build
```

## 🐳 Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

## 🔒 Authentication

Vistara AI uses a multi-tier authentication system:

1. **Public Endpoints**: Health checks
2. **JWT Only**: User-specific features  
3. **Service Only**: Inter-service communication
4. **Either Auth**: Smart planner (supports both JWT and service auth)

## 🧪 Testing

```bash
# Test health endpoint
make test-api

# Test smart planning
make test-smart-plan

# Run all tests
make test
```

## 📁 Project Structure

```
cmd/api/           # Application entry point
internal/handler/  # HTTP handlers
middleware/        # Authentication middleware
pkg/
  dto/            # Data transfer objects
  service/        # Business logic
  util/           # Helper utilities
  validator/      # Input validation
infra/
  config/         # Configuration
  logger/         # Logging utilities
  http/           # HTTP server setup
```

## 🔧 Make Commands

Run `make help` to see all available commands.

## 🤝 Integration with Vistara BE

Vistara AI integrates with Vistara BE to:
- Authenticate users via JWT tokens
- Fetch verified local businesses
- Retrieve tourist attractions  
- Send notifications when plans are generated

## 📝 License

This project is proprietary software of Vistara Studio.
