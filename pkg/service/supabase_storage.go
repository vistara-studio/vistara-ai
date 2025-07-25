package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vistara-studio/vistara-ai/infra/config"
)

// SupabaseStorageService handles file operations with Supabase Storage
type SupabaseStorageService struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
	apiKey     string
	bucket     string
}

// NewSupabaseStorageService creates a new Supabase storage service
func NewSupabaseStorageService(cfg *config.Config) *SupabaseStorageService {
	return &SupabaseStorageService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: cfg.SupabaseURL,
		apiKey:  cfg.SupabaseKey,
		bucket:  cfg.SupabaseBucket,
	}
}

// UploadFileResponse represents the response from file upload
type UploadFileResponse struct {
	Key       string `json:"Key"`
	URL       string `json:"url"`
	PublicURL string `json:"public_url"`
}

// UploadFile uploads a file to Supabase Storage
func (s *SupabaseStorageService) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (*UploadFileResponse, error) {
	// Generate unique filename
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s_%s%s", 
		strings.TrimSuffix(filename, ext),
		uuid.New().String()[:8],
		ext,
	)

	// Create upload URL
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, uniqueFilename)

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(fileContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "3600")

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Get public URL
	publicURL := s.GetPublicURL(uniqueFilename)

	return &UploadFileResponse{
		Key:       uniqueFilename,
		URL:       publicURL,
		PublicURL: publicURL,
	}, nil
}

// UploadMultipleFiles uploads multiple files at once
func (s *SupabaseStorageService) UploadMultipleFiles(ctx context.Context, files []FileUpload) ([]UploadFileResponse, error) {
	var results []UploadFileResponse

	for _, fileUpload := range files {
		result, err := s.UploadFile(ctx, fileUpload.Reader, fileUpload.Filename, fileUpload.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %s: %w", fileUpload.Filename, err)
		}
		results = append(results, *result)
	}

	return results, nil
}

// FileUpload represents a file to be uploaded
type FileUpload struct {
	Reader      io.Reader
	Filename    string
	ContentType string
}

// GetPublicURL returns the public URL for a file
func (s *SupabaseStorageService) GetPublicURL(filename string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, filename)
}

// DeleteFile deletes a file from Supabase Storage
func (s *SupabaseStorageService) DeleteFile(ctx context.Context, filename string) error {
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, filename)

	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListFiles lists files in the bucket
func (s *SupabaseStorageService) ListFiles(ctx context.Context, prefix string, limit int) ([]FileInfo, error) {
	listURL := fmt.Sprintf("%s/storage/v1/object/list/%s", s.baseURL, s.bucket)

	requestBody := map[string]interface{}{
		"prefix": prefix,
		"limit":  limit,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", listURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed with status %d: %s", resp.StatusCode, string(body))
	}

	var files []FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Add public URLs
	for i := range files {
		files[i].PublicURL = s.GetPublicURL(files[i].Name)
	}

	return files, nil
}

// FileInfo represents file information
type FileInfo struct {
	Name         string    `json:"name"`
	ID           string    `json:"id"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PublicURL    string    `json:"public_url"`
}

// UploadDestinationPhotos uploads photos for a destination
func (s *SupabaseStorageService) UploadDestinationPhotos(ctx context.Context, destinationID string, files []FileUpload) ([]string, error) {
	var photoURLs []string

	for i, fileUpload := range files {
		// Create destination-specific filename
		ext := filepath.Ext(fileUpload.Filename)
		filename := fmt.Sprintf("destinations/%s/photo_%d_%s%s", 
			destinationID, 
			i+1,
			uuid.New().String()[:8], 
			ext,
		)

		result, err := s.UploadFile(ctx, fileUpload.Reader, filename, fileUpload.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload photo %d: %w", i+1, err)
		}

		photoURLs = append(photoURLs, result.PublicURL)
	}

	return photoURLs, nil
}

// UploadUserAvatar uploads user avatar
func (s *SupabaseStorageService) UploadUserAvatar(ctx context.Context, userID string, file io.Reader, filename string, contentType string) (string, error) {
	ext := filepath.Ext(filename)
	avatarFilename := fmt.Sprintf("avatars/%s%s", userID, ext)

	result, err := s.UploadFile(ctx, file, avatarFilename, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	return result.PublicURL, nil
}
