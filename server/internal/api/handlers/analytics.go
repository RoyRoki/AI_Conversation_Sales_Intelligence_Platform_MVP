package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-conversation-platform/internal/services/analytics"
	"ai-conversation-platform/internal/services/conversation"
	"ai-conversation-platform/internal/storage/postgres"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService   *analytics.AnalyticsService
	ingestionService   *conversation.IngestionService
	userStorage        *postgres.UserStorage
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(
	analyticsService *analytics.AnalyticsService,
	ingestionService *conversation.IngestionService,
	userStorage *postgres.UserStorage,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		ingestionService: ingestionService,
		userStorage:      userStorage,
	}
}

// GetLeadsResponse represents the response for getting leads
type GetLeadsResponse struct {
	Leads []analytics.PrioritizedLead `json:"leads"`
	Total int                          `json:"total"`
}

// GetLeads handles GET /api/analytics/leads
func (h *AnalyticsHandler) GetLeads(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	var conversationIDs []string
	conversationIDsParam := c.Query("conversation_ids")
	if conversationIDsParam != "" {
		// Parse comma-separated IDs
		conversationIDs = strings.Split(conversationIDsParam, ",")
		for i := range conversationIDs {
			conversationIDs[i] = strings.TrimSpace(conversationIDs[i])
		}
	} else {
		// Fetch all conversations for tenant
		conversations, err := h.ingestionService.ListConversations(tenantID, 1000, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, conv := range conversations {
			conversationIDs = append(conversationIDs, conv.ID)
		}
	}

	if len(conversationIDs) == 0 {
		c.JSON(http.StatusOK, GetLeadsResponse{
			Leads: []analytics.PrioritizedLead{},
			Total: 0,
		})
		return
	}

	leads, err := h.analyticsService.PrioritizeLeads(tenantID, conversationIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Populate customer emails for leads
	for i := range leads {
		// Get conversation to find customer_id
		conv, _, err := h.ingestionService.GetConversation(tenantID, leads[i].ConversationID)
		if err == nil && conv.CustomerID != nil && *conv.CustomerID != "" {
			user, err := h.userStorage.GetUser(tenantID, *conv.CustomerID)
			if err == nil {
				leads[i].CustomerEmail = &user.Email
			}
		}
	}

	c.JSON(http.StatusOK, GetLeadsResponse{
		Leads: leads,
		Total: len(leads),
	})
}

// GetWinProbabilityResponse represents the response for win probability
type GetWinProbabilityResponse struct {
	WinProbability analytics.WinProbability `json:"win_probability"`
}

// GetWinProbability handles GET /api/analytics/conversations/:id/win-probability
func (h *AnalyticsHandler) GetWinProbability(c *gin.Context) {
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

	winProb, err := h.analyticsService.CalculateWinProbability(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetWinProbabilityResponse{
		WinProbability: winProb,
	})
}

// GetChurnRiskResponse represents the response for churn risk
type GetChurnRiskResponse struct {
	ChurnRisk analytics.ChurnRisk `json:"churn_risk"`
}

// GetChurnRisk handles GET /api/analytics/conversations/:id/churn-risk
func (h *AnalyticsHandler) GetChurnRisk(c *gin.Context) {
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

	churnRisk, err := h.analyticsService.CalculateChurnRisk(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetChurnRiskResponse{
		ChurnRisk: churnRisk,
	})
}

// GetQualityResponse represents the response for quality score
type GetQualityResponse struct {
	Quality analytics.QualityScore `json:"quality"`
}

// GetQuality handles GET /api/analytics/conversations/:id/quality (Admin only)
func (h *AnalyticsHandler) GetQuality(c *gin.Context) {
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

	// Check if user is admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	quality, err := h.analyticsService.CalculateQualityScore(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetQualityResponse{
		Quality: quality,
	})
}

// GetTrendsResponse represents the response for trends
type GetTrendsResponse struct {
	Trends analytics.TrendAnalysis `json:"trends"`
}

// GetTrends handles GET /api/analytics/conversations/:id/trends
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
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

	// Get trends through analytics service
	// Note: This requires accessing messages and metadata, so we'll need to add a method to analytics service
	// For now, we'll calculate it inline
	trends, err := h.analyticsService.GetTrends(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetTrendsResponse{
		Trends: trends,
	})
}

// GetCLVResponse represents the response for CLV
type GetCLVResponse struct {
	CLV analytics.CLVEstimate `json:"clv"`
}

// GetCLV handles GET /api/analytics/conversations/:id/clv
func (h *AnalyticsHandler) GetCLV(c *gin.Context) {
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

	clv, err := h.analyticsService.CalculateCLV(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetCLVResponse{
		CLV: clv,
	})
}

// GetSalesCycleResponse represents the response for sales cycle
type GetSalesCycleResponse struct {
	SalesCycle analytics.SalesCyclePrediction `json:"sales_cycle"`
}

// GetSalesCycle handles GET /api/analytics/conversations/:id/sales-cycle
func (h *AnalyticsHandler) GetSalesCycle(c *gin.Context) {
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

	salesCycle, err := h.analyticsService.PredictSalesCycle(tenantID, conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetSalesCycleResponse{
		SalesCycle: salesCycle,
	})
}

// GetDashboardResponse represents the response for dashboard
type GetDashboardResponse struct {
	Metrics analytics.DashboardMetrics `json:"metrics"`
}

// GetDashboard handles GET /api/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	// Get dashboard metrics
	metrics, err := h.analyticsService.GetDashboardMetrics(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetDashboardResponse{
		Metrics: metrics,
	})
}

