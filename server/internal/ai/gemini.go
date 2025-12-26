package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Client represents a Google Gemini API client
type Client struct {
	apiKey    string
	baseURL   string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient() (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// HealthCheck verifies API connectivity
func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("%s/models?key=%s", c.baseURL, c.apiKey)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("gemini health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gemini health check returned status %d", resp.StatusCode)
	}

	return nil
}

// GenerateTextRequest represents a text generation request
type GenerateTextRequest struct {
	Prompt  string
	Context string
}

// GenerateTextResponse represents a text generation response
type GenerateTextResponse struct {
	Text string
}

// GenerateText generates text using Gemini API with retry logic
func (c *Client) GenerateText(req GenerateTextRequest) (*GenerateTextResponse, error) {
	maxRetries := 3
	baseDelay := 1 * time.Second
	
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}
		
		resp, err := c.generateTextRequest(req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		errStr := strings.ToLower(err.Error())
		
		// Check for quota exceeded errors first - these should fail immediately
		// Quota errors mean daily limit is reached and retrying won't help
		// A 429 error with quota-related messages indicates quota exceeded, not rate limit
		if strings.Contains(errStr, "429") {
			// Check if this is a quota error (not just a rate limit)
			if isQuotaExceededError(err) {
				log.Printf("[GEMINI] Quota exceeded (429), failing immediately without retry")
				return nil, err
			}
			// This is a rate limit, not quota - try to extract retry-after time
			if retryAfter := extractRetryAfter(err.Error()); retryAfter > 0 {
				// If we have a retry-after time and haven't exceeded max retries, wait and retry
				if attempt < maxRetries {
					log.Printf("[GEMINI] Rate limited, waiting %.1fs before retry", retryAfter.Seconds())
					time.Sleep(retryAfter)
					continue
				}
			}
			// If we can't extract retry time or exceeded retries, return error
			return nil, err
		}
		
		// Also check for quota errors that aren't 429 (shouldn't happen but be safe)
		if isQuotaExceededError(err) {
			log.Printf("[GEMINI] Quota exceeded, failing immediately without retry")
			return nil, err
		}
		
		// Don't retry on other quota errors or client errors (4xx)
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") || 
		   strings.Contains(err.Error(), "rate limit") ||
		   strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") ||
		   strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "404") {
			return nil, err
		}
		
		// Retry on server errors (5xx) or network errors
		if attempt < maxRetries {
			continue
		}
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// generateTextRequest performs a single API request
func (c *Client) generateTextRequest(req GenerateTextRequest) (*GenerateTextResponse, error) {
	// Use gemini-2.5-flash which is available in v1beta API
	url := fmt.Sprintf("%s/models/gemini-2.5-flash:generateContent?key=%s", c.baseURL, c.apiKey)

	// Build prompt with context if provided
	prompt := req.Prompt
	if req.Context != "" {
		prompt = fmt.Sprintf("Context: %s\n\nQuestion: %s", req.Context, req.Prompt)
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from response
	text := extractTextFromResponse(result)
	if text == "" {
		return nil, fmt.Errorf("no text in response")
	}

	return &GenerateTextResponse{Text: text}, nil
}

// GenerateEmbeddingRequest represents an embedding generation request
type GenerateEmbeddingRequest struct {
	Text string
}

// GenerateEmbeddingResponse represents an embedding generation response
type GenerateEmbeddingResponse struct {
	Embedding []float64
}

// GenerateEmbedding generates embeddings using Gemini API with retry logic
func (c *Client) GenerateEmbedding(req GenerateEmbeddingRequest) (*GenerateEmbeddingResponse, error) {
	maxRetries := 3
	baseDelay := 1 * time.Second
	
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}
		
		resp, err := c.generateEmbeddingRequest(req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		// Check for quota exceeded errors first - these should fail immediately
		// Quota errors mean daily limit is reached and retrying won't help
		if isQuotaExceededError(err) {
			log.Printf("[GEMINI] Quota exceeded, failing immediately without retry")
			return nil, err
		}
		
		// Don't retry on quota errors (429) or client errors (4xx)
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") ||
		   strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") ||
		   strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "404") {
			return nil, err
		}
		
		// Retry on server errors (5xx) or network errors
		if attempt < maxRetries {
			continue
		}
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// generateEmbeddingRequest performs a single embedding API request
func (c *Client) generateEmbeddingRequest(req GenerateEmbeddingRequest) (*GenerateEmbeddingResponse, error) {
	url := fmt.Sprintf("%s/models/embedding-001:embedContent?key=%s", c.baseURL, c.apiKey)

	payload := map[string]interface{}{
		"model": "models/embedding-001",
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{
					"text": req.Text,
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract embedding from response
	embedding := extractEmbeddingFromResponse(result)
	if len(embedding) == 0 {
		return nil, fmt.Errorf("no embedding in response")
	}

	return &GenerateEmbeddingResponse{Embedding: embedding}, nil
}

// extractTextFromResponse extracts text from Gemini API response
func extractTextFromResponse(result map[string]interface{}) string {
	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return ""
	}

	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return ""
	}

	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return ""
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return ""
	}

	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return ""
	}

	text, _ := part["text"].(string)
	return text
}

// extractEmbeddingFromResponse extracts embedding from Gemini API response
func extractEmbeddingFromResponse(result map[string]interface{}) []float64 {
	embedding, ok := result["embedding"].(map[string]interface{})
	if !ok {
		return nil
	}

	values, ok := embedding["values"].([]interface{})
	if !ok {
		return nil
	}

	embeddingVec := make([]float64, 0, len(values))
	for _, v := range values {
		if f, ok := v.(float64); ok {
			embeddingVec = append(embeddingVec, f)
		}
	}

	return embeddingVec
}

// isQuotaExceededError checks if an error indicates quota exceeded (case-insensitive)
// Quota errors should fail immediately without retry
func isQuotaExceededError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	
	// Check for various quota exceeded patterns (case-insensitive)
	// Patterns are ordered by specificity - more specific first
	quotaPatterns := []string{
		"exceeded your current quota",           // Most specific and reliable
		"quota exceeded for metric",             // From error details
		"resource_exhausted",                    // Status field value
		"\"status\"",                            // Check for status field in JSON (very common in quota errors)
		"quota exceeded",                        // Generic fallback
	}
	
	for _, pattern := range quotaPatterns {
		if strings.Contains(errStr, pattern) {
			// Additional check: if pattern is just "status", verify it's in quota error context
			if pattern == "\"status\"" {
				// Make sure this is actually a quota error by checking for other indicators
				if strings.Contains(errStr, "exceeded") || strings.Contains(errStr, "quota") {
					log.Printf("[GEMINI] Quota pattern matched: %s", pattern)
					return true
				}
			} else {
				log.Printf("[GEMINI] Quota pattern matched: %s", pattern)
				return true
			}
		}
	}
	
	return false
}

// extractRetryAfter extracts the retry-after time from Gemini API error messages
// Format: "Please retry in 18.60213991s."
func extractRetryAfter(errorBody string) time.Duration {
	// Look for "Please retry in X.XXs" pattern
	re := regexp.MustCompile(`Please retry in ([\d.]+)s`)
	matches := re.FindStringSubmatch(errorBody)
	if len(matches) > 1 {
		if seconds, err := strconv.ParseFloat(matches[1], 64); err == nil {
			// Add a small buffer (10%) to be safe
			waitTime := time.Duration(seconds * 1.1 * float64(time.Second))
			return waitTime
		}
	}
	return 0
}

