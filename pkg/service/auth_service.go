package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vistara-studio/vistara-ai/infra/config"
)

// AuthService handles authentication with vistara-be
type AuthService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LoginRequest represents login request to vistara-be
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents response from vistara-be login
type LoginResponse struct {
	Message string `json:"message"`
	Payload struct {
		Token string `json:"token"`
	} `json:"payload"`
}

// VistaraBeLoginResponse is the alternative response format
type VistaraBeLoginResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    struct {
		User struct {
			ID       string `json:"id"`
			Email    string `json:"email"`
			Username string `json:"username"`
			Name     string `json:"name"`
			Role     string `json:"role"`
		} `json:"user"`
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	} `json:"data"`
}

// UserProfile represents user profile from vistara-be
type UserProfile struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// ValidateLoginWithVistaraBe validates user login through vistara-be
func (s *AuthService) ValidateLoginWithVistaraBe(email, password string) (*UserProfile, error) {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Make request to vistara-be login endpoint
	req, err := http.NewRequest("POST", s.cfg.VistaraBeURL+"/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service", "vistara-ai") // Service identification

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make login request to vistara-be: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Try to decode the actual vistara-be response format
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	// Check if login was successful
	if loginResp.Message == "login successful" {
		// Login successful, continue processing
	} else {
		return nil, fmt.Errorf("login failed: %s", loginResp.Message)
	}

	// Check if we have a token
	if loginResp.Payload.Token == "" {
		return nil, fmt.Errorf("no token received from vistara-be")
	}

	// Extract user info from JWT token since vistara-be doesn't return user details
	userProfile, err := s.parseJWTToken(loginResp.Payload.Token)
	if err != nil {
		// If parsing fails, create basic profile with email info
		userProfile = &UserProfile{
			ID:       "vistara-be-user",
			Email:    email,
			Username: strings.Split(email, "@")[0], // Use email prefix as username
			Name:     email,
			Role:     "user",
		}
	} else {
		// Fill in missing info from login request
		userProfile.Email = email
		userProfile.Username = strings.Split(email, "@")[0]
		if userProfile.Name == "" {
			userProfile.Name = email
		}
	}

	return userProfile, nil
}

// ValidateUserWithVistaraBe validates user by ID through vistara-be
func (s *AuthService) ValidateUserWithVistaraBe(userID string) (*UserProfile, error) {
	// Make request to vistara-be user profile endpoint
	req, err := http.NewRequest("GET", s.cfg.VistaraBeURL+"/api/users/"+userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user validation request: %w", err)
	}

	req.Header.Set("X-Service", "vistara-ai") // Service identification

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user with vistara-be: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user validation failed with status: %d", resp.StatusCode)
	}

	var userResp struct {
		Success bool `json:"success"`
		Data    struct {
			User UserProfile `json:"user"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	if !userResp.Success {
		return nil, fmt.Errorf("user validation failed")
	}

	return &userResp.Data.User, nil
}

// CheckVistaraBeConnection checks if vistara-be is accessible
func (s *AuthService) CheckVistaraBeConnection() error {
	req, err := http.NewRequest("GET", s.cfg.VistaraBeURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vistara-be is not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vistara-be health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// parseJWTToken extracts user info from JWT token
func (s *AuthService) parseJWTToken(tokenString string) (*UserProfile, error) {
	// This is a simple JWT parser - in production you'd want to validate the signature
	// For now, we'll just extract the payload to get user_id
	
	// Split token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode payload (base64)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse JSON payload
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	// Extract user information
	userID, _ := claims["user_id"].(string)
	if userID == "" {
		return nil, fmt.Errorf("user_id not found in token")
	}

	// Create user profile from available claims
	return &UserProfile{
		ID:       userID,
		Email:    "", // Will be filled from login request
		Username: "", // Will be filled from login request
		Name:     "", // Will be filled from login request
		Role:     "user", // Default role
	}, nil
}

// ParseJWTToken is a public wrapper for parseJWTToken
func (s *AuthService) ParseJWTToken(tokenString string) (*UserProfile, error) {
	return s.parseJWTToken(tokenString)
}
