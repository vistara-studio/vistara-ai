package service

import (
	"log"

	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/util"
)

// SmartPlannerService handles smart trip planning logic
type SmartPlannerService struct {
	geminiService      *GeminiService
	integrationService *IntegrationService
}

// NewSmartPlannerService creates a new smart planner service instance
func NewSmartPlannerService(geminiService *GeminiService, integrationService *IntegrationService) *SmartPlannerService {
	return &SmartPlannerService{
		geminiService:      geminiService,
		integrationService: integrationService,
	}
}

// CreatePlan creates a personalized travel plan using AI with integration data
func (s *SmartPlannerService) CreatePlan(userInput *dto.SmartPlanRequest, duration int, userToken string) (string, error) {
	return s.CreatePlanWithGrounding(userInput, duration, userToken, false)
}

// CreatePlanWithGrounding creates a personalized travel plan using AI with integration data and optional search grounding
func (s *SmartPlannerService) CreatePlanWithGrounding(userInput *dto.SmartPlanRequest, duration int, userToken string, useGrounding bool) (string, error) {
	// Fetch additional data from vistara-be if integration is available
	var localBusinesses []dto.LocalBusiness
	var attractions []dto.TouristAttraction

	if s.integrationService != nil {
		// Fetch local businesses with user token
		if businesses, err := s.integrationService.FetchLocalBusinesses(userInput.Destination, "", userToken); err != nil {
			log.Printf("Warning: Failed to fetch local businesses: %v", err)
		} else {
			localBusinesses = businesses
		}

		// Fetch tourist attractions with user token
		if attr, err := s.integrationService.FetchTouristAttractions(userInput.Destination, userToken); err != nil {
			log.Printf("Warning: Failed to fetch tourist attractions: %v", err)
		} else {
			attractions = attr
		}
	}

	// Format the prompt using validated user input, calculated duration, and integration data
	prompt := util.FormatGeminiPromptWithIntegration(userInput, duration, localBusinesses, attractions)
	
	var logMessage string
	if useGrounding {
		logMessage = "with search grounding for real-time destination information"
	} else {
		logMessage = "with integration data"
	}
	log.Printf("Formatted prompt for AI Smart Planner generation %s", logMessage)

	// Call the Gemini service to generate the itinerary with optional grounding
	var itinerary string
	var err error
	if useGrounding {
		itinerary, err = s.geminiService.GenerateTextWithGrounding(prompt, true)
	} else {
		itinerary, err = s.geminiService.GenerateText(prompt)
	}
	
	if err != nil {
		log.Printf("AI Service error during Smart Planner creation: %v", err)
		return "", err
	}

	// Notify vistara-be about the generated plan if user ID is provided
	if userInput.UserID != nil && *userInput.UserID != "" && s.integrationService != nil {
		if err := s.integrationService.NotifyPlanGenerated(*userInput.UserID, itinerary); err != nil {
			log.Printf("Warning: Failed to notify vistara-be about plan generation: %v", err)
		}
	}

	log.Println("AI Smart Planner itinerary generated successfully")
	return itinerary, nil
}
