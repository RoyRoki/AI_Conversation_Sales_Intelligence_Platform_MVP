package agentassist

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ai-conversation-platform/internal/ai"
	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/rules"
	"ai-conversation-platform/internal/storage/postgres"
)

// PricingRange represents a suggested pricing range
type PricingRange struct {
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
	Validated    bool    `json:"validated"`
	Blocked      bool    `json:"blocked"`
}

// PricingService generates pricing range suggestions
type PricingService struct {
	geminiClient  *ai.Client
	ruleEngine    *rules.RuleEngine
	ruleStorage   *postgres.RuleStorage
}

// NewPricingService creates a new pricing service
func NewPricingService(
	geminiClient *ai.Client,
	ruleEngine *rules.RuleEngine,
	ruleStorage *postgres.RuleStorage,
) *PricingService {
	return &PricingService{
		geminiClient: geminiClient,
		ruleEngine:   ruleEngine,
		ruleStorage:  ruleStorage,
	}
}

// SuggestPricing generates pricing range suggestions based on context
// Never auto-apply pricing suggestions - always requires admin approval
func (s *PricingService) SuggestPricing(
	tenantID string,
	messages []*models.Message,
	context string,
	customerMemory *models.CustomerMemory,
) (*PricingRange, error) {
	log.Printf("[PRICING] generating pricing suggestion tenant=%s", tenantID)

	// Build conversation text
	conversationText := s.buildConversationText(messages)

	// Build prompt for pricing suggestion
	prompt := s.buildPricingPrompt(conversationText, context, customerMemory)

	// Call AI to generate pricing range
	req := ai.GenerateTextRequest{
		Prompt:  prompt,
		Context: context,
	}

	resp, err := s.geminiClient.GenerateText(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pricing suggestion: %w", err)
	}

	// Parse pricing range from response
	pricingRange := s.parsePricingResponse(resp.Text)

	// Validate with rule engine
	rules, err := s.ruleStorage.LoadRules(tenantID)
	if err != nil {
		log.Printf("[PRICING] failed to load rules: %v", err)
		rules = []*models.Rule{}
	}

	// Validate pricing suggestion text
	pricingText := fmt.Sprintf("Price range: $%.2f - $%.2f", pricingRange.MinPrice, pricingRange.MaxPrice)
	validationResult := s.ruleEngine.ValidateOutput(pricingText, rules)

	pricingRange.Validated = validationResult.Passed
	pricingRange.Blocked = validationResult.Blocked

	if validationResult.Blocked {
		log.Printf("[PRICING] pricing suggestion blocked by rule engine")
		return nil, fmt.Errorf("pricing suggestion blocked by rule engine: %s", validationResult.Explanation)
	}

	log.Printf("[PRICING] generated pricing range $%.2f - $%.2f confidence=%.2f", 
		pricingRange.MinPrice, pricingRange.MaxPrice, pricingRange.Confidence)

	return &pricingRange, nil
}

// buildConversationText builds text from messages
func (s *PricingService) buildConversationText(messages []*models.Message) string {
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		parts = append(parts, fmt.Sprintf("%s: %s", msg.Sender, msg.Content))
	}
	return strings.Join(parts, "\n")
}

// buildPricingPrompt builds the prompt for pricing suggestions
func (s *PricingService) buildPricingPrompt(
	conversationText string,
	context string,
	customerMemory *models.CustomerMemory,
) string {
	prompt := `Based on this customer conversation, suggest a pricing range (min and max price).
Consider:
- Customer's pricing sensitivity
- Product context
- Similar customer behavior patterns
- Objection resolution patterns

Return JSON with:
{
  "min_price": 100.00,
  "max_price": 150.00,
  "confidence": 0.85,
  "reasoning": "explanation"
}

Conversation:
` + conversationText

	if context != "" {
		prompt = "Context:\n" + context + "\n\n" + prompt
	}

	if customerMemory != nil {
		memoryInfo := fmt.Sprintf("Customer Pricing Sensitivity: %s\n", customerMemory.PricingSensitivity)
		prompt = memoryInfo + "\n" + prompt
	}

	return prompt
}

// parsePricingResponse parses pricing range from AI response
func (s *PricingService) parsePricingResponse(responseText string) PricingRange {
	// Try to extract JSON
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		// Fallback: return default range
		return PricingRange{
			MinPrice:   100.0,
			MaxPrice:   150.0,
			Confidence: 0.5,
			Reasoning:  "Default pricing range",
		}
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]
	var result struct {
		MinPrice   float64 `json:"min_price"`
		MaxPrice   float64 `json:"max_price"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[PRICING] failed to parse JSON: %v", err)
		return PricingRange{
			MinPrice:   100.0,
			MaxPrice:   150.0,
			Confidence: 0.5,
			Reasoning:  "Default pricing range (parsing failed)",
		}
	}

	return PricingRange{
		MinPrice:   result.MinPrice,
		MaxPrice:   result.MaxPrice,
		Confidence: result.Confidence,
		Reasoning:  result.Reasoning,
	}
}

