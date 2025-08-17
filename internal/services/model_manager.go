package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"

	"github.com/google/uuid"
)

// ModelManager handles model management operations
type ModelManager struct {
	db           database.Database
	ollamaClient *OllamaClient
	logger       *utils.Logger
	downloadProgress map[string]float64 // Track download progress by model ID
	progressMutex    sync.RWMutex       // Protect progress map
	
	// Cache for available models from Ollama library
	availableModelsCache []string
	cacheLastUpdated     time.Time
	cacheTTL             time.Duration
	cacheMutex           sync.RWMutex
}

// NewModelManager creates a new model manager
func NewModelManager(db database.Database, ollamaClient *OllamaClient, logger *utils.Logger) *ModelManager {
	return &ModelManager{
		db:           db,
		ollamaClient: ollamaClient,
		logger:       logger.WithComponent("model_manager"),
		downloadProgress: make(map[string]float64),
		cacheTTL:     24 * time.Hour, // Cache for 24 hours
	}
}

// SyncModels synchronizes local model database with Ollama
func (m *ModelManager) SyncModels(ctx context.Context) error {
	m.logger.Info().Msg("Starting model synchronization with Ollama")

	// Get models with detailed info from Ollama
	ollamaModelsInfo, err := m.ollamaClient.GetModelsWithInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get models from Ollama: %w", err)
	}

	// Get existing models from database
	existingModels, err := m.GetAllModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing models: %w", err)
	}

	// Create a map of existing models for quick lookup
	existingMap := make(map[string]models.Model)
	for _, model := range existingModels {
		existingMap[model.Name] = model
	}

	// Add or update models from Ollama
	for _, ollamaInfo := range ollamaModelsInfo {
		if existing, exists := existingMap[ollamaInfo.Name]; exists {
			// Update existing model with size and status
			needsUpdate := false

			if existing.Status != "available" {
				if err := m.updateModelStatus(ctx, existing.ID, "available"); err != nil {
					m.logger.Error().Err(err).Str("model", ollamaInfo.Name).Msg("Failed to update model status")
				}
			}

			// Update size if different
			if existing.Size != ollamaInfo.Size {
				if err := m.updateModelSize(ctx, existing.ID, ollamaInfo.Size); err != nil {
					m.logger.Error().Err(err).Str("model", ollamaInfo.Name).Msg("Failed to update model size")
				}
			}

			// Update other fields if we have the info
			if ollamaInfo.Family != "" && existing.Family != ollamaInfo.Family {
				needsUpdate = true
			}
			if ollamaInfo.Format != "" && existing.Format != ollamaInfo.Format {
				needsUpdate = true
			}
			if ollamaInfo.Parameters != "" && existing.Parameters != ollamaInfo.Parameters {
				needsUpdate = true
			}
			if ollamaInfo.Quantization != "" && existing.Quantization != ollamaInfo.Quantization {
				needsUpdate = true
			}

			if needsUpdate {
				if err := m.updateModelMetadata(ctx, existing.ID, ollamaInfo); err != nil {
					m.logger.Error().Err(err).Str("model", ollamaInfo.Name).Msg("Failed to update model metadata")
				}
			}
		} else {
			// Add new model with full info
			newModel := models.Model{
				ID:            uuid.New().String(),
				Name:          ollamaInfo.Name,
				DisplayName:   m.generateDisplayName(ollamaInfo.Name),
				Description:   fmt.Sprintf("Model: %s", ollamaInfo.Name),
				Size:          ollamaInfo.Size,
				Family:        ollamaInfo.Family,
				Format:        ollamaInfo.Format,
				Parameters:    ollamaInfo.Parameters,
				Quantization:  ollamaInfo.Quantization,
				Status:        "available",
				IsEnabled:     true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			if err := m.CreateModel(ctx, newModel); err != nil {
				m.logger.Error().Err(err).Str("model", ollamaInfo.Name).Msg("Failed to create model")
			} else {
				m.logger.Info().Str("model", ollamaInfo.Name).Msg("Added new model")
			}
		}
	}

	// Mark models not in Ollama as removed
	ollamaModelSet := make(map[string]bool)
	for _, modelInfo := range ollamaModelsInfo {
		ollamaModelSet[modelInfo.Name] = true
	}

	for _, existing := range existingModels {
		if !ollamaModelSet[existing.Name] && existing.Status == "available" {
			if err := m.updateModelStatus(ctx, existing.ID, "removed"); err != nil {
				m.logger.Error().Err(err).Str("model", existing.Name).Msg("Failed to mark model as removed")
			} else {
				m.logger.Info().Str("model", existing.Name).Msg("Marked model as removed")
			}
		}
	}

	m.logger.Info().Int("ollama_models", len(ollamaModelsInfo)).Msg("Model synchronization completed")
	return nil
}

// GetAllModels retrieves all models from the database
func (m *ModelManager) GetAllModels(ctx context.Context) ([]models.Model, error) {
	query := `
		SELECT id, name, display_name, description, size, family, format,
		       parameters, quantization, status, is_default, is_enabled,
		       supports_embeddings, embedding_dimensions, created_at, updated_at, last_used_at
		FROM models
		ORDER BY is_default DESC, name ASC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query models: %w", err)
	}
	defer rows.Close()

	var modelList []models.Model
	for rows.Next() {
		var model models.Model
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&model.ID, &model.Name, &model.DisplayName, &model.Description,
			&model.Size, &model.Family, &model.Format, &model.Parameters,
			&model.Quantization, &model.Status, &model.IsDefault, &model.IsEnabled,
			&model.SupportsEmbeddings, &model.EmbeddingDimensions, &model.CreatedAt, &model.UpdatedAt, &lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}

		if lastUsedAt.Valid {
			model.LastUsedAt = &lastUsedAt.Time
		}

		modelList = append(modelList, model)
	}

	return modelList, nil
}

// GetAvailableModels retrieves only available and enabled models
func (m *ModelManager) GetAvailableModels(ctx context.Context) ([]models.Model, error) {
	query := `
		SELECT id, name, display_name, description, size, family, format,
		       parameters, quantization, status, is_default, is_enabled,
		       supports_embeddings, embedding_dimensions, created_at, updated_at, last_used_at
		FROM models
		WHERE status = 'available' AND is_enabled = TRUE
		ORDER BY is_default DESC, name ASC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query available models: %w", err)
	}
	defer rows.Close()

	var modelList []models.Model
	for rows.Next() {
		var model models.Model
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&model.ID, &model.Name, &model.DisplayName, &model.Description,
			&model.Size, &model.Family, &model.Format, &model.Parameters,
			&model.Quantization, &model.Status, &model.IsDefault, &model.IsEnabled,
			&model.SupportsEmbeddings, &model.EmbeddingDimensions, &model.CreatedAt, &model.UpdatedAt, &lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}

		if lastUsedAt.Valid {
			model.LastUsedAt = &lastUsedAt.Time
		}

		modelList = append(modelList, model)
	}

	return modelList, nil
}

// GetModelByID retrieves a model by its ID
func (m *ModelManager) GetModelByID(ctx context.Context, id string) (*models.Model, error) {
	query := `
		SELECT id, name, display_name, description, size, family, format,
		       parameters, quantization, status, is_default, is_enabled,
		       supports_embeddings, embedding_dimensions, created_at, updated_at, last_used_at
		FROM models
		WHERE id = $1
	`

	var model models.Model
	var lastUsedAt sql.NullTime

	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID, &model.Name, &model.DisplayName, &model.Description,
		&model.Size, &model.Family, &model.Format, &model.Parameters,
		&model.Quantization, &model.Status, &model.IsDefault, &model.IsEnabled,
		&model.SupportsEmbeddings, &model.EmbeddingDimensions, &model.CreatedAt, &model.UpdatedAt, &lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if lastUsedAt.Valid {
		model.LastUsedAt = &lastUsedAt.Time
	}

	return &model, nil
}

// GetModelByName retrieves a model by its name
func (m *ModelManager) GetModelByName(ctx context.Context, name string) (*models.Model, error) {
	query := `
		SELECT id, name, display_name, description, size, family, format,
		       parameters, quantization, status, is_default, is_enabled,
		       supports_embeddings, embedding_dimensions, created_at, updated_at, last_used_at
		FROM models
		WHERE name = $1
	`

	var model models.Model
	var lastUsedAt sql.NullTime

	err := m.db.QueryRowContext(ctx, query, name).Scan(
		&model.ID, &model.Name, &model.DisplayName, &model.Description,
		&model.Size, &model.Family, &model.Format, &model.Parameters,
		&model.Quantization, &model.Status, &model.IsDefault, &model.IsEnabled,
		&model.SupportsEmbeddings, &model.EmbeddingDimensions, &model.CreatedAt, &model.UpdatedAt, &lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if lastUsedAt.Valid {
		model.LastUsedAt = &lastUsedAt.Time
	}

	return &model, nil
}

// CreateModel creates a new model in the database
func (m *ModelManager) CreateModel(ctx context.Context, model models.Model) error {
	query := `
		INSERT INTO models (id, name, display_name, description, size, family, format,
		                   parameters, quantization, status, is_default, is_enabled,
		                   supports_embeddings, embedding_dimensions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := m.db.ExecContext(ctx, query,
		model.ID, model.Name, model.DisplayName, model.Description,
		model.Size, model.Family, model.Format, model.Parameters,
		model.Quantization, model.Status, model.IsDefault, model.IsEnabled,
		model.SupportsEmbeddings, model.EmbeddingDimensions, model.CreatedAt, model.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Create default configuration for the model
	config := models.GetDefaultConfig()
	config.ID = uuid.New().String()
	config.ModelID = model.ID
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if err := m.CreateModelConfig(ctx, config); err != nil {
		m.logger.Error().Err(err).Str("model_id", model.ID).Msg("Failed to create default config")
	}

	m.logger.Info().Str("model_id", model.ID).Str("name", model.Name).Msg("Model created")
	return nil
}

// UpdateModel updates an existing model
func (m *ModelManager) UpdateModel(ctx context.Context, id string, req models.ModelUpdateRequest) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	argIndex := 1
	if req.DisplayName != nil {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, *req.DisplayName)
		argIndex++
	}
	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.IsDefault != nil {
		setParts = append(setParts, fmt.Sprintf("is_default = $%d", argIndex))
		args = append(args, *req.IsDefault)
		argIndex++
	}
	if req.IsEnabled != nil {
		setParts = append(setParts, fmt.Sprintf("is_enabled = $%d", argIndex))
		args = append(args, *req.IsEnabled)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE models SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	m.logger.Info().Str("model_id", id).Msg("Model updated")
	return nil
}

// DeleteModel marks a model as removed (soft delete)
func (m *ModelManager) DeleteModel(ctx context.Context, id string) error {
	return m.updateModelStatus(ctx, id, "removed")
}

// HardDeleteModel removes a model from both database and Ollama
func (m *ModelManager) HardDeleteModel(ctx context.Context, id string) error {
	// Get model details first
	model, err := m.GetModelByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	m.logger.Info().Str("model_id", id).Str("model_name", model.Name).Msg("Starting hard delete of model")

	// Delete from Ollama first
	if model.Status == "available" {
		if err := m.ollamaClient.DeleteModel(ctx, model.Name); err != nil {
			m.logger.Error().Err(err).Str("model_name", model.Name).Msg("Failed to delete model from Ollama")
			// Continue with database deletion even if Ollama deletion fails
		} else {
			m.logger.Info().Str("model_name", model.Name).Msg("Model deleted from Ollama")
		}
	}

	// Delete from database
	if err := m.deleteModelFromDatabase(ctx, id); err != nil {
		return fmt.Errorf("failed to delete model from database: %w", err)
	}

	m.logger.Info().Str("model_id", id).Str("model_name", model.Name).Msg("Model hard deleted successfully")
	return nil
}

// deleteModelFromDatabase removes a model and its related data from the database
func (m *ModelManager) deleteModelFromDatabase(ctx context.Context, id string) error {
	// Start a transaction to ensure all deletions succeed or fail together
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete model config (CASCADE will handle this, but explicit is better)
	if _, err := tx.ExecContext(ctx, "DELETE FROM model_configs WHERE model_id = $1", id); err != nil {
		return fmt.Errorf("failed to delete model config: %w", err)
	}

	// Delete the model itself
	result, err := tx.ExecContext(ctx, "DELETE FROM models WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RestoreModel restores a soft-deleted model
func (m *ModelManager) RestoreModel(ctx context.Context, id string) error {
	// Get model details first
	model, err := m.GetModelByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	if model.Status != "removed" {
		return fmt.Errorf("model is not in removed state")
	}

	// Check if model still exists in Ollama
	ollamaModels, err := m.ollamaClient.GetModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to check Ollama models: %w", err)
	}

	modelExists := false
	for _, ollamaModel := range ollamaModels {
		if ollamaModel == model.Name {
			modelExists = true
			break
		}
	}

	if modelExists {
		// Model exists in Ollama, restore as available
		return m.updateModelStatus(ctx, id, "available")
	} else {
		// Model doesn't exist in Ollama, mark as error
		return m.updateModelStatus(ctx, id, "error")
	}
}

// GetModelConfig retrieves model configuration
func (m *ModelManager) GetModelConfig(ctx context.Context, modelID string) (*models.ModelConfig, error) {
	query := `
		SELECT id, model_id, temperature, top_p, top_k, repeat_penalty,
		       context_length, max_tokens, system_prompt, created_at, updated_at
		FROM model_configs
		WHERE model_id = $1
	`

	var config models.ModelConfig
	var temperature, topP, repeatPenalty sql.NullFloat64
	var topK, contextLength, maxTokens sql.NullInt64

	err := m.db.QueryRowContext(ctx, query, modelID).Scan(
		&config.ID, &config.ModelID, &temperature, &topP, &topK, &repeatPenalty,
		&contextLength, &maxTokens, &config.SystemPrompt, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model config not found")
		}
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	// Handle nullable fields
	if temperature.Valid {
		config.Temperature = &temperature.Float64
	}
	if topP.Valid {
		config.TopP = &topP.Float64
	}
	if topK.Valid {
		val := int(topK.Int64)
		config.TopK = &val
	}
	if repeatPenalty.Valid {
		config.RepeatPenalty = &repeatPenalty.Float64
	}
	if contextLength.Valid {
		val := int(contextLength.Int64)
		config.ContextLength = &val
	}
	if maxTokens.Valid {
		val := int(maxTokens.Int64)
		config.MaxTokens = &val
	}

	return &config, nil
}

// CreateModelConfig creates a new model configuration
func (m *ModelManager) CreateModelConfig(ctx context.Context, config models.ModelConfig) error {
	query := `
		INSERT INTO model_configs (id, model_id, temperature, top_p, top_k, repeat_penalty,
		                          context_length, max_tokens, system_prompt, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := m.db.ExecContext(ctx, query,
		config.ID, config.ModelID, config.Temperature, config.TopP, config.TopK,
		config.RepeatPenalty, config.ContextLength, config.MaxTokens, config.SystemPrompt,
		config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create model config: %w", err)
	}

	return nil
}

// UpdateModelConfig updates model configuration
func (m *ModelManager) UpdateModelConfig(ctx context.Context, modelID string, req models.ModelConfigUpdateRequest) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	argIndex := 1
	if req.Temperature != nil {
		setParts = append(setParts, fmt.Sprintf("temperature = $%d", argIndex))
		args = append(args, *req.Temperature)
		argIndex++
	}
	if req.TopP != nil {
		setParts = append(setParts, fmt.Sprintf("top_p = $%d", argIndex))
		args = append(args, *req.TopP)
		argIndex++
	}
	if req.TopK != nil {
		setParts = append(setParts, fmt.Sprintf("top_k = $%d", argIndex))
		args = append(args, *req.TopK)
		argIndex++
	}
	if req.RepeatPenalty != nil {
		setParts = append(setParts, fmt.Sprintf("repeat_penalty = $%d", argIndex))
		args = append(args, *req.RepeatPenalty)
		argIndex++
	}
	if req.ContextLength != nil {
		setParts = append(setParts, fmt.Sprintf("context_length = $%d", argIndex))
		args = append(args, *req.ContextLength)
		argIndex++
	}
	if req.MaxTokens != nil {
		setParts = append(setParts, fmt.Sprintf("max_tokens = $%d", argIndex))
		args = append(args, *req.MaxTokens)
		argIndex++
	}
	if req.SystemPrompt != nil {
		setParts = append(setParts, fmt.Sprintf("system_prompt = $%d", argIndex))
		args = append(args, *req.SystemPrompt)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, modelID)

	query := fmt.Sprintf("UPDATE model_configs SET %s WHERE model_id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update model config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model config not found")
	}

	// Handle custom options if provided
	if req.CustomOptions != nil {
		// Store custom options as JSON in a separate operation if needed
		optionsJSON, _ := json.Marshal(req.CustomOptions)
		m.logger.Info().
			Str("model_id", modelID).
			Str("custom_options", string(optionsJSON)).
			Msg("Custom options update requested (not yet implemented)")
	}

	m.logger.Info().Str("model_id", modelID).Msg("Model config updated")
	return nil
}

// GetModelUsageStats retrieves usage statistics for a model
func (m *ModelManager) GetModelUsageStats(ctx context.Context, modelID string) (*models.ModelUsageStats, error) {
	query := `
		SELECT model_id, total_messages, total_tokens, last_used_at
		FROM model_usage_stats
		WHERE model_id = $1
	`

	var stats models.ModelUsageStats
	var lastUsedAt sql.NullTime

	err := m.db.QueryRowContext(ctx, query, modelID).Scan(
		&stats.ModelID, &stats.TotalMessages, &stats.TotalTokens, &lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty stats if no usage found
			return &models.ModelUsageStats{
				ModelID:       modelID,
				TotalMessages: 0,
				TotalTokens:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get model usage stats: %w", err)
	}

	if lastUsedAt.Valid {
		stats.LastUsedAt = &lastUsedAt.Time
	}

	return &stats, nil
}

// GetDefaultModel retrieves the default model
func (m *ModelManager) GetDefaultModel(ctx context.Context) (*models.Model, error) {
	query := `
		SELECT id, name, display_name, description, size, family, format,
		       parameters, quantization, status, is_default, is_enabled,
		       supports_embeddings, embedding_dimensions, created_at, updated_at, last_used_at
		FROM models
		WHERE is_default = TRUE AND is_enabled = TRUE AND status = 'available'
		LIMIT 1
	`

	var model models.Model
	var lastUsedAt sql.NullTime

	err := m.db.QueryRowContext(ctx, query).Scan(
		&model.ID, &model.Name, &model.DisplayName, &model.Description,
		&model.Size, &model.Family, &model.Format, &model.Parameters,
		&model.Quantization, &model.Status, &model.IsDefault, &model.IsEnabled,
		&model.SupportsEmbeddings, &model.EmbeddingDimensions, &model.CreatedAt, &model.UpdatedAt, &lastUsedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no default model found")
		}
		return nil, fmt.Errorf("failed to get default model: %w", err)
	}

	if lastUsedAt.Valid {
		model.LastUsedAt = &lastUsedAt.Time
	}

	return &model, nil
}

// SetDefaultModel sets a model as the default
func (m *ModelManager) SetDefaultModel(ctx context.Context, modelID string) error {
	// First, verify the model exists and is available
	model, err := m.GetModelByID(ctx, modelID)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	if model.Status != "available" {
		return fmt.Errorf("cannot set unavailable model as default")
	}

	// Update the model to be default (trigger will handle unsetting others)
	query := `UPDATE models SET is_default = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	
	result, err := m.db.ExecContext(ctx, query, modelID)
	if err != nil {
		return fmt.Errorf("failed to set default model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	m.logger.Info().Str("model_id", modelID).Str("name", model.Name).Msg("Default model updated")
	return nil
}

// ValidateModel checks if a model is valid for use
func (m *ModelManager) ValidateModel(ctx context.Context, modelName string) error {
	model, err := m.GetModelByName(ctx, modelName)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	if model.Status != "available" {
		return fmt.Errorf("model %s is not available (status: %s)", modelName, model.Status)
	}

	if !model.IsEnabled {
		return fmt.Errorf("model %s is disabled", modelName)
	}

	return nil
}

// updateModelStatus updates the status of a model
func (m *ModelManager) updateModelStatus(ctx context.Context, id, status string) error {
	if !models.ValidateModelStatus(status) {
		return fmt.Errorf("invalid model status: %s", status)
	}

	query := `UPDATE models SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	
	result, err := m.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update model status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	return nil
}

// generateDisplayName generates a user-friendly display name from model name
func (m *ModelManager) generateDisplayName(modelName string) string {
	// Remove common suffixes and format nicely
	name := strings.ReplaceAll(modelName, ":", " ")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	
	// Capitalize first letter of each word
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	
	return strings.Join(words, " ")
}

// updateModelSize updates the size of a model
func (m *ModelManager) updateModelSize(ctx context.Context, id string, size int64) error {
	query := `UPDATE models SET size = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	
	result, err := m.db.ExecContext(ctx, query, size, id)
	if err != nil {
		return fmt.Errorf("failed to update model size: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	return nil
}

// GetModelDetails retrieves detailed model information including config and stats
func (m *ModelManager) GetModelDetails(ctx context.Context, modelID string) (*models.ModelDetailsResponse, error) {
	// Get the model
	model, err := m.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// Get model config
	config, err := m.GetModelConfig(ctx, modelID)
	if err != nil && err.Error() != "model config not found" {
		m.logger.Warn().Err(err).Str("model_id", modelID).Msg("Failed to get model config")
	}

	// Get model stats
	stats, err := m.GetModelUsageStats(ctx, modelID)
	if err != nil {
		m.logger.Warn().Err(err).Str("model_id", modelID).Msg("Failed to get model stats")
	}

	return &models.ModelDetailsResponse{
		Model:  *model,
		Config: config,
		Stats:  stats,
	}, nil
}

// DownloadModel initiates a model download
func (m *ModelManager) DownloadModel(ctx context.Context, req models.ModelDownloadRequest) (*models.ModelDownloadResponse, error) {
	// Check if model already exists
	existingModel, err := m.GetModelByName(ctx, req.Name)
	if err == nil && existingModel.Status == "available" {
		return nil, fmt.Errorf("model %s already exists and is available", req.Name)
	}

	// Create or update model record
	var modelID string
	if existingModel != nil {
		modelID = existingModel.ID
		// Update status to downloading
		if err := m.updateModelStatus(ctx, modelID, "downloading"); err != nil {
			return nil, fmt.Errorf("failed to update model status: %w", err)
		}
	} else {
		// Create new model record
		newModel := models.Model{
			ID:          uuid.New().String(),
			Name:        req.Name,
			DisplayName: req.DisplayName,
			Description: req.Description,
			Status:      "downloading",
			IsEnabled:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		if newModel.DisplayName == "" {
			newModel.DisplayName = m.generateDisplayName(req.Name)
		}
		if newModel.Description == "" {
			newModel.Description = fmt.Sprintf("Model: %s", req.Name)
		}

		if err := m.CreateModel(ctx, newModel); err != nil {
			return nil, fmt.Errorf("failed to create model record: %w", err)
		}
		modelID = newModel.ID
	}

	// Start download in background
	go func() {
		progressChan := make(chan models.ModelDownloadProgress, 100)
		
		// Start the download
		go func() {
			defer close(progressChan)
			if err := m.ollamaClient.PullModel(context.Background(), req.Name, progressChan); err != nil {
				m.logger.Error().Err(err).Str("model", req.Name).Msg("Model download failed")
				m.updateModelStatus(context.Background(), modelID, "error")
				return
			}
		}()

		// Process progress updates
		for progress := range progressChan {
			m.progressMutex.Lock()
			m.downloadProgress[modelID] = progress.Percentage
			m.progressMutex.Unlock()

			if progress.Status == "error" {
				m.updateModelStatus(context.Background(), modelID, "error")
				break
			}
			
			// Check for completion
			if progress.Percentage >= 100 || progress.Status == "success" {
				m.updateModelStatus(context.Background(), modelID, "available")
				
				// Sync to get updated model info
				m.SyncModels(context.Background())
				break
			}
		}

		// Clean up progress tracking
		m.progressMutex.Lock()
		delete(m.downloadProgress, modelID)
		m.progressMutex.Unlock()
	}()

	return &models.ModelDownloadResponse{
		ID:      modelID,
		Name:    req.Name,
		Status:  "downloading",
		Message: "Model download started",
	}, nil
}

// GetAvailableModelsFromOllama retrieves available models from Ollama library
func (m *ModelManager) GetAvailableModelsFromOllama(ctx context.Context) ([]string, error) {
	m.cacheMutex.RLock()
	if time.Since(m.cacheLastUpdated) < m.cacheTTL && len(m.availableModelsCache) > 0 {
		models := make([]string, len(m.availableModelsCache))
		copy(models, m.availableModelsCache)
		m.cacheMutex.RUnlock()
		return models, nil
	}
	m.cacheMutex.RUnlock()

	// Cache is stale or empty, refresh it
	return m.refreshAvailableModelsCache(ctx)
}

// RefreshAvailableModelsCache refreshes the cache of available models
func (m *ModelManager) RefreshAvailableModelsCache(ctx context.Context) error {
	_, err := m.refreshAvailableModelsCache(ctx)
	return err
}

// refreshAvailableModelsCache internal method to refresh cache
func (m *ModelManager) refreshAvailableModelsCache(ctx context.Context) ([]string, error) {
	m.logger.Info().Msg("Refreshing available models cache")

	models, err := m.ollamaClient.GetLibraryModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get library models: %w", err)
	}

	m.cacheMutex.Lock()
	m.availableModelsCache = models
	m.cacheLastUpdated = time.Now()
	m.cacheMutex.Unlock()

	m.logger.Info().Int("model_count", len(models)).Msg("Available models cache refreshed")
	return models, nil
}

// GetModelDownloadStatus retrieves download status for a model
func (m *ModelManager) GetModelDownloadStatus(ctx context.Context, modelID string) (*models.Model, error) {
	model, err := m.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// Add progress information if downloading
	if model.Status == "downloading" {
		m.progressMutex.RLock()
		if progress, exists := m.downloadProgress[modelID]; exists {
			model.Progress = progress
		}
		m.progressMutex.RUnlock()
	}

	return model, nil
}

// GetCacheInfo returns information about the models cache
func (m *ModelManager) GetCacheInfo() map[string]interface{} {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_size":        len(m.availableModelsCache),
		"last_updated":      m.cacheLastUpdated,
		"cache_ttl_hours":   m.cacheTTL.Hours(),
		"cache_valid":       time.Since(m.cacheLastUpdated) < m.cacheTTL,
		"time_until_stale":  m.cacheTTL - time.Since(m.cacheLastUpdated),
	}
}

// updateModelMetadata updates model metadata from Ollama info
func (m *ModelManager) updateModelMetadata(ctx context.Context, id string, info models.OllamaModelDetailedInfo) error {
	query := `
		UPDATE models
		SET family = $1, format = $2, parameters = $3, quantization = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`
	
	result, err := m.db.ExecContext(ctx, query, info.Family, info.Format, info.Parameters, info.Quantization, id)
	if err != nil {
		return fmt.Errorf("failed to update model metadata: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	return nil
}