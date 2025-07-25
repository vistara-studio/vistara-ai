package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/infra/logger"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/service"
	"github.com/vistara-studio/vistara-ai/pkg/util"
	"github.com/vistara-studio/vistara-ai/pkg/validator"
)

// AIHandler handles general AI-related HTTP requests
type AIHandler struct {
	aiHistorianService *service.AIHistorianService
	nusalingoService   *service.NusalingoService
	logger             *logger.Logger
	config             *config.Config
}

// NewAIHandler creates a new instance of AIHandler
func NewAIHandler(aiHistorianService *service.AIHistorianService, nusalingoService *service.NusalingoService, logger *logger.Logger, cfg *config.Config) *AIHandler {
	return &AIHandler{
		aiHistorianService: aiHistorianService,
		nusalingoService:   nusalingoService,
		logger:             logger,
		config:             cfg,
	}
}

// GenerateHistoricalStory handles historical story generation requests
func (h *AIHandler) GenerateHistoricalStory(c *fiber.Ctx) error {
	// Parse request body
	var req dto.HistoricalStoryRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ResponseWithMessage(c, "Invalid request format", fiber.StatusBadRequest, false)
	}

	// Validate request
	if err := h.validateHistoricalStoryRequest(&req); err != nil {
		return util.ResponseWithMessage(c, err.Error(), fiber.StatusBadRequest, false)
	}

	h.logger.Info("Historical story request validated for location: " + req.Location)

	// Generate historical story using AI service
	story, err := h.aiHistorianService.GenerateHistoricalStory(req.Location)
	if err != nil {
		h.logger.Error("AI Service error during historical story generation: " + err.Error())
		return util.ResponseWithMessage(c, "AI historian service is currently unavailable", fiber.StatusServiceUnavailable, false)
	}

	// Set additional response fields
	story.Location = req.Location
	story.GeneratedAt = time.Now()

	// Return successful response
	return util.ResponseWithData(c, story, "Historical story generated successfully", fiber.StatusOK, true)
}

// TranslateText handles language translation requests via Nusalingo
func (h *AIHandler) TranslateText(c *fiber.Ctx) error {
	// Parse request body
	var req dto.NusalingoRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ResponseWithMessage(c, "Invalid request format", fiber.StatusBadRequest, false)
	}

	// Validate request
	if err := h.validateNusalingoRequest(&req); err != nil {
		return util.ResponseWithMessage(c, err.Error(), fiber.StatusBadRequest, false)
	}

	h.logger.Info(fmt.Sprintf("Translation request validated: %s to %s", req.FromLanguage, req.ToLanguage))

	// Translate text using Nusalingo service
	translatedText, err := h.nusalingoService.TranslateText(&req)
	if err != nil {
		h.logger.Error("AI Service error during translation: " + err.Error())
		return util.ResponseWithMessage(c, "Nusalingo translation service is currently unavailable", fiber.StatusServiceUnavailable, false)
	}

	// Prepare response
	response := &dto.NusalingoResponse{
		TranslatedText: translatedText,
		FromLanguage:   req.FromLanguage,
		ToLanguage:     req.ToLanguage,
		OriginalText:   req.Text,
		GeneratedAt:    time.Now(),
	}

	// Return successful response
	return util.ResponseWithData(c, response, "Translation completed successfully", fiber.StatusOK, true)
}

// validateHistoricalStoryRequest validates the historical story request
func (h *AIHandler) validateHistoricalStoryRequest(req *dto.HistoricalStoryRequest) error {
	validationErrors := validator.ValidateStruct(req)
	if validationErrors != nil {
		return fmt.Errorf("validation failed: %v", validationErrors)
	}
	return nil
}

// validateNusalingoRequest validates the Nusalingo translation request
func (h *AIHandler) validateNusalingoRequest(req *dto.NusalingoRequest) error {
	validationErrors := validator.ValidateStruct(req)
	if validationErrors != nil {
		return fmt.Errorf("validation failed: %v", validationErrors)
	}
	return nil
}
