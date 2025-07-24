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

// IntegrationService handles integration with vistara-be backend
type IntegrationService struct {
	config     *config.Config
	httpClient *http.Client
}

// NewIntegrationService creates a new integration service
func NewIntegrationService(cfg *config.Config) *IntegrationService {
	return &IntegrationService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchLocalBusinesses fetches local businesses from vistara-be
func (s *IntegrationService) FetchLocalBusinesses(destination string, businessType string, userToken string) ([]dto.LocalBusiness, error) {
	url := fmt.Sprintf("%s/api/locals", s.config.VistaraBeURL)
	
	// Create request
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

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service", "vistara-ai")
	
	// Use JWT token if provided, otherwise fall back to API key
	if userToken != "" {
		req.Header.Set("Authorization", "Bearer "+userToken)
	} else {
		req.Header.Set("X-API-Key", s.config.APIKey)
	}

	// Make request
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

// FetchTouristAttractions fetches tourist attractions from vistara-be
func (s *IntegrationService) FetchTouristAttractions(destination string, userToken string) ([]dto.TouristAttraction, error) {
	url := fmt.Sprintf("%s/api/tourist-attractions", s.config.VistaraBeURL)
	
	// Create request
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

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service", "vistara-ai")
	
	// Use JWT token if provided, otherwise fall back to API key
	if userToken != "" {
		req.Header.Set("Authorization", "Bearer "+userToken)
	} else {
		req.Header.Set("X-API-Key", s.config.APIKey) // Add API key for service authentication
	}

	// Make request
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

// NotifyPlanGenerated notifies vistara-be that a plan has been generated
func (s *IntegrationService) NotifyPlanGenerated(userID string, planData interface{}) error {
	if userID == "" {
		return nil // Skip notification if no user ID
	}

	url := fmt.Sprintf("%s/api/ai/plan-generated", s.config.VistaraBeURL)
	
	payload := map[string]interface{}{
		"user_id":    userID,
		"plan_data":  planData,
		"timestamp":  time.Now(),
		"service":    "vistara-ai",
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
