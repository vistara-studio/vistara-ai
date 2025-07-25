package dto

import "time"

// NusalingoRequest represents a request for language translation
type NusalingoRequest struct {
	FromLanguage string `json:"from_language" validate:"required,min=2,max=50"`
	ToLanguage   string `json:"to_language" validate:"required,min=2,max=50"`
	Text         string `json:"text" validate:"required,min=1,max=5000"`
}

// NusalingoResponse represents the response from language translation
type NusalingoResponse struct {
	TranslatedText string    `json:"translated_text"`
	FromLanguage   string    `json:"from_language"`
	ToLanguage     string    `json:"to_language"`
	OriginalText   string    `json:"original_text"`
	GeneratedAt    time.Time `json:"generated_at"`
}
