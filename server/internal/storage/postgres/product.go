package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"ai-conversation-platform/internal/models"
)

// ProductStorage handles product-related database operations
type ProductStorage struct {
	client *Client
}

// NewProductStorage creates a new product storage instance
func NewProductStorage(client *Client) *ProductStorage {
	return &ProductStorage{client: client}
}

// CreateProduct creates a new product
func (s *ProductStorage) CreateProduct(tenantID string, product *models.Product) error {
	featuresJSON, _ := json.Marshal(product.Features)
	limitationsJSON, _ := json.Marshal(product.Limitations)
	commonQuestionsJSON, _ := json.Marshal(product.CommonQuestions)

	query := `
		INSERT INTO products (id, tenant_id, name, description, category, price, price_currency, features, limitations, target_audience, common_questions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := s.client.DB.Exec(query,
		product.ID, tenantID, product.Name, product.Description, product.Category,
		product.Price, product.PriceCurrency, string(featuresJSON), string(limitationsJSON),
		product.TargetAudience, string(commonQuestionsJSON), product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

// GetProduct retrieves a product by ID (tenant-scoped)
func (s *ProductStorage) GetProduct(tenantID, productID string) (*models.Product, error) {
	query := `
		SELECT id, tenant_id, name, description, category, price, price_currency, features, limitations, target_audience, common_questions, created_at, updated_at
		FROM products
		WHERE id = $1 AND tenant_id = $2
	`
	product := &models.Product{}
	var featuresJSON, limitationsJSON, commonQuestionsJSON string

	err := s.client.DB.QueryRow(query, productID, tenantID).Scan(
		&product.ID, &product.TenantID, &product.Name, &product.Description, &product.Category,
		&product.Price, &product.PriceCurrency, &featuresJSON, &limitationsJSON,
		&product.TargetAudience, &commonQuestionsJSON, &product.CreatedAt, &product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Unmarshal JSON arrays
	if err := json.Unmarshal([]byte(featuresJSON), &product.Features); err != nil {
		product.Features = []string{}
	}
	if err := json.Unmarshal([]byte(limitationsJSON), &product.Limitations); err != nil {
		product.Limitations = []string{}
	}
	if err := json.Unmarshal([]byte(commonQuestionsJSON), &product.CommonQuestions); err != nil {
		product.CommonQuestions = []string{}
	}

	return product, nil
}

// ListProducts lists all products for a tenant
func (s *ProductStorage) ListProducts(tenantID string) ([]*models.Product, error) {
	query := `
		SELECT id, tenant_id, name, description, category, price, price_currency, features, limitations, target_audience, common_questions, created_at, updated_at
		FROM products
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.client.DB.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var featuresJSON, limitationsJSON, commonQuestionsJSON string

		err := rows.Scan(
			&product.ID, &product.TenantID, &product.Name, &product.Description, &product.Category,
			&product.Price, &product.PriceCurrency, &featuresJSON, &limitationsJSON,
			&product.TargetAudience, &commonQuestionsJSON, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		// Unmarshal JSON arrays
		if err := json.Unmarshal([]byte(featuresJSON), &product.Features); err != nil {
			product.Features = []string{}
		}
		if err := json.Unmarshal([]byte(limitationsJSON), &product.Limitations); err != nil {
			product.Limitations = []string{}
		}
		if err := json.Unmarshal([]byte(commonQuestionsJSON), &product.CommonQuestions); err != nil {
			product.CommonQuestions = []string{}
		}

		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// UpdateProduct updates a product (tenant-scoped)
func (s *ProductStorage) UpdateProduct(tenantID string, product *models.Product) error {
	featuresJSON, _ := json.Marshal(product.Features)
	limitationsJSON, _ := json.Marshal(product.Limitations)
	commonQuestionsJSON, _ := json.Marshal(product.CommonQuestions)

	query := `
		UPDATE products
		SET name = $1, description = $2, category = $3, price = $4, price_currency = $5,
		    features = $6, limitations = $7, target_audience = $8, common_questions = $9, updated_at = $10
		WHERE id = $11 AND tenant_id = $12
	`
	result, err := s.client.DB.Exec(query,
		product.Name, product.Description, product.Category, product.Price, product.PriceCurrency,
		string(featuresJSON), string(limitationsJSON), product.TargetAudience, string(commonQuestionsJSON),
		product.UpdatedAt, product.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// DeleteProduct deletes a product (tenant-scoped)
func (s *ProductStorage) DeleteProduct(tenantID, productID string) error {
	query := `
		DELETE FROM products
		WHERE id = $1 AND tenant_id = $2
	`
	result, err := s.client.DB.Exec(query, productID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

