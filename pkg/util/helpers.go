package util

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// APIResponse represents a standardized API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseWithData creates a standardized API response with data payload
func ResponseWithData(c *fiber.Ctx, data interface{}, message string, statusCode int, success bool) error {
	response := APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	return c.Status(statusCode).JSON(response)
}

// ResponseWithMessage creates a standardized API response without data payload
func ResponseWithMessage(c *fiber.Ctx, message string, statusCode int, success bool) error {
	response := APIResponse{
		Success: success,
		Message: message,
	}
	return c.Status(statusCode).JSON(response)
}

// FormatGeminiPrompt creates a detailed prompt for Gemini AI to generate travel itinerary
func FormatGeminiPrompt(userInput *dto.SmartPlanRequest, duration int) string {
	promptTemplate := `You are a premium travel planning AI for Indonesia. Create a detailed travel itinerary.

**CRITICAL: Your response MUST be ONLY valid JSON. No explanations, no markdown, no additional text. Start with { and end with }.**

User Request:
- Destination: %s
- Travel Dates: %s to %s (%d days)
- Budget: IDR %s
- Activity Preferences: %s
- Travel Style: %s
- Activity Intensity: %s

Generate a JSON response with this EXACT structure:
{
  "itinerary": [
    {
      "day": 1,
      "date": "YYYY-MM-DD",
      "theme": "Day theme matching preferences",
      "activities": [
        {
          "time": "HH:MM",
          "activity": "Activity name",
          "location": "Location with address",
          "description": "Why this matches user preferences",
          "duration": "X hours",
          "cost": "IDR amount or Free",
          "category": "nature|culture|culinary|shopping|family",
          "notes": "Practical tips and transport info"
        }
      ],
      "daily_budget": "IDR amount",
      "intensity_level": "%s"
    }
  ],
  "summary": {
    "total_cost_estimate": "IDR range",
    "highlights": ["Top experience 1", "Top experience 2", "Top experience 3"],
    "travel_tips": ["Practical tip 1", "Local insight 2", "Transport tip 3"],
    "travel_style_notes": "How itinerary matches %s style",
    "last_updated": "July 2025"
  }
}

Create a complete %d-day itinerary for %s focusing on %s activities with %s intensity.`

	// Format dates
	startDateStr := userInput.StartDate.Format("2006-01-02")
	endDateStr := userInput.EndDate.Format("2006-01-02")

	// Format budget
	budgetStr := "Not specified"
	if userInput.Budget != nil {
		budgetStr = fmt.Sprintf("%.0f", *userInput.Budget)
	}

	// Format activity preferences
	preferencesStr := "general sightseeing"
	if len(userInput.ActivityPreferences) > 0 {
		preferencesStr = fmt.Sprintf("%v", userInput.ActivityPreferences)
	}

	// Format travel style
	travelStyleStr := "balanced traveler"
	if userInput.TravelStyle != nil {
		travelStyleStr = *userInput.TravelStyle
	}

	// Format activity intensity
	intensityStr := "balanced"
	if userInput.ActivityIntensity != nil {
		intensityStr = *userInput.ActivityIntensity
	}

	return fmt.Sprintf(promptTemplate,
		userInput.Destination, // destination
		startDateStr,          // start date
		endDateStr,            // end date
		duration,              // duration
		budgetStr,             // budget
		preferencesStr,        // activity preferences
		travelStyleStr,        // travel style
		intensityStr,          // activity intensity
		intensityStr,          // intensity level in JSON
		travelStyleStr,        // travel style notes
		duration,              // duration for final
		userInput.Destination, // destination for final
		preferencesStr,        // preferences for final
		intensityStr,          // intensity for final
	)
}

// FormatOptimizedGeminiPrompt creates an optimized prompt for faster Gemini AI response
func FormatOptimizedGeminiPrompt(userInput *dto.SmartPlanRequest, duration int, businesses []dto.LocalBusiness, attractions []dto.TouristAttraction) string {
	promptTemplate := `Create a %d-day Indonesia travel itinerary for %s. JSON only.

Details: %s to %s, Budget: IDR %s, Preferences: %s, Style: %s, Intensity: %s

Format:
{
  "itinerary": [
    {
      "day": 1,
      "date": "YYYY-MM-DD",
      "activities": [
        {
          "time": "HH:MM",
          "activity": "Name",
          "location": "Place",
          "duration": "X hours",
          "cost": "IDR amount",
          "category": "nature|culture|culinary|shopping"
        }
      ],
      "daily_budget": "IDR amount"
    }
  ],
  "summary": {
    "total_cost": "IDR range",
    "highlights": ["Top 3 experiences"],
    "tips": ["Essential tips"]
  }
}`

	// Quick data preparation
	startDate := userInput.StartDate.Format("2006-01-02")
	endDate := userInput.EndDate.Format("2006-01-02")
	budget := "Not specified"
	if userInput.Budget != nil {
		budget = fmt.Sprintf("%.0f", *userInput.Budget)
	}

	preferences := "general"
	if len(userInput.ActivityPreferences) > 0 {
		preferences = fmt.Sprintf("%v", userInput.ActivityPreferences)
	}

	style := "balanced"
	if userInput.TravelStyle != nil {
		style = *userInput.TravelStyle
	}

	intensity := "moderate"
	if userInput.ActivityIntensity != nil {
		intensity = *userInput.ActivityIntensity
	}

	basePrompt := fmt.Sprintf(promptTemplate, duration, userInput.Destination,
		startDate, endDate, budget, preferences, style, intensity)

	// Add top verified places if available (max 3 each for efficiency)
	if len(businesses) > 0 || len(attractions) > 0 {
		verified := "\n\nInclude if relevant:"

		if len(businesses) > 0 {
			verified += " Businesses: "
			limit := 3
			if len(businesses) < limit {
				limit = len(businesses)
			}
			for i, b := range businesses[:limit] {
				if i > 0 {
					verified += ", "
				}
				verified += b.Name
			}
		}

		if len(attractions) > 0 {
			verified += " Attractions: "
			limit := 3
			if len(attractions) < limit {
				limit = len(attractions)
			}
			for i, a := range attractions[:limit] {
				if i > 0 {
					verified += ", "
				}
				verified += a.Name
			}
		}

		basePrompt += verified
	}

	return basePrompt
}
