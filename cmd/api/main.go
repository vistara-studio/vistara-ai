package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vistara-studio/vistara-ai/infra/config"
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
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: false, // Set to false when using wildcard origins
	}))

	// Health check route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Vistara AI Service is running",
			"version": "1.0.0",
		})
	})

	// Initialize AI services
	geminiService := service.NewGeminiService(cfg)
	smartPlannerService := service.NewSmartPlannerService(geminiService)
	
	// Initialize handlers
	aiHandler := handler.NewAIHandler(smartPlannerService, cfg)

	// Setup routes
	setupRoutes(app, aiHandler, cfg)

	// Start server
	log.Printf("Starting Vistara AI Service on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(app *fiber.App, aiHandler *handler.AIHandler, cfg *config.Config) {
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

	// AI Services routes (with API key authentication)
	// Smart Planner AI - directly under /api/v1
	api.Use("/smart-planner", middleware.APIKeyAuth(cfg))
	api.Post("/smart-planner", aiHandler.GenerateSmartPlan)
	
	// Future AI services can be added here
	// ai.Post("/recommendation-engine", aiHandler.GenerateRecommendations)
	// ai.Post("/travel-assistant", aiHandler.TravelAssistant)
	// ai.Post("/local-guide-matcher", aiHandler.LocalGuideMatcher)
}