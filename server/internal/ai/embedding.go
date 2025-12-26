package ai

import (
	"fmt"
	"log"

	"ai-conversation-platform/internal/storage/chroma"
)

// ContentType represents the type of content to embed
type ContentType string

const (
	ContentTypeProductKnowledge ContentType = "product_knowledge"
	ContentTypeConversationSummary ContentType = "conversation_summary"
	ContentTypeCustomerPreference ContentType = "customer_preference"
)

// EmbeddingService handles selective embedding strategy
type EmbeddingService struct {
	geminiClient *Client
	chromaClient *chroma.Client
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(geminiClient *Client, chromaClient *chroma.Client) *EmbeddingService {
	return &EmbeddingService{
		geminiClient: geminiClient,
		chromaClient: chromaClient,
	}
}

// ShouldEmbed determines if content should be embedded
func (s *EmbeddingService) ShouldEmbed(content string, contentType ContentType) bool {
	if content == "" {
		return false
	}

	switch contentType {
	case ContentTypeProductKnowledge:
		return true // Always embed product knowledge
	case ContentTypeConversationSummary:
		return true // Always embed summaries
	case ContentTypeCustomerPreference:
		return true // Always embed when preferences updated
	default:
		return false // Don't embed raw messages
	}
}

// GenerateEmbedding generates embedding for text using Gemini
func (s *EmbeddingService) GenerateEmbedding(text string) ([]float64, error) {
	req := GenerateEmbeddingRequest{Text: text}
	resp, err := s.geminiClient.GenerateEmbedding(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	return resp.Embedding, nil
}

// StoreEmbedding stores embedding in Chroma DB
func (s *EmbeddingService) StoreEmbedding(collection string, text string, embedding []float64, metadata map[string]interface{}) error {
	// Generate ID if not in metadata
	id := fmt.Sprintf("%s_%d", collection, len(text))
	if idFromMeta, ok := metadata["id"].(string); ok && idFromMeta != "" {
		id = idFromMeta
	}

	req := chroma.AddDocumentsRequest{
		Documents:  []string{text},
		Embeddings: [][]float64{embedding},
		Metadatas:  []map[string]interface{}{metadata},
		IDs:        []string{id},
	}

	if err := s.chromaClient.AddDocuments(collection, req); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	log.Printf("[Embedding] stored collection=%s id=%s text_length=%d", collection, id, len(text))
	return nil
}

// EmbedAndStore generates embedding and stores it
func (s *EmbeddingService) EmbedAndStore(collection string, text string, contentType ContentType, metadata map[string]interface{}) error {
	if !s.ShouldEmbed(text, contentType) {
		return nil
	}

	embedding, err := s.GenerateEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["content_type"] = string(contentType)

	return s.StoreEmbedding(collection, text, embedding, metadata)
}

