package models

import (
	"time"
)

// CustomerMemory stores customer preferences and history
type CustomerMemory struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	CustomerID        string    `json:"customer_id"`
	PreferredLanguage string    `json:"preferred_language"`
	PricingSensitivity string   `json:"pricing_sensitivity"` // "high", "medium", "low"
	ProductInterests  []string  `json:"product_interests"`
	PastObjections    []string  `json:"past_objections"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

