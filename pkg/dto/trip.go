package dto

import "time"

// SmartPlanRequest represents the request payload for smart trip planning
type SmartPlanRequest struct {
	Destination         string    `json:"destination" validate:"required,min=2"`
	StartDate           time.Time `json:"start_date" validate:"required"`
	EndDate             time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	Budget              *float64  `json:"budget" validate:"omitempty,gt=0"`
	ActivityPreferences []string  `json:"activity_preferences" validate:"omitempty"`
	TravelStyle         *string   `json:"travel_style" validate:"omitempty,oneof=solo_traveler romantic_couple family_with_children backpacker luxury_traveler"`
	ActivityIntensity   *string   `json:"activity_intensity" validate:"omitempty,oneof=relaxed balanced full"`

	// Fields for integration with vistara-be
	UserID           *string  `json:"user_id,omitempty"`
	LocalBusinessIDs []string `json:"local_business_ids,omitempty"`
	AttractionIDs    []string `json:"attraction_ids,omitempty"`
}

// SmartPlanResponse represents the response payload for smart trip planning
type SmartPlanResponse struct {
	Plan              interface{} `json:"plan"`
	Destination       string      `json:"destination"`
	StartDate         time.Time   `json:"start_date"`
	EndDate           time.Time   `json:"end_date"`
	Budget            *float64    `json:"budget"`
	TravelStyle       *string     `json:"travel_style"`
	ActivityIntensity *string     `json:"activity_intensity"`
	GeneratedAt       time.Time   `json:"generated_at"`

	// Integration fields
	UserID                 *string             `json:"user_id,omitempty"`
	RecommendedBusinesses  []LocalBusiness     `json:"recommended_businesses,omitempty"`
	RecommendedAttractions []TouristAttraction `json:"recommended_attractions,omitempty"`
}

// LocalBusiness represents a local business for integration
type LocalBusiness struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Address     string  `json:"address"`
	Rating      float64 `json:"rating,omitempty"`
	PriceRange  string  `json:"price_range,omitempty"`
}

// TouristAttraction represents a tourist attraction for integration
type TouristAttraction struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Address     string  `json:"address"`
	Rating      float64 `json:"rating,omitempty"`
	Price       float64 `json:"price,omitempty"`
}

// TripPlanRequest represents the request payload for trip planning (legacy)
type TripPlanRequest struct {
	Destination   string    `json:"destination" validate:"required,min=2"`
	StartDate     time.Time `json:"start_date" validate:"required"`
	EndDate       time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	Budget        *float64  `json:"budget" validate:"omitempty,gt=0"`
	TravelStyle   *string   `json:"travel_style" validate:"omitempty,oneof=budget comfort luxury"`
	Interests     []string  `json:"interests" validate:"omitempty"`
	GroupSize     *int      `json:"group_size" validate:"omitempty,gt=0"`
	Accommodation *string   `json:"accommodation" validate:"omitempty"`
}

// TripPlanResponse represents the response payload for trip planning (legacy)
type TripPlanResponse struct {
	ID                 string      `json:"id"`
	Destination        string      `json:"destination"`
	GeneratedItinerary interface{} `json:"generated_itinerary"`
	CreatedAt          time.Time   `json:"created_at"`
}
