package service

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// AIHistorianService handles AI-powered historical storytelling
type AIHistorianService struct {
	geminiService *GeminiService
}

// NewAIHistorianService creates a new AI historian service instance
func NewAIHistorianService(geminiService *GeminiService) *AIHistorianService {
	return &AIHistorianService{
		geminiService: geminiService,
	}
}

// GenerateHistoricalStory generates a historical story with a quote for a given location
func (s *AIHistorianService) GenerateHistoricalStory(location string) (*dto.HistoricalStoryResponse, error) {
	prompt := s.buildHistorianPrompt(location)
	
	// Call the Gemini service to generate the historical story with enhanced prompting for accuracy
	rawResponse, err := s.geminiService.GenerateTextWithGrounding(prompt, true)
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

// buildHistorianPrompt creates the prompt for the AI historian
func (s *AIHistorianService) buildHistorianPrompt(location string) string {
	return fmt.Sprintf(`Objective: Act as an expert AI historian and a creative storyteller, capable of producing structured data.

Context: The user has provided a location name: %s. This place has deep historical significance, likely documented in ancient manuscripts, chronicles, or texts.

CRITICAL REQUIREMENTS: Use only the most verified, accurate, and authentic historical information. Prioritize reliability and factual accuracy. Draw from established scholarly sources, verified archaeological findings, and authentic historical documents.

Core Task:
1. Research & Identify: Based on your most comprehensive knowledge, find a specific historical manuscript, ancient text, archaeological finding, or scholarly research directly related to %s. Focus on the most credible and well-documented sources.
2. Extract a Quote: From the most reliable and authentic source, identify and extract a short, impactful, and genuine quote that captures a key aspect of the location's history, culture, or description. This quote will be the title.
3. Craft a Narrative: Write a compelling story that explains the context of the quote using verified historical knowledge. Incorporate insights from established archaeological work, peer-reviewed historical analysis, and scholarly consensus. The story should bring the scene described in the quote to life, elaborate on its meaning, and connect it to the broader history of the place for a modern audience.
4. Adopt Persona: The story (content) should be written from the perspective of an engaging local guide or a village elder sharing a tale enriched by both traditional knowledge and scholarly understanding.
5. Accuracy Focus: Ensure all historical details are factually correct and based on credible sources. Avoid speculation or unverified claims.

Strict Output Format:
The entire response must be ONLY a single, clean JSON object with no other text or explanation before or after it. The JSON object must contain exactly two keys:
- "title": A string containing the direct quote from the manuscript or historical source.
- "content": A string containing the story that explains and expands upon the quote, incorporating verified historical knowledge and established research findings.

High-Quality Example (for "Trowulan"):
{
  "title": "Desa-desa makamulya sthananira pranata suraksita...",
  "content": "This line, which translates to 'The noble villages are well-ordered and safely protected,' comes from the ancient Nagarakretagama manuscript written by Mpu Prapanca in 1365 CE. It isn't just a dry description; it's a window into the soul of the Majapahit capital, Trowulan, under King Hayam Wuruk. Archaeological excavations have revealed that these villages were indeed meticulously planned, with sophisticated water management systems that modern urban planners still study today. Imagine walking through those villages in the 14th century. The manuscript tells us of a city not of chaos, but of careful design. Canals, like silver ribbons, crisscrossed the land, not just for travel but to nourish the lush gardens surrounding the brick homes of officials and priests. This quote reveals a kingdom that valued order and security, where the hum of a bustling, prosperous city was a testament to the power and wisdom of its rulers. The story of Trowulan isn't just in its grand temples, but in the quiet pride of these well-kept villages, a truth captured forever in that single, elegant line from one of Java's most important historical texts."
}`, location, location)
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
