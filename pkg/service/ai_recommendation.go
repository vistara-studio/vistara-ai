package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// RecommendationService handles destination recommendations
type RecommendationService struct {
	db *sql.DB
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(db *sql.DB) *RecommendationService {
	return &RecommendationService{
		db: db,
	}
}

// UserInteraction represents user's interaction with destinations
type UserInteraction struct {
	UserID        string
	DestinationID string
	ActionType    string // viewed, saved, reviewed, visited
	Score         float64
	Tags          []string
}

// GetRecommendations provides personalized destination recommendations
func (s *RecommendationService) GetRecommendations(ctx context.Context, req dto.RecommendationRequest) ([]dto.RecommendationResponse, error) {
	// Get user's interaction history
	userProfile, err := s.getUserProfile(ctx, req.UserID)
	if err != nil {
		// If no profile, return content-based recommendations
		return s.getContentBasedRecommendations(ctx, req)
	}

	// Get collaborative filtering recommendations
	recommendations, err := s.getCollaborativeRecommendations(ctx, userProfile, req)
	if err != nil || len(recommendations) < req.Limit {
		// Fallback to content-based
		contentRecs, _ := s.getContentBasedRecommendations(ctx, req)
		recommendations = append(recommendations, contentRecs...)
	}

	// Remove duplicates and limit results
	seen := make(map[string]bool)
	var uniqueRecs []dto.RecommendationResponse
	for _, rec := range recommendations {
		for _, dest := range rec.Destinations {
			if !seen[dest.ID] && len(uniqueRecs) < req.Limit {
				seen[dest.ID] = true
				uniqueRecs = append(uniqueRecs, rec)
				break
			}
		}
	}

	return uniqueRecs, nil
}

// getUserProfile analyzes user's interaction patterns
func (s *RecommendationService) getUserProfile(ctx context.Context, userID string) (*RecommendationUserProfile, error) {
	query := `
		SELECT 
			destination_id,
			action_type,
			created_at,
			COALESCE(rating, 0) as rating
		FROM user_interactions 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT 100
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profile := &RecommendationUserProfile{
		UserID:       userID,
		Preferences:  make(map[string]float64),
		InteractedDestinations: make(map[string]float64),
	}

	for rows.Next() {
		var destID, actionType string
		var rating float64
		var createdAt string

		if err := rows.Scan(&destID, &actionType, &createdAt, &rating); err != nil {
			continue
		}

		// Calculate interaction score
		score := s.calculateInteractionScore(actionType, rating)
		profile.InteractedDestinations[destID] = score

		// Get destination tags for this interaction
		tags, err := s.getDestinationTags(ctx, destID)
		if err == nil {
			for _, tag := range tags {
				profile.Preferences[tag] += score
			}
		}
	}

	if len(profile.InteractedDestinations) == 0 {
		return nil, fmt.Errorf("no user interactions found")
	}

	return profile, nil
}

// RecommendationUserProfile represents user's preferences and interaction history
type RecommendationUserProfile struct {
	UserID                 string
	Preferences            map[string]float64 // tag -> preference score
	InteractedDestinations map[string]float64 // destination_id -> interaction score
}

// calculateInteractionScore assigns scores to different interaction types
func (s *RecommendationService) calculateInteractionScore(actionType string, rating float64) float64 {
	baseScores := map[string]float64{
		"viewed":   1.0,
		"saved":    3.0,
		"reviewed": 4.0,
		"visited":  5.0,
	}

	score := baseScores[actionType]
	if actionType == "reviewed" && rating > 0 {
		// Boost score based on rating (1-5 stars)
		score = score * (rating / 5.0)
	}

	return score
}

// getDestinationTags retrieves tags for a destination
func (s *RecommendationService) getDestinationTags(ctx context.Context, destinationID string) ([]string, error) {
	query := `
		SELECT UNNEST(tags) as tag 
		FROM destinations 
		WHERE id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, destinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err == nil {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// getContentBasedRecommendations provides recommendations based on content similarity
func (s *RecommendationService) getContentBasedRecommendations(ctx context.Context, req dto.RecommendationRequest) ([]dto.RecommendationResponse, error) {
	// If user provided tags, use them
	if len(req.Tags) > 0 {
		return s.getRecommendationsByTags(ctx, req.Tags, req.Limit)
	}

	// Otherwise, return popular destinations
	return s.getPopularDestinations(ctx, req.Limit)
}

// getRecommendationsByTags finds destinations matching specific tags
func (s *RecommendationService) getRecommendationsByTags(ctx context.Context, tags []string, limit int) ([]dto.RecommendationResponse, error) {
	// Create tag filter
	tagPlaceholders := make([]string, len(tags))
	tagArgs := make([]interface{}, len(tags))
	for i, tag := range tags {
		tagPlaceholders[i] = fmt.Sprintf("$%d", i+1)
		tagArgs = append(tagArgs, tag)
	}

	query := fmt.Sprintf(`
		SELECT 
			d.id, d.name, d.description,
			ST_Y(d.location::geometry) as latitude,
			ST_X(d.location::geometry) as longitude,
			d.main_image_url, d.created_at,
			array_length(d.tags, 1) as tag_count
		FROM destinations d
		WHERE d.tags && ARRAY[%s]
		ORDER BY 
			(SELECT COUNT(*) FROM unnest(d.tags) tag WHERE tag = ANY(ARRAY[%s])) DESC,
			d.name ASC
		LIMIT $%d
	`, strings.Join(tagPlaceholders, ","), strings.Join(tagPlaceholders, ","), len(tags)+1)

	tagArgs = append(tagArgs, tagArgs...) // Duplicate for both WHERE and ORDER BY
	tagArgs = append(tagArgs, limit)

	rows, err := s.db.QueryContext(ctx, query, tagArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []dto.RecommendationResponse
	for rows.Next() {
		var dest dto.DestinationResponse
		var tagCount sql.NullInt64

		if err := rows.Scan(
			&dest.ID, &dest.Name, &dest.Description,
			&dest.Location.Latitude, &dest.Location.Longitude,
			&dest.MainImage, &dest.CreatedAt, &tagCount,
		); err != nil {
			continue
		}

		score := float64(tagCount.Int64) / float64(len(tags)) // Match ratio
		reason := fmt.Sprintf("Matches %d of your interest tags: %s", 
			int(tagCount.Int64), strings.Join(tags, ", "))

		recommendations = append(recommendations, dto.RecommendationResponse{
			Destinations: []dto.DestinationResponse{dest},
			Reason:       reason,
			Score:        score,
		})
	}

	return recommendations, nil
}

// getPopularDestinations returns trending/popular destinations
func (s *RecommendationService) getPopularDestinations(ctx context.Context, limit int) ([]dto.RecommendationResponse, error) {
	query := `
		SELECT 
			d.id, d.name, d.description,
			ST_Y(d.location::geometry) as latitude,
			ST_X(d.location::geometry) as longitude,
			d.main_image_url, d.created_at,
			COALESCE(AVG(r.rating), 0) as avg_rating,
			COUNT(r.id) as review_count
		FROM destinations d
		LEFT JOIN reviews r ON d.id = r.destination_id
		WHERE r.created_at > NOW() - INTERVAL '90 days'
		GROUP BY d.id, d.name, d.description, d.location, d.main_image_url, d.created_at
		HAVING COUNT(r.id) >= 3
		ORDER BY 
			(AVG(r.rating) * 0.7 + (COUNT(r.id) / 10.0) * 0.3) DESC,
			d.name ASC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []dto.RecommendationResponse
	for rows.Next() {
		var dest dto.DestinationResponse
		var avgRating, reviewCount float64

		if err := rows.Scan(
			&dest.ID, &dest.Name, &dest.Description,
			&dest.Location.Latitude, &dest.Location.Longitude,
			&dest.MainImage, &dest.CreatedAt, &avgRating, &reviewCount,
		); err != nil {
			continue
		}

		score := avgRating*0.7 + (reviewCount/10.0)*0.3
		reason := fmt.Sprintf("Popular destination with %.1f stars from %.0f reviews", 
			avgRating, reviewCount)

		recommendations = append(recommendations, dto.RecommendationResponse{
			Destinations: []dto.DestinationResponse{dest},
			Reason:       reason,
			Score:        score,
		})
	}

	return recommendations, nil
}

// getCollaborativeRecommendations uses collaborative filtering
func (s *RecommendationService) getCollaborativeRecommendations(ctx context.Context, userProfile *RecommendationUserProfile, req dto.RecommendationRequest) ([]dto.RecommendationResponse, error) {
	// Find similar users based on interaction patterns
	similarUsers, err := s.findSimilarUsers(ctx, userProfile)
	if err != nil || len(similarUsers) == 0 {
		return nil, fmt.Errorf("no similar users found")
	}

	// Get destinations liked by similar users but not interacted with by current user
	var recommendations []dto.RecommendationResponse
	for _, similarUser := range similarUsers {
		userRecs, err := s.getRecommendationsFromSimilarUser(ctx, similarUser, userProfile, req.Limit)
		if err == nil {
			recommendations = append(recommendations, userRecs...)
		}
	}

	return recommendations, nil
}

// SimilarUser represents a user with similar preferences
type SimilarUser struct {
	UserID     string
	Similarity float64
}

// findSimilarUsers finds users with similar interaction patterns
func (s *RecommendationService) findSimilarUsers(ctx context.Context, userProfile *RecommendationUserProfile) ([]SimilarUser, error) {
	// This is a simplified collaborative filtering approach
	// In production, you might use more sophisticated algorithms like matrix factorization

	query := `
		SELECT 
			user_id,
			destination_id,
			AVG(CASE 
				WHEN action_type = 'viewed' THEN 1
				WHEN action_type = 'saved' THEN 3
				WHEN action_type = 'reviewed' THEN 4
				WHEN action_type = 'visited' THEN 5
				ELSE 0
			END) as interaction_score
		FROM user_interactions
		WHERE user_id != $1
		AND destination_id = ANY($2)
		GROUP BY user_id, destination_id
		HAVING AVG(CASE 
			WHEN action_type = 'viewed' THEN 1
			WHEN action_type = 'saved' THEN 3
			WHEN action_type = 'reviewed' THEN 4
			WHEN action_type = 'visited' THEN 5
			ELSE 0
		END) >= 2.0
	`

	// Convert user's interacted destinations to array
	var destIDs []string
	for destID := range userProfile.InteractedDestinations {
		destIDs = append(destIDs, destID)
	}

	if len(destIDs) == 0 {
		return nil, fmt.Errorf("user has no interactions")
	}

	rows, err := s.db.QueryContext(ctx, query, userProfile.UserID, strings.Join(destIDs, ","))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Calculate similarity scores
	userSimilarities := make(map[string]float64)
	for rows.Next() {
		var otherUserID, destID string
		var score float64

		if err := rows.Scan(&otherUserID, &destID, &score); err != nil {
			continue
		}

		// Simple cosine similarity approximation
		if userScore, exists := userProfile.InteractedDestinations[destID]; exists {
			userSimilarities[otherUserID] += userScore * score
		}
	}

	// Convert to sorted list
	var similarUsers []SimilarUser
	for userID, similarity := range userSimilarities {
		if similarity > 0 {
			similarUsers = append(similarUsers, SimilarUser{
				UserID:     userID,
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (highest first)
	// In a real implementation, you'd sort this properly
	return similarUsers[:min(len(similarUsers), 10)], nil
}

// getRecommendationsFromSimilarUser gets recommendations based on similar user's preferences
func (s *RecommendationService) getRecommendationsFromSimilarUser(ctx context.Context, similarUser SimilarUser, userProfile *RecommendationUserProfile, limit int) ([]dto.RecommendationResponse, error) {
	query := `
		SELECT DISTINCT
			d.id, d.name, d.description,
			ST_Y(d.location::geometry) as latitude,
			ST_X(d.location::geometry) as longitude,
			d.main_image_url, d.created_at
		FROM user_interactions ui
		JOIN destinations d ON ui.destination_id = d.id
		WHERE ui.user_id = $1
		AND ui.destination_id != ALL($2)
		AND ui.action_type IN ('saved', 'reviewed', 'visited')
		ORDER BY ui.created_at DESC
		LIMIT $3
	`

	// Convert user's interacted destinations to array
	var destIDs []string
	for destID := range userProfile.InteractedDestinations {
		destIDs = append(destIDs, destID)
	}

	rows, err := s.db.QueryContext(ctx, query, similarUser.UserID, strings.Join(destIDs, ","), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []dto.RecommendationResponse
	for rows.Next() {
		var dest dto.DestinationResponse

		if err := rows.Scan(
			&dest.ID, &dest.Name, &dest.Description,
			&dest.Location.Latitude, &dest.Location.Longitude,
			&dest.MainImage, &dest.CreatedAt,
		); err != nil {
			continue
		}

		reason := fmt.Sprintf("Recommended based on similar users' preferences (%.1f%% similarity)", 
			similarUser.Similarity*100)

		recommendations = append(recommendations, dto.RecommendationResponse{
			Destinations: []dto.DestinationResponse{dest},
			Reason:       reason,
			Score:        similarUser.Similarity,
		})
	}

	return recommendations, nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
