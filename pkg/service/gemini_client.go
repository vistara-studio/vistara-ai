package service

import (
	"context"
	"log"

	"github.com/google/generative-ai-go/genai"
	"github.com/vistara-studio/vistara-ai/infra/config"
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
	return g.GenerateTextWithGrounding(promptText, false)
}

// GenerateTextWithGrounding generates text using Gemini AI with optional search grounding
func (g *GeminiService) GenerateTextWithGrounding(promptText string, useGrounding bool) (string, error) {
	return g.GenerateTextWithSettings(promptText, useGrounding, true)
}

// GenerateTextWithSettings generates text using Gemini AI with configurable settings
func (g *GeminiService) GenerateTextWithSettings(promptText string, useGrounding bool, jsonOutput bool) (string, error) {
	ctx := context.Background()

	// Get model name from config
	modelName := g.config.GeminiModelName
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	var logMessage string
	if useGrounding && g.config.EnableSearchGrounding {
		logMessage = "with search grounding"
	} else {
		logMessage = "without grounding"
	}
	log.Printf("Using Gemini model: %s %s", modelName, logMessage)
	
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

	// Enhanced system instruction for better output format and data accuracy
	var systemInstruction string
	if jsonOutput {
		if useGrounding && g.config.EnableSearchGrounding {
			systemInstruction = "You are an AI assistant that MUST respond ONLY with valid JSON. Never include any text outside the JSON structure. Start with { and end with }. Use your most comprehensive and up-to-date knowledge. Prioritize the most accurate, verified, and recent information from reliable academic sources, official records, and scholarly publications. When providing historical information, cite authentic manuscripts, archaeological findings, and peer-reviewed research."
		} else {
			systemInstruction = "You are an AI assistant that MUST respond ONLY with valid JSON. Never include any text outside the JSON structure. Start with { and end with }."
		}
	} else {
		if useGrounding && g.config.EnableSearchGrounding {
			systemInstruction = "You are a precise AI assistant. Follow the user's instructions exactly. Use your most comprehensive and up-to-date knowledge. Prioritize accuracy and authenticity. Return only what is requested, no additional text or explanations."
		} else {
			systemInstruction = "You are a precise AI assistant. Follow the user's instructions exactly. Return only what is requested, no additional text or explanations."
		}
	}

	// Add system instruction
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(systemInstruction),
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
