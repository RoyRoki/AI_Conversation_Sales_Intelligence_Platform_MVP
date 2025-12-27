package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// RuleHandler handles rule-related HTTP requests
type RuleHandler struct {
	ruleStorage *postgres.RuleStorage
}

// NewRuleHandler creates a new rule handler
func NewRuleHandler(ruleStorage *postgres.RuleStorage) *RuleHandler {
	return &RuleHandler{
		ruleStorage: ruleStorage,
	}
}

// ListRulesRequest represents query parameters for listing rules
type ListRulesRequest struct {
	ActiveOnly bool `form:"active_only"`
}

// ListRulesResponse represents the response for listing rules
type ListRulesResponse struct {
	Rules []*models.Rule `json:"rules"`
	Total int             `json:"total"`
}

// ListRules handles GET /api/rules
func (h *RuleHandler) ListRules(c *gin.Context) {
	var req ListRulesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	rules, err := h.ruleStorage.ListRules(tenantID, req.ActiveOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListRulesResponse{
		Rules: rules,
		Total: len(rules),
	})
}

// GetRuleResponse represents the response for getting a rule
type GetRuleResponse struct {
	Rule *models.Rule `json:"rule"`
}

// GetRule handles GET /api/rules/:id
func (h *RuleHandler) GetRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	rule, err := h.ruleStorage.GetRule(tenantID, ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetRuleResponse{Rule: rule})
}

// CreateRuleRequest represents the request body for creating a rule
type CreateRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"` // "block", "correct", "flag"
	Pattern     string `json:"pattern" binding:"required"`
	Action      string `json:"action" binding:"required"` // "block", "auto_correct", "flag"
	IsActive    bool   `json:"is_active"`
}

// CreateRuleResponse represents the response for creating a rule
type CreateRuleResponse struct {
	Rule *models.Rule `json:"rule"`
}

// CreateRule handles POST /api/rules
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	now := time.Now()
	rule := &models.Rule{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Pattern:     req.Pattern,
		Action:      req.Action,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.ruleStorage.CreateRule(tenantID, rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateRuleResponse{Rule: rule})
}

// UpdateRuleRequest represents the request body for updating a rule
type UpdateRuleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Pattern     string `json:"pattern"`
	Action      string `json:"action"`
	IsActive    *bool  `json:"is_active"`
}

// UpdateRuleResponse represents the response for updating a rule
type UpdateRuleResponse struct {
	Rule *models.Rule `json:"rule"`
}

// UpdateRule handles PUT /api/rules/:id
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	// Get existing rule
	existingRule, err := h.ruleStorage.GetRule(tenantID, ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		existingRule.Name = req.Name
	}
	if req.Description != "" {
		existingRule.Description = req.Description
	}
	if req.Type != "" {
		existingRule.Type = req.Type
	}
	if req.Pattern != "" {
		existingRule.Pattern = req.Pattern
	}
	if req.Action != "" {
		existingRule.Action = req.Action
	}
	if req.IsActive != nil {
		existingRule.IsActive = *req.IsActive
	}
	existingRule.UpdatedAt = time.Now()

	if err := h.ruleStorage.UpdateRule(tenantID, existingRule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UpdateRuleResponse{Rule: existingRule})
}

// DeleteRuleResponse represents the response for deleting a rule
type DeleteRuleResponse struct {
	Message string `json:"message"`
}

// DeleteRule handles DELETE /api/rules/:id
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	if err := h.ruleStorage.DeleteRule(tenantID, ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeleteRuleResponse{Message: "Rule deleted successfully"})
}


