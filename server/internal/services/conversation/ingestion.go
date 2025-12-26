package conversation

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/abadojack/whatlanggo"
	"github.com/google/uuid"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// NormalizedMessage represents a normalized message
type NormalizedMessage struct {
	ConversationID string
	Sender         string
	Message        string
	Timestamp      time.Time
	Channel        string
	Language       string
}

// AnalyzerInterface defines the interface for AI analysis
type AnalyzerInterface interface {
	AnalyzeConversationAsync(tenantID, conversationID string, messages []*models.Message)
}

// AutoReplyInterface defines the interface for auto-reply processing
type AutoReplyInterface interface {
	ProcessAutoReply(tenantID, conversationID string) error
}

// IngestionService handles conversation ingestion
type IngestionService struct {
	conversationStorage *postgres.ConversationStorage
	analyzer            AnalyzerInterface
	autoReplyService    AutoReplyInterface
}

// NewIngestionService creates a new ingestion service
func NewIngestionService(conversationStorage *postgres.ConversationStorage) *IngestionService {
	return &IngestionService{
		conversationStorage: conversationStorage,
	}
}

// SetAnalyzer sets the AI analyzer (optional)
func (s *IngestionService) SetAnalyzer(analyzer AnalyzerInterface) {
	s.analyzer = analyzer
}

// SetAutoReplyService sets the auto-reply service (optional)
func (s *IngestionService) SetAutoReplyService(autoReplyService AutoReplyInterface) {
	s.autoReplyService = autoReplyService
}

// NormalizeMessage normalizes an incoming message into standard schema
func NormalizeMessage(rawMessage string, sender string, channel string, timestamp time.Time, conversationID string) (*NormalizedMessage, error) {
	// Validate sender
	sender = strings.ToLower(strings.TrimSpace(sender))
	if sender != "customer" && sender != "agent" {
		return nil, fmt.Errorf("invalid sender: must be 'customer' or 'agent'")
	}

	// Normalize channel
	channel = normalizeChannel(channel)

	// Auto-detect language
	language := detectLanguage(rawMessage)

	// Use provided timestamp or current time
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return &NormalizedMessage{
		ConversationID: conversationID,
		Sender:         sender,
		Message:        strings.TrimSpace(rawMessage),
		Timestamp:      timestamp,
		Channel:        channel,
		Language:       language,
	}, nil
}

// normalizeChannel normalizes channel name
func normalizeChannel(channel string) string {
	channel = strings.ToLower(strings.TrimSpace(channel))
	if channel == "" {
		return "web"
	}
	// Normalize common channel names
	switch channel {
	case "web", "webchat", "chat", "website":
		return "web"
	case "whatsapp", "wa":
		return "whatsapp"
	case "email", "e-mail":
		return "email"
	default:
		return channel
	}
}

// detectLanguage auto-detects message language
func detectLanguage(text string) string {
	if strings.TrimSpace(text) == "" {
		return "unknown"
	}

	info := whatlanggo.Detect(text)
	if info.IsReliable() {
		return info.Lang.Iso6391()
	}
	return "unknown"
}

// IngestMessage ingests a normalized message into the system and returns the message ID
func (s *IngestionService) IngestMessage(tenantID string, normalized *NormalizedMessage) (string, error) {
	// Create message ID
	messageID := uuid.New().String()

	// Create message model
	message := &models.Message{
		ID:             messageID,
		ConversationID: normalized.ConversationID,
		Sender:         normalized.Sender,
		Content:        normalized.Message,
		Channel:        normalized.Channel,
		Language:       normalized.Language,
		Timestamp:      normalized.Timestamp,
		CreatedAt:      time.Now(),
	}

	// Store message (immutable)
	if err := s.conversationStorage.CreateMessage(message); err != nil {
		return "", fmt.Errorf("failed to store message: %w", err)
	}

	// Trigger async AI analysis if analyzer is set
	if s.analyzer != nil {
		messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, normalized.ConversationID)
		if err == nil {
			s.analyzer.AnalyzeConversationAsync(tenantID, normalized.ConversationID, messages)
		}
	}

	// Trigger auto-reply check if last message was from customer and auto-reply service is set
	if s.autoReplyService != nil && normalized.Sender == "customer" {
		// Process auto-reply asynchronously (don't block message storage)
		go func() {
			if err := s.autoReplyService.ProcessAutoReply(tenantID, normalized.ConversationID); err != nil {
				log.Printf("[INGESTION] auto-reply processing failed: %v", err)
			}
		}()
	}

	return messageID, nil
}

// CreateConversation creates a new conversation
// If customerID is provided, it will check for existing active conversation first
func (s *IngestionService) CreateConversation(tenantID string, customerID *string, productID *string) (*models.Conversation, error) {
	// If customerID is provided, check for existing active conversation
	if customerID != nil && *customerID != "" {
		existingConv, err := s.conversationStorage.FindActiveConversationByCustomer(tenantID, *customerID)
		if err != nil {
			return nil, fmt.Errorf("failed to check for existing conversation: %w", err)
		}
		if existingConv != nil {
			// Return existing active conversation
			return existingConv, nil
		}
	}

	conversationID := uuid.New().String()
	now := time.Now()

	conversation := &models.Conversation{
		ID:         conversationID,
		TenantID:   tenantID,
		CustomerID: customerID,
		ProductID:  productID,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.conversationStorage.CreateConversation(tenantID, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// GetConversation retrieves a conversation with messages
func (s *IngestionService) GetConversation(tenantID, conversationID string) (*models.Conversation, []*models.Message, error) {
	conv, err := s.conversationStorage.GetConversation(tenantID, conversationID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return conv, messages, nil
}

// ListConversations lists conversations for a tenant
func (s *IngestionService) ListConversations(tenantID string, limit, offset int) ([]*models.Conversation, error) {
	conversations, err := s.conversationStorage.ListConversations(tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	return conversations, nil
}

