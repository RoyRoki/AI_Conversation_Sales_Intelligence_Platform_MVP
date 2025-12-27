package ai

import (
	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/chroma"
)

// ConfidenceInputs contains all signals for confidence calculation
type ConfidenceInputs struct {
	Analysis       *models.ConversationMetadata
	ContextScores  []float64
	RuleResults   []bool
	SelfEvaluation float64
}

// ConfidenceScorer calculates confidence scores from multiple signals
type ConfidenceScorer struct{}

// NewConfidenceScorer creates a new confidence scorer
func NewConfidenceScorer() *ConfidenceScorer {
	return &ConfidenceScorer{}
}

// CalculateConfidence computes confidence from multiple signals
func (c *ConfidenceScorer) CalculateConfidence(inputs ConfidenceInputs) float64 {
	contextScore := c.calculateContextRelevance(inputs.ContextScores)
	consistencyScore := c.checkSignalConsistency(inputs.Analysis)
	ruleScore := c.calculateRuleValidation(inputs.RuleResults)
	selfEvalScore := inputs.SelfEvaluation

	// Weighted combination
	confidence := 0.0
	confidence += contextScore * 0.4
	confidence += consistencyScore * 0.3
	confidence += ruleScore * 0.2
	confidence += selfEvalScore * 0.1

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// calculateContextRelevance computes relevance from context scores
func (c *ConfidenceScorer) calculateContextRelevance(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.3 // Low confidence if no context
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	avg := sum / float64(len(scores))
	if avg < 0.3 {
		return 0.3 // Low similarity = low confidence
	}

	return avg
}

// checkSignalConsistency detects contradictions
func (c *ConfidenceScorer) checkSignalConsistency(analysis *models.ConversationMetadata) float64 {
	if analysis == nil {
		return 0.0
	}

	score := 1.0

	// Multiple objections detected = reduced confidence
	if len(analysis.Objections) > 2 {
		score -= 0.2
	}

	// Contradictory signals: positive sentiment but complaint intent
	if analysis.Sentiment == "positive" && analysis.Intent == "complaint" {
		score -= 0.3
	}

	// Contradictory signals: negative sentiment but buying intent
	if analysis.Sentiment == "negative" && analysis.Intent == "buying" {
		score -= 0.2
	}

	// Multiple emotions might indicate confusion
	if len(analysis.Emotions) > 3 {
		score -= 0.1
	}

	if score < 0.0 {
		score = 0.0
	}

	return score
}

// calculateRuleValidation computes score from rule validation
func (c *ConfidenceScorer) calculateRuleValidation(ruleResults []bool) float64 {
	if len(ruleResults) == 0 {
		return 0.5 // Neutral if no rules triggered
	}

	passed := 0
	for _, result := range ruleResults {
		if result {
			passed++
		}
	}

	ratio := float64(passed) / float64(len(ruleResults))
	if ratio < 0.5 {
		return 0.3 // Low confidence if many rules failed
	}

	return ratio
}

// ExtractContextScores extracts scores from retrieved chunks
func ExtractContextScores(chunks []chroma.RetrievedChunk) []float64 {
	scores := make([]float64, 0, len(chunks))
	for _, chunk := range chunks {
		scores = append(scores, chunk.Score)
	}
	return scores
}


