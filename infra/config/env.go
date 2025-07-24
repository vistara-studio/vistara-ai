package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	APIKey           string
	GeminiAPIKey     string
	GeminiModelName  string
	GeminiAPIVersion string
	Port             string
	Environment      string
	LogLevel         string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		APIKey:           getEnv("API_SECRET_KEY", "vistara-ai-default-key"),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		GeminiModelName:  getEnv("GEMINI_MODEL_NAME", "gemini-2.0-flash-exp"),
		GeminiAPIVersion: getEnv("GEMINI_API_VERSION", ""),
		Port:             getEnv("PORT", "8080"),
		Environment:      getEnv("GO_ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}

	// Validate critical settings
	if cfg.GeminiAPIKey == "" && cfg.Environment != "test" {
		log.Fatal("CRITICAL: GEMINI_API_KEY is missing")
	}

	if cfg.APIKey == "vistara-ai-default-key" {
		log.Println("WARNING: Using default API_SECRET_KEY")
	}

	log.Printf("Configuration loaded for environment: %s", cfg.Environment)
	log.Printf("Using Gemini model: %s", cfg.GeminiModelName)

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
