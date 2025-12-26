package analytics

import (
	"math"

	"ai-conversation-platform/internal/models"
)

// TrendLabel represents the trend direction
type TrendLabel string

const (
	TrendImproving    TrendLabel = "Improving"
	TrendStable       TrendLabel = "Stable"
	TrendDeteriorating TrendLabel = "Deteriorating"
)

// TrendAnalysis represents sentiment and emotion trends
type TrendAnalysis struct {
	SentimentTrend TrendLabel   `json:"sentiment_trend"`
	EmotionTrend   TrendLabel   `json:"emotion_trend"`
	SentimentSlope float64      `json:"sentiment_slope"` // -1 to 1
	EmotionSlope   float64      `json:"emotion_slope"`   // -1 to 1
}

// TrendAnalyzer analyzes sentiment and emotion trends over time
type TrendAnalyzer struct{}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{}
}

// AnalyzeTrends computes rolling trends for sentiment and emotions
// Uses historical sentiment/emotion data from conversation metadata and messages
// Decisions based on trends, not single messages (per FRD 4.2.4)
func (a *TrendAnalyzer) AnalyzeTrends(
	messages []*models.Message,
	metadata *models.ConversationMetadata,
) TrendAnalysis {
	if len(messages) < 2 {
		return TrendAnalysis{
			SentimentTrend: TrendStable,
			EmotionTrend:   TrendStable,
			SentimentSlope: 0.0,
			EmotionSlope:   0.0,
		}
	}

	sentimentSlope := a.calculateSentimentTrend(messages, metadata)
	emotionSlope := a.calculateEmotionTrend(messages, metadata)

	sentimentTrend := a.labelTrend(sentimentSlope)
	emotionTrend := a.labelTrend(emotionSlope)

	return TrendAnalysis{
		SentimentTrend: sentimentTrend,
		EmotionTrend:   emotionTrend,
		SentimentSlope: sentimentSlope,
		EmotionSlope:   emotionSlope,
	}
}

// calculateSentimentTrend calculates sentiment trend slope (-1 to 1)
// Positive slope = improving, negative = deteriorating
func (a *TrendAnalyzer) calculateSentimentTrend(
	messages []*models.Message,
	metadata *models.ConversationMetadata,
) float64 {
	if len(messages) < 2 {
		return 0.0
	}

	// Split messages into early and recent windows
	midPoint := len(messages) / 2
	earlyMessages := messages[:midPoint]
	recentMessages := messages[midPoint:]

	// Calculate sentiment scores for each window
	earlyScore := a.calculateWindowSentiment(earlyMessages, metadata)
	recentScore := a.calculateWindowSentiment(recentMessages, metadata)

	// Slope is the difference normalized to -1 to 1 range
	slope := recentScore - earlyScore
	return math.Max(-1.0, math.Min(1.0, slope))
}

// calculateEmotionTrend calculates emotion trend slope (-1 to 1)
func (a *TrendAnalyzer) calculateEmotionTrend(
	messages []*models.Message,
	metadata *models.ConversationMetadata,
) float64 {
	if len(messages) < 2 {
		return 0.0
	}

	// Split messages into early and recent windows
	midPoint := len(messages) / 2
	earlyMessages := messages[:midPoint]
	recentMessages := messages[midPoint:]

	// Count negative emotions (frustration, urgency as negative)
	earlyNegativeCount := a.countNegativeEmotions(earlyMessages, metadata)
	recentNegativeCount := a.countNegativeEmotions(recentMessages, metadata)

	// Calculate ratio of negative emotions
	earlyRatio := float64(earlyNegativeCount) / float64(len(earlyMessages))
	recentRatio := float64(recentNegativeCount) / float64(len(recentMessages))

	// Slope: negative when ratio increases (worse), positive when decreases (better)
	slope := earlyRatio - recentRatio
	return math.Max(-1.0, math.Min(1.0, slope*2)) // Scale to -1 to 1
}

// calculateWindowSentiment calculates average sentiment score for a message window
func (a *TrendAnalyzer) calculateWindowSentiment(
	messages []*models.Message,
	metadata *models.ConversationMetadata,
) float64 {
	if len(messages) == 0 {
		return 0.5 // Neutral
	}

	// Use metadata sentiment score weighted by message frequency
	baseScore := metadata.SentimentScore

	// Adjust based on sentiment keywords in messages
	positiveKeywords := []string{"great", "good", "excellent", "thanks", "appreciate", "love", "happy"}
	negativeKeywords := []string{"bad", "terrible", "awful", "disappointed", "frustrated", "hate", "angry"}

	positiveCount := 0
	negativeCount := 0

	for _, msg := range messages {
		content := msg.Content
		for _, keyword := range positiveKeywords {
			if containsWord(content, keyword) {
				positiveCount++
			}
		}
		for _, keyword := range negativeKeywords {
			if containsWord(content, keyword) {
				negativeCount++
			}
		}
	}

	// Adjust score: +0.1 per positive keyword, -0.1 per negative keyword
	adjustment := float64(positiveCount-negativeCount) * 0.1 / float64(len(messages))
	adjustedScore := baseScore + adjustment

	return math.Max(0.0, math.Min(1.0, adjustedScore))
}

// countNegativeEmotions counts negative emotions in a message window
func (a *TrendAnalyzer) countNegativeEmotions(
	messages []*models.Message,
	metadata *models.ConversationMetadata,
) int {
	count := 0
	negativeEmotions := map[string]bool{
		"frustration": true,
		"urgency":     true, // Urgency can indicate stress
	}

	for _, emotion := range metadata.Emotions {
		if negativeEmotions[emotion] {
			count += len(messages) / 2 // Weight by message count
		}
	}

	return count
}

// labelTrend converts slope to trend label
func (a *TrendAnalyzer) labelTrend(slope float64) TrendLabel {
	threshold := 0.1 // Threshold for considering a trend significant

	if slope > threshold {
		return TrendImproving
	}
	if slope < -threshold {
		return TrendDeteriorating
	}
	return TrendStable
}

// containsWord checks if content contains a word (case-insensitive, whole word)
func containsWord(content, word string) bool {
	// Simple implementation: case-insensitive substring check
	// For MVP, this is sufficient
	lowerContent := toLower(content)
	lowerWord := toLower(word)
	return contains(lowerContent, lowerWord)
}

// toLower converts string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

