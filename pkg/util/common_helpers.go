package util

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

// LogError logs an error with context
func LogError(message string, err error) {
	log.Printf("ERROR: %s - %v", message, err)
}

// LogInfo logs an informational message
func LogInfo(message string) {
	log.Printf("INFO: %s", message)
}

// ParseFloat safely parses a string to float64
func ParseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// ParseInt safely parses a string to int
func ParseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// TruncateString truncates a string to a specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// CalculateDistance calculates the distance between two geographic points using Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
		math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

// toRadians converts degrees to radians
func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// EstimateReadingTime estimates reading time in minutes for a given text
func EstimateReadingTime(text string) float64 {
	wordCount := len(strings.Fields(text))
	// Average reading speed: 200 words per minute
	return float64(wordCount) / 200.0
}

// EstimateNarrationDuration estimates narration duration in seconds
func EstimateNarrationDuration(text string) float64 {
	wordCount := len(strings.Fields(text))
	// Average speaking speed: 150 words per minute
	return float64(wordCount) / 150.0 * 60.0
}

// SanitizeInput removes potentially harmful characters from user input
func SanitizeInput(input string) string {
	// Remove potential SQL injection characters and excessive whitespace
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "'", "")
	input = strings.ReplaceAll(input, "\"", "")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "--", "")
	return input
}

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)
	
	// Replace spaces and special characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	
	slug = result.String()
	
	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	
	// Trim hyphens from beginning and end
	slug = strings.Trim(slug, "-")
	
	return slug
}

// FormatDuration formats a duration in seconds to a human-readable string
func FormatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	
	if duration < time.Minute {
		return strconv.Itoa(int(duration.Seconds())) + " seconds"
	}
	
	minutes := int(duration.Minutes())
	remainingSeconds := int(duration.Seconds()) % 60
	
	if remainingSeconds == 0 {
		return strconv.Itoa(minutes) + " minutes"
	}
	
	return strconv.Itoa(minutes) + "m " + strconv.Itoa(remainingSeconds) + "s"
}

// GenerateAudioFilename generates a filename for audio files
func GenerateAudioFilename(manuscriptID, language string) string {
	timestamp := time.Now().Unix()
	return "narration_" + manuscriptID + "_" + language + "_" + strconv.FormatInt(timestamp, 10) + ".mp3"
}

// ValidateCoordinates validates latitude and longitude values
func ValidateCoordinates(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// CalculateRecommendationScore calculates a score for destination recommendations
func CalculateRecommendationScore(rating float64, reviewCount int, tagMatches int, totalTags int) float64 {
	// Normalize rating (0-5 scale)
	ratingScore := rating / 5.0
	
	// Review count score (logarithmic scale, capped at 100 reviews)
	reviewScore := math.Log(float64(reviewCount+1)) / math.Log(101)
	
	// Tag match score
	tagScore := 0.0
	if totalTags > 0 {
		tagScore = float64(tagMatches) / float64(totalTags)
	}
	
	// Weighted combination
	finalScore := (ratingScore * 0.4) + (reviewScore * 0.3) + (tagScore * 0.3)
	
	return math.Round(finalScore*100) / 100 // Round to 2 decimal places
}

// ExtractLanguageFromCode extracts language name from language code
func ExtractLanguageFromCode(langCode string) string {
	languageMap := map[string]string{
		"id-ID": "Indonesian",
		"en-US": "English",
		"jv-ID": "Javanese",
		"su-ID": "Sundanese",
		"bug":   "Buginese",
	}
	
	if lang, exists := languageMap[langCode]; exists {
		return lang
	}
	
	return langCode
}

// ToJSON converts any object to JSON string safely
func ToJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		LogError("Failed to marshal to JSON", err)
		return "{}"
	}
	return string(bytes)
}

// FromJSON converts JSON string to interface{} safely
func FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// MinInt returns the minimum of two integers
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt returns the maximum of two integers  
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinFloat returns the minimum of two float64 values
func MinFloat(a, b float64) float64 {
	return math.Min(a, b)
}

// MaxFloat returns the maximum of two float64 values
func MaxFloat(a, b float64) float64 {
	return math.Max(a, b)
}

// ClampInt clamps an integer between min and max values
func ClampInt(value, min, max int) int {
	return MinInt(MaxInt(value, min), max)
}

// ClampFloat clamps a float64 between min and max values  
func ClampFloat(value, min, max float64) float64 {
	return MinFloat(MaxFloat(value, min), max)
}

// TimeAgo returns a human-readable "time ago" string
func TimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	
	years := int(duration.Hours() / (24 * 365))
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// FormatPrice formats price with currency
func FormatPrice(amount float64, currency string) string {
	if amount == 0 {
		return "Free"
	}
	
	switch currency {
	case "IDR":
		if amount >= 1000000 {
			return fmt.Sprintf("Rp %.1fM", amount/1000000)
		} else if amount >= 1000 {
			return fmt.Sprintf("Rp %.0fK", amount/1000)
		}
		return fmt.Sprintf("Rp %.0f", amount)
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// FormatRating formats rating for display (e.g., "4.5" or "5.0")
func FormatRating(rating float64) string {
	if rating == 0 {
		return "No rating"
	}
	return fmt.Sprintf("%.1f", rating)
}

// GenerateMapPinColor returns color based on destination category
func GenerateMapPinColor(category string) string {
	colorMap := map[string]string{
		"Cultural":    "#FF6B6B", // Red
		"Historical":  "#4ECDC4", // Teal
		"Religious":   "#45B7D1", // Blue
		"Nature":      "#96CEB4", // Green
		"Adventure":   "#FECA57", // Yellow
		"Beach":       "#48CAE4", // Light Blue
		"City":        "#9B59B6", // Purple
		"Traditional": "#E17055", // Orange
	}
	
	if color, exists := colorMap[category]; exists {
		return color
	}
	return "#6C757D" // Default gray
}

// ValidateImageURL validates if URL is a valid image URL
func ValidateImageURL(url string) bool {
	if url == "" {
		return false
	}
	
	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}
	
	// Check for image extensions
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
	lowerURL := strings.ToLower(url)
	
	for _, ext := range imageExtensions {
		if strings.Contains(lowerURL, ext) {
			return true
		}
	}
	
	return false
}

// BuildImageGallery creates a structured image gallery from URLs
func BuildImageGallery(images []string, maxImages int) []map[string]interface{} {
	var gallery []map[string]interface{}
	
	for i, imageURL := range images {
		if i >= maxImages {
			break
		}
		
		if ValidateImageURL(imageURL) {
			gallery = append(gallery, map[string]interface{}{
				"url":       imageURL,
				"thumbnail": imageURL, // In production, you'd generate thumbnails
				"caption":   "",
				"index":     i,
			})
		}
	}
	
	return gallery
}
