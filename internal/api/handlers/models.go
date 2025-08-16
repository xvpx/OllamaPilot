package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/services"
	"chat_ollama/internal/utils"
)

// ModelsHandler handles model management requests
type ModelsHandler struct {
	modelManager *services.ModelManager
	logger       *utils.Logger
}

// NewModelsHandler creates a new models handler
func NewModelsHandler(db *database.DB, cfg *config.Config, logger *utils.Logger) *ModelsHandler {
	// Create Ollama client
	ollamaClient := services.NewOllamaClient(cfg.OllamaHost, cfg.OllamaTimeout, logger)
	
	// Create model manager
	modelManager := services.NewModelManager(db, ollamaClient, logger)

	return &ModelsHandler{
		modelManager: modelManager,
		logger:       logger.WithComponent("models_handler"),
	}
}

// GetModels handles GET /v1/models
func (h *ModelsHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Check query parameter for available only
	availableOnly := r.URL.Query().Get("available") == "true"

	logger.Info().Bool("available_only", availableOnly).Msg("Getting models list")

	var modelList []models.Model
	var err error

	if availableOnly {
		modelList, err = h.modelManager.GetAvailableModels(ctx)
	} else {
		modelList, err = h.modelManager.GetAllModels(ctx)
	}

	if err != nil {
		logger.Error().Err(err).Msg("Failed to get models")
		apiErr := utils.NewInternalError("Failed to retrieve models", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.ModelListResponse{
		Models: modelList,
		Total:  len(modelList),
	}

	logger.Info().Int("model_count", len(modelList)).Msg("Models retrieved successfully")
	utils.WriteSuccess(w, response)
}

// GetModel handles GET /v1/models/{id}
func (h *ModelsHandler) GetModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Getting model details")

	modelDetails, err := h.modelManager.GetModelDetails(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get model details")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to retrieve model details", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model details retrieved successfully")
	utils.WriteSuccess(w, modelDetails)
}

// UpdateModel handles PUT /v1/models/{id}
func (h *ModelsHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	var req models.ModelUpdateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse model update request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Updating model")

	if err := h.modelManager.UpdateModel(ctx, modelID, req); err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to update model")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to update model", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get updated model details
	modelDetails, err := h.modelManager.GetModelDetails(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get updated model details")
		apiErr := utils.NewInternalError("Model updated but failed to retrieve details", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model updated successfully")
	utils.WriteSuccess(w, modelDetails)
}

// DeleteModel handles DELETE /v1/models/{id} (soft delete)
func (h *ModelsHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Soft deleting model")

	if err := h.modelManager.DeleteModel(ctx, modelID); err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to delete model")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to delete model", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model soft deleted successfully")
	utils.WriteSuccess(w, map[string]string{"message": "Model marked as removed"})
}

// HardDeleteModel handles DELETE /v1/models/{id}/hard (hard delete)
func (h *ModelsHandler) HardDeleteModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute) // Longer timeout for hard delete
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Hard deleting model")

	if err := h.modelManager.HardDeleteModel(ctx, modelID); err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to hard delete model")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to permanently delete model", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model hard deleted successfully")
	utils.WriteSuccess(w, map[string]string{"message": "Model permanently deleted"})
}

// RestoreModel handles POST /v1/models/{id}/restore
func (h *ModelsHandler) RestoreModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Restoring model")

	if err := h.modelManager.RestoreModel(ctx, modelID); err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to restore model")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		if err.Error() == "model is not in removed state" {
			apiErr := utils.NewValidationError("Model is not in removed state", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to restore model", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get updated model details
	modelDetails, err := h.modelManager.GetModelDetails(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get restored model details")
		apiErr := utils.NewInternalError("Model restored but failed to retrieve details", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model restored successfully")
	utils.WriteSuccess(w, modelDetails)
}

// SyncModels handles POST /v1/models/sync
func (h *ModelsHandler) SyncModels(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	logger.Info().Msg("Starting model synchronization")

	if err := h.modelManager.SyncModels(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to sync models")
		apiErr := utils.NewInternalError("Failed to synchronize models", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get updated models list
	modelList, err := h.modelManager.GetAllModels(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get models after sync")
		apiErr := utils.NewInternalError("Models synced but failed to retrieve list", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.ModelListResponse{
		Models: modelList,
		Total:  len(modelList),
	}

	logger.Info().Int("model_count", len(modelList)).Msg("Model synchronization completed")
	utils.WriteSuccess(w, response)
}

// GetModelConfig handles GET /v1/models/{id}/config
func (h *ModelsHandler) GetModelConfig(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Getting model configuration")

	config, err := h.modelManager.GetModelConfig(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get model config")
		
		if err.Error() == "model config not found" {
			apiErr := utils.NewNotFoundError("Model configuration not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to retrieve model configuration", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model configuration retrieved successfully")
	utils.WriteSuccess(w, config)
}

// UpdateModelConfig handles PUT /v1/models/{id}/config
func (h *ModelsHandler) UpdateModelConfig(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	var req models.ModelConfigUpdateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse model config update request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Updating model configuration")

	if err := h.modelManager.UpdateModelConfig(ctx, modelID, req); err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to update model config")
		
		if err.Error() == "model config not found" {
			apiErr := utils.NewNotFoundError("Model configuration not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to update model configuration", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get updated config
	config, err := h.modelManager.GetModelConfig(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get updated model config")
		apiErr := utils.NewInternalError("Configuration updated but failed to retrieve details", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model configuration updated successfully")
	utils.WriteSuccess(w, config)
}

// SetDefaultModel handles POST /v1/models/{id}/default
func (h *ModelsHandler) SetDefaultModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Setting default model")

	if err := h.modelManager.SetDefaultModel(ctx, modelID); err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to set default model")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		if err.Error() == "cannot set unavailable model as default" {
			apiErr := utils.NewValidationError("Cannot set unavailable model as default", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to set default model", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get updated model details
	modelDetails, err := h.modelManager.GetModelDetails(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get updated model details")
		apiErr := utils.NewInternalError("Default model set but failed to retrieve details", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Default model set successfully")
	utils.WriteSuccess(w, modelDetails)
}

// GetModelStats handles GET /v1/models/{id}/stats
func (h *ModelsHandler) GetModelStats(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Getting model usage statistics")

	stats, err := h.modelManager.GetModelUsageStats(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get model stats")
		apiErr := utils.NewInternalError("Failed to retrieve model statistics", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_id", modelID).Msg("Model statistics retrieved successfully")
	utils.WriteSuccess(w, stats)
}

// DownloadModel handles POST /v1/models/download
func (h *ModelsHandler) DownloadModel(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	var req models.ModelDownloadRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse model download request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Validate required fields
	if req.Name == "" {
		apiErr := utils.NewValidationError("Model name is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_name", req.Name).Msg("Starting model download")

	response, err := h.modelManager.DownloadModel(ctx, req)
	if err != nil {
		logger.Error().Err(err).Str("model_name", req.Name).Msg("Failed to start model download")
		
		if err.Error() == fmt.Sprintf("model %s already exists and is available", req.Name) {
			apiErr := utils.NewValidationError("Model already exists and is available", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to start model download", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("model_name", req.Name).Str("model_id", response.ID).Msg("Model download started successfully")
	utils.WriteSuccess(w, response)
}

// GetAvailableModels handles GET /v1/models/available
func (h *ModelsHandler) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Msg("Getting available models for download")

	availableModels, err := h.modelManager.GetAvailableModelsFromOllama(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get available models")
		apiErr := utils.NewInternalError("Failed to retrieve available models", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := struct {
		Models []string `json:"models"`
		Total  int      `json:"total"`
	}{
		Models: availableModels,
		Total:  len(availableModels),
	}

	logger.Info().Int("model_count", len(availableModels)).Msg("Available models retrieved successfully")
	utils.WriteSuccess(w, response)
}

// GetModelDownloadStatus handles GET /v1/models/{id}/download-status
func (h *ModelsHandler) GetModelDownloadStatus(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract model ID from URL path
	modelID := chi.URLParam(r, "modelID")
	if modelID == "" {
		apiErr := utils.NewValidationError("Model ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("model_id", modelID).Msg("Getting model download status")

	model, err := h.modelManager.GetModelDownloadStatus(ctx, modelID)
	if err != nil {
		logger.Error().Err(err).Str("model_id", modelID).Msg("Failed to get model download status")
		
		if err.Error() == "model not found" {
			apiErr := utils.NewNotFoundError("Model not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		
		apiErr := utils.NewInternalError("Failed to retrieve model download status", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Status   string  `json:"status"`
		Progress float64 `json:"progress,omitempty"`
	}{
		ID:       model.ID,
		Name:     model.Name,
		Status:   model.Status,
		Progress: model.Progress,
	}

	logger.Info().Str("model_id", modelID).Str("status", model.Status).Msg("Model download status retrieved successfully")
	utils.WriteSuccess(w, response)
}

// RefreshAvailableModels handles POST /v1/models/available/refresh
func (h *ModelsHandler) RefreshAvailableModels(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second) // Longer timeout for refresh
	defer cancel()

	logger.Info().Msg("Refreshing available models cache")

	if err := h.modelManager.RefreshAvailableModelsCache(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to refresh available models cache")
		apiErr := utils.NewInternalError("Failed to refresh available models cache", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get the updated list
	availableModels, err := h.modelManager.GetAvailableModelsFromOllama(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get available models after refresh")
		apiErr := utils.NewInternalError("Failed to retrieve available models after refresh", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := struct {
		Models    []string               `json:"models"`
		Total     int                    `json:"total"`
		CacheInfo map[string]interface{} `json:"cache_info"`
	}{
		Models:    availableModels,
		Total:     len(availableModels),
		CacheInfo: h.modelManager.GetCacheInfo(),
	}

	logger.Info().Int("model_count", len(availableModels)).Msg("Available models cache refreshed successfully")
	utils.WriteSuccess(w, response)
}

// GetCacheInfo handles GET /v1/models/cache-info
func (h *ModelsHandler) GetCacheInfo(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	logger.Info().Msg("Getting cache information")

	cacheInfo := h.modelManager.GetCacheInfo()

	logger.Info().Msg("Cache information retrieved successfully")
	utils.WriteSuccess(w, cacheInfo)
}