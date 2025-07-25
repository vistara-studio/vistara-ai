package config

import (
	"log"
	"os"
	"strconv"
	"time"

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

	// Performance optimization
	RequestTimeout     time.Duration
	MaxTokens          int32
	ConcurrentRequests int

	// Service-specific timeouts
	SmartPlannerTimeout time.Duration
	HistorianTimeout    time.Duration
	NusalingoTimeout    time.Duration
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
		GeminiModelName:       getEnv("GEMINI_MODEL_NAME", "gemini-2.5-flash"),
		Port:                  getEnv("PORT", "8080"),
		Environment:           getEnv("GO_ENV", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		JWTSecret:             getEnv("JWT_SECRET", "vistara-ai-jwt-secret-change-in-production"),
		AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "*"),
		VistaraBeURL:          getEnv("VISTARA_BE_URL", "http://localhost:8080"),
		EnableSearchGrounding: getEnvBool("ENABLE_SEARCH_GROUNDING", true),
		RequestTimeout:        time.Duration(getEnvInt("REQUEST_TIMEOUT_SECONDS", 60)) * time.Second,
		MaxTokens:             int32(getEnvInt("MAX_TOKENS", 4096)),
		ConcurrentRequests:    getEnvInt("CONCURRENT_REQUESTS", 3),
		SmartPlannerTimeout:   time.Duration(getEnvInt("SMART_PLANNER_TIMEOUT_SECONDS", 90)) * time.Second,
		HistorianTimeout:      time.Duration(getEnvInt("HISTORIAN_TIMEOUT_SECONDS", 45)) * time.Second,
		NusalingoTimeout:      time.Duration(getEnvInt("NUSALINGO_TIMEOUT_SECONDS", 30)) * time.Second,
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

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
