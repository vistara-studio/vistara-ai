package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vistara-studio/vistara-ai/infra/config"
)

// JWTClaims represents the JWT claims structure for vistara-ai generated tokens
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// VistaBeClaims represents the JWT claims structure from vistara-be
type VistaBeClaims struct {
	UserID           string `json:"user_id"`
	IsPremium        bool   `json:"is_premium"`
	PremiumExpiredAt string `json:"premium_expired_at"`
	jwt.RegisteredClaims
}

// JWTAuth creates middleware for mandatory JWT authentication
func JWTAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header is required",
			})
		}

		// Check if it's Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization format. Use 'Bearer <token>'",
			})
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token is required",
			})
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token: " + err.Error(),
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token claims",
			})
		}

		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token has expired",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("jwt_claims", claims)

		return c.Next()
	}
}

// OptionalJWTAuth creates middleware for optional JWT authentication
// If JWT is provided, it validates it; if not, it continues without authentication
func OptionalJWTAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")

		// If no auth header, continue without authentication
		if authHeader == "" {
			return c.Next()
		}

		// Check if it's Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next() // Continue without authentication for non-Bearer tokens
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Next()
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		// If token is valid, store user info
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*JWTClaims); ok {
				// Check expiration
				if claims.ExpiresAt == nil || claims.ExpiresAt.Time.After(time.Now()) {
					c.Locals("user_id", claims.UserID)
					c.Locals("username", claims.Username)
					c.Locals("email", claims.Email)
					c.Locals("role", claims.Role)
					c.Locals("jwt_claims", claims)
					c.Locals("authenticated", true)
				}
			}
		}

		return c.Next()
	}
}

// GenerateJWT generates a new JWT token (utility function for testing)
func GenerateJWT(cfg *config.Config, userID, username, email, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vistara-ai",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
