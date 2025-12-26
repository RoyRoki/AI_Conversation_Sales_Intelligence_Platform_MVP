package postgres

import (
	"database/sql"
	"fmt"

	"ai-conversation-platform/internal/models"
)

// RuleStorage handles rule-related database operations
type RuleStorage struct {
	client *Client
}

// NewRuleStorage creates a new rule storage instance
func NewRuleStorage(client *Client) *RuleStorage {
	return &RuleStorage{client: client}
}

// CreateRule creates a new rule
func (s *RuleStorage) CreateRule(tenantID string, rule *models.Rule) error {
	query := `
		INSERT INTO rules (id, tenant_id, name, description, type, pattern, action, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.client.DB.Exec(query,
		rule.ID, tenantID, rule.Name, rule.Description, rule.Type,
		rule.Pattern, rule.Action, rule.IsActive, rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}
	return nil
}

// GetRule retrieves a rule by ID (tenant-scoped)
func (s *RuleStorage) GetRule(tenantID, ruleID string) (*models.Rule, error) {
	query := `
		SELECT id, tenant_id, name, description, type, pattern, action, is_active, created_at, updated_at
		FROM rules
		WHERE id = $1 AND tenant_id = $2
	`
	rule := &models.Rule{}
	err := s.client.DB.QueryRow(query, ruleID, tenantID).Scan(
		&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Type,
		&rule.Pattern, &rule.Action, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	return rule, nil
}

// ListRules lists rules for a tenant
func (s *RuleStorage) ListRules(tenantID string, activeOnly bool) ([]*models.Rule, error) {
	var query string
	var args []interface{}

	if activeOnly {
		query = `
			SELECT id, tenant_id, name, description, type, pattern, action, is_active, created_at, updated_at
			FROM rules
			WHERE tenant_id = $1 AND is_active = true
			ORDER BY created_at DESC
		`
		args = []interface{}{tenantID}
	} else {
		query = `
			SELECT id, tenant_id, name, description, type, pattern, action, is_active, created_at, updated_at
			FROM rules
			WHERE tenant_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{tenantID}
	}

	rows, err := s.client.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}
	defer rows.Close()

	var rules []*models.Rule
	for rows.Next() {
		rule := &models.Rule{}
		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Type,
			&rule.Pattern, &rule.Action, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rules: %w", err)
	}
	return rules, nil
}

// UpdateRule updates a rule (tenant-scoped)
func (s *RuleStorage) UpdateRule(tenantID string, rule *models.Rule) error {
	query := `
		UPDATE rules
		SET name = $1, description = $2, type = $3, pattern = $4, action = $5, is_active = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`
	result, err := s.client.DB.Exec(query,
		rule.Name, rule.Description, rule.Type, rule.Pattern, rule.Action,
		rule.IsActive, rule.UpdatedAt, rule.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// DeleteRule deletes a rule (tenant-scoped)
func (s *RuleStorage) DeleteRule(tenantID, ruleID string) error {
	query := `
		DELETE FROM rules
		WHERE id = $1 AND tenant_id = $2
	`
	result, err := s.client.DB.Exec(query, ruleID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// LoadRules loads active rules for a tenant (implements RuleLoader interface)
func (s *RuleStorage) LoadRules(tenantID string) ([]*models.Rule, error) {
	return s.ListRules(tenantID, true) // Load only active rules
}

