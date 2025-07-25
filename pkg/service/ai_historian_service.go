package service

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// AIHistorianService handles AI-powered historical storytelling
type AIHistorianService struct {
	geminiService *GeminiService
	config        *config.Config
}

// NewAIHistorianService creates a new AI historian service instance
func NewAIHistorianService(geminiService *GeminiService, cfg *config.Config) *AIHistorianService {
	return &AIHistorianService{
		geminiService: geminiService,
		config:        cfg,
	}
}

// GenerateHistoricalStory generates a historical story with a quote for a given location
func (s *AIHistorianService) GenerateHistoricalStory(location string) (*dto.HistoricalStoryResponse, error) {
	prompt := s.buildHistorianPrompt(location)

	// Call the Gemini service to generate the historical story with enhanced prompting for accuracy and custom timeout
	rawResponse, err := s.geminiService.GenerateTextWithTimeoutAndGrounding(prompt, s.config.HistorianTimeout, true)
	if err != nil {
		log.Printf("AI Service error during historical story generation: %v", err)
		return nil, fmt.Errorf("failed to generate historical story: %w", err)
	}

	// Clean the response from markdown code blocks
	cleanedResponse := s.cleanAIResponse(rawResponse)

	// Parse the JSON response
	var response dto.HistoricalStoryResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &response); err != nil {
		log.Printf("Failed to parse AI historian response as JSON: %v", err)
		log.Printf("Raw response: %s", rawResponse)
		log.Printf("Cleaned response: %s", cleanedResponse)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate response
	if response.Title == "" || response.Content == "" {
		return nil, fmt.Errorf("invalid response: missing title or content")
	}

	return &response, nil
}

// buildHistorianPrompt creates an optimized prompt for the AI historian
func (s *AIHistorianService) buildHistorianPrompt(location string) string {
	return fmt.Sprintf(`Generate a historical story for %s with authentic Indonesian historical context. RESPOND ENTIRELY IN ENGLISH.

Requirements:
1. Find a verified historical manuscript, text, or archaeological source about %s
2. Extract a genuine quote from this source as the title (if original is not in English, provide English translation)
3. Write a compelling story explaining the quote's context and significance IN ENGLISH
4. Use only credible historical sources and archaeological findings
5. ALL content must be written in English language

Output ONLY this JSON format:
{
  "title": "Direct quote from historical source (in English or English translation)",
  "content": "Engaging story in English explaining the quote's historical context and significance. Write from perspective of knowledgeable local guide speaking to international visitors. Include verified historical details and archaeological insights. Use clear, engaging English prose."
}

IMPORTANT: Write everything in English. Focus on authenticity, accuracy, and engaging storytelling for international audiences.`, location, location)
}

// cleanAIResponse removes markdown code blocks and other formatting issues from AI response
func (s *AIHistorianService) cleanAIResponse(response string) string {
	// Remove markdown code blocks (```json and ```)
	re := regexp.MustCompile("```(?:json)?\\s*")
	cleaned := re.ReplaceAllString(response, "")

	// Remove trailing ```
	cleaned = strings.TrimSuffix(cleaned, "```")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Remove any leading/trailing backticks
	cleaned = strings.Trim(cleaned, "`")

	return cleaned
}
