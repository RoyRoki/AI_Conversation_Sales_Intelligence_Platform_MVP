package agentassist

import (
	"time"

	"ai-conversation-platform/internal/models"
)

// TimingWindow represents a suggested timing window for replying
type TimingWindow struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Confidence   float64   `json:"confidence"`
	Reasoning    string    `json:"reasoning"`
}

// TimingService suggests optimal reply timing for agents
type TimingService struct{}

// NewTimingService creates a new timing service
func NewTimingService() *TimingService {
	return &TimingService{}
}

// SuggestTiming calculates optimal reply timing based on past behavior and engagement windows
// Max function size: 40 lines, max parameters: 4
func (s *TimingService) SuggestTiming(
	messages []*models.Message,
	now time.Time,
) TimingWindow {
	if len(messages) == 0 {
		return s.defaultTiming(now)
	}

	// Calculate based on past response behavior
	responsePattern := s.analyzeResponsePattern(messages)
	
	// Calculate engagement windows
	engagementWindow := s.calculateEngagementWindow(now)

	// Combine patterns
	startTime := s.combineTiming(responsePattern, engagementWindow, now)
	endTime := startTime.Add(2 * time.Hour) // 2-hour window

	return TimingWindow{
		StartTime:  startTime,
		EndTime:    endTime,
		Confidence: 0.7,
		Reasoning:   "Based on past response patterns and engagement windows",
	}
}

// analyzeResponsePattern analyzes when customer typically responds
func (s *TimingService) analyzeResponsePattern(messages []*models.Message) time.Time {
	if len(messages) < 2 {
		return time.Now()
	}

	// Find average time between customer messages
	var totalDuration time.Duration
	count := 0
	lastCustomerTime := time.Time{}

	for _, msg := range messages {
		if msg.Sender == "customer" {
			if !lastCustomerTime.IsZero() {
				totalDuration += msg.Timestamp.Sub(lastCustomerTime)
				count++
			}
			lastCustomerTime = msg.Timestamp
		}
	}

	if count == 0 {
		return time.Now()
	}

	avgDuration := totalDuration / time.Duration(count)
	return time.Now().Add(avgDuration)
}

// calculateEngagementWindow calculates optimal time of day based on patterns
func (s *TimingService) calculateEngagementWindow(now time.Time) time.Time {
	hour := now.Hour()
	
	// Peak engagement: 9 AM - 5 PM
	if hour >= 9 && hour < 17 {
		return now
	}
	
	// Before 9 AM: suggest 9 AM
	if hour < 9 {
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	}
	
	// After 5 PM: suggest next day 9 AM
	nextDay := now.Add(24 * time.Hour)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, nextDay.Location())
}

// combineTiming combines response pattern and engagement window
func (s *TimingService) combineTiming(pattern, engagement time.Time, now time.Time) time.Time {
	// Prefer engagement window if pattern is outside business hours
	if engagement.Before(pattern) {
		return engagement
	}
	return pattern
}

// defaultTiming returns default timing window
func (s *TimingService) defaultTiming(now time.Time) TimingWindow {
	startTime := now.Add(1 * time.Hour)
	return TimingWindow{
		StartTime:  startTime,
		EndTime:    startTime.Add(2 * time.Hour),
		Confidence: 0.5,
		Reasoning:   "Default timing window",
	}
}

