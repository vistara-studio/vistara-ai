package config

import (
	"log"
	"os"
	"strconv"

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
	
	// Database configuration
	DatabaseURL      string
	DBHost          string
	DBPort          string
	DBName          string
	DBUser          string
	DBPassword      string
	DBMaxConnections int
	DBSSLMode       string
	
	// Supabase Storage configuration
	SupabaseURL    string
	SupabaseKey    string
	SupabaseBucket string
	
	// Storage configuration
	StorageType string
	StoragePath string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		APIKey:          getEnv("API_SECRET_KEY", "vistara-ai-default-key"),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		GeminiModelName: getEnv("GEMINI_MODEL_NAME", "gemini-2.0-flash-exp"),
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("GO_ENV", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		JWTSecret:       getEnv("JWT_SECRET", "vistara-ai-jwt-secret-change-in-production"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "*"),
		VistaraBeURL:    getEnv("VISTARA_BE_URL", "http://localhost:8080"),
		
		// Database configuration
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBName:          getEnv("DB_NAME", "vistara_ai"),
		DBUser:          getEnv("DB_USER", "vistara"),
		DBPassword:      getEnv("DB_PASSWORD", "vistara123"),
		DBMaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS", 25),
		DBSSLMode:       getEnv("DB_SSL_MODE", "disable"),
		
		// Supabase Storage configuration
		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		SupabaseKey:    getEnv("SUPABASE_KEY", ""),
		SupabaseBucket: getEnv("SUPABASE_BUCKET", "vistara-bucket"),
		
		// Storage configuration
		StorageType: getEnv("STORAGE_TYPE", "local"),
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
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

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	
	return value
}
