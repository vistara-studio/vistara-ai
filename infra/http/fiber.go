package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/infra/logger"
)

// Server represents the HTTP server
type Server struct {
	app    *fiber.App
	config *config.Config
	logger *logger.Logger
}

// New creates a new HTTP server
func New(cfg *config.Config, logger *logger.Logger) *Server {
	app := fiber.New()

	return &Server{
		app:    app,
		config: cfg,
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.app.Listen(":" + s.config.Port)
}

// Stop stops the HTTP server gracefully
func (s *Server) Stop() error {
	return s.app.Shutdown()
}

// App returns the fiber app instance
func (s *Server) App() *fiber.App {
	return s.app
}
