package dto

import "time"

// HistoricalStoryRequest represents a request for historical story generation
type HistoricalStoryRequest struct {
	Location string `json:"location" validate:"required,min=2,max=100"`
}

// HistoricalStoryResponse represents the response from historical story generation
type HistoricalStoryResponse struct {
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Location    string    `json:"location"`
	GeneratedAt time.Time `json:"generated_at"`
}
