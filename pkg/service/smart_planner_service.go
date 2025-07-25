package service

import (
	"log"
	"sync"

	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/util"
)

// SmartPlannerService handles smart trip planning logic
type SmartPlannerService struct {
	geminiService      *GeminiService
	integrationService *IntegrationService
	config             *config.Config
}

// NewSmartPlannerService creates a new smart planner service instance
func NewSmartPlannerService(geminiService *GeminiService, integrationService *IntegrationService, cfg *config.Config) *SmartPlannerService {
	return &SmartPlannerService{
		geminiService:      geminiService,
		integrationService: integrationService,
		config:             cfg,
	}
}

// CreatePlan creates a personalized travel plan using AI with integration data
func (s *SmartPlannerService) CreatePlan(userInput *dto.SmartPlanRequest, duration int, userToken string) (string, error) {
	return s.CreatePlanWithGrounding(userInput, duration, userToken, false)
}

// CreatePlanWithGrounding creates a personalized travel plan using AI with integration data and optional search grounding
func (s *SmartPlannerService) CreatePlanWithGrounding(userInput *dto.SmartPlanRequest, duration int, userToken string, useGrounding bool) (string, error) {
	// Fetch additional data from vistara-be concurrently if integration is available
	var localBusinesses []dto.LocalBusiness
	var attractions []dto.TouristAttraction

	if s.integrationService != nil {
		var wg sync.WaitGroup
		wg.Add(2)

		// Fetch local businesses concurrently
		go func() {
			defer wg.Done()
			if businesses, err := s.integrationService.FetchLocalBusinesses(userInput.Destination, "", userToken); err != nil {
				log.Printf("Warning: Failed to fetch local businesses: %v", err)
			} else {
				localBusinesses = businesses
			}
		}()

		// Fetch tourist attractions concurrently
		go func() {
			defer wg.Done()
			if attr, err := s.integrationService.FetchTouristAttractions(userInput.Destination, userToken); err != nil {
				log.Printf("Warning: Failed to fetch tourist attractions: %v", err)
			} else {
				attractions = attr
			}
		}()

		// Wait for both requests to complete
		wg.Wait()
	}

	// Format the optimized prompt using validated user input, calculated duration, and integration data
	prompt := util.FormatOptimizedGeminiPrompt(userInput, duration, localBusinesses, attractions)

	var logMessage string
	if useGrounding {
		logMessage = "with search grounding for real-time destination information"
	} else {
		logMessage = "with integration data"
	}
	log.Printf("Formatted optimized prompt for AI Smart Planner generation %s", logMessage)

	// Call the Gemini service to generate the itinerary with optional grounding and custom timeout
	var itinerary string
	var err error
	if useGrounding {
		itinerary, err = s.geminiService.GenerateTextWithTimeoutAndGrounding(prompt, s.config.SmartPlannerTimeout, true)
	} else {
		itinerary, err = s.geminiService.GenerateTextWithTimeout(prompt, s.config.SmartPlannerTimeout)
	}

	if err != nil {
		log.Printf("AI Service error during Smart Planner creation: %v", err)
		return "", err
	}

	// Notify vistara-be about the generated plan if user ID is provided (asynchronously)
	if userInput.UserID != nil && *userInput.UserID != "" && s.integrationService != nil {
		go func() {
			if err := s.integrationService.NotifyPlanGenerated(*userInput.UserID, itinerary); err != nil {
				log.Printf("Warning: Failed to notify vistara-be about plan generation: %v", err)
			}
		}()
	}

	log.Println("AI Smart Planner itinerary generated successfully")
	return itinerary, nil
}
