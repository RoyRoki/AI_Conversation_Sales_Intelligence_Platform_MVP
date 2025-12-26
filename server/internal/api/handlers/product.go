package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ai-conversation-platform/internal/ai"
	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productStorage   *postgres.ProductStorage
	embeddingService *ai.EmbeddingService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productStorage *postgres.ProductStorage, embeddingService *ai.EmbeddingService) *ProductHandler {
	return &ProductHandler{
		productStorage:   productStorage,
		embeddingService: embeddingService,
	}
}

// buildProductText creates a comprehensive text representation of a product for embedding
func buildProductText(product *models.Product) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Product: %s", product.Name))
	parts = append(parts, fmt.Sprintf("Description: %s", product.Description))

	if product.Category != "" {
		parts = append(parts, fmt.Sprintf("Category: %s", product.Category))
	}

	if product.Price > 0 {
		parts = append(parts, fmt.Sprintf("Price: %s %.2f", product.PriceCurrency, product.Price))
	}

	if len(product.Features) > 0 {
		parts = append(parts, fmt.Sprintf("Features: %s", strings.Join(product.Features, ", ")))
	}

	if len(product.Limitations) > 0 {
		parts = append(parts, fmt.Sprintf("Limitations: %s", strings.Join(product.Limitations, ", ")))
	}

	if product.TargetAudience != "" {
		parts = append(parts, fmt.Sprintf("Target Audience: %s", product.TargetAudience))
	}

	if len(product.CommonQuestions) > 0 {
		parts = append(parts, fmt.Sprintf("Common Questions: %s", strings.Join(product.CommonQuestions, ", ")))
	}

	return strings.Join(parts, "\n")
}

// embedProduct embeds a product into Chroma DB for semantic search
func (h *ProductHandler) embedProduct(product *models.Product) {
	if h.embeddingService == nil {
		return // Embedding service not available
	}

	productText := buildProductText(product)
	collection := "product_knowledge"
	metadata := map[string]interface{}{
		"id":         product.ID,
		"tenant_id":  product.TenantID,
		"product_id": product.ID,
		"name":       product.Name,
		"category":   product.Category,
	}

	if err := h.embeddingService.EmbedAndStore(
		collection,
		productText,
		ai.ContentTypeProductKnowledge,
		metadata,
	); err != nil {
		log.Printf("[ProductHandler] failed to embed product %s: %v", product.ID, err)
		// Don't fail the request if embedding fails
	} else {
		log.Printf("[ProductHandler] successfully embedded product %s", product.ID)
	}
}

// ListProductsResponse represents the response for listing products
type ListProductsResponse struct {
	Products []*models.Product `json:"products"`
}

// ListProducts handles GET /api/products
// Requires authentication - tenant_id is extracted from JWT token
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Get tenant ID from context (set by JWT middleware)
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	role := c.GetString("role")
	
	log.Printf("[ProductHandler] ListProducts - tenantID from context: %s, userID: %s, role: %s", tenantID, userID, role)
	
	if tenantID == "" {
		// Fallback to query param for backward compatibility
		tenantID = c.Query("tenant_id")
		if tenantID == "" {
			log.Printf("[ProductHandler] tenant_id missing from context and query params - userID: %s, role: %s", userID, role)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "tenant_id is required. Please log out and log in again to refresh your token.",
			})
			return
		}
		log.Printf("[ProductHandler] Using tenant_id from query param: %s", tenantID)
	}

	products, err := h.productStorage.ListProducts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListProductsResponse{Products: products})
}

// GetProductResponse represents the response for getting a product
type GetProductResponse struct {
	Product *models.Product `json:"product"`
}

// GetProduct handles GET /api/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
			return
		}
	}

	product, err := h.productStorage.GetProduct(tenantID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetProductResponse{Product: product})
}

// CreateProductRequest represents the request body for creating a product
type CreateProductRequest struct {
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	Category        string   `json:"category"`
	Price           float64  `json:"price" binding:"required"`
	PriceCurrency   string   `json:"price_currency"`
	Features        []string `json:"features"`
	Limitations     []string `json:"limitations"`
	TargetAudience  string   `json:"target_audience"`
	CommonQuestions []string `json:"common_questions"`
}

// CreateProductResponse represents the response for creating a product
type CreateProductResponse struct {
	Product *models.Product `json:"product"`
}

// CreateProduct handles POST /api/products (admin only)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Check if user is admin
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

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	product := &models.Product{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		Price:           req.Price,
		PriceCurrency:   req.PriceCurrency,
		Features:        req.Features,
		Limitations:     req.Limitations,
		TargetAudience:  req.TargetAudience,
		CommonQuestions: req.CommonQuestions,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if product.PriceCurrency == "" {
		product.PriceCurrency = "INR"
	}

	if err := h.productStorage.CreateProduct(tenantID, product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Embed product into Chroma DB for semantic search (async, non-blocking)
	go h.embedProduct(product)

	c.JSON(http.StatusCreated, CreateProductResponse{Product: product})
}

// UpdateProductRequest represents the request body for updating a product
type UpdateProductRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	Price           float64  `json:"price"`
	PriceCurrency   string   `json:"price_currency"`
	Features        []string `json:"features"`
	Limitations     []string `json:"limitations"`
	TargetAudience  string   `json:"target_audience"`
	CommonQuestions []string `json:"common_questions"`
}

// UpdateProduct handles PUT /api/products/:id (admin only)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Check if user is admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	// Get existing product
	existingProduct, err := h.productStorage.GetProduct(tenantID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.Name != "" {
		existingProduct.Name = req.Name
	}
	if req.Description != "" {
		existingProduct.Description = req.Description
	}
	if req.Category != "" {
		existingProduct.Category = req.Category
	}
	if req.Price > 0 {
		existingProduct.Price = req.Price
	}
	if req.PriceCurrency != "" {
		existingProduct.PriceCurrency = req.PriceCurrency
	}
	if req.Features != nil {
		existingProduct.Features = req.Features
	}
	if req.Limitations != nil {
		existingProduct.Limitations = req.Limitations
	}
	if req.TargetAudience != "" {
		existingProduct.TargetAudience = req.TargetAudience
	}
	if req.CommonQuestions != nil {
		existingProduct.CommonQuestions = req.CommonQuestions
	}
	existingProduct.UpdatedAt = time.Now()

	if err := h.productStorage.UpdateProduct(tenantID, existingProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-embed updated product into Chroma DB for semantic search (async, non-blocking)
	go h.embedProduct(existingProduct)

	c.JSON(http.StatusOK, GetProductResponse{Product: existingProduct})
}

// DeleteProduct handles DELETE /api/products/:id (admin only)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Check if user is admin
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}

	if err := h.productStorage.DeleteProduct(tenantID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

