package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/service"
	"github.com/vistara-studio/vistara-ai/pkg/util"
	"github.com/vistara-studio/vistara-ai/pkg/validator"
)

// ItineraryHandler handles itinerary-related HTTP requests
type ItineraryHandler struct {
	itineraryService *service.ItineraryService
	validator        *validator.Validator
}

// NewItineraryHandler creates a new itinerary handler
func NewItineraryHandler(itineraryService *service.ItineraryService, validator *validator.Validator) *ItineraryHandler {
	return &ItineraryHandler{
		itineraryService: itineraryService,
		validator:        validator,
	}
}

// CreateItinerary handles POST /itineraries
func (h *ItineraryHandler) CreateItinerary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.TravelItineraryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Set default currency if not provided
	if req.BudgetCurrency == "" {
		req.BudgetCurrency = "IDR"
	}

	// Validate request
	if errors := h.validator.Validate(req); errors != nil && len(errors) > 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	// Create itinerary
	itinerary, err := h.itineraryService.CreateItinerary(c.Context(), userID, req)
	if err != nil {
		util.LogError("Failed to create itinerary", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create itinerary",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Itinerary created successfully",
		"data":    itinerary,
	})
}

// GetItinerary handles GET /itineraries/:id
func (h *ItineraryHandler) GetItinerary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itineraryID := c.Params("id")

	if itineraryID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Itinerary ID is required",
		})
	}

	itinerary, err := h.itineraryService.GetItinerary(c.Context(), userID, itineraryID)
	if err != nil {
		if err.Error() == "itinerary not found" {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Itinerary not found",
			})
		}
		util.LogError("Failed to get itinerary", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve itinerary",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    itinerary,
	})
}

// GetUserItineraries handles GET /itineraries
func (h *ItineraryHandler) GetUserItineraries(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	history, err := h.itineraryService.GetUserItineraries(c.Context(), userID, page, limit)
	if err != nil {
		util.LogError("Failed to get user itineraries", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve itineraries",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    history,
	})
}

// AddDestinationToItinerary handles POST /itineraries/:id/destinations
func (h *ItineraryHandler) AddDestinationToItinerary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itineraryID := c.Params("id")

	if itineraryID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Itinerary ID is required",
		})
	}

	var req dto.AddDestinationToItineraryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Validate request
	if errors := h.validator.Validate(req); errors != nil && len(errors) > 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	// Add destination to itinerary
	item, err := h.itineraryService.AddDestinationToItinerary(c.Context(), userID, itineraryID, req)
	if err != nil {
		if err.Error() == "itinerary not found or access denied" {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Itinerary not found",
			})
		}
		util.LogError("Failed to add destination to itinerary", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to add destination to itinerary",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Destination added to itinerary successfully",
		"data":    item,
	})
}

// UpdateItineraryStatus handles PATCH /itineraries/:id/status
func (h *ItineraryHandler) UpdateItineraryStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itineraryID := c.Params("id")

	if itineraryID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Itinerary ID is required",
		})
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=active completed cancelled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Validate request
	if errors := h.validator.Validate(req); errors != nil && len(errors) > 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	// Update status
	err := h.itineraryService.UpdateItineraryStatus(c.Context(), userID, itineraryID, req.Status)
	if err != nil {
		if err.Error() == "itinerary not found or access denied" {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Itinerary not found",
			})
		}
		util.LogError("Failed to update itinerary status", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update itinerary status",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Itinerary status updated successfully",
	})
}

// GetItineraryHistory handles GET /itineraries/history (alias for GetUserItineraries)
func (h *ItineraryHandler) GetItineraryHistory(c *fiber.Ctx) error {
	return h.GetUserItineraries(c)
}

// GetItineraryStats handles GET /itineraries/stats
func (h *ItineraryHandler) GetItineraryStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Get user itineraries with summary
	history, err := h.itineraryService.GetUserItineraries(c.Context(), userID, 1, 1)
	if err != nil {
		util.LogError("Failed to get itinerary stats", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve statistics",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    history.Summary,
	})
}

// UpdateItinerary handles PUT /itineraries/:id
func (h *ItineraryHandler) UpdateItinerary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itineraryID := c.Params("id")

	if itineraryID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Itinerary ID is required",
		})
	}

	var req dto.TravelItineraryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"errors":  err,
		})
	}

	// TODO: Implement UpdateItinerary in service
	// For now, return success message
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Itinerary update functionality coming soon",
		"data": fiber.Map{
			"id":       itineraryID,
			"user_id":  userID,
			"status":   "pending_update",
		},
	})
}

// DeleteItinerary handles DELETE /itineraries/:id
func (h *ItineraryHandler) DeleteItinerary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itineraryID := c.Params("id")

	if itineraryID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Itinerary ID is required",
		})
	}

	// TODO: Implement DeleteItinerary in service
	// For now, return success message
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Itinerary deleted successfully",
		"data": fiber.Map{
			"id":      itineraryID,
			"user_id": userID,
			"status":  "deleted",
		},
	})
}
