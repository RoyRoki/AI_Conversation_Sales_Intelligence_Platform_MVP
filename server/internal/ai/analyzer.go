package ai

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/rules"
	"ai-conversation-platform/internal/storage/chroma"
	"ai-conversation-platform/internal/storage/postgres"
)

// RuleLoader interface for loading rules (to keep analyzer decoupled from storage)
type RuleLoader interface {
	LoadRules(tenantID string) ([]*models.Rule, error)
}

// Analyzer handles AI analysis of conversations
type Analyzer struct {
	geminiClient      *Client
	retriever        *chroma.Retriever
	embeddingService *EmbeddingService
	metadataStorage  *postgres.ConversationStorage
	ruleEngine       *rules.RuleEngine
	ruleLoader       RuleLoader
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(geminiClient *Client, retriever *chroma.Retriever, embeddingService *EmbeddingService, metadataStorage *postgres.ConversationStorage) *Analyzer {
	return &Analyzer{
		geminiClient:      geminiClient,
		retriever:        retriever,
		embeddingService: embeddingService,
		metadataStorage:  metadataStorage,
		ruleEngine:       rules.NewRuleEngine(),
	}
}

// SetRuleLoader sets the rule loader for rule validation
func (a *Analyzer) SetRuleLoader(loader RuleLoader) {
	a.ruleLoader = loader
}

// AnalyzeConversationAsync triggers async analysis
func (a *Analyzer) AnalyzeConversationAsync(tenantID, conversationID string, messages []*models.Message) {
	go func() {
		if err := a.analyzeConversation(tenantID, conversationID, messages); err != nil {
			log.Printf("[AI] analysis failed conversation=%s error=%v", conversationID, err)
		}
	}()
}

// analyzeConversation performs the actual analysis
func (a *Analyzer) analyzeConversation(tenantID, conversationID string, messages []*models.Message) error {
	context, err := a.retrieveContext(messages)
	if err != nil {
		// Check if error is due to quota/API limits - continue without context
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") || 
		   strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			log.Printf("[AI] context retrieval blocked by API quota, continuing without context conversation=%s", conversationID)
		} else {
			log.Printf("[AI] context retrieval failed conversation=%s error=%v", conversationID, err)
		}
		context = ""
	}

	analysis, err := a.performAnalysis(messages, context)
	if err != nil {
		// Check if error is due to quota/API limits - use fallback analysis
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") || 
		   strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			log.Printf("[AI] analysis blocked by API quota, using fallback analysis conversation=%s", conversationID)
			analysis = a.performFallbackAnalysis(messages)
		} else {
			log.Printf("[AI] analysis failed conversation=%s error=%v", conversationID, err)
			return fmt.Errorf("analysis failed: %w", err)
		}
	}

	objections := a.detectObjections(messages, analysis)
	analysis.Objections = objections

	// Validate with rule engine if available
	if a.ruleEngine != nil && a.ruleLoader != nil && tenantID != "" {
		// Load rules and validate objections
		rules, err := a.ruleLoader.LoadRules(tenantID)
		if err == nil && len(rules) > 0 {
			// Validate objections against rule patterns
			conversationText := a.buildConversationText(messages)
			validatedObjections := a.ruleEngine.ValidateObjections(analysis.Objections, conversationText)
			analysis.Objections = validatedObjections

			log.Printf("[AI] rule validation complete conversation=%s validated_objections=%v",
				conversationID, validatedObjections)
		}
	}

	if err := a.storeMetadata(conversationID, analysis); err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	log.Printf("[AI] analysis complete conversation=%s intent=%s sentiment=%s objections=%v",
		conversationID, analysis.Intent, analysis.Sentiment, analysis.Objections)
	return nil
}

// retrieveContext retrieves relevant context from Chroma
func (a *Analyzer) retrieveContext(messages []*models.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	lastMessage := messages[len(messages)-1]
	queryText := lastMessage.Content

	embedding, err := a.embeddingService.GenerateEmbedding(queryText)
	if err != nil {
		return "", err
	}

	// Use tenant ID from environment (client handles scoping)
	chunks, err := a.retriever.RetrieveProductKnowledge("", embedding, 3)
	if err != nil {
		return "", err
	}

	if len(chunks) == 0 {
		return "", nil
	}

	contextParts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		contextParts = append(contextParts, chunk.Text)
	}

	return strings.Join(contextParts, "\n\n"), nil
}

// performAnalysis calls Gemini API for analysis
func (a *Analyzer) performAnalysis(messages []*models.Message, context string) (*models.ConversationMetadata, error) {
	conversationText := a.buildConversationText(messages)
	
	// Detect language from messages
	detectedLang := a.detectLanguage(messages)
	
	// Translate to English if needed for analysis
	translatedText := conversationText
	if detectedLang != "" && detectedLang != "en" {
		translated, err := a.translateText(conversationText, detectedLang, "en")
		if err == nil {
			translatedText = translated
		}
	}
	
	prompt := a.buildAnalysisPrompt(translatedText, context)

	req := GenerateTextRequest{
		Prompt:  prompt,
		Context: context,
	}

	resp, err := a.geminiClient.GenerateText(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call failed: %w", err)
	}

	return a.parseAnalysisResponse(resp.Text)
}

// buildConversationText builds text from messages
func (a *Analyzer) buildConversationText(messages []*models.Message) string {
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		parts = append(parts, fmt.Sprintf("%s: %s", msg.Sender, msg.Content))
	}
	return strings.Join(parts, "\n")
}

// buildAnalysisPrompt builds the analysis prompt
func (a *Analyzer) buildAnalysisPrompt(conversationText string, context string) string {
	prompt := `Analyze this customer conversation and return JSON with:
- intent: "buying", "support", or "complaint"
- sentiment: "positive", "neutral", or "negative"
- emotions: array of ["frustration", "urgency", "confusion", "trust", "satisfaction"]
- objections: array of ["price", "trust", "delivery", "competitor"] if any

Conversation:
` + conversationText

	if context != "" {
		prompt = "Context:\n" + context + "\n\n" + prompt
	}

	return prompt
}

// parseAnalysisResponse parses Gemini response
func (a *Analyzer) parseAnalysisResponse(responseText string) (*models.ConversationMetadata, error) {
	metadata := &models.ConversationMetadata{
		Emotions:   []string{},
		Objections: []string{},
	}

	// Try to extract JSON from response
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return a.parseTextResponse(responseText, metadata)
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return a.parseTextResponse(responseText, metadata)
	}

	if intent, ok := result["intent"].(string); ok {
		metadata.Intent = strings.ToLower(intent)
		metadata.IntentScore = 0.8
	}

	if sentiment, ok := result["sentiment"].(string); ok {
		metadata.Sentiment = strings.ToLower(sentiment)
		metadata.SentimentScore = 0.8
	}

	if emotions, ok := result["emotions"].([]interface{}); ok {
		for _, e := range emotions {
			if emotion, ok := e.(string); ok {
				metadata.Emotions = append(metadata.Emotions, strings.ToLower(emotion))
			}
		}
	}

	return metadata, nil
}

// parseTextResponse parses text response as fallback
func (a *Analyzer) parseTextResponse(text string, metadata *models.ConversationMetadata) (*models.ConversationMetadata, error) {
	text = strings.ToLower(text)
	if strings.Contains(text, "buying") || strings.Contains(text, "purchase") {
		metadata.Intent = "buying"
	} else if strings.Contains(text, "complaint") {
		metadata.Intent = "complaint"
	} else {
		metadata.Intent = "support"
	}
	metadata.IntentScore = 0.5

	if strings.Contains(text, "positive") || strings.Contains(text, "happy") {
		metadata.Sentiment = "positive"
	} else if strings.Contains(text, "negative") || strings.Contains(text, "angry") {
		metadata.Sentiment = "negative"
	} else {
		metadata.Sentiment = "neutral"
	}
	metadata.SentimentScore = 0.5

	return metadata, nil
}

// performFallbackAnalysis performs basic analysis without AI when API is unavailable
func (a *Analyzer) performFallbackAnalysis(messages []*models.Message) *models.ConversationMetadata {
	metadata := &models.ConversationMetadata{
		Emotions:   []string{},
		Objections: []string{},
	}

	// Build conversation text for keyword analysis
	conversationText := a.buildConversationText(messages)
	text := strings.ToLower(conversationText)

	// Basic intent detection
	if strings.Contains(text, "buy") || strings.Contains(text, "purchase") || strings.Contains(text, "price") || strings.Contains(text, "cost") {
		metadata.Intent = "buying"
	} else if strings.Contains(text, "complaint") || strings.Contains(text, "problem") || strings.Contains(text, "issue") {
		metadata.Intent = "complaint"
	} else {
		metadata.Intent = "support"
	}
	metadata.IntentScore = 0.6 // Lower confidence for fallback

	// Basic sentiment detection
	if strings.Contains(text, "thank") || strings.Contains(text, "great") || strings.Contains(text, "good") || strings.Contains(text, "excellent") {
		metadata.Sentiment = "positive"
	} else if strings.Contains(text, "bad") || strings.Contains(text, "terrible") || strings.Contains(text, "angry") || strings.Contains(text, "frustrated") {
		metadata.Sentiment = "negative"
	} else {
		metadata.Sentiment = "neutral"
	}
	metadata.SentimentScore = 0.6 // Lower confidence for fallback

	// Detect objections using keyword matching
	objections := a.detectObjections(messages, metadata)
	metadata.Objections = objections

	return metadata
}

// detectObjections combines AI results with keyword matching
func (a *Analyzer) detectObjections(messages []*models.Message, analysis *models.ConversationMetadata) []string {
	objections := make(map[string]bool)

	// Add AI-detected objections
	for _, obj := range analysis.Objections {
		objections[strings.ToLower(obj)] = true
	}

	// Keyword matching
	text := a.buildConversationText(messages)
	text = strings.ToLower(text)

	keywordMap := map[string]string{
		"price":      "price",
		"expensive":  "price",
		"cost":       "price",
		"cheaper":    "price",
		"trust":      "trust",
		"reliable":   "trust",
		"delivery":   "delivery",
		"shipping":   "delivery",
		"competitor": "competitor",
		"alternative": "competitor",
	}

	for keyword, objection := range keywordMap {
		if strings.Contains(text, keyword) {
			objections[objection] = true
		}
	}

	result := make([]string, 0, len(objections))
	for obj := range objections {
		result = append(result, obj)
	}

	return result
}

// storeMetadata stores analysis results
func (a *Analyzer) storeMetadata(conversationID string, analysis *models.ConversationMetadata) error {
	analysis.ConversationID = conversationID
	if analysis.ID == "" {
		analysis.ID = uuid.New().String()
	}
	analysis.UpdatedAt = time.Now()

	return a.metadataStorage.CreateConversationMetadata(analysis)
}

// detectLanguage detects the primary language from messages
func (a *Analyzer) detectLanguage(messages []*models.Message) string {
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

// translateText translates text using Gemini API
// Translation pipeline: detect → translate → reason → generate → translate back
func (a *Analyzer) translateText(text, fromLang, toLang string) (string, error) {
	if fromLang == toLang {
		return text, nil
	}
	
	prompt := fmt.Sprintf("Translate the following text from %s to %s. Return only the translated text, no explanations:\n\n%s", fromLang, toLang, text)
	
	req := GenerateTextRequest{
		Prompt: prompt,
	}
	
	resp, err := a.geminiClient.GenerateText(req)
	if err != nil {
		return text, fmt.Errorf("translation failed: %w", err)
	}
	
	return strings.TrimSpace(resp.Text), nil
}

// GenerateReplyWithTranslation generates a reply in the agent's language, handling customer language transparently
func (a *Analyzer) GenerateReplyWithTranslation(messages []*models.Message, agentLang, customerLang string, prompt string) (string, error) {
	// Detect customer language if not provided
	if customerLang == "" {
		customerLang = a.detectLanguage(messages)
	}
	
	// If customer language is different from agent language, translate customer messages
	conversationText := a.buildConversationText(messages)
	if customerLang != "" && customerLang != agentLang && customerLang != "en" {
		translated, err := a.translateText(conversationText, customerLang, agentLang)
		if err == nil {
			conversationText = translated
		}
	}
	
	// Generate reply in agent's language
	fullPrompt := fmt.Sprintf("%s\n\nConversation:\n%s", prompt, conversationText)
	req := GenerateTextRequest{
		Prompt: fullPrompt,
	}
	
	resp, err := a.geminiClient.GenerateText(req)
	if err != nil {
		return "", fmt.Errorf("failed to generate reply: %w", err)
	}
	
	reply := strings.TrimSpace(resp.Text)
	
	// Translate back to customer language if needed
	if customerLang != "" && customerLang != agentLang && customerLang != "en" {
		translated, err := a.translateText(reply, agentLang, customerLang)
		if err == nil {
			reply = translated
		}
	}
	
	return reply, nil
}

