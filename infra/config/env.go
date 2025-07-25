package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	APIKey          string
	GeminiAPIKey    string
	GeminiModelName string
	Port            string
	Environment     string
	LogLevel        string

	// JWT configuration
	JWTSecret string

	// CORS configuration for integration with vistara-be
	AllowedOrigins string

	// Service integration
	VistaraBeURL string

	// Search grounding configuration
	EnableSearchGrounding bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		APIKey:                getEnv("API_SECRET_KEY", "vistara-ai-default-key"),
		GeminiAPIKey:          getEnv("GEMINI_API_KEY", ""),
		GeminiModelName:       getEnv("GEMINI_MODEL_NAME", "gemini-2.0-flash-exp"),
		Port:                  getEnv("PORT", "8080"),
		Environment:           getEnv("GO_ENV", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		JWTSecret:             getEnv("JWT_SECRET", "vistara-ai-jwt-secret-change-in-production"),
		AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "*"),
		VistaraBeURL:          getEnv("VISTARA_BE_URL", "http://localhost:8080"),
		EnableSearchGrounding: getEnvBool("ENABLE_SEARCH_GROUNDING", true),
	}

	// Validate critical settings
	if cfg.GeminiAPIKey == "" && cfg.Environment != "test" {
		log.Fatal("CRITICAL: GEMINI_API_KEY is missing")
	}

	if cfg.APIKey == "vistara-ai-default-key" {
		log.Println("WARNING: Using default API_SECRET_KEY")
	}

	if cfg.JWTSecret == "vistara-ai-jwt-secret-change-in-production" {
		log.Println("WARNING: Using default JWT_SECRET, change it in production")
	}

	log.Printf("Configuration loaded for environment: %s", cfg.Environment)
	log.Printf("Using Gemini model: %s", cfg.GeminiModelName)
	log.Printf("Search grounding enabled: %t", cfg.EnableSearchGrounding)

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
