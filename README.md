# Vistara AI - Uniting Journey and Indonesia

## 🌺 Features

- **Cultural AI Assistant**: Intelligent guidance that celebrates Indonesian heritage and traditions
- **Journey-Culture Integration**: Seamlessly connects travel experiences with cultural learning
- **Multi-modal Platform**: Supports education, ticketing, navigation, and translation services
- **Heritage Preservation**: AI-powered content that promotes and preserves Indonesian culture
- **Authentic Experiences**: Connects visitors with genuine local traditions and communitiesre

Vistara AI is an intelligent service that bridges journeys and Indonesian culture using Google's Gemini AI. It's part of the **Vistara** platform - a comprehensive digital ecosystem that integrates education, ticketing, navigation, and local language translation to preserve, promote, and celebrate Indonesian heritage.

The AI service seamlessly integrates with the Vistara platform to provide culturally-rich experiences that connect people with Indonesia's diverse traditions, heritage sites, and authentic local culture.

## 🚀 Quick Start

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
PORT=8080
API_SECRET_KEY=vistara-ai-default-key
JWT_SECRET=vistara-ai-jwt-secret-change-in-production
VISTARA_BE_URL=http://localhost:8080
GO_ENV=development
GEMINI_MODEL_NAME=gemini-2.5-flash
LOG_LEVEL=info
ALLOWED_ORIGINS=*
```

## � Features

- **Smart Itinerary Planning**: AI-powered travel plans using Google Gemini
- **Cultural Integration**: Focuses on Indonesian heritage and cultural experiences
- **Multi-tier Authentication**: Secure user and service-to-service communication
- **Seamless Integration**: Works with Vistara BE for comprehensive travel solutions
- **Flexible API**: Supports various authentication methods for different use cases

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

# Stop development server
make stop
```

## 🐳 Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

## 🔒 Authentication

Vistara AI uses a secure multi-tier authentication system to ensure proper access control:

- **Public Access**: Health checks and basic information
- **User Authentication**: JWT-based authentication for traveler features
- **Service Authentication**: API key authentication for inter-service communication
- **Flexible Security**: Multiple authentication options to support different integration needs

## 🏛️ About Vistara Platform

Vistara is a comprehensive digital ecosystem designed to bridge the gap between modern travel and Indonesia's rich cultural heritage:

- **Education**: Deep cultural learning and interactive heritage storytelling
- **Ticketing**: Seamless access to cultural sites, museums, and traditional experiences
- **Navigation**: Culturally-aware routing that highlights heritage along the journey
- **Translation**: Local language support for authentic cultural immersion

The AI service enhances this ecosystem by providing intelligent assistance that doesn't just guide journeys, but enriches them with cultural meaning and authentic Indonesian experiences.

## 🧪 Testing

```bash
# Test health endpoint
make test-api

# Test AI functionality
make test-smart-plan

# Run all tests
make test

# Reset project to clean state
make reset-setup
```

## 🤝 Integration with Vistara Platform

Vistara AI serves as the intelligent cultural bridge within the broader Vistara ecosystem:

- **Cultural Heritage Focus**: Prioritizes authentic Indonesian cultural sites, traditions, and heritage experiences
- **Community Connection**: Links visitors with local communities, artisans, and cultural practitioners
- **Educational Journey**: Transforms every interaction into a learning opportunity about Indonesian culture
- **Language & Tradition**: Supports local languages and traditional practices for authentic cultural exchange
- **Heritage Preservation**: Contributes to documenting and preserving Indonesia's diverse cultural legacy

This integration ensures that every experience through Vistara not only creates meaningful moments but actively participates in celebrating and preserving Indonesia's rich cultural tapestry - truly *"Menyatukan Perjalanan dan Budaya Indonesia"*.

## 📝 License

This project is proprietary software of Vistara Studio.
