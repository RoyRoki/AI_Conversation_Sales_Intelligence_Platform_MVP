package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-conversation-platform/internal/services/agentassist"
)

// AgentAssistHandler handles agent assist-related HTTP requests
type AgentAssistHandler struct {
	agentAssistService *agentassist.AgentAssistService
}

// NewAgentAssistHandler creates a new agent assist handler
func NewAgentAssistHandler(agentAssistService *agentassist.AgentAssistService) *AgentAssistHandler {
	return &AgentAssistHandler{
		agentAssistService: agentAssistService,
	}
}

// GetSuggestionsRequest represents the request for getting suggestions
type GetSuggestionsRequest struct {
	// No body needed, conversation ID is in path
}

// GetSuggestionsResponse represents the response for getting suggestions
type GetSuggestionsResponse struct {
	Suggestions *agentassist.SuggestionsResponse `json:"suggestions"`
}

// GetSuggestions handles POST /api/conversations/:id/suggestions
func (h *AgentAssistHandler) GetSuggestions(c *gin.Context) {
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

	// Check if user is agent (role check)
	role := c.GetString("role")
	if role != "agent" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "agent access required"})
		return
	}

	// Check if service is available
	if h.agentAssistService == nil {
		log.Printf("[AGENT_ASSIST_HANDLER] agent assist service not initialized")
		c.JSON(http.StatusOK, GetSuggestionsResponse{
			Suggestions: &agentassist.SuggestionsResponse{
				Suggestions: []agentassist.Suggestion{},
				ContextUsed: false,
			},
		})
		return
	}

	suggestions, err := h.agentAssistService.GetReplySuggestions(tenantID, conversationID)
	if err != nil {
		log.Printf("[AGENT_ASSIST_HANDLER] error getting suggestions conversation=%s tenant=%s error=%v", conversationID, tenantID, err)
		// Service should now always return empty suggestions on error, but handle gracefully just in case
		// Return empty suggestions instead of error to allow graceful degradation
		c.JSON(http.StatusOK, GetSuggestionsResponse{
			Suggestions: &agentassist.SuggestionsResponse{
				Suggestions: []agentassist.Suggestion{},
				ContextUsed: false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetSuggestionsResponse{
		Suggestions: suggestions,
	})
}

// GetInsightsResponse represents the response for getting insights
type GetInsightsResponse struct {
	Insights *agentassist.SuggestionsResponse `json:"insights"`
}

// GetInsights handles GET /api/conversations/:id/insights
func (h *AgentAssistHandler) GetInsights(c *gin.Context) {
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

	// Check if user is agent (role check)
	role := c.GetString("role")
	if role != "agent" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "agent access required"})
		return
	}

	// Check if service is available
	if h.agentAssistService == nil {
		log.Printf("[AGENT_ASSIST_HANDLER] agent assist service not initialized")
		c.JSON(http.StatusOK, GetInsightsResponse{
			Insights: &agentassist.SuggestionsResponse{
				Suggestions: []agentassist.Suggestion{},
				ContextUsed: false,
			},
		})
		return
	}

	// Get insights (same as suggestions but focused on metadata)
	suggestions, err := h.agentAssistService.GetReplySuggestions(tenantID, conversationID)
	if err != nil {
		log.Printf("[AGENT_ASSIST_HANDLER] error getting insights conversation=%s tenant=%s error=%v", conversationID, tenantID, err)
		// Service should now always return empty suggestions on error, but handle gracefully just in case
		// Return empty suggestions instead of error to allow graceful degradation
		c.JSON(http.StatusOK, GetInsightsResponse{
			Insights: &agentassist.SuggestionsResponse{
				Suggestions: []agentassist.Suggestion{},
				ContextUsed: false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, GetInsightsResponse{
		Insights: suggestions,
	})
}

