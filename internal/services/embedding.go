package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"chat_ollama/internal/config"
	"chat_ollama/internal/utils"
)

// EmbeddingService handles text embedding generation using Ollama
type EmbeddingService struct {
	client     *http.Client
	ollamaHost string
	logger     *utils.Logger
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response from Ollama embedding API
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(cfg *config.Config, logger *utils.Logger) *EmbeddingService {
	return &EmbeddingService{
		client: &http.Client{
			Timeout: cfg.OllamaTimeout,
		},
		ollamaHost: cfg.OllamaHost,
		logger:     logger.WithComponent("embedding_service"),
	}
}

// GenerateEmbedding generates an embedding for the given text using the specified model
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text, model string) ([]float32, error) {
	if model == "" {
		model = "nomic-embed-text" // Default embedding model
	}

	req := EmbeddingRequest{
		Model:  model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	url := fmt.Sprintf("http://%s/api/embeddings", s.ollamaHost)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	s.logger.Debug().
		Str("model", model).
		Str("text_preview", truncateText(text, 100)).
		Msg("Generating embedding")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	s.logger.Debug().
		Str("model", model).
		Int("embedding_dimensions", len(embeddingResp.Embedding)).
		Msg("Embedding generated successfully")

	return embeddingResp.Embedding, nil
}

// GenerateEmbeddingBatch generates embeddings for multiple texts
func (s *EmbeddingService) GenerateEmbeddingBatch(ctx context.Context, texts []string, model string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		embedding, err := s.GenerateEmbedding(ctx, text, model)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetEmbeddingDimensions returns the dimensions of embeddings for a given model
func (s *EmbeddingService) GetEmbeddingDimensions(ctx context.Context, model string) (int, error) {
	// Generate a test embedding to determine dimensions
	testEmbedding, err := s.GenerateEmbedding(ctx, "test", model)
	if err != nil {
		return 0, fmt.Errorf("failed to determine embedding dimensions: %w", err)
	}
	
	return len(testEmbedding), nil
}

// ValidateEmbeddingModel checks if the embedding model is available
func (s *EmbeddingService) ValidateEmbeddingModel(ctx context.Context, model string) error {
	// Try to generate a small test embedding
	_, err := s.GenerateEmbedding(ctx, "test", model)
	if err != nil {
		return fmt.Errorf("embedding model %s is not available: %w", model, err)
	}
	return nil
}

// GetAvailableEmbeddingModels returns a list of available embedding models
func (s *EmbeddingService) GetAvailableEmbeddingModels(ctx context.Context) ([]string, error) {
	// This would typically query Ollama for available models
	// For now, return common embedding models
	return []string{
		"nomic-embed-text",
		"all-minilm",
		"mxbai-embed-large",
	}, nil
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

// truncateText truncates text to a specified length for logging
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}