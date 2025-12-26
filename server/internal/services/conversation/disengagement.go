package conversation

import (
	"fmt"
	"time"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// DisengagementService handles silence and disengagement detection
type DisengagementService struct {
	conversationStorage *postgres.ConversationStorage
	silenceThreshold    time.Duration
	frequencyThreshold  int // messages per hour
}

// NewDisengagementService creates a new disengagement service
func NewDisengagementService(conversationStorage *postgres.ConversationStorage) *DisengagementService {
	// Default thresholds (configurable via environment variables)
	silenceThreshold := 30 * time.Minute
	frequencyThreshold := 2 // messages per hour

	return &DisengagementService{
		conversationStorage: conversationStorage,
		silenceThreshold:    silenceThreshold,
		frequencyThreshold:  frequencyThreshold,
	}
}

// DisengagementResult represents the result of disengagement detection
type DisengagementResult struct {
	IsDisengaged    bool
	Reason          string
	LastMessageTime time.Time
	MessageCount    int
}

// CheckDisengagement checks if a conversation shows signs of disengagement
func (s *DisengagementService) CheckDisengagement(tenantID, conversationID string) (*DisengagementResult, error) {
	messages, err := s.conversationStorage.GetMessagesByConversation(tenantID, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return &DisengagementResult{
			IsDisengaged: false,
			Reason:       "no messages",
		}, nil
	}

	// Get last message time
	lastMessage := messages[len(messages)-1]
	lastMessageTime := lastMessage.Timestamp
	timeSinceLastMessage := time.Since(lastMessageTime)

	// Check for silence (no reply after threshold)
	if timeSinceLastMessage > s.silenceThreshold {
		return &DisengagementResult{
			IsDisengaged:    true,
			Reason:          "silence",
			LastMessageTime: lastMessageTime,
			MessageCount:    len(messages),
		}, nil
	}

	// Check for sudden drop in message frequency
	if len(messages) >= 3 {
		recentMessages := getRecentMessages(messages, 1*time.Hour)
		if len(recentMessages) < s.frequencyThreshold {
			// Compare with previous hour if available
			previousHourMessages := getMessagesInRange(messages, 2*time.Hour, 1*time.Hour)
			if len(previousHourMessages) > s.frequencyThreshold {
				return &DisengagementResult{
					IsDisengaged:    true,
					Reason:          "frequency_drop",
					LastMessageTime: lastMessageTime,
					MessageCount:    len(messages),
				}, nil
			}
		}
	}

	return &DisengagementResult{
		IsDisengaged:    false,
		Reason:          "active",
		LastMessageTime: lastMessageTime,
		MessageCount:    len(messages),
	}, nil
}

// getRecentMessages gets messages within the specified duration from now
func getRecentMessages(messages []*models.Message, duration time.Duration) []*models.Message {
	cutoff := time.Now().Add(-duration)
	var recent []*models.Message
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Timestamp.After(cutoff) {
			recent = append([]*models.Message{messages[i]}, recent...)
		} else {
			break
		}
	}
	return recent
}

// getMessagesInRange gets messages within a time range
func getMessagesInRange(messages []*models.Message, startOffset, endOffset time.Duration) []*models.Message {
	now := time.Now()
	startTime := now.Add(-startOffset)
	endTime := now.Add(-endOffset)

	var inRange []*models.Message
	for _, msg := range messages {
		if msg.Timestamp.After(endTime) && msg.Timestamp.Before(startTime) {
			inRange = append(inRange, msg)
		}
	}
	return inRange
}

