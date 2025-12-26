package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"ai-conversation-platform/internal/models"
)

// ConversationStorage handles conversation-related database operations
type ConversationStorage struct {
	client *Client
}

// NewConversationStorage creates a new conversation storage instance
func NewConversationStorage(client *Client) *ConversationStorage {
	return &ConversationStorage{client: client}
}

// CreateConversation creates a new conversation
func (s *ConversationStorage) CreateConversation(tenantID string, conv *models.Conversation) error {
	query := `
		INSERT INTO conversations (id, tenant_id, customer_id, product_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.client.DB.Exec(query, conv.ID, tenantID, conv.CustomerID, conv.ProductID, conv.Status, conv.CreatedAt, conv.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

// GetConversation retrieves a conversation by ID (tenant-scoped)
func (s *ConversationStorage) GetConversation(tenantID, conversationID string) (*models.Conversation, error) {
	query := `
		SELECT id, tenant_id, customer_id, product_id, status, created_at, updated_at
		FROM conversations
		WHERE id = $1 AND tenant_id = $2
	`
	conv := &models.Conversation{}
	var customerID sql.NullString
	var productID sql.NullString
	err := s.client.DB.QueryRow(query, conversationID, tenantID).Scan(
		&conv.ID, &conv.TenantID, &customerID, &productID, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if customerID.Valid {
		conv.CustomerID = &customerID.String
	}
	if productID.Valid {
		conv.ProductID = &productID.String
	}
	return conv, nil
}

// UpdateConversation updates conversation status (tenant-scoped)
func (s *ConversationStorage) UpdateConversation(tenantID string, conv *models.Conversation) error {
	query := `
		UPDATE conversations
		SET status = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`
	result, err := s.client.DB.Exec(query, conv.Status, conv.UpdatedAt, conv.ID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("conversation not found")
	}
	return nil
}

// FindActiveConversationByCustomer finds an active conversation for a customer
func (s *ConversationStorage) FindActiveConversationByCustomer(tenantID, customerID string) (*models.Conversation, error) {
	query := `
		SELECT id, tenant_id, customer_id, product_id, status, created_at, updated_at
		FROM conversations
		WHERE tenant_id = $1 AND customer_id = $2 AND status = 'active'
		ORDER BY updated_at DESC
		LIMIT 1
	`
	conv := &models.Conversation{}
	var customerIDVal sql.NullString
	var productID sql.NullString
	err := s.client.DB.QueryRow(query, tenantID, customerID).Scan(
		&conv.ID, &conv.TenantID, &customerIDVal, &productID, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No active conversation found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find active conversation: %w", err)
	}
	if customerIDVal.Valid {
		conv.CustomerID = &customerIDVal.String
	}
	if productID.Valid {
		conv.ProductID = &productID.String
	}
	return conv, nil
}

// ListConversations lists conversations for a tenant with pagination
func (s *ConversationStorage) ListConversations(tenantID string, limit, offset int) ([]*models.Conversation, error) {
	query := `
		SELECT id, tenant_id, customer_id, product_id, status, created_at, updated_at
		FROM conversations
		WHERE tenant_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.client.DB.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*models.Conversation
	for rows.Next() {
		conv := &models.Conversation{}
		var customerID sql.NullString
		var productID sql.NullString
		err := rows.Scan(&conv.ID, &conv.TenantID, &customerID, &productID, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		if customerID.Valid {
			conv.CustomerID = &customerID.String
		}
		if productID.Valid {
			conv.ProductID = &productID.String
		}
		conversations = append(conversations, conv)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %w", err)
	}
	return conversations, nil
}

// CreateMessage creates a new message (immutable)
func (s *ConversationStorage) CreateMessage(msg *models.Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, sender, content, channel, language, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.client.DB.Exec(query,
		msg.ID, msg.ConversationID, msg.Sender, msg.Content,
		msg.Channel, msg.Language, msg.Timestamp, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

// GetMessage retrieves a message by ID
func (s *ConversationStorage) GetMessage(messageID string) (*models.Message, error) {
	query := `
		SELECT id, conversation_id, sender, content, channel, language, timestamp, created_at
		FROM messages
		WHERE id = $1
	`
	msg := &models.Message{}
	err := s.client.DB.QueryRow(query, messageID).Scan(
		&msg.ID, &msg.ConversationID, &msg.Sender, &msg.Content,
		&msg.Channel, &msg.Language, &msg.Timestamp, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return msg, nil
}

// GetMessagesByConversation retrieves all messages for a conversation (tenant-scoped via conversation)
func (s *ConversationStorage) GetMessagesByConversation(tenantID, conversationID string) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender, m.content, m.channel, m.language, m.timestamp, m.created_at
		FROM messages m
		INNER JOIN conversations c ON m.conversation_id = c.id
		WHERE m.conversation_id = $1 AND c.tenant_id = $2
		ORDER BY m.timestamp ASC
	`
	rows, err := s.client.DB.Query(query, conversationID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.Sender, &msg.Content,
			&msg.Channel, &msg.Language, &msg.Timestamp, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}
	return messages, nil
}

// CreateConversationMetadata creates or updates conversation metadata
func (s *ConversationStorage) CreateConversationMetadata(metadata *models.ConversationMetadata) error {
	emotionsJSON, _ := json.Marshal(metadata.Emotions)
	objectionsJSON, _ := json.Marshal(metadata.Objections)

	// SQLite uses different ON CONFLICT syntax, so we'll use a simpler approach
	if s.client.DBType == "sqlite" {
		query := `
			INSERT OR REPLACE INTO conversation_metadata (id, conversation_id, intent, intent_score, sentiment, sentiment_score, emotions, objections, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := s.client.DB.Exec(query,
			metadata.ID, metadata.ConversationID, metadata.Intent, metadata.IntentScore,
			metadata.Sentiment, metadata.SentimentScore,
			string(emotionsJSON), string(objectionsJSON), metadata.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create conversation metadata: %w", err)
		}
		return nil
	}

	query := `
		INSERT INTO conversation_metadata (id, conversation_id, intent, intent_score, sentiment, sentiment_score, emotions, objections, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(conversation_id) DO UPDATE SET
			intent = excluded.intent,
			intent_score = excluded.intent_score,
			sentiment = excluded.sentiment,
			sentiment_score = excluded.sentiment_score,
			emotions = excluded.emotions,
			objections = excluded.objections,
			updated_at = excluded.updated_at
	`
	_, err := s.client.DB.Exec(query,
		metadata.ID, metadata.ConversationID, metadata.Intent, metadata.IntentScore,
		metadata.Sentiment, metadata.SentimentScore,
		string(emotionsJSON), string(objectionsJSON), metadata.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create conversation metadata: %w", err)
	}
	return nil
}

// GetConversationMetadata retrieves metadata for a conversation
func (s *ConversationStorage) GetConversationMetadata(conversationID string) (*models.ConversationMetadata, error) {
	query := `
		SELECT id, conversation_id, intent, intent_score, sentiment, sentiment_score, emotions, objections, updated_at
		FROM conversation_metadata
		WHERE conversation_id = $1
	`
	metadata := &models.ConversationMetadata{}
	var emotionsJSON, objectionsJSON string

	err := s.client.DB.QueryRow(query, conversationID).Scan(
		&metadata.ID, &metadata.ConversationID, &metadata.Intent, &metadata.IntentScore,
		&metadata.Sentiment, &metadata.SentimentScore,
		&emotionsJSON, &objectionsJSON, &metadata.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("metadata not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if err := json.Unmarshal([]byte(emotionsJSON), &metadata.Emotions); err != nil {
		metadata.Emotions = []string{}
	}
	if err := json.Unmarshal([]byte(objectionsJSON), &metadata.Objections); err != nil {
		metadata.Objections = []string{}
	}

	return metadata, nil
}

// UpdateConversationMetadata updates conversation metadata
func (s *ConversationStorage) UpdateConversationMetadata(metadata *models.ConversationMetadata) error {
	emotionsJSON, _ := json.Marshal(metadata.Emotions)
	objectionsJSON, _ := json.Marshal(metadata.Objections)

	query := `
		UPDATE conversation_metadata
		SET intent = $1, intent_score = $2, sentiment = $3, sentiment_score = $4,
		    emotions = $5, objections = $6, updated_at = $7
		WHERE conversation_id = $8
	`
	result, err := s.client.DB.Exec(query,
		metadata.Intent, metadata.IntentScore, metadata.Sentiment, metadata.SentimentScore,
		string(emotionsJSON), string(objectionsJSON), metadata.UpdatedAt, metadata.ConversationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("metadata not found")
	}
	return nil
}

