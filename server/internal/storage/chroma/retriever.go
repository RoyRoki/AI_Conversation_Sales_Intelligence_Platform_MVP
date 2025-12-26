package chroma

import (
	"fmt"
)

// RetrievedChunk represents a retrieved chunk with metadata
type RetrievedChunk struct {
	Text     string
	Score    float64
	Metadata map[string]interface{}
	ID       string
}

// Retriever handles context retrieval from Chroma DB
type Retriever struct {
	client *Client
}

// NewRetriever creates a new retriever
func NewRetriever(client *Client) *Retriever {
	return &Retriever{client: client}
}

// RetrieveContext retrieves top-k relevant chunks from a collection
func (r *Retriever) RetrieveContext(collection string, queryEmbedding []float64, topK int) ([]RetrievedChunk, error) {
	if topK <= 0 {
		topK = 10
	}

	req := QueryRequest{
		QueryEmbeddings: [][]float64{queryEmbedding},
		NResults:        topK,
		Include:         []string{"documents", "metadatas", "distances"},
	}

	resp, err := r.client.Query(collection, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}

	if len(resp.Documents) == 0 || len(resp.Documents[0]) == 0 {
		return []RetrievedChunk{}, nil
	}

	chunks := make([]RetrievedChunk, 0, len(resp.Documents[0]))
	for i := range resp.Documents[0] {
		chunk := RetrievedChunk{
			Text:  resp.Documents[0][i],
			Score: 1.0, // Default score
		}

		if len(resp.Distances) > 0 && len(resp.Distances[0]) > i {
			// Convert distance to similarity score (lower distance = higher similarity)
			distance := resp.Distances[0][i]
			chunk.Score = 1.0 / (1.0 + distance)
		}

		if len(resp.Metadatas) > 0 && len(resp.Metadatas[0]) > i {
			chunk.Metadata = resp.Metadatas[0][i]
		}

		if len(resp.IDs) > 0 && len(resp.IDs[0]) > i {
			chunk.ID = resp.IDs[0][i]
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// RetrieveRelevantConversations retrieves relevant conversation context
func (r *Retriever) RetrieveRelevantConversations(tenantID string, queryEmbedding []float64, topK int) ([]RetrievedChunk, error) {
	// Client will add tenant prefix via getCollectionName
	collection := "conversations"
	return r.RetrieveContext(collection, queryEmbedding, topK)
}

// RetrieveProductKnowledge retrieves relevant product knowledge
func (r *Retriever) RetrieveProductKnowledge(tenantID string, queryEmbedding []float64, topK int) ([]RetrievedChunk, error) {
	// Client will add tenant prefix via getCollectionName
	collection := "product_knowledge"
	return r.RetrieveContext(collection, queryEmbedding, topK)
}

