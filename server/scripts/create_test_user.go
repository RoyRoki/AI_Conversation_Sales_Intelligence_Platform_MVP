package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	"ai-conversation-platform/internal/auth"
	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run create_test_user.go <tenant_id> <email> <password> [role]")
		fmt.Println("Example: go run create_test_user.go tenant1 admin@example.com password123 admin")
		os.Exit(1)
	}

	tenantID := os.Args[1]
	email := os.Args[2]
	password := os.Args[3]
	role := "admin"
	if len(os.Args) > 4 {
		role = os.Args[4]
	}

	// Initialize database client
	dbClient, err := postgres.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()

	// Initialize user storage
	userStorage := postgres.NewUserStorage(dbClient)

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         models.UserRole(role),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = userStorage.CreateUser(tenantID, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("User created successfully!\n")
	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Role: %s\n", user.Role)
	fmt.Printf("Tenant ID: %s\n", user.TenantID)
}

