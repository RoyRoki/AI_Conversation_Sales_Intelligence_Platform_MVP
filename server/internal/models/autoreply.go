package models

import (
	"time"
)

// AutoReplyGlobalConfig represents global auto-reply configuration for a tenant
type AutoReplyGlobalConfig struct {
	TenantID           string    `json:"tenant_id"`
	Enabled            bool      `json:"enabled"`
	ConfidenceThreshold float64  `json:"confidence_threshold"` // 0.0 - 1.0
	UpdatedAt          time.Time `json:"updated_at"`
}

// AutoReplyConversationConfig represents per-conversation auto-reply configuration
type AutoReplyConversationConfig struct {
	ConversationID     string     `json:"conversation_id"`
	Enabled            bool       `json:"enabled"`
	ConfidenceThreshold *float64  `json:"confidence_threshold,omitempty"` // Optional override, nil means use global
	UpdatedAt          time.Time  `json:"updated_at"`
}


