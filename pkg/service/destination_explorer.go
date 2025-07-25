package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/pkg/util"
)

// DestinationService handles destination-related operations
type DestinationService struct {
	db             *sql.DB
	narrationSvc   *NarrationService
	recommendSvc   *RecommendationService
}

// NewDestinationService creates a new destination service
func NewDestinationService(db *sql.DB, narrationSvc *NarrationService, recommendSvc *RecommendationService) *DestinationService {
	return &DestinationService{
		db:           db,
		narrationSvc: narrationSvc,
		recommendSvc: recommendSvc,
	}
}

// GetDestinations retrieves destinations based on search criteria
func (s *DestinationService) GetDestinations(ctx context.Context, req dto.DestinationRequest) ([]dto.DestinationResponse, error) {
	return s.getDestinationsWithUser(ctx, req, nil)
}

// GetDestinationsForUser retrieves destinations with bookmark status for specific user
func (s *DestinationService) GetDestinationsForUser(ctx context.Context, req dto.DestinationRequest, userID string) ([]dto.DestinationResponse, error) {
	return s.getDestinationsWithUser(ctx, req, &userID)
}

// getDestinationsWithUser internal method that handles both with and without user context
func (s *DestinationService) getDestinationsWithUser(ctx context.Context, req dto.DestinationRequest, userID *string) ([]dto.DestinationResponse, error) {
	query := `
		SELECT 
			d.id, d.name, d.description, d.short_description,
			ST_Y(d.location::geometry) as latitude, 
			ST_X(d.location::geometry) as longitude,
			d.main_image_url, d.images, d.category, d.tags,
			d.price_amount, d.price_currency, d.price_unit,
			d.booking_required, d.booking_url, d.open_hours,
			d.created_at,
			COALESCE(AVG(r.rating), 0) as avg_rating,
			COUNT(r.id) as review_count
	`
	
	// Add bookmark status if userID provided
	if userID != nil {
		query += `,
			CASE WHEN ub.user_id IS NOT NULL THEN true ELSE false END as is_bookmarked
		`
	}

	args := []interface{}{}
	whereConditions := []string{}
	argIndex := 1

	// Add distance calculation if coordinates provided
	if req.Lat != nil && req.Long != nil {
		query += fmt.Sprintf(`,
			ST_Distance(
				d.location,
				ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography
			) / 1000 as distance_km
		`, argIndex, argIndex+1)
		args = append(args, *req.Long, *req.Lat)
		argIndex += 2

		// Add radius filter if provided
		if req.Radius != nil {
			whereConditions = append(whereConditions, fmt.Sprintf(`
				ST_DWithin(
					d.location,
					ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography,
					$%d
				)
			`, argIndex, argIndex+1, argIndex+2))
			args = append(args, *req.Long, *req.Lat, *req.Radius*1000) // Convert km to meters
			argIndex += 3
		}
	}

	query += `
		FROM destinations d
		LEFT JOIN reviews r ON d.id = r.destination_id
	`
	
	// Add bookmark join if userID provided
	if userID != nil {
		query += fmt.Sprintf(`
			LEFT JOIN user_bookmarks ub ON d.id = ub.destination_id AND ub.user_id = $%d
		`, argIndex)
		args = append(args, *userID)
		argIndex++
	}

	// Add search filter
	if req.Search != nil && *req.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf(`
			(d.name ILIKE $%d OR d.description ILIKE $%d OR d.category ILIKE $%d)
		`, argIndex, argIndex, argIndex))
		searchTerm := "%" + strings.ToLower(*req.Search) + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	// Add category filter
	if req.Category != nil && *req.Category != "" && *req.Category != "All" {
		whereConditions = append(whereConditions, fmt.Sprintf(`d.category = $%d`, argIndex))
		args = append(args, *req.Category)
		argIndex++
	}

	// Add WHERE clause if conditions exist
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add GROUP BY
	groupByFields := `d.id, d.name, d.description, d.short_description, d.location, 
		         d.main_image_url, d.images, d.category, d.tags, d.price_amount,
		         d.price_currency, d.price_unit, d.booking_required, d.booking_url,
		         d.open_hours, d.created_at`
	
	if userID != nil {
		groupByFields += `, ub.user_id`
	}
	
	query += " GROUP BY " + groupByFields

	// Add ordering
	sortBy := "name"
	if req.SortBy != nil {
		sortBy = *req.SortBy
	}

	switch sortBy {
	case "distance":
		if req.Lat != nil && req.Long != nil {
			query += " ORDER BY distance_km ASC"
		} else {
			query += " ORDER BY d.name ASC"
		}
	case "rating":
		query += " ORDER BY avg_rating DESC, review_count DESC"
	case "popularity":
		query += " ORDER BY review_count DESC, avg_rating DESC"
	default:
		query += " ORDER BY d.name ASC"
	}

	// Add pagination
	limit := 20
	if req.Limit != nil && *req.Limit > 0 && *req.Limit <= 100 {
		limit = *req.Limit
	}
	offset := 0
	if req.Offset != nil && *req.Offset >= 0 {
		offset = *req.Offset
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query destinations: %w", err)
	}
	defer rows.Close()

	var destinations []dto.DestinationResponse
	for rows.Next() {
		var dest dto.DestinationResponse
		var distance sql.NullFloat64
		var avgRating sql.NullFloat64
		var reviewCount sql.NullInt64
		var images, tags []string
		var priceAmount sql.NullFloat64
		var priceCurrency, priceUnit sql.NullString
		var bookingRequired sql.NullBool
		var bookingURL, openHours sql.NullString
		var isBookmarked sql.NullBool

		scanArgs := []interface{}{
			&dest.ID, &dest.Name, &dest.Description, &dest.ShortDesc,
			&dest.Location.Latitude, &dest.Location.Longitude,
			&dest.MainImage, &images, &dest.Category, &tags,
			&priceAmount, &priceCurrency, &priceUnit,
			&bookingRequired, &bookingURL, &openHours,
			&dest.CreatedAt, &avgRating, &reviewCount,
		}

		// Add bookmark status if userID provided
		if userID != nil {
			scanArgs = append(scanArgs, &isBookmarked)
		}

		// Add distance to scan if coordinates were provided
		if req.Lat != nil && req.Long != nil {
			scanArgs = append(scanArgs, &distance)
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan destination: %w", err)
		}

		// Set calculated fields
		if distance.Valid {
			dest.Distance = &distance.Float64
		}
		if avgRating.Valid {
			dest.Rating = avgRating.Float64
		}
		if reviewCount.Valid {
			dest.ReviewCount = int(reviewCount.Int64)
		}
		if userID != nil && isBookmarked.Valid {
			dest.IsBookmarked = isBookmarked.Bool
		}

		dest.Images = images
		dest.Tags = tags

		// Set price info
		if priceAmount.Valid && priceAmount.Float64 > 0 {
			dest.Price = &dto.PriceInfo{
				Amount:   priceAmount.Float64,
				Currency: priceCurrency.String,
				Unit:     priceUnit.String,
			}
		}

		// Set booking info
		if bookingRequired.Valid || bookingURL.Valid || openHours.Valid {
			dest.BookingInfo = &dto.BookingInfo{
				IsAvailable:    true,
				TicketRequired: bookingRequired.Bool,
				BookingURL:     bookingURL.String,
				OpenHours:      openHours.String,
			}
		}

		dest.IsBookmarkable = true // All destinations are bookmarkable

		destinations = append(destinations, dest)
	}

	return destinations, nil
}

// GetDestinationByID retrieves a single destination with full details
func (s *DestinationService) GetDestinationByID(ctx context.Context, destinationID string) (*dto.DestinationResponse, error) {
	// First get the destination
	query := `
		SELECT 
			id, name, description, 
			ST_Y(location::geometry) as latitude, 
			ST_X(location::geometry) as longitude,
			main_image_url, created_at
		FROM destinations 
		WHERE id = $1
	`

	var dest dto.DestinationResponse
	err := s.db.QueryRowContext(ctx, query, destinationID).Scan(
		&dest.ID, &dest.Name, &dest.Description,
		&dest.Location.Latitude, &dest.Location.Longitude,
		&dest.MainImage, &dest.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("destination not found")
		}
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	// Get manuscript details
	manuscript, err := s.getManuscriptByDestinationID(ctx, destinationID)
	if err == nil {
		dest.Manuscript = manuscript
	}

	// Get reviews
	reviews, err := s.getReviewsByDestinationID(ctx, destinationID)
	if err == nil {
		dest.Reviews = reviews
	}

	return &dest, nil
}

// getManuscriptByDestinationID retrieves manuscript for a destination
func (s *DestinationService) getManuscriptByDestinationID(ctx context.Context, destinationID string) (*dto.ManuscriptResponse, error) {
	query := `
		SELECT 
			id, title, cultural_story, original_script_text, 
			translation_text, manuscript_image_url, created_at
		FROM manuscripts 
		WHERE destination_id = $1
	`

	var manuscript dto.ManuscriptResponse
	err := s.db.QueryRowContext(ctx, query, destinationID).Scan(
		&manuscript.ID, &manuscript.Title, &manuscript.CulturalStory,
		&manuscript.OriginalScriptText, &manuscript.TranslationText,
		&manuscript.ManuscriptImageURL, &manuscript.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &manuscript, nil
}

// getReviewsByDestinationID retrieves reviews for a destination
func (s *DestinationService) getReviewsByDestinationID(ctx context.Context, destinationID string) ([]dto.ReviewResponse, error) {
	query := `
		SELECT id, user_id, rating, comment, created_at
		FROM reviews 
		WHERE destination_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`

	rows, err := s.db.QueryContext(ctx, query, destinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []dto.ReviewResponse
	for rows.Next() {
		var review dto.ReviewResponse
		if err := rows.Scan(&review.ID, &review.UserID, &review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			continue // Skip invalid reviews
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

// GenerateNarration generates audio narration for a manuscript
func (s *DestinationService) GenerateNarration(ctx context.Context, manuscriptID string, req dto.NarrationRequest) (*dto.NarrationResponse, error) {
	// Get manuscript text
	var culturalStory string
	query := `SELECT cultural_story FROM manuscripts WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, manuscriptID).Scan(&culturalStory)
	if err != nil {
		return nil, fmt.Errorf("manuscript not found: %w", err)
	}

	// Generate narration using AI service
	narrationResp, err := s.narrationSvc.GenerateAudio(ctx, culturalStory, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate narration: %w", err)
	}

	// Save narration URL to database (optional)
	updateQuery := `
		UPDATE manuscripts 
		SET audio_narration_url = $1, updated_at = NOW() 
		WHERE id = $2
	`
	_, err = s.db.ExecContext(ctx, updateQuery, narrationResp.AudioURL, manuscriptID)
	if err != nil {
		// Log error but don't fail the request
		util.LogError("Failed to update manuscript with narration URL", err)
	}

	return narrationResp, nil
}

// CreateItinerary creates a new itinerary for a user
func (s *DestinationService) CreateItinerary(ctx context.Context, userID string, req dto.ItineraryRequest) (*dto.ItineraryResponse, error) {
	itineraryID := uuid.New().String()

	query := `
		INSERT INTO itineraries (id, user_id, name, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := s.db.ExecContext(ctx, query, itineraryID, userID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create itinerary: %w", err)
	}

	return &dto.ItineraryResponse{
		ID:           itineraryID,
		UserID:       userID,
		Name:         req.Name,
		Destinations: []dto.ItineraryDestinationResponse{},
		CreatedAt:    time.Now(),
	}, nil
}

// AddDestinationToItinerary adds a destination to an itinerary
func (s *DestinationService) AddDestinationToItinerary(ctx context.Context, itineraryID string, req dto.ItineraryDestinationRequest) error {
	// Parse visit date
	visitDate, err := time.Parse("2006-01-02", req.VisitDate)
	if err != nil {
		return fmt.Errorf("invalid visit date format: %w", err)
	}

	// Verify destination exists
	var destinationExists bool
	query := `SELECT EXISTS(SELECT 1 FROM destinations WHERE id = $1)`
	err = s.db.QueryRowContext(ctx, query, req.DestinationID).Scan(&destinationExists)
	if err != nil || !destinationExists {
		return fmt.Errorf("destination not found")
	}

	// Add to itinerary
	insertQuery := `
		INSERT INTO itinerary_destinations (itinerary_id, destination_id, visit_date, notes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (itinerary_id, destination_id) 
		DO UPDATE SET visit_date = $3, notes = $4
	`

	_, err = s.db.ExecContext(ctx, insertQuery, itineraryID, req.DestinationID, visitDate, req.Notes)
	if err != nil {
		return fmt.Errorf("failed to add destination to itinerary: %w", err)
	}

	return nil
}

// GetMapView retrieves destinations for map display
func (s *DestinationService) GetMapView(ctx context.Context, req dto.MapViewRequest) (*dto.MapViewResponse, error) {
	query := `
		SELECT 
			d.id, d.name, 
			ST_Y(d.location::geometry) as latitude, 
			ST_X(d.location::geometry) as longitude,
			d.category, d.main_image_url,
			COALESCE(AVG(r.rating), 0) as avg_rating
		FROM destinations d
		LEFT JOIN reviews r ON d.id = r.destination_id
		WHERE ST_DWithin(
			d.location,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)
		GROUP BY d.id, d.name, d.location, d.category, d.main_image_url
		ORDER BY 
			ST_Distance(
				d.location,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
			) ASC
		LIMIT 50
	`

	// Calculate radius based on zoom level (rough approximation)
	radius := s.calculateRadiusFromZoom(req.Zoom)

	rows, err := s.db.QueryContext(ctx, query, req.Long, req.Lat, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to query map destinations: %w", err)
	}
	defer rows.Close()

	var destinations []dto.MapDestination
	for rows.Next() {
		var dest dto.MapDestination
		var rating sql.NullFloat64

		if err := rows.Scan(
			&dest.ID, &dest.Name,
			&dest.Location.Latitude, &dest.Location.Longitude,
			&dest.Category, &dest.MainImage, &rating,
		); err != nil {
			continue
		}

		if rating.Valid {
			dest.Rating = rating.Float64
		}

		destinations = append(destinations, dest)
	}

	return &dto.MapViewResponse{
		Destinations: destinations,
		Center: dto.LocationResponse{
			Latitude:  req.Lat,
			Longitude: req.Long,
		},
		Zoom: req.Zoom,
	}, nil
}

// calculateRadiusFromZoom calculates search radius based on map zoom level
func (s *DestinationService) calculateRadiusFromZoom(zoom int) float64 {
	// Rough approximation: higher zoom = smaller radius
	baseRadius := 100000.0 // 100km at zoom 1
	return baseRadius / math.Pow(2, float64(zoom-1))
}

// ToggleBookmark adds or removes a destination from user's bookmarks
func (s *DestinationService) ToggleBookmark(ctx context.Context, userID string, req dto.BookmarkRequest) (*dto.BookmarkResponse, error) {
	// Check if destination exists
	var destinationExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM destinations WHERE id = $1)`
	err := s.db.QueryRowContext(ctx, checkQuery, req.DestinationID).Scan(&destinationExists)
	if err != nil || !destinationExists {
		return nil, fmt.Errorf("destination not found")
	}

	var isBookmarked bool
	var message string

	if req.Action == "add" {
		// Add bookmark
		insertQuery := `
			INSERT INTO user_bookmarks (user_id, destination_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, destination_id) DO NOTHING
		`
		_, err = s.db.ExecContext(ctx, insertQuery, userID, req.DestinationID)
		if err != nil {
			return nil, fmt.Errorf("failed to add bookmark: %w", err)
		}
		isBookmarked = true
		message = "Destination bookmarked successfully"

		// Track user interaction
		s.trackUserInteraction(ctx, userID, req.DestinationID, "saved", nil)

	} else if req.Action == "remove" {
		// Remove bookmark
		deleteQuery := `
			DELETE FROM user_bookmarks 
			WHERE user_id = $1 AND destination_id = $2
		`
		_, err = s.db.ExecContext(ctx, deleteQuery, userID, req.DestinationID)
		if err != nil {
			return nil, fmt.Errorf("failed to remove bookmark: %w", err)
		}
		isBookmarked = false
		message = "Bookmark removed successfully"
	}

	return &dto.BookmarkResponse{
		Success:      true,
		IsBookmarked: isBookmarked,
		Message:      message,
	}, nil
}

// IsBookmarked checks if a destination is bookmarked by user
func (s *DestinationService) IsBookmarked(ctx context.Context, userID string, destinationID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_bookmarks WHERE user_id = $1 AND destination_id = $2)
	`, userID, destinationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark: %w", err)
	}
	return exists, nil
}

// GetUserBookmarks retrieves all bookmarked destinations for a user
func (s *DestinationService) GetUserBookmarks(ctx context.Context, userID string, limit, offset int) ([]dto.DestinationResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			d.id, d.name, d.description, d.short_description,
			ST_Y(d.location::geometry) as latitude, 
			ST_X(d.location::geometry) as longitude,
			d.main_image_url, d.images, d.category, d.tags,
			d.price_amount, d.price_currency, d.price_unit,
			d.booking_required, d.booking_url, d.open_hours,
			d.created_at, ub.created_at as bookmarked_at,
			COALESCE(AVG(r.rating), 0) as avg_rating,
			COUNT(r.id) as review_count
		FROM destinations d
		INNER JOIN user_bookmarks ub ON d.id = ub.destination_id
		LEFT JOIN reviews r ON d.id = r.destination_id
		WHERE ub.user_id = $1
		GROUP BY d.id, d.name, d.description, d.short_description, d.location, 
		         d.main_image_url, d.images, d.category, d.tags, d.price_amount,
		         d.price_currency, d.price_unit, d.booking_required, d.booking_url,
		         d.open_hours, d.created_at, ub.created_at
		ORDER BY ub.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var destinations []dto.DestinationResponse
	for rows.Next() {
		var dest dto.DestinationResponse
		var bookmarkedAt time.Time
		var avgRating sql.NullFloat64
		var reviewCount sql.NullInt64
		var images, tags []string
		var priceAmount sql.NullFloat64
		var priceCurrency, priceUnit sql.NullString
		var bookingRequired sql.NullBool
		var bookingURL, openHours sql.NullString

		if err := rows.Scan(
			&dest.ID, &dest.Name, &dest.Description, &dest.ShortDesc,
			&dest.Location.Latitude, &dest.Location.Longitude,
			&dest.MainImage, &images, &dest.Category, &tags,
			&priceAmount, &priceCurrency, &priceUnit,
			&bookingRequired, &bookingURL, &openHours,
			&dest.CreatedAt, &bookmarkedAt,
			&avgRating, &reviewCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}

		// Set calculated fields
		if avgRating.Valid {
			dest.Rating = avgRating.Float64
		}
		if reviewCount.Valid {
			dest.ReviewCount = int(reviewCount.Int64)
		}

		dest.Images = images
		dest.Tags = tags
		dest.IsBookmarked = true
		dest.IsBookmarkable = true

		// Set price info
		if priceAmount.Valid && priceAmount.Float64 > 0 {
			dest.Price = &dto.PriceInfo{
				Amount:   priceAmount.Float64,
				Currency: priceCurrency.String,
				Unit:     priceUnit.String,
			}
		}

		// Set booking info
		if bookingRequired.Valid || bookingURL.Valid || openHours.Valid {
			dest.BookingInfo = &dto.BookingInfo{
				IsAvailable:    true,
				TicketRequired: bookingRequired.Bool,
				BookingURL:     bookingURL.String,
				OpenHours:      openHours.String,
			}
		}

		destinations = append(destinations, dest)
	}

	return destinations, nil
}

// GetCategories retrieves all destination categories
func (s *DestinationService) GetCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	query := `
		SELECT 
			category,
			COUNT(*) as destination_count
		FROM destinations 
		WHERE category IS NOT NULL
		GROUP BY category
		ORDER BY destination_count DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []dto.CategoryResponse
	for rows.Next() {
		var category dto.CategoryResponse
		if err := rows.Scan(&category.Name, &category.Count); err != nil {
			continue
		}
		
		category.ID = strings.ToLower(strings.ReplaceAll(category.Name, " ", "_"))
		category.Icon = s.getCategoryIcon(category.Name)
		categories = append(categories, category)
	}

	return categories, nil
}

// getCategoryIcon returns icon name for category
func (s *DestinationService) getCategoryIcon(category string) string {
	iconMap := map[string]string{
		"Cultural":    "culture",
		"Historical":  "history", 
		"Religious":   "temple",
		"Nature":      "nature",
		"Adventure":   "mountain",
		"Beach":       "beach",
		"City":        "city",
		"Traditional": "heritage",
	}
	
	if icon, exists := iconMap[category]; exists {
		return icon
	}
	return "place"
}

// GetPhotos retrieves photo gallery for a destination
func (s *DestinationService) GetPhotos(ctx context.Context, destinationID string) ([]string, error) {
	query := `
		SELECT images FROM destinations WHERE id = $1
	`

	var images []string
	err := s.db.QueryRowContext(ctx, query, destinationID).Scan(&images)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("destination not found")
		}
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}

	// Also get photos from reviews
	reviewPhotosQuery := `
		SELECT UNNEST(images) as image_url 
		FROM reviews 
		WHERE destination_id = $1 AND images IS NOT NULL
		LIMIT 20
	`

	rows, err := s.db.QueryContext(ctx, reviewPhotosQuery, destinationID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var imageURL string
			if err := rows.Scan(&imageURL); err == nil {
				images = append(images, imageURL)
			}
		}
	}

	return images, nil
}

// trackUserInteraction records user interaction for recommendation engine
func (s *DestinationService) trackUserInteraction(ctx context.Context, userID, destinationID, actionType string, rating *int) {
	query := `
		INSERT INTO user_interactions (user_id, destination_id, action_type, rating, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	
	_, err := s.db.ExecContext(ctx, query, userID, destinationID, actionType, rating)
	if err != nil {
		util.LogError("Failed to track user interaction", err)
	}
}

// Helper function to calculate distance between two coordinates using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
		math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}
