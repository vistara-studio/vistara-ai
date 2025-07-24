package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/service"
	"github.com/vistara-studio/vistara-ai/pkg/util"
	"github.com/vistara-studio/vistara-ai/pkg/validator"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	smartPlannerService *service.SmartPlannerService
	config              *config.Config
}

// NewAIHandler creates a new AI handler
func NewAIHandler(smartPlannerService *service.SmartPlannerService, cfg *config.Config) *AIHandler {
	return &AIHandler{
		smartPlannerService: smartPlannerService,
		config:              cfg,
	}
}

// GenerateSmartPlan handles smart travel planning requests
func (h *AIHandler) GenerateSmartPlan(c *fiber.Ctx) error {
	// Parse request body
	var req dto.SmartPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ResponseWithMessage(c, "Invalid JSON input", fiber.StatusBadRequest, false)
	}

	// Validate request
	if validationErrors := validator.ValidateStruct(&req); validationErrors != nil {
		return util.ResponseWithData(c, validationErrors, "Input validation failed", fiber.StatusBadRequest, false)
	}

	// Custom validation for dates
	if err := h.validateSmartPlanRequest(&req); err != nil {
		return util.ResponseWithMessage(c, err.Error(), fiber.StatusBadRequest, false)
	}

	// Calculate duration from dates
	duration := int(req.EndDate.Sub(req.StartDate).Hours()/24) + 1
	if duration <= 0 {
		return util.ResponseWithMessage(c, "End date must be after start date", fiber.StatusBadRequest, false)
	}

	log.Printf("Smart plan request validated for destination: %s, duration: %d days", req.Destination, duration)

	// Generate smart plan using AI service with calculated duration
	rawItinerary, err := h.smartPlannerService.CreatePlan(&req, duration)
	if err != nil {
		log.Printf("AI Service error during Smart Plan creation: %v", err)
		return util.ResponseWithMessage(c, "AI Smart Planner service is currently unavailable", fiber.StatusServiceUnavailable, false)
	}

	// Clean the response from markdown code blocks
	cleanedItinerary := h.cleanAIResponse(rawItinerary)

	// Parse and validate JSON response
	var parsedItinerary map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedItinerary), &parsedItinerary); err != nil {
		log.Printf("Failed to parse AI response as JSON: %v", err)
		log.Printf("Raw response: %s", rawItinerary)
		log.Printf("Cleaned response: %s", cleanedItinerary)
		return util.ResponseWithMessage(c, "AI service returned invalid response format", fiber.StatusInternalServerError, false)
	}

	// Prepare response
	response := dto.SmartPlanResponse{
		Plan:              parsedItinerary,
		Destination:       req.Destination,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		Budget:            req.Budget,
		TravelStyle:       req.TravelStyle,
		ActivityIntensity: req.ActivityIntensity,
		GeneratedAt:       time.Now(),
	}

	return util.ResponseWithData(c, response, "Smart plan generated successfully", fiber.StatusOK, true)
}

// validateSmartPlanRequest validates the smart plan request
func (h *AIHandler) validateSmartPlanRequest(req *dto.SmartPlanRequest) error {
	// Calculate duration from dates
	duration := int(req.EndDate.Sub(req.StartDate).Hours()/24) + 1
	
	// Check if end date is after start date
	if req.EndDate.Before(req.StartDate) || req.EndDate.Equal(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	// Check minimum and maximum trip duration
	if duration < 1 {
		return fmt.Errorf("trip duration must be at least 1 day")
	}
	if duration > 30 {
		return fmt.Errorf("trip duration cannot exceed 30 days")
	}

	return nil
}

// cleanAIResponse removes markdown code blocks and other formatting issues from AI response
func (h *AIHandler) cleanAIResponse(response string) string {
	// Remove markdown code blocks (```json and ```)
	re := regexp.MustCompile("```(?:json)?\\s*")
	cleaned := re.ReplaceAllString(response, "")
	
	// Remove trailing ```
	cleaned = strings.TrimSuffix(cleaned, "```")
	
	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	// Remove any leading/trailing backticks
	cleaned = strings.Trim(cleaned, "`")
	
	return cleaned
}
