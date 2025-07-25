package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/infra/database"
	appLogger "github.com/vistara-studio/vistara-ai/infra/logger"
	"github.com/vistara-studio/vistara-ai/internal/handler"
	"github.com/vistara-studio/vistara-ai/middleware"
	"github.com/vistara-studio/vistara-ai/pkg/service"
	"github.com/vistara-studio/vistara-ai/pkg/validator"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
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
	authService := service.NewAuthService(cfg)
	
	// Initialize destination service (requires database connection)
	destinationService := service.NewDestinationService(db.GetDB(), nil, nil)
	
	// Initialize itinerary service
	itineraryService := service.NewItineraryService(db.GetDB(), geminiService, destinationService)
	
	// Initialize services that depend on database
	// recommendationService := service.NewRecommendationService(db.GetDB())
	// narrationService := service.NewNarrationService(cfg)

	// Initialize handlers
	aiHandler := handler.NewAIHandler(smartPlannerService, integrationService, appLog, cfg)
	authHandler := handler.NewAuthHandler(cfg, authService)
	
	// Initialize validator
	validatorInstance := validator.NewValidator()
	
	destinationHandler := handler.NewDestinationHandler(destinationService, validatorInstance)
	itineraryHandler := handler.NewItineraryHandler(itineraryService, validatorInstance)

	// Setup routes
	setupRoutes(app, aiHandler, authHandler, destinationHandler, itineraryHandler, cfg)

	// Start server
	log.Printf("Starting Vistara AI Service on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(app *fiber.App, aiHandler *handler.AIHandler, authHandler *handler.AuthHandler, destinationHandler *handler.DestinationHandler, itineraryHandler *handler.ItineraryHandler, cfg *config.Config) {
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

	// Public destination routes (no authentication required)
	destinations := api.Group("/destinations")
	destinations.Get("/", destinationHandler.GetDestinations)                 // Get destinations with optional user context
	destinations.Get("/map", destinationHandler.GetMapView)                   // Get map view of destinations
	destinations.Get("/categories", destinationHandler.GetCategories)         // Get destination categories
	destinations.Get("/:id", destinationHandler.GetDestinationByID)          // Get destination details
	destinations.Get("/:id/photos", destinationHandler.GetPhotos)            // Get destination photos
	// destinations.Get("/:id/reviews", destinationHandler.GetReviews)       // TODO: Implement GetReviews

	// Protected routes (require JWT authentication ONLY)
	protected := api.Group("/user")
	protected.Use(middleware.RequireJWTOnly(cfg))
	protected.Get("/profile", authHandler.GetProfile)
	protected.Post("/smart-planner", aiHandler.GenerateSmartPlan) // JWT-protected version
	
	// Protected destination routes (require authentication)
	protected.Post("/destinations/bookmark", destinationHandler.ToggleBookmark)     // Toggle bookmark
	protected.Get("/destinations/bookmarks", destinationHandler.GetUserBookmarks)   // Get user bookmarks
	// protected.Post("/destinations/:id/review", destinationHandler.AddReview)     // TODO: Implement AddReview

	// Protected itinerary routes (Smart Planner)
	protected.Post("/itineraries", itineraryHandler.CreateItinerary)                      // Create new itinerary
	protected.Get("/itineraries", itineraryHandler.GetUserItineraries)                    // Get user's itineraries
	protected.Get("/itineraries/history", itineraryHandler.GetItineraryHistory)           // Get detailed history
	protected.Get("/itineraries/:id", itineraryHandler.GetItinerary)                      // Get specific itinerary
	protected.Put("/itineraries/:id", itineraryHandler.UpdateItinerary)                   // Update itinerary
	protected.Delete("/itineraries/:id", itineraryHandler.DeleteItinerary)                // Delete itinerary
	protected.Post("/itineraries/:id/destinations", itineraryHandler.AddDestinationToItinerary) // Add destination to itinerary

	// Service-to-service routes (require service authentication ONLY)
	// These are for communication between vistara-be and vistara-ai
	service := api.Group("/service")
	service.Use(middleware.RequireServiceOnly(cfg))
	service.Post("/smart-planner", aiHandler.GenerateSmartPlan)

	// Secure routes (require BOTH JWT AND service authentication)
	// This is the most secure option for sensitive operations
	secure := api.Group("/secure")
	secure.Use(middleware.RequireBothAuth(cfg))
	secure.Post("/smart-planner", aiHandler.GenerateSmartPlan)

	// Legacy route - now requires EITHER JWT OR service auth (more secure than before)
	api.Post("/smart-planner", middleware.RequireEitherAuth(cfg), aiHandler.GenerateSmartPlan)

	// Future AI services can be added here
	// protected.Post("/recommendation-engine", aiHandler.GenerateRecommendations)
	// protected.Post("/travel-assistant", aiHandler.TravelAssistant)
	// service.Post("/local-guide-matcher", aiHandler.LocalGuideMatcher)
}
