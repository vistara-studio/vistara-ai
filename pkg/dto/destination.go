package dto

import (
	"time"
)

// DestinationRequest represents the request for destination search
type DestinationRequest struct {
	Lat       *float64 `json:"lat,omitempty"`
	Long      *float64 `json:"long,omitempty"`
	Radius    *float64 `json:"radius,omitempty"` // in kilometers
	Search    *string  `json:"search,omitempty"`
	Category  *string  `json:"category,omitempty"` // All, Cultural, Historical, etc.
	Limit     *int     `json:"limit,omitempty"`
	Offset    *int     `json:"offset,omitempty"`
	SortBy    *string  `json:"sort_by,omitempty"` // distance, rating, name, popularity
}

// DestinationResponse represents a single destination
type DestinationResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	ShortDesc      string              `json:"short_description"` // For card display
	Location       LocationResponse    `json:"location"`
	MainImage      string              `json:"main_image_url"`
	Images         []string            `json:"images"` // Photo gallery
	Category       string              `json:"category"` // Cultural, Historical, Religious, etc.
	Tags           []string            `json:"tags"`
	Rating         float64             `json:"rating"`
	ReviewCount    int                 `json:"review_count"`
	Price          *PriceInfo          `json:"price,omitempty"`
	Manuscript     *ManuscriptResponse `json:"manuscript,omitempty"`
	Reviews        []ReviewResponse    `json:"reviews,omitempty"`
	Distance       *float64            `json:"distance,omitempty"` // in kilometers
	BookingInfo    *BookingInfo        `json:"booking_info,omitempty"`
	IsBookmarkable bool                `json:"is_bookmarkable"`
	IsBookmarked   bool                `json:"is_bookmarked"`
	CreatedAt      time.Time           `json:"created_at"`
}

// PriceInfo represents pricing information
type PriceInfo struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Unit     string  `json:"unit"` // per person, per group, etc.
}

// BookingInfo represents booking availability
type BookingInfo struct {
	IsAvailable    bool      `json:"is_available"`
	TicketRequired bool      `json:"ticket_required"`
	BookingURL     string    `json:"booking_url,omitempty"`
	OpenHours      string    `json:"open_hours,omitempty"`
	LastEntry      string    `json:"last_entry,omitempty"`
	AvailableDates []string  `json:"available_dates,omitempty"`
}

// LocationResponse represents geographic coordinates
type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ManuscriptResponse represents manuscript details
type ManuscriptResponse struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	CulturalStory       string    `json:"cultural_story"`
	OriginalScriptText  string    `json:"original_script_text"`
	TranslationText     string    `json:"translation_text"`
	ManuscriptImageURL  string    `json:"manuscript_image_url"`
	AudioNarrationURL   *string   `json:"audio_narration_url,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// ReviewResponse represents user review
type ReviewResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserAvatar  string    `json:"user_avatar,omitempty"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
	Images      []string  `json:"images,omitempty"` // Review photos
	IsVerified  bool      `json:"is_verified"` // Verified visitor
	TimeAgo     string    `json:"time_ago"` // "4 months ago"
	CreatedAt   time.Time `json:"created_at"`
}

// NarrationRequest represents request for generating audio narration
type NarrationRequest struct {
	Language    string `json:"language" validate:"required,oneof=id-ID en-US"`
	VoiceGender string `json:"voice_gender" validate:"required,oneof=male female"`
	VoiceStyle  string `json:"voice_style" validate:"omitempty,oneof=dramatic calm friendly"`
}

// NarrationResponse represents response for audio narration
type NarrationResponse struct {
	AudioURL  string    `json:"audio_url"`
	Duration  float64   `json:"duration"` // in seconds
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

// ItineraryRequest represents request to create itinerary
type ItineraryRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

// ItineraryResponse represents itinerary details
type ItineraryResponse struct {
	ID           string                      `json:"id"`
	UserID       string                      `json:"user_id"`
	Name         string                      `json:"name"`
	Destinations []ItineraryDestinationResponse `json:"destinations"`
	CreatedAt    time.Time                   `json:"created_at"`
}

// ItineraryDestinationRequest represents request to add destination to itinerary
type ItineraryDestinationRequest struct {
	DestinationID string `json:"destination_id" validate:"required,uuid"`
	VisitDate     string `json:"visit_date" validate:"required"`
	Notes         string `json:"notes,omitempty"`
}

// ItineraryDestinationResponse represents destination in itinerary
type ItineraryDestinationResponse struct {
	Destination DestinationResponse `json:"destination"`
	VisitDate   time.Time           `json:"visit_date"`
	Notes       string              `json:"notes"`
}

// RecommendationRequest represents request for destination recommendations
type RecommendationRequest struct {
	UserID string   `json:"user_id" validate:"required"`
	Tags   []string `json:"tags,omitempty"`
	Limit  int      `json:"limit" validate:"min=1,max=20"`
}

// RecommendationResponse represents recommended destinations
type RecommendationResponse struct {
	Destinations []DestinationResponse `json:"destinations"`
	Reason       string                `json:"reason"`
	Score        float64               `json:"score"`
}

// MapViewRequest represents request for map view data
type MapViewRequest struct {
	Lat    float64 `json:"lat" validate:"required"`
	Long   float64 `json:"long" validate:"required"`
	Zoom   int     `json:"zoom" validate:"min=1,max=20"`
	Bounds MapBounds `json:"bounds"`
}

// MapBounds represents map viewport bounds
type MapBounds struct {
	NorthEast LocationResponse `json:"north_east"`
	SouthWest LocationResponse `json:"south_west"`
}

// MapViewResponse represents destinations for map display
type MapViewResponse struct {
	Destinations []MapDestination `json:"destinations"`
	Center       LocationResponse `json:"center"`
	Zoom         int              `json:"zoom"`
}

// MapDestination represents simplified destination for map pins
type MapDestination struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Location  LocationResponse `json:"location"`
	Category  string           `json:"category"`
	Rating    float64          `json:"rating"`
	MainImage string           `json:"main_image_url"`
	Price     *PriceInfo       `json:"price,omitempty"`
}

// BookmarkRequest represents request to bookmark/unbookmark destination
type BookmarkRequest struct {
	DestinationID string `json:"destination_id" validate:"required,uuid"`
	Action        string `json:"action" validate:"required,oneof=add remove"`
}

// BookmarkResponse represents bookmark operation result
type BookmarkResponse struct {
	Success       bool   `json:"success"`
	IsBookmarked  bool   `json:"is_bookmarked"`
	Message       string `json:"message"`
}

// CategoryResponse represents destination categories
type CategoryResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Count int    `json:"count"`
}

// FilterRequest represents advanced filtering options
type FilterRequest struct {
	Categories   []string  `json:"categories,omitempty"`
	MinRating    *float64  `json:"min_rating,omitempty"`
	MaxDistance  *float64  `json:"max_distance,omitempty"`
	PriceRange   *PriceRange `json:"price_range,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	HasTicket    *bool     `json:"has_ticket,omitempty"`
	IsBookmarked *bool     `json:"is_bookmarked,omitempty"`
}

// PriceRange represents price filtering range
type PriceRange struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Currency string  `json:"currency"`
}
