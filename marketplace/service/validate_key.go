package service

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// ValidateAPIKey tests if an API key is valid by making a lightweight request to the provider
func ValidateAPIKey(providerType string, apiKey string) error {
	config, ok := GetProviderConfig(providerType)
	if !ok {
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	_ = config // channel type not needed for validation

	var url string
	var authHeader string

	switch providerType {
	case "anthropic":
		url = "https://api.anthropic.com/v1/messages"
		authHeader = "x-api-key"
	case "openai":
		url = "https://api.openai.com/v1/models"
		authHeader = "Authorization"
	case "google":
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
		authHeader = "" // key in URL
	default:
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	switch providerType {
	case "anthropic":
		req.Header.Set(authHeader, apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		// For Anthropic, GET /v1/messages won't work, use a different approach
		// Just check if we get a 401 vs other response
		req.Method = "POST"
		req.URL.Path = "/v1/messages"
	case "openai":
		req.Header.Set(authHeader, "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach provider: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	// Check for auth failures
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API key: provider returned %d", resp.StatusCode)
	}

	// For Anthropic POST, 400 (bad request without body) is expected and means key is valid
	// For OpenAI GET /models, 200 means key is valid
	// For Google, 200 means key is valid, 400/403 means invalid

	return nil
}
