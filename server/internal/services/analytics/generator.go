package analytics

import (
	"log"
	"math"
	"sort"
	"time"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// LeadScore represents a lead score
type LeadScore struct {
	ConversationID string  `json:"conversation_id"`
	Score          float64 `json:"score"` // 0-100
}

// WinProbability represents win probability
type WinProbability struct {
	ConversationID string  `json:"conversation_id"`
	Probability    float64 `json:"probability"` // 0-1
}

// ChurnRisk represents churn risk assessment
type ChurnRisk struct {
	ConversationID string  `json:"conversation_id"`
	RiskScore      float64 `json:"risk_score"` // 0-1
	IsAtRisk       bool    `json:"is_at_risk"`
}

// QualityScore represents conversation quality
type QualityScore struct {
	ConversationID string  `json:"conversation_id"`
	Score          float64 `json:"score"` // 0-100
}

// CLVEstimate represents customer lifetime value estimate
type CLVEstimate struct {
	ConversationID string  `json:"conversation_id"`
	CLV            float64 `json:"clv"`
}

// SalesCyclePrediction represents sales cycle duration prediction
type SalesCyclePrediction struct {
	ConversationID string  `json:"conversation_id"`
	DurationDays   float64 `json:"duration_days"`
}

// LeadContext represents lead context information
type LeadContext struct {
	Source         string  `json:"source"`
	ProductInterest *string `json:"product_interest,omitempty"`
	CustomerType   string  `json:"customer_type"`
	Channel        string  `json:"channel"`
}

// AIInsights represents AI analysis insights
type AIInsights struct {
	Intent          string   `json:"intent"`
	PrimaryObjection *string `json:"primary_objection,omitempty"`
	SentimentTrend  string   `json:"sentiment_trend"`
	Confidence      float64  `json:"confidence"`
}

// EngagementMetrics represents engagement and timing metrics
type EngagementMetrics struct {
	LastMessageTime    string `json:"last_message_time"`
	ResponseDelayRisk  string `json:"response_delay_risk"`
	SilenceDetected    bool   `json:"silence_detected"`
}

// PrioritizedLead represents a prioritized lead
type PrioritizedLead struct {
	ConversationID    string            `json:"conversation_id"`
	CustomerEmail     *string           `json:"customer_email,omitempty"`
	WinProbability    float64           `json:"win_probability"`
	UrgencyScore      float64           `json:"urgency_score"`
	DealValue         float64           `json:"deal_value"`
	PriorityScore     float64           `json:"priority_score"`
	LeadContext       *LeadContext      `json:"lead_context,omitempty"`
	AIInsights        *AIInsights       `json:"ai_insights,omitempty"`
	Engagement        *EngagementMetrics `json:"engagement,omitempty"`
	RecommendedAction *string           `json:"recommended_action,omitempty"`
	LeadStage         *string           `json:"lead_stage,omitempty"`
	RiskFlags         []string          `json:"risk_flags,omitempty"`
}

// AnalyticsConfig contains configurable weights and thresholds
type AnalyticsConfig struct {
	// Lead scoring weights
	LeadScoreIntentWeight      float64
	LeadScoreEngagementWeight  float64
	LeadScoreSentimentWeight   float64

	// Win probability weights
	WinProbIntentWeight        float64
	WinProbSentimentWeight     float64
	WinProbObjectionWeight     float64
	WinProbResponseTimeWeight  float64
	WinProbDurationWeight      float64

	// Churn risk thresholds
	ChurnRiskThreshold         float64

	// Default values
	DefaultDealValue           float64
	DefaultSalesCycleDays      float64
	DefaultCLV                 float64
}

// DefaultAnalyticsConfig returns default configuration
func DefaultAnalyticsConfig() AnalyticsConfig {
	return AnalyticsConfig{
		LeadScoreIntentWeight:     0.4,
		LeadScoreEngagementWeight: 0.3,
		LeadScoreSentimentWeight:  0.3,
		WinProbIntentWeight:       0.25,
		WinProbSentimentWeight:    0.25,
		WinProbObjectionWeight:    0.20,
		WinProbResponseTimeWeight: 0.15,
		WinProbDurationWeight:     0.15,
		ChurnRiskThreshold:        0.6,
		DefaultDealValue:          1000.0,
		DefaultSalesCycleDays:     30.0,
		DefaultCLV:                5000.0,
	}
}

// AnalyticsService orchestrates all analytics calculations
type AnalyticsService struct {
	conversationStorage *postgres.ConversationStorage
	trendAnalyzer       *TrendAnalyzer
	config              AnalyticsConfig
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	conversationStorage *postgres.ConversationStorage,
) *AnalyticsService {
	return &AnalyticsService{
		conversationStorage: conversationStorage,
		trendAnalyzer:       NewTrendAnalyzer(),
		config:              DefaultAnalyticsConfig(),
	}
}

// SetConfig updates the analytics configuration
func (s *AnalyticsService) SetConfig(config AnalyticsConfig) {
	s.config = config
}

// CalculateLeadScore calculates lead score for a conversation
// Weighted sum: buying intent (0.4), engagement (0.3), sentiment trend (0.3)
func (s *AnalyticsService) CalculateLeadScore(
	tenantID, conversationID string,
) (LeadScore, error) {
	conv, err := s.conversationStorage.GetConversation(tenantID, conversationID)
	if err != nil {
		return LeadScore{}, err
	}

	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return LeadScore{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		// If no metadata, return default score
		return LeadScore{ConversationID: conversationID, Score: 50.0}, nil
	}

	// Calculate buying intent signal (0-1)
	intentSignal := s.calculateIntentSignal(metadata)

	// Calculate engagement frequency (0-1)
	engagementSignal := s.calculateEngagementSignal(messages, conv.CreatedAt)

	// Calculate sentiment trend (0-1)
	trends := s.trendAnalyzer.AnalyzeTrends(messages, metadata)
	sentimentTrendSignal := s.trendToSignal(trends.SentimentTrend)

	// Weighted sum
	score := intentSignal*s.config.LeadScoreIntentWeight +
		engagementSignal*s.config.LeadScoreEngagementWeight +
		sentimentTrendSignal*s.config.LeadScoreSentimentWeight

	// Scale to 0-100
	score = score * 100.0
	score = math.Max(0.0, math.Min(100.0, score))

	return LeadScore{
		ConversationID: conversationID,
		Score:          score,
	}, nil
}

// CalculateWinProbability calculates win probability for a conversation
func (s *AnalyticsService) CalculateWinProbability(
	tenantID, conversationID string,
) (WinProbability, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return WinProbability{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		return WinProbability{ConversationID: conversationID, Probability: 0.5}, nil
	}

	conv, err := s.conversationStorage.GetConversation(tenantID, conversationID)
	if err != nil {
		return WinProbability{}, err
	}

	// Intent strength (0-1)
	intentStrength := metadata.IntentScore

	// Sentiment trend (0-1)
	trends := s.trendAnalyzer.AnalyzeTrends(messages, metadata)
	sentimentTrendSignal := s.trendToSignal(trends.SentimentTrend)

	// Objection frequency (inverted: fewer objections = higher probability)
	objectionFrequency := s.calculateObjectionFrequency(messages, metadata)

	// Agent response time (faster = higher probability)
	responseTimeSignal := s.calculateResponseTimeSignal(messages)

	// Conversation duration (optimal range = higher probability)
	durationSignal := s.calculateDurationSignal(conv.CreatedAt, time.Now())

	// Weighted sum
	probability := intentStrength*s.config.WinProbIntentWeight +
		sentimentTrendSignal*s.config.WinProbSentimentWeight +
		(1.0-objectionFrequency)*s.config.WinProbObjectionWeight +
		responseTimeSignal*s.config.WinProbResponseTimeWeight +
		durationSignal*s.config.WinProbDurationWeight

	probability = math.Max(0.0, math.Min(1.0, probability))

	return WinProbability{
		ConversationID: conversationID,
		Probability:    probability,
	}, nil
}

// hasCustomerMessages checks if a conversation has at least one customer message
func (s *AnalyticsService) hasCustomerMessages(tenantID, conversationID string) bool {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		log.Printf("Error getting messages for conversation %s: %v", conversationID, err)
		return false
	}

	// Check if at least one message is from a customer
	for _, msg := range messages {
		if msg.Sender == "customer" {
			return true
		}
	}

	return false
}

// getLeadContext extracts lead context from conversation data
func (s *AnalyticsService) getLeadContext(conv *models.Conversation, messages []*models.Message) *LeadContext {
	// Source: Default to "Global Chat" (can be enhanced with routing logic later)
	source := "Global Chat"
	
	// Product Interest: From conversation.product_id
	var productInterest *string
	if conv.ProductID != nil && *conv.ProductID != "" {
		// For MVP, just use product ID. Can enhance to lookup product name later
		productInterest = conv.ProductID
	}
	
	// Customer Type: Default to "SMB" (can be derived from deal value or metadata later)
	customerType := "SMB"
	
	// Channel: Extract from messages (most common channel), default to "Web"
	channel := "Web"
	if len(messages) > 0 {
		channelCount := make(map[string]int)
		for _, msg := range messages {
			if msg.Channel != "" {
				channelCount[msg.Channel]++
			}
		}
		maxCount := 0
		for ch, count := range channelCount {
			if count > maxCount {
				maxCount = count
				channel = ch
			}
		}
	}
	
	return &LeadContext{
		Source:         source,
		ProductInterest: productInterest,
		CustomerType:   customerType,
		Channel:        channel,
	}
}

// getAIInsights formats AI insights from metadata
func (s *AnalyticsService) getAIInsights(metadata *models.ConversationMetadata, trends TrendAnalysis) *AIInsights {
	if metadata == nil {
		return &AIInsights{
			Intent:         "unknown",
			SentimentTrend: "stable",
			Confidence:     0.0,
		}
	}
	
	intent := metadata.Intent
	if intent == "" {
		intent = "unknown"
	}
	
	// Get primary objection (first one, or most common)
	var primaryObjection *string
	if len(metadata.Objections) > 0 {
		primaryObjection = &metadata.Objections[0]
	}
	
	// Sentiment trend
	sentimentTrend := "stable"
	switch trends.SentimentTrend {
	case TrendImproving:
		sentimentTrend = "improving"
	case TrendDeteriorating:
		sentimentTrend = "deteriorating"
	default:
		sentimentTrend = "stable"
	}
	
	// Confidence: average of intent and sentiment scores
	confidence := (metadata.IntentScore + metadata.SentimentScore) / 2.0
	
	return &AIInsights{
		Intent:          intent,
		PrimaryObjection: primaryObjection,
		SentimentTrend:  sentimentTrend,
		Confidence:      confidence,
	}
}

// getEngagementMetrics calculates engagement metrics from messages
func (s *AnalyticsService) getEngagementMetrics(messages []*models.Message) *EngagementMetrics {
	if len(messages) == 0 {
		return &EngagementMetrics{
			LastMessageTime:   time.Now().Format(time.RFC3339),
			ResponseDelayRisk: "low",
			SilenceDetected:   false,
		}
	}
	
	lastMessage := messages[len(messages)-1]
	lastMessageTime := lastMessage.Timestamp
	timeSinceLastMessage := time.Since(lastMessageTime)
	
	// Response delay risk
	responseDelayRisk := "low"
	if lastMessage.Sender == "customer" {
		if timeSinceLastMessage > 2*time.Hour {
			responseDelayRisk = "high"
		} else if timeSinceLastMessage > 30*time.Minute {
			responseDelayRisk = "medium"
		}
	}
	
	// Silence detected: no messages > 24 hours
	silenceDetected := timeSinceLastMessage > 24*time.Hour
	
	return &EngagementMetrics{
		LastMessageTime:   lastMessageTime.Format(time.RFC3339),
		ResponseDelayRisk: responseDelayRisk,
		SilenceDetected:   silenceDetected,
	}
}

// generateRecommendedAction generates rule-based action suggestions
func (s *AnalyticsService) generateRecommendedAction(metadata *models.ConversationMetadata, urgencyScore float64, engagement *EngagementMetrics) string {
	if metadata == nil {
		return "Continue conversation"
	}
	
	// Check for objections first
	if len(metadata.Objections) > 0 {
		for _, objection := range metadata.Objections {
			switch objection {
			case "price":
				return "Address price objection"
			case "trust":
				return "Provide testimonials/case studies"
			case "delivery":
				return "Clarify delivery timeline"
			case "competitor":
				return "Highlight differentiators"
			}
		}
	}
	
	// High urgency
	if urgencyScore >= 0.7 {
		return "Respond immediately"
	}
	
	// Buying intent
	if metadata.Intent == "buying" {
		return "Move to product-specific chat"
	}
	
	// Response delay risk
	if engagement != nil && engagement.ResponseDelayRisk == "high" {
		return "Send follow-up message"
	}
	
	return "Continue conversation"
}

// determineLeadStage determines lead stage based on conversation data
func (s *AnalyticsService) determineLeadStage(conv *models.Conversation, metadata *models.ConversationMetadata, winProb float64) string {
	if metadata == nil {
		return "discovery"
	}
	
	conversationAge := time.Since(conv.CreatedAt)
	
	// Decision: High win probability (>0.7), objections resolved
	if winProb > 0.7 && len(metadata.Objections) == 0 {
		return "decision"
	}
	
	// Evaluation: Active conversation (1-7 days), buying intent detected
	if conversationAge >= 24*time.Hour && conversationAge <= 7*24*time.Hour && metadata.Intent == "buying" {
		return "evaluation"
	}
	
	// Discovery: New conversation (< 1 day), no strong intent
	return "discovery"
}

// identifyRiskFlags identifies risk indicators
func (s *AnalyticsService) identifyRiskFlags(metadata *models.ConversationMetadata, messages []*models.Message, engagement *EngagementMetrics, trends TrendAnalysis) []string {
	var flags []string
	
	if metadata == nil {
		return flags
	}
	
	// Price objection unresolved
	if len(metadata.Objections) > 0 {
		hasPriceObjection := false
		for _, obj := range metadata.Objections {
			if obj == "price" {
				hasPriceObjection = true
				break
			}
		}
		if hasPriceObjection && len(messages) > 0 {
			// Check if last message is from agent (response sent)
			lastMessage := messages[len(messages)-1]
			if lastMessage.Sender != "agent" {
				flags = append(flags, "Price objection unresolved")
			}
		}
	}
	
	// No follow-up sent yet
	if engagement != nil && len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		if lastMessage.Sender == "customer" {
			timeSinceLastMessage := time.Since(lastMessage.Timestamp)
			if timeSinceLastMessage > 2*time.Hour {
				flags = append(flags, "No follow-up sent yet")
			}
		}
	}
	
	// Silence detected
	if engagement != nil && engagement.SilenceDetected {
		flags = append(flags, "Silence detected")
	}
	
	// Sentiment declining
	if trends.SentimentTrend == TrendDeteriorating {
		flags = append(flags, "Sentiment declining")
	}
	
	return flags
}

// PrioritizeLeads ranks leads by priority
func (s *AnalyticsService) PrioritizeLeads(
	tenantID string,
	conversationIDs []string,
) ([]PrioritizedLead, error) {
	// Deduplicate conversation IDs using a map
	uniqueIDs := make(map[string]bool)
	deduplicatedIDs := make([]string, 0)
	for _, convID := range conversationIDs {
		if !uniqueIDs[convID] {
			uniqueIDs[convID] = true
			deduplicatedIDs = append(deduplicatedIDs, convID)
		}
	}

	// Filter to only include conversations with customer messages
	var filteredIDs []string
	for _, convID := range deduplicatedIDs {
		if s.hasCustomerMessages(tenantID, convID) {
			filteredIDs = append(filteredIDs, convID)
		}
	}

	var leads []PrioritizedLead

	for _, convID := range filteredIDs {
		winProb, err := s.CalculateWinProbability(tenantID, convID)
		if err != nil {
			log.Printf("Error calculating win probability for %s: %v", convID, err)
			continue
		}

		urgencyScore := s.calculateUrgencyScore(tenantID, convID)
		dealValue := s.config.DefaultDealValue // TODO: Get from customer data

		// Priority score = weighted combination
		priorityScore := winProb.Probability*0.5 +
			urgencyScore*0.3 +
			(dealValue/s.config.DefaultDealValue)*0.2

		// Fetch conversation for context
		conv, err := s.conversationStorage.GetConversation(tenantID, convID)
		if err != nil {
			log.Printf("Error getting conversation %s: %v", convID, err)
			continue
		}

		// Fetch messages for engagement metrics
		messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, convID)
		if err != nil {
			log.Printf("Error getting messages for %s: %v", convID, err)
			messages = []*models.Message{}
		}

		// Fetch metadata for AI insights
		metadata, err := s.conversationStorage.GetConversationMetadata(convID)
		if err != nil {
			metadata = nil
		}

		// Get trends for sentiment analysis
		var trends TrendAnalysis
		if metadata != nil {
			trends = s.trendAnalyzer.AnalyzeTrends(messages, metadata)
		} else {
			trends = TrendAnalysis{
				SentimentTrend: TrendStable,
				EmotionTrend:   TrendStable,
			}
		}

		// Build enriched lead data
		leadContext := s.getLeadContext(conv, messages)
		aiInsights := s.getAIInsights(metadata, trends)
		engagement := s.getEngagementMetrics(messages)
		recommendedAction := s.generateRecommendedAction(metadata, urgencyScore, engagement)
		leadStage := s.determineLeadStage(conv, metadata, winProb.Probability)
		riskFlags := s.identifyRiskFlags(metadata, messages, engagement, trends)

		leads = append(leads, PrioritizedLead{
			ConversationID:    convID,
			WinProbability:    winProb.Probability,
			UrgencyScore:      urgencyScore,
			DealValue:         dealValue,
			PriorityScore:     priorityScore,
			LeadContext:       leadContext,
			AIInsights:        aiInsights,
			Engagement:        engagement,
			RecommendedAction: &recommendedAction,
			LeadStage:         &leadStage,
			RiskFlags:         riskFlags,
		})
	}

	// Sort by priority score (descending)
	sort.Slice(leads, func(i, j int) bool {
		return leads[i].PriorityScore > leads[j].PriorityScore
	})

	return leads, nil
}

// CalculateChurnRisk calculates churn risk for a conversation
func (s *AnalyticsService) CalculateChurnRisk(
	tenantID, conversationID string,
) (ChurnRisk, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return ChurnRisk{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		return ChurnRisk{ConversationID: conversationID, RiskScore: 0.3, IsAtRisk: false}, nil
	}

	// Sustained negative sentiment
	trends := s.trendAnalyzer.AnalyzeTrends(messages, metadata)
	negativeSentimentRisk := 0.0
	if trends.SentimentTrend == TrendDeteriorating {
		negativeSentimentRisk = 0.5
	}

	// Repeated unresolved objections
	objectionRisk := s.calculateObjectionFrequency(messages, metadata)

	// Declining engagement
	engagementRisk := s.calculateDecliningEngagement(messages)

	// Combined risk score
	riskScore := negativeSentimentRisk*0.4 + objectionRisk*0.4 + engagementRisk*0.2
	riskScore = math.Max(0.0, math.Min(1.0, riskScore))

	isAtRisk := riskScore >= s.config.ChurnRiskThreshold

	return ChurnRisk{
		ConversationID: conversationID,
		RiskScore:      riskScore,
		IsAtRisk:       isAtRisk,
	}, nil
}

// CalculateQualityScore calculates conversation quality score
func (s *AnalyticsService) CalculateQualityScore(
	tenantID, conversationID string,
) (QualityScore, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return QualityScore{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		return QualityScore{ConversationID: conversationID, Score: 50.0}, nil
	}

	// Response latency (faster = better)
	latencyScore := s.calculateLatencyScore(messages)

	// Sentiment improvement
	trends := s.trendAnalyzer.AnalyzeTrends(messages, metadata)
	sentimentImprovementScore := s.trendToSignal(trends.SentimentTrend)

	// Policy violations (fewer = better) - TODO: Get from rule engine
	policyViolationScore := 1.0 // Default: no violations

	// Conversation completion (completed = better)
	completionScore := s.calculateCompletionScore(messages)

	// Weighted average
	qualityScore := latencyScore*0.3 +
		sentimentImprovementScore*0.3 +
		policyViolationScore*0.2 +
		completionScore*0.2

	qualityScore = qualityScore * 100.0
	qualityScore = math.Max(0.0, math.Min(100.0, qualityScore))

	return QualityScore{
		ConversationID: conversationID,
		Score:          qualityScore,
	}, nil
}

// CalculateCLV estimates customer lifetime value
func (s *AnalyticsService) CalculateCLV(
	tenantID, conversationID string,
) (CLVEstimate, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return CLVEstimate{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		return CLVEstimate{ConversationID: conversationID, CLV: s.config.DefaultCLV}, nil
	}

	// Historical average (using default for MVP)
	historicalAverage := s.config.DefaultCLV

	// Engagement depth multiplier
	engagementMultiplier := s.calculateEngagementDepth(messages)

	// Purchase intent strength multiplier
	intentMultiplier := metadata.IntentScore

	// CLV = base * engagement * intent
	clv := historicalAverage * engagementMultiplier * (0.5 + intentMultiplier*0.5)
	clv = math.Max(s.config.DefaultCLV*0.1, clv) // Minimum 10% of default

	return CLVEstimate{
		ConversationID: conversationID,
		CLV:            clv,
	}, nil
}

// PredictSalesCycle predicts sales cycle duration in days
func (s *AnalyticsService) PredictSalesCycle(
	tenantID, conversationID string,
) (SalesCyclePrediction, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return SalesCyclePrediction{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		return SalesCyclePrediction{
			ConversationID: conversationID,
			DurationDays:   s.config.DefaultSalesCycleDays,
		}, nil
	}

	// Historical average (using default for MVP)
	baseDuration := s.config.DefaultSalesCycleDays

	// Urgency signals reduce duration
	urgencyMultiplier := s.calculateUrgencyScore(tenantID, conversationID)
	if urgencyMultiplier > 0.7 {
		baseDuration *= 0.7 // High urgency = 30% faster
	}

	// High engagement patterns reduce duration
	engagementPattern := s.calculateEngagementPattern(messages)
	if engagementPattern > 0.7 {
		baseDuration *= 0.8 // High engagement = 20% faster
	}

	// Strong intent reduces duration
	if metadata.IntentScore > 0.8 {
		baseDuration *= 0.85 // Strong intent = 15% faster
	}

	durationDays := math.Max(7.0, baseDuration) // Minimum 7 days

	return SalesCyclePrediction{
		ConversationID: conversationID,
		DurationDays:   durationDays,
	}, nil
}

// Helper functions

func (s *AnalyticsService) calculateIntentSignal(metadata *models.ConversationMetadata) float64 {
	if metadata.Intent == "buying" {
		return metadata.IntentScore
	}
	return 0.0
}

func (s *AnalyticsService) calculateEngagementSignal(messages []*models.Message, startTime time.Time) float64 {
	if len(messages) == 0 {
		return 0.0
	}

	duration := time.Since(startTime).Hours()
	if duration < 1.0 {
		duration = 1.0
	}

	messageFrequency := float64(len(messages)) / duration
	// Normalize: >10 messages/hour = 1.0, <1 message/hour = 0.0
	normalized := math.Min(1.0, messageFrequency/10.0)
	return math.Max(0.0, normalized)
}

func (s *AnalyticsService) trendToSignal(trend TrendLabel) float64 {
	switch trend {
	case TrendImproving:
		return 1.0
	case TrendStable:
		return 0.5
	case TrendDeteriorating:
		return 0.0
	default:
		return 0.5
	}
}

func (s *AnalyticsService) calculateObjectionFrequency(messages []*models.Message, metadata *models.ConversationMetadata) float64 {
	if len(messages) == 0 {
		return 0.0
	}
	// Count objections per message
	objectionCount := float64(len(metadata.Objections))
	return math.Min(1.0, objectionCount/float64(len(messages)))
}

func (s *AnalyticsService) calculateResponseTimeSignal(messages []*models.Message) float64 {
	if len(messages) < 2 {
		return 0.5
	}

	var responseTimes []time.Duration
	for i := 1; i < len(messages); i++ {
		if messages[i].Sender == "agent" && messages[i-1].Sender == "customer" {
			responseTime := messages[i].Timestamp.Sub(messages[i-1].Timestamp)
			responseTimes = append(responseTimes, responseTime)
		}
	}

	if len(responseTimes) == 0 {
		return 0.5
	}

	avgResponseTime := time.Duration(0)
	for _, rt := range responseTimes {
		avgResponseTime += rt
	}
	avgResponseTime /= time.Duration(len(responseTimes))

	// Faster response = higher signal (< 1 hour = 1.0, > 24 hours = 0.0)
	hours := avgResponseTime.Hours()
	if hours < 1.0 {
		return 1.0
	}
	if hours > 24.0 {
		return 0.0
	}
	return 1.0 - (hours/24.0)
}

func (s *AnalyticsService) calculateDurationSignal(startTime, endTime time.Time) float64 {
	duration := endTime.Sub(startTime).Hours()

	// Optimal duration: 1-3 days
	if duration >= 24.0 && duration <= 72.0 {
		return 1.0
	}
	// Too short (< 1 hour) or too long (> 7 days) = lower signal
	if duration < 1.0 {
		return 0.3
	}
	if duration > 168.0 { // 7 days
		return 0.3
	}
	// Between 1 hour and 1 day, or 3-7 days = 0.7
	return 0.7
}

func (s *AnalyticsService) calculateUrgencyScore(tenantID, conversationID string) float64 {
	// Check for urgency emotions or keywords
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return 0.5
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		return 0.5
	}

	urgencyKeywords := []string{"urgent", "asap", "immediately", "soon", "quickly", "fast"}
	urgencyCount := 0
	for _, msg := range messages {
		content := toLower(msg.Content)
		for _, keyword := range urgencyKeywords {
			if contains(content, keyword) {
				urgencyCount++
				break
			}
		}
	}

	// Check for urgency emotion
	hasUrgencyEmotion := false
	for _, emotion := range metadata.Emotions {
		if emotion == "urgency" {
			hasUrgencyEmotion = true
			break
		}
	}

	if hasUrgencyEmotion {
		return 0.9
	}

	// Normalize urgency count
	if urgencyCount > 0 {
		score := math.Min(1.0, float64(urgencyCount)/float64(len(messages))*10.0)
		return math.Max(0.5, score)
	}

	return 0.3
}

func (s *AnalyticsService) calculateDecliningEngagement(messages []*models.Message) float64 {
	if len(messages) < 4 {
		return 0.0
	}

	// Compare recent vs early message frequency
	midPoint := len(messages) / 2
	earlyMessages := messages[:midPoint]
	recentMessages := messages[midPoint:]

	if len(earlyMessages) == 0 || len(recentMessages) == 0 {
		return 0.0
	}

	earlyDuration := earlyMessages[len(earlyMessages)-1].Timestamp.Sub(earlyMessages[0].Timestamp).Hours()
	recentDuration := recentMessages[len(recentMessages)-1].Timestamp.Sub(recentMessages[0].Timestamp).Hours()

	if earlyDuration < 0.1 {
		earlyDuration = 0.1
	}
	if recentDuration < 0.1 {
		recentDuration = 0.1
	}

	earlyFreq := float64(len(earlyMessages)) / earlyDuration
	recentFreq := float64(len(recentMessages)) / recentDuration

	if recentFreq < earlyFreq*0.5 {
		return 0.8 // Significant decline
	}
	if recentFreq < earlyFreq*0.8 {
		return 0.5 // Moderate decline
	}
	return 0.0 // No decline
}

func (s *AnalyticsService) calculateLatencyScore(messages []*models.Message) float64 {
	return s.calculateResponseTimeSignal(messages) // Reuse response time logic
}

func (s *AnalyticsService) calculateCompletionScore(messages []*models.Message) float64 {
	if len(messages) == 0 {
		return 0.0
	}

	// Check if conversation appears completed (last message is agent, no customer follow-up for > 1 hour)
	if len(messages) < 2 {
		return 0.5
	}

	lastMessage := messages[len(messages)-1]
	if lastMessage.Sender == "agent" {
		// Check if there's a customer response after
		// For MVP, assume completion if last message is agent
		return 0.8
	}

	return 0.5
}

func (s *AnalyticsService) calculateEngagementDepth(messages []*models.Message) float64 {
	if len(messages) == 0 {
		return 0.5
	}

	// More messages = deeper engagement
	messageCount := float64(len(messages))
	// Normalize: 50+ messages = 2.0x, 10 messages = 1.0x, 1 message = 0.5x
	if messageCount >= 50.0 {
		return 2.0
	}
	if messageCount >= 10.0 {
		return 0.5 + (messageCount-10.0)/40.0*1.5 // Linear interpolation
	}
	return 0.5 + messageCount/10.0*0.5
}

func (s *AnalyticsService) calculateEngagementPattern(messages []*models.Message) float64 {
	if len(messages) < 2 {
		return 0.5
	}

	// Calculate average time between messages (shorter = more engaged)
	var totalGap time.Duration
	count := 0
	for i := 1; i < len(messages); i++ {
		gap := messages[i].Timestamp.Sub(messages[i-1].Timestamp)
		if gap < 24*time.Hour { // Ignore gaps > 24 hours
			totalGap += gap
			count++
		}
	}

	if count == 0 {
		return 0.5
	}

	avgGap := totalGap / time.Duration(count)
	// < 1 hour = 1.0, > 6 hours = 0.0
	hours := avgGap.Hours()
	if hours < 1.0 {
		return 1.0
	}
	if hours > 6.0 {
		return 0.0
	}
	return 1.0 - (hours/6.0)
}

// GetTrends retrieves trend analysis for a conversation
func (s *AnalyticsService) GetTrends(
	tenantID, conversationID string,
) (TrendAnalysis, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return TrendAnalysis{}, err
	}

	metadata, err := s.conversationStorage.GetConversationMetadata(conversationID)
	if err != nil {
		// Return stable trend if no metadata
		return TrendAnalysis{
			SentimentTrend: TrendStable,
			EmotionTrend:   TrendStable,
			SentimentSlope: 0.0,
			EmotionSlope:   0.0,
		}, nil
	}

	return s.trendAnalyzer.AnalyzeTrends(messages, metadata), nil
}

// IntentCount represents an intent with its count
type IntentCount struct {
	Intent string `json:"intent"`
	Count  int    `json:"count"`
}

// ObjectionCount represents an objection with its count
type ObjectionCount struct {
	Objection string `json:"objection"`
	Count     int    `json:"count"`
}

// DashboardMetrics represents dashboard metrics
type DashboardMetrics struct {
	TotalConversations int             `json:"total_conversations"`
	ActiveConversations int             `json:"active_conversations"`
	AverageSentiment    float64         `json:"average_sentiment"`
	WinRate             float64         `json:"win_rate"`
	ChurnRate           float64         `json:"churn_rate"`
	TopIntents          []IntentCount   `json:"top_intents"`
	TopObjections       []ObjectionCount `json:"top_objections"`
}

// GetDashboardMetrics calculates dashboard metrics for a tenant
func (s *AnalyticsService) GetDashboardMetrics(tenantID string) (DashboardMetrics, error) {
	// Get all conversations for tenant (with reasonable limit, nil customerID for admin/agent access to all conversations)
	conversations, err := s.conversationStorage.ListConversations(tenantID, nil, 1000, 0)
	if err != nil {
		return DashboardMetrics{}, err
	}

	totalConversations := len(conversations)
	activeConversations := 0
	totalSentiment := 0.0
	sentimentCount := 0
	totalWinProb := 0.0
	winProbCount := 0
	atRiskCount := 0
	intentMap := make(map[string]int)
	objectionMap := make(map[string]int)

	for _, conv := range conversations {
		// Count active conversations
		if conv.Status == "active" {
			activeConversations++
		}

		// Get metadata for sentiment and intents/objections
		metadata, err := s.conversationStorage.GetConversationMetadata(conv.ID)
		if err == nil {
			// Sentiment
			if metadata.SentimentScore > 0 {
				totalSentiment += metadata.SentimentScore
				sentimentCount++
			}

			// Intents
			if metadata.Intent != "" {
				intentMap[metadata.Intent]++
			}

			// Objections
			for _, objection := range metadata.Objections {
				objectionMap[objection]++
			}
		}

		// Calculate win probability
		winProb, err := s.CalculateWinProbability(tenantID, conv.ID)
		if err == nil {
			totalWinProb += winProb.Probability
			winProbCount++
		}

		// Calculate churn risk
		churnRisk, err := s.CalculateChurnRisk(tenantID, conv.ID)
		if err == nil && churnRisk.IsAtRisk {
			atRiskCount++
		}
	}

	// Calculate averages
	avgSentiment := 0.0
	if sentimentCount > 0 {
		avgSentiment = totalSentiment / float64(sentimentCount)
	}

	winRate := 0.0
	if winProbCount > 0 {
		winRate = totalWinProb / float64(winProbCount)
	}

	churnRate := 0.0
	if totalConversations > 0 {
		churnRate = float64(atRiskCount) / float64(totalConversations)
	}

	// Build top intents (top 5)
	type intentCount struct {
		intent string
		count  int
	}
	intentList := make([]intentCount, 0, len(intentMap))
	for intent, count := range intentMap {
		intentList = append(intentList, intentCount{intent, count})
	}
	sort.Slice(intentList, func(i, j int) bool {
		return intentList[i].count > intentList[j].count
	})
	topIntents := make([]IntentCount, 0, 5)
	for i := 0; i < len(intentList) && i < 5; i++ {
		topIntents = append(topIntents, IntentCount{
			Intent: intentList[i].intent,
			Count:  intentList[i].count,
		})
	}

	// Build top objections (top 5)
	type objectionCount struct {
		objection string
		count     int
	}
	objectionList := make([]objectionCount, 0, len(objectionMap))
	for objection, count := range objectionMap {
		objectionList = append(objectionList, objectionCount{objection, count})
	}
	sort.Slice(objectionList, func(i, j int) bool {
		return objectionList[i].count > objectionList[j].count
	})
	topObjections := make([]ObjectionCount, 0, 5)
	for i := 0; i < len(objectionList) && i < 5; i++ {
		topObjections = append(topObjections, ObjectionCount{
			Objection: objectionList[i].objection,
			Count:     objectionList[i].count,
		})
	}

	return DashboardMetrics{
		TotalConversations: totalConversations,
		ActiveConversations: activeConversations,
		AverageSentiment:    avgSentiment,
		WinRate:             winRate,
		ChurnRate:           churnRate,
		TopIntents:          topIntents,
		TopObjections:       topObjections,
	}, nil
}

