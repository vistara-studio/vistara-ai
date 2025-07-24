package middleware

import (
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/service"
)

// RequireBothAuth middleware that requires BOTH JWT authentication AND service authentication
func RequireBothAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check service authentication
		serviceHeader := c.Get("X-Service")
		if serviceHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Service authentication required (X-Service header missing)",
			})
		}

		// Validate service header (you can add more validation here)
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

		// Then check JWT authentication
		jwtMiddleware := JWTAuth(cfg)
		if err := jwtMiddleware(c); err != nil {
			// If JWT fails, return more specific error
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Both service and user authentication required",
				"error":   "JWT authentication failed",
			})
		}

		return c.Next()
	}
}

// RequireJWTOnly middleware that requires ONLY JWT authentication (for user endpoints)
func RequireJWTOnly(cfg *config.Config) fiber.Handler {
	return JWTAuth(cfg)
}

// RequireServiceOnly middleware that requires ONLY service authentication (for service-to-service)
func RequireServiceOnly(cfg *config.Config) fiber.Handler {
	return ServiceAuth(cfg)
}

// RequireEitherAuth middleware that requires EITHER JWT OR service authentication
func RequireEitherAuth(cfg *config.Config) fiber.Handler {
	authService := service.NewAuthService(cfg)
	
	return func(c *fiber.Ctx) error {
		// Check JWT authentication first
		authHeader := c.Get("Authorization")
		
		jwtValid := false
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != "" {
				// Try to validate as vistara-be token using the auth service
				userProfile, err := authService.ParseJWTToken(tokenString)
				if err == nil && userProfile != nil {
					// Vistara-be token is valid
					c.Locals("user_id", userProfile.ID)
					c.Locals("username", userProfile.Username)
					c.Locals("email", userProfile.Email)
					c.Locals("name", userProfile.Name)
					c.Locals("role", userProfile.Role)
					c.Locals("authenticated", true)
					c.Locals("auth_source", "vistara-be")
					jwtValid = true
				} else {
					// If vistara-be format fails, try vistara-ai format
					aiToken, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
						if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, jwt.ErrSignatureInvalid
						}
						return []byte(cfg.JWTSecret), nil
					})
					
					if err == nil && aiToken.Valid {
						if claims, ok := aiToken.Claims.(*JWTClaims); ok {
							// Store user info in context for vistara-ai token
							c.Locals("user_id", claims.UserID)
							c.Locals("username", claims.Username)
							c.Locals("email", claims.Email)
							c.Locals("role", claims.Role)
							c.Locals("authenticated", true)
							c.Locals("auth_source", "vistara-ai")
							jwtValid = true
						}
					}
				}
			}
		}
		
		if jwtValid {
			return c.Next()
		}
		
		// JWT failed, try service auth
		apiKey := c.Get("X-API-Key")
		
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required - provide either JWT token or valid API key",
			})
		}
		
		if apiKey != cfg.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid API key",
			})
		}
		
		// Service authentication successful
		c.Locals("authenticated", true)
		c.Locals("service", c.Get("X-Service"))
		
		return c.Next()
	}
}
