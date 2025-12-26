package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SuggestionsCache represents cached suggestions in the database
type SuggestionsCache struct {
	ID                    string    `json:"id"`
	ConversationID        string    `json:"conversation_id"`
	LastCustomerMessageID string    `json:"last_customer_message_id"`
	SuggestionsData       string    `json:"suggestions_data"` // JSON string
	ContextUsed           bool      `json:"context_used"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// SuggestionsStorage handles suggestions cache database operations
type SuggestionsStorage struct {
	client *Client
}

// NewSuggestionsStorage creates a new suggestions storage instance
func NewSuggestionsStorage(client *Client) *SuggestionsStorage {
	return &SuggestionsStorage{client: client}
}

// GetSuggestions retrieves cached suggestions for a conversation and last customer message ID
func (s *SuggestionsStorage) GetSuggestions(conversationID, lastCustomerMessageID string) (*SuggestionsCache, error) {
	query := `
		SELECT id, conversation_id, last_customer_message_id, suggestions_data, context_used, created_at, updated_at
		FROM suggestions
		WHERE conversation_id = $1 AND last_customer_message_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	cache := &SuggestionsCache{}
	err := s.client.DB.QueryRow(query, conversationID, lastCustomerMessageID).Scan(
		&cache.ID,
		&cache.ConversationID,
		&cache.LastCustomerMessageID,
		&cache.SuggestionsData,
		&cache.ContextUsed,
		&cache.CreatedAt,
		&cache.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // No cache found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions cache: %w", err)
	}
	
	return cache, nil
}

// SaveSuggestions saves suggestions to cache
func (s *SuggestionsStorage) SaveSuggestions(conversationID, lastCustomerMessageID string, suggestionsData string, contextUsed bool) error {
	now := time.Now()
	id := uuid.New().String()
	
	// Delete old cache entry for this conversation + message ID combination
	deleteQuery := `DELETE FROM suggestions WHERE conversation_id = $1 AND last_customer_message_id = $2`
	_, err := s.client.DB.Exec(deleteQuery, conversationID, lastCustomerMessageID)
	if err != nil {
		return fmt.Errorf("failed to delete old suggestions cache: %w", err)
	}
	
	// Insert new cache entry
	insertQuery := `
		INSERT INTO suggestions (id, conversation_id, last_customer_message_id, suggestions_data, context_used, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = s.client.DB.Exec(insertQuery, id, conversationID, lastCustomerMessageID, suggestionsData, contextUsed, now, now)
	if err != nil {
		return fmt.Errorf("failed to save suggestions cache: %w", err)
	}
	
	return nil
}

// DeleteSuggestions deletes cached suggestions for a conversation (optional, for manual invalidation)
func (s *SuggestionsStorage) DeleteSuggestions(conversationID string) error {
	query := `DELETE FROM suggestions WHERE conversation_id = $1`
	_, err := s.client.DB.Exec(query, conversationID)
	if err != nil {
		return fmt.Errorf("failed to delete suggestions cache: %w", err)
	}
	return nil
}

// ParseSuggestionsData parses the JSON suggestions data string into the SuggestionsResponse structure
func ParseSuggestionsData(data string, target interface{}) error {
	return json.Unmarshal([]byte(data), target)
}

// SerializeSuggestionsData serializes suggestions data to JSON string
func SerializeSuggestionsData(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal suggestions data: %w", err)
	}
	return string(bytes), nil
}

