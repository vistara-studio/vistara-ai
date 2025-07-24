package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// IntegrationService handles communication with vistara-be backend services
type IntegrationService struct {
	config     *config.Config
	httpClient *http.Client
}

// NewIntegrationService creates a new instance of IntegrationService
func NewIntegrationService(cfg *config.Config) *IntegrationService {
	return &IntegrationService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchLocalBusinesses retrieves local businesses from vistara-be API
func (s *IntegrationService) FetchLocalBusinesses(destination string, businessType string, userToken string) ([]dto.LocalBusiness, error) {
	url := fmt.Sprintf("%s/api/locals", s.config.VistaraBeURL)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if destination != "" {
		q.Add("location", destination)
	}
	if businessType != "" {
		q.Add("type", businessType)
	}
	req.URL.RawQuery = q.Encode()

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service", "vistara-ai")

	// Use JWT token if provided, otherwise use service API key as Bearer token
	if userToken != "" {
		req.Header.Set("Authorization", "Bearer "+userToken)
	} else {
		// Use API key as Bearer token for service-to-service authentication
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
		req.Header.Set("X-API-Key", s.config.APIKey)
	}	// Execute HTTP request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch local businesses: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data []dto.LocalBusiness `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// FetchTouristAttractions retrieves tourist attractions from vistara-be API
func (s *IntegrationService) FetchTouristAttractions(destination string, userToken string) ([]dto.TouristAttraction, error) {
	url := fmt.Sprintf("%s/api/tourist-attractions", s.config.VistaraBeURL)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if destination != "" {
		q.Add("location", destination)
	}
	req.URL.RawQuery = q.Encode()

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service", "vistara-ai")
	
	// Use JWT token if provided, otherwise use service API key as Bearer token
	if userToken != "" {
		req.Header.Set("Authorization", "Bearer "+userToken)
	} else {
		// Use API key as Bearer token for service-to-service authentication
		req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
		req.Header.Set("X-API-Key", s.config.APIKey)
	}

	// Execute HTTP request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tourist attractions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data []dto.TouristAttraction `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// NotifyPlanGenerated sends notification to vistara-be when a travel plan is generated
func (s *IntegrationService) NotifyPlanGenerated(userID string, planData interface{}) error {
	if userID == "" {
		return nil // Skip notification if no user ID provided
	}

	url := fmt.Sprintf("%s/api/ai/plan-generated", s.config.VistaraBeURL)

	payload := map[string]interface{}{
		"user_id":   userID,
		"plan_data": planData,
		"timestamp": time.Now(),
		"service":   "vistara-ai",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create notification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service", "vistara-ai")
	req.Header.Set("X-API-Key", s.config.APIKey) // Add API key for service authentication

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification failed with status code: %d", resp.StatusCode)
	}

	return nil
}
