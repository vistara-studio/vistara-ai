package service

import (
	"context"
	"log"
	"time"

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
	return g.GenerateTextWithTimeout(promptText, g.config.RequestTimeout)
}

// GenerateTextWithGrounding generates text using Gemini AI with optional search grounding
func (g *GeminiService) GenerateTextWithGrounding(promptText string, useGrounding bool) (string, error) {
	return g.GenerateTextWithTimeoutAndGrounding(promptText, g.config.RequestTimeout, useGrounding)
}

// GenerateTextWithTimeout generates text using Gemini AI with custom timeout
func (g *GeminiService) GenerateTextWithTimeout(promptText string, timeout time.Duration) (string, error) {
	return g.GenerateTextWithTimeoutAndGrounding(promptText, timeout, false)
}

// GenerateTextWithTimeoutAndGrounding generates text using Gemini AI with custom timeout and optional grounding
func (g *GeminiService) GenerateTextWithTimeoutAndGrounding(promptText string, timeout time.Duration, useGrounding bool) (string, error) {
	return g.GenerateTextWithAdvancedSettings(promptText, timeout, useGrounding, true)
}

// GenerateTextWithAdvancedSettings generates text using Gemini AI with configurable settings
func (g *GeminiService) GenerateTextWithAdvancedSettings(promptText string, timeout time.Duration, useGrounding bool, jsonOutput bool) (string, error) {
	// Create context with custom timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get model name from config
	modelName := g.config.GeminiModelName
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	var logMessage string
	if useGrounding && g.config.EnableSearchGrounding {
		logMessage = "with search grounding"
	} else {
		logMessage = "without grounding"
	}
	log.Printf("Using Gemini model: %s %s (timeout: %v)", modelName, logMessage, timeout)

	model := g.client.GenerativeModel(modelName)

	// Optimized generation settings for gemini-2.5-flash
	temperature := float32(0.2) // Balanced for quality and speed
	topP := float32(0.9)        // High for diverse responses
	topK := int32(30)           // Moderate for good performance
	maxTokens := g.config.MaxTokens

	model.GenerationConfig = genai.GenerationConfig{
		Temperature:     &temperature,
		TopP:            &topP,
		TopK:            &topK,
		MaxOutputTokens: &maxTokens,
	}

	// Optimized system instructions
	var systemInstruction string
	if jsonOutput {
		systemInstruction = "You are an expert AI assistant. Respond ONLY with valid JSON. No explanations, no markdown, just clean JSON starting with { and ending with }."
	} else {
		systemInstruction = "You are an expert AI assistant. Provide direct, concise responses. Return only what is requested."
	}

	// Add system instruction
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(systemInstruction),
		},
	}

	// Generate content with timeout
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
