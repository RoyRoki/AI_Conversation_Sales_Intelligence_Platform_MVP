package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-conversation-platform/internal/auth"
	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userStorage *postgres.UserStorage
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userStorage *postgres.UserStorage) *AuthHandler {
	return &AuthHandler{
		userStorage: userStorage,
	}
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

// LoginResponse represents the response for login
type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userStorage.GetUserByEmail(req.TenantID, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Ensure tenant_id is present (use from request if user doesn't have it)
	tenantID := user.TenantID
	if tenantID == "" {
		tenantID = req.TenantID
	}
	if tenantID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant_id is required"})
		return
	}

	token, err := auth.GenerateToken(user.ID, tenantID, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// CustomerLoginRequest represents the request body for customer login
type CustomerLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

// CustomerLogin handles POST /api/auth/customer-login
// Email-only login that auto-creates customer users if they don't exist
func (h *AuthHandler) CustomerLogin(c *gin.Context) {
	var req CustomerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get or create customer user
	user, err := h.userStorage.GetOrCreateCustomerByEmail(req.TenantID, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate customer"})
		return
	}

	// Ensure user is a customer
	if user.Role != models.RoleCustomer {
		c.JSON(http.StatusForbidden, gin.H{"error": "this endpoint is for customers only"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.TenantID, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

