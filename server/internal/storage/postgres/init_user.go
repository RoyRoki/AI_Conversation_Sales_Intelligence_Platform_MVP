package postgres

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	"ai-conversation-platform/internal/auth"
	"ai-conversation-platform/internal/models"
)

// InitDefaultAdmin creates a default admin user if it doesn't exist
func (s *UserStorage) InitDefaultAdmin() error {
	// Get default admin credentials from environment variables
	tenantID := os.Getenv("DEFAULT_ADMIN_TENANT_ID")
	email := os.Getenv("DEFAULT_ADMIN_EMAIL")
	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")

	// If not set, use defaults
	if tenantID == "" {
		tenantID = "OMX26"
	}
	if email == "" {
		email = "OMX2026@gmail.com"
	}
	if password == "" {
		password = "OMX@2026"
	}

	// Check if user already exists
	_, err := s.GetUserByEmail(tenantID, email)
	if err == nil {
		// User already exists, skip creation
		log.Printf("Default admin user already exists: %s", email)
		return nil
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         models.RoleAdmin,
	}

	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err = s.CreateUser(tenantID, user)
	if err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	log.Printf("Default admin user created successfully: %s (tenant: %s)", email, tenantID)
	return nil
}

