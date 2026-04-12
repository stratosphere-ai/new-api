package service

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// ValidateAPIKey tests if an API key is valid by making a lightweight request to the provider
func ValidateAPIKey(providerType string, apiKey string) error {
	if !IsValidProvider(providerType) {
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	client := &http.Client{Timeout: 15 * time.Second}

	var req *http.Request
	var err error

	switch providerType {
	case "openai":
		req, err = http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)

	case "anthropic":
		// Use GET /v1/models which is a lightweight read-only endpoint
		req, err = http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

	case "google":
		req, err = http.NewRequest("GET",
			fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

	default:
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach provider: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API key: provider returned %d", resp.StatusCode)
	}

	return nil
}
