package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// MemoryHandler handles customer memory-related HTTP requests
type MemoryHandler struct {
	memoryStorage *postgres.MemoryStorage
}

// NewMemoryHandler creates a new memory handler
func NewMemoryHandler(memoryStorage *postgres.MemoryStorage) *MemoryHandler {
	return &MemoryHandler{
		memoryStorage: memoryStorage,
	}
}

// ListMemoriesRequest represents query parameters for listing memories
type ListMemoriesRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// ListMemoriesResponse represents the response for listing memories
type ListMemoriesResponse struct {
	Memories []*models.CustomerMemory `json:"memories"`
	Total    int                       `json:"total"`
}

// ListMemories handles GET /api/memories
func (h *MemoryHandler) ListMemories(c *gin.Context) {
	var req ListMemoriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// Use defaults if query params are not provided
		req.Limit = 100
		req.Offset = 0
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	memories, err := h.memoryStorage.ListMemories(tenantID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListMemoriesResponse{
		Memories: memories,
		Total:    len(memories),
	})
}

// GetMemoryResponse represents the response for getting a memory
type GetMemoryResponse struct {
	Memory *models.CustomerMemory `json:"memory"`
}

// GetMemory handles GET /api/memories/:id
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	memoryID := c.Param("id")
	if memoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memory_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	memory, err := h.memoryStorage.GetMemoryByID(tenantID, memoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetMemoryResponse{Memory: memory})
}

// CreateMemoryRequest represents the request body for creating a memory
type CreateMemoryRequest struct {
	CustomerID        string   `json:"customer_id" binding:"required"`
	PreferredLanguage string   `json:"preferred_language" binding:"required"`
	PricingSensitivity string  `json:"pricing_sensitivity" binding:"required"` // "high", "medium", "low"
	ProductInterests  []string `json:"product_interests"`
	PastObjections    []string `json:"past_objections"`
}

// CreateMemoryResponse represents the response for creating a memory
type CreateMemoryResponse struct {
	Memory *models.CustomerMemory `json:"memory"`
}

// CreateMemory handles POST /api/memories
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate pricing sensitivity
	if req.PricingSensitivity != "high" && req.PricingSensitivity != "medium" && req.PricingSensitivity != "low" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pricing_sensitivity must be 'high', 'medium', or 'low'"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	// Check if memory already exists for this customer
	existingMemory, err := h.memoryStorage.GetMemory(tenantID, req.CustomerID)
	if err == nil && existingMemory != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "memory already exists for this customer"})
		return
	}

	now := time.Now()
	memory := &models.CustomerMemory{
		ID:                uuid.New().String(),
		TenantID:          tenantID,
		CustomerID:        req.CustomerID,
		PreferredLanguage: req.PreferredLanguage,
		PricingSensitivity: req.PricingSensitivity,
		ProductInterests:  req.ProductInterests,
		PastObjections:    req.PastObjections,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.memoryStorage.CreateMemory(tenantID, memory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateMemoryResponse{Memory: memory})
}

// UpdateMemoryRequest represents the request body for updating a memory
type UpdateMemoryRequest struct {
	PreferredLanguage string   `json:"preferred_language"`
	PricingSensitivity string  `json:"pricing_sensitivity"` // "high", "medium", "low"
	ProductInterests  []string `json:"product_interests"`
	PastObjections    []string `json:"past_objections"`
}

// UpdateMemoryResponse represents the response for updating a memory
type UpdateMemoryResponse struct {
	Memory *models.CustomerMemory `json:"memory"`
}

// UpdateMemory handles PUT /api/memories/:id
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	memoryID := c.Param("id")
	if memoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memory_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	// Get existing memory
	existingMemory, err := h.memoryStorage.GetMemoryByID(tenantID, memoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate pricing sensitivity if provided
	if req.PricingSensitivity != "" {
		if req.PricingSensitivity != "high" && req.PricingSensitivity != "medium" && req.PricingSensitivity != "low" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pricing_sensitivity must be 'high', 'medium', or 'low'"})
			return
		}
		existingMemory.PricingSensitivity = req.PricingSensitivity
	}

	// Update fields if provided
	if req.PreferredLanguage != "" {
		existingMemory.PreferredLanguage = req.PreferredLanguage
	}
	if req.ProductInterests != nil {
		existingMemory.ProductInterests = req.ProductInterests
	}
	if req.PastObjections != nil {
		existingMemory.PastObjections = req.PastObjections
	}
	existingMemory.UpdatedAt = time.Now()

	if err := h.memoryStorage.UpdateMemory(tenantID, existingMemory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UpdateMemoryResponse{Memory: existingMemory})
}

// DeleteMemoryResponse represents the response for deleting a memory
type DeleteMemoryResponse struct {
	Message string `json:"message"`
}

// DeleteMemory handles DELETE /api/memories/:id
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	memoryID := c.Param("id")
	if memoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memory_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	if err := h.memoryStorage.DeleteMemory(tenantID, memoryID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeleteMemoryResponse{Message: "Memory deleted successfully"})
}

