package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/service"
	"github.com/vistara-studio/vistara-ai/pkg/util"
	"github.com/vistara-studio/vistara-ai/pkg/validator"
)

// DestinationHandler handles destination-related HTTP requests
type DestinationHandler struct {
	destinationService *service.DestinationService
	validator          *validator.Validator
}

// NewDestinationHandler creates a new destination handler
func NewDestinationHandler(destinationService *service.DestinationService, validator *validator.Validator) *DestinationHandler {
	return &DestinationHandler{
		destinationService: destinationService,
		validator:          validator,
	}
}

// GetDestinations handles GET /v1/destinations
func (h *DestinationHandler) GetDestinations(c *fiber.Ctx) error {
	var req dto.DestinationRequest

	// Parse query parameters
	if lat := c.Query("lat"); lat != "" {
		if latFloat, err := strconv.ParseFloat(lat, 64); err == nil {
			req.Lat = &latFloat
		}
	}

	if long := c.Query("long"); long != "" {
		if longFloat, err := strconv.ParseFloat(long, 64); err == nil {
			req.Long = &longFloat
		}
	}

	if radius := c.Query("radius"); radius != "" {
		if radiusFloat, err := strconv.ParseFloat(radius, 64); err == nil {
			req.Radius = &radiusFloat
		}
	}

	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	if category := c.Query("category"); category != "" {
		req.Category = &category
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = &sortBy
	}

	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil {
			req.Limit = &limitInt
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil {
			req.Offset = &offsetInt
		}
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request parameters",
			"errors":  err,
		})
	}

	// Get destinations with user context if authenticated
	var destinations []dto.DestinationResponse
	var err error

	if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
		destinations, err = h.destinationService.GetDestinationsForUser(c.Context(), req, userID)
	} else {
		destinations, err = h.destinationService.GetDestinations(c.Context(), req)
	}

	if err != nil {
		util.LogError("Failed to get destinations", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve destinations",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    destinations,
		"meta": fiber.Map{
			"count":  len(destinations),
			"limit":  req.Limit,
			"offset": req.Offset,
		},
		"filters": fiber.Map{
			"search":   req.Search,
			"category": req.Category,
			"sort_by":  req.SortBy,
			"location": fiber.Map{
				"lat":    req.Lat,
				"long":   req.Long,
				"radius": req.Radius,
			},
		},
	})
}

// GetDestinationByID handles GET /v1/destinations/{id}
func (h *DestinationHandler) GetDestinationByID(c *fiber.Ctx) error {
	destinationID := c.Params("id")
	if destinationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Destination ID is required",
		})
	}

	destination, err := h.destinationService.GetDestinationByID(c.Context(), destinationID)
	if err != nil {
		if err.Error() == "destination not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Destination not found",
			})
		}

		util.LogError("Failed to get destination", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve destination",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    destination,
	})
}

// GenerateNarration handles POST /v1/manuscripts/{id}/narrate
func (h *DestinationHandler) GenerateNarration(c *fiber.Ctx) error {
	manuscriptID := c.Params("id")
	if manuscriptID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Manuscript ID is required",
		})
	}

	var req dto.NarrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request parameters",
			"errors":  err,
		})
	}

	// Generate narration
	narration, err := h.destinationService.GenerateNarration(c.Context(), manuscriptID, req)
	if err != nil {
		if err.Error() == "manuscript not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Manuscript not found",
			})
		}

		util.LogError("Failed to generate narration", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate audio narration",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    narration,
		"message": "Audio narration generated successfully",
	})
}

// CreateItinerary handles POST /v1/itineraries
func (h *DestinationHandler) CreateItinerary(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User authentication required",
		})
	}

	var req dto.ItineraryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request parameters",
			"errors":  err,
		})
	}

	// Create itinerary
	itinerary, err := h.destinationService.CreateItinerary(c.Context(), userID, req)
	if err != nil {
		util.LogError("Failed to create itinerary", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create itinerary",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    itinerary,
		"message": "Itinerary created successfully",
	})
}

// AddDestinationToItinerary handles POST /v1/itineraries/{itineraryId}/destinations
func (h *DestinationHandler) AddDestinationToItinerary(c *fiber.Ctx) error {
	itineraryID := c.Params("itineraryId")
	if itineraryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Itinerary ID is required",
		})
	}

	var req dto.ItineraryDestinationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request parameters",
			"errors":  err,
		})
	}

	// Add destination to itinerary
	err := h.destinationService.AddDestinationToItinerary(c.Context(), itineraryID, req)
	if err != nil {
		if err.Error() == "destination not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Destination not found",
			})
		}

		util.LogError("Failed to add destination to itinerary", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to add destination to itinerary",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Destination added to itinerary successfully",
	})
}

// TODO: Implement GetRecommendations once RecommendationService is integrated
// GetRecommendations handles GET /v1/recommendations
/*
func (h *DestinationHandler) GetRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User authentication required",
		})
	}

	var req dto.RecommendationRequest
	req.UserID = userID

	// Parse query parameters
	if tags := c.Query("tags"); tags != "" {
		// Split comma-separated tags
		req.Tags = strings.Split(tags, ",")
		for i, tag := range req.Tags {
			req.Tags[i] = strings.TrimSpace(tag)
		}
	}

	req.Limit = 10 // Default limit
	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 && limitInt <= 20 {
			req.Limit = limitInt
		}
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request parameters",
			"errors":  err,
		})
	}

	// Get recommendations
	recommendations, err := h.destinationService.GetRecommendations(c.Context(), req)
	if err != nil {
		util.LogError("Failed to get recommendations", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve recommendations",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    recommendations,
		"meta": fiber.Map{
			"count":   len(recommendations),
			"user_id": userID,
		},
	})
}
*/

// GetMapView handles GET /v1/destinations/map
func (h *DestinationHandler) GetMapView(c *fiber.Ctx) error {
	var req dto.MapViewRequest
	
	// Parse query parameters
	if lat := c.Query("lat"); lat != "" {
		if latFloat, err := strconv.ParseFloat(lat, 64); err == nil {
			req.Lat = latFloat
		}
	}
	
	if long := c.Query("long"); long != "" {
		if longFloat, err := strconv.ParseFloat(long, 64); err == nil {
			req.Long = longFloat
		}
	}
	
	if zoom := c.Query("zoom"); zoom != "" {
		if zoomInt, err := strconv.Atoi(zoom); err == nil {
			req.Zoom = zoomInt
		}
	}
	
	// Parse bounds if provided
	if neLat := c.Query("ne_lat"); neLat != "" {
		if neLong := c.Query("ne_long"); neLong != "" {
			if swLat := c.Query("sw_lat"); swLat != "" {
				if swLong := c.Query("sw_long"); swLong != "" {
					if neLat64, err1 := strconv.ParseFloat(neLat, 64); err1 == nil {
						if neLong64, err2 := strconv.ParseFloat(neLong, 64); err2 == nil {
							if swLat64, err3 := strconv.ParseFloat(swLat, 64); err3 == nil {
								if swLong64, err4 := strconv.ParseFloat(swLong, 64); err4 == nil {
									req.Bounds = dto.MapBounds{
										NorthEast: dto.LocationResponse{Latitude: neLat64, Longitude: neLong64},
										SouthWest: dto.LocationResponse{Latitude: swLat64, Longitude: swLong64},
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request parameters",
			"errors":  err,
		})
	}

	// Get map data
	mapData, err := h.destinationService.GetMapView(c.Context(), req)
	if err != nil {
		util.LogError("Failed to get map view", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve map data",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    mapData,
	})
}

// ToggleBookmark toggles bookmark status for a destination
func (h *DestinationHandler) ToggleBookmark(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	var req dto.BookmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"message": "Invalid request format",
		})
	}

	response, err := h.destinationService.ToggleBookmark(c.Context(), userID, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   response,
	})
}

// GetUserBookmarks retrieves user's bookmarked destinations
func (h *DestinationHandler) GetUserBookmarks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	destinations, err := h.destinationService.GetUserBookmarks(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"destinations": destinations,
			"pagination": fiber.Map{
				"limit":  limit,
				"offset": offset,
				"count":  len(destinations),
			},
		},
	})
}

// GetCategories handles GET /v1/destinations/categories
func (h *DestinationHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.destinationService.GetCategories(c.Context())
	if err != nil {
		util.LogError("Failed to get categories", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve categories",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    categories,
	})
}

// GetPhotos handles GET /v1/destinations/{id}/photos
func (h *DestinationHandler) GetPhotos(c *fiber.Ctx) error {
	destinationID := c.Params("id")
	if destinationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Destination ID is required",
		})
	}

	photos, err := h.destinationService.GetPhotos(c.Context(), destinationID)
	if err != nil {
		if err.Error() == "destination not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Destination not found",
			})
		}

		util.LogError("Failed to get photos", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve photos",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    photos,
	})
}
