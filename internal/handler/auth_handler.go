package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/middleware"
	"github.com/vistara-studio/vistara-ai/pkg/service"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	cfg         *config.Config
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		cfg:         cfg,
		authService: authService,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // in seconds
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

// Login handles user login through vistara-be and returns JWT token
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Validate login with vistara-be
	userProfile, err := h.authService.ValidateLoginWithVistaraBe(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid email or password",
			"error":   err.Error(),
		})
	}

	// Generate JWT token with user data from vistara-be
	token, err := middleware.GenerateJWT(
		h.cfg,
		userProfile.ID,
		userProfile.Username,
		userProfile.Email,
		userProfile.Role,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data": LoginResponse{
			Token:     token,
			ExpiresIn: 86400, // 24 hours in seconds
			UserID:    userProfile.ID,
			Username:  userProfile.Username,
			Email:     userProfile.Email,
			Role:      userProfile.Role,
		},
	})
}

// GetProfile returns the current user's profile from JWT
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Get user info from JWT middleware
	userID := c.Locals("user_id")
	username := c.Locals("username")
	email := c.Locals("email")
	role := c.Locals("role")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile retrieved successfully",
		"data": fiber.Map{
			"user_id":  userID,
			"username": username,
			"email":    email,
			"role":     role,
		},
	})
}

// LoginFallback handles fallback login when vistara-be is not available
func (h *AuthHandler) LoginFallback(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Check if vistara-be is available first
	if err := h.authService.CheckVistaraBeConnection(); err == nil {
		// If vistara-be is available, redirect to main login
		return h.Login(c)
	}

	// Fallback authentication for development/testing when vistara-be is down
	validUsers := map[string]map[string]string{
		"admin@vistara.com": {
			"password": "admin123",
			"username": "admin",
			"role":     "admin",
			"user_id":  "admin-001",
		},
		"user@vistara.com": {
			"password": "user123",
			"username": "user",
			"role":     "user",
			"user_id":  "user-001",
		},
		"test@vistara.com": {
			"password": "test123",
			"username": "testuser",
			"role":     "user",
			"user_id":  "test-001",
		},
	}

	user, exists := validUsers[req.Email]
	if !exists || user["password"] != req.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid email or password",
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(
		h.cfg,
		user["user_id"],
		user["username"],
		req.Email,
		user["role"],
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login successful (fallback mode - vistara-be unavailable)",
		"data": LoginResponse{
			Token:     token,
			ExpiresIn: 86400, // 24 hours in seconds
			UserID:    user["user_id"],
			Username:  user["username"],
			Email:     req.Email,
			Role:      user["role"],
		},
	})
}

// GenerateTestToken generates a test token for development/testing
func (h *AuthHandler) GenerateTestToken(c *fiber.Ctx) error {
	// Only allow in development environment
	if h.cfg.Environment == "production" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Test token generation not allowed in production",
		})
	}

	// Generate test token
	token, err := middleware.GenerateJWT(
		h.cfg,
		"test-user-123",
		"testuser",
		"test@vistara.com",
		"user",
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate test token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Test token generated successfully",
		"data": fiber.Map{
			"token": token,
			"usage": "Add this token to Authorization header as 'Bearer " + token + "'",
		},
	})
}

// CheckConnection checks the connection to vistara-be
func (h *AuthHandler) CheckConnection(c *fiber.Ctx) error {
	if err := h.authService.CheckVistaraBeConnection(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "vistara-be is not available",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "vistara-be connection is healthy",
	})
}
