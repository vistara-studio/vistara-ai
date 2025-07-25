package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vistara-studio/vistara-ai/infra/config"
	appLogger "github.com/vistara-studio/vistara-ai/infra/logger"
	"github.com/vistara-studio/vistara-ai/internal/handler"
	"github.com/vistara-studio/vistara-ai/middleware"
	"github.com/vistara-studio/vistara-ai/pkg/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: true,
	}))

	// Health check route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Vistara AI Service is running",
			"version": "1.0.0",
		})
	})

	// Initialize logger
	appLog := appLogger.New("info")

	// Initialize AI services
	geminiService := service.NewGeminiService(cfg)
	integrationService := service.NewIntegrationService(cfg)
	smartPlannerService := service.NewSmartPlannerService(geminiService, integrationService)
	aiHistorianService := service.NewAIHistorianService(geminiService)
	nusalingoService := service.NewNusalingoService(geminiService)
	authService := service.NewAuthService(cfg)

	// Initialize handlers
	smartPlannerHandler := handler.NewSmartPlannerHandler(smartPlannerService, integrationService, appLog, cfg)
	aiHandler := handler.NewAIHandler(aiHistorianService, nusalingoService, appLog, cfg)
	authHandler := handler.NewAuthHandler(cfg, authService)

	// Setup routes
	setupRoutes(app, smartPlannerHandler, aiHandler, authHandler, cfg)

	// Start server
	log.Printf("Starting Vistara AI Service on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(app *fiber.App, smartPlannerHandler *handler.SmartPlannerHandler, aiHandler *handler.AIHandler, authHandler *handler.AuthHandler, cfg *config.Config) {
	// API group
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "Vistara AI Service",
			"version": "1.0.0",
		})
	})

	// Auth routes (no authentication required)
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)                     // Login via vistara-be
	auth.Post("/login-fallback", authHandler.LoginFallback)    // Fallback when vistara-be is down
	auth.Get("/test-token", authHandler.GenerateTestToken)     // For development/testing
	auth.Get("/check-connection", authHandler.CheckConnection) // Check vistara-be connection

	// Protected routes (require JWT authentication ONLY)
	protected := api.Group("/user")
	protected.Use(middleware.RequireJWTOnly(cfg))
	protected.Get("/profile", authHandler.GetProfile)
	protected.Post("/smart-planner", smartPlannerHandler.GenerateSmartPlan) // JWT-protected version
	protected.Post("/historical-story", aiHandler.GenerateHistoricalStory)  // AI Historian service
	protected.Post("/nusalingo", aiHandler.TranslateText)                   // Nusalingo translation service

	// Service-to-service routes (require service authentication ONLY)
	// These are for communication between vistara-be and vistara-ai
	service := api.Group("/service")
	service.Use(middleware.RequireServiceOnly(cfg))
	service.Post("/smart-planner", smartPlannerHandler.GenerateSmartPlan)
	service.Post("/historical-story", aiHandler.GenerateHistoricalStory)
	service.Post("/nusalingo", aiHandler.TranslateText)

	// Secure routes (require BOTH JWT AND service authentication)
	// This is the most secure option for sensitive operations
	secure := api.Group("/secure")
	secure.Use(middleware.RequireBothAuth(cfg))
	secure.Post("/smart-planner", smartPlannerHandler.GenerateSmartPlan)
	secure.Post("/historical-story", aiHandler.GenerateHistoricalStory)
	secure.Post("/nusalingo", aiHandler.TranslateText)

	// Legacy route - now requires EITHER JWT OR service auth (more secure than before)
	api.Post("/smart-planner", middleware.RequireEitherAuth(cfg), smartPlannerHandler.GenerateSmartPlan)

	// Future AI services can be added here
	// protected.Post("/recommendation-engine", aiHandler.GenerateRecommendations)
	// protected.Post("/travel-assistant", aiHandler.TravelAssistant)
	// service.Post("/local-guide-matcher", aiHandler.LocalGuideMatcher)
}
