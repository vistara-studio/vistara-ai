package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// ItineraryService handles travel itinerary operations
type ItineraryService struct {
	db              *sql.DB
	aiService       *GeminiService
	destinationService *DestinationService
}

// NewItineraryService creates a new itinerary service
func NewItineraryService(db *sql.DB, aiService *GeminiService, destinationService *DestinationService) *ItineraryService {
	return &ItineraryService{
		db:              db,
		aiService:       aiService,
		destinationService: destinationService,
	}
}

// CreateItinerary creates a new travel itinerary with AI-generated plan
func (s *ItineraryService) CreateItinerary(ctx context.Context, userID string, req dto.TravelItineraryRequest) (*dto.TravelItineraryResponse, error) {
	// Generate smart plan using AI service
		planReq := &dto.SmartPlanRequest{
			Destination:         req.DestinationName,
			StartDate:           req.StartDate,
			EndDate:             req.EndDate,
			Budget:              &req.Budget,
			TravelStyle:         nil, // TravelItineraryRequest doesn't have this field
			ActivityIntensity:   nil, // TravelItineraryRequest doesn't have this field  
			ActivityPreferences: req.Interests, // Use interests as activity preferences
		}

		planResp, err := s.aiService.GenerateSmartPlan(planReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI plan: %w", err)
	}

	// Convert AI response to itinerary data
	itineraryData := s.convertAIResponseToItineraryData(planResp, req.StartDate)

	// Marshal itinerary data to JSON
	itineraryJSON, err := json.Marshal(itineraryData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal itinerary data: %w", err)
	}

	// Insert itinerary into database
	itineraryID := uuid.New().String()
	query := `
		INSERT INTO travel_itineraries (
			id, user_id, title, description, destination_name, 
			start_date, end_date, duration_days, total_budget, budget_currency,
			interests, itinerary_data, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err = s.db.QueryRowContext(ctx, query,
		itineraryID, userID, req.Title, req.Description, req.DestinationName,
		req.StartDate, req.EndDate, req.Duration, req.Budget, req.BudgetCurrency,
		pq.Array(req.Interests), string(itineraryJSON), "active",
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create itinerary: %w", err)
	}

	// Track user interaction
	s.trackItineraryInteraction(ctx, userID, itineraryID, "created", map[string]interface{}{
		"destination": req.DestinationName,
		"duration":   req.Duration,
		"budget":     req.Budget,
	})

	// Return complete itinerary response
	return &dto.TravelItineraryResponse{
		ID:              itineraryID,
		UserID:          userID,
		Title:           req.Title,
		Description:     req.Description,
		DestinationName: req.DestinationName,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Duration:        req.Duration,
		TotalBudget:     req.Budget,
		BudgetCurrency:  req.BudgetCurrency,
		Interests:       req.Interests,
		Status:          "active",
		ItineraryData:   itineraryData,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

// GetItinerary retrieves a specific itinerary by ID
func (s *ItineraryService) GetItinerary(ctx context.Context, userID, itineraryID string) (*dto.TravelItineraryResponse, error) {
	query := `
		SELECT id, user_id, title, description, destination_name,
		       start_date, end_date, duration_days, total_budget, budget_currency,
		       interests, itinerary_data, status, created_at, updated_at
		FROM travel_itineraries
		WHERE id = $1 AND user_id = $2
	`

	var itinerary dto.TravelItineraryResponse
	var interestsStr string
	var itineraryDataJSON sql.NullString

	err := s.db.QueryRowContext(ctx, query, itineraryID, userID).Scan(
		&itinerary.ID, &itinerary.UserID, &itinerary.Title, &itinerary.Description,
		&itinerary.DestinationName, &itinerary.StartDate, &itinerary.EndDate,
		&itinerary.Duration, &itinerary.TotalBudget, &itinerary.BudgetCurrency,
		&interestsStr, &itineraryDataJSON, &itinerary.Status,
		&itinerary.CreatedAt, &itinerary.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("itinerary not found")
		}
		return nil, fmt.Errorf("failed to get itinerary: %w", err)
	}

	// Parse interests
	if interestsStr != "" {
		itinerary.Interests = strings.Split(interestsStr, ",")
	}

	// Parse itinerary data
	if itineraryDataJSON.Valid {
		var itineraryData dto.ItineraryData
		if err := json.Unmarshal([]byte(itineraryDataJSON.String), &itineraryData); err == nil {
			itinerary.ItineraryData = &itineraryData
		}
	}

	// Get itinerary destinations
	destinations, err := s.getItineraryDestinations(ctx, itineraryID)
	if err == nil {
		itinerary.Destinations = destinations
	}

	// Track view interaction
	s.trackItineraryInteraction(ctx, userID, itineraryID, "viewed", nil)

	return &itinerary, nil
}

// GetUserItineraries retrieves all itineraries for a user with pagination
func (s *ItineraryService) GetUserItineraries(ctx context.Context, userID string, page, limit int) (*dto.ItineraryHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM travel_itineraries WHERE user_id = $1`
	err := s.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get itineraries
	query := `
		SELECT id, user_id, title, description, destination_name,
		       start_date, end_date, duration_days, total_budget, budget_currency,
		       interests, status, created_at, updated_at
		FROM travel_itineraries
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get itineraries: %w", err)
	}
	defer rows.Close()

	var itineraries []dto.TravelItineraryResponse
	for rows.Next() {
		var itinerary dto.TravelItineraryResponse
		var interestsStr string

		err := rows.Scan(
			&itinerary.ID, &itinerary.UserID, &itinerary.Title, &itinerary.Description,
			&itinerary.DestinationName, &itinerary.StartDate, &itinerary.EndDate,
			&itinerary.Duration, &itinerary.TotalBudget, &itinerary.BudgetCurrency,
			&interestsStr, &itinerary.Status, &itinerary.CreatedAt, &itinerary.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan itinerary: %w", err)
		}

		// Parse interests
		if interestsStr != "" {
			itinerary.Interests = strings.Split(interestsStr, ",")
		}

		itineraries = append(itineraries, itinerary)
	}

	// Get summary statistics
	summary, err := s.getUserItinerarySummary(ctx, userID)
	if err != nil {
		// Don't fail if summary fails, just log and continue
		summary = &dto.ItineraryHistorySummary{}
	}

	totalPages := (totalCount + limit - 1) / limit

	return &dto.ItineraryHistoryResponse{
		Itineraries: itineraries,
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
			TotalItems: totalCount,
		},
		Summary: *summary,
	}, nil
}

// AddDestinationToItinerary adds a destination from explorer to an existing itinerary
func (s *ItineraryService) AddDestinationToItinerary(ctx context.Context, userID, itineraryID string, req dto.AddDestinationToItineraryRequest) (*dto.ItineraryDestinationItem, error) {
	// Verify itinerary belongs to user
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM travel_itineraries WHERE id = $1 AND user_id = $2)`
	err := s.db.QueryRowContext(ctx, checkQuery, itineraryID, userID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("itinerary not found or access denied")
	}

	// Verify destination exists
	destination, err := s.destinationService.GetDestinationByID(ctx, req.DestinationID)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	// Insert itinerary destination
	itemID := uuid.New().String()
	insertQuery := `
		INSERT INTO itinerary_destinations (
			id, itinerary_id, destination_id, day_number, visit_order,
			planned_time, duration_minutes, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at
	`

	var createdAt time.Time
	err = s.db.QueryRowContext(ctx, insertQuery,
		itemID, itineraryID, req.DestinationID, req.DayNumber, req.VisitOrder,
		req.PlannedTime, req.Duration, req.Notes,
	).Scan(&createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add destination to itinerary: %w", err)
	}

	// Track interaction
	s.trackItineraryInteraction(ctx, userID, itineraryID, "added_destination", map[string]interface{}{
		"destination_id": req.DestinationID,
		"day_number":    req.DayNumber,
	})

	// Return the added item
	return &dto.ItineraryDestinationItem{
		ID:            itemID,
		ItineraryID:   itineraryID,
		DestinationID: &req.DestinationID,
		DayNumber:     req.DayNumber,
		VisitOrder:    req.VisitOrder,
		PlannedTime:   req.PlannedTime,
		Duration:      req.Duration,
		Notes:         req.Notes,
		IsCompleted:   false,
		CreatedAt:     createdAt,
		Destination:   destination,
	}, nil
}

// UpdateItineraryStatus updates the status of an itinerary
func (s *ItineraryService) UpdateItineraryStatus(ctx context.Context, userID, itineraryID, status string) error {
	validStatuses := map[string]bool{
		"active":    true,
		"completed": true,
		"cancelled": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `
		UPDATE travel_itineraries 
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`

	result, err := s.db.ExecContext(ctx, query, status, itineraryID, userID)
	if err != nil {
		return fmt.Errorf("failed to update itinerary status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("itinerary not found or access denied")
	}

	// Track status change
	s.trackItineraryInteraction(ctx, userID, itineraryID, "status_changed", map[string]interface{}{
		"new_status": status,
	})

	return nil
}

// Helper functions

func (s *ItineraryService) convertAIResponseToItineraryData(aiResponse *dto.SmartPlanResponse, startDate time.Time) *dto.ItineraryData {
	// Create basic itinerary data structure
	totalBudget := 0.0
	if aiResponse.Budget != nil {
		totalBudget = *aiResponse.Budget
	}

	itineraryData := &dto.ItineraryData{
		TotalBudget: totalBudget,
		Currency:    "USD", // Default currency
		Overview:    fmt.Sprintf("AI-generated travel plan for %s", aiResponse.Destination),
		Tips:        []string{"Follow the generated plan", "Stay flexible with timings", "Check local weather"},
		Days:        []dto.ItineraryDay{},
	}

	// Parse AI response plan (assuming it's a string for now)
	// This is a simplified version - in a real implementation you'd parse the JSON properly
	planText, ok := aiResponse.Plan.(string)
	if !ok {
		// If Plan is not a string, create a simple one-day itinerary
		dayDate := startDate
		itineraryDay := dto.ItineraryDay{
			Day:         1,
			Date:        dayDate,
			Theme:       "Exploration",
			DailyBudget: totalBudget,
			Activities: []dto.ItineraryActivity{
				{
					Time:        "09:00",
					Duration:    180, // 3 hours in minutes
					Name:        "Explore " + aiResponse.Destination,
					Description: "Discover the highlights of the destination",
					Location:    aiResponse.Destination,
					Cost:        totalBudget,
					Tips:        "Start early and bring camera",
				},
			},
			Meals: []dto.ItineraryMeal{
				{
					Type:        "lunch",
					Time:        "12:00",
					Name:        "Local Cuisine",
					Location:    aiResponse.Destination,
					Description: "Try local specialties",
					Cost:        20.0,
					Cuisine:     "Local",
				},
			},
		}
		itineraryData.Days = append(itineraryData.Days, itineraryDay)
		return itineraryData
	}

	// For now, create a simple structure based on the plan text
	// In a real implementation, you'd parse the JSON structure from the AI
	duration := int(aiResponse.EndDate.Sub(aiResponse.StartDate).Hours() / 24)
	if duration < 1 {
		duration = 1
	}

	for i := 0; i < duration; i++ {
		dayDate := startDate.AddDate(0, 0, i)
		dailyBudget := totalBudget / float64(duration)

		itineraryDay := dto.ItineraryDay{
			Day:         i + 1,
			Date:        dayDate,
			Theme:       fmt.Sprintf("Day %d Exploration", i+1),
			DailyBudget: dailyBudget,
			Activities: []dto.ItineraryActivity{
				{
					Time:        "09:00",
					Duration:    180, // 3 hours in minutes
					Name:        fmt.Sprintf("Day %d Activities", i+1),
					Description: planText, // Use the AI plan text
					Location:    aiResponse.Destination,
					Cost:        dailyBudget,
					Tips:        "Check the AI plan for details",
				},
			},
			Meals: []dto.ItineraryMeal{
				{
					Type:        "lunch",
					Time:        "12:00",
					Name:        "Local Cuisine",
					Location:    aiResponse.Destination,
					Description: "Try local specialties",
					Cost:        15.0,
					Cuisine:     "Local",
				},
			},
		}
		itineraryData.Days = append(itineraryData.Days, itineraryDay)
	}

	return itineraryData
}

func (s *ItineraryService) getItineraryDestinations(ctx context.Context, itineraryID string) ([]dto.ItineraryDestinationItem, error) {
	query := `
		SELECT id.id, id.destination_id, id.day_number, id.visit_order,
		       id.planned_time, id.duration_minutes, id.notes, id.is_completed, id.created_at
		FROM itinerary_destinations id
		WHERE id.itinerary_id = $1
		ORDER BY id.day_number, id.visit_order
	`

	rows, err := s.db.QueryContext(ctx, query, itineraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var destinations []dto.ItineraryDestinationItem
	for rows.Next() {
		var item dto.ItineraryDestinationItem
		var destinationID sql.NullString

		err := rows.Scan(
			&item.ID, &destinationID, &item.DayNumber, &item.VisitOrder,
			&item.PlannedTime, &item.Duration, &item.Notes, &item.IsCompleted, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.ItineraryID = itineraryID
		if destinationID.Valid {
			item.DestinationID = &destinationID.String
		}

		destinations = append(destinations, item)
	}

	return destinations, nil
}

func (s *ItineraryService) getUserItinerarySummary(ctx context.Context, userID string) (*dto.ItineraryHistorySummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_itineraries,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_itineraries,
			COALESCE(SUM(total_budget), 0) as total_budget_spent
		FROM travel_itineraries
		WHERE user_id = $1
	`

	var summary dto.ItineraryHistorySummary
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&summary.TotalItineraries,
		&summary.CompletedItineraries,
		&summary.TotalBudgetSpent,
	)

	if err != nil {
		return nil, err
	}

	summary.Currency = "IDR" // Default currency
	summary.MostVisitedCategory = "Culture" // TODO: Calculate from actual data

	return &summary, nil
}

func (s *ItineraryService) trackItineraryInteraction(ctx context.Context, userID, itineraryID, interactionType string, metadata map[string]interface{}) {
	metadataJSON, _ := json.Marshal(metadata)
	
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_interactions (user_id, itinerary_id, interaction_type, metadata, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, userID, itineraryID, interactionType, string(metadataJSON))
	
	if err != nil {
		// Log error but don't fail the main operation
		fmt.Printf("Failed to track interaction: %v\n", err)
	}
}
