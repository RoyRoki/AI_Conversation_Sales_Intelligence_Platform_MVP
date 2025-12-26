package autoreply

import (
	"fmt"
	"log"
	"time"

	"ai-conversation-platform/internal/services/agentassist"
	"ai-conversation-platform/internal/services/conversation"
	"ai-conversation-platform/internal/storage/postgres"
)

// AutoReplyService handles auto-reply functionality
type AutoReplyService struct {
	globalConfigStorage    *postgres.AutoReplyStorage
	conversationConfigStorage *postgres.AutoReplyStorage
	conversationStorage    *postgres.ConversationStorage
	agentAssistService     *agentassist.AgentAssistService
	ingestionService       *conversation.IngestionService
}

// NewAutoReplyService creates a new auto-reply service
func NewAutoReplyService(
	globalConfigStorage *postgres.AutoReplyStorage,
	conversationConfigStorage *postgres.AutoReplyStorage,
	conversationStorage *postgres.ConversationStorage,
	agentAssistService *agentassist.AgentAssistService,
	ingestionService *conversation.IngestionService,
) *AutoReplyService {
	return &AutoReplyService{
		globalConfigStorage:       globalConfigStorage,
		conversationConfigStorage:  conversationConfigStorage,
		conversationStorage:        conversationStorage,
		agentAssistService:         agentAssistService,
		ingestionService:           ingestionService,
	}
}

// EffectiveConfig represents the effective auto-reply configuration for a conversation
type EffectiveConfig struct {
	Enabled            bool
	ConfidenceThreshold float64
	Source             string // "global" or "conversation"
}

// CheckAutoReplyEnabled returns the effective auto-reply configuration
// Conversation config overrides global config if it exists
func (s *AutoReplyService) CheckAutoReplyEnabled(tenantID, conversationID string) (*EffectiveConfig, error) {
	// Check conversation-specific config first
	convConfig, err := s.conversationConfigStorage.GetConversationConfig(conversationID)
	if err == nil {
		// Conversation config exists
		threshold := convConfig.ConfidenceThreshold
		if threshold == nil {
			// Use global threshold if conversation doesn't specify
			globalConfig, err := s.globalConfigStorage.GetGlobalConfig(tenantID)
			if err != nil {
				return nil, fmt.Errorf("failed to get global config: %w", err)
			}
			threshold = &globalConfig.ConfidenceThreshold
		}
		return &EffectiveConfig{
			Enabled:            convConfig.Enabled,
			ConfidenceThreshold: *threshold,
			Source:             "conversation",
		}, nil
	}

	// Fall back to global config
	globalConfig, err := s.globalConfigStorage.GetGlobalConfig(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	return &EffectiveConfig{
		Enabled:            globalConfig.Enabled,
		ConfidenceThreshold: globalConfig.ConfidenceThreshold,
		Source:             "global",
	}, nil
}

// ShouldAutoReply checks if a suggestion meets the confidence threshold
func (s *AutoReplyService) ShouldAutoReply(tenantID, conversationID string, suggestion agentassist.Suggestion) (bool, error) {
	config, err := s.CheckAutoReplyEnabled(tenantID, conversationID)
	if err != nil {
		return false, err
	}

	if !config.Enabled {
		return false, nil
	}

	return suggestion.Confidence >= config.ConfidenceThreshold, nil
}

// ProcessAutoReply processes auto-reply for a conversation after a customer message
// This should be called asynchronously after message ingestion
func (s *AutoReplyService) ProcessAutoReply(tenantID, conversationID string) error {
	log.Printf("[AUTO_REPLY] processing conversation=%s tenant=%s", conversationID, tenantID)

	// 1. Check if auto-reply is enabled
	config, err := s.CheckAutoReplyEnabled(tenantID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to check auto-reply config: %w", err)
	}

	if !config.Enabled {
		log.Printf("[AUTO_REPLY] auto-reply disabled for conversation=%s", conversationID)
		return nil
	}

	// 2. Get conversation messages to check last sender
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	// 3. Only process if last message was from customer (avoid loops)
	lastMessage := messages[len(messages)-1]
	if lastMessage.Sender != "customer" {
		log.Printf("[AUTO_REPLY] last message not from customer, skipping conversation=%s", conversationID)
		return nil
	}

	// 4. Get AI suggestions
	suggestionsResp, err := s.agentAssistService.GetReplySuggestions(tenantID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get suggestions: %w", err)
	}

	if len(suggestionsResp.Suggestions) == 0 {
		log.Printf("[AUTO_REPLY] no suggestions available conversation=%s", conversationID)
		return nil
	}

	// 5. Find best suggestion that meets confidence threshold
	var bestSuggestion *agentassist.Suggestion
	for i := range suggestionsResp.Suggestions {
		sug := &suggestionsResp.Suggestions[i]
		if sug.Confidence >= config.ConfidenceThreshold {
			if bestSuggestion == nil || sug.Confidence > bestSuggestion.Confidence {
				bestSuggestion = sug
			}
		}
	}

	if bestSuggestion == nil {
		log.Printf("[AUTO_REPLY] no suggestion meets confidence threshold (%.2f) conversation=%s", config.ConfidenceThreshold, conversationID)
		return nil
	}

	// 6. Send the message as agent
	normalized, err := conversation.NormalizeMessage(
		bestSuggestion.Text,
		"agent",
		"web",
		time.Now(),
		conversationID,
	)
	if err != nil {
		return fmt.Errorf("failed to normalize auto-reply message: %w", err)
	}

	messageID, err := s.ingestionService.IngestMessage(tenantID, normalized)
	if err != nil {
		return fmt.Errorf("failed to send auto-reply: %w", err)
	}

	log.Printf("[AUTO_REPLY] sent auto-reply message_id=%s conversation=%s confidence=%.2f", messageID, conversationID, bestSuggestion.Confidence)
	return nil
}

