package models

import (
	"time"
)

// Rule represents a policy rule configuration
type Rule struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "block", "correct", "flag"
	Pattern     string    `json:"pattern"` // regex or keyword pattern
	Action      string    `json:"action"` // "block", "auto_correct", "flag"
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}


