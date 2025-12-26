package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client represents a Chroma DB REST API client
type Client struct {
	baseURL  string
	tenantID string
	httpClient *http.Client
}

// NewClient creates a new Chroma DB client
func NewClient() (*Client, error) {
	baseURL := os.Getenv("CHROMA_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		tenantID = "default" // Single tenant MVP
	}

	return &Client{
		baseURL:    baseURL,
		tenantID:   tenantID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// HealthCheck checks if Chroma DB is accessible
func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("%s/api/v2/heartbeat", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("chroma health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("chroma health check returned status %d", resp.StatusCode)
	}

	return nil
}

// getCollectionName returns tenant-scoped collection name
func (c *Client) getCollectionName(baseName string) string {
	return fmt.Sprintf("%s_%s", c.tenantID, baseName)
}

// CreateCollection creates a new collection in Chroma DB
func (c *Client) CreateCollection(name string) error {
	collectionName := c.getCollectionName(name)
	// Try v1 first, fallback to v2 if v1 is deprecated
	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)

	payload := map[string]interface{}{
		"name": collectionName,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetCollection retrieves collection information
func (c *Client) GetCollection(name string) (map[string]interface{}, error) {
	collectionName := c.getCollectionName(name)
	url := fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, collectionName)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("collection not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get collection: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// AddDocumentsRequest represents a request to add documents
type AddDocumentsRequest struct {
	Documents  []string
	Embeddings [][]float64
	Metadatas  []map[string]interface{}
	IDs        []string
}

// AddDocuments adds documents with embeddings to a collection
func (c *Client) AddDocuments(collectionName string, req AddDocumentsRequest) error {
	collection := c.getCollectionName(collectionName)
	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, collection)

	payload := map[string]interface{}{
		"documents":  req.Documents,
		"embeddings": req.Embeddings,
		"metadatas":  req.Metadatas,
		"ids":        req.IDs,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add documents: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// QueryRequest represents a query request
type QueryRequest struct {
	QueryEmbeddings [][]float64
	NResults        int
	Include         []string
}

// QueryResponse represents a query response
type QueryResponse struct {
	IDs       [][]string
	Documents [][]string
	Metadatas [][]map[string]interface{}
	Distances [][]float64
}

// Query performs a semantic search query
func (c *Client) Query(collectionName string, req QueryRequest) (*QueryResponse, error) {
	collection := c.getCollectionName(collectionName)
	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, collection)

	if req.NResults <= 0 {
		req.NResults = 10
	}

	if req.Include == nil {
		req.Include = []string{"documents", "metadatas", "distances"}
	}

	payload := map[string]interface{}{
		"query_embeddings": req.QueryEmbeddings,
		"n_results":        req.NResults,
		"include":          req.Include,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	response := &QueryResponse{}

	if ids, ok := result["ids"].([]interface{}); ok {
		response.IDs = make([][]string, len(ids))
		for i, idList := range ids {
			if idSlice, ok := idList.([]interface{}); ok {
				response.IDs[i] = make([]string, len(idSlice))
				for j, id := range idSlice {
					if idStr, ok := id.(string); ok {
						response.IDs[i][j] = idStr
					}
				}
			}
		}
	}

	if docs, ok := result["documents"].([]interface{}); ok {
		response.Documents = make([][]string, len(docs))
		for i, docList := range docs {
			if docSlice, ok := docList.([]interface{}); ok {
				response.Documents[i] = make([]string, len(docSlice))
				for j, doc := range docSlice {
					if docStr, ok := doc.(string); ok {
						response.Documents[i][j] = docStr
					}
				}
			}
		}
	}

	if metas, ok := result["metadatas"].([]interface{}); ok {
		response.Metadatas = make([][]map[string]interface{}, len(metas))
		for i, metaList := range metas {
			if metaSlice, ok := metaList.([]interface{}); ok {
				response.Metadatas[i] = make([]map[string]interface{}, len(metaSlice))
				for j, meta := range metaSlice {
					if metaMap, ok := meta.(map[string]interface{}); ok {
						response.Metadatas[i][j] = metaMap
					}
				}
			}
		}
	}

	if dists, ok := result["distances"].([]interface{}); ok {
		response.Distances = make([][]float64, len(dists))
		for i, distList := range dists {
			if distSlice, ok := distList.([]interface{}); ok {
				response.Distances[i] = make([]float64, len(distSlice))
				for j, dist := range distSlice {
					if distFloat, ok := dist.(float64); ok {
						response.Distances[i][j] = distFloat
					}
				}
			}
		}
	}

	return response, nil
}

// Delete deletes documents from a collection
func (c *Client) Delete(collectionName string, ids []string) error {
	collection := c.getCollectionName(collectionName)
	url := fmt.Sprintf("%s/api/v1/collections/%s/delete", c.baseURL, collection)

	payload := map[string]interface{}{
		"ids": ids,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete documents: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

