package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"google.golang.org/api/option"
)

// GeminiService handles interactions with Google's Gemini AI
type GeminiService struct {
	client *genai.Client
	config *config.Config
}

// NewGeminiService creates a new Gemini service instance
func NewGeminiService(cfg *config.Config) *GeminiService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		log.Fatal("Failed to create Gemini client:", err)
	}

	log.Println("Gemini client configured successfully")
	return &GeminiService{
		client: client,
		config: cfg,
	}
}

// GenerateText generates text using Gemini AI
func (g *GeminiService) GenerateText(promptText string) (string, error) {
	ctx := context.Background()

	// Get model name from config
	modelName := g.config.GeminiModelName
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	log.Printf("Using Gemini model: %s for travel planning", modelName)
	model := g.client.GenerativeModel(modelName)

	// Configure generation settings for consistent JSON output
	temperature := float32(0.1)
	topP := float32(0.8)
	topK := int32(40)
	maxTokens := int32(8192)

	model.GenerationConfig = genai.GenerationConfig{
		Temperature:     &temperature, // Lower temperature for more consistent responses
		TopP:            &topP,
		TopK:            &topK,
		MaxOutputTokens: &maxTokens,
	}

	// Add system instruction for better JSON compliance
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("You are a travel planning AI that MUST respond ONLY with valid JSON. Never include any text outside the JSON structure. Start with { and end with }."),
		},
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		log.Printf("Error during Gemini API call: %v", err)
		return "", err
	}

	log.Println("Received response from Gemini")

	// Extract text content with proper error handling
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		var result string
		for _, part := range resp.Candidates[0].Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				result += string(textPart)
			}
		}
		if result != "" {
			log.Printf("Successfully extracted text content of length: %d", len(result))
			return result, nil
		}
	}

	log.Printf("Unexpected Gemini response structure: %+v", resp)
	return "", &GeminiError{Message: "Could not extract text content from Gemini response"}
}

// GenerateSmartPlan generates a smart travel plan using Gemini AI
func (g *GeminiService) GenerateSmartPlan(req *dto.SmartPlanRequest) (*dto.SmartPlanResponse, error) {
	// Create prompt for smart planning
	prompt := fmt.Sprintf(`Generate a comprehensive travel plan for %s from %s to %s. 
Budget: %v, Travel Style: %v, Activity Intensity: %v
Activity Preferences: %v

Respond with a detailed JSON itinerary including daily activities, timings, estimated costs, and recommendations.`, 
		req.Destination, 
		req.StartDate.Format("2006-01-02"), 
		req.EndDate.Format("2006-01-02"),
		req.Budget,
		req.TravelStyle,
		req.ActivityIntensity,
		req.ActivityPreferences)

	// Generate text using existing method
	responseText, err := g.GenerateText(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate smart plan: %w", err)
	}

	// Return response
	return &dto.SmartPlanResponse{
		Plan:              responseText,
		Destination:       req.Destination,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		Budget:            req.Budget,
		TravelStyle:       req.TravelStyle,
		ActivityIntensity: req.ActivityIntensity,
		GeneratedAt:       time.Now(),
		UserID:            req.UserID,
	}, nil
}

// Close closes the Gemini service client
func (g *GeminiService) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

// GeminiError represents a custom error for Gemini service
type GeminiError struct {
	Message string
}

func (e *GeminiError) Error() string {
	return e.Message
}
