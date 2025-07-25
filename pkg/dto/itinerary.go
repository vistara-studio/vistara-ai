package dto

import "time"

// TravelItineraryRequest represents request to create or update itinerary
type TravelItineraryRequest struct {
	Title           string    `json:"title" validate:"required,min=3,max=255"`
	Description     string    `json:"description"`
	DestinationName string    `json:"destination_name" validate:"required"`
	StartDate       time.Time `json:"start_date" validate:"required"`
	EndDate         time.Time `json:"end_date" validate:"required"`
	Duration        int       `json:"duration" validate:"required,min=1,max=30"`
	Budget          float64   `json:"budget" validate:"required,min=0"`
	BudgetCurrency  string    `json:"budget_currency"`
	Interests       []string  `json:"interests"`
}

// TravelItineraryResponse represents itinerary response
type TravelItineraryResponse struct {
	ID              string                      `json:"id"`
	UserID          string                      `json:"user_id"`
	Title           string                      `json:"title"`
	Description     string                      `json:"description"`
	DestinationName string                      `json:"destination_name"`
	StartDate       time.Time                   `json:"start_date"`
	EndDate         time.Time                   `json:"end_date"`
	Duration        int                         `json:"duration_days"`
	TotalBudget     float64                     `json:"total_budget"`
	BudgetCurrency  string                      `json:"budget_currency"`
	Interests       []string                    `json:"interests"`
	Status          string                      `json:"status"`
	ItineraryData   *ItineraryData              `json:"itinerary_data,omitempty"`
	Destinations    []ItineraryDestinationItem  `json:"destinations,omitempty"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
}

// ItineraryData represents the AI-generated itinerary plan
type ItineraryData struct {
	Days          []ItineraryDay `json:"days"`
	TotalBudget   float64        `json:"total_budget"`
	Currency      string         `json:"currency"`
	Overview      string         `json:"overview"`
	Tips          []string       `json:"tips"`
	Transportation string        `json:"transportation"`
}

// ItineraryDay represents one day in the itinerary
type ItineraryDay struct {
	Day           int                    `json:"day"`
	Date          time.Time              `json:"date"`
	Theme         string                 `json:"theme"`
	Activities    []ItineraryActivity    `json:"activities"`
	Meals         []ItineraryMeal        `json:"meals"`
	DailyBudget   float64                `json:"daily_budget"`
	Transportation *TransportationInfo   `json:"transportation,omitempty"`
}

// ItineraryActivity represents an activity in the itinerary
type ItineraryActivity struct {
	Time         string  `json:"time"`
	Duration     int     `json:"duration_minutes"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Location     string  `json:"location"`
	Cost         float64 `json:"cost"`
	DestinationID *string `json:"destination_id,omitempty"` // Link to destinations table
	Tips         string  `json:"tips"`
}

// ItineraryMeal represents a meal recommendation
type ItineraryMeal struct {
	Type        string  `json:"type"` // breakfast, lunch, dinner, snack
	Time        string  `json:"time"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	Cuisine     string  `json:"cuisine"`
}

// TransportationInfo represents transportation details
type TransportationInfo struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	Duration    string  `json:"duration"`
	Tips        string  `json:"tips"`
}

// ItineraryDestinationItem represents destination added to itinerary
type ItineraryDestinationItem struct {
	ID            string    `json:"id"`
	ItineraryID   string    `json:"itinerary_id"`
	DestinationID *string   `json:"destination_id"`
	DayNumber     int       `json:"day_number"`
	VisitOrder    int       `json:"visit_order"`
	PlannedTime   string    `json:"planned_time"`
	Duration      int       `json:"duration_minutes"`
	Notes         string    `json:"notes"`
	IsCompleted   bool      `json:"is_completed"`
	CreatedAt     time.Time `json:"created_at"`
	
	// Destination details (if linked)
	Destination   *DestinationResponse `json:"destination,omitempty"`
}

// AddDestinationToItineraryRequest represents request to add destination to itinerary
type AddDestinationToItineraryRequest struct {
	DestinationID string `json:"destination_id" validate:"required"`
	DayNumber     int    `json:"day_number" validate:"required,min=1"`
	VisitOrder    int    `json:"visit_order" validate:"required,min=1"`
	PlannedTime   string `json:"planned_time"`
	Duration      int    `json:"duration_minutes" validate:"min=15"`
	Notes         string `json:"notes"`
}

// ItineraryHistoryResponse represents user's itinerary history
type ItineraryHistoryResponse struct {
	Itineraries []TravelItineraryResponse `json:"itineraries"`
	Pagination  PaginationInfo            `json:"pagination"`
	Summary     ItineraryHistorySummary   `json:"summary"`
}

// ItineraryHistorySummary provides summary statistics
type ItineraryHistorySummary struct {
	TotalItineraries     int     `json:"total_itineraries"`
	CompletedItineraries int     `json:"completed_itineraries"`
	TotalDestinations    int     `json:"total_destinations"`
	TotalBudgetSpent     float64 `json:"total_budget_spent"`
	Currency             string  `json:"currency"`
	MostVisitedCategory  string  `json:"most_visited_category"`
}

// PaginationInfo represents pagination details
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}
