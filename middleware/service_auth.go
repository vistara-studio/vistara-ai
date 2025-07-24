package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/infra/config"
)

// ServiceAuth middleware for service-to-service authentication
func ServiceAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for API key
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "API key is required",
			})
		}

		// Validate API key
		if apiKey != cfg.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid API key",
			})
		}

		// Check for service header (optional but recommended)
		serviceHeader := c.Get("X-Service")
		if serviceHeader != "" {
			// Validate service header if provided
			validServices := map[string]bool{
				"vistara-be": true,
				"vistara-ai": true,
			}

			if !validServices[serviceHeader] {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "Invalid service authentication",
				})
			}
		}

		// Store service info in context
		c.Locals("authenticated", true)
		c.Locals("service", serviceHeader)
		
		return c.Next()
	}
}
