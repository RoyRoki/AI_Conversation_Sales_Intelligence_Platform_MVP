package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ai-conversation-platform/internal/models"
)

// UserStorage handles user-related database operations
type UserStorage struct {
	client *Client
}

// NewUserStorage creates a new user storage instance
func NewUserStorage(client *Client) *UserStorage {
	return &UserStorage{client: client}
}

// CreateUser creates a new user
func (s *UserStorage) CreateUser(tenantID string, user *models.User) error {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.client.DB.Exec(query,
		user.ID, tenantID, user.Email, user.PasswordHash, string(user.Role),
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID (tenant-scoped)
func (s *UserStorage) GetUser(tenantID, userID string) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1 AND tenant_id = $2
	`
	user := &models.User{}
	var roleStr string

	err := s.client.DB.QueryRow(query, userID, tenantID).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
		&roleStr, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Role = models.UserRole(roleStr)
	return user, nil
}

// GetUserByEmail retrieves a user by email (tenant-scoped)
func (s *UserStorage) GetUserByEmail(tenantID, email string) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1 AND tenant_id = $2
	`
	user := &models.User{}
	var roleStr string

	err := s.client.DB.QueryRow(query, email, tenantID).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
		&roleStr, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Role = models.UserRole(roleStr)
	return user, nil
}

// UpdateUser updates user information (tenant-scoped)
func (s *UserStorage) UpdateUser(tenantID string, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6
	`
	result, err := s.client.DB.Exec(query,
		user.Email, user.PasswordHash, string(user.Role), user.UpdatedAt,
		user.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ListUsers lists users for a tenant with pagination
func (s *UserStorage) ListUsers(tenantID string, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.client.DB.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var roleStr string

		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
			&roleStr, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Role = models.UserRole(roleStr)
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}
	return users, nil
}

// GetOrCreateCustomerByEmail gets a customer user by email or creates one if it doesn't exist
func (s *UserStorage) GetOrCreateCustomerByEmail(tenantID, email string) (*models.User, error) {
	// Try to get existing user
	user, err := s.GetUserByEmail(tenantID, email)
	if err == nil {
		// User exists, return it
		return user, nil
	}

	// User doesn't exist, create new customer user
	newUser := &models.User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: "", // No password for customer email-only login
		Role:         models.RoleCustomer,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.CreateUser(tenantID, newUser); err != nil {
		return nil, fmt.Errorf("failed to create customer user: %w", err)
	}

	return newUser, nil
}

