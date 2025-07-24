package service

import (
	"log"

	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/util"
)

// SmartPlannerService handles smart trip planning logic
type SmartPlannerService struct {
	geminiService *GeminiService
}

// NewSmartPlannerService creates a new smart planner service instance
func NewSmartPlannerService(geminiService *GeminiService) *SmartPlannerService {
	return &SmartPlannerService{
		geminiService: geminiService,
	}
}

// CreatePlan creates a personalized travel plan using AI
func (s *SmartPlannerService) CreatePlan(userInput *dto.SmartPlanRequest, duration int) (string, error) {
	// Format the prompt using validated user input and calculated duration
	prompt := util.FormatGeminiPrompt(userInput, duration)
	log.Println("Formatted prompt for AI Smart Planner generation")

	// Call the Gemini service to generate the itinerary
	itinerary, err := s.geminiService.GenerateText(prompt)
	if err != nil {
		log.Printf("AI Service error during Smart Planner creation: %v", err)
		return "", err
	}

	log.Println("AI Smart Planner itinerary generated successfully")
	return itinerary, nil
}
