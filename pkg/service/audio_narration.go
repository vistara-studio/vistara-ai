package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vistara-studio/vistara-ai/pkg/dto"
	"github.com/vistara-studio/vistara-ai/infra/config"
)

// NarrationService handles text-to-speech narration
type NarrationService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewNarrationService creates a new narration service
func NewNarrationService(cfg *config.Config) *NarrationService {
	return &NarrationService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TTSRequest represents request to external TTS service
type TTSRequest struct {
	Text        string `json:"text"`
	Language    string `json:"languageCode"`
	Voice       Voice  `json:"voice"`
	AudioConfig Audio  `json:"audioConfig"`
}

// Voice represents voice configuration
type Voice struct {
	LanguageCode string `json:"languageCode"`
	Name         string `json:"name,omitempty"`
	SsmlGender   string `json:"ssmlGender"`
}

// Audio represents audio configuration
type Audio struct {
	AudioEncoding   string  `json:"audioEncoding"`
	SpeakingRate    float64 `json:"speakingRate,omitempty"`
	Pitch           float64 `json:"pitch,omitempty"`
	VolumeGainDb    float64 `json:"volumeGainDb,omitempty"`
	SampleRateHertz int     `json:"sampleRateHertz,omitempty"`
}

// TTSResponse represents response from external TTS service
type TTSResponse struct {
	AudioContent string `json:"audioContent"`
}

// GenerateAudio generates audio narration from text
func (s *NarrationService) GenerateAudio(ctx context.Context, text string, req dto.NarrationRequest) (*dto.NarrationResponse, error) {
	// Prepare TTS request
	ttsReq := TTSRequest{
		Text:     text,
		Language: req.Language,
		Voice: Voice{
			LanguageCode: req.Language,
			SsmlGender:   strings.ToUpper(req.VoiceGender),
		},
		AudioConfig: Audio{
			AudioEncoding:   "MP3",
			SpeakingRate:    s.getSpeakingRate(req.VoiceStyle),
			Pitch:           s.getPitch(req.VoiceStyle),
			VolumeGainDb:    0.0,
			SampleRateHertz: 24000,
		},
	}

	// Set voice name based on language and gender
	ttsReq.Voice.Name = s.getVoiceName(req.Language, req.VoiceGender, req.VoiceStyle)

	// Call external TTS service (Google Cloud TTS example)
	audioURL, duration, err := s.callTTSService(ctx, ttsReq)
	if err != nil {
		return nil, fmt.Errorf("TTS service failed: %w", err)
	}

	return &dto.NarrationResponse{
		AudioURL:  audioURL,
		Duration:  duration,
		Language:  req.Language,
		CreatedAt: time.Now(),
	}, nil
}

// callTTSService calls external TTS API
func (s *NarrationService) callTTSService(ctx context.Context, req TTSRequest) (string, float64, error) {
	// For demonstration - replace with actual Google Cloud TTS or Azure TTS
	if s.cfg.Environment == "development" {
		return s.mockTTSService(req)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal TTS request: %w", err)
	}

	// Create HTTP request (example for Google Cloud TTS)
	url := fmt.Sprintf("https://texttospeech.googleapis.com/v1/text:synthesize?key=%s", s.cfg.GeminiAPIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("TTS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("TTS service error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ttsResp TTSResponse
	if err := json.NewDecoder(resp.Body).Decode(&ttsResp); err != nil {
		return "", 0, fmt.Errorf("failed to decode TTS response: %w", err)
	}

	// Save audio file and return URL
	audioURL, duration, err := s.saveAudioFile(ttsResp.AudioContent)
	if err != nil {
		return "", 0, fmt.Errorf("failed to save audio file: %w", err)
	}

	return audioURL, duration, nil
}

// mockTTSService provides mock TTS for development
func (s *NarrationService) mockTTSService(req TTSRequest) (string, float64, error) {
	// Generate mock audio URL
	audioURL := fmt.Sprintf("https://storage.vistara.com/audio/mock_%d.mp3", time.Now().Unix())
	
	// Estimate duration (roughly 150 words per minute)
	wordCount := len(strings.Fields(req.Text))
	duration := float64(wordCount) / 150.0 * 60.0 // seconds

	return audioURL, duration, nil
}

// saveAudioFile saves base64 audio content to storage
func (s *NarrationService) saveAudioFile(audioContent string) (string, float64, error) {
	// In production, decode base64 and save to cloud storage (S3, GCS, etc.)
	// For now, return a mock URL
	audioURL := fmt.Sprintf("https://storage.vistara.com/audio/narration_%d.mp3", time.Now().Unix())
	
	// Mock duration calculation
	duration := 45.0 // seconds
	
	return audioURL, duration, nil
}

// getVoiceName returns appropriate voice name based on parameters
func (s *NarrationService) getVoiceName(language, gender, style string) string {
	voiceMap := map[string]map[string]map[string]string{
		"id-ID": {
			"female": {
				"dramatic": "id-ID-Standard-A",
				"calm":     "id-ID-Standard-C", 
				"friendly": "id-ID-Wavenet-A",
			},
			"male": {
				"dramatic": "id-ID-Standard-B",
				"calm":     "id-ID-Standard-D",
				"friendly": "id-ID-Wavenet-B",
			},
		},
		"en-US": {
			"female": {
				"dramatic": "en-US-Neural2-A",
				"calm":     "en-US-Standard-C",
				"friendly": "en-US-Wavenet-A",
			},
			"male": {
				"dramatic": "en-US-Neural2-D",
				"calm":     "en-US-Standard-B", 
				"friendly": "en-US-Wavenet-B",
			},
		},
	}

	if lang, exists := voiceMap[language]; exists {
		if gend, exists := lang[gender]; exists {
			if voice, exists := gend[style]; exists {
				return voice
			}
			// Default to first available voice for gender
			for _, voice := range gend {
				return voice
			}
		}
	}

	// Fallback
	return "id-ID-Standard-A"
}

// getSpeakingRate returns speaking rate based on style
func (s *NarrationService) getSpeakingRate(style string) float64 {
	switch style {
	case "dramatic":
		return 0.85
	case "calm":
		return 0.90
	case "friendly":
		return 1.0
	default:
		return 0.95
	}
}

// getPitch returns pitch based on style  
func (s *NarrationService) getPitch(style string) float64 {
	switch style {
	case "dramatic":
		return -2.0
	case "calm":
		return -1.0
	case "friendly":
		return 1.0
	default:
		return 0.0
	}
}
