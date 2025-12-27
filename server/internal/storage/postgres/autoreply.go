package postgres

import (
	"database/sql"
	"fmt"

	"ai-conversation-platform/internal/models"
)

// AutoReplyStorage handles auto-reply configuration database operations
type AutoReplyStorage struct {
	client *Client
}

// NewAutoReplyStorage creates a new auto-reply storage instance
func NewAutoReplyStorage(client *Client) *AutoReplyStorage {
	return &AutoReplyStorage{client: client}
}

// GetGlobalConfig retrieves global auto-reply configuration for a tenant
func (s *AutoReplyStorage) GetGlobalConfig(tenantID string) (*models.AutoReplyGlobalConfig, error) {
	query := `
		SELECT tenant_id, enabled, confidence_threshold, updated_at
		FROM auto_reply_global
		WHERE tenant_id = $1
	`
	config := &models.AutoReplyGlobalConfig{}
	err := s.client.DB.QueryRow(query, tenantID).Scan(
		&config.TenantID, &config.Enabled, &config.ConfidenceThreshold, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Return default config if not found
		return &models.AutoReplyGlobalConfig{
			TenantID:           tenantID,
			Enabled:            false,
			ConfidenceThreshold: 0.8,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}
	return config, nil
}

// UpdateGlobalConfig updates or creates global auto-reply configuration
func (s *AutoReplyStorage) UpdateGlobalConfig(config *models.AutoReplyGlobalConfig) error {
	if s.client.DBType == "sqlite" {
		query := `
			INSERT OR REPLACE INTO auto_reply_global (tenant_id, enabled, confidence_threshold, updated_at)
			VALUES ($1, $2, $3, $4)
		`
		_, err := s.client.DB.Exec(query, config.TenantID, config.Enabled, config.ConfidenceThreshold, config.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to update global config: %w", err)
		}
		return nil
	}

	query := `
		INSERT INTO auto_reply_global (tenant_id, enabled, confidence_threshold, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(tenant_id) DO UPDATE SET
			enabled = excluded.enabled,
			confidence_threshold = excluded.confidence_threshold,
			updated_at = excluded.updated_at
	`
	_, err := s.client.DB.Exec(query, config.TenantID, config.Enabled, config.ConfidenceThreshold, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update global config: %w", err)
	}
	return nil
}

// GetConversationConfig retrieves per-conversation auto-reply configuration
func (s *AutoReplyStorage) GetConversationConfig(conversationID string) (*models.AutoReplyConversationConfig, error) {
	query := `
		SELECT conversation_id, enabled, confidence_threshold, updated_at
		FROM auto_reply_conversations
		WHERE conversation_id = $1
	`
	config := &models.AutoReplyConversationConfig{}
	var confidenceThreshold sql.NullFloat64
	err := s.client.DB.QueryRow(query, conversationID).Scan(
		&config.ConversationID, &config.Enabled, &confidenceThreshold, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation config: %w", err)
	}
	if confidenceThreshold.Valid {
		config.ConfidenceThreshold = &confidenceThreshold.Float64
	}
	return config, nil
}

// UpdateConversationConfig updates or creates per-conversation auto-reply configuration
func (s *AutoReplyStorage) UpdateConversationConfig(config *models.AutoReplyConversationConfig) error {
	if s.client.DBType == "sqlite" {
		query := `
			INSERT OR REPLACE INTO auto_reply_conversations (conversation_id, enabled, confidence_threshold, updated_at)
			VALUES ($1, $2, $3, $4)
		`
		var confidenceThreshold interface{}
		if config.ConfidenceThreshold != nil {
			confidenceThreshold = *config.ConfidenceThreshold
		}
		_, err := s.client.DB.Exec(query, config.ConversationID, config.Enabled, confidenceThreshold, config.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to update conversation config: %w", err)
		}
		return nil
	}

	query := `
		INSERT INTO auto_reply_conversations (conversation_id, enabled, confidence_threshold, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(conversation_id) DO UPDATE SET
			enabled = excluded.enabled,
			confidence_threshold = excluded.confidence_threshold,
			updated_at = excluded.updated_at
	`
	var confidenceThreshold interface{}
	if config.ConfidenceThreshold != nil {
		confidenceThreshold = *config.ConfidenceThreshold
	}
	_, err := s.client.DB.Exec(query, config.ConversationID, config.Enabled, confidenceThreshold, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update conversation config: %w", err)
	}
	return nil
}


