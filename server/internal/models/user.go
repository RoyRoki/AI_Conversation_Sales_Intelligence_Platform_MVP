package models

import (
	"time"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAgent    UserRole = "agent"
	RoleAdmin    UserRole = "admin"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	PasswordHash string `json:"-"` // Never serialize password
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

