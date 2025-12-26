package agentassist

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ai-conversation-platform/internal/ai"
	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/rules"
	"ai-conversation-platform/internal/storage/chroma"
	"ai-conversation-platform/internal/storage/postgres"
)

// Suggestion represents an AI-generated reply suggestion
type Suggestion struct {
	Text              string   `json:"text"`
	Confidence        float64  `json:"confidence"`
	ProductMatch      bool     `json:"product_match"`
	ProductRecommendations []string `json:"product_recommendations"`
	Reasoning         string   `json:"reasoning"`
}

// SuggestionsResponse represents the response with multiple suggestions
type SuggestionsResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	ContextUsed bool          `json:"context_used"`
	Metadata    *models.ConversationMetadata `json:"metadata"`
}

// AgentAssistService orchestrates agent assist use-case
type AgentAssistService struct {
	analyzer         *ai.Analyzer
	geminiClient     *ai.Client
	retriever        *chroma.Retriever
	embeddingService *ai.EmbeddingService
	ruleEngine       *rules.RuleEngine
	ruleStorage         *postgres.RuleStorage
	conversationStorage *postgres.ConversationStorage
	memoryStorage       *postgres.MemoryStorage
	brandToneStorage    *postgres.BrandToneStorage
	suggestionsStorage  *postgres.SuggestionsStorage
	confidenceScorer    *ai.ConfidenceScorer
}

// NewAgentAssistService creates a new agent assist service
func NewAgentAssistService(
	analyzer *ai.Analyzer,
	geminiClient *ai.Client,
	retriever *chroma.Retriever,
	embeddingService *ai.EmbeddingService,
	ruleEngine *rules.RuleEngine,
	ruleStorage *postgres.RuleStorage,
	conversationStorage *postgres.ConversationStorage,
	memoryStorage *postgres.MemoryStorage,
	brandToneStorage *postgres.BrandToneStorage,
	suggestionsStorage *postgres.SuggestionsStorage,
) *AgentAssistService {
	return &AgentAssistService{
		analyzer:            analyzer,
		geminiClient:        geminiClient,
		retriever:           retriever,
		embeddingService:    embeddingService,
		ruleEngine:          ruleEngine,
		ruleStorage:         ruleStorage,
		conversationStorage: conversationStorage,
		memoryStorage:       memoryStorage,
		brandToneStorage:    brandToneStorage,
		suggestionsStorage:  suggestionsStorage,
		confidenceScorer:    ai.NewConfidenceScorer(),
	}
}

// GetReplySuggestions generates AI reply suggestions for agents
// Flow: check cache → context retrieval → AI generation → rule validation → confidence scoring → return suggestions
func (s *AgentAssistService) GetReplySuggestions(tenantID, conversationID string) (*SuggestionsResponse, error) {
	log.Printf("[AGENT_ASSIST] generating suggestions conversation=%s tenant=%s", conversationID, tenantID)

	// 1. Retrieve conversation context
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return &SuggestionsResponse{
			Suggestions: []Suggestion{},
			ContextUsed: false,
		}, nil
	}

	// 1a. Extract last customer message ID for cache key
	lastCustomerMessageID := s.extractLastCustomerMessageID(messages)
	log.Printf("[AGENT_ASSIST] cache check: storage=%v last_message_id=%s", s.suggestionsStorage != nil, lastCustomerMessageID)
	
	// 1b. Check cache for existing suggestions (before getting metadata since it can change)
	if s.suggestionsStorage != nil && lastCustomerMessageID != "" {
		cached, err := s.suggestionsStorage.GetSuggestions(conversationID, lastCustomerMessageID)
		if err == nil && cached != nil {
			log.Printf("[AGENT_ASSIST] cache hit for conversation=%s last_message=%s", conversationID, lastCustomerMessageID)
			// Parse cached suggestions data (only suggestions array and context_used, not metadata)
			var cachedSuggestions []Suggestion
			if err := json.Unmarshal([]byte(cached.SuggestionsData), &cachedSuggestions); err == nil {
				// Get fresh metadata since it can change
				metadata, _ := s.conversationStorage.GetConversationMetadata(conversationID)
				return &SuggestionsResponse{
					Suggestions: cachedSuggestions,
					ContextUsed: cached.ContextUsed,
					Metadata:    metadata,
				}, nil
			}
			log.Printf("[AGENT_ASSIST] failed to parse cached suggestions, regenerating: %v", err)
		} else {
			log.Printf("[AGENT_ASSIST] cache miss for conversation=%s last_message=%s", conversationID, lastCustomerMessageID)
		}
	}

	// 2. Get conversation metadata (intent, sentiment, etc.)
	metadata, _ := s.conversationStorage.GetConversationMetadata(conversationID)

	// 3. Retrieve context: recent messages, product KB, customer memory
	context, contextScores, err := s.retrieveContext(tenantID, conversationID, messages)
	if err != nil {
		// Check if error is due to quota/API limits
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") || 
		   strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			log.Printf("[AGENT_ASSIST] context retrieval blocked by API quota: %v", err)
		} else {
			log.Printf("[AGENT_ASSIST] context retrieval failed: %v", err)
		}
		context = ""
		contextScores = []float64{}
	}

	// 4. Get customer memory if available
	customerID := s.extractCustomerID(messages)
	var customerMemory *models.CustomerMemory
	if customerID != "" {
		memory, err := s.memoryStorage.GetMemory(tenantID, customerID)
		if err == nil {
			customerMemory = memory
		}
	}

	// 5. Get brand tone
	brandTone, _ := s.getBrandTone(tenantID)

	// 6. Detect customer language for multi-language support
	customerLang := s.detectCustomerLanguage(messages)
	agentLang := "en" // Default agent language (can be configured)

	// 7. Generate AI reply suggestions with product recommendations
	suggestions, err := s.generateReplySuggestions(messages, context, customerMemory, brandTone, metadata, customerLang, agentLang)
	if err != nil {
		// generateReplySuggestions should now always return empty suggestions on error, not nil
		// But keep this as a safety net in case it still returns an error
		log.Printf("[AGENT_ASSIST] AI suggestions generation returned error (using empty suggestions): %v", err)
		return &SuggestionsResponse{
			Suggestions: []Suggestion{},
			ContextUsed: len(context) > 0,
			Metadata:    metadata,
		}, nil
	}

	// 8. Load rules for validation
	rules, err := s.ruleStorage.LoadRules(tenantID)
	if err != nil {
		log.Printf("[AGENT_ASSIST] failed to load rules: %v", err)
		rules = []*models.Rule{}
	}

	// 9. Validate suggestions through rule engine and calculate confidence
	validatedSuggestions := make([]Suggestion, 0, len(suggestions))
	for _, sug := range suggestions {
		// Validate with rule engine
		validationResult := s.ruleEngine.ValidateOutput(sug.Text, rules)

		// Skip blocked suggestions
		if validationResult.Blocked {
			log.Printf("[AGENT_ASSIST] suggestion blocked by rule engine")
			continue
		}

		// Use corrected text if auto-corrected
		if validationResult.CorrectedText != sug.Text {
			sug.Text = validationResult.CorrectedText
			log.Printf("[AGENT_ASSIST] suggestion auto-corrected by rule engine")
		}

		// Calculate confidence score
		confidenceInputs := ai.ConfidenceInputs{
			Analysis:       metadata,
			ContextScores:  contextScores,
			RuleResults:   validationResult.RuleResults,
			SelfEvaluation: sug.Confidence,
		}
		sug.Confidence = s.confidenceScorer.CalculateConfidence(confidenceInputs)

		validatedSuggestions = append(validatedSuggestions, sug)
	}

	log.Printf("[AGENT_ASSIST] generated %d suggestions conversation=%s", len(validatedSuggestions), conversationID)

	response := &SuggestionsResponse{
		Suggestions: validatedSuggestions,
		ContextUsed: len(context) > 0,
		Metadata:    metadata,
	}

	// Save to cache after successful generation (only save suggestions array, not metadata)
	if s.suggestionsStorage != nil && lastCustomerMessageID != "" {
		// Only cache the suggestions array, not the full response (metadata can change)
		suggestionsData, err := json.Marshal(response.Suggestions)
		if err == nil {
			if err := s.suggestionsStorage.SaveSuggestions(conversationID, lastCustomerMessageID, string(suggestionsData), len(context) > 0); err != nil {
				log.Printf("[AGENT_ASSIST] failed to save suggestions to cache: %v", err)
				// Don't fail the request if cache save fails
			} else {
				log.Printf("[AGENT_ASSIST] saved suggestions to cache conversation=%s last_message=%s", conversationID, lastCustomerMessageID)
			}
		}
	}

	return response, nil
}

// retrieveContext retrieves relevant context from Chroma
func (s *AgentAssistService) retrieveContext(tenantID, conversationID string, messages []*models.Message) (string, []float64, error) {
	if len(messages) == 0 {
		return "", []float64{}, nil
	}

	// Use last customer message for context retrieval
	lastMessage := messages[len(messages)-1]
	queryText := lastMessage.Content

	// Generate embedding
	embedding, err := s.embeddingService.GenerateEmbedding(queryText)
	if err != nil {
		// Check if error is due to quota/API limits - return empty context gracefully
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") || 
		   strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			log.Printf("[AGENT_ASSIST] embedding generation blocked by API quota, continuing without context")
			return "", []float64{}, nil
		}
		return "", []float64{}, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Retrieve product knowledge
	productChunks, err := s.retriever.RetrieveProductKnowledge(tenantID, embedding, 5)
	if err != nil {
		// Check if error is due to quota/API limits
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") || 
		   strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			log.Printf("[AGENT_ASSIST] product knowledge retrieval blocked by API quota")
			return "", []float64{}, nil
		}
		return "", []float64{}, fmt.Errorf("failed to retrieve product knowledge: %w", err)
	}

	// Build context from chunks
	contextParts := make([]string, 0, len(productChunks))
	contextScores := make([]float64, 0, len(productChunks))
	for _, chunk := range productChunks {
		contextParts = append(contextParts, chunk.Text)
		contextScores = append(contextScores, chunk.Score)
	}

	context := strings.Join(contextParts, "\n\n")
	return context, contextScores, nil
}

// extractCustomerID extracts customer ID from messages (uses conversation ID as proxy for now)
func (s *AgentAssistService) extractCustomerID(messages []*models.Message) string {
	// For now, use conversation ID as customer identifier
	// In production, this would extract from message metadata or conversation
	if len(messages) > 0 {
		return messages[0].ConversationID
	}
	return ""
}

// extractLastCustomerMessageID extracts the ID of the last customer message
func (s *AgentAssistService) extractLastCustomerMessageID(messages []*models.Message) string {
	// Iterate backwards to find the last customer message
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Sender == "customer" {
			return messages[i].ID
		}
	}
	return "" // No customer message found
}

// generateReplySuggestions generates reply suggestions using AI with multi-language support
func (s *AgentAssistService) generateReplySuggestions(
	messages []*models.Message,
	context string,
	customerMemory *models.CustomerMemory,
	brandTone string,
	metadata *models.ConversationMetadata,
	customerLang string,
	agentLang string,
) ([]Suggestion, error) {
	// Build conversation text
	conversationText := s.buildConversationText(messages)

	// Build prompt with context, customer memory, brand tone, and product recommendations
	prompt := s.buildSuggestionPrompt(conversationText, context, customerMemory, brandTone, metadata)

	// Use analyzer's translation support if languages differ
	if customerLang != "" && customerLang != agentLang && s.analyzer != nil {
		reply, err := s.analyzer.GenerateReplyWithTranslation(messages, agentLang, customerLang, prompt)
		if err == nil {
			// Parse the translated reply as a suggestion
			return []Suggestion{
				{
					Text:       reply,
					Confidence: 0.8,
					Reasoning:  "Generated with multi-language support",
				},
			}, nil
		}
		// If translation fails, log but continue to fallback
		log.Printf("[AGENT_ASSIST] Translation failed, falling back to direct API call: %v", err)
	}

	// Fallback to direct API call
	if s.geminiClient == nil {
		log.Printf("[AGENT_ASSIST] Gemini client not available, returning empty suggestions")
		return []Suggestion{}, nil
	}

	req := ai.GenerateTextRequest{
		Prompt:  prompt,
		Context: context,
	}

	resp, err := s.geminiClient.GenerateText(req)
	if err != nil {
		log.Printf("[AGENT_ASSIST] Gemini API error (full): %v", err)
		errStr := strings.ToLower(err.Error())
		
		// Check for quota/rate limit errors
		if strings.Contains(errStr, "quota") || strings.Contains(errStr, "429") || 
		   strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "resource_exhausted") {
			log.Printf("[AGENT_ASSIST] Gemini API quota/rate limit exceeded, returning empty suggestions")
			// Return empty suggestions instead of error to allow graceful degradation
			return []Suggestion{}, nil
		}
		
		// Check for API key or authentication errors
		if strings.Contains(errStr, "api key") || strings.Contains(errStr, "401") || 
		   strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "invalid key") {
			log.Printf("[AGENT_ASSIST] Gemini API authentication failed, returning empty suggestions")
			return []Suggestion{}, nil
		}
		
		// Check for network or connection errors
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") || 
		   strings.Contains(errStr, "network") || strings.Contains(errStr, "no such host") {
			log.Printf("[AGENT_ASSIST] Gemini API network error, returning empty suggestions: %v", err)
			return []Suggestion{}, nil
		}
		
		// For any other error, log it but still return empty suggestions gracefully
		// This ensures the UI doesn't break even if the AI service has issues
		log.Printf("[AGENT_ASSIST] Gemini API call failed with error (returning empty suggestions): %v", err)
		return []Suggestion{}, nil
	}

	// Parse suggestions from response
	suggestions := s.parseSuggestionsResponse(resp.Text)

	return suggestions, nil
}

// buildConversationText builds text from messages
func (s *AgentAssistService) buildConversationText(messages []*models.Message) string {
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		parts = append(parts, fmt.Sprintf("%s: %s", msg.Sender, msg.Content))
	}
	return strings.Join(parts, "\n")
}

// buildSuggestionPrompt builds the prompt for generating suggestions with product recommendations
func (s *AgentAssistService) buildSuggestionPrompt(
	conversationText string,
	context string,
	customerMemory *models.CustomerMemory,
	brandTone string,
	metadata *models.ConversationMetadata,
) string {
	prompt := `Generate 3 reply suggestions for an agent responding to this customer conversation.
Each suggestion should be:
- Professional and helpful
- Context-aware (use conversation history)
- Product-aware (use product knowledge if relevant)
- Personalized (consider customer preferences if available)
- Include product recommendations based on current intent, similar customer behavior patterns, and objection resolution patterns

For each suggestion, also suggest relevant products if applicable.

Return suggestions as JSON array:
[
  {"text": "suggestion 1", "confidence": 0.85, "reasoning": "why this suggestion", "product_recommendations": ["product1", "product2"]},
  {"text": "suggestion 2", "confidence": 0.80, "reasoning": "why this suggestion", "product_recommendations": []},
  {"text": "suggestion 3", "confidence": 0.75, "reasoning": "why this suggestion", "product_recommendations": ["product3"]}
]

Conversation:
` + conversationText

	// Add context if available
	if context != "" {
		prompt = "Product Knowledge Context:\n" + context + "\n\n" + prompt
	}

	// Add customer memory if available
	if customerMemory != nil {
		memoryInfo := fmt.Sprintf("Customer Preferences:\n- Language: %s\n- Pricing Sensitivity: %s\n- Product Interests: %s\n- Past Objections: %s\n",
			customerMemory.PreferredLanguage,
			customerMemory.PricingSensitivity,
			strings.Join(customerMemory.ProductInterests, ", "),
			strings.Join(customerMemory.PastObjections, ", "))
		prompt = memoryInfo + "\n" + prompt
	}

	// Add brand tone instruction
	if brandTone != "" {
		prompt = fmt.Sprintf("Brand Tone: %s\n\n", brandTone) + prompt
	}

	// Add metadata insights
	if metadata != nil {
		insights := fmt.Sprintf("Conversation Insights:\n- Intent: %s (score: %.2f)\n- Sentiment: %s (score: %.2f)\n- Objections: %s\n",
			metadata.Intent, metadata.IntentScore,
			metadata.Sentiment, metadata.SentimentScore,
			strings.Join(metadata.Objections, ", "))
		prompt = insights + "\n" + prompt
	}

	return prompt
}

// parseSuggestionsResponse parses JSON suggestions from AI response
func (s *AgentAssistService) parseSuggestionsResponse(responseText string) []Suggestion {
	// Try to extract JSON array
	jsonStart := strings.Index(responseText, "[")
	jsonEnd := strings.LastIndex(responseText, "]")
	if jsonStart == -1 || jsonEnd == -1 {
		// Fallback: return single suggestion from text
		return []Suggestion{
			{
				Text:                strings.TrimSpace(responseText),
				Confidence:          0.7,
				Reasoning:            "Generated from AI response",
				ProductRecommendations: []string{},
			},
		}
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]
	var suggestions []struct {
		Text                string   `json:"text"`
		Confidence          float64  `json:"confidence"`
		Reasoning           string   `json:"reasoning"`
		ProductRecommendations []string `json:"product_recommendations"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &suggestions); err != nil {
		// Fallback if JSON parsing fails
		log.Printf("[AGENT_ASSIST] failed to parse JSON suggestions: %v", err)
		return []Suggestion{
			{
				Text:                strings.TrimSpace(responseText),
				Confidence:          0.7,
				Reasoning:            "Generated from AI response",
				ProductRecommendations: []string{},
			},
		}
	}

	result := make([]Suggestion, 0, len(suggestions))
	for _, sug := range suggestions {
		if sug.ProductRecommendations == nil {
			sug.ProductRecommendations = []string{}
		}
		result = append(result, Suggestion{
			Text:                sug.Text,
			Confidence:          sug.Confidence,
			Reasoning:           sug.Reasoning,
			ProductRecommendations: sug.ProductRecommendations,
		})
	}

	return result
}

// detectCustomerLanguage detects customer language from messages
func (s *AgentAssistService) detectCustomerLanguage(messages []*models.Message) string {
	if len(messages) == 0 {
		return ""
	}
	
	// Use language from last customer message if available
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Sender == "customer" && messages[i].Language != "" && messages[i].Language != "unknown" {
			return messages[i].Language
		}
	}
	
	return ""
}

// getBrandTone retrieves brand tone configuration
func (s *AgentAssistService) getBrandTone(tenantID string) (string, error) {
	if s.brandToneStorage == nil {
		return "Professional", nil
	}
	return s.brandToneStorage.GetBrandTone(tenantID)
}

