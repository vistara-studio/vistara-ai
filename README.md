# Vistara AI

AI-powered travel planning service using Google Gemini AI for intelligent itinerary generation.

## Features

- 🤖 AI-powered travel planning with Google Gemini 2.5 Flash
- 📅 Date-based trip planning with validation
- 💰 Budget-optimized recommendations
- 🚀 Fast Go Fiber API
- 🐳 Docker containerization
- 🔒 API key authentication
- 🔗 **Seamless integration with Vistara BE**
- 🏢 **Local business recommendations from verified data**
- 🏛️ **Tourist attraction integration**

## Prerequisites

- Go 1.24.5+
- Docker
- Google Gemini API key

## Quick Start

### 1. Clone Repository
```bash
git clone https://github.com/vistara-studio/vistara-ai.git
cd vistara-ai
```

### 2. Environment Setup
Create `.env` file:
```bash
PORT=5000
GO_ENV=development
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_MODEL_NAME=gemini-2.0-flash-exp
API_SECRET_KEY=vistara-ai-integration-key

# For integration with vistara-be
ALLOWED_ORIGINS=http://localhost:8080,http://localhost:3000
VISTARA_BE_URL=http://localhost:8080
```

### 3. Run with Docker
```bash
make docker-build
make docker-run
```

Or run locally:
```bash
make deps
make run
```

## API Usage

### Health Check
```http
GET /
```

### Smart Travel Planning

**From Vistara BE (Service-to-Service):**
```http
POST /api/v1/smart-planner
X-Service: vistara-be
Content-Type: application/json

{
  "destination": "Bali",
  "start_date": "2025-08-01T00:00:00Z",
  "end_date": "2025-08-05T00:00:00Z",
  "budget": 5000000,
  "travel_style": "romantic_couple",
  "activity_preferences": ["beach", "culture", "culinary"],
  "activity_intensity": "balanced",
  "user_id": "user-123"
}
```

**External API (requires API key):**
```http
POST /api/v1/smart-planner
Authorization: Bearer your_api_key
Content-Type: application/json

{
  "destination": "Bali",
  "start_date": "2025-08-01",
  "end_date": "2025-08-05",
  "budget": 5000000,
  "travel_style": "budget",
  "activity_preferences": ["sightseeing", "adventure"]
}
```

## Development Commands

```bash
make build          # Build application
make run            # Run locally
make dev            # Run with hot reload
make test           # Run tests
make docker-build   # Build Docker image
make docker-run     # Run Docker container
make test-api       # Test API endpoints

# Integration with vistara-be
make test-integration        # Test integration with vistara-be
make test-smart-plan-integration  # Test smart plan with integration
make dev-full               # Start both vistara-ai and vistara-be
```

## Integration with Vistara BE

This service is designed to work seamlessly with the main Vistara backend. Key integration features:

- **Service-to-Service Authentication**: Vistara BE can call this service using the `X-Service` header
- **Local Business Integration**: Fetches verified local businesses from Vistara BE
- **Tourist Attraction Data**: Integrates real attraction data from the main database
- **User Context**: Maintains user context across services
- **Plan Notifications**: Notifies Vistara BE when plans are generated

For detailed integration instructions, see [INTEGRATION_GUIDE.md](./INTEGRATION_GUIDE.md).

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `5000` |
| `GEMINI_API_KEY` | Google Gemini API key | Required |
| `API_KEY` | API authentication key | Required |

## License

Copyright (c) 2025 Muhammad Rafly Ash Shiddiqi
