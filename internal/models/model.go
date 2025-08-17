package models

import (
	"time"
)

// Model represents a language model with metadata
type Model struct {
	ID                   string    `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	DisplayName          string    `json:"display_name" db:"display_name"`
	Description          string    `json:"description" db:"description"`
	Size                 int64     `json:"size" db:"size"`
	Family               string    `json:"family" db:"family"`
	Format               string    `json:"format" db:"format"`
	Parameters           string    `json:"parameters" db:"parameters"`
	Quantization         string    `json:"quantization" db:"quantization"`
	Status               string    `json:"status" db:"status"` // available, downloading, installing, error
	Progress             float64   `json:"progress,omitempty" db:"-"` // Download progress percentage (0-100)
	IsDefault            bool      `json:"is_default" db:"is_default"`
	IsEnabled            bool      `json:"is_enabled" db:"is_enabled"`
	SupportsEmbeddings   bool      `json:"supports_embeddings" db:"supports_embeddings"`
	EmbeddingDimensions  int       `json:"embedding_dimensions" db:"embedding_dimensions"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}

// ModelConfig represents model configuration settings
type ModelConfig struct {
	ID               string                 `json:"id" db:"id"`
	ModelID          string                 `json:"model_id" db:"model_id"`
	Temperature      *float64               `json:"temperature,omitempty" db:"temperature"`
	TopP             *float64               `json:"top_p,omitempty" db:"top_p"`
	TopK             *int                   `json:"top_k,omitempty" db:"top_k"`
	RepeatPenalty    *float64               `json:"repeat_penalty,omitempty" db:"repeat_penalty"`
	ContextLength    *int                   `json:"context_length,omitempty" db:"context_length"`
	MaxTokens        *int                   `json:"max_tokens,omitempty" db:"max_tokens"`
	SystemPrompt     string                 `json:"system_prompt" db:"system_prompt"`
	CustomOptions    map[string]interface{} `json:"custom_options,omitempty"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// ModelUsageStats represents usage statistics for a model
type ModelUsageStats struct {
	ModelID       string `json:"model_id" db:"model_id"`
	TotalMessages int    `json:"total_messages" db:"total_messages"`
	TotalTokens   int    `json:"total_tokens" db:"total_tokens"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}

// ModelInstallRequest represents a request to install a model
type ModelInstallRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ModelInstallResponse represents the response for model installation
type ModelInstallResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ModelListResponse represents the response for listing models
type ModelListResponse struct {
	Models []Model `json:"models"`
	Total  int     `json:"total"`
}

// ModelDetailsResponse represents detailed model information
type ModelDetailsResponse struct {
	Model  Model            `json:"model"`
	Config *ModelConfig     `json:"config,omitempty"`
	Stats  *ModelUsageStats `json:"stats,omitempty"`
}

// ModelUpdateRequest represents a request to update model metadata
type ModelUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
	IsEnabled   *bool   `json:"is_enabled,omitempty"`
}

// ModelConfigUpdateRequest represents a request to update model configuration
type ModelConfigUpdateRequest struct {
	Temperature   *float64               `json:"temperature,omitempty"`
	TopP          *float64               `json:"top_p,omitempty"`
	TopK          *int                   `json:"top_k,omitempty"`
	RepeatPenalty *float64               `json:"repeat_penalty,omitempty"`
	ContextLength *int                   `json:"context_length,omitempty"`
	MaxTokens     *int                   `json:"max_tokens,omitempty"`
	SystemPrompt  *string                `json:"system_prompt,omitempty"`
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

// ValidateModelStatus validates if the model status is valid
func ValidateModelStatus(status string) bool {
	switch status {
	case "available", "downloading", "installing", "error", "removed":
		return true
	default:
		return false
	}
}

// ValidateModelFormat validates if the model format is valid
func ValidateModelFormat(format string) bool {
	switch format {
	case "gguf", "ggml", "safetensors", "pytorch", "onnx":
		return true
	default:
		return false
	}
}

// GetDefaultConfig returns default configuration for a model
func GetDefaultConfig() ModelConfig {
	return ModelConfig{
		Temperature:   floatPtr(0.7),
		TopP:          floatPtr(0.9),
		TopK:          intPtr(40),
		RepeatPenalty: floatPtr(1.1),
		ContextLength: intPtr(4096),
		MaxTokens:     intPtr(2048),
		SystemPrompt:  "",
		CustomOptions: make(map[string]interface{}),
	}
}

// Helper functions for pointer creation
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

// ModelDownloadProgress represents the progress of a model download
type ModelDownloadProgress struct {
	ModelName  string  `json:"model_name"`
	Status     string  `json:"status"`
	Digest     string  `json:"digest,omitempty"`
	Total      int64   `json:"total,omitempty"`
	Completed  int64   `json:"completed,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// OllamaModelInfo represents detailed model information from Ollama
type OllamaModelInfo struct {
	License    string                 `json:"license,omitempty"`
	Modelfile  string                 `json:"modelfile,omitempty"`
	Parameters string                 `json:"parameters,omitempty"`
	Template   string                 `json:"template,omitempty"`
	System     string                 `json:"system,omitempty"`
	Details    OllamaModelDetails     `json:"details,omitempty"`
	ModelInfo  map[string]interface{} `json:"model_info,omitempty"`
}

// OllamaModelDetails represents model details from Ollama
type OllamaModelDetails struct {
	Format            string   `json:"format,omitempty"`
	Family            string   `json:"family,omitempty"`
	Families          []string `json:"families,omitempty"`
	ParameterSize     string   `json:"parameter_size,omitempty"`
	QuantizationLevel string   `json:"quantization_level,omitempty"`
}

// ModelDownloadRequest represents a request to download a model
type ModelDownloadRequest struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ModelDownloadResponse represents the response for model download initiation
type ModelDownloadResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// OllamaModelDetailedInfo represents detailed model information from Ollama with size
type OllamaModelDetailedInfo struct {
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	Family        string `json:"family"`
	Format        string `json:"format"`
	Parameters    string `json:"parameters"`
	Quantization  string `json:"quantization"`
}