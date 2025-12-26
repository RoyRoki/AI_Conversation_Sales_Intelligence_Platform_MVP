package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"ai-conversation-platform/internal/models"
)

// MemoryStorage handles customer memory-related database operations
type MemoryStorage struct {
	client *Client
}

// NewMemoryStorage creates a new memory storage instance
func NewMemoryStorage(client *Client) *MemoryStorage {
	return &MemoryStorage{client: client}
}

// CreateMemory creates a new customer memory record
func (s *MemoryStorage) CreateMemory(tenantID string, memory *models.CustomerMemory) error {
	productInterestsJSON, _ := json.Marshal(memory.ProductInterests)
	pastObjectionsJSON, _ := json.Marshal(memory.PastObjections)

	query := `
		INSERT INTO customer_memory (id, tenant_id, customer_id, preferred_language, pricing_sensitivity, product_interests, past_objections, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.client.DB.Exec(query,
		memory.ID, tenantID, memory.CustomerID, memory.PreferredLanguage,
		memory.PricingSensitivity, string(productInterestsJSON), string(pastObjectionsJSON),
		memory.CreatedAt, memory.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}
	return nil
}

// GetMemory retrieves customer memory by customer ID (tenant-scoped)
func (s *MemoryStorage) GetMemory(tenantID, customerID string) (*models.CustomerMemory, error) {
	query := `
		SELECT id, tenant_id, customer_id, preferred_language, pricing_sensitivity, product_interests, past_objections, created_at, updated_at
		FROM customer_memory
		WHERE customer_id = $1 AND tenant_id = $2
	`
	memory := &models.CustomerMemory{}
	var productInterestsJSON, pastObjectionsJSON string

	err := s.client.DB.QueryRow(query, customerID, tenantID).Scan(
		&memory.ID, &memory.TenantID, &memory.CustomerID, &memory.PreferredLanguage,
		&memory.PricingSensitivity, &productInterestsJSON, &pastObjectionsJSON,
		&memory.CreatedAt, &memory.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("memory not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	if err := json.Unmarshal([]byte(productInterestsJSON), &memory.ProductInterests); err != nil {
		memory.ProductInterests = []string{}
	}
	if err := json.Unmarshal([]byte(pastObjectionsJSON), &memory.PastObjections); err != nil {
		memory.PastObjections = []string{}
	}

	return memory, nil
}

// UpdateMemory updates customer memory (tenant-scoped)
func (s *MemoryStorage) UpdateMemory(tenantID string, memory *models.CustomerMemory) error {
	productInterestsJSON, _ := json.Marshal(memory.ProductInterests)
	pastObjectionsJSON, _ := json.Marshal(memory.PastObjections)

	query := `
		UPDATE customer_memory
		SET preferred_language = $1, pricing_sensitivity = $2, product_interests = $3, past_objections = $4, updated_at = $5
		WHERE customer_id = $6 AND tenant_id = $7
	`
	result, err := s.client.DB.Exec(query,
		memory.PreferredLanguage, memory.PricingSensitivity,
		string(productInterestsJSON), string(pastObjectionsJSON),
		memory.UpdatedAt, memory.CustomerID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

// ListMemories lists customer memories for a tenant with pagination
func (s *MemoryStorage) ListMemories(tenantID string, limit, offset int) ([]*models.CustomerMemory, error) {
	query := `
		SELECT id, tenant_id, customer_id, preferred_language, pricing_sensitivity, product_interests, past_objections, created_at, updated_at
		FROM customer_memory
		WHERE tenant_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.client.DB.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	var memories []*models.CustomerMemory
	for rows.Next() {
		memory := &models.CustomerMemory{}
		var productInterestsJSON, pastObjectionsJSON string

		err := rows.Scan(
			&memory.ID, &memory.TenantID, &memory.CustomerID, &memory.PreferredLanguage,
			&memory.PricingSensitivity, &productInterestsJSON, &pastObjectionsJSON,
			&memory.CreatedAt, &memory.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}

		if err := json.Unmarshal([]byte(productInterestsJSON), &memory.ProductInterests); err != nil {
			memory.ProductInterests = []string{}
		}
		if err := json.Unmarshal([]byte(pastObjectionsJSON), &memory.PastObjections); err != nil {
			memory.PastObjections = []string{}
		}

		memories = append(memories, memory)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memories: %w", err)
	}
	return memories, nil
}

// GetMemoryByID retrieves customer memory by memory ID (tenant-scoped)
func (s *MemoryStorage) GetMemoryByID(tenantID, memoryID string) (*models.CustomerMemory, error) {
	query := `
		SELECT id, tenant_id, customer_id, preferred_language, pricing_sensitivity, product_interests, past_objections, created_at, updated_at
		FROM customer_memory
		WHERE id = $1 AND tenant_id = $2
	`
	memory := &models.CustomerMemory{}
	var productInterestsJSON, pastObjectionsJSON string

	err := s.client.DB.QueryRow(query, memoryID, tenantID).Scan(
		&memory.ID, &memory.TenantID, &memory.CustomerID, &memory.PreferredLanguage,
		&memory.PricingSensitivity, &productInterestsJSON, &pastObjectionsJSON,
		&memory.CreatedAt, &memory.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("memory not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	if err := json.Unmarshal([]byte(productInterestsJSON), &memory.ProductInterests); err != nil {
		memory.ProductInterests = []string{}
	}
	if err := json.Unmarshal([]byte(pastObjectionsJSON), &memory.PastObjections); err != nil {
		memory.PastObjections = []string{}
	}

	return memory, nil
}

// DeleteMemory deletes a customer memory by memory ID (tenant-scoped)
func (s *MemoryStorage) DeleteMemory(tenantID, memoryID string) error {
	query := `
		DELETE FROM customer_memory
		WHERE id = $1 AND tenant_id = $2
	`
	result, err := s.client.DB.Exec(query, memoryID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

