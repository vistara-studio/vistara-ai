package service

import (
	"fmt"
	"log"

	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// NusalingoService handles AI-powered language translation
type NusalingoService struct {
	geminiService *GeminiService
}

// NewNusalingoService creates a new Nusalingo service instance
func NewNusalingoService(geminiService *GeminiService) *NusalingoService {
	return &NusalingoService{
		geminiService: geminiService,
	}
}

// TranslateText translates text from one language to another
func (s *NusalingoService) TranslateText(request *dto.NusalingoRequest) (string, error) {
	prompt := s.buildTranslationPrompt(request)
	
	// Call the Gemini service to generate the translation with enhanced prompting for accuracy
	// Use non-JSON output for plain text translation
	translatedText, err := s.geminiService.GenerateTextWithSettings(prompt, true, false)
	if err != nil {
		log.Printf("AI Service error during translation: %v", err)
		return "", fmt.Errorf("failed to translate text: %w", err)
	}

	// For Nusalingo, we expect plain text response, not JSON
	// So we don't need to parse JSON, just return the cleaned text
	return s.cleanTranslationResponse(translatedText), nil
}

// buildTranslationPrompt creates the prompt for Nusalingo translation
func (s *NusalingoService) buildTranslationPrompt(request *dto.NusalingoRequest) string {
	return fmt.Sprintf(`Objective: Act as a precise and context-aware language translator, specializing in a wide range of languages, including modern, regional, and ancient languages of the Indonesian archipelago (Nusantara).

Context: You will be given three pieces of information:
- from_language: %s
- to_language: %s  
- text: %s

CRITICAL REQUIREMENTS: Use your most comprehensive linguistic knowledge. Prioritize accuracy, cultural nuances, and natural-sounding translations. For regional or ancient languages of Indonesia, draw from authentic linguistic sources and scholarly research.

Core Task:
Translate the provided text from %s to %s. Pay close attention to context, idiomatic expressions, and cultural nuances to provide the most accurate and natural-sounding translation possible. For regional or ancient languages, prioritize the most faithful interpretation of the original meaning.

Strict Output Format:
The response must be ONLY the translated text. Do not include any extra words, explanations, conversational phrases, or formatting. Do not use JSON format. Return only the pure translation.

Examples:
- English to Indonesian: "The quick brown fox jumps over the lazy dog." → "Rubah cokelat yang cepat melompati anjing yang malas."
- Javanese to English: "Sugeng enjing, pripun kabare?" → "Good morning, how are you?"
- Old Javanese to English: "Bhinneka Tunggal Ika" → "Unity in Diversity"

Now translate the text:`, 
		request.FromLanguage, 
		request.ToLanguage, 
		request.Text,
		request.FromLanguage,
		request.ToLanguage)
}

// cleanTranslationResponse cleans the translation response from any unwanted formatting
func (s *NusalingoService) cleanTranslationResponse(response string) string {
	// Remove any markdown formatting
	cleaned := response
	
	// Remove common prefixes that might be added by AI
	prefixes := []string{
		"Translation: ",
		"Translated text: ",
		"The translation is: ",
		"Here is the translation: ",
		"Result: ",
		"Output: ",
	}
	
	for _, prefix := range prefixes {
		if len(cleaned) > len(prefix) && cleaned[:len(prefix)] == prefix {
			cleaned = cleaned[len(prefix):]
			break
		}
	}
	
	// Remove any quotes around the translation
	if len(cleaned) >= 2 && cleaned[0] == '"' && cleaned[len(cleaned)-1] == '"' {
		cleaned = cleaned[1 : len(cleaned)-1]
	}
	
	return cleaned
}
