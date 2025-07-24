# Vistara AI

AI-powered travel planning service using Google Gemini AI for intelligent itinerary generation.

## Features

- 🤖 AI-powered travel planning with Google Gemini 2.5 Flash
- 📅 Date-based trip planning with validation
- 💰 Budget-optimized recommendations
- 🚀 Fast Go Fiber API
- 🐳 Docker containerization
- 🔒 API key authentication

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
ENVIRONMENT=development
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_MODEL=gemini-2.5-flash
API_KEY=your_api_key_here
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
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `5000` |
| `GEMINI_API_KEY` | Google Gemini API key | Required |
| `API_KEY` | API authentication key | Required |

## License

Copyright (c) 2025 Muhammad Rafly Ash Shiddiqi
