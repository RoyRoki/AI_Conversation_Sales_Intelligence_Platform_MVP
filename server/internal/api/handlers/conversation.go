package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/services/conversation"
	"ai-conversation-platform/internal/storage/postgres"
)

// ConversationHandler handles conversation-related HTTP requests
type ConversationHandler struct {
	ingestionService *conversation.IngestionService
	userStorage      *postgres.UserStorage
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(ingestionService *conversation.IngestionService, userStorage *postgres.UserStorage) *ConversationHandler {
	return &ConversationHandler{
		ingestionService: ingestionService,
		userStorage:      userStorage,
	}
}

// CreateConversationRequest represents the request body for creating a conversation
type CreateConversationRequest struct {
	TenantID  string  `json:"tenant_id" binding:"required"`
	ProductID *string `json:"product_id,omitempty"` // Optional product context
}

// CreateConversationResponse represents the response for creating a conversation
type CreateConversationResponse struct {
	Conversation *models.Conversation `json:"conversation"`
}

// CreateConversation handles POST /api/conversations
// Note: This endpoint is kept for backward compatibility, but conversations are now created lazily on first message
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID := c.GetString("user_id")
	userRole := c.GetString("role")
	
	var customerID *string
	if userRole == "customer" {
		customerID = &userID
	}

	conv, err := h.ingestionService.CreateConversation(req.TenantID, customerID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateConversationResponse{Conversation: conv})
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	Sender    string `json:"sender" binding:"required"` // "customer" | "agent"
	Message   string `json:"message" binding:"required"`
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"` // ISO-8601 format
}

// SendMessageResponse represents the response for sending a message
type SendMessageResponse struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Status         string `json:"status"`
}

// SendMessage handles POST /api/conversations/:id/messages
// If conversation_id is "new" or doesn't exist, creates conversation on first message
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	conversationIDParam := c.Param("id")
	
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse timestamp if provided
	var timestamp time.Time
	if req.Timestamp != "" {
		var err error
		timestamp, err = time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp format, use ISO-8601"})
			return
		}
	}

	// Get tenant ID and user ID from context (set by auth middleware)
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	userID := c.GetString("user_id")
	userRole := c.GetString("role")

	var conversationID string
	var customerID *string

	// If conversation_id is "new" or empty, create conversation on first message
	if conversationIDParam == "" || conversationIDParam == "new" {
		// Only customers can create conversations (agents view existing ones)
		if req.Sender == "customer" && userRole == "customer" {
			customerID = &userID
		}

		// Create conversation (will reuse existing active one if customerID provided)
		conv, err := h.ingestionService.CreateConversation(tenantID, customerID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create conversation: %v", err)})
			return
		}
		conversationID = conv.ID
	} else {
		conversationID = conversationIDParam
		// Verify conversation exists
		_, _, err := h.ingestionService.GetConversation(tenantID, conversationID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
			return
		}
	}

	// Normalize message
	normalized, err := conversation.NormalizeMessage(
		req.Message, req.Sender, req.Channel, timestamp, conversationID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ingest message
	messageID, err := h.ingestionService.IngestMessage(tenantID, normalized)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SendMessageResponse{
		MessageID:      messageID,
		ConversationID: conversationID,
		Status:         "sent",
	})
}

// GetConversationResponse represents the response for getting a conversation
type GetConversationResponse struct {
	Conversation *models.Conversation `json:"conversation"`
	Messages     []*models.Message     `json:"messages"`
}

// GetConversation handles GET /api/conversations/:id
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	conv, messages, err := h.ingestionService.GetConversation(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Populate customer email if customer_id exists
	if conv.CustomerID != nil && *conv.CustomerID != "" {
		user, err := h.userStorage.GetUser(tenantID, *conv.CustomerID)
		if err == nil {
			conv.CustomerEmail = &user.Email
		}
	}

	c.JSON(http.StatusOK, GetConversationResponse{
		Conversation: conv,
		Messages:     messages,
	})
}

// ListConversationsRequest represents query parameters for listing conversations
type ListConversationsRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// ListConversationsResponse represents the response for listing conversations
type ListConversationsResponse struct {
	Conversations []*models.Conversation `json:"conversations"`
	Total         int                     `json:"total"`
}

// ListConversations handles GET /api/conversations
func (h *ConversationHandler) ListConversations(c *gin.Context) {
	var req ListConversationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	conversations, err := h.ingestionService.ListConversations(tenantID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Populate customer emails
	for _, conv := range conversations {
		if conv.CustomerID != nil && *conv.CustomerID != "" {
			user, err := h.userStorage.GetUser(tenantID, *conv.CustomerID)
			if err == nil {
				conv.CustomerEmail = &user.Email
			}
		}
	}

	c.JSON(http.StatusOK, ListConversationsResponse{
		Conversations: conversations,
		Total:         len(conversations),
	})
}

