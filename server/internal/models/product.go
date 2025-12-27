package models

import (
	"time"
)

// Product represents a product in the system
type Product struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	Price           float64   `json:"price"`
	PriceCurrency   string    `json:"price_currency"` // "INR", "USD", etc.
	Features        []string  `json:"features"`       // JSON array
	Limitations     []string  `json:"limitations"`   // JSON array
	TargetAudience  string    `json:"target_audience"`
	CommonQuestions []string  `json:"common_questions"` // JSON array
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}


