package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/services/autoreply"
	"ai-conversation-platform/internal/storage/postgres"
)

// AutoReplyHandler handles auto-reply configuration HTTP requests
type AutoReplyHandler struct {
	autoReplyService *autoreply.AutoReplyService
	globalConfigStorage *postgres.AutoReplyStorage
	conversationConfigStorage *postgres.AutoReplyStorage
}

// NewAutoReplyHandler creates a new auto-reply handler
func NewAutoReplyHandler(
	autoReplyService *autoreply.AutoReplyService,
	globalConfigStorage *postgres.AutoReplyStorage,
	conversationConfigStorage *postgres.AutoReplyStorage,
) *AutoReplyHandler {
	return &AutoReplyHandler{
		autoReplyService: autoReplyService,
		globalConfigStorage: globalConfigStorage,
		conversationConfigStorage: conversationConfigStorage,
	}
}

// GetGlobalAutoReplyResponse represents the response for getting global auto-reply config
type GetGlobalAutoReplyResponse struct {
	Config *models.AutoReplyGlobalConfig `json:"config"`
}

// GetGlobalAutoReply handles GET /api/autoreply/global (admin only)
func (h *AutoReplyHandler) GetGlobalAutoReply(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	config, err := h.globalConfigStorage.GetGlobalConfig(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetGlobalAutoReplyResponse{Config: config})
}

// UpdateGlobalAutoReplyRequest represents the request body for updating global auto-reply config
type UpdateGlobalAutoReplyRequest struct {
	Enabled            bool    `json:"enabled"`
	ConfidenceThreshold float64 `json:"confidence_threshold"` // 0.0 - 1.0
}

// UpdateGlobalAutoReply handles PUT /api/autoreply/global (admin only)
func (h *AutoReplyHandler) UpdateGlobalAutoReply(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var req UpdateGlobalAutoReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate confidence threshold
	if req.ConfidenceThreshold < 0.0 || req.ConfidenceThreshold > 1.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confidence_threshold must be between 0.0 and 1.0"})
		return
	}

	config := &models.AutoReplyGlobalConfig{
		TenantID:           tenantID,
		Enabled:            req.Enabled,
		ConfidenceThreshold: req.ConfidenceThreshold,
		UpdatedAt:          time.Now(),
	}

	if err := h.globalConfigStorage.UpdateGlobalConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetGlobalAutoReplyResponse{Config: config})
}

// GetConversationAutoReplyResponse represents the response for getting conversation auto-reply config
type GetConversationAutoReplyResponse struct {
	Config *models.AutoReplyConversationConfig `json:"config"`
	Effective *autoreply.EffectiveConfig       `json:"effective"`
}

// GetConversationAutoReply handles GET /api/conversations/:id/autoreply
func (h *AutoReplyHandler) GetConversationAutoReply(c *gin.Context) {
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

	// Get effective config
	effective, err := h.autoReplyService.CheckAutoReplyEnabled(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Try to get conversation-specific config
	var convConfig *models.AutoReplyConversationConfig
	config, err := h.conversationConfigStorage.GetConversationConfig(conversationID)
	if err == nil {
		convConfig = config
	}

	c.JSON(http.StatusOK, GetConversationAutoReplyResponse{
		Config: convConfig,
		Effective: effective,
	})
}

// UpdateConversationAutoReplyRequest represents the request body for updating conversation auto-reply config
type UpdateConversationAutoReplyRequest struct {
	Enabled            bool     `json:"enabled"`
	ConfidenceThreshold *float64 `json:"confidence_threshold,omitempty"` // Optional override
}

// UpdateConversationAutoReply handles PUT /api/conversations/:id/autoreply (agent/admin)
func (h *AutoReplyHandler) UpdateConversationAutoReply(c *gin.Context) {
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

	var req UpdateConversationAutoReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate confidence threshold if provided
	if req.ConfidenceThreshold != nil {
		if *req.ConfidenceThreshold < 0.0 || *req.ConfidenceThreshold > 1.0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "confidence_threshold must be between 0.0 and 1.0"})
			return
		}
	}

	config := &models.AutoReplyConversationConfig{
		ConversationID:     conversationID,
		Enabled:            req.Enabled,
		ConfidenceThreshold: req.ConfidenceThreshold,
		UpdatedAt:          time.Now(),
	}

	if err := h.conversationConfigStorage.UpdateConversationConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get effective config
	effective, err := h.autoReplyService.CheckAutoReplyEnabled(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetConversationAutoReplyResponse{
		Config: config,
		Effective: effective,
	})
}

// TestAutoReplyResponse represents the response for testing auto-reply
type TestAutoReplyResponse struct {
	WouldSend    bool    `json:"would_send"`
	Confidence   float64 `json:"confidence"`
	Suggestion   string  `json:"suggestion"`
	Reason       string  `json:"reason"`
}

// TestAutoReply handles POST /api/conversations/:id/autoreply/test (admin)
func (h *AutoReplyHandler) TestAutoReply(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

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

	// Check effective config
	effective, err := h.autoReplyService.CheckAutoReplyEnabled(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !effective.Enabled {
		c.JSON(http.StatusOK, TestAutoReplyResponse{
			WouldSend: false,
			Reason:    "Auto-reply is disabled",
		})
		return
	}

	// This would check if a suggestion would be sent
	// For now, return a placeholder
	c.JSON(http.StatusOK, TestAutoReplyResponse{
		WouldSend:  false,
		Confidence: 0.0,
		Reason:     "Test functionality to be implemented",
	})
}

