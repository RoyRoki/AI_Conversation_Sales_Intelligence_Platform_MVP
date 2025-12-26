package models

import (
	"time"
)

// Conversation represents a conversation between customer and agent
type Conversation struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	CustomerID   *string   `json:"customer_id,omitempty"`   // Customer user ID (null for agent-initiated)
	CustomerEmail *string  `json:"customer_email,omitempty"` // Customer email (populated in queries)
	ProductID    *string   `json:"product_id,omitempty"`     // Optional product context
	Status       string    `json:"status"`                   // active, closed, archived
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Sender         string    `json:"sender"` // "customer" | "agent"
	Content        string    `json:"content"`
	Channel        string    `json:"channel"` // "web"
	Language       string    `json:"language"`
	Timestamp      time.Time `json:"timestamp"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationMetadata stores AI analysis results separately from messages
type ConversationMetadata struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Intent         string    `json:"intent"`         // "buying", "support", "complaint"
	IntentScore    float64   `json:"intent_score"`   // 0-1
	Sentiment      string    `json:"sentiment"`      // "positive", "neutral", "negative"
	SentimentScore float64   `json:"sentiment_score"` // 0-1
	Emotions       []string  `json:"emotions"`       // ["frustration", "urgency", etc.]
	Objections     []string  `json:"objections"`     // ["price", "trust", "delivery", "competitor"]
	UpdatedAt      time.Time `json:"updated_at"`
}

