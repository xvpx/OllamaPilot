package handlers

import (
	"context"
	"net/http"
	"time"

	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/services"
	"chat_ollama/internal/utils"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db           database.Database
	ollamaClient *services.OllamaClient
	logger       *utils.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db database.Database, cfg *config.Config, logger *utils.Logger) *HealthHandler {
	// Create Ollama client for health checks
	ollamaClient := services.NewOllamaClient(cfg.OllamaHost, cfg.OllamaTimeout, logger)

	return &HealthHandler{
		db:           db,
		ollamaClient: ollamaClient,
		logger:       logger.WithComponent("health_handler"),
	}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	logger := utils.FromContext(ctx)
	if logger == nil {
		logger = h.logger
	}

	response := models.HealthResponse{
		Status:    models.StatusHealthy,
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Metadata:  make(map[string]interface{}),
	}

	// Check database health
	dbStatus := h.checkDatabaseHealth(ctx)
	response.Services["database"] = dbStatus

	// Check Ollama health
	ollamaStatus := h.checkOllamaHealth(ctx)
	response.Services["ollama"] = ollamaStatus

	// Determine overall status
	overallStatus := models.StatusHealthy
	for _, status := range response.Services {
		if status == models.StatusUnhealthy {
			overallStatus = models.StatusUnhealthy
			break
		} else if status == models.StatusDegraded {
			overallStatus = models.StatusDegraded
		}
	}
	response.Status = overallStatus

	// Add metadata
	if h.db != nil {
		if pgDB, ok := h.db.(*database.PostgresDB); ok {
			if version, err := pgDB.GetVersion(); err == nil {
				response.Metadata["postgres_version"] = version
			}
			// Check pgvector extension
			if err := pgDB.CheckPgvectorExtension(); err == nil {
				response.Metadata["pgvector_enabled"] = true
			} else {
				response.Metadata["pgvector_enabled"] = false
			}
		}
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if overallStatus == models.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == models.StatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	logger.Info().
		Str("overall_status", overallStatus).
		Interface("services", response.Services).
		Msg("Health check completed")

	utils.WriteJSON(w, statusCode, response)
}

// checkDatabaseHealth checks if the database is healthy
func (h *HealthHandler) checkDatabaseHealth(ctx context.Context) string {
	if h.db == nil {
		return models.StatusUnhealthy
	}

	// Create a channel to handle the health check with timeout
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- h.db.HealthCheck()
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			h.logger.Error().Err(err).Msg("Database health check failed")
			return models.StatusUnhealthy
		}
		return models.StatusHealthy
	case <-ctx.Done():
		h.logger.Error().Msg("Database health check timed out")
		return models.StatusUnhealthy
	}
}

// checkOllamaHealth checks if Ollama is healthy
func (h *HealthHandler) checkOllamaHealth(ctx context.Context) string {
	if h.ollamaClient == nil {
		return models.StatusUnhealthy
	}

	// Create a channel to handle the health check with timeout
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- h.ollamaClient.HealthCheck(ctx)
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			h.logger.Error().Err(err).Msg("Ollama health check failed")
			return models.StatusUnhealthy
		}
		return models.StatusHealthy
	case <-ctx.Done():
		h.logger.Error().Msg("Ollama health check timed out")
		return models.StatusUnhealthy
	}
}

// ReadinessCheck handles GET /ready (for Kubernetes readiness probes)
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	logger := utils.FromContext(ctx)
	if logger == nil {
		logger = h.logger
	}

	// For readiness, we only check critical services
	dbStatus := h.checkDatabaseHealth(ctx)

	if dbStatus == models.StatusHealthy {
		logger.Debug().Msg("Readiness check passed")
		utils.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ready",
		})
	} else {
		logger.Warn().Str("db_status", dbStatus).Msg("Readiness check failed")
		utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"reason": "database not healthy",
		})
	}
}

// LivenessCheck handles GET /live (for Kubernetes liveness probes)
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	logger.Debug().Msg("Liveness check")

	// Liveness check is simple - if we can respond, we're alive
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "alive",
	})
}