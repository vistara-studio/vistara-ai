package service

import (
	"fmt"
	"log"

	"github.com/vistara-studio/vistara-ai/infra/config"
	"github.com/vistara-studio/vistara-ai/pkg/dto"
)

// NusalingoService handles AI-powered language translation
type NusalingoService struct {
	geminiService *GeminiService
	config        *config.Config
}

// NewNusalingoService creates a new Nusalingo service instance
func NewNusalingoService(geminiService *GeminiService, cfg *config.Config) *NusalingoService {
	return &NusalingoService{
		geminiService: geminiService,
		config:        cfg,
	}
}

// TranslateText translates text from one language to another
func (s *NusalingoService) TranslateText(request *dto.NusalingoRequest) (string, error) {
	prompt := s.buildTranslationPrompt(request)

	// Call the Gemini service to generate the translation with enhanced prompting for accuracy and custom timeout
	// Use non-JSON output for plain text translation
	translatedText, err := s.geminiService.GenerateTextWithAdvancedSettings(prompt, s.config.NusalingoTimeout, true, false)
	if err != nil {
		log.Printf("AI Service error during translation: %v", err)
		return "", fmt.Errorf("failed to translate text: %w", err)
	}

	// For Nusalingo, we expect plain text response, not JSON
	// So we don't need to parse JSON, just return the cleaned text
	return s.cleanTranslationResponse(translatedText), nil
}

// buildTranslationPrompt creates an optimized prompt for Nusalingo translation
func (s *NusalingoService) buildTranslationPrompt(request *dto.NusalingoRequest) string {
	return fmt.Sprintf(`Translate from %s to %s:

Text: %s

Requirements:
- Accurate translation preserving meaning and cultural context
- Natural-sounding output in target language
- For Indonesian regional/ancient languages: use authentic linguistic sources
- Return ONLY the translated text, no explanations or formatting

Translation:`,
		request.FromLanguage,
		request.ToLanguage,
		request.Text)
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
