package ai

import (
	"fmt"
	"log"
)

const (
	// ConfidenceThreshold is the minimum confidence for AI suggestions
	ConfidenceThreshold = 0.5
)

// FallbackHandler handles graceful fallback when AI fails
type FallbackHandler struct{}

// NewFallbackHandler creates a new fallback handler
func NewFallbackHandler() *FallbackHandler {
	return &FallbackHandler{}
}

// ShouldFallback determines if fallback should be triggered
func (f *FallbackHandler) ShouldFallback(err error, confidence float64) bool {
	if err != nil {
		return true
	}

	if confidence < ConfidenceThreshold {
		return true
	}

	return false
}

// HandleFallback logs fallback event and marks for manual handling
func (f *FallbackHandler) HandleFallback(conversationID string, reason string) error {
	log.Printf("[Fallback] triggered conversation=%s reason=%s", conversationID, reason)
	// No AI suggestions will be shown - agent proceeds normally
	// This is a no-op that just logs - the system continues without AI assistance
	return nil
}

// HandleError wraps error handling with fallback
func (f *FallbackHandler) HandleError(conversationID string, err error) error {
	if err == nil {
		return nil
	}

	reason := fmt.Sprintf("AI error: %v", err)
	return f.HandleFallback(conversationID, reason)
}

// HandleLowConfidence handles low confidence scenarios
func (f *FallbackHandler) HandleLowConfidence(conversationID string, confidence float64) error {
	reason := fmt.Sprintf("low confidence: %.2f", confidence)
	return f.HandleFallback(conversationID, reason)
}

