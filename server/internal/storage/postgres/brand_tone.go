package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

// BrandToneStorage handles brand tone configuration storage
type BrandToneStorage struct {
	client *Client
}

// NewBrandToneStorage creates a new brand tone storage instance
func NewBrandToneStorage(client *Client) *BrandToneStorage {
	return &BrandToneStorage{client: client}
}

// GetBrandTone retrieves brand tone for a tenant
func (s *BrandToneStorage) GetBrandTone(tenantID string) (string, error) {
	query := `
		SELECT tone
		FROM brand_tone
		WHERE tenant_id = $1
		LIMIT 1
	`
	var tone string
	err := s.client.DB.QueryRow(query, tenantID).Scan(&tone)
	if err == sql.ErrNoRows {
		// Return default tone if not configured
		return "Professional", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get brand tone: %w", err)
	}
	return tone, nil
}

// SetBrandTone sets brand tone for a tenant
func (s *BrandToneStorage) SetBrandTone(tenantID, tone string) error {
	// Validate tone
	validTones := map[string]bool{
		"Professional":   true,
		"Friendly":       true,
		"Sales-focused":  true,
	}
	if !validTones[tone] {
		return fmt.Errorf("invalid tone: %s (must be Professional, Friendly, or Sales-focused)", tone)
	}

	query := `
		INSERT INTO brand_tone (tenant_id, tone, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT(tenant_id) DO UPDATE SET
			tone = excluded.tone,
			updated_at = excluded.updated_at
	`
	_, err := s.client.DB.Exec(query, tenantID, tone, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set brand tone: %w", err)
	}
	return nil
}

// DeleteBrandTone deletes brand tone configuration for a tenant
func (s *BrandToneStorage) DeleteBrandTone(tenantID string) error {
	query := `
		DELETE FROM brand_tone
		WHERE tenant_id = $1
	`
	_, err := s.client.DB.Exec(query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete brand tone: %w", err)
	}
	return nil
}

